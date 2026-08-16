#!/usr/bin/env python3
"""產生重製版的背景音樂（標準 MIDI 檔）。

**原版沒有背景音樂。** DOS 版《Wasteland》只有九首 PC 喇叭音效，資料在
執行檔的 `seg005`（864 bytes，`docs/re/44`），其中只有音效 4 是旋律，1.8 秒。
所以 BGM 是重製版自己加的東西，曲子由這支腳本產生——**不是抽自原版**，
也不假裝是原版的音樂（`rulebook/93` 鐵則 1 講的「不可自產」管的是
「別拿自寫合成器冒充原版音色」；這裡沒有原版可冒充，作品是新的）。

目標音源是 **Roland MT-32**：1988 年 DOS 遊戲的高階音源，年代與機種都對得上。
所以聲部安排照 MT-32 的規矩，不是 General MIDI：

  - **MIDI channel 1 不用**。MT-32 的八個聲部預設吃 channel 2–9，
    節奏在 channel 10。放在 channel 1 會整條軌沒聲音，而且是安靜地沒聲音。
  - 音色編號是 **MT-32 的內建音色表**，與 GM 完全不同
    （例如 48 ＝ Str Sect 1、92 ＝ Fr Horn 1、37 ＝ Atmosphere）。

用法（純 stdlib）：

    python3 tools/make_music.py workplace/music
    tools/render_music.sh          # MIDI → MT-32 → ogg
"""

import os
import struct
import sys

TPQ = 480  # 每個四分音符的 tick 數

# MT-32 內建音色（不是 GM）。
MT32 = {
    "strings": 48,   # Str Sect 1
    "horn": 92,      # Fr Horn 1
    "bass": 64,      # Acou Bass1
    "atmos": 37,     # Atmosphere
    "soundtrack": 36,
    "synbass": 28,   # Syn Bass 1
    "synbrass": 24,  # Syn Brass 1
    "guitar": 59,    # Guitar 1
    "harp": 57,      # Harp 1
    "fantasy": 32,
    "timpani": 112,
}

# 節奏聲部的鍵位（MT-32 的節奏音色表，前幾個與 GM 一致）。
KICK, SNARE, HAT, CRASH = 36, 38, 42, 49

# D 小調（D E F G A Bb C）。整份用同一個調，四首曲子換得起來。
N = {"C": 0, "D": 2, "E": 4, "F": 5, "G": 7, "A": 9, "Bb": 10, "B": 11}


def note(name, octave):
    """音名 ＋ 八度 → MIDI 音高（中央 C ＝ C4 ＝ 60）。"""
    return N[name] + (octave + 1) * 12


def vlq(n):
    """MIDI 的變長數量。"""
    out = bytearray([n & 0x7F])
    n >>= 7
    while n:
        out.insert(0, (n & 0x7F) | 0x80)
        n >>= 7
    return bytes(out)


class Track:
    """一條 MIDI 軌。事件先收成 (絕對 tick, 資料)，寫檔時再排序算差值。"""

    def __init__(self, name, channel):
        self.name = name
        self.ch = channel
        self.events = []

    def meta(self, at, kind, payload):
        self.events.append((at, 0, bytes([0xFF, kind]) + vlq(len(payload)) + payload))

    def program(self, at, prog):
        self.events.append((at, 1, bytes([0xC0 | self.ch, prog])))

    def volume(self, at, level):
        self.events.append((at, 1, bytes([0xB0 | self.ch, 7, level])))

    def note(self, at, pitch, dur, vel=90):
        # ⚠ note-off 要排在同一 tick 的 note-on **前面**，否則同音接續會被
        # 後到的 off 砍掉——症狀是連續同音只響第一個。用 order 欄排序解決。
        self.events.append((at, 2, bytes([0x90 | self.ch, pitch, vel])))
        self.events.append((at + dur, 0, bytes([0x80 | self.ch, pitch, 0])))

    def chord(self, at, pitches, dur, vel=80):
        for p in pitches:
            self.note(at, p, dur, vel)

    def encode(self):
        self.events.sort(key=lambda e: (e[0], e[1]))
        out = bytearray()
        last = 0
        for at, _, data in self.events:
            out += vlq(at - last) + data
            last = at
        out += vlq(0) + b"\xff\x2f\x00"  # end of track
        return b"MTrk" + struct.pack(">I", len(out)) + bytes(out)


