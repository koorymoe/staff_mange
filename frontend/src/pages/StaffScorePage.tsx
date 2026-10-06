import { useCallback, useEffect, useState } from 'react'
import { api, type StaffScore, type StaffScoreBoard, type StaffScoreDetail } from '../api'
import MatrixNote from '../components/MatrixNote'
import OwnerSwitch from '../components/OwnerSwitch'
import { SWITCH_MATRIX_SCORING } from '../systemSwitches'
import { useSession } from '../session'

// ═══ تقييم الموظفين — قرار (ع) 10-05 ═══
// النهائي = ٦٠٪ ماتركس (نقاط كل حجز ويوم دوام ومهمة ومشروع) + ٤٠٪ البشر
// (الليدر، الإداري، المراقب، الجودة). درجة تقييم بس — ماكو فلوس.

const pct = (v: number | null) => (v == null ? '—' : `${Math.round(v)}%`)
const tone = (v: number | null) => (v == null ? '#94a3b8' : v >= 80 ? '#059669' : v >= 60 ? '#d97706' : '#dc2626')
const SRC: Record<string, string> = { BOOKING: 'الحجوزات', DAY: 'الحضور', TASK: 'المهام', PROJECT: 'المشاريع' }
const STAGE: Record<string, string> = { LEADER_CREW: 'الليدر', COORD_LEADER: 'الإداري', AUDIT: 'المراقب', QUALITY_CALL: 'الجودة', MONITOR_PERIODIC: 'المراقب (دوري)' }
const ROLE_AR: Record<string, string> = {
  FINANCE: 'محاسب', DESIGNER: 'مصمم', IT_SUPPORT: 'آيتي', TECHNICIAN: 'فني', ENGINEER: 'مهندس', HR_COORDINATOR: 'إداري',
  SALES: 'مبيعات', MEDIA: 'إعلام', QUALITY_ENGINEER: 'جودة', MONITOR: 'مراقب', PROJECT_MANAGER: 'مدير مشاريع',
}

const GROUP_ORDER = ['LEADERS', 'TECHS', 'COORDINATORS', 'SALES', 'TECHNICAL', 'FINANCE', 'MONITORS', 'OTHERS']
const RATERS: Record<string, string> = {
  TECHS: 'الليدر مالتهم بعد كل حجز', LEADERS: 'المراقب + الإداري + الجودة', COORDINATORS: 'المراقب + الجودة',
  SALES: 'المراقب (كل نص شهر)', TECHNICAL: 'المراقب (كل نص شهر)', FINANCE: 'المراقب (كل نص شهر)',
  MONITORS: 'المدير والمالك (كل نص شهر)', OTHERS: 'المراقب (كل نص شهر)',
}

function ScoreBadge({ v, big }: { v: number | null; big?: boolean }) {
  return <b className={big ? 'text-3xl' : 'text-lg'} style={{ color: tone(v) }}>{pct(v)}</b>
}

