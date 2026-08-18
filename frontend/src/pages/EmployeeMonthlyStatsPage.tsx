import { useCallback, useEffect, useState } from 'react'
import { api, type EmployeeMonthlyStats } from '../api'
import KpiBreakdownChart from '../components/KpiBreakdownChart'
import PerformanceCurveModal from '../components/PerformanceCurveModal'

const PRIMARY = '#1a237e'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const PRIMARY_TEXT = 'var(--brand-ink)'
const GOLD = '#c8a45a'
// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const GOLD_TEXT = 'var(--gold-ink)'

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

  const totalCommissionSum = rows.reduce((sum, r) => sum + r.totalCommission, 0)

  const thStyle: React.CSSProperties = {
    padding: '10px 12px', textAlign: 'right', fontSize: '13px', color: 'white',
    background: PRIMARY, whiteSpace: 'nowrap',
  }
  const tdStyle: React.CSSProperties = {
    padding: '10px 12px', textAlign: 'right', fontSize: '13px', borderBottom: '1px solid var(--bd-line)',
  }

  return (
    <div style={{ direction: 'rtl', fontFamily: "'Segoe UI', Tahoma, Arial, sans-serif" }}>
      <div style={{
        background: `linear-gradient(135deg, ${PRIMARY}, #283593)`,
        color: 'white', padding: '20px 30px', borderRadius: '12px', marginBottom: '24px',
      }}>
        <h1 style={{ margin: 0, fontSize: '24px' }}>إحصائيات الموظفين الشهرية</h1>
        <span style={{ color: GOLD_TEXT, fontSize: '14px' }}>نقاط الكي بي اي، العمولات، المبيعات، الحجوزات، والشكاوى — لكل موظف كل شهر</span>
      </div>

      <div style={{
        display: 'flex', gap: '12px', alignItems: 'end', marginBottom: '20px',
        background: 'var(--sf-card)', border: `2px solid ${PRIMARY}`, borderRadius: '12px', padding: '16px',
      }}>
        <div>
          <label style={{ display: 'block', marginBottom: '4px', fontSize: '13px', color: 'var(--t-muted)' }}>الشهر</label>
          <input
            type="month"
            value={month}
            onChange={(e) => setMonth(e.target.value)}
            style={{ padding: '10px 12px', border: '1px solid var(--bd-line)', borderRadius: '8px', fontSize: '14px' }}
          />
        </div>
        <button
          onClick={handleExport}
          disabled={exporting || loading}
          style={{
            // ⚠️ نص غامق ثابت مو `PRIMARY_TEXT`: الخلفية ذهبية **بالوضعين**
            // (زر صامت مو سطح)، فقلب النص لفاتح يخلّيه فاتحاً على ذهبي.
            background: GOLD, color: '#1a237e', border: 'none', padding: '10px 20px',
            borderRadius: '8px', cursor: exporting ? 'not-allowed' : 'pointer', fontWeight: 'bold', fontSize: '14px',
            opacity: exporting ? 0.6 : 1,
          }}
        >
          {exporting ? 'جاري التصدير...' : 'تنزيل إكسل'}
        </button>
      </div>

      {loading && <p style={{ color: 'var(--t-faint)', textAlign: 'center', padding: '40px' }}>جاري التحميل...</p>}
      {error && (
        <div style={{ background: 'var(--sf-danger)', border: '1px solid #fecaca', borderRadius: '8px', padding: '16px', color: 'var(--t-danger)' }}>
          تعذر الاتصال بالخادم: {error}
        </div>
      )}

      {!loading && !error && (
        <div style={{ overflowX: 'auto', background: 'var(--sf-card)', borderRadius: '12px', border: '1px solid var(--bd-line)' }}>
          <table style={{ width: '100%', borderCollapse: 'collapse' }}>
            <thead>
              <tr>
                <th style={thStyle}>الموظف</th>
                <th style={thStyle}>الدور</th>
                <th style={thStyle}>عدد الخدمات التي يعرفها</th>
                <th style={thStyle} title="نسبة زمن الموظف للمتوسط العام بنفس المنظومة ونوع الشغل — فوق ١ أسرع، تحت ١ أبطأ">سرعة العمل</th>
                <th style={thStyle}>نظافة السيارة</th>
                <th style={thStyle}>الشكاوى</th>
                <th style={thStyle}>عدد المبيعات</th>
                <th style={thStyle}>الحجوزات المكتملة</th>
                <th style={thStyle}>كل الحجوزات المسندة</th>
                <th style={thStyle}>حجوزات الصيانة</th>
                <th style={thStyle}>صيانات مجانية</th>
                <th style={thStyle}>أعمال داخل الشركة</th>
                <th style={thStyle}>نوعهن</th>
                <th style={thStyle} title="النقاط الكاملة: الحجوزات + السرعة + التقارير + الحضور + الشكاوى + التقييم اليدوي">نقاط الكي بي اي</th>
                <th style={thStyle} title="التقييمات اليدوية بس (خصومات/مكافآت المدير)">تقييم يدوي</th>
                <th style={thStyle}>قيمة النقاط اليدوية</th>
                <th style={thStyle}>إجمالي العمولة (حجم المبيعات)</th>
                <th style={thStyle}>منحنى الأداء</th>
              </tr>
            </thead>
            <tbody>
              {rows.map((r) => (
                <tr key={r.employeeId}>
                  <td style={tdStyle}>{r.employeeName}</td>
                  <td style={tdStyle}>{r.role}</td>
                  <td style={tdStyle}>{r.servicesKnownCount}</td>
                  {/* فوق ١ أسرع من المتوسط، تحت ١ أبطأ. نعرض عدد العيّنات
                      لأن رقم مبني على عيّنتين مو نفس رقم مبني على عشرين. */}
                  <td style={tdStyle}>
                    {r.workSpeedScore != null ? (
                      <span style={{ fontWeight: 700, color: r.workSpeedScore >= 1 ? '#059669' : '#dc2626' }}>
                        {r.workSpeedScore >= 1 ? '⬆︎' : '⬇︎'} {r.workSpeedScore.toFixed(2)}
                        <span style={{ color: 'var(--t-faint)', fontWeight: 400 }}> ({r.workSpeedSamples})</span>
                      </span>
                    ) : (
                      <span title="ماكو عيّنات كافية لهذا الشهر">—</span>
                    )}
                  </td>
                  <td style={tdStyle}>
                    {r.vehicleCleanlinessScore != null ? `${r.vehicleCleanlinessScore.toFixed(2)} (${r.vehicleRatingsCount})` : '—'}
                  </td>
                  <td style={tdStyle}>{r.complaintsCount}</td>
                  <td style={tdStyle}>{r.salesCount}</td>
                  <td style={tdStyle}>{r.completedBookingsCount}</td>
                  <td style={tdStyle}>{r.totalBookingsCount}</td>
                  <td style={tdStyle}>{r.maintenanceBookingsCount}</td>
                  <td style={tdStyle}>{r.freeMaintenanceCount}</td>
                  <td style={tdStyle}>{r.inHouseWorksCount ?? 0}</td>
                  <td style={{ ...tdStyle, maxWidth: '220px' }}>
                    {r.inHouseWorkTypes?.length ? r.inHouseWorkTypes.join('، ') : '—'}
                  </td>
                  {/* العدد قبل القيمة: الصفحة كانت تعرض الدينار بس،
                      فالمالك يشوف «٠ د.ع» ويظن ماكو نقاط — والحقيقة إن
                      العدد موجود ومحسوب، بس ما كان ينعرض. */}
                  <td style={tdStyle}>
                    <button
                      onClick={() => setKpiFor({ id: r.employeeId, name: r.employeeName })}
                      style={{
                        border: 'none', cursor: 'pointer', borderRadius: '999px', padding: '4px 12px',
                        fontWeight: 800, fontSize: '13px',
                        background: r.smartKpiPoints >= 100 ? '#dcfce7' : r.smartKpiPoints >= 50 ? '#fef3c7' : '#fee2e2',
                        color: r.smartKpiPoints >= 100 ? '#15803d' : r.smartKpiPoints >= 50 ? '#b45309' : '#b91c1c',
                      }}
                      title="اضغط لتشوف مخطط النقاط مفصّلة"
                    >
                      {r.smartKpiPoints} 📊
                    </button>
                  </td>
                  <td style={{ ...tdStyle, color: r.kpiPoints < 0 ? '#b91c1c' : r.kpiPoints > 0 ? '#15803d' : '#64748b' }}>
                    {r.kpiPoints > 0 ? `+${r.kpiPoints}` : r.kpiPoints}
                  </td>
                  <td style={tdStyle}>{fmt(r.kpiPointsValue)} د.ع</td>
                  <td style={{ ...tdStyle, fontWeight: 'bold', color: GOLD_TEXT }}>{fmt(r.totalCommission)} د.ع</td>
                  <td style={tdStyle}>
                    <button
                      onClick={() => setCurveFor({ id: r.employeeId, name: r.employeeName })}
                      style={{ background: PRIMARY, color: 'white', border: 'none', borderRadius: '6px', padding: '6px 10px', fontSize: '12px', cursor: 'pointer', fontWeight: 'bold' }}
                    >
                      📈 عرض
                    </button>
                  </td>
                </tr>
              ))}
              {rows.length === 0 && (
                <tr>
                  <td colSpan={14} style={{ ...tdStyle, textAlign: 'center', color: 'var(--t-faint)', padding: '40px' }}>
                    لا توجد بيانات لهذا الشهر
                  </td>
                </tr>
              )}
            </tbody>
            {rows.length > 0 && (
              <tfoot>
                <tr>
                  <td style={{ ...tdStyle, fontWeight: 'bold' }} colSpan={12}>الإجمالي</td>
                  <td style={{ ...tdStyle, fontWeight: 'bold', color: PRIMARY_TEXT }}>{fmt(totalCommissionSum)} د.ع</td>
                  <td style={tdStyle} />
                </tr>
              </tfoot>
            )}
          </table>
        </div>
      )}

      {kpiFor && (
        <KpiBreakdownChart
          employeeId={kpiFor.id}
          employeeName={kpiFor.name}
          month={month}
          onClose={() => setKpiFor(null)}
        />
      )}

      {curveFor && (
        <PerformanceCurveModal employeeId={curveFor.id} employeeName={curveFor.name} onClose={() => setCurveFor(null)} />
      )}
    </div>
  )
}
