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

function DailyTab() {
  const [stats, setStats] = useState<DailyStats | null>(null)
  useEffect(() => { api.getDailyStats().then(setStats) }, [])
  if (!stats) return <p style={{ color: '#999', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>
  return (
    <div>
      <div style={{ background: 'white', border: `1px solid #e0e0e0`, borderRadius: '12px', padding: '16px', marginBottom: '16px' }}>
        <b>عدد حجوزات اليوم ({stats.date}):</b> <span style={{ color: PRIMARY, fontWeight: 'bold', fontSize: '18px' }}>{stats.totalBookingsToday}</span>
      </div>
      <div style={{ overflowX: 'auto', background: 'white', borderRadius: '12px', border: '1px solid #e0e0e0' }}>
        <table style={{ width: '100%', borderCollapse: 'collapse' }}>
          <thead>
            <tr>
              <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>الموظف</th>
              <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>الدور</th>
              <th style={{ padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white', background: PRIMARY }}>حجوزاته اليوم</th>
            </tr>
          </thead>
          <tbody>
            {stats.employees.map((e) => (
              <tr key={e.employeeId}>
                <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{e.employeeName}</td>
                <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{e.role}</td>
                <td style={{ padding: '10px 12px', fontSize: '13px', borderBottom: '1px solid #eee' }}>{e.bookingsToday}</td>
              </tr>
            ))}
            {stats.employees.length === 0 && (
              <tr><td colSpan={3} style={{ padding: '30px', textAlign: 'center', color: '#999' }}>لا يوجد كادر عنده حجوزات اليوم</td></tr>
            )}
          </tbody>
        </table>
      </div>
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
