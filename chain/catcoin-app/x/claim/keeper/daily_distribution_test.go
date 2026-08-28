package keeper

import (
	"errors"
	"strings"
	"testing"

	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim/types"
)

func TestDailyDistributionCapsAndResetsAtUTCDay(t *testing.T) {
	const dayOneUnix int64 = 1_700_000_000
	dayOne := dayOneUnix / secondsPerUTCDay
	limit := types.DailyClaimLimitUmm

	full, err := consumeDailyDistribution(types.DailyDistribution{UtcDay: dayOne, DistributedUmm: limit - 1}, limit, dayOneUnix, 1)
	if err != nil || full.DistributedUmm != limit {
		t.Fatalf("同日最后 1 umm 应到达上限：state=%+v err=%v", full, err)
	}
	if _, err := consumeDailyDistribution(full, limit, dayOneUnix, 1); !errors.Is(err, errDailyClaimCapReached) {
		t.Fatalf("同日超过 210 MM 必须拒绝：%v", err)
	}

	nextDay, err := consumeDailyDistribution(full, limit, dayOneUnix+secondsPerUTCDay, 1)
	if err != nil || nextDay.UtcDay != dayOne+1 || nextDay.DistributedUmm != 1 {
		t.Fatalf("次日必须从 0 重新统计：state=%+v err=%v", nextDay, err)
	}
	if _, err := consumeDailyDistribution(nextDay, limit, dayOneUnix, 1); err == nil || !strings.Contains(err.Error(), "时间回退") {
		t.Fatalf("区块时间回退必须拒绝：%v", err)
	}
}
