#!/usr/bin/env python3
"""把 `TRANSTBL` 的 800 bytes 拆成 50 組 × 16 的對照表。

`docs/re/56` 讀出它的形狀：每 16 bytes 一組、輸入索引 0–15 對到輸出值 0–15，
第 0 項恆為 0。這支只**拆結構並統計**，不判斷它是什麼表。

用法（純 stdlib）：
    python3 tools/summarize_transtbl.py workplace/orig/wastland/transtbl [輸出.md]
"""

from __future__ import annotations

import sys
from collections import Counter
from pathlib import Path

GROUP = 16


def main() -> None:
    if len(sys.argv) not in (2, 3):
        sys.exit(__doc__)
    data = Path(sys.argv[1]).read_bytes()
    if len(data) % GROUP:
        sys.exit(f"{len(data)} bytes 不是 {GROUP} 的倍數，拆法大概不對")
    groups = [data[i : i + GROUP] for i in range(0, len(data), GROUP)]

    zero = [i for i, g in enumerate(groups) if set(g) == {0}]
    ident = [i for i, g in enumerate(groups) if list(g) == list(range(GROUP))]
    uniq = Counter(bytes(g) for g in groups)

    rows = [
        "# `TRANSTBL` 的結構（工具輸出，不含推論）",
        "",
        f"{len(data)} bytes ＝ **{len(groups)} 組 × {GROUP}**。值域 "
        f"{min(data)}–{max(data)}，相異值 {len(set(data))} 種。",
        "",
        "| 組 | 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 8 | 9 | a | b | c | d | e | f |",
        "|---:|" + "---|" * GROUP,
    ]
    for i, g in enumerate(groups):
        rows.append(f"| {i} | " + " | ".join(f"{v:x}" for v in g) + " |")

    rows += [
        "",
        f"- 全 0 的組：{zero or '無'}（{len(zero)} 組）",
        f"- 恆等（i → i）的組：{ident or '無'}",
        f"- 相異的組：{len(uniq)} 種",
        f"- 第 0 項是 0 的組：{sum(1 for g in groups if g[0] == 0)}／{len(groups)}",
        f"- 第 15 項是 15 的組：{sum(1 for g in groups if g[GROUP - 1] == GROUP - 1)}"
        f"／{len(groups)}",
        "",
        "每一欄（輸入索引）在非全 0 的組裡出現過的輸出值：",
        "",
        "| 輸入 | 輸出值 |",
        "|---:|---|",
    ]
    for c in range(GROUP):
        vals = sorted({g[c] for g in groups if set(g) != {0}})
        rows.append(f"| {c:x} | " + " ".join(f"`{v:x}`" for v in vals) + " |")

    text = "\n".join(rows) + "\n"
    if len(sys.argv) == 3:
        Path(sys.argv[2]).write_text(text, encoding="utf-8")
        print(f"→ {sys.argv[2]}")
    else:
        print(text)


if __name__ == "__main__":
    main()
