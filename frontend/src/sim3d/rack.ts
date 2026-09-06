// ═══ رفّ ١٩ إنچ يتولّد بالكود ═══
//
// الفني الي يشوف السويچ **مسطّحاً على الأرض** بالمشهد ما يتعلّم إن
// مكانه رفّ، ولا يتعلّم ترتيب الأجهزة بيه ولا مسار الكيابل. وأغلب
// مشاريعنا فيها رفّ.
//
// ⚠️⚠️ **هذا يبني الرفّ بس — ما يبني الأجهزة.** الأجهزة تبقى تنبني
// بـ`buildDevice` نفسها بلا أي تعديل، والرفّ يغيّر **وين** ينحط
// جذرها. راسم أجهزة خاص بالرف يعني السويچ يطلع بشكلين مختلفين حسب
// المنظر، وأول تصحيح على هندسته يوصل واحداً وينسى الثاني.
//
// ⚠️ المقاسات قياسية منشورة: الوحدة `U` = ٤٤٫٤٥ مم، وعرض التركيب
// ١٩ إنچ (٤٨٢٫٦ مم)، وعرض الهيكل الشائع ٦٠٠ مم.

import { StandardMaterial } from '@babylonjs/core/Materials/standardMaterial'
import { DynamicTexture } from '@babylonjs/core/Materials/Textures/dynamicTexture'
import { Color3 } from '@babylonjs/core/Maths/math.color'
import { Vector3 } from '@babylonjs/core/Maths/math.vector'
import { CreateBox } from '@babylonjs/core/Meshes/Builders/boxBuilder'
import { TransformNode } from '@babylonjs/core/Meshes/transformNode'
import type { Scene } from '@babylonjs/core/scene'

/** ارتفاع وحدة الرف بالمتر — المقاس القياسي. */
export const U_M = 0.04445
/** عرض التركيب بين الريلين (١٩ إنچ). */
export const MOUNT_W = 0.4826
/** عرض الهيكل الخارجي. */
const FRAME_W = 0.6
/** عمق الهيكل. */
const FRAME_D = 0.8
/** سماكة القائم. */
const POST = 0.02

export interface RackOpts {
  units: number
  position: Vector3
}

export interface BuiltRack {
  root: TransformNode
  /** ارتفاع مركز جهاز يبدي بالوحدة `uFromTop` وارتفاعه `heightU`. */
  slotY: (uFromTop: number, heightU: number) => number
  /** نقطة على الريل الجانبي — منها تمر الكيابل بدل ما تخترق الرفّ. */
  railPoint: (y: number, side: 1 | -1) => Vector3
  setDoor: (open: boolean) => void
  /** عمق الوجه الأمامي — الأجهزة تنركّب عليه. */
  faceZ: number
}

function mat(scene: Scene, hex: string, alpha = 1): StandardMaterial {
  const m = new StandardMaterial(`rk_${hex}_${alpha}`, scene)
  m.diffuseColor = Color3.FromHexString(hex)
  m.specularColor = new Color3(0.08, 0.08, 0.1)
  if (alpha < 1) { m.alpha = alpha; m.backFaceCulling = false }
  return m
}

/**
 * ═══ شريط أرقام الوحدات ═══
 *
 * ⚠️ **الأرقام إجبارية مو زينة.** بلاها الرفّ صندوق، ومعها يصير
 * المتدرّب يقرا «السويچ بالوحدة ١٨» — وهاي بالضبط اللغة الي تنكتب
 * بيها مخططات الرفوف بالميدان، والي راح يسمعها بأول مشروع.
 */
function uStripTexture(scene: Scene, units: number): DynamicTexture {
  const H = Math.max(256, units * 24)
  const tex = new DynamicTexture(`uStrip_${units}`, { width: 64, height: H }, scene, false)
  const ctx = tex.getContext() as CanvasRenderingContext2D
  ctx.fillStyle = '#0d131c'
  ctx.fillRect(0, 0, 64, H)
  const step = H / units
  ctx.textAlign = 'center'
  ctx.textBaseline = 'middle'
  for (let i = 0; i < units; i++) {
    const y = i * step
    ctx.strokeStyle = '#334155'
    ctx.lineWidth = 1
    ctx.beginPath(); ctx.moveTo(0, y); ctx.lineTo(64, y); ctx.stroke()
    // ⚠️ الترقيم من **فوق لتحت** لأن التموضع كله من فوق — ترقيم
    // معاكس للتموضع يخلّي كل قراءة تحتاج طرحاً ذهنياً.
    ctx.fillStyle = (i + 1) % 5 === 0 ? '#94a3b8' : '#475569'
    ctx.font = `${(i + 1) % 5 === 0 ? 'bold ' : ''}13px monospace`
    ctx.fillText(String(i + 1), 32, y + step / 2)
  }
  tex.update()
  return tex
}

