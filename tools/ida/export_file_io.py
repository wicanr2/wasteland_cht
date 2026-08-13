"""DS001：DOS 檔案 I/O 的 sink，以及字串／立即數的引用關係。

要回答的問題：`game1`／`game2` 這些資料檔是誰、在哪裡、怎麼開的。
依靜態溯源的做法從「資料被用掉的那一點」開始——DOS 檔案服務都走 `int 21h`，
所以先把所有 `int 21h` 呼叫點與其服務號抓出來，再往回追檔名從哪來。

同時做兩件與字串有關的事：
  1. 自己掃 ASCII 片段，不只信 IDA 標成 string literal 的那 43 筆；
  2. 對每個片段同時收集 xref 圖的參考**與**把位址當立即數用的指令——
     xref 圖看不到 `mov dx, 1234h` 這種算術式引用，只看 xref 會得到假零。

用法（由 tools/ida.sh 呼叫）：
    tools/ida.sh run tools/ida/export_file_io.py <輸出.json>
"""

from __future__ import annotations

import json
import re
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
    "b5eb39f094e0274165eab5e1584e78ff5b54c7228d8db273573d2bd951ea31a0": (
        "wl.unpacked.exe（tools/unpack_exepack.py 解包）"
    ),
    "cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118": (
        "wl.merged.exe（解包映像＋wla.bin overlay，本專案合成）"
    ),
}

CONTEXT_INSNS = 14

# 只掃 int 21h 會漏掉真正的檔案存取路徑：1980 年代的遊戲常常繞過 DOS
# 直接用 BIOS INT 13h 讀磁區。所以掃全部中斷號，讓分布自己說話。
INT_NAMES = {
    0x08: "IRQ0 timer",
    0x09: "IRQ1 keyboard",
    0x10: "BIOS video",
    0x11: "BIOS equipment list",
    0x12: "BIOS memory size",
    0x13: "BIOS disk",
    0x14: "BIOS serial",
    0x16: "BIOS keyboard",
    0x17: "BIOS printer",
    0x19: "BIOS bootstrap",
    0x1A: "BIOS time",
    0x1C: "timer tick hook",
    0x20: "DOS terminate",
    0x21: "DOS function",
    0x23: "DOS ctrl-break",
    0x24: "DOS critical error",
    0x25: "DOS absolute disk read",
    0x26: "DOS absolute disk write",
    0x27: "DOS TSR",
    0x33: "mouse",
    0x67: "EMS",
}
MIN_RUN = 2
ASCII_RUN = re.compile(rb"[\x20-\x7e]{%d,}" % MIN_RUN)

# DOS INT 21h 的檔案相關服務，只列本專案會用到的；其餘照號碼原樣輸出。
DOS_SERVICES = {
    0x09: "print string",
    0x1A: "set DTA",
    0x25: "set interrupt vector",
    0x2C: "get time",
    0x30: "get DOS version",
    0x35: "get interrupt vector",
    0x3B: "chdir",
    0x3C: "create file",
    0x3D: "open file",
    0x3E: "close file",
    0x3F: "read file",
    0x40: "write file",
    0x41: "delete file",
    0x42: "seek",
    0x43: "get/set attributes",
    0x44: "ioctl",
    0x47: "get current directory",
    0x48: "allocate memory",
    0x49: "free memory",
    0x4A: "resize memory",
    0x4B: "exec",
    0x4C: "terminate",
    0x4E: "find first",
    0x4F: "find next",
}


def hx(v: int) -> str:
    return f"0x{v:X}"


def seg_of(ea: int) -> tuple[str | None, str | None]:
    seg = ida_segment.getseg(ea)
    if seg is None:
        return None, None
    return ida_segment.get_segm_name(seg), hx(ea - seg.start_ea)


def disasm(ea: int) -> str:
    return ida_lines.tag_remove(ida_lines.generate_disasm_line(ea, 0) or "")


def code_segments() -> list:
    out = []
    for ea in idautils.Segments():
        seg = ida_segment.getseg(ea)
        if ida_segment.get_segm_class(seg) == "CODE":
            out.append(seg)
    return out


def preceding_heads(ea: int, count: int) -> list[int]:
    """往前取 count 條指令位址（不跨函式邊界時盡量留在同一函式內）。"""
    func = ida_funcs.get_func(ea)
    low = func.start_ea if func is not None else 0
    out: list[int] = []
    cur = ea
    for _ in range(count):
        prev = idc.prev_head(cur)
        if prev == idc.BADADDR or prev < low:
            break
        out.append(prev)
        cur = prev
    return list(reversed(out))


def guess_ah(context: list[int]) -> int | None:
    """在前文找最後一次寫進 AH／AX 的立即數。找不到就誠實回 None。"""
    value = None
    for ea in context:
        if idc.print_insn_mnem(ea) != "mov":
            continue
        if idc.get_operand_type(ea, 0) != idc.o_reg:
            continue
        if idc.get_operand_type(ea, 1) != idc.o_imm:
            continue
        reg = idc.get_operand_value(ea, 0)
        imm = idc.get_operand_value(ea, 1)
        size = idc.get_item_size(ea)
        # 8-bit AH 的 reg 編號是 4；16-bit AX 是 0（高位元組即服務號）。
        text = disasm(ea)
        if reg == 4 and " ah," in text.lower():
            value = imm & 0xFF
        elif reg == 0 and " ax," in text.lower() and size >= 3:
            value = (imm >> 8) & 0xFF
    return value


