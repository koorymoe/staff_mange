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
  leaderOnly?: boolean
  children?: NavItem[]
}

const I = ({ d }: { d: string }) => (
  <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d={d} /></svg>
)

const navItems: NavItem[] = [
  { to: '/', label: 'الرئيسية', end: true, icon: <I d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z M9 22V12h6v10" /> },
  { to: '/attendance', label: 'الحضور', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="10"/><polyline points="12 6 12 12 16 14"/></svg> },

  // ── الإدارة ──
  {
    to: '/admin-group', label: 'الإدارة', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9"/></svg>,
    roles: ['ADMIN', 'HR_COORDINATOR', 'SALES', 'MONITOR', 'FINANCE', 'PROJECT_MANAGER', 'GPS_ADMIN'],
    children: [
      {
        to: '/mgmt-employees', label: 'إدارة الموظفين', icon: <></>,
        roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'],
        children: [
          { to: '/employees', label: 'إدارة الكوادر', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR'] },
          { to: '/permissions', label: 'الصلاحيات', icon: <></>, roles: ['ADMIN'] },
          { to: '/kpi', label: 'تقييم الأداء', icon: <></>, roles: ['ADMIN', 'MONITOR'], permission: 'kpi_management' },
          { to: '/inventory', label: 'جرد الأدوات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR'], permission: 'inventory' },
          { to: '/stats', label: 'إحصائيات الموظفين', icon: <></>, roles: ['ADMIN'] },
          { to: '/complaints', label: 'الشكاوى', icon: <></>, roles: ['ADMIN', 'SALES', 'HR_COORDINATOR'], permission: 'complaints' },
        ],
      },
      {
        to: '/mgmt-work', label: 'إدارة العمل', icon: <></>,
        roles: ['ADMIN', 'HR_COORDINATOR', 'SALES', 'MONITOR', 'FINANCE'],
        children: [
          { to: '/sales', label: 'حجز جديد', icon: <></>, roles: ['ADMIN', 'SALES'], permission: 'sales_booking' },
          { to: '/customers', label: 'العملاء', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'manage_customers' },
          { to: '/bookings', label: 'الحجوزات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR', 'FINANCE'], permission: 'view_bookings' },
          { to: '/coordinator', label: 'تنسيق الحجوزات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR'], permission: 'coordinator' },
          { to: '/services', label: 'الخدمات', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'] },
          { to: '/missions', label: 'تتبع المهام', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'], permission: 'mission_tracking' },
        ],
      },
      {
        to: '/mgmt-services', label: 'إدارة الخدمات', icon: <></>,
        roles: ['ADMIN', 'GPS_ADMIN', 'GPS_ENGINEER'],
        children: [
          { to: '/gps', label: 'نظام GPS', icon: <></>, roles: ['ADMIN', 'GPS_ADMIN', 'GPS_ENGINEER'], permission: 'gps_system' },
        ],
      },
      {
        to: '/mgmt-projects', label: 'إدارة المشاريع', icon: <></>,
        roles: ['ADMIN', 'PROJECT_MANAGER', 'SALES'],
        children: [
          { to: '/projects', label: 'المشاريع', icon: <></>, roles: ['ADMIN', 'PROJECT_MANAGER'], permission: 'project_management' },
          { to: '/quotations', label: 'عروض الأسعار', icon: <></>, roles: ['ADMIN', 'SALES'], permission: 'quotation_system' },
          { to: '/products', label: 'المنتجات', icon: <></>, roles: ['ADMIN'], permission: 'quotation_system' },
        ],
      },
      {
        to: '/mgmt-finance', label: 'إدارة الحسابات', icon: <></>,
        roles: ['ADMIN', 'FINANCE'],
        children: [
          { to: '/finance', label: 'تدقيق الحسابات', icon: <></>, roles: ['ADMIN', 'FINANCE'], permission: 'finance' },
          { to: '/my-expenses', label: 'المصاريف', icon: <></>, roles: ['ADMIN', 'TECHNICIAN', 'PROJECT_MANAGER'], permission: 'expenses' },
        ],
      },
      {
        to: '/mgmt-procurement', label: 'إدارة المشتريات', icon: <></>,
        roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'],
        children: [
          { to: '/suppliers', label: 'الموردون', icon: <></>, roles: ['ADMIN', 'HR_COORDINATOR', 'MONITOR'] },
        ],
      },
    ],
  },

  { to: '/work-reports', label: 'تقارير العمل', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/><polyline points="14 2 14 8 20 8"/><line x1="16" y1="13" x2="8" y2="13"/><line x1="16" y1="17" x2="8" y2="17"/></svg>, roles: ['TECHNICIAN', 'PROJECT_MANAGER'] },
  { to: '/my-tasks', label: 'مهامي', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M9 11l3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/></svg>, roles: ['TECHNICIAN', 'PROJECT_MANAGER'] },
  { to: '/my-ranking', label: 'تصنيفي', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><polygon points="12 2 15.09 8.26 22 9.27 17 14.14 18.18 21.02 12 17.77 5.82 21.02 7 14.14 2 9.27 8.91 8.26 12 2"/></svg>, roles: ['TECHNICIAN'] },
  { to: '/my-expenses', label: 'مصاريفي', icon: <I d="M12 1v22M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />, roles: ['TECHNICIAN', 'PROJECT_MANAGER'], leaderOnly: true },
  { to: '/my-inventory', label: 'جرد أدواتي', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M14.7 6.3a1 1 0 0 0 0 1.4l1.6 1.6a1 1 0 0 0 1.4 0l3.77-3.77a6 6 0 0 1-7.94 7.94l-6.91 6.91a2.12 2.12 0 0 1-3-3l6.91-6.91a6 6 0 0 1 7.94-7.94l-3.76 3.76z"/></svg>, roles: ['TECHNICIAN'] },
  { to: '/gps/employee', label: 'لوحتي GPS', icon: <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round"><path d="M21 10c0 7-9 13-9 13s-9-6-9-13a9 9 0 0 1 18 0z"/><circle cx="12" cy="10" r="3"/></svg>, roles: ['GPS_ENGINEER'] },
]

const loadStoredEmployee = (): Employee | null => {
  const raw = localStorage.getItem('currentEmployee')
  if (!raw) return null
  try { return JSON.parse(raw) as Employee } catch { return null }
}

function hasActiveChild(item: NavItem, pathname: string): boolean {
  if (!item.children) return pathname === item.to || pathname.startsWith(item.to + '/')
  return item.children.some(c => hasActiveChild(c, pathname))
}

const roleColors: Record<string, string> = {
  ADMIN: 'from-amber-500 to-orange-600',
  SALES: 'from-emerald-500 to-teal-600',
  HR_COORDINATOR: 'from-violet-500 to-purple-600',
  TECHNICIAN: 'from-sky-500 to-blue-600',
  PROJECT_MANAGER: 'from-rose-500 to-pink-600',
  MONITOR: 'from-cyan-500 to-teal-600',
  FINANCE: 'from-lime-500 to-green-600',
  GPS_ADMIN: 'from-indigo-500 to-blue-600',
  GPS_ENGINEER: 'from-blue-500 to-indigo-600',
  QUALITY_ENGINEER: 'from-fuchsia-500 to-purple-600',
}

export default function Layout() {
  const [employee, setEmployeeState] = useState<Employee | null>(loadStoredEmployee)
  const [employeePermissions, setEmployeePermissions] = useState<string[]>([])
  const [expandedGroups, setExpandedGroups] = useState<Record<string, boolean>>({})
  const [collapsed, setCollapsed] = useState(false)
  const location = useLocation()

  const setEmployee = (emp: Employee | null) => {
    setEmployeeState(emp)
    if (emp) localStorage.setItem('currentEmployee', JSON.stringify(emp))
    else localStorage.removeItem('currentEmployee')
  }

  useEffect(() => {
    if (!employee) { setEmployeePermissions([]); return }
    api.getEmployeePermissions(employee.id)
      .then((perms) => setEmployeePermissions(perms.map((p: any) => p.name)))
      .catch(() => setEmployeePermissions([]))
  }, [employee?.id])

  useEffect(() => {
    const autoExpand = (items: NavItem[]) => {
      items.forEach((item) => {
        if (item.children && hasActiveChild(item, location.pathname)) {
          setExpandedGroups((prev) => ({ ...prev, [item.label]: true }))
          autoExpand(item.children)
        }
      })
    }
    autoExpand(navItems)
  }, [location.pathname])

  if (!employee) {
    return (
      <SessionContext.Provider value={{ employee, setEmployee, permissions: employeePermissions }}>
        <Login />
      </SessionContext.Provider>
    )
  }

  const role = employee?.role
  const isVisible = (item: NavItem): boolean => {
    if (item.roles && role && !item.roles.includes(role)) return false
    if (item.permission && role !== 'ADMIN' && !employeePermissions.includes(item.permission)) return false
    if (item.leaderOnly && !(employee as any)?.isLeader && role !== 'ADMIN') return false
    if (item.children) return item.children.some(isVisible)
    return true
  }

  const visibleItems = navItems.filter(isVisible)

  const toggle = (label: string) => setExpandedGroups((p) => ({ ...p, [label]: !p[label] }))

  const gradientClass = roleColors[employee.role] || 'from-blue-500 to-indigo-600'

  const renderNavItem = (item: NavItem, depth: number = 0): React.ReactNode => {
    if (!isVisible(item)) return null

    if (item.children) {
      const kids = item.children.filter(isVisible)
      if (!kids.length) return null
      const open = expandedGroups[item.label]
      const active = hasActiveChild(item, location.pathname)

      if (depth === 0) {
        return (
          <div key={item.label}>
            <button
              onClick={() => toggle(item.label)}
              className={`group flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-semibold transition-all duration-200 ${
                active
                  ? 'bg-white/[0.12] text-white shadow-[inset_0_0_0_1px_rgba(255,255,255,0.1)]'
                  : 'text-blue-200/70 hover:bg-white/[0.06] hover:text-white'
              }`}
            >
              {!collapsed && (
                <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round"
                  className={`transition-transform duration-300 ${open ? 'rotate-180' : ''}`} style={{ flexShrink: 0 }}>
                  <polyline points="6 9 12 15 18 9"/>
                </svg>
              )}
              {!collapsed && <span className="flex-1 text-right">{item.label}</span>}
              <span style={{ flexShrink: 0 }} className="opacity-80 group-hover:opacity-100 transition-opacity">{item.icon}</span>
            </button>
            {open && !collapsed && (
              <div className="mt-1 mr-4 flex flex-col gap-0.5 border-r-2 border-white/[0.08] pr-2 animate-in">
                {kids.map(child => renderNavItem(child, 1))}
              </div>
            )}
          </div>
        )
      }

      return (
        <div key={item.label}>
          <button
            onClick={() => toggle(item.label)}
            className={`group flex w-full items-center gap-2 rounded-lg px-3 py-2 text-[13px] font-bold transition-all duration-200 ${
              active ? 'bg-white/[0.08] text-white' : 'text-slate-400 hover:bg-white/[0.04] hover:text-slate-200'
            }`}
          >
            <svg width="10" height="10" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round"
              className={`transition-transform duration-300 ${open ? 'rotate-180' : ''}`} style={{ flexShrink: 0 }}>
              <polyline points="6 9 12 15 18 9"/>
            </svg>
            <span className="flex-1 text-right">{item.label}</span>
            {active && <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse" style={{ flexShrink: 0 }}/>}
          </button>
          {open && (
            <div className="mr-4 flex flex-col gap-0.5 border-r border-white/[0.06] pr-2">
              {kids.map(child => renderNavItem(child, 2))}
            </div>
          )}
        </div>
      )
    }

    // Leaf
    if (depth === 0) {
      return (
        <NavLink key={item.to} to={item.to} end={item.end}
          className={({ isActive }) =>
            `group relative flex items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-semibold transition-all duration-200 ${
              isActive
                ? 'bg-gradient-to-l from-[#2c5aad]/90 to-[#1e3f7a] text-white shadow-lg shadow-blue-900/30'
                : 'text-blue-200/70 hover:bg-white/[0.06] hover:text-white'
            }`
          }>
          {({ isActive }) => (
            <>
              {isActive && <span className="absolute right-0 top-1/2 -translate-y-1/2 h-6 w-1 rounded-l-full bg-white/80" />}
              {!collapsed && <span className="flex-1 text-right">{item.label}</span>}
              <span style={{ flexShrink: 0 }} className="opacity-80 group-hover:opacity-100 transition-opacity">{item.icon}</span>
            </>
          )}
        </NavLink>
      )
    }

    return (
      <NavLink key={item.to} to={item.to} end={item.end}
        className={({ isActive }) =>
          `group relative rounded-lg px-4 py-1.5 text-right text-[12.5px] font-medium transition-all duration-200 ${
            isActive
              ? 'bg-white/[0.1] text-white font-semibold'
              : 'text-gray-500 hover:bg-white/[0.04] hover:text-gray-300'
          }`
        }>
        {({ isActive }) => (
          <>
            {isActive && <span className="absolute right-0 top-1/2 -translate-y-1/2 h-4 w-0.5 rounded-l-full bg-[#2c5aad]" />}
            {item.label}
          </>
        )}
      </NavLink>
    )
  }

  return (
    <SessionContext.Provider value={{ employee, setEmployee, permissions: employeePermissions }}>
      <div dir="ltr" className="flex min-h-screen bg-[#f0f4f9]">

        {/* ===== Main Area ===== */}
        <div dir="rtl" className="flex flex-1 flex-col">
          {/* Top Header — Glass effect */}
          <header className="sticky top-0 z-30 flex h-16 items-center justify-between border-b border-slate-200/60 bg-white/80 backdrop-blur-xl px-8">
            <div className="flex items-center gap-3">
              <button
                onClick={() => setCollapsed(!collapsed)}
                className="flex h-9 w-9 items-center justify-center rounded-lg text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600"
              >
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <line x1="3" y1="6" x2="21" y2="6"/><line x1="3" y1="12" x2="15" y2="12"/><line x1="3" y1="18" x2="21" y2="18"/>
                </svg>
              </button>
              <div className="h-8 w-px bg-slate-200"/>
              <span className="text-lg font-extrabold text-[#0f2040] tracking-tight">نظام شركة الأماني</span>
            </div>
            <div className="flex items-center gap-3">
              <button className="relative flex h-10 w-10 items-center justify-center rounded-xl text-slate-400 transition-all hover:bg-slate-100 hover:text-slate-600">
                <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9M13.73 21a2 2 0 0 1-3.46 0"/>
                </svg>
              </button>
              <div className="flex items-center gap-3 rounded-xl bg-slate-50/80 px-3 py-1.5 transition-colors hover:bg-slate-100">
                <div className="text-left">
                  <p className="text-sm font-bold text-slate-800">{employee.name}</p>
                  <p className="text-[11px] text-slate-400">{roleLabels[employee.role]}</p>
                </div>
                <div className={`relative flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br ${gradientClass} text-sm font-bold text-white shadow-md`}>
                  {employee.name.charAt(0)}
                  <span className="absolute -bottom-0.5 -left-0.5 h-3 w-3 rounded-full border-2 border-white bg-emerald-500"/>
                </div>
              </div>
            </div>
          </header>

          {/* Content */}
          <main className="flex-1 overflow-y-auto p-8">
            <Outlet />
          </main>
        </div>

        {/* ===== Right Sidebar — Premium ===== */}
        <aside
          dir="rtl"
          className="sticky top-0 flex h-screen flex-col bg-[#0f2040] transition-all duration-300 ease-in-out"
          style={{ width: collapsed ? 72 : 270, minWidth: collapsed ? 72 : 270 }}
        >
          {/* Logo */}
          <div className={`flex items-center ${collapsed ? 'justify-center' : 'justify-center gap-3'} px-4 py-5`}>
            <div className="relative flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br from-[#2c5aad] to-[#1a3a6e] shadow-lg shadow-blue-900/40">
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
                <line x1="12" y1="22.08" x2="12" y2="12"/>
              </svg>
              <span className="absolute -top-1 -left-1 h-3 w-3 rounded-full border-2 border-[#0f2040] bg-emerald-400"/>
            </div>
            {!collapsed && (
              <div>
                <p className="text-sm font-extrabold text-white tracking-tight">الأماني</p>
                <p className="text-[10px] text-blue-300/50 font-medium">Management System</p>
              </div>
            )}
          </div>

          {/* User card */}
          {!collapsed ? (
            <div className="mx-3 mb-4 overflow-hidden rounded-2xl bg-gradient-to-br from-white/[0.08] to-white/[0.03] p-4 shadow-[inset_0_0_0_1px_rgba(255,255,255,0.06)]">
              <div className="flex items-center gap-3">
                <div className="flex-1 text-right">
                  <p className="text-sm font-bold text-white">{employee.name}</p>
                  <p className="text-[11px] text-blue-300/60">{roleLabels[employee.role]}</p>
                </div>
                <div className={`relative flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br ${gradientClass} text-sm font-bold text-white shadow-lg`}>
                  {employee.name.charAt(0)}
                  <span className="absolute -bottom-0.5 -left-0.5 h-3 w-3 rounded-full border-2 border-[#0f2040] bg-emerald-400"/>
                </div>
              </div>
              <div className="mt-3 flex items-center gap-2">
                <span className="h-1.5 w-1.5 rounded-full bg-emerald-400 animate-pulse"/>
                <span className="text-[10px] text-emerald-400/80 font-medium">متصل الآن</span>
                <span className={`mr-auto rounded-full bg-gradient-to-l ${gradientClass} px-2.5 py-0.5 text-[10px] font-bold text-white shadow-sm`}>
                  {roleLabels[employee.role]}
                </span>
              </div>
            </div>
          ) : (
            <div className="mx-auto mb-3">
              <div className={`relative flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br ${gradientClass} text-sm font-bold text-white shadow-lg`}>
                {employee.name.charAt(0)}
                <span className="absolute -bottom-0.5 -left-0.5 h-2.5 w-2.5 rounded-full border-2 border-[#0f2040] bg-emerald-400"/>
              </div>
            </div>
          )}

          {/* Divider */}
          <div className="mx-4 mb-2 h-px bg-gradient-to-l from-transparent via-white/10 to-transparent"/>

          {/* Nav */}
          <nav className="flex flex-1 flex-col gap-0.5 overflow-y-auto px-3 pb-3 scrollbar-thin">
            {visibleItems.map(item => renderNavItem(item, 0))}
          </nav>

          {/* Logout */}
          <div className="mx-3 mb-3 mt-1">
            <div className="h-px bg-gradient-to-l from-transparent via-white/10 to-transparent mb-3"/>
            <button onClick={() => setEmployee(null)}
              className="group flex w-full items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium text-red-300/50 transition-all duration-200 hover:bg-red-500/10 hover:text-red-300">
              {!collapsed && <span className="flex-1 text-right">تسجيل الخروج</span>}
              <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" className="opacity-60 group-hover:opacity-100 transition-opacity">
                <path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4M16 17l5-5-5-5M21 12H9"/>
              </svg>
            </button>
          </div>
        </aside>
      </div>
    </SessionContext.Provider>
  )
}
