import unittest
import contextlib
import io
import json
import os
import tempfile
from pathlib import Path

from PIL import Image, ImageDraw

from hd2_pipeline import clean_frame_edge, crop_boxes, gutter_runs, prepare, sha, tile_indices


class HD2CropTest(unittest.TestCase):
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
