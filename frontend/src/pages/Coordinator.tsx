import { useEffect, useState } from 'react'
import { api, type Booking, type Employee } from '../api'

const techRoles: { key: 'TECH_1' | 'TECH_2' | 'TECH_3'; label: string }[] = [
  { key: 'TECH_1', label: 'الفني الأول' },
  { key: 'TECH_2', label: 'الفني الثاني' },
  { key: 'TECH_3', label: 'الفني الثالث' },
]

export default function Coordinator() {
  const [bookings, setBookings] = useState<Booking[]>([])
  const [matches, setMatches] = useState<Record<string, Employee[]>>({})
  const [supervisors, setSupervisors] = useState<Employee[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // قيم التثبيت (التكلفة المقدرة + العنوان) قبل الضغط على تثبيت
  const [priceDrafts, setPriceDrafts] = useState<Record<string, string>>({})
  const [addressDrafts, setAddressDrafts] = useState<Record<string, string>>({})

  const load = () => {
    setLoading(true)
    api
      .getBookings()
      .then((data) => {
        setBookings(data)
        data
          .filter((b) => b.status === 'CONFIRMED' && !b.transferToProjects && b.service)
          .forEach((b) => loadMatches(b))
      })
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])
  useEffect(() => {
    api.getSupervisors().then(setSupervisors)
  }, [])

  const handleSupervisorChange = async (booking: Booking, employeeId: string) => {
    const updated = await api.assignSupervisor(booking.id, employeeId || null)
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  const loadMatches = async (booking: Booking) => {
    if (!booking.service) return
    const employees = await api.matchEmployees(booking.service.id)
    setMatches((prev) => ({ ...prev, [booking.id]: employees }))
  }

  const handleConfirm = async (booking: Booking, transferToProjects: boolean) => {
    const priceValue = priceDrafts[booking.id]
    const addressValue = addressDrafts[booking.id]
    const updated = await api.confirmBooking(booking.id, {
      confirmedByName: 'الإداري',
      transferToProjects,
      quotedPrice: priceValue ? Number(priceValue) : undefined,
      address: addressValue || undefined,
    })
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    if (!transferToProjects) loadMatches(updated)
  }

  const handleDetailsBlur = async (
    booking: Booking,
    field: 'quotedPrice' | 'address',
    value: string,
  ) => {
    if (field === 'quotedPrice') {
      const num = value === '' ? null : Number(value)
      if (num === booking.quotedPrice) return
      const updated = await api.updateBookingDetails(booking.id, { quotedPrice: num })
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    } else {
      if (value === (booking.address || '')) return
      const updated = await api.updateBookingDetails(booking.id, { address: value })
      setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
    }
  }

  const handleAssign = async (
    booking: Booking,
    role: 'TECH_1' | 'TECH_2' | 'TECH_3',
    employeeId: string,
  ) => {
    if (!employeeId) return
    const updated = await api.assignTechnician(booking.id, { employeeId, role })
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  const handleVehicleChange = async (booking: Booking, assignedVehicle: string) => {
    if (assignedVehicle === (booking.assignedVehicle || '')) return
    const updated = await api.updateBookingDetails(booking.id, { assignedVehicle })
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  const pendingBookings = bookings.filter((b) => b.status === 'PENDING')
  const confirmedBookings = bookings.filter((b) => b.status === 'CONFIRMED')

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">تنسيق الحجوزات (الإداري)</h2>
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
          {/* بانتظار التثبيت */}
          <h3 className="mt-6 mb-3 text-lg font-bold text-brand-800">
            بانتظار التثبيت ({pendingBookings.length})
          </h3>
          <div className="flex flex-col gap-4">
            {pendingBookings.map((booking) => (
              <div
                key={booking.id}
                className="rounded-xl border border-amber-200 bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
              >
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <div>
                    <span className="font-mono text-sm font-semibold text-brand-600">
                      {booking.code}
                    </span>
                    <span className="mr-3 rounded-full bg-amber-100 px-3 py-1 text-xs font-bold text-amber-700">
                      بانتظار التثبيت
                    </span>
                    {booking.priority === 'URGENT' && (
                      <span className="mr-2 rounded-full bg-red-100 px-3 py-1 text-xs font-bold text-red-700">
                        عاجل
                      </span>
                    )}
                  </div>
                  <div className="text-sm text-slate-500">
                    {booking.service?.name || 'بدون خدمة محددة'}
                  </div>
                </div>

                <div className="mt-3 grid grid-cols-1 gap-2 text-sm sm:grid-cols-3">
                  <div>
                    <span className="text-slate-400">الزبون: </span>
                    <span className="font-medium text-brand-800">{booking.customer.name}</span>
                  </div>
                  <div>
                    <span className="text-slate-400">الهاتف: </span>
                    {booking.customer.phone}
                  </div>
                  <div>
                    <span className="text-slate-400">الموقع المسجل: </span>
                    {booking.customer.location || '-'}
                  </div>
                </div>
                {booking.notes && (
                  <p className="mt-2 text-sm text-slate-500">
                    <span className="text-slate-400">ملاحظات الزبون: </span>
                    {booking.notes}
                  </p>
                )}

                <div className="mt-4 rounded-lg bg-slate-50 p-4">
                  <h4 className="text-sm font-bold text-brand-800">
                    بعد الاتفاق مع الزبون، حدد التفاصيل قبل التثبيت
                  </h4>
                  <div className="mt-2 grid grid-cols-1 gap-3 sm:grid-cols-2">
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">
                        التكلفة المقدرة (اختياري)
                      </label>
                      <input
                        type="number"
                        placeholder="مثال: 150000"
                        value={priceDrafts[booking.id] || ''}
                        onChange={(e) =>
                          setPriceDrafts((prev) => ({ ...prev, [booking.id]: e.target.value }))
                        }
                        className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                      />
                    </div>
                    <div>
                      <label className="mb-1 block text-sm font-medium text-slate-600">
                        عنوان تنفيذ المهمة (اختياري)
                      </label>
                      <input
                        placeholder="عنوان مفصل لموقع تنفيذ المهمة"
                        value={addressDrafts[booking.id] || ''}
                        onChange={(e) =>
                          setAddressDrafts((prev) => ({ ...prev, [booking.id]: e.target.value }))
                        }
                        className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                      />
                    </div>
                  </div>
                </div>

                <div className="mt-4 flex flex-wrap gap-2">
                  <button
                    onClick={() => handleConfirm(booking, false)}
                    className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg"
                  >
                    تثبيت وترحيل لكادر الشد
                  </button>
                  <button
                    onClick={() => handleConfirm(booking, true)}
                    className="rounded-lg border border-brand-300 px-4 py-2 text-sm font-medium text-brand-700 transition-colors hover:bg-brand-50"
                  >
                    تثبيت وتحويل لإدارة المشاريع
                  </button>
                </div>
              </div>
            ))}
            {pendingBookings.length === 0 && (
              <p className="text-slate-400">لا توجد حجوزات بانتظار التثبيت.</p>
            )}
          </div>

          {/* تم تثبيتها */}
          <h3 className="mt-8 mb-3 text-lg font-bold text-brand-800">
            تم تثبيتها ({confirmedBookings.length})
          </h3>
          <div className="flex flex-col gap-4">
            {confirmedBookings.map((booking) => (
              <div
                key={booking.id}
                className="rounded-xl border border-emerald-200 bg-white p-6 shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
              >
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
                  </div>
                  <div className="text-sm text-slate-500">
                    {booking.service?.name || 'بدون خدمة محددة'}
                  </div>
                </div>

                <div className="mt-3 grid grid-cols-1 gap-2 text-sm sm:grid-cols-3">
                  <div>
                    <span className="text-slate-400">الزبون: </span>
                    <span className="font-medium text-brand-800">{booking.customer.name}</span>
                  </div>
                  <div>
                    <span className="text-slate-400">الهاتف: </span>
                    {booking.customer.phone}
                  </div>
                  <div>
                    <span className="text-slate-400">الموقع المسجل: </span>
                    {booking.customer.location || '-'}
                  </div>
                </div>

                <div className="mt-3 grid grid-cols-1 gap-3 rounded-lg bg-slate-50 p-4 sm:grid-cols-2">
                  <div>
                    <label className="mb-1 block text-sm font-medium text-slate-600">
                      التكلفة المقدرة
                    </label>
                    <input
                      type="number"
                      placeholder="غير محددة"
                      defaultValue={booking.quotedPrice ?? ''}
                      onBlur={(e) => handleDetailsBlur(booking, 'quotedPrice', e.target.value)}
                      className="w-full rounded-lg border border-slate-300 bg-white px-3 py-2 text-sm outline-none focus:border-brand-500"
                    />
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
                </div>

                {booking.transferToProjects ? (
                  <p className="mt-4 rounded-lg bg-brand-50 px-4 py-2 text-sm text-brand-700">
                    تم تحويل هذا الطلب إلى إدارة المشاريع.
                  </p>
                ) : (
                  <div className="mt-4">
                    <h4 className="text-sm font-bold text-brand-800">توجيه كادر الشد</h4>
                    <p className="mt-1 text-xs text-slate-400">
                      تحديد الكادر والسيارة اختياري - يمكن تثبيت الحجز وتحديدهم لاحقاً.
                    </p>

                    <div className="mt-2 sm:w-1/3">
                      <label className="mb-1 block text-sm font-medium text-slate-600">
                        المشرف المرافق (اختياري)
                      </label>
                      <select
                        value={booking.projectSupervisor?.id || ''}
                        onChange={(e) => handleSupervisorChange(booking, e.target.value)}
                        className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                      >
                        <option value="">-- بدون مشرف --</option>
                        {supervisors.map((s) => (
                          <option key={s.id} value={s.id}>
                            {s.name}
                          </option>
                        ))}
                      </select>
                    </div>

                    {matches[booking.id] && (
                      <div className="mt-2 grid grid-cols-1 gap-3 sm:grid-cols-3">
                        {techRoles.map((tr) => {
                          const assigned = booking.assignments.find((a) => a.role === tr.key)
                          const candidates = matches[booking.id]
                          return (
                            <div key={tr.key}>
                              <label className="mb-1 block text-sm font-medium text-slate-600">
                                {tr.label} (اختياري)
                              </label>
                              <select
                                value={assigned?.employee.id || ''}
                                onChange={(e) => handleAssign(booking, tr.key, e.target.value)}
                                className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                              >
                                <option value="">-- اختر فني --</option>
                                {candidates.length === 0 && (
                                  <option value="" disabled>
                                    لا يوجد موظف متاح يمتلك هذه المهارة
                                  </option>
                                )}
                                {candidates.map((c) => (
                                  <option key={c.id} value={c.id}>
                                    {c.name} ({c.position || 'فني'})
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
                      <input
                        defaultValue={booking.assignedVehicle || ''}
                        onBlur={(e) => handleVehicleChange(booking, e.target.value)}
                        placeholder="مثال: تويوتا هايلوكس - أبيض - 12345"
                        className="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                      />
                    </div>
                    {matches[booking.id]?.length === 0 && (
                      <p className="mt-2 text-sm text-red-600">
                        لا يوجد حالياً أي موظف بالدوام يمتلك مهارة "{booking.service?.name}". يرجى
                        التواصل مع إدارة الكوادر.
                      </p>
                    )}
                  </div>
                )}
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
