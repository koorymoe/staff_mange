import { useEffect, useState } from 'react'
import { thisPeriod } from '../utils/format'
import { apiRequest } from '../api'
import { useSession } from '../session'
import { notify } from '../ui/notify'
import { fmtDateTime } from '../utils/date'

// ═══ 💰 توزيع الفلوس — المالك وحده (ل) 11-24 دازاي ═══
// «ضيفها بالإعدادات… اني احدد كل موظف شكد ياخذ إذا ليدر شكد وإذا فني شكد وإذا أكثر من فني والشركة شكد».
// جدولين: الإحصائيات (شكد دخّل كل موظف — بالتقييم والاستبيانات) والإيراد (صافي ربح الشركة).

type Split = {
  statsLeaderPct: number; statsTechPct: number; statsNoLeaderTechPct: number
  revenueLeaderPct: number; revenueTechPct: number; revenueTechEach: boolean
  updatedByName?: string | null; updatedAt?: string | null
}
type Profit = { invoices: number; total: number; leaders: number; techs: number; company: number }
type Resp = { split: Split; month: string; profit: Profit }

const iqd = (v: number) => `${Math.round(v).toLocaleString('en-US')} د.ع`
const thisMonth = thisPeriod // (11-27) فترة الشركة ٢٦ ← ٢٥

function Pct({ label, hint, value, onChange }: { label: string; hint?: string; value: number; onChange: (v: number) => void }) {
  return (
    <label className="flex items-center justify-between gap-3 rounded-xl bg-slate-50 px-4 py-3">
      <span><b className="text-slate-800">{label}</b>{hint && <small className="block text-xs text-slate-500">{hint}</small>}</span>
      <span className="flex items-center gap-1">
        <input type="number" min={0} max={100} step={1} value={value} onChange={(e) => onChange(Number(e.target.value))}
          className="w-20 rounded-lg border border-slate-300 p-2 text-center text-lg font-bold" dir="ltr" />
        <b className="text-slate-500">٪</b>
      </span>
    </label>
  )
}

