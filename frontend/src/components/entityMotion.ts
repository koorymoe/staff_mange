// ═══════════════════════════════════════════════════════════════════
// محرّك الحركة المحسوبة — الكائن يسوّي أفعالاً ما انسجّلت
// ═══════════════════════════════════════════════════════════════════
//
// طلب (ع): «مانريد ننطي هاي الشخصية كم حركة وتضل تعيدهن… نبرمجه انو
// يعرف كلشي… يعرف ياشر بايده ويوجه الموظف».
//
// ⚠️ **والفرق التقني الي يخلي الطلب ممكناً**:
//   الحركة **المسجّلة** سقفها عدد المقاطع — فما تگدر تشاور على زر ما
//   شفناه قبل، ولا تمشي لشاشة جديدة. والحركة **المحسوبة** بلا سقف:
//   تعطيها هدفاً ← الكود **يحسب** زوايا العظام.
//
// ⚠️ **وثلاثة قيود مقاسة على الشخصية الحالية — ما نتجاوزها بالتمنّي**:
//   ① **ماكو عظام عيون** (قستها: صفر) — فتتبّع المؤشر يصير **بالرأس**،
//      **واسمه «توجّه رأس» مو «تتبّع عيون»** (اتفاق صريح مع (م)).
//   ② **الوسطى والبنصر والخنصر بلا عظام** — فالإمساك واللعب بالطوبة
//      **مستحيلان** لحد إعادة الريغ. **والسبّابة موجودة (٤ عظام)،
//      فالإشارة ممكنة اليوم.**
//   ③ **المقاطع فيها إزاحة جذر** (٤١ مسار على `Hips`) — فالمشي
//      البرمجي فوقها **يضاعف السير** (تحذير (م)، وقسته وطلع صح).
//      لهذا `neutralizeRootMotion` **إلزامية** قبل أي تحريك برمجي.
import type { Scene } from '@babylonjs/core/scene'
import type { Skeleton } from '@babylonjs/core/Bones/skeleton'
import type { AbstractMesh } from '@babylonjs/core/Meshes/abstractMesh'
import { TransformNode } from '@babylonjs/core/Meshes/transformNode'
import { Matrix, Quaternion, Vector3 } from '@babylonjs/core/Maths/math.vector'
import { AnimationGroupMask, AnimationGroupMaskMode } from '@babylonjs/core/Animations/animationGroupMask'
import type { AnimationGroup } from '@babylonjs/core/Animations/animationGroup'
import { BoneIKController } from '@babylonjs/core/Bones/boneIKController'
import { BoneLookController } from '@babylonjs/core/Bones/boneLookController'
import { Space } from '@babylonjs/core/Maths/math.axis'

import { detectNaming, FINGER_KEYS, type BoneNaming, type FingerKey } from './skeletonNaming'

/**
 * أقصى انثناء لمفصل إصبع واحد (راديان ≈ ٧٧°).
 * ⚠️ ثلاث مفاصل × ٧٧° تنطي قبضة كاملة. رفعها أعلى يخلي الإصبع
 * **يدخل بالكف** — والتصادم ما ينحسب بالـrig، فالحد هنا وقاية.
 */
const FINGER_MAX_BEND = 1.35

