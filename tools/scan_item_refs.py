#!/usr/bin/env python3
"""掃全部遊戲資料，找出每個物品編號被誰引用。

回答的是「某個物品在遊戲裡拿得到嗎」。三個來源：

| 來源 | 位置 | 怎麼認 |
|---|---|---|
| 寶箱 | nibble 5 的格 → section 5 記錄 `+0x02` 起每 2 bytes | bit7 設 ＝ 已擲定的物品編號 |
| 腳本攜帶檢查 | nibble 6 → section 6 記錄，opcode 43 的 `+0x07` 起 | `0xFF` 結束 |
| 商店 | 物品表 `+0x02` ＝ 庫存量（`docs/re/42`） | 非 0 ＝ 某家店有賣 |

⚠ **這支報「0 次引用」之前會先自己做正對照**（`~/diagnosis-notes/docs/02`）。
兩個對照量：腳本 opcode 必須全部落在 0–43、寶箱的未擲定類別必須全部落在 0–18。
任何一個超出，就代表走表走歪了——這時候的「0 次」是假零，工具會拒絕下結論。

⚠ **記錄 `+0x00` 不是 opcode**，是 section 型別 `0x10` 那張表的索引，
表裡存的才是 opcode（`docs/re/34` §1）。直接把 `+0x00` 當 opcode 讀，
分布會變成 `{0, 1, 2}` 而且 op 43 一筆都找不到——**看起來像「沒有人用」**。

用法（純 stdlib，不需要 IDA）：
    python3 tools/scan_item_refs.py <wl.merged.exe> <game1> <game2> [輸出.md]
"""

from __future__ import annotations

import importlib.util
import struct
import sys
from collections import Counter
from pathlib import Path

DS = 0x1CE20
SECTION_TABLE = 0xB9E0
SECTION_TYPES = 24
DIM_AT = 0x2C
CHEST_TYPE = 5
SCRIPT_TYPE = 6
OPCODE_TABLE_TYPE = 0x10  # 記錄 +0x00 → opcode 的對照表
OP_CARRY_CHECK = 43
OP_MAX = 43
CLASS_MAX = 18
CHEST_ITEMS_AT = 0x02
CARRY_LIST_AT = 0x07
DONE_MARK = 0x5E  # 寶箱處理完會被寫成 0xDE（docs/re/29 §4）


def load(exe: bytes, g1: bytes, g2: bytes):
    here = Path(__file__).resolve().parent
    spec = importlib.util.spec_from_file_location("_mon", here / "summarize_monsters.py")
    mon = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mon)
    return mon.blocks(exe, g1, g2)


def section_offsets(exe: bytes) -> list[int]:
    hdr = struct.unpack_from("<H", exe, 8)[0] * 16
    base = DS + SECTION_TABLE - 0x10000 + hdr
    return [struct.unpack_from("<H", exe, base + i * 2)[0] for i in range(SECTION_TYPES)]


def u16(b: bytes, off: int | None) -> int | None:
    if off is None or off < 0 or off + 2 > len(b):
        return None
    return struct.unpack_from("<H", b, off)[0]


def array(body: bytes, at: int | None):
    """指標陣列：第一項就是第一筆記錄的位置，所以項數 ＝ (第一項 − 起點) ÷ 2。"""
    start = u16(body, at)
    first = u16(body, start) if start else None
    if not start or not first or not (start < first <= len(body)) or (first - start) % 2:
        return None, 0
    return start, (first - start) // 2


