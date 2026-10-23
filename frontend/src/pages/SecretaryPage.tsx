import { useEffect, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
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
      <p className="sec-empty">السكرتير خاص بالمالك ومدير النظام.</p>
      <Link to="/">ارجع للرئيسية</Link>
    </main>
  )
}

function Secretary() {
  const { employee } = useSession()
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
  const [view, setView] = useState<'home' | 'reminders'>('home')
  const [deferred, setDeferred] = useState<string[]>([])
  const [selected, setSelected] = useState<Reminder | null>(null)
  const [at, setAt] = useState('')
  const [note, setNote] = useState('')
  const [declining, setDeclining] = useState(false)
  const recorder = useRef<MediaRecorder | null>(null)
  const cancelled = useRef(false)
  const mounted = useRef(true)
  const end = useRef<HTMLDivElement>(null)

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
  }, [chats, busy])
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
      content: '#231536',
    })
    head('meta[name="apple-mobile-web-app-title"]', 'meta', {
      name: 'apple-mobile-web-app-title',
      content: 'السكرتير',
    })
    document.title = 'السكرتير'
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
  const send = async () => {
    const text = input.trim()
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
      !proposal.fields.title.trim()
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
    if (!selected || busy || (declining ? !note.trim() : !at)) return
    setBusy(true)
    setError('')
    try {
      await apiRequest(
        `/secretary/reminders/${encodeURIComponent(selected.id)}/${declining ? 'decline' : 'schedule'}`,
        {
          method: 'POST',
          body: JSON.stringify(
            declining ? { reason: note.trim() } : { at, note },
          ),
        },
      )
      setSelected(null)
      setReload((n) => n + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  const items = reminders?.items.filter((x) => !deferred.includes(x.id)) ?? []
  const fields = (key: keyof SecretaryProposal['fields'], value: string) => {
    if (proposal)
      setProposal({ ...proposal, fields: { ...proposal.fields, [key]: value } })
  }

  return (
    <main className="secretary" dir="rtl">
      <div className="sec-shell">
        <header className="sec-header">
          <img
            src={`${import.meta.env.BASE_URL}secretary-icon.svg`}
            alt="السكرتير"
          />
          <div>
            <small>مساحتك الخاصة</small>
            <b>السكرتير</b>
          </div>
          <Link to="/" className="sec-system">
            النظام ↗
          </Link>
        </header>
        {error && (
          <p role="alert" className="sec-error">
            {error}
          </p>
        )}
        {view === 'home' ? (
          <>
            <section className="sec-welcome">
              <small>أهلاً {employee?.name.split(' ')[0]}</small>
              <h1>
                يومك أوضح.
                <br />
                <em>وقتك إلك.</em>
              </h1>
              <p>اسألني عن الشغل، أو خلّيني أرتّب مهمة.</p>
            </section>
            <div className="sec-voice">
              <button
                aria-label={recording ? 'وقف التسجيل وأرسل' : 'ابدأ التسجيل'}
                aria-pressed={recording}
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
              <strong aria-live="polite">
                {recording
                  ? `دا أسمعك · ${seconds} ثانية`
                  : openingMic
                    ? 'دا أفتح المايك…'
                    : 'ضغطة واحچيلي'}
              </strong>
              <span>
                {recording ? 'اضغط مرة ثانية حتى ترسل' : 'أو اكتب طلبك بالأسفل'}
              </span>
              {recording && (
                <button className="sec-cancel-record" onClick={cancelVoice}>
                  إلغاء التسجيل
                </button>
              )}
            </div>
            <button
              className="sec-reminder-banner"
              onClick={() => setView('reminders')}
            >
              <span className="sec-star">✦</span>
              <span>
                <b>تذكيراتي</b>
                <small>مقابلاتك والطلبات اللي تحتاجك</small>
              </span>
              <strong>{reminders ? items.length : '—'}</strong>
              <span>←</span>
            </button>
            <section className="sec-summary">
              <div className="sec-heading">
                <h2>نبض اليوم</h2>
                <span>من النظام</span>
              </div>
              {today ? (
                today.summary.lines.length ? (
                  today.summary.lines.map((x, i) => (
                    <p key={i}>
                      <span>0{i + 1}</span>
                      {x}
                    </p>
                  ))
                ) : (
                  <p>ماكو مستجدات بالملخص هسه.</p>
                )
              ) : (
                <p>دا أجيب ملخص اليوم…</p>
              )}
            </section>
            <section className="sec-conversation">
              <div className="sec-heading">
                <h2>بينك وبين السكرتير</h2>
                <span>المحادثة</span>
              </div>
              {!chats.length && (
                <p className="sec-empty">شنو تحب نتابع أول شي اليوم؟</p>
              )}
              {chats.map((chat) => (
                <article key={chat.id} className={`sec-message ${chat.role}`}>
                  <small>
                    {chat.role === 'user' ? 'إنت' : 'السكرتير'} ·{' '}
                    {clock(chat.at)}
                  </small>
                  {chat.transcript && (
                    <p className="sec-transcript">البصمة: {chat.transcript}</p>
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
                      {c.link?.startsWith('/') && !c.link.startsWith('//') && (
                        <Link to={c.link}>افتح التفاصيل ↗</Link>
                      )}
                    </div>
                  ))}
                  {chat.action?.status === 'SENT' && (
                    <small>✓ المهمة انرسلت</small>
                  )}
                </article>
              ))}
              {busy && (
                <p className="sec-thinking" role="status">
                  ✦ دا أفكر وأرتّب الجواب…
                </p>
              )}
              <div ref={end} />
            </section>
            <form
              className="sec-composer"
              onSubmit={(e) => {
                e.preventDefault()
                void send()
              }}
            >
              <textarea
                aria-label="طلبك للسكرتير"
                value={input}
                rows={1}
                onChange={(e) => {
                  setInput(e.target.value)
                  e.currentTarget.style.height = 'auto'
                  e.currentTarget.style.height = `${Math.min(e.currentTarget.scrollHeight, 150)}px`
                }}
                placeholder="اكتبلي شتحتاج…"
              />
              <button
                disabled={busy || recording || openingMic || !input.trim()}
              >
                إرسال ↑
              </button>
            </form>
          </>
        ) : (
          <section className="sec-reminders">
            <button className="sec-back" onClick={() => setView('home')}>
              → ارجع للمحادثة
            </button>
            <h1>
              تذكيراتي<span>{items.length}</span>
            </h1>
            <p>كل موضوع بمكانه، والقرار بإيدك.</p>
            {(Object.keys(groups) as Reminder['kind'][]).map((kind) => (
              <section key={kind}>
                <h2>
                  {groups[kind]}{' '}
                  <span>{items.filter((x) => x.kind === kind).length}</span>
                </h2>
                {items
                  .filter((x) => x.kind === kind)
                  .map((item) => (
                    <article key={item.id} className="sec-reminder-item">
                      <small>{item.who}</small>
                      <h3>{item.title}</h3>
                      <p>{item.reason}</p>
                      {item.guestName && <p>الضيف: {item.guestName}</p>}
                      {item.preferredAt && (
                        <p>
                          الوقت المقترح: {item.preferredAt.replace('T', ' ')}
                        </p>
                      )}
                      {kind === 'meeting' || kind === 'guest' ? (
                        <div className="sec-actions">
                          <button
                            onClick={() => {
                              setSelected(item)
                              setDeclining(false)
                              setAt(item.preferredAt || '')
                              setNote('')
                            }}
                          >
                            حدد موعد
                          </button>
                          <button
                            className="sec-secondary"
                            onClick={() => {
                              setSelected(item)
                              setDeclining(true)
                              setNote('')
                            }}
                          >
                            ارفض
                          </button>
                        </div>
                      ) : (
                        <div className="sec-actions">
                          {item.link?.startsWith('/') &&
                            !item.link.startsWith('//') && (
                              <Link to={item.link}>افتح التفاصيل</Link>
                            )}
                          <button
                            className="sec-secondary"
                            onClick={() => setDeferred((v) => [...v, item.id])}
                          >
                            أجّل
                          </button>
                        </div>
                      )}
                    </article>
                  ))}
                {!items.some((x) => x.kind === kind) && (
                  <p className="sec-empty">ماكو طلبات بهذا القسم.</p>
                )}
              </section>
            ))}
          </section>
        )}
      </div>
      {selected && (
        <div className="sec-overlay">
          <form
            className="sec-dialog"
            role="dialog"
            aria-modal="true"
            aria-label={declining ? 'رفض الطلب' : 'تحديد موعد'}
            onSubmit={(e) => {
              e.preventDefault()
              void decide()
            }}
          >
            <small>{selected.who}</small>
            <h2>{declining ? 'ليش انرفض الطلب؟' : 'خل نرتّب الموعد'}</h2>
            <p>{selected.title}</p>
            {!declining && (
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
              {declining ? 'سبب الرفض' : 'ملاحظة للموظف'}
              <textarea
                required={declining}
                value={note}
                onChange={(e) => setNote(e.target.value)}
              />
            </label>
            <div className="sec-actions">
              <button disabled={busy}>
                {busy ? 'دا أحفظ…' : declining ? 'تأكيد الرفض' : 'تأكيد الموعد'}
              </button>
              <button
                type="button"
                className="sec-secondary"
                disabled={busy}
                onClick={() => setSelected(null)}
              >
                رجوع
              </button>
            </div>
            {error && (
              <p role="alert" className="sec-error">
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
