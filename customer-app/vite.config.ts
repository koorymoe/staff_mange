import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// تطبيق الزبائن — واجهة مستقلة على دومينها، و/api يروح للباك نفسه.
export default defineConfig({
  base: '/',
  plugins: [react(), tailwindcss()],
  server: { proxy: { '/api': 'http://localhost:4000' } },
})
