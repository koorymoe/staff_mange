import { useEffect, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSession } from '../session'
import { api, type MonthlyAttendanceReport, type EmployeeDailyAttendanceSummary, type OpenSessionResponse, type DailyAttendance } from '../api'
import { countAbsentDays, countLateDays, isLateDay, movementsOf, type Movement } from '../attendanceStats'

/* ───── helpers ───── */

function fmtTime(iso: string | null): string {
  if (!iso) return '—'
  return new Date(iso).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })
}

function fmtHours(minutes: number): string {
  const h = Math.floor(minutes / 60)
  const m = minutes % 60
  if (h === 0) return `${m} دقيقة`
  if (m === 0) return `${h} ساعة`
  return `${h} ساعة و ${m} دقيقة`
}

function elapsedSince(iso: string): string {
  const diffMs = Date.now() - new Date(iso).getTime()
  if (diffMs < 0) return '٠ دقائق'
  const diffMin = Math.floor(diffMs / 60000)
  return fmtHours(diffMin)
}

function currentMonthKey(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

function shiftMonth(month: string, delta: number): string {
  const [y, m] = month.split('-').map(Number)
  const d = new Date(y, m - 1 + delta, 1)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

function monthLabel(month: string): string {
  const [y, m] = month.split('-').map(Number)
  return new Date(y, m - 1, 1).toLocaleDateString('ar-IQ', { year: 'numeric', month: 'long' })
}

function dayLabel(iso: string): string {
  return new Date(iso).toLocaleDateString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'short' })
}

/* ───── Component ───── */

