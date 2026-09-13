// ═══ محرّك عرض شخصية الكيان — يُحمَّل مؤجَّلاً بقصد ═══
//
// ⚠️ **هذا الملف ما يُستورد مباشرة إطلاقاً** — يُجاب بـ`import()`
// من `EntityAvatar.tsx` بس. السبب مقاس ببناء إنتاج: حزمة العارض
// **١.٢ م.ب مضغوطة**، والحزمة الأولية تبقى **١١٧ ك.ب بلا زيادة**
// لأن الاستيراد مؤجَّل. لو انستورد مباشرة، كل موظف يفتح النظام
// يحمّل ١.٢ م.ب حتى لو ما شاف ولا قصة.
import { Engine } from '@babylonjs/core/Engines/engine'
import { Scene } from '@babylonjs/core/scene'
import { ArcRotateCamera } from '@babylonjs/core/Cameras/arcRotateCamera'
import { HemisphericLight } from '@babylonjs/core/Lights/hemisphericLight'
import { DirectionalLight } from '@babylonjs/core/Lights/directionalLight'
import { Matrix, Vector3 } from '@babylonjs/core/Maths/math.vector'
import { Color4 } from '@babylonjs/core/Maths/math.color'
import { SceneLoader } from '@babylonjs/core/Loading/sceneLoader'
import { TransformNode } from '@babylonjs/core/Meshes/transformNode'
import type { AnimationGroup } from '@babylonjs/core/Animations/animationGroup'
import type { Skeleton } from '@babylonjs/core/Bones/skeleton'
import { buildMotionRig, neutralizeRootMotion, type MotionRig } from './entityMotion'
import '@babylonjs/loaders/glTF'

/**
 * أسماء المقاطع بالملف = ثوابت إجراءات القصة بالخادم
 * (`backend-go/internal/model/story.go`). ⚠️ أي اسم ينكتب هنا ومو
 * موجود بالملف **يرجع صامتاً بلا حركة** — فالمحرّك يتحقق ويقع على
 * بديل بدل ما يسكت.
 */
export type ClipName =
  | 'WALK_TO_TARGET' | 'WALK_ALT' | 'RUN_TO_EDGE'
  | 'SPEAK' | 'SPEAK_ALT' | 'CELEBRATE'
  | 'SIT_IDLE' | 'SIT_SPEAK' | 'LAYING_IDLE'
  | 'SHOW_WARNING' | 'POINT_AT_UI'

/** المقاطع الي **تُشغَّل حلقياً** — الباقي يُشغَّل مرة وبس. */
const LOOPING: ReadonlySet<string> = new Set([
  'WALK_TO_TARGET', 'WALK_ALT', 'RUN_TO_EDGE', 'SIT_IDLE', 'LAYING_IDLE',
])

