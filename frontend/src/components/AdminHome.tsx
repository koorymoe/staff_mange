import MatrixCommandCenter from './MatrixCommandCenter'
import { useEffect, useMemo, useState, type ReactNode } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import {
  api,
  type AiSignal,
  type EmployeeDailyAttendanceSummary,
  type FinanceSummary,
  type InternalWorksReport,
  type MonitorDeskCounts,
  type NewEmployeeCurve,
  type StockForecast,
  type TechnicianKpi,
  type TodayPulse,
} from '../api'
import TodayBoard from './TodayBoard'
import ReplacementSuggestionsPanel from './ReplacementSuggestionsPanel'

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

type Board = 'follow' | 'actions' | 'matrix'

const fmt = (n: number) => n.toLocaleString('en-US')
const money = (n: number) => `${fmt(Math.round(n))} د.ع`
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

function Stat({ label, value, hint, tone = 'slate' }: { label: string; value: ReactNode; hint?: ReactNode; tone?: 'slate' | 'green' | 'amber' | 'red' | 'blue' }) {
  const tones = {
    slate: 'text-slate-800', green: 'text-emerald-700', amber: 'text-amber-700', red: 'text-red-700', blue: 'text-brand-700',
  }
  return (
    <div className="rounded-xl bg-slate-50 px-3 py-2.5">
      <p className="text-xs text-slate-500">{label}</p>
      <p className={`mt-0.5 text-xl font-extrabold tabular-nums ${tones[tone]}`}>{value}</p>
      {hint && <p className="mt-0.5 text-[11px] text-slate-400">{hint}</p>}
    </div>
  )
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
        <ul className="divide-y divide-slate-100">
          {curves.map((c) => {
            const last = [...c.weeks].reverse().find((w) => w.complete)
            return (
              <li key={c.employeeId} className="flex items-center justify-between gap-3 py-2">
                <div className="min-w-0">
                  <p className="truncate font-bold text-slate-700">{c.name}{c.isLeader ? ' · ليدر' : ''}</p>
                  <p className="text-[11px] text-slate-400">
                    يومه {fmt(c.daysIn)}
                    {last && ` · آخر أسبوع: ${fmt(last.completed)} منجز، ${fmt(last.stops)} توقف، ${fmt(last.complaints)} شكوى${last.paperPct != null ? `، ورق ${last.paperPct}٪` : ''}`}
                  </p>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-brand-600" title="نقاط كل أسبوع"><Sparkline values={c.weeks.map((w) => w.score)} /></span>
                  <span className={`rounded-full px-2 py-0.5 text-[11px] font-bold ${CURVE_TONE[c.status]}`} title={c.statusNote}>{CURVE_LABEL[c.status]}</span>
                </div>
              </li>
            )
          })}
        </ul>
      )}
      <p className="mt-2 text-[11px] text-slate-400">نقاط الأسبوع = منجز − ٢×توقف − ٣×شكوى. «يحتاج متابعة» = ٣ أسابيع مكتملة بلا تحسّن — للمتابعة والتدريب، مو عقوبة.</p>
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
function FollowBoard() {
  const summary = useLoad(() => api.getDashboardSummary())
  const pulse = useLoad<TodayPulse>(() => api.getTodayPulse())
  const finance = useLoad<FinanceSummary>(() => api.getFinanceSummary())
  const internal = useLoad<InternalWorksReport>(() => api.getInternalWorks(monthKey()))
  const attendance = useLoad<EmployeeDailyAttendanceSummary[]>(() => api.getTodaySummary())
  const signals = useLoad<AiSignal[]>(() => api.getAiSignals())
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

  const present = attendance?.length ?? 0
  const activeNow = attendance?.filter((a) => a.currentlyActive).length ?? 0

  return (
    <div className="space-y-4">
      <Section title="🏢 الشركة بالأرقام">
        {summary === undefined ? <Loading /> : summary === null ? <Missing /> : (
          <div className="grid grid-cols-2 gap-2 sm:grid-cols-4">
            <Stat label="الموظفين" value={fmt(summary.employeeCount)} />
            <Stat label="الزبائن" value={fmt(summary.customerCount)} />
            <Stat label="الحجوزات" value={fmt(summary.bookingCount)} />
            <Stat label="أجهزة الجي بي اس" value={fmt(summary.gpsDeviceCount)} />
          </div>
        )}
      </Section>

      <Section title="⏱️ اليوم" to="/missions" linkLabel="تتبع المهام ←">
        {pulse === undefined ? <Loading /> : pulse === null ? <Missing /> : (
          <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-6">
            <Stat label="حجوزات اليوم" value={fmt(pulse.todayBookings)} hint={`أمس ${fmt(pulse.yesterdayBookings)}`} tone="blue" />
            <Stat label="كوادر بالميدان" value={fmt(pulse.crewInField)} tone="green" />
            <Stat label="مهام مفتوحة" value={fmt(pulse.openMissions)} />
            <Stat label="متأخرة +٢٤ ساعة" value={fmt(pulse.overdueMissions)} tone={pulse.overdueMissions ? 'red' : 'slate'} />
            <Stat label="تنتظر تنسيق" value={fmt(pulse.needsCoordination)} tone={pulse.needsCoordination ? 'amber' : 'slate'} />
            <Stat label="شكاوى جديدة" value={fmt(pulse.newComplaints)} tone={pulse.newComplaints ? 'red' : 'slate'} />
          </div>
        )}
      </Section>

      <div className="grid gap-4 lg:grid-cols-2">
        <Section title="💰 الحسابات" to="/finance">
          {finance === undefined ? <Loading /> : finance === null ? <Missing /> : (
            <div className="grid grid-cols-2 gap-2">
              <Stat label="المحصّل (حجوزات منجزة)" value={money(finance.totalCollected)} tone="green" />
              <Stat label="مدقّق" value={money(finance.verifiedAmount)} hint={`${fmt(finance.verifiedCount)} حجز`} />
              <Stat label="بانتظار التدقيق" value={money(finance.unverifiedAmount)} hint={`${fmt(finance.unverifiedCount)} حجز`} tone={finance.unverifiedCount ? 'amber' : 'slate'} />
              <Stat label="المصاريف" value={money(finance.totalExpenseValue)} hint={`${fmt(finance.pendingExpenses)} بانتظار الموافقة`} tone="red" />
              <Stat label="أنجزت اليوم" value={fmt(finance.todayCompleted)} hint={`انفتح اليوم ${fmt(finance.todayCreated)}`} />
              <Stat label="قيد التنفيذ" value={fmt(finance.inProgressCount)} hint={`مثبّتة ${fmt(finance.confirmedCount)} · معلّقة ${fmt(finance.pendingCount)}`} />
            </div>
          )}
        </Section>

        <Section title="🏭 أعمال داخل الشركة (هذا الشهر)" to="/stats-management">
          {internal === undefined ? <Loading /> : internal === null ? <Missing /> : (
            <>
              <div className="grid grid-cols-3 gap-2">
                <Stat label="داخل الشركة" value={fmt(internal.inHouseCount)} />
                <Stat label="بالموقع" value={fmt(internal.onSiteCount)} />
                <Stat label="المبلغ" value={money(internal.inHouseAmount)} tone="green" />
              </div>
              {internal.services.length > 0 && (
                <ul className="mt-3 space-y-1 text-sm">
                  {internal.services.slice(0, 5).map((s) => (
                    <li key={s.name} className="flex justify-between gap-2 text-slate-600">
                      <span>{s.name}</span>
                      <span className="tabular-nums text-slate-500">{fmt(s.count)} · {money(s.amount)}</span>
                    </li>
                  ))}
                </ul>
              )}
            </>
          )}
        </Section>
      </div>

      <div className="grid gap-4 lg:grid-cols-2">
        <Section title="🧠 ماتركس — تقييمات الموظفين (آخر ٧ أيام)" to="/ai-insights" linkLabel="كل المؤشرات ←">
          {signals === undefined ? <Loading /> : signals === null ? <Missing /> : matrix.length === 0 ? (
            <p className="text-sm text-slate-500">ماكو أحكام على موظفين هالأسبوع ✅</p>
          ) : (
            <ul className="divide-y divide-slate-100">
              {matrix.slice(0, 8).map((m) => (
                <li key={m.id} className="py-2">
                  <div className="flex items-center justify-between gap-2">
                    <span className="font-bold text-slate-700">{m.name}</span>
                    <span className="flex items-center gap-1.5">
                      <span className="text-xs text-slate-400">{fmt(m.count)} حكم</span>
                      <span className={`rounded-full px-2 py-0.5 text-[11px] font-bold ${SEVERITY[m.top]?.cls || ''}`}>{SEVERITY[m.top]?.label || m.top}</span>
                    </span>
                  </div>
                  <p className="mt-0.5 truncate text-xs text-slate-500" title={m.headline}>{m.headline}</p>
                </li>
              ))}
            </ul>
          )}
          <p className="mt-2 text-[11px] text-slate-400">الأحكام توجيه ومراجعة — ماكو غرامة تنطلع منها تلقائياً.</p>
        </Section>

        <Section title="🏆 ترتيب الأداء (هذا الشهر)" to="/kpi">
          {kpi === undefined ? <Loading /> : kpi === null ? <Missing /> : kpi.length === 0 ? (
            <p className="text-sm text-slate-500">ماكو نقاط محسوبة بعد.</p>
          ) : (
            <ol className="space-y-1.5 text-sm">
              {[...kpi].sort((a, b) => b.totalPoints - a.totalPoints).slice(0, 6).map((k, i) => (
                <li key={k.employeeId} className="flex items-center justify-between gap-2">
                  <span className="text-slate-700"><b className="text-slate-400">{i + 1}.</b> {k.employeeName}</span>
                  <span className="tabular-nums font-bold text-brand-700">{fmt(k.totalPoints)}</span>
                </li>
              ))}
            </ol>
          )}
        </Section>
      </div>

      <Section title="🕒 سجل الدوام اليوم" to="/attendance" linkLabel="سجل الدوام ←">
        {attendance === undefined ? <Loading /> : attendance === null ? <Missing /> : (
          <>
            <div className="grid grid-cols-2 gap-2 sm:grid-cols-3">
              <Stat label="سجّلوا حضور اليوم" value={fmt(present)} />
              <Stat label="بالدوام هسه" value={fmt(activeNow)} tone="green" />
              <Stat label="طلعوا" value={fmt(present - activeNow)} />
            </div>
            {present > 0 && (
              <div className="mt-3 flex flex-wrap gap-1.5">
                {attendance.slice(0, 40).map((a) => (
                  <span key={a.employeeId}
                    className={`rounded-full px-2 py-0.5 text-xs ${a.currentlyActive ? 'bg-emerald-50 text-emerald-700' : 'bg-slate-100 text-slate-500'}`}
                    title={`أول دخول ${new Date(a.firstCheckIn).toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })}`}>
                    {a.employee?.name || '—'}
                  </span>
                ))}
              </div>
            )}
          </>
        )}
      </Section>

      <NewEmployeeCurves />

      <TodayBoard finance={finance ?? null} />
    </div>
  )
}

