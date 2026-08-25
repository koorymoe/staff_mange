// ═══ مراحل العمل بالمنظومة ═══
//
// الفني بالميدان ما يشتغل دفعة وحدة — يمشي بمراحل، وكل مرحلة تخلص
// قبل الي بعدها. والمحاكي الي يعرض كلشي مرة وحدة يعلّم عادة غلط.
//
// ⚠️⚠️ **المرحلة تتقدّم من حالة المخطط الحقيقية مو بضغطة زر.** لو
// المتدرّب يگدر يضغط «التالي» بلا ما يخلّص، الشريط يصير زينة —
// وأسوأ: يعطيه إحساساً إنه خلّص وهو ما خلّص. كل مرحلة عدها **شرط
// محسوب**، والشريط يقرا الشرط مو نيّة المستخدم.

import type { LabDoc, SimResult } from './types'

export interface Stage {
  id: string
  label: string
  hint: string
}

export const SOLAR_STAGES: Stage[] = [
  { id: 'build', label: 'بناء النظام', hint: 'حط المكوّنات: ألواح، إنفرتر، بطارية، أحمال.' },
  { id: 'wire', label: 'التوصيل', hint: 'وصّل PV والبطارية بالإنفرتر، والأحمال بمخرج AC.' },
  { id: 'config', label: 'التهيئة', hint: 'اضبط عدد الألواح بالسلسلة وجهد البنك ونافذة MPPT.' },
  { id: 'run', label: 'التشغيل', hint: 'شغّل المحاكاة وشوف القراءات على المكوّنات.' },
  { id: 'diag', label: 'التشخيص', hint: 'ماكو أخطاء خطيرة — المنظومة سليمة.' },
]

export const GENERIC_STAGES: Stage[] = [
  { id: 'build', label: 'بناء النظام', hint: 'حط المكوّنات باللوح.' },
  { id: 'wire', label: 'التوصيل', hint: 'وصّل المنافذ ببعضها.' },
  { id: 'run', label: 'التشغيل', hint: 'شغّل المحاكاة.' },
  { id: 'diag', label: 'التشخيص', hint: 'صلّح الأخطاء الخطيرة.' },
]

export interface StageState {
  stages: Stage[]
  /** رقم المرحلة الحالية — أول وحدة ما اكتملت. */
  current: number
  /** اكتملت لو لا، لكل مرحلة. */
  done: boolean[]
}

/**
 * ═══ حساب المراحل ═══
 *
 * ⚠️ كل شرط **يُقاس من المستند أو من نتيجة المحرّك** — ماكو حالة
 * منفصلة تنخزن. يعني ما تگدر «تخلّص» مرحلة وترجع تكسر شرطها ويبقى
 * الشريط أخضر.
 */
export function computeStages(doc: LabDoc, result: SimResult | null): StageState {
  const stages = doc.domain === 'solar' ? SOLAR_STAGES : GENERIC_STAGES
  const has = (partId: string) => doc.nodes.some((n) => n.partId === partId)
  const errors = (result?.messages ?? []).filter((m) => m.kind === 'error').length

  const done: boolean[] = []
  if (doc.domain === 'solar') {
    const built = has('pv_panel') && has('inverter')
    // التوصيل: الإنفرتر عليه وصلات بمدخل PV ومخرج AC
    const inv = doc.nodes.find((n) => n.partId === 'inverter')
    const linked = (port: string) =>
      !!inv && doc.links.some((l) =>
        (l.from.node === inv.id && l.from.port === port) || (l.to.node === inv.id && l.to.port === port))
    const wired = linked('pv_pos') && linked('pv_neg') && linked('ac_out')
    // التهيئة: ماكو تحذير ستring — يعني النافذة والجهد مضبوطين
    const cfgWarn = (result?.messages ?? []).some((m) =>
      m.kind !== 'info' && /MPPT|Voc|بنك البطاريات/.test(m.text))
    done.push(built)
    done.push(built && wired)
    done.push(built && wired && !!result && !cfgWarn)
    done.push(!!result)
    done.push(!!result && errors === 0)
  } else {
    done.push(doc.nodes.length > 0)
    done.push(doc.links.length > 0)
    done.push(!!result)
    done.push(!!result && errors === 0)
  }

  // ⚠️ المرحلة الحالية = أول وحدة **ما اكتملت**، مو آخر وحدة اكتملت:
  // لو المتدرّب رجع وكسر شرط مرحلة سابقة، الشريط يرجع إلها.
  const current = done.findIndex((d) => !d)
  return { stages, done, current: current === -1 ? stages.length - 1 : current }
}
