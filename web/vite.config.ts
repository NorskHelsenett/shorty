import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'


export default defineConfig({
 base: "/",
 plugins: [react()],
 preview: {
  port: 5173,
  strictPort: true,
 },
 server: {
  port: 5173,
  strictPort: true,
  host: '0.0.0.0',
  origin: "http://localhost:5173",
  hmr: {
    clientPort: 5173,
    protocol: 'ws',   
  },
  fs: {
    strict: true,
  },
  allowedHosts: process.env.VITE_ALLOWED_HOSTS ? process.env.VITE_ALLOWED_HOSTS.split(',') : ['127.0.0.1', 'localhost'],
 },
});