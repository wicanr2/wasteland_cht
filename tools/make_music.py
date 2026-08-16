#!/usr/bin/env python3
"""產生重製版的背景音樂（標準 MIDI 檔），十首。

**原版沒有背景音樂。** DOS 版《Wasteland》只有九首 PC 喇叭音效，資料在
執行檔的 `seg005`（864 bytes，`docs/re/44`），其中只有音效 4 是旋律，1.8 秒。
所以 BGM 是重製版自己加的東西，曲子由這支腳本產生——**不是抽自原版**，
也不假裝是原版的音樂（`rulebook/93` 鐵則 1 講的「不可自產」管的是
「別拿自寫合成器冒充原版音色」；這裡沒有原版可冒充，作品是新的）。

目標音源是 **Roland MT-32／CM-32L**：1988 年 DOS 遊戲的高階音源，年代與機種都對得上。
所以聲部安排照 MT-32 的規矩，不是 General MIDI：

  - **MIDI channel 1 不用**。MT-32 的八個聲部預設吃 channel 2–9，
    節奏在 channel 10。放在 channel 1 會整條軌沒聲音，而且是安靜地沒聲音。
  - 音色編號是 **MT-32 的內建音色表**，與 GM 完全不同。程式變更送的是
    「音色編號 − 1」（音色表 1 起算、MIDI 程式碼 0 起算）。
  - **節奏鍵位不是 GM 的。** 這裡用的四個鍵是量出來的，不是照慣例挑的
    （`tools/mt32_probe.py`）。GM 的腳踏鈸 42 在 MT-32 上根本沒有指派——
    送下去不出聲，而檔案長度、音量、格式全部正常。
  - **同時發聲的音要省。** MT-32 全機只有 32 個 partial，一個音色吃 1–4 個；
    四部各壓一個四音和弦就會開始掉音，而掉的是後來的那些音。
    所以鋪底一律 2–3 個音，主旋律單音。

用法（純 stdlib）：

    python3 tools/make_music.py workplace/music
    tools/render_music.sh          # MIDI → MT-32 → ogg
"""

import os
import struct
import sys

TPQ = 480  # 每個四分音符的 tick 數

# MT-32 內建音色（送出去的程式碼 ＝ 音色表編號 − 1）。
MT32 = {
    "strings": 48,    # 49 Str Sect 1
    "strings2": 49,   # 50 Str Sect 2
    "cello": 54,      # 55 Cello 1
    "contrabass": 56,  # 57 Contrabass
    "horn": 92,       # 93 Fr Horn 1
    "trombone": 90,   # 91 Trombone 1
    "trumpet": 88,    # 89 Trumpet 1
    "brass": 95,      # 96 Brs Sect 1
    "bass": 64,       # 65 Acou Bass1
    "fretless": 70,   # 71 Fretless1
    "synbass": 28,    # 29 Syn Bass 1
    "synbass3": 30,   # 31 Syn Bass 3
    "synbrass": 24,   # 25 Syn Brass 1
    "atmos": 37,      # 38 Atmosphere
    "soundtrack": 36,  # 37 Soundtrack
    "fantasy": 32,    # 33 Fantasy
    "chorale": 34,    # 35 Chorale
    "warmbell": 38,   # 39 Warm Bell
    "icerain": 41,    # 42 Ice Rain
    "glasses": 35,    # 36 Glasses
    "harp": 57,       # 58 Harp 1
    "guitar": 59,     # 60 Guitar 1
    "elecgtr": 61,    # 62 Elec Gtr 1
    "epiano": 3,      # 4 ElecPiano1
    "organ": 9,       # 10 Elec Org 2
    "vibe": 97,       # 98 Vibe 1
    "tubebell": 102,  # 103 Tube Bell
    "flute": 72,      # 73 Flute 1
    "sax": 78,        # 79 Sax 1
    "oboe": 84,       # 85 Oboe
    "timpani": 112,   # 113 Timpani
    "orchehit": 122,  # 123 Orche Hit
    "square": 47,     # 48 SquareWave
}

