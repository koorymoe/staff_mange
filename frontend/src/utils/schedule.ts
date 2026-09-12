// ═══ موعد الحجز ينعرض مدى، مو وقت واحد ═══
//
// الإداري يتصل بالزبون ويكله «نجيك بين ٧ و٨» — لأن الوقت الواحد ما
// يصير وعد يلتزم بيه: الطريق والشغل الي قبله ما ينحسبون بالدقيقة.
// فلازم النظام يعرض نفس الي انحكى للزبون بالضبط.
//
// النهاية تجي من السيرفر (ساعة بعد البداية). لو حجز قديم بلا نهاية،
// نحسبها هنا بنفس القاعدة حتى ما يطلع بلا مدى.
const HOUR_MS = 60 * 60 * 1000

const timeOnly = (d: Date) =>
  d.toLocaleTimeString('ar-IQ', { hour: '2-digit', minute: '2-digit' })

/** «الأحد ٧ آب، من ٧:٠٠ إلى ٨:٠٠» */
export function formatScheduleWindow(
  scheduledAt: string | null | undefined,
  scheduledEndAt?: string | null,
): string {
  if (!scheduledAt) return '—'
  const start = new Date(scheduledAt)
  const end = scheduledEndAt ? new Date(scheduledEndAt) : new Date(start.getTime() + HOUR_MS)
  const day = start.toLocaleDateString('ar-IQ', { weekday: 'long', day: 'numeric', month: 'long' })
  return `${day}، من ${timeOnly(start)} إلى ${timeOnly(end)}`
}

/** «من ٧:٠٠ إلى ٨:٠٠» — بدون التاريخ، للمساحات الضيقة */
export function formatScheduleWindowTime(
  scheduledAt: string | null | undefined,
  scheduledEndAt?: string | null,
): string {
  if (!scheduledAt) return '—'
  const start = new Date(scheduledAt)
  const end = scheduledEndAt ? new Date(scheduledEndAt) : new Date(start.getTime() + HOUR_MS)
  return `من ${timeOnly(start)} إلى ${timeOnly(end)}`
}