def write_midi(path, tracks, bpm):
    head = struct.pack(">HHH", 1, len(tracks) + 1, TPQ)
    tempo = Track("tempo", 0)
    tempo.meta(0, 0x51, struct.pack(">I", int(60_000_000 / bpm))[1:])
    body = b"MThd" + struct.pack(">I", 6) + head + tempo.encode()
    for t in tracks:
        body += t.encode()
    with open(path, "wb") as f:
        f.write(body)
    return len(body)


# ── 曲子 ────────────────────────────────────────────────────────────────
# 和聲進行寫成 (根音名, 八度, [和弦內音])，一小節一個。

W = TPQ * 4  # 全音符 ＝ 一小節
H = TPQ * 2
Q = TPQ
E = TPQ // 2


def theme():
    """主題曲：緩、莊嚴，弦樂鋪底 ＋ 法國號旋律。標題畫面與遊俠中心用。"""
    pad = Track("pad", 1)
    lead = Track("lead", 2)
    bass = Track("bass", 3)
    drum = Track("drum", 9)
    pad.program(0, MT32["strings"])
    lead.program(0, MT32["horn"])
    bass.program(0, MT32["bass"])
    pad.volume(0, 80)
    lead.volume(0, 100)
    bass.volume(0, 85)

    # i – VI – III – VII – i – VI – iv – V（D 小調）
    prog = [
        ("D", ["D", "F", "A"]), ("Bb", ["Bb", "D", "F"]),
        ("F", ["F", "A", "C"]), ("C", ["C", "E", "G"]),
        ("D", ["D", "F", "A"]), ("Bb", ["Bb", "D", "F"]),
        ("G", ["G", "Bb", "D"]), ("A", ["A", "C", "E"]),
    ]
    melody = [
        [("D", 4, H), ("F", 4, Q), ("A", 4, Q)],
        [("G", 4, H), ("F", 4, H)],
        [("E", 4, Q), ("F", 4, Q), ("G", 4, H)],
        [("A", 4, W)],
        [("D", 5, H), ("C", 5, Q), ("A", 4, Q)],
        [("Bb", 4, H), ("A", 4, H)],
        [("G", 4, Q), ("A", 4, Q), ("Bb", 4, H)],
        [("A", 4, W)],
    ]
    for rep in range(2):
        base = rep * 8 * W
        for i, (root, tones) in enumerate(prog):
            at = base + i * W
            pad.chord(at, [note(t, 3) for t in tones], W, 70)
            bass.note(at, note(root, 2), H, 85)
            bass.note(at + H, note(root, 2), H, 70)
            # 第二輪加定音鼓的重拍，把段落推起來。
            if rep == 1:
                drum.note(at, KICK, Q, 95)
                if i % 2 == 1:
                    drum.note(at + H, SNARE, Q, 70)
        for i, bar in enumerate(melody):
            at = base + i * W
            # 第二輪旋律高八度，音量再上去一點。
            oct_shift = 12 if rep == 1 else 0
            vel = 100 if rep == 1 else 88
            for name, octv, dur in bar:
                lead.note(at, note(name, octv) + oct_shift, dur - 20, vel)
                at += dur
    return [pad, lead, bass, drum], 72


