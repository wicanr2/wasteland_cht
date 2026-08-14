"""把執行檔裡某張 16-bit 表倒出來（跳表、位移表、常數表都用得上）。

位址換算見 `docs/re/00-master-index.md` §1：

    線性位址 = 0x1CE20 + <ds 位移>
    檔案位移 = 線性位址 − 0x10000 + MZ header 的 e_cparhdr × 16

跳表的值是 `seg000` 內的位移，所以順便印出線性位址（＋`0x10000`）方便對照
反組譯與函式索引。只輸出位址，不倒出原版內容。

用法：
    python3 tools/dump_word_table.py <wl.merged.exe> <ds 位移> <筆數> [--code|--bytes]

    --code   把每個值當成 seg000 的位移，多印一欄線性位址
    --bytes  改倒 8-bit 表（每列 16 個），例如 ds:D522h 的等級 → 階級編號
"""

from __future__ import annotations

import struct
import sys
from pathlib import Path

DS = 0x1CE20
CODE_BASE = 0x10000


def main() -> None:
    if len(sys.argv) < 4:
        raise SystemExit(__doc__)
    exe = Path(sys.argv[1]).read_bytes()
    offset = int(sys.argv[2], 0)
    count = int(sys.argv[3], 0)
    code = "--code" in sys.argv[4:]
    as_bytes = "--bytes" in sys.argv[4:]

    header_bytes = struct.unpack_from("<H", exe, 8)[0] * 16
    base = DS + offset - CODE_BASE + header_bytes
    print(f"ds:{offset:04X}h（線性 {DS + offset:#07x}，檔案位移 {base:#x}）")

    if as_bytes:
        for i in range(0, count, 16):
            row = exe[base + i : base + min(i + 16, count)]
            print(f"  ds:{offset + i:04X}h  " + " ".join(f"{b:02x}" for b in row))
        return

    for i in range(count):
        value = struct.unpack_from("<H", exe, base + i * 2)[0]
        line = f"  [{i:3}] {value:#06x}"
        if code:
            line += f" → 線性 {CODE_BASE + value:#07x}"
        print(line)


if __name__ == "__main__":
    main()
