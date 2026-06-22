import { useEffect, useState, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'
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

/* ───── Types ───── */

interface GpsStats {
  totalDevices: number
  totalCustomers: number
  devicesByStatus: { status: string; count: number }[]
  totalSims: number
  availableSims: number
  inUseSims: number
}

export default function Dashboard() {
  const { employee, permissions } = useSession()
  const navigate = useNavigate()
  const [gpsStats, setGpsStats] = useState<GpsStats | null>(null)
  const [bookingCount, setBookingCount] = useState(0)
  const [employeeCount, setEmployeeCount] = useState(0)
  const [customerCount, setCustomerCount] = useState(0)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!employee) return
    Promise.all([
      api.getGpsStats().catch(() => null),
      api.getBookings().then((b) => b.length).catch(() => 0),
      api.getEmployees().then((e) => e.length).catch(() => 0),
      api.getCustomers().then((c) => c.length).catch(() => 0),
    ]).then(([gps, bk, emp, cust]) => {
      setGpsStats(gps as GpsStats | null)
      setBookingCount(bk)
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

  // Elapsed time ticker
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

  /* ── Quick access cards with permission filtering ── */
  const quickCards = [
    {
      title: 'حجز جديد',
      desc: 'إنشاء حجز خدمة جديد للعميل',
      color: 'from-blue-500 to-blue-600',
      icon: <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2"><circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="16" /><line x1="8" y1="12" x2="16" y2="12" /></svg>,
      path: '/sales',
      visible: ['ADMIN', 'SALES', 'HR_COORDINATOR'].includes(employee.role),
    },
    {
      title: 'عرض سعر جديد',
      desc: 'إنشاء عرض سعر احترافي للعميل',
      color: 'from-emerald-500 to-emerald-600',
      icon: <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><polyline points="14 2 14 8 20 8" /><line x1="16" y1="13" x2="8" y2="13" /><line x1="16" y1="17" x2="8" y2="17" /></svg>,
      path: '/quotations/new',
      visible: isAdmin || permissions.includes('quotation_system'),
    },
    {
      title: 'إضافة عميل',
      desc: 'تسجيل عميل جديد في النظام',
      color: 'from-violet-500 to-violet-600',
      icon: <svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="8.5" cy="7" r="4" /><line x1="20" y1="8" x2="20" y2="14" /><line x1="23" y1="11" x2="17" y2="11" /></svg>,
      path: '/customers',
      visible: ['ADMIN', 'SALES', 'HR_COORDINATOR', 'MONITOR'].includes(employee.role),
    },
  ].filter(c => c.visible)

  const completedSessions = allTodayRecords.filter(r => r.checkOutTime)

  return (
    <div dir="rtl">
      {/* Hero Banner */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#2c5aad] p-8 text-white shadow-xl">
        <div className="relative z-10 max-w-2xl">
          <h1 className="text-3xl font-bold">مرحباً، {employee.name}</h1>
          <p className="mt-2 text-sm text-blue-200">نظام إدارة شامل لجميع أقسام وخدمات شركة الأماني</p>
          <p className="mt-1 text-sm text-blue-200">المبيعات • العملاء • الحجوزات • GPS • المالية • الموارد البشرية</p>
        </div>
        <div className="absolute left-8 top-1/2 -translate-y-1/2 opacity-10">
          <svg width="180" height="180" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="0.5">
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
            <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
            <line x1="12" y1="22.08" x2="12" y2="12"/>
          </svg>
        </div>
        <div className="absolute left-32 top-6 text-6xl font-extrabold opacity-5">الأماني</div>
      </div>

      {/* Quick Stats - ADMIN ONLY */}
      {isAdmin && (
      <div className="mt-8">
        <h2 className="mb-4 text-right text-lg font-bold text-slate-700">
          <span className="ml-2 inline-block h-2 w-2 rounded-full bg-brand-500"></span>
          نظرة عامة على النظام
        </h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <StatCard
            title="الموظفين"
            value={employeeCount}
            subtitle="إجمالي الموظفين"
            icon={<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" /></svg>}
            iconBg="bg-blue-50"
            loading={loading}
            onClick={() => navigate('/employees')}
          />
          <StatCard
            title="العملاء"
            value={customerCount}
            subtitle="إجمالي العملاء"
            icon={<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="8.5" cy="7" r="4" /><line x1="20" y1="8" x2="20" y2="14" /><line x1="23" y1="11" x2="17" y2="11" /></svg>}
            iconBg="bg-emerald-50"
            loading={loading}
            onClick={() => navigate('/customers')}
          />
          <StatCard
            title="الحجوزات"
            value={bookingCount}
            subtitle="إجمالي الحجوزات"
            icon={<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#8b5cf6" strokeWidth="2"><rect x="3" y="4" width="18" height="18" rx="2" /><line x1="16" y1="2" x2="16" y2="6" /><line x1="8" y1="2" x2="8" y2="6" /><line x1="3" y1="10" x2="21" y2="10" /></svg>}
            iconBg="bg-violet-50"
            loading={loading}
            onClick={() => navigate('/bookings')}
          />
          <StatCard
            title="أجهزة GPS"
            value={gpsStats?.totalDevices || 0}
            subtitle="إجمالي الأجهزة"
            icon={<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" strokeWidth="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" /><circle cx="12" cy="10" r="3" /></svg>}
            iconBg="bg-amber-50"
            loading={loading}
            onClick={() => navigate('/gps')}
          />
        </div>
      </div>
      )}

      {/* Attendance Widget */}
      <div className="mt-8">
        <div className="overflow-hidden rounded-2xl bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="flex items-center gap-2 bg-gradient-to-l from-[#0f2040] to-[#2c5aad] px-6 py-3 text-white">
            <span className="inline-block h-2.5 w-2.5 rounded-full bg-green-400"></span>
            <h3 className="text-base font-bold">الحضور والانصراف</h3>
          </div>
          <div className="p-5">
            {activeRecord ? (
              /* State B: Checked in */
              <div className="flex flex-wrap items-center justify-between gap-4">
                <div className="flex items-center gap-3">
                  <span className="relative flex h-3 w-3">
                    <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-green-400 opacity-75"></span>
                    <span className="relative inline-flex h-3 w-3 rounded-full bg-green-500"></span>
                  </span>
                  <span className="font-bold text-green-700">متواجد</span>
                  <span className="text-sm text-gray-500">منذ {activeRecord.checkInTime}</span>
                  <span className="text-sm text-gray-400">({elapsed})</span>
                </div>
                <div className="relative">
                  <button
                    onClick={() => setShowCheckoutConfirm(true)}
                    className="rounded-lg bg-red-500 px-5 py-2 text-sm font-bold text-white transition hover:bg-red-600"
                  >
                    تسجيل انصراف
                  </button>
                  {showCheckoutConfirm && (
                    <div className="absolute left-0 top-full z-20 mt-2 w-72 rounded-xl border bg-white p-4 shadow-xl">
                      <p className="mb-3 text-sm font-semibold text-gray-800">هل تود تسجيل الانصراف وإغلاق الحساب؟</p>
                      <div className="flex gap-2">
                        <button onClick={handleAttCheckOut} className="flex-1 rounded-lg bg-red-500 px-3 py-2 text-sm font-bold text-white hover:bg-red-600">نعم، انصراف</button>
                        <button onClick={() => setShowCheckoutConfirm(false)} className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm font-bold text-gray-600 hover:bg-gray-50">إلغاء</button>
                      </div>
                    </div>
                  )}
                </div>
              </div>
            ) : completedSessions.length > 0 ? (
              /* State C: Checked out, show sessions */
              <div>
                <div className="mb-3 space-y-2">
                  {completedSessions.map((s, i) => (
                    <div key={i} className="flex items-center gap-3 rounded-lg bg-gray-50 px-4 py-2 text-sm">
                      <span className="font-medium text-gray-600">الجلسة {i + 1}:</span>
                      <span className="text-[#2c5aad] font-bold">{s.checkInTime}</span>
                      <span className="text-gray-400">←</span>
                      <span className="text-[#2c5aad] font-bold">{s.checkOutTime}</span>
                    </div>
                  ))}
                </div>
                <button
                  onClick={handleAttCheckIn}
                  className="rounded-lg bg-green-500 px-5 py-2 text-sm font-bold text-white transition hover:bg-green-600"
                >
                  تسجيل حضور جديد
                </button>
              </div>
            ) : (
              /* State A: Not checked in */
              <button
                onClick={handleAttCheckIn}
                className="rounded-lg bg-green-500 px-6 py-2.5 text-sm font-bold text-white transition hover:bg-green-600"
              >
                تسجيل حضور
              </button>
            )}
          </div>
        </div>
      </div>

      {/* Quick Access Cards */}
      {quickCards.length > 0 && (
      <div className="mt-8">
        <h2 className="mb-4 text-right text-lg font-bold text-slate-700">
          <span className="ml-2 inline-block h-2 w-2 rounded-full bg-brand-500"></span>
          وصول سريع
        </h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          {quickCards.map((card) => (
            <QuickCard
              key={card.path}
              title={card.title}
              desc={card.desc}
              color={card.color}
              icon={card.icon}
              onClick={() => navigate(card.path)}
            />
          ))}
        </div>
      </div>
      )}

      {/* Systems Overview - ADMIN ONLY */}
      {isAdmin && (
      <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* GPS Summary */}
        <div className="rounded-2xl bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="mb-4 flex items-center justify-between">
            <button onClick={() => navigate('/gps')} className="rounded-lg bg-brand-50 px-3 py-1.5 text-xs font-bold text-brand-600 transition hover:bg-brand-100">
              عرض الكل
            </button>
            <h3 className="text-base font-bold text-slate-700">
              <span className="ml-2 inline-block h-2 w-2 rounded-full bg-amber-400"></span>
              نظام GPS
            </h3>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <MiniStatCard title="الأجهزة" value={gpsStats?.totalDevices || 0}
              icon={<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" /><circle cx="12" cy="10" r="3" /></svg>}
              iconBg="bg-emerald-50" />
            <MiniStatCard title="المشتركين" value={gpsStats?.totalCustomers || 0}
              icon={<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /></svg>}
              iconBg="bg-blue-50" />
            <MiniStatCard title="شرائح SIM" value={gpsStats?.totalSims || 0}
              icon={<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#8b5cf6" strokeWidth="2"><rect x="2" y="3" width="20" height="14" rx="2" /><path d="M8 21h8 M12 17v4" /></svg>}
              iconBg="bg-violet-50" />
            <MiniStatCard title="صيانة معلقة" value={pendingMaintenance}
              icon={<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" strokeWidth="2"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" /></svg>}
              iconBg="bg-amber-50" />
          </div>
        </div>

        {/* Sales & Finance Summary */}
        <div className="rounded-2xl bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="mb-4 flex items-center justify-between">
            <button onClick={() => navigate('/finance')} className="rounded-lg bg-brand-50 px-3 py-1.5 text-xs font-bold text-brand-600 transition hover:bg-brand-100">
              عرض الكل
            </button>
            <h3 className="text-base font-bold text-slate-700">
              <span className="ml-2 inline-block h-2 w-2 rounded-full bg-blue-400"></span>
              المبيعات والمالية
            </h3>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <MiniStatCard title="الحجوزات" value={bookingCount}
              icon={<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2"><rect x="3" y="4" width="18" height="18" rx="2" /><line x1="16" y1="2" x2="16" y2="6" /><line x1="8" y1="2" x2="8" y2="6" /><line x1="3" y1="10" x2="21" y2="10" /></svg>}
              iconBg="bg-blue-50" />
            <MiniStatCard title="العملاء" value={customerCount}
              icon={<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="8.5" cy="7" r="4" /></svg>}
              iconBg="bg-emerald-50" />
            <MiniStatCard title="عروض الأسعار" value={0}
              icon={<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#8b5cf6" strokeWidth="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><polyline points="14 2 14 8 20 8" /></svg>}
              iconBg="bg-violet-50" />
            <MiniStatCard title="المعاملات المالية" value={0}
              icon={<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" strokeWidth="2"><circle cx="12" cy="12" r="10" /><path d="M16 8h-6a2 2 0 1 0 0 4h4a2 2 0 0 1 0 4H8 M12 18V6" /></svg>}
              iconBg="bg-amber-50" />
          </div>
        </div>
      </div>
      )}

      {/* HR & Admin Summary - ADMIN ONLY */}
      {isAdmin && (
      <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-2">
        <div className="rounded-2xl bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="mb-4 flex items-center justify-between">
            <button onClick={() => navigate('/employees')} className="rounded-lg bg-brand-50 px-3 py-1.5 text-xs font-bold text-brand-600 transition hover:bg-brand-100">
              عرض الكل
            </button>
            <h3 className="text-base font-bold text-slate-700">
              <span className="ml-2 inline-block h-2 w-2 rounded-full bg-emerald-400"></span>
              الموارد البشرية
            </h3>
          </div>
          <div className="grid grid-cols-2 gap-3">
            <MiniStatCard title="الموظفين" value={employeeCount}
              icon={<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" /></svg>}
              iconBg="bg-blue-50" />
            <MiniStatCard title="الشكاوى" value={0}
              icon={<svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="#ef4444" strokeWidth="2"><path d="M10.29 3.86L1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z" /><line x1="12" y1="9" x2="12" y2="13" /><line x1="12" y1="17" x2="12.01" y2="17" /></svg>}
              iconBg="bg-red-50" />
          </div>
        </div>

        {/* Activity / Alerts */}
        <div className="rounded-2xl bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="mb-4 text-right text-base font-bold text-slate-700">
            <span className="ml-2 inline-block h-2 w-2 rounded-full bg-red-400"></span>
            آخر النشاطات
          </h3>
          <div className="flex h-32 items-center justify-center rounded-xl bg-slate-50 text-slate-300">
            <div className="text-center">
              <svg width="36" height="36" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="mx-auto mb-2">
                <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
              </svg>
              <p className="text-sm">لا توجد نشاطات حديثة</p>
            </div>
          </div>
        </div>
      </div>
      )}
    </div>
  )
}

function StatCard({ title, value, subtitle, icon, iconBg, loading, onClick }: {
  title: string; value: number; subtitle: string; icon: React.ReactNode; iconBg: string; loading: boolean; onClick?: () => void
}) {
  return (
    <div onClick={onClick} className={`flex items-center justify-between rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)] ${onClick ? 'cursor-pointer transition-all hover:shadow-lg hover:-translate-y-0.5' : ''}`}>
      <div className={`flex h-14 w-14 items-center justify-center rounded-xl ${iconBg}`}>{icon}</div>
      <div className="text-right">
        <p className="text-xs font-medium text-slate-400">{title}</p>
        <p className="mt-1 text-3xl font-extrabold text-slate-800">{loading ? '...' : value}</p>
        <p className="text-xs text-slate-400">{subtitle}</p>
      </div>
    </div>
  )
}

function MiniStatCard({ title, value, icon, iconBg }: {
  title: string; value: number; icon: React.ReactNode; iconBg: string
}) {
  return (
    <div className="flex items-center justify-between rounded-xl border border-slate-100 p-4">
      <div className={`flex h-11 w-11 items-center justify-center rounded-xl ${iconBg}`}>{icon}</div>
      <div className="text-right">
        <p className="text-xs text-slate-400">{title}</p>
        <p className="mt-0.5 text-2xl font-extrabold text-slate-800">{value}</p>
      </div>
    </div>
  )
}

function QuickCard({ title, desc, color, icon, onClick }: {
  title: string; desc: string; color: string; icon: React.ReactNode; onClick: () => void
}) {
  return (
    <button onClick={onClick} className={`flex items-center gap-4 rounded-2xl bg-gradient-to-l ${color} p-5 text-right text-white shadow-lg transition-all hover:shadow-xl hover:-translate-y-0.5`}>
      <div className="flex h-14 w-14 items-center justify-center rounded-xl bg-white/20">{icon}</div>
      <div className="flex-1">
        <p className="text-base font-bold">{title}</p>
        <p className="mt-1 text-xs text-white/80">{desc}</p>
      </div>
    </button>
  )
}
