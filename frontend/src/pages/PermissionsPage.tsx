import { useEffect, useState } from 'react'
import { api, type Employee, type Permission } from '../api'

export default function PermissionsPage() {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [selectedEmployeeId, setSelectedEmployeeId] = useState('')
  const [employeePerms, setEmployeePerms] = useState<string[]>([])
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [successMsg, setSuccessMsg] = useState<string | null>(null)

  useEffect(() => {
    setLoading(true)
    Promise.all([api.getEmployees(), api.getPermissions()])
      .then(([emps, perms]) => {
        setEmployees(emps)
        setPermissions(perms)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  useEffect(() => {
    if (!selectedEmployeeId) {
      setEmployeePerms([])
      return
    }
    api.getEmployeePermissions(selectedEmployeeId)
      .then((perms) => setEmployeePerms(perms.map((p) => p.name)))
      .catch(() => setEmployeePerms([]))
  }, [selectedEmployeeId])

  const togglePermission = (permName: string) => {
    setEmployeePerms((prev) =>
      prev.includes(permName) ? prev.filter((p) => p !== permName) : [...prev, permName],
    )
    setSuccessMsg(null)
  }

  const handleSave = async () => {
    if (!selectedEmployeeId) return
    setSaving(true)
    setSuccessMsg(null)
    setError(null)
    try {
      const permIds = permissions
        .filter((p) => employeePerms.includes(p.name))
        .map((p) => p.id)
      await api.setEmployeePermissions(selectedEmployeeId, permIds)
      setSuccessMsg('تم حفظ الصلاحيات بنجاح')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ أثناء الحفظ')
    } finally {
      setSaving(false)
    }
  }

  const filteredEmployees = employees.filter((emp) =>
    emp.name.toLowerCase().includes(search.toLowerCase()),
  )

  const selectedEmployee = employees.find((e) => e.id === selectedEmployeeId)

  return (
    <div dir="rtl">
      <h2 className="text-2xl font-bold text-brand-900">إدارة الصلاحيات</h2>
      <p className="mt-1 text-slate-500">تحديد صلاحيات الوصول لكل موظف في الأنظمة المختلفة.</p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-4 rounded-lg bg-red-50 p-4 text-red-600">{error}</p>
      )}

      {!loading && (
        <div className="mt-6 space-y-6">
          {/* Employee search and select */}
          <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <label className="mb-2 block text-sm font-medium text-slate-600">
              البحث عن موظف
            </label>
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="اكتب اسم الموظف..."
              className="mb-4 w-full max-w-md rounded-lg border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500"
            />

            <label className="mb-1 block text-sm font-medium text-slate-600">اختر الموظف</label>
            <select
              value={selectedEmployeeId}
              onChange={(e) => {
                setSelectedEmployeeId(e.target.value)
                setSuccessMsg(null)
                setError(null)
              }}
              className="w-full max-w-md rounded-lg border border-slate-300 px-3 py-2.5 outline-none focus:border-brand-500"
            >
              <option value="">-- اختر موظف --</option>
              {filteredEmployees.map((emp) => (
                <option key={emp.id} value={emp.id}>
                  {emp.name}
                </option>
              ))}
            </select>
          </div>

          {/* Permission toggles */}
          {selectedEmployeeId && (
            <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <h3 className="mb-2 text-lg font-bold text-brand-800">
                صلاحيات: {selectedEmployee?.name}
              </h3>
              <p className="mb-6 text-sm text-slate-400">
                فعّل أو عطّل الصلاحيات لكل نظام فرعي
              </p>

              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                {permissions.map((perm) => {
                  const isActive = employeePerms.includes(perm.name)
                  return (
                    <div
                      key={perm.id}
                      className={`flex items-center justify-between rounded-lg border px-5 py-4 transition-colors ${
                        isActive
                          ? 'border-brand-500 bg-brand-50'
                          : 'border-slate-200 bg-white'
                      }`}
                    >
                      <span className={`text-sm font-medium ${isActive ? 'text-brand-800' : 'text-slate-500'}`}>
                        {perm.label}
                      </span>
                      <button
                        type="button"
                        role="switch"
                        aria-checked={isActive}
                        onClick={() => togglePermission(perm.name)}
                        className={`relative inline-flex h-6 w-11 shrink-0 cursor-pointer rounded-full transition-colors ${
                          isActive ? 'bg-brand-600' : 'bg-slate-300'
                        }`}
                      >
                        <span
                          className={`pointer-events-none inline-block h-5 w-5 translate-y-0.5 transform rounded-full bg-white shadow ring-0 transition-transform ${
                            isActive ? '-translate-x-5' : '-translate-x-0.5'
                          }`}
                        />
                      </button>
                    </div>
                  )
                })}
              </div>

              {successMsg && (
                <p className="mt-4 rounded-lg bg-emerald-50 p-3 text-sm font-medium text-emerald-700">
                  {successMsg}
                </p>
              )}

              <div className="mt-6">
                <button
                  onClick={handleSave}
                  disabled={saving}
                  className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-10 py-2.5 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
                >
                  {saving ? 'جاري الحفظ...' : 'حفظ الصلاحيات'}
                </button>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
