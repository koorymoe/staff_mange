import { useCallback, useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  api, type GroupPerformance, type LateFocus, type MatrixBusiness as MatrixBusinessData, type MatrixFeedItem,
  type MatrixProposal, type MatrixTrendDay, type MatrixWatch,
} from '../api'
import MatrixBusiness from './MatrixBusiness'
import OwnerSwitch from './OwnerSwitch'
import { evidenceCards } from './matrixEvidence'
import { SWITCH_MATRIX_STAFF_EYE } from '../systemSwitches'
import MatrixEyeGraphic from './MatrixEyeGraphic'
import { GroupPerf } from './MatrixRoleEyes'
import { GROUP_LABEL, eyeColor, type EyeGroup, type EyeMood } from './matrixEyeColors'

// ═══ مركز قيادة ماتركس — شاشة المدير (تصميم (ع)) ═══
// كل رقم من بيانات حقيقية؛ لو ما تكفي نگول «بيانات قليلة» بدل رقم وهمي.
// «تنفيذ» = موافقة على اقتراح معلّق (نفس صندوق القرارات) — ماكو شي ينفّذ بلا ضغطة.

type GroupRep = { group: string; employees: MatrixWatch[]; red: number; alert: number }
type Focus = 'late' | 'perf' | 'profit'

const card = 'rounded-2xl border border-[var(--mx-border)] bg-[var(--mx-card)] shadow-[var(--mx-glow)]'
const fmtIQD = (n: number) => `${Math.round(n).toLocaleString('en-US')} د.ع`

function groupMood(g: GroupRep): EyeMood {
  if (g.red > 0) return 'ANGRY'
  if (g.alert > 0) return 'ALERT'
  return g.employees.length > 0 && g.employees.every((e) => e.mood === 'PLEASED') ? 'PLEASED' : 'CALM'
}

function since(at: string) {
  const m = Math.max(0, Math.round((Date.now() - new Date(at).getTime()) / 60000))
  if (m < 60) return `قبل ${m} د`
  if (m < 1440) return `قبل ${Math.round(m / 60)} س`
  return `قبل ${Math.round(m / 1440)} يوم`
}

// مؤشر الأداء للعرض بس — مو نقاط ولا تقييم رسمي.
// نسبة الإنجاز (الجزئي بنص) − غياب/تأخير اليوم − البطء عن المتوقع.
// ⚠️ چان «١٠٠ − score×٣» فكل من عنده حجوزات مفتوحة يطلع ٠٪ «يحتاج دعم».
type Member = GroupPerformance['members'][number]
// غير الميدانيين: ما ينقاسون بحجوزات ولا سرعة — شغل دورهم بآخر ٣٠ يوم موجود
// لو لا، وحضورهم على جدولهم الحقيقي بس (بلا جدول = بلا خصم).
function perfIndex(m: Member, field = true) {
  if (!field) {
    const done = (m.metrics ?? []).reduce((a, x) => a + x.value, 0)
    let d = done > 0 ? 90 : 55
    if (m.absent) d -= 15
    else if (m.late > 0) d -= Math.min(20, m.late / 3)
    return Math.max(0, Math.min(100, Math.round(d)))
  }
  let v = m.jobs > 0 ? ((m.completed + m.partial * 0.5) / m.jobs) * 100 : 90
  if (m.absent) v -= 15
  else if (m.late > 0) v -= Math.min(20, m.late / 3)
  if (m.speed != null && m.speed > 1.15) v -= Math.min(25, (m.speed - 1) * 40)
  return Math.max(0, Math.min(100, Math.round(v)))
}
function perfStatus(i: number) {
  if (i >= 90) return { t: 'ممتاز', c: 'bg-emerald-500/15 text-[var(--mx-ok)]', bar: '#22c55e', dot: 'bg-emerald-500' }
  if (i >= 75) return { t: 'جيد', c: 'bg-sky-500/15 text-[var(--mx-info)]', bar: '#3b82f6', dot: 'bg-blue-500' }
  if (i >= 60) return { t: 'متابعة', c: 'bg-amber-500/15 text-[var(--mx-warn)]', bar: '#f59e0b', dot: 'bg-amber-500' }
  return { t: 'يحتاج دعم', c: 'bg-red-500/15 text-[var(--mx-bad)]', bar: '#ef4444', dot: 'bg-red-500' }
}
const GROUP_ICON: Record<string, string> = { TECHS: '🔧', LEADERS: '🧭', MONITORS: '🔍', COORDINATORS: '🗂️', FINANCE: '💵', DESIGN: '🎨', QUALITY: '✅', IT: '💻', ADMINS: '👑', STAFF: '🏢' }
function confidence(samples: number) { return samples >= 20 ? 87 : samples >= 8 ? 72 : 50 }

