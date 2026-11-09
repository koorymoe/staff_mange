import { useEffect, useState } from 'react'
import { useSession } from '../session'
import { orgAdmin } from '../orgDeskAccess'
import { orgMonth } from '../orgDesk'
import {
  getOutputs,
  useOutputRegulations,
  type OutputCard,
  type OutputTarget,
} from '../orgOutputs'
import '../pages/OrgPage.css'
import './OrgOutputs.css'

export default function OrgOutputs({
  target,
  allowed = false,
  initialPeriod,
}: {
  target: OutputTarget
  allowed?: boolean
  initialPeriod?: string
}) {
  const { employee } = useSession()
  const permitted =
    !!employee &&
    !!target.id &&
    (orgAdmin(employee) ||
      (target.kind === 'employee' && target.id === employee.id) ||
      allowed)
  return permitted ? (
    <OutputPanel
      key={`${employee!.id}:${target.kind}:${target.id}`}
      target={target}
      initialPeriod={initialPeriod}
    />
  ) : (
    <p role="alert">المخرجات مو ضمن صلاحياتك.</p>
  )
}
export function MyOutputs() {
  const { employee } = useSession()
  return employee ? <OwnOutputs key={employee.id} id={employee.id} /> : null
}
function OwnOutputs({ id }: { id: string }) {
  const [open, setOpen] = useState(false)
  return (
    <section className="org-ui org-my-card" dir="rtl">
      <h2>🎯 مخرجاتي</h2>
      <p>أدائي حسب فقرات لائحة العمل، والبيانات المسجلة بالنظام.</p>
      <button aria-expanded={open} onClick={() => setOpen((v) => !v)}>
        {open ? 'إخفاء مخرجاتي' : 'افتح مخرجاتي'}
      </button>
      {open && <OrgOutputs target={{ kind: 'employee', id }} />}
    </section>
  )
}
function OutputPanel({
  target,
  initialPeriod,
}: {
  target: OutputTarget
  initialPeriod?: string
}) {
  const [period, setPeriod] = useState(() => initialPeriod || orgMonth())
  return (
    <section
      className="org-ui outputs-panel"
      dir="rtl"
      aria-label="لوحة المخرجات"
    >
      <label className="outputs-month">
        شهر المخرجات
        <input
          aria-label="شهر المخرجات"
          type="month"
          value={period}
          onChange={(e) => {
            if (/^\d{4}-(0[1-9]|1[0-2])$/.test(e.target.value))
              setPeriod(e.target.value)
          }}
        />
      </label>
      <OutputData
        key={`${target.kind}:${target.id}:${period}`}
        target={target}
        period={period}
      />
    </section>
  )
}
function OutputData({
  target,
  period,
}: {
  target: OutputTarget
  period: string
}) {
  const { kind, id } = target
  const [data, setData] = useState<OutputCard | null>(null),
    [error, setError] = useState(''),
    [reload, setReload] = useState(0)
  const { clauses, error: regError, retry } = useOutputRegulations(true)
  useEffect(() => {
    let alive = true
    getOutputs({ kind, id }, period)
      .then((card) => {
        if (alive) {
          setData(card)
          setError('')
        }
      })
      .catch((e: Error) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [kind, id, period, reload])
  if (error)
    return (
      <p role="alert" className="org-error">
        {error}{' '}
        <button
          onClick={() => {
            setData(null)
            setError('')
            setReload((n) => n + 1)
          }}
        >
          أعد جلب المخرجات
        </button>
      </p>
    )
  if (!data) return <p role="status">دا أجيب المخرجات…</p>
  const groups = new Map<string, { title: string; rows: OutputCard['rows'] }>()
  for (const row of data.rows ?? []) {
    const clause = clauses?.find(
      (c) => c.no === row.clause || c.code === row.code,
    )
    const key = clause ? String(clause.article) : 'unknown',
      title = clause
        ? `المادة ${clause.article}: ${clause.articleTitle}`
        : 'فقرات بانتظار تصنيف المادة'
    if (!groups.has(key)) groups.set(key, { title, rows: [] })
    groups.get(key)!.rows.push(row)
  }
  return (
    <>
      <header className="org-toolbar outputs-heading">
        <div>
          <h2>{data.label || 'لوحة المخرجات'}</h2>
          <p>
            الفترة: <bdi>{data.period}</bdi> · المشمولون:{' '}
            {data.people.toLocaleString('ar-IQ')}
          </p>
        </div>
        <small>الدرجات من الخادم، مو تقييم تخميني.</small>
      </header>
      {data.weakest?.length > 0 ? (
        <section className="org-warning" aria-label="أضعف الفقرات">
          <h3>⚠️ أضعف الفقرات</h3>
          <ul>
            {data.weakest.map((text, i) => (
              <li key={i}>{text}</li>
            ))}
          </ul>
        </section>
      ) : (
        <p className="outputs-neutral">
          ماكو فقرة ضعيفة ظاهرة بهالشهر، أو البيانات مو كافية.
        </p>
      )}
      {regError && (
        <p role="alert" className="org-error">
          تعذّر تصنيف المواد: {regError}{' '}
          <button onClick={retry}>أعد جلب لائحة المخرجات</button>
        </p>
      )}
      {clauses === null && !regError && <small>دا أجيب أسماء المواد…</small>}
      {[...groups.entries()]
        .sort(([a], [b]) =>
          a === 'unknown' ? 1 : b === 'unknown' ? -1 : Number(a) - Number(b),
        )
        .map(([key, group]) => (
          <section className="outputs-article" key={key}>
            <h3>{group.title}</h3>
            <div className="outputs-grid">
              {group.rows.map((row, i) => {
                const tone =
                  row.score == null
                    ? 'empty'
                    : row.score >= 80
                      ? 'good'
                      : row.score >= 50
                        ? 'medium'
                        : 'low'
                return (
                  <article
                    className={`outputs-row output-${tone}`}
                    key={`${row.code}:${i}`}
                  >
                    <header>
                      <div>
                        <bdi>{row.code}</bdi>
                        <h4>{row.title}</h4>
                      </div>
                      <strong className="outputs-score">
                        {row.score == null
                          ? 'ماكو بيانات'
                          : `${row.score.toLocaleString('ar-IQ')}٪`}
                      </strong>
                    </header>
                    {row.score != null && (
                      <progress
                        max={100}
                        value={Math.min(100, Math.max(0, row.score))}
                        aria-label={`درجة ${row.code}`}
                      />
                    )}
                    <dl className="outputs-details">
                      <div>
                        <dt>ماتركس</dt>
                        <dd>
                          {row.matrixPct == null
                            ? 'غير متوفر'
                            : `${row.matrixPct.toLocaleString('ar-IQ')}٪`}
                        </dd>
                      </div>
                      <div>
                        <dt>مهام بوقتها</dt>
                        <dd>{row.tasksOnTime.toLocaleString('ar-IQ')}</dd>
                      </div>
                      <div>
                        <dt>مهام متأخرة</dt>
                        <dd>{row.tasksLate.toLocaleString('ar-IQ')}</dd>
                      </div>
                      <div>
                        <dt>مهام فاتت</dt>
                        <dd>{row.tasksOverdue.toLocaleString('ar-IQ')}</dd>
                      </div>
                      <div>
                        <dt>خصومات KPI (عدد)</dt>
                        <dd>{row.kpiDeductions.toLocaleString('ar-IQ')}</dd>
                      </div>
                      <div>
                        <dt>شكاوي (عدد)</dt>
                        <dd>{row.complaints.toLocaleString('ar-IQ')}</dd>
                      </div>
                    </dl>
                  </article>
                )
              })}
            </div>
          </section>
        ))}
      {!data.rows?.length && (
        <p className="org-empty">ماكو فقرات راجعة بهالفترة.</p>
      )}
    </>
  )
}
