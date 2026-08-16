import { Fragment, useEffect, useRef, useState } from 'react'
import { api, type Booking, type Employee, type VehicleOption } from '../api'
import { formatCustomerCode } from '../utils/identity'
import { useSession } from '../session'
import { MapViewer } from '../components/MapLazy'
import { formatScheduleWindow } from '../utils/schedule'
import CompletionBadge from '../components/CompletionBadge'
import BookingEditPanel from '../components/BookingEditPanel'
import BookingVisits from '../components/BookingVisits'
import { matches } from '../utils/search'
import { BOOKING_STAGES, currentStage, executionStarted } from '../bookingStage'
import { BUCKET_HEADINGS, DONE_FILTERS, type BookingBucket, type DoneFilter } from './bookingBuckets'

export type { BookingBucket } from './bookingBuckets'

// أسماء كل خدمات الحجز (الزبون ممكن يطلب أكثر من منظومة بنفس الحجز)
function serviceNames(b: { service?: { name: string } | null; services?: { name: string }[] }): string {
  if (b.services && b.services.length > 0) return b.services.map((s) => s.name).join(' + ')
  return b.service?.name || 'بدون خدمة محددة'
}

const techRoles: { key: 'TECH_1' | 'TECH_2' | 'TECH_3'; label: string }[] = [
  { key: 'TECH_1', label: 'الفني الأول' },
  { key: 'TECH_2', label: 'الفني الثاني' },
  { key: 'TECH_3', label: 'الفني الثالث' },
]

