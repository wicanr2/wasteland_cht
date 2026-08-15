"""從 42 個 MSQ 區塊倒出每張地圖的敵人資料表，只印統計不倒出內容。

表的位置寫在記錄區標頭 `P+0x04`／`P+0x05`（`sub_12A4C`，`docs/re/37`），
一筆 8 bytes，`+0x00`／`+0x01` 是 16-bit 基礎血量。

這支的用途是驗一件事：**基礎血量會不會超過 255**。血量的擲法是把高低位
分開各擲一次再合起來（`0x145FD`–`0x14639`），高位非 0 的話結果會多出
256 的倍數——所以「高位是不是永遠 0」直接決定 remake 要不要照抄那條路徑。

表的長度沒有欄位可查，但記錄區標頭整個就是一張 16-bit 位移表，所以**下一個
比它大的位移就是表尾**。⚠ 第 0 筆是全零（型別 0 ＝ 沒有敵人），拿「整筆全零」
當終止條件會一筆都掃不到。

用法（純 stdlib，不需要 IDA）：
    python3 tools/summarize_monsters.py <wl.unpacked.exe> <game1> <game2>
"""

from __future__ import annotations

import struct
import sys
from pathlib import Path

DS = 0x1CE20
BLOCK_TOTAL = 0xBD86
BLOCK_READ = 0xBD22
DIRECTORY = 0xBEC9
MAP_SIZE_SELECTOR = 0xBF1C

ENTRY = 8
HEADER_WORDS = 0x5C // 2


def decrypt(data: bytes, key: int) -> bytearray:
    out = bytearray(len(data))
    for i, b in enumerate(data):
        out[i] = b ^ key
        key = (key + 0x1F) & 0xFF
    return out


def blocks(exe: bytes, g1: bytes, g2: bytes):
    """區塊解析（只到解密後的 body）。多份工具共用這一支，不要各寫一份。"""
    for res_id, label, body, map_size, _tail in blocks_with_tail(exe, g1, g2):
        yield res_id, label, body, map_size


def blocks_with_tail(exe: bytes, g1: bytes, g2: bytes):
    """同上，但多回一個**還沒解壓的 Huffman 尾段**（地圖第 3 層在裡面）。

    尾段是 `span[read:]`——載入器讀進來解密的只到 `read`，剩下的是壓縮資料。
    ⚠ 第 3 層**不在 body 裡**。拿 `body[-D*D:]` 當第 3 層會得到一組看起來
    像樣、實際上是別的東西的數字（`docs/re/48` §3 記過這個坑）。
    """
    header_bytes = struct.unpack_from("<H", exe, 8)[0] * 16

    def at(off: int) -> int:
        return DS + off - 0x10000 + header_bytes

    directory: list[int] = []
    p = at(DIRECTORY)
    while exe[p] != 0xFF:
        directory.append(exe[p])
        p += 1
    total = [struct.unpack_from("<H", exe, at(BLOCK_TOTAL) + i * 2)[0] for i in range(len(directory))]
    read = [struct.unpack_from("<H", exe, at(BLOCK_READ) + i * 2)[0] for i in range(len(directory))]
    selector = exe[at(MAP_SIZE_SELECTOR) : at(MAP_SIZE_SELECTOR) + len(directory)]

    files = {"game1": g1, "game2": g2}
    cursor = {"game1": 0, "game2": 0}
    for res_id in range(len(directory)):
        label = {0x80: "game1", 0x40: "game2"}.get(directory[res_id] & 0xC0)
        if label is None or total[res_id] == 0:
            continue
        off = cursor[label]
        cursor[label] = off + total[res_id]
        span = files[label][off : off + total[res_id]]
        checksum = struct.unpack_from("<H", span, 4)[0]
        body = decrypt(span[6 : read[res_id]], (checksum & 0xFF) ^ (checksum >> 8))
        map_size = 0x1800 if selector[res_id] == 0x40 else 0x600
        yield res_id, label, body, map_size, span[read[res_id] :]


def main() -> None:
    if len(sys.argv) < 4:
        raise SystemExit(__doc__)
    exe, g1, g2 = (Path(p).read_bytes() for p in sys.argv[1:4])

    hi_nonzero = 0
    entries_seen = 0
    hp_lo, hp_hi = 0xFFFF, 0
    ragged = 0
    unbounded: list[int] = []
    print("資源  檔案    表位址   長度   筆數  基礎血量範圍")
    for res_id, label, body, map_size in blocks(exe, g1, g2):
        header = struct.unpack_from(f"<{HEADER_WORDS}H", body, map_size)
        table = header[2]
        if table == 0 or table >= len(body):
            print(f"{res_id:>4}  {label:<6}  {table:#07x}  表位址不在區塊內")
            continue
        # 表尾 ＝ 標頭裡下一個比它大的位移（沒有就用區塊尾）。
        after = [w for w in header if table < w <= len(body)]
        if not after:
            # 標頭裡沒有比它大的位移 → 上界只能用區塊尾，那不是表長。
            unbounded.append(res_id)
            print(f"{res_id:>4}  {label:<6}  {table:#07x}      —     —  無上界，不列入統計")
            continue
        span = min(after) - table
        if span % ENTRY:
            ragged += 1
        n = span // ENTRY
        lo, hi = 0xFFFF, 0
        for i in range(1, n):  # 第 0 筆是「沒有敵人」，永遠全零
            rec = body[table + i * ENTRY : table + (i + 1) * ENTRY]
            base = rec[0] | rec[1] << 8
            if rec[1]:
                hi_nonzero += 1
            lo, hi = min(lo, base), max(hi, base)
        entries_seen += max(0, n - 1)
        if n > 1:
            hp_lo, hp_hi = min(hp_lo, lo), max(hp_hi, hi)
            print(f"{res_id:>4}  {label:<6}  {table:#07x}  {span:>5}  {n - 1:>4}  {lo:>5}–{hi}")
        else:
            print(f"{res_id:>4}  {label:<6}  {table:#07x}  {span:>5}     0  —")

    print()
    print(f"合計 {entries_seen} 筆（不含第 0 筆），基礎血量 {hp_lo}–{hp_hi}，高位非 0 的有 {hi_nonzero} 筆")
    print(f"長度不是 8 的倍數的區塊：{ragged}")
    print(f"上界推不出來、沒列入統計的資源：{unbounded or '無'}")


if __name__ == "__main__":
    main()
