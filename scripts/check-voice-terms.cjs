const assert = require('node:assert/strict');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
(async () => {
  const browser = await chromium.launch({
    executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium',
    args: ['--no-sandbox'],
  });
  const base =
    process.env.VOICE_BASE_URL || 'http://127.0.0.1:5199/staff_mange';
  const errors = [];
  const owner = {
    id: 'voice-owner-test',
    name: 'مدير تجريبي',
    role: 'ADMIN',
    actualRole: 'OWNER',
  };
  async function mount(actor = owner) {
    const page = await browser.newPage({
      viewport: { width: 390, height: 844 },
    });
    page.on('pageerror', (e) => errors.push(e.message));
    const calls = [];
    let terms = [
      {
        term: 'داش كام',
        meaning: 'كاميرا السيارة',
        category: 'خدمة',
        source: 'SYSTEM',
      },
      { term: 'سويچ', meaning: 'موزع شبكة', category: 'شبكات', source: 'SEED' },
      {
        term: 'علّمتها الأولى',
        meaning: 'معلومة اختبار',
        category: 'تعلم',
        source: 'OWNER',
      },
      {
        term: 'المتعلمة الثانية',
        meaning: 'معلومة ثانية',
        category: 'تعلم',
        source: 'LEARNED',
      },
    ];
    let fixes = [
      { wrong: 'دش كاب', right: 'داش كام', source: 'OWNER' },
      { wrong: 'سويش', right: 'سويچ', source: 'SEED' },
    ];
    const flags = {
      failGet: false,
      failPost: false,
      failDelete: false,
      slowSystem: false,
    };
    await page.addInitScript(() => {
      localStorage.setItem('authToken', 'voice-test-token');
      window.prompt = window.confirm = () => {
        throw new Error('Native confirmation forbidden');
      };
    });
    await page.route(base + '/voice-preview', (r) =>
      r.fulfill({
        contentType: 'text/html',
        body: `<html dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body><div id="root"></div><script type="module">import RR from '${base}/@react-refresh';RR.injectIntoGlobalHook(window);window.$RefreshReg$=()=>{};window.$RefreshSig$=()=>x=>x;window.__vite_plugin_react_preamble_installed__=true;const R=await import('${base}/node_modules/.vite/deps/react.js');const React=R.default||R;const D=await import('${base}/node_modules/.vite/deps/react-dom_client.js');const {createRoot}=D.default||D;const {SessionContext}=await import('${base}/src/session.ts');await import('${base}/src/index.css');await import('${base}/src/fonts.css');const {default:Page}=await import('${base}/src/pages/VoiceTermsPage.tsx');const nav=await import('${base}/src/components/navTree.tsx');window.__navAllowed=(actor)=>{const ctx={employee:actor,permissions:['monitoring','auditing','unit_management','agent_matrix'],gpsServiceId:null};return ['/voice-terms','/ai-suggestions'].map(path=>nav.isNavVisible(nav.navItems.find(x=>x.to===path),ctx,true));};window.__root=createRoot(document.getElementById('root'));window.__root.render(React.createElement(SessionContext.Provider,{value:{employee:${JSON.stringify(actor)},permissions:[],setEmployee:()=>{},gpsServiceId:null}},React.createElement(Page)));</script></body></html>`,
      }),
    );
    await page.route('**/api/**', async (r) => {
      const req = r.request(),
        url = new URL(req.url()),
        method = req.method();
      calls.push({
        method,
        path: url.pathname,
        q: url.searchParams.get('q'),
        source: url.searchParams.get('source'),
        term: url.searchParams.get('term'),
        wrong: url.searchParams.get('wrong'),
        body: req.postDataJSON(),
        auth: req.headers().authorization,
      });
      assert.equal(
        url.pathname,
        '/api/voice/terms',
        'No speculative endpoints',
      );
      if (
        (method === 'GET' && flags.failGet) ||
        (method === 'POST' && flags.failPost) ||
        (method === 'DELETE' && flags.failDelete)
      )
        return r.fulfill({
          status: 500,
          contentType: 'application/json',
          body: JSON.stringify({ error: 'فشل تجريبي' }),
        });
      if (method === 'POST') {
        const body = req.postDataJSON();
        terms.unshift({
          term: body.term,
          meaning: body.meaning,
          category: 'تعلم',
          source: 'OWNER',
        });
        if (body.heardAs)
          fixes.unshift({
            wrong: body.heardAs,
            right: body.term,
            source: 'OWNER',
          });
      } else if (method === 'DELETE') {
        if (url.searchParams.has('term'))
          terms = terms.filter((x) => x.term !== url.searchParams.get('term'));
        if (url.searchParams.has('wrong'))
          fixes = fixes.filter(
            (x) => x.wrong !== url.searchParams.get('wrong'),
          );
      }
      const q = url.searchParams.get('q') || '',
        source = url.searchParams.get('source');
      const response =
        method === 'GET'
          ? {
              total: terms.length,
              terms: terms.filter(
                (x) =>
                  (!source || x.source === source) &&
                  (x.term + ' ' + x.meaning).includes(q),
              ),
              fixes,
            }
          : { ok: true };
      if (source === 'SYSTEM' && flags.slowSystem)
        await new Promise((resolve) => setTimeout(resolve, 700));
      await r.fulfill({
        contentType: 'application/json',
        body: JSON.stringify(response),
      });
    });
    await page.goto(base + '/voice-preview');
    return { page, calls, flags };
  }
  const x = await mount(),
    p = x.page;
  const words = p.getByRole('list', { name: 'كلمات القاموس' });
  await words.waitFor();
  assert.equal(await words.locator('li').count(), 4);
  assert.equal(await p.locator('.ear-count strong').innerText(), '٤');
  assert.deepEqual(
    await p.evaluate((actor) => window.__navAllowed(actor), owner),
    [true, true],
  );
  await p.screenshot({ path: '/tmp/voice-terms-mobile.png', fullPage: true });
  // Real debounce: no request for each keystroke, then one final query.
  const search = p.getByLabel('دوّر بالقاموس');
  const before = x.calls.length;
  await search.fill('كام');
  await search.fill('كاميرا');
  await p.waitForTimeout(120);
  assert.equal(x.calls.length, before);
  await words.getByRole('heading', { name: 'داش كام' }).waitFor();
  assert.equal(await words.locator('li').count(), 1);
  assert.equal(x.calls.at(-1).q, 'كاميرا');
  // Whitespace and quickly returning to the loaded query must not stick loading.
  await search.fill('كاميرا ');
  await words.waitFor();
  await search.fill('غير');
  await search.fill('كاميرا ');
  await words.waitFor();
  await search.fill('');
  await p
    .getByRole('group', { name: 'مصدر الكلمات' })
    .getByRole('button', { name: 'علّمتها انت' })
    .click();
  await words.getByRole('heading', { name: 'المتعلمة الثانية' }).waitFor();
  assert.equal(await words.locator('li').count(), 2);
  assert(
    x.calls.some((c) => c.source === 'OWNER') &&
      x.calls.some((c) => c.source === 'LEARNED'),
  );
  x.flags.slowSystem = true;
  await p.getByRole('button', { name: 'النظام', exact: true }).click();
  await p.waitForTimeout(80);
  await p.getByRole('button', { name: 'الورقة', exact: true }).click();
  await words.getByRole('heading', { name: 'سويچ' }).waitFor();
  await p.waitForTimeout(800);
  assert.equal(
    await words.locator('li').count(),
    1,
    'Late response cannot replace selected source',
  );
  assert.equal(await words.getByRole('heading').innerText(), 'سويچ');
  await p.getByRole('button', { name: 'الكل', exact: true }).click();
  await words.getByRole('heading', { name: 'داش كام' }).waitFor();
  // Enter saves JSON and refreshes without navigation.
  await p.getByLabel('الكلمة الصح', { exact: true }).fill('صنف & جديد؟');
  await p.getByLabel('انكتبت غلط هيچ (اختياري)').fill('صنف / غلط؟');
  await p.getByLabel('معناها', { exact: true }).fill('معنى تجريبي');
  await p.getByLabel('معناها', { exact: true }).press('Enter');
  await words.getByRole('heading', { name: 'صنف & جديد؟' }).waitFor();
  const post = x.calls.find((c) => c.method === 'POST');
  assert.deepEqual(post.body, {
    term: 'صنف & جديد؟',
    heardAs: 'صنف / غلط؟',
    meaning: 'معنى تجريبي',
  });
  assert.equal(post.auth, 'Bearer voice-test-token');
  assert.equal(await p.locator('.ear-count strong').innerText(), '٥');
  // Delete is never issued until inline confirmation. Cancel restores focus.
  const del = p.getByRole('button', {
    name: 'حذف الكلمة صنف & جديد؟',
    exact: true,
  });
  await del.click();
  assert(!x.calls.some((c) => c.method === 'DELETE'));
  await p.getByRole('button', { name: 'لا، خلّيها' }).click();
  await del.click();
  x.flags.failDelete = true;
  await p.getByRole('button', { name: 'إي، احذف' }).click();
  await p.getByRole('alert').getByText('فشل تجريبي').waitFor();
  assert(await words.getByRole('heading', { name: 'صنف & جديد؟' }).isVisible());
  x.flags.failDelete = false;
  await p.getByRole('button', { name: 'إي، احذف' }).click();
  await p.getByRole('status').filter({ hasText: 'انشالت الكلمة' }).waitFor();
  assert(
    x.calls.some(
      (c) =>
        c.method === 'DELETE' && c.term === 'صنف & جديد؟' && c.wrong === null,
    ),
  );
  await p.getByRole('button', { name: 'التصحيحات', exact: true }).click();
  const corrections = p.getByRole('list', { name: 'تصحيحات الأذن' });
  await corrections.getByText('صنف / غلط؟', { exact: true }).waitFor();
  await p.screenshot({ path: '/tmp/voice-fixes-mobile.png', fullPage: true });
  await search.fill('صنف');
  await p.waitForTimeout(450);
  assert.equal(
    await corrections.locator('li').count(),
    1,
    'Fixes filtered locally despite API returning all',
  );
  await corrections
    .getByRole('button', { name: 'حذف التصحيح صنف / غلط؟', exact: true })
    .click();
  await p.getByRole('button', { name: 'إي، احذف' }).click();
  await p.getByRole('heading', { name: 'ما لگينا شي بهالبحث' }).waitFor();
  assert(
    x.calls.some(
      (c) =>
        c.method === 'DELETE' && c.wrong === 'صنف / غلط؟' && c.term === null,
    ),
  );
  await search.fill('');
  await corrections.waitFor();
  await p.getByRole('button', { name: 'الورقة', exact: true }).click();
  await corrections.getByText('سويش', { exact: true }).waitFor();
  assert.equal(await corrections.locator('li').count(), 1);
  // Failed saves preserve the editable draft; retry is explicit.
  x.flags.failPost = true;
  await p.getByLabel('الكلمة الصح', { exact: true }).fill('مسودة باقية');
  await p.getByLabel('الكلمة الصح', { exact: true }).press('Enter');
  await p.getByRole('alert').waitFor();
  assert.equal(
    await p.getByLabel('الكلمة الصح', { exact: true }).inputValue(),
    'مسودة باقية',
  );
  x.flags.failPost = false;
  x.flags.failGet = true;
  await p.getByRole('button', { name: 'النظام', exact: true }).click();
  await p.getByRole('button', { name: 'أعد المحاولة' }).waitFor();
  x.flags.failGet = false;
  x.flags.slowSystem = false;
  await p.getByRole('button', { name: 'أعد المحاولة' }).click();
  await p.getByRole('heading', { name: 'ما لگينا شي بهالبحث' }).waitFor();
  await p.getByRole('button', { name: 'الكلمات', exact: true }).click();
  await words.getByRole('heading', { name: 'داش كام' }).waitFor();
  for (const dark of [false, true]) {
    await p.evaluate(
      (d) => document.documentElement.classList.toggle('dark', d),
      dark,
    );
    for (const width of [390, 1366]) {
      await p.setViewportSize({ width, height: 900 });
      assert(
        await p.evaluate(
          () => document.documentElement.scrollWidth <= innerWidth,
        ),
        'No horizontal overflow',
      );
      assert(
        await p
          .getByLabel('الكلمة الصح', { exact: true })
          .evaluate((el) => parseFloat(getComputedStyle(el).fontSize) >= 16),
      );
    }
  }
  await p.getByRole('button', { name: 'الكل', exact: true }).click();
  await words.getByRole('heading', { name: 'سويچ' }).waitFor();
  await p.screenshot({
    path: '/tmp/voice-terms-desktop-dark.png',
    fullPage: true,
  });
  // Conservative contrast: test every foreground at both gradient endpoints.
  await p.evaluate(() => {
    const lum = (hex) =>
      hex
        .match(/\w\w/g)
        .map((v) => parseInt(v, 16) / 255)
        .map((v) => (v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4))
        .reduce((s, v, i) => s + v * [0.2126, 0.7152, 0.0722][i], 0);
    const ratio = (a, b) =>
      (Math.max(lum(a), lum(b)) + 0.05) / (Math.min(lum(a), lum(b)) + 0.05);
    for (const [fg, bg] of [
      ['f4f6fc', '191b22'],
      ['bac4d6', '191b22'],
      ['f6d389', '382c19'],
      ['90ddff', '163244'],
      ['d1d7e2', '303541'],
      ['ffacb9', '191b22'],
      ['8de3b9', '191b22'],
      ['f4f6fc', '075d7c'],
      ['f4f6fc', '92365f'],
    ])
      if (ratio(fg, bg) < 4.5)
        throw new Error('Contrast below 4.5:1: ' + fg + '/' + bg);
  });
  const admin = await mount({ ...owner, actualRole: 'ADMIN' });
  await admin.page.getByRole('list', { name: 'كلمات القاموس' }).waitFor();
  assert(admin.calls.some((c) => c.method === 'GET'));
  for (const role of ['TECHNICIAN', 'MONITOR', 'QUALITY_ENGINEER', 'FINANCE']) {
    const actor = { ...owner, role, actualRole: role };
    const denied = await mount(actor);
    await denied.page.getByRole('alert').waitFor();
    await denied.page.waitForTimeout(400);
    assert.equal(denied.calls.length, 0, 'Denied roles must never probe API');
    assert.deepEqual(
      await denied.page.evaluate((a) => window.__navAllowed(a), actor),
      [false, false],
    );
    await denied.page.close();
  }
  assert.deepEqual(errors, []);
  await browser.close();
  console.log(
    'PASS: owner/admin; denied zero API and hidden nav even with grants; debounce/whitespace; search and filter races; OWNER+LEARNED; Enter save; independent term/fix confirmed deletion with encoded queries; recoverable GET/POST/DELETE errors; 390px/desktop/light/dark; >=4.5 contrast palette. Synthetic APIs only.',
  );
})().catch((e) => {
  console.error(e);
  process.exit(1);
});
