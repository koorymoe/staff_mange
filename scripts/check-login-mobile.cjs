const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const assert = require('node:assert/strict');
(async () => {
  const browser = await chromium.launch({ executablePath: process.env.CHROMIUM_PATH || '/usr/bin/chromium', args: ['--no-sandbox'] });
  const page = await browser.newPage({ viewport: { width: 390, height: 844 }, isMobile: true, hasTouch: true });
  const errors = []; page.on('pageerror', e => errors.push(e.message));
  await page.route('**/api/**', route => route.fulfill({ status: 401, contentType: 'application/json', body: JSON.stringify({ error: 'بيانات تجربة غير صحيحة' }) }));
  await page.goto(process.env.LOGIN_URL || 'http://127.0.0.1:5199/staff_mange/');
  await page.locator('#lx-user').waitFor();
  assert((await page.locator('meta[name=viewport]').getAttribute('content')).includes('maximum-scale=1'));
  for (const width of [277, 320, 375, 390, 430, 844, 1366]) {
    await page.setViewportSize({ width, height: width === 844 ? 390 : 844 });
    await page.locator('#lx-user').fill('test-user');
    await page.locator('#lx-pass').fill('synthetic-password');
    await page.getByRole('button', { name: 'إظهار كلمة المرور', exact: true }).click();
    await page.getByRole('button', { name: 'إخفاء كلمة المرور', exact: true }).click();
    const result = await page.evaluate(() => {
      const card = document.querySelector('.lx-card').getBoundingClientRect();
      return {
        overflow: document.documentElement.scrollWidth > innerWidth,
        cardInside: card.x >= 0 && card.right <= innerWidth,
        inputs: [...document.querySelectorAll('.lx-field input,.cmdlogin-input')].map(el => ({ font: getComputedStyle(el).fontSize, width: el.getBoundingClientRect().width })),
        contained: [...document.querySelectorAll('.lx-field')].every(el => { const a = el.getBoundingClientRect(), b = el.querySelector('input').getBoundingClientRect(); return a.left >= card.left && a.right <= card.right && b.left >= a.left && b.right <= a.right; }),
        scale: visualViewport.scale,
        transform: getComputedStyle(document.querySelector('.lx-card')).transform,
        touch: getComputedStyle(document.querySelector('.lx')).touchAction,
      };
    });
    assert.equal(result.overflow, false, `${width}: document overflow`);
    assert(result.cardInside && result.contained, `${width}: fields/card must fit`);
    assert(result.inputs.every(x => x.font === '16px'));
    assert.equal(result.transform, 'none');
    assert.equal(result.touch, 'pan-y');
    assert.equal(result.scale, 1);
    if (width === 277) await page.screenshot({ path: '/tmp/login-narrow-fixed.png' });
  }
  await page.setViewportSize({ width: 390, height: 844 });
  await page.evaluate(() => document.documentElement.classList.add('dark'));
  assert.equal(await page.locator('#lx-user').evaluate(el => getComputedStyle(el).backgroundColor), 'rgba(0, 0, 0, 0)');
  await page.getByRole('button', { name: 'تسجيل الدخول', exact: true }).click();
  await page.getByRole('alert').filter({ hasText: 'بيانات تجربة غير صحيحة' }).waitFor();
  await page.screenshot({ path: '/tmp/login-mobile-fixed.png' });
  await page.getByRole('button', { name: 'ارفع الستارة' }).click();
  await page.waitForTimeout(700);
  await page.locator('.cmdlogin-input').first().fill('command-test');
  await page.locator('.cmdlogin-input').last().fill('synthetic-password');
  assert.equal(await page.locator('.cmdlogin-input').first().evaluate(el => getComputedStyle(el).fontSize), '16px');
  assert.equal(await page.evaluate(() => visualViewport.scale), 1);
  assert.deepEqual(errors, []);
  console.log('PASS: widths 277/320/375/390/430/844/1366, contained fields, 16px both login forms, fixed scale on focus/reveal, viewport lock, no horizontal overflow, dark input fix, validation failure. Chromium emulation, not real iPhone.');
  await browser.close();
})().catch(e => { console.error(e); process.exit(1); });
