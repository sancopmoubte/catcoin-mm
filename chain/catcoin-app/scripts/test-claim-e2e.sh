#!/usr/bin/env bash
# 在隔离的一验证者候选链上验证真实 MsgClaim 路由；不是正式网络，也不使用任何正式私钥。
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BIN="${ROOT}/build/catcoind"
VOUCHER_TOOL="${ROOT}/build/claimvoucher"
HOME_DIR="${ROOT}/.claim-e2e"
CHAIN_ID="catcoin-claim-e2e-8"
RPC_PORT=27657
RPC="tcp://127.0.0.1:${RPC_PORT}"
ISSUER_DIR="$(mktemp -d)"
NODE_PID=""

cleanup() {
  if [ -n "${NODE_PID}" ] && kill -0 "${NODE_PID}" 2>/dev/null; then
    kill "${NODE_PID}" || true
    wait "${NODE_PID}" || true
  fi
  rm -rf "${ISSUER_DIR}"
}
trap cleanup EXIT

fail() {
  echo "FAIL: $*" >&2
  exit 1
}

wait_for_height() {
  for _ in $(seq 1 40); do
    if curl -fsS "http://127.0.0.1:${RPC_PORT}/status" >/tmp/catcoin-claim-status.json 2>/dev/null; then
      height="$(jq -r '.result.sync_info.latest_block_height' /tmp/catcoin-claim-status.json)"
      if [ "${height}" -ge 2 ]; then
        return 0
      fi
    fi
    sleep 1
  done
  fail "候选节点未在时限内出块"
}

expect_rejected() {
  local name="$1"
  shift
  "$@" >/tmp/catcoin-claim-rejected.json 2>&1 || return 0
  local txhash
  txhash="$(jq -r '.txhash // empty' /tmp/catcoin-claim-rejected.json)"
  [ -n "${txhash}" ] || fail "${name} 没有返回可查询交易哈希"
  for _ in $(seq 1 15); do
    if curl -fsS "http://127.0.0.1:${RPC_PORT}/tx?hash=0x${txhash}" >/tmp/catcoin-claim-deliver.json 2>/dev/null; then
      code="$(jq -r '.result.tx_result.code' /tmp/catcoin-claim-deliver.json)"
      [ "${code}" != "0" ] && return 0
      cat /tmp/catcoin-claim-deliver.json >&2
      fail "${name} 意外在链上成功"
    fi
    sleep 1
  done
  fail "${name} 未在时限内得到链上拒绝结果"
}

cd "${ROOT}"
GOMAXPROCS=2 make build >/tmp/catcoin-claim-build.log
rm -rf "${HOME_DIR}"

"${BIN}" init claim-e2e --chain-id "${CHAIN_ID}" --home "${HOME_DIR}" >/tmp/catcoin-claim-init.log 2>&1
"${BIN}" keys add validator --keyring-backend test --home "${HOME_DIR}" >/tmp/catcoin-claim-validator.json 2>&1
VALIDATOR="$("${BIN}" keys show validator --address --keyring-backend test --home "${HOME_DIR}")"
POOL_ADDRESS="$("${VOUCHER_TOOL}" module-address)"

# 该隔离 E2E 网只保留一个 1 MM 启动验证者和其余领取池；总量仍精确为 21,000,000 MM。
"${BIN}" genesis add-genesis-account "${VALIDATOR}" 100000000umm --keyring-backend test --home "${HOME_DIR}" >/tmp/catcoin-claim-add-validator.log 2>&1
"${BIN}" genesis add-genesis-account "${POOL_ADDRESS}" 2099999900000000umm --home "${HOME_DIR}" >/tmp/catcoin-claim-add-pool.log 2>&1
"${BIN}" genesis gentx validator 100000000umm --chain-id "${CHAIN_ID}" --keyring-backend test --home "${HOME_DIR}" >/tmp/catcoin-claim-gentx.log 2>&1
"${BIN}" genesis collect-gentxs --home "${HOME_DIR}" >/tmp/catcoin-claim-collect-gentxs.log 2>&1

