// ═══ مراحل العمل بالمنظومة ═══
//
// الفني بالميدان ما يشتغل دفعة وحدة — يمشي بمراحل، وكل مرحلة تخلص
// قبل الي بعدها. والمحاكي الي يعرض كلشي مرة وحدة يعلّم عادة غلط.
//
// ⚠️⚠️ **المرحلة تتقدّم من حالة المخطط الحقيقية مو بضغطة زر.** لو
// المتدرّب يگدر يضغط «التالي» بلا ما يخلّص، الشريط يصير زينة —
// وأسوأ: يعطيه إحساساً إنه خلّص وهو ما خلّص. كل مرحلة عدها **شرط
// محسوب**، والشريط يقرا الشرط مو نيّة المستخدم.

import type { LabDoc, SimResult } from './types'

export interface Stage {
  id: string
  label: string
  hint: string
}

export const SOLAR_STAGES: Stage[] = [
  { id: 'build', label: 'بناء النظام', hint: 'حط المكوّنات: ألواح، إنفرتر، بطارية، أحمال.' },
  { id: 'wire', label: 'التوصيل', hint: 'وصّل PV والبطارية بالإنفرتر، والأحمال بمخرج AC.' },
  { id: 'config', label: 'التهيئة', hint: 'اضبط عدد الألواح بالسلسلة وجهد البنك ونافذة MPPT.' },
  { id: 'run', label: 'التشغيل', hint: 'شغّل المحاكاة وشوف القراءات على المكوّنات.' },
  { id: 'diag', label: 'التشخيص', hint: 'ماكو أخطاء خطيرة — المنظومة سليمة.' },
]

export const GENERIC_STAGES: Stage[] = [
  { id: 'build', label: 'بناء النظام', hint: 'حط المكوّنات باللوح.' },
  { id: 'wire', label: 'التوصيل', hint: 'وصّل المنافذ ببعضها.' },
  { id: 'run', label: 'التشغيل', hint: 'شغّل المحاكاة.' },
  { id: 'diag', label: 'التشخيص', hint: 'صلّح الأخطاء الخطيرة.' },
]

export const GPON_STAGES: Stage[] = [
  { id: 'build', label: 'بناء الشبكة', hint: 'حط OLT وسبليتر ووحدات ONT.' },
  { id: 'fiber', label: 'مد الليف', hint: 'وصّل منفذ PON بالسبليتر، ومخارج السبليتر بالONT.' },
  { id: 'budget', label: 'الميزانية الضوئية', hint: 'القدرة الواصلة لكل ONT ضمن حساسيته.' },
  { id: 'config', label: 'تهيئة الخدمة', hint: 'اضبط وضع WAN والـVLAN — لازم يطابق الي بالOLT.' },
  { id: 'diag', label: 'التشخيص', hint: 'ماكو أخطاء — كل المشتركين يسجّلون وعندهم خدمة.' },
]

export const CCTV_STAGES: Stage[] = [
  { id: 'build', label: 'بناء المنظومة', hint: 'حط كاميرات ومسجّلاً وسويچاً إذا احتجت.' },
  { id: 'wire', label: 'التوصيل', hint: 'وصّل الكاميرات بمنافذ PoE أو بالسويچ، والسويچ بالمسجّل.' },
  { id: 'spec', label: 'ضبط الجودة', hint: 'اضبط الدقة والإطارات والكودك — هي الي تقرّر النطاق والتخزين.' },
  { id: 'budget', label: 'التخزين والطاقة', hint: 'أيام التسجيل تكفي، والسحب الليلي ضمن ميزانية PoE.' },
  { id: 'diag', label: 'التشخيص', hint: 'ماكو أخطاء — المنظومة تسجّل كل الكاميرات المدة المطلوبة.' },
]

