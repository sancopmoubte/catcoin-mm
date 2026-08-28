# 猫猫币 MM 静态 PWA 钱包

这是一个**纯静态** React/Vite PWA。它不包含数据库、登录、Node/Express 服务端、私钥 API、签发私钥、验证者密钥或任何水龙头助记词。钱包由浏览器本地创建和加密；余额查询和交易广播直接连接部署者填入的 HTTPS RPC。

> 当前默认配置没有指向临时沙盒端点。请在 `public/catcoin-config.js` 中填入你运营的 HTTPS RPC 与（可选）受控 HTTPS 凭证服务。候选试用 MM 不是正式资产。

## 本地构建

```bash
cd wallet-static
pnpm install
pnpm test
pnpm build
```

产物仅在 `dist/`。可将该目录部署到任何支持 HTTPS 的静态文件托管。用于免费领取的 `issuerEndpoint` 必须是独立服务：它持有私钥并签发短期、地址绑定凭证；**绝不能**把签发私钥放入本 PWA。

## 功能边界

| 功能 | 静态 PWA | 外部依赖 |
|---|---|---|
| 创建/解锁本机钱包、显示恢复短语 | 支持 | 无 |
| 本地签名、余额查询、转账广播 | 支持 | HTTPS RPC |
| 免费领取 | 支持前端入口 | HTTPS 签发服务 + 链上 `x/claim` |
| 每日 210 MM 领取统计 | 读取/显示 | 链上 Query；PWA 不能自行记账 |
| 共识、公开 RPC、签发私钥、验证者 | 不支持 | 必须由独立长期服务运行 |

自写代码采用 Apache-2.0。
