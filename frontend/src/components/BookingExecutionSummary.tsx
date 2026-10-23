import type { BookingExecutionDetail } from '../api'

// ═══ «شنو صار بهذا الحجز فعلاً» ═══
//
// مهندس الجودة كان يتصل بالزبون وهو ما يعرف منو طلع ولا شكد استغرقوا،
// فيسأل سؤال عام («شلون كان الشغل؟») ويكتب ملاحظة عامة. والزبون يجاوب
// جواب عام. فالمتابعة تصير إجراء شكلي.
//
// بهاي التفاصيل يقدر يسأل سؤال محدد: «الفريق وصلكم الساعة ٩ وخلّص
// ١١:٣٠، هل هذا صحيح؟» — وهنا يطلع الجواب الحقيقي.
export default function BookingExecutionSummary({ exec }: { exec: BookingExecutionDetail | null }) {
  if (!exec) return null

  const t = (iso: string | null) =>
    iso ? new Date(iso).toLocaleString('ar-IQ', { weekday: 'short', day: 'numeric', month: 'numeric', hour: '2-digit', minute: '2-digit' }) : '—'

  const duration = (m: number | null) => {
    if (m == null) return '—'
    const h = Math.floor(m / 60)
    const min = m % 60
    if (h === 0) return `${min} دقيقة`
    if (min === 0) return `${h} ساعة`
    return `${h} ساعة و${min} دقيقة`
  }

  const leader = exec.crew.find((c) => c.isLeader)

  return (
    <div className="mt-3 rounded-xl border border-slate-200 bg-slate-50/60 p-3">
      <p className="text-sm font-bold text-[#0f2040]">🔧 تفاصيل التنفيذ</p>

      {/* الكادر — والليدر مؤشّر لأنه هو الي تنزل عليه الغرامة */}
      <div className="mt-2">
        <span className="text-xs text-slate-500">الكادر الي طلع: </span>
        {exec.crew.length === 0 ? (
          <span className="text-xs text-slate-400">ما انسجّل كادر لهذا الحجز</span>
        ) : (
          <span className="inline-flex flex-wrap gap-1.5">
            {exec.crew.map((c) => (
              <span
                key={c.employeeId}
                className={`rounded-full px-2.5 py-0.5 text-xs font-bold ${
                  c.isLeader ? 'bg-amber-100 text-amber-900' : 'bg-white text-slate-700 border border-slate-200'
                }`}
              >
                {c.isLeader && '⭐ '}
                {c.name}
              </span>
            ))}
          </span>
        )}
      </div>
      {leader && (
        <p className="mt-1 text-[11px] text-amber-800">
          ⭐ {leader.name} هو الليدر المسؤول — التقرير السلبي ينخصم منه.
        </p>
      )}

      {/* الأوقات */}
      <div className="mt-3 grid grid-cols-1 gap-2 sm:grid-cols-3">
        <Cell label="بدأ العمل" value={t(exec.startedAt)} />
        <Cell label="تم الإنجاز" value={t(exec.completedAt)} />
        <Cell label="المدة الفعلية" value={duration(exec.durationMinutes)} strong />
      </div>

      {exec.workStoppedAt && (
        <p className="mt-2 rounded-lg bg-amber-50 px-2.5 py-1.5 text-xs text-amber-900">
          ⏸ توقف العمل {t(exec.workStoppedAt)}
          {exec.workStopReason ? ` — ${exec.workStopReason}` : ''}
        </p>
      )}

      {exec.completionNotes && (
        <div className="mt-2 rounded-lg border border-slate-200 bg-white p-2.5">
          <p className="text-xs font-bold text-slate-500">تقرير الإنجاز:</p>
          <p className="mt-0.5 text-sm text-slate-700">{exec.completionNotes}</p>
        </div>
      )}

      {/* الحجز الي أخذ أكثر من يوم — كل يوم بتقريره */}
      {exec.progressReports.length > 0 && (
        <div className="mt-2 rounded-lg border border-amber-200 bg-amber-50/70 p-2.5">
          <p className="text-xs font-bold text-amber-900">
            📋 هذا الحجز أخذ {exec.progressReports.length} أيام:
          </p>
          {exec.progressReports.map((r) => (
            <div key={r.id} className="mt-1.5 text-xs">
              <span className="font-bold text-amber-800">اليوم {r.dayNumber} ({r.percentDone}٪): </span>
              <span className="text-slate-700">{r.workDone}</span>
              {r.blockers && <span className="text-red-700"> — عرقلة: {r.blockers}</span>}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

function Cell({ label, value, strong }: { label: string; value: string; strong?: boolean }) {
  return (
    <div className="rounded-lg border border-slate-200 bg-white px-2.5 py-1.5">
      <p className="text-[11px] text-slate-500">{label}</p>
      <p className={`text-sm ${strong ? 'font-extrabold text-[#0f2040]' : 'font-bold text-slate-700'}`}>{value}</p>
    </div>
  )
}
