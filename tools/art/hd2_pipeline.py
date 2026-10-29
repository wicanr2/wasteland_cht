#!/usr/bin/env python3
"""Prepare reviewed HD2 tile batches; incomplete work never becomes a player bundle."""

import argparse
import hashlib
import json
from io import BytesIO
from pathlib import Path

from PIL import Image, ImageChops


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def local_path(value):
    p = Path(value)
    if p.is_absolute() or ".." in p.parts:
        raise ValueError(f"path must stay inside the project: {value}")
    return p


def tile_indices(batch, limit):
    indices = batch.get("indices", list(range(batch["start"], batch["start"] + batch["count"])))
    skipped = batch.get("superseded_indices", [])
    if (not indices or len(indices) != batch["count"] or len(set(indices)) != len(indices)
            or any(type(i) is not int or i < 0 or i >= limit for i in indices)
            or indices != sorted(indices) or batch["start"] != indices[0]):
        raise ValueError(f"invalid original tile index mapping: {batch['id']}")
    if len(set(skipped)) != len(skipped) or not set(skipped) <= set(indices):
        raise ValueError(f"invalid superseded indices: {batch['id']}")
    return indices, set(skipped)


def gutter_runs(im, axis):
    # Technical magenta frames are reference guides, never game pixels.
    rgb = im.convert("RGB")
    n, m = (rgb.width, rgb.height) if axis == "x" else (rgb.height, rgb.width)
    r, g, b = rgb.split()
    mask = ImageChops.multiply(r.point(lambda p: 255 if p > 190 else 0),
                               b.point(lambda p: 255 if p > 190 else 0))
    mask = ImageChops.multiply(mask, g.point(lambda p: 255 if p < 85 else 0))
    found = []
    start = None
    for i in range(n):
        strip = mask.crop((i, 0, i + 1, m) if axis == "x" else (0, i, m, i + 1))
        count = strip.histogram()[255]
        on = count / m > 0.65
        if on and start is None:
            start = i
        if start is not None and (not on or i == n - 1):
            end = i if on else i - 1
            if end - start >= 2:
                found.append((start, end))
            start = None
    return found


