// بند ٧ج: بدل `.catch(() => {})` الي يبلع الخطأ بصمت — نبقى ما نزعج المستخدم
// (أغلبها تحميلات ثانوية لبطاقات) بس الخطأ ينطبع بالـ console وينحفظ آخر ٢٠ للفحص.
export const swallowed: { at: string; msg: string }[] = []

export function swallow(e: unknown): void {
  const msg = e instanceof Error ? e.message : String(e)
  swallowed.push({ at: new Date().toISOString(), msg })
  if (swallowed.length > 20) swallowed.shift()
  console.warn('[swallowed]', e)
}
