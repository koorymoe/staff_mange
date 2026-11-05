import { useEffect, useLayoutEffect, useState } from 'react'
import { accessApi, type AccessAgent, type AccessGroup, type AccessState, type AccessAudit } from '../../permissionAccess'
import { roleLabel } from '../../roleLabels'
import { matches } from '../../utils/search'
import EmployeeAdminScope from '../../components/EmployeeAdminScope'
import { sourceLabel, tiers, exceptionDeadline, baghdadInput, agentPermission, time, conflict, message } from './shared'
import { Dialog } from './ui'

export function AccessEditor({
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

export function ChangeList({
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
