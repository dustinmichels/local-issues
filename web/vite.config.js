import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

export default defineConfig({
  plugins: [vue(), tailwindcss()],
  build: {
    emptyOutDir: false,
  },
  server: {
    proxy: {
      "/api": "http://localhost:7890",
    },
  },
});
