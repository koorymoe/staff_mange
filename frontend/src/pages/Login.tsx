import { useEffect, useRef, useState } from 'react'
import { api } from '../api'
import CommandLogin from '../command/CommandLogin'
import '../command/theme.css'
import { useSession } from '../session'

const UserIcon = () => <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
const LockIcon = () => <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>

export default function Login() {
  const { setEmployee } = useSession()
  const [username, setUsername] = useState(() => localStorage.getItem('rememberedUser') || '')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [showPass, setShowPass] = useState(false)
  const [remember, setRemember] = useState(() => Boolean(localStorage.getItem('rememberedUser')))
  const [sessionNote] = useState(() => sessionStorage.getItem('sessionEndedReason'))
  const [curtain, setCurtain] = useState(0)
  const dragRef = useRef<{ startY: number; from: number } | null>(null)
  const [dragging, setDragging] = useState(false)

  useEffect(() => { sessionStorage.removeItem('sessionEndedReason') }, [])

  const onHandleDown = (event: React.PointerEvent<HTMLButtonElement>) => {
    event.currentTarget.setPointerCapture(event.pointerId)
    dragRef.current = { startY: event.clientY, from: curtain }
    setDragging(true)
  }
  const onHandleMove = (event: React.PointerEvent<HTMLButtonElement>) => {
    const drag = dragRef.current
    if (!drag) return
    setCurtain(Math.min(1, Math.max(0, drag.from + (drag.startY - event.clientY) / window.innerHeight)))
  }
  const onHandleUp = () => {
    if (!dragRef.current) return
    dragRef.current = null
    setDragging(false)
    setCurtain((value) => value > 0.3 ? 1 : 0)
  }
  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      const employee = await api.login(username, password)
      if (remember) localStorage.setItem('rememberedUser', username)
      else localStorage.removeItem('rememberedUser')
      setEmployee(employee)
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'حدث خطأ غير متوقع')
    } finally { setSubmitting(false) }
  }

  return (
    <main className="login-live-stage" dir="rtl">
      <div className="login-live-scene" aria-hidden="true">
        <img src={`${import.meta.env.BASE_URL}login-live-background.png`} alt="" />
        <span className="login-live-scan" />
        <span className="login-live-particle particle-one" />
        <span className="login-live-particle particle-two" />
        <span className="login-live-particle particle-three" />
      </div>

      <section className="login-live-card" aria-labelledby="login-title">
        <div className="login-live-logo"><img src={`${import.meta.env.BASE_URL}favicon.png?v=3`} alt="شعار شركة الأماني" /></div>
        <header className="login-live-header">
          <h1 id="login-title">شركة الأماني</h1>
          <div className="login-live-subtitle"><span />نظام الإدارة المتكامل<span /></div>
        </header>

        <form className="login-live-form" onSubmit={handleSubmit}>
          <label htmlFor="login-username">اسم المستخدم</label>
          <div className="login-live-control">
            <UserIcon />
            <input id="login-username" required dir="ltr" autoCapitalize="off" autoCorrect="off" spellCheck={false} autoComplete="username" placeholder="أدخل اسم المستخدم" value={username} onChange={(event) => setUsername(event.target.value)} />
          </div>

          <label htmlFor="login-password">كلمة المرور</label>
          <div className="login-live-control">
            <LockIcon />
            <input id="login-password" required type={showPass ? 'text' : 'password'} dir="ltr" autoComplete="current-password" placeholder="أدخل كلمة المرور" value={password} onChange={(event) => setPassword(event.target.value)} />
            <button type="button" className="login-live-eye" onClick={() => setShowPass((value) => !value)} aria-label={showPass ? 'إخفاء كلمة المرور' : 'إظهار كلمة المرور'}>
              <svg viewBox="0 0 24 24" aria-hidden="true">
                {showPass ? <path d="M17.94 17.94A10.1 10.1 0 0 1 12 20c-7 0-11-8-11-8a18.5 18.5 0 0 1 5.06-5.94M9.9 4.24A9 9 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24M1 1l22 22"/> : <><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></>}
              </svg>
            </button>
          </div>

          <div className="login-live-options">
            <label className="login-live-remember"><input type="checkbox" checked={remember} onChange={(event) => setRemember(event.target.checked)} /><span>تذكرني</span></label>
            <button type="button" className="login-live-forgot" onClick={() => setError('تُسترجع كلمة المرور من مدير النظام — راجعه لإعادة ضبطها.')}>نسيت كلمة المرور؟</button>
          </div>

          <button type="submit" disabled={submitting} className="login-live-submit"><span>{submitting ? 'جاري الدخول...' : 'تسجيل الدخول'}</span>{!submitting && <b aria-hidden="true">←</b>}</button>
          {(sessionNote || error) && <p className={`login-live-message ${error ? 'is-error' : ''}`} role="alert">{error || sessionNote}</p>}
        </form>
        <footer className="login-live-secure"><span />♢ دخول آمن ومحمي<span /></footer>
      </section>

      <div className="login-live-platform" aria-hidden="true"><span /><i /></div>
      <button type="button" aria-label="ارفع الستارة" className="curtain-handle" onPointerDown={onHandleDown} onPointerMove={onHandleMove} onPointerUp={onHandleUp} onPointerCancel={onHandleUp} onClick={() => setCurtain((value) => value === 0 ? 1 : 0)} />
      <div className="curtain" inert={curtain === 0 ? true : undefined} style={{ transform: `translateY(${(1 - curtain) * 100}%)`, transition: dragging ? 'none' : 'transform 0.62s cubic-bezier(0.22, 1, 0.36, 1)' }}>
        <CommandLogin />
        <button type="button" className="curtain-close" onClick={() => setCurtain(0)}>▼ إنزال الستارة</button>
      </div>
    </main>
  )
}