export default function MatrixCommandCenter() {
  const [groups, setGroups] = useState<GroupRep[] | null>(null)
  const [feed, setFeed] = useState<MatrixFeedItem[]>([])
  const [trend, setTrend] = useState<MatrixTrendDay[]>([])
  const [late, setLate] = useState<LateFocus | null>(null)
  const [biz, setBiz] = useState<MatrixBusinessData | null>(null)
  const [props, setProps] = useState<MatrixProposal[]>([])
  const [perf, setPerf] = useState<GroupPerformance[]>([])
  const [focus, setFocus] = useState<Focus>('late')
  const [sel, setSel] = useState<string | null>(null)
  const [err, setErr] = useState('')

  const loadAll = useCallback(() => {
    api.getRoleWatch().then((g) => {
      setGroups(g)
      Promise.all(g.filter((x) => x.employees.length > 0).map((x) => api.getGroupPerformance(x.group).catch(() => null)))
        .then((r) => setPerf(r.filter((x): x is GroupPerformance => !!x)))
    }).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر جلب ماتركس'))
    api.getMatrixFeed().then(setFeed).catch(() => {})
    api.getMatrixTrend().then(setTrend).catch(() => {})
    api.getLateFocus().then(setLate).catch(() => {})
    api.getMatrixBusiness().then(setBiz).catch(() => {})
    api.getMatrixProposals('PENDING').then(setProps).catch(() => {})
  }, [])
  useEffect(() => {
    loadAll()
    const t = setInterval(() => { api.getMatrixFeed().then(setFeed).catch(() => {}) }, 60000)
    return () => clearInterval(t)
  }, [loadAll])

  if (err) return <p className="rounded-lg bg-red-50 p-3 text-red-600">{err}</p>
  if (!groups) return <p className="text-[var(--mx-muted)]">ماتركس يجمع البيانات…</p>

  const active = groups.filter((g) => g.employees.length > 0)
  const people = active.reduce((n, g) => n + g.employees.length, 0)
  const red = active.reduce((n, g) => n + g.red, 0)
  const attention = active.reduce((n, g) => n + g.red + g.alert, 0)
  const tasks = active.reduce((n, g) => n + g.employees.reduce((m, e) => m + e.workload.reduce((k, w) => k + w.done + w.left, 0), 0), 0)
  const stability = people ? Math.round(((people - red) / people) * 100) : 100
  const worst: EyeMood = red > 0 ? 'ANGRY' : attention > 0 ? 'ALERT' : 'PLEASED'

  return (
    <div dir="rtl" className="space-y-4 rounded-3xl bg-[var(--mx-bg)] p-3 text-[var(--mx-text)] sm:p-5">
      {/* ── الرأس ── */}
      <div className="grid items-center gap-4 lg:grid-cols-[1fr_auto]">
        <div className="flex flex-wrap items-center gap-4">
          <div className="shrink-0 drop-shadow-[0_0_30px_rgba(56,189,248,0.6)]"><MatrixEyeGraphic group="IT" mood={worst === 'PLEASED' ? 'CALM' : worst} width={190} /></div>
          <div>
            <span className="mb-2 inline-flex items-center gap-1.5 rounded-full border border-emerald-400/30 bg-emerald-500/10 px-3 py-0.5 text-[11px] text-[var(--mx-ok)]">
              <span className="h-2 w-2 animate-pulse rounded-full bg-emerald-400" /> الذكاء التحليلي النشط الآن
            </span>
            <h2 className="text-3xl font-black text-[var(--mx-title)] sm:text-4xl">مركز قيادة <span className="bg-gradient-to-l from-sky-500 to-blue-600 bg-clip-text text-transparent">ماتركس</span></h2>
            <p className="mt-1 text-sm text-[var(--mx-muted)]">يراقب الأداء والموظفين والحجوزات والأرباح لحظياً</p>
          </div>
        </div>
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
          <Kpi value={`${stability}%`} label="استقرار النظام" tone="text-[var(--mx-ok)]" icon="〰️" />
          <Kpi value={attention} label="حالات تحتاج انتباه" tone="text-[var(--mx-bad)]" icon="⚠️" />
          <Kpi value={people} label="موظف تحت المراقبة" tone="text-[var(--mx-accent)]" icon="👥" />
          <Kpi value={tasks} label="مهمة اليوم" tone="text-[var(--mx-accent)]" icon="📦" />
        </div>
      </div>

      {/* مفتاح (ع): يطفي عين ماتركس عن الموظفين ويرجّعها وقت ما يريد */}
      <div className="text-slate-900"><OwnerSwitch switchKey={SWITCH_MATRIX_STAFF_EYE} label="عين ماتركس عند الموظفين" hint="إذا انطفت، العين وتوجيهاتها تختفي من شاشات الموظفين. إنت والمدير تبقى عندكم، وماتركس يكمّل يحلّل." /></div>

      {/* ── يركّز الآن على ── */}
      <div className={`${card} p-3`}>
        <p className="mb-2 text-sm font-extrabold text-[var(--mx-accent)]">🎯 ماتركس يركّز الآن على</p>
        <div className="grid gap-2 sm:grid-cols-3">
          <FocusTab on={focus === 'late'} onClick={() => setFocus('late')} icon="📅" title="الحجوزات المتأخرة" sub="رصد وتنبؤ بالأسباب" />
          <FocusTab on={focus === 'perf'} onClick={() => setFocus('perf')} icon="👥" title="أداء الموظفين" sub="متابعة الإنتاجية والالتزام" />
          <FocusTab on={focus === 'profit'} onClick={() => setFocus('profit')} icon="📊" title="الأرباح اليومية" sub="تحليل الأداء والفرص" />
        </div>
      </div>

      {/* ── اكتشف + البث ── */}
      <div className="grid gap-4 xl:grid-cols-[3fr_2fr]">
        <Discovery focus={focus} late={late} perf={perf} groups={active} biz={biz} proposal={props[0] ?? null} onDone={loadAll} />
        <Feed items={feed} />
      </div>

      {/* ── عيون ماتركس ── */}
      <div className={`${card} p-3`}>
        <p className="mb-3 text-sm font-extrabold text-[var(--mx-accent)]">👁️ عيون ماتركس <span className="text-[11px] font-normal text-[var(--mx-muted)]">كل عين تراقب مجموعة من الموظفين — اضغطها لتقريرها</span></p>
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 md:grid-cols-5 xl:grid-cols-9">
          {active.map((g) => {
            const grp = g.group as EyeGroup
            const mood = groupMood(g)
            const color = eyeColor(grp, mood)
            const bad = g.red > 0
            return (
              <button key={g.group} type="button" onClick={() => setSel(sel === g.group ? null : g.group)}
                className="flex flex-col items-center gap-1 rounded-xl border bg-[var(--mx-sunken)] p-2 text-center transition"
                style={{ borderColor: `color-mix(in srgb, ${color} ${sel === g.group ? 90 : 35}%, transparent)` }}>
                <MatrixEyeGraphic group={grp} mood={mood} width={70} />
                <span className="text-[12px] font-bold" style={{ color: `color-mix(in srgb, ${color} 72%, var(--mx-title))` }}>عن {GROUP_LABEL[grp] ?? g.group}</span>
                <span className="text-[11px] text-[var(--mx-muted)]">{g.employees.length} موظف</span>
                <span className={`rounded-full px-2 py-0.5 text-[10px] font-bold ${bad ? 'bg-red-500/15 text-[var(--mx-bad)]' : g.alert ? 'bg-amber-500/15 text-[var(--mx-warn)]' : 'bg-emerald-500/15 text-[var(--mx-ok)]'}`}>
                  ● {bad ? 'تحتاج انتباه' : g.alert ? 'منتبهة' : 'مستقر'}
                </span>
              </button>
            )
          })}
        </div>
      </div>
      {sel && <div><GroupPerf key={sel} group={sel} /></div>}

      {/* ── الصف الأخير ── */}
      <div className="grid gap-4 lg:grid-cols-2 2xl:grid-cols-4">
        <TrendChart days={trend} />
        <WatchList perf={perf} />
        <Predictions biz={biz} late={late} props={props} />
        <AskMatrix />
      </div>

      <details className="rounded-2xl border border-[var(--mx-border)] bg-[var(--mx-card)] text-[var(--mx-text)]">
        <summary className="cursor-pointer p-3 text-sm font-bold">💼 تفاصيل الأعمال والزبائن</summary>
        <div className="p-3 pt-0"><MatrixBusiness /></div>
      </details>
    </div>
  )
}

