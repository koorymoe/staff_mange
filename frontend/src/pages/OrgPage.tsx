import { useEffect, useRef, useState, type ReactNode } from 'react'
import { Link } from 'react-router-dom'
import { api, fileUrl, type Employee } from '../api'
import { accessApi, type AccessGroup } from '../permissionAccess'
import {
  orgApi,
  orgPeople,
  orgPositions,
  type OrgDivision,
  type OrgPerson,
  type OrgPosition,
  type OrgUnit,
} from '../org'
import { useSession } from '../session'
import DirectiveManager from '../components/DirectiveManager'
import OrgOutputs from '../components/OrgOutputs'
import OrgAssignmentPreview from '../components/OrgAssignmentPreview'
import type { DirectiveTarget } from '../orgDirectives'
import { roleLabel } from '../roleLabels'
import { matches } from '../utils/search'
import './OrgPage.css'

type Selection = {
  employeeId?: string
  divisionId?: string
  unitId?: string
  position?: OrgPosition
}
type Loaded = {
  chart: OrgDivision[] | null
  employees: Employee[] | null
  catalog: AccessGroup[] | null
  errors: string[]
}
const money = (n: number) => new Intl.NumberFormat('en-US').format(n)
function Avatar({ name, photo }: { name: string; photo?: string | null }) {
  const [failed, setFailed] = useState(false)
  return (
    <span className="org-avatar" aria-hidden="true">
      {photo && !failed ? (
        <img src={fileUrl(photo)} alt="" onError={() => setFailed(true)} />
      ) : (
        name.trim().slice(0, 1) || '؟'
      )}
    </span>
  )
}
function Dialog({
  title,
  busy,
  close,
  children,
}: {
  title: string
  busy: boolean
  close: () => void
  children: ReactNode
}) {
  const ref = useRef<HTMLDialogElement>(null)
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
  return (
    <dialog
      ref={ref}
      className="org-ui org-dialog"
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
          disabled={busy}
          onClick={close}
          aria-label="إغلاق النافذة"
        >
          ✕
        </button>
      </header>
      {children}
    </dialog>
  )
}

