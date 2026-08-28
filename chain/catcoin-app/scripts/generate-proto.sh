#!/usr/bin/env bash
# 生成 x/claim 的 Go 类型。输出由 proto 源码决定，不应手工编辑。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SDK_ROOT="${ROOT}/../cosmos-sdk"
GOGO_INCLUDE="$(go env GOPATH)/pkg/mod/github.com/cosmos/gogoproto@v1.7.0"
GATEWAY_ROOT="$(go env GOPATH)/pkg/mod/github.com/grpc-ecosystem/grpc-gateway@v1.16.0"
GATEWAY_INCLUDE="${GATEWAY_ROOT}/third_party/googleapis"
OUT_DIR="${ROOT}/.proto-out"
GENERATED_DIR="${ROOT}/x/claim/types"

if ! command -v protoc >/dev/null 2>&1; then
  echo "未找到 protoc；请先安装 protobuf-compiler。" >&2
  exit 1
fi

if [ ! -x "$(go env GOPATH)/bin/protoc-gen-gogo" ]; then
  echo "未找到 protoc-gen-gogo；请执行：go install github.com/cosmos/gogoproto/protoc-gen-gogo@v1.7.0" >&2
  exit 1
fi

if [ ! -x "$(go env GOPATH)/bin/protoc-gen-grpc-gateway" ]; then
  echo "未找到 protoc-gen-grpc-gateway；请安装与 go.mod 一致的 v1.16.0。" >&2
  exit 1
fi

if [ ! -d "${SDK_ROOT}/proto" ]; then
  echo "未找到已锁定的 Cosmos SDK 子模块；请使用 git clone --recurse-submodules。" >&2
  exit 1
fi

if [ ! -f "${GATEWAY_INCLUDE}/google/api/annotations.proto" ]; then
  echo "未找到 gRPC-Gateway 的 Google API 注释原型。" >&2
  exit 1
fi

rm -rf "${OUT_DIR}"
mkdir -p "${OUT_DIR}"

PATH="$(go env GOPATH)/bin:${PATH}" protoc \
  -I "${ROOT}/proto" \
  -I "${SDK_ROOT}/proto" \
  -I "${GOGO_INCLUDE}" \
  -I "${GATEWAY_INCLUDE}" \
  --gogo_out=plugins=grpc,paths=import:"${OUT_DIR}" \
  --grpc-gateway_out=logtostderr=true,paths=import:"${OUT_DIR}" \
  "${ROOT}/proto/catcoin/claim/v1/tx.proto" \
  "${ROOT}/proto/catcoin/claim/v1/types.proto" \
  "${ROOT}/proto/catcoin/claim/v1/query.proto"

SOURCE_DIR="${OUT_DIR}/github.com/sancopmoubte/my-mobile-chain/chain/catcoin-app/x/claim/types"
if [ ! -d "${SOURCE_DIR}" ]; then
  echo "协议生成输出路径不符合预期。" >&2
  exit 1
fi

find "${GENERATED_DIR}" -maxdepth 1 -type f \( -name '*.pb.go' -o -name '*.pb.gw.go' \) -delete
find "${SOURCE_DIR}" -maxdepth 1 -type f \( -name '*.pb.go' -o -name '*.pb.gw.go' \) -exec mv {} "${GENERATED_DIR}/" \;
rm -rf "${OUT_DIR}"

echo "已生成 x/claim Protobuf Go 源码。"
