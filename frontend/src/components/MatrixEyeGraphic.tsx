import { forwardRef, useId } from 'react'

// ═══ رسم عين ماتركس — مكوّن واحد للهيدر ولعيون المدير ═══
//
// نفس العين بتفاصيل تختلف حسب المجموعة (الدور)، ولون يغلب عليه الشعور:
// هادئة بلون الدور، راضية خضرة، منتبهة صفرة، غاضبة حمرة ببؤبؤ مشقوق
// وعروق وحاجب. والحلقات تدور ببطء — مسح مستمر.

import { eyeColor, type EyeGroup, type EyeMood } from './matrixEyeColors'

interface Props {
  group: EyeGroup
  mood: EyeMood
  /** فتحة الجفن 0..1.2 */
  lid?: number
  width?: number
  /** توهج «شفتك» */
  flash?: boolean
}

// تفاصيل الحلقة الخارجية لكل مجموعة — توقيعها البصري.
function GroupMark({ group, color }: { group: EyeGroup; color: string }) {
  const ticks = (n: number, r1: number, r2: number, w = 1) =>
    Array.from({ length: n }, (_, i) => {
      const a = (i / n) * Math.PI * 2
      return <line key={i} x1={Math.cos(a) * r1} y1={Math.sin(a) * r1} x2={Math.cos(a) * r2} y2={Math.sin(a) * r2} stroke={color} strokeWidth={w} />
    })
  switch (group) {
    case 'ADMINS': // تاج مزدوج
      return <g opacity="0.85"><circle r="15.5" fill="none" stroke={color} strokeWidth="0.7" /><circle r="17.5" fill="none" stroke={color} strokeWidth="0.5" strokeDasharray="1 2" /></g>
    case 'MONITORS': // مسح متقطع
      return <g opacity="0.8">{ticks(24, 14.5, 17, 0.8)}</g>
    case 'COORDINATORS': // نقاط شبكة
      return <g opacity="0.85">{Array.from({ length: 12 }, (_, i) => { const a = (i / 12) * Math.PI * 2; return <circle key={i} cx={Math.cos(a) * 16} cy={Math.sin(a) * 16} r="0.9" fill={color} /> })}</g>
    case 'FINANCE': // حافة عملة
      return <g opacity="0.8"><circle r="16" fill="none" stroke={color} strokeWidth="0.6" />{ticks(36, 16, 17.6, 0.5)}</g>
    case 'TECHS': // مسامير عُدّة
      return <g opacity="0.85">{ticks(8, 14.8, 17.4, 1.4)}<circle r="16" fill="none" stroke={color} strokeWidth="0.4" /></g>
    case 'DESIGN': // ريشة/منحنيات
      return <g opacity="0.85" fill="none" stroke={color} strokeWidth="0.7"><path d="M-16 6 Q-10 -18 0 -16 Q10 -18 16 6" /><path d="M-14 10 Q0 18 14 10" strokeDasharray="2 2" /></g>
    case 'LEADERS': // مثلثات
      return <g opacity="0.85">{Array.from({ length: 6 }, (_, i) => { const a = (i / 6) * Math.PI * 2; const x = Math.cos(a) * 16.5, y = Math.sin(a) * 16.5; return <path key={i} d={`M${x} ${y - 1.4} L${x + 1.3} ${y + 1} L${x - 1.3} ${y + 1}Z`} fill={color} /> })}</g>
    case 'QUALITY': // تصويب
      return <g opacity="0.85" stroke={color} strokeWidth="0.7"><line x1="-18" y1="0" x2="-13" y2="0" /><line x1="13" y1="0" x2="18" y2="0" /><line x1="0" y1="-18" x2="0" y2="-13" /><line x1="0" y1="13" x2="0" y2="18" /><circle r="16" fill="none" strokeDasharray="4 3" /></g>
    case 'IT': // دوائر إلكترونية
      return <g opacity="0.85" stroke={color} strokeWidth="0.6" fill="none"><path d="M-17 -4 h4 v-3 M17 4 h-4 v3 M-4 17 v-4 h-3 M4 -17 v4 h3" /><circle cx="-17" cy="-4" r="0.9" fill={color} /><circle cx="17" cy="4" r="0.9" fill={color} /></g>
    default:
      return <circle r="16" fill="none" stroke={color} strokeWidth="0.5" strokeDasharray="2 3" opacity="0.7" />
  }
}

