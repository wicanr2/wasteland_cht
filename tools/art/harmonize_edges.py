#!/usr/bin/env python3
"""Make AI tile edges deterministic using original edge equivalence classes."""

import argparse
import hashlib
from collections import defaultdict
from pathlib import Path

from PIL import Image


def signature(im: Image.Image, side: str) -> bytes:
    pix = im.load()
    w, h = im.size
    if side == "left":
        seq = [pix[0, y] for y in range(h)]
    elif side == "right":
        seq = [pix[w - 1, y] for y in range(h)]
    elif side == "top":
        seq = [pix[x, 0] for x in range(w)]
    else:
        seq = [pix[x, h - 1] for x in range(w)]
    return bytes(channel for rgba in seq for channel in rgba)


def strip(im: Image.Image, side: str, band: int) -> Image.Image:
    w, h = im.size
    boxes = {
        "left": (0, 0, band, h), "right": (w - band, 0, w, h),
        "top": (0, 0, w, band), "bottom": (0, h - band, w, h),
    }
    return im.crop(boxes[side])


def average(images: list[Image.Image]) -> Image.Image:
    base = images[0]
    out = Image.new("RGBA", base.size)
    src = [im.load() for im in images]
    dst = out.load()
    for y in range(base.height):
        for x in range(base.width):
            dst[x, y] = tuple(sum(p[x, y][c] for p in src) // len(src) for c in range(4))
    return out


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--original", type=Path, required=True)
    ap.add_argument("--input", type=Path, required=True)
    ap.add_argument("--output", type=Path, required=True)
    ap.add_argument("--count", type=int, required=True)
    ap.add_argument("--band", type=int, default=3)
    args = ap.parse_args()

    originals = [Image.open(args.original / f"{i:03d}.png").convert("RGBA") for i in range(args.count)]
    candidates = [Image.open(args.input / f"{i:03d}.png").convert("RGBA") for i in range(args.count)]
    groups: dict[tuple[str, bytes], list[tuple[int, str]]] = defaultdict(list)
    for i, im in enumerate(originals):
        # Horizontal neighbours compare left/right; vertical compare top/bottom.
        for side in ("left", "right"):
            groups[("h", signature(im, side))].append((i, side))
        for side in ("top", "bottom"):
            groups[("v", signature(im, side))].append((i, side))

    replacements: dict[tuple[int, str], Image.Image] = {}
    for members in groups.values():
        if len(members) < 2:
            continue
        avg = average([strip(candidates[i], side, args.band) for i, side in members])
        for i, side in members:
            replacements[(i, side)] = avg

    args.output.mkdir(parents=True, exist_ok=True)
    for i, im in enumerate(candidates):
        w, h = im.size
        for side in ("left", "right", "top", "bottom"):
            rep = replacements.get((i, side))
            if rep is None:
                continue
            at = {
                "left": (0, 0), "right": (w - args.band, 0),
                "top": (0, 0), "bottom": (0, h - args.band),
            }[side]
            im.paste(rep, at)
        im.save(args.output / f"{i:03d}.png", optimize=True)

    digest = hashlib.sha256()
    for i in range(args.count):
        digest.update((args.output / f"{i:03d}.png").read_bytes())
    print(f"harmonized={args.count} groups={sum(len(v) >= 2 for v in groups.values())} sha256={digest.hexdigest()}")


if __name__ == "__main__":
    main()