export interface MotionRig {
  /**
   * يشاور بالإصبع نحو نقطة بالعالم — أي نقطة، بلا مقطع مسجّل.
   *
   * ⚠️ **والهدف يُقصَر على مدى الذراع مع الحفاظ على الاتجاه**: الزر
   * الي بالشريط الجانبي يبعد **متراً ونصف** عن شخصية طولها ١.٧٠ م،
   * فـ«لمسه» مستحيل تشريحياً. الي نريده هو **اتجاه** الإشارة صحيحاً
   * — لهذا القياس المعتمد **زاوية مو مسافة** (`pointingErrorDeg`).
   */
  pointAt(target: Vector3, hand?: 'right' | 'left'): void
  /**
   * يشاور نحو نقطة **بإحداثيات بكسل نسبة للكانفس** — وتنطلع برّا
   * الكانفس عادي (زر بالشريط الجانبي).
   *
   * ⚠️ **الإسقاط العكسي من الكاميرا هو الصحيح**: حسبتها أول مرة
   * بمحاور مخمّنة و«متر لكل بكسل» ثابت، فطلع **خطأ الاتجاه ١٥٢°**
   * (الذراع تشاور بالعكس) — لأن الهدف نزل ورا الشخصية. والكاميرا
   * هي الي تعرف وين «يمين الشاشة» بالعالم، مو أنا.
   */
  pointAtScreen(px: number, py: number, hand?: 'right' | 'left'): Vector3 | null
  /** يوقف الإشارة ويرجّع الذراع لوضعها الطبيعي تدريجياً. */
  stopPointing(): void
  /** الرأس يتوجّه لنقطة — ⚠️ **توجّه رأس**، مو تتبّع عيون. */
  lookAt(target: Vector3 | null): void
  /**
   * **العيون** تتوجّه لنقطة بالعالم — تتبّع عيون حقيقي، مو لفّة رأس.
   *
   * 🔴 **وهذا الي طلبه مالك النظام**: «يتبع الموشر مال الحاسبة».
   * وشكواه «أكو مشكلة بالعيون» كانت على مجسّم **عيونه مرسومة على
   * الصورة** (صفر عظمة عين) — فما گدر يتبع ولا يرمش أبداً.
   *
   * `null` يوقف التتبّع ويرجّع النظر لوضعه الطبيعي.
   * ⚠️ ويرجّع `false` لو الهيكل **ما عنده عظام عيون** — والمستدعي
   * يعرف إن الميزة مو متوفرة بهذا المجسّم بدل ما يظن إنها شغّالة.
   */
  gazeAt(target: Vector3 | null): boolean
  /** نفس `gazeAt` بس بإحداثيات بكسل نسبة للكانفس (موضع المؤشر). */
  gazeAtScreen(px: number, py: number): boolean
  /** هل هذا المجسّم عنده عظام عيون أصلاً؟ — للعرض الصادق بالمختبر. */
  hasEyeBones(): boolean
  /** يمشي للهدف بسرعة متر/ثانية — يرجّع وعداً ينتهي بالوصول. */
  walkTo(target: Vector3, speed?: number): Promise<void>
  /**
   * **خطأ الإشارة بالدرجات**: الزاوية بين (الكتف ← طرف السبّابة)
   * و(الكتف ← الهدف الحقيقي). هذا المقياس الصحيح للإشارة، لأن
   * المسافة تعاقب الكائن على شي مستحيل (هدف خارج مدى ذراعه).
   * `null` لو عظام السبّابة أو الكتف مو موجودة.
   */
  pointingErrorDeg(target: Vector3, hand?: 'right' | 'left'): number | null
  /**
   * **خطأ الإشارة بفضاء الشاشة** بالدرجات: الزاوية بين (الكتف ←
   * طرف السبّابة) و(الكتف ← الزر) **بعد إسقاطهم للشاشة**.
   *
   * ⚠️ **وهذا المقياس المعتمد**، لأن الي يشوفه المستخدم ثنائي
   * الأبعاد: يد ممدودة بعمق مختلف تنطي زاوية عالمية كبيرة وهي
   * بالعين مضبوطة، والعكس صحيح. والخطأ العالمي يضلّل هنا.
   */
  pointingScreenErrorDeg(px: number, py: number, hand?: 'right' | 'left'): number | null
  /** مدى الذراع المقاس بالأمتار (كتف ← يد) — `null` لو ماكو عظام. */
  armReach(hand?: 'right' | 'left'): number | null
  /** موقع طرف السبّابة بالعالم — للفحص المباشر. */
  fingertipWorld(hand?: 'right' | 'left'): { x: number; y: number; z: number } | null
  /** موقع الكتف بالعالم — للفحص المباشر. */
  shoulderWorld(hand?: 'right' | 'left'): { x: number; y: number; z: number } | null
  /**
   * يمشي مسافة نسبية بالأمتار — **مدخله أرقام بسيطة بقصد**، حتى
   * الشاشات ما تستورد `Vector3` من بابل فتدخل الحزمة الأولية
   * (قياس: البارل ١.٢ م.ب، والحزمة الأولية لازم تبقى ١١٧ ك.ب).
   */
  walkBy(dx: number, dz: number, speed?: number): Promise<void>
  /** موقع الجذر — لقياس مسافة السير (هل تتضاعف ولا لا). */
  rootPosition(): { x: number; y: number; z: number }
  /**
   * يوجّه الرأس بزوايا مباشرة (راديان) — مصدرها كاميرا الجهاز.
   *
   * ⚠️ **وهذا مسار منفصل عن `lookAt`**: `lookAt` يحتاج نقطة بالعالم،
   * والكاميرا تنطي **زوايا** جاهزة. وتحويل الزوايا لنقطة وهمية ثم
   * رجوعها لزوايا يضيف خطأ بلا فايدة. `null` يرجّع الرأس للمقطع.
   */
  setHeadPose(pose: { yaw: number; pitch: number; roll: number } | null): void
  /**
   * يقبض الأصابع: `0` ممدود و`1` مقبوض — مصدرها كاميرا الجهاز.
   *
   * 🔴 **غير مكتمل — ولا يُعتمد عليه بعد.** القياس (مو الانطباع):
   * طرف السبّابة تحرّك **٠ مم** ومفصلها الوسطي **٠ مم** والإبهام
   * **٠ مم** بين قبضة كاملة ويد مفتوحة، وتغيّر الصورة **٠٪**.
   * السبب نفس سبب الرأس: **الكتابة اليدوية على العظام داخل حلقة
   * التحديث ما تُرفع للرسم** لمن ما يكون اكو مقطع شغّال. والرأس
   * انحلّ لأن بابل عنده `BoneLookController` جاهز، وماكو مقابل
   * للأصابع — فتحتاج حلاً مستقلاً.
   *
   * ⚠️ **وهي محجوبة أصلاً بالمجسّم**: الوسطى والبنصر والخنصر
   * **بلا عظام** (`fingerBoneReport()` يطلّع الأرقام)، فالقبضة
   * الكاملة مستحيلة على هذا المجسّم مهما انصلّح الكود. فالترتيب
   * الصحيح: **المجسّم الجديد أول**، وبعده هاي.
   */
  setFingerCurl(hand: 'right' | 'left', curls: Partial<Record<string, number>>): void
  /**
   * **يقيس** عظام أصابع كل يد — الأداة الي تحكم على أي مجسّم:
   * الحالي، والمجسّم الجديد لمن يجي. رقم مو انطباع.
   */
  fingerBoneReport(): Record<string, number>
  /** أسماء التعابير المتوفرة بالمجسّم — فاضية يعني ماكو تعابير. */
  expressionNames(): string[]
  /** يطبّق وزن تعبير — يُهمَل لو التعبير مو موجود بالمجسّم. */
  setExpression(name: string, weight: number): void
  dispose(): void
}

/**
 * يحيّد إزاحة الجذر من مقطع: يخلّي المشي **بالمكان**، فالتحريك
 * البرمجي هو الي ينقل الجسم — بلا مضاعفة.
 *
 * ⚠️ **مقاس**: مقاطع المشي والركض فيها **٤١ مسار إزاحة** على
 * `Hips`. وبلا هذا التحييد السير يتضاعف حرفياً.
 *
 * يرجّع دالة تُرجع الإزاحة لو احتجناها (مثلاً لمقطع يُعرض لحاله).
 */