export function ScoreBreakdown({ d, admin, onChange }: { d: StaffScoreDetail; admin?: boolean; onChange?: () => void }) {
  const [all, setAll] = useState(false)
  const cancel = async (id: string) => {
    const note = window.prompt('ليش تلغي هالنقطة؟')
    if (!note) return
    await api.cancelMatrixScore(id, note)
    onChange?.()
  }
  const lost = d.points.filter((p) => p.points < p.maxPoints || p.cancelledAt)
  const shown = all ? d.points : lost.slice(0, 20)
  return (
    <div className="space-y-3">
      <div className="grid grid-cols-3 gap-2 text-center">
        <div className="rounded-xl bg-slate-50 p-2"><p className="text-[11px] text-slate-500">النهائي</p><ScoreBadge v={d.final} big /></div>
        <div className="rounded-xl bg-violet-50 p-2"><p className="text-[11px] text-slate-500">ماتركس (٦٠٪)</p><ScoreBadge v={d.matrixPct} /><p className="text-[11px] text-slate-500">{d.earned} من {d.max} نقطة</p></div>
        <div className="rounded-xl bg-amber-50 p-2"><p className="text-[11px] text-slate-500">البشر (٤٠٪)</p><ScoreBadge v={d.humanPct} /><p className="text-[11px] text-slate-500">{d.humanCount ? `${d.humanAvg?.toFixed(1)}★ من ${d.humanCount} تقييم` : 'محد قيّمه بعد'}</p></div>
      </div>
      {d.reliabilityParts.length > 0 && (
        <div className="rounded-xl border border-sky-200 bg-sky-50/50 p-2">
          <p className="mb-1 flex items-center justify-between text-xs font-extrabold text-sky-900"><span>🛡️ الاعتمادية (رقم منفصل، مو داخل التقييم)</span><ScoreBadge v={d.reliability} /></p>
          <div className="grid gap-1 sm:grid-cols-2">
            {d.reliabilityParts.map((p) => (
              <p key={p.key} className="flex items-center justify-between gap-2 rounded-lg bg-white px-2 py-1 text-xs">
                <span>{p.label} <span className="text-slate-400">· {p.detail}</span></span><b style={{ color: tone(p.pct) }}>{Math.round(p.pct)}%</b>
              </p>
            ))}
          </div>
        </div>
      )}
      <div className="flex flex-wrap gap-2 text-xs">
        {Object.entries(d.bySource).map(([k, [e, m]]) => (
          <span key={k} className="rounded-full bg-slate-100 px-2 py-1">{SRC[k] ?? k}: <b>{e}/{m}</b></span>
        ))}
      </div>
      {d.topLosses.length > 0 && (
        <div className="rounded-xl border border-red-100 bg-red-50/50 p-2 text-xs">
          <p className="mb-1 font-extrabold text-red-800">أكثر شي نزّل تقييمه</p>
          {d.topLosses.map((l) => <p key={l.rule}>• <b>{l.title}</b>: خسر {l.lost} نقطة ({l.count} مرة).{l.advice && <span className="text-slate-600"> 💡 {l.advice}</span>}</p>)}
        </div>
      )}
      <div>
        <div className="mb-1 flex items-center justify-between">
          <p className="text-xs font-extrabold text-slate-700">{all ? 'كل نقاط ماتركس' : 'وين خسر نقاط'}</p>
          <button type="button" onClick={() => setAll(!all)} className="text-xs text-sky-700 underline">{all ? 'بس الي خسرها' : `شوف الكل (${d.points.length})`}</button>
        </div>
        <div className="max-h-80 space-y-1 overflow-y-auto">
          {shown.map((p) => (
            <div key={p.id} className={`flex items-start justify-between gap-2 rounded-lg px-2 py-1.5 text-xs ${p.cancelledAt ? 'bg-slate-50 text-slate-400 line-through' : p.points === p.maxPoints ? 'bg-emerald-50' : p.points === 0 ? 'bg-red-50' : 'bg-amber-50'}`}>
              <span>{p.reason}{p.cancelNote && <span className="no-underline"> (انلغت: {p.cancelNote})</span>}</span>
              <span className="flex shrink-0 items-center gap-1"><b>{p.points}/{p.maxPoints}</b>
                {admin && !p.cancelledAt && p.points < p.maxPoints && <button type="button" onClick={() => void cancel(p.id)} className="rounded border border-slate-300 bg-white px-1 text-[10px]">غلط؟</button>}
              </span>
            </div>
          ))}
          {shown.length === 0 && <p className="text-xs text-slate-400">ماكو.</p>}
        </div>
      </div>
      {d.human.length > 0 && (
        <div>
          <p className="mb-1 text-xs font-extrabold text-slate-700">تقييمات البشر</p>
          <div className="space-y-1">
            {d.human.map((h, i) => (
              <p key={i} className="rounded-lg bg-amber-50/60 px-2 py-1 text-xs">{'★'.repeat(h.score)}{'☆'.repeat(5 - h.score)} · {STAGE[h.stage] ?? h.stage}{h.raterName && ` (${h.raterName})`}{h.code && ` · حجز ${h.code}`}{h.note && ` — «${h.note}»`}</p>
            ))}
          </div>
        </div>
      )}
    </div>
  )
}

