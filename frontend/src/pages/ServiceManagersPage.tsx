import { useEffect, useState } from 'react'
import { api, type Employee, type Service, type ServiceManager } from '../api'

// مسؤول خدمة عام — تعميم فكرة "أبو الجي بي اس" لأي مجموعة خدمات: موظف واحد
// يمكن يكون مسؤول عن أكثر من خدمة سوا (مثال: GPS + منظومات الصوت + الحريق).
export default function ServiceManagersPage() {
  const [managers, setManagers] = useState<ServiceManager[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [services, setServices] = useState<Service[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)

  const [selectedEmployeeId, setSelectedEmployeeId] = useState('')
  const [selectedServiceIds, setSelectedServiceIds] = useState<string[]>([])

  const load = () => {
    setLoading(true)
    Promise.all([api.getServiceManagers(), api.getEmployees(), api.getServices()])
      .then(([m, e, s]) => { setManagers(m); setEmployees(e); setServices(s) })
      .finally(() => setLoading(false))
  }
  useEffect(load, [])

  const grouped = employees
    .map(emp => ({ employee: emp, services: managers.filter(m => m.employee?.id === emp.id).map(m => m.service).filter(Boolean) as Service[] }))
    .filter(g => g.services.length > 0)

  const openEditor = (employeeId: string) => {
    setSelectedEmployeeId(employeeId)
    setSelectedServiceIds(managers.filter(m => m.employee?.id === employeeId).map(m => m.service?.id || ''))
  }

  const toggleService = (id: string) =>
    setSelectedServiceIds(prev => prev.includes(id) ? prev.filter(x => x !== id) : [...prev, id])

  const save = async () => {
    if (!selectedEmployeeId) return alert('اختر الموظف أول')
    setSaving(true)
    try {
      await api.setServiceManagers(selectedEmployeeId, selectedServiceIds)
      setSelectedEmployeeId(''); setSelectedServiceIds([])
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر الحفظ')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <h2 className="mb-2 text-2xl font-bold text-brand-900">مسؤولو الخدمات 🛠️</h2>
      <p className="mb-6 text-sm text-slate-500">
        امنح أي موظف مسؤولية خدمة أو مجموعة خدمات مع بعض (مثال: GPS + منظومات الصوت + الحريق) — يصير هو المسؤول الوحيد عن تفعيلها وجدولتها.
      </p>

      {loading && <p className="py-16 text-center text-slate-400">جاري التحميل...</p>}

      {!loading && (
        <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
          {/* Current assignments */}
          <div className="rounded-2xl bg-white p-6 shadow-sm">
            <h3 className="mb-4 text-lg font-bold text-brand-900">المسؤوليات الحالية</h3>
            {grouped.length === 0 && <p className="text-sm text-slate-400">ماكو مسؤولي خدمات بعد</p>}
            <div className="flex flex-col gap-3">
              {grouped.map(g => (
                <div key={g.employee.id} className="rounded-xl border border-slate-100 p-4">
                  <div className="mb-2 flex items-center justify-between">
                    <span className="font-bold text-slate-800">{g.employee.name}</span>
                    <button onClick={() => openEditor(g.employee.id)} className="text-xs font-bold text-brand-600 hover:underline">تعديل</button>
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {g.services.map(s => (
                      <span key={s.id} className="rounded-full bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700">{s.name}</span>
                    ))}
                  </div>
                </div>
              ))}
            </div>
          </div>

          {/* Editor */}
          <div className="rounded-2xl bg-white p-6 shadow-sm">
            <h3 className="mb-4 text-lg font-bold text-brand-900">تعيين / تعديل مسؤولية</h3>
            <label className="mb-1 block text-xs font-bold text-slate-500">الموظف</label>
            <select value={selectedEmployeeId} onChange={e => openEditor(e.target.value)}
              className="mb-4 w-full rounded-xl border border-slate-200 px-3 py-2.5 text-sm outline-none focus:border-brand-500">
              <option value="">— اختر موظف —</option>
              {employees.map(e => <option key={e.id} value={e.id}>{e.name} ({e.role})</option>)}
            </select>

            <label className="mb-1 block text-xs font-bold text-slate-500">الخدمات المسؤول عنها</label>
            <div className="grid max-h-64 grid-cols-2 gap-2 overflow-y-auto">
              {services.map(s => (
                <label key={s.id} className={`flex cursor-pointer items-center gap-2 rounded-xl border px-3 py-2 text-sm ${
                  selectedServiceIds.includes(s.id) ? 'border-brand-500 bg-brand-50 font-bold text-brand-700' : 'border-slate-200 text-slate-600'
                }`}>
                  <input type="checkbox" checked={selectedServiceIds.includes(s.id)} onChange={() => toggleService(s.id)} />
                  {s.name}
                </label>
              ))}
            </div>

            <button onClick={save} disabled={saving || !selectedEmployeeId}
              className="mt-5 w-full rounded-xl bg-brand-500 py-3 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
              {saving ? 'جاري الحفظ...' : 'حفظ المسؤولية'}
            </button>
          </div>
        </div>
      )}
    </div>
  )
}
