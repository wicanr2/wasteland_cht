"""把 `seg005` 的音效資料倒成人看得懂的形式（`docs/re/44`）。

純 stdlib，不需要 IDA——音效的表與位元組碼都直接躺在執行檔的 `seg005`
（線性 `0x39200`–`0x39560`）。這支只做「照結構排版」，**不加任何語意推論**：
欄位名稱來自 `sub_1CCC9`／`sub_1CD52` 讀寫那些位移的程式碼，
不是從資料長相猜的。

用法：
    python3 tools/summarize_sfx.py <wl.merged.exe> [輸出.md]
"""

from __future__ import annotations

import struct
import sys
from pathlib import Path

SEG005 = 0x39200
SEG005_END = 0x39560
CODE_BASE = 0x10000
PIT_HZ = 1193182

PITCH_TABLE = 0x0174  # 12 個半音的除數（最低八度）
SFX_TABLE = 0x0192  # 音效編號 → 四個聲部的位元組碼位址
LEN_TABLE = 0x01DB  # 音長表（索引 ＝ 音符 byte 的低 5 位）
BYTECODE = 0x01F2  # 位元組碼區的起點
VOICE_SIZE = 0x2E
VOICE_COUNT = 4
SFX_COUNT = (LEN_TABLE - 1 - SFX_TABLE) // 8  # 音長表之前塞得下幾筆

# 聲部結構的欄位。名稱一律標出「哪一段程式碼讀寫它」。
FIELDS = {
    0x00: "剩餘時值（0 ＝ 這個聲部沒在發聲）",
    0x02: "位元組碼指標",
    0x04: "頻率除數（每 tick 加 +0x06）",
    0x06: "滑音增量",
    0x08: "這一 tick 送進 8253 的除數",
    0x0A: "喇叭閘控（or 進 port 61h；0 ＝ 靜音）",
    0x0C: "閘控增量（封套用）",
    0x0E: "音長倍率",
    0x10: "切封套第二段的時值門檻",
    0x12: "移調（半音）",
    0x14: "封套資料基底",
    0x16: "封套目前位移",
    0x18: "封套步進倒數",
    0x1A: "顫音波表基底",
    0x1C: "顫音相位",
    0x1E: "顫音相位增量",
    0x20: "顫音深度",
    0x22: "顫音相位環繞值",
    0x24: "通用計數器（只有 0xFE 迴圈用）",
}

NOTE_NAMES = ["C", "C#", "D", "D#", "E", "F", "F#", "G", "G#", "A", "A#", "B"]


def load_seg005(path: Path) -> bytes:
    raw = path.read_bytes()
    e_cparhdr = struct.unpack_from("<H", raw, 8)[0]
    start = SEG005 - CODE_BASE + e_cparhdr * 16
    return raw[start : start + (SEG005_END - SEG005)]


def field(off: int) -> str:
    return FIELDS.get(off, "未解")


def disasm(seg: bytes, at: int, seen: set[int]) -> list[str]:
    """從 at 解一段位元組碼。回傳每一行的文字。

    停止條件與原版一致：設 `+0x00` 之後那一批就結束（`sub_1CD52` 的
    `cmp bl, 0` → 存回指標並 `retn`）。所以這裡一次只解「一批」，
    下一批要等時值歸零才會再解——輸出用縮排把批次分開。
    """
    out: list[str] = []
    i = at
    for _ in range(256):  # 跑掉的話寧可截斷，也不要無限迴圈
        if i >= len(seg):
            out.append(f"  {i:#06x}  ⚠ 超出 seg005")
            return out
        op = seg[i]
        pc = i
        i += 1
        if op == 0xFF:  # 設定欄位
            f = seg[i]
            val = struct.unpack_from("<H", seg, i + 1)[0]
            i += 3
            out.append(f"  {pc:#06x}  設 +{f:#04x} = {val:#06x}   ; {field(f)}")
            if f == 0x00:
                if val == 0:
                    out.append("          └─ 時值 0 → 指標清成 0，**這一條路到此為止**")
                    return out
                out.append(f"          ── 批次結束，等 {val} 個 tick 再解下一批")
        elif op == 0xFE:  # 迴圈
            f = seg[i]
            tgt = struct.unpack_from("<H", seg, i + 1)[0]
            i += 3
            out.append(
                f"  {pc:#06x}  迴圈 +{f:#04x} → {tgt:#06x}"
                f"   ; 計數器為 0 就無條件跳，否則減 1、非零才跳"
            )
            if tgt in seen:
                # 已經解過就走**不跳**那條路，這樣看得到計數器數完之後的尾巴。
                # ⚠ 但這只是兩條路的其中一條——`0xFE` 走哪一條要看執行期的
                # 計數器，靜態列表**決定不了一首會不會停**。誰會停由
                # `internal/audio` 的測試實跑決定（音效 6 就是靠這樣才發現
                # 它其實是無限循環的）。
                out.append(f"          ↩ {tgt:#06x} 已經解過，往下走計數器數完的那條路")
                continue
            seen.add(tgt)
            i = tgt
            continue
        elif op == 0xFD:  # 換寫入基底
            tgt = struct.unpack_from("<H", seg, i)[0]
            i += 2
            out.append(f"  {pc:#06x}  改寫入基底 → {tgt:#06x}   ; 之後的欄位設定套用到那裡")
            continue
        else:  # 音符
            notes = []
            while True:
                voice, length = op >> 5, op & 0x1F
                pitch_byte = seg[i]
                i += 1
                notes.append((voice, length, pitch_byte & 0x7F))
                more = pitch_byte & 0x80
                if not more:
                    break
                op = seg[i]
                i += 1
            desc = "、".join(
                f"聲部 {v} 音長[{ln}] 音高 {p}"
                f"（{NOTE_NAMES[p % 12]}{p // 12}）"
                for v, ln, p in notes
            )
            out.append(f"  {pc:#06x}  音符 {desc}")
            out.append("          ── 批次結束（時值由音長表決定）")
    out.append(_truncated())
    return out