export interface AvatarHandle {
  play(clip: ClipName): void
  /** يوقف كل المقاطع — للحركة المحسوبة الصافية بلا مقطع تحتها. */
  stopClips(): void
  dispose(): void
  clips(): string[]
  /**
   * منظومة الحركة المحسوبة — `null` لو الشخصية بلا هيكل عظمي.
   * ⚠️ **تُبنى بس لو `motion: true`** حتى الودجة العائمة ما تشيل
   * كلفة حلقة حساب ما تحتاجها.
   */
  motion: MotionRig | null
  /** قياسات حقيقية للتجربة — رقم مو إحساس. */
  stats: AvatarStats
  /** موقع الرأس **بالبكسل** على الكانفس — لقياس التأطير. */
  headPixel(): { x: number; y: number } | null
  /** أبعاد الكانفس بالبكسل. */
  canvasSize(): { w: number; h: number }
  /**
   * يعيد حساب التأطير الآن — للمقاسات الجديدة وللمجسّمات الجديدة.
   * ⚠️ والتأطير وقت التحميل يصير على حالة ما استقرّت بعد، فهاي
   * تنطي الشاشة طريقة تصحيح صريحة بلا إعادة تحميل المشهد.
   */
  reframe(): void
  /**
   * يقرأ إطاراً **من داخل المحرّك** ويرجّع بصمته — عدد البكسلات
   * غير الشفافة ومجموع القنوات.
   *
   * ⚠️ **وهاي الطريقة الوحيدة الي طلعت صادقة**: لقطات المتصفح
   * لكانفس WebGL رجّعت **نفس الصورة بالبت** لأربع وقفات مختلفة
   * (فرق أقصى = صفر) لأن المُركّب يعيد استخدام آخر لقطة. وبنفس
   * اللحظة قياس العظام أثبت حركة **٥١ مم**. فالقياس من داخل
   * المحرّك بعد إجبار الرسم هو الي يوصف الي انرسم فعلاً.
   */
  frameSignature(): Promise<{ opaque: number; sum: number; w: number; h: number } | null>
  /** تقرير الكاميرات — منو الكاميرا الفعّالة فعلاً. */
  cameraReport(): {
    cameras: string[]; active: string | null; ours: string; isOurs: boolean
    radius: number; alpha: number; beta: number; fov: number
    pos: { x: number; y: number; z: number }; tgt: { x: number; y: number; z: number }
    engines: number
  }
  /** أرقام التأطير المحسوبة + موقع نقطة التسليط بالبكسل. */
  framingDebug(): (FramingDebug & { aimPixel: { x: number; y: number } | null }) | null
  /**
   * موقع أي عظمة بالعالم — **أداة قياس** للتجربة.
   *
   * ⚠️ **يُقاس داخل الإطار، ولهذا يرجّع وعداً.** وهاي غلطة قياس
   * كلّفتني وقتاً وكِدت أحكم على ميزة سليمة إنها مكسورة:
   *   ① `headPixel()` يقرأ **مفصل** الرأس، والمفصل هو محور الدوران
   *      نفسه — فلفّ الرأس **ما يزحزحه ولا ملّيمتر**. ثلاث زوايا
   *      مختلفة أعطت **نفس البكسل بالضبط**. فالقياس على عظمة
   *      **بنت** (قمة الرأس) لأنها تتأرجح مع اللفّة.
   *   ② والقراءة **من برّا حلقة الرسم** ترجّع وضع الراحة دائماً:
   *      ترتيب الإطار هو «الحركات تكتب على العظام ← إحنا نكتب
   *      فوقها ← الرسم». فالقراءة بين إطارين تلتقط **كتابة نظام
   *      الحركات** مو كتابتنا. قِستها: القراءة من برّا = **صفر مم**
   *      بينما نفس اللحظة داخل الإطار = **٤٦ مم** والكواتيرنيون
   *      مضبوط. والي يشوفه المستخدم هو المرسوم، فالقياس لازم
   *      يصير بنفس اللحظة الي يرسم بيها.
   */
  boneWorld(name: string): Promise<{ x: number; y: number; z: number } | null>
  /**
   * يغيّر عدد تأثيرات العظام لحظياً — للمقارنة المقاسة ٤ مقابل ٨.
   * ⚠️ **بابل افتراضه ٤**، وملفنا فيه `JOINTS_0`+`JOINTS_1` = ٨،
   * فبالافتراضي **يُقصّ التشوّه صامتاً بلا أي خطأ بالكونسول**.
   */
  setInfluencers(n: number): void
}

/**
 * أرقام التأطير **كما حُسبت فعلاً** — التأطير انكسر ثلاث مرات
 * بأسباب مختلفة، وكل مرة كلّفتني جولة تخمين. فالأرقام تُعرض بدل
 * ما تُخمَّن.
 */
export interface FramingDebug {
  boxMin: { x: number; y: number; z: number }
  boxMax: { x: number; y: number; z: number }
  aim: { x: number; y: number; z: number }
  radius: number
  aspect: number
  renderW: number
  renderH: number
}

export interface AvatarStats {
  /** زمن التحميل بالملي ثانية — مقاس بـ`performance.now()`. */
  loadMs: number
  /** عدد المفاصل بالهيكل (٤١ بالشخصية الحالية). */
  joints: number
  /** أقصى تأثيرات عظام موجودة فعلاً بالملف. */
  influencers: number
  /** عدد الرؤوس — يفسّر كلفة الذاكرة. */
  vertices: number
  /** أسماء المقاطع المحمّلة. */
  clips: string[]
}

