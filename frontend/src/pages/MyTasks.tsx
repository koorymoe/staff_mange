import { useEffect, useState } from 'react'
import { api, type Booking } from '../api'
import { useSession } from '../session'

export default function MyTasks() {
  const { employee } = useSession()
  const [bookings, setBookings] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api
      .getBookings('CONFIRMED')
      .then(setBookings)
      .finally(() => setLoading(false))
  }, [])

  const myTasks = bookings.filter((b) =>
    b.assignments.some((a) => a.employee.id === employee?.id),
  )

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
                    <p className="text-sm text-slate-500">
                      {b.customer.name} - {b.customer.location || 'بدون موقع محدد'}
                    </p>
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
                      {s.service.name}
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
