#!/usr/bin/env python3
"""2026-10-29 定案的現代配樂振奮版，自作 MIDI 十首。復古樂譜另保留。"""

import argparse
import json
from pathlib import Path

import make_music as m


def melody(track, start, bars, velocity):
    for i, phrase in enumerate(bars):
        at = start + i * m.W
        for name, octave, beats in phrase:
            duration = round(beats * m.Q)
            if name:
                track.note(at, m.note(name, octave), duration - 32, velocity)
            at += duration
        if at != start + (i + 1) * m.W:
            raise ValueError("melody bar must contain four beats")


THEME = [
    [("D", 4, 2), ("F", 4, 1), ("A", 4, 1)],
    [("G", 4, 3), ("F", 4, 1)],
    [("E", 4, 1), ("F", 4, 1), ("G", 4, 2)],
    [("A", 4, 3), (None, 0, 1)],
    [("D", 5, 2), ("C", 5, 1), ("A", 4, 1)],
    [("Bb", 4, 2), ("A", 4, 2)],
    [("G", 4, 1), ("A", 4, 1), ("Bb", 4, 2)],
    [("A", 4, 3), (None, 0, 1)],
]
ANSWER = [
    [("C", 5, 1.5), ("A", 4, .5), ("C", 5, 1), ("F", 5, 1)],
    [("E", 5, 2), ("D", 5, 1), ("C", 5, 1)],
    [("Bb", 4, 1), ("D", 5, 1), ("F", 5, 2)],
    [("C", 5, 2), ("A", 4, 1), (None, 0, 1)],
    [("Bb", 4, 1), ("C", 5, 1), ("D", 5, 2)],
    [("E", 5, 2), ("C#", 5, 1), ("E", 5, 1)],
    [("D", 5, 3), ("A", 4, 1)],
    [("C#", 5, 2), ("A", 4, 1), (None, 0, 1)],
]

A = [("D", ["D", "F", "A"]), ("Bb", ["Bb", "D", "F"]),
     ("F", ["F", "A", "C"]), ("C", ["C", "E", "G"]),
     ("D", ["D", "F", "A"]), ("Bb", ["Bb", "D", "F"]),
     ("G", ["G", "Bb", "D"]), ("A", ["A", "C#", "E"])]
B = [("F", ["F", "A", "C"]), ("C", ["C", "E", "G"]),
     ("Bb", ["Bb", "D", "F"]), ("F", ["F", "A", "C"]),
     ("G", ["G", "Bb", "D"]), ("A", ["A", "C#", "E"]),
     ("D", ["D", "F", "A"]), ("A", ["A", "C#", "E"])]


