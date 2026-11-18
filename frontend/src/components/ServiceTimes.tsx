import { useCallback, useEffect, useState } from 'react'
import { apiRequest } from '../api'
import { useSession, isAdminOrOwner } from '../session'

// ═══ ⏱️ الوقت القياسي لكل خدمة (دازاي 11-18 — تقييم الليدر) ═══
// «ماتركس يعرف حجز الكامرات ياخذ ساعة ونص… عشر دقايق فوگ عادي، ساعتين خطأ».
// ماتركس يتعلّم الوقت من الشغل الفعلي (الوسيط لآخر ٩٠ يوم)، والمالك/المدير يعدّله.
interface Row { serviceId: string; name: string; learned: number | null; samples: number; override: number | null }

export default function ServiceTimes() {
  const { employee } = useSession()
  const canEdit = isAdminOrOwner(employee)
  const [rows, setRows] = useState<Row[] | null>(null)
  const [grace, setGrace] = useState<[number, number]>([15, 60])
  const [edit, setEdit] = useState<Record<string, string>>({})
  const [busy, setBusy] = useState('')
  const [err, setErr] = useState('')
  const load = useCallback(() => {
    apiRequest<{ items: Row[]; graceOk: number; gracePartial: number }>('/staff-score/service-times')
      .then((d) => { setRows(d.items); setGrace([d.graceOk, d.gracePartial]) })
      .catch((e) => setErr(e instanceof Error ? e.message : 'تعذر الجلب'))
  }, [])
  useEffect(() => { load() }, [load])
  const save = async (id: string, minutes: number | null) => {
    setBusy(id); setErr('')
    try {
      await apiRequest(`/staff-score/service-times/${id}`, { method: 'PUT', body: JSON.stringify({ minutes }) })
      setEdit((e) => { const n = { ...e }; delete n[id]; return n }); load()
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy('') }
  }
  if (err && !rows) return <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{err}</p>
  if (!rows) return <p className="py-6 text-center text-slate-400">جاري الجلب…</p>
  return (
    <div dir="rtl" className="space-y-3">
      <p className="rounded-xl bg-sky-50 px-4 py-2 text-[12.5px] text-sky-900">
        وقت الليدر بكل حجز ينقاس على هذا الوقت: لحد +{grace[0]} دقيقة فوگه كامل، لحد +{grace[1]} نص، أكثر صفر.
        «ماتركس» = الوسيط من الشغل الفعلي لآخر ٩٠ يوم (يحتاج ٥ حجوزات). التعديل اليدوي يغلبه.
      </p>
      {err && <p className="text-xs text-red-600">{err}</p>}
      <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-xs text-slate-500">
            <tr><th className="p-2 text-right">الخدمة</th><th className="p-2">ماتركس</th><th className="p-2">المعتمد</th>{canEdit && <th className="p-2">تعديل (دقيقة)</th>}</tr>
          </thead>
          <tbody>
            {rows.map((r) => {
              const used = r.override ?? r.learned
              return (
                <tr key={r.serviceId} className="border-t border-slate-100">
                  <td className="p-2 font-bold">{r.name}</td>
                  <td className="p-2 text-center text-slate-600">{r.learned != null ? `${r.learned} د (${r.samples} حجز)` : <span className="text-slate-400">{r.samples ? `${r.samples} حجز — قليل` : 'ماكو بيانات'}</span>}</td>
                  <td className="p-2 text-center">{used != null ? <b>{used} د{r.override != null && ' ✍️'}</b> : <span className="text-slate-400">ما ينقاس</span>}</td>
                  {canEdit && <td className="p-2">
                    <div className="flex items-center justify-center gap-1">
                      <input type="number" min={5} max={1440} value={edit[r.serviceId] ?? ''} placeholder={used != null ? String(used) : ''}
                        onChange={(e) => setEdit((x) => ({ ...x, [r.serviceId]: e.target.value }))}
                        className="w-20 rounded-lg border border-slate-300 px-2 py-1 text-center" />
                      <button type="button" disabled={busy === r.serviceId || !edit[r.serviceId]}
                        onClick={() => void save(r.serviceId, Number(edit[r.serviceId]))}
                        className="rounded-lg bg-brand-500 px-2 py-1 text-xs font-bold text-white disabled:opacity-40">حفظ</button>
                      {r.override != null && <button type="button" disabled={busy === r.serviceId} onClick={() => void save(r.serviceId, null)}
                        className="rounded-lg px-2 py-1 text-xs font-bold text-slate-500 hover:bg-slate-100">رجّعه لماتركس</button>}
                    </div>
                  </td>}
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
