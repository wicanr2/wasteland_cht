#!/usr/bin/env python3
"""Deterministically crop and resize one faithful-HD scene to 288x252."""

import argparse
from pathlib import Path

from PIL import Image


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("input", type=Path)
    ap.add_argument("output", type=Path)
    ap.add_argument("--width", type=int, default=288)
    ap.add_argument("--height", type=int, default=252)
    args = ap.parse_args()

    im = Image.open(args.input).convert("RGB")
    target_ratio = args.width / args.height
    source_ratio = im.width / im.height
    if source_ratio > target_ratio:
        crop_width = round(im.height * target_ratio)
        left = (im.width - crop_width) // 2
        im = im.crop((left, 0, left + crop_width, im.height))
    elif source_ratio < target_ratio:
        crop_height = round(im.width / target_ratio)
        top = (im.height - crop_height) // 2
        im = im.crop((0, top, im.width, top + crop_height))

    im = im.resize((args.width, args.height), Image.Resampling.LANCZOS)
    args.output.parent.mkdir(parents=True, exist_ok=True)
    im.save(args.output, optimize=True)


if __name__ == "__main__":
    main()
