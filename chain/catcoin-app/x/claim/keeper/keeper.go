package keeper

import (
	"context"
	"errors"
	"fmt"

	"cosmossdk.io/collections"
	store "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim"
	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim/types"
)

// Keeper 保存领取状态。所有集合使用同一模块 KVStore，状态由共识复制。
type Keeper struct {
	Schema          collections.Schema
	ClaimRecords    collections.Map[string, types.ClaimRecord]
	UsedCredentials collections.Map[string, types.UsedCredential]
	Authorities     collections.Map[string, types.ClaimAuthority]
	GenesisState    collections.Item[types.GenesisState]

	bankKeeper    claim.BankKeeper
	accountKeeper claim.AccountKeeper
}

// NewKeeper 只接受已存在的模块账户；claim 账户未获 Minter 权限。
func NewKeeper(storeService store.KVStoreService, cdc codec.BinaryCodec, accountKeeper claim.AccountKeeper, bankKeeper claim.BankKeeper) Keeper {
	if accountKeeper.GetModuleAddress(types.ModuleName) == nil {
		panic("claim 模块账户未初始化")
	}
	sb := collections.NewSchemaBuilder(storeService)
	keeper := Keeper{
		ClaimRecords:    collections.NewMap(sb, types.ClaimRecordsPrefix, "claim_records", collections.StringKey, codec.CollValue[types.ClaimRecord](cdc)),
		UsedCredentials: collections.NewMap(sb, types.UsedCredentialsPrefix, "used_credentials", collections.StringKey, codec.CollValue[types.UsedCredential](cdc)),
		Authorities:     collections.NewMap(sb, types.AuthoritiesPrefix, "authorities", collections.StringKey, codec.CollValue[types.ClaimAuthority](cdc)),
		GenesisState:    collections.NewItem(sb, types.GenesisStateKey, "genesis_state", codec.CollValue[types.GenesisState](cdc)),
		bankKeeper:      bankKeeper,
		accountKeeper:   accountKeeper,
	}
	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	keeper.Schema = schema
	return keeper
}

func (k Keeper) GetClaimRecord(ctx context.Context, claimer string) (types.ClaimRecord, error) {
	record, err := k.ClaimRecords.Get(ctx, claimer)
	if errors.Is(err, collections.ErrNotFound) {
		return types.ClaimRecord{Claimer: claimer}, nil
	}
	return record, err
}

func (k Keeper) GetAuthority(ctx context.Context, issuerKeyID string) (types.ClaimAuthority, error) {
	authority, err := k.Authorities.Get(ctx, issuerKeyID)
	if errors.Is(err, collections.ErrNotFound) {
		return types.ClaimAuthority{}, fmt.Errorf("资格签发公钥不存在")
	}
	return authority, err
}

func (k Keeper) GetGenesisState(ctx context.Context) (types.GenesisState, error) {
	state, err := k.GenesisState.Get(ctx)
	if errors.Is(err, collections.ErrNotFound) {
		return types.GenesisState{}, fmt.Errorf("领取模块参数未初始化")
	}
	return state, err
}

func (k Keeper) PoolBalance(ctx context.Context, denom string) uint64 {
	balance := k.bankKeeper.GetBalance(ctx, k.accountKeeper.GetModuleAddress(types.ModuleName), denom)
	if !balance.Amount.IsUint64() {
		return 0
	}
	return balance.Amount.Uint64()
}
