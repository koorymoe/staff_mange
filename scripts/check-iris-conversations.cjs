const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const assert = require('node:assert/strict');
(async () => {
  const browser = await chromium.launch({ executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium', args: ['--no-sandbox'] });
  const page = await browser.newPage({ viewport: { width: 390, height: 844 } });
  const errors = [], requests = [];
  page.on('pageerror', e => errors.push(e.message));
  const stamp = new Date().toISOString();
  let rows = ['a', 'b'].map(id => ({ id, title: `محادثة ${id}`, messages: 1, createdAt: stamp, updatedAt: stamp }));
  const history = Object.fromEntries(rows.map(x => [x.id, [{ id: x.id, at: stamp, role: 'assistant', text: `رسائل ${x.id}`, transcript: null, cards: [], action: null }]]));
  let slowHistory = false, finishHistory, finishMessage;
  await page.addInitScript(() => {
    localStorage.setItem('authToken', 'preview');
    localStorage.setItem('currentEmployee', JSON.stringify({ id: 'test', name: 'مدير تجريبي', role: 'ADMIN', actualRole: 'OWNER', skills: [] }));
    if (!localStorage.getItem('iris-conversation:test')) localStorage.setItem('iris-conversation:test', 'b');
    Object.defineProperty(navigator, 'mediaDevices', { value: { getUserMedia: async () => ({ getTracks: () => [{ stop() {} }] }) } });
    window.MediaRecorder = class {
      constructor(stream) { this.stream = stream; this.mimeType = 'audio/webm'; this.state = 'inactive'; }
      start() { this.state = 'recording'; }
      stop() { this.state = 'inactive'; this.ondataavailable({ data: new Blob(['test audio']) }); this.onstop(); }
    };
  });
  await page.route('**/api/**', async route => {
    const req = route.request(), url = new URL(req.url()), path = url.pathname, method = req.method();
    requests.push({ path, method, body: req.postData() });
    let data = [];
    if (path === '/api/auth/me') data = { id: 'test', name: 'مدير تجريبي', role: 'ADMIN', actualRole: 'OWNER', skills: [] };
    if (path === '/api/secretary/today') data = { greeting: 'أهلاً سيدنا', summary: { lines: [], cards: [] }, pending: [] };
    if (path === '/api/secretary/reminders') data = { counts: { meeting: 0, guest: 0, leave: 0, delete_request: 0 }, items: [] };
    if (path === '/api/notifications') data = { notifications: [], unreadCount: 0 };
    if (path === '/api/secretary/conversations') {
      if (method === 'POST') { const c = { id: 'c', title: 'محادثة جديدة', messages: 0, createdAt: stamp, updatedAt: stamp }; rows.unshift(c); history.c = []; data = c; }
      else data = rows;
    }
    if (path.startsWith('/api/secretary/conversations/')) {
      const id = path.split('/').pop();
      if (method === 'PATCH') rows = rows.map(x => x.id === id ? { ...x, title: req.postDataJSON().title } : x);
      if (method === 'DELETE') rows = rows.filter(x => x.id !== id);
      data = { ok: true };
    }
    if (path === '/api/secretary/history') {
      const id = url.searchParams.get('conversationId');
      assert(id, 'history must be scoped');
      if (slowHistory && id === 'a') await new Promise(r => { finishHistory = r; });
      data = history[id];
    }
    if (path === '/api/secretary/message') {
      const voice = req.headers()['content-type'].includes('multipart');
      const id = voice ? 'c' : req.postDataJSON().conversationId;
      if (voice) assert(req.postData().includes('name="conversationId"\r\n\r\nc'), 'voice must include active ID');
      else { assert.equal(id, 'c'); await new Promise(r => { finishMessage = r; }); }
      data = { messageId: voice ? 'voice-reply' : 'reply', conversationId: id, reply: voice ? 'استلمت الصوت' : 'استلمت النص', transcript: voice ? 'صوت تجريبي' : null, cards: [], proposal: null };
      rows = rows.map(x => x.id === 'c' ? { ...x, title: 'العنوان التلقائي' } : x);
    }
    await route.fulfill({ contentType: 'application/json', body: JSON.stringify(data) });
  });
  const openDrawer = async () => { if ((await page.viewportSize()).width < 900) await page.getByRole('button', { name: 'فتح القائمة', exact: true }).click(); };
  await page.goto(process.env.IRIS_URL || 'http://127.0.0.1:5199/staff_mange/secretary');
  await page.getByText('رسائل b', { exact: true }).waitFor();
  await openDrawer();
  assert.equal(await page.locator('.iris-conversation.selected strong').innerText(), 'محادثة b');
  slowHistory = true;
  await page.locator('.iris-conversation-select').filter({ hasText: 'محادثة a' }).click();
  await page.getByText('دا أجيب رسائل المحادثة…').waitFor();
  await openDrawer();
  // History loading locks sends but must not prevent choosing a different conversation.
  await page.locator('.iris-conversation-select').filter({ hasText: 'محادثة b' }).click();
  await page.getByText('رسائل b', { exact: true }).waitFor();
  finishHistory();
  slowHistory = false;
  await page.waitForTimeout(100);
  assert.equal(await page.getByText('رسائل a', { exact: true }).count(), 0);
  await openDrawer();
  await page.getByRole('button', { name: '＋ محادثة جديدة', exact: true }).click();
  await page.getByLabel('رسالتك لايريس').fill('رسالة فورية بمحادثة جديدة');
  await page.getByRole('button', { name: 'إرسال', exact: true }).click();
  await page.getByText('رسالة فورية بمحادثة جديدة', { exact: true }).waitFor();
  await page.getByText('ايريس يفكّر…', { exact: true }).waitFor();
  finishMessage();
  await page.getByText('استلمت النص', { exact: true }).waitFor();
  await openDrawer();
  await page.getByRole('button', { name: 'خيارات العنوان التلقائي' }).click();
  await page.getByRole('button', { name: 'تسمية', exact: true }).click();
  await page.getByRole('dialog').getByLabel('عنوان المحادثة').fill('اسم جديد');
  await page.getByRole('button', { name: 'حفظ الاسم' }).click();
  await page.getByRole('button', { name: 'خيارات اسم جديد' }).waitFor();
  await page.getByRole('button', { name: 'غلق القائمة', exact: true }).last().click();
  const mic = page.locator('.sec-mic'); await mic.click(); await mic.click();
  await page.getByText('استلمت الصوت', { exact: true }).waitFor();
  await openDrawer();
  await page.getByRole('button', { name: 'خيارات العنوان التلقائي' }).click();
  await page.getByRole('button', { name: 'حذف', exact: true }).click();
  assert.equal(requests.filter(x => x.method === 'DELETE').length, 0);
  await page.getByRole('dialog').getByRole('button', { name: 'إلغاء', exact: true }).click();
  await page.getByRole('button', { name: 'خيارات العنوان التلقائي' }).click();
  await page.getByRole('button', { name: 'حذف', exact: true }).click();
  await page.getByRole('button', { name: 'نعم، احذف المحادثة' }).click();
  await page.getByText('رسائل a', { exact: true }).waitFor();
  await page.reload();
  await page.getByText('رسائل a', { exact: true }).waitFor();
  assert.equal(await page.evaluate(() => localStorage.getItem('iris-conversation:test')), 'a');
  assert(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth));
  await openDrawer();
  await page.waitForTimeout(250);
  await page.screenshot({ path: '/tmp/iris-conversations-mobile.png' });
  await page.setViewportSize({ width: 1366, height: 900 });
  await page.waitForTimeout(250);
  assert.equal(Math.round((await page.locator('.sec-drawer').boundingBox()).x), 0);
  await page.locator('.iris-conversation-select').filter({ hasText: 'محادثة b' }).click();
  await page.getByText('رسائل b', { exact: true }).waitFor();
  await page.screenshot({ path: '/tmp/iris-conversations-desktop.png' });
  await page.evaluate(() => document.documentElement.classList.add('dark'));
  await page.getByRole('button', { name: 'خيارات محادثة b' }).click();
  await page.getByRole('button', { name: 'تسمية', exact: true }).click();
  await page.getByRole('dialog').getByLabel('عنوان المحادثة').fill('تجربة الشاشة العريضة');
  await page.getByRole('button', { name: 'حفظ الاسم' }).click();
  await page.getByRole('button', { name: 'خيارات تجربة الشاشة العريضة' }).waitFor();
  await page.screenshot({ path: '/tmp/iris-conversations-dark.png' });
  assert.deepEqual(errors, []);
  console.log('PASS: persisted selection, scoped histories, rapid switching/late history, new chat, immediate text, scoped JSON + multipart voice, auto title refresh, rename, delete cancel/confirm, reload, mobile/desktop, no overflow. Mock API/media.');
  await browser.close();
})().catch(e => { console.error(e); process.exit(1); });
