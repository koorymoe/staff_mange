import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react(), tailwindcss()],
  // بالسيرفر الفعلي (دومين الشركة) الموقع يكون على الجذر '/'؛ نطاق GitHub
  // Pages القديم يبقى يشتغل بمساره '/staff_mange/' الافتراضي.
  base: process.env.VITE_BASE_PATH || '/staff_mange/',
})
