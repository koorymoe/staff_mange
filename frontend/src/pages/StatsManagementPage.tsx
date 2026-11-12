import { useEffect, useState, type ReactNode } from 'react'
import { api, type DailyStats, type WeeklyStats, type Stats, type InternalWorksReport } from '../api'
import MonthCompare from '../components/MonthCompare'
import BookingCodeChip from '../components/BookingCodeChip'
import ProjectStatsBoard from '../components/ProjectStatsBoard'
import RoleSplit, { MiniMetric, SummaryCard } from '../components/RoleSplit'
import { numIQ as fmt } from '../utils/format'
import PctWhy from '../ui/PctWhy'

const PRIMARY = '#1a237e'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const PRIMARY_TEXT = 'var(--brand-ink)'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const GOLD_TEXT = 'var(--gold-ink)'



function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function StatCard({ label, value }: { label: string; value: ReactNode }) {
  return <SummaryCard label={label} value={value} />
}

function DailyTab() {
  const [date, setDate] = useState(todayStr())
  const [stats, setStats] = useState<DailyStats | null>(null)
  const [err, setErr] = useState('')
  useEffect(() => {
    // ⚠️ جلب آمن: تغيير التاريخ بسرعة ما يخلي رد قديم يكتب فوگ الجديد،
    // والفشل يطلع رسالة بدل «جاري التحميل» للأبد.
    let alive = true
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setStats(null); setErr('')
    api.getDailyStats(date).then((r) => { if (alive) setStats(r) })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذّر جلب الإحصائيات') })
    return () => { alive = false }
  }, [date])

  return (
    <div>
      <div className="bg-sf-card border border-bd-line rounded-[10px] py-3 px-4 mb-4 flex items-center gap-3">
        <label className="text-[13px] text-t-muted">التاريخ</label>
        <input
          type="date"
          value={date}
          onChange={(e) => e.target.value && setDate(e.target.value)}
          className="py-2 px-2.5 border border-bd-line rounded-lg text-[14px]"
        />
        {date !== todayStr() && (
          <button
            onClick={() => setDate(todayStr())}
            style={{ padding: '8px 14px', border: `1px solid ${PRIMARY}`, background: 'var(--sf-card)', color: PRIMARY_TEXT, borderRadius: '8px', fontSize: '13px', fontWeight: 'bold', cursor: 'pointer' }}
          >
            اليوم
          </button>
        )}
      </div>

      {err && <p className="text-t-danger text-center p-10">{err}</p>}
      {!stats && !err && <p className="text-t-faint text-center p-10">جاري التحميل...</p>}

      {stats && (
        <>
          <div className="mb-5 grid grid-cols-2 gap-3 md:grid-cols-4">
            <StatCard label="إجمالي الحجوزات" value={stats.totalBookings} />
            <StatCard label="حجوزات صباحية" value={stats.morningBookings} />
            <StatCard label="حجوزات مسائية" value={stats.eveningBookings} />
            <StatCard label="كادر طلع للحجوزات" value={stats.crewOutCount} />
            <StatCard label="سيارات استُخدمت" value={stats.vehiclesOutCount} />
            <StatCard label="إجمالي عدد الموظفين" value={stats.totalEmployeesCount} />
            <StatCard label="إجمالي المبيعات" value={`${fmt(stats.totalSalesAmount)} د.ع`} />
            <StatCard label="إجمالي الأرباح" value={`${fmt(stats.totalProfitAmount)} د.ع`} />
          </div>

          {stats.employees.length === 0 && <p className="rounded-xl bg-white p-8 text-center text-slate-400">لا يوجد نشاط بهذا اليوم</p>}
          <RoleSplit items={stats.employees}>
            {(list) => (
              <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-3">
                {list.map((e) => {
                  const pct = e.bookingsAssigned > 0 ? Math.round((e.bookingsCompleted / e.bookingsAssigned) * 100) : null
                  return (
                    <div key={e.employeeId} className="flex flex-col gap-2 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
                      <div className="flex items-center justify-between gap-2">
                        <p className="truncate font-extrabold text-slate-800">{e.employeeName}</p>
                        <span className={`shrink-0 rounded-full px-2.5 py-0.5 text-xs font-bold ${e.checkedIn ? 'bg-emerald-100 text-emerald-800' : 'bg-red-100 text-red-700'}`}>
                          {e.checkedIn ? '✔ سجّل حضور' : '✘ ما سجّل حضور'}
                        </span>
                      </div>
                      <div className="grid grid-cols-3 gap-1.5">
                        <MiniMetric label="ترحّلت له" value={e.bookingsAssigned} />
                        <MiniMetric label="نفّذ" value={e.bookingsCompleted} tone={e.bookingsCompleted > 0 ? 'good' : undefined} />
                        <MiniMetric label="ما نفّذ" value={e.bookingsAssigned - e.bookingsCompleted} tone={e.bookingsAssigned - e.bookingsCompleted > 0 ? 'bad' : undefined} />
                      </div>
                      {pct !== null && (
                        <div className="h-1.5 overflow-hidden rounded-full bg-slate-100" title={`نفّذ ${pct}٪`}>
                          <div className="h-full bg-emerald-500" style={{ width: `${pct}%` }} />
                        </div>
                      )}
                    </div>
                  )
                })}
              </div>
            )}
          </RoleSplit>
        </>
      )}
    </div>
  )
}

