// ═══ مساحة العمل الفيزيائية ثلاثية الأبعاد ═══
//
// «أريد شي يبدو وكأنه حقيقي… ماريد شي سهل».
//
// المنظر الفيزيائي بالمخطط الرئيسي (١٩): نفس الحالة بالضبط الي يقراها
// المنظر المنطقي — **مصدر واحد للحقيقة**. الأسلاك هنا وهناك نفس المصفوفة
// بنفس المعرّفات (`ref:terminalId`)، فأي توصيل تسويه بمنظر يبان بالثاني.
// منظر واحد ما يقدر يزيح الثاني ولا يخزن شي لحاله.
//
// ⚠️ ليش الالتقاط بلا حساب إحداثيات؟ كل طرف **جسم مستقل** بالمشهد،
// وBabylon يكلنا أي جسم انضغط. تحويل إحداثيات الماوس لفضاء ثلاثي
// الأبعاد هو أصعب وأهش جزء بأي محرّر ٣د — وإحنه ما نحتاجه أصلاً.
//
// ⚠️ الاستيرادات مفرّقة بمسارات عميقة عمداً (`@babylonjs/core/…`) مو
// `from '@babylonjs/core'`: الحزمة الكاملة تتجاوز ١٥ ميغا، والاستيراد
// الشجري ينزّل الي ينستعمل بس. الصفحة أصلاً `lazy` فما تلمس بقية النظام.

import { useCallback, useEffect, useRef, useState } from 'react'

import '@babylonjs/core/Culling/ray' // ⚠️ إجباري للالتقاط — بدونه scene.pick ترجع null دائماً
import '@babylonjs/core/Rendering/edgesRenderer'
import { ArcRotateCamera } from '@babylonjs/core/Cameras/arcRotateCamera'
import { Engine } from '@babylonjs/core/Engines/engine'
import { DirectionalLight } from '@babylonjs/core/Lights/directionalLight'
import { HemisphericLight } from '@babylonjs/core/Lights/hemisphericLight'
import { ShadowGenerator } from '@babylonjs/core/Lights/Shadows/shadowGenerator'
import { HighlightLayer } from '@babylonjs/core/Layers/highlightLayer'
import { StandardMaterial } from '@babylonjs/core/Materials/standardMaterial'
import { Color3, Color4 } from '@babylonjs/core/Maths/math.color'
import { Vector3 } from '@babylonjs/core/Maths/math.vector'
import { CreateGround } from '@babylonjs/core/Meshes/Builders/groundBuilder'
import { CreateTube } from '@babylonjs/core/Meshes/Builders/tubeBuilder'
import { PointerEventTypes } from '@babylonjs/core/Events/pointerEvents'
import { Matrix } from '@babylonjs/core/Maths/math.vector'
import { Scene as BabylonScene } from '@babylonjs/core/scene'
import type { Mesh } from '@babylonjs/core/Meshes/mesh'

import type { Scene, SimAction, SimDevice, Terminal, Wire } from '../sim/types'
import { buildDevice, type TerminalAnchor } from './deviceGeometry'
import { CABLE_GAUGES, connectorsCompatible, type CableGaugeId } from './cable'

interface Props {
  scene: Scene
  devices: Record<string, SimDevice>
  wires: Wire[]
  onAction: (a: SimAction) => void
  onRemoveWire: (w: Wire) => void
  highlight?: string[]
  readOnly?: boolean
  /** المقطع المختار — يؤثّر على المقاومة وبالتالي على فولتية الحمل. */
  gauge: CableGaugeId
  /** طول السلك الفعلي بالمتر (المسار الحقيقي بالميدان مو مسافة الطاولة). */
  runLengthM: number
}

/** نصف عرض الطاولة بالمتر — الأجهزة تنوزّع بيها حسب `x` بالمشهد. */
const TABLE_HALF = 0.19

