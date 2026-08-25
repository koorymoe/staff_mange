// ═══ المنظر الفيزيائي لمساحة العمل ═══
//
// «أريد محاكاة قوية وثري دي لهاي الأشياء».
//
// ⚠️ **نفس المستند بالضبط** الي يقراه المنظر التخطيطي — مصدر واحد
// للحقيقة (المخطط ١٩). تحرّك قطعة بالمخطط تتحرك هنا، وتضغط قطعة هنا
// تنحدّد هناك. ولا منظر يخزن شي لحاله.
//
// ⚠️ ويعيد استعمال `buildDevice` من محرّك التوصيل **كما هي**: نفس
// المولّد الي يبني القفل وجهاز التتبّع يبني اللوح الشمسي والإنفرتر.
// كتابة مولّد ثانٍ للمشاهد الشمسية يعني شكلين يفترقان بأول تصحيح.

import { useCallback, useEffect, useRef, useState } from 'react'

import '@babylonjs/core/Culling/ray' // ⚠️ إجباري للالتقاط
import { ArcRotateCamera } from '@babylonjs/core/Cameras/arcRotateCamera'
import { Engine } from '@babylonjs/core/Engines/engine'
import { DirectionalLight } from '@babylonjs/core/Lights/directionalLight'
import { HemisphericLight } from '@babylonjs/core/Lights/hemisphericLight'
import { ShadowGenerator } from '@babylonjs/core/Lights/Shadows/shadowGenerator'
import { HighlightLayer } from '@babylonjs/core/Layers/highlightLayer'
import { StandardMaterial } from '@babylonjs/core/Materials/standardMaterial'
import { Color3, Color4 } from '@babylonjs/core/Maths/math.color'
import { Matrix, Vector3 } from '@babylonjs/core/Maths/math.vector'
import { CreateGround } from '@babylonjs/core/Meshes/Builders/groundBuilder'
import { CreateTube } from '@babylonjs/core/Meshes/Builders/tubeBuilder'
import { CreateBox } from '@babylonjs/core/Meshes/Builders/boxBuilder'
import { PointerEventTypes } from '@babylonjs/core/Events/pointerEvents'
import { Scene as BabylonScene } from '@babylonjs/core/scene'
import type { Mesh } from '@babylonjs/core/Meshes/mesh'

import { buildDevice, type TerminalAnchor } from '../sim3d/deviceGeometry'
import type { SimDevice, Terminal } from '../sim/types'
import { PART_BY_ID } from './catalog'
import type { LabDoc, LabNode, PartDef, SimResult } from './types'

interface Props {
  doc: LabDoc
  result: SimResult | null
  selected: string | null
  setSelected: (id: string | null) => void
  /** جهد كل منفذ — الأڤوميتر يطرح بين مسبارين. */
  portV?: Record<string, number>
}

/** ═══ من قطعة اللوح إلى جهاز يفهمه المولّد ═══
 *
 *  ⚠️ محوّل رقيق عمداً: `buildDevice` تشتغل على `SimDevice`، وقطعة
 *  اللوح `PartDef`. المحوّل يترجم بينهما بدل ما ننسخ المولّد.
 *
 *  ⚠️ والقطعة بلا `geo3d` تاخذ صندوقاً محسوباً من أبعادها التخطيطية
 *  (١٠٠ وحدة لوح ≈ ٢٠ سم) — فإضافة قطعة جديدة ما تكسر المشهد، تطلع
 *  صندوقاً بحجم معقول لحد ما تنكتب هندستها.
 */
