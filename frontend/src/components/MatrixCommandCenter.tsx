import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Link } from 'react-router-dom'
import {
  api, type GroupPerformance, type LateFocus, type MatrixBusiness as MatrixBusinessData, type MatrixFeedItem,
  type MatrixProposal, type MatrixWatch,
} from '../api'
import MatrixBusiness from './MatrixBusiness'
import OwnerSwitch from './OwnerSwitch'
import { evidenceCards } from './matrixEvidence'
import { SWITCH_MATRIX_STAFF_EYE } from '../systemSwitches'
import MatrixEyeGraphic from './MatrixEyeGraphic'
import { GroupPerf } from './MatrixRoleEyes'
import { GROUP_LABEL, eyeColor, type EyeGroup, type EyeMood } from './matrixEyeColors'
import MatrixRoleChains from './MatrixRoleChains'
import MatrixWatchEyes from './MatrixWatchEyes'
import { MatrixLearning, MatrixProjectsEye } from './MatrixProjectsEye'
import MatrixDialect from './MatrixDialect'
import MatrixSuggestAccuracyPanel from './MatrixSuggestAccuracy'
import MatrixAutonomy from './MatrixAutonomy'
import { pollVisible } from '../utils/poll'

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
  // الحجز المفتوح الي ما بدا بي أحد (ما انثبت، تأجّل، انتقل) ما ينحسب عليه:
  // چان موظف عليه حجوزات معلّقة بلا ذنبه يطلع ٠٪ «يحتاج دعم».
  const measured = m.jobs - (m.unstarted ?? 0)
  let v = measured > 0 ? ((m.completed + m.partial * 0.5) / measured) * 100 : 90
  if (m.absent) v -= 15
  else if (m.late > 0) v -= Math.min(20, m.late / 3)
  if (m.speed != null && m.speed > 1.15) v -= Math.min(25, (m.speed - 1) * 40)
  return Math.max(0, Math.min(100, Math.round(v)))
}
function confidence(samples: number) { return samples >= 20 ? 87 : samples >= 8 ? 72 : 50 }