export function buildRack(scene: Scene, opts: RackOpts): BuiltRack {
  const units = Math.max(4, Math.round(opts.units))
  const H = units * U_M
  const root = new TransformNode('rack', scene)
  root.position = opts.position.clone()

  // ⚠️ ألوان **أفتح من خلفية المشهد**: الرفّ بلون قريب من الخلفية
  // يختفي عملياً — والمتدرّب يشوف أجهزة طايفة بالهوا ويظن المشهد
  // مكسوراً. والرفّ الحقيقي معدن رمادي يعكس الضوء أصلاً.
  const frameMat = mat(scene, '#3a4757')
  const postMat = mat(scene, '#4a5768')

  const add = (name: string, w: number, h: number, d: number, p: Vector3, m: StandardMaterial) => {
    const b = CreateBox(name, { width: w, height: h, depth: d }, scene)
    b.material = m
    b.position = p
    b.parent = root
    b.isPickable = false
    return b
  }

  // قاعدة وسقف
  add('rk_base', FRAME_W, 0.03, FRAME_D, new Vector3(0, -0.015, 0), frameMat)
  add('rk_top', FRAME_W, 0.03, FRAME_D, new Vector3(0, H + 0.015, 0), frameMat)

  // أربعة قوائم
  for (const sx of [-1, 1]) {
    for (const sz of [-1, 1]) {
      add(`rk_post_${sx}_${sz}`, POST, H, POST,
        new Vector3(sx * (FRAME_W / 2 - POST / 2), H / 2, sz * (FRAME_D / 2 - POST / 2)), postMat)
    }
  }

  // جانبان مصمتان
  for (const sx of [-1, 1]) {
    add(`rk_side_${sx}`, 0.006, H, FRAME_D * 0.92,
      new Vector3(sx * FRAME_W / 2, H / 2, 0), mat(scene, '#26313f'))
  }

  const faceZ = FRAME_D / 2 - 0.02

  // ريلان أماميان بأرقام الوحدات
  for (const sx of [-1, 1]) {
    const rail = CreateBox(`rk_rail_${sx}`, { width: 0.028, height: H, depth: 0.012 }, scene)
    const rm = new StandardMaterial(`rk_railmat_${sx}`, scene)
    rm.diffuseTexture = uStripTexture(scene, units)
    // ⚠️ انبعاث خفيف على شريط الأرقام: بضوء المشهد وحده ما تنقرا
    // بالزوايا المظلمة — وشريط أرقام ما ينقرا مثل ماكو شريط.
    rm.emissiveTexture = uStripTexture(scene, units)
    rm.emissiveColor = new Color3(0.35, 0.35, 0.35)
    rm.specularColor = new Color3(0.05, 0.05, 0.05)
    rail.material = rm
    rail.position = new Vector3(sx * (MOUNT_W / 2 + 0.014), H / 2, faceZ)
    rail.parent = root
    rail.isPickable = false
  }

  // ═══ الباب الأمامي ═══
  // ⚠️ يتحرّك بالدوران حول حافته مثل الباب الحقيقي — مو يختفي.
  // اختفاؤه يخلّي المتدرّب ما يربط «سكّرت الباب» بأي شي مادي.
  const hinge = new TransformNode('rk_hinge', scene)
  hinge.parent = root
  hinge.position = new Vector3(-FRAME_W / 2, H / 2, FRAME_D / 2)
  const door = CreateBox('rk_door', { width: FRAME_W, height: H * 0.98, depth: 0.008 }, scene)
  // ⚠️ شفافية عالية: باب معتم يخفي الأجهزة الي وراه — وهذا عكس
  // فايدته. الباب الحقيقي بأغلب الرفوف زجاج مقوّى أو شبك.
  door.material = mat(scene, '#2a3646', 0.16)
  door.position = new Vector3(FRAME_W / 2, 0, 0)
  door.parent = hinge
  door.isPickable = false

  return {
    root,
    faceZ,
    // ⚠️ التموضع **من فوق**: الوحدة ١ هي الأعلى. عكسها يخالف ترقيم
    // الريل، فالمتدرّب يقرا رقماً ويشوف جهازاً بمحل ثاني.
    slotY: (uFromTop: number, heightU: number) =>
      H - (uFromTop - 1) * U_M - (heightU * U_M) / 2,
    railPoint: (y: number, side: 1 | -1) =>
      new Vector3(opts.position.x + side * (FRAME_W / 2 - 0.01), opts.position.y + y, opts.position.z + faceZ),
    // ⚠️ ~٧٥ درجة مو ١١٠: الباب المفتوح بزاوية كبيرة يصير لوحاً
    // عريضاً بمقدّمة المشهد يغطّي الرفّ نفسه.
    setDoor: (open: boolean) => { hinge.rotation.y = open ? -Math.PI * 0.42 : 0 },
  }
}
