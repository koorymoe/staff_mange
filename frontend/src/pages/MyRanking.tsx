import { thisPeriod, periodRangeLabel } from '../utils/format'
import { useEffect, useState, type ReactNode } from 'react'
import PctWhy from '../ui/PctWhy'
import { type KpiLeaderboardEntry, type StaffScoreDetail } from '../api'
import { api, type Stats, type RoleKpiLeaderboard } from '../api'
import { useSession, isAdminOrOwner } from '../session'
import { todayBaghdad } from '../utils/format'
import { AdministrativeTab as KpiAdministrativeTab } from './KpiPage'
import { roleLabel as arabicRole } from '../roleLabels'
import { tracksFor } from '../rankingTracks'
import PerformanceReviewPage from './PerformanceReviewPage'
import { swallow } from '../utils/swallow'
import { LineIcon } from '../components/LineIcon'
import './MyRanking.css'

const levels = [
  { level: 1, label: 'متدرب', min: 0 },
  { level: 2, label: 'فني مبتدئ', min: 3 },
  { level: 3, label: 'فني', min: 6 },
  { level: 4, label: 'فني متمرس', min: 10 },
  { level: 5, label: 'فني خبير', min: 15 },
]

const BOOKINGS_PER_RANK = 10

// ═══ «التقييم» ═══
//
// «بدل كلمة تصنيفي نخلي كلمة التقييم — من ندخل عليها تفتح واجهة بيها
// تقييمي وتصنيفي، وبيها خيار تقييم فريقي. نفس الآلية الي اشتغلنا بيها
// بأغلب الواجهات».
//
// نفس نمط «مهامي» و«الجرد»: بند واحد بالقائمة، والخيارات من فوگ.
type BoardCol = { label: string; icon: string; get: (e: KpiLeaderboardEntry) => number }
/** قرار (ع) 10-08: أدوار تتقيّم على شغلها هي مو على نقاط KPI — كل دور وحده. */
const CUSTOM_BOARDS: Record<string, { cols: BoardCol[]; explain: [string, string][]; sortNote: string }> = {
  SALES: {
    cols: [
      { label: 'الحجوزات المرحّلة', icon: '🗂️', get: (e) => e.createdBookings ?? 0 },
      { label: 'المهام الإضافية', icon: '📋', get: (e) => e.extraTasksDone ?? 0 },
    ],
    explain: [
      ['🗂️ الحجوزات المرحّلة', 'الحجوزات الي سجّلتها ورحّلتها للشعبة الهندسية بالفترة (الملغاة ما تنحسب).'],
      ['🕐 الالتزام بالدوام', 'من بصمات حضورك وانصرافك — أيام الحضور من أيام العمل.'],
      ['📋 المهام الإضافية', 'المهام الي توجّهت إلك من مسؤولك وخلّصتها.'],
    ],
    sortNote: 'الترتيب على الحجوزات المرحّلة، وبعدها المهام، وعند التساوي الالتزام بالدوام.',
  },
  MONITOR: {
    cols: [
      { label: 'التدقيقات', icon: '🔎', get: (e) => e.monitorAudits ?? 0 },
      { label: 'تدقيق الفواتير (إضافي)', icon: '🧾', get: (e) => e.invoiceAudits ?? 0 },
      { label: 'المهام الإضافية', icon: '📋', get: (e) => e.extraTasksDone ?? 0 },
    ],
    explain: [
      ['🔎 التدقيقات', 'شغلك اليومي: تراجع محطات الحجوزات بصندوق المراقب (توجيه الكوادر، تقييم الإداري لليدرية، اتصال الجودة، اعتماد الفواتير، شغل التقني)، وتقييمات التدقيق والدوري، وتدقيق الشكاوى.'],
      ['🧾 تدقيق الفواتير', 'إضافي — الفواتير الي دقّقتها بنفسك. ما ينقارن بالمحاسب، كل دور لحاله.'],
      ['🕐 الالتزام بالدوام', 'من بصمات حضورك وانصرافك.'],
      ['📋 المهام الإضافية', 'المهام الي توجّهت إلك وخلّصتها.'],
    ],
    sortNote: 'الترتيب على التدقيقات، وبعدها تدقيق الفواتير والمهام، وعند التساوي الالتزام بالدوام.',
  },
}

/** تقييم ماتركس النهائي ويا «منين جا» — نفس FinalScore بالسيرفر (matrix_score_service.go). */
function MatrixPct({ v, s, who, className }: { v: number | null | undefined; s?: StaffScoreDetail | null; who?: string; className?: string }) {
  const r = (n: number) => Math.round(n)
  let formula: ReactNode = <>ماكو نقاط ماتركس ولا تقييم مسؤولين بالشهر، فماكو نسبة.</>
  // (ك) 11-26: إذا السيرفر رجّع بنود الحسبة (الفني والأدوار الجديدة) نشرح منها هي — مو من حسبة ٦٠/٤٠ القديمة
  if (v != null && s?.techParts?.length) {
    const extras: string[] = []
    if (s.extraBonus) extras.push(`+ ${r(s.extraBonus)} بونص شغل برّا دورك`)
    if (s.discipline?.score != null) extras.push(`ويا الانضباط بالدوام ${r(s.discipline.score)}% (يدخل بـ١٥٪)`)
    if (s.revenueAdj) extras.push(`وتغطية الراتب ${s.revenueAdj > 0 ? '+' : ''}${r(s.revenueAdj)}`)
    formula = <>{s.techParts.map((p, i) => <span key={p.key}>{i > 0 && ' + '}{p.label} {r(p.pct)}% × {r(p.weight)}%</span>)}
      {extras.length > 0 && <> {extras.join(' ')}</>} = {r(v)}%.</>
  } else if (v != null && s) {
    const m = s.matrixPct, h = s.humanPct
    formula = m != null && h != null
      ? <>شغلك المسجّل {r(m)}% (جمعت {r(s.earned / 2)} من {r(s.max / 2)} نقطة) وهذا ياخذ ٦٠٪ من الرقم + تقييم مسؤولك {r(h)}% (معدله {s.humanAvg?.toFixed(1)} من ٥، على {s.humanCount} تقييم) وهذا ياخذ ٤٠٪ = {r(v)}%.</>
      : m != null
        ? <>مسؤولك بعده ما قيّمك، فالرقم من شغلك المسجّل بس: {r(s.earned / 2)} من {r(s.max / 2)} نقطة = {r(v)}%.</>
        : h != null
          ? <>بعدك ما جمعت نقاط شغل، فالرقم من تقييم مسؤولك بس: معدله {s.humanAvg?.toFixed(1)} من ٥ = {r(v)}%.</>
          : <>التقييم = {r(v)}% — محسوب من شغلك بالشهر.</>
  } else if (v != null) {
    formula = <>{who ? `تقييم ${who}` : 'التقييم'} = نسبة نقاط ماتركس × ٦٠٪ + تقييم المسؤولين (المعدل من ٥ × ٢٠) × ٤٠٪ = {r(v)}%. إذا واحد منهن ماكو، تنحسب من الثاني وحده.</>
  }
  return (
    <PctWhy title={who ? `تقييم ${who}` : 'تقييمك'} value={v == null ? null : r(v)} className={className} formula={formula}
      meaning={<>رقمك من ١٠٠ <b>للشهر كله</b> (حتى لو مختار «أسبوعي»). يطلع من شغلك المسجّل بالنظام (الحضور، الجرد، الورق، المهام، الحجوزات) ومن تقييم مسؤولك. كل تقصير ينسجل ينزّله، وكل شغل زين يرفعه.</>}
      source={<>نقاط ماتركس المسجلة بالشهر (بلا الملغية) وتقييمات المسؤولين المسموحين لكل مجموعة.</>}>
      {s && s.topLosses.length > 0 && <ul className="space-y-1 text-xs">{s.topLosses.slice(0, 5).map((l) => <li key={l.rule}>• {l.title}: {l.count} مرة</li>)}</ul>}
    </PctWhy>
  )
}

