import { test, expect } from '@playwright/test'
import { API, OWNER, login, prepare, uiLogin } from './helpers'

// (ل) 11-12 دازاي: «استقرار النظام من اضغط عليه المفروض يطلعلي منين جاب هاي النسبة» —
// المراقب يضغط البطاقة ويطلع: الحسبة بالأرقام، معنى الألوان، التوزيع على المجموعات، وسبب كل أحمر.
test.describe.configure({ timeout: 180_000 })

test('استقرار النظام: الضغط يطلّع منين جات النسبة', async ({ page, request }) => {
  const problems: string[] = []
  page.on('pageerror', (e) => problems.push(e.message))
  const mon = await prepare(request, 'MONITOR')
  await request.post(`${API}/attendance/checkin`, { headers: { Authorization: `Bearer ${mon.token}` }, data: {} })
  await uiLogin(page, mon)
  await page.goto('./')
  const cardBtn = page.getByRole('button', { name: /استقرار النظام/ })
  await expect(cardBtn).toContainText('منين جات النسبة؟')
  await expect(cardBtn).toContainText(/\d+٪/)
  await cardBtn.click()
  await expect(page.getByRole('heading', { name: /استقرار النظام \d+%/ })).toBeVisible()
  await expect(page.getByText(/\d+ موظف مو بالأحمر من أصل \d+ موظف تحت المراقبة \(\d+ ناقص \d+ أحمر\) = \d+%/)).toBeVisible()
  await expect(page.getByText('شنو تعني النسبة؟')).toBeVisible()
  await expect(page.getByText(/هذا الي ينزّل النسبة/)).toBeVisible()
  await expect(page.getByText('حسب المجموعة')).toBeVisible()
  await expect(page.getByText(/الحمر \(\d+\) وسبب كل واحد/)).toBeVisible()
  expect(problems).toEqual([])
})

// (ل) 11-12: «امشي على كل النسب بالنظام» — أي نسبة بالرئيسية تنضغط وتطلّع الحسبة والمعنى والمصدر.
test('نسب الرئيسية: كل نسبة تنضغط وتشرح نفسها', async ({ page, request }) => {
  const problems: string[] = []
  page.on('pageerror', (e) => problems.push(e.message))
  page.on('console', (m) => { if (m.type() === 'error' && /validateDOMNesting|cannot be a descendant/.test(m.text())) problems.push(m.text()) })
  const owner = await login(request, OWNER.username, OWNER.password)
  // قاعدة CI فاضية: بلا حجوزات ماكو نسب تنعرض — نسجّل زبون وحجز حتى تطلع «نجاح الأسبوع»
  const h = { Authorization: `Bearer ${owner.token}` }
  const phone = '0790' + String(Math.floor(Math.random() * 1e7)).padStart(7, '0')
  const cu = await request.post(`${API}/customers`, { headers: h, data: { name: `زبون نسب ${Date.now().toString(36)}`, phone } })
  expect(cu.ok(), await cu.text()).toBeTruthy()
  const cust = await cu.json() as { id?: string; customer?: { id: string } }
  const bk = await request.post(`${API}/bookings`, { headers: h, data: { customerId: cust.id ?? cust.customer?.id } })
  expect(bk.ok(), await bk.text()).toBeTruthy()
  await uiLogin(page, owner)
  await page.goto('./')
  await page.getByRole('button', { name: /المتابعة/ }).first().click()
  const pcts = page.locator('[role="button"][aria-label*="منين جات؟"]')
  await expect(pcts.first()).toBeVisible({ timeout: 30_000 })
  const n = Math.min(await pcts.count(), 6)
  for (let i = 0; i < n; i++) {
    await pcts.nth(i).click()
    const dlg = page.getByRole('dialog', { name: /منين جات/ })
    await expect(dlg.getByTestId('pct-formula')).toBeVisible()
    await expect(dlg.getByText('💡 شنو تعني')).toBeVisible()
    await expect(dlg.getByText('📂 منين تجي الأرقام')).toBeVisible()
    await dlg.getByRole('button', { name: 'سد' }).click()
  }
  expect(problems).toEqual([])
})

// (ل) 11-12: ثقة التوقعات المعلّقة صارت من دقتها الحقيقية — للمالك بس، والغريب ما يشوفها.
test('دقة توقعات ماتركس: رقم محسوب مو ثابت', async ({ request }) => {
  const owner = await login(request, OWNER.username, OWNER.password)
  const r = await request.get(`${API}/ai/prediction-accuracy`, { headers: { Authorization: `Bearer ${owner.token}` } })
  expect(r.ok()).toBeTruthy()
  const a = await r.json() as { days: number; correct: number; judged: number; pct: number | null; recent: unknown[] }
  expect(a.days).toBe(30)
  expect(a.pct).toBe(a.judged > 0 ? Math.round((a.correct * 100) / a.judged) : null)
  const tech = await prepare(request, 'TECHNICIAN')
  const t = await request.get(`${API}/ai/prediction-accuracy`, { headers: { Authorization: `Bearer ${tech.token}` } })
  expect(t.ok()).toBeFalsy()
})
