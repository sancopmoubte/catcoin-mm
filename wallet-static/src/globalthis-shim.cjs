// xstream 通过 require("globalthis") 读取这三个导出。此模块只返回浏览器
// 原生 globalThis，不注入 Node.js polyfill，也不读取或保存钱包数据。
function getPolyfill() {
  return globalThis;
}

function shim() {
  return globalThis;
}

function globalThisFactory() {
  return globalThis;
}

globalThisFactory.getPolyfill = getPolyfill;
globalThisFactory.implementation = globalThis;
globalThisFactory.shim = shim;

module.exports = globalThisFactory;
