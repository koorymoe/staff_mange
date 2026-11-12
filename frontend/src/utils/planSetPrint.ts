// ═══ 📐 سيت المخططات — ملف الـPDF (ك) 11-10 ═══
// بطلب دازاي: «تسوي تنزيل ملف ينزل ملف PDF بالألوان الهوية البصرية مال الشركة بي كل
// التصاميم بالإضافة الى عرض السعر». الشكل نفس ورقة «Package Contents» مال الشركة:
// شريط كحلي وشعار. الطباعة بنفس إطار عرض السعر (printProjectQuote) — المستخدم يختار
// «حفظ كـPDF». الصور تنطبع صفحة لكل مخطط؛ PDF وDWG ما ينطبعن جوّه، فيطلعن بقائمة المرفقات.
import fontsCss from '../fonts.css?inline'
import { fileUrl } from '../api'
import { esc } from './printIdentity'

export type PlanPrintGroup = { no: number; name: string; en: string; types: { code: string; name: string }[] }
export type PlanPrintFile = {
  typeCode: string; title: string; fileKey: string; fileType?: string | null
  // (ل) 11-12: فورما المخطط + تاريخه
  createdAt?: string; sheetPos?: 'right' | 'bottom'; sheetName?: string | null; sheetNotes?: string | null; sheetScale?: string | null
}

