#!/usr/bin/env python3
"""Docker 內渲染已定案的十首現代配樂；全數驗證後才安裝。"""

import argparse
import hashlib
import json
import math
import shutil
import subprocess
import sys
import tempfile
from pathlib import Path


def run(args):
    result = subprocess.run(args, capture_output=True, text=True)
    if result.returncode:
        raise RuntimeError(f"{args[0]} failed ({result.returncode}): {result.stderr[-4000:]}")
    return result


def measure(path, filters="anull"):
    result = run(["ffmpeg", "-hide_banner", "-nostats", "-i", str(path),
                  "-af", filters + ",loudnorm=I=-17:TP=-2:LRA=9:print_format=json",
                  "-f", "null", "-"])
    data = json.loads(result.stderr[result.stderr.rindex("{"):])
    return {key: float(data[key]) for key in ("input_i", "input_tp", "input_lra")}


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    if not Path("/.dockerenv").exists():
        raise SystemExit("請在 docker/media.Dockerfile 的 modern stage 內執行")
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("out", type=Path)
    ap.add_argument("--install-to", type=Path)
    args = ap.parse_args()
    root = Path(__file__).resolve().parent.parent
    out = args.out.resolve()
    if out == root / "workplace/music/modern":
        raise SystemExit("先寫候選目錄；驗證後才以 --install-to 安裝")
    out.mkdir(parents=True, exist_ok=True)
    run([sys.executable, str(root / "tools/make_uplifting_music.py"), str(out)])
    score = json.loads((out / "score.json").read_text())
    expected = {"theme", "desert", "night", "town", "facility", "sewer", "vegas", "base", "combat", "ending"}
    if {cue["name"] for cue in score["cues"]} != expected or len(score["cues"]) != 10:
        raise ValueError("必須有十首完整場景曲目")
    sf = Path("/usr/share/sounds/sf2/FluidR3_GM.sf2")
    shutil.copyfile("/usr/share/doc/fluid-soundfont-gm/copyright", out / "FluidR3-GM-LICENSE.txt")
    receipts = []
    for cue in score["cues"]:
        name = cue["name"]
        samples = round(cue["seconds"] * 44100)
        filters = (f"atrim=start_sample={samples}:end_sample={samples * 2},"
                   "asetpts=PTS-STARTPTS,highpass=f=32,lowpass=f=16000")
        gain = None
        files = []
        for stem in ["master", *cue["stems"]]:
            midi = out / f"{name}-{stem}.mid"
            wav = out / f"{name}-{stem}.wav"
            run(["fluidsynth", "-ni", "-o", "synth.cpu-cores=2", "-o",
                 "synth.chorus.active=0", "-F", str(wav), "-r", "44100", str(sf), str(midi)])
            if gain is None:
                levels = measure(wav, filters)
                if not all(math.isfinite(v) for v in levels.values()):
                    raise ValueError(f"silent or invalid master: {name}")
                gain = min(-17 - levels["input_i"], -2 - levels["input_tp"])
            path = out / (f"{name}.ogg" if stem == "master" else f"{name}-{stem}.ogg")
            run(["ffmpeg", "-y", "-loglevel", "error", "-i", str(wav), "-af",
                 filters + f",volume={gain:.6f}dB", "-ar", "44100", "-ac", "2",
                 "-c:a", "libvorbis", "-q:a", "6", str(path)])
            probe = json.loads(run(["ffprobe", "-v", "error", "-show_streams", "-show_format",
                                    "-of", "json", str(path)]).stdout)
            stream = probe["streams"][0]
            duration = float(probe["format"]["duration"])
            if stream["sample_rate"] != "44100" or stream["channels"] != 2 or abs(duration - cue["seconds"]) > .02:
                raise ValueError(f"format/duration mismatch: {path}")
            final = measure(path)
            if not math.isfinite(final["input_i"]) or final["input_tp"] > -1.7:
                raise ValueError(f"silent/clipping output: {path}: {final}")
            files.append({"path": str(path.relative_to(out)), "sha256": sha(path),
                          "midi_sha256": sha(midi), "duration": duration, **final})
            wav.unlink()
        receipts.append({**cue, "samples": samples, "common_gain_db": gain, "files": files})
        print(f"{name}: {cue['bpm']} BPM, {cue['seconds']:.2f} 秒，主混音與四聲部完成", flush=True)
    # Short, level-matched listening sample: theme then combat, with a clear gap.
    run(["ffmpeg", "-y", "-loglevel", "error", "-i", str(out / "theme.ogg"),
         "-i", str(out / "combat.ogg"), "-filter_complex",
         "[0:a]atrim=0:20,asetpts=PTS-STARTPTS,afade=t=out:st=19.5:d=0.5[a];"
         "[1:a]atrim=0:20,asetpts=PTS-STARTPTS,afade=t=in:d=0.1,afade=t=out:st=19.5:d=0.5[b];"
         "[a][b]concat=n=2:v=0:a=1[out]", "-map", "[out]", "-ar", "44100",
         "-c:a", "libvorbis", "-q:a", "4", str(out / "theme-combat-listen.ogg")])
    metadata = {"schema": 1, "status": "approved-direction-technical-checks-passed",
                "profile": "modern", "arrangement": "uplifting-20261029",
                "image": "wasteland-media:bookworm-modern-r1",
                "label": "new remake arrangement", "date_local": "2026-10-29",
                "composer": "project-authored MIDI; no external music-generation service",
                "source_sha256": {p: sha(root / p) for p in
                                  ("tools/make_music.py", "tools/make_uplifting_music.py", "tools/render_modern_music.py")},
                "soundfont_sha256": sha(sf),
                "packages": run(["dpkg-query", "-W", "-f=${Package} ${Version}\n",
                                 "fluidsynth", "fluid-soundfont-gm", "ffmpeg"]).stdout.splitlines(),
                "render": "three cycles; retain middle cycle for reverb continuity; common gain for all stems",
                "sample_rate": 44100, "channels": 2, "codec": "Ogg Vorbis q6",
                "listening": "User approved theme/combat direction on 2026-10-29; remaining cues use the approved arrangement direction; individual listening not claimed",
                "cues": receipts}
    (out / "metadata.json").write_text(json.dumps(metadata, ensure_ascii=False, indent=2) + "\n")
    if args.install_to:
        target = args.install_to.resolve()
        if target == out or target in out.parents or out in target.parents:
            raise ValueError("候選與安裝目錄不能互相包含")
        # Only complete masters, editable one-loop MIDI and rights/metadata are
        # installed. Stems and three-cycle render inputs remain in the source.
        target.parent.mkdir(parents=True, exist_ok=True)
        with tempfile.TemporaryDirectory(prefix=".modern-install-", dir=target.parent) as tmp:
            staged = Path(tmp) / "modern"
            staged.mkdir()
            delivered = json.loads(json.dumps(metadata))
            delivered["delivery"] = "ten masters and one-loop MIDI; stems retained with source render receipt"
            delivered["source_receipt_sha256"] = sha(out / "metadata.json")
            for cue in delivered["cues"]:
                cue["files"] = cue["files"][:1]
                midi = out / f"{cue['name']}.mid"
                cue["single_loop_midi_sha256"] = sha(midi)
                for suffix in ("ogg", "mid"):
                    shutil.copyfile(out / f"{cue['name']}.{suffix}", staged / f"{cue['name']}.{suffix}")
            shutil.copyfile(out / "FluidR3-GM-LICENSE.txt", staged / "FluidR3-GM-LICENSE.txt")
            (staged / "metadata.json").write_text(json.dumps(delivered, ensure_ascii=False, indent=2) + "\n")
            backup = target.with_name(target.name + "-before-uplift-20261029")
            if target.exists():
                if backup.exists():
                    raise FileExistsError(f"舊版備份已存在，未覆寫：{backup}")
                target.rename(backup)
            try:
                staged.rename(target)
            except Exception:
                if backup.exists() and not target.exists():
                    backup.rename(target)
                raise
        print(f"十首驗證通過，已安裝：{target}", flush=True)


if __name__ == "__main__":
    main()
