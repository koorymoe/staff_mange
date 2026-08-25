// ═══ أنواع مختبر المحاكاة ═══
//
// تطابق الي يرجّعه السيرفر (`internal/model/sim.go`). الحقول الي شكلها
// يتغيّر من جهاز لجهاز (`terminals`, `scene`, `steps`) يمرّها Go كـJSON
// خام بلا ما يفهم محتواها — الفهم كله هنا بالواجهة، لأن المحرّك هو الي
// يقراها. هذا الي يخلّي إضافة جهاز جديد **بيانات مو كود**.

/** طرف توصيل على جهاز — سلك ملوّن أو نقطة ربط. */
export interface Terminal {
  id: string
  label: string
  colorHex?: string
  colorName?: string
  /** POWER_POS | POWER_NEG | GND | RELAY_NO | RELAY_COM | RELAY_NC | DATA | INPUT | OUTPUT … */
  kind: string
  signal?: string
  /** نسبة 0..1 من **صندوق الجهاز** — مو من الشاشة. فلمن تنرفع صورة
   *  حقيقية للجهاز بعدين، نفس الإحداثيات تشتغل فوگها بلا إعادة عمل. */
  x: number
  y: number
  description?: string
  /** شنو يصير لو انربط غلط — يطلع بالمرور وبتفسير الغلط. */
  danger?: string
  pairId?: string
  maxConnections?: number
}

/** ═══ الهندسة ثلاثية الأبعاد ═══
 *
 *  الوحدة **متر** ومقياس ١:١ — مثل ما يفرض المخطط الرئيسي (٧٫٢). الجسم
 *  يتولّد بالكود من هذي المواصفة، فماكو ملف موديل ينزّل ولا فنان ٣د
 *  ينتظر. ولمن تجي موديلات مصنّعة حقيقية تنركّب فوگ نفس المراسي. */
export interface DeviceGeometry {
  shape?: 'wall_box' | 'psu_brick' | string
  sizeM?: { w: number; h: number; d: number }
  bodyColorHex?: string
  faceColorHex?: string
  terminalPost?: { radiusM: number; heightM: number }
  features?: GeometryFeature[]
}

export type GeometryFeature =
  | { kind: 'keypad'; x: number; y: number; w: number; h: number; cols: number; rows: number }
  | { kind: 'statusLed'; x: number; y: number; channel?: string }
  | { kind: 'terminalPlate'; x0: number; y0: number; x1: number; y1: number }

/** ═══ نوع الموصّل ═══
 *
 *  التوصيل لازم يكون **مصنّفاً** (٩): RJ45 ما يسنّب على بلوك براغي.
 *  حالياً كل أطرافنا براغي، فالفحص يمر دائماً — بس المحرّك موجود، فأول
 *  جهاز شبكة ينضاف يشتغل الرفض لحاله بلا تعديل بالمحرّك. */
export type ConnectorKind = 'screw_terminal' | 'rj45' | 'dc_jack' | 'sfp' | 'usb'

export interface SimDevice {
  id: string
  categoryId: string
  brand: string
  model: string
  name: string
  summary?: string
  imagePath?: string
  engineKind: 'WIRING' | 'CLI' | 'PANEL' | string
  spec: Record<string, unknown>
  terminals: Terminal[]
  ui: Record<string, unknown>
  geometry?: DeviceGeometry
  status: string
  version: number
  sourceRef?: string
  localPractice?: string
  /** ⚠️ غير المحقّق ما يوصل متدرّباً — السيرفر يفلتره. بس نعرضه للمالك
   *  بعلامة صريحة حتى ما يعتمد عليه بالميدان. */
  verified: boolean
}

export interface SimCategory {
  id: string
  serviceId?: string
  name: string
  description?: string
  imagePath?: string
  sortOrder: number
  serviceName?: string
  exerciseCount: number
}

export interface SimLesson {
  id: string
  categoryId?: string
  deviceId?: string
  title: string
  blocks: LessonBlock[]
  sortOrder: number
  status: string
}

export type LessonBlock =
  | { t: 'text'; md: string }
  | { t: 'warn'; md: string }
  | { t: 'table'; head: string[]; rows: string[][] }
  | { t: 'quiz'; q: string; options: string[]; answer: number; why?: string }
  | { t: 'image'; path: string; caption?: string }

/** جهاز موضوع بالمشهد — `ref` اسمه داخل التمرين، و`x/y` موقعه باللوح. */
export interface SceneDevice {
  ref: string
  deviceId: string
  x: number
  y: number
}
export interface Scene {
  devices: SceneDevice[]
  background?: string
}

