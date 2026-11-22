import { useEffect, useState } from 'react'
import { api, type RevisitReport } from '../api'
import BookingCodeChip from '../components/BookingCodeChip'
import PctWhy from '../ui/PctWhy'

// ═══ 🔁 الرجعات — (ع) 11-07 ═══
// حجز ثاني لنفس الزبون ونفس الخدمة خلال ٣٠ يوم من الإنجاز = الشغل الأول ما انعمل
// صح. أهم مقياس لجودة الفني. ⚠️ للمتابعة والتدريب بس — ماكو أي خصم يطلع منه.
const RANGES = [30, 90, 180]

// (ك) 11-22 زينا: تصميم الرجعات — بطاقات ملوّنة، بحث وتصفية وتصدير، وجدول بنفس حجم جداول تقارير الموظفين.
const ICONS: Record<string, string> = {
  user: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8Zm-7 8a7 7 0 0 1 14 0',
  check: 'M12 21a9 9 0 1 0 0-18 9 9 0 0 0 0 18Zm-4-9 3 3 5-6',
  undo: 'M9 14 4 9l5-5M4 9h10a6 6 0 0 1 0 12h-3',
  bars: 'M6 20V14m6 6V8m6 12V4',
  star: 'M12 3l2.7 5.6 6.1.9-4.4 4.3 1 6.1L12 17l-5.4 2.9 1-6.1L3.2 9.5l6.1-.9L12 3Z',
  cal: 'M7 3v4m10-4v4M4 9h16M5 5h14v15H5z',
  search: 'M11 18a7 7 0 1 0 0-14 7 7 0 0 0 0 14Zm10 3-5-5',
  filter: 'M3 5h18l-7 8v6l-4 2v-8L3 5Z',
  down: 'M12 4v11m-5-5 5 5 5-5M5 20h14',
  clip: 'M9 4h6v3H9zM7 5H5v16h14V5h-2M9 13l2 2 4-4',
  warn: 'M12 3 2 20h20L12 3Zm0 6v5m0 3v.5',
  eq: 'M6 9h12M6 15h12',
}
const Ico = ({ n, className = 'h-4 w-4' }: { n: string; className?: string }) => (
  <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" className={className} aria-hidden="true"><path d={ICONS[n]} /></svg>
)

// (ك) 11-22 زينا: الحدود مثل التصميم — ٢٠٪ وفوك يحتاج متابعة، ١٠–٢٠ متوسط، أقل جيد، وصفر رجعة ممتاز.
type Level = { key: string; label: string; icon: string; pill: string; text: string; bar: string }
const LEVELS: Level[] = [
  { key: 'bad', label: 'يحتاج إلى متابعة', icon: 'warn', pill: 'bg-red-50 text-red-700 ring-red-300', text: 'text-red-600', bar: 'bg-red-500' },
  { key: 'mid', label: 'متوسط', icon: 'eq', pill: 'bg-amber-50 text-amber-700 ring-amber-300', text: 'text-amber-600', bar: 'bg-amber-400' },
  { key: 'good', label: 'جيد', icon: 'check', pill: 'bg-emerald-50 text-emerald-700 ring-emerald-300', text: 'text-emerald-600', bar: 'bg-emerald-500' },
  { key: 'top', label: 'ممتاز', icon: 'star', pill: 'bg-yellow-50 text-yellow-700 ring-yellow-300', text: 'text-slate-700', bar: 'bg-yellow-400' },
]
const levelOf = (rate: number, revisits: number) => revisits === 0 ? LEVELS[3] : rate >= 20 ? LEVELS[0] : rate >= 10 ? LEVELS[1] : LEVELS[2]
// الرقم بمرتبة عشرية وحدة دائماً (50.0%) مثل التصميم
const dec = (v: number) => `${Number.isInteger(v) ? '.0' : ''}%`