def score(combat=False):
    """保留專案 D–F–A 動機，重新寫節奏、回應句、低音與配器。"""
    horn = m.setup("horn melody", 1, "horn", 106, 60)
    trumpet = m.setup("trumpet answer", 2, "trumpet", 91, 72)
    strings = m.setup("moving strings", 3, "strings", 75, 38)
    guitar = m.setup("road guitar", 4, "elecgtr", 62, 90)
    bass = m.setup("driving bass", 5, "synbass" if combat else "bass", 88, 64)
    drums = m.Track("drum kit", 9)
    drums.volume(0, 83)
    chords = (A + B + A + B) if not combat else (A * 4)
    previous = None
    for bar, (root, tones) in enumerate(chords):
        at = bar * m.W
        section = bar // 8
        # Middle section leaves space; final section broadens the melody.
        open_section = section == 1 if not combat else section == 2
        previous = m.voicing(tones, m.note("D", 3), m.note("A", 4), previous)
        for j in range(8):
            pitch = previous[(0, 2, 1, 2, 0, 2, 1, 2)[j]]
            strings.note(at + j * m.E, pitch + 12, m.E - 80,
                         (65 if open_section else 76) + (5 if j % 4 == 0 else 0))
            if not open_section or j % 2 == 0:
                bass.note(at + j * m.E, m.note(root, 2), m.E - 40,
                          91 if j % 4 == 0 else 72)
        # Muted, syncopated guitar supplies an American-road texture.
        for beat in ((0, 1.5, 2.5) if open_section else (0, .75, 1.5, 2, 2.75, 3.5)):
            guitar.chord(at + round(beat * m.Q),
                         [m.note(root, 3), m.note(root, 3) + 7], m.E - 70,
                         65 if open_section else 77)
        for beat in (0, 2):
            drums.note(at + beat * m.Q, m.KICK, m.S, 94 if beat == 0 else 83)
        if not open_section or bar % 2:
            for beat in (1, 3):
                drums.note(at + beat * m.Q, m.SNARE, m.S, 87 if combat else 76)
        for j in range(8):
            drums.note(at + j * m.E, m.HAT, m.S, 48 if j % 2 else 58)
        if bar in (0, 16, 24):
            drums.note(at, m.CRASH, m.Q, 88)
        if bar % 8 == 7:
            for j, key in enumerate((m.TOM_LO, m.TOM_MID, m.TOM_HI, m.SNARE)):
                drums.note(at + 3 * m.Q + j * m.S, key, m.S, 65 + j * 6)

    if not combat:
        melody(horn, 0, THEME, 99)
        melody(trumpet, 8 * m.W, ANSWER, 91)
        melody(horn, 16 * m.W, THEME, 111)
        melody(trumpet, 24 * m.W, ANSWER, 105)
        # Broad brass answers, with rests between phrases.
        for bar in (18, 22, 26, 30):
            trumpet.chord(bar * m.W + 3 * m.Q,
                          [m.note("D", 4), m.note("A", 4)], m.Q - 50, 79)
    else:
        hook = [
            [("D", 5, .5), ("F", 5, .5), ("A", 5, 1), ("G", 5, .5), ("F", 5, .5), ("E", 5, 1)],
            [("D", 5, 2), ("F", 5, 1), (None, 0, 1)],
            [("C", 5, .5), ("E", 5, .5), ("G", 5, 1), ("A", 5, 1), ("G", 5, 1)],
            [("E", 5, 2), (None, 0, 2)],
            [("D", 5, 1), ("F", 5, 1), ("A", 5, 1), ("D", 5, 1)],
            [("F", 5, 1.5), ("D", 5, .5), ("Bb", 4, 1), (None, 0, 1)],
            [("G", 5, 1), ("F", 5, 1), ("D", 5, 1), ("G", 4, 1)],
            [("E", 5, 1), ("C#", 5, 1), ("A", 4, 1), (None, 0, 1)],
        ]
        melody(horn, 0, hook, 105)
        melody(trumpet, 8 * m.W, hook, 98)
        melody(horn, 16 * m.W, THEME, 94)
        melody(trumpet, 24 * m.W, hook, 110)
    for track in (horn, trumpet, strings):
        for bar in range(0, 32, 4):
            track.swell(bar * m.W, 4 * m.W - 1, 98, 117)
    return {"melody": [horn, trumpet], "harmony": [strings, guitar],
            "bass": [bass], "pulse": [drums]}, 144 if combat else 120, 32


def repeated(tracks, ticks, cycles=3):
    out = []
    for track in tracks:
        copy = m.Track(track.name, track.ch)
        for cycle in range(cycles):
            copy.events.extend((at + cycle * ticks, order, data)
                               for at, order, data in track.events)
        copy.meta(cycles * ticks, 0x06, b"preview-end")
        out.append(copy)
    return out


