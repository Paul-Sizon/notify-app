import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  root: '.',
  plugins: [react()],
  build: { outDir: 'dist', emptyOutDir: true },
  server: {
    port: 5173,
    proxy: {
      '/v1': { target: 'https://raspberrypi.taile76757.ts.net', changeOrigin: true, secure: true },
      '/healthz': { target: 'https://raspberrypi.taile76757.ts.net', changeOrigin: true, secure: true },
    },
  },
});
