// الشعار الرسمي لشركة الأماني — نفس ملف الهوية المعتمد، مو شعار A التصميمي.
export function LogoMark({ size = 64, light = false }: { size?: number; light?: boolean }) {
  return (
    <span className={`grid place-items-center ${light ? 'rounded-2xl bg-white/95 shadow-[0_12px_30px_-12px_rgba(0,0,0,.6)]' : ''}`} style={{ width: size, height: size }}>
      <img src="/apple-touch-icon.png" width={size} height={size} alt="شعار شركة الأماني" className="h-full w-full object-contain" />
    </span>
  )
}

export default function Logo({ light = false, size = 64, sub = false }: { light?: boolean; size?: number; sub?: boolean }) {
  return (
    <div className="flex flex-col items-center text-center">
      <LogoMark size={size} light={light} />
      <p className={`mt-2 text-[26px] leading-none font-black ${light ? 'text-white' : 'text-navy'}`}>هندسة</p>
      {sub && <p className={`mt-1.5 text-[13px] font-bold tracking-[0.35em] ${light ? 'text-white/90' : 'text-navy'}`} dir="ltr">ENGINEERING</p>}
    </div>
  )
}
