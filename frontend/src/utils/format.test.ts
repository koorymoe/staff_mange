import { describe, expect, it } from 'vitest'
import { dayBaghdad, dinar, dinarRound, hms, num } from './format'

describe('format', () => {
  it('dayBaghdad: اليوم بتوقيت بغداد مو UTC', () => {
    // ٢٢:٠٠ UTC = ٠١:٠٠ باچر ببغداد (+٣)
    expect(dayBaghdad('2026-11-03T22:00:00Z')).toBe('2026-11-04')
    expect(dayBaghdad('2026-11-03T20:59:00Z')).toBe('2026-11-03')
    expect(dayBaghdad(null)).toBe('')
    expect(dayBaghdad(undefined)).toBe('')
  })
  it('الفلوس بفواصل الآلاف', () => {
    expect(num(1234567)).toBe('1,234,567')
    expect(dinar(2500)).toBe('2,500 د.ع')
    expect(dinarRound(2499.6)).toBe('2,500 د.ع')
  })
  it('hms: ساعة:دقيقة:ثانية', () => {
    expect(hms(0)).toBe('0:00:00')
    expect(hms(27000)).toBe('7:30:00')
    expect(hms(3661)).toBe('1:01:01')
  })
})
