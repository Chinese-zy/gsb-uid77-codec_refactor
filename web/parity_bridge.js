// 对拍桥：从 stdin 读 [{name, hex}]，用页面侧同一套 codec 解码再编码，
// 输出 [{name, frame, reenc}]。Go 测试拿它和 frame.Decode/Encode 的结果逐项比对。
const fs = require("node:fs");
const path = require("node:path");
const Codec = require(path.join(__dirname, "codec.js"));

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

const cases = JSON.parse(fs.readFileSync(0, "utf8"));
const result = cases.map((tc) => {
  const frame = Codec.decode(bytes(tc.hex));
  return { name: tc.name, frame, reenc: hex(Codec.encode(frame)) };
});
process.stdout.write(JSON.stringify(result));
