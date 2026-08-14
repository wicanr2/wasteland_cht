"""解碼原版的 5-bit 打包文字表（`docs/re/17`）。

演算法照 `sub_178B9`／`sub_17B8F`／`sub_17BC7` 逐條實作：

    表基址 +0x00 .. +0x3B   60 bytes 字元對照表（符號 → ASCII，0 ＝ 字串結束）
    表基址 +0x3C ..         16-bit 位移表，**每 4 個字串一項**，位移相對於 +0x3C
    位移指到的地方           5-bit 符號流（每個 byte 由低位往高位取）

    符號 0x1E ＝ 下一個字元轉大寫
    符號 0x1F ＝ escape，再取一個符號後 +0x1E（所以字元集共 60 個）

取第 N 個字串：先跳到第 `N >> 2` 項位移，再往下解掉 `N & 3` 個字串。

用法：
    python3 tools/decode_text.py <wl.merged.exe> [輸出.json]
"""

from __future__ import annotations

import json
import struct
import sys
from pathlib import Path

DS = 0x1CE20
ALPHABET_SIZE = 0x3C

# 執行檔內的九張表。來源是「誰寫 ds:4692h」——那個變數就是目前的字串表基址，
# 全檔 13 個寫入點裡有 9 個寫的是常數（其餘是 MSQ 區塊標頭 +0x00，隨地圖而變）。
EXE_TABLES = {
    0xA703: "sub_16390",
    0xAB3E: "sub_16CBD",
    0xB270: "sub_17ACE（18 個呼叫端，技能／物品／介面）",
    0xCE4B: "sub_1A3E1",
    0xD18E: "0x1B7C2（結局敘述）",
    0xD622: "sub_1BB5D",
    0xDACC: "sub_1BE31",
    0xDBF8: "sub_1C213",
    0xDCED: "sub_1C561",
}
SHIFT_SYMBOL = 0x1E
ESCAPE_SYMBOL = 0x1F


class BitReader:
    """每個 byte 由最低位往最高位取，5 個位元組成一個符號。"""

    def __init__(self, buf: bytes, pos: int) -> None:
        self.buf = buf
        self.pos = pos
        self.cur = 0
        self.left = 0

    def symbol(self) -> int:
        value = 0
        for bit in range(5):
            if self.left == 0:
                self.cur = self.buf[self.pos]
                self.pos += 1
                self.left = 8
            value |= (self.cur & 1) << bit
            self.cur >>= 1
            self.left -= 1
        return value


def decode_run(
    buf: bytes, alpha_base: int, pos: int, count: int
) -> tuple[list[str], bool]:
    """從 pos 解 count 個字串，回傳（解出來的字串, 是否整組都解完）。

    `alpha_base` 是字元對照表的絕對位置。原版沒有邊界檢查——escape 之後的符號
    最大到 0x3D，會讀到 60 bytes 之外，這裡照樣不檢查。

    緩衝區用完時**保留已經解完的字串**（那些是真的），只回報這一組沒解完。
    整組丟掉會連帶丟掉真字串，這是先前少算的原因。
    """
    reader = BitReader(buf, pos)
    out = []
    for _ in range(count):
        chars: list[str] = []
        upper = False
        try:
            while True:
                sym = reader.symbol()
                if sym == SHIFT_SYMBOL:
                    upper = True
                    continue
                if sym == ESCAPE_SYMBOL:
                    sym = reader.symbol() + SHIFT_SYMBOL
                code = buf[alpha_base + sym]
                if code == 0:
                    break
                ch = chr(code)
                if upper and "a" <= ch <= "z":
                    ch = ch.upper()
                upper = False
                chars.append(ch)
        except IndexError:
            return out, False
        out.append("".join(chars))
    return out, True


def decode_table(buf: bytes, base: int, max_groups: int = 512) -> dict:
    alphabet = buf[base : base + ALPHABET_SIZE]
    data = base + ALPHABET_SIZE

    # 位移表沒有長度欄位，但它自己的第一項就是第一個字串的位置，
    # 所以「第一項 ÷ 2」就是表的項數。最後一項在原版裡是沒用到的填充，
    # 值落在區塊外——遇到不遞增就停，並把兩個數字都報出來，不要靜悄悄截掉。
    declared = struct.unpack_from("<H", buf, data)[0] // 2
    offsets: list[int] = []
    for i in range(min(declared, max_groups)):
        off = struct.unpack_from("<H", buf, data + i * 2)[0]
        # 一組只有四個字串，跨距不可能到 1 KB；差太多就是已經讀到表尾的填充。
        if offsets and (off <= offsets[-1] or off - offsets[-1] > 0x400):
            break
        offsets.append(off)

    # 逐組解，讀到緩衝區尾就停——位移表的最後幾項可能指到區塊外，
    # 這時要記下「解到第幾組」，不要整張表放棄。
    strings: list[str] = []
    decoded_groups = 0
    for off in offsets:
        part, complete = decode_run(buf, base, data + off, 4)
        strings.extend(part)
        if not complete:
            break
        decoded_groups += 1

    # 一組固定四個槽，**最後一組不一定四個都用到**——沒用到的槽解出來是雜訊。
    # 字串是靠編號定址的，所以這裡把槽原樣保留，只另外報非空的條數；
    # 不做「看起來像不像文字」的裁切（那是猜測，會把真的字串一起丟掉）。
    return {
        "alphabet": alphabet.hex(),
        "alphabet_text": alphabet.decode("latin1"),
        "declared_groups": declared,
        "group_count": len(offsets),
        "decoded_groups": decoded_groups,
        "slot_count": len(strings),
        "non_empty_count": sum(1 for s in strings if s),
        "strings": strings,
    }


def main() -> None:
    if len(sys.argv) < 2:
        raise SystemExit(__doc__)
    exe = Path(sys.argv[1]).read_bytes()
    if exe[:2] != b"MZ":
        raise SystemExit("要餵解包映像（MZ 執行檔）")
    header = int.from_bytes(exe[8:10], "little") * 16

    tables = []
    slots = 0
    non_empty = 0
    for ds_off, who in EXE_TABLES.items():
        result = decode_table(exe, DS + ds_off - 0x10000 + header)
        result["table_ds_offset"] = f"0x{ds_off:04X}"
        result["table_linear"] = f"0x{DS + ds_off:05X}"
        result["set_by"] = who
        tables.append(result)
        slots += result["slot_count"]
        non_empty += result["non_empty_count"]

    if len(sys.argv) > 2:
        Path(sys.argv[2]).write_text(
            json.dumps(
                {
                    "table_count": len(tables),
                    "slot_total": slots,
                    "non_empty_total": non_empty,
                    "tables": tables,
                },
                ensure_ascii=False,
                indent=2,
            ),
            encoding="utf-8",
        )
        print(f"→ {sys.argv[2]}（{len(tables)} 張表、{slots} 個槽、{non_empty} 條非空）")
    else:
        for result in tables:
            print(f"=== {result['table_ds_offset']} {result['set_by']}")
            for i, s in enumerate(result["strings"]):
                if s:
                    print(f"  {i:>3} {s!r}")


if __name__ == "__main__":
    main()
