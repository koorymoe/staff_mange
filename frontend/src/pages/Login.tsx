import { useEffect, useRef, useState } from 'react'
import { api } from '../api'
import CommandLogin from '../command/CommandLogin'
import '../command/theme.css'
import { useSession } from '../session'

// ═══ شاشة الدخول ═══
//
// «نريد واجهة تسجيل دخول قوية وحلوة وبيها أشياء لايف».
//
// ⚠️ كل الحركة **بالكود مو بصورة**: الصورة ما تتحرك أصلاً، وفوق
// هذا النسخة القديمة چانت خلفية ٢ ميغا تنزّل بكل فتحة — والفني
// بالميدان يدفعها من رصيده. هاي الشاشة وزنها صفر بايت.
//
// ⚠️ والحركة كلها `transform` و`opacity` بس — هذنه الوحيدتان الي
// المتصفح يحرّكهن على كارت الشاشة بلا ما يعيد حساب تخطيط الصفحة.
// أي حركة على `top` أو `width` تخلّي الشاشة تلعثم بالأجهزة الواطية،
// والموظف يفتح هالشاشة كل يوم.
//
// ⚠️ وتخف بالموبايل وتسكت لمن الجهاز يطلب تقليل الحركة — شوف CSS.

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
    <main className="lg-stage" dir="rtl">
      {/* ═══ الخلفية الحيّة ═══ كلها زينة، فمخفية عن قارئ الشاشة */}
      <div className="lg-sky" aria-hidden>
        <span className="lg-aurora lg-aurora-1" />
        <span className="lg-aurora lg-aurora-2" />
        <span className="lg-aurora lg-aurora-3" />
        <span className="lg-floor" />
        <span className="lg-scan" />
        <span className="lg-rays" />
        <span className="lg-halo" />
        <span className="lg-sat" />
        {/* نجوم/غبار ضوئي — مواقعها ثابتة حتى ما تختلف بين رسمة وأخرى */}
        {[...Array(18)].map((_, i) => (
          <span key={i} className="lg-dust" style={{ ['--n' as string]: i }} />
        ))}
      </div>

      <section className="lg-card" aria-labelledby="lg-title">
        <span className="lg-card-ring" aria-hidden />

        <div className="lg-logo">
          <span className="lg-orbit lg-orbit-1" aria-hidden />
          <span className="lg-orbit lg-orbit-2" aria-hidden />
          <span className="lg-radar" aria-hidden />
          <img src={`${import.meta.env.BASE_URL}favicon.png?v=3`} alt="شعار شركة الأماني" />
        </div>

        <h1 id="lg-title" className="lg-brand">شركة الأماني</h1>
        <p className="lg-sub"><span />نظام الإدارة المتكامل<span /></p>

        <form className="lg-form" onSubmit={handleSubmit}>
          <label htmlFor="lg-user">اسم المستخدم</label>
          <div className="lg-field">
            <UserIcon />
            <input
              id="lg-user" required dir="ltr" autoCapitalize="off" autoCorrect="off"
              spellCheck={false} autoComplete="username" placeholder="أدخل اسم المستخدم"
              value={username} onChange={(e) => setUsername(e.target.value)}
            />
          </div>

          <label htmlFor="lg-pass">كلمة المرور</label>
          <div className="lg-field">
            <LockIcon />
            <input
              id="lg-pass" required type={showPass ? 'text' : 'password'} dir="ltr"
              autoComplete="current-password" placeholder="أدخل كلمة المرور"
              value={password} onChange={(e) => setPassword(e.target.value)}
            />
            {/* ⚠️ إظهار كلمة المرور: الفني يكتبها بتلفونه بالميدان بإيد
                وحدة، والغلط المتكرر يقفل الحساب. */}
            <button
              type="button" className="lg-eye"
              onClick={() => setShowPass((v) => !v)}
              aria-label={showPass ? 'إخفاء كلمة المرور' : 'إظهار كلمة المرور'}
            >
              <svg viewBox="0 0 24 24" aria-hidden="true">
                {showPass
                  ? <path d="M17.94 17.94A10.1 10.1 0 0 1 12 20c-7 0-11-8-11-8a18.5 18.5 0 0 1 5.06-5.94M9.9 4.24A9 9 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24M1 1l22 22"/>
                  : <><path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/></>}
              </svg>
            </button>
          </div>

          <div className="lg-row">
            {/* ⚠️ «تذكرني» تحفظ **الاسم بس** مو كلمة المرور: حفظ الكلمة
                بالجهاز يعني أي واحد يفتح تلفون الموظف يدخل باسمه —
                والنظام فيه فلوس وصلاحيات. */}
            <label className="lg-remember">
              <input type="checkbox" checked={remember} onChange={(e) => setRemember(e.target.checked)} />
              <span className="lg-box" aria-hidden />
              تذكرني
            </label>
            <button
              type="button" className="lg-forgot"
              onClick={() => setError('كلمة المرور تنرجع من مدير النظام — راجعه وهو يعيد ضبطها بحسابك.')}
            >
              نسيت كلمة المرور؟
            </button>
          </div>

          <button type="submit" disabled={submitting} className="lg-submit">
            <span className="lg-submit-sweep" aria-hidden />
            {submitting
              ? <><span className="lg-spin" aria-hidden />جاري الدخول…</>
              : <>تسجيل الدخول
                  <svg viewBox="0 0 24 24" aria-hidden="true" className="lg-arrow">
                    <path d="M19 12H5M12 19l-7-7 7-7"/>
                  </svg>
                </>}
          </button>

          {(sessionNote || error) && (
            <p className={`lg-msg ${error ? 'is-error' : ''}`} role="alert">{error || sessionNote}</p>
          )}
        </form>

        <p className="lg-secure">
          <span />
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
          دخول آمن ومحمي
          <span />
        </p>
      </section>

      {/* قاعدة ضوئية تحت البطاقة */}
      <div className="lg-base" aria-hidden><i /><i /><i /></div>

      {/* ═══ الستارة ═══ من ركن أسفل الشاشة تنسحب وتكشف دخول مركز القيادة.
          ⚠️ الشاشتان تنادان **نفس** api.login — الي يقرر لأي نظام تدخل
          هو الرمز مو الشاشة. ماكو طريق دخول ثاني ينضاف. */}
      <button
        type="button" aria-label="ارفع الستارة" className="curtain-handle"
        onPointerDown={onHandleDown} onPointerMove={onHandleMove}
        onPointerUp={onHandleUp} onPointerCancel={onHandleUp}
        onClick={() => setCurtain((v) => (v === 0 ? 1 : 0))}
      />
      {/* ⚠️ inert وهي مسدولة: بدونه حقول الدخول تبقى بالصفحة، والـTab
          ينط عليها وقارئ الشاشة يقراها — نموذج دخول مخفي بالعين بس شغّال. */}
      <div
        className="curtain"
        inert={curtain === 0 ? true : undefined}
        style={{
          transform: `translateY(${(1 - curtain) * 100}%)`,
          transition: dragging ? 'none' : 'transform 0.62s cubic-bezier(0.22, 1, 0.36, 1)',
        }}
      >
        <CommandLogin />
        <button type="button" className="curtain-close" onClick={() => setCurtain(0)}>▼ إنزال الستارة</button>
      </div>
    </main>
  )
}
