// ═══ نحو سويچ على نمط IOS ═══
//
// ⚠️⚠️ **هذا مو IOS مالت سيسكو ولا نسخة منه.** هذا نحو تدريبي مكتوب
// من الأوامر **الشائعة المنشورة** الي أي فني شبكات يعرفها. صور أنظمة
// سيسكو مرخّصة وما تنشغّل برّا شروطها — المخطط الرئيسي يحذّر من هذا
// بالقسم ٤٠ صراحةً، والمخطط نفسه يكول «لا نبني IOS clone» (٤٢).
//
// الهدف: الفني يتعوّد على **الأسلوب** — الأنماط، الاختصار، `?`،
// وشكل الأخطاء. ولمن يجي بالميدان يشوف جهازاً حقيقياً يعرف شيسوّي.
// وأي فرق بين هذا والجهاز الحقيقي يبقى فرقاً — ولهذا التمرين يبقى
// `verified = FALSE` لحد ما فني يجرّبه على سويچ حقيقي بالورشة.
//
// ⚠️ ينتقل للخلفية (`SimCliGrammar`) لمن تجي أداة التأليف. حالياً
// مصدره هنا، والبذرة تاخذ منه — نسخة وحدة مو نسختين.

import type { CliGrammar } from './grammar'

export const CISCO_LIKE: CliGrammar = {
  id: 'simcli_ios_like_switch',
  name: 'Training Switch OS',
  os: '15.2-TRAINING',
  startMode: 'exec',
  banner: [
    '  ⚠️  نظام تدريبي — مو صورة مصنّع. الأوامر شائعة ومنشورة.',
    '',
    'User Access Verification',
    '',
  ],
  modes: [
    // ═══ EXEC للمستخدم ═══ صلاحيات القراءة بس، والمحث `>`
    {
      id: 'exec',
      promptSuffix: '>',
      root: [
        {
          t: 'enable',
          help: 'Turn on privileged commands',
          enter: { mode: 'priv' },
        },
        {
          t: 'show',
          help: 'Show running system information',
          children: [
            { t: 'version', help: 'System hardware and software status', show: 'version' },
            { t: 'clock', help: 'Display the system clock', say: '*04:12:37.115 UTC Mon Mar 1 1993' },
          ],
        },
        { t: 'exit', help: 'Exit from the EXEC', exit: true },
        { t: '?', help: 'List available commands', say: '' },
      ],
    },

    // ═══ EXEC المميّز ═══ المحث `#`
    {
      id: 'priv',
      promptSuffix: '#',
      root: [
        {
          t: 'configure',
          help: 'Enter configuration mode',
          children: [{ t: 'terminal', help: 'Configure from the terminal', enter: { mode: 'config' } }],
        },
        {
          t: 'show',
          help: 'Show running system information',
          children: [
            { t: 'running-config', help: 'Current operating configuration', show: 'running-config' },
            { t: 'version', help: 'System hardware and software status', show: 'version' },
            {
              t: 'vlan',
              help: 'VTP VLAN status',
              children: [{ t: 'brief', help: 'VTP all VLAN status in brief', show: 'vlan-brief' }],
            },
            {
              t: 'interfaces',
              help: 'Interface status and configuration',
              children: [{ t: 'status', help: 'Interface line status', show: 'interfaces-status' }],
            },
          ],
        },
        {
          t: 'write',
          help: 'Write running configuration to memory',
          children: [{ t: 'memory', help: 'Write to NV memory', say: 'Building configuration...\n[OK]' }],
        },
        { t: 'disable', help: 'Turn off privileged commands', exit: true },
        { t: 'exit', help: 'Exit from the EXEC', exit: true },
      ],
    },

    // ═══ التهيئة العامة ═══ المحث `(config)#`
    {
      id: 'config',
      promptSuffix: '(config)#',
      root: [
        {
          t: 'hostname',
          help: "Set system's network name",
          children: [{ t: '<arg>', arg: 'word', help: "This system's network name", set: 'hostname', val: '$1' }],
        },
        {
          t: 'vlan',
          help: 'VLAN commands',
          children: [
            {
              t: '<arg>',
              arg: 'num',
              help: 'ISL VLAN IDs 1-4094',
              set: 'vlans.$1.exists',
              val: 'true',
              enter: { mode: 'config-vlan', ctx: '$1' },
            },
          ],
        },
        {
          t: 'interface',
          help: 'Select an interface to configure',
          children: [
            {
              t: '<arg>',
              arg: 'ifname',
              help: 'Interface name, e.g. GigabitEthernet0/1',
              set: 'interfaces.$1.exists',
              val: 'true',
              enter: { mode: 'config-if', ctx: '$1' },
            },
          ],
        },
        { t: 'exit', help: 'Exit from configure mode', exit: true },
        { t: 'end', help: 'Exit to privileged EXEC mode', endAll: true },
      ],
    },

    // ═══ تهيئة VLAN ═══
    {
      id: 'config-vlan',
      promptSuffix: '(config-vlan)#',
      root: [
        {
          t: 'name',
          help: 'Ascii name of the VLAN',
          children: [{ t: '<arg>', arg: 'word', help: 'The ascii name for the VLAN', set: 'vlans.$ctx.name', val: '$1' }],
        },
        { t: 'exit', help: 'Exit from VLAN configuration', exit: true },
        { t: 'end', help: 'Exit to privileged EXEC mode', endAll: true },
      ],
    },

    // ═══ تهيئة المنفذ ═══ المحث `(config-if)#`
    {
      id: 'config-if',
      promptSuffix: '(config-if)#',
      root: [
        {
          t: 'switchport',
          help: 'Set switching mode characteristics',
          children: [
            {
              t: 'mode',
              help: 'Set trunking mode of the interface',
              children: [
                { t: 'access', help: 'Set trunking mode to ACCESS unconditionally', set: 'interfaces.$ctx.mode', val: 'access' },
                { t: 'trunk', help: 'Set trunking mode to TRUNK unconditionally', set: 'interfaces.$ctx.mode', val: 'trunk' },
              ],
            },
            {
              t: 'access',
              help: 'Set access mode characteristics of the interface',
              children: [
                {
                  t: 'vlan',
                  help: 'Set VLAN when interface is in access mode',
                  children: [
                    { t: '<arg>', arg: 'num', help: 'VLAN ID of the VLAN when this port is in access mode', set: 'interfaces.$ctx.accessVlan', val: '$1' },
                  ],
                },
              ],
            },
            {
              t: 'trunk',
              help: 'Set trunking characteristics of the interface',
              children: [
                {
                  t: 'allowed',
                  help: 'Set allowed VLAN characteristics',
                  children: [
                    {
                      t: 'vlan',
                      help: 'Set allowed VLANs on the trunk',
                      children: [{ t: '<arg>', arg: 'vlanlist', help: 'VLAN IDs of the allowed VLANs', set: 'interfaces.$ctx.trunkVlans', val: '$1' }],
                    },
                  ],
                },
              ],
            },
          ],
        },
        {
          t: 'description',
          help: 'Interface specific description',
          children: [{ t: '<arg>', arg: 'word', help: 'Up to 200 characters describing this interface', set: 'interfaces.$ctx.description', val: '$1' }],
        },
        { t: 'shutdown', help: 'Shutdown the selected interface', set: 'interfaces.$ctx.shutdown', val: 'true' },
        { t: 'exit', help: 'Exit from interface configuration mode', exit: true },
        { t: 'end', help: 'Exit to privileged EXEC mode', endAll: true },
      ],
    },
  ],
}
