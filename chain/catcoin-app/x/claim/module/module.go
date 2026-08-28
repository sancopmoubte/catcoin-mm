package module

import (
	"context"
	"encoding/json"

	gwruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"

	"cosmossdk.io/core/appmodule"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim"
	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim/keeper"
	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim/types"
)

var (
	_ module.AppModuleBasic = AppModuleBasic{}
	_ module.HasGenesis     = AppModule{}
	_ appmodule.AppModule   = AppModule{}
	_ appmodule.HasServices = AppModule{}
)

// AppModuleBasic 处理无需 keeper 的编码、创世和网关注册。
type AppModuleBasic struct{ cdc codec.Codec }

func (AppModuleBasic) Name() string { return types.ModuleName }

func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	types.RegisterLegacyAminoCodec(cdc)
}

func (AppModuleBasic) RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	types.RegisterInterfaces(registry)
}

func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesisState())
}

func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, raw json.RawMessage) error {
	var state types.GenesisState
	if err := cdc.UnmarshalJSON(raw, &state); err != nil {
		return err
	}
	return types.ValidateGenesis(state)
}

func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *gwruntime.ServeMux) {
	if err := types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx)); err != nil {
		panic(err)
	}
}

// AppModule 将领取模块接入 SDK 的消息和查询服务路由。
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

func NewAppModule(cdc codec.Codec, keeper keeper.Keeper) AppModule {
	return AppModule{AppModuleBasic: AppModuleBasic{cdc: cdc}, keeper: keeper}
}

func (AppModule) IsAppModule() {}

func (AppModule) IsOnePerModuleType() {}

func (am AppModule) RegisterServices(registrar grpc.ServiceRegistrar) error {
	types.RegisterMsgServer(registrar, am.keeper)
	types.RegisterQueryServer(registrar, am.keeper)
	return nil
}

func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, raw json.RawMessage) {
	var state types.GenesisState
	cdc.MustUnmarshalJSON(raw, &state)
	if err := am.keeper.InitGenesis(ctx, &state); err != nil {
		panic(err)
	}
}

func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	state, err := am.keeper.ExportGenesis(ctx)
	if err != nil {
		panic(err)
	}
	return cdc.MustMarshalJSON(state)
}

func (AppModule) ConsensusVersion() uint64 { return 1 }

// Keeper 仅供应用接线和链上测试使用，不向钱包暴露任何私钥能力。
func (am AppModule) Keeper() keeper.Keeper { return am.keeper }

var _ claim.BankKeeper
