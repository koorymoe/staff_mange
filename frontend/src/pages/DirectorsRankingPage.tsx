import { useEffect, useState } from 'react'
import { apiRequest } from '../api'

// 🏆 تصنيف المدراء (دازاي 11-01): مدير إدارة مقابل مدير إدارة — على الهيكل الجديد
// (دازاي 11-04): «ماريد بس ارقام، اريد اشوف على شنو اعتمد» + تصنيف النواب + تملي الشاشة.
type Evidence = { label: string; value: string; note?: string }
type Criterion = {
  key: string
  label: string
  weight: number
  score: number | null
  detail: string
  formula: string
  evidence: Evidence[]
}
type Row = {
  rank: number
  divisionName: string
  employeeName: string
  acting: boolean
  people: number
  total: number | null
  criteria: Criterion[]
  divisionTotal?: number | null
  ownTotal?: number | null
}
type Ranking = { period: string; rows: Row[]; deputies: Row[]; units?: Row[] }

const thisMonth = () => new Date().toISOString().slice(0, 7)
const tone = (v: number | null | undefined) =>
  v === null || v === undefined ? 'text-slate-400' : v >= 75 ? 'text-green-600' : v >= 50 ? 'text-amber-600' : 'text-red-600'
const bar = (v: number | null | undefined) =>
  v === null || v === undefined ? 'bg-slate-300' : v >= 75 ? 'bg-green-600' : v >= 50 ? 'bg-amber-500' : 'bg-red-600'
const medal = (r: number) => (r === 1 ? '🥇' : r === 2 ? '🥈' : r === 3 ? '🥉' : `#${r}`)

function CriterionTile({ c, open, onClick }: { c: Criterion; open: boolean; onClick: () => void }) {
  return (
    <button
      type="button"
      onClick={onClick}
      className={`rounded-xl border p-3 text-right transition ${open ? 'border-brand-500 bg-brand-50' : 'border-slate-200 bg-white hover:border-brand-300'}`}
    >
      <div className="flex items-center justify-between gap-2 text-sm">
        <span className="font-semibold text-slate-700">
          {c.label} <span className="text-xs font-normal text-slate-400">{c.weight}٪</span>
        </span>
        <b className={tone(c.score)}>{c.score === null ? '—' : c.score}</b>
      </div>
      <div className="mt-2 h-1.5 overflow-hidden rounded bg-slate-200">
        <div className={`h-full ${bar(c.score)}`} style={{ width: `${c.score ?? 0}%` }} />
      </div>
      <div className="mt-1 flex items-center justify-between text-xs text-slate-500">
        <span>{c.detail}</span>
        <span className="text-brand-600">{open ? 'سدّ ▲' : 'التفاصيل ▼'}</span>
      </div>
    </button>
  )
}