/**
 * يبني مشهداً ويرجّع مقبضاً. يرمي لو فشل التحميل — والمتصل يلتقط
 * ويبقى على النسخة النصية.
 *
 * @param url مسار الملف **مع قاعدة التطبيق** (`import.meta.env.BASE_URL`).
 */
export async function mountAvatar(
  canvas: HTMLCanvasElement,
  url: string,
  initial: ClipName,
  framing: 'bust' | 'full' = 'bust',
  options: { influencers?: number; motion?: boolean } = {},
): Promise<AvatarHandle> {
  const t0 = performance.now()
  const engine = new Engine(canvas, true, { preserveDrawingBuffer: false, stencil: false })
  const scene = new Scene(engine)
  scene.clearColor = new Color4(0, 0, 0, 0)   // شفاف — الكارت وراه يبين
  // ⚠️ **نظام يميني قبل التحميل.** بلاه محمّل glTF يزرع عقدة
  // `__root__` بمقياس **z = −1** حتى يقلب من اليميني للشمالي، ومصفوفة
  // مرآة تخرب حساب الـIK: قِستها — الذراع تتحرّك ١.٧ م بس **بالاتجاه
  // المعاكس**، وخطأ الإشارة ١٦٠° بدل ما يكون دون ١٠°.
  scene.useRightHandedSystem = true

  const result = await SceneLoader.ImportMeshAsync('', url, '', scene)
  const groups: AnimationGroup[] = result.animationGroups

  // ⚠️ **التأثيرات قبل أي شي**: بابل افتراضه ٤ لكل رأس، وملفنا فيه
  // ثمانية. والقصّ يصير **صامتاً** — لا خطأ ولا تحذير، بس تشوّه
  // مختلف عند الكتف والكم. فنرفعها، ونخلي التغيير ممكناً للقياس.
  const wanted = options.influencers ?? 8
  const setInfluencers = (n: number) => {
    for (const m of result.meshes) {
      if (m.skeleton) m.numBoneInfluencers = n
    }
  }
  setInfluencers(wanted)
  // الملف يوصل ومقاطعه موقوفة؛ بابل يشغّل الأول تلقائياً فنوقفها كلها.
  groups.forEach((g) => g.stop())

  // الشخصية ١.٧٠ م والقدمان على Z≈0 (مقيس) — فالكاميرا على ارتفاع
  // الصدر تنظر للأعلى شوي، حتى يبين الوجه والإيدين مو الأرجل.
  // ⚠️ التأطير مقاس على الشاشة مو مخمَّن: بنصف قطر ٣.٠ وهدف ١.١٥
  // طلعت الشخصية صغيرة والوجه ما يبين. الصدر-فوق أوضح — الموظف
  // يحتاج يشوف الوجه والإيد الي تشاور، مو الأرجل.
  // `full` للبوت العائم الصغير: كل الجسم حتى تبين الوقفة والحركة،
  // بلاه تطلع رقبة بمربّع. و`bust` لورقة القصة — الوجه والإيد.
  // ⚠️ **التأطير كله مقاس من العظام، ولا زاوية مكتوبة**.
  // قِستها: الزوايا المكتوبة (`alpha = π/2 + 0.30`) صحّت بنظام
  // بابل الشمالي، وبعد ما حوّلت المشهد لنظام يميني — وهو **شرط**
  // حتى حساب الـIK ما ينقلب — طلع الوجه **مزيّح عن الوسط**. فبدل
  // ما أعاير زاوية ثانية تنكسر بأول تغيير، الكاميرا تتحدد من
  // **اتجاه وجه الشخصية** نفسه: محور الكتفين يعطي الجانب، وضربه
  // بالشمال يعطي الأمام.
  const skel = result.skeletons[0] ?? null
  const skinnedForAim = result.meshes.find((m) => !!m.skeleton) ?? null
  const boneAt = (name: string) => {
    const b = skel?.bones.find((x) => x.name === name)
    return b && skinnedForAim ? b.getAbsolutePosition(skinnedForAim) : null
  }
  const hipsPos = boneAt('mixamorig:Hips')
  const lArm = boneAt('mixamorig:LeftArm')
  const rArm = boneAt('mixamorig:RightArm')
  const center = hipsPos ?? new Vector3(0, 0.92, 0)
  const target = framing === 'full'
    ? new Vector3(center.x, 0.92, center.z)
    : new Vector3(center.x, 1.32, center.z)
  const camera = new ArcRotateCamera('cam', Math.PI / 2, Math.PI / 2, 3.15, target, scene)
  camera.minZ = 0.05
  camera.fov = framing === 'full' ? 0.62 : 0.72

  /**
   * يؤطّر الشخصية من **صندوقها المحيط بالعالم ونسبة الكانفس**.
   *
   * ⚠️ **وهذا التصحيح الثالث للتأطير، والسببان الي كسّرا الي قبله**:
   *   ① **القياس من عظمة الحوض وقت التحميل**: الحوض ينزاح بعدها
   *      (التحييد يرجّعه لوضع الراحة، والمقطع يحرّكه)، فالكاميرا
   *      تبقى مسلّطة على نقطة **قديمة** والشخصية تطلع بحاشية
   *      الإطار — قِستها: الرأس عند **٥٣٦ من ٦٧٤** أفقياً و**٥٨
   *      من ٤١٦** عمودياً، أي مقطوع من فوق وملزوق باليمين.
   *   ② **نصف قطر ثابت (٣.١٥) يهمل نسبة الكانفس**: بابل يثبّت
   *      المجال **عمودياً**، فبكانفس عريض (٦٧٤×٤١٦) الشخصية
   *      تتجاوز العرض. والصح يُحسب من النسبة الفعلية.
   * فالصندوق المحيط يصف **الي انرسم فعلاً**، ويتصحّح لحاله مع أي
   * مجسّم جديد (وهذا مهم لأن مجسّم VRM الجاي بأبعاد مختلفة).
   */
  let lastFraming: FramingDebug | null = null
  const frameCamera = () => {
    let min: Vector3 | null = null
    let max: Vector3 | null = null
    // ⚠️ **الأجسام المكسوّة بالهيكل حصراً**: الملف فيه أجسام مساعدة
    // بلا هيكل، ولمن دخلت بالصندوق زاحت نقطة المركز يساراً فطلعت
    // الشخصية **يمين الوسط** (قِستها: الرأس عند ٤٦٠ من ٦٧٤ والهدف
    // بالمنتصف). والجسم المكسوّ هو الشخصية نفسها.
    for (const m of result.meshes) {
      if (!m.skeleton) continue
      if (!m.getTotalVertices || m.getTotalVertices() === 0) continue
      m.computeWorldMatrix(true)
      m.refreshBoundingInfo({ applySkeleton: true })
      const bb = m.getBoundingInfo().boundingBox
      min = min ? Vector3.Minimize(min, bb.minimumWorld) : bb.minimumWorld.clone()
      max = max ? Vector3.Maximize(max, bb.maximumWorld) : bb.maximumWorld.clone()
    }
    if (!min || !max) return
    const size = max.subtract(min)
    const mid = min.add(max).scale(0.5)
    // `full` يأخذ كل الجسم، و`bust` يأخذ الثلث الأعلى (وجه وإيدين).
    const aim = framing === 'full'
      ? new Vector3(mid.x, mid.y, mid.z)
      : new Vector3(mid.x, min.y + size.y * 0.82, mid.z)
    const wantH = framing === 'full' ? size.y * 1.22 : size.y * 0.46
    const wantW = framing === 'full' ? size.x * 1.25 : size.x * 0.95
    const eng = scene.getEngine()
    const aspect = Math.max(eng.getRenderWidth(), 1) / Math.max(eng.getRenderHeight(), 1)
    const halfV = camera.fov / 2
    // ⚠️ المجال الأفقي يُشتق من العمودي بالنسبة — وبلا هاي الخطوة
    // الكانفس العريض يقصّ الشخصية من الجانب.
    const halfHTan = Math.tan(halfV) * aspect
    const distV = (wantH / 2) / Math.tan(halfV)
    const distH = (wantW / 2) / halfHTan
    const radius = Math.max(distV, distH, 0.6)
    // الاتجاه من **الكتفين المقاسين**: محور الكتفين يعطي الجانب،
    // وضربه بالشمال يعطي الأمام. و`Cross(side, up)` تنطي الظهر.
    let dir = new Vector3(0, 0, 1)
    let side = new Vector3(1, 0, 0)
    if (lArm && rArm) {
      side = rArm.subtract(lArm)
      side.y = 0
      if (side.lengthSquared() > 1e-6) side.normalize()
      const forward = Vector3.Cross(new Vector3(0, 1, 0), side)
      if (forward.lengthSquared() > 1e-6) dir = forward.normalize()
    }
    const tilt = framing === 'full' ? 0.02 : 0.06
    const off = dir.scale(Math.cos(0.25)).add(side.scale(Math.sin(0.25)))
    camera.setTarget(aim)
    camera.setPosition(aim.add(off.scale(radius)).add(new Vector3(0, radius * tilt, 0)))
    lastFraming = {
      boxMin: { x: min.x, y: min.y, z: min.z },
      boxMax: { x: max.x, y: max.y, z: max.z },
      aim: { x: aim.x, y: aim.y, z: aim.z },
      radius,
      aspect,
      renderW: eng.getRenderWidth(),
      renderH: eng.getRenderHeight(),
    }
  }

  const amb = new HemisphericLight('amb', new Vector3(0, 1, 0), scene)
  amb.intensity = 0.85
  const key = new DirectionalLight('key', new Vector3(-0.6, -1, 0.7), scene)
  key.intensity = 1.5

  let current: AnimationGroup | null = null
  const play = (clip: ClipName) => {
    const g = groups.find((x) => x.name === clip)
    if (!g) return                      // اسم مو موجود: نخلي الحالي يكمل
    if (current === g) return
    current?.stop()
    current = g
    g.start(LOOPING.has(clip), 1.0, g.from, g.to)
  }
  // ⚠️ **ما نشغّل المقطع قبل التحييد**: التحييد يقرأ مرجعه من
  // مصفوفة الراحة، بس تشغيل المقطع قبل بناء المنظومة يخلي أول
  // إطار محسوب يجي على عظام متقدّمة بالمشي. فالترتيب: نبني
  // المنظومة ← ثم نشغّل.
  // ═══ الجذر والحركة المحسوبة ═══
  // التحريك البرمجي لازم يصير على **عقدة أب**، مو على عظمة الحوض:
  // المقطع يكتب على العظام كل إطار، فأي إزاحة نحطها عليها تُمحى.
  const root = new TransformNode('entity-root', scene)
  // العقدة الحاملة تمتص إزاحة المقطع، والجذر يحمل المشي المقصود —
  // فصلهما يمنع تنازع الاثنين على نفس الرقم.
  const carrier = new TransformNode('entity-carrier', scene)
  carrier.parent = root
  const skinned = result.meshes.find((m) => !!m.skeleton) ?? null
  const skeleton: Skeleton | null = skinned?.skeleton ?? null
  for (const m of result.meshes) {
    if (!m.parent) m.parent = carrier
  }

  let motion: MotionRig | null = null
  let restoreRoot: (() => void) | null = null
  if (options.motion && skeleton && skinned) {
    // ⚠️ التحييد **قبل** بناء المنظومة: المقاطع فيها ٤١ مسار إزاحة
    // على `Hips` (مقاس)، وبلا تحييدها المشي البرمجي **يتضاعف**.
    restoreRoot = neutralizeRootMotion(skeleton, carrier, skinned)
    motion = buildMotionRig(scene, skeleton, skinned, root, groups)
  }
  play(initial)
  // ⚠️ **بعد** بناء المنظومة وتشغيل المقطع: التأطير يقرأ الصندوق
  // المحيط وقتها، فيوصف الوضع الي راح ينرسم فعلاً.
  frameCamera()

  const stats: AvatarStats = {
    loadMs: Math.round(performance.now() - t0),
    joints: skeleton?.bones.length ?? 0,
    influencers: wanted,
    vertices: result.meshes.reduce((n, m) => n + (m.getTotalVertices?.() ?? 0), 0),
    clips: groups.map((g) => g.name),
  }

  engine.runRenderLoop(() => scene.render())
  // ⚠️ التأطير يُعاد حسابه مع القياس: نسبة الكانفس تدخل بالحساب،
  // فتغيير حجم النافذة بلا إعادة تأطير يقصّ الشخصية من الجانب.
  const onResize = () => { engine.resize(); frameCamera() }
  window.addEventListener('resize', onResize)

  return {
    headPixel() {
      const head = skel?.bones.find((b) => b.name === 'mixamorig:Head')
      if (!head || !skinnedForAim || !scene.activeCamera) return null
      const w = engine.getRenderWidth(), h = engine.getRenderHeight()
      const v = Vector3.Project(
        head.getAbsolutePosition(skinnedForAim),
        Matrix.Identity(), scene.getTransformMatrix(),
        scene.activeCamera.viewport.toGlobal(w, h),
      )
      return { x: v.x, y: v.y }
    },
    canvasSize: () => ({ w: engine.getRenderWidth(), h: engine.getRenderHeight() }),
    async frameSignature() {
      // ⚠️ **نجبر رسمة جديدة قبل القراءة**: المخزن ما يُحفظ
      // (`preserveDrawingBuffer: false`)، فالقراءة بلا رسم تجي
      // على مخزن منتهي.
      scene.render()
      const w = engine.getRenderWidth(), h = engine.getRenderHeight()
      const raw = await engine.readPixels(0, 0, w, h)
      const data = new Uint8Array(raw.buffer, raw.byteOffset, raw.byteLength)
      let opaque = 0, sum = 0
      for (let i = 0; i < data.length; i += 4) {
        if (data[i + 3] > 8) opaque++
        sum += data[i] + data[i + 1] + data[i + 2]
      }
      return { opaque, sum, w, h }
    },
    cameraReport: () => ({
      cameras: scene.cameras.map((c) => c.name),
      active: scene.activeCamera?.name ?? null,
      ours: camera.name,
      isOurs: scene.activeCamera === camera,
      radius: camera.radius,
      alpha: camera.alpha,
      beta: camera.beta,
      fov: camera.fov,
      pos: { x: camera.position.x, y: camera.position.y, z: camera.position.z },
      tgt: { x: camera.target.x, y: camera.target.y, z: camera.target.z },
      engines: 1,
    }),
    reframe: () => frameCamera(),
    framingDebug() {
      if (!lastFraming || !scene.activeCamera) return null
      const w = engine.getRenderWidth(), h = engine.getRenderHeight()
      const v = Vector3.Project(
        new Vector3(lastFraming.aim.x, lastFraming.aim.y, lastFraming.aim.z),
        Matrix.Identity(), scene.getTransformMatrix(),
        scene.activeCamera.viewport.toGlobal(w, h),
      )
      return { ...lastFraming, aimPixel: { x: v.x, y: v.y } }
    },
    boneWorld(name) {
      const b = skel?.bones.find((x) => x.name === name)
      if (!b || !skinnedForAim) return Promise.resolve(null)
      // القياس يصير **داخل** الإطار وبعد كتابتنا — نفس اللحظة الي
      // يُرسم بيها المشهد. مراقب لمرة واحدة حتى ما يكلّف كل إطار.
      return new Promise((resolve) => {
        const once = scene.onAfterAnimationsObservable.addOnce(() => {
          skel?.computeAbsoluteMatrices(true)
          skinnedForAim.computeWorldMatrix(true)
          const v = b.getAbsolutePosition(skinnedForAim)
          resolve({ x: v.x, y: v.y, z: v.z })
        })
        // ⚠️ لو الحلقة موقوفة، الوعد يبقى معلّقاً للأبد — فنحرّر
        // بعد مهلة بدل ما نعلّق الشاشة.
        setTimeout(() => {
          if (once) scene.onAfterAnimationsObservable.remove(once)
          resolve(null)
        }, 1000)
      })
    },
    play,
    stopClips() { current?.stop(); current = null },
    clips: () => groups.map((g) => g.name),
    motion,
    stats,
    setInfluencers(n) { setInfluencers(n); stats.influencers = n },
    dispose() {
      window.removeEventListener('resize', onResize)
      motion?.dispose()
      restoreRoot?.()
      engine.stopRenderLoop()
      scene.dispose()
      engine.dispose()
    },
  }
}
