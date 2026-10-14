export default function Logo({ size = 44 }: { size?: number }) {
  return <img src="/favicon.png" alt="الأماني" width={size} height={size} className="rounded-2xl bg-white object-contain shadow-sm" />
}
