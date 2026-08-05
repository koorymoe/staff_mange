import { Fragment, useEffect, useState, useMemo } from 'react'
import { api, type PersonalTool, type VehicleTool, type OnDemandTool, type ToolRequest, type ToolRequestItem, type Employee, type InventoryCheck, type PersonalToolTemplateItem, type BookingToolCheck, type VehicleToolCheck, type Vehicle, type PersonalToolEvent, type PersonalToolStatus, personalToolStatusLabels, personalToolStatusColors } from '../api'
import { useSession, roleLabels } from '../session'

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

const vehicleToolStatusLabels: Record<string, string> = {
  AVAILABLE: 'متوفرة',
  CHECKED_OUT: 'مصروفة',
  DAMAGED: 'تالفة / مفقودة',
}

const requestStatusColors: Record<ToolRequest['status'], string> = {
  PENDING: 'bg-yellow-100 text-yellow-800',
  APPROVED: 'bg-green-100 text-green-800',
  REJECTED: 'bg-red-100 text-red-800',
  RETURNED: 'bg-gray-100 text-gray-800',
}

// groupTools يجمع أدوات الموظف بالاسم: صف واحد لكل أداة مكتوب جنبه
// العدد، بدل صف مستقل لكل نسخة. الباركود يبقى ظاهر بالتفاصيل لأنه
// يميّز النسخة الواحدة وما ينفع يتجمّع.
function groupTools(tools: PersonalTool[]) {
  const byName = new Map<string, PersonalTool[]>()
  for (const t of tools) {
    const key = t.name.trim()
    const list = byName.get(key)
    if (list) list.push(t)
    else byName.set(key, [t])
  }
  return [...byName.entries()]
    .map(([name, units]) => {
      const counts = new Map<string, number>()
      for (const u of units) counts.set(u.status, (counts.get(u.status) ?? 0) + 1)
      return {
        name,
        units,
        // الأكثر عدداً أول — الحالة الغالبة تبيّن بالنظرة الأولى
        statuses: [...counts.entries()].sort((a, b) => b[1] - a[1]),
      }
    })
    .sort((a, b) => a.name.localeCompare(b.name, 'ar'))
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
  // الموظف المفتوحة عدته بتبويب "أدوات خاصة" (null = عرض كل الموظفين)
  const [selectedKitEmployeeId, setSelectedKitEmployeeId] = useState<string | null>(null)
  // تعديل أداة شخصية
  const [toolEdit, setToolEdit] = useState<PersonalTool | null>(null)
  const [toolEditName, setToolEditName] = useState('')
  const [toolEditBarcode, setToolEditBarcode] = useState('')
  const [toolEditStatus, setToolEditStatus] = useState<PersonalToolStatus>('AVAILABLE')
  const [toolEditNote, setToolEditNote] = useState('')
  const [toolSaving, setToolSaving] = useState(false)
  // سجل حركة أداة
  const [historyTool, setHistoryTool] = useState<PersonalTool | null>(null)
  const [toolEvents, setToolEvents] = useState<PersonalToolEvent[]>([])
  const [historyLoading, setHistoryLoading] = useState(false)
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
  // أي أداة مفتوحة تفاصيلها — المفتاح "رقم الموظف:اسم الأداة"
  const [expandedTool, setExpandedTool] = useState<string | null>(null)
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

  // Vehicle tools
  const [vehicleTools, setVehicleTools] = useState<VehicleTool[]>([])
  const [showVehicleForm, setShowVehicleForm] = useState(false)
  const [vtVehicleId, setVtVehicleId] = useState('')
  const [vtName, setVtName] = useState('')
  const [vtQuantity, setVtQuantity] = useState('1')
  const [vehicles, setVehicles] = useState<Vehicle[]>([])
  // تعديل أداة مركبة
  const [vtEdit, setVtEdit] = useState<VehicleTool | null>(null)
  const [vtEditName, setVtEditName] = useState('')
  const [vtEditQuantity, setVtEditQuantity] = useState('1')
  const [vtEditVehicleId, setVtEditVehicleId] = useState('')
  const [vtEditStatus, setVtEditStatus] = useState<VehicleTool['status']>('AVAILABLE')

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
  // زر "إضافة أداة" الموحّد: يسأل أول شي قياسية لو خاصة. القياسية تروح لكل
  // موظف مستحق، والخاصة تنطلب اسم الموظف وتنضاف له هو بس.
  const [addToolKind, setAddToolKind] = useState<'standard' | 'private'>('standard')
  const [addToolEmployeeId, setAddToolEmployeeId] = useState('')
  // إضافة كميات للمخزون — إداري الكميات يزيد على الكمية الموجودة بأثر مسجّل
  const [stockTarget, setStockTarget] = useState<OnDemandTool | null>(null)
  const [stockQty, setStockQty] = useState('')
  const [stockPrice, setStockPrice] = useState('')
  const [stockSupplier, setStockSupplier] = useState('')

  // تقارير النواقص
  const [bookingToolChecks, setBookingToolChecks] = useState<BookingToolCheck[]>([])
  const [vehicleToolChecks, setVehicleToolChecks] = useState<VehicleToolCheck[]>([])

  const [submitting, setSubmitting] = useState(false)

  // أسماء العدة القياسية — أساس المقارنة. نطبّعها بحذف الفراغات حتى "مفتاح "
  // و"مفتاح" ما ينحسبون أداتين مختلفتين.
  const templateNames = useMemo(
    () => new Set(templateItems.map((t) => t.name.trim())),
    [templateItems],
  )

  // لكل موظف: عدته، شنو ناقص من القياسية، وشنو زايد عليها.
  //
  // ⚠️ الي عنده عدة هم الي يشتغلون بيدينهم بالميدان: الفني والليدر.
  // نفس شرط السيرفر (toolKitEligibleSQL) بالضبط. بدونه الشاشة تعرض
  // «ناقص 37» للمحاسب وموظف المبيعات وإداري الكوادر — ناس ما عندهم
  // عدة أصلاً ولا يتحاسبون عليها، فالتقرير كله يصير ضوضاء.
  const kitSummaries = useMemo(() => {
    const byEmployee = new Map<string, PersonalTool[]>()
    for (const t of personalTools) {
      const list = byEmployee.get(t.employeeId)
      if (list) list.push(t)
      else byEmployee.set(t.employeeId, [t])
    }
    return employees
      .filter((emp) => emp.role === 'TECHNICIAN' || emp.isLeader)
      .map((emp) => {
        const tools = byEmployee.get(emp.id) || []
        const owned = new Set(tools.map((t) => t.name.trim()))
        return {
          employee: emp,
          tools,
          missing: [...templateNames].filter((n) => !owned.has(n)),
          extra: tools.filter((t) => !templateNames.has(t.name.trim())),
        }
      })
      // الي عنده نقص أول القائمة — هذا الي يحتاج انتباه الإداري
      .sort((a, b) => b.missing.length - a.missing.length || a.employee.name.localeCompare(b.employee.name, 'ar'))
  }, [employees, personalTools, templateNames])

  const openToolEdit = (t: PersonalTool) => {
    setToolEdit(t)
    setToolEditName(t.name)
    setToolEditBarcode(t.barcode || '')
    setToolEditStatus((t.status as PersonalToolStatus) || 'AVAILABLE')
    setToolEditNote('')
  }

  const saveToolEdit = async () => {
    if (!toolEdit) return
    setToolSaving(true)
    try {
      await api.updatePersonalTool(toolEdit.id, {
        name: toolEditName.trim(),
        barcode: toolEditBarcode.trim() || undefined,
        status: toolEditStatus,
        note: toolEditNote.trim() || undefined,
      })
      setToolEdit(null)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر حفظ التعديل')
    } finally {
      setToolSaving(false)
    }
  }

  const handleDeleteTool = async (t: PersonalTool) => {
    if (!confirm(`حذف «${t.name}» من عدة هذا الموظف؟ (تبقى بسجل الحركة)`)) return
    try {
      await api.deletePersonalTool(t.id)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر الحذف')
    }
  }

  const openToolHistory = async (t: PersonalTool) => {
    setHistoryTool(t)
    setToolEvents([])
    setHistoryLoading(true)
    try {
      setToolEvents(await api.getToolEvents({ toolId: t.id }))
    } catch {
      setToolEvents([])
    } finally {
      setHistoryLoading(false)
    }
  }

  const vehicleToolStats = useMemo(() => ({
    kinds: vehicleTools.length,
    pieces: vehicleTools.reduce((s, t) => s + (t.quantity || 1), 0),
    vehicles: new Set(vehicleTools.map((t) => t.vehicleId)).size,
    problem: vehicleTools.filter((t) => t.status !== 'AVAILABLE').length,
  }), [vehicleTools])

  const load = () => {
    Promise.all([
      api.getPersonalTools(),
      api.getVehicleTools(),
      api.getOnDemandTools(),
      api.getToolRequests(),
      api.getEmployees(),
      api.getPersonalToolTemplate(),
      // قائمة السيارات لازمة للقائمة المنسدلة بأدوات المركبات — لو الموظف
      // ما عنده صلاحية المركبات نكمل بقائمة فاضية بدل ما تفشل الصفحة كلها.
      api.getVehicles().catch(() => [] as Vehicle[]),
    ])
      .then(([pt, vt, od, tr, emps, tmpl, vhs]) => {
        setPersonalTools(pt)
        setVehicleTools(vt)
        setOnDemandTools(od)
        setToolRequests(tr)
        setEmployees(emps)
        setTemplateItems(tmpl)
        setVehicles(vhs)
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
    if (addToolKind === 'private' && !addToolEmployeeId) {
      alert('اختر الموظف الي عنده هاي الأداة الخاصة')
      return
    }
    setTemplateSubmitting(true)
    try {
      if (addToolKind === 'standard') {
        // قياسية: تنضاف لكل موظف مستحق (فني/ليدر) تلقائياً بالخلفية
        const item = await api.createPersonalToolTemplateItem(templateName)
        setTemplateItems((prev) => [...prev, item])
      } else {
        // خاصة: تنضاف للموظف المختار هو بس، وما تدخل بالعدة القياسية
        await api.createPersonalTool({ employeeId: addToolEmployeeId, name: templateName, barcode: '' })
        load()
      }
      setTemplateName('')
      setAddToolEmployeeId('')
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
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
      await api.createPersonalTool({ employeeId: ptEmployeeId, name: ptName, barcode: '' })
      setPtEmployeeId(''); setPtName('')
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
      await api.createVehicleTool({ vehicleId: vtVehicleId, name: vtName, quantity: Number(vtQuantity) || 1 })
      setVtVehicleId(''); setVtName(''); setVtQuantity('1')
      setShowVehicleForm(false)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const openVehicleToolEdit = (t: VehicleTool) => {
    setVtEdit(t)
    setVtEditName(t.name)
    setVtEditQuantity(String(t.quantity))
    setVtEditVehicleId(t.vehicleId)
    setVtEditStatus(t.status)
  }

  const saveVehicleToolEdit = async () => {
    if (!vtEdit) return
    setSubmitting(true)
    try {
      await api.updateVehicleTool(vtEdit.id, {
        name: vtEditName,
        quantity: Number(vtEditQuantity) || 1,
        vehicleId: vtEditVehicleId,
        status: vtEditStatus,
      })
      setVtEdit(null)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    } finally {
      setSubmitting(false)
    }
  }

  const handleDeleteVehicleTool = async (id: string) => {
    if (!confirm('حذف هذي الأداة من السيارة؟')) return
    try {
      await api.deleteVehicleTool(id)
      load()
    } catch (err) {
      alert(err instanceof Error ? err.message : 'حدث خطأ')
    }
  }

  const handleAddStock = async () => {
    if (!stockTarget || !Number(stockQty)) { alert('اكتب الكمية'); return }
    try {
      await api.addStockIntake({
        toolId: stockTarget.id, quantity: Number(stockQty),
        unitPrice: Number(stockPrice) || null, supplier: stockSupplier || null,
      })
      setStockTarget(null); setStockQty(''); setStockPrice(''); setStockSupplier('')
      load()
    } catch (e) { alert(e instanceof Error ? e.message : 'حدث خطأ') }
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
                  {/* الباركود انشال — ما نحتاجه بإضافة الأداة، ويتولّد بالخلفية */}
                  <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
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
                  </div>
                  <div className="mt-4">
                    <button type="submit" disabled={submitting} className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 disabled:opacity-50">
                      {submitting ? 'جاري الحفظ...' : 'إضافة'}
                    </button>
                  </div>
                </form>
              )}
              {/* بطاقة لكل موظف بدل جدول مسطّح بكل الأدوات مطشّرة — تضغط على
                  الموظف تشوف عدته، والنظام يقارنها بالعدة القياسية وينبّه بالنقص. */}
              {!selectedKitEmployeeId ? (
                <>
                  <div className="mb-4 rounded-xl border border-brand-100 bg-brand-50/60 p-4 text-sm text-brand-800">
                    المقارنة تنعمل تلقائياً بين عدة كل موظف والعدة القياسية ({templateItems.length} أداة).
                    اضغط على أي موظف حتى تشوف عدته بالتفصيل وشنو ناقصه.
                  </div>
                  <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
                    {kitSummaries.map((k) => (
                      <button
                        key={k.employee.id}
                        onClick={() => setSelectedKitEmployeeId(k.employee.id)}
                        className={`rounded-xl border-2 bg-white p-5 text-right shadow-[0_4px_20px_rgba(15,32,64,0.06)] transition-all hover:-translate-y-0.5 hover:shadow-lg ${
                          k.missing.length ? 'border-red-200' : 'border-emerald-200'
                        }`}
                      >
                        <div className="flex items-start justify-between gap-2">
                          <div>
                            <div className="font-bold text-brand-900">{k.employee.name}</div>
                            <div className="text-xs text-slate-400">{roleLabels[k.employee.role] || k.employee.role}</div>
                          </div>
                          <span className={`shrink-0 rounded-full px-3 py-1 text-xs font-bold ${
                            k.missing.length ? 'bg-red-100 text-red-700' : 'bg-emerald-100 text-emerald-700'
                          }`}>
                            {k.missing.length ? `ناقص ${k.missing.length}` : 'مكتملة ✓'}
                          </span>
                        </div>
                        <div className="mt-3 text-sm text-slate-600">
                          عدته: <span className="font-bold text-brand-700">{k.tools.length}</span> أداة
                          {k.extra.length > 0 && <span className="text-slate-400"> · خاصة {k.extra.length}</span>}
                        </div>
                        {k.missing.length > 0 && (
                          <div className="mt-2 text-xs text-red-600">
                            ⚠️ ناقص: {k.missing.slice(0, 3).join('، ')}{k.missing.length > 3 ? ` +${k.missing.length - 3}` : ''}
                          </div>
                        )}
                      </button>
                    ))}
                    {kitSummaries.length === 0 && (
                      <div className="col-span-full rounded-xl bg-white p-8 text-center text-slate-400 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                        لا يوجد موظفين
                      </div>
                    )}
                  </div>
                </>
              ) : (() => {
                const k = kitSummaries.find((x) => x.employee.id === selectedKitEmployeeId)
                if (!k) return null
                return (
                  <div className="rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                    <div className="mb-5 flex flex-wrap items-center justify-between gap-3">
                      <div>
                        <h3 className="text-xl font-bold text-brand-900">عدة: {k.employee.name}</h3>
                        <p className="text-sm text-slate-400">{roleLabels[k.employee.role] || k.employee.role}</p>
                      </div>
                      <button
                        onClick={() => setSelectedKitEmployeeId(null)}
                        className="rounded-lg border border-slate-300 px-4 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50"
                      >
                        ← رجوع لكل الموظفين
                      </button>
                    </div>

                    {k.missing.length > 0 ? (
                      <div className="mb-5 overflow-hidden rounded-xl border border-red-200 bg-white">
                        {/* تقرير النواقص: مو كومة أسماء حمر — أرقام محسوبة
                            وجدول يقول لإداري الكميات شنو يكدر ينطي الحين
                            وشنو لازم يشتريه. */}
                        <div className="flex flex-wrap items-center justify-between gap-3 border-b border-red-100 bg-red-50 px-4 py-3">
                          <div className="font-bold text-red-700">⚠️ تقرير نواقص العدة القياسية</div>
                          <div className="flex flex-wrap gap-2 text-xs font-bold">
                            <span className="rounded-full bg-white px-3 py-1 text-slate-600">
                              العدة القياسية: {templateNames.size}
                            </span>
                            <span className="rounded-full bg-emerald-100 px-3 py-1 text-emerald-700">
                              موجود عنده: {templateNames.size - k.missing.length}
                            </span>
                            <span className="rounded-full bg-red-100 px-3 py-1 text-red-700">
                              ناقص: {k.missing.length}
                            </span>
                            <span className="rounded-full bg-brand-50 px-3 py-1 text-brand-700">
                              الاكتمال: {Math.round(((templateNames.size - k.missing.length) / Math.max(1, templateNames.size)) * 100)}%
                            </span>
                          </div>
                        </div>

                        {/* شريط اكتمال — يبيّن الحالة بنظرة وحدة */}
                        <div className="h-2 w-full bg-slate-100">
                          <div
                            className="h-2 bg-emerald-500 transition-all"
                            style={{ width: `${Math.round(((templateNames.size - k.missing.length) / Math.max(1, templateNames.size)) * 100)}%` }}
                          />
                        </div>

                        <table className="min-w-full divide-y divide-slate-100 text-right text-sm">
                          <thead className="bg-slate-50">
                            <tr>
                              <th className="px-4 py-2 font-semibold text-slate-600">الأداة الناقصة</th>
                              <th className="px-4 py-2 font-semibold text-slate-600">متوفرة بمخزن الكميات؟</th>
                              <th className="px-4 py-2 font-semibold text-slate-600">الإجراء</th>
                            </tr>
                          </thead>
                          <tbody className="divide-y divide-slate-100">
                            {k.missing.map((m) => {
                              const inStock = onDemandTools.find((t) => t.name.trim() === m)
                              const qty = inStock?.availableQuantity ?? 0
                              return (
                                <tr key={m}>
                                  <td className="px-4 py-2 font-medium text-slate-700">{m}</td>
                                  <td className="px-4 py-2">
                                    {qty > 0 ? (
                                      <span className="rounded-full bg-emerald-50 px-3 py-1 text-xs font-bold text-emerald-700">
                                        ✔ موجودة — {qty} بالمخزن
                                      </span>
                                    ) : (
                                      <span className="rounded-full bg-red-50 px-3 py-1 text-xs font-bold text-red-700">
                                        ✘ خالصة من المخزن
                                      </span>
                                    )}
                                  </td>
                                  <td className="px-4 py-2 text-xs text-slate-500">
                                    {qty > 0 ? 'ينطيها إداري الكميات من الرف' : 'لازم تنشترى'}
                                  </td>
                                </tr>
                              )
                            })}
                          </tbody>
                        </table>
                      </div>
                    ) : (
                      <div className="mb-5 rounded-xl border border-emerald-200 bg-emerald-50 p-4 font-bold text-emerald-700">
                        ✓ عدة هذا الموظف مكتملة — كل العدة القياسية موجودة عنده
                      </div>
                    )}

                    <div className="overflow-x-auto">
                      <table className="min-w-full divide-y divide-gray-200 text-right">
                        <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                          <tr>
                            <th className="px-4 py-3 text-sm font-semibold">اسم الأداة</th>
                            <th className="px-4 py-3 text-sm font-semibold">العدد</th>
                            <th className="px-4 py-3 text-sm font-semibold">النوع</th>
                            <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                            <th className="px-4 py-3 text-sm font-semibold"></th>
                          </tr>
                        </thead>
                        <tbody className="divide-y divide-slate-100">
                          {groupTools(k.tools).map((g) => {
                            const open = expandedTool === `${k.employee.id}:${g.name}`
                            return (
                              <Fragment key={g.name}>
                                <tr
                                  className="cursor-pointer hover:bg-slate-50"
                                  onClick={() => setExpandedTool(open ? null : `${k.employee.id}:${g.name}`)}
                                >
                                  <td className="px-4 py-3 font-medium">{g.name}</td>
                                  <td className="px-4 py-3">
                                    <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                                      g.units.length > 1 ? 'bg-brand-100 text-brand-800' : 'bg-slate-100 text-slate-600'
                                    }`}>
                                      × {g.units.length}
                                    </span>
                                  </td>
                                  <td className="px-4 py-3">
                                    {templateNames.has(g.name) ? (
                                      <span className="rounded-full bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700">قياسية</span>
                                    ) : (
                                      <span className="rounded-full bg-amber-50 px-3 py-1 text-xs font-bold text-amber-700">أداة خاصة</span>
                                    )}
                                  </td>
                                  <td className="px-4 py-3">
                                    {/* حالة موحّدة لو كل النسخ نفس الحالة، وإلا تفصيل مختصر */}
                                    <div className="flex flex-wrap gap-1">
                                      {g.statuses.map(([st, n]) => (
                                        <span key={st} className={`rounded-full px-3 py-1 text-xs font-bold ${
                                          personalToolStatusColors[st as PersonalToolStatus] || 'bg-slate-100 text-slate-600'
                                        }`}>
                                          {personalToolStatusLabels[st as PersonalToolStatus] || st}
                                          {g.statuses.length > 1 && ` (${n})`}
                                        </span>
                                      ))}
                                    </div>
                                  </td>
                                  <td className="px-4 py-3 text-xs text-slate-400">{open ? '▲ إخفاء' : '▼ تفاصيل'}</td>
                                </tr>
                                {/* التفاصيل: كل نسخة بباركودها — الباركود يميّز
                                    النسخة الواحدة، فما ينفع يتجمّع بالصف */}
                                {open && g.units.map((t) => (
                                  <tr key={t.id} className="bg-slate-50/60 text-sm">
                                    <td className="px-8 py-2 text-slate-400">↳</td>
                                    <td className="px-4 py-2 font-mono text-xs text-slate-500" colSpan={2}>{t.barcode}</td>
                                    <td className="px-4 py-2">
                                      <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                                        personalToolStatusColors[t.status as PersonalToolStatus] || 'bg-slate-100 text-slate-600'
                                      }`}>
                                        {personalToolStatusLabels[t.status as PersonalToolStatus] || t.status}
                                      </span>
                                    </td>
                                    <td className="px-4 py-2">
                                      {canManageInventory && (
                                        <div className="flex flex-wrap gap-1">
                                          <button onClick={() => openToolEdit(t)} className="rounded-lg bg-brand-50 px-2.5 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100">✎ تعديل</button>
                                          <button onClick={() => openToolHistory(t)} className="rounded-lg bg-slate-100 px-2.5 py-1 text-xs font-bold text-slate-600 hover:bg-slate-200">🕘 السجل</button>
                                          <button onClick={() => handleDeleteTool(t)} className="rounded-lg bg-red-50 px-2.5 py-1 text-xs font-bold text-red-600 hover:bg-red-100">🗑</button>
                                        </div>
                                      )}
                                    </td>
                                  </tr>
                                ))}
                              </Fragment>
                            )
                          })}
                          {k.tools.length === 0 && (
                            <tr><td colSpan={5} className="px-4 py-6 text-center text-slate-400">ما عنده أي أداة مسجلة</td></tr>
                          )}
                        </tbody>
                      </table>
                    </div>
                  </div>
                )
              })()}
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
                      <label className="mb-1 block text-sm font-medium text-slate-600">السيارة</label>
                      <select required value={vtVehicleId} onChange={(e) => setVtVehicleId(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500">
                        <option value="">اختر السيارة</option>
                        {vehicles.map((v) => (
                          <option key={v.id} value={v.id}>{v.name} — {v.plateNumber}</option>
                        ))}
                      </select>
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">اسم الأداة</label>
                      <input required value={vtName} onChange={(e) => setVtName(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
                    </div>
                    <div>
                      {/* الكمية بدل الباركود — نفس الأداة ممكن تتكرر بنفس السيارة */}
                      <label className="mb-1 block text-sm font-medium text-slate-600">الكمية</label>
                      <input required type="number" min="1" value={vtQuantity} onChange={(e) => setVtQuantity(e.target.value)} className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
                    </div>
                  </div>
                  <div className="mt-4">
                    <button type="submit" disabled={submitting} className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 disabled:opacity-50">
                      {submitting ? 'جاري الحفظ...' : 'إضافة'}
                    </button>
                  </div>
                </form>
              )}
              {/* إحصائيات سريعة: كم أداة، كم قطعة إجمالاً، وكم مفقودة/تالفة */}
              <div className="mb-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
                {[
                  { label: 'أنواع الأدوات', value: vehicleToolStats.kinds, color: 'text-brand-700' },
                  { label: 'إجمالي القطع', value: vehicleToolStats.pieces, color: 'text-brand-700' },
                  { label: 'السيارات المجهّزة', value: vehicleToolStats.vehicles, color: 'text-emerald-600' },
                  { label: 'مفقودة / تالفة', value: vehicleToolStats.problem, color: 'text-red-600' },
                ].map((s) => (
                  <div key={s.label} className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                    <div className="text-xs text-slate-400">{s.label}</div>
                    <div className={`mt-1 text-2xl font-bold ${s.color}`}>{s.value}</div>
                  </div>
                ))}
              </div>

              <div className="overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                <div className="overflow-x-auto">
                  <table className="min-w-full divide-y divide-gray-200 text-right">
                    <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
                      <tr>
                        <th className="px-4 py-3 text-sm font-semibold">السيارة</th>
                        <th className="px-4 py-3 text-sm font-semibold">اسم الأداة</th>
                        <th className="px-4 py-3 text-sm font-semibold">الكمية</th>
                        <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                        <th className="px-4 py-3 text-sm font-semibold">تاريخ الإضافة</th>
                        {canManageInventory && <th className="px-4 py-3 text-sm font-semibold">إجراءات</th>}
                      </tr>
                    </thead>
                    <tbody className="divide-y divide-slate-100">
                      {vehicleTools.map((t) => (
                        <tr key={t.id} className="transition-colors hover:bg-slate-50">
                          <td className="px-4 py-3 font-medium">
                            {t.vehicleName || '-'}
                            {t.vehiclePlate && <span className="mr-1 text-xs text-slate-400">({t.vehiclePlate})</span>}
                          </td>
                          <td className="px-4 py-3">{t.name}</td>
                          <td className="px-4 py-3 font-bold text-brand-700">{t.quantity}</td>
                          <td className="px-4 py-3">
                            <span className={`rounded-full px-3 py-1 text-xs font-bold ${
                              t.status === 'AVAILABLE' ? 'bg-emerald-50 text-emerald-700' : 'bg-red-50 text-red-700'
                            }`}>{vehicleToolStatusLabels[t.status] || t.status}</span>
                          </td>
                          <td className="px-4 py-3 text-sm text-slate-500">{new Date(t.createdAt).toLocaleDateString('ar-IQ')}</td>
                          {canManageInventory && (
                            <td className="px-4 py-3">
                              <div className="flex gap-2">
                                <button
                                  onClick={() => openVehicleToolEdit(t)}
                                  className="rounded-lg bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100"
                                >
                                  ✎ تعديل
                                </button>
                                <button
                                  onClick={() => handleDeleteVehicleTool(t.id)}
                                  className="rounded-lg bg-red-50 px-3 py-1 text-xs font-bold text-red-600 hover:bg-red-100"
                                >
                                  🗑 حذف
                                </button>
                              </div>
                            </td>
                          )}
                        </tr>
                      ))}
                      {vehicleTools.length === 0 && (
                        <tr><td colSpan={canManageInventory ? 6 : 5} className="px-4 py-6 text-center text-slate-400">لا توجد أدوات مركبات</td></tr>
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
                              {/* إضافة كمية — عملية لها أثر مسجّل (منو أضاف وشكد ومتى)
                                  بدل ما نعدّل الرقم يدوياً بلا سجل */}
                              <button
                                type="button"
                                onClick={() => { setStockTarget(t); setStockQty(''); setStockPrice(''); setStockSupplier('') }}
                                className="mr-2 rounded-lg bg-emerald-600 px-3 py-1 text-xs font-semibold text-white hover:bg-emerald-700"
                              >
                                ➕ إضافة كمية
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
                              <div className="flex flex-wrap items-center gap-2">
                                <div className="font-medium text-brand-800">{r.reasonLabel}</div>
                                {/* السلة: تخصصية / بدل مفقود / بدل تالف */}
                                {r.kindLabel && (
                                  <span className={`rounded-full px-2.5 py-0.5 text-xs font-bold ${
                                    r.requestKind === 'REPLACE_LOST' ? 'bg-red-100 text-red-800'
                                    : r.requestKind === 'REPLACE_DAMAGED' ? 'bg-amber-100 text-amber-800'
                                    : 'bg-sky-100 text-sky-800'}`}>
                                    {r.kindLabel}
                                  </span>
                                )}
                              </div>
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
                <form onSubmit={handleAddTemplateItem} className="mb-6 rounded-xl border border-white bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                  <p className="mb-3 text-sm font-semibold text-slate-700">إضافة أداة</p>
                  {/* السؤال الأول: قياسية لو خاصة */}
                  <div className="mb-4 flex flex-wrap gap-3">
                    <button
                      type="button"
                      onClick={() => { setAddToolKind('standard'); setAddToolEmployeeId('') }}
                      className={`rounded-lg border px-5 py-3 text-sm font-medium transition-colors ${
                        addToolKind === 'standard'
                          ? 'border-brand-500 bg-brand-50 text-brand-800'
                          : 'border-gray-300 text-slate-600 hover:bg-slate-50'
                      }`}
                    >
                      أداة قياسية
                      <span className="mt-0.5 block text-xs font-normal opacity-70">تنضاف لعدة كل موظف مستحق</span>
                    </button>
                    <button
                      type="button"
                      onClick={() => setAddToolKind('private')}
                      className={`rounded-lg border px-5 py-3 text-sm font-medium transition-colors ${
                        addToolKind === 'private'
                          ? 'border-brand-500 bg-brand-50 text-brand-800'
                          : 'border-gray-300 text-slate-600 hover:bg-slate-50'
                      }`}
                    >
                      أداة خاصة
                      <span className="mt-0.5 block text-xs font-normal opacity-70">لموظف واحد تختاره</span>
                    </button>
                  </div>
                  <div className={`grid grid-cols-1 gap-4 ${addToolKind === 'private' ? 'sm:grid-cols-2' : ''}`}>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">اسم الأداة</label>
                      <input
                        required
                        value={templateName}
                        onChange={(e) => setTemplateName(e.target.value)}
                        placeholder={addToolKind === 'standard' ? 'اسم الأداة القياسية' : 'اسم الأداة الخاصة'}
                        className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
                      />
                    </div>
                    {addToolKind === 'private' && (
                      <div>
                        <label className="mb-1 block text-sm font-medium text-slate-600">الموظف الي عنده هاي الأداة</label>
                        <select
                          required
                          value={addToolEmployeeId}
                          onChange={(e) => setAddToolEmployeeId(e.target.value)}
                          className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
                        >
                          <option value="">اختر الموظف</option>
                          {employees.map((emp) => <option key={emp.id} value={emp.id}>{emp.name}</option>)}
                        </select>
                      </div>
                    )}
                  </div>
                  <button type="submit" disabled={templateSubmitting} className="mt-4 rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3 font-medium text-white shadow-md shadow-brand-900/20 disabled:opacity-50">
                    {templateSubmitting ? 'جاري الإضافة...' : 'إضافة'}
                  </button>
                </form>
              )}
              <p className="mb-4 text-sm text-slate-500">
                العدة القياسية تخص <span className="font-semibold text-slate-700">الفنيين والليدرات</span> بس — أي أداة قياسية تُضاف هنا تنضاف فوراً لعدتهم، وأي فني أو ليدر جديد ياخذ العدة كاملة تلقائياً وقت إنشاء حسابه. المبيعات والتصميم وإدارة المشاريع وبقية الإداريين ما عندهم عدة وما يتحاسبون عليها.
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

      {toolEdit && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" dir="rtl">
          <div className="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-brand-900">تعديل أداة</h3>
            <p className="mt-1 text-sm text-slate-400">{toolEdit.employee?.name || ''}</p>

            <label className="mt-5 block text-sm font-semibold text-slate-700">اسم الأداة</label>
            <input value={toolEditName} onChange={(e) => setToolEditName(e.target.value)} className="mt-1.5 w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500" />

            <label className="mt-4 block text-sm font-semibold text-slate-700">الباركود</label>
            <input value={toolEditBarcode} onChange={(e) => setToolEditBarcode(e.target.value)} className="mt-1.5 w-full rounded-xl border border-slate-300 px-4 py-2.5 font-mono outline-none focus:border-brand-500" />

            <label className="mt-4 block text-sm font-semibold text-slate-700">الحالة</label>
            <select value={toolEditStatus} onChange={(e) => setToolEditStatus(e.target.value as PersonalToolStatus)} className="mt-1.5 w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500">
              {(Object.keys(personalToolStatusLabels) as PersonalToolStatus[]).map((k) => (
                <option key={k} value={k}>{personalToolStatusLabels[k]}</option>
              ))}
            </select>

            <label className="mt-4 block text-sm font-semibold text-slate-700">
              ملاحظة <span className="font-normal text-slate-400">(تنحفظ بسجل الحركة — مفيدة لتوثيق سبب الفقدان)</span>
            </label>
            <input value={toolEditNote} onChange={(e) => setToolEditNote(e.target.value)} placeholder="مثال: ضاعت بموقع الحسينية يوم الخميس" className="mt-1.5 w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500" />

            <div className="mt-5 flex gap-3">
              <button onClick={saveToolEdit} disabled={toolSaving} className="flex-1 rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 py-2.5 font-bold text-white disabled:opacity-50">
                {toolSaving ? 'جاري الحفظ...' : 'حفظ التعديل'}
              </button>
              <button onClick={() => setToolEdit(null)} className="rounded-xl border border-slate-300 px-6 py-2.5 font-medium text-slate-600 hover:bg-slate-50">إلغاء</button>
            </div>
          </div>
        </div>
      )}

      {historyTool && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" dir="rtl">
          <div className="max-h-[85vh] w-full max-w-2xl overflow-auto rounded-2xl bg-white p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-brand-900">سجل حركة: {historyTool.name}</h3>
            <p className="mt-1 text-sm text-slate-400">{historyTool.employee?.name || ''}</p>

            {historyLoading ? (
              <p className="mt-6 text-slate-400">جاري التحميل...</p>
            ) : toolEvents.length === 0 ? (
              <p className="mt-6 rounded-xl bg-slate-50 p-4 text-center text-slate-400">ما اكو حركات مسجلة لهذي الأداة</p>
            ) : (
              <div className="mt-5 space-y-3">
                {toolEvents.map((ev) => {
                  const lost = ev.toStatus === 'LOST'
                  return (
                    <div key={ev.id} className={`rounded-xl border p-4 ${lost ? 'border-red-200 bg-red-50' : 'border-slate-100 bg-slate-50/60'}`}>
                      <div className="flex flex-wrap items-center justify-between gap-2">
                        <span className={`font-bold ${lost ? 'text-red-700' : 'text-brand-800'}`}>
                          {lost ? '⚠️ انفقدت' : ev.eventLabel}
                        </span>
                        <span className="text-xs text-slate-500">
                          {new Date(ev.createdAt).toLocaleString('ar-IQ')}
                        </span>
                      </div>
                      {ev.eventType === 'STATUS_CHANGED' && (
                        <div className="mt-1 text-sm text-slate-600">
                          {ev.fromStatusText} ← <span className="font-bold">{ev.toStatusText}</span>
                        </div>
                      )}
                      {ev.note && <div className="mt-1 text-sm text-slate-600">{ev.note}</div>}
                      <div className="mt-1 text-xs text-slate-400">
                        سجّلها: {ev.actorName || 'غير محدد'}
                      </div>
                    </div>
                  )
                })}
              </div>
            )}

            <button onClick={() => setHistoryTool(null)} className="mt-6 w-full rounded-xl border border-slate-300 py-2.5 font-medium text-slate-600 hover:bg-slate-50">إغلاق</button>
          </div>
        </div>
      )}

      {vtEdit && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" dir="rtl">
          <div className="w-full max-w-lg rounded-2xl bg-white p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-brand-900">تعديل أداة مركبة</h3>

            <label className="mt-5 block text-sm font-semibold text-slate-700">السيارة</label>
            <select
              value={vtEditVehicleId}
              onChange={(e) => setVtEditVehicleId(e.target.value)}
              className="mt-1.5 w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500"
            >
              {vehicles.map((v) => <option key={v.id} value={v.id}>{v.name} — {v.plateNumber}</option>)}
            </select>

            <label className="mt-4 block text-sm font-semibold text-slate-700">اسم الأداة</label>
            <input
              value={vtEditName}
              onChange={(e) => setVtEditName(e.target.value)}
              className="mt-1.5 w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500"
            />

            <div className="mt-4 grid grid-cols-2 gap-4">
              <div>
                <label className="block text-sm font-semibold text-slate-700">الكمية</label>
                <input
                  type="number" min="1"
                  value={vtEditQuantity}
                  onChange={(e) => setVtEditQuantity(e.target.value)}
                  className="mt-1.5 w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500"
                />
              </div>
              <div>
                <label className="block text-sm font-semibold text-slate-700">الحالة</label>
                <select
                  value={vtEditStatus}
                  onChange={(e) => setVtEditStatus(e.target.value as VehicleTool['status'])}
                  className="mt-1.5 w-full rounded-xl border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500"
                >
                  {Object.keys(vehicleToolStatusLabels).map((k) => (
                    <option key={k} value={k}>{vehicleToolStatusLabels[k]}</option>
                  ))}
                </select>
              </div>
            </div>

            <div className="mt-5 flex gap-3">
              <button
                onClick={saveVehicleToolEdit}
                disabled={submitting}
                className="flex-1 rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 py-2.5 font-bold text-white disabled:opacity-50"
              >
                {submitting ? 'جاري الحفظ...' : 'حفظ التعديل'}
              </button>
              <button
                onClick={() => setVtEdit(null)}
                className="rounded-xl border border-slate-300 px-6 py-2.5 font-medium text-slate-600 hover:bg-slate-50"
              >
                إلغاء
              </button>
            </div>
          </div>
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
      {stockTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4" onClick={() => setStockTarget(null)}>
          <div dir="rtl" className="w-full max-w-md rounded-2xl bg-white p-6 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <h3 className="text-lg font-bold" style={{ color: '#1a3a5c' }}>إضافة كمية</h3>
            <p className="mt-1 text-sm text-slate-600">
              {stockTarget.name} — الكمية الحالية: <span className="font-bold">{stockTarget.availableQuantity}</span> من {stockTarget.totalQuantity}
            </p>
            <label className="mt-4 mb-1 block text-sm font-medium text-slate-600">الكمية المضافة *</label>
            <input type="number" value={stockQty} onChange={(e) => setStockQty(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
            <label className="mt-3 mb-1 block text-sm font-medium text-slate-600">سعر القطعة</label>
            <input type="number" value={stockPrice} onChange={(e) => setStockPrice(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
            <label className="mt-3 mb-1 block text-sm font-medium text-slate-600">المورد</label>
            <input value={stockSupplier} onChange={(e) => setStockSupplier(e.target.value)}
              className="w-full rounded-lg border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500" />
            <div className="mt-4 flex gap-3">
              <button onClick={handleAddStock} className="flex-1 rounded-lg px-4 py-3 font-medium text-white" style={{ backgroundColor: '#1a3a5c' }}>إضافة</button>
              <button onClick={() => setStockTarget(null)} className="rounded-lg border border-slate-300 px-4 py-3 font-medium text-slate-700">إلغاء</button>
            </div>
          </div>
        </div>
      )}

    </div>
  )
}
