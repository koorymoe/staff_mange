// ═══ «عرض مشاريع» — صفحة الطباعة (دازاي 11-06) ═══
// نفس تصميم نموذج «مشروع جامعة المكين»: غلاف، نظرة شاملة، فصل لكل فئة بكارتين بالصفحة،
// كشف الكميات والأسعار ويا المبلغ بالحروف، والملحق الفني بأجزائه. A4 وألوان الشركة.
// الأرقام كلها من العرض نفسه — ماكو شي مخترع.
import fontsCss from '../fonts.css?inline'
import { fileUrl } from '../api'
import { arabicAmountWords } from './arabicWords'
import { esc } from './printIdentity'

export type ProjectCover = { label?: string; title?: string; subtitle?: string; city?: string; image?: string }
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
// رأس الفصل مثل النموذج: رقم كبير شفاف ويا خط برتقالي، و«CHAPTER 0n / …»
const chHead = (idx: number, desc: string, en = 'EQUIPMENT', big?: number) => `
    <div class="ch-head"><div><small dir="ltr">CHAPTER ${String(idx).padStart(2, '0')} / ${en}</small><p>${desc}</p></div>${big ? `<div class="bigwrap"><span class="big">${String(big).padStart(2, '0')}</span><i></i></div>` : ''}</div>`
const footer = (page: number, total: number) => `
  <div class="ftr"><span><b>شركة الأماني</b> · <bdi dir="ltr">${PHONE}</bdi></span><span>الصور توضيحية، والاعتماد على الأوصاف الفنية المدونة.</span><span class="pg"><bdi dir="ltr">${String(page).padStart(2, '0')} / ${String(total).padStart(2, '0')}</bdi></span></div>`

