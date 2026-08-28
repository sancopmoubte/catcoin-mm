# 猫猫币（MM）候选链与静态 PWA

这是猫猫币（MM）的**公开候选源码**：Cosmos SDK + CometBFT PoS 应用层、`x/claim` 领取状态机、Rust 轻客户端核心，以及可在 iPhone Safari 中安装的**纯静态 PWA 钱包**。自写代码采用 Apache-2.0。

> **不是正式已发行网络，也不是投资或资产托管服务。** 本仓库不含正式 chain ID、正式创世、分配地址、验证者密钥、钱包助记词、节点状态、签发私钥或任何长期 RPC/凭证服务。发布源码不等于授权启动正式网络。

## 已包含内容

| 目录 | 内容 | 运行边界 |
|---|---|---|
| `chain/catcoin-app` | Cosmos SDK v0.50 候选应用、CometBFT PoS 接线、`x/claim`、Protobuf 和隔离测试 | 仅本地/隔离候选测试；不附创世密钥或运行数据 |
| `wallet-static` | React/Vite 静态 PWA、本机加密钱包、直连 HTTPS RPC、转账、领取和领取池状态读取入口 | 不使用数据库、登录、Express、服务端 API 或硬编码密钥 |
| `light-client` | Rust 轻客户端核心源码 | 仅源代码；不承诺浏览器 PWA 已完成独立轻客户端验证 |
| `docs` | 钻石皮书和领取模块设计 | 描述候选协议与正式发布门槛 |

## 候选规则

| 项目 | 当前候选规则 |
|---|---|
| 共识 | Cosmos SDK + CometBFT PoS |
| 总供应 | 固定 `21,000,000 MM`，即 `2,100,000,000,000,000 umm` |
| 精度 | `1 MM = 100,000,000 umm` |
| 铸币 | 无 `x/mint` 自动增发路径；领取仅从创世存量领取池转账 |
| 免费领取 | 首笔 `1 umm`，每 24 小时翻倍，每地址累计最高 1 MM |
| 防批量 | 地址绑定 Ed25519 一次性凭证、凭证重放防护、每 UTC 日最多从领取池分发 210 MM |
| 手续费/自动收益 | 候选规则为 0 MM 手续费、无自动质押收益 |

“每日 210 MM”是**固定领取池分发速率上限**，不是每日新增 210 MM；它不会改变固定总供应。

## 静态 PWA

```bash
cd wallet-static
pnpm install
pnpm test
pnpm build
```

将 `wallet-static/dist/` 部署到任意 HTTPS 静态托管即可。部署前只需编辑公开文件 `wallet-static/public/catcoin-config.js`，填入你运营的 HTTPS RPC、可选 HTTPS 链 API 和（如启用免费领取）独立 HTTPS 凭证服务地址。HTTPS 链 API 的 `GET /catcoin/claim/v1/pool` 由候选 `x/claim` gRPC-Gateway 返回领取池余额、当前 UTC 日已分发/剩余额度和下次重置时间；PWA 不会自行计算或伪造这些链上数据。**不得**把私钥、助记词、签发密钥或主机密码放进此文件或浏览器代码。

本仓库配置了 GitHub Pages 自动部署；启用 Pages 的 Actions 来源后，静态入口为 <https://sancopmoubte.github.io/catcoin-mm/>。该入口不依赖 Manus 沙盒，且默认不填临时 RPC、链 API 或凭证端点。网页可长期托管和离线打开，但实时链上能力仍取决于独立运营的 HTTPS RPC、链 API 与凭证服务。

## 候选链本地测试

候选应用的 Protobuf 生成流程依赖同级的 Cosmos SDK v0.50.15 源码。先在 `chain/` 下准备该公开上游依赖：

```bash
git clone --branch v0.50.15 --depth 1 https://github.com/cosmos/cosmos-sdk.git chain/cosmos-sdk
cd chain/catcoin-app
make test
bash scripts/test-claim-e2e.sh
```

隔离 E2E 会临时生成测试密钥、启动本地节点并在完成后清理；它不生成正式创世或正式资产。

## 正式发布前的阻塞条件

正式网络仍需要唯一 chain ID、经多方复核的创世分配、至少三名独立验证者和持久基础设施、受保护 RPC、备份恢复演练、签发/隐私/申诉制度、独立安全审计，以及发行方对不可逆创世的单独明确确认。详细边界见 [钻石皮书](docs/catcoin-diamond-paper.md)。
