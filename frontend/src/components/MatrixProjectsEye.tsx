import { useEffect, useState } from 'react'
import { api, type ChainLearning, type ProjectChain, type ProjectsReport } from '../api'
import MatrixNote from './MatrixNote'
import { CHAIN_TONE } from './chainTone'

// ═══ ماتركس ٢٠٥٠ — عين على المشاريع + «شنو تعلّم ماتركس» ═══
// أي أحد يتوجهله مشروع (مشرف، تقني، مصمم) عليه رقابة: كم يوم بالمرحلة
// مقابل المعتاد، آخر حركة، ومنو لمسه وبأي ترتيب.

const dt = (s: string) => new Date(s).toLocaleString('en-GB', { day: '2-digit', month: '2-digit', hour: '2-digit', minute: '2-digit' })
const fm = (m: number) => (m < 60 ? `${m}د` : m < 1440 ? `${Math.round(m / 6) / 10}س` : `${Math.round(m / 144) / 10}ي`)

export function MatrixProjectsEye() {
  const [rep, setRep] = useState<ProjectsReport | null | undefined>(undefined)
  const [open, setOpen] = useState<string | null>(null)
  const [chain, setChain] = useState<{ id: string; c: ProjectChain | null } | null>(null)
  const [showDone, setShowDone] = useState(false)

  useEffect(() => { void api.getMatrixProjects().then(setRep).catch(() => setRep(null)) }, [])
  useEffect(() => {
    if (!open) return
    let alive = true
    void api.getMatrixProject(open).catch(() => null).then((c) => { if (alive) setChain({ id: open, c }) })
    return () => { alive = false }
  }, [open])

  if (rep === undefined) return <p className="text-xs text-slate-400">ماتركس يحلّل المشاريع…</p>
  if (!rep) return null
  const list = rep.projects.filter((p) => showDone || p.status !== 'NA')

  return (
    <div dir="rtl" className="space-y-3 rounded-2xl border border-slate-200 bg-white p-3 text-slate-800">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <p className="text-sm font-extrabold text-[#0f2040]">🏗️ عين ماتركس على المشاريع <span className="text-[11px] font-normal text-slate-500">أي أحد يتوجهله مشروع، بغض النظر عن دوره</span></p>
        <label className="flex items-center gap-1 text-xs"><input type="checkbox" checked={showDone} onChange={(e) => setShowDone(e.target.checked)} /> اعرض المكتملة والمرفوضة</label>
      </div>
      <MatrixNote>{rep.insights.join(' ')}</MatrixNote>
      {rep.people.length > 0 && (
        <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-4">
          {rep.people.map((p) => (
            <div key={p.id} className={`rounded-xl border p-2 text-xs ${p.stuck ? 'border-orange-200 bg-orange-50' : 'border-slate-200 bg-slate-50'}`}>
              <b className="block text-sm">{p.name}</b>
              <span>مفتوح {p.open} · عالق {p.stuck} · مكتمل {p.done} · سكون {p.avgIdle} يوم</span>
              {p.insights.length > 0 && <p className="mt-1 text-slate-600">{p.insights.join(' ')}</p>}
            </div>
          ))}
        </div>
      )}
      <div className="space-y-2">
        {list.map((p) => {
          const t = CHAIN_TONE[p.status]
          return (
            <div key={p.id} className={`rounded-xl border p-2.5 ${t.cls}`}>
              <button type="button" onClick={() => setOpen(open === p.id ? null : p.id)} className="flex w-full flex-wrap items-center justify-between gap-2 text-right">
                <span className="text-sm font-bold">{t.icon} {p.code} — {p.name} <span className="text-[11px] font-normal text-slate-500">· {p.stage}</span></span>
                <span className="text-[11px] text-slate-600">{p.ownerName || 'بلا مسؤول'} · {p.daysInStage}/{p.stageLimit} يوم · آخر حركة قبل {p.idleDays} يوم</span>
              </button>
              <p className="mt-1 text-xs text-slate-700">{p.verdict}</p>
              {open === p.id && (chain?.id !== p.id ? <p className="text-xs text-slate-400">…</p> : chain.c && (
                <div className="mt-2 space-y-2 rounded-lg bg-white/70 p-2 text-xs">
                  <div><b>📈 ترتيب العمل:</b>{chain.c.order.length ? chain.c.order.map((o, i) => <p key={i}>{i + 1}. {o}</p>) : <p className="text-slate-500">ماكو تغيير مراحل مسجّل بعد (السجل بدا من هسه).</p>}</div>
                  {chain.c.delegations.length > 0 && <div><b>📨 التوجيه:</b>{chain.c.delegations.map((d, i) => <p key={i}>{dt(d.at)} — {d.action === 'REVOKE' ? 'انسحب من' : 'توجّه لـ'} {d.name}{d.by ? ` (بيد ${d.by})` : ''}</p>)}</div>}
                  <div><b>👥 منو اشتغل عليه:</b>{chain.c.touchers.length ? chain.c.touchers.map((x) => <p key={x.employeeId}>{x.name} — أول مرة {dt(x.firstAt)}، آخر مرة {dt(x.lastAt)} · {x.actions} حفظ</p>) : <p className="text-slate-500">ماكو أحد حفظ عليه شي من بدا سجل النشاط.</p>}</div>
                </div>
              ))}
            </div>
          )
        })}
        {list.length === 0 && <p className="text-center text-xs text-slate-400">ماكو مشاريع مفتوحة.</p>}
      </div>
    </div>
  )
}

export function MatrixLearning() {
  const [l, setL] = useState<ChainLearning | null>(null)
  useEffect(() => { void api.getChainLearning().then(setL).catch(() => {}) }, [])
  if (!l) return null
  return (
    <details dir="rtl" className="rounded-2xl border border-slate-200 bg-white p-3 text-slate-800">
      <summary className="cursor-pointer text-sm font-extrabold text-[#0f2040]">🧠 شنو تعلّم ماتركس من شغل الشركة <span className="text-[11px] font-normal text-slate-500">الحدود يحسبها لحاله كل ساعة من آخر ٩٠ يوم</span></summary>
      <MatrixNote className="my-2">ما أحط أرقام من راسي — كل حد «متأخر» تعلّمته من شغلكم الحقيقي. {l.servicesOwnLimits > 0 ? `و${l.servicesOwnLimits} خدمة صار إلها حدودها الخاصة لأن عندها عينات كافية.` : 'وكل ما تزيد الحجوزات، كل خدمة تاخذ حدودها الخاصة.'}</MatrixNote>
      <div className="overflow-x-auto">
        <table className="w-full min-w-[520px] text-right text-xs">
          <thead className="bg-slate-100 text-slate-600"><tr><th className="p-2">المحطة</th><th className="p-2">عينات</th><th className="p-2">الوسيط</th><th className="p-2">حد «متأخر»</th><th className="p-2">منين</th></tr></thead>
          <tbody>{l.stations.map((s) => (
            <tr key={s.key} className="border-t border-slate-100"><td className="p-2 font-bold">{s.title}</td><td className="p-2">{s.samples}</td><td className="p-2">{s.samples ? fm(s.medianMin) : '—'}</td><td className="p-2">{fm(s.limitMin)}</td><td className="p-2 text-slate-500">{s.source}</td></tr>
          ))}</tbody>
        </table>
      </div>
    </details>
  )
}
