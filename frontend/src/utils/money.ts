// ═══ «١.٥ هيه نفسها ألف ونص» ═══
//
// (ع): «أريد تنتبهلي — مرات ١.٥ هيه نفسها ألف ونص، فإنت سوي على
// هذا الأساس».
//
// الموظف يكتب الأسعار بالألوف مختصرة: «١.٥» يعني ١٥٠٠، و«٢٥.٥»
// يعني ٢٥٥٠٠. وبلا هذا، يكتب ١.٥ بعرض سعر وينحفظ **دينار ونص**
// فالعرض يوصل الزبون بمجموع غلط بالكامل.
//
// 🔴 **القاعدة محصورة بالكسر بقصد**: التوسيع يصير **بس** لما يكتب
// رقماً **بيه فاصلة** وأقل من ألف. ليش؟
//
//  · ماكو سعر حقيقي بالعراق بكسر دينار — ما إله «١.٥ دينار». فأي
//    كسر صغير هو اختصار ألوف أكيد، مو رقماً يريده.
//  · أما الأرقام الصحيحة فلا نلمسها: «٢٥٠» ممكن تكون ٢٥٠ ألف وممكن
//    تكون ٢٥٠ ديناراً فعلاً (سعر متر كيبل، أو أجرة صغيرة). وتخمينها
//    يضرب سعراً صحيحاً بألف — وهاي **خسارة فلوس** مو إزعاج شاشة.
//
// ⚠️ والتوسيع **يبان دائماً**: الدالة ترجّع `converted` حتى الشاشة
// تگول «كتبت ١.٥ ← حفظناها ١٬٥٠٠». التبديل الصامت بالفلوس ممنوع،
// حتى لو صحيح — الموظف لازم يشوف شنو انحفظ.

export type MoneyInput = { value: number; converted: boolean; typed: string }

/** يفسّر نص سعر: الكسر الأقل من ألف = ألوف مختصرة. */
export function readMoney(raw: string): MoneyInput {
  const typed = (raw ?? '').trim()
  const n = Number(typed)
  if (!Number.isFinite(n) || n <= 0) return { value: Number.isFinite(n) ? n : 0, converted: false, typed }
  // بيه فاصلة عشرية فعلاً (مو «5.» ولا «5»)، وأقل من ألف
  const hasFraction = /^\d*\.\d+$/.test(typed)
  if (hasFraction && n < 1000) return { value: n * 1000, converted: true, typed }
  return { value: n, converted: false, typed }
}

/** صيغة عرض المبلغ — كسور تبان لو موجودة، وبلا كسور وهمية. */
export function money(n: number): string {
  return (n ?? 0).toLocaleString('en-US', { maximumFractionDigits: 2 })
}

// يقرا مبلغ مكتوب بأي أرقام (عربية ٠-٩ أو فارسية ۰-۹ أو إنگليزية) ويشيل
// الفواصل — «٥٠٠٬٠٠٠» و«500,000» نفس الرقم. يرجّع 0 إذا ماكو رقم.
export function parseAmount(s: string): number {
  const ascii = s.replace(/[٠-٩]/g, (d) => String(d.charCodeAt(0) - 0x0660)).replace(/[۰-۹]/g, (d) => String(d.charCodeAt(0) - 0x06f0))
  return Number(ascii.replace(/[^\d]/g, '')) || 0
}
