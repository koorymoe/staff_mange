import { useState } from 'react'
import { api } from '../api'
import { useSession } from '../session'

const orbs = [
  { size: 220, top: '6%', left: '8%', delay: '0s' },
  { size: 90, top: '14%', left: '68%', delay: '1.2s' },
  { size: 140, top: '62%', left: '4%', delay: '2.4s' },
  { size: 60, top: '20%', left: '86%', delay: '0.6s' },
  { size: 260, top: '58%', left: '72%', delay: '1.8s' },
  { size: 40, top: '80%', left: '30%', delay: '3s' },
  { size: 110, top: '4%', left: '42%', delay: '2.1s' },
]

export default function Login() {
  const { setEmployee } = useSession()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      const employee = await api.login(username, password)
      setEmployee(employee)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-gradient-to-br from-[#0f2040] via-[#173463] to-[#2c5aad] p-4">
      {orbs.map((o, i) => (
        <span
          key={i}
          className="login-orb"
          style={{
            width: o.size,
            height: o.size,
            top: o.top,
            left: o.left,
            animationDelay: o.delay,
          }}
        />
      ))}

      <form
        onSubmit={handleSubmit}
        className="login-glass-card relative z-10 flex w-full max-w-[420px] flex-col items-center rounded-[2.5rem] px-8 py-12 sm:rounded-[3rem] sm:px-10"
      >
        <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-gradient-to-br from-[#2c5aad] to-[#1a3a6e] shadow-lg shadow-blue-900/40">
          <svg width="30" height="30" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
            <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
            <line x1="12" y1="22.08" x2="12" y2="12"/>
          </svg>
        </div>
        <h1 className="mt-4 text-center text-2xl font-extrabold text-white tracking-tight">شركة الأماني</h1>
        <p className="mt-1 text-center text-sm text-blue-100/70">نظام الإدارة المتكامل</p>

        <div className="mt-8 w-full">
          <label className="mb-1 block text-sm font-medium text-blue-100/80">اسم المستخدم</label>
          <div className="relative">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
              className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-slate-400">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
            </svg>
            <input
              required
              dir="ltr"
              autoCapitalize="off"
              autoCorrect="off"
              spellCheck={false}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full rounded-xl border border-white/0 bg-white/95 py-2.5 pl-3 pr-10 text-right text-sm text-slate-800 shadow-sm outline-none transition-all focus:ring-2 focus:ring-white/60"
            />
          </div>
        </div>

        <div className="mt-4 w-full">
          <label className="mb-1 block text-sm font-medium text-blue-100/80">كلمة المرور</label>
          <div className="relative">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
              className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-slate-400">
              <rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
            </svg>
            <input
              required
              type="password"
              dir="ltr"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full rounded-xl border border-white/0 bg-white/95 py-2.5 pl-3 pr-10 text-right text-sm text-slate-800 shadow-sm outline-none transition-all focus:ring-2 focus:ring-white/60"
            />
          </div>
        </div>

        {error && (
          <p className="mt-4 w-full rounded-xl bg-red-500/15 p-3 text-center text-sm text-red-100 ring-1 ring-red-400/30">
            {error}
          </p>
        )}

        <button
          type="submit"
          disabled={submitting}
          className="glossy-btn mt-8 w-full rounded-xl bg-gradient-to-l from-white to-blue-50 px-6 py-2.5 font-bold text-brand-800 shadow-lg shadow-blue-950/30 transition-all hover:shadow-xl disabled:opacity-50"
        >
          {submitting ? 'جاري الدخول...' : 'تسجيل الدخول'}
        </button>
      </form>
    </div>
  )
}
