package main

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim/types"
)

// 这是离线资格签发辅助程序。正式签发私钥应置于独立受控系统，不能放入节点、网页或 Git 仓库。
func main() {
	if len(os.Args) < 2 {
		fail("用法：claimvoucher <module-address|generate|sign>")
	}
	switch os.Args[1] {
	case "module-address":
		fmt.Println(authtypes.NewModuleAddress(types.ModuleName).String())
	case "generate":
		generate(os.Args[2:])
	case "sign":
		sign(os.Args[2:])
	default:
		fail("未知子命令：%s", os.Args[1])
	}
}

func generate(args []string) {
	flags := flag.NewFlagSet("generate", flag.ExitOnError)
	privateKeyPath := flags.String("private-key-file", "", "仅本机保存的私钥文件路径")
	flags.Parse(args)
	if *privateKeyPath == "" {
		fail("必须提供 --private-key-file")
	}
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		fail("生成 Ed25519 密钥失败：%v", err)
	}
	if err := os.MkdirAll(filepath.Dir(*privateKeyPath), 0o700); err != nil {
		fail("创建私钥目录失败：%v", err)
	}
	if err := os.WriteFile(*privateKeyPath, []byte(base64.StdEncoding.EncodeToString(privateKey)), 0o600); err != nil {
		fail("写入私钥失败：%v", err)
	}
	printJSON(map[string]string{"ed25519_public_key": base64.StdEncoding.EncodeToString(publicKey)})
}

func sign(args []string) {
	flags := flag.NewFlagSet("sign", flag.ExitOnError)
	privateKeyPath := flags.String("private-key-file", "", "本机私钥文件路径")
	chainID := flags.String("chain-id", "", "候选链 ID")
	claimer := flags.String("claimer", "", "领取地址")
	credentialID := flags.String("credential-id", "", "一次性资格编号")
	expiresAt := flags.Int64("expires-at-unix", 0, "Unix 秒级过期时间")
	issuerKeyID := flags.String("issuer-key-id", "", "链上资格签发公钥编号")
	flags.Parse(args)
	if *privateKeyPath == "" || *chainID == "" || *claimer == "" || *credentialID == "" || *expiresAt == 0 || *issuerKeyID == "" {
		fail("sign 缺少必填参数")
	}
	encoded, err := os.ReadFile(*privateKeyPath)
	if err != nil {
		fail("读取本机私钥失败：%v", err)
	}
	privateKey, err := base64.StdEncoding.DecodeString(string(encoded))
	if err != nil || len(privateKey) != ed25519.PrivateKeySize {
		fail("私钥文件不是有效的 Ed25519 私钥")
	}
	voucher := types.Voucher{
		ChainId:       *chainID,
		Address:       *claimer,
		CredentialId:  *credentialID,
		ExpiresAtUnix: *expiresAt,
		IssuerKeyId:   *issuerKeyID,
	}
	payload, err := types.CanonicalVoucherBytes(voucher)
	if err != nil {
		fail("构造规范凭证载荷失败：%v", err)
	}
	voucher.Signature = ed25519.Sign(ed25519.PrivateKey(privateKey), payload)
	printJSON(voucher)
}

func printJSON(value any) {
	data, err := json.Marshal(value)
	if err != nil {
		fail("编码 JSON 失败：%v", err)
	}
	fmt.Println(string(data))
}

func fail(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if errors.New(message) != nil {
		fmt.Fprintln(os.Stderr, message)
	}
	os.Exit(1)
}
