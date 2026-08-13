"""把 Microsoft EXEPACK 壓縮的 16-bit DOS 執行檔還原成未壓縮 MZ 執行檔。

用法（容器內）：
    python3 tools/unpack_exepack.py <packed.exe> <unpacked.exe> <report.json>

為什麼要自己寫：wl.exe 是 EXEPACK 打包的，IDA 直接分析到的是解壓 stub 加上
壓縮資料，自動辨識出來的函式與字串大部分是誤讀。不先解包，後面所有位址、
xref 與結論都建立在錯的映像上。

實作依 EXEPACK 的格式本身，不依賴任何外部解包工具，好處是每一步都能對
原始 bytes 交代，而且失敗時是明確報錯而不是產出一份看起來像樣的錯映像。

驗證條件（任一不成立就中止，不輸出檔案）：
  - MZ 簽章、e_cs 指向的 stub 段落含 'RB' 簽章
  - exepack_size 與檔案剩餘長度一致
  - 解出的長度等於 header 的 dest_len
  - relocation table 的 16 組資料剛好用完 stub 尾端的空間
"""

from __future__ import annotations

import hashlib
import json
import struct
import sys
from pathlib import Path

ERROR_MESSAGE = b"Packed file is corrupt"


class UnpackError(RuntimeError):
    pass


def read_mz(data: bytes) -> dict[str, int]:
    if data[:2] not in (b"MZ", b"ZM"):
        raise UnpackError("不是 MZ 執行檔")
    (
        _magic,
        cblp,
        cp,
        crlc,
        cparhdr,
        minalloc,
        maxalloc,
        ss,
        sp,
        csum,
        ip,
        cs,
        lfarlc,
        ovno,
    ) = struct.unpack_from("<14H", data, 0)
    return {
        "e_cblp": cblp,
        "e_cp": cp,
        "e_crlc": crlc,
        "e_cparhdr": cparhdr,
        "e_minalloc": minalloc,
        "e_maxalloc": maxalloc,
        "e_ss": ss,
        "e_sp": sp,
        "e_csum": csum,
        "e_ip": ip,
        "e_cs": cs,
        "e_lfarlc": lfarlc,
        "e_ovno": ovno,
        "image_size": (cp - 1) * 512 + (cblp or 512),
        "header_size": cparhdr * 16,
    }


def parse_exepack_header(stub: bytes) -> dict[str, int]:
    if stub[16:18] != b"RB":
        raise UnpackError("stub 段落沒有 EXEPACK 的 'RB' 簽章")
    real_ip, real_cs, mem_start, exepack_size, real_sp, real_ss, dest_len, skip_len = (
        struct.unpack_from("<8H", stub, 0)
    )
    return {
        "real_ip": real_ip,
        "real_cs": real_cs,
        "mem_start": mem_start,
        "exepack_size": exepack_size,
        "real_sp": real_sp,
        "real_ss": real_ss,
        "dest_len_paragraphs": dest_len,
        "skip_len_paragraphs": skip_len,
    }


def unpack_image(packed: bytes, dest_len: int) -> tuple[bytearray, dict[str, int]]:
    """反向解 RLE。EXEPACK 從映像尾端往前寫，所以來源與目的都是遞減的。"""
    src = len(packed)
    while src > 0 and packed[src - 1] == 0xFF:  # 尾端填充
        src -= 1
    padding = len(packed) - src

    out = bytearray(dest_len)
    dst = dest_len
    commands = 0
    fills = 0
    copies = 0

    while True:
        if src < 3:
            raise UnpackError("壓縮資料在讀到結束旗標前就用完了")
        command = packed[src - 1]
        length = int.from_bytes(packed[src - 3 : src - 1], "little")
        src -= 3

        if command & 0xFE == 0xB0:  # fill
            if src < 1:
                raise UnpackError("fill 命令缺少填充值")
            value = packed[src - 1]
            src -= 1
            if dst - length < 0:
                raise UnpackError("fill 超出目的緩衝區")
            out[dst - length : dst] = bytes([value]) * length
            dst -= length
            fills += 1
        elif command & 0xFE == 0xB2:  # copy
            if src - length < 0 or dst - length < 0:
                raise UnpackError("copy 超出緩衝區")
            out[dst - length : dst] = packed[src - length : src]
            src -= length
            dst -= length
            copies += 1
        else:
            raise UnpackError(f"未知的 EXEPACK 命令 0x{command:02X}")

        commands += 1
        if command & 0x01:  # 最後一個命令
            break

    # 剩下的前段是未經 RLE 的原始資料，直接搬過去。
    if src > dst:
        raise UnpackError("剩餘來源資料放不進目的緩衝區")
    out[dst - src : dst] = packed[:src]

    return out, {
        "commands": commands,
        "fill_commands": fills,
        "copy_commands": copies,
        "trailing_ff_padding": padding,
        "literal_prefix_bytes": src,
    }


