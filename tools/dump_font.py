"""把原版的兩套字型畫成文字圖，用來確定字元集的排列順序。

原版有兩套字型、兩套索引空間（`docs/re/14`）：

`colour`（`colorf.fnt`，獨立檔案）
    每個字模 32 bytes ＝ 4 個 EGA 平面 × 8 個像素列，平面連續存放
    （plane0 的 8 列、plane1 的 8 列…）。一個像素的顏色是四個平面同位置的
    位元依 plane0＝bit0 組起來的 4-bit 值。格式來自 overlay slot 19 的畫字元迴圈。

`mono`（內嵌在 `wl.exe`，`seg003:0xCA60`，不是獨立檔案）
    每個字模 8 bytes ＝ 8 個像素列，1 bit 一像素，索引 ＝ ASCII − 0x20。
    格式來自 `sub_1060C`（overlay slot 5）。

用法：
    python3 tools/dump_font.py colour workplace/orig/wastland/colorf.fnt [輸出.txt]
    python3 tools/dump_font.py mono workplace/analysis/unpacked/wl.merged.exe [輸出.txt]
"""

from __future__ import annotations

import sys
from pathlib import Path

ROWS = 8
PLANES = 4
HEX = "0123456789ABCDEF"

# mono 字型在解包映像裡的位置：seg003 起於線性 0x2AE20，字型在 seg003+0xCA60，
# 而線性 → 檔案位移要扣掉載入基底 0x10000 再加上 MZ header（e_cparhdr × 16）。
MONO_LINEAR = 0x2AE20 + 0xCA60
MONO_COUNT = 128  # sub_1060C 的 `and ax, 7Fh`


def decode_colour(raw: bytes) -> list[list[int]]:
    rows = []
    for y in range(ROWS):
        rows.append(
            [
                sum(((raw[p * ROWS + y] >> (7 - x)) & 1) << p for p in range(PLANES))
                for x in range(8)
            ]
        )
    return rows


def decode_mono(raw: bytes) -> list[list[int]]:
    return [[(raw[y] >> (7 - x)) & 1 for x in range(8)] for y in range(ROWS)]


def strip(glyphs: list[list[list[int]]], labels: list[str], palette: str) -> list[str]:
    out = ["  ".join(f"{lab:<10}" for lab in labels).rstrip()]
    for y in range(ROWS):
        out.append(
            "   "
            + "   ".join(
                "".join("." if c == 0 else palette[c % len(palette)] for c in g[y])
                for g in glyphs
            )
        )
    out.append("")
    return out


def render_colour(path: Path) -> str:
    data = path.read_bytes()
    if len(data) % 32:
        raise SystemExit(f"{path} 長度 {len(data)} 不是 32 的倍數")
    count = len(data) // 32
    out = [f"{path.name}：{len(data)} bytes ＝ {count} 個字模（32 bytes、8×8、4 平面）", ""]
    for base in range(0, count, 8):
        group = list(range(base, min(base + 8, count)))
        out += strip(
            [decode_colour(data[g * 32 : (g + 1) * 32]) for g in group],
            [f"{g:>3} 0x{g:02X}" for g in group],
            HEX,
        )
    return "\n".join(out)


def render_mono(path: Path) -> str:
    data = path.read_bytes()
    if data[:2] != b"MZ":
        raise SystemExit(f"{path} 不是 MZ 執行檔——mono 字型要從解包映像取")
    header = int.from_bytes(data[8:10], "little") * 16
    offset = MONO_LINEAR - 0x10000 + header
    font = data[offset : offset + MONO_COUNT * ROWS]
    out = [
        f"{path.name}：seg003+0xCA60（線性 0x{MONO_LINEAR:05X}、檔案位移 0x{offset:05X}）",
        f"{MONO_COUNT} 個字模 × 8 bytes，索引 ＝ ASCII − 0x20",
        "",
    ]
    for base in range(0, MONO_COUNT, 8):
        group = list(range(base, base + 8))
        labels = []
        for g in group:
            code = 0x20 + g
            char = chr(code) if 0x20 <= code < 0x7F else " "
            labels.append(f"{g:>3} {code:02X} {char!r}")
        out += strip(
            [decode_mono(font[g * ROWS : (g + 1) * ROWS]) for g in group], labels, "#"
        )
    return "\n".join(out)


def main() -> None:
    if len(sys.argv) < 3 or sys.argv[1] not in ("colour", "mono"):
        raise SystemExit(__doc__)
    mode, source = sys.argv[1], Path(sys.argv[2])
    text = render_colour(source) if mode == "colour" else render_mono(source)
    if len(sys.argv) > 3:
        Path(sys.argv[3]).write_text(text, encoding="utf-8")
        print(f"→ {sys.argv[3]}")
    else:
        print(text)


if __name__ == "__main__":
    main()
