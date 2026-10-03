// ═══ محرّك الكيابل ═══
//
// المخطط الرئيسي (٩): «المستخدم يسحب سلكاً بصرياً، لكن الخلفية تنشئ
// **Cable Entity** له conductors وgauge وlength وresistance وconnector
// ends». يعني الخط الي تشوفه مو الحقيقة — الحقيقة كيان له خصائص.
//
// ليش يفرق؟ لأن طول السلك ومقطعه يقرّران هبوط الفولتية. سلك ٠٫٥ ملم
// بطول ٤٠ متر لقفل يسحب ٥٥٠ ملّي أمبير يوصّل فولتية أقل من حد التحرير
// — والقفل ما ينفتح **مع إن التوصيل صحيح تماماً**. هذا عطل حقيقي
// يصير بالميدان، وما نگدر نعلّمه بخط مرسوم على شاشة.

import type { ConnectorKind, Terminal } from '../sim/types'

/** مقاومة النحاس النوعية بدرجة ٢٠° — أوم·مم²/م. */
const RHO_COPPER = 0.0172

/** مقاطع الأسلاك الشائعة عدنا للتحكم بالدخول. */
export const CABLE_GAUGES = [
  { id: 'awg22', label: '٢٢ AWG (٠٫٣٣ مم²)', areaMm2: 0.33 },
  { id: 'awg20', label: '٢٠ AWG (٠٫٥٢ مم²)', areaMm2: 0.52 },
  { id: 'awg18', label: '١٨ AWG (٠٫٨٢ مم²)', areaMm2: 0.82 },
  { id: 'mm15', label: '١٫٥ مم² (شائع عدنا)', areaMm2: 1.5 },
] as const

export type CableGaugeId = (typeof CABLE_GAUGES)[number]['id']

/** كيبل واحد بين طرفين — الكيان مو الخط. */
export interface Cable {
  id: string
  fromRef: string
  fromTerminal: string
  toRef: string
  toTerminal: string
  gauge: CableGaugeId
  /** الطول الفعلي بالمتر: مسافة المشهد + الطول المضاف بمسار السحب. */
  lengthM: number
  colorHex: string
}

/** المقاومة الذهاب-وإياب بالأوم. الدائرة تروح وترجع، فالطول ×٢. */
export function cableResistance(cable: Cable): number {
  const g = CABLE_GAUGES.find((x) => x.id === cable.gauge) || CABLE_GAUGES[1]
  return (RHO_COPPER * cable.lengthM * 2) / g.areaMm2
}

/** ═══ توافق الموصّلات (٩) ═══
 *
 *  «RJ45 لا يسنّب على terminal block». حالياً كل أطرافنا براغي فالفحص
 *  يمر دائماً — بس المحرّك موجود من هسه، فأول سويچ ينضاف يشتغل الرفض
 *  لحاله بلا ما ينلمس المحرّك. */
export function connectorOf(t: Terminal): ConnectorKind {
  const k = (t.kind || '').toUpperCase()
  if (k.startsWith('ETH') || k === 'RJ45' || k === 'LAN') return 'rj45'
  if (k === 'DC_JACK') return 'dc_jack'
  if (k === 'SFP') return 'sfp'
  if (k === 'USB') return 'usb'
  return 'screw_terminal'
}

export function connectorsCompatible(a: Terminal, b: Terminal): { ok: boolean; why?: string } {
  const ca = connectorOf(a)
  const cb = connectorOf(b)
  if (ca === cb) return { ok: true }
  return {
    ok: false,
    why: `ما ينسنّب: «${a.label}» موصّله ${LABEL[ca]} و«${b.label}» موصّله ${LABEL[cb]}.`,
  }
}

const LABEL: Record<ConnectorKind, string> = {
  screw_terminal: 'بلوك براغي',
  rj45: 'RJ45',
  dc_jack: 'مقبس تغذية',
  sfp: 'SFP',
  usb: 'USB',
}

/** ═══ هبوط الفولتية ═══
 *
 *  V_drop = I × R. الي يوصل الجهاز = فولتية المصدر − الهبوط. */
export function voltageAtLoad(sourceV: number, loadCurrentA: number, cables: Cable[]): number {
  const r = cables.reduce((sum, c) => sum + cableResistance(c), 0)
  return sourceV - loadCurrentA * r
}
