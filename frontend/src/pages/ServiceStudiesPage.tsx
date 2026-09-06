import { useEffect, useState } from 'react'
import { api, type ServiceStudy, type Employee } from '../api'
import { useSession } from '../session'

function splitList(v: string) {
  return v.split(',').map((s) => s.trim()).filter(Boolean)
}

export default function ServiceStudiesPage() {
  const { permissions, employee } = useSession()
  const canAdd = employee?.role === 'ADMIN' || permissions.includes('unit_technicians')
  const isAdmin = employee?.role === 'ADMIN'
  const employeeId = employee?.id

  const [items, setItems] = useState<ServiceStudy[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [newName, setNewName] = useState('')
  const [saving, setSaving] = useState(false)
  const [busyId, setBusyId] = useState<string | null>(null)
  const [reportDraft, setReportDraft] = useState<Record<string, string>>({})

  const load = () => { api.getServiceStudies().then(setItems).finally(() => setLoading(false)) }
  useEffect(load, [])
  useEffect(() => { if (isAdmin) api.getEmployees().then(setEmployees) }, [isAdmin])

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!newName.trim()) return
    setSaving(true)
    try {
      await api.createServiceStudy(newName.trim())
      setNewName('')
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر إضافة الخدمة')
    } finally {
      setSaving(false)
    }
  }

  const handleAssign = async (id: string, current: string[]) => {
    const options = employees.map((e) => `${e.name} :: ${e.id}`).join('\n')
    const picked = prompt(`اكتب أسماء التقنيين الموكَّلين مفصولين بفاصلة (اختر من):\n${options}`, current.join(', '))
    if (picked === null) return
    const names = splitList(picked)
    const ids = employees.filter((e) => names.includes(e.name)).map((e) => e.id)
    setBusyId(id)
    try {
      await api.assignServiceStudy(id, ids)
      load()
    } finally {
      setBusyId(null)
    }
  }

  const handleAddReport = async (id: string) => {
    const content = reportDraft[id]?.trim()
    if (!content) return
    setBusyId(id)
    try {
      await api.addServiceStudyReport(id, content)
      setReportDraft((d) => ({ ...d, [id]: '' }))
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر رفع التقرير')
    } finally {
      setBusyId(null)
    }
  }

  const handleArchive = async (id: string) => {
    setBusyId(id)
    try {
      await api.archiveServiceStudy(id)
      load()
    } finally {
      setBusyId(null)
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">إدارة الخدمات</h2>
      <p className="mt-1 text-slate-500">خدمة جديدة مقترحة تحتاج دراسة — المدير يوكّل تقنيين محددين، وكل موكَّل يرفع تقارير.</p>

      {canAdd && (
        <form onSubmit={handleCreate} className="mt-4 mb-4 flex gap-2 rounded-xl border border-white bg-white p-4 shadow-sm">
          <input value={newName} onChange={(e) => setNewName(e.target.value)} placeholder="اسم الخدمة المراد فتحها" className="flex-1 rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500" />
          <button type="submit" disabled={saving} className="rounded-lg bg-brand-500 px-4 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
            {saving ? 'جاري...' : '+ إضافة'}
          </button>
        </form>
      )}

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && items.length === 0 && <p className="text-slate-400">لا توجد خدمات مقترحة بعد.</p>}

      <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
        {items.map((s) => {
          const isAssignedToMe = !!employeeId && s.assignedEmployees.some((e) => e.id === employeeId)
          return (
            <div key={s.id} className={`rounded-xl border p-4 shadow-sm ${s.archived ? 'border-slate-200 bg-slate-50' : 'border-white bg-white'}`}>
              <div className="flex items-start justify-between">
                <p className="font-bold text-brand-900">{s.name} {s.archived && <span className="text-xs font-normal text-slate-400">(مؤرشفة)</span>}</p>
                {isAdmin && !s.archived && (
                  <button onClick={() => handleArchive(s.id)} disabled={busyId === s.id} className="rounded-lg bg-slate-100 px-2 py-1 text-xs font-bold text-slate-600 hover:bg-slate-200">🗄 أرشفة</button>
                )}
              </div>
              <div className="mt-2 flex items-center gap-2 text-xs">
                <b className="text-slate-600">الموكَّلون:</b>
                <span className="text-slate-500">{s.assignedEmployees.length > 0 ? s.assignedEmployees.map((e) => e.name).join('، ') : 'لا يوجد بعد'}</span>
                {isAdmin && (
                  <button onClick={() => handleAssign(s.id, s.assignedEmployees.map((e) => e.name))} disabled={busyId === s.id} className="rounded bg-brand-50 px-2 py-0.5 font-bold text-brand-700 hover:bg-brand-100">تعديل</button>
                )}
              </div>

              {s.reports.length > 0 && (
                <div className="mt-3 flex flex-col gap-2">
                  {s.reports.map((r) => (
                    <div key={r.id} className="rounded-lg bg-slate-50 p-2 text-xs">
                      <b className="text-slate-700">{r.employee?.name}</b> — <span className="text-slate-400">{new Date(r.createdAt).toLocaleDateString('ar-IQ')}</span>
                      <p className="mt-1 whitespace-pre-wrap text-slate-600">{r.content}</p>
                    </div>
                  ))}
                </div>
              )}

              {isAssignedToMe && !s.archived && (
                <div className="mt-3">
                  <textarea
                    value={reportDraft[s.id] ?? ''}
                    onChange={(e) => setReportDraft((d) => ({ ...d, [s.id]: e.target.value }))}
                    placeholder="اكتب تقرير/دراسة عن هذه الخدمة..."
                    rows={2}
                    className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                  />
                  <button onClick={() => handleAddReport(s.id)} disabled={busyId === s.id} className="mt-1 rounded-lg bg-brand-500 px-3 py-1.5 text-xs font-bold text-white hover:bg-brand-600 disabled:opacity-50">
                    رفع التقرير
                  </button>
                </div>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
