import { useCallback, useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api, fileUrl, ensureFileToken, type EntityBriefing, type EntityLine } from '../api'
import { useSession } from '../session'

// ═══ الكيان — مراقب ومساعد شخصي لكل موظف ═══
//
// «كيان يهابه ويخافه الموظف، أول ما يفتح النظام يرحّب بيه: آني
// المراقب عليك والمساعد بنفس الوقت… يراقب ويساعد ويتحرك بالشاشة،
// كأنما بشر يراقب بشر».
//
// ⚠️⚠️ **كل جملة يقولها الكيان تجي من الخادم** (`/entity/briefing`)
// محسوبة من نفس بيانات الغرامة الفعلية. هنا **ما ننشئ ولا جملة ولا
// رقم** — تحذير تخترعه الواجهة يعني الموظف يستعجل على شي ما راح
// يصير، وأول مرة يكتشفها يبطّل يقرا أي تحذير بعدها.
//
// ⚠️ الكيان يحلّ محل بوتين قديمين (`AssistantWidget` +
// `ManagerAssistantChat`) — بوت واحد بالشاشة مو اثنين. والمحادثة
// جوّاه تروح لنفس المسارين حسب الدور، بلا تغيير سلوك.

type Msg = { role: 'user' | 'assistant'; text: string }

/** مفاتيح التخزين — الموضع يتذكره المتصفح، والترحيب مرة بالجلسة. */
const POS_KEY = 'entityPos'
const GREETED_KEY = 'entityGreeted'

/** كل شكد يبدّل الكيان السطر المعروض بالفقاعة. */
const LINE_ROTATE_MS = 9000
/** كل شكد يسأل الخادم عن جديد — نفس اتفاقية الاستطلاع بالمشروع. */
const POLL_MS = 60000

