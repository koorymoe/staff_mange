import { describe, expect, it } from 'vitest'
import { askConfirm, askInput, onDialog, pendingDialogs, type DialogReq } from './dialogs'

describe('dialogs', () => {
  it('قبل ما تنرسم الشاشة: الطلب ينتظر بالطابور', async () => {
    onDialog(null)
    const p = askConfirm('تحذف؟')
    expect(pendingDialogs.length).toBe(1)
    const r = pendingDialogs.splice(0)[0]
    if (r.kind === 'confirm') r.resolve(true)
    await expect(p).resolves.toBe(true)
  })
  it('askInput يرجّع النص، والإلغاء null', async () => {
    const got: DialogReq[] = []
    onDialog((r) => got.push(r))
    const a = askInput('السبب؟', 'افتراضي')
    const b = askInput('ثاني؟')
    expect(got.map((r) => r.kind)).toEqual(['input', 'input'])
    expect(got[0].kind === 'input' && got[0].defaultValue).toBe('افتراضي')
    if (got[0].kind === 'input') got[0].resolve('مكتوب')
    if (got[1].kind === 'input') got[1].resolve(null)
    await expect(a).resolves.toBe('مكتوب')
    await expect(b).resolves.toBeNull()
    onDialog(null)
  })
})