def desert():
    """沙漠：空、慢、稀疏。長音鋪底加零星的動機，走很久不膩。"""
    pad = Track("pad", 1)
    voice = Track("voice", 2)
    bass = Track("bass", 3)
    pad.program(0, MT32["atmos"])
    voice.program(0, MT32["soundtrack"])
    bass.program(0, MT32["synbass"])
    pad.volume(0, 70)
    voice.volume(0, 78)
    bass.volume(0, 72)

    # 四個和聲區，每區四小節。持續低音 D 貫穿，像地平線。
    areas = [["D", "A"], ["Bb", "F"], ["F", "C"], ["G", "D"]]
    for i, tones in enumerate(areas):
        at = i * 4 * W
        pad.chord(at, [note(t, 3) for t in tones], 4 * W - 40, 62)
        bass.note(at, note("D", 1), 4 * W - 40, 70)
        # 動機：五個音，每一區換一個起點，像遠處傳來的聲音。
        motif = [("D", 5), ("F", 5), ("E", 5), ("D", 5), ("A", 4)]
        start = at + 2 * W
        for j, (name, octv) in enumerate(motif):
            voice.note(start + j * Q, note(name, octv), Q + Q // 2, 66 - j * 4)
    return [pad, voice, bass], 60


def combat():
    """戰鬥：低音頑固音型 ＋ 銅管刺點 ＋ 鼓。短、循環得起來。"""
    bass = Track("bass", 1)
    brass = Track("brass", 2)
    drum = Track("drum", 9)
    bass.program(0, MT32["synbass"])
    brass.program(0, MT32["synbrass"])
    bass.volume(0, 95)
    brass.volume(0, 92)

    pattern = ["D", "D", "D", "F", "D", "D", "C", "D"]  # 八分音符的頑固音型
    for bar in range(8):
        at = bar * W
        for j, name in enumerate(pattern):
            bass.note(at + j * E, note(name, 2), E - 20, 100)
        drum.note(at, KICK, E, 105)
        drum.note(at + Q, SNARE, E, 92)
        drum.note(at + H, KICK, E, 100)
        drum.note(at + H + Q, SNARE, E, 92)
        for j in range(8):
            drum.note(at + j * E, HAT, E // 2, 55)
        # 每兩小節一次銅管刺點；第 8 小節收在屬和弦，接得回開頭。
        if bar % 2 == 1:
            stab = ["D", "F", "A"] if bar != 7 else ["A", "C", "E"]
            brass.chord(at + H, [note(t, 4) for t in stab], Q, 100)
        if bar == 0:
            drum.note(at, CRASH, Q, 100)
    return [bass, brass, drum], 132


def town():
    """城鎮：稍微溫暖，吉他分解和弦 ＋ 豎琴。商店、醫生、酒吧用。"""
    gtr = Track("gtr", 1)
    harp = Track("harp", 2)
    bass = Track("bass", 3)
    gtr.program(0, MT32["guitar"])
    harp.program(0, MT32["harp"])
    bass.program(0, MT32["bass"])
    gtr.volume(0, 88)
    harp.volume(0, 74)
    bass.volume(0, 80)

    prog = [
        ("D", ["D", "F", "A"]), ("A", ["A", "C", "E"]),
        ("Bb", ["Bb", "D", "F"]), ("F", ["F", "A", "C"]),
        ("G", ["G", "Bb", "D"]), ("D", ["D", "F", "A"]),
        ("A", ["A", "C", "E"]), ("D", ["D", "F", "A"]),
    ]
    for i, (root, tones) in enumerate(prog):
        at = i * W
        # 吉他：上下行分解，每小節八個八分音符。
        seq = tones + [tones[-1]] + list(reversed(tones))[1:]
        for j in range(8):
            name = seq[j % len(seq)]
            gtr.note(at + j * E, note(name, 4), E - 20, 78)
        bass.note(at, note(root, 2), H, 78)
        bass.note(at + H, note(root, 3), H, 66)
        if i % 2 == 0:
            harp.chord(at + Q, [note(t, 5) for t in tones], H, 60)
    return [gtr, harp, bass], 88


TUNES = {"theme": theme, "desert": desert, "combat": combat, "town": town}


def main():
    if len(sys.argv) != 2:
        print(__doc__)
        return 1
    out = sys.argv[1]
    os.makedirs(out, exist_ok=True)
    for name, fn in TUNES.items():
        tracks, bpm = fn()
        path = os.path.join(out, name + ".mid")
        size = write_midi(path, tracks, bpm)
        bars = max(e[0] for t in tracks for e in t.events) / (TPQ * 4)
        print(f"{name:8s} {bpm:3d} BPM  {bars:5.1f} 小節  {size:6d} bytes  → {path}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
