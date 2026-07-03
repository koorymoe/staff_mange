import { useEffect, useState } from 'react'
import type { Employee } from '../api'
import { useSession, roleLabels } from '../session'

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:4000/api'

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = localStorage.getItem('authToken')
  const res = await fetch(`${API_URL}${path}`, {
    ...options,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(options?.headers || {}),
    },
  })
  if (!res.ok) {
    const body = await res.json().catch(() => ({}))
    throw new Error(body.error || `Request failed: ${res.status}`)
  }
  return res.json()
}

interface StatsData {
  totals: {
    totalCustomers: number
    totalBookings: number
    pendingBookings: number
    confirmedBookings: number
    completedBookings: number
    cancelledBookings: number
    urgentPending: number
    totalRevenue: number
    unverifiedRevenue: number
  }
  salesStats: { employeeId: string; name: string; totalTransferred: number; confirmed: number; today: number; thisMonth: number }[]
  coordinatorStats: { employeeId: string; name: string; totalConfirmed: number; today: number; thisMonth: number }[]
  technicianStats: { employeeId: string; name: string; onDuty: boolean; totalAssigned: number; completed: number; revenueHandled: number }[]
  serviceBreakdown: { serviceId: string; name: string; count: number }[]
  roleCounts: { role: string; count: number }[]
}

interface InventoryItem {
  id: string
  name: string
  quantity: number
  employee?: { name: string }
}

