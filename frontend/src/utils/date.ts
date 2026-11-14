// ═══ 📅 صيغة التاريخ الموحّدة (دازاي 11-15) ═══
// جان التاريخ بأربع صيغ (عربي، إنكليزي، ISO خام…) وكل شاشة بدالتها. هسه: عربي وبتوقيت بغداد دائماً.
//   fmtDate(b.createdAt)                     ← ١٥ تشرين الثاني ٢٠٢٦
//   fmtDateTime(b.scheduledAt)               ← التاريخ والساعة
//   fmtDateTime(s, { dateStyle: 'short', timeStyle: 'short' })  ← نفس المنطقة، شكل أقصر
// الفارغ يرجع `empty` (الافتراضي «—»)، والتاريخ الخربان يرجع النص كما هو بدل «Invalid Date».

export type DateInput = string | number | Date | null | undefined
export type DateOpts = Intl.DateTimeFormatOptions & {
  /** النص للقيمة الفارغة */
  empty?: string
  /** السيرفر أحياناً يرجّع وقت UTC بلا Z — هذا يعتبره UTC */
  naiveUtc?: boolean
}

const TZ = 'Asia/Baghdad'
const NAIVE = /^\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}(:\d{2}(\.\d+)?)?$/

function parse(d: Exclude<DateInput, null | undefined>, naiveUtc?: boolean): Date {
  if (d instanceof Date) return d
  if (typeof d === 'string' && naiveUtc && NAIVE.test(d)) return new Date(d.replace(' ', 'T') + 'Z')
  return new Date(d)
}

function fmt(d: DateInput, base: Intl.DateTimeFormatOptions, opts: DateOpts, withTime: boolean): string {
  const { empty = '—', naiveUtc, ...intl } = opts
  if (d === null || d === undefined || d === '') return empty
  const dt = parse(d, naiveUtc)
  if (Number.isNaN(dt.getTime())) return String(d)
  const o: Intl.DateTimeFormatOptions = Object.keys(intl).length ? intl : base
  return withTime ? dt.toLocaleString('ar-IQ', { ...o, timeZone: TZ }) : dt.toLocaleDateString('ar-IQ', { ...o, timeZone: TZ })
}

/** التاريخ بس — عربي، توقيت بغداد. */
export function fmtDate(d: DateInput, opts: DateOpts = {}): string {
  return fmt(d, { year: 'numeric', month: 'long', day: 'numeric' }, opts, false)
}

/** التاريخ والساعة — عربي، توقيت بغداد. */
export function fmtDateTime(d: DateInput, opts: DateOpts = {}): string {
  return fmt(d, { dateStyle: 'medium', timeStyle: 'short' }, opts, true)
}
