package keeper

import (
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

func TestPoolQueryReturnsCurrentUTCDailyDistribution(t *testing.T) {
	storeKey := storetypes.NewKVStoreKey(types.StoreKey)
	ctx := testutil.DefaultContextWithDB(t, storeKey, storetypes.NewTransientStoreKey("claim-query-test")).Ctx
	ctx = ctx.WithBlockTime(time.Unix(1_700_000_000, 0).UTC())
	moduleAddress := sdk.AccAddress([]byte{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1})
	addressCodec := addresscodec.NewBech32Codec("cosmos")
	bank := &memoryBankKeeper{balances: map[string]map[string]cosmossdkmath.Int{
		moduleAddress.String(): {types.Denom: cosmossdkmath.NewInt(777)},
	}}
	account := memoryAccountKeeper{moduleAddress: moduleAddress, codec: coreaddress.Codec(addressCodec)}
	claimKeeper := NewKeeper(runtime.NewKVStoreService(storeKey), codec.NewProtoCodec(cdctypes.NewInterfaceRegistry()), account, bank)
	state := types.DefaultGenesisState()
	state.DailyDistribution = types.DailyDistribution{UtcDay: ctx.BlockTime().Unix() / secondsPerUTCDay, DistributedUmm: 123}
	if err := claimKeeper.InitGenesis(sdk.WrapSDKContext(ctx), state); err != nil {
		t.Fatal(err)
	}

	response, err := claimKeeper.Pool(sdk.WrapSDKContext(ctx), &types.QueryPoolRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if response.AvailableUmm != 777 || response.DailyDistributedUmm != 123 || response.DailyClaimLimitUmm != types.DailyClaimLimitUmm {
		t.Fatalf("领取池查询值错误：%+v", response)
	}
	if response.DailyRemainingUmm != types.DailyClaimLimitUmm-123 || response.UtcDay != state.DailyDistribution.UtcDay {
		t.Fatalf("领取池每日剩余额度错误：%+v", response)
	}
	if response.NextResetAtUnix != (response.UtcDay+1)*secondsPerUTCDay {
		t.Fatalf("下次 UTC 重置时间错误：%d", response.NextResetAtUnix)
	}
}

func TestPoolQueryResetsHistoricalDailyDistributionWithoutWriting(t *testing.T) {
	current := types.DailyDistribution{UtcDay: 10, DistributedUmm: 99}
	next, err := currentDailyDistribution(current, 11*secondsPerUTCDay)
	if err != nil {
		t.Fatal(err)
	}
	if next.UtcDay != 11 || next.DistributedUmm != 0 {
		t.Fatalf("跨 UTC 日查询必须显示零分发，实际：%+v", next)
	}
	if _, err := currentDailyDistribution(current, 9*secondsPerUTCDay); err == nil {
		t.Fatal("区块时间回退必须拒绝读取历史日统计")
	}
}
