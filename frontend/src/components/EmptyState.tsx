/**
 * حالة الفراغ — وأهم شي بيها `reason`.
 *
 * چانت شاشة «تقييم الأداء» تكول «ماكو حجوزات بهذي التصفية» حتى لمن
 * تكون التصفية فارغة والسبب الحقيقي شي ثاني تماماً. رسالة فراغ ما
 * تكول السبب تخلي المستخدم يدوّر بالمكان الغلط.
 */
export default function EmptyState({
  icon = '📭', title, reason, action,
}: {
  icon?: React.ReactNode
  title: string
  /** ليش فارغة — مو شنو الفارغ. */
  reason?: string
  action?: React.ReactNode
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 px-6 py-12 text-center">
      <span className="text-4xl">{icon}</span>
      <p className="text-sm font-bold" style={{ color: 'var(--t-body)' }}>{title}</p>
      {reason && <p className="max-w-md text-xs" style={{ color: 'var(--t-faint)' }}>{reason}</p>}
      {action && <div className="mt-2">{action}</div>}
    </div>
  )
}
