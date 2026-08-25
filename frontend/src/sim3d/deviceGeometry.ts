// ═══ مولّد الأجسام ثلاثية الأبعاد ═══
//
// المخطط الرئيسي (٣) يعتبر الجهاز **توأماً رقمياً** والشكل ثلاثي الأبعاد
// مجرد View لنفس الكائن. فهذا الملف ما يعرّف ولا جهاز — ياخذ `geometry`
// الجايّة من السيرفر ويطلع منها جسماً.
//
// ⚠️ ماكو ملف موديل مصنّع ولا فنان ٣د: الأجسام **تتولّد بالكود**. وهذا
// مقصود مو اختصار — المخطط (٧٫٣) يكول الي يهم فعلاً هو **المراسي
// الدلالية** مو دقة الشكل: بدون ما يعرف المحرّك وين `port.dc_plus` ما
// يقدر يعرف وين ينسنّب السلك. فنبني المراسي أول، والشكل الحقيقي
// ينركّب فوگها بعدين بلا إعادة عمل.
//
// ⚠️ الوحدة **متر** (٧٫٢). جهاز عرضه ١٤٫٥ سم يطلع 0.145 بالمشهد — فلمن
// يجي جهاز ثاني بجنبه، الأحجام تنقارن صح بلا معايرة.

import { Color3 } from '@babylonjs/core/Maths/math.color'
import { Vector3 } from '@babylonjs/core/Maths/math.vector'
import { CreateBox } from '@babylonjs/core/Meshes/Builders/boxBuilder'
import { CreateCylinder } from '@babylonjs/core/Meshes/Builders/cylinderBuilder'
import { StandardMaterial } from '@babylonjs/core/Materials/standardMaterial'
import { DynamicTexture } from '@babylonjs/core/Materials/Textures/dynamicTexture'
import { TransformNode } from '@babylonjs/core/Meshes/transformNode'
import type { Mesh } from '@babylonjs/core/Meshes/mesh'
import type { Scene } from '@babylonjs/core/scene'

import type { DeviceGeometry, SimDevice, Terminal } from '../sim/types'

/** المقاسات الافتراضية لجهاز ما عنده هندسة بعد — حتى ما ينهار المشهد. */
const FALLBACK: Required<Pick<DeviceGeometry, 'sizeM' | 'bodyColorHex' | 'faceColorHex'>> = {
  sizeM: { w: 0.12, h: 0.12, d: 0.03 },
  bodyColorHex: '#3f4756',
  faceColorHex: '#232a36',
}

/** ما يلزم المحرّك يعرفه عن طرف بعد ما ينبني بالمشهد. */
export interface TerminalAnchor {
  terminal: Terminal
  deviceRef: string
  /** «مرساة دلالية» — الاسم الثابت الي يربط الشكل بالمعنى (٧٫٣). */
  semantic: string
  /** نقطة سنّ السلك بالفضاء العالمي. */
  point: Vector3
  post: Mesh
}

export interface BuiltDevice {
  root: TransformNode
  anchors: TerminalAnchor[]
  /** يضوّي/يطفّي LED الحالة — يستعمله المحرّك لمن تتغيّر التغذية. */
  setLed: (color: Color3 | null) => void
}

const c3 = (hex: string | undefined, fallback: string) => Color3.FromHexString(hex || fallback)

function mat(scene: Scene, name: string, hex: string, opts?: { emissive?: Color3; spec?: number }) {
  const m = new StandardMaterial(name, scene)
  m.diffuseColor = c3(hex, '#888888')
  m.specularColor = new Color3(opts?.spec ?? 0.12, opts?.spec ?? 0.12, opts?.spec ?? 0.12)
  if (opts?.emissive) m.emissiveColor = opts.emissive
  return m
}

/** لوحة نص تنرسم على texture — أسماء الأطراف والموديل تنقرا بالمشهد.
 *  ⚠️ RTL: النص العربي ينرسم بـ`direction: rtl` وإلا ينقلب ترتيب الكلمات. */
