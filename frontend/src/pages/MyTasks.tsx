import { useEffect, useState } from 'react'
import { api, type Booking } from '../api'
import { useSession } from '../session'

export default function MyTasks() {
  const { employee } = useSession()
  const [bookings, setBookings] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [amounts, setAmounts] = useState<Record<string, string>>({})
  const [notes, setNotes] = useState<Record<string, string>>({})

  const load = () => {
    api
      .getBookings({ status: 'CONFIRMED' })
      .then(setBookings)
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const myTasks = bookings.filter((b) =>
    b.assignments.some((a) => a.employee.id === employee?.id),
  )

  const handleComplete = async (booking: Booking) => {
    const amountCollected = amounts[booking.id] ? Number(amounts[booking.id]) : undefined
    await api.completeBooking(booking.id, {
      completionNotes: notes[booking.id] || undefined,
      amountCollected,
    })
    setBookings((prev) => prev.filter((b) => b.id !== booking.id))
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
                    <div className="mt-1 grid grid-cols-1 gap-1 text-sm text-slate-500 sm:grid-cols-2">
                      <p>
                        <span className="text-slate-400">الزبون: </span>
                        {b.customer.name}
                      </p>
                      <p>
                        <span className="text-slate-400">الهاتف: </span>
                        {b.customer.phone}
                      </p>
                      <p>
                        <span className="text-slate-400">الموقع: </span>
                        {b.customer.location || 'بدون موقع محدد'}
                      </p>
                      <p>
                        <span className="text-slate-400">السيارة: </span>
                        {b.assignedVehicle || 'لم تحدد'}
                      </p>
                    </div>
                    {b.notes && (
                      <p className="mt-1 text-sm text-slate-500">
                        <span className="text-slate-400">ملاحظات: </span>
                        {b.notes}
                      </p>
                    )}

                    <div className="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-3">
                      <input
                        type="number"
                        placeholder="المبلغ المستلم"
                        value={amounts[b.id] || ''}
                        onChange={(e) => setAmounts((prev) => ({ ...prev, [b.id]: e.target.value }))}
                        className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500"
                      />
                      <input
                        placeholder="ملاحظات الإنجاز"
                        value={notes[b.id] || ''}
                        onChange={(e) => setNotes((prev) => ({ ...prev, [b.id]: e.target.value }))}
                        className="rounded-lg border border-slate-300 px-3 py-2 text-sm outline-none focus:border-brand-500 sm:col-span-1"
                      />
                      <button
                        onClick={() => handleComplete(b)}
                        className="rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg"
                      >
                        تم الإنجاز
                      </button>
                    </div>
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
    </div>
  )
}
