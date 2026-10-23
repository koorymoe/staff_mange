import { useState } from 'react'
import { useSession } from '../session'
import { waApi, type WaAgent, type WaDetail } from '../waApi'
import AgentActionNotice from '../components/AgentActionNotice'
import './WhatsAppPage.css'

export default function WhatsAppSandboxPage() {
  const { employee } = useSession()
  return employee?.actualRole === 'OWNER' ? (
    <Sandbox />
  ) : (
    <p className="p-8 text-slate-500">وضع التجربة لمالك النظام فقط.</p>
  )
}
function Sandbox() {
  const [agent, setAgent] = useState<WaAgent>('QUALITY')
  const [bookingId, setBookingId] = useState('')
  const [text, setText] = useState('')
  const [detail, setDetail] = useState<WaDetail | null>(null)
  const [busy, setBusy] = useState(false)
  const [error, setError] = useState('')
  const start = async () => {
    if (busy || !bookingId.trim()) return
    setBusy(true)
    setError('')
    try {
      const r = await waApi.sandboxStart(agent, bookingId.trim())
      if (!r.conversation.sandbox)
        throw new Error(
          'الخادم رجّع محادثة مو تجريبية. ما راح نستخدمها بوضع التجربة.',
        )
      setDetail({ ...r, actions: [] })
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  const say = async () => {
    if (busy || !detail || !text.trim()) return
    setBusy(true)
    setError('')
    try {
      const r = await waApi.sandboxSay(detail.conversation.id, text.trim())
      setDetail((v) =>
        v
          ? {
              ...v,
              messages: [...v.messages, ...r.messages],
              actions: [...v.actions, ...r.actions],
            }
          : v,
      )
      setText('')
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  return (
    <div className="wa-page">
      <header className="wa-page-heading">
        <div>
          <h1>وضع التجربة</h1>
          <p>اكتب كأنك زبون، وراجع رد الوكيل وإجراءاته.</p>
        </div>
      </header>
      <div className="wa-preview-banner">
        <strong>تجربة معزولة · ماكو واتساب حقيقي</strong>
        <p>
          إجراءات التجربة ما تنفّذ على نظام الشركة. استخدم حجز وبيانات اختبار.
        </p>
      </div>
      {error && (
        <p role="alert" className="wa-error">
          {error}
        </p>
      )}
      <section className="wa-sandbox">
        <form
          className="wa-sandbox-start"
          onSubmit={(e) => {
            e.preventDefault()
            void start()
          }}
        >
          <label>
            جرّب ويا
            <select
              disabled={busy || !!detail}
              value={agent}
              onChange={(e) => setAgent(e.target.value as WaAgent)}
            >
              <option value="QUALITY">وكيل الجودة</option>
              <option value="BOOKING">وكيل الحجوزات</option>
            </select>
          </label>
          <label>
            معرّف الحجز التجريبي
            <input
              required
              disabled={busy || !!detail}
              value={bookingId}
              onChange={(e) => setBookingId(e.target.value)}
              placeholder="معرّف الحجز، مو الكود"
            />
          </label>
          {!detail ? (
            <button disabled={busy || !bookingId.trim()}>
              {busy ? 'دا أفتح…' : 'ابدأ التجربة'}
            </button>
          ) : (
            <button
              type="button"
              disabled={busy}
              onClick={() => {
                setDetail(null)
                setText('')
              }}
            >
              تجربة جديدة
            </button>
          )}
        </form>
        <div className="wa-messages" aria-live="polite">
          {!detail && (
            <div className="wa-empty">
              اختار الوكيل والحجز حتى تبدي المحادثة.
            </div>
          )}
          {detail?.messages.map((m) => (
            <article
              key={m.id}
              className={m.direction === 'IN' ? 'incoming' : 'outgoing'}
            >
              <small>
                {m.author === 'CUSTOMER' ? 'إنت بدور الزبون' : '🤖 الوكيل'}
              </small>
              <p>{m.text}</p>
            </article>
          ))}
          {busy && <p role="status">الوكيل دا يتابع…</p>}
          {detail && <AgentActionNotice actions={detail.actions} />}
        </div>
        {detail && (
          <form
            className="wa-compose"
            onSubmit={(e) => {
              e.preventDefault()
              void say()
            }}
          >
            <textarea
              aria-label="رسالة زبون تجريبية"
              disabled={busy}
              value={text}
              onChange={(e) => setText(e.target.value)}
              placeholder="اكتب كأنك زبون…"
            />
            <button disabled={busy || !text.trim()}>إرسال</button>
          </form>
        )}
      </section>
    </div>
  )
}