# 節奏鍵位：**量出來的**，不是 GM 的慣例（見檔頭與 tools/mt32_probe.py）。
#
#   鍵 30  過零率 313／RMS 1136  最低最響 → 大鼓
#   鍵 27  過零率 2160／尾 0.08  中頻、收得快 → 小鼓
#   鍵 43  過零率 12987／尾 0.02 最高、收得最快 → 腳踏鈸
#   鍵 75  過零率 15783／尾 1.35 最高、拖最長 → 碎音鈸
#   鍵 53／52／31  過零率 483／690／743 → 低中高三個中鼓
KICK, SNARE, HAT, CRASH = 30, 27, 43, 75
TOM_LO, TOM_MID, TOM_HI = 53, 52, 31
RIDE = 61   # 過零 5340／尾 0.55：比腳踏鈸暗、比碎音鈸短

# 量出來**沒有指派**的鍵（CM-32L，`docs/mt32-rhythm-probe.md`）。
# 用到這裡面任何一個都會安靜地沒聲音，所以開跑前擋一次——
# 這正是上一版踩過的坑：GM 的腳踏鈸 42 在這台機器上不存在。
MT32_SILENT_KEYS = {42, 44, 46, 47, 48, 59, 79, 80, 81, 82}


def _check_rhythm_keys():
    used = {"大鼓": KICK, "小鼓": SNARE, "腳踏鈸": HAT, "碎音鈸": CRASH,
            "低中鼓": TOM_LO, "中中鼓": TOM_MID, "高中鼓": TOM_HI, "疊音鈸": RIDE}
    bad = {k: v for k, v in used.items() if v in MT32_SILENT_KEYS}
    if bad:
        raise SystemExit(
            "這些鍵在 MT-32 上沒有指派，會安靜地沒聲音："
            + "、".join(f"{k}={v}" for k, v in bad.items())
            + "\n重跑 tools/mt32_probe.py 挑別的鍵。")


_check_rhythm_keys()

# 同音異名兩種寫法都收：和聲上該寫 Db 的地方硬寫成 C# 會讓譜難讀，
# 而讀不懂的譜改起來就會出錯。
N = {"C": 0, "C#": 1, "Db": 1, "D": 2, "D#": 3, "Eb": 3, "E": 4, "Fb": 4,
     "F": 5, "F#": 6, "Gb": 6, "G": 7, "G#": 8, "Ab": 8, "A": 9,
     "A#": 10, "Bb": 10, "B": 11, "Cb": 11}


def note(name, octave):
    """音名 ＋ 八度 → MIDI 音高（中央 C ＝ C4 ＝ 60）。"""
    return N[name] + (octave + 1) * 12


def voicing(tones, low, high, prev=None):
    """把和弦內音排進 [low, high]，而且盡量貼著上一個聲位走。

    平行移動的原位三和弦聽起來像練習曲；讓每個聲部走最近的音（聲部導進）
    是最省力就能讓和聲聽起來像音樂的一步。
    """
    cands = sorted({note(t, o) for t in tones for o in range(0, 9)
                    if low <= note(t, o) <= high})
    if not cands:
        return []
    if prev is None:
        # 第一個和弦：從低往上取，音跟音之間拉開一點。
        out, last = [], None
        for c in cands:
            if last is None or c - last >= 3:
                out.append(c)
                last = c
            if len(out) == len(tones):
                break
        return out
    out, pool = [], list(cands)
    for p in prev:
        best = min(pool, key=lambda c: (abs(c - p), c))
        out.append(best)
        pool = [c for c in pool if c != best] or list(cands)
    return sorted(out)


def vlq(n):
    """MIDI 的變長數量。"""
    out = bytearray([n & 0x7F])
    n >>= 7
    while n:
        out.insert(0, (n & 0x7F) | 0x80)
        n >>= 7
    return bytes(out)


