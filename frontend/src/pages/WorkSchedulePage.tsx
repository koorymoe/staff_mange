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
  const [fShift, setFShift] = useState('')
  const [fRole, setFRole] = useState('')
  const [fState, setFState] = useState('')
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

  const shown = rows.filter((r) => (!q || r.name.includes(q)) && (!fShift || (r.shift || 'MORNING') === fShift)
    && (!fRole || r.role === fRole) && (!fState || dayKind(r) === fState))
  const kinds = rows.map(dayKind)
  const morning = rows.filter((r) => (r.shift || 'MORNING') === 'MORNING').length
  const roles = [...new Set(rows.map((r) => r.role))]
  const inp = 'rounded-lg border border-slate-300 bg-white px-2 py-1.5 text-xs disabled:border-transparent disabled:bg-transparent'
  const sel = 'rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm'
  const pill = (k: ReturnType<typeof dayKind>, r: WorkScheduleRow) => {
    const t = r.checkIn ? new Date(r.checkIn).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' }) : ''
    const m = { present: ['✅', 'حضر اليوم', 'bg-emerald-50 text-emerald-700'], late: ['⏰', 'متأخر', 'bg-amber-50 text-amber-700'], absent: ['⊗', 'غائب اليوم', 'bg-rose-50 text-rose-700'], notyet: ['🕓', 'ما حضر بعد', 'bg-slate-100 text-slate-600'] }[k]
    return <span className={`inline-flex items-center gap-1.5 rounded-xl px-3 py-1 text-xs font-bold ${m[2]}`}>{m[0]} <span className="leading-tight">{m[1]}{t && <span className="block text-[10px] font-normal">حضر {t}</span>}</span></span>
  }
  return (
    <div className="space-y-4" dir="rtl">
      <div>
        <h1 className="text-2xl font-black text-[#0f2040]">🕘 جدول الدوام</h1>
        <p className="text-sm text-slate-500">إدارة ورديات الموظفين ومتابعة الحضور اليومي</p>
      </div>
      <StatCards items={[
        { label: 'إجمالي الموظفين', value: rows.length, icon: '👥', tone: 'border-sky-100 bg-sky-50/50 text-sky-800' },
        { label: 'الحاضرين اليوم', value: kinds.filter((k) => k === 'present').length, icon: '✅', tone: 'border-emerald-100 bg-emerald-50/50 text-emerald-700' },
        { label: 'المتأخرين', value: kinds.filter((k) => k === 'late').length, icon: '⏰', tone: 'border-amber-100 bg-amber-50/50 text-amber-700' },
        { label: 'الغائبين', value: kinds.filter((k) => k === 'absent').length, icon: '🚫', tone: 'border-rose-100 bg-rose-50/50 text-rose-700' },
        { label: 'الصباحي / المسائي', value: `${morning} / ${rows.length - morning}`, icon: '🔁', tone: 'border-violet-100 bg-violet-50/50 text-violet-700' },
      ]} />
      <div className="flex flex-wrap items-end gap-2 rounded-2xl border border-slate-200 bg-white p-3 shadow-sm">
        <input placeholder="🔍 بحث باسم الموظف…" value={q} onChange={(e) => setQ(e.target.value)} className={`${sel} min-w-[200px] flex-1`} />
        <label className="text-xs text-slate-500">الوجبة<select value={fShift} onChange={(e) => setFShift(e.target.value)} className={`${sel} mt-1 block`}><option value="">جميع الوجبات</option><option value="MORNING">صباحي</option><option value="EVENING">مسائي</option></select></label>
        <label className="text-xs text-slate-500">الدور الوظيفي<select value={fRole} onChange={(e) => setFRole(e.target.value)} className={`${sel} mt-1 block`}><option value="">جميع الأدوار</option>{roles.map((r) => <option key={r} value={r}>{roleLabel(r)}</option>)}</select></label>
        <label className="text-xs text-slate-500">حالة الحضور<select value={fState} onChange={(e) => setFState(e.target.value)} className={`${sel} mt-1 block`}><option value="">جميع الحالات</option><option value="present">حضر</option><option value="late">متأخر</option><option value="absent">غائب</option><option value="notyet">ما حضر بعد</option></select></label>
        <button onClick={() => { setQ(''); setFShift(''); setFRole(''); setFState('') }} className="rounded-xl border border-slate-300 px-4 py-2 text-sm text-slate-600">↻ إعادة تعيين</button>
      </div>
      {!canEdit && <p className="text-xs text-slate-500">عرض فقط — تحديد الأوقات يحتاج صلاحية «تحديد جدول الدوام».</p>}
      {msg && <p className="text-sm text-sky-700">{msg}</p>}
      <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white shadow-sm">
        <table className="w-full min-w-[860px] text-sm">
          <thead className="bg-slate-50 text-xs text-slate-500">
            <tr><th className="p-3 text-right">الموظف</th><th className="p-3 text-right">الدور الوظيفي</th><th className="p-3">الوجبة</th><th className="p-3">يبدأ</th><th className="p-3">ينتهي</th><th className="p-3">حالة اليوم</th>{canEdit && <th className="p-3">إجراءات</th>}</tr>
          </thead>
          <tbody>
            {shown.map((r) => {
              const e = edits[r.id] ?? {}
              const set = (k: keyof WorkScheduleRow, v: string) => setEdits((x) => ({ ...x, [r.id]: { ...x[r.id], [k]: v } }))
              return (
                <tr key={r.id} className="border-t border-slate-100 hover:bg-slate-50/50">
                  <td className="p-2.5"><span className="flex items-center gap-2 font-bold text-[#0f2040]"><span className="grid h-8 w-8 place-items-center rounded-full bg-sky-50 text-[11px] text-[#2c5aad]">{initials(r.name)}</span>{r.name}{r.isLeader && <span className="text-[10px] text-amber-600">قائد</span>}</span></td>
                  <td className="p-2.5 text-xs text-slate-500">{roleLabel(r.role)}</td>
                  <td className="p-2.5 text-center">
                    <select disabled={!canEdit} className={`${inp} ${(e.shift ?? r.shift) === 'EVENING' ? 'bg-violet-50' : ''}`} value={e.shift ?? r.shift ?? 'MORNING'} onChange={(ev) => set('shift', ev.target.value)}>
                      <option value="MORNING">صباحي</option><option value="EVENING">مسائي</option>
                    </select>
                  </td>
                  <td className="p-2.5 text-center"><input type="time" disabled={!canEdit} className={inp} value={e.shiftStart ?? r.shiftStart ?? ''} onChange={(ev) => set('shiftStart', ev.target.value)} /></td>
                  <td className="p-2.5 text-center"><input type="time" disabled={!canEdit} className={inp} value={e.shiftEnd ?? r.shiftEnd ?? ''} onChange={(ev) => set('shiftEnd', ev.target.value)} /></td>
                  <td className="p-2.5 text-center">{pill(dayKind(r), r)}</td>
                  {canEdit && <td className="p-2.5 text-center">{edits[r.id] ? <button onClick={() => save(r)} className="rounded-lg bg-[#0f2040] px-3 py-1 text-xs font-bold text-white">حفظ</button> : <span className="text-slate-300">⋮</span>}</td>}
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

function nowMin() { const d = new Date(); return d.getHours() * 60 + d.getMinutes() }
// حالة اليوم بكلمة وحدة (للبطاقات والفلتر)
function dayKind(r: WorkScheduleRow): 'present' | 'late' | 'absent' | 'notyet' {
  const start = minutes(r.shiftStart)
  if (r.checkIn) {
    const inAt = minutes(new Date(r.checkIn).toTimeString().slice(0, 5))
    return start != null && inAt != null && inAt - start > 10 && inAt - start <= 180 ? 'late' : 'present'
  }
  return start != null && nowMin() > start + 60 ? 'absent' : 'notyet'
}
const initials = (n: string) => n.split(/\s+/).slice(0, 2).map((w) => w[0]).join(' ')

// بطاقات الأرقام — تصميم (ع) 10-13
function StatCards({ items }: { items: { label: string; value: string | number; icon: string; tone: string }[] }) {
  return (
    <div className="grid grid-cols-2 gap-3 lg:grid-cols-5">
      {items.map((x) => (
        <div key={x.label} className={`flex items-center justify-between rounded-2xl border p-4 shadow-sm ${x.tone}`}>
          <div><p className="text-sm font-extrabold">{x.label}</p><p className="mt-1 text-2xl font-black text-[#0f2040] tabular-nums">{x.value}</p><p className="text-[11px] text-slate-500">موظف</p></div>
          <span className="grid h-12 w-12 place-items-center rounded-2xl bg-white/80 text-2xl shadow-sm">{x.icon}</span>
        </div>
      ))}
    </div>
  )
}

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
  const [fRole, setFRole] = useState('')
  const [fState, setFState] = useState('')
  const [open, setOpen] = useState<AttendanceBoardRow | null>(null)
  useEffect(() => { api.getAttendanceBoard(date).then((r) => setRows(r.rows)).catch(() => setRows([])) }, [date])
  const all = rows ?? []
  const kind = (r: AttendanceBoardRow): 'leave' | 'late' | 'present' | 'absent' => {
    if (r.leaveKind) return 'leave'
    if (!r.firstIn) return 'absent'
    const s = minutes(r.shiftStart), i = minutes(new Date(r.firstIn).toTimeString().slice(0, 5))
    return s != null && i != null && i - s > 10 && i - s <= 180 ? 'late' : 'present'
  }
  const list = all.filter((r) => (!q || r.name.includes(q)) && (!fRole || r.role === fRole) && (!fState || kind(r) === fState))
  const ks = all.map(kind)
  const morning = all.filter((r) => (r.shift || 'MORNING') === 'MORNING').length
  const roles = [...new Set(all.map((r) => r.role))]
  const sel = 'rounded-xl border border-slate-300 bg-white px-3 py-2 text-sm'
  const PILL = { present: ['✅ حضر', 'bg-emerald-50 text-emerald-700'], late: ['⏰ متأخر', 'bg-amber-50 text-amber-700'], absent: ['⊗ ما حضر', 'bg-rose-50 text-rose-700'], leave: ['🗓️ إجازة', 'bg-sky-50 text-sky-700'] }
  return (
    <div className="space-y-4" dir="rtl">
      <div>
        <h1 className="text-2xl font-black text-[#0f2040]">📋 دوام الموظفين (يوم وشهر)</h1>
        <p className="text-sm text-slate-500">متابعة حضور وانصراف الموظفين وساعات العمل اليومية</p>
      </div>
      <StatCards items={[
        { label: 'إجمالي الموظفين', value: all.length, icon: '👥', tone: 'border-sky-100 bg-sky-50/50 text-sky-800' },
        { label: 'الحاضرين', value: ks.filter((k) => k === 'present').length, icon: '✅', tone: 'border-emerald-100 bg-emerald-50/50 text-emerald-700' },
        { label: 'المتأخرين', value: ks.filter((k) => k === 'late').length, icon: '⏰', tone: 'border-amber-100 bg-amber-50/50 text-amber-700' },
        { label: 'الغياب', value: ks.filter((k) => k === 'absent').length, icon: '🚫', tone: 'border-rose-100 bg-rose-50/50 text-rose-700' },
        { label: 'الصباحي / المسائي', value: `${morning} / ${all.length - morning}`, icon: '🔁', tone: 'border-violet-100 bg-violet-50/50 text-violet-700' },
      ]} />
      <div className="flex flex-wrap items-end gap-2 rounded-2xl border border-slate-200 bg-white p-3 shadow-sm">
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="🔍 بحث باسم الموظف…" className={`${sel} min-w-[200px] flex-1`} />
        <label className="text-xs text-slate-500">الدور الوظيفي<select value={fRole} onChange={(e) => setFRole(e.target.value)} className={`${sel} mt-1 block`}><option value="">جميع الأدوار</option>{roles.map((r) => <option key={r} value={r}>{roleLabel(r)}</option>)}</select></label>
        <label className="text-xs text-slate-500">حالة الحضور<select value={fState} onChange={(e) => setFState(e.target.value)} className={`${sel} mt-1 block`}><option value="">جميع الحالات</option><option value="present">حضر</option><option value="late">متأخر</option><option value="absent">ما حضر</option><option value="leave">إجازة</option></select></label>
        <label className="text-xs text-slate-500">التاريخ<input type="date" value={date} onChange={(e) => setDate(e.target.value)} className={`${sel} mt-1 block`} /></label>
      </div>
      <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white shadow-sm">
        <table className="w-full min-w-[860px] text-sm">
          <thead className="bg-slate-50"><tr className="text-[12px] text-slate-500">
            {['#', 'الموظف', 'الدور', 'جدوله', 'بدا دوامه', 'انتهى', 'مجموع الساعات', 'الحالة', ''].map((h, i) => <th key={i} className="px-3 py-3 text-right font-bold">{h}</th>)}
          </tr></thead>
          <tbody>
            {rows === null ? <tr><td colSpan={9} className="p-6 text-center text-slate-400">جاري التحميل…</td></tr> : list.map((r, i) => {
              const k = kind(r)
              return (
                <tr key={r.id} className="border-t border-slate-100 hover:bg-slate-50/50">
                  <td className="px-3 py-2.5 text-xs text-slate-400">{i + 1}</td>
                  <td className="px-3"><span className="flex items-center gap-2 font-bold text-[#0f2040]"><span className="grid h-8 w-8 place-items-center rounded-full bg-sky-50 text-[11px] text-[#2c5aad]">{initials(r.name)}</span>{r.name}</span></td>
                  <td className="px-3 text-xs text-slate-500">{roleLabel(r.role)}</td>
                  <td className="px-3 text-xs tabular-nums">{r.shiftStart && r.shiftEnd ? `${r.shiftStart} – ${r.shiftEnd}` : r.shift === 'EVENING' ? 'مسائي' : 'صباحي'}</td>
                  <td className="px-3 font-bold tabular-nums">{hm(r.firstIn)}</td>
                  <td className="px-3 tabular-nums">{r.open ? <span className="text-emerald-600">بالدوام هسه</span> : hm(r.lastOut)}</td>
                  <td className="px-3 tabular-nums">{r.minutes ? dur(r.minutes) : '—'}</td>
                  <td className="px-3"><span className={`rounded-xl px-3 py-1 text-xs font-bold ${PILL[k][1]}`}>{PILL[k][0]}</span></td>
                  <td className="px-3"><button onClick={() => setOpen(r)} className="rounded-lg border border-slate-300 px-2.5 py-1 text-xs font-bold text-[#2c5aad]">📅 جدول الشهر</button></td>
                </tr>
              )
            })}
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
