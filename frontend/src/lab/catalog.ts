// ═══ كتالوگ القطع ═══
//
// ⚠️ **بيانات مو كود.** إضافة قطعة جديدة = سطر بهذا الملف، مو تعديل
// بالمحرّك ولا باللوح (المخطط ٨: «إضافة جهاز عملية content engineering
// وليس feature programming»).
//
// ⚠️ القيم أدناه **نمطية عامة** مأخوذة من الشائع بالسوق — مو كتالوگ
// موديل بعينه. ولهذا كل محتوى المختبر `verified = FALSE` لحد ما فني
// يجرّبه على جهاز حقيقي.

import type { DomainId, PartDef } from './types'

// ═══ الدوائر الكهربائية ═══
const ELECTRICAL: PartDef[] = [
  {
    id: 'dc_source', domain: 'electrical', name: 'مصدر تغذية مستمر', model: 'مصدر جهد · مقاومة داخلية', symbol: 'battery',
    w: 90, h: 60, geo3d: { sizeM: { w: 0.12, h: 0.08, d: 0.06 }, bodyColorHex: '#334155', faceColorHex: '#1e293b', features: [{ kind: 'screen', x: 0.4, y: 0.4, w: 0.4, h: 0.3 }] },
    about: 'مصدر جهد مستمر — البطارية أو المحوّل.',
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
    id: 'resistor', domain: 'electrical', name: 'مقاومة', model: 'عنصر خطّي', symbol: 'resistor',
    w: 90, h: 40,
    geo3d: { sizeM: { w: 0.05, h: 0.015, d: 0.015 }, bodyColorHex: '#a16207', faceColorHex: '#78350f' },
    ports: [
      { id: 'a', label: 'A', kind: 'dc', polarity: 'none', x: 0, y: 0.5 },
      { id: 'b', label: 'B', kind: 'dc', polarity: 'none', x: 1, y: 0.5 },
    ],
    params: [{ id: 'r', label: 'المقاومة', unit: 'Ω', kind: 'number', default: 100, min: 0.001 }],
  },
  {
    id: 'lamp', domain: 'electrical', name: 'لمبة', model: 'حمل مقاوم · عتبة احتراق', symbol: 'lamp',
    w: 70, h: 70, geo3d: { sizeM: { w: 0.07, h: 0.07, d: 0.07 }, bodyColorHex: '#fde68a', faceColorHex: '#fcd34d', features: [{ kind: 'disc', x: 0.5, y: 0.5, r: 0.4, color: '#fef3c7' }] },
    about: 'تضوّي لمن يمر بيها تيار كافٍ، وتحترق لو زاد على حدّها.',
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
    id: 'switch', domain: 'electrical', name: 'مفتاح', model: 'مفتاح مفرد', symbol: 'switch',
    w: 80, h: 40,
    geo3d: { sizeM: { w: 0.07, h: 0.07, d: 0.03 }, bodyColorHex: '#e2e8f0', faceColorHex: '#cbd5e1' },
    ports: [
      { id: 'a', label: 'A', kind: 'dc', polarity: 'none', x: 0, y: 0.5 },
      { id: 'b', label: 'B', kind: 'dc', polarity: 'none', x: 1, y: 0.5 },
    ],
    params: [{ id: 'closed', label: 'مغلق', kind: 'bool', default: false }],
  },
  {
    id: 'fuse', domain: 'electrical', name: 'فيوز', model: 'حماية تيار زائد', symbol: 'fuse',
    w: 80, h: 40, geo3d: { sizeM: { w: 0.05, h: 0.03, d: 0.03 }, bodyColorHex: '#64748b', faceColorHex: '#475569' },
    about: 'ينقطع لمن يزيد التيار على حدّه — يحمي الي وراه.',
    ports: [
      { id: 'a', label: 'A', kind: 'dc', polarity: 'none', x: 0, y: 0.5 },
      { id: 'b', label: 'B', kind: 'dc', polarity: 'none', x: 1, y: 0.5 },
    ],
    params: [{ id: 'iMax', label: 'تيار الفصل', unit: 'A', kind: 'number', default: 5, min: 0.1 }],
  },
  {
    id: 'motor', domain: 'electrical', name: 'محرّك DC', model: 'حمل حثّي مبسّط', symbol: 'motor',
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
    id: 'pv_panel', domain: 'solar', name: 'لوح شمسي', model: '٥٥٠ واط · Vmp ٤١٫٥', symbol: 'pv',
    w: 110, h: 75, geo3d: { sizeM: { w: 1.13, h: 2.28, d: 0.035 }, bodyColorHex: '#1e293b', faceColorHex: '#0f2744', tiltDeg: 28, features: [{ kind: 'pvCells', cols: 6, rows: 12 }] },
    about: 'لوح كهروضوئي — خرجه يتبع الإشعاع ودرجة الحرارة.',
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
    id: 'inverter', domain: 'solar', name: 'إنفرتر هجين', model: 'هجين ٥ كيلوواط · MPPT', symbol: 'inverter',
    w: 120, h: 110,
    geo3d: { sizeM: { w: 0.42, h: 0.58, d: 0.17 }, bodyColorHex: '#cbd5e1', faceColorHex: '#94a3b8', features: [{ kind: 'screen', x: 0.5, y: 0.34, w: 0.5, h: 0.2 }, { kind: 'statusLed', x: 0.3, y: 0.62 }, { kind: 'statusLed', x: 0.5, y: 0.62 }, { kind: 'statusLed', x: 0.7, y: 0.62 }] },
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
    id: 'battery', domain: 'solar', name: 'بطارية', model: 'بنك ٤٨ فولت · ١٠٠ أمبير·ساعة', symbol: 'battery_bank',
    w: 110, h: 80,
    geo3d: { sizeM: { w: 0.44, h: 0.22, d: 0.40 }, bodyColorHex: '#0f172a', faceColorHex: '#1e293b', features: [{ kind: 'screen', x: 0.5, y: 0.45, w: 0.34, h: 0.3 }] },
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
    id: 'load', domain: 'solar', name: 'حمل', model: 'حمل تيار متناوب', symbol: 'load',
    w: 95, h: 70,
    geo3d: { sizeM: { w: 0.36, h: 0.50, d: 0.14 }, bodyColorHex: '#334155', faceColorHex: '#1e293b', features: [{ kind: 'breakerRow', count: 6 }] },
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
    id: 'switch_unmanaged', domain: 'network', name: 'سويچ ٨ منافذ غير مدار', model: '٨ منافذ · بلا إدارة', symbol: 'net_switch',
    w: 120, h: 52, geo3d: { sizeM: { w: 0.22, h: 0.028, d: 0.10 }, bodyColorHex: '#334155', faceColorHex: '#1e293b', features: [{ kind: 'portRow', count: 8, y: 0.5 }, { kind: 'statusLed', x: 0.94, y: 0.5 }] },
    about: 'سويچ بسيط — ما يتهيّأ ولا يعرف VLAN. الشائع بالبيوت والمحلات الصغيرة.',
    ports: swPorts(4, 0),
    params: [{ id: 'hostname', label: 'الاسم', kind: 'text', default: 'SW-DUMB' }],
    danger: 'ما يدعم VLAN — أي عزل منطقي تبنيه بالمشروع ينهار لمن يمر بهذا السويچ.',
  },
  {
    id: 'switch_l2', domain: 'network', name: 'سويچ ٢٤ منفذ مدار', model: '٢٤ منفذ · طبقة ٢ · SFP×٢', symbol: 'net_switch',
    w: 140, h: 60, geo3d: { sizeM: { w: 0.44, h: 0.044, d: 0.20 }, bodyColorHex: '#2b3440', faceColorHex: '#161d27', features: [{ kind: 'portRow', count: 12, y: 0.5 }, { kind: 'statusLed', x: 0.96, y: 0.3 }] },
    about: 'سويچ طبقة ٢ قابل للإدارة — VLAN وترنك، ويتهيّأ بسطر الأوامر.',
    ports: swPorts(4, 2),
    params: [
      { id: 'hostname', label: 'اسم الجهاز', kind: 'text', default: 'SW1' },
      { id: 'poeBudget', label: 'ميزانية PoE', unit: 'W', kind: 'number', default: 0, min: 0,
        help: 'صفر يعني موديل بلا PoE — الكاميرات تحتاج تغذية مستقلة.' },
    ],
  },
  {
    id: 'switch_poe', domain: 'network', name: 'سويچ ٢٤ منفذ PoE+', model: '٢٤ منفذ · PoE+ ١٨٥ واط', symbol: 'net_switch',
    w: 140, h: 60, geo3d: { sizeM: { w: 0.44, h: 0.044, d: 0.22 }, bodyColorHex: '#33404f', faceColorHex: '#161d27', features: [{ kind: 'portRow', count: 12, y: 0.5 }, { kind: 'statusLed', x: 0.96, y: 0.3 }, { kind: 'statusLed', x: 0.96, y: 0.7 }] },
    about: 'يغذّي الكاميرات ونقاط الوصول بنفس كيبل الشبكة.',
    ports: swPorts(4, 2),
    params: [
      { id: 'hostname', label: 'اسم الجهاز', kind: 'text', default: 'SW-POE' },
      { id: 'poeBudget', label: 'ميزانية PoE', unit: 'W', kind: 'number', default: 185, min: 0,
        help: 'المجموع لكل المنافذ مو لكل منفذ — وهذا الي ينسوه الفنيون.' },
    ],
  },
  {
    id: 'switch_l3', domain: 'network', name: 'سويچ طبقة ٣', model: '٢٤ منفذ · توجيه داخلي', symbol: 'net_switch',
    w: 145, h: 62, geo3d: { sizeM: { w: 0.44, h: 0.044, d: 0.24 }, bodyColorHex: '#3b4a5c', faceColorHex: '#161d27', features: [{ kind: 'portRow', count: 12, y: 0.5 }, { kind: 'statusLed', x: 0.96, y: 0.3 }] },
    about: 'يوجّه بين الـVLANات بلا راوتر خارجي.',
    ports: swPorts(4, 2),
    params: [
      { id: 'hostname', label: 'اسم الجهاز', kind: 'text', default: 'SW-L3' },
      { id: 'routing', label: 'التوجيه بين الـVLANات', kind: 'bool', default: true },
      { id: 'poeBudget', label: 'ميزانية PoE', unit: 'W', kind: 'number', default: 0, min: 0 },
    ],
  },
  {
    id: 'router', domain: 'network', name: 'راوتر', model: 'منفذان · توجيه بين شبكتين', symbol: 'net_router',
    w: 120, h: 65,
    geo3d: { sizeM: { w: 0.34, h: 0.044, d: 0.22 }, bodyColorHex: '#2f3b4a', faceColorHex: '#151b24', features: [{ kind: 'portRow', count: 4, y: 0.5 }, { kind: 'statusLed', x: 0.92, y: 0.5 }] },
    ports: [
      { id: 'gi0/0', label: 'Gi0/0', kind: 'eth', x: 0.2, y: 1 },
      { id: 'gi0/1', label: 'Gi0/1', kind: 'eth', x: 0.8, y: 1 },
    ],
    params: [
      { id: 'hostname', label: 'اسم الجهاز', kind: 'text', default: 'R1' },
      { id: 'ip_gi0/0', label: 'عنوان Gi0/0', kind: 'text', default: '192.168.1.1' },
      { id: 'mask_gi0/0', label: 'قناع Gi0/0', kind: 'text', default: '255.255.255.0' },
      { id: 'ip_gi0/1', label: 'عنوان Gi0/1', kind: 'text', default: '10.0.0.1' },
      { id: 'mask_gi0/1', label: 'قناع Gi0/1', kind: 'text', default: '255.255.255.0' },
    ],
    about: 'يربط شبكتين فرعيتين. كل منفذ يحتاج عنواناً **داخل** شبكته، والأجهزة تحتاج بوابتها تؤشّر عليه.',
  },
  {
    id: 'pc', domain: 'network', name: 'حاسبة', model: 'محطة عمل', symbol: 'net_pc',
    w: 85, h: 70,
    geo3d: { sizeM: { w: 0.19, h: 0.42, d: 0.42 }, bodyColorHex: '#1e293b', faceColorHex: '#0f172a', features: [{ kind: 'statusLed', x: 0.5, y: 0.12 }] },
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
    id: 'ip_camera', domain: 'network', name: 'كاميرا شبكة', model: 'كاميرا IP · PoE', symbol: 'net_cam',
    w: 90, h: 65, geo3d: { sizeM: { w: 0.09, h: 0.09, d: 0.16 }, bodyColorHex: '#e2e8f0', faceColorHex: '#cbd5e1', features: [{ kind: 'lens', x: 0.5, y: 0.5, r: 0.3, len: 0.9 }] },
    about: 'كاميرا IP تتغذّى بـPoE من السويچ.',
    ports: [{ id: 'eth0', label: 'Eth0', kind: 'eth', x: 0.5, y: 1 }],
    params: [
      { id: 'name', label: 'الاسم', kind: 'text', default: 'CAM1' },
      { id: 'ip', label: 'العنوان', kind: 'text', default: '192.168.1.51' },
      { id: 'mask', label: 'القناع', kind: 'text', default: '255.255.255.0' },
      { id: 'gw', label: 'البوابة', kind: 'text', default: '192.168.1.1' },
      { id: 'poeW', label: 'سحب PoE نهاراً', unit: 'W', kind: 'number', default: 8, min: 0 },
      // ⚠️ سحب الأشعة **منفصل**: هو الي يخلّي المنظومة تمر نهاراً
      // وتفصل ليلاً. لو جمعناه بالسحب الأساسي، الدرس كله ينمحي.
      { id: 'irW', label: 'زيادة سحب الأشعة (ليلاً)', unit: 'W', kind: 'number', default: 4, min: 0,
        help: 'تشتغل بالظلام بس — وهاي أسوأ حالة، وعليها ينبني التصميم.' },
      { id: 'res', label: 'الدقة', kind: 'select', default: '2mp',
        options: [
          { value: '2mp', label: '٢ ميغابكسل · 1080p' },
          { value: '4mp', label: '٤ ميغابكسل' },
          { value: '5mp', label: '٥ ميغابكسل' },
          { value: '8mp', label: '٨ ميغابكسل · 4K' },
        ] },
      { id: 'fps', label: 'الإطارات بالثانية', kind: 'number', default: 15, min: 1, max: 60,
        help: 'المراقبة ما تحتاج فوگ ١٥ عادةً — الزيادة تاكل قرصاً ونطاقاً.' },
      { id: 'codec', label: 'الكودك', kind: 'select', default: 'h264',
        options: [
          { value: 'h264', label: 'H.264' },
          { value: 'h265', label: 'H.265 · نصف الحجم تقريباً' },
        ] },
    ],
  },
]

