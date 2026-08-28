import { describe, expect, it } from "vitest";
import { normalizeBasePath } from "./deployment";

describe("GitHub Pages 静态部署配置", () => {
  it("将 GitHub Pages 仓库子路径规范化为 Service Worker 可用范围", () => {
    expect(normalizeBasePath("/catcoin-mm/")).toBe("/catcoin-mm/");
    expect(normalizeBasePath("catcoin-mm")).toBe("/catcoin-mm/");
    expect(normalizeBasePath("/")).toBe("/");
  });

  it("拒绝空路径，避免 Service Worker 注册到相对页面路径", () => {
    expect(normalizeBasePath("")).toBe("/");
  });
});
