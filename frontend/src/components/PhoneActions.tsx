import { intlPhoneOrNull } from '../utils/phone'

/**
 * ═══ اتصال + واتساب بجنب رقم الزبون ═══
 *
 * «مو كلمن ورقمه — أريد تحت الرقم يكون علامة الاتصال وعلامة واتساب
 * للتواصل السريع».
 *
 * ⚠️ ينعرض بالشاشات الي (ع) طلبها صراحةً: الحجز الجديد · طابور
 * «ما رد» · طلبات حذف الحجوزات. مو بكل النظام.
 *
 * ⚠️ **ما يظهر بلا رقم صالح**: يستخدم `intlPhoneOrNull` مو
 * `toIntlPhone` — الثانية ترجّع أرقاماً ناقصة كما هي، و
 * `wa.me/<رقم ناقص>` **يفتح ويفشل صامتاً**: الإداري يحسب إنه أرسل
 * وهو لا. فزر يودّي لفراغ أسوأ من ماكو زر.
 *
 * `telegram` اختياري — ينضاف لمن الشاشة تحتاجه.
 */
export default function PhoneActions({
  phone,
  telegram = false,
  className = '',
}: {
  phone?: string | null
  telegram?: boolean
  className?: string
}) {
  const raw = (phone || '').trim()
  const intl = intlPhoneOrNull(raw)
  if (!raw || !intl) return null

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
      {telegram && (
        <a
          href={`https://t.me/+${intl}`}
          target="_blank"
          rel="noreferrer"
          onClick={(e) => e.stopPropagation()}
          title="افتح محادثة تلغرام"
          className="rounded-full bg-sky-50 px-2 py-0.5 text-[11px] font-bold text-sky-700 hover:bg-sky-100"
        >
          تلغرام
        </a>
      )}
    </span>
  )
}