ISSUER_JSON="$("${VOUCHER_TOOL}" generate --private-key-file "${ISSUER_DIR}/issuer.key")"
ISSUER_PUBLIC_KEY="$(printf '%s' "${ISSUER_JSON}" | jq -r '.ed25519_public_key')"
GENESIS="${HOME_DIR}/config/genesis.json"
tmp_genesis="${GENESIS}.tmp"

# 写入测试用公钥（私钥仅在 mktemp 目录，退出时删除），并固定候选经济参数。
jq --arg issuerPublicKey "${ISSUER_PUBLIC_KEY}" '
  .app_state.claim.authorities = [{
    issuer_key_id: "e2e-issuer-1",
    ed25519_public_key: $issuerPublicKey,
    revoked: false,
    activation_height: "1"
  }]
  | .app_state.bank.denom_metadata = [{
    description: "Catcoin candidate test denomination",
    denom_units: [
      {denom: "umm", exponent: 0, aliases: ["uMM"]},
      {denom: "mmc", exponent: 8, aliases: ["MM"]}
    ],
    base: "umm",
    display: "mmc",
    name: "Catcoin",
    symbol: "MM",
    uri: "",
    uri_hash: ""
  }]
  | .app_state.staking.params.bond_denom = "umm"
  | .app_state.gov.params.min_deposit = [{denom: "umm", amount: "100000000"}]
  | .app_state.gov.params.expedited_min_deposit = [{denom: "umm", amount: "500000000"}]
  | del(.app_state.mint)
' "${GENESIS}" >"${tmp_genesis}"
mv "${tmp_genesis}" "${GENESIS}"
"${BIN}" genesis validate-genesis --home "${HOME_DIR}" >/tmp/catcoin-claim-validate-genesis.log 2>&1
jq -e '
  (.app_state | has("mint") | not)
  and (.app_state | has("crisis") | not)
  and (.app_state.bank.denom_metadata[0].base == "umm")
  and (.app_state.bank.denom_metadata[0].display == "mmc")
  and (.app_state.bank.denom_metadata[0].denom_units[] | select(.denom == "mmc") | .exponent == 8)
  and (.app_state.staking.params.bond_denom == "umm")
  and (.app_state.gov.params.min_deposit[0].denom == "umm")
' "${GENESIS}" >/dev/null || fail "候选创世未落实无铸币、无危机费、8 位 umm 或 PoS 治理面额约束"

sed -i -E "s#^laddr = \"tcp://127\.0\.0\.1:26657\"#laddr = \"tcp://127.0.0.1:${RPC_PORT}\"#" "${HOME_DIR}/config/config.toml"
sed -i -E 's#^timeout_commit = "[^"]+"#timeout_commit = "1s"#' "${HOME_DIR}/config/config.toml"
sed -i -E 's#^minimum-gas-prices = "[^"]+"#minimum-gas-prices = "0umm"#' "${HOME_DIR}/config/app.toml"

"${BIN}" start --home "${HOME_DIR}" --minimum-gas-prices 0umm >/tmp/catcoin-claim-node.log 2>&1 &
NODE_PID="$!"
wait_for_height

NOW="$(date +%s)"
VOUCHER="$(${VOUCHER_TOOL} sign --private-key-file "${ISSUER_DIR}/issuer.key" --chain-id "${CHAIN_ID}" --claimer "${VALIDATOR}" --credential-id "credential-1" --expires-at-unix "$((NOW + 600))" --issuer-key-id e2e-issuer-1)"
"${BIN}" tx claim claim --voucher "${VOUCHER}" --from validator --keyring-backend test --chain-id "${CHAIN_ID}" --home "${HOME_DIR}" --node "${RPC}" --yes --broadcast-mode sync --output json >/tmp/catcoin-claim-success.json
sleep 3

