import type { StaffScore } from '../apiTypes'

// ═══ 🔁 قبل/بعد التقييم الجديد (دازاي 11-18) ═══
// لكل فني وليدر: رقمه بالحسبة القديمة (ماتركس ٦٠ + بشر ٤٠) وبالجديدة (البنود + رأي ماتركس)،
// قبل الفلوس والانضباط — حتى المالك يتأكد إن الجديدة منصفة. الأكبر فرق أول.
const r = (n: number) => Math.round(n)

export default function ScoreBeforeAfter({ staff }: { staff: StaffScore[] }) {
  const rows = staff
    .filter((s) => s.techParts?.length)
    .map((s) => {
      const now = s.techParts!.reduce((a, p) => a + p.pct * p.weight / 100, 0)
      return { s, now, old: s.legacyBase ?? null, diff: s.legacyBase != null ? now - s.legacyBase : null }
    })
    .sort((a, b) => Math.abs(b.diff ?? 0) - Math.abs(a.diff ?? 0))
  if (rows.length === 0) return <p className="py-6 text-center text-slate-400">ماكو فنيين أو ليدرية محسوبين بالحسبة الجديدة هالشهر.</p>
  const up = rows.filter((x) => (x.diff ?? 0) > 5).length, down = rows.filter((x) => (x.diff ?? 0) < -5).length
  return (
    <div dir="rtl" className="space-y-3">
      <p className="rounded-xl bg-sky-50 px-4 py-2 text-[12.5px] text-sky-900">
        الرقم قبل الفلوس والانضباط. <b>{up}</b> طلعوا أعلى بأكثر من ٥ نقاط، و<b>{down}</b> نزلوا. الفرق الكبير يستاهل تفتح تفاصيله وتشوف ليش.
      </p>
      <div className="overflow-x-auto rounded-2xl border border-slate-200 bg-white">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 text-xs text-slate-500">
            <tr><th className="p-2 text-right">الموظف</th><th className="p-2">المجموعة</th><th className="p-2">القديم</th><th className="p-2">الجديد</th><th className="p-2">الفرق</th><th className="p-2 text-right">ملاحظات ماتركس</th></tr>
          </thead>
          <tbody>
            {rows.map(({ s, now, old, diff }) => (
              <tr key={s.id} className="border-t border-slate-100">
                <td className="p-2 font-bold">{s.name}</td>
                <td className="p-2 text-center text-xs text-slate-500">{s.groupLabel}</td>
                <td className="p-2 text-center">{old != null ? `${r(old)}%` : '—'}</td>
                <td className="p-2 text-center font-bold">{r(now)}%</td>
                <td className={`p-2 text-center font-bold ${diff == null ? 'text-slate-400' : diff > 5 ? 'text-emerald-600' : diff < -5 ? 'text-red-600' : 'text-slate-500'}`}>
                  {diff == null ? 'جديد' : `${diff > 0 ? '+' : ''}${r(diff)}`}
                </td>
                <td className="p-2 text-xs text-amber-800">{s.matrixNotes?.join(' · ') || '—'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  )
}