export default function MatrixCommandCenter() {
  const [groups, setGroups] = useState<GroupRep[] | null>(null)
  const [feed, setFeed] = useState<MatrixFeedItem[]>([])
  // (ع) 10-13: كل رقم فوگ ينضغط ويطلع مصدره موظف موظف
  const [proof, setProof] = useState<ProofKind | null>(null)
  const [late, setLate] = useState<LateFocus | null>(null)
  const [biz, setBiz] = useState<MatrixBusinessData | null>(null)
  const [props, setProps] = useState<MatrixProposal[]>([])
  const [perf, setPerf] = useState<GroupPerformance[]>([])
  const [focus, setFocus] = useState<Focus>('late')
  const [sel, setSel] = useState<string | null>(null)
  const [vehicles, setVehicles] = useState<{ n: number; high: number } | null>(null)
  const [err, setErr] = useState('')

  const loadAll = useCallback(() => {
    api.getRoleWatch().then((g) => {
      setGroups(g)
      Promise.all(g.filter((x) => x.employees.length > 0).map((x) => api.getGroupPerformance(x.group).catch(() => null)))
        .then((r) => setPerf(r.filter((x): x is GroupPerformance => !!x)))
    }).catch((e) => setErr(e instanceof Error ? e.message : 'تعذر جلب ماتركس'))
    api.getMatrixFeed().then(setFeed).catch(() => {})
    api.getLateFocus().then(setLate).catch(() => {})
    api.getMatrixBusiness().then(setBiz).catch(() => {})
    api.getMatrixProposals('PENDING').then(setProps).catch(() => {})
    api.getMatrixEyes().then((r) => { const v = r.eyes.find((e) => e.key === 'VEHICLES'); setVehicles({ n: v?.items.length ?? 0, high: v?.high ?? 0 }) }).catch(() => {})
  }, [])
  useEffect(() => {
    loadAll()
    return pollVisible(() => { api.getMatrixFeed().then(setFeed).catch(() => {}) }, 60_000)
  }, [loadAll])

  if (err) return <p className="rounded-lg bg-red-50 p-3 text-red-600">{err}</p>
  if (!groups) return <p className="text-[var(--mx-muted)]">ماتركس يجمع البيانات…</p>

  const active = groups.filter((g) => g.employees.length > 0)
  const people = active.reduce((n, g) => n + g.employees.length, 0)
  const red = active.reduce((n, g) => n + g.red, 0)
  const attention = active.reduce((n, g) => n + g.red + g.alert, 0)
  const tasks = active.reduce((n, g) => n + g.employees.reduce((m, e) => m + e.workload.reduce((k, w) => k + w.done + w.left, 0), 0), 0)
  const stability = people ? Math.round(((people - red) / people) * 100) : 100

  return (
    <div dir="rtl" className="space-y-4 rounded-3xl bg-[var(--mx-bg)] p-3 text-[var(--mx-text)] sm:p-5">
      {proof && <ProofModal kind={proof} groups={active} onClose={() => setProof(null)} />}
      {/* ── الرأس ── */}
      <div className="grid items-center gap-4 lg:grid-cols-[1fr_auto]">
        <div className="flex flex-wrap items-center gap-4">
          <BigSharingan />
          <div>
            <span className="mb-2 inline-flex items-center gap-1.5 rounded-full border border-emerald-400/30 bg-emerald-500/10 px-3 py-0.5 text-[11px] text-[var(--mx-ok)]">
              <span className="h-2 w-2 animate-pulse rounded-full bg-emerald-400" /> الذكاء التحليلي النشط الآن
            </span>
            <h2 className="text-3xl font-black text-[var(--mx-title)] sm:text-4xl">مركز قيادة <span className="bg-gradient-to-l from-sky-500 to-blue-600 bg-clip-text text-transparent">ماتركس</span></h2>
            <p className="mt-1 text-sm text-[var(--mx-muted)]">يراقب الأداء والموظفين والحجوزات والأرباح لحظياً</p>
          </div>
        </div>
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
          <Kpi value={`${stability}%`} label="استقرار النظام" tone="text-[var(--mx-ok)]" icon="〰️" onClick={() => setProof('stability')} />
          <Kpi value={attention} label="حالات تحتاج انتباه" tone="text-[var(--mx-bad)]" icon="⚠️" onClick={() => setProof('attention')} />
          <Kpi value={people} label="موظف تحت المراقبة" tone="text-[var(--mx-accent)]" icon="👥" onClick={() => setProof('people')} />
          <Kpi value={tasks} label="شغلة اليوم (منجز + باقي)" tone="text-[var(--mx-accent)]" icon="📦" onClick={() => setProof('tasks')} />
        </div>
      </div>

      {/* ── توقعات ماتركس — فوگ، من الحجوزات (قرار ع 10-13) ── */}
      <Predictions biz={biz} late={late} props={props} />

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
        {/* الاقتراح لازم يخص التركيز: چان أول اقتراح معلّق (جلال) يطلع بكل مكان. */}
        <Discovery focus={focus} late={late} perf={perf} groups={active} biz={biz}
          proposal={focus === 'perf' ? (props.find((p) => p.employeeId) ?? null) : null} onDone={loadAll} />
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
          {/* عين السيارات وعين ماتركس نفسه — بنفس الشبكة (طلب (ع) 10-05). */}
          {([['VEHICLES', vehicles ? `${vehicles.n} بند` : '…', vehicles?.high ? 'ANGRY' : vehicles?.n ? 'ALERT' : 'CALM'], ['MATRIX', 'شغله ودقته وكلفته', 'CALM']] as const).map(([k, sub, mood]) => {
            const color = eyeColor(k, mood)
            return (
              <button key={k} type="button" onClick={() => setSel(sel === k ? null : k)}
                className="flex flex-col items-center gap-1 rounded-xl border bg-[var(--mx-sunken)] p-2 text-center transition"
                style={{ borderColor: `color-mix(in srgb, ${color} ${sel === k ? 90 : 35}%, transparent)` }}>
                <MatrixEyeGraphic group={k} mood={mood} width={70} />
                <span className="text-[12px] font-bold" style={{ color: `color-mix(in srgb, ${color} 72%, var(--mx-title))` }}>عن {GROUP_LABEL[k]}</span>
                <span className="text-[11px] text-[var(--mx-muted)]">{sub}</span>
                <span className={`rounded-full px-2 py-0.5 text-[10px] font-bold ${mood === 'ANGRY' ? 'bg-red-500/15 text-[var(--mx-bad)]' : mood === 'ALERT' ? 'bg-amber-500/15 text-[var(--mx-warn)]' : 'bg-emerald-500/15 text-[var(--mx-ok)]'}`}>
                  ● {mood === 'ANGRY' ? 'تحتاج انتباه' : mood === 'ALERT' ? 'منتبهة' : 'مستقر'}
                </span>
              </button>
            )
          })}
        </div>
      </div>
      {sel === 'VEHICLES' && <MatrixWatchEyes key="v" initialTab="VEHICLES" />}
      {sel === 'MATRIX' && <div className="space-y-4"><MatrixAutonomy /><MatrixSuggestAccuracyPanel /></div>}
      {sel === 'SUPERVISORS' && <MatrixProjectsEye />}
      {sel && sel !== 'VEHICLES' && sel !== 'MATRIX' && sel !== 'SUPERVISORS' && <div><GroupPerf key={sel} group={sel} /></div>}

      {/* ── ماتركس ٢٠٥٠: سلسلة الحجز — كل دور ومحطاته ── */}
      <MatrixRoleChains />
      <MatrixWatchEyes />
      <MatrixSuggestAccuracyPanel />
      <MatrixAutonomy />
      <MatrixProjectsEye />
      <MatrixLearning />
      <MatrixDialect />

      {/* (ع) 10-13: صف «أداء الشركة · قائمة المراقبة · اسأل ماتركس» انشال، والتوقعات صعدت فوگ */}

      <details className="rounded-2xl border border-[var(--mx-border)] bg-[var(--mx-card)] text-[var(--mx-text)]">
        <summary className="cursor-pointer p-3 text-sm font-bold">💼 تفاصيل الأعمال والزبائن</summary>
        <div className="p-3 pt-0"><MatrixBusiness /></div>
      </details>
    </div>
  )
}

