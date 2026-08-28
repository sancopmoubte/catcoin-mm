import { defineConfig, loadEnv } from "vite";
import react from "@vitejs/plugin-react";

// 仅输出静态文件；不注册 API、数据库、登录回调或服务端中间件。
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, ".", "VITE_");
  return {
    plugins: [react()],
    publicDir: "public",
    // GitHub Pages 项目站点位于 /catcoin-mm/；本地预览保持根路径。
    base: env.VITE_BASE_PATH || "/",
    build: { outDir: "dist", sourcemap: true },
  };
});
