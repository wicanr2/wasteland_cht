#!/usr/bin/env python3
"""Prepare reviewed HD2 tile batches; incomplete work never becomes a player bundle."""

import argparse
import hashlib
import json
import shutil
import tempfile
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
    if "source_boxes" in batch:
        boxes = batch["source_boxes"]
        if (batch.get("framed") or len(boxes) != batch["count"] or
                any(len(b) != 4 or any(type(v) is not int for v in b) or
                    not (0 <= b[0] < b[2] <= im.width and 0 <= b[1] < b[3] <= im.height)
                    for b in boxes)):
            raise ValueError(f"invalid measured source rectangles: {batch['id']}")
        return [tuple(b) for b in boxes]
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


EXTRA_GROUPS = {
    "scene": (82, "scenes", (288, 252), "scene"),
    "icons": (10, "icons", (48, 48), "icon"),
    "party-walk": (12, "party-walk", (48, 48), "party-sprite"),
    "title": (1, "fullscreen", (864, 384), "title"),
    "ending": (1, "fullscreen", (864, 384), "ending"),
}

# Generated cutout mattes contain isolated alpha 1–3 background specks. Use a
# visible bounding box for cropping; keep all original alpha inside the crop.
MATTE_VISIBLE_ALPHA = 4


def visible_bbox(alpha):
    return alpha.point(lambda p: 255 if p >= MATTE_VISIBLE_ALPHA else 0).getbbox()


def extra_path(kind, index):
    _, directory, _, _ = EXTRA_GROUPS[kind]
    return Path(directory) / (f"{kind}.png" if kind in ("title", "ending") else f"{index:03d}.png")


