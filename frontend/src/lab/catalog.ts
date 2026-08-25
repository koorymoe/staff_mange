// ═══ كتالوگ القطع ═══
//
// ⚠️ **بيانات مو كود.** إضافة قطعة جديدة = سطر بهذا الملف، مو تعديل
// بالمحرّك ولا باللوح (المخطط ٨: «إضافة جهاز عملية content engineering
// وليس feature programming»).
//
// ⚠️ القيم أدناه **نمطية عامة** مأخوذة من الشائع بالسوق — مو كتالوگ
// موديل بعينه. ولهذا كل محتوى المختبر `verified = FALSE` لحد ما فني
// يجرّبه على جهاز حقيقي.

import type { PartDef } from './types'

// ═══ الدوائر الكهربائية ═══
const ELECTRICAL: PartDef[] = [
  {
    id: 'dc_source', domain: 'electrical', name: 'مصدر تغذية مستمر', symbol: 'battery',
    w: 90, h: 60, about: 'مصدر جهد مستمر — البطارية أو المحوّل.',
    ports: [
      { id: 'pos', label: '+', kind: 'dc', polarity: 'pos', x: 1, y: 0.32 },
      { id: 'neg', label: '−', kind: 'dc', polarity: 'neg', x: 1, y: 0.72 },
    ],
    params: [
      { id: 'v', label: 'الجهد', unit: 'V', kind: 'number', default: 12, min: 1, max: 400 },
      { id: 'rInt', label: 'المقاومة الداخلية', unit: 'Ω', kind: 'number', default: 0.05, min: 0,
        help: 'مصدر مثالي ما موجود — المقاومة الداخلية هي الي تخلّي الجهد ينزل تحت الحمل.' },
    ],
  },
  {
    id: 'resistor', domain: 'electrical', name: 'مقاومة', symbol: 'resistor',
    w: 90, h: 40,
    ports: [
      { id: 'a', label: 'A', kind: 'dc', polarity: 'none', x: 0, y: 0.5 },
      { id: 'b', label: 'B', kind: 'dc', polarity: 'none', x: 1, y: 0.5 },
    ],
    params: [{ id: 'r', label: 'المقاومة', unit: 'Ω', kind: 'number', default: 100, min: 0.001 }],
  },
  {
    id: 'lamp', domain: 'electrical', name: 'لمبة', symbol: 'lamp',
    w: 70, h: 70, about: 'تضوّي لمن يمر بيها تيار كافٍ، وتحترق لو زاد على حدّها.',
    ports: [
      { id: 'a', label: 'A', kind: 'dc', polarity: 'none', x: 0, y: 0.5 },
      { id: 'b', label: 'B', kind: 'dc', polarity: 'none', x: 1, y: 0.5 },
    ],
    params: [
      { id: 'vNom', label: 'الجهد الاسمي', unit: 'V', kind: 'number', default: 12, min: 1 },
      { id: 'pNom', label: 'القدرة', unit: 'W', kind: 'number', default: 5, min: 0.1 },
    ],
  },
  {
    id: 'switch', domain: 'electrical', name: 'مفتاح', symbol: 'switch',
    w: 80, h: 40,
    ports: [
      { id: 'a', label: 'A', kind: 'dc', polarity: 'none', x: 0, y: 0.5 },
      { id: 'b', label: 'B', kind: 'dc', polarity: 'none', x: 1, y: 0.5 },
    ],
    params: [{ id: 'closed', label: 'مغلق', kind: 'bool', default: false }],
  },
  {
    id: 'fuse', domain: 'electrical', name: 'فيوز', symbol: 'fuse',
    w: 80, h: 40, about: 'ينقطع لمن يزيد التيار على حدّه — يحمي الي وراه.',
    ports: [
      { id: 'a', label: 'A', kind: 'dc', polarity: 'none', x: 0, y: 0.5 },
      { id: 'b', label: 'B', kind: 'dc', polarity: 'none', x: 1, y: 0.5 },
    ],
    params: [{ id: 'iMax', label: 'تيار الفصل', unit: 'A', kind: 'number', default: 5, min: 0.1 }],
  },
  {
    id: 'motor', domain: 'electrical', name: 'محرّك DC', symbol: 'motor',
    w: 80, h: 70,
    ports: [
      { id: 'a', label: '+', kind: 'dc', polarity: 'pos', x: 0, y: 0.5 },
      { id: 'b', label: '−', kind: 'dc', polarity: 'neg', x: 1, y: 0.5 },
    ],
    params: [
      { id: 'vNom', label: 'الجهد الاسمي', unit: 'V', kind: 'number', default: 12, min: 1 },
      { id: 'iNom', label: 'التيار الاسمي', unit: 'A', kind: 'number', default: 1.2, min: 0.01 },
    ],
    danger: 'عكس القطبية يدير المحرّك بالاتجاه المعاكس — وببعض الأحمال يكسر الميكانيك.',
  },
]

