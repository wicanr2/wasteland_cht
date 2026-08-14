"""把一張地圖用第 3 層的圖形編號畫成文字圖（`docs/re/24` §2.3）。

第 3 層每格 1 byte 是**圖形編號**，對到 `seg003:0x420 + 編號 × 128`：

    0–9   `IC0_9.WLF` 的十個 16 × 16 圖形
    ≥10   `ALLHTDS` 的圖磚，編號 ＝ 值 − 10（圖磚組載在 0x920 ＝ 0x420 ＋ 10 × 128）

每格用該圖磚出現最多的非 0 顏色代表，所以輸出是「一個字元一格」的縮圖，
不是像素級重現，也不會倒出原版影像資料。

用法：
    python3 tools/render_map.py <wl.merged.exe> <game1> <game2> <allhtds1> <allhtds2> [資源編號]
"""

from __future__ import annotations

import collections
import struct
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from decode_pic import undelta  # noqa: E402
from huffman import decompress, split_all  # noqa: E402

DS = 0x1CE20
BLOCK_TOTAL = 0xBD86
BLOCK_READ = 0xBD22
DIRECTORY = 0xBEC9
MAP_SIZE_SELECTOR = 0xBF1C
DIM_AT = 0x2C
TILESET_AT = 0x30
IC0_9_COUNT = 10  # 圖形編號 0–9 是 IC0_9.WLF，不是圖磚
HEX = "0123456789ABCDEF"


def decrypt(data: bytes, key: int) -> bytearray:
    out = bytearray(len(data))
    for i, c in enumerate(data):
        out[i] = c ^ key
        key = (key + 0x1F) & 0xFF
    return out


def tile_palette(raw: bytes) -> list[str]:
    """每張圖磚取「出現最多的非 0 顏色」當代表字元。"""
    chars = []
    for n in range(len(raw) // 128):
        tile = undelta(raw[n * 128 : (n + 1) * 128], 8)
        hist: collections.Counter = collections.Counter()
        for byte in tile:
            hist[byte >> 4] += 1
            hist[byte & 0xF] += 1
        hist.pop(0, None)
        chars.append(HEX[hist.most_common(1)[0][0]] if hist else ".")
    return chars


def main() -> None:
    if len(sys.argv) < 6:
        raise SystemExit(__doc__)
    exe_p, g1_p, g2_p, h1_p, h2_p = (Path(p) for p in sys.argv[1:6])
    want = int(sys.argv[6]) if len(sys.argv) > 6 else 0

    tilesets = []
    for path in (h1_p, h2_p):
        src = path.read_bytes()
        for block in split_all(src):
            tilesets.append(decompress(src[block["offset"] :])[0])

    exe = exe_p.read_bytes()
    header_bytes = struct.unpack_from("<H", exe, 8)[0] * 16
    at = lambda off: DS + off - 0x10000 + header_bytes  # noqa: E731
    directory: list[int] = []
    p = at(DIRECTORY)
    while exe[p] != 0xFF:
        directory.append(exe[p])
        p += 1
    total = [
        struct.unpack_from("<H", exe, at(BLOCK_TOTAL) + i * 2)[0]
        for i in range(len(directory))
    ]
    read = [
        struct.unpack_from("<H", exe, at(BLOCK_READ) + i * 2)[0]
        for i in range(len(directory))
    ]
    selector = exe[at(MAP_SIZE_SELECTOR) : at(MAP_SIZE_SELECTOR) + len(directory)]

    files = {"game1": g1_p.read_bytes(), "game2": g2_p.read_bytes()}
    cursor = {"game1": 0, "game2": 0}
    for res_id in range(len(directory)):
        label = {0x80: "game1", 0x40: "game2"}.get(directory[res_id] & 0xC0)
        if label is None or total[res_id] == 0:
            continue
        off = cursor[label]
        cursor[label] = off + total[res_id]
        if res_id != want:
            continue
        span = files[label][off : off + total[res_id]]
        checksum = struct.unpack_from("<H", span, 4)[0]
        body = decrypt(span[6 : read[res_id]], (checksum & 0xFF) ^ (checksum >> 8))
        map_size = 0x1800 if selector[res_id] == 0x40 else 0x600
        dim = body[map_size + DIM_AT]
        tileset = body[map_size + TILESET_AT]
        layer3 = decompress(span[read[res_id] :], verify_magic=False)[0]
        chars = tile_palette(tilesets[tileset])
        print(
            f"資源 #{res_id}（{label}）：{dim} × {dim}，圖磚組 {tileset}"
            f"（{len(chars)} 張）"
        )
        out_of_range = 0
        for y in range(dim):
            row = []
            for x in range(dim):
                value = layer3[y * dim + x]
                if value < IC0_9_COUNT:
                    row.append("·")  # IC0_9.WLF，不是圖磚
                elif value - IC0_9_COUNT < len(chars):
                    row.append(chars[value - IC0_9_COUNT])
                else:
                    row.append("?")
                    out_of_range += 1
            print("  " + "".join(row))
        print(f"超出圖磚組範圍的格子：{out_of_range}")
        return
    raise SystemExit(f"找不到資源編號 {want}")


if __name__ == "__main__":
    main()
