import { useState, useEffect, useRef } from 'react'
import { api } from '../api'
import { useSession } from '../session'
import CommandLogin from '../command/CommandLogin'
import { LEFT_SERVICES, RIGHT_SERVICES } from '../utils/loginServices'
import ServiceCard from '../components/ServiceCard'
import '../command/theme.css'

// ⚠️ الكرات صارن **نقاط** صغيرة: بالتصميم القديم كانن كور زرقا
// كبيرة (٢٦٠px) تزاحم البطاقة وتشد النظر عنها. الخلفية دورها تحيط
// بالبطاقة مو تنافسها.
const dots = [
  { size: 4, top: '12%', left: '18%', delay: '0s' },
  { size: 3, top: '28%', left: '9%', delay: '1.4s' },
  { size: 5, top: '68%', left: '14%', delay: '2.6s' },
  { size: 3, top: '84%', left: '26%', delay: '0.8s' },
  { size: 4, top: '16%', left: '78%', delay: '2.1s' },
  { size: 3, top: '46%', left: '92%', delay: '1.1s' },
  { size: 5, top: '74%', left: '84%', delay: '3.1s' },
  { size: 3, top: '8%', left: '58%', delay: '2.4s' },
  { size: 4, top: '90%', left: '62%', delay: '1.7s' },
]

