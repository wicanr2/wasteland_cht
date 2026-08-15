#!/usr/bin/env python3
"""把 `ALLPICS` 交錯其中的參數區拆成 `sub_10A7A` 的那兩張表。

`docs/re/23` §5 把這一區列為未解很久。`sub_10A7A`（overlay slot 16）
把它拆成兩張變動長度的記錄表，這支照那個拆法拆一遍：

| 表 | 分隔 | 一筆長什麼樣 |
|---|---|---|
| A | `0xFF` | 第一個 byte 是 tag，其餘是 byte 串 |
| B | `0xFFFF` | 一個 word 標頭 ＋ `(word >> 12) + 1` bytes 的酬載 |

表 B 的 word 低 12 位是**圖片緩衝區裡的 byte 位移**（一列 48 bytes、
一個 byte 兩個像素），所以 `(x, y) = (位移 % 48 × 2, 位移 ÷ 48)`。

⚠ **這支只拆結構，不解語意。** 「這些 bytes 怎麼疊回圖上」還沒解
（`docs/re/23` §5.1 記了試過哪幾種、差多少），所以輸出裡不要出現
「動畫格」以外的推論。

用法（純 stdlib）：
    python3 tools/summarize_picparams.py <allpics1|allpics2> [輸出.md]
"""

from __future__ import annotations

import struct
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from decode_pic import decompress, split_all  # noqa: E402

PIC_BYTES = 4032  # 96 × 84 packed 4bpp
STRIDE = 48       # 一列的 byte 數


def parse(raw: bytes):
    """拆成 (表 A 的記錄, 表 B 的元素, 分隔數)。"""
    if len(raw) < 2:
        return [], [], 0
    n = struct.unpack_from("<H", raw, 0)[0]
    a_bytes = raw[2 : 2 + n]

    a, start = [], 0
    for i, b in enumerate(a_bytes):
        if b == 0xFF:
            a.append(a_bytes[start:i])
            start = i + 1

    rest = raw[2 + n :]
    b_list, breaks = [], 0
    if len(rest) >= 2:
        m = struct.unpack_from("<H", rest, 0)[0]
        body = rest[2 : 2 + m]
        k = 0
        while k + 2 <= len(body):
            w = struct.unpack_from("<H", body, k)[0]
            if w == 0xFFFF:
                breaks += 1
                k += 2
                continue
            ln = (w >> 12) + 1
            b_list.append((w & 0x0FFF, body[k + 2 : k + 2 + ln]))
            k += 2 + ln
    return a, b_list, breaks


def main() -> None:
    if len(sys.argv) not in (2, 3):
        sys.exit(__doc__)
    src = Path(sys.argv[1])
    data = src.read_bytes()
    blocks = split_all(data)

    rows = [
        f"# `{src.name}` 的參數區（工具輸出，不含推論）",
        "",
        "拆法照 `sub_10A7A`（`docs/re/23` §5.1）。**只拆結構，不解語意。**",
        "",
        "| 圖 | 參數區塊 | 解出 bytes | 表 A 筆數 | 表 B 元素 | 分隔 | 表 B 的 (x, y) 範圍 |",
        "|---:|---:|---:|---:|---:|---:|---|",
    ]
    pic_index = 0
    for i, blk in enumerate(blocks):
        if blk["size"] == PIC_BYTES:
            pic_index += 1
            continue
        try:
            raw, _ = decompress(data[blk["offset"] :])
        except Exception as exc:  # 解不開要看得見，不要靜靜跳過
            rows.append(f"| {pic_index - 1} | {i} | 解壓失敗：{exc} | | | | |")
            continue
        a, b, breaks = parse(raw)
        if b:
            xs = [(off % STRIDE * 2, off // STRIDE) for off, _ in b]
            span = (f"x {min(x for x, _ in xs)}–{max(x for x, _ in xs)}、"
                    f"y {min(y for _, y in xs)}–{max(y for _, y in xs)}")
        else:
            span = "—"
        rows.append(
            f"| {pic_index - 1} | {i} | {len(raw)} | {len(a)} | {len(b)} | {breaks} | {span} |"
        )

    rows += [
        "",
        "表 B 的 (x, y) 範圍是**圖片內座標**（一列 48 bytes、一個 byte 兩個像素）。",
        "範圍窄表示那張圖只有一小塊會變——與實機對拍的殘差位置一致"
        "（`docs/re/54` §3）。",
    ]
    text = "\n".join(rows) + "\n"
    if len(sys.argv) == 3:
        Path(sys.argv[2]).write_text(text, encoding="utf-8")
        print(f"→ {sys.argv[2]}")
    else:
        print(text)


if __name__ == "__main__":
    main()
