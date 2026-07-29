import { useCallback, useEffect, useState } from 'react'
import { api, type EmployeeMonthlyStats } from '../api'

const PRIMARY = '#1a237e'
const GOLD = '#c8a45a'

const fmt = (n: number) => n.toLocaleString('en-IQ')

function getCurrentMonth(): string {
  const now = new Date()
  return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
}

// صفحة إحصائيات الموظفين الشهرية — تجمع نقاط الكي بي اي (نفس الآلية الموجودة
// أصلاً)، سرعة العمل (TODO: تُملأ بعد اكتمال ميزة تقدير مدة تنفيذ العمل)،
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

  const totalCommissionSum = rows.reduce((sum, r) => sum + r.totalCommission, 0)

  const thStyle: React.CSSProperties = {
    padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white',
    background: PRIMARY, whiteSpace: 'nowrap',
  }
  const tdStyle: React.CSSProperties = {
    padding: '10px 12px', textAlign: 'right', fontSize: '13px', borderBottom: '1px solid #eee',
  }

  return (
    <div style={{ direction: 'rtl', fontFamily: "'Segoe UI', Tahoma, Arial, sans-serif" }}>
      <div style={{
        background: `linear-gradient(135deg, ${PRIMARY}, #283593)`,
        color: 'white', padding: '20px 30px', borderRadius: '12px', marginBottom: '24px',
      }}>
        <h1 style={{ margin: 0, fontSize: '24px' }}>إحصائيات الموظفين الشهرية</h1>
        <span style={{ color: GOLD, fontSize: '14px' }}>نقاط الكي بي اي، العمولات، المبيعات، الحجوزات، والشكاوى — لكل موظف كل شهر</span>
      </div>

      <div style={{
        display: 'flex', gap: '12px', alignItems: 'end', marginBottom: '20px',
        background: 'white', border: `2px solid ${PRIMARY}`, borderRadius: '12px', padding: '16px',
      }}>
        <div>
          <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: '#666' }}>الشهر</label>
          <input
            type="month"
            value={month}
            onChange={(e) => setMonth(e.target.value)}
            style={{ padding: '10px 12px', border: '1px solid #ddd', borderRadius: '8px', fontSize: '14px' }}
          />
        </div>
        <button
          onClick={handleExport}
          disabled={exporting || loading}
          style={{
            background: GOLD, color: PRIMARY, border: 'none', padding: '10px 20px',
            borderRadius: '8px', cursor: exporting ? 'not-allowed' : 'pointer', fontWeight: 'bold', fontSize: '14px',
            opacity: exporting ? 0.6 : 1,
          }}
        >
          {exporting ? 'جاري التصدير...' : 'تنزيل إكسل'}
        </button>
      </div>

      {loading && <p style={{ color: '#999', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>}
      {error && (
        <div style={{ background: '#fef2f2', border: '1px solid #fecaca', borderRadius: '8px', padding: '16px', color: '#dc2626' }}>
          تعذر الاتصال بالخادم: {error}
        </div>
      )}

      {!loading && !error && (
        <div style={{ overflowX: 'auto', background: 'white', borderRadius: '12px', border: '1px solid #e0e0e0' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr>
                <th style={thStyle}>الموظف</th>
                <th style={thStyle}>الدور</th>
                <th style={thStyle}>نقاط الكي بي اي</th>
                <th style={thStyle}>سرعة العمل</th>
                <th style={thStyle}>نظافة السيارة</th>
                <th style={thStyle}>الشكاوى</th>
                <th style={thStyle}>عدد المبيعات</th>
                <th style={thStyle}>الحجوزات المكتملة</th>
                <th style={thStyle}>كل الحجوزات المسندة</th>
                <th style={thStyle}>حجوزات الصيانة</th>
                <th style={thStyle}>صيانات مجانية</th>
                <th style={thStyle}>قيمة نقاط الكي بي اي</th>
                <th style={thStyle}>إجمالي العمولة (حجم المبيعات)</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((r) => (
                <tr key={r.employeeId}>
                  <td style={tdStyle}>{r.employeeName}</td>
                  <td style={tdStyle}>{r.role}</td>
                  <td style={tdStyle}>{r.kpiPoints}</td>
                  <td style={tdStyle}>{r.workSpeedScore != null ? r.workSpeedScore.toFixed(2) : '—'}</td>
                  <td style={tdStyle}>
                    {r.vehicleCleanlinessScore != null ? `${r.vehicleCleanlinessScore.toFixed(2)} (${r.vehicleRatingsCount})` : '—'}
                  </td>
                  <td style={tdStyle}>{r.complaintsCount}</td>
                  <td style={tdStyle}>{r.salesCount}</td>
                  <td style={tdStyle}>{r.completedBookingsCount}</td>
                  <td style={tdStyle}>{r.totalBookingsCount}</td>
                  <td style={tdStyle}>{r.maintenanceBookingsCount}</td>
                  <td style={tdStyle}>{r.freeMaintenanceCount}</td>
                  <td style={tdStyle}>{fmt(r.kpiPointsValue)} د.ع</td>
                  <td style={{ ...tdStyle, fontWeight: 'bold', color: GOLD }}>{fmt(r.totalCommission)} د.ع</td>
                </tr>
              ))}
              {rows.length === 0 && (
                <tr>
                  <td colSpan={13} style={{ ...tdStyle, textAlign: 'center', color: '#999', padding: '40px' }}>
                    لا توجد بيانات لهذا الشهر
                  </td>
                </tr>
              )}
            </tbody>
            {rows.length > 0 && (
              <tfoot>
                <tr>
                  <td style={{ ...tdStyle, fontWeight: 'bold' }} colSpan={12}>الإجمالي</td>
                  <td style={{ ...tdStyle, fontWeight: 'bold', color: PRIMARY }}>{fmt(totalCommissionSum)} د.ع</td>
                </tr>
              </tfoot>
            )}
          </table>
        </div>
      )}
    </div>
  )
}
