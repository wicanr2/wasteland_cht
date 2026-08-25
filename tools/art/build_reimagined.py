#!/usr/bin/env python3
"""Build the complete responsive reimagined artpack from accepted HD masters.

This is intentionally deterministic: every shipped PNG can be rebuilt from the
versioned faithful-hd bundle, and every output is recorded in manifest.json.
The transformation is a distinct 3/4 presentation (64x48 terrain footprints,
wide scene plates, and per-party-member layered animation), not a runtime scale.
"""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path

from PIL import Image, ImageChops, ImageDraw, ImageEnhance, ImageFilter, ImageOps


CHAR_TINTS = (
    (214, 154, 92), (104, 165, 196), (182, 103, 93), (116, 176, 111),
    (190, 150, 204), (204, 190, 103), (112, 132, 172),
)
STATES = {"idle": 1, "walk": 4, "attack": 3, "hurt": 2, "down": 1, "interact": 2}
DIRECTIONS = ("n", "ne", "e", "se", "s", "sw", "w", "nw")
WEAPONS = ("rifle", "pistol", "blade", "launcher")


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def save_asset(root: Path, rel: str, image: Image.Image, asset_id: str, kind: str, assets: list[dict]) -> None:
    path = root / rel
    path.parent.mkdir(parents=True, exist_ok=True)
    image.save(path, format="PNG", optimize=True)
    assets.append({
        "id": asset_id, "kind": kind, "path": rel,
        "width": image.width, "height": image.height, "sha256": sha(path),
    })