"${BIN}" query claim claim-record --claimer "${VALIDATOR}" --home "${HOME_DIR}" --node "${RPC}" --output json >/tmp/catcoin-claim-record.json
"${BIN}" query claim pool --home "${HOME_DIR}" --node "${RPC}" --output json >/tmp/catcoin-claim-pool.json
"${BIN}" query bank total --home "${HOME_DIR}" --node "${RPC}" --output json >/tmp/catcoin-claim-total.json

jq -e '.claimed_umm == "1" and .claim_count == "1" and (.last_claim_unix | tonumber) > 0 and (.next_claim_umm // "0") == "0"' /tmp/catcoin-claim-record.json >/dev/null || fail "首次领取状态不正确"
jq -e '.available_umm == "2099999899999999" and .denom == "umm"' /tmp/catcoin-claim-pool.json >/dev/null || fail "领取池余额未按 1 uMM 减少"
jq -e '[.supply[] | select(.denom == "umm") | .amount] | add == "2100000000000000"' /tmp/catcoin-claim-total.json >/dev/null || fail "总供应量不守恒"

EARLY_VOUCHER="$(${VOUCHER_TOOL} sign --private-key-file "${ISSUER_DIR}/issuer.key" --chain-id "${CHAIN_ID}" --claimer "${VALIDATOR}" --credential-id "credential-early" --expires-at-unix "$((NOW + 600))" --issuer-key-id e2e-issuer-1)"
expect_rejected "未满 24 小时的领取" "${BIN}" tx claim claim --voucher "${EARLY_VOUCHER}" --from validator --keyring-backend test --chain-id "${CHAIN_ID}" --home "${HOME_DIR}" --node "${RPC}" --yes --broadcast-mode sync --output json

WRONG_CHAIN_VOUCHER="$(${VOUCHER_TOOL} sign --private-key-file "${ISSUER_DIR}/issuer.key" --chain-id "wrong-chain" --claimer "${VALIDATOR}" --credential-id "credential-wrong-chain" --expires-at-unix "$((NOW + 600))" --issuer-key-id e2e-issuer-1)"
expect_rejected "跨链凭证" "${BIN}" tx claim claim --voucher "${WRONG_CHAIN_VOUCHER}" --from validator --keyring-backend test --chain-id "${CHAIN_ID}" --home "${HOME_DIR}" --node "${RPC}" --yes --broadcast-mode sync --output json

expect_rejected "重复凭证" "${BIN}" tx claim claim --voucher "${VOUCHER}" --from validator --keyring-backend test --chain-id "${CHAIN_ID}" --home "${HOME_DIR}" --node "${RPC}" --yes --broadcast-mode sync --output json

# 重启节点后复用已消费凭证仍必须拒绝，证明防重放状态已写入 IAVL，而非仅驻留进程内存。
kill "${NODE_PID}"
wait "${NODE_PID}" || true
NODE_PID=""
"${BIN}" start --home "${HOME_DIR}" --minimum-gas-prices 0umm >/tmp/catcoin-claim-node-restarted.log 2>&1 &
NODE_PID="$!"
wait_for_height
expect_rejected "重启后的重复凭证" "${BIN}" tx claim claim --voucher "${VOUCHER}" --from validator --keyring-backend test --chain-id "${CHAIN_ID}" --home "${HOME_DIR}" --node "${RPC}" --yes --broadcast-mode sync --output json
"${BIN}" query claim claim-record --claimer "${VALIDATOR}" --home "${HOME_DIR}" --node "${RPC}" --output json >/tmp/catcoin-claim-record-restarted.json
jq -e '.claimed_umm == "1" and .claim_count == "1"' /tmp/catcoin-claim-record-restarted.json >/dev/null || fail "节点重启后领取记录未持久化"

echo "PASS: 隔离候选链已通过真实 MsgClaim 首领、链上记录、领取池扣减、固定供应、提前拒绝、跨链拒绝和凭证重放拒绝验证。"
