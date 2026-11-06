// ═══ «عرض مشاريع» — صفحة الطباعة (دازاي 11-06) ═══
// نفس تصميم نموذج «مشروع جامعة المكين»: غلاف، نظرة شاملة، فصل لكل فئة بكارتين بالصفحة،
// كشف الكميات والأسعار ويا المبلغ بالحروف، والملحق الفني بأجزائه. A4 وألوان الشركة.
// الأرقام كلها من العرض نفسه — ماكو شي مخترع.
import fontsCss from '../fonts.css?inline'
import { fileUrl } from '../api'
import { arabicAmountWords } from './arabicWords'
import { esc } from './printIdentity'

export type ProjectCover = { label?: string; title?: string; subtitle?: string; city?: string }
export type AnnexItem = { title: string; text: string }
export type AnnexPart = { title: string; items: AnnexItem[] }
export type ProjectItem = {
  productName: string
  modelName?: string | null
  specs?: string | null
  category?: string | null
  imageBase64?: string | null
  unit?: string | null
  quantity: number
  unitPrice: number
}
export type ProjectQuote = {
  quotationNumber?: string
  customerName: string
  cover: ProjectCover
  items: ProjectItem[]
  annex: AnnexPart[]
  date: Date
}

const COMPANY_AR = 'شركة الأماني للتجارة العامة والاستثمارات العقارية'
const COMPANY_AR2 = 'والوكالات التجارية محدودة المسؤولية'
const PHONE = '07869787888'
const EMAIL = 'info@alamani.iq'

const n = (v: number) => Math.round(v).toLocaleString('en-US')
const abs = (u: string) => (/^(https?:|data:|blob:)/.test(u) ? u : new URL(u, location.origin).href)

/** المواصفات: كل سطر «عنوان: قيمة» (أو سطر عادي) */
export function parseSpecs(specs?: string | null): { label: string; value: string }[] {
  return (specs ?? '')
    .split(/\r?\n/)
    .map((l) => l.trim())
    .filter(Boolean)
    .map((l) => {
      const i = l.search(/[:：]/)
      return i > 0 ? { label: l.slice(0, i).trim(), value: l.slice(i + 1).trim() } : { label: '', value: l }
    })
}

/** أيقونة توضيحية حسب الفئة لما ماكو صورة للمنتج */
function illustration(category: string): string {
  const c = category
  const screen = '<rect x="20" y="22" width="120" height="74" rx="6" fill="#1e3a8a"/><rect x="26" y="28" width="108" height="62" rx="3" fill="url(#g)"/>'
  const defs = '<defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="#4f6ed8"/><stop offset="1" stop-color="#f59e0b"/></linearGradient></defs>'
  let body = screen + '<rect x="72" y="96" width="16" height="12" fill="#94a3b8"/><rect x="56" y="108" width="48" height="5" rx="2" fill="#64748b"/>'
  if (/طباع|استنساخ|print/i.test(c)) body = '<rect x="30" y="40" width="100" height="58" rx="8" fill="#e2e8f0" stroke="#94a3b8"/><rect x="45" y="22" width="70" height="22" rx="3" fill="#f8fafc" stroke="#94a3b8"/><rect x="45" y="84" width="70" height="26" rx="2" fill="#fff" stroke="#94a3b8"/><rect x="40" y="52" width="30" height="6" rx="2" fill="#1e3a8a"/>'
  else if (/كامير|camera/i.test(c)) body = '<rect x="30" y="45" width="80" height="40" rx="10" fill="#1e3a8a"/><circle cx="95" cy="65" r="14" fill="#0f172a" stroke="#4f6ed8" stroke-width="4"/><rect x="60" y="85" width="10" height="25" fill="#64748b"/>'
  else if (/شبك|network/i.test(c)) body = '<rect x="25" y="50" width="110" height="30" rx="5" fill="#1e3a8a"/>' + [0, 1, 2, 3, 4, 5].map((i) => `<rect x="${35 + i * 16}" y="60" width="10" height="8" fill="#93c5fd"/>`).join('')
  else if (/لابتوب|محمول|laptop/i.test(c)) body = screen + '<path d="M10 100 L150 100 L140 112 L20 112 Z" fill="#64748b"/>'
  return `<svg viewBox="0 0 160 125" class="illus">${defs}${body}</svg>`
}

