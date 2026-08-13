"""把 DS000 清冊 JSON 收攏成可提交的 markdown 摘要（純 stdlib，不需要 IDA）。

用法（容器內）：
    python3 tools/summarize_inventory.py <inventory.json> <輸出.md>

只做整理與排序，不加任何語意推論——推論屬於 docs/re/ 的分析文件，
在這裡混進來會讓「工具輸出」與「人的判斷」分不開。
"""

from __future__ import annotations

import json
import sys
from pathlib import Path


def table(rows: list[list[str]], header: list[str]) -> str:
    out = ["| " + " | ".join(header) + " |", "|" + "---|" * len(header)]
    for row in rows:
        out.append("| " + " | ".join(row) + " |")
    return "\n".join(out)


def main() -> None:
    src, dst = Path(sys.argv[1]), Path(sys.argv[2])
    data = json.loads(src.read_text(encoding="utf-8"))

    funcs = data["functions"]
    strings = data["strings"]
    by_size = sorted(funcs, key=lambda f: -f["size"])
    by_callers = sorted(funcs, key=lambda f: -f["caller_count"])
    orphans = [f for f in funcs if f["caller_count"] == 0]

    lines: list[str] = []
    lines.append("# DS000：`wl.exe` 基準清冊（IDA 匯出）\n")
    lines.append("> 由 `tools/ida/export_inventory.py` 匯出、")
    lines.append("> `tools/summarize_inventory.py` 整理，內容全部是工具輸出，未加語意推論。\n")
    lines.append(
        table(
            [
                ["輸入檔", f"`{data['input_file']}`"],
                ["SHA-256", f"`{data['input_sha256']}`"],
                ["大小", f"{data['input_size']:,} bytes"],
                ["processor", f"`{data['processor']}`"],
                ["segments", str(len(data["segments"]))],
                ["entry points", str(len(data["entries"]))],
                ["自動辨識 functions", str(len(funcs))],
                ["IDA 認定 strings", str(len(strings))],
                ["沒有直接 caller 的 functions", str(len(orphans))],
            ],
            ["項目", "值"],
        )
    )

    lines.append("\n## Segments\n")
    lines.append(
        table(
            [
                [
                    f"`{s['name']}`",
                    s["class"] or "",
                    s["start_ea"],
                    s["end_ea"],
                    f"{s['size']:,}",
                    f"{s['bitness_bits']}-bit",
                ]
                for s in data["segments"]
            ],
            ["name", "class", "start", "end", "size", "bitness"],
        )
    )

    lines.append("\n## Entry points\n")
    lines.append(
        table(
            [
                [
                    str(e["ordinal"]),
                    e["ida_address"],
                    f"{e['segment']}+{e['segment_offset']}",
                    f"`{e['name']}`" if e["name"] else "",
                ]
                for e in data["entries"]
            ],
            ["ordinal", "linear", "segment:offset", "name"],
        )
    )

    lines.append("\n## 最大的 20 個函式\n")
    lines.append(
        table(
            [
                [
                    f["ida_address"],
                    f"{f['segment']}+{f['segment_offset']}",
                    f"`{f['original_name']}`",
                    f"{f['size']:,}",
                    str(f["caller_count"]),
                ]
                for f in by_size[:20]
            ],
            ["linear", "segment:offset", "IDA 名稱", "size", "callers"],
        )
    )

    lines.append("\n## 被呼叫最多的 20 個函式\n")
    lines.append(
        table(
            [
                [
                    f["ida_address"],
                    f"{f['segment']}+{f['segment_offset']}",
                    f"`{f['original_name']}`",
                    str(f["caller_count"]),
                    f"{f['size']:,}",
                ]
                for f in by_callers[:20]
            ],
            ["linear", "segment:offset", "IDA 名稱", "callers", "size"],
        )
    )

    lines.append(f"\n## IDA 認定的字串（全部 {len(strings)} 筆）\n")
    lines.append(
        table(
            [
                [
                    s["ida_address"],
                    str(s["length"]),
                    str(s["xref_count"]),
                    "`" + s["text"].replace("\n", "\\n").replace("|", "\\|") + "`",
                ]
                for s in sorted(strings, key=lambda s: int(s["ida_address"], 16))
            ],
            ["linear", "len", "xrefs", "內容"],
        )
    )

    dst.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"→ {dst}")


main()
