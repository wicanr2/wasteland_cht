#!/usr/bin/env python3
"""Validate a complete numbered faithful-HD scene directory."""

import argparse
from pathlib import Path

from PIL import Image


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("directory", type=Path)
    ap.add_argument("--count", type=int, default=82)
    ap.add_argument("--width", type=int, default=288)
    ap.add_argument("--height", type=int, default=252)
    args = ap.parse_args()

    expected = {f"{i:03d}.png" for i in range(args.count)}
    actual = {p.name for p in args.directory.glob("*.png")}
    missing = sorted(expected - actual)
    extra = sorted(actual - expected)
    errors = []
    for name in sorted(expected & actual):
        with Image.open(args.directory / name) as im:
            if im.format != "PNG":
                errors.append(f"{name}: format={im.format}")
            if im.size != (args.width, args.height):
                errors.append(f"{name}: size={im.size}")
            lo, hi = im.convert("L").getextrema()
            if lo == hi:
                errors.append(f"{name}: blank image")
    if missing or extra or errors:
        for item in missing:
            print(f"MISSING {item}")
        for item in extra:
            print(f"EXTRA {item}")
        for item in errors:
            print(f"ERROR {item}")
        raise SystemExit(1)
    print(f"PASS scenes={args.count} size={args.width}x{args.height}")


if __name__ == "__main__":
    main()
