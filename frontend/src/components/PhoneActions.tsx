import { toIntlPhone } from '../utils/phone'

/**
 * ═══ اتصال + واتساب بجنب رقم الزبون ═══
 *
 * «مو كلمن ورقمه — أريد تحت الرقم يكون علامة الاتصال وعلامة واتساب
 * للتواصل السريع».
 *
 * ⚠️ قرار (ع): هذولا **بمكانين بس** — وقت الحجز الجديد — مو بكل
 * النظام. (زر النسخ هو الي بكل مكان.)
 *
 * ⚠️ ما يظهر بلا رقم: زران يودّيان لفراغ أسوأ من ماكو زر.
 */
export default function PhoneActions({
  phone,
  className = '',
}: {
  phone?: string | null
  className?: string
}) {
  const raw = (phone || '').trim()
  if (!raw) return null
  const intl = toIntlPhone(raw)

  return (
    <span className={`inline-flex items-center gap-1.5 ${className}`}>
      <a
        href={`tel:${raw}`}
        onClick={(e) => e.stopPropagation()}
        title="اتصل بالزبون"
        className="rounded-full bg-brand-50 px-2 py-0.5 text-[11px] font-bold text-brand-700 hover:bg-brand-100"
      >
        📞 اتصال
      </a>
      <a
        href={`https://wa.me/${intl}`}
        target="_blank"
        rel="noreferrer"
        onClick={(e) => e.stopPropagation()}
        title="افتح محادثة واتساب"
        className="rounded-full bg-emerald-50 px-2 py-0.5 text-[11px] font-bold text-emerald-700 hover:bg-emerald-100"
      >
        واتساب
      </a>
    </span>
  )
}
