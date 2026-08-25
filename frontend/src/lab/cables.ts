// ═══ الكيبلات والترانسيفرات ═══
//
// «أريد أنواع السويچات وأنواع الأسلاك وحتى التحويلات الـSFP».
//
// ⚠️ هذا الي يفرّق عن Packet Tracer فعلاً. Packet Tracer عنده أنواع
// كيبلات، بس **ما يحاسبك على الطول ولا على تطابق الترانسيفر مع نوع
// الليف**. وهاي بالضبط الأعطال الي توگف مشاريع عدنا:
//
// ١) **تمديد نحاس أطول من ١٠٠ متر.** يشتغل بالفحص القريب ويفشل
//    بالميدان أو يشتغل متقطّعاً — وهذا أسوأ من ما يشتغل.
// ٢) **ترانسيفر multimode على ليف singlemode** (أو العكس). الضوء
//    يمشي شوي والرابط يقوم وينزل — والفني يبدّل السويچ بلا فايدة.
// ٣) **ترانسيفرات مو متطابقة بالطرفين.** SX على LX ما يشتغلون سوا.
// ٤) **ترانسيفر بمنفذ نحاس أو كيبل نحاس بقفص SFP** — ما ينركّب أصلاً.

import { CABLE_GAUGES } from '../sim3d/cable'
import type { PortKind } from './types'

/** وسط النقل — النحاس والليف ما يتبادلون. */
export type Medium = 'copper' | 'mmf' | 'smf' | 'console'

export interface CableDef {
  id: string
  name: string
  medium: Medium
  /** أقصى مسافة بالمتر — تجاوزها يعني رابطاً غير موثوق. */
  maxM: number
  /** أقصى سرعة بالميغابت. */
  maxMbps: number
  about?: string
}

export const CABLES: CableDef[] = [
  { id: 'cat5e', name: 'Cat5e نحاس', medium: 'copper', maxM: 100, maxMbps: 1000,
    about: 'يكفي للجيجابت لحد ١٠٠ متر — الشائع بالمشاريع العادية عدنا.' },
  { id: 'cat6', name: 'Cat6 نحاس', medium: 'copper', maxM: 100, maxMbps: 1000,
    about: 'أحسن من Cat5e بالتشويش، ونفس حد الـ١٠٠ متر.' },
  { id: 'cat6a', name: 'Cat6a نحاس', medium: 'copper', maxM: 100, maxMbps: 10000,
    about: 'يشيل ١٠ جيجا لحد ١٠٠ متر — للربط بين الرفوف.' },
  { id: 'om3', name: 'ألياف Multimode OM3', medium: 'mmf', maxM: 300, maxMbps: 10000,
    about: 'للمسافات المتوسطة داخل المبنى — أرخص من الـsinglemode.' },
  { id: 'os2', name: 'ألياف Singlemode OS2', medium: 'smf', maxM: 10000, maxMbps: 10000,
    about: 'للمسافات الطويلة بين المباني.' },
  { id: 'console', name: 'كيبل كونسول', medium: 'console', maxM: 5, maxMbps: 1,
    about: 'للتهيئة الأولى بس — مو للبيانات.' },
]

export const CABLE_BY_ID: Record<string, CableDef> = Object.fromEntries(CABLES.map((c) => [c.id, c]))

export interface SfpDef {
  id: string
  name: string
  /** الوسط الي يشتغل عليه هالترانسيفر. */
  medium: Medium
  mbps: number
  maxM: number
}

export const SFPS: SfpDef[] = [
  { id: 'none', name: '— بلا ترانسيفر —', medium: 'copper', mbps: 0, maxM: 0 },
  { id: 'sfp_sx', name: 'SFP 1G SX (multimode)', medium: 'mmf', mbps: 1000, maxM: 550 },
  { id: 'sfp_lx', name: 'SFP 1G LX (singlemode)', medium: 'smf', mbps: 1000, maxM: 10000 },
  { id: 'sfp_t', name: 'SFP 1G RJ45 نحاس', medium: 'copper', mbps: 1000, maxM: 100 },
  { id: 'sfpp_sr', name: 'SFP+ 10G SR (multimode)', medium: 'mmf', mbps: 10000, maxM: 300 },
  { id: 'sfpp_lr', name: 'SFP+ 10G LR (singlemode)', medium: 'smf', mbps: 10000, maxM: 10000 },
]

