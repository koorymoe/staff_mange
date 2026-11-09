import { UR } from './ur'

// ═══ ترجمة الواجهة للأوردو (بلاغ #11، 11-07) ═══
// بدل ما نلف مئات النصوص بالكود: إذا لغة الموظف «ur»، نراقب الصفحة ونبدّل أي نص
// عربي موجود بالقاموس بترجمته (النص، والـplaceholder، والـtitle، والـaria-label).
// النص الي مو بالقاموس يبقى عربي. الأوردو من اليمين لليسار مثل العربي، فالتنسيق ما يتغيّر.

const ATTRS = ['placeholder', 'title', 'aria-label'] as const

export function uiLanguage(): 'ar' | 'ur' {
  try {
    const o = localStorage.getItem('uiLang')
    if (o === 'ur' || o === 'ar') return o
    const e = JSON.parse(localStorage.getItem('currentEmployee') || 'null') as { language?: string } | null
    return e?.language === 'ur' ? 'ur' : 'ar'
  } catch {
    return 'ar'
  }
}

/** يترجم نص واحد — يحافظ على الفراغات حول النص. */
export function trUR(s: string): string | null {
  const t = s.trim()
  if (!t) return null
  const u = UR[t] ?? patternUR(t)
  if (!u) return null
  return s.replace(t, u)
}

// نصوص فيها رقم أو وقت متغيّر — القاموس ما يطابقها حرفياً، فقواعد صغيرة (11-07)
const PATTERNS: [RegExp, (m: RegExpMatchArray) => string][] = [
  [/^(الصباحي|المسائي) من (.*) لـ (.*)$/, (m) => `${m[1] === 'الصباحي' ? 'صبح کی' : 'شام کی'} شفٹ ${m[2]} سے ${m[3]} تک`],
]
function patternUR(t: string): string | null {
  for (const [re, f] of PATTERNS) {
    const m = t.match(re)
    if (m) return f(m)
  }
  return null
}

function translateNode(root: Node) {
  if (root.nodeType === Node.TEXT_NODE) {
    const v = root.nodeValue ?? ''
    const r = trUR(v)
    if (r !== null && r !== v) root.nodeValue = r
    return
  }
  if (root.nodeType !== Node.ELEMENT_NODE) return
  const el = root as Element
  if (el.tagName === 'SCRIPT' || el.tagName === 'STYLE' || el.tagName === 'TEXTAREA') return
  for (const a of ATTRS) {
    const v = el.getAttribute(a)
    if (v) {
      const r = trUR(v)
      if (r !== null && r !== v) el.setAttribute(a, r)
    }
  }
  const walker = document.createTreeWalker(el, NodeFilter.SHOW_TEXT | NodeFilter.SHOW_ELEMENT)
  let n = walker.nextNode()
  while (n) {
    if (n.nodeType === Node.TEXT_NODE) {
      const p = n.parentElement
      if (!p || (p.tagName !== 'TEXTAREA' && p.tagName !== 'SCRIPT' && p.tagName !== 'STYLE')) {
        const v = n.nodeValue ?? ''
        const r = trUR(v)
        if (r !== null && r !== v) n.nodeValue = r
      }
    } else {
      const e = n as Element
      for (const a of ATTRS) {
        const v = e.getAttribute(a)
        if (v) {
          const r = trUR(v)
          if (r !== null && r !== v) e.setAttribute(a, r)
        }
      }
    }
    n = walker.nextNode()
  }
}

let started = false
export function startUrduIfNeeded() {
  if (started || uiLanguage() !== 'ur' || typeof MutationObserver === 'undefined') return
  started = true
  document.documentElement.lang = 'ur'
  document.documentElement.classList.add('lang-ur')
  translateNode(document.body)
  new MutationObserver((muts) => {
    for (const m of muts) {
      if (m.type === 'characterData') translateNode(m.target)
      else if (m.type === 'attributes') translateNode(m.target)
      else m.addedNodes.forEach((x) => translateNode(x))
    }
  }).observe(document.body, { subtree: true, childList: true, characterData: true, attributes: true, attributeFilter: [...ATTRS] })
}

/** تبديل اللغة من الإعدادات: يحفظ ويعيد تحميل الصفحة (أبسط وأضمن من عكس الترجمة). */
export function setUiLanguage(lang: 'ar' | 'ur') {
  try {
    localStorage.setItem('uiLang', lang)
    const e = JSON.parse(localStorage.getItem('currentEmployee') || 'null') as Record<string, unknown> | null
    if (e) localStorage.setItem('currentEmployee', JSON.stringify({ ...e, language: lang }))
  } catch {
    /* وضع خاص */
  }
  location.reload()
}
