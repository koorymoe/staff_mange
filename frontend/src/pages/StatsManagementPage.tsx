import { useEffect, useState } from 'react'
import { api, type DailyStats, type WeeklyStats, type ProjectStageStats, type Stats, type InternalWorksReport } from '../api'
import EmployeeMonthlyStatsPage from './EmployeeMonthlyStatsPage'
import BookingCodeChip from '../components/BookingCodeChip'
import RoleSplit, { MiniMetric, SummaryCard } from '../components/RoleSplit'

const PRIMARY = '#1a237e'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const PRIMARY_TEXT = 'var(--brand-ink)'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const GOLD_TEXT = 'var(--gold-ink)'

const fmt = (n: number) => n.toLocaleString('en-IQ')


function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function StatCard({ label, value }: { label: string; value: string | number }) {
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
      <div style={{ background: 'var(--sf-card)', border: '1px solid var(--bd-line)', borderRadius: '10px', padding: '12px 16px', marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '12px' }}>
        <label style={{ fontSize: '13px', color: 'var(--t-muted)' }}>التاريخ</label>
        <input
          type="date"
          value={date}
          onChange={(e) => e.target.value && setDate(e.target.value)}
          style={{ padding: '8px 10px', border: '1px solid var(--bd-line)', borderRadius: '8px', fontSize: '14px' }}
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

      {err && <p style={{ color: 'var(--t-danger)', textAlign: 'center', padding: '40px' }}>{err}</p>}
      {!stats && !err && <p style={{ color: 'var(--t-faint)', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>}

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
      <div style={{ background: 'var(--sf-card)', border: '1px solid var(--bd-line)', borderRadius: '10px', padding: '12px 16px', display: 'flex', alignItems: 'center', gap: '12px', flexWrap: 'wrap' }}>
        <label style={{ fontSize: '13px', color: 'var(--t-muted)' }}>من</label>
        <input type="date" value={from} max={to} onChange={(e) => e.target.value && setFrom(e.target.value)} style={{ padding: '8px 10px', border: '1px solid var(--bd-line)', borderRadius: '8px', fontSize: '14px' }} />
        <label style={{ fontSize: '13px', color: 'var(--t-muted)' }}>إلى</label>
        <input type="date" value={to} min={from} onChange={(e) => e.target.value && setTo(e.target.value)} style={{ padding: '8px 10px', border: '1px solid var(--bd-line)', borderRadius: '8px', fontSize: '14px' }} />
        <button
          onClick={() => { const r = defaultWeekRange(); setFrom(r.from); setTo(r.to) }}
          style={{ padding: '8px 14px', border: `1px solid ${PRIMARY}`, background: 'var(--sf-card)', color: PRIMARY_TEXT, borderRadius: '8px', fontSize: '13px', fontWeight: 'bold', cursor: 'pointer' }}
        >
          آخر 7 أيام
        </button>
      </div>

      {err && <p style={{ color: 'var(--t-danger)', textAlign: 'center', padding: '40px' }}>{err}</p>}
      {!stats && !err && <p style={{ color: 'var(--t-faint)', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>}

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

function ProjectsTab() {
  const [stats, setStats] = useState<ProjectStageStats[]>([])
  useEffect(() => {
    let alive = true
    api.getProjectStageStats().then((r) => { if (alive) setStats(r) }).catch(() => { if (alive) setStats([]) })
    return () => { alive = false }
  }, [])
  const total = stats.reduce((s, r) => s + r.count, 0)
  return (
    <div style={{ overflowX: 'auto', background: 'var(--sf-card)', borderRadius: '12px', border: '1px solid var(--bd-line)' }}>
      <table style={{ width: '100%', borderCollapse: 'collapse' }}>
        <thead>
          <tr>
            <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>مرحلة المشروع</th>
            <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>عدد المشاريع</th>
          </tr>
        </thead>
        <tbody>
          {stats.map((r) => (
            <tr key={r.stage}>
              <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid var(--bd-line)' }}>{r.stage}</td>
              <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid var(--bd-line)' }}>{r.count}</td>
            </tr>
          ))}
          {stats.length === 0 && (
            <tr><td colSpan={2} style={{ padding: '30px', textAlign: 'center', color: 'var(--t-faint)' }}>لا توجد مشاريع بعد</td></tr>
          )}
        </tbody>
        {stats.length > 0 && (
          <tfoot>
            <tr><td style={{ padding: '10px 12px', fontWeight: 'bold' }}>الإجمالي</td><td style={{ padding: '10px 12px', fontWeight: 'bold', color: PRIMARY_TEXT }}>{total}</td></tr>
          </tfoot>
        )}
      </table>
    </div>
  )
}

function monthStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`
}

/**
 * أكثر خدمة طلبها الزبائن — انتقلت لهنا من صفحة إحصائيات الموظفين،
 * لأن مكانها الصحيح إدارة الإحصائيات.
 */
function ServicesTab() {
  const [stats, setStats] = useState<Stats | null>(null)
  useEffect(() => { api.getStats().then(setStats).catch(() => setStats(null)) }, [])

  if (!stats) return <p style={{ color: 'var(--t-faint)', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>
  const rows = stats.serviceBreakdown
  const max = rows[0]?.count || 1

  return (
    <div style={{ background: 'var(--sf-card)', border: '1px solid var(--bd-line)', borderRadius: '12px', padding: '20px' }}>
      <h3 style={{ color: PRIMARY_TEXT, margin: '0 0 4px 0' }}>أكثر خدمة طلبها الزبائن</h3>
      <p style={{ color: 'var(--t-faint)', fontSize: '13px', margin: '0 0 16px 0' }}>
        مرتّبة من الأكثر طلباً — من شوكت بدت كل خدمة، وشكد دخّلت مبالغ وأرباح
      </p>
      {rows.map((s, i) => (
        <div key={s.serviceId || i} style={{ marginBottom: '14px' }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: '12px' }}>
            <span style={{ fontSize: '13px', width: '140px', color: 'var(--t-body)' }}>{s.name}</span>
            <div style={{ flex: 1, height: '22px', background: 'var(--sf-sunken)', borderRadius: '11px', overflow: 'hidden' }}>
              <div style={{ height: '100%', width: `${(s.count / max) * 100}%`, background: PRIMARY, borderRadius: '11px' }} />
            </div>
            <span style={{ fontSize: '13px', fontWeight: 'bold', width: '40px', color: PRIMARY_TEXT }}>{s.count}</span>
          </div>
          <div style={{ marginRight: '152px', marginTop: '3px', fontSize: '11px', color: 'var(--t-faint)', display: 'flex', gap: '14px', flexWrap: 'wrap' }}>
            <span>📅 من {s.firstAt ? new Date(s.firstAt).toLocaleDateString('ar-IQ') : '—'} إلى {s.lastAt ? new Date(s.lastAt).toLocaleDateString('ar-IQ') : '—'}</span>
            <span>💰 دخل: <b style={{ color: PRIMARY_TEXT }}>{fmt(s.revenue)} د.ع</b></span>
            <span>📈 ربح: <b style={{ color: GOLD_TEXT }}>{fmt(s.profit)} د.ع</b></span>
          </div>
        </div>
      ))}
      <p style={{ marginTop: '12px', fontSize: '11px', color: 'var(--t-faint)' }}>
        الربح = المستلم ناقص كلفة المواد بسعر الجملة. المواد الي ماكو إلها سعر جملة بالكتالوج
        تُحسب بكلفة صفر، يعني الربح المعروض هو الحد الأعلى.
      </p>
      {rows.length === 0 && <p style={{ color: 'var(--t-faint)', fontSize: '13px' }}>ماكو بيانات حجوزات بعد</p>}
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
      <div style={{ background: 'var(--sf-card)', border: '1px solid var(--bd-line)', borderRadius: '10px', padding: '12px 16px', marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '12px' }}>
        <label style={{ fontSize: '13px', color: 'var(--t-muted)' }}>الشهر</label>
        <input type="month" value={month} onChange={(e) => e.target.value && setMonth(e.target.value)}
          style={{ padding: '8px 10px', border: '1px solid var(--bd-line)', borderRadius: '8px', fontSize: '14px' }} />
      </div>

      {!rep && <p style={{ color: 'var(--t-faint)', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>}

      {rep && (
        <>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '12px', marginBottom: '20px' }}>
            <StatCard label="أعمال انخلصت داخل الشركة" value={rep.inHouseCount} />
            <StatCard label="أعمال طلعت للزبون" value={rep.onSiteCount} />
            <StatCard label="نسبة الشغل الداخلي" value={`${share}%`} />
            <StatCard label="مبالغ الشغل الداخلي" value={`${fmt(rep.inHouseAmount)} د.ع`} />
          </div>

          <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '16px', marginBottom: '20px' }}>
            <div style={{ background: 'var(--sf-card)', border: '1px solid var(--bd-line)', borderRadius: '12px', padding: '16px' }}>
              <h4 style={{ color: PRIMARY_TEXT, margin: '0 0 12px 0', fontSize: '15px' }}>شنو انشتغل جوه</h4>
              {rep.services.map((s) => (
                <div key={s.name} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0', borderBottom: '1px solid #f2f2f2', fontSize: '13px' }}>
                  <span>{s.name}</span>
                  <span style={{ fontWeight: 'bold', color: PRIMARY_TEXT }}>{s.count} · {fmt(s.amount)} د.ع</span>
                </div>
              ))}
              {rep.services.length === 0 && <p style={{ color: 'var(--t-faint)', fontSize: '13px' }}>ماكو شغل داخلي بهذا الشهر</p>}
            </div>

            <div style={{ background: 'var(--sf-card)', border: '1px solid var(--bd-line)', borderRadius: '12px', padding: '16px' }}>
              <h4 style={{ color: PRIMARY_TEXT, margin: '0 0 12px 0', fontSize: '15px' }}>منو اشتغل جوه</h4>
              {rep.crew.map((c) => (
                <div key={c.employeeName} style={{ display: 'flex', justifyContent: 'space-between', padding: '6px 0', borderBottom: '1px solid #f2f2f2', fontSize: '13px' }}>
                  <span>{c.employeeName}</span>
                  <span style={{ fontWeight: 'bold', color: GOLD_TEXT }}>{c.count} عمل</span>
                </div>
              ))}
              {rep.crew.length === 0 && <p style={{ color: 'var(--t-faint)', fontSize: '13px' }}>ماكو كادر مسجّل</p>}
            </div>
          </div>

          <div style={{ overflowX: 'auto', background: 'var(--sf-card)', borderRadius: '12px', border: '1px solid var(--bd-line)' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr>
                  {['الكود', 'الخدمة', 'تاريخ الإنجاز', 'المبلغ'].map((h) => (
                    <th key={h} style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>{h}</th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {rep.works.map((w) => (
                  <tr key={w.code} style={{ borderBottom: '1px solid var(--bd-line)' }}>
                    <td style={{ padding: '9px 12px', fontSize: '13px' }}><BookingCodeChip code={w.code} /></td>
                    <td style={{ padding: '9px 12px', fontSize: '13px' }}>{w.serviceName}</td>
                    <td style={{ padding: '9px 12px', fontSize: '13px' }}>
                      {w.completedAt ? new Date(w.completedAt).toLocaleDateString('ar-IQ') : '—'}
                    </td>
                    <td style={{ padding: '9px 12px', fontSize: '13px' }}>{fmt(w.amount)} د.ع</td>
                  </tr>
                ))}
                {rep.works.length === 0 && (
                  <tr><td colSpan={4} style={{ padding: '30px', textAlign: 'center', color: 'var(--t-faint)' }}>ماكو أعمال داخلية بهذا الشهر</td></tr>
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
  const [tab, setTab] = useState<'daily' | 'weekly' | 'monthly' | 'internal' | 'services' | 'projects'>('daily')

  const tabs: { key: typeof tab; label: string }[] = [
    { key: 'daily', label: 'يومية' },
    { key: 'weekly', label: 'أسبوعية' },
    { key: 'monthly', label: 'شهرية' },
    { key: 'internal', label: 'داخل الشركة' },
    { key: 'services', label: 'أكثر خدمة مطلوبة' },
    { key: 'projects', label: 'المشاريع' },
  ]

  return (
    <div style={{ direction: 'rtl', fontFamily: "'Segoe UI', Tahoma, Arial, sans-serif" }}>
      <div style={{
        background: `linear-gradient(135deg, ${PRIMARY}, #283593)`,
        color: 'white', padding: '20px 30px', borderRadius: '12px', marginBottom: '20px',
      }}>
        <h1 style={{ margin: 0, fontSize: '24px' }}>إدارة الإحصائيات</h1>
        <span style={{ color: GOLD_TEXT, fontSize: '14px' }}>يومية، أسبوعية، شهرية، شغل داخل الشركة، أكثر خدمة مطلوبة، ومشاريع — حصراً لمدير النظام</span>
      </div>

      <div style={{ display: 'flex', gap: '8px', marginBottom: '20px' }}>
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
      {tab === 'monthly' && <EmployeeMonthlyStatsPage />}
      {tab === 'internal' && <InternalWorksTab />}
      {tab === 'services' && <ServicesTab />}
      {tab === 'projects' && <ProjectsTab />}
    </div>
  )
}