export default function OwnerMoneySplitPage() {
  const { employee } = useSession()
  const isOwner = employee?.actualRole === 'OWNER'
  const [month, setMonth] = useState(thisMonth())
  const [data, setData] = useState<Resp | null>(null)
  const [s, setS] = useState<Split | null>(null)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  useEffect(() => {
    if (!isOwner) return
    apiRequest<Resp>(`/owner/money-split?month=${month}`).then((d) => { setData(d); setS((x) => x ?? d.split) }).catch((e: Error) => setErr(e.message))
  }, [isOwner, month])
  if (!isOwner) return <p className="py-10 text-center text-slate-500">الصفحة غير موجودة</p>
  if (err) return <p role="alert" className="text-rose-700">{err}</p>
  if (!data || !s) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  const set = (k: keyof Split) => (v: number | boolean) => setS({ ...s, [k]: v })
  const company = 100 - s.revenueLeaderPct - s.revenueTechPct
  const save = async () => {
    setBusy(true)
    try {
      const saved = await apiRequest<Split>('/owner/money-split', { method: 'PUT', body: JSON.stringify(s) })
      setS(saved)
      setData(await apiRequest<Resp>(`/owner/money-split?month=${month}`))
      notify('انحفظت النسب ✅', 'success')
    } catch (e) { notify(e instanceof Error ? e.message : 'تعذر الحفظ') } finally { setBusy(false) }
  }
  const p = data.profit
  return (
    <div className="mx-auto max-w-4xl space-y-4" data-testid="money-split">
      <header>
        <h1 className="text-2xl font-extrabold text-[#0f2040]">💰 توزيع الفلوس</h1>
        <p className="text-sm text-slate-500">النسب تتطبق على كل فواتير الحجوزات (الصافي). المسودة والملغية ما تنحسب.</p>
      </header>

      <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
        <h2 className="text-lg font-bold text-slate-800">📊 الإحصائيات — شكد دخّل كل موظف</h2>
        <p className="text-xs text-slate-500">تطلع بتقييم الموظف («الفلوس الي دخّلها») وباستبيانات الكوادر.</p>
        <Pct label="الليدر" value={s.statsLeaderPct} onChange={set('statsLeaderPct')} />
        <Pct label="الفنيين" hint="يتقاسمونها بيناتهم (فنيين اثنين ← كل واحد النص)" value={s.statsTechPct} onChange={set('statsTechPct')} />
        <Pct label="الفني بحجز بلا ليدر" hint="مثل الـGPS والداش كام — يتقاسمونها إذا أكثر من فني" value={s.statsNoLeaderTechPct} onChange={set('statsNoLeaderTechPct')} />
        <p className="rounded-xl bg-amber-50 px-4 py-2 text-sm text-amber-900">مسؤول الخدمة ما ينحسبله شي.</p>
      </section>

      <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
        <h2 className="text-lg font-bold text-slate-800">🏢 الإيراد — صافي ربح الشركة</h2>
        <Pct label="الليدر" value={s.revenueLeaderPct} onChange={set('revenueLeaderPct')} />
        <Pct label="الفني" value={s.revenueTechPct} onChange={set('revenueTechPct')} />
        <div className="flex flex-wrap gap-2 px-1 text-sm">
          <span className="font-bold text-slate-700">إذا أكثر من فني:</span>
          {([[false, 'يتقاسمون النسبة'], [true, 'كل فني ياخذ النسبة كاملة']] as const).map(([v, l]) => (
            <button key={l} type="button" onClick={() => set('revenueTechEach')(v)}
              className={`rounded-full px-3 py-1 font-bold ${s.revenueTechEach === v ? 'bg-[#0f2040] text-white' : 'bg-slate-100 text-slate-600'}`}>{l}</button>
          ))}
        </div>
        <p className={`rounded-xl px-4 py-3 text-lg font-bold ${company < 0 ? 'bg-rose-50 text-rose-700' : 'bg-emerald-50 text-emerald-800'}`}>
          الشركة: {company}٪ <small className="text-xs font-normal">(الباقي{s.revenueTechEach ? ' — يقل كلما زاد عدد الفنيين' : ''})</small>
        </p>
      </section>

      <div className="flex items-center gap-3">
        <button type="button" disabled={busy || company < 0} onClick={() => void save()} className="rounded-xl bg-[#0f2040] px-6 py-2.5 font-bold text-white disabled:opacity-50">{busy ? 'يحفظ…' : 'احفظ النسب'}</button>
        {data.split.updatedAt && <span className="text-xs text-slate-500">آخر تعديل: {data.split.updatedByName ?? '—'} · {fmtDateTime(data.split.updatedAt)}</span>}
      </div>

      <section className="space-y-3 rounded-2xl bg-white p-4 shadow-sm">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <h2 className="text-lg font-bold text-slate-800">📈 صافي ربح الشركة</h2>
          <input type="month" value={month} onChange={(e) => e.target.value && setMonth(e.target.value)} className="rounded-lg border border-slate-300 p-2" dir="ltr" />
        </div>
        <div className="grid gap-2 sm:grid-cols-4">
          {([['صافي الفواتير', p.total, `${p.invoices} فاتورة`], ['حصة الليدرية', p.leaders, ''], ['حصة الفنيين', p.techs, ''], ['ربح الشركة', p.company, p.total ? `${Math.round((p.company / p.total) * 100)}٪ من الصافي` : '']] as const).map(([k, v, sub], i) => (
            <div key={k} className={`rounded-xl p-3 ${i === 3 ? 'bg-emerald-50 ring-1 ring-emerald-200' : 'bg-slate-50'}`}>
              <p className="text-xs text-slate-500">{k}</p>
              <p className={`text-lg font-extrabold ${i === 3 ? 'text-emerald-800' : 'text-slate-800'}`} dir="ltr">{iqd(v)}</p>
              {sub && <p className="text-xs text-slate-500">{sub}</p>}
            </div>
          ))}
        </div>
        <p className="text-xs text-slate-400">الحساب بالنسب المحفوظة. الفاتورة البلا ليدر ما ينطلع منها حصة ليدر؛ وحصة مسؤول الخدمة تروح للشركة.</p>
      </section>
    </div>
  )
}
