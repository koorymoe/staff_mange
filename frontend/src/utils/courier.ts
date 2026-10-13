// رسول ماتركس (ع 10-13): الرحلة كلها ~٢٠ ثانية
export const COURIER_TRIP_MS = 20_000

// الجفن ينفتح/يغمض ببطء — مسار SVG بلا transition، فنحرّكه بخطوات.
export function tweenLid(from: number, to: number, ms: number, set: (v: number) => void): () => void {
  const start = performance.now()
  let raf = 0
  const step = (t: number) => {
    const k = Math.min(1, (t - start) / ms)
    set(from + (to - from) * k)
    if (k < 1) raf = requestAnimationFrame(step)
  }
  raf = requestAnimationFrame(step)
  return () => cancelAnimationFrame(raf)
}
