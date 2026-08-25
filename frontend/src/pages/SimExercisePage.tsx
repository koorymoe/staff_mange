import { lazy, Suspense, useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api } from '../api'
import SimGate from '../sim/SimGate'
import WiringBoard from '../sim/WiringBoard'
import { evaluateAction, wireExists } from '../sim/evaluate'
import type { SimAction, SimAttempt, SimDevice, SimEvent, SimExercise, Wire } from '../sim/types'
import { CABLE_GAUGES, voltageAtLoad, type Cable, type CableGaugeId } from '../sim3d/cable'

// ⚠️ Babylon تنزّل **بس** لمن يفتح المنظر الفيزيائي: الحزمة ثقيلة،
// وأغلب من يفتح المختبر يبدي بالمنطقي. `lazy` تخلّيها خارج حزمة الدخول.
const Workbench3D = lazy(() => import('../sim3d/Workbench3D'))
const CliTerminal = lazy(() => import('../cli/CliTerminal'))

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

  // ═══ المنظران (١٩) ═══ نفس الحالة بالضبط — `wires` وحدة للاثنين.
  const [view, setView] = useState<'physical' | 'logical'>('physical')
  // ═══ خصائص التمديد ═══ المقطع والطول يقرّران هبوط الفولتية.
  const [gauge, setGauge] = useState<CableGaugeId>('awg20')
  const [runLengthM, setRunLengthM] = useState(12)

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

  // ═══ تقييم تمارين سطر الأوامر ═══
  //
  // ⚠️ التقييم على **الحالة** مو على نص الأمر: الفني يوصل لنفس النتيجة
  // بمسارات مختلفة (ينشئ الـVLAN قبل المنفذ أو بعده)، وأي تقييم يطابق
  // الحروف المكتوبة يعاقبه على طريق صحيح تماماً. وهذا الي يفرّق بين
  // «محاكي» و«امتحان يحفظ الأوامر».
  const onCliState = useCallback((state: Record<string, unknown>, lastCommand: string, mode: string) => {
    const st = ex?.steps?.[stepIndex]
    if (!st || st.expect.op !== 'STATE_EQ' || !st.expect.path) return

    // `__mode` مسار خاص: النمط الحالي مو قيمة بالحالة.
    const actual = st.expect.path === '__mode'
      ? undefined
      : st.expect.path.split('.').reduce<unknown>(
          (a, k) => (typeof a === 'object' && a !== null ? (a as Record<string, unknown>)[k] : undefined), state)

    const ok = st.expect.path === '__mode'
      // ⚠️ `mode` الواصل مو `cliMode` بالحالة: `setCliMode` ما تنطبّق
      // إلا بالرندر الجاي، فقراءة الحالة هنا تعطي النمط **السابق** —
      // والخطوة ما تنجح إلا بأمر زائد بعدها.
      ? mode === st.expect.value
      : String(actual ?? '') === String(st.expect.value ?? '')
    if (!ok) return

    pending.current.push({ kind: 'PASS', stepIndex, atMs: Date.now() - startedAt.current })
    if (lastCommand) { /* الأمر الي نجّح الخطوة — يفيد بالمراجعة لاحقاً */ }
    setFeedback(null)
    setStepIndex((i) => i + 1)
  }, [ex, stepIndex])

  // ═══ محرّك كهربائي مبسّط ═══
  //
  // ⚠️ هذا **مو** جدول جاهز ولا رسالة مكتوبة سلفاً: القراءة تنحسب من
  // مقطع السلك وطوله وسحب الجهاز. غيّر المقطع لـ٢٢ ومدّد ٤٠ متر ← تشوف
  // الفولتية تنزل تحت حد التحرير والقفل ما ينفتح **مع إن التوصيل صحيح**.
  // هذا بالضبط العطل الي يوگف الفني ساعتين بالميدان.
  //
  // ⚠️ درجة الدقة هنا `F1` بسلّم الـFidelity (٦): مقاومة أومية بس، بلا
  // تيار اندفاع ولا حرارة. تكفي لهذا الدرس وما تكفي لامتحان أعطال طاقة —
  // ولمن نحتاج أعلى، الطبقة تتبدّل بلا ما تتغيّر هوية التمرين.
  const power = useMemo(() => {
    // ⚠️ المراجع **ما تنكتب بالكود**: اسم الجهاز بالمشهد (`ref`) يقرّره
    // مؤلّف التمرين مو المحرّك — بهذا التمرين `lock1`/`psu1` مو
    // `lock`/`psu`. ترميزها يدوياً يخلّي المحرّك ما يلگي ولا سلك
    // ويقرا «ماكو تغذية» للأبد. نلگاهن من نوع الجهاز بالمشهد.
    const refOf = (match: (d: SimDevice) => boolean) =>
      (ex?.scene.devices ?? []).find((sd) => {
        const dev = (ex?.devices ?? []).find((d) => d.id === sd.deviceId)
        return dev ? match(dev) : false
      })?.ref
    const lockRef = refOf((d) => d.engineKind === 'WIRING' && (d.terminals?.length ?? 0) > 4)
    const psuRef = refOf((d) => (d.terminals?.length ?? 0) <= 4)
    if (!lockRef || !psuRef) return { state: 'OFF' as const, volts: 0 }

    const has = (a: string, b: string) =>
      wires.some((w) => (w.from === a && w.to === b) || (w.from === b && w.to === a))
    const posOk = has(`${lockRef}:t_red`, `${psuRef}:out_pos`)
    const negOk = has(`${lockRef}:t_black`, `${psuRef}:out_neg`)
    const reversed = has(`${lockRef}:t_red`, `${psuRef}:out_neg`) && has(`${lockRef}:t_black`, `${psuRef}:out_pos`)
    if (reversed) return { state: 'REVERSED' as const, volts: 0 }
    if (!posOk || !negOk) return { state: 'OFF' as const, volts: 0 }

    const spec = (ex?.devices ?? []).find((d) => d.id.includes('keypad'))?.spec as
      | { power?: { currentA?: number } } | undefined
    const loadA = spec?.power?.currentA ?? 0.5
    const cables: Cable[] = [
      { id: 'c+', fromRef: psuRef, fromTerminal: 'out_pos', toRef: lockRef, toTerminal: 't_red', gauge, lengthM: runLengthM, colorHex: '#dc2626' },
      { id: 'c-', fromRef: psuRef, fromTerminal: 'out_neg', toRef: lockRef, toTerminal: 't_black', gauge, lengthM: runLengthM, colorHex: '#111827' },
    ]
    // ⚠️ الذهاب والإياب محسوبان جوّا `cableResistance` — فنمرّر كيبلاً
    // واحداً للحساب وإلا الطول ينحسب أربع مرات.
    const volts = voltageAtLoad(12, loadA, [cables[0]])
    return { state: volts >= 10.8 ? ('OK' as const) : ('LOW' as const), volts }
  }, [wires, gauge, runLengthM, ex])

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
          {/* ═══ مبدّل المنظر (١٩) ═══
              نفس الحالة بالضبط — `wires` وحدة للاثنين. المنظر الفيزيائي
              للتركيب، والمنطقي للسرعة والقراءة. ولا واحد منهما يخزن شي
              لحاله، فالتبديل ما يضيّع ولا سلك. */}
          {/* ⚠️ أدوات التوصيل (المنظران، المقطع، الطول) تخص تمارين
              **التوصيل** بس. عرضها بتمرين سطر أوامر يخلّي المتدرّب
              يدوّر علاقة بين مقطع السلك وأمر `vlan` — وماكو. */}
          {ex.engineKind !== 'CLI' && (
          <div className="mb-2 flex flex-wrap items-center gap-2">
            <div className="inline-flex overflow-hidden rounded-lg ring-1 ring-slate-300">
              {([['physical', '🧊 منظر فيزيائي'], ['logical', '🗺️ منظر منطقي']] as const).map(([v, label]) => (
                <button
                  key={v} onClick={() => setView(v)}
                  className={`px-3 py-1.5 text-xs font-bold transition ${
                    view === v ? 'bg-brand-700 text-white' : 'bg-white text-slate-600 hover:bg-slate-50'}`}
                >{label}</button>
              ))}
            </div>

            {/* خصائص التمديد — تأثيرها حقيقي على القياس، مو زينة */}
            <label className="flex items-center gap-1.5 text-[11px] text-slate-600">
              مقطع السلك
              <select
                value={gauge} onChange={(e) => setGauge(e.target.value as CableGaugeId)}
                className="rounded-md border border-slate-300 px-2 py-1 text-[11px]"
              >
                {CABLE_GAUGES.map((g) => <option key={g.id} value={g.id}>{g.label}</option>)}
              </select>
            </label>
            <label className="flex items-center gap-1.5 text-[11px] text-slate-600">
              طول التمديد (م)
              <input
                type="number" min={1} max={120} value={runLengthM}
                onChange={(e) => setRunLengthM(Math.max(1, Number(e.target.value) || 1))}
                className="w-16 rounded-md border border-slate-300 px-2 py-1 text-[11px]"
              />
            </label>
          </div>
          )}

          {ex.engineKind === 'CLI' ? (
            <Suspense fallback={
              <div className="flex h-[420px] items-center justify-center rounded-xl bg-[#080c10] text-slate-500">
                جاري فتح الجلسة…
              </div>
            }>
              {ex.cliGrammar ? (
                <CliTerminal
                  grammar={ex.cliGrammar}
                  initialState={{ hostname: 'Switch' }}
                  onStateChange={onCliState}
                  readOnly={done || !!finished}
                />
              ) : (
                <p className="rounded-xl bg-red-50 p-4 text-red-600">
                  ما وصل نحو الأوامر لهذا الجهاز — التمرين ما يشتغل بدونه.
                </p>
              )}
            </Suspense>
          ) : view === 'physical' ? (
            <Suspense fallback={
              <div className="flex h-[520px] items-center justify-center rounded-2xl bg-[#0e1219] text-slate-400">
                جاري تحميل المشهد الثلاثي…
              </div>
            }>
              <Workbench3D
                scene={ex.scene} devices={deviceMap} wires={wires}
                onAction={onAction} onRemoveWire={removeWire}
                highlight={showHint && step?.expect.from ? [step.expect.from, step.expect.to ?? ''] : []}
                readOnly={done || !!finished}
                gauge={gauge} runLengthM={runLengthM}
              />
            </Suspense>
          ) : (
            <WiringBoard
              scene={ex.scene} devices={deviceMap} wires={wires}
              onAction={onAction} onRemoveWire={removeWire}
              highlight={showHint && step?.expect.from ? [step.expect.from, step.expect.to ?? ''] : []}
              readOnly={done || !!finished}
            />
          )}

          <p className="mt-2 text-[11px] text-slate-400">
            {ex.engineKind === 'CLI'
              ? 'اضغط داخل الشاشة السوداء حتى تكتب · `?` تعطيك الأوامر المتاحة · Tab يكمّل · ↑ يرجّع آخر أمر · Ctrl-Z يطلّعك للنمط المميّز'
              : view === 'physical'
              ? 'اسحب بالماوس حتى تدور حول الطاولة · عجلة الماوس تقرّب · اضغط طرفاً ثم طرفاً ثانياً حتى توصّلهما · اضغط سلكاً حتى تحذفه'
              : 'اضغط طرفاً ثم طرفاً ثانياً حتى توصّلهما · اضغط سلكاً حتى تحذفه · مرّر على الطرف حتى تشوف وظيفته'}
          </p>

          {/* ═══ الأڤوميتر (١٠) ═══
              «المتدرّب ما يشوف الجواب — يقيسه». القراءة محسوبة من المقطع
              والطول والسحب، فتتغيّر لمن يتغيّر أي واحد منهن. */}
          {ex.engineKind !== 'CLI' && (
          <div className="mt-3 rounded-xl bg-slate-900 p-4 text-slate-100 ring-1 ring-slate-700">
            <div className="flex items-center justify-between gap-3">
              <span className="text-[11px] text-slate-400">قياس على أطراف اللوحة</span>
              <span className="font-mono text-2xl font-bold tabular-nums" style={{
                color: power.state === 'OK' ? '#4ade80' : power.state === 'LOW' ? '#fbbf24' : '#f87171',
              }}>
                {power.state === 'OFF' ? '— — —' : `${power.volts.toFixed(2)} V`}
              </span>
            </div>
            <p className="mt-1.5 text-[11px] leading-relaxed text-slate-300">
              {power.state === 'OFF' && 'ماكو تغذية واصلة — اللوحة مطفية. وصّل الموجب والسالب.'}
              {power.state === 'OK' && '✅ التغذية سليمة — اللوحة تشتغل والقفل يتحرّر.'}
              {power.state === 'LOW' && (
                <>⚠️ <b>هبوط فولتية</b>: واصل أقل من ١٠٫٨ فولت. التوصيل صحيح بس المقطع نحيف أو
                التمديد طويل — اللوحة تشتغل والقفل ما يتحرّر. زيد المقطع أو قصّر التمديد.</>
              )}
              {power.state === 'REVERSED' && '🔥 القطبية معكوسة — بالميدان هذي تحرق اللوحة.'}
            </p>
          </div>
          )}
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