function labelPlate(scene: Scene, text: string, wPx: number, hPx: number, bg: string, fg: string) {
  const tex = new DynamicTexture(`lbl_${text}`, { width: wPx, height: hPx }, scene, false)
  const ctx = tex.getContext() as CanvasRenderingContext2D
  ctx.fillStyle = bg
  ctx.fillRect(0, 0, wPx, hPx)
  ctx.direction = 'rtl'
  ctx.fillStyle = fg
  ctx.font = `bold ${Math.floor(hPx * 0.55)}px "Segoe UI", Tahoma, sans-serif`
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  ctx.fillText(text, wPx / 2, hPx / 2)
  tex.update()
  // ⚠️ قلب **عمودي** للنسيج (V مو U). باليد اليمنى الهندسة صحيحة —
  // الأطراف يسار والأزرار ١٢٣ بمحلاتها — بس الـDynamicTexture تنرسم
  // بمحور V مقلوب فالحروف تطلع مقلوبة. القلب هنا أنظف من الرسم
  // بالمقلوب على الـcanvas: ذاك يكسر تشكيل العربي.
  //
  // ⚠️ جرّبت U أول وطلع الترتيب صح والحروف مقلوبة — القلب الغلط
  // يخفي نص المشكلة ويخلّيها تبين «شبه صحيحة». انفحص بالصورة.
  tex.vScale = -1
  tex.vOffset = 1
  return tex
}

/**
 * يبني جهازاً كاملاً بالمشهد.
 *
 * الطرف (x,y) نسبة ٠..١ من **وجه** الجهاز — نفس الأرقام الي يستعملها
 * المنظر المنطقي ثنائي الأبعاد. منظر واحد ما يقدر يزيح الثاني (١٩).
 */
