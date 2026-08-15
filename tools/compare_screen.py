#!/usr/bin/env python3
"""把 DOSBox 的截圖與我們解碼器的輸出**逐像素**比對。

這是 `CLAUDE.md` §1 說的那件事：DOSBox 是 IDA 的驗證工具。
解碼器解出來的東西要拿原版畫面對過，才算真的對。

    python3 tools/compare_screen.py title <截圖.ppm> <title.pic>
    python3 tools/compare_screen.py icons <截圖.ppm> <ic0_9.wlf> <masks.wlf>

截圖用 PPM（P6）——純 stdlib 讀得動。PNG 先轉一次：

    docker run --rm -v "$PWD/workplace/dosbox/shots:/shots" \\
        --entrypoint sh wasteland-dosbox:latest \\
        -c 'convert /shots/X.png -crop 320x200+0+0 +repage /shots/X.ppm'

比對的是**調色盤索引**不是 RGB：畫面上的每個像素對到最近的 EGA 顏色，
再與解碼器的 4-bit 索引比。這樣避免縮放與色彩空間的雜訊。

⚠ 位置是**掃出來的**，不是寫死的。畫面上的圖在哪一格由搜尋決定，
並且會把最佳與次佳的差距印出來——**差距不夠大就代表沒對準**，
那時候「吻合率 99%」是假的。
"""

from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from decode_pic import undelta  # noqa: E402

# mode 0Dh 的預設 EGA 調色盤（docs/re/23 §5：沒找到設定調色盤的程式碼，
# 推測就是這一組。這支工具正好也在驗證那個推測）。
EGA = [
    (0x00, 0x00, 0x00), (0x00, 0x00, 0xAA), (0x00, 0xAA, 0x00), (0x00, 0xAA, 0xAA),
    (0xAA, 0x00, 0x00), (0xAA, 0x00, 0xAA), (0xAA, 0x55, 0x00), (0xAA, 0xAA, 0xAA),
    (0x55, 0x55, 0x55), (0x55, 0x55, 0xFF), (0x55, 0xFF, 0x55), (0x55, 0xFF, 0xFF),
    (0xFF, 0x55, 0x55), (0xFF, 0x55, 0xFF), (0xFF, 0xFF, 0x55), (0xFF, 0xFF, 0xFF),
]


def read_ppm(path: Path) -> tuple[int, int, bytes]:
    """讀 P6 PPM。標頭的空白與註解都要處理，不要假設它只有一行。"""
    data = path.read_bytes()
    if not data.startswith(b"P6"):
        raise SystemExit(f"{path} 不是 P6 PPM")
    fields, at = [], 2
    while len(fields) < 3:
        while at < len(data) and data[at : at + 1].isspace():
            at += 1
        if data[at : at + 1] == b"#":
            while at < len(data) and data[at] != 0x0A:
                at += 1
            continue
        start = at
        while at < len(data) and not data[at : at + 1].isspace():
            at += 1
        fields.append(int(data[start:at]))
    at += 1  # 標頭後的單一空白
    w, h, maxval = fields
    if maxval != 255:
        raise SystemExit(f"只支援 maxval 255，這份是 {maxval}")
    return w, h, data[at : at + w * h * 3]


def to_index(rgb: tuple[int, int, int]) -> int:
    best, bestd = 0, 1 << 30
    for i, c in enumerate(EGA):
        d = sum((a - b) ** 2 for a, b in zip(rgb, c))
        if d < bestd:
            best, bestd = i, d
    return best


def decode_title(path: Path) -> list[list[int]]:
    raw = undelta(path.read_bytes(), 144)
    out = []
    for y in range(128):
        row = []
        for x in range(288):
            b = raw[y * 144 + (x >> 1)]
            row.append((b >> 4) if (x & 1) == 0 else (b & 0x0F))
        out.append(row)
    return out


