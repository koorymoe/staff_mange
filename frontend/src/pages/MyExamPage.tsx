import { useEffect, useRef, useState } from 'react'
import { api, type TcMyExam } from '../api'

// ═══ اختباري الشهري — (ع) 10-13 ═══
// الموظف يسويه من حسابه: يبدي ← الوكت يمشي ← يسلّم ← النتيجة فوراً.
// بلا خصم نقاط — الي ما ينجح ينضاف لقائمة التدريب بس.
export default function MyExamPage() {
  const [d, setD] = useState<TcMyExam | null>(null)
  const [ans, setAns] = useState<Record<string, number>>({})
  const [left, setLeft] = useState<number | null>(null)
  const [result, setResult] = useState<{ score: number; total: number } | null>(null)
  const [busy, setBusy] = useState(false)
  useEffect(() => { api.tcMyExam().then(setD).catch(() => setD({ exam: null })) }, [])

  const started = d?.attempt && !d.attempt.finishedAt
  const ansRef = useRef(ans)
  useEffect(() => { ansRef.current = ans }, [ans])
  const sent = useRef(false)
  const submit = async () => {
    if (sent.current) return
    sent.current = true
    setBusy(true)
    try { setResult(await api.tcSubmitExam(ansRef.current)) } catch (e) { sent.current = false; alert(e instanceof Error ? e.message : 'تعذّر') } finally { setBusy(false) }
  }
  const submitRef = useRef(submit)
  useEffect(() => { submitRef.current = submit })
  // الوكت يخلص ← يتسلّم لحاله بالي جاوبه
  useEffect(() => {
    if (!started || !d?.exam || !d.attempt) return
    const end = new Date(d.attempt.startedAt).getTime() + d.exam.minutes * 60_000
    const t = window.setInterval(() => {
      const l = Math.max(0, Math.round((end - Date.now()) / 1000))
      setLeft(l)
      if (l === 0) { window.clearInterval(t); void submitRef.current() }
    }, 1000)
    return () => window.clearInterval(t)
  }, [started, d])

  if (!d) return <p className="p-8 text-center text-slate-400">جاري التحميل…</p>
  if (!d.exam) return <p className="m-6 rounded-2xl bg-white p-8 text-center text-slate-500 shadow-sm">ماكو اختبار شهري لدورك هالشهر ✅</p>
  const fin = result ?? (d.attempt?.finishedAt ? { score: d.attempt.score ?? 0, total: d.attempt.total ?? 0 } : null)
  return (
    <div dir="rtl" className="mx-auto max-w-3xl space-y-4">
      <div className="rounded-2xl bg-gradient-to-l from-[#0f2040] to-[#1a3a6c] p-6 text-white">
        <h1 className="text-xl font-extrabold">📝 {d.exam.title}</h1>
        <p className="mt-1 text-sm text-blue-200/80">{d.exam.count} سؤال · {d.exam.minutes} دقيقة · النجاح {d.exam.passPct}%</p>
        {started && !fin && left != null && <p className="mt-2 text-2xl font-black tabular-nums">⏱️ {Math.floor(left / 60)}:{String(left % 60).padStart(2, '0')}</p>}
      </div>
      {fin ? (
        <div className="rounded-2xl bg-white p-8 text-center shadow-sm">
          <p className="text-4xl font-black text-[#0f2040]">{fin.score}/{fin.total}</p>
          <p className={`mt-2 text-lg font-bold ${fin.total && fin.score * 100 / fin.total >= d.exam.passPct ? 'text-emerald-700' : 'text-amber-700'}`}>
            {fin.total && fin.score * 100 / fin.total >= d.exam.passPct ? '✅ ناجح — عاشت إيدك' : '📚 تحتاج شوية تدريب — وحدة التدريب راح تتابعك'}
          </p>
        </div>
      ) : !started ? (
        <div className="rounded-2xl bg-white p-8 text-center shadow-sm">
          <p className="text-slate-600">أول ما تبدي يمشي الوكت ({d.exam.minutes} دقيقة). جاوب بنفسك — هذا حتى نعرف شنو تحتاج تدريب.</p>
          <button onClick={() => api.tcStartExam().then(setD).catch(() => {})} className="mt-4 rounded-xl bg-[#0f2040] px-8 py-3 font-bold text-white">ابدي الاختبار</button>
        </div>
      ) : (
        <>
          {(d.questions ?? []).map((q, i) => (
            <div key={q.id} className="rounded-2xl bg-white p-5 shadow-sm">
              <p className="font-bold text-[#0f2040]">{i + 1}. {q.text}</p>
              <div className="mt-3 grid gap-2">
                {q.options.map((o, j) => (
                  <button key={j} onClick={() => setAns({ ...ans, [q.id]: j })}
                    className={`rounded-xl border px-4 py-2.5 text-right text-sm ${ans[q.id] === j ? 'border-[#2c5aad] bg-sky-50 font-bold text-[#0f2040]' : 'border-slate-200 hover:bg-slate-50'}`}>{o}</button>
                ))}
              </div>
            </div>
          ))}
          <button disabled={busy} onClick={() => { if (Object.keys(ans).length < (d.questions?.length ?? 0) && !confirm('بعدك ما جاوبت كل الأسئلة — تسلّم؟')) return; void submit() }}
            className="w-full rounded-xl bg-[#0f2040] py-3 font-bold text-white disabled:opacity-40">تسليم الاختبار ({Object.keys(ans).length}/{d.questions?.length ?? 0})</button>
        </>
      )}
    </div>
  )
}
