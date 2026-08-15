#!/usr/bin/env python3
"""把「誰會被畫成 `IC0_9.WLF` 的十張疊圖」這件事的**資料面**倒出來。

程式面的四個呼叫點在 `docs/re/48`；這一支只回答資料問得到的部分：

1. 執行檔裡 `ds:AA17h` 那張**敵人種類 → 疊圖編號**的表（6 個有效項）。
2. 42 張地圖的第 3 層裡，值落在 0–9（＝ 這十張，而不是 `ALLHTDS` 圖磚）的格子。
3. 42 張地圖第 1 層 nibble 為 **4／5／9** 的格子數——那三種是 `sub_167CE`
   每走一步會重畫的「會動的格子」，也正是疊圖的三個來源。

⚠ 這支**只排序整理，不加語意**（`CLAUDE.md` §1.2）。「疊圖 6 是 Animal」
這種話要寫在 `docs/re/48`，不寫在這裡。

⚠ 第 1 層的 nibble 是 section 型別，**不是**疊圖編號；nibble 4 的格子畫哪一張
由該筆記錄的 `+0x01` 決定，而且**只有 < 10 才是疊圖**，≥ 10 是 `ALLHTDS` 圖磚
（`sub_18024` 的 `cmp al, 0Ah`）。這個門檻不套的話會憑空多出幾十個假的疊圖。

用法（純 stdlib，不需要 IDA）：
    python3 tools/summarize_icons.py <wl.merged.exe> <game1> <game2> [輸出.md]
"""

from __future__ import annotations

import importlib.util
import struct
import sys
from collections import Counter
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from huffman import decompress  # noqa: E402

DS = 0x1CE20
KIND_ICON_TABLE = 0xAA17  # sub_14664 的 `mov al, [bx-55E9h]`
KIND_ICON_LEN = 6  # 敵人種類 0–5（1–5 有意義，`docs/re/37` §3.2）
DIM_AT = 0x2C  # 記錄區標頭裡的邊長（docs/re/24 §1）
ICON_COUNT = 10  # IC0_9.WLF 的張數；也是 sub_18024 的 `cmp al, 0Ah`
MOVING_NIBBLES = (4, 5, 9)  # sub_167CE 每步重畫的三種


def load(exe: bytes, g1: bytes, g2: bytes):
    """借 summarize_monsters 的區塊解析，避免第二份會漂移的實作。"""
    here = Path(__file__).resolve().parent
    spec = importlib.util.spec_from_file_location("_mon", here / "summarize_monsters.py")
    mon = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mon)
    return mon.blocks_with_tail(exe, g1, g2)


def section_offsets(exe: bytes) -> list[int]:
    hdr = struct.unpack_from("<H", exe, 8)[0] * 16
    base = DS + 0xB9E0 - 0x10000 + hdr
    return [struct.unpack_from("<H", exe, base + i * 2)[0] for i in range(24)]


def u16(b: bytes, off: int | None) -> int | None:
    if off is None or off < 0 or off + 2 > len(b):
        return None
    return struct.unpack_from("<H", b, off)[0]


