import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { api } from '../api'
import PageHeader from '../components/PageHeader'

// ═══ فحص دقة ماتركس — قرار (ع) 10-15 ═══
// كل فحص يشتغل على البيانات الحقيقية: ✓ = حسابات ماتركس سليمة، ✗ = أكو حالات
// تحتاج تصحيح (ويّا أمثلة). «حساب غلط» يبلّغ المالك كل يوم لحد ما ينصلح.

interface Check { key: string; area: string; title: string; why: string; severity: 'ERROR' | 'WARN'; count: number; ok: boolean; examples: string[]; link?: string }

export default function MatrixSelfCheckPage() {
  const [checks, setChecks] = useState<Check[] | null>(null)
  const [reload, setReload] = useState(0)
  const [err, setErr] = useState('')
  useEffect(() => {
    let alive = true
    api.matrixSelfCheck().then((d) => { if (alive) { setChecks(d.checks as Check[]); setErr('') } }).catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [reload])
  const okN = checks?.filter((c) => c.ok).length ?? 0
  const errN = checks?.filter((c) => !c.ok && c.severity === 'ERROR').length ?? 0
  const areas = [...new Set((checks ?? []).map((c) => c.area))]
  return (
    <div dir="rtl" className="space-y-4">
      <PageHeader title="🧪 فحص دقة ماتركس" subtitle="فحوصات على البيانات الحقيقية — هل حسابات ماتركس وتدقيقاته صحيحة؟" />
      {err && <p className="rounded-xl bg-red-50 p-3 text-sm text-red-700">{err}</p>}
      {checks && (
        <div className="flex flex-wrap items-center gap-3 rounded-2xl bg-white p-4 shadow-sm">
          <p className="text-2xl font-black text-[#0f2040]">{okN} / {checks.length} <span className="text-base font-bold text-slate-500">فحص سليم</span></p>
          {errN > 0 ? <span className="rounded-full bg-red-100 px-3 py-1 text-sm font-bold text-red-700">✗ {errN} حساب غلط يحتاج تصحيح</span>
            : <span className="rounded-full bg-emerald-100 px-3 py-1 text-sm font-bold text-emerald-700">✓ ماكو حساب غلط</span>}
          <button onClick={() => setReload((n) => n + 1)} className="mr-auto rounded-lg bg-slate-100 px-3 py-1.5 text-sm">↻ إعادة الفحص</button>
        </div>
      )}
      {areas.map((a) => (
        <section key={a} className="space-y-2">
          <h2 className="text-sm font-bold text-slate-600">{a}</h2>
          {checks!.filter((c) => c.area === a).map((c) => (
            <div key={c.key} className={`rounded-xl border p-3 ${c.ok ? 'border-emerald-200 bg-emerald-50/50' : c.severity === 'ERROR' ? 'border-red-300 bg-red-50' : 'border-amber-300 bg-amber-50'}`}>
              <div className="flex flex-wrap items-center gap-2">
                <span className="text-lg">{c.ok ? '✅' : c.severity === 'ERROR' ? '❌' : '⚠️'}</span>
                <b className="text-sm">{c.title}</b>
                {!c.ok && <span className="rounded-full bg-white px-2 py-0.5 text-xs font-bold">{c.count < 0 ? 'تعذر التشغيل' : `${c.count} حالة`}</span>}
                <span className="text-xs text-slate-500">{c.severity === 'ERROR' ? 'حساب' : 'بيانات'}</span>
                {!c.ok && c.link && <Link to={c.link} className="mr-auto text-xs font-bold text-brand-700 underline">افتح ←</Link>}
              </div>
              <p className="mt-0.5 text-xs text-slate-600">{c.why}</p>
              {!c.ok && c.examples.length > 0 && <p className="mt-1 text-xs text-slate-700">أمثلة: {c.examples.join(' · ')}</p>}
            </div>
          ))}
        </section>
      ))}
    </div>
  )
}
