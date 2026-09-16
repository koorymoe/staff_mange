/**
 * بطاقة رقم واحدة — الشكل الي اختاره صاحب النظام بصورة «تتبع المهام».
 *
 * انبنت لأن نفس البطاقة چانت مكتوبة ثلاث مرات بثلاث ملفات
 * (Dashboard وPerformanceReviewPage وKpiPage) وكل وحدة بألوان مكتوبة
 * بالكود، فولا وحدة منهن تشتغل بالوضع الليلي. الرموز جوّا المكوّن
 * حتى الوضع الليلي يجي مجاناً وين ما انستعمل.
 */

/** النبرات مربوطة برموز index.css مو بألوان مكتوبة بالكود. */
const TONES = {
  default: { tint: 'var(--sf-sunken)', ink: 'var(--t-title)' },
  info: { tint: 'var(--sf-info)', ink: 'var(--t-info)' },
  success: { tint: 'var(--sf-success)', ink: 'var(--t-success)' },
  danger: { tint: 'var(--sf-danger)', ink: 'var(--t-danger)' },
  warning: { tint: 'var(--tint-warning)', ink: 'var(--t-warning)' },
  violet: { tint: 'var(--sf-violet)', ink: 'var(--t-violet)' },
} as const

export type StatTone = keyof typeof TONES

export default function StatTile({
  label, value, hint, icon, tone = 'default', onClick,
}: {
  label: string
  value: React.ReactNode
  hint?: string
  icon?: React.ReactNode
  tone?: StatTone
  /** لو انعطى، البطاقة تصير زر يودّي لشغلها. */
  onClick?: () => void
}) {
  const t = TONES[tone]
  const body = (
    <>
      <div className="flex items-start justify-between gap-2">
        <p className="text-xs font-medium" style={{ color: 'var(--t-muted)' }}>{label}</p>
        {icon && (
          <span
            className="flex h-9 w-9 shrink-0 items-center justify-center rounded-xl text-lg"
            style={{ backgroundColor: t.tint }}
          >
            {icon}
          </span>
        )}
      </div>
      <p className="mt-1 text-2xl font-extrabold" style={{ color: t.ink }}>{value}</p>
      {hint && <p className="mt-0.5 text-[11px]" style={{ color: 'var(--t-faint)' }}>{hint}</p>}
    </>
  )

  const shell =
    'rounded-2xl border p-4 text-right transition-colors ' +
    (onClick ? 'w-full hover:brightness-[0.98] cursor-pointer' : '')

  return onClick ? (
    <button type="button" onClick={onClick} className={shell}
      style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
      {body}
    </button>
  ) : (
    <div className={shell} style={{ backgroundColor: 'var(--sf-card)', borderColor: 'var(--bd-line)' }}>
      {body}
    </div>
  )
}