// ═══ أنظمة إنذار الحريق ═══
//
// ⚠️ أجهزة الزون عدها منفذان (`in`/`out`) لأنها تنربط **بالسلسلة**
// مو بالتوازي — والسلسلة هي الي تخلّي اللوحة تراقب الخط كله بمقاومة
// نهاية وحدة.
const FIRE: PartDef[] = [
  {
    id: 'fire_panel', domain: 'fire', name: 'لوحة إنذار تقليدية', model: 'زونان · دائرة إنذار', symbol: 'fire_panel',
    w: 130, h: 105, geo3d: { sizeM: { w: 0.32, h: 0.40, d: 0.09 }, bodyColorHex: '#7f1d1d', faceColorHex: '#450a0a', features: [{ kind: 'screen', x: 0.5, y: 0.26, w: 0.6, h: 0.22 }, { kind: 'statusLed', x: 0.3, y: 0.56 }, { kind: 'statusLed', x: 0.5, y: 0.56 }, { kind: 'statusLed', x: 0.7, y: 0.56 }] },
    about: 'لوحة زونات: كل زون خط كواشف، ولكل دائرة إنذار صفّاراتها.',
    ports: [
      { id: 'z1', label: 'زون ١', kind: 'signal', x: 0, y: 0.22 },
      { id: 'z2', label: 'زون ٢', kind: 'signal', x: 0, y: 0.45 },
      { id: 's1', label: 'إنذار ١', kind: 'signal', x: 0, y: 0.72 },
      { id: 'bat', label: 'بطارية', kind: 'dc', polarity: 'pos', x: 1, y: 0.72 },
    ],
    params: [
      { id: 'name', label: 'الاسم', kind: 'text', default: 'FACP' },
      { id: 'zones', label: 'عدد الزونات', kind: 'number', default: 2, min: 1, max: 2 },
      { id: 'sounderCircuits', label: 'دوائر الإنذار', kind: 'number', default: 1, min: 1, max: 1 },
      { id: 'standbyMa', label: 'سحب اللوحة بالاستعداد', unit: 'mA', kind: 'number', default: 60, min: 0 },
      { id: 'alarmMa', label: 'سحب اللوحة بالإنذار', unit: 'mA', kind: 'number', default: 120, min: 0 },
    ],
    danger: 'خلط الصفّارات مع الكواشف بنفس الزون يخلّي اللوحة تحسبها كاشفاً معطّلاً.',
  },
  {
    id: 'smoke_detector', domain: 'fire', name: 'كاشف دخان', model: 'ضوئي · سلسلة', symbol: 'detector',
    w: 70, h: 70,
    geo3d: { sizeM: { w: 0.10, h: 0.10, d: 0.045 }, bodyColorHex: '#f1f5f9', faceColorHex: '#e2e8f0', features: [{ kind: 'disc', x: 0.5, y: 0.5, r: 0.36, color: '#cbd5e1' }, { kind: 'statusLed', x: 0.5, y: 0.5 }] },
    ports: [
      { id: 'in', label: 'داخل', kind: 'signal', x: 0, y: 0.5 },
      { id: 'out', label: 'خارج', kind: 'signal', x: 1, y: 0.5 },
    ],
    params: [
      { id: 'standbyMa', label: 'سحب الاستعداد', unit: 'mA', kind: 'number', default: 0.05, min: 0 },
      { id: 'alarmMa', label: 'سحب الإنذار', unit: 'mA', kind: 'number', default: 0.5, min: 0 },
      { id: 'triggered', label: 'مفعّل (محاكاة حريق)', kind: 'bool', default: false },
      { id: 'shorted', label: 'قصر بالدائرة (محاكاة عطل)', kind: 'bool', default: false },
    ],
  },
  {
    id: 'heat_detector', domain: 'fire', name: 'كاشف حرارة', model: 'حراري · سلسلة', symbol: 'detector',
    w: 70, h: 70, geo3d: { sizeM: { w: 0.10, h: 0.10, d: 0.045 }, bodyColorHex: '#e2e8f0', faceColorHex: '#cbd5e1', features: [{ kind: 'disc', x: 0.5, y: 0.5, r: 0.3, color: '#94a3b8' }] },
    about: 'للمطابخ والمرائب — الدخان بيها يعطي إنذارات كاذبة.',
    ports: [
      { id: 'in', label: 'داخل', kind: 'signal', x: 0, y: 0.5 },
      { id: 'out', label: 'خارج', kind: 'signal', x: 1, y: 0.5 },
    ],
    params: [
      { id: 'standbyMa', label: 'سحب الاستعداد', unit: 'mA', kind: 'number', default: 0.05, min: 0 },
      { id: 'alarmMa', label: 'سحب الإنذار', unit: 'mA', kind: 'number', default: 0.5, min: 0 },
      { id: 'triggered', label: 'مفعّل (محاكاة حريق)', kind: 'bool', default: false },
    ],
  },
  {
    id: 'mcp', domain: 'fire', name: 'نقطة إنذار يدوية', model: 'نقطة يدوية · سلسلة', symbol: 'mcp',
    w: 70, h: 70, geo3d: { sizeM: { w: 0.09, h: 0.09, d: 0.04 }, bodyColorHex: '#dc2626', faceColorHex: '#991b1b', features: [{ kind: 'screen', x: 0.5, y: 0.45, w: 0.55, h: 0.35 }] },
    about: 'الصندوق الأحمر عند المخارج — ينكسر زجاجه بالإنذار اليدوي.',
    ports: [
      { id: 'in', label: 'داخل', kind: 'signal', x: 0, y: 0.5 },
      { id: 'out', label: 'خارج', kind: 'signal', x: 1, y: 0.5 },
    ],
    params: [
      { id: 'standbyMa', label: 'سحب الاستعداد', unit: 'mA', kind: 'number', default: 0.05, min: 0 },
      { id: 'triggered', label: 'مفعّل', kind: 'bool', default: false },
    ],
  },
  {
    id: 'sounder', domain: 'fire', name: 'صفّارة', model: 'صفّارة · دائرة إنذار', symbol: 'sounder',
    w: 75, h: 70,
    geo3d: { sizeM: { w: 0.10, h: 0.10, d: 0.06 }, bodyColorHex: '#dc2626', faceColorHex: '#7f1d1d', features: [{ kind: 'grille', x: 0.5, y: 0.5, r: 0.38, rings: 3 }] },
    ports: [
      { id: 'in', label: 'داخل', kind: 'signal', x: 0, y: 0.5 },
      { id: 'out', label: 'خارج', kind: 'signal', x: 1, y: 0.5 },
    ],
    params: [{ id: 'alarmMa', label: 'سحب الإنذار', unit: 'mA', kind: 'number', default: 20, min: 0 }],
    danger: 'تنربط على **دائرة إنذار** مو على زون كواشف.',
  },
  {
    id: 'eol_resistor', domain: 'fire', name: 'مقاومة نهاية (EOL)', model: '٤٫٧ كيلو أوم · رِجلان', symbol: 'eol',
    w: 70, h: 46,
    geo3d: { sizeM: { w: 0.035, h: 0.012, d: 0.012 }, bodyColorHex: '#a16207', faceColorHex: '#78350f' },
    about: 'تنحط بآخر جهاز بالخط. بدونها اللوحة ما تگدر تفرّق بين خط سليم وخط مقطوع.',
    // ⚠️ **منفذان** مو واحد: المقاومة الحقيقية إلها رِجلان، والفني
    // يگدر يحطها بنص الخط غلطاً — وهذا بالضبط الغلط الي نريد نعلّمه.
    // بمنفذ واحد يصير الغلط **مستحيل البناء**، والمحاكي يخفي درساً
    // بدل ما يعلّمه.
    ports: [
      { id: 'in', label: 'رِجل ١', kind: 'signal', x: 0, y: 0.5 },
      { id: 'out', label: 'رِجل ٢', kind: 'signal', x: 1, y: 0.5 },
    ],
    params: [{ id: 'r', label: 'القيمة', unit: 'Ω', kind: 'number', default: 4700, min: 100 }],
    danger: 'محلها **آخر الخط** — أي جهاز وراها يبقى خارج الحماية تماماً.',
  },
  {
    id: 'fire_battery', domain: 'fire', name: 'بطارية احتياط', model: 'احتياط ٧ أمبير·ساعة', symbol: 'battery_bank',
    w: 100, h: 70,
    geo3d: { sizeM: { w: 0.15, h: 0.10, d: 0.065 }, bodyColorHex: '#0f172a', faceColorHex: '#1e293b' },
    ports: [{ id: 'pos', label: '+', kind: 'dc', polarity: 'pos', x: 0, y: 0.5 }],
    params: [
      { id: 'ah', label: 'السعة', unit: 'Ah', kind: 'number', default: 7, min: 1 },
      { id: 'soc', label: 'حالة الشحن', unit: '%', kind: 'number', default: 100, min: 0, max: 100 },
    ],
  },
]

