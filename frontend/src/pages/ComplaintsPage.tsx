import { useEffect, useState } from 'react'
import { api, type Complaint, type ComplaintCustomerStat, type Customer, type Employee } from '../api'
import { useSession } from '../session'

const statusLabels: Record<Complaint['status'], string> = {
  NEW: 'جديدة',
  IN_PROGRESS: 'قيد المعالجة',
  RESOLVED: 'تم الحل',
  CLOSED: 'مغلقة',
}

const statusColors: Record<Complaint['status'], string> = {
  NEW: 'bg-yellow-100 text-yellow-800',
  IN_PROGRESS: 'bg-blue-100 text-blue-800',
  RESOLVED: 'bg-green-100 text-green-800',
  CLOSED: 'bg-gray-100 text-gray-800',
}

// أبو الجودة والمراقب المدقق (والأدمن) يشوفون متابعة/إدارة الشكاوى بس —
// باقي الأدوار الي عندها صلاحية الشكاوى (مثلاً المبيعات) تشوف تسجيل شكوى
// جديدة بس، بدون واجهة المتابعة والإدارة.
const trackingRoles = ['QUALITY_ENGINEER', 'MONITOR', 'ADMIN']
// مهندس الجودة (والأدمن) يشوف تفاصيل الزبون كاملة ويتواصل معه مباشرة.
// المراقب المدقق يشوف بس "إنذار" إنه اكو شكوى تحتاج متابعة، بدون تفاصيل
// الزبون — دوره رقابي بس، مو التواصل المباشر.
const fullDetailRoles = ['QUALITY_ENGINEER', 'ADMIN']

