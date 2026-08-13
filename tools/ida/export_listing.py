"""把整個 CODE 區的每一條指令倒成 JSON，之後所有掃描都在純 Python 做。

動機：每問一個新問題就跑一次 IDA 太慢，而且每支腳本都要重寫「走遍指令」
那段樣板。這支只做一次匯出，把逐指令的原始事實（bytes、助憶碼、運算元型別
與值、讀／寫特徵位元、所屬函式）全部留下，後續的形狀比對（找 RNG、找狀態機、
找欄位存取）就都變成離線的集合運算。

刻意不做任何過濾：不濾 `mov`／`add`／`shl`，因為那些常常正是索引計算與 stride。
讀／寫一律用 IDA 的指令特徵位元（`CF_CHGn`／`CF_USEn`），不解析指令文字。
立即數一律 `& 0xFFFF`（`get_operand_value` 會把 16-bit 立即數符號擴展）。

用法（由 tools/ida.sh 呼叫）：
    tools/ida.sh run tools/ida/export_listing.py <輸出.json>
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

import ida_auto
import ida_bytes
import ida_funcs
import ida_idp
import ida_lines
import ida_nalt
import ida_pro
import ida_segment
import ida_ua
import ida_xref
import idautils
import idc

KNOWN_INPUTS = {
    "b5eb39f094e0274165eab5e1584e78ff5b54c7228d8db273573d2bd951ea31a0",
    "cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118",
}

CHG_BITS = [
    ida_idp.CF_CHG1,
    ida_idp.CF_CHG2,
    ida_idp.CF_CHG3,
    ida_idp.CF_CHG4,
    ida_idp.CF_CHG5,
    ida_idp.CF_CHG6,
]
USE_BITS = [
    ida_idp.CF_USE1,
    ida_idp.CF_USE2,
    ida_idp.CF_USE3,
    ida_idp.CF_USE4,
    ida_idp.CF_USE5,
    ida_idp.CF_USE6,
]

OP_KIND = {
    ida_ua.o_void: "void",
    ida_ua.o_reg: "reg",
    ida_ua.o_mem: "mem",
    ida_ua.o_phrase: "phrase",
    ida_ua.o_displ: "displ",
    ida_ua.o_imm: "imm",
    ida_ua.o_far: "far",
    ida_ua.o_near: "near",
}


def main() -> None:
    if len(sys.argv) < 2:
        ida_pro.qexit(2)
    out_path = Path(sys.argv[1])

    ida_auto.auto_wait()
    sha = ida_nalt.retrieve_input_file_sha256()
    sha = sha.hex() if sha else None
    if sha not in KNOWN_INPUTS:
        out_path.with_suffix(".error.txt").write_text(
            f"輸入檔雜湊 {sha} 不在已知清單內\n", encoding="utf-8"
        )
        ida_pro.qexit(3)

    segs = []
    for seg_ea in idautils.Segments():
        seg = ida_segment.getseg(seg_ea)
        segs.append(
            {
                "name": ida_segment.get_segm_name(seg),
                "class": ida_segment.get_segm_class(seg),
                "start_ea": f"0x{seg.start_ea:X}",
                "end_ea": f"0x{seg.end_ea:X}",
            }
        )

    functions: dict[str, dict] = {}
    for func_ea in idautils.Functions():
        func = ida_funcs.get_func(func_ea)
        if func is None:
            continue
        callers = sorted(
            {
                r.frm
                for r in idautils.XrefsTo(func_ea, 0)
                if r.type in (ida_xref.fl_CN, ida_xref.fl_CF)
            }
        )
        functions[f"0x{func_ea:X}"] = {
            "name": ida_funcs.get_func_name(func_ea),
            "start_ea": f"0x{func.start_ea:X}",
            "end_ea": f"0x{func.end_ea:X}",
            "size": func.end_ea - func.start_ea,
            "segment": ida_segment.get_segm_name(ida_segment.getseg(func_ea)),
            "caller_count": len(callers),
            "callers": [f"0x{c:X}" for c in callers],
        }

    insns = []
    for seg_ea in idautils.Segments():
        seg = ida_segment.getseg(seg_ea)
        if ida_segment.get_segm_class(seg) != "CODE":
            continue
        ea = seg.start_ea
        while ea < seg.end_ea:
            flags = ida_bytes.get_flags(ea)
            size = max(1, ida_bytes.get_item_size(ea))
            if idc.is_code(flags):
                insn = ida_ua.insn_t()
                decoded = ida_ua.decode_insn(insn, ea)
                func = ida_funcs.get_func(ea)
                feature = insn.get_canon_feature() if decoded > 0 else 0
                ops = []
                if decoded > 0:
                    for n in range(6):
                        op = insn.ops[n]
                        if op.type == ida_ua.o_void:
                            break
                        ops.append(
                            {
                                "n": n,
                                "kind": OP_KIND.get(op.type, str(op.type)),
                                "reg": op.reg,
                                "addr": f"0x{op.addr & 0xFFFF:X}",
                                "value": f"0x{op.value & 0xFFFF:X}",
                                "phrase": op.phrase,
                                "dtype_size": ida_ua.get_dtype_size(op.dtype),
                                "write": bool(feature & CHG_BITS[n]),
                                "read": bool(feature & USE_BITS[n]),
                            }
                        )
                insns.append(
                    {
                        "ea": f"0x{ea:X}",
                        "seg": ida_segment.get_segm_name(seg),
                        "func": f"0x{func.start_ea:X}" if func is not None else None,
                        "size": insn.size if decoded > 0 else size,
                        "bytes": ida_bytes.get_bytes(ea, size).hex()
                        if ida_bytes.get_bytes(ea, size)
                        else "",
                        "mnem": idc.print_insn_mnem(ea).lower(),
                        "ops": ops,
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
                "dataset": "full-code-listing",
                "input_sha256": sha,
                "segments": segs,
                "instruction_count": len(insns),
                "function_count": len(functions),
                "functions": functions,
                "instructions": insns,
            },
            ensure_ascii=False,
        ),
        encoding="utf-8",
    )
    ida_pro.qexit(0)


main()
