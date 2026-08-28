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
	return &types.QueryPoolResponse{Denom: state.Denom, AvailableUmm: k.PoolBalance(goCtx, state.Denom)}, nil
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
