package keeper

import (
	"context"

	"cosmossdk.io/collections"
	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim/types"
)

// InitGenesis 将经验证的领取参数、授权公钥、历史记录和凭证消费状态写入模块存储。
func (k Keeper) InitGenesis(ctx context.Context, state *types.GenesisState) error {
	if err := types.ValidateGenesis(*state); err != nil {
		return err
	}
	if err := k.GenesisState.Set(ctx, *state); err != nil {
		return err
	}
	for _, authority := range state.Authorities {
		if err := k.Authorities.Set(ctx, authority.IssuerKeyId, authority); err != nil {
			return err
		}
	}
	for _, record := range state.Records {
		if err := k.ClaimRecords.Set(ctx, record.Claimer, record); err != nil {
			return err
		}
	}
	for _, used := range state.UsedCredentials {
		if err := k.UsedCredentials.Set(ctx, used.CredentialId, used); err != nil {
			return err
		}
	}
	return nil
}

// ExportGenesis 完整导出当前领取状态，以便节点恢复时不丢失限额或重放防护。
func (k Keeper) ExportGenesis(ctx context.Context) (*types.GenesisState, error) {
	state, err := k.GetGenesisState(ctx)
	if err != nil {
		return nil, err
	}
	authorities, err := values(k.Authorities, ctx)
	if err != nil {
		return nil, err
	}
	records, err := values(k.ClaimRecords, ctx)
	if err != nil {
		return nil, err
	}
	used, err := values(k.UsedCredentials, ctx)
	if err != nil {
		return nil, err
	}
	state.Authorities = authorities
	state.Records = records
	state.UsedCredentials = used
	return &state, nil
}

func values[V any](collection collections.Map[string, V], ctx context.Context) ([]V, error) {
	iterator, err := collection.Iterate(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer iterator.Close()
	return iterator.Values()
}
