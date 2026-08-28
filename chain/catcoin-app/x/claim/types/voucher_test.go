package types

import (
	"crypto/ed25519"
	"crypto/rand"
	"testing"
)

func TestVerifyVoucherRejectsCrossChainTamperingAndExpiry(t *testing.T) {
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	authority := ClaimAuthority{IssuerKeyId: "issuer-1", Ed25519PublicKey: publicKey, ActivationHeight: 1}
	voucher := Voucher{ChainId: "catcoin-candidate-8", Address: "cosmos1testaddress", CredentialId: "credential-1", ExpiresAtUnix: 2_000, IssuerKeyId: authority.IssuerKeyId}
	payload, err := CanonicalVoucherBytes(voucher)
	if err != nil {
		t.Fatal(err)
	}
	voucher.Signature = ed25519.Sign(privateKey, payload)

	if err := VerifyVoucher(voucher, authority, "catcoin-candidate-8", voucher.Address, 1_000, 1); err != nil {
		t.Fatalf("有效凭证被拒绝：%v", err)
	}
	if err := VerifyVoucher(voucher, authority, "other-chain", voucher.Address, 1_000, 1); err == nil {
		t.Fatal("跨链凭证必须拒绝")
	}
	tampered := voucher
	tampered.Address = "cosmos1anotheraddress"
	if err := VerifyVoucher(tampered, authority, "catcoin-candidate-8", tampered.Address, 1_000, 1); err == nil {
		t.Fatal("篡改地址后的凭证必须拒绝")
	}
	if err := VerifyVoucher(voucher, authority, "catcoin-candidate-8", voucher.Address, 2_000, 1); err == nil {
		t.Fatal("到期凭证必须拒绝")
	}
}
