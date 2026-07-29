import { useEffect, useState } from 'react'
import { api, type DailyStats, type WeeklyStats, type ProjectStageStats } from '../api'
import EmployeeMonthlyStatsPage from './EmployeeMonthlyStatsPage'

const PRIMARY = '#1a237e'
const GOLD = '#c8a45a'

const fmt = (n: number) => n.toLocaleString('en-IQ')

const roleLabels: Record<string, string> = {
  TECHNICIAN: 'فني/ليدر',
  SALES: 'مبيعات',
}

function todayStr() {
  const d = new Date()
  return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`
}

function StatCard({ label, value }: { label: string; value: string | number }) {
  return (
    <div style={{ background: 'white', border: '1px solid #e0e0e0', borderRadius: '10px', padding: '14px' }}>
      <div style={{ fontSize: '12px', color: '#888' }}>{label}</div>
      <div style={{ fontSize: '20px', fontWeight: 'bold', color: PRIMARY, marginTop: '4px' }}>{value}</div>
    </div>
  )
}

function DailyTab() {
  const [date, setDate] = useState(todayStr())
  const [stats, setStats] = useState<DailyStats | null>(null)
  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setStats(null)
    api.getDailyStats(date).then(setStats)
  }, [date])

  return (
    <div>
      <div style={{ background: 'white', border: '1px solid #e0e0e0', borderRadius: '10px', padding: '12px 16px', marginBottom: '16px', display: 'flex', alignItems: 'center', gap: '12px' }}>
        <label style={{ fontSize: '13px', color: '#666' }}>التاريخ</label>
        <input
          type="date"
          value={date}
          onChange={(e) => e.target.value && setDate(e.target.value)}
          style={{ padding: '8px 10px', border: '1px solid #ddd', borderRadius: '8px', fontSize: '14px' }}
        />
        {date !== todayStr() && (
          <button
            onClick={() => setDate(todayStr())}
            style={{ padding: '8px 14px', border: `1px solid ${PRIMARY}`, background: 'white', color: PRIMARY, borderRadius: '8px', fontSize: '13px', fontWeight: 'bold', cursor: 'pointer' }}
          >
            اليوم
          </button>
        )}
      </div>

      {!stats && <p style={{ color: '#999', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>}

      {stats && (
        <>
          <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '12px', marginBottom: '20px' }}>
            <StatCard label="إجمالي الحجوزات" value={stats.totalBookings} />
            <StatCard label="حجوزات صباحية" value={stats.morningBookings} />
            <StatCard label="حجوزات مسائية" value={stats.eveningBookings} />
            <StatCard label="كادر طلع للحجوزات" value={stats.crewOutCount} />
            <StatCard label="سيارات استُخدمت" value={stats.vehiclesOutCount} />
            <StatCard label="إجمالي عدد الموظفين" value={stats.totalEmployeesCount} />
            <StatCard label="إجمالي المبيعات" value={`${fmt(stats.totalSalesAmount)} د.ع`} />
            <StatCard label="إجمالي الأرباح" value={`${fmt(stats.totalProfitAmount)} د.ع`} />
          </div>

          <div style={{ overflowX: 'auto', background: 'white', borderRadius: '12px', border: '1px solid #e0e0e0' }}>
            <table style={{ width: '100%', borderCollapse: 'collapse' }}>
              <thead>
                <tr>
                  <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>الموظف</th>
                  <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>الدور</th>
                  <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>حجوزات ترحّلت له</th>
                  <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>نفّذ منهن</th>
                  <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>ما نفّذ</th>
                  <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>سجّل حضور؟</th>
                </tr>
              </thead>
              <tbody>
                {stats.employees.map((e) => (
                  <tr key={e.employeeId}>
                    <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{e.employeeName}</td>
                    <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{e.role}</td>
                    <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{e.bookingsAssigned}</td>
                    <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee', color: '#2e7d32', fontWeight: 'bold' }}>{e.bookingsCompleted}</td>
                    <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee', color: '#c62828' }}>{e.bookingsAssigned - e.bookingsCompleted}</td>
                    <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>
                      {e.checkedIn
                        ? <span style={{ color: '#2e7d32', fontWeight: 'bold' }}>✔ إي</span>
                        : <span style={{ color: '#c62828', fontWeight: 'bold' }}>✘ لا</span>}
                    </td>
                  </tr>
                ))}
                {stats.employees.length === 0 && (
                  <tr><td colSpan={6} style={{ padding: '30px', textAlign: 'center', color: '#999' }}>لا يوجد نشاط بهذا اليوم</td></tr>
                )}
              </tbody>
            </table>
          </div>
        </>
      )}
    </div>
  )
}

function WeeklyTab() {
  const [stats, setStats] = useState<WeeklyStats | null>(null)
  useEffect(() => { api.getWeeklyStats().then(setStats) }, [])
  if (!stats) return <p style={{ color: '#999', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>
  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '20px' }}>
      <div>
        <h3 style={{ color: PRIMARY, marginBottom: '10px' }}>إنتاجية الكوادر (آخر 7 أيام)</h3>
        <div style={{ overflowX: 'auto', background: 'white', borderRadius: '12px', border: '1px solid #e0e0e0' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr>
                <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>الموظف</th>
                <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>الدور</th>
                <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>حجوزات منجزة</th>
                <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>حجم المبيعات</th>
              </tr>
            </thead>
            <tbody>
              {stats.crew.map((c) => (
                <tr key={c.employeeId}>
                  <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{c.employeeName}</td>
                  <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{roleLabels[c.role] || c.role}</td>
                  <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{c.completedBookings}</td>
                  <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee', color: GOLD, fontWeight: 'bold' }}>{fmt(c.salesVolume)} د.ع</td>
                </tr>
              ))}
              {stats.crew.length === 0 && (
                <tr><td colSpan={4} style={{ padding: '30px', textAlign: 'center', color: '#999' }}>لا يوجد نشاط بآخر 7 أيام</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
      <div>
        <h3 style={{ color: PRIMARY, marginBottom: '10px' }}>موظفو المبيعات (حسب عدد الحجوزات المدخلة)</h3>
        <div style={{ overflowX: 'auto', background: 'white', borderRadius: '12px', border: '1px solid #e0e0e0' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr>
                <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>الموظف</th>
                <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>حجوزات أدخلها</th>
              </tr>
            </thead>
            <tbody>
              {stats.sales.map((s) => (
                <tr key={s.employeeId}>
                  <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{s.employeeName}</td>
                  <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{s.bookingsEntered}</td>
                </tr>
              ))}
              {stats.sales.length === 0 && (
                <tr><td colSpan={2} style={{ padding: '30px', textAlign: 'center', color: '#999' }}>لا يوجد نشاط بآخر 7 أيام</td></tr>
              )}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}

function ProjectsTab() {
  const [stats, setStats] = useState<ProjectStageStats[]>([])
  useEffect(() => { api.getProjectStageStats().then(setStats) }, [])
  const total = stats.reduce((s, r) => s + r.count, 0)
  return (
    <div style={{ overflowX: 'auto', background: 'white', borderRadius: '12px', border: '1px solid #e0e0e0' }}>
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
              <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{r.stage}</td>
              <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{r.count}</td>
            </tr>
          ))}
          {stats.length === 0 && (
            <tr><td colSpan={2} style={{ padding: '30px', textAlign: 'center', color: '#999' }}>لا توجد مشاريع بعد</td></tr>
          )}
        </tbody>
        {stats.length > 0 && (
          <tfoot>
            <tr><td style={{ padding: '10px 12px', fontWeight: 'bold' }}>الإجمالي</td><td style={{ padding: '10px 12px', fontWeight: 'bold', color: PRIMARY }}>{total}</td></tr>
          </tfoot>
        )}
      </table>
    </div>
  )
}

export default function StatsManagementPage() {
  const [tab, setTab] = useState<'daily' | 'weekly' | 'monthly' | 'projects'>('daily')

  const tabs: { key: typeof tab; label: string }[] = [
    { key: 'daily', label: 'يومية' },
    { key: 'weekly', label: 'أسبوعية' },
    { key: 'monthly', label: 'شهرية' },
    { key: 'projects', label: 'المشاريع' },
  ]

  return (
    <div style={{ direction: 'rtl', fontFamily: "'Segoe UI', Tahoma, Arial, sans-serif" }}>
      <div style={{
        background: `linear-gradient(135deg, ${PRIMARY}, #283593)`,
        color: 'white', padding: '20px 30px', borderRadius: '12px', marginBottom: '20px',
      }}>
        <h1 style={{ margin: 0, fontSize: '24px' }}>إدارة الإحصائيات</h1>
        <span style={{ color: GOLD, fontSize: '14px' }}>يومية، أسبوعية، شهرية، ومشاريع — حصراً لمدير النظام</span>
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
      {tab === 'projects' && <ProjectsTab />}
    </div>
  )
}