/** «؟» صغيرة — تفتح طريقة الحساب مختصرة (قرار (ع) 10-08). */
function HelpDot({ lines }: { lines: string[] }) {
  const [open, setOpen] = useState(false)
  return (
    <span className="relative inline-block">
      <button type="button" onClick={() => setOpen((o) => !o)} aria-label="شلون ينحسب؟"
        className="flex h-5 w-5 items-center justify-center rounded-full border border-slate-300 bg-white text-[11px] font-black text-slate-500 hover:bg-slate-50">؟</button>
      {open && (
        <span className="absolute right-0 top-7 z-20 block w-72 rounded-xl border border-slate-200 bg-white p-3 text-[11px] font-normal leading-relaxed text-slate-600 shadow-lg">
          {lines.map((l) => <span key={l} className="block py-0.5">{l}</span>)}
        </span>
      )}
    </span>
  )
}

/** تقييم ماتركس الي يخص الموظف نفسه: الرقم، وشنو خسّره نقاط ونصيحته. */
function MatrixMine({ score }: { score: StaffScoreDetail | null }) {
  if (!score) return null
  return (
    <div className="mr-card p-5">
      <div className="flex items-center justify-between">
        <h3 className="mr-h flex items-center gap-2 text-sm font-extrabold"><LineIcon n="cal" className="h-5 w-5" /> تقييمك هالشهر</h3>
        <span className="mr-big mr-bigbox text-3xl font-black"><MatrixPct v={score.final} s={score} /></span>
      </div>
      <p className="mr-sub mt-1 text-[11px]">
        {score.month !== thisPeriod() && <b className="text-amber-700">هذا تقييم الشهر الي فات ({periodRangeLabel(score.month)}) — هالشهر بعده بأوله · </b>}
        {score.max > 0 ? `جمعت ${Math.round(score.earned / 2)} نقطة من أصل ${Math.round(score.max / 2)} بشغلك المسجّل` : 'بعدك ما جمعت نقاط هالشهر — تجيك من تقريرك اليومي وحجوزاتك ومهامك وتقييم مسؤولك'}
        {score.humanCount > 0 && ` · مسؤولك قيّمك ${score.humanCount} مرة`}
      </p>
      {score.topLosses.length > 0 && (
        <div className="mt-3 space-y-2">
          <p className="mr-h text-[11px] font-bold">أكثر الأشياء الي نزّلت رقمك:</p>
          {score.topLosses.slice(0, 3).map((l, i) => (
            <div key={l.rule} className={`mr-loss l${i} text-[11px]`}>
              <span className="mr-tile"><LineIcon n={['cal', 'gear', 'layers'][i]} className="h-4 w-4" /></span>
              <div className="min-w-0">
                <b className="block font-bold">{l.title} — صارت {l.count} مرة</b>
                {l.advice && <span className="mt-0.5 block">{l.advice}</span>}
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}

export default function MyRanking() {
  const { employee, permissions } = useSession()
  // «تقييم الموظفين» ما يطلع إلا لمن يقيّم فعلاً — الليدر أو من عنده
  // الصلاحية. تبويب يفتح شاشة تگله «ممنوع» أسوأ من تبويب ما موجود.
  const canReviewTeam = isAdminOrOwner(employee)
    || !!employee?.isLeader
    || (permissions ?? []).includes('performance_review')
  // «تقييم بين الإداريين» — منو يقارن أبو الكوادر والمراقبين ببعض،
  // مو أبو الكوادر نفسه. نفس حارس الخادم بالضبط
  // (`RequireRole("ADMIN","OWNER","MONITOR")`).
  const canSeeEvaluators = isAdminOrOwner(employee) || (employee?.role === 'MONITOR' || permissions.includes('monitoring'))
  // ⚠️⚠️ الإداريون العليا (ADMIN/OWNER) ماعندهم تصنيف شخصي — «عندهم
  // كل الصلاحيات بحكم موقعهم، فيطلعون بكل تصنيف ويزاحمون الي يشتغل
  // الشغل فعلاً» (نفس مبدأ استثناء PermissionLeaderboard بالخادم).
  // بلا هذا الشرط، ADMIN بلا `tracksFor` مطابق يسقط لمقارنة كل
  // حسابات ADMIN ببعض بنقاط KPI/حجوزات — مقارنة بلا معنى لعمل إداري.
  const isTopAdmin = isAdminOrOwner(employee)
  // (ع) 10-13: «تقييم كي بي اي» تبويب هنا — مو شاشة لحالها
  const canKpi = employee?.role === 'ADMIN' || (employee?.role === 'MONITOR' || permissions.includes('monitoring')) || (permissions ?? []).includes('kpi_management')
  const [view, setView] = useState<'mine' | 'team' | 'evaluators' | 'kpi'>('mine')
  const [stats, setStats] = useState<Stats | null>(null)
  const [board, setBoard] = useState<RoleKpiLeaderboard | null>(null)
  const [peers, setPeers] = useState<Map<string, number | null>>(new Map())
  const [myScore, setMyScore] = useState<StaffScoreDetail | null>(null)
  useEffect(() => {
    api.getStaffScorePeers().then((r) => setPeers(new Map(r.map((x) => [x.id, x.final])))).catch(swallow)
    api.getMyScore().then(async (r) => {
      const cur = r?.score ?? null
      // (11-02) أول الشهر بعده ماكو شغل كافي — نعرض الشهر الماضي بداله، مو «ماكو شغل»
      if (cur && cur.max === 0 && cur.humanCount === 0) {
        const d = new Date(todayBaghdad() + 'T12:00:00'); d.setDate(1); d.setMonth(d.getMonth() - 1)
        const prev = await api.getMyScore(d.toISOString().slice(0, 7)).catch(() => null)
        if (prev?.score && (prev.score.max > 0 || prev.score.humanCount > 0)) { setMyScore(prev.score); return }
      }
      setMyScore(cur)
    }).catch(swallow)
  }, [])
  const [evalBoard, setEvalBoard] = useState<RoleKpiLeaderboard | null>(null)
  const [evalPeriod, setEvalPeriod] = useState<'weekly' | 'monthly'>('weekly')
  const [period, setPeriod] = useState<'weekly' | 'monthly'>('weekly')

  // ⚠️ `employee` يجي متأخر من الجلسة — الحالة الافتراضية 'mine' تنكتب
  // وقت أول رسم قبل ما نعرف الدور. لمن يوصل ويطلع إداري عليا، هذا
  // المشتق (لا حالة إضافية ولا Effect) يحوّل العرض لتبويب يشتغل إله
  // فعلاً بدل ما يبقى واقف على تبويب ما راح يظهر أصلاً.
  const effectiveView = isTopAdmin && view === 'mine' ? (canReviewTeam ? 'team' : 'evaluators') : view

  useEffect(() => {
    if (effectiveView === 'evaluators') api.getEvaluatorLeaderboard().then(setEvalBoard).catch(() => setEvalBoard(null))
  }, [effectiveView])

  const isTechnician = employee?.role === 'TECHNICIAN'

  // /stats محمي بصلاحية «إحصائيات الموظفين» — بلاها الطلب ينرفض بلا فايدة.
  const canStats = (permissions ?? []).includes('employee_stats')
  useEffect(() => {
    if (isTechnician && canStats) api.getStats().then(setStats).catch(() => setStats(null))
  }, [isTechnician, canStats])

  // ═══ التصنيف حسب الشغل مو حسب الدور ═══
  //
  // مسارات الموظف = صلاحياته الفعلية. الي عنده «تنسيق الحجوزات»
  // ينقارن بمنسّقي الحجوزات مهما كان اسم دوره — والي يشتغل شغلتين
  // يظهر بتصنيفين.
  // ⚠️ الدور ينمرّر: عائلته هي المسار الأول — بلاه يرجع الترشيح
  // بالصلاحية وحدها وتطلع علّة «إداري الكوادر بتقييمات المبيعات».
  const myTracks = tracksFor(permissions, employee?.role, employee?.isLeader, employee?.division)
  const [trackIdx, setTrackIdx] = useState(0)
  const activeTrack = myTracks[trackIdx] ?? null

  useEffect(() => {
    if (!employee) return
    // بلا صلاحية تصنيف، نرجع للدور — حتى ما يشوف الموظف شاشة فاضية
    // بينما هو فعلاً عنده زملاء بنفس الدور.
    // ⚠️ مسار المبيعات استثناء: صاحب النظام رفض "نفس الشغل" هنا
    // تحديداً — إداري الكوادر ومهندس الجودة يحملون صلاحية
    // sales_booking بحكم دورهم، فكانوا يدخلون ترتيب المبيعات ويخلطونه
    // مع موظفي المبيعات الحقيقيين. `strictRoles` يرجعه لترتيب عائلة
    // الصارم بدل ترتيب الصلاحية.
    const p = activeTrack?.strictRoles?.length
      // العائلة تنرسل مفصولة بفاصلة — الخادم يرشّح `role = ANY(...)`
      ? api.getRoleKpiLeaderboard(activeTrack.strictRoles.join(','), { leader: activeTrack.leader, division: activeTrack.division })
      : activeTrack
      ? api.getPermissionKpiLeaderboard(activeTrack.permission)
      // 🔴 **والإداري العليا ما ننادي إله أصلاً**: الخادم يرفض
      // ADMIN/OWNER بـ٤٠٠ **بقصد** (ماكو تصنيف شخصي للإداريين)،
      // والشاشة تعرف هذا أصلاً (`isTopAdmin` فوگ) — فالنداء چان
      // يطلّع ٤٠٠ بكل فتحة شاشة للمالك، ويخلي الكونسول يمتلي
      // بأخطاء متوقّعة **تخبّي الأخطاء الحقيقية**.
      : isTopAdmin
      ? null
      // الديكور: لوحتهم لحالهم، ما يختلطون ويا الشعبة الهندسية (قرار 10-08).
      : api.getRoleKpiLeaderboard(employee.role, employee.division === 'DECOR' ? { division: 'DECOR' } : undefined)
    if (!p) { return }
    p.then(setBoard).catch(() => setBoard(null))
  }, [employee, activeTrack, isTopAdmin])

  const skillCount = employee?.skills.filter((s) => s.canPerform).length || 0
  const currentLevel = [...levels].reverse().find((l) => skillCount >= l.min) || levels[0]
  const nextLevel = levels.find((l) => l.min > skillCount)

  const myTechStat = stats?.technicianStats.find((s) => s.employeeId === employee?.id)
  const completedCount = myTechStat?.completed || 0

  const rank = Math.floor(completedCount / BOOKINGS_PER_RANK) + 1
  const remainingForNextRank = BOOKINGS_PER_RANK - (completedCount % BOOKINGS_PER_RANK)

  const sortedTechs = stats
    ? [...stats.technicianStats].sort((a, b) => b.completed - a.completed)
    : []
  const myPosition = sortedTechs.findIndex((s) => s.employeeId === employee?.id)

  // قرار (ع) 10-08: المبيعات يتقيّمون على الحجوزات الي رحّلوها للشعبة الهندسية،
  // والالتزام بالحضور، والمهام الإضافية — بس. بلا نقاط KPI ولا حجوزات منجزة.
  // وكل دور يتقيّم على شغله هو: المراقب على تدقيقه (قرار 10-08).
  // (ل) 11-12 دازاي: «المصمم جاي يتحاسب على عدد الحجوزات — هذا غلط، كل واحد يتقيّم حسب شغله».
  // الحجوزات المنجزة ومعدل إنجازها شغل الفنيين والليدرية بس. باقي الأدوار (تصميم، حسابات،
  // كوادر، IT، جودة…) يترتّبون على تقييم ماتركس — الي يحسب شغل كل دور بقواعده هو.
  const fieldTrack = !!activeTrack?.strictRoles?.includes('TECHNICIAN')
  const matrixBoard = {
    cols: [
      { label: 'تقييم ماتركس (شغل دورك)', icon: '🎯', get: (e: KpiLeaderboardEntry) => Math.round(peers.get(e.employeeId) ?? 0) },
      { label: 'المهام الإضافية', icon: '📋', get: (e: KpiLeaderboardEntry) => e.extraTasksDone ?? 0 },
    ],
    explain: [
      ['🎯 تقييم ماتركس', 'محسوب على شغل دورك أنت — مو على الحجوزات: الدوام، تقرير الإنجاز اليومي، مهامك بمواعيدها، ومحطات شغلك (مثلاً المصمم: أقسام التصاميم سلّمها بموعدها، الحسابات: اعتماد الفواتير، الكوادر: توجيه الكوادر)، + تقييم مسؤوليك.'] as [string, string],
      ['📋 المهام الإضافية', 'المهام الي توجّهت إلك وخلّصتها.'] as [string, string],
      ['🕐 الالتزام بالدوام', 'من بصمات حضورك وانصرافك.'] as [string, string],
    ],
    sortNote: 'الترتيب على تقييم ماتركس لشغل دورك، وبعدها المهام، وعند التساوي الالتزام بالدوام.',
  }
  const custom = (activeTrack?.strictRoles?.length === 1 ? CUSTOM_BOARDS[activeTrack.strictRoles[0]] : undefined)
    ?? (activeTrack && !fieldTrack ? matrixBoard : undefined)
  const isSales = !!custom
  const rawList = board ? (period === 'weekly' ? board.weekly : board.monthly) : []
  const list = custom
    ? [...rawList].sort((a, b) => {
      for (const c of custom.cols) { const d = c.get(b) - c.get(a); if (d) return d }
      return b.attendedDays - a.attendedDays
    })
    // قرار (ع) 10-08: الترتيب على تقييم ماتركس (يحسب الشغل الفعلي) بدل نقاط KPI اليدوية.
    : [...rawList].sort((a, b) => (peers.get(b.employeeId) ?? -1) - (peers.get(a.employeeId) ?? -1) || b.points - a.points)
  const myIndex = list.findIndex((e) => e.employeeId === employee?.id)
  const myEntry = myIndex >= 0 ? list[myIndex] : null

  // ── أرقام الموظف بالفترة المختارة ──
  const myRank = myIndex >= 0 ? myIndex + 1 : null
  // اسم المجموعة = المسار (الشغل)، وإذا ماكو مسار نرجع لاسم الدور
  const roleLabel = activeTrack ? activeTrack.label : (employee ? arabicRole(employee.role) : '')
  // معدل الإنجاز: المنجز من الي انكلّف بيه. الي خلّص ٨ من ٨ مو مثل
  // الي خلّص ٨ من ٢٠ — والرقم المطلق لحاله يخفي هذا الفرق.
  const completionRate = myEntry && myEntry.assignedBookings > 0
    ? Math.round((myEntry.completedBookings / myEntry.assignedBookings) * 100)
    : null
  const completionWhy = myEntry && completionRate !== null ? (
    <PctWhy title="معدل الإنجاز" parts={{ numLabel: 'حجز خلّصته', num: myEntry.completedBookings, denLabel: 'حجز انكلّفت بيه', den: myEntry.assignedBookings }}
      meaning={<>شكد من الحجوزات الي انكلّفت بيها بالفترة خلّصتها. ٨٠٪ وفوگ زين. الحجز الي بعده مفتوح ينزّلها، والملغي ما ينحسب.</>}
      source={<>جدول الحجوزات: المكلّف = الحجوزات الي انت بكادرها (أو ليدرها) وانسجلت بالفترة بلا الملغية والمؤرشفة؛ المنجز = الي اكتملت بالفترة.</>} />
  ) : '—'
  const periodDays = period === 'weekly' ? 7 : 30
  const periodLabel = period === 'weekly' ? 'الأسبوع الماضي' : 'الشهر الماضي'
  const top3 = list.slice(0, 3)

  return (
    <div dir="rtl" className="mr-page space-y-5">
      {/* تصميم (ع) 10-13: العنوان والتبويبات بشريط فاتح وحد */}
      <div className="mr-hero space-y-4">
      {/* ═══ العنوان + التبديل ═══ */}
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="flex items-center gap-3">
          <span className="mr-logo"><LineIcon n="trophy" className="h-6 w-6" /></span>
          <div>
            <h2 className="mr-title text-2xl font-black">التقييم</h2>
            <p className="mr-sub text-xs">
              {effectiveView === 'mine' ? `تقييمي وتصنيفي — مقارنة بين الي يشتغلون ${roleLabel}`
                : effectiveView === 'team' ? 'قيّم فريقك عن كل حجز'
                : effectiveView === 'kpi' ? 'خصم وإضافة نقاط الكي بي اي على بنودها الثمانية'
                : 'مقارنة نشاط المراجعة بين الإداريين'}
            </p>
          </div>
        </div>
        <div className="relative flex gap-2">
          {effectiveView === 'mine' && (['monthly', 'weekly'] as const).map((p) => (
            <button
              key={p}
              onClick={() => setPeriod(p)}
              className={`mr-btn ${period === p ? 'on' : ''}`}
            >
              <LineIcon n="cal" /> {p === 'weekly' ? 'أسبوعي' : 'شهري'}
            </button>
          ))}
        </div>
      </div>

      {/* ═══ خيارات من فوگ ═══
          نفس نمط «مهامي» و«الجرد»: بند واحد بالقائمة والاختيار هنا. */}
      {(canReviewTeam || canSeeEvaluators || canKpi) && (
        <div className="mr-tabs">
          {([
            ...(isTopAdmin ? [] : [{ k: 'mine' as const, label: '🏆 تقييمي وتصنيفي' }]),
            ...(canReviewTeam ? [{ k: 'team' as const, label: '👥 تقييم الموظفين' }] : []),
            ...(canSeeEvaluators ? [{ k: 'evaluators' as const, label: '📋 تقييم بين الإداريين' }] : []),
            ...(canKpi ? [{ k: 'kpi' as const, label: '🎯 تقييم كي بي اي' }] : []),
          ]).map((o) => (
            <button
              key={o.k}
              onClick={() => setView(o.k)}
              className={`mr-btn ${effectiveView === o.k ? 'on' : ''}`}
            >
              {o.label}
            </button>
          ))}
        </div>
      )}
      </div>

      {/* تقييم الفريق يعيد استعمال نفس الشاشة بلا نسخ منطقها */}
      {effectiveView === 'team' && <PerformanceReviewPage embedded />}
      {effectiveView === 'kpi' && <KpiAdministrativeTab />}

      {/* ═══ تقييم بين الإداريين ═══
          ⚠️ «نقاط»/«حجوزات مراجعة» هنا **عدد المراجعات المسجَّلة**
          لا نقاط KPI ولا حجوزات مُنجزة — لا معنى لـ«الالتزام بالدوام»
          أو «معدل الإنجاز» لموظف بصفته مقيِّماً، فهذا الجدول أبسط
          من جدول «تقييمي وتصنيفي» بقصد. وما نعرض «دقة التقييم»
          الظاهرة بالتصميم لأنها **غير معرَّفة بالنظام** — لا يوجد
          مقياس فعلي لصحة حكم المقيِّم، وعرض رقم مخترع أسوأ من عدم عرضه. */}
      {effectiveView === 'evaluators' && (
        <div className="space-y-4">
          <div className="flex justify-end gap-2">
            {(['monthly', 'weekly'] as const).map((p) => (
              <button
                key={p}
                onClick={() => setEvalPeriod(p)}
                className={`rounded-lg px-4 py-2 text-xs font-bold transition ${
                  evalPeriod === p ? 'bg-[#2c5aad] text-white shadow-md' : 'border border-slate-300 bg-white text-slate-600 hover:bg-slate-50'
                }`}
              >
                📅 {p === 'weekly' ? 'أسبوعي' : 'شهري'}
              </button>
            ))}
          </div>

          {(() => {
            const evalList = evalBoard ? (evalPeriod === 'weekly' ? evalBoard.weekly : evalBoard.monthly) : []
            const myEvalIndex = evalList.findIndex((e) => e.employeeId === employee?.id)
            const myEvalEntry = myEvalIndex >= 0 ? evalList[myEvalIndex] : null
            return (
              <>
                <div className="grid grid-cols-2 gap-3 lg:grid-cols-2">
                  <RankCard
                    icon="cal" label={evalPeriod === 'weekly' ? 'حجوزات راجعتها هالأسبوع' : 'حجوزات راجعتها هالشهر'}
                    value={myEvalEntry ? String(myEvalEntry.points) : '—'}
                    tone="sky"
                    delta={myEvalEntry?.pointsDelta ?? 0}
                    deltaSuffix="حجز"
                  />
                  <RankCard
                    icon="trophy" label="ترتيبي الحالي"
                    value={myEvalIndex >= 0 ? `#${myEvalIndex + 1}` : '—'}
                    tone="emerald"
                    delta={myEvalEntry?.rankDelta ?? 0}
                    deltaSuffix="مركز"
                    note={myEvalIndex >= 0 ? `من ${evalList.length}` : undefined}
                  />
                </div>

                <div className="overflow-hidden rounded-2xl border border-slate-200 bg-white shadow-sm">
                  <div className="border-b border-slate-100 px-5 py-4">
                    <h3 className="text-sm font-extrabold text-[#0f2040]">👥 ترتيب الإداريين حسب نشاط المراجعة</h3>
                    <p className="mt-1 text-xs text-slate-500">الرقم = كم حجز راجعته وقيّمته {evalPeriod === 'weekly' ? 'هالأسبوع' : 'هالشهر'}. الأكثر مراجعة يطلع أول.</p>
                  </div>
                  <div className="overflow-x-auto">
                    <table className="w-full text-right text-xs">
                      <thead className="bg-slate-50 text-slate-600">
                        <tr>
                          <th className="px-4 py-2.5 font-bold">#</th>
                          <th className="px-4 py-2.5 font-bold">الإداري</th>
                          <th className="px-4 py-2.5 font-bold">كم حجز راجع</th>
                          <th className="px-4 py-2.5 font-bold">مقارنة بالفترة الي فاتت</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-slate-100">
                        {evalList.map((e, i) => {
                          const isMe = e.employeeId === employee?.id
                          return (
                            <tr key={e.employeeId} className={isMe ? 'bg-sky-50/60' : 'hover:bg-slate-50'}>
                              <td className="px-4 py-3">
                                <span className={`inline-flex h-6 w-6 items-center justify-center rounded-md text-[11px] font-black ${
                                  i === 0 ? 'bg-amber-100 text-amber-800'
                                  : i === 1 ? 'bg-slate-200 text-slate-700'
                                  : i === 2 ? 'bg-orange-100 text-orange-800'
                                  : 'text-slate-400'
                                }`}>{i + 1}</span>
                              </td>
                              <td className="px-4 py-3 font-bold text-slate-800">
                                {e.employeeName}{isMe && <span className="mr-1 text-[10px] text-sky-600">(أنت)</span>}
                              </td>
                              <td className="px-4 py-3 font-black text-slate-800">{e.points}</td>
                              <td className="px-4 py-3">
                                {e.pointsDelta === 0 ? (
                                  <span className="text-slate-400">نفس الفترة الي فاتت</span>
                                ) : (
                                  <span className={`font-bold ${e.pointsDelta > 0 ? 'text-emerald-600' : 'text-red-500'}`}>
                                    {e.pointsDelta > 0 ? '▲ أكثر' : '▼ أقل'} بـ{Math.abs(e.pointsDelta)}
                                  </span>
                                )}
                              </td>
                            </tr>
                          )
                        })}
                        {evalList.length === 0 && (
                          <tr><td colSpan={4} className="p-6 text-center text-slate-400">ماكو مراجعات مسجّلة بهذي الفترة</td></tr>
                        )}
                      </tbody>
                    </table>
                  </div>
                </div>
              </>
            )
          })()}
        </div>
      )}

      {effectiveView === 'mine' && (<>
      {/* ═══ مسارات التصنيف ═══
          الموظف الي يشتغل أكثر من شغلة يظهر بأكثر من تصنيف — وهذا
          واقعه، مو خلل. التصنيف الواحد حسب الدور كان يخفيه. */}
      {myTracks.length > 1 && (
        <div className="flex flex-wrap gap-2">
          {myTracks.map((t, i) => (
            <button
              key={t.permission}
              onClick={() => setTrackIdx(i)}
              className={`rounded-xl px-3 py-1.5 text-xs font-bold transition ${
                i === trackIdx
                  ? 'bg-[#0f2040] text-white shadow-md'
                  : 'border border-slate-300 bg-white text-slate-600 hover:bg-slate-50'
              }`}
            >
              {t.icon} {t.label}
            </button>
          ))}
        </div>
      )}

      <p className="mr-note"><LineIcon n="info" className="h-4 w-4 shrink-0" /><span>
        {activeTrack?.strictRoles?.length
          ? <>ⓘ الترتيب بين <b>{activeTrack.label}</b> فقط (عائلة أدوار محدَّدة) — بلا خلط مع أدوار ثانية عندها نفس الصلاحية.</>
          : <>ⓘ الترتيب بين الي يشتغلون <b>نفس الشغل</b> ({roleLabel}) — مو حسب مسمّى الدور.</>}
        {' '}آخر {periodDays} يوم، والمقارنة مع {periodLabel}.
        {myTracks.length > 1 && ` وإنت تشتغل ${myTracks.length} شغلات، فعندك ${myTracks.length} تصنيفات.`}
      </span></p>

      {/* ═══ البطاقات الأربع ═══ */}
      <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
        <RankCard
          icon="trophy" label="ترتيبي الحالي"
          value={myRank ? `#${myRank}` : '—'}
          tone="emerald"
          delta={myEntry?.rankDelta ?? 0}
          deltaSuffix="مركز"
          note={myRank === 1 ? '🎉 حافظ على مركزك!' : myRank ? `من ${list.length}` : undefined}
        />
        {isSales ? (<>
        {custom!.cols.slice(0, 2).map((c, i) => (
          <RankCard key={c.label} icon={i === 0 ? 'cal' : 'clip'} label={c.label} value={myEntry ? String(c.get(myEntry)) : '—'} tone={i === 0 ? 'sky' : 'amber'} />
        ))}
        <RankCard icon="clock" label="أيام الحضور" value={myEntry ? String(myEntry.attendedDays) : '—'} tone="violet" />
        </>) : (<>
        <RankCard
          icon="briefcase" label="تقييم ماتركس"
          note={period === 'weekly' ? 'للشهر كله' : undefined}
          value={<MatrixPct v={myScore?.final ?? (employee ? peers.get(employee.id) : null)} s={myScore} />}
          tone="amber"
          bar={(myScore?.final ?? (employee ? peers.get(employee.id) : null)) ?? undefined}
        />
        <RankCard
          icon="cal" label="حجوزات منجزة"
          value={myEntry ? String(myEntry.completedBookings) : '—'}
          tone="sky"
          note={myEntry && myEntry.assignedBookings > 0 ? `من ${myEntry.assignedBookings} مكلّف بيها` : undefined}
        />
        <RankCard
          icon="trendUp" label="معدل الإنجاز"
          value={completionWhy}
          tone="violet"
          note={completionRate === null ? 'ماكو حجوزات بالفترة' : undefined}
          bar={completionRate ?? undefined}
        />
        </>)}
      </div>

      {/* ═══ أعلى ٣ ═══ */}
      {top3.length === 0 && (
        <p className="mr-card mr-sub px-4 py-6 text-center text-xs">
          ماكو تصنيف بهذي الفترة بعد — التصنيف يبني نفسه من الحجوزات المنجزة والتقييمات.
        </p>
      )}
      {top3.length > 0 && (
        <div className="mr-podium p-5">
          <h3 className="mr-h mb-4 text-sm font-extrabold">
            👑 أعلى ٣ {roleLabel} {period === 'weekly' ? 'هذا الأسبوع' : 'هذا الشهر'}
            {period === 'weekly' && !custom && <span className="mr-sub mr-2 text-[11px] font-bold">(النسبة لتقييم الشهر كله)</span>}
          </h3>
          {/* ═══ المنصّة ═══
              ⚠️ بالموبايل تنقلب **صفوف** مو ثلاث بطاقات جنب بعض.
              ثلاث بطاقات على شاشة ٣٦٠ بكسل تعني ١٠٠ بكسل للوحدة —
              والاسم العربي ينقص لحرفين ونقاط، فالشاشة تصير بلا فايدة:
              تشوف ترتيب بلا ما تعرف منو.
              وترتيب العرض يختلف: بالموبايل ١ ثم ٢ ثم ٣ (من فوگ لتحت
              مثل أي قائمة)، وبالكمبيوتر الأول بالوسط مثل المنصّة. */}
          <div className="flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-center sm:gap-3">
            {[top3[2], top3[0], top3[1]].map((e, slot) => {
              if (!e) return <div key={slot} className="hidden sm:block sm:flex-1" />
              const place = e === top3[0] ? 1 : e === top3[1] ? 2 : 3
              const placeLabel = place === 1 ? 'الأول' : place === 2 ? 'الثاني' : 'الثالث'
              const isMe = e.employeeId === employee?.id
              const orderCls = place === 1 ? 'order-1' : place === 2 ? 'order-2' : 'order-3'
              return (
                <div
                  key={e.employeeId}
                  className={`${orderCls} mr-pod ${place === 1 ? 'first' : ''} flex items-center gap-2.5 p-3 sm:order-none sm:block sm:flex-1 sm:p-3 sm:text-center ${isMe ? 'ring-2 ring-sky-300' : ''}`}
                >
                  <div className="shrink-0 sm:mb-1"><span className={`mr-medal m${place}`}>{place}</span></div>
                  {/* الاسم بلا قصّ: يلتف بسطرين إذا احتاج */}
                  <p className="mr-h min-w-0 flex-1 break-words text-right text-xs font-bold leading-tight sm:text-center">
                    {e.employeeName}
                    {isMe && <span className="mr-1 text-[10px] text-sky-600">(أنت)</span>}
                  </p>
                  <div className="shrink-0 text-left sm:text-center">
                    <p className="mr-h text-2xl font-black leading-none sm:mt-1">
                      {custom ? custom.cols[0].get(e) : <MatrixPct v={peers.get(e.employeeId)} who={e.employeeName} />}
                    </p>
                    {!isSales && e.pointsDelta !== 0 && (
                      <p className={`text-[10px] font-bold ${e.pointsDelta > 0 ? 'text-emerald-600' : 'text-red-500'}`}>
                        {e.pointsDelta > 0 ? '▲' : '▼'} {Math.abs(e.pointsDelta)}
                      </p>
                    )}
                  </div>
                  <span className={`mr-place ${place === 1 ? 'first' : ''} shrink-0 sm:mt-2`}>
                    {placeLabel}
                  </span>
                </div>
              )
            })}
          </div>
        </div>
      )}

      {/* ═══ الجدول + شرح النقاط ═══ */}
      <div className="grid grid-cols-1 gap-4 lg:grid-cols-3">
        {/* الترتيب الكامل */}
        <div className="mr-card overflow-hidden lg:col-span-2">
          <div className="px-5 py-4">
            <h3 className="mr-h flex items-center gap-2 text-sm font-extrabold"><LineIcon n="users" className="h-5 w-5" /> ترتيب {roleLabel}
              <HelpDot lines={custom ? custom.explain.map(([t, d]) => `${t}: ${d.split(' — ')[0]}`) : [
                '🤖 تقييم ماتركس: يقيّم شغلك تلقائياً — شنو سوّيت بوقته، وشنو تركت، والأخطاء.',
                '🗂️ الحجوزات المنجزة: الي انكلّفت بيها وخلصت.',
                '🕐 الالتزام: أيام حضورك من أيام الفترة.',
                'الترتيب على تقييم ماتركس.',
              ]} />
            </h3>
          </div>
          <div className="overflow-x-auto px-3 pb-3">
            <table className="mr-table w-full text-right text-xs">
              <thead className="mr-thead">
                <tr>
                  <th className="rounded-s-lg px-4 py-2.5 font-bold">#</th>
                  <th className="px-4 py-2.5 font-bold">الموظف</th>
                  {isSales ? (<>
                    {custom!.cols.map((c) => <th key={c.label} className="px-4 py-2.5 font-bold">{c.label}</th>)}
                    <th className="px-4 py-2.5 font-bold">الالتزام</th>
                  </>) : (<>
                    <th className="px-4 py-2.5 font-bold">تقييم ماتركس{period === 'weekly' && ' (للشهر)'}</th>
                    <th className="px-4 py-2.5 font-bold">الحجوزات المنجزة</th>
                    <th className="px-4 py-2.5 font-bold">الالتزام</th>
                    <th className="rounded-e-lg px-4 py-2.5 font-bold">مقارنة بالفترة الي فاتت</th>
                  </>)}
                </tr>
              </thead>
              <tbody>
                {list.map((e, i) => {
                  const isMe = e.employeeId === employee?.id
                  // الالتزام = أيام الحضور من أيام الفترة (عدا الجمع)
                  const workDays = period === 'weekly' ? 7 : 30 // ماكو عطلة أسبوعية (قرار 10-08)
                  const rate = Math.min(100, Math.round((e.attendedDays / workDays) * 100))
                  const commit = rate >= 90
                    ? { text: 'ممتاز', cls: 'mr-pill-ok' }
                    : rate >= 75
                      ? { text: 'جيد جداً', cls: 'bg-sky-50 text-sky-700' }
                      : rate >= 50
                        ? { text: 'جيد', cls: 'bg-amber-50 text-amber-700' }
                        : { text: 'يحتاج تحسين', cls: 'bg-red-50 text-red-700' }
                  return (
                    <tr key={e.employeeId} className={isMe ? 'me' : ''}>
                      <td className="px-4 py-3">
                        <span className={`mr-rank ${isMe ? 'me' : i < 3 ? `r${i + 1}` : ''}`}>{i + 1}</span>
                      </td>
                      <td className="px-4 py-3 font-bold">
                        {e.employeeName}{isMe && <span className="mr-1 text-[10px] text-sky-600">(أنت)</span>}
                      </td>
                      {isSales ? (<>
                        {custom!.cols.map((c, ci) => <td key={c.label} className={`px-4 py-3 ${ci === 0 ? 'font-black' : ''}`}>{e.limited ? '—' : c.get(e)}</td>)}
                      </>) : (<>
                        <td className="px-4 py-3 font-black"><MatrixPct v={peers.get(e.employeeId)} who={e.employeeName} /></td>
                        <td className="px-4 py-3">{e.limited ? '—' : e.completedBookings}</td>
                      </>)}
                      <td className="px-4 py-3">
                        {e.limited ? <span className="text-slate-400">—</span> : (
                          <span className={`${commit.cls === 'mr-pill-ok' ? '' : 'rounded-md px-2 py-0.5 text-[10px] font-bold'} ${commit.cls}`} title={`${e.attendedDays} يوم حضور`}>
                            {commit.text}
                          </span>
                        )}
                      </td>
                      {!isSales && <td className="px-4 py-3">
                        {e.pointsDelta === 0 ? (
                          <span className="text-slate-400">نفس الفترة الي فاتت</span>
                        ) : (
                          <span className={`font-bold ${e.pointsDelta > 0 ? 'text-emerald-600' : 'text-red-500'}`}>
                            {e.pointsDelta > 0 ? '▲ أكثر' : '▼ أقل'} بـ{Math.abs(e.pointsDelta)}
                          </span>
                        )}
                      </td>}
                    </tr>
                  )
                })}
                {list.length === 0 && (
                  <tr><td colSpan={6} className="p-6 text-center text-slate-400">ماكو بيانات تقييم بعد</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </div>

        {/* ═══ كيف تنحسب النقاط — الحقيقة مو معادلة مزيّنة ═══ */}
        <div className="space-y-4">
          {/* قرار (ع) 10-08: ماتركس يقيّم كل موظف ويشوف شغله — شكد سوّى وشكد ترك وشنو الأخطاء. */}
          <MatrixMine score={myScore} />

          {/* إنجازاتك بالفترة */}
          {myEntry && (
            <div className="mr-card p-5">
              <h3 className="mr-h mb-3 text-sm font-extrabold">
                🏅 إنجازاتك {period === 'weekly' ? 'هذا الأسبوع' : 'هذا الشهر'}
              </h3>
              <div className="grid grid-cols-3 gap-2">
                <Achievement
                  icon="check" label="أيام الحضور"
                  value={`${myEntry.attendedDays}`}
                  ok={myEntry.attendedDays > 0}
                />
                {isSales ? (<>
                {custom!.cols.slice(0, 2).map((c) => (
                  <Achievement key={c.label} icon="bars" label={c.label} value={`${c.get(myEntry)}`} ok={c.get(myEntry) > 0} />
                ))}
                </>) : (<>
                <Achievement
                  icon="bolt" label="معدل الإنجاز"
                  value={completionWhy}
                  ok={(completionRate ?? 0) >= 80}
                />
                <Achievement
                  icon="bars" label="تقييمات"
                  value={`${myEntry.evaluationCount}`}
                  ok={myEntry.points >= 0}
                />
                </>)}
              </div>
            </div>
          )}
        </div>
      </div>
      {isTechnician && (
        <div className={`grid grid-cols-1 gap-4 ${stats ? 'sm:grid-cols-2' : ''}`}>
          <div className="mr-level flex items-center justify-between gap-4 p-6">
            <div>
            <p className="sub text-sm">المستوى (حسب المهارات المعتمدة)</p>
            <p className="mt-1 text-3xl font-extrabold">
              {currentLevel.level} - {currentLevel.label}
            </p>
            <p className="sub mt-2 text-sm">
              {skillCount} مهارة معتمدة
              {nextLevel
                ? ` - يحتاج ${nextLevel.min - skillCount} مهارة إضافية للترقي إلى "${nextLevel.label}"`
                : ' - وصلت لأعلى مستوى!'}
            </p>
            </div>
            <Emblem />
          </div>

          {/* (ك) 11-26: الرانك من «إحصائيات الموظفين» — بلا الصلاحية چان يطلع «0 حجز · رانك 1» دايماً، فما نعرضه بلا أرقام حقيقية */}
          {stats && <div className="mr-card p-6">
            <h3 className="mr-h mb-2 flex items-center gap-2 text-sm font-extrabold"><LineIcon n="chat" className="h-5 w-5" /> ملاحظات الأداء</h3>
            <div className="flex items-center gap-4">
            <div className="text-center"><p className="mr-sub text-sm">الرانك</p>
            <p className="mr-val emerald mt-1 text-3xl font-extrabold">{rank}</p></div>
            <div className="mr-divider ps-4">
            <p className="mr-sub text-sm">
              {completedCount} حجز منجز - يحتاج {remainingForNextRank} حجز إضافي للصعود للرانك التالي
            </p>
            {myPosition >= 0 && (
              <p className="mr-sub mt-2 text-sm">
                ترتيبك بين الفنيين: <span className="font-bold text-brand-700">#{myPosition + 1}</span> من{' '}
                {sortedTechs.length}
              </p>
            )}
            </div>
            </div>
          </div>}
        </div>
      )}

      </>)}
    </div>
  )
}

/* ───── بطاقة رقم بأعلى الصفحة ───── */

function RankCard({ icon, label, value, tone, delta, deltaSuffix, note, bar }: {
  icon: string; label: string; value: ReactNode
  tone: 'emerald' | 'amber' | 'sky' | 'violet'
  delta?: number; deltaSuffix?: string; note?: string; bar?: number
}) {
  return (
    <div className="mr-card p-4">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="mr-h text-base font-extrabold">{label}</p>
          <p className={`mr-val ${tone} mt-1 text-3xl font-black`}>{value}</p>
          {delta !== undefined && delta !== 0 && (
            <p className={`text-[11px] font-bold ${delta > 0 ? 'text-emerald-600' : 'text-red-500'}`}>
              {/* (دازاي 11-27) «-103 مراجعة» ما يفهمها الموظف — جملة واضحة بدلها */}
              {deltaSuffix === 'مركز'
                ? `${delta > 0 ? '▲ صعدت' : '▼ نزلت'} ${Math.abs(delta)} ${Math.abs(delta) === 1 ? 'مركز' : 'مراكز'} عن الفترة الي فاتت`
                : `${delta > 0 ? '▲ أكثر' : '▼ أقل'} بـ${Math.abs(delta)} ${deltaSuffix ?? ''} عن الفترة الي فاتت`}
            </p>
          )}
          {note && <p className="mr-h mt-0.5 text-xs">{note}</p>}
        </div>
        <span className={`mr-tile ${tone}`}><LineIcon n={icon} className="h-5 w-5" /></span>
      </div>
      {bar !== undefined && (
        <div className="mr-bar mt-3">
          <div className={tone} style={{ width: `${Math.max(0, Math.min(100, bar))}%` }} />
        </div>
      )}
    </div>
  )
}

/* ───── شعار المستوى: نجمة بإكليل غار ───── */

function Emblem() {
  const leaves = [0, 1, 2, 3, 4, 5]
  return (
    <svg viewBox="0 0 100 100" className="h-20 w-20 shrink-0" aria-hidden="true" fill="none" stroke="#e6ecff" strokeWidth="2">
      <circle cx="50" cy="52" r="24" />
      <circle cx="50" cy="52" r="19" strokeWidth="1.2" />
      <path d="m50 39 3.8 7.8 8.6 1.2-6.2 6 1.5 8.5L50 58.5l-7.7 4 1.5-8.5-6.2-6 8.6-1.2L50 39Z" fill="#e6ecff" stroke="none" />
      <path d="m50 12 1.8 3.6 4 .6-2.9 2.8.7 4L50 21.1 46.4 23l.7-4-2.9-2.8 4-.6L50 12Z" fill="#e6ecff" stroke="none" />
      {[-1, 1].map((side) => leaves.map((i) => {
        // ورق الغار على قوس حول الدائرة، كل ورقة مايلة ويا القوس
        const deg = 105 + i * 21
        const a = deg * Math.PI / 180
        const lx = 50 + Math.cos(a) * 33, y = 52 + Math.sin(a) * 33
        const x = side === -1 ? lx : 100 - lx
        const rot = side === -1 ? deg - 180 + 30 : 180 - deg - 30
        return <ellipse key={`${side}${i}`} cx={x} cy={y} rx="2.8" ry="6" fill="#e6ecff" stroke="none"
          transform={`rotate(${rot} ${x} ${y})`} />
      }))}
      <path d="M36 84q14 6 28 0" strokeWidth="3" />
    </svg>
  )
}

/* ───── إنجاز صغير ───── */

function Achievement({ icon, label, value }: {
  icon: string; label: string; value: ReactNode; ok: boolean
}) {
  return (
    <div className="mr-ach">
      <div className={`ic ${icon} flex justify-center`}><LineIcon n={icon} className="h-5 w-5" /></div>
      <p className={`v ${icon} mt-0.5 text-sm font-black`}>{value}</p>
      <p className="mr-sub text-[11px] leading-tight">{label}</p>
    </div>
  )
}
