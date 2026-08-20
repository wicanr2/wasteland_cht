# 119：醫生與訓練師的進場——招呼語的位移每一種設施不一樣，三種都要先選人

日期：2026-08-20 ｜ 接 [`42`](42-facility-loops.md) §5／§6（醫生迴圈）、
[`52`](52-trainer-facility.md)、[`117`](117-save-globals-and-facility-screen.md)（商店那一輪的對拍）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
實機截圖 `workplace/dosbox/shots/50-doc.png`、`54-doc-menu.png`、`60-tr.png`、`61-tr-menu.png`。

商店那一輪（`docs/re/117`）之後，同一套版面的另外兩種設施照樣對一次。
版面完全相同（肖像框 ＋ 選單區 ＋ 名單 ＋ 指令列），差別在**進場那幾行**。

---

## 1. 三個入口讀的欄位不一樣

| 設施 | 跳表 | 入口 | 招牌從記錄 | 招呼語從記錄 | 問句 |
|---|---:|---|---|---|---|
| 醫生 | 0 | `0x1C260` | `+0x07`（13 bytes）| **`+0x03`** → `sub_178A0` | 字串 8:23 `Who wants treatment?` |
| 商店 | 1 | `0x1BE50` | `+0x07` | **`+0x05`** → `ds:DBF4h` | 字串 7:2 `Who wants to enter?` |
| 訓練師 | 2 | `0x1BBA0` | **`+0x04`** | **`+0x03`** → `sub_178A0` | 字串 6:5 `Who wants to enter?` |

三個入口的骨架一樣：抄招牌 → 載圖（醫生 0、商店 1、訓練師 2）→ `sub_1728C`
切成名單模式 → 印招呼語 → 印問句 → **`sub_1721B` 選人** → 選不到就離開；
選到不能行動的人印一句話再回頭重選：

| 設施 | 不能用的那一句 |
|---|---|
| 醫生 | 字串 8:9 `⟨名字⟩ is beyond my help.` |
| 商店 | 字串 7:3 `⟨名字⟩ can't buy anything.` |
| 訓練師 | 字串 6:6 `⟨名字⟩ is in no condition to learn.` |

> **招呼語是「這張地圖自己的字串」**，不是執行檔字串表：`sub_178A0` 收的是
> 地圖字串的編號。所以同一種設施在不同城鎮講不同的話
> （高池鎮商店「Welcome to the shop.」、達爾文村圖書館「Darwin Branch Library.」）。

⚠ **位移拿錯不會壞掉，只會印出別的句子。** remake 原本三種都讀 `+0x05`：
醫生印成「Entering the infirmary.」（那是走進門那一步的**地圖訊息**）、
訓練師的 `+0x05` 根本落在名字欄位裡。畫面上看起來完全正常。

推論等級：**已確認**（三個入口逐行讀出 ＋ 三張實機截圖逐句對上）。

## 2. 醫生的主選單是條件式的

```
0x1C289  ds:DCE0h ← 1        ; Exam 一定可以
0x1C291  ds:DCE2h ← 0        ; Curing
0x1C295  ds:DCE4h ← 0        ; Healing
…選完人之後…
0x1C2D5  角色記錄 +0x28（狀態位元）非 0 → ds:DCE2h ← 1     ; 有病才有 Curing
0x1C2EA  sub_19A41（CON < MAXCON）      → ds:DCE4h ← 1     ; 受傷才有 Healing
0x1C2F4  三個都是 0 → 回去重新選人
0x1C30B  印字串 8:10「"Well ⟨名字⟩, I would recommend:"」＋ 適用的項目
```

實機截圖（滿血、沒病的 Hell Razor）就只有一行 `Exam $25`。
**三個都印的話玩家會以為隨時能治療**。

## 3. `P` 只有商店與醫生有

| 設施 | 外框下緣 | `ds:470Eh`（清單的第三個出口鍵）|
|---|---|---|
| 商店 | `POOL MONEY` | `0x1C088` 設成 `'P'` |
| 醫生 | `POOL MONEY` | `0x1C5B7` 設成 `'P'` |
| 訓練師 | 只有 `MORE!` | **沒有設** |

