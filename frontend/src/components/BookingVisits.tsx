import { useEffect, useState } from 'react'
import { api, type BookingVisit } from '../api'

// ═══ طلعات الحجز ═══
//
// «الحجز الي يكتمل بشكل جزئي… وين ما أوصل بيه، أريد أحسب إنتاجية
// الموظف. حتى لو الحجز نفسه طلعناله أربع أيام، كل مرة طلعناله
// تنحسب حجز للموظف، وكل مرة ينكتب بيها تاريخ وكادر طلع — لأن يجوز
// الكادر يتغيّر. المشكلة الي تصير هسه إن الطلعة الأولى تختفي ويُحسب
// بس الطلعة الثانية».
//
// هنا تنعرض الطلعات كلهن: منو طلع بكل وحدة، ومتى، وشكد أنجز.
// الطلعة الي راحت **ما تنمحي** لمن يتبدّل الكادر — هي واقعة صارت.

const ROLE_LABELS: Record<string, string> = {
  TECH_1: 'الفني الأول',
  TECH_2: 'الفني الثاني',
  TECH_3: 'الفني الثالث',
}

export default function BookingVisits({ bookingId }: { bookingId: string }) {
  const [visits, setVisits] = useState<BookingVisit[]>([])
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let alive = true
    api.getBookingVisits(bookingId)
      .then((v) => { if (alive) setVisits(v) })
      .catch(() => { if (alive) setVisits([]) })
      .finally(() => { if (alive) setLoading(false) })
    return () => { alive = false }
  }, [bookingId])

  if (loading) return <p className="text-xs text-slate-400">جاري تحميل الطلعات...</p>
  // ⚠️ ما نعرض صندوق فاضي: حجز ما طلعوا عليه بعد ما إله شي يتعرض.
  if (visits.length === 0) return null

  return (
    <div className="rounded-xl border border-slate-200 bg-white p-3">
      <p className="mb-2.5 text-[11px] font-extrabold text-[#0f2040]">
        🚚 الطلعات على هذا الحجز
        <span className="mr-1.5 rounded-full bg-brand-50 px-2 py-0.5 text-[10px] font-black text-brand-700">
          {visits.length}
        </span>
        <span className="mr-1.5 font-normal text-slate-400">— كل طلعة تنحسب لكادرها</span>
      </p>

      <div className="space-y-2">
        {visits.map((v) => (
          <div
            key={v.id}
            className={`rounded-lg border px-3 py-2 ${
              v.outcome === 'DONE' ? 'border-emerald-200 bg-emerald-50/60' : 'border-amber-200 bg-amber-50/50'
            }`}
          >
            <div className="flex flex-wrap items-center gap-x-3 gap-y-1">
              <span className="rounded-full bg-white px-2 py-0.5 text-[10px] font-black text-slate-600">
                طلعة {v.visitNumber}
              </span>
              <span className={`text-[11px] font-bold ${v.outcome === 'DONE' ? 'text-emerald-700' : 'text-amber-700'}`}>
                {v.outcome === 'DONE' ? '🏁 خلّصت الحجز' : `⏳ إنجاز جزئي${v.percentDone != null ? ` — ${v.percentDone}%` : ''}`}
              </span>
              <span className="text-[11px] text-slate-500">
                📅 {new Date(v.occurredAt).toLocaleString('ar-IQ', {
                  day: 'numeric', month: 'short', year: 'numeric', hour: '2-digit', minute: '2-digit',
                })}
              </span>
            </div>

            {/* الكادر بالمعرّفات — هذا الي تنعدّ منه الإنتاجية */}
            {v.crew.length > 0 ? (
              <div className="mt-1.5 flex flex-wrap gap-1.5">
                {v.crew.map((c) => (
                  <span
                    key={c.employeeId}
                    className="rounded-lg bg-white px-2 py-0.5 text-[11px] font-bold text-slate-700 shadow-sm"
                  >
                    {c.isLeader && '👑 '}{c.name}
                    <span className="font-normal text-slate-400"> · {ROLE_LABELS[c.role] || c.role}</span>
                  </span>
                ))}
              </div>
            ) : (
              // ⚠️ ما ننكتها بصمت: طلعة بلا كادر مسجّل تعني إن أحداً
              // شغّل الحجز بلا تكليف — وهاي معلومة تنراد تنشاف.
              <p className="mt-1.5 text-[11px] font-bold text-red-600">⚠️ ماكو كادر مسجّل بهاي الطلعة</p>
            )}
          </div>
        ))}
      </div>
    </div>
  )
}
