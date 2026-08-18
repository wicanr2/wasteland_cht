# 103：名片行的 `AMM` 與 `WEAPON` 兩欄

日期：2026-08-17 ｜ 接 `docs/re/15` §4（名片行的欄位順序）、`docs/re/45`（物品資料表）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/15` §4 把 `sub_1708B` 的六個欄位順序釘死了，但只有 `NAME`／`AC`／
`MAX`／`CON` 四欄寫進 remake，`AMM` 與 `WEAPON` 兩欄**一直是空的**。
這一份把那兩欄逐指令讀完。

---

## 1. `AMM`（欄 `0x15`）＝ 裝備武器剩幾發

```
0x170C1  ds:4672h ← 15h                  ; 欄座標
0x170C4  bl ← 1Fh；al ← [記錄 +0x1F]      ; 裝備武器的物品槽號
0x170CC  test al, al；jz 印 0             ; ← 閘一：沒有裝備
0x170D0  shl al, 1；add al, 0BCh          ; bl ← 0xBC ＋ 2n
0x170D6  al ← [記錄 + bl]                 ; 那一格的**附屬 byte**
0x170D8  test al, al；jns 繼續
0x170DC  al ← 0；jz 印 0                  ; ← 閘二：bit7 設起來
0x170E2  and al, 3Fh                      ; 低 6 位 ＝ 剩餘次數
0x170E4  push ax
0x170E5  sub_196B2                        ; → 定址那件物品的資料
0x170E8  sub_199F1                        ; → al ＝ 物品 +0x03 >> 3 ＝ 類別
0x170EB  sub_19D2F                        ; ← 閘三：類別在 ds:CD00h 那張表裡嗎
0x170EE  pop ax
0x170EF  jnb 印 al                        ; CF ＝ 0（在表裡）→ 印次數
0x170F1  al ← 0
loc_170F3:
0x170F3  sub_1769C                        ; 印數字
```

三支輔助：

| 函式 | 做什麼 |
|---|---|
| `sub_19AC8(n)` | `al ← 2n ＋ 0xBB` —— 物品槽 n 的**編號 byte** 位移 |
| `sub_196C9` | 讀 `+0x1F`，非 0 就套 `sub_19AC8`；`ZF ＝ 1` ⇔ 沒有裝備 |
| `sub_19D2F(類別)` | 掃 `ds:CD00h` 找相等的 byte，遇到負值收尾；找到 `CF ＝ 0` |

`ds:CD00h` ＝ `0D 0A 0B 0C 02 03 04 05 06 07 08 09 FF …` ——
就是 remake 的 `ItemClass.Ranged()`（`internal/game/items.go` 早就照這張表寫好了）。

> **`+0x1F` 是 1 起算的。** `sub_19AC8` 算的是 `0xBB ＋ 2n`，而物品陣列從
> `+0xBD` 起（`docs/re/15`），所以 `n ＝ 1` 指的是**第 0 格**。
> 當成 0 起算會整批差一格，而症狀是「顯示的武器是背包裡的下一件」——
> 看起來像資料的問題。

推論等級：**已確認**（三道閘與三支輔助逐指令讀過，`ds:CD00h` 從映像讀出）。

## 2. `WEAPON`（欄 `0x20`）＝ 裝備武器的名字

```
0x1715D  ds:4672h ← 20h
0x17165  sub_196C9                        ; bl ← 槽位移；ZF ⇔ 沒有裝備
0x17168  jz 跳過
0x1716A  bl ← al；al ← [記錄 + bl]         ; 物品編號
0x17172  push ax
0x17173  bl++；al ← [記錄 + bl]            ; 附屬 byte
0x17177  test al, al；jns 不做
0x1717B  sub_19E2A                        ; bit7 設起來 → 反白開（docs/re/111）
0x1717E  pop ax
loc_1717F:
0x1717F  ds:4687h ← 1
0x17185  sub_17AF0                        ; 印物品名
```

名字走物品名稱表（`docs/re/17` §4，索引 ＝ 物品編號 ＋ 36）。

`sub_19E2A` 是**反白開**（`ds:4678h ← 0xFF`，三行，`docs/re/111`）：
bit7 設起來時**武器名字畫成反白**。同一支在 `0x17102` 也被叫一次——
角色記錄 `+0x28`（狀態位元）非 0 時反白。兩處是同一個意思：**這一格有問題**。

推論等級：**已確認**（欄位來源、印出的內容、以及 bit7 的視覺標記）。

## 3. remake 接上去的樣子

| 欄 | 來源 |
|---|---|
| `AMM` | `equippedSlot(c).Value & 0x3F`，過不了三道閘就是 `0` |
| `WEAPON` | `itemName(equippedSlot(c).ID)` |

中文那一條另外組（`rosterRowCJK`）：**狀態字**走 `ui:wound.*`、
**武器名**走物品名的中文。數字欄與角色名不必翻。

⚠ **死亡那一格不是文字是字模**（`game.WoundDead` ＝ `\x7f`，`docs/re/17` §4.4），
不進翻譯目錄。

⚠ 查不到翻譯就整行退回英文——**不要拼出半中半英的名單**（與表頭同一條規矩）。

## 4. 與原版實機截圖對過

`docs/re/47` §5 錄下的名單畫面就是這兩欄的 oracle：

```
1)Hell Razor      AC 0  AMM  0  MAX 28  CON 28  WEAPON Crowbar
2)Angela Deth     AC 0  AMM 18  MAX 27  CON 27  WEAPON VP912 9
3)Thrasher        AC 0  AMM  0  MAX 34  CON 34  WEAPON Knife
4)Snake Vargas    AC 0  AMM 18  MAX 31  CON 31  WEAPON VP912 9
```

四行逐欄對上（`TestRosterMatchesOriginalScreenshot`）。
**第一次跑對出一個既有的錯**：`Knife` 在 remake 顯示成 `Kni`——
名字的單數形是**字根 ＋ 單數字尾**，而 `singular()` 只取了字根
（`docs/re/17` §4.1；修正見 `docs/re/28` §2）。

那個錯**不是這一批改出來的**：`itemName` 一直這樣，USE 清單與商店也受影響。
它躲過了所有測試，因為出貨資料裡大多數名字的單數字尾是空的
（`Crowbar\n\ns\n`），兩種讀法結果一樣。

> **拿實機截圖對帳的價值就在這裡**：它一次驗四行 × 六欄，
> 而其中一欄的錯是別的地方留下來的。

## 5. 可重跑的完整指令

```bash
tools/go.sh test ./internal/play/ -run TestRoster -v
python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xCD00 16 --bytes
```

逐指令的來源是 `workplace/analysis/dumps/listing.json`。

## 6. 這一輪學到的（寫成規則）

- **「這一欄是空的」不會有任何測試變紅。** 名片行的六欄裡有兩欄從頭到尾沒接，
  而所有欄位測試都只驗有接的那四欄。**盤點介面時要對著原版的欄位清單數**，
  不要對著自己的測試數。
- **索引是 0 起算還是 1 起算，看那支換算函式的常數。** `0xBB ＋ 2n` 配上
  `+0xBD` 的陣列起點，等於 n 從 1 起算。**兩個常數要一起看**，
  只看陣列起點會讀成 0 起算而且一路自洽。
- **三道閘裡最容易漏的是最後一道**（類別要在表裡）。前兩道是「有沒有」，
  第三道是「這種武器有沒有彈藥欄」——漏掉的話近戰武器會顯示一個數字，
  而那個數字本身是合法的。
- **驗一個「切字尾」的規則，要挑字尾非空的樣本。** `Crowbar\n\ns\n` 這類
  單數字尾是空的，錯的讀法與對的讀法輸出一樣；只有 `Kni\nfe\nves\n`
  分得出來。**測資要挑會分岔的那一個，不是最常見的那一個。**
