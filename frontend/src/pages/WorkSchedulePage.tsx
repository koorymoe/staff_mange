import { useCallback, useEffect, useState } from 'react'
import { api, type WorkScheduleRow } from '../api'
import { useSession } from '../session'
import { roleLabel } from '../roleLabels'

// ═══ جدول الدوام — كل موظف متى يبدي ومتى ينتهي، والمسؤول يحدده ═══
// العرض: المدير/المالك/الكوادر/المراقب أو صلاحية work_schedule_manage / staff_management
// (نفس حارس الخادم). التعديل: المدير أو work_schedule_manage بس — غيره الخانات قراءة فقط.

function minutes(t?: string | null) {
  if (!t) return null
  const m = /(\d{1,2}):(\d{2})/.exec(t)
  return m ? Number(m[1]) * 60 + Number(m[2]) : null
}

function todayStatus(r: WorkScheduleRow) {
  const start = minutes(r.shiftStart)
  if (!r.checkIn) return { text: 'ما حضر بعد', cls: 'text-slate-400' }
  const inAt = minutes(new Date(r.checkIn).toTimeString().slice(0, 5))
  const hhmm = new Date(r.checkIn).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })
  if (start != null && inAt != null && inAt - start > 10) return { text: `حضر ${hhmm} — متأخر ${inAt - start} د`, cls: 'text-amber-600 font-bold' }
  return { text: `حضر ${hhmm}`, cls: 'text-emerald-600' }
}

export default function WorkSchedulePage() {
  const { employee, permissions } = useSession()
  const role = employee?.role
  const isAdmin = role === 'ADMIN' || role === 'OWNER'
  const canView = isAdmin || role === 'HR_COORDINATOR' || role === 'MONITOR'
    || permissions.includes('work_schedule_manage') || permissions.includes('staff_management')
  const canEdit = isAdmin || permissions.includes('work_schedule_manage')
  const [rows, setRows] = useState<WorkScheduleRow[]>([])
  const [edits, setEdits] = useState<Record<string, Partial<WorkScheduleRow>>>({})
  const [q, setQ] = useState('')
  const [msg, setMsg] = useState('')

  const load = useCallback(() => { api.getWorkSchedule().then(setRows).catch(() => setMsg('تعذّر جلب الجدول')) }, [])
  useEffect(() => { if (canView) load() }, [canView, load])

  if (!canView) return <div className="rounded-2xl bg-red-50 p-10 text-center font-bold text-red-700">🚫 غير مصرح لك بجدول الدوام</div>

  const save = async (r: WorkScheduleRow) => {
    const e = edits[r.id]
    if (!e) return
    try {
      await api.setEmployeeSchedule(r.id, { shift: e.shift ?? r.shift, shiftStart: e.shiftStart ?? r.shiftStart, shiftEnd: e.shiftEnd ?? r.shiftEnd })
      setEdits((x) => { const n = { ...x }; delete n[r.id]; return n })
      setMsg(`انحفظ جدول ${r.name}`)
      load()
    } catch (err) { setMsg(err instanceof Error ? err.message : 'تعذّر الحفظ') }
  }

  const shown = rows.filter((r) => !q || r.name.includes(q))
  const inp = 'rounded border border-slate-300 px-1.5 py-1 text-xs disabled:bg-transparent disabled:border-transparent'
  return (
    <div className="space-y-4" dir="rtl">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h1 className="text-2xl font-extrabold text-[#0f2040]">🕘 جدول الدوام</h1>
        <input placeholder="بحث بالاسم…" value={q} onChange={(e) => setQ(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm" />
      </div>
      {!canEdit && <p className="text-xs text-slate-500">عرض فقط — تحديد الأوقات يحتاج صلاحية «تحديد جدول الدوام».</p>}
      {msg && <p className="text-sm text-sky-700">{msg}</p>}
      <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-xs text-slate-500">
            <tr><th className="p-2 text-right">الموظف</th><th className="p-2 text-right">الدور</th><th className="p-2">الوجبة</th><th className="p-2">يبدي</th><th className="p-2">ينتهي</th><th className="p-2 text-right">اليوم</th>{canEdit && <th />}</tr>
          </thead>
          <tbody>
            {shown.map((r) => {
              const e = edits[r.id] ?? {}
              const set = (k: keyof WorkScheduleRow, v: string) => setEdits((x) => ({ ...x, [r.id]: { ...x[r.id], [k]: v } }))
              const st = todayStatus(r)
              return (
                <tr key={r.id} className="border-t border-slate-100">
                  <td className="p-2 font-bold">{r.name}{r.isLeader && <span className="mr-1 text-[10px] text-amber-600">قائد</span>}</td>
                  <td className="p-2 text-xs text-slate-500">{roleLabel(r.role)}</td>
                  <td className="p-2 text-center">
                    <select disabled={!canEdit} className={inp} value={e.shift ?? r.shift ?? 'MORNING'} onChange={(ev) => set('shift', ev.target.value)}>
                      <option value="MORNING">صباحي</option><option value="EVENING">مسائي</option>
                    </select>
                  </td>
                  <td className="p-2 text-center"><input type="time" disabled={!canEdit} className={inp} value={e.shiftStart ?? r.shiftStart ?? ''} onChange={(ev) => set('shiftStart', ev.target.value)} /></td>
                  <td className="p-2 text-center"><input type="time" disabled={!canEdit} className={inp} value={e.shiftEnd ?? r.shiftEnd ?? ''} onChange={(ev) => set('shiftEnd', ev.target.value)} /></td>
                  <td className={`p-2 text-xs ${st.cls}`}>{st.text}</td>
                  {canEdit && <td className="p-2">{edits[r.id] && <button onClick={() => save(r)} className="rounded bg-brand-600 px-2 py-1 text-xs font-bold text-white">حفظ</button>}</td>}
                </tr>
              )
            })}
          </tbody>
        </table>
      </div>
    </div>
  )
}
