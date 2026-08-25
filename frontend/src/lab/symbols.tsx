// ═══ الرموز التخطيطية ═══
//
// ⚠️ **رموز مو صور.** MATLAB وPacket Tracer الاثنان يستعملون رموزاً
// مسطّحة — لأن الرمز يُقرا بلمحة وهو الي يهم بمخطط فيه ٤٠ قطعة.
// الصورة الفوتوغرافية تشوّش المخطط بدل ما تفيده.
//
// كل رمز يتقيّد بصندوق (0,0,w,h) الجاي من الكتالوگ، فالمنافذ
// (نسبة ٠..١) تنطبق فوگه بلا حساب إضافي.

interface SymProps {
  w: number
  h: number
  /** لون التمييز لمن تكون القطعة محدّدة أو شغّالة. */
  accent: string
  live?: boolean
  params: Record<string, string | number | boolean>
}

const STROKE = '#94a3b8'
const FILL = '#111827'

export function Symbol({ symbol, ...p }: SymProps & { symbol: string }) {
  const S = SYMBOLS[symbol]
  return S ? <S {...p} /> : <GenericBox {...p} />
}

function GenericBox({ w, h, accent }: SymProps) {
  return <rect x={0} y={0} width={w} height={h} rx={6} fill={FILL} stroke={accent} strokeWidth={1.5} />
}

// ═══ الدوائر ═══
const Battery = ({ w, h, accent }: SymProps) => (
  <g>
    <rect x={0} y={0} width={w} height={h} rx={6} fill={FILL} stroke={accent} strokeWidth={1.5} />
    {/* خطّان طويل وقصير — الرمز المتعارف عليه للبطارية */}
    <line x1={w * 0.34} y1={h * 0.18} x2={w * 0.34} y2={h * 0.82} stroke={STROKE} strokeWidth={3} />
    <line x1={w * 0.5} y1={h * 0.32} x2={w * 0.5} y2={h * 0.68} stroke={STROKE} strokeWidth={6} />
    <text x={w * 0.2} y={h * 0.3} fontSize={11} fill="#f87171">+</text>
    <text x={w * 0.2} y={h * 0.8} fontSize={11} fill="#93c5fd">−</text>
  </g>
)

const Resistor = ({ w, h, accent }: SymProps) => (
  <g>
    <rect x={0} y={0} width={w} height={h} rx={4} fill={FILL} stroke={accent} strokeWidth={1.5} />
    {/* زگزاگ — رمز المقاومة */}
    <polyline
      points={`${w * 0.14},${h / 2} ${w * 0.24},${h * 0.22} ${w * 0.36},${h * 0.78} ${w * 0.48},${h * 0.22} ${w * 0.6},${h * 0.78} ${w * 0.72},${h * 0.22} ${w * 0.82},${h / 2} ${w * 0.86},${h / 2}`}
      fill="none" stroke={STROKE} strokeWidth={2} strokeLinejoin="round"
    />
  </g>
)

const Lamp = ({ w, h, accent, live }: SymProps) => (
  <g>
    <circle cx={w / 2} cy={h / 2} r={Math.min(w, h) * 0.34} fill={live ? '#fbbf24' : FILL} stroke={accent} strokeWidth={1.5} />
    <line x1={w * 0.31} y1={h * 0.31} x2={w * 0.69} y2={h * 0.69} stroke={live ? '#78350f' : STROKE} strokeWidth={1.6} />
    <line x1={w * 0.69} y1={h * 0.31} x2={w * 0.31} y2={h * 0.69} stroke={live ? '#78350f' : STROKE} strokeWidth={1.6} />
    {live && <circle cx={w / 2} cy={h / 2} r={Math.min(w, h) * 0.46} fill="#fbbf24" opacity={0.18} />}
  </g>
)

