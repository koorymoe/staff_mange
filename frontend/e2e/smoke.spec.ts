import { test, expect, type Page } from '@playwright/test'

// ٧أ (11-04): نفس الفحص الي چان ينسوى باليد بعد كل تعديل — يدخل المالك، يفتح
// الشاشات المهمة، ويفشل إذا الصفحة طلّعت خطأ أو السيرفر رد ٥xx أو ٤٠٣.
const USER = process.env.E2E_USER ?? 'owner'
const PASS = process.env.E2E_PASS ?? 'Owner-pass-123'

const PAGES: [string, RegExp][] = [
  ['', /./],
  ['bookings', /الحجوزات|حجز/],
  ['leader-invoices', /فواتير/],
  ['customers', /الزبائن/],
  ['complaints', /الشكاو/],
  ['expenses', /المصاريف|مصروف/],
  ['projects', /المشاريع|مشروع/],
  ['org/directors-ranking', /تصنيف القيادة/],
  ['owner-integrity', /سلامة/],
]

async function login(page: Page) {
  await page.goto('./')
  const inputs = page.locator('input')
  await inputs.first().waitFor()
  await inputs.nth(0).fill(USER)
  await inputs.nth(1).fill(PASS)
  await page.keyboard.press('Enter')
  await page.waitForLoadState('networkidle')
  // نوافذ إجبارية لأول دخول (المبادئ) — تنقرا لآخرها وتنقبل
  for (let i = 0; i < 3; i++) {
    const accept = page.getByRole('button', { name: /قرأت وألتزم/ })
    if (!(await accept.count())) break
    await page.evaluate(() => document.querySelectorAll('*').forEach((el) => { if (el.scrollHeight > el.clientHeight + 20) el.scrollTop = el.scrollHeight }))
    await accept.click()
    await page.waitForTimeout(800)
  }
}

test('المالك يفتح الشاشات المهمة بلا أخطاء', async ({ page }) => {
  const problems: string[] = []
  page.on('pageerror', (e) => problems.push(`pageerror: ${e.message}`))
  page.on('response', (r) => {
    const u = r.url()
    if (!u.includes('/api/')) return
    // ٤٠١ قبل الدخول طبيعي (طلبات الخلفية)
    if (r.status() >= 500 || r.status() === 403) problems.push(`${r.status()} ${r.request().method()} ${u}`)
  })
  await login(page)
  for (const [path, heading] of PAGES) {
    await page.goto('./' + path)
    await page.waitForLoadState('networkidle')
    await expect(page.locator('body'), `الصفحة /${path}`).toContainText(heading)
  }
  expect(problems, problems.join('\n')).toEqual([])
})

test('الخروج يمسح الجلسة بالسيرفر (٧ي)', async ({ page, request }) => {
  await login(page)
  const api = process.env.E2E_API ?? 'http://127.0.0.1:4000/api'
  const r = await request.post(`${api}/auth/logout`)
  expect(r.ok()).toBeTruthy()
  const cookie = r.headers()['set-cookie'] ?? ''
  expect(cookie).toContain('sm_session=;')
})