function EvidencePanel({ c }: { c: Criterion }) {
  return (
    <div className="mt-3 rounded-xl border border-brand-200 bg-brand-50/40 p-4">
      <div className="text-sm font-bold text-slate-800">على شنو اعتمد «{c.label}»</div>
      <div className="mt-1 text-xs text-slate-600">
        <b>المعادلة:</b> {c.formula}
      </div>
      {c.score === null && (
        <div className="mt-2 text-xs text-slate-500">ماكو بيانات لهالمعيار هالشهر، فما انحسب عليه (ما ينظلم بصفر).</div>
      )}
      {c.evidence.length === 0 ? (
        <div className="mt-3 text-sm text-slate-500">ماكو سطور تفصيلية — {c.detail}.</div>
      ) : (
        <table className="mt-3 w-full text-sm">
          <thead>
            <tr className="border-b border-slate-200 text-right text-xs text-slate-500">
              <th className="py-1.5 font-semibold">البند</th>
              <th className="py-1.5 font-semibold">القيمة</th>
              <th className="py-1.5 font-semibold">ملاحظة</th>
            </tr>
          </thead>
          <tbody>
            {c.evidence.map((e, i) => (
              <tr key={i} className="border-b border-slate-100 last:border-0">
                <td className="py-1.5 text-slate-800">{e.label}</td>
                <td className="py-1.5 font-semibold text-slate-800">{e.value}</td>
                <td className="py-1.5 text-xs text-slate-500">{e.note ?? ''}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}

function PersonCard({ r, deputy, unit }: { r: Row; deputy: boolean; unit?: boolean }) {
  const [open, setOpen] = useState<string | null>(null)
  const openC = r.criteria.find((c) => c.key === open)
  return (
    <div className="rounded-2xl border border-slate-200 bg-white p-4 shadow-[0_4px_20px_rgba(15,32,64,0.05)]">
      <div className="flex flex-wrap items-center gap-3">
        <span className="min-w-[40px] text-2xl">{r.rank ? medal(r.rank) : '—'}</span>
        <div className="min-w-[160px] flex-1">
          <b className="text-slate-800">{r.employeeName}</b>
          {r.acting && <span className="ms-2 text-xs text-amber-600">بالوكالة</span>}
          <div className="text-xs text-slate-500">
            {unit ? r.divisionName : deputy ? `نائب ${r.divisionName}` : `مدير ${r.divisionName} · ${r.people} موظف`}
          </div>
        </div>
        {deputy && (
          <div className="text-xs text-slate-500">
            الإدارة <b className={tone(r.divisionTotal)}>{r.divisionTotal ?? '—'}</b> · شغله هو{' '}
            <b className={tone(r.ownTotal)}>{r.ownTotal ?? '—'}</b>
          </div>
        )}
        <span className={`text-3xl font-bold ${tone(r.total)}`}>{r.total === null ? 'ماكو بيانات' : r.total}</span>
      </div>
      <div className="mt-3 grid gap-2 [grid-template-columns:repeat(auto-fit,minmax(190px,1fr))]">
        {r.criteria.map((c) => (
          <CriterionTile key={c.key} c={c} open={open === c.key} onClick={() => setOpen(open === c.key ? null : c.key)} />
        ))}
      </div>
      {openC && <EvidencePanel c={openC} />}
    </div>
  )
}

export default function DirectorsRankingPage() {
  const [period, setPeriod] = useState(thisMonth())
  const [tab, setTab] = useState<'directors' | 'deputies' | 'units'>('directors')
  const [data, setData] = useState<Ranking | null>(null)
  const [error, setError] = useState('')
  useEffect(() => {
    let alive = true
    apiRequest<Ranking>(`/org/directors-ranking?period=${period}`)
      .then((d) => alive && setData(d))
      .catch((e: Error) => alive && setError(e.message))
    return () => {
      alive = false
    }
  }, [period])
  const rows = (tab === 'directors' ? data?.rows : tab === 'deputies' ? data?.deputies : data?.units) ?? []
  return (
    <div className="w-full px-4 py-4 md:px-6">
      <div className="flex flex-wrap items-center gap-3">
        <h1 className="m-0 text-xl font-bold text-slate-800">🏆 تصنيف القيادة</h1>
        <input
          type="month"
          value={period}
          className="rounded-lg border border-slate-300 px-2 py-1 text-sm"
          onChange={(e) => {
            if (!e.target.value) return
            setData(null)
            setError('')
            setPeriod(e.target.value)
          }}
        />
        <div className="flex gap-1 rounded-xl bg-slate-100 p-1">
          {(
            [
              ['directors', `المدراء (${data?.rows.length ?? 0})`],
              ['deputies', `النواب (${data?.deputies?.length ?? 0})`],
              ['units', `مدراء ونواب الوحدات (${data?.units?.length ?? 0})`],
            ] as const
          ).map(([k, l]) => (
            <button
              key={k}
              type="button"
              onClick={() => setTab(k)}
              className={`rounded-lg px-4 py-1.5 text-sm font-semibold ${tab === k ? 'bg-brand-600 text-white shadow' : 'text-slate-600'}`}
            >
              {l}
            </button>
          ))}
        </div>
      </div>
      <p className="mt-2 text-sm text-slate-500">
        {tab === 'directors'
          ? 'مدير كل إدارة مقابل الباقين، حسب: المخرجات ٣٠٪ · إنجاز الخطة ٢٠٪ · السرعة ١٥٪ · شكد انحل منها ١٥٪ · المشاكل والشغل الفايت ١٠٪ · الشكاوي والتأخير ١٠٪.'
          : tab === 'deputies'
            ? 'النائب: نص الرقم من إدارته (٥٠٪)، ونصه من شغله هو: تقييمه الشهري ٣٠٪ وبنود الخطة الي برقبته ٢٠٪.'
            : 'مدير ونائب كل وحدة على خطة تحسين وحدتهم: بنوده بالخطة ٥٠٪ (خلص بموعده = ١٠٠، متأخر = −١٠ لكل يوم) · الخطة كاملة ٢٠٪ · تقييمه الشهري ٣٠٪.'}{' '}
        المعيار الي ماله بيانات ما ينحسب عليه. الأدمن والمالك ما ينقيّمون. <b>اضغط على أي معيار حتى تشوف على شنو اعتمد.</b>
      </p>
      {error && <div className="error">{error}</div>}
      {!data && !error && <div className="mt-4 text-slate-400">جاري التحميل…</div>}
      {data && rows.length === 0 && (
        <div className="mt-4 text-slate-500">{tab === 'directors' ? 'ماكو مدراء إدارات معيّنين بالهيكل بعد.' : tab === 'deputies' ? 'ماكو نواب إدارات معيّنين بالهيكل بعد.' : 'ماكو مدراء أو نواب وحدات معيّنين بالهيكل بعد.'}</div>
      )}
      <div className="mt-4 grid gap-3">
        {rows.map((r) => (
          <PersonCard key={`${tab}-${r.divisionName}-${r.employeeName}`} r={r} deputy={tab === 'deputies'} unit={tab === 'units'} />
        ))}
      </div>
    </div>
  )
}
