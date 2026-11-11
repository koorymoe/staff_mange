import { Fragment, useEffect, useMemo, useState } from 'react'
import { api, apiRequest } from '../api'
import type { LateExcuse, StaffScore } from '../apiTypes'
import DisciplineBox from '../components/DisciplineBox'

// ═══ 🔍 مراقبة التقييم (دازاي 11-11) ═══
// «اريد اعرف شلون النظام جاي يحسب… المدراء ماعدهم مشاكل ع شنو يعتمد؟»
// كل رقم ويّاه أدلته، ونسبة تغطية البيانات، وتنبيهات للي ما يتصدّگ:
// رقم ضعيف الأدلة، إدارة ما بلّغت والنظام شايف، قفزة كبيرة، مدير عالي وكادره متأخر،
// وأعذار التأخير منو وافق عليها (حتى تبين المحاباة).

type Criterion = { key: string; label: string; weight: number; score: number | null; detail: string }
type DRow = { divisionName: string; employeeName: string; total: number | null; people: number; criteria: Criterion[]; coverage?: number; weakEvidence?: boolean; unreported?: boolean }
type Ranking = { rows: DRow[]; deputies: DRow[]; units?: DRow[] }
type Audit = { approvers: { name: string; approved: number; rejected: number }[]; excuses: LateExcuse[] }

const thisMonth = () => new Date().toISOString().slice(0, 7)
const f = (v: number | null | undefined) => (v === null || v === undefined ? '—' : Math.round(v * 10) / 10)
const tone = (v: number | null | undefined) => (v == null ? 'text-slate-400' : v >= 75 ? 'text-green-600' : v >= 50 ? 'text-amber-600' : 'text-red-600')

