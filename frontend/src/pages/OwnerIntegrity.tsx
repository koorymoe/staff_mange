import { useState } from 'react'
import { apiRequest } from '../api'
import { useApi } from '../utils/useApi'
import { card } from '../ui/classes'
import { num } from '../utils/format'

// ═══ 🩺 سلامة البيانات (دازاي 11-02 — شبكة الأمان ٢) ═══
// فحص **قراءة بس**: كل سطر يعد صفوف تكسر قاعدة يفترضها النظام. ما يغيّر ولا
// صف — التصليح يصير لاحقاً، بند بند، بعد ما دازاي يشوف «شنو راح يتغير».

type Result = { key: string; domain: string; severity: 'critical' | 'high' | 'medium' | 'low'; title: string; why: string; fix: string; count: number; error?: string; ms: number }
type Latest = {
  run: { startedAt: string; problems: number; critical: number; failed: number; trigger: string } | null
  results: Result[]
  previous: Record<string, number>
  history: { startedAt: string; problems: number; critical: number }[]
}
type Sample = { id: string; info: string }

const SEV: Record<Result['severity'], { label: string; cls: string; rank: number }> = {
  critical: { label: '🔴 خطير', cls: 'bg-red-50 text-red-700 ring-red-200', rank: 0 },
  high: { label: '🟠 عالي', cls: 'bg-orange-50 text-orange-700 ring-orange-200', rank: 1 },
  medium: { label: '🟡 متوسط', cls: 'bg-amber-50 text-amber-700 ring-amber-200', rank: 2 },
  low: { label: '⚪ بسيط', cls: 'bg-slate-50 text-slate-600 ring-slate-200', rank: 3 },
}
const DOMAIN: Record<string, string> = {
  bookings: 'الحجوزات', money: 'الفلوس', people: 'الموظفين والدوام', projects: 'المشاريع',
  customers: 'الزبائن والجودة', ops: 'المخازن والمركبات', system: 'النظام وماتركس',
}