export default function RevisitsPage() {
  const [days, setDays] = useState(30)
  const [rep, setRep] = useState<RevisitReport | null>(null)
  const [err, setErr] = useState('')
  const [q, setQ] = useState('')
  const [lvl, setLvl] = useState('')
  useEffect(() => {
    let alive = true
    api.getRevisits(days)
      .then((r) => { if (alive) { setRep(r); setErr('') } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر') })
    return () => { alive = false }
  }, [days])
  const rate = rep && rep.jobs ? Math.round((1000 * rep.revisits) / rep.jobs) / 10 : 0
  const rows = (rep?.leaders ?? [])
    .filter((l) => (!q.trim() || l.leaderName.includes(q.trim())) && (!lvl || levelOf(l.rate, l.revisits).key === lvl))
    .sort((x, y) => y.rate - x.rate || y.revisits - x.revisits)
  const exportCsv = () => {
    const lines = [['#', 'الليدر', 'إنجاز', 'رجوع', 'نسبة الرجعة', 'مستوى الأداء'], ...rows.map((l, i) => [i + 1, l.leaderName, l.jobs, l.revisits, `${l.rate}%`, levelOf(l.rate, l.revisits).label])]
    const csv = '\uFEFF' + lines.map((r) => r.map((c) => `"${String(c).replace(/"/g, '""')}"`).join(',')).join('\n')
    const a = document.createElement('a')
    a.href = URL.createObjectURL(new Blob([csv], { type: 'text/csv;charset=utf-8' }))
    a.download = `revisits-${days}d.csv`
    a.click()
    URL.revokeObjectURL(a.href)
  }
  const heads: [string, string][] = [['', '#'], ['user', 'الليدر'], ['check', 'إنجاز'], ['undo', 'رجوع'], ['bars', 'نسبة الرجعة'], ['star', 'مستوى الأداء']]
  const ring = Math.min(100, rate)

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex items-start gap-3">
        <span className="rv-icon flex h-16 w-16 shrink-0 items-center justify-center rounded-2xl text-4xl">🔁</span>
        <div>
          <h2 className="text-3xl font-extrabold text-brand-900">الرجعات</h2>
          <p className="mt-1 text-[15px] leading-7 text-slate-500">
            حجز رجع لنفس الزبون ونفس الخدمة خلال {rep?.windowDays ?? 30} يوم من الإنجاز — يعني الشغل الأول يمكن ما انعمل صح.
            تنحسب على ليدر الحجز الأول. للمتابعة والتدريب، مو للخصم.
          </p>
        </div>
      </div>
      <div className="flex flex-wrap items-center gap-2">
        <span className="rv-cal flex h-11 w-11 items-center justify-center rounded-xl border border-slate-200 bg-white text-slate-600"><Ico n="cal" className="h-5 w-5" /></span>
        {RANGES.map((d) => (
          <button key={d} type="button" onClick={() => { if (d !== days) { setRep(null); setDays(d) } }}
            className={`rv-range rounded-full border bg-white px-6 py-2 text-base font-bold transition ${days === d ? 'rv-range-on border-amber-400 text-[#0f2040] ring-2 ring-amber-300/60' : 'border-slate-200 text-slate-600 hover:border-slate-300'}`}>آخر {d} يوم</button>
        ))}
      </div>
      {err && <p className="text-sm text-red-600">{err}</p>}
      {!rep && !err && <p className="text-slate-400">جاري الحساب…</p>}
      {rep && (
        <>
          <div className="grid gap-5 md:grid-cols-3">
            <div className="rv-card rv-blue flex min-h-[7rem] items-center gap-4 rounded-2xl px-6 py-4">
              <span className="rv-ic flex h-14 w-14 items-center justify-center rounded-full"><Ico n="clip" className="h-7 w-7" /></span>
              <div className="flex-1"><p className="text-lg font-bold">حجوزات الإنجاز</p><p className="text-4xl font-extrabold tabular-nums">{rep.jobs}</p></div>
              <Ico n="cal" className="h-14 w-14 opacity-90" />
            </div>
            <div className="rv-card rv-amber flex min-h-[7rem] items-center gap-4 rounded-2xl px-6 py-4">
              <span className="rv-ic flex h-14 w-14 items-center justify-center rounded-full"><Ico n="undo" className="h-7 w-7" /></span>
              <div className="flex-1"><p className="text-lg font-bold">الرجعات</p><p className="text-4xl font-extrabold tabular-nums">{rep.revisits}</p></div>
              <Ico n="clip" className="h-14 w-14 opacity-90" />
            </div>
            <div className="rv-card rv-teal flex min-h-[7rem] items-center gap-4 rounded-2xl px-6 py-4">
              <span className="rv-ic flex h-14 w-14 items-center justify-center rounded-full"><Ico n="bars" className="h-7 w-7" /></span>
              <div className="flex-1"><p className="text-lg font-bold">نسبة الرجعة</p><p className="text-sm opacity-80">من إجمالي الحجوزات</p></div>
              <span className="relative flex h-24 w-24 items-center justify-center rounded-full" style={{ background: `conic-gradient(#facc15 ${ring * 3.6}deg, #0d9488 0)` }}>
                <span className="rv-ring-in flex h-[4.5rem] w-[4.5rem] items-center justify-center rounded-full text-lg font-extrabold">
                  <PctWhy title="نسبة الرجعة" value={rate} suffix={dec(rate)} parts={{ numLabel: 'حجز رجع', num: rep.revisits, denLabel: 'حجز انجز', den: rep.jobs }}
                    meaning={<>من كل 100 حجز انجز، شكد واحد رجع الزبون فتح حجز ثاني لنفس الخدمة خلال {rep.windowDays} يوم — يعني الشغل الأول يمكن ما انعمل صح. كل ما تكون أقل أحسن.</>}
                    source={<>جدول الحجوزات: كل حجز حالته «مكتمل» انجز خلال المدة المختارة (بدون الاستيراد القديم OLD- والشغل الداخلي والمحذوف). وبعدها ندور حجز ثاني لنفس الزبون ونفس الخدمة انفتح خلال {rep.windowDays} يوم من الإنجاز ومو ملغي.</>} />
                </span>
              </span>
            </div>
          </div>

          <div className="rv-panel space-y-4 rounded-3xl border border-slate-200 bg-white p-5 shadow-sm">
            <div className="flex flex-wrap items-center gap-2">
              <label className="flex min-w-[220px] max-w-md flex-1 items-center gap-2 rounded-full border border-slate-200 px-4 py-2.5 text-slate-500">
                <Ico n="search" className="h-5 w-5" />
                <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="البحث عن ليدر…" className="w-full bg-transparent text-[15px] text-slate-800 outline-none" />
              </label>
              <label className="ms-auto flex items-center gap-2 rounded-full border border-slate-200 px-5 py-2.5 text-base font-bold text-slate-700">
                <Ico n="filter" className="h-5 w-5" />
                <select value={lvl} onChange={(e) => setLvl(e.target.value)} className="bg-transparent outline-none" aria-label="تصفية">
                  <option value="">تصفية</option>
                  {LEVELS.map((l) => <option key={l.key} value={l.key}>{l.label}</option>)}
                </select>
              </label>
              <button type="button" onClick={exportCsv} disabled={!rows.length} className="flex items-center gap-2 rounded-full border border-amber-300 px-5 py-2.5 text-base font-bold text-amber-700 disabled:opacity-50">
                <Ico n="down" className="h-5 w-5" />تصدير التقرير
              </button>
            </div>
            <div className="overflow-x-auto rounded-2xl border border-slate-200" data-testid="revisits-table">
              <table className="w-full min-w-[980px] table-fixed text-[13px]">
                <colgroup>{['5%', '25%', '12%', '12%', '28%', '18%'].map((w, i) => <col key={i} style={{ width: w }} />)}</colgroup>
                <thead className="text-xs font-bold">
                  <tr>
                    {heads.map(([ic, h]) => (
                      <th key={h} className="p-0 text-center">
                        <span className="th-box rv-th flex h-[4.75rem] items-center justify-center gap-2 rounded-none border-x-0 px-1.5 text-[15px]">
                          {ic && <span className="th-box-ic"><Ico n={ic} className="h-5 w-5" /></span>}
                          <span className="leading-snug">{h}</span>
                        </span>
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200">
                  {rows.map((l, i) => {
                    const lv = levelOf(l.rate, l.revisits)
                    return (
                      <tr key={l.leaderId} className="h-[3.8rem] align-middle hover:bg-slate-50 [&>td]:border-l [&>td]:border-slate-200 [&>td:last-child]:border-l-0">
                        <td className="px-2 py-2 text-center"><span className="rv-num inline-flex h-8 w-8 items-center justify-center rounded-full bg-slate-100 text-sm font-bold text-slate-600 ring-1 ring-slate-300">{i + 1}</span></td>
                        <td className="px-2 py-2"><span className="flex items-center gap-2 text-[15px] font-bold text-[#0f2040]"><span className="shrink-0 text-amber-500"><Ico n="user" className="h-5 w-5" /></span>{l.leaderName}</span></td>
                        <td className="px-2 py-2 text-center text-[15px] font-bold tabular-nums text-slate-700">{l.jobs}</td>
                        <td className="px-2 py-2 text-center text-[15px] font-bold tabular-nums text-slate-700">{l.revisits}</td>
                        <td className="px-3 py-2">
                          <div className="flex items-center gap-2" dir="ltr">
                            <span className={`w-16 shrink-0 text-[15px] font-extrabold tabular-nums ${lv.text}`}>
                              <PctWhy title={`نسبة رجعة ${l.leaderName}`} value={l.rate} suffix={dec(l.rate)} parts={{ numLabel: 'حجز رجع', num: l.revisits, denLabel: 'حجز انجز وهو ليدره', den: l.jobs }}
                                meaning={<>من الحجوزات الي هذا الليدر كان ليدرها وانجزت، شكد منها رجع الزبون عليها خلال {rep.windowDays} يوم. 20٪ وفوك «يحتاج إلى متابعة»، من 10٪ لـ20٪ «متوسط»، أقل من 10٪ «جيد»، وصفر رجعة «ممتاز».</>}
                                source={<>جدول الحجوزات: كل حجز حالته «مكتمل» انجز خلال المدة المختارة (بدون الاستيراد القديم OLD- والشغل الداخلي والمحذوف). وبعدها ندور حجز ثاني لنفس الزبون ونفس الخدمة انفتح خلال {rep.windowDays} يوم من الإنجاز ومو ملغي.</>} />
                            </span>
                            <span className="rv-track h-2.5 flex-1 overflow-hidden rounded-full bg-slate-200"><span className={`block h-full rounded-full ${lv.bar}`} style={{ width: `${Math.min(100, l.rate)}%` }} /></span>
                          </div>
                        </td>
                        <td className="px-2 py-2 text-center">
                          <span className={`rv-pill-${lv.key} inline-flex items-center gap-1.5 whitespace-nowrap rounded-full px-4 py-1 text-sm font-bold ring-1 ${lv.pill}`}><Ico n={lv.icon} className="h-4 w-4" />{lv.label}</span>
                        </td>
                      </tr>
                    )
                  })}
                </tbody>
              </table>
            </div>
            {rows.length === 0 && <p className="py-4 text-center text-slate-400">{rep.leaders.length ? 'ماكو ليدر يطابق البحث' : 'ماكو حجوزات منجزة بهالمدة'}</p>}
          </div>

          {rep.items.length > 0 && (
            <div className="space-y-2">
              <p className="text-sm font-extrabold text-brand-900">الحجوزات الي رجعت</p>
              {rep.items.map((it) => (
                <div key={it.firstId} className="flex flex-wrap items-center gap-2 rounded-xl bg-white p-3 text-sm shadow-sm">
                  <BookingCodeChip code={it.firstCode} />
                  <span className="text-slate-400">← رجع بعد {it.gapDays} يوم →</span>
                  <BookingCodeChip code={it.secondCode} />
                  <span className="text-slate-600">{it.serviceName ?? '—'}</span>
                  <span className="text-slate-500">· {it.customerName ?? ''}</span>
                  <span className="ms-auto text-xs text-slate-500">الليدر: {it.leaderName ?? '—'}</span>
                </div>
              ))}
            </div>
          )}
        </>
      )}
    </div>
  )
}
