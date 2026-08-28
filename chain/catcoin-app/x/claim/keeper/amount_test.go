package keeper

import (
	"errors"
	"testing"

	"github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim/types"
)

func TestNextAmountCapsExactlyAtOneMM(t *testing.T) {
	state := *types.DefaultGenesisState()
	record := types.ClaimRecord{Claimer: "cosmos1test"}
	var total uint64
	var now int64 = 1_700_000_000

	for i := 0; i < 27; i++ {
		amount, _, err := nextAmount(record, state, now)
		if err != nil {
			t.Fatalf("第 %d 笔不应失败：%v", i+1, err)
		}
		if i == 26 && amount != 32_891_137 {
			t.Fatalf("第 27 笔应截断为 32891137 umm，实际为 %d", amount)
		}
		total += amount
		record.ClaimedUmm += amount
		record.ClaimCount++
		record.LastClaimUnix = now
		now += types.ClaimIntervalSecs
	}
	if total != types.MaxClaimPerAddress || record.ClaimedUmm != types.MaxClaimPerAddress {
		t.Fatalf("累计必须恰为 1 MM：total=%d record=%d", total, record.ClaimedUmm)
	}
	if _, _, err := nextAmount(record, state, now); !errors.Is(err, errClaimCapReached) {
		t.Fatalf("第 28 笔必须永久拒绝，实际错误：%v", err)
	}
}
