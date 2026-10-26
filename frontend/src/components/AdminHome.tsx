import MatrixCommandCenter from './MatrixCommandCenter'
import MatrixChat from './MatrixChat'
import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import {
  api,
  type AiSignal,
  type EmployeeDailyAttendanceSummary,
  type FinanceSummary,
  type InternalWorksReport,
  type NewEmployeeCurve,
  type StockForecast,
  type TechnicianKpi,
  type TodayPulse,
} from '../api'
import TodayBoard from './TodayBoard'
import ActionsTable from './ActionsTable'
import MatrixMissions from './MatrixMissions'
import ReplacementSuggestionsPanel from './ReplacementSuggestionsPanel'
import { pollVisible } from '../utils/poll'
import { drillTo } from '../utils/drill'
import { roleLabel } from '../roleLabels'

// ═══ رئيسية مدير النظام: لوحتين ═══
//
// (ع): «مدير النظام ميريد خانة اسمها الرئيسية — يريد الواجهة تنقسم
// لوحتين». «المتابعة» يشوف بيها كل شي يحتاج مراقبة، و«الإجراءات» كل
// شي ينتظر قراره هو.
//
// 🌳 والذكاء «عروق الشجرة»: أحكام ماتركس تنعرض جوّا لوحة المتابعة نفسها
// يمّ أرقام الموظفين، مو بشاشة معزولة لازم يتذكر يفتحها.
//
// ⚠️ كل قسم يجيب بياناته لحاله ويفشل لحاله: لو مسار وحد تعطّل، باقي
// اللوحة تبقى تشتغل، والقسم يگول «ما وصل» بدل رقم صفر يكذب.

type Board = 'follow' | 'actions' | 'matrix' | 'chat' | 'missions'

const fmt = (n: number) => n.toLocaleString('en-US')
// قرار (ع) 10-09: المبالغ تقريبية لأقرب ألف — بلا كسور مثل ٨٣٢
const money = (n: number) => `${fmt(Math.round(n / 1000) * 1000)} د.ع`
const monthKey = () => new Date().toISOString().slice(0, 7)

const SEVERITY: Record<string, { label: string; rank: number; cls: string }> = {
  CRITICAL: { label: 'حرج', rank: 4, cls: 'bg-red-100 text-red-700' },
  WARN: { label: 'تنبيه', rank: 3, cls: 'bg-amber-100 text-amber-800' },
  WATCH: { label: 'مراقبة', rank: 2, cls: 'bg-sky-100 text-sky-700' },
  INFO: { label: 'معلومة', rank: 1, cls: 'bg-slate-100 text-slate-600' },
}

function useLoad<T>(fn: () => Promise<T>): T | null | undefined {
  // undefined = جاري التحميل، null = فشل
  const [v, setV] = useState<T | null | undefined>(undefined)
  useEffect(() => {
    let alive = true
    fn().then((r) => { if (alive) setV(r) }).catch(() => { if (alive) setV(null) })
    return () => { alive = false }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])
  return v
}

function Section({ title, to, linkLabel = 'التفاصيل ←', children }: { title: string; to?: string; linkLabel?: string; children: ReactNode }) {
  return (
    <section className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <div className="mb-3 flex items-center justify-between gap-2">
        <h3 className="text-base font-extrabold text-[#0f2040]">{title}</h3>
        {to && <Link to={to} className="text-xs font-bold text-brand-600 hover:underline">{linkLabel}</Link>}
      </div>
      {children}
    </section>
  )
}

// (ع): «من اضغط ع شي يوديني عليه حتى أشوف بصورة صحيحة» — كل مربع رقم يفتح
// الشاشة الي تفصّله (to).
type Tone = 'slate' | 'green' | 'amber' | 'red' | 'blue' | 'violet'
const TONE: Record<Tone, { card: string; ink: string; chip: string }> = {
  slate: { card: 'border-slate-200 bg-slate-50/70', ink: 'text-slate-800', chip: 'bg-white text-slate-500' },
  blue: { card: 'border-sky-100 bg-sky-50/70', ink: 'text-sky-800', chip: 'bg-sky-100 text-sky-600' },
  green: { card: 'border-emerald-100 bg-emerald-50/70', ink: 'text-emerald-700', chip: 'bg-emerald-100 text-emerald-600' },
  amber: { card: 'border-amber-100 bg-amber-50/70', ink: 'text-amber-700', chip: 'bg-amber-100 text-amber-600' },
  red: { card: 'border-red-100 bg-red-50/70', ink: 'text-red-700', chip: 'bg-red-100 text-red-600' },
  violet: { card: 'border-violet-100 bg-violet-50/70', ink: 'text-violet-700', chip: 'bg-violet-100 text-violet-600' },
}