function defaultWeekRange() {
  const to = todayStr()
  const d = new Date()
  d.setDate(d.getDate() - 7)
  const from = `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
  return { from, to }
}


function WeeklyTab() {
  const initial = defaultWeekRange()
  const [from, setFrom] = useState(initial.from)
  const [to, setTo] = useState(initial.to)
  const [stats, setStats] = useState<WeeklyStats | null>(null)
  const [err, setErr] = useState('')

  useEffect(() => {
    if (!from || !to) return
    let alive = true
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setStats(null); setErr('')
    api.getWeeklyStats(from, to).then((r) => { if (alive) setStats(r) })
      .catch((e) => { if (alive) setErr(e instanceof Error ? e.message : 'تعذّر جلب الإحصائيات') })
    return () => { alive = false }
  }, [from, to])

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      <div className="bg-sf-card border border-bd-line rounded-[10px] py-3 px-4 flex items-center gap-3 flex-wrap">
        <label className="text-[13px] text-t-muted">من</label>
        <input type="date" value={from} max={to} onChange={(e) => e.target.value && setFrom(e.target.value)} className="py-2 px-2.5 border border-bd-line rounded-lg text-[14px]" />
        <label className="text-[13px] text-t-muted">إلى</label>
        <input type="date" value={to} min={from} onChange={(e) => e.target.value && setTo(e.target.value)} className="py-2 px-2.5 border border-bd-line rounded-lg text-[14px]" />
        <button
          onClick={() => { const r = defaultWeekRange(); setFrom(r.from); setTo(r.to) }}
          style={{ padding: '8px 14px', border: `1px solid ${PRIMARY}`, background: 'var(--sf-card)', color: PRIMARY_TEXT, borderRadius: '8px', fontSize: '13px', fontWeight: 'bold', cursor: 'pointer' }}
        >
          آخر 7 أيام
        </button>
      </div>

      {err && <p className="text-t-danger text-center p-10">{err}</p>}
      {!stats && !err && <p className="text-t-faint text-center p-10">جاري التحميل...</p>}

      {stats && (
        <>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-3">
            <StatCard label="مبيعات صباحية" value={`${fmt(stats.morningSalesAmount)} د.ع`} />
            <StatCard label="مبيعات مسائية" value={`${fmt(stats.eveningSalesAmount)} د.ع`} />
            <StatCard label="إجمالي حجم المبيعات" value={`${fmt(stats.totalSalesAmount)} د.ع`} />
          </div>

          <h3 className="font-bold text-slate-700">أداء كل موظف خلال المدى المحدد</h3>
          {stats.employees.length === 0 && <p className="rounded-xl bg-white p-8 text-center text-slate-400">لا يوجد نشاط بهذا المدى</p>}
          <RoleSplit items={stats.employees}>
            {(list) => (
              <div className="grid gap-3 lg:grid-cols-2">
                {[...list].sort((a, b) => b.kpiPoints - a.kpiPoints).map((r) => (
                  <div key={r.employeeId} className="flex flex-col gap-3 rounded-2xl border border-slate-200 bg-white p-4 shadow-sm">
                    <div className="flex items-start justify-between gap-3">
                      <p className="truncate text-base font-extrabold text-slate-800">{r.employeeName}</p>
                      {/* لون محايد: مقياس الشهر (١٠٠ نقطة) ما يصح على أسبوع. */}
                      <span className="shrink-0 rounded-xl bg-brand-50 px-3 py-1 text-center text-brand-800">
                        <span className="block text-lg font-extrabold leading-none tabular-nums">{r.kpiPoints}</span>
                        <span className="text-[10px] font-bold">نقاط KPI</span>
                      </span>
                    </div>
                    <div className="grid grid-cols-2 gap-1.5 sm:grid-cols-3">
                      <MiniMetric label="الحجوزات المكتملة" value={`${r.completedBookingsCount} من ${r.totalBookingsCount}`} />
                      <MiniMetric label="حجوزات الصيانة" value={r.maintenanceBookingsCount} />
                      <MiniMetric label="صيانات مجانية" value={r.freeMaintenanceCount} />
                      <MiniMetric label="المبيعات" value={r.salesCount} />
                      <MiniMetric label="سرعة العمل" value={r.workSpeedScore != null ? r.workSpeedScore.toFixed(2) : '—'} tone={r.workSpeedScore == null ? undefined : r.workSpeedScore >= 1 ? 'good' : 'bad'} hint="فوق ١ أسرع من المتوسط" />
                      <MiniMetric label="نظافة السيارة" value={r.vehicleCleanlinessScore != null ? `${r.vehicleCleanlinessScore.toFixed(2)} (${r.vehicleRatingsCount})` : '—'} />
                      <MiniMetric label="الشكاوى" value={r.complaintsCount} tone={r.complaintsCount > 0 ? 'bad' : undefined} />
                      <MiniMetric label="قيمة نقاط الكي بي اي" value={`${fmt(r.kpiPointsValue)} د.ع`} />
                      <MiniMetric label="العمولة" value={`${fmt(r.totalCommission)} د.ع`} />
                    </div>
                  </div>
                ))}
              </div>
            )}
          </RoleSplit>
        </>
      )}
    </div>
  )
}

// (ع) 10-13: إحصائيات المشاريع بالتصميم الجديد
function ProjectsTab() {
  return <ProjectStatsBoard />
}

function monthStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

/**
 * أكثر خدمة طلبها الزبائن — انتقلت لهنا من صفحة إحصائيات الموظفين،
 * لأن مكانها الصحيح إدارة الإحصائيات.
 */
type SvcRow = Stats['serviceBreakdown'][number]
function SvcBars({ rows, val, total }: { rows: SvcRow[]; val: (s: SvcRow) => number; total: number }) {
  const max = rows.length ? val(rows[0]) : 1
  return (
    <div>
      {rows.map((s, i) => (
        <div key={s.serviceId || i} className="mb-2.5" data-testid="svc-row">
          <div className="flex items-center gap-3">
            <span className="w-6 text-center text-[13px] font-bold" style={{ color: i < 3 ? GOLD_TEXT : 'var(--t-faint)' }}>{i === 0 ? '🥇' : i === 1 ? '🥈' : i === 2 ? '🥉' : i + 1}</span>
            <span style={{ fontSize: '13px', width: '140px', color: 'var(--t-body)' }}>{s.name}</span>
            <div style={{ flex: 1, height: '20px', background: 'var(--sf-sunken)', borderRadius: '10px', overflow: 'hidden' }}>
              <div style={{ height: '100%', width: `${(val(s) / (max || 1)) * 100}%`, background: PRIMARY, borderRadius: '10px' }} />
            </div>
            <span style={{ fontSize: '13px', fontWeight: 'bold', width: '44px', color: PRIMARY_TEXT }}>{val(s)}</span>
            <span className="w-12 text-[11px] text-t-faint">
              <PctWhy title={`حصة ${s.name}`} parts={{ numLabel: `طلب ${s.name}`, num: val(s), denLabel: 'طلب لكل الخدمات', den: total }}
                meaning={<>شكد من الطلبات كانت على هالخدمة بنفس الفترة.</>}
                source={<>الحجوزات المسجّلة (غير الملغية والمحذوفة) حسب خدمتها، بتاريخ تسجيل الحجز.</>} />
            </span>
          </div>
        </div>
      ))}
      {rows.length === 0 && <p className="text-t-faint text-[13px]">ماكو طلبات بهالفترة بعد.</p>}
    </div>
  )
}

function ServicesTab() {
  const [stats, setStats] = useState<Stats | null>(null)
  useEffect(() => { api.getStats().then(setStats).catch(() => setStats(null)) }, [])

  if (!stats) return <p className="text-t-faint text-center p-10">جاري التحميل...</p>
  // (ل) 11-12 دازاي: «كل خدمة شكد صار عليها طلب هذا الشهر وبالكلي — شنو أكثر الخدمات»
  const all = [...stats.serviceBreakdown].sort((a, b) => b.count - a.count)
  const month = [...stats.serviceBreakdown].filter((s) => s.monthCount > 0).sort((a, b) => b.monthCount - a.monthCount)
  const monthTotal = month.reduce((n, s) => n + s.monthCount, 0)
  const allTotal = all.reduce((n, s) => n + s.count, 0)
  const monthName = new Date().toLocaleDateString('ar-IQ', { month: 'long', year: 'numeric', timeZone: 'Asia/Baghdad' })
  return (
    <div className="grid gap-4 lg:grid-cols-2">
      <div className="bg-sf-card border border-bd-line rounded-xl p-5" data-testid="svc-month">
        <h3 style={{ color: PRIMARY_TEXT, margin: '0 0 4px 0' }}>📅 هذا الشهر — {monthName}</h3>
        <p className="text-t-faint text-[13px] mt-0 mb-4">كل خدمة شكد انطلبت من أول الشهر لهسه · المجموع <b>{monthTotal}</b> طلب{month[0] ? <> · الأكثر: <b style={{ color: GOLD_TEXT }}>{month[0].name}</b></> : null}</p>
        <SvcBars rows={month} val={(s) => s.monthCount} total={monthTotal} />
      </div>
      <div className="bg-sf-card border border-bd-line rounded-xl p-5" data-testid="svc-all">
        <h3 style={{ color: PRIMARY_TEXT, margin: '0 0 4px 0' }}>🗂️ الكلي — من أول النظام</h3>
        <p className="text-t-faint text-[13px] mt-0 mb-4">المجموع <b>{allTotal}</b> طلب{all[0] ? <> · الأكثر: <b style={{ color: GOLD_TEXT }}>{all[0].name}</b></> : null}</p>
        <SvcBars rows={all} val={(s) => s.count} total={allTotal} />
        <div className="mt-4 border-t border-bd-line pt-3 text-[11px] text-t-faint">
          {all.map((s, i) => (
            <p key={s.serviceId || i} className="mb-1">{s.name}: 📅 من {s.firstAt ? new Date(s.firstAt).toLocaleDateString('ar-IQ') : '—'} · 💰 دخل <b style={{ color: PRIMARY_TEXT }}>{fmt(s.revenue)}</b> · 📈 ربح <b style={{ color: GOLD_TEXT }}>{fmt(s.profit)}</b> د.ع</p>
          ))}
          <p className="mt-2">الربح = المستلم ناقص كلفة المواد بسعر الجملة (المواد بلا سعر جملة تنحسب صفر، فالربح هو الحد الأعلى).</p>
        </div>
      </div>
    </div>
  )
}

/** الأعمال المنجزة داخل الشركة خلال شهر — شنو انخلص جوه ومنو اشتغل */
function InternalWorksTab() {
  const [month, setMonth] = useState(monthStr())
  const [rep, setRep] = useState<InternalWorksReport | null>(null)
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setRep(null)
    api.getInternalWorks(month).then(setRep).catch(() => setRep(null))
  }, [month])

  const total = rep ? rep.inHouseCount + rep.onSiteCount : 0
  const share = rep && total > 0 ? Math.round((rep.inHouseCount / total) * 100) : 0

  return (
    <div>
      <div className="bg-sf-card border border-bd-line rounded-[10px] py-3 px-4 mb-4 flex items-center gap-3">
        <label className="text-[13px] text-t-muted">الشهر</label>
        <input type="month" value={month} onChange={(e) => e.target.value && setMonth(e.target.value)}
          className="py-2 px-2.5 border border-bd-line rounded-lg text-[14px]" />
      </div>

      {!rep && <p className="text-t-faint text-center p-10">جاري التحميل...</p>}

      {rep && (
        <>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '12px', marginBottom: '20px' }}>
            <StatCard label="أعمال انخلصت داخل الشركة" value={rep.inHouseCount} />
            <StatCard label="أعمال طلعت للزبون" value={rep.onSiteCount} />
            <StatCard label="نسبة الشغل الداخلي" value={<PctWhy title="نسبة الشغل الداخلي" value={share}
              parts={{ numLabel: 'شغل انخلص داخل الشركة', num: rep.inHouseCount, denLabel: 'حجز مكتمل بهذا الشهر (داخلي + عند الزبون)', den: total }}
              meaning={<>من كل الحجوزات الي خلصت بهذا الشهر، شكد منها انشغلت بالشركة بدل ما الفريق يطلع للزبون. تصعد كل ما يجيبون أغراض أكثر يتصلحن جوه.</>}
              source={<>جدول الحجوزات: الحجوزات المكتملة بهذا الشهر (حسب تاريخ الإنجاز)، والي مكان شغلها «داخل الشركة» مقابل الباقي.</>} />} />
            <StatCard label="مبالغ الشغل الداخلي" value={`${fmt(rep.inHouseAmount)} د.ع`} />
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px', marginBottom: '20px' }}>
            <div className="bg-sf-card border border-bd-line rounded-xl p-4">
              <h4 style={{ color: PRIMARY_TEXT, margin: '0 0 12px 0', fontSize: '15px' }}>شنو انشتغل جوه</h4>
              {rep.services.map((s) => (
                <div key={s.name} className="flex justify-between py-1.5 px-0 border-b border-[#f2f2f2] text-[13px]">
                  <span>{s.name}</span>
                  <span style={{ fontWeight: 'bold', color: PRIMARY_TEXT }}>{s.count} · {fmt(s.amount)} د.ع</span>
                </div>
              ))}
              {rep.services.length === 0 && <p className="text-t-faint text-[13px]">ماكو شغل داخلي بهذا الشهر</p>}
            </div>

            <div className="bg-sf-card border border-bd-line rounded-xl p-4">
              <h4 style={{ color: PRIMARY_TEXT, margin: '0 0 12px 0', fontSize: '15px' }}>منو اشتغل جوه</h4>
              {rep.crew.map((c) => (
                <div key={c.employeeName} className="flex justify-between py-1.5 px-0 border-b border-[#f2f2f2] text-[13px]">
                  <span>{c.employeeName}</span>
                  <span style={{ fontWeight: 'bold', color: GOLD_TEXT }}>{c.count} عمل</span>
                </div>
              ))}
              {rep.crew.length === 0 && <p className="text-t-faint text-[13px]">ماكو كادر مسجّل</p>}
            </div>
          </div>

          <div className="overflow-x-auto bg-sf-card rounded-xl border border-bd-line">
            <table className="w-full border-collapse">
              <thead>
                <tr>
                  {['الكود', 'الخدمة', 'تاريخ الإنجاز', 'المبلغ'].map((h) => (
                    <th key={h} style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rep.works.map((w) => (
                  <tr key={w.code} className="border-b border-bd-line">
                    <td className="py-[9px] px-3 text-[13px]"><BookingCodeChip code={w.code} /></td>
                    <td className="py-[9px] px-3 text-[13px]">{w.serviceName}</td>
                    <td className="py-[9px] px-3 text-[13px]">
                      {w.completedAt ? new Date(w.completedAt).toLocaleDateString('ar-IQ') : '—'}
                    </td>
                    <td className="py-[9px] px-3 text-[13px]">{fmt(w.amount)} د.ع</td>
                  </tr>
                ))}
                {rep.works.length === 0 && (
                  <tr><td colSpan={4} className="p-[30px] text-center text-t-faint">ماكو أعمال داخلية بهذا الشهر</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  )
}

export default function StatsManagementPage() {
  // ٧و (11-05): /stats و/employee-stats انشالن من القائمة — روابطهن تفتح هنا على تبويبهن
  const [tab, setTab] = useState<'daily' | 'weekly' | 'compare' | 'internal' | 'services' | 'projects'>('services')

  const tabs: { key: typeof tab; label: string }[] = [
    // (ل) 11-12 دازاي: تقييمات الموظفين ما تطلع بالإحصائيات — انتقلت لـ«تصنيف المدراء والموظفين»
    { key: 'services', label: '🛠️ الخدمات المطلوبة (الشهر والكلي)' },
    { key: 'daily', label: 'يومية' },
    { key: 'weekly', label: 'أسبوعية' },
    { key: 'compare', label: '📅 مقارنة الأشهر' },
    { key: 'internal', label: 'داخل الشركة' },
    { key: 'projects', label: 'المشاريع' },
  ]

  return (
    <div style={{ direction: 'rtl', fontFamily: "'Segoe UI', Tahoma, Arial, sans-serif" }}>
      <div style={{
        background: `linear-gradient(135deg, ${PRIMARY}, #283593)`,
        color: 'white', padding: '20px 30px', borderRadius: '12px', marginBottom: '20px',
      }}>
        <h1 className="m-0 text-[24px]">📊 الإحصائيات</h1>
        <span style={{ color: GOLD_TEXT, fontSize: '14px' }}>يومية، أسبوعية، شهرية، شغل داخل الشركة، أكثر خدمة مطلوبة، ومشاريع — حصراً لمدير النظام</span>
      </div>

      <div className="flex flex-wrap gap-2 mb-5">
        {tabs.map((t) => (
          <button
            key={t.key}
            onClick={() => setTab(t.key)}
            style={{
              padding: '10px 20px', borderRadius: '8px', border: 'none', cursor: 'pointer',
              fontWeight: 'bold', fontSize: '14px',
              background: tab === t.key ? PRIMARY : 'white',
              color: tab === t.key ? 'white' : PRIMARY,
              boxShadow: tab === t.key ? 'none' : '0 1px 4px rgba(0,0,0,0.1)',
            }}
          >
            {t.label}
          </button>
        ))}
      </div>

      {tab === 'daily' && <DailyTab />}
      {tab === 'weekly' && <WeeklyTab />}
      {tab === 'compare' && <MonthCompare />}
      {tab === 'internal' && <InternalWorksTab />}
      {tab === 'services' && <ServicesTab />}
      {tab === 'projects' && <ProjectsTab />}
    </div>
  )
}
