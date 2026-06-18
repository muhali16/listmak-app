import { defineConfig, loadEnv } from 'vite'
import vue from '@vitejs/plugin-vue'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  // Load env file based on `mode` in the current working directory.
  const env = loadEnv(mode, process.cwd(), '')
  
  // Parse allowedHosts from comma-separated string to array
  const allowedHosts = env.VITE_SERVER_ALLOWED_HOSTS 
    ? env.VITE_SERVER_ALLOWED_HOSTS.split(',').map(host => host.trim())
    : ['localhost']

  return {
    plugins: [vue()],
    server: {
      host: env.VITE_SERVER_HOST || 'localhost',
      port: parseInt(env.VITE_SERVER_PORT) || 5173,
      proxy: {
        '/api': {
          target: env.VITE_API_PROXY_TARGET || 'http://localhost:9001',
          changeOrigin: true,
          secure: false
        }
      },
      allowedHosts: allowedHosts
    }
  }
})