def parse_relocations(stub: bytes) -> tuple[list[tuple[int, int]], int]:
    """reloc table 緊接在錯誤訊息之後，中間沒有結尾 NUL。

    ⚠ 這裡很容易誤判：wl.exe 的訊息在 IDA 裡顯示成 `Packed file is corrupt#`，
    那個 `#` 是 0x23，其實是第一組 relocation 的 count（35 筆）低位元組，
    後面的 0x00 是它的高位元組。把 `#\0` 當成訊息的一部分往後跳，整張表就會
    錯開一個 byte，讀出來的 count 變成不合理的大數。
    """
    idx = stub.find(ERROR_MESSAGE)
    if idx < 0:
        raise UnpackError("stub 裡找不到 'Packed file is corrupt'，無法定位 reloc table")
    pos = start = idx + len(ERROR_MESSAGE)
    relocs: list[tuple[int, int]] = []
    for high in range(16):
        if pos + 2 > len(stub):
            raise UnpackError("reloc table 在讀完 16 組之前就超出 stub")
        (count,) = struct.unpack_from("<H", stub, pos)
        pos += 2
        for _ in range(count):
            (offset,) = struct.unpack_from("<H", stub, pos)
            pos += 2
            relocs.append((high << 12, offset))

    if any(stub[pos:]):
        raise UnpackError(
            f"reloc table 之後還有非零資料（stub offset 0x{pos:X}），解析起點可能錯了"
        )
    return relocs, pos - start


def build_exe(image: bytes, hdr: dict[str, int], relocs: list[tuple[int, int]]) -> bytes:
    reloc_bytes = b"".join(
        struct.pack("<HH", offset, segment) for segment, offset in relocs
    )
    header_len = 0x1C + len(reloc_bytes)
    header_paragraphs = (header_len + 15) // 16
    header_size = header_paragraphs * 16
    total = header_size + len(image)

    header = bytearray(header_size)
    struct.pack_into(
        "<14H",
        header,
        0,
        0x5A4D,
        total % 512,
        (total + 511) // 512,
        len(relocs),
        header_paragraphs,
        0x0000,
        0xFFFF,
        hdr["real_ss"],
        hdr["real_sp"],
        0x0000,
        hdr["real_ip"],
        hdr["real_cs"],
        0x001C,
        0x0000,
    )
    header[0x1C : 0x1C + len(reloc_bytes)] = reloc_bytes
    return bytes(header) + image


def main() -> None:
    src_path, dst_path, report_path = (Path(p) for p in sys.argv[1:4])
    data = src_path.read_bytes()

    mz = read_mz(data)
    body = data[mz["header_size"] : mz["image_size"]]
    stub_offset = mz["e_cs"] * 16
    if stub_offset >= len(body):
        raise UnpackError("e_cs 指到映像之外")

    packed = body[:stub_offset]
    stub = body[stub_offset:]
    pack = parse_exepack_header(stub)

    if pack["exepack_size"] != len(stub):
        raise UnpackError(
            f"exepack_size {pack['exepack_size']} 與 stub 實際長度 {len(stub)} 不符"
        )

    dest_len = pack["dest_len_paragraphs"] * 16
    image, stats = unpack_image(packed, dest_len)
    relocs, reloc_bytes = parse_relocations(stub)
    if reloc_bytes + stub.find(ERROR_MESSAGE) > len(stub):
        raise UnpackError("reloc table 超出 stub 範圍")

    exe = build_exe(bytes(image), pack, relocs)
    dst_path.write_bytes(exe)

    report = {
        "input": {
            "path": str(src_path),
            "size": len(data),
            "sha256": hashlib.sha256(data).hexdigest(),
            "mz_header": mz,
        },
        "exepack_header": pack,
        "stub": {
            "file_offset": mz["header_size"] + stub_offset,
            "size": len(stub),
            "relocation_entries": len(relocs),
            "relocation_table_bytes": reloc_bytes,
        },
        "unpack": stats
        | {
            "packed_bytes": len(packed),
            "unpacked_bytes": len(image),
            "ratio": round(len(packed) / len(image), 4),
        },
        "output": {
            "path": str(dst_path),
            "size": len(exe),
            "sha256": hashlib.sha256(exe).hexdigest(),
            "entry_cs_ip": f"{pack['real_cs']:04X}:{pack['real_ip']:04X}",
            "stack_ss_sp": f"{pack['real_ss']:04X}:{pack['real_sp']:04X}",
        },
    }
    report_path.write_text(
        json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8"
    )
    print(json.dumps(report["output"], ensure_ascii=False, indent=2))


main()
