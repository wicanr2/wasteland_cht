#!/usr/bin/env python3
"""掃 42 個 MSQ 區塊的 nibble 8 記錄，把每一題的模式與答案倒出來。

nibble 8 是**問答**（`docs/re/46` §4）：記錄 `+0x00` 的 bit7 決定「按一個鍵」
還是「打一行字」，`+0x03` 起是答案的字串編號（最後一個的 bit7 是結束標記）。

這支回答兩件事：

1. 單鍵模式與打字模式各用在多少格——`docs/re/46` §7 掛著的那一題。
2. **哪些字串是密語答案**。那些不能翻譯（比對是逐 byte 全等，
   玩家的輸入層是 ASCII，`docs/re/46` §6）。翻譯流程要拿這份清單去擋。

⚠ 結束標記是「bit7 **設起來**」。拿「bit7 ＝ 0 就停」當條件會一題都讀不到，
而且症狀是安靜的「這張地圖沒有問答」。

用法（純 stdlib，不需要 IDA）：
    python3 tools/summarize_questions.py <wl.unpacked.exe> <game1> <game2> \
        <block-strings.json> [輸出.md]

`block-strings.json` 是 `tools/decode_block_text.py` 的產物，用來把答案的
字串編號翻成實際文字——沒有它就只有編號，翻譯流程擋不了東西。
"""

from __future__ import annotations

import struct
import sys
from pathlib import Path

DS = 0x1CE20
SECTION_TABLE = 0xB9E0  # section 型別 → 標頭內位移
SECTION_TYPES = 24
QUESTION_TYPE = 8  # 地圖第 1 層的 nibble
ANSWER_START = 0x03


def load_blocks(exe: bytes, g1: bytes, g2: bytes):
    """借 summarize_monsters 的區塊解析，避免第二份會漂移的實作。"""
    sys.path.insert(0, str(Path(__file__).resolve().parent))
    import importlib.util

    spec = importlib.util.spec_from_file_location(
        "_mon", Path(__file__).resolve().parent / "summarize_monsters.py"
    )
    mon = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mon)
    return mon.blocks(exe, g1, g2)


def section_offsets(exe: bytes) -> list[int]:
    hdr = struct.unpack_from("<H", exe, 8)[0] * 16

    def at(off: int) -> int:
        return DS + off - 0x10000 + hdr

    return [
        struct.unpack_from("<H", exe, at(SECTION_TABLE) + i * 2)[0]
        for i in range(SECTION_TYPES)
    ]


