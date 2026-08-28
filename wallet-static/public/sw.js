const CACHE = "catcoin-static-wallet-v2";
const BASE_URL = new URL("./", self.location.href);
const ASSETS = ["", "index.html", "manifest.webmanifest", "catcoin-config.js"]
  .map(path => new URL(path, BASE_URL).toString());

self.addEventListener("install", event => {
  event.waitUntil(caches.open(CACHE).then(cache => cache.addAll(ASSETS)));
  self.skipWaiting();
});

self.addEventListener("activate", event => event.waitUntil(self.clients.claim()));

// 钱包源码可离线缓存；RPC、凭证与交易请求绝不缓存，始终由网络实时处理。
self.addEventListener("fetch", event => {
  if (event.request.method !== "GET" || new URL(event.request.url).origin !== self.location.origin) return;
  const url = new URL(event.request.url);
  const isNavigation = event.request.mode === "navigate";
  const isPublicConfig = url.pathname.endsWith("/catcoin-config.js");

  if (isNavigation || isPublicConfig) {
    event.respondWith(
      fetch(event.request)
        .then(response => {
          if (response.ok) caches.open(CACHE).then(cache => cache.put(event.request, response.clone()));
          return response;
        })
        .catch(() => caches.match(isNavigation ? new URL("index.html", BASE_URL).toString() : event.request)),
    );
    return;
  }

  event.respondWith(
    caches.match(event.request).then(hit => hit || fetch(event.request).then(response => {
      if (response.ok) caches.open(CACHE).then(cache => cache.put(event.request, response.clone()));
      return response;
    })),
  );
});
