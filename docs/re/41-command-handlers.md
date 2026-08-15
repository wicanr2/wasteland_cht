# 41：四支指令處理程式（Hire／Weapon／Use／Load）

日期：2026-08-15 ｜ 對應盤點 **D2**（戰鬥流程）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/38` 解了指令階段的骨架與逃跑、迴避、攻擊；這一份補上剩下四支。

---

## 1. 四支的共同形狀

`docs/re/38` §1 那個迴圈把處理程式的回傳值寫進 `ds:46DAh + 成員`（指令的參數），
CF 設起來就重問這個人。四支處理程式**全部都是這個形狀**：

```
檢查前提 → 不成立就印一句話、stc（重問）
成立就叫出一個選單 → 回傳選擇（＝ 指令的參數）、clc
```

⚠ **它們都不執行動作。** 換武器沒有真的換、裝填沒有真的填、雇用沒有真的雇——
只是把「選了什麼」記下來，動作在結算階段依 `ds:46D8h`／`ds:46DAh` 進行。
與迴避（處理程式是 `clc; retn`）是同一套設計。

推論等級：**已確認**（四支逐指令讀完，回傳路徑與 `docs/re/38` §1 的寫入點對上）。

## 2. Weapon（碼 3）：`0x1234E`

```
sub_1963A            ; 數物品陣列裡有幾件（§2.1）→ dl
dl ＝ 0 → 印字串 64「You don't have anything.」，stc
ds:7DF3h ← 0x0434    ; 選單視窗的位置
jmp sub_19394        ; 物品清單選單，它的回傳值就是這支的回傳值
```

### 2.1 `sub_1963A`：數物品

```
bl 從 0xBD 每次 +2 數到 0xF9        ; ＝ 角色記錄 +0xBD 起的 30 槽物品陣列
[記錄 + bl] ＝ 0 → 跳過
al 進來是 0xFF 時（sub_1963F 入口）多一道 sub_17AE0 ＋ sub_17AF5 的過濾
ds:4721h ← (bl − 0xBB) >> 1        ; 最後一個符合的槽編號
回傳 dl ＝ 數量
```

`0xBD`／stride 2／30 槽與 `docs/re/15` 的物品陣列完全對上。
兩個入口的差別只在要不要過濾——`0x1963A` 不過濾（全部物品），
`0x1963F` 過濾（`sub_17AE0` 是物品表定址，所以是依物品類別篩）。

## 3. Hire（碼 1）：`0x1236F`

```
ds:46CCh ← 0xFF                    ; 接戰值先設成「不限距離」（docs/re/39 §4）
loc_1382B → al ＝ 可雇用的對象數
al ＝ 0 → 印字串 57「No one is within range.」，stc
sub_12262(al)                      ; 印字串 63「Which group?」並開清單
sub_16D34 → al ＝ 選了第幾個
al ＝ 0xFF → stc（取消，重問）
al ≥ 0xFD → 回去重選
al −= 1；sub_16F20 重畫畫面；clc     ; al 就是回傳值
```

`sub_16F20` 是 `sub_19E2A ＋ sub_18D27 ＋ sub_16F35 × 3 ＋ sub_18D2C ＋ sub_19E30`——
純粹是重畫，不是雇用動作。

## 4. Load／unjam（碼 6）：`0x123E2`

**三道檢查，一道都沒過就什麼都不做。**

```
sub_15C96  ; 有沒有裝備武器（sub_196C9 取裝備，讀它的 +1）
           ; 負 → clc; retn（沒裝備，靜靜結束，不印訊息）
sub_15CA4  ; 這把武器吃不吃彈匣（sub_196B2 → 敵人資料 +0x07）
           ; ＝ 0 → 字串 66「\x0B's weapon can't be reloaded.」
sub_1968C  ; 在物品陣列裡找對應的彈匣
           ; 找不到 → 字串 65「\x0B has no more clips.」
