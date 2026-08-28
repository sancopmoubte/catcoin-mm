package types

import "cosmossdk.io/collections"

const (
	// ModuleName 是领取模块和无铸币权限模块账户的唯一名称。
	ModuleName = "claim"
	StoreKey   = ModuleName
	RouterKey  = ModuleName

	Denom                     = "umm"
	FirstClaimUmm      uint64 = 1
	MaxClaimPerAddress uint64 = 100000000
	// DailyClaimLimitUmm 固定每日最多从存量领取池分发 210 MM；不是增发或通胀。
	DailyClaimLimitUmm uint64 = 210 * 100000000
	ClaimIntervalSecs  int64  = 24 * 60 * 60
)

var (
	ClaimRecordsPrefix    = collections.NewPrefix(0)
	UsedCredentialsPrefix = collections.NewPrefix(1)
	AuthoritiesPrefix     = collections.NewPrefix(2)
	GenesisStateKey       = collections.NewPrefix(3)
)
