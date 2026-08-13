"""把 `wla.bin` 疊到解包映像的 `CS:0000`，產生執行時期的合成映像。

為什麼要這一步：`start` 開機時把 `wla.bin` 讀進 `CS:0000`（見 docs/re/03 §5），
之後 `call` 過去。直接分析 `wl.unpacked.exe` 看到的那 8 KB 是**還沒被覆蓋的**
原始內容，分析它等於分析一段永遠不會執行的程式碼。

輸出的 `wl.merged.exe` 是本專案合成的分析對象，不是原版檔案，
所以另外記一份雜湊，並在報告裡與原版明確區分。

用法（容器內）：
    python3 tools/apply_overlay.py <wl.unpacked.exe> <wla.bin> <wl.merged.exe> <report.json>
"""

from __future__ import annotations

import hashlib
import json
import struct
import sys
from pathlib import Path


def main() -> None:
    base_path, overlay_path, out_path, report_path = (Path(p) for p in sys.argv[1:5])
    base = bytearray(base_path.read_bytes())
    overlay = overlay_path.read_bytes()

    if base[:2] != b"MZ":
        raise SystemExit("基底不是 MZ 執行檔")
    header_size = struct.unpack_from("<H", base, 0x08)[0] * 16
    if header_size + len(overlay) > len(base):
        raise SystemExit("overlay 超出映像範圍")

    original = bytes(base[header_size : header_size + len(overlay)])
    base[header_size : header_size + len(overlay)] = overlay
    out_path.write_bytes(bytes(base))

    differing = sum(1 for a, b in zip(original, overlay) if a != b)
    report = {
        "base": {
            "path": str(base_path),
            "sha256": hashlib.sha256(base_path.read_bytes()).hexdigest(),
            "header_size": header_size,
        },
        "overlay": {
            "path": str(overlay_path),
            "size": len(overlay),
            "sha256": hashlib.sha256(overlay).hexdigest(),
            "load_address_linear": "0x10000",
            "covers_linear": f"0x10000-0x{0x10000 + len(overlay):X}",
        },
        "replaced_region": {
            "original_sha256": hashlib.sha256(original).hexdigest(),
            "differing_bytes": differing,
            "identical": differing == 0,
            "original_head_hex": original[:32].hex(),
            "overlay_head_hex": overlay[:32].hex(),
            "original_nonzero_bytes": sum(1 for b in original if b),
        },
        "output": {
            "path": str(out_path),
            "size": len(base),
            "sha256": hashlib.sha256(bytes(base)).hexdigest(),
            "note": "本專案合成的分析映像，不是原版檔案",
        },
    }
    report_path.write_text(
        json.dumps(report, ensure_ascii=False, indent=2), encoding="utf-8"
    )
    print(json.dumps(report["replaced_region"] | report["output"], ensure_ascii=False, indent=2))


main()
