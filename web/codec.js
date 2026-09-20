// 与 frame/codec.go 共享同一份线格式，见仓库根目录 SPEC.md。
// 小端多字节整数 + LEB128 变长整数；字段严格按字节出现顺序保留，重复不丢。
(function (root, factory) {
  const codec = factory();
  if (typeof module === "object" && module.exports) {
    module.exports = codec;
  }
  root.Codec = codec;
})(typeof globalThis !== "undefined" ? globalThis : this, function () {
  "use strict";

  const KIND_SCALAR = 1;
  const KIND_GROUP = 3;

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
    return Object.prototype.hasOwnProperty.call(names, id) ? names[id] : "未知";
  }

  function readVarint(u8, start, end) {
    let v = 0;
    let o = start;
    for (let i = 0; ; i++) {
      if (o >= end || i >= 10) return null;
      const c = u8[o++];
      if (i === 9 && (c & 0x7f) > 0x01) return null;
      v += (c & 0x7f) * 2 ** (7 * i);
      if ((c & 0x80) === 0) return [v, o];
    }
  }

  function decodeFields(u8, start, end) {
    const fields = [];
    let o = start;
    while (o < end) {
      const keyRead = readVarint(u8, o, end);
      if (keyRead === null) return { fields, ok: false };
      const [id, afterId] = keyRead;
      if (afterId >= end) return { fields, ok: false };
      o = afterId;
      const kind = u8[o++];
      if (kind === KIND_SCALAR) {
        if (o + 4 > end) return { fields, ok: false };
        const value =
          (u8[o] | (u8[o + 1] << 8) | (u8[o + 2] << 16) | (u8[o + 3] << 24)) >>> 0;
        o += 4;
        fields.push({ id, name: nameOf(id), value });
      } else if (kind === KIND_GROUP) {
        const lenRead = readVarint(u8, o, end);
        if (lenRead === null) return { fields, ok: false };
        const [byteLen, afterLen] = lenRead;
        o = afterLen;
        const innerEnd = o + byteLen;
        if (innerEnd > end) return { fields, ok: false };
        const inner = decodeFields(u8, o, innerEnd);
        if (!inner.ok) return { fields, ok: false };
        fields.push({ id, name: nameOf(id), value: inner.fields });
        o = innerEnd;
      } else {
        return { fields, ok: false };
      }
    }
    return { fields, ok: true };
  }

  function decode(u8) {
    if (u8.length < 2) return { version: 0, fields: [] };
    const version = u8[0] | (u8[1] << 8);
    return { version, fields: decodeFields(u8, 2, u8.length).fields };
  }

  function appendVarint(out, v) {
    while (v >= 0x80) {
      out.push((v & 0x7f) | 0x80);
      v = Math.floor(v / 128);
    }
    out.push(v);
    return out;
  }

  function encodeFields(out, fields) {
    for (const field of fields) {
      appendVarint(out, field.id);
      if (Array.isArray(field.value)) {
        const inner = encodeFields([], field.value);
        out.push(KIND_GROUP);
        appendVarint(out, inner.length);
        for (const b of inner) out.push(b);
      } else if (
        Number.isInteger(field.value) &&
        field.value >= 0 &&
        field.value <= 0xffffffff
      ) {
        out.push(KIND_SCALAR);
        out.push(field.value & 0xff, (field.value >>> 8) & 0xff,
          (field.value >>> 16) & 0xff, field.value >>> 24);
      } else {
        throw new Error("codec: unsupported field value");
      }
    }
    return out;
  }

  function encode(frame) {
    const version = frame.version >>> 0;
    return Uint8Array.from(encodeFields(
      [version & 0xff, (version >>> 8) & 0xff],
      frame.fields || []
    ));
  }

  return { names, nameOf, decode, encode };
});
