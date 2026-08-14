"""驗證地圖的三層結構（`docs/re/24`），只輸出結構與統計（不倒出內容）。

一個 MSQ 區塊裡的地圖不是單一平面，而是三層，全部以記錄區標頭 `+0x2C` 的
邊長 D（正方形）為準：

    偏移 0          第 1 層：圖磚編號，**4 bit 一格**（D × D ÷ 2 bytes）
    偏移 D×D÷2      第 2 層：每格 1 byte（D × D bytes）
    Huffman 尾段     第 3 層：每格 1 byte（D × D bytes），載到 `ds:3448h`

前兩層加起來 ＝ 記錄區起點 P（`ds:46B0h`，`0x600` 或 `0x1800`）。

用法：
    python3 tools/summarize_map_layers.py <wl.merged.exe> <game1> <game2> <out.json>
"""

from __future__ import annotations

import json
import struct
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from huffman import decompress  # noqa: E402

DS = 0x1CE20
BLOCK_TOTAL = 0xBD86  # 區塊總長度
BLOCK_READ = 0xBD22  # 載入器實際讀進來、交給 XOR 解密的長度
DIRECTORY = 0xBEC9
MAP_SIZE_SELECTOR = 0xBF1C  # 0x40 → 地圖區 0x1800，其餘 → 0x600
DIM_AT = 0x2C  # 記錄區標頭裡的邊長
TILESET_AT = 0x30  # 記錄區標頭裡的 ALLHTDS 圖磚組編號


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
    total = [
        struct.unpack_from("<H", exe, at(BLOCK_TOTAL) + i * 2)[0]
        for i in range(len(directory))
    ]
    read = [
        struct.unpack_from("<H", exe, at(BLOCK_READ) + i * 2)[0]
        for i in range(len(directory))
    ]

    files = {"game1": g1_p.read_bytes(), "game2": g2_p.read_bytes()}
    cursor = {"game1": 0, "game2": 0}
    blocks = []

    for res_id in range(len(directory)):
        label = {0x80: "game1", 0x40: "game2"}.get(directory[res_id] & 0xC0)
        if label is None or total[res_id] == 0:
            continue
        data = files[label]
        off = cursor[label]
        cursor[label] = off + total[res_id]
        span = data[off : off + total[res_id]]
        checksum = struct.unpack_from("<H", span, 4)[0]
        body = decrypt(span[6 : read[res_id]], (checksum & 0xFF) ^ (checksum >> 8))

        map_size = 0x1800 if selector[res_id] == 0x40 else 0x600
        dim = body[map_size + DIM_AT] if map_size + DIM_AT < len(body) else None
        tail_raw = span[read[res_id] :]
        tail_len = None
        if len(tail_raw) > 8:
            try:
                tail_len = len(decompress(tail_raw, verify_magic=False)[0])
            except Exception as exc:  # 尾段格式不符時記下來，不要靜靜跳過
                tail_len = f"解壓失敗：{exc}"

        entry = {
            "resource_id": res_id,
            "file": label,
            "map_size": f"0x{map_size:04X}",
            "dim": dim,
            "tileset": body[map_size + TILESET_AT] if dim is not None else None,
            "tail_length": tail_len,
        }
        if dim:
            entry["layer1_bytes"] = dim * dim // 2
            entry["layer2_bytes"] = dim * dim
            entry["layers_fill_map_area"] = dim * dim // 2 + dim * dim == map_size
            entry["tail_is_one_byte_per_cell"] = tail_len == dim * dim
        blocks.append(entry)

    checked = [b for b in blocks if b.get("dim")]
    summary = {
        "dataset": "map-layers",
        "block_count": len(blocks),
        "layers_fill_map_area": sum(1 for b in checked if b["layers_fill_map_area"]),
        "tail_is_one_byte_per_cell": sum(
            1 for b in checked if b["tail_is_one_byte_per_cell"]
        ),
        "checked": len(checked),
        "dims": sorted({b["dim"] for b in checked}),
        "tilesets": sorted({b["tileset"] for b in checked}),
        "blocks": blocks,
    }
    out_p.write_text(
        json.dumps(summary, ensure_ascii=False, indent=1), encoding="utf-8"
    )
    print(
        f"{summary['checked']} 個區塊；兩層填滿地圖區 "
        f"{summary['layers_fill_map_area']}／{summary['checked']}，"
        f"尾段每格 1 byte {summary['tail_is_one_byte_per_cell']}／{summary['checked']}；"
        f"邊長 {summary['dims']}，圖磚組 {summary['tilesets']}"
    )


if __name__ == "__main__":
    main()
