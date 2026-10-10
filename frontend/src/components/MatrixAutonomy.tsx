import { useEffect, useState } from 'react'
import { api, type MatrixAutonomyStatus } from '../api'
import MatrixNote from './MatrixNote'
import OwnerSwitch from './OwnerSwitch'
import { SWITCH_MATRIX_AUTO_CREW, SWITCH_MATRIX_MONITOR_ESCALATE } from '../systemSwitches'

// ═══ المرحلة الثالثة: «🤖 ماتركس ينفّذ» ═══
// كل فعل: مفتاح المالك (مطفي افتراضياً) + بوابة دقة + سجل + تراجع.
export default function MatrixAutonomy() {
  const [st, setSt] = useState<MatrixAutonomyStatus | null>(null)
  const [reload, setReload] = useState(0)
  const [busy, setBusy] = useState<string | null>(null)
  useEffect(() => {
    let alive = true
    void api.getMatrixAutonomy().then((r) => { if (alive) setSt(r) }).catch(() => {})
    return () => { alive = false }
  }, [reload])
  if (!st) return null
  const c = st.autoCrew

  const undo = async (id: string) => {
    setBusy(id)
    try { await api.undoMatrixAction(id); setReload((n) => n + 1) } finally { setBusy(null) }
  }

  return (
    <div dir="rtl" className="space-y-3 rounded-2xl border border-emerald-200 bg-white p-3 text-slate-800">
      <p className="text-sm font-extrabold text-emerald-900">🤖 ماتركس ينفّذ <span className="text-[11px] font-normal text-slate-500">أفعال بسيطة تنرجع بضغطة — كل وحدة بمفتاحك</span></p>
      {!st.autopilotOn && <MatrixNote>«ماتركس ينفّذ التذكيرات لحاله» مطفي — وهو المفتاح الأم، فكل الأفعال هنا واقفة.</MatrixNote>}

      <div className="rounded-xl border border-slate-200 p-3 text-xs">
        <b className="text-sm">👷 يكلّف كادر حجز باچر ما انحدد</b>
        <p className="mt-1 text-slate-600">الساعة ٦ المسا: حجز مثبّت باچر وبلا كادر ← ماتركس يكلّف اقتراحه الكامل ويبلّغ المنسق والكادر. المنسق يتراجع من «تنسيق الحجوزات».</p>
        <p className={`mt-2 font-bold ${c.eligible ? 'text-emerald-700' : 'text-amber-700'}`}>
          بوابة الدقة: {c.decided} اقتراح كادر انحسم بآخر ٣٠ يوم، قبول {c.decided ? `${c.pct}%` : '—'} —
          {c.eligible ? ' ✅ مؤهل' : ` بعده مو مؤهل (يحتاج ${c.needDecided} اقتراح وقبول ${c.needPct}%)`}
        </p>
        <div className="mt-2"><OwnerSwitch switchKey={SWITCH_MATRIX_AUTO_CREW} label="ماتركس يكلّف لحاله" hint="حتى لو شغّال، ما ينفّذ إلا إذا البوابة مؤهلة." /></div>
      </div>

      <div className="rounded-xl border border-slate-200 p-3 text-xs">
        <b className="text-sm">⏫ يصعّد بنود المراقب المتأخرة</b>
        <p className="mt-1 text-slate-600">بند بالصندوق صارله أكثر من يومين بلا حكم ← إشعار للمدير مرة باليوم. هسه: {st.monitorEscalate.overdue} بند.</p>
        <div className="mt-2"><OwnerSwitch switchKey={SWITCH_MATRIX_MONITOR_ESCALATE} label="تصعيد المراقب" hint="إشعار بس — ما يغيّر أي بيانات." /></div>
      </div>

      {st.recent.length > 0 && (
        <div className="text-xs">
          <b>آخر الأفعال:</b>
          {st.recent.map((a) => (
            <div key={a.id} className="mt-1 flex flex-wrap items-center justify-between gap-2 rounded-lg bg-slate-50 p-2">
              <span>{a.status === 'UNDONE' ? '↩️' : '✅'} {a.summary} <span className="text-slate-400">· {new Date(a.createdAt).toLocaleString('en-GB')}</span></span>
              {a.status === 'DONE' && (
                <button type="button" disabled={busy === a.id} onClick={() => void undo(a.id)} className="rounded-lg border border-rose-300 px-2 py-0.5 text-rose-700 disabled:opacity-50">تراجع</button>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
