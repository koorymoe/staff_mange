import { COMPANY_ADDRESS, COMPANY_LEGAL_AR, COMPANY_PHONE, COMPANY_UNITED_AR, COMPANY_UNITED_SUB } from '../company'
import type { SheetExtraField, SheetSlot } from '../pages/design/planTypes'
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
  sheetExtra?: SheetExtraField[] | null
}

// ═══ (ل) 11-22 دازاي: الفورما الجديدة — نفس نموذجه حرفياً ═══
// لوغو · اسم الشركة · (اسم المشروع، عنوان المشروع) · (عنوان الحزمة، اسم اللوحة) · (المقياس، التاريخ)
// · الملاحظات · رقم الصفحة. والخانات الإضافية تنحط بعد القسم الي تختاره المصممة.
// مصدر واحد للـPDF والمعاينة بالنظام حتى ما يختلفون.
export type SheetSection = { key: SheetSlot | 'page' | 'extra'; rows: [string, string][]; grow?: boolean }
export const SHEET_SLOTS: { key: SheetSlot; label: string }[] = [
  { key: 'company', label: 'بعد اسم الشركة' }, { key: 'project', label: 'بعد المشروع' },
  { key: 'package', label: 'بعد الحزمة واللوحة' }, { key: 'scale', label: 'بعد المقياس والتاريخ' }, { key: 'notes', label: 'بعد الملاحظات' },
]
export function sheetSections(v: { projectName?: string | null; projectAddress?: string | null; packageName?: string; boardName: string; scale?: string | null; date: string; notes?: string | null; page?: string; extra?: SheetExtraField[] | null }): SheetSection[] {
  const d = (x?: string | null) => (x || '').trim() || '—'
  const base: SheetSection[] = [
    { key: 'company', rows: [] },
    { key: 'project', rows: [['اسم المشروع', d(v.projectName)], ['عنوان المشروع', d(v.projectAddress)]] },
    { key: 'package', rows: [['عنوان الحزمة', d(v.packageName)], ['اسم اللوحة', d(v.boardName)]] },
    { key: 'scale', rows: [['المقياس', d(v.scale)], ['التاريخ', v.date]] },
    { key: 'notes', rows: [['الملاحظات', d(v.notes)]], grow: true },
  ]
  const out: SheetSection[] = []
  for (const sec of base) {
    out.push(sec)
    for (const x of v.extra ?? []) if (x.after === sec.key && x.label.trim()) out.push({ key: 'extra', rows: [[x.label.trim(), d(x.value)]] })
  }
  out.push({ key: 'page', rows: [['رقم الصفحة', v.page ?? '—']] })
  return out
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
  projectAddress?: string | null
  groups: PlanPrintGroup[]
  files: PlanPrintFile[]
  quote: PlanPrintQuote | null
  /** سبب غياب عرض السعر (ماكو مشروع، ماكو عرض، ما عنده صلاحية) */
  quoteNote?: string
  // (ل) 11-10: «ملفين»: PDF التصاميم بلا عرض السعر (العرض PDF وحده بقالبه)
  noQuote?: boolean
  date: Date
}