function Kpi({ value, label, tone, icon }: { value: string | number; label: string; tone: string; icon: string }) {
  return (
    <div className={`${card} min-w-[120px] p-3`}>
      <div className="flex items-start justify-between"><b className={`text-2xl ${tone}`}>{value}</b><span>{icon}</span></div>
      <p className="mt-1 text-[11px] text-[var(--mx-muted)]">{label}</p>
    </div>
  )
}

function FocusTab({ on, onClick, icon, title, sub }: { on: boolean; onClick: () => void; icon: string; title: string; sub: string }) {
  return (
    <button type="button" onClick={onClick}
      className={`flex items-center justify-between gap-2 rounded-xl border px-3 py-2 text-right transition ${on ? 'border-[var(--mx-tab-on-bd)] bg-[var(--mx-tab-on)]' : 'border-[var(--mx-border)] bg-[var(--mx-sunken)] hover:border-[var(--mx-accent)]'}`}>
      <span><b className={`block text-sm ${on ? 'text-[var(--mx-tab-on-tx)]' : 'text-[var(--mx-title)]'}`}>{title}</b><span className="text-[11px] text-[var(--mx-muted)]">{sub}</span></span>
      <span className="text-xl">{icon}</span>
    </button>
  )
}

// ── «ماتركس اكتشف» ──
function Discovery({ focus, late, perf, groups, biz, proposal, onDone }: {
  focus: Focus; late: LateFocus | null; perf: GroupPerformance[]; groups: GroupRep[]; biz: MatrixBusinessData | null; proposal: MatrixProposal | null; onDone: () => void
}) {
  const [modal, setModal] = useState<'sim' | 'evidence' | null>(null)
  const [busy, setBusy] = useState(false)

  const view = useMemo(() => {
    if (focus === 'late') {
      if (!late) return null
      const ch = late.changePct
      return {
        title: late.recent.late === 0 ? 'ماكو حجوزات متأخرة بآخر ٣ أيام ✅'
          : ch != null && ch > 0 ? `زيادة ملحوظة في الحجوزات المتأخرة خلال ٣ أيام الأخيرة` : `${late.recent.late} حجز متأخر بآخر ٣ أيام`,
        sub: `تأخّر ${late.recent.late} من ${late.recent.total} حجز (${late.recentPct}%)` + (ch != null ? ` — ${ch > 0 ? 'ارتفاع' : 'انخفاض'} ${Math.abs(ch)}% مقارنة بالـ٣ أيام الي قبلها (${late.prevPct}%).` : '.'),
        cause: late.cause, impact: late.impact, suggestion: late.suggestion, severity: late.severity, link: '/bookings',
        evidence: { 'حجوزات آخر ٣ أيام': late.recent.total, 'متأخرة': late.recent.late, 'مفتوحة وموعدها فات': late.recent.openNow, 'بلا كادر': late.recent.unstaffed, 'جزئية': late.recent.partial, 'نسبة التأخير قبلها': `${late.prevPct}%` } as Record<string, unknown>,
      }
    }
    if (focus === 'perf') {
      const p = perf.flatMap((g) => g.problems.map((x) => ({ ...x, group: g.group })))[0]
      const members = perf.flatMap((g) => g.members)
      const slow = members.filter((m) => m.speed != null && m.speed >= 1.4).length
      if (!p) return { title: 'أداء الكادر مستقر ✅', sub: `${members.length} موظف — ماكو مشكلة بارزة هسه.`, cause: '—', impact: '—', suggestion: 'كمّلوا بنفس الوتيرة.', severity: 'LOW' as const, link: '', evidence: { 'موظفين': members.length } }
      return {
        title: p.text, sub: `بمجموعة ${GROUP_LABEL[p.group as EyeGroup] ?? p.group}. ${slow ? `${slow} موظف يطوّلون أكثر من المتوقع بـ٤٠٪+.` : ''}`,
        cause: perf.find((g) => g.group === p.group)?.field === false
          ? 'من شغل الدور المسجّل بالنظام بآخر ٣٠ يوم، والحضور مقابل الجدول المسجّل.'
          : 'من مقارنة وقت كل حجز بالمتوقع لنفس الخدمة، والحضور مقابل الجدول.',
        impact: perf.find((g) => g.group === p.group)?.field === false ? 'طابور الدور يتراكم ويتأخر على باقي الأقسام.' : 'تأخّر الحجوزات وضغط على باقي الكادر.',
        suggestion: p.fixes[0] ?? '—', severity: 'MEDIUM' as const, link: p.link ?? '',
        evidence: perfEvidence(p.group, perf, groups),
      }
    }
    if (!biz) return null
    const ch = biz.mtd.lastRevenue > 0 ? Math.round(((biz.mtd.revenue - biz.mtd.lastRevenue) / biz.mtd.lastRevenue) * 100) : null
    return {
      title: ch == null ? `إيراد الشهر لحد اليوم ${fmtIQD(biz.mtd.revenue)}` : `الإيراد ${ch >= 0 ? 'أعلى' : 'أقل'} بـ${Math.abs(ch)}% من نفس الفترة الشهر الماضي`,
      sub: `${biz.mtd.bookings} حجز منجز هالشهر مقابل ${biz.mtd.lastBookings} — ${fmtIQD(biz.mtd.revenue)} مقابل ${fmtIQD(biz.mtd.lastRevenue)}.`,
      cause: ch != null && ch < 0 ? 'حجوزات منجزة أقل، أو فواتير ما انسجّلت بعد.' : 'وتيرة الإنجاز والفوترة.',
      impact: biz.forecast.insufficient ? biz.forecast.basis : `المتوقع لآخر الشهر ${fmtIQD(biz.forecast.expected)}.`,
      suggestion: ch != null && ch < 0 ? 'تابع الزبائن الي استفسروا وما حجزوا، وسجّل الفواتير المعلّقة.' : 'ركّز على الحجوزات المفتوحة حتى تتحول لإيراد.',
      severity: (ch != null && ch <= -15 ? 'HIGH' : 'LOW') as 'HIGH' | 'LOW', link: '/leader-invoices',
      evidence: { 'حجوزات الشهر': biz.mtd.bookings, 'الإيراد': fmtIQD(biz.mtd.revenue), 'نفس الفترة قبل': fmtIQD(biz.mtd.lastRevenue), 'أساس التوقع': biz.forecast.basis } as Record<string, unknown>,
    }
  }, [focus, late, perf, groups, biz])

  const approve = async () => {
    if (!proposal || !confirm(`توافق على اقتراح ماتركس؟\n\n${proposal.title}`)) return
    setBusy(true)
    try { await api.approveMatrixProposal(proposal.id); onDone() } catch (e) { alert(e instanceof Error ? e.message : 'تعذّر التنفيذ') } finally { setBusy(false) }
  }

  const sev = view?.severity === 'HIGH' ? ['أهمية عالية', 'border-red-400/40 text-[var(--mx-bad)]'] : view?.severity === 'MEDIUM' ? ['أهمية متوسطة', 'border-amber-400/40 text-[var(--mx-warn)]'] : ['أهمية منخفضة', 'border-emerald-400/40 text-[var(--mx-ok)]']
  return (
    <section className={`${card} p-4`}>
      <div className="mb-3 flex items-center justify-between">
        <h3 className="text-lg font-extrabold text-[var(--mx-accent)]">✨ ماتركس اكتشف</h3>
        {view && <span className={`rounded-full border px-2.5 py-0.5 text-[11px] ${sev[1]}`}>{sev[0]}</span>}
      </div>
      {!view ? <p className="text-sm text-[var(--mx-muted)]">ماتركس يحلّل…</p> : (
        <>
          <div className="rounded-xl border border-[var(--mx-border)] bg-[var(--mx-sunken)] p-3">
            <b className="block text-base text-[var(--mx-title)]">{view.title}</b>
            <p className="mt-1 text-sm text-[var(--mx-muted)]">{view.sub}</p>
          </div>
          <div className="mt-3 grid gap-2 md:grid-cols-3">
            <Box title="🔗 السبب المحتمل" text={view.cause} />
            <Box title="📊 التأثير المتوقع" text={view.impact} />
            <Box title="💡 اقتراح ماتركس" text={proposal ? proposal.title : view.suggestion} />
          </div>
          <div className="mt-3 grid gap-2 sm:grid-cols-3">
            {proposal ? (
              <button disabled={busy} onClick={approve} className="rounded-xl bg-gradient-to-l from-blue-500 to-blue-700 py-2.5 font-extrabold text-white shadow-[0_0_20px_rgba(59,130,246,0.5)] disabled:opacity-50">🚀 تنفيذ</button>
            ) : view.link ? (
              <Link to={view.link} className="rounded-xl bg-gradient-to-l from-blue-500 to-blue-700 py-2.5 text-center font-extrabold text-white">ودّيني ←</Link>
            ) : <span className="rounded-xl bg-[var(--mx-sunken)] py-2.5 text-center text-sm text-[var(--mx-muted)]">ماكو شي يحتاج قرار</span>}
            <button onClick={() => setModal('sim')} disabled={!proposal} className="rounded-xl border border-[var(--mx-border)] py-2.5 font-bold text-[var(--mx-accent)] disabled:opacity-40">📊 محاكاة</button>
            <button onClick={() => setModal('evidence')} className="rounded-xl border border-[var(--mx-border)] py-2.5 font-bold text-[var(--mx-accent)]">📄 عرض الأدلة</button>
          </div>
        </>
      )}
      {modal && view && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/60 p-4" onClick={() => setModal(null)}>
          <div className="w-full max-w-lg rounded-2xl bg-[var(--mx-card)] p-5 text-[var(--mx-text)]" dir="rtl" onClick={(e) => e.stopPropagation()}>
            {modal === 'evidence' ? (
              <>
                <h4 className="mb-3 font-extrabold">📄 الأدلة (أرقام من النظام)</h4>
                <dl className="space-y-1 text-sm">{Object.entries(view.evidence).map(([k, v]) => {
                  const [label, id] = k.split('|')
                  return (
                    <div key={k} className="flex justify-between gap-3 border-b border-[var(--mx-border)] py-1">
                      <dt className="text-[var(--mx-muted)]">{id ? <Link to={`/matrix/employee/${id}`} className="font-bold text-[var(--mx-accent)] hover:underline">{label} ←</Link> : label}</dt>
                      <dd className="text-left font-bold">{String(v)}</dd>
                    </div>
                  )
                })}</dl>
                {proposal?.evidence && (
                  <div className="mt-3">
                    <p className="mb-1 text-xs font-bold text-[var(--mx-muted)]">💡 أدلة اقتراح ماتركس: {proposal.employeeName ?? ''}</p>
                    <div className="grid grid-cols-2 gap-2">
                      {evidenceCards(proposal).map((c) => (
                        <div key={c.label} className="rounded-lg bg-[var(--mx-sunken)] p-2 text-center"><span>{c.icon}</span><p className="text-[10px] text-[var(--mx-muted)]">{c.label}</p><b className="text-xs">{c.value}</b></div>
                      ))}
                    </div>
                  </div>
                )}
                {view.link && <Link to={view.link} className="mt-3 inline-block text-sm font-bold text-[var(--mx-info)]">ودّيني للمشكلة ←</Link>}
              </>
            ) : proposal && (
              <>
                <h4 className="mb-2 font-extrabold">📊 محاكاة — شنو راح يصير لو وافقت</h4>
                <p className="mb-2 text-sm text-[var(--mx-text)]">{proposal.title}</p>
                <div className="rounded-lg bg-[var(--mx-sunken)] p-3 text-sm">
                  {proposal.kind === 'PREDICTION' && <>ترسل هالرسالة للموظف (بس هو يشوفها):<br /><b>{String(proposal.payload?.message ?? '')}</b></>}
                  {proposal.kind === 'GUIDE_RULE' && <>تنضاف تعليمة لماتركس على شاشة <b dir="ltr">{String(proposal.payload?.route ?? '')}</b>:<br /><b>{String(proposal.payload?.text ?? '')}</b></>}
                  {proposal.kind === 'RULE_TUNE' && <>{proposal.payload?.enabled ? 'تتفعّل' : 'تتعطّل'} تعليمة موجودة.</>}
                </div>
                <p className="mt-2 text-[11px] text-[var(--mx-muted)]">ماكو غرامات ولا نقاط — والمحاكاة ما غيّرت ولا شي.</p>
              </>
            )}
            <button onClick={() => setModal(null)} className="mt-4 rounded-lg border border-[var(--mx-border)] px-4 py-1.5 text-sm">سكّر</button>
          </div>
        </div>
      )}
    </section>
  )
}

