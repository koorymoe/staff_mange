import { useEffect, useRef, useState } from 'react'
import { createPortal } from 'react-dom'
import { api, type MatrixWatch } from '../api'
import { useSession } from '../session'

// ═══ عين ماتركس — «عين الرب» ═══
//
// عين حيّة بالهيدر بكل الشاشات، ويا إطار يومض بخفة حول النظام: الموظف
// يعرف إن أكو متابعة. لونها حسب دوره، وحالتها حسب ماتركس الحقيقي:
//   هادئة  — ماكو عليه تذكير مفتوح.
//   منتبهة — عنده تذكير مفتوح (أصفر).
//   حمرة   — تذكير صعد أو ٣ مفتوحة (حمرة، بؤبؤ مشقوق، إطار أقوى).
// وكل ما يحفظ شي تنفتح وتتوهج لحظة: «شفتك».
//
// ⚠️ العين **تعكس بس** — ما تعاقب ولا تخترع. ضغطها يبين «ليش» بالضبط،
// وترجع هادئة لحالها أول ما ينحل الشي.

type Level = MatrixWatch['level']

// لون العين لكل دور — الحمرة محجوزة للتصعيد بس، فماكو دور أحمر.
const ROLE_COLOR: Record<string, string> = {
  OWNER: '#f5b301',
  ADMIN: '#f5b301',
  MONITOR: '#a855f7',
  HR_COORDINATOR: '#22d3ee',
  FINANCE: '#10b981',
  IT_SUPPORT: '#3b82f6',
  LEADER: '#fb923c',
}
const DEFAULT_COLOR = '#38bdf8'
const ALERT_COLOR = '#facc15'
const RED_COLOR = '#ef4444'

const POLL_MS = 5 * 60 * 1000

