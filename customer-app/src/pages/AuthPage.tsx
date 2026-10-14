import { useEffect, useRef, useState, type FormEvent, type ReactNode } from 'react'
import { api, ApiError, type Account } from '../api'
import Logo from '../components/Logo'
import HeroScene from '../components/HeroScene'

type Step = 'splash' | 'login' | 'signup' | 'verify' | 'forgot' | 'reset'

// ═══ البداية + الدخول + إنشاء الحساب (شاشة 1 و2 بالتصميم) ═══
// الدخول باسم المستخدم. الإيميل للتأكيد بس (رمز 6 أرقام).
export default function AuthPage({ onSession }: { onSession: (token: string, a: Account) => void }) {
  const [step, setStep] = useState<Step>('splash')
  const [f, setF] = useState({ name: '', username: '', email: '', phone: '', password: '', code: '' })
  const [masked, setMasked] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [info, setInfo] = useState('')
  const [cool, setCool] = useState(0)
  const [showPw, setShowPw] = useState(false)

  useEffect(() => { if (step === 'splash') { const t = setTimeout(() => setStep('login'), 2600); return () => clearTimeout(t) } }, [step])
  useEffect(() => { if (cool > 0) { const t = setTimeout(() => setCool((c) => c - 1), 1000); return () => clearTimeout(t) } }, [cool])

  const set = (k: keyof typeof f) => (e: { target: { value: string } }) => setF((x) => ({ ...x, [k]: e.target.value }))
  const go = (s: Step) => { setStep(s); setErr(''); setInfo('') }
  const run = async (fn: () => Promise<void>) => {
    setBusy(true); setErr(''); setInfo('')
    try { await fn() } catch (e) {
      if (e instanceof ApiError && e.data.needVerify) { setMasked(String(e.data.email || '')); setCool(60); go('verify'); setInfo(e.message); return }
      setErr(e instanceof Error ? e.message : 'صار خطأ')
    } finally { setBusy(false) }
  }
  const submit = (e: FormEvent) => {
    e.preventDefault()
    if (step === 'login') return run(async () => { const s = await api.login(f.username, f.password); onSession(s.token, s.account) })
    if (step === 'signup') return run(async () => {
      const r = await api.signup({ name: f.name, username: f.username, email: f.email, phone: f.phone, password: f.password })
      setF((x) => ({ ...x, username: r.username, code: '' })); setMasked(r.email); setCool(60); go('verify')
    })
    if (step === 'verify') return run(async () => { const s = await api.verify(f.username, f.code); onSession(s.token, s.account) })
    if (step === 'forgot') return run(async () => { const r = await api.forgot(f.username); setMasked(r.email); setCool(60); setF((x) => ({ ...x, code: '', password: '' })); go('reset') })
    if (step === 'reset') return run(async () => { const s = await api.reset(f.username, f.code, f.password); onSession(s.token, s.account) })
  }
  const resend = () => run(async () => { const r = await api.resend(f.username, step === 'reset' ? 'reset' : 'signup'); setMasked(r.email); setCool(60); setInfo('دزّينا رمز جديد على إيميلك ✉️') })

  if (step === 'splash') {
    return (
      <div role="button" tabIndex={0} onClick={() => go('login')} className="relative h-full w-full overflow-hidden bg-navy-deep">
        <HeroScene className="kenburns absolute inset-0 h-full w-full" />
        <div className="absolute inset-0 bg-gradient-to-b from-navy-deep/90 via-transparent via-45% to-transparent" />
        <div className="rise relative flex flex-col items-center px-6 pt-[11vh] text-center">
          <Logo light size={96} />
          <p className="mt-5 text-[17px] font-bold text-white/95">حلول تقنية .. لحياة أكثر أماناً</p>
        </div>
        <div className="absolute inset-x-0 bottom-10 flex justify-center gap-1.5">
          <span className="h-1.5 w-8 rounded-full bg-white" /><span className="h-1.5 w-8 rounded-full bg-white/35" /><span className="h-1.5 w-8 rounded-full bg-white/35" />
        </div>
      </div>
    )
  }

  const title: Record<Exclude<Step, 'splash'>, [string, string]> = {
    login: ['', 'خدمات تقنية وهندسية متكاملة\nلبيوت وأعمال أكثر أماناً'],
    signup: ['إنشاء حساب', 'دقيقة وحدة وتصير وية عائلة الأماني'],
    verify: ['تأكيد البريد الإلكتروني', `أرسلنا رمز من 6 أرقام إلى\n${masked}`],
    forgot: ['نسيت كلمة المرور؟', 'اكتب اسم المستخدم ونرسل رمز على بريدك'],
    reset: ['كلمة مرور جديدة', `اكتب الرمز المرسل إلى ${masked}`],
  }

  return (
    <div className="page-in min-h-full bg-white">
      <div className="mx-auto flex min-h-full max-w-md flex-col px-6 pt-12 pb-8">
        {step !== 'login' && (
          <button onClick={() => go(step === 'reset' ? 'forgot' : 'login')} aria-label="رجوع" className="mb-2 -mr-2 grid h-10 w-10 place-items-center rounded-full text-ink active:bg-slate-100">
            <svg width="22" height="22" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.2" strokeLinecap="round"><path d="M9 6l6 6-6 6" /></svg>
          </button>
        )}
        <div className={step === 'login' ? 'mt-6 mb-8' : 'mb-6'}>
          {step === 'login' ? <Logo size={92} /> : <h1 className="text-[24px] font-black text-ink">{title[step][0]}</h1>}
          <p className={`mt-3 text-[14px] leading-7 whitespace-pre-line text-mute ${step === 'login' ? 'text-center font-semibold' : ''}`}>{title[step][1]}</p>
        </div>

        <form onSubmit={submit} className="space-y-3.5">
          {step === 'signup' && <>
            <Input icon={ICON.user} placeholder="الاسم الكامل" value={f.name} onChange={set('name')} autoComplete="name" />
            <Input icon={ICON.phone} placeholder="رقم الهاتف 07xxxxxxxxx" inputMode="tel" dir="ltr" value={f.phone} onChange={set('phone')} autoComplete="tel" />
            <Input icon={ICON.mail} placeholder="البريد الإلكتروني (للتأكيد)" type="email" dir="ltr" value={f.email} onChange={set('email')} autoComplete="email" />
          </>}
          {(step === 'login' || step === 'signup' || step === 'forgot') && (
            <Input icon={ICON.at} placeholder="اسم المستخدم" dir="ltr" autoCapitalize="none" value={f.username}
              onChange={(e) => setF((x) => ({ ...x, username: e.target.value.toLowerCase().replace(/\s/g, '') }))} autoComplete="username"
              hint={step === 'signup' ? 'حروف إنكليزية صغيرة وأرقام، مثل ali.ahmed' : undefined} />
          )}
          {(step === 'verify' || step === 'reset') && <OtpBoxes value={f.code} onChange={(code) => setF((x) => ({ ...x, code }))} />}
          {(step === 'login' || step === 'signup' || step === 'reset') && (
            <Input icon={ICON.lock} placeholder={step === 'reset' ? 'كلمة المرور الجديدة' : 'كلمة المرور'} type={showPw ? 'text' : 'password'} dir="ltr"
              value={f.password} onChange={set('password')} autoComplete={step === 'login' ? 'current-password' : 'new-password'}
              hint={step !== 'login' ? '8 أحرف على الأقل' : undefined}
              end={<button type="button" onClick={() => setShowPw((v) => !v)} aria-label="إظهار كلمة المرور" className="text-mute">{showPw ? ICON.eyeOff : ICON.eye}</button>} />
          )}

          {step === 'login' && (
            <div className="flex items-center justify-between pt-1 text-[13px]">
              <label className="flex items-center gap-2 text-mute"><input type="checkbox" defaultChecked className="h-4 w-4 accent-[#0d3b8c]" /> تذكرني</label>
              <button type="button" onClick={() => go('forgot')} className="font-bold text-navy">نسيت كلمة المرور؟</button>
            </div>
          )}

          {err && <p className="rounded-xl bg-rose-50 px-3 py-2.5 text-[13px] font-bold text-rose-600">{err}</p>}
          {info && <p className="rounded-xl bg-emerald-50 px-3 py-2.5 text-[13px] font-bold text-emerald-700">{info}</p>}

          <button disabled={busy || (step === 'verify' || step === 'reset' ? f.code.length !== 6 : false)}
            className="mt-2 h-[52px] w-full rounded-xl bg-navy-2 text-[16px] font-extrabold text-white shadow-[0_10px_20px_-10px_rgba(13,59,140,.7)] transition active:scale-[.98] disabled:opacity-50">
            {busy ? <span className="inline-flex gap-1">{[0, 1, 2].map((i) => <span key={i} className="dot h-2 w-2 rounded-full bg-white" style={{ animationDelay: `${i * .2}s` }} />)}</span>
              : { login: 'تسجيل الدخول', signup: 'إنشاء الحساب', verify: 'تأكيد', forgot: 'إرسال الرمز', reset: 'حفظ ودخول' }[step]}
          </button>

          {(step === 'verify' || step === 'reset') && (
            <button type="button" disabled={cool > 0 || busy} onClick={resend} className="w-full py-2 text-[13px] font-bold text-navy disabled:text-mute">
              {cool > 0 ? `إعادة الإرسال بعد 0:${String(cool).padStart(2, '0')}` : 'ما وصلني الرمز؟ إعادة الإرسال'}
            </button>
          )}
        </form>

        <div className="mt-auto pt-10 text-center text-[14px]">
          {step === 'login' && <p className="text-mute">ليس لديك حساب؟ <button onClick={() => go('signup')} className="font-extrabold text-navy">إنشاء حساب</button></p>}
          {step === 'signup' && <p className="text-mute">عندك حساب؟ <button onClick={() => go('login')} className="font-extrabold text-navy">تسجيل الدخول</button></p>}
        </div>
      </div>
    </div>
  )
}

