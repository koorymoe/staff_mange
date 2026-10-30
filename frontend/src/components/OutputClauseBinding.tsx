import { useRef, useState } from 'react'
import { apiRequest } from '../api'
import { useSession } from '../session'
import { orgAdmin } from '../orgDeskAccess'
import type { OrgClause } from '../orgDesk'
import '../pages/OrgPage.css'
import './OrgOutputs.css'

export default function OutputClauseBinding(props: {
  kind: 'task' | 'criterion'
  id: string
  clause?: number | null
  clauses: OrgClause[] | null
  onSaved?: (clause: number | null) => void
}) {
  const { employee } = useSession()
  return employee && orgAdmin(employee) ? (
    <Binding
      key={`${employee.id}:${props.kind}:${props.id}:${props.clause ?? ''}`}
      {...props}
    />
  ) : null
}
function Binding({
  kind,
  id,
  clause,
  clauses,
  onSaved,
}: {
  kind: 'task' | 'criterion'
  id: string
  clause?: number | null
  clauses: OrgClause[] | null
  onSaved?: (clause: number | null) => void
}) {
  const [value, setValue] = useState(clause == null ? '' : String(clause)),
    [confirmed, setConfirmed] = useState(clause == null ? '' : String(clause)),
    [busy, setBusy] = useState(false),
    [error, setError] = useState(''),
    [notice, setNotice] = useState('')
  const lock = useRef(false)
  async function save() {
    if (lock.current || value === confirmed || !clauses) return
    lock.current = true
    setBusy(true)
    setError('')
    setNotice('')
    try {
      const next = value ? Number(value) : null
      await apiRequest(
        `/org/${kind === 'task' ? 'tasks' : 'kpi-criteria'}/${encodeURIComponent(id)}/clause`,
        { method: 'PUT', body: JSON.stringify({ clause: next }) },
      )
      setConfirmed(value)
      setNotice('انحفظ ربط الفقرة.')
      onSaved?.(next)
    } catch (e) {
      setError((e as Error).message)
    } finally {
      lock.current = false
      setBusy(false)
    }
  }
  return (
    <div className="org-ui output-binding" aria-label="ربط فقرة المخرجات">
      <label>
        الفقرة
        <select
          aria-label="فقرة المخرجات"
          value={value}
          disabled={busy || clauses === null}
          onChange={(e) => setValue(e.target.value)}
        >
          <option value="">
            {clauses === null ? 'اللائحة مو متوفرة بعد' : 'بلا ربط'}
          </option>
          {clause != null && !clauses?.some((c) => c.no === clause) && (
            <option value={clause}>
              فقرة {clause} (مو فعّالة باللائحة الحالية)
            </option>
          )}
          {[
            ...new Map(
              (clauses ?? []).map((c) => [c.article, c.articleTitle]),
            ).entries(),
          ]
            .sort(([a], [b]) => a - b)
            .map(([article, title]) => (
              <optgroup key={article} label={`المادة ${article}: ${title}`}>
                {(clauses ?? [])
                  .filter((c) => c.article === article)
                  .map((c) => (
                    <option key={c.no} value={c.no}>
                      {c.code} · {c.title}
                    </option>
                  ))}
              </optgroup>
            ))}
        </select>
      </label>
      <button
        type="button"
        disabled={busy || value === confirmed || clauses === null}
        onClick={() => void save()}
      >
        {busy ? 'دا أحفظ…' : 'حفظ الفقرة'}
      </button>
      {error && (
        <p role="alert" className="org-error">
          {error}
        </p>
      )}
      {notice && <p role="status">{notice}</p>}
    </div>
  )
}
