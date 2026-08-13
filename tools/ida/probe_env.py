"""環境探針：確認 IDAPython 在容器內可用，並記錄版本與資料庫基本狀態。

用法（容器內）：
    idat -A -S"/workspace/tools/ida/probe_env.py /output/probe.json" /tmp/wl.exe

只讀，不改資料庫。輸出必須寫檔——headless 的 print 不進 stdout，
exit code 0 也不代表腳本真的跑到底。
"""

from __future__ import annotations

import json
import platform
import sys
from pathlib import Path

import ida_auto
import ida_ida
import ida_nalt
import ida_pro
import ida_segment
import idautils


def kernel_version() -> str:
    """9.4 把 get_kernel_version 從 ida_idaapi 移走了，逐一試而不是假設。"""
    import idaapi

    for getter in ("get_kernel_version", "get_ida_version"):
        fn = getattr(idaapi, getter, None)
        if callable(fn):
            return str(fn())
    return f"IDA_SDK_VERSION={getattr(idaapi, 'IDA_SDK_VERSION', 'unknown')}"


def main() -> None:
    if len(sys.argv) < 2:
        ida_pro.qexit(2)
    out = Path(sys.argv[1])

    ida_auto.auto_wait()

    segments = []
    for ea in idautils.Segments():
        seg = ida_segment.getseg(ea)
        segments.append(
            {
                "name": ida_segment.get_segm_name(seg),
                "start_ea": f"0x{seg.start_ea:X}",
                "end_ea": f"0x{seg.end_ea:X}",
                "class": ida_segment.get_segm_class(seg),
                "bitness": seg.bitness,
            }
        )

    payload = {
        "python_version": platform.python_version(),
        "python_executable": sys.executable,
        "ida_version": kernel_version(),
        "input_file": ida_nalt.get_root_filename(),
        "input_sha256": ida_nalt.retrieve_input_file_sha256().hex(),
        "processor": ida_ida.inf_get_procname(),
        "counts": {
            "segments": len(segments),
            "functions": len(list(idautils.Functions())),
            "names": len(list(idautils.Names())),
        },
        "segments": segments,
    }

    out.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")
    ida_pro.qexit(0)


main()
