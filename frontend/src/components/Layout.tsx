import { useState, useEffect } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { api, type Employee, type EmployeeRole } from '../api'
import { SessionContext, roleLabels } from '../session'
import Login from '../pages/Login'

interface NavItem {
  to: string
  label: string
  end?: boolean
  roles?: EmployeeRole[] // undefined = visible to everyone
  permission?: string
}

const navItems: NavItem[] = [
  { to: '/', label: 'الرئيسية', end: true, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'FINANCE', 'PROJECT_MANAGER'] },
  { to: '/sales', label: 'حجز جديد', roles: ['ADMIN', 'SALES'] },
  { to: '/coordinator', label: 'تنسيق الحجوزات', roles: ['ADMIN', 'HR_COORDINATOR'] },
  { to: '/my-tasks', label: 'مهامي', roles: ['TECHNICIAN', 'PROJECT_MANAGER'] },
  { to: '/my-ranking', label: 'تصنيفي', roles: ['TECHNICIAN'] },
  { to: '/my-expenses', label: 'مصاريفي', roles: ['ADMIN', 'TECHNICIAN', 'PROJECT_MANAGER'] },
  { to: '/my-inventory', label: 'جرد أدواتي', roles: ['TECHNICIAN'] },
  { to: '/employees', label: 'إدارة الكوادر', roles: ['ADMIN', 'HR_COORDINATOR'] },
  { to: '/customers', label: 'الزبائن', roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'] },
  { to: '/bookings', label: 'الحجوزات', roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'FINANCE'] },
  { to: '/services', label: 'الخدمات', roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'] },
  { to: '/finance', label: 'تدقيق الحسابات', roles: ['ADMIN', 'FINANCE'] },
  { to: '/kpi', label: 'تقييم الأداء', roles: ['ADMIN', 'MONITOR'], permission: 'kpi_management' },
  { to: '/complaints', label: 'الشكاوى', roles: ['ADMIN', 'SALES', 'HR_COORDINATOR'], permission: 'complaints' },
  { to: '/inventory', label: 'جرد الأدوات', roles: ['ADMIN', 'HR_COORDINATOR'], permission: 'inventory' },
  { to: '/permissions', label: 'الصلاحيات', roles: ['ADMIN'] },
  { to: '/quotations', label: 'عروض الأسعار', roles: ['ADMIN', 'SALES'], permission: 'quotation_system' },
  { to: '/products', label: 'المنتجات', roles: ['ADMIN'], permission: 'quotation_system' },
  { to: '/gps', label: 'نظام GPS', roles: ['ADMIN', 'GPS_ADMIN'], permission: 'gps_system' },
  { to: '/gps/customers', label: 'زبائن GPS', roles: ['ADMIN', 'GPS_ADMIN'], permission: 'gps_system' },
  { to: '/gps/devices', label: 'أجهزة GPS', roles: ['ADMIN', 'GPS_ADMIN'], permission: 'gps_system' },
  { to: '/gps/sims', label: 'شرائح SIM', roles: ['ADMIN', 'GPS_ADMIN'], permission: 'gps_system' },
  { to: '/gps/renewals', label: 'تجديد الاشتراكات', roles: ['ADMIN', 'GPS_ADMIN'], permission: 'gps_system' },
  { to: '/gps/maintenance', label: 'صيانة GPS', roles: ['ADMIN', 'GPS_ADMIN'], permission: 'gps_system' },
  { to: '/gps/employee', label: 'لوحة موظف GPS', roles: ['GPS_ENGINEER'] },
]

const loadStoredEmployee = (): Employee | null => {
  const raw = localStorage.getItem('currentEmployee')
  if (!raw) return null
  try {
    return JSON.parse(raw) as Employee
  } catch {
    return null
  }
}

export default function Layout() {
  const [employee, setEmployeeState] = useState<Employee | null>(loadStoredEmployee)
  const [employeePermissions, setEmployeePermissions] = useState<string[]>([])

  const setEmployee = (emp: Employee | null) => {
    setEmployeeState(emp)
    if (emp) localStorage.setItem('currentEmployee', JSON.stringify(emp))
    else localStorage.removeItem('currentEmployee')
  }

  useEffect(() => {
    if (!employee) {
      setEmployeePermissions([])
      return
    }
    api.getEmployeePermissions(employee.id)
      .then((perms) => setEmployeePermissions(perms.map((p: any) => p.name)))
      .catch(() => setEmployeePermissions([]))
  }, [employee?.id])

  if (!employee) {
    return (
      <SessionContext.Provider value={{ employee, setEmployee }}>
        <Login />
      </SessionContext.Provider>
    )
  }

  const role = employee?.role
  const visibleItems = navItems.filter((item) => {
    if (role === 'ADMIN') return true
    if (item.roles && role && !item.roles.includes(role)) return false
    if (item.permission && !employeePermissions.includes(item.permission)) return false
    return true
  })

  return (
    <SessionContext.Provider value={{ employee, setEmployee }}>
      <div className="flex min-h-screen bg-slate-100">
        <aside className="flex w-64 flex-col bg-gradient-to-b from-brand-500 via-brand-700 to-brand-900 text-white shadow-xl">
          <div className="border-b border-white/10 px-6 py-6">
            <h1 className="text-xl font-extrabold text-white">شركة الأماني</h1>
            <p className="mt-1 text-sm text-brand-200">نظام الإدارة المتكامل</p>
          </div>

          <div className="border-b border-white/10 px-4 py-4">
            <p className="text-sm font-bold text-white">{employee.name}</p>
            <p className="text-xs text-brand-200">{roleLabels[employee.role]}</p>
            <button
              onClick={() => setEmployee(null)}
              className="mt-2 w-full rounded-lg border border-white/20 bg-white/10 px-2 py-1.5 text-xs font-medium text-white transition-colors hover:bg-white/20"
            >
              تسجيل الخروج
            </button>
          </div>

          <nav className="flex flex-col gap-1 p-4 overflow-y-auto flex-1">
            {visibleItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                end={item.end}
                className={({ isActive }) =>
                  `rounded-lg px-4 py-3 text-sm font-medium transition-all ${
                    isActive
                      ? 'bg-white text-brand-900 shadow-lg shadow-black/10'
                      : 'text-brand-100 hover:bg-white/10'
                  }`
                }
              >
                {item.label}
              </NavLink>
            ))}
          </nav>
        </aside>
        <main className="flex-1 p-8">
          <Outlet />
        </main>
      </div>
    </SessionContext.Provider>
  )
}
