#!/usr/bin/env python3
"""Verify the final Windows x86_64 ZIP, including UTF-8 and DLL closure."""

from __future__ import annotations

import struct
import sys
import zipfile

UTF8 = 0x800
SYSTEM_DLLS = {
    "advapi32.dll", "bcrypt.dll", "comdlg32.dll", "d3d11.dll",
    "d3dcompiler_47.dll", "dinput8.dll", "dxgi.dll", "gdi32.dll",
    "imm32.dll", "kernel32.dll", "ntdll.dll", "ole32.dll", "oleaut32.dll",
    "opengl32.dll", "rpcrt4.dll", "secur32.dll", "setupapi.dll",
    "shell32.dll", "shcore.dll", "user32.dll", "userenv.dll", "uuid.dll",
    "version.dll", "winmm.dll", "ws2_32.dll", "xinput1_4.dll", "xinput9_1_0.dll",
}


def u16(data: bytes, offset: int) -> int:
    return struct.unpack_from("<H", data, offset)[0]


def u32(data: bytes, offset: int) -> int:
    return struct.unpack_from("<I", data, offset)[0]


def pe_imports(data: bytes) -> tuple[int, list[str]]:
    if data[:2] != b"MZ":
        raise ValueError("不是 MZ")
    pe = u32(data, 0x3C)
    if data[pe:pe + 4] != b"PE\0\0":
        raise ValueError("不是 PE")
    machine = u16(data, pe + 4)
    nsec = u16(data, pe + 6)
    opt = pe + 24
    dd = opt + (112 if u16(data, opt) == 0x20B else 96)
    imp_rva = u32(data, dd + 8)
    sec = pe + 24 + u16(data, pe + 20)
    sections = []
    for i in range(nsec):
        pos = sec + i * 40
        sections.append((u32(data, pos + 12), u32(data, pos + 16), u32(data, pos + 20)))

    def offset(rva: int):
        for va, size, pointer in sections:
            if va <= rva < va + max(size, 1):
                return pointer + rva - va
        return None

    names: list[str] = []
    pos = offset(imp_rva)
    while pos is not None:
        name_rva = u32(data, pos + 12)
        if name_rva == 0 and u32(data, pos) == 0:
            break
        name_pos = offset(name_rva)
        if name_pos is None:
            raise ValueError("匯入 DLL 名稱 RVA 無法定址")
        end = data.index(b"\0", name_pos)
        names.append(data[name_pos:end].decode("ascii"))
        pos += 20
    return machine, sorted(set(names), key=str.lower)


def main() -> None:
    path = sys.argv[1]
    with zipfile.ZipFile(path) as archive:
        infos = archive.infolist()
        if not infos:
            raise SystemExit("ZIP 是空的")
        bad_utf8 = [info.filename for info in infos if not info.flag_bits & UTF8]
        if bad_utf8:
            raise SystemExit(f"{len(bad_utf8)} 筆未設定 UTF-8 bit 11")
        names = {info.filename for info in infos}
        roots = {name.split("/", 1)[0] for name in names if "/" in name}
        if len(roots) != 1:
            raise SystemExit(f"ZIP 頂層不唯一：{sorted(roots)}")
        root = next(iter(roots))
        required = {
            f"{root}/bin/wasteland.exe", f"{root}/bin/wl-setup.exe",
            f"{root}/開始遊戲.bat", f"{root}/Windows-DLL說明.txt",
            f"{root}/artpacks/faithful-hd/manifest.json",
            f"{root}/artpacks/reimagined/manifest.json",
        }
        missing = required - names
        if missing:
            raise SystemExit(f"缺少必要檔案：{sorted(missing)}")
        is_full = "-local-full-" in root
        for folder in ("data", "eten", "music", "build"):
            prefix = f"{root}/{folder}/"
            present = any(name.startswith(prefix) and not name.endswith("/") for name in names)
            if is_full and not present:
                raise SystemExit(f"完整版缺少 {folder}/ 的實際檔案")
            if not is_full and present:
                raise SystemExit(f"公開版不應含有 {folder}/")

        batch = archive.read(f"{root}/開始遊戲.bat")
        if b"\n" in batch.replace(b"\r\n", b""):
            raise SystemExit("開始遊戲.bat 含非 CRLF 換行")

        bundled_dlls = {name.rsplit("/", 1)[-1].lower() for name in names if name.lower().endswith(".dll")}
        report = []
        for exe in ("wasteland.exe", "wl-setup.exe"):
            machine, imports = pe_imports(archive.read(f"{root}/bin/{exe}"))
            if machine != 0x8664:
                raise SystemExit(f"{exe} machine={machine:#x}，不是 x86_64")
            external = []
            for dll in imports:
                low = dll.lower()
                if low in SYSTEM_DLLS or low.startswith(("api-ms-win-", "ext-ms-win-")):
                    continue
                if low not in bundled_dlls:
                    external.append(dll)
            if external:
                raise SystemExit(f"{exe} 缺少非系統 DLL：{external}")
            report.append(f"{exe}: {', '.join(imports) or '(no imports)'}")
    print(f"[windows-zip] PASS entries={len(infos)} utf8={len(infos)} bundled_dlls={len(bundled_dlls)}")
    for line in report:
        print(f"  {line}")


if __name__ == "__main__":
    main()
