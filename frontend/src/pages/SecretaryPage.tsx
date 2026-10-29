import { useEffect, useMemo, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { startAgentRebrand } from '../utils/agentBrand'
import {
  api,
  apiRequest,
  streamSecretaryMessage,
  type SecretaryChat,
  type SecretaryConversation,
  type SecretaryProposal,
  type SecretaryReply,
  type SecretaryToday,
  type ExtraTask,
} from '../api'
import { useSession } from '../session'
import './SecretaryPage.css'
import { irisHonorific } from '../irisIdentity'
type IrisTrace = {
  text: string
  tools: number
  seconds: number
  phase: 'live' | 'done' | 'error' | 'stopped'
  open: boolean
}
type IrisChat = SecretaryChat & { trace?: IrisTrace }

// (دازاي 10-29) «البصمة تضل ظاهرة أسمعها براحتي» — من السجل تنجاب من الخادم وقت ما يضغط.
function VoiceBubble({ chat }: { chat: IrisChat }) {
  const [url, setUrl] = useState(chat.audioUrl ?? '')
  const [loading, setLoading] = useState(false)
  const [failed, setFailed] = useState(false)
  if (url) return <audio className="sec-voice-play" controls src={url} preload="metadata" />
  if (!chat.hasAudio) return null
  return (
    <button
      type="button"
      className="sec-voice-load"
      disabled={loading}
      onClick={async () => {
        setLoading(true)
        setFailed(false)
        try {
          setUrl(URL.createObjectURL(await api.secretaryAudio(chat.id)))
        } catch {
          setFailed(true)
        } finally {
          setLoading(false)
        }
      }}
    >
      {loading ? '…' : failed ? 'ما انجابت — جرّب مرة ثانية' : '▶ اسمع البصمة'}
    </button>
  )
}
const streamClock = () => performance.now()

type Reminder = {
  id: string
  kind: 'meeting' | 'guest' | 'delete_request' | 'leave'
  title: string
  who: string
  reason: string
  guestName: string | null
  preferredAt: string | null
  link: string | null
  createdAt: string
}
type Agenda = {
  id: string
  kind: 'meeting' | 'guest' | 'mine'
  title: string
  who: string
  note: string
  at: string | null
  doneAt: string | null
}
type Diary = {
  id: string
  kind: string
  body: string
  at: string | null
  doneAt: string | null
  createdAt: string
}
type Reminders = { counts: Record<Reminder['kind'], number>; items: Reminder[] }
const groups = {
  meeting: 'طلبات مقابلة',
  guest: 'ضيوف',
  delete_request: 'حذف حجوزات',
  leave: 'إجازات',
}
const clock = (value: string) =>
  new Date(value).toLocaleTimeString('ar-IQ', {
    timeZone: 'Asia/Baghdad',
    hour: '2-digit',
    minute: '2-digit',
  })

function IrisIcon({
  kind,
}: {
  kind: 'calendar' | 'bell' | 'check' | 'send' | 'file' | 'pen'
}) {
  const paths = {
    calendar: (
      <>
        <rect x="3" y="5" width="18" height="16" rx="3" />
        <path d="M7 3v4m10-4v4M3 11h18M8 15h2m4 0h2m-8 3h2" />
      </>
    ),
    bell: (
      <>
        <path d="M18 8a6 6 0 0 0-12 0c0 7-3 8-3 8h18s-3-1-3-8M10 20h4" />
      </>
    ),
    check: (
      <>
        <circle cx="12" cy="12" r="9" />
        <path d="m8 12 3 3 5-6" />
      </>
    ),
    send: (
      <>
        <path d="m21 3-7 18-4-7-7-4 18-7ZM10 14l5-5" />
      </>
    ),
    file: (
      <>
        <path d="M14 3H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V9l-6-6Z" />
        <path d="M14 3v6h6M8 13h8m-8 4h5" />
      </>
    ),
    pen: (
      <>
        <path d="m15 4 5 5M4 20l5-1L21 7a2 2 0 0 0-5-5L4 14l-1 7Z" />
      </>
    ),
  }
  return (
    <svg viewBox="0 0 24 24" aria-hidden="true" className="iris-icon">
      {paths[kind]}
    </svg>
  )
}

export default function SecretaryPage() {
  const { employee } = useSession()
  const allowed = employee?.actualRole === 'OWNER' || employee?.role === 'ADMIN'
  return allowed ? (
    <Secretary key={employee.id} />
  ) : (
    <main className="secretary">
      <p className="sec-empty">ايريس خاص بالمالك ومدير النظام.</p>
    </main>
  )
}

function Secretary() {
  const navigate = useNavigate()
  // (دازاي 10-29) الصفحة برّا Layout، فالاسم الي اختاره المالك (كمال) ما چان يتبدّل هنا
  useEffect(() => { startAgentRebrand() }, [])
  const { employee } = useSession()
  const honorific = irisHonorific(employee?.certificate)
  const [today, setToday] = useState<SecretaryToday | null>(null)
  const [chats, setChats] = useState<IrisChat[]>([])
  const [streaming, setStreaming] = useState(false)
  const [retryAudio, setRetryAudio] = useState<Blob | null>(null)
  // (دازاي 10-29) البصمة ما تنرسل لحالها: تنحفظ هنا، يسمعها، وهو يضغط إرسال.
  const [draftAudio, setDraftAudio] = useState<Blob | null>(null)
  const draftUrl = useMemo(
    () => (draftAudio ? URL.createObjectURL(draftAudio) : ''),
    [draftAudio],
  )
  const holdStart = useRef(0)
  const storageKey = `iris-conversation:${employee?.id}`
  const [conversations, setConversations] = useState<SecretaryConversation[]>(
    [],
  )
  const [conversationId, setConversationId] = useState<string | null>(null)
  const [conversationsReady, setConversationsReady] = useState(false)
  const [conversationListError, setConversationListError] = useState('')
  const [historyLoading, setHistoryLoading] = useState(false)
  const [historyFailed, setHistoryFailed] = useState(false)
  const [historyRetry, setHistoryRetry] = useState(0)
  const [conversationReload, setConversationReload] = useState(0)
  const [conversationBusy, setConversationBusy] = useState(false)
  const [conversationMenu, setConversationMenu] = useState<string | null>(null)
  const [conversationEdit, setConversationEdit] = useState<{
    item: SecretaryConversation
    deleting: boolean
  } | null>(null)
  const [conversationTitle, setConversationTitle] = useState('')
  const initializedConversations = useRef(false)
  const [reminders, setReminders] = useState<Reminders | null>(null)
  const [proposal, setProposal] = useState<SecretaryProposal | null>(null)
  const [input, setInput] = useState('')
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const [reload, setReload] = useState(0)
  const [recording, setRecording] = useState(false)
  const [seconds, setSeconds] = useState(0)
  const [openingMic, setOpeningMic] = useState(false)
  const [view, setView] = useState<'chat' | 'agenda' | 'reminders' | 'diary'>(
    'chat',
  )
  const [drawer, setDrawer] = useState(false)
  const [agenda, setAgenda] = useState<Agenda[]>([])
  const [agendaLoaded, setAgendaLoaded] = useState(false)
  const [urgentTasks, setUrgentTasks] = useState<ExtraTask[] | null>(null)
  const [urgentOpen, setUrgentOpen] = useState(false)
  const [online, setOnline] = useState(() => navigator.onLine)
  const [connected, setConnected] = useState(false)
  const [diary, setDiary] = useState<Diary[]>([])
  const [search, setSearch] = useState('')
  const [agendaBody, setAgendaBody] = useState('')
  const [diaryBody, setDiaryBody] = useState('')
  const body = view === 'agenda' ? agendaBody : diaryBody
  const setBody = (value: string) =>
    view === 'agenda' ? setAgendaBody(value) : setDiaryBody(value)
  const [wide, setWide] = useState(
    () => window.matchMedia('(min-width: 900px)').matches,
  )
  const [newAt, setNewAt] = useState('')
  const [adding, setAdding] = useState(false)
  const [deleteId, setDeleteId] = useState<string | null>(null)
  const [deleteConfirmed, setDeleteConfirmed] = useState(false)
  const [calendarNow, setCalendarNow] = useState(() => Date.now())
  const [selected, setSelected] = useState<Reminder | null>(null)
  const [at, setAt] = useState('')
  const [note, setNote] = useState('')
  const [declining, setDeclining] = useState(false)
  const recorder = useRef<MediaRecorder | null>(null)
  const cancelled = useRef(false)
  const mounted = useRef(true)
  const end = useRef<HTMLDivElement>(null)
  const streamController = useRef<AbortController | null>(null)
  const streamStarted = useRef(0)
  const skipHistory = useRef<string | null>(null)
  const viewportRoot = useRef<HTMLElement>(null)

  useEffect(() => {
    const viewport = window.visualViewport
    const update = () => {
      const root = viewportRoot.current
      if (!root) return
      root.style.setProperty(
        '--iris-viewport-height',
        `${viewport?.height ?? window.innerHeight}px`,
      )
      root.style.setProperty(
        '--iris-viewport-top',
        `${viewport?.offsetTop ?? 0}px`,
      )
      const keyboard =
        document.activeElement?.matches('textarea, input') &&
        (viewport?.height ?? window.innerHeight) <
          document.documentElement.clientHeight - 120
      root.classList.toggle('iris-keyboard-open', Boolean(keyboard))
    }
    update()
    viewport?.addEventListener('resize', update)
    viewport?.addEventListener('scroll', update)
    window.addEventListener('resize', update)
    document.addEventListener('focusin', update)
    document.addEventListener('focusout', update)
    return () => {
      viewport?.removeEventListener('resize', update)
      viewport?.removeEventListener('scroll', update)
      window.removeEventListener('resize', update)
      document.removeEventListener('focusin', update)
      document.removeEventListener('focusout', update)
    }
  }, [])

  useEffect(() => {
    const update = () => setOnline(navigator.onLine)
    window.addEventListener('online', update)
    window.addEventListener('offline', update)
    return () => {
      window.removeEventListener('online', update)
      window.removeEventListener('offline', update)
    }
  }, [])
  useEffect(() => {
    let alive = true
    api
      .getExtraTasks()
      .then((rows) => {
        if (alive)
          setUrgentTasks(
            rows.filter(
              (x) =>
                x.priority === 'URGENT' &&
                (x.status === 'NEW' || x.status === 'IN_PROGRESS'),
            ),
          )
      })
      .catch((e) => {
        if (alive) {
          setUrgentTasks(null)
          setError(e.message)
        }
      })
    return () => {
      alive = false
    }
  }, [reload])

  useEffect(() => {
    const media = window.matchMedia('(min-width: 900px)')
    const update = () => setWide(media.matches)
    media.addEventListener('change', update)
    return () => media.removeEventListener('change', update)
  }, [])

  useEffect(() => {
    const timer = window.setInterval(() => setCalendarNow(Date.now()), 60000)
    return () => window.clearInterval(timer)
  }, [])

  useEffect(() => {
    mounted.current = true
    return () => {
      mounted.current = false
      streamController.current?.abort()
      cancelled.current = true
      const r = recorder.current
      if (r?.state === 'recording') r.stop()
      r?.stream.getTracks().forEach((t) => t.stop())
    }
  }, [])
  useEffect(() => {
    let alive = true
    api
      .secretaryConversations()
      .then((rows) => {
        if (!alive) return
        setConversations(rows)
        setConversationListError('')
        if (!initializedConversations.current) {
          let saved: string | null = null
          try {
            saved = localStorage.getItem(storageKey)
          } catch {
            /* Storage can be disabled. */
          }
          const initialId =
            rows.find((row) => row.id === saved)?.id ?? rows[0]?.id ?? null
          setHistoryLoading(Boolean(initialId))
          setConversationId(initialId)
          initializedConversations.current = true
        }
        setConversationsReady(true)
      })
      .catch((e) => {
        if (alive) {
          setError(e.message)
          setConversationListError(e.message)
        }
      })
    return () => {
      alive = false
    }
  }, [conversationReload, storageKey])
  useEffect(() => {
    let alive = true
    if (!conversationId || skipHistory.current === conversationId) return
    api
      .secretaryHistory(50, conversationId)
      .then((rows) => {
        if (alive) {
          setChats(rows)
          setHistoryLoading(false)
          setHistoryFailed(false)
        }
      })
      .catch((e) => {
        if (alive) {
          setError(e.message)
          setHistoryLoading(false)
          setHistoryFailed(true)
        }
      })
    return () => {
      alive = false
    }
  }, [conversationId, historyRetry])
  useEffect(() => {
    if (!conversationsReady) return
    try {
      if (conversationId) localStorage.setItem(storageKey, conversationId)
      else localStorage.removeItem(storageKey)
    } catch {
      /* A private browser must still be usable. */
    }
  }, [conversationId, conversationsReady, storageKey])
  useEffect(() => {
    let alive = true
    api
      .secretaryToday()
      .then((x) => {
        if (alive) {
          setToday(x)
          setConnected(true)
        }
      })
      .catch((e) => {
        if (alive) {
          setError(e.message)
          setConnected(false)
        }
      })
    apiRequest<Reminders>('/secretary/reminders')
      .then((x) => {
        if (alive) setReminders(x)
      })
      .catch((e) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [reload])
  useEffect(() => {
    const scroller = end.current?.closest<HTMLElement>('.sec-chat-scroll')
    if (chats.length || busy)
      scroller?.scrollTo({ top: scroller.scrollHeight, behavior: 'smooth' })
  }, [chats, busy, view])
  useEffect(() => {
    if (!streaming) return
    const timer = window.setInterval(
      () =>
        setChats((rows) =>
          rows.map((chat) =>
            chat.trace?.phase === 'live'
              ? {
                  ...chat,
                  trace: {
                    ...chat.trace,
                    seconds: Math.floor(
                      (streamClock() - streamStarted.current) / 1000,
                    ),
                  },
                }
              : chat,
          ),
        ),
      1000,
    )
    return () => window.clearInterval(timer)
  }, [streaming])
  // (دازاي 10-29) «بس اطلع من التطبيق يوكف» — الخادم يكمل الجواب، فمن يرجع نجيب السجل.
  useEffect(() => {
    const onVisible = () => {
      if (document.visibilityState !== 'visible' || streamController.current) return
      skipHistory.current = null
      setHistoryRetry((n) => n + 1)
    }
    document.addEventListener('visibilitychange', onVisible)
    return () => document.removeEventListener('visibilitychange', onVisible)
  }, [])
  useEffect(() => {
    if (!recording) return
    const timer = window.setInterval(() => setSeconds((n) => n + 1), 1000)
    return () => window.clearInterval(timer)
  }, [recording])
  useEffect(() => {
    const title = document.title
    const edits: (() => void)[] = []
    const head = (
      selector: string,
      tag: 'link' | 'meta',
      attributes: Record<string, string>,
    ) => {
      const existing = document.head.querySelector(selector)
      const node = existing || document.createElement(tag)
      const old = Object.keys(attributes).map((key) => [
        key,
        node.getAttribute(key),
      ])
      Object.entries(attributes).forEach(([key, value]) =>
        node.setAttribute(key, value),
      )
      if (!existing) document.head.append(node)
      edits.push(() => {
        if (!existing) node.remove()
        else
          old.forEach(([key, value]) =>
            value === null
              ? node.removeAttribute(key!)
              : node.setAttribute(key!, value!),
          )
      })
    }
    const base = import.meta.env.BASE_URL
    head('link[rel="manifest"]', 'link', {
      rel: 'manifest',
      href: `${base}secretary.webmanifest`,
    })
    head('link[rel="apple-touch-icon"]', 'link', {
      rel: 'apple-touch-icon',
      href: `${base}secretary-icon-180.png`,
    })
    head('meta[name="viewport"]', 'meta', {
      name: 'viewport',
      content:
        'width=device-width, initial-scale=1, maximum-scale=1, viewport-fit=cover',
    })
    head('meta[name="theme-color"]', 'meta', {
      name: 'theme-color',
      content: '#101013',
    })
    head('meta[name="apple-mobile-web-app-title"]', 'meta', {
      name: 'apple-mobile-web-app-title',
      content: 'ايريس',
    })
    document.title = 'ايريس'
    return () => {
      edits.reverse().forEach((restore) => restore())
      document.title = title
    }
  }, [])

  const acceptReply = (
    text: string,
    reply: SecretaryReply,
    pendingId: string,
    liveId: string,
    audio: boolean,
  ) => {
    const stamp = new Date().toISOString()
    setChats((v) => [
      ...v.map(
        (chat): IrisChat =>
          chat.id === pendingId
            ? {
                ...chat,
                text: audio ? 'رسالة صوتية' : text,
                transcript: reply.transcript,
              }
            : chat.id !== liveId
              ? chat
              : {
                  id: reply.messageId,
                  at: stamp,
                  role: 'assistant',
                  text: reply.reply,
                  transcript: null,
                  cards: reply.cards,
                  action: reply.proposal
                    ? {
                        id: reply.proposal.actionId,
                        status: 'PROPOSED',
                        result: null,
                      }
                    : null,
                  trace: chat.trace
                    ? {
                        ...chat.trace,
                        phase: 'done',
                        open: false,
                        seconds: Math.floor(
                          (streamClock() - streamStarted.current) / 1000,
                        ),
                      }
                    : undefined,
                },
      ),
    ])
    setProposal(reply.proposal)
    setReload((n) => n + 1)
    if (reply.conversationId && !conversationId) {
      skipHistory.current = reply.conversationId
      setHistoryLoading(false)
      setConversationId(reply.conversationId)
    }
  }
  const conversationLocked =
    busy ||
    recording ||
    openingMic ||
    conversationBusy ||
    historyLoading ||
    !conversationsReady
  const conversationSwitchLocked =
    busy || recording || openingMic || conversationBusy || !conversationsReady
  const selectConversation = (id: string | null) => {
    if (id !== conversationId) {
      skipHistory.current = null
      setRetryAudio(null)
      setHistoryLoading(Boolean(id))
      setHistoryFailed(false)
      setChats([])
      setProposal(null)
      setInput('')
      setConversationId(id)
    }
    setView('chat')
    setDrawer(false)
    setConversationMenu(null)
    setError('')
  }
  const newConversation = async () => {
    if (conversationLocked) return
    setConversationBusy(true)
    setError('')
    try {
      const item = await api.secretaryNewConversation()
      setConversations((rows) => [item, ...rows])
      selectConversation(item.id)
      setConversationReload((n) => n + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setConversationBusy(false)
    }
  }
  const saveConversation = async () => {
    if (!conversationEdit || conversationLocked) return
    const { item, deleting } = conversationEdit
    const title = conversationTitle.trim()
    if (!deleting && (!title || [...title].length > 80)) return
    setConversationBusy(true)
    setError('')
    try {
      if (deleting) {
        await api.secretaryDeleteConversation(item.id)
        const remaining = conversations.filter((row) => row.id !== item.id)
        setConversations(remaining)
        if (conversationId === item.id)
          selectConversation(remaining[0]?.id ?? null)
      } else {
        await api.secretaryRenameConversation(item.id, title)
        setConversations((rows) =>
          rows.map((row) => (row.id === item.id ? { ...row, title } : row)),
        )
      }
      setConversationEdit(null)
      setConversationReload((n) => n + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setConversationBusy(false)
    }
  }
  const runStream = async (text: string, audio?: Blob) => {
    if (streamController.current || conversationLocked || historyFailed) return
    const abort = new AbortController()
    streamController.current = abort
    streamStarted.current = streamClock()
    setStreaming(true)
    setBusy(true)
    setError('')
    setRetryAudio(null)
    if (!audio) setInput('')
    const pendingId = `local-${crypto.randomUUID()}`
    const liveId = `live-${crypto.randomUUID()}`
    const at = new Date().toISOString()
    setChats((v) => [
      ...v,
      {
        id: pendingId,
        at,
        role: 'user',
        text: audio ? 'رسالة صوتية — دا أحوّلها لنص…' : text,
        transcript: null,
        audioUrl: audio ? URL.createObjectURL(audio) : undefined,
        cards: [],
        action: null,
      },
      {
        id: liveId,
        at,
        role: 'assistant',
        text: '',
        transcript: null,
        cards: [],
        action: null,
        trace: { text: '', tools: 0, seconds: 0, phase: 'live', open: true },
      },
    ])
    try {
      await streamSecretaryMessage(
        {
          text: audio ? undefined : text,
          audio,
          conversationId: conversationId ?? undefined,
        },
        (event) => {
          if (!mounted.current || abort.signal.aborted) return
          if (event.type === 'done') {
            acceptReply(text, event.message, pendingId, liveId, Boolean(audio))
            return
          }
          setChats((rows) =>
            rows.map((chat) => {
              if (chat.id !== liveId || !chat.trace) return chat
              if (event.type === 'text') return { ...chat, text: event.text }
              if (event.type === 'thinking')
                return {
                  ...chat,
                  trace: { ...chat.trace, text: chat.trace.text + event.text },
                }
              if (event.type === 'tool')
                return {
                  ...chat,
                  trace: {
                    ...chat.trace,
                    tools: chat.trace.tools + 1,
                    text: `${chat.trace.text}\n🔎 checking ${event.name}…\n`,
                  },
                }
              return chat
            }),
          )
        },
        abort.signal,
      )
    } catch (e) {
      if (!mounted.current) return
      setError(
        abort.signal.aborted
          ? 'وقّفت الرد. السؤال ممكن محفوظ بالخادم؛ ما أعدنا إرساله.'
          : (e as Error).message,
      )
      setChats((rows) =>
        rows.map((chat) =>
          chat.id === liveId && chat.trace
            ? {
                ...chat,
                trace: {
                  ...chat.trace,
                  phase: abort.signal.aborted ? 'stopped' : 'error',
                  open: false,
                  seconds: Math.floor(
                    (streamClock() - streamStarted.current) / 1000,
                  ),
                },
              }
            : chat.id === pendingId && audio
              ? { ...chat, text: 'رسالة صوتية — ما اكتمل الرد عليها' }
              : chat,
        ),
      )
      if (audio) setRetryAudio(audio)
      else setInput(text)
    } finally {
      if (mounted.current) {
        setBusy(false)
        setStreaming(false)
        setConversationReload((n) => n + 1)
      }
      if (streamController.current === abort) streamController.current = null
    }
  }
  const send = async (raw = input) => {
    const text = raw.trim()
    if (!text && draftAudio) {
      const blob = draftAudio
      setDraftAudio(null)
      await runStream('', blob)
      return
    }
    if (!text) return
    await runStream(text)
  }
  const voice = async () => {
    if (
      busy ||
      openingMic ||
      conversationBusy ||
      historyLoading ||
      historyFailed ||
      !conversationsReady
    )
      return
    if (recorder.current?.state === 'recording') {
      recorder.current.stop()
      setRecording(false)
      return
    }
    setOpeningMic(true)
    setError('')
    cancelled.current = false
    try {
      const stream = await navigator.mediaDevices.getUserMedia({ audio: true })
      if (!mounted.current) {
        stream.getTracks().forEach((t) => t.stop())
        return
      }
      const r = new MediaRecorder(stream)
      const chunks: Blob[] = []
      recorder.current = r
      r.ondataavailable = (e) => {
        if (e.data.size) chunks.push(e.data)
      }
      r.onstop = async () => {
        stream.getTracks().forEach((t) => t.stop())
        recorder.current = null
        if (cancelled.current || !mounted.current) return
        const blob = new Blob(chunks, { type: r.mimeType })
        if (!blob.size) return
        setDraftAudio(blob)
      }
      r.start()
      setSeconds(0)
      setRecording(true)
    } catch {
      setError('ما گدرت أفتح المايك. اسمح بالمايك من المتصفح أو اكتبلي.')
    } finally {
      setOpeningMic(false)
    }
  }
  const cancelVoice = () => {
    cancelled.current = true
    recorder.current?.stop()
    setRecording(false)
    setSeconds(0)
  }
  const confirm = async () => {
    if (
      !proposal ||
      busy ||
      !proposal.fields.employeeId ||
      !proposal.fields.title.trim() ||
      !proposal.fields.dueAt
    )
      return
    setBusy(true)
    setError('')
    try {
      const r = await api.secretaryConfirm(proposal.actionId, proposal.fields)
      setChats((v) => [
        ...v,
        {
          id: `sent-${proposal.actionId}`,
          at: new Date().toISOString(),
          role: 'assistant',
          text: 'انرسلت المهمة.',
          transcript: null,
          cards: [r.card],
          action: { id: proposal.actionId, status: 'SENT', result: null },
        },
      ])
      setProposal(null)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  const cancelProposal = async () => {
    if (!proposal || busy) return
    setBusy(true)
    try {
      await api.secretaryCancel(proposal.actionId)
      setProposal(null)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  const decide = async () => {
    if (!selected || busy) return
    const meeting = selected.kind === 'meeting' || selected.kind === 'guest'
    if (meeting && (declining ? !note.trim() : !at)) return
    if (selected.kind === 'delete_request' && !declining && !deleteConfirmed) {
      setDeleteConfirmed(true)
      return
    }
    setBusy(true)
    setError('')
    try {
      if (meeting)
        await apiRequest(
          `/secretary/reminders/${encodeURIComponent(selected.id)}/${declining ? 'decline' : 'schedule'}`,
          {
            method: 'POST',
            body: JSON.stringify(
              declining ? { reason: note.trim() } : { at, note },
            ),
          },
        )
      else {
        const id = selected.id.substring(selected.id.indexOf(':') + 1)
        if (selected.kind === 'leave')
          await api.decideLeave(id, !declining, note)
        else await api.decideBookingDelete(id, !declining, note)
      }
      setSelected(null)
      setDeleteConfirmed(false)
      setReload((n) => n + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  useEffect(() => {
    let alive = true
    apiRequest<Agenda[]>('/secretary/agenda')
      .then((x) => {
        if (alive) {
          setAgenda(x)
          setAgendaLoaded(true)
        }
      })
      .catch((e) => {
        if (alive) {
          setError(e.message)
          setAgendaLoaded(false)
        }
      })
    return () => {
      alive = false
    }
  }, [reload])
  useEffect(() => {
    let alive = true
    apiRequest<Diary[]>(
      `/secretary/notes?kind=DIARY&q=${encodeURIComponent(search)}`,
    )
      .then((x) => {
        if (alive) setDiary(x)
      })
      .catch((e) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [reload, search])
  // 👁️ (دازاي 10-29) مذكرة بيها اسم موظف ← كمال يتحقق من أرقامه ويتابعه
  const [watch, setWatch] = useState<{ employee: string; summary: string; verdict: string } | null>(null)
  const saveNote = async (kind: 'AGENDA' | 'DIARY') => {
    if (!body.trim() || busy) return
    setBusy(true)
    setError('')
    setWatch(null)
    try {
      const saved = await apiRequest<{ watch?: { employee: string; summary: string; verdict: string } }>('/secretary/notes', {
        method: 'POST',
        body: JSON.stringify({
          kind,
          body: body.trim(),
          ...(kind === 'AGENDA' && newAt ? { at: newAt } : {}),
        }),
      })
      setBody('')
      setNewAt('')
      setAdding(false)
      setWatch(saved?.watch ?? null)
      setReload((n) => n + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  const changeNote = async (id: string, done?: boolean) => {
    if (busy) return
    setBusy(true)
    setError('')
    try {
      await apiRequest(
        `/secretary/notes/${encodeURIComponent(id)}${done === undefined ? '' : '/done'}`,
        {
          method: done === undefined ? 'DELETE' : 'PUT',
          ...(done === undefined ? {} : { body: JSON.stringify({ done }) }),
        },
      )
      setDeleteId(null)
      setReload((n) => n + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  const dayKey = (v: string) =>
    new Intl.DateTimeFormat('en-CA', {
      timeZone: 'Asia/Baghdad',
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    }).format(new Date(v))
  const todayKey = dayKey(new Date(calendarNow).toISOString())
  const tomorrowKey = dayKey(new Date(calendarNow + 86400000).toISOString())
  const bucket = (v: string | null) =>
    !v
      ? 'بلا وقت'
      : dayKey(v) <= todayKey
        ? 'اليوم'
        : dayKey(v) === tomorrowKey
          ? 'باچر'
          : 'بعدين'
  const displayDate = (v: string) =>
    new Date(v).toLocaleString('ar-IQ', {
      timeZone: 'Asia/Baghdad',
      dateStyle: 'medium',
      timeStyle: 'short',
    })
  const items = reminders?.items ?? []
  const fields = (key: keyof SecretaryProposal['fields'], value: string) => {
    if (proposal)
      setProposal({ ...proposal, fields: { ...proposal.fields, [key]: value } })
  }

  const tabs = [
    { id: 'chat' as const, label: 'المحادثة', icon: '💬', count: chats.length },
    {
      id: 'agenda' as const,
      label: 'جدول أعمالي',
      icon: '📅',
      count: agenda.filter((x) => !x.doneAt).length,
    },
    {
      id: 'reminders' as const,
      label: 'تذكيراتي',
      icon: '🔔',
      count: items.length,
    },
    { id: 'diary' as const, label: 'مذكراتي', icon: '📔', count: diary.length },
  ]
  const openDecision = (item: Reminder, reject: boolean) => {
    setSelected(item)
    setDeclining(reject)
    setAt(
      item.preferredAt
        ? `${dayKey(item.preferredAt)}T${new Intl.DateTimeFormat('en-GB', { timeZone: 'Asia/Baghdad', hour: '2-digit', minute: '2-digit' }).format(new Date(item.preferredAt))}`
        : '',
    )
    setNote('')
    setDeleteConfirmed(false)
    setError('')
  }
  const todaysMeetings = agenda.filter(
    (x) =>
      (x.kind === 'meeting' || x.kind === 'guest') &&
      x.at &&
      dayKey(x.at) === todayKey,
  )
  return (
    <main ref={viewportRoot} className="secretary" dir="rtl">
      {conversationEdit && (
        <div className="sec-overlay">
          <form
            className="sec-dialog"
            role="dialog"
            aria-modal="true"
            aria-labelledby="iris-conversation-dialog"
            onSubmit={(e) => {
              e.preventDefault()
              void saveConversation()
            }}
          >
            <h2 id="iris-conversation-dialog">
              {conversationEdit.deleting ? 'حذف المحادثة؟' : 'تسمية المحادثة'}
            </h2>
            {conversationEdit.deleting ? (
              <p>
                راح تنحذف «{conversationEdit.item.title}» ورسائلها نهائياً.
                متأكد؟
              </p>
            ) : (
              <label>
                عنوان المحادثة
                <input
                  autoFocus
                  required
                  maxLength={80}
                  value={conversationTitle}
                  onChange={(e) => setConversationTitle(e.target.value)}
                />
              </label>
            )}
            <div className="sec-actions">
              <button
                className={
                  conversationEdit.deleting ? 'sec-danger' : 'sec-primary'
                }
                disabled={
                  conversationBusy ||
                  (!conversationEdit.deleting && !conversationTitle.trim())
                }
              >
                {conversationEdit.deleting ? 'نعم، احذف المحادثة' : 'حفظ الاسم'}
              </button>
              <button
                type="button"
                className="sec-secondary"
                disabled={conversationBusy}
                onClick={() => setConversationEdit(null)}
              >
                إلغاء
              </button>
            </div>
            {error && (
              <p className="sec-error" role="alert">
                {error}
              </p>
            )}
          </form>
        </div>
      )}
      {drawer && (
        <button
          className="sec-backdrop"
          aria-label="غلق القائمة"
          onClick={() => setDrawer(false)}
        />
      )}
      <aside
        className={`sec-drawer ${drawer ? 'open' : ''}`}
        inert={!wide && !drawer}
      >
        <div className="sec-drawer-brand">
          <span className="sec-logo">إ</span>
          <div>
            <strong>ايريس</strong>
            <small>شغلك، بترتيب أوضح</small>
          </div>
          <button
            className="sec-close-drawer"
            aria-label="غلق القائمة"
            onClick={() => setDrawer(false)}
          >
            ×
          </button>
        </div>
        <section className="iris-conversations" aria-label="محادثات ايريس">
          <button
            className="sec-primary iris-new-chat"
            disabled={conversationLocked}
            onClick={newConversation}
          >
            ＋ محادثة جديدة
          </button>
          {!conversationsReady && !conversationListError && (
            <p role="status">دا أجيب محادثاتك…</p>
          )}
          {conversationListError && (
            <p className="sec-error">
              تعذر جلب المحادثات. اضغط تحديث المحادثات حتى تعيد المحاولة.
            </p>
          )}
          <button
            className="iris-conversation-refresh"
            disabled={conversationBusy}
            onClick={() => setConversationReload((n) => n + 1)}
          >
            تحديث المحادثات
          </button>
          {conversationsReady && !conversations.length && (
            <small>ابدأ محادثة جديدة ويا ايريس.</small>
          )}
          <div className="iris-conversation-list">
            {conversations.map((item) => (
              <div
                key={item.id}
                className={`iris-conversation ${conversationId === item.id ? 'selected' : ''}`}
              >
                <button
                  className="iris-conversation-select"
                  disabled={conversationSwitchLocked}
                  aria-current={conversationId === item.id ? 'true' : undefined}
                  onClick={() => selectConversation(item.id)}
                >
                  <strong>{item.title}</strong>
                  <time dateTime={item.updatedAt}>
                    {displayDate(item.updatedAt)}
                  </time>
                </button>
                <button
                  className="iris-conversation-more"
                  aria-label={`خيارات ${item.title}`}
                  aria-expanded={conversationMenu === item.id}
                  disabled={conversationLocked}
                  onClick={() =>
                    setConversationMenu(
                      conversationMenu === item.id ? null : item.id,
                    )
                  }
                >
                  ⋯
                </button>
                {conversationMenu === item.id && (
                  <div className="iris-conversation-actions">
                    <button
                      onClick={() => {
                        setConversationEdit({ item, deleting: false })
                        setConversationTitle(item.title)
                        setConversationMenu(null)
                      }}
                    >
                      تسمية
                    </button>
                    <button
                      onClick={() => {
                        setConversationEdit({ item, deleting: true })
                        setConversationMenu(null)
                      }}
                    >
                      حذف
                    </button>
                  </div>
                )}
              </div>
            ))}
          </div>
        </section>
        <nav aria-label="أقسام ايريس">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              aria-current={view === tab.id ? 'page' : undefined}
              className={view === tab.id ? 'selected' : ''}
              onClick={() => {
                setView(tab.id)
                setDrawer(false)
                setError('')
              }}
            >
              <span>{tab.icon}</span>
              {tab.label}
              <b>{tab.count}</b>
            </button>
          ))}
        </nav>
        <div className="sec-drawer-foot">مساحتك الخاصة للإدارة</div>
      </aside>
      <section className="sec-workspace">
        <header className="sec-topbar iris-topbar">
          <div className="iris-header-tools">
            <button
              className="sec-menu"
              aria-label="فتح القائمة"
              onClick={() => setDrawer((v) => !v)}
            >
              ☰
            </button>
            <button
              className="iris-bell"
              aria-label={`تذكيراتي · ${items.length} طلب`}
              onClick={() => setView('reminders')}
            >
              <IrisIcon kind="bell" />
              {items.length > 0 && <span />}
            </button>
          </div>
          <div className="iris-brand">
            <img
              src={`${import.meta.env.BASE_URL}apple-touch-icon.png`}
              alt="شعار الأماني"
            />
            <h1>ايريس</h1>
            <small>شركة الأماني</small>
          </div>
          <span
            className={`iris-presence ${online && connected ? 'connected' : ''}`}
            role="status"
          >
            <i />
            {!online
              ? 'بدون اتصال'
              : connected
                ? 'متصل بالنظام'
                : 'جار الاتصال'}
          </span>
        </header>
        {error && (
          <p role="alert" className="sec-error">
            {error}
          </p>
        )}
        {view === 'chat' ? (
          <>
            <div className="iris-stat-cards">
              <button
                onClick={() => setView('agenda')}
                className="iris-stat meetings"
              >
                <IrisIcon kind="calendar" />
                <span>
                  اجتماعات اليوم
                  <b>{agendaLoaded ? todaysMeetings.length : '—'}</b>
                </span>
                <span aria-hidden="true">‹</span>
              </button>
              <button
                onClick={() => setUrgentOpen(true)}
                className="iris-stat urgent"
              >
                <IrisIcon kind="check" />
                <span>
                  مهام عاجلة<b>{urgentTasks?.length ?? '—'}</b>
                </span>
                <span aria-hidden="true">‹</span>
              </button>
            </div>
            <div className="sec-chat-scroll">
              <div className="sec-chat-inner">
                {historyLoading && <p role="status">دا أجيب رسائل المحادثة…</p>}
                {historyFailed && (
                  <button
                    className="sec-secondary"
                    onClick={() => {
                      setHistoryLoading(true)
                      setHistoryFailed(false)
                      setHistoryRetry((n) => n + 1)
                    }}
                  >
                    أعد تحميل الرسائل
                  </button>
                )}
                {!chats.length && !historyLoading && !historyFailed && (
                  <div className="iris-welcome">
                    <span className="iris-avatar">
                      <img
                        src={`${import.meta.env.BASE_URL}apple-touch-icon.png`}
                        alt=""
                      />
                    </span>
                    <div>
                      <h2>أهلاً {honorific}.</h2>
                      <p>كيف أساعدك اليوم؟</p>
                    </div>
                  </div>
                )}
                {chats.map((chat) => (
                  <article key={chat.id} className={`sec-message ${chat.role}`}>
                    {chat.role === 'assistant' && (
                      <span className="iris-avatar">
                        <img
                          src={`${import.meta.env.BASE_URL}apple-touch-icon.png`}
                          alt=""
                        />
                      </span>
                    )}
                    <small>{clock(chat.at)}</small>
                    {chat.trace && (
                      <div className="iris-thinking" dir="ltr">
                        <button
                          type="button"
                          aria-expanded={chat.trace.open}
                          onClick={() =>
                            setChats((rows) =>
                              rows.map((row) =>
                                row.id === chat.id && row.trace
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
                          {chat.trace.phase === 'live'
                            ? `Thinking · ${chat.trace.seconds}s · ${chat.trace.tools} tools`
                            : `${chat.trace.phase === 'done' ? 'Thought for' : chat.trace.phase === 'stopped' ? 'Stopped after' : 'Interrupted after'} ${chat.trace.seconds}s · ${chat.trace.tools} tools`}
                        </button>
                        {chat.trace.open && (
                          <pre>{chat.trace.text || 'Waiting for Iris…'}</pre>
                        )}
                      </div>
                    )}
                    {chat.trace &&
                      (chat.trace.phase === 'error' ||
                        chat.trace.phase === 'stopped') && (
                        <small className="iris-stream-warning">
                          الرد غير مكتمل. إيقاف البث أو انقطاعه ما يضمن إلغاء
                          إجراء بالخادم؛ راجع المحادثة قبل إعادة الطلب.
                        </small>
                      )}
                    <VoiceBubble key={chat.id} chat={chat} />
                    {chat.transcript && (
                      <p className="sec-transcript">
                        البصمة: {chat.transcript}
                      </p>
                    )}
                    <p>{chat.text}</p>
                    {chat.cards.map((c, i) => (
                      <div className="sec-result" key={i}>
                        <span>{c.status}</span>
                        <h3>
                          <IrisIcon
                            kind={
                              /اجتماع|موعد/.test(c.title) ? 'calendar' : 'file'
                            }
                          />
                          {c.title}
                        </h3>
                        {c.owner && <p>المسؤول: {c.owner}</p>}
                        {c.lines.map((line, j) => (
                          <p key={j}>{line}</p>
                        ))}
                        {c.link && (
                          <button
                            className="iris-card-action"
                            onClick={() =>
                              c.link!.startsWith('/secretary/files/')
                                ? void api
                                    .downloadSecretaryFile(
                                      c.link!,
                                      c.title.replace(/^📊\s*/, '') + '.xlsx',
                                    )
                                    .catch((e: Error) => window.alert(e.message))
                                : navigate(c.link!)
                            }
                          >
                            <IrisIcon kind="file" />
                            {c.link.startsWith('/secretary/files/')
                              ? 'نزّل الملف'
                              : 'افتح'}
                          </button>
                        )}
                        {/اجتماع|موعد/.test(c.title) && (
                          <button
                            className="iris-card-action"
                            onClick={() => setView('agenda')}
                          >
                            <IrisIcon kind="calendar" />
                            عرض جدول الأعمال
                          </button>
                        )}
                      </div>
                    ))}
                    {chat.action?.status === 'SENT' && (
                      <small>✓ المهمة انرسلت</small>
                    )}
                  </article>
                ))}
                <div ref={end} />
              </div>
            </div>
            <div className="sec-composer-wrap">
              {retryAudio && (
                <div className="iris-audio-retry" role="status">
                  <span>
                    نسخة البصمة لإعادة المحاولة موجودة بالذاكرة مؤقتاً.
                  </span>
                  <button
                    type="button"
                    disabled={conversationLocked || historyFailed}
                    onClick={() => void runStream('', retryAudio)}
                  >
                    إعادة إرسال البصمة
                  </button>
                  <button
                    type="button"
                    disabled={streaming}
                    onClick={() => setRetryAudio(null)}
                  >
                    احذف البصمة المؤقتة
                  </button>
                </div>
              )}
              <div className="iris-quick-actions" aria-label="اقتراحات سريعة">
                {[
                  {
                    label: 'ترتيب اجتماع',
                    prompt: 'رتب لي اجتماع',
                    icon: 'calendar' as const,
                  },
                  {
                    label: 'تلخيص التقارير',
                    prompt: 'لخّص لي تقارير اليوم',
                    icon: 'file' as const,
                  },
                  {
                    label: 'إعداد رسالة',
                    prompt: 'ساعدني بإعداد رسالة',
                    icon: 'pen' as const,
                  },
                ].map((action) => (
                  <button
                    key={action.label}
                    disabled={conversationLocked || historyFailed}
                    onClick={() => void send(action.prompt)}
                  >
                    <IrisIcon kind={action.icon} />
                    {action.label}
                  </button>
                ))}
              </div>
              {draftAudio && !recording && (
                <div className="sec-voice-draft" role="status">
                  <audio controls src={draftUrl} />
                  <button type="button" onClick={() => setDraftAudio(null)}>
                    احذف
                  </button>
                </div>
              )}
              {recording && (
                <div className="sec-recording" role="status">
                  ● دا أسمعك · {seconds} ثانية
                  <button onClick={cancelVoice}>إلغاء التسجيل</button>
                </div>
              )}
              <form
                className="sec-composer"
                onSubmit={(e) => {
                  e.preventDefault()
                  void send()
                }}
              >
                <button
                  type="button"
                  aria-pressed={recording}
                  aria-label={recording ? 'وقف التسجيل' : 'اضغط وسجّل'}
                  className={recording ? 'sec-mic live' : 'sec-mic'}
                  // اضغط وخلّي إيدك = يسجّل، تشيلها = يوقف (وما يرسل).
                  // ضغطة سريعة = يبدأ، وضغطة ثانية = يوقف.
                  onPointerDown={(e) => {
                    e.preventDefault()
                    if (recording) return
                    holdStart.current = Date.now()
                    setDraftAudio(null)
                    void voice()
                  }}
                  onPointerUp={() => {
                    if (!holdStart.current) {
                      if (recorder.current?.state === 'recording') void voice()
                      return
                    }
                    const held = Date.now() - holdStart.current
                    holdStart.current = 0
                    if (held > 400 && recorder.current?.state === 'recording') void voice()
                  }}
                  onContextMenu={(e) => e.preventDefault()}
                  disabled={
                    busy ||
                    openingMic ||
                    conversationBusy ||
                    historyLoading ||
                    historyFailed ||
                    !conversationsReady
                  }
                >
                  {recording ? (
                    '■'
                  ) : (
                    <svg viewBox="0 0 24 24" aria-hidden="true">
                      <rect x="9" y="2" width="6" height="13" rx="3" />
                      <path d="M5 10v2a7 7 0 0 0 14 0v-2M12 19v3M8 22h8" />
                    </svg>
                  )}
                </button>
                <textarea
                  aria-label="رسالتك لايريس"
                  disabled={streaming}
                  rows={1}
                  value={input}
                  onChange={(e) => {
                    setInput(e.target.value)
                    e.currentTarget.style.height = 'auto'
                    e.currentTarget.style.height = `${Math.min(e.currentTarget.scrollHeight, 150)}px`
                  }}
                  placeholder={
                    openingMic ? 'دا أفتح المايك…' : 'اكتب رسالتك هنا…'
                  }
                  // (دازاي 10-29) Enter يرسل، وShift+Enter سطر جديد
                  onKeyDown={(e) => {
                    if (e.key === 'Enter' && !e.shiftKey && !e.nativeEvent.isComposing) {
                      e.preventDefault()
                      e.currentTarget.form?.requestSubmit()
                    }
                  }}
                />
                {streaming ? (
                  <button
                    key="iris-stop"
                    type="button"
                    className="sec-send iris-stop"
                    aria-label="إيقاف الرد"
                    onClick={(event) => {
                      event.preventDefault()
                      streamController.current?.abort()
                    }}
                  >
                    ■
                  </button>
                ) : (
                  <button
                    key="iris-send"
                    className="sec-send"
                    aria-label="إرسال"
                    disabled={
                      conversationLocked ||
                      historyFailed ||
                      (!input.trim() && !draftAudio)
                    }
                  >
                    <IrisIcon kind="send" />
                  </button>
                )}
              </form>
              <small>رسائلك ومتابعاتك بمكان واحد.</small>
            </div>
          </>
        ) : (
          <div className="sec-section-scroll">
            <div className="sec-section-inner">
              {view === 'agenda' && (
                <>
                  <section className="sec-summary">
                    <div className="sec-heading">
                      <h2>نبض اليوم</h2>
                      <span>من النظام</span>
                    </div>
                    {today ? (
                      today.summary.lines.map((line, i) => (
                        <p key={i}>{line}</p>
                      ))
                    ) : (
                      <p>دا أجيب ملخص اليوم…</p>
                    )}
                  </section>
                  <div className="sec-heading">
                    <div>
                      <h2>جدول أعمالي</h2>
                      <p>كل مواعيدك بتوقيت بغداد.</p>
                    </div>
                    <button
                      className="sec-primary"
                      onClick={() => {
                        setAdding((v) => !v)
                        setBody('')
                        setNewAt('')
                      }}
                    >
                      ＋ أضف للجدول
                    </button>
                  </div>
                  {adding && (
                    <form
                      className="sec-panel sec-note-form"
                      onSubmit={(e) => {
                        e.preventDefault()
                        void saveNote('AGENDA')
                      }}
                    >
                      <label>
                        شنو عندك بالجدول؟
                        <textarea
                          required
                          value={body}
                          onChange={(e) => setBody(e.target.value)}
                        />
                      </label>
                      <label>
                        الوقت · اختياري
                        <input
                          type="datetime-local"
                          value={newAt}
                          onChange={(e) => setNewAt(e.target.value)}
                        />
                      </label>
                      <div className="sec-actions">
                        <button disabled={busy || !body.trim()}>
                          حفظ بالجدول
                        </button>
                        <button
                          type="button"
                          className="sec-secondary"
                          onClick={() => setAdding(false)}
                        >
                          إلغاء
                        </button>
                      </div>
                    </form>
                  )}
                  {['اليوم', 'باچر', 'بعدين', 'بلا وقت'].map((label) => (
                    <section className="sec-agenda-group" key={label}>
                      <h3>
                        {label}
                        <span>
                          {agenda.filter((x) => bucket(x.at) === label).length}
                        </span>
                      </h3>
                      {agenda
                        .filter((x) => bucket(x.at) === label)
                        .map((item) => (
                          <article
                            className={`sec-panel ${item.doneAt ? 'done' : ''}`}
                            key={item.id}
                          >
                            <small>
                              {item.at ? displayDate(item.at) : 'بدون موعد'} ·{' '}
                              {item.kind === 'mine'
                                ? 'من جدولك'
                                : item.kind === 'guest'
                                  ? 'ضيف'
                                  : 'مقابلة'}
                            </small>
                            <h3>{item.title}</h3>
                            {item.who && <p>{item.who}</p>}
                            {item.note && <p>{item.note}</p>}
                            {item.kind === 'mine' && (
                              <div className="sec-actions">
                                <button
                                  className="sec-approve"
                                  disabled={busy}
                                  onClick={() =>
                                    void changeNote(item.id, !item.doneAt)
                                  }
                                >
                                  {item.doneAt ? '↶ رجّع للمعلّق' : '✓ خلص'}
                                </button>
                                <button
                                  className="sec-danger"
                                  onClick={() => setDeleteId(item.id)}
                                >
                                  حذف
                                </button>
                              </div>
                            )}
                          </article>
                        ))}
                      {!agenda.some((x) => bucket(x.at) === label) && (
                        <p className="sec-empty">ماكو بنود.</p>
                      )}
                    </section>
                  ))}
                </>
              )}
              {view === 'reminders' && (
                <>
                  <div className="sec-heading">
                    <div>
                      <h2>تذكيراتي</h2>
                      <p>راجع الطلب وخذ قرارك من هنا.</p>
                    </div>
                    <span className="sec-pending">{items.length} ينتظر</span>
                  </div>
                  {(Object.keys(groups) as Reminder['kind'][]).map((kind) => (
                    <section className="sec-agenda-group" key={kind}>
                      <h3>
                        {groups[kind]}
                        <span>
                          {items.filter((x) => x.kind === kind).length}
                        </span>
                      </h3>
                      {items
                        .filter((x) => x.kind === kind)
                        .map((item) => (
                          <article className="sec-panel" key={item.id}>
                            <small>{item.who}</small>
                            <h3>{item.title}</h3>
                            <p>{item.reason}</p>
                            {item.guestName && <p>الضيف: {item.guestName}</p>}
                            {item.preferredAt && (
                              <p>
                                الوقت المقترح: {displayDate(item.preferredAt)}
                              </p>
                            )}
                            <div className="sec-actions">
                              <button
                                className="sec-approve"
                                onClick={() => openDecision(item, false)}
                              >
                                {kind === 'meeting' || kind === 'guest'
                                  ? 'حدد موعد'
                                  : 'وافق'}
                              </button>
                              <button
                                className="sec-danger"
                                onClick={() => openDecision(item, true)}
                              >
                                ارفض
                              </button>
                            </div>
                          </article>
                        ))}
                      {!items.some((x) => x.kind === kind) && (
                        <p className="sec-empty">ماكو طلبات بهذا القسم.</p>
                      )}
                    </section>
                  ))}
                </>
              )}
              {view === 'diary' && (
                <>
                  <div className="sec-heading">
                    <div>
                      <h2>مذكراتي</h2>
                      <p>ايريس يكدر يرجعلها لمن تسأله: «شنو سويت ويا فلان؟»</p>
                    </div>
                  </div>
                  <form
                    className="sec-panel sec-note-form"
                    onSubmit={(e) => {
                      e.preventDefault()
                      void saveNote('DIARY')
                    }}
                  >
                    <label>
                      شنو صار اليوم؟
                      <textarea
                        required
                        value={body}
                        onChange={(e) => setBody(e.target.value)}
                        placeholder="اكتب ملاحظتك حتى ترجع إلها…"
                      />
                    </label>
                    <button
                      className="sec-primary"
                      disabled={busy || !body.trim()}
                    >
                      حفظ المذكرة
                    </button>
                  </form>
                  {watch && (
                    <div className="sec-panel" role="status">
                      <strong>👁️ متابعة {watch.employee}</strong>
                      <p>{watch.summary}</p>
                      <p>{watch.verdict}</p>
                    </div>
                  )}
                  <label className="sec-search">
                    بحث بمذكراتك
                    <input
                      type="search"
                      value={search}
                      onChange={(e) => setSearch(e.target.value)}
                      placeholder="اسم، موضوع، أو كلمة…"
                    />
                  </label>
                  {[...new Set(diary.map((x) => dayKey(x.createdAt)))]
                    .sort()
                    .reverse()
                    .map((day) => (
                      <section className="sec-agenda-group" key={day}>
                        <h3>{day}</h3>
                        {diary
                          .filter((x) => dayKey(x.createdAt) === day)
                          .map((item) => (
                            <article className="sec-panel" key={item.id}>
                              <small>{clock(item.createdAt)}</small>
                              <p className="sec-diary-body">{item.body}</p>
                              <button
                                className="sec-delete-note"
                                onClick={() => setDeleteId(item.id)}
                              >
                                حذف
                              </button>
                            </article>
                          ))}
                      </section>
                    ))}
                  {!diary.length && (
                    <p className="sec-empty">
                      {search
                        ? 'ما لگيت مذكرات بهذا البحث.'
                        : 'بعدك ما عندك مذكرات.'}
                    </p>
                  )}
                </>
              )}
            </div>
          </div>
        )}
      </section>
      {urgentOpen && (
        <div className="sec-overlay">
          <div
            className="sec-dialog"
            role="dialog"
            aria-modal="true"
            aria-label="المهام العاجلة"
          >
            <div className="sec-heading">
              <h2>المهام العاجلة</h2>
              <button
                aria-label="غلق المهام العاجلة"
                className="sec-secondary"
                onClick={() => setUrgentOpen(false)}
              >
                ×
              </button>
            </div>
            <p>المهام العاجلة الجديدة أو قيد التنفيذ من النظام.</p>
            {urgentTasks === null ? (
              <p className="sec-empty">
                ما توفرت بيانات المهام. جرّب تحديث البيانات.
              </p>
            ) : !urgentTasks.length ? (
              <p className="sec-empty">ماكو مهام عاجلة مفتوحة.</p>
            ) : (
              urgentTasks.map((task) => (
                <article key={task.id} className="sec-panel">
                  <h3>{task.title}</h3>
                  <p>{task.description}</p>
                  <p>المسؤول: {task.assignedToName || 'غير متوفر'}</p>
                  <small>
                    {task.dueAt ? displayDate(task.dueAt) : 'بدون موعد'} ·{' '}
                    {task.status === 'NEW' ? 'جديدة' : 'قيد التنفيذ'}
                  </small>
                </article>
              ))
            )}
            <button
              className="sec-primary"
              onClick={() => {
                setReload((n) => n + 1)
              }}
            >
              تحديث البيانات
            </button>
          </div>
        </div>
      )}
      {deleteId && (
        <div className="sec-overlay">
          <div
            className="sec-dialog"
            role="alertdialog"
            aria-modal="true"
            aria-label="تأكيد حذف المذكرة"
          >
            <h2>تحذف هذا البند؟</h2>
            <p>الحذف نهائي وما تكدر ترجّعه.</p>
            <div className="sec-actions">
              <button
                className="sec-danger"
                disabled={busy}
                onClick={() => void changeNote(deleteId)}
              >
                إي، احذفه
              </button>
              <button
                className="sec-secondary"
                disabled={busy}
                onClick={() => setDeleteId(null)}
              >
                رجوع
              </button>
            </div>
            {error && (
              <p role="alert" className="sec-error">
                {error}
              </p>
            )}
          </div>
        </div>
      )}
      {selected && (
        <div className="sec-overlay">
          <form
            className="sec-dialog"
            role="dialog"
            aria-modal="true"
            aria-label="قرار الطلب"
            onSubmit={(e) => {
              e.preventDefault()
              void decide()
            }}
          >
            <small>{selected.who}</small>
            <h2>
              {deleteConfirmed
                ? 'متأكد؟ الحجز ينحذف'
                : declining
                  ? 'رفض الطلب'
                  : selected.kind === 'meeting' || selected.kind === 'guest'
                    ? 'حدد موعد'
                    : 'موافقة على الطلب'}
            </h2>
            <p>{selected.title}</p>
            <p>{selected.reason}</p>
            {!declining &&
              (selected.kind === 'meeting' || selected.kind === 'guest') && (
                <label>
                  التاريخ والوقت · بغداد
                  <input
                    required
                    type="datetime-local"
                    value={at}
                    onChange={(e) => setAt(e.target.value)}
                  />
                </label>
              )}
            <label>
              {declining ? 'سبب الرفض' : 'ملاحظة'}
              <textarea
                required={declining}
                value={note}
                onChange={(e) => setNote(e.target.value)}
              />
            </label>
            {deleteConfirmed && (
              <p className="sec-error">
                الموافقة تحذف الحجز نهائياً. راجع الطلب قبل التأكيد.
              </p>
            )}
            <div className="sec-actions">
              <button
                className={
                  declining || deleteConfirmed ? 'sec-danger' : 'sec-approve'
                }
                disabled={busy}
              >
                {busy
                  ? 'دا أحفظ…'
                  : deleteConfirmed
                    ? 'إي، وافق واحذف الحجز'
                    : declining
                      ? 'تأكيد الرفض'
                      : selected.kind === 'meeting' || selected.kind === 'guest'
                        ? 'تأكيد الموعد'
                        : 'تأكيد الموافقة'}
              </button>
              <button
                type="button"
                className="sec-secondary"
                disabled={busy}
                onClick={() => {
                  setSelected(null)
                  setDeleteConfirmed(false)
                }}
              >
                رجوع
              </button>
            </div>
            {error && (
              <p className="sec-error" role="alert">
                {error}
              </p>
            )}
          </form>
        </div>
      )}
      {proposal && (
        <div className="sec-overlay">
          <form
            className="sec-dialog"
            role="dialog"
            aria-modal="true"
            aria-label="مسودة المهمة"
            onSubmit={(e) => {
              e.preventDefault()
              void confirm()
            }}
          >
            <small>مسودة · تحتاج تأكيدك</small>
            <h2>فهمت منك…</h2>
            <p>{proposal.summary}</p>
            <label>
              الموظف
              <select
                required
                value={proposal.fields.employeeId}
                aria-label="الموظف"
                onChange={(e) => {
                  const candidate = proposal.candidates.find(
                    (c) => c.id === e.target.value,
                  )
                  if (candidate)
                    setProposal({
                      ...proposal,
                      fields: {
                        ...proposal.fields,
                        employeeId: candidate.id,
                        employeeName: candidate.name,
                      },
                    })
                }}
              >
                <option value="">اختار الموظف</option>
                {(proposal.candidates.length
                  ? proposal.candidates
                  : [
                      {
                        id: proposal.fields.employeeId,
                        name: proposal.fields.employeeName,
                      },
                    ]
                ).map((c) => (
                  <option key={c.id} value={c.id}>
                    {c.name}
                  </option>
                ))}
              </select>
            </label>
            <label>
              عنوان المهمة
              <input
                required
                value={proposal.fields.title}
                onChange={(e) => fields('title', e.target.value)}
              />
            </label>
            <label>
              التفاصيل
              <textarea
                value={proposal.fields.description}
                onChange={(e) => fields('description', e.target.value)}
              />
            </label>
            <label>
              الأولوية
              <select
                value={proposal.fields.priority}
                onChange={(e) => fields('priority', e.target.value)}
              >
                <option value="NORMAL">عادية</option>
                <option value="URGENT">عاجلة</option>
              </select>
            </label>
            <label>
              آخر موعد للإنجاز · بغداد *
              <input
                type="datetime-local"
                required
                value={proposal.fields.dueAt}
                onChange={(e) => fields('dueAt', e.target.value)}
              />
            </label>
            <div className="sec-actions">
              <button disabled={busy}>إرسال المهمة</button>
              <button
                disabled={busy}
                type="button"
                className="sec-secondary"
                onClick={cancelProposal}
              >
                إلغاء
              </button>
            </div>
            {error && (
              <p className="sec-error" role="alert">
                {error}
              </p>
            )}
          </form>
        </div>
      )}
    </main>
  )
}
