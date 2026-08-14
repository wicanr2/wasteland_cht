"""把 UTF-8 的翻譯 TSV 編譯成 Go 讀得動的 Big5 目錄（`docs/spec/11`）。

    translations/zh-Hant/*.tsv  →  translations/zh-Hant.cat

**Big5 編碼的知識放在這裡，Go 不需要依賴任何編碼函式庫。**

編譯時擋三種錯（`docs/spec/11` §5）：

1. **格數超過原文**——訊息視窗只有 6 行 × 38 格，超過就會爆版面。
2. **控制碼數量不符**——`\\x0A`／`\\x0C`／`\\x0E`／`\\x0F` 是文字變形的分段，
   少一個就會選錯段（`docs/re/28`）。
3. **Big5 缺字**——倚天字型是 Big5 排列，編不出來的字要換掉。

用法：
    python3 tools/build_lang.py [translations 目錄]
"""

from __future__ import annotations

import struct
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from extract_strings import unescape  # noqa: E402

MAGIC = b"WLCAT\0"
VERSION = 1

# 這些控制碼是文字變形的分段，數量必須與原文一致（docs/re/28）。
VARIANT_CODES = (0x0A, 0x0C, 0x0E, 0x0F)

# 少數全形符號在 Python 的 big5 codec 與 Big5 表之間有歧義，手動補。
MANUAL_BIG5 = {"～": b"\xa1\xe3"}


def cells(text: str) -> int:
    """算排版格數。中文與英數都是一格（docs/spec/10 §3），控制碼不算。"""
    return sum(1 for ch in text if ord(ch) >= 0x20 and ch != "\x7f")


def to_big5(text: str) -> tuple[bytes, list[str]]:
    """轉成 Big5，回傳（bytes, 編不出來的字）。控制碼與 ASCII 原樣保留。"""
    out = bytearray()
    missing: list[str] = []
    for ch in text:
        if ord(ch) < 0x80:
            out.append(ord(ch))
            continue
        if ch in MANUAL_BIG5:
            out += MANUAL_BIG5[ch]
            continue
        try:
            out += ch.encode("big5")
        except UnicodeEncodeError:
            missing.append(ch)
    return bytes(out), missing


def read_tsv(path: Path) -> dict[str, str]:
    rows: dict[str, str] = {}
    for lineno, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        if not line.strip() or line.startswith("#"):
            continue
        if "\t" not in line:
            raise SystemExit(f"{path}:{lineno}：沒有 TAB，這一行不是 key<TAB>text")
        key, text = line.split("\t", 1)
        if key in rows:
            raise SystemExit(f"{path}:{lineno}：key {key} 重複")
        rows[key] = unescape(text)
    return rows


def main() -> None:
    root = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("translations")
    src_dir, zh_dir = root / "source", root / "zh-Hant"
    if not zh_dir.is_dir():
        raise SystemExit(f"{zh_dir} 不存在——還沒有任何翻譯")

    source: dict[str, str] = {}
    for p in sorted(src_dir.glob("*.tsv")):
        source.update(read_tsv(p))

    zh: dict[str, str] = {}
    for p in sorted(zh_dir.glob("*.tsv")):
        zh.update(read_tsv(p))

    errors: list[str] = []
    entries: list[tuple[str, bytes]] = []
    for key, text in sorted(zh.items()):
        orig = source.get(key)
        if orig is None:
            errors.append(f"{key}：原文裡沒有這個 key（抽取工具改過？）")
            continue
        if cells(text) > cells(orig):
            errors.append(
                f"{key}：譯文 {cells(text)} 格 > 原文 {cells(orig)} 格，訊息視窗會爆"
            )
        for code in VARIANT_CODES:
            a, b = orig.count(chr(code)), text.count(chr(code))
            if a != b:
                errors.append(f"{key}：控制碼 \\x{code:02X} 原文 {a} 個、譯文 {b} 個")
        data, missing = to_big5(text)
        if missing:
            errors.append(f"{key}：Big5 編不出 {''.join(sorted(set(missing)))}")
            continue
        entries.append((key, data))

    if errors:
        for e in errors:
            print("錯誤：" + e, file=sys.stderr)
        raise SystemExit(f"{len(errors)} 個問題，沒有產生 .cat")

    out = bytearray(MAGIC)
    out += struct.pack("<HI", VERSION, len(entries))
    for key, data in entries:
        kb = key.encode("utf-8")
        out += struct.pack("<H", len(kb)) + kb
        out += struct.pack("<H", len(data)) + data

    dest = root / "zh-Hant.cat"
    dest.write_bytes(bytes(out))
    print(f"{len(entries)} 條 → {dest}（{len(out)} bytes；原文共 {len(source)} 條）")


if __name__ == "__main__":
    main()
