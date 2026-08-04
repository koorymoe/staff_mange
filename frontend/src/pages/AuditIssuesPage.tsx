import { useEffect, useState } from 'react'
import { api, type AuditIssue } from '../api'

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
        <h1 className="text-2xl font-bold text-white">💸 بلاغات أخطاء التدقيق</h1>
        <p className="mt-1 text-sm text-blue-200">
          أشّرها المحاسب وهو يدقق الحجوزات — تحتاج متابعة منك. ({open.length} مفتوح)
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
                ✔ تابعته — أغلق البلاغ
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
