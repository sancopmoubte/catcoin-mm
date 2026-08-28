import { fromBech32 } from "@cosmjs/encoding";

export const ATOMIC_PER_MM = 100_000_000n;
export const DAILY_CLAIM_LIMIT_UMM = 21_000_000_000n;

export type PublicNetworkConfig = {
  networkLabel: string;
  chainId: string;
  rpcEndpoint: string;
  apiEndpoint: string;
  issuerEndpoint: string;
  denom: string;
  displaySymbol: string;
  addressPrefix: string;
};

export const defaultConfig: PublicNetworkConfig = {
  networkLabel: "候选试用网络（非正式资产）",
  chainId: "catcoin-claim-trial-8",
  rpcEndpoint: "",
  apiEndpoint: "",
  issuerEndpoint: "",
  denom: "umm",
  displaySymbol: "MM",
  addressPrefix: "cosmos",
};

export function formatMm(amount: string | bigint): string {
  const raw = BigInt(amount);
  const whole = raw / ATOMIC_PER_MM;
  const fraction = (raw % ATOMIC_PER_MM).toString().padStart(8, "0").replace(/0+$/, "");
  return fraction ? `${whole}.${fraction}` : whole.toString();
}

export function mmToAtomic(input: string): string {
  const normalized = input.trim();
  if (!/^\d+(\.\d{1,8})?$/.test(normalized)) throw new Error("金额须为大于 0 且最多 8 位小数的 MM 数字");
  const [whole, fraction = ""] = normalized.split(".");
  const value = BigInt(whole) * ATOMIC_PER_MM + BigInt((fraction + "00000000").slice(0, 8));
  if (value <= 0n) throw new Error("金额必须大于 0");
  return value.toString();
}

export function isHttpsUrl(value: string): boolean {
  try { return new URL(value).protocol === "https:"; } catch { return false; }
}

export function isReadyForChain(config: PublicNetworkConfig): boolean {
  return Boolean(config.chainId.trim() && config.denom.trim() && isHttpsUrl(config.rpcEndpoint));
}

export function isReadyForPoolQuery(config: PublicNetworkConfig): boolean {
  return isHttpsUrl(config.apiEndpoint);
}

export type ClaimPoolStatus = {
  denom: string;
  availableUmm: string;
  dailyClaimLimitUmm: string;
  dailyDistributedUmm: string;
  dailyRemainingUmm: string;
  utcDay: string;
  nextResetAtUnix: string;
};

function unsignedInteger(value: unknown, field: string): string {
  const normalized = typeof value === "number" ? String(value) : typeof value === "string" ? value : "";
  if (!/^\d+$/.test(normalized)) throw new Error(`领取池接口的 ${field} 不是无符号整数`);
  return normalized;
}

// gRPC-Gateway 的 JSON 会把 uint64/int64 编码为字符串；同时接受数值以兼容受控网关。
export function parseClaimPoolStatus(value: unknown): ClaimPoolStatus {
  if (!value || typeof value !== "object") throw new Error("领取池接口未返回对象");
  const source = value as Record<string, unknown>;
  const read = (camel: string, snake: string) => source[camel] ?? source[snake];
  const denom = String(source.denom ?? "").trim();
  if (!denom) throw new Error("领取池接口未返回 denom");
  return {
    denom,
    availableUmm: unsignedInteger(read("availableUmm", "available_umm"), "available_umm"),
    dailyClaimLimitUmm: unsignedInteger(read("dailyClaimLimitUmm", "daily_claim_limit_umm"), "daily_claim_limit_umm"),
    dailyDistributedUmm: unsignedInteger(read("dailyDistributedUmm", "daily_distributed_umm"), "daily_distributed_umm"),
    dailyRemainingUmm: unsignedInteger(read("dailyRemainingUmm", "daily_remaining_umm"), "daily_remaining_umm"),
    utcDay: unsignedInteger(read("utcDay", "utc_day"), "utc_day"),
    nextResetAtUnix: unsignedInteger(read("nextResetAtUnix", "next_reset_at_unix"), "next_reset_at_unix"),
  };
}

export function isWalletAddress(address: string, prefix: string): boolean {
  try {
    const decoded = fromBech32(address.trim());
    return decoded.prefix === prefix && decoded.data.length === 20;
  } catch { return false; }
}
