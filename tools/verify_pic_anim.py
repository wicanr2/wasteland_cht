#!/usr/bin/env python3
"""重建 `ALLPICS` 的局部動畫，並與實機截圖逐像素比對（A9 的驗收）。

`docs/re/23` §5 把 `ALLPICS` 交錯的參數區列為未解很久。這一支照
`sub_10A7A`（拆表）與 `sub_10B11`（疊格）把動畫真的**畫出來**，
再拿實機截圖驗——**讀得懂指令不等於重建得出輸出**，所以驗收是像素數而不是敘述。

解法（`docs/re/23` §5.1、§5.3）：

| 項目 | 規則 |
|---|---|
| 表 A 一筆 | `(延遲, 格號)` 交錯，第一個 byte 是**初始延遲**（第三趟 `0x10AF1` 抽走它） |
| 表 B 元素 | word 標頭 ＋ `(w >> 12) + 1` bytes 酬載 |
| 相位 | `p ← w & 3`，起點 x ← 欄 × 8 ＋ **2p**（前 p 對像素缺席，落在低位） |
| 位置 | `v ← (w >> 2) & 0x3FF`；`列, 欄 ← divmod(v, 12)` |
| 酬載 | 一個 byte 兩個像素（高 nibble 在左），**XOR** 進畫面 |
| 累積 | XOR 疊上去之後**不還原**，所以畫面 ＝ 底圖 ⊕ 格0 ⊕ … ⊕ 格k |

用法（純 stdlib）：

    python3 tools/verify_pic_anim.py workplace/orig/wastland/allpics1 3 \\
        --screen workplace/dosbox/shots/db05.ppm --at 8,8
"""

from __future__ import annotations

import argparse
import struct
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from decode_pic import undelta  # noqa: E402
from huffman import decompress, split_all  # noqa: E402

PIC_BYTES = 4032
WIDTH, HEIGHT, STRIDE = 96, 84, 48
PLANE_ROW = 12   # 一列 12 個螢幕 byte（96 ÷ 8）


def load(path: Path, index: int):
    """回傳（第 index 張圖的像素, 它後面那個參數區的原始 bytes）。"""
    data = path.read_bytes()
    blocks = split_all(data)
    seen = -1
    for i, blk in enumerate(blocks):
        if blk["size"] != PIC_BYTES:
            continue
        seen += 1
        if seen != index:
            continue
        pic = undelta(decompress(data[blk["offset"]:])[0], STRIDE)
        if i + 1 >= len(blocks) or blocks[i + 1]["size"] == PIC_BYTES:
            return pic, b""
        return pic, decompress(data[blocks[i + 1]["offset"]:])[0]
    raise SystemExit(f"{path.name} 裡沒有第 {index} 張圖")


def parse(par: bytes):
    """拆成（表 A 的播放順序, 表 B 的每一格）。"""
    if len(par) < 2:
        return [], []
    n = struct.unpack_from("<H", par, 0)[0]
    a_bytes = par[2:2 + n]

    order, start = [], 0
    for i, byte in enumerate(a_bytes):
        if byte != 0xFF:
            continue
        rec = a_bytes[start:i]
        start = i + 1
        # (延遲, 格號) 交錯：第一個 byte 是初始延遲，所以格號在奇數位
        order += [rec[k + 1] for k in range(0, len(rec) - 1, 2)]

    rest = par[2 + n:]
    frames, cur = [], []
    if len(rest) >= 2:
        m = struct.unpack_from("<H", rest, 0)[0]
        body = rest[2:2 + m]
        k = 0
        while k + 2 <= len(body):
            w = struct.unpack_from("<H", body, k)[0]
            if w == 0xFFFF:
                frames.append(cur)
                cur = []
                k += 2
                continue
            ln = (w >> 12) + 1
            row, col = divmod((w >> 2) & 0x3FF, PLANE_ROW)
            cur.append((col * 8 + 2 * (w & 3), row, body[k + 2:k + 2 + ln]))
            k += 2 + ln
    if cur:
        frames.append(cur)
    return order, frames


def to_pixels(buf: bytes) -> list[list[int]]:
    return [[(buf[y * STRIDE + (x >> 1)] >> 4) if (x & 1) == 0
             else (buf[y * STRIDE + (x >> 1)] & 0x0F)
             for x in range(WIDTH)] for y in range(HEIGHT)]


def apply_frame(px: list[list[int]], frame) -> None:
    for x0, y0, payload in frame:
        if not 0 <= y0 < HEIGHT:
            continue
        for j, byte in enumerate(payload):
            for half, nib in ((0, byte >> 4), (1, byte & 0x0F)):
                x = x0 + j * 2 + half
                if 0 <= x < WIDTH:
                    px[y0][x] ^= nib


def main() -> None:
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("allpics")
    ap.add_argument("index", type=int)
    ap.add_argument("--screen", help="實機截圖（PPM）")
    ap.add_argument("--at", default="8,8", help="圖在畫面上的左上角，預設 8,8")
    args = ap.parse_args()

    pic, par = load(Path(args.allpics), args.index)
    order, frames = parse(par)
    print(f"表 A 播放順序：{order}")
    print(f"表 B 格數：{len(frames)}（元素數 {[len(f) for f in frames]}）")

    px = to_pixels(pic)
    if not args.screen:
        return

    from compare_screen import read_ppm, to_index  # 只有比對時才需要
    ox, oy = (int(v) for v in args.at.split(","))
    sw, _, raw = read_ppm(Path(args.screen))
    scr = {}
    for y in range(HEIGHT):
        for x in range(WIDTH):
            o = ((oy + y) * sw + ox + x) * 3
            scr[(x, y)] = to_index((raw[o], raw[o + 1], raw[o + 2]))

    def diff() -> int:
        return sum(1 for y in range(HEIGHT) for x in range(WIDTH)
                   if scr[(x, y)] != px[y][x])

    print(f"底圖（一格都沒疊）：差 {diff()} 像素")
    best = (diff(), -1)
    for step, fi in enumerate(order):
        if fi >= len(frames):
            print(f"  第 {step} 步：格 {fi} 不存在，跳過")
            continue
        apply_frame(px, frames[fi])
        d = diff()
        print(f"  疊到第 {step} 步（格 {fi}）：差 {d} 像素")
        if d < best[0]:
            best = (d, step)
    print(f"\n最小差異：第 {best[1]} 步、{best[0]} 像素")
    sys.exit(0 if best[0] == 0 else 1)


if __name__ == "__main__":
    main()
