// 报文编解码，规矩见 SPEC.md，对照表见 testdata/golden.json。
// 本文件同时给页面（全局函数）和 node 测试（module.exports）用。

const names = {
  1: "电压",
  2: "成组",
  3: "湿度",
  4: "气压",
  5: "风速",
  6: "云量",
  8: "计数",
  150: "温度",
};

function nameOf(id) {
  return names[id] || "未知";
}

function readU16(u8, o) {
  return [u8[o] | (u8[o + 1] << 8), o + 2];
}

function readU32(u8, o) {
  return [
    (u8[o] | (u8[o + 1] << 8) | (u8[o + 2] << 16) | (u8[o + 3] << 24)) >>> 0,
    o + 4,
  ];
}

// LEB128：低 7 位有效，最高位 1 表示还有后续字节，不越过 end。
function readVarint(u8, o, end) {
  let v = 0;
  let shift = 0;
  while (o < end) {
    const b = u8[o++];
    v += (b & 0x7f) * 2 ** shift;
    if ((b & 0x80) === 0) break;
    shift += 7;
  }
  return [v, o];
}

function readFields(u8, o, end) {
  const fields = [];
  while (o < end) {
    let key;
    [key, o] = readVarint(u8, o, end);
    if (o >= end) break;
    const kind = u8[o++];
    if (kind === 1) {
      if (o + 4 > end) break;
      let val;
      [val, o] = readU32(u8, o);
      fields.push({ name: nameOf(key), value: val });
    } else if (kind === 3) {
      let n;
      [n, o] = readVarint(u8, o, end);
      if (n > end - o) break;
      const inner = readFields(u8, o, o + n);
      o += n;
      fields.push({ name: nameOf(key), value: inner });
    } else {
      break;
    }
  }
  return fields;
}

function decode(u8) {
  if (u8.length < 2) return { version: 0, fields: [] };
  const [version, o] = readU16(u8, 0);
  return { version, fields: readFields(u8, o, u8.length) };
}

if (typeof module !== "undefined" && module.exports) {
  module.exports = { decode };
}
