import { describe, expect, it } from 'vitest'
import { beforeAll, vi } from 'vitest'
import { buildPlanSetHtml } from './planSetPrint'

beforeAll(() => { vi.stubGlobal('location', { origin: 'https://x.test' }) })

// (ل) 11-12 دازاي: غلاف قبل كل مخطط بيه اسمه، وفورما جوّه كل مخطط (لوغو، الاسم، الملاحظات، السكيل، التاريخ).
describe('PDF سيت المخططات: الغلاف والفورما', () => {
  const groups = [{ no: 7, name: 'الأبواب والشبابيك', en: 'Doors & Windows', types: [{ code: '7.1', name: 'مخطط الأبواب' }] }]
  const base = { title: 'سيت دار مشيد', projectName: 'PRJ-1', groups, quote: null, noQuote: true, date: new Date('2026-11-12T10:00:00Z') }
  it('غلاف قبل كل مخطط، والفورما على اليمين بالافتراضي ويا اللوغو والتاريخ', () => {
    const html = buildPlanSetHtml({ ...base, files: [
      { typeCode: '7.1', title: 'a.png', fileKey: '/api/files/designs/a', fileType: 'image/png', createdAt: '2026-11-10T09:00:00Z' },
      { typeCode: '7.1', title: 'b.png', fileKey: '/api/files/designs/b', fileType: 'image/png', sheetPos: 'bottom', sheetName: 'الطابق الأرضي: أبواب وشبابيك', sheetNotes: 'القياسات بالمتر', sheetScale: '1:100' },
    ] }, '/logo.png')
    // صورتين من نفس النوع = غلاف واحد قبلهن
    expect(html.match(/class="page land pcover"/g)).toHaveLength(1)
    expect(html).toContain('2 لوحات')
    expect(html.match(/class="page land plan sheet-right"/g)).toHaveLength(1)
    expect(html.match(/class="page land plan sheet-bottom"/g)).toHaveLength(1)
    // الغلاف قبل مخططه
    expect(html.indexOf('class="page land pcover"')).toBeLessThan(html.indexOf('class="page land plan sheet-right"'))
    expect(html).toContain('<h2>مخطط الأبواب</h2>')
    expect(html).toContain('<span>الطابق الأرضي: أبواب وشبابيك</span>')
    expect(html).toContain('1:100')
    expect(html).toContain('القياسات بالمتر')
    expect(html).toContain('10/11/2026') // تاريخ رفع المخطط
    expect(html).not.toContain('src="LOGO"')
    expect(html.match(/class="tb-logo"><img src="[^"]*logo.png"/g)).toHaveLength(2)
  })
  it('PDF و DWG ما يطلعلهن غلاف (ما ينطبعن)', () => {
    const html = buildPlanSetHtml({ ...base, files: [{ typeCode: '7.1', title: 'x.pdf', fileKey: '/api/files/designs/x', fileType: 'application/pdf' }] }, '/logo.png')
    expect(html).not.toContain('class="page land pcover"')
  })
  it('الترقيم يتسلسل بالملف: 10.2 و10.4 يصيرن 10.1 و10.2، والقسم يبقى', () => {
    const g = [{ no: 10, name: 'حزمة العرض', en: 'Boards', types: [{ code: '10.1', name: 'أ' }, { code: '10.2', name: 'اللقطات الرئيسية' }, { code: '10.3', name: 'ج' }, { code: '10.4', name: 'اللقطات المنظورية' }] }]
    const img = (c: string) => ({ typeCode: c, title: c, fileKey: '/api/files/designs/' + c, fileType: 'image/png' })
    const html = buildPlanSetHtml({ ...base, groups: g, files: [img('10.4'), img('10.2'), img('10.2')] }, '/logo.png')
    const codes = [...html.matchAll(/class="pc-code" dir="ltr">([^<]+)</g)].map((m) => m[1])
    expect(codes).toEqual(['10.1', '10.2'])
    expect(html.indexOf('اللقطات الرئيسية')).toBeLessThan(html.indexOf('اللقطات المنظورية'))
    expect(html).toContain('01 / 02')
  })
})
