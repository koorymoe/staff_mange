// ═══ نواة مساحة العمل ═══
//
// «أريد كأنما آني فاتح ماتلاب وجاي أربط دوائر كهربائية أو جاي أربط
// سويچات وراوترات، بس شي منظّم ومرتّب وقوي».
//
// ⚠️ **النواة ما تعرف شنو يعني «لوح شمسي» ولا «راوتر»** — هذا أهم قرار
// معماري بالمخطط الرئيسي (٤). النواة تعرف: عقدة، منفذ، وصلة، ومحرّك.
// وقوانين الكهرباء تعيش بمحرّك الكهرباء، وقوانين الشبكة بمحرّك الشبكة.
// بدون هالفصل، أول ما نضيف مجالاً رابعاً ينكسر كلشي.

/** نوع المنفذ — يقرّر شنو ينربط بشنو. */
export type PortKind = 'dc' | 'ac' | 'eth' | 'sfp' | 'signal' | 'spk' | 'fiber'

export interface PortDef {
  id: string
  label: string
  kind: PortKind
  /** القطبية للمنافذ الكهربائية — تقرّر الغلط القاتل. */
  polarity?: 'pos' | 'neg' | 'none'
  /** موقع المنفذ على رمز القطعة، نسبة ٠..١ من صندوقها. */
  x: number
  y: number
}

/** حقل قابل للتحرير بلوحة الخصائص. */
export interface ParamDef {
  id: string
  label: string
  unit?: string
  kind: 'number' | 'text' | 'select' | 'bool'
  default: string | number | boolean
  options?: { value: string; label: string }[]
  min?: number
  max?: number
  help?: string
}

/** تعريف قطعة بالكتالوگ. */
export interface PartDef {
  id: string
  domain: DomainId
  name: string
  /** سطر الموديل تحت الاسم بالمكتبة.
   *
   *  ⚠️ **أوصاف عامة مو أسماء مصنّعين**: كل محتوى المختبر مبني على
   *  أعراف منشورة مو كتالوگات، وكتابة «Cisco 2960X» تحت رمز عام
   *  تعطي انطباعاً إن القيم مأخوذة من كتالوگ ذاك الموديل — وهذا
   *  مو صحيح، ويناقض شريط السلامة الي فوگ الشاشة. */
  model?: string
  /** رمز الرسم — يختاره العارض. */
  symbol: string
  /** أبعاد الرمز بوحدات اللوح. */
  w: number
  h: number
  ports: PortDef[]
  params: ParamDef[]
  /** ═══ الهندسة ثلاثية الأبعاد ═══
   *
   *  نفس شكل `DeviceGeometry` بمحرّك التوصيل — **عمداً**: هذا الي
   *  يخلّي `buildDevice` تنعاد استعمالها كما هي بلا محوّل ولا نسخة
   *  ثانية من مولّد الأجسام.
   *
   *  ⚠️ الوحدة **متر** ١:١ مثل بقية المشاهد الثلاثية. القطعة بلا
   *  هندسة تنرسم صندوقاً عاماً — فإضافة قطعة ما تكسر المشهد. */
  geo3d?: {
    sizeM: { w: number; h: number; d: number }
    bodyColorHex?: string
    faceColorHex?: string
    /** ميلان حول المحور الأفقي — الألواح تنصب مائلة مو مسطّحة. */
    tiltDeg?: number
    features?: unknown[]
  }
  /** شرح قصير يطلع بلوحة الخصائص وبالكتالوگ. */
  about?: string
  /** تحذير سلامة يطلع لمن تنحط القطعة. */
  danger?: string
}

export type DomainId = 'electrical' | 'solar' | 'network' | 'fire' | 'audio' | 'gpon'

/** نسخة قطعة موضوعة باللوح. */
export interface LabNode {
  id: string
  partId: string
  x: number
  y: number
  /** دوران بمضاعفات ٩٠ درجة. */
  rot: 0 | 90 | 180 | 270
  label?: string
  params: Record<string, string | number | boolean>
  /** ═══ حالة سطر الأوامر للجهاز ═══
   *
   *  ⚠️ هاي **مو نسخة ثانية** من الإعدادات — هي **المصدر الوحيد** لأي
   *  شي ينهيّأ بالكونسول (الـVLAN، اسم الجهاز، وضع المنفذ). محرّك
   *  الشبكة يقراها منها مباشرة.
   *
   *  ليش يفرق؟ لأن الفني بالميدان ما «يختار VLAN للحاسبة» — يهيّئ
   *  **منفذ السويچ**. أي محاكي يخلّي الـVLAN خاصية بالحاسبة يعلّم
   *  عادة غلط، ولمن يجي الفني بالميدان ما يعرف وين يدوّر. */
  cliState?: Record<string, unknown>
}

export interface LabLink {
  id: string
  from: { node: string; port: string }
  to: { node: string; port: string }
  /** ⚠️ الكيبل **كيان له خصائص** مو خط: نوعه وطوله والترانسيفر بطرفيه.
   *  هذا الي يخلّي «الشبكة ما تشتغل بعد ١٤٠ متر» و«ترانسيفر multimode
   *  على ألياف singlemode» أعطالاً تنكشف بالمحاكي مو بالميدان. */
  params?: Record<string, string | number | boolean>
}

export interface LabDoc {
  domain: DomainId
  nodes: LabNode[]
  links: LabLink[]
}

// ═══ نتيجة المحاكاة ═══
//
// ⚠️ المحرّك ما يعدّل المستند أبداً — يرجّع **نتيجة منفصلة**. لو كتب
// النتائج جوّا العقد، ما تگدر تفرّق بين «إعداد حطّه المستخدم» و«رقم
// حسبه المحرّك»، وأول حفظ يخزن نتائج محاكاة كأنها تصميم.
export interface SimResult {
  ok: boolean
  /** رسائل عامة تطلع بشريط النتائج. */
  messages: { kind: 'info' | 'warn' | 'error'; text: string }[]
  /** قياسات لكل عقدة — تطلع كشارة فوگ القطعة. */
  nodeReadings: Record<string, { text: string; tone?: 'ok' | 'warn' | 'bad' }[]>
  /** حالة كل وصلة — لونها بالرسم. */
  linkState: Record<string, 'ok' | 'off' | 'bad'>
  /** جهد كل عقدة كهربائية بالفولت — للفحص والتشخيص. */
  netVoltages?: Record<string, number>
}

export interface DomainEngine {
  id: DomainId
  name: string
  // ⚠️ الكتالوگ ينمرّر **دائماً**، بس المحرّك يقدر يتجاهله: TypeScript
  // تسمح للتنفيذ ياخذ وسائط أقل من التوقيع. محرّك الكهرباء يحتاجه
  // (يقرا منه المنافذ)، ومحرّكا الشبكة والشمسية لا — قوانينهم على
  // الوسائط مو على الشكل، فيكتبون `run(doc)` وبس.
  run(doc: LabDoc, catalog: Record<string, PartDef>): SimResult
}
