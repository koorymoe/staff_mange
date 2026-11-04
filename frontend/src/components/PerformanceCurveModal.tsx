import { useEffect, useState } from 'react'
import { api, type EmployeePerformanceCurve } from '../api'

// ⚠️ نسخة **النص** تنقلب بالوضع الليلي، والأصل يبقى للأسطح:
// نفس اللون يخدم عنواناً غامقاً على أبيض، ورأس جدول كحلي عليه نص أبيض.
// قلب الاثنين سوا يكسر واحداً منهما — نفس فخّ --color-white.
const PRIMARY_TEXT = 'var(--brand-ink)'
const BLUE = '#2a78d6'
const ORANGE = '#eb6834'

const monthLabel = (m: string) => {
  const [, mm] = m.split('-')
  const names = ['يناير', 'فبراير', 'مارس', 'أبريل', 'مايو', 'يونيو', 'يوليو', 'أغسطس', 'سبتمبر', 'أكتوبر', 'نوفمبر', 'ديسمبر']
  return names[Number(mm) - 1] || m
}

// AnimatedStat: عرض قيمة شهر واحد بس (بدون خلط أشهر ببعض) — رقم متحرك (count-up)
// مع شريط تعبئة متحرك، فاهمين إن الإحصائية شهرية بحتة مو نافذة متعددة الأشهر.
function AnimatedStat({
  title, unit, color, value, formatValue,
}: {
  title: string
  unit: string
  color: string
  value: number
  formatValue: (n: number) => string
}) {
  const [displayValue, setDisplayValue] = useState(0)
  const [fillPct, setFillPct] = useState(0)

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setDisplayValue(0)
    setFillPct(0)
    const start = performance.now()
    const duration = 900
    let raf = 0
    const tick = (now: number) => {
      const t = Math.min(1, (now - start) / duration)
      const eased = 1 - Math.pow(1 - t, 3)
      setDisplayValue(Math.round(value * eased))
      setFillPct(eased * 100)
      if (t < 1) raf = requestAnimationFrame(tick)
    }
    raf = requestAnimationFrame(tick)
    return () => cancelAnimationFrame(raf)
  }, [value])

  return (
    <div style={{ background: 'var(--sf-card)', border: '1px solid var(--bd-line)', borderRadius: '12px', padding: '20px', flex: 1, minWidth: '260px' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '10px' }}>
        <span style={{ width: '10px', height: '10px', borderRadius: '50%', background: color, display: 'inline-block' }} />
        <h4 style={{ margin: 0, fontSize: '14px', color: PRIMARY_TEXT }}>{title}</h4>
      </div>
      <div style={{ fontSize: '32px', fontWeight: 'bold', color, lineHeight: 1.2 }}>
        {formatValue(displayValue)} <span style={{ fontSize: '14px', color: 'var(--t-faint)', fontWeight: 'normal' }}>{unit}</span>
      </div>
      <div style={{ marginTop: '14px', height: '8px', borderRadius: '4px', background: 'var(--sf-sunken)', overflow: 'hidden' }}>
        <div style={{ width: `${fillPct}%`, height: '100%', background: color, borderRadius: '4px', transition: 'width 60ms linear' }} />
      </div>
    </div>
  )
}

export default function PerformanceCurveModal({ employeeId, employeeName, onClose }: { employeeId: string; employeeName: string; onClose: () => void }) {
  const [curve, setCurve] = useState<EmployeePerformanceCurve | null>(null)
  const [month, setMonth] = useState(() => {
    const now = new Date()
    return `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`
  })

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setCurve(null)
    api.getEmployeePerformanceCurve(employeeId, month, 1).then(setCurve).catch(() => setCurve(null))
  }, [employeeId, month])

  const monthPoints = curve?.points[curve.points.length - 1]?.points ?? 0
  const monthCommission = curve?.commission[curve.commission.length - 1]?.amount ?? 0

  return (
    <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', zIndex: 60, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '16px' }} onClick={onClose}>
      <div
        style={{ background: 'var(--sf-sunken)', borderRadius: '16px', padding: '20px', width: '100%', maxWidth: '620px', maxHeight: '90vh', overflowY: 'auto', direction: 'rtl' }}
        onClick={(e) => e.stopPropagation()}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', flexWrap: 'wrap', gap: '10px' }}>
          <h3 style={{ margin: 0, color: PRIMARY_TEXT, fontSize: '18px' }}>📈 أداء {employeeName}</h3>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <label style={{ fontSize: '13px', color: 'var(--t-muted)' }}>الشهر</label>
            <input
              type="month" value={month} onChange={(e) => e.target.value && setMonth(e.target.value)}
              style={{ padding: '6px 10px', border: '1px solid var(--bd-line)', borderRadius: '8px', fontSize: '13px' }}
            />
            <button onClick={onClose} style={{ background: 'none', border: 'none', fontSize: '22px', color: 'var(--t-faint)', cursor: 'pointer' }}>×</button>
          </div>
        </div>

        {!curve && <p style={{ textAlign: 'center', color: 'var(--t-faint)', padding: '30px' }}>جاري التحميل...</p>}

        {curve && (
          <>
            <p style={{ margin: '0 0 12px', fontSize: '13px', color: 'var(--t-faint)' }}>بيانات شهر {monthLabel(month)} فقط — بدون خلط مع أي شهر آخر.</p>
            <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
              <AnimatedStat title="نقاط الكي بي اي" unit="نقطة" color={BLUE} value={monthPoints} formatValue={(n) => n.toLocaleString('en-IQ')} />
              <AnimatedStat title="المبالغ المحصّلة نسبة إلى النقاط" unit="د.ع" color={ORANGE} value={monthCommission} formatValue={(n) => n.toLocaleString('en-IQ')} />
            </div>
          </>
        )}
      </div>
    </div>
  )
}
