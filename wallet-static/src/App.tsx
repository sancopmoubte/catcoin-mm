import { DirectSecp256k1HdWallet, Registry } from "@cosmjs/proto-signing";
import { coin, defaultRegistryTypes, SigningStargateClient, StargateClient } from "@cosmjs/stargate";
import { Any } from "cosmjs-types/google/protobuf/any";
import { AuthInfo, TxBody, TxRaw } from "cosmjs-types/cosmos/tx/v1beta1/tx";
import { BinaryWriter } from "cosmjs-types/binary";
import { FormEvent, useMemo, useState } from "react";
import { defaultConfig, formatMm, isReadyForChain, isWalletAddress, mmToAtomic, PublicNetworkConfig } from "./wallet";

declare global {
  interface Window { CATCOIN_CONFIG?: Partial<PublicNetworkConfig>; }
}

type StoredWallet = { address: string; encrypted: string };
type Voucher = { chainId: string; address: string; credentialId: string; expiresAtUnix: string | number; issuerKeyId: string; signature: string };
type Activity = { title: string; detail: string; txHash?: string };

const STORAGE_KEY = "catcoin-static-wallet-v1";
const MSG_CLAIM_TYPE_URL = "/catcoin.claim.v1.MsgClaim";
const ZERO_FEE = { amount: [], gas: "250000" };

function readStoredWallet(): StoredWallet | null {
  try { const raw = localStorage.getItem(STORAGE_KEY); return raw ? JSON.parse(raw) as StoredWallet : null; } catch { return null; }
}

function readPublicConfig(): PublicNetworkConfig {
  const configured = window.CATCOIN_CONFIG && typeof window.CATCOIN_CONFIG === "object" ? window.CATCOIN_CONFIG : {};
  return { ...defaultConfig, ...configured };
}

function b64ToBytes(value: string): Uint8Array {
  const binary = atob(value);
  return Uint8Array.from(binary, letter => letter.charCodeAt(0));
}

function encodeVoucher(voucher: Voucher, writer: BinaryWriter): void {
  if (voucher.chainId) writer.uint32(10).string(voucher.chainId);
  if (voucher.address) writer.uint32(18).string(voucher.address);
  if (voucher.credentialId) writer.uint32(26).string(voucher.credentialId);
  if (BigInt(voucher.expiresAtUnix) !== 0n) writer.uint32(32).int64(BigInt(voucher.expiresAtUnix));
  if (voucher.issuerKeyId) writer.uint32(42).string(voucher.issuerKeyId);
  const signature = voucher.signature ? b64ToBytes(voucher.signature) : new Uint8Array();
  if (signature.length) writer.uint32(50).bytes(signature);
}

const MsgClaim = {
  encode(value: { claimer: string; voucher: Voucher }, writer: BinaryWriter = BinaryWriter.create()): BinaryWriter {
    writer.uint32(10).string(value.claimer);
    writer.uint32(18).fork(); encodeVoucher(value.voucher, writer); writer.ldelim();
    return writer;
  },
  fromPartial(value: Partial<{ claimer: string; voucher: Voucher }>) {
    if (!value.claimer || !value.voucher) throw new Error("领取交易缺少地址或凭证");
    return value as { claimer: string; voucher: Voucher };
  },
};

function registry() { return new Registry([...defaultRegistryTypes, [MSG_CLAIM_TYPE_URL, MsgClaim as never]]); }

function normalizeVoucher(raw: Record<string, unknown>): Voucher {
  return {
    chainId: String(raw.chainId ?? raw.chain_id ?? ""), address: String(raw.address ?? ""),
    credentialId: String(raw.credentialId ?? raw.credential_id ?? ""), expiresAtUnix: String(raw.expiresAtUnix ?? raw.expires_at_unix ?? "0"),
    issuerKeyId: String(raw.issuerKeyId ?? raw.issuer_key_id ?? ""), signature: String(raw.signature ?? ""),
  };
}