def find_icons(shot: Path, icons_path: Path, masks_path: Path) -> None:
    """在截圖裡找出十張疊圖各出現在哪裡。

    判準是原版的合成式 `螢幕 ← (背景 AND 遮罩) OR 疊圖`：**遮罩為 0 的像素
    背景被清掉**，螢幕上就等於疊圖本身。所以只比遮罩 0 的那些像素要不要全等，
    完全不必知道底下是什麼地形——這正是「疊圖畫在哪」可以脫離地圖單獨驗的原因。

    ⚠ 0 號的遮罩**全 0、圖形也全 0**——它的作用是把一格塗黑。可比像素有 256 個
    但全部要求「黑」，所以畫面上任何一塊 16 × 16 純黑都會吻合。它的命中數
    **不能當成「這裡有 0 號」**，報出來只是為了不讓它靜靜消失。
    """
    sys.path.insert(0, str(Path(__file__).resolve().parent))
    from dump_icons import mask_bits, planar_to_indexed  # noqa: PLC0415

    w, h, pix = read_ppm(shot)
    idx = [
        [to_index((pix[(y * w + x) * 3], pix[(y * w + x) * 3 + 1], pix[(y * w + x) * 3 + 2]))
         for x in range(w)]
        for y in range(h)
    ]
    raw_i = icons_path.read_bytes()
    raw_m = masks_path.read_bytes()
    n = len(raw_i) // 128
    print(f"截圖 {w} × {h}；疊圖 {n} 張")

    for k in range(n):
        img = planar_to_indexed(raw_i[k * 128 : (k + 1) * 128])
        msk = mask_bits(raw_m[k * 32 : (k + 1) * 32])
        want = [(x, y, img[y][x]) for y in range(16) for x in range(16) if not msk[y][x]]
        if not want:
            print(f"疊圖 {k}：遮罩全 1，沒有可比的像素——跳過（不是「沒出現」）")
            continue
        hits = [
            (ox, oy)
            for oy in range(h - 16 + 1)
            for ox in range(w - 16 + 1)
            if all(idx[oy + y][ox + x] == v for x, y, v in want)
        ]
        print(f"疊圖 {k}（{len(want)} 個可比像素）：{len(hits)} 處吻合 {hits[:8]}")


def main() -> None:
    if len(sys.argv) == 5 and sys.argv[1] == "icons":
        find_icons(Path(sys.argv[2]), Path(sys.argv[3]), Path(sys.argv[4]))
        return
    if len(sys.argv) != 4 or sys.argv[1] != "title":
        raise SystemExit(__doc__)
    shot = Path(sys.argv[2])
    w, h, pix = read_ppm(shot)
    want = decode_title(Path(sys.argv[3]))
    ph, pw = len(want), len(want[0])
    print(f"截圖 {w} × {h}；解碼器 {pw} × {ph}")

    # 先把整張截圖換成索引，省得每個 offset 重算。
    idx = [
        [to_index((pix[(y * w + x) * 3], pix[(y * w + x) * 3 + 1], pix[(y * w + x) * 3 + 2]))
         for x in range(w)]
        for y in range(h)
    ]

    scores: list[tuple[int, int, int]] = []
    for oy in range(0, min(40, h - ph + 1)):
        for ox in range(0, min(40, w - pw + 1)):
            hit = 0
            for y in range(0, ph, 2):  # 先粗掃，每兩列取一列
                r_i, r_w = idx[oy + y], want[y]
                hit += sum(1 for x in range(pw) if r_i[ox + x] == r_w[x])
            scores.append((hit, ox, oy))
    scores.sort(reverse=True)
    if not scores:
        raise SystemExit("截圖比圖片還小，沒得比")

    _, ox, oy = scores[0]
    total = pw * ph
    hit = sum(
        1
        for y in range(ph)
        for x in range(pw)
        if idx[oy + y][ox + x] == want[y][x]
    )
    print(f"最佳位置 ({ox}, {oy})：{hit}/{total} 吻合 = {hit / total * 100:.2f}%")

    # ⚠ 沒有這一段的話，「對不準」與「解碼錯」長得一模一樣。
    if len(scores) > 1:
        second = max(s for s in scores if (s[1], s[2]) != (ox, oy))
        gap = scores[0][0] - second[0]
        print(f"次佳位置 ({second[1]}, {second[2]}) 差 {gap} 個像素（粗掃）")
        if gap < pw:  # 少於一列的差距 ＝ 沒有明顯的最佳解
            print("⚠ 最佳與次佳差距太小 —— 這代表**沒對準**，上面的吻合率不可信")

    if hit != total:
        bad = [
            (x, y, idx[oy + y][ox + x], want[y][x])
            for y in range(ph)
            for x in range(pw)
            if idx[oy + y][ox + x] != want[y][x]
        ]
        print(f"不吻合 {len(bad)} 個像素，前 10 個（x, y, 畫面, 解碼器）：")
        for b in bad[:10]:
            print("   ", b)


if __name__ == "__main__":
    main()
