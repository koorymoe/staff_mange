import { describe, expect, it } from 'vitest'
import { ageDays, byShortageAge } from './shortageAge'

const now = Date.parse('2026-11-18T12:00:00Z')
describe('shortageAge', () => {
  it('reads naive timestamps as UTC', () => {
    expect(ageDays('2026-11-15 12:00:00', now)).toBe(3)
    expect(ageDays('2026-11-18T11:00:00Z', now)).toBe(0)
  })
  it('filters by min age, oldest first', () => {
    const rows = [{ id: 'a', checkedAt: '2026-11-17T12:00:00Z' }, { id: 'b', checkedAt: '2026-11-08T12:00:00Z' }, { id: 'c', checkedAt: '2026-11-14T12:00:00Z' }]
    expect(byShortageAge(rows, 2, now).map((r) => r.id)).toEqual(['b', 'c'])
    expect(byShortageAge(rows, 0, now).map((r) => r.id)).toEqual(['b', 'c', 'a'])
  })
})