// بطاقة رقم بتصميم (ع) 10-04: لون خفيف، أيقونة بدائرة، والرقم كبير.
// (ع): «من اضغط ع شي يوديني عليه» — كل بطاقة تفتح الشاشة الي تفصّلها (to).
// تصميم (ع) 10-13: سهم ‹ تحت لكل بطاقة تودّي لتفاصيلها، وشريط نسبة اختياري.
function Stat({ label, value, hint, tone = 'slate', to, icon, big, pct, onClick, active }: { label: string; value: ReactNode; hint?: ReactNode; tone?: Tone; to?: string; icon?: string; big?: boolean; pct?: number | null; onClick?: () => void; active?: boolean }) {
  const t = TONE[tone]
  const p = pct == null || !isFinite(pct) ? null : Math.max(0, Math.min(100, Math.round(pct)))
  const body = (
    <>
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className={`text-[13px] font-extrabold leading-snug ${tone === 'slate' ? 'text-slate-700' : t.ink}`}>{label}</p>
          <p className={`mt-1 font-extrabold tabular-nums text-[#0f2040] ${big ? 'text-3xl' : 'text-2xl'}`}>{value}</p>
          {/* (ع) 10-16: سطر الشرح ينقرا كامل — مو مقصوص */}
          {hint && <p className="mt-0.5 text-[11px] leading-snug text-slate-500">{hint}</p>}
        </div>
        <div className="flex shrink-0 flex-col items-center gap-2">
          {icon && <span className={`grid place-items-center rounded-2xl ${big ? 'h-14 w-14 text-2xl' : 'h-12 w-12 text-xl'} ${t.chip}`}>{icon}</span>}
          {(to || onClick) && <span className="grid h-6 w-6 place-items-center rounded-full bg-white text-xs text-slate-500 shadow-sm">‹</span>}
        </div>
      </div>
      {p != null && (
        <div className="mt-2 flex items-center gap-2">
          <div className="h-2 flex-1 rounded-full bg-white/80"><div className={`h-2 rounded-full ${tone === 'blue' ? 'bg-sky-500' : tone === 'amber' ? 'bg-amber-500' : 'bg-emerald-500'}`} style={{ width: `${p}%` }} /></div>
          <span className={`text-[11px] font-extrabold tabular-nums ${t.ink}`}>{p}%</span>
        </div>
      )}
    </>
  )
  const cls = `block rounded-2xl border p-3.5 transition ${t.card}${active ? ' ring-2 ring-brand-500' : ''}`
  if (onClick) return <button type="button" onClick={onClick} className={`${cls} w-full text-right hover:-translate-y-0.5 hover:shadow-md`}>{body}</button>
  return to
    ? <Link to={to} className={`${cls} hover:-translate-y-0.5 hover:shadow-md`}>{body}</Link>
    : <div className={cls}>{body}</div>
}

function Missing() {
  return <p className="text-xs text-slate-400">ما وصلت البيانات — افتح التفاصيل.</p>
}

function Loading() {
  return <div className="h-16 animate-pulse rounded-xl bg-slate-100" />
}

// ═══ ماتركس — منحنى تعلّم الموظف الجديد ═══
// خط أسبوعي صغير لنقاط كل أسبوع (منجز − ٢×توقف − ٣×شكوى). الحالة
// محسوبة بالخادم بقاعدة ثابتة: ٣ أسابيع مكتملة متتالية بلا تحسّن = يحتاج متابعة.
function Sparkline({ values }: { values: number[] }) {
  if (values.length < 2) return <span className="text-[11px] text-slate-400">—</span>
  const w = 84, h = 22
  const min = Math.min(...values), max = Math.max(...values)
  const span = max - min || 1
  const pts = values.map((v, i) => `${(i / (values.length - 1)) * w},${h - 2 - ((v - min) / span) * (h - 4)}`).join(' ')
  return (
    <svg width={w} height={h} viewBox={`0 0 ${w} ${h}`} className="shrink-0" aria-hidden>
      <polyline points={pts} fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinejoin="round" strokeLinecap="round" />
    </svg>
  )
}

const CURVE_TONE: Record<NewEmployeeCurve['status'], string> = {
  OK: 'bg-emerald-50 text-emerald-700',
  NEEDS_FOLLOWUP: 'bg-red-50 text-red-700',
  INSUFFICIENT: 'bg-slate-100 text-slate-500',
}
const CURVE_LABEL: Record<NewEmployeeCurve['status'], string> = {
  OK: 'ماشي طبيعي',
  NEEDS_FOLLOWUP: 'يحتاج متابعة',
  INSUFFICIENT: 'ماكو بيانات كافية بعد',
}

