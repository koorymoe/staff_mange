import { useEffect, useState } from 'react'
import { accessApi, type AccessGroup, type RoleDrift } from '../../permissionAccess'
import { roleLabel } from '../../roleLabels'
import { matches } from '../../utils/search'
import { conflict, message } from './shared'
import { Dialog } from './ui'
import { ChangeList } from './AccessEditor'

export function DriftPage({
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
