import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api, type MatrixSuggestAccuracy } from '../api'
import MatrixNote from './MatrixNote'
import PctWhy from '../ui/PctWhy'
import { swallow } from '../utils/swallow'

// ═══ دقة اقتراحات ماتركس — المرحلة الثانية ═══
// كم مرة المنسقين مشوا على اقتراحه، وهل الحجوزات الي مشت عليه طلعت بوقتها
// أكثر من الي انغيّرت. هذا المقياس الي على أساسه (ع) يقرر يوم ينطي ماتركس
// صلاحية تنفيذ — مو قبل.
export default function MatrixSuggestAccuracyPanel() {
  const [a, setA] = useState<MatrixSuggestAccuracy | null>(null)
  const [days, setDays] = useState(30)
  useEffect(() => {
    let alive = true
    void api.getMatrixSuggestAccuracy(days).then((r) => { if (alive) setA(r) }).catch(swallow)
    return () => { alive = false }
  }, [days])
  if (!a) return null
  return (
    <div dir="rtl" className="space-y-3 rounded-2xl border border-violet-200 bg-white p-3 text-slate-800">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-sm font-extrabold text-violet-900">🎯 دقة اقتراحات ماتركس <span className="text-[11px] font-normal text-slate-500">ماتركس يقترح الموعد والكادر، والمنسق يقرر</span></p>
        <select value={days} onChange={(e) => setDays(Number(e.target.value))} className="rounded-lg border border-slate-200 px-2 py-1 text-xs">
          {[7, 30, 90].map((d) => <option key={d} value={d}>آخر {d} يوم</option>)}
        </select>
      </div>
      <MatrixNote>{a.insights.join(' ')}{a.pending ? ` و${a.pending} اقتراح ينتظر قرار المنسق.` : ''}</MatrixNote>
      <div className="grid gap-2 sm:grid-cols-2">
        {a.kinds.map((k) => (
          <div key={k.kind} className="rounded-xl border border-slate-200 bg-slate-50 p-3 text-xs">
            <b className="text-sm">{k.title}</b>
            <p className="mt-1 text-2xl font-black text-violet-700">{k.decided ? <PctWhy title={`قبول ${k.title}`} value={k.acceptPct} parts={{ numLabel: 'اقتراح اعتمده المنسق كامل', num: k.accepted, denLabel: 'اقتراح انحسم', den: k.decided }}
              meaning={<>شكد مرة المنسق مشى على {k.title} بدون تغيير بآخر {a.days} يوم. «اعتمد جزء منه» ({k.partial}) و«غيّره» ({k.changed}) ما ينحسبن قبول. كل ما تعلى يعني اقتراحاته صايبة أكثر.</>}
              source={<>اقتراحات ماتركس المسجّلة على الحجوزات وقرار المنسق عليها، بآخر {a.days} يوم.</>} /> : <span className="text-sm font-bold text-slate-400">بعد ماكو قرارات</span>}</p>
            <p>✅ اعتمده {k.accepted} · 🟡 اعتمد جزء منه {k.partial} · ✏️ غيّره {k.changed}</p>
            {(k.onTimeAccepted != null || k.onTimeChanged != null) && (
              <p className="mt-1 text-slate-600">طلع بوقته: لمن اعتمدوه <PctWhy title="طلع بوقته لمن اعتمدوا الاقتراح" value={k.onTimeAccepted ?? null}
                formula={<>من {k.sampleAccepted} حجز اعتمدوا بيه {k.title} وعنده وقت بداية مسجّل، الي بدا خلال ٣٠ دقيقة من موعده ÷ {k.sampleAccepted} = <b>{k.onTimeAccepted ?? '—'}%</b>.</>}
                meaning={<>إذا هاي أعلى من نسبة «لمن انغيّر» ({k.onTimeChanged ?? '—'}%) يعني المشي على اقتراح ماتركس يطلّع الشغل بوقته أكثر.</>}
                source={<>الحجوزات الي انحسم اقتراحها «اعتمده» بآخر {a.days} يوم: موعدها المحدد ووقت ما بدا الكادر الشغل فعلاً.</>} /> ({k.sampleAccepted}) · لمن انغيّر <PctWhy title="طلع بوقته لمن انغيّر الاقتراح" value={k.onTimeChanged ?? null}
                formula={<>من {k.sampleChanged} حجز غيّروا بيه {k.title} وعنده وقت بداية مسجّل، الي بدا خلال ٣٠ دقيقة من موعده ÷ {k.sampleChanged} = <b>{k.onTimeChanged ?? '—'}%</b>.</>}
                meaning={<>نقارنها ويا «لمن اعتمدوه» ({k.onTimeAccepted ?? '—'}%): إذا هاي أعلى، يعني تغيير المنسق چان أحسن من اقتراح ماتركس.</>}
                source={<>الحجوزات الي انحسم اقتراحها «غيّره» بآخر {a.days} يوم: موعدها ووقت بداية الشغل الفعلي.</>} /> ({k.sampleChanged})</p>
            )}
          </div>
        ))}
      </div>
      {a.recent.length > 0 && (
        <div className="text-xs">
          <b>آخر القرارات:</b>
          {a.recent.map((r, i) => <p key={i}>{r.outcome} — <Link to={`/bookings?focus=${r.bookingId}`} className="font-bold text-sky-700 underline">{r.code}</Link> · {r.kind}</p>)}
        </div>
      )}
    </div>
  )
}
