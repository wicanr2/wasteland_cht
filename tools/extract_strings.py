"""把原版的 4,827 條文字抽成翻譯目錄的原文檔（`docs/spec/11`）。

輸出兩個 TSV 到 `translations/source/`：

    exe.tsv      執行檔九張表，key ＝ exe:<表編號>:<槽>
    blocks.tsv   42 個地圖區塊，key ＝ blk:<檔名>:<資源編號>:<槽>

控制碼寫成 `\\xNN`——`\\r`、`\\x06`、`\\x10` 這些在原版是有意義的，
翻譯時**不能刪掉也不能自己加**（`docs/re/14` §4、`docs/re/28`）。

**這兩個檔是機器產生的，不要手改。** 重跑之後內容有變，代表抽取工具改了，
要回頭看，不是把原文改掉。

用法：
    python3 tools/extract_strings.py <wl.merged.exe> <game1> <game2> [輸出目錄]
"""

from __future__ import annotations

import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from decode_text import decode_tables  # noqa: E402
from decode_block_text import decode_blocks  # noqa: E402


def escape(s: str) -> str:
    """把控制碼寫成 \\xNN，其餘原樣保留（TAB 與換行一定要跳脫）。"""
    out = []
    for ch in s:
        o = ord(ch)
        if ch == "\\":
            out.append("\\\\")
        elif o < 0x20 or o == 0x7F or ch == "\t":
            out.append(f"\\x{o:02X}")
        else:
            out.append(ch)
    return "".join(out)


def unescape(s: str) -> str:
    """escape 的反向——build_lang.py 與測試都用它做 round-trip 檢查。"""
    out = []
    i = 0
    while i < len(s):
        if s[i] == "\\" and i + 1 < len(s):
            if s[i + 1] == "\\":
                out.append("\\")
                i += 2
                continue
            if s[i + 1] == "x" and i + 3 < len(s):
                out.append(chr(int(s[i + 2 : i + 4], 16)))
                i += 4
                continue
        out.append(s[i])
        i += 1
    return "".join(out)


def write_tsv(path: Path, rows: list[tuple[str, str]], header: str) -> int:
    lines = [f"# {header}", "# 機器產生，不要手改（tools/extract_strings.py）"]
    for key, text in rows:
        lines.append(f"{key}\t{escape(text)}")
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    return len(rows)


def main() -> None:
    if len(sys.argv) < 4:
        raise SystemExit(__doc__)
    exe = Path(sys.argv[1])
    game1, game2 = Path(sys.argv[2]), Path(sys.argv[3])
    outdir = Path(sys.argv[4]) if len(sys.argv) > 4 else Path("translations/source")
    outdir.mkdir(parents=True, exist_ok=True)

    exe_rows: list[tuple[str, str]] = []
    for ti, table in enumerate(decode_tables(exe.read_bytes())):
        for slot, text in enumerate(table["strings"]):
            if text:
                exe_rows.append((f"exe:{ti}:{slot}", text))

    blk_rows: list[tuple[str, str]] = []
    for block in decode_blocks(exe.read_bytes(), game1.read_bytes(), game2.read_bytes()):
        for slot, text in enumerate(block["strings"]):
            if text:
                blk_rows.append((f"blk:{block['file']}:{block['resource_id']}:{slot}", text))

    n1 = write_tsv(outdir / "exe.tsv", exe_rows, "執行檔的九張 5-bit 打包字串表")
    n2 = write_tsv(outdir / "blocks.tsv", blk_rows, "42 個地圖區塊的字串表")
    print(f"exe {n1} 條、blocks {n2} 條，合計 {n1 + n2} 條 → {outdir}")


if __name__ == "__main__":
    main()
