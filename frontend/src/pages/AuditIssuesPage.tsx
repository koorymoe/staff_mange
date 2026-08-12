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

export default function AuditIssuesPage() {
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
      <div className="rounded-2xl p-6 shadow-sm" style={{ backgroundColor: '#1a3a5c' }}>
        <h1 className="text-2xl font-bold text-white">
          {asAccountant ? '💸 أخطاء الفواتير الي أشّرتها' : '💸 بلاغات أخطاء التدقيق'}
        </h1>
        <p className="mt-1 text-sm text-blue-200">
          {asAccountant
            ? `هاي البلاغات الي أرسلتها إنت وأنت تدقق — تابع شنو صار بيها. (${open.length} لسه مفتوح)`
            : `أشّرها المحاسب وهو يدقق الحجوزات — تدقّق بيها على الليدر. (${open.length} مفتوح)`}
        </p>
      </div>

      {loading && <p className="text-center text-slate-400">جاري التحميل...</p>}

      <div className="space-y-3">
        {items.map((i) => (
          <div key={i.id} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
            <div className="flex flex-wrap items-start justify-between gap-3">
              <div>
                <p className="font-bold text-slate-800">
                  {i.bookingCode}
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
                className="mt-3 rounded-lg bg-emerald-600 px-4 py-2 text-sm font-medium text-white hover:bg-emerald-700 disabled:opacity-50">
                {asAccountant ? '✔ انحلّت — أغلق البلاغ' : '✔ تأكدت من الليدر — أغلق البلاغ'}
              </button>
            )}
          </div>
        ))}
        {!loading && items.length === 0 && (
          <p className="rounded-2xl bg-white p-8 text-center text-slate-400 shadow-sm">ماكو بلاغات</p>
        )}
      </div>
    </div>
  )
}
