// ═══ فيزياء اللوح الكهروضوئي — منحنى IV حقيقي ═══
//
// «أريده أكثر دقة بكل دقة».
//
// المحرّك القديم چان يضرب: `عدد الألواح × القدرة × ساعات الشمس`.
// وهذا **يشتغل بالورقة وينهار بالميدان**، لأن اللوح مو مصدر قدرة
// ثابت — هو **منحنى**. القدرة الي يعطيها تعتمد على الجهد الي يسحبه
// منه الإنفرتر، وعلى الإشعاع، وعلى **حرارة الخلية** (مو حرارة الجو).
//
// وثلاثة أشياء ما چان يگدر يفسّرها الضرب الخطي، وهاي بالضبط الي
// توگف الفني:
//
// ١) **ليش المنظومة تنتج أقل بالصيف الحار من الربيع المعتدل؟**
//    الإشعاع أعلى بالصيف، بس حرارة الخلية توصل ٦٥ درجة فينزل الجهد
//    ٪١٣ — والضرب الخطي يقول العكس.
// ٢) **ليش Voc بالبرد يحرق الإنفرتر؟** لأن Voc يرتفع بالبرودة،
//    و«القدرة» ما إلها علاقة — الي يحرق هو **الجهد بلا حمل** بصباح
//    شتوي صافي.
// ٣) **ليش ستring بظل جزئي ينهار كله؟** لأن الخلايا بالسلسلة تمرّر
//    **نفس التيار**، فأضعف خلية تحكم السلسلة.
//
// ⚠️⚠️ **درجة الدقة انتقلت من `F1` لـ`F3`**: منحنى IV بنموذج الدايود
// الأحادي + تصحيح حراري + تتبّع MPP حقيقي + محاكاة يوم كامل بخطوات.
// الباقي برّا النطاق: التشتّت الطيفي، تدهور LID/PID، تأثير الغبار
// التراكمي، وتعدّد قمم المنحنى بالتظليل المعقّد.
//
// ⚠️⚠️ والمعاملات **قوانين فيزياء وأعراف منشورة — مو كتالوگ موديل**:
// ثابت بولتزمان وشحنة الإلكترون قيم كونية، ومعاملات الحرارة
// (٪٠٫٣٣−/درجة للجهد) أعراف صناعية منشورة. ماكو ولا رقم مأخوذ من
// ورقة بيانات شركة، والمحتوى يبقى `verified = FALSE`.

/** شحنة الإلكترون (كولوم). */
const Q = 1.602176634e-19
/** ثابت بولتزمان (جول/كلفن). */
const K = 1.380649e-23
/** صفر مئوي بالكلفن. */
const T0K = 273.15
/** ظروف الاختبار القياسية: ١٠٠٠ واط/م² و٢٥ درجة. */
export const STC_IRR = 1000
export const STC_T = 25

/** مواصفات اللوح — كلها من لوحة بيانات اللوح الظاهرة عليه. */
export interface PanelSpec {
  /** جهد الدارة المفتوحة عند STC. */
  voc: number
  /** تيار القصر عند STC. */
  isc: number
  /** جهد أعظم قدرة عند STC. */
  vmp: number
  /** تيار أعظم قدرة عند STC. */
  imp: number
  /** عدد الخلايا بالسلسلة داخل اللوح (٦٠ · ٧٢ · ١٤٤ نصفية). */
  cells: number
  /** معامل حرارة الجهد (نسبة/درجة، سالب). */
  tcV?: number
  /** معامل حرارة التيار (نسبة/درجة، موجب). */
  tcI?: number
}

/** ⚠️ أعراف صناعية منشورة للسليكون البلوري — مو كتالوگ موديل. */
const DEF_TC_V = -0.0033
const DEF_TC_I = 0.0005

export interface PanelModel {
  /** التيار عند جهد معيّن. */
  currentAt: (v: number) => number
  /** القدرة عند جهد معيّن. */
  powerAt: (v: number) => number
  /** Voc بعد التصحيح الحراري والإشعاعي. */
  voc: number
  /** Isc بعد التصحيح. */
  isc: number
  /** حرارة الخلية المستعملة. */
  cellT: number
  /** معامل المثالية المعاير — للفحص والعرض. */
  nIdeal: number
  /** المقاومة التسلسلية المعايرة (أوم). */
  rs: number
}

