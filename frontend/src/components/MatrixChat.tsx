import { Fragment, useEffect, useRef, useState, type ReactNode } from 'react'
import {
  api,
  streamMatrixChat,
  type MatrixChatMessage,
  type MatrixChatRow,
  type MatrixLesson,
} from '../api'
import { useSession } from '../session'
import './MatrixChat.css'
import ScreenshotRedactor from './ScreenshotRedactor'

// ═══ 💬 دردشة ماتركس — المالك والمدير (قرار (ع) 10-10) ═══
// «نسولف ويا ويرد مثل أي ذكاء اصطناعي». ماتركس يتذكر المحادثة، ويفتح أدواته
// (قراءة بس) لما يحتاج أرقام الشركة.
//
// (ع) 10-11: «اخذ سكرين من النظام ادزه لماتركس، يكلي هاي ليش، واذا غلط
// يتعلم» — إرفاق صورة (زر/لصق/سحب)، وزر «غلط — صحّحه» يحفظ درس دائم.
const TOOL_AR: Record<string, string> = {
  money_month: 'فلوس الشهر',
  invoices_status: 'الفواتير',
  late_bookings: 'الحجوزات المتأخرة',
  team_performance: 'أداء الفرق',
  employee_profile: 'ملف موظف',
  attendance_today: 'دوام اليوم',
  project_delays: 'تأخير المشاريع',
  procurement_watch: 'شغل المخازن',
  staff_scores: 'التقييم',
  system_guide: 'دليل النظام',
  lookup_code: 'فتح كود',
  service_economics: 'ربحية الخدمات',
}
const SUGGEST = [
  'شلون الشركة اليوم؟',
  'أي خدمة أكثر ربح للساعة؟',
  'شنو لو زدنا فني ثالث للكاميرات؟',
  'منو الموظف الي يحتاج متابعة هالأسبوع وليش؟',
  'شنو المشاريع المتأخرة؟',
]

// تصغير الصورة بالمتصفح: سكرين 4K يصير ~300 كيلو، والكلفة تنزل
function shrink(file: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const img = new Image()
    const url = URL.createObjectURL(file)
    img.onload = () => {
      const scale = Math.min(1, 1600 / img.width)
      const c = document.createElement('canvas')
      c.width = Math.round(img.width * scale)
      c.height = Math.round(img.height * scale)
      c.getContext('2d')?.drawImage(img, 0, 0, c.width, c.height)
      URL.revokeObjectURL(url)
      resolve(c.toDataURL('image/jpeg', 0.85))
    }
    img.onerror = () => {
      URL.revokeObjectURL(url)
      reject(new Error('الصورة ما انقرت'))
    }
    img.src = url
  })
}

