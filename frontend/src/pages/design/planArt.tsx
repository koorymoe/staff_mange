// ═══ رسومات «سيت المخططات» — (ك) 11-10 بطلب دازاي «اريد هذا بالضبط» ═══
// الرسومات مرسومة SVG بالكود (عمارة ونخيل وبلوبرنت ولفّات مخططات) بدل صور جاهزة:
// ما تحتاج ملفات ولا مزوّد خارجي، وتطلع حادة بكل حجم.
import type { ReactNode } from 'react'

const I = ({ children, size = 22 }: { children: ReactNode; size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.8" strokeLinecap="round" strokeLinejoin="round" aria-hidden>{children}</svg>
)
export const IconCity = (p: { size?: number }) => <I {...p}><path d="M3 21h18M5 21V8l5-3v16M10 21V3l9 4v14" /><path d="M13 9h3M13 12h3M13 15h3M7 11h1M7 14h1M7 17h1" /></I>
export const IconFolder = (p: { size?: number }) => <I {...p}><path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z" /></I>
export const IconLayers = (p: { size?: number }) => <I {...p}><path d="m12 3 9 5-9 5-9-5z" /><path d="m3 13 9 5 9-5" /></I>
export const IconDoc = (p: { size?: number }) => <I {...p}><path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z" /><path d="M14 3v5h5M9 13h6M9 17h6" /></I>
export const IconPin = (p: { size?: number }) => <I {...p}><path d="M12 21s-7-6.2-7-11a7 7 0 0 1 14 0c0 4.8-7 11-7 11z" /><circle cx="12" cy="10" r="2.5" /></I>
export const IconCal = (p: { size?: number }) => <I {...p}><rect x="3" y="5" width="18" height="16" rx="2" /><path d="M3 10h18M8 3v4M16 3v4" /></I>
export const IconArrow = (p: { size?: number }) => <I {...p}><path d="M19 12H5M11 6l-6 6 6 6" /></I>
export const IconHeart = ({ on, size = 18 }: { on: boolean; size?: number }) => (
  <svg width={size} height={size} viewBox="0 0 24 24" fill={on ? '#f43f5e' : 'none'} stroke={on ? '#f43f5e' : 'currentColor'} strokeWidth="1.8" aria-hidden><path d="M12 20s-7-4.4-9-9a5 5 0 0 1 9-3 5 5 0 0 1 9 3c-2 4.6-9 9-9 9z" /></svg>
)
export const IconCheck = (p: { size?: number }) => <I {...p}><path d="m5 12 5 5 9-10" /></I>
export const IconBolt = (p: { size?: number }) => <I {...p}><path d="M13 2 4 14h7l-1 8 9-12h-7z" /></I>
export const IconClock = (p: { size?: number }) => <I {...p}><circle cx="12" cy="12" r="9" /><path d="M12 7v5l3 2" /></I>
export const IconUpload = (p: { size?: number }) => <I {...p}><path d="M7 18a5 5 0 0 1-.5-10A6 6 0 0 1 18 9a4 4 0 0 1 0 9" /><path d="M12 12v8M9 15l3-3 3 3" /></I>
export const IconImage = (p: { size?: number }) => <I {...p}><rect x="3" y="4" width="18" height="16" rx="2" /><circle cx="9" cy="10" r="2" /><path d="m21 17-5-5-9 8" /></I>
export const IconCalc = (p: { size?: number }) => <I {...p}><rect x="5" y="3" width="14" height="18" rx="2" /><path d="M8 7h8M8 11h1M12 11h1M16 11h0M8 15h1M12 15h1M8 18h1M12 18h1M16 15v3" /></I>
export const IconPdf = (p: { size?: number }) => <I {...p}><path d="M14 3H7a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10a2 2 0 0 0 2-2V8z" /><path d="M14 3v5h5" /><text x="7.5" y="17" fontSize="5.5" fill="currentColor" stroke="none" fontWeight="700">PDF</text></I>

/** عمارة حديثة بنخيل — للبانر ولكروت المشاريع الي ما بيها صورة. `seed` يغيّر الشكل شوية. */
export function BuildingArt({ seed = 0, className = '' }: { seed?: number; className?: string }) {
  const floors = 3 + (seed % 2)
  const cols = 7 + (seed % 3)
  const sky = ['#f6b26b', '#9ec5fe', '#c4b5fd'][seed % 3]
  return (
    <svg viewBox="0 0 600 300" preserveAspectRatio="xMidYMid slice" className={className} aria-hidden>
      <defs>
        <linearGradient id={`sky${seed}`} x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#3b4f9c" /><stop offset=".7" stopColor={sky} /><stop offset="1" stopColor="#fde7c4" /></linearGradient>
        <linearGradient id={`glass${seed}`} x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#3a4f7a" /><stop offset="1" stopColor="#1c2742" /></linearGradient>
        <linearGradient id={`fade${seed}`} x1="0" y1="0" x2="0" y2="1"><stop offset=".55" stopColor="#0f1a36" stopOpacity="0" /><stop offset="1" stopColor="#0f1a36" stopOpacity=".95" /></linearGradient>
      </defs>
      <rect width="600" height="300" fill={`url(#sky${seed})`} />
      {/* المبنى */}
      <g>
        <rect x="120" y={230 - floors * 42} width="360" height={floors * 42 + 10} fill="#efe9df" />
        <rect x="112" y={224 - floors * 42} width="376" height="10" fill="#d8d0c2" />
        {Array.from({ length: floors }).map((_, f) =>
          Array.from({ length: cols }).map((__, c) => {
            const w = 340 / cols
            const lit = (f * 3 + c + seed) % 4 === 0
            return <rect key={`${f}-${c}`} x={130 + c * w} y={236 - floors * 42 + f * 42} width={w - 8} height="30" fill={lit ? '#ffd890' : `url(#glass${seed})`} opacity={lit ? 0.9 : 1} />
          }))}
        <rect x="270" y="200" width="60" height="40" fill="#ffd890" opacity=".85" />
      </g>
      {/* النخيل */}
      {[40, 90, 515, 565, 210, 400].map((x, i) => (
        <g key={x} transform={`translate(${x} ${i > 3 ? 238 : 240}) scale(${i > 3 ? 0.6 : 1})`}>
          <path d="M0 0 C 2 -40, -3 -70, 2 -100" stroke="#6b4f2a" strokeWidth="5" fill="none" />
          {[-60, -25, 15, 50, 85, 130].map((a) => <path key={a} d="M2 -100 q 20 -8 38 10" stroke="#2f6b3a" strokeWidth="6" fill="none" transform={`rotate(${a} 2 -100)`} />)}
        </g>
      ))}
      <rect y="240" width="600" height="60" fill="#3d5f3a" />
      <rect width="600" height="300" fill={`url(#fade${seed})`} />
    </svg>
  )
}

/** لفّات مخططات زرقة — بلوحة «خطوات إنشاء سيت». */
export function RollsArt({ className = '' }: { className?: string }) {
  return (
    <svg viewBox="0 0 200 120" className={className} aria-hidden>
      <path d="M20 95 L120 40 L185 70 L85 118 Z" fill="#dbe4ff" stroke="#93a8e8" />
      {[0, 1, 2, 3].map((i) => <path key={i} d={`M${40 + i * 18} ${88 - i * 10} l70 32`} stroke="#93a8e8" strokeWidth=".8" />)}
      {[[30, 60], [55, 45], [80, 72]].map(([x, y], i) => (
        <g key={i} transform={`translate(${x} ${y}) rotate(-28)`}>
          <rect width="90" height="20" rx="10" fill="#4f6fd8" />
          <rect width="90" height="20" rx="10" fill="url(#rollshine)" opacity=".5" />
          <ellipse cx="90" cy="10" rx="6" ry="10" fill="#c7d4ff" stroke="#4f6fd8" />
          <ellipse cx="90" cy="10" rx="2" ry="4" fill="#4f6fd8" />
        </g>
      ))}
      <defs><linearGradient id="rollshine" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#fff" /><stop offset="1" stopColor="#fff" stopOpacity="0" /></linearGradient></defs>
    </svg>
  )
}

/** خطوط بلوبرنت شفافة فوك البانر. */
export function BlueprintLines({ className = '' }: { className?: string }) {
  return (
    <svg viewBox="0 0 400 240" className={className} aria-hidden fill="none" stroke="#9fb6ff" strokeWidth=".8" opacity=".35">
      <path d="M60 220V90l90-50 90 50v130M150 40v180M60 140h180M90 220v-60h40v60M180 220v-70h40v70" />
      <path d="M260 220V120h110v100M260 150h110M300 120v100M330 120v100" />
      <path d="M20 225h370M30 230v-10M380 230v-10" />
      <circle cx="150" cy="95" r="14" />
    </svg>
  )
}
