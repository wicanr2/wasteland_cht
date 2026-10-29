#!/usr/bin/env python3
"""Verify the actual versioned package; run in the packaging Docker image."""
import argparse
import hashlib
import json
from pathlib import Path
import plistlib
import re
import struct
import subprocess
import tempfile
import zipfile

REPO = Path(__file__).resolve().parents[1]


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def architectures(path):
    data = path.read_bytes()
    if data[:4] == b"\x7fELF":
        assert data[4:6] == b"\x02\x01" and struct.unpack_from("<H", data, 18)[0] == 62
        return ["linux-amd64"]
    if data[:2] == b"MZ":
        pe = struct.unpack_from("<I", data, 0x3C)[0]
        assert data[pe:pe + 4] == b"PE\0\0" and struct.unpack_from("<H", data, pe + 4)[0] == 0x8664
        return ["windows-amd64"]
    assert data[:4] == b"\xca\xfe\xba\xbe", path
    count = struct.unpack_from(">I", data, 4)[0]
    cpus = {struct.unpack_from(">I", data, 8 + i * 20)[0] for i in range(count)}
    assert count == 2 and cpus == {0x01000007, 0x0100000C}, (path, cpus)
    return ["darwin-amd64", "darwin-arm64"]


def verify(package, version, mode, platform):
    assert re.fullmatch(r"v\.\d+\.\d+\.\d+-\d{8}", version), version
    pkg = f"wasteland-cht-{version}-{platform}-{mode}"
    assert package.name == pkg + (".AppImage" if platform == "linux-x64" else ".zip")
    result = {"schema": 1, "version": version, "mode": mode, "platform": platform,
              "package": str(package), "bytes": package.stat().st_size, "sha256": sha(package),
              "rights": "local-original-derived" if mode == "full-local" else "engine-and-translations"}
    with tempfile.TemporaryDirectory(prefix="wl-package-verify-") as temp:
        temp = Path(temp)
        if platform == "linux-x64":
            subprocess.run([str(package.resolve()), "--appimage-extract"], cwd=temp,
                           stdout=subprocess.DEVNULL, check=True, timeout=60)
            root = temp / "squashfs-root/usr"
            assert (temp / "squashfs-root/AppRun").stat().st_mode & 0o111
        else:
            with zipfile.ZipFile(package) as archive, package.open("rb") as raw:
                assert archive.testzip() is None
                for info in archive.infolist():
                    parts = Path(info.filename).parts
                    assert parts and parts[0] == pkg and ".." not in parts and not info.filename.startswith("/")
                    assert info.flag_bits & 0x800, info.filename
                    raw.seek(info.header_offset + 6)
                    assert struct.unpack("<H", raw.read(2))[0] & 0x800, info.filename
                    if platform == "macos-universal" and any(s in info.filename for s in ("/bin/", "/MacOS/")) and not info.is_dir():
                        assert (info.external_attr >> 16) & 0o111, info.filename
                archive.extractall(temp)
                result["zip_entries"] = len(archive.infolist())
            root = temp / pkg
            if platform == "macos-universal":
                app = root / "荒野遊俠.app/Contents"
                plist = plistlib.loads((app / "Info.plist").read_bytes())
                assert plist["WastelandReleaseVersion"] == version
                assert plist["CFBundleShortVersionString"] == version[2:].rsplit("-", 1)[0]
                root = app / "Resources"
            else:
                batch = (root / "開始遊戲.bat").read_bytes()
                assert b"\n" not in batch.replace(b"\r\n", b"")
                subprocess.run(["python3", str(REPO / "tools/verify_windows_zip.py"), str(package)], check=True)
        manifest = (root / "PACKAGE-MANIFEST.txt").read_text()
        assert f"Version: {version}\n" in manifest and f"Package: {pkg}\n" in manifest
        assert "LicenseRef-RRSAL-1.0" in (root / "LICENSE").read_text()
        result["build_manifest"] = manifest
        suffix = ".exe" if platform == "windows-x64" else ""
        result["architectures"] = {}
        for name in ("wasteland", "wl-setup"):
            binary = root / "bin" / (name + suffix)
            result["architectures"][name] = architectures(binary)
            assert result["architectures"][name] == {
                "linux-x64": ["linux-amd64"], "windows-x64": ["windows-amd64"],
                "macos-universal": ["darwin-amd64", "darwin-arm64"],
            }[platform]
            assert version.encode() in binary.read_bytes(), binary
        for name in ("zh-Hant.cat", "paragraphs-zh-Hant.cat"):
            assert sha(root / "translations" / name) == sha(REPO / "translations" / name)
        assert not any((root / "artpacks" / old).exists() for old in ("faithful-hd", "reimagined"))
        if mode == "patch":
            assert not any((root / "artpacks").rglob("*.*"))
            assert not any((root / name).exists() for name in ("data", "eten", "music", "build"))
            forbidden = {"game1", "game2", "wl.exe", "wla.bin", "curs", "transtbl", "paragraphs.txt"}
            for path in root.rglob("*"):
                if path.is_file():
                    assert path.name.lower() not in forbidden, path
                    assert path.suffix.lower() not in {".png", ".ogg", ".wlf", ".fnt", ".pic", ".cpa"}, path
        else:
            assert (root / "build/wl.merged.exe").is_file()
            assert sha(root / "build/wl.merged.exe") == "cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118"
            for name in ("wl.exe", "wla.bin", "game1", "game2"):
                assert sha(root / "data" / name) == sha(REPO / "workplace/orig/wastland" / name)
            assert (root / "eten/STDFONT.24").stat().st_size > 0
            result["art_assets"] = {}
            for style in ("hd2-a", "hd2-b"):
                art = root / "artpacks" / style
                assert sha(art / "manifest.json") == sha(REPO / "artpacks" / style / "manifest.json")
                art_manifest = json.loads((art / "manifest.json").read_text())
                assert art_manifest["id"] == style and len(art_manifest["assets"]) == 1157
                assert len(list((art / "assets").rglob("*.png"))) == 1157
                for asset in art_manifest["assets"]:
                    path = art / asset["path"]
                    assert sha(path) == asset["sha256"], path
                    assert struct.unpack(">II", path.read_bytes()[16:24]) == (asset["width"], asset["height"])
                result["art_assets"][style] = 1157
            result["music"] = {}
            for variant in ("retro", "modern"):
                songs = sorted((root / "music" / variant).glob("*.ogg"))
                assert len(songs) == 10
                result["music"][variant] = {p.name: sha(p) for p in songs}
                for song in songs:
                    assert sha(song) == sha(REPO / "workplace/music" / variant / song.name)
            assert (root / "music/modern/FluidR3-GM-LICENSE.txt").is_file()
        result["status"] = "verified-package-content-and-architecture"
    return result


if __name__ == "__main__":
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("package", type=Path)
    ap.add_argument("--version", required=True)
    ap.add_argument("--mode", choices=["patch", "full-local"], required=True)
    ap.add_argument("--platform", choices=["linux-x64", "windows-x64", "macos-universal"], required=True)
    ap.add_argument("--receipt", type=Path, required=True)
    args = ap.parse_args()
    result = verify(args.package, args.version, args.mode, args.platform)
    args.receipt.parent.mkdir(parents=True, exist_ok=True)
    args.receipt.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
    print(f"PASS {args.mode} {args.platform}: {result['bytes']} bytes {result['sha256']}")
