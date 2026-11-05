import { useCallback, useEffect, useState } from 'react'
import { api, type Employee } from '../../api'
import { accessApi, type AccessAgent, type AccessGroup } from '../../permissionAccess'
import { roleLabel } from '../../roleLabels'
import { matches } from '../../utils/search'
import { defaults, message } from './shared'
import { Dialog } from './ui'
import { AccessEditor } from './AccessEditor'
import { DriftPage } from './DriftPage'

export function AccessPage({
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
