package keeper

import (
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"strings"
	"testing"
	"time"

	coreaddress "cosmossdk.io/core/address"
	cosmossdkmath "cosmossdk.io/math"
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/codec"
	addresscodec "github.com/cosmos/cosmos-sdk/codec/address"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	"github.com/cosmos/cosmos-sdk/testutil"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim/types"
)

type memoryAccountKeeper struct {
	moduleAddress sdk.AccAddress
	codec         coreaddress.Codec
}

func (m memoryAccountKeeper) GetModuleAddress(string) sdk.AccAddress { return m.moduleAddress }
func (m memoryAccountKeeper) AddressCodec() coreaddress.Codec        { return m.codec }

type memoryBankKeeper struct {
	balances map[string]map[string]cosmossdkmath.Int
}

func (m *memoryBankKeeper) GetBalance(_ context.Context, account sdk.AccAddress, denom string) sdk.Coin {
	amount := m.balances[account.String()][denom]
	return sdk.NewCoin(denom, amount)
}

func (m *memoryBankKeeper) SendCoinsFromModuleToAccount(_ context.Context, _ string, recipient sdk.AccAddress, coins sdk.Coins) error {
	for _, coin := range coins {
		module := sdk.AccAddress([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}).String()
		if m.balances[module][coin.Denom].LT(coin.Amount) {
			return errClaimCapReached
		}
		m.balances[module][coin.Denom] = m.balances[module][coin.Denom].Sub(coin.Amount)
		if _, exists := m.balances[recipient.String()]; !exists {
			m.balances[recipient.String()] = map[string]cosmossdkmath.Int{}
		}
		current, exists := m.balances[recipient.String()][coin.Denom]
		if !exists {
			current = cosmossdkmath.ZeroInt()
		}
		m.balances[recipient.String()][coin.Denom] = current.Add(coin.Amount)
	}
	return nil
}

func TestClaimWritesDeterministicStateAndRejectsReuse(t *testing.T) {
	t.Helper()
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("claim-test")).Ctx
	ctx = ctx.WithChainID("catcoin-candidate-8")
	firstBlockTime := time.Unix(1_700_000_000, 0).UTC()
	ctx = ctx.WithBlockTime(firstBlockTime).WithBlockHeight(1)

	moduleAddress := sdk.AccAddress([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
	claimer := sdk.AccAddress([]byte{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2})
	addressCodec := addresscodec.NewBech32Codec("cosmos")
	claimerText, err := addressCodec.BytesToString(claimer)
	if err != nil {
		t.Fatal(err)
	}
	_, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	publicKey := privateKey.Public().(ed25519.PublicKey)
	bank := &memoryBankKeeper{balances: map[string]map[string]cosmossdkmath.Int{
		moduleAddress.String(): {types.Denom: cosmossdkmath.NewInt(100)},
	}}
	account := memoryAccountKeeper{moduleAddress: moduleAddress, codec: addressCodec}
	cdc := codec.NewProtoCodec(cdctypes.NewInterfaceRegistry())
	claimKeeper := NewKeeper(runtime.NewKVStoreService(storeKey), cdc, account, bank)
	state := types.DefaultGenesisState()
	state.Authorities = []types.ClaimAuthority{{IssuerKeyId: "issuer-1", Ed25519PublicKey: publicKey, ActivationHeight: 1}}
	if err := claimKeeper.InitGenesis(sdk.WrapSDKContext(ctx), state); err != nil {
		t.Fatal(err)
	}

	firstVoucher := signedVoucher(t, privateKey, ctx.ChainID(), claimerText, "credential-1", firstBlockTime.Add(time.Hour).Unix())
	first, err := claimKeeper.Claim(sdk.WrapSDKContext(ctx), &types.MsgClaim{Claimer: claimerText, Voucher: firstVoucher})
	if err != nil {
		t.Fatal(err)
	}
	if first.ClaimedUmm != 1 || first.TotalClaimedUmm != 1 || first.NextClaimUmm != 2 || first.NextClaimAtUnix != firstBlockTime.Add(24*time.Hour).Unix() {
		t.Fatalf("首次领取响应错误：%+v", first)
	}
	if got := bank.GetBalance(context.Background(), moduleAddress, types.Denom).Amount.Uint64(); got != 99 {
		t.Fatalf("领取池应减少 1，实际为 %d", got)
	}

	earlyVoucher := signedVoucher(t, privateKey, ctx.ChainID(), claimerText, "credential-early", firstBlockTime.Add(time.Hour).Unix())
	if _, err := claimKeeper.Claim(sdk.WrapSDKContext(ctx), &types.MsgClaim{Claimer: claimerText, Voucher: earlyVoucher}); err == nil || !strings.Contains(err.Error(), "时间未到") {
		t.Fatalf("未到 24 小时必须拒绝，实际错误：%v", err)
	}
	used, err := claimKeeper.UsedCredentials.Has(sdk.WrapSDKContext(ctx), "credential-early")
	if err != nil || used {
		t.Fatalf("失败领取不得消费凭证：used=%v err=%v", used, err)
	}

	nextCtx := ctx.WithBlockTime(firstBlockTime.Add(24 * time.Hour)).WithBlockHeight(2)
	secondVoucher := signedVoucher(t, privateKey, nextCtx.ChainID(), claimerText, "credential-2", nextCtx.BlockTime().Add(time.Hour).Unix())
	second, err := claimKeeper.Claim(sdk.WrapSDKContext(nextCtx), &types.MsgClaim{Claimer: claimerText, Voucher: secondVoucher})
	if err != nil {
		t.Fatal(err)
	}
	if second.ClaimedUmm != 2 || second.TotalClaimedUmm != 3 {
		t.Fatalf("第二次领取应翻倍为 2，实际：%+v", second)
	}
	if _, err := claimKeeper.Claim(sdk.WrapSDKContext(nextCtx), &types.MsgClaim{Claimer: claimerText, Voucher: secondVoucher}); err == nil || !strings.Contains(err.Error(), "已被消费") {
		t.Fatalf("重放凭证必须拒绝，实际错误：%v", err)
	}
	if got := bank.GetBalance(context.Background(), moduleAddress, types.Denom).Amount.Uint64(); got != 97 {
		t.Fatalf("领取池应累计减少 3，实际为 %d", got)
	}
}

func signedVoucher(t *testing.T, privateKey ed25519.PrivateKey, chainID, claimer, credentialID string, expiresAt int64) types.Voucher {
	t.Helper()
	voucher := types.Voucher{ChainId: chainID, Address: claimer, CredentialId: credentialID, ExpiresAtUnix: expiresAt, IssuerKeyId: "issuer-1"}
	payload, err := types.CanonicalVoucherBytes(voucher)
	if err != nil {
		t.Fatal(err)
	}
	voucher.Signature = ed25519.Sign(privateKey, payload)
	return voucher
}
