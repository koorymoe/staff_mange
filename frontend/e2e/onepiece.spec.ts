import { test, expect } from '@playwright/test'
import { API, login, prepare, uiLogin } from './helpers'

// 🏴‍☠️ ون بيس (11-10): مصفوفة الصلاحيات ولوحة الأدمن — للمالك بس، والأسرار مخفية.
test('المالك: مصفوفة الصلاحيات ولوحة الأدمن', async ({ page, request }) => {
  test.setTimeout(120_000)
  const owner = await login(request, process.env.E2E_OWNER ?? 'owner', process.env.E2E_OWNER_PASS ?? 'Owner-pass-123')
  await uiLogin(page, owner)

  await page.goto('./core-matrix')
  await expect(page.getByRole('heading', { name: /مصفوفة الصلاحيات/ })).toBeVisible({ timeout: 15_000 })
  const total = Number(await page.getByRole('button', { name: /كل الأفعال/ }).locator('div').first().textContent())
  expect(total, 'كل مسار إله فعل').toBeGreaterThan(1000)
  await page.fill('#core-matrix-q', 'people.suspend')
  await expect(page.getByText('يوقف حساب موظف')).toBeVisible()
  await page.screenshot({ path: 'test-results/onepiece-matrix.png', fullPage: false })

  await page.goto('./core-admin')
  await page.fill('#core-admin-filter', 'Employee')
  await page.getByRole('button', { name: /^Employee\s/ }).first().click()
  await expect(page.getByText(/صف · صفحة/)).toBeVisible({ timeout: 15_000 })
  // كلمة السر تطلع مخفية — وما تطلع قيمة حقيقية بالجدول أبداً
  const html = await page.content()
  expect(html).toContain('•••')
  expect(html).not.toMatch(/\$2[aby]\$/)
  await page.screenshot({ path: 'test-results/onepiece-admin.png', fullPage: false })
})

test('الفني: الصفحتين مو موجودات والـAPI يرد 404', async ({ page, request }) => {
  test.setTimeout(120_000)
  const tech = await prepare(request, 'TECHNICIAN')
  const h = { Authorization: `Bearer ${tech.token}` }
  expect((await request.get(`${API}/core/matrix`, { headers: h })).status()).toBe(404)
  expect((await request.get(`${API}/core/admin/tables`, { headers: h })).status()).toBe(404)
  await uiLogin(page, tech)
  await page.goto('./core-admin')
  // حارس الواجهة العام يمنعه قبل الصفحة، أو الصفحة نفسها — بالحالتين ما يشوف شي
  await expect(page.getByText(/الصفحة غير موجودة|ما عندك صلاحية لهاي الشاشة/)).toBeVisible({ timeout: 15_000 })
})