def arrange_scene(name):
    """原有八首的旋律留作線索，重寫律動、回應與聲部配置。"""
    tracks, _ = m.TUNES[name]()
    config = {
        "desert": (90, 24, [A[0]] * 4 + [A[6]] * 4 + [A[1]] * 4 +
                   [A[3]] * 4 + [A[2]] * 4 + [A[0]] * 4, 25, 22, .65),
        "night": (76, 16, [("A", ["A", "C", "E"])] * 4 +
                  [("F", ["F", "A", "C"])] * 4 + [("G", ["G", "B", "D"])] * 4 +
                  [("A", ["A", "C", "E"])] * 4, 25, 11, .35),
        "town": (112, 24, [("G", ["G", "Bb", "D"]), ("D", ["D", "F", "A"]),
                 ("Eb", ["Eb", "G", "Bb"]), ("Bb", ["Bb", "D", "F"]),
                 ("C", ["C", "Eb", "G"]), ("G", ["G", "Bb", "D"]),
                 ("D", ["D", "F", "A"]), ("G", ["G", "Bb", "D"])] * 3,
                 48, 40, .75),
        "facility": (116, 16, [("Bb", ["Bb", "D", "F"]), ("G", ["G", "Bb", "D"]),
                     ("Eb", ["Eb", "G", "Bb"]), ("F", ["F", "A", "C"]),
                     ("Bb", ["Bb", "D", "F"]), ("Eb", ["Eb", "G", "Bb"]),
                     ("C", ["C", "Eb", "G"]), ("F", ["F", "A", "C"])] * 2,
                     25, 11, .65),
        "sewer": (88, 20, [("C", ["C", "Eb", "G"])] * 4 +
                  [("Ab", ["Ab", "C", "Eb"])] * 4 + [("C", ["C", "Eb", "G"])] * 4 +
                  [("Bb", ["Bb", "D", "F"])] * 4 + [("G", ["G", "Bb", "D"])] * 4,
                  42, 60, .55),
        "vegas": (126, 24, [("F", ["F", "Ab", "C"]), ("Bb", ["Bb", "D", "F"]),
                  ("Eb", ["Eb", "G", "Bb"]), ("Ab", ["Ab", "C", "Eb"]),
                  ("Db", ["Db", "F", "Ab"]), ("G", ["G", "Bb", "Db"]),
                  ("C", ["C", "E", "G"]), ("F", ["F", "Ab", "C"])] * 3,
                  28, 57, .8),
        "base": (138, 32, [("E", ["E", "G", "B"])] * 2 + [("C", ["C", "E", "G"])] * 2 +
                 [("G", ["G", "B", "D"])] * 2 + [("A", ["A", "C", "E"]),
                 ("B", ["B", "D", "F#"])], 48, 60, 1.0),
        "ending": (104, 24, A[:2] + [A[6], A[7]] + A[:2] + [A[6], A[7]] +
                   [A[1], A[2], A[3], ("G", ["G", "B", "D"]),
                    ("D", ["D", "F#", "A"]), A[7], ("G", ["G", "B", "D"]),
                    ("D", ["D", "F#", "A"])] +
                   [("D", ["D", "F#", "A"]), A[7], A[1], ("G", ["G", "B", "D"]),
                    ("D", ["D", "F#", "A"]), ("G", ["G", "B", "D"]), A[7],
                    ("D", ["D", "F#", "A"])], 48, 60, .85),
    }
    bpm, bars, progression, motion_program, answer_program, intensity = config[name]
    if name == "base":
        progression *= 4
    if len(progression) != bars:
        raise ValueError(f"bad progression length: {name}")
    groups = {"melody": [], "harmony": [], "bass": [], "pulse": []}
    for track in tracks:
        category = ("pulse" if track.ch == 9 else "bass" if track.name == "low" else
                    "harmony" if track.name in ("pad", "gtr", "ep") else "melody")
        # An early melody gives the player something to follow immediately.
        if name in ("town", "vegas") and track.name in ("flute", "sax"):
            track.events += [(at - 8 * m.W, order, data)
                             for at, order, data in list(track.events)
                             if 8 * m.W <= at < 16 * m.W and data[0] & 0xF0 in (0x80, 0x90)]
        if name == "ending" and track.name == "lead":
            changed = []
            for at, order, data in track.events:
                if data[0] & 0xF0 in (0x80, 0x90):
                    pitch = data[1] - 12
                    # Original self-authored bar 19 had F# against a Bb chord.
                    if 18 * m.W <= at < 19 * m.W and pitch == m.note("F#", 4):
                        pitch -= 1
                    data = bytes([data[0], pitch, *data[2:]])
                changed.append((at, order, data))
            track.events = changed
        # Replace distant FX pads with a clear small-ensemble sound.
        program = {"pad": 48, "voice": 22 if name == "desert" else 11,
                   "bell": 4 if name == "night" else 14, "flute": 40,
                   "ep": 0 if name == "vegas" else 4, "brass": 57,
                   "lead": 60}.get(track.name)
        if program is not None and track.ch != 9:
            track.events = [(at, order, bytes([data[0], program]) if data[0] & 0xF0 == 0xC0 else data)
                            for at, order, data in track.events]
        groups[category].append(track)

    motion = m.setup("forward motion", 5, "guitar", 72 if intensity < .8 else 80, 42)
    motion.program(0, motion_program)
    answer = m.setup("hopeful answer", 6, "horn", 76 if intensity < .8 else 89, 78)
    answer.program(0, answer_program)
    pulse = m.Track("new pulse", 9)
    # Share the source percussion channel, with one intentional mix level.
    pulse.volume(0, 71 if intensity < .8 else 83)
    previous = None
    for bar, (root, tones) in enumerate(progression):
        at = bar * m.W
        previous = m.voicing(tones, m.note("C", 3), m.note("C", 5), previous)
        for j in range(8):
            if name == "night" and j % 2:
                continue
            pitch = previous[(0, 2, 1, 2, 0, 2, 1, 2)[j]]
            motion.note(at + j * m.E, pitch, m.E - 55,
                        int((67 if j % 4 == 0 else 57) * (.8 + intensity * .25)))
        # Brass/chime response uses this bar's chord, so every new phrase fits.
        if name in ("desert", "night", "sewer", "facility", "base") and bar % 4 in (0, 1):
            chord = [m.note(t, 4) for t in tones]
            if bar % 4 == 0:
                for beat, pitch in ((0, chord[0]), (1, chord[1]), (2, chord[2])):
                    answer.note(at + beat * m.Q, pitch, (m.H if beat == 2 else m.Q) - 35,
                                69 + round(intensity * 17))
            else:
                answer.note(at + m.Q, chord[2], m.H - 35, 66 + round(intensity * 15))
        if name == "vegas" and bar % 4 == 0:
            answer.chord(at + 3 * m.Q, previous[-2:], m.Q - 40, 79)
        if name == "ending" and bar < 2:
            for j, pitch in enumerate([m.note("D", 4), m.note("F", 4), m.note("A", 4)]):
                answer.note(at + j * m.Q, pitch, m.Q - 40, 79 + bar * 5)
        if name not in ("base", "vegas"):
            for beat in ((0,) if name == "night" else (0, 2)):
                pulse.note(at + beat * m.Q, m.KICK, m.S, int(74 * intensity))
            if name != "night":
                for beat in (1, 3):
                    pulse.note(at + beat * m.Q, m.SNARE, m.S, int(67 * intensity))
            for j in range(8):
                if name == "night" and j % 2:
                    continue
                pulse.note(at + j * m.E, m.HAT, m.S, int((42 if j % 2 else 52) * intensity))
        if name in ("town", "ending", "base") and bar in (0, bars // 2):
            pulse.note(at, m.CRASH, m.Q, int(77 * intensity))
    groups["melody"].append(answer)
    groups["harmony"].append(motion)
    groups["pulse"].append(pulse)
    return groups, bpm, bars


def main():
    ap = argparse.ArgumentParser(description=__doc__)
    ap.add_argument("out", type=Path)
    args = ap.parse_args()
    args.out.mkdir(parents=True, exist_ok=True)
    m.configure_profile("modern")
    cues = []
    for name in m.TUNES:
        stems, bpm, bars = (score(combat=name == "combat") if name in ("theme", "combat")
                            else arrange_scene(name))
        ticks = bars * m.W
        all_tracks = [t for group in stems.values() for t in group]
        m.write_midi(args.out / f"{name}.mid", repeated(all_tracks, ticks, cycles=1), bpm)
        for stem, tracks in {"master": all_tracks, **stems}.items():
            m.write_midi(args.out / f"{name}-{stem}.mid", repeated(tracks, ticks), bpm)
        cues.append({"name": name, "bpm": bpm, "bars": bars,
                     "seconds": bars * 240 / bpm, "stems": list(stems)})
    (args.out / "score.json").write_text(json.dumps({
        "status": "approved-direction-render-candidate", "cues": cues,
        "identity": "1988 American post-nuclear road, rusted military metal, damaged neon",
        "motif": "project-authored D-F-A, G-F, E-F-G, A; new response phrases",
        "source": "tools/make_music.py and tools/make_uplifting_music.py",
        "randomness": "none", "render_cycles": 3,
    }, ensure_ascii=False, indent=2) + "\n")


if __name__ == "__main__":
    main()