function NewEmployeeCurves() {
  const curves = useLoad<NewEmployeeCurve[]>(() => api.getNewEmployeeCurves())
  return (
    <Section title="🌱 ماتركس — منحنى الموظفين الجدد (أول ٩٠ يوم)">
      {curves === undefined ? <Loading /> : curves === null ? <Missing /> : curves.length === 0 ? (
        <p className="text-sm text-slate-500">ماكو فنيين أو ليدرية جدد بآخر ٩٠ يوم.</p>
      ) : (
        <div className="overflow-x-auto">
          <table className="w-full min-w-[760px] text-sm">
            <thead>
              <tr className="border-b border-slate-100 text-[12px] text-slate-500">
                <th className="py-2 text-right font-bold">#</th>
                <th className="py-2 text-right font-bold">الموظف</th>
                <th className="py-2 text-right font-bold">فترة المتابعة</th>
                <th className="py-2 text-center font-bold">الحالة</th>
                <th className="py-2 text-center font-bold">مؤشر الأداء (كل أسبوع)</th>
                <th className="py-2 text-center font-bold"><i className="me-1 inline-block h-2 w-2 rounded-full bg-emerald-500" />منجز</th>
                <th className="py-2 text-center font-bold"><i className="me-1 inline-block h-2 w-2 rounded-full bg-amber-500" />توقف</th>
                <th className="py-2 text-center font-bold"><i className="me-1 inline-block h-2 w-2 rounded-full bg-red-500" />شكاوي</th>
              </tr>
            </thead>
            <tbody>
              {curves.map((c, i) => {
                const last = [...c.weeks].reverse().find((w) => w.complete)
                return (
                  <tr key={c.employeeId} className="border-b border-slate-50 hover:bg-slate-50/60">
                    <td className="py-2.5 text-xs font-bold text-slate-400">{String(i + 1).padStart(2, '0')}</td>
                    <td className="py-2.5">
                      <Link to={`/matrix/employee/${c.employeeId}`} className="flex items-center gap-2 font-extrabold text-[#0f2040] hover:underline">
                        <span className="grid h-8 w-8 shrink-0 place-items-center rounded-full bg-sky-50 text-xs">👤</span>
                        {c.name}{c.isLeader ? ' — ليدر' : ''}
                      </Link>
                    </td>
                    <td className="py-2.5 text-xs">
                      <p className="font-bold text-slate-700">منذ {fmt(c.daysIn)} يوم</p>
                      <p className="text-[11px] text-slate-400">الأسبوع {fmt(Math.min(13, Math.ceil(c.daysIn / 7) || 1))} من ١٣</p>
                    </td>
                    <td className="py-2.5 text-center">
                      <span className={`rounded-full px-2.5 py-1 text-[11px] font-bold ${CURVE_TONE[c.status]}`} title={c.statusNote}>{CURVE_LABEL[c.status]}</span>
                    </td>
                    <td className="py-2.5 text-center">
                      <span className={`inline-block ${c.status === 'NEEDS_FOLLOWUP' ? 'text-red-500' : 'text-emerald-600'}`} title="نقاط كل أسبوع"><Sparkline values={c.weeks.map((w) => w.score)} /></span>
                    </td>
                    <td className="py-2.5 text-center text-base font-extrabold text-emerald-700 tabular-nums">{fmt(last?.completed ?? 0)}</td>
                    <td className={`py-2.5 text-center text-base font-extrabold tabular-nums ${(last?.stops ?? 0) > 0 ? 'text-amber-600' : 'text-slate-400'}`}>{fmt(last?.stops ?? 0)}</td>
                    <td className={`py-2.5 text-center text-base font-extrabold tabular-nums ${(last?.complaints ?? 0) > 0 ? 'text-red-600' : 'text-slate-400'}`}>{fmt(last?.complaints ?? 0)}</td>
                  </tr>
                )
              })}
            </tbody>
          </table>
        </div>
      )}
      <p className="mt-2 text-[11px] text-slate-400">نقاط الأسبوع = منجز − ٢×توقف − ٣×شكوى. «يحتاج متابعة» = ٣ أسابيع مكتملة بلا تحسّن — للمتابعة والتدريب، مو عقوبة.</p>
    </Section>
  )
}

// ═══ «منو فاتح النظام هسه» — طلب (ع) 10-04 ═══
// أخضر = استخدم النظام بآخر ٥ دقايق، رمادي = بآخر ساعة. يتحدّث كل دقيقة.
function OnlineNow() {
  const [rows, setRows] = useState<Awaited<ReturnType<typeof api.getOnlineEmployees>> | null | undefined>(undefined)
  const [now, setNow] = useState(() => Date.now())
  useEffect(() => {
    let alive = true
    const load = () => api.getOnlineEmployees().then((r) => { if (alive) { setRows(r); setNow(Date.now()) } }).catch(() => { if (alive) setRows(null) })
    load()
    const stop = pollVisible(load, 60_000)
    return () => { alive = false; stop() }
  }, [])
  const mins = (s: string) => Math.max(0, Math.round((now - new Date(s).getTime()) / 60000))
  const live = (rows ?? []).filter((r) => mins(r.lastSeenAt) <= 5)
  const recent = (rows ?? []).filter((r) => mins(r.lastSeenAt) > 5)
  return (
    <Section title={`🟢 المتصلين هسه (${live.length})`}>
      {rows === undefined ? <Loading /> : rows === null ? <Missing /> : rows.length === 0 ? (
        <p className="text-sm text-slate-500">محد فاتح النظام هسه.</p>
      ) : (
        <div className="space-y-2">
          <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-6">
            {live.map((r) => (
              <div key={r.id} className="flex items-center gap-2.5 rounded-2xl border border-emerald-100 bg-emerald-50/50 px-3 py-2.5">
                <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-sky-50 text-sm">👤</span>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-extrabold text-[#0f2040]">{r.name}</p>
                  <p className="truncate text-[11px] text-slate-500">{roleLabel(r.role)} · {mins(r.lastSeenAt) <= 1 ? 'هسه' : `قبل ${mins(r.lastSeenAt)} د`}</p>
                </div>
                <i className="h-2.5 w-2.5 shrink-0 rounded-full bg-emerald-500" />
              </div>
            ))}
          </div>
          {recent.length > 0 && (
            <div className="flex flex-wrap items-center gap-1.5">
              <span className="text-[11px] text-slate-400">ظهروا بآخر ساعة:</span>
              {recent.map((r) => (
                <span key={r.id} className="inline-flex items-center gap-1 rounded-full bg-slate-100 px-2.5 py-0.5 text-[11px] text-slate-600">
                  <i className="h-1.5 w-1.5 rounded-full bg-slate-400" />{r.name} · قبل {mins(r.lastSeenAt)} د
                </span>
              ))}
            </div>
          )}
        </div>
      )}
      <p className="mt-2 text-[11px] text-slate-400">«متصل» = استخدم النظام بآخر ٥ دقايق (أي ضغطة أو فتح شاشة).</p>
    </Section>
  )
}

