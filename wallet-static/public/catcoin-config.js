// 这是公开网络配置，不是密钥文件。仅填写临时候选节点的公开 HTTPS 端点。

// Codespaces 是临时云端开发电脑；停止或重建后，下面的端点会失效或变化。

window.CATCOIN_CONFIG = {
  
  networkLabel: "catcoin-pos-1 临时 Codespace 开发网（非正式资产）",
  
  chainId: "catcoin-pos-1",
  
  rpcEndpoint: "https://shiny-broccoli-gxqvr7j4qvwv29vpw-26657.app.github.dev",
  
  // 节点启用 REST API 后用于查询领取池；未提供凭证服务时领取按钮不会伪造成功。
  
  apiEndpoint: "https://shiny-broccoli-gxqvr7j4qvwv29vpw-1317.app.github.dev",
  
  issuerEndpoint: "",
  
  denom: "umm",
  
  displaySymbol: "MM",
  
  addressPrefix: "cosmos"
    
};











