import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, type StoryWithScene } from '../api'
import { useSession } from '../session'

// ═══ مشهد الكيان — القصة تنعرض وتُقَرّ ═══
//
// المراقب يضغط «خصم نقطة» ← شخصيته تدخل من الحافة حاملة ورقة ←
// تسلّمها ← الورقة تنفتح وتنقرا ← الموظف يقرّ بالاطلاع.
//
// ⚠️⚠️ **هذي النسخة النصية، وهي مو مؤقتة**: عقد الكيان يفرض
// **fallback نصي كامل** لمّا يفشل Rive أو يكون المتصفح على
// `prefers-reduced-motion`. فهذا الملف **نصف المنتج الدائم** —
// والفن لمّا يوصل يركب فوقه، ما يبدّله.
//
// ⚠️ **ولا جملة ولا رقم ينولد هنا**: كل شي يجي من `payload` الخادم
// المحسوب من الحدث الرسمي. جملة تخترعها الواجهة تعني تحذيراً بلا
// حدث وراه — وأول مرة يكتشفها الموظف يبطّل يقرا أي تحذير بعدها.

/**
 * استطلاع هادئ لمّا ماكو قصة تنتظر.
 *
 * ⚠️ **لگّاه الفحص الحي**: چان ٦٠ ثانية مثل بقية الاستطلاع، فأول
 * قصة تتأخر **دقيقة كاملة** — المراقب يخصم والموظف ما يشوف شي.
 * صار ١٥ث لأن الاستعلام **رخيص فعلاً**: فهرس جزئي على المعلّقة
 * وصف واحد بس، مو جلب قائمة.
 */
const IDLE_POLL_MS = 15000
/**
 * استطلاع سريع لمّا يكون بالطابور شي.
 *
 * ⚠️ **هذا بديل SSE بقرار مقاس**: `WriteTimeout` بالخادم ٣٠ ثانية،
 * فاتصال SSE ينقطع كل نصف دقيقة — يطلع أسوأ من الاستطلاع مو أحسن.
 * خمس ثوانٍ تأخير أقصى، والموظف أصلاً ما يعرف متى انضغط الزر.
 */
const ACTIVE_POLL_MS = 5000

function prefersReducedMotion(): boolean {
  try {
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches
  } catch {
    return false
  }
}

/** نصّ آمن من `payload` — بلا تخمين ولا قيمة مخترعة. */
function text(payload: Record<string, unknown>, key: string): string {
  const v = payload?.[key]
  return typeof v === 'string' ? v : ''
}

function num(payload: Record<string, unknown>, key: string): number | null {
  const v = payload?.[key]
  return typeof v === 'number' ? v : null
}

