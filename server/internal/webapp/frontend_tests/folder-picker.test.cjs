const test = require("node:test");
const assert = require("node:assert/strict");

test("folder picker reports unsupported without desktop bridge", async () => {
  delete global.lorivoDesktop;
  delete global.LorivoDesktop;
  delete require.cache[require.resolve("./modules/folder-picker.js")];
  const picker = require("./modules/folder-picker.js");

  assert.deepEqual(await picker.pickFolder({ title: "Library" }), { supported: false, path: "" });
});

test("folder picker normalizes desktop bridge path results", async () => {
  global.lorivoDesktop = {
    pickFolder: async request => ({ path: `  ${request.currentPath}\\Movies  ` }),
  };
  delete global.LorivoDesktop;
  delete require.cache[require.resolve("./modules/folder-picker.js")];
  const picker = require("./modules/folder-picker.js");

  assert.equal(picker.normalizePickedPath({ path: "  C:\\Media  " }), "C:\\Media");
  assert.deepEqual(await picker.pickFolder({ title: "Library", currentPath: "D:\\Media", purpose: "library" }), {
    supported: true,
    path: "D:\\Media\\Movies",
  });
});
