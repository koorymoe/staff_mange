import { useEffect, useState } from 'react'
import { api, type Booking } from '../api'
import { useSession } from '../session'

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

function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

export default function WorkReportPage() {
  const { employee: currentUser } = useSession()
  const [bookings, setBookings] = useState<Booking[]>([])
  const [completedToday, setCompletedToday] = useState<Booking[]>([])
  const [reportedBookingIds, setReportedBookingIds] = useState<Set<string>>(new Set())
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [forms, setForms] = useState<Record<string, ReportForm>>({})
  const [submitting, setSubmitting] = useState(false)

  const load = () => {
    if (!currentUser) return
    Promise.all([
      api.getBookings({ status: 'IN_PROGRESS' }),
      api.getBookings({ status: 'COMPLETED', date: todayStr() }),
      api.getWorkReports(currentUser.id),
    ])
      .then(([inProgress, completed, myReports]) => {
        const isMine = (b: Booking) =>
          b.assignments.some((a) => a.employee.id === currentUser.id) ||
          b.projectSupervisor?.id === currentUser.id
        setBookings(inProgress.filter(isMine))
        setCompletedToday(completed.filter(isMine))
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
              <div className="flex items-center justify-between p-5">
                <div className="flex items-center gap-4">
                  <div className="flex h-12 w-12 items-center justify-center rounded-xl bg-brand-50 text-lg font-bold text-brand-700">
                    {booking.sequenceNumber || '#'}
                  </div>
                  <div>
                    <p className="font-bold text-brand-900">{booking.code}</p>
                    <p className="text-sm text-slate-500">{booking.customer?.name}</p>
                  </div>
                </div>
                <div className="flex items-center gap-4">
                  <div className="text-left">
                    <p className="text-sm text-slate-500">{booking.service?.name || '-'}</p>
                    <p className="text-xs text-slate-400">
                      {booking.scheduledAt
                        ? new Date(booking.scheduledAt).toLocaleDateString('ar-IQ')
                        : '-'}
                    </p>
                  </div>
                  <button
                    onClick={() => handleToggle(booking.id)}
                    className={`rounded-xl px-5 py-2.5 font-medium text-white shadow transition-all ${
                      expanded
                        ? 'bg-slate-400 hover:bg-slate-500'
                        : 'bg-gradient-to-l from-brand-500 to-brand-800 hover:shadow-lg hover:shadow-brand-900/30'
                    }`}
                  >
                    {expanded ? 'إغلاق' : 'رفع تقرير'}
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

  return (
    <div>
      <div>
        <h2 className="text-2xl font-bold text-brand-900">تقرير العمل</h2>
        <p className="mt-1 text-slate-500">
          حجوزاتك المنجزة اليوم وبانتظار تقرير، وحجوزاتك الجارية.
        </p>
      </div>

      <div className="mt-6">
        <h3 className="mb-3 text-lg font-bold text-brand-800">📋 حجوزات اليوم — بانتظار التقرير</h3>
        {needsReportToday.length === 0 ? (
          <div className="rounded-2xl border border-white bg-white p-6 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <p className="text-slate-400">ما عندك حجوزات منجزة اليوم تحتاج تقرير — ممتاز!</p>
          </div>
        ) : (
          <div className="space-y-4">{needsReportToday.map(renderCard)}</div>
        )}
      </div>

      {alreadyReportedToday.length > 0 && (
        <div className="mt-8">
          <h3 className="mb-3 text-lg font-bold text-emerald-700">✔ حجوزات اليوم — تم رفع تقريرها</h3>
          <div className="space-y-2">
            {alreadyReportedToday.map((b) => (
              <div key={b.id} className="flex items-center justify-between rounded-xl border border-emerald-100 bg-emerald-50 px-5 py-3">
                <div>
                  <p className="font-bold text-emerald-900">{b.code}</p>
                  <p className="text-sm text-emerald-700">{b.customer?.name}</p>
                </div>
                <span className="text-sm font-bold text-emerald-600">✔ تم الإرسال</span>
              </div>
            ))}
          </div>
        </div>
      )}

      <div className="mt-8">
        <h3 className="mb-3 text-lg font-bold text-brand-800">🚧 الحجوزات الجارية</h3>
        {bookings.length === 0 ? (
          <div className="rounded-2xl border border-white bg-white p-6 text-center shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            <p className="text-slate-400">لا توجد حجوزات جارية مسندة إليك حالياً</p>
          </div>
        ) : (
          <div className="space-y-4">{bookings.map(renderCard)}</div>
        )}
      </div>
    </div>
  )
}