const Switch = ({ w, h, accent, params }: SymProps) => {
  const closed = !!params.closed
  return (
    <g>
      <rect x={0} y={0} width={w} height={h} rx={4} fill={FILL} stroke={accent} strokeWidth={1.5} />
      <circle cx={w * 0.22} cy={h / 2} r={3} fill={STROKE} />
      <circle cx={w * 0.78} cy={h / 2} r={3} fill={STROKE} />
      {/* الذراع يميل لمن ينفتح — يُقرا من بعيد */}
      <line
        x1={w * 0.22} y1={h / 2}
        x2={w * 0.78} y2={closed ? h / 2 : h * 0.2}
        stroke={closed ? '#4ade80' : '#f87171'} strokeWidth={2.4} strokeLinecap="round"
      />
    </g>
  )
}

const Fuse = ({ w, h, accent }: SymProps) => (
  <g>
    <rect x={0} y={0} width={w} height={h} rx={4} fill={FILL} stroke={accent} strokeWidth={1.5} />
    <rect x={w * 0.2} y={h * 0.28} width={w * 0.6} height={h * 0.44} rx={3} fill="none" stroke={STROKE} strokeWidth={1.6} />
    <line x1={w * 0.2} y1={h / 2} x2={w * 0.8} y2={h / 2} stroke={STROKE} strokeWidth={1.6} />
  </g>
)

const Motor = ({ w, h, accent }: SymProps) => (
  <g>
    <circle cx={w / 2} cy={h / 2} r={Math.min(w, h) * 0.36} fill={FILL} stroke={accent} strokeWidth={1.5} />
    <text x={w / 2} y={h / 2 + 5} fontSize={15} fill={STROKE} textAnchor="middle" fontWeight="bold">M</text>
  </g>
)

// ═══ الطاقة الشمسية ═══
const Pv = ({ w, h, accent, live }: SymProps) => (
  <g>
    <rect x={0} y={0} width={w * 0.8} height={h} rx={4} fill={live ? '#1e3a5f' : FILL} stroke={accent} strokeWidth={1.5} />
    {[0.2, 0.4, 0.6, 0.8].map((f) => (
      <line key={f} x1={w * 0.8 * f} y1={0} x2={w * 0.8 * f} y2={h} stroke={STROKE} strokeWidth={0.8} opacity={0.6} />
    ))}
    {[0.33, 0.66].map((f) => (
      <line key={f} x1={0} y1={h * f} x2={w * 0.8} y2={h * f} stroke={STROKE} strokeWidth={0.8} opacity={0.6} />
    ))}
    {live && <circle cx={w * 0.18} cy={h * 0.2} r={5} fill="#fbbf24" opacity={0.8} />}
  </g>
)

const Inverter = ({ w, h, accent, live }: SymProps) => (
  <g>
    <rect x={0} y={0} width={w} height={h} rx={7} fill={FILL} stroke={accent} strokeWidth={1.5} />
    {/* الرمز المتعارف: مربّع بقطر، DC على جهة وموجة AC على الجهة الثانية */}
    <line x1={w * 0.28} y1={h * 0.75} x2={w * 0.72} y2={h * 0.25} stroke={STROKE} strokeWidth={1.6} />
    <line x1={w * 0.3} y1={h * 0.33} x2={w * 0.44} y2={h * 0.33} stroke={STROKE} strokeWidth={2} />
    <path d={`M ${w * 0.56} ${h * 0.68} q ${w * 0.05} ${-h * 0.12} ${w * 0.1} 0 q ${w * 0.05} ${h * 0.12} ${w * 0.1} 0`}
      fill="none" stroke={STROKE} strokeWidth={2} />
    {live && <circle cx={w * 0.5} cy={h * 0.9} r={3.5} fill="#4ade80" />}
  </g>
)

const BatteryBank = ({ w, h, accent, params }: SymProps) => {
  const soc = Math.max(0, Math.min(100, Number(params.soc ?? 0)))
  return (
    <g>
      <rect x={0} y={0} width={w * 0.85} height={h} rx={5} fill={FILL} stroke={accent} strokeWidth={1.5} />
      <rect x={w * 0.85} y={h * 0.35} width={w * 0.06} height={h * 0.3} fill={STROKE} />
      <rect x={4} y={4} width={(w * 0.85 - 8) * (soc / 100)} height={h - 8} rx={3}
        fill={soc > 50 ? '#22c55e' : soc > 20 ? '#f59e0b' : '#ef4444'} opacity={0.55} />
      <text x={w * 0.42} y={h / 2 + 5} fontSize={13} fill="#e2e8f0" textAnchor="middle" fontWeight="bold">{soc}%</text>
    </g>
  )
}

