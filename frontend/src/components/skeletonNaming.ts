// ═══════════════════════════════════════════════════════════════════
// خريطة أسماء العظام — نمط الهيكل يُكتشَف، مو مفروضاً
// ═══════════════════════════════════════════════════════════════════
//
// ⚠️ **السبب مقاس، مو احتياطاً نظرياً**: محرّك الحركة انبنى على أسماء
// **ميكسامو** (`mixamorig:RightHandIndex1`). ولمن جرّبنا مجسّماً ثانياً
// طلع نمطه **Rigify** (`DEF-f_index.01.R`)، والنتيجة إن كل قياس رجّع
// **صفر**: صفر عظمة إصبع وصفر تعبير — والمجسّم فيه **٣٠ عظمة إصبع**
// و**٣١ تعبير** فعلاً. يعني المشكلة كانت بكودنا مو بالمجسّم، والخطأ
// كان **صامتاً** (ولا رسالة) لأن `find` يرجّع `undefined` بهدوء.
//
// فبدل ما نعلّق المحرّك على مجسّم واحد، نكتشف النمط من أسماء العظام
// الموجودة فعلاً، ونرجّع خريطة موحّدة. وأي مجسّم جاي (VRM أو غيره)
// يحتاج بس نمطاً جديداً هنا.

export type RigKind = 'mixamo' | 'rigify' | 'unknown'
export type Hand = 'right' | 'left'

/** الأصابع بالترتيب — أسماء موحّدة بيننا، مستقلة عن نمط المجسّم. */
export const FINGER_KEYS = ['Thumb', 'Index', 'Middle', 'Ring', 'Pinky'] as const
export type FingerKey = typeof FINGER_KEYS[number]

export interface BoneNaming {
  kind: RigKind
  hips: string
  head: string
  /** عظمة قمة الرأس — للقياس (المفصل ما يتحرّك باللفّة). */
  headTop: string | null
  eye(hand: Hand): string | null
  arm(hand: Hand): string
  foreArm(hand: Hand): string
  hand(hand: Hand): string
  /** عظام مفصل إصبع واحد بالترتيب من القاعدة — فاضية لو ماكو. */
  fingerBones(hand: Hand, finger: FingerKey): string[]
  /** طرف السبّابة — للقياس وللإشارة. */
  fingerTip(hand: Hand): string[]
  /** نمط سلسلة الذراع — لمسك الحركة (`AnimationGroupMask`). */
  armChain(hand: Hand): RegExp
  /**
   * محور انثناء مفصل الإصبع **بفضاء العظمة**.
   *
   * ⚠️ **وهذا الي خلا القبض يطلع صفر مم**: كان مكتوباً `Z` لكل
   * الأنماط. وبـRigify العظمة ممتدة على `Y`، فالدوران حول `Z`
   * صار **لَفّة حول طول الإصبع** — ما تزحزح المفاصل الي بعدها ولا
   * ملّيمتر، وما تبيّن بالصورة. والانثناء الحقيقي حول `X`.
   */
  fingerBendAxis: readonly [number, number, number]
  /**
   * تصحيح وضعية الراحة لمّا المجسّم **ما بيه ولا مقطع حركة**.
   *
   * 🔴 **السبب مقاس من الملف نفسه**: مجسّم Rigify الي رفعناه فيه
   * **صفر مقاطع** (المصدّر ما حفظ وقفاته)، فوضعية الراحة هي الي
   * تُعرض — وهي **أذرع مفتوحة بزاوية ٥٩.٢° عن العمود**. وهاي تبيّن
   * للموظف كأن الشخصية **مكسورة** مو «واقفة ساكنة».
   *
   * والأرقام محسوبة من هيكل الـglTF مو مخمّنة: العظمة ممتدة على
   * `Y` المحلي، ومحور الإنزال هو `Z` المحلي (مقاس: `[0,0,∓1]`)،
   * والزاوية الطبيعية تخلي الذراع ~١٤° عن الجسم.
   *
   * `null` يعني «ما يحتاج تصحيحاً» — ميكسامو عنده مقاطعه.
   */
  restArmDrop: { angleRad: number } | null
}

