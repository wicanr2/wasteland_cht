#!/usr/bin/env python3
"""Assemble indexed tile PNG files into deterministic batch atlases."""

import argparse
from pathlib import Path

from PIL import Image


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("input", type=Path)
    ap.add_argument("output", type=Path)
    ap.add_argument("--count", type=int, required=True)
    ap.add_argument("--cols", type=int, default=4)
    ap.add_argument("--batch", type=int, default=16)
    args = ap.parse_args()
    args.output.mkdir(parents=True, exist_ok=True)
    for start in range(0, args.count, args.batch):
        end = min(start + args.batch, args.count)
        opened = [Image.open(args.input / f"{i:03d}.png").convert("RGBA") for i in range(start, end)]
        cell = opened[0].size
        if any(im.size != cell for im in opened):
            raise SystemExit(f"batch {start}-{end - 1} 尺寸不一致")
        rows = (len(opened) + args.cols - 1) // args.cols
        atlas = Image.new("RGBA", (args.cols * cell[0], rows * cell[1]), (0, 0, 0, 0))
        for local, im in enumerate(opened):
            atlas.alpha_composite(im, ((local % args.cols) * cell[0], (local // args.cols) * cell[1]))
        atlas.save(args.output / f"atlas-{start:03d}-{end - 1:03d}.png", optimize=True)


if __name__ == "__main__":
    main()
