import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api'
import { useSession } from '../session'

interface GpsStats {
  totalDevices: number
  totalCustomers: number
  devicesByStatus: { status: string; count: number }[]
  totalSims: number
  availableSims: number
  inUseSims: number
}

export default function Dashboard() {
  const { employee } = useSession()
  const navigate = useNavigate()
  const [gpsStats, setGpsStats] = useState<GpsStats | null>(null)
  const [bookingCount, setBookingCount] = useState(0)
  const [employeeCount, setEmployeeCount] = useState(0)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!employee) return
    Promise.all([
      api.getGpsStats().catch(() => null),
      api.getBookings().then((b) => b.length).catch(() => 0),
      api.getEmployees().then((e) => e.length).catch(() => 0),
    ]).then(([gps, bk, emp]) => {
      setGpsStats(gps as GpsStats | null)
      setBookingCount(bk)
      setEmployeeCount(emp)
    }).finally(() => setLoading(false))
  }, [employee])

  if (!employee) return null

  const activeDevices = gpsStats?.devicesByStatus?.find((d) => d.status === 'ACTIVE')?.count || 0
  const pendingMaintenance = gpsStats?.devicesByStatus?.find((d) => d.status === 'MAINTENANCE')?.count || 0

  return (
    <div dir="rtl">
      {/* Hero Banner */}
      <div className="relative overflow-hidden rounded-2xl bg-gradient-to-l from-[#1a2744] to-[#2a4a7f] p-8 text-white shadow-xl">
        <div className="relative z-10 max-w-lg">
          <h1 className="text-2xl font-bold">مرحباً بك في لوحة التحكم</h1>
          <p className="mt-2 text-sm text-blue-200">نظام شامل لإدارة جميع خدمات الشركة</p>
          <p className="mt-1 text-sm text-blue-200">تتبع وإدارة جميع الأجهزة والمشتركين والاشتراكات والصيانة بسهولة وكفاءة</p>
          <button
            onClick={() => navigate('/gps')}
            className="mt-5 rounded-lg bg-brand-500 px-6 py-2.5 text-sm font-bold text-white shadow-lg transition-all hover:bg-brand-400 hover:shadow-xl"
          >
            عرض التفاصيل
          </button>
        </div>
        {/* Decorative elements */}
        <div className="absolute left-8 top-1/2 -translate-y-1/2 opacity-20">
          <svg width="200" height="200" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="0.5">
            <path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />
            <circle cx="12" cy="10" r="3" />
          </svg>
        </div>
        <div className="absolute left-32 top-6 text-6xl font-extrabold opacity-10">GPS</div>
      </div>

      {/* GPS Summary Section */}
      <div className="mt-8">
        <h2 className="mb-4 text-right text-lg font-bold text-slate-700">
          <span className="ml-2 inline-block h-2 w-2 rounded-full bg-brand-500"></span>
          ملخص نظام GPS
        </h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <StatCard
            title="الأجهزة"
            value={gpsStats?.totalDevices || 0}
            subtitle="جهاز نشط"
            icon={<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" /><circle cx="12" cy="10" r="3" /></svg>}
            iconBg="bg-emerald-50"
            loading={loading}
          />
          <StatCard
            title="المشتركين"
            value={gpsStats?.totalCustomers || 0}
            subtitle="مشترك نشط"
            icon={<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2"><path d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="9" cy="7" r="4" /><path d="M23 21v-2a4 4 0 0 0-3-3.87" /><path d="M16 3.13a4 4 0 0 1 0 7.75" /></svg>}
            iconBg="bg-blue-50"
            loading={loading}
          />
          <StatCard
            title="الاشتراكات"
            value={gpsStats?.inUseSims || 0}
            subtitle="اشتراك فعال"
            icon={<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#8b5cf6" strokeWidth="2"><rect x="2" y="3" width="20" height="14" rx="2" /><path d="M8 21h8 M12 17v4" /></svg>}
            iconBg="bg-violet-50"
            loading={loading}
          />
          <StatCard
            title="الصيانة"
            value={pendingMaintenance}
            subtitle="أجهزة تحتاج صيانة"
            icon={<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" strokeWidth="2"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" /></svg>}
            iconBg="bg-amber-50"
            loading={loading}
          />
        </div>
      </div>

      {/* Middle Section: Chart + Today Stats */}
      <div className="mt-8 grid grid-cols-1 gap-6 lg:grid-cols-2">
        {/* Placeholder chart area */}
        <div className="rounded-2xl bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="mb-4 flex items-center justify-between">
            <div className="rounded-lg border border-slate-200 px-3 py-1.5 text-xs text-slate-500">اخر 7 أيام</div>
            <h3 className="text-base font-bold text-slate-700">نظرة عامة على الأجهزة</h3>
          </div>
          <div className="flex h-48 items-center justify-center rounded-xl bg-slate-50 text-slate-300">
            <div className="text-center">
              <svg width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.5" className="mx-auto mb-2">
                <polyline points="22 12 18 12 15 21 9 3 6 12 2 12" />
              </svg>
              <p className="text-sm">الرسم البياني</p>
            </div>
          </div>
        </div>

        {/* Today Stats */}
        <div className="rounded-2xl bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="mb-4 text-right text-base font-bold text-slate-700">إحصائيات اليوم</h3>
          <div className="grid grid-cols-2 gap-3">
            <MiniStatCard title="العملاء الجدد" value={gpsStats?.totalCustomers || 0}
              icon={<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="8.5" cy="7" r="4" /><line x1="20" y1="8" x2="20" y2="14" /><line x1="23" y1="11" x2="17" y2="11" /></svg>}
              iconBg="bg-emerald-50" />
            <MiniStatCard title="الاشتراكات الجديدة" value={gpsStats?.inUseSims || 0}
              icon={<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2"><rect x="2" y="3" width="20" height="14" rx="2" /><path d="M8 21h8 M12 17v4" /><line x1="17" y1="8" x2="17" y2="14" /><line x1="20" y1="11" x2="14" y2="11" /></svg>}
              iconBg="bg-blue-50" />
            <MiniStatCard title="أجهزة تحت الصيانة" value={pendingMaintenance}
              icon={<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#f59e0b" strokeWidth="2"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" /></svg>}
              iconBg="bg-amber-50" />
            <MiniStatCard title="تنبيهات اليوم" value={0}
              icon={<svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#ef4444" strokeWidth="2"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9" /><path d="M13.73 21a2 2 0 0 1-3.46 0" /></svg>}
              iconBg="bg-red-50" />
          </div>
        </div>
      </div>

      {/* Financial Summary */}
      <div className="mt-8">
        <h2 className="mb-4 text-right text-lg font-bold text-slate-700">
          <span className="ml-2 inline-block h-2 w-2 rounded-full bg-brand-500"></span>
          ملخص مالي سريع
        </h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          <FinanceCard title="إجمالي الإيرادات" value="---" subtitle="اليوم"
            icon={<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#3b82f6" strokeWidth="2"><circle cx="12" cy="12" r="10" /><path d="M16 8h-6a2 2 0 1 0 0 4h4a2 2 0 0 1 0 4H8 M12 18V6" /></svg>}
            iconBg="bg-blue-50" />
          <FinanceCard title="المصروفات" value="---" subtitle="اليوم"
            icon={<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#ef4444" strokeWidth="2"><circle cx="12" cy="12" r="10" /><line x1="15" y1="9" x2="9" y2="15" /><line x1="9" y1="9" x2="15" y2="15" /></svg>}
            iconBg="bg-red-50" />
          <FinanceCard title="صافي الأرباح" value="---" subtitle="اليوم"
            icon={<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#8b5cf6" strokeWidth="2"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18" /><polyline points="17 6 23 6 23 12" /></svg>}
            iconBg="bg-violet-50" />
          <FinanceCard title="المعاملات" value={String(bookingCount)} subtitle="اليوم"
            icon={<svg width="28" height="28" viewBox="0 0 24 24" fill="none" stroke="#10b981" strokeWidth="2"><rect x="1" y="4" width="22" height="16" rx="2" /><line x1="1" y1="10" x2="23" y2="10" /></svg>}
            iconBg="bg-emerald-50" />
        </div>
      </div>
    </div>
  )
}

function StatCard({ title, value, subtitle, icon, iconBg, loading }: {
  title: string; value: number; subtitle: string; icon: React.ReactNode; iconBg: string; loading: boolean
}) {
  return (
    <div className="flex items-center justify-between rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
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

function FinanceCard({ title, value, subtitle, icon, iconBg }: {
  title: string; value: string; subtitle: string; icon: React.ReactNode; iconBg: string
}) {
  return (
    <div className="flex items-center justify-between rounded-2xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
      <div className={`flex h-14 w-14 items-center justify-center rounded-xl ${iconBg}`}>{icon}</div>
      <div className="text-right">
        <p className="text-xs font-medium text-slate-400">{title}</p>
        <p className="mt-1 text-2xl font-extrabold text-slate-800">{value}</p>
        <p className="text-xs text-slate-400">{subtitle}</p>
      </div>
    </div>
  )
}
