import { imgServiceUrl } from '../api'

// ═══ رسمة الخدمة ═══
// إذا المالك رافع صورة للخدمة تطلع هي. وإذا لا، رسمة SVG أنيقة حسب نوع
// الخدمة (كاميرا، شبكة، طاقة شمسية…) على خلفية متدرجة — بلا صور وهمية.

type Kind = 'camera' | 'network' | 'solar' | 'smart' | 'gps' | 'alarm' | 'sound' | 'power' | 'lock' | 'phone' | 'tv' | 'build' | 'tool'

const RULES: [RegExp, Kind][] = [
  [/كامير|camera|ip|انلوك|مراقب/i, 'camera'], [/شبك|انترنت|network|ستلايت|wifi/i, 'network'], [/شمس|طاقة|solar/i, 'solar'],
  [/ذكي|smart|بيوت/i, 'smart'], [/gps|جي بي|تتبع/i, 'gps'], [/حريق|انذار|إنذار/i, 'alarm'], [/صوت|sound|سماع/i, 'sound'],
  [/كهرب|power/i, 'power'], [/اقفال|أقفال|قفل|بصمة/i, 'lock'], [/تليفون|بدال|هاتف|حاكي/i, 'phone'], [/شاش|tv/i, 'tv'],
  [/حداد|نجار|جبس|صباغ|سيرام|تاسيس|تأسيس|ديكور|لبخ|بناء/i, 'build'],
]
const kindOf = (name: string): Kind => RULES.find(([re]) => re.test(name))?.[1] ?? 'tool'

const TONES: Record<Kind, [string, string, string]> = {
  camera: ['#e8eefb', '#cddcf7', '#0d3b8c'], network: ['#e9f0fb', '#d3e1f6', '#123f9c'], solar: ['#fff4dc', '#fde5ad', '#c47a00'],
  smart: ['#e7f6f1', '#c9ecdf', '#0b7a5a'], gps: ['#e8f1ff', '#cfe0ff', '#1d4ed8'], alarm: ['#fdeceb', '#f9d2cf', '#c0262d'],
  sound: ['#eef0f6', '#d9deea', '#334155'], power: ['#fff6dd', '#fde8a6', '#b26b00'], lock: ['#eef1fb', '#d8def6', '#1e3a8a'],
  phone: ['#ecf6fb', '#d0e9f6', '#0369a1'], tv: ['#eef0fb', '#dcdff7', '#3730a3'], build: ['#f6f1ea', '#eadfce', '#8a5a2b'], tool: ['#eef2f8', '#dbe3ef', '#0d3b8c'],
}

