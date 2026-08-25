#!/usr/bin/env python3
"""Validate WIP modern tiles against the authoritative original inventory."""

import argparse
import json
from pathlib import Path

from PIL import Image


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--inventory", type=Path, default=Path("artwork/manifests/original-art-inventory.json"))
    ap.add_argument("--root", type=Path, default=Path("artpacks/faithful-hd/assets"))
    ap.add_argument("--tile-size", type=int, default=48)
    ap.add_argument("--tileset", type=int, action="append")
    args = ap.parse_args()

    inv = json.loads(args.inventory.read_text(encoding="utf-8"))
    selected = set(args.tileset or [t["id"] for t in inv["tilesets"]])
    failed = False
    for spec in inv["tilesets"]:
        tid = spec["id"]
        if tid not in selected:
            continue
        directory = args.root / f"tileset-{tid}"
        missing = []
        bad = []
        for index in range(spec["count"]):
            path = directory / f"{index:03d}.png"
            if not path.is_file():
                missing.append(index)
                continue
            try:
                with Image.open(path) as im:
                    if im.format != "PNG" or im.size != (args.tile_size, args.tile_size):
                        bad.append(f"{index}:{im.format} {im.size[0]}x{im.size[1]}")
            except Exception as exc:  # malformed image is a gate failure
                bad.append(f"{index}:{exc}")
        extra = sorted(
            p.name for p in directory.glob("[0-9][0-9][0-9].png")
            if int(p.stem) >= spec["count"]
        ) if directory.is_dir() else []
        state = "PASS" if not missing and not bad and not extra else "FAIL"
        print(f"tileset {tid}: {state} count={spec['count']} missing={missing} bad={bad} extra={extra}")
        failed |= state == "FAIL"
    if failed:
        raise SystemExit(1)


if __name__ == "__main__":
    main()
