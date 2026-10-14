import { NavLink } from 'react-router-dom'

const P = { fill: 'none', stroke: 'currentColor', strokeWidth: 1.9, strokeLinecap: 'round' as const, strokeLinejoin: 'round' as const }
const items = [
  { to: '/', label: 'الرئيسية', icon: <path {...P} d="M3 11l9-8 9 8v9a1 1 0 0 1-1 1h-5v-6H9v6H4a1 1 0 0 1-1-1z" /> },
  { to: '/services', label: 'الخدمات', icon: <g {...P}><rect x="3" y="3" width="7" height="7" rx="2" /><rect x="14" y="3" width="7" height="7" rx="2" /><rect x="3" y="14" width="7" height="7" rx="2" /><rect x="14" y="14" width="7" height="7" rx="2" /></g> },
  { to: '/offers', label: 'العروض', icon: <g {...P}><path d="M20 12v9H4v-9M2 7h20v5H2zM12 21V7M12 7H7.5a2.5 2.5 0 1 1 0-5C11 2 12 7 12 7zM12 7h4.5a2.5 2.5 0 1 0 0-5C13 2 12 7 12 7z" /></g> },
  { to: '/bookings', label: 'حجوزاتي', icon: <g {...P}><path d="M9 4h6a1 1 0 0 1 1 1v1H8V5a1 1 0 0 1 1-1z" /><path d="M16 5h2a1 1 0 0 1 1 1v14a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1h2M9 12h6M9 16h4" /></g> },
  { to: '/account', label: 'حسابي', icon: <g {...P}><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></g> },
]

export default function BottomNav({ badge }: { badge: number }) {
  return (
    <nav className="safe-b fixed inset-x-0 bottom-0 z-30 border-t border-line bg-white">
      <div className="mx-auto grid max-w-lg grid-cols-5">
        {items.map((it) => (
          <NavLink key={it.to} to={it.to} end={it.to === '/'} className={({ isActive }) => `relative flex flex-col items-center gap-0.5 pt-2 pb-2 text-[11px] font-bold transition ${isActive ? 'text-navy' : 'text-[#98a2b8]'}`}>
            {({ isActive }) => (
              <>
                <svg width="24" height="24" viewBox="0 0 24 24" className={isActive ? 'drop-shadow-[0_2px_4px_rgba(13,59,140,.25)]' : ''}>{it.icon}</svg>
                {it.label}
                {isActive && <span className="absolute top-0 h-[3px] w-8 rounded-b-full bg-navy" />}
                {it.to === '/offers' && badge > 0 && <span className="absolute top-1 left-1/2 ml-2 grid h-4 min-w-4 place-items-center rounded-full bg-gold px-1 text-[9px] font-black text-navy-deep">{badge}</span>}
              </>
            )}
          </NavLink>
        ))}
      </div>
    </nav>
  )
}
