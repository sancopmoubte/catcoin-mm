package types

import (
	"fmt"
	"strings"

	"crypto/ed25519"
)

// DefaultGenesisState 默认禁用领取资格；正式候选创世必须在受控环境写入真实公钥。
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Authorities:           []ClaimAuthority{},
		Denom:                 Denom,
		FirstClaimUmm:         FirstClaimUmm,
		MaxClaimPerAddressUmm: MaxClaimPerAddress,
		DailyClaimLimitUmm:    DailyClaimLimitUmm,
		DailyDistribution:     DailyDistribution{},
		ClaimIntervalSeconds:  ClaimIntervalSecs,
		Records:               []ClaimRecord{},
		UsedCredentials:       []UsedCredential{},
	}
}

// ValidateGenesis 阻止任何参数漂移、私钥污染、无效公钥和重复状态键。
func ValidateGenesis(state GenesisState) error {
	if state.Denom != Denom {
		return fmt.Errorf("领取基础面额必须为 %s", Denom)
	}
	if state.FirstClaimUmm != FirstClaimUmm {
		return fmt.Errorf("首笔领取必须为 %d umm", FirstClaimUmm)
	}
	if state.MaxClaimPerAddressUmm != MaxClaimPerAddress {
		return fmt.Errorf("单地址累计上限必须为 %d umm", MaxClaimPerAddress)
	}
	if state.DailyClaimLimitUmm != DailyClaimLimitUmm {
		return fmt.Errorf("每日领取上限必须为 %d umm", DailyClaimLimitUmm)
	}
	if state.DailyDistribution.UtcDay < 0 || state.DailyDistribution.DistributedUmm > DailyClaimLimitUmm {
		return fmt.Errorf("每日领取统计无效")
	}
	if state.ClaimIntervalSeconds != ClaimIntervalSecs {
		return fmt.Errorf("领取间隔必须为 %d 秒", ClaimIntervalSecs)
	}

	authorityIDs := map[string]struct{}{}
	for _, authority := range state.Authorities {
		if err := ValidateAuthority(authority); err != nil {
			return err
		}
		if _, exists := authorityIDs[authority.IssuerKeyId]; exists {
			return fmt.Errorf("资格签发公钥编号重复：%s", authority.IssuerKeyId)
		}
		authorityIDs[authority.IssuerKeyId] = struct{}{}
	}

	claimers := map[string]struct{}{}
	for _, record := range state.Records {
		if strings.TrimSpace(record.Claimer) == "" {
			return fmt.Errorf("领取记录缺少地址")
		}
		if record.ClaimedUmm > MaxClaimPerAddress {
			return fmt.Errorf("地址 %s 的领取总额超过上限", record.Claimer)
		}
		if _, exists := claimers[record.Claimer]; exists {
			return fmt.Errorf("领取记录地址重复：%s", record.Claimer)
		}
		claimers[record.Claimer] = struct{}{}
	}

	credentialIDs := map[string]struct{}{}
	for _, used := range state.UsedCredentials {
		if strings.TrimSpace(used.CredentialId) == "" || strings.TrimSpace(used.Claimer) == "" || used.UsedAtUnix <= 0 {
			return fmt.Errorf("已消费资格凭证记录无效")
		}
		if _, exists := credentialIDs[used.CredentialId]; exists {
			return fmt.Errorf("已消费资格凭证编号重复：%s", used.CredentialId)
		}
		credentialIDs[used.CredentialId] = struct{}{}
	}

	return nil
}

// ValidateAuthority 只允许链上持有 32 字节 Ed25519 公钥；任何私钥长度都将被拒绝。
func ValidateAuthority(authority ClaimAuthority) error {
	if !validIdentifier(authority.IssuerKeyId) {
		return fmt.Errorf("资格签发公钥编号无效")
	}
	if len(authority.Ed25519PublicKey) != ed25519.PublicKeySize {
		return fmt.Errorf("资格签发公钥必须为 %d 字节 Ed25519 公钥", ed25519.PublicKeySize)
	}
	if authority.ActivationHeight < 0 {
		return fmt.Errorf("资格签发公钥生效高度不能为负数")
	}
	return nil
}

func validIdentifier(value string) bool {
	if len(value) == 0 || len(value) > 128 || strings.Contains(value, "|") {
		return false
	}
	return strings.TrimSpace(value) == value
}
