import { describe, expect, it } from 'vitest'
import { parseAmount, readMoney } from './money'

describe('money', () => {
  it('parseAmount: أرقام عربية وفارسية وفواصل', () => {
    expect(parseAmount('٥٠٠٬٠٠٠')).toBe(500000)
    expect(parseAmount('500,000')).toBe(500000)
    expect(parseAmount('۲۵۰۰')).toBe(2500)
    expect(parseAmount('ماكو رقم')).toBe(0)
  })
  it('readMoney: «١.٥» = ألف ونص، والصحيح ما ينلمس', () => {
    expect(readMoney('1.5')).toEqual({ value: 1500, converted: true, typed: '1.5' })
    expect(readMoney('25.5').value).toBe(25500)
    expect(readMoney('250')).toEqual({ value: 250, converted: false, typed: '250' })
    expect(readMoney('1500.5').converted).toBe(false)
  })
})
