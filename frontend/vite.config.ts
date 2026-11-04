import { execSync } from 'node:child_process'
import { createRequire } from 'node:module'
import path from 'node:path'
import { defineConfig } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

// https://vite.dev/config/
/**
 * بصمة البناء — تُحقن وقت البناء وتُعرض بشاشة المختبر للمالك.
 *
 * 🔴 **السبب من الواقع مو تزييناً**: سؤال «هل وصل التحديث؟» تكرّر
 * **ثلاث مرات بيوم واحد**. ومالك النظام شاف إصلاحاً مرفوعاً ومقاساً
 * (نزول الأذرع ٢٣.٦ سم) وما بيّن عنده — لأن `docker compose up -d`
 * **يعيد استخدام صورة الواجهة القديمة** بلا `--build`، فالخادم صار
 * جديداً والواجهة لا. وكل مرة انصرف وقت بجدال «مرفوع/مو مرفوع».
 *
 * فالبصمة تخلي **النظام يجاوب** بدل ما نخمّن.
 *
 * ⚠️ ولا تفشّل البناء لو گيت مو موجود (البناء داخل حاوية بلا
 * تاريخ گيت): نرجّع `unknown` ونكمل.
 */
function buildStamp(): string {
  try {
    return execSync('git rev-parse --short HEAD', { stdio: ['ignore', 'pipe', 'ignore'] })
      .toString().trim() || 'unknown'
  } catch {
    return 'unknown'
  }
}

// المسار الكامل لملف داخل نسخة الإنتاج (الحزمة ما تسمح باستيراده بالاسم)
const rrProd = (f: string) => path.join(path.dirname(createRequire(import.meta.url).resolve('react-router')), '..', 'production', f)

export default defineConfig(({ command }) => ({
  plugins: [react(), tailwindcss()],
  // (11-04 — تعديلات ٧ك) react-router 7 يصدّر نسخة التطوير افتراضياً (أكبر وبيها
  // فحوصات تطوير). بالبناء نوجّهه لنسخة الإنتاج الي جوّا نفس الحزمة.
  resolve: command === 'build' ? {
    alias: [
      { find: /^react-router$/, replacement: rrProd('index.mjs') },
      { find: /^react-router\/dom$/, replacement: rrProd('dom-export.mjs') },
    ],
  } : {},
  define: {
    __BUILD_COMMIT__: JSON.stringify(buildStamp()),
    __BUILD_TIME__: JSON.stringify(new Date().toISOString()),
  },
  // بالسيرفر الفعلي (دومين الشركة) الموقع يكون على الجذر '/'؛ نطاق GitHub
  // Pages القديم يبقى يشتغل بمساره '/staff_mange/' الافتراضي.
  base: process.env.VITE_BASE_PATH || '/staff_mange/',
}))
