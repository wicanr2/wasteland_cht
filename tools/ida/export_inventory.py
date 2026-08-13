"""DS000：wl.exe 基準清冊（函式、entry、segment、字串、呼叫關係）。

非破壞性匯出：只讀資料庫，不改名、不加註、不寫回。

用法（由 tools/ida.sh 呼叫）：
    tools/ida.sh run tools/ida/export_inventory.py <輸出.json>

每筆結論都要能追回 binary 身分，所以輸出一律帶輸入檔 SHA-256；
與 docs/re/01-baseline.md 記載的雜湊不符時直接失敗，不產出檔案。
"""

from __future__ import annotations

import json
import sys
from pathlib import Path

import ida_auto
import ida_bytes
import ida_entry
import ida_funcs
import ida_ida
import ida_nalt
import ida_pro
import ida_segment
import idautils

# 只認這兩份輸入。結論只能套用到這裡列出的 SHA-256，換一份 binary 就要重驗。
KNOWN_INPUTS = {
    "098aef9b4fe4fea3b8d0d134f82fe11a6dac608839ebd175e168cf0271b93b4f": "wl.exe（EXEPACK 打包原版）",
    "b5eb39f094e0274165eab5e1584e78ff5b54c7228d8db273573d2bd951ea31a0": "wl.unpacked.exe（tools/unpack_exepack.py 解包）",
    "cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118": "wl.merged.exe（解包映像＋wla.bin overlay，本專案合成）",
}


def hx(value: int) -> str:
    return f"0x{value:X}"


def seg_of(ea: int) -> tuple[str | None, str | None]:
    seg = ida_segment.getseg(ea)
    if seg is None:
        return None, None
    return ida_segment.get_segm_name(seg), hx(ea - seg.start_ea)


def collect_segments() -> list[dict]:
    out = []
    for ea in idautils.Segments():
        seg = ida_segment.getseg(ea)
        out.append(
            {
                "name": ida_segment.get_segm_name(seg),
                "class": ida_segment.get_segm_class(seg),
                "start_ea": hx(seg.start_ea),
                "end_ea": hx(seg.end_ea),
                "size": seg.end_ea - seg.start_ea,
                "bitness_bits": 16 << seg.bitness,
            }
        )
    return out


def collect_entries() -> list[dict]:
    out = []
    for i in range(ida_entry.get_entry_qty()):
        ordinal = ida_entry.get_entry_ordinal(i)
        ea = ida_entry.get_entry(ordinal)
        name, offset = seg_of(ea)
        out.append(
            {
                "ordinal": ordinal,
                "ida_address": hx(ea),
                "segment": name,
                "segment_offset": offset,
                "name": ida_entry.get_entry_name(ordinal),
            }
        )
    return out


def collect_functions() -> list[dict]:
    out = []
    for ea in idautils.Functions():
        func = ida_funcs.get_func(ea)
        if func is None:
            continue
        callers = sorted({f"0x{x.frm:X}" for x in idautils.XrefsTo(ea) if x.iscode})
        name, offset = seg_of(ea)
        out.append(
            {
                "ida_address": hx(ea),
                "segment": name,
                "segment_offset": offset,
                "original_name": ida_funcs.get_func_name(ea),
                "size": func.end_ea - func.start_ea,
                "caller_count": len(callers),
                "callers": callers[:32],
                "is_library": bool(func.flags & ida_funcs.FUNC_LIB),
                "no_return": bool(func.flags & ida_funcs.FUNC_NORET),
            }
        )
    return out


def collect_strings() -> list[dict]:
    out = []
    sc = idautils.Strings()
    sc.setup()
    for item in sc:
        raw = ida_bytes.get_strlit_contents(item.ea, item.length, item.strtype)
        if raw is None:
            continue
        name, offset = seg_of(item.ea)
        xrefs = sorted({f"0x{x.frm:X}" for x in idautils.XrefsTo(item.ea)})
        out.append(
            {
                "ida_address": hx(item.ea),
                "segment": name,
                "segment_offset": offset,
                "length": item.length,
                "text": raw.decode("cp437", errors="replace"),
                "xref_count": len(xrefs),
                "xrefs": xrefs[:16],
            }
        )
    return out


def main() -> None:
    if len(sys.argv) < 2:
        ida_pro.qexit(2)
    out_path = Path(sys.argv[1])

    ida_auto.auto_wait()

    sha256 = ida_nalt.retrieve_input_file_sha256().hex()
    if sha256 not in KNOWN_INPUTS:
        out_path.with_suffix(".error.txt").write_text(
            f"輸入檔雜湊 {sha256} 不在已知清單內，拒絕匯出\n", encoding="utf-8"
        )
        ida_pro.qexit(3)

    payload = {
        "dataset": "DS000-inventory",
        "input_file": ida_nalt.get_root_filename(),
        "input_identity": KNOWN_INPUTS[sha256],
        "input_sha256": sha256,
        "input_size": ida_nalt.retrieve_input_file_size(),
        "processor": ida_ida.inf_get_procname(),
        "segments": collect_segments(),
        "entries": collect_entries(),
        "functions": collect_functions(),
        "strings": collect_strings(),
    }

    out_path.write_text(
        json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8"
    )
    ida_pro.qexit(0)


main()
