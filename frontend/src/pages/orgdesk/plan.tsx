import { useRef, useState } from 'react'
import { type OrgPerson, type OrgUnit } from '../../org'
import { orgDeskWrite, type DeskPlan, type DeskComment, type DeskIssue, type DeskItem, type DeskMetric, type OrgDesk } from '../../orgDesk'
import { canAddItems, dueOrder, teamPeople, number, percent, date, planStatusNames, field } from './shared'
import { DeskForm } from './editors'
import { isAdminOrOwner, useSession } from '../../session'

export function SolutionEditor({
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

export function PlanControls({
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
  // (دازاي 11-08) المالك ومدير النظام يرجّعون الخطة للتعديل حتى بعد الاعتماد
  const { employee } = useSession()
  const admin = isAdminOrOwner(employee)
  const [action, setAction] = useState<
    'submit' | 'approve' | 'return' | 'ai' | 'reopen' | null
  >(null)
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
            <button className="org-primary" onClick={() => setAction('ai')}>
              🤖 ماتركس يكتب الخطة
            </button>
            <button
              disabled={!plan.items?.length}
              onClick={() => setAction('submit')}
            >
              {unit ? 'ارفع لمدير الإدارة' : 'ارفع لمدير «الإدارة»'}
            </button>
            {!plan.items?.length && <small>ضيف بنود للخطة گبل الرفع.</small>}
          </>
        )}
        {admin && plan.status === 'APPROVED' && (
          <button className="org-danger" onClick={() => setAction('reopen')}>
            ↩️ ارجعها للتعديل
          </button>
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
            action === 'ai'
              ? '🤖 ماتركس يكتب خطة التحسين'
              : action === 'submit'
              ? 'تأكيد رفع الخطة'
              : action === 'approve'
                ? 'اعتماد الخطة'
                : action === 'reopen'
                  ? 'إرجاع الخطة المعتمدة للتعديل'
                  : 'إرجاع الخطة'
          }
          label={
            action === 'ai'
              ? 'اكتب المسودة'
              : action === 'submit'
              ? 'تأكيد الرفع'
              : action === 'approve'
                ? 'تأكيد الاعتماد'
                : 'تأكيد الإرجاع'
          }
          close={() => setAction(null)}
          saved={saved}
          submit={async (d) => {
            if (action === 'ai')
              return orgDeskWrite(
                `plans/${encodeURIComponent(plan.id)}/draft-ai`,
                { guidance: field(d, 'guidance') },
              )
            if (action === 'submit')
              return orgDeskWrite(
                `plans/${encodeURIComponent(plan.id)}/submit`,
                {},
              )
            const note = field(d, 'note')
            if ((action === 'return' || action === 'reopen') && [...note].length < 5)
              throw new Error('اكتب شنو لازم يتعدل حتى ترجّعها')
            if (action === 'reopen')
              return orgDeskWrite(`plans/${encodeURIComponent(plan.id)}/reopen`, { note })
            return orgDeskWrite(`plans/${encodeURIComponent(plan.id)}/review`, {
              approve: action === 'approve',
              note,
            })
          }}
        >
          <p>{plan.title}</p>
          {action === 'ai' ? (
            <>
              <p>
                ماتركس يقرا أرقام الوحدة الحقيقية (المشاكل المفتوحة، والمقاييس،
                والخطط السابقة) ويكتب ٤–٧ بنود بمواعيد، ويوزعها على المدير
                والنائب. تگدر تعدّل أو تشيل أي بند قبل الرفع، ومن تنعتمد كل بند
                يصير مهمة بموعده.
              </p>
              <label>
                توجيهك لماتركس (اختياري)
                <textarea
                  name="guidance"
                  rows={3}
                  placeholder="مثلاً: ركّز على تأخير التسليم وتدريب الكادر الجديد"
                />
              </label>
            </>
          ) : action === 'submit' ? (
            <p>
              الرفع يرسل الخطة للمراجعة ويوقف إضافة البنود لحد ما ترجع للتعديل.
            </p>
          ) : (
            <label>
              {action === 'return' || action === 'reopen' ? 'ملاحظة الإرجاع' : 'ملاحظة (اختياري)'}
              <textarea
                name="note"
                required={action === 'return' || action === 'reopen'}
                minLength={action === 'return' || action === 'reopen' ? 5 : undefined}
                rows={4}
              />
            </label>
          )}
        </DeskForm>
      )}
    </section>
  )
}

export function UnitDevelopment({
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
  const { employee: viewer } = useSession()
  const admin = isAdminOrOwner(viewer)
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
              {editable && canAddItems(plan, admin) && (
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
          {add && editable && canAddItems(plan, admin) && (
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

export function ProgressRing({ value }: { value: number }) {
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

export function Comments({
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

export function PlanCard({
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
      {item.byMatrixDraft && (
        <p className="desk-meta">
          🤖 كتبه ماتركس{item.reason ? ` — ليش: ${item.reason}` : ''}
        </p>
      )}
      {item.taskId && <TaskChip item={item} />}
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

export function CreatePlan({
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

export function ItemEditor({
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

// (دازاي 11-08) البند كمهمة: حالة المهمة المربوطة بنفس مكان البند
function TaskChip({ item }: { item: DeskItem }) {
  const d = (s?: string | null) => (s ? new Date(s).toLocaleDateString('en-GB', { timeZone: 'Asia/Baghdad' }) : '')
  if (item.taskStatus === 'DONE' || item.taskDoneAt) {
    return <p className="desk-meta">📋 المهمة: ✅ خلصت {d(item.taskDoneAt)}</p>
  }
  if (item.taskStatus === 'CANCELLED') return <p className="desk-meta">📋 المهمة: انلغت</p>
  const late = item.taskLateDays ?? 0
  if (late >= 1) {
    return <p className="desk-meta desk-overdue">📋 المهمة: ⚠️ متأخرة {late} يوم (موعدها {d(item.taskDueAt)})</p>
  }
  return <p className="desk-meta">📋 المهمة: ⏳ شغّالة{item.taskDueAt ? ` — موعدها ${d(item.taskDueAt)}` : ''}</p>
}
