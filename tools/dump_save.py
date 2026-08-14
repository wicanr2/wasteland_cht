"""解開原版的存檔區並列出結構（`docs/re/30`）。

存檔藏在 `GAME1`／`GAME2` 末端的一個 MSQ 資源裡，與地圖區塊同一套加密：

    +0x00  magic 'msq0'／'msq1'
    +0x04  checksum（16-bit）＝ 0 − Σ 明文位元組
    +0x06  0x800 加密段：8 × 256 bytes（第 0 筆是全域狀態，其餘是角色）
    +0x806 0xA00 未加密段（出廠全零）

⚠ 位移是 **32-bit**：原版 `mov dx, 53C5h / mov cx, 2` 是 `cx:dx`，
所以是 `0x000253C5` 不是 `0x53C5`。

只輸出結構與統計，不倒出原版內容（名字會顯示，因為那是判斷「解對了」的依據）。

用法：
    python3 tools/dump_save.py <game1> <game2>
"""

from __future__ import annotations

import struct
import sys
from pathlib import Path

# 存檔資源的檔案位移（原版 sub_18744 的 seek 目標，cx:dx）。
SAVE_AT = {"game1": 0x000253C5, "game2": 0x00028BC7}
PLAIN_LEN = 0x800
TAIL_LEN = 0xA00
RECORD = 256
SLOT_GROUPS = (0x00, 0x0E, 0x1C, 0x2A)  # 四組隊伍槽表，各 14 bytes
GLOBAL_AT = 0x78  # ds:464Eh–465Bh 那 14 bytes
SERIAL_AT = 0xF5  # 32-bit 存檔序號
PLACE_AT = 0xD0  # 地點名稱（＝ 記憶體 ds:7201h，全域記錄載到 0x7131）


def decrypt(raw: bytes, checksum: int) -> bytearray:
    out = bytearray(len(raw))
    key = (checksum & 0xFF) ^ (checksum >> 8)
    for i, c in enumerate(raw):
        out[i] = c ^ key
        key = (key + 0x1F) & 0xFF
    return out


def ascii_of(raw: bytes) -> str:
    return "".join(chr(c) if 0x20 <= c < 0x7F else "." for c in raw)


def dump(path: Path, at: int) -> None:
    data = path.read_bytes()
    magic = data[at : at + 4]
    checksum = struct.unpack_from("<H", data, at + 4)[0]
    body = decrypt(data[at + 6 : at + 6 + PLAIN_LEN], checksum)
    ok = ((-sum(body)) & 0xFFFF) == checksum
    tail = data[at + 6 + PLAIN_LEN : at + 6 + PLAIN_LEN + TAIL_LEN]

    print(f"=== {path.name} @ {at:#x}  magic={magic!r}  checksum={checksum:#06x} "
          f"{'✓ 驗證通過' if ok else '✗ 驗證失敗'}")
    print(f"    未加密尾段 {TAIL_LEN} bytes：非零 {sum(1 for c in tail if c)} 個")

    g = body[:RECORD]
    print("    ── 第 0 筆：全域狀態 ──")
    for n, base in enumerate(SLOT_GROUPS):
        slot = g[base : base + 14]
        members = [b for b in slot[:8] if b]
        print(f"      隊伍組 {n}：成員 {members}  座標 ({slot[8]}, {slot[9]}) "
              f"地圖 {slot[10]}  傳送 ({slot[11]}, {slot[12]})")
    gl = g[GLOBAL_AT : GLOBAL_AT + 14]
    print(f"      視窗原點 ({gl[0]}, {gl[1]})  32-bit 累計 {int.from_bytes(gl[2:5], 'little')}"
          f"  隊伍人數 {gl[8]}  目前地圖 {gl[7]}"
          f"  時鐘 {gl[12]:02d}:{gl[11]:02d}")
    print(f"      存檔序號 {struct.unpack_from('<I', g, SERIAL_AT)[0]}"
          f"  地點 {ascii_of(g[PLACE_AT:PLACE_AT + 13])!r}")

    print("    ── 第 1–7 筆：角色 ──")
    for r in range(1, 8):
        rec = body[r * RECORD : (r + 1) * RECORD]
        name = ascii_of(rec[:14]).rstrip(".")
        if not name:
            print(f"      {r}：（空）")
            continue
        print(f"      {r}：{name!r}  屬性 {list(rec[0x0E:0x15])}  "
              f"MAXCON {struct.unpack_from('<H', rec, 0x1B)[0]}  "
              f"CON {struct.unpack_from('<h', rec, 0x1D)[0]}  "
              f"階級 {ascii_of(rec[0x32:0x3E]).rstrip('.')!r}")


def main() -> None:
    if len(sys.argv) < 3:
        raise SystemExit(__doc__)
    for path, key in zip(sys.argv[1:3], ("game1", "game2")):
        dump(Path(path), SAVE_AT[key])


if __name__ == "__main__":
    main()
