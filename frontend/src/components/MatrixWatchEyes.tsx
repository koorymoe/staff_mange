import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type MatrixEyesReport } from '../api'
import MatrixNote from './MatrixNote'

// ═══ عيون الرقابة — المخزن، السيارات، رأي الزبائن، الآيتي، الطلبات المعلّقة ═══
// طلب (ع) 10-05: «لازم اكو رقابه بكل مكان» و«تروح للمراقب ويا تحليل ماتركس».
// كل بند: شنو الغلط بالأرقام، منو يخصه، وشنو يقترح ماتركس.

const ICON: Record<string, string> = { REVENUE: '💰', STOCK: '📦', VEHICLES: '🚗', CUSTOMERS: '⭐', IT: '🖥️', PENDING: '⏳' }
const SEV: Record<string, { cls: string; label: string }> = {
  HIGH: { cls: 'border-red-300 bg-red-50', label: '🔴 مهم' },
  MEDIUM: { cls: 'border-amber-300 bg-amber-50', label: '🟠 يحتاج متابعة' },
  LOW: { cls: 'border-slate-200 bg-slate-50', label: '🔵 للعلم' },
}

export default function MatrixWatchEyes({ initialTab = 'REVENUE' }: { initialTab?: string } = {}) {
  const [r, setR] = useState<MatrixEyesReport | null>(null)
  const [tab, setTab] = useState(initialTab)
  const [all, setAll] = useState(false)
  useEffect(() => { void api.getMatrixEyes().then(setR).catch(() => {}) }, [])
  if (!r) return <p className="text-xs text-slate-400">ماتركس يفحص…</p>
  const eye = r.eyes.find((e) => e.key === tab) ?? r.eyes[0]
  const items = all ? eye.items : eye.items.slice(0, 12)

  return (
    <div dir="rtl" className="space-y-3 rounded-2xl border border-slate-200 bg-slate-50/60 p-3 text-slate-800 sm:p-4">
      <div>
        <p className="text-base font-extrabold text-[#0f2040]">👁️ عيون الرقابة</p>
        <p className="text-[11px] text-slate-500">الفلوس والإيرادات، المخزن، السيارات، رأي الزبائن، الآيتي، والطلبات المعلّقة. ماتركس يفحصها كل يوم بالقواعد (بلا كلفة).</p>
      </div>
      <MatrixNote>{r.insights.join(' ')}</MatrixNote>
      <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-6">
        {r.eyes.map((e) => (
          <button key={e.key} type="button" onClick={() => { setTab(e.key); setAll(false) }}
            className={`rounded-xl border p-2.5 text-right transition ${tab === e.key ? 'border-[#0f2040] bg-white shadow' : 'border-slate-200 bg-white/70 hover:bg-white'}`}>
            <p className="text-sm font-bold">{ICON[e.key]} {e.title}</p>
            <p className="text-[11px] text-slate-500">{e.items.length ? `${e.items.length} بند${e.high ? ` · ${e.high} مهم` : ''}` : '✅ ماكو شي'}</p>
          </button>
        ))}
      </div>

      <p className="text-xs text-slate-600">{eye.summary}</p>
      <div className="space-y-2">
        {items.map((it, i) => (
          <div key={i} className={`rounded-xl border p-3 ${SEV[it.severity].cls}`}>
            <div className="flex flex-wrap items-start justify-between gap-2">
              <p className="text-sm font-bold text-slate-800">{it.title}</p>
              <span className="shrink-0 text-[11px] font-bold">{SEV[it.severity].label}</span>
            </div>
            <p className="mt-1 text-xs text-slate-700">{it.detail}</p>
            <p className="mt-1 text-xs text-violet-800">🤖 {it.advice}</p>
            <p className="mt-1 text-[11px] text-slate-500">
              {it.ownerName && <>يخص: <b>{it.ownerName}</b> · </>}
              {it.at && new Date(it.at).toLocaleDateString('en-GB')}
              {it.bookingId && <> · <Link to={`/bookings?focus=${it.bookingId}`} className="text-sky-700 underline">افتح الحجز</Link></>}
            </p>
          </div>
        ))}
        {eye.items.length > 12 && !all && (
          <button type="button" onClick={() => setAll(true)} className="text-xs font-bold text-sky-700 hover:underline">عرض كل البنود ({eye.items.length})</button>
        )}
      </div>

      {tab === 'CUSTOMERS' && r.customers.length > 0 && (
        <div className="rounded-xl border border-slate-200 bg-white p-3">
          <p className="mb-2 text-xs font-extrabold text-slate-700">رضا الزبائن عن كل موظف (آخر ٦٠ يوم){r.teamHappyPct != null && <span className="font-normal text-slate-500"> — الفريق {r.teamHappyPct}%</span>}</p>
          <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
            {r.customers.map((c) => (
              <div key={c.employeeId} className="rounded-lg border border-slate-200 p-2 text-xs">
                <div className="flex items-center justify-between"><b>{c.name}</b><b style={{ color: c.happyPct == null ? '#94a3b8' : c.happyPct >= 80 ? '#059669' : c.happyPct >= 60 ? '#d97706' : '#dc2626' }}>{c.happyPct == null ? '—' : `${c.happyPct}%`}</b></div>
                <p className="text-slate-500">{c.happy} راضي · {c.unhappy} عنده ملاحظة · {c.complaints} شكوى{c.avgRating != null ? ` · تقييم ${c.avgRating.toFixed(1)}` : ''}</p>
                {c.notes.map((n, i) => <p key={i} className="text-slate-600">{n}</p>)}
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}
