import { useEffect, useState } from 'react'
import { api, AI_METRIC_LABELS, type AiCatalog, type AiMetric, type AiSignal } from '../api'

// ═══ إحصائيات ومؤشرات الذكاء الاصطناعي ═══
//
// ⚠️ للمالك ومدير النظام حصراً — طلب صريح. تحليل «ليش هذا الموظف
// وقّف الشغل» بيد زميله يتحول لسلاح داخلي، ويخلي الموظفين يخافون
// يكتبون السبب الحقيقي فتضيع الفايدة كلها.
//
// الفكرة المركزية بهاي الشاشة: **الأدلة تنعرض قبل الحكم**.
// الحكم رأي، والأدلة حقائق. لو النظام گال «الموظف مقصّر» ولقيت
// الأدلة تگول غير — الأدلة هي الصح. لهذا ما نخبّيها.

const SEVERITY: Record<string, { label: string; cls: string }> = {
  INFO: { label: 'معلومة', cls: 'bg-slate-100 text-slate-700' },
  WATCH: { label: 'راقب', cls: 'bg-amber-100 text-amber-800' },
  WARN: { label: 'تنبيه', cls: 'bg-orange-100 text-orange-800' },
  CRITICAL: { label: 'خطير', cls: 'bg-red-100 text-red-800' },
}

// تسمية عربية لكل حقيقة — المفتاح الإنجليزي ما يقراه أحد.
const FACT_LABELS: Record<string, string> = {
  stopReason: 'السبب الي كتبه الموظف',
  stoppedAtHour: 'ساعة التوقف',
  minutesToShiftEnd: 'باقي على نهاية الدوام (دقيقة)',
  workedMinutes: 'اشتغل قبل التوقف (دقيقة)',
  procurementRequests: 'طلبات مواد لهذا الحجز',
  requestedBeforeStop: 'طلب المادة قبل ما يوقّف',
  lastRequestStatus: 'حالة آخر طلب',
  cartItemsTotal: 'مواد بسلة الزبون',
  cartItemsAfterStart: 'مواد انضافت بعد ما بدأ الشغل',
  stopsLast30Days: 'توقفاته بآخر ٣٠ يوم',
}

const fmtFact = (v: unknown) => (v === true ? 'إي' : v === false ? 'لا' : v === null || v === '' ? '—' : String(v))

