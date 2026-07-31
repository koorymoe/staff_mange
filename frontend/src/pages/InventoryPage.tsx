import { useEffect, useState } from 'react'
import { api, type PersonalTool, type VehicleTool, type OnDemandTool, type ToolRequest, type ToolRequestItem, type Employee, type InventoryCheck, type PersonalToolTemplateItem, type BookingToolCheck, type VehicleToolCheck } from '../api'
import { useSession } from '../session'

type TabKey = 'todaychecks' | 'personal' | 'vehicle' | 'ondemand' | 'requests' | 'template' | 'reports'

const tabs: { key: TabKey; label: string }[] = [
  { key: 'todaychecks', label: 'نتائج جرد اليوم' },
  { key: 'personal', label: 'أدوات خاصة' },
  { key: 'vehicle', label: 'أدوات المركبات' },
  { key: 'ondemand', label: 'أدوات حسب الحاجة' },
  { key: 'requests', label: 'طلبات الأدوات' },
  { key: 'template', label: 'العدة القياسية' },
  { key: 'reports', label: 'تقارير النواقص' },
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
  const { employee: currentUser, permissions } = useSession()
  const isAdmin = currentUser?.role === 'ADMIN'
  // "جرد الأدوات" الآن صلاحية مفعّلة فعلياً بالباك-إند (requireHROrInventory) —
  // أي موظف مُنح صلاحية inventory من صفحة الصلاحيات يقدر يدير الأدوات، حتى لو
  // دوره مختلف عن HR_COORDINATOR (مثلاً PROCUREMENT_ADMIN).
  const canManageInventory = isAdmin || currentUser?.role === 'HR_COORDINATOR' || permissions.includes('inventory')
  // إداري الكميات مسؤول عن توفير وإضافة "أدوات حسب الحاجة" (نفس مسؤوليته
  // بطلبات المواد الناقصة من الموظفين) — بالإضافة للأدمن.
  const canManageOnDemand = isAdmin || currentUser?.role === 'PROCUREMENT_ADMIN'
  // موافقة/رفض طلبات الأدوات صارت صلاحية مستقلة (tool_requests_approve) بدل
  // ما تكون مشتقة من "جرد الأدوات" أو من الدور — إداري الكميات كان ينحرم منها
  // مهما انمنحت له صلاحيات ثانية. الأدوار القديمة باقية حتى ما ينقطع أحد.
  const canApproveRequests =
    isAdmin ||
    permissions.includes('tool_requests_approve') ||
    currentUser?.role === 'HR_COORDINATOR' ||
    currentUser?.role === 'MONITOR'
  // حذف الطلب يمحي أثره نهائياً — لمدير النظام والمالك فقط، مطابق للباك إند.
  const canDeleteRequests = isAdmin
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

  // العدة القياسية (PersonalToolTemplateItem)
  const [templateItems, setTemplateItems] = useState<PersonalToolTemplateItem[]>([])
  const [templateName, setTemplateName] = useState('')
  const [templateSubmitting, setTemplateSubmitting] = useState(false)

  // تقارير النواقص
  const [bookingToolChecks, setBookingToolChecks] = useState<BookingToolCheck[]>([])
  const [vehicleToolChecks, setVehicleToolChecks] = useState<VehicleToolCheck[]>([])

  const [submitting, setSubmitting] = useState(false)

  const load = () => {
    Promise.all([
      api.getPersonalTools(),
      api.getVehicleTools(),
      api.getOnDemandTools(),
      api.getToolRequests(),
      api.getEmployees(),
      api.getPersonalToolTemplate(),
    ])
      .then(([pt, vt, od, tr, emps, tmpl]) => {
        setPersonalTools(pt)
        setVehicleTools(vt)
        setOnDemandTools(od)
        setToolRequests(tr)
        setEmployees(emps)
        setTemplateItems(tmpl)
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  useEffect(() => {
    if (activeTab !== 'reports') return
    api.getVehicleToolChecks().then(setVehicleToolChecks).catch(() => setVehicleToolChecks([]))
    api.getAllBookingToolChecks().then(setBookingToolChecks).catch(() => setBookingToolChecks([]))
  }, [activeTab])

  const handleAddTemplateItem = async (e: React.FormEvent) => {
    e.preventDefault()
    setTemplateSubmitting(true)
    try {
      const item = await api.createPersonalToolTemplateItem(templateName)
      setTemplateItems((prev) => [...prev, item])
      setTemplateName('')
    } catch { /* ignore */ }
    finally { setTemplateSubmitting(false) }
  }

  const handleDeleteTemplateItem = async (id: string) => {
    try {
      await api.deletePersonalToolTemplateItem(id)
      setTemplateItems((prev) => prev.filter((t) => t.id !== id))
    } catch { /* ignore */ }
  }

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

  // الموافقة تمر بنافذة: إذا الأداة موجودة بالشركة تنوافق مباشرة، وإذا مو
  // موجودة لازم إداري الكميات يدخل سعر شرائها وينفتح طلب مشتريات للمحاسب.
  const [approveTarget, setApproveTarget] = useState<ToolRequestItem | null>(null)
  const [approvePrice, setApprovePrice] = useState('')
  const [approveError, setApproveError] = useState('')
  const [approving, setApproving] = useState(false)

  const confirmApprove = async () => {
    if (!currentUser || !approveTarget) return
    const inStock = (approveTarget.tool?.availableQuantity ?? 0) > 0
    const price = parseFloat(approvePrice)
    if (!inStock && (!approvePrice.trim() || isNaN(price) || price <= 0)) {
      setApproveError('الأداة مو متوفرة — لازم تدخل سعر الشراء حتى يتحول الطلب للمحاسب')
      return
    }
    setApproving(true)
    setApproveError('')
    try {
      await api.approveToolRequest(approveTarget.id, currentUser.id, inStock ? undefined : price)
      setApproveTarget(null)
      setApprovePrice('')
      load()
    } catch (err) {
      setApproveError(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setApproving(false)
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

  const handleDeleteRequest = async (id: string) => {
    if (!confirm('حذف طلب الأداة هذا نهائياً؟')) return
    try {
      await api.deleteToolRequest(id)
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
                      <th className="px-4 py-3 text-sm font-semibold">السبب والشرح</th>
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
                        <td className="px-4 py-3 max-w-xs">
                          {r.reasonLabel ? (
                            <>
                              <div className="font-medium text-brand-800">{r.reasonLabel}</div>
                              {r.description && <div className="mt-0.5 text-xs text-slate-500">{r.description}</div>}
                            </>
                          ) : (
                            <span className="text-slate-400">-</span>
                          )}
                          {r.purchasePrice != null && (
                            <div className="mt-1 text-xs font-bold text-amber-700">
                              انشترت بسعر {r.purchasePrice.toLocaleString('ar-IQ')} د.ع — تحوّلت للمحاسب
                            </div>
                          )}
                        </td>
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
                                    onClick={() => { setApproveTarget(r); setApprovePrice(''); setApproveError('') }}
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
                              {canDeleteRequests && (
                                <button
                                  onClick={() => handleDeleteRequest(r.id)}
                                  className="rounded-lg bg-slate-100 px-3 py-1 text-xs font-medium text-slate-600 hover:bg-slate-200"
                                >
                                  حذف
                                </button>
                              )}
                            </div>
                          </td>
                        )}
                      </tr>
                    ))}
                    {toolRequests.length === 0 && (
                      <tr>
                        <td colSpan={canApproveRequests ? 8 : 7} className="px-4 py-6 text-center text-slate-400">
                          لا توجد طلبات أدوات
                        </td>
                      </tr>
                    )}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Personal Tool Standard Kit Template Tab */}
          {activeTab === 'template' && (
            <div>
              {canManageInventory && (
                <form onSubmit={handleAddTemplateItem} className="mb-6 flex gap-3 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <input
                    required
                    value={templateName}
                    onChange={(e) => setTemplateName(e.target.value)}
                    placeholder="اسم الأداة القياسية"
                    className="flex-1 rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
                  />
                  <button type="submit" disabled={templateSubmitting} className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 disabled:opacity-50">
                    {templateSubmitting ? 'جاري الإضافة...' : 'إضافة للعدة القياسية'}
                  </button>
                </form>
              )}
              <p className="mb-4 text-sm text-slate-500">
                أي أداة تُضاف هنا تنضاف فوراً لعدة كل الموظفين الحاليين، وأي موظف جديد ياخذ العدة القياسية كاملة تلقائياً وقت إنشاء حسابه.
              </p>
              <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200 text-right">
                    <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                      <tr>
                        <th className="px-4 py-3 text-sm font-semibold">اسم الأداة</th>
                        {canManageInventory && <th className="px-4 py-3 text-sm font-semibold">إجراءات</th>}
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {templateItems.map((item) => (
                        <tr key={item.id} className="transition-colors hover:bg-slate-50">
                          <td className="px-4 py-3 font-medium">{item.name}</td>
                          {canManageInventory && (
                            <td className="px-4 py-3">
                              <button
                                onClick={() => handleDeleteTemplateItem(item.id)}
                                className="rounded-lg bg-red-100 px-3 py-1 text-xs font-medium text-red-700 hover:bg-red-200"
                              >
                                حذف
                              </button>
                            </td>
                          )}
                        </tr>
                      ))}
                      {templateItems.length === 0 && (
                        <tr><td colSpan={canManageInventory ? 2 : 1} className="px-4 py-6 text-center text-slate-400">لا توجد عناصر بالعدة القياسية</td></tr>
                      )}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          )}

          {/* Missing-tool Reports Tab: booking tool checks + vehicle tool checks */}
          {activeTab === 'reports' && (
            <div className="space-y-8">
              <div>
                <h3 className="mb-3 text-lg font-bold text-slate-700">نواقص الأدوات الشخصية عند استلام الحجوزات</h3>
                <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200 text-right">
                      <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                        <tr>
                          <th className="px-4 py-3 text-sm font-semibold">الموظف</th>
                          <th className="px-4 py-3 text-sm font-semibold">الحجز</th>
                          <th className="px-4 py-3 text-sm font-semibold">الأدوات الناقصة</th>
                          <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-100">
                        {bookingToolChecks.map((c) => (
                          <tr key={c.id} className="transition-colors hover:bg-slate-50">
                            <td className="px-4 py-3 font-medium">{c.employee?.name || '-'}</td>
                            <td className="px-4 py-3 font-mono text-xs text-slate-500">{c.bookingId}</td>
                            <td className="px-4 py-3">{c.missingItems || 'لا يوجد نقص'}</td>
                            <td className="px-4 py-3 text-slate-500">{new Date(c.checkedAt).toLocaleDateString('ar-IQ')}</td>
                          </tr>
                        ))}
                        {bookingToolChecks.length === 0 && (
                          <tr><td colSpan={4} className="px-4 py-6 text-center text-slate-400">لا توجد فحوصات مسجّلة</td></tr>
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>

              <div>
                <h3 className="mb-3 text-lg font-bold text-slate-700">نواقص أدوات المركبات العامة عند بدء المهام (ليدر)</h3>
                <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <div className="overflow-x-auto">
                    <table className="min-w-full divide-y divide-gray-200 text-right">
                      <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                        <tr>
                          <th className="px-4 py-3 text-sm font-semibold">الموظف</th>
                          <th className="px-4 py-3 text-sm font-semibold">المركبة</th>
                          <th className="px-4 py-3 text-sm font-semibold">الأدوات الناقصة</th>
                          <th className="px-4 py-3 text-sm font-semibold">التاريخ</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-100">
                        {vehicleToolChecks.map((c) => (
                          <tr key={c.id} className="transition-colors hover:bg-slate-50">
                            <td className="px-4 py-3 font-medium">{c.employee?.name || '-'}</td>
                            <td className="px-4 py-3 font-mono text-xs text-slate-500">{c.vehicleId}</td>
                            <td className="px-4 py-3">{c.missingToolNames || 'لا يوجد نقص'}</td>
                            <td className="px-4 py-3 text-slate-500">{new Date(c.createdAt).toLocaleDateString('ar-IQ')}</td>
                          </tr>
                        ))}
                        {vehicleToolChecks.length === 0 && (
                          <tr><td colSpan={4} className="px-4 py-6 text-center text-slate-400">لا توجد فحوصات مسجّلة</td></tr>
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      {approveTarget && (() => {
        const inStock = (approveTarget.tool?.availableQuantity ?? 0) > 0
        return (
          <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" dir="rtl">
            <div className="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl">
              <h3 className="text-lg font-bold text-brand-900">موافقة على طلب أداة</h3>

              <div className="mt-4 space-y-2 rounded-xl bg-slate-50 p-4 text-sm">
                <div className="flex justify-between"><span className="text-slate-500">الموظف</span><span className="font-medium">{approveTarget.employee?.name || '-'}</span></div>
                <div className="flex justify-between"><span className="text-slate-500">الأداة</span><span className="font-medium">{approveTarget.tool?.name || '-'}</span></div>
                {approveTarget.reasonLabel && (
                  <div className="flex justify-between gap-4"><span className="text-slate-500">السبب</span><span className="font-medium text-left">{approveTarget.reasonLabel}</span></div>
                )}
                {approveTarget.description && (
                  <div className="border-t border-slate-200 pt-2">
                    <div className="text-slate-500">شرح الموظف</div>
                    <div className="mt-0.5 text-slate-700">{approveTarget.description}</div>
                  </div>
                )}
              </div>

              {inStock ? (
                <p className="mt-4 rounded-xl bg-emerald-50 p-4 text-sm font-medium text-emerald-700">
                  ✓ الأداة متوفرة بالشركة ({approveTarget.tool?.availableQuantity} قطعة) — تنصرف من المخزن مباشرة بدون شراء.
                </p>
              ) : (
                <>
                  <p className="mt-4 rounded-xl bg-amber-50 p-4 text-sm font-medium text-amber-800">
                    ⚠️ الأداة مو متوفرة بالشركة — لازم تنشترى. اكتب سعر الشراء وراح ينفتح طلب مشتريات تلقائياً يوصل للمحاسب.
                  </p>
                  <label className="mt-4 block text-sm font-semibold text-slate-700">سعر الشراء (دينار عراقي)</label>
                  <input
                    type="number"
                    min="0"
                    value={approvePrice}
                    onChange={(e) => { setApprovePrice(e.target.value); setApproveError('') }}
                    placeholder="مثال: 25000"
                    className="mt-1.5 w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500"
                  />
                </>
              )}

              {approveError && <p className="mt-3 rounded-lg bg-red-50 p-3 text-sm text-red-600">{approveError}</p>}

              <div className="mt-5 flex gap-3">
                <button
                  onClick={confirmApprove}
                  disabled={approving}
                  className="flex-1 rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 py-2.5 font-bold text-white disabled:opacity-50"
                >
                  {approving ? 'جاري...' : inStock ? 'تأكيد الموافقة' : 'موافقة وإرسال للمحاسب'}
                </button>
                <button
                  onClick={() => setApproveTarget(null)}
                  className="rounded-xl border border-slate-300 px-6 py-2.5 font-medium text-slate-600 hover:bg-slate-50"
                >
                  إلغاء
                </button>
              </div>
            </div>
          </div>
        )
      })()}
    </div>
  )
}
