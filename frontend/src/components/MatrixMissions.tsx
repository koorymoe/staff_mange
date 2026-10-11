import { useCallback, useEffect, useState } from 'react'
import { api, type Employee, type MatrixMission } from '../api'

// ═══ 🎯 مهام ماتركس «راقب فلان» — قرار (ع) 10-11 ═══
// تختار موظف وكم يوم: ماتركس يكتب يومية كل ليلة (٨م) وبالنهاية خلاصة.
export default function MatrixMissions() {
  const [rows, setRows] = useState<MatrixMission[]>([])
  const [emps, setEmps] = useState<Employee[]>([])
  const [emp, setEmp] = useState('')
  const [note, setNote] = useState('')
  const [days, setDays] = useState(7)
  const [open, setOpen] = useState<{ mission: MatrixMission; logs: { day: string; text: string }[] } | null>(null)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)

  const load = useCallback(() => { api.getMatrixMissions().then(setRows).catch(() => setRows([])) }, [])
  useEffect(load, [load])
  useEffect(() => {
    api.getEmployees().then((r) => setEmps(r.filter((e) => e.status === 'ACTIVE' && e.role !== 'ADMIN' && e.role !== 'OWNER')
      .sort((a, b) => a.name.localeCompare(b.name, 'ar')))).catch(() => setEmps([]))
  }, [])

  const create = async () => {
    if (!emp) { setErr('اختار الموظف'); return }
    setBusy(true); setErr('')
    try {
      await api.createMatrixMission({ employeeId: emp, note: note.trim() || undefined, days })
      setEmp(''); setNote(''); load()
    } catch (e) { setErr(e instanceof Error ? e.message : 'تعذر') } finally { setBusy(false) }
  }

  const input = 'rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-violet-500'
  return (
    <div dir="rtl" className="space-y-4">
      <div className="rounded-2xl border border-violet-200 bg-violet-50/50 p-4">
        <p className="font-extrabold text-violet-950">🎯 كلّف ماتركس يراقب موظف</p>
        <p className="mb-3 text-xs text-violet-800">كل ليلة الساعة ٨ يكتب سطر بالي صار (حضوره، شنو خلّص، المتأخر، الفواتير)، وبآخر يوم يكتبلك خلاصة ونصيحة.</p>
        <div className="grid gap-2 sm:grid-cols-[1fr_1fr_120px_auto]">
          <select value={emp} onChange={(e) => setEmp(e.target.value)} className={input}>
            <option value="">— اختار الموظف —</option>
            {emps.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
          </select>
          <input value={note} onChange={(e) => setNote(e.target.value)} placeholder="ليش؟ (اختياري) مثلاً: تأخيره زايد" className={input} />
          <select value={days} onChange={(e) => setDays(Number(e.target.value))} className={input}>
            {[3, 7, 14, 30].map((d) => <option key={d} value={d}>{d} يوم</option>)}
          </select>
          <button type="button" disabled={busy} onClick={() => void create()} className="rounded-lg bg-violet-600 px-4 py-2 text-sm font-bold text-white disabled:opacity-50">ابدأ المراقبة</button>
        </div>
        {err && <p className="mt-2 text-xs font-bold text-red-700">{err}</p>}
      </div>

      <div className="space-y-2">
        {rows.length === 0 && <p className="p-6 text-center text-sm text-slate-400">ماكو مهام مراقبة بعد.</p>}
        {rows.map((m) => (
          <div key={m.id} className="flex flex-wrap items-center justify-between gap-2 rounded-xl border border-slate-200 bg-white px-4 py-3">
            <div>
              <p className="font-bold text-[#0f2040]">{m.employeeName}
                <span className={`mr-2 rounded-full px-2 py-0.5 text-[11px] font-bold ${m.active ? 'bg-emerald-50 text-emerald-700' : m.summary ? 'bg-violet-50 text-violet-700' : 'bg-slate-100 text-slate-500'}`}>
                  {m.active ? '● يراقب' : m.summary ? '✓ الخلاصة جاهزة' : 'موقوفة'}
                </span>
              </p>
              <p className="text-xs text-slate-500">{m.startsOn} ← {m.endsOn}{m.note && <> · {m.note}</>}</p>
            </div>
            <div className="flex gap-2">
              <button type="button" onClick={() => api.getMatrixMission(m.id).then(setOpen).catch(() => {})} className="rounded-lg border border-violet-300 px-3 py-1.5 text-xs font-bold text-violet-800">اليومية والخلاصة</button>
              {m.active && <button type="button" onClick={() => { if (window.confirm('توقف المراقبة؟')) api.stopMatrixMission(m.id).then(load) }} className="rounded-lg px-3 py-1.5 text-xs font-bold text-slate-500 hover:text-red-600">إيقاف</button>}
            </div>
          </div>
        ))}
      </div>

      {open && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setOpen(null)}>
          <div className="max-h-[80vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-5 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <div className="mb-3 flex items-center justify-between">
              <p className="text-base font-extrabold text-violet-900">🎯 {open.mission.employeeName}</p>
              <button type="button" onClick={() => setOpen(null)} className="text-xs font-bold text-slate-500">✕ سكّر</button>
            </div>
            {open.mission.summary && (
              <div className="mb-3 whitespace-pre-wrap rounded-xl border border-violet-200 bg-violet-50/60 p-3 text-sm leading-relaxed text-violet-950">
                <b>🤖 خلاصة ماتركس:</b>{'\n'}{open.mission.summary}
              </div>
            )}
            <p className="mb-1 text-xs font-bold text-slate-600">اليومية</p>
            {open.logs.length === 0 && <p className="text-xs text-slate-400">أول سطر ينكتب الليلة الساعة ٨.</p>}
            <div className="space-y-1">
              {open.logs.map((l) => <p key={l.day} className="rounded-lg bg-slate-50 px-3 py-1.5 text-xs text-slate-700"><b>{l.day}</b> — {l.text}</p>)}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
