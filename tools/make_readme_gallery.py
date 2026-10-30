#!/usr/bin/env python3
"""Build the versioned README gallery from verified packages, in Docker.

Run tools/readme_shots.sh. Runtime PNGs are copied byte for byte; the asset
contact sheets read the final v2 Windows ZIP, not a draft art directory.
"""
import gzip
import hashlib
import io
import json
import re
import sys
from collections import Counter
from pathlib import Path
import zipfile

from PIL import Image, ImageDraw, ImageFont, ImageOps


ROOT = Path(__file__).resolve().parents[1]
OUT = ROOT / "docs/images"
V1 = "v.1.0.0-20261029"
V2 = sys.argv[1] if len(sys.argv)>1 else "v.2.0.0-20261030"
if not re.fullmatch(r"v\.\d+\.\d+\.\d+-\d{8}",V2):
    raise SystemExit("A complete release version is required")
BG = "#101921"
PANEL = "#1b2833"
INK = "#edf0ec"
MUTED = "#a8b9c4"
GOLD = "#e8c579"
FONTS = {}
FONT_BYTES = gzip.decompress(
    (ROOT / "internal/presentation/fonts/GenSenRounded-TW-Regular.otf.gz").read_bytes()
)
RECEIPT = {"schema": 1, "purpose": "README versioned runtime and asset gallery",
           "runtime": [], "assets": {}, "outputs": [], "labels": []}


def sha(data):
    return hashlib.sha256(data).hexdigest()


def output_path(path):
    path.parent.mkdir(exist_ok=True)
    for target in (path.parent, path):
        if target.exists():
            assert target.stat().st_uid == 1000 and target.stat().st_gid == 1000, target
    return path


def font(size):
    if size not in FONTS:
        FONTS[size] = ImageFont.truetype(io.BytesIO(FONT_BYTES), size)
    return FONTS[size]


