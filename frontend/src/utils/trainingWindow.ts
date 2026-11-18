// (دازاي 11-18) وقت التدريب اليومي بتوقيت بغداد — فارغ = بلا وقت (التنبيه يبين دائماً).
export function baghdadClock(now: Date): string {
  return now.toLocaleTimeString('en-GB', { timeZone: 'Asia/Baghdad', hour: '2-digit', minute: '2-digit', hour12: false })
}

export function inTrainingWindow(from: string | null | undefined, to: string | null | undefined, now: Date): boolean {
  if (!from || !to) return true
  const t = baghdadClock(now)
  return t >= from && t < to
}