export function neutralizeRootMotion(
  skeleton: Skeleton,
  carrier: TransformNode,
  mesh: AbstractMesh,
): () => void {
  const naming = detectNaming(skeleton.bones.map((b) => b.name))
  const hips = skeleton.bones.find((b) => b.name === naming.hips)
  if (!hips) return () => {}

  // ⚠️ **التحييد يصير على عقدة حاملة، مو بالكتابة على عظمة الحوض.**
  //
  // جرّبت الكتابة على العظمة بطريقتين وفشلتا **بالقياس**:
  //   ① مرجع من الوضع الحالي وقت البناء ← نقطة عشوائية على مسار
  //      المشي، تختلف كل تحميل: قمة الرأس طلعت z = ٠.٥١ و٠.٦٢
  //      و٠.٧٧ و**١.٨٨** بأربع تحميلات.
  //   ② مرجع من `getRestMatrix()` ثم `getBindMatrix()` ← الحوض
  //      طلع عند **z ≈ ٢.٢٣ م** والرأس عند **−٠.١٣ م**، أي حوض
  //      يبعد مترين عن راسه — إحداثيات العظام هنا **بفضاء مقيّس**
  //      (ميكسامو يجي بالسنتيمترات مع مقياس على الجذر)، فالأرقام
  //      الي أكتبها بالأمتار تخرّب الهيكل. ومعها انزاحت الشخصية
  //      **١.٢٧ م** وطلعت بحاشية الإطار.
  //
  // فالحل ما يكتب على العظام إطلاقاً: نقيس **إزاحة الحوض بالعالم**
  // (وهاي وحدات المشهد، معروفة) ونطرحها من عقدة حاملة. فالمقطع
  // يمشي بحرّيته، والجسم يبقى بمكانه — ولا نحتاج نعرف فضاء العظام.
  let baseline: { x: number; z: number } | null = null
  const observer = () => {
    skeleton.computeAbsoluteMatrices(true)
    mesh.computeWorldMatrix(true)
    const w = hips.getAbsolutePosition(mesh)
    if (!baseline) {
      // أول إطار **بعد** تشغيل المقطع = إطار المقطع صفر دائماً،
      // فالمرجع ثابت بين التحميلات (وهذا الي كان ناقصاً).
      baseline = { x: w.x, z: w.z }
      return
    }
    // تصحيح مباشر: إزاحة العقدة تنعكس ١:١ على العالم، فخطوة واحدة
    // تكفي ولا نحتاج حلقة تقارب.
    carrier.position.x -= w.x - baseline.x
    carrier.position.z -= w.z - baseline.z
  }
  const scene = skeleton.getScene()
  scene.onAfterAnimationsObservable.add(observer)
  return () => {
    scene.onAfterAnimationsObservable.removeCallback(observer)
    carrier.position.set(0, 0, 0)
  }
}

/**
 * يبني منظومة الحركة المحسوبة فوق شخصية محمّلة.
 *
 * @param root العقدة الي تنقل الجسم كله (التحريك البرمجي يصير عليها)
 */
