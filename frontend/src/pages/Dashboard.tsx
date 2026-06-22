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

  if (!employee) return null

  const isAdmin = employee.role === 'ADMIN'
  const pendingMaintenance = gpsStats?.devicesByStatus?.find((d) => d.status === 'MAINTENANCE')?.count || 0

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

      {/* Quick Access Cards */}
      <div className="mt-8">
        <h2 className="mb-4 text-right text-lg font-bold text-slate-700">
          <span className="ml-2 inline-block h-2 w-2 rounded-full bg-brand-500"></span>
          وصول سريع
        </h2>
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
          <QuickCard
            title="حجز جديد"
            desc="إنشاء حجز خدمة جديد للعميل"
            color="from-blue-500 to-blue-600"
            icon={<svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2"><circle cx="12" cy="12" r="10" /><line x1="12" y1="8" x2="12" y2="16" /><line x1="8" y1="12" x2="16" y2="12" /></svg>}
            onClick={() => navigate('/sales')}
          />
          <QuickCard
            title="عرض سعر جديد"
            desc="إنشاء عرض سعر احترافي للعميل"
            color="from-emerald-500 to-emerald-600"
            icon={<svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z" /><polyline points="14 2 14 8 20 8" /><line x1="16" y1="13" x2="8" y2="13" /><line x1="16" y1="17" x2="8" y2="17" /></svg>}
            onClick={() => navigate('/quotations/new')}
          />
          <QuickCard
            title="إضافة عميل"
            desc="تسجيل عميل جديد في النظام"
            color="from-violet-500 to-violet-600"
            icon={<svg width="32" height="32" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2"><path d="M16 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2" /><circle cx="8.5" cy="7" r="4" /><line x1="20" y1="8" x2="20" y2="14" /><line x1="23" y1="11" x2="17" y2="11" /></svg>}
            onClick={() => navigate('/customers')}
          />
        </div>
      </div>

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
