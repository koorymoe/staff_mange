import { describe, expect, it } from 'vitest'
import { inTrainingWindow } from './trainingWindow'

describe('inTrainingWindow', () => {
  const at = (iso: string) => new Date(iso)
  it('no window = always', () => expect(inTrainingWindow(null, null, at('2026-11-18T05:00:00Z'))).toBe(true))
  it('uses Baghdad time', () => {
    // 07:30 UTC = 10:30 بغداد
    expect(inTrainingWindow('10:00', '11:00', at('2026-11-18T07:30:00Z'))).toBe(true)
    expect(inTrainingWindow('10:00', '11:00', at('2026-11-18T08:00:00Z'))).toBe(false)
    expect(inTrainingWindow('10:00', '11:00', at('2026-11-18T06:59:00Z'))).toBe(false)
  })
})
