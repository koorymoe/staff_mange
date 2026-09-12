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

/** نسيج خلايا اللوح — شبكة خطوط على أزرق داكن.
 *
 *  ⚠️ ينبني **مرة وحدة لكل شكل** ويتشارك بين كل الألواح: نسيج لكل
 *  لوح يعني ذاكرة GPU تتضاعف مع كل لوح ينضاف للمشهد. */
const pvTexCache = new Map<string, DynamicTexture>()
function pvTexture(scene: Scene, cols: number, rows: number): DynamicTexture {
  const key = `${cols}x${rows}`
  // ⚠️ المفتاح يشمل **المشهد**: المشهد ينهدم لمن تسكّر الصفحة،
  // والنسيج ينهدم وياه — فنسيج مخبّأ من مشهد ميت يعطي لوحاً أسود.
  const hit = pvTexCache.get(key)
  if (hit && hit.getScene() === scene) return hit
  const W = 256, H = 512
  const tex = new DynamicTexture(`pv_${key}`, { width: W, height: H }, scene, false)
  const ctx = tex.getContext() as CanvasRenderingContext2D
  ctx.fillStyle = '#0f2744'
  ctx.fillRect(0, 0, W, H)
  ctx.fillStyle = '#12325a'
  const cw = W / cols, ch = H / rows
  for (let r = 0; r < rows; r++) {
    for (let c = 0; c < cols; c++) {
      ctx.fillRect(c * cw + 1.5, r * ch + 1.5, cw - 3, ch - 3)
    }
  }
  // شريط الأشرطة الموصلة — تفصيلة صغيرة بس هي الي تخلّيه يشبه لوحاً
  ctx.strokeStyle = '#93a4bd'
  ctx.lineWidth = 1
  for (let c = 0; c < cols; c++) {
    for (const f of [0.33, 0.66]) {
      ctx.beginPath()
      ctx.moveTo(c * cw + cw * f, 0)
      ctx.lineTo(c * cw + cw * f, H)
      ctx.stroke()
    }
  }
  tex.update()
  tex.vScale = -1
  tex.vOffset = 1
  pvTexCache.set(key, tex)
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
    } else if (f.kind === 'pvCells') {
      // ═══ خلايا اللوح ═══
      //
      // ⚠️ الخلايا **جسم واحد بشبكة خطوط** مو ٧٢ صندوقاً: لوح واحد
      // بـ٧٢ خلية × عشرة ألواح بالمشهد = ٧٢٠ جسماً، والمشهد ينهار.
      // الخطوط تعطي نفس الانطباع بجسم واحد.
      const cell = CreateBox(`cells_${deviceRef}`, { width: w * 0.94, height: h * 0.94, depth: 0.001 }, scene)
      const cm = new StandardMaterial(`m_cells_${deviceRef}`, scene)
      cm.diffuseTexture = pvTexture(scene, f.cols, f.rows)
      cm.specularColor = new Color3(0.4, 0.45, 0.55)
      cell.material = cm
      cell.position = new Vector3(0, 0, faceZ + 0.0008)
      cell.parent = root
      cell.isPickable = false
    } else if (f.kind === 'screen') {
      const sw = f.w * w, sh = f.h * h
      const scr = CreateBox(`scr_${deviceRef}`, { width: sw, height: sh, depth: 0.002 }, scene)
      const sm = new StandardMaterial(`m_scr_${deviceRef}`, scene)
      sm.diffuseColor = Color3.FromHexString('#0b2b22')
      sm.emissiveColor = Color3.FromHexString('#0b3a2c')
      scr.material = sm
      scr.position = onFace(f.x, f.y, faceZ + 0.0016)
      scr.parent = root
      scr.isPickable = false
    } else if (f.kind === 'breakerRow') {
      // صف قواطع — الرمز الي يميّز لوحة التوزيع بلمحة.
      const bw = (w * 0.8) / f.count
      for (let i = 0; i < f.count; i++) {
        const br = CreateBox(`brk_${deviceRef}_${i}`, { width: bw * 0.7, height: h * 0.16, depth: 0.004 }, scene)
        br.material = mat(scene, `m_brk_${deviceRef}`, '#e2e8f0')
        br.position = new Vector3(-w * 0.4 + bw * (i + 0.5), h * 0.1, faceZ + 0.003)
        br.parent = root
        br.isPickable = false
      }
    } else if (f.kind === 'portRow') {
      // ═══ صف منافذ RJ45 ═══
      //
      // ⚠️ هذا الي يخلّي السويچ **يُعرف بلمحة**: صندوق ١U بلا صف
      // منافذ يشبه أي جهاز رفّي ثاني. والمنافذ تنرسم **بجسمين**
      // (فتحة وإطار) مو بأربعة وعشرين جسماً منفصلاً.
      const pw = (w * 0.86) / f.count
      for (let i = 0; i < f.count; i++) {
        const px = -w * 0.43 + pw * (i + 0.5)
        const hole = CreateBox(`port_${deviceRef}_${i}`, { width: pw * 0.68, height: h * 0.42, depth: 0.004 }, scene)
        hole.material = mat(scene, `m_port_${deviceRef}`, '#05070b')
        hole.position = new Vector3(px, (0.5 - f.y) * h, faceZ + 0.002)
        hole.parent = root
        hole.isPickable = false
      }
    } else if (f.kind === 'disc') {
      // جهاز سقفي مدوّر — كاشف أو سماعة سقف.
      const disc = CreateCylinder(`disc_${deviceRef}`, { diameter: f.r * 2 * w, height: d * 0.5 }, scene)
      disc.rotation.x = Math.PI / 2
      disc.position = onFace(f.x, f.y, faceZ + d * 0.25)
      disc.material = mat(scene, `m_disc_${deviceRef}`, f.color ?? '#e2e8f0')
      disc.parent = root
      disc.isPickable = false
    } else if (f.kind === 'lens') {
      // عدسة الكاميرا — أسطوانة بارزة، وهي علامة الكاميرا الوحيدة.
      const lens = CreateCylinder(`lens_${deviceRef}`, { diameter: f.r * 2 * w, height: f.len * d }, scene)
      lens.rotation.x = Math.PI / 2
      lens.position = onFace(f.x, f.y, faceZ + (f.len * d) / 2)
      lens.material = mat(scene, `m_lens_${deviceRef}`, '#0b1220', { spec: 0.7 })
      lens.parent = root
      lens.isPickable = false
    } else if (f.kind === 'grille') {
      // شبك السماعة — حلقات متداخلة.
      const rings = f.rings ?? 3
      for (let i = 1; i <= rings; i++) {
        const ring = CreateCylinder(`gr_${deviceRef}_${i}`, {
          diameter: f.r * 2 * w * (i / rings), height: 0.002,
        }, scene)
        ring.rotation.x = Math.PI / 2
        ring.position = onFace(f.x, f.y, faceZ + 0.002 + i * 0.0006)
        ring.material = mat(scene, `m_gr_${deviceRef}_${i}`, i % 2 ? '#475569' : '#1e293b')
        ring.parent = root
        ring.isPickable = false
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
