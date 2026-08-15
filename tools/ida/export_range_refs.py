"""掃全部指令，找出立即數或記憶體位移落在指定範圍內的引用。

用途：知道「某個結構在 ds:X 起」之後，反過來問「誰在碰它」。
xref 圖只涵蓋 IDA 認得出來的參考，把位址當純數字算的程式碼在圖上是空的，
所以這裡直接掃運算元數值。

⚠ 16-bit 的 `get_operand_value` 會把立即數符號擴展（0x91C5 → 0xFFFF…91C5），
一律 `& 0xFFFF` 之後再比對，否則會得到安靜的零命中。

用法（由 tools/ida.sh 呼叫）：
    tools/ida.sh run tools/ida/export_range_refs.py <輸出.json> 0x800 0x900

⚠ 範圍是**半開區間** [lo, hi)。要查單一位址得寫 `0xA5C5 0xA5C6`——
寫成 `0xA5C5 0xA5C5` 會回零命中，而那和「真的沒人碰」長得一模一樣。
下界 ≥ 上界時本腳本直接拒絕產出，不讓它變成安靜的假零。
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_lines
import ida_nalt
import ida_pro
import ida_segment
import idautils
import idc

KNOWN_INPUTS = {
    "b5eb39f094e0274165eab5e1584e78ff5b54c7228d8db273573d2bd951ea31a0",
    "cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118",
}


def hx(v: int) -> str:
    return f"0x{v:X}"


def main() -> None:
    if len(sys.argv) < 4:
        ida_pro.qexit(2)
    out_path = Path(sys.argv[1])
    lo = int(sys.argv[2], 0)
    hi = int(sys.argv[3], 0)
    if hi <= lo:
        raise SystemExit(
            f"範圍 [{lo:#x}, {hi:#x}) 是空的——半開區間，"
            f"查單一位址要寫 {lo:#x} {lo + 1:#x}"
        )

    ida_auto.auto_wait()
    sha = ida_nalt.retrieve_input_file_sha256()
    sha = sha.hex() if sha else None
    if sha not in KNOWN_INPUTS:
        out_path.with_suffix(".error.txt").write_text(
            f"輸入檔雜湊 {sha} 不在已知清單內\n", encoding="utf-8"
        )
        ida_pro.qexit(3)

    hits = []
    scanned = 0
    for seg_ea in idautils.Segments():
        seg = ida_segment.getseg(seg_ea)
        if ida_segment.get_segm_class(seg) != "CODE":
            continue
        ea = seg.start_ea
        while ea < seg.end_ea:
            if idc.is_code(ida_bytes.get_flags(ea)):
                scanned += 1
                for n in range(3):
                    otype = idc.get_operand_type(ea, n)
                    if otype not in (idc.o_imm, idc.o_mem, idc.o_displ):
                        continue
                    value = idc.get_operand_value(ea, n) & 0xFFFF
                    if not (lo <= value < hi):
                        continue
                    func = ida_funcs.get_func(ea)
                    hits.append(
                        {
                            "ida_address": hx(ea),
                            "function": (
                                ida_funcs.get_func_name(func.start_ea)
                                if func is not None
                                else None
                            ),
                            "operand": n,
                            "operand_type": otype,
                            "value": hx(value),
                            "disasm": ida_lines.tag_remove(
                                ida_lines.generate_disasm_line(ea, 0) or ""
                            ),
                        }
                    )
            nxt = idc.next_head(ea, seg.end_ea)
            if nxt in (idc.BADADDR, None) or nxt <= ea:
                break
            ea = nxt

    out_path.write_text(
        json.dumps(
            {
                "input_sha256": sha,
                "range": [hx(lo), hx(hi)],
                "instructions_scanned": scanned,
                "hit_count": len(hits),
                "hits": hits,
            },
            ensure_ascii=False,
            indent=2,
        ),
        encoding="utf-8",
    )
    ida_pro.qexit(0)


main()
