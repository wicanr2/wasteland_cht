#!/usr/bin/env python3
"""把 wl-atlas 倒出來的 JSON 整理成攻略用的表格。

純 stdlib、不需要遊戲資料——只吃 `cmd/wl-atlas` 的輸出與 `translations/`
裡的譯名。**只排序整理，不加語意推論**（`CLAUDE.md` §1.2 的兩段式）：
「這扇門要開鎖或撬棍」是這裡輸出的事實，「所以先去買撬棍」是攻略正文的判斷。

    tools/go.sh run ./cmd/wl-atlas -dir workplace/orig/wastland \\
        -lang translations/zh-Hant.cat -out workplace/atlas.json
    python3 tools/summarize_walkthrough.py workplace/atlas.json docs/walkthrough/generated
"""

import json
import os
import sys

# 屬性條件的參數是**角色記錄的位移**，不是屬性編號（docs/re/21 §1）。
ATTRS = {
    0x0E: "力量",
    0x0F: "智力",
    0x10: "幸運",
    0x11: "速度",
    0x12: "敏捷",
    0x13: "靈巧",
    0x14: "魅力",
}

# 設施跳表的前五格（`internal/game/facility.go`，docs/re/42、52、72、73）。
FACILITY = {
    0: "醫生",
    1: "商店",
    2: "訓練師",
    3: "角色管理（遊俠中心）",
    4: "結局",
}


def load_tsv(path):
    out = {}
    if not os.path.exists(path):
        return out
    with open(path, encoding="utf-8") as f:
        for line in f:
            line = line.rstrip("\n")
            if not line or line.startswith("#"):
                continue
            key, _, value = line.partition("\t")
            out[key] = value
    return out


def clean(text):
    """把控制碼換成看得懂的樣子，並壓成一行。"""
    if not text:
        return ""
    for bad, good in (("\\x0A", ""), ("\\x0B", "〈名字〉"), ("\\x0F", "〈數〉"),
                      ("\\x11", ""), ("\\x10", ""), ("\\x0D", " "), ("\\x0E", "")):
        text = text.replace(bad, good)
    return " ".join(text.replace("\r", " ").replace("\n", " ").split())


class Names:
    """技能與物品的名字，英中各一份。"""

    def __init__(self, root):
        src = load_tsv(os.path.join(root, "translations/source/exe.tsv"))
        zh = load_tsv(os.path.join(root, "translations/zh-Hant/exe-skills-items.tsv"))
        self.en = {}
        self.zh = {}
        for key, value in src.items():
            if key.startswith("exe:2:"):
                self.en[int(key.split(":")[2])] = clean(value)
        for key, value in zh.items():
            if key.startswith("exe:2:"):
                self.zh[int(key.split(":")[2])] = clean(value)

    def skill(self, n):
        return self._pair(n, 1 <= n <= 35)

    def item(self, n):
        return self._pair(36 + n, 0 <= n <= 94)

    def _pair(self, slot, known):
        if not known:
            return f"編號 {slot}（超出字串表，未解）"
        en, zh = self.en.get(slot, ""), self.zh.get(slot, "")
        if zh and en:
            return f"{zh}（{en}）"
        return zh or en or f"編號 {slot}"


def condition_text(cond, names):
    kind, ident, level = cond["kind"], cond["id"], cond["value"]
    if kind == "技能":
        return f"技能 {names.skill(ident)}，難度 {level}"
    if kind == "物品":
        return f"物品 {names.item(ident)}"
    if kind == "屬性":
        attr = ATTRS.get(ident, f"記錄 +{ident:#04x}（未解）")
        return f"屬性 {attr}，難度 {level}"
    if kind == "隊伍人數":
        return f"隊伍剛好 {ident} 人"
    if kind == "金錢":
        return f"身上有 ${ident}"
    return f"未解型別（{cond['raw']}）"


def string_of(block, slot):
    entry = block.get(slot)
    if not entry:
        return "", ""
    return clean(entry.get("en", "")), clean(entry.get("zh", ""))


def first_byte(entry):
    """記錄的 `+0x00`；記錄取不到時（section 指標是 0）回 0。"""
    parts = entry["bytes"].split()
    return int(parts[0], 16) if parts else 0