/** تاريخ المخطط بالفورما = يوم رفعه (بتوقيت بغداد) */
export function sheetDate(f: { createdAt?: string }, fallback = new Date()) {
  return (f.createdAt ? new Date(f.createdAt) : fallback).toLocaleDateString('en-GB', { timeZone: 'Asia/Baghdad' })
}
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
// (ل) 11-12: ألوان دليل الهوية الي دزه دازاي — كحلي أساسي، ذهبي، أزرق ثانوي، خلفية، نص
const NAVY = '#0B2D5B'
const GOLD = '#D4AF37'
const BLUE = '#1E4A8A'
// موجات الغلاف (نفس بانر الهوية): كحلي غامق وموجة ذهبية تحت
const WAVES = `<svg class="pc-waves" viewBox="0 0 1123 794" preserveAspectRatio="none" aria-hidden="true">
  <defs><linearGradient id="g" x1="0" x2="1"><stop offset="0" stop-color="#f3dc8a"/><stop offset=".5" stop-color="${GOLD}"/><stop offset="1" stop-color="#9c7a1e"/></linearGradient>
  <linearGradient id="b" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="${BLUE}" stop-opacity=".55"/><stop offset="1" stop-color="${NAVY}" stop-opacity="0"/></linearGradient></defs>
  <path d="M0 0 H1123 V170 C900 260 700 120 470 180 C270 232 120 160 0 210 Z" fill="url(#b)"/>
  <path d="M0 640 C220 560 430 700 700 650 C880 616 1010 560 1123 590 V794 H0 Z" fill="#071f42"/>
  <path d="M0 655 C230 575 430 715 700 665 C880 631 1010 575 1123 605" fill="none" stroke="url(#g)" stroke-width="7"/>
  <path d="M0 690 C250 620 470 745 760 700 C930 674 1040 630 1123 650" fill="none" stroke="url(#g)" stroke-width="2" opacity=".7"/>
</svg>`
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

  // (ل) 11-12 دازاي: «قبل كل تصميم غلاف يوضح هذا تصميم شنو — مرتب وحلو، بألوان فخمة، وبيه اسم المخطط»،
  // و«جوه كل تصميم فورما: لوغو الأماني، اسم المخطط، الملاحظات، السكيل، التاريخ» — مكانها تختاره المصممة.
  const pages = images.map((f, i) => {
    const t = typeName.get(f.typeCode)
    const name = (f.sheetName || '').trim() || t?.name || f.title || f.typeCode
    const group = t ? `${t.group.no}. ${t.group.name}` : ''
    const date = sheetDate(f, p.date)
    const cover = `<section class="page land pcover">${WAVES}
      <div class="pc-top"><img src="LOGO" alt=""><div><b>شركة الأماني</b><small dir="ltr">Al-Amani Company</small></div>
        <span class="pc-no" dir="ltr">${String(i + 1).padStart(2, '0')} / ${String(images.length).padStart(2, '0')}</span></div>
      <div class="pc-body"><span class="pc-code" dir="ltr">${esc(f.typeCode)}</span>
        <h2>${esc(name)}</h2>${group ? `<p class="pc-group">${esc(group)}${t?.group.en ? ` · <bdi dir="ltr">${esc(t.group.en)}</bdi>` : ''}</p>` : ''}
        <p class="pc-tag"><i></i><span>${esc(p.title)}${p.projectName ? ` · ${esc(p.projectName)}` : ''}</span><i></i></p></div>
      <div class="pc-foot"><span dir="ltr">${esc(date)}</span><span>معاً .. نحو مستقبل أفضل</span></div>
      ${foot(page++)}</section>`
    const cells = [['اسم المخطط', name], ['الملاحظات', (f.sheetNotes || '').trim() || '—'], ['السكيل', (f.sheetScale || '').trim() || '—'], ['التاريخ', date]]
    const block = `<div class="tb"><div class="tb-logo"><img src="LOGO" alt=""><small>الأماني</small></div>
      ${cells.map(([k, v]) => `<div class="tb-cell"><b>${k}</b><span${k === 'السكيل' || k === 'التاريخ' ? ' dir="ltr"' : ''}>${esc(v)}</span></div>`).join('')}</div>`
    const pos = f.sheetPos === 'bottom' ? 'bottom' : 'right'
    const sheet = `<section class="page land plan sheet-${pos}">
      ${pos === 'right' ? block : ''}<div class="img"><img src="${esc(abs(fileUrl(f.fileKey)))}" alt=""></div>${pos === 'bottom' ? block : ''}
      ${foot(page++)}</section>`
    return cover + sheet
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
@page land { size: A4 landscape; margin: 0 }
.page.land { page: land; width: 297mm; height: 210mm; padding: 10mm 10mm 14mm }
/* ── غلاف المخطط: كحلي غامق وذهبي ── */
.page.land.pcover { background: linear-gradient(135deg, #123a70 0%, ${NAVY} 45%, #071f42 100%); color: #fff; justify-content: space-between; padding: 14mm 18mm 16mm }
.pc-waves { position: absolute; inset: 0; width: 100%; height: 100%; z-index: 0 }
.pcover > *:not(.pc-waves) { position: relative; z-index: 1 }
.pc-top { display: flex; align-items: center; gap: 5mm }
.pc-top img { width: 20mm; height: 20mm; border-radius: 50%; background: #fff; padding: 1.5mm; box-shadow: 0 0 0 0.8mm ${GOLD} }
.pc-top b { display: block; font-size: 16px } .pc-top small { color: ${GOLD}; font-size: 11px; font-weight: 700 }
.pc-no { margin-inline-start: auto; font-size: 13px; letter-spacing: 2px; color: ${GOLD} }
.pc-body { text-align: center; margin-top: -10mm }
.pc-code { display: inline-block; background: ${GOLD}; color: ${NAVY}; border-radius: 10mm; padding: 1mm 7mm; font-weight: 800; letter-spacing: 2px }
.pc-body h2 { font-size: 42px; margin: 6mm 0 3mm; line-height: 1.25; text-shadow: 0 2px 14px rgba(0,0,0,.35) }
.pc-group { color: ${GOLD}; font-size: 16px; margin: 0; font-weight: 700 }
.pc-tag { display: flex; align-items: center; justify-content: center; gap: 4mm; margin: 7mm 0 0; color: #e2e8f0; font-size: 14px }
.pc-tag i { width: 22mm; height: 0.5mm; background: ${GOLD} }
.pc-foot { display: flex; justify-content: space-between; gap: 6mm; font-size: 11px; color: #fff; font-weight: 700 }
.pcover .foot { display: none }
/* ── صفحة المخطط ويا الفورما ── */
.plan { gap: 4mm } .plan.sheet-right { flex-direction: row }
/* ⚠️ grid مو flex-column: ويا flex-column كروم يطبع صفحة الفورما-تحت بالطول (portrait) ويقص المخطط */
.plan.sheet-bottom { display: grid; grid-template-rows: minmax(0, 1fr) 26mm; gap: 4mm }
.plan .img { flex: 1; min-width: 0 }
.tb { display: flex; border: 0.5mm solid ${NAVY}; border-radius: 2mm; overflow: hidden; background: #fff; font-size: 10px }
.sheet-right .tb { flex-direction: column; width: 46mm } .sheet-bottom .tb { flex-direction: row; height: 26mm }
.tb-logo { display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 1mm; padding: 2mm; background: ${NAVY}; border-bottom: 1mm solid ${GOLD}; color: #fff; font-weight: 700; font-size: 12px }
.tb-logo img { width: 15mm; height: 15mm; border-radius: 50%; background: #fff; padding: 1mm }
.sheet-bottom .tb-logo { width: 30mm }
.tb-cell { flex: 1; padding: 2mm 2.5mm; display: flex; flex-direction: column; gap: 1mm; min-width: 0 }
.sheet-right .tb-cell { border-top: 0.3mm solid ${NAVY} } .sheet-bottom .tb-cell { border-inline-start: 0.3mm solid ${NAVY} }
.sheet-right .tb-cell:nth-child(3) { flex: 2 } .sheet-bottom .tb-cell:nth-child(3) { flex: 2 }
.tb-cell b { color: ${NAVY}; font-size: 10.5px } .tb-cell span { font-size: 12px; white-space: pre-wrap; overflow-wrap: anywhere; color: #1e293b }
@media print { body { background: #fff } .page { margin: 0 } }
</style></head><body>${cover}${contents}${pages}${attach}${p.noQuote ? '' : quote}</body></html>`.replace(/src="LOGO"/g, `src="${esc(abs(logoUrl))}"`)
}
