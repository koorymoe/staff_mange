const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const assert = require('node:assert/strict');
(async () => {
  const browser = await chromium.launch({
    executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium',
    args: ['--no-sandbox'],
  });
  const base = process.env.IRIS_BASE_URL || 'http://127.0.0.1:5199/staff_mange',
    errors = [];
  const owner = {
    id: 'iris-owner-test',
    name: 'مدير تجريبي',
    role: 'ADMIN',
    actualRole: 'OWNER',
    certificate: 'دكتوراه',
  };
  const transformed = await (
    await fetch(base + '/src/pages/SecretaryPage.tsx')
  ).text();
  const routerPath = transformed.match(
    /["']([^"']*react-router-dom\.js[^"']*)["']/,
  )[1];
  const at = '2026-10-28T09:00:00Z';
  async function mount(actor = owner, empty = false) {
    const page = await browser.newPage({
      viewport: { width: 390, height: 844 },
    });
    page.on('pageerror', (e) => {
      errors.push(e.message);
      console.log(e.message);
    });
    const calls = [];
    let rows = empty
      ? []
      : [
          {
            id: 'conv-a',
            title: 'محادثة تجريبية',
            messages: 0,
            createdAt: at,
            updatedAt: at,
          },
        ];
    await page.addInitScript(() => {
      localStorage.setItem('authToken', 'iris-test-token');
      Object.defineProperty(navigator, 'mediaDevices', {
        value: {
          getUserMedia: async () => ({
            getTracks: () => [
              {
                stop() {
                  window.__tracksStopped = (window.__tracksStopped || 0) + 1;
                },
              },
            ],
          }),
        },
      });
      window.MediaRecorder = class {
        constructor(stream) {
          this.stream = stream;
          this.mimeType = 'audio/webm';
          this.state = 'inactive';
        }
        start() {
          this.state = 'recording';
        }
        stop() {
          this.state = 'inactive';
          this.ondataavailable({ data: new Blob(['synthetic audio']) });
          this.onstop();
        }
      };
      const original = window.fetch.bind(window);
      window.__posts = [];
      window.__mode = 'stream';
      window.fetch = async (url, init) => {
        if (!String(url).includes('/secretary/message?stream=1'))
          return original(url, init);
        const audio = init.body instanceof FormData;
        window.__posts.push({
          url: String(url),
          method: init.method,
          headers: init.headers,
          audio,
          body: audio
            ? {
                conversationId: init.body.get('conversationId'),
                size: init.body.get('file')?.size,
              }
            : JSON.parse(init.body),
        });
        if (window.__mode === 'jsonError')
          return new Response(JSON.stringify({ error: 'خطأ تحقق تجريبي' }), {
            status: 400,
            headers: { 'Content-Type': 'application/json' },
          });
        let closed = false;
        return new Response(
          new ReadableStream({
            start(c) {
              window.__emit = (event, split = true) => {
                const b = new TextEncoder().encode(
                  'data: ' + JSON.stringify(event) + '\r\n\r\n',
                );
                if (split) {
                  for (let i = 0; i < b.length; i++)
                    c.enqueue(b.slice(i, i + 1));
                } else c.enqueue(b);
              };
              window.__end = () => {
                closed = true;
                c.close();
              };
              init.signal.addEventListener('abort', () => {
                window.__aborted = true;
                if (!closed) {
                  closed = true;
                  c.error(new DOMException('Aborted', 'AbortError'));
                }
              });
            },
            cancel() {
              closed = true;
            },
          }),
          { headers: { 'Content-Type': 'text/event-stream' } },
        );
      };
    });
    await page.route(base + '/iris-stream-preview', (r) =>
      r.fulfill({
        contentType: 'text/html',
        body: `<html dir="rtl"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head><body><div id="root"></div><script type="module">import RR from '${base}/@react-refresh';RR.injectIntoGlobalHook(window);window.$RefreshReg$=()=>{};window.$RefreshSig$=()=>x=>x;window.__vite_plugin_react_preamble_installed__=true;const R=await import('${base}/node_modules/.vite/deps/react.js');const React=R.default||R;const D=await import('${base}/node_modules/.vite/deps/react-dom_client.js');const {createRoot}=D.default||D;const router=await import('${routerPath}');const {SessionContext}=await import('${base}/src/session.ts');await import('${base}/src/index.css');await import('${base}/src/fonts.css');const {default:Page}=await import('${base}/src/pages/SecretaryPage.tsx');window.__root=createRoot(document.getElementById('root'));window.__root.render(React.createElement(router.BrowserRouter,{basename:'/staff_mange'},React.createElement(SessionContext.Provider,{value:{employee:${JSON.stringify(actor)},permissions:[],setEmployee:()=>{},gpsServiceId:null}},React.createElement(Page))));</script></body></html>`,
      }),
    );
    await page.route('**/api/**', async (r) => {
      const req = r.request(),
        u = new URL(req.url()),
        p = u.pathname;
      calls.push({ p, method: req.method(), q: u.search });
      let data = [];
      if (p === '/api/ai/agent-names')
        data = { IRIS: 'كمال', MATRIX: 'ماتركس' };
      else if (p === '/api/secretary/conversations') data = rows;
      else if (p === '/api/secretary/today')
        data = {
          greeting: 'أهلاً دكتور',
          summary: { lines: [], cards: [] },
          pending: [],
        };
      else if (p === '/api/secretary/reminders')
        data = {
          counts: { meeting: 0, guest: 0, delete_request: 0, leave: 0 },
          items: [],
        };
      else if (p.endsWith('/cancel')) data = { ok: true };
      await r.fulfill({
        contentType: 'application/json',
        body: JSON.stringify(data),
      });
    });
    await page.goto(base + '/iris-stream-preview');
    return {
      page,
      calls,
      activate: () => {
        rows = [
          {
            id: 'conv-a',
            title: 'محادثة تجريبية',
            messages: 2,
            createdAt: at,
            updatedAt: at,
          },
        ];
      },
    };
  }
  const x = await mount(owner, true),
    p = x.page;
  const input = p.getByLabel('رسالتك لايريس');
  await input.fill('رسالة فورية');
  await p.getByRole('button', { name: 'إرسال', exact: true }).click();
  await p.getByText('رسالة فورية', { exact: true }).waitFor();
  await p.getByRole('button', { name: /Thinking/ }).waitFor();
  await p.waitForFunction(() => window.__posts.length === 1);
  let post = await p.evaluate(() => window.__posts[0]);
  assert.equal(post.method, 'POST');
  assert.equal(post.headers.Authorization, 'Bearer iris-test-token');
  assert.equal(post.headers.Accept, 'text/event-stream');
  assert.equal(post.body.conversationId, undefined);
  await p.evaluate(() => window.__emit({ type: 'thinking', text: 'Check ' }));
  await p.evaluate(() =>
    window.__emit({ type: 'thinking', text: 'company status.' }),
  );
  await p.evaluate(() =>
    window.__emit({ type: 'tool', name: 'company_summary' }),
  );
  await p
    .locator('.iris-thinking pre')
    .filter({ hasText: 'Check company status.' })
    .waitFor();
  await p
    .locator('.iris-thinking pre')
    .filter({ hasText: 'checking company_summary' })
    .waitFor();
  assert.equal(await p.locator('.iris-thinking').getAttribute('dir'), 'ltr');
  await p.evaluate(() => window.__emit({ type: 'text', text: 'جواب أولي' }));
  await p.getByText('جواب أولي', { exact: true }).waitFor();
  await p.evaluate(() =>
    window.__emit({ type: 'text', text: 'جواب بديل كامل' }),
  );
  await p.getByText('جواب بديل كامل', { exact: true }).waitFor();
  assert.equal(await p.getByText('جواب أولي', { exact: true }).count(), 0);
  assert.equal(await p.locator('.sec-result').count(), 0);
  assert.equal(await p.getByRole('dialog').count(), 0);
  const reply = {
    messageId: 'saved-iris',
    conversationId: 'conv-a',
    transcript: null,
    reply: 'الرد النهائي المحفوظ',
    cards: [
      {
        kind: 'summary',
        title: 'اجتماع تجريبي',
        status: 'جاهز',
        owner: 'الإدارة',
        updatedAt: at,
        lines: ['تفاصيل اصطناعية للاختبار'],
        link: null,
      },
    ],
    proposal: {
      actionId: 'proposal-test',
      kind: 'task',
      fields: {
        employeeId: '',
        employeeName: '',
        title: 'مهمة تجريبية',
        description: 'تفاصيل تجريبية',
        priority: 'NORMAL',
        dueAt: '',
      },
      candidates: [
        { id: 'a', name: 'فني تجريبي ألف' },
        { id: 'b', name: 'فني تجريبي باء' },
      ],
    },
  };
  x.activate();
  await p.evaluate(
    (message) => window.__emit({ type: 'done', message }),
    reply,
  );
  await p.getByRole('heading', { name: 'فهمت منك…' }).waitFor();
  assert.equal(
    await p
      .getByRole('dialog')
      .getByLabel('الموظف', { exact: true })
      .locator('option')
      .count(),
    3,
  );
  assert.equal(
    await p.getByRole('dialog').getByLabel('عنوان المهمة').inputValue(),
    'مهمة تجريبية',
  );
  assert(!x.calls.some((c) => c.p.endsWith('/confirm')));
  await p
    .getByRole('dialog')
    .getByRole('button', { name: 'إلغاء', exact: true })
    .click();
  await p.getByText('الرد النهائي المحفوظ', { exact: true }).waitFor();
  await p
    .getByRole('heading', { name: 'اجتماع تجريبي', exact: true })
    .waitFor();
  const thought = p.getByRole('button', { name: /Thought for \d+s · 1 tools/ });
  await thought.waitFor();
  assert.equal(await thought.getAttribute('aria-expanded'), 'false');
  await thought.click();
  await p.locator('.iris-thinking pre').waitFor();
  assert(
    !x.calls.some((c) => c.p === '/api/secretary/history'),
    'new conversation history must not overwrite streamed done/cards/trace',
  );
  assert.equal(await p.getByText('رسالة فورية', { exact: true }).count(), 1);
  assert.equal(
    await p.evaluate(() =>
      localStorage.getItem('iris-conversation:iris-owner-test'),
    ),
    'conv-a',
  );
  await p.screenshot({ path: '/tmp/iris-stream-mobile.png' });
  const contrast = async () =>
    p
      .locator('.iris-thinking')
      .first()
      .evaluate((el) => {
        const lum = (rgb) =>
          rgb
            .match(/[\d.]+/g)
            .slice(0, 3)
            .map(Number)
            .map((v) => v / 255)
            .map((v) =>
              v <= 0.04045 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4,
            )
            .reduce((sum, v, i) => sum + v * [0.2126, 0.7152, 0.0722][i], 0);
        const bg = lum(getComputedStyle(el).backgroundColor),
          fg = lum(getComputedStyle(el.querySelector('pre')).color);
        return (Math.max(bg, fg) + 0.05) / (Math.min(bg, fg) + 0.05);
      });
  assert((await contrast()) >= 4.5, 'light Thinking contrast');
  await p.getByRole('button', { name: 'ابدأ التسجيل' }).click();
  await p.getByRole('button', { name: 'وقف التسجيل وأرسل' }).click();
  await p.getByText('رسالة صوتية — دا أحوّلها لنص…', { exact: true }).waitFor();
  await p.waitForFunction(() => window.__posts.length === 2);
  post = await p.evaluate(() => window.__posts[1]);
  assert(post.audio && post.body.size > 0);
  assert.equal(post.body.conversationId, 'conv-a');
  assert.equal(post.headers['Content-Type'], undefined);
  await p.evaluate((message) => window.__emit({ type: 'done', message }), {
    ...reply,
    messageId: 'voice-ok',
    transcript: 'نص البصمة التجريبي',
    reply: 'فهمت البصمة',
    cards: [],
    proposal: null,
  });
  await p.getByText('البصمة: نص البصمة التجريبي', { exact: true }).waitFor();
  await p.getByText('فهمت البصمة', { exact: true }).waitFor();
  assert(await p.evaluate(() => window.__tracksStopped > 0));
  await input.fill('اختبار إيقاف النص');
  await p.getByRole('button', { name: 'إرسال', exact: true }).click();
  await p.waitForFunction(() => window.__posts.length === 3);
  await p.evaluate(() => window.__emit({ type: 'text', text: 'رد جزئي' }));
  await p.getByRole('button', { name: 'إيقاف الرد' }).click();
  await p.getByRole('alert').filter({ hasText: 'وقّفت الرد' }).waitFor();
  await p.waitForTimeout(150);
  assert.equal(await p.evaluate(() => window.__posts.length), 3);
  assert.equal(await input.inputValue(), 'اختبار إيقاف النص');
  assert(await p.evaluate(() => window.__aborted));
  await p.getByText('الرد غير مكتمل.', { exact: false }).waitFor();
  await p.evaluate(() => (window.__mode = 'jsonError'));
  await p.getByRole('button', { name: 'إرسال', exact: true }).click();
  await p.getByRole('alert').filter({ hasText: 'خطأ تحقق تجريبي' }).waitFor();
  assert.equal(await input.inputValue(), 'اختبار إيقاف النص');
  await p.evaluate(() => (window.__mode = 'stream'));
  await p.getByRole('button', { name: 'إرسال', exact: true }).click();
  await p.waitForFunction(() => window.__posts.length === 5);
  await p.evaluate(() =>
    window.__emit({ type: 'error', text: 'خطأ بث اصطناعي' }),
  );
  await p.getByRole('alert').filter({ hasText: 'خطأ بث اصطناعي' }).waitFor();
  assert.equal(await input.inputValue(), 'اختبار إيقاف النص');
  await p.getByRole('button', { name: 'إرسال', exact: true }).click();
  await p.waitForFunction(() => window.__posts.length === 6);
  await p.evaluate(() => window.__end());
  await p.getByRole('alert').filter({ hasText: 'انقطع البث' }).waitFor();
  await p.getByRole('button', { name: 'ابدأ التسجيل' }).click();
  await p.getByRole('button', { name: 'وقف التسجيل وأرسل' }).click();
  await p.waitForFunction(() => window.__posts.length === 7);
  await p.getByRole('button', { name: 'إيقاف الرد' }).click();
  await p.getByRole('button', { name: 'إعادة إرسال البصمة' }).waitFor();
  assert.equal(await p.evaluate(() => window.__posts.length), 7);
  await p.getByRole('button', { name: 'إعادة إرسال البصمة' }).click();
  await p.waitForFunction(() => window.__posts.length === 8);
  assert((await p.evaluate(() => window.__posts[7])).audio);
  await p.evaluate((message) => window.__emit({ type: 'done', message }), {
    ...reply,
    messageId: 'voice-retried',
    transcript: 'بصمة إعادة المحاولة',
    reply: 'تمت إعادة المحاولة',
    cards: [],
    proposal: null,
  });
  await p.getByText('تمت إعادة المحاولة', { exact: true }).waitFor();
  assert.equal(
    await p.getByRole('button', { name: 'إعادة إرسال البصمة' }).count(),
    0,
  );
  await p.setViewportSize({ width: 1366, height: 900 });
  await p.evaluate(() => document.documentElement.classList.add('dark'));
  await p.screenshot({ path: '/tmp/iris-stream-desktop-dark.png' });
  assert((await contrast()) >= 4.5, 'dark Thinking contrast');
  assert(
    await p.evaluate(() => document.documentElement.scrollWidth <= innerWidth),
  );
  const admin = await mount({ ...owner, actualRole: 'ADMIN' });
  await admin.page.getByLabel('رسالتك لايريس').fill('اختبار الأدمن');
  await admin.page.getByRole('button', { name: 'إرسال', exact: true }).click();
  await admin.page.waitForFunction(() => window.__posts.length === 1);
  assert(
    !admin.calls.some(
      (c) => c.p.startsWith('/api/ai/') && c.p !== '/api/ai/agent-names',
    ),
  );
  await admin.page.evaluate(() => window.__root.unmount());
  assert(
    await admin.page.evaluate(() => window.__aborted),
    'unmount must abort',
  );
  const denied = await mount({
    ...owner,
    role: 'TECHNICIAN',
    actualRole: null,
  });
  await denied.page.getByText('ايريس خاص بالمالك ومدير النظام.').waitFor();
  assert.equal(denied.calls.length, 0);
  assert.equal(await denied.page.evaluate(() => window.__posts.length), 0);
  assert.deepEqual(errors, []);
  console.log(
    'PASS: Iris immediate/new-conversation SSE, fragmented UTF8, additive thinking/tools, replaced text, done cards/proposal/candidates, no history overwrite, scoped multipart/transcript, stop without retry, JSON/SSE/EOF errors restore draft, explicit in-memory audio retry, admin access/denied zero calls, unmount abort, mobile/desktop/dark. Synthetic transport/media/APIs.',
  );
  await browser.close();
})().catch((e) => {
  console.error(e);
  process.exit(1);
});