async function requestVoucher(address: string, config: PublicNetworkConfig): Promise<Voucher> {
  if (!config.issuerEndpoint.startsWith("https://")) throw new Error("免费领取需要配置独立的 HTTPS 凭证服务");
  const response = await fetch(`${config.issuerEndpoint.replace(/\/$/, "")}/v1/trial-voucher`, {
    method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ address }),
  });
  const body = await response.json().catch(() => ({})) as { voucher?: Record<string, unknown>; message?: string; error?: string };
  if (!response.ok || !body.voucher) throw new Error(body.message || body.error || "凭证服务未返回有效凭证");
  const voucher = normalizeVoucher(body.voucher);
  if (voucher.chainId !== config.chainId || voucher.address !== address || !voucher.signature) throw new Error("凭证与当前网络或本机地址不匹配");
  return voucher;
}

async function claim(address: string, voucher: Voucher, config: PublicNetworkConfig): Promise<string> {
  const client = await StargateClient.connect(config.rpcEndpoint);
  try {
    const claimRegistry = registry();
    const value = { claimer: address, voucher };
    const message = Any.fromPartial({ typeUrl: MSG_CLAIM_TYPE_URL, value: claimRegistry.encode({ typeUrl: MSG_CLAIM_TYPE_URL, value }) });
    const bodyBytes = TxBody.encode(TxBody.fromPartial({ messages: [message], memo: "Catcoin static PWA candidate claim" })).finish();
    const authInfoBytes = AuthInfo.encode(AuthInfo.fromPartial({ signerInfos: [] })).finish();
    const result = await client.broadcastTx(TxRaw.encode(TxRaw.fromPartial({ bodyBytes, authInfoBytes, signatures: [] })).finish());
    if (result.code !== 0) throw new Error(result.rawLog || `领取交易失败，代码 ${result.code}`);
    return result.transactionHash;
  } finally { client.disconnect(); }
}

