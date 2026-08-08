import { useEffect, useState } from 'react'
import { api, type BookingProgressReport } from '../api'

// ═══ «وين وصل الي قبلنا؟» ═══
//
// هذا الجزء هو الفايدة الحقيقية من الإنجاز الجزئي. بدونه الكادر الجاي
// يوصل الموقع ويبدي يسأل الزبون شنو انسوّى أمس — والزبون مو مسؤول
// يعرف، وأحياناً يجاوب غلط فينعاد شغل منجز أو ينتنسى شغل ناقص.
//
// نعرض كل يوم بترتيبه: منو اشتغل، شنو خلّص، شنو باقي، وشنو عرقلهم.
export default function BookingProgressTimeline({ bookingId }: { bookingId: string }) {
  const [rows, setRows] = useState<BookingProgressReport[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.getBookingProgress(bookingId)
      .then(setRows)
      .catch(() => setRows([]))
      .finally(() => setLoading(false))
  }, [bookingId])

  if (loading) return <p className="text-xs text-slate-400">جاري تحميل تقارير الأيام...</p>
  if (rows.length === 0) return null

  const last = rows[rows.length - 1]

  return (
    <div className="mt-3 rounded-xl border border-amber-200 bg-amber-50/60 p-3">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h4 className="text-sm font-bold text-amber-900">
          📋 وين وصل الشغل — {rows.length} {rows.length === 1 ? 'يوم' : 'أيام'}
        </h4>
        <span className="rounded-full bg-amber-200/70 px-2.5 py-0.5 text-xs font-bold text-amber-900">
          منجز {last.percentDone}٪
        </span>
      </div>

      {/* الباقي أول شي وبأوضح شكل — هذا الي يحتاجه الكادر الحين */}
      <div className="mt-2 rounded-lg border border-amber-300 bg-white p-2.5">
        <p className="text-xs font-bold text-slate-500">الباقي عليكم:</p>
        <p className="mt-0.5 text-sm font-bold text-[#0f2040]">{last.remainingWork}</p>
      </div>

      <div className="mt-3 space-y-2">
        {rows.map((r) => (
          <div key={r.id} className="rounded-lg border border-amber-100 bg-white/70 p-2.5 text-xs">
            <div className="flex flex-wrap items-center gap-2">
              <span className="rounded-full bg-amber-100 px-2 py-0.5 font-bold text-amber-800">
                اليوم {r.dayNumber}
              </span>
              <span className="text-slate-400">
                {new Date(r.createdAt).toLocaleDateString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'numeric' })}
              </span>
              <span className="font-bold text-slate-600">{r.percentDone}٪</span>
              {r.crewSnapshot && <span className="text-slate-500">👷 {r.crewSnapshot}</span>}
            </div>
            <p className="mt-1.5">
              <span className="font-bold text-emerald-700">انخلص: </span>
              <span className="text-slate-700">{r.workDone}</span>
            </p>
            <p className="mt-1">
              <span className="font-bold text-amber-700">باقي: </span>
              <span className="text-slate-700">{r.remainingWork}</span>
            </p>
            {r.blockers && (
              <p className="mt-1">
                <span className="font-bold text-red-600">عرقلة: </span>
                <span className="text-slate-700">{r.blockers}</span>
              </p>
            )}
            {r.materialsUsed && (
              <p className="mt-1 text-slate-500">📦 مواد: {r.materialsUsed}</p>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
