const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const assert = require('node:assert/strict');
(async () => {
  const browser = await chromium.launch({
    executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium',
    args: ['--no-sandbox'],
  });
  const errors = [];
  const base =
    process.env.ACCESS_BASE_URL || 'http://127.0.0.1:5199/staff_mange';
  const owner = {
    id: 'owner-test',
    name: 'مدير تجريبي',
    role: 'ADMIN',
    actualRole: 'OWNER',
  };
  const perm = (name, label) => ({ name, label, grantable: true });
  const catalog = [
    {
      key: 'monitor',
      title: 'صلاحيات المراقب',
      bundle: 'monitoring',
      perms: [
        perm('monitoring', 'حزمة المراقب'),
        perm('monitor_one', 'متابعة الحضور'),
        perm('monitor_two', 'متابعة العمل'),
      ],
    },
    {
      key: 'finance',
      title: 'الحسابات',
      perms: [
        perm('finance', 'عرض الحسابات'),
        perm('finance_audit', 'التدقيق المالي'),
      ],
    },
    {
      key: 'staff',
      title: 'الموظفين',
      perms: [perm('employee_create', 'إضافة موظف')],
    },
    {
      key: 'agents',
      title: '🤖 الوكلاء',
      perms: [
        perm('agent_matrix', 'ماتركس'),
        perm('agent_quality', 'الجودة'),
        perm('agent_booking', 'الحجوزات'),
      ],
    },
    {
      key: 'other',
      title: 'أخرى',
      perms: [
        perm('extra', 'صلاحية قديمة'),
        perm('protected', 'صلاحية مستثناة'),
        ...[
          'quotation_create',
          'quotation_edit_own',
          'quotation_manage_all',
        ].map((n, i) => perm(n, `عروض مستوى ${i}`)),
      ],
    },
  ];
  const agents = [
    { agent: 'MATRIX', name: 'ماتركس', editable: false },
    { agent: 'IRIS', name: 'ايريس', editable: false },
    { agent: 'QUALITY', name: 'ابو الجودة', editable: true },
    { agent: 'BOOKING', name: 'ابو الحجوزات', editable: true },
  ];
  const employees = [
    {
      id: 'a',
      name: 'موظف ألف',
      role: 'TECHNICIAN',
      secondaryRoles: [],
      isLeader: false,
    },
    {
      id: 'b',
      name: 'موظف باء',
      role: 'TECHNICIAN',
      secondaryRoles: [],
      isLeader: false,
    },
    {
      id: 'admin',
      name: 'مدير تجريبي',
      role: 'ADMIN',
      secondaryRoles: [],
      isLeader: false,
    },
  ];
  const fresh = (id) => ({
    employee: employees.find((e) => e.id === id),
    version: 3,
    perms:
      id === 'a'
        ? [
            { name: 'monitor_one', source: 'TEMPLATE' },
            { name: 'extra', source: 'MANUAL' },
            { name: 'protected', source: 'EXCEPTION', reason: 'مهمة خاصة' },
          ]
        : [
            { name: 'finance', source: 'TEMPLATE' },
            ...(id === 'b'
              ? [{ name: 'employee_create', source: 'MANUAL' }]
              : []),
          ],
    missingFromTemplate: id === 'a' ? ['monitor_two'] : [],
  });
  const states = Object.fromEntries(employees.map((e) => [e.id, fresh(e.id)]));
  let driftRows = [
    {
      employeeId: 'a',
      name: 'موظف ألف',
      role: 'TECHNICIAN',
      status: 'DIFF',
      extra: ['extra', 'finance_audit'],
      missing: ['monitor_two'],
      exceptions: ['protected'],
      version: 3,
    },
    {
      employeeId: 'b',
      name: 'موظف باء',
      role: 'UNASSIGNED',
      status: 'NO_TEMPLATE',
      extra: [],
      missing: [],
      exceptions: [],
      version: 3,
    },
  ];
  async function mount(actor = owner) {
    const page = await browser.newPage({
      viewport: { width: 390, height: 844 },
    });
    page.on('pageerror', (e) => {
      errors.push(e.message);
      console.log('PAGE ERROR', e.message);
    });
    const calls = [];
    let denySave = false,
      delay = false,
      finish,
      failed = false;
    await page.route(base + '/access-preview', (r) =>
      r.fulfill({
        contentType: 'text/html',
        body: `<html dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body><div id="root" style="padding:16px"></div><script type="module">import RR from '${base}/@react-refresh'; RR.injectIntoGlobalHook(window); window.$RefreshReg$=()=>{}; window.$RefreshSig$=()=>x=>x; window.__vite_plugin_react_preamble_installed__=true; const {default:React}=await import('${base}/node_modules/.vite/deps/react.js'); const D=await import('${base}/node_modules/.vite/deps/react-dom_client.js'); const {createRoot}=D.default||D; const {SessionContext}=await import('${base}/src/session.ts'); await import('${base}/src/index.css'); await import('${base}/src/fonts.css'); const {default:Page}=await import('${base}/src/pages/PermissionsPage.tsx'); createRoot(document.getElementById('root')).render(React.createElement(SessionContext.Provider,{value:{employee:${JSON.stringify(actor)},permissions:[],setEmployee:()=>{},gpsServiceId:null}},React.createElement(Page)));</script></body></html>`,
      }),
    );
    await page.route('**/api/**', async (route) => {
      const req = route.request(),
        u = new URL(req.url()),
        path = u.pathname,
        method = req.method(),
        body = req.postData() ? req.postDataJSON() : null;
      calls.push({ path, method, body, query: u.search });
      let data = [],
        status = 200;
      if (path === '/api/employees') data = employees;
      else if (path.endsWith('/admin-scope')) data = { roles: [] };
      else if (path === '/api/permissions/catalog')
        data = catalog.map((g) => ({
          ...g,
          perms: g.perms.map((p) =>
            actor.actualRole === 'OWNER'
              ? p
              : {
                  ...p,
                  grantable: p.name !== 'finance_audit',
                  lockReason:
                    p.name === 'finance_audit'
                      ? 'المالك يمنحها بيده'
                      : undefined,
                },
          ),
        }));
      else if (path === '/api/ai/agents') {
        assert(
          actor.actualRole === 'OWNER' || actor.role === 'ADMIN',
          'agent names are readable by admins on avg db1e5066',
        );
        data = agents;
      } else if (path === '/api/ai/agents/QUALITY') {
        agents[2].name = body.name;
        data = { agent: 'QUALITY', name: body.name };
      } else if (path.endsWith('/access')) {
        const id = path.split('/')[3];
        if (method === 'GET' && delay && id === 'a')
          await new Promise((r) => {
            finish = r;
          });
        if (method === 'GET' && failed && id === 'a') {
          status = 500;
          data = { error: 'فشل تجريبي بجلب الصلاحيات' };
        } else if (method === 'PUT') {
          if (denySave) {
            status = 409;
            data = { error: 'زميل عدّل صلاحيات هذا الموظف هسه — حدّث الصفحة' };
          } else {
            assert.equal(body.version, states[id].version);
            states[id].perms = states[id].perms.filter(
              (p) => !body.remove.includes(p.name),
            );
            for (const name of body.add)
              states[id].perms.push({ name, source: 'MANUAL' });
            states[id].version++;
            data = states[id];
          }
        } else data = states[id];
      } else if (path === '/api/permissions/role-drift') data = driftRows;
      else if (path.endsWith('/align')) {
        assert.equal(body.version, 3);
        driftRows[0] = {
          ...driftRows[0],
          extra: ['finance_audit'],
          missing: [],
          status: 'DIFF',
          version: 4,
        };
        data = driftRows[0];
      } else if (path === '/api/permissions/exceptions') {
        const id = body?.employeeId ?? u.searchParams.get('employeeId'),
          name = body?.permission ?? u.searchParams.get('permission');
        if (method === 'POST') {
          assert.equal(body.version, states[id].version);
          states[id].version++;
          states[id].perms = states[id].perms.filter((p) => p.name !== name);
          states[id].perms.push({
            name,
            source: 'EXCEPTION',
            reason: body.reason,
            expiresAt: body.expiresAt,
          });
        } else
          states[id].perms = states[id].perms.map((p) =>
            p.name === name ? { name, source: 'MANUAL' } : p,
          );
        data = { ok: true };
      } else if (path === '/api/permissions/audit-log')
        data = [
          {
            id: 'log',
            action: 'ADD',
            permission: 'extra',
            actorName: 'مدير تجربة',
            reason: 'سبب تجريبي',
            createdAt: '2026-10-26T08:00:00Z',
          },
        ];
      await route.fulfill({
        status,
        contentType: 'application/json',
        body: JSON.stringify(data),
      });
    });
    await page.goto(base + '/access-preview');
    return {
      page,
      calls,
      deny: () => {
        denySave = true;
      },
      allow: () => {
        denySave = false;
      },
      slow: () => {
        delay = true;
      },
      release: () => {
        delay = false;
        finish();
      },
      fail: () => {
        failed = true;
      },
      recover: () => {
        failed = false;
      },
    };
  }
  const x = await mount(),
    p = x.page;
  await p.getByLabel('اختر الموظف').selectOption('a');
  await p.getByRole('heading', { name: 'موظف ألف', exact: true }).waitFor();
  await p.getByText('صلاحيات المراقب', { exact: true }).click();
  await p
    .locator('.access-group')
    .filter({ hasText: 'صلاحيات المراقب' })
    .getByRole('button', { name: 'انطِ الحزمة كاملة', exact: true })
    .click();
  assert.equal(x.calls.filter((c) => c.method === 'PUT').length, 0);
  await p.getByRole('switch', { name: 'متابعة العمل', exact: true }).click();
  assert.equal(
    await p
      .getByRole('switch', { name: 'حزمة المراقب', exact: true })
      .getAttribute('aria-checked'),
    'false',
  );
  await p.getByRole('switch', { name: 'متابعة العمل', exact: true }).click();
  await p.getByLabel('اختر الموظف').selectOption('b');
  await p.getByRole('dialog').getByRole('button', { name: 'ابقَ هنا' }).click();
  assert.equal(await p.getByLabel('اختر الموظف').inputValue(), 'a');
  await p.getByRole('button', { name: 'راجع واحفظ' }).click();
  await p
    .getByRole('dialog')
    .getByLabel('سبب التغيير')
    .fill('مراجعة توزيع العمل');
  await p.getByRole('button', { name: 'تأكيد وحفظ' }).click();
  await p.getByText('انحفظت الصلاحيات وسجّلنا سبب التغيير.').waitFor();
  const save = x.calls.find((c) => c.method === 'PUT');
  assert.deepEqual(save.body.add, ['monitor_two']);
  assert.deepEqual(save.body.remove, []);
  assert.equal(save.body.version, 3);
  await p.getByText('الحسابات', { exact: true }).click();
  await p.getByRole('switch', { name: 'عرض الحسابات', exact: true }).click();
  await p.getByRole('button', { name: 'راجع واحفظ' }).click();
  await p.getByRole('dialog').getByLabel('سبب التغيير').fill('تغيير مع تعارض');
  x.deny();
  await p.getByRole('button', { name: 'تأكيد وحفظ' }).click();
  await p.getByRole('dialog').getByRole('alert').waitFor();
  assert(await p.getByRole('button', { name: 'تأكيد وحفظ' }).isDisabled());
  assert.equal(x.calls.filter((c) => c.method === 'PUT').length, 2);
  x.allow();
  await p
    .getByRole('dialog')
    .getByRole('button', { name: 'حدّث وراجع — يلغي المسوّدة' })
    .click();
  await p.getByRole('heading', { name: 'موظف ألف', exact: true }).waitFor();
  await p.getByText('الحسابات', { exact: true }).click();
  const financeRow = p.locator('.access-permission').filter({
    has: p.getByRole('switch', { name: 'عرض الحسابات', exact: true }),
  });
  await financeRow
    .getByRole('button', { name: 'استثناء أو صلاحية مؤقتة' })
    .click();
  await p
    .getByRole('dialog')
    .getByLabel('سبب الاستثناء')
    .fill('تكليف مؤقت بالتغطية');
  await p
    .getByRole('dialog')
    .getByLabel('تنتهي بموعد', { exact: false })
    .fill('2027-10-26T15:00');
  await p.getByRole('button', { name: 'تأكيد الاستثناء' }).click();
  await p.getByRole('heading', { name: 'موظف ألف', exact: true }).waitFor();
  const ex = x.calls.find(
    (c) => c.path === '/api/permissions/exceptions' && c.method === 'POST',
  );
  assert.equal(ex.body.expiresAt, '2027-10-26T12:00:00.000Z');
  await p.getByText('الحسابات', { exact: true }).click();
  await p
    .locator('.access-permission')
    .filter({
      has: p.getByRole('switch', { name: 'عرض الحسابات', exact: true }),
    })
    .getByRole('button', { name: 'تعديل الاستثناء / مدته' })
    .click();
  assert.equal(
    await p
      .getByRole('dialog')
      .getByLabel('تنتهي بموعد', { exact: false })
      .inputValue(),
    '2027-10-26T15:00',
  );
  assert.equal(
    await p.getByRole('dialog').getByLabel('سبب الاستثناء').inputValue(),
    'تكليف مؤقت بالتغطية',
  );
  await p
    .getByRole('dialog')
    .getByRole('button', { name: 'إلغاء', exact: true })
    .click();
  await p
    .locator('.access-permission')
    .filter({
      has: p.getByRole('switch', { name: 'عرض الحسابات', exact: true }),
    })
    .getByRole('button', { name: 'إزالة تثبيت الاستثناء' })
    .click();
  assert.equal(x.calls.filter((c) => c.method === 'DELETE').length, 0);
  await p.getByRole('button', { name: 'أزل تثبيت الاستثناء' }).click();
  await p.getByRole('heading', { name: 'موظف ألف', exact: true }).waitFor();
  assert(
    states.a.perms.some((q) => q.name === 'finance'),
    'exception removal must not revoke permission',
  );
  await p.getByText('🤖 الوكلاء', { exact: true }).click();
  await p
    .getByText('🔒 للمدير وحده — المالك ومدير النظام، بلا مفتاح منح.')
    .waitFor();
  assert.equal(await p.getByRole('switch', { name: /ايريس/ }).count(), 0);
  await p
    .locator('.access-agent')
    .filter({ hasText: 'ابو الجودة' })
    .getByRole('button', { name: 'تعديل الاسم' })
    .click();
  await p.getByRole('dialog').getByLabel('الاسم الظاهر').fill('وكيل تجريبي');
  await p.getByRole('button', { name: 'حفظ الاسم', exact: true }).click();
  await p
    .getByRole('switch', { name: 'توجيه وكيل تجريبي', exact: true })
    .waitFor();
  await p
    .getByRole('switch', { name: 'توجيه وكيل تجريبي', exact: true })
    .click();
  await p.getByRole('button', { name: 'إلغاء المسوّدة' }).click();
  await p.getByRole('tab', { name: 'مطابقة الصنف' }).click();
  await p.getByRole('button', { name: 'شوف الفرق وطابقه ويا صنفه' }).click();
  await p.getByRole('dialog').getByText('استثناءات باقية · 1').waitFor();
  await p
    .getByRole('dialog')
    .getByLabel('سبب المطابقة')
    .fill('مراجعة قالب المنصب');
  await p.getByRole('button', { name: 'تأكيد المطابقة' }).click();
  await p
    .getByText(
      'انطبقت التغييرات المسموحة. بقيت فروقات مقفولة؛ راجعها بالكتالوگ.',
    )
    .waitFor();
  const align = x.calls.find((c) => c.path.endsWith('/align'));
  assert.equal(align.body.version, 3);
  await p.getByRole('button', { name: 'شوف التفاصيل', exact: true }).click();
  assert.equal(
    await p
      .getByRole('dialog')
      .getByRole('button', { name: 'تأكيد المطابقة' })
      .count(),
    0,
  );
  await p.keyboard.press('Escape');
  await p.evaluate(() => window.scrollTo(0, 0));
  await p.screenshot({ path: '/tmp/access-drift-mobile.png' });
  assert(
    await p.evaluate(() => document.documentElement.scrollWidth <= innerWidth),
  );
  await p.getByRole('tab', { name: 'صلاحيات الموظف' }).click();
  await p.getByRole('heading', { name: 'موظف ألف', exact: true }).waitFor();
  await p.getByLabel('دوّر على صلاحية').fill('ايريس');
  await p
    .getByText('🔒 للمدير وحده — المالك ومدير النظام، بلا مفتاح منح.')
    .waitFor();
  await p.getByLabel('دوّر على صلاحية').fill('');
  await p.getByText('عروض الأسعار — اختار درجة وحدة', { exact: true }).click();
  await p.getByRole('switch', { name: 'عروض مستوى 0' }).click();
  await p.getByRole('switch', { name: 'عروض مستوى 1' }).click();
  assert.equal(
    await p
      .getByRole('switch', { name: 'عروض مستوى 0' })
      .getAttribute('aria-checked'),
    'false',
  );
  await p.getByRole('button', { name: 'إلغاء المسوّدة' }).click();
  await p.getByText('صلاحيات المراقب', { exact: true }).click();
  await p.evaluate(() => window.scrollTo(0, 0));
  await p.screenshot({ path: '/tmp/access-mobile.png' });
  const contrast = async () =>
    p.locator('.access-employee').evaluate((card) => {
      const luminance = (rgb) =>
        rgb
          .match(/[\d.]+/g)
          .slice(0, 3)
          .map(Number)
          .map((x) => x / 255)
          .map((x) => (x <= 0.04045 ? x / 12.92 : ((x + 0.055) / 1.055) ** 2.4))
          .reduce((sum, x, i) => sum + x * [0.2126, 0.7152, 0.0722][i], 0);
      const bg = luminance(getComputedStyle(card).backgroundColor);
      return ['p', '.access-eyebrow'].map((selector) => {
        const fg = luminance(
          getComputedStyle(card.querySelector(selector)).color,
        );
        return (Math.max(fg, bg) + 0.05) / (Math.min(fg, bg) + 0.05);
      });
    });
  assert(
    (await contrast()).every((ratio) => ratio >= 4.5),
    'light body and muted text contrast',
  );
  await p.setViewportSize({ width: 1366, height: 900 });
  await p.evaluate(() => document.documentElement.classList.add('dark'));
  await p.evaluate(() => window.scrollTo(0, 0));
  await p.screenshot({ path: '/tmp/access-desktop-dark.png' });
  assert(
    (await contrast()).every((ratio) => ratio >= 4.5),
    'dark body and muted text contrast',
  );
  // Admin may read agent names on current avg, never perform owner-only writes.
  const admin = await mount({ ...owner, actualRole: 'ADMIN' }),
    a = admin.page;
  await a.getByLabel('اختر الموظف').selectOption('a');
  await a.getByText('الحسابات', { exact: true }).click();
  assert(
    await a
      .getByRole('switch', { name: 'التدقيق المالي', exact: true })
      .isDisabled(),
  );
  await a.getByLabel('اختر الموظف').selectOption('admin');
  await a.getByRole('heading', { name: 'مدير تجريبي', exact: true }).waitFor();
  await a.getByText('الحسابات', { exact: true }).click();
  assert(
    await a
      .getByRole('switch', { name: 'عرض الحسابات', exact: true })
      .isDisabled(),
  );
  assert(
    !admin.calls.some(
      (c) => c.path.startsWith('/api/ai/') && c.method !== 'GET',
    ),
  );
  const race = await mount();
  race.slow();
  await race.page.getByLabel('اختر الموظف').selectOption('a');
  await race.page.getByLabel('اختر الموظف').selectOption('b');
  await race.page
    .getByRole('heading', { name: 'موظف باء', exact: true })
    .waitFor();
  race.release();
  await race.page
    .getByText('👥 الأدوار الي يكدر يضيفها ويعدّلها', { exact: true })
    .waitFor();
  assert.equal(
    await race.page
      .locator('input[type=checkbox]')
      .first()
      .evaluate((el) => getComputedStyle(el).display),
    'none',
  );
  await race.page.waitForTimeout(100);
  assert.equal(
    await race.page
      .getByRole('heading', { name: 'موظف ألف', exact: true })
      .count(),
    0,
  );
  const failed = await mount();
  failed.fail();
  await failed.page.getByLabel('اختر الموظف').selectOption('a');
  await failed.page.getByRole('alert').waitFor();
  assert.equal(await failed.page.getByRole('switch').count(), 0);
  assert.equal(
    await failed.page.getByRole('button', { name: 'راجع واحفظ' }).count(),
    0,
  );
  const denied = await mount({
    ...owner,
    role: 'TECHNICIAN',
    actualRole: null,
  });
  await denied.page.getByRole('alert').waitFor();
  assert.equal(denied.calls.length, 0);
  assert.deepEqual(errors, []);
  console.log(
    'PASS: catalog bundles/partial toggles, draft review, delta+version save, unsaved switch confirmation, 409 freeze/refresh, temporary RFC3339 Baghdad exception, metadata removal, owner-only rename, Iris lock, drift preview/confirmation/NO_TEMPLATE, admin locks/no owner writes, failed fetch no edit, late employee race, unauthorized zero API calls, mobile/desktop/dark. Synthetic APIs.',
  );
  await browser.close();
})().catch((e) => {
  console.error(e);
  process.exit(1);
});
