# 111：`sub_19E2A` 是「反白開」——卡彈的武器名與生病的隊員會被反白

日期：2026-08-18 ｜ 接 [`103`](103-roster-line-columns.md) §2（`AMM`／`WEAPON` 兩欄）、
[`14`](14-fonts-and-text-encoding.md) §「控制碼 0x01／0x02」（反白旗標）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/103` §2 掛著「武器附屬 byte 的 bit7 設起來時 `sub_19E2A` 多做什麼」。
答案是**三行**：

```
0x19E2A  mov  al, 0FFh
0x19E2C  mov  ds:4678h, al      ; ← 反白旗標
0x19E2F  retn
```

`ds:4678h` 早就解過了（`docs/re/14`）：**文字繪製時非 0 就走
`lodsb ／ not al ／ stosb` 那條路 ＝ 反白**。`sub_19E30` 是同一支的關閉版
（`ds:4678h ← 0`），兩支正好對應文字控制碼 `0x01`／`0x02`。

所以 bit7 不是「多做一件事」，是**把接下來那幾個字畫成反白**。

---

## 1. 名單那一行哪幾段會反白（`sub_1708B`）

整支 `0x1708B`–`0x17190` 倒出來讀，**開關與欄位切換的順序就是答案**：

```
0x17094  ds:4678h ← ds:471Fh          ; 整行的反白由呼叫端決定
0x1709A  印序號與 `>`
0x170A5  sub_19E30                    ; 反白關
0x170A8  sub_171B9                    ; 印名字
0x170AB  ds:4672h ← 0x11              ; 游標欄 → AC
0x170BC  印 AC
0x170BF  ds:4672h ← 0x15              ; 游標欄 → AMM
0x170C4  bl ← 0x1F                    ; 裝備索引
0x170CE  == 0 → AMM 印 0
0x170D0  槽 ← 索引 × 2 ＋ 0xBC；al ← 附屬 byte
0x170DA  **bit7 設 → al ← 0**         ; AMM 欄印 0（`docs/re/103` §2）
0x170E2  al &= 0x3F，再過三道閘        ; 過不了也印 0
0x170F6  bl ← 0x28；狀態位元 != 0 → sub_19E2A   ; ← **反白開**
0x17105  ds:4672h ← 0x18              ; 游標欄 → **MAX**
0x1711C  sub_176C5                    ; 印 MAXCON（16-bit）
0x1711F  sub_19E30                    ; ← **反白關**
0x17122  ds:4672h ← 0x1C              ; 游標欄 → CON（旗標已經關掉了）
0x1715D  ds:4672h ← 0x20              ; 游標欄 → WEAPON
0x1717B  武器附屬 byte bit7 → sub_19E2A          ; ← 反白開
0x17185  sub_17AF0                    ; 印武器名
0x17188  jmp  sub_19E30               ; ← 反白關
```

兩處反白的意思不一樣，但形狀一樣：**它是「這一格有問題」的統一標記**。

| 條件 | 反白的是 |
|---|---|
| 角色記錄 `+0x28`（八個狀態位元）非 0 | **`MAX`（體力上限）那一欄**——開在 `0x170F6`、關在 `0x1711F`，中間只印了 MAXCON |
| 裝備武器的附屬 byte `bit7` | **武器名字** |
| `ds:471Fh` 非 0 | 整行（呼叫端設，用途另解）|

**反白的範圍只到欄位本身的字**：欄與欄之間不是印空白，是把游標欄寫進
`ds:4672h`（`0x17105`、`0x17122`、`0x1715D`）——中間那幾格根本沒有被印過，
旗標碰不到它們。

推論等級：**已確認**（`sub_19E2A` 三行直讀，`ds:4678h` 的消費端在 `docs/re/14`
已確認，開關與 `ds:4672h` 的順序在同一支函式裡逐行讀出）。

## 2. bit7 到底是什麼

`internal/game/resolve.go` 早就把 bit7 當成**卡彈**（`jammedFlag = 0x80`），
依據是 `Load` 指令的字串（`Load/unjam`）與 `AMM` 欄印 0。這一輪補上第三個證據：
**卡彈的武器名在名單上是反白的**——原版特地給它一個視覺標記，
與「這件武器現在不能用」完全一致。

官方手冊那一條：

> `Load` — Load/unjam：把武器裝填，或把卡住的武器排除故障。
> （`docs/manual/06-command-summary.md`）

## 3. remake 這一側

| 項目 | 狀態 |
|---|---|
| bit7 ＝ 卡彈，`AMM` 欄印 0 | **已接**（`internal/play/combat.go` 的 `ammoColumn`）|
| `Load` 指令清掉整個彈匣 | **已接**（`internal/game/resolve.go` 的 `ResolveLoad`）|
| **反白** | **已接**（兩條路都補了）|

`RosterRow` 多兩個旗標（`CONInverse`／`WeaponInverse`），
`InverseAt(lay, maxCON, weapon)` 算出「哪幾欄要反白」（`lay` 是中英各一套的
欄座標），兩條繪製路各拿去用：

| 路 | 怎麼畫 |
|---|---|
| 低解（英文）| `Frame.DrawLineInverse` → 既有的 `DrawGlyph(..., inverse)` |
| 高解（中文）| `HiFrame.FillCell` 塗滿整格，再用**背景色**畫字 |

⚠ **高解那一層沒有背景像素可以取反**：原版的反白是畫字時
`lodsb ／ not al ／ stosb`（`docs/re/14`），前景背景對調；
倚天字模只有「有筆劃的點」，所以拆成兩步。

⚠ **反白範圍要用「這一次真的要畫的那段字」算**。中英兩版的欄座標不同
（中文的武器欄要 14 格才放得下最長的名字），拿另一套去畫會反白到隔壁欄，
而畫面上只是「反白的位置有點怪」。中文那一版是排版函式組出來的一整行字串，
所以 `rosterFieldsCJK` 照欄座標把兩欄切回來——與排版共用同一組常數（`cjkRoster`），
兩邊不會漂。

⚠ 只反白欄位本身的字，**不含後面的補白**——這是 §1 讀出來的，不是取保守。

## 4. 這一輪學到的（寫成規則）

- **「未解」之前先看那支函式有幾行。** 這一支三行、而且它寫的變數在另一份筆記裡
  早就解過了。掛著「未解」一個月的東西，成本是**一次 `dump` 加一次 grep**。
- **同一個旗標被兩處設起來，先假設它們共用一個語意**（「這一格有問題」），
  而不是各自代表不同的東西。
- **旗標「開在哪裡」不等於「作用在哪一欄」——要讀到關掉那一行為止。**
  這一支在開與關之間只印了一欄，而開的位置看起來像在講後面那一整段。
  反錯一欄的畫面照樣有一塊反白，**沒有任何症狀**。
- **欄與欄之間是游標移動，不是印空白。** 判斷「反白涵蓋到哪裡」要看游標欄
  （`ds:4672h`）寫在哪，不能看輸出的字串長什麼樣。

## 5. 可重跑

```bash
docker run --rm --log-opt max-size=10m --log-opt max-file=3 --network none \
  -v "$PWD:/w" -w /w -u "$(id -u):$(id -g)" python:3.12-slim python3 -c "
import json; d=json.load(open('workplace/analysis/dumps/listing.json'))
for i in d['instructions']:
    ea=int(i['ea'],16)
    if 0x19E2A<=ea<0x19E36 or 0x1708B<=ea<0x17190: print(hex(ea), i['disasm'])
print('--- ds:4678h 的讀寫點 ---')
for i in d['instructions']:
    if 'ds:4678h' in i['disasm']: print(i['ea'], i['disasm'].strip())"
```
