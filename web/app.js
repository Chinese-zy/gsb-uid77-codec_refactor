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

function readU16(u8, o) {
  return [u8[o] | (u8[o + 1] << 8), o + 2];
}

function readU32(u8, o) {
  return [
    (u8[o] | (u8[o + 1] << 8) | (u8[o + 2] << 16) | (u8[o + 3] << 24)) >>> 0,
    o + 4,
  ];
}

function readVarint(u8, o) {
  let n = 0;
  let shift = 0;
  while (o < u8.length) {
    const b = u8[o++];
    n |= (b & 0x7f) << shift;
    if ((b & 0x80) === 0) break;
    shift += 7;
  }
  return [n, o];
}

function readFields(u8, o, end) {
  const fields = [];
  while (o < end) {
    let key;
    [key, o] = readVarint(u8, o);
    if (o >= end) break;
    const kind = u8[o++];
    if (kind === 1) {
      if (o + 4 > end) break;
      let val;
      [val, o] = readU32(u8, o);
      fields.push({ name: names[key] || "未知", value: val });
    } else if (kind === 3) {
      let n;
      [n, o] = readVarint(u8, o);
      const innerEnd = Math.min(end, o + n);
      const inner = readFields(u8, o, innerEnd);
      o = innerEnd;
      fields.push({ name: names[key] || "未知", value: inner });
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

document.getElementById("sieve").addEventListener("change", (ev) => {
  const want = ev.target.value;
  for (const li of document.querySelectorAll("#list li")) {
    li.hidden = want !== "全部" && li.dataset.kind !== want;
  }
});

fetch("sample.bin")
  .then((res) => res.arrayBuffer())
  .then((buf) => {
    const u8 = new Uint8Array(buf);
    const hex = Array.from(u8).map((b) => b.toString(16).padStart(2, "0")).join(" ");
    document.getElementById("hex").textContent = hex;
    document.getElementById("page").textContent = JSON.stringify(decode(u8), null, 2);
  });

fetch("/api/frame")
  .then((res) => res.text())
  .then((text) => {
    document.getElementById("api").textContent = text;
  });
