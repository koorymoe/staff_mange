import { useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { type Employee, type EmployeeRole } from '../api'
import { SessionContext, roleLabels } from '../session'
import Login from '../pages/Login'

interface NavItem {
  to: string
  label: string
  end?: boolean
  roles?: EmployeeRole[] // undefined = visible to everyone
}

const navItems: NavItem[] = [
  { to: '/', label: 'الرئيسية', end: true, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'FINANCE', 'PROJECT_MANAGER'] },
  { to: '/sales', label: 'حجز جديد', roles: ['ADMIN', 'SALES'] },
  { to: '/coordinator', label: 'تنسيق الحجوزات', roles: ['ADMIN', 'HR_COORDINATOR'] },
  { to: '/my-tasks', label: 'مهامي', roles: ['ADMIN', 'TECHNICIAN', 'PROJECT_MANAGER'] },
  { to: '/my-ranking', label: 'تصنيفي', roles: ['TECHNICIAN'] },
  { to: '/my-expenses', label: 'مصاريفي', roles: ['ADMIN', 'TECHNICIAN', 'PROJECT_MANAGER'] },
  { to: '/employees', label: 'إدارة الكوادر', roles: ['ADMIN', 'HR_COORDINATOR'] },
  { to: '/customers', label: 'الزبائن', roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'] },
  { to: '/bookings', label: 'الحجوزات', roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'FINANCE'] },
  { to: '/services', label: 'الخدمات', roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'] },
  { to: '/finance', label: 'تدقيق الحسابات', roles: ['ADMIN', 'FINANCE'] },
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

  const setEmployee = (emp: Employee | null) => {
    setEmployeeState(emp)
    if (emp) localStorage.setItem('currentEmployee', JSON.stringify(emp))
    else localStorage.removeItem('currentEmployee')
  }

  if (!employee) {
    return (
      <SessionContext.Provider value={{ employee, setEmployee }}>
        <Login />
      </SessionContext.Provider>
    )
  }

  const role = employee?.role
  const visibleItems = navItems.filter((item) => !item.roles || (role && item.roles.includes(role)))

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

          <nav className="flex flex-col gap-1 p-4">
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
