import { useCallback, useEffect, useMemo, useState } from 'react'
import { api, type EmployeeMonthlyStats } from '../api'
import KpiBreakdownChart from '../components/KpiBreakdownChart'
import PerformanceCurveModal from '../components/PerformanceCurveModal'
import { roleChipColor, roleGradient, roleLabel } from '../roleLabels'


const fmt = (n: number) => n.toLocaleString('en-IQ')

function getCurrentMonth(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

// صفحة إحصائيات الموظفين الشهرية — تجمع نقاط الكي بي اي (نفس الآلية الموجودة
// أصلاً)، سرعة العمل (نسبة زمن الموظف للمتوسط العام بنفس المنظومة)،
// نظافة السيارة (من تقييم السائقين بعد المهمة الموجود أصلاً)، الشكاوى، عدد
// المبيعات، الحجوزات المكتملة، ومجموع العمولة الشهرية — OWNER/ADMIN فقط،
// مقيّدة بـRequireAdmin بنفس نمط بقية الصفحات الحساسة (permissions،
// service-managers).
export default function EmployeeMonthlyStatsPage() {
  const [month, setMonth] = useState(getCurrentMonth())
  const [rows, setRows] = useState<EmployeeMonthlyStats[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [exporting, setExporting] = useState(false)
  const [kpiFor, setKpiFor] = useState<{ id: string; name: string } | null>(null)
  const [curveFor, setCurveFor] = useState<{ id: string; name: string } | null>(null)

  const load = useCallback(() => {
    api.getEmployeeMonthlyStats(month)
      .then(setRows)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [month])

  useEffect(() => { load() }, [load])

  const handleExport = async () => {
    setExporting(true)
    try {
      await api.exportEmployeeMonthlyStats(month)
    } catch (err) {
      alert(err instanceof Error ? err.message : 'تعذر تصدير الملف')
    } finally {
      setExporting(false)
    }
  }

  // ═══ قرار (ع) 10-07: «كل دور ينفصل عن الثاني» و«مرتب مو مثل الإكسل» ═══
  // تبويب لكل دور، ملخص الدور فوگ، وبطاقة مرتبة لكل موظف — نفس الأرقام
  // كلها، بس كل رقم بمجموعته (الشغل · الجودة · الفلوس · النقاط).
  const roles = useMemo(() => {
    const m = new Map<string, number>()
    rows.forEach((r) => m.set(r.role, (m.get(r.role) ?? 0) + 1))
    return [...m.entries()].sort((a, b) => b[1] - a[1]).map(([role, n]) => ({ role, n }))
  }, [rows])
  const [pickedRole, setPickedRole] = useState('')
  const role = roles.some((r) => r.role === pickedRole) ? pickedRole : (roles[0]?.role ?? '')
  const shown = useMemo(() => rows.filter((r) => r.role === role)
    .sort((a, b) => b.smartKpiPoints - a.smartKpiPoints), [rows, role])

  const sum = (f: (r: EmployeeMonthlyStats) => number) => shown.reduce((t, r) => t + f(r), 0)
  // ⚠️ الرواتب الفاضية تنطرح من المجموع **ومن المقام سوا** — لو جمعناها
  // صفراً تطلع نسبة تغطية كاذبة عالية.
  const salarySum = sum((r) => r.salary ?? 0)
  const revenueOfSalaried = sum((r) => (r.salary ? r.revenueBrought : 0))
  const coverage = salarySum > 0 ? (revenueOfSalaried / salarySum) * 100 : null
  const avgKpi = shown.length ? Math.round(sum((r) => r.smartKpiPoints) / shown.length) : 0

  return (
    <div dir="rtl" className="space-y-4">
      <div className="flex flex-wrap items-end justify-between gap-3 rounded-2xl bg-gradient-to-l from-[#1a237e] to-[#283593] p-5 text-white shadow">
        <div>
          <h1 className="text-2xl font-extrabold">إحصائيات الموظفين الشهرية</h1>
          <p className="mt-1 text-sm text-white/80">كل دور لحاله — الشغل، الجودة، الفلوس، ونقاط الكي بي اي لكل موظف</p>
        </div>
        <div className="flex items-end gap-2">
          <label className="text-xs text-white/80">
            <span className="mb-1 block">الشهر</span>
            <input type="month" value={month} onChange={(e) => { setLoading(true); setMonth(e.target.value) }}
              className="rounded-lg border-0 bg-white px-3 py-2 text-sm font-bold text-[#1a237e] [color-scheme:light]" />
          </label>
          <button type="button" onClick={handleExport} disabled={exporting || loading}
            className="rounded-lg bg-[#c8a45a] px-4 py-2 text-sm font-bold text-[#1a237e] disabled:opacity-60">
            {exporting ? 'جاري التصدير…' : '⬇️ تنزيل إكسل'}
          </button>
        </div>
      </div>

      {loading && <p className="py-10 text-center text-slate-400">جاري التحميل…</p>}
      {error && <p className="rounded-xl border border-red-200 bg-red-50 p-4 text-red-700">تعذر الاتصال بالخادم: {error}</p>}
      {!loading && !error && rows.length === 0 && <p className="rounded-xl bg-white p-10 text-center text-slate-400">لا توجد بيانات لهذا الشهر</p>}

      {!loading && !error && rows.length > 0 && (
        <>
          {/* تبويب لكل دور */}
          <div className="flex flex-wrap gap-2">
            {roles.map(({ role: r, n }) => {
              const on = r === role
              const c = roleChipColor(r)
              return (
                <button key={r} type="button" onClick={() => setPickedRole(r)}
                  className={`flex items-center gap-2 rounded-xl px-4 py-2 text-sm font-bold shadow-sm transition ${on ? `bg-gradient-to-l ${roleGradient(r)} text-white` : `${c.bg} ${c.text} hover:brightness-95`}`}>
                  {roleLabel(r)}
                  <span className={`rounded-full px-2 text-xs tabular-nums ${on ? 'bg-white/25' : 'bg-white'}`}>{n}</span>
                </button>
              )
            })}
          </div>

          {/* ملخص الدور */}
          <div className="grid grid-cols-2 gap-3 md:grid-cols-5">
            <Summary label="عدد الموظفين" value={String(shown.length)} />
            <Summary label="متوسط نقاط الكي بي اي" value={String(avgKpi)} tone={avgKpi >= 100 ? 'good' : avgKpi >= 50 ? 'mid' : 'bad'} />
            <Summary label="المبالغ الي دخّلوها" value={`${fmt(sum((r) => r.revenueBrought))} د.ع`} />
            <Summary label="تغطية رواتبهم" value={coverage === null ? '—' : `${coverage.toFixed(0)}%`} tone={coverage === null ? undefined : coverage >= 100 ? 'good' : 'bad'} />
            <Summary label="العمولات" value={`${fmt(sum((r) => r.totalCommission))} د.ع`} />
          </div>

          {/* بطاقة لكل موظف */}
          <div className="grid gap-3 lg:grid-cols-2">
            {shown.map((r) => (
              <EmployeeCard key={r.employeeId} r={r}
                onKpi={() => setKpiFor({ id: r.employeeId, name: r.employeeName })}
                onCurve={() => setCurveFor({ id: r.employeeId, name: r.employeeName })} />
            ))}
          </div>
        </>
      )}

      {kpiFor && (
        <KpiBreakdownChart employeeId={kpiFor.id} employeeName={kpiFor.name} month={month} onClose={() => setKpiFor(null)} />
      )}
      {curveFor && (
        <PerformanceCurveModal employeeId={curveFor.id} employeeName={curveFor.name} onClose={() => setCurveFor(null)} />
      )}
    </div>
  )
}

const TONE: Record<string, string> = { good: 'text-emerald-700', mid: 'text-amber-700', bad: 'text-red-700' }

function Summary({ label, value, tone }: { label: string; value: string; tone?: 'good' | 'mid' | 'bad' }) {
  return (
    <div className="rounded-xl border border-slate-200 bg-white p-3 shadow-sm">
      <p className="text-xs text-slate-500">{label}</p>
      <p className={`mt-1 text-lg font-extrabold tabular-nums ${tone ? TONE[tone] : 'text-slate-800'}`}>{value}</p>
    </div>
  )
}

function Metric({ label, value, hint, tone }: { label: string; value: React.ReactNode; hint?: string; tone?: 'good' | 'mid' | 'bad' }) {
  return (
    <div className="rounded-lg bg-slate-50 px-3 py-2" title={hint}>
      <p className="text-[11px] text-slate-500">{label}</p>
      <p className={`text-sm font-bold tabular-nums ${tone ? TONE[tone] : 'text-slate-800'}`}>{value}</p>
    </div>
  )
}

function Group({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <div>
      <p className="mb-1.5 text-[11px] font-bold tracking-wide text-slate-400">{title}</p>
      <div className="grid grid-cols-2 gap-1.5 sm:grid-cols-3">{children}</div>
    </div>
  )
}

function EmployeeCard({ r, onKpi, onCurve }: { r: EmployeeMonthlyStats; onKpi: () => void; onCurve: () => void }) {
  const kpiTone = r.smartKpiPoints >= 100 ? 'bg-emerald-100 text-emerald-800' : r.smartKpiPoints >= 50 ? 'bg-amber-100 text-amber-800' : 'bg-red-100 text-red-800'
  const cov = r.salaryCoverage
  return (
    <div className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate text-base font-extrabold text-slate-800">{r.employeeName}</p>
          <p className="text-xs text-slate-500">{roleLabel(r.role)} · يعرف {r.servicesKnownCount} خدمة</p>
        </div>
        <button type="button" onClick={onKpi} title="اضغط تشوف النقاط مفصّلة"
          className={`shrink-0 rounded-xl px-3 py-1.5 text-center ${kpiTone}`}>
          <span className="block text-xl font-extrabold leading-none tabular-nums">{r.smartKpiPoints}</span>
          <span className="text-[10px] font-bold">نقاط KPI 📊</span>
        </button>
      </div>

      <Group title="الشغل">
        <Metric label="الحجوزات المكتملة" value={`${r.completedBookingsCount} من ${r.totalBookingsCount}`} />
        <Metric label="حجوزات الصيانة" value={r.maintenanceBookingsCount} />
        <Metric label="صيانات مجانية" value={r.freeMaintenanceCount} />
        <Metric label="أعمال داخل الشركة" value={r.inHouseWorksCount ?? 0} hint={r.inHouseWorkTypes?.join('، ')} />
        <Metric label="المبيعات" value={r.salesCount} />
      </Group>

      <Group title="الجودة">
        <Metric label="سرعة العمل" hint="فوق ١ أسرع من المتوسط، تحت ١ أبطأ — وبين القوسين عدد العيّنات"
          value={r.workSpeedScore == null ? '—' : <>{r.workSpeedScore >= 1 ? '⬆︎' : '⬇︎'} {r.workSpeedScore.toFixed(2)} <span className="text-[11px] font-normal text-slate-400">({r.workSpeedSamples})</span></>}
          tone={r.workSpeedScore == null ? undefined : r.workSpeedScore >= 1 ? 'good' : 'bad'} />
        <Metric label="نظافة السيارة" value={r.vehicleCleanlinessScore == null ? '—' : `${r.vehicleCleanlinessScore.toFixed(2)} (${r.vehicleRatingsCount})`} />
        <Metric label="الشكاوى" value={r.complaintsCount} tone={r.complaintsCount > 0 ? 'bad' : undefined} />
      </Group>

      <Group title="الفلوس">
        <Metric label="المبالغ الي دخّلها" value={`${fmt(r.revenueBrought)} د.ع`} hint="صافي فواتير الليدر الي رفعها هو بهذا الشهر" />
        <Metric label="راتبه" value={r.salary === null ? '—' : `${fmt(r.salary)} د.ع`} />
        <Metric label="تغطية الراتب" value={cov === null ? '—' : `${cov.toFixed(0)}%`} hint="الإيراد ÷ الراتب — والمطلوب يعبر ١٠٠٪" tone={cov === null ? undefined : cov >= 100 ? 'good' : 'bad'} />
        <Metric label="العمولة" value={`${fmt(r.totalCommission)} د.ع`} />
        <Metric label="تقييم يدوي" value={r.kpiPoints > 0 ? `+${r.kpiPoints}` : r.kpiPoints} tone={r.kpiPoints > 0 ? 'good' : r.kpiPoints < 0 ? 'bad' : undefined} />
        <Metric label="قيمة النقاط اليدوية" value={`${fmt(r.kpiPointsValue)} د.ع`} />
      </Group>

      {cov !== null && (
        <div className="h-1.5 overflow-hidden rounded-full bg-slate-100" title="تغطية الراتب">
          <div className={`h-full ${cov >= 100 ? 'bg-emerald-500' : 'bg-red-400'}`} style={{ width: `${Math.min(100, cov)}%` }} />
        </div>
      )}

      <button type="button" onClick={onCurve}
        className="self-start rounded-lg border border-brand-200 bg-brand-50 px-3 py-1.5 text-xs font-bold text-brand-700">
        📈 منحنى الأداء
      </button>
    </div>
  )
}
