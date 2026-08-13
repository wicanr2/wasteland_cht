"""解析 MSQ 區塊的內部佈局，只輸出結構與統計（不倒出內容）。

佈局來自 `docs/re/16`：

    +0x0000            地圖（P bytes，P ＝ 0x600 或 0x1800，由執行檔內的選擇表決定）
    +P                 記錄區標頭（0x5C bytes）
    +P+0x5C            第一個 section
    …                  各 section 依標頭裡的位移排列

標頭裡的位移不是連續存放的，要透過執行檔內 `ds:B9E0h` 那張「section 型別 → 標頭位移」
的表去查（`sub_17CB1` 就是這樣做的）。

用法：
    python3 tools/summarize_msq_layout.py <wl.merged.exe> <game1> <game2> <out.json>
"""

from __future__ import annotations

import json
import struct
import sys
from pathlib import Path

DS = 0x1CE20
BLOCK_LENGTHS = 0xBD86
DIRECTORY = 0xBEC9
MAP_SIZE_SELECTOR = 0xBF1C  # 每個資源一個 byte：0x40 → 地圖 0x1800，其餘 → 0x600
SECTION_TABLE = 0xB9E0  # section 型別 → 標頭內位移
SECTION_TYPES = 24
HEADER_SIZE = 0x5C
# 只有這幾個型別會經由 sub_17CB1 取用，也就是「指標陣列 ＋ 記錄」那種形狀
POINTER_ARRAY_TYPES = (3, 5, 15, 16, 17)


def decrypt(data: bytes, key: int) -> bytearray:
    out = bytearray(len(data))
    for i, c in enumerate(data):
        out[i] = c ^ key
        key = (key + 0x1F) & 0xFF
    return out


def main() -> None:
    exe_p, g1_p, g2_p, out_p = (Path(p) for p in sys.argv[1:5])
    exe = exe_p.read_bytes()
    header_bytes = struct.unpack_from("<H", exe, 8)[0] * 16
    at = lambda off: DS + off - 0x10000 + header_bytes  # noqa: E731

    directory: list[int] = []
    p = at(DIRECTORY)
    while exe[p] != 0xFF:
        directory.append(exe[p])
        p += 1

    selector = exe[at(MAP_SIZE_SELECTOR) : at(MAP_SIZE_SELECTOR) + len(directory)]
    block_len = [
        struct.unpack_from("<H", exe, at(BLOCK_LENGTHS) + i * 2)[0]
        for i in range(len(directory))
    ]
    section_offsets = [
        struct.unpack_from("<H", exe, at(SECTION_TABLE) + i * 2)[0]
        for i in range(SECTION_TYPES)
    ]

    files = {"game1": g1_p.read_bytes(), "game2": g2_p.read_bytes()}
    cursor = {"game1": 0, "game2": 0}
    blocks = []

    for res_id, raw in enumerate(directory):
        label = {0x80: "game1", 0x40: "game2"}.get(raw & 0xC0)
        if label is None or block_len[res_id] == 0:
            continue
        data = files[label]
        off = cursor[label]
        cursor[label] = off + block_len[res_id]
        span = data[off : off + block_len[res_id]]
        checksum = struct.unpack_from("<H", span, 4)[0]
        body = decrypt(span[6:], (checksum & 0xFF) ^ (checksum >> 8))

        map_size = 0x1800 if selector[res_id] == 0x40 else 0x600
        entry: dict = {
            "resource_id": res_id,
            "file": label,
            "body_length": len(body),
            "map_size": f"0x{map_size:04X}",
            "map_cells": map_size * 2,
        }
        if map_size + HEADER_SIZE > len(body):
            entry["note"] = "地圖區比區塊還大，佈局不適用"
            blocks.append(entry)
            continue

        sections = {}
        for kind, rel in enumerate(section_offsets):
            if rel == 0:
                continue
            start = struct.unpack_from("<H", body, map_size + rel)[0]
            if start == 0:
                continue
            sections[kind] = start
        ordered = sorted(set(sections.values()))

        entry["header_at"] = f"0x{map_size:04X}"
        entry["first_section"] = (
            f"0x{ordered[0]:04X}" if ordered else None
        )
        entry["first_section_is_header_end"] = bool(
            ordered and ordered[0] == map_size + HEADER_SIZE
        )
        entry["sections"] = {}
        for kind, start in sorted(sections.items()):
            nxt = next((s for s in ordered if s > start), len(body))
            info: dict = {
                "header_offset": f"+0x{section_offsets[kind]:02X}",
                "start": f"0x{start:04X}",
                "end": f"0x{nxt:04X}",
                "size": nxt - start,
            }
            if kind in POINTER_ARRAY_TYPES and start + 2 <= len(body):
                first = struct.unpack_from("<H", body, start)[0]
                plausible = start < first <= len(body) and (first - start) % 2 == 0
                if plausible:
                    count = (first - start) // 2
                    pointers = [
                        struct.unpack_from("<H", body, start + 2 * k)[0]
                        for k in range(count)
                    ]
                    plausible = all(q == 0 or first <= q <= len(body) for q in pointers)
                    info["record_count"] = count
                    info["empty_slots"] = sum(1 for q in pointers if q == 0)
                info["pointer_array"] = plausible
            entry["sections"][kind] = info
        blocks.append(entry)

    summary = {
        "dataset": "msq-block-layout",
        "block_count": len(blocks),
        "first_section_at_header_end": sum(
            1 for b in blocks if b.get("first_section_is_header_end")
        ),
        "pointer_array_ok": {
            str(t): sum(
                1
                for b in blocks
                for k, v in b.get("sections", {}).items()
                if k == t and v.get("pointer_array")
            )
            for t in POINTER_ARRAY_TYPES
        },
        "pointer_array_present": {
            str(t): sum(
                1
                for b in blocks
                for k, v in b.get("sections", {}).items()
                if k == t and "pointer_array" in v
            )
            for t in POINTER_ARRAY_TYPES
        },
    }

    out_p.write_text(
        json.dumps(
            {
                "summary": summary,
                "section_type_to_header_offset": {
                    str(i): f"+0x{v:02X}" for i, v in enumerate(section_offsets) if v
                },
                "blocks": blocks,
            },
            ensure_ascii=False,
            indent=2,
        ),
        encoding="utf-8",
    )
    print(json.dumps(summary, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
