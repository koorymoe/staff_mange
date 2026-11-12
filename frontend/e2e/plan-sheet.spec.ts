import { test, expect } from '@playwright/test'
import fs from 'fs'
import { API, OWNER, login, uiLogin } from './helpers'

// (ل) 11-12 دازاي: فورما المخطط — المصممة تختار المكان (يمين/جوّه) وتعدّل الاسم والملاحظات والسكيل،
// واللوغو والتاريخ ثابتين. تنحفظ وترجع بنفس القيم.
test.describe.configure({ timeout: 180_000 })

test('فورما المخطط: تعديل وحفظ ومعاينة', async ({ page, request }) => {
  const problems: string[] = []
  page.on('pageerror', (e) => problems.push(e.message))
  page.on('response', (r) => { if (r.url().includes('/api/') && r.status() >= 500) problems.push(`${r.status()} ${r.url()}`) })
  const owner = await login(request, OWNER.username, OWNER.password)
  const h = { Authorization: `Bearer ${owner.token}` }
  const set = await (await request.post(`${API}/plan-sets`, { headers: h, data: { title: `سيت فورما ${Date.now().toString(36)}` } })).json() as { id: string }
  const up = await request.post(`${API}/files?folder=designs`, { headers: h, multipart: { file: { name: 'plan.png', mimeType: 'image/png', buffer: fs.readFileSync('public/favicon.png') } } })
  expect(up.ok(), await up.text()).toBeTruthy()
  const f = await up.json() as { url: string; type: string }
  const add = await request.post(`${API}/plan-sets/${set.id}/files`, { headers: h, data: { typeCode: '7.1', title: 'plan.png', fileKey: f.url, fileType: f.type } })
  expect(add.ok(), await add.text()).toBeTruthy()

  await uiLogin(page, owner)
  await page.goto(`./plan-sets/${set.id}`)
  await page.getByRole('button', { name: /جدول الكميات/ }).first().click()
  await page.getByRole('button', { name: '🗂️ الفورما' }).first().click()
  const ed = page.getByTestId('sheet-editor')
  await expect(ed.getByText(/لوغو الأماني والتاريخ/)).toBeVisible()
  await ed.getByRole('button', { name: /جوّه/ }).click()
  await ed.getByLabel('اسم المخطط').fill('الطابق الأرضي: أبواب وشبابيك')
  await ed.getByLabel('الملاحظات').fill('القياسات بالمتر')
  await ed.getByLabel('السكيل').fill('1:100')
  await expect(ed.getByTestId('sheet-preview')).toContainText('1:100')
  await ed.getByRole('button', { name: 'احفظ الفورما' }).click()
  await expect(ed).toBeHidden()
  const d = await (await request.get(`${API}/plan-sets/${set.id}`, { headers: h })).json() as { items: { sheetPos: string; sheetName: string; sheetNotes: string; sheetScale: string }[] }
  expect(d.items[0]).toMatchObject({ sheetPos: 'bottom', sheetName: 'الطابق الأرضي: أبواب وشبابيك', sheetNotes: 'القياسات بالمتر', sheetScale: '1:100' })
  await page.getByRole('button', { name: '🗂️ الفورما' }).first().click()
  await expect(page.getByTestId('sheet-editor').getByLabel('السكيل')).toHaveValue('1:100')
  expect(problems).toEqual([])
})
