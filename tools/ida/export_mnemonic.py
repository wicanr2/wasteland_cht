"""掃指定助憶碼的所有出現，列出所屬函式與運算元。

找特定機制時很有用：RNG 找 `mul`、音效找 `out`、位元運算找 `rol`／`ror`。
比「猜哪個函式是什麼」可靠，因為它問的是指令本身。

用法（由 tools/ida.sh 呼叫）：
    tools/ida.sh run tools/ida/export_mnemonic.py <輸出.json> mul imul rol
"""

from __future__ import annotations

import json
import sys
from collections import Counter
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


def main() -> None:
    if len(sys.argv) < 3:
        ida_pro.qexit(2)
    out_path = Path(sys.argv[1])
    wanted = {m.lower() for m in sys.argv[2:]}

    ida_auto.auto_wait()
    sha = ida_nalt.retrieve_input_file_sha256()
    sha = sha.hex() if sha else None
    if sha not in KNOWN_INPUTS:
        out_path.with_suffix(".error.txt").write_text(
            f"輸入檔雜湊 {sha} 不在已知清單內\n", encoding="utf-8"
        )
        ida_pro.qexit(3)

    hits = []
    for seg_ea in idautils.Segments():
        seg = ida_segment.getseg(seg_ea)
        if ida_segment.get_segm_class(seg) != "CODE":
            continue
        ea = seg.start_ea
        while ea < seg.end_ea:
            if idc.is_code(ida_bytes.get_flags(ea)):
                mnem = idc.print_insn_mnem(ea).lower()
                if mnem in wanted:
                    func = ida_funcs.get_func(ea)
                    hits.append(
                        {
                            "ida_address": f"0x{ea:X}",
                            "mnemonic": mnem,
                            "function": (
                                ida_funcs.get_func_name(func.start_ea)
                                if func is not None
                                else None
                            ),
                            "function_start": (
                                f"0x{func.start_ea:X}" if func is not None else None
                            ),
                            "function_size": (
                                func.end_ea - func.start_ea if func is not None else None
                            ),
                            "disasm": ida_lines.tag_remove(
                                ida_lines.generate_disasm_line(ea, 0) or ""
                            ),
                        }
                    )
            nxt = idc.next_head(ea, seg.end_ea)
            if nxt in (idc.BADADDR, None) or nxt <= ea:
                break
            ea = nxt

    by_func = Counter(h["function"] for h in hits)
    out_path.write_text(
        json.dumps(
            {
                "input_sha256": sha,
                "mnemonics": sorted(wanted),
                "hit_count": len(hits),
                "by_function": by_func.most_common(),
                "hits": hits,
            },
            ensure_ascii=False,
            indent=2,
        ),
        encoding="utf-8",
    )
    ida_pro.qexit(0)


main()
