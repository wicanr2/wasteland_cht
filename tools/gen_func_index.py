"""產生函式索引：把 IDA 清冊與已寫的逆向筆記對照，標出誰分析過、誰還沒。

CLAUDE.md 要求「讀任何 sub_XXXXX 之前先查函式索引」——筆記多起來之後，
光靠記憶一定會重讀已經解過的函式。這支就是那份索引的產生器。

用法（容器內）：
    python3 tools/gen_func_index.py <inventory-merged.json> <docs 目錄> <輸出.md>
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

FUNC_RE = re.compile(r"\bsub_([0-9A-Fa-f]{4,6})\b")
ADDR_RE = re.compile(r"0x([0-9A-Fa-f]{4,6})\b")


def main() -> None:
    inv_p, docs_p, out_p = (Path(p) for p in sys.argv[1:4])
    inv = json.loads(inv_p.read_text(encoding="utf-8"))

    mentioned: dict[int, set[str]] = {}
    for md in sorted(docs_p.rglob("*.md")):
        # generated/ 底下是工具輸出的完整清冊，列了每一個函式，
        # 算進來會讓「已分析」失真。人寫的筆記才算數。
        if md.name.startswith("00-function-index") or "generated" in md.parts:
            continue
        text = md.read_text(encoding="utf-8", errors="replace")
        addrs = {int(m, 16) for m in FUNC_RE.findall(text)}
        addrs |= {int(m, 16) for m in ADDR_RE.findall(text)}
        rel = str(md.relative_to(docs_p.parent))
        for a in addrs:
            mentioned.setdefault(a, set()).add(rel)

    funcs = sorted(inv["functions"], key=lambda f: int(f["ida_address"], 16))
    rows = []
    for f in funcs:
        ea = int(f["ida_address"], 16)
        docs = sorted(mentioned.get(ea, ()))
        rows.append((ea, f, docs))

    analysed = [r for r in rows if r[2]]
    lines = [
        "# 00：函式索引\n",
        "> 由 `tools/gen_func_index.py` 產生。**讀任何 `sub_XXXXX` 之前先查這張表**——",
        "> 筆記超過二三十份之後，靠記憶一定會重讀已經解過的函式。\n",
        f"輸入：`{inv['input_identity']}`，SHA-256 `{inv['input_sha256']}`\n",
        f"- 自動辨識函式：**{len(funcs)}**",
        f"- 已在筆記中出現：**{len(analysed)}**",
        f"- 尚未碰過：**{len(funcs) - len(analysed)}**\n",
        "## 已分析（依呼叫端數量排序）\n",
    ]
    lines.append("| 位址 | segment:offset | 大小 | callers | 出現於 |")
    lines.append("|---|---|---:|---:|---|")
    for ea, f, docs in sorted(analysed, key=lambda r: -r[1]["caller_count"]):
        seg = f"{f['segment']}+{f['segment_offset']}"
        doc = "、".join(d.replace("docs/re/", "") for d in docs)
        lines.append(
            f"| `{f['ida_address']}` | {seg} | {f['size']} | {f['caller_count']} | {doc} |"
        )

    lines.append("\n## 尚未碰過的函式（依大小排序，前 60 個）\n")
    lines.append("大的通常是主邏輯，是後續分析的優先對象。\n")
    lines.append("| 位址 | segment:offset | 大小 | callers |")
    lines.append("|---|---|---:|---:|")
    rest = [r for r in rows if not r[2]]
    for ea, f, _ in sorted(rest, key=lambda r: -r[1]["size"])[:60]:
        seg = f"{f['segment']}+{f['segment_offset']}"
        lines.append(f"| `{f['ida_address']}` | {seg} | {f['size']} | {f['caller_count']} |")

    out_p.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"函式 {len(funcs)}，已分析 {len(analysed)}，未碰 {len(funcs) - len(analysed)}")


main()