// ═══ الصوت والإذاعة ═══
//
// ⚠️ منافذ السماعات نوعها `spk` مو `signal`: هذا الي يخلّي اللوح
// يعطي الوصلة **خصائص خط سماعات** (مقطع وطول) بدل خصائص كيبل شبكة،
// ويمنع ربط سماعة بمنفذ شبكة.
const AUDIO: PartDef[] = [
  {
    id: 'amplifier', domain: 'audio', name: 'مكبّر صوت', model: '١٢٠ واط · ١٠٠ فولت أو لو-Z', symbol: 'amplifier',
    w: 130, h: 80,
    geo3d: { sizeM: { w: 0.44, h: 0.088, d: 0.32 }, bodyColorHex: '#1f2937', faceColorHex: '#111827', features: [{ kind: 'screen', x: 0.35, y: 0.5, w: 0.3, h: 0.4 }, { kind: 'statusLed', x: 0.9, y: 0.35 }, { kind: 'statusLed', x: 0.9, y: 0.65 }] },
    about: 'خط ١٠٠ فولت للمسافات الطويلة والمناطق، أو مقاومة منخفضة للقاعات القريبة.',
    ports: [
      { id: 'out1', label: 'مخرج ١', kind: 'spk', x: 1, y: 0.35 },
      { id: 'out2', label: 'مخرج ٢', kind: 'spk', x: 1, y: 0.68 },
      { id: 'mic', label: 'ميكروفون', kind: 'signal', x: 0, y: 0.5 },
    ],
    params: [
      { id: 'name', label: 'الاسم', kind: 'text', default: 'AMP1' },
      { id: 'pRated', label: 'القدرة', unit: 'W', kind: 'number', default: 120, min: 10 },
      { id: 'mode', label: 'وضع المخرج', kind: 'select', default: '100v',
        options: [{ value: '100v', label: 'خط ١٠٠ فولت' }, { value: 'lowz', label: 'مقاومة منخفضة' }] },
      { id: 'minOhm', label: 'أدنى مقاومة (وضع لو-Z)', unit: 'Ω', kind: 'number', default: 4, min: 1,
        help: 'النزول تحتها يعني تيار زائد — فصل أو احتراق.' },
    ],
    danger: 'ربط مخرجَي مكبّرين ببعض يحرق الاثنين.',
  },
  {
    id: 'ceiling_speaker', domain: 'audio', name: 'سماعة سقف', model: 'سقفية · تاب قابل للاختيار', symbol: 'speaker',
    w: 80, h: 75,
    geo3d: { sizeM: { w: 0.20, h: 0.20, d: 0.09 }, bodyColorHex: '#f8fafc', faceColorHex: '#e2e8f0', features: [{ kind: 'grille', x: 0.5, y: 0.5, r: 0.42, rings: 4 }] },
    ports: [
      { id: 'in', label: 'داخل', kind: 'spk', x: 0, y: 0.5 },
      { id: 'out', label: 'خارج', kind: 'spk', x: 1, y: 0.5 },
    ],
    params: [
      { id: 'mode', label: 'الوضع', kind: 'select', default: '100v',
        options: [{ value: '100v', label: 'بمحوّل ١٠٠ فولت' }, { value: 'lowz', label: 'مقاومة منخفضة' }] },
      { id: 'tapW', label: 'التاب', unit: 'W', kind: 'select', default: 6,
        options: [{ value: '3', label: '٣ واط' }, { value: '6', label: '٦ واط' }, { value: '12', label: '١٢ واط' }, { value: '24', label: '٢٤ واط' }] },
      { id: 'ohm', label: 'المقاومة (وضع لو-Z)', unit: 'Ω', kind: 'number', default: 8, min: 1 },
    ],
  },
  {
    id: 'horn_speaker', domain: 'audio', name: 'بوق خارجي', model: 'بوق خارجي · ٢٤ واط', symbol: 'horn',
    w: 85, h: 70, geo3d: { sizeM: { w: 0.22, h: 0.16, d: 0.24 }, bodyColorHex: '#475569', faceColorHex: '#334155', features: [{ kind: 'grille', x: 0.5, y: 0.5, r: 0.38, rings: 2 }] },
    about: 'للساحات والمخازن — صوت عالٍ ومدى بعيد.',
    ports: [
      { id: 'in', label: 'داخل', kind: 'spk', x: 0, y: 0.5 },
      { id: 'out', label: 'خارج', kind: 'spk', x: 1, y: 0.5 },
    ],
    params: [
      { id: 'mode', label: 'الوضع', kind: 'select', default: '100v',
        options: [{ value: '100v', label: 'بمحوّل ١٠٠ فولت' }, { value: 'lowz', label: 'مقاومة منخفضة' }] },
      { id: 'tapW', label: 'التاب', unit: 'W', kind: 'select', default: 24,
        options: [{ value: '10', label: '١٠ واط' }, { value: '15', label: '١٥ واط' }, { value: '24', label: '٢٤ واط' }, { value: '30', label: '٣٠ واط' }] },
      { id: 'ohm', label: 'المقاومة (وضع لو-Z)', unit: 'Ω', kind: 'number', default: 8, min: 1 },
    ],
  },
  {
    id: 'paging_mic', domain: 'audio', name: 'ميكروفون نداء', model: 'ميكروفون نداء', symbol: 'mic',
    w: 75, h: 70,
    geo3d: { sizeM: { w: 0.11, h: 0.05, d: 0.13 }, bodyColorHex: '#1e293b', faceColorHex: '#0f172a', features: [{ kind: 'statusLed', x: 0.5, y: 0.5 }] },
    ports: [{ id: 'out', label: 'خرج', kind: 'signal', x: 1, y: 0.5 }],
    params: [{ id: 'name', label: 'الاسم', kind: 'text', default: 'MIC1' }],
  },
]