function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function addDays(dateStr: string, delta: number) {
  const d = new Date(`${dateStr}T00:00:00`)
  d.setDate(d.getDate() + delta)
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function formatDateArabic(dateStr: string) {
  return new Date(`${dateStr}T00:00:00`).toLocaleDateString('ar-IQ', { weekday: 'long', year: 'numeric', month: 'long', day: 'numeric' })
}

// موعد الحجز الفعلي (الموعد المحدد للزبون) هو المعيار الصح للفلترة بالتاريخ —
// لو ما محدد موعد بعد (لسا ما تنسّق)، نرجع لتاريخ التسجيل حتى الحجز يبقى قابل
// للعثور عليه بدل ما يضيع من كل الفلاتر.
function relevantDate(b: Booking): string {
  return b.scheduledAt || b.createdAt
}

const techRoleLabels: Record<string, string> = {
  TECH_1: 'الفني الأول',
  TECH_2: 'الفني الثاني',
  TECH_3: 'الفني الثالث',
}

// ═══ سلال الحجوزات ═══
// كل تبويب بشاشة «الحجوزات» يفتح نفس القائمة بسلّة مختلفة.
export default function BookingsList({ bucket = 'all' }: { bucket?: BookingBucket } = {}) {
  const { employee, permissions } = useSession()
  const canSeeStats = employee?.role === 'ADMIN' || employee?.role === 'MONITOR' || permissions.includes('monitoring')
  const [bookings, setBookings] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [search, setSearch] = useState('')
  const [expandedId, setExpandedId] = useState<string | null>(null)
  // ═══ «بانتظار التثبيت» بلا فلتر تاريخ ═══
  //
  // «الحجوزات الجديدة ما يحتاج بيها فلاتر — أي حجز جديد يجي هنا».
  //
  // ⚠️ وهذا ما كان ترتيب بس: الشاشة تبدي بفلتر **يوم اليوم**، فالحجز
  // الي انسجّل أمس وما أحد حچى وية زبونه بعد **ما يطلع أصلاً** —
  // الإداري يشوف «لا توجد حجوزات» ويظن ماكو شغل، والحجز يقعد بلا
  // متابعة لحد ما ينسى. الطابور الي ينتظر تصرّف ما ينفلتر بالتاريخ:
  // كله لازم ينشاف، أقدمه أول.
  const [selectedDate, setSelectedDate] = useState<string | null>(bucket === 'pending' ? null : todayStr())
  const [selectedMonth, setSelectedMonth] = useState<string | null>(null)
  const [doneFilter, setDoneFilter] = useState<DoneFilter>('ALL')
  const dateInputRef = useRef<HTMLInputElement>(null)
  const monthInputRef = useRef<HTMLInputElement>(null)
  const isAdmin = employee?.role === 'ADMIN'
  // ═══ التعديل صار بالصلاحية مو بالدور ═══
  //
  // «هنا نحتاج زر تعديل لتفاصيل الحجز، مثلاً الوقت أو الكادر».
  //
  // الزر جان موجود — بس محجوب على `role === 'ADMIN'` وحده. فإداري
  // الكوادر الي عنده صلاحية التنسيق يشوف الحجز ولا يكدر يلمسه، مع
  // إن **السيرفر يسمحله** (`requireBookingEdit` و`requireBookingCoord`
  // الاثنين يقبلون بالصلاحية مو بالدور).
  //
  // ⚠️ القوائم هنا تطابق حراس السيرفر بالضبط: زر يطلع وينرفض بـ٤٠٣
  // أسوأ من زر ما يطلع.
  const owner = employee?.role === 'OWNER'
  const canEditDetails = isAdmin || owner
    || ['coordinator', 'crew_management', 'view_bookings', 'mission_tracking', 'sales_booking']
      .some((p) => permissions.includes(p))
  const canEditCrew = isAdmin || owner
    || ['coordinator', 'crew_management'].some((p) => permissions.includes(p))
  // طلب حذف حجز: الإداري والمراقب ومدير النظام، أو أي واحد ينمنح الصلاحية
  const canRequestDelete = isAdmin || employee?.role === 'OWNER' ||
    employee?.role === 'HR_COORDINATOR' || employee?.role === 'MONITOR' ||
    permissions.includes('booking_delete_request')

  const requestDelete = async (bookingId: string, code: string) => {
    const reason = prompt(`سبب طلب حذف الحجز ${code}؟ (تجريبي، ملغى، مكرر...)`)
    if (!reason || !reason.trim()) return
    try {
      await api.requestBookingDelete(bookingId, reason.trim())
      alert('انرفع طلب الحذف — المراقب أو مدير النظام راح يبت بيه')
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر رفع الطلب')
    }
  }
  const [technicians, setTechnicians] = useState<Employee[]>([])
  // ═══ وين الليدر؟ ═══
  // «هنا بالتعديل مال الحجز، كون أكدر أخلي الليدر — وين الليدر؟».
  // الليدر بالنظام حقل مستقل (`projectSupervisor` — «تيم ليدر») مو
  // دور من أدوار الفنيين الثلاثة. شاشة التنسيق عندها هالخانة من
  // زمان، وشاشة الحجوزات نسيتها — فالإداري الي يعدّل الكادر من هنا
  // يلگه ثلاث خانات فنيين وماكو وين يحط الليدر.
  const [supervisors, setSupervisors] = useState<Employee[]>([])
  const [vehicles, setVehicles] = useState<VehicleOption[]>([])
  const [mapBooking, setMapBooking] = useState<Booking | null>(null)
  // الشخصيات المهمة: أي موظف يقدر يعلّم زبون بضغطة زر. نجيب المعرّفات بس
  // (بدون تفاصيل) حتى نعرف أي زر يكون مضغوط.
  const [vipIds, setVipIds] = useState<string[]>([])
  const [vipBusy, setVipBusy] = useState<string | null>(null)
  const [editingId, setEditingId] = useState<string | null>(null)
  const [assigning, setAssigning] = useState(false)

  useEffect(() => {
    if (canEditCrew) {
      api.getEmployees().then((all) => setTechnicians(all.filter((e) => e.role === 'TECHNICIAN')))
      // المركبات بيانات مساعدة هنا — مو كل من يشوف الحجوزات عنده صلاحية
      // المركبات، فالرفض ما يجوز يكسر الصفحة (شوف Coordinator.tsx)
      api.getVehicleOptions().then(setVehicles).catch(() => setVehicles([]))
      // ⚠️ الرفض ما يكسر الشاشة: مو كل من يعدّل الكادر عنده صلاحية
      // قائمة المشرفين — وقتها تبقى الخانة فاضية بدل ما تطفّي الصفحة.
      api.getSupervisors().then(setSupervisors).catch(() => setSupervisors([]))
    }
  }, [canEditCrew])

  useEffect(() => { api.getVipCustomerIds().then(setVipIds).catch(() => {}) }, [])

  // تعليم/إزالة "شخصية مهمة" — التعليم متاح لأي موظف، أما إزالة التعليم
  // فمحصورة بمدير النظام (السيرفر يفرضها كمان).
  const toggleVip = async (booking: Booking) => {
    const customerId = booking.customer?.id
    if (!customerId) return
    const already = vipIds.includes(customerId)
    if (already && !isAdmin) {
      alert('إزالة التعليم متاحة لمدير النظام فقط')
      return
    }
    setVipBusy(customerId)
    try {
      if (already) {
        await api.unmarkVipCustomer(customerId)
        setVipIds((prev) => prev.filter((id) => id !== customerId))
      } else {
        await api.markVipCustomer({
          customerId,
          bookingId: booking.id,
          requestSummary: booking.service?.name || undefined,
        })
        setVipIds((prev) => [...prev, customerId])
      }
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تنفيذ العملية')
    } finally {
      setVipBusy(null)
    }
  }

  // ═══ مسار التثبيت — من «بانتظار التثبيت» للتنسيق ═══
  //
  // «الحجوزات الجديدة الي ما متفقين وية الزبون المفروض تطلع
  // بانتظار التثبيت. من نضغط (تواصل وية الزبون) و(ترحيل لكادر الشد)
  // يله يترحّل لتنسيق الحجوزات. أني ما أريد الحجوزات مباشرة تطلع
  // بتنسيق الحجوزات».
  //
  // ⚠️ قبل، الحجز الجديد كان يوصل شاشة التنسيق **بلحظة تسجيله**:
  // المنسّق يشوف حجوزات ما أحد حچى وية زبونها بعد، ويبدي يدوّر
  // مواعيد وكوادر لشغل يمكن ما يصير أصلاً.
  const [flowBusy, setFlowBusy] = useState<string | null>(null)

  const markContacted = async (booking: Booking) => {
    setFlowBusy(booking.id)
    try {
      const updated = await api.markConfirmationContacted(booking.id)
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تسجيل التواصل')
    } finally { setFlowBusy(null) }
  }

  /** يثبّت الحجز ويرحّله: لكادر الشد (يوصل التنسيق) أو لإدارة
   *  المشاريع (يبقى مقفل عندهم لحد ما يوصل مرحلة التنفيذ). */
  const confirmAndTransfer = async (booking: Booking, toProjects: boolean) => {
    const where = toProjects ? 'إدارة المشاريع' : 'كادر الشد'
    if (!confirm(`تثبيت الحجز ${booking.code} وترحيله لـ${where}؟`)) return
    setFlowBusy(booking.id)
    try {
      const updated = await api.confirmBooking(booking.id, {
        confirmedByName: employee?.name || '',
        confirmedByEmployeeId: employee?.id,
        transferToProjects: toProjects,
      })
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر التثبيت')
    } finally { setFlowBusy(null) }
  }

  const handleVehicleChange = async (booking: Booking, assignedVehicle: string) => {
    setAssigning(true)
    try {
      const updated = await api.updateBookingDetails(booking.id, { assignedVehicle })
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تعديل السيارة')
    } finally {
      setAssigning(false)
    }
  }

  /** تعيين/إزالة الليدر (تيم ليدر) — نفس مسار السيرفر الي تستعمله
   *  شاشة التنسيق، حتى ما يصير مسارين للشغلة الوحدة. */
  const handleSupervisor = async (booking: Booking, employeeId: string) => {
    setAssigning(true)
    try {
      const updated = await api.assignSupervisor(booking.id, employeeId || null)
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تعيين الليدر')
    } finally {
      setAssigning(false)
    }
  }

  const handleReassign = async (booking: Booking, role: 'TECH_1' | 'TECH_2' | 'TECH_3', employeeId: string) => {
    if (!employeeId) return
    setAssigning(true)
    try {
      const updated = await api.assignTechnician(booking.id, { employeeId, role })
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تعديل التكليف')
    } finally {
      setAssigning(false)
    }
  }

  useEffect(() => {
    // نطلب من السيرفر يوم واحد بس (لو محدد) بدل ما نجيب كل أرشيف الحجوزات التاريخي
    // ونفلتره بالواجهة — نفس المشكلة اللي صلحناها بصفحة تنسيق الحجوزات، وهذي الصفحة
    // كانت أسوأ لأنها ما تفلتر بالحالة إطلاقاً (كل الحجوزات من كل الأزمنة).
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setLoading(true)
    // فلتر الشهر يحتاج كل الحجوزات (نفلترها بالواجهة)، أما فلتر يوم واحد فنطلبه
    // من السيرفر مباشرة (أخف على قاعدة البيانات) — الاثنين ما ينفعون بنفس الوقت.
    api
      .getBookings(selectedMonth ? undefined : selectedDate ? { date: selectedDate } : undefined)
      .then(setBookings)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [selectedDate, selectedMonth])

  // ═══ شنو يطلع بهاي الشاشة ═══
  //
  // «الحجز الجديد بعدني ما مثبته ولا مكلف كادر ولا منسق ولا متواصل وية
  // الزبون — ما أريده يطلع هنا إلا بحال تم التواصل وتحديد الكادر
  // والتاريخ».
  //
  // الحجز الي لسه ما انسّق مو «حجز اليوم» — هو **طلب** واصل للتنسيق.
  // عرضه هنا يخلط الشغل المؤكد بالشغل الي لسه ما انولد، ويخلي عدّ
  // «شكد عدنا اليوم» كذب.
  //
  // ⚠️ ما ينختفي بصمت: العدد ينعرض فوق مع رابط لشاشة التنسيق. إخفاء
  // بلا إشارة يخلي الإداري يظن الحجز ضاع ويعيد تسجيله.
  // ⚠️ الإخفاء الصامت انلغى.
  //
  // كانت الشاشة تعرض **بس** الحجز الي عنده تثبيت وموعد وكادر — أي
  // حجز ناقصه وحدة منهن يختفي. يعني الإداري يثبّت حجز بلا كادر
  // (وهذا مشروع تماماً: يجوز يثبّته اليوم ويكلّف الكادر بعد أربع
  // أيام) فيروح يدوّر عليه بالقائمة وما يلگاه — فيظن إنه ضاع
  // ويعيد تسجيله.
  //
  // هسه السلّة هي الي تقرر شنو يطلع، والإداري يعرف بأي سلّة هو.
  //
  // ═══ والسلال ما تتداخل ═══
  //
  // «أريد الحجوزات الي ما مثبتة أصلاً، الي هنّ الحجوزات المرحّلة
  // حديثاً حتى نتواصل وية الزبون — هذا الي بانتظار التثبيت».
  //
  // فالحجز يمشي بسلّة وحدة بكل لحظة: بانتظار التثبيت ← مثبّتة ←
  // مكلّفة. لو طلع بسلّتين، الإداري يعالجه مرتين أو ينساه لأنه
  // «شافه بالسلّة الثانية».
  const inBucket = (b: (typeof bookings)[number]) => {
    const crewed = (b.assignments?.length ?? 0) > 0
    const started = executionStarted(b)
    switch (bucket) {
      // بانتظار التثبيت: المرحّل حديثاً الي ننتظر نتواصل وية زبونه.
      // ⚠️ الملغى والمنجز ما يبقون هنا — هذي سلّة شغل، مو أرشيف.
      case 'pending':
        return !b.confirmedAt && !crewed && !started
          && b.status !== 'CANCELLED' && b.status !== 'COMPLETED'
      // مثبّتة: انثبّت وبعده ما انكلّف عليه كادر ولا بدا التنفيذ.
      // الموعد مو شرط — يجوز يتثبّت اليوم وينحدد موعده بعد أيام.
      case 'confirmed':
        return !!b.confirmedAt && !crewed && !started
          && b.status !== 'COMPLETED' && b.status !== 'CANCELLED'
      // مكلّفة: انكلّف عليه كادر **أو** بدا التنفيذ فعلاً.
      // «أو» مو «و»: حجز باشر بيه الليدر بلا ما ينسجّل تكليف رسمي
      // لازم يبقى مرئي بمكان ما — وهذا مكانه.
      case 'assigned':
        // ⚠️ المنجز طلع من «مكلّفة»: صارله سلّته الخاصة بتفرّعاتها،
        // وبقاؤه بالاثنين يعني الإداري يشوفه بمكانين ويظن إنه لسه شغّال.
        return (crewed || started) && b.status !== 'CANCELLED'
          && b.status !== 'COMPLETED' && b.status !== 'PARTIAL'
      // ═══ تم الإنجاز ═══
      // «الاتجاه الأول الي هو تم الإنجاز، وراها بنود نتفرّع».
      // المنجز جزئياً يدخل هنا كمان: هو شغل صار فعلاً، وطلعته
      // انحسبت لكادرها — بس الحجز ما خلص بعد.
      case 'done': {
        const isDone = b.status === 'COMPLETED' || b.status === 'PARTIAL'
        if (!isDone) return false
        switch (doneFilter) {
          case 'PARTIAL': return b.status === 'PARTIAL'
          case 'ALL': return true
          default: return b.status === 'COMPLETED' && b.completionState === doneFilter
        }
      }
      default: return true
    }
  }

  const filtered = bookings
    .filter(inBucket)
    .filter((b) => {
      if (search.trim() && !matches([b.code, b.customer?.name, b.customer?.code, b.customer?.phone], search)) return false
      if (selectedMonth && !relevantDate(b).startsWith(selectedMonth)) return false
      return true
    })
    // الأحدث للأقدم دايماً — هذا الترتيب الافتراضي المطلوب.
    // ⚠️ إلا طابور الانتظار: هناك **الأقدم أول**، لأنه الي منتظر
    // أكثر — الزبون الي مسجّل من ثلاث أيام وما أحد حچى وياه أولى
    // من الي انسجّل قبل ساعة.
    .sort((a, b) => bucket === 'pending'
      ? new Date(relevantDate(a)).getTime() - new Date(relevantDate(b)).getTime()
      : new Date(relevantDate(b)).getTime() - new Date(relevantDate(a)).getTime())

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">{BUCKET_HEADINGS[bucket].title}</h2>
      <p className="mt-1 text-slate-500">
        {BUCKET_HEADINGS[bucket].next}
        {bucket === 'all' && canSeeStats && (
          <> إحصائية "أكثر الخدمات طلباً" لكل الخدمات صارت بصفحة <a href="/stats" className="text-brand-600 hover:underline">إحصائيات الموظفين</a>.</>
        )}
      </p>


      {/* ═══ تفرّعات «تم الإنجاز» ═══
          خيار واحد بالأعلى («تم الإنجاز») والتفصيل جوّاه — نفس نمط
          «زبون ما رد» و«حجوزات ملغية». */}
      {bucket === 'done' && (
        <div className="mt-4 inline-flex flex-wrap gap-1 rounded-xl bg-slate-100 p-1">
          {DONE_FILTERS.map((f) => (
            <button
              key={f.key}
              onClick={() => setDoneFilter(f.key)}
              className={`rounded-lg px-3 py-1.5 text-[11px] font-bold transition-colors ${
                doneFilter === f.key ? 'bg-white text-[#0f2040] shadow-sm' : 'text-slate-500 hover:text-slate-700'
              }`}
            >
              {f.label}
            </button>
          ))}
        </div>
      )}

      <div className="mt-6 flex flex-wrap items-center gap-3">
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="بحث برقم الحجز، اسم الزبون، كود الزبون، أو رقم الهاتف..."
          className="w-full rounded-lg border border-slate-300 px-3 py-2 outline-none focus:border-brand-500 sm:w-96"
        />

        {/* أدوات التاريخ ما تطلع بطابور الانتظار — شوف التعليق فوق */}
        {bucket !== 'pending' && (
        <div className="flex items-center gap-1 rounded-lg border border-slate-300 bg-white px-1 py-1">
          <button
            onClick={() => { setSelectedMonth(null); setSelectedDate((d) => addDays(d || todayStr(), 1)) }}
            className="rounded-md px-2 py-1 text-slate-500 hover:bg-slate-100"
            title="اليوم التالي"
          >
            ▶
          </button>
          <div className="relative">
            <button
              onClick={() => dateInputRef.current?.showPicker?.()}
              className="min-w-[180px] rounded-md px-2 py-1 text-sm font-medium text-brand-800 hover:bg-slate-100"
            >
              📅 {selectedMonth ? 'فلتر شهر مفعّل' : selectedDate ? formatDateArabic(selectedDate) : 'كل الحجوزات'}
            </button>
            <input
              ref={dateInputRef}
              type="date"
              value={selectedDate || ''}
              onChange={(e) => { if (e.target.value) { setSelectedMonth(null); setSelectedDate(e.target.value) } }}
              className="absolute inset-0 h-full w-full cursor-pointer opacity-0"
            />
          </div>
          <button
            onClick={() => { setSelectedMonth(null); setSelectedDate((d) => addDays(d || todayStr(), -1)) }}
            className="rounded-md px-2 py-1 text-slate-500 hover:bg-slate-100"
            title="اليوم السابق"
          >
            ◀
          </button>
        </div>
        )}

        {bucket !== 'pending' && selectedDate !== todayStr() && !selectedMonth && (
          <button
            onClick={() => setSelectedDate(todayStr())}
            className="rounded-lg border border-brand-200 px-3 py-2 text-sm font-medium text-brand-700 hover:bg-brand-50"
          >
            اليوم
          </button>
        )}

        {bucket !== 'pending' && (
        <button
          onClick={() => { setSelectedMonth(null); setSelectedDate((d) => (d === null ? todayStr() : null)) }}
          className={`rounded-lg border px-3 py-2 text-sm font-medium ${
            selectedDate === null && !selectedMonth
              ? 'border-brand-500 bg-brand-500 text-white'
              : 'border-slate-300 text-slate-600 hover:bg-slate-50'
          }`}
          title="كل الحجوزات (بدون فلتر تاريخ)، مرتبة من الأحدث للأقدم"
        >
          📋 كل الحجوزات
        </button>
        )}

        {/* ═══ فلتر الشهر ═══
            ⚠️ كان زر وفوقه خانة شهر **شفافة تغطيه بالكامل**: الضغطة
            تروح للخانة المخفية مو للزر، و`showPicker()` ما تنستدعى
            أبداً — فالفلتر ما يفتح ولا مرة. صارت خانة ظاهرة عادية،
            تشتغل بكل متصفح بلا حيلة. */}
        {bucket !== 'pending' && (
        <label className={`flex items-center gap-2 rounded-lg border px-3 py-1.5 text-sm font-medium ${
          selectedMonth ? 'border-brand-500 bg-brand-50 text-brand-800' : 'border-slate-300 text-slate-600'
        }`}>
          <span className="whitespace-nowrap">🗓️ فلتر شهر</span>
          <input
            ref={monthInputRef}
            type="month"
            value={selectedMonth || ''}
            onChange={(e) => {
              if (e.target.value) { setSelectedDate(null); setSelectedMonth(e.target.value) }
              else { setSelectedMonth(null); setSelectedDate(todayStr()) }
            }}
            className="rounded-md border border-slate-300 bg-white px-2 py-1 text-xs outline-none focus:border-brand-500"
          />
        </label>
        )}
        {bucket !== 'pending' && selectedMonth && (
          <button
            onClick={() => { setSelectedMonth(null); setSelectedDate(todayStr()) }}
            className="rounded-lg border border-slate-300 px-3 py-2 text-sm font-medium text-slate-600 hover:bg-slate-50"
          >
            إلغاء فلتر الشهر
          </button>
        )}
      </div>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
          تعذر الاتصال بالخادم: {error}
        </p>
      )}

      {!loading && !error && (
        <div className="mt-6 overflow-hidden rounded-xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="overflow-x-auto">
          <table className="w-full text-right">
            <thead className="bg-gradient-to-l from-brand-500 to-brand-800 text-white">
              <tr>
                <th className="px-4 py-3 text-sm font-semibold">رقم الحجز</th>
                <th className="px-4 py-3 text-sm font-semibold">الزبون</th>
                <th className="px-4 py-3 text-sm font-semibold">كود الزبون</th>
                <th className="px-4 py-3 text-sm font-semibold">الخدمة</th>
                <th className="px-4 py-3 text-sm font-semibold">الموظف الذي سجل</th>
                <th className="px-4 py-3 text-sm font-semibold">السيارة</th>
                <th className="px-4 py-3 text-sm font-semibold">موعد التنفيذ</th>
                <th className="px-4 py-3 text-sm font-semibold">الحالة</th>
                <th className="px-4 py-3 text-sm font-semibold"></th>
              </tr>
            </thead>
            <tbody className="divide-y divide-slate-100">
              {filtered.map((b) => (
                <Fragment key={b.id}>
                  <tr>
                    <td className="px-4 py-3 font-mono text-sm font-semibold text-brand-600">
                      {b.code}
                    </td>
                    <td className="px-4 py-3">{b.customer?.name || 'زبون غير معروف'}</td>
                    <td className="px-4 py-3 font-mono text-sm text-slate-500">{formatCustomerCode(b.customer) || '-'}</td>
                    <td className="px-4 py-3 text-slate-600">{serviceNames(b)}</td>
                    <td className="px-4 py-3 text-slate-600">{b.transferEmployee?.name || '-'}</td>
                    <td className="px-4 py-3 text-slate-500">{b.assignedVehicle || '-'}</td>
                    <td className="px-4 py-3 text-slate-500">
                      {b.scheduledAt
                        ? formatScheduleWindow(b.scheduledAt, b.scheduledEndAt)
                        : <span className="text-amber-600">لم يُنسَّق بعد ({new Date(b.createdAt).toLocaleDateString('ar-IQ')})</span>}
                    </td>
                    <td className="px-4 py-3">
                      {/* الحالة التفصيلية بدل «مثبت/منجز» الخام: تبيّن هل
                          انكلّف كادر عليه، وهل خلّص فاتورته وتقريره. نفس
                          اللغة الي بتنسيق الحجوزات بالضبط. */}
                      <CompletionBadge booking={b} />
                    </td>
                    <td className="px-4 py-3">
                      {/* مسار التثبيت — يطلع بسلّة «بانتظار التثبيت» بس،
                          لأنها الوحيدة الي هذي الخطوات تعني بيها شي. */}
                      {bucket === 'pending' && canEditDetails && (
                        <div className="mb-1.5 flex flex-wrap gap-1">
                          {!b.confirmationContactedAt ? (
                            <button
                              onClick={() => markContacted(b)}
                              disabled={flowBusy === b.id}
                              className="rounded-lg bg-amber-500 px-2.5 py-1 text-[11px] font-bold text-white hover:bg-amber-600 disabled:opacity-50"
                            >
                              📞 تواصلت وية الزبون
                            </button>
                          ) : (
                            <>
                              <span className="rounded-lg bg-emerald-50 px-2 py-1 text-[10px] font-bold text-emerald-700">
                                ✓ تواصلنا
                              </span>
                              <button
                                onClick={() => confirmAndTransfer(b, false)}
                                disabled={flowBusy === b.id}
                                className="rounded-lg bg-brand-600 px-2.5 py-1 text-[11px] font-bold text-white hover:bg-brand-700 disabled:opacity-50"
                              >
                                ✅ ثبّت ورحّل لكادر الشد
                              </button>
                              <button
                                onClick={() => confirmAndTransfer(b, true)}
                                disabled={flowBusy === b.id}
                                className="rounded-lg bg-violet-600 px-2.5 py-1 text-[11px] font-bold text-white hover:bg-violet-700 disabled:opacity-50"
                              >
                                🏗 ثبّت ورحّل لإدارة المشاريع
                              </button>
                            </>
                          )}
                        </div>
                      )}
                      <button
                        onClick={() => setExpandedId(expandedId === b.id ? null : b.id)}
                        className="rounded-lg border border-brand-200 px-3 py-1 text-xs font-medium text-brand-700 transition-colors hover:bg-brand-50"
                      >
                        {expandedId === b.id ? 'إخفاء' : 'التفاصيل'}
                      </button>
                    </td>
                  </tr>
                  {expandedId === b.id && (
                    <tr>
                      <td colSpan={9} className="bg-slate-50 px-4 py-4">
                        {/* ═══ وين وصل الحجز ═══
                            «أريد متابعة لحالة الحجز مرحلة بعد مرحلة».
                            ⚠️ المراحل تگول «وين وصل» مو «شنو ناقص عليك»:
                            كل مرحلة اختيارية لحالها، والحجز يتثبّت بلا
                            كادر والكادر ينكلّف بعدها بأيام. */}
                        <StageStrip booking={b} />

                        {/* ═══ الطلعات ═══
                            الحجز الي ياخذ أربع أيام أربع طلعات، وكل
                            وحدة إلها كادرها وتاريخها — والطلعة الي
                            راحت ما تنمحي لمن يتبدّل الكادر. */}
                        <div className="mt-3">
                          <BookingVisits bookingId={b.id} />
                        </div>

                        <div className="mt-4 grid grid-cols-1 gap-3 text-sm sm:grid-cols-2 lg:grid-cols-4">
                          <div>
                            <p className="text-slate-400">اسم الزبون</p>
                            <p className="mt-1 font-bold text-slate-700">{b.customer?.name || 'زبون غير معروف'}</p>
                            {b.customer?.id && (
                              <button
                                onClick={() => toggleVip(b)}
                                disabled={vipBusy === b.customer?.id}
                                className={`mt-1 rounded-full px-3 py-1 text-xs font-bold transition-colors disabled:opacity-50 ${
                                  vipIds.includes(b.customer.id)
                                    ? 'bg-amber-400 text-amber-950'
                                    : 'bg-slate-100 text-slate-500 hover:bg-amber-100 hover:text-amber-700'
                                }`}
                              >
                                {vipIds.includes(b.customer.id) ? '⭐ شخصية مهمة' : '☆ تعليم كشخصية مهمة'}
                              </button>
                            )}
                            {canRequestDelete && (
                              <button
                                onClick={() => requestDelete(b.id, b.code)}
                                className="mt-1 mr-2 rounded-full bg-red-50 px-3 py-1 text-xs font-bold text-red-700 transition-colors hover:bg-red-100"
                              >
                                🗑️ اطلب حذف الحجز
                              </button>
                            )}
                          </div>
                          <div>
                            <p className="text-slate-400">رقم هاتف الزبون</p>
                            <p className="mt-1 font-bold text-brand-700" dir="ltr">{b.customer?.phone || '-'}</p>
                          </div>
                          {/* تعديل الحجز نفسه: الخدمة والسعر والموعد.
                              قبل، التعديل الوحيد هنا جان تغيير الكادر —
                              وأي تغيير بطلب الزبون يحتاج إلغاء الحجز
                              وإعادة إنشائه. */}
                          {canEditDetails && (
                            <div className="sm:col-span-2">
                              <BookingEditPanel
                                booking={b}
                                onSaved={(u) => setBookings((prev) => prev.map((x) => (x.id === u.id ? u : x)))}
                              />
                            </div>
                          )}
                          <div>
                            <div className="flex items-center justify-between">
                              <p className="text-slate-400">الكادر الذي تم تكليفه</p>
                              {canEditCrew && (
                                <button
                                  type="button"
                                  onClick={() => setEditingId(editingId === b.id ? null : b.id)}
                                  className="rounded-md bg-brand-50 px-2 py-0.5 text-xs font-bold text-brand-700 hover:bg-brand-100"
                                >
                                  {editingId === b.id ? 'إغلاق التعديل' : 'تعديل'}
                                </button>
                              )}
                            </div>
                            {editingId === b.id ? (
                              <div className="mt-2 flex flex-col gap-2">
                                {/* ⚠️ الليدر أول الخانات: هو المسؤول عن
                                    الحجز، والفنيين وراه. */}
                                <div>
                                  <label className="mb-0.5 block text-xs font-bold text-amber-700">👑 تيم ليدر</label>
                                  <select
                                    value={b.projectSupervisor?.id || ''}
                                    disabled={assigning}
                                    onChange={(e) => handleSupervisor(b, e.target.value)}
                                    className="w-full rounded-lg border border-amber-300 px-2 py-1.5 text-sm outline-none focus:border-amber-500"
                                  >
                                    <option value="">-- بدون تيم ليدر --</option>
                                    {supervisors.map((sv) => (
                                      <option key={sv.id} value={sv.id}>{sv.name}</option>
                                    ))}
                                  </select>
                                </div>
                                {techRoles.map((tr) => {
                                  const current = b.assignments.find((a) => a.role === tr.key)
                                  return (
                                    <div key={tr.key}>
                                      <label className="mb-0.5 block text-xs text-slate-400">{tr.label}</label>
                                      <select
                                        value={current?.employee.id || ''}
                                        disabled={assigning}
                                        onChange={(e) => handleReassign(b, tr.key, e.target.value)}
                                        className="w-full rounded-lg border border-slate-300 px-2 py-1.5 text-sm outline-none focus:border-brand-500"
                                      >
                                        <option value="">-- اختر --</option>
                                        {technicians.map((t) => (
                                          <option key={t.id} value={t.id}>{t.name}{t.isLeader ? ' (ليدر)' : ''}</option>
                                        ))}
                                      </select>
                                    </div>
                                  )
                                })}
                              </div>
                            ) : b.assignments.length > 0 || b.projectSupervisor ? (
                              <ul className="mt-1 list-inside list-disc text-slate-700">
                                {/* الليدر ينعرض حتى لو ماكو ولا فني مكلّف:
                                    هو المسؤول، وغيابه من العرض يخلّي
                                    الحجز يبان بلا مسؤول وهو إله. */}
                                {b.projectSupervisor && (
                                  <li className="font-bold text-amber-800">
                                    👑 {b.projectSupervisor.name}
                                    <span className="font-normal text-slate-400"> (تيم ليدر)</span>
                                  </li>
                                )}
                                {/* ═══ وين الليدر؟ ═══
                                    «أني محدد ليدر، جاي يطلعلي فقط فني — وين
                                    الليدر؟». الليدر مو دور تكليف منفصل: هو
                                    موظف عليه علم `isLeader` وينكلّف كأي فني.
                                    فكانت القائمة تگول «الفني الأول» بس،
                                    والإداري يشوف حجزه بلا ليدر مع إنه حدده.
                                    ⚠️ العلم يجي من السيرفر مع بيانات الموظف —
                                    ما ننسخه ولا نستنتجه بالمتصفح. */}
                                {b.assignments.map((a) => (
                                  <li key={a.id}>
                                    {a.employee.name}
                                    <span className="text-slate-400"> ({techRoleLabels[a.role] || a.role})</span>
                                    {a.employee.isLeader && (
                                      <span className="mr-1.5 rounded-full bg-amber-100 px-2 py-0.5 text-[10px] font-black text-amber-800">
                                        👑 ليدر
                                      </span>
                                    )}
                                  </li>
                                ))}
                              </ul>
                            ) : (
                              <p className="mt-1 text-slate-400">لم يتم تكليف أحد بعد</p>
                            )}
                          </div>
                          <div>
                            <p className="text-slate-400">السيارة المخصصة</p>
                            {editingId === b.id ? (
                              <select
                                value={b.assignedVehicle || ''}
                                disabled={assigning}
                                onChange={(e) => handleVehicleChange(b, e.target.value)}
                                className="mt-1 w-full rounded-lg border border-slate-300 px-2 py-1.5 text-sm outline-none focus:border-brand-500"
                              >
                                <option value="">-- اختر سيارة --</option>
                                {vehicles.map((v) => (
                                  <option key={v.id} value={`${v.name} - ${v.plateNumber}`}>
                                    {v.name} ({v.plateNumber}){v.color ? ` - ${v.color}` : ''}
                                  </option>
                                ))}
                              </select>
                            ) : (
                              <p className="mt-1 text-slate-700">{b.assignedVehicle || '-'}</p>
                            )}
                          </div>
                          <div>
                            <p className="text-slate-400">عنوان تنفيذ المهمة</p>
                            <p className="mt-1 text-slate-700">{b.address || '-'}</p>
                            {b.mapLatitude != null && b.mapLongitude != null ? (
                              <button
                                onClick={() => setMapBooking(b)}
                                className="text-xs text-brand-500 hover:underline"
                              >
                                عرض الموقع على الخريطة 🗺️
                              </button>
                            ) : b.mapLocation && (
                              <a href={b.mapLocation} target="_blank" rel="noreferrer" className="text-xs text-brand-500 hover:underline">
                                فتح على الخريطة
                              </a>
                            )}
                          </div>
                          <div>
                            <p className="text-slate-400">التكلفة المقدرة</p>
                            <p className="mt-1 text-slate-700">
                              {b.quotedPrice != null ? b.quotedPrice.toLocaleString() : '-'}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">الدفعة المقدمة</p>
                            <p className="mt-1 text-slate-700">
                              {b.advancePaid != null ? b.advancePaid.toLocaleString() : '-'}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">وقت تسجيل الحجز</p>
                            <p className="mt-1 text-slate-700">
                              {new Date(b.createdAt).toLocaleString('ar-IQ')}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">الموعد المحدد للزبون</p>
                            <p className="mt-1 text-slate-700">
                              {b.scheduledAt ? formatScheduleWindow(b.scheduledAt, b.scheduledEndAt) : 'غير محدد'}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">وقت إنجاز الحجز</p>
                            <p className="mt-1 text-slate-700">
                              {b.completedAt ? new Date(b.completedAt).toLocaleString('ar-IQ') : 'لم يُنجز بعد'}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">المبلغ المستلم</p>
                            <p className="mt-1 text-slate-700">
                              {b.amountCollected != null ? b.amountCollected.toLocaleString() : '-'}
                              {b.amountCollected != null && (
                                <span className="mr-2 text-xs text-slate-400">
                                  ({b.amountVerified ? 'مدقق' : 'بانتظار التدقيق'})
                                </span>
                              )}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">ملاحظات الإنجاز</p>
                            <p className="mt-1 text-slate-700">{b.completionNotes || '-'}</p>
                          </div>
                          <div>
                            <p className="text-slate-400">ملاحظات الحجز</p>
                            <p className="mt-1 text-slate-700">{b.notes || '-'}</p>
                          </div>
                          <div>
                            <p className="text-slate-400">من أكد الحجز</p>
                            <p className="mt-1 text-slate-700">{b.confirmedByEmployee?.name || b.confirmedByName || '-'}</p>
                          </div>
                          <div>
                            <p className="text-slate-400">من عدّل الحجز</p>
                            <p className="mt-1 text-slate-700">
                              {b.lastEditedBy?.name || '-'}
                              {b.lastEditedBy && b.lastEditedAt && (
                                <span className="text-xs text-slate-400"> ({new Date(b.lastEditedAt).toLocaleString('ar-IQ')})</span>
                              )}
                            </p>
                          </div>
                          <div>
                            <p className="text-slate-400">موظف المبيعات (مصدر الزبون)</p>
                            <p className="mt-1 text-slate-700">{b.transferEmployee?.name || '-'}</p>
                          </div>
                          {b.adminNotes && (
                            <div className="col-span-full">
                              <p className="text-slate-400">ملاحظات الإدارة</p>
                              <p className="mt-1 whitespace-pre-line rounded-lg bg-amber-50 border border-amber-200 p-2 text-sm text-amber-800">{b.adminNotes}</p>
                            </div>
                          )}
                        </div>
                      </td>
                    </tr>
                  )}
                </Fragment>
              ))}
              {filtered.length === 0 && (
                <tr>
                  <td colSpan={9} className="px-4 py-6 text-center text-slate-400">
                    لا توجد حجوزات
                  </td>
                </tr>
              )}
            </tbody>
          </table>
          </div>
        </div>
      )}

      {mapBooking && mapBooking.mapLatitude != null && mapBooking.mapLongitude != null && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={() => setMapBooking(null)}>
          <div className="w-full max-w-2xl rounded-2xl bg-white p-4 shadow-xl" onClick={(e) => e.stopPropagation()}>
            <div className="mb-3 flex items-center justify-between">
              <h3 className="font-bold text-brand-900">موقع الحجز {mapBooking.code}</h3>
              <button onClick={() => setMapBooking(null)} className="rounded-lg px-3 py-1 text-sm text-slate-500 hover:bg-slate-100">✕ إغلاق</button>
            </div>
            <MapViewer lat={mapBooking.mapLatitude} lng={mapBooking.mapLongitude} height={380} />
          </div>
        </div>
      )}
    </div>
  )
}

