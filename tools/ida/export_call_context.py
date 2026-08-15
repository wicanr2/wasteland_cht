"""列出某個目標的所有呼叫端，並把每個呼叫點**前面**若干條指令原樣倒出來。

用途：回答「呼叫這支的時候暫存器裡放的是什麼」。這種問題沒辦法用 xref 清單
回答——xref 只說「這裡有人叫」，不說「叫的時候 `al` 是幾」。

⚠ 倒出來的是**未過濾的**指令序列，包含 `mov`／`shl`／`and`。看起來像樣板的
那幾條常常正是索引計算。要縮短輸出請調 --before，不要在這裡先濾。

⚠ 這支**不做語意判斷**：它不會說「這個呼叫傳的是 7」，只會把指令倒出來
讓人自己讀。立即數的符號擴展問題（`get_operand_value` 的 16-bit 陷阱，
`docs/re/03` §7）在這裡不存在，因為輸出的是反組譯文字與原始 bytes。

用法（由 tools/ida.sh 呼叫）：
    tools/ida.sh run tools/ida/export_call_context.py <輸出.json> 0x1000C [--before 24]
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


def line(ea: int) -> dict:
    return {
        "ea": f"0x{ea:X}",
        "seg": f"{ida_segment.get_segm_name(ida_segment.getseg(ea))}:{ea - ida_segment.getseg(ea).start_ea:04X}",
        "bytes": ida_bytes.get_bytes(ea, idc.get_item_size(ea)).hex(),
        "text": ida_lines.tag_remove(ida_lines.generate_disasm_line(ea, 0)),
    }


def main() -> None:
    if len(sys.argv) < 3:
        ida_pro.qexit(2)
    out_path = Path(sys.argv[1])
    args = sys.argv[2:]
    before = 24
    if "--before" in args:
        i = args.index("--before")
        before = int(args[i + 1])
        del args[i : i + 2]
    targets = [int(a, 16) if a.startswith("0x") else idc.get_name_ea_simple(a) for a in args]

    ida_auto.auto_wait()
    sha = ida_nalt.retrieve_input_file_sha256()
    sha = sha.hex() if isinstance(sha, (bytes, bytearray)) else sha
    if sha not in KNOWN_INPUTS:
        out_path.write_text(json.dumps({"error": "unknown input", "sha256": sha}))
        ida_pro.qexit(3)

    result = {"input_sha256": sha, "before": before, "targets": []}
    for t in targets:
        sites = []
        for xref in idautils.CodeRefsTo(t, 0):
            f = ida_funcs.get_func(xref)
            # 往回走：用函式起點當下界，避免走進上一個函式。
            lo = f.start_ea if f else xref
            chain = []
            ea = xref
            for _ in range(before):
                prev = idc.prev_head(ea, lo)
                if prev == idc.BADADDR or prev < lo:
                    break
                chain.append(prev)
                ea = prev
            sites.append(
                {
                    "call_ea": f"0x{xref:X}",
                    "in_function": ida_funcs.get_func_name(xref) or "",
                    "func_start": f"0x{f.start_ea:X}" if f else None,
                    "context": [line(e) for e in reversed(chain)] + [line(xref)],
                }
            )
        sites.sort(key=lambda s: int(s["call_ea"], 16))
        result["targets"].append(
            {"target": f"0x{t:X}", "name": ida_funcs.get_func_name(t) or "", "sites": sites, "count": len(sites)}
        )

    out_path.write_text(json.dumps(result, indent=1))
    ida_pro.qexit(0)


main()