def label(im, text, rect, size=28, fill=INK, center=False):
    """Every caption must fit its explicit rectangle without clipping."""
    x0, y0, x1, y1 = rect
    draw = ImageDraw.Draw(im)
    for px in range(size, 9, -1):
        face = font(px)
        b = draw.textbbox((0, 0), text, font=face)
        if b[2] - b[0] <= x1 - x0 and b[3] - b[1] <= y1 - y0:
            break
    else:
        raise ValueError(f"Caption does not fit: {text}")
    x = x0 - b[0] + ((x1 - x0 - (b[2] - b[0])) // 2 if center else 0)
    y = y0 - b[1] + (y1 - y0 - (b[3] - b[1])) // 2
    box = draw.textbbox((x, y), text, font=face)
    assert x0 <= box[0] <= box[2] <= x1 and y0 <= box[1] <= box[3] <= y1
    draw.text((x, y), text, font=face, fill=fill)
    RECEIPT["labels"].append({"text": text, "safe": rect, "ink": box, "size": px})


def paste(im, source, rect, pixel=False):
    x0, y0, x1, y1 = rect
    resample = Image.Resampling.NEAREST if pixel else Image.Resampling.LANCZOS
    preview = ImageOps.contain(source.convert("RGBA"), (x1-x0, y1-y0), resample)
    xy = (x0 + (x1-x0-preview.width)//2, y0 + (y1-y0-preview.height)//2)
    im.paste(preview, xy, preview)


def save(im, name):
    target = output_path(OUT / name)
    im.save(target, optimize=True)
    RECEIPT["outputs"].append({"path": str(target.relative_to(ROOT)),
        "sha256": sha(target.read_bytes()), "size": list(im.size)})


def runtime(version, shots):
    source = ROOT / "dist-all" / version / "smoke/full-local-linux-gui"
    receipt = json.loads((source / "result.json").read_text())
    delivery = json.loads((source.parents[1] / "SHA256SUMS.json").read_text())
    assert receipt["status"] == "passed" and receipt["version"] == version
    assert receipt["exit_code"] == 0
    assert any(f["sha256"] == receipt["package_sha256"] for f in delivery["files"])
    hashes = {x["path"]: x["sha256"] for x in receipt["screenshots"]}
    for name, original in shots.items():
        data = (source / original).read_bytes()
        assert sha(data) == hashes[original], original
        with Image.open(io.BytesIO(data)) as image:
            assert image.size == (960, 720)
        destination = output_path(OUT / version.split("-", 1)[0].replace("v.", "v") / name)
        destination.write_bytes(data)
        RECEIPT["runtime"].append({"path": str(destination.relative_to(ROOT)),
            "sha256": sha(data), "source": str((source / original).relative_to(ROOT)),
            "version": version, "source_commit": delivery["source_commit"],
            "package_sha256": receipt["package_sha256"],
            "entry": receipt["entry"], "platform": "Linux Xvfb",
            "capture": "normal packaged launcher; no debug entry; uncropped logical 960x720"})


def overview(packs):
    im = Image.new("RGB", (1920, 1908), BG)
    label(im, "WASTELAND 荒野遊俠 · V2.0.0 素材總覽", (32, 24, 1888, 82), 44, GOLD)
    label(im, "兩種重繪風格，每版 1,157 張；以下為精選，完整清單另附總表。",
          (32, 90, 1888, 132), 28)
    for col, (mode, pack) in enumerate(packs.items()):
        x = 32 + col * 944
        title = "A 細緻像素" if col == 0 else "B 手繪高解析"
        ImageDraw.Draw(im).rounded_rectangle((x, 154, x+912, 1836), 12, fill=PANEL)
        label(im, title, (x+24, 172, x+888, 220), 36, GOLD)
        paste(im, pack["images"]["fullscreen.title"], (x+24, 232, x+888, 616), col == 0)
        label(im, "場景與人物 · 82 張", (x+24, 632, x+888, 672), 27, GOLD)
        for i, index in enumerate((0, 2, 13, 31, 49, 75)):
            px, py = x+24+(i % 3)*288, 688+(i // 3)*272
            paste(im, pack["images"][f"scene.{index:03}"], (px, py, px+276, py+242), col == 0)
            label(im, f"場景 {index:03}", (px, py+244, px+276, py+266), 17, MUTED, True)
        label(im, "地形 · 九組共 1,051 張，精選 72 張", (x+24, 1240, x+888, 1280), 27, GOLD)
        # Fixed index sampling, shared by A and B; no visual cherry-picking.
        tiles = [a for a in pack["assets"] if a["kind"] == "tile"]
        selected = []
        for group in range(9):
            part = [a for a in tiles if a["path"].split("/")[1] == f"tileset-{group}"]
            selected.extend(part[i*(len(part)-1)//7] for i in range(8))
        for i, asset in enumerate(selected):
            px, py = x+24+(i % 16)*54, 1292+(i // 16)*54
            paste(im, pack["images"][asset["id"]], (px, py, px+48, py+48), col == 0)
        label(im, "隊伍動畫 · 12 格", (x+24, 1584, x+888, 1624), 27, GOLD)
        for i, asset in enumerate(a for a in pack["assets"] if a["kind"] == "party-sprite"):
            px = x+24+i*64
            paste(im, pack["images"][asset["id"]], (px, 1636, px+56, 1692), col == 0)
        label(im, "圖示 · 10 張", (x+24, 1708, x+888, 1748), 27, GOLD)
        for i, asset in enumerate(a for a in pack["assets"] if a["kind"] == "icon"):
            px = x+24+i*64
            paste(im, pack["images"][asset["id"]], (px, 1760, px+56, 1816), col == 0)
    label(im, f"最終 {V2} 封包素材展示拼圖；不是遊戲截圖或可匯入的圖包。",
          (32, 1852, 1888, 1892), 24, MUTED)
    save(im, "v2.0.0/assets-overview.png")


def full_sheet(mode, pack):
    """Place every manifest asset exactly once, with a searchable index."""
    groups = []
    for n in range(9):
        group = [a for a in pack["assets"] if a["path"].startswith(f"assets/tileset-{n}/")]
        groups.append((f"地形組 {n} · {len(group)} 張", group, 32, 54, 70, 48, 48))
    groups += [
        ("隊伍動畫 · 12 格", [a for a in pack["assets"] if a["kind"] == "party-sprite"], 24, 72, 86, 64, 64),
        ("圖示 · 10 張", [a for a in pack["assets"] if a["kind"] == "icon"], 24, 72, 86, 64, 64),
        ("場景與人物 · 82 張", [a for a in pack["assets"] if a["kind"] == "scene"], 8, 216, 208, 204, 178),
        ("標題與結局 · 2 張", [a for a in pack["assets"] if a["kind"] in ("title", "ending")], 2, 876, 418, 864, 384),
    ]
    height = 150 + sum(54 + ((len(group)+cols-1)//cols)*cell_h + 20
                       for _, group, cols, _, cell_h, _, _ in groups) + 64
    im = Image.new("RGB", (1824, height), BG)
    title = "HD2 A 細緻像素" if mode == "hd2-a" else "HD2 B 手繪高解析"
    label(im, f"V2.0.0 · {title} · 全 1,157 張", (32, 24, 1792, 82), 40, GOLD)
    label(im, "依封包 manifest 索引排列；所有地形、場景、隊伍、圖示、標題與結局均列入。",
          (32, 92, 1792, 132), 26, MUTED)
    y, included = 150, []
    for title, group, cols, cell_w, cell_h, width, height in groups:
        label(im, title, (32, y, 1792, y+44), 28, GOLD)
        y += 54
        for i, asset in enumerate(group):
            x, py = 32+(i % cols)*cell_w, y+(i // cols)*cell_h
            paste(im, pack["images"][asset["id"]], (x, py, x+width, py+height), mode == "hd2-a")
            text = asset["id"].split(".")[-1]
            label(im, text, (x, py+height+2, x+width, py+cell_h-2), 13, MUTED, True)
            included.append(asset["id"])
        y += ((len(group)+cols-1)//cols)*cell_h+20
    assert len(included) == len(set(included)) == 1157
    assert set(included) == set(pack["images"])
    RECEIPT["assets"][mode]["sheet_ids"] = included
    label(im, f"來源：{V2} 正式本機完整包。此圖僅展示素材，完整圖包留在本機。",
          (32, y+4, 1792, y+44), 23, MUTED)
    save(im, f"v2.0.0/assets-{mode}-all.png")


def comparison():
    im = Image.new("RGB", (1968, 844), BG)
    label(im, "V1.0.0 舊版", (16, 16, 976, 66), 36, GOLD)
    label(im, "V2.0.0 新版", (992, 16, 1952, 66), 36, GOLD)
    for i, name in enumerate(("v1.0.0/map-hd2-b.png", "v2.0.0/map-hd2-b.png")):
        with Image.open(OUT / name) as screenshot:
            im.paste(screenshot, (16+i*976, 80))
    label(im, "同為 HD2 B，地圖視窗相同；呈現字型、邊框與工具列差異。訊息與動畫影格不同。",
          (16, 808, 1952, 836), 24, MUTED)
    save(im, "v2.0.0/ui-before-after.png")


def events_sheet(packs):
    """Every one of the original 82 scene indices, with both HD styles."""
    inventory = json.loads((ROOT / "artwork/manifests/original-art-inventory.json").read_text())
    assert inventory["scenes"] == {"count":82, "width":96, "height":84}
    expected = {f"scene.{i:03}" for i in range(82)}
    for pack in packs.values():
        assert {a["id"] for a in pack["assets"] if a["kind"]=="scene"} == expected
    im = Image.new("RGB", (2400, 3196), BG)
    label(im, "V2.0.0 · 全部 82 張事件／人物場景 · HD2 A／B 對照", (32,20,2368,78), 40, GOLD)
    label(im, "每個編號左側為 A 細緻像素，右側為 B 手繪高解析；000–081 全數列入，共 164 張。",
          (32,86,2368,126), 26, MUTED)
    entries=[]
    for i in range(82):
        x, y = 32+(i%6)*390, 148+(i//6)*212
        label(im, f"{i:03}    A                        B", (x,y,x+374,y+30), 19, GOLD, True)
        row={"index":i,"styles":{}}
        for offset,(mode,pack) in enumerate(packs.items()):
            asset=next(a for a in pack["assets"] if a["id"]==f"scene.{i:03}")
            hd=pack["images"][asset["id"]]
            assert hd.size==(288,252)
            source=ROOT/f"workplace/art-reference/scenes/{i:03}.png"
            with Image.open(source) as original:
                assert original.size==(96,84)
                original=original.convert("RGBA")
                nearest=original.resize(hd.size,Image.Resampling.NEAREST)
                assert hd.tobytes()!=nearest.tobytes(), (mode,i,"original upscale fallback")
                row["original_sha256"]=sha(source.read_bytes())
            px=x+offset*190
            paste(im,hd,(px,y+36,px+184,y+197),mode=="hd2-a")
            row["styles"][mode]={"sha256":asset["sha256"],"size":[288,252],
                                      "is_original_nearest_upscale":False}
        entries.append(row)
    label(im,f"來源：最終 {V2} 本機封包；素材展示拼圖，非遊戲截圖。",(32,3136,2368,3180),25,MUTED)
    RECEIPT["events"]={"count_per_style":82,"indices":entries,"status":"all-hd-no-original-upscale-fallback"}
    save(im,"v2.0.0/events-hd2-all.png")


def chrome_comparison():
    im=Image.new("RGB",(1968,844),BG)
    label(im,"HD 框線關閉 · 簡約細框",(16,16,976,66),36,GOLD)
    label(im,"HD 原版框線開啟 · A／B 預設",(992,16,1952,66),36,GOLD)
    for i,name in enumerate(("map-hd2-b-off.png","map-hd2-b.png")):
        with Image.open(OUT/"v2.0.0"/name) as screenshot:
            im.paste(screenshot,(16+i*976,80))
    label(im,"最終封包實機截圖；F2 的 C 框線可開關，地圖區域像素相同。",(16,808,1952,836),24,MUTED)
    save(im,"v2.0.0/frame-on-off.png")


def promo_preview():
    """Use reviewed final-video frames, retaining their provenance."""
    source = ROOT / "dist-all" / V2 / "promo"
    receipt = source / "metadata.json"
    if not receipt.is_file():
        return
    data = json.loads(receipt.read_text())
    if "hd-chrome-on-off-on" not in data.get("validation", []):
        return
    assert data["status"] == "passed" and data["manual_frame_review"] is True
    assert data["version"] == V2 and sha((source / data["file"]).read_bytes()) == data["sha256"]
    reviewed = {frame["path"]: frame["sha256"] for frame in data["reviewed_frames"]}
    samples = [
        ("02-hd2-a-live.png", "HD2 A · HD 原版框線"),
        ("03-hd2-b-live.png", "HD2 B · HD 原版框線"),
        ("09-chrome-off-map.png", "F2／C 關閉框線 · 27 秒"),
        ("10-chrome-restored-map.png", "F2／C 恢復框線 · 37 秒"),
    ]
    im = Image.new("RGB", (1328, 920), BG)
    label(im, "V2.0.0 · 重錄推廣影片 · 75 秒", (24, 16, 1304, 64), 34, GOLD)
    frames = []
    for i, (name, title) in enumerate(samples):
        path = "frames/" + name
        content = (source / path).read_bytes()
        assert sha(content) == reviewed[path], path
        x, y = 16 + (i % 2) * 656, 80 + (i // 2) * 400
        label(im, title, (x, y, x+640, y+32), 26, GOLD)
        with Image.open(io.BytesIO(content)) as frame:
            assert frame.size == (1280, 720)
            paste(im, frame, (x, y+36, x+640, y+396))
        frames.append({"path":str((source/path).relative_to(ROOT)), "sha256":sha(content)})
    label(im, "成片擷取影格拼圖；影片另含繁簡、字型／字級與遊俠中心操作。", (24, 884, 1304, 912), 23, MUTED)
    RECEIPT["promo"] = {"version": V2, "file": str((source/data["file"]).relative_to(ROOT)),
        "sha256":data["sha256"], "metadata_sha256":sha(receipt.read_bytes()), "frames":frames}
    save(im, "v2.0.0/promo-preview.png")


def main():
    assert OUT.is_dir()
    for directory in (OUT, ROOT/"tools"):
        assert directory.stat().st_uid == 1000 and directory.stat().st_gid == 1000
    runtime(V1, {"title.png": "00-title.png", "map-original.png": "01-map-original.png",
                 "map-hd2-b.png": "05-saved.png", "settings.png": "03-settings-b-modern.png"})
    runtime(V2, {"title.png": "00-title.png", "map-original.png": "01-original.png",
        "map-hd2-a.png": "05-art-a.png", "map-hd2-b.png": "06-art-b-modern.png",
        "character.png": "12-character.png", "items-hant.png": "13-items.png",
        "journal-hant.png": "15-journal.png", "settings-hant.png": "16-settings.png",
        "settings-hans.png": "20-hans-settings.png", "map-hans.png": "21-hans-map.png",
        "items-hans.png": "23-hans-items.png", "journal-hans.png": "25-hans-journal.png",
        "ranger-center.png": "10-ranger-center.png", "font-noto-16.png": "03-noto-16.png",
        "font-noto-18.png": "03-noto-18.png", "font-noto-20.png": "04-noto-20.png"})
    if V2 != "v.2.0.0-20261029":
        runtime(V2,{"map-hd2-a-off.png":"05-art-a-off.png",
                    "map-hd2-b-off.png":"06-art-b-modern-off.png"})
        chrome_comparison()
    archive = ROOT / f"dist-all/{V2}/full-local/windows-x64/wasteland-cht-{V2}-windows-x64-full-local.zip"
    delivery = json.loads((ROOT / f"dist-all/{V2}/SHA256SUMS.json").read_text())
    archive_sha = sha(archive.read_bytes())
    assert any(f["sha256"] == archive_sha for f in delivery["files"])
    RECEIPT["package"] = {"path": str(archive.relative_to(ROOT)), "sha256": archive_sha,
                          "source_commit": delivery["source_commit"]}
    packs = {}
    with zipfile.ZipFile(archive) as z:
        for mode in ("hd2-a", "hd2-b"):
            paths = [p for p in z.namelist() if p.endswith(f"/{mode}/manifest.json")]
            assert len(paths) == 1
            manifest_path = paths[0]
            data = z.read(manifest_path)
            manifest = json.loads(data)
            counts = Counter(a["kind"] for a in manifest["assets"])
            assert counts == {"tile":1051, "scene":82, "party-sprite":12, "icon":10, "title":1, "ending":1}
            images = {}
            for asset in manifest["assets"]:
                content = z.read(manifest_path.removesuffix("manifest.json") + asset["path"])
                assert sha(content) == asset["sha256"], asset["id"]
                with Image.open(io.BytesIO(content)) as source:
                    assert source.size == (asset["width"], asset["height"])
                    images[asset["id"]] = source.convert("RGBA")
            packs[mode] = {"assets":manifest["assets"], "images":images}
            RECEIPT["assets"][mode] = {"manifest_sha256":sha(data), "counts":dict(counts)}
            full_sheet(mode, packs[mode])
    overview(packs)
    events_sheet(packs)
    comparison()
    promo_preview()
    RECEIPT["generator"] = {"path": "tools/make_readme_gallery.py", "sha256": sha(Path(__file__).read_bytes())}
    RECEIPT["limits"] = "Runtime views show actual packages, not full-playthrough or original parity evidence. Asset sheets are labelled compositions, not gameplay."
    output_path(OUT / "gallery.json").write_text(json.dumps(RECEIPT, ensure_ascii=False, indent=2)+"\n", encoding="utf-8")
    print(f"PASS: {len(RECEIPT['runtime'])} verified runtime screenshots; 2 x 1157 assets; 164 HD events; {len(RECEIPT['outputs'])} compositions")


if __name__ == "__main__":
    main()