function Box({ title, text }: { title: string; text: string }) {
  return (
    <div className="rounded-xl border border-[var(--mx-border)] bg-[var(--mx-sunken)] p-3">
      <b className="text-sm text-[var(--mx-info)]">{title}</b>
      <p className="mt-1 text-[13px] leading-6 text-[var(--mx-text)]">{text}</p>
    </div>
  )
}

// ── البث المباشر ──
const FEED_ICON: Record<MatrixFeedItem['kind'], [string, string, string]> = {
  BOOKING: ['📅', 'text-[var(--mx-info)]', '/bookings'],
  INVOICE: ['💵', 'text-[var(--mx-ok)]', '/leader-invoices'],
  ACTION: ['🤖', 'text-[var(--mx-violet)]', '/?board=matrix'],
  ESCALATED: ['⚠️', 'text-[var(--mx-bad)]', '/?board=matrix'],
  PROPOSAL: ['💡', 'text-[var(--mx-warn)]', '/matrix/decisions'],
}
function Feed({ items }: { items: MatrixFeedItem[] }) {
  return (
    <section className={`${card} p-4`}>
      <h3 className="text-lg font-extrabold text-[var(--mx-accent)]">📡 البث المباشر لماتركس</h3>
      <p className="mb-3 text-[11px] text-[var(--mx-muted)]">مراقبة لحظية لأهم الأحداث في النظام</p>
      {items.length === 0 ? <p className="text-sm text-[var(--mx-muted)]">ماكو أحداث بآخر ٣ أيام.</p> : (
        <ul className="max-h-[330px] space-y-2 overflow-y-auto">
          {items.map((it, i) => {
            const [icon, cls, to] = FEED_ICON[it.kind]
            return (
              <li key={i}>
                <Link to={to} className="flex items-center gap-3 rounded-xl border border-[var(--mx-border)] bg-[var(--mx-sunken)] p-2.5 hover:border-[var(--mx-accent)]">
                  <span className={`text-xl ${cls}`}>{icon}</span>
                  <span className="min-w-0 flex-1"><b className="block truncate text-[13px] text-[var(--mx-title)]">{it.title}</b><span className="block truncate text-[11px] text-[var(--mx-muted)]">{it.sub} · {since(it.at)}</span></span>
                  <span className="text-[var(--mx-muted)]">‹</span>
                </Link>
              </li>
            )
          })}
        </ul>
      )}
    </section>
  )
}