export default function StaffScorePage() {
  const { employee } = useSession()
  const admin = employee?.role === 'ADMIN' || employee?.actualRole === 'OWNER'
  const [month, setMonth] = useState('')
  const [b, setB] = useState<StaffScoreBoard | null>(null)
  const [sel, setSel] = useState<string | null>(null)
  const [d, setD] = useState<StaffScoreDetail | null>(null)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const [tab, setTab] = useState('')
  const load = useCallback(() => { void api.getStaffScoreBoard(month).then(setB).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [month])
  useEffect(load, [load])
  useEffect(() => { if (sel) void api.getStaffScore(sel, month).then(setD).catch(() => {}) }, [sel, month])
  const reloadDetail = () => { if (sel) void api.getStaffScore(sel, month).then(setD); load() }
  const run = async () => { setBusy(true); try { await api.runStaffScore(); load(); if (sel) reloadDetail() } finally { setBusy(false) } }

  if (err) return <p className="text-red-600">{err}</p>
  if (!b) return <p className="text-slate-400">ماتركس يحسب…</p>
  // قرار (ع) 10-06: كل مجموعة ترتيبها وحدها (الليدرية، الفنيين، المبيعات…).
  const groups = GROUP_ORDER.filter((g) => b.staff.some((s) => s.group === g))
  const g = groups.includes(tab) ? tab : groups[0] ?? ''
  const shown = b.staff.filter((s) => s.group === g)
  const trend = (s: StaffScore) => s.final == null || s.prevFinal == null ? '' : s.final > s.prevFinal + 2 ? ' ↑' : s.final < s.prevFinal - 2 ? ' ↓' : ''

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-end justify-between gap-2">
        <div>
          <h2 className="text-2xl font-bold text-brand-900">🏅 تقييم الموظفين</h2>
          <p className="text-sm text-slate-500">ماتركس ٦٠٪ (نقاط كل حجز ويوم دوام ومهمة ومشروع) + البشر ٤٠٪ (الليدر، الإداري، المراقب، الجودة). درجة تقييم بس، بلا فلوس.</p>
        </div>
        <div className="flex items-center gap-2">
          <input type="month" value={month || b.month} onChange={(e) => setMonth(e.target.value)} className="rounded-lg border border-slate-200 px-2 py-1 text-sm" />
          {admin && <button type="button" disabled={busy} onClick={() => void run()} className="rounded-lg bg-violet-700 px-3 py-1.5 text-xs font-bold text-white disabled:opacity-50">🔄 احسب هسه</button>}
        </div>
      </div>
      {admin && <OwnerSwitch switchKey={SWITCH_MATRIX_SCORING} label="ماتركس يقيّم الموظفين بالنقاط لحاله" hint="شغّال = يحسب كل ساعة. مطفي = ما يحسب لحاله (تقدر تحسب يدوي بزر «احسب هسه» حتى تجرّب)." />}
      {!b.on && <p className="rounded-xl bg-amber-50 p-2 text-xs text-amber-900">⏸️ الحساب التلقائي مطفي. الأرقام هنا من آخر حساب يدوي.</p>}
      <MatrixNote>{b.insights.join(' ')}</MatrixNote>
      <div className="flex flex-wrap gap-1.5">
        {groups.map((k) => (
          <button key={k} type="button" onClick={() => { setTab(k); setSel(null) }}
            className={`rounded-full px-3 py-1 text-sm font-bold ${g === k ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-700 ring-1 ring-slate-200'}`}>
            {b.staff.find((s) => s.group === k)?.groupLabel} <span className="text-xs opacity-70">{b.staff.filter((s) => s.group === k).length}</span>
          </button>
        ))}
      </div>
      {g && <p className="text-xs text-slate-500">👥 منو يقيّم {b.staff.find((s) => s.group === g)?.groupLabel}: {RATERS[g]}</p>}
      <div className="grid gap-4 lg:grid-cols-[1fr_1.2fr]">
        <div className="space-y-1.5">
          {shown.map((s, i) => (
            <button key={s.id} type="button" onClick={() => setSel(sel === s.id ? null : s.id)}
              className={`flex w-full items-center justify-between gap-2 rounded-xl border px-3 py-2 text-right ${sel === s.id ? 'border-[#0f2040] bg-white shadow' : 'border-slate-200 bg-white/80 hover:bg-white'}`}>
              <span className="flex items-center gap-2">
                <span className="w-5 text-xs text-slate-400">{s.final == null ? '' : i + 1}</span>
                <span><b className="text-sm">{s.name}</b> <span className="text-xs text-slate-500">{ROLE_AR[s.role] ?? s.role}</span>
                  {s.noHuman && <span className="mr-1 rounded bg-slate-100 px-1 text-[10px] text-slate-500">بلا تقييم بشري</span>}</span>
              </span>
              <span className="flex items-center gap-3 text-xs">
                <span className="hidden text-slate-500 sm:inline">ماتركس {pct(s.matrixPct)} · بشر {pct(s.humanPct)}</span>
                <span title="الاعتمادية" className="rounded-full px-2 py-0.5 font-bold" style={{ color: tone(s.reliability), background: 'var(--sf-sunken, #f1f5f9)' }}>🛡️ {pct(s.reliability)}</span>
                <span className="w-14 text-left"><ScoreBadge v={s.final} />{trend(s)}</span>
              </span>
            </button>
          ))}
        </div>
        <div className="rounded-2xl border border-slate-200 bg-white p-3 lg:sticky lg:top-4 lg:self-start">
          {!sel && <p className="p-6 text-center text-sm text-slate-400">اختار موظف حتى تشوف شلون انحسب تقييمه، نقطة بنقطة.</p>}
          {sel && d?.id !== sel && <p className="text-xs text-slate-400">…</p>}
          {d && d.id === sel && <><p className="mb-2 text-base font-extrabold">{d.name}</p><ScoreBreakdown d={d} admin={admin} onChange={reloadDetail} /></>}
        </div>
      </div>
    </div>
  )
}
