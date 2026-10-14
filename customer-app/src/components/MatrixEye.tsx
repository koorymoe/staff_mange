// عين ماتركس — هوية المساعد الذكي بالتطبيق
export default function MatrixEye({ size = 40, busy = false }: { size?: number; busy?: boolean }) {
  return (
    <svg width={size} height={size} viewBox="0 0 64 64" aria-hidden className="shrink-0">
      <defs><radialGradient id="mx-iris" cx="50%" cy="50%" r="50%"><stop offset="0%" stopColor="#ff4d4d" /><stop offset="100%" stopColor="#9b0000" /></radialGradient></defs>
      <circle cx="32" cy="32" r="31" fill="#0d3b8c" />
      <g className="blink">
        <path d="M7 32 Q32 11 57 32 Q32 53 7 32 Z" fill="#fff" />
        <circle cx="32" cy="32" r="11.5" fill="url(#mx-iris)" />
        <g style={{ transformOrigin: '32px 32px', animation: `spin ${busy ? '1s' : '6s'} linear infinite` }}>
          {[0, 120, 240].map((a) => <path key={a} transform={`rotate(${a} 32 32)`} d="M32 23 a2.5 2.5 0 1 1 -0.1 0 q3 1 2 4" fill="#111" />)}
        </g>
        <circle cx="32" cy="32" r="3" fill="#111" />
      </g>
    </svg>
  )
}
