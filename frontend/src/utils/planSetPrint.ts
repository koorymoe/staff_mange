// ═══ 📐 سيت المخططات — ملف الـPDF (ك) 11-10 ═══
// بطلب دازاي: «تسوي تنزيل ملف ينزل ملف PDF بالألوان الهوية البصرية مال الشركة بي كل
// التصاميم بالإضافة الى عرض السعر». الشكل نفس ورقة «Package Contents» مال الشركة:
// شريط كحلي وشعار. الطباعة بنفس إطار عرض السعر (printProjectQuote) — المستخدم يختار
// «حفظ كـPDF». الصور تنطبع صفحة لكل مخطط؛ PDF وDWG ما ينطبعن جوّه، فيطلعن بقائمة المرفقات.
import fontsCss from '../fonts.css?inline'
import { fileUrl } from '../api'
import { esc } from './printIdentity'

export type PlanPrintGroup = { no: number; name: string; en: string; types: { code: string; name: string }[] }
export type PlanPrintFile = { typeCode: string; title: string; fileKey: string; fileType?: string | null }
export type PlanPrintQuote = {
  quotationNumber: string
  customerName: string
  items: { productName: string; unit: string; quantity: number; unitPrice: number; totalPrice: number }[]
  grandTotal: number
  discountValue: number
  netTotal: number
}
export type PlanPrintInput = {
  title: string
  projectName?: string | null
  groups: PlanPrintGroup[]
  files: PlanPrintFile[]
  quote: PlanPrintQuote | null
  /** سبب غياب عرض السعر (ماكو مشروع، ماكو عرض، ما عنده صلاحية) */
  quoteNote?: string
  // (ل) 11-10: «ملفين»: PDF التصاميم بلا عرض السعر (العرض PDF وحده بقالبه)
  noQuote?: boolean
  date: Date
}

const COMPANY = 'شركة الأماني المتحدة للتجارة والخدمات العامة والمقاولات والاستثمارات والتطوير العقاري محدودة المسؤولية'
const NAVY = '#2b3074'
const IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/webp']

const n = (v: number) => Math.round(v).toLocaleString('en-US')
const abs = (u: string) => (/^(https?:|data:|blob:)/.test(u) ? u : new URL(u, location.origin).href)
export const isImageFile = (f: { fileType?: string | null }) => IMAGE_TYPES.includes(f.fileType ?? '')

const band = `<div class="band"><img src="LOGO" alt=""><span>${COMPANY}</span></div>`
const foot = (page: number) => `<div class="foot"><span>شركة الأماني المتحدة</span><bdi dir="ltr">${String(page).padStart(2, '0')}</bdi></div>`

