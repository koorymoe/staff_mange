import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { useSession } from '../session'
import {
  waApi,
  allowedWaAgents,
  type WaAgent,
  type WaConversation,
  type WaDetail,
} from '../waApi'
import AgentActionNotice from '../components/AgentActionNotice'
import './WhatsAppPage.css'
import { pollVisible } from '../utils/poll'

export default function WhatsAppPage() {
  const { employee, permissions } = useSession()
  const allowed = allowedWaAgents(
    employee?.role,
    employee?.actualRole,
    permissions,
  )
  return allowed.length ? (
    <Conversations key={allowed.join(',')} allowed={allowed} />
  ) : (
    <p className="p-8 text-slate-500">
      هاي الشاشة للإدارة ومسؤولي الجودة والحجوزات المخولين.
    </p>
  )
}
function Conversations({ allowed }: { allowed: WaAgent[] }) {
  const { employee } = useSession()
  const [agent, setAgent] = useState<WaAgent>(allowed[0])
  const [rows, setRows] = useState<WaConversation[]>([])
  const [selected, setSelected] = useState('')
  const [detail, setDetail] = useState<WaDetail | null>(null)
  const [status, setStatus] = useState('')
  const [search, setSearch] = useState('')
  const [text, setText] = useState('')
  const [busy, setBusy] = useState(false)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [confirm, setConfirm] = useState<'release' | 'close' | null>(null)
  const [reload, setReload] = useState(0)
  useEffect(() => pollVisible(() => setReload((n) => n + 1), 15000), [])
  useEffect(() => {
    let alive = true
    waApi
      .list(agent, status)
      .then((x) => {
        if (alive) {
          setRows(x.filter((row) => row.agent === agent))
          setSelected((prev) => (x.some((r) => r.id === prev) ? prev : ''))
          setLoading(false)
        }
      })
      .catch((e) => {
        if (alive) {
          setError(e.message)
          setLoading(false)
        }
      })
    return () => {
      alive = false
    }
  }, [agent, status, reload])
  useEffect(() => {
    if (!selected) return
    let alive = true
    waApi
      .detail(selected)
      .then((x) => {
        if (alive && allowed.includes(x.conversation.agent)) setDetail(x)
      })
      .catch((e) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [selected, reload, allowed])
  const act = async (kind: 'takeover' | 'release' | 'close') => {
    if (!selected || busy) return
    setBusy(true)
    setError('')
    try {
      await waApi[kind](selected)
      setConfirm(null)
      setReload((n) => n + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  const send = async () => {
    if (
      !selected ||
      !text.trim() ||
      busy ||
      detail?.conversation.status !== 'HUMAN'
    )
      return
    setBusy(true)
    setError('')
    try {
      await waApi.send(selected, text.trim())
      setText('')
      setReload((n) => n + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  const current = detail?.conversation
  return (
    <div className="wa-page">
      <header className="wa-page-heading">
        <div>
          <h1>محادثات الواتساب</h1>
          <p>راجع الحديث، واعرف منو يتابع الزبون.</p>
        </div>
        {employee?.actualRole === 'OWNER' && (
          <Link to="/whatsapp-sandbox">وضع التجربة ←</Link>
        )}
      </header>
      {error && (
        <p className="wa-error" role="alert">
          {error}
        </p>
      )}
      <div className="wa-grid">
        <aside className="wa-list">
          <label>
            الوكيل
            <select
              aria-label="الوكيل"
              value={agent}
              onChange={(e) => {
                setAgent(e.target.value as WaAgent)
                setSelected('')
                setDetail(null)
                setRows([])
                setLoading(true)
                setError('')
              }}
            >
              {allowed.map((a) => (
                <option key={a} value={a}>
                  {a === 'QUALITY' ? 'وكيل الجودة' : 'وكيل الحجوزات'}
                </option>
              ))}
            </select>
          </label>
          <input
            aria-label="ابحث بالمحادثات"
            type="search"
            placeholder="ابحث بالاسم أو كود الحجز…"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <select
            aria-label="فلتر الحالة"
            value={status}
            onChange={(e) => {
              setStatus(e.target.value)
              setSelected('')
              setDetail(null)
              setRows([])
              setLoading(true)
            }}
          >
            <option value="">كل الحالات</option>
            <option value="HUMAN">تحتاج إنسان</option>
            <option value="BOT">ويا الوكيل</option>
            <option value="CLOSED">مغلقة</option>
          </select>
          {loading && <p className="wa-empty">دا أجيب المحادثات…</p>}
          {rows
            .filter((x) =>
              `${x.customerName} ${x.bookingCode}`.includes(search),
            )
            .map((x) => (
              <button
                key={x.id}
                aria-pressed={x.id === selected}
                onClick={() => {
                  setSelected(x.id)
                  setDetail(null)
                  setText('')
                  setError('')
                }}
              >
                <b>{x.customerName}</b>
                <small>
                  {x.bookingCode} ·{' '}
                  {x.status === 'HUMAN'
                    ? '👤 إنسان'
                    : x.status === 'BOT'
                      ? '🤖 الوكيل'
                      : 'مغلقة'}
                  {x.sandbox ? ' · تجربة' : ''}
                </small>
                <p>{x.lastMessage}</p>
                {x.unread > 0 && <span>{x.unread} غير مقروءة</span>}
              </button>
            ))}
          {!loading && !rows.length && (
            <p className="wa-empty">ماكو محادثات بهالفلتر.</p>
          )}
        </aside>
        <section className="wa-conversation">
          {!selected ? (
            <div className="wa-empty">اختار محادثة حتى تشوف تفاصيلها.</div>
          ) : !current ? (
            <div className="wa-empty">دا أجيب الرسائل…</div>
          ) : (
            <>
              <header>
                <div>
                  <h2>{current.customerName}</h2>
                  <p>
                    {current.bookingCode} ·{' '}
                    {current.status === 'HUMAN'
                      ? '👤 إنسان'
                      : current.status === 'BOT'
                        ? '🤖 الوكيل'
                        : 'مغلقة'}
                    {current.sandbox ? ' · تجربة' : ''}
                  </p>
                </div>
                <div className="wa-toolbar">
                  {current.status === 'BOT' && (
                    <button
                      disabled={busy}
                      onClick={() => void act('takeover')}
                    >
                      استلم المحادثة
                    </button>
                  )}
                  {current.status === 'HUMAN' && (
                    <button
                      disabled={busy}
                      onClick={() => setConfirm('release')}
                    >
                      رجّعها للوكيل
                    </button>
                  )}
                  {current.status !== 'CLOSED' && (
                    <button disabled={busy} onClick={() => setConfirm('close')}>
                      أغلق المحادثة
                    </button>
                  )}
                </div>
              </header>
              <div className="wa-messages">
                {detail.messages.map((m) => (
                  <article
                    key={m.id}
                    className={m.direction === 'IN' ? 'incoming' : 'outgoing'}
                  >
                    <small>
                      {m.author === 'CUSTOMER'
                        ? 'الزبون'
                        : m.author === 'HUMAN'
                          ? `👤 ${m.authorName || 'الموظف'}`
                          : '🤖 الوكيل'}{' '}
                      ·{' '}
                      {new Date(m.createdAt).toLocaleTimeString('ar-IQ', {
                        timeZone: 'Asia/Baghdad',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </small>
                    <p>{m.text}</p>
                  </article>
                ))}
                <AgentActionNotice
                  actions={detail.actions}
                  onChanged={() => setReload((n) => n + 1)}
                />
              </div>
              <form
                className="wa-compose"
                onSubmit={(e) => {
                  e.preventDefault()
                  void send()
                }}
              >
                <textarea
                  aria-label="رد الموظف"
                  disabled={current.status !== 'HUMAN' || busy}
                  value={text}
                  onChange={(e) => setText(e.target.value)}
                  placeholder={
                    current.status === 'HUMAN'
                      ? 'اكتب ردك للزبون…'
                      : 'استلم المحادثة حتى تگدر ترد'
                  }
                />
                <button
                  disabled={current.status !== 'HUMAN' || !text.trim() || busy}
                >
                  إرسال
                </button>
              </form>
            </>
          )}
        </section>
      </div>
      {confirm && (
        <div className="wa-overlay">
          <div
            role="dialog"
            aria-modal="true"
            aria-label="تأكيد تغيير حالة المحادثة"
            className="wa-dialog"
          >
            <h2>
              {confirm === 'release' ? 'ترجّعها للوكيل؟' : 'تغلق المحادثة؟'}
            </h2>
            <p>
              {confirm === 'release'
                ? 'الوكيل يرجع يتابع رسائل الزبون.'
                : 'المحادثة تتأشّر مغلقة.'}
            </p>
            <button disabled={busy} onClick={() => void act(confirm)}>
              تأكيد
            </button>
            <button disabled={busy} onClick={() => setConfirm(null)}>
              رجوع
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
