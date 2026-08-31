import { useEffect, useState } from 'react'
import { api, type Booking } from '../api'

// صفحة تدقيق للمراقب (صلاحية crew_management) — تعرض الحجوزات الموجّهة/المسندة
// من موظف مبيعات (أو غيره) لكن لسه ما ثبّتها الإداري (حالة PENDING). الهدف:
// يقدر المراقب يقارن ويدقق هل الإداري تواصل فعلاً مع الزبون وأقفل الاتفاق قبل
// التثبيت (حقل confirmationContactedAt) أو لسه ما تواصل، من مصدر مستقل عن
// المسار العام /api/bookings.
/** ⚠️ `embedded`: نفس الشاشة بالضبط بلا ترويستها — تنضمّ بمكتب
 *  المراقب. **ما ننسخ المحتوى**: نسختان تفترقان بأول تصحيح، فالمراقب
 *  يشوف صفاً بشاشة ومحلولاً بالثانية ويفقد الثقة بالاثنتين. */
interface EmbeddedProps { embedded?: boolean }

export default function MonitorCrewBookingsPage({ embedded }: EmbeddedProps = {}) {
  const [bookings, setBookings] = useState<Booking[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api.getPendingAudit()
      .then(setBookings)
      .catch((e) => setError(e instanceof Error ? e.message : 'حدث خطأ'))
      .finally(() => setLoading(false))
  }, [])

  return (
    <div>
      {!embedded && (
        <>
          <h2 className="text-2xl font-bold text-brand-900">تدقيق تنسيق الحجوزات</h2>
          <p className="mt-1 text-slate-500">
            الحجوزات الموجّهة من المبيعات (أو غيرهم) وما زالت بانتظار تثبيت الإداري — لمقارنة
            هل الإداري تواصل فعلاً مع الزبون وأقفل الاتفاق قبل التثبيت.
          </p>
        </>
      )}

      {loading && <p className="mt-6 text-slate-400">جاري التحميل...</p>}
      {error && <p className="mt-6 rounded-lg bg-red-50 p-4 text-red-600">تعذر الاتصال بالخادم: {error}</p>}

      {!loading && !error && (
        <div className="mt-6 flex flex-col gap-3">
          {bookings.map((b) => (
            <div
              key={b.id}
              className="rounded-xl border border-amber-200 bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]"
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <span className="font-mono text-sm font-semibold text-brand-600">{b.code}</span>
                <span className="rounded-full bg-amber-100 px-3 py-1 text-xs font-bold text-amber-700">
                  بانتظار التثبيت
                </span>
              </div>
              <div className="mt-2 grid grid-cols-1 gap-1 text-sm text-slate-500 sm:grid-cols-3">
                <p><span className="text-slate-400">الزبون: </span>{b.customer?.name || 'زبون غير معروف'}</p>
                <p><span className="text-slate-400">الهاتف: </span>{b.customer?.phone || '-'}</p>
                <p>
                  <span className="text-slate-400">وجّهه: </span>
                  {b.transferEmployee?.name || 'غير محدد'}
                </p>
              </div>
              <div className="mt-3">
                {b.confirmationContactedAt ? (
                  <div className="rounded-lg bg-emerald-50 px-3 py-2 text-xs font-bold text-emerald-700">
                    ✅ الإداري تواصل مع الزبون وأقفل الاتفاق ({b.confirmationContactedBy?.name || 'غير معروف'}) —{' '}
                    {new Date(b.confirmationContactedAt).toLocaleString('ar-IQ')}
                  </div>
                ) : (
                  <div className="rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-700">
                    ⏳ الإداري لسه ما تواصل مع الزبون (ما ضغط "تم")
                  </div>
                )}
              </div>
            </div>
          ))}
          {bookings.length === 0 && (
            <p className="text-slate-400">لا توجد حجوزات بانتظار التثبيت حالياً.</p>
          )}
        </div>
      )}
    </div>
  )
}
