import { forwardRef, useId } from 'react'
import type { EyeMood } from './matrixEyeColors'

// ═══ عيون المالك — رينغان وشارنغان (طلب (ع) 10-04) ═══
// بحساب المالك بس، بدل عين ماتركس. مرسومة بالكود (SVG) وحركتها CSS بس —
// ما تثقّل الشغل. الضغط عليها يفتح نفس تقرير ماتركس.
// ref = مجموعة البؤبؤين، حتى يتبعن الماوس مثل العين العادية.

interface Props { mood: EyeMood; lid?: number; width?: number }

const OwnerEyes = forwardRef<SVGGElement, Props>(function OwnerEyes({ mood, lid = 0.95, width = 96 }, pupilRef) {
  const uid = useId().replace(/:/g, '')
  const angry = mood === 'ANGRY'
  const h = 15 * Math.max(lid, 0.08)
  const eye = (cx: number) => `M${cx - 22} 0 Q${cx} ${-h} ${cx + 22} 0 Q${cx} ${h} ${cx - 22} 0Z`
  const glow = angry ? '#ef4444' : '#a78bfa'

  // توموي: دائرة صغيرة وذيلها منحني — ثلاثة حول البؤبؤ بزوايا ١٢٠°.
  const tomoe = (a: number) => (
    <g key={a} transform={`rotate(${a})`}>
      <circle cx="0" cy="-6.3" r="1.7" fill="#111" />
      <path d="M1.6 -6.6 Q3.6 -4.6 2.4 -2.4" fill="none" stroke="#111" strokeWidth="1.1" strokeLinecap="round" />
    </g>
  )

  return (
    <svg viewBox="-56 -20 112 40" width={width} height={(width * 40) / 112} className={`owner-eyes ${angry ? 'is-angry' : ''}`} style={{ overflow: 'visible' }}>
      <defs>
        <radialGradient id={`rin${uid}`} cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="#e9d5ff" />
          <stop offset="100%" stopColor="#a78bfa" />
        </radialGradient>
        <radialGradient id={`shar${uid}`} cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="#ff4d4d" />
          <stop offset="100%" stopColor="#9f0d0d" />
        </radialGradient>
        <clipPath id={`cl1${uid}`}><path d={eye(-26)} /></clipPath>
        <clipPath id={`cl2${uid}`}><path d={eye(26)} /></clipPath>
        <filter id={`gl${uid}`} x="-50%" y="-50%" width="200%" height="200%">
          <feGaussianBlur stdDeviation="1.4" result="b" />
          <feMerge><feMergeNode in="b" /><feMergeNode in="SourceGraphic" /></feMerge>
        </filter>
      </defs>

      {/* بياض العينين */}
      <path d={eye(-26)} fill="#f8fafc" stroke={glow} strokeWidth="1.3" filter={`url(#gl${uid})`} />
      <path d={eye(26)} fill="#f8fafc" stroke={angry ? '#ef4444' : '#f87171'} strokeWidth="1.3" filter={`url(#gl${uid})`} />

      <g ref={pupilRef}>
        {/* الرينغان — يسار: حلقات متداخلة */}
        <g clipPath={`url(#cl1${uid})`}>
          <g transform="translate(-26 0)" className="owner-rinnegan">
            <circle r="10.5" fill={`url(#rin${uid})`} />
            {[1.6, 3.4, 5.2, 7, 8.8, 10.4].map((r) => <circle key={r} r={r} fill="none" stroke="#4c1d95" strokeWidth="0.7" />)}
            <circle r="1.3" fill="#1e1b4b" />
          </g>
        </g>
        {/* الشارنغان — يمين: ثلاث توموي تدور */}
        <g clipPath={`url(#cl2${uid})`}>
          <g transform="translate(26 0)">
            <circle r="10.5" fill={`url(#shar${uid})`} />
            <circle r="6.3" fill="none" stroke="#111" strokeWidth="0.6" opacity="0.7" />
            <g className="owner-sharingan-spin">{[0, 120, 240].map(tomoe)}</g>
            <circle r="2.3" fill="#111" />
          </g>
        </g>
      </g>
    </svg>
  )
})

export default OwnerEyes
