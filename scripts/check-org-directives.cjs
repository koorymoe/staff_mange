const assert = require('node:assert/strict');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
(async () => {
  const browser = await chromium.launch({
    executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium',
    args: ['--no-sandbox'],
  });
  const base =
    process.env.DIRECTIVE_BASE_URL || 'http://127.0.0.1:5199/staff_mange';
  const transformed = await (
      await fetch(base + '/src/pages/OrgDeskPage.tsx')
    ).text(),
    routerPath = transformed.match(
      /["']([^"']*react-router-dom\.js[^"']*)["']/,
    )[1];
  const errors = [];
  const owner = {
    id: 'directive-owner',
    name: 'مالك تجريبي',
    role: 'ADMIN',
    actualRole: 'OWNER',
    status: 'ACTIVE',
  };
  const employee = {
    id: 'directive-employee',
    name: 'موظف تجريبي',
    role: 'TECHNICIAN',
    actualRole: 'TECHNICIAN',
    status: 'ACTIVE',
  };
  async function mount(screen, actor = owner, tier = 'admin', audience = 3) {
    const page = await browser.newPage({
      viewport: { width: 390, height: 844 },
    });
    page.on('pageerror', (e) => {
      errors.push(e.message);
      console.log('PAGE ERROR', e.message);
    });
    const calls = [],
      flags = { fail: '', delayAck: 0 };
    const row = (
      id,
      title,
      scope = 'UNIT',
      divisionId = 'eng',
      unitId = 'design',
    ) => ({
      id,
      title,
      scope,
      divisionId,
      unitId,
      place: scope === 'COMPANY' ? 'الشركة كلها' : 'الهندسية / التصميم',
      body: 'نص توجيه مصطنع واضح ومفصل يلتزم بيه كادر الاختبار.',
      clause: 6,
      clauseCode: 'م2/ف1',
      issuedBy: 'مدير تجريبي',
      issuedAt: '2026-10-30T06:00:00Z',
      archivedAt: null,
      ackAt: null,
      acks: 1,
      audience: 4,
    });
    const mine = [
      {
        ...row('acked', 'توجيه مقرّ سابقاً'),
        ackAt: '2026-10-30T07:00:00Z',
        issuedAt: '2026-10-30T08:00:00Z',
      },
      row('pending/1 ?', 'توجيه ينتظر إقراري'),
      {
        ...row('pending-2', 'توجيه ثاني بانتظاري'),
        issuedAt: '2026-10-29T07:00:00Z',
      },
      {
        ...row('archived-mine', 'موقوف ما ينطلب إقراره'),
        archivedAt: '2026-10-30T08:00:00Z',
      },
    ];
    const managed = [
      row('design-row', 'توجيه الوحدة'),
      row('company-row', 'توجيه الشركة', 'COMPANY', null, null),
      row('division-row', 'توجيه الإدارة', 'DIVISION', 'eng', null),
      { ...row('stopped', 'توجيه موقوف'), archivedAt: '2026-10-30T08:00:00Z' },
      { ...row('zero', 'توجيه بلا جمهور'), audience: 0, acks: 0 },
      row('other', 'توجيه إدارة ثانية', 'UNIT', 'ops', 'fleet'),
    ];
    const membership =
      tier === 'none'
        ? []
        : [
            {
              id: 'post',
              employeeId: actor.id,
              employeeName: actor.name,
              divisionId: 'eng',
              divisionName: 'الهندسية',
              unitId: tier === 'unit' ? 'design' : null,
              position: 'DEPUTY',
              allowanceIQD: 0,
              startedAt: '2026-10-01T06:00:00Z',
              note: '',
              acting: false,
            },
          ];
    const chart = [
      {
        id: 'eng',
        name: 'الهندسية',
        parentId: null,
        summary: 'قسم هندسي تجريبي',
        permissions: [],
        people: tier === 'unit' ? [] : membership,
        units: [
          {
            id: 'design',
            divisionId: 'eng',
            name: 'التصميم',
            permissions: [],
            people: tier === 'unit' ? membership : [],
          },
          {
            id: 'projects',
            divisionId: 'eng',
            name: 'المشاريع',
            permissions: [],
            people: [],
          },
        ],
      },
      {
        id: 'ops',
        name: 'العمليات',
        parentId: null,
        summary: 'تشغيل',
        permissions: [],
        people: [],
        units: [
          {
            id: 'fleet',
            divisionId: 'ops',
            name: 'المركبات',
            permissions: [],
            people: [],
          },
        ],
      },
    ];
    const caps =
      tier === 'none'
        ? []
        : tier === 'unit'
          ? [
              {
                scope: 'UNIT',
                divisionId: 'eng',
                unitId: 'design',
                label: 'الهندسية / التصميم',
              },
            ]
          : tier === 'division'
            ? [
                { scope: 'DIVISION', divisionId: 'eng', label: 'الهندسية' },
                {
                  scope: 'UNIT',
                  divisionId: 'eng',
                  unitId: 'design',
                  label: 'الهندسية / التصميم',
                },
              ]
            : [
                { scope: 'COMPANY', label: 'الشركة كلها' },
                { scope: 'DIVISION', divisionId: 'eng', label: 'الهندسية' },
                {
                  scope: 'UNIT',
                  divisionId: 'eng',
                  unitId: 'design',
                  label: 'الهندسية / التصميم',
                },
              ];
    await page.addInitScript(() => {
      localStorage.setItem('authToken', 'directive-test-token');
      window.prompt = window.confirm = () => {
        throw new Error('Native dialog forbidden');
      };
    });
    await page.route(base + '/directives-preview', (r) =>
      r.fulfill({
        contentType: 'text/html',
        body: `<html dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body><div id="root" style="padding:12px"></div><script type="module">import RR from '${base}/@react-refresh';RR.injectIntoGlobalHook(window);window.$RefreshReg$=()=>{};window.$RefreshSig$=()=>x=>x;window.__vite_plugin_react_preamble_installed__=true;const R=await import('${base}/node_modules/.vite/deps/react.js');const React=R.default||R;const D=await import('${base}/node_modules/.vite/deps/react-dom_client.js');const {createRoot}=D.default||D;const router=await import('${routerPath}');const {SessionContext}=await import('${base}/src/session.ts');await import('${base}/src/index.css');await import('${base}/src/fonts.css');const file=${JSON.stringify(screen)}==='home'?'components/MyDirectives':${JSON.stringify(screen)}==='decor'?'pages/DecorTodayPage':${JSON.stringify(screen)}==='chart'?'pages/OrgPage':${JSON.stringify(screen)}==='desk'?'pages/OrgDeskPage':'components/DirectiveManager';const {default:Page}=await import('${base}/src/'+file+'.tsx');const page=React.createElement(Page,{targets:${JSON.stringify(caps)}});const content=${JSON.stringify(screen)}==='desk'?React.createElement(router.Routes,null,React.createElement(router.Route,{path:'directives-preview',element:React.createElement(router.Navigate,{to:'/org/desk/eng',replace:true})}),React.createElement(router.Route,{path:'org/desk/:divisionId',element:page})):page;createRoot(document.getElementById('root')).render(React.createElement(router.BrowserRouter,{basename:'/staff_mange'},React.createElement(SessionContext.Provider,{value:{employee:${JSON.stringify(actor)},permissions:[],setEmployee:()=>{},gpsServiceId:null}},content)));</script></body></html>`,
      }),
    );
    await page.route('**/api/**', async (r) => {
      const req = r.request(),
        u = new URL(req.url()),
        path = u.pathname,
        method = req.method(),
        body = req.postDataJSON();
      calls.push({ path, method, body, auth: req.headers().authorization });
      const send = (data, status = 200) =>
        r.fulfill({
          status,
          contentType: 'application/json',
          body: JSON.stringify(data),
        });
      if (flags.fail === path)
        return send({ error: 'فشل تجريبي بلا حفظ' }, 500);
      if (path === '/api/org/directives/mine') return send(mine);
      if (path === '/api/org/directives' && method === 'GET')
        return send(
          tier === 'none'
            ? []
            : tier === 'unit'
              ? managed.filter((d) => d.unitId === 'design')
              : tier === 'division'
                ? managed.filter((d) => d.divisionId === 'eng')
                : managed,
        );
      if (path.endsWith('/ack')) {
        assert.equal(method, 'POST');
        if (flags.delayAck)
          await new Promise((resolve) => setTimeout(resolve, flags.delayAck));
        const id = decodeURIComponent(path.split('/').at(-2));
        mine.find((d) => d.id === id).ackAt = '2026-10-30T09:00:00Z';
        return send({ ok: true });
      }
      if (path === '/api/org/directives' && method === 'POST') {
        if (body.scope === 'COMPANY') {
          assert.equal(actor.role, 'ADMIN');
          assert(
            !Object.hasOwn(body, 'divisionId') &&
              !Object.hasOwn(body, 'unitId'),
          );
        }
        if (body.scope === 'DIVISION') {
          assert(tier === 'admin' || tier === 'division');
          assert(!Object.hasOwn(body, 'unitId'));
        }
        if (body.scope === 'UNIT')
          assert(
            tier === 'admin' ||
              tier === 'division' ||
              (tier === 'unit' && body.unitId === 'design'),
          );
        managed.unshift({
          ...row(
            'created',
            'إصدار جديد',
            body.scope,
            body.divisionId || null,
            body.unitId || null,
          ),
          title: body.title,
          body: body.body,
          clause: body.clause,
        });
        return send({ id: 'created', audience });
      }
      if (path.endsWith('/archive')) {
        assert.equal(method, 'POST');
        const id = decodeURIComponent(path.split('/').at(-2));
        managed.find((d) => d.id === id).archivedAt = '2026-10-30T10:00:00Z';
        return send({ ok: true });
      }
      if (path === '/api/org/regulations')
        return send([
          {
            no: 6,
            article: 2,
            articleTitle: 'الجودة والأداء التشغيلي',
            title: 'الدقة والسرعة',
            code: 'م2/ف1',
          },
        ]);
      if (path === '/api/org/me') return send(membership);
      if (path === '/api/bookings') return send([]);
      if (path === '/api/org/chart') {
        assert.equal(actor.role, 'ADMIN');
        return send(chart);
      }
      if (path === '/api/employees') {
        assert.equal(actor.role, 'ADMIN');
        return send([actor]);
      }
      if (path === '/api/permissions/catalog') {
        assert.equal(actor.role, 'ADMIN');
        return send([]);
      }
      if (path.startsWith('/api/org/desk/'))
        return send({
          division: chart[0],
          period: u.searchParams.get('period'),
          canEdit: tier === 'admin' || tier === 'division',
          editableUnits:
            tier === 'admin' || tier === 'division'
              ? ['design', 'projects']
              : tier === 'unit'
                ? ['design']
                : [],
          canReview: tier === 'admin' || tier === 'division',
          issues: [],
          plan: null,
          unitPlans: [],
          comments: [],
          authority: [],
          metrics: [],
        });
      throw new Error('Unexpected API ' + method + ' ' + path);
    });
    await page.goto(base + '/directives-preview');
    return { page, calls, flags };
  }
  const home = await mount('home', employee, 'none'),
    p = home.page;
  await p
    .getByRole('heading', { name: 'توجيه ينتظر إقراري', exact: true })
    .waitFor();
  assert.equal(
    await p.locator('.directive-card h3').first().innerText(),
    'توجيه ينتظر إقراري',
  );
  assert.equal(
    await p
      .getByRole('heading', { name: 'موقوف ما ينطلب إقراره', exact: true })
      .count(),
    0,
  );
  assert(home.calls.every((c) => c.path === '/api/org/directives/mine'));
  await p.screenshot({
    path: '/tmp/directives-mine-mobile.png',
    fullPage: true,
  });
  const pending = p.locator('.directive-card').filter({
    has: p.getByRole('heading', { name: 'توجيه ينتظر إقراري', exact: true }),
  });
  home.flags.fail = '/api/org/directives/pending%2F1%20%3F/ack';
  await pending
    .getByRole('button', { name: 'اطّلعت وألتزم', exact: true })
    .click();
  await p.getByRole('alert').waitFor();
  assert.equal(await pending.getByText('✓ أقرّيت', { exact: true }).count(), 0);
  home.flags.fail = '';
  home.flags.delayAck = 80;
  const count = home.calls.length;
  await pending
    .getByRole('button', { name: 'اطّلعت وألتزم', exact: true })
    .evaluate((el) => {
      el.click();
      el.click();
    });
  await pending.getByText('✓ أقرّيت', { exact: true }).waitFor();
  assert.equal(
    home.calls.slice(count).filter((c) => c.path.endsWith('/ack')).length,
    1,
    'Duplicate click locked',
  );
  assert.equal(await pending.getByRole('button').count(), 0);
  assert(
    home.calls
      .filter((c) => c.method === 'POST')
      .every(
        (c) =>
          c.path.endsWith('/ack') && c.auth === 'Bearer directive-test-token',
      ),
  );
  assert(
    !home.calls.some(
      (c) =>
        c.path === '/api/org/directives' ||
        c.path.startsWith('/api/org/desk') ||
        c.path === '/api/org/chart',
    ),
  );
  const admin = await mount('manager'),
    a = admin.page;
  await a.getByRole('heading', { name: 'توجيه الوحدة', exact: true }).waitFor();
  const zero = a.locator('.directive-card').filter({
    has: a.getByRole('heading', { name: 'توجيه بلا جمهور', exact: true }),
  });
  assert.equal(
    await zero.getByRole('progressbar').evaluate((el) => el.value),
    0,
  );
  const stopped = a.locator('.directive-card').filter({
    has: a.getByRole('heading', { name: 'توجيه موقوف', exact: true }),
  });
  assert.equal(
    await stopped.getByRole('button', { name: 'إيقاف', exact: true }).count(),
    0,
  );
  await a.getByRole('button', { name: '+ إصدار توجيه', exact: true }).click();
  let d = a.getByRole('dialog');
  assert.equal(
    await d
      .getByLabel('نطاق التوجيه', { exact: true })
      .locator('option')
      .count(),
    3,
  );
  await d.getByLabel('نطاق التوجيه', { exact: true }).selectOption('UNIT');
  await d.getByLabel('عنوان التوجيه', { exact: true }).fill('توجيه تجربة وحدة');
  await d
    .getByLabel('نص التوجيه', { exact: true })
    .fill('الالتزام بتعليمات الاختبار بصورة صحيحة');
  await d.getByLabel('فقرة التوجيه', { exact: true }).selectOption('6');
  assert.equal(
    await d
      .getByLabel('فقرة التوجيه', { exact: true })
      .locator('optgroup')
      .first()
      .getAttribute('label'),
    'المادة 2: الجودة والأداء التشغيلي',
  );
  admin.flags.fail = '/api/org/directives';
  await d
    .getByRole('button', { name: 'إصدار وإرسال التوجيه', exact: true })
    .click();
  await d.getByRole('alert').waitFor();
  assert.equal(
    await d.getByLabel('عنوان التوجيه', { exact: true }).inputValue(),
    'توجيه تجربة وحدة',
  );
  admin.flags.fail = '';
  await a.screenshot({ path: '/tmp/directives-issue-mobile.png' });
  await d
    .getByRole('button', { name: 'إصدار وإرسال التوجيه', exact: true })
    .click();
  await a.getByRole('status').filter({ hasText: 'وصل لـ ٣ شخص' }).waitFor();
  const post = admin.calls
    .filter((c) => c.method === 'POST' && c.path === '/api/org/directives')
    .at(-1);
  assert.deepEqual(post.body, {
    scope: 'UNIT',
    divisionId: 'eng',
    unitId: 'design',
    title: 'توجيه تجربة وحدة',
    body: 'الالتزام بتعليمات الاختبار بصورة صحيحة',
    clause: 6,
  });
  const row = a.locator('.directive-card').filter({
    has: a.getByRole('heading', { name: 'توجيه الوحدة', exact: true }),
  });
  await row.getByRole('button', { name: 'إيقاف', exact: true }).click();
  assert(!admin.calls.some((c) => c.path.endsWith('/archive')));
  await a
    .getByRole('dialog')
    .getByRole('button', { name: 'إلغاء', exact: true })
    .click();
  assert(!admin.calls.some((c) => c.path.endsWith('/archive')));
  await row.getByRole('button', { name: 'إيقاف', exact: true }).click();
  admin.flags.fail = '/api/org/directives/design-row/archive';
  await a
    .getByRole('dialog')
    .getByRole('button', { name: 'تأكيد الإيقاف', exact: true })
    .click();
  await a.getByRole('dialog').getByRole('alert').waitFor();
  assert.equal(await row.getByText('موقوف', { exact: true }).count(), 0);
  admin.flags.fail = '';
  await a
    .getByRole('dialog')
    .getByRole('button', { name: 'تأكيد الإيقاف', exact: true })
    .click();
  await row.getByText('موقوف', { exact: true }).waitFor();
  await a.screenshot({
    path: '/tmp/directives-followup-mobile.png',
    fullPage: true,
  });
  await a.setViewportSize({ width: 1366, height: 1000 });
  await a.evaluate(() => document.documentElement.classList.add('dark'));
  await a.screenshot({
    path: '/tmp/directives-followup-dark.png',
    fullPage: true,
  });
  await a.getByRole('button', { name: '+ إصدار توجيه', exact: true }).click();
  d = a.getByRole('dialog');
  await d.getByLabel('نطاق التوجيه', { exact: true }).selectOption('UNIT');
  await d.getByLabel('نطاق التوجيه', { exact: true }).selectOption('COMPANY');
  await d
    .getByLabel('عنوان التوجيه', { exact: true })
    .fill('توجيه شركة تجريبي');
  await d
    .getByLabel('نص التوجيه', { exact: true })
    .fill('نص تعميم شركة للاختبار فقط');
  await d
    .getByRole('button', { name: 'إصدار وإرسال التوجيه', exact: true })
    .click();
  await a
    .getByRole('heading', { name: 'توجيه شركة تجريبي', exact: true })
    .waitFor();
  const companyPost = admin.calls
    .filter((c) => c.method === 'POST' && c.path === '/api/org/directives')
    .at(-1);
  assert(
    !Object.hasOwn(companyPost.body, 'divisionId') &&
      !Object.hasOwn(companyPost.body, 'unitId'),
  );
  const head = {
    id: 'head-directive',
    name: 'رئيس تجريبي',
    role: 'ENGINEER',
    actualRole: 'ENGINEER',
    status: 'ACTIVE',
  };
  const unit = await mount('desk', head, 'unit'),
    u = unit.page;
  await u.getByRole('button', { name: 'التوجيهات', exact: true }).click();
  await u.getByRole('button', { name: '+ إصدار توجيه', exact: true }).click();
  d = u.getByRole('dialog');
  assert.equal(
    await d
      .getByLabel('نطاق التوجيه', { exact: true })
      .locator('option')
      .count(),
    1,
  );
  assert.equal(
    await d.getByLabel('نطاق التوجيه', { exact: true }).inputValue(),
    'UNIT',
  );
  assert.equal(
    await d
      .getByLabel('مكان التوجيه', { exact: true })
      .locator('option')
      .count(),
    1,
  );
  await d.getByLabel('عنوان التوجيه', { exact: true }).fill('خاص بالوحدة فقط');
  await d
    .getByLabel('نص التوجيه', { exact: true })
    .fill('نص طويل لوحدة التصميم التجريبية');
  await d
    .getByRole('button', { name: 'إصدار وإرسال التوجيه', exact: true })
    .click();
  await u
    .getByRole('heading', { name: 'خاص بالوحدة فقط', exact: true })
    .waitFor();
  assert(!unit.calls.some((c) => c.path === '/api/org/chart'));
  assert(
    unit.calls.some(
      (c) =>
        c.method === 'POST' &&
        c.body?.scope === 'UNIT' &&
        c.body.unitId === 'design',
    ),
  );
  const division = await mount('desk', head, 'division'),
    v = division.page;
  await v.getByRole('button', { name: 'التوجيهات', exact: true }).click();
  await v.getByRole('button', { name: '+ إصدار توجيه', exact: true }).click();
  d = v.getByRole('dialog');
  assert.equal(
    await d
      .getByLabel('نطاق التوجيه', { exact: true })
      .locator('option')
      .count(),
    2,
  );
  assert.equal(
    await d.getByLabel('نطاق التوجيه', { exact: true }).inputValue(),
    'DIVISION',
  );
  await d.getByRole('button', { name: 'إلغاء', exact: true }).click();
  const chart = await mount('chart'),
    c = chart.page;
  await c.getByRole('button', { name: '📣 إصدار توجيه', exact: true }).click();
  await c.getByRole('button', { name: '+ إصدار توجيه', exact: true }).click();
  d = c.getByRole('dialog');
  await d.getByLabel('نطاق التوجيه', { exact: true }).selectOption('DIVISION');
  assert.equal(
    await d
      .getByLabel('مكان التوجيه', { exact: true })
      .locator('option')
      .count(),
    2,
  );
  await d.getByRole('button', { name: 'إلغاء', exact: true }).click();
  const empty = await mount('manager', employee, 'none');
  await empty.page
    .getByText('ماكو توجيهات ضمن متابعتك هسه.', { exact: true })
    .waitFor();
  assert.equal(
    await empty.page
      .getByRole('button', { name: '+ إصدار توجيه', exact: true })
      .count(),
    0,
  );
  assert(empty.calls.every((c) => c.path === '/api/org/directives'));
  home.flags.fail = '/api/org/directives/mine';
  await p.reload();
  await p.getByRole('alert').waitFor();
  home.flags.fail = '';
  await p
    .getByRole('button', { name: 'أعد جلب توجيهاتي', exact: true })
    .click();
  await p
    .getByRole('heading', { name: 'توجيه ثاني بانتظاري', exact: true })
    .waitFor();
  for (const page of [p, a, u, c])
    for (const width of [390, 1366]) {
      await page.setViewportSize({ width, height: 900 });
      assert(
        await page.evaluate(
          () => document.documentElement.scrollWidth <= innerWidth,
        ),
      );
      assert(
        await page
          .locator('.org-ui input,.org-ui select,.org-ui textarea')
          .evaluateAll((rows) =>
            rows.every((el) => parseFloat(getComputedStyle(el).fontSize) >= 16),
          ),
      );
    }
  const decor = await mount('decor', employee, 'none');
  await decor.page
    .getByRole('heading', { name: '🪚 شغلي اليوم', exact: true })
    .waitFor();
  await decor.page
    .getByRole('heading', { name: 'توجيه ينتظر إقراري', exact: true })
    .waitFor();
  assert(
    decor.calls.every(
      (c) =>
        c.path === '/api/bookings' || c.path === '/api/org/directives/mine',
    ),
  );
  for (const page of [p, a]) {
    const ratio = await page
      .locator('.directive-muted')
      .first()
      .evaluate((el) => {
        const lum = (rgb) =>
          rgb
            .match(/[\d.]+/g)
            .slice(0, 3)
            .map(Number)
            .map((n) => n / 255)
            .map((n) =>
              n <= 0.04045 ? n / 12.92 : ((n + 0.055) / 1.055) ** 2.4,
            )
            .reduce((s, n, i) => s + n * [0.2126, 0.7152, 0.0722][i], 0);
        const a = lum(getComputedStyle(el).backgroundColor),
          b = lum(getComputedStyle(el.querySelector('.directive-body')).color);
        return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);
      });
    assert(ratio >= 4.5, 'Grey cards must remain readable in both themes');
  }
  assert.equal(
    await a
      .locator('.directive-card bdi')
      .first()
      .evaluate((el) => getComputedStyle(el).direction),
    'ltr',
  );
  assert.deepEqual(errors, []);
  await browser.close();
  console.log(
    'PASS: personal pending-first/ack re-fetch/archived excluded/duplicate lock/encoded ID and retries; no privileged personal calls; admin/division/unit issuance scopes and payloads; API clauses; audience feedback; managed counts/zero audience/archive confirm/cancel/error/grey; actual desk and chart integration; no-rank empty; mobile/desktop/dark. Synthetic APIs only.',
  );
})().catch((e) => {
  console.error(e);
  process.exit(1);
});
