#!/usr/bin/env python3
"""把 42 個 MSQ 區塊的 nibble 5（寶箱）逐格倒出來。

回答的是「**哪一格有什麼**」。記錄 `+0x02` 起每兩個 byte 一組（`docs/re/29` §4）：

| 第一個 byte | 意思 |
|---|---|
| `0x00` | 清單結束 |
| `0xFF` | 結束並跳到隊伍名單畫面 |
| `0x5E` | 特例：改寫成 `0xDE`，後面兩個 byte 各擲一次骰 |
| `< 0x80` | **物品類別**，第一次踩到才擲出是哪一件（`sub_15453` ＋ `sub_18E41`）|
| `≥ 0x80` | 出貨就決定好的物品編號（bit7 ＝ 已決定）|

第二個 byte 是數量；bit7 設著的話**數量也擲一次**。

⚠ 三個會製造假象的地方：

1. **指標表的第 0 項可以是 0**（資源 40 就是），拿它推表長會讓整張表消失。
   這裡走 `scan_item_refs.array()`，那一支已經改成找第一個非 0 的項。
2. **出貨地圖沒有 nibble 5 的格子不代表這張圖沒有寶箱**：腳本的 op 8／9／34／37
   會往 section 5 寫東西，問答與批次改寫也會把某一格變成 nibble 5。
   所以「格」欄空白的記錄照樣列出來。
3. **`0x5E` 不是類別 94**。它是特例標記，當成類別讀會多出一批不存在的「類 94」。

用法（純 stdlib，不需要 IDA）：
    python3 tools/summarize_chests.py <wl.merged.exe> <game1> <game2> [輸出.md]
"""

from __future__ import annotations

import importlib.util
import re
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
CHEST_TYPE = 5
ITEMS_AT = 0x02
DONE_MARK = 0x5E
CLASS_MAX = 18


def _scan():
    spec = importlib.util.spec_from_file_location("_s", HERE / "scan_item_refs.py")
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def item_names() -> dict[int, str]:
    """物品名字取自 `docs/re/generated/ida94/items.md`（`tools/dump_items.py` 的產物）。

    沒有那份就只印編號——**不要在這裡另外解一次物品表**，
    兩份實作會漂移。
    """
    p = HERE.parent / "docs/re/generated/ida94/items.md"
    if not p.exists():
        return {}
    out = {}
    for line in p.read_text(encoding="utf-8").splitlines():
        m = re.match(r"\|\s*(\d+)\s*\|\s*([^|]+?)\s*\|", line)
        if m:
            out[int(m.group(1))] = m.group(2)
    return out


def main() -> None:
    if len(sys.argv) not in (4, 5):
        sys.exit(__doc__)
    S = _scan()
    exe, g1, g2 = (Path(p).read_bytes() for p in sys.argv[1:4])
    offs = S.section_offsets(exe)
    names = item_names()

    def say(v: int) -> str:
        if v == DONE_MARK:
            return "〔擲骰特例〕"
        if v & 0x80:
            n = v & 0x7F
            return f"{n} {names.get(n, '')}".strip()
        return f"類別 {v}" + (f"（超出 0–{CLASS_MAX}）" if v > CLASS_MAX else "")

    rows = [
        "# 寶箱逐格內容（工具輸出，不含推論）",
        "",
        "`已定` ＝ 出貨資料就寫死的物品編號；`類別` ＝ 第一次踩到才擲。",
        "「格」空白 ＝ 出貨地圖沒有格子指到這一筆（要靠改寫才到得了）。",
        "",
        "| 資源 | 檔案 | 記錄 | 格 | 內容 |",
        "|---:|---|---:|---|---|",
    ]
    cells = recs = aliased = 0
    bad = []
    for res_id, label, body, map_size in S.load(exe, g1, g2):
        dim = body[map_size + S.DIM_AT] if map_size + S.DIM_AT < len(body) else 0
        if not dim or dim * dim + dim * dim // 2 > len(body):
            continue
        layer2 = body[dim * dim // 2 : dim * dim // 2 + dim * dim]
        start, count = S.array(body, map_size + offs[CHEST_TYPE])
        if not start:
            continue
        # section 5 的指標與別的型別撞在一起 ＝ **這張圖沒有寶箱**，
        # 指到的是別人的資料（世界地圖 0 與 26 就是這樣）。
        if any(
            t != CHEST_TYPE and S.u16(body, map_size + o) == start
            for t, o in enumerate(offs)
        ):
            aliased += 1
            continue
        where: dict[int, list[str]] = {}
        for x, y in S.cells_of(body, dim, CHEST_TYPE):
            where.setdefault(layer2[y * dim + x], []).append(f"({x},{y})")
            cells += 1
        # ⚠ 記錄的長度要用**下一筆記錄的位址**夾住。清單雖然以 `0x00`／`0xFF` 結束，
        # 但指標表的項數是推出來的（`array()`），推多了就會有幾筆指到別的 section
        # ——那些「記錄」讀下去會一路吃進字串，症狀是「類別」欄冒出 ASCII 碼。
        ptrs = sorted({q for q in (S.u16(body, start + 2 * j) for j in range(count)) if q})
        # section 的尾巴同樣要夾：最後一筆記錄後面接的是**別的 section**，
        # 不夾住就會把鄰居的資料讀成物品。
        others = [
            q
            for q in (S.u16(body, map_size + o) for o in offs)
            if q and q > start
        ]
        section_end = min(others) if others else len(body)
        for i in range(count):
            p = S.u16(body, start + 2 * i)
            if not p or p + ITEMS_AT >= len(body):
                continue
            after = [q for q in ptrs if q > p]
            limit = min(after[0], section_end) if after else section_end
            k, items, off_range = p + ITEMS_AT, [], False
            while k + 1 < limit and body[k] not in (0x00, 0xFF):
                v = body[k]
                if v != DONE_MARK and not v & 0x80 and v > CLASS_MAX:
                    off_range = True
                items.append(say(v))
                k += 2
            if not items:
                continue
            # 逐筆的正對照：這一筆走出值域，就**不報它的內容**。
            # 全域的「0 次」結論由下面的 unreadable 數字決定可不可信。
            if off_range:
                bad.append((res_id, i))
                continue
            recs += 1
            rows.append(
                f"| {res_id} | {label} | {i} | {'、'.join(where.get(i, []))} | "
                f"{'、'.join(items)} |"
            )

    rows[3:3] = [
        f"寶箱格 **{cells}** 個、讀得準的記錄 **{recs}** 筆。"
        f"section 5 的指標與別的型別撞在一起而跳過的區塊：**{aliased}**"
        "（那些圖沒有寶箱）。",
        "",
        f"走出類別值域而**不報內容**的記錄：**{len(bad)}**"
        + (f"（{bad}）" if bad else "")
        + "。",
        "",
        (
            "沒有走歪的記錄，所以「某個物品沒出現」可以當成結論。"
            if not bad
            else "⚠ **有記錄走歪了。下面少了那幾筆，"
            "所以「某個物品沒出現」這種結論要先排除它們。**"
        ),
    ]
    out = "\n".join(rows) + "\n"
    if len(sys.argv) == 5:
        Path(sys.argv[4]).write_text(out, encoding="utf-8")
        print(f"→ {sys.argv[4]}")
    else:
        print(out)


if __name__ == "__main__":
    main()
