import { test, expect } from '@playwright/test'
import { prepare, uiLogin } from './helpers'

// ٧أ (11-04): كل دور يدخل ويفتح **كل رابط بقائمته** — أي خطأ صفحة أو ٥xx أو ٤٠٣ يفشّل.
// ٤٠٣ مهم بالذات: بالنظام ٤٠٣ ينسجّل «مخالفة» على الموظف، فشاشة بقائمته ترجّع ٤٠٣ = ظلم.
// الحسابات تنسوى بالسيرفر (المالك يسويها)، والنوافذ الإجبارية (المبادئ، الاستمارة) تتعبّى بالسيرفر.
const ROLES = ['MONITOR', 'FINANCE', 'TECHNICIAN', 'HR_COORDINATOR', 'QUALITY_ENGINEER', 'SALES'] as const

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
    const hubs = ['/leaves', '/staff-leaves', '/suppliers', '/tech-suppliers', '/supplier-relations', '/quality-desk', '/staff-score']
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
