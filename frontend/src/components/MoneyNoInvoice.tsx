// ═══ فلوس بلا فاتورة — التحذير الي يجاوب «ليش؟» ═══
//
// (ع): «المحاسب يجي يشوف أكو حجز كامل وجايه فلوس بس ماجايه فاتورة —
// منو الإداري الي أكّده ومنو الليدر المسؤول وليش جايبة فلوس بدون
// فاتورة».
//
// النظام ما يخزّن «سبباً» مكتوباً لغياب الفاتورة — وما نخترع واحداً.
// الجواب العملي هو **الحقائق**: الفاتورة شغل الليدر وهو ما سوّاها،
// هذا اسمه ورقمه حتى يتصل بيه هسه، وهذا الإداري الي أكّد الحجز لو
// احتاج يرجع للأصل. وهذا الي يخلي الحجز ينحل بدل ما يضل معلّقاً.
//
// ⚠️ ما يظهر إلا لما **يجتمع** الشرطان: فلوس موجودة وفاتورة مفقودة.
// لأن حجز بلا فلوس وبلا فاتورة مو مشكلة محاسبة، وحجز بفاتورة خلص.
// ولو المبلغ صفر ماكو شي ينتطالب بيه.

type Props = {
  collected?: number | null
  hasInvoice: boolean
  leaderName?: string | null
  leaderPhone?: string | null
  /** مجانية مؤشَّرة — هاي مو «فاتورة ناقصة»، هاي قرار مسجَّل */
  isFree?: boolean
  className?: string
}

export default function MoneyNoInvoice({
  collected, hasInvoice, leaderName, leaderPhone,
  isFree = false, className = '',
}: Props) {
  const amount = collected ?? 0
  if (hasInvoice || isFree || amount <= 0) return null

  return (
    <div
      dir="rtl"
      className={`rounded-lg border border-red-200 bg-red-50/70 px-3 py-2 text-[11px] leading-relaxed ${className}`}
    >
      <p className="font-extrabold text-red-800">
        🔴 جايبة فلوس بلا فاتورة — {amount.toLocaleString()}
      </p>
      <p className="mt-0.5 text-red-700">
        الفاتورة شغل الليدر وما انسوّت لهذا الحجز.
      </p>
      {/* ⚠️ الأسماء ما تتكرر هنا: سطر المسؤولية فوق الصف يعرضهن
          أصلاً (رأس الهوية بشاشة المحاسب، وسطر «الليدر/أكّده»
          بالتدقيق اليومي). الي يزيده هذا الصندوق هو **الفعل**:
          زر اتصال جاهز بالليدر حتى ينحل الحجز بمكالمة مو بمطاردة. */}
      {leaderPhone ? (
        <a
          href={`tel:${leaderPhone}`}
          className="mt-1.5 inline-block rounded-lg bg-red-600 px-2.5 py-1 font-extrabold text-white hover:bg-red-700"
        >
          ☎ اتصل بالليدر{leaderName ? ` — ${leaderName}` : ''}
        </a>
      ) : (
        // «—» مو رقم غلط: ماكو رقم مسجَّل فعلاً، والمحاسب لازم يعرف
        // إنه ماكو بدل ما ينسب المسؤولية لواحد غلط.
        <p className="mt-1 font-bold text-amber-800">
          ☎ — ماكو رقم مسجَّل{leaderName ? ` لليدر ${leaderName}` : ' وماكو ليدر مسجَّل على الحجز'}
        </p>
      )}
    </div>
  )
}
