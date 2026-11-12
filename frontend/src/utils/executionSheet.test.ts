import { describe, expect, it } from 'vitest'
import type { Booking } from '../api'
import { clockOf, emptySheet } from './executionSheet'

// بلاغ #34 (مؤمل 11-12): وقت البدء والانتهاء بتقرير الخدمة يتعبّى من الحجز (بتوقيت بغداد).
describe('تقرير الخدمة: الوقت من الحجز', () => {
  it('clockOf: ساعة بغداد، الدقيقة لأقرب ٥ لتحت، والوقت بالضبط بالملاحظة', () => {
    expect(clockOf('2026-11-12T06:47:00Z')).toEqual({ h: '09', m: '45', note: 'من الحجز: 09:47' })
    expect(clockOf(null)).toEqual({ h: '', m: '', note: '' })
    expect(clockOf('غلط')).toEqual({ h: '', m: '', note: '' })
  })
  it('emptySheet يعبّي البدء (startedAt وإلا arrivedAt) والانتهاء (completedAt)', () => {
    const b = { id: 'b', startedAt: null, arrivedAt: '2026-11-12T05:03:00Z', completedAt: '2026-11-12T11:30:00Z', customer: {}, assignments: [] } as unknown as Booking
    const t = emptySheet(b).time
    expect(t).toMatchObject({ startH: '08', startM: '00', endH: '14', endM: '30', endNote: 'من الحجز: 14:30' })
  })
  it('حجز بلا أوقات = فاضي يعبّيه بإيده', () => {
    const t = emptySheet({ id: 'b', customer: {}, assignments: [] } as unknown as Booking).time
    expect(t.startH + t.endH).toBe('')
  })
})
