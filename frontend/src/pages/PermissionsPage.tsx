import { useEffect, useState } from 'react'
import { api, type Employee, type Permission } from '../api'

export default function PermissionsPage() {
  const [employees, setEmployees] = useState<Employee[]>([])
  const [permissions, setPermissions] = useState<Permission[]>([])
  const [selectedEmployeeId, setSelectedEmployeeId] = useState('')
  const [employeePerms, setEmployeePerms] = useState<string[]>([])
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [error, setError] = useState<string | null>(null)

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
      .then((perms) => setEmployeePerms(perms.map((p) => p.permissionName)))
      .catch(() => setEmployeePerms([]))
  }, [selectedEmployeeId])

  const togglePermission = (permName: string) => {
    setEmployeePerms((prev) =>
      prev.includes(permName) ? prev.filter((p) => p !== permName) : [...prev, permName],
    )
  }

  const handleSave = async () => {
    if (!selectedEmployeeId) return
    setSaving(true)
    try {
      await api.setEmployeePermissions(selectedEmployeeId, employeePerms)
      alert('تم حفظ الصلاحيات بنجاح')
    } catch (e) {
      alert(e instanceof Error ? e.message : 'حدث خطأ')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">إدارة الصلاحيات</h2>
      <p className="mt-1 text-slate-500">تحديد صلاحيات الوصول لكل موظف في الأنظمة المختلفة.</p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {!loading && !error && (
        <div className="mt-6 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="mb-6">
            <label className="mb-1 block text-sm font-medium text-slate-600">اختر الموظف</label>
            <select
              value={selectedEmployeeId}
              onChange={(e) => setSelectedEmployeeId(e.target.value)}
              className="w-full max-w-md rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500"
            >
              <option value="">-- اختر موظف --</option>
              {employees.map((emp) => (
                <option key={emp.id} value={emp.id}>
                  {emp.name}
                </option>
              ))}
            </select>
          </div>

          {selectedEmployeeId && (
            <>
              <h3 className="mb-4 text-lg font-bold text-brand-800">الصلاحيات</h3>
              <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                {permissions.map((perm) => (
                  <label
                    key={perm.id}
                    className={`flex cursor-pointer items-center gap-3 rounded-lg border px-4 py-3 text-sm transition-colors ${
                      employeePerms.includes(perm.name)
                        ? 'border-brand-500 bg-brand-50 text-brand-800'
                        : 'border-slate-200 text-slate-500'
                    }`}
                  >
                    <input
                      type="checkbox"
                      checked={employeePerms.includes(perm.name)}
                      onChange={() => togglePermission(perm.name)}
                      className="h-4 w-4 accent-brand-700"
                    />
                    {perm.label}
                  </label>
                ))}
              </div>

              <div className="mt-6">
                <button
                  onClick={handleSave}
                  disabled={saving}
                  className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-8 py-2.5 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
                >
                  {saving ? 'جاري الحفظ...' : 'حفظ الصلاحيات'}
                </button>
              </div>
            </>
          )}
        </div>
      )}
    </div>
  )
}
