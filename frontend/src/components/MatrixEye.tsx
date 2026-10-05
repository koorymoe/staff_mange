import { useCallback, useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { Link, useLocation } from 'react-router-dom'
import { api, type MatrixWatch } from '../api'
import { useSession } from '../session'
import { isEnabled, SWITCH_MATRIX_STAFF_EYE } from '../systemSwitches'
import MatrixEyeGraphic from './MatrixEyeGraphic'
import { eyeColor, type EyeGroup, type EyeMood } from './matrixEyeColors'

// ═══ عين ماتركس — «عين الرب» ═══
//
// قرار (ع): العين **ما تكون موجودة دايماً**. تطلع لما الموظف يشتغل على
// شاشة شغله ويسوي إجراء — المراقب يدقق، المحاسب يعتمد، الليدر يكمّل
// ورقه — وتحچي بأرقامه الحقيقية: «دققت ٢، باقي ٣». وتختفي بعد شوية.
// والاستثناء: لو حالته **حمرة** (تذكير صعد أو ٣ مفتوحة)، تطلع بأي شاشة
// يدخلها — التصعيد ما ينخفي.
//
// ⚠️ تعكس بس — ما تعاقب ولا تخترع. كل جملة برقم من الخادم.

const VISIBLE_MS = 12000
const SPEAK_MS = 7000
const POLL_MS = 5 * 60 * 1000

// شاشات الشغل الاحتياطية لكل مجموعة — لو الأرقام ما وصلت، العين تبقى تعرف
// وين يشتغل الموظف.
const FALLBACK_ROUTES: Record<string, string[]> = {
  MONITORS: ['/monitor-desk', '/daily-audit', '/monitor-inbox', '/leader-invoices', '/audit-issues', '/finance', '/crew-bookings-audit'],
  COORDINATORS: ['/coordinator', '/bookings', '/staff-requests', '/stage-buckets'],
  FINANCE: ['/leader-invoices', '/finance', '/daily-audit', '/expenses', '/revolving-fund'],
  LEADERS: ['/my-tasks', '/leader-invoices/new', '/work-reports', '/missions'],
  TECHS: ['/my-tasks', '/missions', '/my-inventory', '/attendance'],
  DESIGN: ['/design-gallery', '/design-forms', '/unit-design'],
  QUALITY: ['/quality-follow-ups', '/quality', '/complaints'],
  IT: ['/it-assets', '/it-stats'],
  SALES: ['/bookings', '/customers'],
  PROJECTS: ['/projects', '/staff-requests'],
  GPS: ['/gps', '/gps/devices', '/gps/customers', '/gps/sims'],
  ADMINS: [],
  STAFF: [],
}

function currentPath(): string {
  return window.location.pathname.replace(/^\/staff_mange/, '') || '/'
}

function isWorkRoute(w: MatrixWatch | null, path: string): boolean {
  if (!w) return false
  if (w.workload.some((x) => onRoute(path, x.route))) return true
  return (FALLBACK_ROUTES[w.group] ?? []).some((r) => onRoute(path, r))
}

function onRoute(path: string, route: string): boolean {
  const base = route.split('?')[0]
  return path === base || path.startsWith(base + '/')
}

// جملة ماتركس حسب الموقف — صارمة، برقم، بلا مدح فاضي.
function speech(w: MatrixWatch, path: string, afterAction: boolean): string | null {
  if (w.mood === 'ANGRY' && w.items.length > 0) {
    const esc = w.items.find((i) => i.escalated) ?? w.items[0]
    return w.items.some((i) => i.escalated)
      ? `ذكّرتك وما انحل: ${esc.summary}. القضية وصلت للمدير.`
      : `عندك ${w.open} تذكيرات مفتوحة وما انحلت. أولها: ${esc.summary}.`
  }
  const here = w.workload.find((x) => onRoute(path, x.route))
  if (afterAction && here) {
    if (here.left === 0) {
      const all = w.workload.every((x) => x.left === 0)
      return all
        ? `${here.verb} ${here.done} اليوم وخلّصت كل الي عليك. مسجّل.`
        : `خلّصت «${here.label}». باقي عليك: ${w.workload.filter((x) => x.left > 0).map((x) => `${x.left} ${x.label}`).join('، ')}.`
    }
    return here.done > 0
      ? `${here.verb} ${here.done} اليوم. باقي ${here.left} من «${here.label}». كمّلها.`
      : `باقي ${here.left} من «${here.label}». ماتركس يتابع.`
  }
  if (afterAction && w.open > 0) return `مسجّل. وعندك ${w.open} تذكير مفتوح بعده.`
  if (afterAction) return 'ماتركس يتابع شغلك هنا.'
  return null
}

export default function MatrixEye() {
  const { employee } = useSession()
  const location = useLocation()
  const path = location.pathname
  const [watch, setWatch] = useState<MatrixWatch | null>(null)
  const [visible, setVisible] = useState(false)
  const [flash, setFlash] = useState(false)
  const [blink, setBlink] = useState(false)
  const [open, setOpen] = useState(false)
  const [say, setSay] = useState<string | null>(null)
  const pupilRef = useRef<SVGGElement>(null)
  const boxRef = useRef<HTMLButtonElement>(null)
  const hideT = useRef(0)
  // مفتاح (ع): «عين ماتركس عند الموظفين» — ينطفي ويرجع. المدير والمالك تبقى عندهم.
  const [staffOff, setStaffOff] = useState(false)
  useEffect(() => { isEnabled(SWITCH_MATRIX_STAFF_EYE).then((on) => setStaffOff(!on)).catch(() => {}) }, [path])
  const sayT = useRef(0)

  const group = (watch?.projectMode ? 'SUPERVISORS' : watch?.group ?? 'STAFF') as EyeGroup
  const mood = (watch?.mood ?? 'CALM') as EyeMood
  const color = eyeColor(group, mood)
  const angry = mood === 'ANGRY'
  // المالك والمدير: العين ظاهرة دايماً وتنطيهم تقارير (طلب (ع)) —
  // للموظفين تبقى تطلع عند الإجراء بس.
  const isBoss = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'
  const shown = visible || angry || open || isBoss

  const speak = useCallback((text: string | null) => {
    window.clearTimeout(sayT.current)
    setSay(text)
    if (text) sayT.current = window.setTimeout(() => setSay(null), SPEAK_MS)
  }, [])

  // الحالة: أول تحميل وكل ٥ دقايق.
  useEffect(() => {
    let alive = true
    const load = () => api.getMyWatch().then((w) => { if (alive) setWatch(w) }).catch(() => {})
    load()
    const t = window.setInterval(load, POLL_MS)
    // بعد كل حفظ: التذكير الي انحل ينسكّر بالخادم، فنعيد الحالة فوراً.
    window.addEventListener('matrix-refresh', load)
    return () => { alive = false; window.clearInterval(t); window.removeEventListener('matrix-refresh', load) }
  }, [])

  // الحمرة: أول ما يدخل أي شاشة، العين تحچي.
  useEffect(() => {
    if (!watch || watch.mood !== 'ANGRY') return
    const t = window.setTimeout(() => speak(speech(watch, path, false)), 600)
    return () => window.clearTimeout(t)
  }, [path, watch, speak])

  // الظهور: (١) حفظ ناجح (matrix-saw) — بأرقام جديدة من الخادم.
  // (٢) أي ضغطة زر داخل محتوى شاشة شغله — فتح فاتورة، تدقيق، تفاصيل.
  // وأول دخول للشاشة ما يطلّعها: «من يدخل ماموجودة، من يشتغل تطلع».
  const watchRef = useRef<MatrixWatch | null>(null)
  useEffect(() => { watchRef.current = watch }, [watch])
  const lastReveal = useRef(0)
  useEffect(() => {
    const reveal = (w: MatrixWatch) => {
      const p = currentPath()
      if (!isWorkRoute(w, p) && w.mood !== 'ANGRY') return
      lastReveal.current = Date.now()
      setVisible(true)
      setFlash(true)
      window.setTimeout(() => setFlash(false), 1300)
      speak(speech(w, p, true))
      window.clearTimeout(hideT.current)
      hideT.current = window.setTimeout(() => setVisible(false), VISIBLE_MS)
    }
    const onSaw = () => {
      api.getMyWatch().then((w) => { setWatch(w); reveal(w) }).catch(() => { if (watchRef.current) reveal(watchRef.current) })
    }
    // أي ضغطة زر بأي شاشة: ماتركس يسأل عقل التوجيه (تعليمات المدير +
    // هايكو لو موجود). إذا عنده توجيه، العين تطلع وتحچيه. وبشاشة الشغل
    // بلا تعليمة خاصة، تحچي بأرقام الشغل.
    const showWith = (w: MatrixWatch, text: string) => {
      lastReveal.current = Date.now()
      setVisible(true)
      setFlash(true)
      window.setTimeout(() => setFlash(false), 1300)
      speak(text)
      window.clearTimeout(hideT.current)
      hideT.current = window.setTimeout(() => setVisible(false), VISIBLE_MS)
      void w
    }
    const onClick = (e: MouseEvent) => {
      const t = e.target as HTMLElement | null
      const btn = t?.closest('button, a, [role="button"], [role="tab"]') as HTMLElement | null
      if (!t || !btn || !t.closest('main')) return
      if (Date.now() - lastReveal.current < 2500) return
      const w = watchRef.current
      if (!w) return
      const p = currentPath()
      const label = (btn.innerText || btn.getAttribute('aria-label') || btn.title || '').replace(/\s+/g, ' ').trim().slice(0, 60)
      lastReveal.current = Date.now()
      api.getMatrixGuide(p, label).then((g) => {
        if (g.text) showWith(w, g.text)
        else if (isWorkRoute(w, p) || w.mood === 'ANGRY') reveal(w)
      }).catch(() => { if (isWorkRoute(w, p)) reveal(w) })
    }
    window.addEventListener('matrix-saw', onSaw)
    document.addEventListener('click', onClick, true)
    return () => {
      window.removeEventListener('matrix-saw', onSaw)
      document.removeEventListener('click', onClick, true)
      window.clearTimeout(hideT.current)
      window.clearTimeout(sayT.current)
    }
  }, [speak])

  // رمشة — الغاضبة ما ترمش، تحدّق.
  useEffect(() => {
    if (angry || !shown) return
    let t = 0
    const loop = () => {
      t = window.setTimeout(() => { setBlink(true); window.setTimeout(() => setBlink(false), 130); loop() }, 3000 + Math.random() * 3500)
    }
    loop()
    return () => window.clearTimeout(t)
  }, [angry, shown])

  // البؤبؤ يتبع الماوس — بالـDOM مباشرة.
  useEffect(() => {
    let raf = 0
    const onMove = (e: MouseEvent) => {
      cancelAnimationFrame(raf)
      raf = requestAnimationFrame(() => {
        const box = boxRef.current?.getBoundingClientRect()
        const g = pupilRef.current
        if (!box || !g) return
        const dx = e.clientX - (box.left + box.width / 2)
        const dy = e.clientY - (box.top + box.height / 2)
        const d = Math.hypot(dx, dy) || 1
        const k = Math.min(1, d / 280)
        g.setAttribute('transform', `translate(${((dx / d) * 9 * k).toFixed(2)} ${((dy / d) * 4.5 * k).toFixed(2)})`)
      })
    }
    window.addEventListener('mousemove', onMove)
    return () => { window.removeEventListener('mousemove', onMove); cancelAnimationFrame(raf) }
  }, [])

  if (!employee) return null
  if (staffOff && employee.role !== 'ADMIN' && employee.actualRole !== 'OWNER') return null

  const lid = !shown ? 0 : blink ? 0.06 : flash ? 1.15 : angry ? 0.6 : mood === 'PLEASED' ? 0.75 : 0.95
  const title = angry ? 'ماتركس يراقبك عن قرب' : 'ماتركس'

  return (
    <>
      {shown && createPortal(
        <div aria-hidden className={`matrix-eye-frame ${angry ? 'is-red' : ''} ${flash ? 'is-flash' : ''}`} style={{ ['--eye' as string]: color }} />,
        document.body,
      )}

      <div className={`matrix-eye-slot ${shown ? 'is-shown' : ''}`}>
        <button
          ref={boxRef}
          type="button"
          onClick={() => setOpen((o) => !o)}
          title={title}
          aria-label={title}
          className={`matrix-eye ${angry ? 'is-red' : ''} ${flash ? 'is-flash' : ''}`}
          style={{ ['--eye' as string]: color }}
        >
          <MatrixEyeGraphic ref={pupilRef} group={group} mood={mood} lid={lid} flash={flash} width={76} />
          {watch && watch.open > 0 && shown && (
            <span className="absolute -top-1 -end-1 grid h-4 min-w-4 place-items-center rounded-full px-1 text-[10px] font-extrabold text-white" style={{ background: angry ? '#ef4444' : '#ca8a04' }}>{watch.open}</span>
          )}
        </button>

        {say && shown && !open && (
          <div dir="rtl" className="matrix-eye-say" style={{ ['--eye' as string]: color }}>
            <b>ماتركس:</b> {say}
          </div>
        )}

        {open && (
          <div dir="rtl" className="absolute start-0 top-full z-50 mt-2 w-80 max-w-[calc(100vw-2rem)] rounded-2xl border border-slate-200 bg-white p-3 text-sm shadow-xl">
            <p className="font-extrabold" style={{ color: mood === 'CALM' ? undefined : color }}>
              {angry ? '🔴 ماتركس يراقبك عن قرب' : mood === 'ALERT' ? '🟡 ماتركس منتبه' : mood === 'PLEASED' ? '🟢 ماتركس راضي عن شغلك اليوم' : '👁️ ماتركس يتابع'}
            </p>
            {watch && watch.workload.length > 0 && (
              <div className="mt-2">
                <p className="mb-1 text-xs font-bold text-slate-500">شغلك اليوم</p>
                <ul className="space-y-1">
                  {watch.workload.map((w) => (
                    <li key={w.key} className="rounded-lg bg-slate-50 px-2 py-1.5">
                      <div className="flex items-center justify-between gap-2">
                        <Link to={w.route} onClick={() => setOpen(false)} className="font-bold text-slate-800 hover:underline">{w.label} ←</Link>
                        <span className="text-xs tabular-nums">
                          {w.done > 0 && <span className="text-emerald-700">{w.verb} {w.done} · </span>}
                          <b className={w.left > 0 ? 'text-amber-700' : 'text-emerald-700'}>{w.left > 0 ? `باقي ${w.left}` : 'خالص ✓'}</b>
                        </span>
                      </div>
                      {/* أزرار تودّي للشي نفسه مباشرة — فاتورة هالحجز، مو قائمة عامة */}
                      {(w.links?.length || w.codes?.length) ? (
                        <div className="mt-1 flex flex-wrap gap-1">
                          {(w.links ?? (w.codes ?? []).map((c) => ({ label: c, to: w.route }))).slice(0, 12).map((l) => (
                            <Link key={l.label} to={l.to} onClick={() => setOpen(false)} className="rounded-md bg-white px-1.5 py-0.5 text-[11px] font-bold text-brand-700 ring-1 ring-slate-200 hover:bg-brand-50">{l.label}</Link>
                          ))}
                        </div>
                      ) : null}
                    </li>
                  ))}
                </ul>
              </div>
            )}
            {watch && watch.items.length > 0 && (
              <div className="mt-2">
                <p className="mb-1 text-xs font-bold text-slate-500">تذكيرات ما انحلت</p>
                <ul className="max-h-52 space-y-1.5 overflow-y-auto">
                  {watch.items.map((it, i) => (
                    <li key={i} className="rounded-lg bg-slate-50 p-2">
                      <p className="text-slate-800">{it.escalated && '⬆️ '}{it.summary}</p>
                      <p className="text-[11px] text-slate-500">{it.label} · من {new Date(it.since).toLocaleDateString('ar-IQ', { timeZone: 'Asia/Baghdad' })}{it.escalated ? ' · وصل للمدير' : ''}</p>
                    </li>
                  ))}
                </ul>
              </div>
            )}
            {isBoss && <BossReport onGo={() => setOpen(false)} />}
            {!isBoss && watch && watch.workload.length === 0 && watch.items.length === 0 && <p className="mt-1 text-xs text-slate-500">ماكو عليك شي مفتوح.</p>}
          </div>
        )}
      </div>
    </>
  )
}

// ═══ تقرير المالك/المدير من العين — يتجمّع لمن تنفتح بس ═══
function BossReport({ onGo }: { onGo: () => void }) {
  const [r, setR] = useState<{ attention: number; red: number; people: number; props: number; latePct: number | null; lateOpen: number } | null>(null)
  useEffect(() => {
    let alive = true
    Promise.all([
      api.getRoleWatch().catch(() => []),
      api.getMatrixProposals('PENDING').catch(() => []),
      api.getLateFocus().catch(() => null),
    ]).then(([g, p, l]) => {
      if (!alive) return
      setR({
        attention: g.reduce((n, x) => n + x.red + x.alert, 0), red: g.reduce((n, x) => n + x.red, 0),
        people: g.reduce((n, x) => n + x.employees.length, 0), props: p.length,
        latePct: l ? l.recentPct : null, lateOpen: l ? l.recent.openNow : 0,
      })
    })
    return () => { alive = false }
  }, [])
  if (!r) return <p className="mt-2 text-xs text-slate-400">ماتركس يجمع التقرير…</p>
  const row = (icon: string, text: string, to: string, tone = 'text-slate-800') => (
    <Link to={to} onClick={onGo} className={`flex items-center justify-between gap-2 rounded-lg bg-slate-50 px-2 py-1.5 hover:bg-sky-50 ${tone}`}>
      <span>{icon} {text}</span><span className="text-xs text-sky-700">←</span>
    </Link>
  )
  return (
    <div className="mt-2 space-y-1.5">
      <p className="text-xs font-bold text-slate-500">📊 تقرير ماتركس الآن</p>
      {row('👥', `${r.people} موظف تحت المراقبة — ${r.attention} يحتاجون انتباه${r.red ? ` (${r.red} حمر)` : ''}`, '/?board=matrix', r.red ? 'text-red-700' : 'text-slate-800')}
      {row('💡', r.props ? `${r.props} اقتراح ينتظر قرارك` : 'ماكو اقتراحات تنتظرك', '/matrix/decisions', r.props ? 'text-amber-700' : 'text-slate-800')}
      {r.latePct != null && row('⏰', `الحجوزات المتأخرة آخر ٣ أيام: ${r.latePct}%${r.lateOpen ? ` — ${r.lateOpen} مفتوحة وموعدها فات` : ''}`, '/?board=matrix', r.latePct >= 30 ? 'text-red-700' : 'text-slate-800')}
      {row('👁️', 'افتح مركز قيادة ماتركس', '/?board=matrix', 'font-bold text-brand-700')}
    </div>
  )
}