export function buildMotionRig(
  scene: Scene,
  skeleton: Skeleton,
  mesh: AbstractMesh,
  root: TransformNode,
  groups: AnimationGroup[] = [],
): MotionRig {
  const skinnedMesh = mesh
  const bone = (n: string) => skeleton.bones.find((b) => b.name === n) ?? null
  const firstBone = (names: readonly string[]) => {
    for (const n of names) { const b = bone(n); if (b) return b }
    return null
  }
  // نمط الأسماء يُكتشَف من الهيكل نفسه — انظر `skeletonNaming.ts`.
  const naming: BoneNaming = detectNaming(skeleton.bones.map((b) => b.name))
  const fingerAxis = new Vector3(...naming.fingerBendAxis)

  // ═══ أهداف الحركة: عُقد غير مرئية يتبعها الـIK ═══
  const ikTarget = new TransformNode('entity-ik-target', scene)
  // القطب يحدد **اتجاه المرفق** — بلاه المرفق يلتوي لجهة غلط
  // والإشارة تطلع مكسورة، وهاي مو عيب بالـIK بل قطب ناقص.
  const poleTarget = new TransformNode('entity-ik-pole', scene)
  const lookTarget = new TransformNode('entity-look-target', scene)

  const ik: Record<'right' | 'left', BoneIKController | null> = { right: null, left: null }
  for (const side of ['right', 'left'] as const) {
    const fore = bone(naming.foreArm(side))
    if (!fore) continue
    ik[side] = new BoneIKController(mesh, fore, {
      targetMesh: ikTarget,
      poleTargetMesh: poleTarget,
      // ⚠️ حدّ الزاوية يمنع الذراع تنقلب للورا — الـIK بلا حد
      // يلگى «حلاً» رياضياً صحيحاً وتشريحياً مستحيلاً.
      // ⚠️ حدّ الزاوية **لا يوصل ١٨٠°**: قِستها — بـ٠.٩٨π الذراع
      // **تتجمّد تماماً** (طرف الإصبع ما يتحرّك ولا ملّيمتر بستة
      // قياسات متتالية)، لأن الهدف بمدى كامل يطلّع `acos` خارج
      // مجاله داخل الحلّال فتصير المصفوفة NaN. و٠.٩٢π تمدّ الذراع
      // بلا ما تكسر الحلّال.
      maxAngle: Math.PI * 0.92,
    })
    // ⚠️ **سلطة كاملة للـIK على الذراع** (`1`) مو مزجاً جزئياً.
    // قِستها: بمزج ٠.٢٨ والمقطع شغّال، الـIK يمزج كل إطار من قيمة
    // **متحركة** (المشي يهزّ الذراع) فما يتقارب أبداً — الخطأ يتذبذب
    // بين ٨° و١٩°. وبسلطة كاملة يتقارب لـ**٠.٣°**، والنعومة تجي من
    // تحريك **الهدف** تدريجياً مو من مزج العظمة.
    ik[side]!.slerpAmount = 1
  }

  const headBone = bone(naming.head)
  const look = headBone
    ? new BoneLookController(mesh, headBone, lookTarget.position, {
      // حدود بشرية: الرأس ما يلتف ٩٠° ولا يرفع لفوق بلا حد
      maxYaw: Math.PI * 0.42,
      maxPitch: Math.PI * 0.22,
      slerpAmount: 0.22,
    })
    : null

  // ═══ العيون ═══
  //
  // ⚠️ **بـ`BoneLookController` مثل الرأس** — وهذا **المسار الوحيد
  // المقاس إنه يوصل للرسم**: الكتابة اليدوية على العظام رجّعت **صفر
  // مم وصفر تغيّر بالصورة** بكل محاولة (نفس علّة قبض الأصابع).
  //
  // ⚠️ و**حدود العين أضيق هواي من الرأس**: العين تلف ~٣٠° أفقياً
  // و~٢٠° عمودياً وبعدها **يلتفت الرأس**. وبلا هالحدود تطلع نظرة
  // «الشيطان» — العين تلف لآخر المحجر وتبين مرعبة مو منتبهة.
  //
  // ⚠️ ويُبنى **بس لو الهيكل عنده عظام عيون**: ميكسامو ما ينطي
  // عظام عيون إطلاقاً (`naming.eye()` يرجّع `null`)، فالميزة
  // **تسكت بهدوء** بدل ما تكسر.
  const eyeLook: Record<'right' | 'left', BoneLookController | null> = { right: null, left: null }
  const eyeTarget = new TransformNode('entity-eye-target', scene)
  for (const side of ['right', 'left'] as const) {
    const name = naming.eye(side)
    const eyeBone = name ? bone(name) : null
    if (!eyeBone) continue
    eyeLook[side] = new BoneLookController(mesh, eyeBone, eyeTarget.position, {
      maxYaw: Math.PI * 0.17,
      maxPitch: Math.PI * 0.11,
      // ⚠️ أسرع من الرأس (٠.٢٢): العين تسبق الرأس عند البشر، ولو
      // تحرّكت بنفس بطئه تبين كأنها زجاجية.
      slerpAmount: 0.35,
    })
  }
  let gazing = false

  // ⚠️ **الإحداثيات لازم تمرّ بالمجسم**: `getAbsolutePosition()` بلا
  // تمرير المجسم ترجّع فضاء الهيكل مو العالم — وهاي غلطة قِستها:
  // طلعت أخطاء الإشارة **١٨٦–٢٢٣ سم** وهي أصلاً غلط قياس مو غلط IK.
  const worldOf = (b: ReturnType<typeof bone>) =>
    b ? b.getAbsolutePosition(skinnedMesh) : null
  const fingertip = (hand: 'right' | 'left') => {
    // آخر عظمة بالسلسلة هي **طرف** الإصبع بميكسامو (`...Index4`)؛
    // القياس من `Index1` يخلي الإشارة تبين أدق من واقعها.
    return worldOf(firstBone(naming.fingerTip(hand)))
  }
  const shoulder = (hand: 'right' | 'left') =>
    worldOf(bone(naming.arm(hand)))
  /** مدى الذراع مقاس من العظام نفسها — مو ثابتاً مكتوباً. */
  const reachOf = (hand: 'right' | 'left'): number | null => {
    const a = worldOf(bone(naming.arm(hand)))
    const b = worldOf(bone(naming.foreArm(hand)))
    const c = worldOf(bone(naming.hand(hand)))
    if (!a || !b || !c) return null
    return Vector3.Distance(a, b) + Vector3.Distance(b, c)
  }

  /**
   * ⚠️ **طبقات الحركة**: المقطع يحرّك الجسم، و**الـIK يملك الذراع**.
   * قِستها: بلا هذا الفصل، مقطع المشي يهزّ الذراع كل إطار فالخطأ
   * يتذبذب **٨°–١٩°** ولا يتقارب أبداً، مهما عايرت المزج — لأنه
   * تنازع مو معايرة. وبالمسك (استثناء سلسلة الذراع من المقطع)
   * الـIK يمسك الذراع وباقي الجسم يكمل مشيه.
   */
  const armChainNames = (hand: 'right' | 'left'): string[] => {
    const re = naming.armChain(hand)
    return skeleton.bones.map((b) => b.name).filter((n) => re.test(n))
  }
  const applyArmMask = (hand: 'right' | 'left' | null) => {
    for (const g of groups) {
      if (!hand) { g.mask = null; g.syncWithMask(true); continue }
      g.mask = new AnimationGroupMask(armChainNames(hand), AnimationGroupMaskMode.Exclude)
      g.syncWithMask(true)
    }
  }

  /**
   * القطب يُحسب **من الكتف** مو من الهدف: القطب المعلّق بالهدف يدور
   * معاه فالمرفق يلتوي لجهة غريبة. والمرفق الطبيعي يبقى **تحت**
   * الكتف وشوي للخارج والورا.
   */
  const polePointFor = (hand: 'right' | 'left', sh: Vector3) =>
    sh.add(new Vector3(hand === 'right' ? 0.45 : -0.45, -0.8, -0.25))

  // ═══ الأصابع: مقاسة مو مفترضة ═══
  // ⚠️ **ميكسامو يسمّي سلاسل الأصابع بنمط ثابت** (`RightHandIndex1..4`)،
  // **بس وجودها مو مضمون**: قِستها على شخصيتنا — الإبهام والسبّابة
  // موجودان، والوسطى والبنصر والخنصر **صفر عظام**. فالكود يشتغل على
  // الموجود ويسكت عن الناقص، و`fingerBoneReport()` يطلّع الحقيقة.
  /** عظام إصبع واحد بالترتيب من القاعدة — بلا عظمة الطرف (ما تنثني). */
  const fingerBones = (hand: 'right' | 'left', finger: FingerKey) => {
    const out = []
    for (const n of naming.fingerBones(hand, finger)) {
      const b = bone(n)
      if (b) out.push(b)
    }
    return out
  }
  /** الزاوية المطبَّقة فعلاً على كل عظمة إصبع — للدوران التفاضلي. */
  const fingerAngle = new Map<string, number>()

  /**
   * انحناء الأصابع المطلوب هذا الإطار. يُطبَّق بحلقة التحديث مو
   * فوراً، لأن المقطع يكتب على العظام كل إطار — فالتطبيق المباشر
   * **يُمحى** (نفس سبب ترتيب الـIK).
   */
  const curlWanted: Record<'right' | 'left', Record<string, number>> =
    { right: {}, left: {} }

  /**
   * التعابير: نبحث بكل الأجسام مو بواحد.
   *
   * ⚠️ **غلطة قِستها**: كان الكود يقرأ `skinnedMesh.morphTargetManager`
   * والـ`skinnedMesh` هو **أول** جسم مكسوّ يلگاه المحمّل — وبالمجسّم
   * الجديد طلع **الشعر** (٢٢ ألف رأس، بلا تعابير)، فرجّع «صفر
   * تعابير» والمجسّم فيه **٣١** على جسم اسمه `CC_Base_Body`. والوجه
   * ما يكون دايماً أول جسم، فالبحث لازم يشمل الكل.
   */
  const morphOwners = scene.meshes
    .filter((m) => m.morphTargetManager && m.morphTargetManager.numTargets > 0)
    .map((m) => m.morphTargetManager!)

  let pointing: 'right' | 'left' | null = null
  /** بكسلات الزر — التصحيح النهائي يصير عليها (فضاء الشاشة). */
  let pointPixel: { x: number; y: number } | null = null
  let aimDir: Vector3 | null = null
  let looking = false
  let walk: { to: Vector3; speed: number; done: () => void } | null = null

  // ⚠️ **بعد الحركات مو قبلها.** قِستها: نفس الكود بـ
  // `onBeforeRenderObservable` طلع خطأ إشارة **٩٨°** لمن يكون اكو
  // مقطع شغّال، و**١٣°** لمن أوقّف المقاطع — لأن المقطع يكتب فوق
  // زوايا العظام بعد ما الـIK يحسبها. والحل ترتيب مو معايرة.
  const observer = scene.onAfterAnimationsObservable.add(() => {
    const dt = scene.getEngine().getDeltaTime() / 1000
    if (pointing && ik[pointing]) {
      // ⚠️ **نحدّث المصفوفات قبل أي قياس**: العظام تحرّكت هذا الإطار
      // بالمقطع (الحوض والعمود)، ومصفوفاتها المطلقة تُحسب متأخرة —
      // فالقياس بلا هذا السطر يقرأ إطاراً قديماً، والخطأ يطلع بقدر
      // سرعة الحركة (تذبذب ٨°–٢٠° بالمشي، وصفر تقريباً بالوقوف).
      skeleton.computeAbsoluteMatrices(true)
      skinnedMesh.computeWorldMatrix(true)
      // ═══ تصحيح مغلق: الهدف هو **طرف الإصبع** مو الرسغ ═══
      // الـIK يوجّه الرسغ، والسبّابة تبقى بانحناءها — فيبقى فرق
      // زاوي ثابت. فنقيس الفرق كل إطار ونصحّح اتجاه الهدف خطوة
      // خطوة. بلا هذا يبقى خطأ ~١٣° وهو فرق يبيّن للعين.
      const sh = shoulder(pointing)
      const tip = fingertip(pointing)
      const reach = reachOf(pointing)
      if (sh && tip && reach && aimDir && pointPixel && scene.activeCamera) {
        // ═══ التصحيح المغلق **بفضاء الشاشة** ═══
        // نقيس وين طرف السبّابة طالع **بالبكسل** ووين الزر، وندوّر
        // اتجاه الإشارة بمقدار الفرق. والتصحيح بالشاشة مو بالعالم
        // لأن هذا الي تشوفه العين — واختلاف العمق يخرب الحساب
        // العالمي بينما الشاشة تحكم عليه مباشرة.
        const cam = scene.activeCamera
        const eng = scene.getEngine()
        const vp = cam.viewport.toGlobal(eng.getRenderWidth(), eng.getRenderHeight())
        const tm = scene.getTransformMatrix()
        const s2 = Vector3.Project(sh, Matrix.Identity(), tm, vp)
        const t2 = Vector3.Project(tip, Matrix.Identity(), tm, vp)
        const want = new Vector3(pointPixel.x - s2.x, pointPixel.y - s2.y, 0)
        const have = new Vector3(t2.x - s2.x, t2.y - s2.y, 0)
        if (want.length() > 1 && have.length() > 1) {
          // زاوية الفرق بالشاشة، ثم ندوّر اتجاه الهدف بنفس المقدار
          // حول محور نظر الكاميرا — فالنتيجة تنطبق بالعين.
          const a = Math.atan2(have.y, have.x)
          const b = Math.atan2(want.y, want.x)
          let diff = b - a
          while (diff > Math.PI) diff -= 2 * Math.PI
          while (diff < -Math.PI) diff += 2 * Math.PI
          // ⚠️ **مو `getForwardRay()`**: يحتاج استيراد `Ray` الجانبي،
          // وبلاه يرمي استثناءً **كل إطار** فتتجمّد الذراع تماماً
          // (قِستها: طرف الإصبع ثابت بستة قياسات، والسبب استثناء
          // «Ray needs to be imported» مو معايرة الحلّال). واستيراد
          // `Ray` يكبّر الحزمة بلا فائدة — واتجاه الكاميرا يكفي.
          const axis = cam.getDirection(Vector3.Forward())
          // ⚠️ y بالشاشة ينزل، فدوران الشاشة معاكس لدوران العالم
          // حول محور النظر — والإشارة تنقلب لو أهملنا هاي.
          // ⚠️ **تصحيح صغير للصقل فقط** (٠.٢٥): الحلقة بكسب عالي
          // تتباعد بالأهداف الجانبية — قِستها ٩٧°←١٢٠°. والهدف صار
          // صحيحاً أصلاً بقطع المستوى، فالحلقة تلطّف الباقي بس.
          const q = Quaternion.RotationAxis(axis, -diff * 0.25)
          aimDir = aimDir.applyRotationQuaternion(q).normalize()
          ikTarget.position.copyFrom(sh.add(aimDir.scale(reach * 0.97)))
          poleTarget.position.copyFrom(polePointFor(pointing, sh))
        }
      }
      ik[pointing]!.update()
    }
    if (looking && look) look.update()
    if (gazing) {
      for (const side of ['right', 'left'] as const) eyeLook[side]?.update()
    }
    // ═══ انحناء الأصابع من الكاميرا ═══
    let wroteFingers = false
    for (const hand of ['right', 'left'] as const) {
      const wanted = curlWanted[hand]
      for (const finger of FINGER_KEYS) {
        const v = wanted[finger]
        if (v === undefined) continue
        const bones = fingerBones(hand, finger)
        for (const b of bones) {
          // ⚠️ **مسار `rotate` مو `setRotationQuaternion`**: الثاني
          // (بفضاء محلي) ما يوصل للرسم — قِسته: صفر مم حركة وصفر
          // تغيّر بالصورة، وجرّبت معاه إجبار التركيب و`prepare` ولا
          // شي نفع. و`rotate` يكتب على المصفوفة مباشرة بمسار ثاني.
          // والدوران **تفاضلي**: نتذكر الزاوية المطبَّقة وندوّر
          // بالفرق بس، وإلا تتراكم كل إطار وينلوي الإصبع.
          const applied = fingerAngle.get(b.name) ?? 0
          const goal = v * FINGER_MAX_BEND
          const delta = goal - applied
          if (Math.abs(delta) > 1e-4) {
            b.rotate(fingerAxis, -delta, Space.LOCAL)
            fingerAngle.set(b.name, goal)
            wroteFingers = true
          }
        }
      }
    }
    // ⚠️ **دفع الهياكل للرسم إلزامي لمن ماكو مقطع شغّال.**
    //
    // قِستها من داخل المحرّك (مو بلقطة متصفح): لفّة الرأس وقبض
    // الأصابع غيّرتا **صفر بالمئة** من الإطار وقت ما المقاطع
    // موقوفة، **ونفس الكود** غيّر **١٩.٢٪** لمن يكون اكو مقطع
    // شغّال. والسبب إن نظام الحركات هو الي يعلّم الهيكل «متسّخ»
    // كل إطار فتُرفع مصفوفات العظام للكارت؛ وبلا مقطع، كتابتنا
    // تضلّ بالذاكرة **وما تُرفع** — فالعظمة تتحرّك بالحساب
    // والصورة ما تتغيّر. و`prepare(true)` يتجاهل فحص الإطار
    // ويرفعها. وبالإنتاج دايماً اكو مقطع سكون شغّال، فهاي تحمينا
    // من حالة ما تبيّن إلا بالتجربة.
    if (wroteFingers) {
      // نحدّث المصفوفات **المطلقة** من المحلية الي كتبناها هالآن،
      // ثم ندفعها للرسم. الترتيب مهم: `prepare` وحده يرفع مصفوفات
      // **مخزّنة** من تمرير الحركات، فكتابتنا ما تبان.
      skeleton.computeAbsoluteMatrices(true)
      skinnedMesh.computeWorldMatrix(true)
      skeleton.prepare(true)
    }
    if (walk) {
      const dir = walk.to.subtract(root.position)
      dir.y = 0
      const dist = dir.length()
      if (dist < 0.05) {
        const finish = walk.done
        walk = null
        finish()
      } else {
        dir.normalize()
        root.position.addInPlace(dir.scale(Math.min(walk.speed * dt, dist)))
        // يستقبل جهة سيره — بلاها يمشي جانبياً مثل السلطعون
        root.rotation.y = Math.atan2(dir.x, dir.z)
      }
    }
  })

  return {
    pointAt(target, hand = 'right') {
      // نقصّر الهدف لمدى الذراع **بنفس الاتجاه**: الـIK بهدف بعيد
      // يشدّ الذراع لأقصاها ويطلع وضعاً متيبّساً، والاتجاه يضيع.
      const sh = shoulder(hand)
      const reach = reachOf(hand)
      let aim = target
      if (sh && reach) {
        const d = target.subtract(sh)
        const len = d.length()
        aimDir = d.scale(1 / Math.max(len, 1e-4))
        // ⚠️ **هامش أمان ٪٣ إلزامي**: الهدف على المدى الكامل بالضبط
        // يخرّب الحلّال (NaN) — وهامش صغير يمدّ الذراع بلا كسر.
        aim = sh.add(d.scale((reach * 0.97) / Math.max(len, 1e-4)))
      }
      ikTarget.position.copyFrom(aim)
      if (pointing !== hand) applyArmMask(hand)
      // القطب يُحسب: تحت الهدف وجانبه، فالمرفق ينزل طبيعياً
      if (sh) poleTarget.position.copyFrom(polePointFor(hand, sh))
      pointing = hand
    },
    pointAtScreen(px, py, hand = 'right') {
      const cam = scene.activeCamera
      const sh = shoulder(hand)
      if (!cam || !sh) return null
      const eng = scene.getEngine()
      const w = eng.getRenderWidth(), h = eng.getRenderHeight()
      const view = cam.getViewMatrix(), proj = cam.getProjectionMatrix()
      // ⚠️ **الطريقة الي طلعت مستقرة**: نبني شعاعاً من الكاميرا يمرّ
      // ببكسل الزر، ونقطعه بمستوى يمرّ **بالكتف** وعمودي على نظر
      // الكاميرا. فبكسلات الشاشة تتوزّع على مستوى بعمق الجسم نفسه،
      // والاتجاه يصير صحيحاً للأعلى وللجانبين وللأسفل.
      //
      // ⚠️ وجرّبت قبلها طريقتين وفشلتا بالقياس:
      //   ① إسقاط عكسي بعمق ثابت ٠.٥ ← كل البكسلات تتحوّل لمنطقة
      //      وحدة، فالوقفة ما تتغيّر أبداً لأي زر.
      //   ② حلقة تصحيح تدوّر الاتجاه حول محور النظر ← تستقر للأعلى
      //      (٢°) و**تتباعد** للجانبين والأسفل (٩٧°←١٢٠°).
      const near = Vector3.Unproject(new Vector3(px, py, 0), w, h, Matrix.Identity(), view, proj)
      const far = Vector3.Unproject(new Vector3(px, py, 1), w, h, Matrix.Identity(), view, proj)
      const dir = far.subtract(near)
      const n = cam.getDirection(Vector3.Forward())
      const denom = Vector3.Dot(dir, n)
      // شعاع موازي للمستوى (نظرياً مستحيل بكاميرا منظورية) ← نرجّع
      // بلا إشارة بدل ما نقسّم على صفر ونطلّع إحداثيات لا نهائية.
      if (Math.abs(denom) < 1e-6) return null
      const t = Vector3.Dot(sh.subtract(near), n) / denom
      const world = near.add(dir.scale(t))
      this.pointAt(world, hand)
      this.lookAt(world)
      pointPixel = { x: px, y: py }
      return world
    },
    stopPointing() {
      pointing = null; aimDir = null; pointPixel = null
      applyArmMask(null)          // ترجع الذراع للمقطع
    },
    pointingErrorDeg(target, hand = 'right') {
      const tip = fingertip(hand)
      const sh = shoulder(hand)
      if (!tip || !sh) return null
      const a = tip.subtract(sh)
      const b = target.subtract(sh)
      if (a.length() < 1e-4 || b.length() < 1e-4) return null
      const cos = Vector3.Dot(a.normalize(), b.normalize())
      return (Math.acos(Math.min(1, Math.max(-1, cos))) * 180) / Math.PI
    },
    pointingScreenErrorDeg(px, py, hand = 'right') {
      const cam = scene.activeCamera
      const sh = shoulder(hand)
      const tip = fingertip(hand)
      if (!cam || !sh || !tip) return null
      const eng = scene.getEngine()
      const w = eng.getRenderWidth(), h = eng.getRenderHeight()
      const vp = cam.viewport.toGlobal(w, h)
      const proj = (v: Vector3) =>
        Vector3.Project(v, Matrix.Identity(), scene.getTransformMatrix(), vp)
      const s2 = proj(sh), t2 = proj(tip)
      const ax = t2.x - s2.x, ay = t2.y - s2.y
      const bx = px - s2.x, by = py - s2.y
      const la = Math.hypot(ax, ay), lb = Math.hypot(bx, by)
      if (la < 1 || lb < 1) return null
      const cos = (ax * bx + ay * by) / (la * lb)
      return (Math.acos(Math.min(1, Math.max(-1, cos))) * 180) / Math.PI
    },
    armReach: (hand = 'right') => reachOf(hand),
    fingertipWorld: (hand = 'right') => {
      const v = fingertip(hand)
      return v ? { x: v.x, y: v.y, z: v.z } : null
    },
    shoulderWorld: (hand = 'right') => {
      const v = shoulder(hand)
      return v ? { x: v.x, y: v.y, z: v.z } : null
    },
    rootPosition: () => ({ x: root.position.x, y: root.position.y, z: root.position.z }),
    setHeadPose(pose) {
      // ⚠️ **زوايا الكاميرا تتحوّل لنقطة نظر، والتوجيه يصير
      // بـ`BoneLookController` مال بابل — مو بكتابة يدوية على
      // العظمة.**
      //
      // الكتابة اليدوية (`setYawPitchRoll` داخل حلقة التحديث)
      // **ما تُرفع للرسم**، وهذا مقاس من داخل المحرّك مو استنتاج:
      // بنفس الإطار واللحظة، الإشارة بالـIK غيّرت **١٨.٩٨٪** من
      // الصورة، وكتابتي على الرأس غيّرت **٠٪** وقمة الرأس
      // تحرّكت **٠ مم**. ونفس الاستدعاء من برّا الحلقة يحرّكها
      // **٧١ مم** — يعني الكتابة صحيحة بس ضايعة بين تمرير
      // الحركات والرسم. وبدل ما أحاول أجبر الرفع بثلاث طرق
      // (`prepare(true)` و`computeAbsoluteMatrices` والاثنين
      // سوا — وكلها طلعت **٠٪**)، أستخدم الأداة الي القياس يثبت
      // إنها تُرسم.
      if (!pose) { looking = false; return }
      const head = worldOf(headBone)
      if (!head) return
      // اتجاه الوجه يُقاس من محور الكتفين، وعليه تنطبق الزوايا.
      let side = new Vector3(1, 0, 0)
      const la = worldOf(bone(naming.arm('left')))
      const ra = worldOf(bone(naming.arm('right')))
      if (la && ra) {
        side = ra.subtract(la)
        side.y = 0
        if (side.lengthSquared() > 1e-6) side.normalize()
      }
      const fwd = Vector3.Cross(new Vector3(0, 1, 0), side)
      if (fwd.lengthSquared() < 1e-6) return
      fwd.normalize()
      // ⚠️ الميل موجب = ينظر **للأسفل** (نفس اصطلاح ميدياپايپ)،
      // فالإشارة معكوسة على y.
      const dir = fwd.scale(Math.cos(pose.yaw)).add(side.scale(Math.sin(pose.yaw)))
      dir.y = -Math.sin(pose.pitch)
      dir.normalize()
      // متر واحد كافي: `BoneLookController` يهمّه الاتجاه.
      this.lookAt(head.add(dir.scale(1.0)))
    },
    setFingerCurl(hand, curls) {
      const dest = curlWanted[hand]
      for (const [k, v] of Object.entries(curls)) {
        if (typeof v !== 'number') continue
        // أسماء ميدياپايپ صغيرة (`index`) وأسماء ميكسامو بحرف كبير
        // (`Index`) — التطبيع هنا حتى الشاشة ما تهتم بالفرق.
        const key = k.charAt(0).toUpperCase() + k.slice(1)
        dest[key] = Math.min(1, Math.max(0, v))
      }
    },
    fingerBoneReport() {
      const out: Record<string, number> = {}
      for (const hand of ['right', 'left'] as const) {
        for (const f of FINGER_KEYS) {
          out[`${hand}.${f}`] = fingerBones(hand, f).length
        }
      }
      return out
    },
    expressionNames() {
      const out = new Set<string>()
      for (const mgr of morphOwners) {
        for (let i = 0; i < mgr.numTargets; i++) out.add(mgr.getTarget(i).name)
      }
      return [...out]
    },
    setExpression(name, weight) {
      const w = Math.min(1, Math.max(0, weight))
      // نطبّقه على **كل** جسم فيه نفس التعبير: الوجه والذقن والحواجب
      // أجسام منفصلة، ولو حرّكنا الوجه وحده تصير الابتسامة والذقن
      // مو متطابقين.
      for (const mgr of morphOwners) {
        for (let i = 0; i < mgr.numTargets; i++) {
          const t = mgr.getTarget(i)
          if (t.name === name) t.influence = w
        }
      }
    },
    lookAt(target) {
      if (!target || !look) { looking = false; return }
      lookTarget.position.copyFrom(target)
      look.target = lookTarget.position
      looking = true
    },
    walkTo(target, speed = 0.9) {
      // ⚠️ مشية جديدة تلغي الي قبلها بلا ما يبقى وعد معلّق للأبد
      if (walk) { const old = walk.done; walk = null; old() }
      return new Promise<void>((resolve) => { walk = { to: target.clone(), speed, done: resolve } })
    },
    walkBy(dx, dz, speed = 0.9) {
      return this.walkTo(root.position.add(new Vector3(dx, 0, dz)), speed)
    },
    dispose() {
      applyArmMask(null)
      // نوقف التوجيه والكتابة على الأصابع قبل فكّ المراقب، وإلا
      // تبقى العظام على آخر وضع كتبناه والمقطع ما يسترجعها.
      looking = false
      curlWanted.right = {}
      curlWanted.left = {}
      // نرجّع كل إصبع بالدوران المعاكس لنفس المقدار المطبَّق.
      for (const [name, ang] of fingerAngle) {
        const b = bone(name)
        if (b && Math.abs(ang) > 1e-4) b.rotate(fingerAxis, ang, Space.LOCAL)
      }
      fingerAngle.clear()
      scene.onAfterAnimationsObservable.remove(observer)
      ikTarget.dispose(); poleTarget.dispose(); lookTarget.dispose()
    },
  }
}

/**
 * يحوّل موقع عنصر بالصفحة لنقطة بالعالم الثلاثي — حتى الكائن يشاور
 * على **زر حقيقي** بالواجهة مو على إحداثيات مخترعة.
 *
 * ⚠️ **تحويل تقريبي بقصد**: يفترض أن المشهد معروض بمستطيل معلوم
 * وبمقياس متر واحد = `metersPerPixel`. القياس النهائي بالتجربة:
 * أعطيه ثلاثة أزرار وأقيس مسافة طرف الإصبع عن مركز الزر.
 */
export function domPointToWorld(
  el: Element,
  canvas: HTMLCanvasElement,
  metersPerPixel = 0.0032,
): Vector3 {
  const r = el.getBoundingClientRect()
  const c = canvas.getBoundingClientRect()
  const cx = r.left + r.width / 2 - (c.left + c.width / 2)
  const cy = r.top + r.height / 2 - (c.top + c.height / 2)
  // y بالشاشة ينزل، وبالعالم يطلع — فالإشارة تنعكس
  return new Vector3(cx * metersPerPixel, 1.35 - cy * metersPerPixel, 0.35)
}
