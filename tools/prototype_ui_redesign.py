#!/usr/bin/env python3
"""Render a disposable UI proposal, without changing the game or release.

Run in the existing Docker tool image. Inputs are the verified local release's
actual GUI screenshots and the pinned fonts recorded in ui-redesign-font-inputs.json.
The HTML previews presentation choices only; save/load and audio are not connected.
"""

from __future__ import annotations

import hashlib
import json
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "workplace/art-preview"
GUI = ROOT / "dist-all/v.1.0.0-20261029/smoke/full-local-linux-gui"
SCALE = 3
SIZE = (960, 720)
MAP = (24, 84, 888, 468)
BG = "#101820"
PANEL = "#17232e"
EDGE = "#46535d"
TEXT = "#e7ebed"
MUTED = "#9daab5"
ACCENT = "#e2c58c"
FONT_PATHS = {
    "noto": OUT / "ui-redesign-NotoSansCJKtc-Regular.otf",
    "rounded": OUT / "ui-redesign-GenSenRounded-TW-Regular.otf",
}
CHECKS: list[dict] = []


def sha(p: Path) -> str:
    return hashlib.sha256(p.read_bytes()).hexdigest()


def box(r):
    return tuple(round(v * SCALE) for v in r)


class Canvas:
    def __init__(self, family="rounded", size=SIZE):
        self.im = Image.new("RGB", tuple(x * SCALE for x in size), BG)
        self.draw = ImageDraw.Draw(self.im)
        self.family = family

    def rect(self, r, fill=PANEL, edge=EDGE, radius=5):
        self.draw.rounded_rectangle(
            box(r), radius=radius * SCALE, fill=fill, outline=edge, width=SCALE
        )

    def text(self, value, r, pixels=16, color=TEXT, center=False):
        font = ImageFont.truetype(str(FONT_PATHS[self.family]), pixels * SCALE)
        x0, y0, x1, y1 = box(r)
        # Use actual font metrics, and reject overflow rather than clip it.
        bounds = self.draw.textbbox((0, 0), value, font=font, anchor="lt")
        w, h = bounds[2] - bounds[0], bounds[3] - bounds[1]
        assert w <= x1 - x0 and h <= y1 - y0, (value, r, pixels, bounds)
        x = (x0 + x1 - w) // 2 if center else x0
        y = (y0 + y1 - h) // 2
        self.draw.text((x, y), value, font=font, fill=color, anchor="lt")
        CHECKS.append({"text": value, "font": self.family, "px": pixels,
                       "safe_rect": list(r), "measured": [w / SCALE, h / SCALE]})

    def finish(self):
        return self.im.resize(tuple(v // SCALE for v in self.im.size), Image.Resampling.LANCZOS)


def render(family, pixels, mode, map_im):
    c = Canvas(family)
    c.draw.rectangle(box((0, 0, 960, 44)), fill=PANEL)
    c.draw.line(box((0, 43, 960, 43)), fill=EDGE, width=SCALE)
    c.rect((12, 6, 204, 38))
    c.text("畫面", (24, 10, 58, 34), color=MUTED)
    for key, label, r in [
        ("original", "原版", (64, 9, 120, 35)),
        ("hd2-a", "A", (124, 9, 158, 35)),
        ("hd2-b", "B", (162, 9, 196, 35)),
    ]:
        if mode == key:
            c.rect(r, fill="#30404c", edge=ACCENT, radius=3)
        c.text(label, r, color=ACCENT if mode == key else MUTED, center=True)
    for r in [(212, 6, 316, 38), (324, 6, 436, 38), (444, 6, 564, 38),
              (572, 6, 684, 38), (692, 6, 772, 38), (780, 6, 860, 38), (868, 6, 948, 38)]:
        c.rect(r)
    c.text("字型 " + ("Noto" if family == "noto" else "圓體"), (218, 10, 310, 34), center=True)
    c.text("−", (330, 10, 354, 34), color=ACCENT, center=True)
    c.text(f"{pixels} px", (354, 10, 406, 34), center=True)
    c.text("+", (406, 10, 430, 34), color=ACCENT, center=True)
    c.text("配樂 現代", (450, 10, 558, 34), center=True)
    c.text("音量 − 6 +", (578, 10, 678, 34), center=True)
    c.text("存檔 F5", (698, 10, 766, 34), pixels=15, center=True)
    c.text("讀檔 F9", (786, 10, 854, 34), pixels=15, center=True)
    c.text("設定 F2", (874, 10, 942, 34), pixels=15, center=True)

    c.text("荒野遊俠", (20, 54, 220, 76), pixels=20, color=ACCENT)
    caption = {"original": "原版美術", "hd2-a": "HD2 A · 細緻像素", "hd2-b": "HD2 B · 手繪高解析"}[mode]
    c.text(caption, (232, 54, 560, 76), pixels=14, color=MUTED)
    c.text("01:00", (806, 54, 892, 76), pixels=16, color=MUTED, center=True)
    c.rect((20, 80, 892, 472), fill="#090f14", edge="#667078", radius=5)
    c.rect((904, 80, 948, 472), radius=5)
    c.text("輻射", (908, 96, 944, 120), pixels=12, color=MUTED, center=True)
    for i in range(12):
        y = 144 + 20 * i
        c.rect((917, y, 935, y + 12), fill="#293944", edge="#3a4c58", radius=2)
    c.text("0", (908, 416, 944, 452), pixels=18, color=MUTED, center=True)

    c.rect((16, 486, 948, 654), radius=6)
    c.text("訊息", (32, 498, 140, 520), pixels=14, color=ACCENT)
    c.text("隊伍 1", (816, 498, 932, 520), pixels=14, color=MUTED, center=True)
    c.draw.line(box((32, 528, 932, 528)), fill="#31414e", width=SCALE)
    c.text("已讀取快速存檔。", (32, 542, 916, 542 + pixels + 8), pixels=pixels)
    c.text("PageUp / PageDown  捲動", (32, 620, 520, 644), pixels=13, color=MUTED)
    commands = [("U", "使用"), ("E", "遭遇"), ("O", "排隊"), ("D", "分隊"),
                ("V", "換隊"), ("S", "存檔"), ("R", "無線"), ("P", "手札")]
    for i, (key, label) in enumerate(commands):
        x = 16 + i * 117
        c.rect((x, 666, x + 109, 704), radius=5)
        c.text(key, (x + 10, 672, x + 30, 698), pixels=15, color=ACCENT, center=True)
        c.text(label, (x + 38, 672, x + 99, 698), pixels=16, center=True)
    im = c.finish()
    # Preserve all viewport pixels from the real renderer. No resampling/cropping
    # beyond its established 864x384 viewport is applied to these art assets.
    im.paste(map_im, MAP[:2])
    assert im.crop(MAP).tobytes() == map_im.tobytes()
    return im


def comparison(left, right, left_label, right_label, filename):
    c = Canvas("noto", (1968, 792))
    c.text(left_label, (24, 16, 968, 60), pixels=22, color=ACCENT)
    c.text(right_label, (1008, 16, 1952, 60), pixels=22, color=ACCENT)
    im = c.finish()
    im.paste(left, (16, 64))
    im.paste(right, (992, 64))
    im.save(OUT / filename)


def html_preview():
    (OUT / "ui-redesign-preview.html").write_text("""<!doctype html>
<html lang="zh-Hant"><meta charset="utf-8"><meta name="viewport" content="width=device-width">
<title>荒野遊俠：介面排版草稿</title>
<style>
body{margin:24px;background:#0b1219;color:#e7ebed;font:16px/1.7 system-ui,sans-serif}
main{max-width:960px;margin:auto}h1{font-size:22px;color:#e2c58c;margin:0}p{color:#aab7c1}
.stage{position:relative;aspect-ratio:4/3;width:100%;box-shadow:0 0 0 1px #46535d}
img{display:block;width:100%;height:100%}button{cursor:pointer} .hit{position:absolute;
background:transparent;border:0;color:transparent;border-radius:5px;top:.83%;height:4.45%}
.hit:hover,.hit:focus-visible{background:#e2c58c22;outline:2px solid #e2c58c}
.hit:active{background:#e2c58c44}label{color:#e2c58c}a{color:#e2c58c}
</style><main><h1>介面排版草稿</h1>
<p>點工具列的「原版／A／B」、「字型」和「−／+」比較。這是排版草稿，存讀檔與音訊尚未接入遊戲。</p>
<div class="stage"><img id="screen" alt="介面排版草稿：圓體 18 像素，HD2 B">
<button class="hit" style="left:6.67%;width:5.83%" onclick="art('original')" aria-label="原版美術"></button>
<button class="hit" style="left:12.92%;width:3.54%" onclick="art('hd2-a')" aria-label="HD2 A 美術"></button>
<button class="hit" style="left:16.88%;width:3.54%" onclick="art('hd2-b')" aria-label="HD2 B 美術"></button>
<button class="hit" style="left:22.08%;width:10.83%" onclick="fontChoice()" aria-label="切換 Noto 或源泉圓體"></button>
<button class="hit" style="left:33.75%;width:3.13%" onclick="size(-2)" aria-label="縮小內文字級"></button>
<button class="hit" style="left:42.29%;width:3.13%" onclick="size(2)" aria-label="放大內文字級"></button>
</div><p id="status" aria-live="polite"></p>
<p>地圖取自已驗證封包的實際畫面，保持相同的 864×384 可視範圍。內文可比較 16／18／20 像素，工具列為 16 像素。</p>
<p>字型來源：<a href="https://github.com/notofonts/noto-cjk">Noto Sans CJK TC</a>、
<a href="https://github.com/ButTaiwan/gensen-font">源泉圓體</a>。
固定來源、SHA-256 與授權見 <a href="ui-redesign-font-inputs.json">字型來源紀錄</a>。</p>
</main><script>
let family='rounded', pixels=18, mode='hd2-b';
function update(){const s=document.getElementById('screen');s.src=`ui-redesign-${family}-${pixels}-${mode}.png`;
s.alt=`${family==='rounded'?'源泉圓體':'Noto Sans CJK TC'} ${pixels} px / ${mode}`;
document.getElementById('status').textContent=s.alt+'　　工具列單列常駐，細邊框。';}
function art(v){mode=v;update()}function fontChoice(){family=family==='noto'?'rounded':'noto';update()}
function size(d){pixels=Math.max(16,Math.min(20,pixels+d));update()}update();
</script></html>""", encoding="utf-8")


def main():
    assert (OUT.stat().st_uid, OUT.stat().st_gid) == (1000, 1000)
    font_inputs = json.loads((OUT / "ui-redesign-font-inputs.json").read_text())
    for src in font_inputs["sources"]:
        assert sha(ROOT / src["file"]) == src["sha256"], src["file"]
    shots = {"original": (GUI / "01-map-original.png", (24, 24, 888, 408)),
             "hd2-a": (GUI / "02-settings.png", MAP),
             "hd2-b": (GUI / "07-loaded.png", MAP)}
    maps = {k: Image.open(p).convert("RGB").crop(r) for k, (p, r) in shots.items()}
    for m in maps.values():
        assert m.size == (864, 384)
    for family in FONT_PATHS:
        for pixels in (16, 18, 20):
            for mode, map_im in maps.items():
                render(family, pixels, mode, map_im).save(OUT / f"ui-redesign-{family}-{pixels}-{mode}.png")
    rounded = Image.open(OUT / "ui-redesign-rounded-18-hd2-b.png")
    noto = Image.open(OUT / "ui-redesign-noto-18-hd2-b.png")
    current = Image.open(GUI / "07-loaded.png").convert("RGB")
    assert current.size == SIZE
    comparison(current, rounded, "目前：24 px 點陣字、原版放大邊框", "草稿：18 px 圓體、細邊框、單列工具列", "ui-redesign-before-after.png")
    comparison(noto, rounded, "A　Noto Sans CJK TC · 18 px", "B　源泉圓體 · 18 px", "ui-redesign-font-comparison.png")

    sample = Canvas("noto", (960, 336))
    for x, family in [(20, "noto"), (500, "rounded")]:
        sample.family = family
        sample.text("Noto Sans CJK TC" if family == "noto" else "源泉圓體 TW", (x, 14, x+440, 46), pixels=20, color=ACCENT)
        for i, px in enumerate((16, 18, 20)):
            y = 68 + i * 84
            sample.text(f"{px} px　隊伍、裝備、輻射與治療", (x, y, x+440, y+30), pixels=px)
            sample.text("Hell Razor　CON 25/25　$ 1,250", (x, y+32, x+440, y+62), pixels=px, color=MUTED)
    sample.finish().save(OUT / "ui-redesign-font-samples.png")
    html_preview()
    outputs = sorted(OUT.glob("ui-redesign-*.png")) + [OUT / "ui-redesign-preview.html"]
    receipt = {
        "status": "disposable prototype; not production; awaiting visual choice",
        "reference": "~/cht/mm3 single-row Noto toolbar; ~/cht/mm33 does not exist",
        "canvas": list(SIZE), "toolbar_px": 16, "body_sizes_px": [16, 18, 20],
        "proposal_default_px": 18, "line_height_at_18_px": 26,
        "map_source_pixels_preserved": True, "viewport": list(MAP),
        "screenshot_inputs": [{"path": str(p.relative_to(ROOT)), "sha256": sha(p), "crop": list(r)} for p,r in shots.values()],
        "font_inputs_sha256": sha(OUT / "ui-redesign-font-inputs.json"),
        "script_sha256": sha(Path(__file__)), "checked_text_boxes": len(CHECKS),
        "text_bounds": CHECKS,
        "outputs": [{"path": str(p.relative_to(ROOT)), "sha256": sha(p), "bytes": p.stat().st_size} for p in outputs],
    }
    (OUT / "ui-redesign-review.json").write_text(json.dumps(receipt, ensure_ascii=False, indent=2)+"\n")
    assert all((p.stat().st_uid, p.stat().st_gid) == (1000, 1000) for p in outputs)
    print(f"Rendered 18 native-size options and 3 comparisons; {len(CHECKS)} text-safe rectangles checked.")
    print("All map pixels equal the verified GUI inputs. Game code, saves and release packages unchanged.")


if __name__ == "__main__":
    main()