export default function ScoreMonitorPage() {
  const [month, setMonth] = useState(thisMonth())
  const [staff, setStaff] = useState<StaffScore[] | null>(null)
  const [rank, setRank] = useState<Ranking | null>(null)
  const [audit, setAudit] = useState<Audit | null>(null)
  const [open, setOpen] = useState<string | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => {
    api.getStaffScoreBoard(month).then((b) => { setStaff(b.staff); setErr('') }).catch((e: Error) => setErr(e.message))
    apiRequest<Ranking>(`/org/directors-ranking?period=${month}`).then(setRank).catch(() => setRank(null))
    apiRequest<Audit>(`/late-excuses/audit?month=${month}`).then(setAudit).catch(() => setAudit(null))
  }, [month])

  const leaders = useMemo(() => [...(rank?.rows ?? []), ...(rank?.deputies ?? []), ...(rank?.units ?? [])], [rank])
  const alerts = useMemo(() => {
    const out: { t: string; sev: 'red' | 'amber' }[] = []
    for (const r of rank?.rows ?? []) {
      const disc = r.criteria.find((c) => c.key === 'discipline')?.score
      if (r.unreported) out.push({ sev: 'red', t: `«${r.divisionName}» ما بلّغت ولا مشكلة هالشهر، بس النظام شايف محطات فاتت (${r.criteria.find((c) => c.key === 'few')?.detail ?? ''})` })
      if (r.weakEvidence) out.push({ sev: 'amber', t: `رقم ${r.employeeName} (${r.divisionName}) ضعيف الأدلة: ${Math.round((r.coverage ?? 0) * 100)}٪ بس من المعايير عليها بيانات` })
      if (r.total != null && r.total >= 80 && disc != null && disc < 70) out.push({ sev: 'red', t: `${r.employeeName} درجته ${f(r.total)} بس انضباط كادره ${f(disc)}` })
    }
    for (const s of staff ?? []) {
      if (s.final != null && s.prevFinal != null && Math.abs(s.final - s.prevFinal) >= 15) out.push({ sev: 'amber', t: `${s.name}: قفزة ${s.final > s.prevFinal ? '⬆️' : '⬇️'} من ${f(s.prevFinal)} لـ ${f(s.final)}` })
      if (s.discipline && s.discipline.late >= 3) out.push({ sev: 'red', t: `${s.name}: ${s.discipline.late} تأخير بلا عذر (${s.discipline.lateMinutes} دقيقة) و${s.discipline.absent} غياب` })
    }
    for (const a of audit?.approvers ?? []) {
      if (a.approved >= 5 && a.rejected === 0) out.push({ sev: 'amber', t: `${a.name} وافق على ${a.approved} عذر وما رفض ولا واحد — تأكد مو محاباة` })
    }
    return out
  }, [rank, staff, audit])

  return (
    <div dir="rtl" className="space-y-4 p-4">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h1 className="text-2xl font-black text-[#0f2040]">🔍 مراقبة التقييم</h1>
          <p className="text-xs text-slate-500">كل رقم بأدلته، وشكد بياناته حقيقية، والي ما يتصدّگ</p>
        </div>
        <input id="mon-month" type="month" value={month} onChange={(e) => setMonth(e.target.value)} className="rounded-lg border border-slate-300 px-2 py-1 text-sm" />
      </div>
      {err && <p className="rounded-xl bg-red-50 p-2 text-sm text-red-700">{err}</p>}

      <section className="space-y-1 rounded-2xl bg-white p-4 shadow-sm">
        <h2 className="text-base font-extrabold">🚨 تنبيهات ({alerts.length})</h2>
        {alerts.length === 0 ? <p className="text-sm text-slate-500">ماكو شي غريب هالشهر 👌</p> : alerts.map((a, i) => (
          <p key={i} className={`rounded-lg p-2 text-sm ${a.sev === 'red' ? 'bg-red-50 text-red-800' : 'bg-amber-50 text-amber-900'}`}>{a.t}</p>
        ))}
      </section>

      <section className="overflow-x-auto rounded-2xl bg-white p-4 shadow-sm">
        <h2 className="mb-2 text-base font-extrabold">🏛️ المدراء — منين جاي الرقم</h2>
        <table className="w-full min-w-[720px] text-sm">
          <thead><tr className="text-right text-xs text-slate-500"><th className="p-1">المدير</th><th>الإدارة</th><th>الرقم</th><th>تغطية البيانات</th>{rank?.rows[0]?.criteria.map((c) => <th key={c.key} className="p-1">{c.label} <small>({c.weight})</small></th>)}</tr></thead>
          <tbody>
            {(rank?.rows ?? []).map((r) => (
              <tr key={r.divisionName + r.employeeName} className="border-t border-slate-100">
                <td className="p-1 font-bold">{r.employeeName}{r.unreported && ' 🚩'}</td>
                <td>{r.divisionName}</td>
                <td className={`font-black ${tone(r.total)}`}>{f(r.total)}</td>
                <td className={r.weakEvidence ? 'font-bold text-amber-700' : ''}>{Math.round((r.coverage ?? 0) * 100)}٪{r.weakEvidence && ' ⚠️'}</td>
                {r.criteria.map((c) => <td key={c.key} title={c.detail} className={tone(c.score)}>{f(c.score)}</td>)}
              </tr>
            ))}
          </tbody>
        </table>
        <p className="mt-1 text-[11px] text-slate-500">حط المؤشر على أي رقم حتى تشوف سببه. التفصيل الكامل بأدلته بصفحة «تصنيف المدراء». {leaders.length > (rank?.rows.length ?? 0) && `النواب ورؤساء الوحدات (${leaders.length - (rank?.rows.length ?? 0)}) هناك هم.`}</p>
      </section>

      <section className="overflow-x-auto rounded-2xl bg-white p-4 shadow-sm">
        <h2 className="mb-2 text-base font-extrabold">👥 الموظفين — تفصيل كل تقييم</h2>
        <table className="w-full min-w-[720px] text-sm">
          <thead><tr className="text-right text-xs text-slate-500"><th className="p-1">الموظف</th><th>النهائي</th><th>الشهر الي فات</th><th>ماتركس</th><th>البشر</th><th>الانضباط</th><th>تأخير بلا عذر</th><th>غياب</th><th>الاعتمادية</th><th /></tr></thead>
          <tbody>
            {(staff ?? []).map((s) => (
              <Fragment key={s.id}>
                <tr className="border-t border-slate-100">
                  <td className="p-1 font-bold">{s.name} <small className="font-normal text-slate-500">{s.groupLabel}</small></td>
                  <td className={`font-black ${tone(s.final)}`}>{f(s.final)}</td>
                  <td>{f(s.prevFinal)}</td>
                  <td>{f(s.matrixPct)}</td>
                  <td>{s.humanCount ? f(s.humanPct) : <span className="text-amber-700">محد قيّمه</span>}</td>
                  <td className={tone(s.discipline?.score)}>{s.discipline ? f(s.discipline.score) : '—'}</td>
                  <td className={s.discipline?.late ? 'font-bold text-red-700' : ''}>{s.discipline?.late ?? '—'}</td>
                  <td className={s.discipline?.absent ? 'font-bold text-rose-800' : ''}>{s.discipline?.absent ?? '—'}</td>
                  <td>{f(s.reliability)}</td>
                  <td>{s.discipline && s.discipline.items.length > 0 && <button type="button" onClick={() => setOpen(open === s.id ? null : s.id)} className="rounded bg-slate-100 px-2 py-0.5 text-xs">الأيام</button>}</td>
                </tr>
                {open === s.id && s.discipline && <tr><td colSpan={10} className="p-2"><DisciplineBox d={s.discipline} /></td></tr>}
              </Fragment>
            ))}
          </tbody>
        </table>
        <p className="mt-1 text-[11px] text-slate-500">النهائي = ماتركس ٦٠٪ + البشر ٤٠٪، ومن أول الشهر الجاي ٨٥٪ منها + الانضباط ١٥٪.</p>
      </section>

      <section className="space-y-2 rounded-2xl bg-white p-4 shadow-sm">
        <h2 className="text-base font-extrabold">⏰ أعذار التأخير — منو وافق على شنو</h2>
        {!audit || audit.excuses.length === 0 ? <p className="text-sm text-slate-500">ماكو أعذار هالشهر</p> : (
          <>
            <div className="flex flex-wrap gap-2 text-sm">{audit.approvers.map((a) => <span key={a.name} className="rounded-full bg-slate-100 px-3 py-1">{a.name}: قبل {a.approved} · رفض {a.rejected}</span>)}</div>
            <ul className="space-y-1 text-sm">{audit.excuses.map((x) => (
              <li key={x.id}><b>{x.employeeName}</b> · {x.date.slice(0, 10)} · {x.reason} · <span className="text-slate-600">{x.status === 'PENDING' ? 'بانتظار' : `${x.status === 'APPROVED' ? 'قبله' : 'رفضه'} ${x.decidedBy ?? ''}`}</span></li>
            ))}</ul>
          </>
        )}
      </section>
    </div>
  )
}