export default function Workbench3D({
  scene, devices, wires, onAction, onRemoveWire, highlight, readOnly, gauge, runLengthM,
}: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const anchorsRef = useRef<Map<string, TerminalAnchor>>(new Map())
  const engineRef = useRef<Engine | null>(null)
  const sceneRef = useRef<BabylonScene | null>(null)
  const cableMeshRef = useRef<Mesh[]>([])
  const hlRef = useRef<HighlightLayer | null>(null)

  const [armed, setArmed] = useState<string | null>(null)
  const [hover, setHover] = useState<TerminalAnchor | null>(null)
  const [reject, setReject] = useState<string | null>(null)
  /** اسم الطرف المسلّح — **حالة** مو قراءة من ref أثناء الرندر. */
  const [armedLabel, setArmedLabel] = useState('')
  const [ready, setReady] = useState(false)

  // ⚠️ الدوال تنقرا من ref داخل حلقة الرسم: لو انقرت من الإغلاق مباشرة
  // تبقى النسخة الأولى للأبد (المشهد ينبني مرة وحدة والحلقة تعيش بعده).
  const cb = useRef({ onAction, onRemoveWire, readOnly, armed })
  /** الطرف المسلّح — يُقرا داخل حلقة الأحداث فلازم يكون بـref هم. */
  const armedRef = useRef<string | null>(null)
  // الأسلاك تنقرا هي هم داخل حلقة الأحداث.
  const wiresRef = useRef<Wire[]>(wires)
  // ⚠️ التزامن بتأثير مو بجسم الرندر: الكتابة على ref أثناء الرندر تكسر
  // الرندر النقي (وقاعدة react-hooks تمنعها).
  useEffect(() => {
    cb.current = { onAction, onRemoveWire, readOnly, armed }
    wiresRef.current = wires
    armedRef.current = armed
  })

  // ═══ بناء المشهد — مرة وحدة ═══
  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const engine = new Engine(canvas, true, { preserveDrawingBuffer: true, stencil: true })
    const bs = new BabylonScene(engine)
    bs.clearColor = new Color4(0.055, 0.075, 0.11, 1)
    // ⚠️ **اليد اليمنى إجبارية هنا**. Babylon افتراضه اليد اليسرى (+Z
    // داخل الشاشة)، فالكاميرا الواگفة بـ+Z تشوف المشهد **من ورا** —
    // كل شي ينقلب مرآة: الأرقام تنقرا ٣٢١، والنص معكوس، والأطراف
    // تطلع بالجهة الغلط. وهذا هم اصطلاح glTF الي يعتمده المخطط
    // (٧٫٤) لأصول العرض — فالتوافق مضمون لمن تجي موديلات حقيقية.
    bs.useRightHandedSystem = true
    engineRef.current = engine
    sceneRef.current = bs

    // ⚠️ alpha = **+**PI/2 مو −: موقع الكاميرا يجي من
    // (cos α·sin β, cos β, sin α·sin β)، فالسالب يحطها بـ−Z يعني **ورا**
    // الأجهزة — تشوف ظهورها والأطراف كلها بالجهة الثانية.
    const camera = new ArcRotateCamera('cam', Math.PI / 2, Math.PI / 2.75, 0.42, new Vector3(0, 0.01, 0), bs)
    camera.attachControl(canvas, true)
    camera.lowerRadiusLimit = 0.22
    camera.upperRadiusLimit = 1.4
    camera.upperBetaLimit = Math.PI / 2.02 // ما ينزل تحت الطاولة
    camera.wheelPrecision = 220
    camera.panningSensibility = 3200
    camera.minZ = 0.01

    new HemisphericLight('hemi', new Vector3(0, 1, 0), bs).intensity = 0.95
    // الضوء يجي من فوگ وقدّام حتى تبان الأطراف بارزة عن الوجه.
    const dir = new DirectionalLight('dir', new Vector3(-0.35, -1, -0.6), bs)
    dir.position = new Vector3(0.35, 1.1, 0.7)
    dir.intensity = 1.05
    const shadows = new ShadowGenerator(1024, dir)
    shadows.useBlurExponentialShadowMap = true
    shadows.blurScale = 2

    // الطاولة — مرجع بصري للحجم. بدونها الأجهزة تطوف بالفراغ وما تحس بحجمها.
    const table = CreateGround('table', { width: 1.1, height: 0.75 }, bs)
    const tm = new StandardMaterial('m_table', bs)
    tm.diffuseColor = Color3.FromHexString('#1b2231')
    tm.specularColor = new Color3(0.05, 0.05, 0.05)
    table.material = tm
    table.position.y = -0.085
    table.receiveShadows = true
    table.isPickable = false

    hlRef.current = new HighlightLayer('hl', bs)

    // ═══ الأجهزة ═══
    const anchors = new Map<string, TerminalAnchor>()
    for (const sd of scene.devices ?? []) {
      const dev = devices[sd.deviceId]
      if (!dev) continue
      // x بالمشهد نسبة ٠..١ من اليسار — نفس الي يستعمله المنظر المنطقي.
      // بالكاميرا الافتراضية (alpha=+π/2) محور +X يطلع **يمين** الشاشة،
      // فـ`x` بالمشهد ينمشي مثل ما هو — والقفل (0.68) يبقى يمين مثل
      // المنظر المنطقي. (انفحص بالصورة مو بالاشتقاق.)
      const worldX = (sd.x - 0.5) * 2 * TABLE_HALF
      const built = buildDevice(bs, dev, sd.ref, new Vector3(worldX, 0, 0), 0)
      built.root.getChildMeshes().forEach((m) => shadows.addShadowCaster(m as Mesh))
      for (const a of built.anchors) anchors.set(`${sd.ref}:${a.terminal.id}`, a)
    }
    anchorsRef.current = anchors

    // ═══ التفاعل ═══
    bs.onPointerObservable.add((info) => {
      const picked = info.pickInfo?.pickedMesh
      const meta = picked?.metadata as { terminalId?: string; deviceRef?: string; wireIndex?: number } | undefined

      if (info.type === PointerEventTypes.POINTERMOVE) {
        const id = meta?.terminalId ? `${meta.deviceRef}:${meta.terminalId}` : null
        setHover(id ? anchors.get(id) || null : null)
        canvas.style.cursor = meta ? 'pointer' : 'grab'
        return
      }
      if (info.type !== PointerEventTypes.POINTERPICK) return
      if (cb.current.readOnly) return

      // حذف كيبل
      if (meta?.wireIndex !== undefined) {
        const w = wiresRef.current[meta.wireIndex]
        if (w) cb.current.onRemoveWire(w)
        return
      }
      if (!meta?.terminalId) return
      const id = `${meta.deviceRef}:${meta.terminalId}`
      setReject(null)

      // ⚠️ المنطق **برّا** دالة تحديث الحالة. حطّيته جوّاها أول مرة
      // وطلع تحذير React حقيقي: «تحديث Runner أثناء رندر Workbench3D».
      // السبب إن دالة التحديث تنفّذها React **وقت الرندر**، وأي نداء
      // لـ`onAction` جوّاها يحدّث حالة الصفحة الأم بنص الرندر. الحالة
      // السابقة تنقرا من ref، والتحديثات تصير نداءات صريحة.
      const prev = armedRef.current
      if (!prev) { setArmed(id); return }
      if (prev === id) { setArmed(null); return }

      const a = anchors.get(prev)
      const b2 = anchors.get(id)
      if (a && b2) {
        // توافق الموصّلات قبل أي شي (٩) — الرفض هنا **مو غلط تدريبي**
        // ولا ينحسب على المتدرّب: هو استحالة فيزيائية، مثل ما ما تگدر
        // تسنّب RJ45 ببلوك براغي بالميدان.
        const compat = connectorsCompatible(a.terminal, b2.terminal)
        if (!compat.ok) {
          setReject(compat.why || 'ما ينسنّب')
          setArmed(null)
          return
        }
      }
      setArmed(null)
      cb.current.onAction({ op: 'CONNECT', from: prev, to: id })
    })

    // ⚠️ منفذ فحص **بالتطوير بس**: الفحص الآلي لازم يعرف وين يطلع الطرف
    // على الشاشة حتى يضغطه، وحسابه من برّا يعني إعادة كتابة إسقاط
    // الكاميرا بالفحص — فحص يفحص نفسه مو المشهد. `import.meta.env.DEV`
    // يشيله كلياً من بناء الإنتاج.
    if (import.meta.env.DEV) {
      const w = window as unknown as { __simEngine?: Engine; __BJS?: unknown }
      w.__simEngine = engine
      w.__BJS = { Vector3, Matrix }
    }

    engine.runRenderLoop(() => bs.render())
    const onResize = () => engine.resize()
    window.addEventListener('resize', onResize)
    setReady(true)

    return () => {
      window.removeEventListener('resize', onResize)
      engine.stopRenderLoop()
      bs.dispose()
      engine.dispose()
      engineRef.current = null
      sceneRef.current = null
    }
    // المشهد ينبني مرة وحدة للتمرين — تغيّر الأسلاك ينعالج بتأثير منفصل.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  // ═══ رسم الكيابل ═══
  const drawCables = useCallback(() => {
    const bs = sceneRef.current
    if (!bs) return
    for (const m of cableMeshRef.current) m.dispose()
    cableMeshRef.current = []

    wires.forEach((w, i) => {
      const a = anchorsRef.current.get(w.from)
      const b = anchorsRef.current.get(w.to)
      if (!a || !b) return

      // ⚠️ ترهّل الكيبل (catenary): الخط المستقيم بين طرفين يبيّن «رسم
      // برنامج»، والترهّل يبيّن «سلك». ومو زخرفة بس — الطول الفعلي
      // ينحسب من هذا المسار مو من المسافة المستقيمة.
      const dist = Vector3.Distance(a.point, b.point)
      const sag = Math.min(0.055, dist * 0.34)
      const pts: Vector3[] = []
      const N = 26
      for (let s = 0; s <= N; s++) {
        const t = s / N
        const p = Vector3.Lerp(a.point, b.point, t)
        p.y -= Math.sin(Math.PI * t) * sag
        p.z += Math.sin(Math.PI * t) * 0.02 // ينط للأمام حتى ما يدخل بالجهاز
        pts.push(p)
      }
      const tube = CreateTube(`cable_${i}`, { path: pts, radius: 0.0022, tessellation: 10 }, bs)
      const m = new StandardMaterial(`m_cable_${i}`, bs)
      // لون السلك من **الطرف** الي انربط بيه — الفني يتابع اللون بالميدان.
      m.diffuseColor = Color3.FromHexString(a.terminal.colorHex || '#9ca3af')
      m.specularColor = new Color3(0.25, 0.25, 0.25)
      tube.material = m
      tube.metadata = { wireIndex: i }
      tube.isPickable = true
      cableMeshRef.current.push(tube)
    })
  }, [wires])

  useEffect(() => { if (ready) drawCables() }, [ready, drawCables])

  // ═══ التوهّج: الطرف المسلّح والتلميح ═══
  useEffect(() => {
    const hl = hlRef.current
    if (!hl) return
    hl.removeAllMeshes()
    if (armed) {
      const a = anchorsRef.current.get(armed)
      if (a) {
        hl.addMesh(a.post, Color3.FromHexString('#38bdf8'))
        setArmedLabel(a.terminal.label)
      }
    } else {
      setArmedLabel('')
    }
    for (const id of highlight || []) {
      const a = anchorsRef.current.get(id)
      if (a) hl.addMesh(a.post, Color3.FromHexString('#fbbf24'))
    }
  }, [armed, highlight, ready])

  // Esc يلغي التسليح — نفس سلوك المنظر المنطقي.
  useEffect(() => {
    if (!armed) return
    const onKey = (e: KeyboardEvent) => { if (e.key === 'Escape') setArmed(null) }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [armed])

  const gaugeLabel = CABLE_GAUGES.find((g) => g.id === gauge)?.label ?? ''

  return (
    <div className="relative overflow-hidden rounded-2xl bg-[#0e1219] ring-1 ring-slate-700">
      <canvas ref={canvasRef} className="block h-[520px] w-full touch-none outline-none" />

      {/* شريط الحالة — المقطع والطول يأثّرون على الفولتية فلازم يبقون بالعين */}
      <div className="pointer-events-none absolute right-3 top-3 rounded-lg bg-black/55 px-3 py-2 text-[11px] leading-relaxed text-slate-200 backdrop-blur">
        <div>المقطع: <b>{gaugeLabel}</b></div>
        <div>طول التمديد: <b>{runLengthM} م</b></div>
        <div>الأسلاك: <b>{wires.length}</b></div>
      </div>

      {/* بطاقة الطرف تحت المؤشر */}
      {hover && (
        <div className="pointer-events-none absolute bottom-3 right-3 max-w-xs rounded-lg bg-black/70 px-3 py-2 text-xs text-slate-100 backdrop-blur">
          <div className="flex items-center gap-2">
            <span className="inline-block h-3 w-3 rounded-full ring-1 ring-white/40"
              style={{ background: hover.terminal.colorHex || '#9ca3af' }} />
            <b>{hover.terminal.label}</b>
            <span className="text-slate-400">· {hover.terminal.signal}</span>
          </div>
          {hover.terminal.description && <p className="mt-1 text-slate-300">{hover.terminal.description}</p>}
          {hover.terminal.danger && <p className="mt-1 text-red-300">⚠️ {hover.terminal.danger}</p>}
        </div>
      )}

      {/* رفض التوصيل — استحالة فيزيائية، مو غلط تدريبي */}
      {reject && (
        <div className="absolute left-1/2 top-3 -translate-x-1/2 rounded-lg bg-slate-900/90 px-4 py-2 text-xs text-amber-200 ring-1 ring-amber-500/40">
          🔌 {reject}
        </div>
      )}

      {armed && (
        <div className="pointer-events-none absolute bottom-3 left-3 rounded-lg bg-sky-500/90 px-3 py-1.5 text-xs font-bold text-white">
          مسلّح: {armedLabel} — اضغط الطرف الثاني (Esc يلغي)
        </div>
      )}
    </div>
  )
}

export type { Terminal }