export default function StoryScene() {
  const { employee } = useSession()
  const navigate = useNavigate()
  const [story, setStory] = useState<StoryWithScene | null>(null)
  const [pending, setPending] = useState(0)
  const [phase, setPhase] = useState<'arriving' | 'reading'>('arriving')
  const [busy, setBusy] = useState(false)
  const reduced = useMemo(() => prefersReducedMotion(), [])
  // نمنع تكرار إعلان «وصلت» لنفس القصة عند كل استطلاع
  const announced = useRef<string | null>(null)

  const load = useCallback(() => {
    if (!employee) return
    api.getNextStory()
      .then((res) => {
        setStory((prev) => (prev && prev.id === res.story?.id ? prev : res.story))
        setPending(res.pending)
      })
      .catch(() => { /* المشهد ما يكسر الشاشة — الإشعارات تبقى تشتغل */ })
  }, [employee])

  useEffect(() => {
    load()
    const iv = setInterval(() => {
      // ⚠️ الصفحة مخفية ← ما نستطلع: كيان يشتغل بالخلفية يحرق
      // بطارية الفني بالميدان، وهذا سبب حقيقي يخلّيه يسكّر النظام.
      if (document.hidden) return
      load()
    }, story ? ACTIVE_POLL_MS : IDLE_POLL_MS)
    const onVisible = () => { if (!document.hidden) load() }
    document.addEventListener('visibilitychange', onVisible)
    return () => {
      clearInterval(iv)
      document.removeEventListener('visibilitychange', onVisible)
    }
  }, [load, story])

  // ── وصلت للجهاز: مرحلة `DELIVERED` ──
  //
  // ⚠️ **«وصلت» مو «انقرأت»**: هذي تسجّل إن الجهاز استلمها بس.
  // القراءة إقرار الموظف نفسه تحت.
  useEffect(() => {
    if (!story || announced.current === story.id) return
    announced.current = story.id
    setPhase('arriving')
    void api.advanceStory(story.id, 'DELIVERED', 0).catch(() => {})
    // الحركة زينة والمعنى إلزامي: على `reduced-motion` نقفز للقراءة فوراً
    const delay = reduced ? 0 : (story.scene[0]?.durationMs ?? 800)
    const t = setTimeout(() => {
      setPhase('reading')
      void api.advanceStory(story.id, 'SEEN', 1).catch(() => {})
    }, delay)
    return () => clearTimeout(t)
  }, [story, reduced])

  const acknowledge = useCallback(async () => {
    if (!story || busy) return
    setBusy(true)
    try {
      await api.advanceStory(story.id, 'ACKNOWLEDGED', story.scene.length)
      announced.current = null
      setStory(null)
      load()
    } finally {
      setBusy(false)
    }
  }, [story, busy, load])

  const openDetails = useCallback(async () => {
    if (!story) return
    await api.advanceStory(story.id, 'OPENED', 3).catch(() => {})
    const link = text(story.payload, 'link')
    if (link) navigate(link)
  }, [story, navigate])

  if (!story) return null

  const title = text(story.payload, 'title') || story.label
  const reason = text(story.payload, 'reason')
  const dinar = num(story.payload, 'dinar')
  const merged = num(story.payload, 'mergedCount') ?? 1
  const lines = Array.isArray(story.payload.lines) ? (story.payload.lines as string[]) : []
  const link = text(story.payload, 'link')
  const warning = story.priority >= 80

  return (
    <div
      dir="rtl"
      className="fixed inset-0 z-[80] flex items-center justify-center bg-black/50 p-4"
      role="dialog"
      aria-modal="true"
      aria-label={story.label}
    >
      <div
        className={`w-full max-w-md overflow-hidden rounded-2xl shadow-2xl ${
          reduced || phase === 'reading' ? 'opacity-100' : 'opacity-0'
        } transition-opacity duration-300`}
        style={{ background: 'var(--sf-1, #fff)', color: 'var(--t-1, #0f172a)' }}
      >
        {/* ── المُرسِل: اسمه ظاهر بقصد ── */}
        <div
          className="flex items-center gap-3 px-5 py-4"
          style={{ background: warning ? '#7f1d1d' : '#134e4a', color: '#fff' }}
        >
          <span className="text-2xl" aria-hidden>{warning ? '📋' : '✉️'}</span>
          <div className="flex-1">
            <div className="text-sm font-bold">{story.label}</div>
            <div className="text-xs opacity-90">
              {story.senderName ? `وصلتك من: ${story.senderName}` : 'رسالة من النظام'}
            </div>
          </div>
          {pending > 1 && (
            <span className="rounded-full bg-white/20 px-2 py-1 text-xs font-bold">
              +{pending - 1} بالانتظار
            </span>
          )}
        </div>

        {/* ── الورقة: محتواها كله من الخادم ── */}
        <div className="space-y-3 px-5 py-5">
          <p className="text-base font-bold">{title}</p>
          {reason && (
            <p className="text-sm" style={{ color: 'var(--t-2, #475569)' }}>
              السبب: {reason}
            </p>
          )}
          {dinar !== null && dinar > 0 && (
            <p className="text-sm font-bold" style={{ color: '#b91c1c' }}>
              الكلفة: {dinar.toLocaleString('en-US')} د.ع
            </p>
          )}
          {/* ⚠️ التجميع: عدة بنود بمشهد واحد بدل عدة مشاهد ورا بعض */}
          {merged > 1 && lines.length > 0 && (
            <ul className="space-y-1 rounded-lg p-3 text-sm" style={{ background: 'var(--sf-2, #f1f5f9)' }}>
              {lines.map((l, i) => <li key={i}>• {l}</li>)}
            </ul>
          )}
          <p className="text-xs" style={{ color: 'var(--t-3, #64748b)' }}>
            {new Date(story.createdAt).toLocaleString('en-GB')}
          </p>
        </div>

        <div className="flex gap-2 px-5 pb-5">
          <button
            type="button"
            onClick={acknowledge}
            disabled={busy}
            className="flex-1 rounded-xl px-4 py-2.5 text-sm font-bold text-white disabled:opacity-60"
            style={{ background: warning ? '#b91c1c' : '#0f2040' }}
          >
            {busy ? '…' : 'اطّلعت ✓'}
          </button>
          {link && (
            <button
              type="button"
              onClick={openDetails}
              className="rounded-xl px-4 py-2.5 text-sm font-bold"
              style={{ background: 'var(--sf-2, #e2e8f0)', color: 'var(--t-1, #0f172a)' }}
            >
              عرض التفاصيل
            </button>
          )}
        </div>
      </div>
    </div>
  )
}
