import { useEffect, useState } from 'react'
import { NavLink, Outlet } from 'react-router-dom'
import { api, type Employee, type EmployeeRole } from '../api'
import { SessionContext, roleLabels } from '../session'

interface NavItem {
  to: string
  label: string
  end?: boolean
  roles?: EmployeeRole[] // undefined = visible to everyone
}

const navItems: NavItem[] = [
  { to: '/', label: 'الرئيسية', end: true },
  { to: '/sales', label: 'حجز جديد', roles: ['ADMIN', 'SALES'] },
  { to: '/coordinator', label: 'تنسيق الحجوزات', roles: ['ADMIN', 'HR_COORDINATOR'] },
  { to: '/my-tasks', label: 'مهامي', roles: ['ADMIN', 'TECHNICIAN', 'PROJECT_MANAGER'] },
  { to: '/employees', label: 'إدارة الكوادر', roles: ['ADMIN', 'HR_COORDINATOR'] },
  { to: '/customers', label: 'الزبائن', roles: ['ADMIN', 'SALES', 'HR_COORDINATOR'] },
  { to: '/services', label: 'الخدمات' },
]

export default function Layout() {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [employee, setEmployeeState] = useState<Employee | null>(null)

  useEffect(() => {
    api.getEmployees().then((emps) => {
      setEmployees(emps)
      const savedId = localStorage.getItem('currentEmployeeId')
      const found = emps.find((e) => e.id === savedId)
      setEmployeeState(found || null)
    })
  }, [])

  const setEmployee = (emp: Employee | null) => {
    setEmployeeState(emp)
    if (emp) localStorage.setItem('currentEmployeeId', emp.id)
    else localStorage.removeItem('currentEmployeeId')
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
            <label className="mb-1 block text-xs font-medium text-brand-200">أنا الآن:</label>
            <select
              value={employee?.id || ''}
              onChange={(e) => setEmployee(employees.find((emp) => emp.id === e.target.value) || null)}
              className="w-full rounded-lg border border-white/20 bg-white/10 px-2 py-2 text-sm text-white outline-none [&>option]:text-brand-900"
            >
              <option value="">-- اختر اسمك --</option>
              {employees.map((emp) => (
                <option key={emp.id} value={emp.id}>
                  {emp.name} ({roleLabels[emp.role]})
                </option>
              ))}
            </select>
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
