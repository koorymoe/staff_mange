import { test, expect } from '@playwright/test'
import { API, OWNER, login, prepare, uiLogin } from './helpers'

// (ل) 11-12 دازاي — وحدة التدريب: المنهج شجرة (فني ← داش كام ← قفل باب ← تنصيب ← الحفر)،
// والماتشينغ: تدريب إلزامي ← بدي ← تقرير وتقييم من مسؤول التدريب ← يطلع بتقارير التدريب للمالك.
test.describe.configure({ timeout: 180_000 })

test('وحدة التدريب: شجرة المنهج + الماتشينغ + تقرير للمالك', async ({ page, request }) => {
  const problems: string[] = []
  page.on('pageerror', (e) => problems.push(e.message))
  page.on('response', (r) => { if (r.url().includes('/api/') && r.status() >= 500) problems.push(`${r.status()} ${r.url()}`) })
  const tech = await prepare(request, 'TECHNICIAN')
  const techName = (tech.employee as { name: string }).name
  const tag = Date.now().toString(36)
  const owner = await login(request, OWNER.username, OWNER.password)
  await uiLogin(page, owner)
  await page.goto('./training-center')

  // ── المنهج ──
  await page.getByRole('button', { name: /مفردات المنهج/ }).click()
  await page.getByRole('button', { name: /^فني/ }).click()
  const track = `داش كام ${tag}`
  await page.getByPlaceholder('عنوان المسار (مثلاً: داش كام)').fill(track)
  await page.getByRole('button', { name: '+ مسار جديد' }).click()
  const levels: [string, string, string][] = [[track, '+ تفرع أول', 'قفل باب'], ['قفل باب', '+ تفرع ثاني', 'تنصيب قفل باب'], ['تنصيب قفل باب', '+ تفرع ثالث', 'الحفر']]
  for (const [parent, button, child] of levels) {
    const row = page.locator('li > div', { has: page.getByText(parent, { exact: true }) }).last()
    await row.getByRole('button', { name: button }).click()
    await page.getByPlaceholder(new RegExp(`تحت «${parent}»`)).fill(child)
    await page.getByRole('button', { name: 'إضافة', exact: true }).click()
    await expect(page.getByText(child, { exact: true })).toBeVisible()
  }
  // التفرع الثالث ما تنزل تحته شي، والي بيه تفرعات ما ينحذف
  const third = page.locator('li > div', { has: page.getByText('الحفر', { exact: true }) }).last()
  await expect(third.getByRole('button', { name: /^\+/ })).toHaveCount(0)
  const trackRow = page.locator('li > div', { has: page.getByText(track, { exact: true }) }).first()
  await expect(trackRow.getByRole('button', { name: 'حذف' })).toHaveCount(0)

  // ── الماتشينغ ──
  await page.getByRole('button', { name: /الماتشينغ/ }).click()
  await expect(page.getByRole('heading', { name: /منو يحتاج تدريب/ })).toBeVisible()
  const topic = `الحفر ${tag}`
  const addBox = page.locator('div', { has: page.getByRole('heading', { name: /أضف تدريب إلزامي/ }) }).last()
  await addBox.locator('select').selectOption({ label: techName })
  await addBox.getByPlaceholder('يتدرب على شنو').fill(topic)
  await addBox.getByPlaceholder('السبب').fill('الليدر بلّغ إنه ما يعرف يحفر')
  await addBox.getByRole('button', { name: 'إضافة' }).click()
  const need = page.getByTestId('need-row').filter({ hasText: topic })
  await expect(need.getByText(/إلزامي/)).toBeVisible()
  await expect(need.getByRole('button', { name: 'ما يحتاج' })).toHaveCount(0) // الإلزامي ما ينلغي
  await need.getByRole('button', { name: '📖 بدي التدريب' }).click()
  await expect(need.getByText('📖 بالتدريب')).toBeVisible()
  await need.getByRole('button', { name: /خلص — اكتب التقرير/ }).click()
  await need.getByPlaceholder(/تقرير التدريب/).fill('تدرب يومين على الحفر وطبّق بحجز')
  await need.getByRole('button', { name: '4 من ٥' }).click()
  await need.getByRole('button', { name: 'أرسل التقرير للمالك' }).click()
  await expect(page.getByTestId('need-row').filter({ hasText: topic })).toHaveCount(0)

  // ── تقارير التدريب ──
  await page.getByRole('button', { name: /تقارير التدريب/ }).click()
  const rep = page.getByTestId('training-report').filter({ hasText: topic })
  await expect(rep.getByText('تدرب يومين على الحفر وطبّق بحجز')).toBeVisible()
  await expect(rep.getByLabel('التقييم 4 من ٥')).toBeVisible()
  expect(problems).toEqual([])

  // الفني ما يدخل الماتشينغ
  const r = await request.get(`${API}/training-center/matching`, { headers: { Authorization: `Bearer ${tech.token}` } })
  expect([403, 404]).toContain(r.status())
})
