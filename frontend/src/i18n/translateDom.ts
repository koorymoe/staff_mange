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
// المدد بالعربي («1 يوم 23 س»، «30 د») — للرسائل الي يكتبها السيرفر
const dur = (x: string) => x.replace(/(\d+)\s*يوم/g, '$1 دن').replace(/(\d+)\s*س(?![؀-ۿ])/g, '$1 گھنٹے').replace(/(\d+)\s*د(?![؀-ۿ])/g, '$1 منٹ')
const tw = (x: string) => UR[x.trim()] ?? x.trim()
// «عندك 3 حجوزات اليوم، 2 …» — بنود ملخّص الواجبات
const DUE: [RegExp, string][] = [
  [/^(\d+) تسجيل الحضور اليوم$/, '$1 آج کی حاضری'],
  [/^(\d+) حجوزات اليوم موادها ما تجهزت$/, '$1 آج کی بکنگ جن کا سامان تیار نہیں'],
  [/^(\d+) حجوزات ناقصها فاتورة أو تقرير$/, '$1 بکنگ جن کا بل یا رپورٹ باقی ہے'],
  [/^(\d+) حجوزات اليوم$/, '$1 آج کی بکنگ'],
]
// أيام الأسبوع والأشهر — للتواريخ الي تطلع من المتصفح بالعربي
const DAYS: Record<string, string> = { 'الأحد': 'اتوار', 'الاثنين': 'پیر', 'الثلاثاء': 'منگل', 'الأربعاء': 'بدھ', 'الخميس': 'جمعرات', 'الجمعة': 'جمعہ', 'السبت': 'ہفتہ' }
const MONTHS: Record<string, string> = {
  'كانون الثاني': 'جنوری', 'شباط': 'فروری', 'آذار': 'مارچ', 'نيسان': 'اپریل', 'أيار': 'مئی', 'حزيران': 'جون',
  'تموز': 'جولائی', 'آب': 'اگست', 'أيلول': 'ستمبر', 'تشرين الأول': 'اکتوبر', 'تشرين الثاني': 'نومبر', 'كانون الأول': 'دسمبر',
}
const MONTH_RE = '(كانون الثاني|شباط|آذار|نيسان|أيار|حزيران|تموز|آب|أيلول|تشرين الأول|تشرين الثاني|كانون الأول)'
const DAY_RE = '(الأحد|الاثنين|الثلاثاء|الأربعاء|الخميس|الجمعة|السبت)'
const PATTERNS: [RegExp, (m: RegExpMatchArray) => string][] = [
  [new RegExp(`^${DAY_RE}، (\\S+) ${MONTH_RE}، من (.+) إلى (.+)$`), (m) => `${DAYS[m[1]]}، ${m[2]} ${MONTHS[m[3]]}، ${m[4]} سے ${m[5]} تک`],
  [new RegExp(`^${DAY_RE}، (\\S+) ${MONTH_RE} (\\S+)$`), (m) => `${DAYS[m[1]]}، ${m[2]} ${MONTHS[m[3]]} ${m[4]}`],
  [new RegExp(`^${MONTH_RE} (\\S+)$`), (m) => `${MONTHS[m[1]]} ${m[2]}`],
  [/^من (.+) · (\d+) دقيقة$/, (m) => `${m[1]} سے · ${m[2]} منٹ`],
  [/^(الالتزام|السرعة|جودة التنفيذ) (\d+) من ٥$/, (m) => `${UR[m[1]]} ٥ میں سے ${m[2]}`],
  [/^من (\d+) مكلّف بيها$/, (m) => `${m[1]} میں سے جو آپ کو دی گئیں`],
  [/^من (\d+)$/, (m) => `${m[1]} میں سے`],
  [/^الإشعارات — (\d+) غير مقروء$/, (m) => `اطلاعات — ${m[1]} نہ پڑھی گئیں`],
  [/^انحسبلك (\d+) من (\d+) بشغلك$/, (m) => `آپ کے کام سے ${m[2]} میں سے ${m[1]} شمار ہوئے`],
  [/^⏳ بانتظار التقييم — باقي (\d+) من (\d+)$/, (m) => `⏳ تشخیص کا انتظار — ${m[2]} میں سے ${m[1]} باقی`],
  [/^- يحتاج (\d+) مهارة إضافية للترقي إلى "(.+)"$/, (m) => `- "${tw(m[2])}" تک ترقی کے لیے ${m[1]} مزید مہارتیں درکار`],
  [/^حجز (\S+) · (.+?): صارت بوقتها — صارت خلال (.+?) \(المعتاد ما يتجاوز (.+?)\)\.$/,
    (m) => `بکنگ ${m[1]} · ${tw(m[2])}: وقت پر ہوا — ${dur(m[3])} میں ہوا (عام طور پر ${dur(m[4])} سے زیادہ نہیں)۔`],
  [/^حجز (\S+) · (.+?): صارت بس بيها غلط — صارت\.$/, (m) => `بکنگ ${m[1]} · ${tw(m[2])}: ہوا مگر غلطی کے ساتھ — ہو گیا۔`],
  [/^حجز (\S+) · (.+?): صارت بس متأخرة — (.+)$/, (m) => `بکنگ ${m[1]} · ${tw(m[2])}: ہوا مگر تاخیر سے — ${dur(m[3])}`],
  [/^(الصباحي|المسائي) من (.*) لـ (.*)$/, (m) => `${m[1] === 'الصباحي' ? 'صبح کی' : 'شام کی'} شفٹ ${m[2]} سے ${m[3]} تک`],
  // (دازاي 11-13) «نص عربي ونص باكستاني» — النصوص الي بيها رقم
  [/^(.+) (\d+%|—) — منين جات؟$/, (m) => `${tw(m[1])} ${m[2]} — یہ کہاں سے آیا؟`],
  [/^(.+) \((\d+)\)$/, (m) => (UR[m[1].trim()] ? `${UR[m[1].trim()]} (${m[2]})` : m[0])],
  [/^(\d+) دقيقة · (\d+) جلسات$/, (m) => `${m[1]} منٹ · ${m[2]} سیشن`],
  [/^(\d+) دقيقة$/, (m) => `${m[1]} منٹ`],
  [/^(\d+) (أيام|يوم)$/, (m) => `${m[1]} دن`],
  [/^عندك (.+)\. كمّل شغلك أول\.$/, (m) => `آپ کے پاس ${m[1].split('، ').map((x) => { for (const [re, r] of DUE) if (re.test(x)) return x.replace(re, r); return x }).join('، ')} ہیں۔ پہلے اپنا کام مکمل کریں۔`],
  [/^حجز (\S+) · (.+?): ما صارت ووقتها فات — ما صارت لحد هسه\. صارلها (.+?) تنتظر، والمعتاد ما يتجاوز (.+?)\.( ماكو تقرير عمل\.)?$/,
    (m) => `بکنگ ${m[1]} · ${tw(m[2])}: نہیں ہوا اور وقت گزر گیا — ابھی تک نہیں ہوا۔ ${dur(m[3])} سے انتظار، عام طور پر ${dur(m[4])} سے زیادہ نہیں۔${m[5] ? ' کام کی رپورٹ نہیں ہے۔' : ''}`],
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
