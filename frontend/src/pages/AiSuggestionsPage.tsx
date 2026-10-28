import { useEffect, useRef, useState } from 'react'
import { apiRequest } from '../api'
import { useSession } from '../session'
import '../components/AiBrains.css'

type Decision = 'ACCEPTED' | 'REJECTED' | 'DONE'
type Suggestion = {
  id: string
  agent: string
  title: string
  details: string
  needsCode: boolean
  status: 'NEW' | Decision
  createdAt: string
}
const statuses = {
  NEW: 'جديد',
  ACCEPTED: 'مقبول',
  REJECTED: 'مرفوض',
  DONE: 'منجز',
}
const names: Record<string, string> = {
  MATRIX: 'ماتركس',
  IRIS: 'ايريس',
  QUALITY: 'وكيل الجودة',
  BOOKING: 'وكيل الحجوزات',
}
export default function AiSuggestionsPage() {
  const { employee } = useSession()
  return employee?.actualRole === 'OWNER' || employee?.role === 'ADMIN' ? (
    <Suggestions />
  ) : (
    <p>اقتراحات الذكاء للمالك ومدير النظام فقط.</p>
  )
}
function Suggestions() {
  const [filter, setFilter] = useState<'all' | 'code' | 'new'>('all')
  const [items, setItems] = useState<Suggestion[] | null>(null)
  const [reload, setReload] = useState(0)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [pending, setPending] = useState<{
    item: Suggestion
    status: Decision
  } | null>(null)
  const [busy, setBusy] = useState(false)
  const dialog = useRef<HTMLElement>(null)
  useEffect(() => {
    if (!pending) return
    const previous = document.activeElement as HTMLElement | null
    const overflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    dialog.current?.querySelector<HTMLButtonElement>('button')?.focus()
    return () => {
      document.body.style.overflow = overflow
      if (previous?.isConnected) previous.focus()
    }
  }, [pending])
  useEffect(() => {
    let alive = true
    const query = new URLSearchParams({ status: filter === 'new' ? 'NEW' : '' })
    if (filter === 'code') query.set('needsCode', '1')
    apiRequest<Suggestion[]>(`/ai/suggestions?${query}`)
      .then((rows) => {
        if (alive) {
          setItems(rows)
          setError('')
        }
      })
      .catch((e) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [filter, reload])
  const decide = async () => {
    if (!pending || busy) return
    setBusy(true)
    setError('')
    setSuccess('')
    try {
      await apiRequest(
        `/ai/suggestions/${encodeURIComponent(pending.item.id)}/decide`,
        { method: 'POST', body: JSON.stringify({ status: pending.status }) },
      )
      setPending(null)
      setItems(null)
      setSuccess('انحفظ قرارك. هذا تحديث حالة فقط، مو تنفيذ تلقائي.')
      setReload((n) => n + 1)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      setBusy(false)
    }
  }
  return (
    <div className="ai-suggestions" dir="rtl">
      <header>
        <h1>🧠 اقتراحات الذكاء</h1>
        <p>
          أفكار الوكلاء بمكان واحد. الموافقة تسجّل قرارك؛ ما تشغّل كود ولا تنفّذ
          التعديل وحدها.
        </p>
      </header>
      <div className="ai-suggestion-filters" aria-label="فلتر الاقتراحات">
        {(
          [
            ['all', 'الكل'],
            ['code', 'يحتاج برمجة'],
            ['new', 'الجديد'],
          ] as const
        ).map(([key, label]) => (
          <button
            key={key}
            aria-pressed={filter === key}
            disabled={busy}
            onClick={() => {
              setFilter(key)
              setItems(null)
              setError('')
              setSuccess('')
              if (key === filter) setReload((n) => n + 1)
            }}
          >
            {label}
          </button>
        ))}
      </div>
      {error && (
        <p role="alert" className="ai-brain-notice">
          {error}{' '}
          {!pending && (
            <button
              onClick={() => {
                setItems(null)
                setReload((n) => n + 1)
              }}
            >
              أعد المحاولة
            </button>
          )}
        </p>
      )}
      {success && <p role="status">{success}</p>}
      {!items && !error && <p role="status">دا أجيب اقتراحات الوكلاء…</p>}
      {items?.length === 0 && (
        <p className="ai-brain-notice">ماكو اقتراحات بهالفلتر لحد هسه.</p>
      )}
      {items?.map((item) => (
        <article className="ai-suggestion-card" key={item.id}>
          <div className="ai-suggestion-meta">
            <span>{names[item.agent.toUpperCase()] ?? item.agent}</span>
            <span className="ai-suggestion-badge">
              {statuses[item.status] ?? item.status}
            </span>
            {item.needsCode && (
              <span className="ai-suggestion-badge">⌘ يحتاج برمجة</span>
            )}
            <small>
              {new Intl.DateTimeFormat('ar-IQ', {
                timeZone: 'Asia/Baghdad',
                dateStyle: 'medium',
                timeStyle: 'short',
              }).format(new Date(item.createdAt))}
            </small>
          </div>
          <h2>{item.title}</h2>
          <p>{item.details}</p>
          <div className="ai-suggestion-actions">
            {(
              [
                ['ACCEPTED', 'وافق', 'accept'],
                ['REJECTED', 'ارفض', 'reject'],
                ['DONE', 'تم', ''],
              ] as const
            ).map(([status, label, cls]) => (
              <button
                key={status}
                className={cls}
                disabled={busy || item.status === status}
                onClick={() => {
                  setPending({ item, status })
                  setError('')
                }}
              >
                {label}
              </button>
            ))}
          </div>
        </article>
      ))}
      {pending && (
        <div
          className="ai-suggestion-dialog-overlay"
          onKeyDown={(e) => {
            if (e.key === 'Escape' && !busy) setPending(null)
            if (e.key === 'Tab') {
              const buttons = [
                ...(dialog.current?.querySelectorAll<HTMLButtonElement>(
                  'button:not(:disabled)',
                ) ?? []),
              ]
              const first = buttons[0],
                last = buttons.at(-1)
              if (!first) e.preventDefault()
              else if (e.shiftKey && document.activeElement === first) {
                e.preventDefault()
                last?.focus()
              } else if (!e.shiftKey && document.activeElement === last) {
                e.preventDefault()
                first.focus()
              }
            }
          }}
        >
          <section
            ref={dialog}
            role="dialog"
            aria-modal="true"
            aria-label="تأكيد قرار الاقتراح"
            className="ai-suggestion-dialog"
          >
            <h2>تأكيد: {statuses[pending.status]}</h2>
            <p>{pending.item.title}</p>
            <p>هذا يغيّر حالة الاقتراح فقط. ما ينفّذ برمجة أو إجراء بالنظام.</p>
            {error && <p role="alert">{error}</p>}
            <div className="ai-suggestion-actions">
              <button autoFocus disabled={busy} onClick={() => void decide()}>
                {busy ? 'دا أحفظ…' : 'تأكيد القرار'}
              </button>
              <button disabled={busy} onClick={() => setPending(null)}>
                إلغاء
              </button>
            </div>
          </section>
        </div>
      )}
    </div>
  )
}
