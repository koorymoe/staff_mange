import { useEffect, useState } from 'react'
import { api, type PersonalTool, type VehicleTool, type OnDemandTool, type ToolRequest, type Employee, type InventoryCheck } from '../api'
import { useSession } from '../session'

type TabKey = 'todaychecks' | 'personal' | 'vehicle' | 'ondemand' | 'requests'

const tabs: { key: TabKey; label: string }[] = [
  { key: 'todaychecks', label: 'نتائج جرد اليوم' },
  { key: 'personal', label: 'أدوات خاصة' },
  { key: 'vehicle', label: 'أدوات المركبات' },
  { key: 'ondemand', label: 'أدوات حسب الحاجة' },
  { key: 'requests', label: 'طلبات الأدوات' },
]

const requestStatusLabels: Record<ToolRequest['status'], string> = {
  PENDING: 'معلّق',
  APPROVED: 'موافق عليه',
  REJECTED: 'مرفوض',
  RETURNED: 'مسترجع',
}

const requestStatusColors: Record<ToolRequest['status'], string> = {
  PENDING: 'bg-yellow-100 text-yellow-800',
  APPROVED: 'bg-green-100 text-green-800',
  REJECTED: 'bg-red-100 text-red-800',
  RETURNED: 'bg-gray-100 text-gray-800',
}

