import { useCallback, useEffect, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../api'
import SimGate from '../sim/SimGate'
import WiringBoard from '../sim/WiringBoard'
import { evaluateAction, wireExists } from '../sim/evaluate'
import type { SimAction, SimAttempt, SimDevice, SimEvent, SimExercise, Wire } from '../sim/types'

// ═══ شاشة التمرين ═══
//
// «يجي يلگه أنواع الكاميرات… ومحاكي يجيب الكامرة يشدها، شنو يحتاج
// إلها كيبل، وشلون يبرمجها».
//
// ⚠️ الغلط **ما ينهي المحاولة**: يعرض العاقبة («لو هذا كان بالميدان چان
// احترقت اللوحة») ويخلّيه يعيد. إنهاء المحاولة يعاقب، والعرض يعلّم.
//
// ⚠️ والدرجة **من السيرفر** مو محسوبة هنا: السيرفر يحسبها من أوزان
// الخطوات المخزونة عنده ويتحقّق من سجل الأحداث. لهذا **لازم** نرسل
// حدث `PASS` لكل خطوة تنجح — بدونها الدرجة تطلع صفراً.

export default function SimExercisePage() {
  return <SimGate><Runner /></SimGate>
}

function Runner() {
  const { id = '' } = useParams()
  const [ex, setEx] = useState<SimExercise | null>(null)
  const [attempt, setAttempt] = useState<SimAttempt | null>(null)
  const [loading, setLoading] = useState(true)
  const [err, setErr] = useState<string | null>(null)

  const [wires, setWires] = useState<Wire[]>([])
  const [stepIndex, setStepIndex] = useState(0)
  const [hintsUsed, setHintsUsed] = useState(0)
  const [wrongCount, setWrongCount] = useState(0)
  const [showHint, setShowHint] = useState(false)
  const [feedback, setFeedback] = useState<{ say: string; fatal?: boolean } | null>(null)
  const [finished, setFinished] = useState<SimAttempt | null>(null)
  const [saving, setSaving] = useState(false)

  // ⚠️ ما ننادي Date.now() **أثناء الرندر**: الرندر لازم يكون نقياً
  // (وهذا الي تمنعه قاعدة react-hooks). ننصّبها بأول تأثير.
  const startedAt = useRef(0)
  const pending = useRef<SimEvent[]>([])

  useEffect(() => {
    let alive = true
    void (async () => {
      startedAt.current = Date.now()
      try {
        const e = await api.getSimExercise(id)
        if (!alive) return
        setEx(e)
        const a = await api.startSimAttempt(id)
        if (!alive) return
        setAttempt(a)
        // ⚠️ نستأنف من وين وقف: تمرين توصيل ١٥ سلك مو شي ينخلص بخمس دقائق،
        // والفني يوقّف ويرجع. الحالة محفوظة بالسيرفر.
        setWires(a.state?.wires ?? [])
        setStepIndex(a.state?.stepIndex ?? 0)
        setHintsUsed(a.hintsUsed ?? 0)
        setWrongCount(a.wrongCount ?? 0)
      } catch (e2) {
        if (alive) setErr(e2 instanceof Error ? e2.message : 'تعذر فتح التمرين')
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [id])

  const deviceMap: Record<string, SimDevice> = {}
  for (const d of ex?.devices ?? []) deviceMap[d.id] = d

  const steps = ex?.steps ?? []
  const step = steps[stepIndex]
  const done = !!ex && stepIndex >= steps.length

  const push = (e: SimEvent) => pending.current.push(e)
  const ms = () => Date.now() - startedAt.current

  const persist = useCallback(async (nextWires: Wire[], nextStep: number, hints: number, wrongs: number) => {
    if (!attempt) return
    const events = pending.current
    pending.current = []
    try {
      await api.saveSimProgress(attempt.id, {
        state: { wires: nextWires, stepIndex: nextStep },
        stepsPassed: nextStep, hintsUsed: hints, wrongCount: wrongs, events,
      })
    } catch { /* الحفظ المرحلي ما يوقف التمرين — الأحداث تنعاد بالإنهاء */
      pending.current = [...events, ...pending.current]
    }
  }, [attempt])

  const onAction = (a: SimAction) => {
    if (done || !step) return
    if (wireExists(wires, a.from, a.to)) {
      setFeedback({ say: 'هذا السلك موصول أصلاً.' })
      return
    }
    push({ kind: 'ACTION', stepIndex, atMs: ms(), payload: { from: a.from, to: a.to } })
    const v = evaluateAction(step, a)

    if (v.result === 'PASS') {
      const nextWires = [...wires, { from: a.from, to: a.to }]
      const nextStep = stepIndex + 1
      setWires(nextWires)
      setStepIndex(nextStep)
      setFeedback(null)
      setShowHint(false)
      // ⚠️ حدث PASS **ضروري** — السيرفر يحدّ الدرجة بعدد هالأحداث.
      push({ kind: 'PASS', stepIndex, atMs: ms() })
      void persist(nextWires, nextStep, hintsUsed, wrongCount)
      return
    }

    const nextWrong = wrongCount + 1
    setWrongCount(nextWrong)
    setFeedback({ say: v.say ?? 'غلط.', fatal: v.fatal })
    push({ kind: 'WRONG', stepIndex, atMs: ms(), payload: { say: v.say ?? '' } })
    void persist(wires, stepIndex, hintsUsed, nextWrong)
  }

  const useHint = () => {
    if (showHint) return
    setShowHint(true)
    const n = hintsUsed + 1
    setHintsUsed(n)
    push({ kind: 'HINT', stepIndex, atMs: ms() })
    void persist(wires, stepIndex, n, wrongCount)
  }

  const removeWire = (w: Wire) => {
    const next = wires.filter((x) => !(x.from === w.from && x.to === w.to))
    setWires(next)
    push({ kind: 'ACTION', stepIndex, atMs: ms(), payload: { removed: `${w.from}→${w.to}` } })
    void persist(next, stepIndex, hintsUsed, wrongCount)
  }

  const finish = async () => {
    if (!attempt) return
    setSaving(true)
    const events = pending.current
    pending.current = []
    try {
      const a = await api.finishSimAttempt(attempt.id, {
        state: { wires, stepIndex },
        stepsPassed: stepIndex, hintsUsed, wrongCount,
        durationSec: Math.round(ms() / 1000), events,
      })
      setFinished(a)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر إنهاء المحاولة')
      pending.current = events
    } finally { setSaving(false) }
  }

  if (loading) return <p className="text-slate-400">جاري التحميل…</p>
  if (err && !ex) return <p className="rounded-lg bg-red-50 p-4 text-red-600">{err}</p>
  if (!ex) return null

  const unverified = (ex.devices ?? []).some((d) => !d.verified) || !ex.verified

  return (
    <div dir="rtl" className="space-y-4">
      {/* ⚠️ شريط السلامة **دائم** مو نافذة تنضغط مرة: النافذة تنضغط
          بأول أسبوع وتنتسى، وهذا تدريب على أسلاك وكهرباء. */}
      <div className="rounded-xl bg-amber-50 px-4 py-3 text-[12.5px] font-bold leading-relaxed text-amber-900 ring-1 ring-amber-200">
        ⚠️ هذا تدريب محاكاة — لا تعتمد عليه بالميدان. الرجوع لكتالوگ الشركة
        المصنّعة إلزامي قبل أي توصيل حقيقي.
        {unverified && (
          <span className="mt-1 block font-normal">
            ومعلومات هذا الجهاز <b>غير محقّقة ميدانياً</b> — ألوان الأسلاك تختلف بين مصنّع وآخر.
          </span>
        )}
      </div>

      <div className="flex flex-wrap items-start justify-between gap-3">
        <Link to="/simulator-lab" className="text-sm text-brand-700 hover:underline">← رجوع للمختبر</Link>
        <div className="text-right">
          <h2 className="text-xl font-bold text-brand-900">{ex.title}</h2>
          {ex.brief && <p className="mt-0.5 text-sm text-slate-500">{ex.brief}</p>}
        </div>
      </div>

      {/* ⚠️ أول بوابة «افتحه من الحاسبة» بالنظام: سحب أسلاك بشاشة ٦ إنچ
          ما يشتغل، والتنازل بالمحرّك أسوأ من رسالة صريحة. */}
      <div className="rounded-2xl bg-white p-8 text-center shadow md:hidden">
        <p className="text-lg font-bold text-slate-700">🖥️ افتحه من الحاسبة</p>
        <p className="mt-2 text-sm text-slate-500">
          لوحة توصيل الأسلاك تحتاج شاشة كبيرة — افتح المختبر من الحاسبة أو اللوحي.
        </p>
      </div>

      <div className="hidden gap-4 md:grid md:grid-cols-[1fr_20rem]">
        <div>
          <WiringBoard
            scene={ex.scene} devices={deviceMap} wires={wires}
            onAction={onAction} onRemoveWire={removeWire}
            highlight={showHint && step?.expect.from ? [step.expect.from, step.expect.to ?? ''] : []}
            readOnly={done || !!finished}
          />
          <p className="mt-2 text-[11px] text-slate-400">
            اضغط طرفاً ثم طرفاً ثانياً حتى توصّلهما · اضغط سلكاً حتى تحذفه · مرّر على الطرف حتى تشوف وظيفته
          </p>
        </div>

        <aside className="space-y-3">
          {finished ? (
            <div className={`rounded-2xl p-5 text-center shadow ${finished.status === 'PASSED' ? 'bg-emerald-50' : 'bg-amber-50'}`}>
              <p className="text-3xl font-black text-brand-900">{finished.score}</p>
              <p className="mt-1 text-sm font-bold text-slate-700">
                {finished.status === 'PASSED' ? '✅ نجحت' : 'ما وصلت درجة النجاح'}
              </p>
              <p className="mt-2 text-[11px] text-slate-500">
                تلميحات: {finished.hintsUsed} · أغلاط: {finished.wrongCount}
              </p>
              <Link to="/simulator-lab" className="mt-4 block rounded-lg bg-brand-600 px-4 py-2 text-sm font-bold text-white">
                رجوع للمختبر
              </Link>
            </div>
          ) : (
            <>
              <div className="rounded-2xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                <p className="text-[11px] font-bold text-slate-500">
                  الخطوة {Math.min(stepIndex + 1, steps.length)} من {steps.length}
                </p>
                {step ? (
                  <>
                    <p className="mt-1.5 text-[15px] font-bold text-brand-900">{step.title}</p>
                    <p className="mt-1 text-sm leading-relaxed text-slate-600">{step.instruction}</p>
                    {showHint && step.hint && (
                      <p className="mt-2 rounded-lg bg-sky-50 px-3 py-2 text-[12px] leading-relaxed text-sky-900">
                        💡 {step.hint}
                      </p>
                    )}
                    {!showHint && step.hint && (
                      <button onClick={useHint} className="mt-2 text-[11px] font-bold text-brand-700 hover:underline">
                        💡 أريد تلميح (ينقّص من الدرجة)
                      </button>
                    )}
                  </>
                ) : (
                  <p className="mt-2 text-sm font-bold text-emerald-700">✅ خلّصت كل الخطوات</p>
                )}
              </div>

              {feedback && (
                <div className={`rounded-2xl p-4 text-sm leading-relaxed shadow ${feedback.fatal ? 'bg-red-600 text-white' : 'bg-red-50 text-red-800'}`}>
                  {feedback.fatal && <p className="mb-1 text-xs font-black">⚠️ لو هذا كان بالميدان</p>}
                  {feedback.say}
                  <button onClick={() => setFeedback(null)}
                    className={`mt-2 block text-[11px] font-bold underline ${feedback.fatal ? 'text-white/90' : 'text-red-700'}`}>
                    فهمت — أعيد
                  </button>
                </div>
              )}

              <div className="rounded-2xl border border-white bg-white p-4 text-[11px] text-slate-500 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
                <p>أسلاك موصولة: <b className="text-slate-700">{wires.length}</b></p>
                <p className="mt-1">تلميحات: {hintsUsed} · أغلاط: {wrongCount}</p>
              </div>

              {done && (
                <button onClick={finish} disabled={saving}
                  className="w-full rounded-xl bg-emerald-700 px-4 py-3 text-sm font-bold text-white hover:bg-emerald-800 disabled:opacity-60">
                  {saving ? 'يحسب الدرجة…' : 'أنهِ التمرين واحسب الدرجة'}
                </button>
              )}
              {err && <p className="rounded-lg bg-red-50 p-3 text-xs text-red-600">{err}</p>}
            </>
          )}

          <ol className="space-y-1 rounded-2xl border border-white bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
            {steps.map((s, i) => (
              <li key={s.index} className={`text-[12px] ${i < stepIndex ? 'text-emerald-700' : i === stepIndex ? 'font-bold text-brand-900' : 'text-slate-400'}`}>
                {i < stepIndex ? '✅' : i === stepIndex ? '▸' : '○'} {s.title || s.instruction}
              </li>
            ))}
          </ol>
        </aside>
      </div>
    </div>
  )
}
