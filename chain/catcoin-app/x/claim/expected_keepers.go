package claim

import (
	"context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper 只暴露领取池转账与余额查询；没有任何 MintCoins 权限或接口。
type BankKeeper interface {
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
}

// AccountKeeper 只用于确认 claim 模块账户存在及校验 Bech32 地址。
type AccountKeeper interface {
	GetModuleAddress(name string) sdk.AccAddress
	AddressCodec() address.Codec
}
