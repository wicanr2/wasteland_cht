#!/usr/bin/env python3
"""在分析映像裡掃某個函式的直接呼叫點（far call ＋ 同段 near call）。

不需要 IDA。用途是替 IDA 的 xref 做一次獨立的正對照——尤其是要下
「這個東西沒有人呼叫」這種結論的時候，全檔掃描比 xref 圖說得清楚。

    python3 tools/scan_callers.py <映像> <段>:<位移>

段要**自己查**（IDA 的 `segment:offset` 那一欄），不接受只給線性位址——
函式位址推不出它屬於哪個段，`0x1CBD3` 猜成 `0x1CBD:0x0003` 會掃到零筆。

⚠ 兩個會製造安靜零命中的坑（`docs/re/44` §10）：

1. **檔案裡存的是重定位前的段值。** MZ 的 relocation table 會在載入時把
   `0x1000` 加上去，所以執行期的 `0x1CB6` 在檔案裡是 `0x0CB6`。
   拿執行期的段值去掃，一筆都不會中，而且長得和「真的沒人呼叫」一模一樣。
2. **段基底不能從函式位址推。** `0x1CBD3` 屬於段 `0x1CB6`（位移 `0x73`），
   不是 `0x1CBD:0x0003`。這兩種寫法指的是同一個線性位址，但機器碼裡存的是
   段與位移，猜錯就掃到零筆——而且一樣是安靜的零。

抓不到間接呼叫（`call far ptr [x]`）。下「沒有呼叫端」的結論時要一起講。
"""

import re
import struct
import sys


def scan(path: str, seg: int, off: int) -> int:
    img = open(path, "rb").read()
    hdr = struct.unpack_from("<H", img, 8)[0] * 16

    def linear(file_off: int) -> int:
        return file_off - hdr + 0x10000

    disk_seg = seg - 0x1000  # 重定位前的段值

    print(f"目標 {seg * 16 + off:#07x} ＝ {seg:#06x}:{off:#06x}")
    print(f"檔案裡的段值（重定位前）{disk_seg:#06x}")
    print()

    pat = bytes([0x9A]) + struct.pack("<HH", off, disk_seg)
    far = [m.start() for m in re.finditer(re.escape(pat), img)]
    print(f"far call：{len(far)} 筆")
    for f in far:
        print(f"  檔案 {f:#07x}  線性 {linear(f):#07x}  前 8 bytes {img[f - 8:f].hex(' ')}")

    # 同段的 near call：只有目標所在的那個段裡的程式碼跳得到。
    flo = seg * 16 - 0x10000 + hdr
    fhi = flo + 0x10000
    near = []
    for f in range(max(flo, 0), min(fhi, len(img) - 3)):
        if img[f] != 0xE8:
            continue
        rel = struct.unpack_from("<h", img, f + 1)[0]
        if ((f + 3 - flo) + rel) & 0xFFFF == off:
            near.append(f)
    print(f"\n同段 near call：{len(near)} 筆")
    for f in near:
        print(f"  檔案 {f:#07x}  線性 {linear(f):#07x}  前 8 bytes {img[f - 8:f].hex(' ')}")

    total = len(far) + len(near)
    if total == 0:
        # 零命中與「段給錯」長得一模一樣，所以講清楚而不是靜靜回 0。
        print("\n⚠ 零命中。下「沒有呼叫端」的結論之前，先拿一個**已知有呼叫端**的"
              "位址跑同一支腳本做正對照，確認段與位移沒給錯。")
    return total


if __name__ == "__main__":
    if len(sys.argv) != 3 or ":" not in sys.argv[2]:
        sys.exit(__doc__)
    a, b = sys.argv[2].split(":")
    scan(sys.argv[1], int(a, 16), int(b, 16))