export default function EntityCompanion() {
  const { employee } = useSession()
  const navigate = useNavigate()
  const isManager = employee?.role === 'ADMIN' || employee?.role === 'MONITOR' || employee?.actualRole === 'OWNER'

  const [brief, setBrief] = useState<EntityBriefing | null>(null)
  const [open, setOpen] = useState(false)
  const [bubble, setBubble] = useState<string | null>(null)
  const [lineIdx, setLineIdx] = useState(0)
  const [messages, setMessages] = useState<Msg[]>([])
  const [input, setInput] = useState('')
  const [sending, setSending] = useState(false)
  // اتجاه النظر: الكيان ينقلب ناحية المؤشر — هذي الي تعطي إحساس المراقبة
  const [facing, setFacing] = useState(1)

  // ── الموضع: يُسحب باليد ويتذكر مكانه ──
  const [pos, setPos] = useState<{ x: number; y: number }>(() => {
    try {
      const raw = localStorage.getItem(POS_KEY)
      if (raw) return JSON.parse(raw) as { x: number; y: number }
    } catch { /* تخزين مقفول أو قيمة خربانة — ننزل للافتراضي */ }
    return { x: 20, y: Math.max(80, window.innerHeight - 190) }
  })
  const drag = useRef<{ dx: number; dy: number; moved: boolean } | null>(null)

  // ── جلب التقرير ──
  const load = useCallback(() => {
    if (!employee) return
    api.getEntityBriefing().then(setBrief).catch(() => { /* الكيان ما يكسر الشاشة */ })
  }, [employee])

  useEffect(() => {
    ensureFileToken().finally(load)
    const iv = setInterval(load, POLL_MS)
    return () => clearInterval(iv)
  }, [load])

  // ── الترحيب: مرة وحدة بالجلسة ──
  useEffect(() => {
    if (!brief) return
    if (sessionStorage.getItem(GREETED_KEY)) return
    sessionStorage.setItem(GREETED_KEY, '1')
    setBubble(brief.greeting)
    const t = setTimeout(() => setBubble(null), 7000)
    return () => clearTimeout(t)
  }, [brief])

  // ── تدوير سطور التحذير بالفقاعة ──
  useEffect(() => {
    const lines = brief?.lines || []
    if (lines.length === 0) return
    const iv = setInterval(() => {
      setLineIdx((i) => (i + 1) % lines.length)
    }, LINE_ROTATE_MS)
    return () => clearInterval(iv)
  }, [brief])

  // ── نداء من شاشة ثانية: «اسأل الذكاء عن هذا الموظف» ──
  //
  // ⚠️ الكيان ورث حدث `open-manager-chat` من البوت القديم الي انشال.
  // بدون هذا السطر يصير زر «اسأل الذكاء الاصطناعي عن هذا الموظف»
  // بإدارة الكوادر يضغط وما يفتح ولا شي — زر ميّت بلا ما ينتبه أحد.
  useEffect(() => {
    const onOpen = (e: Event) => {
      const prefill = (e as CustomEvent<{ prefill?: string }>).detail?.prefill
      setOpen(true)
      if (prefill) setInput(prefill)
    }
    window.addEventListener('open-manager-chat', onOpen)
    return () => window.removeEventListener('open-manager-chat', onOpen)
  }, [])

  // ── الالتفات ناحية المؤشر ──
  useEffect(() => {
    const onMove = (e: PointerEvent) => setFacing(e.clientX < pos.x + 40 ? -1 : 1)
    window.addEventListener('pointermove', onMove)
    return () => window.removeEventListener('pointermove', onMove)
  }, [pos.x])

  // ── السحب باليد ──
  const onPointerDown = (e: React.PointerEvent) => {
    drag.current = { dx: e.clientX - pos.x, dy: e.clientY - pos.y, moved: false }
    ;(e.target as HTMLElement).setPointerCapture?.(e.pointerId)
  }
  const onPointerMove = (e: React.PointerEvent) => {
    if (!drag.current) return
    const x = Math.min(Math.max(0, e.clientX - drag.current.dx), window.innerWidth - 90)
    const y = Math.min(Math.max(0, e.clientY - drag.current.dy), window.innerHeight - 110)
    drag.current.moved = true
    setPos({ x, y })
  }
  const onPointerUp = () => {
    if (!drag.current) return
    // ⚠️ السحب ما يفتح اللوحة: بدون هذا الفحص كل تحريك للكيان يفتح
    // لوحته بوجه المستخدم، فيصير سحبه عقوبة مو ميزة.
    const moved = drag.current.moved
    drag.current = null
    try { localStorage.setItem(POS_KEY, JSON.stringify(pos)) } catch { /* تخزين مقفول */ }
    if (!moved) setOpen((o) => !o)
  }

  const send = async () => {
    const text = input.trim()
    if (!text || sending) return
    setInput('')
    setMessages((prev) => [...prev, { role: 'user', text }])
    setSending(true)
    try {
      // نفس تفريع الأدوار الموجود قبل الكيان — بلا تغيير سلوك المحادثة
      const res = isManager
        ? await api.managerChatAssistant(text, messages)
        : await api.askAssistant(text)
      setMessages((prev) => [...prev, { role: 'assistant', text: res.reply }])
    } catch (err) {
      setMessages((prev) => [...prev, { role: 'assistant', text: err instanceof Error ? err.message : 'صار خطأ، جرب مرة ثانية' }])
    } finally {
      setSending(false)
    }
  }

  if (!employee) return null

  const mood = brief?.mood || 'HAPPY'
  const lines = brief?.lines || []
  const activeLine: EntityLine | undefined = lines[lineIdx % Math.max(1, lines.length)]
  const urgentCount = lines.filter((l) => l.urgent).length
  const photo = moodImage(brief, mood)

  return (
    <div
      className="fixed z-[65] select-none"
      style={{ left: pos.x, top: pos.y }}
    >
      {/* ═══ اللوحة: واجباتي وغراماتي والمحادثة ═══ */}
      {open && (
        <div className="mb-3 flex max-h-[70vh] w-[22rem] flex-col overflow-hidden rounded-2xl border border-white bg-white shadow-2xl">
          <div className={`flex items-center justify-between px-4 py-3 text-white ${moodHeader(mood)}`}>
            <span className="font-bold">{moodTitle(mood)}</span>
            <button onClick={() => setOpen(false)} className="text-white/80 hover:text-white">✕</button>
          </div>

          <div className="flex-1 space-y-3 overflow-y-auto p-3">
            {brief && (
              <div className="rounded-xl bg-slate-50 p-3 text-xs text-slate-600 ring-1 ring-slate-100">
                <div className="flex items-center justify-between font-bold text-[#0f2040]">
                  <span>رصيد انضباطك</span>
                  <span className={brief.points < 100 ? 'text-red-600' : 'text-emerald-600'}>
                    {brief.points} / 100
                  </span>
                </div>
                {brief.dinarAtRisk > 0 && (
                  <p className="mt-1 text-red-600">
                    ⚠️ معرّض تخسر {brief.dinarAtRisk.toLocaleString('en-US')} د.ع لو ما تحرّكت هسه.
                  </p>
                )}
                {brief.persona && <p className="mt-1 text-slate-400">{brief.persona}</p>}
              </div>
            )}

            <div>
              <p className="mb-1 text-xs font-bold text-[#0f2040]">واجباتك ومتابعتها</p>
              {lines.length === 0 && (
                <p className="rounded-xl bg-emerald-50 px-3 py-2 text-xs font-bold text-emerald-700">
                  ✅ ماكو شي عليك هسه — شغلك نظيف، وآني أراقب.
                </p>
              )}
              {lines.map((l, i) => (
                <button
                  key={i}
                  onClick={() => { if (l.link) { navigate(l.link); setOpen(false) } }}
                  className={`mb-1.5 block w-full rounded-xl px-3 py-2 text-right text-xs font-semibold ${
                    l.urgent ? 'bg-red-50 text-red-700 ring-1 ring-red-200' : 'bg-amber-50 text-amber-800 ring-1 ring-amber-100'
                  }`}
                >
                  {l.urgent ? '⛔ ' : '⏳ '}{l.text}
                </button>
              ))}
            </div>

            <p className="rounded-lg bg-[#0f2040] px-3 py-2 text-[11px] font-bold text-white">
              🗣️ أي شغلة تتأخر عليها راح توصل للمدير بتقريري — الأحسن نخلصها هسه.
            </p>

            {/* المحادثة — نفس المساعد الذكي الي كان بالبوت القديم */}
            <div className="space-y-2">
              {messages.map((m, i) => (
                <div
                  key={i}
                  className={`max-w-[85%] rounded-xl px-3 py-2 text-sm ${
                    m.role === 'user' ? 'mr-auto bg-brand-500 text-white' : 'ml-auto bg-slate-100 text-slate-700'
                  }`}
                >
                  {m.text}
                </div>
              ))}
              {sending && <div className="ml-auto max-w-[85%] rounded-xl bg-slate-100 px-3 py-2 text-sm text-slate-400">...</div>}
            </div>
          </div>

          <div className="flex gap-2 border-t border-slate-100 p-3">
            <input
              value={input}
              onChange={(e) => setInput(e.target.value)}
              onKeyDown={(e) => e.key === 'Enter' && send()}
              placeholder="اسألني عن أي شي بشغلك..."
              className="flex-1 rounded-lg border border-gray-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
            <button onClick={send} disabled={sending}
              className="rounded-lg bg-brand-600 px-3 py-2 text-sm font-bold text-white disabled:opacity-50">
              إرسال
            </button>
          </div>
        </div>
      )}

      {/* ═══ فقاعة الكلام ═══ */}
      {!open && (bubble || activeLine) && (
        <button
          onClick={() => {
            if (bubble) { setBubble(null); return }
            if (activeLine?.link) navigate(activeLine.link)
          }}
          className={`mb-2 block max-w-[16rem] rounded-2xl px-3 py-2 text-right text-[11px] font-bold shadow-lg ${
            bubble ? 'bg-white text-[#0f2040] ring-1 ring-slate-200'
              : activeLine?.urgent ? 'bg-red-600 text-white' : 'bg-amber-400 text-[#3a2a00]'
          }`}
        >
          {bubble || activeLine?.text}
        </button>
      )}

      {/* ═══ الكيان نفسه ═══ */}
      <div
        onPointerDown={onPointerDown}
        onPointerMove={onPointerMove}
        onPointerUp={onPointerUp}
        title="اسحبني، أو اضغط تشوف واجباتك"
        className={`relative h-[86px] w-[76px] cursor-grab active:cursor-grabbing ${mood === 'ANGRY' ? 'entity-alert' : 'entity-float'}`}
        style={{ transform: `scaleX(${facing})` }}
      >
        {photo ? (
          <img src={fileUrl(photo)} alt="" draggable={false}
            className="h-full w-full rounded-2xl object-cover shadow-xl ring-2 ring-white/60" />
        ) : (
          // ⚠️ الشخصية ما انولدت بعد — الكيان يشتغل بإيموجي الموظف
          // المعتمد أصلاً بدل ما يختفي. الميزة ما تنتظر التوليد.
          <div className={`flex h-full w-full items-center justify-center rounded-2xl text-3xl shadow-xl ring-2 ring-white/60 ${moodHeader(mood)}`}>
            {employee.attendanceIcon || '🤖'}
          </div>
        )}
        {urgentCount > 0 && (
          <span className="absolute -top-1 -right-1 flex h-5 min-w-[20px] items-center justify-center rounded-full bg-red-600 px-1 text-[10px] font-black text-white ring-2 ring-white">
            {urgentCount}
          </span>
        )}
      </div>
    </div>
  )
}

/** صورة الملامح حسب المزاج — وترجع لهادئ لو الملمح المطلوب ما انولد. */
function moodImage(b: EntityBriefing | null, mood: string): string | undefined {
  if (!b || b.characterState !== 'READY') return undefined
  if (mood === 'ANGRY') return b.angryUrl || b.calmUrl
  if (mood === 'HAPPY') return b.happyUrl || b.calmUrl
  return b.calmUrl
}

function moodHeader(mood: string): string {
  if (mood === 'ANGRY') return 'bg-gradient-to-br from-red-500 to-red-800'
  if (mood === 'WATCHING') return 'bg-gradient-to-br from-amber-500 to-amber-700'
  return 'bg-gradient-to-br from-brand-500 to-brand-800'
}

function moodTitle(mood: string): string {
  if (mood === 'ANGRY') return '⛔ عندك تأخير — لازم نحچي'
  if (mood === 'WATCHING') return '👀 آني أراقب شغلك'
  return '🙂 شغلك نظيف — كمّل هيچ'
}
