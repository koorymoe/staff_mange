// مشهد البداية: فيلا حديثة بالغسق، ألواح شمسية على السطح، وكاميرا مراقبة —
// رسمة SVG بأسلوب التصميم (بدل صورة فوتوغرافية).
export default function HeroScene({ className = '' }: { className?: string }) {
  return (
    <svg viewBox="0 0 390 520" preserveAspectRatio="xMidYMax slice" className={className} aria-hidden>
      <defs>
        <linearGradient id="sky" x1="0" y1="0" x2="0" y2="1">
          <stop offset="0" stopColor="#0a2a6b" /><stop offset=".55" stopColor="#1f4fa8" /><stop offset="1" stopColor="#6d8fd0" />
        </linearGradient>
        <linearGradient id="wall" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#e9edf5" /><stop offset="1" stopColor="#c9d2e3" /></linearGradient>
        <linearGradient id="glass" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#ffd58a" /><stop offset="1" stopColor="#f5a524" /></linearGradient>
        <linearGradient id="panel" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stopColor="#1a3d8f" /><stop offset="1" stopColor="#0b1f4d" /></linearGradient>
        <linearGradient id="lawn" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stopColor="#2f6b45" /><stop offset="1" stopColor="#173b27" /></linearGradient>
      </defs>
      <rect width="390" height="520" fill="url(#sky)" />
      {[[40, 60], [90, 30], [300, 50], [350, 90], [200, 25], [250, 80], [140, 95]].map(([x, y], i) => <circle key={i} cx={x} cy={y} r={i % 3 ? 1.2 : 1.8} fill="#fff" opacity=".7" />)}
      <circle cx="320" cy="140" r="26" fill="#ffe7b0" opacity=".9" />
      {/* أشجار خلفية */}
      <path d="M0 360 Q40 300 80 350 Q110 290 150 345 L150 420 H0 Z" fill="#163a5f" opacity=".75" />
      <path d="M300 350 Q340 290 390 330 V420 H300 Z" fill="#163a5f" opacity=".75" />
      {/* الفيلا */}
      <g>
        <rect x="70" y="270" width="250" height="150" fill="url(#wall)" />
        <rect x="150" y="215" width="190" height="80" fill="#dfe5f0" />
        <rect x="60" y="262" width="270" height="12" fill="#2b3550" />
        <rect x="140" y="207" width="210" height="10" fill="#2b3550" />
        {/* ألواح شمسية */}
        <g transform="translate(160 188) skewX(-18)">
          {[0, 1, 2, 3].map((i) => <rect key={i} x={i * 42} y="0" width="38" height="20" rx="2" fill="url(#panel)" stroke="#5b86d6" strokeWidth="1" />)}
          {[0, 1, 2, 3].map((i) => <path key={'l' + i} d={`M${i * 42 + 13} 0 V20 M${i * 42 + 26} 0 V20 M${i * 42} 10 H${i * 42 + 38}`} stroke="#5b86d6" strokeWidth=".7" />)}
        </g>
        {/* شبابيك مضيئة */}
        <rect x="165" y="228" width="70" height="52" fill="url(#glass)" opacity=".95" />
        <rect x="250" y="228" width="75" height="52" fill="url(#glass)" opacity=".85" />
        <rect x="88" y="296" width="80" height="100" fill="url(#glass)" />
        <rect x="190" y="300" width="58" height="120" fill="#3a2a1a" />
        <rect x="262" y="296" width="44" height="70" fill="url(#glass)" opacity=".9" />
        <path d="M128 296 V396 M88 346 H168 M200 228 V280 M287 228 V280" stroke="#2b3550" strokeWidth="3" />
        {/* كاميرا على الجدار */}
        <g transform="translate(300 300)">
          <rect x="-4" y="-2" width="8" height="14" rx="2" fill="#f3f5f9" />
          <rect x="-26" y="-12" width="30" height="16" rx="7" fill="#fbfcfe" stroke="#9aa6bd" />
          <circle cx="-20" cy="-4" r="4.5" fill="#0b1f4d" /><circle cx="-20" cy="-4" r="1.6" fill="#5b86d6" />
        </g>
      </g>
      {/* حديقة */}
      <rect x="0" y="410" width="390" height="110" fill="url(#lawn)" />
      <path d="M0 418 H390" stroke="#4c8a5f" strokeWidth="3" />
      <rect x="180" y="420" width="70" height="100" fill="#cfd6e3" opacity=".6" />
      {[30, 345].map((x) => <g key={x}><rect x={x - 3} y="360" width="6" height="55" fill="#3b2a1a" /><circle cx={x} cy="352" r="26" fill="#245a3a" /></g>)}
      <rect width="390" height="520" fill="url(#sky)" opacity=".18" />
    </svg>
  )
}
