import { test, expect } from '@playwright/test'
import { API, OWNER, login, prepare, uiLogin } from './helpers'

// بلاغ #33 (مؤمل 11-12): عرض سعر انسوّى قبل ما الزبون يصعد مشروع — من كارت المشروع
// «🔗 ربط عرض موجود» يطلع عرض نفس الزبون ويرتبط. والي ماعنده عروض = 404.
test.describe.configure({ timeout: 180_000 })

test('ربط عرض سعر موجود بمشروع', async ({ page, request }) => {
  const problems: string[] = []
  page.on('pageerror', (e) => problems.push(e.message))
  page.on('response', (r) => { if (r.url().includes('/api/') && r.status() >= 500) problems.push(`${r.status()} ${r.url()}`) })
  const owner = await login(request, OWNER.username, OWNER.password)
  const h = { Authorization: `Bearer ${owner.token}` }
  const tag = Date.now().toString(36)
  const name = `زبون ربط ${tag}`
  const phone = '0781' + String(Math.floor(Math.random() * 1e7)).padStart(7, '0')
  const q = await request.post(`${API}/quotations`, { headers: h, data: { customerName: name, customerPhone: phone, items: [] } })
  expect(q.ok(), await q.text()).toBeTruthy()
  const quote = await q.json() as { id: string; quotationNumber: string }
  const p = await request.post(`${API}/projects`, { headers: h, data: { name, phone } })
  expect(p.ok(), await p.text()).toBeTruthy()
  const proj = await p.json() as { id: string }
  const up = await request.put(`${API}/projects/${proj.id}`, { headers: h, data: { name, phone, stage: '3. عرض السعر' } })
  expect(up.ok(), await up.text()).toBeTruthy()

  await uiLogin(page, owner)
  await page.goto('./projects')
  await page.getByPlaceholder(/بحث|ابحث/).first().fill(name)
  const card = page.locator('div', { hasText: name }).filter({ has: page.getByRole('button', { name: '🔗 ربط عرض موجود' }) }).last()
  await card.getByRole('button', { name: '🔗 ربط عرض موجود' }).click()
  const dlg = page.getByTestId('link-quotation')
  const row = dlg.locator('div.border', { hasText: quote.quotationNumber })
  await expect(row.getByText('نفس الزبون')).toBeVisible()
  await row.getByRole('button', { name: 'ربط' }).click()
  await expect(dlg).toBeHidden()

  const by = await (await request.get(`${API}/quotation-by-project/${proj.id}`, { headers: h })).json() as { id: string }
  expect(by.id).toBe(quote.id)
  await card.getByRole('button', { name: '🔗 ربط عرض موجود' }).click()
  await expect(dlg.getByText(`مربوط بعرض ${quote.quotationNumber}`)).toBeVisible()
  expect(problems).toEqual([])

  const tech = await prepare(request, 'TECHNICIAN')
  const r = await request.get(`${API}/projects/${proj.id}/quotation-candidates`, { headers: { Authorization: `Bearer ${tech.token}` } })
  expect(r.status()).toBe(404)
})
