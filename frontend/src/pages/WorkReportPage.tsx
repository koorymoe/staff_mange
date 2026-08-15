import { useEffect, useState } from 'react'
import { api, type Booking, type WorkReport } from '../api'
import { useSession } from '../session'
import EntityIdentity from '../components/EntityIdentity'

type WorkStatus = 'COMPLETED' | 'STOPPED' | null

interface ReportForm {
  workStatus: WorkStatus
  events: string
  customerRequests: string
  cleanedPlace: boolean
  gaveInfo: boolean
  tookPhotos: boolean
  additionalNotes: string
  stopReason: string
  stopNotes: string
}

const emptyForm: ReportForm = {
  workStatus: null,
  events: '',
  customerRequests: '',
  cleanedPlace: false,
  gaveInfo: false,
  tookPhotos: false,
  additionalNotes: '',
  stopReason: '',
  stopNotes: '',
}

export default function WorkReportPage() {
  const { employee: currentUser } = useSession()
  const [bookings, setBookings] = useState<Booking[]>([])
  const [completedToday, setCompletedToday] = useState<Booking[]>([])
  const [reportedBookingIds, setReportedBookingIds] = useState<Set<string>>(new Set())
  // تقاريري — نحتفظ بيها كاملة حتى نحسب أرقام اليوم والأسبوع
  const [myReports, setMyReports] = useState<WorkReport[]>([])
  const [tab, setTab] = useState<'pending' | 'running' | 'done'>('pending')
  // ⚠️ الوقت ينثبت مرة وحدة بفتح الشاشة: قراءة الساعة أثناء الرندر
  // تخلي النتيجة تتغيّر بلا سبب والأرقام تنط قدام الموظف. وفوق هنا
  // مو تحت — الخطّافات لازم تنستدعى قبل أي return مبكر.
  const [{ todayKey, weekAgo }] = useState(() => ({
    todayKey: new Date().toDateString(),
    weekAgo: Date.now() - 7 * 24 * 60 * 60 * 1000,
  }))
  const [search, setSearch] = useState('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [forms, setForms] = useState<Record<string, ReportForm>>({})
  const [submitting, setSubmitting] = useState(false)

  const load = () => {
    if (!currentUser) return
    Promise.all([
      api.getBookings({ status: 'IN_PROGRESS' }),
      // بدون قيد تاريخ اليوم — حتى الحجز المنجز من كم يوم ونسى الموظف يسوي
      // تقريره يضل يطلع هنا لين يسويه، مو يختفي بس لأن اليوم تغيّر.
      api.getBookings({ status: 'COMPLETED' }),
      api.getWorkReports(currentUser.id),
    ])
      .then(([inProgress, completed, myReports]) => {
        const isMine = (b: Booking) =>
          b.assignments.some((a) => a.employee.id === currentUser.id) ||
          b.projectSupervisor?.id === currentUser.id
        setBookings(inProgress.filter(isMine))
        setCompletedToday(completed.filter(isMine))
        setMyReports(myReports)
        setReportedBookingIds(new Set(myReports.map((r) => r.bookingId)))
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }
  useEffect(load, [currentUser])

  const getForm = (id: string) => forms[id] || emptyForm
  const updateForm = (id: string, patch: Partial<ReportForm>) =>
    setForms((prev) => ({ ...prev, [id]: { ...getForm(id), ...patch } }))

  const handleToggle = (id: string) => {
    setExpandedId(expandedId === id ? null : id)
  }

  const handleSubmit = async (bookingId: string) => {
    const form = getForm(bookingId)
    if (form.workStatus === 'COMPLETED' && !form.events.trim()) {
      alert('يرجى كتابة تقرير الأحداث والمشاكل')
      return
    }
    if (form.workStatus === 'STOPPED' && !form.stopReason.trim()) {
      alert('يرجى كتابة سبب التوقف')
      return
    }
    setSubmitting(true)
    try {
      await api.createWorkReport({
        bookingId,
        workStatus: form.workStatus!,
        events: form.events || undefined,
        extraRequests: form.customerRequests || undefined,
        cleanedSite: form.cleanedPlace,
        gaveInfo: form.gaveInfo,
        tookPhotos: form.tookPhotos,
        stopReason: form.stopReason || undefined,
        notes: form.additionalNotes || form.stopNotes || undefined,
      })
      setExpandedId(null)
      setForms((prev) => {
        const next = { ...prev }
        delete next[bookingId]
        return next
      })
      setReportedBookingIds((prev) => new Set(prev).add(bookingId))
      load() // نعيد الجلب حتى تنضبط أرقام اليوم والأسبوع
      alert('تم إرسال التقرير بنجاح ✓')
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر إرسال التقرير')
    } finally {
      setSubmitting(false)
    }
  }

  if (loading) return <p className="mt-6 text-slate-400">جاري التحميل...</p>
  if (error)
    return (
      <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">
        تعذر الاتصال بالخادم: {error}
      </p>
    )

  const needsReportToday = completedToday.filter((b) => !reportedBookingIds.has(b.id))
  const alreadyReportedToday = completedToday.filter((b) => reportedBookingIds.has(b.id))

  const renderCard = (booking: Booking) => {
    const expanded = expandedId === booking.id
    const form = getForm(booking.id)

    return (
      <div
        key={booking.id}
        className="rounded-2xl border border-white bg-white shadow-[0_4px_20px_rgba(15,32,64,0.06)] overflow-hidden"
      >
              {/* Booking header */}
              {/* ⚠️ متجاوب بقصد: الفني يفتح هاي الشاشة **من الموبايل وهو
                  بالميدان**، مو من كمبيوتر بالمكتب. الصف الأفقي ينضغط
                  ويصير غير مقروء على شاشة ٦ إنچ — فينكسر لأعمدة، والزر
                  يصير بعرض الشاشة كاملة حتى ينضغط بالإبهام. */}
              <div className="flex flex-col gap-3 p-4 sm:flex-row sm:items-center sm:justify-between sm:p-5">
                <div className="flex items-center gap-3 sm:gap-4">
                  <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-brand-50 text-base font-bold text-brand-700 sm:h-12 sm:w-12 sm:text-lg">
                    {booking.sequenceNumber || '#'}
                  </div>
                  <div className="min-w-0">
                    <p className="font-mono text-sm font-black text-brand-900 sm:text-base">{booking.code}</p>
                    <p className="truncate text-xs text-slate-500 sm:text-sm">{booking.customer?.name}</p>
                    {booking.customer?.phone && (
                      <a href={`tel:${booking.customer.phone}`} className="text-[11px] font-bold text-brand-700 underline sm:hidden">
                        📞 {booking.customer.phone}
                      </a>
                    )}
                  </div>
                </div>
                <div className="hidden md:block">
                  {/* الفني يكتب تقرير عن حجز — لازم يشوف لمنو ومنو الليدر
                      المسؤول قبل ما يوقّع على «تم الإنجاز». */}
                  <EntityIdentity booking={booking} />
                </div>
                <div className="flex flex-col gap-2 sm:flex-row sm:items-center sm:gap-4">
                  <div className="flex flex-wrap gap-x-4 gap-y-1 text-[11px] text-slate-500 sm:block sm:text-left sm:text-xs">
                    <span className="sm:block sm:text-sm">🔧 {booking.service?.name || '—'}</span>
                    <span className="sm:block sm:text-xs sm:text-slate-400">
                      📅 {booking.scheduledAt ? new Date(booking.scheduledAt).toLocaleDateString('ar-IQ') : '—'}
                    </span>
                  </div>
                  <button
                    onClick={() => handleToggle(booking.id)}
                    className={`w-full shrink-0 rounded-xl px-5 py-3 text-sm font-bold text-white shadow transition-all sm:w-auto sm:py-2.5 ${
                      expanded
                        ? 'bg-slate-400 hover:bg-slate-500'
                        : 'bg-gradient-to-l from-brand-500 to-brand-800 hover:shadow-lg hover:shadow-brand-900/30'
                    }`}
                  >
                    {expanded ? 'إغلاق' : '📤 رفع التقرير'}
                  </button>
                </div>
              </div>

              {/* Report form */}
              {expanded && (
                <div className="border-t border-slate-100 p-5">
                  {/* Work status toggle */}
                  <p className="mb-3 font-semibold text-brand-800">حالة العمل</p>
                  <div className="grid grid-cols-2 gap-4">
                    <button
                      type="button"
                      onClick={() => updateForm(booking.id, { workStatus: 'COMPLETED' })}
                      className={`rounded-2xl border-2 p-5 text-center font-bold transition-all ${
                        form.workStatus === 'COMPLETED'
                          ? 'border-green-500 bg-green-50 text-green-700 shadow-md'
                          : 'border-slate-200 bg-white text-slate-500 hover:border-green-300'
                      }`}
                    >
                      <span className="mb-1 block text-3xl">&#10003;</span>
                      تم إنجاز العمل
                    </button>
                    <button
                      type="button"
                      onClick={() => updateForm(booking.id, { workStatus: 'STOPPED' })}
                      className={`rounded-2xl border-2 p-5 text-center font-bold transition-all ${
                        form.workStatus === 'STOPPED'
                          ? 'border-amber-500 bg-amber-50 text-amber-700 shadow-md'
                          : 'border-slate-200 bg-white text-slate-500 hover:border-amber-300'
                      }`}
                    >
                      <span className="mb-1 block text-3xl">&#9724;</span>
                      توقف العمل
                    </button>
                  </div>

                  {/* COMPLETED fields */}
                  {form.workStatus === 'COMPLETED' && (
                    <div className="mt-5 space-y-4">
                      <div>
                        <label className="mb-1 block text-sm font-medium text-slate-600">
                          الأحداث والمشاكل <span className="text-red-500">*</span>
                        </label>
                        <textarea
                          required
                          value={form.events}
                          onChange={(e) => updateForm(booking.id, { events: e.target.value })}
                          rows={4}
                          placeholder="اكتب تقريراً مفصلاً عن كل ما حدث"
                          className="w-full rounded-xl border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
                        />
                      </div>
                      <div>
                        <label className="mb-1 block text-sm font-medium text-slate-600">
                          طلبات الزبون الإضافية
                        </label>
                        <textarea
                          value={form.customerRequests}
                          onChange={(e) =>
                            updateForm(booking.id, { customerRequests: e.target.value })
                          }
                          rows={2}
                          className="w-full rounded-xl border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
                        />
                      </div>
                      <div className="space-y-2">
                        <label className="block text-sm font-medium text-slate-600">
                          قائمة التحقق
                        </label>
                        {[
                          { key: 'cleanedPlace' as const, label: 'تم تنظيف المكان' },
                          { key: 'gaveInfo' as const, label: 'تم إعطاء المعلومات اللازمة للزبون' },
                          { key: 'tookPhotos' as const, label: 'تم تصوير العمل' },
                        ].map(({ key, label }) => (
                          <label
                            key={key}
                            className="flex items-center gap-3 rounded-xl border border-slate-200 bg-slate-50 px-4 py-3 cursor-pointer hover:bg-slate-100 transition-colors"
                          >
                            <input
                              type="checkbox"
                              checked={form[key]}
                              onChange={(e) =>
                                updateForm(booking.id, { [key]: e.target.checked })
                              }
                              className="h-5 w-5 rounded accent-brand-600"
                            />
                            <span className="font-medium text-slate-700">{label}</span>
                          </label>
                        ))}
                      </div>
                      <div>
                        <label className="mb-1 block text-sm font-medium text-slate-600">
                          ملاحظات إضافية
                        </label>
                        <textarea
                          value={form.additionalNotes}
                          onChange={(e) =>
                            updateForm(booking.id, { additionalNotes: e.target.value })
                          }
                          rows={2}
                          className="w-full rounded-xl border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
                        />
                      </div>
                    </div>
                  )}

                  {/* STOPPED fields */}
                  {form.workStatus === 'STOPPED' && (
                    <div className="mt-5 space-y-4">
                      <div>
                        <label className="mb-1 block text-sm font-medium text-slate-600">
                          سبب التوقف <span className="text-red-500">*</span>
                        </label>
                        <textarea
                          required
                          value={form.stopReason}
                          onChange={(e) =>
                            updateForm(booking.id, { stopReason: e.target.value })
                          }
                          rows={4}
                          placeholder="اكتب سبباً مفصلاً لتوقف العمل"
                          className="w-full rounded-xl border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
                        />
                      </div>
                      <div>
                        <label className="mb-1 block text-sm font-medium text-slate-600">
                          ملاحظات
                        </label>
                        <textarea
                          value={form.stopNotes}
                          onChange={(e) =>
                            updateForm(booking.id, { stopNotes: e.target.value })
                          }
                          rows={2}
                          className="w-full rounded-xl border border-gray-300 px-4 py-3 text-right outline-none focus:border-brand-500"
                        />
                      </div>
                    </div>
                  )}

                  {/* Submit */}
                  {form.workStatus && (
                    <div className="mt-5">
                      <button
                        onClick={() => handleSubmit(booking.id)}
                        disabled={submitting}
                        className="w-full rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-6 py-3.5 font-bold text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg hover:shadow-brand-900/30 disabled:opacity-50"
                      >
                        {submitting ? 'جاري الإرسال...' : 'إرسال التقرير'}
                      </button>
                    </div>
                  )}
                </div>
              )}
      </div>
    )
  }

  // ═══ الأرقام + الفلاتر ═══
  //
  // ⚠️ الأرقام كلها حقيقية من بيانات الموظف. التصميم كان بيه «قيد
  // المراجعة» — وهاي حالة **ما موجودة بالنظام**: التقرير ينرفع
  // وخلص، ماكو دورة مراجعة ولا حالة تتغيّر. فحطينا محلها رقم يفيد
  // فعلاً (تقارير الأسبوع). عرض رقم لحالة ما تنوجد يخلي الموظف
  // ينتظر مراجعة ما راح تجي.
  const reportedToday = myReports.filter(r => new Date(r.createdAt).toDateString() === todayKey).length
  const reportedThisWeek = myReports.filter(r => new Date(r.createdAt).getTime() >= weekAgo).length

  const matches = (b: Booking) => {
    const q = search.trim()
    if (!q) return true
    return `${b.code} ${b.customer?.name ?? ''}`.includes(q)
  }
  const pendingList = needsReportToday.filter(matches)
  const doneList = alreadyReportedToday.filter(matches)
  const runningList = bookings.filter(matches)

  return (
    <div dir="rtl" className="space-y-4 sm:space-y-5">
      {/* ═══ العنوان ═══ */}
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex items-center gap-3">
          <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-sky-50 text-lg sm:h-11 sm:w-11 sm:text-xl">📄</span>
          <div className="min-w-0">
            <h2 className="text-xl font-black text-[#0f2040] sm:text-2xl">تقرير العمل</h2>
            <p className="text-[11px] text-slate-500 sm:text-xs">ارفع تقرير الحجوزات المنجزة وتابع الي بانتظار الرفع</p>
          </div>
        </div>
        <p className="rounded-xl border border-sky-200 bg-sky-50 px-3 py-2 text-[10px] leading-relaxed text-sky-800 sm:text-[11px]">
          ℹ️ <b>معلومة:</b> يترفع تقرير منفصل لكل حجز بعد الإنجاز
        </p>
      </div>

      {/* ═══ الأرقام — ٢×٢ بالموبايل و٤ بالشاشة الكبيرة ═══ */}
      <div className="grid grid-cols-2 gap-2.5 lg:grid-cols-4 sm:gap-3">
        <ReportStat icon="⏳" label="بانتظار الرفع" value={needsReportToday.length} tone="amber" unit="حجز" />
        <ReportStat icon="📤" label="تم الرفع اليوم" value={reportedToday} tone="sky" unit="تقرير" />
        <ReportStat icon="📅" label="تقارير الأسبوع" value={reportedThisWeek} tone="violet" unit="تقرير" />
        <ReportStat icon="✅" label="إجمالي تقاريري" value={myReports.length} tone="emerald" unit="تقرير" />
      </div>

      {/* ═══ الفلاتر ═══ */}
      <div className="flex flex-wrap items-center gap-2">
        <div className="flex flex-wrap gap-1.5">
          {([
            { k: 'pending', label: `بانتظار الرفع (${pendingList.length})` },
            { k: 'running', label: `جارية (${runningList.length})` },
            { k: 'done', label: `مرفوعة (${doneList.length})` },
          ] as const).map(t => (
            <button
              key={t.k}
              onClick={() => setTab(t.k)}
              className={`rounded-lg px-3 py-2 text-[11px] font-bold transition sm:text-xs ${
                tab === t.k ? 'bg-[#2c5aad] text-white shadow-md' : 'border border-slate-300 bg-white text-slate-600 hover:bg-slate-50'
              }`}
            >
              {t.label}
            </button>
          ))}
        </div>
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="🔍 ابحث برقم الحجز أو اسم العميل"
          className="min-w-[180px] flex-1 rounded-lg border border-slate-300 px-3 py-2 text-xs outline-none focus:border-sky-500"
        />
      </div>

      {/* ═══ القوائم ═══ */}
      {tab === 'pending' && (
        pendingList.length === 0
          ? <EmptyBox text="ما عندك حجوزات منجزة تحتاج تقرير — ممتاز!" icon="🎉" />
          : <div className="space-y-3">{pendingList.map(renderCard)}</div>
      )}

      {tab === 'running' && (
        runningList.length === 0
          ? <EmptyBox text="ماكو حجوزات جارية مسندة إلك حالياً" icon="🚧" />
          : <div className="space-y-3">{runningList.map(renderCard)}</div>
      )}

      {tab === 'done' && (
        doneList.length === 0
          ? <EmptyBox text="ماكو تقارير مرفوعة بعد" icon="📄" />
          : <div className="space-y-2">
              {doneList.map((b) => (
                <div key={b.id} className="flex flex-wrap items-center justify-between gap-2 rounded-xl border border-emerald-200 bg-emerald-50 px-4 py-3">
                  <div className="min-w-0">
                    <p className="font-mono text-sm font-black text-emerald-900">{b.code}</p>
                    <p className="truncate text-xs text-emerald-700">{b.customer?.name}</p>
                  </div>
                  <span className="rounded-full bg-emerald-100 px-3 py-1 text-[11px] font-bold text-emerald-700">✅ انرفع التقرير</span>
                </div>
              ))}
            </div>
      )}
    </div>
  )
}

/* ───── بطاقة رقم ───── */

function ReportStat({ icon, label, value, tone, unit }: {
  icon: string; label: string; value: number; unit: string
  tone: 'amber' | 'sky' | 'violet' | 'emerald'
}) {
  const tones: Record<string, { t: string; b: string }> = {
    amber:   { t: 'text-amber-700',   b: 'bg-amber-50' },
    sky:     { t: 'text-sky-700',     b: 'bg-sky-50' },
    violet:  { t: 'text-violet-700',  b: 'bg-violet-50' },
    emerald: { t: 'text-emerald-700', b: 'bg-emerald-50' },
  }
  // الصفر ما ينلوّن: «٠ بانتظار الرفع» خبر زين مو تحذير
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

function EmptyBox({ text, icon }: { text: string; icon: string }) {
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-8 text-center">
      <p className="text-3xl">{icon}</p>
      <p className="mt-2 text-sm font-bold text-slate-500">{text}</p>
    </div>
  )
}
