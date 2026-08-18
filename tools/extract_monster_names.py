"""把 42 個地圖區塊的**明文敵人名字表**抽成翻譯目錄的原文檔。

輸出 `translations/source/monsters.tsv`，key ＝ `monster:<原文>`（去重）。

名字表在記錄區標頭 `+0x02`／`+0x03` 指到的位移，NUL 分隔、明文 ASCII
（`docs/re/09` §3）。第 0 條是空的（前導 NUL），所以**編號從 1 起算**，
與敵人資料表同一套（`docs/re/114` §6）。

表以 **`0xFF`** 結束。⚠ **不要拿記錄區標頭 `+0x31` 當筆數**：那是
「隨機遭遇擲哪一種」的上限（`sub_16890`），不擲遭遇的地圖它是 0，
而那些地圖照樣有靜態遭遇也照樣有名字。用它會在半數地圖上抽到零條。

單複數用 `\\n` 分段（字根／單數字尾／複數字尾，`docs/re/17` §4.1），
原樣保留成 `\\x0A`——翻譯時**不要刪掉**，顯示層要靠它取單數形。

用法：
    python3 tools/extract_monster_names.py <wl.merged.exe> <game1> <game2> [輸出目錄]
"""

from __future__ import annotations

import struct
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))
from decode_block_text import walk_blocks  # noqa: E402

HDR_NAMES = 0x02  # 名字表位移（記錄區標頭）
MAX_NAMES = 64  # 防呆上限：真正的表最長 15 條左右
MAX_NAME_LEN = 40


def escape(s: str) -> str:
    """控制碼寫成 \\xNN，其餘原樣（與 extract_strings.py 同一套）。"""
    out = []
    for ch in s:
        if ch == "\t":
            out.append("\\t")
        elif ord(ch) < 0x20 or ord(ch) == 0x7F:
            out.append(f"\\x{ord(ch):02X}")
        else:
            out.append(ch)
    return "".join(out)


def names_of(body: bytes, header_at: int) -> tuple[list[str], str]:
    """回傳這個區塊的名字清單（索引 1 起算）與一句備註。"""
    if header_at + HDR_NAMES + 1 >= len(body):
        return [], "標頭超出區塊"
    off = struct.unpack_from("<H", body, header_at + HDR_NAMES)[0]
    if off == 0 or off >= len(body):
        return [], f"名字表位移 {off:#x} 超出區塊"

    out: list[str] = []
    p = off
    while p < len(body) and body[p] != 0xFF and len(out) < MAX_NAMES:
        end = body.find(b"\x00", p)
        if end < 0 or end - p > MAX_NAME_LEN:
            return out, "沒有結尾的 NUL 或名字過長"
        text = body[p:end].decode("latin-1")
        if any(c != "\n" and (c < " " or c > "~") for c in text):
            return out, f"第 {len(out)} 條有非 ASCII 位元組，停在這裡"
        out.append(text)
        p = end + 1
    note = ""
    if len(out) >= MAX_NAMES:
        note = f"超過 {MAX_NAMES} 條還沒遇到 0xFF——大概沒解對"
    return out, note


def main() -> None:
    if len(sys.argv) not in (4, 5):
        sys.exit(__doc__)
    exe = Path(sys.argv[1]).read_bytes()
    game1 = Path(sys.argv[2]).read_bytes()
    game2 = Path(sys.argv[3]).read_bytes()
    outdir = Path(sys.argv[4]) if len(sys.argv) == 5 else Path("translations/source")
    outdir.mkdir(parents=True, exist_ok=True)

    seen: dict[str, list[str]] = {}
    notes: list[str] = []
    for res_id, label, body, header_at, _ in walk_blocks(exe, game1, game2):
        names, note = names_of(bytes(body), header_at)
        if note:
            notes.append(f"{label}#{res_id}：{note}")
        for i, n in enumerate(names):
            # 第 0 條是空的（前導 NUL），而**沒有字母的不是名字**——
            # 表尾之後偶爾會多解出一兩個位元組（`\x0A`、`2`）。
            if i == 0 or not any(c.isalpha() for c in n):
                continue
            seen.setdefault(n, []).append(f"{label}#{res_id}:{i}")

    lines = [f"monster:{escape(n)}\t{escape(n)}" for n in sorted(seen)]
    path = outdir / "monsters.tsv"
    path.write_text("\n".join(lines) + "\n", encoding="utf-8")
    print(f"敵人名字 {len(seen)} 條（去重前 {sum(len(v) for v in seen.values())} 筆）→ {path}")
    for n in notes:
        print("  ⚠", n)


if __name__ == "__main__":
    main()
