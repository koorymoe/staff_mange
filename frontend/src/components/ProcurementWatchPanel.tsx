import { useEffect, useState } from 'react'
import { api, type ProcurementWatch, type ProcDecision } from '../api'

// ═══ 📦 شغل المخازن — مكتب المراقب (قرار (ع) 10-10) ═══
// «كل شغلة يشتغلها أبو الكميات لازم تكون موجودة بالنظام» — نفس البنود الي
// ماتركس يقيّمه ويذكّره عليها: الرد على الطلبات، متابعة السيارات، النواقص، الأسطول.
const hours = (a: string, b?: string | null) => ((b ? new Date(b) : new Date()).getTime() - new Date(a).getTime()) / 3600000
const age = (h: number) => (h < 24 ? `${Math.round(h)} ساعة` : `${Math.round(h / 24)} يوم`)

function Section({ title, rows, pendingWord }: { title: string; rows: ProcDecision[]; pendingWord: string }) {
  const open = rows.filter((r) => !r.decidedAt)
  const done = rows.filter((r) => r.decidedAt)
  const avg = done.length ? done.reduce((s, r) => s + hours(r.createdAt, r.decidedAt), 0) / done.length : null
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-3">
      <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
        <b className="text-sm text-[#0f2040]">{title}</b>
        <span className="text-xs text-slate-500">{open.length} {pendingWord} · {done.length} انحسم{avg != null && ` · معدّل الرد ${age(avg)}`}</span>
      </div>
      {open.length === 0 ? <p className="text-xs text-emerald-700">✅ ماكو شي معلّق</p> : (
        <div className="space-y-1">
          {open.map((r) => {
            const h = hours(r.createdAt)
            return (
              <p key={r.id} className={`flex justify-between gap-2 rounded-lg px-2 py-1 text-xs ${h > 48 ? 'bg-red-50 text-red-800' : h > 24 ? 'bg-amber-50 text-amber-800' : 'bg-slate-50 text-slate-700'}`}>
                <span>{r.label}</span><b className="whitespace-nowrap">{age(h)}</b>
              </p>
            )
          })}
        </div>
      )}
    </div>
  )
}

export default function ProcurementWatchPanel() {
  const [d, setD] = useState<ProcurementWatch | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => {
    let alive = true
    api.getProcurementWatch().then((r) => { if (alive) setD(r) }).catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [])
  if (err) return <p className="text-sm text-red-600">{err}</p>
  if (!d) return <p className="text-sm text-slate-400">جاري التحميل…</p>
  const fullDays = d.vehicleDays.filter((x) => x.total > 0 && x.rated >= x.total).length
  return (
    <div dir="rtl" className="space-y-3">
      <p className="text-xs text-slate-500">البنود من {d.since} وطالع — ماتركس يقيّم أبو الكميات عليها ويذكّره، والي يتأخر +٤٨ ساعة يوصلك.</p>
      <div className="grid gap-3 lg:grid-cols-2">
        <Section title="📋 طلبات المواد" rows={d.materials} pendingWord="معلّق" />
        <Section title="🧰 طلبات الأدوات" rows={d.tools} pendingWord="معلّق" />
        <Section title="📦 نواقص الجرد" rows={d.shortages} pendingWord="مفتوح" />
        <div className="rounded-2xl border border-slate-200 bg-white p-3">
          <div className="mb-2 flex items-center justify-between">
            <b className="text-sm text-[#0f2040]">🚗 متابعة السيارات</b>
            <span className="text-xs text-slate-500">{fullDays} من {d.vehicleDays.length} يوم كاملة</span>
          </div>
          <p className={`mb-2 rounded-lg px-2 py-1 text-xs ${d.unratedToday.length ? 'bg-amber-50 text-amber-800' : 'bg-emerald-50 text-emerald-700'}`}>
            {d.unratedToday.length ? `اليوم بعد ما انقيّمت: ${d.unratedToday.join('، ')}` : '✅ اليوم كل السيارات انقيّمت'}
          </p>
          <div className="flex flex-wrap gap-1">
            {d.vehicleDays.map((x) => (
              <span key={x.day} title={`${x.day}: ${x.rated} من ${x.total}`}
                className={`rounded px-1.5 py-0.5 text-[10px] font-bold ${x.rated >= x.total ? 'bg-emerald-100 text-emerald-800' : x.rated > 0 ? 'bg-amber-100 text-amber-800' : 'bg-red-100 text-red-700'}`}>
                {x.day.slice(5)} {x.rated}/{x.total}
              </span>
            ))}
          </div>
        </div>
      </div>
      <div className="rounded-2xl border border-slate-200 bg-white p-3">
        <b className="text-sm text-[#0f2040]">🔧 صيانة ووثائق فات موعدها</b>
        {d.fleetOverdue.length === 0 ? <p className="mt-1 text-xs text-emerald-700">✅ ماكو</p>
          : <ul className="mt-1 space-y-0.5 text-xs text-red-700">{d.fleetOverdue.map((f) => <li key={f}>• {f}</li>)}</ul>}
      </div>
    </div>
  )
}
