#!/usr/bin/env python3
"""Compose and validate a local v2 trailer from newly recorded package footage."""
import gzip
import hashlib
import io
import json
from pathlib import Path
import re
import shutil
import subprocess
import sys
import tempfile

from PIL import Image, ImageDraw, ImageFont


def sha(p):
    with p.open("rb") as f:
        return hashlib.file_digest(f, "sha256").hexdigest()


def run(args, **kwargs):
    return subprocess.run(args, check=True, timeout=180, **kwargs)


def probe(p):
    return json.loads(subprocess.check_output(["ffprobe", "-v", "error", "-show_format", "-show_streams", "-of", "json", str(p)]))


def main():
    version = sys.argv[1]
    assert re.fullmatch(r"v\.\d+\.\d+\.\d+-\d{8}", version)
    raw, out = Path("/raw"), Path("/out")
    capture = json.loads((raw / "capture.json").read_text())
    assert capture["status"] == "passed" and capture["version"] == version and capture["exit_code"] == 0
    package = Path(f"/src/dist-all/{version}/full-local/linux-x64/wasteland-cht-{version}-linux-x64-full-local.AppImage")
    assert sha(package) == capture["package_sha256"]
    assert sha(raw / "driver.sh") == capture["driver_sha256"]
    for f in capture["clips"] + capture["music"]:
        assert sha(raw / f["path"]) == f["sha256"]
    movie = out / f"wasteland-cht-{version}-promo.mp4"
    assert not movie.exists(), "Refusing to overwrite a delivered trailer"
    fonts = gzip.decompress(Path("/src/internal/presentation/fonts/GenSenRounded-TW-Regular.otf.gz").read_bytes())
    fields = []

    def label(im, text, rect, maximum=32, color=(231, 235, 237)):
        draw = ImageDraw.Draw(im)
        x0, y0, x1, y1 = rect
        for size in range(maximum, 11, -1):
            font = ImageFont.truetype(io.BytesIO(fonts), size)
            box = draw.textbbox((0, 0), text, font=font)
            if box[2] - box[0] <= x1-x0 and box[3] - box[1] <= y1-y0:
                break
        else:
            raise AssertionError(f"Text cannot fit: {text}")
        x = x0 + (x1-x0-(box[2]-box[0]))//2 - box[0]
        y = y0 + (y1-y0-(box[3]-box[1]))//2 - box[1]
        ink = draw.textbbox((x,y), text, font=font)
        assert x0 <= ink[0] <= ink[2] <= x1 and y0 <= ink[1] <= ink[3] <= y1
        draw.text((x,y), text, font=font, fill=color)
        fields.append({"text":text,"safe":rect,"ink":ink,"pixels":size})

    out.mkdir(exist_ok=True)
    frames = out / "frames"; frames.mkdir(exist_ok=True)
    board = [
        {"kind":"card", "name":"intro", "seconds":5},
        {"kind":"live", "name":"original-live", "seconds":8, "left":["原版", "美術"], "right":["v2.0", "繁體中文"]},
        {"kind":"live", "name":"hd2-a-live", "seconds":9, "left":["HD2 A", "細緻像素"], "right":["同一世界", "自由行走"]},
        {"kind":"live", "name":"hd2-b-live", "seconds":9, "left":["HD2 B", "手繪高解析"], "right":["原版素材", "重新繪製"]},
        {"kind":"live", "name":"language-live", "offset":4, "seconds":12, "left":["繁體中文", "／", "简体中文"], "right":["工具列", "直接切換"]},
        {"kind":"live", "name":"toolbar-live", "offset":8, "seconds":12, "left":["圓體", "Noto", "16／18／20"], "right":["細邊框", "單列工具列"]},
        {"kind":"live", "name":"facility-live", "seconds":11, "left":["遊俠中心", "隊伍管理"], "right":["完整畫面", "實際操作"]},
        {"kind":"card", "name":"outro", "seconds":7},
    ]
    assert sum(x["seconds"] for x in board) == 73
    with tempfile.TemporaryDirectory(prefix="wl-promo-") as temporary:
        temp = Path(temporary)
        segments = []
        time = 0
        for i, item in enumerate(board):
            item["start"] = time; time += item["seconds"]; item["end"] = time
            png = temp / f"{i:02}.png"
            segment = temp / f"{i:02}.mp4"
            if item["kind"] == "card":
                im = Image.new("RGB", (1280,720), (16,24,32))
                d = ImageDraw.Draw(im)
                d.rounded_rectangle((64,56,1216,664), radius=12, outline=(70,83,93), width=2)
                if item["name"] == "intro":
                    label(im,"WASTELAND",(112,112,1168,230),90,(226,197,140))
                    label(im,"荒野遊俠",(112,250,1168,360),78)
                    label(im,"v2.0",(112,386,1168,460),54,(226,197,140))
                    label(im,"繁簡中文 · 三種美術 · 全新介面",(112,490,1168,548),34)
                else:
                    label(im,"荒野遊俠 v2.0",(112,110,1168,202),64,(226,197,140))
                    label(im,"繁體中文／简体中文",(112,242,1168,292),36)
                    label(im,"原版／HD2 A／HD2 B",(112,310,1168,360),36)
                    label(im,"Linux · Windows · macOS 本機測試版",(112,384,1168,434),32)
                    label(im,"配樂為專案自作，非 1988 年原聲",(112,472,1168,516),26)
                    label(im,"github.com/wicanr2/wasteland_cht",(112,534,1168,572),26)
                label(im,version,(112,604,1168,640),22,(157,170,181))
                im.save(png)
                run(["ffmpeg","-v","error","-y","-loop","1","-i",str(png),"-t",str(item["seconds"]),"-r","25",
                     "-vf",f"fade=t=in:st=0:d=0.2,fade=t=out:st={item['seconds']-.2}:d=0.2,format=yuv420p",
                     "-threads","2","-c:v","libx264","-preset","veryfast","-crf","20",str(segment)])
            else:
                assert float(probe(raw/(item["name"]+".mp4"))["format"]["duration"]) >= item.get("offset",0)+item["seconds"]
                overlay = Image.new("RGBA",(1280,720),(0,0,0,0))
                d=ImageDraw.Draw(overlay)
                d.rectangle((0,0,159,719),fill=(16,24,32,255))
                d.rectangle((1120,0,1279,719),fill=(16,24,32,255))
                for side,x in [("left",8),("right",1128)]:
                    lines=item[side]
                    for n,text in enumerate(lines):
                        label(overlay,text,(x,270+n*62,x+144,324+n*62),28,(226,197,140) if n==0 else (231,235,237))
                label(overlay,"v2.0",(1130,654,1270,704),22,(157,170,181))
                overlay.save(png)
                run(["ffmpeg","-v","error","-y","-ss",str(item.get("offset",0)),"-i",str(raw/(item["name"]+".mp4")),"-i",str(png),
                     "-filter_complex","[0:v]fps=25,pad=1280:720:160:0:color=0x101820[v];[v][1:v]overlay=0:0,format=yuv420p",
                     "-t",str(item["seconds"]),"-an","-threads","2","-c:v","libx264","-preset","veryfast","-crf","20",str(segment)])
            segments.append(segment)
            print("COMPOSED",item["name"],flush=True)
        playlist=temp/"concat.txt"
        playlist.write_text("\n".join(f"file '{p}'" for p in segments)+"\n")
        silent=temp/"silent.mp4"
        run(["ffmpeg","-v","error","-y","-f","concat","-safe","0","-i",str(playlist),"-c","copy",str(silent)])
        assert abs(float(probe(silent)["format"]["duration"])-73)<.1
        theme,combat=raw/"music/theme.ogg",raw/"music/combat.ogg"
        assert float(probe(theme)["format"]["duration"])>=45
        assert float(probe(combat)["format"]["duration"])>=30
        candidate=temp/"candidate.mp4"
        audio=("[1:a]asplit=2[t0][t1];[t0]atrim=0:28,asetpts=PTS-STARTPTS[a0];"
               "[2:a]atrim=0:30,asetpts=PTS-STARTPTS[a1];[t1]atrim=28:45,asetpts=PTS-STARTPTS[a2];"
               "[a0][a1]acrossfade=d=1:c1=tri:c2=tri[a01];[a01][a2]acrossfade=d=1:c1=tri:c2=tri,"
               "volume=0.9,afade=t=in:st=0:d=0.8,afade=t=out:st=71:d=2,aresample=48000[a]")
        run(["ffmpeg","-v","error","-y","-i",str(silent),"-i",str(theme),"-i",str(combat),"-filter_complex",audio,
             "-map","0:v:0","-map","[a]","-c:v","copy","-c:a","aac","-b:a","192k","-movflags","+faststart",str(candidate)])
        info=probe(candidate)
        v=next(x for x in info["streams"] if x["codec_type"]=="video")
        a=next(x for x in info["streams"] if x["codec_type"]=="audio")
        assert (v["width"],v["height"],v["codec_name"],a["codec_name"])==(1280,720,"h264","aac")
        assert abs(float(v["duration"])-73)<.1 and abs(float(a["duration"])-73)<.1
        run(["ffmpeg","-v","error","-xerror","-i",str(candidate),"-threads","2","-f","null","-"],stdout=subprocess.DEVNULL)
        volume=run(["ffmpeg","-hide_banner","-nostats","-i",str(candidate),"-vn","-af","volumedetect","-f","null","-"],capture_output=True,text=True).stderr
        mean=float(re.search(r"mean_volume: ([-\d.]+) dB",volume)[1])
        peak=float(re.search(r"max_volume: ([-\d.]+) dB",volume)[1])
        assert mean > -45 and peak < 0,(mean,peak)
        scan=run(["ffmpeg","-hide_banner","-nostats","-i",str(candidate),"-an","-vf",
                  "blackdetect=d=0.2:pix_th=0.02,freezedetect=n=-50dB:d=3","-threads","2","-f","null","-"],capture_output=True,text=True).stderr
        black=[float(x) for x in re.findall(r"black_duration:([\d.]+)",scan)]
        assert not black or max(black)<=.4,black
        freezes=[]
        started=None
        for line in scan.splitlines():
            m=re.search(r"freeze_start: ([\d.]+)",line)
            if m: started=float(m[1])
            m=re.search(r"freeze_end: ([\d.]+)",line)
            if m and started is not None:
                end=float(m[1])
                movement=max((max(0,min(end,b)-max(started,a)) for a,b in [(5,13),(13,22),(22,31),(55,66)]),default=0)
                assert movement<=6,(started,end)
                freezes.append({"start":started,"end":end,"movement_overlap":movement}); started=None
        # Whole-frame average difference dilutes small, valid text changes over
        # a largely static map. Recognize the actual language/size glyph regions
        # against the recorded PNG states instead. The mask comes from reference
        # differences; half the nearest reference separation gives disjoint sets.
        def ui_states(start, duration, crop, refs):
            x0,y0,x1,y1=crop; w,h=x1-x0,y1-y0
            templates=[Image.open(raw/(name+'.png')).convert('RGB').crop(crop).tobytes() for name in refs.values()]
            mask=[i for i in range(w*h*3) if len({t[i] for t in templates})>1]
            assert mask
            def distance(a,b): return sum(abs(a[i]-b[i]) for i in mask)/len(mask)
            separation=min(distance(a,b) for i,a in enumerate(templates) for b in templates[i+1:])
            assert separation>0
            b=subprocess.check_output(['ffmpeg','-v','error','-ss',str(start),'-i',str(candidate),'-t',str(duration),
                '-vf',f'crop={w}:{h}:{x0+160}:{y0},fps=2,format=rgb24','-f','rawvideo','-'],timeout=60)
            stride=w*h*3; assert len(b)%stride==0
            counts={key:0 for key in refs}; sequence=[]; best_errors={key:[] for key in refs}
            for i in range(0,len(b),stride):
                frame=b[i:i+stride]; errors=[distance(frame,t) for t in templates]
                best=min(range(len(errors)),key=errors.__getitem__)
                if errors[best] < separation/2:
                    key=list(refs)[best]; counts[key]+=1; best_errors[key].append(round(errors[best],3))
                    if not sequence or sequence[-1]!=key: sequence.append(key)
            print('UI STATES',counts,'sequence',sequence,'reference separation',separation,flush=True)
            return {'samples':counts,'sequence':sequence,'reference_separation':separation,'errors':best_errors}
        language_states=ui_states(31,12,(612,162,844,204),{'zh-Hant':'lang-hant','zh-Hans':'lang-hans'})
        assert all(n>=4 for n in language_states['samples'].values()),language_states
        assert language_states['sequence']==['zh-Hant','zh-Hans'],language_states
        size_states=ui_states(43,12,(354,10,406,34),{'16':'font-16','18':'font-18','20':'font-20'})
        assert all(n>=2 for n in size_states['samples'].values()),size_states
        samples=[]
        for i,item in enumerate(board):
            t=item["start"]+item["seconds"]/2
            p=frames/f"{i:02}-{item['name']}.png"
            run(["ffmpeg","-v","error","-y","-ss",str(t),"-i",str(candidate),"-frames:v","1",str(p)])
            samples.append(p)
        sheet=Image.new("RGB",(1024,1248),(16,24,32))
        for i,p in enumerate(samples):
            im=Image.open(p).convert("RGB"); im=im.resize((512,288),Image.Resampling.LANCZOS)
            x=(i%2)*512;y=(i//2)*312
            sheet.paste(im,(x,y+24)); label(sheet,board[i]["name"],(x+6,y,x+506,y+24),18)
        sheet.save(out/"contact-sheet.png")
        shutil.copyfile(candidate,movie)
        shutil.copyfile(raw/"capture.json",out/"capture.json")
        shutil.copyfile(raw/"driver.sh",out/"record-driver.sh")
        if (raw/"metadata-writer.sh").is_file(): shutil.copyfile(raw/"metadata-writer.sh",out/"capture-metadata-writer.sh")
        shutil.copyfile(Path(__file__),out/"render-driver.py")
        (out/"ffprobe.json").write_text(json.dumps(info,ensure_ascii=False,indent=2)+"\n")
        (out/"analysis.log").write_text(volume+"\n"+scan)
        receipt={"schema":1,"status":"passed","manual_frame_review":False,"version":version,"file":movie.name,"sha256":sha(movie),"bytes":movie.stat().st_size,
          "duration":73,"video":{"width":1280,"height":720,"codec":"h264","fps":25},
          "audio":{"codec":"aac","mean_db":mean,"peak_db":peak,"mix":"modern theme 28s + combat 30s + theme 17s; two 1s crossfades"},
          "validation":["whole-file-decode","duration-match","non-silent","peak-below-zero","black-frame-scan","movement-freeze-scan","language-glyph-states","size-glyph-states","caption-ink-bounds"],
          "intentional_stills":[[0,5],[66,73]],"small_ui_change_intervals":[[31,43],[43,55]],
          "language_states":language_states,"size_states":size_states,"freeze_intervals":freezes,"text_fields":fields,"storyboard":board,
          "capture_manifest_sha256":sha(raw/"capture.json"),"package_sha256":capture["package_sha256"],
          "renderer_sha256":sha(Path(__file__)),"source_music":capture["music"],
          "rights":"LOCAL ONLY: original/derived game visuals, existing approved project score; not uploaded or publicly licensed",
          "limits":"Technical audio checks only; no new human listening acceptance or native Windows/macOS validation."}
        (out/"metadata.json").write_text(json.dumps(receipt,ensure_ascii=False,indent=2)+"\n")
        print("PASS",movie.name,"73 seconds",mean,"dB",peak,"dB",flush=True)


if __name__=="__main__": main()
