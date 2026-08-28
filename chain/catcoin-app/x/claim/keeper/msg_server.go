package keeper

import (
	"context"
	"errors"
	"fmt"

	cosmossdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim/types"
)

var _ types.MsgServer = Keeper{}

// Claim 执行完整原子领取：校验资格、检查重放与额度、转账、再写入记录；任一步失败均不提交缓存状态。
func (k Keeper) Claim(goCtx context.Context, msg *types.MsgClaim) (*types.MsgClaimResponse, error) {
	if msg == nil {
		return nil, fmt.Errorf("领取消息不能为空")
	}
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	sdkCtx := sdk.UnwrapSDKContext(goCtx)
	claimer, err := k.accountKeeper.AddressCodec().StringToBytes(msg.Claimer)
	if err != nil {
		return nil, fmt.Errorf("领取地址编码无效: %w", err)
	}
	cacheCtx, write := sdkCtx.CacheContext()
	ctx := sdk.WrapSDKContext(cacheCtx)

	state, err := k.GetGenesisState(ctx)
	if err != nil {
		return nil, err
	}
	authority, err := k.GetAuthority(ctx, msg.Voucher.IssuerKeyId)
	if err != nil {
		return nil, err
	}
	if err := types.VerifyVoucher(msg.Voucher, authority, cacheCtx.ChainID(), msg.Claimer, cacheCtx.BlockTime().Unix(), cacheCtx.BlockHeight()); err != nil {
		return nil, err
	}
	used, err := k.UsedCredentials.Has(ctx, msg.Voucher.CredentialId)
	if err != nil {
		return nil, err
	}
	if used {
		return nil, fmt.Errorf("资格凭证已被消费")
	}

	record, err := k.GetClaimRecord(ctx, msg.Claimer)
	if err != nil {
		return nil, err
	}
	amount, nextAt, err := nextAmount(record, state, cacheCtx.BlockTime().Unix())
	if err != nil {
		return nil, err
	}
	daily, err := consumeDailyDistribution(state.DailyDistribution, state.DailyClaimLimitUmm, cacheCtx.BlockTime().Unix(), amount)
	if err != nil {
		return nil, err
	}
	if k.PoolBalance(ctx, state.Denom) < amount {
		return nil, fmt.Errorf("领取池余额不足")
	}

	coins := sdk.NewCoins(sdk.NewCoin(state.Denom, cosmossdkmath.NewIntFromUint64(amount)))
	if err := k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, sdk.AccAddress(claimer), coins); err != nil {
		return nil, err
	}
	record.Claimer = msg.Claimer
	record.ClaimedUmm += amount
	record.ClaimCount++
	record.LastClaimUnix = cacheCtx.BlockTime().Unix()
	if err := k.ClaimRecords.Set(ctx, msg.Claimer, record); err != nil {
		return nil, err
	}
	if err := k.UsedCredentials.Set(ctx, msg.Voucher.CredentialId, types.UsedCredential{
		CredentialId: msg.Voucher.CredentialId,
		Claimer:      msg.Claimer,
		UsedAtUnix:   cacheCtx.BlockTime().Unix(),
	}); err != nil {
		return nil, err
	}
	state.DailyDistribution = daily
	if err := k.GenesisState.Set(ctx, state); err != nil {
		return nil, err
	}

	next, _, nextErr := nextAmount(record, state, record.LastClaimUnix+state.ClaimIntervalSeconds)
	if nextErr != nil && !errors.Is(nextErr, errClaimCapReached) {
		return nil, nextErr
	}
	if !errors.Is(nextErr, errClaimCapReached) {
		nextAt = record.LastClaimUnix + state.ClaimIntervalSeconds
	}
	write()
	return &types.MsgClaimResponse{
		ClaimedUmm:      amount,
		TotalClaimedUmm: record.ClaimedUmm,
		NextClaimUmm:    next,
		NextClaimAtUnix: nextAt,
	}, nil
}

var errClaimCapReached = errors.New("地址领取上限已达到")
var errDailyClaimCapReached = errors.New("今日全网领取额度已用完")

const secondsPerUTCDay int64 = 24 * 60 * 60

// consumeDailyDistribution 在交易缓存上下文中消费每日全网额度。
// 当日额度不足时整笔领取失败，外层不会调用 write，因此凭证、余额和统计均不会被提交。
func consumeDailyDistribution(current types.DailyDistribution, limit uint64, nowUnix int64, amount uint64) (types.DailyDistribution, error) {
	if nowUnix < 0 {
		return types.DailyDistribution{}, fmt.Errorf("区块时间早于 Unix epoch")
	}
	utcDay := nowUnix / secondsPerUTCDay
	if utcDay < current.UtcDay {
		return types.DailyDistribution{}, fmt.Errorf("区块时间回退，拒绝领取")
	}
	if utcDay > current.UtcDay {
		current = types.DailyDistribution{UtcDay: utcDay}
	}
	if current.DistributedUmm > limit || amount > limit-current.DistributedUmm {
		return types.DailyDistribution{}, errDailyClaimCapReached
	}
	current.DistributedUmm += amount
	return current, nil
}

// nextAmount 只用无符号整数计算，末笔自动截断到剩余额度，永不溢出或超过 1 MM。
func nextAmount(record types.ClaimRecord, state types.GenesisState, nowUnix int64) (uint64, int64, error) {
	if record.ClaimedUmm >= state.MaxClaimPerAddressUmm {
		return 0, 0, errClaimCapReached
	}
	if record.ClaimCount > 0 {
		nextAt := record.LastClaimUnix + state.ClaimIntervalSeconds
		if nowUnix < nextAt {
			return 0, nextAt, fmt.Errorf("下次领取时间未到")
		}
	}
	remaining := state.MaxClaimPerAddressUmm - record.ClaimedUmm
	amount := state.FirstClaimUmm
	for count := uint64(0); count < record.ClaimCount; count++ {
		if amount >= remaining || amount > state.MaxClaimPerAddressUmm/2 {
			return remaining, 0, nil
		}
		amount *= 2
	}
	if amount > remaining {
		amount = remaining
	}
	return amount, 0, nil
}