export const ACCESS_STAGES: Stage[] = [
  { id: 'build', label: 'بناء الباب', hint: 'حط قفلاً ووحدة تحكم ومغذّياً — وأشّر رقم الباب على كل قطعة.' },
  { id: 'egress', label: 'مسار الخروج', hint: 'كل باب لازم عليه زر خروج أو حسّاس أو كسر زجاج.' },
  { id: 'safety', label: 'السلامة', hint: 'مخارج الطوارئ بأقفال مغناطيسية، ومربوطة بتماس الحريق.' },
  { id: 'power', label: 'التغذية والبطارية', hint: 'الجهد يوصل للقفل، والبطارية تغطّي ٢٤ ساعة.' },
  { id: 'diag', label: 'التشخيص', hint: 'ماكو أخطاء — كل باب آمن ويشتغل.' },
]

export interface StageState {
  stages: Stage[]
  /** رقم المرحلة الحالية — أول وحدة ما اكتملت. */
  current: number
  /** اكتملت لو لا، لكل مرحلة. */
  done: boolean[]
}

/**
 * ═══ حساب المراحل ═══
 *
 * ⚠️ كل شرط **يُقاس من المستند أو من نتيجة المحرّك** — ماكو حالة
 * منفصلة تنخزن. يعني ما تگدر «تخلّص» مرحلة وترجع تكسر شرطها ويبقى
 * الشريط أخضر.
 */
// ⚠️⚠️ **مرحلة التشخيص ما تكتمل إلا إذا كل الي قبلها اكتملت.**
// بلا هالشرط، لوح فيه قطعة وحدة يعطي «صفر أخطاء» فتطلع آخر مرحلة
// خضراء — والمتدرّب يقرا «خلّصت» وهو ما بنى شي. الشريط الي يكذب
// بمرحلة وحدة يفقد معناه كله.
export function computeStages(doc: LabDoc, result: SimResult | null): StageState {
  const stages = doc.domain === 'solar' ? SOLAR_STAGES
    : doc.domain === 'gpon' ? GPON_STAGES
    : doc.domain === 'cctv' ? CCTV_STAGES
    : doc.domain === 'access' ? ACCESS_STAGES : GENERIC_STAGES
  const has = (partId: string) => doc.nodes.some((n) => n.partId === partId)
  const errors = (result?.messages ?? []).filter((m) => m.kind === 'error').length

  const done: boolean[] = []
  if (doc.domain === 'solar') {
    const built = has('pv_panel') && has('inverter')
    // التوصيل: الإنفرتر عليه وصلات بمدخل PV ومخرج AC
    const inv = doc.nodes.find((n) => n.partId === 'inverter')
    const linked = (port: string) =>
      !!inv && doc.links.some((l) =>
        (l.from.node === inv.id && l.from.port === port) || (l.to.node === inv.id && l.to.port === port))
    const wired = linked('pv_pos') && linked('pv_neg') && linked('ac_out')
    // التهيئة: ماكو تحذير ستring — يعني النافذة والجهد مضبوطين
    const cfgWarn = (result?.messages ?? []).some((m) =>
      m.kind !== 'info' && /MPPT|Voc|بنك البطاريات/.test(m.text))
    done.push(built)
    done.push(built && wired)
    done.push(built && wired && !!result && !cfgWarn)
    done.push(!!result && built && wired)
    done.push(!!result && errors === 0 && done.every(Boolean))
  } else if (doc.domain === 'access') {
    const built = (has('mag_lock') || has('electric_strike')) && has('ac_controller') && has('ac_psu')
    const msgs = result?.messages ?? []
    // ⚠️ الأنماط تطابق **نص المحرّك حرفياً** — النمط الي ما ينطابق
    // يعني مرحلة تبقى خضراء دائماً، وهذا انلقى مرتين بمرحلة الكاميرات.
    const noErr = (re: RegExp) => !!result && !msgs.some((m) => m.kind === 'error' && re.test(m.text))
    const egressOk = noErr(/وسيلة خروج حر|كسر الزجاج/)
    const safeOk = noErr(/مخرج طوارئ وعليه قفل كهربائي|بلا ربط بإنذار الحريق|تماس الحريق \*\*مجسور\*\*/)
    const powerOk = noErr(/الجهد الواصل|أمبير·ساعة|فوگ طاقته|دايود/)
    done.push(built)
    done.push(built && egressOk)
    done.push(built && safeOk)
    done.push(built && powerOk)
    done.push(!!result && errors === 0 && done.every(Boolean))
  } else if (doc.domain === 'cctv') {
    const built = has('ip_camera') && has('nvr')
    const wired = doc.links.length > 0
    const msgs = result?.messages ?? []
    // ⚠️ «ضبط الجودة» يكتمل لمن ما يبقى تحذير إطارات — مو لمن يلمس
    // المستخدم الخانة. المرحلة تقرا الشرط مو النية.
    const specOk = !!result && !msgs.some((m) => /إطاراً/.test(m.text))
    const budgetOk = !!result && !msgs.some((m) =>
      m.kind === 'error' && /(يوم تسجيل|بالليل|منفذ PoE|حد إدخال|قنوات)/.test(m.text))
    done.push(built)
    done.push(built && wired)
    done.push(built && wired && specOk)
    done.push(built && wired && budgetOk)
    done.push(!!result && errors === 0 && done.every(Boolean))
  } else if (doc.domain === 'gpon') {
    const built = has('olt') && has('ont')
    const fiber = doc.links.length > 0
    const msgs = result?.messages ?? []
    const budgetOk = !!result && !msgs.some((m) => m.kind === 'error' && /القدرة الواصلة|ما يسجّل|إشباع/.test(m.text))
    const cfgOk = !!result && !msgs.some((m) => m.kind === 'error' && /PPPoE|VLAN/.test(m.text))
    done.push(built)
    done.push(built && fiber)
    done.push(budgetOk && built && fiber)
    done.push(cfgOk && built && fiber)
    done.push(!!result && errors === 0 && done.every(Boolean))
  } else {
    done.push(doc.nodes.length > 0)
    done.push(doc.links.length > 0)
    done.push(!!result && doc.links.length > 0)
    done.push(!!result && errors === 0 && done.every(Boolean))
  }

  // ⚠️ المرحلة الحالية = أول وحدة **ما اكتملت**، مو آخر وحدة اكتملت:
  // لو المتدرّب رجع وكسر شرط مرحلة سابقة، الشريط يرجع إلها.
  const current = done.findIndex((d) => !d)
  return { stages, done, current: current === -1 ? stages.length - 1 : current }
}

