#!/usr/bin/env python3
"""Split an AI grid into deterministic indexed PNG cells and a clean atlas."""

import argparse
from pathlib import Path

from PIL import Image


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("input", type=Path)
    ap.add_argument("output", type=Path)
    ap.add_argument("--cols", type=int, required=True)
    ap.add_argument("--rows", type=int, required=True)
    ap.add_argument("--cell", type=int, required=True)
    ap.add_argument("--start", type=int, default=0)
    ap.add_argument("--count", type=int, default=0)
    ap.add_argument("--inset", type=int, default=0)
    args = ap.parse_args()

    if args.cols < 1 or args.rows < 1 or args.cell < 1 or args.inset < 0:
        ap.error("cols、rows、cell 必須大於零；inset 不得為負")
    total = args.cols * args.rows
    count = args.count or total
    if count < 1 or count > total:
        ap.error(f"count 必須在 1–{total}")

    src = Image.open(args.input).convert("RGBA")
    args.output.mkdir(parents=True, exist_ok=True)
    atlas = Image.new("RGBA", (args.cols * args.cell, args.rows * args.cell), (0, 0, 0, 0))

    for local in range(count):
        col, row = local % args.cols, local // args.cols
        # round keeps all source pixels accounted for when dimensions are not
        # divisible by the requested grid (a common image-generator output).
        x0 = round(col * src.width / args.cols) + args.inset
        x1 = round((col + 1) * src.width / args.cols) - args.inset
        y0 = round(row * src.height / args.rows) + args.inset
        y1 = round((row + 1) * src.height / args.rows) - args.inset
        if x1 <= x0 or y1 <= y0:
            raise SystemExit(f"cell {local} inset 後為空")
        tile = src.crop((x0, y0, x1, y1)).resize(
            (args.cell, args.cell), Image.Resampling.LANCZOS
        )
        index = args.start + local
        tile.save(args.output / f"{index:03d}.png", optimize=True)
        atlas.alpha_composite(tile, (col * args.cell, row * args.cell))

    atlas.save(args.output / f"atlas-{args.start:03d}-{args.start + count - 1:03d}.png", optimize=True)


if __name__ == "__main__":
    main()
