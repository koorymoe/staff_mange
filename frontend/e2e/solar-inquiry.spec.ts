import { test, expect } from '@playwright/test'
import { API, OWNER, login, prepare, uiLogin } from './helpers'

// (ل) 11-12 دازاي: استفسار طاقة شمسية من الأول للأخير — تسجيل ← التقني «ما رد» ← ابو الجودة
// يتصل والزبون يگول التقني ما اتصل. والي ماله صلاحية ما يشوف (404).
test.describe.configure({ timeout: 180_000 })

test('استفسار طاقة شمسية: تسجيل ← تقني ← ابو الجودة', async ({ page, request }) => {
  const problems: string[] = []
  page.on('pageerror', (e) => problems.push(e.message))
  page.on('response', (r) => { if (r.url().includes('/api/') && r.status() >= 500) problems.push(`${r.status()} ${r.url()}`) })
  const owner = await login(request, OWNER.username, OWNER.password)
  await uiLogin(page, owner)
  await page.goto('./solar-inquiries')
  const name = `زبون شمسي ${Date.now().toString(36)}`
  await page.locator('#si-name').fill(name)
  await page.locator('#si-phone').fill('07701234567')
  await page.locator('#si-note').fill('يسأل عن منظومة ٥ كيلو')
  await page.getByRole('button', { name: 'سجّل وارسل للتقنيين' }).click()
  await expect(page.getByText('انسجل الاستفسار وانرسل للتقنيين ✅')).toBeVisible()
  const card = page.locator('div.rounded-xl.border', { hasText: name })
  await expect(card.getByText('ينتظر التقني')).toBeVisible()

  await card.getByRole('button', { name: '📞 آخذه أنا' }).click()
  await card.getByRole('button', { name: 'سجّل نتيجة الاتصال' }).click()
  await card.getByRole('button', { name: 'ما رد', exact: true }).click()
  await card.getByRole('button', { name: 'احفظ', exact: true }).click()
  await expect(page.getByText('انسجل اتصالك وراح لابو الجودة يتأكد')).toBeVisible()

  await page.getByRole('button', { name: /ينتظر ابو الجودة/ }).click()
  const c2 = page.locator('div.rounded-xl.border', { hasText: name })
  await c2.getByRole('button', { name: '⭐ اتصلت بالزبون (ابو الجودة)' }).click()
  await expect(c2.getByText('التقني كاتب «ما رد» — اسأل الزبون: اتصل بيك التقني؟')).toBeVisible()
  await c2.getByRole('button', { name: 'إي', exact: true }).first().click()
  const grp = (l: string) => c2.locator('fieldset', { hasText: l })
  await grp('اتصل بيك التقني؟').getByRole('button', { name: 'لا', exact: true }).click()
  await grp('شنو چان يريد؟').getByRole('button', { name: 'يريد ينصب منظومة' }).click()
  await grp('انطوك عرض سعر؟').getByRole('button', { name: 'لا', exact: true }).click()
  await grp('العرض مناسب ومنافس للسوق؟').getByRole('button', { name: 'ما يدري' }).click()
  await c2.getByRole('button', { name: 'احفظ المتابعة' }).click()
  await expect(page.getByText('انسجلت المتابعة ✅')).toBeVisible()
  await page.getByRole('button', { name: /مسكّرة/ }).click()
  await expect(page.locator('div.rounded-xl.border', { hasText: name }).getByText('مسكّر', { exact: true })).toBeVisible()
  expect(problems).toEqual([])

  const sales = await prepare(request, 'SALES')
  const r = await request.get(`${API}/solar-inquiries`, { headers: { Authorization: `Bearer ${sales.token}` } })
  expect(r.status()).toBe(404)
})
