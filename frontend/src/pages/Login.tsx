import { useEffect, useState } from 'react'
import { api } from '../api'
import { useSession } from '../session'

export default function Login() {
  const { setEmployee } = useSession()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [rememberMe, setRememberMe] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

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
    <div className="login-container">
      <div className="login-card">
        <div className="logo">
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="white" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
            <rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
          </svg>
        </div>
        <h2>أهلاً بعودتك 👋</h2>
        <p>سجّل الدخول للوصول لحساب نظام شركة الأماني</p>

        <form className="login-form" onSubmit={handleSubmit}>
          <div className="input-group">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
            </svg>
            <input
              type="text"
              placeholder="اسم المستخدم"
              required
              dir="ltr"
              autoCapitalize="off"
              autoCorrect="off"
              spellCheck={false}
              value={username}
              onChange={(e) => setUsername(e.target.value)}
            />
          </div>

          <div className="input-group">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
            </svg>
            <input
              type={showPassword ? 'text' : 'password'}
              placeholder="كلمة المرور"
              required
              dir="ltr"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
            />
            <svg
              className="toggle-password"
              width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
              onClick={() => setShowPassword((s) => !s)}
            >
              {showPassword ? (
                <path d="M17.94 17.94A10.94 10.94 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24M1 1l22 22"/>
              ) : (
                <>
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
                </>
              )}
            </svg>
          </div>

          <div className="remember-forgot">
            <label className="remember">
              <input type="checkbox" checked={rememberMe} onChange={(e) => setRememberMe(e.target.checked)} />
              تذكرني
            </label>
            <a href="#" className="forgot" onClick={(e) => { e.preventDefault(); alert('تواصل مع مدير النظام لإعادة تعيين كلمة المرور') }}>
              نسيت كلمة المرور؟
            </a>
          </div>

          {error && <p className="login-error">{error}</p>}

          <button type="submit" className="login-btn" disabled={submitting}>
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

        <p className="system-note">نظام الإدارة المتكامل — شركة الأماني</p>
      </div>
    </div>
  )
}
