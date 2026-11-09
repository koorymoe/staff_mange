import { describe, it, expect, vi } from 'vitest'
import { swallow, swallowed } from './swallow'

describe('swallow', () => {
  it('records and caps at 20', () => {
    const w = vi.spyOn(console, 'warn').mockImplementation(() => {})
    for (let i = 0; i < 25; i++) swallow(new Error('e' + i))
    expect(swallowed.length).toBe(20)
    expect(swallowed[19].msg).toBe('e24')
    expect(w).toHaveBeenCalled()
  })
})
