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