export default function Login() {
  const { setEmployee } = useSession()
  const [username, setUsername] = useState(() => localStorage.getItem('rememberedUser') || '')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)
  const [showPass, setShowPass] = useState(false)
  // ⚠️ «تذكرني» تحفظ الاسم بس — شوف التعليق عند الخانة.
  const [remember, setRemember] = useState(() => !!localStorage.getItem('rememberedUser'))
  // سبب انتهاء الجلسة يوصل من api.handleSessionExpired عبر sessionStorage —
  // بدل alert الي كان يوقف الصفحة وينتظر ضغطة (وينفتح عدة مرات بنفس الوقت).
  // قراءة صافية بلا حذف: StrictMode ينفّذ مُهيّئ useState مرتين، فلو حذفنا
  // هنا راح تطلع القراءة الثانية فاضية وتضيع الرسالة. الحذف بـuseEffect.
  const [sessionNote] = useState(() => sessionStorage.getItem('sessionEndedReason'))
  useEffect(() => { sessionStorage.removeItem('sessionEndedReason') }, [])

  // ═══ الستارة ═══
  //
  // من ركن أسفل الشاشة تنسحب ستارة وتكشف دخول مركز القيادة.
  //
  // ⚠️ الشاشتين تنادن **نفس** `api.login`. الي يقرر لأي نظام تدخل هو
  // الرمز مو الشاشة: يوزرك + رمزك العادي من الشاشة الحمراء يدخّلك
  // نظام الشركة، ويوزرك + رمز القيادة من الشاشة الزرقاء يدخّلك مركز
  // القيادة. ماكو طريق دخول ثاني ينضاف — بس منظر ثاني لنفس الطريق.
  //
  // 0 = مسدولة، 1 = مرفوعة كلها. الكسر بينهن للسحب بالإصبع.
  const [curtain, setCurtain] = useState(0)
  const dragRef = useRef<{ startY: number; from: number } | null>(null)
  const [dragging, setDragging] = useState(false)

  const onHandleDown = (e: React.PointerEvent<HTMLButtonElement>) => {
    e.currentTarget.setPointerCapture(e.pointerId)
    dragRef.current = { startY: e.clientY, from: curtain }
    setDragging(true)
  }
  const onHandleMove = (e: React.PointerEvent<HTMLButtonElement>) => {
    const d = dragRef.current
    if (!d) return
    // السحب لفوق = قيمة أكبر. نقسم على ارتفاع الشاشة حتى السحب يمشي
    // وية الإصبع بنفس المسافة مهما كان حجم الجهاز.
    const next = d.from + (d.startY - e.clientY) / window.innerHeight
    setCurtain(Math.min(1, Math.max(0, next)))
  }
  const onHandleUp = () => {
    if (!dragRef.current) return
    dragRef.current = null
    setDragging(false)
    // تلتصق بالطرفين: نص ستارة مرفوعة ما إلها معنى
    setCurtain((c) => (c > 0.3 ? 1 : 0))
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    setError(null)
    try {
      const employee = await api.login(username, password)
      if (remember) localStorage.setItem('rememberedUser', username)
      else localStorage.removeItem('rememberedUser')
      setEmployee(employee)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="login-stage relative flex min-h-screen items-center justify-center overflow-hidden p-4">
      {/* ═══ الخلفية ═══ شبكة وتوهّج، بلا أي صورة تنحمّل */}
      <div className="login-grid" aria-hidden />
      {dots.map((o, i) => (
        <span
          key={i}
          className="login-dot"
          style={{ width: o.size, height: o.size, top: o.top, left: o.left, animationDelay: o.delay }}
        />
      ))}

      {/* ═══ خدمات الشركة حول البطاقة ═══
          ⚠️ تنخفي تحت 1280px: هاي زينة تحيط ببطاقة الدخول، وبالموبايل
          تصير جدار نصوص يدفن الخانات الي جاي عشانها الموظف. */}
      <div className="pointer-events-none absolute right-6 top-1/2 z-10 hidden -translate-y-1/2 flex-col gap-7 xl:flex 2xl:right-16">
        {RIGHT_SERVICES.map((it) => <ServiceCard key={it.title} item={it} side="right" />)}
      </div>
      <div className="pointer-events-none absolute left-6 top-1/2 z-10 hidden -translate-y-1/2 flex-col gap-7 xl:flex 2xl:left-16">
        {LEFT_SERVICES.map((it) => <ServiceCard key={it.title} item={it} side="left" />)}
      </div>

      <form
        onSubmit={handleSubmit}
        className="login-card relative z-20 flex w-full max-w-[420px] flex-col items-center rounded-[2.2rem] px-7 py-10 sm:px-9"
      >
        <div className="login-logo flex h-20 w-20 items-center justify-center rounded-3xl">
          <img src={`${import.meta.env.BASE_URL}favicon.png?v=3`} alt="شعار شركة الأماني" className="h-12 w-12 object-contain" />
        </div>

        <h1 className="mt-5 text-center text-3xl font-black tracking-tight text-white">شركة الأماني</h1>
        {/* السطر الفرعي بين خطين — مثل التصميم */}
        <div className="mt-2 flex w-full items-center gap-3">
          <span className="h-px flex-1 bg-gradient-to-l from-transparent to-sky-400/40" />
          <p className="whitespace-nowrap text-[13px] text-sky-200/80">نظام الإدارة المتكامل</p>
          <span className="h-px flex-1 bg-gradient-to-r from-transparent to-sky-400/40" />
        </div>

        <div className="mt-7 w-full">
          <label className="mb-1.5 block text-[13px] font-medium text-sky-100/80">اسم المستخدم</label>
          <div className="relative">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
              className="pointer-events-none absolute right-3.5 top-1/2 -translate-y-1/2 text-sky-300/70">
              <path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>
            </svg>
            <input
              required
              dir="ltr"
              autoCapitalize="off"
              autoCorrect="off"
              spellCheck={false}
              placeholder="أدخل اسم المستخدم"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="login-field w-full rounded-xl py-3 pl-3 pr-11 text-right text-sm outline-none"
            />
          </div>
        </div>

        <div className="mt-4 w-full">
          <label className="mb-1.5 block text-[13px] font-medium text-sky-100/80">كلمة المرور</label>
          <div className="relative">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"
              className="pointer-events-none absolute right-3.5 top-1/2 -translate-y-1/2 text-sky-300/70">
              <rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0 1 10 0v4"/>
            </svg>
            <input
              required
              type={showPass ? 'text' : 'password'}
              dir="ltr"
              placeholder="أدخل كلمة المرور"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="login-field w-full rounded-xl py-3 pl-11 pr-11 text-right text-sm outline-none"
            />
            {/* ⚠️ إظهار كلمة المرور: الموظف يكتبها بتلفونه بالميدان
                وبإيد وحدة — والغلط المتكرر يقفل الحساب. */}
            <button
              type="button"
              onClick={() => setShowPass((v) => !v)}
              aria-label={showPass ? 'إخفاء كلمة المرور' : 'إظهار كلمة المرور'}
              className="absolute left-3 top-1/2 -translate-y-1/2 text-sky-300/70 transition hover:text-sky-200"
            >
              {showPass ? (
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M17.94 17.94A10.07 10.07 0 0 1 12 20c-7 0-11-8-11-8a18.45 18.45 0 0 1 5.06-5.94M9.9 4.24A9.12 9.12 0 0 1 12 4c7 0 11 8 11 8a18.5 18.5 0 0 1-2.16 3.19m-6.72-1.07a3 3 0 1 1-4.24-4.24"/><line x1="1" y1="1" x2="23" y2="23"/>
                </svg>
              ) : (
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
                  <path d="M1 12s4-8 11-8 11 8 11 8-4 8-11 8-11-8-11-8z"/><circle cx="12" cy="12" r="3"/>
                </svg>
              )}
            </button>
          </div>
        </div>

        <div className="mt-3.5 flex w-full items-center justify-between">
          <label className="flex cursor-pointer items-center gap-2 text-[12px] text-sky-100/80">
            {/* ⚠️ «تذكرني» تحفظ **الاسم بس** مو كلمة المرور: حفظ
                الكلمة بالجهاز يعني أي واحد يفتح تلفون الموظف يدخل
                باسمه — والنظام فيه فلوس وصلاحيات. */}
            <input
              type="checkbox"
              checked={remember}
              onChange={(e) => setRemember(e.target.checked)}
              className="h-4 w-4 accent-sky-500"
            />
            تذكرني
          </label>
          <button
            type="button"
            onClick={() => setError('كلمة المرور تنرجع من مدير النظام — راجعه وهو يعيد ضبطها بحسابك.')}
            className="text-[12px] text-sky-300/80 underline-offset-4 transition hover:text-sky-200 hover:underline"
          >
            نسيت كلمة المرور؟
          </button>
        </div>

        {sessionNote && !error && (
          <div className="mt-4 w-full rounded-xl bg-amber-400/15 px-4 py-3 text-center text-[13px] font-bold text-amber-100 ring-1 ring-amber-300/30">
            {sessionNote}
          </div>
        )}
        {error && (
          <p className="mt-4 w-full rounded-xl bg-red-500/15 p-3 text-center text-[13px] text-red-100 ring-1 ring-red-400/30">
            {error}
          </p>
        )}

        <button
          type="submit"
          disabled={submitting}
          className="login-submit mt-6 flex w-full items-center justify-center gap-2 rounded-xl px-6 py-3 text-base font-extrabold text-white disabled:opacity-60"
        >
          {submitting ? 'جاري الدخول...' : 'تسجيل الدخول'}
          {!submitting && (
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round">
              <line x1="19" y1="12" x2="5" y2="12"/><polyline points="12 19 5 12 12 5"/>
            </svg>
          )}
        </button>

        <div className="mt-5 flex w-full items-center gap-3">
          <span className="h-px flex-1 bg-white/10" />
          <p className="flex items-center gap-1.5 whitespace-nowrap text-[11px] text-sky-200/60">
            <svg width="13" height="13" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
              <path d="M12 22s8-4 8-10V5l-8-3-8 3v7c0 6 8 10 8 10z"/>
            </svg>
            دخول آمن ومحمي
          </p>
          <span className="h-px flex-1 bg-white/10" />
        </div>
      </form>

      {/* مقبض الستارة — ركن أسفل الشاشة */}
      <button
        type="button"
        aria-label="ارفع الستارة"
        className="curtain-handle"
        onPointerDown={onHandleDown}
        onPointerMove={onHandleMove}
        onPointerUp={onHandleUp}
        onPointerCancel={onHandleUp}
        onClick={() => setCurtain((c) => (c === 0 ? 1 : 0))}
      />

      {/* الستارة نفسها.
          محلها الطبيعي برّا الشاشة من تحت، فما تغطي شي وهي مسدولة.
          ⚠️ inert لما تكون مسدولة: بدونه حقول الدخول الحمراء تبقى
          موجودة بالصفحة، والـTab ينط عليها وقارئ الشاشة يقراها —
          يعني نموذج دخول مخفي بالعين بس شغّال فعلياً. */}
      <div
        className="curtain"
        inert={curtain === 0 ? true : undefined}
        style={{
          transform: `translateY(${(1 - curtain) * 100}%)`,
          // بلا انتقال أثناء السحب — وإلا الستارة تتأخر ورا الإصبع
          // ⚠️ نقرا **حالة** مو `ref` أثناء الرندر: قراءة الـref
          // بالرندر ما تعيد الرسم لمن تتغيّر، فالانتقال يبقى معطّلاً
          // بعد ما يخلص السحب.
          transition: dragging ? 'none' : 'transform 0.62s cubic-bezier(0.22, 1, 0.36, 1)',
        }}
      >
        <CommandLogin />
        <button type="button" className="curtain-close" onClick={() => setCurtain(0)}>
          ▼ إنزال الستارة
        </button>
      </div>
    </div>
  )
}
