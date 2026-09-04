import { useEffect, useRef, useState } from 'react'
import { api, type Booking, type Employee, type CartItem, type Product, type JobDurationEstimate, type VehicleOption } from '../api'
import { formatCustomerCode } from '../utils/identity'
import { executionStarted } from '../bookingStage'
import { useSession } from '../session'
import { LocationPicker } from '../components/MapLazy'
import CompletionBadge from '../components/CompletionBadge'
import BookingEditPanel from '../components/BookingEditPanel'
import BookingLifecycleActions from '../components/BookingLifecycleActions'
import { COMPLETION_ORDER, completionLabel } from '../components/completionStates'
import { matches as searchMatches } from '../utils/search'
import EntityIdentity from '../components/EntityIdentity'
import BookingTimelineView from '../components/BookingTimeline'
import { promptChoice } from '../utils/promptChoice'
import { bookingDeleteChannelLabels, bookingDeleteTypeLabels, BOOKING_NO_ANSWER_CHOICE, bookingNoAnswerLabel, type BookingDeleteChannel, type BookingDeleteRequestType } from '../api'

const DELETE_CHANNEL_OPTIONS: [BookingDeleteChannel, string][] =
  (Object.entries(bookingDeleteChannelLabels) as [BookingDeleteChannel, string][])
// ⚠️ «الزبون ما رد» خيار بنفس القائمة بس **مو طلب حذف** — يُعالَج
// بفرع مستقل ينقل الحجز للانتظار فوراً، بلا ما يوصل الخادم كنوع طلب.
type DeleteTypeChoice = BookingDeleteRequestType | typeof BOOKING_NO_ANSWER_CHOICE
const DELETE_TYPE_OPTIONS: [DeleteTypeChoice, string][] = [
  ...(Object.entries(bookingDeleteTypeLabels) as [BookingDeleteRequestType, string][]),
  [BOOKING_NO_ANSWER_CHOICE, bookingNoAnswerLabel],
]

// أسماء كل خدمات الحجز (الزبون ممكن يطلب أكثر من منظومة بنفس الحجز)
function serviceNames(b: { service?: { name: string } | null; services?: { name: string }[] }): string {
  if (b.services && b.services.length > 0) return b.services.map((s) => s.name).join(' + ')
  return b.service?.name || 'بدون خدمة محددة'
}

