#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
etunpack.py — 倚天中文系統 ETUNPACK V1.00 壓縮字型解包器(純 Python)

用途:把倚天 3.53 光碟裡的 24 點字型 STD.24M / STD.24K / STD.24L /
STD.24R / STD.24B / STD.24S 解成未壓縮的 stdfont.24 點陣檔。

格式(逆向自 ETUNPACK.EXE,DISK2 / FILES 目錄下的 DOS 執行檔):

  檔頭
    0x00  char[16]  "ETUNPACK V1.00\0\0"
    0x10  uint32LE  dir_ofs    目錄項位置(實測 0x20)
    0x14  uint32LE  data_ofs   壓縮碼流起點(實測 0x40)

  目錄項(位於 dir_ofs)
    +0x00 uint32LE  項目大小(0x20)
    +0x04 char[12]  輸出檔名,例:"stdfont.24"
    +0x12 uint8     xor_start  差分還原起始「列」
    +0x13 uint8     xor_end    差分還原結束「列」(不含)

  碼流(位於 data_ofs)
    1) 三棵靜態 Huffman 樹的樹形描述,依序 table0 / table1 / table2。
       每棵樹對 256 個 symbol 依 0..255 順序描述:
         5 bits  = 碼長 L(L==0 表示該 symbol 未使用)
         L bits  = 該 symbol 的 Huffman 碼(MSB first)
       解碼端一邊讀一邊長樹:節點編號從 0x100(根)起算,
       新節點取 0x101, 0x102, ...;葉子直接存 symbol 值(< 0x100)。
    2) 字圖資料:每字 72 bytes(24x24,每列 3 bytes),每個 byte 用
       「該 byte 在列內的欄位」對應的樹解碼 —— 也就是 table index 每
       輸出一個 byte 就 (idx+1)%3。因為 72 是 3 的倍數,每字開頭必回到 0。
    3) 每解完一個字,做垂直差分還原(壓縮端存的是與上一列的 XOR):
         for i in range(xor_start*3, xor_end*3):
             glyph[i+3] ^= glyph[i]
       xor_end 隨字體不同(明/圓/黑 = 22、楷/宋 = 20、隸 = 18),
       代表該字體實際用到的列數;超出範圍的末幾列不做差分。

  位元讀取:MSB first,byte 用完才取下一個 byte。

字數 13094 與每字 72 bytes 是 ETUNPACK.EXE 內寫死的常數
(0x3326 / 0x48),此處沿用。輸出 13094 * 72 = 942768 bytes。

用法:
    python3 etunpack.py STD.24M [輸出路徑]
"""

import struct
import sys

GLYPH_BYTES = 72        # 24x24,每列 3 bytes
NUM_GLYPHS = 13094      # ETUNPACK.EXE 內的 0x3326
ROW_BYTES = 3
NUM_TABLES = 3
ROOT = 0x100


class BitReader:
    """MSB-first 位元讀取器,對應 ETUNPACK.EXE 的 getbits()。"""

    def __init__(self, data):
        self.data = data
        self.pos = 0
        self.cur = 0
        self.left = 0
        self.overrun = False

    def bit(self):
        if self.left == 0:
            if self.pos >= len(self.data):
                # 碼流提前用完(STD.24L 隸書就少了最後 430 字)。
                # 原版 ETUNPACK.EXE 在這裡會繼續讀舊緩衝區的殘值,
                # 尾端本來就是垃圾;這裡改用 0 補,行為較可預期。
                self.overrun = True
                self.cur = 0
                self.left = 8
            else:
                self.cur = self.data[self.pos]
                self.pos += 1
                self.left = 8
        b = (self.cur >> 7) & 1
        self.cur = (self.cur << 1) & 0xFF
        self.left -= 1
        return b

    def bits(self, n):
        v = 0
        for _ in range(n):
            v = (v << 1) | self.bit()
        return v


def read_tree(br):
    """讀一棵樹形描述,回傳 (left, right) 兩個 dict:node -> 子節點/symbol。"""
    left = {}
    right = {}
    next_node = ROOT + 1
    for sym in range(256):
        length = br.bits(5)
        if length == 0:
            continue          # 該 symbol 未使用
        node = ROOT
        # 前 length-1 個 bit 走(必要時新增)內部節點
        for _ in range(length - 1):
            arr = right if br.bit() else left
            nxt = arr.get(node)
            if nxt is None:
                nxt = next_node
                next_node += 1
                arr[node] = nxt
            node = nxt
        # 最後 1 個 bit 掛上葉子
        arr = right if br.bit() else left
        arr[node] = sym
    return left, right


def decode(data, xor_start, xor_end,
           num_glyphs=NUM_GLYPHS, glyph_bytes=GLYPH_BYTES):
    br = BitReader(data)
    trees = [read_tree(br) for _ in range(NUM_TABLES)]

    out = bytearray()
    tbl = 0
    truncated_at = None
    lo = xor_start * ROW_BYTES
    hi = xor_end * ROW_BYTES

    for gi in range(num_glyphs):
        if br.overrun and truncated_at is None:
            truncated_at = gi
        g = bytearray(glyph_bytes)
        for i in range(glyph_bytes):
            left, right = trees[tbl]
            node = ROOT
            while True:
                arr = right if br.bit() else left
                node = arr[node]
                if node < 0x100:
                    break
            g[i] = node
            tbl = (tbl + 1) % NUM_TABLES
        # 垂直差分還原:本列 XOR 上一列
        for i in range(lo, hi):
            g[i + ROW_BYTES] ^= g[i]
        out += g
    return bytes(out), truncated_at


def unpack(path):
    raw = open(path, 'rb').read()
    if raw[:14] != b'ETUNPACK V1.00':
        raise ValueError("不是 ETUNPACK V1.00 檔案: %s" % path)
    dir_ofs, data_ofs = struct.unpack_from('<II', raw, 0x10)
    name = raw[dir_ofs + 4:dir_ofs + 16].split(b'\0')[0].decode('ascii')
    xor_start = raw[dir_ofs + 0x12]
    xor_end = raw[dir_ofs + 0x13]
    out, truncated_at = decode(raw[data_ofs:], xor_start, xor_end)
    return name, xor_start, xor_end, out, truncated_at


def main(argv):
    if len(argv) < 2:
        print(__doc__)
        return 1
    name, xs, xe, out, trunc = unpack(argv[1])
    dst = argv[2] if len(argv) > 2 else name
    with open(dst, 'wb') as f:
        f.write(out)
    print("%s -> %s  (%d bytes, %d chars, xor rows %d..%d)"
          % (argv[1], dst, len(out), len(out) // GLYPH_BYTES, xs, xe))
    if trunc is not None:
        print("警告:壓縮碼流在第 %d 字就用完,之後的字非有效資料" % trunc)
    return 0


if __name__ == '__main__':
    sys.exit(main(sys.argv))