const MatrixEyeGraphic = forwardRef<SVGGElement, Props>(function MatrixEyeGraphic({ group, mood, lid = 0.9, width = 76, flash = false }, pupilRef) {
  const uid = useId().replace(/:/g, '')
  const color = eyeColor(group, mood)
  const angry = mood === 'ANGRY'
  const pleased = mood === 'PLEASED'
  const h = 17 * lid
  const eyePath = `M-30 0 Q0 ${-h} 30 0 Q0 ${h} -30 0Z`
  const pupilW = angry ? 1.4 : flash ? 2.6 : pleased ? 4.6 : 3.8

  return (
    <svg viewBox="-34 -20 68 40" width={width} height={(width * 40) / 68} className={`mx-eye-svg ${angry ? 'is-angry' : ''}`} style={{ overflow: 'visible' }}>
      <defs>
        <radialGradient id={`ir${uid}`} cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="#fff" stopOpacity="0.95" />
          <stop offset="28%" stopColor={color} />
          <stop offset="70%" stopColor={color} stopOpacity="0.55" />
          <stop offset="100%" stopColor="#020617" />
        </radialGradient>
        <radialGradient id={`sc${uid}`} cx="50%" cy="50%" r="50%">
          <stop offset="60%" stopColor="#0b1220" />
          <stop offset="100%" stopColor={angry ? '#3b0a0a' : '#111827'} />
        </radialGradient>
        <clipPath id={`cl${uid}`}><path d={eyePath} /></clipPath>
        <filter id={`gl${uid}`} x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation={flash ? 2.4 : 1.2} result="b" />
          <feMerge><feMergeNode in="b" /><feMergeNode in="SourceGraphic" /></feMerge>
        </filter>
      </defs>

      {/* هالة خارجية + توقيع المجموعة (يدور) */}
      <g className="mx-orbit" filter={`url(#gl${uid})`}><GroupMark group={group} color={color} /></g>

      {/* بياض العين */}
      <path d={eyePath} fill={`url(#sc${uid})`} stroke={color} strokeWidth="1.4" filter={`url(#gl${uid})`} />

      <g clipPath={`url(#cl${uid})`}>
        {/* عروق حمرة بالغضب */}
        {angry && (
          <g stroke="#b91c1c" strokeWidth="0.45" fill="none" opacity="0.9" className="mx-veins">
            <path d="M-29 1 q6 -2 9 1 t7 0" /><path d="M29 -1 q-6 2 -9 -1 t-7 0" /><path d="M-26 -3 q4 1 6 -1" /><path d="M26 3 q-4 -1 -6 1" />
          </g>
        )}
        <g ref={pupilRef}>
          {/* القزحية: خطوط شعاعية حتى تبين حيّة */}
          <circle r="10.5" fill={`url(#ir${uid})`} />
          <g className="mx-iris" stroke={color} strokeOpacity="0.55" strokeWidth="0.35">
            {Array.from({ length: 28 }, (_, i) => { const a = (i / 28) * Math.PI * 2; return <line key={i} x1={Math.cos(a) * 4.5} y1={Math.sin(a) * 4.5} x2={Math.cos(a) * 10} y2={Math.sin(a) * 10} /> })}
          </g>
          <circle r="10.5" fill="none" stroke={color} strokeWidth="0.8" className="mx-ring" strokeDasharray="6 2 1 2" />
          <circle r="7" fill="none" stroke="#fff" strokeOpacity="0.18" strokeWidth="0.4" className="mx-ring-rev" strokeDasharray="1 3" />
          <ellipse rx={pupilW} ry={angry ? 6.5 : 5.2} fill="#000" />
          {angry && <ellipse rx={pupilW + 0.6} ry="6.9" fill="none" stroke="#ef4444" strokeWidth="0.5" />}
          <circle cx="-3.4" cy="-3.6" r="1.7" fill="#fff" opacity="0.9" />
          <circle cx="3" cy="2.8" r="0.7" fill="#fff" opacity="0.5" />
        </g>
        {/* خط مسح يعبر العين */}
        <rect x="-30" y="-0.6" width="60" height="1.2" fill={color} opacity="0.35" className="mx-scan" />
      </g>

      {/* حاجب غاضب */}
      {angry && <path d="M-27 -16 L-6 -9 M27 -16 L6 -9" stroke="#ef4444" strokeWidth="2.6" strokeLinecap="round" filter={`url(#gl${uid})`} />}
    </svg>
  )
})

export default MatrixEyeGraphic
