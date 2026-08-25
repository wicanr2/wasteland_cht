#!/usr/bin/env python3
"""Validate faithful-hd map overlay icons and transparent safety margins."""

import argparse
from pathlib import Path

from PIL import Image


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--root", type=Path, default=Path("artpacks/faithful-hd/assets/icons"))
    args = ap.parse_args()
    failed = []
    for index in range(10):
        p = args.root / f"{index:03d}.png"
        if not p.is_file():
            failed.append(f"{index}:missing")
            continue
        with Image.open(p) as opened:
            im = opened.convert("RGBA")
        if im.size != (48, 48):
            failed.append(f"{index}:size={im.size}")
            continue
        alpha = im.getchannel("A")
        extrema = alpha.getextrema()
        if index == 0:
            if extrema != (0, 0):
                failed.append(f"0:not-empty alpha={extrema}")
            continue
        if extrema[1] == 0:
            failed.append(f"{index}:empty")
        border = list(alpha.crop((0, 0, 48, 1)).getdata())
        border += list(alpha.crop((0, 47, 48, 48)).getdata())
        border += list(alpha.crop((0, 0, 1, 48)).getdata())
        border += list(alpha.crop((47, 0, 48, 48)).getdata())
        if any(border):
            failed.append(f"{index}:touches-border")
    if failed:
        raise SystemExit("icons: FAIL " + " ".join(failed))
    print("icons: PASS count=10 size=48x48 icon0=transparent border=transparent")


if __name__ == "__main__":
    main()
