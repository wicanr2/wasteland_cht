"""解密 MSQ 區塊，並用原版自己的 checksum 條件驗證。

演算法照 `sub_11A59`（docs/re/08）逐條實作：

    key = lo(checksum) XOR hi(checksum)
    每個 byte：plain = cipher XOR key，然後 key = (key + 0x1F) & 0xFF

驗證條件也是原版的：解密後所有 byte 的負和（16-bit）要等於 checksum。
不符就報 FAIL，不要為了讓報表好看而放寬。

只輸出統計，不倒出解密內容（原版資產不散布）。

用法（容器內）：
    python3 tools/decrypt_msq.py <wl.unpacked.exe> <game1> <game2> <out.json>
"""

from __future__ import annotations

import json
import re
import struct
import sys
from pathlib import Path

DS = 0x1CE20
BLOCK_LENGTHS = 0xBD86  # 依資源編號排列的區塊總長度，兩個檔共用同一張表
READ_LENGTHS = 0xBD22
DIRECTORY = 0xBEC9
WORD = re.compile(rb"[A-Za-z]{4,}")


def decrypt(data: bytes, key: int) -> bytearray:
    out = bytearray(len(data))
    for i, c in enumerate(data):
        out[i] = c ^ key
        key = (key + 0x1F) & 0xFF
    return out


def main() -> None:
    exe_p, g1_p, g2_p, out_p = (Path(p) for p in sys.argv[1:5])
    exe = exe_p.read_bytes()
    files = {"game1": g1_p.read_bytes(), "game2": g2_p.read_bytes()}

    header = struct.unpack_from("<H", exe, 8)[0] * 16
    at = lambda off: DS + off - 0x10000 + header  # noqa: E731

    directory = []
    p = at(DIRECTORY)
    while exe[p] != 0xFF:
        directory.append(exe[p])
        p += 1

    block_len = [
        struct.unpack_from("<H", exe, at(BLOCK_LENGTHS) + i * 2)[0]
        for i in range(len(directory))
    ]
    read_len = [
        struct.unpack_from("<H", exe, at(READ_LENGTHS) + i * 2)[0]
        for i in range(len(directory))
    ]

    # 資源編號在兩個檔案裡各自累進，順序照目錄。
    cursor = {"game1": 0, "game2": 0}
    results = []
    for res_id, raw in enumerate(directory):
        disk = raw & 0xC0
        label = {0x80: "game1", 0x40: "game2"}.get(disk)
        length = block_len[res_id]
        if label is None or length == 0:
            results.append(
                {
                    "resource_id": res_id,
                    "raw": f"0x{raw:02X}",
                    "file": label,
                    "block_length": length,
                    "note": "目錄高 2 bits 為 0x00 或長度 0，沒有對應區塊",
                }
            )
            continue

        data = files[label]
        off = cursor[label]
        cursor[label] = off + length
        span = data[off : off + length]
        magic = span[:4].decode("latin1")
        checksum = struct.unpack_from("<H", span, 4)[0]
        body = decrypt(span[6:], (checksum & 0xFF) ^ (checksum >> 8))

        total = 0
        verified_at = None
        for i, b in enumerate(body):
            total = (total - b) & 0xFFFF
            if total == checksum:
                verified_at = i + 1
                break

        words = WORD.findall(bytes(body))
        results.append(
            {
                "resource_id": res_id,
                "raw": f"0x{raw:02X}",
                "file": label,
                "offset": off,
                "offset_hex": f"0x{off:X}",
                "block_length": length,
                "read_length": read_len[res_id],
                "magic": magic,
                "magic_ok": magic in ("msq0", "msq1"),
                "checksum": f"0x{checksum:04X}",
                "checksum_verified_at": verified_at,
                "checksum_ok": verified_at is not None,
                "ascii_word_count": len(words),
                "sample_words": [w.decode("latin1") for w in words[:8]],
            }
        )

    ok = sum(1 for r in results if r.get("checksum_ok"))
    total_blocks = sum(1 for r in results if r.get("magic"))
    out_p.write_text(
        json.dumps(
            {
                "directory_entries": len(directory),
                "blocks": total_blocks,
                "checksum_verified": ok,
                "results": results,
            },
            ensure_ascii=False,
            indent=2,
        ),
        encoding="utf-8",
    )
    print(f"區塊 {total_blocks} 個，magic 正確 "
          f"{sum(1 for r in results if r.get('magic_ok'))} 個，checksum 通過 {ok} 個")
    texty = [r for r in results if r.get("ascii_word_count", 0) > 50]
    print(f"解密後含 50 個以上英文單字的區塊：{len(texty)} 個")
    for r in texty[:8]:
        print(f"  #{r['resource_id']:>2} {r['file']} 單字={r['ascii_word_count']:>4} "
              f"{r['sample_words'][:5]}")


main()
