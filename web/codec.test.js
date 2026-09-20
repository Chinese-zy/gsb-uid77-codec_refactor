// JS 侧的写死字节对照表，与 frame/codec_test.go 独立维护、内容相同。
// 运行：node --test web/
const test = require("node:test");
const assert = require("node:assert/strict");
const { decode, encode } = require("./codec.js");

function f(id, value) {
  return { id, name: require("./codec.js").nameOf(id), value };
}

function bytes(hex) {
  const clean = hex.replace(/\s+/g, "");
  const u8 = new Uint8Array(clean.length / 2);
  for (let i = 0; i < u8.length; i++) {
    u8[i] = parseInt(clean.slice(i * 2, i * 2 + 2), 16);
  }
  return u8;
}

function hex(u8) {
  return Array.from(u8).map((b) => b.toString(16).padStart(2, "0")).join("");
}

// [名称, 输入字节, 期望结果, 是否合法报文]；非法报文只要求安全停下、保序，
// 不要求 encode 还原坏字节（与 frame/codec_test.go 的规矩一致）。
const cases = [
  ["不足两字节", "01", { version: 0, fields: [] }, false],
  ["空帧只带版本", "0100", { version: 1, fields: [] }, true],
  ["标量小端", "0200" + "030104000000", { version: 2, fields: [f(3, 4)] }, true],
  ["版本大小端区分", "0001", { version: 256, fields: [] }, true],
  ["标量高位字节", "0000" + "0501fcfdfcfc",
    { version: 0, fields: [f(5, 0xfcfcfdfc)] }, true],
  ["重复字段保序不去重", "0000" + "010101000000" + "010102000000",
    { version: 0, fields: [f(1, 1), f(1, 2)] }, true],
  ["LEB128 温度 id 150", "0000" + "96010107000000",
    { version: 0, fields: [f(150, 7)] }, true],
  ["LEB128 多续位", "0000" + "8080080101000000",
    { version: 0, fields: [f(0x20000, 1)] }, true],
  ["成组含两字段", "0100" + "02030c" + "030104000000" + "040105000000",
    { version: 1, fields: [f(2, [f(3, 4), f(4, 5)])] }, true],
  ["成组后还可接标量", "0100" + "020300" + "030109000000",
    { version: 1, fields: [f(2, []), f(3, 9)] }, true],
  ["截断标量丢弃当前字段", "0000" + "03010400", { version: 0, fields: [] }, false],
  ["悬垂 varint 不炸", "0000" + "9680", { version: 0, fields: [] }, false],
  ["未知 kind 停止", "0000" + "0102", { version: 0, fields: [] }, false],
  ["成组长度越界", "0000" + "020310" + "030104000000",
    { version: 0, fields: [] }, false],
  ["成组内层截断整体丢弃", "0000" + "020303" + "0301",
    { version: 0, fields: [] }, false],
];

test("写死字节对照表：decode 与合法报文重编码", () => {
  for (const [name, rawHex, want, valid] of cases) {
    const got = decode(bytes(rawHex));
    assert.deepEqual(got, want, `decode: ${name}`);
    if (valid) {
      assert.equal(hex(encode(got)), rawHex, `encode: ${name}`);
    }
  }
});

test("encode 拒绝非法 value", () => {
  assert.throws(() => encode({
    version: 0,
    fields: [{ id: 1, name: "电压", value: 1.5 }],
  }));
});

test("sample.bin 基线", () => {
  const fs = require("node:fs");
  const path = require("node:path");
  const raw = new Uint8Array(fs.readFileSync(path.join(__dirname, "sample.bin")));
  const want = {
    version: 1,
    fields: [
      f(2, [f(3, 5), f(4, 6), f(5, 7), f(6, 8)]),
      f(1, 1000),
      f(8, 3),
      f(150, 7),
    ],
  };
  assert.deepEqual(decode(raw), want);
  assert.deepEqual(Array.from(encode(decode(raw))), Array.from(raw));
});
