"""強制把指定位址分析成程式碼再倒出來（IDA 自動分析漏掉的區段）。

有些位址 IDA 沒建成程式碼——通常是只有間接跳表指到它、沒有直接 call。
`docs/re/26` 的事件分派表 `ds:AA87h` 就是這種：nibble 5／8／9 的處理函式
（`0x15280`／`0x15160`／`0x14410`）在資料庫裡是未定義的位元組。

這支在**資料庫的暫存副本**上做三件事（`tools/ida.sh run` 已保證是副本）：

    del_items → create_insn → add_func

然後把該函式逐指令倒進 JSON。canonical `.i64` 不受影響。

用法：
    tools/ida.sh run tools/ida/export_forced.py <out.json> 0x15280 0x15160 …
"""

from __future__ import annotations

import json
import os
import sys

import ida_auto
import ida_bytes
import ida_funcs
import ida_ida
import ida_kernwin
import ida_nalt
import idautils
import idc

KNOWN_SHA256 = {
    "cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118",  # wl.merged.exe
    "b5eb39f0...31a0",  # 佔位：解包版的雜湊由 docs/re/01 提供
}
MAX_INSN = 400


def bail(msg: str) -> None:
    sys.stderr.write(msg + "\n")
    ida_kernwin.qexit(1)


def force_function(ea: int) -> dict:
    """把 ea 起的位元組轉成程式碼並建成函式，回傳逐指令內容。"""
    info = {"requested": f"{ea:#x}"}

    func = ida_funcs.get_func(ea)
    if func is None:
        # 先清掉可能被誤判成資料的位元組，再逐條建指令。
        ida_bytes.del_items(ea, ida_bytes.DELIT_EXPAND, 16)
        if not idc.create_insn(ea):
            info["error"] = "create_insn 失敗（這個位址不是有效的指令開頭？）"
            return info
        if not ida_funcs.add_func(ea):
            info["note"] = "add_func 失敗，改用線性反組譯到第一個 retn"
        ida_auto.auto_wait()
        func = ida_funcs.get_func(ea)

    insns = []
    cur = ea
    end = func.end_ea if func else None
    while len(insns) < MAX_INSN:
        if not ida_bytes.is_code(ida_bytes.get_flags(cur)):
            ida_bytes.del_items(cur, ida_bytes.DELIT_EXPAND, 16)
            if not idc.create_insn(cur):
                break
        size = idc.get_item_size(cur)
        insns.append(
            {
                "ea": f"{cur:#x}",
                "bytes": ida_bytes.get_bytes(cur, size).hex(),
                "disasm": idc.GetDisasm(cur),
                "mnem": idc.print_insn_mnem(cur),
            }
        )
        mnem = idc.print_insn_mnem(cur)
        cur += size
        if mnem in ("retn", "retf", "iret"):
            break
        if end is not None and cur >= end:
            break

    info["function_start"] = f"{func.start_ea:#x}" if func else None
    info["function_end"] = f"{func.end_ea:#x}" if func else None
    info["instruction_count"] = len(insns)
    info["instructions"] = insns
    info["callers"] = sorted(
        {f"{x:#x}" for x in idautils.CodeRefsTo(ea, 0)}
    )
    return info


def main() -> None:
    if len(sys.argv) < 3:
        bail("用法：export_forced.py <out.json> <位址…>")
    out_path = sys.argv[1]
    targets = [int(a, 16) for a in sys.argv[2:]]

    ida_auto.auto_wait()
    sha = ida_nalt.retrieve_input_file_sha256()
    sha = sha.hex() if sha else None

    result = {
        "dataset": "forced-functions",
        "input_sha256": sha,
        "min_ea": f"{ida_ida.inf_get_min_ea():#x}",
        "functions": [force_function(ea) for ea in targets],
    }
    with open(out_path, "w", encoding="utf-8") as fh:
        json.dump(result, fh, ensure_ascii=False, indent=1)
    # headless 的 print 不進 stdout，所以寫一份大小到檔案旁邊備查。
    sys.stderr.write(f"寫出 {out_path}（{os.path.getsize(out_path)} bytes）\n")
    ida_kernwin.qexit(0)


main()