const Load = ({ w, h, accent, live }: SymProps) => (
  <g>
    <rect x={0} y={0} width={w} height={h} rx={6} fill={FILL} stroke={accent} strokeWidth={1.5} />
    <path d={`M ${w * 0.28} ${h * 0.62} L ${w * 0.5} ${h * 0.24} L ${w * 0.5} ${h * 0.48} L ${w * 0.72} ${h * 0.48} L ${w * 0.5} ${h * 0.82} L ${w * 0.5} ${h * 0.62} Z`}
      fill={live ? '#fbbf24' : 'none'} stroke={STROKE} strokeWidth={1.6} strokeLinejoin="round" />
  </g>
)

// ═══ الشبكات ═══
const NetSwitch = ({ w, h, accent, live }: SymProps) => (
  <g>
    <rect x={0} y={0} width={w} height={h} rx={5} fill={FILL} stroke={accent} strokeWidth={1.5} />
    {/* أربعة أسهم متقاطعة — الرمز المتعارف للسويچ */}
    <g stroke={STROKE} strokeWidth={1.6} fill="none">
      <path d={`M ${w * 0.3} ${h * 0.34} L ${w * 0.7} ${h * 0.34}`} markerEnd="" />
      <path d={`M ${w * 0.64} ${h * 0.28} L ${w * 0.7} ${h * 0.34} L ${w * 0.64} ${h * 0.4}`} />
      <path d={`M ${w * 0.7} ${h * 0.62} L ${w * 0.3} ${h * 0.62}`} />
      <path d={`M ${w * 0.36} ${h * 0.56} L ${w * 0.3} ${h * 0.62} L ${w * 0.36} ${h * 0.68}`} />
    </g>
    {live && <circle cx={w * 0.08} cy={h * 0.16} r={3} fill="#4ade80" />}
  </g>
)

const NetRouter = ({ w, h, accent, live }: SymProps) => (
  <g>
    <ellipse cx={w / 2} cy={h / 2} rx={w * 0.44} ry={h * 0.38} fill={FILL} stroke={accent} strokeWidth={1.5} />
    <g stroke={STROKE} strokeWidth={1.6} fill="none">
      <path d={`M ${w * 0.3} ${h * 0.42} L ${w * 0.7} ${h * 0.42}`} />
      <path d={`M ${w * 0.64} ${h * 0.36} L ${w * 0.7} ${h * 0.42} L ${w * 0.64} ${h * 0.48}`} />
      <path d={`M ${w * 0.7} ${h * 0.6} L ${w * 0.3} ${h * 0.6}`} />
      <path d={`M ${w * 0.36} ${h * 0.54} L ${w * 0.3} ${h * 0.6} L ${w * 0.36} ${h * 0.66}`} />
    </g>
    {live && <circle cx={w * 0.12} cy={h * 0.2} r={3} fill="#4ade80" />}
  </g>
)

const NetPc = ({ w, h, accent, live }: SymProps) => (
  <g>
    <rect x={w * 0.1} y={0} width={w * 0.8} height={h * 0.62} rx={3} fill={FILL} stroke={accent} strokeWidth={1.5} />
    <rect x={w * 0.16} y={h * 0.08} width={w * 0.68} height={h * 0.4} fill={live ? '#0f3d2e' : '#0b1220'} />
    <rect x={w * 0.3} y={h * 0.66} width={w * 0.4} height={h * 0.08} fill={STROKE} />
    <rect x={w * 0.18} y={h * 0.76} width={w * 0.64} height={h * 0.07} rx={2} fill={STROKE} />
  </g>
)

const NetCam = ({ w, h, accent, live }: SymProps) => (
  <g>
    <rect x={w * 0.12} y={h * 0.22} width={w * 0.56} height={h * 0.4} rx={5} fill={FILL} stroke={accent} strokeWidth={1.5} />
    <path d={`M ${w * 0.68} ${h * 0.32} L ${w * 0.9} ${h * 0.22} L ${w * 0.9} ${h * 0.62} L ${w * 0.68} ${h * 0.52} Z`}
      fill={FILL} stroke={accent} strokeWidth={1.5} />
    <line x1={w * 0.4} y1={h * 0.62} x2={w * 0.4} y2={h * 0.82} stroke={STROKE} strokeWidth={2} />
    {live && <circle cx={w * 0.22} cy={h * 0.32} r={3} fill="#f87171" />}
  </g>
)