export default function AiInsightsPage() {
  const [signals, setSignals] = useState<AiSignal[]>([])
  const [catalog, setCatalog] = useState<AiCatalog | null>(null)
  const [metrics, setMetrics] = useState<AiMetric[]>([])
  const [loading, setLoading] = useState(true)
  const [busy, setBusy] = useState(false)
  const [err, setErr] = useState<string | null>(null)

  // ⚠️ الجلب يصير بمكان **واحد** داخل الـeffect، وزر التحليل يطلبه
  // برفع العدّاد بدل ما ينادي نسخة ثانية من نفس الكود. هذا يمنع
  // سباق الطلبات (`alive`) ويخلّي مسار الجلب واحد ما ينحرف.
  const [reload, setReload] = useState(0)

  useEffect(() => {
    let alive = true
    void (async () => {
      try {
        const [s, c, m] = await Promise.all([api.getAiSignals(), api.getAiCatalog(), api.getAiMetrics()])
        if (alive) { setSignals(s); setCatalog(c); setMetrics(m) }
      } catch (e) {
        if (alive) setErr(e instanceof Error ? e.message : 'تعذر الجلب')
      } finally {
        if (alive) setLoading(false)
      }
    })()
    return () => { alive = false }
  }, [reload])

  const analyze = async () => {
    setBusy(true)
    setErr(null)
    try {
      await api.runAiProcess()
      // المؤشرات تنحسب من الأدلة، فلازم تنعاد بعد التحليل مباشرة —
      // وإلا تبقى تعرض أرقام ما تشمل الي انحلل هسه.
      await api.recomputeAiMetrics()
      setReload((n) => n + 1)
    } catch (e) {
      setErr(e instanceof Error ? e.message : 'تعذر التحليل')
    } finally {
      setBusy(false)
    }
  }

  const analyzed = signals.filter((s) => s.verdict)

  return (
    <div dir="rtl" className="space-y-5">
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">🧠 إحصائيات ومؤشرات الذكاء الاصطناعي</h2>
          <p className="mt-1 text-sm text-slate-500">
            النظام يراقب الأحداث، يجمع الأدلة من كل مكان، ويطلّع تفسيراً. <b>المالك ومدير النظام بس.</b>
          </p>
        </div>
        <button
          onClick={analyze}
          disabled={busy}
          className="rounded-xl bg-gradient-to-l from-brand-500 to-brand-800 px-4 py-2.5 text-sm font-bold text-white disabled:opacity-50"
        >
          {busy ? 'جاري التحليل...' : '⚙️ حلّل الإشارات المعلّقة'}
        </button>
      </div>

      {err && <p className="rounded-lg bg-red-50 p-4 text-red-600">{err}</p>}

      {/* حالة الهيكلة — شنو شغّال وشنو ينتظر المنصّة */}
      {catalog && (
        <div className="rounded-2xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <div className="flex flex-wrap items-center gap-2">
            <h3 className="font-bold text-[#0f2040]">حالة الهيكلة</h3>
            <span className="rounded-full bg-slate-100 px-2.5 py-1 text-xs font-bold text-slate-700">
              المحلّل الحالي: {catalog.judge}
            </span>
            <span className={`rounded-full px-2.5 py-1 text-xs font-bold ${catalog.platformLinked ? 'bg-emerald-100 text-emerald-800' : 'bg-amber-100 text-amber-800'}`}>
              {catalog.platformLinked ? 'منصّة ذكاء مربوطة' : 'ماكو منصّة — قواعد حتمية بس'}
            </span>
          </div>
          {/* ⚠️ نگولها صريحة: الشغّال اليوم قواعد يقينية ما تخمّن.
              لما ننشترك بمنصّة تجي فوقها، وما ينرمي شي من الي انبنى. */}
          <p className="mt-2 text-xs leading-6 text-slate-500">
            الي يشتغل اليوم <b>قواعد حتمية</b>: النظام يجمع الحقائق من الجداول (الساعة، طلبات المواد، سلة الزبون،
            سجل الموظف) ويفسّرها بقواعد مكتوبة — بلا اشتراك بأي منصّة. لما ننشترك، المنصّة تجي <b>فوق نفس الأدلة</b>
            وتضيف التفسير الأعمق والصياغة. الي انبنى هسه ما ينرمي.
          </p>

          <div className="mt-3 grid grid-cols-1 gap-3 sm:grid-cols-2">
            <div>
              <p className="mb-1.5 text-xs font-bold text-slate-500">الأحداث الي ينراقبها</p>
              {catalog.signals.map((s) => (
                <p key={s.key} className="text-xs text-slate-600">
                  {s.ready ? '✅' : '⏳'} <b>{s.label}</b> — {s.detail}
                </p>
              ))}
            </div>
            <div>
              <p className="mb-1.5 text-xs font-bold text-slate-500">المؤشرات</p>
              {catalog.metrics.map((m) => (
                <p key={m.key} className="text-xs text-slate-600">
                  {m.ready ? '✅' : '⏳'} <b>{m.label}</b> — {m.detail}
                </p>
              ))}
            </div>
          </div>
        </div>
      )}

      {/* ═══ المؤشرات ═══ */}
      {metrics.length > 0 && (
        <div className="rounded-2xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
          <h3 className="mb-3 font-bold text-[#0f2040]">📊 المؤشرات — آخر ٣٠ يوم</h3>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-4">
            {metrics.map((m) => (
              <div key={m.id} className="rounded-xl border border-slate-200 bg-slate-50/60 p-3">
                <p className="text-[11px] font-bold text-slate-500">
                  {AI_METRIC_LABELS[m.metricKey] || m.metricKey}
                </p>
                <p className="mt-1 text-2xl font-extrabold text-[#0f2040]">
                  {m.metricKey === 'STOP_MINUTES_AVG'
                    ? `${Math.round(m.value)} دقيقة`
                    : `${m.value.toFixed(0)}%`}
                </p>
                {/* ⚠️ عدد العيّنات ينعرض دائماً: «٥٠٪» من عيّنتين مو
                    مثل «٥٠٪» من مئتين، والقرار يختلف. */}
                <p className="mt-0.5 text-[10.5px] text-slate-400">من {m.sampleCount} حالة</p>
              </div>
            ))}
          </div>
        </div>
      )}

      {loading && <p className="text-slate-400">جاري التحميل...</p>}
      {!loading && analyzed.length === 0 && (
        <p className="rounded-xl border border-white bg-white p-8 text-center text-slate-400">
          ماكو تحليلات بعد — الإشارات تتراكم مع الشغل، واضغط «حلّل» حتى تطلع النتيجة.
        </p>
      )}

      <div className="space-y-3">
        {analyzed.map((s) => {
          const v = s.verdict!
          const sev = SEVERITY[v.severity] || SEVERITY.INFO
          const facts = s.evidence?.facts || {}
          const gaps = s.evidence?.gaps || []
          return (
            <div key={s.id} className="rounded-2xl border border-white bg-white p-5 shadow-[0_4px_20px_rgba(15,32,64,0.06)]">
              <div className="flex flex-wrap items-center gap-2">
                <span className={`rounded-full px-2.5 py-1 text-xs font-bold ${sev.cls}`}>{sev.label}</span>
                <h3 className="font-bold text-[#0f2040]">{v.headline}</h3>
                <span className="text-xs text-slate-400">
                  {s.employeeName && <>👤 {s.employeeName} · </>}
                  {new Date(s.occurredAt).toLocaleString('en-GB')}
                </span>
                {/* ⚠️ الثقة الواطية تنقال صراحة — القرار على أساس حكم
                    «مو متأكد» أسوأ من ماكو حكم أصلاً. */}
                <span className={`mr-auto rounded-lg px-2 py-1 text-[11px] font-bold ${v.confidence >= 70 ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-600'}`}>
                  {v.confidence >= 70 ? `ثقة ${v.confidence}%` : `مو متأكد (${v.confidence}%)`}
                </span>
              </div>

              {v.reasoning && <p className="mt-2 text-sm leading-7 text-slate-700">{v.reasoning}</p>}
              {v.suggestion && (
                <p className="mt-2 rounded-lg bg-sky-50 px-3 py-2 text-sm text-sky-900">💡 {v.suggestion}</p>
              )}
              {v.blameEmployeeName && (
                <p className="mt-2 text-xs font-bold text-amber-800">⚠️ يحتاج مراجعة مع: {v.blameEmployeeName}</p>
              )}

              {/* الأدلة — الحقائق الي انبنى عليها الحكم */}
              <details className="mt-3">
                <summary className="cursor-pointer text-xs font-bold text-slate-500">
                  📋 الأدلة الي انبنى عليها الحكم ({Object.keys(facts).length})
                </summary>
                <div className="mt-2 grid grid-cols-1 gap-x-6 gap-y-1 rounded-xl bg-slate-50 p-3 text-xs sm:grid-cols-2">
                  {Object.entries(facts).map(([k, val]) => (
                    <p key={k} className="flex justify-between gap-2">
                      <span className="text-slate-500">{FACT_LABELS[k] || k}</span>
                      <b className="text-slate-800">{fmtFact(val)}</b>
                    </p>
                  ))}
                </div>
                {/* ⚠️ الفجوات تنعرض مو تنخبّى: الحكم على أدلة ناقصة
                    لازم يعرفه الي يقراه. */}
                {gaps.length > 0 && (
                  <p className="mt-2 rounded-lg bg-amber-50 px-3 py-2 text-xs text-amber-900">
                    ⚠️ ما قدرنا نجمع: {gaps.join(' · ')}
                  </p>
                )}
              </details>

              <p className="mt-2 text-[11px] text-slate-400">
                المصدر: {v.source === 'RULES' ? 'قواعد النظام (حتمية)' : `منصّة ذكاء${v.modelName ? ` — ${v.modelName}` : ''}`}
              </p>
            </div>
          )
        })}
      </div>
    </div>
  )
}