/** المتوقّع من الخطوة. بهالمرحلة `CONNECT` بس — الباقي يجي مع محرّكات لاحقة. */
export interface Expect {
  op: 'CONNECT' | 'DISCONNECT' | 'STATE_EQ' | 'ENERGIZE' | 'LAB_CHECK' | string
  from?: string
  to?: string
  path?: string
  value?: unknown
  /** فحوص تحديات مساحة العمل — **كلها** لازم تنجح.
   *  ⚠️ الفحص على **نتيجة المحرّك** مو على شكل المخطط: التحدي يكول
   *  «خلّي الكاميرات تشتغل» مو «حط سويچاً بالإحداثي الفلاني»، فأي
   *  طريق صحيح ينجح. */
  checks?: import('../lab/checks').LabCheck[]
}

/** تفسير غلط محدّد — `match` يطابق حركة بعينها، و`matchAny` يمسك الباقي. */
export interface WrongCase {
  match?: Expect
  matchAny?: boolean
  say: string
  penalty?: number
  /** ⚠️ القاتل **ما ينهي المحاولة** — يعرض العاقبة ويخلّيه يعيد. */
  fatal?: boolean
}

export interface Step {
  index: number
  title?: string
  instruction: string
  expect: Expect
  hint?: string
  wrong?: WrongCase[]
  weight?: number
  hintPenalty?: number
  wrongPenalty?: number
}

export interface SimExercise {
  id: string
  categoryId: string
  title: string
  brief?: string
  engineKind: string
  difficulty: number
  timeLimitSec?: number
  passScore: number
  maxAttempts?: number
  scene: Scene
  steps: Step[]
  status: string
  version: number
  sourceRef?: string
  localPractice?: string
  verified: boolean
  /** أجهزة المشهد كاملة — السيرفر يجيبهن مع التمرين بنداء واحد. */
  devices?: SimDevice[]
  /** شجرة أوامر الجهاز لتمارين `CLI` — تجي مع نفس النداء. */
  cliGrammar?: import('../cli/grammar').CliGrammar
  bestScore?: number
  passed: boolean
}

export interface SimAttempt {
  id: string
  exerciseId: string
  exerciseVersion: number
  employeeId: string
  status: 'IN_PROGRESS' | 'PASSED' | 'FAILED' | 'ABANDONED' | string
  score?: number
  stepsTotal: number
  stepsPassed: number
  hintsUsed: number
  wrongCount: number
  durationSec?: number
  state: BoardState
  startedAt: string
  finishedAt?: string
}

/** حالة اللوح — تنحفظ بالسيرفر فيگدر يوقّف ويكمّل باچر. */
export interface BoardState {
  wires?: Wire[]
  stepIndex?: number
}

/** سلك بين طرفين. المعرّف بصيغة `ref:terminalId` — مثل `lock1:t_red`. */
export interface Wire {
  from: string
  to: string
}

/** حركة يسوّيها المتدرّب باللوح. */
export interface SimAction {
  op: 'CONNECT' | 'DISCONNECT'
  from: string
  to: string
}

/** حدث ينرسل للسيرفر — **ضروري**: السيرفر يقبل عدد الخطوات الناجحة
 *  بحدود أحداث `PASS` المسجّلة فعلاً، فبلا إرسالها الدرجة تطلع صفراً. */
export interface SimEvent {
  stepIndex?: number
  kind: 'ACTION' | 'HINT' | 'WRONG' | 'PASS' | 'RESET'
  payload?: Record<string, unknown>
  atMs: number
}

export const TERMINAL_KIND_LABELS: Record<string, string> = {
  POWER_POS: 'موجب التغذية',
  POWER_NEG: 'سالب التغذية',
  GND: 'أرضي',
  RELAY_NO: 'ريلاي — تماس مفتوح (NO)',
  RELAY_NC: 'ريلاي — تماس مغلق (NC)',
  RELAY_COM: 'ريلاي — مشترك (COM)',
  DATA: 'داتا',
  INPUT: 'مدخل',
  OUTPUT: 'مخرج',
  ETHERNET: 'شبكة',
  PV_POS: 'موجب الألواح',
  PV_NEG: 'سالب الألواح',
  BAT_POS: 'موجب البطارية',
  BAT_NEG: 'سالب البطارية',
  AC_L: 'فاز (L)',
  AC_N: 'نيوترال (N)',
  UNUSED: 'غير مستعمل',
}
