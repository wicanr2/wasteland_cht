"""把段落書的中文翻譯編成遊戲內手札用的目錄（`docs/spec/19` §5）。

    translations/zh-Hant/paragraphs/*.md  →  translations/paragraphs-zh-Hant.cat

沿用 `tools/build_lang.py` 的 `.cat` 格式與 Big5 編碼，key 是 `para:<編號>`，
所以 Go 那邊不必多一種讀檔器（`internal/lang` 直接讀）。

來源是 markdown 而不是 TSV：段落是好幾百字的散文，塞進單行 TSV 沒辦法審閱。
檔案結構鏡射英文原文 `docs/paragraphs/`，段落標題一律 `## 段落 N`。
段落之間的空行轉成兩個 `\\x0D`（手札的換行與遊戲其餘文字一致）。

擋三種錯：

1. **編號重複或超出 1–162**——手札要湊得齊 162 段，少一段不能靜靜過去。
2. **Big5 編不出來的字**——倚天字型是 Big5 排列。
3. **段落是空的**——只有標題沒有正文。

用法：
    python3 tools/build_paragraphs.py
"""

from __future__ import annotations

import re
import struct
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from build_lang import MAGIC, VERSION, to_big5  # noqa: E402

TOTAL = 162
HEADING = re.compile(r"^##\s*段落\s*(\d+)\s*$")

ROOT = Path(__file__).resolve().parent.parent
SRC = ROOT / "translations" / "zh-Hant" / "paragraphs"
OUT = ROOT / "translations" / "paragraphs-zh-Hant.cat"


def parse(path: Path) -> dict[int, str]:
    """把一個 markdown 檔拆成「編號 → 正文」。"""
    out: dict[int, str] = {}
    num: int | None = None
    buf: list[str] = []

    def flush() -> None:
        if num is None:
            return
        body = "\n".join(buf).strip("\n")
        if num in out:
            raise SystemExit(f"{path}：段落 {num} 出現兩次")
        out[num] = body

    for line in path.read_text(encoding="utf-8").splitlines():
        if m := HEADING.match(line):
            flush()
            num, buf = int(m.group(1)), []
            continue
        if num is not None:
            buf.append(line)
    flush()
    return out


def to_text(body: str) -> str:
    """markdown 的段落分隔（空行）→ 兩個 `\\x0D`，段內換行 → 一個。"""
    paras = [" ".join(p.split()) for p in re.split(r"\n\s*\n", body) if p.strip()]
    return "\x0D\x0D".join(paras)


def main() -> None:
    if not SRC.is_dir():
        raise SystemExit(f"{SRC} 不存在——段落書的中文翻譯還沒開始")

    paragraphs: dict[int, str] = {}
    for p in sorted(SRC.glob("*.md")):
        for num, body in parse(p).items():
            if num in paragraphs:
                raise SystemExit(f"段落 {num} 在兩個檔案裡都有")
            paragraphs[num] = body

    errors: list[str] = []
    entries: list[tuple[str, bytes]] = []
    for num in sorted(paragraphs):
        if not 1 <= num <= TOTAL:
            errors.append(f"段落 {num} 超出 1–{TOTAL}")
            continue
        text = to_text(paragraphs[num])
        if not text:
            errors.append(f"段落 {num} 只有標題沒有正文")
            continue
        data, missing = to_big5(text)
        if missing:
            errors.append(f"段落 {num}：Big5 編不出 {''.join(sorted(set(missing)))}")
            continue
        entries.append((f"para:{num}", data))

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
    OUT.write_bytes(bytes(out))

    done = sorted(paragraphs)
    missing_nums = [n for n in range(1, TOTAL + 1) if n not in paragraphs]
    print(f"{len(entries)} 段 → {OUT}（{len(out)} bytes）")
    if missing_nums:
        print(f"還沒翻的 {len(missing_nums)} 段：{missing_nums[0]}–{missing_nums[-1]}")
    else:
        print(f"162 段全部翻完（{done[0]}–{done[-1]}）")


if __name__ == "__main__":
    main()
