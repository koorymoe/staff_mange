import { money } from '../utils/money'
import DisciplineBox from '../components/DisciplineBox'
import { useCallback, useEffect, useState, type ReactNode } from 'react'
import { api, type StaffScore, type StaffScoreBoard, type StaffScoreDetail } from '../api'
import MatrixNote from '../components/MatrixNote'
import OwnerSwitch from '../components/OwnerSwitch'
import { GroupRanking, PairCompare, PendingHumanBadge } from '../components/StaffCompare'
import { SWITCH_MATRIX_SCORING } from '../systemSwitches'
import { useSession, isAdminOrOwner } from '../session'
import StaffPointsBoard from '../components/StaffPointsBoard'
import { askInput } from '../ui/dialogs'
import { swallow } from '../utils/swallow'
import PctWhy from '../ui/PctWhy'

// ═══ تقييم الموظفين — قرار (ع) 10-05 ═══
// النهائي = ٦٠٪ ماتركس (نقاط كل حجز ويوم دوام ومهمة ومشروع) + ٤٠٪ البشر
// (الليدر، الإداري، المراقب، الجودة). درجة تقييم بس — ماكو فلوس.

const pct = (v: number | null) => (v == null ? '—' : `${Math.round(v)}%`)
const tone = (v: number | null) => (v == null ? '#94a3b8' : v >= 80 ? '#059669' : v >= 60 ? '#d97706' : '#dc2626')
const SRC: Record<string, string> = { BOOKING: 'الحجوزات', DAY: 'الحضور', TASK: 'المهام', PROJECT: 'المشاريع', PROCUREMENT: 'المخازن', MONITOR: 'التدقيق', DUTY: 'واجب التقييم', REPORT: 'تقرير الإنجاز' }
const STAGE: Record<string, string> = { LEADER_CREW: 'الليدر', COORD_LEADER: 'الإداري', AUDIT: 'المراقب', QUALITY_CALL: 'الجودة', MONITOR_PERIODIC: 'المراقب (دوري)' }
const ROLE_AR: Record<string, string> = {
  FINANCE: 'محاسب', DESIGNER: 'مصمم', IT_SUPPORT: 'آيتي', TECHNICIAN: 'فني', ENGINEER: 'مهندس', HR_COORDINATOR: 'إداري',
  SALES: 'مبيعات', MEDIA: 'إعلام', PUBLIC_RELATIONS: 'علاقات عامة', QUALITY_ENGINEER: 'جودة', MONITOR: 'مراقب', PROJECT_MANAGER: 'مدير مشاريع', PROCUREMENT_ADMIN: 'إداري كميات',
}

const GROUP_ORDER = ['LEADERS', 'TECHS', 'DECOR', 'COORDINATORS', 'SALES', 'TECHNICAL', 'FINANCE', 'PROCUREMENT', 'MONITORS', 'QUALITY', 'DESIGN', 'IT', 'PROJECTS', 'OTHERS']
const REV_SRC: Record<string, string> = { BOOKED: 'الي سجّلها', WORK: 'الي اشتغلها (حصته من الكادر)', PROJECT: 'دفعات المشاريع الي هو مسؤولها' }

const RATERS: Record<string, string> = {
  TECHS: 'الليدر مالتهم بعد كل حجز', LEADERS: 'المراقب + الإداري + الجودة', COORDINATORS: 'المراقب + الجودة',
  SALES: 'المراقب (كل نص شهر)', TECHNICAL: 'المراقب (كل نص شهر)', FINANCE: 'المراقب (كل نص شهر)',
  PROCUREMENT: 'المراقب (كل نص شهر) + ماتركس على الطلبات والسيارات والجرد',
  MONITORS: 'المدير والمالك (كل نص شهر) + ماتركس على سرعة التدقيق', OTHERS: 'المراقب (كل نص شهر)',
  QUALITY: 'المراقب (كل نص شهر)', DESIGN: 'المراقب (كل نص شهر)', IT: 'المراقب (كل نص شهر)',
  PROJECTS: 'المراقب (كل نص شهر) + ماتركس على متابعة المشاريع',
  DECOR: 'الليدر مالتهم بعد كل حجز (والليدر: المراقب + الإداري + الجودة)',
}