function Kpi({ value, label, tone, icon, onClick }: { value: string | number; label: string; tone: string; icon: string; onClick?: () => void }) {
  return (
    <button type="button" onClick={onClick} title="اضغط حتى تشوف منين جا الرقم"
      className={`${card} min-w-[120px] p-3 text-right transition hover:-translate-y-0.5 hover:border-[var(--mx-accent)]`}>
      <div className="flex items-start justify-between"><b className={`text-2xl ${tone}`}>{value}</b><span>{icon}</span></div>
      <p className="mt-1 text-[11px] text-[var(--mx-muted)]">{label} ‹</p>
    </button>
  )
}

// ═══ إثبات الأرقام — (ع) 10-13: «من اضغط عليهن يوديني لشي يثبت صحتهن» ═══
type ProofKind = 'stability' | 'attention' | 'people' | 'tasks'
type WatchGroup = { group: string; employees: MatrixWatch[]; red: number; alert: number }

function ProofModal({ kind, groups, onClose }: { kind: ProofKind; groups: WatchGroup[]; onClose: () => void }) {
  const all = groups.flatMap((g) => g.employees.map((e) => ({ ...e, group: g.group })))
  const people = all.length
  const red = all.filter((e) => e.level === 'RED')
  const alert = all.filter((e) => e.level === 'ALERT')
  const name = (e: MatrixWatch) => e.name || '—'
  const glabel = (g: string) => GROUP_LABEL[g as EyeGroup] ?? g
  const emp = (e: MatrixWatch, extra?: React.ReactNode) => (
    <li key={e.id} className="rounded-lg bg-[var(--mx-sunken)] px-3 py-2">
      <div className="flex items-center justify-between gap-2">
        {e.id ? <Link to={`/matrix/employee/${e.id}`} className="font-bold text-[var(--mx-title)] hover:underline">{name(e)} ←</Link> : <b>{name(e)}</b>}
        <span className="text-[11px] text-[var(--mx-muted)]">{glabel(e.group)} · {e.level === 'RED' ? '🔴 أحمر' : e.level === 'ALERT' ? '🟡 منتبه' : '🟢 مستقر'}</span>
      </div>
      {extra}
    </li>
  )
  const reasons = (e: MatrixWatch) => e.items.length > 0 && (
    <ul className="mt-1 space-y-0.5 text-[12px] text-[var(--mx-muted)]">
      {e.items.map((it, i) => <li key={i}>• {it.escalated ? '⬆️ ' : ''}{it.summary} <span className="text-[10px]">({it.label} · من {new Date(it.since).toLocaleDateString('ar-IQ')})</span></li>)}
    </ul>
  )
  const work = (e: MatrixWatch) => (
    <ul className="mt-1 space-y-0.5 text-[12px]">
      {e.workload.map((w) => (
        <li key={w.key} className="flex items-center justify-between gap-2">
          <Link to={w.route} className="text-[var(--mx-accent)] hover:underline">{w.label} ←</Link>
          <span className="tabular-nums text-[var(--mx-muted)]">{w.verb} {w.done} · باقي {w.left} = <b className="text-[var(--mx-title)]">{w.done + w.left}</b></span>
        </li>
      ))}
    </ul>
  )
  let title: string, how: string, body: React.ReactNode
  if (kind === 'stability') {
    title = `〰️ استقرار النظام ${people ? Math.round(((people - red.length) / people) * 100) : 100}%`
    how = `الحسبة: (كل الموظفين تحت المراقبة ${people} − الحمر ${red.length}) ÷ ${people}. الأحمر = تذكير صعد للمدير أو ٣ تذكيرات مفتوحة.`
    body = red.length ? <ul className="space-y-1.5">{red.map((e) => emp(e, reasons(e)))}</ul> : <p>ماكو موظف أحمر ✅</p>
  } else if (kind === 'attention') {
    title = `⚠️ ${red.length + alert.length} حالة تحتاج انتباه`
    how = `= الحمر (${red.length}) + المنتبه عليهم (${alert.length}). تحت كل اسم التذكيرات المفتوحة الي خلّته هيچ.`
    body = <ul className="space-y-1.5">{[...red, ...alert].map((e) => emp(e, reasons(e)))}</ul>
  } else if (kind === 'people') {
    title = `👥 ${people} موظف تحت المراقبة`
    how = 'كل موظف فعّال عنده دور يراقبه ماتركس (المالك والمدير ما ينحسبون). مجمّعين حسب المجموعة.'
    body = (
      <div className="space-y-3">
        {groups.filter((g) => g.employees.length).map((g) => (
          <div key={g.group}>
            <p className="mb-1 text-sm font-extrabold text-[var(--mx-title)]">{glabel(g.group)} ({g.employees.length})</p>
            <ul className="grid gap-1 sm:grid-cols-2">{g.employees.map((e) => emp({ ...e, group: g.group }))}</ul>
          </div>
        ))}
      </div>
    )
  } else {
    const withWork = all.filter((e) => e.workload.length > 0)
    const total = all.reduce((n, e) => n + e.workload.reduce((k, w) => k + w.done + w.left, 0), 0)
    title = `📦 ${total} شغلة اليوم`
    how = 'مو «مهام» بالمعنى العادي: هو مجموع شغل اليوم لكل موظف حسب دوره (مثلاً حجوزات يدققها المراقب، فواتير يعتمدها المحاسب، ورق الليدر) = الي خلّصه + الي باقي عليه. كل سطر يودّيك لشاشته.'
    body = withWork.length ? <ul className="space-y-1.5">{withWork.map((e) => emp(e, work(e)))}</ul> : <p>ماكو شغل محسوب اليوم.</p>
  }
  return (
    <div className="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/50 p-4" onClick={onClose}>
      <div dir="rtl" className="w-full max-w-3xl rounded-2xl bg-[var(--mx-bg)] p-5 text-[var(--mx-text)] shadow-2xl" onClick={(e) => e.stopPropagation()}>
        <div className="mb-2 flex items-center justify-between">
          <h3 className="text-lg font-extrabold text-[var(--mx-title)]">{title}</h3>
          <button onClick={onClose} className="text-2xl text-[var(--mx-muted)]">×</button>
        </div>
        <p className="mb-3 rounded-lg bg-[var(--mx-sunken)] px-3 py-2 text-[12px]">🧮 {how}</p>
        <div className="max-h-[70vh] overflow-y-auto">{body}</div>
      </div>
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
  const [modal, setModal] = useState<'sim' | 'evidence' | 'details' | null>(null)
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
    // (ع) 10-13: المقارنة حجوزات بحجوزات — دفعات المشاريع تنسجّل بالجملة وتخرّب النسبة
    const bRev = biz.mtd.revenue - (biz.mtd.projectRevenue ?? 0)
    const bLast = biz.mtd.lastRevenue - (biz.mtd.lastProjectRevenue ?? 0)
    const ch = bLast > 0 ? Math.round(((bRev - bLast) / bLast) * 100) : null
    return {
      title: ch == null ? `إيراد الحجوزات لحد اليوم ${fmtIQD(bRev)}` : `إيراد الحجوزات ${ch >= 0 ? 'أعلى' : 'أقل'} بـ${Math.abs(ch)}% من نفس الفترة الشهر الماضي`,
      sub: `${biz.mtd.bookings} حجز منجز هالشهر مقابل ${biz.mtd.lastBookings} — ${fmtIQD(bRev)} مقابل ${fmtIQD(bLast)}. ودفعات مشاريع انسجّلت هالشهر ${fmtIQD(biz.mtd.projectRevenue ?? 0)}.`,
      cause: ch != null && ch < 0 ? 'حجوزات منجزة أقل، أو فواتير ما انسجّلت بعد.' : 'وتيرة الإنجاز والفوترة.',
      impact: biz.forecast.insufficient ? biz.forecast.basis : `المتوقع لآخر الشهر ${fmtIQD(biz.forecast.expected)}.`,
      suggestion: ch != null && ch < 0 ? 'تابع الزبائن الي استفسروا وما حجزوا، وسجّل الفواتير المعلّقة.' : 'ركّز على الحجوزات المفتوحة حتى تتحول لإيراد.',
      severity: (ch != null && ch <= -15 ? 'HIGH' : 'LOW') as 'HIGH' | 'LOW', link: '/leader-invoices',
      evidence: { 'حجوزات الشهر': biz.mtd.bookings, 'إيراد الحجوزات': fmtIQD(bRev), 'نفس الفترة قبل': fmtIQD(bLast), 'دفعات مشاريع هالشهر': fmtIQD(biz.mtd.projectRevenue ?? 0), 'أساس التوقع': biz.forecast.basis } as Record<string, unknown>,
    }
  }, [focus, late, perf, groups, biz])

  // (ع): «من يگلي موظفين مقصرين يطلعلي منو، ووين مقصرين، واذا اكو سبب».
  const weak = useMemo(() => perf.flatMap((g) => g.members.map((m) => {
    const field = g.field !== false
    const why: string[] = []
    if (m.absent) why.push('ما سجّل حضور اليوم (عنده جدول، ومو بإجازة)')
    else if (m.late > 0) why.push(`تأخّر اليوم ${m.late} دقيقة عن جدوله`)
    if (field) {
      const stuck = m.open - (m.unstarted ?? 0)
      if (stuck > 0) why.push(`${stuck} حجز بدا بي وما خلّصه`)
      if (m.partial > 0) why.push(`${m.partial} حجز جزئي`)
      if (m.speed != null && m.speed >= 1.4) why.push(`يطوّل ×${m.speed.toFixed(1)} عن المتوقع لنفس الخدمة`)
    } else {
      const done = (m.metrics ?? []).filter((x) => x.label !== 'عمليات بالنظام').reduce((a, x) => a + x.value, 0)
      if ((m.metrics ?? []).length > 1 && done === 0) why.push('النظام ما لگى شغل مسجّل إله بآخر ٣٠ يوم (شغله برّا النظام؟)')
    }
    return { ...m, group: g.group, idx: perfIndex(m, field), why }
  })).filter((m) => m.idx < 75 && m.why.length > 0).sort((a, b) => a.idx - b.idx), [perf])

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
          <div className={`mt-3 grid gap-2 ${proposal ? 'sm:grid-cols-4' : 'sm:grid-cols-3'}`}>
            {proposal ? (
              <button disabled={busy} onClick={approve} className="rounded-xl bg-gradient-to-l from-blue-500 to-blue-700 py-2.5 font-extrabold text-white shadow-[0_0_20px_rgba(59,130,246,0.5)] disabled:opacity-50">🚀 تنفيذ</button>
            ) : view.link ? (
              <Link to={view.link} className="rounded-xl bg-gradient-to-l from-blue-500 to-blue-700 py-2.5 text-center font-extrabold text-white">ودّيني ←</Link>
            ) : <span className="rounded-xl bg-[var(--mx-sunken)] py-2.5 text-center text-sm text-[var(--mx-muted)]">ماكو شي يحتاج قرار</span>}
            <button onClick={() => setModal('details')} className="rounded-xl border border-[var(--mx-border)] py-2.5 font-bold text-[var(--mx-accent)]">
              {focus === 'late' ? `📋 الحجوزات (${late?.items?.length ?? 0})` : focus === 'perf' ? `👥 منو؟ (${weak.length})` : `📋 وين الإيراد (${biz?.uninvoiced?.length ?? 0})`}
            </button>
            {proposal && <button onClick={() => setModal('sim')} className="rounded-xl border border-[var(--mx-border)] py-2.5 font-bold text-[var(--mx-accent)]">📊 محاكاة</button>}
            <button onClick={() => setModal('evidence')} className="rounded-xl border border-[var(--mx-border)] py-2.5 font-bold text-[var(--mx-accent)]">📄 عرض الأدلة</button>
          </div>
        </>
      )}
      {modal && view && (
        <div className="fixed inset-0 z-50 grid place-items-center bg-black/60 p-4" onClick={() => setModal(null)}>
          <div className="w-full max-w-lg rounded-2xl bg-[var(--mx-card)] p-5 text-[var(--mx-text)]" dir="rtl" onClick={(e) => e.stopPropagation()}>
            {modal === 'details' ? (
              <div className="max-h-[70vh] overflow-y-auto">
                {focus === 'late' ? (
                  <>
                    <h4 className="mb-3 font-extrabold">📋 الحجوزات المتأخرة (آخر ٣ أيام)</h4>
                    {(late?.items ?? []).length === 0 ? <p className="text-sm text-[var(--mx-muted)]">ماكو ✅</p> : (
                      <ul className="space-y-2 text-sm">
                        {(late?.items ?? []).map((b) => (
                          <li key={b.id} className="rounded-lg bg-[var(--mx-sunken)] p-2">
                            <div className="flex flex-wrap items-center justify-between gap-2">
                              <Link to={`/bookings?focus=${b.id}`} className="font-extrabold text-[var(--mx-accent)] hover:underline">📋 {b.code} ←</Link>
                              <span className="text-xs font-bold text-[var(--mx-bad)]">{b.status === 'COMPLETED' ? `خلص متأخر ${b.daysLate} يوم` : `مفتوح — فات موعده بـ${b.daysLate} يوم`}</span>
                            </div>
                            <p className="mt-1 text-xs text-[var(--mx-muted)]">
                              {b.customer ?? '—'} · {b.service ?? '—'} · 📅 {new Date(b.scheduledAt).toLocaleDateString('ar-IQ', { timeZone: 'Asia/Baghdad' })}
                              {' · '}{b.unstaffed ? <b className="text-[var(--mx-warn)]">بلا كادر</b> : <>👷 {b.leader ?? 'بلا ليدر'}</>}
                              {b.partial && <b className="text-[var(--mx-warn)]"> · جزئي</b>}
                            </p>
                            <p className="mt-0.5 text-xs">💬 {b.reason ?? <span className="text-[var(--mx-muted)]">ماكو سبب مكتوب</span>}</p>
                          </li>
                        ))}
                      </ul>
                    )}
                  </>
                ) : focus === 'profit' ? (
                  <>
                    <h4 className="mb-2 font-extrabold">🧾 منجزة هالشهر وما عليها فاتورة</h4>
                    <p className="mb-2 text-[11px] text-[var(--mx-muted)]">إيراد ما انحسب بعد — كل وحدة تنفتح على الحجز. (القديم OLD والكشف مستثنين.)</p>
                    {(biz?.uninvoiced ?? []).length === 0 ? <p className="text-sm text-[var(--mx-muted)]">كلها مفوترة ✅</p> : (
                      <ul className="space-y-1.5 text-sm">
                        {(biz?.uninvoiced ?? []).map((b) => (
                          <li key={b.id} className="flex flex-wrap items-center justify-between gap-2 rounded-lg bg-[var(--mx-sunken)] p-2">
                            <Link to={`/bookings?focus=${b.id}`} className="font-extrabold text-[var(--mx-accent)] hover:underline">📋 {b.code} ←</Link>
                            <span className="text-xs text-[var(--mx-muted)]">{b.customer ?? '—'} · {b.service ?? '—'} · 👷 {b.leader ?? 'بلا ليدر'} · خلص {new Date(b.completedAt).toLocaleDateString('ar-IQ', { timeZone: 'Asia/Baghdad' })}</span>
                          </li>
                        ))}
                      </ul>
                    )}
                    {(biz?.inquirers ?? []).length > 0 && (
                      <>
                        <h4 className="mb-2 mt-4 font-extrabold">📞 زبائن استفسروا أكثر من مرة وما حجزوا</h4>
                        <ul className="space-y-1.5 text-sm">
                          {(biz?.inquirers ?? []).slice(0, 15).map((c) => (
                            <li key={c.id} className="flex items-center justify-between gap-2 rounded-lg bg-[var(--mx-sunken)] p-2">
                              <Link to={`/customers?q=${encodeURIComponent(c.name)}`} className="font-bold text-[var(--mx-accent)] hover:underline">👤 {c.name} ←</Link>
                              <span className="text-xs text-[var(--mx-muted)]">{c.archived} استفسار · آخرها {new Date(c.lastAt).toLocaleDateString('ar-IQ', { timeZone: 'Asia/Baghdad' })}</span>
                            </li>
                          ))}
                        </ul>
                      </>
                    )}
                  </>
                ) : (
                  <>
                    <h4 className="mb-3 font-extrabold">👥 الموظفين الي يحتاجون متابعة، ووين</h4>
                    {weak.length === 0 ? <p className="text-sm text-[var(--mx-muted)]">ماكو أحد يحتاج متابعة هسه ✅</p> : (
                      <ul className="space-y-2 text-sm">
                        {weak.map((m) => (
                          <li key={m.id} className="rounded-lg bg-[var(--mx-sunken)] p-2">
                            <div className="flex flex-wrap items-center justify-between gap-2">
                              <Link to={`/matrix/employee/${m.id}`} className="font-extrabold text-[var(--mx-accent)] hover:underline">👤 {m.name} — تقريره ←</Link>
                              <span className="text-xs text-[var(--mx-muted)]">{GROUP_LABEL[m.group as EyeGroup] ?? m.group} · مؤشر {m.idx}%</span>
                            </div>
                            <ul className="mt-1 list-inside list-disc text-xs">{m.why.map((w) => <li key={w}>{w}</li>)}</ul>
                          </li>
                        ))}
                      </ul>
                    )}
                    <p className="mt-2 text-[11px] text-[var(--mx-muted)]">الأسباب أرقام من النظام — التفاصيل والسبب الي كتبه الموظف بتقريره. ماكو نقاط ولا غرامات.</p>
                  </>
                )}
              </div>
            ) : modal === 'evidence' ? (
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

// ── توقعات ماتركس (أعمال بس — بلا توقعات شخصية) ──
function Predictions({ biz, late, props }: { biz: MatrixBusinessData | null; late: LateFocus | null; props: MatrixProposal[] }) {
  const items: { text: string; change: number | null; conf: number }[] = []
  if (biz) {
    const ch = biz.mtd.lastBookings > 0 ? Math.round(((biz.mtd.bookings - biz.mtd.lastBookings) / biz.mtd.lastBookings) * 100) : null
    items.push({ text: `المتوقع لآخر الشهر ${biz.forecast.expectedJobs} حجز`, change: ch, conf: biz.forecast.insufficient ? 50 : confidence(biz.mtd.bookings) })
    const proj = biz.mtd.projectRevenue ?? 0
    const bR = biz.mtd.revenue - proj, bL = biz.mtd.lastRevenue - (biz.mtd.lastProjectRevenue ?? 0)
    const rc = bL > 0 ? Math.round(((bR - bL) / bL) * 100) : null
    // إيراد الحجوزات المتوقع بس — دفعات المشاريع كتل تنسجّل بأي وكت فما تنتوقّع
    items.push({ text: `إيراد الحجوزات المتوقع لآخر الشهر ${fmtIQD(Math.max(0, biz.forecast.expected - proj))}`, change: rc, conf: biz.forecast.insufficient ? 50 : confidence(biz.mtd.bookings) })
    if (proj > 0) items.push({ text: `دفعات مشاريع مقبوضة هالشهر لحد هسه ${fmtIQD(proj)} (ما تدخل بالتوقع)`, change: null, conf: 100 })
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

// ═══ العين الكبيرة بمركز القيادة — قرار (ع) 10-13 ═══
// شارنغان دائمي (نفس عين المالك)، وضغطتين عليها تحوّلها مانغكيو دقيقة
// وترجع. هنا بس — عين الشريط العلوي ما تتأثر.
// (ع) 10-13: بلا حواجب — نرسمها دايماً «هادئة» (الشارنغان يفرض اللون الأحمر)
function BigSharingan() {
  const [mangekyo, setMangekyo] = useState(false)
  const t = useRef(0)
  useEffect(() => () => window.clearTimeout(t.current), [])
  // رمشة طبيعية كل ٣–٦ ثواني
  const [blink, setBlink] = useState(false)
  useEffect(() => {
    let id = 0
    const loop = () => { id = window.setTimeout(() => { setBlink(true); window.setTimeout(() => setBlink(false), 140); loop() }, 3000 + Math.random() * 3000) }
    loop()
    return () => window.clearTimeout(id)
  }, [])
  const toggle = () => {
    window.clearTimeout(t.current)
    setMangekyo(true)
    t.current = window.setTimeout(() => setMangekyo(false), 60_000)
  }
  return (
    <button type="button" onDoubleClick={toggle} title="ضغطتين: مانغكيو شارنغان لمدة دقيقة"
      className={`shrink-0 select-none transition ${mangekyo ? 'drop-shadow-[0_0_40px_rgba(220,38,38,0.85)]' : 'drop-shadow-[0_0_30px_rgba(185,28,28,0.55)]'}`}
      style={{ ['--eye' as string]: '#b91c1c' }}>
      <MatrixEyeGraphic group="ADMINS" mood="CALM" width={190} lid={blink ? 0.06 : 0.95} sharingan={mangekyo ? 2 : 1} />
    </button>
  )
}
