import { fromBech32 } from "@cosmjs/encoding";

export const ATOMIC_PER_MM = 100_000_000n;
export const DAILY_CLAIM_LIMIT_UMM = 21_000_000_000n;

export type PublicNetworkConfig = {
  networkLabel: string;
  chainId: string;
  rpcEndpoint: string;
  issuerEndpoint: string;
  denom: string;
  displaySymbol: string;
  addressPrefix: string;
};

export const defaultConfig: PublicNetworkConfig = {
  networkLabel: "候选试用网络（非正式资产）",
  chainId: "catcoin-claim-trial-8",
  rpcEndpoint: "",
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

export function isWalletAddress(address: string, prefix: string): boolean {
  try {
    const decoded = fromBech32(address.trim());
    return decoded.prefix === prefix && decoded.data.length === 20;
  } catch { return false; }
}
