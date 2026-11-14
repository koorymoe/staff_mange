import { describe, expect, it } from 'vitest'
import { bookingServiceLabel, NO_SERVICE_LABEL } from './bookingServiceLabel'

describe('bookingServiceLabel', () => {
  it('الخدمة الرئيسية أول، وباقي الخدمات بلا تكرار', () => {
    expect(bookingServiceLabel({ service: { name: 'كاميرات' } })).toBe('كاميرات')
    expect(bookingServiceLabel({ service: { name: 'كاميرات' }, services: [{ name: 'كاميرات' }, { name: 'إنذار' }] })).toBe('كاميرات، إنذار')
  })
  it('بلا رئيسية: من قائمة الخدمات', () => {
    expect(bookingServiceLabel({ service: null, services: [{ name: 'شبكات' }, { name: 'سمارت' }] })).toBe('شبكات، سمارت')
  })
  it('الصيانة: «صيانة» + خدمة الزبون إذا معروفة', () => {
    expect(bookingServiceLabel({ bookingType: 'MAINTENANCE', customer: { services: ['كاميرات'] } })).toBe('صيانة — كاميرات')
    expect(bookingServiceLabel({ bookingType: 'MAINTENANCE', customer: null })).toBe('صيانة')
  })
  it('ماكو شي: تحذير واضح', () => {
    expect(bookingServiceLabel({ bookingType: 'REGULAR' })).toBe(NO_SERVICE_LABEL)
    expect(bookingServiceLabel(null)).toBe(NO_SERVICE_LABEL)
  })
})
