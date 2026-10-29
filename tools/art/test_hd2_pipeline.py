import unittest
import contextlib
import io
import json
import os
import tempfile
from pathlib import Path

from PIL import Image, ImageDraw

from hd2_pipeline import build_bundle, clean_frame_edge, crop_boxes, gutter_runs, prepare, prepare_extras, sha, tile_indices


class HD2CropTest(unittest.TestCase):
    def test_measured_walk_rectangles_align_uneven_rows(self):
        im = Image.new("RGBA", (64, 160))
        draw = ImageDraw.Draw(im)
        draw.rectangle((20, 12, 44, 51), fill=(0, 255, 0, 255))
        draw.rectangle((20, 85, 44, 124), fill=(0, 255, 0, 255))
        batch = {"id": "walk", "cols": 1, "rows": 2, "count": 2,
                 "source_boxes": [[0, 0, 64, 64], [0, 73, 64, 137]]}
        boxes = crop_boxes(im, batch)
        self.assertEqual(im.crop(boxes[0]).tobytes(), im.crop(boxes[1]).tobytes())
        with self.assertRaisesRegex(ValueError, "measured source rectangles"):
            crop_boxes(im, {**batch, "framed": True})
        with self.assertRaisesRegex(ValueError, "measured source rectangles"):
            crop_boxes(im, {**batch, "source_boxes": [[0, 0, 64, 200]]})

    def test_rectangular_illustrations_keep_their_measured_slots(self):
        previous = Path.cwd()
        with tempfile.TemporaryDirectory() as tmp:
            try:
                os.chdir(tmp)
                originals = Path("workplace/art-reference/scenes")
                originals.mkdir(parents=True)
                im = Image.new("RGBA", (170, 150), (255, 0, 255, 255))
                draw = ImageDraw.Draw(im)
                colors = [(20, 40, 60, 255), (80, 100, 120, 255), (140, 160, 180, 255), (200, 220, 240, 255)]
                for i, color in enumerate(colors):
                    x0, x1 = ((5, 73), (79, 165))[i % 2]
                    y0, y1 = ((5, 62), (68, 145))[i // 2]
                    draw.rectangle((x0, y0, x1 - 1, y1 - 1), fill=color)
                    Image.new("RGB", (96, 84)).save(originals / f"{i:03d}.png")
                path = Path("scenes.png")
                im.save(path)
                batch = {"id": "scenes", "mode": "hd2-b", "kind": "scene", "status": "reviewed",
                         "start": 0, "count": 4, "cols": 2, "rows": 2, "framed": True,
                         "source": str(path), "reference": str(path),
                         "source_sha256": sha(path), "reference_sha256": sha(path)}
                pending, _, coverage, _ = prepare_extras({"illustration_batches": [batch]}, "hd2-b", Path("assets"))
                self.assertEqual(len(pending), 4)
                for i, (dest, raw) in enumerate(pending):
                    out = Image.open(io.BytesIO(raw))
                    self.assertEqual(str(dest), f"assets/scenes/{i:03d}.png")
                    self.assertEqual(out.size, (288, 252))
                    self.assertEqual(out.getpixel((144, 126)), colors[i])
                self.assertEqual(coverage[0]["ready"], 4)
            finally:
                os.chdir(previous)

    def test_sprite_transparency_and_literal_pink_survive(self):
        previous = Path.cwd()
        with tempfile.TemporaryDirectory() as tmp:
            try:
                os.chdir(tmp)
                im = Image.new("RGBA", (100, 50))
                ImageDraw.Draw(im).rectangle((0, 0, 49, 49), fill=(0, 0, 0, 255))
                ImageDraw.Draw(im).rectangle((60, 10, 89, 39), fill=(255, 0, 255, 255))
                im.putpixel((50, 0), (200, 200, 200, 2))
                im.putpixel((99, 49), (200, 200, 200, 3))
                path = Path("icons.png")
                im.save(path)
                batch = {"id": "icons", "mode": "hd2-a", "kind": "icons", "status": "reviewed",
                         "start": 0, "count": 2, "cols": 2, "rows": 1,
                         "source": str(path), "reference": str(path),
                         "source_sha256": sha(path), "reference_sha256": sha(path)}
                pending, _, _, _ = prepare_extras({"illustration_batches": [batch]}, "hd2-a", Path("assets"))
                self.assertEqual(Image.open(io.BytesIO(pending[0][1])).getextrema(),
                                 ((0, 0), (0, 0), (0, 0), (255, 255)))
                sprite = Image.open(io.BytesIO(pending[1][1]))
                self.assertEqual(sprite.getpixel((24, 24)), (255, 0, 255, 255))
                self.assertEqual(sprite.getpixel((0, 0))[3], 0)
                Image.new("RGBA", (100, 50)).save(path)
                batch["source_sha256"] = batch["reference_sha256"] = sha(path)
                with self.assertRaisesRegex(ValueError, "zero must be opaque black"):
                    prepare_extras({"illustration_batches": [batch]}, "hd2-a", Path("assets"))
            finally:
                os.chdir(previous)

    def test_party_frames_share_scale_when_a_leg_is_raised(self):
        previous = Path.cwd()
        with tempfile.TemporaryDirectory() as tmp:
            try:
                os.chdir(tmp)
                im = Image.new("RGBA", (100, 50))
                draw = ImageDraw.Draw(im)
                for i, bottom in enumerate((39, 34)):
                    x = i * 50
                    draw.rectangle((x + 20, 10, x + 29, 19), fill=(255, 0, 0, 255))
                    draw.rectangle((x + 20, 20, x + 29, bottom), fill=(0, 170, 0, 255))
                path = Path("party.png")
                im.save(path)
                batch = {"id": "party", "mode": "hd2-a", "kind": "party-walk", "status": "reviewed",
                         "start": 0, "count": 2, "cols": 2, "rows": 1,
                         "source": str(path), "reference": str(path),
                         "source_sha256": sha(path), "reference_sha256": sha(path)}
                pending, _, _, _ = prepare_extras({"illustration_batches": [batch]}, "hd2-a", Path("assets"))
                head_pixels = []
                for _, raw in pending:
                    sprite = Image.open(io.BytesIO(raw))
                    head_pixels.append([(x, y) for y in range(48) for x in range(48)
                                        if sprite.getpixel((x, y)) == (255, 0, 0, 255)])
                self.assertEqual(head_pixels[0], head_pixels[1])
                self.assertTrue(head_pixels[0])
            finally:
                os.chdir(previous)

    def test_incomplete_coverage_cannot_create_a_player_bundle(self):
        previous = Path.cwd()
        with tempfile.TemporaryDirectory() as tmp:
            try:
                os.chdir(tmp)
                receipt = {"mode": "hd2-a", "coverage": [{"ready": 0, "required": 1}], "extra_coverage": []}
                with self.assertRaisesRegex(ValueError, "incomplete HD2"):
                    build_bundle(receipt, "hd2-a")
                self.assertFalse(Path("artpacks").exists())
            finally:
                os.chdir(previous)

    def test_material_batch_keeps_original_slot_numbers(self):
        batch = {"id": "blue-planes", "start": 2, "count": 9,
                 "indices": [2, 14, 17, 18, 20, 33, 52, 82, 136]}
        with self.assertRaisesRegex(ValueError, "index mapping"):
            tile_indices({**batch, "indices": [2] * 9}, 163)
        with self.assertRaisesRegex(ValueError, "superseded indices"):
            tile_indices({**batch, "superseded_indices": [3]}, 163)

    def test_regrouped_material_reaches_its_original_png_slots(self):
        previous = Path.cwd()
        with tempfile.TemporaryDirectory() as tmp:
            try:
                os.chdir(tmp)
                Path("artwork/manifests").mkdir(parents=True)
                Path("artwork/manifests/original-art-inventory.json").write_text(
                    json.dumps({"tilesets": [{"id": 2, "count": 16}]}))
                originals = Path("workplace/art-reference/tiles/tileset-2")
                originals.mkdir(parents=True)
                for n in range(16):
                    Image.new("RGB", (16, 16)).save(originals / f"{n:03d}.png")
                old = Image.new("RGB", (128, 128))
                draw = ImageDraw.Draw(old)
                for n in range(16):
                    x, y = (n % 4) * 32, (n // 4) * 32
                    draw.rectangle((x, y, x + 31, y + 31), fill=(n * 10, 20, 30))
                old.save("old.png")
                new = Image.new("RGB", (128, 32))
                draw = ImageDraw.Draw(new)
                draw.rectangle((0, 0, 31, 31), fill=(85, 85, 255))
                draw.rectangle((32, 0, 63, 31), fill=(84, 86, 250))
                new.save("new.png")
                batches = []
                for name, start, count, rows in (("old", 0, 16, 4), ("new", 2, 2, 1)):
                    path = Path(name + ".png")
                    batches.append({"id": name, "mode": "hd2-a", "tileset": 2,
                                    "start": start, "count": count, "cols": 4, "rows": rows,
                                    "status": "reviewed", "source": str(path), "source_sha256": sha(path),
                                    "reference": str(path), "reference_sha256": sha(path)})
                batches[0]["superseded_indices"] = [2, 14]
                batches[1]["indices"] = [2, 14]
                ledger = Path("ledger.json")
                ledger.write_text(json.dumps({"batches": batches}))
                with contextlib.redirect_stdout(io.StringIO()):
                    prepare(ledger, "hd2-a")
                out = Path("artwork/work/hd2-a/assets/tileset-2")
                self.assertEqual(Image.open(out / "002.png").getpixel((24, 24)), (85, 85, 255, 255))
                self.assertEqual(Image.open(out / "014.png").getpixel((24, 24)), (84, 86, 250, 255))
                self.assertEqual(Image.open(out / "001.png").getpixel((24, 24)), (10, 20, 30, 255))
                self.assertEqual(len(list(out.glob("[0-9][0-9][0-9].png"))), 16)
                before = {str(p): sha(p) for p in out.glob("*.png")}
                batches[0].pop("superseded_indices")
                ledger.write_text(json.dumps({"batches": batches}))
                with self.assertRaisesRegex(ValueError, "duplicate tile"):
                    prepare(ledger, "hd2-a")
                self.assertEqual(before, {str(p): sha(p) for p in out.glob("*.png")})
            finally:
                os.chdir(previous)

    def test_exact_frame_coverage_boundary(self):
        im = Image.new("RGB", (9, 20))
        draw = ImageDraw.Draw(im)
        draw.rectangle((0, 0, 2, 12), fill=(255, 0, 255))  # exactly 65%: excluded
        draw.rectangle((3, 0, 5, 13), fill=(255, 0, 255))  # 70%: included
        self.assertEqual(gutter_runs(im, "x"), [(3, 5)])

    def test_measured_frames_follow_unequal_cells(self):
        # Equal division once put roof-colored strips into road tiles. A model
        # can keep four columns while giving them unequal widths and heights.
        im = Image.new("RGB", (510, 480), (255, 0, 255))
        xs = [(7, 101), (109, 229), (237, 362), (370, 503)]
        ys = [(7, 111), (119, 239), (247, 352), (360, 473)]
        colors = []
        draw = ImageDraw.Draw(im)
        for n in range(16):
            color = (10 + n * 10, 80, 40)
            colors.append(color)
            x0, x1 = xs[n % 4]
            y0, y1 = ys[n // 4]
            draw.rectangle((x0, y0, x1 - 1, y1 - 1), fill=color)
        boxes = crop_boxes(im, {"id": "unequal", "cols": 4, "rows": 4,
                                "count": 16, "framed": True})
        for n, box in enumerate(boxes):
            self.assertEqual(im.crop(box).getcolors(),
                             [(im.crop(box).width * im.crop(box).height, colors[n])])

    def test_missing_separator_is_rejected(self):
        im = Image.new("RGB", (400, 400), (25, 25, 25))
        with self.assertRaisesRegex(ValueError, "frame count differs"):
            crop_boxes(im, {"id": "missing", "cols": 4, "rows": 4,
                            "count": 16, "framed": True})

    def test_only_technical_edge_color_is_extended(self):
        tile = Image.new("RGBA", (256, 256), (20, 30, 20, 255))
        original = Image.new("RGB", (16, 16), (0, 170, 0))
        for x in range(256):
            tile.putpixel((x, 0), (255, 0, 255, 255))
        # Genuine original purple and an interior purple motif remain intact.
        original.putpixel((0, 0), (170, 0, 170))
        tile.putpixel((0, 0), (170, 0, 170, 255))
        tile.putpixel((100, 100), (200, 0, 200, 255))
        cleaned, count = clean_frame_edge(tile, original)
        self.assertGreater(count, 0)
        self.assertEqual(cleaned.getpixel((50, 0)), (20, 30, 20, 255))
        self.assertEqual(cleaned.getpixel((0, 0)), (170, 0, 170, 255))
        self.assertEqual(cleaned.getpixel((100, 100)), (200, 0, 200, 255))
        self.assertEqual(tile.getpixel((50, 0)), (255, 0, 255, 255))

    def test_large_contamination_is_rejected(self):
        with self.assertRaisesRegex(ValueError, "no uncontaminated"):
            clean_frame_edge(Image.new("RGB", (256, 256), (255, 0, 255)),
                             Image.new("RGB", (16, 16), (0, 170, 0)))

    def test_muted_violet_mountain_is_not_technical_magenta(self):
        tile = Image.new("RGB", (256, 256), (160, 20, 175))
        cleaned, count = clean_frame_edge(tile, Image.new("RGB", (16, 16), (170, 85, 0)))
        self.assertEqual(count, 0)
        self.assertEqual(cleaned.tobytes(), tile.tobytes())


if __name__ == "__main__":
    unittest.main()
