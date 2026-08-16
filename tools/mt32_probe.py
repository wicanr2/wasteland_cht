#!/usr/bin/env python3
"""量 MT-32／CM-32L 的節奏鍵位：哪些鍵會發聲、發出來的聲音是低是高。

**為什麼要量。** MT-32 的節奏音色表與 General MIDI 不一樣，而且不是全鍵盤都有
指派——送一個沒指派的鍵下去，它就是不出聲，`mt32emu-smf2wav` 也不會抱怨。
症狀是「鼓聲少了一半」，而檔案長度、音量、格式全部正常。

這支不猜，它跑一次真的算繪：每個鍵各敲一下，再回頭量每一段的能量與過零率。

  能量  → 這個鍵有沒有指派（RMS 貼近 0 ＝ 沒有）
  過零率 → 大鼓那類低頻每秒過零幾百次，鈸那類高頻上萬次

過零率分不出小鼓與手鼓，但分得出「低頻打擊 vs 金屬類」，
挑大鼓／小鼓／腳踏鈸／碎音鈸夠用了。

用法：

    python3 tools/mt32_probe.py gen workplace/probe/rhythm.mid
    # 用 render_music.sh 同一套容器算成 wav
    python3 tools/mt32_probe.py analyze workplace/probe/rhythm.wav
"""

import struct
import sys
import wave

TPQ = 480
STEP = TPQ          # 每個鍵隔一個四分音符
DUR = TPQ * 3 // 4
LO, HI = 24, 88     # 掃 24–87，涵蓋 MT-32 文件講的節奏區間
BPM = 120           # 四分音符 ＝ 0.5 秒，換算窗格好算
RHYTHM_CH = 9       # MIDI channel 10（0 起算）＝ 節奏部


def vlq(n):
    out = bytearray([n & 0x7F])
    n >>= 7
    while n:
        out.insert(0, (n & 0x7F) | 0x80)
        n >>= 7
    return bytes(out)


def gen(path):
    ev = []
    for i, key in enumerate(range(LO, HI)):
        at = i * STEP
        ev.append((at, 1, bytes([0x90 | RHYTHM_CH, key, 100])))
        ev.append((at + DUR, 0, bytes([0x80 | RHYTHM_CH, key, 0])))
    ev.sort(key=lambda e: (e[0], e[1]))

    body = bytearray()
    body += vlq(0) + b"\xff\x51\x03" + struct.pack(">I", int(60_000_000 / BPM))[1:]
    last = 0
    for at, _, data in ev:
        body += vlq(at - last) + data
        last = at
    body += vlq(0) + b"\xff\x2f\x00"

    trk = b"MTrk" + struct.pack(">I", len(body)) + bytes(body)
    head = b"MThd" + struct.pack(">I", 6) + struct.pack(">HHH", 0, 1, TPQ)
    with open(path, "wb") as f:
        f.write(head + trk)
    print(f"{HI - LO} 個鍵 → {path}（每鍵 0.5 秒，共 {(HI - LO) * 0.5:.1f} 秒）")


def analyze(path):
    with wave.open(path, "rb") as w:
        ch, width, rate, n = w.getnchannels(), w.getsampwidth(), w.getframerate(), w.getnframes()
        raw = w.readframes(n)
    if width != 2:
        print(f"只讀 16-bit，這份是 {width * 8}-bit", file=sys.stderr)
        return 1
    total = len(raw) // 2
    samples = struct.unpack("<%dh" % total, raw)
    # 只取左聲道，節奏部沒有做 pan 差異，兩聲道判斷一樣。
    mono = samples[::ch]
    print(f"{path}：{rate} Hz、{ch} 聲道、{len(mono) / rate:.1f} 秒\n")

    def rms_of(seg):
        if not seg:
            return 0.0
        return (sum(s * s for s in seg) / len(seg)) ** 0.5

    win = int(rate * 0.5)
    silent, sounding = [], []
    print(f"{'鍵':>4} {'RMS':>8} {'過零率/秒':>10} {'尾巴':>6}  判讀")
    for i, key in enumerate(range(LO, HI)):
        full = mono[i * win:(i + 1) * win]
        if not full:
            break
        # 只看敲下去那 0.3 秒：後面是殘響，會把安靜的鍵也墊高。
        seg = full[:int(rate * 0.3)]
        rms = rms_of(seg)
        zc = sum(1 for a, b in zip(seg, seg[1:]) if (a >= 0) != (b >= 0))
        zcr = zc * rate / len(seg)
        # 尾巴 ＝ 0.3–0.5 秒的能量 ÷ 前 50 ms 的能量。
        # **這一項才分得出腳踏鈸與碎音鈸**：兩者都是高頻，差別在收得快不快。
        head = rms_of(full[:int(rate * 0.05)])
        tail = rms_of(full[int(rate * 0.3):])
        ratio = tail / head if head > 1 else 0.0
        if rms < 30:
            note = "—（沒有指派）"
            silent.append(key)
        else:
            if zcr < 1500:
                note = "低頻打擊"
            elif zcr < 5000:
                note = "中頻"
            else:
                note = "金屬／長衰減" if ratio > 0.25 else "金屬／短衰減"
            sounding.append((key, rms, zcr, ratio))
        print(f"{key:>4} {rms:>8.0f} {zcr:>10.0f} {ratio:>6.2f}  {note}")

    print(f"\n有聲 {len(sounding)} 鍵、無聲 {len(silent)} 鍵")
    if silent:
        print("無聲的鍵：", " ".join(str(k) for k in silent))
    if sounding:
        def top(pred, key, n=3):
            xs = [t for t in sounding if pred(t)]
            return sorted(xs, key=key, reverse=True)[:n]

        print("\n候選（照量到的數字排，不是照 GM 的慣例）：")
        for label, pred, key in (
            ("大鼓　　（低頻、響）", lambda t: t[2] < 800, lambda t: t[1]),
            ("小鼓　　（中頻、響）", lambda t: 1800 < t[2] < 4000, lambda t: t[1]),
            ("腳踏鈸　（高頻、收得快）", lambda t: t[2] > 5000 and t[3] < 0.25, lambda t: t[1]),
            ("碎音鈸　（高頻、拖得長）", lambda t: t[2] > 5000 and t[3] >= 0.25, lambda t: t[1]),
        ):
            xs = top(pred, key)
            got = "、".join(f"{k}（RMS {r:.0f}／過零 {z:.0f}／尾 {q:.2f}）" for k, r, z, q in xs)
            print(f"  {label}：{got or '沒有符合的'}")
    return 0


def main():
    if len(sys.argv) != 3 or sys.argv[1] not in ("gen", "analyze"):
        print(__doc__)
        return 1
    return gen(sys.argv[2]) or 0 if sys.argv[1] == "gen" else analyze(sys.argv[2])


if __name__ == "__main__":
    sys.exit(main())
