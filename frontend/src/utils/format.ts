// ═══ (DRY 11-02) دوال التنسيق المشتركة ═══
// چانت كل شاشة تكتب نسختها: نفس «fmt» و«money» و«today» بعشرات الملفات،
// وبعضها اختلف بلا قصد (today بتوقيت غرينتش ضد بغداد). هسه مصدر واحد.
// ⚠️ الأسماء تبيّن الشكل بالضبط — الشاشة تختار الي يناسبها بلا ما تغيّر عرضها.

/** 1,234 (بلا تقريب) */
export const num = (n: number) => n.toLocaleString('en-US')
/** 1,235 (مقرّب لأقرب عدد صحيح) */
export const numRound = (n: number) => Math.round(n).toLocaleString('en-US')
/** صيغة en-IQ (نفس شكل الأرقام الي چانت بشاشات الأسعار والإحصاءات) */
export const numIQ = (n: number) => n.toLocaleString('en-IQ')
/** 1,234 د.ع */
export const dinar = (n: number) => n.toLocaleString('en-US') + ' د.ع'
/** 1,235 د.ع — مقرّب، والفارغ = 0 */
export const dinarRound = (n: number) => `${Math.round(n || 0).toLocaleString('en-US')} د.ع`
/** مقرّب لأقرب ألف: 1,235,000 د.ع */
export const dinarK = (n: number) => `${num(Math.round(n / 1000) * 1000)} د.ع`
/** تاريخ اليوم بتوقيت بغداد YYYY-MM-DD */
export const todayBaghdad = () => new Date().toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' })
/** (DRY 11-04) تاريخ YYYY-MM-DD لأي وقت بتوقيت بغداد — فاضي إذا ماكو */
export const dayBaghdad = (iso: string | Date | null | undefined) => (iso ? new Date(iso).toLocaleDateString('en-CA', { timeZone: 'Asia/Baghdad' }) : '')
/** ثواني ← س:دد:ثث */
export const hms = (sec: number) => `${Math.floor(sec / 3600)}:${String(Math.floor((sec % 3600) / 60)).padStart(2, '0')}:${String(sec % 60).padStart(2, '0')}`
/** تاريخ ووقت عربي متوسط */
export const dateTimeAr = (s: string) => new Date(s).toLocaleString('ar-IQ', { dateStyle: 'medium', timeStyle: 'short' })

/** (دازاي 11-27) فترة الشركة الحالية «YYYY-MM» (من ٢٦ لـ٢٥) — يوم ٢٦ وبعده = الشهر الجاي. */
export const thisPeriod = () => {
  const d = todayBaghdad()
  const [y, m, day] = d.split('-').map(Number)
  if (day < 26) return d.slice(0, 7)
  return new Date(Date.UTC(y, m, 1)).toISOString().slice(0, 7)
}
/** حدود الفترة للعرض: «26/10 ← 25/11» */
export const periodRangeLabel = (p: string) => {
  const [y, m] = p.split('-').map(Number)
  const prev = new Date(Date.UTC(y, m - 2, 1))
  return `26/${prev.getUTCMonth() + 1} ← 25/${m}`
}