export default function App() {
  const [config, setConfig] = useState<PublicNetworkConfig>(readPublicConfig);
  const [storedWallet, setStoredWallet] = useState<StoredWallet | null>(readStoredWallet);
  const [wallet, setWallet] = useState<DirectSecp256k1HdWallet | null>(null);
  const [password, setPassword] = useState("");
  const [mnemonic, setMnemonic] = useState("");
  const [balance, setBalance] = useState("0");
  const [height, setHeight] = useState("—");
  const [recipient, setRecipient] = useState("");
  const [amount, setAmount] = useState("0.00000001");
  const [notice, setNotice] = useState("请先配置 HTTPS RPC，再创建或解锁钱包。");
  const [busy, setBusy] = useState<"create" | "unlock" | "refresh" | "claim" | "send" | null>(null);
  const [activity, setActivity] = useState<Activity[]>([]);

  const ready = useMemo(() => isReadyForChain(config), [config]);
  const address = storedWallet?.address;
  const unlocked = Boolean(wallet && address);
  const updateConfig = (key: keyof PublicNetworkConfig, value: string) => setConfig(current => ({ ...current, [key]: value.trim() }));
  const addActivity = (item: Activity) => setActivity(current => [item, ...current].slice(0, 5));

  async function refresh() {
    if (!ready) { setNotice("请在下方填入你的 HTTPS RPC 地址和正确 chain ID。静态网页不能自行提供节点。"); return; }
    setBusy("refresh");
    try {
      const endpoint = config.rpcEndpoint.replace(/\/$/, "");
      const statusResponse = await fetch(`${endpoint}/status`);
      if (!statusResponse.ok) throw new Error("RPC 状态接口不可用");
      const status = await statusResponse.json();
      const network = String(status?.result?.node_info?.network ?? "");
      if (network !== config.chainId) throw new Error(`网络不匹配：节点为 ${network || "未知"}`);
      setHeight(String(status?.result?.sync_info?.latest_block_height ?? "—"));
      if (address) {
        const client = await StargateClient.connect(endpoint);
        try { setBalance((await client.getBalance(address, config.denom)).amount || "0"); } finally { client.disconnect(); }
      }
      setNotice("已从配置的候选链 RPC 读取状态。");
    } catch (error) { setNotice(error instanceof Error ? error.message : "读取候选链失败"); } finally { setBusy(null); }
  }

  async function createWallet(event: FormEvent) {
    event.preventDefault();
    if (password.length < 8) { setNotice("本机钱包密码至少需要 8 位。"); return; }
    setBusy("create");
    try {
      const nextWallet = await DirectSecp256k1HdWallet.generate(24, { prefix: config.addressPrefix });
      const [account] = await nextWallet.getAccounts();
      const nextStored = { address: account.address, encrypted: await nextWallet.serialize(password) };
      localStorage.setItem(STORAGE_KEY, JSON.stringify(nextStored));
      setStoredWallet(nextStored); setWallet(nextWallet); setMnemonic(nextWallet.mnemonic); setPassword("");
      setNotice("本机钱包已创建。请离线抄写恢复短语；关闭后本页不会再显示它。");
      await refresh();
    } catch (error) { setNotice(error instanceof Error ? error.message : "创建钱包失败"); } finally { setBusy(null); }
  }

  async function unlockWallet(event: FormEvent) {
    event.preventDefault();
    if (!storedWallet) return;
    setBusy("unlock");
    try {
      setWallet(await DirectSecp256k1HdWallet.deserialize(storedWallet.encrypted, password));
      setPassword(""); setNotice("本机钱包已解锁。"); await refresh();
    } catch { setNotice("密码不正确或本机钱包数据损坏。"); } finally { setBusy(null); }
  }

  async function claimCandidate() {
    if (!wallet || !address) { setNotice("请先创建或解锁本机钱包。"); return; }
    if (!ready) { setNotice("请先配置候选链 HTTPS RPC。"); return; }
    setBusy("claim");
    try {
      const voucher = await requestVoucher(address, config);
      const txHash = await claim(address, voucher, config);
      addActivity({ title: "候选试用领取已广播", detail: "额度由链上地址、凭证与每日限制共同决定", txHash });
      setNotice("领取交易已写入候选链。请刷新余额确认。"); await refresh();
    } catch (error) { setNotice(error instanceof Error ? error.message : "领取失败"); } finally { setBusy(null); }
  }

  async function send(event: FormEvent) {
    event.preventDefault();
    if (!wallet || !address) { setNotice("请先解锁本机钱包。"); return; }
    if (!ready) { setNotice("请先配置候选链 HTTPS RPC。"); return; }
    if (!isWalletAddress(recipient, config.addressPrefix)) { setNotice(`收款地址必须是有效的 ${config.addressPrefix}1… 地址。`); return; }
    setBusy("send");
    try {
      const client = await SigningStargateClient.connectWithSigner(config.rpcEndpoint, wallet);
      try {
        const result = await client.sendTokens(address, recipient, [coin(mmToAtomic(amount), config.denom)], ZERO_FEE, "Catcoin static PWA transfer");
        if (result.code !== 0) throw new Error(result.rawLog || `转账失败，代码 ${result.code}`);
        addActivity({ title: "本机签名转账已广播", detail: `发送 ${amount} ${config.displaySymbol}；链规则为 0 MM 费用`, txHash: result.transactionHash });
        setRecipient(""); setNotice("转账交易已广播，请刷新余额确认。");
      } finally { client.disconnect(); }
      await refresh();
    } catch (error) { setNotice(error instanceof Error ? error.message : "转账失败"); } finally { setBusy(null); }
  }

  return <main className="shell">
    <section className="app" aria-label="猫猫币静态 PWA 钱包">
      <header><div><span className="brand-mark">M</span><p>猫猫币 <small>MM · 纯静态 PWA 钱包</small></p></div><span className={ready ? "pill ready" : "pill"}>{ready ? "RPC 已配置" : "等待 RPC 配置"}</span></header>
      <section className="warning"><strong>候选试用网络，不是正式资产。</strong> 静态网页只在本机签名并直连你配置的 HTTPS 服务；不会保存、上传或托管你的钱包私钥。</section>
      <section className="balance-card"><span>本机钱包余额</span><h1>{formatMm(balance)} <small>{config.displaySymbol}</small></h1><p>{address || "创建或解锁钱包后显示地址"}</p><button className="outline" onClick={() => void refresh()} disabled={busy !== null}>{busy === "refresh" ? "正在读取…" : "刷新链上余额"}</button></section>

      <section className="card config-card"><h2>1. 公开网络配置</h2><p>此处只填公开 HTTPS 地址，不填助记词、私钥或密码。候选节点停机时，钱包仍可离线保留，但不能查询或广播。</p><div className="fields"><label>网络名称<input value={config.networkLabel} onChange={event => updateConfig("networkLabel", event.target.value)} /></label><label>Chain ID<input value={config.chainId} onChange={event => updateConfig("chainId", event.target.value)} /></label><label>HTTPS RPC<input inputMode="url" placeholder="https://rpc.example.com" value={config.rpcEndpoint} onChange={event => updateConfig("rpcEndpoint", event.target.value)} /></label><label>HTTPS 凭证服务（领取可选）<input inputMode="url" placeholder="https://issuer.example.com" value={config.issuerEndpoint} onChange={event => updateConfig("issuerEndpoint", event.target.value)} /></label></div></section>

      {!storedWallet ? <section className="card"><h2>2. 创建本机钱包</h2><p>钱包只加密保存在此浏览器。本页没有登录、数据库或后台钱包托管。</p><form onSubmit={createWallet}><label>设置本机钱包密码<input type="password" autoComplete="new-password" minLength={8} value={password} onChange={event => setPassword(event.target.value)} placeholder="至少 8 位" required /></label><button disabled={busy !== null}>{busy === "create" ? "正在创建…" : "创建并解锁"}</button></form></section> : !unlocked ? <section className="card"><h2>2. 解锁本机钱包</h2><p className="address">{address}</p><form onSubmit={unlockWallet}><label>本机钱包密码<input type="password" autoComplete="current-password" value={password} onChange={event => setPassword(event.target.value)} required /></label><button disabled={busy !== null}>{busy === "unlock" ? "正在解锁…" : "解锁钱包"}</button></form></section> : <>
        <section className="card"><h2>2. 领取候选试用 MM</h2><p>链上规则：首笔 0.00000001 MM、每 24 小时递增、每地址最多 1 MM；全网每个 UTC 日最多从固定领取池分发 210 MM。领取需要独立 HTTPS 凭证服务，额度和防批量规则由链执行。</p><button onClick={() => void claimCandidate()} disabled={busy !== null || !config.issuerEndpoint}>{busy === "claim" ? "正在领取…" : "一键领取候选试用 MM"}</button></section>
        <section className="card"><h2>3. 本机签名转账</h2><p>私钥不离开当前设备。交易直接发送至配置的 RPC，候选规则为 0 MM 手续费。</p><form onSubmit={send}><label>收款地址<input value={recipient} onChange={event => setRecipient(event.target.value)} placeholder={`${config.addressPrefix}1…`} required /></label><label>金额（MM）<input inputMode="decimal" value={amount} onChange={event => setAmount(event.target.value)} required /></label><button disabled={busy !== null || !ready}>{busy === "send" ? "正在签名并广播…" : "签名并发送"}</button></form></section>
      </>}

      {mnemonic && <section className="recovery"><h2>立即离线抄写恢复短语</h2><code>{mnemonic}</code><p>不要截图、不要发送给任何人。关闭后本页不会再次显示。</p><button className="outline" onClick={() => setMnemonic("")}>我已安全抄写</button></section>}
      <section className="status"><strong>状态：</strong>{notice}<span>区块高度：{height} · 候选每日领取上限：210 MM（固定池分发，不增发）</span></section>
      {activity.length > 0 && <section className="card"><h2>本机操作记录</h2>{activity.map(item => <article className="activity" key={`${item.title}-${item.txHash}`}><strong>{item.title}</strong><p>{item.detail}</p>{item.txHash && <code>{item.txHash}</code>}</article>)}</section>}
      <footer>纯静态构建 · 本地钱包 · 公开 RPC · 不含数据库、登录、服务端 API 或任何硬编码密钥</footer>
    </section>
  </main>;
}