// ═══ معايير السلامة والاختبار السريع ═══
//
// ⚠️ **قائمة تنعلّم لحالها من نتيجة المحرّك** — مو خانات يأشّر عليها
// المتدرّب. قائمة سلامة يأشّر عليها المستخدم بنفسه تنعبّي بثانيتين
// بلا ما ينفحص شي، وتصير ورقة تُملأ مو فحصاً يُجرى. وهذا بالضبط الي
// يصير بالميدان لمن يكون النموذج ورقياً.

export interface SafetyCheck {
  id: string
  label: string
  /** يمر · ينتظر (ماكو بيانات بعد) · يفشل */
  state: 'pass' | 'idle' | 'fail'
  why?: string
}

export interface QuickTest {
  checks: SafetyCheck[]
  /** الدرجة من ١٠٠ — نسبة الفحوص الناجحة من المفحوصة فعلاً. */
  score: number
  /** كم فحص انفحص فعلاً (مو `idle`). */
  tested: number
}

/**
 * ⚠️ الدرجة تنحسب من **الفحوص الي انفحصت فعلاً** بس. لو حسبنا
 * الـ`idle` كفشل، مخطط ما انشغّل بعد يعطي صفراً ويخوّف المتدرّب بلا
 * سبب. ولو حسبناها كنجاح، لوح فاضي يعطي ١٠٠ — وهذا أسوأ.
 */
