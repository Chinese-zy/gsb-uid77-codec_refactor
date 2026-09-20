document.getElementById("sieve").addEventListener("change", (ev) => {
  const want = ev.target.value;
  for (const li of document.querySelectorAll("#list li")) {
    li.hidden = want !== "全部" && li.dataset.kind !== want;
  }
});

function dumpBytes(u8) {
  const lines = ["偏移   原始字节（每行 16 字节）", "----   ------------------------"];
  for (let row = 0; row < u8.length; row += 16) {
    const chunk = Array.from(u8.slice(row, row + 16));
    const hex = chunk.map((b) => b.toString(16).padStart(2, "0")).join(" ")
      .padEnd(47, " ");
    lines.push(
      row.toString(16).padStart(4, "0") + "   " + hex
    );
  }
  lines.push("", `共 ${u8.length} 字节`);
  return lines.join("\n");
}

function sameFrame(a, b) {
  return JSON.stringify(a) === JSON.stringify(b);
}

let pageFrame = null;
let apiFrame = null;

function checkParity() {
  if (pageFrame === null || apiFrame === null) return;
  const parity = document.getElementById("parity");
  if (sameFrame(pageFrame, apiFrame)) {
    parity.textContent = "页面与接口解析结果一致（字段顺序逐项相同）。";
    parity.className = "ok";
  } else {
    parity.textContent = "页面与接口解析结果不一致，按 SPEC.md 视为缺陷。";
    parity.className = "bad";
  }
}

fetch("sample.bin")
  .then((res) => res.arrayBuffer())
  .then((buf) => {
    const u8 = new Uint8Array(buf);
    document.getElementById("hex").textContent = dumpBytes(u8);
    pageFrame = Codec.decode(u8);
    document.getElementById("page").textContent = JSON.stringify(pageFrame, null, 2);
    checkParity();
  });

fetch("/api/frame")
  .then((res) => res.json())
  .then((frame) => {
    apiFrame = frame;
    document.getElementById("api").textContent = JSON.stringify(frame, null, 2);
    checkParity();
  });
