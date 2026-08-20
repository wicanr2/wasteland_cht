#!/usr/bin/env python3
"""把 42 個區塊的條件閘（nibble 2）獎懲參數倒出來。

參數在記錄 `+0x08`／`+0x09`（`docs/re/67` §1）：

| 欄位 | 意思 |
|---|---|
| `+0x08` 低 7 位 | 角色記錄的欄位位移（`0x1D` ＝ CON、`0x15` ＝ 金錢、`0x21` ＝ 經驗）|
| `+0x08` bit7 | 1 ＝ 固定值；0 ＝ 擲 (`+0x09` & 0x7F) 顆 d6 |
| `+0x09` 低 7 位 | 量或骰數 |
| `+0x09` bit7 | 1 ＝ 減、0 ＝ 加 |

**外加一欄 `+0x00` 的 bit0**：扣 CON 時它會被抄進 `ds:46EFh`
（`0x1423C`），非 0 ＝ **這一次跳過護甲吸收**（`docs/re/55` §1）。
那一位決定「穿好一點有沒有用」，所以在這裡當成一等公民印出來。

⚠ 懲罰落在**沒通過檢定的人**身上，逐個角色各自判定（`docs/re/67` §3）——
不是整隊一起。

用法（純 stdlib，不需要 IDA）：
    python3 tools/summarize_gate_penalties.py <wl.merged.exe> <game1> <game2> [輸出.md]
"""

from __future__ import annotations

import importlib.util
import sys
from pathlib import Path

HERE = Path(__file__).resolve().parent
GATE_TYPE = 2
FIELD_AT, AMOUNT_AT = 0x08, 0x09
FIELDS = {0x1D: "CON", 0x15: "金錢", 0x21: "經驗"}


def _scan():
    spec = importlib.util.spec_from_file_location("_s", HERE / "scan_item_refs.py")
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def main() -> None:
    if len(sys.argv) not in (4, 5):
        sys.exit(__doc__)
    S = _scan()
    exe, g1, g2 = (Path(p).read_bytes() for p in sys.argv[1:4])
    offs = S.section_offsets(exe)

    rows = [
        "| 資源 | 檔案 | 記錄 | 格數 | 欄位 | 加減 | 量 | 護甲 |",
        "|---:|---|---:|---:|---|---|---|---|",
    ]
    con_total = con_bypass = 0
    aliased = 0
    for res_id, label, body, map_size in S.load(exe, g1, g2):
        dim = body[map_size + S.DIM_AT] if map_size + S.DIM_AT < len(body) else 0
        if not dim or dim * dim + dim * dim // 2 > len(body):
            continue
        layer2 = body[dim * dim // 2 : dim * dim // 2 + dim * dim]
        start, count = S.array(body, map_size + offs[GATE_TYPE])
        if not start:
            continue
        if any(
            t != GATE_TYPE and S.u16(body, map_size + o) == start
            for t, o in enumerate(offs)
        ):
            aliased += 1
            continue
        cells: dict[int, int] = {}
        for x, y in S.cells_of(body, dim, GATE_TYPE):
            cells[layer2[y * dim + x]] = cells.get(layer2[y * dim + x], 0) + 1
        for i in range(count):
            p = S.u16(body, start + 2 * i)
            if not p or p + AMOUNT_AT >= len(body):
                continue
            f, q = body[p + FIELD_AT], body[p + AMOUNT_AT]
            if f == 0:
                continue
            field = f & 0x7F
            amount = (
                f"{q & 0x7F}" if f & 0x80 else f"{q & 0x7F} 顆 d6"
            )
            armour = "—"
            if field == 0x1D:
                con_total += 1
                if body[p] & 1:
                    con_bypass += 1
                    armour = "**跳過**"
                else:
                    armour = "照扣"
            rows.append(
                f"| {res_id} | {label} | {i} | {cells.get(i, 0)} | "
                f"{FIELDS.get(field, hex(field))} | {'減' if q & 0x80 else '加'} | "
                f"{amount} | {armour} |"
            )

    head = [
        "# 條件閘的獎懲參數（工具輸出，不含推論）",
        "",
        f"有懲罰的記錄 **{len(rows) - 2}** 筆，其中扣 CON 的 **{con_total}** 筆："
        f"**{con_bypass}** 筆跳過護甲吸收、**{con_total - con_bypass}** 筆照扣。",
        f"section 2 的指標與別的型別撞在一起而跳過的區塊：**{aliased}**。",
        "",
        "「格數」是出貨地圖有幾格指到這一筆；0 ＝ 要靠改寫才到得了。",
        "",
    ]
    out = "\n".join(head + rows) + "\n"
    if len(sys.argv) == 5:
        Path(sys.argv[4]).write_text(out, encoding="utf-8")
        print(f"→ {sys.argv[4]}")
    else:
        print(out)


if __name__ == "__main__":
    main()
