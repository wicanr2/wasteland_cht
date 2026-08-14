"""把 UTF-8 的翻譯 TSV 編譯成 Go 讀得動的 Big5 目錄（`docs/spec/11`）。

    translations/zh-Hant/*.tsv  →  translations/zh-Hant.cat

**Big5 編碼的知識放在這裡，Go 不需要依賴任何編碼函式庫。**

除了逐 key 的翻譯檔，還吃一份**文本層**的共用翻譯 `_shared.tsv`
（`<原文>\t<譯文>`）：42 個地圖區塊各自帶一份戰鬥動詞與開門訊息，
同樣的英文在 1,571 個 key 裡重複出現，逐 key 翻只會讓同一句話被翻出好幾種版本。
逐 key 的翻譯**永遠優先**，共用層只補沒被翻到的。

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


def escape_for_error(s: str) -> str:
    """報錯時把控制碼印成看得見的形式。"""
    return "".join(ch if ch.isprintable() else f"\\x{ord(ch):02X}" for ch in s)


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

    # 文本層：原文 → 譯文。用在 42 個區塊各帶一份的戰鬥動詞那類重複文本。
    shared: dict[str, str] = {}
    shared_path = zh_dir / "_shared.tsv"
    if shared_path.is_file():
        for lineno, line in enumerate(
            shared_path.read_text(encoding="utf-8").splitlines(), 1
        ):
            if not line.strip() or line.startswith("#"):
                continue
            if "\t" not in line:
                raise SystemExit(f"{shared_path}:{lineno}：沒有 TAB")
            orig, text = line.split("\t", 1)
            key = unescape(orig)
            if key in shared:
                raise SystemExit(f"{shared_path}:{lineno}：這段原文已經翻過了")
            shared[key] = unescape(text)

    # ⚠ 跨檔的重複 key 會靜靜覆蓋掉一份翻譯——同一條被翻兩次而且兩份不同時，
    # 出來的是哪一份取決於檔名排序。這種錯不會有任何症狀，所以要擋。
    zh: dict[str, str] = {}
    owner: dict[str, str] = {}
    dupes: list[str] = []
    for p in sorted(zh_dir.glob("*.tsv")):
        if p.name.startswith("_"):
            continue  # 底線開頭的是文本層，不是逐 key 的翻譯檔
        for key, text in read_tsv(p).items():
            if key in zh:
                dupes.append(f"{key}：{owner[key]} 與 {p.name} 都翻了")
                continue
            zh[key], owner[key] = text, p.name
    if dupes:
        for d in dupes:
            print("錯誤：" + d, file=sys.stderr)
        raise SystemExit(f"{len(dupes)} 個重複的 key，沒有產生 .cat")

    # 共用層只補沒被逐 key 翻到的。⚠ 原文對不上的共用項要擋下來——
    # 打錯一個字的話它會安靜地一條都套不上，而症狀就只是「進度沒動」。
    texts = set(source.values())
    if unused := [o for o in shared if o not in texts]:
        for o in unused:
            print(f"錯誤：共用翻譯的原文對不上任何字串：{escape_for_error(o)}", file=sys.stderr)
        raise SystemExit(f"{len(unused)} 條共用翻譯沒有對應的原文")
    from_shared = 0
    for key, text in source.items():
        if key in zh:
            continue
        if (t := shared.get(text)) is not None:
            zh[key], owner[key] = t, "_shared.tsv"
            from_shared += 1

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
    print(
        f"{len(entries)} 條 → {dest}（{len(out)} bytes；原文共 {len(source)} 條；"
        f"其中 {from_shared} 條來自共用文本層）"
    )


if __name__ == "__main__":
    main()