// ═══ الألياف الضوئية GPON ═══
//
// ⚠️ منافذ الليف نوعها `fiber`: اللوح يعطي وصلاتها **طولاً بالكيلومتر
// وعدد لحامات** بدل «نوع كيبل Cat6»، ويمنع ربط ليف بمنفذ نحاس.
const GPON: PartDef[] = [
  {
    id: 'olt', domain: 'gpon', name: 'OLT — طرفية الخط الضوئي', model: 'GPON Class B+ · ٨ منافذ PON',
    symbol: 'olt', w: 145, h: 62,
    geo3d: { sizeM: { w: 0.44, h: 0.088, d: 0.30 }, bodyColorHex: '#243244', faceColorHex: '#141b26',
      features: [{ kind: 'portRow', count: 8, y: 0.5 }, { kind: 'statusLed', x: 0.96, y: 0.3 }] },
    about: 'الجهاز الي بالمقسّم — يغذّي مئات المشتركين بليف واحد لكل منفذ PON.',
    ports: [
      { id: 'pon1', label: 'PON 1', kind: 'fiber', x: 0.2, y: 1 },
      { id: 'pon2', label: 'PON 2', kind: 'fiber', x: 0.5, y: 1 },
      { id: 'uplink', label: 'صاعد', kind: 'sfp', x: 0.85, y: 1 },
    ],
    params: [
      { id: 'name', label: 'الاسم', kind: 'text', default: 'OLT-1' },
      { id: 'txDbm', label: 'قدرة الإرسال', unit: 'dBm', kind: 'number', default: 3, min: -5, max: 10,
        help: 'فئة B+ ترسل بين +١٫٥ و+٥ ديسيبل·ملّي واط.' },
      { id: 'ponCapacity', label: 'سعة منفذ PON', kind: 'number', default: 64, min: 8, max: 128 },
      { id: 'serviceVlan', label: 'VLAN الخدمة', kind: 'number', default: 35, min: 1, max: 4094,
        help: 'لازم يطابق الي بالONT — وإلا يسجّل بلا إنترنت.' },
    ],
  },
  {
    id: 'splitter', domain: 'gpon', name: 'سبليتر ضوئي', model: 'مقسّم بصري · نسبة قابلة للاختيار',
    symbol: 'splitter', w: 95, h: 70,
    geo3d: { sizeM: { w: 0.14, h: 0.035, d: 0.10 }, bodyColorHex: '#475569', faceColorHex: '#334155' },
    about: 'يقسّم الليف الواحد على عدة مشتركين — وكل تقسيم ياكل من الميزانية الضوئية.',
    ports: [
      { id: 'in', label: 'داخل', kind: 'fiber', x: 0, y: 0.5 },
      { id: 'o1', label: 'خرج ١', kind: 'fiber', x: 1, y: 0.2 },
      { id: 'o2', label: 'خرج ٢', kind: 'fiber', x: 1, y: 0.4 },
      { id: 'o3', label: 'خرج ٣', kind: 'fiber', x: 1, y: 0.6 },
      { id: 'o4', label: 'خرج ٤', kind: 'fiber', x: 1, y: 0.8 },
    ],
    params: [
      { id: 'ratio', label: 'نسبة التقسيم', kind: 'select', default: '8',
        options: [
          { value: '2', label: '1:2 · فقد ٣٫٦ dB' }, { value: '4', label: '1:4 · فقد ٧٫٣ dB' },
          { value: '8', label: '1:8 · فقد ١٠٫٥ dB' }, { value: '16', label: '1:16 · فقد ١٣٫٧ dB' },
          { value: '32', label: '1:32 · فقد ١٧٫٠ dB' }, { value: '64', label: '1:64 · فقد ٢٠٫٥ dB' },
        ] },
    ],
    danger: 'كل زيادة بنسبة التقسيم تاكل من الميزانية — 1:64 ياكل ٢٠ ديسيبل، وهذا نص الميزانية تقريباً.',
  },
  {
    id: 'ont', domain: 'gpon', name: 'ONT — وحدة المشترك', model: 'راوتر ضوئي · WiFi + LAN',
    symbol: 'ont', w: 100, h: 66,
    geo3d: { sizeM: { w: 0.19, h: 0.035, d: 0.13 }, bodyColorHex: '#e2e8f0', faceColorHex: '#cbd5e1',
      features: [{ kind: 'statusLed', x: 0.2, y: 0.5 }, { kind: 'statusLed', x: 0.35, y: 0.5 }, { kind: 'statusLed', x: 0.5, y: 0.5 }] },
    about: 'الجهاز الي ببيت الزبون — يستقبل الضوء ويطلّع إنترنت وWiFi.',
    ports: [
      { id: 'pon', label: 'PON', kind: 'fiber', x: 0, y: 0.5 },
      { id: 'lan1', label: 'LAN 1', kind: 'eth', x: 1, y: 0.35 },
      { id: 'lan2', label: 'LAN 2', kind: 'eth', x: 1, y: 0.65 },
    ],
    params: [
      { id: 'name', label: 'الاسم', kind: 'text', default: 'ONT-1' },
      { id: 'rxMin', label: 'أدنى حساسية', unit: 'dBm', kind: 'number', default: -27, min: -40, max: -10 },
      { id: 'rxMax', label: 'حد الإشباع', unit: 'dBm', kind: 'number', default: -8, min: -20, max: 0,
        help: 'قدرة **أعلى** من هذا تعمي المستقبل — الONT ما يشتغل.' },
      { id: 'wanMode', label: 'وضع WAN', kind: 'select', default: 'pppoe',
        options: [{ value: 'pppoe', label: 'PPPoE' }, { value: 'dhcp', label: 'DHCP' }, { value: 'static', label: 'ثابت' }] },
      { id: 'pppoeUser', label: 'اسم مستخدم PPPoE', kind: 'text', default: '' },
      { id: 'wanVlan', label: 'معرّف VLAN', kind: 'number', default: 35, min: 1, max: 4094 },
      { id: 'ssid', label: 'اسم الشبكة اللاسلكية', kind: 'text', default: 'Home-WiFi' },
    ],
  },
]