export const SFP_BY_ID: Record<string, SfpDef> = Object.fromEntries(SFPS.map((s) => [s.id, s]))

export const MEDIUM_AR: Record<Medium, string> = {
  copper: 'نحاس',
  mmf: 'ألياف Multimode',
  smf: 'ألياف Singlemode',
  console: 'كونسول',
}

/** خصائص الوصلة الي تنعرض بلوحة الخصائص. */
export const LINK_PARAMS = [
  { id: 'cable', label: 'نوع الكيبل', kind: 'select' as const, default: 'cat6',
    options: CABLES.map((c) => ({ value: c.id, label: c.name })) },
  { id: 'lengthM', label: 'الطول', unit: 'م', kind: 'number' as const, default: 15, min: 1, max: 20000,
    help: 'الطول الفعلي بالمسار مو المسافة المستقيمة — الكيبل يلف بالسقف والمجاري.' },
  { id: 'sfpA', label: 'ترانسيفر الطرف الأول', kind: 'select' as const, default: 'none',
    options: SFPS.map((s) => ({ value: s.id, label: s.name })) },
  { id: 'sfpB', label: 'ترانسيفر الطرف الثاني', kind: 'select' as const, default: 'none',
    options: SFPS.map((s) => ({ value: s.id, label: s.name })) },
]

/** ═══ خصائص خط السماعات ═══
 *
 *  ⚠️ خط الصوت **مو كيبل شبكة**: ما إله Cat6 ولا ترانسيفر — إله
 *  **مقطع بالمليمتر المربّع** وطول، ومنهما تنحسب خسارة الخط. عرض
 *  خانة «نوع الكيبل Cat6» على خط سماعات يربك الفني ويعلّم غلط. */
export const SPK_LINK_PARAMS = [
  { id: 'gauge', label: 'مقطع السلك', kind: 'select' as const, default: 'mm15',
    options: CABLE_GAUGES.map((g) => ({ value: g.id, label: g.label })) },
  { id: 'lengthM', label: 'طول الخط', unit: 'م', kind: 'number' as const, default: 30, min: 1, max: 2000,
    help: 'خط الـ١٠٠ فولت انخترع أصلاً عشان المسافات الطويلة — بس الخسارة تبقى موجودة.' },
]

/** ═══ أي خصائص تنعطى لوصلة بين منفذين؟ ═══
 *
 *  ⚠️ مركزية عمداً: بدونها كل محرّك يخمّن شكل خصائص وصلاته، وأول
 *  مجال جديد يكتب صيغة رابعة. */
export function linkParamsFor(kind: PortKind) {
  if (kind === 'eth' || kind === 'sfp') return LINK_PARAMS
  if (kind === 'spk') return SPK_LINK_PARAMS
  return null
}

export interface LinkCheck {
  ok: boolean
  /** سرعة الرابط الفعلية بالميغابت — أقل عنصر بالسلسلة. */
  mbps: number
  problems: string[]
}

/**
 * ═══ فحص الرابط ═══
 *
 * ⚠️ **أضعف حلقة تقرّر**: الرابط بسرعة أقل عنصر، ويفشل لو أي عنصر
 * ما يطابق. هذا بالضبط سلوك الميدان — ترانسيفر ١٠ جيجا على ليف
 * قصير ما ينفع لو الطرف الثاني ١ جيجا.
 */