// ═══ الطاقة الشمسية ═══
const SOLAR: PartDef[] = [
  {
    id: 'pv_panel', domain: 'solar', name: 'لوح شمسي', symbol: 'pv',
    w: 110, h: 75, about: 'لوح كهروضوئي — خرجه يتبع الإشعاع ودرجة الحرارة.',
    ports: [
      { id: 'pos', label: '+', kind: 'dc', polarity: 'pos', x: 1, y: 0.3 },
      { id: 'neg', label: '−', kind: 'dc', polarity: 'neg', x: 1, y: 0.7 },
    ],
    params: [
      { id: 'pmax', label: 'قدرة اللوح', unit: 'W', kind: 'number', default: 550, min: 10 },
      { id: 'vmp', label: 'جهد أعظم قدرة Vmp', unit: 'V', kind: 'number', default: 41.5, min: 1 },
      { id: 'voc', label: 'جهد الدارة المفتوحة Voc', unit: 'V', kind: 'number', default: 49.8, min: 1,
        help: 'Voc يرتفع لمن تبرد الأجواء — وهو الي يحرق الإنفرتر بالشتاء لو الستring طويل.' },
      { id: 'count', label: 'عدد الألواح بالسلسلة', kind: 'number', default: 6, min: 1, max: 30 },
    ],
  },
  {
    id: 'inverter', domain: 'solar', name: 'إنفرتر هجين', symbol: 'inverter',
    w: 120, h: 110,
    ports: [
      { id: 'pv_pos', label: 'PV +', kind: 'dc', polarity: 'pos', x: 0, y: 0.22 },
      { id: 'pv_neg', label: 'PV −', kind: 'dc', polarity: 'neg', x: 0, y: 0.42 },
      { id: 'bat_pos', label: 'BAT +', kind: 'dc', polarity: 'pos', x: 0, y: 0.66 },
      { id: 'bat_neg', label: 'BAT −', kind: 'dc', polarity: 'neg', x: 0, y: 0.86 },
      { id: 'ac_out', label: 'AC OUT', kind: 'ac', polarity: 'none', x: 1, y: 0.5 },
    ],
    params: [
      { id: 'pRated', label: 'القدرة', unit: 'W', kind: 'number', default: 5000, min: 100 },
      { id: 'mpptMin', label: 'أدنى جهد MPPT', unit: 'V', kind: 'number', default: 120, min: 10 },
      { id: 'mpptMax', label: 'أعلى جهد MPPT', unit: 'V', kind: 'number', default: 450, min: 20 },
      { id: 'vdcMax', label: 'أقصى جهد DC مسموح', unit: 'V', kind: 'number', default: 500, min: 20,
        help: 'تجاوزه يحرق الإنفرتر — وهذا أكثر غلط يصير بتركيب الستring.' },
      { id: 'batV', label: 'جهد بنك البطاريات', unit: 'V', kind: 'select', default: '48',
        options: [{ value: '12', label: '١٢ فولت' }, { value: '24', label: '٢٤ فولت' }, { value: '48', label: '٤٨ فولت' }] },
    ],
    danger: 'ربط PV بمداخل البطارية يحرق الإنفرتر فوراً — المداخل تشبه بعضها بأغلب الموديلات.',
  },
  {
    id: 'battery', domain: 'solar', name: 'بطارية', symbol: 'battery_bank',
    w: 110, h: 80,
    ports: [
      { id: 'pos', label: '+', kind: 'dc', polarity: 'pos', x: 1, y: 0.3 },
      { id: 'neg', label: '−', kind: 'dc', polarity: 'neg', x: 1, y: 0.7 },
    ],
    params: [
      { id: 'v', label: 'جهد البطارية', unit: 'V', kind: 'select', default: '48',
        options: [{ value: '12', label: '١٢ فولت' }, { value: '24', label: '٢٤ فولت' }, { value: '48', label: '٤٨ فولت' }] },
      { id: 'ah', label: 'السعة', unit: 'Ah', kind: 'number', default: 100, min: 1 },
      { id: 'soc', label: 'حالة الشحن', unit: '%', kind: 'number', default: 80, min: 0, max: 100 },
      { id: 'dod', label: 'أقصى تفريغ مسموح', unit: '%', kind: 'number', default: 80, min: 10, max: 100 },
    ],
  },
  {
    id: 'load', domain: 'solar', name: 'حمل', symbol: 'load',
    w: 95, h: 70,
    ports: [{ id: 'ac_in', label: 'AC', kind: 'ac', polarity: 'none', x: 0, y: 0.5 }],
    params: [
      { id: 'p', label: 'القدرة', unit: 'W', kind: 'number', default: 1200, min: 1 },
      { id: 'hours', label: 'ساعات التشغيل', unit: 'س', kind: 'number', default: 5, min: 0.1, max: 24 },
      { id: 'name', label: 'الاسم', kind: 'text', default: 'أحمال المنزل' },
    ],
  },
]