export function quickTest(doc: LabDoc, result: SimResult | null): QuickTest {
  const msgs = result?.messages ?? []
  const hasErr = (re: RegExp) => msgs.some((m) => m.kind === 'error' && re.test(m.text))
  const hasWarn = (re: RegExp) => msgs.some((m) => m.kind !== 'info' && re.test(m.text))
  const ran = !!result

  const mk = (id: string, label: string, ok: boolean, why?: string): SafetyCheck =>
    ({ id, label, state: !ran ? 'idle' : ok ? 'pass' : 'fail', why: ok ? undefined : why })

  const checks: SafetyCheck[] = []

  if (doc.domain === 'solar') {
    checks.push(mk('polarity', 'القطبية ومداخل الإنفرتر', !hasErr(/مدخل البطارية|مدخل PV/),
      'ألواح على مدخل بطارية أو العكس — يحرق الإنفرتر فوراً.'))
    checks.push(mk('voc', 'جهد الستring ببرد الشتاء', !hasErr(/Voc/),
      'Voc بالبرد فوگ حد الإنفرتر — يحرقه بأول صباح بارد.'))
    checks.push(mk('mppt', 'نافذة MPPT', !hasWarn(/MPPT/),
      'الستring برّا نافذة MPPT — الإنفرتر ما يشتغل بأعلى كفاءة أو ما يبدي أصلاً.'))
    checks.push(mk('bank', 'مطابقة جهد بنك البطاريات', !hasErr(/بنك البطاريات/),
      'جهد البنك مو مطابق لإعداد الإنفرتر — ما يشحن ولا يشتغل.'))
    checks.push(mk('load', 'قدرة الأحمال ضمن الحد', !hasErr(/قدرة الإنفرتر/),
      'الأحمال فوگ قدرة الإنفرتر — يفصل على حمل زائد.'))
    // ⚠️ `&&` مو `||`: الشرطان **لازم** الاثنان — الاحتياطي يكفي
    // **و** الإنتاج اليومي يكفي. بـ`||` الفحص ينجح لو واحد منهما
    // تمام، فمنظومة احتياطها ساعة ونص تطلع «سليمة».
    checks.push(mk('backup', 'الاحتياطي والإنتاج اليومي',
      !hasWarn(/تغطّي الأحمال/) && !hasWarn(/ما يكفي، تحتاج ألواحاً/),
      'الاحتياطي أو الإنتاج اليومي ما يكفي الأحمال.'))
  } else if (doc.domain === 'access') {
    // ⚠️ **القاتل أول واحد** بالقائمة: ترتيب الفحوص هو ترتيب
    // الأولوية الي يقراها المتدرّب، وفحص يموّت ما ينحط بالآخر.
    checks.push(mk('failsafe', '☠️ مخارج الطوارئ تنفتح بانقطاع التيار',
      !hasErr(/مخرج طوارئ وعليه قفل كهربائي/),
      'باب مخرج طوارئ بقفل يبقى مقفلاً بانقطاع التيار — يحبس الناس.'))
    checks.push(mk('egress', 'كل باب عليه مخرج حر',
      !hasErr(/وسيلة خروج حر/) && !hasErr(/كسر الزجاج/),
      'باب بلا زر خروج أو حسّاس أو زجاج — أو زجاج مربوط بالوحدة.'))
    // ⚠️ النمط بقدر الجملة الي تخصّه: `/إنذار الحريق/` لحاله يخطف
    // رسالة البطارية (تذكر «حساب على طريقة إنذار الحريق» كمقارنة)
    // فيفشّل فحصاً سليماً — وفحص يفشل بلا سبب يخلّي المتدرّب يدوّر
    // على عطل ما موجود.
    checks.push(mk('fire', 'الأقفال تفصل بإنذار الحريق',
      !hasErr(/بلا ربط بإنذار الحريق/) && !hasErr(/تماس الحريق \*\*مجسور\*\*/),
      'قفل مغناطيسي بلا تماس حريق — الحريق يصير والكهرباء شغّالة.'))
    checks.push(mk('voltage', 'الجهد يوصل للقفل', !hasErr(/الجهد الواصل/),
      'هبوط جهد على مسافة القفل — يمسك ضعيفاً أو ما يفتح.'))
    checks.push(mk('battery', 'البطارية تغطّي ٢٤ ساعة', !hasErr(/أمبير·ساعة/) && !hasErr(/فوگ طاقته/),
      'البطارية ما تكفي أو السحب فوگ طاقة المغذّي.'))
    checks.push(mk('diode', 'دايود على كل ملف قفل', !hasErr(/دايود/),
      'بلا دايود، نبضة الرجوع تاكل اللوحة خلال أشهر.'))
  } else if (doc.domain === 'cctv') {
    checks.push(mk('channels', 'عدد القنوات يكفي الكاميرات', !hasErr(/على مسجّل \d+ قنوات/),
      'كاميرات أكثر من قنوات المسجّل — الزيادة ما تنسجّل أصلاً.'))
    checks.push(mk('inrate', 'إدخال المسجّل يستوعب البث', !hasErr(/حد إدخال/),
      'مجموع البث فوگ حد المسجّل — تسجيل متقطّع.'))
    checks.push(mk('storage', 'أيام التسجيل المطلوبة', !hasErr(/يوم تسجيل/),
      'القرص ما يكفي المدة — المسجّل يمسح الأقدم تلقائياً.'))
    checks.push(mk('poeNight', 'سحب PoE **ليلاً**', !hasErr(/بالليل/),
      'السحب الليلي فوگ الميزانية — الكاميرات تفصل بالظلام بس.'))
    checks.push(mk('uplink', 'الرابط الصاعد ما يخنق البث', !hasErr(/عنق الزجاجة/),
      'كل الكاميرات تمر برابط سعته أقل من مجموع بثها.'))
  } else if (doc.domain === 'gpon') {
    checks.push(mk('budget', 'الميزانية الضوئية لكل مشترك',
      !msgs.some((m) => m.kind === 'error' && /القدرة الواصلة/.test(m.text)),
      'اكو مشترك برّا حدود الحساسية — ما يسجّل أو مشبع.'))
    checks.push(mk('margin', 'هامش ٣ ديسيبل على الأقل',
      !hasWarn(/بلا هامش/), 'مسار بلا هامش — أي لحام أو اتساخ يطيّح الخدمة.'))
    checks.push(mk('capacity', 'سعة منفذ PON', !hasErr(/سعته/),
      'مشتركون أكثر من سعة المنفذ — الزيادة ما تنسجّل.'))
    checks.push(mk('service', 'تهيئة الخدمة (WAN وVLAN)',
      !hasErr(/PPPoE|VLAN/), 'ONT مسجّل ضوئياً بلا خدمة — الضوء أخضر وماكو إنترنت.'))
    checks.push(mk('splitter', 'مخارج السبليتر ضمن نسبته', !hasErr(/أكثر من مخارجه/),
      'سبليتر عليه مخارج أكثر من نسبته.'))
  } else {
    checks.push(mk('errors', 'ماكو أخطاء خطيرة', msgs.filter((m) => m.kind === 'error').length === 0,
      'بعدها أخطاء خطيرة بالمخطط.'))
    checks.push(mk('links', 'كل الوصلات سليمة',
      Object.values(result?.linkState ?? {}).every((s) => s !== 'bad'),
      'اكو وصلة مكسورة.'))
  }

  const tested = checks.filter((c) => c.state !== 'idle').length
  const passed = checks.filter((c) => c.state === 'pass').length
  return { checks, tested, score: tested ? Math.round((passed / tested) * 100) : 0 }
}