export function buildPlanSetHtml(p: PlanPrintInput, logoUrl: string): string {
  const typeName = new Map<string, { group: PlanPrintGroup; name: string }>()
  for (const g of p.groups) for (const t of g.types) typeName.set(t.code, { group: g, name: t.name })
  const order = new Map<string, number>()
  p.groups.forEach((g, gi) => g.types.forEach((t, ti) => order.set(t.code, gi * 100 + ti)))
  const files = [...p.files].sort((a, b) => (order.get(a.typeCode) ?? 9999) - (order.get(b.typeCode) ?? 9999))
  const images = files.filter(isImageFile)
  const others = files.filter((f) => !isImageFile(f))
  const present = new Set(files.map((f) => f.typeCode))
  const day = p.date.toLocaleDateString('en-GB', { timeZone: 'Asia/Baghdad' })
  let page = 1

  const cover = `<section class="page cover">${band}
    <div class="cover-body"><small dir="ltr">PLANS PACKAGE</small><h1>${esc(p.title)}</h1>
      ${p.projectName ? `<p>${esc(p.projectName)}</p>` : ''}<p class="muted">${esc(day)} · ${files.length} مخطط</p></div>
    ${foot(page++)}</section>`

  // «Package Contents»: الأقسام العشرة، والي بيه ملفات يتعلّم
  const contents = `<section class="page">${band}<div class="box">Package Contents</div><ol class="contents">
    ${p.groups.map((g) => `<li class="${g.types.some((t) => present.has(t.code)) ? 'on' : ''}"><b>${g.no}</b><span dir="ltr">${esc(g.en)}</span><span>${esc(g.name)}</span></li>`).join('')}
  </ol>${foot(page++)}</section>`

  const pages = images.map((f) => {
    const t = typeName.get(f.typeCode)
    return `<section class="page">${band}
      <div class="ph"><span class="tag">${esc(f.typeCode)}</span><div><b>${esc(t?.name ?? f.typeCode)}</b><small>${esc(t ? `${t.group.no}. ${t.group.name}` : '')}${f.title && f.title !== f.typeCode ? ` · ${esc(f.title)}` : ''}</small></div></div>
      <div class="img"><img src="${esc(abs(fileUrl(f.fileKey)))}" alt=""></div>${foot(page++)}</section>`
  }).join('')

  const attach = others.length ? `<section class="page">${band}<div class="box">ملفات مرفقة (PDF / أوتوكاد)</div>
    <p class="muted">هاي الملفات ما تنطبع جوّه الحزمة — تنفتح من النظام أو تنرسل ويا الملف.</p>
    <table><thead><tr><th>النوع</th><th>الاسم</th><th>الصيغة</th></tr></thead><tbody>
    ${others.map((f) => `<tr><td>${esc(f.typeCode)} ${esc(typeName.get(f.typeCode)?.name ?? '')}</td><td>${esc(f.title)}</td><td dir="ltr">${esc(f.fileType === 'application/pdf' ? 'PDF' : f.fileType === 'image/vnd.dwg' ? 'DWG' : (f.fileType ?? '—'))}</td></tr>`).join('')}
    </tbody></table>${foot(page++)}</section>` : ''

  const q = p.quote
  const quote = `<section class="page">${band}<div class="box">عرض السعر</div>
    ${q ? `<p><b>رقم العرض:</b> <bdi dir="ltr">${esc(q.quotationNumber)}</bdi> · <b>الزبون:</b> ${esc(q.customerName)}</p>
      <table><thead><tr><th>#</th><th>البند</th><th>الوحدة</th><th>الكمية</th><th>السعر</th><th>المجموع</th></tr></thead><tbody>
      ${q.items.map((it, i) => `<tr><td>${i + 1}</td><td>${esc(it.productName)}</td><td>${esc(it.unit)}</td><td>${n(it.quantity)}</td><td>${n(it.unitPrice)}</td><td>${n(it.totalPrice)}</td></tr>`).join('')}
      </tbody></table>
      <div class="tot"><span>المجموع: <b>${n(q.grandTotal)}</b></span>${q.discountValue ? `<span>الخصم: <b>${n(q.discountValue)}</b></span>` : ''}<span class="net">الصافي: <b>${n(q.netTotal)}</b> د.ع</span></div>`
    : `<p class="muted">${esc(p.quoteNote || 'ماكو عرض سعر مربوط بهذا السيت.')}</p>`}
    ${foot(page++)}</section>`

  return `<!doctype html><html lang="ar" dir="rtl"><head><meta charset="utf-8"><title>${esc(p.title)}</title><style>${fontsCss}
@page { size: A4; margin: 0 }
* { box-sizing: border-box }
body { margin: 0; font-family: 'Tajawal', sans-serif; color: #1e293b; -webkit-print-color-adjust: exact; print-color-adjust: exact; background: #e2e8f0 }
.page { width: 210mm; height: 297mm; padding: 12mm 12mm 16mm; position: relative; overflow: hidden; background: #fff; page-break-after: always; margin: 0 auto; display: flex; flex-direction: column; gap: 6mm }
.band { background: ${NAVY}; color: #fff; border-radius: 6mm; padding: 4mm 6mm; display: flex; align-items: center; gap: 5mm; font-size: 11px; font-weight: 700 }
.band img { width: 18mm; height: 18mm; border-radius: 50%; background: #fff; object-fit: contain; padding: 1.5mm }
.cover-body { flex: 1; display: flex; flex-direction: column; justify-content: center; align-items: center; text-align: center }
.cover-body small { letter-spacing: 4px; color: #94a3b8 } .cover-body h1 { color: ${NAVY}; font-size: 34px; margin: 4mm 0 }
.cover-body p { margin: 1mm 0; font-size: 15px }
.muted { color: #64748b } .box { border: 1px solid ${NAVY}; border-radius: 3mm; padding: 2mm 6mm; align-self: flex-start; font-weight: 700; color: ${NAVY} }
.contents { list-style: none; margin: 0; padding: 0; display: grid; gap: 3mm }
.contents li { display: grid; grid-template-columns: 10mm 1fr 1fr; gap: 3mm; padding: 2.5mm 4mm; border-radius: 2mm; background: #f1f5f9; color: #94a3b8 }
.contents li.on { background: #eef0fb; color: ${NAVY} } .contents li span[dir=ltr] { text-align: left }
.ph { display: flex; align-items: center; gap: 4mm } .ph b { display: block; color: ${NAVY}; font-size: 16px } .ph small { color: #64748b }
.tag { background: ${NAVY}; color: #fff; border-radius: 2mm; padding: 1.5mm 3mm; font-weight: 700 }
.img { flex: 1; display: flex; align-items: center; justify-content: center; border: 1px solid #e2e8f0; border-radius: 3mm; overflow: hidden; min-height: 0 }
.img img { max-width: 100%; max-height: 100%; object-fit: contain }
table { width: 100%; border-collapse: collapse; font-size: 12px } th { background: ${NAVY}; color: #fff; padding: 2mm } td { border-bottom: 1px solid #e2e8f0; padding: 2mm; text-align: center }
.tot { display: flex; gap: 6mm; justify-content: flex-end; font-size: 13px } .tot .net b { color: ${NAVY}; font-size: 16px }
.foot { position: absolute; bottom: 6mm; left: 12mm; right: 12mm; display: flex; justify-content: space-between; font-size: 10px; color: #94a3b8; border-top: 1px solid #e2e8f0; padding-top: 2mm }
@media print { body { background: #fff } .page { margin: 0 } }
</style></head><body>${cover}${contents}${pages}${attach}${p.noQuote ? '' : quote}</body></html>`.replace(/src="LOGO"/g, `src="${esc(abs(logoUrl))}"`)
}
