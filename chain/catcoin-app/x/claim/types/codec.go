package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RegisterInterfaces 注册唯一可广播的领取消息，确保交易通过 SDK 的消息服务路由执行。
func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations((*sdk.Msg)(nil), &MsgClaim{})
}

// RegisterLegacyAminoCodec 保持为空：新模块只使用 Protobuf 交易编码。
func RegisterLegacyAminoCodec(_ *codec.LegacyAmino) {}
