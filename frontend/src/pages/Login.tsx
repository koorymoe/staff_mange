import { useEffect, useState } from 'react'
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

// أيقونات خافتة بزوايا الخلفية — ترمز لخدمات الشركة (جي بي اس، كاميرات، طاقة شمسية، إنذار حريق)
const bgIcons = [
  { top: '8%', right: '6%', size: 130, rotate: '-8deg', d: 'M23 7l-7 5 7 5V7z M1 5h15v14H1z', label: 'camera' }, // كاميرا
  { top: '68%', right: '10%', size: 150, rotate: '4deg', d: 'M12 2 2 7l10 5 10-5-10-5z M2 17l10 5 10-5 M2 12l10 5 10-5', label: 'solar' }, // طاقة شمسية (رمزي)
  { top: '72%', left: '4%', size: 110, rotate: '-4deg', d: 'M5 12.55a11 11 0 0 1 14.08 0 M1.42 9a16 16 0 0 1 21.16 0 M8.53 16.11a6 6 0 0 1 6.95 0 M12 20h.01', label: 'network' }, // شبكة/جي بي اس
  { top: '4%', left: '10%', size: 90, rotate: '6deg', d: 'M12 9v4M12 17h.01M10.29 3.86 1.82 18a2 2 0 0 0 1.71 3h16.94a2 2 0 0 0 1.71-3L13.71 3.86a2 2 0 0 0-3.42 0z', label: 'alert' }, // إنذار
]

export default function Login() {
  const { setEmployee } = useSession()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [rememberMe, setRememberMe] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [showForgotHint, setShowForgotHint] = useState(false)

  useEffect(() => {
    const saved = localStorage.getItem('rememberedUsername')
    if (saved) setUsername(saved)
  }, [])

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      const employee = await api.login(username, password)
      if (rememberMe) localStorage.setItem('rememberedUsername', username)
      else localStorage.removeItem('rememberedUsername')
      setEmployee(employee)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="relative flex min-h-screen items-center justify-center overflow-hidden bg-gradient-to-br from-[#050b1a] via-[#0f2040] to-[#1a3a6e] p-4">
      {/* شبكة تقنية خافتة بالخلفية */}
      <div
        className="pointer-events-none absolute inset-0 opacity-[0.07]"
        style={{
          backgroundImage:
            'linear-gradient(rgba(125,163,222,0.6) 1px, transparent 1px), linear-gradient(90deg, rgba(125,163,222,0.6) 1px, transparent 1px)',
          backgroundSize: '42px 42px',
        }}
      />

      {bgIcons.map((ic, i) => (
        <svg
          key={i}
          width={ic.size}
          height={ic.size}
          viewBox="0 0 24 24"
          fill="none"
          stroke="currentColor"
          strokeWidth="1"
          strokeLinecap="round"
          strokeLinejoin="round"
          className="pointer-events-none absolute text-blue-300/[0.08]"
          style={{ top: ic.top, right: ic.right, left: ic.left, transform: `rotate(${ic.rotate})` }}
        >
          <path d={ic.d} />
        </svg>
      ))}

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
        <div className="flex h-16 w-16 items-center justify-center rounded-full bg-gradient-to-br from-[#4a7bcb] to-[#152c58] shadow-lg shadow-blue-900/50">
          <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
          </svg>
        </div>
        <h1 className="mt-4 text-center text-2xl font-extrabold text-white tracking-tight">شركة الأماني</h1>
        <p className="mt-1 text-center text-sm text-blue-100/70">مرحباً بعودتك، سجّل الدخول للمتابعة</p>

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
              type={showPassword ? 'text' : 'password'}
              dir="ltr"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full rounded-xl border border-white/0 bg-white/95 py-2.5 pl-10 pr-10 text-right text-sm text-slate-800 shadow-sm outline-none transition-all focus:ring-2 focus:ring-white/60"
            />
            <button
              type="button"
              onClick={() => setShowPassword((s) => !s)}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-600"
              tabIndex={-1}
            >
              {showPassword ? (
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M17.94 17.94A10.94 10.94 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/>
                </svg>
              ) : (
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
                </svg>
              )}
            </button>
          </div>
        </div>

        <div className="mt-3 flex w-full items-center justify-between text-xs">
          <label className="flex items-center gap-1.5 text-blue-100/80">
            <input
              type="checkbox"
              checked={rememberMe}
              onChange={(e) => setRememberMe(e.target.checked)}
              className="h-3.5 w-3.5 rounded border-white/30 accent-brand-400"
            />
            تذكرني
          </label>
          <button
            type="button"
            onClick={() => setShowForgotHint((s) => !s)}
            className="font-medium text-blue-200 hover:text-white hover:underline"
          >
            نسيت كلمة المرور؟
          </button>
        </div>
        {showForgotHint && (
          <p className="mt-2 w-full rounded-lg bg-white/10 p-2.5 text-center text-xs text-blue-100">
            تواصل مع مدير النظام لإعادة تعيين كلمة المرور
          </p>
        )}

        {error && (
          <p className="mt-4 w-full rounded-xl bg-red-500/15 p-3 text-center text-sm text-red-100 ring-1 ring-red-400/30">
            {error}
          </p>
        )}

        <button
          type="submit"
          disabled={submitting}
          className="glossy-btn mt-6 flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-l from-[#4a7bcb] to-[#1a376e] px-6 py-2.5 font-bold text-white shadow-lg shadow-blue-950/40 transition-all hover:shadow-xl disabled:opacity-50"
        >
          {submitting ? 'جاري الدخول...' : (
            <>
              تسجيل الدخول
              <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
                <line x1="19" y1="12" x2="5" y2="12"/><polyline points="12 5 5 12 12 19"/>
              </svg>
            </>
          )}
        </button>
      </form>
    </div>
  )
}