export default function ComplaintsPage() {
  const { employee: currentUser } = useSession()
  const canTrack = !!currentUser && trackingRoles.includes(currentUser.role)
  const canSeeFullDetail = !!currentUser && fullDetailRoles.includes(currentUser.role)
  const [complaints, setComplaints] = useState<Complaint[]>([])
  const [, setCustomers] = useState<Customer[]>([])
  const [employees, setEmployees] = useState<Employee[]>([])
  const [stats, setStats] = useState<ComplaintCustomerStat[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [showForm, setShowForm] = useState(false)

  // Form state
  const [phone, setPhone] = useState('')
  const [foundCustomer, setFoundCustomer] = useState<Customer | null>(null)
  const [searchingCustomer, setSearchingCustomer] = useState(false)
  const [description, setDescription] = useState('')
  const [submitting, setSubmitting] = useState(false)

  // Resolution modal
  const [resolvingId, setResolvingId] = useState<string | null>(null)
  const [resolution, setResolution] = useState('')

  const load = () => {
    setLoading(true)
    Promise.all([api.getComplaints(), api.getCustomers(), api.getEmployees()])
      .then(([comps, custs, emps]) => {
        setComplaints(comps)
        setCustomers(custs)
        setEmployees(emps)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])
  // تقرير عدد شكاوى كل زبون — منفصل تماماً عن إحصائيات الحجوزات
  useEffect(() => {
    if (canTrack) api.getComplaintStats().then(setStats).catch(() => {})
  }, [canTrack])
  // الأدوار الي بس تسجل شكوى (بدون متابعة) تشوف الفورم مباشرة بدون زر تبديل
  useEffect(() => {
    if (currentUser && !canTrack) setShowForm(true)
  }, [currentUser, canTrack])

  const handleLookup = async () => {
    if (!phone.trim()) return
    setSearchingCustomer(true)
    try {
      const c = await api.lookupCustomer(phone.trim())
      setFoundCustomer(c)
      if (!c) alert('لم يتم العثور على زبون بهذا الرقم')
    } catch {
      setFoundCustomer(null)
    } finally {
      setSearchingCustomer(false)
    }
  }

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!foundCustomer || !currentUser) return
    setSubmitting(true)
    try {
      await api.createComplaint({
        customerId: foundCustomer.id,
        description,
        createdByEmployeeId: currentUser.id,
      })
      setPhone('')
      setFoundCustomer(null)
      setDescription('')
      setShowForm(false)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const handleAssign = async (complaintId: string, employeeId: string) => {
    try {
      await api.updateComplaint(complaintId, { assignedToEmployeeId: employeeId })
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  const handleStatusChange = async (complaintId: string, status: Complaint['status']) => {
    try {
      await api.updateComplaint(complaintId, { status })
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  const handleResolve = async () => {
    if (!resolvingId || !resolution.trim()) return
    try {
      await api.resolveComplaint(resolvingId, resolution)
      setResolvingId(null)
      setResolution('')
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  return (
    <div>
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">شكاوى الصيانة</h2>
          <p className="mt-1 text-slate-500">إدارة شكاوى العملاء ومتابعة حالتها.</p>
        </div>
        {!canTrack && (
          <button
            onClick={() => setShowForm(!showForm)}
            className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30"
          >
            شكوى جديدة
          </button>
        )}
      </div>

      {/* New complaint form */}
      {!canTrack && showForm && (
        <form
          onSubmit={handleSubmit}
          className="mt-6 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
        >
          <h3 className="mb-4 text-lg font-bold text-brand-800">تسجيل شكوى جديدة</h3>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">رقم هاتف الزبون</label>
              <div className="flex gap-2">
                <input
                  value={phone}
                  onChange={(e) => setPhone(e.target.value)}
                  placeholder="أدخل رقم الهاتف"
                  className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
                />
                <button
                  type="button"
                  onClick={handleLookup}
                  disabled={searchingCustomer}
                  className="whitespace-nowrap rounded-lg bg-brand-100 px-4 py-3 text-sm font-medium text-brand-700 transition-colors hover:bg-brand-200 disabled:opacity-50"
                >
                  {searchingCustomer ? 'جاري البحث...' : 'بحث'}
                </button>
              </div>
              {foundCustomer && (
                <p className="mt-2 rounded-lg bg-green-50 px-3 py-2 text-sm text-green-700">
                  الزبون: {foundCustomer.name} - {foundCustomer.phone}
                </p>
              )}
            </div>
            <div>
              <label className="mb-1 block text-sm font-medium text-slate-600">وصف الشكوى</label>
              <textarea
                required
                value={description}
                onChange={(e) => setDescription(e.target.value)}
                rows={3}
                className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
              />
            </div>
          </div>
          <div className="mt-4">
            <button
              type="submit"
              disabled={submitting || !foundCustomer}
              className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
            >
              {submitting ? 'جاري الحفظ...' : 'تسجيل الشكوى'}
            </button>
          </div>
        </form>
      )}

      {/* Resolution modal */}
      {canTrack && resolvingId && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="w-full max-w-md rounded-xl bg-white p-6 shadow-xl">
            <h3 className="mb-4 text-lg font-bold text-brand-800">حل الشكوى</h3>
            <textarea
              value={resolution}
              onChange={(e) => setResolution(e.target.value)}
              rows={4}
              placeholder="أدخل تفاصيل الحل..."
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
            />
            <div className="mt-4 flex gap-3">
              <button
                onClick={handleResolve}
                className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-2 font-medium text-white shadow-md"
              >
                حفظ الحل
              </button>
              <button
                onClick={() => { setResolvingId(null); setResolution('') }}
                className="rounded-lg border border-slate-300 px-6 py-2 font-medium text-slate-600"
              >
                إلغاء
              </button>
            </div>
          </div>
        </div>
      )}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {canTrack && stats.length > 0 && (
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="border-b border-slate-100 px-4 py-3 text-sm font-bold text-brand-800">
            تقرير الشكاوى حسب الزبون
          </h3>
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 text-right text-sm">
              <thead className="bg-slate-50">
                <tr>
                  <th className="px-4 py-2 font-semibold text-slate-600">الزبون</th>
                  <th className="px-4 py-2 font-semibold text-slate-600">الهاتف</th>
                  <th className="px-4 py-2 font-semibold text-slate-600">عدد الشكاوى</th>
                  <th className="px-4 py-2 font-semibold text-slate-600">مفتوحة حالياً</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {stats.map((s) => (
                  <tr key={s.customerId}>
                    <td className="px-4 py-2 font-medium">{s.customerName}</td>
                    <td className="px-4 py-2 text-slate-500">{s.customerPhone}</td>
                    <td className="px-4 py-2 font-bold text-brand-700">{s.complaintCount}</td>
                    <td className="px-4 py-2">{s.openCount > 0 ? `${s.openCount} ⚠️` : '-'}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {canTrack && !loading && !error && !canSeeFullDetail && (
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 text-right">
              <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                <tr>
                  <th className="px-4 py-3 text-sm font-semibold">تنبيه</th>
                  <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                  <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {complaints.map((c) => (
                  <tr key={c.id} className="transition-colors hover:bg-slate-50">
                    <td className="px-4 py-3 font-medium text-amber-700">
                      ⚠️ شكوى تحتاج متابعة من مهندس الجودة
                    </td>
                    <td className="px-4 py-3">
                      <span className={`rounded-full px-3 py-1 text-xs font-bold ${statusColors[c.status]}`}>
                        {statusLabels[c.status]}
                      </span>
                    </td>
                    <td className="px-4 py-3 text-slate-500">
                      {new Date(c.createdAt).toLocaleDateString('ar-IQ')}
                    </td>
                  </tr>
                ))}
                {complaints.length === 0 && (
                  <tr>
                    <td colSpan={3} className="px-4 py-6 text-center text-slate-400">
                      لا توجد شكاوى بعد
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {canSeeFullDetail && !loading && !error && (
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="overflow-x-auto">
            <table className="min-w-full divide-y divide-gray-200 text-right">
              <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                <tr>
                  <th className="px-4 py-3 text-sm font-semibold">الزبون</th>
                  <th className="px-4 py-3 text-sm font-semibold">الوصف</th>
                  <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                  <th className="px-4 py-3 text-sm font-semibold">المسؤول</th>
                  <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
                  <th className="px-4 py-3 text-sm font-semibold">الحل</th>
                  <th className="px-4 py-3 text-sm font-semibold">إجراءات</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-100">
                {complaints.map((c) => (
                  <tr key={c.id} className="transition-colors hover:bg-slate-50">
                    <td className="px-4 py-3 font-medium">
                      {c.customer.name}
                      <div className="text-xs font-normal text-slate-400">{c.customer.phone}</div>
                    </td>
                    <td className="max-w-xs truncate px-4 py-3 text-slate-600">{c.description}</td>
                    <td className="px-4 py-3">
                      <span className={`rounded-full px-3 py-1 text-xs font-bold ${statusColors[c.status]}`}>
                        {statusLabels[c.status]}
                      </span>
                    </td>
                    <td className="px-4 py-3">
                      <select
                        value={c.assignedToEmployeeId || ''}
                        onChange={(e) => handleAssign(c.id, e.target.value)}
                        className="rounded-lg border border-slate-300 px-2 py-1 text-sm outline-none focus:border-brand-500"
                      >
                        <option value="">غير محدد</option>
                        {employees.map((emp) => (
                          <option key={emp.id} value={emp.id}>{emp.name}</option>
                        ))}
                      </select>
                    </td>
                    <td className="px-4 py-3 text-slate-500">
                      {new Date(c.createdAt).toLocaleDateString('ar-IQ')}
                    </td>
                    <td className="max-w-xs truncate px-4 py-3 text-slate-500">{c.resolution || '-'}</td>
                    <td className="px-4 py-3">
                      <div className="flex gap-2">
                        {c.status === 'NEW' && (
                          <button
                            onClick={() => handleStatusChange(c.id, 'IN_PROGRESS')}
                            className="rounded-lg bg-blue-100 px-3 py-1 text-xs font-medium text-blue-700 hover:bg-blue-200"
                          >
                            بدء المعالجة
                          </button>
                        )}
                        {(c.status === 'NEW' || c.status === 'IN_PROGRESS') && (
                          <button
                            onClick={() => setResolvingId(c.id)}
                            className="rounded-lg bg-green-100 px-3 py-1 text-xs font-medium text-green-700 hover:bg-green-200"
                          >
                            حل
                          </button>
                        )}
                        {c.status === 'RESOLVED' && (
                          <button
                            onClick={() => handleStatusChange(c.id, 'CLOSED')}
                            className="rounded-lg bg-gray-100 px-3 py-1 text-xs font-medium text-gray-700 hover:bg-gray-200"
                          >
                            إغلاق
                          </button>
                        )}
                      </div>
                    </td>
                  </tr>
                ))}
                {complaints.length === 0 && (
                  <tr>
                    <td colSpan={7} className="px-4 py-6 text-center text-slate-400">
                      لا توجد شكاوى بعد
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}
    </div>
  )
}
