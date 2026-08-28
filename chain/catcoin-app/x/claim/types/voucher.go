package types

import (
	"crypto/ed25519"
	"fmt"
	"strconv"
	"strings"
)

const voucherDomain = "catcoin.claim.v1"

// CanonicalVoucherBytes 返回唯一且带域分隔的签名载荷，防止跨链与跨消息重放。
func CanonicalVoucherBytes(voucher Voucher) ([]byte, error) {
	fields := []string{
		voucherDomain,
		voucher.ChainId,
		voucher.Address,
		voucher.CredentialId,
		strconv.FormatInt(voucher.ExpiresAtUnix, 10),
		voucher.IssuerKeyId,
	}
	for _, field := range fields {
		if field == "" || strings.Contains(field, "|") || strings.TrimSpace(field) != field {
			return nil, fmt.Errorf("资格凭证签名字段无效")
		}
	}
	if !validIdentifier(voucher.CredentialId) || !validIdentifier(voucher.IssuerKeyId) {
		return nil, fmt.Errorf("资格凭证编号无效")
	}
	return []byte(strings.Join(fields, "|")), nil
}

// VerifyVoucher 验证签名、时间、链 ID 与领取地址绑定；金额不属于凭证，完全由链上计算。
func VerifyVoucher(voucher Voucher, authority ClaimAuthority, expectedChainID, expectedAddress string, nowUnix int64, height int64) error {
	if voucher.ChainId != expectedChainID || voucher.ChainId == "" {
		return fmt.Errorf("资格凭证链 ID 不匹配")
	}
	if voucher.Address != expectedAddress || voucher.Address == "" {
		return fmt.Errorf("资格凭证地址不匹配")
	}
	if voucher.ExpiresAtUnix <= nowUnix {
		return fmt.Errorf("资格凭证已过期")
	}
	if voucher.IssuerKeyId != authority.IssuerKeyId || authority.Revoked || height < authority.ActivationHeight {
		return fmt.Errorf("资格签发公钥不可用")
	}
	if err := ValidateAuthority(authority); err != nil {
		return err
	}
	payload, err := CanonicalVoucherBytes(voucher)
	if err != nil {
		return err
	}
	if len(voucher.Signature) != ed25519.SignatureSize || !ed25519.Verify(ed25519.PublicKey(authority.Ed25519PublicKey), payload, voucher.Signature) {
		return fmt.Errorf("资格凭证签名无效")
	}
	return nil
}

// ValidateBasic 在签名验证前尽早拒绝结构不完整的领取消息。
func (msg MsgClaim) ValidateBasic() error {
	if strings.TrimSpace(msg.Claimer) == "" || strings.Contains(msg.Claimer, "|") {
		return fmt.Errorf("领取地址无效")
	}
	if msg.Voucher.Address != msg.Claimer {
		return fmt.Errorf("凭证地址必须等于领取交易签名者")
	}
	if msg.Voucher.ExpiresAtUnix <= 0 {
		return fmt.Errorf("凭证过期时间无效")
	}
	if _, err := CanonicalVoucherBytes(msg.Voucher); err != nil {
		return err
	}
	if len(msg.Voucher.Signature) != ed25519.SignatureSize {
		return fmt.Errorf("资格凭证签名长度无效")
	}
	return nil
}
