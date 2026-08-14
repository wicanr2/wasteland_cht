"""列出還沒翻的字串，或把某個區塊的未翻條目倒成可以直接編輯的 TSV。

翻譯還剩一千多條散在三十幾張地圖上，每一輪都要先問「哪一塊最大、這塊還缺哪幾條」。
這支把那兩個問題做成一個指令，判定邏輯與 `tools/build_lang.py` 完全一致
（逐 key 檔優先，`_shared.tsv` 依**原文**補上，底線開頭的檔案不是逐 key 檔）。

用法：
    python3 tools/untranslated.py                 # 依未翻條數排序列出所有區塊
    python3 tools/untranslated.py game2 20        # 倒出該區塊的未翻條目（已跳脫，可直接貼進 TSV）
"""

from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from build_lang import read_tsv  # noqa: E402
from extract_strings import escape, unescape  # noqa: E402

ROOT = Path(__file__).resolve().parent.parent
ZH = ROOT / "translations" / "zh-Hant"
SRC = ROOT / "translations" / "source"


def load() -> tuple[dict[str, str], set[str], set[str]]:
    source: dict[str, str] = {}
    for p in sorted(SRC.glob("*.tsv")):
        source.update(read_tsv(p))
    done: set[str] = set()
    for p in sorted(ZH.glob("*.tsv")):
        if p.name.startswith("_"):
            continue
        done |= set(read_tsv(p))
    # ⚠ 共用層是以**原文**為 key，那一欄同樣是跳脫過的，要 unescape 才對得上原文。
    # 忘記這一步不會報錯，只會讓這支工具把一千多條已經翻好的字串報成未翻。
    shared: set[str] = set()
    shared_path = ZH / "_shared.tsv"
    if shared_path.is_file():
        for line in shared_path.read_text(encoding="utf-8").splitlines():
            if not line.strip() or line.startswith("#"):
                continue
            shared.add(unescape(line.split("\t", 1)[0]))
    return source, done, shared


def pending(source, done, shared) -> list[tuple[str, str]]:
    return [
        (k, v)
        for k, v in source.items()
        if k not in done and v not in shared
    ]


def main() -> None:
    source, done, shared = load()
    todo = pending(source, done, shared)

    if len(sys.argv) == 3:
        prefix = f"blk:{sys.argv[1]}:{sys.argv[2]}:"
        rows = [(k, v) for k, v in todo if k.startswith(prefix)]
        for k, v in sorted(rows, key=lambda kv: int(kv[0].rsplit(":", 1)[1])):
            print(f"{k}\t{escape(v)}")
        print(f"# {len(rows)} 條未翻", file=sys.stderr)
        return

    counts: dict[str, int] = {}
    for k, _ in todo:
        counts[k.rsplit(":", 1)[0]] = counts.get(k.rsplit(":", 1)[0], 0) + 1
    for blk, n in sorted(counts.items(), key=lambda kv: -kv[1]):
        print(f"{n:5d}  {blk}")
    print(f"總計 {len(todo)} 條未翻／原文 {len(source)} 條")


if __name__ == "__main__":
    main()