class Track:
    """一條 MIDI 軌。事件先收成 (絕對 tick, 排序權, 資料)，寫檔時再算差值。"""

    def __init__(self, name, channel):
        self.name = name
        self.ch = channel
        self.events = []

    def meta(self, at, kind, payload):
        self.events.append((at, 0, bytes([0xFF, kind]) + vlq(len(payload)) + payload))

    def cc(self, at, num, val):
        self.events.append((at, 1, bytes([0xB0 | self.ch, num, max(0, min(127, val))])))

    def program(self, at, prog):
        self.events.append((at, 1, bytes([0xC0 | self.ch, prog])))

    def volume(self, at, level):
        self.cc(at, 7, level)

    def pan(self, at, val):
        """CC10：0 ＝ 左、64 ＝ 中、127 ＝ 右。分開擺讓四個聲部不擠成一團。"""
        self.cc(at, 10, val)

    def expression(self, at, level):
        """CC11：樂句起伏用這個，不要動 CC7（那是整條軌的基準音量）。"""
        self.cc(at, 11, level)

    def swell(self, at, dur, lo, hi, steps=16):
        """一段漸強／漸弱。靜態音量是這種曲子聽起來最像機器的地方。"""
        for i in range(steps + 1):
            self.expression(at + dur * i // steps, lo + (hi - lo) * i // steps)

    def note(self, at, pitch, dur, vel=90):
        # ⚠ note-off 要排在同一 tick 的 note-on **前面**，否則同音接續會被
        # 後到的 off 砍掉——症狀是連續同音只響第一個。用排序權解決。
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


def setup(name, ch, timbre, vol, pan=64):
    t = Track(name, ch)
    t.program(0, MT32[timbre])
    t.volume(0, vol)
    t.pan(0, pan)
    t.expression(0, 127)
    return t


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


# ── 節拍常數 ────────────────────────────────────────────────────────────
W = TPQ * 4   # 一小節（四四拍）
H = TPQ * 2
Q = TPQ
E = TPQ // 2
S = TPQ // 4


def drums(track, bar, pattern, vel=1.0):
    """把一個字串節奏型鋪進一小節。每個字元 ＝ 一個十六分音符。

        k 大鼓  s 小鼓  h 腳踏鈸  c 碎音鈸  r 疊音鈸  . 休止
    """
    keys = {"k": (KICK, 104), "s": (SNARE, 96), "h": (HAT, 58),
            "c": (CRASH, 100), "r": (RIDE, 66)}
    for i, c in enumerate(pattern):
        if c not in keys:
            continue
        key, v = keys[c]
        track.note(bar + i * S, key, S, max(1, min(127, int(v * vel))))


# ═══ 曲子 ═══════════════════════════════════════════════════════════════
#
# 每首都刻意用不同的調與速度。整套用同一個調是上一版最明顯的問題——
# 從沙漠走進賭場再走進基地，聽起來會像同一首歌沒停過。


def theme():
    """主題曲：D 小調、莊嚴。標題畫面用。

    兩段：A 段弦樂鋪底加法國號的主題，B 段轉到相對大調（F）讓它有個出口，
    再收回 A。24 小節，約 76 秒。
    """
    pad = setup("pad", 1, "strings", 82, 54)
    lead = setup("lead", 2, "horn", 102, 74)
    low = setup("low", 3, "cello", 88, 64)
    perc = Track("perc", 9)

    A = [("D", ["D", "F", "A"]), ("Bb", ["Bb", "D", "F"]),
         ("F", ["F", "A", "C"]), ("C", ["C", "E", "G"]),
         ("D", ["D", "F", "A"]), ("Bb", ["Bb", "D", "F"]),
         ("G", ["G", "Bb", "D"]), ("A", ["A", "C#", "E"])]
    B = [("F", ["F", "A", "C"]), ("C", ["C", "E", "G"]),
         ("Bb", ["Bb", "D", "F"]), ("F", ["F", "A", "C"]),
         ("G", ["G", "Bb", "D"]), ("A", ["A", "C#", "E"]),
         ("D", ["D", "F", "A"]), ("A", ["A", "C#", "E"])]

    melA = [[("D", 4, H), ("F", 4, Q), ("A", 4, Q)],
            [("G", 4, H + Q), ("F", 4, Q)],
            [("E", 4, Q), ("F", 4, Q), ("G", 4, H)],
            [("A", 4, W)],
            [("D", 5, H), ("C", 5, Q), ("A", 4, Q)],
            [("Bb", 4, H), ("A", 4, H)],
            [("G", 4, Q), ("A", 4, Q), ("Bb", 4, H)],
            [("A", 4, W)]]
    melB = [[("C", 5, H), ("A", 4, H)],
            [("G", 4, W)],
            [("Bb", 4, H), ("D", 5, H)],
            [("C", 5, H + Q), ("A", 4, Q)],
            [("Bb", 4, Q), ("C", 5, Q), ("D", 5, H)],
            [("E", 5, H), ("C#", 5, H)],
            [("D", 5, W)],
            [("A", 4, H), ("A", 4, H)]]

    prev = None
    for si, (prog, mel, oct_shift, base_vel) in enumerate(
            [(A, melA, 0, 84), (B, melB, 0, 94), (A, melA, 12, 100)]):
        base = si * 8 * W
        for i, (root, tones) in enumerate(prog):
            at = base + i * W
            prev = voicing(tones, note("D", 3), note("A", 4), prev)
            pad.chord(at, prev, W - 30, 66 + si * 4)
            low.note(at, note(root, 2), H, 82)
            low.note(at + H, note(root, 2), H - 20, 66)
            if si >= 1:
                perc.note(at, KICK, Q, 92)
                if i % 2 == 1:
                    perc.note(at + H, SNARE, Q, 66)
            if si == 2 and i == 0:
                perc.note(at, CRASH, Q, 96)
        # 樂句的漸強：每四小節一個弧線，靜態音量最像機器。
        for p in range(2):
            lead.swell(base + p * 4 * W, 4 * W, 86, 118 if p else 104)
        for i, bar in enumerate(mel):
            at = base + i * W
            for name, octv, dur in bar:
                lead.note(at, note(name, octv) + oct_shift, dur - 25, base_vel)
                at += dur
    return [pad, lead, low, perc], 76


def desert():
    """沙漠（白天）：空、慢、稀疏。長音鋪底加零星的動機，走很久不膩。

    D 多利安（小調但六級升高），比純小調少一點悲情、多一點乾。
    24 小節、56 BPM，約 103 秒。
    """
    pad = setup("pad", 1, "atmos", 72, 44)
    voice = setup("voice", 2, "soundtrack", 80, 84)
    low = setup("low", 3, "contrabass", 76, 64)
    bell = setup("bell", 4, "warmbell", 60, 100)

    areas = [["D", "A"], ["D", "G"], ["Bb", "F"], ["C", "G"],
             ["F", "C"], ["D", "A"]]
    motifs = [[("D", 5), ("F", 5), ("E", 5), ("D", 5), ("A", 4)],
              [("A", 4), ("C", 5), ("D", 5)],
              [("F", 5), ("E", 5), ("C", 5), ("D", 5)],
              [("G", 4), ("A", 4), ("C", 5)],
              [("C", 5), ("D", 5), ("F", 5), ("E", 5)],
              [("D", 5), ("A", 4), ("D", 4)]]

    prev = None
    for i, tones in enumerate(areas):
        at = i * 4 * W
        prev = voicing(tones, note("D", 3), note("D", 5), prev)
        pad.chord(at, prev, 4 * W - 60, 60)
        pad.swell(at, 4 * W, 78, 108 if i % 2 else 92)
        # 持續低音 D 貫穿，像地平線；只有第三、五區換一次根音。
        root = "Bb" if i == 2 else ("F" if i == 4 else "D")
        low.note(at, note(root, 1), 4 * W - 60, 68)
        m = motifs[i]
        start = at + 2 * W
        for j, (name, octv) in enumerate(m):
            voice.note(start + j * Q, note(name, octv), Q + E, 70 - j * 5)
        # 遠處的鐘：每兩區一次，落在小節線之前半拍，不要對齊。
        if i % 2 == 1:
            bell.note(at + 3 * W + E * 3, note("D", 6), H, 52)
    return [pad, voice, low, bell], 56


def night():
    """沙漠（夜間）：更暗、更慢、幾乎沒有旋律。

    A 小調，兩個聲部加一個偶爾出現的鐘。16 小節、50 BPM，約 77 秒。
    夜間換曲是重製版的決定，門檻取原版的晝夜門檻（6 時與 18 時，`docs/re/27`）。
    """
    pad = setup("pad", 1, "chorale", 66, 40)
    low = setup("low", 2, "contrabass", 72, 64)
    bell = setup("bell", 3, "icerain", 56, 96)

    areas = [["A", "E"], ["F", "C"], ["G", "D"], ["A", "E"]]
    prev = None
    for i, tones in enumerate(areas):
        at = i * 4 * W
        prev = voicing(tones, note("A", 2), note("A", 4), prev)
        pad.chord(at, prev, 4 * W - 80, 54)
        pad.swell(at, 4 * W, 64, 96)
        low.note(at, note(tones[0], 1), 4 * W - 80, 62)
        for j, p in enumerate([("A", 5), ("C", 6), ("E", 5)]):
            bell.note(at + (1 + j) * W + Q, note(*p), H, 46 - j * 6)
    return [pad, low, bell], 50


def town():
    """城鎮（一般 32 × 32 的室內地圖）：溫暖一點，吉他分解和弦 ＋ 豎琴。

    G 小調。24 小節、92 BPM，約 63 秒。
    """
    gtr = setup("gtr", 1, "guitar", 90, 48)
    harp = setup("harp", 2, "harp", 74, 88)
    low = setup("low", 3, "bass", 82, 64)
    flute = setup("flute", 4, "flute", 78, 76)

    prog = [("G", ["G", "Bb", "D"]), ("D", ["D", "F", "A"]),
            ("Eb", ["Eb", "G", "Bb"]), ("Bb", ["Bb", "D", "F"]),
            ("C", ["C", "Eb", "G"]), ("G", ["G", "Bb", "D"]),
            ("D", ["D", "F", "A"]), ("G", ["G", "Bb", "D"])]
    tune = [None, None,
            [("D", 5, H), ("Eb", 5, H)],
            [("D", 5, W)],
            [("C", 5, Q), ("D", 5, Q), ("Eb", 5, H)],
            [("D", 5, H), ("Bb", 4, H)],
            [("C", 5, H), ("A", 4, H)],
            [("G", 4, W)]]

    prev = None
    for rep in range(3):
        base = rep * 8 * W
        for i, (root, tones) in enumerate(prog):
            at = base + i * W
            prev = voicing(tones, note("G", 3), note("D", 5), prev)
            # 吉他：上下行分解，每小節八個八分音符。
            seq = prev + [prev[-1]] + list(reversed(prev))[1:]
            for j in range(8):
                gtr.note(at + j * E, seq[j % len(seq)], E - 25, 74 + (6 if j == 0 else 0))
            low.note(at, note(root, 2), H, 78)
            low.note(at + H, note(root, 3), H - 20, 62)
            if rep == 0 and i % 2 == 0:
                harp.chord(at + Q, [p + 12 for p in prev[:2]], H, 54)
            # 長笛的旋律從第二輪才進來，第三輪整段都在。
            if rep >= 1 and tune[i]:
                a = at
                for name, octv, dur in tune[i]:
                    flute.note(a, note(name, octv), dur - 25, 74 + rep * 8)
                    a += dur
    return [gtr, harp, low, flute], 92


def facility():
    """設施內（商店、醫生、訓練師、遊俠中心）：短、明亮、不搶戲。

    降 B 大調——整套裡唯一的大調，走進門會有「安全了」的感覺。
    16 小節、100 BPM，約 38 秒。
    """
    ep = setup("ep", 1, "epiano", 88, 56)
    vibe = setup("vibe", 2, "vibe", 76, 84)
    low = setup("low", 3, "bass", 78, 64)

    prog = [("Bb", ["Bb", "D", "F"]), ("G", ["G", "Bb", "D"]),
            ("Eb", ["Eb", "G", "Bb"]), ("F", ["F", "A", "C"]),
            ("Bb", ["Bb", "D", "F"]), ("Eb", ["Eb", "G", "Bb"]),
            ("C", ["C", "Eb", "G"]), ("F", ["F", "A", "C"])]
    prev = None
    for rep in range(2):
        base = rep * 8 * W
        for i, (root, tones) in enumerate(prog):
            at = base + i * W
            prev = voicing(tones, note("Bb", 3), note("F", 5), prev)
            for j, p in enumerate(prev):
                ep.note(at + j * E, p, H, 72)
            ep.chord(at + H + Q, prev, Q, 60)
            low.note(at, note(root, 2), Q, 76)
            low.note(at + H, note(root, 2), Q, 64)
            if rep == 1:
                vibe.note(at + Q, prev[-1] + 12, H, 58)
    return [ep, vibe, low], 100


def sewer():
    """下水道與地底：C 小調、低、濕。旋律幾乎沒有，靠低頻與零星的滴水。

    20 小節、64 BPM，約 75 秒。滴水用中鼓的最高那顆，落點刻意不對齊拍子。
    """
    pad = setup("pad", 1, "fantasy", 68, 40)
    low = setup("low", 2, "synbass3", 84, 64)
    drip = Track("drip", 9)
    voice = setup("voice", 3, "glasses", 62, 92)

    areas = [["C", "G"], ["Ab", "Eb"], ["C", "G"], ["Bb", "F"], ["G", "D"]]
    prev = None
    for i, tones in enumerate(areas):
        at = i * 4 * W
        prev = voicing(tones, note("C", 3), note("C", 5), prev)
        pad.chord(at, prev, 4 * W - 60, 56)
        pad.swell(at, 4 * W, 60, 92)
        low.note(at, note(tones[0], 1), 2 * W, 76)
        low.note(at + 2 * W, note(tones[1], 1), 2 * W - 60, 64)
        # 滴水：每區五滴，位置用區號錯開，聽起來不會像節拍器。
        for j in range(5):
            pos = at + ((j * 7 + i * 3) % 16) * (W // 4) + (j % 2) * S
            drip.note(pos, TOM_HI, S, 40 + (j % 3) * 8)
        if i % 2 == 0:
            voice.note(at + 2 * W, note(tones[0], 5), W, 48)
    return [pad, low, voice, drip], 64


def vegas():
    """拉斯維加斯：整套裡唯一帶爵士味的一首。

    F 小調配七和弦、電鋼琴與行走低音——廢土上的賭城仍然是賭城。
    這一首刻意與別首不同調也不同語彙：從沙漠走進來會明顯感覺換了地方。
    24 小節、108 BPM，約 53 秒。
    """
    ep = setup("ep", 1, "epiano", 92, 52)
    sax = setup("sax", 2, "sax", 86, 80)
    low = setup("low", 3, "fretless", 84, 64)
    perc = Track("perc", 9)

    # ii–V–i 的循環，加上七音讓它有爵士味。
    prog = [("F", ["F", "Ab", "C", "Eb"]), ("Bb", ["Bb", "D", "F", "Ab"]),
            ("Eb", ["Eb", "G", "Bb", "Db"]), ("Ab", ["Ab", "C", "Eb", "G"]),
            ("Db", ["Db", "F", "Ab", "C"]), ("G", ["G", "Bb", "Db", "F"]),
            ("C", ["C", "E", "G", "Bb"]), ("F", ["F", "Ab", "C", "Eb"])]
    line = [None,
            [("C", 5, Q), ("Eb", 5, E), ("F", 5, E), ("Ab", 5, H)],
            None,
            [("G", 5, Q), ("F", 5, E), ("Eb", 5, E), ("C", 5, H)],
            None,
            [("Db", 5, E), ("C", 5, E), ("Bb", 4, Q), ("Ab", 4, H)],
            [("Bb", 4, Q), ("C", 5, Q), ("Eb", 5, H)],
            [("F", 5, W)]]

    prev = None
    for rep in range(3):
        base = rep * 8 * W
        for i, (root, tones) in enumerate(prog):
            at = base + i * W
            # 電鋼琴只彈三個音（省 partial），根音交給低音部。
            prev = voicing(tones[1:], note("Eb", 3), note("Eb", 5), prev)
            ep.chord(at + E, prev, H, 66)
            ep.chord(at + H + E, prev, Q, 56)
            # 行走低音：根音 → 五度 → 根音 → 半音接近下一個和弦。
            nxt = prog[(i + 1) % 8][0]
            walk = [note(root, 2), note(tones[2], 2), note(root, 2), note(nxt, 2) - 1]
            for j, p in enumerate(walk):
                low.note(at + j * Q, p, Q - 20, 80)
            # 疊音鈸打四分、小鼓落在二四拍的反拍。
            for j in range(4):
                perc.note(at + j * Q, RIDE, E, 54 if j % 2 else 62)
            perc.note(at + Q + E, SNARE, S, 44)
            perc.note(at + 3 * Q + E, SNARE, S, 48)
            if rep >= 1 and line[i]:
                a = at
                for name, octv, dur in line[i]:
                    sax.note(a, note(name, octv), dur - 25, 82 + rep * 6)
                    a += dur
    return [ep, sax, low, perc], 108


def base():
    """科奇斯基地（地圖 0x10–0x14）：機械、緊繃、不放鬆。

    E 小調，低音走十六分音符的頑固音型，銅管一段一段疊上來。
    32 小節、126 BPM，約 61 秒。
    """
    low = setup("low", 1, "synbass", 96, 64)
    brass = setup("brass", 2, "synbrass", 90, 50)
    pad = setup("pad", 3, "fantasy", 70, 86)
    perc = Track("perc", 9)

    # 五個音的頑固音型，長度刻意不是四的倍數：與四四拍錯開，
    # 循環起來聽不出來從哪裡開始。
    ost = ["E", "E", "G", "E", "B"]
    roots = ["E", "E", "C", "C", "G", "G", "A", "B"]
    prev = None
    for bar in range(32):
        at = bar * W
        root = roots[bar % 8]
        for j in range(16):
            p = note(ost[j % 5], 2)
            low.note(at + j * S, p, S - 10, 96 if j % 4 == 0 else 78)
        drums(perc, at, "k..hs..hk..hs..h" if bar % 4 != 3 else "k..hs..hk.khs.sk")
        if bar % 8 == 0:
            perc.note(at, CRASH, Q, 100)
        # 銅管：每兩小節一次短音，第三段之後改長音把張力壓上去。
        tones = {"E": ["E", "G", "B"], "C": ["C", "E", "G"],
                 "G": ["G", "B", "D"], "A": ["A", "C", "E"],
                 "B": ["B", "D", "F#"]}[root]
        prev = voicing(tones, note("E", 3), note("E", 5), prev)
        if bar >= 16:
            brass.chord(at, prev, W - 40, 88)
        elif bar % 2 == 1:
            brass.chord(at + H, prev, Q, 82)
        if bar >= 8:
            pad.chord(at, [prev[0] + 12], W - 40, 54)
    return [low, brass, pad, perc], 126


def combat():
    """戰鬥：低音頑固音型 ＋ 銅管刺點 ＋ 鼓。

    D 小調、138 BPM、36 小節，約 63 秒——上一版只有 18 秒，
    一場架就會循環好幾遍。四段結構：進場、加壓、鼓過門、收回開頭。
    """
    low = setup("low", 1, "synbass", 100, 64)
    brass = setup("brass", 2, "synbrass", 94, 48)
    lead = setup("lead", 3, "square", 78, 84)
    perc = Track("perc", 9)

    pattern = ["D", "D", "D", "F", "D", "D", "C", "D"]
    stabs = {0: ["D", "F", "A"], 1: ["Bb", "D", "F"],
             2: ["C", "E", "G"], 3: ["A", "C#", "E"]}
    riff = [("D", 5, E), ("F", 5, E), ("A", 5, Q), ("G", 5, E), ("F", 5, E), ("E", 5, Q)]

    prev = None
    for bar in range(36):
        at = bar * W
        for j, name in enumerate(pattern):
            oct_ = 2 if bar < 8 else (2 if j % 2 == 0 else 1)
            low.note(at + j * E, note(name, oct_), E - 20, 102 if j == 0 else 88)
        if bar % 8 == 7:
            # 過門：三個中鼓由低到高滾上去，接回開頭。
            for j, k in enumerate([TOM_LO, TOM_LO, TOM_MID, TOM_MID, TOM_HI, TOM_HI, SNARE, SNARE]):
                perc.note(at + j * E, k, E, 78 + j * 4)
        else:
            drums(perc, at, "k..hs..hk.khs..h" if bar % 2 else "k..hs..hk..hs.sh")
        if bar % 16 == 0:
            perc.note(at, CRASH, Q, 104)
        prev = voicing(stabs[(bar // 2) % 4], note("D", 4), note("D", 5), prev)
        if bar % 2 == 1:
            brass.chord(at + H, prev, Q, 96)
        # 主奏從第 12 小節進來，第 24 小節之後高八度。
        if bar >= 12 and bar % 4 == 0:
            a, shift = at, 12 if bar >= 24 else 0
            for name, octv, dur in riff:
                lead.note(a, note(name, octv) + shift, dur - 20, 80)
                a += dur
    return [low, brass, lead, perc], 138


def ending():
    """結局：整套裡唯一從小調走到大調的一首。

    D 小調起、D 大調收（皮卡第三度）——四把鑰匙、四個站台、240 步的倒數
    之後，這裡是唯一該讓人鬆一口氣的地方。24 小節、66 BPM，約 87 秒。
    """
    pad = setup("pad", 1, "strings2", 86, 50)
    lead = setup("lead", 2, "oboe", 92, 78)
    low = setup("low", 3, "cello", 84, 64)
    bell = setup("bell", 4, "tubebell", 70, 100)
    perc = Track("perc", 9)

    prog = ([("D", ["D", "F", "A"]), ("Bb", ["Bb", "D", "F"]),
             ("G", ["G", "Bb", "D"]), ("A", ["A", "C#", "E"])] * 2 +
            [("Bb", ["Bb", "D", "F"]), ("F", ["F", "A", "C"]),
             ("C", ["C", "E", "G"]), ("G", ["G", "B", "D"]),
             ("D", ["D", "F#", "A"]), ("A", ["A", "C#", "E"]),
             ("G", ["G", "B", "D"]), ("D", ["D", "F#", "A"])] +
            [("D", ["D", "F#", "A"]), ("A", ["A", "C#", "E"]),
             ("Bb", ["Bb", "D", "F"]), ("G", ["G", "B", "D"]),
             ("D", ["D", "F#", "A"]), ("G", ["G", "B", "D"]),
             ("A", ["A", "C#", "E"]), ("D", ["D", "F#", "A"])])
    mel = [None, None, [("D", 5, H), ("F", 5, H)], [("E", 5, W)],
           [("F", 5, H), ("G", 5, H)], [("A", 5, W)],
           [("G", 5, H), ("F", 5, H)], [("E", 5, W)],
           [("D", 5, Q), ("F", 5, Q), ("A", 5, H)], [("G", 5, W)],
           [("E", 5, H), ("G", 5, H)], [("F#", 5, W)],
           [("A", 5, H), ("D", 6, H)], [("C#", 6, W)],
           [("B", 5, H), ("A", 5, H)], [("F#", 5, W)],
           [("A", 5, W)], [("G", 5, H), ("E", 5, H)],
           [("F#", 5, H), ("A", 5, H)], [("B", 5, W)],
           [("A", 5, H), ("F#", 5, H)], [("G", 5, H), ("B", 5, H)],
           [("A", 5, W)], [("D", 6, W)]]

    prev = None
    for i, (root, tones) in enumerate(prog):
        at = i * W
        prev = voicing(tones, note("D", 3), note("A", 4), prev)
        pad.chord(at, prev, W - 30, 62 + min(i, 16))
        low.note(at, note(root, 2), H, 80)
        low.note(at + H, note(root, 2), H - 20, 64)
        if i >= 8:
            bell.note(at, note(root, 5), H, 46 + (i - 8) * 2)
        if i >= 16 and i % 4 == 0:
            perc.note(at, TOM_LO, Q, 70)
        if mel[i]:
            a = at
            for name, octv, dur in mel[i]:
                lead.note(a, note(name, octv), dur - 30, 84 + min(i, 20))
                a += dur
    for p in range(6):
        lead.swell(p * 4 * W, 4 * W, 84, 100 + p * 4)
    return [pad, lead, low, bell, perc], 66


TUNES = {
    "theme": theme,
    "desert": desert,
    "night": night,
    "town": town,
    "facility": facility,
    "sewer": sewer,
    "vegas": vegas,
    "base": base,
    "combat": combat,
    "ending": ending,
}


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
        ticks = max(e[0] for t in tracks for e in t.events)
        bars = ticks / W
        secs = ticks / TPQ * 60 / bpm
        print(f"{name:9s} {bpm:3d} BPM  {bars:5.1f} 小節  {secs:5.1f} 秒  "
              f"{len(tracks)} 軌  {size:6d} bytes")
    return 0


if __name__ == "__main__":
    sys.exit(main())