所以訓練師沒有 `P`。remake 三種都寫成「換下一個人」，兩種是錯的、一種是多的。

## 4. 訓練師的清單有三欄

實機（`61-tr-menu.png`）：

```
Skill points = 1
     IQ PTS LVL   SKILL
1>    3   4    2  Brawling
2>    3   2    1  Climb
…
9>    6   4    2  Perception
MORE!
```

表頭是字串 6:3（`   IQ PTS LVL   SKILL`）。三欄分別是**IQ 需求**
（技能資料 `+0x00 >> 3`）、**這一級要幾點**（`SkillCost`）、**目前等級**。

⚠ **技能表的第 0 格不是技能**（與物品表第 0 筆同一種佔位）：原版的第一列是
`Brawling`。remake 原本列出來，畫面上多一行沒有名字的 `skill 0`。

## 5. remake 這一側

| 項目 | 狀態 |
|---|---|
| 招呼語位移依設施種類 | **已接**（`greetingAt`）|
| 招呼語先過 `textlayout.Render`（吃掉控制碼）＋ 依面板寬折行 | **已接**（`wrapCells`）|
| 三種設施都問「誰要進去／誰要治療」 | **已接**（`asksWho`、`whoKey`、`whoPrompt`、`cantMessage`）|
| 醫生主選單條件式 ＋ `I would recommend` | **已接**|
| `P`：商店與醫生集中金錢、訓練師沒有 | **已接**|
| 訓練師表頭與三欄、跳過第 0 格 | **已接**（`trainableSkills`、`facility.skillrow`）|

## 6. 可重跑

```bash
# 醫生（高池鎮 Infirmary，門在 (13, 29)）
tools/go.sh run ./cmd/wl-save -dir workplace/dosbox/game -map 10 -at 13,30
tools/dosbox.sh "wait:6;key:Return;wait:3;key:p;wait:4;key:i;wait:2;key:y;wait:4;\
shot:50-doc;key:1;wait:3;shot:54-doc-menu;key:e;wait:4;shot:55-doc-exam"

# 訓練師（達爾文村 Library，門在 (24, 20)）
tools/go.sh run ./cmd/wl-save -dir workplace/dosbox/game -map 21 -at 24,21
tools/dosbox.sh "wait:6;key:Return;wait:3;key:p;wait:4;key:i;wait:2;key:y;wait:4;\
shot:60-tr;key:1;wait:3;shot:61-tr-menu"

# remake 同一條路
tools/go.sh run ./cmd/wl-play -script "map=10:13:30,up,Y,1" -trace
tools/go.sh run ./cmd/wl-play -script "map=21:24:21,up,Y,1" -trace

# 設施的門在哪：atlas 裡「落點改寫成 nibble 6」的傳送格
python3 -c "
import json; d=json.load(open('workplace/atlas.json'))
for m in d['maps']:
    for t in (m.get('teleports') or []):
        b=[int(x,16) for x in t['bytes'].split()]
        if len(b)>5 and b[4]!=0xFF and (b[4]&0x7F)==6:
            print(m['resource'], t.get('cells'), '→ nibble 6 記錄', b[5])"
```

## 7. 這一輪學到的（寫成規則）

- **同一族的三個入口，欄位位移不一定一樣。** 商店讀 `+0x05`、醫生與訓練師讀
  `+0x03`，招牌也分別在 `+0x07` 與 `+0x04`。**把一種設施的版面推廣到另一種
  之前，先把另一種的入口也讀一遍**——它們長得像，但不是同一段程式碼。
- **選單少一項比多一項難發現。** 醫生的條件式選單少了 Healing／Curing 時，
  畫面上只是「這家醫生只提供檢查」；多印兩項則要玩家按下去才會發現沒反應。
  **有旗標的地方就要問「這三個旗標什麼時候是 0」。**
- **外框上的標籤是免費的斷言**：`POOL MONEY` 有沒有出現，直接回答了
  「這一層有沒有這個鍵」——比在程式碼裡找按鍵表快得多。