def cells_of(body: bytes, dim: int, nibble: int):
    for i, b in enumerate(body[: dim * dim // 2]):
        for half, n in ((0, b >> 4), (1, b & 0x0F)):
            if n == nibble:
                idx = i * 2 + half
                yield idx % dim, idx // dim


def main() -> None:
    if len(sys.argv) not in (4, 5):
        sys.exit(__doc__)
    exe, g1, g2 = (Path(p).read_bytes() for p in sys.argv[1:4])
    offs = section_offsets(exe)

    chest_items = Counter()
    carry_items = Counter()
    ops = Counter()
    chest_cells = script_recs = 0
    no_script = out_of_table = 0
    bad_class = Counter()
    bad_op = Counter()

    for _res, _label, body, map_size in load(exe, g1, g2):
        dim = body[map_size + DIM_AT] if map_size + DIM_AT < len(body) else 0
        if not dim or dim * dim + dim * dim // 2 > len(body):
            continue
        layer2 = body[dim * dim // 2 : dim * dim // 2 + dim * dim]

        # 寶箱
        start, count = array(body, map_size + offs[CHEST_TYPE])
        for x, y in cells_of(body, dim, CHEST_TYPE):
            rec = layer2[y * dim + x]
            p = u16(body, start + 2 * rec) if start and rec < count else None
            if p is None or p + CHEST_ITEMS_AT >= len(body):
                continue
            chest_cells += 1
            k = p + CHEST_ITEMS_AT
            while k + 1 < len(body) and body[k] not in (0x00, 0xFF):
                v = body[k]
                if v & 0x80:
                    chest_items[v & 0x7F] += 1
                elif v > CLASS_MAX and v != DONE_MARK:
                    bad_class[v] += 1
                k += 2

        # 腳本。⚠ 沒有 nibble 6 的格子就沒有腳本——這種區塊的 section 6
        # 指標常常與別的 section 指到同一處（＝ 這個 section 不存在），
        # 照樣走下去會把別人的資料讀成 opcode。
        if not any(True for _ in cells_of(body, dim, SCRIPT_TYPE)):
            no_script += 1
            continue
        start, count = array(body, map_size + offs[SCRIPT_TYPE])
        opbase = u16(body, map_size + offs[OPCODE_TABLE_TYPE])
        if not start or not opbase:
            no_script += 1
            continue
        # 0x10 表裡存的是 opcode，所以它的有效長度 ＝ 開頭連續 ≤ 43 的那一段。
        table_len = 0
        while True:
            v = u16(body, opbase + 2 * table_len)
            if v is None or v > OP_MAX:
                break
            table_len += 1
        if table_len == 0:
            no_script += 1
            continue
        for i in range(count):
            p = u16(body, start + 2 * i)
            if p is None or p >= len(body):
                continue
            sel = body[p]
            if sel & 0x80:
                continue  # bit7 ＝ 設施畫面不是腳本（docs/spec/09 §2）
            if sel >= table_len:
                out_of_table += 1
                continue
            op = u16(body, opbase + 2 * sel)
            if op is None:
                continue
            script_recs += 1
            if op > OP_MAX:
                bad_op[op] += 1
                continue
            ops[op] += 1
            if op == OP_CARRY_CHECK:
                k = p + CARRY_LIST_AT
                while k < len(body) and body[k] != 0xFF:
                    carry_items[body[k]] += 1
                    k += 1

    trustworthy = not bad_class and not bad_op
    rows = [
        "# 物品編號被誰引用（工具輸出，不含推論）",
        "",
        f"寶箱格 **{chest_cells}** 個、腳本記錄 **{script_recs}** 筆。",
        "",
        "## 正對照（決定下面的「0 次」可不可信）",
        "",
        f"- 腳本 opcode 超出 0–{OP_MAX} 的：**{sum(bad_op.values())}**"
        + (f"（{dict(bad_op)}）" if bad_op else ""),
        f"- 寶箱未擲定類別超出 0–{CLASS_MAX} 的：**{sum(bad_class.values())}**"
        + (f"（{dict(bad_class)}）" if bad_class else ""),
        f"- 沒有 nibble 6 的格子而跳過的區塊：**{no_script}**（那些區塊沒有腳本）",
        f"- `+0x00` 指到 opcode 表外而跳過的記錄：**{out_of_table}**",
        "",
        (
            "兩個都是 0，代表兩張表都走對了，**下面的「0 次」才是真的沒有人用**。"
            if trustworthy
            else "⚠ **有超出值域的項，代表走表走歪了。下面的次數不可信，不要拿去下「沒有人用」的結論。**"
        ),
        "",
        "## 腳本 opcode 分布",
        "",
        "、".join(f"{k}×{v}" for k, v in sorted(ops.items())) or "（一筆都沒有）",
        "",
        "## 每個物品編號的引用次數",
        "",
        "| 物品 | 寶箱 | 腳本攜帶檢查 |",
        "|---:|---:|---:|",
    ]
    for n in sorted(set(chest_items) | set(carry_items)):
        rows.append(f"| {n} | {chest_items.get(n, 0)} | {carry_items.get(n, 0)} |")

    everyone = set(range(95))
    unused = sorted(everyone - set(chest_items) - set(carry_items))
    rows += [
        "",
        f"## 兩個來源都沒出現的編號（{len(unused)} 個）",
        "",
        "、".join(str(n) for n in unused) or "（沒有）",
        "",
        "⚠ 這不等於「遊戲裡拿不到」——起始裝備（`sub_1C9DE`）與商店庫存"
        "（物品表 `+0x02`）是另外兩條路，不在這份掃描裡。",
    ]

    text = "\n".join(rows) + "\n"
    if len(sys.argv) == 5:
        Path(sys.argv[4]).write_text(text, encoding="utf-8")
        print(f"→ {sys.argv[4]}")
    else:
        print(text)


if __name__ == "__main__":
    main()
