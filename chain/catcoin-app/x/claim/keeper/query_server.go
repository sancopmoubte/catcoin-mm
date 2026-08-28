package keeper

import (
	"context"
	"errors"

	"cosmossdk.io/collections"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim/types"
)

var _ types.QueryServer = Keeper{}

func (k Keeper) ClaimRecord(goCtx context.Context, request *types.QueryClaimRecordRequest) (*types.QueryClaimRecordResponse, error) {
	if request == nil || request.Claimer == "" {
		return nil, errors.New("领取地址不能为空")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	state, err := k.GetGenesisState(goCtx)
	if err != nil {
		return nil, err
	}
	record, err := k.GetClaimRecord(goCtx, request.Claimer)
	if err != nil {
		return nil, err
	}
	next, nextAt, amountErr := nextAmount(record, state, ctx.BlockTime().Unix())
	capped := errors.Is(amountErr, errClaimCapReached)
	if amountErr != nil && !capped {
		next = 0
	}
	return &types.QueryClaimRecordResponse{
		Record:          &record,
		NextClaimUmm:    next,
		NextClaimAtUnix: nextAt,
		Capped:          capped,
		ClaimedUmm:      record.ClaimedUmm,
		ClaimCount:      record.ClaimCount,
		LastClaimUnix:   record.LastClaimUnix,
	}, nil
}

func (k Keeper) Pool(goCtx context.Context, _ *types.QueryPoolRequest) (*types.QueryPoolResponse, error) {
	state, err := k.GetGenesisState(goCtx)
	if err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	daily, err := currentDailyDistribution(state.DailyDistribution, ctx.BlockTime().Unix())
	if err != nil {
		return nil, err
	}
	if daily.DistributedUmm > state.DailyClaimLimitUmm {
		return nil, errors.New("每日领取统计超过固定上限")
	}
	return &types.QueryPoolResponse{
		Denom:              state.Denom,
		AvailableUmm:       k.PoolBalance(goCtx, state.Denom),
		DailyClaimLimitUmm: state.DailyClaimLimitUmm,
		DailyDistributedUmm: daily.DistributedUmm,
		DailyRemainingUmm:  state.DailyClaimLimitUmm - daily.DistributedUmm,
		UtcDay:             daily.UtcDay,
		NextResetAtUnix:    (daily.UtcDay + 1) * secondsPerUTCDay,
	}, nil
}

// currentDailyDistribution 将历史日统计按当前区块 UTC 日解释为已重置的零分发状态。
// 它只用于查询，不写入任何链上状态；交易仍只能由 consumeDailyDistribution 消费额度。
func currentDailyDistribution(current types.DailyDistribution, nowUnix int64) (types.DailyDistribution, error) {
	if nowUnix < 0 {
		return types.DailyDistribution{}, errors.New("区块时间早于 Unix epoch")
	}
	utcDay := nowUnix / secondsPerUTCDay
	if utcDay < current.UtcDay {
		return types.DailyDistribution{}, errors.New("区块时间回退，拒绝读取领取统计")
	}
	if utcDay > current.UtcDay {
		return types.DailyDistribution{UtcDay: utcDay}, nil
	}
	return current, nil
}

func (k Keeper) Authority(goCtx context.Context, request *types.QueryAuthorityRequest) (*types.QueryAuthorityResponse, error) {
	if request == nil || request.IssuerKeyId == "" {
		return nil, errors.New("资格签发公钥编号不能为空")
	}
	authority, err := k.Authorities.Get(goCtx, request.IssuerKeyId)
	if errors.Is(err, collections.ErrNotFound) {
		return nil, errors.New("资格签发公钥不存在")
	}
	if err != nil {
		return nil, err
	}
	return &types.QueryAuthorityResponse{Authority: &authority}, nil
}
