import { useState, useEffect } from 'react'
import { NavLink, Outlet, useLocation } from 'react-router-dom'
import { api, type Employee, type EmployeeRole } from '../api'
import { SessionContext, roleLabels } from '../session'
import Login from '../pages/Login'

interface NavItem {
  to: string
  label: string
  icon: React.ReactNode
  end?: boolean
  roles?: EmployeeRole[]
  permission?: string
  children?: NavItem[]
}

const SvgIcon = ({ d, color = 'currentColor' }: { d: string; color?: string }) => (
  <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke={color} strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
    <path d={d} />
  </svg>
)

const navItems: NavItem[] = [
  { to: '/', label: 'الرئيسية', end: true, icon: <SvgIcon d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z M9 22V12h6v10" />, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'FINANCE', 'PROJECT_MANAGER'] },
  { to: '/sales', label: 'حجز جديد', icon: <SvgIcon d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />, roles: ['ADMIN', 'SALES'] },
  { to: '/customers', label: 'العملاء', icon: <SvgIcon d="M17 21v-2a4 4 0 0 0-4-4H5a4 4 0 0 0-4 4v2 M23 21v-2a4 4 0 0 0-3-3.87 M16 3.13a4 4 0 0 1 0 7.75 M9 7a4 4 0 1 0 0-0.01" />, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'] },
  { to: '/bookings', label: 'الحجوزات', icon: <SvgIcon d="M19 4H5a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6a2 2 0 0 0-2-2zM16 2v4M8 2v4M3 10h18" />, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'FINANCE'] },
  {
    to: '/gps', label: 'GPS', icon: <SvgIcon d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z M12 7a3 3 0 1 0 0 6 3 3 0 0 0 0-6z" />,
    roles: ['ADMIN', 'GPS_ADMIN'],
    permission: 'gps_system',
    children: [
      { to: '/gps/devices', label: 'الأجهزة', icon: <></>, roles: ['ADMIN', 'GPS_ADMIN'], permission: 'gps_system' },
      { to: '/gps/customers', label: 'الزبائن', icon: <></>, roles: ['ADMIN', 'GPS_ADMIN'], permission: 'gps_system' },
      { to: '/gps/renewals', label: 'الاشتراكات', icon: <></>, roles: ['ADMIN', 'GPS_ADMIN'], permission: 'gps_system' },
      { to: '/gps/maintenance', label: 'الصيانة', icon: <></>, roles: ['ADMIN', 'GPS_ADMIN'], permission: 'gps_system' },
    ],
  },
  {
    to: '/finance', label: 'المالية', icon: <SvgIcon d="M2 17l10-10 4 4L22 5 M22 5v6h-6" />,
    roles: ['ADMIN', 'FINANCE'],
    children: [
      { to: '/finance', label: 'تدقيق الحسابات', icon: <></>, roles: ['ADMIN', 'FINANCE'], end: true },
      { to: '/my-expenses', label: 'المصاريف', icon: <></>, roles: ['ADMIN', 'TECHNICIAN', 'PROJECT_MANAGER'] },
    ],
  },
  {
    to: '/admin', label: 'الإدارة', icon: <SvgIcon d="M12 15a3 3 0 1 0 0-6 3 3 0 0 0 0 6z M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83" />,
    roles: ['ADMIN', 'HR_COORDINATOR'],
    children: [
      { to: '/employees', label: 'إدارة الكوادر', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR'] },
      { to: '/permissions', label: 'الصلاحيات', icon: <></>, roles: ['ADMIN'] },
      { to: '/services', label: 'الخدمات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'] },
      { to: '/kpi', label: 'تقييم الأداء', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'kpi_management' },
      { to: '/inventory', label: 'جرد الأدوات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR'], permission: 'inventory' },
      { to: '/complaints', label: 'الشكاوى', icon: <></>, roles: ['ADMIN', 'SALES', 'HR_COORDINATOR'], permission: 'complaints' },
      { to: '/quotations', label: 'عروض الأسعار', icon: <></>, roles: ['ADMIN', 'SALES'], permission: 'quotation_system' },
      { to: '/products', label: 'المنتجات', icon: <></>, roles: ['ADMIN'], permission: 'quotation_system' },
    ],
  },
  { to: '/coordinator', label: 'تنسيق الحجوزات', icon: <SvgIcon d="M16 4h2a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2h2 M15 2H9a1 1 0 0 0-1 1v2a1 1 0 0 0 1 1h6a1 1 0 0 0 1-1V3a1 1 0 0 0-1-1z" />, roles: ['ADMIN', 'HR_COORDINATOR'] },
  { to: '/my-tasks', label: 'مهامي', icon: <SvgIcon d="M9 11l3 3L22 4 M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11" />, roles: ['TECHNICIAN', 'PROJECT_MANAGER'] },
  { to: '/my-ranking', label: 'تصنيفي', icon: <SvgIcon d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z" />, roles: ['TECHNICIAN'] },
  { to: '/my-expenses', label: 'مصاريفي', icon: <SvgIcon d="M12 1v22M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />, roles: ['TECHNICIAN', 'PROJECT_MANAGER'] },
  { to: '/my-inventory', label: 'جرد أدواتي', icon: <SvgIcon d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z" />, roles: ['TECHNICIAN'] },
  { to: '/gps/employee', label: 'لوحة GPS', icon: <SvgIcon d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z" />, roles: ['GPS_ENGINEER'] },
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
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({})
  const location = useLocation()

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

  // Auto-expand group containing current route
  useEffect(() => {
    navItems.forEach((item) => {
      if (item.children) {
        const isChildActive = item.children.some((c) => location.pathname.startsWith(c.to))
        const isParentActive = location.pathname === item.to || location.pathname.startsWith(item.to + '/')
        if (isChildActive || isParentActive) {
          setExpandedGroups((prev) => ({ ...prev, [item.label]: true }))
        }
      }
    })
  }, [location.pathname])

  if (!employee) {
    return (
      <SessionContext.Provider value={{ employee, setEmployee }}>
        <Login />
      </SessionContext.Provider>
    )
  }

  const role = employee?.role

  const isItemVisible = (item: NavItem): boolean => {
    if (role === 'ADMIN') return true
    if (item.roles && role && !item.roles.includes(role)) return false
    if (item.permission && !employeePermissions.includes(item.permission)) return false
    return true
  }

  const visibleItems = navItems.filter((item) => {
    if (isItemVisible(item)) return true
    if (item.children && item.children.some(isItemVisible)) return true
    return false
  })

  const toggleGroup = (label: string) => {
    setExpandedGroups((prev) => ({ ...prev, [label]: !prev[label] }))
  }

  return (
    <SessionContext.Provider value={{ employee, setEmployee }}>
      <div className="flex min-h-screen flex-col bg-slate-50">
        {/* Top Header Bar */}
        <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-slate-200 bg-white px-6 shadow-sm">
          <div className="flex items-center gap-3">
            {/* Notification bell */}
            <button className="relative rounded-full p-2 text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600">
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9 M13.73 21a2 2 0 0 1-3.46 0" />
              </svg>
            </button>
            {/* User info */}
            <div className="flex items-center gap-3">
              <div className="text-right">
                <div className="text-sm font-bold text-slate-800">{employee.name}</div>
                <div className="text-xs text-slate-400">{roleLabels[employee.role]}</div>
              </div>
              <div className="flex h-9 w-9 items-center justify-center rounded-full bg-gradient-to-br from-brand-500 to-brand-700 text-sm font-bold text-white">
                {employee.name.charAt(0)}
              </div>
            </div>
          </div>
          {/* Company logo */}
          <div className="flex items-center gap-3">
            <span className="text-lg font-extrabold text-brand-900">نظام شركة الأماني</span>
            <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-brand-50">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="#1e3a5f" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z" />
                <polyline points="3.27 6.96 12 12.01 20.73 6.96" />
                <line x1="12" y1="22.08" x2="12" y2="12" />
              </svg>
            </div>
          </div>
        </header>

        <div className="flex flex-1">
          {/* Right Sidebar */}
          <aside className="sticky top-16 flex h-[calc(100vh-4rem)] w-64 flex-col overflow-y-auto border-l border-slate-200 bg-[#1e2a3a] text-white">
            <nav className="flex flex-1 flex-col gap-0.5 p-3 pt-4">
              {visibleItems.map((item) => {
                if (item.children) {
                  const visibleChildren = item.children.filter(isItemVisible)
                  if (visibleChildren.length === 0) return null
                  const isExpanded = expandedGroups[item.label]
                  const isGroupActive = visibleChildren.some((c) => location.pathname.startsWith(c.to)) || location.pathname === item.to
                  return (
                    <div key={item.label}>
                      <button
                        onClick={() => toggleGroup(item.label)}
                        className={`flex w-full items-center justify-between rounded-xl px-4 py-3 text-sm font-medium transition-all ${
                          isGroupActive ? 'bg-white/10 text-white' : 'text-slate-300 hover:bg-white/5 hover:text-white'
                        }`}
                      >
                        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
                          className={`transition-transform ${isExpanded ? 'rotate-180' : ''}`}>
                          <polyline points="6 9 12 15 18 9" />
                        </svg>
                        <div className="flex items-center gap-3">
                          <span>{item.label}</span>
                          <span className="opacity-80">{item.icon}</span>
                        </div>
                      </button>
                      {isExpanded && (
                        <div className="mr-6 mt-1 flex flex-col gap-0.5 border-r border-white/10 pr-3">
                          {visibleChildren.map((child) => (
                            <NavLink
                              key={child.to}
                              to={child.to}
                              end={child.end}
                              className={({ isActive }) =>
                                `rounded-lg px-4 py-2 text-right text-sm font-medium transition-all ${
                                  isActive
                                    ? 'bg-brand-500/20 text-brand-200'
                                    : 'text-slate-400 hover:bg-white/5 hover:text-white'
                                }`
                              }
                            >
                              {child.label}
                            </NavLink>
                          ))}
                        </div>
                      )}
                    </div>
                  )
                }

                return (
                  <NavLink
                    key={item.to}
                    to={item.to}
                    end={item.end}
                    className={({ isActive }) =>
                      `flex items-center justify-end gap-3 rounded-xl px-4 py-3 text-sm font-medium transition-all ${
                        isActive
                          ? 'bg-brand-500 text-white shadow-lg shadow-brand-900/30'
                          : 'text-slate-300 hover:bg-white/5 hover:text-white'
                      }`
                    }
                  >
                    <span>{item.label}</span>
                    <span className="opacity-80">{item.icon}</span>
                  </NavLink>
                )
              })}
            </nav>

            {/* Logout at bottom */}
            <div className="border-t border-white/10 p-3">
              <button
                onClick={() => setEmployee(null)}
                className="flex w-full items-center justify-end gap-3 rounded-xl px-4 py-3 text-sm font-medium text-slate-400 transition-all hover:bg-white/5 hover:text-white"
              >
                <span>تسجيل الخروج</span>
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4 M16 17l5-5-5-5 M21 12H9" />
                </svg>
              </button>
            </div>
          </aside>

          {/* Main Content */}
          <main className="flex-1 overflow-y-auto p-8">
            <Outlet />
          </main>
        </div>
      </div>
    </SessionContext.Provider>
  )
}
