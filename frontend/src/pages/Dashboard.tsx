import { useEffect, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'
import type { Booking } from '../api'
import { useSession } from '../session'

/* ───── Attendance localStorage helpers ───── */

const todayKey = () => new Date().toISOString().slice(0, 10)

interface AttendanceRecord {
  employeeId: string
  employeeName: string
  date: string
  checkInTime: string | null
  checkOutTime: string | null
}

function getAttendanceRecords(): AttendanceRecord[] {
  try { return JSON.parse(localStorage.getItem('attendance_records') || '[]') } catch { return [] }
}
function saveAttendanceRecords(records: AttendanceRecord[]) {
  localStorage.setItem('attendance_records', JSON.stringify(records))
}
function getMyTodayRecord(employeeId: string): AttendanceRecord | undefined {
  return getAttendanceRecords().find(r => r.employeeId === employeeId && r.date === todayKey() && !r.checkOutTime)
}
function getMyTodayRecords(employeeId: string): AttendanceRecord[] {
  return getAttendanceRecords().filter(r => r.employeeId === employeeId && r.date === todayKey())
}

function elapsedSince(timeStr: string): string {
  const digits = timeStr.replace(/[^\d:]/g, '')
  const [h, m] = digits.split(':').map(Number)
  if (isNaN(h) || isNaN(m)) return '--'
  const now = new Date()
  const checkInDate = new Date()
  checkInDate.setHours(h, m, 0, 0)
  const diffMs = now.getTime() - checkInDate.getTime()
  if (diffMs < 0) return '٠ دقائق'
  const diffMin = Math.floor(diffMs / 60000)
  const hours = Math.floor(diffMin / 60)
  const mins = diffMin % 60
  if (hours === 0) return `${mins} دقيقة`
  return `${hours} ساعة و ${mins} دقيقة`
}

function formatTime(): string {
  return new Date().toLocaleTimeString('ar-SA', { hour: '2-digit', minute: '2-digit' })
}

function formatDate(): string {
  return new Date().toLocaleDateString('ar-IQ', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })
}

/* ───── Types ───── */

interface GpsStats {
  totalDevices: number
  totalCustomers: number
  devicesByStatus: { status: string; count: number }[]
  totalSims: number
  availableSims: number
  inUseSims: number
}

/* ───── Animated Counter ───── */
function AnimatedNumber({ value, loading }: { value: number; loading: boolean }) {
  const [display, setDisplay] = useState(0)
  useEffect(() => {
    if (loading) return
    let start = 0
    const duration = 800
    const step = Math.ceil(value / (duration / 16))
    const timer = setInterval(() => {
      start += step
      if (start >= value) { setDisplay(value); clearInterval(timer) }
      else setDisplay(start)
    }, 16)
    return () => clearInterval(timer)
  }, [value, loading])
  if (loading) return <span className="inline-block h-8 w-16 animate-pulse rounded-lg bg-slate-200" />
  return <>{display.toLocaleString('ar-SA')}</>
}

/* ───── Progress Ring ───── */
function ProgressRing({ percent, size = 56, stroke = 5, color }: { percent: number; size?: number; stroke?: number; color: string }) {
  const r = (size - stroke) / 2
  const circ = 2 * Math.PI * r
  const offset = circ - (percent / 100) * circ
  return (
    <svg width={size} height={size} className="-rotate-90">
      <circle cx={size/2} cy={size/2} r={r} fill="none" stroke="currentColor" strokeWidth={stroke} className="text-slate-100" />
      <circle cx={size/2} cy={size/2} r={r} fill="none" stroke={color} strokeWidth={stroke}
        strokeDasharray={circ} strokeDashoffset={offset} strokeLinecap="round"
        className="transition-all duration-1000 ease-out" />
    </svg>
  )
}