function ScoreBadge({ v, big }: { v: number | null; big?: boolean }) {
  return <b className={big ? 'text-3xl' : 'text-lg'} style={{ color: tone(v) }}>{pct(v)}</b>
}

const r1 = (v: number | null | undefined) => (v == null ? '—' : Math.round(v * 10) / 10)
const SCORE_SRC = <>نقاط ماتركس للموظف بهذا الشهر (كل حجز ويوم دوام ومهمة ومشروع وطلب مخزن — كل نقطة وياها سببها بالقائمة تحت)، وتقييمات البشر بالنجوم (الليدر، الإداري، المراقب، الجودة).</>

// نفس شكل ScoreBadge بس الرقم ينضغط ويطلّع منين جا.
function WhyBadge({ d, kind, big }: { d: StaffScoreDetail; kind: 'final' | 'matrix' | 'human' | 'reliability'; big?: boolean }) {
  const v = kind === 'final' ? d.final : kind === 'matrix' ? d.matrixPct : kind === 'human' ? d.humanPct : d.reliability
  const cls = big ? 'text-3xl' : 'text-lg'
  if (v == null) return <ScoreBadge v={null} big={big} />
  const disc = d.discipline?.score
  let formula: ReactNode
  let meaning: ReactNode
  let title: string
  if (kind === 'matrix') {
    title = 'رقم ماتركس'
    formula = <>أخذ {d.earned} نقطة من أصل {d.max} نقطة ممكنة → {d.earned} ÷ {d.max} × 100 = <b>{pct(v)}</b></>
    meaning = <>شكد أخذ من النقاط الي جان يگدر ياخذها من شغله الفعلي. كل قاعدة ما التزم بيها تنزّل النقاط (شوف «أكثر شي نزّل تقييمه»). وزنه ٦٠٪ من النهائي.</>
  } else if (kind === 'human') {
    title = 'رقم البشر'
    formula = <>معدل نجومه {r1(d.humanAvg)}★ من {d.humanCount} تقييم، والنجمة = 20 → {r1(d.humanAvg)} × 20 = <b>{pct(v)}</b></>
    meaning = <>شلون يشوفوه الي يقيّموه (من 5 نجوم). 5★ = 100. وزنه ٤٠٪ من النهائي.</>
  } else if (kind === 'reliability') {
    title = 'الاعتمادية'
    formula = <>معدل {d.reliabilityParts.length} أجزاء (كل جزء من 100): {d.reliabilityParts.map((p) => `${p.label} ${Math.round(p.pct)}`).join(' + ')} ÷ {d.reliabilityParts.length} = <b>{pct(v)}</b></>
    meaning = <>رقم منفصل عن التقييم: هل تعتمد عليه؟ الحضور والجرد والورق والمهام نسبة من نقاط ماتركس، والشكاوى تنزّل 25 لكل شكوى، والإجازات 10 لكل يوم فوق يومين، والمشاكل ويا الزملاء وضياع العدّة 25 لكل وحدة، ونقاط الـKPI 20 لكل نقطة.</>
  } else {
    title = 'التقييم النهائي'
    const base = d.matrixPct != null && d.humanPct != null
      ? <>(٦٠٪ × ماتركس {r1(d.matrixPct)}) + (٤٠٪ × البشر {r1(d.humanPct)}) = {r1(0.6 * d.matrixPct + 0.4 * d.humanPct)}</>
      : d.matrixPct != null ? <>ماكو تقييم بشر بعد، فناخذ ماتركس بس = {r1(d.matrixPct)}</> : <>ماكو نقاط ماتركس، فناخذ البشر بس = {r1(d.humanPct)}</>
    formula = <>{base}{!!d.revenueAdj && <><br />{d.revenueAdj > 0 ? '+' : '−'} {r1(Math.abs(d.revenueAdj))} نقطة من تغطية راتبه ({r1(d.salaryCoverage ?? 0)}٪) — {d.revenueAdj > 0 ? 'دخّل أكثر من راتبه' : 'ما دخّل قد راتبه، فينزل بنسبة النقص'}</>}{!!d.extraBonus && <><br />+ {r1(d.extraBonus)} نقطة شغل إضافي برّا دوره ({d.extraJobs} شغلة، نقطة لكل وحدة لحد ٥)</>}{disc != null && <><br />وبعدها الانضباط ياخذ ١٥٪: (٨٥٪ × الرقم فوگ) + (١٥٪ × الانضباط {r1(disc)})</>} = <b>{pct(v)}</b></>
    meaning = <>درجة تقييم الموظف بهذا الشهر. 80 وفوق أخضر، 60–80 أصفر، أقل أحمر. تصعد إذا أخذ نقاط ماتركس كاملة ونجوم عالية وحضر بوقته.</>
  }
  return <b className={cls} style={{ color: tone(v) }}><PctWhy title={title} value={Math.round(v)} formula={formula} meaning={meaning}
    source={kind === 'reliability' ? <>نقاط ماتركس للحضور والجرد والورق والمهام، وشكاوى الزبائن، والإجازات، والمشاكل الوظيفية، وضياع العدّة، ونقاط الـKPI لهذا الشهر.</> : SCORE_SRC} /></b>
}