function card(it: ProjectItem, no: number, otherChapter?: string): string {
  const specs = parseSpecs(it.specs)
  const img = it.imageBase64 ? `<img src="${esc(abs(fileUrl(it.imageBase64)))}" alt="">` : illustration(it.category ?? it.productName)
  // كثافة المواصفات: قليلة = أسطر أوسع؛ كثيرة = أضيق — حتى البطاقة تمتلي بلا فراغ ولا طفح
  const dens = specs.length <= 4 ? 'roomy' : specs.length >= 10 ? 'tight' : ''
  const nn = String(no).padStart(2, '0')
  // (بلاغ #17 — مؤمل 11-07) مثل النموذج: الكلام يمين والصورة يسار، والمواصفات نقاط «العنوان: القيمة»
  return `
  <div class="card ${dens}">
    <div class="card-body">
      <div class="card-top"><small class="eq-tag" dir="ltr">EQUIPMENT / ${nn}</small>${otherChapter ? `<span class="eq">${esc(otherChapter)}</span>` : ''}</div>
      <div class="ttl"><span class="no"><bdi dir="ltr">${nn}</bdi></span><h3>${esc(it.productName)}</h3></div>
      ${it.modelName ? `<bdi class="model" dir="ltr">${esc(it.modelName)}</bdi>` : ''}
      ${specs.length ? `<ul class="specs">${specs.map((s) => `<li>${s.label ? `<b>${esc(s.label)}:</b> ` : ''}<span dir="auto">${esc(s.value)}</span></li>`).join('')}</ul>` : '<div class="specs-empty">—</div>'}
      <div class="money">
        <div><small>الكمية</small><b><bdi dir="ltr">${n(it.quantity)}</bdi></b></div>
        <div><small>السعر المفرد</small><b><bdi dir="ltr">${n(it.unitPrice)}</bdi></b></div>
        <div class="tot"><small>الإجمالي (د.ع)</small><b><bdi dir="ltr">${n(it.quantity * it.unitPrice)}</bdi></b></div>
      </div>
    </div>
    <div class="card-img">
      <div class="pic">${img}</div>
      <span class="dia"></span>
      ${it.imageBase64 ? '' : '<small>صورة توضيحية للمنتج</small>'}
    </div>
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
  // صفحات المنتجات (مثل نموذج دازاي): كارتين بكل صفحة بالتسلسل — حتى لو الفئتين مختلفات، ماكو نص صفحة فاضي.
  // وإذا العدد فردي، آخر منتج ينعرض فوق جدول الكميات بصفحة العرض المالي.
  const flat = chapters.flatMap((ch, ci) => ch.items.map((x) => ({ ...x, ch: ch.name, idx: ci + 1 })))
  const lastAlone = flat.length % 2 === 1 ? flat[flat.length - 1] : null
  const paged = lastAlone ? flat.slice(0, -1) : flat
  const chapterPages: { name: string; idx: number; cards: { it: ProjectItem; no: number; ch: string; idx: number }[] }[] = []
  for (let i = 0; i < paged.length; i += 2) {
    const pair = paged.slice(i, i + 2)
    chapterPages.push({ name: pair[0].ch, idx: pair[0].idx, cards: pair })
  }
  const annexParts = q.annex.filter((p) => p.items.length)
  const totalPages = 2 + chapterPages.length + 1 + annexParts.length
  let page = 1
  const date = q.date.toLocaleDateString('en-US', { month: 'long', year: 'numeric' }).toUpperCase()
  const city = (q.cover.city || 'KARBALA, IRAQ').toUpperCase()
  // (بلاغ #13) صورة الواجهة: إذا الموظف رفع صورة وحدة تنعرض هي؛ وإلا ٣ منتجات (المصوّرة أول) على «مسرح» والوسطى أكبر
  const withImg = [...items.filter((i) => i.imageBase64), ...items.filter((i) => !i.imageBase64)].slice(0, 3)
  const staged = withImg.length === 3 ? [withImg[1], withImg[0], withImg[2]] : withImg
  const collage = q.cover.image
    ? `<img class="hero-img" src="${esc(abs(fileUrl(q.cover.image)))}" alt="">`
    : staged.map((it, i) => `<div class="col-item${staged.length === 3 && i === 1 ? ' main' : ''}">${it.imageBase64 ? `<img src="${esc(abs(fileUrl(it.imageBase64)))}" alt="">` : illustration(it.category ?? it.productName)}</div>`).join('')

  const coverPage = `
  <section class="page cover">
    <svg class="arc" viewBox="0 0 400 400"><circle cx="300" cy="330" r="190" fill="none" stroke="rgba(255,255,255,.07)" stroke-width="46"/></svg>
    <div class="cover-top">
      <div class="hdr-co light"><div><b>${COMPANY_AR}</b><b>${COMPANY_AR2}</b></div><img src="LOGO" alt=""></div>
      <p class="cover-label">${esc(q.cover.label || q.customerName)}</p>
      <h1>${esc(q.cover.title || 'عرض تجهيز فني وتجاري')}</h1>
      <div class="bar"></div>
      ${q.cover.subtitle ? `<p class="cover-sub">${esc(q.cover.subtitle)}</p>` : ''}
    </div>
    <div class="collage${q.cover.image ? ' has-hero' : ''}"><span class="glow"></span>${collage}</div>
    <div class="cover-line"></div>
    <div class="cover-bottom">
      <div><h3>ملف التجهيز الفني والتجاري</h3><small>عرض مشاريع · ${items.length} بنداً وملحق فني متكامل</small>${q.quotationNumber ? `<small>رقم العرض <bdi dir="ltr">${esc(q.quotationNumber)}</bdi></small>` : ''}</div>
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
    <div class="scope">${chapters.map((c, i) => `<div><small dir="ltr">${String(i + 1).padStart(2, '0')} / SCOPE</small><h4>${esc(c.name)}</h4><p>${c.items.map((x) => esc(x.it.productName)).slice(0, 3).join('، ')}${c.items.length > 3 ? ` و${c.items.length - 3} غيرها` : ''}</p></div>`).join('')}</div>
    <p class="note">التجهيز والأسعار أدناه يستندان إلى قائمة المنتجات المقدمة، ويُستكملان بالشروط الفنية والتنفيذية في نهاية الملف.</p>
    ${footer(page, totalPages)}
  </section>`
  page++

  const chapterHtml = chapterPages.map((cp, pi) => {
    const s = `
  <section class="page">${header(cp.name)}
    ${chHead(cp.idx, 'المنتجات والمواصفات الفنية والأسعار - بالدينار العراقي', 'EQUIPMENT', pi + 1)}
    ${cp.cards.map((c) => card(c.it, c.no, c.ch !== cp.name ? c.ch : undefined)).join('')}
    ${footer(page, totalPages)}
  </section>`
    page++
    return s
  }).join('')

  const summary = `
  <section class="page fin${lastAlone ? ' with-card' : ''}">${header('العرض المالي والتجهيز')}
    ${lastAlone ? `${chHead(chapters.length + 1, 'البند الأخير والملخص المالي التفصيلي لأسعار المشروع', 'FINANCIAL SUMMARY')}<div class="fin-card">${card(lastAlone.it, lastAlone.no, lastAlone.ch)}</div>` : ''}
    <h3 class="sec">كشف الكميات والأسعار <small dir="ltr">IQD · ${items.length} ITEMS</small></h3>
    <table class="tbl">
      <thead><tr><th class="c">ت</th><th>اسم المنتج</th><th class="c">العدد</th><th class="num">سعر المفرد</th><th class="num">الإجمالي</th></tr></thead>
      <tbody>${chapters.map((ch) => `<tr class="grp"><td colspan="5">${esc(ch.name)}</td></tr>` + ch.items.map(({ it, no }) => `<tr><td class="c"><bdi dir="ltr">${String(no).padStart(2, '0')}</bdi></td><td class="nm">${esc(it.productName)}</td><td class="c"><bdi dir="ltr">${n(it.quantity)}</bdi></td><td class="num"><bdi dir="ltr">${n(it.unitPrice)}</bdi></td><td class="num"><b><bdi dir="ltr">${n(it.quantity * it.unitPrice)}</bdi></b></td></tr>`).join('')).join('')}</tbody>
      <tfoot><tr><td colspan="2">المجموع</td><td class="c"><bdi dir="ltr">${n(units)}</bdi></td><td></td><td class="num"><bdi dir="ltr">${n(total)}</bdi></td></tr></tfoot>
    </table>
    <div class="grand"><div><h3>القيمة الإجمالية للمشروع</h3><p>${esc(arabicAmountWords(total))}</p></div><div class="gv"><b><bdi dir="ltr">${n(total)}</bdi></b><small>IQD</small></div></div>
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

.hdr-co.light img { filter: brightness(0) invert(1) }
.cover { position: relative } .arc { position: absolute; top: 0; left: -60mm; width: 190mm; height: 190mm; pointer-events: none }
.cover-top { position: relative; padding: 16mm 14mm 40mm; min-height: 160mm }
.cover h1 { font-size: 42px; line-height: 1.4; max-width: 150mm }
.collage { position: relative; height: 92mm; margin-top: -40mm }
.glow { position: absolute; width: 70mm; height: 70mm; border-radius: 50%; background: radial-gradient(circle, rgba(245,158,11,.22), transparent 70%); right: 40%; top: 6mm }
.col-item { position: relative; width: 44mm; height: 52mm } .col-item img { filter: drop-shadow(0 8px 10px rgba(15,23,42,.18)) }
.cover-line { margin: 10mm 14mm 0; border-top: 1px solid #cbd5e1 }
.cover-bottom { padding-top: 6mm }
.ov-hero { padding: 14mm 12mm 26mm } .ov-hero h2 { font-size: 28px; line-height: 1.4 }
.scope div { border-top: 3px solid #f59e0b; min-height: 36mm }
.ch-head .bigwrap { position: relative } .ch-head .big { font-size: 78px; color: #e8edf5 } .ch-head .bigwrap i { position: absolute; right: 0; bottom: 6px; width: 44px; height: 3px; background: #f59e0b }
.card { height: 108mm }
.card-img { position: relative } .dia { position: absolute; bottom: 14mm; right: 12mm; width: 10mm; height: 10mm; border: 1.5px solid #f6c77a; transform: rotate(45deg); opacity: .8 }

.ch-head { justify-content: space-between } .ch-head .bigwrap i { right: auto; left: 0 }
.arc { left: -90mm; top: 10mm; width: 230mm; height: 230mm } .arc circle { stroke: rgba(255,255,255,.09) }
.fin.with-card .fin-card .card { height: 74mm; margin-bottom: 4mm } .fin.with-card .fin-card .specs li { padding: 1px 0; font-size: 9.5px } .fin.with-card .fin-card .circle { width: 36mm; height: 36mm }
.fin.with-card .tbl { font-size: 9.5px } .fin.with-card .tbl th { padding: 4px } .fin.with-card .tbl td { padding: 3px 6px }
.fin.with-card .grand { margin-top: 4mm; padding: 5mm 7mm } .fin.with-card .gv b { font-size: 24px } .fin .sec { margin: 2mm 0 }
.arc { z-index: 1 } .cover-top > * { position: relative; z-index: 2 } .collage { z-index: 2 }

/* ═══ تنسيق v2 (دازاي 11-06: أقوى من النموذج) ═══ */
.cover h1 { font-size: 44px; text-wrap: balance; max-width: 160mm }
.arc { left: auto; right: 90mm; top: 40mm; width: 180mm; height: 180mm } .arc circle { stroke: rgba(255,255,255,.12) }
.cover-bottom h3 { font-size: 18px; margin-bottom: 4px } .cover-bottom small { font-size: 10.5px; line-height: 1.7 }
.card { display: grid; grid-template-columns: 38% 62%; border: 1px solid #dbe3f0; border-radius: 14px; overflow: hidden; height: 112mm; margin-bottom: 5mm; border-left: 0; box-shadow: 0 2px 10px rgba(15,23,42,.05) }
.card-img { background: linear-gradient(180deg, #eef2fb, #f8fafc); display: flex; flex-direction: column; align-items: center; justify-content: center; gap: 8px; padding: 8mm 6mm; border-left: 1px solid #e2e8f0 }
.pic { width: 100%; height: 62mm; display: flex; align-items: center; justify-content: center } .pic img, .pic svg { max-width: 100%; max-height: 100%; object-fit: contain; filter: drop-shadow(0 10px 12px rgba(15,23,42,.15)) }
.model-chip { background: #1e3a8a; color: #fff; font-size: 9px; font-weight: 700; padding: 3px 10px; border-radius: 20px; letter-spacing: .5px; max-width: 100%; overflow: hidden; text-overflow: ellipsis; white-space: nowrap }
.card-img small { font-size: 8.5px; color: #94a3b8 }
.card-body { padding: 5mm 6mm 0; display: flex; flex-direction: column; min-width: 0 }
.card-top { display: flex; align-items: center; gap: 8px; margin-bottom: 2px } .no { background: #f59e0b; color: #fff; font-weight: 800; font-size: 13px; border-radius: 8px; padding: 2px 9px } .eq { background: #eef2ff; color: #1e3a8a; font-size: 9px; padding: 3px 8px; border-radius: 10px }
.card h3 { margin: 4px 0 8px; color: #0f2a6b; font-size: 16px; line-height: 1.45 }
table.specs { width: 100%; border-collapse: collapse; margin: 0 0 auto; table-layout: fixed } table.specs th, table.specs td { font-size: 10.5px; padding: 4px 6px; border-bottom: 1px solid #eef2f7; vertical-align: top; text-align: right; word-wrap: break-word }
table.specs th { width: 34%; color: #1e3a8a; font-weight: 700 } table.specs td { color: #334155 } table.specs tr:nth-child(odd) td, table.specs tr:nth-child(odd) th { background: #f8fafc }
.card.roomy table.specs th, .card.roomy table.specs td { padding: 11px 8px; font-size: 12.5px } .card.tight table.specs th, .card.tight table.specs td { padding: 2px 6px; font-size: 9.5px }
.specs-empty { margin: auto 0; color: #cbd5e1 }
.money { display: grid; grid-template-columns: 1fr 1fr 1.3fr; margin: 4mm -6mm 0; border-top: 1px solid #e2e8f0 } .money div { padding: 3mm 4mm; text-align: center; border-left: 1px solid #eef2f7 } .money div:last-child { border-left: 0 }
.money small { display: block; font-size: 8.5px; color: #64748b; margin-bottom: 2px } .money b { font-size: 15px; color: #0f172a; font-variant-numeric: tabular-nums }
.money .tot { background: #1e3a8a } .money .tot small { color: #c7d2fe } .money .tot b { color: #fff; font-size: 16px }
.fin.with-card .fin-card .card { height: 72mm } .fin.with-card .fin-card .pic { height: 38mm } .fin.with-card .fin-card table.specs th, .fin.with-card .fin-card table.specs td { padding: 1px 6px; font-size: 9px }
.page.fin { display: flex; flex-direction: column } .page.fin .grand { margin-top: auto; margin-bottom: 14mm }
.tbl { font-variant-numeric: tabular-nums } .tbl th.num, .tbl td.num { text-align: left; padding-left: 10px } .tbl th.c, .tbl td.c { text-align: center; width: 9% }
.tbl tr.grp td { background: #eef2fb !important; color: #1e3a8a; font-weight: 700; text-align: right; font-size: 10px; padding: 4px 8px }
.tbl tfoot td { background: #0f2a6b; color: #fff; font-weight: 800; padding: 6px 8px; text-align: right } .tbl tfoot td.num { text-align: left } .tbl tfoot td.c { text-align: center }
.ftr .pg bdi, .ftr bdi { unicode-bidi: isolate }
.card { height: 101mm } table.specs { flex: 0 0 auto; margin: 0 } .money { margin-top: auto }
.card-body { overflow: hidden }

/* ═══ v3 — ملاحظات مؤمل 11-07 (بلاغات #12–#18) ═══ */
/* #12 #14: الشعار واسم الشركة يمين الغلاف وأوضح */
.cover-top .hdr-co { justify-content: flex-start } .cover-top .hdr-co img { width: 58px; height: 58px } .cover-top .hdr-co div { font-size: 13px; line-height: 1.5 }
.cover-label { margin-top: 26mm } .cover h1 { font-size: 46px; line-height: 1.35 }
/* #13: مسرح الصور بالواجهة — الوسطى أكبر، والكل على نفس الأرضية؛ أو صورة وحدة مرفوعة */
.collage { align-items: flex-end; justify-content: center; gap: 6mm; padding: 10mm 12mm 12mm }
.col-item { width: 46mm; height: 48mm; align-items: flex-end } .col-item.main { width: 62mm; height: 64mm }
.collage.has-hero { padding: 0; overflow: hidden } .hero-img { width: 100%; height: 100%; object-fit: cover; border-radius: 18px }
/* #15: النص الطويل سطرين مو سطر واحد ناعم */
.ov-hero h2 { max-width: 150mm; text-wrap: balance } .ov-hero p { max-width: 128mm; font-size: 13px; line-height: 1.8; text-wrap: balance }
.stats b { font-size: 30px } .stats div { padding: 14px 10px }
/* #16: محاور المشروع — وسم صغير، عنوان واضح، ووصف سطرين */
.scope div { padding: 12px 14px } .scope small { font-size: 10px; letter-spacing: 1.5px } .scope h4 { font-size: 16px; margin: 6px 0 4px }
.scope p { font-size: 11.5px; line-height: 1.7; display: -webkit-box; -webkit-line-clamp: 2; -webkit-box-orient: vertical; overflow: hidden }
/* #17: الكلام يمين والصورة يسار، والمواصفات نقاط */
.card { grid-template-columns: 62% 38% } .card-img { border-left: 0; border-right: 1px solid #e2e8f0; position: relative }
.card-top { justify-content: space-between } .eq-tag { color: #94a3b8; font-size: 8.5px; letter-spacing: 1.5px }
.ttl { display: flex; align-items: baseline; gap: 8px } .ttl .no { background: none; color: #f59e0b; font-size: 22px; padding: 0; font-weight: 800 } .ttl h3 { margin: 2px 0 2px; font-size: 17px }
.card .model { display: block; color: #4f6ed8; font-size: 10px; font-weight: 700; margin-bottom: 6px; text-align: right }
ul.specs { list-style: none; padding: 0; margin: 0; flex: 0 0 auto } ul.specs li { font-size: 10.5px; padding: 3px 0; border-bottom: 1px dashed #eef2f7; color: #334155; line-height: 1.6 }
ul.specs li::before { content: '•'; color: #f59e0b; margin-left: 6px; font-weight: 800 } ul.specs b { color: #1e3a8a }
.card.roomy ul.specs li { font-size: 12.5px; padding: 7px 0 } .card.tight ul.specs li { font-size: 9.5px; padding: 1px 0 }
.dia { bottom: 10mm; right: auto; left: 10mm }
/* #18: الأرقام بسطر بسيط مثل النموذج، والإجمالي برتقالي */
.money { margin: auto 0 0; border-top: 1px solid #e2e8f0; grid-template-columns: 1fr 1fr 1.3fr } .money div { border-left: 0; padding: 3mm 2mm }
.money .tot { background: none } .money .tot small { color: #64748b } .money .tot b { color: #f59e0b; font-size: 17px }
.fin.with-card .fin-card ul.specs li { padding: 0; font-size: 9px }
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
