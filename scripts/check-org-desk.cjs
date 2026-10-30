const assert = require('node:assert/strict');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
(async () => {
  const browser = await chromium.launch({
    executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium',
    args: ['--no-sandbox'],
  });
  const base = process.env.DESK_BASE_URL || 'http://127.0.0.1:5199/staff_mange';
  const transformed = await (
    await fetch(base + '/src/pages/OrgDeskPage.tsx')
  ).text();
  const routerPath = transformed.match(
    /["']([^"']*react-router-dom\.js[^"']*)["']/,
  )[1];
  const errors = [];
  const owner = {
    id: 'desk-owner',
    name: 'مالك تجريبي',
    role: 'ADMIN',
    actualRole: 'OWNER',
  };
  const head = {
    id: 'desk-head',
    name: 'مدير تجريبي',
    role: 'ENGINEER',
    actualRole: 'ENGINEER',
  };
  async function mount(actor = owner, options = {}) {
    const page = await browser.newPage({
      viewport: { width: 390, height: 844 },
    });
    page.on('pageerror', (e) => {
      errors.push(e.message);
      console.log('PAGE ERROR', e.message);
    });
    const calls = [],
      flags = { fail: '', hideDesk: false, delayMonth: '' };
    let releaseMonth;
    const membership = options.noHead
      ? []
      : [
          {
            id: 'assignment',
            employeeId: actor.id,
            employeeName: actor.name,
            divisionId: options.root ? 'admin' : 'eng',
            divisionName: options.root ? 'الإدارة' : 'الهندسية',
            unitId: options.unit ? 'design' : null,
            position: options.member ? 'MEMBER' : 'DEPUTY',
            allowanceIQD: 0,
            startedAt: '2026-10-29T06:00:00Z',
            acting: false,
            note: '',
          },
        ];
    const units = [
      {
        id: 'design',
        divisionId: 'eng',
        name: 'التصميم',
        permissions: [],
        people: membership,
      },
    ];
    const divisions = [
      {
        id: 'eng',
        name: 'الهندسية',
        parentId: 'admin',
        summary: 'تطوير الشغل الهندسي',
        permissions: [],
        units,
        people: options.unit ? [] : membership,
      },
      {
        id: 'ops',
        name: 'العمليات',
        parentId: 'admin',
        summary: 'متابعة العمل',
        permissions: [],
        people: [],
        units: [],
      },
    ];
    let issues = [
      {
        id: 'low',
        title: 'مشكلة ثانوية',
        problem: 'تفاصيل مشكلة تجريبية ثانية',
        impact: 'تأثير محدود',
        scope: 'SPECIFIC',
        principle: null,
        severity: 'LOW',
        status: 'OPEN',
        resolution: null,
        authorName: 'نائب تجريبي',
        unitId: null,
        createdAt: '2026-10-29T06:00:00Z',
      },
      {
        id: 'high',
        title: 'تأخير المشاريع',
        problem: 'تأخير تجريبي بمواعيد المشاريع',
        impact: 'تأخير التسليم للزبون',
        scope: 'GENERAL',
        principle: 6,
        severity: 'HIGH',
        status: 'OPEN',
        resolution: null,
        authorName: 'مدير تجريبي',
        unitId: 'design',
        createdAt: '2026-10-29T06:00:00Z',
      },
    ];
    let items = [
      {
        id: 'matrix-item',
        kind: 'PROBLEM',
        title: 'تقليل التأخير',
        issueId: 'high',
        metric: 'late',
        baseline: 12,
        target: 2,
        current: 6,
        progress: 60,
        byMatrix: true,
        manualProgress: 0,
        doneAt: null,
      },
      {
        id: 'manual-item',
        kind: 'DEVELOPMENT',
        title: 'تطوير دليل العمل',
        issueId: null,
        metric: null,
        baseline: null,
        target: null,
        current: null,
        progress: 20,
        byMatrix: false,
        manualProgress: 20,
        doneAt: null,
      },
    ];
    let authority = [
        { decision: 'شراء العُدد', mode: 'DECIDE', limitIQD: 100000 },
      ],
      comments = [];
    let hasPlan = !options.noPlan;
    const metrics = [
      { key: 'late', label: 'نسبة التأخير', lowerBetter: true, unit: '%' },
      { key: 'score', label: 'معدل التقييم', lowerBetter: false, unit: '%' },
    ];
    await page.addInitScript(() => {
      localStorage.setItem('authToken', 'desk-test-token');
      window.prompt = window.confirm = () => {
        throw new Error('Native dialogs forbidden');
      };
    });
    await page.route(base + '/org/desk**', (r) =>
      r.fulfill({
        contentType: 'text/html',
        body: `<html dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body><div id="root" style="padding:12px"></div><script type="module">import RR from '${base}/@react-refresh';RR.injectIntoGlobalHook(window);window.$RefreshReg$=()=>{};window.$RefreshSig$=()=>x=>x;window.__vite_plugin_react_preamble_installed__=true;const R=await import('${base}/node_modules/.vite/deps/react.js');const React=R.default||R;const D=await import('${base}/node_modules/.vite/deps/react-dom_client.js');const {createRoot}=D.default||D;const router=await import('${routerPath}');const {SessionContext}=await import('${base}/src/session.ts');await import('${base}/src/index.css');await import('${base}/src/fonts.css');const {default:Page}=await import('${base}/src/pages/OrgDeskPage.tsx');const nav=await import('${base}/src/components/navTree.tsx');window.__deskVisible=(actor,items)=>nav.isNavVisible(nav.navItems.find(x=>x.to==='/org/desk'),{employee:actor,permissions:['monitoring','unit_management'],gpsServiceId:null,orgMemberships:items},true);window.__pathAllowed=(actor,items,path)=>nav.isPathAllowed(path,{employee:actor,permissions:['monitoring','unit_management'],gpsServiceId:null,orgMemberships:items});const content=React.createElement(router.Routes,null,React.createElement(router.Route,{path:'org/desk',element:React.createElement(Page)}),React.createElement(router.Route,{path:'org/desk/:divisionId',element:React.createElement(Page)}));createRoot(document.getElementById('root')).render(React.createElement(router.BrowserRouter,{basename:'/staff_mange'},React.createElement(SessionContext.Provider,{value:{employee:${JSON.stringify(actor)},permissions:[],setEmployee:()=>{},gpsServiceId:null}},content)));</script></body></html>`,
      }),
    );
    await page.route('**/api/**', async (r) => {
      const req = r.request(),
        url = new URL(req.url()),
        path = url.pathname,
        method = req.method(),
        body = req.postDataJSON();
      calls.push({
        path,
        method,
        body,
        period: url.searchParams.get('period'),
        auth: req.headers().authorization,
      });
      const send = (data, status = 200) =>
        r.fulfill({
          status,
          contentType: 'application/json',
          body: JSON.stringify(data),
        });
      if (path === '/api/org/me') return send(membership);
      if (path === '/api/org/chart') {
        assert(actor.role === 'ADMIN', 'No chart for non-admin heads');
        return send(divisions);
      }
      if (path.startsWith('/api/org/desk/')) {
        if (flags.hideDesk) return send({ error: 'مو موجود' }, 404);
        const id = decodeURIComponent(path.split('/').at(-1)),
          period = url.searchParams.get('period');
        const data = {
          division: divisions.find((d) => d.id === id) || {
            ...divisions[0],
            id,
            name: id === 'admin' ? 'الإدارة' : id,
          },
          period,
          canEdit:
            !options.unit &&
            (!options.root || id === 'admin') &&
            !options.readOnly,
          issues,
          plan: hasPlan
            ? {
                id: 'plan-test',
                period,
                title: 'خطة ' + period,
                progress: items.length
                  ? Math.round(
                      items.reduce((sum, item) => sum + item.progress, 0) /
                        items.length,
                    )
                  : 0,
                items: items.map((i) => ({
                  ...i,
                  current:
                    period === '2026-09' && i.byMatrix ? null : i.current,
                })),
              }
            : null,
          comments,
          authority,
          metrics,
        };
        const snapshot = JSON.parse(JSON.stringify(data));
        if (period === flags.delayMonth)
          await new Promise((resolve) => {
            releaseMonth = resolve;
          });
        return send(snapshot);
      }
      if (flags.fail === path)
        return send({ error: 'تعذر الحفظ التجريبي' }, 500);
      if (path === '/api/org/issues' && method === 'POST') {
        issues.push({
          id: 'new-issue',
          ...body,
          unitId: body.unitId || null,
          status: 'OPEN',
          resolution: null,
          authorName: 'مدير تجريبي',
          createdAt: '2026-10-29T07:00:00Z',
        });
        return send({ id: 'new-issue' });
      }
      if (path.startsWith('/api/org/issues/') && method === 'PATCH') {
        Object.assign(
          issues.find((i) => i.id === path.split('/').at(-1)),
          body,
        );
        return send({ ok: true });
      }
      if (path === '/api/org/plans') {
        hasPlan = true;
        items = [];
        return send({ id: 'plan-test' });
      }
      if (path.endsWith('/items')) {
        if (body.metric === 'late' && body.target >= 12)
          return send(
            { error: 'الهدف لازم يكون أحسن من الرقم الحالي (12)' },
            400,
          );
        items.push({
          id: 'new-item',
          ...body,
          issueId: body.issueId || null,
          metric: body.metric || null,
          baseline: body.metric ? 12 : null,
          target: body.target ?? null,
          current: body.metric ? 6 : null,
          progress: 0,
          byMatrix: !!body.metric,
          manualProgress: 0,
          doneAt: null,
        });
        return send({ id: 'new-item' });
      }
      if (path.startsWith('/api/org/plan-items/')) {
        const i = items.find((i) => i.id === path.split('/').at(-1));
        if (body.manualProgress != null) {
          assert(!i.byMatrix);
          i.manualProgress = body.manualProgress;
          i.progress = body.manualProgress;
        }
        if (body.done) {
          i.doneAt = '2026-10-29T08:00:00Z';
          i.progress = 100;
        }
        return send({ ok: true });
      }
      if (path === '/api/org/comments') {
        comments.push({
          ...body,
          authorName: 'نائب تجريبي',
          createdAt: '2026-10-29T08:00:00Z',
        });
        return send({ ok: true });
      }
      if (path.startsWith('/api/org/authority/')) {
        assert(actor.role === 'ADMIN');
        assert.equal(method, 'PUT');
        authority = body;
        return send({ ok: true });
      }
      throw new Error('Unexpected API ' + method + ' ' + path);
    });
    await page.goto(
      base +
        '/org/desk' +
        (options.gateway ? '' : '/' + (options.divisionId || 'eng')),
    );
    return { page, calls, flags, membership, release: () => releaseMonth?.() };
  }
  const x = await mount(),
    p = x.page;
  const waitDesk = () => p.locator('.desk-intro h2').waitFor();
  await waitDesk();
  assert.equal(
    await p.locator('.desk-issue h3').first().innerText(),
    'تأخير المشاريع',
  );
  await p.getByText('المبدأ 6: السرعة', { exact: true }).waitFor();
  await p.screenshot({
    path: '/tmp/org-desk-issues-mobile.png',
    fullPage: true,
  });
  await p.getByRole('button', { name: '+ مشكلة جديدة', exact: true }).click();
  let dialog = p.getByRole('dialog');
  assert.equal(
    await dialog
      .getByLabel('المبدأ المكسور', { exact: true })
      .locator('option')
      .count(),
    21,
  );
  await dialog
    .getByLabel('عنوان المشكلة', { exact: true })
    .fill('مشكلة جديدة تجريبية');
  await dialog
    .getByLabel('شرح المشكلة', { exact: true })
    .fill('تفاصيل مصطنعة للمشكلة الجديدة');
  await dialog
    .getByLabel('مضارها', { exact: true })
    .fill('ضرر تجريبي على الإنجاز');
  await dialog.getByLabel('الوحدة', { exact: true }).selectOption('design');
  await dialog.getByLabel('النطاق', { exact: true }).selectOption('GENERAL');
  await dialog.getByLabel('المبدأ المكسور', { exact: true }).selectOption('20');
  await dialog.getByLabel('الخطورة', { exact: true }).selectOption('HIGH');
  await dialog.getByRole('button', { name: 'حفظ', exact: true }).click();
  await p
    .getByRole('heading', { name: 'مشكلة جديدة تجريبية', exact: true })
    .waitFor();
  const newIssue = x.calls.find((c) => c.path === '/api/org/issues');
  assert.equal(newIssue.body.principle, 20);
  assert.equal(newIssue.body.divisionId, 'eng');
  assert.equal(newIssue.auth, 'Bearer desk-test-token');
  const firstIssue = p.locator('.desk-issue').filter({
    has: p.getByRole('heading', { name: 'تأخير المشاريع', exact: true }),
  });
  await firstIssue
    .getByRole('button', { name: 'تغيير الحالة', exact: true })
    .click();
  dialog = p.getByRole('dialog');
  await dialog.getByLabel('الحالة', { exact: true }).selectOption('RESOLVED');
  await dialog.getByRole('button', { name: 'حفظ', exact: true }).click();
  assert(
    !x.calls.some(
      (c) => c.method === 'PATCH' && c.path === '/api/org/issues/high',
    ),
  );
  await dialog
    .getByLabel('شرح الحل', { exact: true })
    .fill('تم حل المشكلة باختبار مصطنع');
  await dialog.getByRole('button', { name: 'حفظ', exact: true }).click();
  await firstIssue.getByText('انحلت', { exact: true }).waitFor();
  await firstIssue
    .getByLabel('تعليق جديد', { exact: true })
    .fill('تعليق النائب على الحل');
  await firstIssue
    .getByRole('button', { name: 'إرسال التعليق', exact: true })
    .click();
  await firstIssue
    .getByText('تعليق النائب على الحل', { exact: true })
    .waitFor();
  await p.getByRole('button', { name: 'خطة التحسين', exact: true }).click();
  await p.getByRole('img', { name: 'إنجاز الخطة 40%' }).waitFor();
  const matrixCard = p.locator('.desk-plan-card').filter({
    has: p.getByRole('heading', { name: 'تقليل التأخير', exact: true }),
  });
  const manualCard = p.locator('.desk-plan-card').filter({
    has: p.getByRole('heading', { name: 'تطوير دليل العمل', exact: true }),
  });
  assert.equal(await matrixCard.getByRole('slider').count(), 0);
  await matrixCard.getByText('يحسبها ماتركس', { exact: true }).waitFor();
  await manualCard.getByRole('slider').fill('55');
  await manualCard
    .getByRole('button', { name: 'احفظ النسبة', exact: true })
    .click();
  await manualCard.getByRole('slider').waitFor();
  assert.equal(await manualCard.getByRole('slider').inputValue(), '55');
  await manualCard.getByRole('button', { name: 'خلص ✓', exact: true }).click();
  await manualCard.getByText('✓ خلص', { exact: false }).waitFor();
  await matrixCard
    .getByLabel('تعليق جديد', { exact: true })
    .fill('ملاحظة النائب على البند');
  await matrixCard
    .getByRole('button', { name: 'إرسال التعليق', exact: true })
    .click();
  await matrixCard
    .getByText('ملاحظة النائب على البند', { exact: true })
    .waitFor();
  await p.screenshot({ path: '/tmp/org-desk-plan-mobile.png', fullPage: true });
  await p.getByRole('button', { name: '+ إضافة بند', exact: true }).click();
  dialog = p.getByRole('dialog');
  await dialog
    .getByLabel('عنوان البند', { exact: true })
    .fill('بند محسوب جديد');
  await dialog.getByLabel('ربط بمشكلة', { exact: true }).selectOption('high');
  await dialog.getByLabel('المقياس', { exact: true }).selectOption('late');
  await dialog.getByLabel('الهدف', { exact: false }).fill('20');
  await dialog.getByText('أقل أحسن', { exact: false }).first().waitFor();
  await dialog.getByRole('button', { name: 'حفظ', exact: true }).click();
  await dialog
    .getByRole('alert')
    .getByText('الهدف لازم يكون أحسن من الرقم الحالي (12)', { exact: true })
    .waitFor();
  assert.equal(
    await dialog.getByLabel('الهدف', { exact: false }).inputValue(),
    '20',
  );
  await dialog.getByLabel('الهدف', { exact: false }).fill('4');
  await dialog.getByRole('button', { name: 'حفظ', exact: true }).click();
  await p
    .getByRole('heading', { name: 'بند محسوب جديد', exact: true })
    .waitFor();
  assert(
    !x.calls.some(
      (c) => c.path.endsWith('/matrix-item') && c.body?.manualProgress != null,
    ),
  );
  await p.getByRole('button', { name: 'الصلاحيات (DOA)', exact: true }).click();
  await p
    .getByRole('button', { name: 'تعديل جدول الصلاحيات', exact: true })
    .click();
  dialog = p.getByRole('dialog');
  await dialog.getByRole('button', { name: '+ إضافة صف', exact: true }).click();
  await dialog.getByLabel('القرار 2', { exact: true }).fill('صرف تجريبي');
  await dialog
    .getByLabel('طريقة الحسم 2', { exact: true })
    .selectOption('ESCALATE');
  await dialog.getByLabel('السقف بالدينار 2', { exact: true }).fill('٥٠');
  await dialog.getByRole('alert').waitFor();
  await dialog.getByRole('checkbox').check();
  await dialog.getByRole('button', { name: 'حفظ', exact: true }).click();
  assert(!x.calls.some((c) => c.method === 'PUT'));
  await dialog.getByLabel('السقف بالدينار 2', { exact: true }).fill('٥٠٠٠٠');
  assert.equal(
    await dialog.getByLabel('السقف بالدينار 2', { exact: true }).inputValue(),
    '50,000',
  );
  await dialog.getByRole('button', { name: 'شيل الصف 1', exact: true }).click();
  x.flags.fail = '/api/org/authority/eng';
  await dialog.getByRole('button', { name: 'حفظ', exact: true }).click();
  await dialog
    .getByRole('alert')
    .getByText('تعذر الحفظ التجريبي', { exact: true })
    .waitFor();
  assert.equal(
    await dialog.getByLabel('القرار 1', { exact: true }).inputValue(),
    'صرف تجريبي',
  );
  x.flags.fail = '';
  await dialog.getByRole('button', { name: 'حفظ', exact: true }).click();
  await p.getByRole('cell', { name: 'صرف تجريبي', exact: true }).waitFor();
  assert.deepEqual(x.calls.filter((c) => c.method === 'PUT').at(-1).body, [
    { decision: 'صرف تجريبي', mode: 'ESCALATE', limitIQD: 50000 },
  ]);
  await p.screenshot({
    path: '/tmp/org-desk-authority-mobile.png',
    fullPage: true,
  });
  await p.getByRole('button', { name: 'الفريق', exact: true }).click();
  await p.getByRole('heading', { name: 'التصميم', exact: true }).waitFor();
  await p.getByText(owner.name, { exact: true }).first().waitFor();
  await p.screenshot({ path: '/tmp/org-desk-team-mobile.png', fullPage: true });
  await p.getByRole('button', { name: 'خطة التحسين', exact: true }).click();
  await p.getByLabel('الشهر', { exact: true }).fill('2026-09');
  await p.getByRole('heading', { name: 'خطة 2026-09', exact: true }).waitFor();
  await p
    .getByText('الرقم الحالي مو متوفر لهالفترة؛ ما نعتبره صفر.', {
      exact: true,
    })
    .first()
    .waitFor();
  x.flags.delayMonth = '2026-07';
  await p.getByLabel('الشهر', { exact: true }).fill('2026-07');
  await p.waitForFunction(() =>
    document.querySelector('[role="status"]')?.textContent.includes('دا أجيب'),
  );
  await p.waitForTimeout(80);
  await p.getByLabel('الشهر', { exact: true }).fill('2026-08');
  await p.getByRole('heading', { name: 'خطة 2026-08', exact: true }).waitFor();
  x.release();
  await p.waitForTimeout(100);
  assert.equal(
    await p.getByRole('heading', { name: 'خطة 2026-07', exact: true }).count(),
    0,
  );
  await p.setViewportSize({ width: 1366, height: 1000 });
  await p.screenshot({
    path: '/tmp/org-desk-plan-desktop.png',
    fullPage: true,
  });
  await p.evaluate(() => document.documentElement.classList.add('dark'));
  await p.screenshot({
    path: '/tmp/org-desk-plan-desktop-dark.png',
    fullPage: true,
  });
  const reader = await mount(head, { unit: true }),
    r = reader.page;
  await r.locator('.desk-intro').waitFor();
  assert(!reader.calls.some((c) => c.path === '/api/org/chart'));
  assert.equal(
    await r.getByRole('button', { name: 'تغيير الحالة', exact: true }).count(),
    0,
  );
  assert(
    await r
      .getByRole('button', { name: '+ مشكلة جديدة', exact: true })
      .isVisible(),
  );
  await r.getByRole('button', { name: 'خطة التحسين', exact: true }).click();
  assert.equal(await r.getByRole('slider').count(), 0);
  assert.equal(
    await r.getByRole('button', { name: '+ إضافة بند', exact: true }).count(),
    0,
  );
  assert.equal(
    await r.getByRole('button', { name: 'خلص ✓', exact: true }).count(),
    0,
  );
  await r.getByRole('button', { name: 'الصلاحيات (DOA)', exact: true }).click();
  assert.equal(
    await r
      .getByRole('button', { name: 'تعديل جدول الصلاحيات', exact: true })
      .count(),
    0,
  );
  const starter = await mount(head, { noPlan: true }),
    s = starter.page;
  await s.locator('.desk-intro').waitFor();
  await s.getByRole('button', { name: 'خطة التحسين', exact: true }).click();
  await s.getByRole('button', { name: 'ابدي خطة هالشهر', exact: true }).click();
  await s
    .getByRole('dialog')
    .getByRole('button', { name: 'حفظ', exact: true })
    .click();
  await s.getByRole('button', { name: '+ إضافة بند', exact: true }).waitFor();
  assert(
    starter.calls.some(
      (c) =>
        c.path === '/api/org/plans' &&
        /^\d{4}-\d{2}$/.test(c.body.period) &&
        c.body.divisionId === 'eng',
    ),
  );
  const emptyReader = await mount(head, { unit: true, noPlan: true });
  await emptyReader.page.locator('.desk-intro').waitFor();
  await emptyReader.page
    .getByRole('button', { name: 'خطة التحسين', exact: true })
    .click();
  assert.equal(
    await emptyReader.page
      .getByRole('button', { name: 'ابدي خطة هالشهر', exact: true })
      .count(),
    0,
  );
  const directory = await mount(owner, { gateway: true });
  await directory.page
    .getByRole('link', { name: 'العمليات ←', exact: true })
    .waitFor();
  assert(!directory.calls.some((c) => c.path.startsWith('/api/org/desk/')));
  await directory.page
    .getByRole('link', { name: 'العمليات ←', exact: true })
    .click();
  await directory.page
    .locator('.desk-intro h2')
    .filter({ hasText: 'العمليات' })
    .waitFor();
  const root = await mount(head, { root: true, divisionId: 'ops' });
  await root.page
    .locator('.desk-intro h2')
    .filter({ hasText: 'العمليات' })
    .waitFor();
  assert.equal(
    await root.page
      .getByLabel('اختيار الإدارة', { exact: true })
      .locator('option')
      .count(),
    9,
  );
  assert(!root.calls.some((c) => c.path === '/api/org/chart'));
  for (const options of [
    { noHead: true },
    { member: true },
    { divisionId: 'ops' },
  ]) {
    const denied = await mount(head, options);
    await denied.page.getByRole('alert').waitFor();
    assert(denied.calls.every((c) => c.path === '/api/org/me'));
    assert.equal(
      await denied.page.evaluate(
        ({ actor, items }) => window.__deskVisible(actor, items),
        {
          actor: head,
          items: options.noHead
            ? []
            : options.member
              ? [{ position: 'MEMBER', divisionId: 'eng', unitId: null }]
              : denied.membership,
        },
      ),
      !options.noHead && !options.member,
    );
  }
  const hidden = await mount(owner);
  await hidden.page.locator('.desk-intro').waitFor();
  hidden.flags.hideDesk = true;
  await hidden.page.getByLabel('الشهر', { exact: true }).fill('2025-01');
  await hidden.page
    .getByRole('alert')
    .filter({ hasText: 'مو موجود' })
    .waitFor();
  assert.equal(
    await hidden.page
      .getByRole('button', { name: '+ مشكلة جديدة', exact: true })
      .count(),
    0,
  );
  for (const page of [p, r, s])
    for (const width of [390, 1366]) {
      await page.setViewportSize({ width, height: 900 });
      assert(
        await page.evaluate(
          () => document.documentElement.scrollWidth <= innerWidth,
        ),
      );
      assert(
        await page
          .locator('.org-ui input,.org-ui textarea,.org-ui select')
          .evaluateAll((rows) =>
            rows.every((el) => parseFloat(getComputedStyle(el).fontSize) >= 16),
          ),
      );
    }
  assert.deepEqual(errors, []);
  await browser.close();
  console.log(
    'PASS: head-only gate/nav and zero desk/chart calls for unassigned/member/wrong scope; root management all desks without chart; owner directory; severity/principles; issue/status/required solution/comments; computed/manual plan progress, done, metric target error preserves draft, create plan; DOA owner-only full replacement with small IQD and retry; history/null current/month race/404; mobile/desktop/dark. Synthetic APIs only.',
  );
})().catch((e) => {
  console.error(e);
  process.exit(1);
});