def _truncated() -> str:
    return "  ⚠ 解了 256 個指令還沒結束，截斷"


def main() -> None:
    if len(sys.argv) < 2:
        raise SystemExit(__doc__)
    seg = load_seg005(Path(sys.argv[1]))
    lines: list[str] = []
    w = lines.append

    w("# 音效資料（`seg005`）")
    w("")
    w("由 `tools/summarize_sfx.py` 產生。**只排版，不加語意推論。**")
    w("")
    w("⚠ `0xFE` 走哪一條路要看執行期的計數器，所以**這份列表決定不了一首會不會停**。")
    w("誰會停由 `internal/audio` 的測試實跑決定。")
    w("")
    w(f"`seg005` 線性 {SEG005:#x}–{SEG005_END:#x}，共 {len(seg)} bytes。")
    w("")

    w("## 1. 聲部結構（4 個 × 0x2E bytes，從 `seg005:0x0000` 起）")
    w("")
    w("| 位移 | 用途 |")
    w("|---|---|")
    for off in sorted(FIELDS):
        w(f"| `+{off:#04x}` | {FIELDS[off]} |")
    w("")

    w(f"## 2. 音高表（`seg005:{PITCH_TABLE:#06x}`，12 個半音）")
    w("")
    w("| 半音 | 除數 | 頻率 |")
    w("|---|---|---|")
    for n in range(12):
        div = struct.unpack_from("<H", seg, PITCH_TABLE + n * 2)[0]
        w(f"| {NOTE_NAMES[n]} | `{div:#06x}` | {PIT_HZ / div:.2f} Hz |")
    w("")
    w("八度靠右移：音高 ÷ 12 ＝ 八度數，餘數查表，除數 `shr` 八度數。")
    w("")

    w(f"## 3. 音長表（`seg005:{LEN_TABLE:#06x}`，索引 ＝ 音符 byte 的低 5 位）")
    w("")
    usable = 0x01EF - LEN_TABLE  # 0x1EF 起是全域變數，不是表
    w(f"表後面 `{0x01EF:#06x}` 起就是全域變數，所以**只有索引 0–{usable - 1} 落在表內**。")
    w("")
    row = [f"{seg[LEN_TABLE + i]:#04x}" for i in range(usable)]
    w("| 索引 | " + " | ".join(str(i) for i in range(usable)) + " |")
    w("|---" * (usable + 1) + "|")
    w("| 值 | " + " | ".join(row) + " |")
    w("")

    w(f"## 4. 音效表（`seg005:{SFX_TABLE:#06x}`，每筆 4 個 word ＝ 四個聲部）")
    w("")
    w("| 音效 | 聲部 0 | 聲部 1 | 聲部 2 | 聲部 3 |")
    w("|---|---|---|---|---|")
    starts: list[tuple[int, list[int]]] = []
    for n in range(SFX_COUNT):
        v = list(struct.unpack_from("<4H", seg, SFX_TABLE + n * 8))
        starts.append((n, v))
        cells = " | ".join(f"`{x:#06x}`" if x else "—" for x in v)
        w(f"| {n} | {cells} |")
    w("")

    w(f"## 5. 位元組碼（`seg005:{BYTECODE:#06x}` 起）")
    w("")
    for n, v in starts:
        w(f"### 音效 {n}")
        w("")
        w("```")
        for vi, addr in enumerate(v):
            if not addr:
                continue
            w(f"聲部 {vi} ← {addr:#06x}")
            w("\n".join(disasm(seg, addr, {addr})))
        w("```")
        w("")

    text = "\n".join(lines) + "\n"
    if len(sys.argv) > 2:
        Path(sys.argv[2]).write_text(text, encoding="utf-8")
        print(f"→ {sys.argv[2]}")
    else:
        print(text)


if __name__ == "__main__":
    main()
