.PHONY: proto build test clean

# 先生成消息与查询类型，再构建；该流程不依赖 Docker 或 buf。
proto:
	bash ./scripts/generate-proto.sh

build: proto
	mkdir -p ./build
	go build -o ./build/catcoind ./cmd/catcoind
	go build -o ./build/claimvoucher ./cmd/claimvoucher

test: proto
	go test ./...

clean:
	rm -rf ./build ./x/claim/types/*.pb.go
