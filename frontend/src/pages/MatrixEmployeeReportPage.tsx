import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { api, type EmployeeReport, type ReportLine } from '../api'
import MatrixEyeGraphic from '../components/MatrixEyeGraphic'
import { GROUP_LABEL, type EyeGroup } from '../components/matrixEyeColors'
import { useSession } from '../session'

// ═══ تقرير ماتركس عن موظف — بالعراقي، بأرقام حقيقية ═══
// دوامه مقابل جدوله، وكل حجز بمراحله ومنو أخّر، وإجازاته ونمطها،
// و«شخصيته بالشغل» من ٣٠ يوم. للمدير والمالك، والمراقب يقراه (بلا أزرار الإدارة).

const todayBaghdad = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })
const TONE: Record<ReportLine['tone'], string> = {
  OK: 'border-emerald-200 bg-emerald-50 text-emerald-900',
  WARN: 'border-amber-200 bg-amber-50 text-amber-900',
  BAD: 'border-red-200 bg-red-50 text-red-900',
  INFO: 'border-slate-200 bg-slate-50 text-slate-800',
}
const ICON: Record<ReportLine['tone'], string> = { OK: '✅', WARN: '⚠️', BAD: '🔴', INFO: 'ℹ️' }

// كل مشكلة: «ودّيني» + حلول مقترحة، وزر يحوّل الحل تعليمة لماتركس
// (تنحفظ معطّلة — المدير يفعّلها من تعليمات ماتركس، يعني اقتراح مو قرار).
function Lines({ lines, group }: { lines: ReportLine[]; group?: string }) {
  const [saved, setSaved] = useState<Record<string, boolean>>({})
  // «حوّلها تعليمة» = POST guide-rules (requireAdmin) — للمراقب ينخفي حتى ما ينسجل انتهاك.
  const isAdmin = useSession().employee?.role === 'ADMIN'
  if (lines.length === 0) return <p className="text-xs text-slate-400">ماكو شي.</p>
  const toRule = async (fix: string) => {
    try {
      await api.createGuideRule({ route: '/', match: '', groups: group ?? '', text: fix, enabled: false, priority: 5, onlyIfPending: false })
      setSaved((x) => ({ ...x, [fix]: true }))
    } catch { alert('تعذّر الحفظ') }
  }
  return (
    <ul className="space-y-1.5">
      {lines.map((l, i) => (
        <li key={i} className={`rounded-lg border px-3 py-1.5 text-sm ${TONE[l.tone]}`}>
          <div className="flex items-start justify-between gap-2">
            <span>{ICON[l.tone]} {l.text}</span>
            {l.link && <Link to={l.link} className="shrink-0 rounded bg-white/70 px-2 py-0.5 text-[11px] font-bold text-sky-700 hover:bg-white">ودّيني ←</Link>}
          </div>
          {l.fixes && l.fixes.length > 0 && (
            <div className="mt-1.5 space-y-1 border-t border-black/5 pt-1.5">
              <p className="text-[11px] font-bold opacity-70">💡 حلول مقترحة:</p>
              {l.fixes.map((f) => (
                <div key={f} className="flex items-center justify-between gap-2 text-xs">
                  <span>• {f}</span>
                  {saved[f]
                    ? <span className="shrink-0 text-[10px] text-emerald-700">انحفظت معطّلة ✓</span>
                    : isAdmin && <button onClick={() => toRule(f)} className="shrink-0 rounded border border-current/20 px-1.5 py-0.5 text-[10px] opacity-70 hover:opacity-100">🔁 حوّلها تعليمة</button>}
                </div>
              ))}
            </div>
          )}
        </li>
      ))}
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
  const navigate = useNavigate()
  const isAdmin = useSession().employee?.role === 'ADMIN'
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
          {isAdmin
            ? <Link to="/?board=matrix" className="text-sm text-brand-700 underline">← عيون ماتركس</Link>
            : <button onClick={() => navigate(-1)} className="text-sm text-brand-700 underline">← رجوع</button>}
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

          <Section title={`📝 شنو سوّى ${rep.day === todayBaghdad() ? 'اليوم' : 'بهاليوم'} (${(rep.activity ?? []).reduce((n, a) => n + a.count, 0)} فعل)`}>
            {(rep.activity ?? []).length === 0 ? (
              <p className="text-xs text-slate-400">ما سجّل ولا فعل بالنظام هاليوم (حفظ، قرار، إضافة، تعديل).</p>
            ) : (
              <>
                <div className="mb-2 flex flex-wrap gap-1.5">
                  {Object.entries((rep.activity ?? []).reduce<Record<string, number>>((m, a) => { m[a.area] = (m[a.area] ?? 0) + a.count; return m }, {}))
                    .sort((a, b) => b[1] - a[1])
                    .map(([area, n]) => <span key={area} className="rounded-full bg-sky-50 px-2.5 py-0.5 text-[11px] font-bold text-sky-800">{area}: {n}</span>)}
                </div>
                <ol className="max-h-80 space-y-1 overflow-y-auto border-r-2 border-sky-200 pr-3">
                  {(rep.activity ?? []).map((a, i) => (
                    <li key={i} className="flex gap-3 text-sm">
                      <span className="w-12 shrink-0 font-mono text-xs text-slate-400">{a.at}</span>
                      <span className="text-slate-800">{a.text}{a.count > 1 && <b className="mr-1 text-sky-700">×{a.count}</b>}</span>
                    </li>
                  ))}
                </ol>
              </>
            )}
          </Section>

          <div className="grid gap-4 lg:grid-cols-2">
            <Section title="🕐 الدوام"><Lines lines={rep.attendance} group={rep.group} /></Section>
            <Section title="🌴 الإجازات"><Lines lines={rep.leaves} /></Section>
          </div>

          {/* الحجوزات والسرعة للميدانيين بس — المحاسب والمراقب بشغل دورهم. */}
          {rep.field !== false && <Section title={`🔧 الحجوزات (${rep.jobs.length})`}>
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
          </Section>}

          {rep.field === false && rep.role && (
            <Section title={`${rep.role.title} (آخر ٣٠ يوم)`}>
              <div className="mb-3 grid grid-cols-2 gap-2 sm:grid-cols-4">
                {rep.role.metrics.map((m) => (
                  <div key={m.label} className="rounded-xl bg-slate-50 p-2 text-center ring-1 ring-slate-100">
                    <b className="block text-xl text-[#0f2040]">{m.value}</b>
                    <span className="text-[11px] text-slate-500">{m.label}</span>
                  </div>
                ))}
              </div>
              <Lines lines={rep.role.lines} group={rep.group} />
            </Section>
          )}

          {rep.field !== false && <Section title="📈 الأداء (آخر ٣٠ يوم)">
            <Lines lines={rep.performance ?? []} group={rep.group} />
            {rep.slowJobs?.length > 0 && (
              <div className="mt-3">
                <p className="mb-1 text-xs font-bold text-slate-600">🐢 أبطأ حجوزاته مقابل المتوقع لنفس الخدمة:</p>
                <table className="w-full text-xs">
                  <thead className="text-slate-400"><tr><th className="text-right">الحجز</th><th className="text-right">الخدمة</th><th>أخذ</th><th>المتوقع</th><th>الفرق</th></tr></thead>
                  <tbody>
                    {rep.slowJobs.map((j) => (
                      <tr key={j.bookingId} className="border-t border-slate-100">
                        <td className="py-1 font-mono">{j.code}</td><td>{j.service}</td>
                        <td className="text-center">{Math.round(j.actual / 6) / 10} س</td>
                        <td className="text-center">{Math.round(j.expected / 6) / 10} س</td>
                        <td className="text-center font-bold text-red-600">×{(j.actual / Math.max(1, j.expected)).toFixed(1)}</td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            )}
          </Section>}

          <div className="grid gap-4 lg:grid-cols-2">
            <Section title="🧭 شخصيته بالشغل (آخر ٣٠ يوم)"><Lines lines={rep.behavior} group={rep.group} /></Section>
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
