const CACHE = "catcoin-static-wallet-v1";
const ASSETS = ["/", "/index.html", "/manifest.webmanifest", "/catcoin-config.js"];

self.addEventListener("install", event => {
  event.waitUntil(caches.open(CACHE).then(cache => cache.addAll(ASSETS)));
  self.skipWaiting();
});

self.addEventListener("activate", event => event.waitUntil(self.clients.claim()));

// 钱包源码可离线缓存；RPC、凭证与交易请求绝不缓存，始终由网络实时处理。
self.addEventListener("fetch", event => {
  if (event.request.method !== "GET" || new URL(event.request.url).origin !== self.location.origin) return;
  event.respondWith(caches.match(event.request).then(hit => hit || fetch(event.request)));
});
