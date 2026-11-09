import * as Sentry from '@sentry/react'

// ═══ Sentry (دازاي 11-05) — يلزم أخطاء تطبيق الزبائن. يتحمّل بس إذا VITE_SENTRY_DSN موجود. ═══
// ⚠️ قاعدة صاحب النظام: ماكو أسماء موظفين/زبائن ولا مبالغ ولا تلفونات تطلع لجهة خارجية.
// فكل نص ورابط يتنظف هنا قبل الإرسال، وماكو بيانات مستخدم ولا كوكيز ولا تسجيل شاشة (replay).

const reEmail = /[\w.+-]+@[\w-]+(\.[\w-]+)+/g
const reUUID = /[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/gi
const reDigits = /\d{4,}/g
const reArabic = /[؀-ۿ][؀-ۿ\s،]*/g

export function scrub(s?: string): string | undefined {
  if (!s) return s
  return s.replace(reEmail, '[email]').replace(reUUID, '[id]').replace(reDigits, '[n]').replace(reArabic, '[ar] ')
}

// الرابط: بلا query (بيه بحث وأسماء)، ومعرّفات المسار تتبدّل
export function scrubUrl(u?: string): string | undefined {
  if (!u) return u
  try {
    const url = new URL(u, location.origin)
    const path = url.pathname.split('/').map((p) => (/\d|[0-9a-f]{16,}|%/i.test(p) ? ':id' : p)).join('/')
    return url.origin + path
  } catch {
    return '[url]'
  }
}

export function initSentry(dsn: string, app: string) {
  Sentry.init({
    dsn,
    environment: 'production',
    maxBreadcrumbs: 20,
    initialScope: { tags: { app } },
    beforeBreadcrumb(b) {
      if (b.category === 'console') return null // console بيه بيانات
      b.message = scrub(b.message)
      if (b.data) {
        const url = typeof b.data.url === 'string' ? scrubUrl(b.data.url) : undefined
        b.data = url ? { url, method: b.data.method, status_code: b.data.status_code } : undefined
      }
      return b
    },
    beforeSend(e) {
      delete e.user
      if (e.request) {
        e.request = { url: scrubUrl(e.request.url) }
      }
      e.message = scrub(e.message)
      for (const x of e.exception?.values ?? []) {
        x.value = scrub(x.value)
        for (const f of x.stacktrace?.frames ?? []) {
          delete f.context_line
          delete f.pre_context
          delete f.post_context
          delete f.vars
        }
      }
      delete e.extra
      return e
    },
  })
}

export const captureException = (err: unknown) => Sentry.captureException(err)