/**
 * ═══ حرارة الخلية ═══
 *
 * ⚠️⚠️ **حرارة الخلية مو حرارة الجو** — وهذا أكثر شي ينُسى. اللوح
 * تحت شمس الظهر يوصل **٢٥–٣٠ درجة فوگ** حرارة الجو. فبيوم ٤٥ درجة
 * (وهذا عادي عدنا بالصيف) الخلية تكون ٧٦ درجة، والجهد ينزل ٪١٧.
 *
 * المعادلة القياسية بـNOCT: الخلية = الجو + (NOCT−٢٠)×الإشعاع/٨٠٠.
 * وNOCT الشائع ٤٥ درجة.
 */
export function cellTemp(ambientC: number, irradiance: number, noct = 45): number {
  return ambientC + ((noct - 20) * irradiance) / 800
}

interface Fit { a: number; rs: number; nIdeal: number }

/** ذاكرة المعايرة — المعايرة بحث رقمي، وإعادته لكل نقطة على المنحنى
 *  تخلّي رسم منحنى واحد يعيدها ١٢٠ مرة بلا فايدة. */
const fitCache = new Map<string, Fit>()

/**
 * ═══ معايرة النموذج على نقاط لوحة البيانات ═══
 *
 * ⚠️⚠️ **هنا انكسر النموذج الأول.** أول محاولة ثبّتُ فيها معامل
 * المثالية `n = 1.2` تخميناً وأهملتُ المقاومة التسلسلية — فطلع
 * `Pmax = 425` واط بدل ٥٥٠، **خطأ ٪٢٣**. والفحص هو الي مسكها: بلاه
 * چان انبنى فوگها عارض منحنيات ودرس كامل، وكلهم يعرضون رقماً غلطاً
 * بثقة.
 *
 * الصح إن `n` و`Rs` **ما ينتخمّنون — ينعايرون** على نقاط اللوحة:
 *
 * ١) المنحنى لازم يمرّ بـ`(Vmp, Imp)` بالضبط ← منها `Rs` تنحسب
 *    تحليلياً لكل `n` مفترَض.
 * ٢) وقمة المنحنى لازم تصير **عند** `Vmp` ← وهذا الي يختار `n`.
 *
 * فيصير بحثاً بمتغيّر واحد، ونتيجته إن `Pmax = Vmp × Imp` بالضبط —
 * مو تقريباً.
 */
function fitPanel(spec: PanelSpec): Fit {
  const key = `${spec.voc}|${spec.isc}|${spec.vmp}|${spec.imp}|${spec.cells}`
  const hit = fitCache.get(key)
  if (hit) return hit

  const vtCell = (K * (STC_T + T0K)) / Q
  let best: Fit = { a: 1.2 * vtCell * spec.cells, rs: 0, nIdeal: 1.2 }
  let bestErr = Infinity

  for (let n = 0.8; n <= 2.2; n += 0.005) {
    const a = n * vtCell * spec.cells
    const is = spec.isc / (Math.exp(spec.voc / a) - 1)
    // من I(Vmp) = Imp:  Vmp + Imp·Rs = a·ln((Isc−Imp)/Is + 1)
    const rs = (a * Math.log((spec.isc - spec.imp) / is + 1) - spec.vmp) / spec.imp
    if (!Number.isFinite(rs) || rs < 0 || rs > 2) continue

    // ⚠️ نفحص إن القمة فعلاً عند Vmp — مو نفترضها.
    const cur = (v: number) => solveI(v, spec.isc, is, a, rs)
    let peakV = 0, peakP = -1
    for (let k = 1; k <= 160; k++) {
      const v = (spec.voc * k) / 160
      const p = v * cur(v)
      if (p > peakP) { peakP = p; peakV = v }
    }
    const err = Math.abs(peakV - spec.vmp)
    if (err < bestErr) { bestErr = err; best = { a, rs, nIdeal: n } }
  }
  fitCache.set(key, best)
  return best
}

/**
 * ═══ حلّ التيار عند جهد معيّن ═══
 *
 * ⚠️ المعادلة **ضمنية** لمن تدخل `Rs`: التيار يظهر بالطرفين
 * (`I = Isc − Is(exp((V+I·Rs)/a) − 1)`). نحلّها بتكرار نقطة ثابتة
 * مخمَّد — والتخميد ضروري: بلاه التكرار يتذبذب ويتباعد قرب Voc.
 */
