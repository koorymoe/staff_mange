import { useEffect, useState } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import { orgApi, type OrgDivision } from '../../org'
import { useSession } from '../../session'
import DirectiveManager from '../../components/DirectiveManager'
import MyOrgCard from '../../components/MyOrgCard'
import MyDirectives from '../../components/MyDirectives'
import OrgOutputs from '../../components/OrgOutputs'
import type { DirectiveTarget } from '../../orgDirectives'
import { orgDeskGet, orgMonth, orgRegulations, type OrgClause, type OrgDesk } from '../../orgDesk'
import { type Tab, canAddItems, dueOrder, teamPeople, number, statuses, date, type Modal } from './shared'
import { SolutionEditor, PlanControls, UnitDevelopment, ProgressRing, Comments, PlanCard, CreatePlan, ItemEditor } from './plan'
import { Team, IssueEditor, StatusEditor, AuthorityEditor } from './editors'

export function DeskDirectory({
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

export function DeskContent({
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
  // (دازاي 11-01) النائب اطّلاع بس — والمدير بالوكالة يجيه تعيين DIRECTOR فيصير يقرر
  const deputyHere =
    !data.canEdit &&
    [...(data.division.people ?? []), ...(data.division.units ?? []).flatMap((u) => u.people ?? [])].some(
      (p) => p.employeeId === employee?.id && p.position === 'DEPUTY',
    )
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
      {deputyHere && (
        <p role="note" className="org-notice">
          👁️ أنت نائب بهذا المكان — اطّلاع بس، والقرار للمدير. تصير مدير بالوكالة تلقائياً لمن المدير ياخذ إجازة موافق عليها، وترجع نائب لمن يرجع.
        </p>
      )}
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
            // (دازاي 11-01) «مكاني» جوّا مكتب الإدارة — مو خيار ثاني بالقائمة
            ['mine', '🏛️ مكاني'],
            ['issues', 'المشاكل'],
            ['plan', 'خطة التحسين'],
            ['units', 'تطوير الوحدات'],
            ['directives', 'التوجيهات'],
            ['outputs', 'المخرجات'],
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
      {tab === 'mine' && <><MyDirectives /><MyOrgCard /></>}
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
                  {issue.source === 'MATRIX' && (
                    <span className="desk-badge matrix">🤖 شخّصها ماتركس</span>
                  )}
                  {issue.source === 'MONITOR' && (
                    <span className="desk-badge">👁️ من المراقب</span>
                  )}
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
            <p className="desk-meta">تعديل الجدول للمالك ومدير النظام بس. الجدول مرجع مكتوب للمدير وماتركس وكمال: «يقرر» يحسمه المدير لحد السقف، و«يرفع» لازم يرفعه للإدارة.</p>
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
      {tab === 'outputs' && <OfficeOutputs data={data} period={period} />}
      {tab === 'directives' && <DirectiveManager divisionId={data.division.id} targets={[
        ...(admin ? [{scope:'COMPANY',label:'الشركة كلها'} as DirectiveTarget] : []),
        ...(data.canEdit ? [{scope:'DIVISION',divisionId:data.division.id,label:data.division.name} as DirectiveTarget] : []),
        ...(data.division.units ?? []).filter(u => (data.editableUnits ?? []).includes(u.id)).map(u => ({scope:'UNIT' as const,divisionId:data.division.id,unitId:u.id,label:`${data.division.name} / ${u.name}`})),
      ]}/>}
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

export function OfficeOutputs({ data, period }: { data: OrgDesk; period: string }) {
  const [scope, setScope] = useState<'division' | 'unit'>('division')
  const [unitId, setUnitId] = useState('')
  // Visible office units are not necessarily units whose outputs we may read.
  const units = (data.division.units ?? []).filter((unit) =>
    (data.editableUnits ?? []).includes(unit.id),
  )
  const selected = units.find((unit) => unit.id === unitId)
  return (
    <section className="org-panel">
      <div className="org-actions">
        <button aria-pressed={scope === 'division'} onClick={() => setScope('division')}>
          مخرجات الإدارة
        </button>
        {units.length > 0 && (
          <button aria-pressed={scope === 'unit'} onClick={() => {
            setScope('unit')
            if (!selected) setUnitId(units[0].id)
          }}>
            مخرجات الوحدة
          </button>
        )}
      </div>
      {scope === 'unit' && (
        <label>وحدة المخرجات
          <select aria-label="وحدة المخرجات" value={unitId} onChange={(e) => setUnitId(e.target.value)}>
            {units.map((unit) => <option key={unit.id} value={unit.id}>{unit.name}</option>)}
          </select>
        </label>
      )}
      {scope === 'division' ? (
        <OrgOutputs key={`division:${data.division.id}`} target={{ kind: 'division', id: data.division.id }} allowed initialPeriod={period} />
      ) : selected && (
        <OrgOutputs key={`unit:${selected.id}`} target={{ kind: 'unit', id: selected.id }} allowed initialPeriod={period} />
      )}
    </section>
  )
}
