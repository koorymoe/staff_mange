// ═══════════════════════════════════════════════════════════════════
// مشي الكيان **داخل الصفحة** — لا تمثال محطوط بزاوية
// ═══════════════════════════════════════════════════════════════════
//
// 🔴 **السبب حرفياً من مالك النظام**: «ما تكون طايفة، وإنما تگدر
// تمشي داخل الصفحات وع الأسطر مال النظام والواجهات… ما ريدها تمثال
// ومحطوط». وقبل هذا الملف، الكيان **صندوق `fixed` بزاوية الشاشة**
// يُسحب باليد ويتذكّر مكانه — يعني محطوط فعلاً، وشكواه دقيقة.
//
// ⚠️ **والمشي بتحريك الصندوق نفسه مو بكانفس بحجم الشاشة**:
// كانفس WebGL يغطّي الصفحة كلها يعني **إعادة رسم كل الشاشة كل إطار**
// على أجهزة الموظفين، ويسرق نقرات الماوس من الواجهة تحته. وتحريك
// صندوق ٩٦×١١٢ يخلي التصيير **بنفس كلفته الحالية بالضبط**.
//
// ⚠️ **واحترام الموظف قبل العرض**: الكيان ما يمشي وهو يكتب، ولا
// يوقف فوق زر، ويرجع **لمكانه الي حطّه المستخدم** لمّا يخلص. البوت
// الي يلعب فوق شغل الموظف يتحول لعقوبة مو مساعدة.

/** نقطة بالشاشة (بكسلات CSS، نسبة للنافذة). */
export interface Spot { x: number; y: number }

/** حجم صندوق الكيان — نفس الي بـ`EntityCompanion`. */
export const ENTITY_W = 96
export const ENTITY_H = 112

/**
 * سرعة المشي **بالبكسل/ثانية**.
 *
 * ⚠️ الرقم مقصود: أسرع منه يبين «يطير» ويشتّت، وأبطأ منه يخلي
 * الموظف ينتظره. و٤٢٠ بكسل/ث تعني عبور شاشة ١٤٤٠ بـ٣.٤ ثانية.
 */
const SPEED_PX_S = 420

/** أقصر مسافة تستاهل مشياً — أقل منها نقفز بهدوء. */
const MIN_WALK_PX = 40

/**
 * هل الموظف يكتب الآن؟
 *
 * ⚠️ **يُفحص قبل كل مشية**: لو مشى وهو يكتب بحقل، الحركة تسحب عينه
 * من السطر الي يكتبه — وهذا إزعاج مو مساعدة.
 */
export function userIsTyping(): boolean {
  const el = document.activeElement as HTMLElement | null
  if (!el) return false
  const tag = el.tagName
  return tag === 'INPUT' || tag === 'TEXTAREA' || tag === 'SELECT' || el.isContentEditable
}

/** يحترم إعداد «قلّل الحركة» بالجهاز — الحركة تنطفي كلياً. */
export function reducedMotion(): boolean {
  try {
    return window.matchMedia('(prefers-reduced-motion: reduce)').matches
  } catch {
    return false
  }
}

/** يقيّد النقطة داخل النافذة حتى الكيان ما يطلع برّا الشاشة. */
export function clampToViewport(p: Spot): Spot {
  return {
    x: Math.min(Math.max(8, p.x), Math.max(8, window.innerWidth - ENTITY_W - 8)),
    y: Math.min(Math.max(8, p.y), Math.max(8, window.innerHeight - ENTITY_H - 8)),
  }
}

/**
 * يلگى العنصر الي يقصده سطر التقرير، من **رابطه**.
 *
 * ⚠️ **بالرابط مو بالنص**: نصوص الأسطر تتغيّر مع البيانات (أرقام
 * وأسماء)، والرابط ثابت وهو نفس الي يفتحه الزر. ونجرّب الشكلين
 * (`/x` و`#/x`) لأن الراوتر يكتبها بشكلين حسب المكان.
 */
