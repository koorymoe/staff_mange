import { Fragment, useState, useEffect, useContext } from 'react'
import { useSearchParams } from 'react-router-dom'
import { api, procurementTypeLabels, type ProcurementRequest, type ProcurementRequestType, type ProcurementStats, type Booking } from '../api'
import { SessionContext } from '../session'
import BookingCodeChip from '../components/BookingCodeChip'

const statusLabels: Record<string, string> = {
  PENDING: 'بانتظار التوفير',
  IN_PROGRESS: 'جاري التوفير',
  FULFILLED: 'تم التوفير',
  REJECTED: 'مرفوض',
}
const statusColors: Record<string, string> = {
  PENDING: 'bg-amber-100 text-amber-700',
  IN_PROGRESS: 'bg-blue-100 text-blue-700',
  FULFILLED: 'bg-emerald-100 text-emerald-700',
  REJECTED: 'bg-red-100 text-red-700',
}

export default function ProcurementPage() {
  const { employee, permissions } = useContext(SessionContext)
  // توفير المواد وتحديد حالتها (توفير/رفض) يقتصر على إداري الكميات والأدمن فقط —
  // أي موظف ثاني (فني، مدير مشاريع، مراقب...) يشوف الطلبات وحالتها بس بدون تحكم.
  const canManageProcurement = employee?.role === 'ADMIN' || employee?.role === 'OWNER' || employee?.role === 'PROCUREMENT_ADMIN'
  const isAdminLike = employee?.role === 'ADMIN' || employee?.role === 'OWNER'
  // كل نوع طلب له صلاحية مستقلة تُمنح يدوياً بصفحة الصلاحيات — ما تنجر مع أي دور.
  // الأدمن/المالك يقدر ينشئ الاثنين دائماً.
  const canCreatePersonal = isAdminLike || permissions.includes('procurement_personal')
  const canCreateCustomer = isAdminLike || permissions.includes('procurement_customer')
  // الطلب اليدوي: الإداري أو الليدر يكتب المادة بالتفصيل بدون حجز ولا كتالوك
  const canCreateManual = isAdminLike || permissions.includes('procurement_manual')
  const [searchParams] = useSearchParams()

  const [requests, setRequests] = useState<ProcurementRequest[]>([])
  const [stats, setStats] = useState<ProcurementStats | null>(null)
  const [bookings, setBookings] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [activeTab, setActiveTab] = useState<ProcurementRequestType>('CUSTOMER_PRODUCT')
  const [suppliers, setSuppliers] = useState<{ id: string; companyName: string; ownerName: string }[]>([])

  // Form state
  const [showForm, setShowForm] = useState(false)
  const [formRequestType, setFormRequestType] = useState<ProcurementRequestType>('CUSTOMER_PRODUCT')
  const [formBookingId, setFormBookingId] = useState('')
  const [formNotes, setFormNotes] = useState('')
  type FormItem = { productName: string; quantity: number; model: string; spec: string; unitPrice: string; itemNotes: string }
  const emptyItem = (): FormItem => ({ productName: '', quantity: 1, model: '', spec: '', unitPrice: '', itemNotes: '' })
  const [formItems, setFormItems] = useState<FormItem[]>([emptyItem()])
  const [submitting, setSubmitting] = useState(false)

  // Expanded rows
  const [expandedRows, setExpandedRows] = useState<Set<string>>(new Set())

  // Fulfill modal
  const [fulfillModal, setFulfillModal] = useState<ProcurementRequest | null>(null)
  const [fulfillCost, setFulfillCost] = useState('')
  const [fulfillNotes, setFulfillNotes] = useState('')
  const [fulfillSupplierId, setFulfillSupplierId] = useState('')
  const [fulfillItems, setFulfillItems] = useState<{ id: string; productName: string; quantity: number; unitPrice: string; totalPrice: string; fulfilled: boolean }[]>([])
  const [fulfilling, setFulfilling] = useState(false)

  const loadData = async () => {
    try {
      const [reqs, st, bks, sups] = await Promise.all([
        api.getProcurementRequests(),
        api.getProcurementStats(),
        // القائمة المنسدلة تربط الطلب بحجز شغّال — ماكو معنى نعرض
        // حجوزات منجزة وملغاة من سنين، ولا نسحب الأرشيف كله عشانها.
        api.getBookings({ status: ['PENDING', 'CONFIRMED', 'IN_PROGRESS'] }),
        // الموردين المضافين مسبقاً — أبو الكميات لازم يحدد منهم
        api.getSupplierOptions().catch(() => []),
      ])
      setRequests(reqs)
      setStats(st)
      setBookings(bks)
      setSuppliers(sups ?? [])
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ في تحميل البيانات')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    // `loadData` is async and only sets state after its awaits resolve; false positive.
    // eslint-disable-next-line react-hooks/set-state-in-effect
    loadData()
  }, [])

  // إذا وصلنا من زر "اطلب مادة" بشاشة الحجز (?bookingId=...) نفتح فورم الطلب
  // ونعبّي الحجز تلقائياً بدل ما يدوّر عليه المستخدم من القائمة المنسدلة،
  // ونحدد النوع تلقائياً "طلب للزبون" لأنه مربوط بحجز.
  useEffect(() => {
    const bookingId = searchParams.get('bookingId')
    // Pre-filling the request form from the ?bookingId= URL param is a derived-state
    // sync from the router, not a fetch.
    if (bookingId) {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setFormBookingId(bookingId)
      setFormRequestType('CUSTOMER_PRODUCT')
      setActiveTab('CUSTOMER_PRODUCT')
      setShowForm(true)
    }
  }, [searchParams])

  // عند فتح الفورم أول مرة، إذا الموظف عنده صلاحية وحدة بس، نحددها تلقائياً.
  useEffect(() => {
    if (!showForm) return
    // Auto-selecting the request type when the employee only has one permission is a
    // derived-state sync from props, not a fetch.
    const allowed: ProcurementRequestType[] = []
    if (canCreatePersonal) allowed.push('PERSONAL_SUPPLY')
    if (canCreateCustomer) allowed.push('CUSTOMER_PRODUCT')
    if (canCreateManual) allowed.push('MANUAL_SUPPLY')
    // eslint-disable-next-line react-hooks/set-state-in-effect
    if (allowed.length === 1) setFormRequestType(allowed[0])
  }, [showForm, canCreatePersonal, canCreateCustomer, canCreateManual])

  // أنواع الطلبات المسموحة لهذا الموظف — تحدد الراديو وتبويب الجدول
  const allowedTypes: ProcurementRequestType[] = []
  if (canCreatePersonal) allowedTypes.push('PERSONAL_SUPPLY')
  if (canCreateCustomer) allowedTypes.push('CUSTOMER_PRODUCT')
  if (canCreateManual) allowedTypes.push('MANUAL_SUPPLY')

  const handleSubmit = async () => {
    if (formItems.some(i => !i.productName.trim() || i.quantity < 1)) {
      setError('يرجى ملء جميع بيانات المواد')
      return
    }
    if (!employee) return
    setSubmitting(true)
    setError('')
    try {
      await api.createProcurementRequest({
        requestedById: employee.id,
        bookingId: formRequestType === 'CUSTOMER_PRODUCT' ? (formBookingId || undefined) : undefined,
        requestType: formRequestType,
        notes: formNotes || undefined,
        items: formItems.map(i => ({
          productName: i.productName.trim(),
          quantity: i.quantity,
          // السعر التقريبي الي يكتبه الطالب — أبو الكميات يصحّحه وقت الشراء
          unitPrice: i.unitPrice ? parseFloat(i.unitPrice) : null,
          model: i.model.trim() || null,
          spec: i.spec.trim() || null,
          itemNotes: i.itemNotes.trim() || null,
        })),
      })
      setFormBookingId('')
      setFormNotes('')
      setFormItems([emptyItem()])
      setShowForm(false)
      await loadData()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ في إنشاء الطلب')
    } finally {
      setSubmitting(false)
    }
  }

  const handleStatusUpdate = async (id: string, status: string) => {
    try {
      await api.updateProcurementStatus(id, status)
      await loadData()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ في تحديث الحالة')
    }
  }

  const openFulfillModal = (req: ProcurementRequest) => {
    setFulfillModal(req)
    setFulfillCost(req.totalCost?.toString() || '')
    setFulfillNotes(req.fulfillmentNotes || '')
    setFulfillSupplierId(req.supplierId || '')
    setFulfillItems(req.items.map(i => ({
      id: i.id,
      productName: i.productName,
      quantity: i.quantity,
      unitPrice: i.unitPrice?.toString() || '',
      totalPrice: i.totalPrice?.toString() || '',
      fulfilled: i.fulfilled,
    })))
  }

  const handleFulfill = async () => {
    if (!fulfillModal || !employee) return
    // بدون مورد ما نعرف من وين انجابت المادة — الباك إند يرفض أصلاً،
    // بس نوقفها هنا حتى ما يضيع شغل المستخدم بطلب فاشل.
    if (!fulfillSupplierId) {
      setError('لازم تحدد المورد الي انجابت منه المادة')
      return
    }
    setFulfilling(true)
    setError('')
    try {
      await api.fulfillProcurementRequest(fulfillModal.id, {
        fulfilledById: employee.id,
        supplierId: fulfillSupplierId,
        totalCost: fulfillCost ? parseFloat(fulfillCost) : undefined,
        fulfillmentNotes: fulfillNotes || undefined,
        items: fulfillItems.map(i => ({
          id: i.id,
          unitPrice: i.unitPrice ? parseFloat(i.unitPrice) : undefined,
          totalPrice: i.totalPrice ? parseFloat(i.totalPrice) : undefined,
          fulfilled: i.fulfilled,
        })),
      })
      setFulfillModal(null)
      await loadData()
    } catch (e) {
      setError(e instanceof Error ? e.message : 'حدث خطأ في تنفيذ الطلب')
    } finally {
      setFulfilling(false)
    }
  }

  const toggleRow = (id: string) => {
    setExpandedRows(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const fmt = (n: number | null | undefined) =>
    n != null ? n.toLocaleString('ar-IQ', { minimumFractionDigits: 2, maximumFractionDigits: 2 }) : '—'

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[#2c5aad]" />
      </div>
    )
  }

  return (
    <div dir="rtl" className="p-6 space-y-6">
      <h1 className="text-2xl font-bold text-[#0f2040]">طلبات المواد والمشتريات</h1>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 rounded-lg p-3 text-sm">{error}</div>
      )}

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
          {[
            { label: 'إجمالي المصروفات', value: `${fmt(stats.totalSpent)} د.ع`, color: 'text-[#2c5aad]' },
            { label: 'طلبات معلقة', value: stats.pendingCount.toString(), color: 'text-amber-600' },
            { label: 'تم التوفير', value: stats.fulfilledCount.toString(), color: 'text-emerald-600' },
            { label: 'مصروفات الشهر', value: `${fmt(stats.monthlySpent)} د.ع`, color: 'text-[#2c5aad]' },
          ].map((card, i) => (
            <div key={i} className="rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)] p-5">
              <p className="text-sm text-slate-500 mb-1">{card.label}</p>
              <p className={`text-2xl font-bold ${card.color}`}>{card.value}</p>
            </div>
          ))}
        </div>
      )}

      {/* New Request Form */}
      <div className="rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <button
          onClick={() => setShowForm(!showForm)}
          className="w-full flex items-center justify-between p-4 text-right font-semibold text-[#0f2040] hover:bg-slate-50 rounded-xl transition"
        >
          <span>+ طلب جديد</span>
          <svg className={`w-5 h-5 transition-transform ${showForm ? 'rotate-180' : ''}`} fill="none" viewBox="0 0 24 24" stroke="currentColor"><path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" /></svg>
        </button>
        {showForm && (
          <div className="p-4 pt-0 space-y-4 border-t">
            {allowedTypes.length > 1 && (
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">نوع الطلب</label>
                <div className="flex flex-wrap gap-4">
                  {allowedTypes.map(t => (
                    <label key={t} className="flex items-center gap-1.5 text-sm">
                      <input
                        type="radio"
                        name="requestType"
                        checked={formRequestType === t}
                        onChange={() => setFormRequestType(t)}
                      />
                      {procurementTypeLabels[t]}
                    </label>
                  ))}
                </div>
              </div>
            )}
            {allowedTypes.length === 0 && (
              <p className="text-sm text-red-600">لا تملك صلاحية تقديم أي نوع من طلبات المشتريات.</p>
            )}

            {formRequestType === 'CUSTOMER_PRODUCT' && (
              <div>
                <label className="block text-sm font-medium text-slate-700 mb-1">الحجز (اختياري)</label>
                <select
                  value={formBookingId}
                  onChange={e => setFormBookingId(e.target.value)}
                  className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
                >
                  <option value="">بدون حجز</option>
                  {bookings.map(b => (
                    <option key={b.id} value={b.id}>
                      <BookingCodeChip code={b.code} /> - {b.customer?.name || ''}
                    </option>
                  ))}
                </select>
              </div>
            )}

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">المواد المطلوبة</label>
              <div className="space-y-2">
                {formItems.map((item, idx) => {
                  const patch = (p: Partial<FormItem>) => {
                    const next = [...formItems]
                    next[idx] = { ...next[idx], ...p }
                    setFormItems(next)
                  }
                  return (
                    <div key={idx} className="rounded-lg border border-slate-200 p-3">
                      <div className="flex items-center gap-2">
                        <input
                          type="text"
                          placeholder="اسم المادة"
                          value={item.productName}
                          onChange={e => patch({ productName: e.target.value })}
                          className="flex-1 border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
                        />
                        <input
                          type="number"
                          min={1}
                          placeholder="الكمية"
                          value={item.quantity}
                          onChange={e => patch({ quantity: parseInt(e.target.value) || 1 })}
                          className="w-24 border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
                        />
                        {formItems.length > 1 && (
                          <button
                            onClick={() => setFormItems(formItems.filter((_, i) => i !== idx))}
                            className="text-red-500 hover:text-red-700 text-xl leading-none"
                          >×</button>
                        )}
                      </div>
                      {/* تفاصيل الطلب اليدوي — الطالب يحدد بالضبط شنو يريد */}
                      {formRequestType === 'MANUAL_SUPPLY' && (
                        <div className="mt-2 grid grid-cols-1 gap-2 sm:grid-cols-3">
                          <input
                            type="text"
                            placeholder="الموديل"
                            value={item.model}
                            onChange={e => patch({ model: e.target.value })}
                            className="border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
                          />
                          <input
                            type="text"
                            placeholder="النوعية / المواصفات"
                            value={item.spec}
                            onChange={e => patch({ spec: e.target.value })}
                            className="border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
                          />
                          <input
                            type="number"
                            step="0.01"
                            placeholder="السعر التقريبي للوحدة"
                            value={item.unitPrice}
                            onChange={e => patch({ unitPrice: e.target.value })}
                            className="border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
                          />
                          <input
                            type="text"
                            placeholder="ملاحظات المادة (ترتيب، لون، بديل مقبول...)"
                            value={item.itemNotes}
                            onChange={e => patch({ itemNotes: e.target.value })}
                            className="sm:col-span-3 border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
                          />
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
              <button
                onClick={() => setFormItems([...formItems, emptyItem()])}
                className="mt-2 text-sm text-[#2c5aad] hover:underline"
              >+ إضافة مادة</button>
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">ملاحظات</label>
              <textarea
                value={formNotes}
                onChange={e => setFormNotes(e.target.value)}
                rows={2}
                className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
              />
            </div>

            <button
              onClick={handleSubmit}
              disabled={submitting || allowedTypes.length === 0}
              className="bg-[#2c5aad] text-white px-6 py-2 rounded-lg text-sm font-medium hover:bg-[#1a3a6e] disabled:opacity-50 transition"
            >
              {submitting ? 'جاري الإرسال...' : 'إرسال الطلب'}
            </button>
          </div>
        )}
      </div>

      {/* Tabs */}
      <div className="flex gap-2 border-b border-slate-200">
        <button
          onClick={() => setActiveTab('CUSTOMER_PRODUCT')}
          className={`px-4 py-2 text-sm font-semibold border-b-2 transition ${activeTab === 'CUSTOMER_PRODUCT' ? 'border-[#2c5aad] text-[#2c5aad]' : 'border-transparent text-slate-500 hover:text-slate-700'}`}
        >
          طلبات منتجات الزبائن
        </button>
        <button
          onClick={() => setActiveTab('PERSONAL_SUPPLY')}
          className={`px-4 py-2 text-sm font-semibold border-b-2 transition ${activeTab === 'PERSONAL_SUPPLY' ? 'border-[#2c5aad] text-[#2c5aad]' : 'border-transparent text-slate-500 hover:text-slate-700'}`}
        >
          طلبات احتياجات شخصية
        </button>
        <button
          onClick={() => setActiveTab('MANUAL_SUPPLY')}
          className={`px-4 py-2 text-sm font-semibold border-b-2 transition ${activeTab === 'MANUAL_SUPPLY' ? 'border-[#2c5aad] text-[#2c5aad]' : 'border-transparent text-slate-500 hover:text-slate-700'}`}
        >
          طلبات مواد يدوية
        </button>
      </div>

      {/* Requests Table */}
      <div className="rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)] overflow-hidden">
        <div className="overflow-x-auto">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-gradient-to-l from-[#2c5aad] to-[#1a3a6e] text-white">
                <th className="px-4 py-3 text-right font-medium">الكود</th>
                <th className="px-4 py-3 text-right font-medium">مقدم الطلب</th>
                {activeTab === 'CUSTOMER_PRODUCT' && <th className="px-4 py-3 text-right font-medium">الحجز</th>}
                <th className="px-4 py-3 text-right font-medium">الحالة</th>
                <th className="px-4 py-3 text-right font-medium">التكلفة</th>
                <th className="px-4 py-3 text-right font-medium">التاريخ</th>
                <th className="px-4 py-3 text-right font-medium">إجراءات</th>
              </tr>
            </thead>
            <tbody>
              {requests.filter(r => r.requestType === activeTab).length === 0 && (
                <tr><td colSpan={activeTab === 'CUSTOMER_PRODUCT' ? 7 : 6} className="text-center py-8 text-slate-400">لا توجد طلبات</td></tr>
              )}
              {requests.filter(r => r.requestType === activeTab).map(req => (
                <Fragment key={req.id}>
                  <tr className="border-b border-slate-100 hover:bg-slate-50 cursor-pointer" onClick={() => toggleRow(req.id)}>
                    <td className="px-4 py-3 font-mono text-xs">{req.code}</td>
                    <td className="px-4 py-3">{req.requestedBy.name}</td>
                    {activeTab === 'CUSTOMER_PRODUCT' && (
                      <td className="px-4 py-3">{req.booking ? `${req.booking.code} - ${req.booking.customer?.name}` : '—'}</td>
                    )}
                    <td className="px-4 py-3">
                      <span className={`px-2 py-1 rounded-full text-xs font-medium ${statusColors[req.status]}`}>
                        {statusLabels[req.status]}
                      </span>
                    </td>
                    <td className="px-4 py-3">{fmt(req.totalCost)}</td>
                    <td className="px-4 py-3 text-xs text-slate-500">{new Date(req.createdAt).toLocaleDateString('ar-IQ')}</td>
                    <td className="px-4 py-3">
                      {canManageProcurement ? (
                        <div className="flex gap-1 flex-wrap" onClick={e => e.stopPropagation()}>
                          {req.status === 'PENDING' && (
                            <>
                              <button onClick={() => handleStatusUpdate(req.id, 'IN_PROGRESS')} className="px-2 py-1 text-xs bg-blue-100 text-blue-700 rounded hover:bg-blue-200 transition">جاري التوفير</button>
                              <button onClick={() => handleStatusUpdate(req.id, 'REJECTED')} className="px-2 py-1 text-xs bg-red-100 text-red-700 rounded hover:bg-red-200 transition">رفض</button>
                            </>
                          )}
                          {(req.status === 'PENDING' || req.status === 'IN_PROGRESS') && (
                            <button onClick={() => openFulfillModal(req)} className="px-2 py-1 text-xs bg-emerald-100 text-emerald-700 rounded hover:bg-emerald-200 transition">تم التوفير</button>
                          )}
                        </div>
                      ) : (
                        <span className="text-xs text-slate-400">للاطّلاع فقط</span>
                      )}
                    </td>
                  </tr>
                  {expandedRows.has(req.id) && (
                    <tr key={`${req.id}-items`} className="bg-slate-50">
                      <td colSpan={activeTab === 'CUSTOMER_PRODUCT' ? 7 : 6} className="px-6 py-3">
                        <div className="space-y-1">
                          <p className="text-xs font-semibold text-slate-600 mb-2">المواد المطلوبة:</p>
                          {req.items.map(item => (
                            <div key={item.id} className="flex gap-4 text-xs text-slate-700">
                              <span className="font-medium">{item.productName}</span>
                              {item.model && <span>الموديل: {item.model}</span>}
                              {item.spec && <span>النوعية: {item.spec}</span>}
                              <span>الكمية: {item.quantity}</span>
                              {item.unitPrice != null && <span>سعر الوحدة: {fmt(item.unitPrice)}</span>}
                              {item.totalPrice != null && <span>الإجمالي: {fmt(item.totalPrice)}</span>}
                              <span className={item.fulfilled ? 'text-emerald-600' : 'text-amber-600'}>
                                {item.fulfilled ? 'تم التوفير' : 'لم يوفر'}
                              </span>
                              {item.itemNotes && <span className="text-slate-500">({item.itemNotes})</span>}
                            </div>
                          ))}
                          {req.supplier && <p className="text-xs text-slate-500 mt-2">المورد: <b>{req.supplier.name}</b></p>}
                          {req.notes && <p className="text-xs text-slate-500 mt-2">ملاحظات: {req.notes}</p>}
                          {req.fulfilledBy && <p className="text-xs text-slate-500">وفّره: {req.fulfilledBy.name}</p>}
                          {req.fulfillmentNotes && <p className="text-xs text-slate-500">ملاحظات التوفير: {req.fulfillmentNotes}</p>}
                        </div>
                      </td>
                    </tr>
                  )}
                </Fragment>
              ))}
            </tbody>
          </table>
        </div>
      </div>

      {/* Fulfill Modal */}
      {fulfillModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40">
          <div className="bg-white rounded-xl shadow-xl w-full max-w-lg mx-4 max-h-[90vh] overflow-y-auto p-6 space-y-4" dir="rtl">
            <div className="flex items-center justify-between">
              <h2 className="text-lg font-bold text-[#0f2040]">تنفيذ الطلب - {fulfillModal.code}</h2>
              <button onClick={() => setFulfillModal(null)} className="text-slate-400 hover:text-slate-600 text-xl">×</button>
            </div>

            <div className="space-y-3">
              <p className="text-sm font-semibold text-slate-600">المواد:</p>
              {fulfillItems.map((item, idx) => (
                <div key={item.id} className="border border-slate-200 rounded-lg p-3 space-y-2">
                  <div className="flex items-center justify-between">
                    <span className="text-sm font-medium">{item.productName} (الكمية: {item.quantity})</span>
                    <label className="flex items-center gap-1 text-xs">
                      <input
                        type="checkbox"
                        checked={item.fulfilled}
                        onChange={e => {
                          const next = [...fulfillItems]
                          next[idx] = { ...next[idx], fulfilled: e.target.checked }
                          setFulfillItems(next)
                        }}
                      />
                      تم التوفير
                    </label>
                  </div>
                  <div className="flex gap-2">
                    <div className="flex-1">
                      <label className="text-xs text-slate-500">سعر الوحدة</label>
                      <input
                        type="number"
                        step="0.01"
                        value={item.unitPrice}
                        onChange={e => {
                          const next = [...fulfillItems]
                          next[idx] = { ...next[idx], unitPrice: e.target.value }
                          setFulfillItems(next)
                        }}
                        className="w-full border border-slate-300 rounded px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
                      />
                    </div>
                    <div className="flex-1">
                      <label className="text-xs text-slate-500">الإجمالي</label>
                      <input
                        type="number"
                        step="0.01"
                        value={item.totalPrice}
                        onChange={e => {
                          const next = [...fulfillItems]
                          next[idx] = { ...next[idx], totalPrice: e.target.value }
                          setFulfillItems(next)
                        }}
                        className="w-full border border-slate-300 rounded px-2 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
                      />
                    </div>
                  </div>
                </div>
              ))}
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">
                المورد <span className="text-red-500">*</span>
              </label>
              <select
                value={fulfillSupplierId}
                onChange={e => setFulfillSupplierId(e.target.value)}
                className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
              >
                <option value="">اختر المورد الي انجابت منه المادة</option>
                {suppliers.map(sp => (
                  <option key={sp.id} value={sp.id}>{sp.companyName}{sp.ownerName ? ` — ${sp.ownerName}` : ''}</option>
                ))}
              </select>
              {suppliers.length === 0 && (
                <p className="mt-1 text-xs text-amber-600">ما اكو موردين مضافين — ضيفهم من شاشة الموردين أول.</p>
              )}
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">التكلفة الإجمالية</label>
              <input
                type="number"
                step="0.01"
                value={fulfillCost}
                onChange={e => setFulfillCost(e.target.value)}
                className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
              />
            </div>

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">ملاحظات التوفير</label>
              <textarea
                value={fulfillNotes}
                onChange={e => setFulfillNotes(e.target.value)}
                rows={2}
                className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
              />
            </div>

            <div className="flex gap-2 justify-end">
              <button onClick={() => setFulfillModal(null)} className="px-4 py-2 text-sm text-slate-600 hover:bg-slate-100 rounded-lg transition">إلغاء</button>
              <button
                onClick={handleFulfill}
                disabled={fulfilling}
                className="px-6 py-2 text-sm bg-emerald-700 text-white rounded-lg font-medium hover:bg-emerald-800 disabled:opacity-50 transition"
              >
                {fulfilling ? 'جاري التنفيذ...' : 'تأكيد التوفير'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