// ═══ الشبكات ═══
// ⚠️ منافذ الوصلات (`kind: 'sfp'`) **مو نفس** منافذ النحاس: محرّك
// اللوح يرفض ربطهن ببعض، ومحرّك الشبكة يطلب ترانسيفراً بيهن. هذا الي
// يخلّي «ركّبت كيبل نحاس بقفص SFP» غلطاً ينكشف بالمحاكي.
const swPorts = (n: number, sfp: number) => [
  ...Array.from({ length: n }, (_, i) => ({
    id: `gi0/${i + 1}`, label: `Gi0/${i + 1}`, kind: 'eth' as const,
    x: 0.06 + (i * 0.72) / Math.max(1, n - 1), y: 1,
  })),
  ...Array.from({ length: sfp }, (_, i) => ({
    id: `sfp${i + 1}`, label: `SFP${i + 1}`, kind: 'sfp' as const,
    x: 0.86 + i * 0.1, y: 1,
  })),
]

const NETWORK: PartDef[] = [
  {
    id: 'switch_unmanaged', domain: 'network', name: 'سويچ ٨ منافذ غير مدار', symbol: 'net_switch',
    w: 120, h: 52, about: 'سويچ بسيط — ما يتهيّأ ولا يعرف VLAN. الشائع بالبيوت والمحلات الصغيرة.',
    ports: swPorts(4, 0),
    params: [{ id: 'hostname', label: 'الاسم', kind: 'text', default: 'SW-DUMB' }],
    danger: 'ما يدعم VLAN — أي عزل منطقي تبنيه بالمشروع ينهار لمن يمر بهذا السويچ.',
  },
  {
    id: 'switch_l2', domain: 'network', name: 'سويچ ٢٤ منفذ مدار', symbol: 'net_switch',
    w: 140, h: 60, about: 'سويچ طبقة ٢ قابل للإدارة — VLAN وترنك، ويتهيّأ بسطر الأوامر.',
    ports: swPorts(4, 2),
    params: [
      { id: 'hostname', label: 'اسم الجهاز', kind: 'text', default: 'SW1' },
      { id: 'poeBudget', label: 'ميزانية PoE', unit: 'W', kind: 'number', default: 0, min: 0,
        help: 'صفر يعني موديل بلا PoE — الكاميرات تحتاج تغذية مستقلة.' },
    ],
  },
  {
    id: 'switch_poe', domain: 'network', name: 'سويچ ٢٤ منفذ PoE+', symbol: 'net_switch',
    w: 140, h: 60, about: 'يغذّي الكاميرات ونقاط الوصول بنفس كيبل الشبكة.',
    ports: swPorts(4, 2),
    params: [
      { id: 'hostname', label: 'اسم الجهاز', kind: 'text', default: 'SW-POE' },
      { id: 'poeBudget', label: 'ميزانية PoE', unit: 'W', kind: 'number', default: 185, min: 0,
        help: 'المجموع لكل المنافذ مو لكل منفذ — وهذا الي ينسوه الفنيون.' },
    ],
  },
  {
    id: 'switch_l3', domain: 'network', name: 'سويچ طبقة ٣', symbol: 'net_switch',
    w: 145, h: 62, about: 'يوجّه بين الـVLANات بلا راوتر خارجي.',
    ports: swPorts(4, 2),
    params: [
      { id: 'hostname', label: 'اسم الجهاز', kind: 'text', default: 'SW-L3' },
      { id: 'routing', label: 'التوجيه بين الـVLANات', kind: 'bool', default: true },
      { id: 'poeBudget', label: 'ميزانية PoE', unit: 'W', kind: 'number', default: 0, min: 0 },
    ],
  },
  {
    id: 'router', domain: 'network', name: 'راوتر', symbol: 'net_router',
    w: 120, h: 65,
    ports: [
      { id: 'gi0/0', label: 'Gi0/0', kind: 'eth', x: 0.2, y: 1 },
      { id: 'gi0/1', label: 'Gi0/1', kind: 'eth', x: 0.8, y: 1 },
    ],
    params: [
      { id: 'hostname', label: 'اسم الجهاز', kind: 'text', default: 'R1' },
      { id: 'ip0', label: 'عنوان Gi0/0', kind: 'text', default: '192.168.1.1' },
      { id: 'ip1', label: 'عنوان Gi0/1', kind: 'text', default: '10.0.0.1' },
    ],
  },
  {
    id: 'pc', domain: 'network', name: 'حاسبة', symbol: 'net_pc',
    w: 85, h: 70,
    ports: [{ id: 'eth0', label: 'Eth0', kind: 'eth', x: 0.5, y: 1 }],
    params: [
      { id: 'name', label: 'الاسم', kind: 'text', default: 'PC1' },
      { id: 'ip', label: 'العنوان', kind: 'text', default: '192.168.1.10' },
      { id: 'mask', label: 'القناع', kind: 'text', default: '255.255.255.0' },
      { id: 'gw', label: 'البوابة', kind: 'text', default: '192.168.1.1' },
    ],
    // ⚠️ ماكو خانة VLAN هنا عمداً: الجهاز الطرفي ما يختار VLANه —
    // **منفذ السويچ** هو الي يحطّه بيه. الخانة هنا تعلّم عادة غلط.
  },
  {
    id: 'ip_camera', domain: 'network', name: 'كاميرا شبكة', symbol: 'net_cam',
    w: 90, h: 65, about: 'كاميرا IP تتغذّى بـPoE من السويچ.',
    ports: [{ id: 'eth0', label: 'Eth0', kind: 'eth', x: 0.5, y: 1 }],
    params: [
      { id: 'name', label: 'الاسم', kind: 'text', default: 'CAM1' },
      { id: 'ip', label: 'العنوان', kind: 'text', default: '192.168.1.51' },
      { id: 'mask', label: 'القناع', kind: 'text', default: '255.255.255.0' },
      { id: 'gw', label: 'البوابة', kind: 'text', default: '192.168.1.1' },
      { id: 'poeW', label: 'سحب PoE', unit: 'W', kind: 'number', default: 8, min: 0 },
    ],
  },
]

export const PARTS: PartDef[] = [...ELECTRICAL, ...SOLAR, ...NETWORK]

export const PART_BY_ID: Record<string, PartDef> = Object.fromEntries(PARTS.map((p) => [p.id, p]))

export const DOMAINS: { id: 'electrical' | 'solar' | 'network'; name: string; icon: string; about: string }[] = [
  { id: 'network', name: 'الشبكات', icon: '🌐', about: 'سويچات وراوترات وحاسبات وكاميرات — وصّل وهيّئ واختبر الاتصال.' },
  { id: 'solar', name: 'الطاقة الشمسية', icon: '☀️', about: 'ألواح وإنفرتر وبطاريات وأحمال — افحص الستring والتوازن قبل الميدان.' },
  { id: 'electrical', name: 'الدوائر الكهربائية', icon: '⚡', about: 'مصادر ومقاومات ولمبات ومفاتيح — الدائرة تنحل فعلاً بقوانين كيرشوف.' },
]

/** القيم الابتدائية لقطعة — تنسخ من الكتالوگ لمن تنحط باللوح. */
export function defaultParams(part: PartDef): Record<string, string | number | boolean> {
  return Object.fromEntries(part.params.map((p) => [p.id, p.default]))
}