// ── رسم ٣٠ يوم: المنجز (أعمدة)، الإيراد والالتزام بالموعد (خطوط) ──
function TrendChart({ days }: { days: MatrixTrendDay[] }) {
  const W = 340, H = 170, P = 22
  const maxC = Math.max(1, ...days.map((d) => d.completed))
  const maxR = Math.max(1, ...days.map((d) => d.revenue))
  const x = (i: number) => P + (i * (W - P * 2)) / Math.max(1, days.length - 1)
  const line = (vals: number[], max: number) => vals.map((v, i) => `${i ? 'L' : 'M'}${x(i).toFixed(1)} ${(H - P - (v / max) * (H - P * 2)).toFixed(1)}`).join(' ')
  const onTime = days.map((d) => (d.scheduled ? (d.onTime / d.scheduled) * 100 : 0))
  return (
    <section className={`${card} p-4`}>
      <h3 className="text-sm font-extrabold text-[var(--mx-accent)]">📈 أداء الشركة (آخر ٣٠ يوم)</h3>
      <div className="mt-1 flex gap-3 text-[10px] text-[var(--mx-muted)]">
        <span><i className="inline-block h-2 w-2 rounded-full bg-sky-400" /> الحجوزات المنجزة</span>
        <span><i className="inline-block h-2 w-2 rounded-full bg-emerald-400" /> الأرباح</span>
        <span><i className="inline-block h-2 w-2 rounded-full bg-violet-400" /> الالتزام بالموعد</span>
      </div>
      {days.length === 0 ? <p className="mt-6 text-sm text-[var(--mx-muted)]">ماكو بيانات.</p> : (
        <svg viewBox={`0 0 ${W} ${H}`} className="mt-2 w-full">
          {days.map((d, i) => {
            const h = (d.completed / maxC) * (H - P * 2)
            return <rect key={d.day} x={x(i) - 3} y={H - P - h} width="6" height={h} rx="1.5" fill="url(#cbar)" />
          })}
          <defs><linearGradient id="cbar" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#38bdf8" /><stop offset="1" stopColor="#1d4ed8" /></linearGradient></defs>
          <path d={line(days.map((d) => d.revenue), maxR)} fill="none" stroke="#34d399" strokeWidth="1.6" />
          <path d={line(onTime, 100)} fill="none" stroke="#a78bfa" strokeWidth="1.4" strokeDasharray="3 2" />
          {days.map((d, i) => i % 5 === 0 && <text key={d.day} x={x(i)} y={H - 6} fontSize="8" fill="var(--mx-axis)" textAnchor="middle">{d.day}</text>)}
        </svg>
      )}
    </section>
  )
}

