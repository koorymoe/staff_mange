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
   * يغيّر عدد تأثيرات العظام لحظياً — للمقارنة المقاسة ٤ مقابل ٨.
   * ⚠️ **بابل افتراضه ٤**، وملفنا فيه `JOINTS_0`+`JOINTS_1` = ٨،
   * فبالافتراضي **يُقصّ التشوّه صامتاً بلا أي خطأ بالكونسول**.
   */
  setInfluencers(n: number): void
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
  if (lArm && rArm) {
    const side = rArm.subtract(lArm)
    side.y = 0
    side.normalize()
    // ⚠️ **ترتيب الضرب الاتجاهي يحدد الوجه من الظهر**: بـ
    // `Cross(side, up)` طلعت الكاميرا **ورا الشخصية** (قِستها
    // بالصورة: ظهر وسط الإطار). والصحيح `Cross(up, side)`.
    const forward = Vector3.Cross(new Vector3(0, 1, 0), side).normalize()
    const radius = framing === 'full' ? 3.15 : 1.75
    // إزاحة بسيطة بالزاوية (٠.٢٥ راديان) تنطي عمقاً بدل مسطّح تماماً
    const tilt = framing === 'full' ? 0.02 : 0.06
    const dir = forward.scale(Math.cos(0.25)).add(side.scale(Math.sin(0.25)))
    camera.setPosition(target.add(dir.scale(radius)).add(new Vector3(0, radius * tilt, 0)))
    camera.setTarget(target)
  }
  camera.fov = framing === 'full' ? 0.62 : 0.72
  camera.minZ = 0.05

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
  play(initial)

  // ═══ الجذر والحركة المحسوبة ═══
  // التحريك البرمجي لازم يصير على **عقدة أب**، مو على عظمة الحوض:
  // المقطع يكتب على العظام كل إطار، فأي إزاحة نحطها عليها تُمحى.
  const root = new TransformNode('entity-root', scene)
  const skinned = result.meshes.find((m) => !!m.skeleton) ?? null
  const skeleton: Skeleton | null = skinned?.skeleton ?? null
  for (const m of result.meshes) {
    if (!m.parent) m.parent = root
  }

  let motion: MotionRig | null = null
  let restoreRoot: (() => void) | null = null
  if (options.motion && skeleton && skinned) {
    // ⚠️ التحييد **قبل** بناء المنظومة: المقاطع فيها ٤١ مسار إزاحة
    // على `Hips` (مقاس)، وبلا تحييدها المشي البرمجي **يتضاعف**.
    restoreRoot = neutralizeRootMotion(skeleton)
    motion = buildMotionRig(scene, skeleton, skinned, root, groups)
  }

  const stats: AvatarStats = {
    loadMs: Math.round(performance.now() - t0),
    joints: skeleton?.bones.length ?? 0,
    influencers: wanted,
    vertices: result.meshes.reduce((n, m) => n + (m.getTotalVertices?.() ?? 0), 0),
    clips: groups.map((g) => g.name),
  }

  engine.runRenderLoop(() => scene.render())
  const onResize = () => engine.resize()
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
