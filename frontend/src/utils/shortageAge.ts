// (طلب ماتركس 0ef76f3b-3 — 11-18) فلتر الجرد حسب عمر النقص.
// checkedAt من الخادم TIMESTAMP بلا منطقة = UTC.
export function ageDays(checkedAt: string, now: number): number {
  const s = /[zZ]|[+-]\d\d:?\d\d$/.test(checkedAt) ? checkedAt : checkedAt.replace(' ', 'T') + 'Z'
  const t = new Date(s).getTime()
  return Number.isNaN(t) ? 0 : Math.max(0, Math.floor((now - t) / 86_400_000))
}

/** النواقص الي عمرها minDays فما فوق، الأقدم أول. */
export function byShortageAge<T extends { checkedAt: string }>(rows: T[], minDays: number, now: number): T[] {
  return rows
    .map((r) => ({ r, a: ageDays(r.checkedAt, now) }))
    .filter((x) => x.a >= minDays)
    .sort((x, y) => y.a - x.a)
    .map((x) => x.r)
}