export default function AttendancePage() {
  const { employee, permissions } = useSession()
  const navigate = useNavigate()

  const [openSession, setOpenSession] = useState<OpenSessionResponse | null>(null)
  const [loading, setLoading] = useState(true)
  const [justCheckedIn, setJustCheckedIn] = useState(false)
  const [justCheckedOut, setJustCheckedOut] = useState(false)
  const [elapsed, setElapsed] = useState('')
  const [month, setMonth] = useState(currentMonthKey())
  const [report, setReport] = useState<MonthlyAttendanceReport | null>(null)
  const [todaySummary, setTodaySummary] = useState<EmployeeDailyAttendanceSummary[]>([])
  const [error, setError] = useState('')
  const [viewedEmployeeId, setViewedEmployeeId] = useState<string>('')
  // «عرض السجل» و«عرض التفاصيل» يفتحون نفس اللوحة — الحركات الكاملة
  const [showAllMovements, setShowAllMovements] = useState(false)

  const isAdmin = employee?.role === 'ADMIN' || employee?.role === 'OWNER' || employee?.role === 'MONITOR' || permissions.includes('monitoring')
  // تصدير جدول الدوام مو لكل موظف — للمراقب ومدير النظام والمالك بس،
  // أو لمن ينطيه المدير صلاحية المراقبة صراحةً.
  const canExport = isAdmin

  const loadOpenSession = useCallback(() => {
    if (!employee) return
    api.getMyOpenSession().then(setOpenSession).catch(() => setOpenSession(null)).finally(() => setLoading(false))
  }, [employee])

  const targetEmployeeId = viewedEmployeeId || employee?.id || ''

  const loadReport = useCallback(() => {
    if (!targetEmployeeId) return
    api.getMonthlyAttendance(targetEmployeeId, month).then(setReport).catch(() => setReport(null))
  }, [targetEmployeeId, month])

  useEffect(() => { loadOpenSession() }, [loadOpenSession])
  useEffect(() => { loadReport() }, [loadReport])
  useEffect(() => {
    if (isAdmin) api.getTodaySummary().then(setTodaySummary).catch(() => setTodaySummary([]))
  }, [isAdmin])

  const today = openSession?.open ?? null

  useEffect(() => {
    if (!today?.checkIn) return
    // Initializing the elapsed-time display synchronously (before the 60s tick) so the
    // timer doesn't show a stale/blank value for a minute is intentional here.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setElapsed(elapsedSince(today.checkIn))
    const interval = setInterval(() => setElapsed(elapsedSince(today.checkIn)), 60000)
    return () => clearInterval(interval)
  }, [today])

  const handleCheckIn = useCallback(async () => {
    try {
      setError('')
      await api.checkIn()
      loadOpenSession()
      loadReport()
      setJustCheckedIn(true)
      setTimeout(() => navigate('/'), 2000)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر تسجيل الحضور')
    }
  }, [navigate, loadReport, loadOpenSession])

  const handleCheckOut = useCallback(async () => {
    try {
      setError('')
      await api.checkOut()
      loadOpenSession()
      loadReport()
      setJustCheckedOut(true)
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذر تسجيل الانصراف')
    }
  }, [loadReport, loadOpenSession])

  if (!employee || loading) return null

  /* ── Just checked in: success screen ── */
  if (justCheckedIn && today) {
    return (
      <div dir="rtl" className="flex min-h-[80vh] items-center justify-center p-4">
        <div className="w-full max-w-md rounded-2xl bg-white p-10 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="mx-auto mb-6 flex h-24 w-24 items-center justify-center rounded-full bg-green-100">
            <svg className="h-14 w-14 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2.5}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
            </svg>
          </div>
          <h2 className="mb-2 text-2xl font-bold text-green-700">شكراً! تم تسجيل حضورك</h2>
          <p className="text-lg text-gray-600">وقت الحضور: {fmtTime(today.checkIn)}</p>
          <p className="mt-4 text-sm text-gray-400">جارٍ التحويل إلى لوحة التحكم...</p>
        </div>
      </div>
    )
  }

  // ── أرقام الشهر ──
  const shiftStart = employee.shiftStart ?? null
  const lateDays = countLateDays(report, shiftStart)
  const absentDays = countAbsentDays(report, month)
  const todayMinutes = openSession?.totalMinutes ?? 0

  // حركات اليوم: من جلسات اليوم نفسه بالتقرير الشهري
  const todayKey = new Date().toISOString().slice(0, 10)
  const todayRecord: DailyAttendance | null =
    report?.days.find((d) => d.date.slice(0, 10) === todayKey) ?? null
  const movements = movementsOf(todayRecord)

  return (
    <div dir="rtl" className="mx-auto max-w-6xl space-y-5 p-4">
      {/* ═══ العنوان ═══ */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <span className="flex h-11 w-11 items-center justify-center rounded-xl bg-sky-50 text-xl">🕐</span>
          <div>
            <h1 className="text-2xl font-black text-[#0f2040]">الحضور</h1>
            <p className="text-xs text-slate-500">متابعة حضورك وانصرافك اليومي</p>
          </div>
        </div>
        <span className="rounded-lg bg-white px-3 py-1.5 text-xs font-bold text-slate-600 shadow-sm">
          📅 {new Date().toLocaleDateString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' })}
        </span>
      </div>

      {error && (
        <div className="rounded-xl border-2 border-red-300 bg-red-50 px-4 py-3 text-sm font-bold text-red-700">{error}</div>
      )}

      {/* ═══ البطاقات الأربع ═══ */}
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <StatCard
          icon="🛡️" label="حالة اليوم"
          value={today ? 'متواجد' : 'غير مسجّل'}
          tone={today ? 'emerald' : 'slate'}
        />
        <StatCard icon="🕐" label="وقت الحضور" value={fmtTime(todayRecord?.firstCheckIn ?? today?.checkIn ?? null)} tone="sky" />
        <StatCard icon="⏱️" label="إجمالي الساعات هذا الشهر" value={report ? fmtHours(report.totalMinutes) : '—'} tone="violet" />
        <StatCard icon="📅" label="أيام الحضور هذا الشهر" value={report ? `${report.daysPresent} أيام` : '—'} tone="amber" />
      </div>

      {/* ═══ حالة الحضور + ملخص اليوم ═══ */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        {/* اللوحة الكبيرة */}
        <div className="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm lg:col-span-2">
          <h2 className="mb-5 flex items-center gap-2 text-base font-extrabold text-[#0f2040]">
            🕐 حالة الحضور
          </h2>

          <div className="flex flex-col items-center gap-5 sm:flex-row-reverse sm:items-center sm:justify-between">
            {/* الدائرة */}
            <div className="relative flex h-40 w-40 shrink-0 items-center justify-center">
              <span className={`absolute inset-0 rounded-full ${today ? 'bg-emerald-100' : 'bg-slate-100'}`} />
              <span className={`absolute inset-4 rounded-full ${today ? 'bg-emerald-200/70' : 'bg-slate-200/70'}`} />
              {today && <span className="absolute inset-0 animate-ping rounded-full bg-emerald-300 opacity-30" />}
              <button
                onClick={today ? undefined : handleCheckIn}
                disabled={!!today}
                className={`relative z-10 flex h-24 w-24 items-center justify-center rounded-full shadow-lg transition ${
                  today
                    ? 'bg-gradient-to-br from-emerald-500 to-emerald-600 cursor-default'
                    : 'bg-gradient-to-br from-sky-500 to-sky-600 hover:scale-105'
                }`}
                aria-label={today ? 'متواجد' : 'سجّل حضورك'}
              >
                <svg className="h-11 w-11 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
              </button>
            </div>

            {/* النص والأزرار */}
            <div className="flex-1 text-center sm:text-right">
              <h3 className="text-2xl font-black text-[#0f2040]">
                {today ? 'متواجد بالشركة' : 'ما سجّلت حضورك بعد'}
              </h3>
              <p className={`mt-1 text-sm font-bold ${today ? 'text-emerald-600' : 'text-slate-500'}`}>
                {today ? '✅ تم تسجيل الحضور بنجاح' : 'اضغط الزر حتى تسجّل حضورك'}
              </p>

              {today && (
                <div className="mt-4 inline-block rounded-xl border border-slate-200 bg-slate-50 px-5 py-3">
                  <p className="text-[11px] text-slate-500">وقت الحضور</p>
                  <p className="text-lg font-black text-emerald-700">{fmtTime(today.checkIn)}</p>
                  <p className="mt-0.5 text-[11px] text-slate-500">هاي الجلسة: {elapsed}</p>
                </div>
              )}

              {justCheckedOut && (
                <p className="mt-3 text-sm font-bold text-emerald-600">شكراً لك، أحسنت العمل اليوم!</p>
              )}

              <div className="mt-5 flex flex-wrap justify-center gap-2 sm:justify-start">
                {today ? (
                  <button
                    onClick={handleCheckOut}
                    className="rounded-xl bg-[#2c5aad] px-5 py-2.5 text-sm font-bold text-white shadow-md transition hover:bg-[#24488c]"
                  >
                    ⬅️ تسجيل الانصراف
                  </button>
                ) : (
                  <button
                    onClick={handleCheckIn}
                    className="rounded-xl bg-emerald-600 px-5 py-2.5 text-sm font-bold text-white shadow-md transition hover:bg-emerald-700"
                  >
                    ✅ تسجيل الحضور
                  </button>
                )}
                <button
                  onClick={() => setShowAllMovements((v) => !v)}
                  className="rounded-xl border border-slate-300 bg-white px-5 py-2.5 text-sm font-bold text-slate-600 transition hover:bg-slate-50"
                >
                  ☰ عرض السجل
                </button>
              </div>
            </div>
          </div>
        </div>

        {/* ملخص اليوم */}
        <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
          <h2 className="mb-4 flex items-center gap-2 text-sm font-extrabold text-[#0f2040]">📋 ملخص اليوم</h2>
          <div className="space-y-2.5">
            <SummaryRow icon="🕐" label="إجمالي ساعات اليوم" value={fmtHours(todayMinutes)} />
            <SummaryRow
              icon="⏱️" label="الجلسة الحالية"
              value={today ? `${fmtTime(today.checkIn)} — الآن` : 'ماكو جلسة مفتوحة'}
            />
            <SummaryRow
              icon="🔁" label="عدد جلسات اليوم"
              value={todayRecord ? `${todayRecord.sessions.length}` : '0'}
            />
          </div>
          <button
            onClick={() => setShowAllMovements((v) => !v)}
            className="mt-4 flex w-full items-center justify-between rounded-xl border border-slate-200 px-3 py-2 text-xs font-bold text-slate-600 transition hover:bg-slate-50"
          >
            <span>{showAllMovements ? 'إخفاء التفاصيل' : 'عرض التفاصيل'}</span>
            <span>{showAllMovements ? '⌄' : '‹'}</span>
          </button>
        </div>
      </div>

      {/* ═══ سجل الدوام الشهري + آخر الحركات ═══ */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        <div className="lg:col-span-2">
          <MonthlyView
            month={month} setMonth={setMonth} report={report}
            employeeId={targetEmployeeId} canExport={canExport}
            shiftStart={shiftStart} lateDays={lateDays} absentDays={absentDays}
          />
        </div>
        <MovementsPanel movements={movements} expanded={showAllMovements} onToggle={() => setShowAllMovements((v) => !v)} />
      </div>

      {/* Employee picker for admins */}
      {isAdmin && (
        <div className="flex items-center gap-3 rounded-2xl bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <label className="text-sm font-semibold text-gray-600">عرض سجل موظف آخر:</label>
          <select
            value={viewedEmployeeId}
            onChange={(e) => setViewedEmployeeId(e.target.value)}
            className="rounded-lg border border-gray-200 px-3 py-1.5 text-sm"
          >
            <option value="">أنا ({employee.name})</option>
            {todaySummary.filter(s => s.employeeId !== employee.id).map(s => (
              <option key={s.employeeId} value={s.employeeId}>{s.employee?.name || s.employeeId}</option>
            ))}
          </select>
        </div>
      )}

      {/* Admin table */}
      {isAdmin && <AdminTable records={todaySummary} />}
    </div>
  )
}

/* ───── بطاقة رقم علوية ───── */

function StatCard({ icon, label, value, tone }: {
  icon: string; label: string; value: string
  tone: 'emerald' | 'sky' | 'violet' | 'amber' | 'slate'
}) {
  const tones: Record<string, string> = {
    emerald: 'text-emerald-600 bg-emerald-50',
    sky: 'text-sky-600 bg-sky-50',
    violet: 'text-violet-600 bg-violet-50',
    amber: 'text-amber-600 bg-amber-50',
    slate: 'text-slate-500 bg-slate-100',
  }
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex items-start justify-between gap-2">
        <p className="text-[11px] font-medium text-slate-500">{label}</p>
        <span className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-lg text-sm ${tones[tone]}`}>{icon}</span>
      </div>
      <p className={`mt-1.5 text-lg font-black ${tones[tone].split(' ')[0]}`}>{value}</p>
    </div>
  )
}

/* ───── سطر بملخص اليوم ───── */

function SummaryRow({ icon, label, value }: { icon: string; label: string; value: string }) {
  return (
    <div className="flex items-center justify-between gap-2 rounded-xl border border-slate-100 bg-slate-50/60 px-3 py-2.5">
      <div className="min-w-0">
        <p className="text-[11px] text-slate-500">{label}</p>
        <p className="truncate text-xs font-bold text-slate-800">{value}</p>
      </div>
      <span className="shrink-0 text-base">{icon}</span>
    </div>
  )
}

/* ───── آخر الحركات ─────
 *
 * الخروج بنص اليوم والرجوع بعده = استراحة، مو انصراف. التمييز
 * يخلي الموظف يشوف يومه مثل ما صار فعلاً، بدل «تسجيل انصراف» أربع
 * مرات بيوم واحد. */

function MovementsPanel({ movements, expanded, onToggle }: {
  movements: Movement[]; expanded: boolean; onToggle: () => void
}) {
  const shown = expanded ? movements : movements.slice(0, 4)
  const style: Record<Movement['kind'], { color: string; icon: string }> = {
    in:    { color: 'text-emerald-600 bg-emerald-50', icon: '→' },
    break: { color: 'text-amber-600 bg-amber-50',     icon: '⏸' },
    back:  { color: 'text-sky-600 bg-sky-50',         icon: '↩' },
    out:   { color: 'text-slate-600 bg-slate-100',    icon: '←' },
  }

  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
      <h2 className="mb-4 flex items-center gap-2 text-sm font-extrabold text-[#0f2040]">🕓 آخر الحركات</h2>

      {shown.length === 0 ? (
        <p className="py-6 text-center text-xs text-slate-400">ماكو حركات اليوم</p>
      ) : (
        <div className="space-y-2.5">
          {shown.map((m, i) => (
            <div key={`${m.at}-${i}`} className="flex items-center justify-between gap-2">
              <div className="flex items-center gap-2.5">
                <span className={`flex h-7 w-7 items-center justify-center rounded-full text-xs font-bold ${style[m.kind].color}`}>
                  {style[m.kind].icon}
                </span>
                <span className="text-xs font-bold text-slate-700">{m.label}</span>
              </div>
              <div className="text-left">
                <p className="text-[11px] font-bold text-slate-600">{fmtTime(m.at)}</p>
                <p className="text-[10px] text-slate-400">اليوم</p>
              </div>
            </div>
          ))}
        </div>
      )}

      {movements.length > 4 && (
        <button
          onClick={onToggle}
          className="mt-4 flex w-full items-center justify-between rounded-xl border border-slate-200 px-3 py-2 text-xs font-bold text-slate-600 transition hover:bg-slate-50"
        >
          <span>{expanded ? 'عرض أقل' : `عرض جميع الحركات (${movements.length})`}</span>
          <span>{expanded ? '⌄' : '‹'}</span>
        </button>
      )}
    </div>
  )
}

/* ───── Monthly attendance table ───── */

function MonthlyView({ month, setMonth, report, employeeId, canExport, shiftStart, lateDays, absentDays }: {
  canExport: boolean
  month: string
  setMonth: (m: string) => void
  report: MonthlyAttendanceReport | null
  employeeId: string
  shiftStart: string | null
  lateDays: number
  absentDays: number
}) {
  const canGoForward = month < currentMonthKey()
  const [exporting, setExporting] = useState(false)

  const handleExport = async () => {
    if (!employeeId) return
    setExporting(true)
    try {
      await api.exportEmployeeAttendance(employeeId, month)
    } finally {
      setExporting(false)
    }
  }

  return (
    <div className="overflow-hidden rounded-2xl bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
      <div className="flex items-center justify-between bg-gradient-to-l from-[#0f2040] to-[#2c5aad] px-8 py-5 text-white">
        <button onClick={() => setMonth(shiftMonth(month, -1))} className="rounded-lg p-2 hover:bg-white/10" aria-label="الشهر السابق">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M15 18l-6-6 6-6" /></svg>
        </button>
        <div className="text-center">
          <h2 className="text-xl font-bold">سجل الدوام الشهري</h2>
          <p className="text-sm text-blue-200">{monthLabel(month)}</p>
        </div>
        <button
          onClick={() => canGoForward && setMonth(shiftMonth(month, 1))}
          disabled={!canGoForward}
          className="rounded-lg p-2 hover:bg-white/10 disabled:opacity-30"
          aria-label="الشهر التالي"
        >
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M9 18l6-6-6-6" /></svg>
        </button>
      </div>

      {canExport && (
        <div className="flex items-center justify-between border-b border-gray-100 px-8 py-3">
          <button
            onClick={handleExport}
            disabled={exporting || !report || report.days.length === 0}
            className="rounded-lg bg-emerald-50 px-4 py-1.5 text-xs font-bold text-emerald-700 hover:bg-emerald-100 disabled:opacity-40"
          >
            {exporting ? 'جارٍ التصدير...' : 'تصدير Excel'}
          </button>
        </div>
      )}

      {/* الأرقام الأربعة */}
      {report && (
        <div className="grid grid-cols-2 gap-px border-b border-slate-100 bg-slate-100 sm:grid-cols-4">
          <MiniStat icon="🕐" label="إجمالي الساعات" value={fmtHours(report.totalMinutes)} tone="text-sky-700" />
          <MiniStat icon="📅" label="أيام الحضور" value={`${report.daysPresent} أيام`} tone="text-emerald-700" />
          <MiniStat
            icon="⏰" label="أيام التأخير"
            value={shiftStart ? `${lateDays} يوم` : 'ماكو دوام محدد'}
            tone={lateDays > 0 ? 'text-amber-700' : 'text-slate-600'}
            hint={shiftStart ? `الدوام ${shiftStart}` : 'حدد وقت الدوام بملف الموظف'}
          />
          <MiniStat
            icon="🚫" label="أيام بلا بصمة"
            value={`${absentDays} يوم`}
            tone={absentDays > 0 ? 'text-red-700' : 'text-slate-600'}
            hint="عدا الجمعة — وممكن تكون إجازة مصدّقة"
          />
        </div>
      )}

      {!report || report.days.length === 0 ? (
        <p className="p-8 text-center text-gray-400">لا توجد سجلات حضور بهذا الشهر</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-right text-sm">
            <thead className="bg-gray-50 text-gray-600">
              <tr>
                <th className="px-6 py-3 font-semibold">اليوم</th>
                <th className="px-6 py-3 font-semibold">وقت الحضور</th>
                <th className="px-6 py-3 font-semibold">وقت الانصراف</th>
                <th className="px-6 py-3 font-semibold">عدد الساعات</th>
                <th className="px-6 py-3 font-semibold">الحالة</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {report.days.map((d) => {
                // الحالة: لسه شغّال / تأخر / مكتمل. التأخير يطلع بعموده
                // حتى يعرف الموظف أي يوم بالضبط انحسب عليه.
                const late = isLateDay(d, shiftStart)
                const badge = d.stillOpen
                  ? { text: 'مفتوح', cls: 'bg-sky-50 text-sky-700' }
                  : late
                    ? { text: 'متأخر', cls: 'bg-amber-50 text-amber-700' }
                    : { text: 'مكتمل', cls: 'bg-emerald-50 text-emerald-700' }
                return (
                <tr key={d.date} className="hover:bg-gray-50">
                  <td className="px-6 py-4 font-medium text-gray-800">{dayLabel(d.firstCheckIn)}</td>
                  <td className="px-6 py-4 text-gray-600">{fmtTime(d.firstCheckIn)}</td>
                  <td className="px-6 py-4 text-gray-600">
                    {d.stillOpen ? <span className="text-amber-600 font-semibold">لم يسجل انصراف بعد</span> : fmtTime(d.lastCheckOut)}
                  </td>
                  <td className="px-6 py-4 text-gray-600">
                    {fmtHours(d.totalMinutes)}
                    {d.sessions.length > 1 && (
                      <span className="mr-1 text-[11px] text-slate-400">({d.sessions.length} جلسات)</span>
                    )}
                  </td>
                  <td className="px-6 py-4">
                    <span className={`inline-block rounded-full px-2.5 py-1 text-[11px] font-bold ${badge.cls}`}>
                      ● {badge.text}
                    </span>
                  </td>
                </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

/* ───── Admin attendance table (today summary + correction) ───── */

function AdminTable({ records }: { records: EmployeeDailyAttendanceSummary[] }) {
  const [exporting, setExporting] = useState(false)

  const handleExport = async () => {
    setExporting(true)
    try {
      await api.exportTodayAttendance()
    } finally {
      setExporting(false)
    }
  }

  return (
    <div className="overflow-hidden rounded-2xl bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
      <div className="flex items-center justify-between bg-gradient-to-l from-[#0f2040] to-[#2c5aad] px-8 py-5 text-white">
        <div>
          <h2 className="text-xl font-bold">حضور الموظفين اليوم</h2>
          <p className="mt-1 text-sm text-blue-200">ملخص كل جلسات كل موظف باليوم الحالي</p>
        </div>
        <button
          onClick={handleExport}
          disabled={exporting}
          className="rounded-lg bg-white/10 px-4 py-1.5 text-xs font-bold text-white hover:bg-white/20 disabled:opacity-40"
        >
          {exporting ? 'جارٍ التصدير...' : 'تصدير إكسل لكل الموظفين'}
        </button>
      </div>

      {records.length === 0 ? (
        <p className="p-8 text-center text-gray-400">لا توجد سجلات حضور اليوم</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full text-right text-sm">
            <thead className="bg-gray-50 text-gray-600">
              <tr>
                <th className="px-6 py-3 font-semibold">اسم الموظف</th>
                <th className="px-6 py-3 font-semibold">وقت الدخول الأول</th>
                <th className="px-6 py-3 font-semibold">وقت الخروج الأخير</th>
                <th className="px-6 py-3 font-semibold">عدد الجلسات</th>
                <th className="px-6 py-3 font-semibold">مجموع الساعات اليوم</th>
                <th className="px-6 py-3 font-semibold">الحالة</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {records.map((r) => {
                const status = r.currentlyActive ? 'نشط الآن' : 'منتهي'
                const color = r.currentlyActive ? 'text-green-600 bg-green-50' : 'text-amber-600 bg-amber-50'
                return (
                  <tr key={r.employeeId} className="hover:bg-gray-50">
                    <td className="px-6 py-4 font-medium text-gray-800">{r.employee?.name || '—'}</td>
                    <td className="px-6 py-4 text-gray-600">{fmtTime(r.firstCheckIn)}</td>
                    <td className="px-6 py-4 text-gray-600">{fmtTime(r.lastCheckOut)}</td>
                    <td className="px-6 py-4 text-gray-600">{r.sessionsCount}</td>
                    <td className="px-6 py-4 text-gray-600">{fmtHours(r.totalMinutes)}</td>
                    <td className="px-6 py-4">
                      <span className={`inline-block rounded-full px-3 py-1 text-xs font-bold ${color}`}>{status}</span>
                    </td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

/* ───── رقم صغير بترويسة الجدول الشهري ───── */

function MiniStat({ icon, label, value, tone, hint }: {
  icon: string; label: string; value: string; tone: string; hint?: string
}) {
  return (
    <div className="bg-white p-4 text-center" title={hint}>
      <div className="mb-1 flex items-center justify-center gap-1.5">
        <span className="text-sm">{icon}</span>
        <p className="text-[11px] text-slate-500">{label}</p>
      </div>
      <p className={`text-base font-black ${tone}`}>{value}</p>
    </div>
  )
}