def main() -> None:
    if len(sys.argv) not in (5, 6):
        sys.exit(__doc__)
    exe, g1, g2 = (Path(p).read_bytes() for p in sys.argv[1:4])
    import json

    texts = {
        b["resource_id"]: b["strings"]
        for b in json.loads(Path(sys.argv[4]).read_text())["blocks"]
    }

    def raw(res_id: int, sid: int) -> str:
        pool = texts.get(res_id) or []
        return pool[sid] if sid < len(pool) else ""

    def typable(res_id: int, sid: int) -> bool:
        """這條字串**打得出來嗎**。

        不是猜測：原版的輸入緩衝區是 16 bytes（`docs/re/46` §4），
        所以超過 15 個字元的字串永遠比不中；含控制碼的同理
        （輸入層丟掉 `< 0x20` 的字元）。
        """
        t = raw(res_id, sid)
        # ⚠ 也要擋純空白：長度與字元範圍都過得了，但那不是答案。
        return bool(t.strip()) and len(t) <= 15 and all(c >= " " for c in t)

    def say(res_id: int, sid: int) -> str:
        pool = texts.get(res_id) or []
        if sid >= len(pool):
            return f"{sid}?"  # 超出這個區塊的字串數 —— 多半是 0xFF 那種哨兵
        return repr(pool[sid])[1:-1] or f"{sid}(空)"
    offsets = section_offsets(exe)
    rel = offsets[QUESTION_TYPE]
    if rel == 0:
        sys.exit(f"section 型別 {QUESTION_TYPE} 在 ds:B9E0h 的位移是 0，掃不了")

    rows = ["# nibble 8 的問答（工具輸出，不含推論）", ""]
    rows.append("| 資源 | 檔案 | 記錄 | 模式 | 題目 | 答案 |")
    rows.append("|---:|---|---:|---|---|---|")

    single = typed = 0
    unterminated = 0
    unreachable = 0
    answer_strings: dict[tuple[str, int], set[int]] = {}
    blocks_with = 0

    for res_id, label, body, map_size in load_blocks(exe, g1, g2):
        start = struct.unpack_from("<H", body, map_size + rel)[0]
        if start == 0 or start + 2 > len(body):
            continue
        first = struct.unpack_from("<H", body, start)[0]
        if not (start < first <= len(body)) or (first - start) % 2:
            continue  # 不是指標陣列，跳過並在結尾報數
        count = (first - start) // 2
        seen_here = False
        for i in range(count):
            ptr = struct.unpack_from("<H", body, start + 2 * i)[0]
            if ptr == 0 or ptr + ANSWER_START >= len(body):
                continue
            rec = body[ptr:]
            prompt = rec[0]
            answers: list[int] = []
            done = False
            for k in range(ANSWER_START, min(len(rec), ANSWER_START + 32)):
                answers.append(rec[k] & 0x7F)
                if rec[k] & 0x80:
                    done = True
                    break
            if not done:
                unterminated += 1
                continue
            mode = "單鍵" if prompt & 0x80 else "打字"
            if prompt & 0x80:
                single += 1
            else:
                hits = [a for a in answers if typable(res_id, a)]
                if not hits:
                    # 一條都打不出來 → 這筆不是打字問答（多半是別的用途的記錄）
                    unreachable += 1
                    continue
                typed += 1
                answer_strings.setdefault((label, res_id), set()).update(hits)
            seen_here = True
            rows.append(
                f"| {res_id} | {label} | {i} | {mode} | {say(res_id, prompt & 0x7F)[:40]} | "
                + "、".join(say(res_id, a) for a in answers)
                + " |"
            )
        if seen_here:
            blocks_with += 1

    rows += [
        "",
        f"合計：**打字 {typed} 題、單鍵 {single} 題**，分布在 {blocks_with} 個區塊。",
        f"讀不到結束標記而跳過的記錄：{unterminated}。",
        f"標成打字模式、但**沒有一條答案打得出來**（超過 15 字或含控制碼）"
        f"而排除的記錄：{unreachable}——那些不是問答。",
        "",
        "## 不能翻譯的字串（打字題的答案）",
        "",
        "比對是逐 byte 全等而輸入層是 ASCII（`docs/re/46` §6），",
        "所以這些字串一旦翻成中文，玩家就永遠打不出來。",
        "",
        "| 資源 | 字串編號 | 內容 |",
        "|---:|---|---|",
    ]
    for label, res_id in sorted(answer_strings):
        ids = sorted(answer_strings[(label, res_id)])
        rows.append(
            f"| {res_id} | "
            + "、".join(str(i) for i in ids)
            + " | "
            + "、".join(say(res_id, i) for i in ids)
            + " |"
        )
    if not answer_strings:
        rows.append("| — | 一題都沒有 |")

    # 機器可讀的清單，給 tools/build_lang.py 當守則用。
    guard = [
        "# 打字問答的答案：**不可翻譯**（`docs/re/46` §6）。",
        "# 比對是逐 byte 全等而輸入層是 ASCII，翻成中文玩家永遠打不出來。",
        "# 由 tools/summarize_questions.py 產生，不要手改。",
        "# key\t原文",
    ]
    for label, res_id in sorted(answer_strings):
        # ⚠ 檔案不能用資源編號猜（資源 27 在 game1），要帶著走。
        for sid in sorted(answer_strings[(label, res_id)]):
            guard.append(f"blk:{label}:{res_id}:{sid}\t{raw(res_id, sid)}")
    Path("translations/must-not-translate.tsv").write_text(
        "\n".join(guard) + "\n", encoding="utf-8"
    )

    text = "\n".join(rows) + "\n"
    if len(sys.argv) == 6:
        Path(sys.argv[5]).write_text(text, encoding="utf-8")
        print(f"→ {sys.argv[5]}")
    else:
        print(text)


if __name__ == "__main__":
    main()
