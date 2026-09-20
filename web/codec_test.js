// 页面这一侧跑和接口同一份字节对照表：node web/codec_test.js

const assert = require("node:assert");
const fs = require("node:fs");
const path = require("node:path");
const { decode } = require("./codec.js");

const golden = JSON.parse(
  fs.readFileSync(path.join(__dirname, "..", "testdata", "golden.json"), "utf8"),
);

for (const c of golden.cases) {
  const u8 = Uint8Array.from(Buffer.from(c.hex.replace(/\s+/g, ""), "hex"));
  const got = JSON.parse(JSON.stringify(decode(u8)));
  assert.deepStrictEqual(got, c.expect, `对不上：${c.name}`);
}

console.log(`ok ${golden.cases.length} cases`);
