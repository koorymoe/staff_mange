import { test, expect, type Page } from '@playwright/test'
import { API, OWNER, login, prepare, uiLogin } from './helpers'

// ميزات 11-10/11-11 (دازاي): أعذار التأخير (الموظف يقدّم والمالك يقبل)، مراقبة التقييم،
// وسيت المخططات (نموذج جاهز + زرين PDF + عرض السعر المربوط) — بلا خطأ ولا ٥xx/٤٠٣.
function watch(page: Page) {
  const problems: string[] = []
  page.on('pageerror', (e) => problems.push(`pageerror: ${e.message}`))
  page.on('response', (r) => {
    if (r.url().includes('/api/') && (r.status() >= 500 || r.status() === 403)) problems.push(`${r.status()} ${r.url()}`)
  })
  return problems
}

test('عذر تأخير: الفني يقدّم من صفحة الحضور، والمالك يقبله، والفني يشوفه مقبول', async ({ browser, request }) => {
  const tech = await prepare(request, 'TECHNICIAN')
  const reason = `عطل سيارة ${Date.now().toString(36)}`
  const p1 = await browser.newPage()
  const pr1 = watch(p1)
  await uiLogin(p1, tech)
  await p1.goto('./attendance')
  await expect(p1.getByText('⏰ أعذار التأخير والانضباط')).toBeVisible()
  await p1.locator('#lx-reason').fill(reason)
  await p1.getByRole('button', { name: 'قدّم العذر' }).click()
  await expect(p1.getByText('انرسل العذر لمسؤولك ✅')).toBeVisible()
  await expect(p1.getByText(reason)).toBeVisible()

  const owner = await login(request, OWNER.username, OWNER.password)
  const p2 = await browser.newPage()
  const pr2 = watch(p2)
  await uiLogin(p2, owner)
  await p2.goto('./attendance')
  const row = p2.locator('div', { hasText: reason }).filter({ has: p2.getByRole('button', { name: 'قبول' }) }).last()
  await row.getByRole('button', { name: 'قبول' }).click()
  await expect(p2.locator('div', { hasText: reason }).getByText(/مقبول ✅/).first()).toBeVisible()

  await p1.reload()
  await expect(p1.locator('li', { hasText: reason }).getByText('مقبول ✅')).toBeVisible()
  // الغريب (فني ثاني) ما يگدر يبت — مو موجود
  const other = await prepare(request, 'TECHNICIAN')
  const mine = await (await request.get(`${API}/late-excuses/mine`, { headers: { Authorization: `Bearer ${tech.token}` } })).json()
  const id = mine.excuses[0].id
  const r = await request.post(`${API}/late-excuses/${id}/decide`, { headers: { Authorization: `Bearer ${other.token}` }, data: { approve: false } })
  expect(r.status()).toBe(404)
  expect([...pr1, ...pr2]).toEqual([])
})

test('مراقبة التقييم تنفتح للمالك بجداولها، وللفني لا', async ({ page, request }) => {
  const problems = watch(page)
  const owner = await login(request, OWNER.username, OWNER.password)
  await uiLogin(page, owner)
  await page.goto('./org/score-monitor')
  await expect(page.getByRole('heading', { name: '🔍 مراقبة التقييم' })).toBeVisible()
  await expect(page.getByText('🏛️ المدراء — منين جاي الرقم')).toBeVisible()
  await expect(page.getByText('👥 الموظفين — تفصيل كل تقييم')).toBeVisible()
  expect(problems).toEqual([])
  const tech = await prepare(request, 'TECHNICIAN')
  const r = await request.get(`${API}/late-excuses/audit`, { headers: { Authorization: `Bearer ${tech.token}` } })
  expect(r.status()).toBe(404)
})

test('سيت المخططات: نموذج جاهز، وزر PDF التصاميم، وعرض السعر يفتح مربوط بالسيت', async ({ page, request }) => {
  const problems = watch(page)
  const owner = await login(request, OWNER.username, OWNER.password)
  const h = { Authorization: `Bearer ${owner.token}` }
  const tag = Date.now().toString(36)
  const tpl = await (await request.post(`${API}/plan-sets`, { headers: h, data: { title: `نموذج ${tag}`, isTemplate: true } })).json()
  const set = await (await request.post(`${API}/plan-sets`, { headers: h, data: { title: `سيت ${tag}` } })).json()
  await uiLogin(page, owner)
  await page.goto(`./plan-sets/${set.id}`)
  await expect(page.getByRole('button', { name: /PDF التصاميم/ })).toBeVisible()
  await page.getByText('📋 خذ من نموذج جاهز').click()
  await expect(page.locator('#ps-copy-from option', { hasText: `نموذج ${tag}` })).toHaveCount(1)
  await page.getByRole('button', { name: '☆ خلّيه نموذج جاهز' }).click()
  await expect(page.getByRole('button', { name: '⭐ نموذج جاهز — شيله' })).toBeVisible()
  await page.getByRole('button', { name: '⭐ نموذج جاهز — شيله' }).click()
  await page.getByRole('button', { name: /عرض السعر/ }).click()
  await expect(page).toHaveURL(new RegExp(`quotations/project/new\\?planSet=${set.id}`))
  await expect(page.getByText('🖼️ صور العرض — ارفع أو ارفق صور')).toBeVisible()
  expect(tpl.id).toBeTruthy()
  expect(problems).toEqual([])
})
