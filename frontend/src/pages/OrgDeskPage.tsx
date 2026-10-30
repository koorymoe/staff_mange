import { useEffect, useRef, useState, type ReactNode } from 'react'
import { Link, Navigate, useNavigate, useParams } from 'react-router-dom'
import {
  orgApi,
  orgPositions,
  type OrgDivision,
  type OrgPerson,
  type OrgUnit,
} from '../org'
import { useSession } from '../session'
import {
  orgAdmin,
  orgDeskAllowed,
  orgDeskSeeds,
  orgHeads,
  orgSeesAll,
  useOrgMemberships,
} from '../orgDeskAccess'
import {
  orgDeskGet,
  orgDeskWrite,
  orgIQD,
  orgMonth,
  orgRegulations,
  type OrgClause,
  type DeskPlan,
  type AuthorityRow,
  type DeskComment,
  type DeskIssue,
  type DeskItem,
  type DeskMetric,
  type OrgDesk,
} from '../orgDesk'
import './OrgPage.css'
import './OrgDeskPage.css'

type Tab = 'issues' | 'plan' | 'units' | 'authority' | 'team'
const canAddItems = (plan: DeskPlan) =>
  plan.status === 'DRAFT' || plan.status === 'RETURNED'
const dueOrder = (a: DeskItem, b: DeskItem) =>
  !a.dueDate
    ? b.dueDate
      ? 1
      : 0
    : !b.dueDate
      ? -1
      : a.dueDate.localeCompare(b.dueDate)
const teamPeople = (division: OrgDivision) => [
  ...new Map(
    [
      ...(division.people ?? []),
      ...(division.units ?? []).flatMap((u) => u.people ?? []),
    ].map((p) => [p.employeeId, p]),
  ).values(),
]
const number = (n: number | null | undefined) =>
  n == null
    ? 'غير متوفر'
    : new Intl.NumberFormat('ar-IQ', { maximumFractionDigits: 2 }).format(n)
const percent = (n: number) =>
  Math.min(100, Math.max(0, Number.isFinite(n) ? n : 0))
const statuses = {
  OPEN: 'مفتوحة',
  IN_PROGRESS: 'قيد المعالجة',
  RESOLVED: 'انحلت',
}
const date = (value: string) => {
  const d = new Date(value)
  return Number.isNaN(d.getTime())
    ? 'التاريخ مو متوفر'
    : new Intl.DateTimeFormat('ar-IQ', {
        timeZone: 'Asia/Baghdad',
        dateStyle: 'medium',
        timeStyle: 'short',
      }).format(d)
}

