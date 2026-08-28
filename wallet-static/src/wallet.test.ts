import { describe, expect, it } from "vitest";
import { DAILY_CLAIM_LIMIT_UMM, defaultConfig, formatMm, isHttpsUrl, isReadyForChain, isReadyForPoolQuery, mmToAtomic, parseClaimPoolStatus } from "./wallet";

describe("静态钱包公开配置与金额", () => {
  it("以 8 位精度处理 MM，且每日额度只是固定领取池分发速率", () => {
    expect(mmToAtomic("0.00000001")).toBe("1");
    expect(mmToAtomic("210")).toBe(DAILY_CLAIM_LIMIT_UMM.toString());
    expect(formatMm(DAILY_CLAIM_LIMIT_UMM)).toBe("210");
  });

  it("拒绝非 HTTPS RPC，并要求部署者显式提供可用公开端点", () => {
    expect(isHttpsUrl("http://example.test")).toBe(false);
    expect(isHttpsUrl("https://rpc.example.test")).toBe(true);
    expect(isReadyForChain(defaultConfig)).toBe(false);
    expect(isReadyForChain({ ...defaultConfig, rpcEndpoint: "https://rpc.example.test" })).toBe(true);
    expect(isReadyForPoolQuery(defaultConfig)).toBe(false);
    expect(isReadyForPoolQuery({ ...defaultConfig, apiEndpoint: "https://api.example.test" })).toBe(true);
  });

  it("解析真实领取池查询，兼容 gRPC-Gateway 的下划线 JSON 字段", () => {
    expect(parseClaimPoolStatus({
      denom: "umm", available_umm: "2099999700000000", daily_claim_limit_umm: "21000000000",
      daily_distributed_umm: "123", daily_remaining_umm: "20999999877", utc_day: "19675", next_reset_at_unix: "1700006400",
    })).toMatchObject({ denom: "umm", dailyDistributedUmm: "123", dailyRemainingUmm: "20999999877" });
    expect(() => parseClaimPoolStatus({ denom: "umm", availableUmm: "not-a-number" })).toThrow("available_umm");
  });
});
