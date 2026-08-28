// 这是公开网络配置，不是密钥文件。部署者填入自己的 HTTPS 端点后重新加载页面即可。
// 当前候选试用服务可能停止或重置，因此默认不指向任何临时沙盒 URL。
window.CATCOIN_CONFIG = {
  networkLabel: "候选试用网络（非正式资产）",
  chainId: "catcoin-claim-trial-8",
  rpcEndpoint: "",
  // 仅用于 GET /catcoin/claim/v1/pool；可与 RPC 同源，也可使用运营方的 HTTPS API 网关。
  apiEndpoint: "",
  issuerEndpoint: "",
  denom: "umm",
  displaySymbol: "MM",
  addressPrefix: "cosmos"
};
