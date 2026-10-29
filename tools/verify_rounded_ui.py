#!/usr/bin/env python3
"""在 Docker 內核對新版介面證據。--reviewed 表示本輪已逐張人工檢視。"""
from __future__ import annotations

import argparse
import gzip
import hashlib
import json
import os
import struct
import subprocess
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "workplace/art-preview"


def sha(p: Path) -> str:
    h = hashlib.sha256()
    with p.open("rb") as f:
        for block in iter(lambda: f.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--reviewed", action="store_true")
    args = parser.parse_args()
    assert args.reviewed, "請先檢視 GUI 圖片，再用 --reviewed 保存人工檢視結果"
    approved = json.loads((ROOT / "artwork/manifests/ui-rounded-approved.json").read_text())
    ref = ROOT / approved["reference_png"]
    assert sha(ref) == approved["sha256"]
    fonts = json.loads((ROOT / "internal/presentation/fonts/sources.json").read_text())
    for source in fonts["sources"]:
        p = ROOT / source["embedded_path"]
        assert sha(p) == source["embedded_sha256"], p
        raw = gzip.decompress(p.read_bytes()) if p.suffix == ".gz" else p.read_bytes()
        assert hashlib.sha256(raw).hexdigest() == source["sha256"], p

    gui_dir = OUT / "ui-gui"
    gui = json.loads((gui_dir / "result.json").read_text())
    assert gui["status"] == "passed" and gui["exit_code"] == 0
    assert gui["map_load_restored"] and gui["quit_blocks_toolbar"]
    assert sha(ROOT / "workplace/game-test/wasteland-ui-dev") == gui["binary_sha256"]
    for f in gui["files"]:
        assert sha(gui_dir / f["path"]) == f["sha256"], f
        assert Image.open(gui_dir / f["path"]).size == (960, 720), f

    a = Image.open(ref).convert("RGB")
    b = Image.open(gui_dir / "09-loaded.png").convert("RGB")
    viewport = tuple(approved["map_rect"])
    assert a.crop(viewport).tobytes() == b.crop(viewport).tobytes(), "定版與實作地圖像素不同"
    comparison = Image.new("RGB", (1968, 792), "#101820")
    font = ImageFont.truetype(str(OUT / "ui-redesign-NotoSansCJKtc-Regular.otf"), 22)
    d = ImageDraw.Draw(comparison)
    d.text((24, 20), "定版草稿 · B 圓體 18 px", font=font, fill="#e2c58c")
    d.text((1008, 20), "實際遊戲 · 相同讀檔畫面", font=font, fill="#e2c58c")
    comparison.paste(a, (16, 64))
    comparison.paste(b, (992, 64))
    picture = OUT / "ui-implemented-vs-approved.png"
    comparison.save(picture)

    binaries = []
    for name, platform in [
        ("wasteland-ui-dev", "linux-amd64"),
        ("wasteland-ui-dev-windows.exe", "windows-amd64"),
        ("wasteland-ui-dev-darwin-amd64", "darwin-amd64"),
        ("wasteland-ui-dev-darwin-arm64", "darwin-arm64"),
    ]:
        p = ROOT / "workplace/game-test" / name
        with p.open("rb") as f:
            header = f.read(65536)
        if platform.startswith("linux"):
            assert header[:4] == b"\x7fELF" and struct.unpack_from("<H", header, 18)[0] == 62
        elif platform.startswith("windows"):
            assert header[:2] == b"MZ"
            pe = struct.unpack_from("<I", header, 0x3C)[0]
            assert header[pe:pe+4] == b"PE\0\0" and struct.unpack_from("<H", header, pe+4)[0] == 0x8664
        else:
            assert header[:4] == b"\xcf\xfa\xed\xfe"
            cpu = struct.unpack_from("<I", header, 4)[0]
            assert cpu == (0x1000007 if platform.endswith("amd64") else 0x100000C)
            ncmds = struct.unpack_from("<I", header, 16)[0]
            at, minimum = 32, []
            for _ in range(ncmds):
                cmd, size = struct.unpack_from("<II", header, at)
                if cmd == 0x32:
                    minimum.append(struct.unpack_from("<I", header, at+12)[0])
                elif cmd == 0x24:
                    minimum.append(struct.unpack_from("<I", header, at+8)[0])
                at += size
            assert minimum and all(v == 0xB0000 for v in minimum), (platform, minimum)
        binaries.append({"path": str(p.relative_to(ROOT)), "platform": platform,
                         "bytes": p.stat().st_size, "sha256": sha(p)})

    old_root = ROOT / "dist-all/v.1.0.0-20261029"
    old = json.loads((old_root / "SHA256SUMS.json").read_text())
    for package in old["packages"]:
        assert sha(old_root / package["path"]) == package["sha256"]

    inputs = [ROOT / p for p in [
        "go.mod", "go.sum", "cmd/wasteland/main.go", "internal/play/play.go",
        "internal/play/charview.go", "internal/play/presentation.go", "internal/play/presentation_test.go",
        "internal/render/hires.go", "internal/ui/ui.go", "internal/presentation/canvas.go",
        "internal/presentation/canvas_test.go", "translations/zh-Hant.cat",
        "tools/check_rounded_ui_gui.sh", "tools/verify_rounded_ui.py",
        "artwork/manifests/ui-rounded-approved.json", "docs/spec/33-rounded-ui.md",
    ]] + sorted((ROOT / "internal/presentation/fonts").iterdir())
    reviewed = ["00-title", "02-noto-16", "03-noto-20", "04-hd2-a", "09-loaded", "18-combat", "21-ranger-center"]
    review = {
        "status": "passed", "scope": "approved modern UI; local development binaries, not a release",
        "source_commit": subprocess.check_output(["git", "rev-parse", "HEAD"], cwd=ROOT, text=True).strip(),
        "validation_tracked_changes": bool(subprocess.check_output(["git", "status", "--porcelain", "--untracked-files=no"], cwd=ROOT, text=True).strip()),
        "tool_images": {key: os.environ.get(key, "not-recorded") for key in ("WL_GO_IMAGE_ID", "WL_OSXCROSS_IMAGE_ID")},
        "manual_review": reviewed, "gui_receipt_sha256": sha(gui_dir / "result.json"),
        "approved_reference_sha256": sha(ref), "map_pixels_equal_approved_reference": True,
        "expected_differences": ["字型光柵器抗鋸齒", "輻射欄依實際儀器與讀數顯示；草稿為樣本數值"],
        "comparison": {"path": str(picture.relative_to(ROOT)), "sha256": sha(picture)},
        "binaries": binaries, "old_release_packages_unchanged": len(old["packages"]),
        "native_windows_macos": "not-run", "packaging": "not rebuilt; no new version or publication",
        "full_suite_log": "workplace/game-test/ui-rounded-go-checks.txt",
        "final_ui_tests_and_vet_log": "workplace/game-test/ui-rounded-final-targeted.txt",
        "inputs": [{"path": str(p.relative_to(ROOT)), "sha256": sha(p)} for p in inputs],
    }
    dest = OUT / "ui-implementation-review.json"
    dest.write_text(json.dumps(review, ensure_ascii=False, indent=2)+"\n")
    for p in [picture, dest]:
        assert (p.stat().st_uid, p.stat().st_gid) == (1000, 1000), p
    print("PASS: font sources/licenses, GUI hashes, identical map viewport, four binaries and six preserved packages.")
    print(dest.relative_to(ROOT))


if __name__ == "__main__":
    main()