export default function Dashboard() {
  const { employee, permissions } = useSession()
  const navigate = useNavigate()
  const [gpsStats, setGpsStats] = useState<GpsStats | null>(null)
  const [bookingCount, setBookingCount] = useState(0)
  const [employeeCount, setEmployeeCount] = useState(0)
  const [customerCount, setCustomerCount] = useState(0)
  const [bookings, setBookings] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [currentTime, setCurrentTime] = useState(formatTime())

  useEffect(() => {
    const t = setInterval(() => setCurrentTime(formatTime()), 60000)
    return () => clearInterval(t)
  }, [])

  useEffect(() => {
    if (!employee) return
    Promise.all([
      api.getGpsStats().catch(() => null),
      api.getBookings().catch(() => [] as Booking[]),
      api.getEmployees().then((e) => e.length).catch(() => 0),
      api.getCustomers().then((c) => c.length).catch(() => 0),
    ]).then(([gps, bk, emp, cust]) => {
      setGpsStats(gps as GpsStats | null)
      setBookings(bk as Booking[])
      setBookingCount((bk as Booking[]).length)
      setEmployeeCount(emp)
      setCustomerCount(cust)
    }).finally(() => setLoading(false))
  }, [employee])

  /* ── Attendance widget state ── */
  const [activeRecord, setActiveRecord] = useState<AttendanceRecord | undefined>(undefined)
  const [allTodayRecords, setAllTodayRecords] = useState<AttendanceRecord[]>([])
  const [elapsed, setElapsed] = useState('')
  const [showCheckoutConfirm, setShowCheckoutConfirm] = useState(false)

  useEffect(() => {
    if (!employee) return
    setActiveRecord(getMyTodayRecord(employee.id))
    setAllTodayRecords(getMyTodayRecords(employee.id))
  }, [employee])

  useEffect(() => {
    if (!activeRecord?.checkInTime || activeRecord.checkOutTime) return
    setElapsed(elapsedSince(activeRecord.checkInTime))
    const interval = setInterval(() => {
      setElapsed(elapsedSince(activeRecord.checkInTime!))
    }, 60000)
    return () => clearInterval(interval)
  }, [activeRecord])

  const handleAttCheckIn = useCallback(() => {
    if (!employee) return
    const records = getAttendanceRecords()
    const now = new Date().toLocaleTimeString('ar-SA', { hour: '2-digit', minute: '2-digit' })
    const rec: AttendanceRecord = { employeeId: employee.id, employeeName: employee.name, date: todayKey(), checkInTime: now, checkOutTime: null }
    records.push(rec)
    saveAttendanceRecords(records)
    setActiveRecord(rec)
    setAllTodayRecords(getMyTodayRecords(employee.id))
  }, [employee])

  const handleAttCheckOut = useCallback(() => {
    if (!employee || !activeRecord) return
    const records = getAttendanceRecords()
    const rec = records.find(r => r.employeeId === employee.id && r.date === todayKey() && !r.checkOutTime)
    if (rec) {
      rec.checkOutTime = new Date().toLocaleTimeString('ar-SA', { hour: '2-digit', minute: '2-digit' })
      saveAttendanceRecords(records)
      setActiveRecord(undefined)
      setAllTodayRecords(getMyTodayRecords(employee.id))
    }
    setShowCheckoutConfirm(false)
  }, [employee, activeRecord])

  if (!employee) return null

  const isAdmin = employee.role === 'ADMIN'
  const pendingMaintenance = gpsStats?.devicesByStatus?.find((d) => d.status === 'MAINTENANCE')?.count || 0

  const quickCards = [
    {
      title: 'حجز جديد',
      desc: 'إنشاء حجز خدمة جديد للعميل',
      gradient: 'from-blue-500 via-blue-600 to-blue-700',
      iconPath: 'M12 4v16m8-8H4',
      path: '/sales',
      visible: ['ADMIN', 'SALES', 'HR_COORDINATOR'].includes(employee.role),
    },
    {
      title: 'عرض سعر جديد',
      desc: 'إنشاء عرض سعر احترافي للعميل',
      gradient: 'from-emerald-500 via-emerald-600 to-emerald-700',
      iconPath: 'M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z',
      path: '/quotations/new',
      visible: isAdmin || permissions.includes('quotation_system'),
    },
    {
      title: 'إضافة عميل',
      desc: 'تسجيل عميل جديد في النظام',
      gradient: 'from-violet-500 via-violet-600 to-violet-700',
      iconPath: 'M18 9v3m0 0v3m0-3h3m-3 0h-3m-2-5a4 4 0 11-8 0 4 4 0 018 0zM3 20a6 6 0 0112 0v1H3v-1z',
      path: '/customers',
      visible: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'].includes(employee.role),
    },
    {
      title: 'تتبع المهام',
      desc: 'متابعة المهام الميدانية والفرق',
      gradient: 'from-amber-500 via-amber-600 to-amber-700',
      iconPath: 'M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2m-6 9l2 2 4-4',
      path: '/missions',
      visible: isAdmin || ['HR_COORDINATOR', 'MONITOR'].includes(employee.role),
    },
  ].filter(c => c.visible)

  const completedSessions = allTodayRecords.filter(r => r.checkOutTime)

  const kpiCards = [
    { title: 'الموظفين', value: employeeCount, color: '#3b82f6', bg: 'from-blue-500/10 to-blue-500/5', ring: 75, path: '/employees',
      icon: <path d="M17 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2M9 7a4 4 0 100-8 4 4 0 000 8zM23 21v-2a4 4 0 00-3-3.87M16 3.13a4 4 0 010 7.75" /> },
    { title: 'العملاء', value: customerCount, color: '#10b981', bg: 'from-emerald-500/10 to-emerald-500/5', ring: 60, path: '/customers',
      icon: <><path d="M16 21v-2a4 4 0 00-4-4H5a4 4 0 00-4 4v2" /><circle cx="8.5" cy="7" r="4" /><line x1="20" y1="8" x2="20" y2="14" /><line x1="23" y1="11" x2="17" y2="11" /></> },
    { title: 'الحجوزات', value: bookingCount, color: '#8b5cf6', bg: 'from-violet-500/10 to-violet-500/5', ring: 45, path: '/bookings',
      icon: <><rect x="3" y="4" width="18" height="18" rx="2" /><line x1="16" y1="2" x2="16" y2="6" /><line x1="8" y1="2" x2="8" y2="6" /><line x1="3" y1="10" x2="21" y2="10" /></> },
    { title: 'أجهزة GPS', value: gpsStats?.totalDevices || 0, color: '#f59e0b', bg: 'from-amber-500/10 to-amber-500/5', ring: 85, path: '/gps',
      icon: <><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0118 0z" /><circle cx="12" cy="10" r="3" /></> },
  ]

  return (
    <div dir="rtl" className="space-y-6">
      {/* ═══ Hero Section ═══ */}
      <div className="relative overflow-hidden rounded-3xl bg-gradient-to-l from-[#0a1628] via-[#1a3a6e] to-[#2c5aad] p-8 text-white">
        {/* Decorative elements */}
        <div className="absolute inset-0 overflow-hidden">
          <div className="absolute -left-20 -top-20 h-64 w-64 rounded-full bg-white/5 blur-3xl" />
          <div className="absolute -bottom-10 left-1/3 h-48 w-48 rounded-full bg-blue-400/10 blur-2xl" />
          <div className="absolute right-10 top-10 h-32 w-32 rounded-full bg-blue-300/5 blur-xl" />
          {/* Grid pattern */}
          <div className="absolute inset-0 opacity-[0.03]" style={{ backgroundImage: 'radial-gradient(circle, white 1px, transparent 1px)', backgroundSize: '24px 24px' }} />
        </div>

        <div className="relative z-10 flex flex-wrap items-center justify-between gap-6">
          <div className="flex-1">
            <div className="mb-1 flex items-center gap-2">
              <span className="inline-flex items-center gap-1.5 rounded-full bg-white/10 px-3 py-1 text-xs font-medium backdrop-blur-sm">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" />
                متصل
              </span>
            </div>
            <h1 className="text-3xl font-bold tracking-tight">مرحباً، {employee.name}</h1>
            <p className="mt-2 text-sm text-blue-200/80">{formatDate()}</p>
            <p className="mt-1 text-xs text-blue-300/60">نظام إدارة شامل — شركة الأماني</p>
          </div>

          {/* Time display */}
          <div className="flex flex-col items-center rounded-2xl bg-white/10 px-6 py-4 backdrop-blur-md">
            <span className="text-4xl font-bold tracking-wider tabular-nums">{currentTime}</span>
            <span className="mt-1 text-xs text-blue-200/70">التوقيت المحلي</span>
          </div>
        </div>
      </div>

      {/* ═══ KPI Cards - ADMIN ═══ */}
      {isAdmin && (
        <div className="grid grid-cols-2 gap-4 lg:grid-cols-4">
          {kpiCards.map((kpi) => (
            <button key={kpi.title} onClick={() => navigate(kpi.path)}
              className="group relative overflow-hidden rounded-2xl bg-white p-5 text-right shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.04)] transition-all duration-300 hover:shadow-[0_8px_32px_rgba(44,90,173,0.12)] hover:-translate-y-0.5">
              {/* Gradient background */}
              <div className={`absolute inset-0 bg-gradient-to-bl ${kpi.bg} opacity-0 transition-opacity group-hover:opacity-100`} />
              <div className="relative flex items-start justify-between">
                <div className="relative">
                  <ProgressRing percent={kpi.ring} color={kpi.color} />
                  <div className="absolute inset-0 flex items-center justify-center">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke={kpi.color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                      {kpi.icon}
                    </svg>
                  </div>
                </div>
                <div>
                  <p className="text-xs font-medium text-slate-400">{kpi.title}</p>
                  <p className="mt-1 text-3xl font-black text-slate-800">
                    <AnimatedNumber value={kpi.value} loading={loading} />
                  </p>
                </div>
              </div>
              {/* Hover arrow */}
              <div className="absolute bottom-3 left-3 opacity-0 transition-all group-hover:opacity-100 group-hover:translate-x-0 translate-x-2">
                <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke={kpi.color} strokeWidth="2.5" strokeLinecap="round"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
              </div>
            </button>
          ))}
        </div>
      )}

      {/* ═══ Bookings Notification Panel ═══ */}
      {['ADMIN', 'HR_COORDINATOR', 'MONITOR'].includes(employee.role) && (() => {
        const pending = bookings.filter(b => b.status === 'PENDING')
        const confirmed = bookings.filter(b => b.status === 'CONFIRMED')
        const inProgress = bookings.filter(b => b.status === 'IN_PROGRESS')
        if (pending.length === 0 && confirmed.length === 0 && inProgress.length === 0) return null
        return (
          <div className="space-y-4">
            {/* Pending - urgent notification */}
            {pending.length > 0 && (
              <div className="rounded-2xl border-2 border-amber-200 bg-gradient-to-l from-amber-50 to-orange-50 p-5 shadow-lg shadow-amber-100/50">
                <div className="mb-4 flex items-center justify-between">
                  <button onClick={() => navigate('/coordinator')} className="rounded-lg bg-amber-500 px-4 py-1.5 text-xs font-bold text-white transition hover:bg-amber-600">
                    تنسيق الحجوزات ←
                  </button>
                  <div className="flex items-center gap-3">
                    <span className="relative flex h-3 w-3">
                      <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-amber-400 opacity-75" />
                      <span className="relative inline-flex h-3 w-3 rounded-full bg-amber-500" />
                    </span>
                    <h3 className="text-base font-extrabold text-amber-800">
                      حجوزات بانتظار التثبيت
                      <span className="mr-2 inline-flex h-7 w-7 items-center justify-center rounded-full bg-amber-500 text-sm font-black text-white">{pending.length}</span>
                    </h3>
                  </div>
                </div>
                <div className="space-y-2 max-h-64 overflow-y-auto">
                  {pending.map(b => (
                    <div key={b.id} onClick={() => navigate('/coordinator')} className="flex cursor-pointer items-center justify-between rounded-xl bg-white/80 px-4 py-3 transition hover:bg-white hover:shadow-md">
                      <div className="flex items-center gap-2">
                        {b.priority === 'URGENT' && <span className="rounded-full bg-red-100 px-2 py-0.5 text-[10px] font-bold text-red-600">عاجل</span>}
                        <span className="text-xs text-slate-400">{new Date(b.createdAt).toLocaleDateString('ar-IQ')}</span>
                      </div>
                      <div className="flex items-center gap-3 text-right">
                        <div>
                          <span className="text-sm font-bold text-amber-700">{b.code}</span>
                          <span className="mr-2 text-sm font-medium text-slate-700">{b.customer.name}</span>
                        </div>
                        {b.service && <span className="rounded-lg bg-amber-100 px-2 py-0.5 text-[11px] font-medium text-amber-700">{b.service.name}</span>}
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Confirmed + In Progress - summary row */}
            {(confirmed.length > 0 || inProgress.length > 0) && (
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                {confirmed.length > 0 && (
                  <button onClick={() => navigate('/coordinator')} className="group flex items-center justify-between rounded-2xl border border-blue-200 bg-blue-50 p-4 text-right transition hover:bg-blue-100 hover:shadow-md">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2" strokeLinecap="round" className="opacity-40 group-hover:opacity-100"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
                    <div className="flex items-center gap-3">
                      <div>
                        <span className="text-sm font-bold text-blue-800">حجوزات مثبتة بحاجة تنسيق</span>
                        <p className="text-xs text-blue-500 mt-0.5">تعيين كوادر وموعد ومركبة</p>
                      </div>
                      <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-blue-500 text-lg font-black text-white shadow-lg shadow-blue-200">{confirmed.length}</div>
                    </div>
                  </button>
                )}
                {inProgress.length > 0 && (
                  <button onClick={() => navigate('/bookings')} className="group flex items-center justify-between rounded-2xl border border-violet-200 bg-violet-50 p-4 text-right transition hover:bg-violet-100 hover:shadow-md">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="#8b5cf6" strokeWidth="2" strokeLinecap="round" className="opacity-40 group-hover:opacity-100"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
                    <div className="flex items-center gap-3">
                      <div>
                        <span className="text-sm font-bold text-violet-800">حجوزات جاري تنفيذها</span>
                        <p className="text-xs text-violet-500 mt-0.5">متابعة الكوادر الميدانية</p>
                      </div>
                      <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-violet-500 text-lg font-black text-white shadow-lg shadow-violet-200">{inProgress.length}</div>
                    </div>
                  </button>
                )}
              </div>
            )}
          </div>
        )
      })()}

      {/* ═══ Sales Level Card ═══ */}
      {employee.role === 'SALES' && (
        <div className="rounded-2xl bg-white p-5 shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.04)]">
          <div className="flex items-center gap-4">
            <div className="relative">
              <ProgressRing percent={Math.min(100, (bookingCount % 10) * 10)} color="#10b981" />
              <div className="absolute inset-0 flex items-center justify-center">
                <span className="text-lg font-extrabold text-emerald-600">{Math.floor(bookingCount / 10) + 1}</span>
              </div>
            </div>
            <div className="flex-1">
              <h3 className="text-sm font-bold text-[#0f2040]">مستوى المبيعات</h3>
              <p className="text-xs text-slate-400 mt-0.5">كل 10 حجوزات = مستوى جديد</p>
              <div className="mt-2 flex items-center gap-4">
                <div className="flex items-center gap-1.5">
                  <span className="h-2 w-2 rounded-full bg-emerald-500"/>
                  <span className="text-xs text-slate-500"><span className="font-bold text-[#0f2040]">{bookingCount}</span> حجز</span>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="h-2 w-2 rounded-full bg-blue-500"/>
                  <span className="text-xs text-slate-500"><span className="font-bold text-[#0f2040]">{10 - (bookingCount % 10)}</span> للمستوى التالي</span>
                </div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ═══ Attendance + Quick Access Row ═══ */}
      <div className={`grid gap-4 ${quickCards.length > 0 ? 'grid-cols-1 lg:grid-cols-5' : 'grid-cols-1'}`}>
        {/* Attendance Widget */}
        <div className={`${quickCards.length > 0 ? 'lg:col-span-2' : ''} overflow-hidden rounded-2xl bg-white shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.04)]`}>
          <div className="flex items-center justify-between bg-gradient-to-l from-[#0f2040] to-[#1e4276] px-5 py-3">
            <div className="flex items-center gap-2 text-white">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-white/15">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round"><circle cx="12" cy="12" r="10" /><polyline points="12 6 12 12 16 14" /></svg>
              </div>
              <h3 className="text-sm font-bold">الحضور والانصراف</h3>
            </div>
            {activeRecord && (
              <span className="flex items-center gap-1.5 rounded-full bg-emerald-500/20 px-2.5 py-0.5 text-xs text-emerald-300">
                <span className="relative flex h-2 w-2"><span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-75" /><span className="relative inline-flex h-2 w-2 rounded-full bg-emerald-400" /></span>
                متواجد
              </span>
            )}
          </div>
          <div className="p-5">
            {activeRecord ? (
              <div className="space-y-3">
                <div className="flex items-center gap-3 rounded-xl bg-emerald-50 px-4 py-3">
                  <div className="flex h-10 w-10 items-center justify-center rounded-full bg-emerald-100">
                    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="#059669" strokeWidth="2.5" strokeLinecap="round"><polyline points="20 6 9 17 4 12" /></svg>
                  </div>
                  <div className="flex-1">
                    <p className="text-sm font-bold text-emerald-800">تم تسجيل الحضور</p>
                    <p className="text-xs text-emerald-600">منذ {activeRecord.checkInTime} • {elapsed}</p>
                  </div>
                </div>
                <div className="relative">
                  <button onClick={() => setShowCheckoutConfirm(true)}
                    className="w-full rounded-xl bg-gradient-to-l from-red-500 to-red-600 px-4 py-2.5 text-sm font-bold text-white shadow-lg shadow-red-500/20 transition-all hover:shadow-xl hover:shadow-red-500/30">
                    تسجيل انصراف
                  </button>
                  {showCheckoutConfirm && (
                    <div className="absolute left-0 right-0 top-full z-20 mt-2 rounded-xl border border-slate-100 bg-white p-4 shadow-2xl">
                      <p className="mb-3 text-sm font-semibold text-gray-800">هل تود تسجيل الانصراف؟</p>
                      <div className="flex gap-2">
                        <button onClick={handleAttCheckOut} className="flex-1 rounded-xl bg-red-500 px-3 py-2 text-sm font-bold text-white">نعم</button>
                        <button onClick={() => setShowCheckoutConfirm(false)} className="flex-1 rounded-xl border border-slate-200 px-3 py-2 text-sm font-bold text-slate-600">إلغاء</button>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            ) : completedSessions.length > 0 ? (
              <div className="space-y-2">
                {completedSessions.map((s, i) => (
                  <div key={i} className="flex items-center gap-3 rounded-xl bg-slate-50 px-4 py-2.5 text-sm">
                    <div className="flex h-7 w-7 items-center justify-center rounded-full bg-blue-100 text-xs font-bold text-blue-600">{i + 1}</div>
                    <span className="font-bold text-[#2c5aad]">{s.checkInTime}</span>
                    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#94a3b8" strokeWidth="2"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
                    <span className="font-bold text-[#2c5aad]">{s.checkOutTime}</span>
                  </div>
                ))}
                <button onClick={handleAttCheckIn}
                  className="w-full rounded-xl bg-gradient-to-l from-emerald-500 to-emerald-600 px-4 py-2.5 text-sm font-bold text-white shadow-lg shadow-emerald-500/20 transition-all hover:shadow-xl">
                  تسجيل حضور جديد
                </button>
              </div>
            ) : (
              <div className="text-center">
                <div className="mx-auto mb-3 flex h-16 w-16 items-center justify-center rounded-full bg-emerald-50">
                  <svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2" strokeLinecap="round"><circle cx="12" cy="12" r="10" /><polyline points="12 6 12 12 16 14" /></svg>
                </div>
                <p className="mb-3 text-sm text-slate-400">لم تسجل حضورك بعد</p>
                <button onClick={handleAttCheckIn}
                  className="w-full rounded-xl bg-gradient-to-l from-emerald-500 to-emerald-600 px-6 py-3 text-sm font-bold text-white shadow-lg shadow-emerald-500/20 transition-all hover:shadow-xl hover:shadow-emerald-500/30 hover:-translate-y-0.5">
                  تسجيل حضور
                </button>
              </div>
            )}
          </div>
        </div>

        {/* Quick Access */}
        {quickCards.length > 0 && (
          <div className="lg:col-span-3 grid grid-cols-1 gap-3 sm:grid-cols-2">
            {quickCards.map((card) => (
              <button key={card.path} onClick={() => navigate(card.path)}
                className={`group relative overflow-hidden rounded-2xl bg-gradient-to-bl ${card.gradient} p-5 text-right text-white shadow-lg transition-all duration-300 hover:shadow-2xl hover:-translate-y-1`}>
                {/* Shine effect */}
                <div className="absolute inset-0 bg-gradient-to-l from-white/0 via-white/10 to-white/0 translate-x-full transition-transform duration-700 group-hover:-translate-x-full" />
                <div className="relative flex items-center gap-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-white/20 backdrop-blur-sm">
                    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><path d={card.iconPath} /></svg>
                  </div>
                  <div className="flex-1">
                    <p className="text-base font-bold">{card.title}</p>
                    <p className="mt-0.5 text-xs text-white/70">{card.desc}</p>
                  </div>
                  <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" className="opacity-40 transition-all group-hover:opacity-100 group-hover:-translate-x-1"><path d="M19 12H5M12 19l-7-7 7-7" /></svg>
                </div>
              </button>
            ))}
          </div>
        )}
      </div>

      {/* ═══ System Panels - ADMIN ═══ */}
      {isAdmin && (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {/* GPS Panel */}
          <SystemPanel title="نظام GPS" color="#f59e0b" dotColor="bg-amber-400" actionLabel="عرض الكل" onAction={() => navigate('/gps')}>
            <div className="grid grid-cols-2 gap-3">
              <MiniKPI label="الأجهزة" value={gpsStats?.totalDevices || 0} color="#10b981" />
              <MiniKPI label="المشتركين" value={gpsStats?.totalCustomers || 0} color="#3b82f6" />
              <MiniKPI label="شرائح SIM" value={gpsStats?.totalSims || 0} color="#8b5cf6" />
              <MiniKPI label="صيانة معلقة" value={pendingMaintenance} color={pendingMaintenance > 0 ? '#ef4444' : '#f59e0b'} />
            </div>
          </SystemPanel>

          {/* Sales Panel */}
          <SystemPanel title="المبيعات والمالية" color="#3b82f6" dotColor="bg-blue-400" actionLabel="عرض الكل" onAction={() => navigate('/finance')}>
            <div className="grid grid-cols-2 gap-3">
              <MiniKPI label="الحجوزات" value={bookingCount} color="#3b82f6" />
              <MiniKPI label="العملاء" value={customerCount} color="#10b981" />
              <MiniKPI label="عروض الأسعار" value={0} color="#8b5cf6" />
              <MiniKPI label="المعاملات" value={0} color="#f59e0b" />
            </div>
          </SystemPanel>

          {/* HR Panel */}
          <SystemPanel title="الموارد البشرية" color="#10b981" dotColor="bg-emerald-400" actionLabel="عرض الكل" onAction={() => navigate('/employees')}>
            <div className="grid grid-cols-2 gap-3">
              <MiniKPI label="الموظفين" value={employeeCount} color="#3b82f6" />
              <MiniKPI label="الشكاوى" value={0} color="#ef4444" />
            </div>
          </SystemPanel>

          {/* Activity Panel */}
          <div className="rounded-2xl bg-white p-6 shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.04)]">
            <h3 className="mb-4 flex items-center gap-2 text-right text-sm font-bold text-slate-700">
              <span className="h-2 w-2 rounded-full bg-red-400" />
              آخر النشاطات
            </h3>
            <div className="flex h-28 items-center justify-center rounded-xl border-2 border-dashed border-slate-100">
              <div className="text-center">
                <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="#cbd5e1" strokeWidth="1.5" strokeLinecap="round" className="mx-auto mb-2">
                  <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
                </svg>
                <p className="text-xs text-slate-300">لا توجد نشاطات حديثة</p>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}

/* ═══ Sub-components ═══ */

function SystemPanel({ title, color: _color, dotColor, actionLabel, onAction, children }: {
  title: string; color: string; dotColor: string; actionLabel: string; onAction: () => void; children: React.ReactNode
}) {
  return (
    <div className="rounded-2xl bg-white p-6 shadow-[0_1px_3px_rgba(0,0,0,0.04),0_8px_24px_rgba(0,0,0,0.04)]">
      <div className="mb-4 flex items-center justify-between">
        <button onClick={onAction}
          className="rounded-lg bg-slate-50 px-3 py-1.5 text-xs font-bold text-slate-500 transition hover:bg-brand-50 hover:text-brand-600">
          {actionLabel} ←
        </button>
        <h3 className="flex items-center gap-2 text-sm font-bold text-slate-700">
          <span className={`h-2 w-2 rounded-full ${dotColor}`} />
          {title}
        </h3>
      </div>
      {children}
    </div>
  )
}

function MiniKPI({ label, value, color }: { label: string; value: number; color: string }) {
  return (
    <div className="group flex items-center justify-between rounded-xl border border-slate-100 p-3.5 transition-colors hover:border-slate-200 hover:bg-slate-50/50">
      <div className="flex h-10 w-10 items-center justify-center rounded-lg" style={{ backgroundColor: color + '12' }}>
        <span className="text-lg font-black" style={{ color }}>{value}</span>
      </div>
      <p className="text-xs font-medium text-slate-500">{label}</p>
    </div>
  )
}
