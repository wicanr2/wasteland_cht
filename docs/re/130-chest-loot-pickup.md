# 130：寶箱的撿拾流程——「誰要撿？」與逐件拿取

日期：2026-08-25 ｜ 接 `docs/re/29` §4（nibble 5 的內容生成）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/29` §4 解了「第一次踩到才擲出內容並寫回記錄」，但停在
`0xFF → 0x1534B`。這一份把 `0x1534B` 之後讀完：**內容決定之後，
原版切到名單畫面問「誰要撿」，選了人再開一張物品清單逐件拿**。

---

## 1. 流程總覽（`0x15280` 的後半）

```
0x1534B  sub_1728C                 ; 切到名單畫面（ds:46B9h ← 1）
0x1534E  al ← 0x10；sub_16CB2      ; 印字串 16 `Who wants loot?`
0x15353  sub_1721B(0)              ; 挑隊員（與 USE 的「Which player?」同一支）
0x15358  ＝ 0 → sub_163C4；clc；retn   ; 取消 → 回地圖（東西留在原地）
0x15361  sub_17029；sub_172BB
0x15369  CF ＝ 1 → 印字串 17 ` can't get any.` → 回 0x1534B    ; 倒下的人不能撿
0x15372  sub_15488                 ; 建物品清單（表頭 ＝ 字串 19 `\x0B, take an item`）
0x15375  sub_16DB4；sub_16D34      ; 畫一頁 ＋ 等按鍵（docs/re/53 的清單框架）
0x15387  ＝ 0xFF → 回 0x1534B      ; ESC ＝ 換人／收手
0x1538B  ≥ 0xFD → 重畫
0x1538F  索引×2 ＝ 記錄裡那一對的位移（清單第 n 列 ↔ 記錄 +0x02+2n）
```

## 2. 拿一件（`0x153CB`）

```
0x153CB  sub_17AE0                 ; 物品表定址
0x153CE  sub_1968A                 ; 選中那個人找空的物品槽
0x153D1  bl ＝ 0xFF → 0x1540B      ; 沒空槽（`You can't carry any more.`，字串 18）
0x153D6  al ← 物品資料 +0x04       ; 容量
0x153DF  sub_1968A → 46B5h         ; 槽位址
0x153E9  [槽] ← 物品編號；[槽+1] ← 容量    ; **附屬 byte ＝ 容量（發滿）**，
                                          ; 與起始裝備同一條規則（docs/re/21 §5.1）
0x153F3  數量 byte（那一對的第二個）−1
0x15403  ＝ 0 → 0x1541F            ; 這一項拿完
0x15405  重畫清單，回等按鍵
```

## 3. 拿完一項與拿完整格（`0x1541F`）

```
0x1541F  記錄[那一對的第一個 byte] ← 0     ; 這一項從清單上消失
0x1542B  ds:469Dh −1                       ; 剩幾項
0x15432  ≠ 1 → 重畫清單
0x15436  al ← 0；sub_169B1                 ; **全部拿完 → 用位移 0 改寫這一格**
0x1543C  sub_163C4                         ; 回地圖
```

**寶箱記錄的 `+0x00`／`+0x01` 是「拿完之後這一格變什麼」的改寫對**——
布袋圖示會消失，就是這一步把第 1 層 nibble 改掉的。

## 4. 錢的特例（物品編號 `0x5E`）

`docs/re/29` §4 的「`0x5E` 改寫成 `0xDE`、後面兩個 byte 各擲一次骰」在
拿取端對上了（`0x153A3`）：

```
0x153A3  ＝ 0x5E → 讀那一對之後的兩個 byte → ds:466Bh／466Ch（金額低／高）
0x153B9  [第二個 byte] ← 0xFF；ds:466Dh ← 0
0x153C2  sub_19895                 ; 把 466B–466D（24-bit）加進選中那個人的金錢
0x153C8  jmp 0x1541F               ; 這一項直接算拿完
```

所以 **`0x5E` ＝ 現金**：它占 3 bytes（`0x5E`、金額低、金額高），
第一次踩到時兩個金額 byte 各 `roll(原值)` 一次（`sub_15441` ×2）——
資料裡寫的是上限，實拿的是 1..上限。

## 5. 用到的字串（表 1）

| 編號 | 內容 | 時機 |
|---:|---|---|
| 16 | `Who wants loot?` | 進名單畫面 |
| 17 | `\x0B can't get any.` | 選到不能行動的人 |
| 18 | `You can't carry any more.` | 物品槽滿（`0x1540B`） |
| 19 | `\x0B, take an item\x0D    #    ITEM` | 物品清單表頭（`sub_15488` 印，逐列回呼 `0x549C`） |

類別的比對對象是**物品類別（`+0x03 >> 3`，`docs/re/45`）**，不是 `+0x03` 的
原始 byte——出貨物品表的 `+0x03` 全是「類別 << 3 ＋ 低位旗標」的形狀，
沒有任何一筆原始值等於出貨寶箱唯一用到的類別 1；類別 1 ＝ 近戰
（`Ax`／`Knife` 那一組），與「寶箱擲出一件近戰武器」相符。

推論等級：§1–§3 **已確認**（逐指令讀完，字串編號與譯文對上）；
類別比對走 `>> 3` **強證據**（`sub_199F1` 未逐行讀，由表的形狀與資料反推）；
§4 的 `sub_19895` ＝「把 24-bit 金額加進角色金錢」**強證據**
（466B–466D 三個 byte 的組裝與金錢欄位是 24-bit 相符，函式本體未逐行讀）；
字串 18 在 `0x1540B` 印 **強證據**（該 chunk 未倒，從訊息內容與位置推）。

## 6. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/chest_tail.json 0x15280 0x1534B 0x15441
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/chest_tail2.json 0x15361 0x1532A 0x1529E
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/chest_tail4.json 0x153CB 0x1541F 0x15488
```

## 7. 這一輪學到的（寫成規則）

- **「開名單畫面」常常不是終點，是選人介面的開場。** `sub_1728C` 之後
  幾乎一定接 `sub_1721B`（挑隊員）——看到前者就往後找後者。
- **同一條「附屬 byte ＝ 容量」的規則在第三個地方出現**（起始裝備、商店、
  撿拾）。發物品的路徑一律走同一支 helper，不要各自為政。
