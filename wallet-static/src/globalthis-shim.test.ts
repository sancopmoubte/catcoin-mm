import { expect, test } from "vitest";
// CommonJS 垫片必须维持 xstream 所需的函数和具名属性。
declare const require: (moduleId: string) => {
  (): typeof globalThis;
  getPolyfill: () => typeof globalThis;
  implementation: typeof globalThis;
  shim: () => typeof globalThis;
};

const globalThisFactory = require("./globalthis-shim.cjs");

test("浏览器 globalThis 垫片保留 xstream 所需导出", () => {
  expect(globalThisFactory()).toBe(globalThis);
  expect(globalThisFactory.getPolyfill()).toBe(globalThis);
  expect(globalThisFactory.implementation).toBe(globalThis);
  expect(globalThisFactory.shim()).toBe(globalThis);
});
