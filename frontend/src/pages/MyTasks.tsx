import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { formatScheduleWindow } from '../utils/schedule'
import CompletionBadge from '../components/CompletionBadge'
import { api, type Booking, type PersonalTool } from '../api'
import PartialCompleteDialog from '../components/PartialCompleteDialog'
import BookingProgressTimeline from '../components/BookingProgressTimeline'
import EntityIdentity from '../components/EntityIdentity'
import MyExtraTasks from '../components/MyExtraTasks'
import LeaderInvoicesListPage from './LeaderInvoicesListPage'
import WorkReportPage from './WorkReportPage'
import { useSession } from '../session'
import { useSaveGuard } from '../useSaveGuard'
import SaveError from '../components/SaveError'

// ═══ الخيارات الثلاثة ═══
// الحجوزات = شغلك اليوم · فواتيري = فواتير شغلك · تقاريري = تقارير شغلك.
// ⚠️ برّا المكوّن: مصفوفة تنبني بكل رندر تخلي React يعيد بناء الأزرار
// بلا داعي.
const TABS = [
  { key: 'bookings' as const, label: 'الحجوزات', icon: '📋' },
  { key: 'invoices' as const, label: 'فواتيري', icon: '🧾' },
  { key: 'reports' as const, label: 'تقاريري', icon: '📝' },
]

type TabKey = (typeof TABS)[number]['key']

function elapsedSince(iso: string): string {
  const diffMin = Math.max(0, Math.floor((Date.now() - new Date(iso).getTime()) / 60000))
  const h = Math.floor(diffMin / 60)
  const m = diffMin % 60
  return h > 0 ? `${h} ساعة و ${m} دقيقة` : `${m} دقيقة`
}

// خريطة الطريق تفتح داخل بوب-أب فوق نفس الصفحة (بدون تحويل الفني لصفحة ثانية
// جوه النظام) — وزر "فتح بتطبيق الخرائط" يفتح تطبيق خرائط خارجي للتنقل الفعلي.
function DirectionsModal({ booking, onClose }: { booking: Booking; onClose: () => void }) {
  const mapRef = useRef<HTMLDivElement>(null)
  const lat = booking.mapLatitude
  const lng = booking.mapLongitude

  useEffect(() => {
    if (!mapRef.current || lat == null || lng == null) return
    const map = L.map(mapRef.current).setView([lat, lng], 15)
    L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; OpenStreetMap',
    }).addTo(map)
    L.marker([lat, lng]).addTo(map)
    return () => { map.remove() }
  }, [lat, lng])

  return (
    <div className="fixed inset-0 z-[70] flex items-center justify-center bg-black/50 p-4">
      <div className="w-full max-w-lg rounded-2xl bg-white shadow-2xl">
        <div className="flex items-center justify-between border-b border-slate-100 p-4">
          <h3 className="font-bold text-brand-900">🗺️ الطريق لموقع الزبون</h3>
          <button onClick={onClose} className="text-2xl text-slate-400 hover:text-slate-600">×</button>
        </div>
        <div className="p-4">
          <p className="mb-3 text-sm text-slate-500">{booking.address || booking.customer?.location || 'بدون عنوان محدد'}</p>
          {lat != null && lng != null ? (
            <div ref={mapRef} className="h-64 w-full rounded-xl border" />
          ) : (
            <p className="rounded-lg bg-amber-50 p-3 text-sm text-amber-700">ما اكو إحداثيات محددة لهذا الموقع.</p>
          )}
        </div>
        <div className="flex gap-3 border-t border-slate-100 p-4">
          {lat != null && lng != null && (
            <a
              href={`https://www.google.com/maps/dir/?api=1&destination=${lat},${lng}`}
              target="_blank" rel="noreferrer"
              className="flex-1 rounded-lg bg-brand-500 py-2.5 text-center text-sm font-bold text-white hover:bg-brand-600"
            >
              فتح بتطبيق الخرائط للتنقل ←
            </a>
          )}
          <button onClick={onClose} className="rounded-lg border border-slate-200 px-6 py-2.5 text-sm font-bold text-slate-600 hover:bg-slate-50">إغلاق</button>
        </div>
      </div>
    </div>
  )
}