// ═══ الكاميرات والمراقبة ═══
//
// ⚠️ الكاميرا هنا **نفس** قطعة الشبكات بالضبط (`ip_camera`) — مو
// نسخة ثانية. نسختان تعني خاصية تنضاف لوحدة وتنسى بالثانية، وفني
// يتدرّب على كاميرا ما تشبه الي بمختبر الشبكات.
//
// الي ينضاف هنا هو **الي يفرق بمنظومة مراقبة**: المسجّل والقرص
// والشاشة.
const CCTV: PartDef[] = [
  {
    id: 'nvr', domain: 'cctv', name: 'مسجّل شبكي NVR', model: 'قنوات · قرص · منافذ PoE', symbol: 'nvr',
    w: 150, h: 62,
    geo3d: { sizeM: { w: 0.38, h: 0.05, d: 0.32 }, bodyColorHex: '#28313d', faceColorHex: '#141a22', features: [{ kind: 'portRow', count: 8, y: 0.5 }, { kind: 'statusLed', x: 0.94, y: 0.35 }, { kind: 'disc', x: 0.2, y: 0.5, r: 0.12 }] },
    about: 'قلب المنظومة: يسجّل ويخزّن، وحدوده (القنوات والإدخال وPoE) هي الي تقرّر شكد كاميرا تنفع.',
    ports: [
      { id: 'lan', label: 'LAN', kind: 'eth', x: 0.1, y: 1 },
      ...Array.from({ length: 4 }, (_, i) => ({
        id: `poe${i + 1}`, label: `PoE${i + 1}`, kind: 'eth' as const, x: 0.34 + i * 0.16, y: 1,
      })),
    ],
    params: [
      { id: 'name', label: 'الاسم', kind: 'text', default: 'NVR1' },
      { id: 'channels', label: 'عدد القنوات', kind: 'number', default: 8, min: 1,
        help: 'الكاميرات فوگ هذا العدد ما تنسجّل — والمنظومة تبدو شغّالة.' },
      { id: 'maxInMbps', label: 'حد الإدخال', unit: 'Mbps', kind: 'number', default: 80, min: 1,
        help: 'مجموع بثّ كل الكاميرات لازم يبقى تحته.' },
      { id: 'diskTb', label: 'سعة القرص', unit: 'TB', kind: 'number', default: 4, min: 0 },
      { id: 'retentionDays', label: 'أيام التسجيل المطلوبة', unit: 'يوم', kind: 'number', default: 30, min: 1 },
      { id: 'poePorts', label: 'منافذ PoE', kind: 'number', default: 4, min: 0 },
      { id: 'poeBudget', label: 'ميزانية PoE', unit: 'W', kind: 'number', default: 50, min: 0,
        help: 'ميزانية المسجّل أصغر من ميزانية السويچ عادةً — وهذا الي يفاجئ الفني.' },
    ],
    danger: 'الحدود الثلاثة (قنوات · إدخال · PoE) تنكسر بصمت: المنظومة تشتغل جزئياً وتبدو سليمة.',
  },
  {
    id: 'cctv_monitor', domain: 'cctv', name: 'شاشة عرض', model: 'شاشة مراقبة', symbol: 'monitor',
    w: 100, h: 70,
    geo3d: { sizeM: { w: 0.6, h: 0.36, d: 0.05 }, bodyColorHex: '#1e293b', faceColorHex: '#0b1220', features: [{ kind: 'screen', x: 0.5, y: 0.45, w: 0.86, h: 0.72 }] },
    about: 'شاشة العرض — ما تأثّر على التسجيل، بس بلاها ماكو مراقبة حيّة.',
    ports: [{ id: 'hdmi', label: 'HDMI', kind: 'signal', x: 0.5, y: 1 }],
    params: [{ id: 'name', label: 'الاسم', kind: 'text', default: 'شاشة' }],
  },
]


