import { defineConfig, loadEnv } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), '');
  const apiBase = env.API_BASE;
  if (!apiBase) {
    throw new Error('API_BASE env var is required (e.g. https://host.example.ts.net)');
  }
  return {
    root: '.',
    plugins: [react()],
    build: { outDir: 'dist', emptyOutDir: true },
    server: {
      port: 5173,
      proxy: {
        '/v1': { target: apiBase, changeOrigin: true, secure: true },
        '/healthz': { target: apiBase, changeOrigin: true, secure: true },
      },
    },
  };
});
