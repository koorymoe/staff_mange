import { test, expect } from '@playwright/test'
import { API, prepare, uiLogin } from './helpers'

// بلاغ #11 (11-07): فني لغته أوردو يشوف شاشاته بالأوردو، وبموبايل صغير الشريط العلوي
// ما يطلع برا الشاشة (چان يطلع ٢٢٠px والجرس ما ينوصل).
test('فني أوردو: الجرد بالأوردو والشريط يكفي موبايل ٣٧٥', async ({ page, request }) => {
  test.setTimeout(180_000)
  await page.setViewportSize({ width: 375, height: 740 })
  const session = await prepare(request, 'TECHNICIAN')
  const h = { Authorization: `Bearer ${session.token}` }
  const r = await request.put(`${API}/me/language`, { headers: h, data: { language: 'ur' } })
  expect(r.ok(), await r.text()).toBeTruthy()
  // (11-12) وقت الدوام (٨–٤) الشاشة تطلب تسجيل الحضور قبل أي شي — چان الاختبار يفشل بس بالنهار.
  // نسجّل حضوره أول حتى نوصل للجرد بأي ساعة.
  const ci = await request.post(`${API}/attendance/checkin`, { headers: h, data: {} })
  expect(ci.ok() || ci.status() === 409, `تسجيل الحضور: ${ci.status()} ${await ci.text()}`).toBeTruthy()
  await uiLogin(page, { ...session, employee: { ...(session.employee as object), language: 'ur' } })

  await page.goto('./my-inventory')
  await expect(page.locator('html')).toHaveAttribute('lang', 'ur')
  await expect(page.getByText('اوزار اور آلات کی جانچ').first()).toBeVisible({ timeout: 15_000 })

  // كل أزرار الشريط العلوي داخل الشاشة
  const out = await page.evaluate(() =>
    [...document.querySelectorAll('header button')]
      .map((b) => b.getBoundingClientRect())
      .filter((x) => x.width > 0 && (x.left < -1 || x.right > innerWidth + 1)).length)
  expect(out, 'أزرار الشريط طالعة برا الشاشة').toBe(0)
})