export function findLinkTarget(link: string): HTMLElement | null {
  const clean = link.replace(/^#/, '')
  const sels = [
    `a[href$="${clean}"]`,
    `a[href$="#${clean}"]`,
    `[data-entity-target="${clean}"]`,
  ]
  for (const s of sels) {
    const el = document.querySelector<HTMLElement>(s)
    // ⚠️ العنصر المخفي ما ينفع هدفاً: الكيان يمشي لمكان فاضي
    // ويشاور على لا شي. `offsetParent === null` تكشف المخفي.
    if (el && el.offsetParent !== null) return el
  }
  return null
}

/**
 * كم عنصراً تفاعلياً يغطّيه صندوق الكيان لو وقف بهالنقطة؟
 *
 * ⚠️ **يُفحَص الصندوق كله مو المركز**: الكيان **١١٢ بكسل طولاً**
 * وصفوف القائمة الجانبية **٤٠** — فمركزه يگدر يكون بمكان فاضي
 * وطرفاه فوق زرّين.
 *
 * ⚠️ **ودرس قياس**: ظنّيت إن الكيان يغطّي «التقييم» و«العمل»، وطلع
 * **قياسي هو الغلط**: كنت أقيس `getBoundingClientRect` للغلاف
 * الخارجي — وهو يحتوي **فقاعة الكلام** (لعرض ١٦rem) فصندوقه أوسع
 * هواي من الشخصية. والقياس من داخل التطبيق على صندوق ٩٦×١١٢ رجّع
 * **صفر عناصر تفاعلية** بالنقاط الخمس كلها. فالفحص هنا يأخذ
 * **نقطة الوقوف** مو غلاف المكوّن.
 */
export function interactiveHits(p: Spot): number {
  const pts: [number, number][] = [
    [p.x + ENTITY_W / 2, p.y + ENTITY_H / 2],
    [p.x + 6, p.y + 6],
    [p.x + ENTITY_W - 6, p.y + 6],
    [p.x + 6, p.y + ENTITY_H - 6],
    [p.x + ENTITY_W - 6, p.y + ENTITY_H - 6],
  ]
  const seen = new Set<Element>()
  for (const [x, y] of pts) {
    if (x < 0 || y < 0 || x > window.innerWidth || y > window.innerHeight) continue
    const stack = document.elementsFromPoint(x, y) as HTMLElement[]
    // ⚠️ نتجاهل عقد الكيان نفسه، وإلا يحسب نفسه «عنصراً يغطّيه».
    const top = stack.find((el) => !el.closest('[data-entity-root]'))
    const act = top?.closest('button, a, input, textarea, select, [role="button"]')
    if (act) seen.add(act)
  }
  return seen.size
}

/**
 * يختار **مكان وقوف فاضي** قريباً من العنصر.
 *
 * 🔴 **التغطية عيب حقيقي مو تفصيلاً**: لو وقف فوق زر، الموظف ما
 * يگدر يضغطه — فالمساعدة تصير مانعاً. فنولّد مرشّحات حول العنصر
 * ونأخذ **أول واحد ما يغطّي ولا عنصر**، وإذا كلها مشغولة نأخذ
 * **الأقل تغطية** — «أحسن موجود» بدل ما نلغي المشي كله.
 *
 * والترتيب مقصود: الجهة (يسار أولاً لأن الواجهة عربية) ثم إزاحة
 * عمودية تدريجية — أقرب مكان مقبول يفوز.
 */
export function pickStandSpot(rect: DOMRect, el?: HTMLElement | null): Spot {
  const gap = 12
  // 🔴 **أولاً: خارج حدود الحاوية الي فيها الهدف** — وهاي قاعدة
  // حاسمة بدل فحص «شنو تحتي»، والسبب مقاس:
  // الهدف كان صفّاً بالقائمة الجانبية، والفحص وقت الاختيار رجّع
  // **صفر تغطية** لأن القائمة **لسه ما انرسمت** (تعتمد على بيانات
  // تجي بنداء ثانٍ) — ولمّا وصل، وقف وغطّى **ثلاثة** عناصر:
  // «التقييم» و«جدول دوامي» و«العمل». يعني الفحص اللحظي يكذب
  // بسبب الترتيب الزمني، وحدود الحاوية **معلومة وثابتة**.
  const host = el?.closest('nav, aside, header, [role="navigation"]') as HTMLElement | null
  if (host) {
    const hr = host.getBoundingClientRect()
    // نوقف بالجهة الي فيها مساحة الصفحة: يسار الحاوية لو هي بأقصى
    // اليمين (واجهتنا عربية فالقائمة يمين)، وإلا يمينها.
    const outLeft = hr.left - ENTITY_W - gap
    const outRight = hr.right + gap
    const x = outLeft >= 8 ? outLeft : outRight
    return clampToViewport({ x, y: rect.top + rect.height / 2 - ENTITY_H / 2 })
  }
  const sides = [rect.left - ENTITY_W - gap, rect.right + gap]
  const midY = rect.top + rect.height / 2 - ENTITY_H / 2
  const dys = [0, -ENTITY_H * 0.6, ENTITY_H * 0.6, -ENTITY_H * 1.2, ENTITY_H * 1.2]
  let best: Spot | null = null
  let bestHits = Infinity
  const consider = (cand: Spot): Spot | null => {
    const hits = interactiveHits(cand)
    if (hits === 0) return cand
    if (hits < bestHits) { bestHits = hits; best = cand }
    return null
  }
  // ① جنب العنصر مباشرةً — الأقرب والأوضح
  for (const dy of dys) {
    for (const x of sides) {
      const ok = consider(clampToViewport({ x, y: midY + dy }))
      if (ok) return ok
    }
  }
  // ② 🔴 **مسح أفقي للخارج** — والسبب مقاس: العنصر الهدف كان داخل
  // **القائمة الجانبية**، وهي عمود صفوف ٤٠ بكسل متلاصقة. فأي صندوق
  // بطول ١١٢ جنبها يغطّي صفّين–ثلاثة، و**كل** المرشّحات القريبة
  // طلعت مشغولة (قِستها: وقف وغطّى «التقييم» و«العمل»).
  // فنمشي للخارج بخطوات لين نطلع من العمود المزدحم لمساحة الصفحة.
  for (let step = 1; step <= 12; step++) {
    for (const dy of [0, -ENTITY_H * 0.8, ENTITY_H * 0.8]) {
      const ok = consider(clampToViewport({
        x: rect.left - ENTITY_W - gap - step * 56,
        y: midY + dy,
      }))
      if (ok) return ok
    }
  }
  // ③ ولو كل الشاشة مشغولة، نأخذ **الأقل تغطية** — أحسن موجود بدل
  // ما نلغي المشي كله ونرجع لسلوك «التمثال المحطوط».
  return best ?? clampToViewport({ x: sides[0], y: midY })
}

/**
 * يمشّي الكيان من نقطة لنقطة بسرعة بشرية، ويرجّع **دالة إلغاء**.
 *
 * ⚠️ **الحركة بـ`requestAnimationFrame` مو بانتقال CSS**: المكوّن
 * يحتاج يعرف **الموضع الحقيقي كل إطار** (حتى يقلب وجهه ويوقف مقطع
 * المشي بالوصول)، وانتقال CSS يخفي الموضع الوسطي عن الكود.
 *
 * ⚠️ و**الزمن من `performance.now()` مو من عدد الإطارات**: بجهاز
 * بطيء أو تبويب بالخلفية عدد الإطارات ينزل، فالمشي المربوط بالإطارات
 * يصير بطيئاً بشكل عشوائي. والمربوط بالوقت يمشي بنفس السرعة دائماً.
 */
export function walk(
  from: Spot,
  to: Spot,
  onStep: (p: Spot, dir: 1 | -1) => void,
  onDone: () => void,
): () => void {
  const target = clampToViewport(to)
  const dx = target.x - from.x
  const dy = target.y - from.y
  const dist = Math.hypot(dx, dy)
  if (dist < MIN_WALK_PX || reducedMotion()) {
    onStep(target, dx < 0 ? -1 : 1)
    onDone()
    return () => {}
  }
  const ms = (dist / SPEED_PX_S) * 1000
  const dir: 1 | -1 = dx < 0 ? -1 : 1
  const t0 = performance.now()
  let raf = 0
  let cancelled = false
  const tick = () => {
    if (cancelled) return
    const t = Math.min(1, (performance.now() - t0) / ms)
    // تسهيل بسيط بالبداية والنهاية — المشي الخطّي يبين آلياً
    const e = t < 0.5 ? 2 * t * t : 1 - Math.pow(-2 * t + 2, 2) / 2
    onStep({ x: from.x + dx * e, y: from.y + dy * e }, dir)
    if (t >= 1) { onDone(); return }
    raf = requestAnimationFrame(tick)
  }
  raf = requestAnimationFrame(tick)
  return () => { cancelled = true; cancelAnimationFrame(raf) }
}
