import { useEffect, useRef, useState, type ReactNode } from 'react'
import { orgPositions, type OrgDivision } from '../../org'
import { orgDeskWrite, orgIQD, type OrgClause, type AuthorityRow, type DeskIssue } from '../../orgDesk'
import { statuses, field } from './shared'

export function Team({ division }: { division: OrgDivision }) {
  const group = (name: string, people = division.people ?? []) => (
    <section className="org-panel" key={name}>
      <h3>{name}</h3>
      {(['DIRECTOR', 'DEPUTY', 'MEMBER'] as const).map((position) => (
        <div className="desk-team-position" key={position}>
          <h4>{orgPositions[position]}</h4>
          {people.some((p) => p.position === position) ? (
            <ul>
              {people
                .filter((p) => p.position === position)
                .map((p) => (
                  <li key={p.id}>
                    {p.employeeName}
                    {p.acting && ' · بالوكالة'}
                  </li>
                ))}
            </ul>
          ) : (
            <p className="desk-meta">
              {position === 'MEMBER'
                ? 'بلا أعضاء'
                : `بلا ${orgPositions[position]}`}
            </p>
          )}
        </div>
      ))}
    </section>
  )
  return (
    <div className="desk-cards">
      {group(
        division.name,
        (division.people ?? []).filter((p) => !p.unitId),
      )}
      {(division.units ?? []).map((u) => group(u.name, u.people ?? []))}
    </div>
  )
}

export function DeskForm({
  title,
  children,
  close,
  saved,
  submit,
  label = 'حفظ',
}: {
  title: string
  children: ReactNode
  close: () => void
  saved: () => void
  submit: (data: FormData) => Promise<unknown>
  label?: string
}) {
  const ref = useRef<HTMLDialogElement>(null),
    lock = useRef(false)
  const [busy, setBusy] = useState(false),
    [error, setError] = useState('')
  useEffect(() => {
    const previous = document.activeElement as HTMLElement | null
    const dialog = ref.current
    const overflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    dialog?.showModal()
    return () => {
      dialog?.close()
      document.body.style.overflow = overflow
      if (previous?.isConnected) previous.focus()
    }
  }, [])
  async function save(form: HTMLFormElement) {
    if (lock.current) return
    lock.current = true
    setError('')
    const data = new FormData(form)
    setBusy(true)
    try {
      await submit(data)
      saved()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      lock.current = false
      setBusy(false)
    }
  }
  return (
    <dialog
      ref={ref}
      className="org-ui org-dialog desk-dialog"
      dir="rtl"
      aria-label={title}
      onCancel={(e) => {
        e.preventDefault()
        if (!busy) close()
      }}
    >
      <header className="org-toolbar">
        <h2>{title}</h2>
        <button
          disabled={busy}
          type="button"
          aria-label="إغلاق النافذة"
          onClick={close}
        >
          ✕
        </button>
      </header>
      <form
        onSubmit={(e) => {
          e.preventDefault()
          void save(e.currentTarget)
        }}
      >
        <fieldset disabled={busy}>
          {children}
          {error && (
            <p role="alert" className="org-error">
              {error}
            </p>
          )}
          <div className="org-actions">
            <button type="submit" className="org-primary">
              {busy ? 'دا أحفظ…' : label}
            </button>
            <button type="button" onClick={close}>
              إلغاء
            </button>
          </div>
        </fieldset>
      </form>
    </dialog>
  )
}

