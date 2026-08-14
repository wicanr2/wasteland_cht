"""從逐指令 JSON 掃出「以某個基址暫存器存取的記錄欄位」。

原版的記錄結構（角色、地圖記錄、隊伍）都不是用固定位址存取的，而是
`mov di, ds:46B5h` ＋ `mov bl, <常數>` ＋ `mov al, [bx+di]`。
固定位移掃描（`export_memops.py`）看不到這種存取，所以要另外做一次
極小範圍的資料流：在一個函式內線性追蹤 `bl`／`bx` 的常數值與 `inc`／`dec`，
碰到 `[bx+di]` 就把當下的 `bl` 記成欄位位移。

刻意保守：只認「常數載入 ＋ 加減一」這條鏈，`bl` 一旦被算出來（`shl`、
從記憶體讀）就放棄追蹤，寧可少報也不要報錯的位移。少報的部分會出現在
`unresolved` 計數裡，不會靜悄悄消失。

用法（純 stdlib，不需要 IDA）：
    python3 tools/summarize_record_fields.py <listing.json> [基址...]

    基址預設為角色記錄與地圖記錄那幾個：46b5 46b7 46c8 4661 46ae 46c6 4665
"""

from __future__ import annotations

import json
import re
import sys
from pathlib import Path

MOV_CONST = re.compile(r"^mov\s+(bl|bx),\s*([0-9A-Fa-f]+h|\d+)$")
MOV_BASE = re.compile(r"^mov\s+di,\s*ds:([0-9A-Fa-f]+)h$")
INCDEC = re.compile(r"^(inc|dec)\s+bl$")
CLOBBER = re.compile(r"^(mov|add|sub|shl|shr|and|or|xor|pop|lodsb)\b.*\bb[lx]\b")

DEFAULT_BASES = ["46b5", "46b7", "46c8", "4661", "46ae", "46c6", "4665"]


def parse_const(text: str) -> int:
    return int(text[:-1], 16) if text.lower().endswith("h") else int(text)


def main() -> None:
    if len(sys.argv) < 2:
        raise SystemExit(__doc__)
    listing = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
    bases = {b.lower() for b in (sys.argv[2:] or DEFAULT_BASES)}

    functions = listing["functions"]

    # IDA 沒建成函式的程式碼佔全檔約兩成，而且正是記錄存取最密集的地方。
    # 這些指令沒有函式可以分組，就以 `retn` 為界切成偽區塊，
    # 標成 `blk_<起始位址>`，免得整批被跳過（那會是安靜的漏報）。
    by_func: dict[str, list[dict]] = {}
    pseudo: str | None = None
    for insn in listing["instructions"]:
        func = insn["func"]
        if func is None:
            if pseudo is None:
                pseudo = f"blk_{insn['ea']}"
            by_func.setdefault(pseudo, []).append(insn)
            if insn["mnem"] in ("retn", "retf"):
                pseudo = None
            continue
        pseudo = None
        by_func.setdefault(func, []).append(insn)

    fields: dict[str, dict[int, list[dict]]] = {}
    unresolved: dict[str, int] = {}

    for func, insns in by_func.items():
        if func is None:
            continue

        # **往前跳**的目標要清掉追蹤狀態：線性掃描會把兩條互斥分支的 `inc bl`
        # 疊加，報出根本不存在的位移（踩過雷：`sub_14193` 兩條分支各 inc 一次，
        # 被算成 `+0x0A`，實際上兩邊都是 `+0x09`）。
        #
        # 往回跳（迴圈）不清：迴圈裡位移每輪都在變，任何單一值都不算「這一輪的」，
        # 而迴圈**進入前**的那個值正是要找的東西（陣列起點）。
        targets = set()
        for insn in insns:
            if not insn["mnem"].startswith("j"):
                continue
            here = int(insn["ea"], 16)
            for op in insn["ops"]:
                if op["kind"] != "near":
                    continue
                dest = int(op["addr"], 16)
                if (dest & 0xFFFF) > (here & 0xFFFF):
                    targets.add(op["addr"])

        offset: int | None = None
        base: str | None = None
        for insn in insns:
            if insn["ea"] in targets or f"0x{int(insn['ea'], 16) & 0xFFFF:X}" in targets:
                offset = None
                base = None
            text = " ".join(insn["disasm"].split())
            m = MOV_CONST.match(text)
            if m:
                offset = parse_const(m.group(2)) & 0xFF
                continue
            m = MOV_BASE.match(text)
            if m:
                base = m.group(1).lower()
                continue
            m = INCDEC.match(text)
            if m:
                if offset is not None:
                    offset = (offset + (1 if m.group(1) == "inc" else -1)) & 0xFF
                continue
            if "[bx+di]" in text:
                if base in bases:
                    if offset is None:
                        unresolved[base] = unresolved.get(base, 0) + 1
                    else:
                        fields.setdefault(base, {}).setdefault(offset, []).append(
                            {
                                "ida_address": insn["ea"],
                                "function": func,
                                "disasm": insn["disasm"],
                            }
                        )
                continue
            # bl／bx 被非常數方式改掉就放棄追蹤，避免報出錯的位移
            if CLOBBER.match(text) and not MOV_CONST.match(text):
                offset = None

    out = {
        "dataset": "record-field-map",
        "input_sha256": listing.get("input_sha256"),
        "bases": sorted(bases),
        "unresolved_accesses": unresolved,
        "fields": {
            base: {
                f"+0x{off:02X}": {
                    "access_count": len(sites),
                    "functions": sorted({s["function"] for s in sites}),
                    "sites": sites,
                }
                for off, sites in sorted(entries.items())
            }
            for base, entries in sorted(fields.items())
        },
    }
    print(json.dumps(out, ensure_ascii=False, indent=2))

    for base in sorted(fields):
        total = sum(len(v) for v in fields[base].values())
        print(
            f"# ds:{base.upper()}h：{len(fields[base])} 個位移、{total} 處存取、"
            f"{unresolved.get(base, 0)} 處位移追不到",
            file=sys.stderr,
        )
    print(f"# 函式總數 {len(functions)}", file=sys.stderr)


if __name__ == "__main__":
    main()