def grade(im: Image.Image, saturation: float = 1.10, contrast: float = 1.08) -> Image.Image:
    out = ImageEnhance.Color(im.convert("RGBA")).enhance(saturation)
    out = ImageEnhance.Contrast(out).enhance(contrast)
    # Warm highlights and cooler shadows keep desert/ruin materials readable.
    warm = Image.new("RGBA", out.size, (42, 20, 4, 0))
    warm.putalpha(out.getchannel("A").point(lambda a: a // 12))
    return Image.alpha_composite(out, warm)


def terrain(im: Image.Image) -> Image.Image:
    src = grade(im)
    src = src.resize((64, 56), Image.Resampling.LANCZOS)
    # A shallow 3/4 footprint: preserve silhouettes while exposing a lit front lip.
    out = Image.new("RGBA", (64, 48))
    out.alpha_composite(src.crop((0, 0, 64, 48)))
    lip = src.crop((0, 40, 64, 48)).filter(ImageFilter.GaussianBlur(0.35))
    lip = ImageEnhance.Brightness(lip).enhance(0.72)
    out.alpha_composite(lip, (0, 40))
    return out


def wide_plate(im: Image.Image, size: tuple[int, int]) -> Image.Image:
    src = grade(im, 1.06, 1.05).convert("RGB")
    bg = ImageOps.fit(src, size, method=Image.Resampling.LANCZOS).filter(ImageFilter.GaussianBlur(18))
    bg = ImageEnhance.Brightness(bg).enhance(0.52).convert("RGBA")
    fg = ImageOps.contain(src, (int(size[0] * .82), int(size[1] * .92)), Image.Resampling.LANCZOS).convert("RGBA")
    x, y = (size[0] - fg.width) // 2, (size[1] - fg.height) // 2
    shadow = Image.new("RGBA", size)
    sd = ImageDraw.Draw(shadow)
    sd.rounded_rectangle((x - 8, y - 8, x + fg.width + 8, y + fg.height + 8), 12, fill=(0, 0, 0, 150))
    bg = Image.alpha_composite(bg, shadow.filter(ImageFilter.GaussianBlur(8)))
    bg.alpha_composite(fg, (x, y))
    return bg


def tint_sprite(src: Image.Image, tint: tuple[int, int, int]) -> Image.Image:
    src = src.convert("RGBA").resize((52, 52), Image.Resampling.LANCZOS)
    gray = ImageOps.grayscale(src)
    colored = ImageOps.colorize(gray, (24, 25, 23), tint).convert("RGBA")
    colored.putalpha(src.getchannel("A"))
    out = Image.new("RGBA", (64, 64))
    out.alpha_composite(colored, (6, 8))
    return out


def pose(base: Image.Image, state: str, frame: int, direction: int) -> Image.Image:
    im = base
    if direction in (2, 3, 4):
        im = ImageOps.mirror(im)
    if direction in (1, 3, 5, 7):
        # Diagonal silhouettes are narrower and offset toward their facing side.
        narrow = im.resize((56, 64), Image.Resampling.LANCZOS)
        tmp = Image.new("RGBA", (64, 64)); tmp.alpha_composite(narrow, (4, 0)); im = tmp
    if state == "walk":
        im = ImageChops.offset(im, 0, (0, -2, 0, 2)[frame % 4])
    elif state == "attack":
        im = im.rotate((0, -7, 5)[frame % 3], resample=Image.Resampling.BICUBIC, center=(32, 44))
    elif state == "hurt":
        im = im.rotate((-8, 8)[frame % 2], resample=Image.Resampling.BICUBIC, center=(32, 48))
        red = Image.new("RGBA", im.size, (180, 35, 25, 55)); im = Image.alpha_composite(im, red)
    elif state == "down":
        im = im.rotate(82, resample=Image.Resampling.BICUBIC, expand=False, center=(32, 46))
        im = ImageChops.offset(im, 0, 12)
    elif state == "interact":
        im = im.rotate((-3, 3)[frame % 2], resample=Image.Resampling.BICUBIC, center=(32, 42))
    return im


def weapon(kind: str, direction: int) -> Image.Image:
    im = Image.new("RGBA", (64, 64))
    d = ImageDraw.Draw(im)
    colors = {"rifle": (92, 72, 45, 255), "pistol": (120, 125, 124, 255),
              "blade": (190, 202, 203, 255), "launcher": (69, 86, 56, 255)}
    lengths = {"rifle": 26, "pistol": 14, "blade": 22, "launcher": 29}
    c, length = colors[kind], lengths[kind]
    d.rounded_rectangle((31, 27, 31 + length, 31), 2, fill=c, outline=(25, 23, 20, 255))
    d.rectangle((29, 30, 36, 34), fill=(53, 43, 31, 255))
    # Asset orientation starts east and is rotated in 45-degree increments.
    return im.rotate(direction * 45, resample=Image.Resampling.BICUBIC, center=(32, 32))


def build(source: Path, output: Path) -> None:
    src_manifest = json.loads((source / "manifest.json").read_text(encoding="utf-8"))
    by_id = {a["id"]: a for a in src_manifest["assets"]}
    assets: list[dict] = []

    tile_counts = (66, 141, 163, 107, 127, 118, 90, 104, 135)
    for tileset, count in enumerate(tile_counts):
        for index in range(count):
            src = Image.open(source / by_id[f"tile.{tileset}.{index:03d}"]["path"])
            save_asset(output, f"assets/tileset-{tileset}/{index:03d}.png", terrain(src),
                       f"tile.{tileset}.{index:03d}", "terrain-3q", assets)

    for index in range(10):
        src = Image.open(source / by_id[f"icon.{index:03d}"]["path"])
        icon = grade(src).resize((64, 64), Image.Resampling.LANCZOS)
        save_asset(output, f"assets/icons/{index:03d}.png", icon,
                   f"icon.{index:03d}", "world-object", assets)

    for index in range(82):
        src = Image.open(source / by_id[f"scene.{index:03d}"]["path"])
        save_asset(output, f"assets/scenes/{index:03d}.png", wide_plate(src, (768, 432)),
                   f"scene.{index:03d}", "scene-wide", assets)
    for name in ("title", "ending"):
        src = Image.open(source / by_id[f"fullscreen.{name}"]["path"])
        save_asset(output, f"assets/fullscreen/{name}.png", wide_plate(src, (1280, 720)),
                   f"fullscreen.{name}", "fullscreen-wide", assets)

    cardinal = [Image.open(source / by_id[f"party.walk.{i:03d}"]["path"]) for i in range(12)]
    for character, tint in enumerate(CHAR_TINTS):
        for direction, direction_name in enumerate(DIRECTIONS):
            # Cardinal source selection is deterministic; diagonal identity comes from pose().
            source_dir = ((direction + 1) // 2) % 4
            for state, frames in STATES.items():
                for frame in range(frames):
                    base = tint_sprite(cardinal[source_dir * 3 + (frame % 3)], tint)
                    im = pose(base, state, frame, direction)
                    rel = f"assets/characters/{character}/{direction_name}/{state}-{frame:02d}.png"
                    save_asset(output, rel, im,
                               f"character.{character}.{direction_name}.{state}.{frame:02d}",
                               "character-layer", assets)

    for kind in WEAPONS:
        for direction, direction_name in enumerate(DIRECTIONS):
            save_asset(output, f"assets/weapons/{kind}-{direction_name}.png", weapon(kind, direction),
                       f"weapon.{kind}.{direction_name}", "weapon-layer", assets)

    manifest = {
        "schema": 1, "id": "wasteland-reimagined-v1", "mode": "reimagined",
        "canvas": {"width": 1280, "height": 720, "responsive": True,
                   "max_view_cols": 25, "max_view_rows": 15},
        "assets": assets,
    }
    output.mkdir(parents=True, exist_ok=True)
    (output / "manifest.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(f"built {len(assets)} assets in {output}")


def main() -> None:
    p = argparse.ArgumentParser()
    p.add_argument("--source", type=Path, default=Path("artpacks/faithful-hd"))
    p.add_argument("--output", type=Path, default=Path("artpacks/reimagined"))
    args = p.parse_args()
    build(args.source, args.output)


if __name__ == "__main__":
    main()
