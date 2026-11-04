import { useEffect, useRef, useState, type ReactNode } from 'react'
import { useSession } from '../session'
import { orgAdmin } from '../orgDeskAccess'
import { orgRegulations, type OrgClause } from '../orgDesk'
import {
  directivesApi,
  directiveDate,
  directiveTargetKey,
  type OrgDirective,
  type DirectiveScope,
  type DirectiveTarget,
} from '../orgDirectives'
import '../pages/OrgPage.css'
import './OrgDirectives.css'

export default function DirectiveManager({
  targets,
  divisionId,
}: {
  targets: DirectiveTarget[]
  divisionId?: string
}) {
  const { employee } = useSession()
  return employee ? (
    <ManagedList
      key={`${employee.id}:${divisionId || 'all'}`}
      targets={targets.filter(
        (t) => t.scope !== 'COMPANY' || orgAdmin(employee),
      )}
      divisionId={divisionId}
    />
  ) : null
}
function ManagedList({
  targets,
  divisionId,
}: {
  targets: DirectiveTarget[]
  divisionId?: string
}) {
  const [items, setItems] = useState<OrgDirective[] | null>(null),
    [error, setError] = useState(''),
    [reload, setReload] = useState(0),
    [notice, setNotice] = useState(''),
    [issuing, setIssuing] = useState(false),
    [archiving, setArchiving] = useState<OrgDirective | null>(null)
  const [clauses, setClauses] = useState<OrgClause[] | null>(null),
    [regError, setRegError] = useState(''),
    [regReload, setRegReload] = useState(0)
  useEffect(() => {
    let alive = true
    directivesApi
      .managed()
      .then((rows) => {
        if (alive) {
          setItems(rows ?? [])
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
  useEffect(() => {
    if (!targets.length) return
    let alive = true
    orgRegulations()
      .then((rows) => {
        if (alive) {
          setClauses(rows ?? [])
          setRegError('')
        }
      })
      .catch((e: Error) => {
        if (alive) setRegError(e.message)
      })
    return () => {
      alive = false
    }
  }, [regReload, targets.length])
  const refresh = () => {
    setItems(null)
    setError('')
    setReload((n) => n + 1)
  }
  const rows = (items ?? [])
    .filter(
      (d) =>
        !divisionId || d.scope === 'COMPANY' || d.divisionId === divisionId,
    )
    .sort(
      (a, b) =>
        Number(!!a.archivedAt) - Number(!!b.archivedAt) ||
        b.issuedAt.localeCompare(a.issuedAt),
    )
  return (
    <section
      className="org-ui directive-manager"
      dir="rtl"
      aria-label="إدارة التوجيهات"
    >
      <header className="org-toolbar">
        <div>
          <h2>📣 التوجيهات والمتابعة</h2>
          <p>التوجيه سياسة عمل بنطاق واضح، وإقرار الموظف يبقى مسجّل.</p>
        </div>
        {targets.length > 0 && (
          <button
            className="org-primary"
            onClick={() => {
              setIssuing(true)
              setNotice('')
            }}
          >
            + إصدار توجيه
          </button>
        )}
      </header>
      {notice && (
        <p role="status" className="org-notice">
          {notice}
        </p>
      )}
      {error && (
        <p role="alert" className="org-error">
          {error} <button onClick={refresh}>أعد جلب المتابعة</button>
        </p>
      )}
      {regError && (
        <p role="alert" className="org-error">
          لائحة العمل: {regError}{' '}
          <button onClick={() => setRegReload((n) => n + 1)}>
            أعد جلب لائحة التوجيهات
          </button>
        </p>
      )}
      {!items && !error && <p role="status">دا أجيب متابعة التوجيهات…</p>}
      {items && !error && rows.length === 0 && (
        <p className="org-empty">ماكو توجيهات ضمن متابعتك هسه.</p>
      )}
      <div className="directive-grid">
        {rows.map((d) => {
          const value =
            d.audience > 0
              ? Math.min(100, Math.max(0, (d.acks / d.audience) * 100))
              : 0
          return (
            <article
              key={d.id}
              className={`directive-card directive-followup ${d.archivedAt ? 'directive-muted' : ''}`}
            >
              <header>
                <span className="directive-place">{d.place}</span>
                {d.archivedAt && <strong>موقوف</strong>}
              </header>
              <h3>{d.title}</h3>
              <p className="directive-body">{d.body}</p>
              <p className="directive-meta">
                {d.clauseCode
                  ? `فقرة اللائحة: ${d.clauseCode}`
                  : 'بلا فقرة مرتبطة'}
              </p>
              <p className="directive-meta">
                أصدره: {d.issuedBy || 'غير متوفر'} · {directiveDate(d.issuedAt)}
              </p>
              <p>
                <bdi dir="ltr">
                  {d.acks.toLocaleString('ar-IQ')}/
                  {d.audience.toLocaleString('ar-IQ')}
                </bdi>{' '}
                أقرّوا
              </p>
              <progress
                value={value}
                max={100}
                aria-label={`إقرارات ${d.title}`}
              />
              {d.audience === 0 && (
                <small>ماكو جمهور فعّال بهالنطاق حالياً.</small>
              )}
              {d.archivedAt ? (
                <small>توقف: {directiveDate(d.archivedAt)}</small>
              ) : (
                <button className="org-danger" onClick={() => setArchiving(d)}>
                  إيقاف
                </button>
              )}
            </article>
          )
        })}
      </div>
      {issuing && targets.length > 0 && (
        <IssueDirective
          targets={targets}
          clauses={clauses}
          close={() => setIssuing(false)}
          saved={(audience) => {
            setIssuing(false)
            setNotice(`وصل لـ ${audience.toLocaleString('ar-IQ')} شخص`)
            refresh()
          }}
        />
      )}
      {archiving && (
        <DirectiveDialog
          title="إيقاف التوجيه"
          label="تأكيد الإيقاف"
          close={() => setArchiving(null)}
          saved={() => {
            setArchiving(null)
            setNotice('توقف التوجيه، وما انحذف سجله.')
            refresh()
          }}
          submit={() => directivesApi.archive(archiving.id)}
        >
          <h3>{archiving.title}</h3>
          <p>توقف التوجيه عن الجمهور. هذا إيقاف، مو حذف للسجل أو الإقرارات.</p>
        </DirectiveDialog>
      )}
    </section>
  )
}
function DirectiveDialog({
  title,
  label,
  close,
  saved,
  submit,
  children,
}: {
  title: string
  label: string
  close: () => void
  saved: () => void
  submit: (data: FormData) => Promise<unknown>
  children: ReactNode
}) {
  const ref = useRef<HTMLDialogElement>(null),
    lock = useRef(false)
  const [busy, setBusy] = useState(false),
    [error, setError] = useState('')
  useEffect(() => {
    const previous = document.activeElement as HTMLElement | null,
      dialog = ref.current,
      overflow = document.body.style.overflow
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
    const body = new FormData(form)
    lock.current = true
    setBusy(true)
    setError('')
    try {
      await submit(body)
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
      className="org-ui org-dialog directive-form"
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
          type="button"
          aria-label="إغلاق النافذة"
          disabled={busy}
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
            <button className="org-primary" type="submit">
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
function IssueDirective({
  targets,
  clauses,
  close,
  saved,
}: {
  targets: DirectiveTarget[]
  clauses: OrgClause[] | null
  close: () => void
  saved: (audience: number) => void
}) {
  const [scope, setScope] = useState<DirectiveScope>(targets[0].scope),
    [target, setTarget] = useState(directiveTargetKey(targets[0]))
  const result = useRef<{ audience: number } | null>(null)
  const options = targets.filter((t) => t.scope === scope)
  return (
    <DirectiveDialog
      title="إصدار توجيه"
      label="إصدار وإرسال التوجيه"
      close={close}
      saved={() => {
        if (result.current) saved(result.current.audience)
      }}
      submit={async (data) => {
        const choice = options.find((t) => directiveTargetKey(t) === target),
          title = String(data.get('title') || '').trim(),
          body = String(data.get('body') || '').trim(),
          clause = String(data.get('clause') || '')
        if (!choice) throw new Error('اختار نطاق مسموح للتوجيه.')
        if ([...title].length < 3 || [...body].length < 10)
          throw new Error('اكتب عنوان التوجيه ونصه (١٠ حروف على الأقل)')
        result.current = await directivesApi.issue({
          scope: choice.scope,
          ...(choice.scope !== 'COMPANY'
            ? { divisionId: choice.divisionId }
            : {}),
          ...(choice.scope === 'UNIT' ? { unitId: choice.unitId } : {}),
          title,
          body,
          clause: clause ? Number(clause) : null,
        })
      }}
    >
      <label>
        النطاق
        <select
          aria-label="نطاق التوجيه"
          value={scope}
          onChange={(e) => {
            const next = e.target.value as DirectiveScope
            setScope(next)
            setTarget(
              directiveTargetKey(targets.find((t) => t.scope === next)!),
            )
          }}
        >
          {(['COMPANY', 'DIVISION', 'UNIT'] as const)
            .filter((s) => targets.some((t) => t.scope === s))
            .map((s) => (
              <option key={s} value={s}>
                {{ COMPANY: 'الشركة', DIVISION: 'الإدارة', UNIT: 'الوحدة' }[s]}
              </option>
            ))}
        </select>
      </label>
      {scope !== 'COMPANY' && (
        <label>
          المكان
          <select
            aria-label="مكان التوجيه"
            value={target}
            onChange={(e) => setTarget(e.target.value)}
          >
            {options.map((t) => (
              <option key={directiveTargetKey(t)} value={directiveTargetKey(t)}>
                {t.label}
              </option>
            ))}
          </select>
        </label>
      )}
      <label>
        عنوان التوجيه
        <input name="title" required minLength={3} />
      </label>
      <label>
        نص التوجيه
        <textarea name="body" required minLength={10} rows={5} />
      </label>
      <label>
        فقرة اللائحة (اختياري)
        <select
          name="clause"
          aria-label="فقرة التوجيه"
          disabled={clauses === null}
        >
          <option value="">
            {clauses === null ? 'اللائحة مو متوفرة بعد' : 'بلا فقرة مرتبطة'}
          </option>
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
      <p className="directive-meta">
        الإرسال يصدر التوجيه على النطاق المحدد ويطلب إقرار العاملين بيه.
      </p>
    </DirectiveDialog>
  )
}