// ═══ ماتركس — المخزون قبل ما يخلص ═══
function StockForecastSection() {
  const [reload, setReload] = useState(0)
  const [f, setF] = useState<StockForecast | null | undefined>(undefined)
  useEffect(() => {
    let alive = true
    api.getStockForecast().then((r) => { if (alive) setF(r) }).catch(() => { if (alive) setF(null) })
    return () => { alive = false }
  }, [reload])
  // جرد: الرصيد الموجود هسه — ماتركس يطرح المصروف بعده لحاله.
  const count = async (id: string, name: string, current?: number) => {
    const v = prompt(`جرد «${name}»: شگد موجود هسه بالمخزن؟`, current != null ? String(Math.max(0, Math.round(current))) : '')
    if (v === null || v.trim() === '' || Number.isNaN(Number(v))) return
    try { await api.setMaterialStock(id, Number(v)); setReload((x) => x + 1) } catch (e) { alert(e instanceof Error ? e.message : 'تعذّر الحفظ') }
  }
  return (
    <Section title="📦 ماتركس — المواد قبل ما تخلص" to="/procurement" linkLabel="طلبات المواد ←">
      {f === undefined ? <Loading /> : f === null ? <Missing /> : (
        <>
          <p className={`mb-2 rounded-xl px-3 py-2 text-xs ${f.stockTracked ? 'bg-slate-50 text-slate-600' : 'bg-amber-50 text-amber-800'}`}>{f.note}</p>
          {f.items.length > 0 && (
            <ul className="space-y-1 text-sm">
              {f.items.slice(0, 10).map((m) => (
                <li key={m.key} className="flex items-center justify-between gap-2 text-slate-600">
                  <span className="truncate">{m.name}</span>
                  <span className="flex shrink-0 items-center gap-2 tabular-nums text-slate-500">
                    {m.daysLeft != null ? (
                      <b className={m.daysLeft <= 7 ? 'text-red-600' : m.daysLeft <= 14 ? 'text-amber-600' : 'text-emerald-700'}>
                        باقي {fmt(Math.max(0, m.remaining ?? 0))} · يكفي ~{m.daysLeft} يوم
                      </b>
                    ) : <>{m.avgDaily.toLocaleString('en-US')}/يوم · يحتاج ~{fmt(m.need14Days)} لـ{f.horizonDays} يوم</>}
                    {!m.key.startsWith('name:') && (
                      <button onClick={() => count(m.key, m.name, m.remaining)} className="rounded border border-slate-200 px-1.5 text-[10px] text-slate-500 hover:bg-slate-50" title="سجّل الرصيد الموجود هسه">جرد</button>
                    )}
                  </span>
                </li>
              ))}
            </ul>
          )}
          <p className="mt-2 text-[11px] text-slate-400">من فواتير الليدرية آخر {f.windowDays} يوم، المواد الي انصرفت {f.minUsages} مرات فأكثر.</p>
        </>
      )}
    </Section>
  )
}