// ═══ لوحة الإجراءات ═══
function ActionTile({ to, icon, label, count, hint }: { to: string; icon: string; label: string; count: number | null | undefined; hint: string }) {
  const hot = typeof count === 'number' && count > 0
  return (
    <Link to={to}
      className={`flex items-start justify-between gap-3 rounded-2xl border p-4 transition-colors ${hot ? 'border-amber-300 bg-amber-50 hover:bg-amber-100' : 'border-slate-200 bg-white hover:bg-slate-50'}`}>
      <div>
        <p className="font-extrabold text-[#0f2040]">{icon} {label}</p>
        <p className="mt-0.5 text-xs text-slate-500">{hint}</p>
      </div>
      <span className={`min-w-10 rounded-xl px-2.5 py-1 text-center text-lg font-extrabold tabular-nums ${hot ? 'bg-amber-500 text-white' : 'bg-slate-100 text-slate-500'}`}>
        {count === undefined ? '…' : count === null ? '—' : fmt(count)}
      </span>
    </Link>
  )
}

function ActionsBoard() {
  const leaves = useLoad(() => api.getLeavePendingCount())
  const staff = useLoad(() => api.getStaffRequests())
  const deletes = useLoad(() => api.getBookingDeleteRequestCounts())
  const desk = useLoad<MonitorDeskCounts>(() => api.getMonitorDeskCounts())
  const dups = useLoad(() => api.getDuplicateCandidates(undefined, 'PENDING'))
  const achievements = useLoad(() => api.getAchievements({ limit: 200 }))
  const matrix = useLoad(() => api.getMatrixDecisions())

  const n = <T,>(v: T | null | undefined, f: (x: T) => number) => (v === undefined ? undefined : v === null ? null : f(v))

  return (
    <div className="space-y-4">
      <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-3">
        <ActionTile to="/matrix/decisions" icon="🤖" label="صندوق قرارات ماتركس" hint="شنو سوّى لحاله وشنو ينتظرك" count={n(matrix, (x) => x.pending.length + x.unstaffed.length + (x.proposals ?? 0))} />
        <ActionTile to="/leaves" icon="🌴" label="طلبات الإجازات" hint="تنتظر موافقتك" count={n(leaves, (x) => x.count)} />
        <ActionTile to="/staff-requests" icon="👷" label="طلبات الكادر" hint="المشاريع تطلب كادر" count={n(staff, (x) => x.filter((r) => r.status === 'PENDING').length)} />
        <ActionTile to="/booking-delete-requests" icon="🗑️" label="طلبات حذف الحجوزات" hint="تنتظر قرار" count={n(deletes, (x) => x.awaitingReview)} />
        <ActionTile to="/monitor" icon="🧠" label="صندوق المراقب" hint="بضمنه أحكام ماتركس" count={n(desk, (x) => x.inbox)} />
        <ActionTile to="/duplicate-review" icon="👯" label="تكرارات مشتبه بيها" hint="حجوزات وزبائن" count={n(dups, (x) => x.length)} />
        <ActionTile to="/achievements" icon="📋" label="إنجازات بلا تقييم" hint="تقارير الموظفين اليومية" count={n(achievements, (x) => x.filter((a) => a.reviewStatus === 'PENDING').length)} />
      </div>

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
            { to: '/permissions', label: '🔐 الصلاحيات' },
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
  const board: Board = raw === 'actions' || raw === 'matrix' ? raw : 'follow'
  const tab = (b: Board, label: string) => (
    <button type="button" onClick={() => setParams(b === 'follow' ? {} : { board: b })}
      className={`rounded-xl px-4 py-2 text-sm font-extrabold transition-colors ${board === b ? 'bg-[#0f2040] text-white' : 'bg-white text-slate-600 ring-1 ring-slate-200 hover:bg-slate-50'}`}>
      {label}
    </button>
  )
  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h2 className="text-2xl font-extrabold text-[#0f2040]">{name ? `هلا ${name}` : 'مدير النظام'}</h2>
          <p className="text-sm text-slate-500">كل شي بالشركة بمكان واحد — شوف وتحرك.</p>
        </div>
        <div className="flex gap-2">
          {tab('follow', '📊 المتابعة')}
          {tab('actions', '⚡ الإجراءات')}
          {tab('matrix', '👁️ ماتركس')}
        </div>
      </div>
      {board === 'follow' ? <FollowBoard /> : board === 'actions' ? <ActionsBoard /> : <MatrixCommandCenter />}
    </div>
  )
}
