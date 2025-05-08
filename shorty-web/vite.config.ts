import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react-swc'

// https://vite.dev/config/
// export default defineConfig({
//   plugins: [react()],
// })

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
    clientPort: 5173, // Angi klientporten for Hot Module Replacement
    protocol: 'ws',   // Angi WebSocket-protokoll
  },
 },
});