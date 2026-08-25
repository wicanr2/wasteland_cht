#!/usr/bin/env python3
"""Deterministically crop and resize one 9:4 faithful-HD full-width picture."""

import argparse
from pathlib import Path

from PIL import Image


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("input", type=Path)
    ap.add_argument("output", type=Path)
    ap.add_argument("--width", type=int, default=864)
    ap.add_argument("--height", type=int, default=384)
    args = ap.parse_args()

    im = Image.open(args.input).convert("RGB")
    target = args.width / args.height
    source = im.width / im.height
    if source > target:
        width = round(im.height * target)
        left = (im.width - width) // 2
        im = im.crop((left, 0, left + width, im.height))
    elif source < target:
        height = round(im.width / target)
        top = (im.height - height) // 2
        im = im.crop((0, top, im.width, top + height))
    im = im.resize((args.width, args.height), Image.Resampling.LANCZOS)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    im.save(args.output, optimize=True)


if __name__ == "__main__":
    main()
