#!/usr/bin/env python3
"""Remove a light checkerboard connected to the image boundary."""

import argparse
from collections import deque
from pathlib import Path

from PIL import Image


def background(px: tuple[int, int, int, int], minimum: int, chroma: int) -> bool:
    r, g, b, _ = px
    return min(r, g, b) >= minimum and max(r, g, b) - min(r, g, b) <= chroma


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("input", type=Path)
    ap.add_argument("output", type=Path)
    ap.add_argument("--minimum", type=int, default=220)
    ap.add_argument("--chroma", type=int, default=12)
    args = ap.parse_args()
    im = Image.open(args.input).convert("RGBA")
    pix = im.load()
    seen: set[tuple[int, int]] = set()
    q: deque[tuple[int, int]] = deque()
    for x in range(im.width):
        q.extend(((x, 0), (x, im.height - 1)))
    for y in range(im.height):
        q.extend(((0, y), (im.width - 1, y)))
    while q:
        x, y = q.popleft()
        if (x, y) in seen or not background(pix[x, y], args.minimum, args.chroma):
            continue
        seen.add((x, y))
        pix[x, y] = (0, 0, 0, 0)
        if x: q.append((x - 1, y))
        if x + 1 < im.width: q.append((x + 1, y))
        if y: q.append((x, y - 1))
        if y + 1 < im.height: q.append((x, y + 1))
    args.output.parent.mkdir(parents=True, exist_ok=True)
    im.save(args.output, optimize=True)
    print(f"transparent_pixels={len(seen)} total={im.width * im.height}")


if __name__ == "__main__":
    main()