// ── قائمة مراقبة الموظفين (تصميم (ع)) ──
function WatchList({ perf }: { perf: GroupPerformance[] }) {
  // كل موظف ينقاس بشغل دوره، والترتيب داخل مجموعته: أضعف واحد من كل مجموعة أول،
  // حتى ما ينقارن محاسب ويا فني بنفس المسطرة.
  const all = perf.flatMap((g) => g.members.map((m) => ({ ...m, group: g.group, idx: perfIndex(m, g.field !== false) })))
  const rank = new Map<string, number>()
  for (const g of perf) {
    [...g.members].map((m) => ({ id: m.id, i: perfIndex(m, g.field !== false) })).sort((a, b) => a.i - b.i).forEach((x, n) => rank.set(x.id, n))
  }
  const rows = [...all].sort((a, b) => (rank.get(a.id) ?? 0) - (rank.get(b.id) ?? 0) || a.idx - b.idx).slice(0, 6)
  const good = all.filter((m) => m.idx >= 75).length
  const support = all.filter((m) => m.idx < 60).length
  return (
    <section className={`${card} flex flex-col p-4`}>
      <div className="mb-3 flex items-start justify-between gap-2">
        <div className="flex items-center gap-2">
          <span className="grid h-10 w-10 place-items-center rounded-xl bg-sky-500/15 text-lg">👥</span>
          <div>
            <h3 className="text-base font-extrabold text-[var(--mx-title)]">قائمة مراقبة الموظفين</h3>
            <p className="text-[11px] text-[var(--mx-muted)]">متابعة أداء الموظفين ومستويات الإنتاجية</p>
          </div>
        </div>
        <span className="rounded-xl bg-sky-500/10 px-3 py-1 text-center"><b className="block text-lg text-[var(--mx-accent)]">{all.length}</b><span className="text-[10px] text-[var(--mx-muted)]">موظف</span></span>
      </div>
      {rows.length === 0 ? <p className="text-sm text-[var(--mx-muted)]">ماتركس يحسب…</p> : (
        <table className="w-full text-[12px]">
          <thead><tr className="bg-[var(--mx-sunken)] text-[var(--mx-muted)]"><th className="rounded-r-lg p-2 text-right">الموظف</th><th className="text-right">القسم</th><th>مؤشر الأداء</th><th className="rounded-l-lg">الحالة</th></tr></thead>
          <tbody>
            {rows.map((m) => {
              const st = perfStatus(m.idx)
              return (
                <tr key={m.id} className="border-b border-[var(--mx-border)]">
                  <td className="py-2">
                    <Link to={`/matrix/employee/${m.id}`} className="flex items-center gap-2 font-bold text-[var(--mx-title)] hover:text-[var(--mx-accent)]">
                      <span className="grid h-7 w-7 shrink-0 place-items-center rounded-full bg-sky-500/15 text-[11px] text-[var(--mx-accent)]">{m.name.trim().split(/\s+/).slice(0, 2).map((w) => w[0]).join(' ')}</span>
                      <span className="truncate">{m.name}</span>
                    </Link>
                  </td>
                  <td><span className="rounded-lg bg-[var(--mx-sunken)] px-2 py-0.5 text-[11px] text-[var(--mx-text)]">{GROUP_ICON[m.group] ?? ''} {GROUP_LABEL[m.group as EyeGroup] ?? m.group}</span></td>
                  <td className="px-2">
                    <b className="text-[var(--mx-title)]">{m.idx}%</b>
                    <div className="mt-0.5 h-1.5 w-20 rounded-full bg-[var(--mx-sunken)]"><div className="h-1.5 rounded-full" style={{ width: `${m.idx}%`, background: st.bar }} /></div>
                  </td>
                  <td className="text-center"><span className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-[10px] font-bold ${st.c}`}><i className={`h-1.5 w-1.5 rounded-full ${st.dot}`} />{st.t}</span></td>
                </tr>
              )
            })}
          </tbody>
        </table>
      )}
      {all.length > 0 && (
        <div className="mt-3 flex items-center gap-2 rounded-xl bg-[var(--mx-sunken)] p-2.5">
          <span className="text-lg">📊</span>
          <div className="text-[11px]">
            <b className="block text-[var(--mx-accent)]">{good} من أصل {all.length} موظف أداؤهم جيد أو ممتاز</b>
            <span className="text-[var(--mx-muted)]">{support > 0 ? `يوجد ${support} يحتاج دعم إضافي لتحسين أدائه.` : 'ماكو أحد يحتاج دعم هسه.'} · المؤشر للمتابعة بس — مو نقاط.</span>
          </div>
        </div>
      )}
    </section>
  )
}

// ── توقعات ماتركس (أعمال بس — بلا توقعات شخصية) ──
function Predictions({ biz, late, props }: { biz: MatrixBusinessData | null; late: LateFocus | null; props: MatrixProposal[] }) {
  const items: { text: string; change: number | null; conf: number }[] = []
  if (biz) {
    const ch = biz.mtd.lastBookings > 0 ? Math.round(((biz.mtd.bookings - biz.mtd.lastBookings) / biz.mtd.lastBookings) * 100) : null
    items.push({ text: `المتوقع لآخر الشهر ${biz.forecast.expectedJobs} حجز`, change: ch, conf: biz.forecast.insufficient ? 50 : confidence(biz.mtd.bookings) })
    const rc = biz.mtd.lastRevenue > 0 ? Math.round(((biz.mtd.revenue - biz.mtd.lastRevenue) / biz.mtd.lastRevenue) * 100) : null
    items.push({ text: `الأرباح المتوقعة ${fmtIQD(biz.forecast.expected)}`, change: rc, conf: biz.forecast.insufficient ? 50 : confidence(biz.mtd.bookings) })
  }
  if (late) items.push({ text: `الحجوزات المتأخرة: ${late.recentPct}% من آخر ٣ أيام`, change: late.changePct == null ? null : -late.changePct, conf: confidence(late.recent.total) })
  const pred = props.filter((p) => p.kind === 'PREDICTION').length
  if (pred) items.push({ text: `${pred} توقّع ينتظر قرارك (موظفين ما يلحگون شغل اليوم)`, change: null, conf: 72 })
  return (
    <section className={`${card} p-4`}>
      <h3 className="text-sm font-extrabold text-[var(--mx-accent)]">📉 توقعات ماتركس</h3>
      <p className="mb-2 text-[11px] text-[var(--mx-muted)]">توقعات مبنية على تحليل البيانات الحالية</p>
      <ul className="space-y-2">
        {items.map((it, i) => (
          <li key={i} className="flex items-center gap-2 rounded-xl border border-[var(--mx-border)] bg-[var(--mx-sunken)] p-2">
            <span className={`min-w-[60px] rounded-lg px-2 py-1 text-center text-xs font-bold ${it.change == null ? 'bg-slate-600/30 text-[var(--mx-text)]' : it.change >= 0 ? 'bg-emerald-500/15 text-[var(--mx-ok)]' : 'bg-red-500/15 text-[var(--mx-bad)]'}`}>
              {it.change == null ? '—' : `${it.change >= 0 ? '↗ +' : '↘ '}${it.change}%`}
            </span>
            <span className="flex-1 text-[12px] text-[var(--mx-text)]">{it.text}</span>
            <span className="rounded-full border border-[var(--mx-border)] px-1.5 py-0.5 text-[10px] text-[var(--mx-accent)]" title="الثقة حسب كمية البيانات">{it.conf}%{it.conf <= 50 && ' · بيانات قليلة'}</span>
          </li>
        ))}
      </ul>
    </section>
  )
}

// ── اسأل ماتركس (للمدير — قراءة بس) ──
function AskMatrix() {
  const [q, setQ] = useState('')
  const [a, setA] = useState<{ answer: string; source: string } | null>(null)
  const [busy, setBusy] = useState(false)
  const ask = async (text: string) => {
    if (!text.trim() || busy) return
    setBusy(true); setA(null)
    try { setA(await api.askMatrix(text)) } catch (e) { setA({ answer: e instanceof Error ? e.message : 'تعذّر الجواب', source: 'ERR' }) } finally { setBusy(false) }
  }
  return (
    <section className={`${card} flex flex-col p-4`}>
      <h3 className="text-sm font-extrabold text-[var(--mx-accent)]">💬 اسأل ماتركس</h3>
      <p className="mb-2 text-[11px] text-[var(--mx-muted)]">اطرح سؤالك واحصل على تحليل فوري</p>
      <form onSubmit={(e) => { e.preventDefault(); ask(q) }} className="flex gap-2">
        <input value={q} onChange={(e) => setQ(e.target.value)} placeholder="مثال: ما سبب تأخر الحجوزات؟" className="min-w-0 flex-1 rounded-xl border border-[var(--mx-border)] bg-[var(--mx-sunken)] px-3 py-2 text-sm text-[var(--mx-title)] placeholder:text-[var(--mx-muted)]" />
        <button disabled={busy} className="rounded-xl bg-blue-600 px-3 text-white disabled:opacity-50">➤</button>
      </form>
      <div className="mt-2 flex flex-wrap gap-1.5">
        {['تحليل الأرباح اليوم', 'أداء الموظفين', 'مقترحات التحسين', 'منو متأخر اليوم؟'].map((s) => (
          <button key={s} onClick={() => { setQ(s); ask(s) }} className="rounded-full border border-[var(--mx-border)] px-2.5 py-1 text-[11px] text-[var(--mx-text)] hover:border-[var(--mx-accent)]">{s}</button>
        ))}
      </div>
      {busy && <p className="mt-3 text-xs text-[var(--mx-muted)]">ماتركس يفكّر…</p>}
      {a && (
        <div className="mt-3 whitespace-pre-line rounded-xl border border-[var(--mx-border)] bg-[var(--mx-sunken)] p-3 text-[13px] leading-6 text-[var(--mx-text)]">
          {a.answer}
          {a.source !== 'ERR' && <p className="mt-1 text-[10px] text-[var(--mx-muted)]">{a.source === 'MODEL' ? 'تحليل بالنموذج (بلا أسماء)' : 'من أرقام النظام مباشرة'}</p>}
        </div>
      )}
    </section>
  )
}

// ═══ دليل «أداء الموظفين» حسب شغل كل واحد ═══
// كادر الميدان (ليدرية وفنيين): الحجوزات والسرعة. الباقين (مراقب، محاسب،
// تنسيق، جودة...): شغلهم اليوم من «شغلك اليوم» — مو «حجوزات 0» الي ما تعني شي.
// المفتاح «الاسم|المعرّف» حتى النافذة تسوي الاسم رابط لتقريره.
const FIELD_GROUPS = new Set(['TECHS', 'LEADERS'])
function perfEvidence(group: string, perf: GroupPerformance[], groups: GroupRep[]): Record<string, unknown> {
  const out: Record<string, unknown> = {}
  if (FIELD_GROUPS.has(group)) {
    for (const m of (perf.find((g) => g.group === group)?.members ?? []).slice(0, 8)) {
      const speed = m.speed == null ? '' : m.speed > 10 ? '، السرعة: بيانات تحتاج مراجعة' : `، السرعة ×${m.speed}`
      out[`${m.name}|${m.id}`] = `حجوزات ${m.jobs}: منجز ${m.completed}، جزئي ${m.partial}، مفتوح ${m.open}${speed}${m.absent ? '، ما حضر اليوم' : m.late ? `، تأخّر ${m.late}د` : ''}`
    }
    return out
  }
  for (const e of (groups.find((g) => g.group === group)?.employees ?? []).slice(0, 8)) {
    const work = e.workload.length
      ? e.workload.map((w) => `${w.label}: ${w.done ? `${w.verb} ${w.done}، ` : ''}${w.left ? `باقي ${w.left}` : 'خالص ✓'}`).join(' · ')
      : 'ماكو شغل مسجّل عليه اليوم'
    out[`${e.name ?? '—'}|${e.id ?? ''}`] = work + (e.open ? ` · ${e.open} تذكير مفتوح` : '')
  }
  return out
}
