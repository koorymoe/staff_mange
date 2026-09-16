import { useState } from 'react'
import { copyText } from '../utils/clipboard'

/**
 * ═══ كود الحجز + زر نسخ ═══
 *
 * «بجانب كل كود حجز أو رقم حجز أريد علامة نسخ، حتى بس أضغط عليها
 * ينسخ الحجز — الموضوع متعب من أكعد أحدد وأنسخ وألصق».
 *
 * ⚠️ الأنماط بالنظام مختلفة (font-mono، حبة داكنة، خط عريض)، فالمكوّن
 * ما يفرض شكلاً: `className` تجي من مكان الاستعمال، وهو يضيف الزر بس.
 * بدون هذا، كل موضع ينبدّل يتغيّر شكله ويصير التبديل «إعادة تصميم».
 */
export default function BookingCodeChip({
  code,
  className = '',
  title,
}: {
  code?: string | null
  className?: string
  /** نص بديل للتلميح (افتراضياً: انسخ كود الحجز) */
  title?: string
}) {
  const [copied, setCopied] = useState(false)
  if (!code) return null

  const doCopy = async (e: React.MouseEvent) => {
    // ⚠️ الكود غالباً جوّا صف/بطاقة قابلة للضغط — بلا هذا، النسخ
    // يفتح التفاصيل بنفس الضغطة والمستخدم يظن الزر ما اشتغل.
    e.stopPropagation()
    e.preventDefault()
    const ok = await copyText(code)
    if (!ok) return
    setCopied(true)
    setTimeout(() => setCopied(false), 1400)
  }

  return (
    <span className={`inline-flex items-center gap-1 ${className}`}>
      {code}
      <button
        type="button"
        onClick={doCopy}
        title={title || 'انسخ كود الحجز'}
        aria-label={title || 'انسخ كود الحجز'}
        className="rounded px-1 text-[11px] leading-none opacity-60 transition hover:opacity-100"
      >
        {copied ? '✓' : '⧉'}
      </button>
    </span>
  )
}