export function buildDevice(
  scene: Scene,
  device: SimDevice,
  deviceRef: string,
  position: Vector3,
  rotationY: number,
): BuiltDevice {
  const g = device.geometry || {}
  const size = g.sizeM || FALLBACK.sizeM
  const { w, h, d } = size
  const root = new TransformNode(`dev_${deviceRef}`, scene)
  root.position = position
  root.rotation.y = rotationY

  // ═══ الجسم ═══
  const body = CreateBox(`body_${deviceRef}`, { width: w, height: h, depth: d }, scene)
  body.material = mat(scene, `m_body_${deviceRef}`, g.bodyColorHex || FALLBACK.bodyColorHex)
  body.parent = root
  body.isPickable = false

  // الوجه الأمامي: لوح رفيع بلون أغمق حتى يبان الفرق بين الوجه والجوانب.
  const faceZ = d / 2 + 0.0006
  const face = CreateBox(`face_${deviceRef}`, { width: w * 0.985, height: h * 0.985, depth: 0.001 }, scene)
  face.material = mat(scene, `m_face_${deviceRef}`, g.faceColorHex || FALLBACK.faceColorHex)
  face.position = new Vector3(0, 0, faceZ)
  face.parent = root
  face.isPickable = false

  // ⚠️ (x,y) بالمواصفة: x من **اليسار** وy من **الأعلى** — نفس اصطلاح
  // المنظر ثنائي الأبعاد. المحور Y بالمشهد يزيد لفوگ، فينقلب.
  const onFace = (x: number, y: number, z = faceZ) => new Vector3((x - 0.5) * w, (0.5 - y) * h, z)

  // ═══ الملامح ═══
  let ledMesh: Mesh | null = null
  for (const f of g.features || []) {
    if (f.kind === 'terminalPlate') {
      const pw = (f.x1 - f.x0) * w
      const ph = (f.y1 - f.y0) * h
      const plate = CreateBox(`plate_${deviceRef}`, { width: pw, height: ph, depth: 0.0015 }, scene)
      plate.material = mat(scene, `m_plate_${deviceRef}`, '#0f141d')
      plate.position = onFace((f.x0 + f.x1) / 2, (f.y0 + f.y1) / 2, faceZ + 0.0008)
      plate.parent = root
      plate.isPickable = false
    } else if (f.kind === 'keypad') {
      const kw = f.w * w
      const kh = f.h * h
      const pad = CreateBox(`pad_${deviceRef}`, { width: kw, height: kh, depth: 0.0015 }, scene)
      pad.material = mat(scene, `m_pad_${deviceRef}`, '#161c27')
      pad.position = onFace(f.x, f.y, faceZ + 0.0008)
      pad.parent = root
      pad.isPickable = false
      // الأزرار — تنبني من cols/rows مو مكتوبة وحدة وحدة.
      const keys = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '*', '0', '#']
      const bw = (kw / f.cols) * 0.74
      const bh = (kh / f.rows) * 0.72
      const keyMat = mat(scene, `m_key_${deviceRef}`, '#cbd2dd')
      for (let r = 0; r < f.rows; r++) {
        for (let c = 0; c < f.cols; c++) {
          const i = r * f.cols + c
          const b = CreateBox(`key_${deviceRef}_${i}`, { width: bw, height: bh, depth: 0.0022 }, scene)
          b.material = keyMat
          b.position = new Vector3(
            pad.position.x + (c - (f.cols - 1) / 2) * (kw / f.cols),
            pad.position.y - (r - (f.rows - 1) / 2) * (kh / f.rows),
            faceZ + 0.0022,
          )
          b.parent = root
          b.isPickable = false
          if (keys[i]) {
            const t = CreateBox(`keyt_${deviceRef}_${i}`, { width: bw, height: bh, depth: 0.0004 }, scene)
            const tm = new StandardMaterial(`m_keyt_${deviceRef}_${i}`, scene)
            tm.diffuseTexture = labelPlate(scene, keys[i], 64, 64, '#cbd2dd', '#1f2937')
            tm.specularColor = Color3.Black()
            t.material = tm
            t.position = new Vector3(b.position.x, b.position.y, faceZ + 0.0034)
            t.parent = root
            t.isPickable = false
          }
        }
      }
    } else if (f.kind === 'statusLed') {
      const led = CreateCylinder(`led_${deviceRef}`, { diameter: 0.006, height: 0.002 }, scene)
      led.rotation.x = Math.PI / 2
      led.position = onFace(f.x, f.y, faceZ + 0.001)
      led.material = mat(scene, `m_led_${deviceRef}`, '#334155')
      led.parent = root
      led.isPickable = false
      ledMesh = led
    }
  }

  // ═══ لوحة الموديل ═══ حتى يعرف المتدرّب شنو الجهاز الي يشتغل عليه.
  const nameW = w * 0.9
  const nameH = h * 0.1
  const plate = CreateBox(`nameplate_${deviceRef}`, { width: nameW, height: nameH, depth: 0.0006 }, scene)
  const nm = new StandardMaterial(`m_name_${deviceRef}`, scene)
  nm.diffuseTexture = labelPlate(scene, device.model, 512, 64, '#111827', '#94a3b8')
  nm.specularColor = Color3.Black()
  plate.material = nm
  plate.position = new Vector3(0, -h / 2 + nameH * 0.75, faceZ + 0.0012)
  plate.parent = root
  plate.isPickable = false

  // ═══ الأطراف — المراسي الدلالية ═══
  //
  // ⚠️ كل طرف **جسم مستقل قابل للالتقاط**. هذا الي يخلّي التوصيل بالفضاء
  // الثلاثي بلا أي حساب إحداثيات: تضغط الجسم، Babylon يكلك أي واحد
  // انضغط، والمعنى مخزون بـmetadata مو مستنتج من الموقع.
  const postSpec = g.terminalPost || { radiusM: 0.0035, heightM: 0.0055 }
  const anchors: TerminalAnchor[] = []
  for (const t of device.terminals) {
    const post = CreateCylinder(`term_${deviceRef}_${t.id}`, {
      diameter: postSpec.radiusM * 2,
      height: postSpec.heightM,
    }, scene)
    post.rotation.x = Math.PI / 2
    post.position = onFace(t.x, t.y, faceZ + postSpec.heightM / 2)
    post.material = mat(scene, `m_term_${deviceRef}_${t.id}`, t.colorHex || '#9ca3af', { spec: 0.35 })
    post.parent = root
    post.isPickable = true
    post.metadata = {
      semantic: `port.${t.id}`,
      terminalId: t.id,
      deviceRef,
    }
    anchors.push({
      terminal: t,
      deviceRef,
      semantic: `port.${t.id}`,
      point: Vector3.Zero(), // تنحسب بعد ما ينثبّت المشهد — تحت.
      post,
    })
  }

  // ⚠️ نقطة السنّ لازم تكون **عالمية** مو محلية: الجهاز مدوّر ومزاح،
  // والسلك ينرسم بفضاء المشهد. `computeWorldMatrix(true)` إجباري لأن
  // Babylon ما يحدّث المصفوفة إلا عند الرسم — وإحنه نحتاجها قبله.
  root.computeWorldMatrix(true)
  for (const a of anchors) {
    a.post.computeWorldMatrix(true)
    a.point = a.post.getAbsolutePosition().clone()
  }

  const setLed = (color: Color3 | null) => {
    if (!ledMesh) return
    const m = ledMesh.material as StandardMaterial
    m.emissiveColor = color || Color3.Black()
    m.diffuseColor = color || Color3.FromHexString('#334155')
  }
  setLed(null)

  return { root, anchors, setLed }
}