// ═══ لوحة المتابعة ═══
// monitor: نسخة المراقب — بلا حسابات الشركة (بس التدقيق) وبلا نداءات المدير الحصرية.
export function FollowBoard({ monitor = false }: { monitor?: boolean } = {}) {
  const summary = useLoad(() => api.getDashboardSummary())
  const pulse = useLoad<TodayPulse>(() => api.getTodayPulse())
  const finance = useLoad<FinanceSummary>(() => api.getFinanceSummary())
  const internal = useLoad<InternalWorksReport>(() => monitor ? Promise.reject(new Error('skip')) : api.getInternalWorks(monthKey()))
  const attendance = useLoad<EmployeeDailyAttendanceSummary[]>(() => api.getTodaySummary())
  const signals = useLoad<AiSignal[]>(() => monitor ? Promise.reject(new Error('skip')) : api.getAiSignals())
  const kpi = useLoad<TechnicianKpi[]>(() => api.getKpiLeaderboard(monthKey()))
  // الوقت ينثبت بفتح الشاشة — قراءة الساعة أثناء الرسم تخلي النتيجة تتغيّر بلا سبب.
  const [since] = useState(() => Date.now() - 7 * 24 * 3600 * 1000)

  // ماتركس لكل موظف: آخر ٧ أيام، أعلى خطورة وآخر حكم.
  const matrix = useMemo(() => {
    if (!signals) return []
    const by = new Map<string, { name: string; count: number; top: string; headline: string; at: number }>()
    for (const s of signals) {
      const v = s.verdict
      if (!v) continue
      const at = new Date(s.occurredAt || s.createdAt).getTime()
      if (at < since) continue
      const id = v.blameEmployeeId || s.employeeId
      const name = v.blameEmployeeName || s.employeeName
      if (!id || !name) continue
      const cur = by.get(id) || { name, count: 0, top: 'INFO', headline: v.headline, at: 0 }
      cur.count++
      if ((SEVERITY[v.severity]?.rank || 0) > (SEVERITY[cur.top]?.rank || 0)) cur.top = v.severity
      if (at > cur.at) { cur.at = at; cur.headline = v.headline }
      by.set(id, cur)
    }
    return [...by.entries()]
      .map(([id, r]) => ({ id, ...r }))
      .sort((a, b) => (SEVERITY[b.top]?.rank || 0) - (SEVERITY[a.top]?.rank || 0) || b.count - a.count)
  }, [signals, since])

  const [attFilter, setAttFilter] = useState<'all' | 'active' | 'left' | 'absent'>('all')
  const [absentees, setAbsentees] = useState<{ id: string; name: string }[] | null>(null)
  useEffect(() => {
    if (attFilter !== 'absent' || absentees !== null || !attendance) return
    const here = new Set(attendance.map((x) => x.employeeId))
    api.getEmployees()
      .then((all) => setAbsentees(all.filter((e) => e.status === 'ACTIVE' && !here.has(e.id)).map((e) => ({ id: e.id, name: e.name }))))
      .catch(() => setAbsentees([]))
  }, [attFilter, absentees, attendance])
  const present = attendance?.length ?? 0
  const activeNow = attendance?.filter((a) => a.currentlyActive).length ?? 0

  return (
    <div className="space-y-4">
      {/* تصميم (ع) 10-13: المؤشرات الرئيسية أول شي */}
      <Section title="📊 المؤشرات الرئيسية اليوم" to="/missions" linkLabel="تتبع المهام كامل ←">
        {pulse === undefined ? <Loading /> : pulse === null ? <Missing /> : (
          <div className="grid grid-cols-2 gap-3 sm:grid-cols-3 lg:grid-cols-5">
            {/* 🔴 (ع) 10-16 «الرقم = القائمة»: كل بطاقة تفتح نفس الي عدّته بالضبط */}
            <Stat big icon="📅" tone="blue" label="حجوزات اليوم" to={drillTo('today_scheduled')} value={fmt(pulse.todayBookings)} hint={`موعدها اليوم · أمس جانت ${fmt(pulse.yesterdayBookings)}`} />
            <Stat big icon="📄" tone="blue" label="المهام المفتوحة" to={drillTo('open_missions')} value={fmt(pulse.openMissions)} hint="حجوزات عليها مهمة بعدها ما خلصت — كل الفترات" />
            <Stat big icon="⚠️" tone={pulse.newComplaints ? 'red' : 'slate'} label="الشكاوى الجديدة" to="/complaints?status=NEW" value={fmt(pulse.newComplaints)} hint="شكاوى بعد ما أحد اشتغل عليها" />
            <Stat big icon="⏳" tone={pulse.overdueMissions ? 'amber' : 'slate'} label="المتأخرات" to={drillTo('overdue')} value={fmt(pulse.overdueMissions)} hint="تواصل +ساعتين · كادر ما طلع · تنفيذ طوّل · ورق ناقص +٢٤ ساعة" />
            <Stat big icon="👷" tone="green" label="الكوادر بالميدان" to={drillTo('in_field')} value={fmt(pulse.crewInField)} hint="عدد الموظفين الطالعين هسه — القائمة بالحجوزات الي طالعين عليها" />
          </div>
        )}
      </Section>

      <div className="grid gap-4 lg:grid-cols-2">
        <Section title={monitor ? '🪙 حسابات التدقيق' : '💰 الحسابات'} to={monitor ? undefined : '/finance'}>
          {finance === undefined ? <Loading /> : finance === null ? <Missing /> : (
            <>
            {(() => {
              const booked = finance.bookingsCollected ?? finance.totalCollected
              const projects = finance.projectsCollected ?? 0
              const audited = finance.verifiedAmount + finance.unverifiedAmount
              return (
                <div className="grid grid-cols-1 gap-3 sm:grid-cols-2">
                  {/* قرار (ع) 10-09: فلوس الحجوزات مفصولة عن فلوس المشاريع. النِّسب كلها من نفس الأرقام */}
                  {!monitor && <>
                  <Stat icon="💳" label="المقبوض من الحجوزات" to="/finance" value={money(booked)} hint={`نصيبها من كل المقبوض ${money(finance.totalCollected)}`} tone="green" pct={finance.totalCollected ? (booked / finance.totalCollected) * 100 : null} />
                  <Stat icon="🏗️" label="المقبوض من المشاريع" to="/project-payments" value={money(projects)} hint={`نصيبها من كل المقبوض ${money(finance.totalCollected)}`} tone="green" pct={finance.totalCollected ? (projects / finance.totalCollected) * 100 : null} />
                  </>}
                  <Stat icon="🪙" tone="blue" label="مدقّق" to={drillTo('verified')} value={money(finance.verifiedAmount)} hint={`${fmt(finance.verifiedCount)} حجز المحاسب دقّق مبلغه · من أصل ${money(audited)} — كل الأيام`} pct={audited ? (finance.verifiedAmount / audited) * 100 : null} />
                  <Stat icon="⏰" label="بانتظار التدقيق" to={drillTo('unverified')} value={money(finance.unverifiedAmount)} hint={`${fmt(finance.unverifiedCount)} حجز منجز بعده ما تدقّق — كل الأيام`} tone={finance.unverifiedCount ? 'amber' : 'slate'} />
                </div>
              )
            })()}
            <div className="mt-3 grid grid-cols-3 gap-2">
              {!monitor && <Stat icon="🧾" label="المصاريف" to="/expenses" value={money(finance.totalExpenseValue)} hint={`${fmt(finance.pendingExpenses)} بانتظار الموافقة`} tone="red" />}
              <Stat icon="✅" tone="violet" label="أنجزت اليوم" to={drillTo('completed_today')} value={fmt(finance.todayCompleted)} hint={`تأشّرت منجز اليوم · وانسجّل اليوم ${fmt(finance.todayCreated)} جديد`} />
              <Stat icon="⚙️" label="قيد التنفيذ" to={drillTo('in_progress')} value={fmt(finance.inProgressCount)} hint={`مثبّتة ${fmt(finance.confirmedCount)} · معلّقة ${fmt(finance.pendingCount)}`} />
            </div>
            </>
          )}
        </Section>

        {!monitor && <Section title="🏭 أعمال داخل الشركة (هذا الشهر)" to="/stats-management">
          {internal === undefined ? <Loading /> : internal === null ? <Missing /> : (
            <>
              <div className="grid grid-cols-3 gap-2">
                <Stat icon="🏢" tone="blue" label="داخل الشركة" to={drillTo('inhouse_month')} value={fmt(internal.inHouseCount)} hint="حجز انخلص جوّا الشركة هذا الشهر" />
                <Stat icon="📍" tone="blue" label="بالموقع" to={drillTo('onsite_month')} value={fmt(internal.onSiteCount)} hint="حجز انخلص عند الزبون هذا الشهر" />
                <Stat icon="💵" label="المبلغ" to="/stats-management" value={money(internal.inHouseAmount)} tone="green" hint="مبلغ الشغل الداخلي هذا الشهر — التفصيل بالإحصائيات" />
              </div>
              {internal.services.length > 0 && (
                <div className="mt-4">
                  <p className="mb-2 text-sm font-extrabold text-[#0f2040]">☰ تفاصيل سريعة</p>
                  <ul className="space-y-1.5 text-sm">
                    {internal.services.slice(0, 6).map((sv, i) => (
                      <li key={sv.name}>
                        <Link to="/stats-management" className="flex items-center gap-2 rounded-xl bg-slate-50 px-3 py-2 hover:bg-sky-50">
                          <i className={`h-2.5 w-2.5 shrink-0 rounded-full ${['bg-sky-500', 'bg-violet-500', 'bg-amber-500', 'bg-emerald-500', 'bg-red-500', 'bg-slate-400'][i % 6]}`} />
                          <span className="flex-1 font-bold text-slate-700">{sv.name}</span>
                          <span className="tabular-nums text-slate-500">{fmt(sv.count)} · {money(sv.amount)}</span>
                          <span className="grid h-6 w-6 place-items-center rounded-full bg-white text-xs text-slate-500 shadow-sm">‹</span>
                        </Link>
                      </li>
                    ))}
                  </ul>
                </div>
              )}
            </>
          )}
        </Section>}
      </div>

      {/* «شنو صاير اليوم» و«شغلي اليوم» والرسم وأكثر الكوادر — ترتيب (ع) 10-04 */}
      <TodayBoard finance={finance ?? null} />

      <div className="grid gap-4 lg:grid-cols-2">
        {!monitor && <Section title="🧠 ماتركس — تقييمات الموظفين (آخر ٧ أيام)" to="/ai-insights" linkLabel="كل المؤشرات ←">
          {signals === undefined ? <Loading /> : signals === null ? <Missing /> : matrix.length === 0 ? (
            <p className="text-sm text-slate-500">ماكو أحكام على موظفين هالأسبوع ✅</p>
          ) : (
            <ul className="divide-y divide-slate-100">
              {matrix.slice(0, 8).map((m) => (
                <li key={m.id}>
                  <Link to={`/matrix/employee/${m.id}`} className="flex items-center gap-3 py-2 hover:bg-slate-50">
                    <span className="grid h-9 w-9 shrink-0 place-items-center rounded-full bg-sky-50 text-sm">👤</span>
                    <div className="min-w-0 flex-1">
                      <p className="truncate font-bold text-slate-700">{m.name}</p>
                      <p className="truncate text-xs text-slate-500" title={m.headline}>{m.headline}</p>
                    </div>
                    <span className="shrink-0 text-xs text-slate-400">{fmt(m.count)} حكم</span>
                    <span className={`shrink-0 rounded-full px-2.5 py-0.5 text-[11px] font-bold ${SEVERITY[m.top]?.cls || ''}`}>{SEVERITY[m.top]?.label || m.top}</span>
                    <span className="shrink-0 text-slate-300">‹</span>
                  </Link>
                </li>
              ))}
            </ul>
          )}
          <p className="mt-2 text-[11px] text-slate-400">الأحكام توجيه ومراجعة — ماكو غرامة تنطلع منها تلقائياً.</p>
        </Section>}

        <Section title="🏆 ترتيب الأداء (هذا الشهر)" to="/kpi">
          {kpi === undefined ? <Loading /> : kpi === null ? <Missing /> : kpi.length === 0 ? (
            <p className="text-sm text-slate-500">ماكو نقاط محسوبة بعد.</p>
          ) : (
            (() => {
              const rows = [...kpi].sort((a, b) => b.totalPoints - a.totalPoints).slice(0, 6)
              const max = Math.max(1, ...rows.map((r) => r.totalPoints))
              const medal = ['bg-amber-100 text-amber-700', 'bg-slate-200 text-slate-600', 'bg-orange-100 text-orange-700']
              const pill = ['bg-emerald-50 text-emerald-700', 'bg-sky-50 text-sky-700', 'bg-violet-50 text-violet-700']
              return (
                <ol className="space-y-2 text-sm">
                  {rows.map((k, i) => (
                    <li key={k.employeeId} className="flex items-center gap-3">
                      <span className={`grid h-7 w-7 shrink-0 place-items-center rounded-full text-xs font-extrabold ${medal[i] ?? 'bg-slate-100 text-slate-500'}`}>{i + 1}</span>
                      <span className="grid h-8 w-8 shrink-0 place-items-center rounded-full bg-sky-50 text-sm">👤</span>
                      <span className="w-36 truncate font-bold text-slate-700">{k.employeeName}</span>
                      <div className="h-2 flex-1 rounded-full bg-slate-100"><div className="h-2 rounded-full bg-gradient-to-l from-sky-500 to-indigo-500" style={{ width: `${(k.totalPoints / max) * 100}%` }} /></div>
                      <span className={`w-14 rounded-full px-2 py-0.5 text-center text-xs font-extrabold tabular-nums ${pill[i] ?? 'bg-slate-50 text-slate-700'}`}>{fmt(k.totalPoints)}</span>
                    </li>
                  ))}
                </ol>
              )
            })()
          )}
        </Section>
      </div>

      <OnlineNow />

      <Section title="🕒 سجل الدوام اليوم" to="/attendance" linkLabel="سجل الدوام ←">
        {attendance === undefined ? <Loading /> : attendance === null ? <Missing /> : (
          <>
            <div className="grid grid-cols-2 gap-3 lg:grid-cols-4">
              {/* (ع) 10-16: البطاقة تفلتر الأسماء جوّا بدل ما الأربعة يودّن لنفس الشاشة */}
              <Stat big icon="👥" tone="green" label="إجمالي الحضور اليوم" onClick={() => setAttFilter('all')} active={attFilter === 'all'} value={fmt(present)} hint={summary ? `من ${fmt(summary.employeeCount)} موظف — اضغط تشوف أسماءهم جوّا` : undefined} />
              <Stat big icon="🕘" tone="blue" label="على الدوام هسه" onClick={() => setAttFilter('active')} active={attFilter === 'active'} value={fmt(activeNow)} hint="مسجّل حضور ولسه ما طلع" />
              <Stat big icon="🚪" tone="amber" label="انصرفوا" onClick={() => setAttFilter('left')} active={attFilter === 'left'} value={fmt(present - activeNow)} hint="سجّلوا حضور وطلعوا" />
              {summary && <Stat big icon="❌" tone="red" label="ما سجّلوا حضور" onClick={() => setAttFilter('absent')} active={attFilter === 'absent'} value={fmt(Math.max(0, summary.employeeCount - present))} hint="من كل الموظفين (بيهم الإجازات والعطل)" />}
            </div>
            {attFilter === 'absent' ? (
              <div className="mt-3 flex flex-wrap items-center gap-1.5">
                <span className="text-xs font-bold text-slate-500">❌ ما سجّلوا حضور اليوم</span>
                {absentees === null ? <span className="text-xs text-slate-400">جاري التحميل…</span>
                  : absentees.length === 0 ? <span className="text-xs text-emerald-600">الكل سجّل ✓</span>
                  : absentees.map((e) => <span key={e.id} className="rounded-full bg-red-50 px-3 py-1 text-xs font-bold text-red-700">{e.name}</span>)}
              </div>
            ) : present > 0 && (
              <div className="mt-3 flex flex-wrap items-center gap-1.5">
                <span className="text-xs font-bold text-slate-500">👥 {attFilter === 'active' ? 'على الدوام هسه' : attFilter === 'left' ? 'انصرفوا' : 'الموظفون اليوم'}</span>
                {attendance.filter((a) => attFilter === 'all' || (attFilter === 'active' ? a.currentlyActive : !a.currentlyActive)).slice(0, 60).map((a) => (
                  <span key={a.employeeId}
                    className={`rounded-full px-3 py-1 text-xs font-bold ${a.currentlyActive ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'}`}
                    title={`أول دخول ${new Date(a.firstCheckIn).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })}`}>
                    {a.employee?.name || '—'}
                  </span>
                ))}
              </div>
            )}
          </>
        )}
      </Section>

      {!monitor && <NewEmployeeCurves />}

      {/* ما موجودة بتصميم (ع) — نزلت جوّا لحد ما يقرر تبقى لو تنشال */}
      <div className="grid gap-4 lg:grid-cols-2">
      <Section title="🏢 الشركة بالأرقام (عام)">
        {summary === undefined ? <Loading /> : summary === null ? <Missing /> : (
          <div className="grid grid-cols-2 gap-2 sm:grid-cols-2">
            <Stat icon="👥" tone="violet" label="الموظفين" hint="موظف وموظفة — يفتح «إدارة الموظفين»" to="/staff-management-desk" value={fmt(summary.employeeCount)} />
            <Stat icon="🤝" tone="blue" label="العملاء" hint="عميل بالنظام" to="/customers" value={fmt(summary.customerCount)} />
            <Stat icon="📦" tone="green" label="الحجوزات" hint="حجز بالنظام" to="/bookings" value={fmt(summary.bookingCount)} />
            <Stat icon="📡" tone="amber" label="أجهزة الجي بي اس" hint="جهاز مسجّل" to="/gps/devices" value={fmt(summary.gpsDeviceCount)} />
          </div>
        )}
      </Section>

        {pulse && (
          <Section title="🕐 تنتظر تنسيق">
            <Stat icon="🕐" label="حجوزات تنتظر تنسيق" to={drillTo('needs_coord')} value={fmt(pulse.needsCoordination)} tone={pulse.needsCoordination ? 'amber' : 'slate'}
              hint="مثبّتة وناقصها موعد أو كادر — اضغط تشوف كل وحدة وين واكفة وليش" />
          </Section>
        )}
      </div>
    </div>
  )
}

