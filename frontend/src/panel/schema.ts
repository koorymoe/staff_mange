// ═══ مخطط واجهة إعدادات الجهاز ═══
//
// «واجهات الترمنال الـCLI وGUI بالبرمجة — أريد شي يبدو وكأنه حقيقي».
//
// الـCLI انبنى بشجرة أوامر **بيانات**. وهاي أخته: واجهة الويب الي
// يفتحها الفني بالمتصفّح على الراوتر — وتنوصف بمخطط، **مو بمكوّن
// React لكل جهاز**.
//
// ⚠️ ليش بيانات؟ لأن كل راوتر عنده واجهته، ولو انكتبت كمكوّن فكل
// جهاز جديد يعني نسخة من الواجهة — والنسخ تفترق بأول تصحيح. نفس
// الفخ الي تجنّبناه بـ`CliGrammar`، ونفس الحل.
//
// ⚠️ و`engineKind: 'PANEL'` موجود بالنموذج من أول يوم وما انستُعمل —
// المكان چان محجوزاً لهذا.

export type PanelFieldKind = 'text' | 'password' | 'number' | 'select' | 'bool' | 'readonly'

export interface PanelField {
  id: string
  label: string
  kind: PanelFieldKind
  /** مفتاح بـ`params` العقدة — **هذا الي يربط الواجهة بالمحاكاة**.
   *
   *  ⚠️ الكتابة تروح لنفس المكان الي يقرا منه المحرّك، فتغيير
   *  الـVLAN بالواجهة يغيّر نتيجة المحاكاة بلا أي ربط إضافي. أي
   *  تخزين منفصل يعني قيمتين تفترقان: وحدة تشوفها بالواجهة ووحدة
   *  يحسب بيها المحرّك. */
  path?: string
  /** قيمة **محسوبة** تتجدد مع المحاكاة — للحقول القرائية.
   *  مثل قدرة الاستقبال: تنقرا من الميزانية الضوئية مو تنخزن. */
  computed?: string
  options?: { value: string; label: string }[]
  unit?: string
  help?: string
  placeholder?: string
  /** يخفي الحقل إلا لمن تكون قيمة حقل ثانٍ مطابقة — مثل بيانات
   *  PPPoE الي ما تظهر إلا بوضع PPPoE. */
  showWhen?: { path: string; equals: string }
}

export interface PanelSection {
  title: string
  note?: string
  fields: PanelField[]
}

export interface PanelTab {
  id: string
  label: string
  sections: PanelSection[]
}

export interface PanelSchema {
  id: string
  /** اسم الموديل بالشريط العلوي. */
  name: string
  /** سطر تحته — الفئة أو الوصف. */
  brandLine?: string
  /** العنوان الي يفتحه الفني بالمتصفّح. */
  address?: string
  /** تحذير يطلع بأعلى الواجهة. */
  warn?: string
  tabs: PanelTab[]
}

/** قيم الحقول القابلة للتحرير بمسوّدة. */
export type PanelDraft = Record<string, string | number | boolean>

/** يجمع القيم الابتدائية من `params` حسب المخطط. */
export function draftFrom(schema: PanelSchema, params: Record<string, unknown>): PanelDraft {
  const d: PanelDraft = {}
  for (const tab of schema.tabs) {
    for (const sec of tab.sections) {
      for (const f of sec.fields) {
        if (f.kind === 'readonly' || !f.path) continue
        const v = params[f.path]
        d[f.path] = (v as string | number | boolean) ?? (f.kind === 'bool' ? false : '')
      }
    }
  }
  return d
}

/**
 * ═══ التطبيق ═══
 *
 * ⚠️ يرجّع **مستنداً جديداً** ما يعدّل المدخل: نفس قاعدة بقية النظام،
 * والتعديل بالمكان يخلّي React ما تحس بالتغيير أحياناً.
 */
export function applyDraft(
  params: Record<string, unknown>, draft: PanelDraft,
): Record<string, unknown> {
  return { ...params, ...draft }
}

/** يقرّر إذا الحقل يظهر بالمسوّدة الحالية. */
export function fieldVisible(f: PanelField, draft: PanelDraft): boolean {
  if (!f.showWhen) return true
  return String(draft[f.showWhen.path] ?? '') === f.showWhen.equals
}
