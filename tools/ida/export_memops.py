"""全檔掃描「直接定址的記憶體存取」，逐筆標出讀／寫。

為什麼不用 xref：這份資料庫裡 IDA 只替 22 個位址建了資料 xref，
絕大多數 `mov ax, ds:46B7h` 這種直接定址並沒有進 xref 圖。
拿 xref 問「誰在寫這個變數」會得到安靜的零命中，和「真的沒人寫」長得一模一樣。

讀／寫的判定用 IDA 自己的指令特徵位元（`CF_CHGn`／`CF_USEn`），
不解析指令文字——`push` 的第 0 個運算元是來源不是目的，用文字判會全部判錯。

段基底：本專案的解包映像裡 DS 對到 `seg002` 起點，所以
`ds:XXXX` 的線性位址 ＝ seg002.start_ea ＋ XXXX。腳本會同時記錄
原始 `op.addr` 與 IDA 自己算出的線性位址（`map_data_ea`），兩者都留著供校正。

用法（由 tools/ida.sh 呼叫）：
    tools/ida.sh run tools/ida/export_memops.py <輸出.json>
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


def linear_of(insn, op) -> int | None:
    """把 o_mem 運算元換成線性位址，優先用 IDA 自己的對應。"""
    try:
        ea = ida_ua.map_data_ea(insn, op)
        if ea not in (idc.BADADDR, None):
            return ea
    except Exception:
        pass
    return None


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

    ops: list[dict] = []
    func_info: dict[str, dict] = {}

    for seg_ea in idautils.Segments():
        seg = ida_segment.getseg(seg_ea)
        if ida_segment.get_segm_class(seg) != "CODE":
            continue
        ea = seg.start_ea
        while ea < seg.end_ea:
            if not idc.is_code(ida_bytes.get_flags(ea)):
                nxt = idc.next_head(ea, seg.end_ea)
                if nxt in (idc.BADADDR, None) or nxt <= ea:
                    break
                ea = nxt
                continue

            insn = ida_ua.insn_t()
            if ida_ua.decode_insn(insn, ea) > 0:
                feature = insn.get_canon_feature()
                func = ida_funcs.get_func(ea)
                fstart = f"0x{func.start_ea:X}" if func is not None else None
                if func is not None and fstart not in func_info:
                    callers = {
                        r.frm
                        for r in idautils.XrefsTo(func.start_ea, 0)
                        if r.type in (ida_xref.fl_CN, ida_xref.fl_CF)
                    }
                    func_info[fstart] = {
                        "name": ida_funcs.get_func_name(func.start_ea),
                        "size": func.end_ea - func.start_ea,
                        "caller_count": len(callers),
                        "segment": ida_segment.get_segm_name(
                            ida_segment.getseg(func.start_ea)
                        ),
                    }
                for n in range(6):
                    op = insn.ops[n]
                    if op.type == ida_ua.o_void:
                        break
                    if op.type != ida_ua.o_mem:
                        continue
                    ops.append(
                        {
                            "ida_address": f"0x{ea:X}",
                            "function_start": fstart,
                            "mnemonic": idc.print_insn_mnem(ea).lower(),
                            "op_index": n,
                            "op_addr": f"0x{op.addr & 0xFFFF:X}",
                            "linear": (
                                f"0x{linear_of(insn, op):X}"
                                if linear_of(insn, op) is not None
                                else None
                            ),
                            "dtype_size": ida_ua.get_dtype_size(op.dtype),
                            "write": bool(feature & CHG_BITS[n]),
                            "read": bool(feature & USE_BITS[n]),
                            "disasm": ida_lines.tag_remove(
                                ida_lines.generate_disasm_line(ea, 0) or ""
                            ),
                        }
                    )

            nxt = idc.next_head(ea, seg.end_ea)
            if nxt in (idc.BADADDR, None) or nxt <= ea:
                break
            ea = nxt

    # 校正用：已經有 xref 的位址，看 op.addr／linear 對不對得上
    calibration = []
    for entry in ops[:400]:
        site = int(entry["ida_address"], 16)
        drefs = [f"0x{t:X}" for t in idautils.DataRefsFrom(site)]
        if drefs:
            calibration.append({**entry, "dataref_targets": drefs})

    out_path.write_text(
        json.dumps(
            {
                "dataset": "direct-memory-operands",
                "input_sha256": sha,
                "segments": segs,
                "op_count": len(ops),
                "calibration_samples": calibration[:40],
                "functions": func_info,
                "ops": ops,
            },
            ensure_ascii=False,
            indent=2,
        ),
        encoding="utf-8",
    )
    ida_pro.qexit(0)


main()