export function ScoreBreakdown({ d, admin, onChange }: { d: StaffScoreDetail; admin?: boolean; onChange?: () => void }) {
  const [all, setAll] = useState(false)
  const cancel = async (id: string) => {
    const note = (await askInput('ليش تلغي هالنقطة؟'))
    if (!note) return
    await api.cancelMatrixScore(id, note)
    onChange?.()
  }
  const lost = d.points.filter((p) => p.points < p.maxPoints || p.cancelledAt)
  const shown = all ? d.points : lost.slice(0, 20)
  return (
    <div className="space-y-3">
      <div className="grid grid-cols-3 gap-2 text-center">
        <div className="rounded-xl bg-slate-50 p-2"><p className="text-[11px] text-slate-500">النهائي</p><WhyBadge d={d} kind="final" big /></div>
        <div className="rounded-xl bg-violet-50 p-2"><p className="text-[11px] text-slate-500">ماتركس (٦٠٪)</p><WhyBadge d={d} kind="matrix" /><p className="text-[11px] text-slate-500">{d.earned} من {d.max} نقطة</p></div>
        <div className="rounded-xl bg-amber-50 p-2"><p className="text-[11px] text-slate-500">البشر (٤٠٪)</p><WhyBadge d={d} kind="human" /><p className="text-[11px] text-slate-500">{d.humanCount ? `${d.humanAvg?.toFixed(1)}★ من ${d.humanCount} تقييم` : 'محد قيّمه بعد'}</p></div>
      </div>
      {d.discipline && <DisciplineBox d={d.discipline} />}
      {d.reliabilityParts.length > 0 && (
        <div className="rounded-xl border border-sky-200 bg-sky-50/50 p-2">
          <p className="mb-1 flex items-center justify-between text-xs font-extrabold text-sky-900"><span>🛡️ الاعتمادية (رقم منفصل، مو داخل التقييم)</span><WhyBadge d={d} kind="reliability" /></p>
          <div className="grid gap-1 sm:grid-cols-2">
            {d.reliabilityParts.map((p) => (
              <p key={p.key} className="flex items-center justify-between gap-2 rounded-lg bg-white px-2 py-1 text-xs">
                <span>{p.label} <span className="text-slate-400">· {p.detail}</span></span><b style={{ color: tone(p.pct) }}><PctWhy title={p.label} value={Math.round(p.pct)}
                  formula={<>{p.detail} → <b>{Math.round(p.pct)}%</b></>}
                  meaning={<>جزء من الاعتمادية (من 100). {p.key === 'COMPLAINTS' ? 'كل شكوى زبون تنزّل 25.' : p.key === 'LEAVES' ? 'يومين إجازة بالشهر طبيعي، وكل يوم فوقهن ينزّل 10.' : p.key === 'PEERS' || p.key === 'TOOLS' ? 'كل وحدة تنزّل 25.' : p.key === 'KPI' ? 'كل نقطة KPI عليه تنزّل 20.' : 'النقاط الي أخذها من النقاط الممكنة بهذا البند.'}</>}
                  source={<>{p.key === 'COMPLAINTS' ? 'شكاوى الزبائن المسجّلة عليه بهذا الشهر.' : p.key === 'LEAVES' ? 'الإجازات والغيابات المسجّلة إله بهذا الشهر.' : p.key === 'PEERS' ? 'المشاكل الوظيفية المسجّلة عليه.' : p.key === 'TOOLS' ? 'سجل العدّة الي ضاعت أو تلفت.' : p.key === 'KPI' ? 'نقاط الـKPI المسجّلة عليه.' : 'نقاط ماتركس لهذا البند بهذا الشهر.'}</>} /></b>
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
      {d.revenueKind && (
        <div className={`rounded-xl border p-2 text-xs ${d.salaryCoverage == null ? 'border-slate-200 bg-slate-50' : d.salaryCoverage >= 100 ? 'border-emerald-200 bg-emerald-50/60' : 'border-amber-200 bg-amber-50/60'}`} data-testid="revenue">
          <p className="font-extrabold">💰 دخّل للشركة هالشهر: {money(d.revenue ?? 0)}
            <span className="font-normal text-slate-500"> · من {d.revenueJobs ?? 0} {d.revenueKind === 'PROJECT' ? 'مشروع' : 'حجز'} {REV_SRC[d.revenueKind]}</span></p>
          {d.salaryCoverage != null
            ? <p className="mt-0.5">يغطي راتبه: <PctWhy title="تغطية الراتب" value={Math.round(d.salaryCoverage)}
                formula={<>الي دخّله {money(d.revenue ?? 0)} ÷ راتب الأيام الي مرّت من الفترة × 100 = <b>{Math.round(d.salaryCoverage)}٪</b></>}
                meaning={<>كل موظف لازم يدخّل أكثر من راتبه. تحت ١٠٠٪ نهائيه ينزل بنسبة النقص (الي ما دخّل شي ينزل للنص)، وفوك ١٠٠٪ ياخذ زيادة لحد ١٠ نقاط.</>}
                source={<>فواتير الحجوزات (غير المسودة والملغية) {REV_SRC[d.revenueKind]}، والراتب من ملف الموظف.</>} /></p>
            : <p className="mt-0.5 text-slate-500">ماكو راتب مسجّل، فما تنحسب التغطية.</p>}
        </div>
      )}
      {!!d.extraJobs && (
        <div className="rounded-xl border border-emerald-100 bg-emerald-50/60 p-2 text-xs" data-testid="extra-work">
          <p className="mb-1 font-extrabold text-emerald-800">➕ شغل إضافي برّا دوره: {d.extraJobs} (+{r1(d.extraBonus)} نقطة)</p>
          <p className="mb-1 text-slate-600">ما يتقيّم عليه مثل أهل هالشغل — ينكتب هنا ويزيد تقييمه بس.</p>
          {(d.extraList ?? []).slice(0, 10).map((x) => <p key={x}>• {x}</p>)}
        </div>
      )}
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
  const admin = isAdminOrOwner(employee)
  const [month, setMonth] = useState('')
  const [b, setB] = useState<StaffScoreBoard | null>(null)
  const [sel, setSel] = useState<string | null>(null)
  const [d, setD] = useState<StaffScoreDetail | null>(null)
  const [err, setErr] = useState('')
  const [busy, setBusy] = useState(false)
  const [tab, setTab] = useState('')
  // قرار (ع) 10-10: «مقارنة» — ترتيب الأفضل للأسوأ، و«قارن بموظف آخر»
  const [view, setView] = useState<'list' | 'compare' | 'points'>('points')
  const [cmpA, setCmpA] = useState<string | null>(null)
  const [cmpB, setCmpB] = useState<string | null>(null)
  const load = useCallback(() => { void api.getStaffScoreBoard(month).then(setB).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر')) }, [month])
  useEffect(load, [load])
  useEffect(() => { if (sel) void api.getStaffScore(sel, month).then(setD).catch(swallow) }, [sel, month])
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
      <div className="flex gap-1.5">
        {([['points', '🏅 رصيد النقاط'], ['list', '📋 التقييم'], ['compare', '⚖️ المقارنة']] as const).map(([k, l]) => (
          <button key={k} type="button" onClick={() => setView(k)}
            className={`rounded-lg px-3 py-1.5 text-sm font-bold ${view === k ? 'bg-violet-700 text-white' : 'bg-white text-slate-700 ring-1 ring-slate-200'}`}>{l}</button>
        ))}
      </div>
      {view === 'compare' && <GroupRanking staff={shown} />}
      {view === 'points' && g && <StaffPointsBoard month={month || b.month} group={g} />}
      {g && view !== 'points' && <p className="text-xs text-slate-500">👥 منو يقيّم {b.staff.find((s) => s.group === g)?.groupLabel}: {RATERS[g]}</p>}
      {view === 'list' && <div className="grid gap-4 lg:grid-cols-[1fr_1.2fr]">
        <div className="space-y-1.5">
          {shown.map((s, i) => (
            <div key={s.id} className="flex items-stretch gap-1">
            <button type="button" onClick={() => { setCmpA(null); setSel(sel === s.id ? null : s.id) }}
              className={`flex w-full items-center justify-between gap-2 rounded-xl border px-3 py-2 text-right ${sel === s.id ? 'border-[#0f2040] bg-white shadow' : 'border-slate-200 bg-white/80 hover:bg-white'}`}>
              <span className="flex items-center gap-2">
                <span className="w-5 text-xs text-slate-400">{s.final == null ? '' : i + 1}</span>
                <span><b className="text-sm">{s.name}</b> <span className="text-xs text-slate-500">{ROLE_AR[s.role] ?? s.role}</span>
                  {s.noHuman && <PendingHumanBadge />}</span>
              </span>
              <span className="flex items-center gap-3 text-xs">
                <span className="hidden text-slate-500 sm:inline">ماتركس {pct(s.matrixPct)} · بشر {pct(s.humanPct)}</span>
                <span title="الاعتمادية" className="rounded-full px-2 py-0.5 font-bold" style={{ color: tone(s.reliability), background: 'var(--sf-sunken, #f1f5f9)' }}>🛡️ {pct(s.reliability)}</span>
                <span className="w-14 text-left"><ScoreBadge v={s.final} />{trend(s)}</span>
              </span>
            </button>
            <button type="button" title="قارن بموظف آخر" onClick={() => { setSel(null); setCmpA(s.id); setCmpB(null) }}
              className={`shrink-0 rounded-xl border px-2 text-xs font-bold ${cmpA === s.id ? 'border-violet-600 bg-violet-600 text-white' : 'border-slate-200 bg-white text-violet-700 hover:bg-violet-50'}`}>⚖️<span className="hidden sm:inline"> قارن</span></button>
            </div>
          ))}
        </div>
        <div className="rounded-2xl border border-slate-200 bg-white p-3 lg:sticky lg:top-4 lg:self-start">
          {cmpA && shown.some((s) => s.id === cmpA) ? (
            <PairCompare a={shown.find((s) => s.id === cmpA)!} b={shown.find((s) => s.id === cmpB) ?? null}
              peers={shown} onPick={(id) => setCmpB(id || null)} onClose={() => setCmpA(null)} />
          ) : !sel && <p className="p-6 text-center text-sm text-slate-400">اختار موظف حتى تشوف شلون انحسب تقييمه، نقطة بنقطة — أو اضغط «⚖️ قارن» جنب اسمه.</p>}
          {sel && d?.id !== sel && <p className="text-xs text-slate-400">…</p>}
          {!cmpA && d && d.id === sel && <><p className="mb-2 text-base font-extrabold">{d.name}</p><ScoreBreakdown d={d} admin={admin} onChange={reloadDetail} /></>}
        </div>
      </div>}
    </div>
  )
}
