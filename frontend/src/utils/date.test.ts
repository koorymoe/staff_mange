import { describe, expect, it } from 'vitest'
import { fmtDate, fmtDateTime } from './date'

describe('date', () => {
  it('الفارغ والخربان', () => {
    expect(fmtDate(null)).toBe('—')
    expect(fmtDateTime(undefined, { empty: 'بلا موعد' })).toBe('بلا موعد')
    expect(fmtDate('مو تاريخ')).toBe('مو تاريخ')
  })
  it('توقيت بغداد: ٢٢:٣٠ UTC = اليوم الجاي ببغداد', () => {
    const a = fmtDate('2026-01-01T22:30:00Z', { day: 'numeric' })
    const b = fmtDate('2026-01-02T10:00:00Z', { day: 'numeric' })
    expect(a).toBe(b)
  })
  it('naiveUtc يعتبر الوقت بلا Z توقيت عالمي', () => {
    expect(fmtDateTime('2026-01-01 22:30:00', { naiveUtc: true, day: 'numeric' }))
      .toBe(fmtDateTime('2026-01-01T22:30:00Z', { day: 'numeric' }))
  })
})
