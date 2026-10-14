import { useEffect, useState, type FormEvent, type ReactNode } from 'react'
import { api, ApiError, type Account } from '../api'
import Logo from '../components/Logo'
import MatrixEye from '../components/MatrixEye'

type Step = 'login' | 'signup' | 'verify' | 'forgot' | 'reset'

// ═══ الدخول وإنشاء الحساب ═══
// الدخول باسم المستخدم + الباسورد. الإيميل للتأكيد بس (رمز 6 أرقام).
export default function AuthPage({ onSession }: { onSession: (token: string, a: Account) => void }) {
  const [step, setStep] = useState<Step>('login')
  const [f, setF] = useState({ name: '', username: '', email: '', phone: '', password: '', code: '' })
  const [maskedEmail, setMaskedEmail] = useState('')
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [info, setInfo] = useState('')
  const [cool, setCool] = useState(0)

  useEffect(() => {
    if (cool <= 0) return
    const t = setTimeout(() => setCool((c) => c - 1), 1000)
    return () => clearTimeout(t)
  }, [cool])

  const set = (k: keyof typeof f) => (e: { target: { value: string } }) => setF((x) => ({ ...x, [k]: e.target.value }))
  const go = (s: Step) => { setStep(s); setErr(''); setInfo('') }

  const run = async (fn: () => Promise<void>) => {
    setBusy(true); setErr(''); setInfo('')
    try { await fn() } catch (e) {
      if (e instanceof ApiError && e.data.needVerify) {
        setMaskedEmail(String(e.data.email || '')); setCool(60); go('verify'); setInfo(e.message); return
      }
      setErr(e instanceof Error ? e.message : 'صار خطأ')
    } finally { setBusy(false) }
  }

  const submit = (e: FormEvent) => {
    e.preventDefault()
    if (step === 'login') return run(async () => { const s = await api.login(f.username, f.password); onSession(s.token, s.account) })
    if (step === 'signup') return run(async () => {
      const r = await api.signup({ name: f.name, username: f.username, email: f.email, phone: f.phone, password: f.password })
      setF((x) => ({ ...x, username: r.username, code: '' })); setMaskedEmail(r.email); setCool(60); go('verify')
    })
    if (step === 'verify') return run(async () => { const s = await api.verify(f.username, f.code); onSession(s.token, s.account) })
    if (step === 'forgot') return run(async () => { const r = await api.forgot(f.username); setMaskedEmail(r.email); setCool(60); setF((x) => ({ ...x, code: '', password: '' })); go('reset') })
    if (step === 'reset') return run(async () => { const s = await api.reset(f.username, f.code, f.password); onSession(s.token, s.account) })
  }

  const resend = () => run(async () => {
    const r = await api.resend(f.username, step === 'reset' ? 'reset' : 'signup')
    setMaskedEmail(r.email); setCool(60); setInfo('دزّينا رمز جديد ✉️')
  })

  const titles: Record<Step, [string, string]> = {
    login: ['أهلاً بيك 👋', 'سجّل دخولك وخلي ماتركس يخدمك'],
    signup: ['حساب جديد', 'دقيقة وحدة وتصير وية عائلة الأماني'],
    verify: ['تأكيد الإيميل', `دزّينا رمز من 6 أرقام على ${maskedEmail}`],
    forgot: ['نسيت كلمة المرور؟', 'اكتب اسم المستخدم ونرسلك رمز على إيميلك'],
    reset: ['كلمة مرور جديدة', `اكتب الرمز الي وصلك على ${maskedEmail}`],
  }

  return (
    <div className="min-h-full bg-gradient-to-b from-navy via-[#1e40af] to-[#f4f6fb] to-55%">
      <div className="mx-auto flex min-h-full max-w-md flex-col px-5 pt-10 pb-8">
        <div className="mb-6 flex items-center gap-3 text-white">
          <Logo size={52} />
          <div>
            <p className="text-xl font-black">الأماني</p>
            <p className="text-xs text-blue-100">للتجارة العامة والاستثمارات والوكالات التجارية</p>
          </div>
          <span className="mr-auto"><MatrixEye size={44} /></span>
        </div>

        <form onSubmit={submit} className="rise rounded-3xl bg-white p-6 shadow-xl shadow-navy/10">
          <h1 className="text-2xl font-black text-ink">{titles[step][0]}</h1>
          <p className="mt-1 mb-5 text-sm text-slate-500">{titles[step][1]}</p>

          {step === 'signup' && <>
            <Field label="الاسم الكامل"><input required value={f.name} onChange={set('name')} className={inp} autoComplete="name" /></Field>
            <Field label="رقم الهاتف"><input required inputMode="tel" dir="ltr" placeholder="07xxxxxxxxx" value={f.phone} onChange={set('phone')} className={inp} autoComplete="tel" /></Field>
            <Field label="الإيميل (للتأكيد بس)"><input required type="email" dir="ltr" value={f.email} onChange={set('email')} className={inp} autoComplete="email" /></Field>
          </>}
          {(step === 'login' || step === 'signup' || step === 'forgot') && (
            <Field label="اسم المستخدم" hint={step === 'signup' ? 'حروف إنكليزية صغيرة وأرقام، مثل ali.ahmed' : undefined}>
              <input required dir="ltr" autoCapitalize="none" autoCorrect="off" value={f.username} onChange={(e) => setF((x) => ({ ...x, username: e.target.value.toLowerCase().replace(/\s/g, '') }))} className={inp} autoComplete="username" />
            </Field>
          )}
          {(step === 'verify' || step === 'reset') && (
            <Field label="رمز التحقق">
              <input required inputMode="numeric" dir="ltr" maxLength={6} autoComplete="one-time-code" value={f.code}
                onChange={(e) => setF((x) => ({ ...x, code: e.target.value.replace(/\D/g, '').slice(0, 6) }))}
                className={`${inp} text-center text-2xl font-black tracking-[0.5em]`} />
            </Field>
          )}
          {(step === 'login' || step === 'signup' || step === 'reset') && (
            <Field label={step === 'reset' ? 'كلمة المرور الجديدة' : 'كلمة المرور'} hint={step !== 'login' ? '8 أحرف على الأقل' : undefined}>
              <input required type="password" dir="ltr" minLength={step === 'login' ? 1 : 8} value={f.password} onChange={set('password')} className={inp}
                autoComplete={step === 'login' ? 'current-password' : 'new-password'} />
            </Field>
          )}

          {err && <p className="mb-3 rounded-xl bg-rose-50 px-3 py-2 text-sm font-bold text-rose-700">{err}</p>}
          {info && <p className="mb-3 rounded-xl bg-emerald-50 px-3 py-2 text-sm font-bold text-emerald-700">{info}</p>}

          <button disabled={busy} className="w-full rounded-2xl bg-navy py-3.5 text-base font-black text-white shadow-lg shadow-navy/25 transition active:scale-[.98] disabled:opacity-60">
            {busy ? '…' : { login: 'دخول', signup: 'إنشاء الحساب', verify: 'تأكيد', forgot: 'أرسل الرمز', reset: 'حفظ ودخول' }[step]}
          </button>

          {(step === 'verify' || step === 'reset') && (
            <button type="button" disabled={cool > 0 || busy} onClick={resend} className="mt-3 w-full text-sm font-bold text-navy disabled:text-slate-400">
              {cool > 0 ? `إعادة الإرسال بعد ${cool} ثانية` : 'ما وصلني — أرسل مرة ثانية'}
            </button>
          )}

          <div className="mt-5 space-y-2 text-center text-sm">
            {step === 'login' && <>
              <p className="text-slate-500">ما عندك حساب؟ <button type="button" onClick={() => go('signup')} className="font-black text-sun">سوّي حساب</button></p>
              <button type="button" onClick={() => go('forgot')} className="text-slate-400">نسيت كلمة المرور؟</button>
            </>}
            {step !== 'login' && <button type="button" onClick={() => go('login')} className="font-bold text-slate-500">→ رجوع لتسجيل الدخول</button>}
          </div>
        </form>

        <p className="mt-6 text-center text-xs text-slate-400">بإنشاء الحساب توافق على استخدام بياناتك لخدمتك وتواصلنا وياك فقط.</p>
      </div>
    </div>
  )
}

const inp = 'w-full rounded-2xl border border-slate-200 bg-slate-50 px-4 py-3 text-base outline-none transition focus:border-navy focus:bg-white focus:ring-4 focus:ring-navy/10'

function Field({ label, hint, children }: { label: string; hint?: string; children: ReactNode }) {
  return (
    <label className="mb-4 block">
      <span className="mb-1.5 block text-sm font-bold text-slate-600">{label}</span>
      {children}
      {hint && <span className="mt-1 block text-[11px] text-slate-400">{hint}</span>}
    </label>
  )
}
