import { test, expect, type APIRequestContext, type Page } from '@playwright/test'

// ٧أ (11-04): كل دور يدخل ويفتح **كل رابط بقائمته** — أي خطأ صفحة أو ٥xx أو ٤٠٣ يفشّل.
// ٤٠٣ مهم بالذات: بالنظام ٤٠٣ ينسجّل «مخالفة» على الموظف، فشاشة بقائمته ترجّع ٤٠٣ = ظلم.
// الحسابات تنسوى بالسيرفر (المالك يسويها)، والنوافذ الإجبارية (المبادئ، الاستمارة) تتعبّى بالسيرفر.
const API = process.env.E2E_API ?? 'http://127.0.0.1:4000/api'
const OWNER = { username: process.env.E2E_USER ?? 'owner', password: process.env.E2E_PASS ?? 'Owner-pass-123' }
const PASS = 'E2e-pass-1234'
const ROLES = ['MONITOR', 'FINANCE', 'TECHNICIAN', 'HR_COORDINATOR', 'QUALITY_ENGINEER', 'SALES'] as const
const RUN = Date.now().toString(36)

// الدخول محدود ٨ بالدقيقة لكل عنوان — إذا انضرب الحد ننتظر وتعيد (هذا سلوك صحيح مو خلل).
async function login(req: APIRequestContext, username: string, password: string) {
  for (let i = 0; i < 3; i++) {
    const r = await req.post(`${API}/auth/login`, { data: { username, password } })
    if (r.status() === 429) { await new Promise((ok) => setTimeout(ok, 62_000)); continue }
    expect(r.ok(), `دخول ${username}: ${r.status()}`).toBeTruthy()
    // الرد مسطّح: حقول الموظف + token + realm (مثل api.login بالواجهة)
    const { token, realm, ...employee } = (await r.json()) as { token: string; realm?: string }
    return { token, realm: realm ?? 'staff', employee }
  }
  throw new Error('حد الدخول')
}
let ownerToken = ''

async function prepare(req: APIRequestContext, role: string) {
  if (!ownerToken) ownerToken = (await login(req, OWNER.username, OWNER.password)).token
  const owner = ownerToken
  const username = `e2e_${role.toLowerCase()}_${RUN}`
  const phone = '079' + String(Math.floor(Math.random() * 1e8)).padStart(8, '0')
  const c = await req.post(`${API}/employees`, {
    headers: { Authorization: `Bearer ${owner}` },
    data: { name: `اختبار ${role} تلقائي`, username, password: PASS, role, phone, certificate: 'اختبار', position: 'اختبار' },
  })
  expect(c.ok(), `إنشاء ${role}: ${await c.text()}`).toBeTruthy()
  const session = await login(req, username, PASS)
  const t = session.token
  const h = { Authorization: `Bearer ${t}` }
  const st = await (await req.get(`${API}/privacy-policy/status`, { headers: h })).json()
  if (st.needsAcceptance) await req.post(`${API}/privacy-policy/accept`, { headers: h, data: { hash: st.hash ?? '' } })
  const gate = await (await req.get(`${API}/me/work-gate`, { headers: h })).json()
  if (gate.survey) {
    const s = await req.post(`${API}/me/profile-survey`, {
      headers: h,
      data: { major: 'هندسة حاسبات', degree: 'بكالوريوس', selfSalaryIQD: null, monthlyHours: null, currentJob: 'اختبار', companyRating: 4, companyNote: '', contractEnd: '' },
    })
    expect(s.ok(), await s.text()).toBeTruthy()
  }
  const g2 = await (await req.get(`${API}/me/work-gate`, { headers: h })).json()
  if (g2.missingDay) await req.post(`${API}/me/daily-report`, { headers: h, data: { day: g2.missingDay, body: 'تقرير اختبار تلقائي' } })
  return session
}

// نفس الي تسويه شاشة الدخول بعد النجاح — بلا دخول ثاني يستهلك من الحد.
// ⚠️ الجلسة تنحط **قبل** ما تنفتح الصفحة: لو انحطت بعد، طلب خلفي من شاشة الدخول
// (بلا جلسة، يرجع ٤٠١) يوصل بعدها والواجهة تفهمه «الجلسة انتهت» وتطلّع الحساب.
async function uiLogin(page: Page, s: { token: string; realm: string; employee: unknown }) {
  await page.addInitScript((x) => {
    if (!localStorage.getItem('authToken')) {
      localStorage.setItem('authToken', x.token)
      localStorage.setItem('authRealm', x.realm)
      localStorage.setItem('currentEmployee', JSON.stringify(x.employee))
    }
  }, s)
  await page.goto('./')
  await page.waitForLoadState('networkidle')
}

for (const role of ROLES) {
  test(`${role}: كل روابط القائمة تنفتح بلا خطأ ولا ٤٠٣`, async ({ page, request }) => {
    test.setTimeout(240_000)
    const session = await prepare(request, role)
    const problems: string[] = []
    let where = '/'
    page.on('pageerror', (e) => problems.push(`[${where}] pageerror: ${e.message}`))
    page.on('response', (r) => {
      if (!r.url().includes('/api/')) return
      if (r.status() >= 500 || r.status() === 403) problems.push(`[${where}] ${r.status()} ${r.request().method()} ${r.url().replace(/^.*\/api/, '/api')}`)
    })
    await uiLogin(page, session)
    const base = new URL(page.url()).pathname.replace(/\/[^/]*$/, '/')
    const hrefs = await page.locator(`a[href^="${base}"]`).evaluateAll((as) => [...new Set(as.map((a) => (a as HTMLAnchorElement).getAttribute('href') ?? ''))])
    const links = hrefs.filter((h) => h && !h.includes('logout')).slice(0, 40)
    expect(links.length, `${role}: القائمة فاضية`).toBeGreaterThan(0)
    for (const href of links) {
      where = href
      await page.goto(href)
      await page.waitForLoadState('networkidle')
    }
    // ٧و (11-05): الشاشات الموحّدة بتبويبات — كل تبويب يطلع لازم ينفتح بلا ٤٠٣.
    // نبدي من «الإجازات» (مفتوحة للكل) ومن أي شاشة موحّدة موجودة بقائمته.
    const hubs = ['/leaves', '/staff-leaves', '/suppliers', '/tech-suppliers', '/supplier-relations', '/quality-desk']
    const starts = hubs.filter((h) => h === '/leaves' || links.some((l) => l.endsWith(h)))
    for (const h of starts) {
      where = `hub ${h}`
      await page.goto(base.replace(/\/$/, '') + h)
      await page.waitForLoadState('networkidle')
      const n = await page.locator('main [role=tablist] [role=tab]').count()
      for (let i = 0; i < n; i++) {
        where = `hub ${h} tab#${i}`
        await page.locator('main [role=tablist] [role=tab]').nth(i).click()
        await page.waitForLoadState('networkidle')
      }
    }
    expect(problems, `${role}:\n${problems.join('\n')}`).toEqual([])
  })
}
