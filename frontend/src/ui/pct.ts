// نسبة مقرّبة أو null إذا المقام صفر — تنستخدم ويا PctWhy.
export function pctOf(num: number, den: number) {
  return den > 0 ? Math.round((num / den) * 100) : null
}

// (دازاي 11-15) لون حلقة التقدّم يتبع النسبة: ٨٠+ أخضر، ٥٠+ عنبري، والباقي برتقالي.
export function ringColor(p: number) {
  return p >= 80 ? '#16a34a' : p >= 50 ? '#f59e0b' : '#fb923c'
}
