import { useState } from 'react'
import { apiRequest, type Service } from '../api'

// ⏱📷 لكل خدمة: وقت الالتزام ويا الزبون (ساعات من تسجيل الحجز) وأقل صور «بعد الشغل» قبل «تم».
// (دازاي 11-06: ٤٨ ساعة للكل افتراضياً، والصور صفر = بلا شرط)
export default function ServiceRulesEditor({ services }: { services: Service[] }) {
  const [vals, setVals] = useState<Record<string, { sla: number; photos: number }>>({})
  const [msg, setMsg] = useState<Record<string, string>>({})
  const [q, setQ] = useState('')
  const get = (s: Service) => vals[s.id] ?? { sla: s.slaHours ?? 48, photos: s.minPhotos ?? 0 }
  const save = async (s: Service) => {
    const v = get(s)
    try {
      await apiRequest(`/services/${s.id}/rules`, { method: 'PUT', body: JSON.stringify({ slaHours: v.sla, minPhotos: v.photos }) })
      setMsg((m) => ({ ...m, [s.id]: '✅' }))
    } catch (e) {
      setMsg((m) => ({ ...m, [s.id]: e instanceof Error ? e.message : 'تعذر' }))
    }
  }
  if (services.length === 0) return null
  return (
    <section className="rounded-2xl bg-white p-4 shadow-sm" aria-label="الالتزام والصور">
      <h3 className="mb-1 text-[14px] font-bold text-brand-900">⏱ وقت الالتزام و📷 الصور قبل «تم»</h3>
      <p className="mb-2 text-[12px] text-slate-600">
        وقت الالتزام: كم ساعة من تسجيل الحجز لازم يخلص (يبين للمنسّق، وينبّه قبل ٦ ساعات، ويصعد للمدير إذا فات).
        الصور: أقل عدد صور «بعد الشغل» قبل ما الفني يگدر يسوي «تم» — صفر يعني بلا شرط.
      </p>
      <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="🔍 دوّر على الخدمة" className="mb-2 w-full rounded-lg border border-slate-300 px-3 py-1.5 text-[12.5px]" />
      <div className="space-y-1.5">
        {services
          .filter((s) => !q.trim() || s.name.toLowerCase().includes(q.trim().toLowerCase()))
          .map((s) => {
            const v = get(s)
            return (
              <div key={s.id} className="flex flex-wrap items-center gap-2 text-[12.5px]">
                <span className="min-w-[120px] font-bold text-slate-700">{s.name}</span>
                <label className="flex items-center gap-1">
                  ⏱
                  <input type="number" min={1} max={720} value={v.sla} aria-label={`ساعات الالتزام — ${s.name}`}
                    onChange={(e) => setVals((x) => ({ ...x, [s.id]: { ...v, sla: Number(e.target.value) } }))}
                    className="w-16 rounded border border-slate-300 px-1 py-0.5" />
                  ساعة
                </label>
                <label className="flex items-center gap-1">
                  📷
                  <input type="number" min={0} max={20} value={v.photos} aria-label={`أقل صور — ${s.name}`}
                    onChange={(e) => setVals((x) => ({ ...x, [s.id]: { ...v, photos: Number(e.target.value) } }))}
                    className="w-12 rounded border border-slate-300 px-1 py-0.5" />
                  صور
                </label>
                <button type="button" onClick={() => void save(s)} className="rounded-lg bg-slate-100 px-2 py-0.5 text-xs font-bold">حفظ</button>
                {msg[s.id] && <span className="text-xs">{msg[s.id]}</span>}
              </div>
            )
          })}
      </div>
    </section>
  )
}
