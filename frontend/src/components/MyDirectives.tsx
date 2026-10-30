import { useEffect, useRef, useState } from 'react'
import {
  directivesApi,
  directiveDate,
  type OrgDirective,
} from '../orgDirectives'
import { useSession } from '../session'
import '../pages/OrgPage.css'
import './OrgDirectives.css'

export default function MyDirectives() {
  const { employee } = useSession()
  return employee ? <MyDirectiveList key={employee.id} /> : null
}
function MyDirectiveList() {
  const [items, setItems] = useState<OrgDirective[] | null>(null),
    [error, setError] = useState(''),
    [notice, setNotice] = useState(''),
    [reload, setReload] = useState(0),
    [busy, setBusy] = useState<string | null>(null)
  const lock = useRef(false)
  useEffect(() => {
    let alive = true
    directivesApi
      .mine()
      .then((rows) => {
        if (alive) {
          setItems(
            (rows ?? [])
              .filter((d) => !d.archivedAt)
              .sort(
                (a, b) =>
                  Number(!!a.ackAt) - Number(!!b.ackAt) ||
                  b.issuedAt.localeCompare(a.issuedAt),
              ),
          )
          setError('')
        }
      })
      .catch((e: Error) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [reload])
  const refresh = () => {
    setItems(null)
    setError('')
    setReload((n) => n + 1)
  }
  async function ack(item: OrgDirective) {
    if (lock.current || item.ackAt || item.archivedAt) return
    lock.current = true
    setBusy(item.id)
    setError('')
    setNotice('')
    try {
      await directivesApi.ack(item.id)
      setNotice('انحفظ إقرارك.')
      refresh()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      lock.current = false
      setBusy(null)
    }
  }
  return (
    <section
      className="org-ui org-my-card directive-mine"
      dir="rtl"
      aria-label="توجيهات بانتظار إقرارك"
    >
      <header className="org-toolbar">
        <h2>📣 توجيهات بانتظار إقرارك</h2>
        {items && (
          <span className="directive-count">
            {items.filter((d) => !d.ackAt).length.toLocaleString('ar-IQ')}{' '}
            بانتظارك
          </span>
        )}
      </header>
      <p>اقرأ التوجيه كامل، وبعدين أقرّ إنك اطّلعت وتلتزم بيه.</p>
      {notice && (
        <p role="status" className="org-notice">
          {notice}
        </p>
      )}
      {error && (
        <p role="alert" className="org-error">
          {error}{' '}
          <button disabled={!!busy} onClick={refresh}>
            أعد جلب توجيهاتي
          </button>
        </p>
      )}
      {!items && !error && <p role="status">دا أجيب توجيهاتك…</p>}
      {items?.length === 0 && (
        <p className="org-empty">ماكو توجيهات بانتظارك هسه.</p>
      )}
      <div className="directive-grid">
        {items?.map((item) => (
          <article
            key={item.id}
            className={`directive-card ${item.ackAt ? 'directive-muted' : ''}`}
          >
            <header>
              <span className="directive-place">{item.place}</span>
              {item.ackAt && (
                <strong className="directive-confirmed">✓ أقرّيت</strong>
              )}
            </header>
            <h3>{item.title}</h3>
            <p className="directive-body">{item.body}</p>
            <p className="directive-meta">
              {item.clauseCode
                ? `فقرة اللائحة: ${item.clauseCode}`
                : 'بلا فقرة مرتبطة'}
            </p>
            <p className="directive-meta">
              أصدره: {item.issuedBy || 'غير متوفر'} ·{' '}
              {directiveDate(item.issuedAt)}
            </p>
            {item.ackAt ? (
              <small>الإقرار: {directiveDate(item.ackAt)}</small>
            ) : (
              <button
                className="org-primary"
                disabled={!!busy}
                onClick={() => void ack(item)}
              >
                {busy === item.id ? 'دا أسجّل إقرارك…' : 'اطّلعت وألتزم'}
              </button>
            )}
          </article>
        ))}
      </div>
      {items && items.length >= 100 && (
        <small>المعروض آخر ١٠٠ توجيه كحد أقصى.</small>
      )}
    </section>
  )
}