function solveI(v: number, iph: number, is: number, a: number, rs: number): number {
  let i = iph
  for (let k = 0; k < 60; k++) {
    const next = iph - is * (Math.exp((v + i * rs) / a) - 1)
    const damped = i + 0.35 * (next - i)
    if (Math.abs(damped - i) < 1e-9) { i = damped; break }
    i = Math.max(-1, Math.min(iph * 1.2, damped))
  }
  return Math.max(0, i)
}

/**
 * ═══ نموذج اللوح عند إشعاع وحرارة معيّنين ═══
 *
 * المعايرة تصير **مرة وحدة عند الظروف القياسية**، وبعدها القيم
 * تُترجم لظروف التشغيل: التيار يتبع الإشعاع خطّياً، والجهد يتبع
 * الحرارة، و`Rs` و`n` ثوابت الجهاز ما تتغيّر.
 *
 * ⚠️ **التيار خطّي مع الإشعاع والجهد لوغاريتمي** — وهاي الي تخلّي
 * يوم غائم «يشتغل» وما ينتج: الفني يقيس الجهد فيلگاه شبه طبيعي،
 * والتيار (الي ما يقيسه) نازل للثلث.
 */
export function panelModel(spec: PanelSpec, irradiance: number, ambientC: number): PanelModel {
  return panelModelAtCell(spec, irradiance, cellTemp(ambientC, irradiance, 45))
}

/**
 * ═══ نفس النموذج بحرارة **خلية** معطاة ═══
 *
 * ⚠️⚠️ **الفرق بين الدالتين هو الفرق الي وگع بيه فحصي أنا.** الظروف
 * القياسية (STC) معرَّفة بحرارة **خلية** ٢٥ درجة — مو حرارة جو ٢٥.
 * وبجو ٢٥ وإشعاع كامل تصير الخلية ٥٦ درجة، فالجهد ينزل ٪١٠ ويطلع
 * «النموذج غلط» وهو صحيح.
 *
 * ولهذا الدالتان منفصلتان بالاسم: `panelModel` للميدان (تعرف حرارة
 * الجو)، وهاي للمعايرة والمقارنة بلوحة البيانات. خلطهما يخلّي كل
 * مقارنة بلوحة بيانات تطلع فاشلة بـ٪١٠ بلا سبب ظاهر.
 */
export function panelModelAtCell(spec: PanelSpec, irradiance: number, cellC: number): PanelModel {
  const fit = fitPanel(spec)
  const cellT = cellC
  const dT = cellT - STC_T
  const tcV = spec.tcV ?? DEF_TC_V
  const tcI = spec.tcI ?? DEF_TC_I
  const g = Math.max(0, irradiance) / STC_IRR

  if (g <= 0) {
    return {
      currentAt: () => 0, powerAt: () => 0,
      voc: 0, isc: 0, cellT, nIdeal: fit.nIdeal, rs: fit.rs,
    }
  }

  // `a` يتناسب مع الحرارة المطلقة (الجهد الحراري ∝ T)
  const a = (fit.a * (cellT + T0K)) / (STC_T + T0K)
  const isc = spec.isc * g * (1 + tcI * dT)
  const voc = Math.max(0, spec.voc * (1 + tcV * dT) + a * Math.log(g))
  const is = voc > 0 ? isc / (Math.exp(voc / a) - 1) : 0

  const currentAt = (v: number): number => {
    if (v < 0) return isc
    if (v >= voc) return 0
    return solveI(v, isc, is, a, fit.rs)
  }
  return {
    currentAt, powerAt: (v) => v * currentAt(v),
    voc, isc, cellT, nIdeal: fit.nIdeal, rs: fit.rs,
  }
}

export interface MppResult {
  /** جهد أعظم قدرة. */
  vmp: number
  /** تيار أعظم قدرة. */
  imp: number
  /** أعظم قدرة (واط). */
  pmax: number
}

/**
 * ═══ تتبّع نقطة أعظم قدرة ═══
 *
 * ⚠️ **بحث فعلي على المنحنى** مو قراءة رقم مخزون. هذا الي يخلّي
 * تأثير الحرارة والإشعاع والتظليل يظهر **لحاله** بلا ما نكتب له
 * قاعدة: تنزل القدرة لأن النقطة العظمى نفسها انتقلت.
 *
 * مسح خشن ثم تنقيح — أسرع من نيوتن وأمتن (المنحنى ما بيه اشتقاق
 * تحليلي بسيط عند الأطراف).
 */
