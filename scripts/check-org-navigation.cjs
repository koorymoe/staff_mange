// Real Layout + synthetic sessions/APIs; read-only fixtures, never production.
const assert = require('node:assert/strict');
const fs = require('node:fs');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
(async () => {
  const base = 'http://127.0.0.1:5199/staff_mange';
  const browser = await chromium.launch({
    executablePath: '/usr/bin/chromium',
    args: ['--no-sandbox'],
  });
  const model = fs.readFileSync('backend-go/internal/model/org.go', 'utf8');
  const unit = (id) => {
    const line = model.split('\n').find((s) => s.includes(`ID: "${id}"`));
    assert(line, `Missing seed ${id}`);
    const parse = (key) =>
      [
        ...(
          line.match(new RegExp(key + ': \\[\\]string\\{([^}]+)\\}'))?.[1] || ''
        ).matchAll(/"([^"]+)"/g),
      ].map((m) => m[1]);
    return { all: parse('Perms'), heads: parse('Head') };
  };
  const member = (id) => {
    const u = unit(id);
    return u.all.filter((p) => !u.heads.includes(p));
  };
  const ops = [
    ...new Set(
      ['ops.audit', 'ops.fleet', 'ops.staff', 'ops.bookings'].flatMap(
        (id) => unit(id).all,
      ),
    ),
  ];
  const monitorSource = fs
    .readFileSync('backend-go/internal/model/permission.go', 'utf8')
    .split('\n')
    .find((s) => /^\s*"MONITOR":/.test(s));
  const monitor = [...monitorSource.matchAll(/"([^"]+)"/g)]
    .slice(1)
    .map((m) => m[1]);
  const cases = [
    {
      key: 'owner',
      name: 'مالك',
      role: 'ADMIN',
      actualRole: 'OWNER',
      permissions: [],
    },
    {
      key: 'admin',
      name: 'أدمن',
      role: 'ADMIN',
      actualRole: 'ADMIN',
      permissions: [],
    },
    {
      key: 'ops',
      name: 'مدير عمليات',
      role: 'HR_COORDINATOR',
      permissions: ops,
      division: 'ops',
      position: 'DIRECTOR',
    },
    {
      key: 'bookings',
      name: 'عضو حجوزات',
      role: 'HR_COORDINATOR',
      permissions: member('ops.bookings'),
      division: 'ops',
      unit: 'ops.bookings',
    },
    {
      key: 'fleet',
      name: 'عضو عُدد',
      role: 'PROCUREMENT_ADMIN',
      permissions: member('ops.fleet'),
      division: 'ops',
      unit: 'ops.fleet',
    },
    { key: 'monitor', name: 'مراقب', role: 'MONITOR', permissions: monitor },
  ];
  const code = await (await fetch(base + '/src/components/Layout.tsx')).text();
  const routerPath = code.match(
    /["']([^"']*react-router-dom\.js[^"']*)["']/,
  )[1];
  const pattern =
    /const visibleItems = organizeOrgNavigation\(prune\(navItems, false, new Set\(\)\)\);/;
  assert(
    pattern.test(code),
    'Layout instrumentation must match actual prune call',
  );
  const instrumented = code.replace(
    pattern,
    `const beforeItems=prune(navItems,false,new Set());const visibleItems=organizeOrgNavigation(beforeItems);
    const flatten=(items,group='')=>items.flatMap(it=>it.children?flatten(it.children,it.label):it.divider?[]:[{item:it,to:it.to,label:it.labelFor?it.labelFor({employee,permissions:employeePermissions,gpsServiceId}):it.label,group}]);
    const before=flatten(beforeItems),after=flatten(visibleItems);
    window.__sidebarSnapshot={before:before.map(({item,...row})=>row),after:after.map(({item,...row})=>row),permissions:employeePermissions,originalLeaves:after.every(({item})=>before.some(b=>b.item===item)),groups:visibleItems.filter(x=>x.children).map(x=>({to:x.to,label:x.label,children:flatten(x.children).map(x=>x.to)}))};
    window.__expandSidebar=()=>{const pairs=[];const walk=items=>items.forEach(it=>{if(it.children){pairs.push([it.to||it.label,true]);walk(it.children)}});walk(visibleItems);setExpandedGroups(Object.fromEntries(pairs))};
    window.__setSidebarGroups=keys=>setExpandedGroups(Object.fromEntries(keys.map(key=>[key,true])));
  `,
  );
  const errors = [],
    reports = [];
  const stubs = [
    'PrivacyPolicyGate',
    'EmployeeAvatar',
    'LiveAlerts',
    'ThemeToggle',
    'MatrixEye',
    'MatrixCourier',
    'RemoteTimerChip',
    'GlobalSearch',
    'LastBookingPrompt',
    'AutoCheckoutPrompt',
    'JustificationPrompt',
    'MyExamBanner',
    'AnnouncementTicker',
    'SettingsPanel',
  ];
  for (const c of cases) {
    const actor = {
      id: 'nav-' + c.key,
      name: c.name + ' تجريبي',
      role: c.role,
      actualRole: c.actualRole || c.role,
      status: 'ACTIVE',
      isLeader: false,
    };
    const page = await browser.newPage({
        viewport: { width: 1366, height: 900 },
      }),
      calls = [];
    page.on('pageerror', (e) => errors.push(e.message));
    await page.addInitScript((actor) => {
      localStorage.setItem('authToken', 'synthetic-nav-token');
      localStorage.setItem('currentEmployee', JSON.stringify(actor));
      window.prompt = window.confirm = () => {
        throw Error('Native dialogs forbidden');
      };
    }, actor);
    await page.route('**/src/components/*.tsx*', (r) => {
      const name = new URL(r.request().url()).pathname
        .split('/')
        .pop()
        .split('.')[0];
      if (name === 'Layout')
        return r.fulfill({
          contentType: 'text/javascript',
          body: instrumented,
        });
      if (stubs.includes(name))
        return r.fulfill({
          contentType: 'text/javascript',
          body: 'export default function Stub(){return null}',
        });
      if (['ErrorBoundary', 'AttendanceGate'].includes(name))
        return r.fulfill({
          contentType: 'text/javascript',
          body: 'export default function Stub({children}){return children}',
        });
      if (name === 'InspectBar')
        return r.fulfill({
          contentType: 'text/javascript',
          body: 'export function InspectBar(){return null}',
        });
      return r.continue();
    });
    await page.route('**/api/**', (r) => {
      const req = r.request(),
        path = new URL(req.url()).pathname;
      calls.push(path);
      assert.equal(
        req.method(),
        'GET',
        'Sidebar grouping must not mutate anything',
      );
      let data;
      if (path === '/api/auth/me') data = actor;
      else if (path === '/api/permissions/employee/' + actor.id)
        data = c.permissions.map((name) => ({ name, label: name }));
      else if (path === '/api/org/me')
        data = c.division
          ? [
              {
                id: 'test-assignment',
                employeeId: actor.id,
                divisionId: c.division,
                unitId: c.unit || null,
                position: c.position || 'MEMBER',
              },
            ]
          : [];
      else if (path === '/api/notifications')
        data = { notifications: [], unreadCount: 0 };
      else if (path === '/api/services') data = [];
      else if (path === '/api/ai/agent-names') data = {};
      else if (path === '/api/files/token')
        data = { token: 'synthetic-file-token' };
      else {
        errors.push('Unexpected API ' + path);
        data = {};
      }
      return r.fulfill({
        contentType: 'application/json',
        body: JSON.stringify(data),
      });
    });
    await page.route(base + '/nav-preview', (r) =>
      r.fulfill({
        contentType: 'text/html',
        body: `<html dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body><div id="root"></div><script type="module">
      import RR from '${base}/@react-refresh';RR.injectIntoGlobalHook(window);window.$RefreshReg$=()=>{};window.$RefreshSig$=()=>x=>x;window.__vite_plugin_react_preamble_installed__=true;
      const R=await import('${base}/node_modules/.vite/deps/react.js'),React=R.default||R;const D=await import('${base}/node_modules/.vite/deps/react-dom_client.js'),{createRoot}=D.default||D;
      await import('${base}/src/index.css');await import('${base}/src/fonts.css');const router=await import('${routerPath}');const {default:Layout}=await import('${base}/src/components/Layout.tsx');
      createRoot(document.getElementById('root')).render(React.createElement(router.BrowserRouter,{basename:'/staff_mange'},React.createElement(router.Routes,null,React.createElement(router.Route,{element:React.createElement(Layout)},React.createElement(router.Route,{path:'*',element:React.createElement('h1',null,'فحص القائمة — بيانات مصطنعة')})))));
      </script></body></html>`,
      }),
    );
    await page.goto(base + '/nav-preview');
    await page.waitForFunction(
      (n) =>
        window.__sidebarSnapshot?.permissions.length === n &&
        document.querySelector('.app-sidebar nav'),
      c.permissions.length,
    );
    await page.waitForTimeout(100);
    const snap = await page.evaluate(() => window.__sidebarSnapshot);
    const unique = (rows) => [...new Set(rows.map((x) => x.to))].sort();
    assert.deepEqual(
      unique(snap.after),
      unique(snap.before),
      c.name + ': missing/new screens',
    );
    assert(snap.originalLeaves, c.name + ': leaf permission/context changed');
    if (c.key === 'owner') {
      const groups = await page.evaluate(async (base) => {
        const nav = await import(base + '/src/components/navTree.tsx');
        const context = (role, permissions) => ({
          employee: { id: 'expansion-test', role, actualRole: role },
          permissions,
          gpsServiceId: null,
          orgMemberships: [],
        });
        return [
          nav.orgNavigationGroupForPath(
            '/bookings/record-test',
            context('HR_COORDINATOR', ['view_bookings']),
          ),
          nav.orgNavigationGroupForPath(
            '/work-reports',
            context('HR_COORDINATOR', ['unit_service', 'invoice_gps']),
          ),
          nav.orgNavigationGroupForPath(
            '/my-projects',
            context('TECHNICAL', ['my_projects']),
          ),
        ];
      }, base);
      assert.deepEqual(
        groups,
        ['/unit-bookings', '/unit-technicians', undefined],
        'Deep links and personal field work use their actual visible group',
      );
    }
    assert(
      snap.groups.every((g) => g.children.length > 0),
      'No empty unit',
    );
    assert(
      !snap.groups
        .find((g) => g.to === '/unit-hr')
        ?.children.some((p) =>
          [
            '/bookings',
            '/missions',
            '/customers',
            '/bookings-archive',
          ].includes(p),
        ),
    );
    for (const path of [
      '/bookings',
      '/missions',
      '/customers',
      '/bookings-archive',
    ])
      if (snap.after.some((x) => x.to === path))
        assert(
          snap.groups
            .find((g) => g.to === '/unit-bookings')
            ?.children.includes(path),
        );
    for (const path of [
      '/audit-issues',
      '/duplicate-review',
      '/leader-invoices',
    ])
      if (snap.after.some((x) => x.to === path))
        assert(
          snap.groups
            .find((g) => g.to === '/unit-auditing')
            ?.children.includes(path),
        );
    await page.evaluate(() => window.__expandSidebar());
    const fixed = [
      { to: '/privacy-policy', label: 'سياسة الخصوصية', group: 'روابط ثابتة' },
    ];
    if (c.role === 'ADMIN')
      fixed.push({
        to: '/announcements',
        label: 'لوحة الإعلانات',
        group: 'روابط ثابتة',
      });
    if (c.actualRole === 'OWNER')
      for (const [to, label] of [
        ['/owner-security', 'لوحة المراقبة الخلفية'],
        ['/command-code', 'رمز مركز القيادة'],
        ['/owner-backups', 'النسخ الاحتياطية'],
        ['/assistant-conversations', 'محادثات المساعد الذكي'],
      ])
        fixed.push({ to, label, group: 'روابط المالك الثابتة' });
    snap.before.push(...fixed);
    snap.after.push(...fixed);
    const rendered = await page
      .locator('.app-sidebar a')
      .evaluateAll((els) =>
        [
          ...new Set(
            els.map(
              (el) =>
                (new URL(el.href).pathname.replace('/staff_mange', '') || '/') +
                new URL(el.href).search,
            ),
          ),
        ].sort(),
      );
    assert.deepEqual(
      rendered,
      unique(snap.before),
      'Actual DOM screens must match before snapshot',
    );
    reports.push({
      account: c.name,
      role: c.role,
      actualRole: actor.actualRole,
      permissions: c.permissions,
      countBefore: unique(snap.before).length,
      countAfter: unique(snap.after).length,
      before: snap.before,
      after: snap.after,
    });
    if (c.key === 'ops') {
      await page.evaluate(() => {
        window.__setSidebarGroups([]);
        history.pushState({}, '', '/staff_mange/bookings');
        window.dispatchEvent(new PopStateEvent('popstate'));
      });
      await page
        .getByRole('button', { name: 'الحجوزات والتنسيق', exact: true })
        .locator('..')
        .locator('a[href$="/bookings"]')
        .waitFor();
      assert.equal(
        await page
          .getByRole('button', { name: 'إدارة الكوادر', exact: true })
          .locator('..')
          .locator('a')
          .count(),
        0,
        'Bookings deep link must not expand the old HR group',
      );
      await page.evaluate(() =>
        window.__setSidebarGroups(['/unit-bookings', '/unit-auditing']),
      );
      await page
        .getByRole('button', { name: 'التدقيق', exact: true })
        .locator('..')
        .locator('a')
        .first()
        .waitFor();
      await page.locator('.app-sidebar nav').evaluate((el) => {
        el.scrollTop = el.scrollHeight;
      });
      await page
        .getByRole('button', { name: 'التدقيق', exact: true })
        .scrollIntoViewIfNeeded();
      await page.screenshot({ path: '/tmp/org-nav-desktop.png' });
      await page.setViewportSize({ width: 390, height: 844 });
      await page.locator('header button').first().click();
      await page
        .getByRole('button', { name: 'الحجوزات والتنسيق', exact: true })
        .scrollIntoViewIfNeeded();
      await page.screenshot({ path: '/tmp/org-nav-mobile.png' });
      assert(
        await page.evaluate(
          () => document.documentElement.scrollWidth <= innerWidth,
        ),
      );
    }
    assert(
      !calls.some((p) => p === '/api/org/chart'),
      'Grouping must not fetch chart',
    );
    console.log(
      c.name,
      unique(snap.before).length,
      '→',
      unique(snap.after).length,
    );
    await page.close();
  }
  fs.writeFileSync(
    '/tmp/org-navigation-comparison.json',
    JSON.stringify(reports, null, 2) + '\n',
  );
  assert.deepEqual(errors, []);
  await browser.close();
  console.log(
    'PASS: real Layout before/after exact unique routes and leaf guards for six synthetic accounts; OrgSeed bundles, grouping, DOM links, zero chart/writes, mobile/desktop.',
  );
})().catch((e) => {
  console.error(e);
  process.exit(1);
});
