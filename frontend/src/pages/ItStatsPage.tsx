import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, IT_KIND_LABELS, IT_LOG_LABELS, IT_STATUS_LABELS, type ItAssetKind, type ItAssetStatus, type ItStats } from '../api'
import { useSession } from '../session'
import ReplacementSuggestionsPanel from '../components/ReplacementSuggestionsPanel'

// ═══ إحصائيات تقنية المعلومات ═══
// شكد جهاز عدنا ومن أي نوع، شنو بالتصليح، شنو ضمانه قرب يخلص، وشكد
// صرفنا على التصليح بآخر ٣ أشهر.

function Card({ label, value, tone = 'text-slate-800' }: { label: string; value: string | number; tone?: string }) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-4">
      <p className="text-xs text-slate-500">{label}</p>
      <p className={`mt-1 text-2xl font-extrabold tabular-nums ${tone}`}>{value}</p>
    </div>
  )
}

export default function ItStatsPage() {
  const { employee, permissions } = useSession()
  const canManage = employee?.role === 'ADMIN' || permissions.includes('it_assets')
  const [s, setS] = useState<ItStats | null>(null)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    api.getItStats().then((r) => { if (alive) setS(r) }).catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر جلب الإحصائيات') })
    return () => { alive = false }
  }, [])

  if (err) return <p className="rounded-lg bg-red-50 p-4 text-red-600">{err}</p>
  if (!s) return <p className="text-slate-400">جاري التحميل…</p>

  const fmt = (n: number) => n.toLocaleString('en-US')
  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">📊 إحصائيات تقنية المعلومات</h2>
          <p className="text-sm text-slate-500">حالة أجهزة الشركة بلمحة.</p>
        </div>
        {canManage && <Link to="/it-assets" className="rounded-xl bg-brand-50 px-4 py-2 text-sm font-bold text-brand-700 hover:bg-brand-100">🖥️ الجرد ←</Link>}
      </div>

      <div className="grid grid-cols-2 gap-3 md:grid-cols-5">
        <Card label="كل الأجهزة" value={fmt(s.total)} tone="text-brand-700" />
        <Card label="شغّال" value={fmt(s.byStatus.ACTIVE || 0)} tone="text-emerald-700" />
        <Card label="بالتصليح" value={fmt(s.byStatus.REPAIR || 0)} tone={s.byStatus.REPAIR ? 'text-amber-700' : 'text-slate-800'} />
        <Card label="تصليحات آخر ٩٠ يوم" value={fmt(s.repairCount90d)} />
        <Card label="كلفة التصليح (٩٠ يوم)" value={`${fmt(Math.round(s.repairCost90d))} د.ع`} tone="text-red-700" />
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <section className="rounded-2xl border border-slate-200 bg-white p-4">
          <h3 className="mb-3 font-extrabold text-[#0f2040]">حسب النوع</h3>
          {s.total === 0 ? <p className="text-sm text-slate-400">ماكو أجهزة مسجّلة.</p> : (
            <ul className="space-y-2">
              {Object.entries(s.byKind).sort((a, b) => b[1] - a[1]).map(([k, n]) => (
                <li key={k}>
                  <div className="flex justify-between text-sm"><span>{IT_KIND_LABELS[k as ItAssetKind] || k}</span><b className="tabular-nums">{fmt(n)}</b></div>
                  <div className="mt-1 h-2 rounded-full bg-slate-100"><div className="h-2 rounded-full bg-brand-500" style={{ width: `${(n / s.total) * 100}%` }} /></div>
                </li>
              ))}
            </ul>
          )}
          <div className="mt-4 flex flex-wrap gap-2">
            {Object.entries(s.byStatus).map(([k, n]) => (
              <span key={k} className="rounded-full bg-slate-100 px-2.5 py-1 text-xs text-slate-600">{IT_STATUS_LABELS[k as ItAssetStatus] || k}: <b>{fmt(n)}</b></span>
            ))}
            {s.unassigned > 0 && <span className="rounded-full bg-amber-50 px-2.5 py-1 text-xs text-amber-800">شغّال بلا مكان ولا مستخدم: <b>{fmt(s.unassigned)}</b></span>}
          </div>
        </section>

        <section className="rounded-2xl border border-slate-200 bg-white p-4">
          <h3 className="mb-3 font-extrabold text-[#0f2040]">⚠️ يحتاج انتباه</h3>
          {s.inRepair.length === 0 && s.warrantySoon.length === 0 ? <p className="text-sm text-slate-400">ماكو شي يحتاج انتباه ✅</p> : (
            <ul className="space-y-1.5 text-sm">
              {s.inRepair.map((a) => (
                <li key={`r${a.id}`} className="flex justify-between gap-2"><span>🔧 {a.name}</span><span className="text-xs text-amber-700">بالتصليح</span></li>
              ))}
              {s.warrantySoon.map((a) => (
                <li key={`w${a.id}`} className="flex justify-between gap-2"><span>🛡️ {a.name}</span><span className="text-xs text-slate-500">الضمان {a.warrantyUntil?.slice(0, 10)}</span></li>
              ))}
            </ul>
          )}
          {/* ماتركس: جهاز تصلّح ٣ مرات فأكثر بـ١٨٠ يوم — يطلع بس لصاحب it_assets أو المدير */}
          <div className="mt-3 border-t border-slate-100 pt-3">
            <p className="mb-2 text-xs font-bold text-slate-500">🔁 اقتراح استبدال</p>
            <ReplacementSuggestionsPanel part="it" />
          </div>
        </section>
      </div>

      <section className="rounded-2xl border border-slate-200 bg-white p-4">
        <h3 className="mb-3 font-extrabold text-[#0f2040]">آخر الأعمال</h3>
        {s.recentLogs.length === 0 ? <p className="text-sm text-slate-400">ماكو سجل بعد.</p> : (
          <ul className="divide-y divide-slate-100">
            {s.recentLogs.map((l) => (
              <li key={l.id} className="py-2 text-sm">
                <div className="flex justify-between gap-2 text-xs text-slate-500">
                  <span><b className="text-slate-700">{l.assetName}</b> · {IT_LOG_LABELS[l.kind] || l.kind}</span>
                  <span>{new Date(l.createdAt).toLocaleDateString('ar-IQ')} · {l.employeeName || '—'}</span>
                </div>
                <p className="mt-0.5 text-slate-700">{l.note}</p>
              </li>
            ))}
          </ul>
        )}
      </section>
    </div>
  )
}
