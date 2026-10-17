import type { ReactNode } from 'react'
import { useNavigate } from 'react-router-dom'
import { SHORT } from '../lib'

// شريط علوي أبيض بعنوان بالوسط وزر رجوع — مثل التصميم
export function TopBar({ title, back = true, end }: { title: string; back?: boolean; end?: ReactNode }) {
  const nav = useNavigate()
  return (
    <header className="safe-t sticky top-0 z-20 bg-white/95 backdrop-blur">
      <div className="relative flex h-14 items-center justify-center px-4">
        {back && (
          <button aria-label="رجوع" onClick={() => (history.length > 1 ? nav(-1) : nav('/'))} className="absolute right-3 grid h-10 w-10 place-items-center rounded-full text-ink active:bg-slate-100">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round"><path d="M9 6l6 6-6 6" /></svg>
          </button>
        )}
        <h1 className="text-[17px] font-extrabold text-ink">{title}</h1>
        {end && <div className="absolute left-3">{end}</div>}
      </div>
    </header>
  )
}

const PILL: Record<string, string> = {
  PENDING: 'bg-sky-50 text-sky-700', CONFIRMED: 'bg-blue-50 text-blue-700', WAITING: 'bg-amber-50 text-amber-700',
  IN_PROGRESS: 'bg-amber-50 text-amber-700', PARTIAL: 'bg-amber-50 text-amber-700', COMPLETED: 'bg-emerald-50 text-emerald-700', CANCELLED: 'bg-rose-50 text-rose-600',
}

export function StatusPill({ status, label }: { status: string; label?: string }) {
  return <span className={`inline-flex shrink-0 items-center rounded-lg px-3 py-1 text-[12px] font-bold ${PILL[status] ?? 'bg-slate-100 text-slate-600'}`}>{label ?? SHORT[status] ?? status}</span>
}

export function Skel({ className = '' }: { className?: string }) {
  return <div className={`skel rounded-2xl ${className}`} />
}

export function Stars({ value, size = 14 }: { value: number; size?: number }) {
  return (
    <span className="inline-flex" dir="ltr">
      {[1, 2, 3, 4, 5].map((i) => (
        <svg key={i} width={size} height={size} viewBox="0 0 24 24" fill={i <= Math.round(value) ? '#f5b335' : '#e5e7eb'}><path d="M12 2l3.1 6.3 6.9 1-5 4.9 1.2 6.8L12 17.8 5.8 21l1.2-6.8-5-4.9 6.9-1z" /></svg>
      ))}
    </span>
  )
}

export function Empty({ icon, title, children }: { icon: string; title: string; children?: ReactNode }) {
  return (
    <div className="card rise mx-auto mt-6 p-8 text-center">
      <p className="text-5xl">{icon}</p>
      <p className="mt-3 text-base font-extrabold text-ink">{title}</p>
      {children}
    </div>
  )
}

