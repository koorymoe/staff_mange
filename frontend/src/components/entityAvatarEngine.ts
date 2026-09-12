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
import { Vector3 } from '@babylonjs/core/Maths/math.vector'
import { Color4 } from '@babylonjs/core/Maths/math.color'
import { SceneLoader } from '@babylonjs/core/Loading/sceneLoader'
import type { AnimationGroup } from '@babylonjs/core/Animations/animationGroup'
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
  dispose(): void
  clips(): string[]
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
): Promise<AvatarHandle> {
  const engine = new Engine(canvas, true, { preserveDrawingBuffer: false, stencil: false })
  const scene = new Scene(engine)
  scene.clearColor = new Color4(0, 0, 0, 0)   // شفاف — الكارت وراه يبين

  const result = await SceneLoader.ImportMeshAsync('', url, '', scene)
  const groups: AnimationGroup[] = result.animationGroups
  // الملف يوصل ومقاطعه موقوفة؛ بابل يشغّل الأول تلقائياً فنوقفها كلها.
  groups.forEach((g) => g.stop())

  // الشخصية ١.٧٠ م والقدمان على Z≈0 (مقيس) — فالكاميرا على ارتفاع
  // الصدر تنظر للأعلى شوي، حتى يبين الوجه والإيدين مو الأرجل.
  // ⚠️ التأطير مقاس على الشاشة مو مخمَّن: بنصف قطر ٣.٠ وهدف ١.١٥
  // طلعت الشخصية صغيرة والوجه ما يبين. الصدر-فوق أوضح — الموظف
  // يحتاج يشوف الوجه والإيد الي تشاور، مو الأرجل.
  // `full` للبوت العائم الصغير: كل الجسم حتى تبين الوقفة والحركة،
  // بلاه تطلع رقبة بمربّع. و`bust` لورقة القصة — الوجه والإيد.
  const camera = framing === 'full'
    ? new ArcRotateCamera('cam', Math.PI / 2 + 0.22, Math.PI / 2 - 0.02, 3.15,
      new Vector3(0, 0.92, 0), scene)
    : new ArcRotateCamera('cam', Math.PI / 2 + 0.30, Math.PI / 2 - 0.04, 1.75,
      new Vector3(0, 1.32, 0), scene)
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

  engine.runRenderLoop(() => scene.render())
  const onResize = () => engine.resize()
  window.addEventListener('resize', onResize)

  return {
    play,
    clips: () => groups.map((g) => g.name),
    dispose() {
      window.removeEventListener('resize', onResize)
      engine.stopRenderLoop()
      scene.dispose()
      engine.dispose()
    },
  }
}