function toDevice(part: PartDef, node: LabNode): SimDevice {
  const g = part.geo3d
  const sizeM = g?.sizeM ?? { w: (part.w / 100) * 0.2, h: (part.h / 100) * 0.2, d: 0.06 }
  return {
    id: part.id,
    categoryId: '',
    brand: '',
    model: String(node.params.name ?? node.params.hostname ?? part.name),
    name: part.name,
    engineKind: 'WIRING',
    spec: {},
    // منافذ اللوح (نسبة ٠..١ من الرمز) تصير أطرافاً بنفس النسبة على الوجه
    terminals: part.ports.map<Terminal>((p) => ({
      id: p.id,
      label: p.label,
      colorHex: p.polarity === 'pos' ? '#dc2626' : p.polarity === 'neg' ? '#111827'
        : p.kind === 'ac' ? '#f59e0b' : p.kind === 'spk' ? '#a855f7' : '#0ea5e9',
      kind: p.kind.toUpperCase(),
      x: p.x, y: p.y,
    })),
    ui: {},
    geometry: {
      sizeM,
      bodyColorHex: g?.bodyColorHex ?? '#3f4756',
      faceColorHex: g?.faceColorHex ?? '#232a36',
      terminalPost: { radiusM: Math.min(0.006, sizeM.w * 0.03), heightM: 0.006 },
      features: (g?.features ?? []) as never,
    },
    status: 'DRAFT',
    version: 1,
    verified: false,
  }
}

/** لون الكيبل من قطبية منفذه — أحمر/أسود للمستمر وأصفر للمتناوب.
 *  ⚠️ نفس اصطلاح الميدان: الفني يتتبّع اللون مو التسمية. */
function cableColor(part: PartDef | undefined, portId: string): string {
  const p = part?.ports.find((x) => x.id === portId)
  if (!p) return '#94a3b8'
  if (p.kind === 'ac') return '#f59e0b'
  if (p.kind === 'eth' || p.kind === 'sfp') return '#38bdf8'
  if (p.kind === 'spk') return '#a855f7'
  return p.polarity === 'pos' ? '#dc2626' : p.polarity === 'neg' ? '#1f2937' : '#64748b'
}

/** مفتاح الألوان لكل مجال — الفني يتتبّع اللون مو التسمية. */
const LEGEND: Record<string, [string, string][]> = {
  solar: [['#dc2626', 'موجب DC'], ['#1f2937', 'سالب DC'], ['#f59e0b', 'تيار متناوب']],
  electrical: [['#dc2626', 'موجب'], ['#1f2937', 'سالب'], ['#64748b', 'بلا قطبية']],
  network: [['#0ea5e9', 'شبكة RJ45'], ['#38bdf8', 'قفص SFP']],
  fire: [['#64748b', 'خط زون'], ['#f59e0b', 'دائرة إنذار'], ['#dc2626', 'تغذية']],
  audio: [['#a855f7', 'خط سماعات'], ['#f59e0b', 'إشارة ميكروفون']],
}

/** وحدات اللوح ← أمتار. اللوح مساحة رسم، والمشهد عالم حقيقي.
 *
 *  ⚠️ الرقم انضبط **بالصورة مو بالحساب**: بـ٠٫٠١٢ الأجهزة تنباعد
 *  سبعة أمتار واللوح الشمسي ٢٫٢٨ متر — فاللوح يهيمن والباقي يطلع
 *  نقاطاً. المسافات بالمنظومة الحقيقية أقرب من هيچ: الإنفرتر
 *  والبطاريات بغرفة وحدة. */
const M_PER_UNIT = 0.0055

