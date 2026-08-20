#!/usr/bin/env python3
"""把框邊上那些標籤（`ESC`、`POOL MONEY`、`ROSTER ON`…）倒出來。

兩張表（`docs/re/126`）：

- `ds:CA70h` 起是**彩色字型的字模串**，`0x00` 結束。編碼與 `sub_17451` 一致：
  `索引 ＝ (字元 & 0xDF) − 0x29`，空白是 `0x33`，頭尾各有一個蓋子。
- `ds:CBBDh` 起是版面表，17 筆 × 6 bytes：`欄／列／長度／動作碼／字模串指標`。

⚠ **這一族字串在映像裡不是 ASCII**，`grep POOL` 一筆都不會中——
那個零命中與「這東西不存在」長得一模一樣。

用法（純 stdlib，不需要 IDA）：
    python3 tools/summarize_box_labels.py <wl.merged.exe> [輸出.md]
"""

from __future__ import annotations

import struct
import sys
from pathlib import Path

DS = 0x1CE20
LABEL_TABLE = 0xCBBD
LABEL_COUNT = 17
RECORD = 6
CAP_LEFT = (0x16, 0x17)
CAP_RIGHT = (0x28, 0x32)
SPACE = 0x33


def main() -> None:
    if len(sys.argv) not in (2, 3):
        sys.exit(__doc__)
    img = Path(sys.argv[1]).read_bytes()
    hdr = struct.unpack_from("<H", img, 8)[0] * 16

    def at(off: int) -> int:
        return DS + off - 0x10000 + hdr

    def text(ptr: int) -> str:
        out = []
        i = at(ptr)
        while img[i] != 0:
            g = img[i]
            if g in CAP_LEFT or g in CAP_RIGHT:
                out.append("|")  # 頭尾的蓋子
            elif g == SPACE:
                out.append(" ")
            else:
                out.append(chr(g + 0x29))
            i += 1
        return "".join(out)

    rows = [
        "# 框邊標籤（工具輸出，不含推論）",
        "",
        f"版面表 `ds:{LABEL_TABLE:04X}h`，{LABEL_COUNT} 筆 × {RECORD} bytes。",
        "`|` 是頭尾的蓋子字模。",
        "",
        "| # | 標籤 | 欄 | 列 | 長度 | 動作碼 | 字模串 |",
        "|---:|---|---:|---:|---:|---:|---|",
    ]
    for n in range(LABEL_COUNT):
        base = at(LABEL_TABLE + n * RECORD)
        col, row, length, action = img[base : base + 4]
        ptr = struct.unpack_from("<H", img, base + 4)[0]
        rows.append(
            f"| {n} | `{text(ptr)}` | {col} | {row} | {length} | "
            f"{action:#04x} | `{ptr:#06x}` |"
        )
    out = "\n".join(rows) + "\n"
    if len(sys.argv) == 3:
        Path(sys.argv[2]).write_text(out, encoding="utf-8")
        print(f"→ {sys.argv[2]}")
    else:
        print(out)


if __name__ == "__main__":
    main()
