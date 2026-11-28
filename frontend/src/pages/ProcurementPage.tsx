import Modal from '../ui/Modal'
import { Fragment, useState, useEffect, useContext } from 'react'
import PurchaseControlPanel from '../components/PurchaseControlPanel'
import { purchaseApi } from '../api'
import { useSearchParams } from 'react-router-dom'
import { api, procurementTypeLabels, type ProcurementRequest, type ProcurementRequestType, type ProcurementStats, type Booking } from '../api'
import { SessionContext, hasRole, isAdminOrOwner } from '../session'
import BookingCodeChip from '../components/BookingCodeChip'
import { swallow } from '../utils/swallow'
import { BTN_GHOST, BTN_SUCCESS } from '../ui/classes'
import { STATUS_LABELS } from '../statusLabels'
import { LineIcon } from '../components/LineIcon'
import './ProcurementRequests.css'

const statusLabels: Record<string, string> = {
  PENDING: 'بانتظار التوفير',
  IN_PROGRESS: 'جاري التوفير',
  FULFILLED: 'تم التوفير',
  REJECTED: STATUS_LABELS.REJECTED,
}

export default function ProcurementPage() {
  const { employee, permissions } = useContext(SessionContext)
  // توفير المواد وتحديد حالتها (توفير/رفض) يقتصر على إداري الكميات والأدمن فقط —
  // أي موظف ثاني (فني، مدير مشاريع، مراقب...) يشوف الطلبات وحالتها بس بدون تحكم.
  const canManageProcurement = isAdminOrOwner(employee) || hasRole(employee, 'PROCUREMENT_ADMIN')
  const isAdminLike = isAdminOrOwner(employee)
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
  const [q, setQ] = useState('')
  const [statusF, setStatusF] = useState('')
  const [filterOpen, setFilterOpen] = useState(false)
  const [menuFor, setMenuFor] = useState<string | null>(null)
  // القائمة تطفو فوگ الصفحة (fixed) — داخل الجدول القابل للتمرير چانت تنقص
  const [menuPos, setMenuPos] = useState<{ top?: number; bottom?: number; left: number }>({ top: 0, left: 0 })

  // Fulfill modal
  const [fulfillModal, setFulfillModal] = useState<ProcurementRequest | null>(null)
  const [fulfillCost, setFulfillCost] = useState('')
  const [fulfillNotes, setFulfillNotes] = useState('')
  const [fulfillSupplierId, setFulfillSupplierId] = useState('')
  const [fulfillItems, setFulfillItems] = useState<{ id: string; productName: string; quantity: number; unitPrice: string; totalPrice: string; fulfilled: boolean }[]>([])
  const [fulfilling, setFulfilling] = useState(false)
  // (ع) 10-15: بعد الشراء ينطلب تقييم المورد مباشرة
  const [reviewFor, setReviewFor] = useState<ProcurementRequest | null>(null)
  // شرط العرضين: الشراء فوق الحد لازم وياه عرض مورد ثاني
  const [quoteThreshold, setQuoteThreshold] = useState(500000)
  const [quotes, setQuotes] = useState<{ supplierId: string; totalPrice: string }[]>([])
  useEffect(() => { purchaseApi.policy().then((p) => setQuoteThreshold(p.quoteThreshold)).catch(swallow) }, [])
  const canVerifyPurchase = isAdminLike || hasRole(employee, 'FINANCE') || permissions.includes('finance_audit')
  const canWatchPurchase = canVerifyPurchase || employee?.role === 'MONITOR' || permissions.includes('monitoring')

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

  // (11-17) 🔔 «ذكّر المدير»: طلب معلّق من يومين فأكثر، ومرة باليوم — نفس حارس الخادم
  const [reminded, setReminded] = useState<Record<string, boolean>>({})
  const [nowMs] = useState(() => Date.now())
  const canRemind = (req: ProcurementRequest) => {
    if (req.status !== 'PENDING' && req.status !== 'IN_PROGRESS') return false
    const day = 24 * 3600 * 1000
    if (nowMs - new Date(req.createdAt).getTime() < 2 * day) return false
    return !reminded[req.id] && !(req.remindedAt && nowMs - new Date(req.remindedAt).getTime() < day)
  }
  const handleRemind = async (id: string) => {
    try {
      await api.remindProcurement(id)
      setReminded((x) => ({ ...x, [id]: true }))
      setError('')
    } catch (e) {
      setError(e instanceof Error ? e.message : 'تعذّر إرسال التذكير')
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
      // (ك) 11-28 زينا: يبدي مأشّر — چان فارغ فالطلب ينحفظ «تم التوفير» والمادة «لم يوفر» (PR-0004)
      fulfilled: true,
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
        quotes: quotes.filter((q) => q.supplierId && Number(q.totalPrice) > 0).map((q) => ({ supplierId: q.supplierId, totalPrice: Number(q.totalPrice) })),
        items: fulfillItems.map(i => ({
          id: i.id,
          unitPrice: i.unitPrice ? parseFloat(i.unitPrice) : undefined,
          totalPrice: i.totalPrice ? parseFloat(i.totalPrice) : undefined,
          fulfilled: i.fulfilled,
        })),
      })
      const done = fulfillModal
      setFulfillModal(null)
      setQuotes([])
      await loadData()
      setReviewFor({ ...done, supplierId: fulfillSupplierId, supplier: { id: fulfillSupplierId, name: suppliers.find((x) => x.id === fulfillSupplierId)?.companyName ?? '' } })
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
    n != null ? n.toLocaleString('ar-IQ', { minimumFractionDigits: 0, maximumFractionDigits: 2 }) : '—'

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-[#2c5aad]" />
      </div>
    )
  }

  return (
    <div dir="rtl" className="p-6 space-y-6">
      {/* (ك) 11-22 زينا: الرأس والبطاقات مثل التصميم — الألوان بـindex.css (.pc-*) */}
      <div className="flex items-center gap-3">
        <span className="pc-logo flex h-12 w-12 items-center justify-center rounded-xl"><LineIcon n="cart" className="h-7 w-7" /></span>
        <h1 className="text-3xl font-extrabold text-[#0f2040]">طلبات المواد والمشتريات</h1>
      </div>

      {error && (
        <div className="bg-red-50 border border-red-200 text-red-700 rounded-lg p-3 text-sm">{error}</div>
      )}

      {/* Stats Cards */}
      {stats && (
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
          {[
            { label: 'إجمالي المصروفات', value: `${fmt(stats.totalSpent)} د.ع`, color: 'pc-v-blue', icon: 'wallet' },
            { label: 'طلبات معلقة', value: stats.pendingCount.toString(), color: 'pc-v-amber', icon: 'clock' },
            { label: 'تم التوفير', value: stats.fulfilledCount.toString(), color: 'pc-v-green', icon: 'bars' },
            { label: 'مصروفات الشهر', value: `${fmt(stats.monthlySpent)} د.ع`, color: 'pc-v-blue', icon: 'layers' },
          ].map((card, i) => (
            <div key={i} className="pc-card flex items-center gap-4 rounded-2xl border px-5 py-5">
              <span className="pc-ic flex h-14 w-14 shrink-0 items-center justify-center rounded-xl"><LineIcon n={card.icon} className="h-8 w-8" /></span>
              <div className="min-w-0 flex-1 text-right">
                <p className="text-lg font-bold">{card.label}</p>
                <p className={`text-2xl font-extrabold tabular-nums ${card.color}`}>{card.value}</p>
              </div>
            </div>
          ))}
        </div>
      )}

      <PurchaseControlPanel requests={requests} canBuy={canManageProcurement} canVerify={canVerifyPurchase} canWatch={canWatchPurchase}
        canDecide={isAdminLike} reviewFor={reviewFor} onReviewClose={() => setReviewFor(null)} />

      {/* New Request Form — (ك) 11-27 زينا: تنفتح من «اضافة طلب جديد» بالجدول، والشريط القديم انشال */}
      {showForm && (
      <div className="rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
        <div className="flex items-center justify-between p-4 font-semibold text-[#0f2040]">
          <span>+ طلب جديد</span>
          <button type="button" onClick={() => setShowForm(false)} aria-label="سد الاستمارة" className="rounded-lg px-2 py-1 text-lg leading-none text-slate-500 hover:bg-slate-100">×</button>
        </div>
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
      )}

      {/* (ك) 11-27 زينا: التبويبات والجدول مثل التصميم — الألوان بـProcurementRequests.css (.pr-*) */}
      <div className="pr-tabs">
        {([
          { k: 'CUSTOMER_PRODUCT' as const, label: 'طلبات منتجات الزبائن', icon: 'cart' },
          { k: 'PERSONAL_SUPPLY' as const, label: 'طلبات احتياجات شخصية', icon: 'user' },
          { k: 'MANUAL_SUPPLY' as const, label: 'طلبات مواد يدوية', icon: 'box' },
        ]).map((t) => (
          <button key={t.k} type="button" onClick={() => setActiveTab(t.k)} className={`pr-tab ${activeTab === t.k ? 'on' : ''}`}>
            <LineIcon n={t.icon} className="h-6 w-6" /> {t.label}
          </button>
        ))}
      </div>

      {(() => {
        const tabTitle = activeTab === 'CUSTOMER_PRODUCT' ? 'طلبات منتجات الزبائن' : activeTab === 'PERSONAL_SUPPLY' ? 'طلبات احتياجات شخصية' : 'طلبات مواد يدوية'
        const tabSub = activeTab === 'CUSTOMER_PRODUCT' ? 'إدارة ومتابعة طلبات الزبائن للمنتجات' : activeTab === 'PERSONAL_SUPPLY' ? 'إدارة ومتابعة طلبات الاحتياجات الشخصية' : 'إدارة ومتابعة طلبات المواد اليدوية'
        const tabIcon = activeTab === 'CUSTOMER_PRODUCT' ? 'cart' : activeTab === 'PERSONAL_SUPPLY' ? 'user' : 'box'
        const isCust = activeTab === 'CUSTOMER_PRODUCT'
        const cols = isCust ? 7 : 6
        const term = q.trim().toLowerCase()
        const rows = requests.filter(r => r.requestType === activeTab)
          .filter(r => !statusF || r.status === statusF)
          .filter(r => !term || [r.code, r.requestedBy.name, r.booking?.code, r.booking?.customer?.name].some(v => (v ?? '').toLowerCase().includes(term)))
        return (
      <div className="pr-panel">
        <div className="pr-head">
          <div className="flex items-center gap-3">
            <span className="pr-logo"><LineIcon n={tabIcon} className="h-7 w-7" /></span>
            <div>
              <h2 className="pr-title">{tabTitle}</h2>
              <p className="pr-sub">{tabSub}</p>
            </div>
          </div>
          <div className="pr-tools">
            <label className="pr-search">
              <LineIcon n="search" className="h-5 w-5" />
              <input value={q} onChange={e => setQ(e.target.value)} placeholder="البحث في الطلبات..." />
            </label>
            <div className="relative">
              <button type="button" className={`pr-filter ${statusF ? 'on' : ''}`} onClick={() => setFilterOpen(o => !o)} aria-label="فلترة حسب الحالة">
                <LineIcon n="filter" className="h-5 w-5" />
              </button>
              {filterOpen && (
                <div className="pr-menu" onMouseLeave={() => setFilterOpen(false)}>
                  {[['', 'كل الحالات'], ...Object.entries(statusLabels)].map(([k, l]) => (
                    <button key={k} type="button" className={statusF === k ? 'on' : ''} onClick={() => { setStatusF(k); setFilterOpen(false) }}>{l}</button>
                  ))}
                </div>
              )}
            </div>
            <button type="button" className="pr-add" onClick={() => { setShowForm(true); window.scrollTo({ top: 0, behavior: 'smooth' }) }}>
              <LineIcon n="plus" className="h-5 w-5" /> اضافة طلب جديد
            </button>
          </div>
        </div>

        <div className="overflow-x-auto">
          <table className="pr-table">
            <thead>
              <tr>
                <th><span><LineIcon n="tag" /> الكود</span></th>
                <th><span><LineIcon n="user" /> مقدم الطلب</span></th>
                {isCust && <th><span><LineIcon n="bookmark" /> الحجز</span></th>}
                <th className="c"><span><LineIcon n="layers" /> الحالة</span></th>
                <th className="c"><span><LineIcon n="db" /> التكلفة</span></th>
                <th className="c"><span><LineIcon n="cal" /> التاريخ</span></th>
                <th className="c act"><span><LineIcon n="gear" /> إجراءات</span></th>
              </tr>
            </thead>
            <tbody>
              {rows.length === 0 && (
                <tr><td colSpan={cols} className="pr-empty">لا توجد طلبات</td></tr>
              )}
              {rows.map(req => (
                <Fragment key={req.id}>
                  <tr className="pr-row" onClick={() => toggleRow(req.id)}>
                    <td><span className="pr-code">{req.code}</span></td>
                    <td><span className="pr-who"><LineIcon n="user" className="h-5 w-5" /> {req.requestedBy.name}</span></td>
                    {isCust && (
                      <td className="c">{req.booking ? `${req.booking.code} - ${req.booking.customer?.name}` : '—'}</td>
                    )}
                    <td className="c">
                      <span className={`pr-status s-${req.status}`}>
                        {req.status === 'FULFILLED' && <LineIcon n="check" className="h-4 w-4" />} {statusLabels[req.status]}
                      </span>
                    </td>
                    <td className="c num">{fmt(req.totalCost)}</td>
                    <td className="c num">{new Date(req.createdAt).toLocaleDateString('ar-IQ')}</td>
                    <td className="c" onClick={e => e.stopPropagation()}>
                      <div className="relative inline-block">
                        <button type="button" className="pr-dots" aria-label="إجراءات" onClick={e => { const r = e.currentTarget.getBoundingClientRect(); setMenuPos(r.bottom + 220 > window.innerHeight ? { bottom: window.innerHeight - r.top + 6, left: r.left } : { top: r.bottom + 6, left: r.left }); setMenuFor(m => m === req.id ? null : req.id) }}>•••</button>
                        {menuFor === req.id && (
                          <div className="pr-menu float" style={menuPos} onMouseLeave={() => setMenuFor(null)}>
                            <button type="button" onClick={() => { toggleRow(req.id); setMenuFor(null) }}>{expandedRows.has(req.id) ? 'اخفي المواد' : 'شوف المواد'}</button>
                            {canManageProcurement ? (<>
                              {req.status === 'PENDING' && (<>
                                <button type="button" onClick={() => { setMenuFor(null); handleStatusUpdate(req.id, 'IN_PROGRESS') }}>جاري التوفير</button>
                                <button type="button" className="bad" onClick={() => { setMenuFor(null); handleStatusUpdate(req.id, 'REJECTED') }}>رفض</button>
                              </>)}
                              {canRemind(req) && (
                                <button type="button" title="الطلب معلّق من يومين فأكثر — يوصل إشعار للمدير (مرة باليوم)" onClick={() => { setMenuFor(null); handleRemind(req.id) }}>🔔 ذكّر المدير</button>
                              )}
                              {reminded[req.id] && <span className="ok">✓ انذكّر</span>}
                              {(req.status === 'PENDING' || req.status === 'IN_PROGRESS') && (
                                <button type="button" className="good" onClick={() => { setMenuFor(null); openFulfillModal(req) }}>تم التوفير</button>
                              )}
                            </>) : <span className="muted">للاطّلاع فقط</span>}
                          </div>
                        )}
                      </div>
                    </td>
                  </tr>
                  {expandedRows.has(req.id) && (
                    <tr key={`${req.id}-items`} className="pr-items">
                      <td colSpan={cols} className="px-6 py-3">
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
        )
      })()}

      {/* Fulfill Modal */}
      {fulfillModal && (
        <Modal title={<>تنفيذ الطلب - {fulfillModal.code}</>} onClose={() => setFulfillModal(null)} size="md" busy={fulfilling} closeOnBackdrop={false} bodyClassName="space-y-4 p-6">

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

            {(() => {
              const total = fulfillCost ? parseFloat(fulfillCost) : fulfillItems.reduce((a, i) => a + (parseFloat(i.totalPrice) || 0), 0)
              const need = quoteThreshold > 0 && total >= quoteThreshold
              const okCount = quotes.filter((q) => q.supplierId && q.supplierId !== fulfillSupplierId && Number(q.totalPrice) > 0).length
              return (
                <div className={`rounded-lg border p-3 ${need && okCount === 0 ? 'border-red-300 bg-red-50' : 'border-slate-200'}`}>
                  <p className="text-sm font-bold text-slate-700">🧾 عروض الموردين الثانيين {need ? <span className="text-red-600">(إلزامي — الشراء فوق {quoteThreshold.toLocaleString('en-US')} د.ع)</span> : <span className="font-normal text-slate-400">(اختياري)</span>}</p>
                  {quotes.map((q, i) => (
                    <div key={i} className="mt-2 flex gap-2">
                      <select value={q.supplierId} onChange={(e) => setQuotes(quotes.map((x, j) => (j === i ? { ...x, supplierId: e.target.value } : x)))}
                        className="flex-1 rounded-lg border border-slate-300 px-2 py-1.5 text-sm">
                        <option value="">المورد</option>
                        {suppliers.filter((x) => x.id !== fulfillSupplierId).map((x) => <option key={x.id} value={x.id}>{x.companyName}</option>)}
                      </select>
                      <input type="number" placeholder="سعره الكلي" value={q.totalPrice} onChange={(e) => setQuotes(quotes.map((x, j) => (j === i ? { ...x, totalPrice: e.target.value } : x)))}
                        className="w-36 rounded-lg border border-slate-300 px-2 py-1.5 text-sm" />
                      <button onClick={() => setQuotes(quotes.filter((_, j) => j !== i))} className="text-xl text-red-500">×</button>
                    </div>
                  ))}
                  <button onClick={() => setQuotes([...quotes, { supplierId: '', totalPrice: '' }])} className="mt-2 text-sm text-[#2c5aad] hover:underline">+ عرض مورد ثاني</button>
                </div>
              )
            })()}

            <div>
              <label className="block text-sm font-medium text-slate-700 mb-1">ملاحظات التوفير</label>
              <textarea
                value={fulfillNotes}
                onChange={e => setFulfillNotes(e.target.value)}
                rows={2}
                className="w-full border border-slate-300 rounded-lg px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-[#2c5aad]"
              />
            </div>

            <div className={BTN_GHOST}>
              <button onClick={() => setFulfillModal(null)} className="px-4 py-2 text-sm text-slate-600 hover:bg-slate-100 rounded-lg transition">إلغاء</button>
              <button
                onClick={handleFulfill}
                disabled={fulfilling}
                className={BTN_SUCCESS}
              >
                {fulfilling ? 'جاري التنفيذ...' : 'تأكيد التوفير'}
              </button>
            </div>
        </Modal>
      )}
    </div>
  )
}
