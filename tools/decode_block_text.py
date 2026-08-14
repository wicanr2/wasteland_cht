"""解出 42 個 MSQ 區塊各自的 5-bit 打包字串表（`docs/re/18`）。

關鍵在解密長度：`sub_11A59` 只解 **區塊標頭第一個 word** 那麼多 bytes，
不是整個區塊。而同一個 word 又被 `sub_1790B` 當成字串表的基址
（`ds:4692h`）——換句話說 **字串表就從加密區結束的地方開始，而且不加密**。
先前把整個區塊都拿去 XOR，等於把字串表攪爛，所以看起來是高熵資料。

每個區塊有自己的字元對照表，不能套執行檔那張。

用法：
    python3 tools/decode_block_text.py <wl.merged.exe> <game1> <game2> <out.json>
"""

from __future__ import annotations

import json
import struct
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from decode_text import decode_table  # noqa: E402

DS = 0x1CE20
BLOCK_LENGTHS = 0xBD86
READ_LENGTHS = 0xBD22  # 載入器實際讀進緩衝區的長度——字串只能解到這裡為止
DIRECTORY = 0xBEC9
MAP_SIZE_SELECTOR = 0xBF1C


def decrypt_block(raw: bytes, checksum: int, header_at: int) -> tuple[bytearray, int]:
    """照 sub_11A59：先解到標頭+2 讀出長度，再解到那個長度為止，其餘保持原樣。"""
    out = bytearray(raw)
    key = (checksum & 0xFF) ^ (checksum >> 8)
    first = min(header_at + 2, len(raw))
    for i in range(first):
        out[i] = raw[i] ^ key
        key = (key + 0x1F) & 0xFF
    if header_at + 2 > len(raw):
        return out, len(raw)
    length = struct.unpack_from("<H", out, header_at)[0]
    for i in range(first, min(length, len(raw))):
        out[i] = raw[i] ^ key
        key = (key + 0x1F) & 0xFF
    return out, length


def main() -> None:
    exe_p, g1_p, g2_p, out_p = (Path(p) for p in sys.argv[1:5])
    exe = exe_p.read_bytes()
    header_bytes = struct.unpack_from("<H", exe, 8)[0] * 16
    at = lambda off: DS + off - 0x10000 + header_bytes  # noqa: E731

    directory: list[int] = []
    p = at(DIRECTORY)
    while exe[p] != 0xFF:
        directory.append(exe[p])
        p += 1
    selector = exe[at(MAP_SIZE_SELECTOR) : at(MAP_SIZE_SELECTOR) + len(directory)]
    block_len = [
        struct.unpack_from("<H", exe, at(BLOCK_LENGTHS) + i * 2)[0]
        for i in range(len(directory))
    ]
    read_len = [
        struct.unpack_from("<H", exe, at(READ_LENGTHS) + i * 2)[0]
        for i in range(len(directory))
    ]

    files = {"game1": g1_p.read_bytes(), "game2": g2_p.read_bytes()}
    cursor = {"game1": 0, "game2": 0}
    blocks = []
    total = 0

    for res_id, raw_flag in enumerate(directory):
        label = {0x80: "game1", 0x40: "game2"}.get(raw_flag & 0xC0)
        if label is None or block_len[res_id] == 0:
            continue
        data = files[label]
        off = cursor[label]
        cursor[label] = off + block_len[res_id]
        span = data[off : off + block_len[res_id]]
        checksum = struct.unpack_from("<H", span, 4)[0]
        header_at = 0x1800 if selector[res_id] == 0x40 else 0x600
        # ⚠ 只餵「載入器實際讀進來」的那一段。多餵的部分是壓縮的尾段，
        # 解 5-bit 會解出看起來像文字的雜訊，而且一路解到檔案結束。
        body, length = decrypt_block(span[6 : read_len[res_id]], checksum, header_at)

        entry: dict = {
            "resource_id": res_id,
            "file": label,
            "body_length": len(body),
            "header_at": f"0x{header_at:04X}",
            "encrypted_length": f"0x{length:04X}",
            "string_table_at": f"0x{length:04X}",
        }
        if length + 0x40 >= len(body):
            entry["note"] = "字串表超出區塊，跳過"
            blocks.append(entry)
            continue
        try:
            table = decode_table(bytes(body), length, max_groups=256)
        except Exception as exc:  # noqa: BLE001
            entry["error"] = str(exc)
            blocks.append(entry)
            continue
        strings = table["strings"]
        entry["group_count"] = table["group_count"]
        entry["slot_count"] = len(strings)
        entry["alphabet"] = table["alphabet"]
        entry["strings"] = strings
        total += sum(1 for s in strings if s)
        blocks.append(entry)

    decoded = [b for b in blocks if "slot_count" in b]
    summary = {
        "block_count": len(blocks),
        "blocks_with_text": len(decoded),
        "slot_total": sum(b.get("slot_count", 0) for b in blocks),
        "non_empty_strings": total,
        "longest": max(
            ((len(s), b["resource_id"]) for b in decoded for s in b["strings"]),
            default=(0, None),
        )[0],
    }
    out_p.write_text(
        json.dumps({"summary": summary, "blocks": blocks}, ensure_ascii=False, indent=2),
        encoding="utf-8",
    )
    print(json.dumps(summary, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
