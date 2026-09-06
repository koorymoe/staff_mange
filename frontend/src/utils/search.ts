// ═══ البحث العربي — أداة موحّدة لكل خانات البحث بالنظام ═══
//
// المشكلة الي تحلها: البحث بالنظام كان يقارن النص حرفياً، والعربي ما
// ينكتب بشكل واحد. يعني:
//
//   المستخدم يكتب      بالقاعدة مخزون      النتيجة قبل
//   ─────────────      ──────────────      ───────────
//   احمد               أحمد                 ٠ نتائج ❌
//   فاطمه              فاطمة                ٠ نتائج ❌
//   مصطفي              مصطفى                ٠ نتائج ❌
//   موسسة              مؤسسة                ٠ نتائج ❌
//   ٠٧٧٠١٢٣٤٥٦٧        07701234567          ٠ نتائج ❌
//   "احمد " (مسافة)    أحمد                 ٠ نتائج ❌
//
// وهذا مو خطأ نظري: الموظف يدور على زبون موجود، يطلعله «ماكو نتائج»،
// فيفتح ملف زبون جديد — وينخلق زبون مكرر. أو يكول للزبون «انت مو
// بالنظام». الحل إننا نطبّع الطرفين قبل المقارنة.
//
// ما نغيّر البيانات المخزونة أبداً — نطبّع نسخة بالذاكرة وقت المقارنة بس.

const ARABIC_INDIC = '٠١٢٣٤٥٦٧٨٩'
const EXTENDED_INDIC = '۰۱۲۳۴۵۶۷۸۹' // الفارسية — تطلع من بعض لوحات المفاتيح

/**
 * normalize يوحّد شكل النص حتى المقارنة تشتغل مهما اختلفت طريقة الكتابة:
 *
 *  • الهمزات: أ إ آ ٱ → ا     (احمد = أحمد)
 *  • التاء المربوطة: ة → ه    (فاطمة = فاطمه)
 *  • الألف المقصورة: ى → ي    (مصطفى = مصطفي)
 *  • الهمزة على واو/ياء: ؤ ئ → و ي   (مؤسسة = موسسة)
 *  • التشكيل والتطويل: ينحذفون  (مُحَمَّد = محمد، محمـــد = محمد)
 *  • الأرقام العربية والفارسية → إنجليزية  (٠٧٧٠ = 0770)
 *  • المسافات المكررة والأطراف: تنضغط لمسافة وحدة
 *  • الإنجليزي: يصير صغير  (AHMED = ahmed)
 */
export function normalize(input: string | null | undefined): string {
  if (!input) return ''
  let s = String(input)

  // التشكيل والتطويل أول شي — حتى ما يقطعون المطابقة بالخطوات الجاية
  s = s.replace(/[ً-ْٰـ]/g, '')

  s = s
    .replace(/[أإآٱ]/g, 'ا')
    .replace(/ة/g, 'ه')
    .replace(/ى/g, 'ي')
    .replace(/ؤ/g, 'و')
    .replace(/ئ/g, 'ي')

  // الأرقام: لوحة مفاتيح الموبايل بالعربي تطلّع ٠١٢، والمخزون بالقاعدة
  // إنجليزي — بدون هذا السطر البحث برقم الهاتف من الموبايل ما يشتغل أبداً
  s = s.replace(/[٠-٩]/g, (d) => String(ARABIC_INDIC.indexOf(d)))
  s = s.replace(/[۰-۹]/g, (d) => String(EXTENDED_INDIC.indexOf(d)))

  return s.replace(/\s+/g, ' ').trim().toLowerCase()
}

/**
 * normalizeDigits يشيل كل شي مو رقم — للبحث برقم الهاتف تحديداً.
 * الموظف يكتب «0770 123 4567» أو «0770-123-4567» والمخزون «07701234567».
 */
export function digitsOnly(input: string | null | undefined): string {
  return normalize(input).replace(/\D/g, '')
}

/**
 * matches يفحص هل النص المطلوب موجود بأي حقل من الحقول المعطاة.
 *
 * يفصل الأرقام عن الحروف تلقائياً: لو المستخدم كتب أرقام بس، يقارن
 * أرقام مقابل أرقام (فيلكه رقم الهاتف مهما كان مكتوب بشرطات أو مسافات).
 *
 *   matches([c.name, c.phone, c.code], search)
 */
export function matches(fields: (string | null | undefined)[], query: string): boolean {
  const q = normalize(query)
  if (!q) return true

  const qDigits = digitsOnly(query)
  // «أرقام بس» — يعني على الأغلب رقم هاتف أو كود
  const numericQuery = qDigits.length > 0 && qDigits === q.replace(/\s/g, '')

  for (const field of fields) {
    if (field === null || field === undefined) continue
    if (normalize(field).includes(q)) return true
    if (numericQuery && digitsOnly(field).includes(qDigits)) return true
  }
  return false
}

/**
 * makeFilter يرجّع دالة فلترة جاهزة — للاستعمال مع useMemo:
 *
 *   const filtered = useMemo(
 *     () => rows.filter(makeFilter(search, (r) => [r.name, r.phone])),
 *     [rows, search],
 *   )
 */
export function makeFilter<T>(query: string, pick: (row: T) => (string | null | undefined)[]) {
  const q = normalize(query)
  if (!q) return () => true
  return (row: T) => matches(pick(row), query)
}
