// شعار الأماني: حرف A بقمّة كحلية وخط ذهبي — بأسلوب التصميم المقترح.
export function LogoMark({ size = 64, light = false }: { size?: number; light?: boolean }) {
  return (
    <svg width={size} height={size * 0.82} viewBox="0 0 100 82" aria-hidden>
      <path d="M50 2 L96 80 H72 L50 41 L28 80 H4 Z" fill={light ? '#fff' : '#0d3b8c'} />
      <path d="M58 54 L72 80 H58 L48 62 Z" fill="#f5b335" />
    </svg>
  )
}

export default function Logo({ light = false, size = 64, sub = true }: { light?: boolean; size?: number; sub?: boolean }) {
  return (
    <div className="flex flex-col items-center text-center">
      <LogoMark size={size} light={light} />
      <p className={`mt-2 text-[26px] leading-none font-black ${light ? 'text-white' : 'text-navy'}`}>شركة الأماني</p>
      {sub && <p className={`mt-1.5 text-[13px] font-bold tracking-[0.35em] ${light ? 'text-white/90' : 'text-navy'}`} dir="ltr">AL AMANI</p>}
    </div>
  )
}
