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
