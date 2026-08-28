# 猫猫币 MM 静态 PWA 钱包

这是一个**纯静态** React/Vite PWA。它不包含数据库、登录、Node/Express 服务端、私钥 API、签发私钥、验证者密钥或任何水龙头助记词。钱包由浏览器本地创建和加密；余额查询和交易广播直接连接部署者填入的 HTTPS RPC。

> 当前默认配置没有指向临时沙盒端点。请在 `public/catcoin-config.js` 中填入你运营的 HTTPS RPC、用于读取领取池的 HTTPS 链 API 与（可选）受控 HTTPS 凭证服务。候选试用 MM 不是正式资产。

## 本地构建

```bash
cd wallet-static
pnpm install
pnpm test
pnpm build
```

产物仅在 `dist/`。可将该目录部署到任何支持 HTTPS 的静态文件托管。`apiEndpoint` 只用于读取候选链接口 `GET /catcoin/claim/v1/pool`，该接口由 `x/claim` 的 gRPC-Gateway 提供，可与 RPC 同源或由独立 HTTPS API 网关提供。用于免费领取的 `issuerEndpoint` 必须是独立服务：它持有私钥并签发短期、地址绑定凭证；**绝不能**把签发私钥放入本 PWA。

## GitHub Pages 静态部署

公开仓库的 `.github/workflows/deploy-pages.yml` 会在 `main` 分支的 `wallet-static/` 文件变更后构建本目录，并发布到 GitHub Pages 的仓库子路径。启用 Pages 的 Actions 来源后，固定入口为 <https://sancopmoubte.github.io/catcoin-mm/>，不依赖 Manus 沙盒。构建时会将资源、配置文件、PWA 启动范围和 Service Worker 自动置于 `/catcoin-mm/` 路径；本地构建仍保持根路径。

该静态站点默认不配置 RPC 或凭证端点。沙盒回收后，GitHub Pages 页面和已缓存的离线界面仍可打开；实时余额、转账广播和受控领取仍分别需要你自行运行的 HTTPS RPC 与凭证服务。

## 功能边界

| 功能 | 静态 PWA | 外部依赖 |
|---|---|---|
| 创建/解锁本机钱包、显示恢复短语 | 支持 | 无 |
| 本地签名、余额查询、转账广播 | 支持 | HTTPS RPC |
| 免费领取 | 支持前端入口 | HTTPS 签发服务 + 链上 `x/claim` |
| 每日 210 MM 领取统计、领取池余额 | 读取/显示 | `apiEndpoint` 的 HTTPS 链 API；PWA 不能自行记账 |
| 共识、公开 RPC、签发私钥、验证者 | 不支持 | 必须由独立长期服务运行 |

自写代码采用 Apache-2.0。