// ═══ لوحة الإجراءات ═══
function ActionsBoard() {
  return (
    <div className="space-y-4">
      {/* جدول الإجراءات بتصميم (ع) 10-04 — بدل مربعات العدّادات */}
      <ActionsTable />

      <div className="grid gap-4 lg:grid-cols-2">
        <StockForecastSection />
        <Section title="🔁 ماتركس — مقترح استبداله" to="/it-stats" linkLabel="إحصائيات الـIT ←">
          <ReplacementSuggestionsPanel part="any" compact />
        </Section>
      </div>

      <section className="rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
        <h3 className="mb-3 text-base font-extrabold text-[#0f2040]">⚡ إجراءات سريعة</h3>
        <div className="flex flex-wrap gap-2">
          {[
            { to: '/extra-tasks', label: '➕ أعطي مهمة إضافية' },
            { to: '/sales', label: '📅 حجز جديد' },
            { to: '/coordinator', label: '🗂️ تنسيق الحجوزات' },
            { to: '/weekly-report', label: '📊 التقرير الأسبوعي' },
            { to: '/finance', label: '🧾 تدقيق الحسابات' },
            { to: '/complaints', label: '📣 الشكاوى' },
            { to: '/staff-management-desk', label: '👥 إدارة الموظفين' },
            { to: '/access', label: '🔐 الصلاحيات والوصول' },
          ].map((a) => (
            <Link key={a.to} to={a.to}
              className="rounded-xl bg-brand-50 px-3 py-2 text-sm font-bold text-brand-700 transition-colors hover:bg-brand-100">
              {a.label}
            </Link>
          ))}
        </div>
      </section>
    </div>
  )
}