export function findMpp(m: PanelModel, steps = 200): MppResult {
  if (m.voc <= 0) return { vmp: 0, imp: 0, pmax: 0 }
  let bestV = 0, bestP = -1
  for (let i = 1; i <= steps; i++) {
    const v = (m.voc * i) / steps
    const p = m.powerAt(v)
    if (p > bestP) { bestP = p; bestV = v }
  }
  // تنقيح حول القمة
  const w = m.voc / steps
  for (let i = -20; i <= 20; i++) {
    const v = bestV + (w * i) / 20
    if (v <= 0 || v >= m.voc) continue
    const p = m.powerAt(v)
    if (p > bestP) { bestP = p; bestV = v }
  }
  return { vmp: bestV, imp: m.currentAt(bestV), pmax: bestP }
}

/** نقطة على المنحنى — للرسم. */
export interface IvPoint { v: number; i: number; p: number }

export function ivCurve(m: PanelModel, points = 120): IvPoint[] {
  const out: IvPoint[] = []
  if (m.voc <= 0) return out
  for (let k = 0; k <= points; k++) {
    const v = (m.voc * k) / points
    const i = m.currentAt(v)
    out.push({ v, i, p: v * i })
  }
  return out
}

/**
 * ═══ منحنى الإشعاع خلال اليوم ═══
 *
 * ⚠️ **جيبي مو ثابت.** «٥ ساعات شمس ذروة» رقم مفيد للتقدير السريع
 * وكارثة للتصميم: المنظومة ما تشتغل ٥ ساعات بكامل قدرتها وتنطفي —
 * هي تطلع وتنزل. والحمل الي يشتغل الساعة ٧ صباحاً يشوف ٪٢٠ من
 * القدرة، مو ٪١٠٠.
 *
 * `peakIrr` ذروة الظهر (واط/م²) و`dayHours` طول النهار.
 */
export function irradianceAt(hour: number, peakIrr = 1000, sunrise = 6, sunset = 18): number {
  if (hour <= sunrise || hour >= sunset) return 0
  const t = (hour - sunrise) / (sunset - sunrise)
  return peakIrr * Math.sin(Math.PI * t)
}

export interface DayPoint {
  hour: number
  irradiance: number
  ambientC: number
  cellT: number
  pmax: number
  vmp: number
}

/**
 * ═══ محاكاة يوم كامل ═══
 *
 * ⚠️ الطاقة = **تكامل** المنحنى مو ضرب. والفرق مو تجميلياً: منظومة
 * محسوبة بـ«٥ ساعات ذروة» تطلع أعلى من الواقع لأن الحرارة تاكل من
 * الذروة بالضبط بالساعات الي فيها أعلى إشعاع.
 */
export function simulateDay(
  spec: PanelSpec,
  opts: { panels: number; peakIrr?: number; minC?: number; maxC?: number; stepH?: number },
): { points: DayPoint[]; kwh: number; peakW: number } {
  const step = opts.stepH ?? 0.5
  const minC = opts.minC ?? 22
  const maxC = opts.maxC ?? 40
  const points: DayPoint[] = []
  let wh = 0
  let peakW = 0
  for (let h = 0; h <= 24; h += step) {
    const irr = irradianceAt(h, opts.peakIrr ?? 1000)
    // ⚠️ الحرارة تتأخّر عن الشمس: أحرّ ساعة ~٣ عصراً مو الظهر.
    const ambientC = minC + (maxC - minC) * Math.max(0, Math.sin((Math.PI * (h - 5)) / 16))
    if (irr <= 0) { points.push({ hour: h, irradiance: 0, ambientC, cellT: ambientC, pmax: 0, vmp: 0 }); continue }
    const m = panelModel(spec, irr, ambientC)
    const mpp = findMpp(m, 120)
    const total = mpp.pmax * opts.panels
    wh += total * step
    peakW = Math.max(peakW, total)
    points.push({ hour: h, irradiance: irr, ambientC, cellT: m.cellT, pmax: total, vmp: mpp.vmp })
  }
  return { points, kwh: wh / 1000, peakW }
}