// Convert an ISO date string to the local "YYYY-MM-DDTHH:mm" format expected by datetime-local inputs
const toLocalInput = (iso: string) => {
  const d = new Date(iso)
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`
}

// ساعات العمل وطول كل موعد (بالساعات) - تستخدم لاقتراح أقرب موعد فاضي
const WORK_START_HOUR = 8
const LAST_SLOT_HOUR = 22
const SLOT_HOURS = 2


const techRoles: { key: 'TECH_1' | 'TECH_2' | 'TECH_3'; label: string }[] = [
  { key: 'TECH_1', label: 'الفني الأول' },
  { key: 'TECH_2', label: 'الفني الثاني' },
  { key: 'TECH_3', label: 'الفني الثالث' },
]

export default function Coordinator() {
  const { employee: currentUser, permissions } = useSession()
  // منو يكدر يطلب حذف حجز — نفس قائمة شاشة الحجوزات بالضبط، حتى ما
  // يصير الزر يطلع بشاشة ويختفي بالثانية لنفس الموظف.
  const canRequestDelete = currentUser?.role === 'ADMIN' || currentUser?.role === 'OWNER'
    || currentUser?.role === 'HR_COORDINATOR' || currentUser?.role === 'MONITOR'
    || (permissions ?? []).includes('booking_delete_request')
  // الحذف (الأرشفة) قرار إداري — المنسّق يأجّل ويحط بالانتظار، بس ما يحذف
  const isAdmin = currentUser?.role === 'ADMIN' || currentUser?.role === 'OWNER'
  const [bookings, setBookings] = useState<Booking[]>([])
  const [matches, setMatches] = useState<Record<string, Employee[]>>({})
  // الحجز الي قيد التثبيت حالياً — الأزرار كانت بلا أي إشارة انتظار، فالمستخدم
  // يحس النظام بطيء أو معلّق ويضغط عدة مرات.
  const [search, setSearch] = useState('')
  const [supervisors, setSupervisors] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // قيم التثبيت (التكلفة المقدرة + العنوان + الموعد) قبل الضغط على تثبيت
  // مسوّدات المواعيد الي لسه ما انحفظن — تنحجز بالمواعيد المتاحة حتى
  // ما ينختار نفس الوقت لحجزين بنفس اللحظة.
  // ⚠️ صارت للقراءة بس بعد ما انشال قسم «بانتظار التثبيت» (التثبيت
  // صار بمحطته)، وخلّيناها لأن حجز المواعيد يعتمد عليها.
  const [scheduleDrafts] = useState<Record<string, string>>({})

  // سلة المنتجات
  const [cartItems, setCartItems] = useState<Record<string, CartItem[]>>({})
  const [cartOpen, setCartOpen] = useState<Record<string, boolean>>({})
  const [cartForm, setCartForm] = useState<Record<string, { productName: string; quantity: string; unitPrice: string; notes: string }>>({})
  const [products, setProducts] = useState<Product[]>([])
  const [vehicles, setVehicles] = useState<VehicleOption[]>([])
  const [scheduleMode, setScheduleMode] = useState<Record<string, 'slots' | 'manual'>>({})
  const [editingLocationId, setEditingLocationId] = useState<string | null>(null)

  // تقدير مدة العمل المتعلَّم تلقائياً من بيانات فعلية سابقة (بدون رقم مفروض
  // يدوياً) — يُعرض بجانب اختيار الفنيين حتى المنسق يشوف تقدير واقعي قبل التثبيت.
  const [durationEstimates, setDurationEstimates] = useState<Record<string, JobDurationEstimate>>({})
  const estimateAsked = useRef<Set<string>>(new Set())

  // ═══ ليش تقدير المدة صار بمفتاح مو بحجز ═══
  //
  // «هاي الصفحة ثكيلة بالتحميل، هاي مشكلة، لازم أسرع شوي».
  //
  // كان هذا الأثر يدزّ **طلب مستقل لكل حجز مثبّت**: عشرين حجز = عشرين
  // نداء. والأسوأ إنه معلّق على `bookings` كلها — وكل حفظ بالشاشة
  // (سعر، موعد، مشرف، فني) يبدّل مصفوفة الحجوزات، فينعاد الأثر
  // ويعيد **نفس** العشرين نداء من الصفر. فالصفحة تثقل بكل ضغطة.
  //
  // والتقدير أصلاً ما يعتمد على الحجز، يعتمد على أربعة أرقام
  // (منظومة، نوع شغل، عدد، حجم كادر) — وأغلب الحجوزات تشترك بنفس
  // الأربعة. فصار المفتاح هو الأربعة: نطلب مرة وحدة لكل تركيبة
  // مختلفة، وما نعيد طلب تركيبة وصلتنا.
  const estimateKeyOf = (b: Booking) => {
    const crewSize = 1 + b.assignments.filter((a) => a.role === 'TECH_1' || a.role === 'TECH_2' || a.role === 'TECH_3').length
    const itemCount = b.deviceCount || b.systemCount || 1
    const jobType: 'MAINTENANCE' | 'INSTALL' = b.bookingType === 'MAINTENANCE' ? 'MAINTENANCE' : 'INSTALL'
    return { key: `${b.systemType}|${jobType}|${itemCount}|${crewSize}`, systemName: b.systemType!, jobType, itemCount, crewSize }
  }

  useEffect(() => {
    const wanted = new Map<string, ReturnType<typeof estimateKeyOf>>()
    for (const b of bookings) {
      if (b.status !== 'CONFIRMED' || !b.systemType) continue
      const k = estimateKeyOf(b)
      if (!wanted.has(k.key)) wanted.set(k.key, k)
    }
    for (const k of wanted.values()) {
      // ⚠️ سجل الطلبات بـ`ref` مو بالحالة: الحالة ما تنكتب إلا بعد
      // ما يرجع الجواب، فطلبين لنفس التركيبة بنفس اللحظة يمرّون
      // الاثنين. الـ`ref` ينكتب فوراً فيمنع الثاني.
      if (estimateAsked.current.has(k.key)) continue
      estimateAsked.current.add(k.key)
      api
        .getJobDurationEstimate({ systemName: k.systemName, jobType: k.jobType, itemCount: k.itemCount, crewSize: k.crewSize })
        .then((estimate) => setDurationEstimates((p) => ({ ...p, [k.key]: estimate })))
        // لو فشل، نشيله من السجل حتى يعاود بالتحميل الجاي
        .catch(() => { estimateAsked.current.delete(k.key) })
    }
  }, [bookings])

  const getAvailableSlots = (excludeId?: string) => {
    const takenKeys = new Set<string>()
    const toKey = (d: Date) => `${d.getFullYear()}-${d.getMonth()}-${d.getDate()}-${d.getHours()}`
    bookings
      .filter(b => b.scheduledAt && (b.status === 'CONFIRMED' || b.status === 'PENDING' || b.status === 'IN_PROGRESS'))
      .forEach(b => takenKeys.add(toKey(new Date(b.scheduledAt!))))
    Object.entries(scheduleDrafts).forEach(([id, val]) => {
      if (id !== excludeId && val) takenKeys.add(toKey(new Date(val)))
    })

    const slots: { value: string; label: string }[] = []
    const slot = new Date()
    slot.setMinutes(0, 0, 0)
    if (slot.getHours() < WORK_START_HOUR) slot.setHours(WORK_START_HOUR)
    else if (slot.getHours() > LAST_SLOT_HOUR) { slot.setDate(slot.getDate() + 1); slot.setHours(WORK_START_HOUR) }
    else { slot.setHours(slot.getHours() + 1); if (slot.getHours() > LAST_SLOT_HOUR) { slot.setDate(slot.getDate() + 1); slot.setHours(WORK_START_HOUR) } }

    for (let i = 0; i < 200 && slots.length < 30; i++) {
      const key = toKey(slot)
      if (!takenKeys.has(key)) {
        const pad = (n: number) => String(n).padStart(2, '0')
        const value = `${slot.getFullYear()}-${pad(slot.getMonth() + 1)}-${pad(slot.getDate())}T${pad(slot.getHours())}:00`
        const label = slot.toLocaleString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'numeric', hour: '2-digit', minute: '2-digit' })
        slots.push({ value, label })
      }
      slot.setHours(slot.getHours() + SLOT_HOURS)
      if (slot.getHours() > LAST_SLOT_HOUR || slot.getHours() < WORK_START_HOUR) { slot.setDate(slot.getDate() + 1); slot.setHours(WORK_START_HOUR) }
    }
    return slots
  }

  // ═══ ليش كل حفظ لازم يمر من هنا ═══
  //
  // كانت إحدى عشر عملية بهاي الشاشة تنادي السيرفر بـawait عارية بلا
  // try: تعيين مشرف، تحديد موعد، تعديل سعر وعنوان وموقع، تكليف فني،
  // إضافة وحذف من السلة.
  //
  // لمن يفشل أي واحد منهن — انقطاع، صلاحية ناقصة، سيرفر يعيد التشغيل —
  // الوعد يترفض بلا ما يمسكه أحد: ما تطلع رسالة، والخانة تبقى تعرض
  // القيمة الجديدة الي كتبها الموظف (لأن الحالة المحلية ما انرجعت).
  //
  // فالإداري يشوف السعر الجديد قدامه ويكمّل شغله وهو مطمّن — والسعر
  // ما وصل قاعدة البيانات إطلاقاً. وينكشف بعد أيام لمن يفتح الحجز
  // ويلگى الرقم القديم، وما يعرف منو غيّره ولا متى.
  //
  // الحفظ الي يفشل بصمت أخطر من الي يفشل بصوت.
  const [saveError, setSaveError] = useState<string | null>(null)

  /** يحدّث حجز بالقائمة، ويعرض سبب الفشل بدل ما يبلعه. */
  const applyUpdate = async (what: string, fn: () => Promise<Booking>) => {
    try {
      const updated = await fn()
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
      setSaveError(null)
      return updated
    } catch (e) {
      setSaveError(`تعذر ${what}: ${e instanceof Error ? e.message : 'خطأ غير متوقع'}`)
      return null
    }
  }

  /** نفس الفكرة للعمليات الي ما ترجّع حجز (السلة). */
  const runAction = async (what: string, fn: () => Promise<unknown>) => {
    try {
      await fn()
      setSaveError(null)
      return true
    } catch (e) {
      setSaveError(`تعذر ${what}: ${e instanceof Error ? e.message : 'خطأ غير متوقع'}`)
      return false
    }
  }

  const loadCart = async (bookingId: string) => {
    const items = await api.getCartItems(bookingId)
    setCartItems(prev => ({ ...prev, [bookingId]: items }))
  }

  const addCartItem = async (bookingId: string) => {
    const form = cartForm[bookingId]
    if (!form?.productName || !form?.quantity || !form?.unitPrice) return
    const ok = await runAction('إضافة المادة للسلة', () => api.addCartItem(bookingId, {
      productName: form.productName,
      quantity: Number(form.quantity),
      unitPrice: Number(form.unitPrice),
      notes: form.notes || undefined,
    }))
    // ⚠️ ما نفضّي الخانات إلا بعد نجاح الحفظ: تفضيتها عند الفشل تخلي
    // الإداري يعيد كتابة المادة كلها من الصفر بلا ما يعرف ليش راحت.
    if (!ok) return
    setCartForm(prev => ({ ...prev, [bookingId]: { productName: '', quantity: '', unitPrice: '', notes: '' } }))
    loadCart(bookingId)
  }

  const removeCartItem = async (bookingId: string, itemId: string) => {
    if (!(await runAction('حذف المادة من السلة', () => api.deleteCartItem(itemId)))) return
    loadCart(bookingId)
  }

  const toggleCart = (bookingId: string) => {
    const isOpen = !cartOpen[bookingId]
    setCartOpen(prev => ({ ...prev, [bookingId]: isOpen }))
    if (isOpen && !cartItems[bookingId]) loadCart(bookingId)
  }

  const loadMatches = async (bookingsForService: Booking[], serviceId: string) => {
    const employees = await api.matchEmployees(serviceId)
    setMatches((prev) => {
      const next = { ...prev }
      for (const b of bookingsForService) next[b.id] = employees
      return next
    })
  }

  const load = () => {
    api
      // نجيب الحجوزات الفعّالة بس (قيد التنسيق) لا كل الأرشيف التاريخي — هذي الصفحة
      // ما تستخدم إطلاقاً حجوزات IN_PROGRESS/COMPLETED/CANCELLED، وجلبها كلها كان يبطّئ
      // الصفحة كثير مع تراكم آلاف الحجوزات القديمة.
      // ═══ ⚠️ WAITING انشالت من هنا — قرار لاحق يعدّل قراراً أسبق ═══
      //
      // چانت تنجاب حتى «يضل الحجز مرئي للمنسّق حتى يعاود الاتصال».
      // بس صاحب النظام رجع وقال صراحة: «زبون ما رد ماريده يضهرلي
      // بمكانين، اريده يضهر فقط بالمكان مال الزبون ما رد… اني بعدين
      // ارجعله يدويا».
      //
      // فالحجز الي الزبون ما رد عليه صار له **مكان واحد**: طابور
      // «زبون ما رد» بشاشة المحطات (NO_ANSWER_BEFORE/AFTER_CONFIRM)،
      // ومنه يرجّعه صاحب النظام يدوياً. ظهوره هنا وهناك سوا چان
      // يخلّي المنسّق يشتغل على حجز مؤجَّل قصداً.
      .getBookings({ status: ['PENDING', 'CONFIRMED'] })
      .then((all) => {
        // ═══ الي بدا التنفيذ ما يبقى بشاشة التنسيق ═══
        //
        // «المفروض من يتنسق ويستلمهن الليدر ويبدا التنفيذ بيهن يختفن
        // من هاي الواجهة — يرحن لحجوزات مكلفة».
        //
        // فلتر الحالة فوق يشيل IN_PROGRESS، بس الحجز يبقى CONFIRMED
        // ولو الليدر وصل الموقع وباشر (`arrivedAt`/`startedAt`) —
        // فيضل معروض هنا وكأنه ينتظر تنسيق، والمنسّق يروح يغيّر موعده
        // أو كادره وهم شغالين بيه فعلاً. `executionStarted` تشيله،
        // وسلّة «حجوزات مكلّفة» تستلمه بنفس الدالة.
        // ═══ ما يوصل هنا إلا الي انثبّت وانرحّل لكادر الشد ═══
        //
        // «أني ما أريد الحجوزات مباشرة تطلع بتنسيق الحجوزات. لازم
        // الحجوزات الجديدة الي ما متفقين وية الزبون تطلع بانتظار
        // التثبيت، ومن نضغط (تواصل وية الزبون) و(ترحيل لكادر الشد)
        // يله يترحّل لتنسيق الحجوزات».
        //
        // ⚠️ الحجز الجديد كان يوصل هنا **بلحظة تسجيله**: المنسّق
        // يشوف حجوزات ما أحد حچى وية زبونها، ويبدي يدوّر مواعيد
        // وكوادر لشغل يمكن ما يصير أصلاً.
        //
        const data = all.filter((b) => !executionStarted(b) && !!b.confirmedAt)
        setBookings(data)
        // حجز المشاريع **المفتوح** (وصل التنفيذ) ياخذ مرشحين مثل أي
        // حجز عادي — نفس كادر الشد هو الي راح ينفّذ. المقفول بس
        // ينستثنى لأنه لسه بإجراءات إدارة المشاريع.
        const candidates = data.filter((b) => b.status === 'CONFIRMED' && !b.projectLocked)
        // نجمّع الحجوزات حسب نفس الخدمة ونطلب موظفي المطابقة مرة وحدة لكل خدمة بدل
        // طلب منفصل لكل حجز (كان يسوي طلب HTTP مستقل لكل حجز بيها نفس الخدمة).
        // الحجز بلا خدمة محددة (مثلاً حجز صيانة) جان ينشال من هنا،
        // فما يوصله طلب مطابقة، فما تنعرض خانات الفنيين أصلاً وما
        // يكدر المنسّق يكلّف أحد. هسه يدخل بمفتاح فاضي ويجيب كل
        // الكوادر بلا فحص مهارة خدمة معيّنة.
        const byService = new Map<string, Booking[]>()
        for (const b of candidates) {
          const sid = b.service?.id || ''
          byService.set(sid, [...(byService.get(sid) || []), b])
        }
        byService.forEach((bs, sid) => loadMatches(bs, sid))
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])
  useEffect(() => {
    // بيانات مساعدة (مشرفين/منتجات/مركبات): منسّق الحجوزات ممكن ما عنده
    // صلاحية المركبات أو المنتجات — وقتها نخلي القائمة فارغة بدل ما ينكسر
    // كل الصفحة. بدون هذا الـcatch كان الرفض يطلع Unhandled rejection
    // ويطفّي الصفحة كلها، ويُحسب محاولة وصول غير مخوّلة على الموظف.
    api.getSupervisors().then(setSupervisors).catch(() => setSupervisors([]))
    api.getProducts().then(setProducts).catch(() => setProducts([]))
    api.getVehicleOptions().then(setVehicles).catch(() => setVehicles([]))
  }, [])

  // ═══ طلب حذف الحجز من شاشة التنسيق ═══
  //
  // «ضيفلي هنانه خيار طلب حذف الحجز».
  //
  // الحجز التجريبي أو المكرر أو الي الزبون تراجع عنه يوصل هنا مثل
  // أي حجز، والمنسّق يشوفه قدّامه — بس ما عنده وين يأشّر عليه. فيبقى
  // بالطابور يزاحم الشغل الحقيقي، أو المنسّق يفتح شاشة الحجوزات
  // يدوّره من جديد حتى يطلب حذفه.
  //
  // ⚠️ **طلب** حذف مو حذف: المراقب أو مدير النظام هو الي يبتّ. نفس
  // المسار الي بشاشة الحجوزات بالضبط — ما نسوي مسار ثاني للشغلة
  // الوحدة.
  const requestDelete = async (booking: Booking) => {
    const reason = prompt(`سبب طلب حذف الحجز ${booking.code}؟ (تجريبي، ملغى، مكرر...)`)
    if (!reason || !reason.trim()) return
    const channel = promptChoice('من وين اجه طلب الحذف؟', DELETE_CHANNEL_OPTIONS)
    if (!channel) return
    const requestType = promptChoice('شنو نوع الطلب؟', DELETE_TYPE_OPTIONS)
    if (!requestType) return
    // «الزبون ما رد» مسار مختلف تماماً: ما ينفتح طلب حذف ولا ينطر
    // قرار المراقب — الحجز ينزاح فوراً لطابور الانتظار الموجود أصلاً.
    if (requestType === BOOKING_NO_ANSWER_CHOICE) {
      try {
        await api.markBookingWaiting(booking.id, reason.trim())
      setSaveError(null)
      alert('تأشّر «الزبون ما رد» — الحجز انزاح لطابور الانتظار، ترجّعه يدوياً وقت ما تريد')
      } catch (e) {
      setSaveError(`تعذر تأشير «الزبون ما رد»: ${e instanceof Error ? e.message : 'خطأ غير متوقع'}`)
      }
      return
    }
    try {
      await api.requestBookingDelete(booking.id, reason.trim(), channel, requestType)
      setSaveError(null)
      alert('انرفع طلب الحذف — المراقب أو مدير النظام راح يبت بيه')
    } catch (e) {
      setSaveError(`تعذر رفع طلب الحذف: ${e instanceof Error ? e.message : 'خطأ غير متوقع'}`)
    }
  }

  // ═══ «تم الإنجاز» بدون تفاصيل — للمالك وحده ═══
  //
  // «هذني حجوزات قديمة احنا مشتغّليهن وما نعرف الكادر الي طلع ولا
  // التكلفة، فنريده ينكتب عليه تم الإنجاز بشكل كامل بدون تفاصيل،
  // وبعدين نكمل المحتاجيهن».
  //
  // ⚠️ هذا **مو** تأشير إنجاز عادي: الحجز ينعلّم «تسوية إدارية» حتى
  // ينستثنى من غرامات الفاتورة والتقرير. بدون العلامة، تنظيف
  // الطابور يتحوّل بعد ٢٤ ساعة لغرامات على ليدرات وإداريين، على شغل
  // صار قبل ما يوجد النظام ومحد يكدر يوثّقه.
  //
  // ⚠️ وما تنسجّل طلعة: ما نعرف منو طلع، وتسجيل الكادر الحالي يعني
  // إنتاجية مبنية على تخمين.
  // ⚠️ `role` مطبّع: جلسة الواجهة تحوّل دور المالك لـ'ADMIN' حتى
  // يشتغل كل شي مبني على `role === 'ADMIN'` تلقائياً، والدور الحقيقي
  // ينحفظ بـ`actualRole`. فحص `role === 'OWNER'` ما ينجح **أبداً** —
  // وهذا الي خلّى الزر ما يطلع للمالك نفسه.
  const isOwner = currentUser?.actualRole === 'OWNER' || currentUser?.role === 'OWNER'

  const settleLegacy = async (booking: Booking) => {
    if (!confirm(
      `تأشير الحجز ${booking.code} «تم الإنجاز بدون تفاصيل»؟\n\n`
      + 'راح ينقفل منجزاً بلا كادر ولا مبالغ، وينعلّم كتسوية إدارية '
      + '(مستثنى من غرامات الفاتورة والتقرير).',
    )) return
    const note = prompt('ملاحظة (اختيارية) — مثلاً: حجز قديم قبل النظام') || ''
    try {
      const updated = await api.settleLegacyBooking(booking.id, note.trim())
      setBookings((prev) => prev.filter((b) => b.id !== updated.id))
      setSaveError(null)
    } catch (e) {
      setSaveError(`تعذر التأشير: ${e instanceof Error ? e.message : 'خطأ غير متوقع'}`)
    }
  }

  const handleSupervisorChange = async (booking: Booking, employeeId: string) => {
    await applyUpdate('تعيين المشرف', () => api.assignSupervisor(booking.id, employeeId || null))
  }

  const handleScheduleChange = async (booking: Booking, value: string) => {
    if (!value) return
    await applyUpdate('تحديد الموعد', () => api.scheduleBooking(booking.id, value, currentUser?.id))
  }

  const handleDetailsBlur = async (
    booking: Booking,
    field: 'quotedPrice' | 'address' | 'mapLocation',
    value: string,
  ) => {
    if (field === 'quotedPrice') {
      const num = value === '' ? null : Number(value)
      if (num === booking.quotedPrice) return
      await applyUpdate('حفظ المبلغ المقدّر', () => api.updateBookingDetails(booking.id, { quotedPrice: num }))
    } else if (field === 'mapLocation') {
      if (value === (booking.mapLocation || '')) return
      let lat: number | null = null, lng: number | null = null
      let m = value.match(/@(-?\d+\.?\d*),(-?\d+\.?\d*)/)
      if (!m) m = value.match(/[?&]q=(-?\d+\.?\d*),(-?\d+\.?\d*)/)
      if (!m) m = value.match(/place\/(-?\d+\.?\d*),(-?\d+\.?\d*)/)
      if (!m) m = value.match(/^(-?\d+\.?\d*)\s*,\s*(-?\d+\.?\d*)$/)
      if (m) { lat = parseFloat(m[1]); lng = parseFloat(m[2]) }
      await applyUpdate('حفظ الموقع', () => api.updateBookingDetails(booking.id, {
        mapLocation: value,
        mapLatitude: lat,
        mapLongitude: lng,
      }))
    } else {
      if (value === (booking.address || '')) return
      await applyUpdate('حفظ العنوان', () => api.updateBookingDetails(booking.id, { address: value }))
    }
  }

  // تحديد الموقع مباشرة من الخريطة بدل الاعتماد على تحليل رابط ملصوق (اللي يفشل بكثير من روابط
  // خرائط كوكل المختصرة لأنها ما تحمل الإحداثيات أصلاً بالرابط نفسه)
  const handleMapPointChange = async (booking: Booking, point: { lat: number; lng: number } | null) => {
    if (!point) return
    await applyUpdate('حفظ نقطة الخريطة', () => api.updateBookingDetails(booking.id, {
      mapLatitude: point.lat,
      mapLongitude: point.lng,
    }))
  }

  // إلغاء تكليف موظف من خانة — الحجز يبقى مثبّت بلا كادر.
  const handleUnassign = async (booking: Booking, role: string, name: string) => {
    if (!confirm(`إلغاء تكليف «${name}» من هذا الحجز؟\nالحجز يبقى مثبّت، بس بلا كادر بهاي الخانة.`)) return
    await applyUpdate('إلغاء التكليف', () => api.unassignTechnician(booking.id, role))
  }

  const handleAssign = async (
    booking: Booking,
    role: 'TECH_1' | 'TECH_2' | 'TECH_3',
    employeeId: string,
  ) => {
    if (!employeeId) return
    // ⚠️ أخطر وحدة بالقائمة: فشلها الصامت يخلي الإداري يظن الفني
    // مكلّف، والفني ما يوصله شي — فالحجز يوصل يومه وماكو أحد رايح له.
    await applyUpdate('تكليف الفني', () => api.assignTechnician(booking.id, { employeeId, role }))
  }

  const handleVehicleChange = async (booking: Booking, assignedVehicle: string) => {
    if (assignedVehicle === (booking.assignedVehicle || '')) return
    await applyUpdate('تحديد المركبة', () => api.updateBookingDetails(booking.id, { assignedVehicle }))
  }

  // ═══ متابعة الإنجاز وحمل الليدرات ═══
  // الإداري ما جان يشوف منو خلّص شغله كامل ومنو أنجز وترك الورق وراه،
  // ولا جان يشوف كل ليدر شكد عنده حجوزات قبل ما يكلّفه — فيكلّف واحد
  // مزحوم وواحد ثاني فاضي.
  const [stateFilter, setStateFilter] = useState<string>('ALL')
  const [leaderFilter, setLeaderFilter] = useState<string>('ALL')

  // حمل كل ليدر: الحجوزات الي لسه شغالة عليه (مو منجزة ولا ملغاة)
  const leaderLoad = (() => {
    const map = new Map<string, { name: string; active: Booking[]; done: Booking[] }>()
    for (const b of bookings) {
      for (const a of b.assignments || []) {
        // الليدر بالحجز = الموظف المعيّن الي مؤشّر «تيم ليدر» بملفه
        // (الأدوار TECH_1/2/3 هي لبقية الكادر، مو للليدر).
        if (!a.employee?.isLeader) continue
        const id = a.employee?.id
        if (!id) continue
        if (!map.has(id)) map.set(id, { name: a.employee.name, active: [], done: [] })
        const entry = map.get(id)!
        if (b.status === 'COMPLETED' || b.status === 'CANCELLED') entry.done.push(b)
        else entry.active.push(b)
      }
    }
    return [...map.entries()]
      .map(([id, v]) => ({ id, ...v }))
      .sort((x, y) => x.active.length - y.active.length) // الأخف حمل أول — حتى يكلّفه
  })()

  // الحجوزات الي انتهت (منجزة أو متوقفة) — هاي الي إلها حالة اكتمال
  const finishedBookings = bookings
    .filter((b) => b.status === 'COMPLETED' || b.status === 'CANCELLED')
    .filter((b) => matchesSearch(b))
    .filter((b) => stateFilter === 'ALL' || b.completionState === stateFilter)
    .filter((b) =>
      leaderFilter === 'ALL' ||
      (b.assignments || []).some((a) => a.employee?.isLeader && a.employee?.id === leaderFilter),
    )

  const stateCounts = COMPLETION_ORDER.reduce<Record<string, number>>((acc, st) => {
    acc[st] = bookings.filter((b) => b.completionState === st).length
    return acc
  }, {})

  // بحث بكود الحجز، كود الزبون، رقم هاتفه، أو اسمه
  const matchesSearch = (b: Booking) => {
    // ⚠️ كان يبحث بـcustomer.code وهو ما ينرسل مع الحجز — يعني البحث
    // بكود الزبون جان ما يطابق ولا حجز أبداً، بصمت.
    return searchMatches([b.code, formatCustomerCode(b.customer), b.customer?.phone, b.customer?.name], search)
  }
  const confirmedBookings = bookings.filter((b) => b.status === 'CONFIRMED' && matchesSearch(b))

  // ═══ التثبيت شي والتنسيق شي ثاني ═══
  //
  // «لازم تنعزل الحجوزات المثبتة الي بيها كادر عن المثبتة بدون تنسيق —
  // هذا يرجع للحجوزات الي تحتاج تنسيق ينسقهن».
  //
  // قبل، الحل كان إجبار الإداري يحدد كادر وموعد وقت التثبيت. وهذا غلط
  // بالواقع: «الحجز بعد أسبوع — اني شمدريني بعد أسبوع منو موجود؟».
  // فيضطر يحط كادر عشوائي حتى تمر الشاشة، فيطلع تكليف كذب بالجدول،
  // والكادر الي انكتب اسمه ما يدري بيه أصلاً.
  //
  // فالتثبيت يضل حر: تواصلت مع الزبون واتفقتوا → ثبّت. أما الكادر
  // والموعد فينحطّون لمن يجي وقتهم، والحجز يضل بطابور التنسيق لحد
  // ذاك الوقت.
  //
  // ⚠️ الحجز المثبّت بلا كادر **ما يصير يطلع بجدول اليوم** — شاشة
  // الحجوزات تشيله (isCoordinated هناك) لأنه شغل ما انولد بعد.
  // عرضه ويّا شغل اليوم يخلي عدّ «شكد عدنا اليوم» كذب، والكادر
  // يتحضّر لحجز محد كلّفه بيه.
  const isFullyCoordinated = (b: Booking) =>
    !!b.scheduledAt && (b.assignments?.length ?? 0) > 0

  // المحتاجة تنسيق فوگ: هاي الي عليها شغل. المنسّقة تحت — خلصت.
  const confirmedNeedingCoordination = confirmedBookings.filter((b) => !isFullyCoordinated(b))
  const confirmedDone = confirmedBookings.filter(isFullyCoordinated)
  const orderedConfirmed = [...confirmedNeedingCoordination, ...confirmedDone]
  // ═══ كشف تعارض الموعد ═══
  //
  // «من أخلي وقت للحجز ينطيني تنبيه إنه اكو حجز بهذا الوقت فقط حتى
  // أعرف، أو يقترحلي أوقات متاحة».
  //
  // ⚠️ تنبيه مو منع: يجوز الإداري **يقصد** يحط حجزين بنفس الساعة
  // (كادرين مختلفين، أو زبونين بنفس البناية). المنع يخليه يدور على
  // حيلة يتجاوزها، والتنبيه يخليه يقرر وهو شايف.
  const conflictsFor = (bookingId: string, value: string) => {
    if (!value) return []
    const t = new Date(value).getTime()
    if (Number.isNaN(t)) return []
    const HOUR = 60 * 60 * 1000
    return bookings.filter((b) => {
      if (b.id === bookingId || !b.scheduledAt) return false
      if (b.status === 'CANCELLED' || b.status === 'COMPLETED') return false
      // نفس الساعة تقريباً — فرق أقل من ساعة يعتبر تعارض
      return Math.abs(new Date(b.scheduledAt).getTime() - t) < HOUR
    })
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">تنسيق الحجوزات (الإداري)</h2>

      {/* ⚠️ الشريط لاصق بأعلى الشاشة (sticky): الفشل يصير غالباً وأنت
          بنص القائمة عند حجز بعيد عن الرأس، ورسالة تطلع فوگ برّا مجال
          النظر ما تنقرا — يعني نرجع لنفس الفشل الصامت. */}
      {saveError && (
        <div className="sticky top-2 z-20 mt-3 flex items-start justify-between gap-3 rounded-xl border-2 border-red-300 bg-red-50 px-4 py-3 shadow-lg">
          <div>
            <p className="text-sm font-extrabold text-red-800">⚠️ {saveError}</p>
            <p className="mt-0.5 text-xs text-red-700">
              التغيير <b>ما انحفظ</b> — جرّب مرة ثانية. إذا تكرر، دز صورة الرسالة للدعم.
            </p>
          </div>
          <button
            onClick={() => setSaveError(null)}
            className="shrink-0 rounded-lg border border-red-300 bg-white px-2.5 py-1 text-xs font-bold text-red-700"
          >
            إخفاء
          </button>
        </div>
      )}
      <p className="mt-1 text-slate-500">
        ثبّت الحجز مع الزبون مع تحديد التكلفة والعنوان، ثم وجّهه لكادر الشد أو لإدارة المشاريع وحدد
        الفنيين المتاحين (اختياري الآن، يمكن تحديدهم لاحقاً).
      </p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>
      )}

      {!loading && !error && (
        <>
          {/* ⚠️ جدول «المواعيد المحجوزة» انشال من هنا.
              كان يعرض كل المواعيد القادمة بأعلى الشاشة — عشرات الصفوف
              يقراهن الإداري **بعينه** ويقارنهن بالوقت الي بذهنه، وبعدها
              يمرّر نص شاشة حتى يوصل الحجز الي يشتغل عليه.
              بدله: التحذير يجي **بلحظة اختيار الوقت** لهذا الحجز
              بالذات، ويگول منو الزبون المتعارض ويقترح بديل. */}

          {/* ═══ حمل الليدرات: منو فاضي ومنو مزحوم ═══
              الإداري ما يكلّف بالحدس — يشوف كل ليدر شكد عنده حجوزات
              شغالة وأي حجوزات، والقائمة مرتبة من الأخف حملاً. */}
          <div className="mt-6 rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <h3 className="mb-3 text-sm font-bold text-[#0f2040]">👷 حمل الليدرات — منو تكدر تكلّفه</h3>
            {leaderLoad.length === 0 && <p className="text-xs text-slate-400">ماكو ليدرات مكلّفين بحجوزات حالياً.</p>}
            <div className="grid grid-cols-1 gap-2 sm:grid-cols-2 lg:grid-cols-3">
              {leaderLoad.map((l) => (
                <button
                  key={l.id}
                  type="button"
                  onClick={() => setLeaderFilter(leaderFilter === l.id ? 'ALL' : l.id)}
                  className={`rounded-xl border p-3 text-right transition-all ${
                    leaderFilter === l.id ? 'border-brand-500 bg-brand-50' : 'border-slate-200 bg-white hover:border-slate-300'
                  }`}
                >
                  <div className="flex items-center justify-between gap-2">
                    <span className="text-sm font-bold text-slate-800">{l.name}</span>
                    <span
                      className={`rounded-full px-2 py-0.5 text-[11px] font-bold ${
                        l.active.length === 0
                          ? 'bg-emerald-50 text-emerald-700'
                          : l.active.length <= 2
                            ? 'bg-sky-50 text-sky-700'
                            : 'bg-red-50 text-red-700'
                      }`}
                    >
                      {l.active.length === 0 ? 'فاضي' : `${l.active.length} حجز شغّال`}
                    </span>
                  </div>
                  {l.active.length > 0 && (
                    <p className="mt-1 text-[11px] leading-relaxed text-slate-500">
                      {l.active.map((b) => b.code).join('، ')}
                    </p>
                  )}
                  <p className="mt-0.5 text-[10px] text-slate-400">أنجز {l.done.length} حجز</p>
                </button>
              ))}
            </div>
          </div>

          {/* ═══ متابعة الإنجاز ═══ */}
          <div className="mt-4 rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <h3 className="mb-3 text-sm font-bold text-[#0f2040]">📋 متابعة الإنجاز — منو خلّص ورقه ومنو لا</h3>
            <div className="flex flex-wrap gap-2">
              <button
                type="button"
                onClick={() => setStateFilter('ALL')}
                className={`rounded-full border px-3 py-1 text-[11px] font-bold ${
                  stateFilter === 'ALL' ? 'border-brand-500 bg-brand-50 text-brand-800' : 'border-slate-200 text-slate-500'
                }`}
              >
                الكل
              </button>
              {COMPLETION_ORDER.map((st) => (
                <button
                  key={st}
                  type="button"
                  onClick={() => setStateFilter(stateFilter === st ? 'ALL' : st)}
                  className={`rounded-full border px-3 py-1 text-[11px] font-bold ${
                    stateFilter === st ? 'border-brand-500 bg-brand-50 text-brand-800' : 'border-slate-200 text-slate-500'
                  }`}
                >
                  {completionLabel(st)} ({stateCounts[st] ?? 0})
                </button>
              ))}
            </div>
            {(stateFilter !== 'ALL' || leaderFilter !== 'ALL') && (
              <div className="mt-3 space-y-1.5">
                {finishedBookings.length === 0 && (
                  <p className="text-xs text-slate-400">ماكو حجوزات بهذي الحالة.</p>
                )}
                {finishedBookings.map((b) => (
                  <div key={b.id} className="flex flex-wrap items-center gap-2 rounded-lg border border-slate-100 px-3 py-2">
                    {/* الهوية الكاملة بدل الكود+الاسم: المنسّق يتصل
                        بالزبون من هنا مباشرة بلا ما يفتح الحجز. */}
                    <EntityIdentity booking={b} className="flex-1 border-0 bg-transparent px-0 py-0" />
                    <span className="mr-auto"><CompletionBadge booking={b} /></span>
                    {/* قصة الحجز والأوقات — تنفتح بالطلب حتى ما نجيب
                        خط زمني لكل حجز بالقائمة. */}
                    <details className="w-full">
                      <summary className="cursor-pointer text-xs font-bold text-brand-700">🕒 شوف قصة الحجز والأوقات</summary>
                      <BookingTimelineView bookingId={b.id} />
                    </details>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* بحث بكود الحجز أو كود الزبون أو رقمه أو اسمه */}
          <div className="mt-6 rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="🔍 بحث بكود الحجز، كود الزبون، رقم الهاتف، أو اسم الزبون..."
              className="w-full rounded-lg border border-slate-300 px-4 py-2.5 outline-none focus:border-brand-500"
            />
            {search.trim() && (
              <p className="mt-2 text-xs text-slate-500">
                النتائج: {confirmedBookings.length} حجز مثبّت
              </p>
            )}
          </div>

          {/* ═══ «بانتظار التثبيت» انشالت من هنا ═══
              «الحجوزات جاي تعبر بالمراحل — عندي حجز لم ينسق بعد،
              شعدنا يطلع بتم التثبيت؟ أريدك تلتزم بالآلية الصحيحة».

              هذا القسم كان **باب ثاني** يثبّت الحجز من شاشة التنسيق
              مباشرة، فيتخطى البوابة الي اتفقنا عليها: تواصل وية
              الزبون ← ثبّت ورحّل. يعني حجز يوصل «تم التثبيت» بلا ما
              يمر بـ«بانتظار التثبيت» ولا أحد حچى وية زبونه.

              التثبيت صار بمكان واحد: محطة «بانتظار التثبيت». وشاشة
              التنسيق تستلم **المثبّت** بس — وهذا الي يخلّي المراحل
              تمشي وحدة بعد وحدة بدل ما تتقافز. */}
          {/* تم تثبيتها — مقسومة: المحتاجة تنسيق فوگ، المنسّقة تحت */}
          <h3 className="mt-8 mb-1 text-lg font-bold text-brand-800">
            تم تثبيتها ({confirmedBookings.length})
          </h3>
          <p className="mb-3 text-xs text-slate-500">
            {confirmedNeedingCoordination.length > 0 ? (
              <>
                <span className="font-bold text-amber-700">
                  ⏳ {confirmedNeedingCoordination.length} تحتاج تنسيق
                </span>
                {' '}(بلا كادر أو بلا موعد — ما تطلع بجدول اليوم لحد ما تنسّقها)
                {confirmedDone.length > 0 && <> · <span className="font-bold text-emerald-700">✅ {confirmedDone.length} منسّقة</span></>}
              </>
            ) : (
              <span className="font-bold text-emerald-700">✅ كلها منسّقة — كادر وموعد محددين</span>
            )}
          </p>
          <div className="flex flex-col gap-4">
            {orderedConfirmed.map((booking, idx) => (
              <div key={booking.id}>
              {/* فاصل بصري بين الي محتاجة شغل والي خلصت */}
              {idx === confirmedNeedingCoordination.length && confirmedNeedingCoordination.length > 0 && (
                <div className="mb-4 flex items-center gap-3">
                  <div className="h-px flex-1 bg-emerald-200" />
                  <span className="rounded-full bg-emerald-50 px-3 py-1 text-xs font-bold text-emerald-700">
                    ✅ منسّقة بالكامل ({confirmedDone.length})
                  </span>
                  <div className="h-px flex-1 bg-emerald-200" />
                </div>
              )}
              <div
                className={`rounded-xl border bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)] ${
                  isFullyCoordinated(booking) ? 'border-emerald-200' : 'border-amber-300 bg-amber-50/30'
                }`}
              >
                {/* شريط يگول بالضبط شنو ناقص — بدونه الإداري ما يعرف
                    ليش هذا الحجز مو بجدول اليوم */}
                {!isFullyCoordinated(booking) && (
                  <div className="mb-3 rounded-lg border border-amber-300 bg-amber-50 px-3 py-2 text-xs font-bold text-amber-800">
                    ⏳ مثبّت بس يحتاج تنسيق — ناقص{' '}
                    {!booking.scheduledAt && (booking.assignments?.length ?? 0) === 0
                      ? 'الموعد والكادر'
                      : !booking.scheduledAt
                        ? 'الموعد'
                        : 'الكادر'}
                    <span className="mr-1 font-normal text-amber-700">
                      · ما راح يطلع بجدول اليوم ولا يوصل الكادر لحد ما تكمّله
                    </span>
                  </div>
                )}
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <span className="font-mono text-sm font-semibold text-brand-600">
                      {booking.code}
                    </span>
                    <span className="mr-3 rounded-full bg-emerald-100 px-3 py-1 text-xs font-bold text-emerald-700">
                      تم التثبيت
                    </span>
                    {booking.priority === 'URGENT' && (
                      <span className="mr-2 rounded-full bg-red-100 px-3 py-1 text-xs font-bold text-red-700">
                        عاجل
                      </span>
                    )}
                    {booking.status === 'CONFIRMED' && booking.materialsReadyAt && (
                      <span className="mr-2 rounded-full bg-amber-100 px-3 py-1 text-xs font-bold text-amber-700">
                        ⏰ المواد جاهزة، بانتظار انطلاق الفريق ({booking.materialsReadyBy?.name || '-'})
                      </span>
                    )}
                  </div>
                  <div className="flex items-center gap-2">
                    <span className="text-sm text-slate-500">{serviceNames(booking)}</span>
                    {isOwner && (
                      <button
                        onClick={() => settleLegacy(booking)}
                        className="rounded-full bg-emerald-700 px-3 py-1 text-xs font-bold text-white transition-colors hover:bg-emerald-800"
                        title="حجز قديم ما نعرف كادره ولا تكلفته — ينقفل منجزاً بلا تفاصيل، ومستثنى من الغرامات"
                      >
                        ✅ تم الإنجاز (بدون تفاصيل)
                      </button>
                    )}
                    {canRequestDelete && (
                      <button
                        onClick={() => requestDelete(booking)}
                        className="rounded-full bg-red-50 px-3 py-1 text-xs font-bold text-red-700 transition-colors hover:bg-red-100"
                      >
                        🗑️ اطلب حذف الحجز
                      </button>
                    )}
                  </div>
                </div>

                <div className="mt-3 grid grid-cols-1 gap-2 text-sm sm:grid-cols-3">
                  <div>
                    <span className="text-slate-400">الزبون: </span>
                    <span className="font-medium text-brand-800">{booking.customer?.name || 'زبون غير معروف'}</span>
                  </div>
                  <div>
                    <span className="text-slate-400">الهاتف: </span>
                    {booking.customer?.phone || '-'}
                  </div>
                  <div>
                    <span className="text-slate-400">الموقع المسجل: </span>
                    {booking.customer?.location || '-'}
                  </div>
                </div>

                {/* تعديل الخدمة والسعر والموعد سوه — الخدمة ما جان بيها
                    تعديل أصلاً، فالزبون يغيّر طلبه والإداري ما يكدر
                    يعكسه بالحجز. */}
                <BookingEditPanel
                  booking={booking}
                  onSaved={(u) => setBookings((prev) => prev.map((x) => (x.id === u.id ? u : x)))}
                />

                {/* قرارات الزبون: أجّل، ما رد، ألغى.
                    مخفية للمقفول — السيرفر يرفضها أصلاً، فعرض أزرار
                    تطلّع خطأ بس يضيّع وقت المنسّق. */}
                {!booking.projectLocked && (
                  <BookingLifecycleActions
                    booking={booking}
                    canArchive={isAdmin}
                    onChanged={(u) => setBookings((prev) => prev.map((x) => (x.id === u.id ? u : x)))}
                  />
                )}

                <div className="mt-3 grid grid-cols-1 gap-3 rounded-lg bg-slate-50 p-4 sm:grid-cols-2">
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      كلفة العمل التقديرية (بدون المواد)
                    </label>
                    <input
                      type="number"
                      placeholder="غير محددة"
                      defaultValue={booking.quotedPrice ?? ''}
                      onBlur={(e) => handleDetailsBlur(booking, 'quotedPrice', e.target.value)}
                      className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                    {(cartItems[booking.id]?.length ?? 0) > 0 && (
                      <p className="mt-1 text-xs text-slate-400">
                        + كلفة المواد من السلة: {(cartItems[booking.id] || []).reduce((sum, i) => sum + i.totalPrice, 0).toLocaleString()} د.ع (تلقائي)
                      </p>
                    )}
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      عنوان تنفيذ المهمة
                    </label>
                    <input
                      placeholder="غير محدد"
                      defaultValue={booking.address || ''}
                      onBlur={(e) => handleDetailsBlur(booking, 'address', e.target.value)}
                      className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
                  </div>
                  <div className="sm:col-span-2">
                    <div className="mb-1 flex items-center justify-between">
                      <label className="block text-sm font-medium text-slate-600">موقع تنفيذ المهمة على الخريطة</label>
                      <button
                        type="button"
                        onClick={() => setEditingLocationId(editingLocationId === booking.id ? null : booking.id)}
                        className="rounded-lg bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700 hover:bg-brand-100"
                      >
                        {editingLocationId === booking.id ? 'إخفاء الخريطة' : 'تعديل الموقع على الخريطة'}
                      </button>
                    </div>
                    {booking.mapLatitude != null && booking.mapLongitude != null ? (
                      <a
                        href={`https://www.openstreetmap.org/?mlat=${booking.mapLatitude}&mlon=${booking.mapLongitude}#map=17/${booking.mapLatitude}/${booking.mapLongitude}`}
                        target="_blank"
                        rel="noreferrer"
                        className="inline-block text-xs text-brand-500 hover:underline"
                      >
                        فتح الموقع المحدد حالياً على الخريطة
                      </a>
                    ) : (
                      <p className="text-xs text-slate-400">ما تحدد موقع لهذا الحجز بعد</p>
                    )}
                    {editingLocationId === booking.id && (
                      <div className="mt-2">
                        <LocationPicker
                          value={booking.mapLatitude != null && booking.mapLongitude != null
                            ? { lat: booking.mapLatitude, lng: booking.mapLongitude }
                            : null}
                          onChange={(point) => handleMapPointChange(booking, point)}
                        />
                      </div>
                    )}
                  </div>
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      الموعد المحدد
                    </label>
                    <div className="flex gap-2">
                      {scheduleMode[`c_${booking.id}`] === 'manual' ? (
                        <input
                          type="datetime-local"
                          defaultValue={booking.scheduledAt ? toLocalInput(booking.scheduledAt) : ''}
                          onBlur={(e) => handleScheduleChange(booking, e.target.value)}
                          className="flex-1 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                        />
                      ) : (
                        <select
                          defaultValue=""
                          onChange={(e) => { if (e.target.value) handleScheduleChange(booking, e.target.value) }}
                          className="flex-1 rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                        >
                          <option value="">
                            {booking.scheduledAt
                              ? `الحالي: ${new Date(booking.scheduledAt).toLocaleString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'numeric', hour: '2-digit', minute: '2-digit' })}`
                              : '-- اختر موعد --'}
                          </option>
                          {getAvailableSlots(booking.id).map(s => (
                            <option key={s.value} value={s.value} disabled={s.label.includes('محجوز')}>
                              {s.label}
                            </option>
                          ))}
                        </select>
                      )}
                      <button
                        type="button"
                        onClick={() => setScheduleMode(prev => ({ ...prev, [`c_${booking.id}`]: prev[`c_${booking.id}`] === 'manual' ? 'slots' : 'manual' }))}
                        className="rounded-lg border border-slate-300 px-3 py-2 text-xs text-slate-600 transition-colors hover:bg-slate-50"
                      >
                        {scheduleMode[`c_${booking.id}`] === 'manual' ? 'المواعيد' : 'يدوي'}
                      </button>
                    </div>

                    {/* ═══ تنبيه التعارض ═══
                        «خلي من أخلي وقت للحجز ينطيني تنبيه إنو أكو حجز
                        بهذا الوقت».
                        ⚠️ كان معلّقاً بقسم «بانتظار التثبيت» الي
                        انشال، فانفقد وياه. رجّعته لخانة الموعد
                        الفعّالة — هاي المكان الي ينحدد بيه الموعد
                        فعلاً هسه. */}
                    {(() => {
                      const clash = conflictsFor(booking.id, booking.scheduledAt || '')
                      if (clash.length === 0) return null
                      return (
                        <p className="mt-1.5 rounded-lg bg-amber-50 px-2.5 py-1.5 text-[11px] font-bold text-amber-800">
                          ⚠️ أكو {clash.length} حجز بنفس الوقت تقريباً:{' '}
                          {clash.slice(0, 3).map((c) => c.code).join('، ')}
                          {clash.length > 3 && ` +${clash.length - 3}`}
                        </p>
                      )
                    })()}
                    {booking.scheduleLogs?.length > 0 && (
                      <details className="mt-1">
                        <summary className="cursor-pointer text-xs text-slate-400 hover:text-slate-600">سجل التعديلات ({booking.scheduleLogs.length})</summary>
                        <div className="mt-1 max-h-32 space-y-1 overflow-y-auto rounded-lg bg-slate-50 p-2">
                          {booking.scheduleLogs.map(log => (
                            <p key={log.id} className="text-[11px] text-slate-500">
                              <span className="font-bold text-slate-700">{log.changedBy.name}</span>
                              {' غيّر من '}
                              <span className="text-red-500">{log.oldTime ? new Date(log.oldTime).toLocaleString('ar-IQ') : 'بدون موعد'}</span>
                              {' إلى '}
                              <span className="text-emerald-600">{new Date(log.newTime).toLocaleString('ar-IQ')}</span>
                              {' — '}
                              <span className="text-slate-400">{new Date(log.createdAt).toLocaleString('ar-IQ')}</span>
                            </p>
                          ))}
                        </div>
                      </details>
                    )}
                  </div>
                </div>

                {/* ═══ حجز عند إدارة المشاريع ═══
                    قبل، الحجز يختفي من دنيا المنسّق ويطلع سطر ميّت «تم
                    التحويل» وبس — ما يعرف وين وصل ولا متى يرجع له، مع
                    إن **نفس كادر الشد** هو الي راح ينفّذ.
                    هسه يضل ظاهر: مقفول لحد ما المشروع يوصل «٥. البدء
                    بالتنفيذ»، وبعدها ينفتح بنفس كتلة التوجيه العادية. */}
                {booking.projectLocked ? (
                  <div className="mt-4 rounded-lg border border-amber-300 bg-amber-50 px-4 py-3">
                    <p className="text-sm font-bold text-amber-900">🔒 حجز مشاريع — بانتظار إكمال الإجراءات</p>
                    <p className="mt-1 text-xs text-amber-800">
                      عند إدارة المشاريع (اتصال، كشف، عرض سعر، عقد). أول ما يوصل مرحلة «البدء بالتنفيذ»
                      راح ينفتح هنا وتنسّقه بكادر الشد مثل أي حجز، وراح يوصلك إشعار.
                    </p>
                  </div>
                ) : (
                  <div className="mt-4">
                    {booking.transferToProjects && (
                      <div className="mb-3 rounded-lg border border-emerald-300 bg-emerald-50 px-4 py-2">
                        <p className="text-sm font-bold text-emerald-900">🏗️ حجز مشاريع وصل مرحلة التنفيذ — جاهز للتنسيق</p>
                      </div>
                    )}
                    <h4 className="text-sm font-bold text-brand-800">توجيه كادر الشد</h4>
                    <p className="mt-1 text-xs text-slate-400">
                      تحديد الكادر والسيارة اختياري - يمكن تثبيت الحجز وتحديدهم لاحقاً.
                    </p>

                    <div className="mt-2 grid grid-cols-1 gap-3 sm:grid-cols-2">
                      <div>
                        <label className="mb-1 block text-sm font-medium text-slate-600">
                          تيم ليدر (اختياري)
                        </label>
                        <select
                          value={booking.projectSupervisor?.id || ''}
                          onChange={(e) => handleSupervisorChange(booking, e.target.value)}
                          className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                        >
                          <option value="">-- بدون تيم ليدر --</option>
                          {supervisors.map((s) => (
                            <option key={s.id} value={s.id}>
                              {s.name}
                            </option>
                          ))}
                        </select>
                      </div>
                      {booking.expenseResponsible && (
                        <div className="flex items-center gap-2 rounded-lg border border-emerald-200 bg-emerald-50 px-3 py-2 self-end">
                          <span className="text-xs text-emerald-700">المسؤول عن المصاريف:</span>
                          <span className="text-xs font-bold text-emerald-800">{booking.expenseResponsible.name}</span>
                        </div>
                      )}
                    </div>

                    {booking.systemType && (
                      <div className="mt-2">
                        {durationEstimates[estimateKeyOf(booking).key]?.expectedMinutes != null ? (
                          <div className="rounded-lg border border-brand-200 bg-brand-50 px-3 py-2 text-xs font-bold text-brand-700">
                            ⏱️ الوقت المتوقع (متعلَّم من {durationEstimates[estimateKeyOf(booking).key].sampleCount} عيّنة سابقة):{' '}
                            {Math.round(durationEstimates[estimateKeyOf(booking).key].expectedMinutes!)} دقيقة
                          </div>
                        ) : (
                          <div className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-500">
                            ⏳ لا توجد بيانات كافية بعد ({durationEstimates[estimateKeyOf(booking).key]?.sampleCount ?? 0}/
                            {durationEstimates[estimateKeyOf(booking).key]?.minSamples ?? 5} عينات)
                          </div>
                        )}
                      </div>
                    )}

                    {/* الخانات تنعرض دائماً — قبل، لو ما وصلت قائمة
                        المطابقة (حجز بلا خدمة أو طلب فشل) تختفي خانات
                        الفنيين كلها وما يكدر يكلّف أحد. */}
                    {(
                      <div className="mt-2 grid grid-cols-1 gap-3 sm:grid-cols-3">
                        {techRoles.map((tr) => {
                          const assigned = booking.assignments.find((a) => a.role === tr.key)
                          const candidates = matches[booking.id] || []
                          return (
                            <div key={tr.key}>
                              <div className="mb-1 flex items-center justify-between gap-2">
                                <label className="block text-sm font-medium text-slate-600">
                                  {tr.label} (اختياري)
                                </label>
                                {/* ⚠️ زر الإلغاء ما كان موجود: القائمة
                                    تبدّل الموظف بس، فالإداري الي يكلّف
                                    بالغلط ما عنده أي طريق يشيله — يضطر
                                    يحط واحد ثاني والحجز يبقى بكادر
                                    ما يخصه. */}
                                {assigned && (
                                  <button
                                    onClick={() => handleUnassign(booking, tr.key, assigned.employee.name)}
                                    className="rounded-lg px-2 py-0.5 text-[11px] font-bold text-red-600 hover:bg-red-50"
                                  >
                                    ✖ إلغاء التكليف
                                  </button>
                                )}
                              </div>
                              <select
                                value={assigned?.employee.id || ''}
                                onChange={(e) => handleAssign(booking, tr.key, e.target.value)}
                                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                              >
                                <option value="">-- اختر فني --</option>
                                {candidates.length === 0 && (
                                  <option value="" disabled>
                                    {matches[booking.id] ? 'ماكو كادر متاح' : 'جاري تحميل الكوادر...'}
                                  </option>
                                )}
                                {candidates.map((c) => (
                                  <option key={c.id} value={c.id}>
                                    {c.hasRequiredSkill === false ? `⚠️ ${c.name} (لا يمتلك المهارة المطلوبة)` : `${c.name} (${c.position || 'فني'})`}
                                  </option>
                                ))}
                              </select>
                            </div>
                          )
                        })}
                      </div>
                    )}
                    <div className="mt-3 sm:w-1/3">
                      <label className="mb-1 block text-sm font-medium text-slate-600">
                        السيارة المخصصة للمهمة (اختياري)
                      </label>
                      <select
                        value={booking.assignedVehicle || ''}
                        onChange={(e) => handleVehicleChange(booking, e.target.value)}
                        className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                      >
                        <option value="">-- اختر سيارة --</option>
                        {vehicles.map((v) => (
                          <option key={v.id} value={`${v.name} - ${v.plateNumber}`}>
                            {v.name} ({v.plateNumber}){v.color ? ` - ${v.color}` : ''}
                          </option>
                        ))}
                      </select>
                    </div>
                    {booking.assignments.some((a) => {
                      const candidate = matches[booking.id]?.find((c) => c.id === a.employee.id)
                      return candidate && candidate.hasRequiredSkill === false
                    }) && (
                      <p className="mt-2 rounded-lg bg-amber-50 border border-amber-200 px-4 py-2 text-sm text-amber-700">
                        ⚠️ تحذير: الفني المختار لا يمتلك المهارة المطلوبة لهذه الخدمة
                      </p>
                    )}
                  </div>
                )}

                {/* سلة المنتجات */}
                <div className="mt-4 border-t border-slate-100 pt-4">
                  <button
                    onClick={() => toggleCart(booking.id)}
                    className="flex items-center gap-2 rounded-lg bg-gradient-to-l from-violet-500 to-violet-700 px-4 py-2 text-sm font-medium text-white shadow-md transition-all hover:shadow-lg"
                  >
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>
                    سلة الزبون
                    {cartItems[booking.id]?.length ? (
                      <span className="rounded-full bg-white/30 px-2 py-0.5 text-xs">{cartItems[booking.id].length}</span>
                    ) : null}
                  </button>

                  {cartOpen[booking.id] && (
                    <div className="mt-3 rounded-xl border border-violet-100 bg-violet-50/50 p-4">
                      {/* قائمة العناصر */}
                      {(cartItems[booking.id] || []).length > 0 && (
                        <div className="mb-4 overflow-x-auto">
                          <table className="w-full text-sm">
                            <thead>
                              <tr className="border-b border-violet-200 text-right text-xs text-violet-600">
                                <th className="pb-2 pr-2">المنتج</th>
                                <th className="pb-2 pr-2">الكمية</th>
                                <th className="pb-2 pr-2">سعر الوحدة</th>
                                <th className="pb-2 pr-2">المجموع</th>
                                <th className="pb-2 pr-2">ملاحظات</th>
                                <th className="pb-2"></th>
                              </tr>
                            </thead>
                            <tbody>
                              {(cartItems[booking.id] || []).map(item => (
                                <tr key={item.id} className="border-b border-violet-100">
                                  <td className="py-2 pr-2 font-medium text-brand-800">{item.productName}</td>
                                  <td className="py-2 pr-2">{item.quantity}</td>
                                  <td className="py-2 pr-2">{item.unitPrice.toLocaleString()}</td>
                                  <td className="py-2 pr-2 font-bold text-violet-700">{item.totalPrice.toLocaleString()}</td>
                                  <td className="py-2 pr-2 text-xs text-slate-500">{item.notes || '-'}</td>
                                  <td className="py-2">
                                    <button
                                      onClick={() => removeCartItem(booking.id, item.id)}
                                      className="rounded p-1 text-red-400 transition-colors hover:bg-red-50 hover:text-red-600"
                                    >
                                      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M3 6h18"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                                    </button>
                                  </td>
                                </tr>
                              ))}
                            </tbody>
                            <tfoot>
                              <tr className="font-bold text-violet-800">
                                <td colSpan={3} className="pt-3 pr-2">التكلفة الكلية</td>
                                <td className="pt-3 pr-2">{(cartItems[booking.id] || []).reduce((sum, i) => sum + i.totalPrice, 0).toLocaleString()}</td>
                                <td colSpan={2}></td>
                              </tr>
                            </tfoot>
                          </table>
                        </div>
                      )}

                      {/* إضافة عنصر جديد */}
                      <div className="grid grid-cols-2 gap-2 sm:grid-cols-5">
                        <div className="relative">
                          <select
                            value={cartForm[booking.id]?.productName || ''}
                            onChange={e => {
                              const prod = products.find(p => p.name === e.target.value)
                              setCartForm(prev => ({
                                ...prev,
                                [booking.id]: {
                                  ...prev[booking.id] || { productName: '', quantity: '', unitPrice: '', notes: '' },
                                  productName: e.target.value,
                                  unitPrice: prod?.defaultPrice ? String(prod.defaultPrice) : prev[booking.id]?.unitPrice || '',
                                },
                              }))
                            }}
                            className="w-full rounded-lg border border-violet-200 bg-white px-3 py-2 text-sm outline-none focus:border-violet-500"
                          >
                            <option value="">-- اختر منتج --</option>
                            {products.map(p => (
                              <option key={p.id} value={p.name}>{p.name} {p.defaultPrice ? `(${p.defaultPrice.toLocaleString()})` : ''}</option>
                            ))}
                          </select>
                          {(() => {
                            const selProd = products.find(p => p.name === cartForm[booking.id]?.productName)
                            return selProd?.imageBase64 ? (
                              <img src={selProd.imageBase64} className="mt-1 h-12 w-12 rounded border border-violet-200 object-contain bg-white" />
                            ) : null
                          })()}
                        </div>
                        <input
                          type="number"
                          placeholder="الكمية"
                          value={cartForm[booking.id]?.quantity || ''}
                          onChange={e => setCartForm(prev => ({ ...prev, [booking.id]: { ...prev[booking.id] || { productName: '', quantity: '', unitPrice: '', notes: '' }, quantity: e.target.value } }))}
                          className="rounded-lg border border-violet-200 bg-white px-3 py-2 text-sm outline-none focus:border-violet-500"
                        />
                        <input
                          type="number"
                          placeholder="سعر الوحدة"
                          value={cartForm[booking.id]?.unitPrice || ''}
                          onChange={e => setCartForm(prev => ({ ...prev, [booking.id]: { ...prev[booking.id] || { productName: '', quantity: '', unitPrice: '', notes: '' }, unitPrice: e.target.value } }))}
                          className="rounded-lg border border-violet-200 bg-white px-3 py-2 text-sm outline-none focus:border-violet-500"
                        />
                        <input
                          placeholder="ملاحظات (اختياري)"
                          value={cartForm[booking.id]?.notes || ''}
                          onChange={e => setCartForm(prev => ({ ...prev, [booking.id]: { ...prev[booking.id] || { productName: '', quantity: '', unitPrice: '', notes: '' }, notes: e.target.value } }))}
                          className="rounded-lg border border-violet-200 bg-white px-3 py-2 text-sm outline-none focus:border-violet-500"
                        />
                        <button
                          onClick={() => addCartItem(booking.id)}
                          className="rounded-lg bg-violet-600 px-4 py-2 text-sm font-medium text-white transition-colors hover:bg-violet-700"
                        >
                          إضافة
                        </button>
                      </div>
                    </div>
                  )}
                </div>
              </div>
              </div>
            ))}
            {confirmedBookings.length === 0 && (
              <p className="text-slate-400">لا توجد حجوزات مثبتة بعد.</p>
            )}
          </div>
        </>
      )}
    </div>
  )
}