function OtpBoxes({ value, onChange }: { value: string; onChange: (v: string) => void }) {
  const ref = useRef<HTMLInputElement>(null)
  return (
    <div className="relative" onClick={() => ref.current?.focus()}>
      <input ref={ref} autoFocus inputMode="numeric" autoComplete="one-time-code" maxLength={6} value={value}
        onChange={(e) => onChange(e.target.value.replace(/\D/g, '').slice(0, 6))} className="absolute inset-0 opacity-0" aria-label="رمز التحقق" />
      <div className="flex justify-between gap-2" dir="ltr">
        {[0, 1, 2, 3, 4, 5].map((i) => (
          <span key={i} className={`grid h-14 flex-1 place-items-center rounded-xl border-2 text-2xl font-black text-ink transition ${
            value.length === i ? 'border-navy bg-white' : value[i] ? 'border-navy/30 bg-blue-50/50' : 'border-line bg-[#f7f9fc]'}`}>{value[i] ?? ''}</span>
        ))}
      </div>
    </div>
  )
}

function Input({ icon, hint, end, ...p }: React.InputHTMLAttributes<HTMLInputElement> & { icon: ReactNode; hint?: string; end?: ReactNode }) {
  return (
    <label className="block">
      <span className="flex h-[52px] items-center gap-3 rounded-xl border border-line bg-white px-4 transition focus-within:border-navy focus-within:ring-4 focus-within:ring-navy/10">
        <span className="text-mute">{icon}</span>
        <input required {...p} className="h-full min-w-0 flex-1 bg-transparent text-[15px] text-ink outline-none placeholder:text-right placeholder:text-[#a3acbf]" />
        {end}
      </span>
      {hint && <span className="mt-1 block px-1 text-[11px] text-mute">{hint}</span>}
    </label>
  )
}

