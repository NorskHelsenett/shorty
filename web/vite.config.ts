/// <reference types="vite/client" />
import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react-swc'
import * as process from 'process'
import fs from 'fs'
import path from 'path'

// Function to get array environment variables
function getEnvArray(variable: string | undefined, name: string, defaultValue: string[]): string[] {
    if (!variable){
        console.warn(`Environment variable ${name} is not set; using standard value: ${defaultValue}`)
        return defaultValue;
    }
    console.info(`Environment variable ${name} is set; using value: ${variable.split(',')}`)
    return variable.split(',');
}

// Function to get string environment variables
function getEnvString(variable: string | undefined, name: string, defaultValue: string): string {
    if (!variable){
        console.warn(`Environment variable ${name} is not set; using standard value: ${defaultValue}`)
        return defaultValue;
    }
    console.info(`Environment variable ${name} is set; using value: ${variable}`)
    return variable;
}

// Generate runtime config file when running build
function generateRuntimeConfig(env: Record<string, string>) {
  // Create runtime config content with default or configured values
  const runtimeConfig = {
    AUTH_URL: getEnvString(env.VITE_AUTH_URL, "VITE_AUTH_URL", "http://localhost:5556/dex"),
    API_URL: getEnvString(env.VITE_API_URL, "VITE_API_URL", "http://localhost:8880"),
    REDIRECT_URI: getEnvString(env.VITE_REDIRECT_URI, "VITE_REDIRECT_URI", "http://localhost:5173")
  };

  // Ensure the public directory exists
  const publicDir = path.resolve(process.cwd(), 'public');
  if (!fs.existsSync(publicDir)) {
    fs.mkdirSync(publicDir, { recursive: true });
  }

  // Write the runtime config file
  const configContent = `// Runtime configuration - these values can be replaced at deployment time
window.RUNTIME_CONFIG = ${JSON.stringify(runtimeConfig, null, 2)};`;
  
  fs.writeFileSync(path.join(publicDir, 'config.js'), configContent);
  console.log('Generated runtime config.js with current environment values');
}

export default defineConfig(({ mode, command }) => {
  // Load env files based on mode (.env, .env.production, etc)
  const env = loadEnv(mode, process.cwd(), '')
  
  // If we're building the app, generate the runtime config
  if (command === 'build') {
    generateRuntimeConfig(env);
  }
  
  // Get allowed hosts from environment variable or use default
  const ALLOWED_HOSTS = getEnvArray(process.env.VITE_ALLOWED_HOSTS || env.VITE_ALLOWED_HOSTS, "VITE_ALLOWED_HOSTS", ['127.0.0.1', 'localhost'])

  // Define runtime environment variable values for development and build
  const runtimeEnvs = {
    AUTH_URL: getEnvString(env.VITE_AUTH_URL, "VITE_AUTH_URL", "http://localhost:5556/dex"),
    API_URL: getEnvString(env.VITE_API_URL, "VITE_API_URL", "http://localhost:8880"),
    REDIRECT_URI: getEnvString(env.VITE_REDIRECT_URI, "VITE_REDIRECT_URI", "http://localhost:5173"),
  };

  return {
    base: "/admin/",
    plugins: [
      react(),
      // Custom plugin to inject runtime config during development
      {
        name: 'inject-runtime-config',
        transformIndexHtml(html) {
          // Only add the script tag in development mode, in production it will be a physical file
          if (command === 'serve') {
            const runtimeScript = `<script>
              window.RUNTIME_CONFIG = ${JSON.stringify(runtimeEnvs, null, 2)};
            </script>`;
            
            return html.replace('</head>', `${runtimeScript}</head>`);
          }
          return html;
        }
      },
    ],
    preview: {
      port: parseInt(getEnvString(env.VITE_PORT, "VITE_PORT", "5173")),
      strictPort: true,
    },
    server: {
      port: parseInt(getEnvString(env.VITE_PORT, "VITE_PORT", "5173")),
      strictPort: true,
      host: '0.0.0.0',
      origin: getEnvString(env.VITE_ORIGIN, "VITE_ORIGIN", "http://localhost:5173"),
      hmr: {
        clientPort: parseInt(getEnvString(env.VITE_CLIENT_PORT, "VITE_CLIENT_PORT", "5173")),
        protocol: getEnvString(env.VITE_HMR_PROTOCOL, "VITE_HMR_PROTOCOL", "ws"),   
      },
      fs: {
        strict: true,
      },
      allowedHosts: ALLOWED_HOSTS
    },
  }
});