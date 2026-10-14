import { NavLink } from 'react-router-dom'

const items = [
  { to: '/', label: 'الرئيسية', icon: 'M3 11l9-8 9 8v9a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1z' },
  { to: '/matrix', label: 'ماتركس', icon: 'M2 12s4-7 10-7 10 7 10 7-4 7-10 7S2 12 2 12zm10 3a3 3 0 1 0 0-6 3 3 0 0 0 0 6z' },
  { to: '/bookings', label: 'حجوزاتي', icon: 'M7 3v3m10-3v3M4 8h16M5 5h14a1 1 0 0 1 1 1v13a1 1 0 0 1-1 1H5a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1zm4 8l2 2 4-4' },
  { to: '/account', label: 'حسابي', icon: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8zm-8 9a8 8 0 0 1 16 0' },
]

export default function BottomNav({ badge }: { badge: number }) {
  return (
    <nav className="safe-b fixed inset-x-0 bottom-0 z-30 border-t border-slate-200 bg-white/95 backdrop-blur">
      <div className="mx-auto grid max-w-lg grid-cols-4">
        {items.map((it) => (
          <NavLink key={it.to} to={it.to} end={it.to === '/'}
            className={({ isActive }) => `relative flex flex-col items-center gap-0.5 py-2.5 text-[11px] font-bold ${isActive ? 'text-navy' : 'text-slate-400'}`}>
            {({ isActive }) => (
              <>
                <span className={`grid h-8 w-12 place-items-center rounded-full transition ${isActive ? 'bg-navy/10' : ''}`}>
                  <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.9" strokeLinecap="round" strokeLinejoin="round"><path d={it.icon} /></svg>
                </span>
                {it.label}
                {it.to === '/' && badge > 0 && <span className="absolute top-1 left-1/2 ml-3 grid h-4 min-w-4 place-items-center rounded-full bg-sun px-1 text-[9px] text-white">{badge}</span>}
              </>
            )}
          </NavLink>
        ))}
      </div>
    </nav>
  )
}
