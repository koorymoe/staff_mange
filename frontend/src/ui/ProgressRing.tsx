import type { ReactNode } from 'react'
import { ringColor } from './pct'

// ═══ (دازاي 11-15) حلقة التقدّم الموحّدة ═══
// چانت منسوخة ٤ مرات وكل وحدة بحجم ولون. هسه وحدة: الحجم والسماكة واللون
// يتغيّرون بالـprops، والوسط (children) والتسمية تحته اختياريين.
// ⚠️ إذا اللون ما انطى، يتبع النسبة: ٨٠+ أخضر، ٥٠+ عنبري، والباقي برتقالي.

export default function ProgressRing({
  percent, size = 56, stroke = 5, color, track = '#eef2f7', children, label, labelClassName = 'text-[8px] text-slate-400', animate,
}: {
  percent: number
  size?: number
  stroke?: number
  color?: string
  track?: string
  children?: ReactNode
  label?: ReactNode
  labelClassName?: string
  animate?: boolean
}) {
  const p = Math.max(0, Math.min(100, isFinite(percent) ? percent : 0))
  const c = color ?? ringColor(p)
  const r = (size - stroke) / 2
  const circ = 2 * Math.PI * r
  const hasCenter = children != null || label != null
  return (
    <div className="relative shrink-0" style={{ width: size, height: size }}>
      <svg width={size} height={size} viewBox={`0 0 ${size} ${size}`} className="-rotate-90">
        <circle cx={size / 2} cy={size / 2} r={r} fill="none" stroke={track} strokeWidth={stroke} />
        <circle
          cx={size / 2} cy={size / 2} r={r} fill="none" stroke={c} strokeWidth={stroke} strokeLinecap="round"
          strokeDasharray={`${(p / 100) * circ} ${circ}`}
          className={animate ? 'transition-all duration-1000 ease-out' : undefined}
        />
      </svg>
      {hasCenter && (
        <div className="absolute inset-0 flex flex-col items-center justify-center leading-none">
          {children}
          {label != null && <span className={labelClassName}>{label}</span>}
        </div>
      )}
    </div>
  )
}
