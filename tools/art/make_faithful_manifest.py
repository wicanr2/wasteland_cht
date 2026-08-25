#!/usr/bin/env python3
"""Build the complete validated faithful-hd bundle manifest."""

import hashlib
import json
from pathlib import Path

from PIL import Image


ROOT = Path("artpacks/faithful-hd")
COUNTS = [66, 141, 163, 107, 127, 118, 90, 104, 135]


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def asset(asset_id: str, kind: str, rel: Path, size: tuple[int, int]) -> dict:
    path = ROOT / rel
    if not path.is_file():
        raise SystemExit(f"missing: {path}")
    with Image.open(path) as im:
        if im.format != "PNG" or im.size != size:
            raise SystemExit(f"invalid {path}: {im.format} {im.size}, want PNG {size}")
    return {
        "id": asset_id,
        "kind": kind,
        "path": str(rel),
        "width": size[0],
        "height": size[1],
        "sha256": digest(path),
    }


def main() -> None:
    assets = []
    for tileset, count in enumerate(COUNTS):
        for index in range(count):
            rel = Path("assets") / f"tileset-{tileset}" / f"{index:03d}.png"
            assets.append(asset(f"tile.{tileset}.{index:03d}", "tile", rel, (48, 48)))
    for index in range(10):
        rel = Path("assets/icons") / f"{index:03d}.png"
        assets.append(asset(f"icon.{index:03d}", "icon", rel, (48, 48)))
    for index in range(12):
        rel = Path("assets/party-walk") / f"{index:03d}.png"
        assets.append(asset(f"party.walk.{index:03d}", "party-sprite", rel, (48, 48)))
    for index in range(82):
        rel = Path("assets/scenes") / f"{index:03d}.png"
        assets.append(asset(f"scene.{index:03d}", "scene", rel, (288, 252)))
    assets.append(asset("fullscreen.title", "title", Path("assets/fullscreen/title.png"), (864, 384)))
    assets.append(asset("fullscreen.ending", "ending", Path("assets/fullscreen/ending.png"), (864, 384)))

    manifest = {
        "schema": 1,
        "id": "faithful-hd-v1",
        "mode": "faithful-hd",
        "canvas": {
            "width": 960,
            "height": 720,
            "responsive": False,
        },
        "assets": assets,
    }
    (ROOT / "manifest.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n"
    )
    print(f"PASS assets={len(assets)} canvas=960x720 fixed-4:3")


if __name__ == "__main__":
    main()
