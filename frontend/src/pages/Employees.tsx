import { useEffect, useState } from 'react'
import { api, type Employee, type Service } from '../api'

export default function Employees() {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [services, setServices] = useState<Service[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [selectedId, setSelectedId] = useState<string | null>(null)

  // New employee form
  const [name, setName] = useState('')
  const [certificate, setCertificate] = useState('')
  const [position, setPosition] = useState('')
  const [phone, setPhone] = useState('')
  const [submitting, setSubmitting] = useState(false)

  const load = () => {
    setLoading(true)
    Promise.all([api.getEmployees(), api.getServices()])
      .then(([emps, svcs]) => {
        setEmployees(emps)
        setServices(svcs)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleAddEmployee = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await api.createEmployee({
        name,
        certificate: certificate || null,
        position: position || null,
        phone: phone || null,
      })
      setName('')
      setCertificate('')
      setPosition('')
      setPhone('')
      load()
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const selectedEmployee = employees.find((emp) => emp.id === selectedId) || null

  const toggleSkill = async (employee: Employee, serviceId: string) => {
    const current = new Map(employee.skills.map((s) => [s.serviceId, s.canPerform]))
    const newValue = !current.get(serviceId)
    current.set(serviceId, newValue)

    const skills = services.map((svc) => ({
      serviceId: svc.id,
      canPerform: current.get(svc.id) ?? false,
    }))

    const updated = await api.updateEmployeeSkills(employee.id, skills)
    setEmployees((prev) => prev.map((emp) => (emp.id === employee.id ? updated : emp)))
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">إدارة الكوادر</h2>
      <p className="mt-1 text-slate-500">
        إدارة بيانات الموظفين وتحديد المهارات (الخدمات) التي يستطيع كل موظف تنفيذها.
      </p>

      <form
        onSubmit={handleAddEmployee}
        className="mt-6 grid grid-cols-1 gap-4 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] sm:grid-cols-4"
      >
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">الاسم</label>
          <input
            required
            value={name}
            onChange={(e) => setName(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">الشهادة</label>
          <input
            value={certificate}
            onChange={(e) => setCertificate(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">المنصب</label>
          <input
            value={position}
            onChange={(e) => setPosition(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div>
          <label className="mb-1 block text-sm font-medium text-slate-600">الهاتف</label>
          <input
            value={phone}
            onChange={(e) => setPhone(e.target.value)}
            className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
          />
        </div>
        <div className="sm:col-span-4">
          <button
            type="submit"
            disabled={submitting}
            className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
          >
            {submitting ? 'جاري الحفظ...' : 'إضافة موظف'}
          </button>
        </div>
      </form>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {!loading && !error && (
        <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)] lg:col-span-1">
            <table className="w-full text-right">
              <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                <tr>
                  <th className="px-4 py-3 text-sm font-semibold">الاسم</th>
                  <th className="px-4 py-3 text-sm font-semibold">المنصب</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {employees.map((emp) => (
                  <tr
                    key={emp.id}
                    onClick={() => setSelectedId(emp.id)}
                    className={`cursor-pointer transition-colors ${
                      selectedId === emp.id ? 'bg-brand-50' : 'hover:bg-slate-50'
                    }`}
                  >
                    <td className="px-4 py-3 font-medium">{emp.name}</td>
                    <td className="px-4 py-3 text-slate-500">{emp.position || '-'}</td>
                  </tr>
                ))}
                {employees.length === 0 && (
                  <tr>
                    <td colSpan={2} className="px-4 py-6 text-center text-slate-400">
                      لا يوجد موظفين بعد
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>

          <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] lg:col-span-2">
            {!selectedEmployee && (
              <p className="text-slate-400">اختر موظفاً من القائمة لعرض/تعديل مهاراته.</p>
            )}
            {selectedEmployee && (
              <div>
                <h3 className="text-lg font-bold text-brand-800">
                  مهارات: {selectedEmployee.name}
                </h3>
                <p className="mt-1 text-sm text-slate-500">
                  حدد الخدمات التي يستطيع هذا الموظف تنفيذها. يستخدم النظام هذه القائمة
                  لاقتراح الموظف تلقائياً عند إنشاء حجز جديد.
                </p>
                <div className="mt-4 grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
                  {services.map((service) => {
                    const skill = selectedEmployee.skills.find(
                      (s) => s.serviceId === service.id,
                    )
                    const canPerform = skill?.canPerform ?? false
                    return (
                      <label
                        key={service.id}
                        className={`flex cursor-pointer items-center gap-2 rounded-lg border px-3 py-2 text-sm transition-colors ${
                          canPerform
                            ? 'border-brand-500 bg-brand-50 text-brand-800'
                            : 'border-slate-200 text-slate-500'
                        }`}
                      >
                        <input
                          type="checkbox"
                          checked={canPerform}
                          onChange={() => toggleSkill(selectedEmployee, service.id)}
                          className="h-4 w-4 accent-brand-700"
                        />
                        {service.name}
                      </label>
                    )
                  })}
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