// عارض صغير: **عريض**، وأسطر «- » قوائم
function inline(t: string): ReactNode[] {
  return t
    .split(/(\*\*[^*]+\*\*)/g)
    .map((p, i) =>
      p.startsWith('**') && p.endsWith('**') ? (
        <b key={i}>{p.slice(2, -2)}</b>
      ) : (
        <Fragment key={i}>{p}</Fragment>
      ),
    )
}
function Rich({ text }: { text: string }) {
  return (
    <>
      {text.split('\n').map((l, i) => {
        const m = l.match(/^\s*(?:[-•*]|\d+[.)])\s+(.*)$/)
        if (m)
          return (
            <p key={i} className="flex gap-1.5">
              <span className="text-violet-500">•</span>
              <span>{inline(m[1])}</span>
            </p>
          )
        if (/^#{1,3}\s/.test(l))
          return (
            <p key={i} className="mt-1 font-extrabold">
              {inline(l.replace(/^#+\s/, ''))}
            </p>
          )
        return l.trim() === '' ? (
          <div key={i} className="h-2" />
        ) : (
          <p key={i}>{inline(l)}</p>
        )
      })}
    </>
  )
}

export default function MatrixChat() {
  const { employee, permissions } = useSession()
  const admin = employee?.actualRole === 'OWNER' || employee?.role === 'ADMIN'
  // (دازاي 10-29) صاحب «محادثة ماتركس» يحاچيه هم — أدواته بحدود صلاحياته (بالخادم)،
  // والتصحيح والدروس تبقى للمدير لأنها تغيّر ماتركس للكل.
  return admin || permissions.includes('agent_matrix') ? (
    <MatrixChatView admin={admin} />
  ) : (
    <p>محادثة ماتركس تحتاج صلاحية «محادثة ماتركس».</p>
  )
}
type Trace = {
  text: string
  tools: number
  seconds: number
  phase: 'live' | 'done' | 'stopped' | 'error'
  open: boolean
}
type DisplayMessage = MatrixChatMessage & { trace?: Trace; live?: boolean }
const clock = () => performance.now()
function MatrixChatView({ admin }: { admin: boolean }) {
  const [chats, setChats] = useState<MatrixChatRow[]>([])
  const [enabled, setEnabled] = useState(true)
  const [cur, setCur] = useState<string | null>(null)
  const [msgs, setMsgs] = useState<DisplayMessage[]>([])
  const [reload, setReload] = useState(0)
  const [historyLoading, setHistoryLoading] = useState(false)
  const [historyFailed, setHistoryFailed] = useState(false)
  const [historyReload, setHistoryReload] = useState(0)
  const [text, setText] = useState('')
  // image: النسخة المعتمدة (مغطّاة) — هي بس تنرسل. draft: الأصلية، تبقى بالجهاز.
  const [image, setImage] = useState<string | null>(null)
  const [draft, setDraft] = useState<string | null>(null)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState('')
  const [fixing, setFixing] = useState<string | null>(null)
  const [fixText, setFixText] = useState('')
  const [learned, setLearned] = useState<Set<string>>(new Set())
  const [lessons, setLessons] = useState<MatrixLesson[] | null>(null)
  const [zoom, setZoom] = useState<string | null>(null)
  const end = useRef<HTMLDivElement>(null)
  const fileRef = useRef<HTMLInputElement>(null)
  const controller = useRef<AbortController | null>(null)
  const skipHistory = useRef<string | null>(null)
  const mounted = useRef(true)
  const started = useRef(0)

  const loadChats = () => setReload((n) => n + 1)
  useEffect(() => {
    let alive = true
    api
      .getMatrixChats()
      .then((r) => {
        if (alive) {
          setChats(r.chats)
          setEnabled(r.enabled)
        }
      })
      .catch((e) => {
        if (alive) setErr(e.message)
      })
    return () => {
      alive = false
    }
  }, [reload])
  useEffect(() => {
    mounted.current = true
    return () => {
      mounted.current = false
      controller.current?.abort()
    }
  }, [])
  useEffect(() => {
    if (!cur || skipHistory.current === cur) return
    let alive = true
    api
      .getMatrixChat(cur)
      .then((r) => {
        if (alive) {
          setMsgs(r)
          setHistoryLoading(false)
          setHistoryFailed(false)
        }
      })
      .catch((e) => {
        if (alive) {
          setErr(e.message)
          setHistoryLoading(false)
          setHistoryFailed(true)
        }
      })
    return () => {
      alive = false
    }
  }, [cur, historyReload])
  useEffect(() => {
    if (!busy) return
    const timer = window.setInterval(
      () =>
        setMsgs((rows) =>
          rows.map((m) =>
            m.trace?.phase === 'live'
              ? {
                  ...m,
                  trace: {
                    ...m.trace,
                    seconds: Math.floor((clock() - started.current) / 1000),
                  },
                }
              : m,
          ),
        ),
      1000,
    )
    return () => window.clearInterval(timer)
  }, [busy])
  useEffect(() => {
    const scroller = end.current?.parentElement
    scroller?.scrollTo({ top: scroller.scrollHeight, behavior: 'smooth' })
  }, [msgs, busy])

  const attach = async (f: File | null | undefined) => {
    if (busy || !f || !f.type.startsWith('image/')) return
    try {
      setImage(null)
      setDraft(await shrink(f))
      setErr('')
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'الصورة ما انقرت')
    }
  }

  const newChat = () => {
    if (busy) return
    skipHistory.current = null
    setCur(null)
    setMsgs([])
    setImage(null)
    setText('')
    setErr('')
    setHistoryLoading(false)
    setHistoryFailed(false)
    setFixing(null)
  }
  const selectChat = (id: string) => {
    if (busy || id === cur) return
    skipHistory.current = null
    setMsgs([])
    setCur(id)
    setHistoryLoading(true)
    setHistoryFailed(false)
    setErr('')
    setImage(null)
    setFixing(null)
  }

  const send = async (t: string) => {
    const body = t.trim()
    if (
      (!body && !image) ||
      busy ||
      controller.current ||
      historyLoading ||
      historyFailed ||
      !enabled
    )
      return
    setErr('')
    setBusy(true)
    const abort = new AbortController()
    controller.current = abort
    started.current = clock()
    const img = image
    const userId = `local-user-${crypto.randomUUID()}`
    const liveId = `local-answer-${crypto.randomUUID()}`
    const at = new Date().toISOString()
    setMsgs((rows) => [
      ...rows,
      {
        id: userId,
        role: 'USER',
        text: body || 'شنو هاي الشاشة؟',
        steps: null,
        image: img,
        createdAt: at,
      },
      {
        id: liveId,
        role: 'ASSISTANT',
        text: '',
        steps: null,
        createdAt: at,
        live: true,
        trace: { text: '', tools: 0, seconds: 0, phase: 'live', open: true },
      },
    ])
    setText('')
    setImage(null)
    let id = cur
    try {
      if (!id) {
        const c = await api.createMatrixChat(abort.signal)
        if (!mounted.current) return
        if (abort.signal.aborted)
          throw new DOMException('Aborted', 'AbortError')
        id = c.id
        skipHistory.current = c.id
        setCur(c.id)
      }
      window.dispatchEvent(new Event('matrix-mangekyo')) // (ع) 10-13: ماتركس يشتغل — مانغكيو
      await streamMatrixChat(
        id,
        body,
        img ?? undefined,
        (event) => {
          if (!mounted.current || abort.signal.aborted) return
          setMsgs((rows) =>
            rows.map((m) => {
              if (m.id !== liveId || !m.trace) return m
              const trace = m.trace
              if (event.type === 'thinking')
                return {
                  ...m,
                  trace: { ...trace, text: trace.text + event.text },
                }
              if (event.type === 'tool')
                return {
                  ...m,
                  trace: {
                    ...trace,
                    tools: trace.tools + 1,
                    text: `${trace.text}\n🔎 checking ${event.name}…\n`,
                  },
                }
              if (event.type === 'text') return { ...m, text: event.text }
              if (event.type === 'done')
                return {
                  ...event.message,
                  trace: {
                    ...trace,
                    seconds: Math.floor((clock() - started.current) / 1000),
                    phase: 'done',
                    open: false,
                  },
                }
              return m
            }),
          )
        },
        abort.signal,
      )
    } catch (e) {
      if (!mounted.current) return
      const stopped = abort.signal.aborted
      setErr(
        stopped
          ? 'وقّفت الرد. نصك رجع للخانة؛ ممكن السؤال محفوظ بالمحادثة.'
          : e instanceof Error
            ? e.message
            : 'تعذر الإرسال',
      )
      setMsgs((rows) =>
        rows.map((m) =>
          m.id === liveId && m.trace
            ? {
                ...m,
                trace: {
                  ...m.trace,
                  phase: stopped ? 'stopped' : 'error',
                  open: false,
                  seconds: Math.floor((clock() - started.current) / 1000),
                },
              }
            : m,
        ),
      )
      setText(body)
      setImage(img)
    } finally {
      if (mounted.current) {
        setBusy(false)
        loadChats()
      }
      if (controller.current === abort) controller.current = null
    }
  }

  const saveFix = async (msgId: string) => {
    if (!cur || fixText.trim().length < 5) return
    try {
      await api.correctMatrixChat(cur, msgId, fixText.trim())
      setLearned((s) => new Set(s).add(msgId))
      setMsgs((m) => [
        ...m,
        {
          id: 'fix' + msgId,
          role: 'USER',
          text: 'تصحيح: ' + fixText.trim(),
          steps: null,
          createdAt: new Date().toISOString(),
        },
      ])
      setFixing(null)
      setFixText('')
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر حفظ التصحيح')
    }
  }

  const openLessons = () => {
    if (!admin) return
    api
      .getMatrixLessons()
      .then(setLessons)
      .catch(() => setLessons([]))
  }

  return (
    <div
      dir="rtl"
      className="grid h-[calc(100vh-180px)] min-h-[520px] gap-3 lg:grid-cols-[260px_1fr]"
    >
      <aside className="hidden flex-col overflow-hidden rounded-2xl border border-slate-200 bg-white lg:flex">
        <button
          type="button"
          disabled={busy}
          onClick={newChat}
          className="m-3 rounded-xl bg-gradient-to-l from-violet-600 to-fuchsia-600 px-3 py-2 text-sm font-bold text-white"
        >
          ＋ محادثة جديدة
        </button>
        <div className="flex-1 space-y-1 overflow-y-auto px-2 pb-2">
          {chats.map((c) => (
            <div
              key={c.id}
              className={`group flex items-center gap-1 rounded-lg px-2 py-2 text-sm ${cur === c.id ? 'bg-violet-50 font-bold text-violet-900' : 'text-slate-700 hover:bg-slate-50'}`}
            >
              <button
                type="button"
                disabled={busy}
                onClick={() => selectChat(c.id)}
                className="min-w-0 flex-1 truncate text-right"
              >
                {c.title}
              </button>
              <button
                type="button"
                disabled={busy}
                title="حذف"
                onClick={() => {
                  if (window.confirm('تحذف المحادثة؟'))
                    api.deleteMatrixChat(c.id).then(() => {
                      if (cur === c.id) newChat()
                      loadChats()
                    })
                }}
                className="hidden text-xs text-slate-400 hover:text-red-600 group-hover:block"
              >
                ✕
              </button>
            </div>
          ))}
          {chats.length === 0 && (
            <p className="p-3 text-xs text-slate-400">ماكو محادثات بعد.</p>
          )}
        </div>
        {admin && (
<button
          type="button"
          onClick={openLessons}
          className="m-3 mt-0 rounded-xl border border-violet-200 px-3 py-2 text-xs font-bold text-violet-800 hover:bg-violet-50"
        >
          🧠 دروس ماتركس
        </button>)}
      </aside>

      <section
        className="flex min-h-0 flex-col overflow-hidden rounded-2xl border border-slate-200 bg-gradient-to-b from-white to-violet-50/40"
        onDragOver={(e) => e.preventDefault()}
        onDrop={(e) => {
          e.preventDefault()
          void attach(e.dataTransfer.files?.[0])
        }}
      >
        <header className="flex items-center justify-between gap-2 border-b border-slate-100 px-4 py-3">
          <div className="flex min-w-0 items-center gap-2">
            <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-gradient-to-br from-violet-600 to-fuchsia-600 text-lg text-white">
              🤖
            </span>
            <div className="min-w-0">
              <b className="text-[#0f2040]">ماتركس</b>
              <p className="truncate text-[11px] text-slate-500">
                {enabled
                  ? 'يسولف، يقرا السكرين، ويفتش بالنظام بنفسه — وإذا غلط صحّحه ويتعلّم'
                  : 'يحتاج مفتاح هايكو حتى يسولف'}
              </p>
            </div>
          </div>
          {/* الموبايل: القائمة الجانبية مخفية، فالمحادثات هنا */}
          <div className="flex shrink-0 items-center gap-1 lg:hidden">
            <select
              aria-label="محادثات ماتركس"
              disabled={busy}
              value={cur ?? ''}
              onChange={(e) =>
                e.target.value ? selectChat(e.target.value) : newChat()
              }
              className="max-w-[130px] rounded-lg border border-slate-200 px-1 py-1 text-base"
            >
              <option value="">＋ جديدة</option>
              {chats.map((c) => (
                <option key={c.id} value={c.id}>
                  {c.title}
                </option>
              ))}
            </select>
            {admin && (
<button
              type="button"
              onClick={openLessons}
              className="rounded-lg px-2 py-1 text-base"
              title="دروس ماتركس"
            >
              🧠
            </button>)}
          </div>
        </header>

        <div className="flex-1 space-y-3 overflow-y-auto px-4 py-4">
          {historyLoading && <p role="status">دا أجيب المحادثة…</p>}
          {historyFailed && (
            <button
              type="button"
              onClick={() => {
                setHistoryLoading(true)
                setHistoryFailed(false)
                setHistoryReload((n) => n + 1)
              }}
            >
              أعد جلب المحادثة
            </button>
          )}
          {msgs.length === 0 && !busy && (
            <div className="mx-auto mt-6 max-w-lg text-center">
              <p className="text-2xl">👋</p>
              <p className="mt-2 font-bold text-[#0f2040]">
                هلا، شتحب نحچي اليوم؟
              </p>
              <p className="mt-1 text-xs text-slate-500">
                تكدر تلصق سكرين من النظام (Ctrl+V) أو تسحبه هنا، وماتركس يشرحلك
                شنو بيه وليش.
              </p>
              <div className="mt-4 flex flex-wrap justify-center gap-2">
                {SUGGEST.map((s) => (
                  <button
                    key={s}
                    type="button"
                    disabled={!enabled}
                    onClick={() => void send(s)}
                    className="rounded-full border border-violet-200 bg-white px-3 py-1.5 text-xs font-bold text-violet-800 hover:bg-violet-50 disabled:opacity-40"
                  >
                    {s}
                  </button>
                ))}
              </div>
            </div>
          )}
          {msgs.map((m) => (
            <div
              key={m.id}
              className={`flex ${m.role === 'USER' ? 'justify-start' : 'justify-end'}`}
            >
              <div
                className={`max-w-[85%] ${fixing === m.id ? 'w-[min(85%,420px)]' : ''} rounded-2xl px-4 py-2.5 text-sm leading-relaxed shadow-sm ${m.role === 'USER' ? 'rounded-br-md bg-[#0f2040] text-white' : 'rounded-bl-md border border-violet-100 bg-white text-slate-800'}`}
              >
                {m.trace && (
                  <div className="matrix-thinking" dir="ltr">
                    <button
                      type="button"
                      aria-expanded={m.trace.open}
                      onClick={() =>
                        setMsgs((rows) =>
                          rows.map((row) =>
                            row.id === m.id && row.trace
                              ? {
                                  ...row,
                                  trace: {
                                    ...row.trace,
                                    open: !row.trace.open,
                                  },
                                }
                              : row,
                          ),
                        )
                      }
                    >
                      {m.trace.phase === 'live'
                        ? `Thinking · ${m.trace.seconds}s · ${m.trace.tools} tools`
                        : `${m.trace.phase === 'done' ? 'Thought for' : m.trace.phase === 'stopped' ? 'Stopped after' : 'Interrupted after'} ${m.trace.seconds}s · ${m.trace.tools} tools`}
                    </button>
                    {m.trace.open && (
                      <pre>{m.trace.text || 'Waiting for the model…'}</pre>
                    )}
                  </div>
                )}
                {m.image && (
                  <button
                    type="button"
                    onClick={() => setZoom(m.image ?? null)}
                    className="mb-2 block"
                  >
                    <img
                      src={m.image}
                      alt="سكرين"
                      className="max-h-48 rounded-lg border border-white/30"
                    />
                  </button>
                )}
                {m.role === 'USER' ? (
                  <p className="whitespace-pre-wrap">{m.text}</p>
                ) : (
                  <Rich text={m.text} />
                )}
                {m.role === 'ASSISTANT' && !m.live && (
                  <div className="mt-1.5 flex flex-wrap items-center justify-between gap-2 border-t border-slate-100 pt-1 text-[10px] text-slate-400">
                    <span>
                      {m.steps ? (
                        <>
                          🔎 فتّش:{' '}
                          {[...new Set(m.steps.split(','))]
                            .map((s) => TOOL_AR[s] ?? s)
                            .join('، ')}
                        </>
                      ) : (
                        ''
                      )}
                    </span>
                    {!admin ? null : learned.has(m.id) ? (
                      <span className="font-bold text-emerald-600">
                        ✓ ماتركس تعلّم
                      </span>
                    ) : (
                      <button
                        type="button"
                        disabled={busy}
                        onClick={() => {
                          setFixing(m.id)
                          setFixText('')
                        }}
                        className="font-bold text-rose-500 hover:text-rose-700"
                      >
                        👎 غلط — صحّحه
                      </button>
                    )}
                  </div>
                )}
                {fixing === m.id && (
                  <div className="mt-2 space-y-1.5">
                    <textarea
                      autoFocus
                      value={fixText}
                      onChange={(e) => setFixText(e.target.value)}
                      rows={2}
                      placeholder="شنو الصح؟ مثلاً: الحجز بانتظار موافقة الزبون يعني الزبون رايح يرجّع خبر، مو إن الموظف مقصّر"
                      className="w-full rounded-lg border border-rose-200 px-2 py-1.5 text-xs outline-none focus:border-rose-400"
                    />
                    <div className="flex gap-1.5">
                      <button
                        type="button"
                        disabled={fixText.trim().length < 5}
                        onClick={() => void saveFix(m.id)}
                        className="rounded-lg bg-rose-600 px-3 py-1 text-xs font-bold text-white disabled:opacity-40"
                      >
                        علّمه
                      </button>
                      <button
                        type="button"
                        onClick={() => setFixing(null)}
                        className="rounded-lg border px-3 py-1 text-xs text-slate-600"
                      >
                        إلغاء
                      </button>
                    </div>
                  </div>
                )}
              </div>
            </div>
          ))}
          <div ref={end} />
        </div>

        {err && (
          <p
            role="alert"
            className="mx-4 mb-2 rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700"
          >
            {err}
          </p>
        )}
        {image && (
          <div className="mx-3 mb-1 flex items-center gap-2 rounded-xl border border-violet-200 bg-white p-2">
            <img src={image} alt="المرفق" className="h-14 rounded-md border" />
            <span className="flex-1 text-xs text-slate-500">
              الصورة جاهزة — اكتب سؤالك أو دزها هيچ
            </span>
            <button
              type="button"
              onClick={() => setImage(null)}
              className="text-xs font-bold text-slate-400 hover:text-red-600"
            >
              ✕
            </button>
          </div>
        )}
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void send(text)
          }}
          className="flex items-end gap-2 border-t border-slate-100 bg-white p-3"
        >
          <input
            ref={fileRef}
            type="file"
            accept="image/*"
            className="hidden"
            onChange={(e) => {
              void attach(e.target.files?.[0])
              e.target.value = ''
            }}
          />
          <button
            type="button"
            disabled={!enabled || busy}
            onClick={() => fileRef.current?.click()}
            title="إرفاق صورة"
            className="h-11 rounded-xl border border-slate-300 px-3 text-lg disabled:opacity-40"
          >
            📎
          </button>
          <textarea
            aria-label="رسالتك لماتركس"
            value={text}
            onChange={(e) => setText(e.target.value)}
            rows={1}
            disabled={!enabled || busy || historyLoading || historyFailed}
            onKeyDown={(e) => {
              if (e.key === 'Enter' && !e.shiftKey) {
                e.preventDefault()
                void send(text)
              }
            }}
            onPaste={(e) => {
              const f = [...e.clipboardData.files].find((x) =>
                x.type.startsWith('image/'),
              )
              if (f) {
                e.preventDefault()
                void attach(f)
              }
            }}
            placeholder={
              enabled
                ? 'اكتب لماتركس أو الصق سكرين… (Enter للإرسال)'
                : 'ماتركس يحتاج مفتاح هايكو'
            }
            className="max-h-40 min-h-[44px] min-w-0 flex-1 resize-y rounded-xl border border-slate-300 px-3 py-2.5 text-base outline-none focus:border-violet-500 disabled:bg-slate-50"
          />
          {busy ? (
            <button
              key="stop"
              type="button"
              onClick={(event) => {
                event.preventDefault()
                controller.current?.abort()
              }}
              className="h-11 rounded-xl border border-red-300 px-4 font-bold text-red-700"
            >
              إيقاف
            </button>
          ) : (
            <button
              key="send"
              type="submit"
              disabled={
                historyLoading ||
                historyFailed ||
                (!text.trim() && !image) ||
                !enabled
              }
              className="h-11 rounded-xl bg-gradient-to-l from-violet-600 to-fuchsia-600 px-5 text-sm font-bold text-white disabled:opacity-40"
            >
              إرسال
            </button>
          )}
        </form>
      </section>

      {draft && (
        <ScreenshotRedactor
          src={draft}
          onDone={(redacted) => {
            setImage(redacted)
            setDraft(null)
          }}
          onCancel={() => setDraft(null)}
        />
      )}

      {zoom && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4"
          onClick={() => setZoom(null)}
        >
          <img
            src={zoom}
            alt="سكرين"
            className="max-h-full max-w-full rounded-xl"
          />
        </div>
      )}

      {lessons && (
        <div
          className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4"
          onClick={() => setLessons(null)}
        >
          <div
            dir="rtl"
            className="max-h-[80vh] w-full max-w-lg overflow-y-auto rounded-2xl bg-white p-5 shadow-xl"
            onClick={(e) => e.stopPropagation()}
          >
            <div className="mb-3 flex items-center justify-between">
              <p className="text-base font-extrabold text-violet-900">
                🧠 دروس ماتركس ({lessons.length})
              </p>
              <button
                type="button"
                onClick={() => setLessons(null)}
                className="text-xs font-bold text-slate-500"
              >
                ✕ سكّر
              </button>
            </div>
            <p className="mb-3 text-xs text-slate-500">
              كل تصحيح تكتبه يصير درس، وماتركس يلتزم بيه بكل محادثة جاية. إذا
              درس صار قديم أو غلط، طفّيه.
            </p>
            {lessons.length === 0 && (
              <p className="p-4 text-center text-sm text-slate-400">
                ماكو دروس بعد — من يغلط ماتركس اضغط «👎 غلط — صحّحه».
              </p>
            )}
            <div className="space-y-2">
              {lessons.map((l) => (
                <div
                  key={l.id}
                  className="rounded-xl border border-slate-200 p-3"
                >
                  <p className="text-sm font-bold text-slate-800">{l.text}</p>
                  {l.context && (
                    <p className="mt-1 text-[11px] text-slate-400">
                      على سؤال: «{l.context}»
                    </p>
                  )}
                  <div className="mt-1.5 flex items-center justify-between text-[11px] text-slate-400">
                    <span>
                      {l.by ?? '—'} ·{' '}
                      {new Date(l.createdAt).toLocaleDateString('ar-IQ')}
                    </span>
                    <button
                      type="button"
                      onClick={() => {
                        if (window.confirm('تطفّي هذا الدرس؟'))
                          api
                            .disableMatrixLesson(l.id)
                            .then(() =>
                              setLessons(
                                (x) => x?.filter((y) => y.id !== l.id) ?? null,
                              ),
                            )
                      }}
                      className="font-bold text-rose-500 hover:text-rose-700"
                    >
                      طفّيه
                    </button>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