export default function StatsPage() {
  const { employee } = useSession()
  const role = employee?.role
  const [stats, setStats] = useState<StatsData | null>(null)
  const [employees, setEmployees] = useState<Employee[]>([])
  const [inventory, setInventory] = useState<InventoryItem[]>([])
  const [tab, setTab] = useState<'overview' | 'sales' | 'technicians' | 'coordinators'>('overview')

  useEffect(() => {
    if (role !== 'ADMIN') return
    request<StatsData>('/stats').then(setStats).catch(() => {})
    request<Employee[]>('/employees').then(setEmployees).catch(() => {})
    request<InventoryItem[]>('/inventory/personal').then(setInventory).catch(() => {})
  }, [role])

  if (role !== 'ADMIN') return <div className="p-8 text-center text-gray-500">غير مصرح</div>
  if (!stats) return <div className="flex items-center justify-center h-64"><div className="w-8 h-8 border-4 border-[var(--color-brand-500)] border-t-transparent rounded-full animate-spin" /></div>

  const totalEmp = employees.length
  const activeEmp = employees.filter(e => e.onDuty).length

  const tabs = [
    { key: 'overview' as const, label: 'نظرة عامة' },
    { key: 'sales' as const, label: 'المبيعات' },
    { key: 'technicians' as const, label: 'الفنيون' },
    { key: 'coordinators' as const, label: 'المنسقون' },
  ]

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-[var(--color-brand-900)]">إحصائيات الموظفين</h1>

      {/* Tabs */}
      <div className="flex gap-2 border-b border-gray-200 pb-2">
        {tabs.map(t => (
          <button key={t.key} onClick={() => setTab(t.key)}
            className={`px-4 py-2 rounded-t-lg text-sm font-medium transition-colors ${tab === t.key ? 'bg-[var(--color-brand-500)] text-white' : 'text-gray-600 hover:bg-gray-100'}`}>
            {t.label}
          </button>
        ))}
      </div>

      {tab === 'overview' && (
        <div className="space-y-6">
          {/* Employee counts */}
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
            <Card title="إجمالي الموظفين" value={totalEmp} color="blue" />
            <Card title="في الخدمة" value={activeEmp} color="green" />
            <Card title="خارج الخدمة" value={totalEmp - activeEmp} color="red" />
            <Card title="إجمالي العملاء" value={stats.totals.totalCustomers} color="purple" />
          </div>

          {/* Role distribution */}
          <div className="bg-white rounded-xl shadow-sm border p-6">
            <h2 className="text-lg font-semibold mb-4">توزيع الأدوار</h2>
            <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-3">
              {stats.roleCounts.map(r => (
                <div key={r.role} className="flex items-center justify-between bg-gray-50 rounded-lg p-3">
                  <span className="text-sm text-gray-700">{roleLabels[r.role as keyof typeof roleLabels] || r.role}</span>
                  <span className="text-lg font-bold text-[var(--color-brand-500)]">{r.count}</span>
                </div>
              ))}
            </div>
          </div>

          {/* Booking stats */}
          <div className="bg-white rounded-xl shadow-sm border p-6">
            <h2 className="text-lg font-semibold mb-4">إحصائيات الحجوزات</h2>
            <div className="grid grid-cols-2 sm:grid-cols-4 gap-4">
              <MiniCard label="إجمالي" value={stats.totals.totalBookings} />
              <MiniCard label="قيد الانتظار" value={stats.totals.pendingBookings} className="text-yellow-600" />
              <MiniCard label="مؤكدة" value={stats.totals.confirmedBookings} className="text-blue-600" />
              <MiniCard label="مكتملة" value={stats.totals.completedBookings} className="text-green-600" />
              <MiniCard label="ملغاة" value={stats.totals.cancelledBookings} className="text-red-600" />
              <MiniCard label="عاجلة معلقة" value={stats.totals.urgentPending} className="text-orange-600" />
              <MiniCard label="إجمالي الإيرادات" value={`${(stats.totals.totalRevenue / 1000).toFixed(0)}K`} className="text-green-700" />
              <MiniCard label="غير مدققة" value={`${(stats.totals.unverifiedRevenue / 1000).toFixed(0)}K`} className="text-amber-600" />
            </div>
          </div>

          {/* Services breakdown */}
          <div className="bg-white rounded-xl shadow-sm border p-6">
            <h2 className="text-lg font-semibold mb-4">الخدمات الأكثر طلباً</h2>
            <div className="space-y-2">
              {stats.serviceBreakdown.slice(0, 8).map((s, i) => {
                const max = stats.serviceBreakdown[0]?.count || 1
                return (
                  <div key={s.serviceId || i} className="flex items-center gap-3">
                    <span className="text-sm w-32 truncate text-gray-700">{s.name}</span>
                    <div className="flex-1 h-6 bg-gray-100 rounded-full overflow-hidden">
                      <div className="h-full bg-[var(--color-brand-500)] rounded-full transition-all" style={{ width: `${(s.count / max) * 100}%` }} />
                    </div>
                    <span className="text-sm font-medium w-10 text-left">{s.count}</span>
                  </div>
                )
              })}
            </div>
          </div>

          {/* Inventory summary */}
          {inventory.length > 0 && (
            <div className="bg-white rounded-xl shadow-sm border p-6">
              <h2 className="text-lg font-semibold mb-4">ملخص الجرد</h2>
              <p className="text-gray-600 text-sm mb-2">إجمالي الأدوات المسجلة: <span className="font-bold">{inventory.length}</span></p>
              <p className="text-gray-600 text-sm">إجمالي الكمية: <span className="font-bold">{inventory.reduce((s, i) => s + i.quantity, 0)}</span></p>
            </div>
          )}
        </div>
      )}

      {tab === 'sales' && (
        <div className="bg-white rounded-xl shadow-sm border overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="text-right p-3 font-medium">#</th>
                <th className="text-right p-3 font-medium">الموظف</th>
                <th className="text-right p-3 font-medium">إجمالي التحويلات</th>
                <th className="text-right p-3 font-medium">مؤكدة</th>
                <th className="text-right p-3 font-medium">اليوم</th>
                <th className="text-right p-3 font-medium">هذا الشهر</th>
                <th className="text-right p-3 font-medium">نسبة النجاح</th>
              </tr>
            </thead>
            <tbody>
              {stats.salesStats.map((s, i) => (
                <tr key={s.employeeId} className="border-t hover:bg-gray-50">
                  <td className="p-3 font-bold text-[var(--color-brand-500)]">{i + 1}</td>
                  <td className="p-3 font-medium">{s.name}</td>
                  <td className="p-3">{s.totalTransferred}</td>
                  <td className="p-3 text-green-600 font-medium">{s.confirmed}</td>
                  <td className="p-3">{s.today}</td>
                  <td className="p-3">{s.thisMonth}</td>
                  <td className="p-3">
                    <span className={`font-medium ${s.totalTransferred ? (s.confirmed / s.totalTransferred > 0.5 ? 'text-green-600' : 'text-amber-600') : 'text-gray-400'}`}>
                      {s.totalTransferred ? `${Math.round((s.confirmed / s.totalTransferred) * 100)}%` : '-'}
                    </span>
                  </td>
                </tr>
              ))}
              {stats.salesStats.length === 0 && <tr><td colSpan={7} className="p-8 text-center text-gray-400">لا توجد بيانات</td></tr>}
            </tbody>
          </table>
        </div>
      )}

      {tab === 'technicians' && (
        <div className="bg-white rounded-xl shadow-sm border overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="text-right p-3 font-medium">#</th>
                <th className="text-right p-3 font-medium">الفني</th>
                <th className="text-right p-3 font-medium">الحالة</th>
                <th className="text-right p-3 font-medium">المهام</th>
                <th className="text-right p-3 font-medium">مكتملة</th>
                <th className="text-right p-3 font-medium">الإيرادات</th>
                <th className="text-right p-3 font-medium">نسبة الإنجاز</th>
              </tr>
            </thead>
            <tbody>
              {stats.technicianStats.map((t, i) => (
                <tr key={t.employeeId} className="border-t hover:bg-gray-50">
                  <td className="p-3 font-bold text-[var(--color-brand-500)]">{i + 1}</td>
                  <td className="p-3 font-medium">{t.name}</td>
                  <td className="p-3">
                    <span className={`inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-xs ${t.onDuty ? 'bg-green-100 text-green-700' : 'bg-red-100 text-red-700'}`}>
                      <span className={`w-1.5 h-1.5 rounded-full ${t.onDuty ? 'bg-green-500' : 'bg-red-500'}`} />
                      {t.onDuty ? 'متاح' : 'غير متاح'}
                    </span>
                  </td>
                  <td className="p-3">{t.totalAssigned}</td>
                  <td className="p-3 text-green-600 font-medium">{t.completed}</td>
                  <td className="p-3">{t.revenueHandled.toLocaleString()} د.ع</td>
                  <td className="p-3">
                    <span className={`font-medium ${t.totalAssigned ? (t.completed / t.totalAssigned > 0.5 ? 'text-green-600' : 'text-amber-600') : 'text-gray-400'}`}>
                      {t.totalAssigned ? `${Math.round((t.completed / t.totalAssigned) * 100)}%` : '-'}
                    </span>
                  </td>
                </tr>
              ))}
              {stats.technicianStats.length === 0 && <tr><td colSpan={7} className="p-8 text-center text-gray-400">لا توجد بيانات</td></tr>}
            </tbody>
          </table>
        </div>
      )}

      {tab === 'coordinators' && (
        <div className="bg-white rounded-xl shadow-sm border overflow-hidden">
          <table className="w-full text-sm">
            <thead className="bg-gray-50">
              <tr>
                <th className="text-right p-3 font-medium">#</th>
                <th className="text-right p-3 font-medium">المنسق</th>
                <th className="text-right p-3 font-medium">إجمالي التأكيدات</th>
                <th className="text-right p-3 font-medium">اليوم</th>
                <th className="text-right p-3 font-medium">هذا الشهر</th>
              </tr>
            </thead>
            <tbody>
              {stats.coordinatorStats.map((c, i) => (
                <tr key={c.employeeId} className="border-t hover:bg-gray-50">
                  <td className="p-3 font-bold text-[var(--color-brand-500)]">{i + 1}</td>
                  <td className="p-3 font-medium">{c.name}</td>
                  <td className="p-3 text-green-600 font-medium">{c.totalConfirmed}</td>
                  <td className="p-3">{c.today}</td>
                  <td className="p-3">{c.thisMonth}</td>
                </tr>
              ))}
              {stats.coordinatorStats.length === 0 && <tr><td colSpan={5} className="p-8 text-center text-gray-400">لا توجد بيانات</td></tr>}
            </tbody>
          </table>
        </div>
      )}
    </div>
  )
}

function Card({ title, value, color }: { title: string; value: number | string; color: string }) {
  const colors: Record<string, string> = {
    blue: 'bg-blue-50 text-blue-700 border-blue-200',
    green: 'bg-green-50 text-green-700 border-green-200',
    red: 'bg-red-50 text-red-700 border-red-200',
    purple: 'bg-purple-50 text-purple-700 border-purple-200',
  }
  return (
    <div className={`rounded-xl border p-4 ${colors[color] || colors.blue}`}>
      <p className="text-xs opacity-70 mb-1">{title}</p>
      <p className="text-2xl font-bold">{value}</p>
    </div>
  )
}

function MiniCard({ label, value, className = '' }: { label: string; value: number | string; className?: string }) {
  return (
    <div className="bg-gray-50 rounded-lg p-3 text-center">
      <p className="text-xs text-gray-500 mb-1">{label}</p>
      <p className={`text-xl font-bold ${className}`}>{value}</p>
    </div>
  )
}
