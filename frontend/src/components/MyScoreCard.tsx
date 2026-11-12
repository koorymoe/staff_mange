import { useEffect, useState } from 'react'
import { api, type StaffScoreDetail } from '../api'
import { ScoreBreakdown } from '../pages/StaffScorePage'
import { swallow } from '../utils/swallow'
import PctWhy from '../ui/PctWhy'

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
          <span className="block text-[11px] text-slate-500">ماتركس ٦٠٪ + تقييم مسؤوليك ٤٠٪{d.reliability != null && <> · 🛡️ الاعتمادية <b><PctWhy title="الاعتمادية" value={Math.round(d.reliability)}
            formula={<>معدّل {d.reliabilityParts.length} جزء: ({d.reliabilityParts.map((p) => `${Math.round(p.pct)}`).join(' + ')}) ÷ {d.reliabilityParts.length} = <b>{Math.round(d.reliability)}%</b></>}
            meaning={<>شكد نگدر نعتمد عليك بالأشياء الأساسية: الحضور، الجرد، الورق، المهام، الشكاوى، الإجازات، المشاكل ويا الزملاء، العدّة والـKPI. كل جزء يبدي من ١٠٠ وينزل بقدر الي عليك.</>}
            source={<>سجلات هالشهر: الحضور، جرد العدّة، أوراق الحجوزات، المهام، شكاوى الزبائن، الإجازات، والملاحظات الوظيفية.</>}>
            <ul className="space-y-0.5 text-xs">{d.reliabilityParts.map((p) => <li key={p.key}>• {p.label}: <b>{Math.round(p.pct)}%</b> — {p.detail}</li>)}</ul>
          </PctWhy></b></>}{tip && <> · أكثر شي نزّله: <b>{tip.title}</b></>}</span>
        </span>
        <b className="text-2xl" style={{ color }}>{d.final == null ? '—' : <PctWhy title="تقييمي هالشهر" value={Math.round(d.final)}
          formula={(() => {
            const m = d.matrixPct, h = d.humanPct
            const base = m != null && h != null ? 0.6 * m + 0.4 * h : m ?? h ?? 0
            const ds = d.discipline?.score ?? null
            return <>
              <p>{m != null && h != null
                ? <>ماتركس {Math.round(m)}% ({d.earned} من {d.max} نقطة) × ٦٠٪ + تقييم مسؤوليك {Math.round(h)}% (معدّل {d.humanAvg?.toFixed(1)} من ٥ بـ{d.humanCount} تقييم) × ٤٠٪ = {Math.round(base)}%</>
                : m != null ? <>بس ماتركس: {Math.round(m)}% ({d.earned} من {d.max} نقطة) — لأن ماكو تقييم من مسؤوليك هالشهر.</>
                : <>بس تقييم مسؤوليك: {Math.round(h ?? 0)}% (معدّل {d.humanAvg?.toFixed(1)} من ٥) — لأن ماكو نقاط ماتركس هالشهر.</>}</p>
              {ds != null && <p>بعدين الانضباط بالدوام: {Math.round(base)}% × ٨٥٪ + انضباطك {Math.round(ds)}% × ١٥٪ = <b>{Math.round(d.final ?? 0)}%</b></p>}
            </>
          })()}
          meaning={<>رقمك هالشهر. ٨٠% فما فوق أخضر، ٦٠–٧٩ برتقالي، أقل أحمر. ترفعه لمن تكسب نقاط ماتركس (شغل بوقته، ورق كامل…)، ومسؤوليك يقيّموك زين، وتحضر بوقتك.{tip ? ` أكثر شي نزّله: ${tip.title}.` : ''}{d.prevFinal != null ? ` الشهر الماضي چان ${Math.round(d.prevFinal)}%.` : ''}</>}
          source={<>نقاط ماتركس المسجّلة عليك هالشهر من الحجوزات والحضور والمهام، وتقييمات مسؤوليك (من ١ لـ٥)، وسجل حضورك مقابل جدولك. التفصيل الكامل تفتحه بالضغط على البطاقة.</>} />}</b>
      </button>
      {tip?.advice && !open && <p className="mt-1 text-xs text-violet-800">💡 {tip.advice}</p>}
      {open && <div className="mt-3"><ScoreBreakdown d={d} /></div>}
    </div>
  )
}
