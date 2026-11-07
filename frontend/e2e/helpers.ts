import { expect, type APIRequestContext, type Page } from '@playwright/test'

// مشترك بين اختبارات المتصفح (11-07): إنشاء حساب بدور، الدخول، وزرع الجلسة قبل فتح الصفحة.
export const API = process.env.E2E_API ?? 'http://127.0.0.1:4000/api'
export const OWNER = { username: process.env.E2E_USER ?? 'owner', password: process.env.E2E_PASS ?? 'Owner-pass-123' }
export const PASS = 'E2e-pass-1234'
export const RUN = Date.now().toString(36)

// الدخول محدود ٨ بالدقيقة لكل عنوان — إذا انضرب الحد ننتظر وتعيد (هذا سلوك صحيح مو خلل).
export async function login(req: APIRequestContext, username: string, password: string) {
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

export async function prepare(req: APIRequestContext, role: string) {
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
export async function uiLogin(page: Page, s: { token: string; realm: string; employee: unknown }) {
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

