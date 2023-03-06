import { defineConfig } from "vite";
import { nodeResolve } from "@rollup/plugin-node-resolve";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import * as path from "path";

export default defineConfig({
  plugins: [svelte()],
  // https://github.com/vitejs/vite/issues/7385#issuecomment-1286606298
  resolve: {
    alias: {
      "#src": path.resolve(__dirname, "./src/"),
      "#lib": path.resolve(__dirname, "./src/lib/"),
    },
  },
});
