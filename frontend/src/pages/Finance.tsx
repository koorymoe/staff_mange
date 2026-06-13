import { useEffect, useState } from 'react'
import { api, type Booking } from '../api'

export default function Finance() {
  const [bookings, setBookings] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const load = () => {
    setLoading(true)
    api
      .getBookings('COMPLETED')
      .then(setBookings)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }

  useEffect(load, [])

  const handleVerify = async (booking: Booking) => {
    const updated = await api.verifyAmount(booking.id)
    setBookings((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))
  }

  return (
    <div>
      <h2 className="text-2xl font-bold text-brand-900">تدقيق الحسابات</h2>
      <p className="mt-1 text-slate-500">مراجعة المبالغ المستلمة من الحجوزات المنجزة وتدقيقها.</p>

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && (
        <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>
      )}

      <div className="mt-6 flex flex-col gap-4">
        {bookings.map((b) => (
          <div
            key={b.id}
            className="rounded-xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
          >
            <div className="flex flex-wrap items-center justify-between gap-2">
              <div>
                <span className="font-mono text-sm font-semibold text-brand-600">{b.code}</span>
                <span className="mr-3 text-sm font-medium text-brand-800">{b.customer.name}</span>
              </div>
              <span
                className={`rounded-full px-3 py-1 text-xs font-bold ${
                  b.amountVerified
                    ? 'bg-emerald-100 text-emerald-700'
                    : 'bg-amber-100 text-amber-700'
                }`}
              >
                {b.amountVerified ? 'تم التدقيق' : 'بانتظار التدقيق'}
              </span>
            </div>
            <p className="mt-2 text-sm text-slate-500">
              <span className="text-slate-400">المبلغ المستلم: </span>
              {b.amountCollected != null ? b.amountCollected.toLocaleString() : 'غير مسجل'}
            </p>
            {b.completionNotes && (
              <p className="text-sm text-slate-500">
                <span className="text-slate-400">ملاحظات الفني: </span>
                {b.completionNotes}
              </p>
            )}
            {!b.amountVerified && (
              <button
                onClick={() => handleVerify(b)}
                className="mt-3 rounded-lg bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2 text-sm font-medium text-white shadow-md shadow-brand-900/20 transition-all hover:shadow-lg"
              >
                تأكيد التدقيق
              </button>
            )}
          </div>
        ))}
        {!loading && bookings.length === 0 && (
          <p className="text-slate-400">لا توجد حجوزات منجزة بعد.</p>
        )}
      </div>
    </div>
  )
}