def cells_text(entry, limit=6):
    cells = entry.get("cells") or []
    shown = ", ".join(f"({x},{y})" for x, y in cells[:limit])
    if entry["cell_count"] == 0:
        return "—"
    if entry["cell_count"] > limit:
        shown += f" …共 {entry['cell_count']} 格"
    return shown


def inbound_names(atlas):
    """每張地圖是「從哪一句話走進來的」——那句話就是地點名。"""
    out = {}
    for src in atlas["maps"]:
        block = {s["slot"]: s for s in src["strings"]}
        for tele in src["teleports"] or []:
            if tele["back"] or tele["to_resource"] < 0:
                continue
            slot = first_byte(tele) & 0x3F
            en, zh = string_of(block, slot)
            if not en:
                continue
            bucket = out.setdefault(tele["to_resource"], {})
            key = (en, zh)
            bucket[key] = bucket.get(key, 0) + max(tele["cell_count"], 1)
    best = {}
    for res, bucket in out.items():
        best[res] = sorted(bucket.items(), key=lambda kv: -kv[1])
    return best


HEAD = ("<!-- 這個檔由 tools/summarize_walkthrough.py 產生，不要手改。 -->\n"
        "<!-- 重跑：tools/go.sh run ./cmd/wl-atlas -dir workplace/orig/wastland "
        "-lang translations/zh-Hant.cat -out workplace/atlas.json && "
        "python3 tools/summarize_walkthrough.py workplace/atlas.json "
        "docs/walkthrough/generated -->\n\n")


def write(path, title, intro, body):
    with open(path, "w", encoding="utf-8") as f:
        f.write(HEAD)
        f.write(f"# {title}\n\n{intro}\n\n")
        f.write(body)
        if not body.endswith("\n"):
            f.write("\n")


def maps_doc(atlas, names, out_dir):
    inbound = inbound_names(atlas)
    lines = ["| 資源 | 檔案 | 邊長 | 遭遇分母 | 一步幾分鐘 | 走進來時畫面上那句話 |",
             "|---:|---|---:|---:|---:|---|"]
    for m in atlas["maps"]:
        label = "、".join(f"{zh or en}" for (en, zh), _ in inbound.get(m["resource"], [])[:2])
        denom = m["encounter_denominator"]
        lines.append(f"| {m['resource']} | {m['file']} | {m['dim']} | "
                     f"{denom if denom else '不擲'} | {m['step_minutes']:g} | {label or '（沒有具名入口）'} |")
    lines.append("")
    lines.append("## 通往哪裡")
    lines.append("")
    lines.append("| 資源 | → 資源 | 目的座標 | 要先問一句 | 走進來時那句話 | 幾格 |")
    lines.append("|---:|---:|---|---|---|---:|")
    for m in atlas["maps"]:
        block = {s["slot"]: s for s in m["strings"]}
        for tele in m["teleports"] or []:
            slot = first_byte(tele) & 0x3F
            en, zh = string_of(block, slot)
            target = "回程（原路退出）" if tele["back"] else str(tele["to_resource"])
            where = f"({tele['to_x']},{tele['to_y']})"
            if tele["relative"]:
                where += " 相對"
            lines.append(f"| {m['resource']} | {target} | {where} | "
                         f"{'是' if tele['asks_first'] else ''} | {zh or en} | {tele['cell_count']} |")
    write(os.path.join(out_dir, "maps.md"), "地圖一覽與連結",
          "42 個 MSQ 區塊、它們之間的傳送關係，以及走進去時畫面上那一句話"
          "（nibble 10 記錄 `+0x00` 的低 6 位，`docs/re/60` §3.1）。",
          "\n".join(lines))


def gates_doc(atlas, names, out_dir):
    parts = []
    for m in atlas["maps"]:
        gates = [g for g in (m["gates"] or []) if g["conditions"]]
        if not gates:
            continue
        block = {s["slot"]: s for s in m["strings"]}
        parts.append(f"## 資源 {m['resource']}（{m['file']}）\n")
        parts.append("| 記錄 | 格子 | 擋住時的訊息 | 過得去的條件（依序試） |")
        parts.append("|---:|---|---|---|")
        for g in gates:
            en, zh = string_of(block, g["message_slot"])
            conds = "<br>".join(condition_text(c, names) for c in g["conditions"])
            parts.append(f"| {g['record']} | {cells_text(g)} | {zh or en} | {conds} |")
        parts.append("")
    write(os.path.join(out_dir, "gates.md"), "條件閘：哪一格要什麼",
          "nibble 2 的條件串列（`docs/re/65`）。**依序試到第一個成功為止**，"
          "所以同一格通常給了好幾條路：技能、屬性檢定，或身上有某件東西"
          "（物品那一條會消耗一次）。難度是條件 byte 的低 5 位。\n\n"
          "格子欄的 `—` 表示出貨資料裡沒有格子指到，要等劇情把某一格改寫過去"
          "——沙漠高溫的三段就是這樣（`docs/re/75`）。",
          "\n".join(parts))


