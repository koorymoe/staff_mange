// المبلغ بالحروف العربية (دازاي 11-06، عرض المشاريع): 222945000 ←
// «مائتان واثنان وعشرون مليوناً وتسعمائة وخمسة وأربعون ألف دينار عراقي».
const ONES = ['', 'واحد', 'اثنان', 'ثلاثة', 'أربعة', 'خمسة', 'ستة', 'سبعة', 'ثمانية', 'تسعة', 'عشرة',
  'أحد عشر', 'اثنا عشر', 'ثلاثة عشر', 'أربعة عشر', 'خمسة عشر', 'ستة عشر', 'سبعة عشر', 'ثمانية عشر', 'تسعة عشر']
const TENS = ['', '', 'عشرون', 'ثلاثون', 'أربعون', 'خمسون', 'ستون', 'سبعون', 'ثمانون', 'تسعون']
const HUNDREDS = ['', 'مائة', 'مائتان', 'ثلاثمائة', 'أربعمائة', 'خمسمائة', 'ستمائة', 'سبعمائة', 'ثمانمائة', 'تسعمائة']

function below1000(n: number): string {
  const h = Math.floor(n / 100), r = n % 100
  const parts: string[] = []
  if (h) parts.push(HUNDREDS[h])
  if (r) {
    if (r < 20) parts.push(ONES[r])
    else {
      const o = r % 10, t = Math.floor(r / 10)
      parts.push(o ? `${ONES[o]} و${TENS[t]}` : TENS[t])
    }
  }
  return parts.join(' و')
}

// [مفرد، مثنى، جمع (٣–١٠)، تمييز (١١ فما فوق)]
const SCALES: [string, string, string, string][] = [
  ['ألف', 'ألفان', 'آلاف', 'ألف'],
  ['مليون', 'مليونان', 'ملايين', 'مليوناً'],
  ['مليار', 'ملياران', 'مليارات', 'ملياراً'],
]

function scaled(n: number, s: [string, string, string, string]): string {
  if (n === 1) return s[0]
  if (n === 2) return s[1]
  if (n >= 3 && n <= 10) return `${below1000(n)} ${s[2]}`
  return `${below1000(n)} ${s[3]}`
}

export function arabicNumberWords(value: number): string {
  let n = Math.floor(Math.abs(value))
  if (n === 0) return 'صفر'
  const groups: number[] = []
  while (n > 0) { groups.push(n % 1000); n = Math.floor(n / 1000) }
  const parts: string[] = []
  for (let i = groups.length - 1; i >= 0; i--) {
    const g = groups[i]
    if (!g) continue
    parts.push(i === 0 ? below1000(g) : scaled(g, SCALES[i - 1] ?? SCALES[2]))
  }
  return parts.join(' و')
}

export function arabicAmountWords(value: number, currency = 'دينار عراقي'): string {
  return `${arabicNumberWords(value)} ${currency}`
}
