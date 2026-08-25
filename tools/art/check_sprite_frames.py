#!/usr/bin/env python3
"""Validate indexed animation frames for size, alpha content and safety border."""

import argparse
from pathlib import Path

from PIL import Image, ImageChops


def main() -> None:
    ap = argparse.ArgumentParser()
    ap.add_argument("--root", type=Path, required=True)
    ap.add_argument("--count", type=int, required=True)
    ap.add_argument("--size", type=int, default=48)
    ap.add_argument("--groups", type=int, default=1, help="方向群組數；每群影格應有變化")
    args = ap.parse_args()
    failed = []
    frames = []
    for index in range(args.count):
        p = args.root / f"{index:03d}.png"
        if not p.is_file():
            failed.append(f"{index}:missing")
            continue
        im = Image.open(p).convert("RGBA")
        frames.append(im)
        if im.size != (args.size, args.size): failed.append(f"{index}:size={im.size}")
        alpha = im.getchannel("A")
        if alpha.getextrema()[1] == 0: failed.append(f"{index}:empty")
        border = alpha.crop((0, 0, args.size, 1)).getbbox() or alpha.crop((0, args.size-1, args.size, args.size)).getbbox()
        border = border or alpha.crop((0, 0, 1, args.size)).getbbox() or alpha.crop((args.size-1, 0, args.size, args.size)).getbbox()
        if border: failed.append(f"{index}:touches-border")
    if args.count % args.groups:
        failed.append("count-not-divisible-by-groups")
    elif len(frames) == args.count:
        per = args.count // args.groups
        for group in range(args.groups):
            seq = frames[group*per:(group+1)*per]
            if all(ImageChops.difference(seq[0], other).getbbox() is None for other in seq[1:]):
                failed.append(f"group-{group}:no-animation")
    if failed:
        raise SystemExit("sprite frames: FAIL " + " ".join(failed))
    print(f"sprite frames: PASS count={args.count} size={args.size}x{args.size} groups={args.groups}")


if __name__ == "__main__":
    main()