const MIXAMO: BoneNaming = {
  kind: 'mixamo',
  hips: 'mixamorig:Hips',
  head: 'mixamorig:Head',
  headTop: 'mixamorig:HeadTop_End',
  // ⚠️ ميكسامو **ما ينطي عظام عيون إطلاقاً** — مقاس: صفر.
  eye: () => null,
  arm: (h) => `mixamorig:${h === 'right' ? 'Right' : 'Left'}Arm`,
  foreArm: (h) => `mixamorig:${h === 'right' ? 'Right' : 'Left'}ForeArm`,
  hand: (h) => `mixamorig:${h === 'right' ? 'Right' : 'Left'}Hand`,
  fingerBones: (h, f) => {
    const side = h === 'right' ? 'Right' : 'Left'
    return [1, 2, 3].map((i) => `mixamorig:${side}Hand${f}${i}`)
  },
  fingerTip: (h) => {
    const side = h === 'right' ? 'Right' : 'Left'
    return [`mixamorig:${side}HandIndex4`, `mixamorig:${side}HandIndex3`]
  },
  armChain: (h) =>
    new RegExp(`^mixamorig:${h === 'right' ? 'Right' : 'Left'}(Shoulder|Arm|ForeArm|Hand)`),
  fingerBendAxis: [0, 0, 1],
  // ميكسامو عنده ١١ مقطعاً يكتبون على العظام — فأي تصحيح ينمحي.
  restArmDrop: null,
}

/**
 * Rigify: عظام التشويه تجي بسابقة `DEF-`، والجهة لاحقة `.L`/`.R`،
 * والأصابع `f_index.01` بثلاث مفاصل، والإبهام `thumb.01`.
 */
const RIGIFY: BoneNaming = {
  kind: 'rigify',
  hips: 'DEF-spine',
  head: 'DEF-spine.006',
  headTop: null,
  eye: (h) => `DEF-eye.${h === 'right' ? 'R' : 'L'}`,
  arm: (h) => `DEF-upper_arm.${h === 'right' ? 'R' : 'L'}`,
  foreArm: (h) => `DEF-forearm.${h === 'right' ? 'R' : 'L'}`,
  hand: (h) => `DEF-hand.${h === 'right' ? 'R' : 'L'}`,
  fingerBones: (h, f) => {
    const s = h === 'right' ? 'R' : 'L'
    // الإبهام باسمه لحاله، والأربعة الباقية بسابقة `f_`.
    const base = f === 'Thumb' ? 'thumb' : `f_${f.toLowerCase()}`
    return ['01', '02', '03'].map((i) => `DEF-${base}.${i}.${s}`)
  },
  fingerTip: (h) => {
    const s = h === 'right' ? 'R' : 'L'
    // ماكو عظمة طرف بـRigify، فآخر مفصل بالسبّابة هو الأقرب للطرف.
    return [`DEF-f_index.03.${s}`, `DEF-f_index.02.${s}`]
  },
  armChain: (h) =>
    new RegExp(`^DEF-(shoulder|upper_arm|forearm|hand)\\.${h === 'right' ? 'R' : 'L'}`),
  fingerBendAxis: [1, 0, 0],
  // ٥٩.٢° مقاسة − ١٤° نبقّيها فرجة عن الجسم = ٤٥°.
  restArmDrop: { angleRad: (45 * Math.PI) / 180 },
}

/**
 * يكتشف النمط من أسماء العظام الموجودة.
 * ⚠️ الاكتشاف **بعظمة مميّزة لكل نمط** مو بالعدد: عدد العظام يتغيّر
 * بين مجسّم ومجسّم بنفس النمط.
 */
export function detectNaming(boneNames: readonly string[]): BoneNaming {
  const set = new Set(boneNames)
  if (set.has(MIXAMO.hips)) return MIXAMO
  if (boneNames.some((n) => n.startsWith('DEF-') && n.includes('spine'))) return RIGIFY
  // نمط مجهول: نرجّع ميكسامو حتى السلوك يبقى معروفاً، والشاشة
  // تعرض «—» لأن الأسماء ما تنلگى — وهذا أصدق من تخمين أسماء.
  return { ...MIXAMO, kind: 'unknown' }
}
