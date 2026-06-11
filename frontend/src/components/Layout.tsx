import { NavLink, Outlet } from 'react-router-dom'

const navItems = [
  { to: '/', label: 'الرئيسية', end: true },
  { to: '/employees', label: 'إدارة الكوادر' },
  { to: '/customers', label: 'الزبائن' },
  { to: '/services', label: 'الخدمات' },
]

export default function Layout() {
  return (
    <div className="flex min-h-screen bg-slate-100">
      <aside className="flex w-64 flex-col bg-brand-900 text-white">
        <div className="border-b border-white/10 px-6 py-6">
          <h1 className="text-xl font-extrabold text-gold-400">شركة الأماني</h1>
          <p className="mt-1 text-sm text-brand-200">نظام الإدارة المتكامل</p>
        </div>
        <nav className="flex flex-col gap-1 p-4">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
              className={({ isActive }) =>
                `rounded-lg px-4 py-3 text-sm font-medium transition-colors ${
                  isActive
                    ? 'bg-gold-500 text-brand-900'
                    : 'text-brand-100 hover:bg-brand-800'
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
  )
}