function Icon({ kind, c }: { kind: Kind; c: string }) {
  const s = { fill: 'none', stroke: c, strokeWidth: 3.2, strokeLinecap: 'round' as const, strokeLinejoin: 'round' as const }
  switch (kind) {
    case 'camera': return <g><rect x="18" y="34" width="52" height="26" rx="10" fill="#fff" stroke={c} strokeWidth="3.2" /><circle cx="56" cy="47" r="8" fill={c} /><circle cx="56" cy="47" r="3" fill="#fff" /><path d="M26 60 L22 78 H40" {...s} /><path d="M70 41 L82 36 V58 L70 53" {...s} /></g>
    case 'network': return <g><rect x="18" y="52" width="64" height="20" rx="6" fill="#fff" stroke={c} strokeWidth="3.2" /><path d="M30 52 L26 24 M50 52 V20 M70 52 L74 24" {...s} /><circle cx="30" cy="62" r="2.5" fill={c} /><circle cx="40" cy="62" r="2.5" fill={c} /><circle cx="50" cy="62" r="2.5" fill="#22c55e" /></g>
    case 'solar': return <g><circle cx="74" cy="22" r="8" fill="#f5b335" /><path d="M14 66 L28 34 H86 L72 66 Z" fill="#1e3a8a" stroke={c} strokeWidth="2" /><path d="M33 34 L23 66 M52 34 L42 66 M71 34 L61 66 M21 50 H79" stroke="#93c5fd" strokeWidth="1.6" /><path d="M48 66 V80 M36 80 H60" {...s} /></g>
    case 'smart': return <g><path d="M18 46 L50 20 L82 46 V78 H18 Z" fill="#fff" stroke={c} strokeWidth="3.2" strokeLinejoin="round" /><path d="M38 56 a14 14 0 0 1 24 0 M43 62 a7 7 0 0 1 14 0" {...s} /><circle cx="50" cy="68" r="3" fill={c} /></g>
    case 'gps': return <g><path d="M50 82 C50 82 24 56 24 38 a26 26 0 0 1 52 0 C76 56 50 82 50 82 Z" fill={c} /><circle cx="50" cy="38" r="10" fill="#fff" /></g>
    case 'alarm': return <g><rect x="22" y="22" width="56" height="56" rx="14" fill={c} /><circle cx="50" cy="50" r="15" fill="#fff" /><circle cx="50" cy="50" r="8" fill={c} opacity=".85" /></g>
    case 'sound': return <g><path d="M22 40 H36 L56 24 V76 L36 60 H22 Z" fill={c} /><path d="M66 36 a18 18 0 0 1 0 28 M74 28 a30 30 0 0 1 0 44" {...s} /></g>
    case 'power': return <g><rect x="28" y="18" width="44" height="64" rx="8" fill="#fff" stroke={c} strokeWidth="3.2" /><path d="M54 28 L40 52 H52 L46 72 L62 44 H50 Z" fill="#f5b335" stroke={c} strokeWidth="1.6" strokeLinejoin="round" /></g>
    case 'lock': return <g><rect x="26" y="44" width="48" height="36" rx="8" fill={c} /><path d="M36 44 V34 a14 14 0 0 1 28 0 V44" {...s} /><circle cx="50" cy="60" r="5" fill="#fff" /></g>
    case 'phone': return <g><rect x="34" y="16" width="32" height="68" rx="8" fill="#fff" stroke={c} strokeWidth="3.2" /><rect x="40" y="24" width="20" height="40" rx="3" fill={c} opacity=".15" /><circle cx="50" cy="74" r="3" fill={c} /></g>
    case 'tv': return <g><rect x="16" y="22" width="68" height="44" rx="6" fill={c} /><rect x="21" y="27" width="58" height="34" rx="3" fill="#93c5fd" opacity=".6" /><path d="M40 78 H60 M50 66 V78" {...s} /></g>
    case 'build': return <g><path d="M22 78 V44 L50 26 L78 44 V78" fill="#fff" stroke={c} strokeWidth="3.2" strokeLinejoin="round" /><rect x="42" y="56" width="16" height="22" fill={c} /><path d="M14 46 L50 22 L86 46" {...s} /></g>
    default: return <g><path d="M62 22 a14 14 0 0 0 -16 18 L22 64 a7 7 0 0 0 10 10 L56 50 a14 14 0 0 0 18 -16 l-9 9 -8 -2 -2 -8 Z" fill={c} /></g>
  }
}

// صورة افتراضية حقيقية لكل نوع (CC0 — public/img/CREDITS.md)
// الصور الأصلية الجديدة تتقدم على المكتبة العامة، والباقي يبقى له صورة مخصصة.
const photoOf = (name: string) => {
  const kind = kindOf(name)
  if (kind === 'camera') return '/img/camera-premium.png'
  if (kind === 'gps') return '/img/gps-premium.png'
  if (kind === 'solar') return '/img/solar-premium.png'
  if (kind === 'network') return '/img/network-premium.png'
  if (kind === 'smart') return '/img/smart-premium.png'
  return `/img/${kind}.webp`
}

export default function ServiceArt({ id, name, hasImage, className = '', iconOnly = false, badge = false }: { id?: string; name: string; hasImage?: boolean; className?: string; iconOnly?: boolean; badge?: boolean }) {
  const k = kindOf(name)
  const [a, b, c] = TONES[k]
  if (badge) {
    // أيقونة الخدمات الرئيسية: دائرة ملوّنة متدرجة وأيقونة مجسّمة (مثل التصميم)
    return (
      <span className={`grid place-items-center rounded-2xl ${className}`} style={{ background: `linear-gradient(145deg, ${a}, ${b})`, boxShadow: `inset 0 -3px 0 ${b}, 0 6px 14px -8px ${c}` }}>
        <svg viewBox="0 0 100 100" className="h-[70%] w-[70%] drop-shadow-[0_3px_3px_rgba(15,29,58,.25)]" aria-hidden><Icon kind={k} c={c} /></svg>
      </span>
    )
  }
  if (iconOnly) return <svg viewBox="0 0 100 100" className={className} aria-hidden><Icon kind={k} c={c} /></svg>
  return <img src={hasImage && id ? imgServiceUrl(id) : photoOf(name)} alt={name} loading="lazy" className={`service-photo bg-[#e9eef7] object-cover ${className}`} />
}
