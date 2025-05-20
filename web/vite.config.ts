/// <reference types="vite/client" />
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'


function getEnv(variable: string | undefined, name: string, defaultValue: string[]): string[] {
    if (!variable){
        console.warn(`Enveroment variable ${name} is not set; using standard value: ${defaultValue}`)
        return defaultValue;
    }
    return variable.split(',');
}

const ALLOWED_HOSTS = getEnv(import.meta.env.VITE_ALLOWED_HOSTS, "VITE_ALLOWED_HOSTS", ['127.0.0.1', 'localhost']);

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
  allowedHosts: ALLOWED_HOSTS
 },
});