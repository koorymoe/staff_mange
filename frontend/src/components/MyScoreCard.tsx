import { useEffect, useState } from 'react'
import { api, type StaffScoreDetail } from '../api'
import { ScoreBreakdown } from '../pages/StaffScorePage'
import { swallow } from '../utils/swallow'
import PctWhy from '../ui/PctWhy'
import ScoreSteps from './ScoreSteps'

// «🏅 تقييمي هالشهر» — الموظف يشوف رقمه وليش، وشلون يرفعه. ما يشوف غيره.
export default function MyScoreCard() {
  const [d, setD] = useState<StaffScoreDetail | null>(null)
  const [open, setOpen] = useState(false)
  useEffect(() => { void api.getMyScore().then((r) => { if (r?.score && r.score.max + r.score.humanCount > 0) setD(r.score) }).catch(swallow) }, [])
  if (!d) return null
  const v = d.final ?? 0
  const color = v >= 80 ? '#059669' : v >= 60 ? '#d97706' : '#dc2626'
  const tip = d.topLosses[0]
  return (
    <div dir="rtl" className="rounded-2xl border border-slate-200 bg-white p-3 shadow-sm">
      <button type="button" onClick={() => setOpen(!open)} className="flex w-full items-center justify-between gap-3 text-right">
        <span>
          <b className="text-sm text-slate-800">🏅 تقييمي هالشهر</b>
          <span className="block text-[11px] text-slate-500">٦٠٪ من شغلك المسجّل + ٤٠٪ من تقييم مسؤولك{d.reliability != null && <> · 🛡️ الالتزام <b><PctWhy title="الالتزام" value={Math.round(d.reliability)}
            formula={<>معدّل {d.reliabilityParts.length} جزء: ({d.reliabilityParts.map((p) => `${Math.round(p.pct)}`).join(' + ')}) ÷ {d.reliabilityParts.length} = <b>{Math.round(d.reliability)}%</b></>}
            meaning={<>شكد نگدر نعتمد عليك بالأشياء الأساسية: الحضور، الجرد، الورق، المهام، الشكاوى، الإجازات، المشاكل ويا الزملاء، العدّة والـKPI. كل جزء يبدي من ١٠٠ وينزل بقدر الي عليك.</>}
            source={<>سجلات هالشهر: الحضور، جرد العدّة، أوراق الحجوزات، المهام، شكاوى الزبائن، الإجازات، والملاحظات الوظيفية.</>}>
            <ul className="space-y-0.5 text-xs">{d.reliabilityParts.map((p) => <li key={p.key}>• {p.label}: <b>{Math.round(p.pct)}%</b> — {p.detail}</li>)}</ul>
          </PctWhy></b></>}{tip && <> · أكثر شي نزّله: <b>{tip.title}</b></>}</span>
        </span>
        <b className="text-2xl" style={{ color }}>{d.final == null ? '—' : <PctWhy title="تقييمي هالشهر" value={Math.round(d.final)}
          formula={<ScoreSteps d={d} />}
          meaning={<>رقمك هالشهر. ٨٠% فما فوق أخضر، ٦٠–٧٩ برتقالي، أقل أحمر. ترفعه لمن تكسب نقاط ماتركس (شغل بوقته، ورق كامل…)، ومسؤوليك يقيّموك زين، وتحضر بوقتك، وشغلك يدخّل للشركة قد راتبك.{tip ? ` أكثر شي نزّله: ${tip.title}.` : ''}{d.prevFinal != null ? ` الشهر الماضي چان ${Math.round(d.prevFinal)}%.` : ''}{d.trend != null && Math.abs(d.trend) >= 2 ? (d.trend > 0 ? ` 📈 صاعد ${Math.round(d.trend)} عن معدّلك بآخر ٣ أشهر — كمّل!` : ` 📉 نازل ${Math.round(-d.trend)} عن معدّلك بآخر ٣ أشهر.`) : ''}</>}
          source={<>نقاط ماتركس المسجّلة عليك هالشهر من الحجوزات والحضور والمهام، وتقييمات مسؤوليك (من ١ لـ٥)، وسجل حضورك مقابل جدولك. التفصيل الكامل تفتحه بالضغط على البطاقة.</>} />}</b>
      </button>
      {tip?.advice && !open && <p className="mt-1 text-xs text-violet-800">💡 {tip.advice}</p>}
      {open && <div className="mt-3"><ScoreBreakdown d={d} mine /></div>}
    </div>
  )
}