// ═══ التحكم بالدخول والأقفال ═══
//
// ⚠️ **رقم الباب خاصية على كل قطعة** — لأن الفحص يشتغل على «الباب»
// مو على «القطعة». بلاه، زر خروج على باب A يبين كأنه يغطّي باب B،
// والمنظومة تمر بفحص وهي فيها باب بلا مخرج. وهذا بالضبط شلون تنكتب
// مخططات التحكم بالدخول بالميدان: كل شي مؤشّر عليه رقم بابه.
const doorParam = { id: 'door', label: 'رقم الباب', kind: 'number' as const, default: 1, min: 1,
  help: 'كل قطع الباب الواحد لازم نفس الرقم — الفحوص تنحسب لكل باب على حدة.' }

const ACCESS: PartDef[] = [
  {
    id: 'ac_controller', domain: 'access', name: 'وحدة تحكم بالدخول', model: 'وحدة بابين · ريلايات', symbol: 'ac_ctrl',
    w: 140, h: 66,
    geo3d: { sizeM: { w: 0.26, h: 0.19, d: 0.06 }, bodyColorHex: '#334155', faceColorHex: '#1e293b', features: [{ kind: 'terminalPlate', x: 0.5, y: 0.7, w: 0.8, h: 0.3 }, { kind: 'statusLed', x: 0.88, y: 0.2 }] },
    about: 'تقرأ البطاقة وتقرّر، وريلايها يقطع تغذية القفل. ⚠️ عدنا الريلاي يقطع **الموجب** مو السالب.',
    ports: [
      { id: 'pwr', label: 'تغذية', kind: 'dc', x: 0.1, y: 0 },
      { id: 'lock1', label: 'قفل', kind: 'dc', x: 0.3, y: 1 },
      { id: 'rdr1', label: 'قارئ', kind: 'signal', x: 0.55, y: 1 },
      { id: 'rex1', label: 'خروج', kind: 'signal', x: 0.78, y: 1 },
    ],
    params: [
      { id: 'name', label: 'الاسم', kind: 'text', default: 'وحدة التحكم' },
      { id: 'standbyMa', label: 'سحب الوحدة', unit: 'mA', kind: 'number', default: 120, min: 0 },
    ],
  },
  {
    id: 'mag_lock', domain: 'access', name: 'قفل مغناطيسي', model: 'fail-safe · يسحب باستمرار', symbol: 'mag_lock',
    w: 120, h: 56,
    geo3d: { sizeM: { w: 0.25, h: 0.04, d: 0.04 }, bodyColorHex: '#94a3b8', faceColorHex: '#64748b', features: [{ kind: 'terminalPlate', x: 0.85, y: 0.5, w: 0.2, h: 0.6 }] },
    about: 'يحتاج كهرباء حتى **يقفل** — فينفتح بانقطاع التيار. هذا المطلوب على مخارج الطوارئ.',
    ports: [{ id: 'in', label: 'تغذية', kind: 'dc', x: 0.5, y: 0 }],
    params: [
      { id: 'name', label: 'اسم الباب', kind: 'text', default: 'الباب الرئيسي' },
      doorParam,
      // ⚠️ السؤال **بلغة الميدان**: الفني يعرف إذا الباب مخرج طوارئ،
      // وما يعرف بالضرورة شنو يعني fail-safe. المحاكي هو الي يستنتج.
      { id: 'isEgress', label: 'الباب مخرج طوارئ؟', kind: 'bool', default: false,
        help: 'مخرج الطوارئ لازم قفله ينفصل بانقطاع التيار — وإلا يحبس الناس.' },
      { id: 'holdMa', label: 'سحب الإمساك', unit: 'mA', kind: 'number', default: 500, min: 0,
        help: 'يسحبه **٢٤ ساعة** — لأنه يسحب حتى يبقى مقفلاً.' },
      { id: 'minV', label: 'أقل جهد للإمساك', unit: 'V', kind: 'number', default: 10.5, min: 0 },
      { id: 'diode', label: 'دايود على الملف', kind: 'bool', default: true,
        help: 'بلاه، نبضة الرجوع تاكل تلامسات الريلاي — والعطل يظهر بعد أشهر.' },
    ],
    danger: 'على مخرج طوارئ لازم يكون هذا النوع — والنوع الثاني يحبس الناس بالحريق.',
  },
  {
    id: 'electric_strike', domain: 'access', name: 'قفل كهربائي (استرايك)', model: 'fail-secure · يسحب لحظة الفتح', symbol: 'strike',
    w: 110, h: 56,
    geo3d: { sizeM: { w: 0.03, h: 0.13, d: 0.03 }, bodyColorHex: '#cbd5e1', faceColorHex: '#94a3b8', features: [] },
    about: 'يحتاج كهرباء حتى **يفتح** — فيبقى مقفلاً بانقطاع التيار. ⚠️ ما ينفع على مخرج طوارئ.',
    ports: [{ id: 'in', label: 'تغذية', kind: 'dc', x: 0.5, y: 0 }],
    params: [
      { id: 'name', label: 'اسم الباب', kind: 'text', default: 'باب المخزن' },
      doorParam,
      { id: 'isEgress', label: 'الباب مخرج طوارئ؟', kind: 'bool', default: false },
      { id: 'pulseMa', label: 'سحب الفتح', unit: 'mA', kind: 'number', default: 350, min: 0,
        help: 'لحظياً بس — ساكن بلا تيار، عكس المغناطيسي.' },
      { id: 'minV', label: 'أقل جهد للفتح', unit: 'V', kind: 'number', default: 10, min: 0 },
      { id: 'diode', label: 'دايود على الملف', kind: 'bool', default: true },
    ],
    danger: 'يبقى **مقفلاً** بانقطاع التيار — ممنوع على مخارج الطوارئ.',
  },
  {
    id: 'card_reader', domain: 'access', name: 'قارئ بطاقات', model: 'Wiegand · حد ~١٥٠ متر', symbol: 'reader',
    w: 80, h: 66,
    geo3d: { sizeM: { w: 0.08, h: 0.12, d: 0.02 }, bodyColorHex: '#1e293b', faceColorHex: '#0f172a', features: [{ kind: 'statusLed', x: 0.5, y: 0.2 }] },
    about: 'يقرا البطاقة ويرسلها للوحدة. ⚠️ فوگ ~١٥٠ متر الإشارة تصير متقطعة.',
    ports: [{ id: 'data', label: 'بيانات', kind: 'signal', x: 0.5, y: 1 }],
    params: [
      { id: 'name', label: 'الاسم', kind: 'text', default: 'قارئ' },
      doorParam,
      { id: 'standbyMa', label: 'سحب القارئ', unit: 'mA', kind: 'number', default: 60, min: 0 },
    ],
  },
  {
    id: 'exit_button', domain: 'access', name: 'زر خروج', model: 'تماس جاف', symbol: 'exit_btn',
    w: 76, h: 56,
    geo3d: { sizeM: { w: 0.08, h: 0.08, d: 0.02 }, bodyColorHex: '#e2e8f0', faceColorHex: '#22c55e', features: [{ kind: 'disc', x: 0.5, y: 0.5, r: 0.3 }] },
    about: 'الخروج من جوّا بلا بطاقة — ⚠️ إلزامي على كل باب.',
    ports: [{ id: 'out', label: 'تماس', kind: 'signal', x: 0.5, y: 1 }],
    params: [{ id: 'name', label: 'الاسم', kind: 'text', default: 'زر خروج' }, doorParam],
  },
  {
    id: 'rex_motion', domain: 'access', name: 'حسّاس خروج', model: 'حركة · فوگ الباب', symbol: 'rex',
    w: 84, h: 56,
    geo3d: { sizeM: { w: 0.11, h: 0.05, d: 0.05 }, bodyColorHex: '#f1f5f9', faceColorHex: '#cbd5e1', features: [{ kind: 'lens', x: 0.5, y: 0.5, r: 0.25, len: 0.6 }] },
    about: 'يفتح تلقائياً لمن أحد يقرب من جوّا — بديل زر الخروج.',
    ports: [{ id: 'out', label: 'تماس', kind: 'signal', x: 0.5, y: 1 }],
    params: [{ id: 'name', label: 'الاسم', kind: 'text', default: 'حسّاس خروج' }, doorParam],
  },
  {
    id: 'break_glass', domain: 'access', name: 'كسر زجاج للطوارئ', model: 'يقطع تغذية القفل مباشرة', symbol: 'break_glass',
    w: 80, h: 62,
    geo3d: { sizeM: { w: 0.09, h: 0.09, d: 0.03 }, bodyColorHex: '#dc2626', faceColorHex: '#991b1b', features: [{ kind: 'disc', x: 0.5, y: 0.5, r: 0.28 }] },
    about: 'آخر خط دفاع. ⚠️ لازم يقطع **تغذية القفل مباشرة** — مو يمر بوحدة التحكم.',
    ports: [{ id: 'out', label: 'قطع', kind: 'dc', x: 0.5, y: 1 }],
    params: [{ id: 'name', label: 'الاسم', kind: 'text', default: 'كسر زجاج' }, doorParam],
    danger: 'مربوط بوحدة التحكم = إحساس أمان كاذب: وحدة معلّقة تخلّيه بلا فايدة.',
  },
  {
    id: 'ac_psu', domain: 'access', name: 'مغذّي مع بطارية', model: '١٢ فولت · بطارية احتياط', symbol: 'ac_psu',
    w: 120, h: 66,
    geo3d: { sizeM: { w: 0.2, h: 0.25, d: 0.09 }, bodyColorHex: '#475569', faceColorHex: '#1e293b', features: [{ kind: 'terminalPlate', x: 0.5, y: 0.75, w: 0.7, h: 0.25 }, { kind: 'statusLed', x: 0.85, y: 0.2 }] },
    about: 'يغذّي المنظومة ويشحن البطارية. ⚠️ البطارية هي الي تقرّر شكد تصمد المنظومة بانقطاع.',
    ports: [{ id: 'out', label: 'خرج', kind: 'dc', x: 0.5, y: 1 }],
    params: [
      { id: 'name', label: 'الاسم', kind: 'text', default: 'مغذّي' },
      { id: 'voltage', label: 'الجهد', unit: 'V', kind: 'number', default: 12, min: 0 },
      { id: 'maxA', label: 'أقصى تيار', unit: 'A', kind: 'number', default: 3, min: 0 },
      { id: 'ah', label: 'سعة البطارية', unit: 'Ah', kind: 'number', default: 7, min: 0 },
    ],
  },
  {
    id: 'fire_relay', domain: 'access', name: 'تماس إنذار الحريق', model: 'يقطع الأقفال بالإنذار', symbol: 'fire_relay',
    w: 100, h: 56,
    geo3d: { sizeM: { w: 0.09, h: 0.06, d: 0.03 }, bodyColorHex: '#7f1d1d', faceColorHex: '#450a0a', features: [{ kind: 'statusLed', x: 0.5, y: 0.5 }] },
    about: 'تماس من لوحة الحريق يقطع تغذية الأقفال المغناطيسية. ⚠️ الحريق يصير **والكهرباء شغّالة**.',
    ports: [{ id: 'nc', label: 'NC', kind: 'dc', x: 0.5, y: 1 }],
    params: [
      { id: 'name', label: 'الاسم', kind: 'text', default: 'تماس الحريق' },
      { id: 'bypassed', label: 'مجسور', kind: 'bool', default: false,
        help: 'أحد يجسّره حتى «يوگف الإزعاج» — والمنظومة تبدو شغّالة تماماً.' },
    ],
    danger: 'جسره يعني الأقفال ما تفصل بالحريق — والمنظومة تبدو سليمة.',
  },
]

