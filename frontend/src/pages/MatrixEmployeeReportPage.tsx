import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { api, type EmployeeReport, type ReportLine } from '../api'
import MatrixEyeGraphic from '../components/MatrixEyeGraphic'
import { GROUP_LABEL, type EyeGroup } from '../components/matrixEyeColors'

// ═══ تقرير ماتركس عن موظف — بالعراقي، بأرقام حقيقية ═══
// دوامه مقابل جدوله، وكل حجز بمراحله ومنو أخّر، وإجازاته ونمطها،
// و«شخصيته بالشغل» من ٣٠ يوم. للمدير والمالك بس.

const todayBaghdad = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })
const TONE: Record<ReportLine['tone'], string> = {
  OK: 'border-emerald-200 bg-emerald-50 text-emerald-900',
  WARN: 'border-amber-200 bg-amber-50 text-amber-900',
  BAD: 'border-red-200 bg-red-50 text-red-900',
  INFO: 'border-slate-200 bg-slate-50 text-slate-800',
}
const ICON: Record<ReportLine['tone'], string> = { OK: '✅', WARN: '⚠️', BAD: '🔴', INFO: 'ℹ️' }

function Lines({ lines }: { lines: ReportLine[] }) {
  if (lines.length === 0) return <p className="text-xs text-slate-400">ماكو شي.</p>
  return (
    <ul className="space-y-1.5">
      {lines.map((l, i) => <li key={i} className={`rounded-lg border px-3 py-1.5 text-sm ${TONE[l.tone]}`}>{ICON[l.tone]} {l.text}</li>)}
    </ul>
  )
}

function Section({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-4">
      <h3 className="mb-2 font-extrabold text-[#0f2040]">{title}</h3>
      {children}
    </section>
  )
}

export default function MatrixEmployeeReportPage() {
  const { id = '' } = useParams()
  const [day, setDay] = useState(todayBaghdad())
  const [rep, setRep] = useState<EmployeeReport | null>(null)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    let alive = true
    api.getEmployeeReport(id, day)
      .then((r) => { if (alive) { setRep(r); setErr(null) } })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذر جلب التقرير') })
    return () => { alive = false }
  }, [id, day])

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          {rep && <MatrixEyeGraphic group={(rep.group || 'STAFF') as EyeGroup} mood="CALM" width={70} />}
          <div>
            <h2 className="text-2xl font-bold text-brand-900">تقرير ماتركس: {rep?.name ?? '…'}</h2>
            <p className="text-sm text-slate-500">{rep ? GROUP_LABEL[rep.group as EyeGroup] ?? rep.group : ''} · أرقام حقيقية من النظام، بلا تخمين.</p>
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Link to="/?board=matrix" className="text-sm text-brand-700 underline">← عيون ماتركس</Link>
          <input type="date" value={day} onChange={(e) => setDay(e.target.value)} className="rounded-lg border border-slate-300 px-3 py-1.5 text-sm" />
        </div>
      </div>

      {err && <p className="rounded-lg bg-red-50 p-3 text-red-600">{err}</p>}
      {!rep && !err && <p className="text-slate-400">ماتركس يحلّل…</p>}

      {rep && (
        <>
          <section className="rounded-2xl bg-[#0b1220] p-4 text-slate-100">
            <p className="mb-1 text-xs text-slate-400">خلاصة ماتركس {rep.summaryBy === 'MODEL' ? '(🧠 هايكو)' : '(📏 القواعد)'}</p>
            <p className="text-base leading-8">{rep.summary}</p>
          </section>

          <div className="grid gap-4 lg:grid-cols-2">
            <Section title="🕐 الدوام"><Lines lines={rep.attendance} /></Section>
            <Section title="🌴 الإجازات"><Lines lines={rep.leaves} /></Section>
          </div>

          <Section title={`🔧 الحجوزات (${rep.jobs.length})`}>
            {rep.jobs.length === 0 ? <p className="text-xs text-slate-400">ماكو حجوزات بهاليوم.</p> : (
              <div className="space-y-3">
                {rep.jobs.map((j) => (
                  <div key={j.code} className="rounded-xl border border-slate-200 p-3">
                    <p className="font-bold text-slate-800">{j.code} <span className="text-xs font-normal text-slate-500">{j.service} · {j.status}</span></p>
                    <ol className="my-2 flex flex-wrap items-center gap-1 text-[12px]">
                      {j.stages.map((st, i) => (
                        <li key={i} className="flex items-center gap-1">
                          {i > 0 && <span className={`text-[10px] ${st.gap > 45 ? 'font-bold text-red-600' : 'text-slate-400'}`}>{st.gap > 0 ? `${st.gap >= 120 ? `${Math.round(st.gap / 60)}س` : `${st.gap}د`} ←` : '←'}</span>}
                          <span className="rounded-md bg-slate-100 px-1.5 py-0.5"><b>{st.label}</b> {st.at}</span>
                        </li>
                      ))}
                    </ol>
                    <Lines lines={j.lines} />
                  </div>
                ))}
              </div>
            )}
          </Section>

          <div className="grid gap-4 lg:grid-cols-2">
            <Section title="🧭 شخصيته بالشغل (آخر ٣٠ يوم)"><Lines lines={rep.behavior} /></Section>
            <Section title="🤖 تذكيرات ماتركس"><Lines lines={rep.reminders} /></Section>
          </div>

          {rep.workload && rep.workload.length > 0 && (
            <Section title="📋 شغله اليوم">
              <ul className="space-y-1 text-sm">
                {rep.workload.map((w) => <li key={w.key}>{w.label}: {w.done > 0 && <span className="text-emerald-700">{w.verb} {w.done} · </span>}<b className={w.left ? 'text-amber-700' : 'text-emerald-700'}>{w.left ? `باقي ${w.left}` : 'خالص'}</b></li>)}
              </ul>
            </Section>
          )}
        </>
      )}
    </div>
  )
}
