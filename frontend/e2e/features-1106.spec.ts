import { test, expect, type Page } from '@playwright/test'

// ميزات 11-06 (دازاي): قوالب الملحق، صحة الوكلاء + الامتحان، أفعال الوكلاء، الالتزام والصور،
// الصيانة الوقائية، ومسار اليوم — كل وحدة تنفتح بلا خطأ ولا ٥xx/٤٠٣.
const USER = process.env.E2E_USER ?? 'owner'
const PASS = process.env.E2E_PASS ?? 'Owner-pass-123'

async function login(page: Page) {
  await page.goto('./')
  const inputs = page.locator('input')
  await inputs.first().waitFor()
  await inputs.nth(0).fill(USER)
  await inputs.nth(1).fill(PASS)
  await page.keyboard.press('Enter')
  await page.waitForLoadState('networkidle')
  for (let i = 0; i < 3; i++) {
    const accept = page.getByRole('button', { name: /قرأت وألتزم/ })
    if (!(await accept.count())) break
    await page.evaluate(() => document.querySelectorAll('*').forEach((el) => { if (el.scrollHeight > el.clientHeight + 20) el.scrollTop = el.scrollHeight }))
    await accept.click()
    await page.waitForTimeout(800)
  }
}

function watch(page: Page) {
  const problems: string[] = []
  page.on('pageerror', (e) => problems.push(`pageerror: ${e.message}`))
  page.on('response', (r) => {
    if (r.url().includes('/api/') && (r.status() >= 500 || r.status() === 403)) problems.push(`${r.status()} ${r.url()}`)
  })
  return problems
}

test('قوالب الملحق: نسخ كقالب جديد وحفظ', async ({ page }) => {
  const problems = watch(page)
  await login(page)
  await page.goto('./quotations/project/new')
  await page.getByRole('button', { name: '⚙️ قوالب الملحق' }).click()
  const dlg = page.getByRole('dialog', { name: 'قوالب الملحق' })
  await dlg.getByRole('button', { name: '📄 نسخ كقالب جديد' }).click()
  const name = 'قالب e2e ' + Date.now().toString(36)
  await dlg.getByLabel('اسم القالب').fill(name)
  await dlg.getByRole('button', { name: '💾 احفظ كقالب جديد' }).click()
  await expect(dlg).toContainText('✅ انحفظ القالب')
  await dlg.getByRole('button', { name: 'سد' }).click()
  await expect(page.locator('select option', { hasText: name })).toHaveCount(1)
  expect(problems, problems.join('\n')).toEqual([])
})

test('صحة الوكلاء وأفعالهم ومكتب الجودة والخدمات تنفتح', async ({ page }) => {
  const problems = watch(page)
  await login(page)
  await page.goto('./ai-suggestions')
  await page.getByRole('button', { name: /صحة الوكلاء/ }).click()
  await expect(page.getByRole('button', { name: /امتحن ماتركس/ })).toBeVisible()
  await page.getByRole('button', { name: /أفعال الوكلاء/ }).click()
  await expect(page.locator('body')).toContainText(/أفعال|ماكو أفعال/)
  await page.goto('./quality-desk')
  await page.getByRole('button', { name: /صيانة وقائية/ }).click()
  await expect(page.locator('body')).toContainText(/زبون/)
  await page.goto('./service-managers')
  await expect(page.getByRole('region', { name: 'الالتزام والصور' })).toBeVisible()
  await page.goto('./coordinator')
  await page.waitForLoadState('networkidle')
  expect(problems, problems.join('\n')).toEqual([])
})
