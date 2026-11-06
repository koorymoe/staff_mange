import { useEffect, useRef, useState } from 'react'
import { apiRequest } from '../api'
import { useSession, isAdminOrOwner } from '../session'
import '../components/AiBrains.css'

type Decision = 'ACCEPTED' | 'REJECTED' | 'DONE'
type Suggestion = {
  id: string
  agent: string
  title: string
  details: string
  needsCode: boolean
  status: 'NEW' | 'VERIFYING' | Decision
  createdAt: string
  verifyResult?: 'FOUND' | 'NOT_FOUND' | null
  verifyNote?: string | null
}
const statuses = {
  NEW: 'جديد',
  ACCEPTED: 'مقبول',
  REJECTED: 'مرفوض',
  DONE: 'منجز',
  VERIFYING: '⏳ ينتظر تأكد الوكيل',
}
const names: Record<string, string> = {
  MATRIX: 'ماتركس',
  IRIS: 'ايريس',
  QUALITY: 'وكيل الجودة',
  BOOKING: 'وكيل الحجوزات',
}
export default function AiSuggestionsPage() {
  const { employee } = useSession()
  return isAdminOrOwner(employee) ? (
    <Suggestions />
  ) : (
    <p>اقتراحات الذكاء للمالك ومدير النظام فقط.</p>
  )
}
function Suggestions() {
  const [filter, setFilter] = useState<
    'all' | 'code' | 'new' | 'verify' | 'lessons' | 'health'
  >('all')
  const extra = filter === 'lessons' || filter === 'health'
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
    if (filter === 'lessons' || filter === 'health') return
    let alive = true
    const query = new URLSearchParams({
      status: filter === 'new' ? 'NEW' : filter === 'verify' ? 'VERIFYING' : '',
    })
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
            ['verify', 'ينتظر التأكد'],
            ['lessons', '📘 دروس مقترحة'],
            ['health', '🩺 صحة الوكلاء'],
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
      {filter === 'lessons' && <ProposedLessons />}
      {filter === 'health' && <AgentHealth />}
      {!extra && !items && !error && <p role="status">دا أجيب اقتراحات الوكلاء…</p>}
      {!extra && items?.length === 0 && (
        <p className="ai-brain-notice">ماكو اقتراحات بهالفلتر لحد هسه.</p>
      )}
      {!extra && items?.map((item) => (
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
          {item.verifyResult && (
            <p className="ai-brain-notice" role="note">
              {item.verifyResult === 'FOUND'
                ? '✅ الوكيل تأكد إنه منضاف'
                : '⚠️ الوكيل ما لگاه منضاف — رجع لحالة «مقبول»'}
              {item.verifyNote ? `: ${item.verifyNote}` : ''}
            </p>
          )}
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
                disabled={
                  busy ||
                  item.status === status ||
                  (status === 'DONE' && item.status === 'VERIFYING')
                }
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
            <p>
              {pending.status === 'DONE'
                ? 'الاقتراح يصير «ينتظر تأكد الوكيل»: الوكيل صاحبه يتأكد بأول فرصة إنه صدگ انضاف، ويكتب النتيجة هنا.'
                : 'هذا يغيّر حالة الاقتراح فقط. ما ينفّذ برمجة أو إجراء بالنظام.'}
            </p>
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

// ═══ 📘 دروس اقترحها الوكلاء على نفسهم (دازاي 11-06) — ما تدخل تعليماتهم إلا بموافقة المالك ═══
type Lesson = {
  id: string
  agent: string
  text: string
  context: string | null
  createdAt: string
}
function ProposedLessons() {
  const [rows, setRows] = useState<Lesson[] | null>(null)
  const [err, setErr] = useState('')
  const [reload, setReload] = useState(0)
  useEffect(() => {
    apiRequest<Lesson[]>('/matrix/lessons?status=PROPOSED')
      .then(setRows)
      .catch((e) => setErr(e.message))
  }, [reload])
  const act = async (id: string, approve: boolean) => {
    try {
      await apiRequest(
        `/matrix/lessons/${encodeURIComponent(id)}${approve ? '/approve' : ''}`,
        { method: approve ? 'POST' : 'DELETE' },
      )
      setReload((n) => n + 1)
    } catch (e) {
      setErr((e as Error).message)
    }
  }
  if (err) return <p role="alert" className="ai-brain-notice">{err}</p>
  if (!rows) return <p role="status">دا أجيب الدروس…</p>
  if (rows.length === 0)
    return <p className="ai-brain-notice">ماكو دروس تنتظر موافقتك.</p>
  return (
    <>
      <p>
        الوكيل لاحظ غلط بنفسه واقترح هالدرس. «اعتمد» يدخله بتعليماته من الرسالة
        الجاية، و«ارفض» يشيله.
      </p>
      {rows.map((l) => (
        <article className="ai-suggestion-card" key={l.id}>
          <div className="ai-suggestion-meta">
            <span>{names[l.agent.toUpperCase()] ?? l.agent}</span>
          </div>
          <h2>{l.text}</h2>
          {l.context && <p>ليش: {l.context}</p>}
          <div className="ai-suggestion-actions">
            <button className="accept" onClick={() => void act(l.id, true)}>
              اعتمد
            </button>
            <button className="reject" onClick={() => void act(l.id, false)}>
              ارفض
            </button>
          </div>
        </article>
      ))}
    </>
  )
}

// ═══ 🩺 صحة الوكلاء: الأدوات الفاشلة، الأسئلة الي ما انجاوبت، وتقييم المستخدمين ═══
type Health = Record<
  string,
  {
    tools: {
      tool: string
      calls: number
      failures: number
      avgMs: number
      lastError: string | null
    }[]
    gaps: { kind: string; count: number; missing: string }[]
    good: number
    bad: number
    badNotes: string[]
    proposedLessons: number
    /** امتحان الوكيل (11-06): آخر الجولات، والأخيرة بتفاصيلها */
    evals?: {
      build: string
      total: number
      passed: number
      at: string
      details?: { id: string; pass: boolean; why?: string; steps: string[] }[]
    }[]
  }
>
function AgentHealth() {
  const [h, setH] = useState<Health | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => {
    apiRequest<Health>('/ai/health')
      .then(setH)
      .catch((e) => setErr(e.message))
  }, [])
  const [evalMsg, setEvalMsg] = useState('')
  const runEval = async () => {
    setEvalMsg('')
    try {
      const r = await apiRequest<{ status: string }>('/ai/evals/run', { method: 'POST' })
      setEvalMsg(r.status)
    } catch (e) {
      setEvalMsg(e instanceof Error ? e.message : 'تعذر')
    }
  }
  if (err) return <p role="alert" className="ai-brain-notice">{err}</p>
  if (!h) return <p role="status">دا أجيب صحة الوكلاء…</p>
  return (
    <>
      <p>آخر ٧ أيام (والأسئلة الي ما انجاوبت آخر ١٤ يوم).</p>
      <p>
        <button type="button" className="ai-suggestion-badge" onClick={() => void runEval()}>
          🧪 امتحن ماتركس هسه
        </button>{' '}
        {evalMsg && <span role="status">{evalMsg}</span>}
      </p>
      {Object.entries(h).map(([agent, a]) => {
        const failing = a.tools.filter((t) => t.failures > 0)
        const total = a.tools.reduce((n, t) => n + t.calls, 0)
        return (
          <article className="ai-suggestion-card" key={agent}>
            <div className="ai-suggestion-meta">
              <span>{names[agent] ?? agent}</span>
              <span className="ai-suggestion-badge">
                👍 {a.good} · 👎 {a.bad}
              </span>
              <span className="ai-suggestion-badge">{total} نداء أداة</span>
              {a.proposedLessons > 0 && (
                <span className="ai-suggestion-badge">
                  📘 {a.proposedLessons} درس ينتظرك
                </span>
              )}
            </div>
            {a.evals && a.evals.length > 0 && (
              <>
                <h2>
                  🧪 الامتحان: {a.evals[0].passed} من {a.evals[0].total} (
                  {Math.round((a.evals[0].passed * 100) / a.evals[0].total)}٪)
                </h2>
                <p>
                  {a.evals
                    .map((e) => `${Math.round((e.passed * 100) / e.total)}٪ ${new Date(e.at).toLocaleDateString('ar-IQ')}`)
                    .join(' ← ')}
                </p>
                {(a.evals[0].details ?? [])
                  .filter((d) => !d.pass)
                  .map((d) => (
                    <p key={d.id}>
                      ❌ {d.id}: {d.why}
                      {d.steps.length > 0 ? ` (استعمل: ${d.steps.join('، ')})` : ''}
                    </p>
                  ))}
              </>
            )}
            <h2>أدوات تفشل</h2>
            {failing.length === 0 ? (
              <p>ماكو فشل ✅</p>
            ) : (
              failing.slice(0, 8).map((t) => (
                <p key={t.tool}>
                  {t.tool}: {t.failures} من {t.calls}
                  {t.lastError ? ` — ${t.lastError}` : ''}
                </p>
              ))
            )}
            <h2>أسئلة ما گدر يجاوبها</h2>
            {a.gaps.length === 0 ? (
              <p>ماكو ✅</p>
            ) : (
              a.gaps.slice(0, 8).map((g) => (
                <p key={g.kind}>
                  {g.kind} ×{g.count}: {g.missing}
                </p>
              ))
            )}
            {a.badNotes.length > 0 && (
              <>
                <h2>شنو ما عجب المستخدمين</h2>
                {a.badNotes.map((n, i) => (
                  <p key={i}>👎 {n}</p>
                ))}
              </>
            )}
          </article>
        )
      })}
    </>
  )
}
