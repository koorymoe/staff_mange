import { useEffect, useMemo, useState } from 'react'
import { api, type EmployeePerformanceCurve } from '../api'

const PRIMARY = '#1a237e'
const BLUE = '#2a78d6'
const ORANGE = '#eb6834'
const MUTED = '#898781'
const GRID = '#e1e0d9'

const monthLabel = (m: string) => {
  const [, mm] = m.split('-')
  const names = ['يناير', 'فبراير', 'مارس', 'أبريل', 'مايو', 'يونيو', 'يوليو', 'أغسطس', 'سبتمبر', 'أكتوبر', 'نوفمبر', 'ديسمبر']
  return names[Number(mm) - 1] || m
}

function AnimatedLineChart({
  title, unit, color, values, formatValue,
}: {
  title: string
  unit: string
  color: string
  values: { month: string; value: number }[]
  formatValue: (n: number) => string
}) {
  const [drawn, setDrawn] = useState(false)
  const [hover, setHover] = useState<number | null>(null)

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setDrawn(false)
    const t = setTimeout(() => setDrawn(true), 50)
    return () => clearTimeout(t)
  }, [values])

  const W = 480
  const H = 180
  const PAD = { top: 16, right: 16, bottom: 28, left: 16 }
  const max = Math.max(1, ...values.map((v) => v.value))
  const stepX = values.length > 1 ? (W - PAD.left - PAD.right) / (values.length - 1) : 0

  const points = values.map((v, i) => ({
    x: PAD.left + i * stepX,
    y: PAD.top + (H - PAD.top - PAD.bottom) * (1 - v.value / max),
    ...v,
  }))

  const linePath = points.map((p, i) => `${i === 0 ? 'M' : 'L'} ${p.x} ${p.y}`).join(' ')
  const areaPath = points.length > 0
    ? `${linePath} L ${points[points.length - 1].x} ${H - PAD.bottom} L ${points[0].x} ${H - PAD.bottom} Z`
    : ''

  const pathLength = useMemo(() => 1200, [])

  return (
    <div style={{ background: 'white', border: '1px solid #e0e0e0', borderRadius: '12px', padding: '16px', flex: 1, minWidth: '320px' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '8px' }}>
        <span style={{ width: '10px', height: '10px', borderRadius: '50%', background: color, display: 'inline-block' }} />
        <h4 style={{ margin: 0, fontSize: '14px', color: PRIMARY }}>{title}</h4>
      </div>
      <svg viewBox={`0 0 ${W} ${H}`} style={{ width: '100%', height: 'auto', overflow: 'visible' }}>
        {[0.25, 0.5, 0.75].map((f) => (
          <line key={f} x1={PAD.left} x2={W - PAD.right} y1={PAD.top + (H - PAD.top - PAD.bottom) * f} y2={PAD.top + (H - PAD.top - PAD.bottom) * f} stroke={GRID} strokeWidth={1} />
        ))}
        <defs>
          <linearGradient id={`grad-${title}`} x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" stopColor={color} stopOpacity="0.22" />
            <stop offset="100%" stopColor={color} stopOpacity="0" />
          </linearGradient>
        </defs>
        {areaPath && (
          <path d={areaPath} fill={`url(#grad-${title})`} style={{ opacity: drawn ? 1 : 0, transition: 'opacity 900ms ease 500ms' }} />
        )}
        {linePath && (
          <path
            d={linePath} fill="none" stroke={color} strokeWidth={2.5} strokeLinecap="round" strokeLinejoin="round"
            strokeDasharray={pathLength}
            strokeDashoffset={drawn ? 0 : pathLength}
            style={{ transition: 'stroke-dashoffset 1200ms cubic-bezier(0.4, 0, 0.2, 1)' }}
          />
        )}
        {points.map((p, i) => (
          <g key={i} onMouseEnter={() => setHover(i)} onMouseLeave={() => setHover(null)}>
            <circle cx={p.x} cy={p.y} r={hover === i ? 6 : 4} fill="white" stroke={color} strokeWidth={2}
              style={{ opacity: drawn ? 1 : 0, transition: `opacity 400ms ease ${600 + i * 80}ms`, cursor: 'pointer' }} />
            <rect x={p.x - stepX / 2} y={0} width={Math.max(stepX, 1)} height={H} fill="transparent" />
            <text x={p.x} y={H - 6} textAnchor="middle" fontSize="10" fill={MUTED}>{monthLabel(p.month)}</text>
          </g>
        ))}
        {hover !== null && points[hover] && (
          <g>
            <line x1={points[hover].x} x2={points[hover].x} y1={PAD.top} y2={H - PAD.bottom} stroke={color} strokeWidth={1} strokeDasharray="3,3" opacity={0.5} />
            <rect x={Math.min(Math.max(points[hover].x - 40, 0), W - 80)} y={2} width={80} height={20} rx={5} fill={PRIMARY} />
            <text x={Math.min(Math.max(points[hover].x, 40), W - 40)} y={16} textAnchor="middle" fontSize="11" fill="white" fontWeight="bold">
              {formatValue(points[hover].value)} {unit}
            </text>
          </g>
        )}
      </svg>
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
    api.getEmployeePerformanceCurve(employeeId, month, 6).then(setCurve).catch(() => setCurve(null))
  }, [employeeId, month])

  return (
    <div style={{ position: 'fixed', inset: 0, background: 'rgba(0,0,0,0.5)', zIndex: 60, display: 'flex', alignItems: 'center', justifyContent: 'center', padding: '16px' }} onClick={onClose}>
      <div
        style={{ background: '#f7f7f9', borderRadius: '16px', padding: '20px', width: '100%', maxWidth: '820px', maxHeight: '90vh', overflowY: 'auto', direction: 'rtl' }}
        onClick={(e) => e.stopPropagation()}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '16px', flexWrap: 'wrap', gap: '10px' }}>
          <h3 style={{ margin: 0, color: PRIMARY, fontSize: '18px' }}>📈 منحنى أداء {employeeName}</h3>
          <div style={{ display: 'flex', alignItems: 'center', gap: '8px' }}>
            <label style={{ fontSize: '13px', color: '#666' }}>الشهر</label>
            <input
              type="month" value={month} onChange={(e) => e.target.value && setMonth(e.target.value)}
              style={{ padding: '6px 10px', border: '1px solid #ddd', borderRadius: '8px', fontSize: '13px' }}
            />
            <button onClick={onClose} style={{ background: 'none', border: 'none', fontSize: '22px', color: '#999', cursor: 'pointer' }}>×</button>
          </div>
        </div>

        {!curve && <p style={{ textAlign: 'center', color: '#999', padding: '30px' }}>جاري تحميل المنحنى...</p>}

        {curve && (
          <div style={{ display: 'flex', gap: '16px', flexWrap: 'wrap' }}>
            <AnimatedLineChart
              title={`نقاط الكي بي اي (6 أشهر تنتهي بـ ${monthLabel(month)})`}
              unit="نقطة"
              color={BLUE}
              values={curve.points.map((p) => ({ month: p.month, value: p.points }))}
              formatValue={(n) => n.toLocaleString('en-IQ')}
            />
            <AnimatedLineChart
              title={`المبالغ المحصّلة نسبة إلى النقاط (6 أشهر تنتهي بـ ${monthLabel(month)})`}
              unit="د.ع"
              color={ORANGE}
              values={curve.commission.map((c) => ({ month: c.month, value: c.amount }))}
              formatValue={(n) => n.toLocaleString('en-IQ')}
            />
          </div>
        )}
      </div>
    </div>
  )
}