def crop_boxes(im, batch):
    cols, rows = batch["cols"], batch["rows"]
    if batch.get("framed"):
        gx, gy = gutter_runs(im, "x"), gutter_runs(im, "y")
        if len(gx) != cols + 1 or len(gy) != rows + 1:
            raise ValueError(f"magenta frame count differs: {len(gx)}/{len(gy)} for {batch['id']}")
        xs = [(gx[i][1] + 1, gx[i + 1][0]) for i in range(cols)]
        ys = [(gy[i][1] + 1, gy[i + 1][0]) for i in range(rows)]
    else:
        xc = batch.get("x_cuts", [i / cols for i in range(cols + 1)])
        yc = batch.get("y_cuts", [i / rows for i in range(rows + 1)])
        if len(xc) != cols + 1 or len(yc) != rows + 1 or xc[0] != 0 or yc[0] != 0 or xc[-1] != 1 or yc[-1] != 1:
            raise ValueError(f"bad measured cut grid: {batch['id']}")
        xs = [(round(xc[i] * im.width), round(xc[i + 1] * im.width)) for i in range(cols)]
        ys = [(round(yc[i] * im.height), round(yc[i + 1] * im.height)) for i in range(rows)]
    boxes = [(xs[n % cols][0], ys[n // cols][0], xs[n % cols][1], ys[n // cols][1])
             for n in range(batch["count"])]
    if any(x1 <= x0 or y1 <= y0 for x0, y0, x1, y1 in boxes):
        raise ValueError(f"empty tile rectangle: {batch['id']}")
    return boxes


def clean_frame_edge(tile, original):
    """Extend generated paint over a thin technical-frame fringe, in place.

    Original pixels are consulted only to protect genuine purple motifs; no
    original color is copied into the output. Interior artwork is untouched.
    """
    def magenta(pixel):
        r, g, b = pixel[:3]
        # Use the same bright technical color as gutter_runs. Muted violet
        # mountain paint is genuine artwork even at the boundary of a tile.
        return r > 190 and b > 190 and g < 85

    def original_purple(pixel):
        r, g, b = pixel[:3]
        return min(r, b) > 100 and g < min(r, b) * 0.65

    out = tile.copy()
    width, height = tile.size
    band = max(2, round(min(width, height) / 100))
    source = original.convert("RGB")
    changed = 0
    for y in range(height):
        edge_x = (range(width) if y < band or y >= height - band else
                  (*range(band), *range(width - band, width)))
        for x in edge_x:
            if not magenta(tile.getpixel((x, y))):
                continue
            ox, oy = x * source.width // width, y * source.height // height
            if original_purple(source.getpixel((ox, oy))):
                continue
            # Same material immediately inside this generated tile, at the
            # same coordinates. Do not resize or move the material footprint.
            ix = min(max(x, band), width - band - 1)
            iy = min(max(y, band), height - band - 1)
            paint = tile.getpixel((ix, iy))
            if magenta(paint):
                # A softened frame can tint the first sample too. Search only
                # the immediately adjacent generated paint, within this band.
                candidates = [(abs(dx) + abs(dy), ix + dx, iy + dy)
                              for dy in range(-band, band + 1)
                              for dx in range(-band, band + 1)
                              if band <= ix + dx < width - band
                              and band <= iy + dy < height - band
                              and not magenta(tile.getpixel((ix + dx, iy + dy)))]
                if not candidates:
                    raise ValueError(f"no uncontaminated frame-edge paint near {x},{y}")
                _, ix, iy = min(candidates)
                paint = tile.getpixel((ix, iy))
            out.putpixel((x, y), paint)
            changed += 1
    return out, changed


def make_guides(tid):
    inv = json.loads(Path("artwork/manifests/original-art-inventory.json").read_text())
    count = next(t["count"] for t in inv["tilesets"] if t["id"] == tid)
    out = Path("workplace/art-reference/hd2-guides")
    out.mkdir(parents=True, exist_ok=True)
    for start in range(0, count, 16):
        end = min(start + 16, count)
        rows = (end - start + 3) // 4
        cell, gap = 128, 16
        sheet = Image.new("RGB", (4 * cell + 5 * gap, rows * cell + (rows + 1) * gap), (255, 0, 255))
        for n in range(rows * 4):
            tile = Image.new("RGB", (cell, cell))
            if start + n < end:
                p = Path(f"workplace/art-reference/tiles/tileset-{tid}/{start+n:03d}.png")
                tile = Image.open(p).convert("RGB").resize((cell, cell), Image.Resampling.NEAREST)
            sheet.paste(tile, (gap + (n % 4) * (cell + gap), gap + (n // 4) * (cell + gap)))
        sheet.save(out / f"tileset-{tid}-{start:03d}-{end-1:03d}.png")
    print(f"reference guides: {out}, tileset={tid}, count={count}")


def prepare(ledger_path, mode):
    ledger = json.loads(ledger_path.read_text(encoding="utf-8"))
    inventory = json.loads(Path("artwork/manifests/original-art-inventory.json").read_text())
    counts = {t["id"]: t["count"] for t in inventory["tilesets"]}
    out = Path("artwork/work") / mode / "assets"
    seen = set()
    receipts = []
    pending = []
    for batch in ledger["batches"]:
        if batch["mode"] != mode or batch["status"] != "reviewed":
            continue
        tid, start, count = batch["tileset"], batch["start"], batch["count"]
        cols, rows = batch["cols"], batch["rows"]
        if tid not in counts or start < 0 or count < 1:
            raise ValueError(f"batch indices exceed original inventory: {batch['id']}")
        indices, skipped = tile_indices(batch, counts[tid])
        if cols != 4 or rows < 1 or rows > 4 or count > cols * rows:
            raise ValueError(f"unreviewable batch grid: {batch['id']}")
        source = local_path(batch["source"])
        if sha(source) != batch["source_sha256"]:
            raise ValueError(f"source changed after review: {source}")
        reference = local_path(batch["reference"])
        if sha(reference) != batch["reference_sha256"]:
            raise ValueError(f"reference changed after review: {reference}")
        im = Image.open(source).convert("RGBA")
        # A is authored pixel art; B is painted. This only resamples and cuts the
        # generated source, without inventing surfaces or replacing original pixels.
        resample = Image.Resampling.NEAREST if mode == "hd2-a" else Image.Resampling.LANCZOS
        tiles = out / f"tileset-{tid}"
        atlas = Image.new("RGBA", (cols * 48, rows * 48))
        files = []
        boxes = crop_boxes(im, batch)
        for n in range(count):
            index = indices[n]
            if index in skipped:
                continue
            if (tid, index) in seen:
                raise ValueError(f"duplicate tile {tid}/{index}")
            seen.add((tid, index))
            col, row = n % cols, n // cols
            box = boxes[n]
            tile = im.crop(box)
            fringe_pixels = 0
            original = Path(f"workplace/art-reference/tiles/tileset-{tid}/{index:03d}.png")
            if batch.get("framed"):
                tile, fringe_pixels = clean_frame_edge(tile, Image.open(original))
            tile = tile.resize((48, 48), resample)
            if tile.getextrema()[3] != (255, 255):
                raise ValueError(f"terrain tile unexpectedly transparent: {tid}/{index}")
            path = tiles / f"{index:03d}.png"
            encoded = BytesIO()
            tile.save(encoded, format="PNG", optimize=True)
            data = encoded.getvalue()
            pending.append((path, data))
            atlas.paste(tile, (col * 48, row * 48))
            files.append({"index": index, "source_box": box, "path": str(path),
                          "sha256": hashlib.sha256(data).hexdigest(),
                          "frame_fringe_pixels": fringe_pixels,
                          "footprint_reference_sha256": sha(original)})
        atlas_path = tiles / (f"atlas-{batch['id']}.png" if "indices" in batch else
                              f"atlas-{start:03d}-{start + count - 1:03d}.png")
        encoded = BytesIO()
        atlas.save(encoded, format="PNG", optimize=True)
        data = encoded.getvalue()
        pending.append((atlas_path, data))
        receipts.append({"batch": batch["id"], "source_sha256": sha(source),
                         "resampling": "nearest" if mode == "hd2-a" else "lanczos",
                         "files": files, "atlas": str(atlas_path),
                         "atlas_sha256": hashlib.sha256(data).hexdigest()})
    # Validate every reviewed batch before replacing any existing WIP output.
    for path, data in pending:
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_bytes(data)
    # This named WIP directory contains only reproducible pipeline output.
    # Remove stale slots excluded from the reviewed ledger after a rejection.
    for path in out.glob("tileset-*/[0-9][0-9][0-9].png"):
        tid = int(path.parent.name.split("-")[-1])
        if (tid, int(path.stem)) not in seen:
            path.unlink()
    coverage = [{"tileset": tid, "ready": sum((tid, i) in seen for i in range(count)),
                 "required": count} for tid, count in sorted(counts.items())]
    receipt = {"schema": 1, "mode": mode, "status": "wip-not-player-bundle",
               "pipeline": "tools/art/hd2_pipeline.py", "ledger_sha256": sha(ledger_path),
               "pipeline_sha256": sha(Path(__file__)),
               "coverage": coverage, "batches": receipts}
    path = Path("artwork/work") / mode / "receipt.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"mode": mode, "coverage": coverage, "receipt": str(path)}))


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--ledger", type=Path, default=Path("artwork/manifests/hd2-production.json"))
    ap.add_argument("--mode", choices=["hd2-a", "hd2-b"])
    ap.add_argument("--guides", action="store_true")
    ap.add_argument("--tileset", type=int, default=3)
    args = ap.parse_args()
    if args.guides:
        make_guides(args.tileset)
    elif args.mode:
        prepare(args.ledger, args.mode)
    else:
        ap.error("--mode or --guides is required")


if __name__ == "__main__":
    main()
