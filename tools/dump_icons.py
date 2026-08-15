#!/usr/bin/env python3
"""把 `IC0_9.WLF` 的十張疊圖與 `MASKS.WLF` 的十張遮罩畫成一張 PNG。

為什麼要有這支：`docs/re/24` §2.3 把這張表的**位置與筆數**算清楚了
（`seg003:0x0420`，10 × 128 bytes，4 平面 16 × 16），但「第 N 張畫的是什麼」
程式碼裡沒有寫——它只有編號。要回答那一題只能**把圖畫出來看**，
再拿程式碼裡傳那個編號的地方（`tools/scan_icon_users.py`）去綁語意。

輸出三列，每列十格，左到右是編號 0–9：

| 列 | 內容 |
|---|---|
| 上 | 疊圖本身（EGA 16 色） |
| 中 | 遮罩（白 ＝ 位元 1 ＝ 保留背景） |
| 下 | 合成結果：`(背景 AND 遮罩) OR 疊圖`，背景用格紋代表任意地形 |

⚠ 下面那一列**不是**原版畫面，是用原版的合成公式重畫的示意；
真正的驗收是 `tools/compare_screen.py` 的逐像素對拍。

用法（純 stdlib）：
    python3 tools/dump_icons.py <ic0_9.wlf> <masks.wlf> <輸出.png>
"""

from __future__ import annotations

import struct
import sys
import zlib
from pathlib import Path

W = H = 16
TILE_BYTES = 128
MASK_BYTES = 32
SCALE = 6
GAP = 2

# mode 0Dh 的預設 EGA 調色盤（與 tools/compare_screen.py 同一組）。
EGA = [
    (0x00, 0x00, 0x00), (0x00, 0x00, 0xAA), (0x00, 0xAA, 0x00), (0x00, 0xAA, 0xAA),
    (0xAA, 0x00, 0x00), (0xAA, 0x00, 0xAA), (0xAA, 0x55, 0x00), (0xAA, 0xAA, 0xAA),
    (0x55, 0x55, 0x55), (0x55, 0x55, 0xFF), (0x55, 0xFF, 0x55), (0x55, 0xFF, 0xFF),
    (0xFF, 0x55, 0x55), (0xFF, 0x55, 0xFF), (0xFF, 0xFF, 0x55), (0xFF, 0xFF, 0xFF),
]


def planar_to_indexed(raw: bytes) -> list[list[int]]:
    """EGA 4 平面（平面連續）→ 索引圖。與 internal/assets/pic.go 同一套。"""
    row_bytes = W // 8
    plane = row_bytes * H
    out = []
    for y in range(H):
        row = []
        for x in range(W):
            v = 0
            for p in range(4):
                b = raw[p * plane + y * row_bytes + (x >> 3)]
                v |= ((b >> (7 - (x & 7))) & 1) << p
            row.append(v)
        out.append(row)
    return out


def mask_bits(raw: bytes) -> list[list[int]]:
    """一個位元一個像素，16 列 × 2 bytes。1 ＝ 保留背景。"""
    return [
        [(raw[y * 2 + (x >> 3)] >> (7 - (x & 7))) & 1 for x in range(W)]
        for y in range(H)
    ]


def write_png(path: Path, pix: list[list[tuple[int, int, int]]]) -> None:
    h, w = len(pix), len(pix[0])
    raw = b"".join(
        b"\x00" + bytes(v for px in row for v in px) for row in pix
    )

    def chunk(tag: bytes, body: bytes) -> bytes:
        return (
            struct.pack(">I", len(body))
            + tag
            + body
            + struct.pack(">I", zlib.crc32(tag + body) & 0xFFFFFFFF)
        )

    path.write_bytes(
        b"\x89PNG\r\n\x1a\n"
        + chunk(b"IHDR", struct.pack(">IIBBBBB", w, h, 8, 2, 0, 0, 0))
        + chunk(b"IDAT", zlib.compress(raw, 9))
        + chunk(b"IEND", b"")
    )


def ascii_art(img: list[list[int]]) -> list[str]:
    """給文件用的字元圖：`.` ＝ 顏色 0，其餘印十六進位的顏色編號。"""
    return ["".join("." if v == 0 else f"{v:X}" for v in row) for row in img]


def main() -> None:
    if len(sys.argv) != 4:
        sys.exit(__doc__)
    icons_raw = Path(sys.argv[1]).read_bytes()
    masks_raw = Path(sys.argv[2]).read_bytes()
    out = Path(sys.argv[3])

    n = len(icons_raw) // TILE_BYTES
    if len(masks_raw) // MASK_BYTES != n:
        sys.exit(
            f"疊圖 {n} 張、遮罩 {len(masks_raw) // MASK_BYTES} 張——數量不一致，"
            "先確認檔案，不要硬畫"
        )

    icons = [planar_to_indexed(icons_raw[i * TILE_BYTES:(i + 1) * TILE_BYTES]) for i in range(n)]
    masks = [mask_bits(masks_raw[i * MASK_BYTES:(i + 1) * MASK_BYTES]) for i in range(n)]

    cell = W * SCALE
    sheet_w = n * cell + (n + 1) * GAP
    sheet_h = 3 * cell + 4 * GAP
    canvas = [[(0x20, 0x20, 0x20)] * sheet_w for _ in range(sheet_h)]

    def blit(col: int, row: int, get) -> None:
        x0 = GAP + col * (cell + GAP)
        y0 = GAP + row * (cell + GAP)
        for y in range(cell):
            for x in range(cell):
                canvas[y0 + y][x0 + x] = get(x // SCALE, y // SCALE)

    for i in range(n):
        blit(i, 0, lambda x, y, i=i: EGA[icons[i][y][x]])
        blit(i, 1, lambda x, y, i=i: (0xFF, 0xFF, 0xFF) if masks[i][y][x] else (0, 0, 0))

        def composite(x, y, i=i):
            # 背景用格紋代表「任意地形」，才看得出遮罩挖掉了哪一塊。
            bg = 6 if ((x >> 2) + (y >> 2)) & 1 else 2
            v = (bg if masks[i][y][x] else 0) | icons[i][y][x]
            return EGA[v & 0x0F]

        blit(i, 2, composite)

    write_png(out, canvas)
    print(f"→ {out}（{n} 張，上：疊圖／中：遮罩／下：合成示意）")

    for i in range(n):
        cover = sum(1 for row in masks[i] for b in row if not b)
        ink = sum(1 for row in icons[i] for v in row if v)
        print(f"\n# 疊圖 {i}：遮罩挖掉 {cover}/256 像素，圖形非零 {ink}/256 像素")
        for a, b in zip(ascii_art(icons[i]), ("".join("#" if v else " " for v in r) for r in masks[i])):
            print(f"  {a}  |{b}|")


if __name__ == "__main__":
    main()
