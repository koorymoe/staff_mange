// ═══ حقن الأعطال ═══
//
// التدريب الحقيقي مو «ابنِ منظومة سليمة» — هو **«خذ منظومة خربانة
// ولگِ الخلل»**. الفني بالميدان نادراً يبني من الصفر؛ أغلب شغله
// يروح لموقع فيه شي مركّب وما يشتغل.
//
// ═══ القرار المعماري ═══
//
// ⚠️ **العطل = تعديل خفي على خصائص القطعة أو الوصلة** — مو حالة
// جديدة تفهمها المحرّكات. يعني: كونكتر وسخ = `extraLossDb: 3.5`
// على الوصلة، وظل على لوح = `shadeFactor: 0.4`.
//
// ليش هيچ؟ لأن المحرّكات **ما تحتاج تعرف بالأعطال أصلاً**. تشوف
// خصائص وتحسب. البديل — أن كل محرّك يفحص «هل اكو عطل من نوع
// كذا؟» — يعني كل عطل جديد يلمس خمسة محرّكات، وأول عطل ينُسى
// بمحرّك يصير عطلاً **ما يظهر** والمتدرّب يدوّر على شي مو موجود.
//
// ⚠️ والعطل **مخفي عن المتدرّب** بوضع التشخيص: عطل مكتوب على
// الشاشة مو عطل — هو إجابة. الي يشوفه المتدرّب هو **العَرَض** بس،
// ويلگي السبب بالقياس.

import type { LabDoc, DomainId } from './types'

export interface FaultDef {
  id: string
  label: string
  /** المجالات الي ينطبق عليها — فاضي يعني كلها. */
  domains?: DomainId[]
  target: 'node' | 'link'
  /** ينطبق على هالقطع بس — فاضي يعني أي قطعة. */
  parts?: string[]
  /** شنو يشوفه المتدرّب — للمدرّب، حتى يعرف شنو حقن. */
  symptom: string
  /** التعديل الخفي على الخصائص. */
  params: Record<string, string | number | boolean>
}

export const FAULTS: FaultDef[] = [
  // ═══ الألياف الضوئية ═══
  {
    id: 'dirty_connector', label: 'كونكتر وسخ', domains: ['gpon'], target: 'link',
    symptom: 'فقد زائد ٣٫٥ ديسيبل — القدرة تنزل والرابط يقوم وينزل.',
    params: { extraLossDb: 3.5 },
  },
  {
    id: 'fiber_bend', label: 'ليف مثني بزاوية حادة', domains: ['gpon'], target: 'link',
    symptom: 'فقد زائد ٢ ديسيبل — يصير لمن ينحشر الليف بمجرى ضيق.',
    params: { extraLossDb: 2 },
  },
  {
    id: 'bad_splice', label: 'لحام رديء', domains: ['gpon'], target: 'link',
    symptom: 'فقد زائد ١٫٥ ديسيبل بلحام واحد — اللحام السليم ياكل ٠٫١ بس.',
    params: { extraLossDb: 1.5 },
  },
  {
    id: 'fiber_cut', label: 'ليف مقطوع', domains: ['gpon'], target: 'link',
    symptom: 'ماكو ضوء أبداً — الONT ما يسجّل.',
    params: { extraLossDb: 60 },
  },

  // ═══ الشبكات ═══
  {
    id: 'link_down', label: 'كيبل مقطوع', domains: ['network'], target: 'link',
    symptom: 'الرابط ميت — الأجهزة وراه ما توصل.',
    params: { forceDown: true },
  },
  {
    id: 'cable_stretched', label: 'تمديد أطول من الحد', domains: ['network'], target: 'link',
    symptom: 'الرابط يشتغل بالفحص القريب ويتقطّع بالميدان.',
    params: { lengthM: 140 },
  },

  // ═══ الطاقة الشمسية ═══
  {
    id: 'shaded_string', label: 'ظل على الستring', domains: ['solar'], target: 'node',
    parts: ['pv_panel'],
    symptom: 'الإنتاج ينزل ٦٠٪ — عمود أو شجرة أو لوح جاره.',
    params: { shadeFactor: 0.4 },
  },
  {
    id: 'weak_battery', label: 'بطارية متدهورة', domains: ['solar'], target: 'node',
    parts: ['battery'],
    symptom: 'السعة الفعلية ثلث الاسمية — الاحتياطي ينتهي بسرعة.',
    params: { ah: 33 },
  },
  {
    id: 'loose_terminal', label: 'طرف مرتخٍ', domains: ['solar'], target: 'node',
    parts: ['pv_panel'],
    symptom: 'جهد الستring ينزل — كأن ألواحاً ناقصة.',
    params: { count: 3 },
  },

  // ═══ إنذار الحريق ═══
  {
    id: 'zone_open', label: 'دائرة زون مفتوحة', domains: ['fire'], target: 'link',
    symptom: 'اللوحة تصفّر «عطل» — كأن ماكو مقاومة نهاية.',
    params: { open: true },
  },
  {
    id: 'detector_short', label: 'قصر بكاشف', domains: ['fire'], target: 'node',
    parts: ['smoke_detector', 'heat_detector'],
    symptom: 'اللوحة تقرا عطلاً مو إنذاراً.',
    params: { shorted: true },
  },
  {
    id: 'weak_fire_battery', label: 'بطارية إنذار ضعيفة', domains: ['fire'], target: 'node',
    parts: ['fire_battery'],
    symptom: 'ما تغطّي ٢٤ ساعة استعداد.',
    params: { ah: 1.2 },
  },

  // ═══ الصوت ═══
  {
    id: 'wrong_tap', label: 'تاب سماعة أعلى من المفروض', domains: ['audio'], target: 'node',
    parts: ['ceiling_speaker', 'horn_speaker'],
    symptom: 'التحميل يزيد بلا ما يلاحظ الفني — يظهر لمن يجمع الخط.',
    params: { tapW: 30 },
  },
]

export function faultsFor(domain: DomainId, target: 'node' | 'link', partId?: string): FaultDef[] {
  return FAULTS.filter((f) =>
    f.target === target &&
    (!f.domains || f.domains.includes(domain)) &&
    (!f.parts || (partId ? f.parts.includes(partId) : false)))
}

export const FAULT_BY_ID: Record<string, FaultDef> = Object.fromEntries(FAULTS.map((f) => [f.id, f]))

/**
 * ═══ دمج الأعطال قبل التشغيل ═══
 *
 * ⚠️ يرجّع **مستنداً جديداً** — المستند الأصلي ما ينلمس. لو دمجنا
 * بالمكان، العطل يصير جزءاً من التصميم: يتحفظ مع المخطط ويطلع
 * كأنه اختيار المستخدم، وما تگدر تشيله.
 */
export function withFaults(doc: LabDoc): LabDoc {
  return {
    ...doc,
    nodes: doc.nodes.map((n) => {
      const f = n.fault ? FAULT_BY_ID[n.fault] : null
      return f ? { ...n, params: { ...n.params, ...f.params } } : n
    }),
    links: doc.links.map((l) => {
      const f = l.fault ? FAULT_BY_ID[l.fault] : null
      return f ? { ...l, params: { ...(l.params ?? {}), ...f.params } } : l
    }),
  }
}

/** عدد الأعطال المحقونة — للوحة المدرّب. */
export function countFaults(doc: LabDoc): number {
  return doc.nodes.filter((n) => n.fault).length + doc.links.filter((l) => l.fault).length
}