export function IssueEditor({
  division,
  editableUnits,
  regulations,
  close,
  saved,
}: {
  division: OrgDivision
  editableUnits: string[]
  regulations: OrgClause[] | null
  close: () => void
  saved: () => void
}) {
  return (
    <DeskForm
      title="مشكلة جديدة"
      close={close}
      saved={saved}
      submit={(d) =>
        orgDeskWrite('issues', {
          divisionId: division.id,
          ...(field(d, 'unitId') ? { unitId: field(d, 'unitId') } : {}),
          title: field(d, 'title'),
          problem: field(d, 'problem'),
          impact: field(d, 'impact'),
          scope: field(d, 'scope'),
          principle: field(d, 'principle')
            ? Number(field(d, 'principle'))
            : null,
          severity: field(d, 'severity'),
          solutionPlan: field(d, 'solutionPlan'),
        })
      }
    >
      <label>
        عنوان المشكلة
        <input name="title" minLength={3} required />
      </label>
      <label>
        شرح المشكلة
        <textarea name="problem" minLength={10} required rows={3} />
      </label>
      <label>
        مضارها
        <textarea name="impact" required rows={2} />
      </label>
      <label>
        خطة الحل (اختياري)
        <textarea name="solutionPlan" rows={3} />
      </label>
      <label>
        الوحدة
        <select name="unitId" aria-label="الوحدة">
          <option value="">على مستوى الإدارة</option>
          {(division.units ?? [])
            .filter((u) => editableUnits.includes(u.id))
            .map((u) => (
              <option key={u.id} value={u.id}>
                {u.name}
              </option>
            ))}
        </select>
      </label>
      <label>
        النطاق
        <select name="scope" aria-label="النطاق">
          <option value="SPECIFIC">خاصة</option>
          <option value="GENERAL">عامة</option>
        </select>
      </label>
      <label>
        فقرة لائحة العمل
        <select
          name="principle"
          aria-label="فقرة لائحة العمل"
          disabled={regulations === null}
        >
          <option value="">
            {regulations === null ? 'اللائحة مو متوفرة بعد' : 'مو محددة'}
          </option>
          {[
            ...new Map(
              (regulations ?? []).map((c) => [c.article, c.articleTitle]),
            ).entries(),
          ]
            .sort(([a], [b]) => a - b)
            .map(([article, title]) => (
              <optgroup key={article} label={`المادة ${article}: ${title}`}>
                {(regulations ?? [])
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
      <label>
        الخطورة
        <select name="severity" aria-label="الخطورة" defaultValue="MEDIUM">
          <option value="LOW">واطية</option>
          <option value="MEDIUM">متوسطة</option>
          <option value="HIGH">عالية</option>
        </select>
      </label>
    </DeskForm>
  )
}

export function StatusEditor({
  issue,
  close,
  saved,
}: {
  issue: DeskIssue
  close: () => void
  saved: () => void
}) {
  const [status, setStatus] = useState(issue.status)
  return (
    <DeskForm
      title="تغيير حالة المشكلة"
      close={close}
      saved={saved}
      submit={async (d) => {
        if (status === 'RESOLVED' && !field(d, 'resolution'))
          throw new Error('اشرح شلون انحلت المشكلة.')
        return orgDeskWrite(
          `issues/${encodeURIComponent(issue.id)}`,
          { status, resolution: field(d, 'resolution') },
          'PATCH',
        )
      }}
    >
      <h3>{issue.title}</h3>
      <label>
        الحالة
        <select
          aria-label="الحالة"
          value={status}
          onChange={(e) => setStatus(e.target.value as DeskIssue['status'])}
        >
          {Object.entries(statuses).map(([key, label]) => (
            <option value={key} key={key}>
              {label}
            </option>
          ))}
        </select>
      </label>
      <label>
        شرح الحل{status !== 'RESOLVED' && ' (اختياري)'}
        <textarea
          name="resolution"
          required={status === 'RESOLVED'}
          defaultValue={issue.resolution ?? ''}
          rows={4}
        />
      </label>
    </DeskForm>
  )
}

export function AuthorityEditor({
  id,
  rows,
  close,
  saved,
}: {
  id: string
  rows: AuthorityRow[]
  close: () => void
  saved: () => void
}) {
  const [draft, setDraft] = useState(() =>
    rows.map((r, i) => ({
      ...r,
      key: `old-${i}`,
      amount:
        r.limitIQD == null
          ? ''
          : new Intl.NumberFormat('en-US').format(r.limitIQD),
    })),
  )
  return (
    <DeskForm
      title="تعديل جدول الصلاحيات"
      close={close}
      saved={saved}
      submit={async () => {
        const output = draft.map((r) => {
          const n = orgIQD(r.amount)
          if (!r.decision.trim())
            throw new Error('اكتب القرار أو احذف الصف الفارغ.')
          if (r.amount.trim() && (!n.valid || n.small))
            throw new Error(
              'اكتب السقف كامل بالدينار، بلا كسور أو اختصار ألوف.',
            )
          return {
            decision: r.decision.trim(),
            mode: r.mode,
            limitIQD: r.amount.trim() ? n.value : null,
          }
        })
        return orgDeskWrite(
          `authority/${encodeURIComponent(id)}`,
          output,
          'PUT',
        )
      }}
    >
      <p className="org-warning">
        الحفظ يبدّل الجدول كله. الصف الي تشيله هنا ينشال من الجدول عند الحفظ.
      </p>
      {draft.map((r, i) => (
        <section className="desk-authority-row" key={r.key}>
          <label>
            القرار {i + 1}
            <input
              required
              value={r.decision}
              onChange={(e) =>
                setDraft((d) =>
                  d.map((row) =>
                    row.key === r.key
                      ? { ...row, decision: e.target.value }
                      : row,
                  ),
                )
              }
            />
          </label>
          <label>
            طريقة الحسم
            <select
              aria-label={`طريقة الحسم ${i + 1}`}
              value={r.mode}
              onChange={(e) =>
                setDraft((d) =>
                  d.map((row) =>
                    row.key === r.key
                      ? { ...row, mode: e.target.value as AuthorityRow['mode'] }
                      : row,
                  ),
                )
              }
            >
              <option value="DECIDE">يقرر لحاله</option>
              <option value="ESCALATE">يرفعه</option>
            </select>
          </label>
          <label>
            السقف بالدينار {i + 1}
            <input
              inputMode="numeric"
              dir="ltr"
              value={r.amount}
              onChange={(e) => {
                const n = orgIQD(e.target.value)
                setDraft((d) =>
                  d.map((row) =>
                    row.key === r.key
                      ? {
                          ...row,
                          amount: n.valid
                            ? new Intl.NumberFormat('en-US').format(n.value)
                            : e.target.value,
                        }
                      : row,
                  ),
                )
              }}
              placeholder="فاضي = ما محدد"
            />
          </label>
          {r.amount.trim() &&
            (!orgIQD(r.amount).valid || orgIQD(r.amount).small) && (
              <p role="alert" className="org-warning">
                اكتب المبلغ كامل مثل 50,000؛ الصفر مسموح.
              </p>
            )}
          <button
            type="button"
            className="org-danger"
            onClick={() =>
              setDraft((d) => d.filter((row) => row.key !== r.key))
            }
          >
            شيل الصف {i + 1}
          </button>
        </section>
      ))}
      <button
        type="button"
        onClick={() =>
          setDraft((d) => [
            ...d,
            {
              key: crypto.randomUUID(),
              decision: '',
              mode: 'DECIDE',
              limitIQD: null,
              amount: '',
            },
          ])
        }
      >
        + إضافة صف
      </button>
      <label className="desk-check">
        <input type="checkbox" required />
        راجعت الجدول وأوافق على استبداله بالكامل
      </label>
    </DeskForm>
  )
}