def prepare_extras(ledger, mode, out):
    pending, receipts, seen = [], [], set()
    for batch in ledger.get("illustration_batches", []):
        if batch["mode"] != mode or batch["status"] != "reviewed":
            continue
        kind = batch["kind"]
        if kind not in EXTRA_GROUPS:
            raise ValueError(f"unknown illustration group: {kind}")
        limit, _, size, _ = EXTRA_GROUPS[kind]
        indices, skipped = tile_indices(batch, limit)
        if (skipped and kind != "scene") or not 1 <= batch["cols"] <= 5 or not 1 <= batch["rows"] <= 4 or len(indices) > batch["cols"] * batch["rows"]:
            raise ValueError(f"invalid illustration grid: {batch['id']}")
        source, reference = local_path(batch["source"]), local_path(batch["reference"])
        if sha(source) != batch["source_sha256"] or sha(reference) != batch["reference_sha256"]:
            raise ValueError(f"illustration source changed after review: {batch['id']}")
        im = Image.open(source).convert("RGBA")
        boxes = crop_boxes(im, batch)
        files = []
        resample = Image.Resampling.NEAREST if mode == "hd2-a" else Image.Resampling.LANCZOS
        party_bounds = None
        if kind == "party-walk":
            # A shared crop and scale preserves body size and the foot anchor
            # when one animation frame has a raised leg.
            bounds = [visible_bbox(im.crop(b).resize((256, 256), resample).getchannel("A")) for b in boxes]
            if any(b is None for b in bounds):
                raise ValueError("party walk has an empty frame")
            party_bounds = (min(b[0] for b in bounds), min(b[1] for b in bounds),
                            max(b[2] for b in bounds), max(b[3] for b in bounds))
        for index, box in zip(indices, boxes):
            if index in skipped:
                continue
            if (kind, index) in seen:
                raise ValueError(f"duplicate illustration: {kind}/{index}")
            seen.add((kind, index))
            tile = im.crop(box)
            sprite = kind in ("icons", "party-walk")
            if sprite:
                # Only technical separator fringe is removed. Sprite colors,
                # including actual pink figures, are not chroma-keyed.
                band = max(2, round(min(tile.size) / 100))
                for y in range(tile.height):
                    for x in (range(tile.width) if y < band or y >= tile.height - band else
                              (*range(band), *range(tile.width - band, tile.width))):
                        r, g, b, a = tile.getpixel((x, y))
                        if r > 190 and b > 190 and g < 85:
                            tile.putpixel((x, y), (0, 0, 0, 0))
                alpha = tile.getchannel("A")
                bbox = visible_bbox(alpha)
                if kind == "icons" and index == 0:
                    if tile.getextrema() != ((0, 0), (0, 0), (0, 0), (255, 255)):
                        raise ValueError("icon zero must be opaque black, as in the decoded source mask")
                    tile = Image.new("RGBA", size, (0, 0, 0, 255))
                else:
                    if bbox is None or alpha.getextrema()[0] != 0 or bbox[0] < band or bbox[1] < band or bbox[2] > tile.width - band or bbox[3] > tile.height - band:
                        raise ValueError(f"sprite needs a clean transparent margin: {kind}/{index}")
                    figure = (tile.resize((256, 256), resample).crop(party_bounds)
                              if party_bounds else tile.crop(bbox))
                    scale = min(40 / figure.width, 40 / figure.height)
                    fitted = (max(1, round(figure.width * scale)), max(1, round(figure.height * scale)))
                    figure = figure.resize(fitted, resample)
                    tile = Image.new("RGBA", size)
                    tile.alpha_composite(figure, ((size[0] - fitted[0]) // 2, (size[1] - fitted[1]) // 2))
            else:
                if batch.get("framed"):
                    original = Path(f"workplace/art-reference/scenes/{index:03d}.png")
                    tile, _ = clean_frame_edge(tile, Image.open(original))
                tile = tile.resize(size, resample)
                if tile.getextrema()[3] != (255, 255):
                    raise ValueError(f"illustration unexpectedly transparent: {kind}/{index}")
            path = out / extra_path(kind, index)
            encoded = BytesIO()
            tile.save(encoded, format="PNG", optimize=True)
            data = encoded.getvalue()
            pending.append((path, data))
            files.append({"index": index, "source_box": box, "path": str(path),
                          "sha256": hashlib.sha256(data).hexdigest()})
        receipts.append({"batch": batch["id"], "kind": kind,
                         "source_sha256": sha(source), "files": files,
                         "visible_alpha_threshold": MATTE_VISIBLE_ALPHA if kind in ("icons", "party-walk") else None})
    if any(kind == "icons" for kind, _ in seen) and ("icons", 0) not in seen:
        # Original slot zero is a black clearing cell, not artwork. Preserve
        # its decoded mask semantics without asking a generator to invent it.
        empty = ledger.get("empty_icon", {})
        source = local_path(empty.get("reference", ""))
        if not source.is_file() or sha(source) != empty.get("sha256"):
            raise ValueError("missing verified original clearing icon")
        original = Image.open(source).convert("RGBA")
        if original.size != (16, 16) or original.getextrema() != ((0, 0), (0, 0), (0, 0), (255, 255)):
            raise ValueError("original clearing icon is not opaque black")
        tile = original.resize((48, 48), Image.Resampling.NEAREST)
        encoded = BytesIO()
        tile.save(encoded, format="PNG", optimize=True)
        path, data = out / extra_path("icons", 0), encoded.getvalue()
        pending.append((path, data))
        seen.add(("icons", 0))
        receipts.append({"batch": "canonical-black-clearing-cell", "kind": "icons",
                         "source_sha256": sha(source), "files": [{"index": 0, "path": str(path),
                         "sha256": hashlib.sha256(data).hexdigest()}]})
    coverage = [{"kind": kind, "ready": sum((kind, i) in seen for i in range(v[0])),
                 "required": v[0]} for kind, v in EXTRA_GROUPS.items()]
    return pending, receipts, coverage, seen


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
    extra_pending, illustrations, extra_coverage, extra_seen = prepare_extras(ledger, mode, out)
    pending.extend(extra_pending)
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
    for kind, (limit, _, _, _) in EXTRA_GROUPS.items():
        for index in range(limit):
            path = out / extra_path(kind, index)
            if path.exists() and (kind, index) not in extra_seen:
                path.unlink()
    coverage = [{"tileset": tid, "ready": sum((tid, i) in seen for i in range(count)),
                 "required": count} for tid, count in sorted(counts.items())]
    receipt = {"schema": 1, "mode": mode, "status": "wip-not-player-bundle",
               "pipeline": "tools/art/hd2_pipeline.py", "ledger_sha256": sha(ledger_path),
               "pipeline_sha256": sha(Path(__file__)),
               "coverage": coverage, "batches": receipts,
               "extra_coverage": extra_coverage, "illustrations": illustrations}
    path = Path("artwork/work") / mode / "receipt.json"
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(receipt, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
    print(json.dumps({"mode": mode, "coverage": coverage, "receipt": str(path)}))
    return receipt


def build_bundle(receipt, mode):
    """Install a complete local candidate. No original or retired HD asset is used."""
    if mode not in ("hd2-a", "hd2-b") or receipt["mode"] != mode:
        raise ValueError("bundle identity mismatch")
    if any(c["ready"] != c["required"] for c in receipt["coverage"] + receipt["extra_coverage"]):
        raise ValueError("incomplete HD2 coverage; player bundle was not created")
    root = Path("artpacks")
    root.mkdir(exist_ok=True)
    target = root / mode
    if target.exists():
        raise FileExistsError(f"existing local bundle must be retained before replacement: {target}")
    assets, expected = [], {}
    for batch in receipt["batches"] + receipt["illustrations"]:
        for file in batch["files"]:
            expected[file["path"]] = file["sha256"]
    entries = []
    for c in receipt["coverage"]:
        for i in range(c["required"]):
            entries.append((f"tile.{c['tileset']}.{i:03d}", "tile", Path(f"tileset-{c['tileset']}/{i:03d}.png"), (48, 48)))
    for kind, (count, _, size, asset_kind) in EXTRA_GROUPS.items():
        for i in range(count):
            asset_id = (f"fullscreen.{kind}" if kind in ("title", "ending") else
                        f"party.walk.{i:03d}" if kind == "party-walk" else
                        f"icon.{i:03d}" if kind == "icons" else f"scene.{i:03d}")
            entries.append((asset_id, asset_kind, extra_path(kind, i), size))
    if len(entries) != 1157:
        raise ValueError(f"unexpected complete inventory: {len(entries)}")
    with tempfile.TemporaryDirectory(prefix=f".{mode}-", dir=root) as tmp:
        staged = Path(tmp) / mode
        staged.mkdir()
        for asset_id, kind, rel, size in entries:
            source = Path("artwork/work") / mode / "assets" / rel
            if sha(source) != expected.get(str(source)):
                raise ValueError(f"WIP asset changed after preparation: {source}")
            with Image.open(source) as im:
                if im.format != "PNG" or im.size != size:
                    raise ValueError(f"invalid complete asset: {source}")
            dest = staged / "assets" / rel
            dest.parent.mkdir(parents=True, exist_ok=True)
            shutil.copyfile(source, dest)
            assets.append({"id": asset_id, "kind": kind, "path": str(Path("assets") / rel),
                           "width": size[0], "height": size[1], "sha256": sha(dest)})
        manifest = {"schema": 1, "id": mode, "mode": "faithful-hd",
                    "canvas": {"width": 960, "height": 720, "responsive": False}, "assets": assets}
        (staged / "manifest.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n")
        bundle_receipt = dict(receipt, status="complete-local-candidate",
                              manifest_sha256=sha(staged / "manifest.json"))
        (staged / "source-receipt.json").write_text(json.dumps(bundle_receipt, ensure_ascii=False, indent=2) + "\n")
        staged.rename(target)
    print(f"Local original-derived HD2 candidate: {target}, assets={len(assets)}; not a public release")


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("--ledger", type=Path, default=Path("artwork/manifests/hd2-production.json"))
    ap.add_argument("--mode", choices=["hd2-a", "hd2-b"])
    ap.add_argument("--guides", action="store_true")
    ap.add_argument("--bundle", action="store_true", help="完整覆蓋後才建立新的本機玩家包")
    ap.add_argument("--tileset", type=int, default=3)
    args = ap.parse_args()
    if args.guides:
        make_guides(args.tileset)
    elif args.mode:
        receipt = prepare(args.ledger, args.mode)
        if args.bundle:
            build_bundle(receipt, args.mode)
    else:
        ap.error("--mode or --guides is required")


if __name__ == "__main__":
    main()