export default function OrgPage({
  view = 'chart',
}: {
  view?: 'chart' | 'assign'
}) {
  const { employee } = useSession()
  return employee?.actualRole === 'OWNER' || employee?.role === 'ADMIN' ? (
    <OrgWorkspace key={`${employee.id}:${view}`} view={view} />
  ) : (
    <p role="alert">الهيكل التنظيمي للمالك ومدير النظام بس.</p>
  )
}
function OrgWorkspace({ view }: { view: 'chart' | 'assign' }) {
  const { employee: currentUser } = useSession()
  const [data, setData] = useState<Loaded | null>(null)
  const [reload, setReload] = useState(0)
  const [loading, setLoading] = useState(true)
  const [selection, setSelection] = useState<Selection | null>(null)
  const [person, setPerson] = useState<OrgPerson | null>(null)
  const [permissionUnit, setPermissionUnit] = useState<OrgUnit | null>(null)
  const [filter, setFilter] = useState('')
  const [search, setSearch] = useState('')
  const [notice, setNotice] = useState('')
  const [showDirectives, setShowDirectives] = useState(false)
  useEffect(() => {
    let alive = true
    Promise.allSettled([
      orgApi.chart(),
      api.getEmployees(),
      accessApi.catalog(),
    ]).then(([c, e, p]) => {
      if (!alive) return
      setData({
        chart: c.status === 'fulfilled' ? (c.value ?? []) : null,
        employees: e.status === 'fulfilled' ? e.value : null,
        catalog: p.status === 'fulfilled' ? p.value : null,
        errors: [c, e, p].flatMap((r, i) =>
          r.status === 'rejected'
            ? [
                `${['الهيكل', 'الموظفين', 'أسماء الصلاحيات'][i]}: ${(r.reason as Error).message}`,
              ]
            : [],
        ),
      })
      setLoading(false)
    })
    return () => {
      alive = false
    }
  }, [reload])
  const refresh = () => {
    setLoading(true)
    setReload((n) => n + 1)
  }
  const chart = data?.chart ?? []
  const employees = data?.employees ?? []
  const people = orgPeople(chart)
  const active = employees.filter((e) => e.status === 'ACTIVE')
  const canManage = (id: string) => {
    if (currentUser?.actualRole === 'OWNER') return true
    const target = employees.find((e) => e.id === id)
    return (
      !!target &&
      id !== currentUser?.id &&
      !['OWNER', 'ADMIN'].includes(target.actualRole ?? target.role)
    )
  }
  const unplaced = active.filter(
    (e) => !people.some((p) => p.employeeId === e.id),
  )
  const labels = new Map(
    data?.catalog?.flatMap((g) =>
      g.perms.map((p) => [p.name, p.label] as const),
    ) ?? [],
  )
  const assigned = (s: Selection) => {
    setSelection(s)
    setNotice('')
  }
  const saved = () => {
    setSelection(null)
    setPerson(null)
    setNotice('انحفظ التغيير.')
    refresh()
  }
  const seat = (
    rows: OrgPerson[],
    position: 'DIRECTOR' | 'DEPUTY',
    place: Selection,
  ) => {
    const occupants = rows.filter((p) => p.position === position)
    return (
      <div className="org-seat">
        <small>{orgPositions[position]}</small>
        {occupants.length ? (
          occupants.map((p) => (
            <button
              key={p.id}
              className="org-person"
              disabled={loading}
              onClick={() => setPerson(p)}
            >
              <Avatar
                name={p.employeeName}
                photo={employees.find((e) => e.id === p.employeeId)?.photoUrl}
              />
              <span>
                {p.employeeName}
                {p.acting && <small>بالوكالة</small>}
              </span>
            </button>
          ))
        ) : (
          <div className="org-vacant">
            <span>بلا {orgPositions[position]}</span>
            <button
              disabled={loading || !data?.employees}
              onClick={() => assigned({ ...place, position })}
            >
              عيّن
            </button>
          </div>
        )}
      </div>
    )
  }
  const members = (rows: OrgPerson[]) => (
    <div className="org-members">
      <small>
        الأعضاء:{' '}
        {rows
          .filter((p) => p.position === 'MEMBER')
          .length.toLocaleString('ar-IQ')}
      </small>
      <div>
        {rows
          .filter((p) => p.position === 'MEMBER')
          .map((p) => (
            <button key={p.id} disabled={loading} onClick={() => setPerson(p)}>
              {p.employeeName}
            </button>
          ))}
      </div>
    </div>
  )
  function division(d: OrgDivision, visited: Set<string>): ReactNode {
    if (visited.has(d.id))
      return (
        <p role="alert" key={d.id}>
          تعذّر ترتيب «{d.name}»: حلقة بتسلسل الإدارات.
        </p>
      )
    const path = new Set([...visited, d.id])
    const local = (d.people ?? []).filter((p) => !p.unitId)
    const children = chart.filter((c) => c.parentId === d.id)
    return (
      <div className="org-branch" key={d.id}>
        <DivisionPanel name={d.name} summary={d.summary}>
          <div className="org-seats">
            {seat(local, 'DIRECTOR', { divisionId: d.id })}
            {seat(local, 'DEPUTY', { divisionId: d.id })}
          </div>
          {members(local)}
          <button
            className="org-add"
            disabled={loading || !data?.employees}
            onClick={() => assigned({ divisionId: d.id, position: 'MEMBER' })}
          >
            + عيّن بالإدارة
          </button>
          <div className="org-units">
            {(d.units ?? []).map((u) => (
              <section className="org-unit" key={u.id}>
                <button
                  className="org-unit-title"
                  onClick={() => setPermissionUnit(u)}
                >
                  {u.name}
                  <small>شنو تفتح هالوحدة؟ ←</small>
                </button>
                <div className="org-seats">
                  {seat(u.people ?? [], 'DIRECTOR', {
                    divisionId: d.id,
                    unitId: u.id,
                  })}
                  {seat(u.people ?? [], 'DEPUTY', {
                    divisionId: d.id,
                    unitId: u.id,
                  })}
                </div>
                {members(u.people ?? [])}
                <button
                  className="org-add"
                  disabled={loading || !data?.employees}
                  onClick={() =>
                    assigned({
                      divisionId: d.id,
                      unitId: u.id,
                      position: 'MEMBER',
                    })
                  }
                >
                  + عيّن بالوحدة
                </button>
              </section>
            ))}
          </div>
        </DivisionPanel>
        {children.length > 0 && (
          <div className="org-children">
            {children.map((child) => division(child, path))}
          </div>
        )}
      </div>
    )
  }
  const roots = chart.filter(
    (d) => !d.parentId || !chart.some((p) => p.id === d.parentId),
  )
  const reachable = new Set<string>()
  const visit = (id: string) => {
    if (reachable.has(id)) return
    reachable.add(id)
    chart.filter((d) => d.parentId === id).forEach((d) => visit(d.id))
  }
  roots.forEach((d) => visit(d.id))
  const remaining = chart.filter((d) => !reachable.has(d.id))
  const shown = active.filter(
    (e) =>
      matches([e.name, e.position, roleLabel(e.role)], search) &&
      (!filter ||
        (filter === 'none'
          ? unplaced.some((p) => p.id === e.id)
          : people.some(
              (p) => p.employeeId === e.id && p.divisionId === filter,
            ))),
  )
  return (
    <main className="org-ui org-page" dir="rtl">
      <header className="org-toolbar">
        <div>
          <p className="org-kicker">شركة الأماني · المرحلة الأولى</p>
          <h1>{view === 'chart' ? '🏛️ الهيكل التنظيمي' : 'ترتيب الموظفين'}</h1>
          <p>
            إدارات ووحدات، ومسؤولية واضحة لكل مكان. الأدوار القديمة باقية مثل ما
            هي.
          </p>
        </div>
        <nav>
          <Link to="/org">الشجرة</Link>
          <Link to="/org/assign">ترتيب الموظفين</Link>
          <button disabled={loading || !data?.chart} aria-expanded={showDirectives} onClick={() => setShowDirectives(v => !v)}>📣 إصدار توجيه</button>
          <button
            className="org-primary"
            disabled={loading || !data?.chart || !data?.employees}
            onClick={() => assigned({})}
          >
            + تعيين جديد
          </button>
        </nav>
      </header>
      {showDirectives && data?.chart && <section className="org-panel"><DirectiveManager targets={[
        {scope:'COMPANY',label:'الشركة كلها'},
        ...chart.flatMap(d => [{scope:'DIVISION' as const,divisionId:d.id,label:d.name},...(d.units ?? []).map(u => ({scope:'UNIT' as const,divisionId:d.id,unitId:u.id,label:`${d.name} / ${u.name}`}))]),
      ] as DirectiveTarget[]}/></section>}
      {notice && (
        <p role="status" className="org-notice">
          {notice}
        </p>
      )}
      {!!data?.errors.length && (
        <div role="alert" className="org-error">
          {data.errors.map((e) => (
            <p key={e}>{e}</p>
          ))}
          <button disabled={loading} onClick={refresh}>
            أعد المحاولة
          </button>
        </div>
      )}
      {loading && <p role="status">دا أحدّث الهيكل والموظفين…</p>}
      {view === 'chart' && data?.chart && (
        <>
          <section className="org-top">
            <span>الإدارة العليا</span>
            <h2>مدير الشركة (المالك) + مدير النظام</h2>
            <p>الإدارات والوحدات أدناه حسب التسلسل المعتمد بالنظام.</p>
          </section>
          {chart.length === 0 ? (
            <p className="org-panel">بعد ما موجود هيكل تنظيمي.</p>
          ) : (
            <div className="org-tree">
              {roots.map((d) => division(d, new Set()))}
              {remaining.length > 0 && (
                <section className="org-panel">
                  <p role="alert">إدارات تحتاج مراجعة التسلسل الأبوي.</p>
                  {remaining.map((d) => division(d, new Set()))}
                </section>
              )}
            </div>
          )}
        </>
      )}
      {view === 'assign' && data?.chart && data.employees && (
        <>
          <section className="org-panel org-toolbar">
            <h2>
              {loading ? '—' : unplaced.length.toLocaleString('ar-IQ')} موظف بلا
              مكان
            </h2>
            <div className="org-filters">
              <label>
                فلتر المكان
                <select
                  aria-label="فلتر المكان"
                  disabled={loading}
                  value={filter}
                  onChange={(e) => setFilter(e.target.value)}
                >
                  <option value="">الكل</option>
                  <option value="none">بلا مكان</option>
                  {chart.map((d) => (
                    <option value={d.id} key={d.id}>
                      {d.name}
                    </option>
                  ))}
                </select>
              </label>
              <label>
                بحث بالموظفين
                <input
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  placeholder="اسم الموظف…"
                />
              </label>
            </div>
          </section>
          <div className="org-table-wrap">
            <table>
              <thead>
                <tr>
                  <th>الموظف</th>
                  <th>مكانه الحالي</th>
                  <th>الإجراء</th>
                </tr>
              </thead>
              <tbody>
                {shown.map((e) => (
                  <tr key={e.id}>
                    <td>
                      <div className="org-person">
                        <Avatar name={e.name} photo={e.photoUrl} />
                        <span>
                          {e.name}
                          <small>{roleLabel(e.role)}</small>
                        </span>
                      </div>
                    </td>
                    <td>
                      {people.some((p) => p.employeeId === e.id) ? (
                        <ul>
                          {people
                            .filter((p) => p.employeeId === e.id)
                            .map((p) => (
                              <li key={p.id}>
                                <button
                                  disabled={loading}
                                  onClick={() => setPerson(p)}
                                >
                                  {chart.find((d) => d.id === p.divisionId)
                                    ?.name ?? p.divisionId}{' '}
                                  {p.unitId &&
                                    `· ${chart.flatMap((d) => d.units ?? []).find((u) => u.id === p.unitId)?.name ?? p.unitId}`}{' '}
                                  · {orgPositions[p.position]}
                                </button>
                              </li>
                            ))}
                        </ul>
                      ) : (
                        <span className="org-unplaced">بلا مكان</span>
                      )}
                    </td>
                    <td>
                      <button
                        disabled={loading || !canManage(e.id)}
                        onClick={() => assigned({ employeeId: e.id })}
                      >
                        {canManage(e.id) ? 'عيّن' : 'المالك وحده'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
            {shown.length === 0 && (
              <p className="org-empty">ماكو موظفين بهالفلتر.</p>
            )}
          </div>
        </>
      )}
      {selection && data?.chart && data.employees && (
        <AssignDialog
          chart={chart}
          catalog={data.catalog}
          employees={active.filter((e) => canManage(e.id))}
          initial={selection}
          close={() => setSelection(null)}
          saved={saved}
        />
      )}
      {person && (
        <PersonDialog
          person={person}
          canEnd={canManage(person.employeeId)}
          chart={chart}
          close={() => setPerson(null)}
          saved={saved}
        />
      )}
      {permissionUnit && (
        <Dialog
          title={`صلاحيات ${permissionUnit.name}`}
          busy={false}
          close={() => setPermissionUnit(null)}
        >
          <p>المنصب هو الي يفتح صلاحيات الوحدة تلقائياً.</p>
          {!data?.catalog && (
            <p role="alert">
              تعذّر جلب الأسماء العربية. أعد تحميل الهيكل حتى تنعرض.
            </p>
          )}
          <ul className="org-permissions">
            {(permissionUnit.permissions ?? []).map((p) => (
              <li key={p}>
                {labels.get(p) ?? `صلاحية غير معرّفة بالقاموس (${p})`}
              </li>
            ))}
          </ul>
          {!permissionUnit.permissions?.length && (
            <p>ماكو صلاحيات مضافة لهالوحدة.</p>
          )}
        </Dialog>
      )}
    </main>
  )
}
function DivisionPanel({
  name,
  summary,
  children,
}: {
  name: string
  summary: string
  children: ReactNode
}) {
  const [open, setOpen] = useState(() => window.innerWidth >= 768)
  return (
    <details
      className="org-division"
      open={open}
      onToggle={(e) => setOpen(e.currentTarget.open)}
    >
      <summary>
        <h2>{name}</h2>
        <p>{summary}</p>
        <small>افتح تفاصيل الإدارة ⌄</small>
      </summary>
      <div className="org-division-body">{children}</div>
    </details>
  )
}
function AssignDialog({
  chart,
  catalog,
  employees,
  initial,
  close,
  saved,
}: {
  chart: OrgDivision[]
  catalog: AccessGroup[] | null
  employees: Employee[]
  initial: Selection
  close: () => void
  saved: () => void
}) {
  const [employeeId, setEmployeeId] = useState(initial.employeeId ?? '')
  const [query, setQuery] = useState('')
  const [divisionId, setDivisionId] = useState(initial.divisionId ?? '')
  const [unitId, setUnitId] = useState(initial.unitId ?? '')
  const [position, setPosition] = useState<OrgPosition>(
    initial.position ?? 'MEMBER',
  )
  const [allowance, setAllowance] = useState('0')
  const [note, setNote] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const lock = useRef(false)
  const digits = allowance
    .replace(/[٠-٩]/g, (d) => String(d.charCodeAt(0) - 1632))
    .replace(/[۰-۹]/g, (d) => String(d.charCodeAt(0) - 1776))
    .replace(/[,٬\s]/g, '')
  const amount = Number(digits)
  const validAmount = /^\d+$/.test(digits) && Number.isSafeInteger(amount)
  const small = validAmount && amount > 0 && amount < 1000
  const filtered = employees.filter(
    (e) =>
      e.id === employeeId ||
      matches([e.name, e.position, roleLabel(e.role)], query),
  )
  const division = chart.find((d) => d.id === divisionId)
  const submit = async () => {
    if (lock.current || !employeeId || !divisionId || !validAmount || small)
      return
    lock.current = true
    setBusy(true)
    setError('')
    try {
      await orgApi.assign({
        employeeId,
        divisionId,
        ...(unitId ? { unitId } : {}),
        position,
        allowanceIQD: amount,
        note: note.trim(),
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
    <Dialog title="تعيين بالهيكل" busy={busy} close={close}>
      <form
        onSubmit={(e) => {
          e.preventDefault()
          void submit()
        }}
        className="org-form"
      >
        <label>
          دوّر على موظف
          <input
            value={query}
            disabled={busy}
            onChange={(e) => setQuery(e.target.value)}
            placeholder="اكتب اسم الموظف…"
          />
        </label>
        <label>
          الموظف
          <select
            required
            disabled={busy}
            value={employeeId}
            aria-label="الموظف"
            onChange={(e) => setEmployeeId(e.target.value)}
          >
            <option value="">اختار موظف</option>
            {filtered.map((e) => (
              <option key={e.id} value={e.id}>
                {e.name} · {roleLabel(e.role)} · {e.id.slice(-6)}
              </option>
            ))}
          </select>
        </label>
        <label>
          الإدارة
          <select
            required
            disabled={busy}
            value={divisionId}
            aria-label="الإدارة"
            onChange={(e) => {
              setDivisionId(e.target.value)
              setUnitId('')
            }}
          >
            <option value="">اختار إدارة</option>
            {chart.map((d) => (
              <option value={d.id} key={d.id}>
                {d.name}
              </option>
            ))}
          </select>
        </label>
        <label>
          الوحدة (اختياري)
          <select
            disabled={busy || !division}
            value={unitId}
            aria-label="الوحدة (اختياري)"
            onChange={(e) => setUnitId(e.target.value)}
          >
            <option value="">على مستوى الإدارة</option>
            {(division?.units ?? []).map((u) => (
              <option key={u.id} value={u.id}>
                {u.name}
              </option>
            ))}
          </select>
        </label>
        <label>
          المنصب
          <select
            disabled={busy}
            value={position}
            aria-label="المنصب"
            onChange={(e) => setPosition(e.target.value as OrgPosition)}
          >
            {Object.entries(orgPositions).map(([value, label]) => (
              <option key={value} value={value}>
                {label}
              </option>
            ))}
          </select>
        </label>
        <label>
          المخصص بالدينار
          <input
            required
            dir="ltr"
            inputMode="numeric"
            value={allowance}
            disabled={busy}
            aria-invalid={!validAmount || small}
            onChange={(e) => {
              const raw = e.target.value
                .replace(/[٠-٩]/g, (d) => String(d.charCodeAt(0) - 1632))
                .replace(/[۰-۹]/g, (d) => String(d.charCodeAt(0) - 1776))
                .replace(/[,٬\s]/g, '')
              setAllowance(
                /^\d+$/.test(raw) && Number.isSafeInteger(Number(raw))
                  ? money(Number(raw))
                  : raw,
              )
            }}
          />
        </label>
        {small && (
          <p role="alert" className="org-warning">
            كتبت {money(amount)} دينار بس. إذا تقصد ألوف، اكتب المبلغ كامل مثل
            50,000. الصفر مسموح.
          </p>
        )}
        {!validAmount && <p role="alert">اكتب مبلغ صحيح بالدينار، بلا كسور.</p>}
        <label className="org-wide">
          ملاحظة
          <textarea
            value={note}
            disabled={busy}
            onChange={(e) => setNote(e.target.value)}
            rows={3}
          />
        </label>
        <OrgAssignmentPreview
          division={division}
          unitId={unitId}
          position={position}
          catalog={catalog}
        />
        {error && (
          <p className="org-error org-wide" role="alert">
            {error}
          </p>
        )}
        <div className="org-wide org-actions">
          <button
            className="org-primary"
            type="submit"
            disabled={
              busy || !validAmount || small || !employeeId || !divisionId
            }
          >
            {busy ? 'دا أحفظ…' : 'حفظ التعيين'}
          </button>
          <button type="button" disabled={busy} onClick={close}>
            إلغاء
          </button>
        </div>
      </form>
    </Dialog>
  )
}
function PersonDialog({
  person,
  canEnd,
  chart,
  close,
  saved,
}: {
  person: OrgPerson
  canEnd: boolean
  chart: OrgDivision[]
  close: () => void
  saved: () => void
}) {
  const [ending, setEnding] = useState(false)
  const [outputsOpen, setOutputsOpen] = useState(false)
  const [reason, setReason] = useState('')
  const [error, setError] = useState('')
  const [busy, setBusy] = useState(false)
  const lock = useRef(false)
  const division = chart.find((d) => d.id === person.divisionId)
  const date = new Date(person.startedAt)
  const end = async () => {
    if (!reason.trim() || lock.current) return
    lock.current = true
    setBusy(true)
    setError('')
    try {
      await orgApi.end(person.id, reason.trim())
      saved()
    } catch (e) {
      setError((e as Error).message)
    } finally {
      lock.current = false
      setBusy(false)
    }
  }
  return (
    <Dialog title={`منصب ${person.employeeName}`} busy={busy} close={close}>
      <dl className="org-person-info">
        <dt>المكان</dt>
        <dd>
          {division?.name ?? person.divisionId}{' '}
          {person.unitId &&
            `· ${division?.units?.find((u) => u.id === person.unitId)?.name ?? person.unitId}`}
        </dd>
        <dt>المنصب</dt>
        <dd>
          {orgPositions[person.position]}
          {person.acting && ' · بالوكالة'}
        </dd>
        <dt>المخصص</dt>
        <dd>
          <bdi>{money(person.allowanceIQD)}</bdi> د.ع
        </dd>
        <dt>من متى</dt>
        <dd>
          {Number.isNaN(date.getTime())
            ? 'التاريخ مو متوفر'
            : new Intl.DateTimeFormat('ar-IQ', {
                timeZone: 'Asia/Baghdad',
                dateStyle: 'medium',
              }).format(date)}
        </dd>
        <dt>ملاحظة</dt>
        <dd>{person.note || 'ماكو ملاحظة'}</dd>
      </dl>
      <button aria-expanded={outputsOpen} onClick={() => setOutputsOpen((v) => !v)}>
        🎯 المخرجات
      </button>
      {outputsOpen && <OrgOutputs target={{ kind: 'employee', id: person.employeeId }} />}
      {!canEnd ? (
        <p>إنهاء هالمنصب للمالك وحده.</p>
      ) : !ending ? (
        <button className="org-danger" onClick={() => setEnding(true)}>
          إنهاء المنصب
        </button>
      ) : (
        <form
          onSubmit={(e) => {
            e.preventDefault()
            void end()
          }}
        >
          <h3>تأكيد إنهاء المنصب</h3>
          <p>ينتهي هالمنصب فقط، مو حساب الموظف ولا مناصبه الثانية.</p>
          <label>
            سبب الإنهاء
            <textarea
              required
              value={reason}
              disabled={busy}
              onChange={(e) => setReason(e.target.value)}
              rows={3}
            />
          </label>
          {error && (
            <p className="org-error" role="alert">
              {error}
            </p>
          )}
          <div className="org-actions">
            <button
              className="org-danger"
              disabled={busy || !reason.trim()}
              type="submit"
            >
              {busy ? 'دا أنهي…' : 'تأكيد الإنهاء'}
            </button>
            <button
              disabled={busy}
              type="button"
              onClick={() => setEnding(false)}
            >
              تراجع
            </button>
          </div>
        </form>
      )}
    </Dialog>
  )
}
