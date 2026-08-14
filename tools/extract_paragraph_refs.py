"""從原文字串裡抽出「哪一條敘述會叫玩家去讀第幾段」。

輸出 `docs/re/generated/paragraph-refs.tsv`：`<字串 key>\t<段落編號>`。

這份**只有 key 與數字、不含任何原文**，所以與其他衍生資料一樣入版控
（`translations/source/` 是原文，依 `CLAUDE.md` §7 不入版控）。

⚠ **一定要從英文原文抽，不能在執行期解析翻譯過的文字。**
翻譯之後那句話變成「請看第 23 段。」，格式隨譯者而變；把偵測放在執行期，
等於讓每個譯者的用字決定遊戲讀不讀得到段落。這張表是編譯期產物，
翻譯怎麼寫都不影響。

`docs/re/33` §2 的例外（`game2` 資源 15：前綴與數字是執行期拼出來的）
在這裡抽不到，會列在輸出的註解裡。

用法（純 stdlib）：
    python3 tools/extract_paragraph_refs.py [translations/source/blocks.tsv] [輸出]
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

# 原文的兩種寫法：整條就這一句，或混在句子中間。
REF = re.compile(r"[Rr]ead paragraph\s+(\d+)")

# 沒有編號的前綴（docs/re/33 §2）：編號由選單選擇在執行期接上去。
BARE = re.compile(r"[Rr]ead paragraph\s*$")


def main() -> None:
    src = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("translations/source/blocks.tsv")
    out = Path(sys.argv[2]) if len(sys.argv) > 2 else Path("docs/re/generated/paragraph-refs.tsv")

    rows: list[tuple[str, int]] = []
    runtime: list[str] = []
    for line in src.read_text(encoding="utf-8").splitlines():
        if line.startswith("#") or "\t" not in line:
            continue
        key, text = line.split("\t", 1)
        if m := REF.search(text):
            rows.append((key, int(m.group(1))))
        elif BARE.search(text.rstrip("\\x0D").rstrip()):
            runtime.append(key)

    numbers = sorted({n for _, n in rows})
    lines = [
        "# 字串 key → 段落編號（tools/extract_paragraph_refs.py 產生，不要手改）",
        f"# {len(rows)} 條引用、{len(numbers)} 個不同編號",
    ]
    if runtime:
        lines.append("# 編號在執行期才拼上去、這裡抽不到的（docs/re/33 §2）：")
        lines += [f"#   {k}" for k in runtime]
    lines += [f"{k}\t{n}" for k, n in sorted(rows)]
    out.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"{len(rows)} 條引用、{len(numbers)} 個不同編號 → {out}")
    if runtime:
        print(f"另有 {len(runtime)} 條的編號是執行期拼的，沒有列進表裡")


if __name__ == "__main__":
    main()
