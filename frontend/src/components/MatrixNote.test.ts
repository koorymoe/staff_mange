import { describe, it, expect } from 'vitest'
import { splitSentences } from './matrixNoteText'

// (دازاي 11-13) رسالة ماتركس الطويلة تنقسم: أول جملة ملخّص، والباقي نقاط.
describe('splitSentences', () => {
  it('splits on full stops, semicolons and new lines', () => {
    expect(splitSentences('يحتاج متابعة. حضر بوقته 16 من 26 يوم؛ متوسط تأخيره 6 ساعات.\nنمط: 9 من 12 إجازاته الأحد.'))
      .toEqual(['يحتاج متابعة.', 'حضر بوقته 16 من 26 يوم؛', 'متوسط تأخيره 6 ساعات.', 'نمط: 9 من 12 إجازاته الأحد.'])
  })
  it('keeps a short message whole', () => {
    expect(splitSentences('ماكو شي مهم')).toEqual(['ماكو شي مهم'])
  })
})