三道都過 → clc（接受這個指令）
```

印訊息的收尾是 `sub_1789C(4)` → 印訊息 → `sub_1789C(6)` 再 `stc`；
`sub_1789C` 就是 `jmp word ptr ds:B265h`（每字元處理器，`docs/re/14` §4）。

### 4.1 `sub_1968C`：在物品陣列裡找一件東西

```
bl 從 0xBD 每次 +2 到 0xF9
[記錄 + bl] ＝ 要找的值 → clc，bl 停在那一格
找完沒有 → bl ＝ 0xFF，stc
```

同一支也被裝備／交易用到，是**物品陣列的線性查找**。

## 5. Use（碼 7）：`0x1240B`

唯一一支會**寫進別的地方**的：

```
ds:0A426h ← 成員編號
sub_13AE4 → CF 設就 sub_1728C（切回地圖）＋ stc
    ; sub_13AE4 先印字串 4「Use: Item / Skill / Attribute」，
    ; ⚠ **索引照字母表 ds:A5E8h（53 49 41 ＝ SIA），不是顯示順序**：
    ;   0 ＝ Skill、1 ＝ Item、2 ＝ Attribute（docs/re/92 §2）
ds:0A425h ← dl
sub_12738 → al；al |= ds:0A425h        ; 把兩半合成一個 byte
bl ← ds:0A426h；di ← ds:46B7h
al ← [bx+di]                            ; 隊伍槽表 → 這個成員的角色編號
[0xA9FD + 角色編號] ← 剛才那個 byte      ; ← 每個角色一格的「這回合要用什麼」
sub_1728C                               ; 切回地圖視窗
clc
```

⚠ **索引是角色編號不是隊伍位置。** `ds:46B7h` 是隊伍槽表（`docs/re/15`），
所以同一個角色換到別的隊伍位置，這一格還是跟著他。

## 6. 用到的字串

| 字串 | 內容 | 誰印 |
|---|---|---|
| 4 | `Use:  Item / Skill / Attribute` | `sub_13AE4` |
| 57 | `No one is within range.` | Hire |
| 63 | `Which group?` | Hire 的清單標題（`sub_12262`） |
| 64 | `You don't have anything.` | Weapon |
| 65 | `\x0B has no more clips.` | Load |
| 66 | `\x0B's weapon can't be reloaded.` | Load |

## 7. 還沒解的

- `loc_1382B`：算「可雇用的對象數」的那一段。
- `sub_19394`：物品清單選單的完整行為（回傳值的編碼）。
- `sub_13AE4` 之後三條分派（物品／技能／屬性）各自做什麼。
- `sub_12738`：與選擇合成那個 byte 的低半部。
- `ds:A9FDh` 那個 byte 的欄位配置（哪幾個 bit 是類別、哪幾個是編號）。

## 8. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/handlers.json \
  0x1963A 0x1382B 0x12262 0x16D34 0x16F20 0x19394 0x13AE4 0x12738 \
  0x15C96 0x15CA4 0x1968C 0x1789C
```

`0x1234E`–`0x1243F` 沒有被 IDA 建成函式，要用 `export_listing.py` 的逐指令 JSON 讀。

## 9. 這一輪學到的（寫成規則）

- **一組同類的處理程式，先找它們的共同形狀再讀個別內容。** 這四支加上
  已解的迴避、攻擊、逃跑，全部都是「檢查 → 選參數 → 回傳」；
  知道形狀之後，每一支只剩「檢查什麼、選什麼」兩個問題要回答。
- **「處理程式裡沒有動作」不是讀漏了。** 換武器沒換、裝填沒填——
  動作在結算階段。指令階段與結算階段分離是回合制的常見設計，
  在指令處理程式裡找動作會一直找不到。
- **同一個陣列被三支函式掃過（數、找、選），三支的邊界一致才算讀懂。**
  `0xBD` 起、stride 2、到 `0xF9` 為止在 `sub_1963A`／`sub_1968C` 都出現，
  與 `docs/re/15` 的 30 槽物品陣列三方吻合。
