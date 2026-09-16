import { useCallback, useEffect, useRef, useState } from 'react'
import { api } from '../api'
import CommandLogin from '../command/CommandLogin'
import '../command/theme.css'
import { useSession } from '../session'

// ═══ شاشة الدخول ═══
//
// «اريد واجهة مضيئة قويه… وأشياء متحركة ورياكشن… بالنصف… والشعار
// واضح وجميل».
//
// ⚠️ «رياكشن» يعني الشاشة **ترد** على المستخدم مو بس تتحرك لحالها:
// البطاقة تميل ورا الماوس، وضوء يتبع المؤشر، والزر ينبض بالضغط،
// والخانة ترتفع تسميتها لمن تكتب بيها. الحركة الي ما ترد تصير
// خلفية متحركة — والمستخدم يمل منها بأسبوع.
//
// ⚠️ وكل الحركة `transform`/`opacity` بس، والميلان ينحسب برسمة
// وحدة (`requestAnimationFrame`) ويكتب **متغيّرات CSS** بدل ما
// يعيد رسم React — تحريك الحالة مع كل حركة ماوس يخنق الصفحة.
//
// ⚠️ والميلان ينطفي باللمس: التلفون ماكو بيه مؤشر، وحساب الميلان
// من اللمسة يخلّي البطاقة تنط تحت الإصبع.

const UserIcon = () => <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/></svg>
const LockIcon = () => <svg viewBox="0 0 24 24" aria-hidden="true"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/></svg>

export default function Login() {
  const { setEmployee } = useSession()
  const [username, setUsername] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [showPass, setShowPass] = useState(false)
  const [sessionNote] = useState(() => sessionStorage.getItem('sessionEndedReason'))
  const [curtain, setCurtain] = useState(0)
  const dragRef = useRef<{ startY: number; from: number } | null>(null)
  const [dragging, setDragging] = useState(false)

  const cardRef = useRef<HTMLDivElement>(null)
  const frameRef = useRef(0)

  useEffect(() => {
    sessionStorage.removeItem('sessionEndedReason')
    // ⚠️ «تذكرني» انشالت بطلب صاحب النظام — ننضّف الاسم المحفوظ من
    // الأجهزة الي چانت مأشّرة عليها، وإلا يبقى مخزوناً للأبد بلا
    // خانة تشيله.
    localStorage.removeItem('rememberedUser')
  }, [])

  // ═══ الميلان وضوء المؤشر ═══
  const onPointerMove = useCallback((e: React.PointerEvent<HTMLDivElement>) => {
    if (e.pointerType !== 'mouse') return
    const card = cardRef.current
    if (!card) return
    const rect = card.getBoundingClientRect()
    const x = (e.clientX - rect.left) / rect.width
    const y = (e.clientY - rect.top) / rect.height
    cancelAnimationFrame(frameRef.current)
    frameRef.current = requestAnimationFrame(() => {
      card.style.setProperty('--rx', `${(0.5 - y) * 7}deg`)
      card.style.setProperty('--ry', `${(x - 0.5) * 9}deg`)
      card.style.setProperty('--mx', `${x * 100}%`)
      card.style.setProperty('--my', `${y * 100}%`)
    })
  }, [])
  const onPointerLeave = useCallback(() => {
    const card = cardRef.current
    if (!card) return
    cancelAnimationFrame(frameRef.current)
    card.style.setProperty('--rx', '0deg')
    card.style.setProperty('--ry', '0deg')
  }, [])
  useEffect(() => () => cancelAnimationFrame(frameRef.current), [])

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
      setEmployee(await api.login(username, password))
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'حدث خطأ غير متوقع')
    } finally { setSubmitting(false) }
  }

  return (
    <main className="lx" dir="rtl" onPointerMove={onPointerMove} onPointerLeave={onPointerLeave}>
      {/* ═══ الخلفية المضيئة ═══ كتل لون تتموّج + شبكة ناعمة */}
      <div className="lx-sky" aria-hidden>
        <span className="lx-blob lx-b1" />
        <span className="lx-blob lx-b2" />
        <span className="lx-blob lx-b3" />
        <span className="lx-grid" />
        <span className="lx-floor" />
        <div className="lx-rings"><i /><i /><i /><i /></div>
        <span className="lx-beam" />
        <span className="lx-beam lx-beam-2" />
        {[...Array(14)].map((_, i) => (
          <span key={i} className="lx-spark" style={{ ['--n' as string]: i }} />
        ))}
      </div>

      <div className="lx-card" ref={cardRef}>
        <span className="lx-glow" aria-hidden />

        {/* ═══ الشعار ═══ حلقة متدرّجة تدور حوله وهالة تنبض */}
        <div className="lx-logo">
          <span className="lx-logo-ring" aria-hidden />
          <span className="lx-logo-pulse" aria-hidden />
          <span className="lx-logo-disc">
            <img src={`${import.meta.env.BASE_URL}favicon.png?v=3`} alt="شعار شركة الأماني" />
          </span>
        </div>

        <h1 className="lx-title">شركة الأماني</h1>
        <p className="lx-tag">نظام الإدارة المتكامل</p>

        <form className="lx-form" onSubmit={handleSubmit}>
          {/* ⚠️ التسمية ترتفع فوگ الخانة لمن تكتب — «رياكشن» يخلّي
              الموظف يعرف وين هو بلا ما يدوّر. */}
          <div className={`lx-field ${username ? 'has-val' : ''}`}>
            <UserIcon />
            <input
              id="lx-user" required dir="ltr" autoCapitalize="off" autoCorrect="off"
              spellCheck={false} autoComplete="username" placeholder=" "
              value={username} onChange={(e) => setUsername(e.target.value)}
            />
            <label htmlFor="lx-user">اسم المستخدم</label>
            <span className="lx-underline" aria-hidden />
          </div>

          <div className={`lx-field ${password ? 'has-val' : ''}`}>
            <LockIcon />
            <input
              id="lx-pass" required type={showPass ? 'text' : 'password'} dir="ltr"
              autoComplete="current-password" placeholder=" "
              value={password} onChange={(e) => setPassword(e.target.value)}
            />
            <label htmlFor="lx-pass">كلمة المرور</label>
            <span className="lx-underline" aria-hidden />
            {/* ⚠️ إظهار كلمة المرور: الفني يكتبها بتلفونه بالميدان بإيد
                وحدة، والغلط المتكرر يقفل الحساب. */}
            <button
              type="button" className="lx-eye"
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

          <button type="submit" disabled={submitting} className="lx-submit">
            <span className="lx-submit-bg" aria-hidden />
            <span className="lx-submit-txt">
              {submitting
                ? <><span className="lx-spin" aria-hidden />جاري الدخول…</>
                : <>تسجيل الدخول
                    <svg viewBox="0 0 24 24" aria-hidden="true" className="lx-arrow">
                      <path d="M19 12H5M12 19l-7-7 7-7"/>
                    </svg>
                  </>}
            </span>
          </button>

          {(sessionNote || error) && (
            <p className={`lx-msg ${error ? 'is-error' : ''}`} role="alert">{error || sessionNote}</p>
          )}
        </form>

        <p className="lx-secure">
          <svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/></svg>
          دخول آمن ومحمي
        </p>
      </div>

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