export default function MyTasks() {
  // كل حفظ بهاي الشاشة يمر من هنا — الفشل ينعرض بدل ما ينبلع
  const guard = useSaveGuard()
  const { employee, permissions } = useSession()
  const navigate = useNavigate()
  const [bookings, setBookings] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [amounts, setAmounts] = useState<Record<string, string>>({})
  const [advances, setAdvances] = useState<Record<string, string>>({})
  const [notes, setNotes] = useState<Record<string, string>>({})
  const [, setTick] = useState(0)

  // مودال شيك الأدوات الشخصية قبل "استلام" الحجز — كل الأدوات معلّمة تلقائياً
  // (مؤشرة) كموجودة، والموظف يشيل التأشير فقط عن الناقص عنده (أسرع من ما يعلّم
  // كل أداة لحالها).
  const [toolsModalBooking, setToolsModalBooking] = useState<Booking | null>(null)
  const [directionsFor, setDirectionsFor] = useState<Booking | null>(null)
  const [personalTools, setPersonalTools] = useState<PersonalTool[]>([])
  const [checkedTools, setCheckedTools] = useState<Record<string, boolean>>({})
  const [toolsLoading, setToolsLoading] = useState(false)
  const [submittingAccept, setSubmittingAccept] = useState(false)

  // ── الخيارات الثلاثة ──
  // الصلاحية هي الي تقرر منو يشوف شنو، مو الدور: محاسب انطيته صلاحية
  // فواتير يشوف تبويب فواتيره، وفني بلا الصلاحية ما يشوفه.
  const [tab, setTab] = useState<TabKey>('bookings')
  const perms = permissions ?? []
  const canSeeInvoices = employee?.role === 'ADMIN' || perms.includes('leader_invoices_view') || perms.includes('execution_cost')
  const canSeeReports = employee?.role === 'ADMIN' || perms.includes('work_reports')

  useEffect(() => {
    const t = setInterval(() => setTick((v) => v + 1), 30000)
    return () => clearInterval(t)
  }, [])

  const load = () => {
    // المنجزة تنجلب بعد — الموظف لازم يضل يشوف شغله الي خلّصه
    // بتفاصيله، مو يختفي عنه أول ما يضغط «تم».
    // ⚠️ نطلب «مهامي» من السيرفر مو نجيب حجوزات الشركة ونفلترها هنا.
    //
    // الفلترة بالحالة كانت **تخبّي الحجز عن الليدر**: الإداري يحط
    // الليدر بخانة الليدر قبل ما يثبّت الحجز، فالحجز يضل PENDING —
    // وهذي الشاشة كانت تطلب CONFIRMED وIN_PROGRESS وCOMPLETED بس.
    // النتيجة: الليدر ما يشوف ولا شي، والحل الي كانوا يسوونه إنهم
    // يحطونه بخانة الفني هم حتى يطلعله.
    //
    // ومسار السيرفر يغطي الاثنين: المكلّف بجدول التعيينات **والليدر**
    // بعمود المشرف — وهو الي كان ناقص.
    //
    // وفايدة ثانية: الفني ما عاد ينزّل حجوزات الشركة كلها بزبائنها
    // على تلفونه حتى يشوف مهامه هو.
    api.getBookings({ assignedTo: 'me' })
      .then(setBookings)
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  // الحجز يخصني إذا انكلّفت بيه ككادر، **أو** إذا أني التيم ليدر
  // مالته. الليدر ينحفظ بعمود مستقل مو بجدول التعيينات — بدون هذا
  // الشرط الليدر ما يشوف ولا حجز بشاشته وهو المسؤول عنه.
  // ═══ منو يسوق مسار الحجز؟ ═══
  // الليدر. هو الي يجهّز المواد، ويأشّر الانطلاق، ويبدي العمل،
  // وينهيه. الفني ما يحتاج يضغط ولا وحدة منهن — أول ما يضغط الليدر
  // «انطلقنا» يصير الفريق كله منطلق تلقائياً، لأن الحدث على الحجز
  // مو على كل موظف لحاله.
  //
  // شغل الفني الوحيد: يجرد أدواته وأدوات السيارة وينطي «تم».
  const amLeaderOf = (b: Booking) => {
    if (!employee) return false
    if (b.projectSupervisor?.id) return b.projectSupervisor.id === employee.id
    // حجز ما انتحدد له تيم ليدر: أي ليدر مكلّف بيه يسوق المسار
    return !!employee.isLeader && b.assignments.some((a) => a.employee.id === employee.id)
  }

  const isMine = (b: Booking) =>
    b.assignments.some((a) => a.employee.id === employee?.id) ||
    b.projectSupervisor?.id === employee?.id
  const myTasks = bookings.filter((b) => isMine(b) && b.status !== 'COMPLETED')
  // شغلي الي خلّصته — يضل ظاهر بتفاصيله وبحالة ورقه
  const myDone = bookings
    .filter((b) => isMine(b) && b.status === 'COMPLETED')
    .sort((a, b) => new Date(b.completedAt || 0).getTime() - new Date(a.completedAt || 0).getTime())

  const handleArrive = async (booking: Booking) => {
    const updated = await guard.run('تسجيل الوصول', () => api.markArrived(booking.id))
    if (!updated) return
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  // نعتبرها الاستلام الفعلي (تنفيذ startBooking) — تُستدعى مباشرة لو الموظف
  // ما عنده أدوات شخصية مسجلة أصلاً (نتخطى المودال حتى لا يعلق بواجهة فاضية)،
  // أو بعد ما يضغط "تم" بمودال شيك الأدوات.
  const doStart = async (bookingId: string, missingToolIds?: string[]) => {
    const updated = await guard.run('بدء العمل', () => api.startBooking(bookingId, missingToolIds))
    if (!updated) return
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  // جرد الفني: يفتح نفس مودال الأدوات بس ما يبدي العمل — البدء بيد
  // الليدر. toolsOnly تفرّق بين الحالتين.
  const [toolsOnly, setToolsOnly] = useState(false)
  const openToolsCheck = async (booking: Booking, onlyTools: boolean) => {
    setToolsOnly(onlyTools)
    if (!employee) { if (!onlyTools) await doStart(booking.id); return }
    setToolsModalBooking(booking)
    setToolsLoading(true)
    try {
      const tools = await api.getPersonalTools(employee.id)
      if (tools.length === 0) {
        setToolsModalBooking(null)
        if (!onlyTools) await doStart(booking.id)
        return
      }
      setPersonalTools(tools)
      const allChecked: Record<string, boolean> = {}
      tools.forEach((t) => { allChecked[t.id] = true })
      setCheckedTools(allChecked)
    } catch {
      setToolsModalBooking(null)
      if (!onlyTools) await doStart(booking.id)
    } finally {
      setToolsLoading(false)
    }
  }

  const handleConfirmToolsCheck = async () => {
    if (!toolsModalBooking) return
    setSubmittingAccept(true)
    try {
      const missingToolIds = personalTools.filter((t) => !checkedTools[t.id]).map((t) => t.id)
      if (toolsOnly) {
        // جرد بس — ما نبدي العمل. البدء بيد الليدر.
        const missingNames = personalTools.filter((t) => !checkedTools[t.id]).map((t) => t.name)
        await api.createInventoryCheck({
          complete: missingNames.length === 0,
          missingItems: missingNames.length ? missingNames.join('، ') : undefined,
        })
      } else {
        await doStart(toolsModalBooking.id, missingToolIds)
      }
      setToolsModalBooking(null)
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تأكيد الاستلام')
    } finally {
      setSubmittingAccept(false)
    }
  }

  const handleMaterialsReady = async (booking: Booking) => {
    const updated = await guard.run('تأشير جاهزية المواد', () => api.setMaterialsReady(booking.id))
    if (!updated) return
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  // ═══ بعد الإنجاز: الورق ما ينتسى ═══
  // الإنجاز لحاله ما يكفي — باقي فاتورة التكاليف المربوطة بالحجز وتقرير
  // العمل. قبل، الحجز جان يختفي من الشاشة أول ما يضغط «تم الإنجاز»
  // فينسى الورق ويطلع الحجز «منجز» وهو ناقص. هسه تطلع له مطالبة بيها
  // الاثنين — بس مو إجبارية: يكدر يأجلها بـ«بعدين» ويرجع لها.
  const [paperwork, setPaperwork] = useState<{ booking: Booking; stopped: boolean } | null>(null)
  const [stopFor, setStopFor] = useState<Booking | null>(null)
  const [partialFor, setPartialFor] = useState<Booking | null>(null)
  const [stopReason, setStopReason] = useState('')
  const [stopping, setStopping] = useState(false)
  // ⚠️ الوقت ينثبت بفتح الشاشة مو بكل رندر — وفوق هنا لأن الخطّافات
  // لازم تنستدعى قبل أي return مبكر.
  const [{ todayKey, monthAgo }] = useState(() => ({
    todayKey: new Date().toDateString(),
    monthAgo: Date.now() - 30 * 24 * 60 * 60 * 1000,
  }))

  const handleComplete = async (booking: Booking) => {
    const amountCollected = amounts[booking.id] ? Number(amounts[booking.id]) : undefined
    const advancePaid = advances[booking.id] ? Number(advances[booking.id]) : undefined
    // ⚠️ ما نشيل المهمة من القائمة إلا بعد نجاح الحفظ: شيلها عند
    // الفشل يخلي الفني يظن إنه سلّم الشغل والمبلغ، والحجز يبقى مفتوح
    // بالنظام وماكو منو يدري.
    if (!(await guard.run('إنهاء الحجز', () => api.completeBooking(booking.id, {
      completionNotes: notes[booking.id] || undefined,
      amountCollected,
      advancePaid,
    })))) return
    setBookings((prev) => prev.filter((b) => b.id !== booking.id))
    setPaperwork({ booking, stopped: false })
  }

  // توقف العمل: السبب إجباري، وبعده ينطلب تقرير — والتقرير يتأجل إذا حب.
  const handleStopWork = async () => {
    if (!stopFor || !stopReason.trim()) return
    setStopping(true)
    try {
      const updated = await api.stopBookingWork(stopFor.id, stopReason.trim())
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
      setPaperwork({ booking: updated, stopped: true })
      setStopFor(null)
      setStopReason('')
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تسجيل توقف العمل')
    } finally {
      setStopping(false)
    }
  }

  return (
    <>
      <SaveError message={guard.error} onClose={guard.clear} />
    <div dir="rtl">
      {/* ═══ العنوان ═══ */}
      <div className="flex items-center gap-3">
        <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-brand-50 text-lg sm:h-11 sm:w-11 sm:text-xl">☑️</span>
        <div className="min-w-0">
          <h2 className="text-xl font-black text-[#0f2040] sm:text-2xl">مهامي</h2>
          <p className="text-[11px] text-slate-500 sm:text-xs">
            المهام المكلّف بيها حالياً وإجراءات التنفيذ المعتمدة
          </p>
        </div>
      </div>

      {/* ═══ ثلاثة خيارات جوّا الشاشة مو ثلاث بنود بالقائمة ═══
          «مهامي» و«فواتير الليدر» و«التقارير» كانوا ثلاث بنود منفصلة
          بالقائمة الجانبية — والثلاثة نفس الشغل: حجزك، فاتورة حجزك،
          وتقرير حجزك. الموظف يخلص شغلة ولازم يطلع للقائمة ويدور على
          البند الثاني حتى يكمّلها.

          هسه القائمة بيها «مهامي» وحدها، والثلاثة خيارات جوّاها.

          ⚠️ الخيار ما يطلع إلا للي عنده صلاحيته: فاتح «مهامي» بلا
          صلاحية فواتير ما يشوف تبويب فواتير — تبويب يفتح شاشة تگله
          «ممنوع» أسوأ من تبويب ما موجود. */}
      {!loading && (
        <div className="mt-4 grid grid-cols-3 gap-1.5 rounded-2xl border border-slate-200 bg-white p-1.5 shadow-[0_2px_12px_rgba(15,32,64,0.05)] sm:inline-grid sm:gap-2">
          {TABS.filter((t) => t.key === 'bookings'
            || (t.key === 'invoices' && canSeeInvoices)
            || (t.key === 'reports' && canSeeReports)).map((t) => (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={`rounded-xl px-3 py-2 text-[11px] font-extrabold transition sm:px-5 sm:text-xs ${
                tab === t.key
                  ? 'bg-gradient-to-l from-brand-500 to-brand-800 text-white shadow-md'
                  : 'text-slate-600 hover:bg-slate-50'
              }`}
            >
              {t.icon} {t.label}
            </button>
          ))}
        </div>
      )}

      {/* التبويبان الثانيان يعيدان استعمال نفس الشاشتين الموجودتين —
          ما ننسخ منطق الفواتير ولا التقارير بمكان ثاني، وإلا صارت
          نسختين تفترقن أول تعديل. */}
      {tab === 'invoices' && <div className="mt-4"><LeaderInvoicesListPage /></div>}
      {tab === 'reports' && <div className="mt-4"><WorkReportPage /></div>}

      {loading && tab === 'bookings' && <p className="mt-6 text-slate-400">جاري التحميل...</p>}

      {/* ═══ الأرقام — ٢×٢ بالموبايل ═══
          ⚠️ هاي شاشة **الميدان**: الفني يفتحها من تلفونه وهو بالسيارة
          أو عند الزبون. أربع بطاقات بصف واحد تنضغط وتصير أرقام ما
          تنقرا على شاشة ٦ إنچ. */}
      {!loading && tab === 'bookings' && (
        <div className="mt-4 grid grid-cols-2 gap-2.5 lg:grid-cols-4 sm:gap-3">
          <TaskStat
            icon="📅" label="مهام اليوم" tone="sky" unit="مهام"
            value={myTasks.filter((b) => b.scheduledAt && new Date(b.scheduledAt).toDateString() === todayKey).length}
          />
          <TaskStat
            icon="⚙️" label="قيد التنفيذ" tone="violet" unit="مهمة"
            value={myTasks.filter((b) => b.status === 'IN_PROGRESS').length}
          />
          <TaskStat
            icon="⏳" label="بانتظار البدء" tone="amber" unit="مهام"
            value={myTasks.filter((b) => b.status !== 'IN_PROGRESS').length}
          />
          <TaskStat
            icon="✅" label="مكتملة هذا الشهر" tone="emerald" unit="مهمة"
            value={myDone.filter((b) => b.completedAt && new Date(b.completedAt).getTime() >= monthAgo).length}
          />
        </div>
      )}

      {!loading && tab === 'bookings' && (
        <div className="mt-4 sm:mt-6">
          <div>
            {/* المهام الموجّهة من المدير فوق مهام الحجوزات: شغل موجّه
                لك بالاسم، ولو انحط بأسفل الصفحة راح ينتنسى. */}
            <MyExtraTasks />

            <h3 className="mb-3 font-bold text-brand-800">المهام الحالية</h3>
            <div className="flex flex-col gap-3">
              {myTasks.map((b) => {
                const myRole = b.assignments.find((a) => a.employee.id === employee?.id)?.role
                return (
                  <div
                    key={b.id}
                    className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
                  >
                    {/* هوية كاملة: الفني كان يشوف كود الحجز بس */}
                    <EntityIdentity booking={b} variant="full" className="mb-2" />

                    {/* ملاحظة الإداري للكادر — هاي كانت تنقال بالتلفون
                        وتضيع. الي ما كان بالمكالمة ما يعرفها. */}
                    {b.crewNotes && (
                      <div className="mb-2 rounded-lg border border-sky-200 bg-sky-50 px-3 py-2 text-sm">
                        <span className="font-bold text-sky-900">📋 ملاحظة الإداري:</span>{' '}
                        <span className="text-sky-900">{b.crewNotes}</span>
                        {b.crewNotesByName && (
                          <span className="mr-1 text-[11px] text-sky-700">— {b.crewNotesByName}</span>
                        )}
                      </div>
                    )}
                    {/* الحجز المعيّن لليدر قبل التثبيت: يشوفه بس يعرف
                        إنه لسه ما انثبت، فما يروح للزبون بلا موعد. */}
                    {b.status === 'PENDING' && (
                      <div className="mb-2 rounded-lg border border-amber-200 bg-amber-50 px-3 py-2 text-xs font-bold text-amber-800">
                        ⏳ لسه ما انثبت — انتظر التثبيت والموعد قبل ما تتحرك
                      </div>
                    )}
                    <div className="flex items-center justify-between">
                      <span className="font-mono text-sm font-semibold text-brand-600">
                        {b.code}
                      </span>
                      <span className="rounded-full bg-brand-50 px-3 py-1 text-xs font-bold text-brand-700">
                        {myRole === 'TECH_1'
                          ? 'الفني الأول'
                          : myRole === 'TECH_2'
                            ? 'الفني الثاني'
                            : 'الفني الثالث'}
                      </span>
                    </div>
                    <p className="mt-2 text-sm font-medium text-brand-800">{b.service?.name}</p>
                    {b.scheduledAt && (
                      <p className="mt-1 inline-block rounded-lg bg-amber-50 px-2 py-1 text-sm font-bold text-amber-800">
                        🕒 الموعد: {formatScheduleWindow(b.scheduledAt, b.scheduledEndAt)}
                      </p>
                    )}
                    <div className="mt-1 grid grid-cols-1 gap-1 text-sm text-slate-500 sm:grid-cols-2">
                      <p>
                        <span className="text-slate-400">الزبون: </span>
                        {b.customer?.name || 'زبون غير معروف'}
                      </p>
                      <p>
                        <span className="text-slate-400">الهاتف: </span>
                        {b.customer?.phone || '-'}
                      </p>
                      <p>
                        <span className="text-slate-400">العنوان: </span>
                        {b.address || b.customer?.location || 'بدون موقع محدد'}
                        {' '}
                        <button
                          type="button"
                          onClick={() => setDirectionsFor(b)}
                          className="mr-1 rounded-md bg-brand-50 px-2 py-0.5 text-xs font-bold text-brand-700 hover:bg-brand-100"
                        >
                          🗺️ الطريق
                        </button>
                      </p>
                      <p>
                        <span className="text-slate-400">السيارة: </span>
                        {b.assignedVehicle || 'لم تحدد'}
                      </p>
                      <p>
                        <span className="text-slate-400">التكلفة المقدرة: </span>
                        {b.quotedPrice != null ? b.quotedPrice.toLocaleString() : 'غير محددة'}
                      </p>
                    </div>
                    {b.notes && (
                      <p className="mt-1 text-sm text-slate-500">
                        <span className="text-slate-400">ملاحظات: </span>
                        {b.notes}
                      </p>
                    )}

                    {b.status === 'CONFIRMED' ? (
                      amLeaderOf(b) ? (
                      /* ═══ الليدر يسوق المسار ═══ */
                      <div className="mt-3 space-y-2">
                        {b.materialsReadyAt ? (
                          <div className="rounded-lg border-2 border-red-300 bg-red-50 px-4 py-3 text-center">
                            <p className="font-bold text-red-700">⏰ المواد جاهزة — انطلق الآن!</p>
                            <p className="mt-1 text-xs text-red-600">
                              جهّزها {b.materialsReadyBy?.name || 'تيم ليدر الفريق'} من {elapsedSince(b.materialsReadyAt)}
                            </p>
                          </div>
                        ) : (
                          <button
                            onClick={() => handleMaterialsReady(b)}
                            className="w-full rounded-lg bg-gradient-to-l from-purple-500 to-purple-700 px-4 py-3 text-sm font-bold text-white shadow-md transition-all hover:shadow-lg"
                          >
                            📦 تم تجهيز المواد — أبلغ الفريق
                          </button>
                        )}
                        {b.arrivedAt ? (
                          <div className="rounded-lg bg-sky-50 px-3 py-2 text-xs font-bold text-sky-700">
                            📍 وصلنا للزبون من {elapsedSince(b.arrivedAt)}
                          </div>
                        ) : (
                          <button
                            onClick={() => handleArrive(b)}
                            className="w-full rounded-lg bg-gradient-to-l from-sky-500 to-sky-700 px-4 py-3 text-sm font-bold text-white shadow-md transition-all hover:shadow-lg"
                          >
                            📍 انطلقنا / وصلنا للزبون — عن الفريق كله
                          </button>
                        )}
                        <button
                          onClick={() => openToolsCheck(b, false)}
                          className="w-full rounded-lg bg-gradient-to-l from-amber-500 to-amber-700 px-4 py-3 text-sm font-bold text-white shadow-md transition-all hover:shadow-lg"
                        >
                          ✅ بدأنا بالعمل — عن الفريق كله
                        </button>
                      </div>
                      ) : (
                      /* ═══ الفني: يتابع بس، وشغله جرد أدواته ═══ */
                      <div className="mt-3 space-y-2">
                        <div className="rounded-lg border border-slate-200 bg-slate-50 px-3 py-2 text-xs text-slate-600">
                          {b.arrivedAt
                            ? `📍 الفريق وصل للزبون من ${elapsedSince(b.arrivedAt)}`
                            : b.materialsReadyAt
                              ? `📦 المواد جاهزة من ${elapsedSince(b.materialsReadyAt)} — انتظر إشارة الليدر`
                              : '⏳ بانتظار الليدر يجهّز المواد'}
                          <span className="mt-1 block text-[11px] text-slate-400">
                            الانطلاق وبدء العمل بيد الليدر — ما تحتاج تضغط شي.
                          </span>
                        </div>
                        <button
                          onClick={() => openToolsCheck(b, true)}
                          className="w-full rounded-lg bg-gradient-to-l from-emerald-500 to-emerald-700 px-4 py-3 text-sm font-bold text-white shadow-md transition-all hover:shadow-lg"
                        >
                          🧰 جردت أدواتي وأدوات السيارة — تم
                        </button>
                      </div>
                      )
                    ) : (
                      <div className="mt-3">
                        <div className="mb-2 rounded-lg bg-amber-50 px-3 py-1 text-xs font-bold text-amber-700">
                          🔄 جاري التنفيذ
                        </div>
                        <div className="grid grid-cols-1 gap-2 sm:grid-cols-4">
                          <input
                            type="number"
                            placeholder="المبلغ المستلم"
                            value={amounts[b.id] || ''}
                            onChange={(e) => setAmounts((prev) => ({ ...prev, [b.id]: e.target.value }))}
                            className="rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none focus:border-brand-500 sm:py-2"
                          />
                          <input
                            type="number"
                            placeholder="دفعة مقدمة (إن وجدت)"
                            value={advances[b.id] || ''}
                            onChange={(e) => setAdvances((prev) => ({ ...prev, [b.id]: e.target.value }))}
                            className="rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none focus:border-brand-500 sm:py-2"
                          />
                          <input
                            placeholder="ملاحظات الإنجاز"
                            value={notes[b.id] || ''}
                            onChange={(e) => setNotes((prev) => ({ ...prev, [b.id]: e.target.value }))}
                            className="rounded-lg border border-slate-300 px-3 py-2.5 text-sm outline-none focus:border-brand-500 sm:py-2"
                          />
                          <button
                            onClick={() => handleComplete(b)}
                            className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 text-sm font-bold text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg sm:py-2"
                          >
                            ✅ تم الإنجاز
                          </button>
                          {/* الشغل الي ياخذ أكثر من يوم: بدل ما يأشّر
                              «تم الإنجاز» على شغل ناقص أو «توقف العمل»
                              وكأن الشغل فشل. */}
                          <button
                            onClick={() => setPartialFor(b)}
                            className="rounded-lg border border-amber-400 bg-amber-50 px-4 py-3 text-sm font-bold text-amber-800 transition-all hover:bg-amber-100 sm:py-2"
                          >
                            🔄 إنجاز جزئي (نكمل باچر)
                          </button>
                        </div>
                        {/* تقارير الأيام الفائتة — الكادر يقراها قبل ما يبدي */}
                        <BookingProgressTimeline bookingId={b.id} booking={b} />
                        {b.workStoppedAt ? (
                          <div className="mt-2 rounded-lg border border-slate-300 bg-slate-50 px-3 py-2 text-xs">
                            <span className="font-bold text-slate-700">⏸ العمل متوقف</span>
                            <span className="text-slate-500"> — {b.workStopReason}</span>
                            <button
                              onClick={async () => {
                                const u = await api.resumeBookingWork(b.id)
                                setBookings((prev) => prev.map((x) => (x.id === u.id ? u : x)))
                              }}
                              className="mr-2 rounded px-2 py-0.5 font-bold text-brand-700 underline"
                            >
                              رجعت أكمّل
                            </button>
                          </div>
                        ) : (
                          <button
                            onClick={() => { setStopFor(b); setStopReason('') }}
                            className="mt-2 w-full rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-bold text-slate-600 transition-all hover:bg-slate-50"
                          >
                            ⏸ توقف العمل
                          </button>
                        )}
                        <div className="hidden">
                        </div>
                        {employee?.isLeader && (
                          <button
                            onClick={() => navigate(`/leader-invoices/new?bookingId=${b.id}`)}
                            className="mt-2 w-full rounded-lg border border-brand-300 bg-brand-50 px-4 py-2 text-sm font-bold text-brand-700 transition-all hover:bg-brand-100"
                          >
                            🧾 إنشاء فاتورة ليدر لهذا الحجز
                          </button>
                        )}
                      </div>
                    )}

                    {(b.status === 'CONFIRMED' || b.status === 'IN_PROGRESS') && (
                      <button
                        onClick={() => navigate(`/procurement?bookingId=${b.id}`)}
                        className="mt-2 w-full rounded-lg border border-slate-300 bg-white px-4 py-2 text-sm font-medium text-slate-600 transition-all hover:bg-slate-50"
                      >
                        🧰 اطلب مادة ناقصة لهذا الحجز
                      </button>
                    )}
                  </div>
                )
              })}
              {myTasks.length === 0 && (
                <p className="text-slate-400">لا توجد مهام مسندة إليك حالياً.</p>
              )}
            </div>
          </div>

        </div>
      )}

      {toolsModalBooking && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-brand-900">تأكيد استلام الحجز {toolsModalBooking.code}</h3>
            <p className="mt-1 text-sm font-medium text-amber-700">
              علّم فقط الأداة الناقصة عندك — الباقي مؤشر مسبقاً كموجود.
            </p>

            {toolsLoading ? (
              <p className="mt-4 text-slate-400">جاري التحميل...</p>
            ) : (
              <div className="mt-4 max-h-64 space-y-2 overflow-y-auto">
                {personalTools.map((tool) => (
                  <label
                    key={tool.id}
                    className="flex cursor-pointer items-center gap-3 rounded-lg border border-slate-200 px-3 py-2"
                  >
                    <input
                      type="checkbox"
                      checked={!!checkedTools[tool.id]}
                      onChange={() =>
                        setCheckedTools((prev) => ({ ...prev, [tool.id]: !prev[tool.id] }))
                      }
                      className="h-5 w-5 accent-brand-600"
                    />
                    <span className="text-sm font-medium text-brand-900">{tool.name}</span>
                  </label>
                ))}
              </div>
            )}

            <div className="mt-6 flex gap-2">
              <button
                onClick={() => setToolsModalBooking(null)}
                className="flex-1 rounded-lg border border-slate-300 px-4 py-2.5 text-sm font-medium text-slate-600 transition-all hover:bg-slate-50"
              >
                إلغاء
              </button>
              <button
                onClick={handleConfirmToolsCheck}
                disabled={submittingAccept || toolsLoading}
                className="flex-1 rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2.5 text-sm font-bold text-white shadow-md transition-all hover:shadow-lg disabled:opacity-50"
              >
                {submittingAccept ? 'جارٍ التأكيد...' : 'تم'}
              </button>
            </div>
          </div>
        </div>
      )}

      {partialFor && (
        <PartialCompleteDialog
          booking={partialFor}
          onClose={() => setPartialFor(null)}
          onDone={() => {
            // الحجز رجع للإداري — ما عاد بمهام هذا الموظف
            setBookings((prev) => prev.filter((x) => x.id !== partialFor.id))
            setPartialFor(null)
          }}
        />
      )}

      {/* ═══ سبب توقف العمل — إجباري ═══ */}
      {stopFor && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-[#0f2040]">⏸ توقف العمل — حجز {stopFor.code}</h3>
            <p className="mt-1 text-xs text-slate-500">
              اكتب سبب التوقف. بدونه ما ينفع لا للمتابعة ولا للتقرير.
            </p>
            <textarea
              value={stopReason}
              onChange={(e) => setStopReason(e.target.value)}
              rows={3}
              placeholder="مثال: الزبون مو موجود بالموقع / المواد ناقصة / عطل بالكهرباء"
              className="mt-3 w-full rounded-xl border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
            />
            <div className="mt-4 flex gap-2">
              <button
                onClick={handleStopWork}
                disabled={!stopReason.trim() || stopping}
                className="flex-1 rounded-xl bg-gradient-to-l from-slate-600 to-slate-800 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
              >
                {stopping ? 'جارٍ التسجيل...' : 'سجّل توقف العمل'}
              </button>
              <button
                onClick={() => setStopFor(null)}
                className="rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-600"
              >
                رجوع
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ═══ بعد الإنجاز أو التوقف: الورق الباقي ═══
          مو إجباري — يكدر يأجله بـ«بعدين». بس ما يختفي بالسكوت مثل قبل. */}
      {paperwork && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
          <div className="w-full max-w-md rounded-2xl bg-white p-6 shadow-2xl">
            <h3 className="text-lg font-bold text-[#0f2040]">
              {paperwork.stopped ? '⏸ العمل توقف' : '✅ تم الإنجاز'} — حجز {paperwork.booking.code}
            </h3>
            <p className="mt-1 text-sm text-slate-600">
              {paperwork.stopped
                ? 'باقي عليك تقرير يوضّح شنو صار ووين وقف الشغل.'
                : 'باقي عليك ورقتين حتى يطلع الحجز «منجز بشكل كامل»:'}
            </p>
            <div className="mt-4 space-y-2">
              {!paperwork.stopped && (
                <button
                  onClick={() => navigate(`/leader-invoices/new?bookingId=${paperwork.booking.id}`)}
                  className="w-full rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-3 text-sm font-bold text-white"
                >
                  🧾 سوّي فاتورة التكاليف الآن
                </button>
              )}
              <button
                onClick={() => navigate(`/work-reports?bookingId=${paperwork.booking.id}`)}
                className="w-full rounded-xl border-2 border-brand-300 bg-brand-50 px-4 py-3 text-sm font-bold text-brand-700"
              >
                📝 سوّي تقرير العمل الآن
              </button>
              <button
                onClick={() => setPaperwork(null)}
                className="w-full rounded-xl border border-slate-300 px-4 py-2.5 text-sm font-bold text-slate-500"
              >
                بعدين
              </button>
            </div>
            <p className="mt-3 text-[11px] leading-relaxed text-amber-700">
              ⚠ الحجز راح يبقى مؤشّر «منجز بدون فاتورة/تقرير» بتنسيق الحجوزات لين تخلّصهن.
            </p>
          </div>
        </div>
      )}

      {/* ═══ شغلي الي خلّصته ═══
          قبل، الحجز يختفي من الشاشة أول ما ينضغط «تم الإنجاز» — فالموظف
          ما عاد يشوف شغله ولا تفاصيله ولا يعرف باقي عليه ورق لو لا. */}
      {myDone.length > 0 && (
        <div className="mt-8">
          <h3 className="mb-3 font-bold text-brand-800">✅ شغلي المنجز ({myDone.length})</h3>
          <div className="flex flex-col gap-2">
            {myDone.map((b) => (
              <div key={b.id} className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                <div className="flex flex-wrap items-center gap-2">
                  <span className="font-mono text-sm font-semibold text-brand-600">{b.code}</span>
                  <span className="text-sm font-bold text-slate-700">{b.customer?.name || '—'}</span>
                  <span className="text-xs text-slate-400">{(b.services?.length ? b.services.map((s) => s.name).join(" + ") : b.service?.name) || "—"}</span>
                  <span className="mr-auto"><CompletionBadge booking={b} /></span>
                </div>
                <div className="mt-2 grid grid-cols-1 gap-1 text-xs text-slate-500 sm:grid-cols-2">
                  <p>🏁 انتهى: {b.completedAt ? new Date(b.completedAt).toLocaleString('ar-IQ') : '—'}</p>
                  <p>💰 المستلم: {(b.amountCollected ?? 0).toLocaleString()} د.ع</p>
                  {b.address && <p className="sm:col-span-2">📍 {b.address}</p>}
                  {b.completionNotes && <p className="sm:col-span-2">📝 {b.completionNotes}</p>}
                </div>
                {/* الورق الباقي — يضل قدامه لين يخلّصه */}
                {(!b.hasInvoice || !b.hasReport) && amLeaderOf(b) && (
                  <div className="mt-2 flex flex-wrap gap-2">
                    {!b.hasInvoice && (
                      <button
                        onClick={() => navigate(`/leader-invoices/new?bookingId=${b.id}`)}
                        className="rounded-lg border border-brand-300 bg-brand-50 px-3 py-1.5 text-xs font-bold text-brand-700"
                      >
                        🧾 باقي عليك الفاتورة
                      </button>
                    )}
                    {!b.hasReport && (
                      <button
                        onClick={() => navigate(`/work-reports?bookingId=${b.id}`)}
                        className="rounded-lg border border-brand-300 bg-brand-50 px-3 py-1.5 text-xs font-bold text-brand-700"
                      >
                        📝 باقي عليك التقرير
                      </button>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        </div>
      )}

      {directionsFor && <DirectionsModal booking={directionsFor} onClose={() => setDirectionsFor(null)} />}
    </div>
    </>
  )
}

/* ───── بطاقة رقم ميدانية ───── */

function TaskStat({ icon, label, value, tone, unit }: {
  icon: string; label: string; value: number; unit: string
  tone: 'sky' | 'violet' | 'amber' | 'emerald'
}) {
  const tones: Record<string, { t: string; b: string }> = {
    sky:     { t: 'text-sky-700',     b: 'bg-sky-50' },
    violet:  { t: 'text-violet-700',  b: 'bg-violet-50' },
    amber:   { t: 'text-amber-700',   b: 'bg-amber-50' },
    emerald: { t: 'text-emerald-700', b: 'bg-emerald-50' },
  }
  // الصفر ما ينلوّن: «٠ قيد التنفيذ» مو مشكلة — تلوينه يخلي الفني
  // يحس إن أكو شي ناقص عليه وهو أنجز كلشي.
  const c = value > 0 ? tones[tone] : { t: 'text-slate-500', b: 'bg-slate-100' }
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-3 shadow-sm sm:p-4">
      <div className="flex items-start justify-between gap-1.5">
        <p className="text-[10px] font-medium leading-tight text-slate-500 sm:text-[11px]">{label}</p>
        <span className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-lg text-xs sm:h-8 sm:w-8 sm:text-sm ${c.b}`}>{icon}</span>
      </div>
      <p className={`mt-1 text-xl font-black sm:text-2xl ${c.t}`}>{value}</p>
      <p className="text-[9px] text-slate-400 sm:text-[10px]">{unit}</p>
    </div>
  )
}
