"""用執行檔內的位移表把 `game1`／`game2` 切成資源。

定址方式見 docs/re/06：位移表在 `wl.exe` 的資料段，不在資料檔裡。
每個資源開頭 4 bytes 是 magic（`msq0`／`msq1`），切點正確與否就看這個——
42 個切點全部落在 magic 上，機率上不可能是巧合。

只輸出統計與少量 hex，不倒出資料內容（原版資產不散布）。

用法（容器內）：
    python3 tools/split_resources.py <wl.unpacked.exe> <game1> <game2> <out.json>
"""

from __future__ import annotations

import hashlib
import json
import struct
import sys
from pathlib import Path

DS = 0x1CE20
TABLE_A = 0xBC7A  # ds:9168h == 0x80 走這張
TABLE_B = 0xBCCA
DIRECTORY = 0xBEC9
PRINTABLE = set(range(0x20, 0x7F)) | {0x0A, 0x0D, 0x09}


def read_tables(exe: bytes, ceilings: tuple[int, int]) -> dict:
    header = struct.unpack_from("<H", exe, 8)[0] * 16

    def at(ds_off: int) -> int:
        return DS + ds_off - 0x10000 + header

    def offsets(base: int, count: int) -> list[int]:
        return [struct.unpack_from("<I", exe, at(base) + i * 4)[0] for i in range(count)]

    # 表長度不寫死：遞增且落在對應資料檔大小之內才算。
    def grow(base: int, ceiling: int, limit: int = 64) -> list[int]:
        """遞增且落在檔案大小之內才算數。

        只看「是否遞增」會多收兩筆：表 B 後面兩個垃圾值 0x13852187、0x25652407
        碰巧也是遞增的，但遠超過 game2 的大小。
        """
        vals = [struct.unpack_from("<I", exe, at(base) + i * 4)[0] for i in range(limit)]
        out = [vals[0]]
        for v in vals[1:]:
            if v <= out[-1] or v >= ceiling:
                break
            out.append(v)
        return out

    directory = []
    p = at(DIRECTORY)
    while True:
        b = exe[p]
        if b == 0xFF:
            break
        directory.append(b)
        p += 1

    return {
        "table_a": grow(TABLE_A, ceilings[0]),
        "table_b": grow(TABLE_B, ceilings[1]),
        "directory": directory,
        "offsets_fn": offsets,
    }


def describe(blob: bytes) -> dict:
    if not blob:
        return {"empty": True}
    printable = sum(1 for b in blob if b in PRINTABLE)
    return {
        "head_hex": blob[:16].hex(),
        "printable_ratio": round(printable / len(blob), 3),
        "distinct_bytes": len(set(blob)),
        "zero_ratio": round(blob.count(0) / len(blob), 3),
    }


def split(data: bytes, table: list[int], label: str) -> list[dict]:
    out = []
    for i, off in enumerate(table):
        end = table[i + 1] if i + 1 < len(table) else len(data)
        if off >= len(data):
            out.append({"index": i, "offset": off, "error": "位移超出檔案"})
            continue
        span = data[off:end]
        magic = span[:4]
        body = span[4:]
        out.append(
            {
                "file": label,
                "index": i,
                "offset": off,
                "offset_hex": f"0x{off:X}",
                "span_length": len(span),
                "magic": magic.decode("latin1"),
                "magic_hex": magic.hex(),
                "body": describe(body),
            }
        )
    return out


def main() -> None:
    exe_p, g1_p, g2_p, out_p = (Path(p) for p in sys.argv[1:5])
    exe = exe_p.read_bytes()
    g1 = g1_p.read_bytes()
    g2 = g2_p.read_bytes()
    tables = read_tables(exe, (len(g1), len(g2)))

    directory = []
    for res_id, raw in enumerate(tables["directory"]):
        disk = raw & 0xC0
        idx = raw & 0x3F
        directory.append(
            {
                "resource_id": res_id,
                "raw": f"0x{raw:02X}",
                "disk_bits": f"0x{disk:02X}",
                "table_index": idx,
                "table": {0x80: "A", 0x40: "B"}.get(disk),
                "in_range": (
                    idx < len(tables["table_a"])
                    if disk == 0x80
                    else idx < len(tables["table_b"]) if disk == 0x40 else None
                ),
            }
        )

    payload = {
        "inputs": {
            "exe_sha256": hashlib.sha256(exe).hexdigest(),
            "game1_sha256": hashlib.sha256(g1).hexdigest(),
            "game2_sha256": hashlib.sha256(g2).hexdigest(),
            "game1_size": len(g1),
            "game2_size": len(g2),
        },
        "table_a_entries": len(tables["table_a"]),
        "table_b_entries": len(tables["table_b"]),
        "directory_entries": len(tables["directory"]),
        "directory": directory,
        "game1_resources": split(g1, tables["table_a"], "game1"),
        "game2_resources": split(g2, tables["table_b"], "game2"),
    }
    out_p.write_text(json.dumps(payload, ensure_ascii=False, indent=2), encoding="utf-8")

    for label in ("game1", "game2"):
        rows = payload[f"{label}_resources"]
        magics = {r.get("magic") for r in rows}
        print(f"{label}: {len(rows)} 個資源，開頭 magic = {magics}")


main()
