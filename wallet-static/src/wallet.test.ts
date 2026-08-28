import { describe, expect, it } from "vitest";
import { DAILY_CLAIM_LIMIT_UMM, defaultConfig, formatMm, isHttpsUrl, isReadyForChain, mmToAtomic } from "./wallet";

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
  });
});