export default function OrgDeskPage() {
  const { employee } = useSession()
  const { divisionId } = useParams()
  const membership = useOrgMemberships(employee)
  if (!employee) return <p>سجّل دخولك حتى تفتح المكتب.</p>
  if (!orgAdmin(employee) && membership.error)
    return (
      <p role="alert">
        {membership.error}{' '}
        <button onClick={membership.retry}>أعد المحاولة</button>
      </p>
    )
  if (!orgAdmin(employee) && membership.items === null)
    return <p role="status">دا أتحقق من منصبك…</p>
  if (!orgDeskAllowed(employee, membership.items, divisionId))
    return <p role="alert">ما عندك منصب إداري يفتح هالمكتب.</p>
  const heads = orgHeads(membership.items)
  const choices = orgSeesAll(heads)
    ? orgDeskSeeds
    : [
        ...new Map(
          heads.map((p) => [
            p.divisionId,
            { id: p.divisionId, name: p.divisionName || p.divisionId },
          ]),
        ).values(),
      ]
  return (
    <DeskDirectory
      key={employee.id}
      admin={orgAdmin(employee)}
      choices={choices}
      divisionId={divisionId}
      managementHead={orgSeesAll(heads)}
    />
  )
}
function DeskDirectory({
  admin,
  choices,
  divisionId,
  managementHead,
}: {
  admin: boolean
  choices: { id: string; name: string }[]
  divisionId?: string
  managementHead: boolean
}) {
  const [chart, setChart] = useState<OrgDivision[] | null>(null)
  const [error, setError] = useState('')
  const [reload, setReload] = useState(0)
  const [period, setPeriod] = useState(orgMonth)
  const [tab, setTab] = useState<Tab>('issues')
  const navigate = useNavigate()
  useEffect(() => {
    if (!admin) return
    let alive = true
    orgApi
      .chart()
      .then((rows) => {
        if (alive) {
          setChart(rows ?? [])
          setError('')
        }
      })
      .catch((e: Error) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [admin, reload])
  const list = admin ? (chart ?? []) : choices
  if (!divisionId && !admin && list.length === 1)
    return (
      <Navigate to={`/org/desk/${encodeURIComponent(list[0].id)}`} replace />
    )
  return (
    <main className="org-ui desk-page" dir="rtl">
      <header className="org-toolbar">
        <div>
          <p className="org-kicker">الأماني · مسؤولية وخطة ونتيجة</p>
          <h1>🏛️ مكتب الإدارة</h1>
        </div>
        <div className="org-filters">
          <label>
            الإدارة
            <select
              aria-label="اختيار الإدارة"
              value={divisionId ?? ''}
              onChange={(e) => {
                if (e.target.value)
                  navigate(`/org/desk/${encodeURIComponent(e.target.value)}`)
              }}
            >
              <option value="">اختار إدارة</option>
              {divisionId && !list.some((d) => d.id === divisionId) && (
                <option value={divisionId}>{divisionId}</option>
              )}
              {list.map((d) => (
                <option key={d.id} value={d.id}>
                  {d.name}
                </option>
              ))}
            </select>
          </label>
          <label>
            الشهر
            <input
              type="month"
              aria-label="الشهر"
              required
              value={period}
              onChange={(e) => {
                if (/^\d{4}-(0[1-9]|1[0-2])$/.test(e.target.value))
                  setPeriod(e.target.value)
              }}
            />
          </label>
        </div>
      </header>
      {error && (
        <p role="alert" className="org-error">
          تعذّر جلب أسماء الإدارات: {error}{' '}
          <button onClick={() => setReload((n) => n + 1)}>أعد المحاولة</button>
        </p>
      )}
      {!divisionId ? (
        <section className="org-panel">
          <h2>اختار المكتب الي تريد تراجعه</h2>
          {admin && !chart && !error && <p role="status">دا أجيب الإدارات…</p>}
          <div className="desk-directory">
            {list.map((d) => (
              <Link key={d.id} to={`/org/desk/${encodeURIComponent(d.id)}`}>
                {d.name} ←
              </Link>
            ))}
          </div>
        </section>
      ) : (
        <DeskContent
          key={`${divisionId}:${period}`}
          id={divisionId}
          period={period}
          admin={admin}
          reviewDivision={admin || (managementHead && divisionId !== 'admin')}
          tab={tab}
          setTab={setTab}
        />
      )}
    </main>
  )
}

type Modal =
  | { kind: 'issue' }
  | { kind: 'status'; issue: DeskIssue }
  | { kind: 'plan' }
  | { kind: 'item' }
  | { kind: 'authority' }
  | { kind: 'solution'; issue: DeskIssue }
function DeskContent({
  id,
  period,
  admin,
  tab,
  setTab,
  reviewDivision,
}: {
  id: string
  period: string
  admin: boolean
  tab: Tab
  setTab: (tab: Tab) => void
  reviewDivision: boolean
}) {
  const { employee } = useSession()
  const [data, setData] = useState<OrgDesk | null>(null)
  const [error, setError] = useState('')
  const [reload, setReload] = useState(0)
  const [modal, setModal] = useState<Modal | null>(null)
  const [notice, setNotice] = useState('')
  const [openedUnit, setOpenedUnit] = useState<string | null>(null)
  const [regulations, setRegulations] = useState<OrgClause[] | null>(null)
  const [regError, setRegError] = useState('')
  const [regReload, setRegReload] = useState(0)
  useEffect(() => {
    let alive = true
    orgRegulations()
      .then((rows) => {
        if (alive) {
          setRegulations(rows ?? [])
          setRegError('')
        }
      })
      .catch((e: Error) => {
        if (alive) setRegError(e.message)
      })
    return () => {
      alive = false
    }
  }, [regReload])
  useEffect(() => {
    let alive = true
    orgDeskGet(id, period)
      .then((result) => {
        if (alive) {
          setData(result)
          setError('')
        }
      })
      .catch((e: Error) => {
        if (alive) setError(e.message)
      })
    return () => {
      alive = false
    }
  }, [id, period, reload])
  const refresh = () => {
    setData(null)
    setError('')
    setReload((n) => n + 1)
  }
  const saved = () => {
    setModal(null)
    setNotice('انحفظ التغيير.')
    refresh()
  }
  if (error)
    return (
      <p role="alert" className="org-error">
        {error} <button onClick={refresh}>أعد المحاولة</button>
      </p>
    )
  if (!data) return <p role="status">دا أجيب مكتب الإدارة…</p>
  const issues = [...(data.issues ?? [])].sort(
    (a, b) =>
      ({ HIGH: 0, MEDIUM: 1, LOW: 2 })[a.severity] -
      { HIGH: 0, MEDIUM: 1, LOW: 2 }[b.severity],
  )
  return (
    <>
      <section className="desk-intro org-panel">
        <div>
          <h2>{data.division.name}</h2>
          <p>{data.division.summary}</p>
        </div>
        <span className="desk-badge">
          {data.canEdit
            ? 'تدير خطة الإدارة والحالات'
            : (data.editableUnits ?? []).length
              ? 'تدير خطط وحداتك، وتبلّغ وتعلّق'
              : data.canReview && reviewDivision
                ? 'تراجع خطة الإدارة، وتبلّغ وتعلّق'
                : 'للقراءة، الإبلاغ والتعليق'}
        </span>
      </section>
      {notice && (
        <p role="status" className="org-notice">
          {notice}
        </p>
      )}
      {regError && (
        <p className="org-error" role="alert">
          لائحة العمل: {regError}{' '}
          <button onClick={() => setRegReload((n) => n + 1)}>
            أعد جلب اللائحة
          </button>
        </p>
      )}
      <nav className="desk-tabs" aria-label="أقسام مكتب الإدارة">
        {(
          [
            ['issues', 'المشاكل'],
            ['plan', 'خطة التحسين'],
            ['units', 'تطوير الوحدات'],
            ['authority', 'الصلاحيات (DOA)'],
            ['team', 'الفريق'],
          ] as const
        ).map(([key, label]) => (
          <button
            key={key}
            aria-pressed={tab === key}
            onClick={() => setTab(key)}
          >
            {label}
          </button>
        ))}
      </nav>
      {tab === 'issues' && (
        <>
          <div className="org-toolbar">
            <h2>المشاكل · {number(issues.length)}</h2>
            <button
              className="org-primary"
              onClick={() => setModal({ kind: 'issue' })}
            >
              + مشكلة جديدة
            </button>
          </div>
          <div className="desk-cards">
            {issues.map((issue) => (
              <article
                key={issue.id}
                className={`desk-issue severity-${issue.severity}`}
              >
                <header>
                  <span className={`desk-badge severity-${issue.severity}`}>
                    {
                      { HIGH: 'عالية', MEDIUM: 'متوسطة', LOW: 'واطية' }[
                        issue.severity
                      ]
                    }
                  </span>
                  <span className={`desk-badge status-${issue.status}`}>
                    {statuses[issue.status]}
                  </span>
                </header>
                <h3>{issue.title}</h3>
                <h4>المشكلة</h4>
                <p>{issue.problem}</p>
                <h4>مضارها</h4>
                <p>{issue.impact || 'ما مذكورة'}</p>
                <p className="desk-meta">
                  {issue.scope === 'GENERAL' ? 'عامة' : 'خاصة'}
                  {issue.unitId &&
                    ` · ${data.division.units?.find((u) => u.id === issue.unitId)?.name ?? issue.unitId}`}
                </p>
                {issue.principle != null && (
                  <p className="desk-principle">
                    {regulations?.find((c) => c.no === issue.principle)
                      ? `${regulations.find((c) => c.no === issue.principle)!.code} · ${regulations.find((c) => c.no === issue.principle)!.title}`
                      : `الفقرة رقم ${issue.principle} · ${regulations ? 'مو موجودة باللائحة الحالية' : 'اسم الفقرة مو متوفر بعد'}`}
                  </p>
                )}
                <p className="desk-meta">
                  كتبها: {issue.authorName || 'غير متوفر'} ·{' '}
                  {date(issue.createdAt)}
                </p>
                {issue.resolution && (
                  <p className="desk-resolution">الحل: {issue.resolution}</p>
                )}
                {issue.solutionPlan && (
                  <section className="desk-solution">
                    <h4>خطة الحل</h4>
                    <p>{issue.solutionPlan}</p>
                  </section>
                )}
                {(data.canEdit ||
                  (!!employee && issue.authorId === employee.id) ||
                  (!!issue.unitId &&
                    (data.editableUnits ?? []).includes(issue.unitId))) && (
                  <button onClick={() => setModal({ kind: 'solution', issue })}>
                    خطة الحل
                  </button>
                )}
                {data.canEdit && (
                  <button onClick={() => setModal({ kind: 'status', issue })}>
                    تغيير الحالة
                  </button>
                )}
                <Comments
                  targetType="ISSUE"
                  targetId={issue.id}
                  comments={data.comments ?? []}
                  saved={saved}
                />
              </article>
            ))}
          </div>
          {!issues.length && (
            <p className="org-empty">
              ماكو مشاكل مسجلة. إذا شايف شي يحتاج تحسين، بلّغ عنه.
            </p>
          )}
        </>
      )}
      {tab === 'plan' && (
        <>
          {data.plan ? (
            <>
              <section className="org-panel desk-plan-heading">
                <ProgressRing value={data.plan.progress} />
                <div>
                  <h2>{data.plan.title}</h2>
                  <p>خطة {data.plan.period}</p>
                  {data.canEdit && canAddItems(data.plan) && (
                    <button
                      className="org-primary"
                      onClick={() => setModal({ kind: 'item' })}
                    >
                      + إضافة بند
                    </button>
                  )}
                </div>
              </section>
              <PlanControls
                plan={data.plan}
                canEdit={data.canEdit}
                canReview={data.canReview && reviewDivision}
                unit={false}
                saved={saved}
              />
              {(['PROBLEM', 'DEVELOPMENT'] as const).map((kind) => (
                <section key={kind} className="desk-plan-section">
                  <h2>
                    {kind === 'PROBLEM' ? '🔧 حل المشاكل' : '🚀 تطوير الأقسام'}
                  </h2>
                  <div className="desk-cards">
                    {(data.plan!.items ?? [])
                      .filter((item) => item.kind === kind)
                      .sort(dueOrder)
                      .map((item) => (
                        <PlanCard
                          key={item.id}
                          item={item}
                          metrics={data.metrics ?? []}
                          issues={issues}
                          canEdit={
                            data.canEdit ||
                            (!!employee && item.ownerId === employee.id)
                          }
                          comments={data.comments ?? []}
                          saved={saved}
                        />
                      ))}
                  </div>
                  {!(data.plan!.items ?? []).some((i) => i.kind === kind) && (
                    <p className="org-empty">بعد ما مضاف بند بهالقسم.</p>
                  )}
                </section>
              ))}
            </>
          ) : (
            <section className="org-panel org-empty">
              <h2>بعد ماكو خطة لهالشهر</h2>
              <p>المشاكل تنحط بخطوات قابلة للقياس، والنسبة تبين من النتيجة.</p>
              {data.canEdit && (
                <button
                  className="org-primary"
                  onClick={() => setModal({ kind: 'plan' })}
                >
                  ابدي خطة هالشهر
                </button>
              )}
            </section>
          )}
        </>
      )}
      {tab === 'authority' && (
        <section className="org-panel">
          <div className="org-toolbar">
            <div>
              <h2>جدول الصلاحيات (DOA)</h2>
              <p>شنو تقرر لحالك، وشنو ترفعه للإدارة.</p>
            </div>
            {admin && (
              <button
                className="org-primary"
                onClick={() => setModal({ kind: 'authority' })}
              >
                تعديل جدول الصلاحيات
              </button>
            )}
          </div>
          <div className="org-table-wrap">
            <table>
              <thead>
                <tr>
                  <th>القرار</th>
                  <th>طريقة الحسم</th>
                  <th>السقف بالدينار</th>
                </tr>
              </thead>
              <tbody>
                {(data.authority ?? []).map((r, i) => (
                  <tr key={`${r.decision}:${i}`}>
                    <td>{r.decision}</td>
                    <td>{r.mode === 'DECIDE' ? 'يقرر لحاله' : 'يرفعه'}</td>
                    <td>
                      {r.limitIQD == null ? (
                        'ما محدد'
                      ) : (
                        <bdi>
                          {new Intl.NumberFormat('en-US').format(r.limitIQD)}{' '}
                          د.ع
                        </bdi>
                      )}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
          {!data.authority?.length && (
            <p className="org-empty">بعد ما محدد جدول صلاحيات لهالإدارة.</p>
          )}
          {!admin && (
            <p className="desk-meta">تعديل الجدول للمالك ومدير النظام بس.</p>
          )}
        </section>
      )}
      {tab === 'units' && (
        <UnitDevelopment
          data={data}
          period={period}
          openedUnit={openedUnit}
          openUnit={setOpenedUnit}
          viewerId={employee?.id}
          saved={saved}
        />
      )}
      {tab === 'team' && <Team division={data.division} />}
      {modal?.kind === 'issue' && (
        <IssueEditor
          division={data.division}
          editableUnits={data.editableUnits ?? []}
          regulations={regulations}
          close={() => setModal(null)}
          saved={saved}
        />
      )}
      {modal?.kind === 'status' && data.canEdit && (
        <StatusEditor
          issue={modal.issue}
          close={() => setModal(null)}
          saved={saved}
        />
      )}
      {modal?.kind === 'plan' && data.canEdit && (
        <CreatePlan
          id={id}
          period={period}
          close={() => setModal(null)}
          saved={saved}
        />
      )}
      {modal?.kind === 'item' &&
        data.canEdit &&
        data.plan &&
        canAddItems(data.plan) && (
          <ItemEditor
            planId={data.plan.id}
            issues={issues}
            metrics={data.metrics ?? []}
            people={teamPeople(data.division)}
            close={() => setModal(null)}
            saved={saved}
          />
        )}
      {modal?.kind === 'authority' && admin && (
        <AuthorityEditor
          id={id}
          rows={data.authority ?? []}
          close={() => setModal(null)}
          saved={saved}
        />
      )}
      {modal?.kind === 'solution' && (
        <SolutionEditor
          issue={modal.issue}
          close={() => setModal(null)}
          saved={saved}
        />
      )}
    </>
  )
}
function SolutionEditor({
  issue,
  close,
  saved,
}: {
  issue: DeskIssue
  close: () => void
  saved: () => void
}) {
  return (
    <DeskForm
      title="خطة الحل"
      close={close}
      saved={saved}
      submit={async (d) => {
        const plan = field(d, 'plan')
        if ([...plan].length < 10)
          throw new Error('اكتب خطة الحل (١٠ حروف على الأقل)')
        return orgDeskWrite(
          `issues/${encodeURIComponent(issue.id)}/solution`,
          { plan },
          'PUT',
        )
      }}
    >
      <h3>{issue.title}</h3>
      <label>
        خطة الحل
        <textarea
          name="plan"
          aria-label="خطة الحل"
          required
          minLength={10}
          rows={5}
          defaultValue={issue.solutionPlan ?? ''}
        />
      </label>
    </DeskForm>
  )
}
const planStatusNames = {
  DRAFT: 'مسودة',
  SUBMITTED: 'مرفوعة',
  APPROVED: 'معتمدة',
  RETURNED: 'مرجعة للتعديل',
}
function PlanControls({
  plan,
  canEdit,
  canReview,
  unit,
  saved,
}: {
  plan: DeskPlan
  canEdit: boolean
  canReview: boolean
  unit: boolean
  saved: () => void
}) {
  const [action, setAction] = useState<'submit' | 'approve' | 'return' | null>(
    null,
  )
  return (
    <section className="org-panel desk-plan-controls">
      <div className="org-toolbar">
        <span className={`desk-plan-status plan-${plan.status}`}>
          {planStatusNames[plan.status] ?? 'حالة غير معروفة'}
        </span>
        {plan.submittedAt && <small>انرفعت: {date(plan.submittedAt)}</small>}
      </div>
      {plan.status === 'RETURNED' && (
        <p className="org-warning desk-review-note">
          ملاحظة الإرجاع: {plan.reviewNote || 'ما مذكورة'}
        </p>
      )}
      <div className="org-actions">
        {canEdit && canAddItems(plan) && (
          <>
            <button
              disabled={!plan.items?.length}
              onClick={() => setAction('submit')}
            >
              {unit ? 'ارفع لمدير الإدارة' : 'ارفع لمدير «الإدارة»'}
            </button>
            {!plan.items?.length && <small>ضيف بنود للخطة گبل الرفع.</small>}
          </>
        )}
        {canReview && plan.status === 'SUBMITTED' && (
          <>
            <button
              className="org-primary"
              onClick={() => setAction('approve')}
            >
              اعتماد
            </button>
            <button className="org-danger" onClick={() => setAction('return')}>
              إرجاع
            </button>
          </>
        )}
      </div>
      {action && (
        <DeskForm
          title={
            action === 'submit'
              ? 'تأكيد رفع الخطة'
              : action === 'approve'
                ? 'اعتماد الخطة'
                : 'إرجاع الخطة'
          }
          label={
            action === 'submit'
              ? 'تأكيد الرفع'
              : action === 'approve'
                ? 'تأكيد الاعتماد'
                : 'تأكيد الإرجاع'
          }
          close={() => setAction(null)}
          saved={saved}
          submit={async (d) => {
            if (action === 'submit')
              return orgDeskWrite(
                `plans/${encodeURIComponent(plan.id)}/submit`,
                {},
              )
            const note = field(d, 'note')
            if (action === 'return' && [...note].length < 5)
              throw new Error('اكتب شنو لازم يتعدل حتى ترجّعها')
            return orgDeskWrite(`plans/${encodeURIComponent(plan.id)}/review`, {
              approve: action === 'approve',
              note,
            })
          }}
        >
          <p>{plan.title}</p>
          {action === 'submit' ? (
            <p>
              الرفع يرسل الخطة للمراجعة ويوقف إضافة البنود لحد ما ترجع للتعديل.
            </p>
          ) : (
            <label>
              {action === 'return' ? 'ملاحظة الإرجاع' : 'ملاحظة (اختياري)'}
              <textarea
                name="note"
                required={action === 'return'}
                minLength={action === 'return' ? 5 : undefined}
                rows={4}
              />
            </label>
          )}
        </DeskForm>
      )}
    </section>
  )
}
function UnitDevelopment({
  data,
  period,
  openedUnit,
  openUnit,
  viewerId,
  saved,
}: {
  data: OrgDesk
  period: string
  openedUnit: string | null
  openUnit: (id: string | null) => void
  viewerId?: string
  saved: () => void
}) {
  const [create, setCreate] = useState(false),
    [add, setAdd] = useState(false)
  const units = data.division.units ?? []
  const unit = units.find((u) => u.id === openedUnit)
  const plan = (data.unitPlans ?? []).find((p) => p.unitId === openedUnit)
  const editable = !!unit && (data.editableUnits ?? []).includes(unit.id)
  if (!unit)
    return (
      <div className="desk-cards">
        {units.map((u) => {
          const p = (data.unitPlans ?? []).find((p) => p.unitId === u.id)
          return (
            <section className="org-panel desk-unit-summary" key={u.id}>
              <h2>{u.name}</h2>
              {p ? (
                <>
                  <span className={`desk-plan-status plan-${p.status}`}>
                    {planStatusNames[p.status] ?? 'حالة غير معروفة'}
                  </span>
                  <p>الإنجاز: {number(p.progress)}٪</p>
                  <progress
                    max={100}
                    value={percent(p.progress)}
                    aria-label={`إنجاز وحدة ${u.name}`}
                  />
                  <button onClick={() => openUnit(u.id)}>افتح {u.name}</button>
                </>
              ) : (data.editableUnits ?? []).includes(u.id) ? (
                <button
                  className="org-primary"
                  onClick={() => {
                    openUnit(u.id)
                    setCreate(true)
                  }}
                >
                  ابدي خطة الوحدة · {u.name}
                </button>
              ) : (
                <p className="desk-meta">ما رفعت خطة بعد</p>
              )}
            </section>
          )
        })}
        {!units.length && (
          <p className="org-empty">بعد ما مضافة وحدات لهالإدارة.</p>
        )}
      </div>
    )
  return (
    <section className="desk-unit-plan">
      <div className="org-toolbar">
        <h2>تطوير {unit.name}</h2>
        <button
          onClick={() => {
            setCreate(false)
            setAdd(false)
            openUnit(null)
          }}
        >
          رجع للوحدات ←
        </button>
      </div>
      {plan ? (
        <>
          <section className="org-panel desk-plan-heading">
            <ProgressRing value={plan.progress} />
            <div>
              <h3>{plan.title}</h3>
              <p>{plan.period}</p>
              {editable && canAddItems(plan) && (
                <button className="org-primary" onClick={() => setAdd(true)}>
                  + إضافة بند للوحدة
                </button>
              )}
            </div>
          </section>
          <PlanControls
            plan={plan}
            canEdit={editable}
            canReview={data.canReview && data.canEdit}
            unit
            saved={saved}
          />
          <div className="desk-cards">
            {[...(plan.items ?? [])].sort(dueOrder).map((item) => (
              <PlanCard
                key={item.id}
                item={item}
                metrics={data.metrics ?? []}
                issues={data.issues ?? []}
                canEdit={editable || (!!viewerId && item.ownerId === viewerId)}
                comments={data.comments ?? []}
                saved={saved}
              />
            ))}
          </div>
          {!plan.items?.length && (
            <p className="org-empty">ضيف أول بند وحدد المسؤول والموعد.</p>
          )}
          {add && editable && canAddItems(plan) && (
            <ItemEditor
              planId={plan.id}
              issues={data.issues ?? []}
              metrics={data.metrics ?? []}
              people={teamPeople(data.division)}
              close={() => setAdd(false)}
              saved={saved}
            />
          )}
        </>
      ) : (
        <section className="org-panel">
          <p>ما رفعت خطة بعد</p>
          {editable && (
            <button onClick={() => setCreate(true)}>ابدي خطة الوحدة</button>
          )}
        </section>
      )}
      {create && editable && !plan && (
        <CreatePlan
          id={data.division.id}
          unit={unit}
          period={period}
          close={() => setCreate(false)}
          saved={saved}
        />
      )}
    </section>
  )
}
function ProgressRing({ value }: { value: number }) {
  const progress = percent(value)
  return (
    <div
      className="desk-ring"
      role="img"
      aria-label={`إنجاز الخطة ${progress}%`}
    >
      <svg viewBox="0 0 120 120" aria-hidden="true">
        <circle cx="60" cy="60" r="50" className="desk-ring-track" />
        <circle
          cx="60"
          cy="60"
          r="50"
          className="desk-ring-value"
          strokeDasharray={`${progress * Math.PI} ${100 * Math.PI}`}
          transform="rotate(-90 60 60)"
        />
      </svg>
      <span>
        {number(progress)}٪<small>إنجاز الخطة</small>
      </span>
    </div>
  )
}
function Comments({
  targetType,
  targetId,
  comments,
  saved,
}: {
  targetType: 'ISSUE' | 'PLAN_ITEM'
  targetId: string
  comments: DeskComment[]
  saved: () => void
}) {
  const [text, setText] = useState(''),
    [error, setError] = useState(''),
    [busy, setBusy] = useState(false)
  const lock = useRef(false)
  async function submit() {
    if (lock.current || text.trim().length < 2) return
    lock.current = true
    setBusy(true)
    setError('')
    try {
      await orgDeskWrite('comments', {
        targetType,
        targetId,
        text: text.trim(),
      })
      saved()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      lock.current = false
      setBusy(false)
    }
  }
  return (
    <section className="desk-comments" aria-label="التعليقات">
      <h4>
        {targetType === 'PLAN_ITEM' ? 'تعليقات وملاحظات النائب' : 'التعليقات'}
      </h4>
      {comments
        .filter((c) => c.targetType === targetType && c.targetId === targetId)
        .map((c, i) => (
          <div className="desk-comment" key={c.id ?? `${c.createdAt}:${i}`}>
            <strong>{c.authorName || 'غير متوفر'}</strong>
            <small>{date(c.createdAt)}</small>
            <p>{c.text}</p>
          </div>
        ))}
      <form
        onSubmit={(e) => {
          e.preventDefault()
          void submit()
        }}
      >
        <label>
          تعليق جديد
          <textarea
            minLength={2}
            required
            disabled={busy}
            value={text}
            onChange={(e) => setText(e.target.value)}
            rows={2}
          />
        </label>
        {error && (
          <p className="org-error" role="alert">
            {error}
          </p>
        )}
        <button disabled={busy || text.trim().length < 2}>
          {busy ? 'دا أرسل…' : 'إرسال التعليق'}
        </button>
      </form>
    </section>
  )
}
function PlanCard({
  item,
  metrics,
  issues,
  canEdit,
  comments,
  saved,
}: {
  item: DeskItem
  metrics: DeskMetric[]
  issues: DeskIssue[]
  canEdit: boolean
  comments: DeskComment[]
  saved: () => void
}) {
  const [manual, setManual] = useState(item.manualProgress),
    [busy, setBusy] = useState(false),
    [error, setError] = useState('')
  const lock = useRef(false),
    metric = metrics.find((m) => m.key === item.metric)
  async function update(body: { manualProgress?: number; done?: boolean }) {
    if (
      !canEdit ||
      lock.current ||
      ((item.byMatrix || !!item.metric) && body.manualProgress != null)
    )
      return
    lock.current = true
    setBusy(true)
    setError('')
    try {
      await orgDeskWrite(
        `plan-items/${encodeURIComponent(item.id)}`,
        body,
        'PATCH',
      )
      saved()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      lock.current = false
      setBusy(false)
    }
  }
  return (
    <article className="org-panel desk-plan-card">
      <div className="org-toolbar">
        <h3>{item.title}</h3>
        <strong>{number(item.progress)}٪</strong>
      </div>
      <progress
        aria-label={`إنجاز ${item.title}`}
        max={100}
        value={percent(item.progress)}
      />
      {item.issueId && (
        <p className="desk-meta">
          المشكلة:{' '}
          {issues.find((i) => i.id === item.issueId)?.title ??
            'المشكلة المرتبطة مو متوفرة'}
        </p>
      )}
      <div className="desk-item-meta">
        <span>المسؤول: {item.ownerName || 'مو محدد'}</span>
        <span>
          الموعد: <bdi>{item.dueDate || 'مو محدد'}</bdi>
        </span>
        {item.overdue && <span className="desk-overdue">متأخر</span>}
      </div>
      {item.byMatrix || item.metric ? (
        <>
          <span className="desk-badge matrix">يحسبها ماتركس</span>
          <div className="desk-measurement">
            <span>
              الأساس
              <strong>
                {number(item.baseline)} {metric?.unit}
              </strong>
            </span>
            <span aria-hidden="true">←</span>
            <span>
              الحالي
              <strong>
                {number(item.current)} {metric?.unit}
              </strong>
            </span>
            <span aria-hidden="true">←</span>
            <span>
              الهدف
              <strong>
                {number(item.target)} {metric?.unit}
              </strong>
            </span>
          </div>
          <p className="desk-meta">
            {metric?.label ?? item.metric}
            {metric && ` · ${metric.lowerBetter ? 'أقل أحسن' : 'أكثر أحسن'}`}
          </p>
          {item.current == null && (
            <p className="desk-meta">
              الرقم الحالي مو متوفر لهالفترة؛ ما نعتبره صفر.
            </p>
          )}
        </>
      ) : (
        canEdit &&
        !item.doneAt && (
          <div className="desk-manual">
            <label>
              نسبة الإنجاز اليدوية: {number(manual)}٪
              <input
                type="range"
                min={0}
                max={100}
                step={1}
                disabled={busy}
                value={manual}
                onChange={(e) => setManual(Number(e.target.value))}
              />
            </label>
            <button
              disabled={busy || manual === item.manualProgress}
              onClick={() => void update({ manualProgress: manual })}
            >
              احفظ النسبة
            </button>
          </div>
        )
      )}
      {error && (
        <p role="alert" className="org-error">
          {error}
        </p>
      )}
      {item.doneAt ? (
        <p className="desk-resolution">✓ خلص · {date(item.doneAt)}</p>
      ) : (
        canEdit && (
          <button disabled={busy} onClick={() => void update({ done: true })}>
            خلص ✓
          </button>
        )
      )}
      <Comments
        targetType="PLAN_ITEM"
        targetId={item.id}
        comments={comments}
        saved={saved}
      />
    </article>
  )
}
function Team({ division }: { division: OrgDivision }) {
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

function DeskForm({
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
const field = (data: FormData, key: string) =>
  String(data.get(key) ?? '').trim()
function IssueEditor({
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
function StatusEditor({
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
function CreatePlan({
  id,
  period,
  unit,
  close,
  saved,
}: {
  id: string
  period: string
  unit?: OrgUnit
  close: () => void
  saved: () => void
}) {
  return (
    <DeskForm
      title={unit ? 'ابدي خطة الوحدة' : 'ابدي خطة هالشهر'}
      close={close}
      saved={saved}
      submit={(d) =>
        orgDeskWrite('plans', {
          divisionId: id,
          period,
          ...(unit ? { unitId: unit.id } : {}),
          title: field(d, 'title'),
        })
      }
    >
      <p>خطة الشهر: {period}</p>
      <label>
        عنوان الخطة
        <input
          name="title"
          required
          defaultValue={`خطة ${unit?.name || 'التحسين'} ${period}`}
        />
      </label>
    </DeskForm>
  )
}
function ItemEditor({
  planId,
  issues,
  metrics,
  people,
  close,
  saved,
}: {
  planId: string
  issues: DeskIssue[]
  metrics: DeskMetric[]
  people: OrgPerson[]
  close: () => void
  saved: () => void
}) {
  const [metric, setMetric] = useState('')
  const selected = metrics.find((m) => m.key === metric)
  return (
    <DeskForm
      title="إضافة بند للخطة"
      close={close}
      saved={saved}
      submit={(d) =>
        orgDeskWrite(`plans/${encodeURIComponent(planId)}/items`, {
          kind: field(d, 'kind'),
          title: field(d, 'title'),
          ...(field(d, 'issueId') ? { issueId: field(d, 'issueId') } : {}),
          ...(metric ? { metric, target: Number(field(d, 'target')) } : {}),
          ...(field(d, 'dueDate') ? { dueDate: field(d, 'dueDate') } : {}),
          ...(field(d, 'ownerId') ? { ownerId: field(d, 'ownerId') } : {}),
        })
      }
    >
      <label>
        النوع
        <select name="kind" aria-label="نوع البند">
          <option value="PROBLEM">🔧 حل مشكلة</option>
          <option value="DEVELOPMENT">🚀 تطوير الأقسام</option>
        </select>
      </label>
      <label>
        عنوان البند
        <input name="title" minLength={3} required />
      </label>
      <label>
        الموعد
        <input name="dueDate" type="date" />
      </label>
      <label>
        المسؤول
        <select name="ownerId" aria-label="المسؤول عن البند">
          <option value="">مو محدد</option>
          {people.map((p) => (
            <option key={p.employeeId} value={p.employeeId}>
              {p.employeeName}
            </option>
          ))}
        </select>
      </label>
      <label>
        ربط بمشكلة (اختياري)
        <select name="issueId" aria-label="ربط بمشكلة">
          <option value="">بلا ربط</option>
          {issues.map((i) => (
            <option key={i.id} value={i.id}>
              {i.title}
            </option>
          ))}
        </select>
      </label>
      <label>
        المقياس (اختياري)
        <select
          aria-label="المقياس"
          value={metric}
          onChange={(e) => setMetric(e.target.value)}
        >
          <option value="">نسبة يدوية</option>
          {metrics.map((m) => (
            <option key={m.key} value={m.key}>
              {m.label}
            </option>
          ))}
        </select>
      </label>
      {selected && (
        <>
          <label>
            الهدف · {selected.unit} ·{' '}
            {selected.lowerBetter ? 'أقل أحسن' : 'أكثر أحسن'}
            <input
              name="target"
              type="number"
              step="any"
              min={0}
              max={selected.unit === '%' ? 100 : undefined}
              required
            />
          </label>
          <p className="desk-meta">
            خط الأساس يجي من النظام؛ الهدف لازم يكون أحسن منه. إذا ماكو أرقام
            سابقة، اختار نسبة يدوية.
          </p>
        </>
      )}
    </DeskForm>
  )
}
function AuthorityEditor({
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
