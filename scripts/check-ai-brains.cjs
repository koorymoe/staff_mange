const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const assert = require('node:assert/strict');
(async () => {
  const browser = await chromium.launch({
    executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium',
    args: ['--no-sandbox'],
  });
  const base =
      process.env.BRAINS_BASE_URL || 'http://127.0.0.1:5199/staff_mange',
    errors = [];
  const transformed = await (
    await fetch(base + '/src/pages/AiAgentsPage.tsx')
  ).text();
  const routerPath = transformed.match(
    /["']([^"']*react-router-dom\.js[^"']*)["']/,
  )[1];
  const owner = {
    id: 'owner-test',
    name: 'مدير تجريبي',
    role: 'ADMIN',
    actualRole: 'OWNER',
  };
  const now = '2026-10-28T08:00:00Z';
  async function mount(screen, actor = owner) {
    const page = await browser.newPage({
      viewport: { width: 390, height: 844 },
    });
    page.on('pageerror', (e) => {
      errors.push(e.message);
      console.log('PAGE ERROR', e.message);
    });
    const calls = [];
    let finishCreate,
      delayCreate = true,
      delayFilter = false,
      finishFilter,
      denyDecision = false;
    let brains = {
      internal: {
        model: 'work-model',
        chatModel: 'chat-model',
        spentTodayUSD: 1,
        dailyLimitUSD: 2,
      },
      external: { model: 'external-model', spentTodayUSD: 3, dailyLimitUSD: 2 },
    };
    let items = [
      {
        id: 's1',
        agent: 'IRIS',
        title: 'اقتراح تجريبي',
        details: 'تفاصيل تجريبية فقط',
        needsCode: true,
        status: 'NEW',
        createdAt: now,
      },
      {
        id: 's2',
        agent: 'MATRIX',
        title: 'اقتراح ثانٍ',
        details: 'تفاصيل ثانية',
        needsCode: false,
        status: 'ACCEPTED',
        createdAt: now,
      },
    ];
    await page.addInitScript(() => {
      localStorage.setItem('authToken', 'mock-token');
      const original = window.fetch.bind(window);
      window.__mode = 'stream';
      window.__posts = [];
      window.fetch = async (url, init) => {
        if (!String(url).match(/\/matrix\/chats\/[^/]+\/stream$/))
          return original(url, init);
        window.__posts.push({
          url: String(url),
          method: init.method,
          headers: init.headers,
          body: JSON.parse(init.body),
        });
        if (window.__mode === 'jsonError')
          return new Response(
            JSON.stringify({ error: 'وصلت سقف اليوم التجريبي' }),
            { status: 429, headers: { 'Content-Type': 'application/json' } },
          );
        let closed = false;
        const body = new ReadableStream({
          start(controller) {
            window.__emit = (event, split = false) => {
              const bytes = new TextEncoder().encode(
                'data: ' + JSON.stringify(event) + '\r\n\r\n',
              );
              if (split) {
                for (let i = 0; i < bytes.length; i++)
                  controller.enqueue(bytes.slice(i, i + 1));
              } else controller.enqueue(bytes);
            };
            window.__end = () => {
              closed = true;
              controller.close();
            };
            init.signal.addEventListener('abort', () => {
              window.__aborted = true;
              if (!closed) {
                closed = true;
                controller.error(new DOMException('Aborted', 'AbortError'));
              }
            });
          },
          cancel() {
            closed = true;
            window.__cancelled = true;
          },
        });
        return new Response(body, {
          headers: { 'Content-Type': 'text/event-stream' },
        });
      };
    });
    await page.route(base + '/brains-preview', (r) =>
      r.fulfill({
        contentType: 'text/html',
        body: `<html dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body><div id="root" style="padding:16px"></div><script type="module">import RR from '${base}/@react-refresh';RR.injectIntoGlobalHook(window);window.$RefreshReg$=()=>{};window.$RefreshSig$=()=>x=>x;window.__vite_plugin_react_preamble_installed__=true;const R=await import('${base}/node_modules/.vite/deps/react.js');const React=R.default||R;const D=await import('${base}/node_modules/.vite/deps/react-dom_client.js');const {createRoot}=D.default||D;const router=await import('${routerPath}');const {SessionContext}=await import('${base}/src/session.ts');await import('${base}/src/index.css');await import('${base}/src/fonts.css');const {default:Page}=await import('${base}/src/${screen}.tsx');createRoot(document.getElementById('root')).render(React.createElement(router.BrowserRouter,{basename:'/staff_mange'},React.createElement(SessionContext.Provider,{value:{employee:${JSON.stringify(actor)},permissions:[],setEmployee:()=>{},gpsServiceId:null}},React.createElement(Page))));</script></body></html>`,
      }),
    );
    await page.route('**/api/**', async (route) => {
      const req = route.request(),
        u = new URL(req.url()),
        p = u.pathname,
        method = req.method(),
        body = req.postData() ? req.postDataJSON() : null;
      calls.push({ p, method, body, q: u.search });
      let data = [];
      if (p === '/api/matrix/chats' && method === 'GET')
        data = {
          enabled: true,
          chats: [{ id: 'chat', title: 'محادثة تجريبية', updatedAt: now }],
        };
      else if (p === '/api/matrix/chats' && method === 'POST') {
        if (delayCreate)
          await new Promise((r) => {
            finishCreate = r;
          });
        data = { id: 'chat', title: 'محادثة تجريبية', updatedAt: now };
      } else if (p === '/api/matrix/chats/chat') data = [];
      else if (p.includes('/correct')) data = { ok: true };
      else if (p === '/api/ai/status') {
        assert.equal(actor.actualRole, 'OWNER', 'status must be owner only');
        data = {
          brains,
          matrix: { on: true },
          secretary: { on: true },
          quality: { on: true },
          booking: { on: true },
        };
      } else if (p === '/api/ai/agents') {
        assert.equal(actor.actualRole, 'OWNER');
        data = [
          { agent: 'MATRIX', name: 'ماتركس', editable: false },
          { agent: 'IRIS', name: 'ايريس', editable: false },
          { agent: 'QUALITY', name: 'وكيل الجودة', editable: false },
          { agent: 'BOOKING', name: 'وكيل الحجوزات', editable: false },
        ];
      } else if (p.startsWith('/api/ai/budgets/')) {
        assert.equal(actor.actualRole, 'OWNER');
        const brain = p.split('/').pop();
        brains[brain].dailyLimitUSD = body.dailyUSD;
        data = { brain, ...brains[brain] };
      } else if (p === '/api/ai/suggestions') {
        const result = items.filter(
          (x) =>
            (!u.searchParams.get('status') ||
              x.status === u.searchParams.get('status')) &&
            (!u.searchParams.get('needsCode') || x.needsCode),
        );
        if (delayFilter && u.searchParams.get('needsCode'))
          await new Promise((r) => {
            finishFilter = r;
          });
        data = result;
      } else if (p.endsWith('/decide')) {
        if (denyDecision)
          return route.fulfill({
            status: 500,
            contentType: 'application/json',
            body: JSON.stringify({ error: 'فشل قرار تجريبي' }),
          });
        items = items.map((x) =>
          x.id === p.split('/')[4] ? { ...x, status: body.status } : x,
        );
        data = { ok: true };
      }
      await route.fulfill({
        contentType: 'application/json',
        body: JSON.stringify(data),
      });
    });
    await page.goto(base + '/brains-preview');
    return {
      page,
      calls,
      releaseCreate: () => {
        delayCreate = false;
        finishCreate();
      },
      denyDecision: () => {
        denyDecision = true;
      },
      allowDecision: () => {
        denyDecision = false;
      },
      slowFilter: () => {
        delayFilter = true;
      },
      releaseFilter: () => {
        delayFilter = false;
        finishFilter();
      },
    };
  }
  const m = await mount('components/MatrixChat'),
    p = m.page;
  await p.getByLabel('رسالتك لماتركس').fill('رسالة فورية قبل إنشاء المحادثة');
  await p.getByRole('button', { name: 'إرسال', exact: true }).click();
  await p
    .getByText('رسالة فورية قبل إنشاء المحادثة', { exact: true })
    .waitFor();
  await p.getByRole('button', { name: /Thinking/ }).waitFor();
  assert.equal(await p.evaluate(() => window.__posts.length), 0);
  m.releaseCreate();
  await p.waitForFunction(() => window.__posts.length === 1);
  const post = await p.evaluate(() => window.__posts[0]);
  assert.equal(post.method, 'POST');
  assert.equal(post.headers.Authorization, 'Bearer mock-token');
  assert.equal(
    m.calls.filter((x) => x.p === '/api/matrix/chats/chat').length,
    0,
    'new history must not erase optimistic messages',
  );
  await p.evaluate(() =>
    window.__emit({ type: 'thinking', text: 'Check ' }, true),
  );
  await p.evaluate(() =>
    window.__emit({ type: 'thinking', text: 'company data.' }, true),
  );
  await p.evaluate(() =>
    window.__emit({ type: 'tool', name: 'booking_summary' }, true),
  );
  await p
    .locator('.matrix-thinking pre')
    .filter({ hasText: 'Check company data.' })
    .waitFor();
  await p
    .locator('.matrix-thinking pre')
    .filter({ hasText: 'checking booking_summary' })
    .waitFor();
  assert.equal(await p.locator('.matrix-thinking').getAttribute('dir'), 'ltr');
  await p.evaluate(() =>
    window.__emit({ type: 'text', text: 'جواب أولي' }, true),
  );
  await p.getByText('جواب أولي', { exact: true }).waitFor();
  await p.evaluate(() =>
    window.__emit({ type: 'text', text: 'جواب بديل كامل' }, true),
  );
  await p.getByText('جواب بديل كامل', { exact: true }).waitFor();
  assert.equal(await p.getByText('جواب أولي', { exact: true }).count(), 0);
  assert.equal(
    await p.getByRole('button', { name: '👎 غلط — صحّحه' }).count(),
    0,
  );
  await p.evaluate(
    (at) =>
      window.__emit(
        {
          type: 'done',
          message: {
            id: 'saved-answer',
            role: 'ASSISTANT',
            text: 'الجواب المحفوظ النهائي',
            steps: 'booking_summary',
            createdAt: at,
          },
        },
        true,
      ),
    now,
  );
  await p.getByText('الجواب المحفوظ النهائي', { exact: true }).waitFor();
  const thought = p.getByRole('button', { name: /Thought for \d+s · 1 tools/ });
  await thought.waitFor();
  assert.equal(await thought.getAttribute('aria-expanded'), 'false');
  await thought.click();
  await p.locator('.matrix-thinking pre').waitFor();
  await p.getByRole('button', { name: '👎 غلط — صحّحه' }).click();
  await p
    .getByPlaceholder('شنو الصح؟', { exact: false })
    .fill('تصحيح تجريبي بدون بيانات حقيقية');
  await p.getByRole('button', { name: 'علّمه', exact: true }).click();
  await p.getByText('✓ ماتركس تعلّم', { exact: true }).waitFor();
  assert(
    m.calls.some(
      (c) => c.p.includes('saved-answer') && c.p.includes('/correct'),
    ),
  );
  await p.screenshot({ path: '/tmp/matrix-stream-mobile.png' });
  const thinkingContrast = async () =>
    p
      .locator('.matrix-thinking')
      .first()
      .evaluate((el) => {
        const light = (rgb) =>
          rgb
            .match(/[\d.]+/g)
            .slice(0, 3)
            .map(Number)
            .map((v) => v / 255)
            .map((v) =>
              v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4,
            )
            .reduce((sum, v, i) => sum + v * [0.2126, 0.7152, 0.0722][i], 0);
        const bg = light(getComputedStyle(el).backgroundColor),
          fg = light(getComputedStyle(el.querySelector('pre')).color);
        return (Math.max(bg, fg) + 0.05) / (Math.min(bg, fg) + 0.05);
      });
  assert((await thinkingContrast()) >= 4.5, 'light Thinking contrast');
  await p.getByLabel('رسالتك لماتركس').fill('اختبار إيقاف');
  await p.getByRole('button', { name: 'إرسال', exact: true }).click();
  await p.waitForFunction(() => window.__posts.length === 2);
  await p.evaluate(() =>
    window.__emit({ type: 'text', text: 'رد جزئي لا ينحسب محفوظ' }),
  );
  await p.getByRole('button', { name: 'إيقاف', exact: true }).click();
  await p.waitForTimeout(300);
  await p.getByRole('alert').filter({ hasText: 'وقّفت الرد' }).waitFor();
  assert.equal(
    await p.getByLabel('رسالتك لماتركس').inputValue(),
    'اختبار إيقاف',
  );
  assert(await p.evaluate(() => window.__aborted));
  assert.equal(
    await p.evaluate(() => window.__posts.length),
    2,
    'stop must never auto-resend',
  );
  await p.getByLabel('رسالتك لماتركس').fill('خطأ بث');
  await p.getByRole('button', { name: 'إرسال', exact: true }).click();
  await p.waitForFunction(() => window.__posts.length === 3);
  await p.evaluate(() =>
    window.__emit({ type: 'error', text: 'فشل بث تجريبي' }, true),
  );
  await p.getByRole('alert').filter({ hasText: 'فشل بث تجريبي' }).waitFor();
  assert.equal(await p.getByLabel('رسالتك لماتركس').inputValue(), 'خطأ بث');
  await p.evaluate(() => (window.__mode = 'jsonError'));
  await p.getByRole('button', { name: 'إرسال', exact: true }).click();
  await p
    .getByRole('alert')
    .filter({ hasText: 'وصلت سقف اليوم التجريبي' })
    .waitFor();
  assert.equal(await p.getByLabel('رسالتك لماتركس').inputValue(), 'خطأ بث');
  await p.evaluate(() => (window.__mode = 'stream'));
  await p.getByRole('button', { name: 'إرسال', exact: true }).click();
  await p.waitForFunction(() => window.__posts.length === 5);
  await p.evaluate(() => window.__end());
  await p.getByRole('alert').filter({ hasText: 'انقطع البث' }).waitFor();
  await p.setViewportSize({ width: 1366, height: 900 });
  await p.evaluate(() => document.documentElement.classList.add('dark'));
  await p.screenshot({ path: '/tmp/matrix-stream-desktop-dark.png' });
  assert((await thinkingContrast()) >= 4.5, 'dark Thinking contrast');
  assert(
    await p.evaluate(() => document.documentElement.scrollWidth <= innerWidth),
  );
  const b = await mount('pages/AiAgentsPage');
  await b.page.getByText('chat-model', { exact: true }).waitFor();
  assert.equal(
    await b.page.getByRole('progressbar').first().getAttribute('aria-valuenow'),
    '50',
  );
  assert.equal(
    await b.page.getByRole('progressbar').last().getAttribute('aria-valuenow'),
    '100',
  );
  await b.page.getByLabel('السقف اليومي بالدولار — الداخلي').fill('0.09');
  await b.page
    .locator('.ai-brain-card')
    .first()
    .getByRole('button', { name: 'حفظ السقف' })
    .click();
  assert.equal(b.calls.filter((c) => c.method === 'PUT').length, 0);
  await b.page.getByLabel('السقف اليومي بالدولار — الداخلي').fill('4');
  await b.page
    .locator('.ai-brain-card')
    .first()
    .getByRole('button', { name: 'حفظ السقف' })
    .click();
  await b.page.getByText('انحفظ السقف اليومي.', { exact: true }).waitFor();
  assert.equal(b.calls.find((c) => c.method === 'PUT').body.dailyUSD, 4);
  await b.page.screenshot({ path: '/tmp/brains-budget-mobile.png' });
  const admin = await mount('pages/AiAgentsPage', {
    ...owner,
    actualRole: 'ADMIN',
  });
  await admin.page.getByText('هاي الشاشة لمالك النظام فقط.').waitFor();
  assert.equal(admin.calls.length, 0);
  const s = await mount('pages/AiSuggestionsPage', {
      ...owner,
      actualRole: 'ADMIN',
    }),
    q = s.page;
  await q
    .getByRole('heading', { name: 'اقتراح تجريبي', exact: true })
    .waitFor();
  s.slowFilter();
  const requested = q.waitForRequest((r) => r.url().includes('needsCode=1'));
  await q.getByRole('button', { name: 'يحتاج برمجة', exact: true }).click();
  await requested;
  await q.getByRole('button', { name: 'الجديد', exact: true }).click();
  await q
    .getByRole('heading', { name: 'اقتراح تجريبي', exact: true })
    .waitFor();
  s.releaseFilter();
  await q.waitForTimeout(100);
  assert.equal(
    await q.getByRole('heading', { name: 'اقتراح ثانٍ', exact: true }).count(),
    0,
  );
  await q.getByRole('button', { name: 'وافق', exact: true }).click();
  assert.equal(s.calls.filter((c) => c.method === 'POST').length, 0);
  await q.getByRole('button', { name: 'تأكيد القرار' }).click();
  await q.getByText('ماكو اقتراحات بهالفلتر لحد هسه.').waitFor();
  assert.equal(
    s.calls.find((c) => c.method === 'POST').body.status,
    'ACCEPTED',
  );
  await q.getByRole('button', { name: 'الكل', exact: true }).click();
  await q
    .getByRole('heading', { name: 'اقتراح تجريبي', exact: true })
    .waitFor();
  const first = q.locator('.ai-suggestion-card').filter({
    has: q.getByRole('heading', { name: 'اقتراح تجريبي', exact: true }),
  });
  s.denyDecision();
  await first.getByRole('button', { name: 'ارفض', exact: true }).click();
  await q.getByRole('button', { name: 'تأكيد القرار' }).click();
  await q
    .getByRole('dialog')
    .getByRole('alert')
    .filter({ hasText: 'فشل قرار تجريبي' })
    .waitFor();
  s.allowDecision();
  await q.getByRole('button', { name: 'تأكيد القرار' }).click();
  await first.getByText('مرفوض', { exact: true }).waitFor();
  await first.getByRole('button', { name: 'تم', exact: true }).click();
  await q.getByRole('button', { name: 'تأكيد القرار' }).click();
  await first.getByText('منجز', { exact: true }).waitFor();
  await q.screenshot({ path: '/tmp/ai-suggestions-mobile.png' });
  await q.setViewportSize({ width: 1366, height: 900 });
  await q.evaluate(() => document.documentElement.classList.add('dark'));
  await q.screenshot({ path: '/tmp/ai-suggestions-desktop-dark.png' });
  assert(
    await q.evaluate(() => document.documentElement.scrollWidth <= innerWidth),
  );
  for (const screen of [
    'components/MatrixChat',
    'pages/AiSuggestionsPage',
    'pages/AiAgentsPage',
  ]) {
    const denied = await mount(screen, {
      ...owner,
      role: 'TECHNICIAN',
      actualRole: null,
    });
    await denied.page.waitForTimeout(150);
    assert.equal(denied.calls.length, 0);
    assert.equal(await denied.page.evaluate(() => window.__posts.length), 0);
  }
  assert.deepEqual(errors, []);
  console.log(
    'PASS: optimistic before create, byte-fragmented SSE/UTF8/CRLF, accumulated thinking/tools, replaced text, saved ID correction, abort/error/JSON-limit/EOF recovery, owner-only status/budgets + progress, admin suggestions filter race/decision, denied zero requests, mobile/desktop/dark. Synthetic streaming transport and APIs.',
  );
  await browser.close();
})().catch((e) => {
  console.error(e);
  process.exit(1);
});