const COMPANY = COMPANY_LEGAL_AR
// (ل) 11-12: ألوان دليل الهوية الي دزه دازاي — كحلي أساسي، ذهبي، أزرق ثانوي، خلفية، نص
const NAVY = '#0B2D5B'
const GOLD = '#D4AF37'
// موجات الغلاف (نفس بانر الهوية): كحلي غامق وموجة ذهبية تحت
// (ل) 11-22 دازاي «سويلنه هاي بدل القديمة»: غلاف فخم — شريط ذهبي بزخرفة إسلامية فوگ اليسار، مبنى زجاجي
// يمين، مخطط ملفوف يسار، وأيقونات التصميم بالوسط. كله SVG حتى ينطبع حاد بالـA3.
const WAVES = `<svg class="pc-waves" viewBox="0 0 1123 794" preserveAspectRatio="xMidYMid slice" aria-hidden="true">
  <defs>
    <radialGradient id="bg" cx=".55" cy=".45" r=".8"><stop offset="0" stop-color="#1b3f78"/><stop offset=".55" stop-color="${NAVY}"/><stop offset="1" stop-color="#061a38"/></radialGradient>
    <linearGradient id="g" x1="0" x2="1"><stop offset="0" stop-color="#9c7a1e"/><stop offset=".45" stop-color="#f3dc8a"/><stop offset="1" stop-color="${GOLD}"/></linearGradient>
    <linearGradient id="gv" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#f3dc8a"/><stop offset="1" stop-color="#9c7a1e"/></linearGradient>
    <linearGradient id="glass" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#2a5a9e" stop-opacity=".55"/><stop offset=".6" stop-color="#123a70" stop-opacity=".35"/><stop offset="1" stop-color="${GOLD}" stop-opacity=".25"/></linearGradient>
    <linearGradient id="refl" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="${GOLD}" stop-opacity=".25"/><stop offset="1" stop-color="${GOLD}" stop-opacity="0"/></linearGradient>
    <linearGradient id="roll" x1="0" y1="0" x2="0" y2="1"><stop offset="0" stop-color="#3b6cb0"/><stop offset=".5" stop-color="#1a3f75"/><stop offset="1" stop-color="#0a2550"/></linearGradient>
    <pattern id="lat" width="34" height="34" patternUnits="userSpaceOnUse"><g fill="none" stroke="${GOLD}" stroke-width="1.1" opacity=".75">
      <path d="M17 2 L21 13 L32 17 L21 21 L17 32 L13 21 L2 17 L13 13 Z"/><circle cx="17" cy="17" r="5"/><path d="M0 0 L5 5 M34 0 L29 5 M0 34 L5 29 M34 34 L29 29"/></g></pattern>
    <pattern id="bp" width="16" height="16" patternUnits="userSpaceOnUse"><path d="M16 0H0V16" fill="none" stroke="#8fb3e8" stroke-opacity=".35" stroke-width=".6"/></pattern>
    <clipPath id="band"><path d="M0 0 H330 C250 120 150 260 0 420 Z"/></clipPath>
    <filter id="glow"><feGaussianBlur stdDeviation="6"/></filter>
  </defs>
  <rect width="1123" height="794" fill="url(#bg)"/>
  <!-- الشريط الذهبي والزخرفة -->
  <g clip-path="url(#band)"><rect width="400" height="450" fill="#0a2550"/><rect width="400" height="450" fill="url(#lat)"/></g>
  <path d="M330 0 C250 120 150 260 0 420" fill="none" stroke="url(#g)" stroke-width="10"/>
  <path d="M380 0 C290 150 170 300 0 470" fill="none" stroke="url(#g)" stroke-width="4" opacity=".8"/>
  <path d="M420 0 C330 170 190 330 0 520" fill="none" stroke="${GOLD}" stroke-width="1.5" opacity=".5"/>
  <path d="M330 0 C250 120 150 260 0 420" fill="none" stroke="#f3dc8a" stroke-width="18" opacity=".25" filter="url(#glow)"/>
  <!-- المبنى الزجاجي -->
  <polygon points="830,610 1123,560 1123,330 900,400 830,430" fill="url(#glass)" stroke="#8fb3e8" stroke-opacity=".5" stroke-width="1.2"/>
  <line x1="867" y1="604" x2="867" y2="418" stroke="#a9c6f0" stroke-opacity=".35"/>
  <line x1="903" y1="598" x2="903" y2="399" stroke="#a9c6f0" stroke-opacity=".35"/>
  <line x1="940" y1="591" x2="940" y2="387" stroke="#a9c6f0" stroke-opacity=".35"/>
  <line x1="976" y1="585" x2="976" y2="376" stroke="#a9c6f0" stroke-opacity=".35"/>
  <line x1="1013" y1="579" x2="1013" y2="364" stroke="#a9c6f0" stroke-opacity=".35"/>
  <line x1="1050" y1="572" x2="1050" y2="353" stroke="#a9c6f0" stroke-opacity=".35"/>
  <line x1="1086" y1="566" x2="1086" y2="341" stroke="#a9c6f0" stroke-opacity=".35"/>
  <line x1="830" y1="580" x2="1123" y2="522" stroke="#a9c6f0" stroke-opacity=".3"/>
  <line x1="830" y1="550" x2="1123" y2="483" stroke="#a9c6f0" stroke-opacity=".3"/>
  <line x1="830" y1="520" x2="1123" y2="445" stroke="#a9c6f0" stroke-opacity=".3"/>
  <line x1="830" y1="490" x2="1123" y2="407" stroke="#a9c6f0" stroke-opacity=".3"/>
  <line x1="830" y1="460" x2="1123" y2="368" stroke="#a9c6f0" stroke-opacity=".3"/>
  <polygon points="760,620 830,610 830,430 760,450" fill="#123a70" fill-opacity=".7" stroke="#8fb3e8" stroke-opacity=".4"/>
  <line x1="760" y1="592" x2="830" y2="580" stroke="${GOLD}" stroke-opacity="0.35"/>
  <line x1="760" y1="563" x2="830" y2="550" stroke="${GOLD}" stroke-opacity="0.25"/>
  <line x1="760" y1="535" x2="830" y2="520" stroke="${GOLD}" stroke-opacity="0.35"/>
  <line x1="760" y1="507" x2="830" y2="490" stroke="${GOLD}" stroke-opacity="0.25"/>
  <line x1="760" y1="478" x2="830" y2="460" stroke="${GOLD}" stroke-opacity="0.35"/>
  <polygon points="880,560 1123,525" fill="none"/><line x1="860" y1="520" x2="1123" y2="480" stroke="${GOLD}" stroke-width="3" opacity=".8"/><line x1="860" y1="470" x2="1123" y2="425" stroke="${GOLD}" stroke-width="2" opacity=".6"/>
  <rect x="760" y="618" width="363" height="40" fill="url(#refl)"/>
  <!-- الأرضية والخطوط الذهبية -->
  <path d="M0 650 L1123 590" stroke="${GOLD}" stroke-width="2" opacity=".7"/>
  <path d="M260 794 L1123 610" stroke="url(#g)" stroke-width="1.5" opacity=".5"/>
  <path d="M0 700 Q560 560 1123 690" fill="none" stroke="#8fb3e8" stroke-opacity=".12" stroke-width="40"/>
  <!-- المخطط الملفوف -->
  <g transform="rotate(14 160 620)">
    <rect x="-40" y="575" width="300" height="80" rx="4" fill="url(#roll)"/><rect x="-40" y="575" width="300" height="80" fill="url(#bp)"/>
    <path d="M20 600 h80 v30 h-50 M130 590 v50 M160 610 h70" fill="none" stroke="#cfe0fa" stroke-opacity=".5"/>
    <ellipse cx="260" cy="615" rx="18" ry="40" fill="#0a2550" stroke="url(#gv)" stroke-width="7"/>
    <ellipse cx="260" cy="615" rx="8" ry="20" fill="#061a38"/>
    <ellipse cx="260" cy="615" rx="30" ry="50" fill="${GOLD}" opacity=".18" filter="url(#glow)"/>
  </g>
  <path d="M300 700 l120 -20 l60 50 l-120 22 Z" fill="url(#bp)" opacity=".7"/>
</svg>`
const ICO = (d: string) => `<span class="pc-ico"><svg viewBox="0 0 24 24" fill="none" stroke="${GOLD}" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">${d}</svg></span>`
const ICONS = [
  '<path d="M12 3 21 8 12 13 3 8Z"/><path d="M3 12l9 5 9-5"/><path d="M3 16l9 5 9-5"/>',
  '<rect x="3" y="4" width="18" height="14"/><path d="M3 11h8V4M11 11v7M15 4v6h6"/><path d="M14 20l6-6 2 2-6 6h-2z"/>',
  '<circle cx="12" cy="4" r="1.6"/><path d="M11 5.5 5 21M13 5.5 19 21M7.5 15h9"/>',
  '<path d="M12 2 21 7v10l-9 5-9-5V7z"/><path d="M3 7l9 5 9-5M12 12v10"/>',
].map(ICO).join('')
const IMAGE_TYPES = ['image/jpeg', 'image/png', 'image/webp']

