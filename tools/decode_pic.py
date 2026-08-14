"""解碼原版的圖片格式（`docs/re/23`）。

原版的圖片是同一套格式，只有尺寸不同：

    packed 4bpp（一個 byte 兩個像素，高位在左）
    列間 XOR delta：out[n + stride] = in[n + stride] XOR out[n]
    **XOR 的回看距離就是一列的 byte 數**

| 來源 | stride | 尺寸 | 解碼函式 |
|---|---:|---|---|
| `ALLPICS1/2` 的 4,032 bytes 子區塊 | 48 | 96 × 84 | overlay slot 2（`sub_10144`） |
| `TITLE.PIC`（18,432 bytes） | 144 | 288 × 128 | `start` 內嵌（`docs/re/03` §6） |

只輸出文字圖與統計，不倒出原版影像資料。

用法：
    python3 tools/decode_pic.py title  workplace/orig/wastland/title.pic
    python3 tools/decode_pic.py allpics workplace/orig/wastland/allpics1 [子區塊編號]
"""

from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from huffman import decompress, split_all  # noqa: E402

HEX = "0123456789ABCDEF"
PROFILES = {
    "allpics": {"stride": 48, "width": 96, "height": 84},
    "title": {"stride": 144, "width": 288, "height": 128},
    "tile": {"stride": 8, "width": 16, "height": 16},
}


def undelta(buf: bytes, stride: int) -> bytearray:
    """照原版：以 word 為單位，`out[n+stride] ^= out[n]`，n 由 0 開始每次 +2。

    因為 n ≥ stride 之後讀到的是已經解過的內容，所以這是滾動的自參考解碼，
    不是單純的 XOR——順序不能顛倒。
    """
    out = bytearray(buf)
    n = 0
    while n + stride + 1 < len(out):
        out[n + stride] ^= out[n]
        out[n + stride + 1] ^= out[n + 1]
        n += 2
    return out


def render(buf: bytes, width: int, height: int, stride: int, cols: int = 96) -> list[str]:
    """把 4bpp 的像素畫成文字圖；`cols` 是輸出寬度（過寬會橫向取樣）。"""
    step = max(1, width // cols)
    lines = []
    for y in range(height):
        row = []
        for x in range(0, width, step):
            byte = buf[y * stride + (x >> 1)]
            value = (byte >> 4) if (x & 1) == 0 else (byte & 0xF)
            row.append("." if value == 0 else HEX[value])
        lines.append("".join(row))
    return lines


def main() -> None:
    if len(sys.argv) < 3 or sys.argv[1] not in PROFILES:
        raise SystemExit(__doc__)
    profile = PROFILES[sys.argv[1]]
    source = Path(sys.argv[2]).read_bytes()

    if sys.argv[1] == "tile":
        blocks = split_all(source)
        block = int(sys.argv[3]) if len(sys.argv) > 3 else 0
        first = int(sys.argv[4]) if len(sys.argv) > 4 else 0
        raw, _ = decompress(source[blocks[block]["offset"] :])
        print(
            f"{len(blocks)} 個子區塊；第 {block} 塊 {len(raw)} bytes ＝ "
            f"{len(raw) // 128} 個 16 × 16 圖磚"
        )
        for n in range(first, min(first + 4, len(raw) // 128)):
            tile = undelta(raw[n * 128 : (n + 1) * 128], 8)
            print(f"\n  圖磚 #{n}")
            for line in render(tile, 16, 16, 8, cols=16):
                print("  " + line)
        return

    if sys.argv[1] == "allpics":
        blocks = split_all(source)
        pictures = [b for b in blocks if b["size"] == 4032]
        index = int(sys.argv[3]) if len(sys.argv) > 3 else 0
        print(
            f"{len(blocks)} 個子區塊，其中 {len(pictures)} 個是 4,032 bytes 的圖片；"
            f"顯示第 {index} 張"
        )
        raw, _ = decompress(source[pictures[index]["offset"] :])
    else:
        raw = source

    stride = profile["stride"]
    expected = stride * profile["height"]
    if len(raw) != expected:
        raise SystemExit(
            f"長度 {len(raw)} 與 {profile['width']}×{profile['height']}"
            f"（{expected} bytes）不符"
        )

    pixels = undelta(raw, stride)
    from collections import Counter

    hist = Counter()
    for byte in pixels:
        hist[byte >> 4] += 1
        hist[byte & 0xF] += 1
    used = sorted(k for k, v in hist.items() if v)
    print(
        f"{profile['width']} × {profile['height']}，4bpp，"
        f"用到 {len(used)} 種顏色：{used}"
    )
    print()
    for line in render(pixels, profile["width"], profile["height"], stride):
        print("  " + line)


if __name__ == "__main__":
    main()
