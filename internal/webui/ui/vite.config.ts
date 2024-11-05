import { svelte } from "@sveltejs/vite-plugin-svelte";
import { defineConfig } from "vite";

// https://vite.dev/config/
export default defineConfig({
  plugins: [svelte()],
  server: {
    proxy: {
      "/events": {
        target: "http://localhost:9876",
        changeOrigin: true,
      },
    },
  },
});
