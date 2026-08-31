import { useEffect, useState } from 'react'
import { api, type AuditIssue } from '../api'
import { useSession } from '../session'

/**
 * بلاغات أخطاء التدقيق.
 *
 * المحاسب يأشّرها وهو يدقق، والنظام يوجّهها:
 *   - «المبلغ غير مطابق» → الرقابة والجودة
 *   - «خطأ بالسعر»        → الرقابة والإداري
 *
 * كل واحد يشوف الي يخصه بس — الفلترة تصير بالسيرفر حسب دوره.
 */

const fmt = (n: number | null) => (n == null ? '—' : n.toLocaleString('en-IQ') + ' د.ع')

/** ⚠️ `embedded`: نفس الشاشة بالضبط بلا ترويستها — تنضمّ بمكتب
 *  المراقب. **ما ننسخ المحتوى**: نسختان تفترقان بأول تصحيح، فالمراقب
 *  يشوف صفاً بشاشة ومحلولاً بالثانية ويفقد الثقة بالاثنتين. */
interface EmbeddedProps { embedded?: boolean }

export default function AuditIssuesPage({ embedded }: EmbeddedProps = {}) {
  // ═══ المحاسب مو مراقب ═══
  // نفس الشاشة، بس معناها يختلف: عند المحاسب **صادر** — هو الي أشّر
  // الأخطاء ويتابع شنو صار بيها. وعند المراقب **وارد للتدقيق** — يروح
  // يتأكد من الليدر ليش عنده أخطاء. السيرفر يفلتر (المحاسب يشوف بلاغاته
  // بس)، وهنا نغيّر العنوان والأعمدة حتى الواجهة تگول نفس الكلام.
  const { employee } = useSession()
  const asAccountant = employee?.role === 'FINANCE'
  const [items, setItems] = useState<AuditIssue[]>([])
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState<string | null>(null)

  const load = () => {
    api.getAuditIssues()
      .then(setItems)
      .catch(() => setItems([]))
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  const resolve = async (i: AuditIssue) => {
    setBusy(i.id)
    try {
      await api.resolveAuditIssue(i.id)
      load()
    } finally {
      setBusy(null)
    }
  }

  const open = items.filter((i) => i.status === 'OPEN')

  return (
    <div dir="rtl" className="space-y-6">
      {!embedded && (
        <div
          className="relative overflow-hidden rounded-2xl p-6 shadow-md"
          style={{ background: 'linear-gradient(135deg, #1a3a5c 0%, #24507e 55%, #2f6ba8 100%)' }}
        >
          <span
            aria-hidden
            className="pointer-events-none absolute -left-16 -top-24 h-64 w-64 rounded-full opacity-20"
            style={{ background: 'radial-gradient(circle, #c8a45a 0%, transparent 70%)' }}
          />
          <div className="relative flex flex-wrap items-end justify-between gap-4">
            <div>
              <h1 className="text-2xl font-black text-white">
                {asAccountant ? '💸 أخطاء الفواتير الي أشّرتها' : '💸 بلاغات أخطاء التدقيق'}
              </h1>
              <p className="mt-1 max-w-xl text-sm text-blue-100">
                {asAccountant
                  ? 'هاي البلاغات الي أرسلتها إنت وأنت تدقق — تابع شنو صار بيها.'
                  : 'أشّرها المحاسب وهو يدقق الحجوزات — تدقّق بيها على الليدر.'}
              </p>
            </div>
            <div className="rounded-xl bg-amber-400/20 px-4 py-2 text-center ring-1 ring-amber-200/40 backdrop-blur">
              <p className="text-2xl font-black leading-none text-amber-100">{open.length}</p>
              <p className="mt-1 text-[11px] text-amber-50">لسه مفتوح</p>
            </div>
          </div>
        </div>
      )}

      {loading && <p className="text-center text-slate-400">جاري التحميل...</p>}

      <div className="space-y-3">
        {items.map((i) => (
          <div
            key={i.id}
            /* شريط بلون البلاغ: المفتوح يشدّ العين والمنحسم يهدى — الشاشة
               تطول بالوقت وكلها تبين نفس الشي بلا هذا. */
            className={`relative overflow-hidden rounded-xl border bg-white p-4 pr-5 shadow-sm transition hover:shadow-md ${
              i.status !== 'OPEN' ? 'border-slate-200 opacity-75'
                : i.kind === 'MISMATCH' ? 'border-amber-200' : 'border-red-200'
            }`}
          >
            <span
              aria-hidden
              className={`absolute inset-y-0 right-0 w-1.5 ${
                i.status !== 'OPEN' ? 'bg-slate-300' : i.kind === 'MISMATCH' ? 'bg-amber-400' : 'bg-red-400'
              }`}
            />
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="font-bold text-slate-800">
                  <span className="rounded-lg bg-slate-100 px-2 py-0.5 font-mono text-sm text-slate-700">{i.bookingCode}</span>
                  <span className="mr-2 text-sm font-normal text-slate-500">{i.customerName}</span>
                </p>
                <p className="mt-1 text-sm">
                  <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                    i.kind === 'MISMATCH' ? 'bg-amber-100 text-amber-800' : 'bg-red-100 text-red-800'
                  }`}>
                    {i.kindLabel}
                  </span>
                  <span className="mr-2 text-xs text-slate-400">← {i.routedTo}</span>
                </p>
                <p className="mt-2 text-sm text-slate-600">
                  المبلغ بالفاتورة: <b>{fmt(i.expectedAmount)}</b> · المسجّل بالنظام: <b>{fmt(i.actualAmount)}</b>
                </p>
                {i.note && <p className="mt-1 text-sm text-slate-600">📝 {i.note}</p>}
                <p className="mt-1 text-xs text-slate-400">
                  سجّله: {i.raisedByName} · {new Date(i.createdAt).toLocaleDateString('ar-IQ')}
                </p>
                {/* الليدر يظهر للمراقب بس — هو الي يروح يسأله. المحاسب
                    يعرف ليدره أصلاً من الفاتورة الي دققها. */}
                {!asAccountant && i.leaderName && (
                  <p className="mt-1 text-sm font-bold text-slate-700">👷 الليدر: {i.leaderName}</p>
                )}
              </div>
              <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                i.status === 'OPEN' ? 'bg-amber-100 text-amber-800' : 'bg-slate-200 text-slate-600'
              }`}>
                {i.status === 'OPEN' ? 'مفتوح' : 'انحسم'}
              </span>
            </div>

            {i.status === 'OPEN' && (
              <button disabled={busy === i.id} onClick={() => resolve(i)}
                className="mt-3 rounded-lg bg-emerald-700 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-800 disabled:opacity-50">
                {asAccountant ? '✔ انحلّت — أغلق البلاغ' : '✔ تأكدت من الليدر — أغلق البلاغ'}
              </button>
            )}
          </div>
        ))}
        {!loading && items.length === 0 && (
          <div className="rounded-2xl border border-dashed border-slate-200 bg-white p-10 text-center shadow-sm">
            <p className="text-4xl">✅</p>
            <p className="mt-3 font-bold text-slate-600">ماكو بلاغات</p>
            <p className="mx-auto mt-1 max-w-sm text-sm text-slate-400">
              {asAccountant
                ? 'ما أشّرت أي خطأ لحد هسه — البلاغ ينسجّل لمن تختار «غير مطابق» أو «خطأ بالسعر» بالتدقيق اليومي.'
                : 'ماكو خطأ أشّره المحاسب — لمن يأشّر واحد يوصلك هنا باسم الليدر صاحب الفاتورة.'}
            </p>
          </div>
        )}
      </div>
    </div>
  )
}
