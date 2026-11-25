import { useEffect, useState } from 'react'
import { apiRequest } from '../api'

// ═══ 🧩 مؤشر هشاشة النظام (ل) 11-25 دازاي ═══
// «في حال غياب موظف — واكتشاف إذا تم الاعتماد على موظف واحد لإنجاز نفس العمل (مثلاً واحد بس يشتغل الأقفال)».

type Svc = { serviceId: string; serviceName: string; jobs: number; workers: number; topName: string | null; topJobs: number; topShare: number; skilled: number; risk: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'OK' }
type Person = { employeeId: string; name: string; soleFor: string[]; mainFor: string[]; onLeave: boolean }
type Rep = { days: number; services: Svc[]; people: Person[]; score: number }
const RISK = {
  CRITICAL: ['🔴 حرج — شخص واحد بس', 'bg-rose-50 text-rose-800'],
  HIGH: ['🟠 عالي — واحد يسوي ٨٠٪+', 'bg-orange-50 text-orange-800'],
  MEDIUM: ['🟡 متوسط — اثنين بس', 'bg-amber-50 text-amber-800'],
  OK: ['🟢 متين', 'bg-emerald-50 text-emerald-800'],
} as const

export default function FragilityPage() {
  const [days, setDays] = useState(90)
  const [d, setD] = useState<Rep | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => { apiRequest<Rep>(`/employee-stats/fragility?days=${days}`).then(setD).catch((e: Error) => setErr(e.message)) }, [days])
  if (err) return <p role="alert" className="text-rose-700">{err}</p>
  if (!d) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  const color = d.score >= 50 ? 'text-rose-700' : d.score >= 25 ? 'text-amber-700' : 'text-emerald-700'
  return (
    <div className="space-y-3" data-testid="fragility">
      <section className="flex flex-wrap items-center justify-between gap-3 rounded-2xl bg-white p-4 shadow-sm">
        <div>
          <h2 className="text-lg font-extrabold text-[#0f2040]">🧩 مؤشر هشاشة النظام</h2>
          <p className="text-xs text-slate-500">من الحجوزات المكتملة بآخر {d.days} يوم: كل خدمة كم موظف يشتغلها، ووين الشغل معتمد على شخص واحد.</p>
        </div>
        <div className="flex items-center gap-3">
          <select value={days} onChange={(e) => setDays(Number(e.target.value))} className="rounded-lg border border-slate-300 p-1.5 text-sm">
            {[30, 90, 180, 365].map((n) => <option key={n} value={n}>آخر {n} يوم</option>)}
          </select>
          <div className="text-center"><p className={`text-3xl font-extrabold ${color}`}>{d.score}</p><p className="text-[10px] text-slate-500">٠ متين · ١٠٠ هش</p></div>
        </div>
      </section>

      {d.people.length > 0 && (
        <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
          <h3 className="font-bold">👤 إذا غاب هذا الموظف…</h3>
          {d.people.map((p) => (
            <div key={p.employeeId} className="rounded-xl bg-slate-50 p-3 text-sm">
              <b>{p.name}</b>{p.onLeave && <span className="ms-2 rounded-full bg-rose-100 px-2 text-xs font-bold text-rose-800">بإجازة هسه!</span>}
              {p.soleFor.length > 0 && <p className="text-rose-700">ماكو غيره يسوي: {p.soleFor.join('، ')}</p>}
              {p.mainFor.length > 0 && <p className="text-orange-700">هو يسوي أغلب: {p.mainFor.join('، ')}</p>}
            </div>
          ))}
          <p className="text-xs text-slate-400">الحل: درّب موظف ثاني على هاي الخدمات (من «التدريب») ووزّع حجوزاتها عليه.</p>
        </section>
      )}

      <section className="overflow-x-auto rounded-2xl bg-white p-4 shadow-sm">
        <table className="w-full min-w-[720px] text-sm">
          <thead className="bg-slate-50 text-xs text-slate-600"><tr>{['الخدمة', 'الخطر', 'الحجوزات', 'كم موظف اشتغلها', 'أكثر واحد', 'حصته', 'يگدرون حسب المهارات'].map((h) => <th key={h} className="p-2 text-center">{h}</th>)}</tr></thead>
          <tbody className="divide-y">
            {d.services.map((s) => (
              <tr key={s.serviceId} className="text-center">
                <td className="p-2 text-start font-bold">{s.serviceName}</td>
                <td><span className={`rounded-full px-2 py-0.5 text-xs font-bold ${RISK[s.risk][1]}`}>{RISK[s.risk][0]}</span></td>
                <td>{s.jobs}</td><td>{s.workers}</td><td>{s.topName ?? '—'}</td><td>{s.topShare}٪</td><td>{s.skilled || '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </section>
    </div>
  )
}