const header = (title: string) => `
  <div class="hdr">
    <div class="hdr-co"><div><b>${COMPANY_AR}</b><b>${COMPANY_AR2}</b></div><img src="LOGO" alt=""></div>
    <div class="hdr-t"><small>AL-AMANI / PROJECT CATALOG</small><h2>${esc(title)}</h2></div>
  </div>`
const footer = (page: number, total: number) => `
  <div class="ftr"><span><b>شركة الأماني</b> · ${PHONE}</span><span>الصور توضيحية، والاعتماد على الأوصاف الفنية المدونة.</span><span class="pg">${String(page).padStart(2, '0')} / ${String(total).padStart(2, '0')}</span></div>`

function card(it: ProjectItem, no: number): string {
  const specs = parseSpecs(it.specs)
  const img = it.imageBase64 ? `<img src="${esc(abs(fileUrl(it.imageBase64)))}" alt="">` : illustration(it.category ?? it.productName)
  return `
  <div class="card">
    <div class="card-body">
      <div class="card-top"><span class="eq">EQUIPMENT / ${String(no).padStart(2, '0')}</span><span class="no">${String(no).padStart(2, '0')}</span></div>
      <h3>${esc(it.productName)}</h3>
      ${it.modelName ? `<div class="model" dir="ltr">${esc(it.modelName)}</div>` : ''}
      <ul class="specs">${specs.map((s) => `<li>${s.label ? `<b>${esc(s.label)}:</b> ` : ''}<bdi>${esc(s.value)}</bdi></li>`).join('')}</ul>
      <div class="money">
        <div><small>الكمية</small><b>${n(it.quantity)}</b></div>
        <div><small>السعر المفرد د.ع</small><b>${n(it.unitPrice)}</b></div>
        <div><small>الإجمالي د.ع</small><b class="tot">${n(it.quantity * it.unitPrice)}</b></div>
      </div>
    </div>
    <div class="card-img"><div class="circle">${img}</div><small>${it.imageBase64 ? '' : 'صورة توضيحية للمنتج'}</small></div>
  </div>`
}

