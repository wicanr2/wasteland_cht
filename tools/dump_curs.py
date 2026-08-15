#!/usr/bin/env python3
"""把 `CURS` 的 2,048 bytes 拆成 8 個 32 × 16 的 EGA 4 平面圖形。

版面（`docs/re/57`）：

    8 個單位 × 256 bytes
    一個單位 ＝ 32 × 16、EGA 4 平面**平面連續**（與 `IC0_9.WLF` 同一套）
    平面大小 ＝ 4 bytes × 16 列 ＝ 64，一個單位 ＝ 4 × 64 ＝ 256

⚠ **平面是連續的，不是逐列交錯。** 用逐列交錯解出來每隔一列全空，
看起來像「圖案只有一半」——那是版面錯了，不是資料稀疏。
判斷方法是拿 `IC0_9.WLF` 用同一支公式解一次做正對照（`docs/re/57` §1）。

只輸出文字圖與統計，不倒出原版影像資料。

用法（純 stdlib）：
    python3 tools/dump_curs.py workplace/orig/wastland/curs [輸出.md]
"""

from __future__ import annotations

import sys
from pathlib import Path

W, H = 32, 16
ROW = W // 8          # 一列 4 bytes
PLANE = ROW * H       # 一個平面 64 bytes
UNIT = PLANE * 4      # 一個圖形 256 bytes
HEX = "0123456789abcdef"


def decode(raw: bytes, unit: int) -> list[list[int]]:
    """EGA 4 平面（平面連續）→ 索引圖。與 tools/dump_icons.py 同一套公式。"""
    base = unit * UNIT
    out = []
    for y in range(H):
        row = []
        for x in range(W):
            v = 0
            for p in range(4):
                b = raw[base + p * PLANE + y * ROW + (x >> 3)]
                v |= ((b >> (7 - (x & 7))) & 1) << p
            row.append(v)
        out.append(row)
    return out


def main() -> None:
    if len(sys.argv) not in (2, 3):
        sys.exit(__doc__)
    raw = Path(sys.argv[1]).read_bytes()
    if len(raw) % UNIT:
        sys.exit(f"{len(raw)} bytes 不是 {UNIT} 的倍數，版面大概不對")
    count = len(raw) // UNIT

    rows = [
        f"# `CURS` 的 {count} 個圖形（工具輸出，不含推論）",
        "",
        f"版面 {W} × {H}、EGA 4 平面平面連續，一個圖形 {UNIT} bytes。",
        "`.` 是 0，其餘是十六進位的索引值。",
        "",
    ]
    for u in range(count):
        img = decode(raw, u)
        used = sorted({v for r in img for v in r if v})
        last = max((y for y, r in enumerate(img) if any(r)), default=-1)
        rows += [
            f"## 圖形 {u}",
            "",
            f"用到的顏色：{' '.join(f'`{v:x}`' for v in used) or '無'}；"
            f"最後一列有內容的是第 {last} 列。",
            "",
            "```",
        ]
        rows += ["".join("." if v == 0 else HEX[v] for v in r) for r in img]
        rows += ["```", ""]

    text = "\n".join(rows) + "\n"
    if len(sys.argv) == 3:
        Path(sys.argv[2]).write_text(text, encoding="utf-8")
        print(f"→ {sys.argv[2]}")
    else:
        print(text)


if __name__ == "__main__":
    main()