// ═══ إنذار الحريق ═══
const FirePanel = ({ w, h, accent, live }: SymProps) => (
  <g>
    <rect x={0} y={0} width={w} height={h} rx={6} fill="#3f1616" stroke={accent} strokeWidth={1.5} />
    <rect x={w * 0.12} y={h * 0.1} width={w * 0.76} height={h * 0.3} rx={3} fill="#0b1220" />
    {/* شريط مؤشرات — أحمر/أصفر/أخضر مثل اللوحة الحقيقية */}
    {[['#ef4444', 0.2], ['#f59e0b', 0.34], [live ? '#22c55e' : '#334155', 0.48]].map(([c, x], i) => (
      <circle key={i} cx={w * (x as number)} cy={h * 0.55} r={4} fill={c as string} />
    ))}
    <text x={w / 2} y={h * 0.86} fontSize={11} fill="#fca5a5" textAnchor="middle" fontWeight="bold">FACP</text>
  </g>
)

const Detector = ({ w, h, accent, live }: SymProps) => (
  <g>
    <circle cx={w / 2} cy={h / 2} r={Math.min(w, h) * 0.38} fill={FILL} stroke={accent} strokeWidth={1.5} />
    <circle cx={w / 2} cy={h / 2} r={Math.min(w, h) * 0.22} fill="none" stroke={STROKE} strokeWidth={1.2} />
    <circle cx={w / 2} cy={h / 2} r={3} fill={live ? '#ef4444' : '#475569'} />
  </g>
)

const Mcp = ({ w, h, accent }: SymProps) => (
  <g>
    <rect x={w * 0.16} y={h * 0.16} width={w * 0.68} height={h * 0.68} rx={4} fill="#7f1d1d" stroke={accent} strokeWidth={1.5} />
    <rect x={w * 0.28} y={h * 0.3} width={w * 0.44} height={h * 0.24} rx={2} fill="#e2e8f0" opacity={0.85} />
    <text x={w / 2} y={h * 0.74} fontSize={9} fill="#fecaca" textAnchor="middle" fontWeight="bold">MCP</text>
  </g>
)

const Sounder = ({ w, h, accent, live }: SymProps) => (
  <g>
    {/* بوق — مثلث مع موجات صوت */}
    <path d={`M ${w * 0.18} ${h * 0.38} L ${w * 0.42} ${h * 0.38} L ${w * 0.62} ${h * 0.18} L ${w * 0.62} ${h * 0.82} L ${w * 0.42} ${h * 0.62} L ${w * 0.18} ${h * 0.62} Z`}
      fill={live ? '#ef4444' : FILL} stroke={accent} strokeWidth={1.5} strokeLinejoin="round" />
    {[0.72, 0.84].map((f, i) => (
      <path key={i} d={`M ${w * f} ${h * (0.34 + i * 0.04)} Q ${w * (f + 0.08)} ${h * 0.5} ${w * f} ${h * (0.66 - i * 0.04)}`}
        fill="none" stroke={live ? '#fca5a5' : STROKE} strokeWidth={1.4} />
    ))}
  </g>
)

const Eol = ({ w, h, accent }: SymProps) => (
  <g>
    <rect x={w * 0.1} y={h * 0.28} width={w * 0.8} height={h * 0.44} rx={4} fill={FILL} stroke={accent} strokeWidth={1.5} />
    <polyline points={`${w * 0.2},${h / 2} ${w * 0.32},${h * 0.32} ${w * 0.46},${h * 0.68} ${w * 0.6},${h * 0.32} ${w * 0.72},${h * 0.68} ${w * 0.82},${h / 2}`}
      fill="none" stroke="#fbbf24" strokeWidth={1.8} strokeLinejoin="round" />
  </g>
)