def main() -> None:
    if len(sys.argv) not in (4, 5):
        sys.exit(__doc__)
    exe, g1, g2 = (Path(p).read_bytes() for p in sys.argv[1:4])

    hdr = struct.unpack_from("<H", exe, 8)[0] * 16
    tbl_off = DS + KIND_ICON_TABLE - 0x10000 + hdr
    kind_icon = list(exe[tbl_off : tbl_off + KIND_ICON_LEN])

    rows = [
        "# 十張疊圖的資料面（工具輸出，不含推論）",
        "",
        "## 1. `ds:AA17h`：敵人種類 → 疊圖編號",
        "",
        f"線性 `0x{DS + KIND_ICON_TABLE:X}`，檔案位移 `0x{tbl_off:X}`，"
        f"原始 bytes `{exe[tbl_off:tbl_off + KIND_ICON_LEN].hex(' ')}`。",
        "",
        "| 敵人種類（記錄 `+0x06`） | 疊圖編號 |",
        "|---:|---:|",
    ]
    rows += [f"| {i} | {v} |" for i, v in enumerate(kind_icon)]

    offs = section_offsets(exe)
    layer3 = Counter()
    nibbles: dict[int, Counter] = {}
    n4_icon = Counter()
    n4_tile = 0
    n4_unresolved = 0
    blocks = 0

    tail_failed: list[int] = []
    for res_id, label, body, map_size, tail_raw in load(exe, g1, g2):
        dim = body[map_size + DIM_AT] if map_size + DIM_AT < len(body) else 0
        if not dim or dim * dim + dim * dim // 2 > len(body):
            continue
        blocks += 1
        l1 = body[: dim * dim // 2]
        l2 = body[dim * dim // 2 : dim * dim // 2 + dim * dim]
        # 第 3 層在 Huffman 尾段（docs/re/24 §2），**不在 body 裡**。
        try:
            tail = decompress(tail_raw, verify_magic=False)[0]
        except Exception:
            tail_failed.append(res_id)
            tail = b""

        c = Counter()
        cells4 = []
        for i, b in enumerate(l1):
            for half, n in ((0, b >> 4), (1, b & 0x0F)):
                if n in MOVING_NIBBLES:
                    c[n] += 1
                    if n == 4:
                        idx = i * 2 + half
                        cells4.append((idx % dim, idx // dim))
        if c:
            nibbles[res_id] = c

        for v in tail:
            if v < ICON_COUNT:
                layer3[v] += 1

        # nibble 4 的格子：記錄 +0x01 < 10 才是疊圖。
        start = u16(body, map_size + offs[4])
        first = u16(body, start) if start else None
        ok = bool(start and first and start < first <= len(body) and (first - start) % 2 == 0)
        for x, y in cells4:
            rec = l2[y * dim + x]
            p = u16(body, start + 2 * rec) if ok and rec < (first - start) // 2 else None
            if p is None or p + 1 >= len(body):
                n4_unresolved += 1
                continue
            v = body[p + 1] & 0x7F
            if v < ICON_COUNT:
                n4_icon[v] += 1
            else:
                n4_tile += 1

    rows += [
        "",
        "## 2. 第 3 層值落在 0–9 的格子（＝ 背景直接用這十張）",
        "",
        f"掃了 {blocks} 個有地圖的區塊。",
        "",
        "| 疊圖編號 | 格子數 |",
        "|---:|---:|",
    ]
    rows += [f"| {i} | {layer3.get(i, 0)} |" for i in range(ICON_COUNT)]
    rows += ["", f"合計 {sum(layer3.values())} 格。"]
    if tail_failed:
        rows.append(f"⚠ 尾段解壓失敗而沒算進去的區塊：{tail_failed}——**這不是零**。")

    rows += [
        "",
        "## 3. 會動的三種 nibble（`sub_167CE` 每走一步重畫）",
        "",
        "| 資源 | nibble 4 | nibble 5 | nibble 9 |",
        "|---:|---:|---:|---:|",
    ]
    for res_id in sorted(nibbles):
        c = nibbles[res_id]
        rows.append(f"| {res_id} | {c.get(4, 0)} | {c.get(5, 0)} | {c.get(9, 0)} |")
    tot = Counter()
    for c in nibbles.values():
        tot.update(c)
    rows.append(f"| **合計** | {tot.get(4, 0)} | {tot.get(5, 0)} | {tot.get(9, 0)} |")

    rows += [
        "",
        "### nibble 4 的格子：記錄 `+0x01` 指到什麼",
        "",
        f"畫成疊圖（`< {ICON_COUNT}`）：{sum(n4_icon.values())} 格 → "
        + ("、".join(f"編號 {k} × {v}" for k, v in sorted(n4_icon.items())) or "無"),
        "",
        f"畫成 `ALLHTDS` 圖磚（`>= {ICON_COUNT}`）：{n4_tile} 格。",
        "",
        f"記錄查不到而略過：{n4_unresolved} 格"
        + ("（**不是零，要在筆記裡交代**）" if n4_unresolved else ""),
    ]

    text = "\n".join(rows) + "\n"
    if len(sys.argv) == 5:
        Path(sys.argv[4]).write_text(text, encoding="utf-8")
        print(f"→ {sys.argv[4]}")
    else:
        print(text)


if __name__ == "__main__":
    main()
