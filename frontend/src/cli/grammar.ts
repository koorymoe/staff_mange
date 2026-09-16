// ═══ نحو الـCLI ═══
//
// «واجهات الترمنال الـCLI… أريد شي يبدو وكأنه حقيقي، ماريد شي سهل».
//
// اللي يخلّي Packet Tracer يحس حقيقي مو الرسوم — هو إن الترمنال **يرد
// مثل الجهاز**: يقبل الاختصار (`int gi0/1`)، ويطلع `% Invalid input
// detected at '^' marker.` والسهم تحت **الكلمة الغلط بالضبط**، ويكلك
// `% Ambiguous command` لمن الاختصار يطابق أمرين. الفني الي اشتغل على
// سويچ حقيقي يعرف هالتفاصيل بالغريزة — وأي واحدة تنقص تكشف اللعبة.
//
// ⚠️ النحو **بيانات مو كود**: شجرة تنخزن بجدول `SimCliGrammar`. إضافة
// هواوي VRP أو TP-Link ما تحتاج تعديل بالمحرّك — تحتاج شجرة ثانية.
// (المخطط ٨: «إضافة جهاز عملية content engineering وليس feature
// programming».)

/** نوع الوسيط — يقرّر شنو يُقبل وشنو يطلع بالمساعدة. */
export type ArgKind = 'word' | 'num' | 'ipv4' | 'mask' | 'ifname' | 'vlanlist'

export interface CliNode {
  /** الكلمة الحرفية، أو `<arg>` لعقدة وسيط. */
  t: string
  /** نص المساعدة الي يطلع مع `?` — بالإنجليزي مثل الجهاز الحقيقي. */
  help: string
  /** نوع الوسيط لعقد `<arg>`. */
  arg?: ArgKind
  children?: CliNode[]

  // ═══ الأثر عند اكتمال الأمر ═══
  /** مسار بحالة الجهاز — `$1`..`$9` تنبدّل بالوسائط بالترتيب. */
  set?: string
  /** القيمة — `$1` وسيط، أو حرفية، أو `true`/`false`. */
  val?: string
  /** يحذف المسار بدل ما يكتبه (أمر `no`). */
  unset?: boolean
  /** يدخل نمطاً جديداً — `$1` مسموحة بالسياق (مثل اسم المنفذ). */
  enter?: { mode: string; ctx?: string }
  /** يطلع نمطاً واحداً للورا. */
  exit?: boolean
  /** يطلع مباشرة للنمط المسمّى — `end` بسيسكو ترجّع لـEXEC المميّز،
   *  و`return` بـVRP ترجّع لنمط المستخدم. الوجهة تنسمّى صراحةً. */
  endTo?: string
  /** يطبع خرجاً محسوباً من الحالة. */
  show?: string
  /** نص ثابت يطلع بعد التنفيذ (مثل تحذير). */
  say?: string
}

export interface CliMode {
  id: string
  /** لاحقة المحث: `>` أو `#` أو `(config)#` … */
  promptSuffix: string
  /** ═══ قالب المحث الكامل ═══
   *
   *  بعض الأنظمة **تحيط** الاسم بدل ما تلحقه: هواوي VRP تكتب
   *  `<Huawei>` بنمط المستخدم و`[Huawei]` بنمط النظام
   *  و`[Huawei-GigabitEthernet0/0/1]` بنمط المنفذ. اللاحقة لحالها
   *  ما تگدر تسوي هذا.
   *
   *  يقبل `$host` و`$ctx`. إذا مو موجود، يبقى `host + promptSuffix`
   *  — فالأنظمة الموجودة ما تتأثر. */
  promptTemplate?: string
  /** يستعمل سياق النمط باللاحقة، مثل `(config-if)#`. */
  root: CliNode[]
}

export interface CliGrammar {
  id: string
  name: string
  os: string
  /** النمط الي يبدي بيه الجلسة. */
  startMode: string
  modes: CliMode[]
  /** سطور ترحيب تطلع أول ما يفتح الترمنال. */
  banner?: string[]
  /** ═══ صيغة خرج `show`/`display` ═══
   *
   *  ⚠️ الحالة **وحدة** لكل الأنظمة (`hostname`, `interfaces.X.accessVlan`)
   *  — وهذا صح، الجهاز واحد والي يتغيّر هو **لغة عرضه**. بلا هالحقل،
   *  `display current-configuration` على جهاز VRP تطلع
   *  `switchport mode access` بصيغة سيسكو — وأي فني هواوي يعرف
   *  إنها لعبة بثانية. */
  showStyle?: 'ios' | 'vrp'
}

/** حالة الجهاز — شجرة بسيطة يكتب بيها `set` ويقرا منها `show`. */
export type CliState = Record<string, unknown>
