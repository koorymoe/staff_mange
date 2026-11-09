import { describe, expect, it } from 'vitest'
import { arabicAmountWords, arabicNumberWords } from './arabicWords'

describe('arabicNumberWords', () => {
  it('مبلغ نموذج جامعة المكين', () => {
    expect(arabicAmountWords(222945000)).toBe('مائتان واثنان وعشرون مليوناً وتسعمائة وخمسة وأربعون ألف دينار عراقي')
  })
  it('حالات المفرد والمثنى والجمع', () => {
    expect(arabicNumberWords(0)).toBe('صفر')
    expect(arabicNumberWords(1000)).toBe('ألف')
    expect(arabicNumberWords(2000)).toBe('ألفان')
    expect(arabicNumberWords(5000)).toBe('خمسة آلاف')
    expect(arabicNumberWords(1250000)).toBe('مليون ومائتان وخمسون ألف')
    expect(arabicNumberWords(925000)).toBe('تسعمائة وخمسة وعشرون ألف')
    expect(arabicNumberWords(3500000)).toBe('ثلاثة ملايين وخمسمائة ألف')
    expect(arabicNumberWords(11)).toBe('أحد عشر')
    expect(arabicNumberWords(101)).toBe('مائة وواحد')
  })
})
