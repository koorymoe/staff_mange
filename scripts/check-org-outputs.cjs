// Synthetic HTTP fixtures only: never connects to the production database.
const assert = require('node:assert/strict');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
(async () => {
  const browser = await chromium.launch({
    executablePath: '/usr/bin/chromium',
    args: ['--no-sandbox'],
  });
  const base = 'http://127.0.0.1:5199/staff_mange';
  const transformed = await (
    await fetch(base + '/src/pages/OrgDeskPage.tsx')
  ).text();
  const routerPath = transformed.match(
    /["']([^"']*react-router-dom\.js[^"']*)["']/,
  )[1];
  const owner = {
    id: 'owner-test',
    name: 'مالك تجريبي',
    role: 'ADMIN',
    actualRole: 'OWNER',
    status: 'ACTIVE',
  };
  const member = {
    id: 'self-test',
    name: 'موظف تجريبي',
    role: 'TECHNICIAN',
    actualRole: 'TECHNICIAN',
    status: 'ACTIVE',
  };
  const clauses = Array.from({ length: 20 }, (_, i) => ({
    no: i + 1,
    article: i < 5 ? 2 : 3,
    articleTitle: i < 5 ? 'الجودة والأداء' : 'المسؤولية',
    title: 'فقرة تجريبية ' + (i + 1),
    code: `م${i < 5 ? 2 : 3}/ف${i + 1}`,
  }));
  const errors = [];
  async function mount(screen, actor = owner, extra = {}) {
    const page = await browser.newPage({
      viewport: { width: 390, height: 844 },
    });
    page.on('pageerror', (e) => {
      errors.push(e.message);
      console.error('PAGE ERROR', e.message);
    });
    const calls = [],
      flags = {
        failOutput: false,
        failBinding: false,
        failReg: false,
        delay: '',
      };
    let clause = null;
    const task = () => ({
      id: 'task-1',
      title: 'مهمة اختبار',
      assignedToId: member.id,
      assignedToName: member.name,
      assignedByName: owner.name,
      priority: 'NORMAL',
      status: 'NEW',
      overdue: false,
      createdAt: '2026-10-30T06:00:00Z',
      clause,
    });
    const division = {
      id: 'eng',
      name: 'الهندسية',
      summary: 'اختبار مصطنع',
      people: [
        {
          id: 'assignment-1',
          employeeId: member.id,
          employeeName: member.name,
          divisionId: 'eng',
          unitId: null,
          position: 'DIRECTOR',
          allowanceIQD: 100000,
          startedAt: '2026-10-01',
          acting: false,
        },
      ],
      permissions: [],
      units: [
        {
          id: 'design',
          divisionId: 'eng',
          name: 'التصميم',
          people: [],
          permissions: [],
        },
        {
          id: 'other',
          divisionId: 'eng',
          name: 'وحدة ثانية',
          people: [],
          permissions: [],
        },
      ],
    };
    await page.route('**/api/**', async (r) => {
      const req = r.request(),
        u = new URL(req.url()),
        path = u.pathname.replace(/^.*\/api/, ''),
        body = req.postDataJSON();
      calls.push({
        path,
        method: req.method(),
        query: Object.fromEntries(u.searchParams),
        body,
      });
      const ok = (data) =>
        r.fulfill({
          contentType: 'application/json',
          body: JSON.stringify(data),
        });
      const fail = (message) =>
        r.fulfill({
          status: 404,
          contentType: 'application/json',
          body: JSON.stringify({ error: message }),
        });
      if (path === '/org/regulations')
        return flags.failReg ? fail('اللائحة غير متوفرة') : ok(clauses);
      if (path === '/org/outputs') {
        const period = u.searchParams.get('period');
        if (flags.delay === period)
          await new Promise((resolve) => setTimeout(resolve, 450));
        if (flags.failOutput) return fail('مو موجود');
        return ok({
          label: 'مخرجات اختبار ' + period,
          period,
          people: 1,
          weakest: ['م2/ف3: تحتاج متابعة'],
          rows: clauses.map((c, i) => ({
            code: c.code,
            title: c.title,
            score: [80, 50, 49, 0, null][i % 5],
            matrixPct: i === 4 ? null : 75,
            tasksOnTime: 3,
            tasksLate: 2,
            tasksOverdue: 1,
            kpiDeductions: 2,
            complaints: 1,
          })),
        });
      }
      if (path === '/org/me')
        return ok(
          screen === 'desk'
            ? [
                {
                  id: 'head',
                  employeeId: actor.id,
                  divisionId: 'eng',
                  unitId: 'design',
                  position: 'DIRECTOR',
                },
              ]
            : [],
        );
      if (path === '/org/my-items' || path === '/org/directives/mine')
        return ok([]);
      if (path === '/org/chart') return ok([division]);
      if (path === '/org/desk/eng')
        return ok({
          division,
          period: '2026-10',
          canEdit: false,
          editableUnits: ['design'],
          canReview: false,
          issues: [],
          plan: null,
          unitPlans: [],
          comments: [],
          authority: [],
          metrics: [],
        });
      if (path === '/employees') return ok([member]);
      if (path === '/permissions/catalog') return ok([]);
      if (path === '/extra-tasks') return ok([task()]);
      if (path === '/kpi') return ok([]);
      if (path === '/kpi-criteria' && req.method() === 'GET')
        return ok([
          {
            id: 'criterion-1',
            label: 'معيار اختبار',
            createdAt: '2026-10-30',
            clause,
          },
        ]);
      if (/^\/org\/(tasks|kpi-criteria)\/.+\/clause$/.test(path)) {
        if (flags.failBinding) return fail('تعذّر حفظ الفقرة');
        clause = body.clause;
        return ok({ ok: true });
      }
      if (path === '/extra-tasks/task-1/cancel')
        return ok({ ...task(), status: 'CANCELLED' });
      if (path === '/kpi-criteria/criterion-1' && req.method() === 'DELETE')
        return ok({ ok: true });
      errors.push('Unexpected API ' + req.method() + ' ' + path);
      return fail('Unexpected API');
    });
    await page.route(base + '/outputs-preview', (r) =>
      r.fulfill({
        contentType: 'text/html',
        body: `<html dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body><div id="root" style="padding:12px"></div><script type="module">
      import RR from '${base}/@react-refresh';RR.injectIntoGlobalHook(window);window.$RefreshReg$=()=>{};window.$RefreshSig$=()=>x=>x;window.__vite_plugin_react_preamble_installed__=true;
      window.prompt=window.confirm=()=>{throw Error('Native dialog forbidden')};localStorage.setItem('authToken','synthetic-only');
      const R=await import('${base}/node_modules/.vite/deps/react.js'), React=R.default||R;
      const D=await import('${base}/node_modules/.vite/deps/react-dom_client.js'),{createRoot}=D.default||D;
      const router=await import('${routerPath}');const {SessionContext}=await import('${base}/src/session.ts');await import('${base}/src/index.css');await import('${base}/src/fonts.css');
      const screen=${JSON.stringify(screen)};const file=screen==='self'?'components/MyOrgCard':screen==='desk'?'pages/OrgDeskPage':screen==='chart'?'pages/OrgPage':screen==='tasks'?'pages/ExtraTasksPage':screen==='kpi'?'pages/KpiPage':'components/OrgOutputs';
      const module=await import('${base}/src/'+file+'.tsx'), Page=screen==='kpi'?module.AdministrativeTab:module.default;
      const view=React.createElement(Page,${JSON.stringify(extra)});
      const content=screen==='desk'?React.createElement(router.Routes,null,React.createElement(router.Route,{path:'outputs-preview',element:React.createElement(router.Navigate,{to:'/org/desk/eng',replace:true})}),React.createElement(router.Route,{path:'org/desk/:divisionId',element:view})):view;
      createRoot(document.getElementById('root')).render(React.createElement(router.BrowserRouter,{basename:'/staff_mange'},React.createElement(SessionContext.Provider,{value:{employee:${JSON.stringify(actor)},permissions:['kpi_criteria_management'],setEmployee:()=>{},gpsServiceId:null}},content)));
      </script></body></html>`,
      }),
    );
    await page.goto(base + '/outputs-preview');
    return { page, calls, flags };
  }
  async function noOverflow(page) {
    assert(
      await page.evaluate(
        () => document.documentElement.scrollWidth <= innerWidth + 1,
      ),
    );
  }
  async function contrast(page) {
    const ratios = await page
      .locator('.outputs-score, .outputs-details dt')
      .evaluateAll((els) => {
        const rgb = (s) =>
          s
            .match(/[\d.]+/g)
            .slice(0, 3)
            .map(Number);
        const lum = (s) =>
          rgb(s)
            .map((x) => {
              x /= 255;
              return x <= 0.04045 ? x / 12.92 : ((x + 0.055) / 1.055) ** 2.4;
            })
            .reduce((n, x, i) => n + x * [0.2126, 0.7152, 0.0722][i], 0);
        return els.map((el) => {
          let p = el,
            bg;
          while (p) {
            bg = getComputedStyle(p).backgroundColor;
            if (bg !== 'rgba(0, 0, 0, 0)') break;
            p = p.parentElement;
          }
          const a = lum(getComputedStyle(el).color),
            b = lum(bg || 'rgb(255,255,255)');
          return (Math.max(a, b) + 0.05) / (Math.min(a, b) + 0.05);
        });
      });
    assert(
      ratios.every((r) => r >= 4.5),
      JSON.stringify(ratios),
    );
  }
  const self = await mount('self', member),
    p = self.page;
  await p.getByRole('button', { name: 'افتح مخرجاتي' }).click();
  await p.locator('.outputs-row').last().waitFor();
  assert.equal(await p.locator('.outputs-row').count(), 20);
  assert.equal(await p.locator('.output-good').count(), 4);
  assert.equal(await p.locator('.output-medium').count(), 4);
  assert.equal(await p.locator('.output-low').count(), 8);
  assert.equal(await p.locator('.output-empty progress').count(), 0);
  assert.equal(await p.locator('.outputs-article').count(), 2);
  assert(
    self.calls
      .filter((c) => c.path === '/org/outputs')
      .every(
        (c) =>
          c.query.employeeId === member.id &&
          !c.query.unitId &&
          !c.query.divisionId,
      ),
  );
  assert(
    !self.calls.some(
      (c) => c.path === '/org/chart' || c.path.startsWith('/org/desk/'),
    ),
  );
  await noOverflow(p);
  await contrast(p);
  assert(
    await p
      .getByLabel('شهر المخرجات')
      .evaluate((el) => parseFloat(getComputedStyle(el).fontSize) >= 16),
  );
  await p.screenshot({ path: '/tmp/outputs-self-mobile.png', fullPage: true });
  self.flags.delay = '2026-09';
  await p.getByLabel('شهر المخرجات').fill('2026-09');
  await p.waitForTimeout(70);
  await p.getByLabel('شهر المخرجات').fill('2026-08');
  await p.getByRole('heading', { name: 'مخرجات اختبار 2026-08' }).waitFor();
  await p.waitForTimeout(500);
  assert.equal(
    await p.getByRole('heading', { name: 'مخرجات اختبار 2026-09' }).count(),
    0,
  );
  self.flags.failOutput = true;
  await p.getByLabel('شهر المخرجات').fill('2026-07');
  await p.getByRole('button', { name: 'أعد جلب المخرجات' }).waitFor();
  assert.equal(await p.locator('.outputs-row').count(), 0);
  self.flags.failOutput = false;
  self.flags.failReg = true;
  await p.getByRole('button', { name: 'أعد جلب المخرجات' }).click();
  await p.locator('.outputs-row').last().waitFor();
  await p.getByLabel('شهر المخرجات').fill('2026-06');
  await p.getByRole('button', { name: 'أعد جلب لائحة المخرجات' }).waitFor();
  assert.equal(
    await p
      .getByRole('heading', { name: 'فقرات بانتظار تصنيف المادة' })
      .count(),
    1,
  );
  self.flags.failReg = false;
  await p.getByRole('button', { name: 'أعد جلب لائحة المخرجات' }).click();
  await p.getByRole('heading', { name: 'المادة 2: الجودة والأداء' }).waitFor();
  await p.close();
  const denied = await mount('denied', member, {
    target: { kind: 'employee', id: 'another-person' },
  });
  await denied.page.getByRole('alert').waitFor();
  assert.equal(denied.calls.length, 0);
  await denied.page.close();
  const desk = await mount('desk', member);
  await desk.page
    .getByRole('button', { name: 'المخرجات', exact: true })
    .click();
  await desk.page.locator('.outputs-row').last().waitFor();
  await desk.page
    .getByRole('button', { name: 'مخرجات الوحدة', exact: true })
    .click();
  await desk.page.locator('.outputs-row').last().waitFor();
  assert.equal(
    await desk.page.getByLabel('وحدة المخرجات').locator('option').count(),
    1,
  );
  assert(
    desk.calls.some(
      (c) => c.path === '/org/outputs' && c.query.divisionId === 'eng',
    ),
  );
  assert(
    desk.calls.some(
      (c) => c.path === '/org/outputs' && c.query.unitId === 'design',
    ),
  );
  assert(!desk.calls.some((c) => c.query.unitId === 'other'));
  await desk.page.setViewportSize({ width: 1366, height: 900 });
  await desk.page.evaluate(() =>
    document.documentElement.classList.add('dark'),
  );
  await noOverflow(desk.page);
  await contrast(desk.page);
  await desk.page.screenshot({
    path: '/tmp/outputs-unit-desktop-dark.png',
    fullPage: true,
  });
  await desk.page.close();
  const chart = await mount('chart');
  await chart.page.locator('summary').filter({ hasText: 'الهندسية' }).click();
  await chart.page
    .getByRole('button')
    .filter({ hasText: member.name })
    .first()
    .click();
  await chart.page
    .getByRole('button', { name: '🎯 المخرجات', exact: true })
    .click();
  await chart.page.locator('.outputs-row').last().waitFor();
  assert(
    chart.calls.some(
      (c) => c.path === '/org/outputs' && c.query.employeeId === member.id,
    ),
  );
  await noOverflow(chart.page);
  await chart.page.screenshot({ path: '/tmp/outputs-person-mobile.png' });
  await chart.page.close();
  for (const screen of ['tasks', 'kpi']) {
    const t = await mount(screen),
      pg = t.page;
    const select = pg.getByLabel('فقرة المخرجات', { exact: true });
    try {
      await select.waitFor({ timeout: 8000 });
    } catch (e) {
      console.log(screen, t.calls, await pg.locator('body').innerText());
      throw e;
    }
    await select.selectOption('1');
    assert.equal(t.calls.filter((c) => c.method === 'PUT').length, 0);
    t.flags.failBinding = true;
    await pg.getByRole('button', { name: 'حفظ الفقرة' }).click();
    await pg
      .getByRole('alert')
      .filter({ hasText: 'تعذّر حفظ الفقرة' })
      .waitFor();
    assert.equal(await select.inputValue(), '1');
    t.flags.failBinding = false;
    await pg.getByRole('button', { name: 'حفظ الفقرة' }).click();
    await pg.getByRole('status').filter({ hasText: 'انحفظ' }).first().waitFor();
    await pg.waitForTimeout(150);
    assert(t.calls.some((c) => c.method === 'PUT' && c.body.clause === 1));
    await select.selectOption('');
    await pg.getByRole('button', { name: 'حفظ الفقرة' }).click();
    await pg.waitForTimeout(150);
    assert(t.calls.some((c) => c.method === 'PUT' && c.body.clause === null));
    assert.equal(
      t.calls.filter((c) => c.path === '/org/regulations').length,
      1,
    );
    await noOverflow(pg);
    await pg.screenshot({
      path: `/tmp/outputs-${screen}-binding-mobile.png`,
      fullPage: true,
    });
    if (screen === 'tasks') {
      await pg.getByRole('button', { name: 'إلغاء', exact: true }).click();
      const dialog = pg.getByRole('dialog');
      await dialog.waitFor();
      assert(
        await dialog
          .getByRole('button', { name: 'تأكيد', exact: true })
          .isDisabled(),
      );
      await dialog.getByRole('button', { name: 'تراجع' }).click();
      assert(!t.calls.some((c) => c.path.endsWith('/cancel')));
      await pg.getByRole('button', { name: 'إلغاء', exact: true }).click();
      await dialog
        .getByLabel('السبب', { exact: true })
        .fill('سبب اختبار مصطنع');
      await dialog.getByRole('button', { name: 'تأكيد', exact: true }).click();
      await pg.waitForTimeout(150);
      assert(
        t.calls.some(
          (c) =>
            c.path.endsWith('/cancel') && c.body.reason === 'سبب اختبار مصطنع',
        ),
      );
    } else {
      await pg.getByTitle('حذف', { exact: true }).click();
      await pg.getByRole('dialog').waitFor();
      await pg.keyboard.press('Escape');
      assert(!t.calls.some((c) => c.method === 'DELETE'));
      await pg.getByTitle('حذف', { exact: true }).click();
      await pg
        .getByRole('dialog')
        .getByRole('button', { name: 'تأكيد', exact: true })
        .click();
      await pg.waitForTimeout(150);
      assert(t.calls.some((c) => c.method === 'DELETE'));
    }
    await pg.close();
    const restricted = await mount(screen, member);
    await restricted.page
      .getByText(screen === 'tasks' ? 'مهمة اختبار' : 'معيار اختبار', {
        exact: true,
      })
      .first()
      .waitFor();
    assert.equal(
      await restricted.page
        .getByLabel('فقرة المخرجات', { exact: true })
        .count(),
      0,
    );
    assert(
      !restricted.calls.some(
        (c) => c.path === '/org/regulations' || c.method === 'PUT',
      ),
    );
    await restricted.page.close();
  }
  assert.deepEqual(errors, []);
  await browser.close();
  console.log(
    'PASS: 20 dynamic clauses/article grouping, null vs zero, score boundaries, month race/error/retry, safe self/head scopes, admin task/KPI binding/unlink/error, nonadmin zero privileged calls, inline cancellation confirmations, 390/1366/dark/overflow. Synthetic APIs only.',
  );
})().catch((e) => {
  console.error(e);
  process.exit(1);
});