export default function Scene3D({ doc, result, selected, setSelected, portV }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const sceneRef = useRef<BabylonScene | null>(null)
  const anchorsRef = useRef<Map<string, TerminalAnchor>>(new Map())
  const rootsRef = useRef<Map<string, Mesh[]>>(new Map())
  const cablesRef = useRef<Mesh[]>([])
  /** أعلى كل قطعة بالفضاء العالمي — منها تنموضع اللافتة. */
  const topsRef = useRef<Map<string, Vector3>>(new Map())
  const labelsRef = useRef<HTMLDivElement>(null)
  const labelEls = useRef<Map<string, HTMLDivElement>>(new Map())
  const hlRef = useRef<HighlightLayer | null>(null)
  const [ready, setReady] = useState(false)
  const [hover, setHover] = useState<string | null>(null)
  /** ═══ الأڤوميتر ═══
   *  ⚠️ مسباران مثل الجهاز الحقيقي: الأول أحمر والثاني أسود، والقراءة
   *  **فرق** بينهما. أداة تعطي «جهد نقطة» تعلّم عادة ما توجد بالميدان. */
  const [meter, setMeter] = useState(false)
  const [probes, setProbes] = useState<string[]>([])

  const cb = useRef({ setSelected, meter: false })
  useEffect(() => { cb.current = { setSelected, meter } })

  // ═══ بناء المشهد ═══
  //
  // ⚠️ يعاد بناؤه لمن **تتغيّر مجموعة القطع أو الوصلات** — مو مع كل
  // تعديل خاصية. إعادة البناء مع كل ضغطة مفتاح بخانة رقمية تخلّي
  // المشهد يترمّش، والكاميرا ترجع لمحلها الأول.
  const topology = doc.nodes.map((n) => `${n.id}:${n.partId}:${n.x},${n.y}`).join('|') +
    '#' + doc.links.map((l) => `${l.from.node}.${l.from.port}-${l.to.node}.${l.to.port}`).join('|')

  useEffect(() => {
    const canvas = canvasRef.current
    if (!canvas) return
    const engine = new Engine(canvas, true, { preserveDrawingBuffer: true, stencil: true })
    const bs = new BabylonScene(engine)
    // ⚠️ اليد اليمنى — نفس سبب مشهد التوصيل: بلاها كل شي ينقلب مرآة.
    bs.useRightHandedSystem = true
    bs.clearColor = new Color4(0.03, 0.05, 0.09, 1)
    sceneRef.current = bs

    const camera = new ArcRotateCamera('cam', Math.PI / 2.4, Math.PI / 3.4, 6, new Vector3(0, 0.4, 0), bs)
    camera.attachControl(canvas, true)
    camera.lowerRadiusLimit = 1.2
    camera.upperRadiusLimit = 24
    camera.upperBetaLimit = Math.PI / 2.05
    camera.wheelPrecision = 18
    camera.panningSensibility = 120
    camera.minZ = 0.05

    new HemisphericLight('hemi', new Vector3(0, 1, 0), bs).intensity = 0.85
    // ⚠️ الشمس تجي من زاوية عالية: الألواح مائلة، وضوء عمودي يخلّيها
    // تبين مسطّحة بلا عمق.
    const sun = new DirectionalLight('sun', new Vector3(-0.45, -1, -0.35), bs)
    sun.position = new Vector3(6, 12, 5)
    sun.intensity = 1.15
    const shadows = new ShadowGenerator(1024, sun)
    shadows.useBlurExponentialShadowMap = true
    shadows.blurScale = 2

    const ground = CreateGround('ground', { width: 24, height: 24 }, bs)
    const gm = new StandardMaterial('m_ground', bs)
    gm.diffuseColor = Color3.FromHexString('#0d1522')
    gm.specularColor = new Color3(0.02, 0.02, 0.02)
    ground.material = gm
    ground.receiveShadows = true
    ground.isPickable = false

    hlRef.current = new HighlightLayer('hl3d', bs)

    // ═══ القطع ═══
    const anchors = new Map<string, TerminalAnchor>()
    const roots = new Map<string, Mesh[]>()
    const tops = new Map<string, Vector3>()
    let minX = Infinity, maxX = -Infinity, minZ = Infinity, maxZ = -Infinity

    for (const n of doc.nodes) {
      const part = PART_BY_ID[n.partId]
      if (!part) continue
      const dev = toDevice(part, n)
      const size = dev.geometry?.sizeM ?? { w: 0.2, h: 0.2, d: 0.06 }
      // موقع اللوح (x يمين، y نزول) ← أرضية المشهد (X, Z)
      const wx = (n.x + part.w / 2) * M_PER_UNIT
      const wz = (n.y + part.h / 2) * M_PER_UNIT
      minX = Math.min(minX, wx); maxX = Math.max(maxX, wx)
      minZ = Math.min(minZ, wz); maxZ = Math.max(maxZ, wz)

      const built = buildDevice(bs, dev, n.id, new Vector3(wx, size.h / 2 + 0.02, wz), 0)

      // ⚠️ الميلان **بعد** البناء: `buildDevice` تبني بالمستوي الرأسي،
      // واللوح الشمسي ينصب مائلاً. الميلان على الجذر يشيل الأطراف
      // وياه، ونقاط السنّ تنعاد حسابها تحت.
      const tilt = part.geo3d?.tiltDeg
      if (tilt) {
        built.root.rotation.x = -(90 - tilt) * (Math.PI / 180)
        built.root.position.y = size.h * Math.sin((tilt * Math.PI) / 180) / 2 + 0.35
        // قائم يسند اللوح — بدونه يطوف بالهوا.
        const leg = CreateBox(`leg_${n.id}`, { width: 0.05, height: 0.7, depth: 0.05 }, bs)
        leg.material = new StandardMaterial(`m_leg_${n.id}`, bs)
        ;(leg.material as StandardMaterial).diffuseColor = Color3.FromHexString('#475569')
        leg.position = new Vector3(wx, 0.35, wz + size.h * 0.25)
        leg.isPickable = false
        shadows.addShadowCaster(leg)
      }

      built.root.computeWorldMatrix(true)
      // ⚠️ نقطة اللافتة **فوگ** الجسم مو بمركزه: لافتة بالمركز تغطّي
      // القطعة نفسها، والفني يقرا الرقم وما يشوف الي يقيسه.
      tops.set(n.id, new Vector3(
        built.root.position.x,
        built.root.position.y + (tilt ? size.h * Math.sin((tilt * Math.PI) / 180) : size.h) / 2 + 0.16,
        built.root.position.z,
      ))
      const meshes = built.root.getChildMeshes() as Mesh[]
      for (const m of meshes) { shadows.addShadowCaster(m); m.computeWorldMatrix(true) }
      roots.set(n.id, meshes)

      for (const a of built.anchors) {
        a.post.computeWorldMatrix(true)
        a.point = a.post.getAbsolutePosition().clone()
        anchors.set(`${n.id}:${a.terminal.id}`, a)
        // ⚠️ كل جسم بالقطعة يحمل معرّفها: الضغط على جسم اللوح نفسه
        // لازم يحدّدها، مو بس الضغط على طرف.
      }
      for (const m of meshes) {
        m.isPickable = true
        m.metadata = { ...(m.metadata ?? {}), nodeId: n.id }
      }
    }
    anchorsRef.current = anchors
    rootsRef.current = roots
    topsRef.current = tops

    // ═══ الكيابل ═══
    for (const l of doc.links) {
      const a = anchors.get(`${l.from.node}:${l.from.port}`)
      const b = anchors.get(`${l.to.node}:${l.to.port}`)
      if (!a || !b) continue
      const dist = Vector3.Distance(a.point, b.point)
      const sag = Math.min(0.22, dist * 0.16)
      const pts: Vector3[] = []
      for (let i = 0; i <= 24; i++) {
        const t = i / 24
        const p = Vector3.Lerp(a.point, b.point, t)
        p.y -= Math.sin(Math.PI * t) * sag
        pts.push(p)
      }
      const tube = CreateTube(`cab_${l.id}`, { path: pts, radius: 0.016, tessellation: 8 }, bs)
      const tm = new StandardMaterial(`m_cab_${l.id}`, bs)
      tm.diffuseColor = Color3.FromHexString(cableColor(PART_BY_ID[doc.nodes.find((n) => n.id === l.from.node)?.partId ?? ''], l.from.port))
      tm.specularColor = new Color3(0.2, 0.2, 0.2)
      tube.material = tm
      tube.isPickable = false
      shadows.addShadowCaster(tube)
      cablesRef.current.push(tube)
    }

    // ═══ ضبط الكاميرا على المحتوى ═══
    if (Number.isFinite(minX)) {
      const cx = (minX + maxX) / 2, cz = (minZ + maxZ) / 2
      const span = Math.max(maxX - minX, maxZ - minZ, 2)
      camera.setTarget(new Vector3(cx, 0.55, cz))
      camera.radius = Math.max(3.2, span * 1.15 + 2.2)
    }

    // ═══ التفاعل ═══
    bs.onPointerObservable.add((info) => {
      const meta = info.pickInfo?.pickedMesh?.metadata as { nodeId?: string } | undefined
      if (info.type === PointerEventTypes.POINTERMOVE) {
        setHover(meta?.nodeId ?? null)
        canvas.style.cursor = meta?.nodeId ? 'pointer' : 'grab'
        return
      }
      // ⚠️ `POINTERTAP` مو `POINTERPICK`: الثاني ما ينطلق إلا لمن
      // ينضغط **جسم**، فالضغط على الفراغ ما يلغي التحديد أبداً —
      // والفني يضل يدوّر شلون يلغيه. و`TAP` ما ينطلق بعد سحب
      // الكاميرا، فتدوير المشهد ما يلغي تحديدك.
      if (info.type !== PointerEventTypes.POINTERTAP) return

      // ⚠️ بوضع القياس الضغط **يلمس طرفاً** مو يحدّد قطعة: خلط
      // الوضعين يعني كل لمسة مسبار تبدّل لوحة الخصائص وتشتّت الفني.
      if (cb.current.meter) {
        const t = info.pickInfo?.pickedMesh?.metadata as { terminalId?: string; deviceRef?: string } | undefined
        if (t?.terminalId) {
          const id = `${t.deviceRef}:${t.terminalId}`
          setProbes((prev) => (prev.length >= 2 ? [id] : [...prev, id]))
        }
        return
      }
      cb.current.setSelected(meta?.nodeId ?? null)
    })

    // ═══ موضعة اللافتات ═══
    //
    // ⚠️ تنكتب على **DOM مباشرة** مو بحالة React: الموضعة تصير مع كل
    // رسمة (٦٠ مرة بالثانية وأنت تدوّر الكاميرا)، ورفعها لحالة React
    // يعني ٦٠ رندراً بالثانية للاستوديو كله — الصفحة تختنق.
    //
    // ⚠️ ولافتة **ورا الكاميرا** تنخفى: الإسقاط يرجّع إحداثيات صالحة
    // حتى للي ورا الظهر، فبلا هالفحص تطلع لافتات مقلوبة بالزوايا.
    bs.onAfterRenderObservable.add(() => {
      const w = engine.getRenderWidth(), h = engine.getRenderHeight()
      const vp = bs.activeCamera?.viewport.toGlobal(w, h)
      if (!vp) return
      const rect = canvas.getBoundingClientRect()
      const sx = rect.width / w, sy = rect.height / h
      for (const [id, el] of labelEls.current) {
        const pos = topsRef.current.get(id)
        if (!pos) { el.style.display = 'none'; continue }
        const q = Vector3.Project(pos, Matrix.Identity(), bs.getTransformMatrix(), vp)
        if (q.z > 1 || q.z < 0) { el.style.display = 'none'; continue }
        el.style.display = ''
        el.style.transform = `translate(-50%,-100%) translate(${q.x * sx}px, ${q.y * sy}px)`
      }
    })

    engine.runRenderLoop(() => bs.render())
    const onResize = () => engine.resize()
    window.addEventListener('resize', onResize)
    setReady(true)

    if (import.meta.env.DEV) {
      ;(window as unknown as { __labScene?: Engine }).__labScene = engine
    }

    return () => {
      window.removeEventListener('resize', onResize)
      engine.stopRenderLoop()
      bs.dispose()
      engine.dispose()
      sceneRef.current = null
      cablesRef.current = []
      setReady(false)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [topology])

  // ═══ اللافتات ═══
  //
  // ⚠️ تنبنى لمن **تتغيّر النتيجة** بس — مو مع كل رسمة. الموضعة
  // شي والمحتوى شي: المحتوى يتغيّر لمن تشغّل المحاكاة، والموضعة
  // تتغيّر لمن تدوّر الكاميرا.
  useEffect(() => {
    const host = labelsRef.current
    if (!host) return
    host.replaceChildren()
    labelEls.current.clear()
    if (!result) return

    for (const n of doc.nodes) {
      const rs = result.nodeReadings[n.id] ?? []
      if (rs.length === 0) continue
      const el = document.createElement('div')
      el.className = 'absolute left-0 top-0 rounded-lg bg-black/70 px-2 py-1 text-center leading-tight backdrop-blur-sm ring-1 ring-white/10'
      el.style.willChange = 'transform'
      const name = document.createElement('div')
      name.className = 'text-[10px] font-bold text-slate-300'
      name.textContent = String(n.params.name ?? n.params.hostname ?? PART_BY_ID[n.partId]?.name ?? '')
      el.appendChild(name)
      for (const r of rs.slice(0, 3)) {
        const v = document.createElement('div')
        v.className = `font-mono text-[11.5px] font-bold ${
          r.tone === 'bad' ? 'text-red-300' : r.tone === 'warn' ? 'text-amber-300' : 'text-emerald-300'}`
        v.textContent = r.text
        el.appendChild(v)
      }
      host.appendChild(el)
      labelEls.current.set(n.id, el)
    }
  }, [result, doc.nodes, ready])

  // ═══ التمييز ═══
  const highlight = useCallback(() => {
    const hl = hlRef.current
    if (!hl) return
    hl.removeAllMeshes()
    for (const [id, color] of [[selected, '#38bdf8'], [hover, '#94a3b8']] as const) {
      if (!id || (id === hover && hover === selected)) continue
      for (const m of rootsRef.current.get(id) ?? []) hl.addMesh(m, Color3.FromHexString(color))
    }
    // المسبارات: الأول أحمر والثاني أسود — نفس ألوان الجهاز الحقيقي.
    probes.forEach((pid, i) => {
      const a = anchorsRef.current.get(pid)
      if (a) hl.addMesh(a.post, Color3.FromHexString(i === 0 ? '#ef4444' : '#e2e8f0'))
    })
  }, [selected, hover, probes])
  useEffect(() => { if (ready) highlight() }, [ready, highlight])

  // ⚠️ القراءة **فرق** بين المسبارين — والسالب يعني إنك عاكس
  // المسبارين، مثل الجهاز الحقيقي بالضبط.
  const reading = probes.length === 2 && portV
    ? (portV[probes[0]] ?? 0) - (portV[probes[1]] ?? 0)
    : null
  const probeLabel = (i: number) => {
    const id = probes[i]
    if (!id) return ''
    const [nodeId, portId] = id.split(':')
    const nd = doc.nodes.find((n) => n.id === nodeId)
    const pt = PART_BY_ID[nd?.partId ?? '']?.ports.find((x) => x.id === portId)
    return pt?.label ?? portId
  }

  const hasVoltages = !!portV && Object.keys(portV).length > 0

  const sel = doc.nodes.find((n) => n.id === (hover ?? selected))
  const selPart = sel ? PART_BY_ID[sel.partId] : null
  const readings = sel ? result?.nodeReadings[sel.id] ?? [] : []

  return (
    <div className="relative h-full w-full overflow-hidden rounded-xl bg-[#070c14] ring-1 ring-slate-800">
      <canvas ref={canvasRef} className="block h-full w-full touch-none outline-none" />

      {/* ⚠️ `pointer-events-none` على الحاوية: اللافتات تطوف فوگ اللوح،
          وبدونها تحجب الضغط على القطع الي تحتها. */}
      <div ref={labelsRef} className="pointer-events-none absolute inset-0 overflow-hidden" />

      {doc.nodes.length === 0 && (
        <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
          <p className="text-[12px] text-slate-600">اللوح فاضي — حط قطعاً من المكتبة.</p>
        </div>
      )}

      {/* بطاقة القطعة تحت المؤشر */}
      {sel && selPart && (
        <div className="pointer-events-none absolute bottom-3 right-3 max-w-xs rounded-xl bg-black/70 px-3 py-2 text-xs text-slate-100 backdrop-blur">
          <b>{String(sel.params.name ?? sel.params.hostname ?? selPart.name)}</b>
          {selPart.model && <span className="block text-[10px] text-slate-400">{selPart.model}</span>}
          {readings.map((r, i) => (
            <span key={i} className={`mt-0.5 block font-mono text-[11px] ${
              r.tone === 'bad' ? 'text-red-300' : r.tone === 'warn' ? 'text-amber-300' : 'text-emerald-300'}`}>
              {r.text}
            </span>
          ))}
        </div>
      )}

      {/* الأڤوميتر — ما يطلع إلا بمجال عنده جهود محسوبة.
          ⚠️ أداة قياس على مشهد شبكات تقرا «0.0 V» على منفذ RJ45 —
          رقم غلط، والرقم الغلط أسوأ من ما اكو رقم. */}
      {hasVoltages && (
      <div className="absolute right-3 top-3 w-56">
        <button
          onClick={() => { setMeter((v) => !v); setProbes([]) }}
          className={`w-full rounded-lg px-3 py-1.5 text-[11.5px] font-bold transition ${
            meter ? 'bg-amber-500 text-black' : 'bg-black/60 text-slate-300 ring-1 ring-slate-700 backdrop-blur hover:bg-black/80'}`}
        >
          🔌 {meter ? 'أطفِ الأڤوميتر' : 'أڤوميتر'}
        </button>

        {meter && (
          <div className="mt-1.5 rounded-xl bg-black/80 p-3 text-center ring-1 ring-amber-500/40 backdrop-blur">
            <p className="font-mono text-2xl font-black tabular-nums text-amber-300">
              {reading === null ? '— — —' : `${reading.toFixed(1)} V`}
            </p>
            <p className="mt-1 text-[10.5px] leading-relaxed text-slate-400">
              {probes.length === 0 && 'المس الطرف الأول (مسبار أحمر)'}
              {probes.length === 1 && 'المس الطرف الثاني (مسبار أسود)'}
              {probes.length === 2 && (
                <>
                  {probeLabel(0)} ↔ {probeLabel(1)}
                  <button onClick={() => setProbes([])} className="mt-1 block w-full text-amber-400 underline">
                    قياس جديد
                  </button>
                </>
              )}
            </p>
          </div>
        )}
      </div>
      )}

      {/* مفتاح ألوان الأسلاك — حسب المجال.
          ⚠️ مفتاح «موجب DC / سالب DC» على مشهد شبكات ما إله معنى —
          الكيابل هناك نحاس وألياف. مفتاح ما يخص يشوّش مو يوضّح. */}
      <div className="pointer-events-none absolute bottom-3 left-3 rounded-lg bg-black/55 px-2.5 py-2 text-[10px] text-slate-300 backdrop-blur">
        {(LEGEND[doc.domain] ?? LEGEND.electrical).map(([c, t]) => (
          <span key={t} className="flex items-center gap-1.5">
            <span className="inline-block h-1.5 w-4 rounded" style={{ background: c }} />
            {t}
          </span>
        ))}
      </div>
    </div>
  )
}