const i = (d: ReactNode) => <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round">{d}</svg>
const ICON = {
  user: i(<><circle cx="12" cy="8" r="4" /><path d="M4 21a8 8 0 0 1 16 0" /></>),
  phone: i(<path d="M22 16.9v3a2 2 0 0 1-2.2 2 19.8 19.8 0 0 1-8.6-3.1 19.5 19.5 0 0 1-6-6A19.8 19.8 0 0 1 2.1 4.2 2 2 0 0 1 4.1 2h3a2 2 0 0 1 2 1.7c.1.9.4 1.8.7 2.7a2 2 0 0 1-.5 2.1L8 9.8a16 16 0 0 0 6 6l1.3-1.3a2 2 0 0 1 2.1-.4c.9.3 1.8.6 2.7.7a2 2 0 0 1 1.7 2z" />),
  mail: i(<><rect x="3" y="5" width="18" height="14" rx="2" /><path d="M3 7l9 6 9-6" /></>),
  at: i(<><circle cx="12" cy="12" r="4" /><path d="M16 8v5a3 3 0 0 0 6 0v-1a10 10 0 1 0-4 8" /></>),
  lock: i(<><rect x="4" y="11" width="16" height="10" rx="2" /><path d="M8 11V7a4 4 0 0 1 8 0v4" /></>),
  eye: i(<><path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7S2 12 2 12z" /><circle cx="12" cy="12" r="3" /></>),
  eyeOff: i(<><path d="M17.9 17.9A10 10 0 0 1 12 19c-6 0-10-7-10-7a18 18 0 0 1 5.1-5.9M9.9 4.2A9 9 0 0 1 12 4c6 0 10 7 10 7a18 18 0 0 1-2.2 3.2M1 1l22 22" /></>),
}