def collect_interrupts() -> list[dict]:
    """掃全部 `int n`。IDA 認得的中斷指令才算，避免把資料裡的 0xCD 當指令。"""
    out = []
    for seg in code_segments():
        ea = seg.start_ea
        while ea < seg.end_ea:
            if idc.is_code(ida_bytes.get_flags(ea)) and idc.print_insn_mnem(ea) == "int":
                number = idc.get_operand_value(ea, 0)
                context = preceding_heads(ea, CONTEXT_INSNS)
                ah = guess_ah(context)
                func = ida_funcs.get_func(ea)
                name, offset = seg_of(ea)
                out.append(
                    {
                        "ida_address": hx(ea),
                        "segment": name,
                        "segment_offset": offset,
                        "interrupt": hx(number),
                        "interrupt_name": INT_NAMES.get(number),
                        "function": (
                            ida_funcs.get_func_name(func.start_ea)
                            if func is not None
                            else None
                        ),
                        "function_start": (
                            hx(func.start_ea) if func is not None else None
                        ),
                        "ah": hx(ah) if ah is not None else None,
                        "service": (
                            DOS_SERVICES.get(ah)
                            if ah is not None and number == 0x21
                            else None
                        ),
                        "context": [
                            {"ida_address": hx(c), "disasm": disasm(c)} for c in context
                        ],
                    }
                )
            nxt = idc.next_head(ea, seg.end_ea)
            if nxt in (idc.BADADDR, None) or nxt <= ea:
                break
            ea = nxt
    return out


def collect_ascii_runs() -> list[dict]:
    runs = []
    for ea in idautils.Segments():
        seg = ida_segment.getseg(ea)
        data = ida_bytes.get_bytes(seg.start_ea, seg.end_ea - seg.start_ea)
        if not data:
            continue
        segname = ida_segment.get_segm_name(seg)
        for match in ASCII_RUN.finditer(data):
            start = seg.start_ea + match.start()
            text = match.group().decode("ascii")
            runs.append(
                {
                    "ida_address": hx(start),
                    "segment": segname,
                    "segment_offset": hx(match.start()),
                    "length": len(text),
                    "text": text,
                    "in_code_segment": ida_segment.get_segm_class(seg) == "CODE",
                    "xrefs": sorted({hx(x.frm) for x in idautils.XrefsTo(start)})[:16],
                    "immediate_refs": [],
                }
            )
    return runs


def attach_immediate_refs(runs: list[dict]) -> dict[str, int]:
    """掃全部指令的立即數，落在某個 ASCII 片段內就記回去。

    這是 xref 圖看不到的那一半：位址被當純數字算的引用不會進 xref。
    """
    index: dict[int, dict] = {}
    for run in runs:
        base = int(run["ida_address"], 16)
        for off in range(run["length"]):
            index.setdefault(base + off, run)

    # 16-bit 程式的立即數是 segment-relative offset，所以同時試「段基址 + 立即數」。
    seg_bases = [ida_segment.getseg(ea).start_ea for ea in idautils.Segments()]
    scanned = 0
    matched = 0

    for seg in code_segments():
        ea = seg.start_ea
        while ea < seg.end_ea:
            if idc.is_code(ida_bytes.get_flags(ea)):
                scanned += 1
                for n in range(3):
                    if idc.get_operand_type(ea, n) != idc.o_imm:
                        continue
                    # ⚠ get_operand_value 會把 16-bit 立即數符號擴展成 64-bit
                    # （0x91C5 → 0xFFFFFFFFFFFF91C5）。不遮罩就永遠對不上任何
                    # 位址，而且症狀是安靜的零命中，看起來像「沒有人引用」。
                    imm = idc.get_operand_value(ea, n) & 0xFFFF
                    for base in seg_bases:
                        target = base + imm
                        run = index.get(target)
                        if run is None:
                            continue
                        entry = {
                            "ida_address": hx(ea),
                            "disasm": disasm(ea),
                            "target": hx(target),
                            "via_segment_base": hx(base),
                        }
                        if entry not in run["immediate_refs"]:
                            run["immediate_refs"].append(entry)
                            matched += 1
            nxt = idc.next_head(ea, seg.end_ea)
            if nxt in (idc.BADADDR, None) or nxt <= ea:
                break
            ea = nxt
    return {"instructions_scanned": scanned, "immediate_matches": matched}


def main() -> None:
    if len(sys.argv) < 2:
        ida_pro.qexit(2)
    out_path = Path(sys.argv[1])

    ida_auto.auto_wait()

    sha256 = ida_nalt.retrieve_input_file_sha256().hex()
    if sha256 not in KNOWN_INPUTS:
        out_path.with_suffix(".error.txt").write_text(
            f"輸入檔雜湊 {sha256} 不在已知清單內（本腳本只跑解包版），拒絕匯出\n",
            encoding="utf-8",
        )
        ida_pro.qexit(3)

    interrupts = collect_interrupts()
    runs = collect_ascii_runs()
    scan_stats = attach_immediate_refs(runs)

    payload = {
        "dataset": "DS001-file-io",
        "input_file": ida_nalt.get_root_filename(),
        "input_sha256": sha256,
        "input_identity": KNOWN_INPUTS[sha256],
        "scan_stats": scan_stats | {"ascii_runs": len(runs), "interrupt_calls": len(interrupts)},
        "interrupt_calls": interrupts,
        "ascii_runs": runs,
    }
    out_path.write_text(
        json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8"
    )
    ida_pro.qexit(0)


main()
