// 先启动 demo，并用 playwright-cli open http://127.0.0.1:8080 打开浏览器。
// 运行：playwright-cli run-code --filename captcha/example/slider/browser-regression.js
async (page) => {
  const drag = async () => {
    const slider = page.locator('#slider');
    await slider.focus();
    await slider.press('ArrowRight');
    await page.waitForTimeout(160);
    await slider.press('ArrowRight');
    await page.waitForTimeout(210);
    await slider.press('ArrowRight');
    await page.waitForTimeout(100);
    await slider.press('End');
  };

  // 同时覆盖关闭按钮和 Escape，以及旧请求的成功、失败响应。
  for (const cancel of ['button', 'escape']) {
    await page.goto('http://127.0.0.1:8080');
    await page.waitForFunction(() => !document.querySelector('#slider').disabled);
    let releaseOld, releaseNew, oldReceived, oldFinished;
    const oldGate = new Promise(resolve => { releaseOld = resolve; });
    const newGate = new Promise(resolve => { releaseNew = resolve; });
    const received = new Promise(resolve => { oldReceived = resolve; });
    const finished = new Promise(resolve => { oldFinished = resolve; });
    const verifyRoute = async route => {
      const response = await route.fetch();
      oldReceived();
      await oldGate;
      try {
        if (cancel === 'escape') await route.fulfill({ status: 500, body: 'old failure' });
        else await route.fulfill({ response });
      } catch {
        // 关闭弹窗会中止旧请求，浏览器可能已撤销该拦截。
      } finally {
        oldFinished();
      }
    };
    const generateRoute = async route => {
      const response = await route.fetch();
      await newGate;
      await route.fulfill({ response });
    };
    await page.route('**/api/captcha/verify', verifyRoute);
    try {
      await drag();
      await received;
      await page.route('**/api/captcha/generate', generateRoute);
      if (cancel === 'button') await page.locator('#close').click();
      else await page.keyboard.press('Escape');
      const generated = page.waitForRequest(r => r.url().endsWith('/api/captcha/generate'), { timeout: 3000 });
      await page.locator('#open').click();
      await generated;
      releaseOld();
      await finished;
      await page.waitForTimeout(50);
      const pending = await page.evaluate(() => ({
        sliderDisabled: document.querySelector('#slider').disabled,
        refreshDisabled: document.querySelector('#refresh').disabled,
        status: document.querySelector('#status').textContent,
      }));
      if (!pending.sliderDisabled || !pending.refreshDisabled || pending.status !== '正在准备验证') {
        throw new Error(`旧响应修改了新弹窗：${JSON.stringify(pending)}`);
      }
      releaseNew();
      await page.waitForFunction(() => !document.querySelector('#slider').disabled);
    } finally {
      releaseOld();
      releaseNew();
      await page.unroute('**/api/captcha/verify', verifyRoute);
      await page.unroute('**/api/captcha/generate', generateRoute);
    }
    await drag();
    await page.waitForFunction(() => document.querySelector('#track').classList.contains('success'));
  }
  return '关闭按钮、Escape、旧成功/失败响应隔离、新挑战验证均通过';
}
