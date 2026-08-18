#!/usr/bin/env python3
"""在整份反組譯裡找「誰引用了 [lo, hi) 這段位址」。

與 `tools/ida/export_range_refs.py` 的差別：那一支在 IDA 裡跑、比對的是
`get_operand_value()`；這一支在 IDA 外面跑，讀 `export_listing.py` 產的
JSON，比對的是**反組譯文字裡的十六進位字面值**。

⚠ 16 位元的同一個位址在反組譯裡有兩種長相：`[bx+7E0Bh]` 與 `[bx-81F5h]`。
只 grep 前者會漏掉後者，而症狀是安靜的零命中——與「真的沒人引用」
長得一模一樣（`docs/re/112` §1）。所以這裡把每個字面值**帶正負號正規化成
16 位元**再比。

正對照是強制的：`ds:A4E0h`（設施跳表，全檔唯一引用寫成 `[bx-5B20h]`）
一定要有命中，否則直接失敗——掃描器有洞的時候，零命中什麼都證明不了。

用法：
    python3 tools/scan_addr_refs.py <listing.json> <lo> <hi>
"""

import json
import re
import sys

LITERAL = re.compile(r"([+-])?\b([0-9A-F][0-9A-Fa-f]*)h")

# 正對照：這個位址在整份反組譯裡只出現一次，而且是負位移的寫法。
CONTROL_LO, CONTROL_HI = 0xA4E0, 0xA4E2


def literals(text):
    """反組譯文字裡的每個十六進位字面值，正規化成 16 位元無號。"""
    out = []
    for sign, digits in LITERAL.findall(text):
        value = int(digits, 16)
        if sign == "-":
            value = -value
        out.append(value & 0xFFFF)
    return out


def scan(instructions, lo, hi):
    hits = []
    for n, ins in enumerate(instructions):
        if any(lo <= v < hi for v in literals(ins["disasm"])):
            prev = instructions[n - 1]["disasm"] if n else ""
            hits.append((int(ins["ea"], 16), ins["disasm"], prev))
    return hits


def main(argv):
    if len(argv) != 4:
        sys.exit(__doc__)
    path, lo, hi = argv[1], int(argv[2], 0), int(argv[3], 0)
    if lo >= hi:
        sys.exit(f"範圍是半開區間 [lo, hi)，{lo:#x} >= {hi:#x} 一定回零命中")

    instructions = json.load(open(path))["instructions"]

    control = scan(instructions, CONTROL_LO, CONTROL_HI)
    if not control:
        sys.exit(
            f"正對照失敗：ds:{CONTROL_LO:04X}h 應該有命中卻是零。"
            "掃描器有洞，這一次的結果不能用。"
        )
    print(f"正對照 ds:{CONTROL_LO:04X}h：{len(control)} 命中 ✓（共 {len(instructions)} 條指令）")

    hits = scan(instructions, lo, hi)
    print(f"[{lo:04X}, {hi:04X}) {len(hits)} 命中")
    for ea, disasm, prev in hits:
        print(f"  {ea:07x}  {disasm:<38} <- {prev}")
    print(
        "⚠ 比對的是字面值，所以**存進去的立即數**也會命中（那是值不是位址）；"
        "先把位址載進暫存器再間接存取的則掃不到。"
    )


if __name__ == "__main__":
    main(sys.argv)
