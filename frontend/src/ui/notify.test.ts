import { describe, expect, it } from 'vitest'
import { notify, onNotify, pendingToasts, type Toast } from './notify'

describe('notify', () => {
  it('قبل ما تنرسم الشاشة: الرسائل تنتظر بالطابور وبالترتيب', () => {
    onNotify(null)
    pendingToasts.splice(0)
    notify('أول')
    notify('تم ✅', 'success')
    expect(pendingToasts.map((t) => [t.message, t.kind])).toEqual([['أول', 'error'], ['تم ✅', 'success']])
    expect(pendingToasts[0].id).toBeLessThan(pendingToasts[1].id)
    pendingToasts.splice(0)
  })
  it('بعد ما تنرسم: توصل للمستمع مباشرة', () => {
    const got: Toast[] = []
    onNotify((t) => got.push(t))
    notify('معلومة', 'info')
    expect(got).toHaveLength(1)
    expect(got[0].kind).toBe('info')
    expect(pendingToasts).toHaveLength(0)
    onNotify(null)
  })
})
