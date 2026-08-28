import { defineConfig } from "vite";
import react from "@vitejs/plugin-react";

// 仅输出静态文件；不注册 API、数据库、登录回调或服务端中间件。
export default defineConfig({
  plugins: [react()],
  publicDir: "public",
  build: { outDir: "dist", sourcemap: true },
});
