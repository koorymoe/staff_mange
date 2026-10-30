const assert = require('node:assert/strict');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
(async () => {
  const browser = await chromium.launch({
    executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium',
    args: ['--no-sandbox'],
  });
  const base = process.env.ORG_BASE_URL || 'http://127.0.0.1:5199/staff_mange';
  const transformed = await (
    await fetch(base + '/src/pages/OrgPage.tsx')
  ).text();
  const routerPath = transformed.match(
    /["']([^"']*react-router-dom\.js[^"']*)["']/,
  )[1];
  const errors = [];
  const owner = {
    id: 'owner-org',
    name: 'مالك تجريبي',
    role: 'ADMIN',
    actualRole: 'OWNER',
  };
  const staff = {
    id: 'staff-a',
    name: 'فني تجريبي',
    role: 'TECHNICIAN',
    actualRole: 'TECHNICIAN',
  };
  async function mount(actor = owner, view = 'chart', emptyMe = false) {
    const page = await browser.newPage({
      viewport: { width: 390, height: 844 },
    });
    page.on('pageerror', (e) => {
      errors.push(e.message);
      console.log('PAGE ERROR', e.message);
    });
    const calls = [],
      flags = {
        conflict: false,
        failEnd: false,
        failChart: false,
        failEmployees: false,
        failCatalog: false,
      };
    const employees = [
      { ...owner, role: 'OWNER', status: 'ACTIVE' },
      {
        id: 'admin-org',
        name: 'مدير نظام تجريبي',
        role: 'ADMIN',
        status: 'ACTIVE',
      },
      {
        id: 'boss-org',
        name: 'مدير إدارة تجريبي',
        role: 'MONITOR',
        status: 'ACTIVE',
        photoUrl:
          'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" width="50" height="50"%3E%3Crect width="50" height="50" fill="%232563eb"/%3E%3C/svg%3E',
      },
      { ...staff, status: 'ACTIVE' },
      {
        id: 'staff-b',
        name: 'مهندس تجريبي',
        role: 'ENGINEER',
        status: 'ACTIVE',
      },
      {
        id: 'inactive',
        name: 'موظف غير فعال',
        role: 'TECHNICIAN',
        status: 'INACTIVE',
      },
    ];
    let people = [
      {
        id: 'assignment-head',
        employeeId: 'boss-org',
        employeeName: 'مدير إدارة تجريبي',
        divisionId: 'admin',
        unitId: null,
        position: 'DIRECTOR',
        allowanceIQD: 100000,
        acting: false,
        startedAt: '2026-10-29T06:00:00Z',
        note: 'بيانات مصطنعة',
      },
      {
        id: 'assignment-owner',
        employeeId: owner.id,
        employeeName: owner.name,
        divisionId: 'it',
        unitId: null,
        position: 'MEMBER',
        allowanceIQD: 0,
        acting: false,
        startedAt: '2026-10-29T06:00:00Z',
        note: '',
      },
    ];
    const myItems = [
      {
        id: 'own-item',
        title: 'بندي الآمن',
        kind: 'DEVELOPMENT',
        metric: null,
        manualProgress: 10,
        progress: 10,
        doneAt: null,
        dueDate: '2026-10-01',
        ownerId: actor.id,
        ownerName: actor.name,
        overdue: true,
        divisionName: 'الهندسية',
        unitName: 'التصميم',
        period: '2026-10',
        planTitle: 'خطة اختبار',
        planStatus: 'APPROVED',
      },
    ];
    const divisions = [
      {
        id: 'admin',
        name: 'الإدارة',
        parentId: null,
        summary: 'مراقبة كل الوحدات والحسابات والرقابة',
        permissions: ['monitoring'],
        units: [
          {
            id: 'accounts',
            divisionId: 'admin',
            name: 'الحسابات',
            permissions: ['finance'],
          },
        ],
      },
      ...[
        'العمليات',
        'تطوير الأعمال',
        'الهندسية',
        'الجودة',
        'سلاسل التوريد',
        'العقود والإعلام',
      ].map((name, i) => ({
        id: 'division-' + i,
        name,
        parentId: 'admin',
        summary: 'وصف إدارة تجريبي',
        permissions: [],
        units:
          i === 2
            ? [
                {
                  id: 'engineering-service',
                  divisionId: 'division-2',
                  name: 'خدمات هندسية',
                  permissions: ['inventory', 'unknown_test'],
                },
              ]
            : [],
      })),
      {
        id: 'it',
        name: 'التقنية IT',
        parentId: null,
        summary: 'وحدة مستقلة',
        permissions: [],
        units: [
          {
            id: 'it.main',
            divisionId: 'it',
            name: 'وحدة التقنية',
            permissions: [],
          },
        ],
      },
    ];
    const chart = () =>
      divisions.map((d) => ({
        ...d,
        people: people.filter((p) => p.divisionId === d.id && !p.unitId),
        units: d.units.map((u) => ({
          ...u,
          people: people.filter((p) => p.unitId === u.id),
        })),
      }));
    await page.addInitScript(() => {
      localStorage.setItem('authToken', 'org-test-token');
      window.prompt = window.confirm = () => {
        throw new Error('Native dialog forbidden');
      };
    });
    await page.route(base + '/org-preview', (r) =>
      r.fulfill({
        contentType: 'text/html',
        body: `<html dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body><div id="root" style="padding:12px"></div><script type="module">import RR from '${base}/@react-refresh';RR.injectIntoGlobalHook(window);window.$RefreshReg$=()=>{};window.$RefreshSig$=()=>x=>x;window.__vite_plugin_react_preamble_installed__=true;const R=await import('${base}/node_modules/.vite/deps/react.js');const React=R.default||R;const D=await import('${base}/node_modules/.vite/deps/react-dom_client.js');const {createRoot}=D.default||D;const router=await import('${routerPath}');const {SessionContext}=await import('${base}/src/session.ts');await import('${base}/src/index.css');await import('${base}/src/fonts.css');const {default:Org}=await import('${base}/src/pages/OrgPage.tsx');const {default:Me}=await import('${base}/src/components/MyOrgCard.tsx');const nav=await import('${base}/src/components/navTree.tsx');window.__navAllowed=actor=>{const ctx={employee:actor,permissions:['monitoring','auditing','unit_management','unit_hr'],gpsServiceId:null};return ['/org','/my-org'].map(path=>nav.isNavVisible(nav.navItems.find(x=>x.to===path),ctx,true));};window.__pathAllowed=actor=>nav.isPathAllowed('/org/assign',{employee:actor,permissions:['monitoring','unit_management'],gpsServiceId:null});const page=${JSON.stringify(view)}==='me'?React.createElement(Me):React.createElement(Org,{view:${JSON.stringify(view)}});createRoot(document.getElementById('root')).render(React.createElement(router.BrowserRouter,{basename:'/staff_mange'},React.createElement(SessionContext.Provider,{value:{employee:${JSON.stringify(actor)},permissions:[],setEmployee:()=>{},gpsServiceId:null}},page)));</script></body></html>`,
      }),
    );
    await page.route('**/api/**', async (r) => {
      const req = r.request(),
        path = new URL(req.url()).pathname,
        method = req.method();
      const body = req.postDataJSON();
      calls.push({ path, method, body, auth: req.headers().authorization });
      const send = (data, status = 200) =>
        r.fulfill({
          status,
          contentType: 'application/json',
          body: JSON.stringify(data),
        });
      if (path === '/api/org/chart')
        return flags.failChart
          ? send({ error: 'تعذر جلب الهيكل' }, 500)
          : send(chart());
      if (path === '/api/employees')
        return flags.failEmployees
          ? send({ error: 'تعذر جلب الموظفين' }, 500)
          : send(employees);
      if (path === '/api/permissions/catalog')
        return flags.failCatalog
          ? send({ error: 'تعذر جلب القاموس' }, 500)
          : send([
              {
                key: 'main',
                title: 'مجموعة',
                perms: [
                  { name: 'inventory', label: 'الجرد', grantable: true },
                  { name: 'finance', label: 'الحسابات', grantable: true },
                ],
              },
            ]);
      if (path === '/api/org/me')
        return send(
          emptyMe
            ? []
            : [
                {
                  ...people[0],
                  employeeId: actor.id,
                  employeeName: actor.name,
                  position:
                    actor.id === 'staff-a' ? 'MEMBER' : people[0].position,
                  divisionName: 'الإدارة',
                  unitId: 'accounts',
                  unitName: 'الحسابات',
                },
              ],
        );
      if (path === '/api/org/my-items') return send(emptyMe ? [] : myItems);
      if (path === '/api/org/plan-items/own-item' && method === 'PATCH') {
        Object.assign(myItems[0], {
          manualProgress: body.manualProgress,
          progress: body.manualProgress,
        });
        return send({ ok: true });
      }
      if (path === '/api/org/assign') {
        if (flags.conflict)
          return send({ error: 'هذا المكان بي مدير حالياً' }, 409);
        const e = employees.find((e) => e.id === body.employeeId);
        const p = {
          id: 'new-assignment',
          ...body,
          unitId: body.unitId || null,
          employeeName: e.name,
          acting: false,
          startedAt: '2026-10-29T09:00:00Z',
        };
        people.push(p);
        return send(p);
      }
      if (path.endsWith('/end')) {
        if (flags.failEnd) return send({ error: 'تعذر إنهاء المنصب' }, 500);
        const id = decodeURIComponent(path.split('/').at(-2));
        people = people.filter((p) => p.id !== id);
        return send({ ok: true });
      }
      throw new Error('Unexpected endpoint: ' + path);
    });
    await page.goto(base + '/org-preview');
    return { page, calls, flags };
  }
  const x = await mount(),
    p = x.page;
  await p.getByRole('heading', { name: 'الإدارة', exact: true }).waitFor();
  assert.deepEqual(await p.evaluate((a) => window.__navAllowed(a), owner), [
    true,
    true,
  ]);
  assert.equal(await p.locator('.org-division').count(), 8);
  assert.equal(
    await p.locator('.org-division[open]').count(),
    0,
    'Mobile starts collapsed',
  );
  await p.getByRole('heading', { name: 'الهندسية', exact: true }).click();
  await p.getByRole('button', { name: 'خدمات هندسية', exact: false }).click();
  await p.getByRole('dialog').getByText('الجرد', { exact: true }).waitFor();
  await p
    .getByRole('dialog')
    .getByText('صلاحية غير معرّفة بالقاموس (unknown_test)', { exact: true })
    .waitFor();
  await p.getByRole('button', { name: 'إغلاق النافذة' }).click();
  await p.screenshot({ path: '/tmp/org-chart-mobile.png', fullPage: true });
  await p.getByRole('button', { name: '+ تعيين جديد', exact: true }).click();
  let dialog = p.getByRole('dialog');
  await dialog.getByLabel('دوّر على موظف').fill('فني');
  await dialog.getByLabel('الموظف', { exact: true }).selectOption('staff-a');
  await dialog
    .getByLabel('الإدارة', { exact: true })
    .selectOption('division-2');
  await dialog
    .getByLabel('الوحدة (اختياري)')
    .selectOption('engineering-service');
  await dialog.getByLabel('المنصب', { exact: true }).selectOption('DIRECTOR');
  await dialog.getByLabel('المخصص بالدينار').fill('٥٠');
  await dialog.getByRole('alert').waitFor();
  assert(
    await dialog.getByRole('button', { name: 'حفظ التعيين' }).isDisabled(),
  );
  assert(!x.calls.some((c) => c.method === 'POST'));
  await dialog.getByLabel('المخصص بالدينار').fill('٥٠٠٠٠');
  assert.equal(
    await dialog.getByLabel('المخصص بالدينار').inputValue(),
    '50,000',
  );
  await dialog.getByLabel('ملاحظة').fill('اختبار تعيين');
  x.flags.conflict = true;
  await dialog.getByRole('button', { name: 'حفظ التعيين' }).click();
  await dialog
    .getByRole('alert')
    .getByText('هذا المكان بي مدير حالياً')
    .waitFor();
  assert.equal(
    await dialog.getByLabel('الموظف', { exact: true }).inputValue(),
    'staff-a',
  );
  await p.screenshot({ path: '/tmp/org-assign-dialog-mobile.png' });
  x.flags.conflict = false;
  await dialog.getByLabel('المنصب', { exact: true }).selectOption('DEPUTY');
  await dialog.getByRole('button', { name: 'حفظ التعيين' }).click();
  await p.getByRole('button', { name: 'فني تجريبي', exact: true }).waitFor();
  const post = x.calls.filter((c) => c.path === '/api/org/assign').at(-1);
  assert.deepEqual(post.body, {
    employeeId: 'staff-a',
    divisionId: 'division-2',
    unitId: 'engineering-service',
    position: 'DEPUTY',
    allowanceIQD: 50000,
    note: 'اختبار تعيين',
  });
  assert.equal(post.auth, 'Bearer org-test-token');
  await p.getByRole('button', { name: 'فني تجريبي', exact: true }).click();
  dialog = p.getByRole('dialog');
  await dialog.getByText('50,000', { exact: true }).waitFor();
  await dialog
    .getByRole('button', { name: 'إنهاء المنصب', exact: true })
    .click();
  assert(
    await dialog.getByRole('button', { name: 'تأكيد الإنهاء' }).isDisabled(),
  );
  await dialog.getByLabel('سبب الإنهاء').fill('انتهت مهمة الاختبار');
  x.flags.failEnd = true;
  await dialog.getByRole('button', { name: 'تأكيد الإنهاء' }).click();
  await dialog.getByRole('alert').waitFor();
  assert.equal(
    await dialog.getByLabel('سبب الإنهاء').inputValue(),
    'انتهت مهمة الاختبار',
  );
  x.flags.failEnd = false;
  await dialog.getByRole('button', { name: 'تأكيد الإنهاء' }).click();
  await p.getByRole('status').filter({ hasText: 'انحفظ التغيير' }).waitFor();
  await p.waitForTimeout(100);
  assert.equal(
    await p.getByRole('button', { name: 'فني تجريبي', exact: true }).count(),
    0,
  );
  assert(
    x.calls.some(
      (c) =>
        c.path.endsWith('/new-assignment/end') &&
        c.body.reason === 'انتهت مهمة الاختبار',
    ),
  );
  await p.setViewportSize({ width: 1366, height: 1000 });
  await p.locator('.org-division').evaluateAll((rows) =>
    rows.forEach((el) => {
      el.open = true;
      el.dispatchEvent(new Event('toggle'));
    }),
  );
  await p.screenshot({ path: '/tmp/org-chart-desktop.png', fullPage: true });
  await p.evaluate(() => document.documentElement.classList.add('dark'));
  await p.screenshot({
    path: '/tmp/org-chart-desktop-dark.png',
    fullPage: true,
  });
  const placement = await mount(owner, 'assign'),
    t = placement.page;
  await t.getByRole('table').waitFor();
  assert.equal(await t.locator('tbody tr').count(), 5, 'Only active employees');
  await t.getByLabel('فلتر المكان').selectOption('none');
  assert.equal(await t.locator('tbody tr').count(), 3);
  await t.getByLabel('فلتر المكان').selectOption('admin');
  assert.equal(await t.locator('tbody tr').count(), 1);
  await t.getByLabel('فلتر المكان').selectOption('');
  await t.screenshot({ path: '/tmp/org-employees-mobile.png', fullPage: true });
  const staffRow = t.getByRole('row').filter({ hasText: 'فني تجريبي' });
  await staffRow.getByRole('button', { name: 'عيّن', exact: true }).click();
  dialog = t.getByRole('dialog');
  assert.equal(
    await dialog.getByLabel('الموظف', { exact: true }).inputValue(),
    'staff-a',
  );
  await dialog.getByLabel('الإدارة', { exact: true }).selectOption('admin');
  await dialog.getByLabel('الوحدة (اختياري)').selectOption('accounts');
  await dialog.getByLabel('الإدارة', { exact: true }).selectOption('it');
  assert.equal(await dialog.getByLabel('الوحدة (اختياري)').inputValue(), '');
  await dialog.getByLabel('المخصص بالدينار').fill('0');
  await dialog.getByRole('button', { name: 'حفظ التعيين' }).click();
  await t.getByRole('table').waitFor();
  const divPost = placement.calls.find((c) => c.path === '/api/org/assign');
  assert(!Object.hasOwn(divPost.body, 'unitId'));
  assert.equal(divPost.body.allowanceIQD, 0);
  const adminActor = {
    id: 'admin-org',
    name: 'مدير نظام تجريبي',
    role: 'ADMIN',
    actualRole: 'ADMIN',
  };
  const admin = await mount(adminActor, 'assign');
  await admin.page.getByRole('table').waitFor();
  for (const name of ['مالك تجريبي', 'مدير نظام تجريبي'])
    assert(
      await admin.page
        .getByRole('row')
        .filter({ hasText: name })
        .getByRole('button', { name: 'المالك وحده' })
        .isDisabled(),
    );
  await admin.page
    .getByRole('button', { name: '+ تعيين جديد', exact: true })
    .click();
  assert.equal(
    await admin.page
      .getByRole('dialog')
      .getByLabel('الموظف', { exact: true })
      .locator('option')
      .count(),
    4,
    'Self/owner/admin excluded, only active others',
  );
  await admin.page.getByRole('button', { name: 'إلغاء', exact: true }).click();
  await admin.page
    .getByRole('row')
    .filter({ hasText: 'مالك تجريبي' })
    .getByRole('button', { name: 'التقنية IT', exact: false })
    .click();
  assert.equal(
    await admin.page
      .getByRole('dialog')
      .getByRole('button', { name: 'إنهاء المنصب', exact: true })
      .count(),
    0,
  );
  const mine = await mount(staff, 'me');
  await mine.page.getByText('الحسابات', { exact: false }).waitFor();
  await mine.page
    .getByRole('heading', { name: 'بندي الآمن', exact: true })
    .waitFor();
  assert(
    mine.calls.every((c) =>
      ['/api/org/me', '/api/org/my-items'].includes(c.path),
    ),
  );
  await mine.page.getByRole('slider').fill('65');
  await mine.page
    .getByRole('button', { name: 'احفظ نسبة بندي', exact: true })
    .click();
  await mine.page.getByRole('slider').waitFor();
  assert.equal(await mine.page.getByRole('slider').inputValue(), '65');
  assert(
    mine.calls.some(
      (c) =>
        c.method === 'PATCH' &&
        c.path === '/api/org/plan-items/own-item' &&
        c.body.manualProgress === 65,
    ),
  );
  assert(
    !mine.calls.some(
      (c) => c.path.startsWith('/api/org/desk') || c.path === '/api/org/chart',
    ),
  );
  await mine.page.screenshot({ path: '/tmp/org-my-place-mobile.png' });
  const empty = await mount(staff, 'me', true);
  await empty.page.getByText('بعدك ما انوضعت بالهيكل').waitFor();
  for (const view of ['chart', 'assign']) {
    const denied = await mount(staff, view);
    await denied.page.getByRole('alert').waitFor();
    await denied.page.waitForTimeout(100);
    assert.equal(denied.calls.length, 0, 'Staff must never probe admin API');
    assert.deepEqual(
      await denied.page.evaluate((a) => window.__navAllowed(a), staff),
      [false, true],
    );
    assert.equal(
      await denied.page.evaluate((a) => window.__pathAllowed(a), staff),
      false,
    );
  }
  for (const page of [p, t, mine.page])
    for (const width of [390, 1366]) {
      await page.setViewportSize({ width, height: 900 });
      assert(
        await page.evaluate(
          () => document.documentElement.scrollWidth <= innerWidth,
        ),
        'No page horizontal overflow',
      );
      assert(
        await page
          .locator('.org-ui input')
          .evaluateAll((rows) =>
            rows.every((el) => parseFloat(getComputedStyle(el).fontSize) >= 16),
          ),
      );
    }
  x.flags.failChart = true;
  await p.reload();
  await p.getByRole('alert').filter({ hasText: 'تعذر جلب الهيكل' }).waitFor();
  assert(
    await p
      .getByRole('button', { name: '+ تعيين جديد', exact: true })
      .isDisabled(),
  );
  x.flags.failChart = false;
  await p.getByRole('button', { name: 'أعد المحاولة' }).click();
  await p.getByRole('heading', { name: 'الإدارة', exact: true }).waitFor();
  x.flags.failEmployees = true;
  await p.reload();
  await p.getByRole('alert').filter({ hasText: 'تعذر جلب الموظفين' }).waitFor();
  assert(
    await p
      .getByRole('button', { name: '+ تعيين جديد', exact: true })
      .isDisabled(),
  );
  x.flags.failEmployees = false;
  await p.getByRole('button', { name: 'أعد المحاولة' }).click();
  await p.getByRole('heading', { name: 'الإدارة', exact: true }).waitFor();
  for (const dark of [false, true]) {
    await p.evaluate(
      (d) => document.documentElement.classList.toggle('dark', d),
      dark,
    );
    await p.evaluate(() => {
      const rgb = (value) =>
        value
          .match(/[\d.]+/g)
          .slice(0, 3)
          .map(Number);
      const lum = (value) =>
        rgb(value)
          .map((v) => v / 255)
          .map((v) => (v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4))
          .reduce((s, v, i) => s + v * [0.2126, 0.7152, 0.0722][i], 0);
      const panel = document.querySelector('.org-division');
      const css = getComputedStyle(panel);
      const muted = getComputedStyle(panel.querySelector('small')).color;
      const a = lum(css.backgroundColor),
        b = lum(muted);
      if ((Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05) < 4.5)
        throw new Error('Org muted text contrast below 4.5:1');
    });
  }
  assert.deepEqual(errors, []);
  await browser.close();
  console.log(
    'PASS: chart/assign/end/placement/security regression; own me/my-items safe reads + own progress PATCH for ordinary employee, zero desk/chart; 390/1366/dark/overflow. Synthetic data and APIs.',
  );
})().catch((e) => {
  console.error(e);
  process.exit(1);
});
