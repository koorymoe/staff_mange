import { useState } from 'react'
import { api } from '../api'
import { useSession } from '../session'

// أبراج المدينة بالخلفية الفنية — ارتفاعات وعروض عشوائية الشكل بس ثابتة (مو Math.random حتى ما تتغير كل تحديث)
const buildings = [
  22, 38, 16, 55, 30, 70, 24, 45, 60, 20, 33, 48, 27, 65, 19, 40, 52, 28, 36, 58, 23, 44,
]

export default function Login() {
  const { setEmployee } = useSession()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
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
    <div className="flex min-h-screen bg-[#0a1730] lg:flex-row-reverse">
      {/* ===== لوحة المشهد الفني — تختفي بالموبايل ===== */}
      <div className="login-scene relative hidden w-1/2 lg:block">
        <div className="login-moon" />
        <div className="login-skyline">
          {buildings.map((h, i) => (
            <div key={i} className="bld" style={{ height: `${h}%`, width: `${100 / buildings.length}%` }} />
          ))}
        </div>
        <div className="login-scene-water" />

        <div className="relative z-10 flex h-full flex-col justify-between p-12">
          <div className="flex items-center gap-3">
            <div className="flex h-12 w-12 items-center justify-center rounded-2xl bg-gradient-to-br from-[#4a7bcb] to-[#152c58] shadow-lg shadow-blue-900/40">
              <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
                <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
                <line x1="12" y1="22.08" x2="12" y2="12"/>
              </svg>
            </div>
            <span className="text-lg font-extrabold text-white tracking-tight">شركة الأماني</span>
          </div>

          <div className="max-w-sm">
            <h2 className="text-3xl font-extrabold leading-tight text-white">
              نظام إدارة متكامل<br />بين إيديك بكل مكان
            </h2>
            <p className="mt-4 text-sm leading-relaxed text-blue-100/70">
              جي بي اس، كاميرات، صيانة مركبات، وكل عمليات الشركة — بمكان وحد، بمتابعة لحظية.
            </p>
          </div>
        </div>
      </div>

      {/* ===== لوحة تسجيل الدخول ===== */}
      <div className="relative flex w-full items-center justify-center overflow-hidden bg-gradient-to-br from-[#0f2040] via-[#173463] to-[#2c5aad] p-6 lg:w-1/2">
        <form onSubmit={handleSubmit} className="relative z-10 flex w-full max-w-sm flex-col">
          <div className="flex h-14 w-14 items-center justify-center rounded-2xl bg-gradient-to-br from-[#4a7bcb] to-[#152c58] shadow-lg shadow-blue-900/40 lg:hidden">
            <svg width="26" height="26" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/>
              <polyline points="3.27 6.96 12 12.01 20.73 6.96"/>
              <line x1="12" y1="22.08" x2="12" y2="12"/>
            </svg>
          </div>

          <h1 className="mt-6 text-2xl font-extrabold text-white tracking-tight lg:mt-0">أهلاً بعودتك 👋</h1>
          <p className="mt-1 text-sm text-blue-100/70">سجّل الدخول للوصول لحسابك بنظام شركة الأماني</p>

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

          {error && (
            <p className="mt-4 w-full rounded-xl bg-red-500/15 p-3 text-center text-sm text-red-100 ring-1 ring-red-400/30">
              {error}
            </p>
          )}

          <button
            type="submit"
            disabled={submitting}
            className="glossy-btn mt-8 flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-l from-[#4a7bcb] to-[#1a376e] px-6 py-2.5 font-bold text-white shadow-lg shadow-blue-950/40 transition-all hover:shadow-xl disabled:opacity-50"
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

          <p className="mt-6 text-center text-xs text-blue-100/50">نظام الإدارة المتكامل — شركة الأماني</p>
        </form>
      </div>
    </div>
  )
}
