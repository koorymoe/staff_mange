import { useEffect, useRef, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import L from 'leaflet'
import 'leaflet/dist/leaflet.css'
import { formatScheduleWindow } from '../utils/schedule'
import { api, type Booking, type PersonalTool } from '../api'
import { useSession } from '../session'

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
  const { employee } = useSession()
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

  useEffect(() => {
    const t = setInterval(() => setTick((v) => v + 1), 30000)
    return () => clearInterval(t)
  }, [])

  const load = () => {
    Promise.all([api.getBookings({ status: 'CONFIRMED' }), api.getBookings({ status: 'IN_PROGRESS' })])
      .then(([confirmed, inProgress]) => setBookings([...confirmed, ...inProgress]))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const myTasks = bookings.filter((b) =>
    b.assignments.some((a) => a.employee.id === employee?.id),
  )

  const handleArrive = async (booking: Booking) => {
    const updated = await api.markArrived(booking.id)
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  // نعتبرها الاستلام الفعلي (تنفيذ startBooking) — تُستدعى مباشرة لو الموظف
  // ما عنده أدوات شخصية مسجلة أصلاً (نتخطى المودال حتى لا يعلق بواجهة فاضية)،
  // أو بعد ما يضغط "تم" بمودال شيك الأدوات.
  const doStart = async (bookingId: string, missingToolIds?: string[]) => {
    const updated = await api.startBooking(bookingId, missingToolIds)
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  const handleStart = async (booking: Booking) => {
    if (!employee) {
      await doStart(booking.id)
      return
    }
    setToolsModalBooking(booking)
    setToolsLoading(true)
    try {
      const tools = await api.getPersonalTools(employee.id)
      if (tools.length === 0) {
        // لا توجد أدوات مسجلة لهذا الموظف — نتخطى المودال بالكامل ونكمل الاستلام
        // عادي، بدل ما نعلقه بشاشة فاضية بلا فايدة.
        setToolsModalBooking(null)
        await doStart(booking.id)
        return
      }
      setPersonalTools(tools)
      // كل الأدوات مؤشرة (موجودة) افتراضياً — الموظف يشيل التأشير فقط عن الناقص.
      const allChecked: Record<string, boolean> = {}
      tools.forEach((t) => { allChecked[t.id] = true })
      setCheckedTools(allChecked)
    } catch {
      // تعذر جلب الأدوات — لا نمنع الاستلام، نكمل عادي بدون شيك.
      setToolsModalBooking(null)
      await doStart(booking.id)
    } finally {
      setToolsLoading(false)
    }
  }

  const handleConfirmToolsCheck = async () => {
    if (!toolsModalBooking) return
    setSubmittingAccept(true)
    try {
      const missingToolIds = personalTools.filter((t) => !checkedTools[t.id]).map((t) => t.id)
      await doStart(toolsModalBooking.id, missingToolIds)
      setToolsModalBooking(null)
    } catch (e) {
      alert(e instanceof Error ? e.message : 'تعذر تأكيد الاستلام')
    } finally {
      setSubmittingAccept(false)
    }
  }

  const handleMaterialsReady = async (booking: Booking) => {
    const updated = await api.setMaterialsReady(booking.id)
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  // ═══ بعد الإنجاز: الورق ما ينتسى ═══
  // الإنجاز لحاله ما يكفي — باقي فاتورة التكاليف المربوطة بالحجز وتقرير
  // العمل. قبل، الحجز جان يختفي من الشاشة أول ما يضغط «تم الإنجاز»
  // فينسى الورق ويطلع الحجز «منجز» وهو ناقص. هسه تطلع له مطالبة بيها
  // الاثنين — بس مو إجبارية: يكدر يأجلها بـ«بعدين» ويرجع لها.
  const [paperwork, setPaperwork] = useState<{ booking: Booking; stopped: boolean } | null>(null)
  const [stopFor, setStopFor] = useState<Booking | null>(null)
  const [stopReason, setStopReason] = useState('')
  const [stopping, setStopping] = useState(false)

  const handleComplete = async (booking: Booking) => {
    const amountCollected = amounts[booking.id] ? Number(amounts[booking.id]) : undefined
    const advancePaid = advances[booking.id] ? Number(advances[booking.id]) : undefined
    await api.completeBooking(booking.id, {
      completionNotes: notes[booking.id] || undefined,
      amountCollected,
      advancePaid,
    })
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
    <div>
      <h2 className="text-2xl font-bold text-brand-900">مهامي</h2>
      <p className="mt-1 text-slate-500">
        المهام المكلف بها حالياً، ومهاراتك المعتمدة بالنظام.
      </p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}

      {!loading && (
        <div className="mt-6 grid grid-cols-1 gap-6 lg:grid-cols-3">
          <div className="lg:col-span-2">
            <h3 className="mb-3 font-bold text-brand-800">المهام الحالية</h3>
            <div className="flex flex-col gap-3">
              {myTasks.map((b) => {
                const myRole = b.assignments.find((a) => a.employee.id === employee?.id)?.role
                return (
                  <div
                    key={b.id}
                    className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
                  >
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
                      <div className="mt-3 space-y-2">
                        {b.materialsReadyAt ? (
                          <div className="rounded-lg border-2 border-red-300 bg-red-50 px-4 py-3 text-center">
                            <p className="font-bold text-red-700">⏰ المواد جاهزة — انطلق الآن!</p>
                            <p className="mt-1 text-xs text-red-600">
                              جهّزها {b.materialsReadyBy?.name || 'تيم ليدر الفريق'} من {elapsedSince(b.materialsReadyAt)}
                            </p>
                          </div>
                        ) : employee?.isLeader ? (
                          <button
                            onClick={() => handleMaterialsReady(b)}
                            className="w-full rounded-lg bg-gradient-to-l from-purple-500 to-purple-700 px-4 py-3 text-sm font-bold text-white shadow-md transition-all hover:shadow-lg"
                          >
                            📦 تم تجهيز المواد — أبلغ الفريق
                          </button>
                        ) : null}
                        {b.arrivedAt ? (
                          <div className="rounded-lg bg-sky-50 px-3 py-2 text-xs font-bold text-sky-700">
                            📍 وصلت للزبون من {elapsedSince(b.arrivedAt)}
                          </div>
                        ) : (
                          <button
                            onClick={() => handleArrive(b)}
                            className="w-full rounded-lg bg-gradient-to-l from-sky-500 to-sky-700 px-4 py-3 text-sm font-bold text-white shadow-md transition-all hover:shadow-lg"
                          >
                            📍 وصلت للزبون
                          </button>
                        )}
                        <button
                          onClick={() => handleStart(b)}
                          className="w-full rounded-lg bg-gradient-to-l from-amber-500 to-amber-700 px-4 py-3 text-sm font-bold text-white shadow-md transition-all hover:shadow-lg"
                        >
                          ✅ تم الاستلام — بدأت بالعمل
                        </button>
                      </div>
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
                            className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                          />
                          <input
                            type="number"
                            placeholder="دفعة مقدمة (إن وجدت)"
                            value={advances[b.id] || ''}
                            onChange={(e) => setAdvances((prev) => ({ ...prev, [b.id]: e.target.value }))}
                            className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                          />
                          <input
                            placeholder="ملاحظات الإنجاز"
                            value={notes[b.id] || ''}
                            onChange={(e) => setNotes((prev) => ({ ...prev, [b.id]: e.target.value }))}
                            className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                          />
                          <button
                            onClick={() => handleComplete(b)}
                            className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg"
                          >
                            تم الإنجاز
                          </button>
                        </div>
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

          <div>
            <h3 className="mb-3 font-bold text-brand-800">مهاراتي المعتمدة</h3>
            <div className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <div className="flex flex-wrap gap-2">
                {employee?.skills
                  .filter((s) => s.canPerform)
                  .map((s) => (
                    <span
                      key={s.id}
                      className="rounded-full bg-brand-50 px-3 py-1 text-xs font-medium text-brand-700"
                    >
                      {s.skill.name}
                    </span>
                  ))}
                {(!employee || employee.skills.filter((s) => s.canPerform).length === 0) && (
                  <p className="text-sm text-slate-400">لم يتم تحديد مهارات بعد.</p>
                )}
              </div>
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

      {directionsFor && <DirectionsModal booking={directionsFor} onClose={() => setDirectionsFor(null)} />}
    </div>
  )
}
