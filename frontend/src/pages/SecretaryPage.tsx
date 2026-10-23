import { useEffect, useRef, useState } from 'react'
import {
  api,
  apiRequest,
  type SecretaryChat,
  type SecretaryProposal,
  type SecretaryReply,
  type SecretaryToday,
} from '../api'
import { useSession } from '../session'
import './SecretaryPage.css'
import { irisHonorific } from '../irisIdentity'

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
  const { employee } = useSession()
  const honorific = irisHonorific(employee?.certificate)
  const [today, setToday] = useState<SecretaryToday | null>(null)
  const [chats, setChats] = useState<SecretaryChat[]>([])
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
      cancelled.current = true
      const r = recorder.current
      if (r?.state === 'recording') r.stop()
      r?.stream.getTracks().forEach((t) => t.stop())
    }
  }, [])
  useEffect(() => {
    let alive = true
    api
      .secretaryHistory(50)
      .then((x) => {
        if (alive) setChats(x)
      })
      .catch((e) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [])
  useEffect(() => {
    let alive = true
    api
      .secretaryToday()
      .then((x) => {
        if (alive) setToday(x)
      })
      .catch((e) => {
        if (alive) setError(e.message)
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
    if (chats.length || busy)
      end.current?.scrollIntoView({ behavior: 'smooth', block: 'end' })
  }, [chats, busy, view])
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
      content: '#2563eb',
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

  const acceptReply = (text: string, reply: SecretaryReply) => {
    const stamp = new Date().toISOString()
    setChats((v) => [
      ...v,
      {
        id: `user-${reply.messageId}`,
        at: stamp,
        role: 'user',
        text,
        transcript: reply.transcript,
        cards: [],
        action: null,
      },
      {
        id: reply.messageId,
        at: stamp,
        role: 'assistant',
        text: reply.reply,
        transcript: null,
        cards: reply.cards,
        action: reply.proposal
          ? { id: reply.proposal.actionId, status: 'PROPOSED', result: null }
          : null,
      },
    ])
    setProposal(reply.proposal)
    setReload((n) => n + 1)
  }
  const send = async (raw = input) => {
    const text = raw.trim()
    if (!text || busy || recording || openingMic) return
    setBusy(true)
    setError('')
    setInput('')
    try {
      acceptReply(text, await api.secretaryMessage(text))
    } catch (e) {
      setError((e as Error).message)
      setInput(text)
    } finally {
      setBusy(false)
    }
  }
  const voice = async () => {
    if (busy || openingMic) return
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
        setBusy(true)
        try {
          const reply = await api.secretaryVoice(blob)
          if (mounted.current)
            acceptReply(reply.transcript || 'رسالة صوتية', reply)
        } catch (e) {
          if (mounted.current) setError((e as Error).message)
        } finally {
          if (mounted.current) setBusy(false)
        }
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
        if (alive) setAgenda(x)
      })
      .catch((e) => {
        if (alive) setError(e.message)
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
  const saveNote = async (kind: 'AGENDA' | 'DIARY') => {
    if (!body.trim() || busy) return
    setBusy(true)
    setError('')
    try {
      await apiRequest('/secretary/notes', {
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
  return (
    <main className="secretary" dir="rtl">
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
        <header className="sec-topbar">
          <button
            className="sec-menu"
            aria-label="فتح القائمة"
            onClick={() => setDrawer((v) => !v)}
          >
            ☰
          </button>
          <h1>{tabs.find((t) => t.id === view)?.label}</h1>
          <span className="sec-top-status">أهلاً {honorific} · ايريس</span>
        </header>
        {error && (
          <p role="alert" className="sec-error">
            {error}
          </p>
        )}
        {view === 'chat' ? (
          <>
            <div className="sec-chat-scroll">
              <div className="sec-chat-inner">
                {!chats.length && (
                  <div className="sec-chat-empty">
                    <span className="sec-logo">إ</span>
                    <h2>أهلاً {honorific}</h2>
                    <p>اسأل عن الشغل أو وجّه مهمة بصوتك.</p>
                    <div className="sec-suggestions">
                      {[
                        'شنو صاير اليوم؟',
                        'منو متأخر؟',
                        'وجّه مهمة',
                        'شوفلي الحجوزات',
                      ].map((text) => (
                        <button
                          key={text}
                          disabled={busy}
                          onClick={() => void send(text)}
                        >
                          {text}
                          <span>←</span>
                        </button>
                      ))}
                    </div>
                  </div>
                )}
                {chats.map((chat) => (
                  <article key={chat.id} className={`sec-message ${chat.role}`}>
                    <small>
                      {chat.role === 'user' ? 'إنت' : 'ايريس'} ·{' '}
                      {clock(chat.at)}
                    </small>
                    {chat.transcript && (
                      <p className="sec-transcript">
                        البصمة: {chat.transcript}
                      </p>
                    )}
                    <p>{chat.text}</p>
                    {chat.cards.map((c, i) => (
                      <div className="sec-result" key={i}>
                        <span>{c.status}</span>
                        <h3>{c.title}</h3>
                        {c.owner && <p>المسؤول: {c.owner}</p>}
                        {c.lines.map((line, j) => (
                          <p key={j}>{line}</p>
                        ))}
                      </div>
                    ))}
                    {chat.action?.status === 'SENT' && (
                      <small>✓ المهمة انرسلت</small>
                    )}
                  </article>
                ))}
                {busy && (
                  <p className="sec-thinking" role="status">
                    دا أفكر وأرتّب الجواب…
                  </p>
                )}
                <div ref={end} />
              </div>
            </div>
            <div className="sec-composer-wrap">
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
                  aria-label={recording ? 'وقف التسجيل وأرسل' : 'ابدأ التسجيل'}
                  className={recording ? 'sec-mic live' : 'sec-mic'}
                  onClick={voice}
                  disabled={busy || openingMic}
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
                  aria-label="طلبك للسكرتير"
                  rows={1}
                  value={input}
                  onChange={(e) => {
                    setInput(e.target.value)
                    e.currentTarget.style.height = 'auto'
                    e.currentTarget.style.height = `${Math.min(e.currentTarget.scrollHeight, 150)}px`
                  }}
                  placeholder={openingMic ? 'دا أفتح المايك…' : 'اكتب طلبك…'}
                />
                <button
                  className="sec-send"
                  disabled={busy || recording || openingMic || !input.trim()}
                >
                  إرسال ↑
                </button>
              </form>
              <small>المهام ما تنرسل إلا بعد تأكيدك.</small>
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