// ═══ الصوت ═══
const Amplifier = ({ w, h, accent, live, params }: SymProps) => (
  <g>
    <rect x={0} y={0} width={w} height={h} rx={5} fill={FILL} stroke={accent} strokeWidth={1.5} />
    {/* مثلث التكبير — الرمز المتعارف */}
    <path d={`M ${w * 0.2} ${h * 0.25} L ${w * 0.6} ${h * 0.5} L ${w * 0.2} ${h * 0.75} Z`}
      fill="none" stroke={STROKE} strokeWidth={1.8} strokeLinejoin="round" />
    <text x={w * 0.76} y={h * 0.44} fontSize={10} fill="#94a3b8" textAnchor="middle">
      {String(params.mode ?? '100v') === '100v' ? '100V' : 'LOW-Z'}
    </text>
    <text x={w * 0.76} y={h * 0.66} fontSize={10} fill="#64748b" textAnchor="middle">{String(params.pRated ?? '')}W</text>
    {live && <circle cx={w * 0.08} cy={h * 0.12} r={3} fill="#4ade80" />}
  </g>
)

const SpeakerSym = ({ w, h, accent, live }: SymProps) => (
  <g>
    <rect x={w * 0.18} y={h * 0.3} width={w * 0.18} height={h * 0.4} fill={FILL} stroke={accent} strokeWidth={1.5} />
    <path d={`M ${w * 0.36} ${h * 0.3} L ${w * 0.62} ${h * 0.12} L ${w * 0.62} ${h * 0.88} L ${w * 0.36} ${h * 0.7} Z`}
      fill={live ? '#334155' : FILL} stroke={accent} strokeWidth={1.5} strokeLinejoin="round" />
    {live && [0.7, 0.82].map((f, i) => (
      <path key={i} d={`M ${w * f} ${h * (0.3 + i * 0.05)} Q ${w * (f + 0.07)} ${h * 0.5} ${w * f} ${h * (0.7 - i * 0.05)}`}
        fill="none" stroke="#4ade80" strokeWidth={1.4} />
    ))}
  </g>
)

const HornSym = ({ w, h, accent, live }: SymProps) => (
  <g>
    <path d={`M ${w * 0.14} ${h * 0.38} L ${w * 0.38} ${h * 0.38} L ${w * 0.72} ${h * 0.1} L ${w * 0.72} ${h * 0.9} L ${w * 0.38} ${h * 0.62} L ${w * 0.14} ${h * 0.62} Z`}
      fill={live ? '#334155' : FILL} stroke={accent} strokeWidth={1.5} strokeLinejoin="round" />
    <line x1={w * 0.72} y1={h * 0.1} x2={w * 0.72} y2={h * 0.9} stroke={STROKE} strokeWidth={1.4} />
  </g>
)

const MicSym = ({ w, h, accent }: SymProps) => (
  <g>
    <rect x={w * 0.36} y={h * 0.12} width={w * 0.28} height={h * 0.42} rx={w * 0.14} fill={FILL} stroke={accent} strokeWidth={1.5} />
    <path d={`M ${w * 0.24} ${h * 0.46} a ${w * 0.26} ${h * 0.26} 0 0 0 ${w * 0.52} 0`} fill="none" stroke={STROKE} strokeWidth={1.6} />
    <line x1={w / 2} y1={h * 0.72} x2={w / 2} y2={h * 0.88} stroke={STROKE} strokeWidth={1.6} />
  </g>
)

const SYMBOLS: Record<string, (p: SymProps) => React.ReactElement> = {
  amplifier: Amplifier,
  speaker: SpeakerSym,
  horn: HornSym,
  mic: MicSym,
  fire_panel: FirePanel,
  detector: Detector,
  mcp: Mcp,
  sounder: Sounder,
  eol: Eol,
  battery: Battery,
  resistor: Resistor,
  lamp: Lamp,
  switch: Switch,
  fuse: Fuse,
  motor: Motor,
  pv: Pv,
  inverter: Inverter,
  battery_bank: BatteryBank,
  load: Load,
  net_switch: NetSwitch,
  net_router: NetRouter,
  net_pc: NetPc,
  net_cam: NetCam,
}
