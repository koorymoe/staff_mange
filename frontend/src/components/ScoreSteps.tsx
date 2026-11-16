import type { StaffScore } from '../apiTypes'

// ═══ خطوات التقييم النهائي بنفس ترتيب الخادم (دازاي 11-16) ═══
// «٧٨٪ بالحسبة وطالع ٤٢٪» — الحسبة چانت تعرض أول سطر بس. الخادم يحسب بهالترتيب
// (matrix_score_service.compute): ماتركس والبشر ← + الشغل الإضافي ← × تغطية الراتب ← الانضباط.
const r = (n: number) => Math.round(n)

export default function ScoreSteps({ d }: { d: StaffScore }) {
  const m = d.matrixPct, h = d.humanPct
  const base = m != null && h != null ? 0.6 * m + 0.4 * h : m ?? h ?? 0
  const afterExtra = Math.min(100, base + (d.extraBonus ?? 0))
  const afterRevenue = afterExtra + (d.revenueAdj ?? 0)
  const ds = d.discipline?.score ?? null
  return (
    <ol className="list-decimal space-y-1 pr-4">
      <li>{m != null && h != null
        ? <>ماتركس {r(m)}% ({d.earned} من {d.max} نقطة) × ٦٠٪ + تقييم البشر {r(h)}% (معدّل {d.humanAvg?.toFixed(1)} من ٥ بـ{d.humanCount} تقييم) × ٤٠٪ = <b>{r(base)}%</b></>
        : m != null ? <>ماتركس بس: {r(m)}% ({d.earned} من {d.max} نقطة){d.aiOnly ? ' — تقييمه من ماتركس بس' : ' — ماكو تقييم بشر هالشهر'} = <b>{r(base)}%</b></>
        : <>تقييم البشر بس: {r(h ?? 0)}% = <b>{r(base)}%</b></>}</li>
      {!!d.extraBonus && <li>+ {r(d.extraBonus)} نقطة شغل إضافي برّا دوره ({d.extraJobs} شغلة، نقطة لكل وحدة لحد ٥) = <b>{r(afterExtra)}%</b></li>}
      {!!d.revenueAdj && <li>{d.revenueAdj > 0
        ? <>+ {r(d.revenueAdj)} لأن دخّل للشركة أكثر من راتبه (تغطية {r(d.salaryCoverage ?? 0)}٪)</>
        : <>× تغطية الراتب: دخّل للشركة {r(d.salaryCoverage ?? 0)}٪ من راتبه، فالرقم ينزل بنسبة النقص (لحد النص) → − {r(Math.abs(d.revenueAdj))}</>} = <b>{r(afterRevenue)}%</b></li>}
      {ds != null && <li>الانضباط بالدوام: {r(afterRevenue)}% × ٨٥٪ + انضباطه {r(ds)}% × ١٥٪ = <b>{r(d.final ?? 0)}%</b></li>}
    </ol>
  )
}