export default function OwnerIntegrity() {
  const [reload, setReload] = useState(0)
  const { data, error } = useApi(() => apiRequest<Latest>('/integrity'), [reload])
  const [busy, setBusy] = useState(false)
  const [runErr, setRunErr] = useState('')
  const [open, setOpen] = useState<string | null>(null)
  const [samples, setSamples] = useState<Record<string, Sample[] | string>>({})
  const [onlyBad, setOnlyBad] = useState(true)

  const runNow = async () => {
    setBusy(true); setRunErr('')
    try { await apiRequest('/integrity/run', { method: 'POST' }); setReload((n) => n + 1) }
    catch (e) { setRunErr(e instanceof Error ? e.message : 'تعذر الفحص') }
    finally { setBusy(false) }
  }
  const toggle = async (key: string) => {
    if (open === key) { setOpen(null); return }
    setOpen(key)
    if (samples[key]) return
    try { const rows = await apiRequest<Sample[]>(`/integrity/${encodeURIComponent(key)}/sample`); setSamples((s) => ({ ...s, [key]: rows })) }
    catch (e) { setSamples((s) => ({ ...s, [key]: e instanceof Error ? e.message : 'تعذر' })) }
  }

  const results = [...(data?.results ?? [])]
    .filter((r) => !onlyBad || r.count > 0 || r.error)
    .sort((a, b) => SEV[a.severity].rank - SEV[b.severity].rank || b.count - a.count)
  const bad = (data?.results ?? []).filter((r) => r.count > 0)

  return (
    <div dir="rtl" className="space-y-4">
      <div className={card}>
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h1 className="text-2xl font-extrabold text-[#0f2040]">🩺 سلامة البيانات</h1>
            <p className="text-sm text-slate-500">فحص <b>قراءة بس</b> — يعد كل خلل بالبيانات وما يغيّر ولا صف. يشتغل لحاله كل يوم، ويبلّغك إذا خلل خطير زاد.</p>
          </div>
          <button disabled={busy} onClick={() => void runNow()} className="rounded-xl bg-[#0f2040] px-4 py-2 text-sm font-bold text-white disabled:opacity-50">
            {busy ? '⏳ يفحص…' : '🔄 افحص هسه'}
          </button>
        </div>
        {(error || runErr) && <p className="mt-2 text-sm text-red-600">{error || runErr}</p>}
        {data?.run ? (
          <div className="mt-3 grid grid-cols-2 gap-2 sm:grid-cols-4">
            <Stat k="آخر فحص" v={new Date(data.run.startedAt).toLocaleString('ar-IQ', { dateStyle: 'medium', timeStyle: 'short' })} />
            <Stat k="فحوصات بيها خلل" v={`${bad.length} من ${data.results.length}`} bad={bad.length > 0} />
            <Stat k="صفوف خطيرة" v={num(data.run.critical)} bad={data.run.critical > 0} />
            <Stat k="فحوصات تعذّرت" v={String(data.run.failed)} bad={data.run.failed > 0} />
          </div>
        ) : data && <p className="mt-3 text-sm text-slate-500">ماكو فحص بعد — اضغط «افحص هسه».</p>}
      </div>

      {data?.run && (
        <div className={card}>
          <label className="mb-2 flex items-center gap-2 text-sm text-slate-600">
            <input type="checkbox" checked={onlyBad} onChange={(e) => setOnlyBad(e.target.checked)} /> الي بيها خلل بس
          </label>
          <div className="divide-y divide-slate-100">
            {results.map((r) => {
              const prev = data.previous[r.key]
              const delta = prev === undefined ? null : r.count - prev
              const s = samples[r.key]
              return (
                <div key={r.key} className="py-2.5">
                  <button onClick={() => void toggle(r.key)} className="flex w-full flex-wrap items-center gap-2 text-right">
                    <span className={`rounded-full px-2 py-0.5 text-[11px] font-bold ring-1 ${SEV[r.severity].cls}`}>{SEV[r.severity].label}</span>
                    <span className="text-[11px] text-slate-400">{DOMAIN[r.domain] ?? r.domain}</span>
                    <span className="flex-1 font-bold text-[#0f2040]">{r.title}</span>
                    {r.error
                      ? <span className="text-xs font-bold text-slate-400">تعذّر</span>
                      : <span className={`font-mono text-lg font-black ${r.count > 0 ? 'text-red-600' : 'text-emerald-600'}`} dir="ltr">{num(r.count)}</span>}
                    {delta !== null && delta !== 0 && <span className={`text-xs font-bold ${delta > 0 ? 'text-red-500' : 'text-emerald-600'}`} dir="ltr">{delta > 0 ? `+${delta}` : delta}</span>}
                  </button>
                  {open === r.key && (
                    <div className="mt-2 space-y-1 rounded-xl bg-slate-50 p-3 text-sm">
                      <p><b>ليش مهم:</b> {r.why}</p>
                      <p><b>شلون ينصلح:</b> {r.fix} <span className="text-xs text-slate-400">(ما ينصلح إلا بموافقتك)</span></p>
                      {r.error && <p className="text-xs text-red-600">{r.error}</p>}
                      {r.count > 0 && (s === undefined ? <p className="text-xs text-slate-400">جاري جلب الأمثلة…</p>
                        : typeof s === 'string' ? <p className="text-xs text-red-600">{s}</p>
                        : <ul className="list-disc pr-5 text-xs text-slate-600">{s.map((x) => <li key={x.id}>{x.info}</li>)}</ul>)}
                    </div>
                  )}
                </div>
              )
            })}
            {results.length === 0 && <p className="py-6 text-center text-emerald-700">✅ ماكو خلل بأي فحص</p>}
          </div>
        </div>
      )}
    </div>
  )
}

function Stat({ k, v, bad }: { k: string; v: string; bad?: boolean }) {
  return (
    <div className="rounded-xl bg-slate-50 p-3">
      <p className="text-[11px] text-slate-500">{k}</p>
      <p className={`text-lg font-black ${bad ? 'text-red-600' : 'text-[#0f2040]'}`}>{v}</p>
    </div>
  )
}
