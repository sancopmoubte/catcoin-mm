.PHONY: chain-test claim-e2e static-test static-build

# 候选链测试不创建正式创世或主网进程。
chain-test:
	$(MAKE) -C chain/catcoin-app test

claim-e2e:
	bash chain/catcoin-app/scripts/test-claim-e2e.sh

# 静态 PWA 不依赖数据库或 Node 服务端，仅输出可部署的前端文件。
static-test:
	cd wallet-static && pnpm test

static-build:
	cd wallet-static && pnpm build
