#!/usr/bin/env python3
"""Fit alpha sprites inside a deterministic transparent safety rectangle."""

import argparse
from pathlib import Path

from PIL import Image


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("input", type=Path)
    ap.add_argument("output", type=Path)
    ap.add_argument("--count", type=int, required=True)
    ap.add_argument("--canvas", type=int, default=48)
    ap.add_argument("--max", dest="maximum", type=int, default=40)
    args = ap.parse_args()
    args.output.mkdir(parents=True, exist_ok=True)
    for index in range(args.count):
        im = Image.open(args.input / f"{index:03d}.png").convert("RGBA")
        bbox = im.getchannel("A").getbbox()
        dst = Image.new("RGBA", (args.canvas, args.canvas), (0, 0, 0, 0))
        if bbox:
            sprite = im.crop(bbox)
            ratio = min(args.maximum / sprite.width, args.maximum / sprite.height, 1.0)
            size = (max(1, round(sprite.width * ratio)), max(1, round(sprite.height * ratio)))
            if size != sprite.size:
                sprite = sprite.resize(size, Image.Resampling.LANCZOS)
            dst.alpha_composite(sprite, ((args.canvas - size[0]) // 2, (args.canvas - size[1]) // 2))
        dst.save(args.output / f"{index:03d}.png", optimize=True)


if __name__ == "__main__":
    main()
