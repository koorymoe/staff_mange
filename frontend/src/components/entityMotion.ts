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

/** أسماء عظام ميكسامو الي نحتاجها — مقاسة موجودة بالشخصية. */
const BONE = {
  hips: 'mixamorig:Hips',
  head: 'mixamorig:Head',
  rightArm: 'mixamorig:RightArm',
  rightForeArm: 'mixamorig:RightForeArm',
  leftArm: 'mixamorig:LeftArm',
  leftForeArm: 'mixamorig:LeftForeArm',
} as const

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
export function neutralizeRootMotion(skeleton: Skeleton): () => void {
  const hips = skeleton.bones.find((b) => b.name === BONE.hips)
  if (!hips) return () => {}
  // نجمّد الإزاحة الأفقية ونخلي العمودية (النزول والطلوع بالخطوة).
  const original = hips.getPosition().clone()
  const observer = () => {
    const p = hips.getPosition()
    hips.setPosition(new Vector3(original.x, p.y, original.z))
  }
  // ⚠️ **بعد** تحديث الهياكل مو قبله — وإلا المقطع يكتب فوق تصحيحنا
  // والتحييد ما ينفّذ (وهاي غلطة تطلع كأنها «التحييد ما نفع»).
  const scene = skeleton.getScene()
  scene.onAfterAnimationsObservable.add(observer)
  return () => { scene.onAfterAnimationsObservable.removeCallback(observer) }
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

  // ═══ أهداف الحركة: عُقد غير مرئية يتبعها الـIK ═══
  const ikTarget = new TransformNode('entity-ik-target', scene)
  // القطب يحدد **اتجاه المرفق** — بلاه المرفق يلتوي لجهة غلط
  // والإشارة تطلع مكسورة، وهاي مو عيب بالـIK بل قطب ناقص.
  const poleTarget = new TransformNode('entity-ik-pole', scene)
  const lookTarget = new TransformNode('entity-look-target', scene)

  const ik: Record<'right' | 'left', BoneIKController | null> = { right: null, left: null }
  for (const side of ['right', 'left'] as const) {
    const fore = bone(side === 'right' ? BONE.rightForeArm : BONE.leftForeArm)
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

  const headBone = bone(BONE.head)
  const look = headBone
    ? new BoneLookController(mesh, headBone, lookTarget.position, {
      // حدود بشرية: الرأس ما يلتف ٩٠° ولا يرفع لفوق بلا حد
      maxYaw: Math.PI * 0.42,
      maxPitch: Math.PI * 0.22,
      slerpAmount: 0.22,
    })
    : null

  // ⚠️ **الإحداثيات لازم تمرّ بالمجسم**: `getAbsolutePosition()` بلا
  // تمرير المجسم ترجّع فضاء الهيكل مو العالم — وهاي غلطة قِستها:
  // طلعت أخطاء الإشارة **١٨٦–٢٢٣ سم** وهي أصلاً غلط قياس مو غلط IK.
  const worldOf = (b: ReturnType<typeof bone>) =>
    b ? b.getAbsolutePosition(skinnedMesh) : null
  const fingertip = (hand: 'right' | 'left') => {
    const side = hand === 'right' ? 'Right' : 'Left'
    // آخر عظمة بالسلسلة هي **طرف** الإصبع بميكسامو (`...Index4`)؛
    // القياس من `Index1` يخلي الإشارة تبين أدق من واقعها.
    return worldOf(bone(`mixamorig:${side}HandIndex4`) ?? bone(`mixamorig:${side}HandIndex3`))
  }
  const shoulder = (hand: 'right' | 'left') =>
    worldOf(bone(hand === 'right' ? BONE.rightArm : BONE.leftArm))
  /** مدى الذراع مقاس من العظام نفسها — مو ثابتاً مكتوباً. */
  const reachOf = (hand: 'right' | 'left'): number | null => {
    const side = hand === 'right' ? 'Right' : 'Left'
    const a = worldOf(bone(hand === 'right' ? BONE.rightArm : BONE.leftArm))
    const b = worldOf(bone(hand === 'right' ? BONE.rightForeArm : BONE.leftForeArm))
    const c = worldOf(bone(`mixamorig:${side}Hand`))
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
    const side = hand === 'right' ? 'Right' : 'Left'
    return skeleton.bones
      .map((b) => b.name)
      .filter((n) => new RegExp(`^mixamorig:${side}(Shoulder|Arm|ForeArm|Hand)`).test(n))
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
