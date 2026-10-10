// ═══ نحو سويچ على نمط VRP ═══
//
// ⚠️⚠️ **هذا مو VRP مالت هواوي ولا نسخة منها.** نحو تدريبي مكتوب من
// الأوامر **الشائعة المنشورة** بمناهج الشبكات. صور أنظمة المصنّعين
// مرخّصة وما تنشغّل برّا شروطها (المخطط ٤٠ و٤٢).
//
// ⚠️ وهذا الملف هو **الإثبات العملي** إن المعمار بيانات مو كود: محرّك
// الأوامر ما انلمس ولا سطر عشانه — بس شجرة ثانية. لو چان النحو مكتوباً
// بالكود، كل نظام جديد يعني نسخة من المحرّك.
//
// الفرق الحقيقي عن نمط IOS، والي يلخبط الفني الي يبدّل بينهم:
//   • المحث **يحيط** الاسم: `<SW>` بنمط المستخدم و`[SW]` بنمط النظام
//   • `system-view` بدل `configure terminal`
//   • `quit` بدل `exit` — و`return` بدل `end`
//   • `sysname` بدل `hostname`
//   • المنفذ ينهيّأ بخطوتين: `port link-type access` ثم
//     `port default vlan 20` — مو أمراً واحداً
//   • `display` بدل `show`
//   • أسماء المنافذ بثلاث خانات `GigabitEthernet0/0/1` مو خانتين

import type { CliGrammar } from './grammar'

export const HUAWEI_VRP: CliGrammar = {
  id: 'simcli_vrp_like_switch',
  name: 'Training VRP-style OS',
  os: 'V200R-TRAINING',
  startMode: 'user',
  showStyle: 'vrp',
  banner: [
    '  ⚠️  نظام تدريبي — مو صورة مصنّع. الأوامر شائعة ومنشورة.',
    '',
    'Info: The max number of VTY users is 10.',
    '',
  ],
  modes: [
    // ═══ نمط المستخدم ═══ `<SW>`
    {
      id: 'user',
      promptSuffix: '>',
      promptTemplate: '<$host>',
      root: [
        { t: 'system-view', help: 'Enter system view', enter: { mode: 'system' } },
        {
          t: 'display',
          help: 'Display information',
          children: [
            { t: 'version', help: 'Display version information', show: 'version' },
            { t: 'current-configuration', help: 'Display current configuration', show: 'running-config' },
            {
              t: 'vlan',
              help: 'Display VLAN information',
              children: [{ t: 'brief', help: 'Brief VLAN information', show: 'vlan-brief' }],
            },
          ],
        },
        { t: 'quit', help: 'Exit from current mode', exit: true },
      ],
    },

    // ═══ نمط النظام ═══ `[SW]`
    {
      id: 'system',
      promptSuffix: ']',
      promptTemplate: '[$host]',
      root: [
        {
          t: 'sysname',
          help: 'Set the host name',
          children: [{ t: '<arg>', arg: 'word', help: 'Host name', set: 'hostname', val: '$1' }],
        },
        {
          t: 'vlan',
          help: 'Create VLAN or enter VLAN view',
          children: [
            {
              t: '<arg>', arg: 'num', help: 'VLAN ID <1-4094>',
              set: 'vlans.$1.exists', val: 'true',
              enter: { mode: 'vlan-view', ctx: '$1' },
            },
          ],
        },
        {
          t: 'interface',
          help: 'Enter interface view',
          children: [
            {
              t: '<arg>', arg: 'ifname', help: 'Interface name, e.g. GigabitEthernet0/0/1',
              set: 'interfaces.$1.exists', val: 'true',
              enter: { mode: 'if-view', ctx: '$1' },
            },
          ],
        },
        {
          t: 'display',
          help: 'Display information',
          children: [
            { t: 'current-configuration', help: 'Display current configuration', show: 'running-config' },
            { t: 'this', help: 'Display configuration of current view', show: 'running-config' },
            {
              t: 'vlan',
              help: 'Display VLAN information',
              children: [{ t: 'brief', help: 'Brief VLAN information', show: 'vlan-brief' }],
            },
          ],
        },
        { t: 'save', help: 'Save current configuration', say: 'Are you sure to continue? [Y/N]:y\nIt will take several minutes to save configuration file, please wait...\nConfiguration file had been saved successfully' },
        { t: 'quit', help: 'Return to user view', exit: true },
        { t: 'return', help: 'Return to user view', endTo: 'user' },
      ],
    },

    // ═══ نمط الـVLAN ═══ `[SW-vlan20]`
    {
      id: 'vlan-view',
      promptSuffix: ']',
      promptTemplate: '[$host-vlan$ctx]',
      root: [
        {
          t: 'description',
          help: 'Specify VLAN description',
          children: [{ t: '<arg>', arg: 'word', help: 'Description', set: 'vlans.$ctx.name', val: '$1' }],
        },
        {
          t: 'name',
          help: 'Specify VLAN name',
          children: [{ t: '<arg>', arg: 'word', help: 'VLAN name', set: 'vlans.$ctx.name', val: '$1' }],
        },
        { t: 'quit', help: 'Return to system view', exit: true },
        { t: 'return', help: 'Return to user view', endTo: 'user' },
      ],
    },

    // ═══ نمط المنفذ ═══ `[SW-GigabitEthernet0/0/1]`
    {
      id: 'if-view',
      promptSuffix: ']',
      promptTemplate: '[$host-$ctx]',
      root: [
        {
          t: 'port',
          help: 'Port configuration',
          children: [
            {
              t: 'link-type',
              help: 'Set the link type of the port',
              children: [
                { t: 'access', help: 'Access port', set: 'interfaces.$ctx.mode', val: 'access' },
                { t: 'trunk', help: 'Trunk port', set: 'interfaces.$ctx.mode', val: 'trunk' },
                { t: 'hybrid', help: 'Hybrid port', set: 'interfaces.$ctx.mode', val: 'hybrid' },
              ],
            },
            {
              t: 'default',
              help: 'Set the default VLAN of the port',
              children: [
                {
                  t: 'vlan',
                  help: 'Default VLAN of the access port',
                  children: [{ t: '<arg>', arg: 'num', help: 'VLAN ID <1-4094>', set: 'interfaces.$ctx.accessVlan', val: '$1' }],
                },
              ],
            },
            {
              t: 'trunk',
              help: 'Trunk port configuration',
              children: [
                {
                  t: 'allow-pass',
                  help: 'Allow VLANs to pass',
                  children: [
                    {
                      t: 'vlan',
                      help: 'VLANs allowed on the trunk',
                      children: [{ t: '<arg>', arg: 'vlanlist', help: 'VLAN IDs', set: 'interfaces.$ctx.trunkVlans', val: '$1' }],
                    },
                  ],
                },
              ],
            },
          ],
        },
        {
          t: 'description',
          help: 'Specify the description of the interface',
          children: [{ t: '<arg>', arg: 'word', help: 'Description', set: 'interfaces.$ctx.description', val: '$1' }],
        },
        { t: 'shutdown', help: 'Shut down the interface', set: 'interfaces.$ctx.shutdown', val: 'true' },
        { t: 'display', help: 'Display information', children: [{ t: 'this', help: 'Configuration of current view', show: 'running-config' }] },
        { t: 'quit', help: 'Return to system view', exit: true },
        { t: 'return', help: 'Return to user view', endTo: 'user' },
      ],
    },
  ],
}
