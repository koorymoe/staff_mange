import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type SalesOpportunity } from '../api'
import { useSession } from '../session'
import BookingCodeChip from '../components/BookingCodeChip'
import PhoneActions from '../components/PhoneActions'

/**
 * ═══ 💡 فرص البيع (ماتركس) ═══
 *
 * قواعد بس من بيانات الحجوزات: خدمة مكمّلة للي عنده، وصيانة سنوية
 * مستحقة (آخر إنجاز قبل ١١–١٣ شهر). بلا ضمان — ماكو بيانات ضمان
 * بالحجوزات أصلاً.
 *
 * ⚠️ الحارس نفس الخادم بالضبط: ADMIN/OWNER أو صلاحية sales_booking —
 * غيرهم ما نطلب أصلاً.
 */
type ReasonFilter = 'ALL' | SalesOpportunity['reason']

export default function SalesOpportunitiesPage() {
  const { employee, permissions } = useSession()
  const allowed = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'
    || (permissions ?? []).includes('sales_booking')
  const [rows, setRows] = useState<SalesOpportunity[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [reason, setReason] = useState<ReasonFilter>('ALL')
  const [reload, setReload] = useState(0)

  useEffect(() => {
    if (!allowed) return
    let alive = true
    api.getSalesOpportunities()
      .then((r) => { if (alive) { setRows(r); setError('') } })
      .catch((e: Error) => { if (alive) setError(e.message) })
      .finally(() => { if (alive) setLoading(false) })
    return () => { alive = false }
  }, [allowed, reload])

  const counts = useMemo(() => ({
    ALL: rows.length,
    CROSS_SELL: rows.filter((r) => r.reason === 'CROSS_SELL').length,
    MAINTENANCE_DUE: rows.filter((r) => r.reason === 'MAINTENANCE_DUE').length,
  }), [rows])
  const shown = reason === 'ALL' ? rows : rows.filter((r) => r.reason === reason)

  if (!allowed) {
    return <div className="p-6 text-center text-slate-500">هاي الشاشة للمبيعات والإدارة بس.</div>
  }

  const tabs: { key: ReasonFilter; label: string }[] = [
    { key: 'ALL', label: 'الكل' },
    { key: 'MAINTENANCE_DUE', label: '🔧 صيانة سنوية مستحقة' },
    { key: 'CROSS_SELL', label: '➕ خدمة مكمّلة' },
  ]

  return (
    <div className="space-y-4 p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h1 className="text-xl font-bold text-[#0f2040]">💡 فرص البيع</h1>
          <p className="text-sm text-slate-500">
            زبائن ممكن نتصل بيهم: صيانة سنوية فات عليها ١١–١٣ شهر، أو خدمة تكمّل الي عندهم.
          </p>
        </div>
        <button
          onClick={() => { setLoading(true); setReload((n) => n + 1) }}
          className="rounded-lg border border-slate-200 bg-white px-3 py-1.5 text-sm hover:bg-slate-50"
        >
          تحديث
        </button>
      </div>

      <div className="flex flex-wrap gap-2">
        {tabs.map((t) => (
          <button
            key={t.key}
            onClick={() => setReason(t.key)}
            className={`rounded-full px-3 py-1 text-sm font-semibold ${reason === t.key ? 'bg-brand-600 text-white' : 'bg-slate-100 text-slate-700 hover:bg-slate-200'}`}
          >
            {t.label} ({counts[t.key]})
          </button>
        ))}
      </div>

      {error && <div className="rounded-lg bg-red-50 p-3 text-sm text-red-700">{error}</div>}
      {loading ? (
        <div className="p-6 text-center text-slate-400">جاري الحساب…</div>
      ) : shown.length === 0 ? (
        <div className="rounded-xl border border-dashed border-slate-200 p-6 text-center text-slate-500">ماكو فرص حالياً.</div>
      ) : (
        <div className="grid gap-3 md:grid-cols-2">
          {shown.map((o) => (
            <div key={`${o.reason}-${o.customerId}-${o.suggestedService}`} className="rounded-xl border border-slate-200 bg-white p-4 shadow-sm">
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span className={`rounded-full px-2.5 py-0.5 text-xs font-bold ${o.reason === 'MAINTENANCE_DUE' ? 'bg-amber-100 text-amber-800' : 'bg-emerald-100 text-emerald-800'}`}>
                  {o.reasonLabel}
                </span>
                <BookingCodeChip code={o.customerCode} title="انسخ كود الزبون" className="font-mono text-xs text-slate-500" />
              </div>
              <div className="mt-2 font-bold text-[#0f2040]">{o.customerName}</div>
              <div className="mt-1 flex flex-wrap items-center gap-2 text-sm text-slate-600">
                <span dir="ltr">{o.customerPhone}</span>
                <PhoneActions phone={o.customerPhone} />
              </div>
              <div className="mt-2 text-sm text-slate-700">
                عنده: <b>{o.basedOnService}</b> · آخر إنجاز {new Date(o.lastBookingCompleted).toLocaleDateString('ar-IQ')}
              </div>
              <div className="mt-1 text-sm text-slate-700">نقترح: <b>{o.suggestedService}</b></div>
              <div className="mt-3">
                <Link
                  to={`/sales?customerId=${encodeURIComponent(o.customerId)}`}
                  className="inline-block rounded-lg bg-brand-600 px-3 py-1.5 text-sm font-bold text-white hover:bg-brand-700"
                >
                  احجز
                </Link>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
