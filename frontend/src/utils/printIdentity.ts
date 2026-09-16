import type { IdentityFields } from './identity'

// ═══ رأس الهوية للطباعة — مصدر واحد ═══
//
// الورقة الي تطلع من النظام وتوصل بيد الزبون لازم تعرّف نفسها: كود
// الحجز، كود الزبون، الاسم، الهاتف، والليدر المسؤول. قبل، كل شاشة
// طباعة تبني رأسها بالإيد — والنتيجة إن كل ورقة تعرض شي مختلف.
//
// ⚠️ التهريب إجباري ومو اختياري. اسم زبون بيه `<` يكسر الصفحة
// المطبوعة، واسم بيه `<img onerror=...>` يصير XSS مخزّن ينفّذ بجلسة
// الي يفتح المعاينة. لهذا `esc()` بهذا الملف نفسه — حتى ما ينلگى
// طريق يبني الرأس بلا تهريب.

/** يهرّب أي نص قبل ما ينحط داخل HTML. */
export function esc(v: unknown): string {
  return String(v ?? '')
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#39;')
}

/** ستايل الرأس — ينحط مرة وحدة بـ<style> مال صفحة الطباعة. */
export const printIdentityCss = `
.ident-strip{display:flex;flex-wrap:wrap;gap:4px 14px;border:1px solid #cbd5e1;
  border-radius:8px;background:#f8fafc;padding:7px 11px;font-size:11.5px;
  margin:8px 0;direction:rtl}
.ident-strip .k{color:#64748b}
.ident-strip .v{font-weight:700;color:#0f2040}
.ident-strip .warn{color:#b45309;font-weight:700}
`

/** يبني شريط الهوية HTML. يرجّع نص فاضي لو ماكو ولا معلومة —
 *  صندوق فاضي بالورقة أسوأ من لا شي.
 *
 *  ⚠️ الحقل الفارغ **ينختفي** ما يطلع «-»: سطر مليان شرطات يوسّخ
 *  الورقة ويخلي الي مهم ما ينشاف. الاستثناء الوحيد الليدر — غيابه
 *  معلومة بحد ذاته فينكتب «بلا ليدر» صراحةً. */
export function printIdentityHtml(id: IdentityFields): string {
  const parts: string[] = []
  const add = (label: string, value?: string | null) => {
    if (!value) return
    parts.push(`<span><span class="k">${esc(label)}:</span> <span class="v">${esc(value)}</span></span>`)
  }
  add('كود الحجز', id.bookingCode)
  add('كود الزبون', id.customerCode)
  add('الزبون', id.customerName)
  add('الهاتف', id.customerPhone)
  add('الخدمة', id.serviceName)
  add('العنوان', id.address)
  if (id.leaderName) add('الليدر', id.leaderName)
  else parts.push('<span class="warn">بلا ليدر</span>')

  // الليدر لحاله ما يعرّف ورقة — لو ماكو غيره ما نطبع الشريط أصلاً
  if (parts.length <= 1) return ''
  return `<div class="ident-strip">${parts.join('')}</div>`
}
