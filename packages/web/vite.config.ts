import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  build: {
    rollupOptions: {
      external: ['./src/gen/indexall/v1/tag_connect', './src/gen/indexall/v1/resource_connect'],
    },
  },
});
