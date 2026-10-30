#!/usr/bin/env python3
"""Disposable HD original-frame proposal; not connected to the game.

Docker: read /src only, mount workplace/art-preview writable at /out.
Reference: docs/re/124-screen-border-and-right-column.md. This is a modern
redrawing of the silver bevel and right title/meter, not original-pixel parity.
"""
import gzip
import hashlib
import io
import json
from pathlib import Path

from PIL import Image, ImageDraw, ImageFont


ROOT = Path('/src')
OUT = Path('/out')
GUI = ROOT / 'dist-all/v.2.0.0-20261029/smoke/full-local-linux-gui'
SCALE = 3
FONTS = {}
FONT = gzip.decompress((ROOT / 'internal/presentation/fonts/GenSenRounded-TW-Regular.otf.gz').read_bytes())


def font(px):
    if px not in FONTS:
        FONTS[px] = ImageFont.truetype(io.BytesIO(FONT), px*SCALE)
    return FONTS[px]


def rect(r):
    return tuple(int(n*SCALE) for n in r)


def label(im, text, r, px=16, fill='#e7ebed'):
    d = ImageDraw.Draw(im)
    x0,y0,x1,y1 = rect(r)
    f = font(px)
    b = d.textbbox((0,0),text,font=f)
    assert b[2]-b[0] <= x1-x0 and b[3]-b[1] <= y1-y0, text
    xy = (x0+(x1-x0-b[2]+b[0])//2-b[0], y0+(y1-y0-b[3]+b[1])//2-b[1])
    d.text(xy,text,font=f,fill=fill)


def metal(im, r, width=8):
    """Silver, lit top/left and recessed lower/right, with mitered joins."""
    d = ImageDraw.Draw(im)
    x0,y0,x1,y1 = rect(r)
    for offset in range(width*SCALE):
        n = offset/SCALE
        if n < 1: light,dark='#e9edf0','#11181e'
        elif n < 3: light,dark='#b5bdc3','#454f58'
        elif n < width-2: light,dark='#929ca5','#66717b'
        elif n < width-1: light,dark='#454f58','#b5bdc3'
        else: light,dark='#11181e','#e9edf0'
        a,b,c,e=x0+offset,y0+offset,x1-1-offset,y1-1-offset
        d.line((a,b,c,b),fill=light)
        d.line((a,b,a,e),fill=light)
        d.line((a,e,c,e),fill=dark)
        d.line((c,b,c,e),fill=dark)


def frame(source):
    im = source.convert('RGB').resize((960*SCALE,720*SCALE),Image.Resampling.NEAREST)
    metal(im,(16,76,896,476))
    metal(im,(16,486,948,654))
    d=ImageDraw.Draw(im)
    d.rectangle(rect((904,80,948,472)),fill='#111b25')
    metal(im,(904,80,948,472),4)
    for i,ch in enumerate('WASTELAND'):
        # Native-size glyphs, rather than an enlarged copy of the old bitmap.
        label(im,ch,(910,88+i*22,942,110+i*22),24,'#d8dee3')
    d.rectangle(rect((912,292,940,416)),fill='#152744')
    metal(im,(912,292,940,416),3)
    for i in range(12):
        y=298+i*9
        d.rectangle(rect((918,y,934,y+5)),fill='#102a54')
        d.line(rect((918,y,934,y)),fill='#315c9a',width=SCALE)
    label(im,'輻射',(908,424,944,444),12,'#a8b8c5')
    label(im,'—',(908,446,944,464),14,'#a8b8c5')
    label(im,'HD 原版框線',(578,54,788,76),14,'#e2c58c')
    result=im.resize((960,720),Image.Resampling.LANCZOS)
    # Raster antialiasing must never alter map/party pixels.
    result.paste(source.crop((24,84,888,468)),(24,84))
    assert result.crop((24,84,888,468)).tobytes()==source.crop((24,84,888,468)).tobytes()
    return result


def write(im,name):
    path=OUT/name
    assert not path.exists() or path.stat().st_uid==1000
    im.save(path)
    return {'path':'workplace/art-preview/'+name,'sha256':hashlib.sha256(path.read_bytes()).hexdigest()}


def compare(left,right,labels,name):
    im=Image.new('RGB',(1968*SCALE,812*SCALE),'#101820')
    label(im,labels[0],(16,12,976,58),26,'#e2c58c')
    label(im,labels[1],(992,12,1952,58),26,'#e2c58c')
    label(im,'框線樣稿，尚未接入遊戲；地圖像素、字型與工具列沿用實際畫面。',
          (16,786,1952,810),18,'#a8b8c5')
    result=im.resize((1968,812),Image.Resampling.LANCZOS)
    result.paste(left,(16,64)); result.paste(right,(992,64))
    return write(result,name)


def main():
    assert OUT.is_dir() and OUT.stat().st_uid==1000
    a=Image.open(GUI/'05-art-a.png').convert('RGB')
    b=Image.open(GUI/'06-art-b-modern.png').convert('RGB')
    aa,bb=frame(a),frame(b)
    results=[write(aa,'hd-frame-proposal-a-on.png'),write(bb,'hd-frame-proposal-b-on.png')]
    results.append(compare(aa,bb,('A · HD 原版框線開啟','B · HD 原版框線開啟'),'hd-frame-proposal-ab.png'))
    results.append(compare(b,bb,('框線關閉 · 保留目前簡約介面','框線開啟 · HD 原版風格'),'hd-frame-proposal-on-off.png'))
    # Show the proposed toggle in the actual settings header, beside language.
    settings=bb.copy()
    old=Image.open(GUI/'16-settings.png').convert('RGB')
    settings.paste(old.crop((16,154,948,654)),(16,154))
    im=settings.resize((960*SCALE,720*SCALE),Image.Resampling.NEAREST)
    metal(im,(16,154,948,654))
    ImageDraw.Draw(im).rounded_rectangle(rect((356,162,600,204)),radius=4*SCALE,
        outline='#e2c58c',fill='#253747',width=SCALE)
    label(im,'C 框線：HD 原版',(364,170,592,196),16,'#e2c58c')
    results.append(write(im.resize((960,720),Image.Resampling.LANCZOS),'hd-frame-proposal-settings.png'))
    payload={'status':'disposable-proposal-awaiting-user-style-decision','recipe':
        'tools/prototype_hd_chrome.py','source':'v.2.0.0-20261029 final Linux GUI',
        'map_rect':[24,84,888,468],'map_pixels_unchanged':True,'files':results}
    p=OUT/'hd-frame-proposal.json'; assert not p.exists() or p.stat().st_uid==1000
    p.write_text(json.dumps(payload,ensure_ascii=False,indent=2)+'\n')
    print('PROPOSAL: A/B on, off/on comparison and F2 toggle; map pixels preserved')


if __name__=='__main__':
    main()
