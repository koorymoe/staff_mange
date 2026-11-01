// ═══ 🏗️ مفكك أوامر المشاريع (دازاي 11-01) ═══
// يقرا جملة عربية وحدة ويطلع الحقول — **محلي بالكامل**، ولا كلمة تطلع برّا
// الجهاز (أسماء المالكين ممنوع توصل لأي ذكاء خارجي).

export type Stage = 'TENDER' | 'SIGNED' | 'EXECUTION' | 'HANDOVER' | 'CLOSED'
export const STAGE_LABEL: Record<Stage, string> = {
  TENDER: 'عطاء', SIGNED: 'توقيع', EXECUTION: 'تنفيذ', HANDOVER: 'تسليم', CLOSED: 'مغلق',
}
const STAGE_WORDS: [RegExp, Stage][] = [
  [/عطاء|مناقصة/, 'TENDER'], [/توقيع|موقع العقد|متعاقد/, 'SIGNED'], [/تنفيذ|عمل|شغال/, 'EXECUTION'],
  [/تسليم/, 'HANDOVER'], [/مغلق|منتهي|خلص|سكر/, 'CLOSED'],
]

export type ProjectFields = {
  name?: string
  location?: string
  owner?: string
  contractIQD?: number
  durationMonths?: number
  startDate?: string
  stage?: Stage
  progress?: number
}
export type Parsed =
  | { kind: 'create'; fields: ProjectFields; missing: string[] }
  | { kind: 'update'; target: string; fields: ProjectFields }
  | { kind: 'list'; stage?: Stage }
  | { kind: 'help' }
  | { kind: 'error'; message: string }

/** أرقام عربية/فارسية ← إنگليزية، والفارزة العربية ← نقطة */
export function normDigits(s: string) {
  return s
    .replace(/[٠-٩]/g, (d) => String('٠١٢٣٤٥٦٧٨٩'.indexOf(d)))
    .replace(/[۰-۹]/g, (d) => String('۰۱۲۳۴۵۶۷۸۹'.indexOf(d)))
    .replace(/٫/g, '.')
}

/** «450 مليون» «1.2 مليار» «450,000,000» «٧٥٠ الف» ← رقم */
export function parseMoney(raw: string): number | undefined {
  const s = normDigits(raw).replace(/،/g, ',')
  const m = s.match(/(\d[\d,]*(?:\.\d+)?)\s*(مليار|مليون|ملايين|الف|ألف|آلاف|k|m|b)?/i)
  if (!m) return undefined
  const n = Number(m[1].replace(/,/g, ''))
  if (!isFinite(n)) return undefined
  const unit = (m[2] || '').toLowerCase()
  const mult = /مليار|b/.test(unit) ? 1e9 : /مليون|ملايين|m/.test(unit) ? 1e6 : /الف|ألف|آلاف|k/.test(unit) ? 1e3 : 1
  return Math.round(n * mult)
}

function stageOf(s: string): Stage | undefined {
  for (const [re, st] of STAGE_WORDS) if (re.test(s)) return st
  return undefined
}

function parseDate(s: string): string | undefined {
  const t = normDigits(s)
  let m = t.match(/(\d{4})[-/](\d{1,2})[-/](\d{1,2})/)
  if (m) return `${m[1]}-${m[2].padStart(2, '0')}-${m[3].padStart(2, '0')}`
  m = t.match(/(\d{1,2})[-/](\d{1,2})[-/](\d{4})/)
  if (m) return `${m[3]}-${m[2].padStart(2, '0')}-${m[1].padStart(2, '0')}`
  return undefined
}

/** يقسّم الأمر لقطع «مفتاح قيمة» بالفوارز */
function pieces(s: string) {
  return s.split(/[،,؛;\n]+/).map((x) => x.trim()).filter(Boolean)
}

