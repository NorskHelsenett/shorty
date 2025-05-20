/// <reference types="vite/client" />
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react-swc'
import * as process from 'process'

function getEnv(variable: string | undefined, name: string, defaultValue: string[]): string[] {
    if (!variable){
        console.warn(`Environment variable ${name} is not set; using standard value: ${defaultValue}`)
        return defaultValue;
    }
    console.info(`Environment variable ${name} is set; using value: ${variable.split(',')}`)
    return variable.split(',');
}

export default defineConfig(({ mode }) => {
  // Load env files based on mode (.env, .env.production, etc)
  const env = loadEnv(mode, process.cwd(), '')
  
  // Get allowed hosts from environment variable or use default
  const ALLOWED_HOSTS = getEnv(process.env.VITE_ALLOWED_HOSTS || env.VITE_ALLOWED_HOSTS, "VITE_ALLOWED_HOSTS", ['127.0.0.1', 'localhost'])

  return {
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
  }
});