export const PARTS: PartDef[] = [...ELECTRICAL, ...SOLAR, ...NETWORK, ...FIRE, ...AUDIO, ...GPON, ...CCTV, ...ACCESS]

/** ═══ قطع مشتركة بين مجالين ═══
 *
 * ⚠️ **نفس الكائن بمعرّف واحد** — الكتالوگ يفهرس بالمعرّف، فالنسخة
 * الثانية چانت تدهس الأولى. الي نسويه: نخلّي الكتالوگ يعرض القطعة
 * بمجالين، بلا ما ننسخ تعريفها.
 *
 * ليش يفرق: كاميرا بخصائص مختلفة بين مختبرين تعني فنياً يتدرّب على
 * جهاز ما موجود — وأول خاصية تنضاف لوحدة وتنسى بالثانية تكسر الدرس.
 */
export const SHARED_IN_CCTV = ['ip_camera', 'switch_poe', 'switch_l2']

export const PART_BY_ID: Record<string, PartDef> = Object.fromEntries(PARTS.map((p) => [p.id, p]))

/** قطع مجال معيّن — تحسب المشتركة. */
export function partsForDomain(domain: DomainId): PartDef[] {
  const own = PARTS.filter((p) => p.domain === domain)
  if (domain !== 'cctv') return own
  return [...SHARED_IN_CCTV.map((id) => PART_BY_ID[id]).filter(Boolean), ...own]
}