export function buildProjectQuoteHtml(q: ProjectQuote, logoUrl: string): string {
  const items = q.items.filter((i) => i.productName.trim())
  const total = items.reduce((a, i) => a + i.quantity * i.unitPrice, 0)
  const units = items.reduce((a, i) => a + i.quantity, 0)
  // الفصول بترتيب أول ظهور للفئة
  const chapters: { name: string; items: { it: ProjectItem; no: number }[] }[] = []
  items.forEach((it, i) => {
    const name = it.category?.trim() || 'التجهيزات'
    let ch = chapters.find((c) => c.name === name)
    if (!ch) chapters.push((ch = { name, items: [] }))
    ch.items.push({ it, no: i + 1 })
  })
  // صفحات الفصول: كارتين بالصفحة
  const chapterPages: { name: string; idx: number; cards: { it: ProjectItem; no: number }[] }[] = []
  chapters.forEach((ch, ci) => {
    for (let i = 0; i < ch.items.length; i += 2) chapterPages.push({ name: ch.name, idx: ci + 1, cards: ch.items.slice(i, i + 2) })
  })
  const annexParts = q.annex.filter((p) => p.items.length)
  const totalPages = 2 + chapterPages.length + 1 + annexParts.length
  let page = 1
  const date = q.date.toLocaleDateString('en-US', { month: 'long', year: 'numeric' }).toUpperCase()
  const city = (q.cover.city || 'KARBALA, IRAQ').toUpperCase()
  const collage = items.slice(0, 4).map((it) => `<div class="col-item">${it.imageBase64 ? `<img src="${esc(abs(fileUrl(it.imageBase64)))}" alt="">` : illustration(it.category ?? it.productName)}</div>`).join('')

  const coverPage = `
  <section class="page cover">
    <div class="cover-top">
      <div class="hdr-co light"><div><b>${COMPANY_AR}</b><b>${COMPANY_AR2}</b></div><img src="LOGO" alt=""></div>
      <p class="cover-label">${esc(q.cover.label || q.customerName)}</p>
      <h1>${esc(q.cover.title || 'عرض تجهيز فني وتجاري')}</h1>
      <div class="bar"></div>
      ${q.cover.subtitle ? `<p class="cover-sub">${esc(q.cover.subtitle)}</p>` : ''}
    </div>
    <div class="collage">${collage}</div>
    <div class="cover-bottom">
      <div><h3>ملف التجهيز الفني والتجاري</h3><small>عرض مشاريع | ${items.length} بنداً وملحق فني متكامل${q.quotationNumber ? ` · رقم العرض ${esc(q.quotationNumber)}` : ''}</small></div>
      <div class="prep" dir="ltr"><small>PREPARED BY <b>AL-AMANI</b></small><small>${esc(date)} · ${esc(city)}</small></div>
    </div>
  </section>`
  page++

  const overview = `
  <section class="page">${header('نظرة شاملة على المشروع')}
    <div class="ov-hero">
      <small dir="ltr">INTEGRATED PROJECT TECHNOLOGY</small>
      <h2>${esc(q.cover.title || 'تجهيز تقني متكامل')}</h2>
      <p>مجموعة مختارة من التجهيزات، ضمن عرض منظم يوضح التجهيز والكميات والتكلفة والالتزامات الفنية.</p>
    </div>
    <div class="stats">
      <div><b>${n(items.length)}</b><small>بند تجهيز</small></div>
      <div><b>${n(units)}</b><small>وحدة ومعدة</small></div>
      <div><b>${(total / 1e6).toLocaleString('en-US', { maximumFractionDigits: 3 })}</b><small>مليون دينار عراقي</small></div>
    </div>
    <h3 class="sec">محاور المشروع <small dir="ltr">/ PROJECT SCOPE</small></h3>
    <div class="scope">${chapters.map((c, i) => `<div><small dir="ltr">${String(i + 1).padStart(2, '0')}</small><h4>${esc(c.name)}</h4><p>${c.items.map((x) => esc(x.it.productName)).slice(0, 4).join('، ')}${c.items.length > 4 ? '…' : ''}</p></div>`).join('')}</div>
    <p class="note">التجهيز والأسعار أدناه يستندان إلى قائمة المنتجات المقدمة، ويُستكملان بالشروط الفنية والتنفيذية في نهاية الملف.</p>
    ${footer(page, totalPages)}
  </section>`
  page++

  const chapterHtml = chapterPages.map((cp) => {
    const s = `
  <section class="page">${header(cp.name)}
    <div class="ch-head"><span class="big">${String(cp.idx).padStart(2, '0')}</span><div><small dir="ltr">CHAPTER ${String(cp.idx).padStart(2, '0')}</small><p>المنتجات والمواصفات الفنية والأسعار - بالدينار العراقي</p></div></div>
    ${cp.cards.map((c) => card(c.it, c.no)).join('')}
    ${footer(page, totalPages)}
  </section>`
    page++
    return s
  }).join('')

  const summary = `
  <section class="page">${header('العرض المالي والتجهيز')}
    <h3 class="sec">كشف الكميات والأسعار <small dir="ltr">IQD · ${items.length} ITEMS</small></h3>
    <table class="tbl">
      <thead><tr><th>ت</th><th>اسم المنتج</th><th>العدد</th><th>سعر المفرد</th><th>الإجمالي</th></tr></thead>
      <tbody>${items.map((it, i) => `<tr><td>${String(i + 1).padStart(2, '0')}</td><td class="nm">${esc(it.productName)}</td><td>${n(it.quantity)}</td><td>${n(it.unitPrice)}</td><td><b>${n(it.quantity * it.unitPrice)}</b></td></tr>`).join('')}</tbody>
    </table>
    <div class="grand"><div><h3>القيمة الإجمالية للمشروع</h3><p>${esc(arabicAmountWords(total))}</p></div><div class="gv"><b>${n(total)}</b><small>IQD</small></div></div>
    ${footer(page, totalPages)}
  </section>`
  page++

  let clause = 0
  const annexHtml = annexParts.map((p, pi) => {
    const last = pi === annexParts.length - 1
    const s = `
  <section class="page">${header('الملحق الفني ونطاق العمل')}
    <div class="ax-head"><small dir="ltr">IMPLEMENTATION / TECHNICAL ANNEX · PART ${String(pi + 1).padStart(2, '0')} OF ${String(annexParts.length).padStart(2, '0')}</small><h2>${esc(p.title)}</h2></div>
    ${p.items.map((it) => { clause++; return `<div class="clause"><span class="cn">${String(clause).padStart(2, '0')}</span><div><h4>${esc(it.title)}</h4><p>${esc(it.text)}</p></div></div>` }).join('')}
    <p class="note">هذه البنود مرتبطة بعرض التجهيز وتحدد إجراءات العمل وحقوق الطرفين.</p>
    ${last ? `<div class="closing"><div><h3>من التخطيط إلى التسليم…<br>حلول تقنية تستحق الثقة.</h3><small>شركة الأماني · هاتف ${PHONE} · ${EMAIL}</small></div><img src="LOGO" alt=""></div>` : ''}
    ${footer(page, totalPages)}
  </section>`
    page++
    return s
  }).join('')

  const css = fontsCss.replace(/url\((['"]?)([^)'"]+)\1\)/g, (_m, _q, u: string) => `url(${abs(u)})`)
  return `<!doctype html><html lang="ar" dir="rtl"><head><meta charset="utf-8"><title>${esc(q.cover.title || 'عرض مشاريع')}</title>
<style>${css}
@page { size: A4; margin: 0 }
* { box-sizing: border-box }
body { margin: 0; font-family: 'Tajawal', sans-serif; color: #1e293b; -webkit-print-color-adjust: exact; print-color-adjust: exact; background: #e2e8f0 }
.page { width: 210mm; height: 297mm; padding: 16mm 14mm 18mm; position: relative; overflow: hidden; background: #fff; page-break-after: always; margin: 0 auto }
.page::after { content: ''; position: absolute; bottom: 0; left: 0; right: 0; height: 9px; background: linear-gradient(90deg, #f59e0b 0 13%, #1e3a8a 13% 100%) }
.hdr { display: flex; justify-content: space-between; align-items: center; border-bottom: 1px solid #e2e8f0; padding-bottom: 12px; margin-bottom: 18px; position: relative }
.hdr::after { content: ''; position: absolute; bottom: -2px; right: 0; width: 26%; height: 3px; background: #f59e0b }
.hdr-co { display: flex; align-items: center; gap: 10px } .hdr-co div { display: flex; flex-direction: column; font-size: 11px; color: #1e3a8a; text-align: right }
.hdr-co img { width: 44px; height: 44px; object-fit: contain }
.hdr-co.light div { color: #fff }
.hdr-t { text-align: left } .hdr-t small { color: #94a3b8; letter-spacing: 2px; font-size: 9px } .hdr-t h2 { margin: 0; color: #1e3a8a; font-size: 22px }
.cover { padding: 0; background: #e5e9f0 }
.cover-top { background: linear-gradient(160deg, #1e3a8a, #2a4fb0); color: #fff; padding: 18mm 14mm 34mm; min-height: 150mm }
.cover-top .hdr-co { justify-content: flex-end }
.cover-label { margin: 22mm 0 6px; font-size: 13px; opacity: .85 }
.cover h1 { font-size: 40px; line-height: 1.35; margin: 0 }
.bar { width: 48mm; height: 4px; background: #f59e0b; border-radius: 4px; margin: 16px 0 }
.cover-sub { font-size: 14px; opacity: .9; margin: 0 }
.collage { margin: -24mm 14mm 0; background: #fff; border-radius: 18px; height: 78mm; display: flex; align-items: center; justify-content: space-around; padding: 10mm; box-shadow: 0 10px 30px rgba(15,23,42,.12) }
.col-item { width: 40mm; height: 40mm; display: flex; align-items: center; justify-content: center } .col-item img, .col-item svg { max-width: 100%; max-height: 100%; object-fit: contain }
.cover-bottom { display: flex; justify-content: space-between; align-items: flex-end; padding: 16mm 14mm 0 } .cover-bottom h3 { margin: 0; color: #1e3a8a }
.cover-bottom small { color: #64748b; display: block } .prep { text-align: left } .prep b { color: #f59e0b }
.ov-hero { background: linear-gradient(160deg, #1e3a8a, #2a4fb0); color: #fff; border-radius: 16px; padding: 12mm 12mm 22mm } .ov-hero small { opacity: .7; letter-spacing: 2px; font-size: 9px }
.ov-hero h2 { margin: 6px 0; font-size: 24px } .ov-hero p { opacity: .9; font-size: 12px; margin: 0 }
.stats { display: flex; gap: 10px; margin: -14mm 8mm 10mm } .stats div { flex: 1; background: #fff; border-radius: 12px; text-align: center; padding: 10px; box-shadow: 0 4px 14px rgba(15,23,42,.1) }
.stats b { display: block; font-size: 24px; color: #1e3a8a } .stats small { color: #64748b }
.sec { color: #1e3a8a; margin: 4mm 0 } .sec small { color: #94a3b8; font-size: 10px }
.scope { display: grid; grid-template-columns: 1fr 1fr; gap: 10px } .scope div { border: 1px solid #e2e8f0; border-radius: 12px; padding: 10px 12px; background: #f8fafc; min-height: 30mm }
.scope small { color: #f59e0b; font-weight: 700 } .scope h4 { margin: 4px 0; color: #1e3a8a } .scope p { margin: 0; font-size: 11px; color: #475569 }
.note { border-right: 4px solid #f59e0b; padding: 4px 10px; font-size: 10px; color: #64748b; margin-top: 8mm }
.ch-head { display: flex; align-items: center; gap: 12px; margin-bottom: 8px } .ch-head .big { font-size: 64px; font-weight: 800; color: #e2e8f0; line-height: 1 }
.ch-head small { color: #94a3b8; letter-spacing: 2px; font-size: 9px } .ch-head p { margin: 0; font-size: 11px; color: #64748b }
.card { display: flex; border: 1px solid #e2e8f0; border-radius: 14px; overflow: hidden; height: 104mm; margin-bottom: 6mm; border-left: 4px solid #1e3a8a }
.card-body { flex: 1.6; padding: 10px 14px; display: flex; flex-direction: column }
.card-top { display: flex; justify-content: space-between; align-items: center } .eq { background: #eef2ff; color: #1e3a8a; font-size: 8px; padding: 3px 8px; border-radius: 10px; letter-spacing: 1px } .no { color: #f59e0b; font-weight: 800; font-size: 20px }
.card h3 { margin: 4px 0 0; color: #1e3a8a; font-size: 17px } .model { color: #4f6ed8; font-size: 10px; font-weight: 700; text-align: right }
.specs { list-style: none; padding: 0; margin: 6px 0; flex: 1; overflow: hidden } .specs li { font-size: 10.5px; padding: 3px 0; border-bottom: 1px solid #f1f5f9 } .specs li::before { content: '•'; color: #f59e0b; margin-left: 6px } .specs b { color: #1e3a8a }
.money { display: flex; justify-content: space-between; border-top: 1px solid #e2e8f0; padding-top: 6px } .money div { text-align: center } .money small { display: block; font-size: 9px; color: #94a3b8 } .money b { font-size: 15px; color: #1e293b } .money .tot { color: #f59e0b }
.card-img { flex: 1; background: linear-gradient(180deg, #eef2ff, #f8fafc); display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 6px }
.circle { width: 46mm; height: 46mm; border-radius: 50%; border: 1px solid #dbe3f4; display: flex; align-items: center; justify-content: center; padding: 6mm } .circle img, .circle svg { max-width: 100%; max-height: 100%; object-fit: contain } .card-img small { font-size: 9px; color: #94a3b8 }
.illus { width: 100%; height: auto }
.tbl { width: 100%; border-collapse: collapse; font-size: 11px } .tbl th { background: #1e3a8a; color: #fff; padding: 7px; font-weight: 600 } .tbl td { padding: 6px 7px; border-bottom: 1px solid #e2e8f0; text-align: center } .tbl td.nm { text-align: right; font-weight: 600 } .tbl tr:nth-child(even) td { background: #f8fafc }
.grand { margin-top: 8mm; background: linear-gradient(160deg, #1e3a8a, #2a4fb0); color: #fff; border-radius: 14px; padding: 8mm; display: flex; justify-content: space-between; align-items: center } .grand h3 { margin: 0 } .grand p { margin: 4px 0 0; font-size: 11px; opacity: .9 }
.gv { text-align: left } .gv b { display: block; font-size: 30px } .gv small { color: #f59e0b }
.ax-head small { color: #94a3b8; letter-spacing: 2px; font-size: 9px } .ax-head h2 { margin: 2px 0 8px; color: #1e3a8a }
.clause { display: flex; gap: 10px; border: 1px solid #e2e8f0; border-radius: 12px; padding: 8px 12px; margin-bottom: 4mm; background: #fbfcff } .cn { color: #f59e0b; font-weight: 800 } .clause h4 { margin: 0 0 3px; color: #1e3a8a; font-size: 13px } .clause p { margin: 0; font-size: 10.2px; line-height: 1.75; color: #334155; text-align: justify }
.closing { margin-top: 6mm; background: linear-gradient(160deg, #1e3a8a, #2a4fb0); color: #fff; border-radius: 14px; padding: 8mm; display: flex; justify-content: space-between; align-items: center } .closing h3 { margin: 0 0 6px; line-height: 1.6 } .closing small { opacity: .8 } .closing img { width: 26mm; filter: brightness(0) invert(1) }
.ftr { position: absolute; bottom: 12mm; left: 14mm; right: 14mm; display: flex; justify-content: space-between; font-size: 9px; color: #94a3b8; border-top: 1px solid #e2e8f0; padding-top: 6px } .ftr b { color: #1e3a8a } .pg { color: #f59e0b; font-weight: 700 }
@media print { body { background: #fff } .page { margin: 0 } }
</style></head><body>${coverPage}${overview}${chapterHtml}${summary}${annexHtml}</body></html>`.replace(/src="LOGO"/g, `src="${esc(abs(logoUrl))}"`)
}

/** يطبع العرض بإطار مخفي (نفس طريقة عرض السعر القديم) بعد ما تخلص الصور والخطوط. */
export function printProjectQuote(html: string) {
  const frame = document.createElement('iframe')
  frame.style.cssText = 'position:fixed;right:0;bottom:0;width:0;height:0;border:0;visibility:hidden'
  document.body.appendChild(frame)
  const w = frame.contentWindow
  if (!w) { frame.remove(); return }
  w.document.open(); w.document.write(html); w.document.close()
  w.onafterprint = () => setTimeout(() => frame.remove(), 500)
  const imgs = Array.from(w.document.images).map((img) => (img.complete ? Promise.resolve() : new Promise<void>((r) => { img.onload = () => r(); img.onerror = () => r() })))
  void Promise.race([Promise.all([...imgs, w.document.fonts?.ready]), new Promise((r) => setTimeout(r, 6000))])
    .then(() => setTimeout(() => { w.focus(); w.print() }, 200))
}
