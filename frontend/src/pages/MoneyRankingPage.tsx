import { useEffect, useState } from 'react'
import { apiRequest } from '../api'

// ═══ 💰 الفلوس والأهداف — ترتيب الموظفين للمدراء (ل) 11-24 دازاي ═══
// «الإحصائيات: كل موظف شكد دخّل وشكد باقيله وشكد طلع عنده أخطاء وغيره صحح، وترتيب منو الأفضل ومنو الأردأ».
// بلا رواتب — الرواتب والراتب المتوقع للمالك بس بـ«تكاليف الشركة».

type Row = { rank: number; name: string; shift: 'MORNING' | 'EVENING' | null; target: number; daily: number; brought: number; remaining: number; forecast: number; pct: number; dailyNeeded: number; errors: number; fixes: number }
const iqd = (v: number) => Math.round(v).toLocaleString('en-US')
const thisMonth = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' }).slice(0, 7)

export default function MoneyRankingPage() {
  const [month, setMonth] = useState(thisMonth())
  const [d, setD] = useState<{ rows: Row[]; downtimeDays: number } | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => { apiRequest<{ rows: Row[]; downtimeDays: number }>(`/employee-stats/money-ranking?month=${month}`).then(setD).catch((e: Error) => setErr(e.message)) }, [month])
  if (err) return <p role="alert" className="text-rose-700">{err}</p>
  if (!d) return <p className="py-10 text-center text-slate-400">جاري التحميل…</p>
  return (
    <section className="space-y-3 rounded-2xl bg-white p-4 shadow-sm" data-testid="money-ranking">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h2 className="text-lg font-extrabold text-[#0f2040]">💰 الفلوس والأهداف{d.downtimeDays > 0 && <small className="ms-2 text-xs font-normal text-amber-700">({d.downtimeDays} يوم توقف محسوب)</small>}</h2>
        <input type="month" value={month} onChange={(e) => e.target.value && setMonth(e.target.value)} className="rounded-lg border border-slate-300 p-1.5 text-sm" dir="ltr" />
      </div>
      {d.rows.length === 0 ? <p className="text-sm text-slate-500">ماكو أهداف بعد — المالك يحدد الموظفين ورواتبهم بـ«تكاليف الشركة والأهداف».</p> : (
        <div className="overflow-x-auto">
          <table className="w-full min-w-[820px] text-sm">
            <thead className="bg-slate-50 text-xs text-slate-600"><tr>{['#', 'الموظف', 'الهدف', 'دخّل', 'باقي', 'لازم باليوم', 'التوقع لآخر الشهر', 'أخطاء فنية', 'صلّح لغيره'].map((h) => <th key={h} className="p-2 text-center">{h}</th>)}</tr></thead>
            <tbody className="divide-y">
              {d.rows.map((r) => (
                <tr key={r.rank} className="text-center">
                  <td className="p-2 font-bold text-slate-500">{r.rank <= 3 ? ['🥇', '🥈', '🥉'][r.rank - 1] : r.rank}</td>
                  <td className="p-2 text-start font-bold">{r.name}{r.shift && <small className="ms-1 text-xs font-normal text-slate-500">({r.shift === 'MORNING' ? 'صباحي' : 'مسائي'})</small>}</td>
                  <td dir="ltr">{iqd(r.target)}</td>
                  <td dir="ltr">
                    <b className={r.pct >= 100 ? 'text-emerald-700' : r.pct >= 60 ? 'text-amber-700' : 'text-rose-700'}>{iqd(r.brought)}</b> <small className="text-slate-500">({r.pct}٪)</small>
                    <div className="mx-auto mt-1 h-1.5 w-24 overflow-hidden rounded-full bg-slate-100"><div className={`h-full ${r.pct >= 100 ? 'bg-emerald-500' : r.pct >= 60 ? 'bg-amber-500' : 'bg-rose-500'}`} style={{ width: `${Math.min(100, Math.max(0, r.pct))}%` }} /></div>
                  </td>
                  <td dir="ltr">{iqd(r.remaining)}</td>
                  <td dir="ltr">{iqd(r.dailyNeeded || r.daily)}</td>
                  <td dir="ltr" className={r.forecast >= r.target ? 'text-emerald-700' : 'text-rose-700'}>{iqd(r.forecast)}</td>
                  <td>{r.errors > 0 ? <b className="text-rose-700">{r.errors}</b> : '—'}</td>
                  <td>{r.fixes > 0 ? <b className="text-emerald-700">{r.fixes}</b> : '—'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}
      <p className="text-xs text-slate-400">«دخّل» حسب نسب «توزيع الفلوس»، ويا تصحيحات الأخطاء الفنية الموافق عليها. الترتيب حسب نسبة الوصول للهدف.</p>
    </section>
  )
}
