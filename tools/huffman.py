"""Wasteland 的 Huffman 解壓（`sub_11AE8` ／ `sub_11C28` ／ `sub_11B83`）。

容器格式（docs/re/10）：

    +0  4 bytes  解壓後大小
    +4  3 bytes  'm' 's' 'q'
    +7  1 byte   磁碟編號
    +8  …        位元流：先是前序編碼的 Huffman 樹，接著是編碼資料

樹的節點在原版是 5 bytes（左 2、右 2、值 1），根固定在 `ds:9509h`，
其餘節點從 `ds:950Eh` 起依序配置、上限 768 個（`sub_11CA4`）。
這裡用 dict 模擬，位址保留成 key 以便對照原版。

位元順序 MSB first（`sub_11C54`：mask 由 `0x80` 起，測試後右移）。

用法（容器內）：
    python3 tools/huffman.py <輸入檔> [輸出檔]
"""

from __future__ import annotations

import struct
import sys
from pathlib import Path

ROOT = 0x9509
FIRST_NODE = 0x950E
NODE_SIZE = 5
MAX_NODES = 0x300


class Bits:
    def __init__(self, data: bytes, pos: int):
        self.data = data
        self.pos = pos
        self.cur = data[pos]
        self.pos += 1
        self.mask = 0x80

    def bit(self) -> int:
        if self.mask == 0:
            self.cur = self.data[self.pos]
            self.pos += 1
            self.mask = 0x80
        value = 1 if self.cur & self.mask else 0
        self.mask >>= 1
        return value

    def byte(self) -> int:
        v = 0
        for _ in range(8):
            v = ((v << 1) | self.bit()) & 0xFF
        return v


class Tree:
    def __init__(self):
        self.nodes: dict[int, list[int]] = {ROOT: [0, 0, 0]}
        self.next = FIRST_NODE
        self.count = 0

    def alloc(self) -> int:
        if self.count >= MAX_NODES:
            raise RuntimeError("節點數超過原版上限 768")
        addr = self.next
        self.next += NODE_SIZE
        self.count += 1
        self.nodes[addr] = [0, 0, 0]
        return addr

    def build(self, bits: Bits, addr: int, separator: bool) -> None:
        if bits.bit():  # 1 → 葉節點
            self.nodes[addr][2] = bits.byte()
            return
        left = self.alloc()
        right = self.alloc()
        self.nodes[addr][0] = left
        self.nodes[addr][1] = right
        self.build(bits, left, separator)
        if separator:
            bits.bit()  # sub_11C28 在兩次遞迴之間多讀一個 bit（0x11C44）
        self.build(bits, right, separator)


def decompress(
    data: bytes, separator: bool = True, verify_magic: bool = True
) -> tuple[bytes, dict]:
    """verify_magic=False 對應 `sub_11AE8` 的 AL=0 路徑。

    MSQ 區塊的尾段用的就是那條路徑：一樣有 8 bytes header，但第 5–7 bytes
    不是 `msq`，原版在 AL=0 時直接跳過驗證（`0x11B11`）。
    """
    out_size = struct.unpack_from("<I", data, 0)[0]
    magic = data[4:7]
    disk = data[7]
    if verify_magic and magic != b"msq":
        raise ValueError(f"magic 不是 msq：{magic!r}")

    bits = Bits(data, 8)
    tree = Tree()
    sys.setrecursionlimit(10000)
    tree.build(bits, ROOT, separator)

    out = bytearray()
    nodes = tree.nodes
    while len(out) < out_size:
        node = ROOT
        while nodes[node][0] != 0:
            node = nodes[node][1] if bits.bit() else nodes[node][0]
        out.append(nodes[node][2])

    return bytes(out), {
        "declared_size": out_size,
        "disk": disk,
        "tree_nodes": tree.count,
        "bytes_consumed": bits.pos,
        "input_size": len(data),
        "separator_bit": separator,
    }


def split_all(data: bytes) -> list[dict]:
    """整個檔案是子區塊串接：下一個的起點就是上一個位元流消耗完的位置。

    這是實測出來的——`allhtds1` 第一塊消耗到 5122，而檔案裡下一個 `msq`
    出現在 5126，正好差 4 bytes 的大小前綴。
    """
    out = []
    pos = 0
    while pos + 8 <= len(data) and data[pos + 4 : pos + 7] == b"msq":
        blob, info = decompress(data[pos:])
        printable = sum(1 for b in blob if 0x20 <= b < 0x7F or b in (10, 13))
        out.append(
            {
                "index": len(out),
                "offset": pos,
                "offset_hex": f"0x{pos:X}",
                "size": len(blob),
                "disk": info["disk"],
                "tree_nodes": info["tree_nodes"],
                "compressed_bytes": info["bytes_consumed"],
                "consumed_total": pos + info["bytes_consumed"],
                "printable_ratio": round(printable / len(blob), 3),
                "head_hex": blob[:16].hex(),
            }
        )
        pos += info["bytes_consumed"]
    return out


def main() -> None:
    src = Path(sys.argv[1])
    data = src.read_bytes()
    if "--all" in sys.argv:
        blocks = split_all(data)
        used = blocks[-1]["consumed_total"] if blocks else 0
        print(f"{src.name}: {len(blocks)} 個子區塊，用掉 {used}/{len(data)} bytes")
        for b in blocks:
            print(f"  [{b['index']:>2}] {b['offset_hex']:>7} 解出={b['size']:>6} "
                  f"壓縮={b['compressed_bytes']:>6} 節點={b['tree_nodes']:>3} "
                  f"可列印={b['printable_ratio']:.3f} {b['head_hex'][:24]}")
        return
    last = None
    for separator in (True, False):
        try:
            out, info = decompress(data, separator)
        except Exception as exc:  # noqa: BLE001
            last = f"separator={separator}：{exc}"
            continue
        printable = sum(1 for b in out if 0x20 <= b < 0x7F or b in (10, 13))
        print(
            f"separator={separator} 解出 {len(out)} bytes（宣告 {info['declared_size']}）"
            f" 樹節點={info['tree_nodes']} 消耗={info['bytes_consumed']}/{info['input_size']}"
            f" 可列印={printable / len(out):.3f}"
        )
        print("  前 48 bytes:", out[:48].hex())
        if len(sys.argv) > 2:
            Path(sys.argv[2]).write_bytes(out)
        return
    print("兩種讀法都失敗：", last)


if __name__ == "__main__":
    main()
