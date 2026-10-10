// ═══ واجهة وحدة المشترك الضوئية (ONT) ═══
//
// ⚠️⚠️ **مو واجهة مصنّع بعينه**: هيكل عام لواجهات الONT الشائعة —
// نفس التبويبات ونفس الحقول الي يلگاها الفني بأغلب الأجهزة. الأسماء
// وصفية، والأرقام قياسية منشورة. `verified = FALSE` مثل بقية
// المحتوى لحد ما يجرّبها فني على جهاز حقيقي.
//
// ⚠️ الترتيب مقصود: **الحالة أول**. الفني الي يفتح واجهة ONT عند
// زبون أول شي يشوفه هو «هل الجهاز مسجّل؟ وكم القدرة الواصلة؟» —
// وهذولا يقرّران إذا المشكلة بالليف أو بالإعدادات. واجهة تبدي
// بتبويب WAN تخلّيه يعدّل إعدادات وهو ما يعرف إذا الضوء أصلاً واصل.

import type { PanelSchema } from './schema'

export const ONT_PANEL: PanelSchema = {
  id: 'panel_ont_generic',
  name: 'راوتر ضوئي — وحدة مشترك',
  brandLine: 'GPON ONT · WiFi + LAN + VoIP',
  address: 'http://192.168.100.1',
  warn: 'واجهة تدريبية — هيكل عام لواجهات الONT الشائعة، مو واجهة موديل بعينه.',
  tabs: [
    {
      id: 'status',
      label: 'الحالة',
      sections: [
        {
          title: 'حالة الوصلة الضوئية',
          note: 'هذي القيم **تُقاس** — الجهاز يقراها من مستقبله الضوئي، وما تنكتب بالإعدادات.',
          fields: [
            { id: 'reg', label: 'حالة التسجيل بالOLT', kind: 'readonly', computed: 'ponStatus' },
            { id: 'rx', label: 'قدرة الاستقبال', kind: 'readonly', computed: 'rxDbm', unit: 'dBm',
              help: 'المدى السليم بين −٢٧ و−٨. تحته ما يسجّل، وفوگه يشبع المستقبل.' },
            { id: 'loss', label: 'الفقد الكلي على المسار', kind: 'readonly', computed: 'lossDb', unit: 'dB' },
            { id: 'splits', label: 'عدد السبليترات بالمسار', kind: 'readonly', computed: 'splitters' },
          ],
        },
        {
          title: 'حالة الخدمة',
          fields: [
            { id: 'svc', label: 'الإنترنت', kind: 'readonly', computed: 'serviceStatus' },
            { id: 'wanip', label: 'عنوان WAN', kind: 'readonly', computed: 'wanIp' },
          ],
        },
      ],
    },
    {
      id: 'wan',
      label: 'WAN',
      sections: [
        {
          title: 'إعدادات الاتصال',
          note: '⚠️ التسجيل الضوئي شي والخدمة شي: الضوء يصير أخضر وأنت بلا إنترنت لو هذي الإعدادات غلط.',
          fields: [
            {
              id: 'mode', label: 'وضع الاتصال', kind: 'select', path: 'wanMode',
              options: [
                { value: 'pppoe', label: 'PPPoE' },
                { value: 'dhcp', label: 'DHCP (تلقائي)' },
                { value: 'static', label: 'عنوان ثابت' },
              ],
            },
            { id: 'user', label: 'اسم المستخدم', kind: 'text', path: 'pppoeUser',
              placeholder: 'user@isp', showWhen: { path: 'wanMode', equals: 'pppoe' },
              help: 'يجي من مزوّد الخدمة — بلاه الجهاز يسجّل وما ينطي إنترنت.' },
            { id: 'pass', label: 'كلمة المرور', kind: 'password', path: 'pppoePass',
              showWhen: { path: 'wanMode', equals: 'pppoe' } },
            { id: 'vlan', label: 'معرّف VLAN', kind: 'number', path: 'wanVlan',
              help: '⚠️ لازم يطابق VLAN الخدمة بالOLT بالضبط — أكثر بلاغ «الليف سليم والنت ما يجي» سببه هذا.' },
          ],
        },
      ],
    },
    {
      id: 'lan',
      label: 'LAN',
      sections: [
        {
          title: 'الشبكة الداخلية',
          fields: [
            { id: 'lanIp', label: 'عنوان الجهاز', kind: 'text', path: 'lanIp', placeholder: '192.168.100.1' },
            { id: 'lanMask', label: 'قناع الشبكة', kind: 'text', path: 'lanMask', placeholder: '255.255.255.0' },
            { id: 'dhcp', label: 'خادم DHCP', kind: 'bool', path: 'dhcpOn' },
            { id: 'dhcpFrom', label: 'بداية المدى', kind: 'text', path: 'dhcpFrom', placeholder: '192.168.100.10',
              showWhen: { path: 'dhcpOn', equals: 'true' } },
            { id: 'dhcpTo', label: 'نهاية المدى', kind: 'text', path: 'dhcpTo', placeholder: '192.168.100.200',
              showWhen: { path: 'dhcpOn', equals: 'true' } },
          ],
        },
      ],
    },
    {
      id: 'wifi',
      label: 'الشبكة اللاسلكية',
      sections: [
        {
          title: 'نطاق ٢٫٤ جيجاهرتز',
          fields: [
            { id: 'w24', label: 'تفعيل', kind: 'bool', path: 'wifi24On' },
            { id: 'ssid', label: 'اسم الشبكة (SSID)', kind: 'text', path: 'ssid' },
            { id: 'wpass', label: 'كلمة المرور', kind: 'password', path: 'wifiPass',
              help: 'أقل شي ٨ محارف — والجهاز يرفض الأقصر.' },
            { id: 'ch', label: 'القناة', kind: 'select', path: 'wifiChannel',
              options: [
                { value: 'auto', label: 'تلقائي' }, { value: '1', label: '١' },
                { value: '6', label: '٦' }, { value: '11', label: '١١' },
              ],
              help: 'عدنا بالعمارات الازدحام شديد — ١ و٦ و١١ هن الوحيدات الي ما يتداخلن.' },
          ],
        },
        {
          title: 'نطاق ٥ جيجاهرتز',
          fields: [
            { id: 'w5', label: 'تفعيل', kind: 'bool', path: 'wifi5On' },
            { id: 'ssid5', label: 'اسم الشبكة (SSID)', kind: 'text', path: 'ssid5' },
          ],
        },
      ],
    },
    {
      id: 'security',
      label: 'الأمان',
      sections: [
        {
          title: 'الحماية',
          fields: [
            { id: 'fw', label: 'الجدار الناري', kind: 'bool', path: 'firewallOn' },
            { id: 'remote', label: 'الإدارة عن بُعد', kind: 'bool', path: 'remoteMgmt',
              help: '⚠️ تفعيلها يفتح واجهة الجهاز على الإنترنت — ما تنفعّل إلا بحاجة، وبكلمة مرور قوية.' },
            { id: 'admPass', label: 'كلمة مرور الجهاز', kind: 'password', path: 'adminPass',
              help: 'تركها على الافتراضية يعني أي واحد بالشبكة يدخل ويغيّر.' },
          ],
        },
      ],
    },
  ],
}
