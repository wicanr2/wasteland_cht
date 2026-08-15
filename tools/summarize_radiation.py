#!/usr/bin/env python3
"""把 42 張地圖裡 nibble 9（輻射區）格子指到的記錄倒出來。

`docs/re/55` 讀出輻射結算會用記錄的兩個 byte：

    +0x00 的 bit0 → ds:46EFh（這一次結算跳不跳過護甲吸收）
    +0x01         → 擲幾顆 d6 當傷害

這支只**列出實際資料**，不下語意判斷——「輻射傷害會不會被護甲擋」這種問題
要由資料回答，不是由讀程式碼的人猜。

用法（純 stdlib）：
    python3 tools/summarize_radiation.py <wl.merged.exe> <game1> <game2> [輸出.md]
"""

from __future__ import annotations

import importlib.util
import struct
import sys
from collections import Counter
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

DS = 0x1CE20
DIM_AT = 0x2C
RAD_NIBBLE = 9


def _icons():
    here = Path(__file__).resolve().parent
    spec = importlib.util.spec_from_file_location("_icons", here / "summarize_icons.py")
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def main() -> None:
    if len(sys.argv) not in (4, 5):
        sys.exit(__doc__)
    ic = _icons()
    exe = Path(sys.argv[1]).read_bytes()
    g1 = Path(sys.argv[2]).read_bytes()
    g2 = Path(sys.argv[3]).read_bytes()
    offs = ic.section_offsets(exe)

    rows = [
        "# nibble 9（輻射區）的記錄（工具輸出，不含推論）",
        "",
        "`+0x00` 是記錄第一個 byte，`+0x01` 是第二個。欄位語意見 `docs/re/55`。",
        "",
        "| 資源 | 格數 | 記錄數 | `+0x00` 的值 | `+0x00` bit0 | `+0x01` 的值 |",
        "|---:|---:|---:|---|---|---|",
    ]
    total_cells = 0
    all_b0, all_bit0, all_b1 = Counter(), Counter(), Counter()

    for res_id, _label, body, map_size, _tail in ic.load(exe, g1, g2):
        dim = body[map_size + DIM_AT] if map_size + DIM_AT < len(body) else 0
        if not dim or dim * dim + dim * dim // 2 > len(body):
            continue
        l1 = body[: dim * dim // 2]
        l2 = body[dim * dim // 2 : dim * dim // 2 + dim * dim]

        cells = []
        for i, b in enumerate(l1):
            for half, n in ((0, b >> 4), (1, b & 0x0F)):
                if n == RAD_NIBBLE:
                    idx = i * 2 + half
                    cells.append((idx % dim, idx // dim))
        if not cells:
            continue

        start = ic.u16(body, map_size + offs[RAD_NIBBLE])
        first = ic.u16(body, start) if start else None
        ok = bool(start and first and start < first <= len(body) and (first - start) % 2 == 0)
        count = (first - start) // 2 if ok else 0

        b0, bit0, b1 = Counter(), Counter(), Counter()
        recs = set()
        for x, y in cells:
            rec = l2[y * dim + x]
            recs.add(rec)
            p = ic.u16(body, start + 2 * rec) if ok and rec < count else None
            if p is None or p + 1 >= len(body):
                b0["查不到"] += 1
                continue
            b0[body[p]] += 1
            bit0[body[p] & 1] += 1
            b1[body[p + 1]] += 1
        total_cells += len(cells)
        all_b0.update(b0)
        all_bit0.update(bit0)
        all_b1.update(b1)
        rows.append(
            f"| {res_id} | {len(cells)} | {len(recs)} | {fmt(b0)} | {fmt(bit0)} | {fmt(b1)} |"
        )

    rows += [
        "",
        f"合計 **{total_cells} 格**。",
        "",
        f"- `+0x00` 的值：{fmt(all_b0)}",
        f"- `+0x00` 的 bit0：{fmt(all_bit0)}",
        f"- `+0x01` 的值：{fmt(all_b1)}",
    ]
    text = "\n".join(rows) + "\n"
    if len(sys.argv) == 5:
        Path(sys.argv[4]).write_text(text, encoding="utf-8")
        print(f"→ {sys.argv[4]}")
    else:
        print(text)


def fmt(c: Counter) -> str:
    return "、".join(f"`{k}`×{v}" for k, v in sorted(c.items(), key=lambda kv: str(kv[0])))


if __name__ == "__main__":
    main()