export default function MatrixEye() {
  const { employee } = useSession()
  const [watch, setWatch] = useState<MatrixWatch | null>(null)
  const [saw, setSaw] = useState(0) // عدّاد «شفتك» — كل زيادة تشغّل الومضة
  const [flash, setFlash] = useState(false)
  const [open, setOpen] = useState(false)
  const [blink, setBlink] = useState(false)
  const pupilRef = useRef<SVGGElement>(null)
  const boxRef = useRef<HTMLButtonElement>(null)

  const role = employee?.actualRole === 'OWNER' ? 'OWNER' : employee?.isLeader ? 'LEADER' : (employee?.role ?? '')
  const level: Level = watch?.level ?? 'CALM'
  const base = ROLE_COLOR[role] ?? DEFAULT_COLOR
  const color = level === 'RED' ? RED_COLOR : level === 'ALERT' ? ALERT_COLOR : base

  // الحالة: أول تحميل، وكل ٥ دقايق، وبعد كل حفظ (حتى تنطفي الحمرة أول ما ينحل).
  useEffect(() => {
    let alive = true
    api.getMyWatch().then((w) => { if (alive) setWatch(w) }).catch(() => {})
    const t = window.setInterval(() => {
      api.getMyWatch().then((w) => { if (alive) setWatch(w) }).catch(() => {})
    }, POLL_MS)
    return () => { alive = false; window.clearInterval(t) }
  }, [saw])

  // «شفتك»: كل إجراء ناجح بأي شاشة.
  useEffect(() => {
    let timer = 0
    const onSaw = () => {
      setFlash(true)
      window.clearTimeout(timer)
      timer = window.setTimeout(() => { setFlash(false); setSaw((n) => n + 1) }, 1400)
    }
    window.addEventListener('matrix-saw', onSaw)
    return () => { window.removeEventListener('matrix-saw', onSaw); window.clearTimeout(timer) }
  }, [])

  // رمشة كل كم ثانية — الحمرة ما ترمش (تحدّق).
  useEffect(() => {
    if (level === 'RED') return
    let t = 0
    const loop = () => {
      t = window.setTimeout(() => {
        setBlink(true)
        window.setTimeout(() => setBlink(false), 140)
        loop()
      }, 3500 + Math.random() * 4000)
    }
    loop()
    return () => window.clearTimeout(t)
  }, [level])

  // البؤبؤ يتبع الماوس — بالـDOM مباشرة حتى ما نعيد الرسم بكل حركة.
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
        const k = Math.min(1, d / 300)
        g.setAttribute('transform', `translate(${((dx / d) * 6 * k).toFixed(2)} ${((dy / d) * 3 * k).toFixed(2)})`)
      })
    }
    window.addEventListener('mousemove', onMove)
    return () => { window.removeEventListener('mousemove', onMove); cancelAnimationFrame(raf) }
  }, [])

  if (!employee) return null

  // فتحة الجفن: ترمش → تنسد، شفتك → تنفتح واسعة، حمرة → ضيقة غاضبة.
  const lid = blink ? 0.08 : flash ? 1.15 : level === 'RED' ? 0.62 : 0.9
  const pupilW = level === 'RED' ? 1.6 : flash ? 3.2 : 4.4
  const title = level === 'RED' ? 'ماتركس يراقبك عن قرب' : level === 'ALERT' ? 'ماتركس منتبه' : 'ماتركس يتابع'

  return (
    <>
      {/* الإطار الومّاض حول الشاشة كلها — بالـbody مباشرة، لأن الهيدر
          عنده تأثيرات تخلّي fixed ينحصر داخله. وما يمسك الضغطات أبداً. */}
      {createPortal(
        <div
          aria-hidden
          className={`matrix-eye-frame ${level === 'RED' ? 'is-red' : ''} ${flash ? 'is-flash' : ''}`}
          style={{ ['--eye' as string]: color }}
        />,
        document.body,
      )}

      <div className="relative">
        <button
          ref={boxRef}
          type="button"
          onClick={() => setOpen((o) => !o)}
          title={title}
          aria-label={title}
          className={`matrix-eye ${level === 'RED' ? 'is-red' : ''} ${flash ? 'is-flash' : ''}`}
          style={{ ['--eye' as string]: color }}
        >
          <svg viewBox="-30 -16 60 32" width="58" height="31">
            <defs>
              <radialGradient id="mx-iris" cx="50%" cy="50%" r="50%">
                <stop offset="0%" stopColor="#fff" stopOpacity="0.95" />
                <stop offset="35%" stopColor={color} />
                <stop offset="100%" stopColor="#020617" />
              </radialGradient>
              <clipPath id="mx-lid">
                <path d={`M-27 0 Q0 ${-17 * lid} 27 0 Q0 ${17 * lid} -27 0Z`} />
              </clipPath>
            </defs>
            {/* بياض العين داخل الجفن */}
            <path d={`M-27 0 Q0 ${-17 * lid} 27 0 Q0 ${17 * lid} -27 0Z`} fill="#0b1220" stroke={color} strokeWidth="1.6" className="mx-lid" />
            <g clipPath="url(#mx-lid)">
              <g ref={pupilRef}>
                <circle r="9.5" fill="url(#mx-iris)" />
                <circle r="9.5" fill="none" stroke={color} strokeOpacity="0.7" strokeWidth="0.8" className="mx-ring" />
                <ellipse rx={pupilW} ry="5.5" fill="#000" />
                <circle cx="-3" cy="-3" r="1.6" fill="#fff" opacity="0.85" />
              </g>
            </g>
            {/* حاجب غاضب بالحمرة */}
            {level === 'RED' && <path d="M-24 -13 L-4 -8 M24 -13 L4 -8" stroke={RED_COLOR} strokeWidth="2.2" strokeLinecap="round" />}
          </svg>
          {watch && watch.open > 0 && (
            <span className="absolute -top-1 -end-1 grid h-4 min-w-4 place-items-center rounded-full px-1 text-[10px] font-extrabold text-white" style={{ background: color === base ? RED_COLOR : color }}>{watch.open}</span>
          )}
        </button>

        {open && (
          <div dir="rtl" className="absolute start-0 top-full z-50 mt-2 w-80 max-w-[calc(100vw-2rem)] rounded-2xl border border-slate-200 bg-white p-3 text-sm shadow-xl">
            <p className="font-extrabold" style={{ color: level === 'CALM' ? undefined : color }}>
              {level === 'RED' ? '🔴 ماتركس يراقبك عن قرب' : level === 'ALERT' ? '🟡 ماتركس منتبه' : '👁️ ماتركس يتابع — كلشي تمام'}
            </p>
            {watch && watch.items.length > 0 ? (
              <>
                <p className="mb-2 text-xs text-slate-500">
                  {level === 'RED' ? 'عندك أشياء ذكّرك بيها وما انحلت. العين ترجع هادئة أول ما تخلصها.' : 'عندك شي ذكّرك بي ماتركس:'}
                </p>
                <ul className="max-h-64 space-y-1.5 overflow-y-auto">
                  {watch.items.map((it, i) => (
                    <li key={i} className="rounded-lg bg-slate-50 p-2">
                      <p className="text-slate-800">{it.escalated && '⬆️ '}{it.summary}</p>
                      <p className="text-[11px] text-slate-500">{it.label} · من {new Date(it.since).toLocaleDateString('ar-IQ', { timeZone: 'Asia/Baghdad' })}{it.escalated ? ' · وصل للمدير' : ''}</p>
                    </li>
                  ))}
                </ul>
              </>
            ) : (
              <p className="mt-1 text-xs text-slate-500">ماكو عليك ولا تذكير مفتوح.</p>
            )}
          </div>
        )}
      </div>
    </>
  )
}