/* ───── شريط مراحل الحجز ───── */

function StageStrip({ booking }: { booking: Booking }) {
  const now = currentStage(booking)
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-3">
      <p className="mb-2.5 text-[11px] font-extrabold text-[#0f2040]">
        📍 وين وصل الحجز — <span className="text-brand-700">{now.done(booking) ? 'انجز' : now.label}</span>
      </p>
      {/* ⚠️ تمرير أفقي جوّا الشريط: سبع مراحل على شاشة موبايل تنضغط
          لنقاط بلا أسماء، فتصير زينة ما تفيد. */}
      <div className="flex gap-1 overflow-x-auto pb-1">
        {BOOKING_STAGES.map((st, i) => {
          const done = st.done(booking)
          const isNow = st.key === now.key && !done
          const at = st.at(booking)
          return (
            <div key={st.key} className="flex min-w-[92px] flex-1 flex-col items-center gap-1">
              <div className="flex w-full items-center">
                <span className={`h-1 flex-1 rounded-full ${i === 0 ? 'bg-transparent' : done ? 'bg-emerald-400' : 'bg-slate-200'}`} />
                <span className={`flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-[11px] ${
                  done ? 'bg-emerald-500 text-white'
                    : isNow ? 'bg-amber-400 text-white ring-2 ring-amber-200'
                    : 'bg-slate-200 text-slate-400'
                }`}>
                  {done ? '✓' : st.icon}
                </span>
                <span className={`h-1 flex-1 rounded-full ${i === BOOKING_STAGES.length - 1 ? 'bg-transparent' : done ? 'bg-emerald-400' : 'bg-slate-200'}`} />
              </div>
              <span className={`text-center text-[9.5px] leading-tight ${
                done ? 'font-bold text-emerald-700' : isNow ? 'font-bold text-amber-700' : 'text-slate-400'
              }`}>
                {st.label}
              </span>
              {/* الوقت ينعرض للمرحلة الي خلصت بس — المرحلة الي ما
                  صارت ما إلها وقت، وشرطة مكانها ضجيج. */}
              {done && at && (
                <span className="text-[9px] text-slate-400">
                  {new Date(at).toLocaleDateString('ar-IQ', { day: 'numeric', month: 'short' })}
                </span>
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
