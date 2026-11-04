import {
  useCallback,
  useEffect,
  useLayoutEffect,
  useId,
  useRef,
  useState,
  type ReactNode,
} from 'react'
import { api, ApiError, type Employee } from '../api'
import {
  accessApi,
  type AccessAgent,
  type AccessGroup,
  type AccessState,
  type AccessAudit,
  type RoleDrift,
} from '../permissionAccess'
import { useSession } from '../session'
import { roleLabel } from '../roleLabels'
import { matches } from '../utils/search'
import EmployeeAdminScope from '../components/EmployeeAdminScope'
import './PermissionsPage.css'

const sourceLabel = {
  TEMPLATE: 'قالب المنصب',
  MANUAL: 'منح يدوي',
  EXCEPTION: 'استثناء معتمد',
  ORG: 'منصب بالهيكل',
}
const tiers = ['quotation_create', 'quotation_edit_own', 'quotation_manage_all']
function exceptionDeadline(value: string) {
  if (!value) return null
  const at = new Date(`${value}:00+03:00`)
  if (!Number.isFinite(at.getTime()) || at.getTime() <= Date.now()) {
    throw new Error('موعد الانتهاء لازم بالمستقبل، بتوقيت بغداد.')
  }
  return at
}
function baghdadInput(at: string) {
  const parts = Object.fromEntries(
    new Intl.DateTimeFormat('en-GB', {
      timeZone: 'Asia/Baghdad',
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
      hourCycle: 'h23',
    })
      .formatToParts(new Date(at))
      .map((part) => [part.type, part.value]),
  )
  return `${parts.year}-${parts.month}-${parts.day}T${parts.hour}:${parts.minute}`
}
const agentPermission = {
  MATRIX: 'agent_matrix',
  QUALITY: 'agent_quality',
  BOOKING: 'agent_booking',
}
const defaults: AccessAgent[] = [
  { agent: 'MATRIX', name: 'ماتركس', editable: false },
  { agent: 'IRIS', name: 'ايريس', editable: false },
  { agent: 'QUALITY', name: 'وكيل الجودة', editable: false },
  { agent: 'BOOKING', name: 'وكيل الحجوزات', editable: false },
]
const time = (at: string) =>
  new Intl.DateTimeFormat('ar-IQ', {
    timeZone: 'Asia/Baghdad',
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(at))
const conflict = (e: unknown) => e instanceof ApiError && e.status === 409
const message = (e: unknown) =>
  e instanceof Error ? e.message : 'تعذر إكمال العملية. جرّب مرة ثانية.'

function Dialog({
  title,
  children,
  cancel,
  busy,
}: {
  title: string
  children: ReactNode
  cancel: () => void
  busy: boolean
}) {
  const id = useId()
  const ref = useRef<HTMLElement>(null)
  useEffect(() => {
    const previous = document.activeElement as HTMLElement | null
    const overflow = document.body.style.overflow
    document.body.style.overflow = 'hidden'
    const first = ref.current?.querySelector<HTMLElement>(
      'textarea, input, button',
    )
    first?.focus()
    return () => {
      document.body.style.overflow = overflow
      if (previous?.isConnected) previous.focus()
    }
  }, [])
  return (
    <div
      className="access-overlay"
      onKeyDown={(e) => {
        if (e.key === 'Escape' && !busy) {
          e.stopPropagation()
          cancel()
        }
        if (e.key === 'Tab') {
          const items = [
            ...(ref.current?.querySelectorAll<HTMLElement>(
              'button:not(:disabled), input:not(:disabled), textarea:not(:disabled), select:not(:disabled), [tabindex="0"]',
            ) ?? []),
          ]
          const first = items[0],
            last = items.at(-1)
          if (e.shiftKey && document.activeElement === first) {
            e.preventDefault()
            last?.focus()
          } else if (!e.shiftKey && document.activeElement === last) {
            e.preventDefault()
            first?.focus()
          }
        }
      }}
    >
      <section
        ref={ref}
        className="access-dialog"
        role="dialog"
        aria-modal="true"
        aria-labelledby={id}
      >
        <header>
          <h3 id={id}>{title}</h3>
          <button
            type="button"
            aria-label="غلق النافذة"
            disabled={busy}
            onClick={cancel}
          >
            ×
          </button>
        </header>
        {children}
      </section>
    </div>
  )
}

export default function PermissionsPage({
  embedded = false,
}: { embedded?: boolean } = {}) {
  const { employee } = useSession()
  const owner = employee?.actualRole === 'OWNER'
  if (!owner && employee?.role !== 'ADMIN')
    return (
      <p className="p-6" role="alert">
        الصلاحيات والوصول للمالك ومدير النظام فقط.
      </p>
    )
  return <AccessPage owner={owner} embedded={embedded} />
}

function AccessPage({
  owner,
  embedded,
}: {
  owner: boolean
  embedded: boolean
}) {
  const [groups, setGroups] = useState<AccessGroup[] | null>(null)
  const [employees, setEmployees] = useState<Employee[]>([])
  const [selected, setSelected] = useState('')
  const [tab, setTab] = useState<'access' | 'drift'>('access')
  const [search, setSearch] = useState('')
  const [dirty, setDirty] = useState(false)
  const [working, setWorking] = useState(false)
  const [navigation, setNavigation] = useState<{
    employee?: string
    tab?: 'access' | 'drift'
  } | null>(null)
  const [reload, setReload] = useState(0)
  const [error, setError] = useState('')
  const [agents, setAgents] = useState<AccessAgent[]>(defaults)
  const [agentError, setAgentError] = useState('')
  const [agentReload, setAgentReload] = useState(0)
  const [editAgent, setEditAgent] = useState<AccessAgent | null>(null)
  const [agentName, setAgentName] = useState('')
  const onDirty = useCallback((value: boolean) => setDirty(value), [])
  const onWorking = useCallback((value: boolean) => setWorking(value), [])
  useEffect(() => {
    let alive = true
    Promise.all([accessApi.catalog(), api.getEmployees()])
      .then(([catalog, rows]) => {
        if (alive) {
          setGroups(catalog)
          setEmployees(rows)
          setError('')
        }
      })
      .catch((e) => {
        if (alive) setError(message(e))
      })
    return () => {
      alive = false
    }
  }, [reload])
  // (لوفي 10-27) أسماء الوكلاء صارت تنقرا للأدمن هم — التعديل يبقى للمالك.
  useEffect(() => {
    let alive = true
    accessApi
      .agents()
      .then((rows) => {
        if (alive) {
          setAgents(rows)
          setAgentError('')
        }
      })
      .catch((e) => {
        if (alive) setAgentError(message(e))
      })
    return () => {
      alive = false
    }
  }, [agentReload])
  useEffect(() => {
    if (!dirty) return
    const warn = (event: BeforeUnloadEvent) => {
      event.preventDefault()
      event.returnValue = ''
    }
    window.addEventListener('beforeunload', warn)
    return () => window.removeEventListener('beforeunload', warn)
  }, [dirty])
  const navigate = (next: { employee?: string; tab?: 'access' | 'drift' }) => {
    if (next.employee === selected || next.tab === tab) return
    if (dirty) {
      setNavigation(next)
      return
    }
    if (next.employee !== undefined) setSelected(next.employee)
    if (next.tab) setTab(next.tab)
  }
  const rename = async () => {
    if (!editAgent || !owner || working) return
    const name = agentName.trim()
    if ([...name].length < 2 || [...name].length > 30) {
      setAgentError('الاسم من حرفين إلى ٣٠ حرف.')
      return
    }
    setWorking(true)
    try {
      await accessApi.renameAgent(editAgent.agent, name)
      setEditAgent(null)
      setAgentReload((n) => n + 1)
    } catch (e) {
      setAgentError(message(e))
    } finally {
      setWorking(false)
    }
  }
  const allPerms = groups?.flatMap((g) => g.perms) ?? []
  const label = (name: string) =>
    allPerms.find((p) => p.name === name)?.label ?? name
  return (
    <div className="access-page" dir="rtl">
      {!embedded && (
        <header className="access-heading">
          <span>إدارة النظام / الوصول</span>
          <h2>الصلاحيات والوصول</h2>
          <p>انطِ كل موظف اللي يحتاجه، وشوف مصدر كل صلاحية قبل ما تغيّرها.</p>
        </header>
      )}
      <div
        className="access-tabs"
        role="tablist"
        aria-label="الصلاحيات والوصول"
      >
        <button
          role="tab"
          aria-selected={tab === 'access'}
          disabled={working}
          onClick={() => navigate({ tab: 'access' })}
        >
          صلاحيات الموظف
        </button>
        <button
          role="tab"
          aria-selected={tab === 'drift'}
          disabled={working}
          onClick={() => navigate({ tab: 'drift' })}
        >
          مطابقة الصنف
        </button>
      </div>
      {error && (
        <div className="access-error" role="alert">
          {error}{' '}
          <button onClick={() => setReload((n) => n + 1)}>أعد المحاولة</button>
        </div>
      )}
      {!groups && !error && <p role="status">دا أجيب مجموعات الصلاحيات…</p>}
      {groups && tab === 'access' && (
        <>
          <section className="access-card access-picker">
            <label>
              البحث عن موظف
              <input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="الاسم أو المنصب…"
              />
            </label>
            <label>
              اختر الموظف
              <select
                value={selected}
                disabled={working}
                onChange={(e) => navigate({ employee: e.target.value })}
              >
                <option value="">اختار موظف حتى تشوف صلاحياته</option>
                {employees
                  .filter(
                    (e) =>
                      e.id === selected ||
                      matches([e.name, e.jobTitle, e.position], search),
                  )
                  .map((e) => (
                    <option key={e.id} value={e.id}>
                      {e.name} · {roleLabel(e.role)}
                    </option>
                  ))}
              </select>
            </label>
          </section>
          {!selected ? (
            <div className="access-card access-empty">
              اختار الموظف من فوق. التغييرات ما تنحفظ إلا بعد مراجعتك وتأكيدك.
            </div>
          ) : (
            <AccessEditor
              key={selected}
              id={selected}
              groups={groups}
              owner={owner}
              agents={agents}
              agentError={agentError}
              retryAgents={() => setAgentReload((n) => n + 1)}
              renameAgent={(agent) => {
                setAgentName(agent.name)
                setEditAgent(agent)
              }}
              onDirty={onDirty}
              onWorking={onWorking}
            />
          )}
        </>
      )}
      {groups && tab === 'drift' && (
        <DriftPage groups={groups} label={label} onWorking={onWorking} />
      )}
      {navigation && (
        <Dialog
          title="عندك تغييرات ما انحفظت"
          busy={false}
          cancel={() => setNavigation(null)}
        >
          <p>إذا تكمل، نلغي المسوّدة فقط. الصلاحيات المحفوظة ما تتغيّر.</p>
          <div className="access-actions">
            <button
              className="access-primary"
              onClick={() => {
                setDirty(false)
                if (navigation.employee !== undefined)
                  setSelected(navigation.employee)
                if (navigation.tab) setTab(navigation.tab)
                setNavigation(null)
              }}
            >
              اترك المسوّدة وكمل
            </button>
            <button onClick={() => setNavigation(null)}>ابقَ هنا</button>
          </div>
        </Dialog>
      )}
      {editAgent && (
        <Dialog
          title={`تسمية ${editAgent.name}`}
          busy={working}
          cancel={() => setEditAgent(null)}
        >
          <form
            onSubmit={(e) => {
              e.preventDefault()
              void rename()
            }}
          >
            <label>
              الاسم الظاهر
              <input
                autoFocus
                required
                minLength={2}
                maxLength={30}
                value={agentName}
                onChange={(e) => setAgentName(e.target.value)}
              />
            </label>
            <p>هذا اسم عام بالنظام، مو اسم خاص لهذا الموظف.</p>
            {agentError && (
              <p role="alert" className="access-error">
                {agentError}
              </p>
            )}
            <div className="access-actions">
              <button className="access-primary" disabled={working}>
                حفظ الاسم
              </button>
              <button
                type="button"
                disabled={working}
                onClick={() => setEditAgent(null)}
              >
                إلغاء
              </button>
            </div>
          </form>
        </Dialog>
      )}
    </div>
  )
}

function AccessEditor({
  id,
  groups,
  owner,
  agents,
  agentError,
  retryAgents,
  renameAgent,
  onDirty,
  onWorking,
}: {
  id: string
  groups: AccessGroup[]
  owner: boolean
  agents: AccessAgent[]
  agentError: string
  retryAgents: () => void
  renameAgent: (agent: AccessAgent) => void
  onDirty: (value: boolean) => void
  onWorking: (value: boolean) => void
}) {
  const [state, setState] = useState<AccessState | null>(null)
  const [draft, setDraft] = useState<string[]>([])
  const [search, setSearch] = useState('')
  const [reload, setReload] = useState(0)
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const [busy, setBusy] = useState(false)
  const [stale, setStale] = useState(false)
  const [review, setReview] = useState(false)
  const [reason, setReason] = useState('')
  const [exception, setException] = useState<{
    name: string
    removing: boolean
  } | null>(null)
  const [expires, setExpires] = useState('')
  const [audit, setAudit] = useState<AccessAudit[] | null>(null)
  const [auditError, setAuditError] = useState('')
  const [auditReload, setAuditReload] = useState(0)
  useEffect(() => {
    let alive = true
    accessApi
      .get(id)
      .then((value) => {
        if (alive) {
          setState(value)
          setDraft(value.perms.map((p) => p.name))
          setStale(false)
          setError('')
        }
      })
      .catch((e) => {
        if (alive) setError(message(e))
      })
    return () => {
      alive = false
    }
  }, [id, reload])
  useEffect(() => {
    let alive = true
    accessApi
      .audit(id)
      .then((rows) => {
        if (alive) {
          setAudit(rows)
          setAuditError('')
        }
      })
      .catch((e) => {
        if (alive) setAuditError(message(e))
      })
    return () => {
      alive = false
    }
  }, [id, auditReload])
  const perms = groups.flatMap((g) => g.perms)
  const label = (name: string) =>
    perms.find((p) => p.name === name)?.label ?? name
  const saved = state?.perms.map((p) => p.name) ?? []
  const add = draft.filter((p) => !saved.includes(p))
  const remove = saved.filter((p) => !draft.includes(p))
  const dirty = Boolean(state && (add.length || remove.length))
  useLayoutEffect(() => {
    onDirty(dirty)
    return () => onDirty(false)
  }, [dirty, onDirty])
  useLayoutEffect(() => {
    onWorking(busy)
    return () => onWorking(false)
  }, [busy, onWorking])
  const targetLocked =
    !owner &&
    (state?.employee.role === 'ADMIN' || state?.employee.role === 'OWNER')
  const disabled = !state || busy || stale || targetLocked
  const can = (name: string) =>
    !disabled && perms.some((p) => p.name === name && p.grantable)
  const toggle = (name: string, group: AccessGroup) => {
    if (!can(name)) return
    setSuccess('')
    setDraft((current) => {
      if (current.includes(name))
        return current.filter(
          (p) => p !== name && !(name !== group.bundle && p === group.bundle),
        )
      if (name === group.bundle)
        return [
          ...new Set([
            ...current,
            ...group.perms.filter((p) => p.grantable).map((p) => p.name),
          ]),
        ]
      return [
        ...current.filter((p) => !tiers.includes(name) || !tiers.includes(p)),
        name,
      ]
    })
  }
  const failure = (e: unknown) => {
    setError(message(e))
    if (conflict(e)) setStale(true)
  }
  const refresh = () => {
    setState(null)
    setReview(false)
    setException(null)
    setReload((n) => n + 1)
    setAuditReload((n) => n + 1)
  }
  const save = async () => {
    if (!state || disabled || !dirty || reason.trim().length < 3) return
    setBusy(true)
    setError('')
    try {
      const value = await accessApi.save(id, {
        add,
        remove,
        reason: reason.trim(),
        version: state.version,
      })
      setState(value)
      setDraft(value.perms.map((p) => p.name))
      setReview(false)
      setReason('')
      setSuccess('انحفظت الصلاحيات وسجّلنا سبب التغيير.')
      setAuditReload((n) => n + 1)
    } catch (e) {
      failure(e)
    } finally {
      setBusy(false)
    }
  }
  const saveException = async () => {
    if (!exception || !state || !can(exception.name) || dirty) return
    if (!exception.removing && reason.trim().length < 3) return
    setBusy(true)
    setError('')
    try {
      if (exception.removing)
        await accessApi.removeException(id, exception.name, state.version)
      else {
        const at = exceptionDeadline(expires)
        await accessApi.exception({
          employeeId: id,
          permission: exception.name,
          reason: reason.trim(),
          version: state.version,
          ...(at ? { expiresAt: at.toISOString() } : {}),
        })
      }
      setException(null)
      setSuccess('انحفظ تعديل الاستثناء.')
      refresh()
    } catch (e) {
      failure(e)
    } finally {
      setBusy(false)
    }
  }
  const row = (name: string, group: AccessGroup, title?: string) => {
    const perm = group.perms.find((p) => p.name === name)
    if (!perm)
      return (
        <p className="access-muted">
          صلاحية {title ?? name} مو معرّفة بالكتالوگ.
        </p>
      )
    const original = state?.perms.find((p) => p.name === name)
    return (
      <div className="access-permission" key={name}>
        <div>
          <strong>{title ?? perm.label}</strong>
          <small>
            {!perm.grantable
              ? `🔒 ${perm.lockReason ?? 'مقفولة'}`
              : original
                ? sourceLabel[original.source]
                : state?.missingFromTemplate.includes(name)
                  ? 'ناقصة من قالب المنصب'
                  : 'غير ممنوحة'}
          </small>
          {original?.reason && <small>السبب: {original.reason}</small>}
          {original?.expiresAt && (
            <small>تنتهي: {time(original.expiresAt)}</small>
          )}
          {(add.includes(name) || remove.includes(name)) && (
            <small className="access-pending">
              {add.includes(name)
                ? 'راح تنضاف بعد الحفظ'
                : 'راح تنشال بعد الحفظ'}
            </small>
          )}
          {perm.grantable && !targetLocked && (
            <button
              className="access-link"
              disabled={disabled || dirty}
              onClick={() => {
                setException({ name, removing: false })
                setExpires(
                  original?.expiresAt ? baghdadInput(original.expiresAt) : '',
                )
                setReason(original?.reason ?? '')
                setError('')
              }}
            >
              {original?.source === 'EXCEPTION'
                ? 'تعديل الاستثناء / مدته'
                : 'استثناء أو صلاحية مؤقتة'}
            </button>
          )}
          {original?.source === 'EXCEPTION' &&
            perm.grantable &&
            !targetLocked && (
              <button
                className="access-link"
                disabled={disabled || dirty}
                onClick={() => {
                  setException({ name, removing: true })
                  setError('')
                }}
              >
                إزالة تثبيت الاستثناء
              </button>
            )}
        </div>
        <button
          className="access-switch"
          type="button"
          role="switch"
          aria-label={title ?? perm.label}
          aria-checked={draft.includes(name)}
          disabled={!can(name)}
          onClick={() => toggle(name, group)}
        >
          <span />
        </button>
      </div>
    )
  }
  const quotation = groups
    .flatMap((g) => g.perms)
    .filter((p) => tiers.includes(p.name))
  const displayGroups = [
    ...groups.map((g) => ({
      ...g,
      perms: g.perms.filter((p) => !tiers.includes(p.name)),
    })),
    ...(quotation.length
      ? [
          {
            key: 'quotation',
            title: 'عروض الأسعار — اختار درجة وحدة',
            perms: quotation,
          },
        ]
      : []),
  ]
  const shownGroups = displayGroups.filter(
    (g) =>
      (g.perms.length || g.key === 'agents') &&
      matches(
        [
          g.title,
          ...g.perms.flatMap((p) => [p.label, p.name]),
          ...(g.key === 'agents' ? agents.map((agent) => agent.name) : []),
        ],
        search,
      ),
  )
  return (
    <>
      {error && (
        <div className="access-error" role="alert">
          {error}
          {(stale || !state) && (
            <button disabled={busy} onClick={refresh}>
              حدّث وراجع — يلغي المسوّدة
            </button>
          )}
        </div>
      )}
      {success && (
        <p className="access-success" role="status">
          {success}
        </p>
      )}
      {!state ? (
        !error && <p role="status">دا أجيب صلاحيات الموظف…</p>
      ) : (
        <>
          <section className="access-card access-employee">
            <div>
              <span className="access-eyebrow">صلاحيات الموظف</span>
              <h3>{state.employee.name}</h3>
              <p>
                {[state.employee.role, ...(state.employee.secondaryRoles ?? [])]
                  .map(roleLabel)
                  .join(' · ')}
                {state.employee.isLeader ? ' · ليدر' : ''}
              </p>
            </div>
            <span className="access-badge">{saved.length} صلاحية مسجّلة</span>
          </section>
          {targetLocked && (
            <p className="access-notice">
              🔒 صلاحيات المالك والمدير يعدّلها المالك بس. العرض متاح بدون
              تعديل.
            </p>
          )}
          {(state.employee.role === 'ADMIN' ||
            state.employee.role === 'OWNER') && (
            <p className="access-notice">
              الوصول الإداري يعتمد على الدور أيضاً؛ إزالة مفتاح هنا ما تلغي
              صلاحيات الدور الإداري.
            </p>
          )}
          <label className="access-search">
            دوّر على صلاحية
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="اسم الصلاحية أو المجموعة…"
            />
          </label>
          {search && !shownGroups.length && (
            <p className="access-notice">
              ما لگينا صلاحية بهذا الاسم. جرّب كلمة أقصر أو امسح البحث.
            </p>
          )}
          {shownGroups.map((group) => {
            const grantablePerms = group.perms.filter((p) => p.grantable)
            const complete =
              grantablePerms.length > 0 &&
              grantablePerms.every((p) => draft.includes(p.name))
            return (
              <details
                className="access-card access-group"
                key={`${group.key}-${search ? 'search' : 'all'}`}
                open={search ? true : undefined}
              >
                <summary>
                  <strong>{group.title}</strong>
                  <span>
                    {group.perms.filter((p) => draft.includes(p.name)).length}{' '}
                    من {group.perms.length} مفعّلة
                  </span>
                </summary>
                {group.key !== 'quotation' && (
                  <div className="access-group-tools">
                    <button
                      disabled={
                        disabled ||
                        complete ||
                        !group.perms.some((p) => p.grantable)
                      }
                      onClick={() => {
                        setSuccess('')
                        setDraft((current) => [
                          ...new Set([
                            ...current,
                            ...group.perms
                              .filter((p) => p.grantable)
                              .map((p) => p.name),
                          ]),
                        ])
                      }}
                    >
                      {!grantablePerms.length
                        ? 'ماكو صلاحيات قابلة للمنح'
                        : complete
                          ? 'الحزمة مفعّلة'
                          : 'انطِ الحزمة كاملة'}
                    </button>
                    <small>القابل للمنح فقط. ما تنحفظ لحد ما تراجع.</small>
                  </div>
                )}
                {'bundle' in group && group.bundle && (
                  <p className="access-notice">
                    مفتاح الحزمة يفتح المجموعة كاملة. إطفاء صلاحية فرعية يطفي
                    مفتاح الحزمة بالمسوّدة حتى يصير الاختيار جزئي.
                  </p>
                )}
                {group.key === 'agents' ? (
                  <>
                    {agentError && (
                      <p className="access-error">
                        {agentError}{' '}
                        <button onClick={retryAgents}>
                          أعد جلب أسماء الوكلاء
                        </button>
                      </p>
                    )}
                    <p className="access-notice">
                      التوجيه ما يمنح الموظف صلاحيات إضافية، ولا يوسّع صلاحيات
                      الوكيل.
                    </p>
                    {agents.map((agent) => (
                      <div className="access-agent" key={agent.agent}>
                        <header>
                          <strong>🤖 {agent.name}</strong>
                          {owner && agent.editable ? (
                            <button
                              disabled={busy}
                              onClick={() => renameAgent(agent)}
                            >
                              تعديل الاسم
                            </button>
                          ) : (
                            <small>
                              {agent.agent === 'MATRIX' ||
                              agent.agent === 'IRIS'
                                ? 'اسم ثابت'
                                : 'الاسم العام يعدّله المالك'}
                            </small>
                          )}
                        </header>
                        {agent.agent === 'IRIS' ? (
                          <p>
                            🔒 للمدير وحده — المالك ومدير النظام، بلا مفتاح منح.
                          </p>
                        ) : (
                          row(
                            agentPermission[agent.agent],
                            group,
                            `توجيه ${agent.name}`,
                          )
                        )}
                      </div>
                    ))}
                    {group.perms
                      .filter(
                        (p) => !Object.values(agentPermission).includes(p.name),
                      )
                      .map((p) => row(p.name, group))}
                  </>
                ) : (
                  group.perms
                    .filter((p) =>
                      matches([p.name, p.label, group.title], search),
                    )
                    .map((p) => row(p.name, group))
                )}
              </details>
            )
          })}
          {dirty && (
            <div className="access-savebar">
              <strong>
                {add.length} إضافات · {remove.length} إزالة
              </strong>
              <span>مسوّدة — بعد ما انحفظت</span>
              <button
                className="access-primary"
                disabled={disabled}
                onClick={() => {
                  setReview(true)
                  setReason('')
                  setError('')
                }}
              >
                راجع واحفظ
              </button>
              <button
                disabled={busy}
                onClick={() => {
                  setDraft(saved)
                  setSuccess('')
                }}
              >
                إلغاء المسوّدة
              </button>
            </div>
          )}
          {!dirty &&
            !targetLocked &&
            !stale &&
            (saved.includes('employee_create') ||
              saved.includes('employee_edit_all')) && (
              <EmployeeAdminScope employeeId={id} />
            )}
          <details className="access-card">
            <summary>سجل تغييرات الصلاحيات</summary>
            {auditError && (
              <p className="access-error">
                {auditError}{' '}
                <button onClick={() => setAuditReload((n) => n + 1)}>
                  أعد جلب السجل
                </button>
              </p>
            )}
            {!audit && !auditError && <p role="status">دا أجيب السجل…</p>}
            {audit?.length === 0 && (
              <p className="access-muted">ماكو تغييرات مسجّلة لحد هسه.</p>
            )}
            {audit?.map((item, index) => (
              <article className="access-audit" key={item.id ?? index}>
                <strong>
                  {item.action === 'ADD'
                    ? 'إضافة'
                    : item.action === 'REMOVE'
                      ? 'إزالة'
                      : item.action === 'UNPIN'
                        ? 'شيل تثبيت استثناء'
                        : 'انتهاء مؤقتة'}{' '}
                  · {label(item.permission)}
                </strong>
                <p>
                  {item.actorName ?? 'النظام'} · {time(item.createdAt)}
                </p>
                <small>{item.reason ?? 'بدون سبب مسجّل'}</small>
              </article>
            ))}
          </details>
        </>
      )}
      {review && (
        <Dialog
          title="راجع تغييرات الصلاحيات"
          busy={busy}
          cancel={() => setReview(false)}
        >
          <ChangeList title="راح تنضاف" names={add} label={label} />
          <ChangeList title="راح تنشال" names={remove} label={label} />
          <form
            onSubmit={(e) => {
              e.preventDefault()
              void save()
            }}
          >
            <label>
              سبب التغيير
              <textarea
                autoFocus
                required
                minLength={3}
                value={reason}
                onChange={(e) => setReason(e.target.value)}
              />
            </label>
            {error && (
              <p role="alert" className="access-error">
                {error}
              </p>
            )}
            {stale && (
              <button type="button" disabled={busy} onClick={refresh}>
                حدّث وراجع — يلغي المسوّدة
              </button>
            )}
            <div className="access-actions">
              <button
                className="access-primary"
                disabled={disabled || reason.trim().length < 3}
              >
                تأكيد وحفظ
              </button>
              <button
                type="button"
                disabled={busy}
                onClick={() => setReview(false)}
              >
                ارجع للتعديل
              </button>
            </div>
          </form>
        </Dialog>
      )}
      {exception && (
        <Dialog
          title={
            exception.removing
              ? 'إزالة تثبيت الاستثناء؟'
              : `استثناء: ${label(exception.name)}`
          }
          busy={busy}
          cancel={() => setException(null)}
        >
          <form
            onSubmit={(e) => {
              e.preventDefault()
              void saveException()
            }}
          >
            {exception.removing ? (
              <p>
                نشيل صفة الاستثناء فقط. الصلاحية تبقى ممنوحة، وممكن المطابقة
                تشيلها لاحقاً إذا مو من قالب المنصب.
              </p>
            ) : (
              <>
                <p>
                  التأكيد يمنح الصلاحية ويثبتها استثناء معتمد؛ مطابقة الصنف ما
                  تشيلها.
                </p>
                <label>
                  سبب الاستثناء
                  <textarea
                    required
                    minLength={3}
                    value={reason}
                    onChange={(e) => setReason(e.target.value)}
                  />
                </label>
                <label>
                  تنتهي بموعد — اختياري (توقيت بغداد)
                  <input
                    type="datetime-local"
                    value={expires}
                    onChange={(e) => setExpires(e.target.value)}
                  />
                </label>
                <p>
                  بدون موعد: استثناء مستمر. بموعد: الخادم يشيل الصلاحية كلها من
                  ينتهي وقتها.
                </p>
              </>
            )}
            {error && (
              <p role="alert" className="access-error">
                {error}
              </p>
            )}
            {stale && (
              <button type="button" disabled={busy} onClick={refresh}>
                حدّث وراجع — يلغي المسوّدة
              </button>
            )}
            <div className="access-actions">
              <button
                className="access-primary"
                disabled={
                  disabled || (!exception.removing && reason.trim().length < 3)
                }
              >
                {exception.removing ? 'أزل تثبيت الاستثناء' : 'تأكيد الاستثناء'}
              </button>
              <button
                type="button"
                disabled={busy}
                onClick={() => setException(null)}
              >
                إلغاء
              </button>
            </div>
          </form>
        </Dialog>
      )}
    </>
  )
}

function ChangeList({
  title,
  names,
  label,
}: {
  title: string
  names: string[]
  label: (name: string) => string
}) {
  return (
    <section className="access-change">
      <h4>
        {title} · {names.length}
      </h4>
      {names.length ? (
        <ul>
          {names.map((name) => (
            <li key={name}>{label(name)}</li>
          ))}
        </ul>
      ) : (
        <p>ماكو</p>
      )}
    </section>
  )
}

function DriftPage({
  groups,
  label,
  onWorking,
}: {
  groups: AccessGroup[]
  label: (name: string) => string
  onWorking: (value: boolean) => void
}) {
  const [rows, setRows] = useState<RoleDrift[] | null>(null)
  const [reload, setReload] = useState(0)
  const [error, setError] = useState('')
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState('ALL')
  const [selected, setSelected] = useState<RoleDrift | null>(null)
  const [reason, setReason] = useState('')
  const [busy, setBusy] = useState(false)
  const [stale, setStale] = useState(false)
  const [success, setSuccess] = useState('')
  useEffect(() => {
    let alive = true
    accessApi
      .drift()
      .then((value) => {
        if (alive) {
          setRows(value)
          setError('')
          setStale(false)
        }
      })
      .catch((e) => {
        if (alive) setError(message(e))
      })
    return () => {
      alive = false
    }
  }, [reload])
  useEffect(() => {
    onWorking(busy)
    return () => onWorking(false)
  }, [busy, onWorking])
  const grantable = (name: string) =>
    groups.some((g) => g.perms.some((p) => p.name === name && p.grantable))
  const align = async () => {
    if (
      !selected ||
      busy ||
      stale ||
      reason.trim().length < 3 ||
      selected.status !== 'DIFF'
    )
      return
    setBusy(true)
    setError('')
    try {
      const value = await accessApi.align(selected.employeeId, {
        reason: reason.trim(),
        version: selected.version,
      })
      setRows(
        (old) =>
          old?.map((r) => (r.employeeId === value.employeeId ? value : r)) ??
          null,
      )
      setSelected(null)
      setSuccess(
        value.status === 'MATCH'
          ? 'تمت المطابقة، والاستثناءات باقية.'
          : 'انطبقت التغييرات المسموحة. بقيت فروقات مقفولة؛ راجعها بالكتالوگ.',
      )
    } catch (e) {
      setError(message(e))
      if (conflict(e)) setStale(true)
    } finally {
      setBusy(false)
    }
  }
  return (
    <section className="access-drift">
      <div className="access-stats">
        {(['MATCH', 'DIFF', 'NO_TEMPLATE'] as const).map((key) => (
          <button
            key={key}
            aria-pressed={status === key}
            onClick={() => setStatus(status === key ? 'ALL' : key)}
          >
            <strong>
              {rows ? rows.filter((r) => r.status === key).length : '—'}
            </strong>
            {key === 'MATCH'
              ? 'مطابقون'
              : key === 'DIFF'
                ? 'عندهم اختلاف'
                : 'بلا قالب'}
          </button>
        ))}
      </div>
      <p className="access-notice">
        المطابقة تخص الصلاحيات فقط، ما تغيّر منصب الموظف ولا دوره. الاستثناءات
        المعتمدة والمقفول يبقون.
      </p>
      <label>
        ابحث عن موظف
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="الاسم أو الصنف…"
        />
      </label>
      {error && (
        <p className="access-error" role="alert">
          {error}
        </p>
      )}
      <button
        disabled={busy}
        onClick={() => {
          setRows(null)
          setSelected(null)
          setReload((n) => n + 1)
        }}
      >
        تحديث المقارنة
      </button>
      {success && (
        <p className="access-success" role="status">
          {success}
        </p>
      )}
      {!rows && !error && <p role="status">دا أقارن قوالب المناصب…</p>}
      {rows?.length === 0 && <p>ماكو موظفين للمقارنة.</p>}
      {rows
        ?.filter(
          (r) =>
            (status === 'ALL' || r.status === status) &&
            matches([r.name, roleLabel(r.role)], search),
        )
        .map((item) => (
          <article
            className="access-card access-drift-row"
            key={item.employeeId}
          >
            <div>
              <h3>{item.name}</h3>
              <p>{roleLabel(item.role)}</p>
              <span className="access-badge">
                {item.status === 'MATCH'
                  ? '✓ مطابق'
                  : item.status === 'NO_TEMPLATE'
                    ? 'يحتاج تحديد صنفه — بلا قالب'
                    : `${item.extra.length} زايدة · ${item.missing.length} ناقصة`}
              </span>
              <small>{item.exceptions.length} استثناءات معتمدة</small>
            </div>
            <button
              disabled={busy}
              onClick={() => {
                setSelected(item)
                setReason('')
                setError('')
                setStale(false)
              }}
            >
              {item.status === 'DIFF'
                ? 'شوف الفرق وطابقه ويا صنفه'
                : 'شوف التفاصيل'}
            </button>
          </article>
        ))}
      {selected && (
        <Dialog
          title={`مطابقة الصنف: ${selected.name}`}
          busy={busy}
          cancel={() => setSelected(null)}
        >
          <p>
            {roleLabel(selected.role)} ·{' '}
            {selected.status === 'NO_TEMPLATE'
              ? 'هذا الموظف ما إله قالب؛ حدد صنفه أول.'
              : 'هاي مقارنة قبل التنفيذ؛ ما تغير شي لحد تأكيدك.'}
          </p>
          <ChangeList
            title="راح تنضاف"
            names={selected.missing.filter(grantable)}
            label={label}
          />
          <ChangeList
            title="راح تنشال"
            names={selected.extra.filter(grantable)}
            label={label}
          />
          <ChangeList
            title="استثناءات باقية"
            names={selected.exceptions}
            label={label}
          />
          <ChangeList
            title="مقفولة — ما تتغيّر"
            names={[...selected.missing, ...selected.extra].filter(
              (n) => !grantable(n),
            )}
            label={label}
          />
          {error && (
            <p className="access-error" role="alert">
              {error}
            </p>
          )}
          {stale && (
            <button
              onClick={() => {
                setRows(null)
                setSelected(null)
                setReload((n) => n + 1)
              }}
            >
              حدّث المقارنة وراجع الفرق
            </button>
          )}
          {selected.status === 'DIFF' && (
            <form
              onSubmit={(e) => {
                e.preventDefault()
                void align()
              }}
            >
              <label>
                سبب المطابقة
                <textarea
                  required
                  minLength={3}
                  value={reason}
                  onChange={(e) => setReason(e.target.value)}
                />
              </label>
              <div className="access-actions">
                <button
                  className="access-primary"
                  disabled={
                    busy ||
                    stale ||
                    reason.trim().length < 3 ||
                    ![...selected.extra, ...selected.missing].some(grantable)
                  }
                >
                  تأكيد المطابقة
                </button>
                <button
                  type="button"
                  disabled={busy}
                  onClick={() => setSelected(null)}
                >
                  إلغاء
                </button>
              </div>
            </form>
          )}
        </Dialog>
      )}
    </section>
  )
}
