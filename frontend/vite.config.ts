import { defineConfig } from 'vite'
import vue from '@vitejs/plugin-vue'
import { resolve } from 'path'
import type { ProxyOptions } from 'vite'

/**
 * The app talks to the API through the relative baseURL "/api/v1", so both the
 * dev and preview servers must proxy "/api" to the Go backend.
 *
 * IMPORTANT: by default an upstream connection error (e.g. the backend is still
 * compiling and refuses the connection, ECONNREFUSED) is emitted as an
 * unhandled 'error' event on the proxy, which *kills the whole server process*.
 * That is what took down the preview server in the middle of a long E2E run.
 * This handler answers with 502 instead, so a transient upstream failure can
 * never bring the frontend down.
 */
const apiProxy: ProxyOptions = {
  target: 'http://localhost:8000',
  changeOrigin: true,
  configure: (proxy) => {
    proxy.on('error', (err, _req, res) => {
      console.warn(`[proxy] /api upstream error: ${err?.message ?? err}`)
      // `res` is a ServerResponse for normal requests, but the proxy also emits
      // errors for sockets that never produced one — guard before writing.
      const response = res as unknown as {
        headersSent?: boolean
        writeHead?: (code: number, headers: Record<string, string>) => void
        end?: (body?: string) => void
      }
      if (response && response.writeHead && !response.headersSent) {
        response.writeHead(502, { 'Content-Type': 'application/json; charset=utf-8' })
        response.end(JSON.stringify({ message: 'upstream API unavailable' }))
      }
    })
  },
}

export default defineConfig({
  plugins: [vue()],
  resolve: {
    alias: {
      '@': resolve(__dirname, 'src')
    }
  },
  server: {
    port: 5173,
    host: '0.0.0.0',
    proxy: { '/api': apiProxy }
  },
  // Serving a built bundle is far more stable than the dev server, which
  // recompiles modules on demand and accumulated enough state to die during a
  // long E2E run. Both share the crash-proof proxy above.
  preview: {
    port: 5173,
    strictPort: true,
    host: '0.0.0.0',
    proxy: { '/api': apiProxy }
  }
})
