import { beforeAll, describe, expect, it, vi } from 'vitest'

// البيئة node: الملف يقرا location و localStorage عند التحميل (عبر api.ts)
vi.stubGlobal('location', { origin: 'http://localhost', href: 'http://localhost/' })
vi.stubGlobal('localStorage', { getItem: () => null, setItem: () => {}, removeItem: () => {} })
let mod: typeof import('./projectQuotePrint')
beforeAll(async () => { mod = await import('./projectQuotePrint') })

describe('parseSpecs', () => {
  it('يقسم «عنوان: قيمة» ويشيل الفاضي', () => {
    expect(mod.parseSpecs('الدقة: 4MP\n\n  العدسة：2.8mm \nبلا عنوان')).toEqual([
      { label: 'الدقة', value: '4MP' },
      { label: 'العدسة', value: '2.8mm' },
      { label: '', value: 'بلا عنوان' },
    ])
    expect(mod.parseSpecs(null)).toEqual([])
  })
})

describe('buildProjectQuoteHtml', () => {
  const evil = '<img src=x onerror=alert(1)>'
  const build = () => mod.buildProjectQuoteHtml(
    {
      customerName: evil,
      cover: { title: evil, label: evil, subtitle: evil, city: evil },
      items: [{ productName: evil, modelName: evil, specs: `${evil}: ${evil}`, category: evil, quantity: 2, unitPrice: 1500, unit: evil }],
      annex: [{ title: evil, items: [{ title: evil, text: evil }] }],
      date: new Date(2026, 10, 6),
    },
    '/logo.png',
  )
  it('ما يمرر HTML من نص المستخدم (كل الحقول)', () => {
    const html = build()
    expect(html).not.toContain('<img src=x')
    expect(html).toContain('&lt;img src=x')
  })
  it('الإجمالي = الكمية × السعر', () => {
    expect(build()).toContain('3,000')
  })
})
