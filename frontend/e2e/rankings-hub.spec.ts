import { test, expect } from '@playwright/test'
import { OWNER, login, uiLogin } from './helpers'

// (ل) 11-12 دازاي: التقييمات كلها بمكان واحد «تصنيف المدراء والموظفين»، والإحصائيات بلا تقييمات
// بس «كل خدمة شكد انطلبت هذا الشهر وبالكلي».
test.describe.configure({ timeout: 180_000 })

test('تصنيف المدراء والموظفين بمكان واحد + الإحصائيات بالخدمات', async ({ page, request }) => {
  const problems: string[] = []
  page.on('pageerror', (e) => problems.push(e.message))
  page.on('response', (r) => { if (r.url().includes('/api/') && r.status() >= 500) problems.push(`${r.status()} ${r.url()}`) })
  const owner = await login(request, OWNER.username, OWNER.password)
  await uiLogin(page, owner)
  await page.goto('./org/directors-ranking')
  await expect(page.getByRole('heading', { name: '🏆 تصنيف المدراء والموظفين' })).toBeVisible()
  const tabs = page.getByRole('tablist').getByRole('tab')
  for (const name of [/المدراء/, /الموظفين — ماتركس/, /المؤشرات/, /أرقام الشهر/, /أرقام الأقسام/]) await expect(tabs.filter({ hasText: name })).toHaveCount(1)
  await tabs.filter({ hasText: /الموظفين — ماتركس/ }).click()
  await expect(page).toHaveURL(/staff-score/)
  await expect(page.getByRole('heading', { name: '🏆 تصنيف المدراء والموظفين' })).toBeVisible()
  await tabs.filter({ hasText: /أرقام الأقسام/ }).click()
  await expect(page).toHaveURL(/staff-numbers/)

  await page.goto('./stats-management')
  await expect(page.getByTestId('svc-month')).toContainText('هذا الشهر')
  await expect(page.getByTestId('svc-all')).toContainText('الكلي')
  await expect(page.getByRole('button', { name: /إحصائيات الموظفين/ })).toHaveCount(0)
  await expect(page.getByRole('button', { name: /شهرية \(لكل موظف\)/ })).toHaveCount(0)
  expect(problems).toEqual([])
})