const n = (v: number) => Math.round(v).toLocaleString('en-US')
const abs = (u: string) => (/^(https?:|data:|blob:)/.test(u) ? u : new URL(u, location.origin).href)
export const isImageFile = (f: { fileType?: string | null }) => IMAGE_TYPES.includes(f.fileType ?? '')

const band = `<div class="band"><img src="LOGO" alt=""><span>${COMPANY}</span></div>`
// (ل) 11-22 دازاي: «هنا يكون عنوان الشركة والرقم… بتصميم مرتب» — شريط كحلي رفيع بأيقونات، بلا كلمات زايدة
const PIN = '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M12 2a7 7 0 0 0-7 7c0 5.2 7 13 7 13s7-7.8 7-13a7 7 0 0 0-7-7zm0 9.5A2.5 2.5 0 1 1 12 6.5a2.5 2.5 0 0 1 0 5z"/></svg>'
const TEL = '<svg viewBox="0 0 24 24" aria-hidden="true"><path d="M6.6 10.8a15.1 15.1 0 0 0 6.6 6.6l2.2-2.2a1 1 0 0 1 1-.25 11.4 11.4 0 0 0 3.6.57 1 1 0 0 1 1 1V20a1 1 0 0 1-1 1A17 17 0 0 1 3 4a1 1 0 0 1 1-1h3.5a1 1 0 0 1 1 1c0 1.25.2 2.45.57 3.57a1 1 0 0 1-.25 1z"/></svg>'
const foot = (page: number) => `<div class="foot">${COMPANY_ADDRESS ? `<span class="fi">${PIN}<span>${esc(COMPANY_ADDRESS)}</span></span>` : ''}<span class="fi">${TEL}<bdi dir="ltr">${COMPANY_PHONE}</bdi></span><b class="fp" dir="ltr">${String(page).padStart(2, '0')}</b></div>`

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

  // «Package Contents»: (ل) 11-22 دازاي: «اذا المخططات المعمارية مموجودة راسا تنحذف» — بس الأقسام الي بيها ملفات
  // (ل) 11-22 دازاي: «تسلسل جديد حسب الموجود» — الأقسام الموجودة بالسيت تترقّم 1، 2، 3… (مو رقم الكتالوج)
  const shownGroups = p.groups.filter((g) => g.types.some((t) => present.has(t.code)))
  const seqNo = new Map(shownGroups.map((g, i) => [g.no, i + 1]))
  const contents = `<section class="page">${band}<div class="box">Package Contents</div><ol class="contents">
    ${shownGroups.map((g) => `<li class="on"><b>${seqNo.get(g.no)}</b><span dir="ltr">${esc(g.en)}</span><span>${esc(g.name)}</span></li>`).join('')}
  </ol>${foot(page++)}</section>`

  // (ل) 11-12 دازاي: «قبل كل تصميم غلاف يوضح هذا تصميم شنو — مرتب وحلو، بألوان فخمة، وبيه اسم المخطط»،
  // و«جوه كل تصميم فورما: لوغو الأماني، اسم المخطط، الملاحظات، السكيل، التاريخ» — مكانها تختاره المصممة.
  // (ل) 11-12 دازاي: «غلاف قبل كل نوع مو قبل كل صورة» — صور النوع الواحد (مثلاً ٧ لقطات رئيسية)
  // ورا غلاف واحد. و«الترقيم بالـPDF يتسلسل»: رقم القسم يبقى، والنوع يتسلسل حسب الموجود
  // بالملف (10.2 و10.4 ← 10.1 و10.2) — مو أرقام النظام الداخلية.
  const typeOrder: string[] = []
  const byType = new Map<string, PlanPrintFile[]>()
  for (const f of images) {
    if (!byType.has(f.typeCode)) { byType.set(f.typeCode, []); typeOrder.push(f.typeCode) }
    byType.get(f.typeCode)!.push(f)
  }
  const seqInGroup = new Map<number, number>()
  const pages = typeOrder.map((code, i) => {
    const fs = byType.get(code)!
    const t = typeName.get(code)
    const gno = t?.group.no
    const k = gno == null ? 0 : (seqInGroup.get(gno) ?? 0) + 1
    if (gno != null) seqInGroup.set(gno, k)
    const seq = gno == null ? code : `${seqNo.get(gno) ?? gno}.${k}`
    const coverName = (fs.length === 1 && (fs[0].sheetName || '').trim()) || t?.name || fs[0].title || code
    // (ل) 11-22 دازاي: سطر القسم وعدد اللوحات انحذفن من الغلاف
    const cover = `<section class="page land pcover">${WAVES}
      <div class="pc-top"><img src="LOGO" alt=""><div><b>شركة الأماني</b><small dir="ltr">Al-Amani Company</small></div>
        <span class="pc-no" dir="ltr"><i></i>${String(i + 1).padStart(2, '0')} / ${String(typeOrder.length).padStart(2, '0')}</span></div>
      <div class="pc-body"><p class="pc-codeline"><i></i><span class="pc-code" dir="ltr">${esc(seq)}</span><i></i></p>
        <h2>${esc(coverName)}</h2>
        <p class="pc-tag"><i></i><span>${esc(p.title)}${p.projectName ? ` - ${esc(p.projectName)}` : ''}</span><i></i></p>
        <div class="pc-icons">${ICONS}</div></div>
      <div class="pc-foot"><span dir="ltr">${esc(sheetDate(fs[0], p.date))}</span><span>معاً ... نحو مستقبل أفضل</span></div>
      ${foot(page++)}</section>`
    const sheets = fs.map((f, j) => {
      const name = (f.sheetName || '').trim() || (fs.length > 1 ? `${t?.name ?? code} ${j + 1}` : t?.name || f.title || code)
      const date = sheetDate(f, p.date)
      // (ل) 11-22 دازاي: الفورما الجديدة (sheetSections) — رقم الصفحة آخر خانة
      const secs = sheetSections({ projectName: p.projectName, projectAddress: p.projectAddress, packageName: t?.group.name, boardName: name,
        scale: f.sheetScale, date, notes: f.sheetNotes, page: String(page).padStart(2, '0'), extra: f.sheetExtra })
      const ltr = new Set(['المقياس', 'التاريخ', 'رقم الصفحة'])
      const block = `<div class="tb"><div class="tb-logo"><img src="LOGO" alt=""></div>
        ${secs.map((sec) => sec.key === 'company'
          ? `<div class="tb-co"><b>${esc(COMPANY_UNITED_AR)}</b><small>${esc(COMPANY_UNITED_SUB)}</small></div>`
          : `<div class="tb-sec${sec.grow ? ' grow' : ''}${sec.key === 'page' ? ' tb-page' : ''}">${sec.rows.map(([k, v]) => `<div class="tb-cell"><b>${esc(k)}:</b><span${ltr.has(k) ? ' dir="ltr"' : ''}>${esc(v)}</span></div>`).join('')}</div>`).join('')}</div>`
      const pos = f.sheetPos === 'bottom' ? 'bottom' : 'right'
      page++
      return `<section class="page land plan sheet-${pos}">
        ${pos === 'right' ? block : ''}<div class="img"><img src="${esc(abs(fileUrl(f.fileKey)))}" alt=""></div>${pos === 'bottom' ? block : ''}
        </section>`
    }).join('')
    return cover + sheets
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
/* (ل) 11-12 دازاي: «الملف لازم A3 — كل الصفحات نفس الحجم». حجم واحد لكل الصفحات: بدونه
   نافذة الطباعة تحشر الصفحات بالورقة المختارة وتطلع الأغلفة والصور مقصوصة ع النص */
@page { size: A3 landscape; margin: 0 }
* { box-sizing: border-box }
body { margin: 0; font-family: 'Tajawal', sans-serif; color: #1e293b; -webkit-print-color-adjust: exact; print-color-adjust: exact; background: #e2e8f0 }
.page { width: 420mm; height: 297mm; padding: 16mm 20mm 26mm; break-inside: avoid; position: relative; overflow: hidden; background: #fff; page-break-after: always; margin: 0 auto; display: flex; flex-direction: column; gap: 8mm }
.band { background: ${NAVY}; color: #fff; border-radius: 6mm; padding: 5mm 8mm; display: flex; align-items: center; gap: 6mm; font-size: 15px; font-weight: 700 }
.band img { width: 24mm; height: 24mm; border-radius: 50%; background: #fff; object-fit: contain; padding: 1.5mm }
.cover-body { flex: 1; display: flex; flex-direction: column; justify-content: center; align-items: center; text-align: center }
.cover-body small { letter-spacing: 4px; color: #94a3b8 } .cover-body h1 { color: ${NAVY}; font-size: 52px; margin: 6mm 0 }
.cover-body p { margin: 1.5mm 0; font-size: 20px }
.muted { color: #64748b } .box { border: 1px solid ${NAVY}; border-radius: 3mm; padding: 3mm 8mm; font-size: 18px; align-self: flex-start; font-weight: 700; color: ${NAVY} }
.contents { list-style: none; margin: 0; padding: 0; display: grid; gap: 3mm }
.contents { grid-template-columns: 1fr 1fr; gap: 4mm 8mm } .contents li { display: grid; grid-template-columns: 12mm 1fr 1fr; gap: 4mm; padding: 4mm 6mm; font-size: 16px; border-radius: 2mm; background: #f1f5f9; color: #94a3b8 }
.contents li.on { background: #eef0fb; color: ${NAVY} } .contents li span[dir=ltr] { text-align: left }
.ph { display: flex; align-items: center; gap: 4mm } .ph b { display: block; color: ${NAVY}; font-size: 16px } .ph small { color: #64748b }
.tag { background: ${NAVY}; color: #fff; border-radius: 2mm; padding: 1.5mm 3mm; font-weight: 700 }
.img { flex: 1; display: flex; align-items: center; justify-content: center; border: 1px solid #e2e8f0; border-radius: 3mm; overflow: hidden; min-height: 0 }
.img img { max-width: 100%; max-height: 100%; object-fit: contain }
table { width: 100%; border-collapse: collapse; font-size: 15px } th { background: ${NAVY}; color: #fff; padding: 2mm } td { border-bottom: 1px solid #e2e8f0; padding: 2mm; text-align: center }
.tot { display: flex; gap: 6mm; justify-content: flex-end; font-size: 13px } .tot .net b { color: ${NAVY}; font-size: 16px }
.foot { position: absolute; bottom: 7mm; left: 20mm; right: 20mm; display: flex; align-items: center; gap: 10mm; background: ${NAVY}; color: #fff; border-radius: 3mm; padding: 2.5mm 6mm; font-size: 15px; font-weight: 700; border-bottom: 1mm solid ${GOLD} }
.foot .fi { display: flex; align-items: center; gap: 2.5mm } .foot svg { width: 5mm; height: 5mm; fill: ${GOLD} }
.foot .fp { margin-right: auto; background: ${GOLD}; color: ${NAVY}; border-radius: 2mm; padding: 0.5mm 3mm }
.page.land { padding: 14mm 16mm 20mm }
/* ── غلاف المخطط: كحلي غامق وذهبي ── */
.page.land.pcover { background: ${NAVY}; color: #fff; justify-content: space-between; padding: 18mm 24mm 18mm }
.pc-waves { position: absolute; inset: 0; width: 100%; height: 100%; z-index: 0 }
.pcover > *:not(.pc-waves) { position: relative; z-index: 1 }
.pc-top { display: flex; align-items: center; gap: 5mm }
.pc-top img { width: 30mm; height: 30mm; border-radius: 50%; background: #fff; padding: 1.5mm; box-shadow: 0 0 0 1mm ${GOLD}, 0 0 8mm rgba(212,175,55,.35) }
.pc-top b { display: block; font-size: 30px } .pc-top small { color: ${GOLD}; font-size: 17px; font-weight: 700 }
.pc-no { margin-inline-start: auto; font-size: 15px; letter-spacing: 1px; color: #e2e8f0; display: flex; align-items: center; gap: 4mm }
.pc-no i { width: 14mm; height: 0.5mm; background: ${GOLD} }
.pc-body { text-align: center; margin-top: -26mm }
.pc-codeline, .pc-tag { display: flex; align-items: center; justify-content: center; gap: 6mm; margin: 0 }
.pc-codeline i { width: 18mm; height: 0.5mm; background: ${GOLD} }
.pc-code { display: inline-block; background: linear-gradient(180deg, #f3dc8a, ${GOLD} 55%, #9c7a1e); color: ${NAVY}; border-radius: 10mm; padding: 1.5mm 10mm; font-size: 26px; font-weight: 800; box-shadow: 0 1mm 4mm rgba(0,0,0,.35) }
.pc-body h2 { font-size: 84px; margin: 3mm 0 0; line-height: 1.25; background: linear-gradient(180deg, #fff 30%, #cdd6e4); -webkit-background-clip: text; background-clip: text; color: transparent; filter: drop-shadow(0 3px 6px rgba(0,0,0,.45)) }
.pc-body h2::after { content: ''; display: block; width: 60mm; height: 0.8mm; margin: 2mm auto 0; background: radial-gradient(${GOLD}, transparent 70%) }
.pc-tag { margin: 5mm 0 0; color: #fff; font-size: 24px; font-weight: 700 }
.pc-tag i { width: 30mm; height: 0.5mm; background: ${GOLD} }
.pc-icons { display: flex; justify-content: center; gap: 7mm; margin-top: 9mm }
.pc-ico { width: 24mm; height: 24mm; border-radius: 5mm; display: grid; place-items: center; background: linear-gradient(160deg, rgba(255,255,255,.14), rgba(255,255,255,.03)); border: 0.4mm solid rgba(212,175,55,.55); box-shadow: inset 0 0 4mm rgba(212,175,55,.15), 0 1.5mm 4mm rgba(0,0,0,.4) }
.pc-ico svg { width: 13mm; height: 13mm }
.pc-foot { display: flex; justify-content: space-between; gap: 6mm; font-size: 19px; color: #fff; font-weight: 700 }
.pc-foot span { border-inline-start: 1mm solid ${GOLD}; padding-inline-start: 4mm }
.pcover .foot { display: none }
/* ── صفحة المخطط ويا الفورما ── */
.plan { gap: 4mm } .plan.sheet-right { flex-direction: row }
/* ⚠️ grid مو flex-column: ويا flex-column كروم يطبع صفحة الفورما-تحت بالطول (portrait) ويقص المخطط */
.plan.sheet-bottom { display: grid; grid-template-rows: minmax(0, 1fr) 36mm; gap: 4mm }
.plan .img { flex: 1; min-width: 0 }
.tb { display: flex; border: 0.5mm solid ${NAVY}; border-radius: 1.5mm; overflow: hidden; background: #fff; font-size: 10px; color: #1e293b }
/* (ل) 11-22 دازاي «قلل من العرض»: 64 ← 50 ملم */
.sheet-right .tb { flex-direction: column; width: 50mm; flex: 0 0 50mm } .sheet-bottom .tb { flex-direction: row; height: 36mm }
.tb > * { border-color: ${NAVY} } .sheet-right .tb > * + * { border-top: 0.3mm solid ${NAVY} } .sheet-bottom .tb > * + * { border-inline-start: 0.3mm solid ${NAVY} }
.tb-logo { display: grid; place-items: center; padding: 3mm 2mm }
.tb-logo img { width: 26mm; height: 26mm; object-fit: contain }
.sheet-bottom .tb-logo { width: 34mm } .sheet-bottom .tb-logo img { width: 24mm; height: 24mm }
.tb-co { padding: 2mm 2.5mm; text-align: center; display: flex; flex-direction: column; justify-content: center; gap: 1mm }
.tb-co b { font-size: 15px; color: ${NAVY} } .tb-co small { font-size: 8.5px; line-height: 1.5; color: #334155 }
.sheet-bottom .tb-co { width: 52mm }
.tb-sec { padding: 2mm 2.5mm; display: flex; flex-direction: column; gap: 2.5mm; min-width: 0 }
.sheet-right .tb-sec { min-height: 20mm } .tb-sec.grow { flex: 1 } .sheet-bottom .tb-sec { flex: 1 }
.tb-cell { display: flex; flex-direction: column; gap: 0.6mm; min-width: 0 }
.tb-cell b { color: ${NAVY}; font-size: 11.5px } .tb-cell span { font-size: 12.5px; white-space: pre-wrap; overflow-wrap: anywhere }
.tb-page { background: #f8fafc; min-height: 0 !important } .tb-page span { font-size: 20px; font-weight: 800; color: ${NAVY} }
@media print { html, body { margin: 0; background: #fff } .page { margin: 0 } }
</style></head><body>${cover}${contents}${pages}${attach}${p.noQuote ? '' : quote}</body></html>`.replace(/src="LOGO"/g, `src="${esc(abs(logoUrl))}"`)
}
