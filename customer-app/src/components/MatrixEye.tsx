// عين ماتركس — الشارنغان الصغيرة الي ترمش، هوية المساعد بالتطبيق.
export default function MatrixEye({ size = 40, busy = false }: { size?: number; busy?: boolean }) {
  return (
    <svg width={size} height={size} viewBox="0 0 64 64" aria-hidden className="shrink-0">
      <defs>
        <radialGradient id="mx-iris" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="#ff4d4d" />
          <stop offset="100%" stopColor="#9b0000" />
        </radialGradient>
      </defs>
      <circle cx="32" cy="32" r="31" fill="#0f172a" />
      <g className="blink">
        <path d="M6 32 Q32 10 58 32 Q32 54 6 32 Z" fill="#fff" />
        <circle cx="32" cy="32" r="12" fill="url(#mx-iris)" />
        <g style={{ transformOrigin: '32px 32px', animation: `spin ${busy ? '1s' : '6s'} linear infinite` }}>
          {[0, 120, 240].map((a) => (
            <path key={a} transform={`rotate(${a} 32 32)`} d="M32 22.5 a2.6 2.6 0 1 1 -0.1 0 q3 1 2 4" fill="#111" />
          ))}
        </g>
        <circle cx="32" cy="32" r="3.2" fill="#111" />
        <circle cx="32" cy="32" r="8" fill="none" stroke="#111" strokeOpacity=".5" strokeWidth=".8" />
      </g>
      <style>{'@keyframes spin{to{transform:rotate(360deg)}}'}</style>
    </svg>
  )
}
