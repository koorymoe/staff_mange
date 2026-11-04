import { defineConfig } from 'vitest/config'

// اختبارات الوحدة بـ src بس — e2e/ لـ Playwright (npm run e2e).
export default defineConfig({
  test: { include: ['src/**/*.test.ts', 'src/**/*.test.tsx'], environment: 'node' },
})
