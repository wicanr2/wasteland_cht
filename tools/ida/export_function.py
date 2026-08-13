"""DS-tool：把指定函式完整倒出來（位址、bytes、未過濾組語、呼叫關係、資料引用）。

這是所有逐函式分析的基礎工具。刻意不做任何過濾——把「看起來像樣板」的
`mov`／`add`／`shl` 濾掉，濾掉的往往正是索引計算與 stride。

用法（由 tools/ida.sh 呼叫，位址或名稱都收）：
    tools/ida.sh run tools/ida/export_function.py <輸出.json> 0x116AC sub_11854
    tools/ida.sh run tools/ida/export_function.py <輸出.json> 0x116AC --callers
        --callers 會把每個目標函式的直接呼叫端也一起倒出來
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
import ida_name
import ida_pro
import ida_segment
import idautils
import idc

KNOWN_INPUTS = {
    "b5eb39f094e0274165eab5e1584e78ff5b54c7228d8db273573d2bd951ea31a0": (
        "wl.unpacked.exe（tools/unpack_exepack.py 解包）"
    ),
    "cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118": (
        "wl.merged.exe（解包映像＋wla.bin overlay，本專案合成）"
    ),
}


def hx(v: int) -> str:
    return f"0x{v:X}"


def seg_of(ea: int) -> tuple[str | None, str | None]:
    seg = ida_segment.getseg(ea)
    if seg is None:
        return None, None
    return ida_segment.get_segm_name(seg), hx(ea - seg.start_ea)


def resolve(token: str) -> int | None:
    if token.lower().startswith("0x"):
        try:
            return int(token, 16)
        except ValueError:
            return None
    ea = ida_name.get_name_ea(idc.BADADDR, token)
    return None if ea == idc.BADADDR else ea


def dump_function(ea: int) -> dict | None:
    func = ida_funcs.get_func(ea)
    if func is None:
        return None

    lines = []
    callees = []
    data_refs = []
    for head in idautils.Heads(func.start_ea, func.end_ea):
        size = idc.get_item_size(head)
        raw = ida_bytes.get_bytes(head, size) or b""
        segname, offset = seg_of(head)
        lines.append(
            {
                "ida_address": hx(head),
                "segment_offset": f"{segname}+{offset}",
                "bytes": raw.hex(),
                "disasm": ida_lines.tag_remove(
                    ida_lines.generate_disasm_line(head, 0) or ""
                ),
            }
        )
        for xref in idautils.XrefsFrom(head, 0):
            if xref.iscode and idc.print_insn_mnem(head) in ("call", "jmp"):
                target = ida_funcs.get_func(xref.to)
                if target is not None and target.start_ea != func.start_ea:
                    entry = {
                        "from": hx(head),
                        "target": hx(target.start_ea),
                        "name": ida_funcs.get_func_name(target.start_ea),
                    }
                    if entry not in callees:
                        callees.append(entry)
            elif not xref.iscode:
                entry = {
                    "from": hx(head),
                    "target": hx(xref.to),
                    "name": ida_name.get_name(xref.to) or None,
                }
                if entry not in data_refs:
                    data_refs.append(entry)

    callers = []
    for xref in idautils.XrefsTo(func.start_ea):
        if not xref.iscode:
            continue
        caller = ida_funcs.get_func(xref.frm)
        callers.append(
            {
                "from": hx(xref.frm),
                "function": (
                    ida_funcs.get_func_name(caller.start_ea)
                    if caller is not None
                    else None
                ),
                "function_start": hx(caller.start_ea) if caller is not None else None,
                "disasm": ida_lines.tag_remove(
                    ida_lines.generate_disasm_line(xref.frm, 0) or ""
                ),
            }
        )

    segname, offset = seg_of(func.start_ea)
    return {
        "ida_address": hx(func.start_ea),
        "end_ea": hx(func.end_ea),
        "segment": segname,
        "segment_offset": offset,
        "original_name": ida_funcs.get_func_name(func.start_ea),
        "size": func.end_ea - func.start_ea,
        "instruction_count": len(lines),
        "callers": callers,
        "callees": callees,
        "data_refs": data_refs,
        "listing": lines,
    }


def main() -> None:
    if len(sys.argv) < 3:
        ida_pro.qexit(2)
    out_path = Path(sys.argv[1])
    tokens = [t for t in sys.argv[2:] if not t.startswith("--")]
    with_callers = "--callers" in sys.argv[2:]

    ida_auto.auto_wait()

    sha256 = ida_nalt.retrieve_input_file_sha256()
    sha256 = sha256.hex() if sha256 else None
    if sha256 not in KNOWN_INPUTS:
        out_path.with_suffix(".error.txt").write_text(
            f"輸入檔雜湊 {sha256} 不在已知清單內，拒絕匯出\n", encoding="utf-8"
        )
        ida_pro.qexit(3)

    wanted: list[int] = []
    unresolved: list[str] = []
    for token in tokens:
        ea = resolve(token)
        if ea is None:
            unresolved.append(token)
        elif ea not in wanted:
            wanted.append(ea)

    functions = []
    seen: set[int] = set()
    queue = list(wanted)
    while queue:
        ea = queue.pop(0)
        func = ida_funcs.get_func(ea)
        if func is None or func.start_ea in seen:
            continue
        seen.add(func.start_ea)
        dumped = dump_function(ea)
        if dumped is None:
            continue
        functions.append(dumped)
        if with_callers and ea in wanted:
            for caller in dumped["callers"]:
                if caller["function_start"]:
                    queue.append(int(caller["function_start"], 16))

    payload = {
        "dataset": "function-dump",
        "input_sha256": sha256,
        "input_identity": KNOWN_INPUTS[sha256],
        "requested": tokens,
        "unresolved": unresolved,
        "with_callers": with_callers,
        "functions": functions,
    }
    out_path.write_text(
        json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8"
    )
    ida_pro.qexit(0)


main()