export function checkLink(
  cableId: string, lengthM: number, sfpAId: string, sfpBId: string,
  aIsSfpCage: boolean, bIsSfpCage: boolean,
): LinkCheck {
  const cable = CABLE_BY_ID[cableId] ?? CABLES[1]
  const A = SFP_BY_ID[sfpAId] ?? SFP_BY_ID.none
  const B = SFP_BY_ID[sfpBId] ?? SFP_BY_ID.none
  const problems: string[] = []
  let mbps = cable.maxMbps

  // ١) قفص SFP بلا ترانسيفر — ما ينربط أصلاً
  if (aIsSfpCage && A.id === 'none') problems.push('منفذ SFP بلا ترانسيفر — لازم تركّب وحدة بالقفص قبل الكيبل.')
  if (bIsSfpCage && B.id === 'none') problems.push('الطرف الثاني منفذ SFP بلا ترانسيفر.')

  // ٢) ترانسيفر بمنفذ نحاس عادي — ما يدخل بالقفص
  if (!aIsSfpCage && A.id !== 'none') problems.push('ركّبت ترانسيفر على منفذ نحاس عادي — الترانسيفر يدخل بقفص SFP بس.')
  if (!bIsSfpCage && B.id !== 'none') problems.push('الطرف الثاني: ترانسيفر على منفذ نحاس عادي.')

  // ٢ب) ليف على منفذ نحاس عادي — مستحيل فيزيائياً
  //
  // ⚠️ هاي انكشفت بالفحص: قبلها كيبل ألياف على منفذ RJ45 چان **يمر
  // بلا اعتراض** لأن كل القواعد الثانية تفحص الترانسيفر، وبلا ترانسيفر
  // ماكو شي تفحصه. الليف ما يدخل بمنفذ نحاس أصلاً — لا بقابس ولا
  // بشكل. القاعدة لازم تكون على **المنفذ** مو على الترانسيفر.
  if (cable.medium === 'mmf' || cable.medium === 'smf') {
    if (!aIsSfpCage) problems.push('كيبل ألياف على منفذ نحاس RJ45 — ما يدخل أصلاً. الليف يحتاج قفص SFP وترانسيفراً.')
    if (!bIsSfpCage) problems.push('الطرف الثاني: كيبل ألياف على منفذ نحاس RJ45.')
  }

  // ٣) تطابق الترانسيفر مع نوع الكيبل
  for (const [t, side] of [[A, 'الأول'], [B, 'الثاني']] as const) {
    if (t.id === 'none') continue
    if (t.medium !== cable.medium) {
      problems.push(
        `ترانسيفر الطرف ${side} (${MEDIUM_AR[t.medium]}) مو مطابق للكيبل (${MEDIUM_AR[cable.medium]}). ` +
        'بالميدان الرابط يقوم وينزل والفني يبدّل السويچ بلا فايدة.',
      )
    }
    mbps = Math.min(mbps, t.mbps)
  }

  // ٤) تطابق الطرفين ببعضهم
  if (A.id !== 'none' && B.id !== 'none' && A.mbps !== B.mbps) {
    problems.push(`ترانسيفرات مو متطابقة: ${A.name} مع ${B.name} — لازم يكونون بنفس السرعة والنوع.`)
  }

  // ٥) المسافة
  const limit = Math.min(cable.maxM, ...[A, B].filter((t) => t.id !== 'none').map((t) => t.maxM))
  if (lengthM > limit) {
    problems.push(
      `الطول ${lengthM} متر وحد هذا الرابط ${limit} متر. ` +
      (cable.medium === 'copper'
        ? 'النحاس فوگ ١٠٠ متر يشتغل بالفحص القريب ويتقطّع بالميدان — وهذا أسوأ من ما يشتغل.'
        : 'الضوء ما يوصل — تحتاج نوعاً ثانياً أو مقوّياً.'),
    )
  }

  // ٦) كيبل كونسول للبيانات
  if (cable.medium === 'console') problems.push('كيبل الكونسول للتهيئة بس — ما ينقل بيانات الشبكة.')

  return { ok: problems.length === 0, mbps: problems.length ? 0 : mbps, problems }
}