function fieldsFrom(parts: string[]): ProjectFields {
  const f: ProjectFields = {}
  for (const p of parts) {
    const t = normDigits(p)
    let m: RegExpMatchArray | null
    if ((m = p.match(/^(?:الموقع|موقعه|المكان|في)\s*[:：]?\s*(.+)$/))) f.location = m[1].trim()
    else if ((m = p.match(/^(?:المالك|الجهة|جهة|لصالح|الزبون)\s*[:：]?\s*(.+)$/))) f.owner = m[1].trim()
    else if (/^(?:العقد|القيمة|قيمة|بقيمة|المبلغ)/.test(p)) f.contractIQD = parseMoney(p)
    else if ((m = t.match(/^(?:المدة|مدته|مدة)\s*[:：]?\s*(\d+)\s*(سنة|سنوات|سنين|شهر|اشهر|أشهر|شهور)?/))) {
      const n = Number(m[1]); f.durationMonths = /سن/.test(m[2] || '') ? n * 12 : n
    } else if (/^(?:يبدي|يبدأ|البدء|البداية|تاريخ)/.test(p)) f.startDate = parseDate(p)
    else if ((m = t.match(/^(?:مرحلة|المرحلة)\s*[:：]?\s*(.+)$/))) {
      f.stage = stageOf(m[1])
      const pr = m[1].match(/(\d{1,3})\s*%?/); if (pr) f.progress = Math.min(100, Number(pr[1]))
    } else if ((m = t.match(/^(?:انجاز|إنجاز|الانجاز|نسبة)\s*[:：]?\s*(\d{1,3})/))) f.progress = Math.min(100, Number(m[1]))
    else if ((m = t.match(/^(\d{1,3})\s*%$/))) f.progress = Math.min(100, Number(m[1]))
  }
  return f
}

export function parseCommand(input: string): Parsed {
  const s = input.trim().replace(/\s+/g, ' ')
  if (!s) return { kind: 'error', message: 'اكتب أمر' }
  if (/^(مساعدة|help|\?)$/i.test(s)) return { kind: 'help' }
  let m = s.match(/^(?:عرض|اعرض|القائمة|كلهن)\s*(.*)$/)
  if (m) return { kind: 'list', stage: m[1] ? stageOf(m[1]) : undefined }

  m = s.match(/^(?:مشروع جديد|مشروع)\s+(.+)$/)
  if (m) {
    const [first, ...rest] = pieces(m[1])
    const fields: ProjectFields = { name: first, ...fieldsFrom(rest) }
    const missing: string[] = []
    if (!fields.location) missing.push('الموقع')
    if (!fields.owner) missing.push('المالك')
    if (!fields.contractIQD) missing.push('قيمة العقد')
    if (!fields.durationMonths) missing.push('المدة')
    if (!fields.startDate) missing.push('تاريخ البدء')
    return { kind: 'create', fields, missing }
  }

  // تحديث: «<اسم> مرحلة تنفيذ 35%» أو «<اسم>، انجاز 40»
  // ⚠️ \b ما يشتغل ويا العربي بـJS — الحد = مسافة أو فارزة قبل الكلمة
  const kw = /(?:^|[\s،,])(?=(?:مرحلة|المرحلة|انجاز|إنجاز|الانجاز|نسبة|الموقع|المالك|الجهة|العقد|القيمة|المدة|يبدي|يبدأ)(?:[\s:：]|$)|[\d٠-٩]{1,3}\s*%)/
  const idx = s.search(kw)
  if (idx > 0) {
    const target = s.slice(0, idx).replace(/[،,]\s*$/, '').trim()
    const restRaw = s.slice(idx).replace(/^[\s،,]+/, '')
    // «مرحلة تنفيذ 35%» قطعة وحدة تكفي، والباقي بالفوارز
    const fields = fieldsFrom(pieces(restRaw).flatMap((p) => p.split(/\s+(?=(?:انجاز|إنجاز|نسبة)\s)/)))
    if (Object.keys(fields).length) return { kind: 'update', target, fields }
  }
  return { kind: 'error', message: 'ما فهمت الأمر — اكتب «مساعدة»' }
}

export const fmtIQD = (n: number) => {
  if (n >= 1e9) return `${(n / 1e9).toLocaleString('en-US', { maximumFractionDigits: 2 })} مليار`
  if (n >= 1e6) return `${(n / 1e6).toLocaleString('en-US', { maximumFractionDigits: 2 })} مليون`
  return n.toLocaleString('en-US')
}
