import { useEffect, useState } from 'react'
import { api, type Employee, type Service, type TrainingMaterial } from '../api'

export default function TrainingManagement() {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [services, setServices] = useState<Service[]>([])
  const [selectedEmployeeId, setSelectedEmployeeId] = useState<string>('')
  const [assignedServiceIds, setAssignedServiceIds] = useState<string[]>([])
  const [selectedServiceId, setSelectedServiceId] = useState<string>('')
  const [materials, setMaterials] = useState<TrainingMaterial[]>([])
  const [newTitle, setNewTitle] = useState('')
  const [newUrl, setNewUrl] = useState('')
  const [newType, setNewType] = useState<'VIDEO' | 'DOCUMENT'>('VIDEO')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    api.getEmployees().then(setEmployees)
    api.getServices().then(setServices)
  }, [])

  useEffect(() => {
    // Guard-clause reset when selection is cleared.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!selectedEmployeeId) { setAssignedServiceIds([]); return }
    api.getTrainingAssignments(selectedEmployeeId).then(list => setAssignedServiceIds(list.map(s => s.id)))
  }, [selectedEmployeeId])

  useEffect(() => {
    // Guard-clause reset when selection is cleared.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (!selectedServiceId) { setMaterials([]); return }
    api.getTrainingMaterials(selectedServiceId).then(setMaterials)
  }, [selectedServiceId])

  const selectedEmployee = employees.find(e => e.id === selectedEmployeeId)

  const toggleService = (serviceId: string) => {
    setAssignedServiceIds(prev =>
      prev.includes(serviceId) ? prev.filter(id => id !== serviceId) : [...prev, serviceId]
    )
  }

  const saveAssignments = async () => {
    if (!selectedEmployeeId) return
    setSaving(true)
    try {
      await api.setTrainingAssignments(selectedEmployeeId, assignedServiceIds)
    } finally {
      setSaving(false)
    }
  }

  const toggleTraineeMode = async (isTrainee: boolean) => {
    if (!selectedEmployeeId) return
    if (isTrainee) {
      await api.updateEmployee(selectedEmployeeId, { isTrainee: true })
    } else {
      // إنهاء التدريب يسجل تلقائياً تقييم إيجابي يعكس اجتيازه، بدل تعطيل الوضع بس
      await api.completeTraining(selectedEmployeeId)
    }
    setEmployees(prev => prev.map(e => e.id === selectedEmployeeId ? { ...e, isTrainee } : e))
  }

  const addMaterial = async () => {
    if (!selectedServiceId || !newTitle || !newUrl) return
    const material = await api.createTrainingMaterial({ serviceId: selectedServiceId, title: newTitle, url: newUrl, type: newType })
    setMaterials(prev => [...prev, material])
    setNewTitle('')
    setNewUrl('')
  }

  const removeMaterial = async (id: string) => {
    await api.deleteTrainingMaterial(id)
    setMaterials(prev => prev.filter(m => m.id !== id))
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">إدارة التدريب</h2>
      <p className="mt-1 text-slate-500">حدد صلاحية التدريب للموظفين الجدد والمواد التي يشاهدونها حسب تخصصهم.</p>

      <div className="mt-6 grid gap-6 lg:grid-cols-2">
        {/* Employee assignment */}
        <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
          <h3 className="mb-3 font-bold text-brand-900">1. تعيين موظف للتدريب</h3>
          <select value={selectedEmployeeId} onChange={e => setSelectedEmployeeId(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
            <option value="">اختر الموظف</option>
            {employees.map(e => <option key={e.id} value={e.id}>{e.name}</option>)}
          </select>

          {selectedEmployee && (
            <>
              <label className="mt-4 flex items-center gap-2 rounded-lg bg-amber-50 p-3 text-sm font-bold text-amber-700">
                <input type="checkbox" checked={selectedEmployee.isTrainee}
                  onChange={e => toggleTraineeMode(e.target.checked)} />
                وضع التدريب مفعّل (يشوف صفحة التدريب فقط)
              </label>

              <p className="mt-4 mb-2 text-xs font-bold text-slate-500">الخدمات المخصصة للتدريب:</p>
              <div className="flex flex-col gap-1.5">
                {services.map(s => (
                  <label key={s.id} className="flex items-center gap-2 rounded-lg border border-slate-200 px-3 py-2 text-sm">
                    <input type="checkbox" checked={assignedServiceIds.includes(s.id)}
                      onChange={() => toggleService(s.id)} />
                    {s.name}
                  </label>
                ))}
              </div>
              <button onClick={saveAssignments} disabled={saving}
                className="mt-3 w-full rounded-lg bg-brand-500 py-2 text-sm font-bold text-white hover:bg-brand-600 disabled:opacity-50">
                {saving ? 'جاري الحفظ...' : 'حفظ التعيين'}
              </button>
            </>
          )}
        </div>

        {/* Materials management */}
        <div className="rounded-2xl border border-slate-200 bg-white p-5 shadow-sm">
          <h3 className="mb-3 font-bold text-brand-900">2. إدارة المواد التدريبية</h3>
          <select value={selectedServiceId} onChange={e => setSelectedServiceId(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm">
            <option value="">اختر الخدمة</option>
            {services.map(s => <option key={s.id} value={s.id}>{s.name}</option>)}
          </select>

          {selectedServiceId && (
            <>
              <div className="mt-4 flex flex-col gap-2 rounded-lg bg-slate-50 p-3">
                <input value={newTitle} onChange={e => setNewTitle(e.target.value)} placeholder="عنوان المادة"
                  className="rounded-lg border border-slate-300 px-3 py-2 text-sm" />
                <input value={newUrl} onChange={e => setNewUrl(e.target.value)} placeholder="رابط الفيديو أو المستند"
                  className="rounded-lg border border-slate-300 px-3 py-2 text-sm" />
                <select value={newType} onChange={e => setNewType(e.target.value as 'VIDEO' | 'DOCUMENT')}
                  className="rounded-lg border border-slate-300 px-3 py-2 text-sm">
                  <option value="VIDEO">فيديو</option>
                  <option value="DOCUMENT">مستند</option>
                </select>
                <button onClick={addMaterial}
                  className="rounded-lg bg-emerald-500 py-2 text-sm font-bold text-white hover:bg-emerald-600">
                  + إضافة مادة
                </button>
              </div>

              <div className="mt-3 flex flex-col gap-2">
                {materials.map(m => (
                  <div key={m.id} className="flex items-center justify-between rounded-lg border border-slate-200 px-3 py-2 text-sm">
                    <button onClick={() => removeMaterial(m.id)} className="text-xs font-bold text-red-500 hover:underline">حذف</button>
                    <span className="font-bold text-brand-900">{m.title}</span>
                  </div>
                ))}
                {materials.length === 0 && <p className="text-center text-xs text-slate-400">لا توجد مواد بعد</p>}
              </div>
            </>
          )}
        </div>
      </div>
    </div>
  )
}