export default function AdminHome({ name }: { name?: string }) {
  const [params, setParams] = useSearchParams()
  const raw = params.get('board')
  // «مكتب المدير»: بلا اختيار ينفتح على الثلاثة سوية، وواحنا نختار.
  const board: Board | 'desk' = raw === 'actions' || raw === 'matrix' || raw === 'follow' || raw === 'chat' || raw === 'missions' ? raw : 'desk'
  const tab = (b: Board | 'desk', label: string) => (
    <button type="button" onClick={() => setParams({ board: b })}
      className={`rounded-xl px-4 py-2 text-sm font-extrabold transition-colors ${board === b ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600 ring-1 ring-slate-200 hover:bg-slate-50'}`}>
      {label}
    </button>
  )
  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-extrabold text-[#0f2040]">{name ? `هلا يا ${name}` : 'مدير النظام'}</h2>
          <p className="text-sm text-slate-500">كل شي بالشركة بمكان واحد — شوف وتحرك.</p>
        </div>
        <div className="flex flex-wrap gap-2">
          {tab('desk', '🏛️ مكتب المدير')}
          {tab('follow', '📊 المتابعة')}
          {tab('actions', '⚡ الإجراءات')}
          {tab('matrix', '👁️ ماتركس')}
          {tab('chat', '💬 دردشة ماتركس')}
          {tab('missions', '🎯 مهام ماتركس')}
        </div>
      </div>
      {board === 'desk' ? (
        <div className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {([
            ['follow', '📊', 'المتابعة', 'الشركة بالأرقام، ومتابعة اليوم، والحسابات، والدوام، وترتيب الأداء.', 'from-sky-500 to-indigo-600'],
            ['actions', '⚡', 'الإجراءات', 'كل شي ينتظر قرارك أو متابعة — بجدول واحد ويودّيك لشاشته.', 'from-amber-500 to-orange-600'],
            ['matrix', '👁️', 'ماتركس', 'مركز قيادة ماتركس: شنو اكتشف، وعيون الموظفين، والتوقعات.', 'from-violet-500 to-fuchsia-600'],
            ['chat', '💬', 'دردشة ماتركس', 'سولف ويا ماتركس: يفتش بالنظام بنفسه ويجاوبك مثل أي ذكاء اصطناعي.', 'from-fuchsia-500 to-pink-600'],
            ['missions', '🎯', 'مهام ماتركس', 'كلّف ماتركس يراقب موظف أيام: يومية كل ليلة وخلاصة بالنهاية.', 'from-emerald-500 to-teal-600'],
          ] as const).map(([b, icon, title, desc, grad]) => (
            <button key={b} type="button" onClick={() => setParams({ board: b })}
              className="group rounded-3xl border border-slate-200 bg-white p-6 text-right shadow-sm transition hover:-translate-y-1 hover:shadow-lg">
              <span className={`grid h-16 w-16 place-items-center rounded-2xl bg-gradient-to-br ${grad} text-3xl text-white shadow-md`}>{icon}</span>
              <h3 className="mt-4 text-xl font-extrabold text-[#0f2040]">{title}</h3>
              <p className="mt-1 text-sm text-slate-500">{desc}</p>
              <p className="mt-4 text-sm font-bold text-brand-600 group-hover:underline">افتح ←</p>
            </button>
          ))}
        </div>
      ) : board === 'follow' ? <FollowBoard /> : board === 'actions' ? <ActionsBoard /> : board === 'chat' ? <MatrixChat /> : board === 'missions' ? <MatrixMissions /> : <MatrixCommandCenter />}
    </div>
  )
}
