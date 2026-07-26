import { useEffect, useState } from 'react'
import { api, type TeamInventoryToolCatalogItem, type TeamInventoryCheck, type TeamInventoryShortageReason, type Employee } from '../api'
import { useSession } from '../session'

const reasonLabels: Record<TeamInventoryShortageReason, string> = {
  FORGOTTEN: 'نسيان في مكان معين',
  DAMAGED: 'يجب جلب القطعة المتلوفة "تلف"',
  UNKNOWN: 'لا اعرف',
}

type PersonKey = 'LEADER' | 'EMPLOYEE1' | 'EMPLOYEE2'

interface ToolState {
  present: boolean
  reason: TeamInventoryShortageReason | ''
}

export default function TeamInventoryCheckPage() {
  const { employee } = useSession()
  const [tools, setTools] = useState<TeamInventoryToolCatalogItem[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [checks, setChecks] = useState<TeamInventoryCheck[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  const [employee1Id, setEmployee1Id] = useState('')
  const [employee2Id, setEmployee2Id] = useState('')
  // state[toolName][personKey] = { present, reason }
  const [state, setState] = useState<Record<string, Record<PersonKey, ToolState>>>({})

  const load = async () => {
    try {
      const [toolList, empList, checkList] = await Promise.all([
        api.getTeamInventoryTools(),
        api.getEmployees(),
        api.getTeamInventoryChecks(),
      ])
      setTools(toolList)
      setEmployees(empList)
      setChecks(checkList)
      setState((prev) => {
        const next: Record<string, Record<PersonKey, ToolState>> = { ...prev }
        for (const tool of toolList) {
          if (!next[tool.name]) {
            next[tool.name] = {
              LEADER: { present: true, reason: '' },
              EMPLOYEE1: { present: true, reason: '' },
              EMPLOYEE2: { present: true, reason: '' },
            }
          }
        }
        return next
      })
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setLoading(false)
    }
  }

  // eslint-disable-next-line react-hooks/set-state-in-effect
  useEffect(() => { load() }, [])

  const setToolState = (toolName: string, person: PersonKey, patch: Partial<ToolState>) => {
    setState((prev) => ({
      ...prev,
      [toolName]: {
        ...prev[toolName],
        [person]: { ...prev[toolName]?.[person], ...patch },
      },
    }))
  }

  const otherEmployees = employees.filter((e) => e.id !== employee?.id)

  const handleSubmit = async () => {
    const items: { toolName: string; personRole: PersonKey; present: boolean; reason?: TeamInventoryShortageReason | null }[] = []
    for (const tool of tools) {
      const roles: PersonKey[] = ['LEADER']
      if (employee1Id) roles.push('EMPLOYEE1')
      if (employee2Id) roles.push('EMPLOYEE2')
      for (const role of roles) {
        const s = state[tool.name]?.[role]
        if (!s) continue
        if (!s.present && !s.reason) {
          alert(`اختر سبب النقص لأداة "${tool.name}"`)
          return
        }
        items.push({ toolName: tool.name, personRole: role, present: s.present, reason: s.present ? null : (s.reason || null) })
      }
    }
    if (items.length === 0) {
      alert('لا توجد أدوات لتسجيل الجرد')
      return
    }
    setSubmitting(true)
    try {
      await api.createTeamInventoryCheck({
        employee1Id: employee1Id || null,
        employee2Id: employee2Id || null,
        items,
      })
      await load()
      alert('تم حفظ جرد الفريق بنجاح')
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر حفظ الجرد')
    } finally {
      setSubmitting(false)
    }
  }

  const columns: { key: PersonKey; label: string }[] = [
    { key: 'LEADER', label: 'الليدر' },
    ...(employee1Id ? [{ key: 'EMPLOYEE1' as PersonKey, label: 'الموظف الأول' }] : []),
    ...(employee2Id ? [{ key: 'EMPLOYEE2' as PersonKey, label: 'الموظف الثاني' }] : []),
  ]

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">جرد العدد (جرد الفريق)</h2>
      <p className="mt-1 text-slate-500">اختر أفراد الفريق ثم حدد حالة كل أداة لكل شخص</p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>}

      {!loading && !error && (
        <>
          <div className="mt-6 grid grid-cols-1 gap-4 rounded-xl bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">الموظف الأول</label>
              <select value={employee1Id} onChange={(e) => setEmployee1Id(e.target.value)}
                className="w-full rounded-lg border border-slate-200 px-3 py-2">
                <option value="">-- اختر --</option>
                {otherEmployees.map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
              </select>
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">الموظف الثاني</label>
              <select value={employee2Id} onChange={(e) => setEmployee2Id(e.target.value)}
                className="w-full rounded-lg border border-slate-200 px-3 py-2">
                <option value="">-- اختر --</option>
                {otherEmployees.filter((e) => e.id !== employee1Id).map((e) => <option key={e.id} value={e.id}>{e.name}</option>)}
              </select>
            </div>
          </div>

          <div className="mt-6 overflow-hidden rounded-xl bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <div className="overflow-x-auto">
              <table className="min-w-full divide-y divide-gray-200 text-right">
                <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                  <tr>
                    <th className="px-4 py-3 text-sm font-semibold">الأداة</th>
                    {columns.map((c) => (
                      <th key={c.key} className="px-4 py-3 text-sm font-semibold">{c.label}</th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-100">
                  {tools.map((tool) => (
                    <tr key={tool.id}>
                      <td className="px-4 py-3 font-medium">{tool.name}</td>
                      {columns.map((c) => {
                        const s = state[tool.name]?.[c.key] || { present: true, reason: '' }
                        return (
                          <td key={c.key} className="px-4 py-3">
                            <div className="flex flex-col gap-1">
                              <label className="flex items-center gap-2 text-sm">
                                <input type="checkbox" checked={s.present}
                                  onChange={(e) => setToolState(tool.name, c.key, { present: e.target.checked, reason: e.target.checked ? '' : s.reason })} />
                                متوفرة
                              </label>
                              {!s.present && (
                                <select value={s.reason} onChange={(e) => setToolState(tool.name, c.key, { reason: e.target.value as TeamInventoryShortageReason })}
                                  className="rounded border border-slate-200 px-2 py-1 text-xs">
                                  <option value="">-- سبب النقص --</option>
                                  {Object.entries(reasonLabels).map(([k, v]) => <option key={k} value={k}>{v}</option>)}
                                </select>
                              )}
                            </div>
                          </td>
                        )
                      })}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          <button onClick={handleSubmit} disabled={submitting}
            className="mt-6 w-full rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 py-3 font-bold text-white shadow-md disabled:opacity-50">
            {submitting ? 'جارٍ الحفظ...' : 'حفظ جرد الفريق'}
          </button>

          <h3 className="mt-10 text-lg font-bold text-brand-900">جلسات الجرد السابقة</h3>
          <div className="mt-3 space-y-3">
            {checks.length === 0 ? (
              <p className="text-slate-400">لا توجد جلسات سابقة</p>
            ) : (
              checks.map((c) => (
                <div key={c.id} className="rounded-xl bg-white p-4 shadow-[0_2px_10px_rgba(15,32,64,0.05)]">
                  <div className="text-sm text-slate-500">
                    الليدر: {c.leader?.name || '—'} · الموظف الأول: {c.employee1?.name || '—'} · الموظف الثاني: {c.employee2?.name || '—'}
                  </div>
                  <div className="mt-1 text-xs text-slate-400">{new Date(c.createdAt).toLocaleString('ar-IQ')}</div>
                  <div className="mt-2 text-sm">
                    {c.items.filter((i) => !i.present).length === 0
                      ? <span className="text-green-600 font-bold">جميع الأدوات متوفرة</span>
                      : <span className="text-amber-700 font-bold">{c.items.filter((i) => !i.present).length} أداة ناقصة</span>}
                  </div>
                </div>
              ))
            )}
          </div>
        </>
      )}
    </div>
  )
}
