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

```
0x17094  ds:4678h ← ds:471Fh          ; 整行的反白由呼叫端決定
0x1709D  印序號與 `>`
0x170A5  sub_19E30                    ; 反白關
0x170C4  bl ← 0x1F                    ; 裝備索引
0x170CE  == 0 → AMM 印 0
0x170D0  槽 ← 索引 × 2 ＋ 0xBC；al ← 附屬 byte
0x170DA  **bit7 設 → al ← 0**         ; AMM 欄印 0（`docs/re/103` §2）
0x170E2  al &= 0x3F，再過三道閘        ; 過不了也印 0
0x170F6  bl ← 0x28；狀態位元 != 0 → sub_19E2A   ; ← **生病的人反白**
0x1717B  武器附屬 byte bit7 → sub_19E2A          ; ← **卡彈的武器名反白**
```

兩處反白的意思不一樣，但形狀一樣：**它是「這一格有問題」的統一標記**。

| 條件 | 反白的是 |
|---|---|
| 角色記錄 `+0x28`（八個狀態位元）非 0 | 接在後面那一欄（傷勢／體力那一段）|
| 裝備武器的附屬 byte `bit7` | **武器名字** |
| `ds:471Fh` 非 0 | 整行（呼叫端設，用途另解）|

推論等級：**已確認**（`sub_19E2A` 三行直讀，`ds:4678h` 的消費端在 `docs/re/14` 已確認）。

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
| **反白** | **沒接**：名單目前一律正常字。低解那條路有 `render.Frame.DrawGlyph(..., inverse)`，高解（中文）那條路還沒有反白 |

⚠ 要補反白得**兩條路都補**：中文名單走 `HiFrame`，那一層現在只有「畫上前景色」，
沒有「整格填滿再挖空」。只補一條會變成「英文畫面看得到卡彈、中文畫面看不到」。

## 4. 這一輪學到的（寫成規則）

- **「未解」之前先看那支函式有幾行。** 這一支三行、而且它寫的變數在另一份筆記裡
  早就解過了。掛著「未解」一個月的東西，成本是**一次 `dump` 加一次 grep**。
- **同一個旗標被兩處設起來，先假設它們共用一個語意**（「這一格有問題」），
  而不是各自代表不同的東西。

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
