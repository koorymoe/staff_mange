import { defineConfig } from '@playwright/test'

// ٧أ (11-04): اختبارات متصفح سريعة على الشاشات المهمة. تحتاج السيرفر شغّال
// (E2E_API، الافتراضي localhost:4000) والواجهة (E2E_BASE). بالـ CI تنبني وتشتغل بخطوة لحالها.
export default defineConfig({
  testDir: './e2e',
  timeout: 60_000,
  retries: process.env.CI ? 1 : 0,
  workers: 1,
  reporter: process.env.CI ? [['list'], ['html', { open: 'never' }]] : 'list',
  use: {
    baseURL: process.env.E2E_BASE ?? 'http://127.0.0.1:5199/staff_mange/',
    viewport: { width: 1400, height: 1000 },
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
  },
})