export default function InventoryPage() {
  const { employee: currentUser } = useSession()
  const isAdmin = currentUser?.role === 'ADMIN'
  const canManageInventory = isAdmin || currentUser?.role === 'HR_COORDINATOR'
  // إداري الكميات مسؤول عن توفير وإضافة "أدوات حسب الحاجة" (نفس مسؤوليته
  // بطلبات المواد الناقصة من الموظفين) — بالإضافة للأدمن.
  const canManageOnDemand = isAdmin || currentUser?.role === 'PROCUREMENT_ADMIN'
  // موافقة/رفض طلبات الأدوات: الأدمن، إداري الكوادر، وكمان المراقب (على عكس
  // "جرد الأدوات" الي المراقب بيه اطلاع فقط — هذي طلبات موافقة منفصلة).
  const canApproveRequests = canManageInventory || currentUser?.role === 'MONITOR'
  const [resolvingId, setResolvingId] = useState<string | null>(null)
  const handleResolveCheck = async (id: string) => {
    setResolvingId(id)
    try {
      const updated = await api.resolveInventoryCheck(id)
      setTodaysChecks((prev) => prev.map((c) => (c.id === updated.id ? updated : c)))
    } catch { /* ignore */ }
    finally { setResolvingId(null) }
  }
  const [activeTab, setActiveTab] = useState<TabKey>('todaychecks')
  const [employees, setEmployees] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Today's inventory checks
  const [todaysChecks, setTodaysChecks] = useState<InventoryCheck[]>([])
  useEffect(() => { api.getTodaysInventoryChecks().then(setTodaysChecks).catch(() => setTodaysChecks([])) }, [])

  // Personal tools
  const [personalTools, setPersonalTools] = useState<PersonalTool[]>([])
  const [showPersonalForm, setShowPersonalForm] = useState(false)
  const [ptEmployeeId, setPtEmployeeId] = useState('')
  const [ptName, setPtName] = useState('')
  const [ptBarcode, setPtBarcode] = useState('')

  // Vehicle tools
  const [vehicleTools, setVehicleTools] = useState<VehicleTool[]>([])
  const [showVehicleForm, setShowVehicleForm] = useState(false)
  const [vtVehicleId, setVtVehicleId] = useState('')
  const [vtName, setVtName] = useState('')
  const [vtBarcode, setVtBarcode] = useState('')

  // On-demand tools
  const [onDemandTools, setOnDemandTools] = useState<OnDemandTool[]>([])
  const [showOnDemandForm, setShowOnDemandForm] = useState(false)
  const [odName, setOdName] = useState('')
  const [odBarcode, setOdBarcode] = useState('')
  const [odQuantity, setOdQuantity] = useState('')

  // Tool requests
  const [toolRequests, setToolRequests] = useState<ToolRequest[]>([])

  const [submitting, setSubmitting] = useState(false)

  const load = () => {
    setLoading(true)
    Promise.all([
      api.getPersonalTools(),
      api.getVehicleTools(),
      api.getOnDemandTools(),
      api.getToolRequests(),
      api.getEmployees(),
    ])
      .then(([pt, vt, od, tr, emps]) => {
        setPersonalTools(pt)
        setVehicleTools(vt)
        setOnDemandTools(od)
        setToolRequests(tr)
        setEmployees(emps)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleAddPersonalTool = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await api.createPersonalTool({ employeeId: ptEmployeeId, name: ptName, barcode: ptBarcode })
      setPtEmployeeId(''); setPtName(''); setPtBarcode('')
      setShowPersonalForm(false)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const handleAddVehicleTool = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await api.createVehicleTool({ vehicleId: vtVehicleId, name: vtName, barcode: vtBarcode })
      setVtVehicleId(''); setVtName(''); setVtBarcode('')
      setShowVehicleForm(false)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const handleAddOnDemandTool = async (e: React.FormEvent) => {
    e.preventDefault()
    setSubmitting(true)
    try {
      await api.createOnDemandTool({ name: odName, barcode: odBarcode, totalQuantity: Number(odQuantity) })
      setOdName(''); setOdBarcode(''); setOdQuantity('')
      setShowOnDemandForm(false)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const handleApproveRequest = async (id: string) => {
    if (!currentUser) return
    try {
      await api.approveToolRequest(id, currentUser.id)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  const handleRejectRequest = async (id: string) => {
    try {
      await api.rejectToolRequest(id)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  const handleReturnRequest = async (id: string) => {
    try {
      await api.returnToolRequest(id)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">إدارة المخزون والأدوات</h2>
      <p className="mt-1 text-slate-500">تتبع الأدوات الخاصة والمركبات وأدوات حسب الحاجة وطلبات الأدوات.</p>

      {/* Tabs */}
      <div className="mt-6 flex gap-1 rounded-xl bg-slate-100 p-1">
        {tabs.map((tab) => (
          <button
            key={tab.key}
            onClick={() => setActiveTab(tab.key)}
            className={`flex-1 rounded-lg px-4 py-2.5 text-sm font-medium transition-all ${
              activeTab === tab.key
                ? 'bg-gradient-to-l from-brand-500 to-brand-800 text-white shadow-md'
                : 'text-slate-600 hover:text-brand-700'
            }`}
          >
            {tab.label}
          </button>
        ))}
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {!loading && !error && (
        <div className="mt-6">
          {/* Today's Inventory Checks Tab */}
          {activeTab === 'todaychecks' && (
            <div className="space-y-3">
              <p className="text-sm text-slate-500">
                نتائج جرد الفنيين لهذا اليوم — لمين اكو نقص، وفرلهم البديل قبل ما يطلعون للحجز.
              </p>
              {todaysChecks.length === 0 && (
                <div className="rounded-xl border border-white bg-white p-8 text-center text-slate-400 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  لا يوجد أي فني سجّل جرد أدواته اليوم بعد
                </div>
              )}
              {todaysChecks.map((c) => (
                <div key={c.id} className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <div className="flex items-center justify-between">
                    <span className="font-bold text-slate-800">{c.employee?.name || '-'}</span>
                    <span className={`rounded-full px-3 py-1 text-xs font-bold ${c.complete ? 'bg-emerald-100 text-emerald-700' : 'bg-amber-100 text-amber-700'}`}>
                      {c.complete ? 'العدة كاملة' : 'فيه نقص'}
                    </span>
                  </div>
                  <p className="mt-1 text-xs text-slate-400">{new Date(c.checkedAt).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })}</p>
                  {!c.complete && c.missingItems && (
                    <p className="mt-2 rounded-lg bg-amber-50 px-3 py-2 text-sm text-amber-800">الناقص: {c.missingItems}</p>
                  )}
                  {!c.complete && (
                    <div className="mt-2 flex items-center justify-between">
                      {c.resolved ? (
                        <span className="rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-700">
                          ✓ تم التوفير{c.resolvedBy ? ` بواسطة ${c.resolvedBy.name}` : ''}
                        </span>
                      ) : (
                        <span className="rounded-full bg-slate-100 px-3 py-1 text-xs font-bold text-slate-500">لسه ماتوفر</span>
                      )}
                      {canManageInventory && !c.resolved && (
                        <button
                          onClick={() => handleResolveCheck(c.id)}
                          disabled={resolvingId === c.id}
                          className="rounded-lg bg-emerald-600 px-3 py-1.5 text-xs font-bold text-white hover:bg-emerald-700 disabled:opacity-50"
                        >
                          {resolvingId === c.id ? 'جاري...' : 'تم توفير الاحتياج ✓'}
                        </button>
                      )}
                    </div>
                  )}
                </div>
              ))}
            </div>
          )}

          {/* Personal Tools Tab */}
          {activeTab === 'personal' && (
            <div>
              <div className="mb-4 flex justify-end">
                <button
                  onClick={() => setShowPersonalForm(!showPersonalForm)}
                  className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30"
                >
                  إضافة أداة خاصة
                </button>
              </div>
              {showPersonalForm && (
                <form onSubmit={handleAddPersonalTool} className="mb-6 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">الموظف</label>
                      <select required value={ptEmployeeId} onChange={(e) => setPtEmployeeId(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500">
                        <option value="">اختر الموظف</option>
                        {employees.map((emp) => <option key={emp.id} value={emp.id}>{emp.name}</option>)}
                      </select>
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">اسم الأداة</label>
                      <input required value={ptName} onChange={(e) => setPtName(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">الباركود</label>
                      <input required value={ptBarcode} onChange={(e) => setPtBarcode(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
                    </div>
                  </div>
                  <div className="mt-4">
                    <button type="submit" disabled={submitting} className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 disabled:opacity-50">
                      {submitting ? 'جاري الحفظ...' : 'إضافة'}
                    </button>
                  </div>
                </form>
              )}
              <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200 text-right">
                    <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                      <tr>
                        <th className="px-4 py-3 text-sm font-semibold">الموظف</th>
                        <th className="px-4 py-3 text-sm font-semibold">اسم الأداة</th>
                        <th className="px-4 py-3 text-sm font-semibold">الباركود</th>
                        <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {personalTools.map((t) => (
                        <tr key={t.id} className="transition-colors hover:bg-slate-50">
                          <td className="px-4 py-3 font-medium">{t.employee?.name || '-'}</td>
                          <td className="px-4 py-3">{t.name}</td>
                          <td className="px-4 py-3 font-mono text-sm text-slate-500">{t.barcode}</td>
                          <td className="px-4 py-3">
                            <span className="rounded-full bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700">{t.status}</span>
                          </td>
                        </tr>
                      ))}
                      {personalTools.length === 0 && (
                        <tr><td colSpan={4} className="px-4 py-6 text-center text-slate-400">لا توجد أدوات خاصة</td></tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}

          {/* Vehicle Tools Tab */}
          {activeTab === 'vehicle' && (
            <div>
              <div className="mb-4 flex justify-end">
                <button
                  onClick={() => setShowVehicleForm(!showVehicleForm)}
                  className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30"
                >
                  إضافة أداة مركبة
                </button>
              </div>
              {showVehicleForm && (
                <form onSubmit={handleAddVehicleTool} className="mb-6 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">رقم لوحة المركبة</label>
                      <input required value={vtVehicleId} onChange={(e) => setVtVehicleId(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">اسم الأداة</label>
                      <input required value={vtName} onChange={(e) => setVtName(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">الباركود</label>
                      <input required value={vtBarcode} onChange={(e) => setVtBarcode(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
                    </div>
                  </div>
                  <div className="mt-4">
                    <button type="submit" disabled={submitting} className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 disabled:opacity-50">
                      {submitting ? 'جاري الحفظ...' : 'إضافة'}
                    </button>
                  </div>
                </form>
              )}
              <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200 text-right">
                    <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                      <tr>
                        <th className="px-4 py-3 text-sm font-semibold">لوحة المركبة</th>
                        <th className="px-4 py-3 text-sm font-semibold">اسم الأداة</th>
                        <th className="px-4 py-3 text-sm font-semibold">الباركود</th>
                        <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {vehicleTools.map((t) => (
                        <tr key={t.id} className="transition-colors hover:bg-slate-50">
                          <td className="px-4 py-3 font-medium">{t.vehicleId}</td>
                          <td className="px-4 py-3">{t.name}</td>
                          <td className="px-4 py-3 font-mono text-sm text-slate-500">{t.barcode}</td>
                          <td className="px-4 py-3">
                            <span className="rounded-full bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700">{t.status}</span>
                          </td>
                        </tr>
                      ))}
                      {vehicleTools.length === 0 && (
                        <tr><td colSpan={4} className="px-4 py-6 text-center text-slate-400">لا توجد أدوات مركبات</td></tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}

          {/* On-Demand Tools Tab */}
          {activeTab === 'ondemand' && (
            <div>
              <div className="mb-4 flex justify-end">
                <button
                  onClick={() => setShowOnDemandForm(!showOnDemandForm)}
                  className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30"
                >
                  إضافة أداة
                </button>
              </div>
              {showOnDemandForm && (
                <form onSubmit={handleAddOnDemandTool} className="mb-6 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">اسم الأداة</label>
                      <input required value={odName} onChange={(e) => setOdName(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">الباركود</label>
                      <input required value={odBarcode} onChange={(e) => setOdBarcode(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">الكمية الإجمالية</label>
                      <input required type="number" min="1" value={odQuantity} onChange={(e) => setOdQuantity(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
                    </div>
                  </div>
                  <div className="mt-4">
                    <button type="submit" disabled={submitting} className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 disabled:opacity-50">
                      {submitting ? 'جاري الحفظ...' : 'إضافة'}
                    </button>
                  </div>
                </form>
              )}
              <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200 text-right">
                    <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                      <tr>
                        <th className="px-4 py-3 text-sm font-semibold">اسم الأداة</th>
                        <th className="px-4 py-3 text-sm font-semibold">الباركود</th>
                        <th className="px-4 py-3 text-sm font-semibold">الكمية الإجمالية</th>
                        <th className="px-4 py-3 text-sm font-semibold">الكمية المتاحة</th>
                        {canManageOnDemand && <th className="px-4 py-3 text-sm font-semibold">تعديل</th>}
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {onDemandTools.map((t) => (
                        <tr key={t.id} className="transition-colors hover:bg-slate-50">
                          <td className="px-4 py-3 font-medium">{t.name}</td>
                          <td className="px-4 py-3 font-mono text-sm text-slate-500">{t.barcode}</td>
                          <td className="px-4 py-3">{t.totalQuantity}</td>
                          <td className="px-4 py-3">
                            <span className={`rounded-full px-3 py-1 text-sm font-bold ${t.availableQuantity > 0 ? 'bg-green-50 text-green-700' : 'bg-red-50 text-red-700'}`}>
                              {t.availableQuantity}
                            </span>
                          </td>
                          {canManageOnDemand && (
                            <td className="px-4 py-3">
                              <button
                                type="button"
                                onClick={async () => {
                                  const input = prompt(`عدّل الكمية المتاحة لأداة "${t.name}" (الإجمالية: ${t.totalQuantity})`, String(t.availableQuantity))
                                  if (input === null) return
                                  const num = Number(input)
                                  if (Number.isNaN(num) || num < 0) { alert('رقم غير صحيح'); return }
                                  try {
                                    const updated = await api.updateOnDemandTool(t.id, { availableQuantity: num })
                                    setOnDemandTools((prev) => prev.map((x) => (x.id === updated.id ? updated : x)))
                                  } catch (e) {
                                    alert(e instanceof Error ? e.message : 'تعذر التعديل')
                                  }
                                }}
                                className="rounded-lg bg-slate-100 px-3 py-1 text-xs font-semibold text-slate-600 hover:bg-slate-200"
                              >
                                تعديل
                              </button>
                            </td>
                          )}
                        </tr>
                      ))}
                      {onDemandTools.length === 0 && (
                        <tr><td colSpan={canManageOnDemand ? 5 : 4} className="px-4 py-6 text-center text-slate-400">لا توجد أدوات حسب الحاجة</td></tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}

          {/* Tool Requests Tab */}
          {activeTab === 'requests' && (
            <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <div className="overflow-x-auto">
                <table className="min-w-full divide-y divide-gray-200 text-right">
                  <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                    <tr>
                      <th className="px-4 py-3 text-sm font-semibold">الموظف</th>
                      <th className="px-4 py-3 text-sm font-semibold">الأداة</th>
                      <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                      <th className="px-4 py-3 text-sm font-semibold">تاريخ الطلب</th>
                      <th className="px-4 py-3 text-sm font-semibold">تاريخ الاستلام</th>
                      <th className="px-4 py-3 text-sm font-semibold">تاريخ الإرجاع</th>
                      {canApproveRequests && <th className="px-4 py-3 text-sm font-semibold">إجراءات</th>}
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-slate-100">
                    {toolRequests.map((r) => (
                      <tr key={r.id} className="transition-colors hover:bg-slate-50">
                        <td className="px-4 py-3 font-medium">{r.employee?.name || '-'}</td>
                        <td className="px-4 py-3">{r.tool?.name || '-'}</td>
                        <td className="px-4 py-3">
                          <span className={`rounded-full px-3 py-1 text-xs font-bold ${requestStatusColors[r.status]}`}>
                            {requestStatusLabels[r.status]}
                          </span>
                        </td>
                        <td className="px-4 py-3 text-slate-500">
                          {new Date(r.requestedAt).toLocaleDateString('ar-IQ')}
                        </td>
                        <td className="px-4 py-3 text-slate-500">
                          {r.approvedAt ? new Date(r.approvedAt).toLocaleDateString('ar-IQ') : '-'}
                        </td>
                        <td className="px-4 py-3 text-slate-500">
                          {r.returnedAt ? new Date(r.returnedAt).toLocaleDateString('ar-IQ') : '-'}
                        </td>
                        {canApproveRequests && (
                          <td className="px-4 py-3">
                            <div className="flex gap-2">
                              {r.status === 'PENDING' && (
                                <>
                                  <button
                                    onClick={() => handleApproveRequest(r.id)}
                                    className="rounded-lg bg-green-100 px-3 py-1 text-xs font-medium text-green-700 hover:bg-green-200"
                                  >
                                    موافقة
                                  </button>
                                  <button
                                    onClick={() => handleRejectRequest(r.id)}
                                    className="rounded-lg bg-red-100 px-3 py-1 text-xs font-medium text-red-700 hover:bg-red-200"
                                  >
                                    رفض
                                  </button>
                                </>
                              )}
                              {r.status === 'APPROVED' && (
                                <button
                                  onClick={() => handleReturnRequest(r.id)}
                                  className="rounded-lg bg-blue-100 px-3 py-1 text-xs font-medium text-blue-700 hover:bg-blue-200"
                                >
                                  استرجاع
                                </button>
                              )}
                            </div>
                          </td>
                        )}
                      </tr>
                    ))}
                    {toolRequests.length === 0 && (
                      <tr>
                        <td colSpan={canApproveRequests ? 7 : 6} className="px-4 py-6 text-center text-slate-400">
                          لا توجد طلبات أدوات
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
