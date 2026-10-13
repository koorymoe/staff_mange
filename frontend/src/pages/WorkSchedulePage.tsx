import { lazy, Suspense, useCallback, useEffect, useState } from 'react'
import { api, type AttendanceBoardRow, type WorkScheduleRow } from '../api'
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

function ScheduleTab() {
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

// ═══ (ع) 10-13: جدول الدوام بثلاث واجهات — الجدول، دوام الموظفين (يوم/شهر)، طلبات الإجازات ═══
const LeavesPage = lazy(() => import('./LeavesPage'))

export default function WorkSchedulePage() {
  const [tab, setTab] = useState<'schedule' | 'board' | 'leaves'>('schedule')
  return (
    <div dir="rtl" className="w-full space-y-4">
      <div className="flex flex-wrap gap-2 rounded-2xl bg-white p-2 shadow-sm">
        {([['schedule', '🕘 جدول الدوام'], ['board', '📋 دوام الموظفين (يوم وشهر)'], ['leaves', '🗓️ طلبات الإجازات']] as const).map(([k, l]) => (
          <button key={k} onClick={() => setTab(k)} className={`rounded-xl px-4 py-2 text-sm font-bold ${tab === k ? 'bg-[#0f2040] text-white' : 'text-slate-600 hover:bg-slate-100'}`}>{l}</button>
        ))}
      </div>
      {tab === 'schedule' && <ScheduleTab />}
      {tab === 'board' && <AttendanceBoard />}
      {tab === 'leaves' && <Suspense fallback={<p className="text-slate-400">جاري التحميل…</p>}><LeavesPage /></Suspense>}
    </div>
  )
}

const hm = (s: string | null | undefined) => s ? new Date(s).toLocaleTimeString('ar-IQ', { timeZone: 'Asia/Baghdad', hour: '2-digit', minute: '2-digit' }) : '—'
const dur = (m: number) => `${Math.floor(m / 60)}س ${m % 60}د`

function AttendanceBoard() {
  const [date, setDate] = useState(() => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' }))
  const [rows, setRows] = useState<AttendanceBoardRow[] | null>(null)
  const [q, setQ] = useState('')
  const [open, setOpen] = useState<AttendanceBoardRow | null>(null)
  useEffect(() => { api.getAttendanceBoard(date).then((r) => setRows(r.rows)).catch(() => setRows([])) }, [date])
  const list = (rows ?? []).filter((r) => !q || r.name.includes(q))
  const present = (rows ?? []).filter((r) => r.firstIn).length
  const onLeave = (rows ?? []).filter((r) => r.leaveKind).length
  return (
    <div className="space-y-3">
      <div className="flex flex-wrap items-center gap-2 rounded-2xl bg-white p-3 shadow-sm">
        <input type="date" value={date} onChange={(e) => setDate(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm" />
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="بحث بالاسم…" className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm" />
        <span className="text-sm text-slate-600">حضروا <b className="text-emerald-700">{present}</b> · بإجازة <b className="text-amber-700">{onLeave}</b> · ما حضروا <b className="text-red-600">{(rows?.length ?? 0) - present - onLeave}</b></span>
      </div>
      <div className="overflow-x-auto rounded-2xl bg-white shadow-sm">
        <table className="w-full min-w-[760px] text-sm">
          <thead><tr className="border-b text-[12px] text-slate-500">
            {['الموظف', 'الدور', 'جدوله', 'بدا دوامه', 'انتهى', 'مجموع الساعات', 'الحالة', ''].map((h) => <th key={h} className="px-3 py-2 text-right font-bold">{h}</th>)}
          </tr></thead>
          <tbody>
            {rows === null ? <tr><td colSpan={8} className="p-6 text-center text-slate-400">جاري التحميل…</td></tr> : list.map((r) => (
              <tr key={r.id} className="border-b border-slate-50">
                <td className="px-3 py-2 font-bold text-[#0f2040]">{r.name}</td>
                <td className="px-3 text-xs text-slate-500">{roleLabel(r.role)}</td>
                <td className="px-3 text-xs">{r.shiftStart && r.shiftEnd ? `${r.shiftStart} – ${r.shiftEnd}` : r.shift === 'EVENING' ? 'مسائي' : 'صباحي'}</td>
                <td className="px-3 font-bold tabular-nums">{hm(r.firstIn)}</td>
                <td className="px-3 tabular-nums">{r.open ? <span className="text-emerald-600">بالدوام هسه</span> : hm(r.lastOut)}</td>
                <td className="px-3 tabular-nums">{r.minutes ? dur(r.minutes) : '—'}</td>
                <td className="px-3 text-xs font-bold">{r.leaveKind ? <span className="text-amber-700">🗓️ إجازة</span> : r.firstIn ? <span className="text-emerald-700">حضر</span> : <span className="text-red-600">ما حضر</span>}</td>
                <td className="px-3"><button onClick={() => setOpen(r)} className="rounded-lg border border-slate-300 px-2.5 py-1 text-xs font-bold text-[#2c5aad]">📅 جدول الشهر</button></td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {open && <MonthModal emp={open} onClose={() => setOpen(null)} />}
    </div>
  )
}

function MonthModal({ emp, onClose }: { emp: AttendanceBoardRow; onClose: () => void }) {
  const [month, setMonth] = useState(() => new Date().toISOString().slice(0, 7))
  const [rep, setRep] = useState<Awaited<ReturnType<typeof api.getMonthlyAttendance>> | null>(null)
  useEffect(() => { api.getMonthlyAttendance(emp.id, month).then(setRep).catch(() => setRep(null)) }, [emp.id, month])
  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/40 p-4" onClick={onClose}>
      <div dir="rtl" className="w-full max-w-2xl rounded-2xl bg-white p-5 shadow-2xl" onClick={(e) => e.stopPropagation()}>
        <div className="mb-3 flex flex-wrap items-center justify-between gap-2">
          <h3 className="text-lg font-extrabold text-[#0f2040]">📅 دوام {emp.name}</h3>
          <div className="flex items-center gap-2">
            <input type="month" value={month} onChange={(e) => setMonth(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1 text-sm" />
            <button onClick={() => void api.exportEmployeeAttendance(emp.id, month)} className="rounded-lg bg-[#0f2040] px-3 py-1.5 text-xs font-bold text-white">⬇️ Excel (بيه أيام الإجازة)</button>
            <button onClick={onClose} className="text-2xl text-slate-400">×</button>
          </div>
        </div>
        {rep && <p className="mb-2 text-sm text-slate-600">أيام الحضور <b>{rep.daysPresent}</b> · مجموع الساعات <b>{dur(rep.totalMinutes)}</b></p>}
        <div className="max-h-[65vh] overflow-y-auto">
          <table className="w-full text-sm">
            <thead><tr className="text-[12px] text-slate-500"><th className="text-right">اليوم</th><th className="text-right">بدا</th><th className="text-right">انتهى</th><th className="text-right">الساعات</th></tr></thead>
            <tbody>
              {(rep?.days ?? []).map((d) => {
                const s = d.sessions ?? []
                return (
                  <tr key={String(d.date)} className="border-t border-slate-100">
                    <td className="py-1.5">{new Date(d.date).toLocaleDateString('ar-IQ', { weekday: 'short', day: 'numeric', month: 'short' })}</td>
                    <td className="tabular-nums">{hm(s[0]?.checkIn)}</td>
                    <td className="tabular-nums">{hm(s[s.length - 1]?.checkOut)}</td>
                    <td className="tabular-nums">{dur(d.totalMinutes)}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
