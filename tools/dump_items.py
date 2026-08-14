#!/usr/bin/env python3
"""倒出物品資料表（95 筆 × 8 bytes），配上執行檔字串表裡的物品名與技能名。

不需要 IDA。

    python3 tools/dump_items.py <game1> <exe-strings.json> [輸出.md]

⚠ **這張表不在執行檔裡，在存檔區。** 映像裡 `ds:7A31h` 那段是零，
表是開檔時才讀進來的（`0x186A9`，`docs/re/45` §2）：

    偏移 = 0x000253C5 ＋ 0x1206 ＋ ds:BE20h[存檔槽]      （game1）
           0x00028BC7 ＋ 0x1206                          （game2）

每個存檔槽一份，所以**賣東西改到的 `+0x02` 會跟著存檔走**。
區塊是 MSQ 加密的，解法照 `docs/re/08`：

    key = lo(checksum) XOR hi(checksum)；每 byte 解完 key += 0x1F

解完用原版自己的條件驗（明文位元組的負和 ＝ checksum），不符就報 FAIL，
不要為了讓輸出好看而放寬。

索引：`sub_17AE0` 的基址是 `0x7A39` ＝ 表首 ＋ 8，位址 ＝ `0x7A39 + 8 × 索引`，
所以**索引 0 落在表的第 1 筆**（第 0 筆沒有人定址得到）。物品名 ＝ 字串表 2 的
第 `36 + 索引` 條，技能名 ＝ 同一張表的第 `+0x05` 條。
"""

from __future__ import annotations

import json
import struct
import sys
from pathlib import Path

SAVE_BASE = {"game1": 0x000253C5, "game2": 0x00028BC7}
ITEM_DELTA = 0x1206  # 存檔資源 → 物品表資源
SLOT_STRIDE = 0x2FE  # ds:BE20h 的三個槽：0 / 0x2FE / 0x5FC
ITEM_BYTES = 0x2F8   # 95 × 8
NAME_BASE = 36       # 字串表 2 的第 36 條是第 0 號物品

# 類別 ＝ `+0x03 >> 3`。名稱是本專案依表的內容命名的，原版沒有這些字串。
CLASS_NAMES = {
    1: "近戰／徒手",
    2: "手槍與投擲",
    3: "火焰／卡賓",
    4: "步槍",
    6: "衝鋒槍",
    7: "突擊步槍",
    8: "反戰車（輕）",
    9: "反戰車（重）",
    10: "雷射手槍",
    11: "能量（中）",
    12: "能量（重）",
    13: "爆裂物",
    14: "彈藥",
    15: "護甲",
    16: "一般物品",
    17: "可賣雜物",
    18: "劇情物品",
}


def decrypt(blob: bytes, at: int, length: int) -> tuple[bytes, bool]:
    """解一個 MSQ 區塊，回 (明文, 驗證是否通過)。"""
    checksum = struct.unpack_from("<H", blob, at + 4)[0]
    key = (checksum & 0xFF) ^ (checksum >> 8)
    out = bytearray()
    for b in blob[at + 6 : at + 6 + length]:
        out.append(b ^ key)
        key = (key + 0x1F) & 0xFF
    return bytes(out), ((-sum(out)) & 0xFFFF) == checksum


def main(argv: list[str]) -> int:
    if len(argv) not in (3, 4):
        sys.exit(__doc__)
    blob = Path(argv[1]).read_bytes()
    names = json.loads(Path(argv[2]).read_text())["tables"][2]["strings"]

    at = SAVE_BASE["game1"] + ITEM_DELTA
    if at + 6 + ITEM_BYTES > len(blob) or blob[at : at + 3] != b"msq":
        sys.exit(f"{at:#x} 不是 MSQ 區塊——確認給的是 game1")
    table, ok = decrypt(blob, at, ITEM_BYTES)

    def name(i: int) -> str:
        s = names[NAME_BASE + i] if NAME_BASE + i < len(names) else ""
        # 物品名帶單複數變形碼（`docs/re/28`）：前綴 ＋ 單數尾 ＋ 複數尾。
        # ⚠ 只取第一段會得到 `Kni`——單數是**前兩段接起來**。
        parts = s.split("\n")
        if len(parts) >= 3:
            return parts[0] + parts[1]
        return parts[0] or f"（第 {i} 條沒有名字）"

    rows = ["# 物品資料表（工具輸出，不含推論）", ""]
    rows.append(f"來源 `game1` 偏移 `{at:#x}`，checksum 驗證：{'通過' if ok else '**FAIL**'}")
    rows.append("")
    rows.append("| # | 名稱 | 基礎價 | 類別 | 彈匣 | 技能 | 骰數 | 彈藥 |")
    rows.append("|---:|---|---:|---|---:|---|---:|---|")
    for idx in range(95):
        e = table[(idx + 1) * 8 : (idx + 2) * 8]
        if len(e) < 8:
            break
        price = struct.unpack_from("<H", e, 0)[0]
        cls = e[3] >> 3
        skill = names[e[5]] if e[5] and e[5] < len(names) else "—"
        ammo = name(e[7] - 0) if e[7] else "—"
        rows.append(
            f"| {idx} | {name(idx)} | {price} | {cls}（{CLASS_NAMES.get(cls, '未命名')}）"
            f" | {e[4] or '—'} | {skill} | {e[6]} | {ammo} |"
        )

    text = "\n".join(rows) + "\n"
    if len(argv) == 4:
        Path(argv[3]).write_text(text, encoding="utf-8")
        print(f"→ {argv[3]}")
    else:
        print(text)
    return 0 if ok else 1


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