def passwords_doc(atlas, out_dir):
    parts = []
    for m in atlas["maps"]:
        menus = [x for x in (m["menus"] or []) if x["valid"] and x["answers"]]
        if not menus:
            continue
        parts.append(f"## 資源 {m['resource']}（{m['file']}）\n")
        parts.append("| 記錄 | 格子 | 模式 | 題目 | 收得下的答案 |")
        parts.append("|---:|---|---|---|---|")
        for x in menus:
            prompt = clean(x.get("prompt_zh", "")) or clean(x.get("prompt_en", ""))
            answers = "、".join(clean(a) for a in x["answers"] if clean(a))
            parts.append(f"| {x['record']} | {cells_text(x)} | {x['mode']} | "
                         f"{prompt} | {answers} |")
        parts.append("")
    write(os.path.join(out_dir, "passwords.md"), "問答與密語",
          "nibble 8 的問答（`docs/re/46`）。打字題的答案在資料裡就是大寫，"
          "輸入層會把按鍵轉成大寫再逐 byte 全等比對——**沒有模糊比對**。"
          "單鍵題列的是收得下的按鍵。\n\n"
          "格子欄的 `—` 表示出貨資料裡沒有格子指到這一筆：對話是一串，"
          "答對之後才把那一格改寫成下一題（`docs/re/46` §4.1）。",
          "\n".join(parts))


def facilities_doc(atlas, out_dir):
    parts = []
    for m in atlas["maps"]:
        rows = m["facilities"] or []
        chests = m["chests"] or []
        if not rows and not chests:
            continue
        parts.append(f"## 資源 {m['resource']}（{m['file']}）\n")
        if rows:
            parts.append("| 記錄 | 格子 | 是什麼 | 招牌 |")
            parts.append("|---:|---|---|---|")
            for f in rows:
                if f["kind"] == "facility":
                    what = FACILITY.get(f["jump_index"], f"跳表 {f['jump_index']}（未解）")
                elif f["opcode"] >= 0:
                    what = f"腳本 opcode {f['opcode']}"
                else:
                    what = "腳本（opcode 查不到）"
                parts.append(f"| {f['record']} | {cells_text(f)} | {what} | "
                             f"{(f.get('name') or '').strip()} |")
            parts.append("")
        if chests:
            parts.append("| 藏東西的格（nibble 5）| 記錄 | 前 32 bytes |")
            parts.append("|---|---:|---|")
            for c in chests:
                parts.append(f"| {cells_text(c)} | {c['record']} | `{c['bytes']}` |")
            parts.append("")
    write(os.path.join(out_dir, "facilities.md"), "設施、腳本與藏東西的格",
          "nibble 6 是設施與地圖腳本（跳表 `ds:A4E0h`，0–4 是設施、5 以上是"
          "腳本 opcode，`docs/re/34`、`docs/re/79`）；nibble 5 是踩上去有東西的格"
          "（內容生成還沒解，只列 bytes）。\n\n"
          "**商店與醫生在出貨資料裡都沒有格子指到**（格子欄是 `—`）："
          "傳送進去之後由收尾改寫把腳下那一格換成設施才進得了門"
          "（`docs/re/72` §2）。招牌那一欄就是畫面上方那一行字。",
          "\n".join(parts))


def main():
    if len(sys.argv) != 3:
        print(__doc__)
        return 1
    atlas_path, out_dir = sys.argv[1], sys.argv[2]
    root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    with open(atlas_path, encoding="utf-8") as f:
        atlas = json.load(f)
    os.makedirs(out_dir, exist_ok=True)
    names = Names(root)
    maps_doc(atlas, names, out_dir)
    gates_doc(atlas, names, out_dir)
    passwords_doc(atlas, out_dir)
    facilities_doc(atlas, out_dir)
    print(f"{len(atlas['maps'])} 張地圖 → {out_dir}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
