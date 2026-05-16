const test = require("node:test");
const assert = require("node:assert/strict");

test("desktop bridge reports unsupported when bridge is absent", async () => {
  delete global.xuvaDesktop;
  delete global.XuvaDesktop;
  delete require.cache[require.resolve("./modules/desktop-bridge.js")];
  const bridge = require("./modules/desktop-bridge.js");

  assert.deepEqual(bridge.capabilities(), {
    available: false,
    canPickFolder: false,
    canRestart: false,
  });
  assert.deepEqual(await bridge.restartServer(), { supported: false });
});

test("desktop bridge restart calls bridge restartServer when available", async () => {
  let called = 0;
  global.xuvaDesktop = {
    restartServer: async () => {
      called += 1;
    },
  };
  delete global.XuvaDesktop;
  delete require.cache[require.resolve("./modules/desktop-bridge.js")];
  const bridge = require("./modules/desktop-bridge.js");

  assert.equal(bridge.capabilities().canRestart, true);
  assert.deepEqual(await bridge.restartServer(), { supported: true });
  assert.equal(called, 1);
});
