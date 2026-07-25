import { useEffect, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useSession } from '../session'
import { api, type MonthlyAttendanceReport, type EmployeeDailyAttendanceSummary, type OpenSessionResponse } from '../api'

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

  const isAdmin = employee?.role === 'ADMIN' || employee?.role === 'MONITOR' || permissions.includes('monitoring')

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

  return (
    <div dir="rtl" className="mx-auto max-w-4xl space-y-8 p-4">
      {/* Status card */}
      <div className="overflow-hidden rounded-2xl bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <div className="bg-gradient-to-l from-[#0f2040] to-[#2c5aad] px-8 py-6 text-white">
          <h1 className="text-2xl font-bold">حالة الحضور</h1>
          <p className="mt-1 text-sm text-blue-200">{employee.name}</p>
        </div>

        <div className="p-8 text-center">
          {error && (
            <div className="mb-4 rounded-lg bg-red-50 px-4 py-2 text-sm font-semibold text-red-600">{error}</div>
          )}

          {!today ? (
            <div className="flex flex-col items-center gap-6">
              <button
                onClick={handleCheckIn}
                className="group relative flex h-40 w-40 items-center justify-center rounded-full bg-gradient-to-br from-green-500 to-emerald-600 shadow-xl transition hover:scale-105 hover:shadow-2xl"
              >
                <span className="absolute inset-0 animate-ping rounded-full bg-green-400 opacity-20" />
                <span className="absolute inset-0 animate-pulse rounded-full bg-green-400 opacity-10" />
                <svg className="relative z-10 h-16 w-16 text-white" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z" />
                </svg>
              </button>
              <span className="text-2xl font-bold text-gray-800">متواجد بالشركة</span>
              <p className="text-gray-500">اضغط لتسجيل حضورك</p>
              {openSession && openSession.totalMinutes > 0 && (
                <p className="text-sm font-semibold text-[#2c5aad]">دوامك اليوم لحد الحين: {fmtHours(openSession.totalMinutes)}</p>
              )}
            </div>
          ) : (
            <>
              <div className="mx-auto mb-4 flex h-20 w-20 items-center justify-center rounded-full bg-green-100">
                <svg className="h-10 w-10 text-green-600" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z" />
                </svg>
              </div>
              <p className="mb-1 text-xl font-bold text-gray-800">
                أنت متواجد منذ الساعة {fmtTime(today.checkIn)}
              </p>
              <p className="mb-2 text-lg text-gray-500">مدة هذي الجلسة: {elapsed}</p>
              {openSession && (
                <p className="mb-8 text-sm font-semibold text-[#2c5aad]">مجموع دوامك اليوم: {fmtHours(openSession.totalMinutes)}</p>
              )}
              {justCheckedOut && (
                <p className="mb-4 text-lg font-semibold text-green-600">شكراً لك، أحسنت العمل اليوم!</p>
              )}
              <button
                onClick={handleCheckOut}
                className="inline-flex items-center gap-3 rounded-full bg-gradient-to-l from-red-600 to-amber-500 px-10 py-4 text-lg font-bold text-white shadow-lg transition hover:scale-105 hover:shadow-xl"
              >
                <svg className="h-7 w-7" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                  <path strokeLinecap="round" strokeLinejoin="round" d="M17 16l4-4m0 0l-4-4m4 4H7m6 4v1a2 2 0 01-2 2H5a2 2 0 01-2-2V7a2 2 0 012-2h6a2 2 0 012 2v1" />
                </svg>
                انصراف
              </button>
            </>
          )}
        </div>
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

      {/* Monthly attendance record */}
      <MonthlyView month={month} setMonth={setMonth} report={report} employeeId={targetEmployeeId} />

      {/* Admin table */}
      {isAdmin && <AdminTable records={todaySummary} />}
    </div>
  )
}

/* ───── Monthly attendance table ───── */

function MonthlyView({ month, setMonth, report, employeeId }: {
  month: string
  setMonth: (m: string) => void
  report: MonthlyAttendanceReport | null
  employeeId: string
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

      <div className="flex items-center justify-between border-b border-gray-100 px-8 py-3">
        <button
          onClick={handleExport}
          disabled={exporting || !report || report.days.length === 0}
          className="rounded-lg bg-emerald-50 px-4 py-1.5 text-xs font-bold text-emerald-700 hover:bg-emerald-100 disabled:opacity-40"
        >
          {exporting ? 'جارٍ التصدير...' : 'تصدير Excel'}
        </button>
      </div>

      {/* Summary */}
      {report && (
        <div className="grid grid-cols-2 divide-x divide-x-reverse divide-gray-100 border-b border-gray-100 sm:grid-cols-2">
          <div className="p-5 text-center">
            <p className="text-2xl font-bold text-[#2c5aad]">{report.daysPresent}</p>
            <p className="text-sm text-gray-500">يوم حضور</p>
          </div>
          <div className="p-5 text-center">
            <p className="text-2xl font-bold text-[#2c5aad]">{fmtHours(report.totalMinutes)}</p>
            <p className="text-sm text-gray-500">إجمالي ساعات الدوام</p>
          </div>
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
                <th className="px-6 py-3 font-semibold">عدد الجلسات</th>
                <th className="px-6 py-3 font-semibold">عدد الساعات</th>
              </tr>
            </thead>
            <tbody className="divide-y divide-gray-100">
              {report.days.map((d) => (
                <tr key={d.date} className="hover:bg-gray-50">
                  <td className="px-6 py-4 font-medium text-gray-800">{dayLabel(d.firstCheckIn)}</td>
                  <td className="px-6 py-4 text-gray-600">{fmtTime(d.firstCheckIn)}</td>
                  <td className="px-6 py-4 text-gray-600">
                    {d.stillOpen ? <span className="text-amber-600 font-semibold">لم يسجل انصراف بعد</span> : fmtTime(d.lastCheckOut)}
                  </td>
                  <td className="px-6 py-4 text-gray-600">{d.sessions.length > 1 ? d.sessions.length : '—'}</td>
                  <td className="px-6 py-4 text-gray-600">{fmtHours(d.totalMinutes)}</td>
                </tr>
              ))}
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
