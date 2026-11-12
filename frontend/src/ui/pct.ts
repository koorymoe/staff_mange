// نسبة مقرّبة أو null إذا المقام صفر — تنستخدم ويا PctWhy.
export function pctOf(num: number, den: number) {
  return den > 0 ? Math.round((num / den) * 100) : null
}