export const DOMAINS: { id: DomainId; name: string; icon: string; about: string }[] = [
  { id: 'network', name: 'الشبكات', icon: '🌐', about: 'سويچات وراوترات وحاسبات وكاميرات — وصّل وهيّئ واختبر الاتصال.' },
  { id: 'solar', name: 'الطاقة الشمسية', icon: '☀️', about: 'ألواح وإنفرتر وبطاريات وأحمال — افحص الستring والتوازن قبل الميدان.' },
  { id: 'electrical', name: 'الدوائر الكهربائية', icon: '⚡', about: 'مصادر ومقاومات ولمبات ومفاتيح — الدائرة تنحل فعلاً بقوانين كيرشوف.' },
  { id: 'fire', name: 'إنذار الحريق', icon: '🔥', about: 'زونات ومقاومة نهاية وصفّارات وبطارية — اللوحة تقرا عطلاً لو الخط غلط.' },
  { id: 'audio', name: 'الصوت والإذاعة', icon: '🔊', about: 'مكبّرات وسماعات وخط ١٠٠ فولت — التحميل الزائد يحرق المكبّر.' },
  { id: 'gpon', name: 'الألياف الضوئية', icon: '🔬', about: 'OLT وسبليترات وONT — الميزانية الضوئية تقرّر منو يسجّل ومنو لا.' },
  { id: 'cctv', name: 'الكاميرات والمراقبة', icon: '📹', about: 'كاميرات ومسجّل وقرص — النطاق وأيام التخزين وسحب PoE الليلي تنحسب فعلاً.' },
  { id: 'access', name: 'الأقفال والتحكم بالدخول', icon: '🔐', about: 'أبواب وأقفال وقارئات — والفحص يقرا كل باب على حدة، لأن باباً واحداً غلط يكفي.' },
]

/** القيم الابتدائية لقطعة — تنسخ من الكتالوگ لمن تنحط باللوح. */
export function defaultParams(part: PartDef): Record<string, string | number | boolean> {
  return Object.fromEntries(part.params.map((p) => [p.id, p.default]))
}
