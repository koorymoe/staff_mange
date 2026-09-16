import { useEffect, useState } from 'react'
import { api, type TechnicianKpi } from '../api'

// ═══ مخطط نقاط الكي بي اي — من وين طلعت نقاط الموظف ═══
//
// الرقم الإجمالي لوحده ما يفيد: الموظف يشوف «٨٧ نقطة» وما يعرف شنو
// يسوي حتى يخليها ١٠٠، والمدير يشوف «٤٥» وما يعرف المشكلة وين.
//
// المخطط يفصّلها للستة عناصر الي تتكوّن منها فعلاً، فينبيّن بنظرة
// وحدة: العمود الواطي هو المشكلة.
//
// ملاحظة: أكو عنصر يطلع بالسالب (الشكاوى والخصومات اليدوية). نرسمها
// بالأحمر تحت الخط، مو نخفيها — العمود السالب هو أهم شي بالمخطط.
export default function KpiBreakdownChart({
  employeeId,
  employeeName,
  month,
  onClose,
}: {
  employeeId: string
  employeeName: string
  month?: string
  onClose: () => void
}) {
  const [data, setData] = useState<TechnicianKpi | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    api.getTechnicianKpi(employeeId, month)
      .then(setData)
      .catch((e) => setError(e instanceof Error ? e.message : 'تعذر جلب النقاط'))
      .finally(() => setLoading(false))
  }, [employeeId, month])

  const bars = data
    ? [
        { label: 'الحجوزات المكتملة', points: data.breakdown.completedBookings.points, detail: `${data.breakdown.completedBookings.count} حجز` },
        { label: 'سرعة الإنجاز', points: data.breakdown.completionSpeed.points, detail: data.breakdown.completionSpeed.avgMinutes ? `متوسط ${Math.round(data.breakdown.completionSpeed.avgMinutes)} دقيقة` : 'ماكو بيانات' },
        { label: 'تقارير العمل', points: data.breakdown.workReports.points, detail: `${data.breakdown.workReports.fullReports} كامل من ${data.breakdown.workReports.count}` },
        { label: 'الحضور', points: data.breakdown.attendance.points, detail: `${data.breakdown.attendance.daysPresent} من ${data.breakdown.attendance.totalDays} يوم` },
        { label: 'شكاوى الزبائن', points: data.breakdown.complaints.points, detail: data.breakdown.complaints.count > 0 ? `${data.breakdown.complaints.count} شكوى` : 'ماكو شكاوى' },
        { label: 'خصومات يدوية', points: data.breakdown.manualDeductions.points, detail: `${data.breakdown.manualDeductions.count} خصم` },
      ]
    : []

  // السلّم يغطي الموجب والسالب سوه، وإلا العمود السالب ينقص من تحت
  const maxAbs = Math.max(20, ...bars.map((b) => Math.abs(b.points)))

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" onClick={onClose}>
      <div
        className="max-h-[90vh] w-full max-w-3xl overflow-y-auto rounded-2xl bg-white p-5 shadow-xl"
        onClick={(e) => e.stopPropagation()}
      >
        <div className="flex items-start justify-between gap-3">
          <div>
            <h3 className="text-lg font-extrabold text-[#0f2040]">📊 نقاط الكي بي اي — {employeeName}</h3>
            {data && <p className="text-xs text-slate-500">الفترة: {data.period}</p>}
          </div>
          {data && (
            <div className="shrink-0 rounded-xl bg-[#0f2040] px-4 py-2 text-center text-white">
              <p className="text-[11px] opacity-80">المجموع</p>
              <p className="text-2xl font-extrabold">{data.totalPoints}</p>
            </div>
          )}
        </div>

        {loading && <p className="mt-6 text-center text-slate-400">جاري التحميل...</p>}
        {error && <p className="mt-6 rounded-lg bg-red-50 p-3 text-sm font-bold text-red-700">{error}</p>}

        {data && (
          <>
            {/* الأعمدة */}
            <div className="mt-5 flex items-end justify-around gap-2 border-b border-slate-200 pb-1" style={{ height: '260px' }}>
              {bars.map((b) => {
                const h = (Math.abs(b.points) / maxAbs) * 100
                const negative = b.points < 0
                return (
                  <div key={b.label} className="flex h-full flex-1 flex-col items-center justify-end">
                    <span className={`mb-1 text-sm font-extrabold ${negative ? 'text-red-600' : 'text-[#0f2040]'}`}>
                      {b.points}
                    </span>
                    <div
                      className={`w-full max-w-[64px] rounded-t ${negative ? 'bg-red-500' : 'bg-[#6aa84f]'}`}
                      style={{ height: `${Math.max(h, 2)}%` }}
                      title={`${b.label}: ${b.points} نقطة — ${b.detail}`}
                    />
                  </div>
                )
              })}
            </div>

            {/* الأسماء تحت الأعمدة */}
            <div className="flex justify-around gap-2">
              {bars.map((b) => (
                <div key={b.label} className="flex-1 pt-2 text-center">
                  <p className="text-[11px] font-bold leading-tight text-slate-700">{b.label}</p>
                  <p className="text-[10px] text-slate-400">{b.detail}</p>
                </div>
              ))}
            </div>

            {/* الي يحتاج انتباه */}
            {bars.some((b) => b.points < 0) && (
              <p className="mt-4 rounded-lg bg-red-50 px-3 py-2 text-xs font-bold text-red-800">
                ⚠️ الأعمدة الحمر تنقص من المجموع — هذي الي تحتاج شغل عليها.
              </p>
            )}
          </>
        )}

        <button
          onClick={onClose}
          className="mt-5 w-full rounded-lg border border-slate-300 px-4 py-2 text-sm font-bold text-slate-600"
        >
          إغلاق
        </button>
      </div>
    </div>
  )
}
