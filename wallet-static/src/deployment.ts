/** 将 Vite 的根路径规范为可用于 PWA 资源与 Service Worker scope 的绝对目录路径。 */
export function normalizeBasePath(basePath: string): string {
  const withLeadingSlash = basePath.startsWith("/") ? basePath : `/${basePath}`;
  return withLeadingSlash.endsWith("/") ? withLeadingSlash : `${withLeadingSlash}/`;
}

export const appBasePath = normalizeBasePath(import.meta.env.BASE_URL);
