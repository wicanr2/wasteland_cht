# 92：`USE` 指令 —— Skill／Item／Attribute 三選一，與施用的骨架

日期：2026-08-15 ｜ 接 `docs/re/91`（指令列的七個入口）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`USE` 是指令列七項裡最深的一支，也是**擋住「走完主線」驗收**的那一項——
很多劇情觸發要用技能或物品。這一份解開它的第一層與三條分支的入口。

---

## 1. 挑人：`0x13A80`

```
0x13A80  sub_19727；sub_1728C
0x13A86  al ← ds:4653h（隊伍人數）
0x13A89  ＝ 1 → loc_13AAF          ; **一個人就不問**，直接用他
0x13A8D  al ← 2；sub_16CB2         ; 印訊息 2（`Which player?`）
0x13A92  al ← 0；sub_1721B         ; 挑隊員（1–9）
0x13A97  ＝ 0 → jmp sub_163C4      ; 取消 → 回地圖
loc_13AAF:
0x13AAF  ds:A5E5h ← al             ; 記住是誰
0x13AB2  sub_172BB                 ; CON ≤ 0（docs/re/89 §2）
0x13AB5  → 不行就印訊息 3 回地圖
loc_13AC2:
0x13AC2  sub_13AE4（挑要用什麼）
0x13AC8  失敗 → 回頭重問
0x13ACB  ds:4661h ← ds:A5E5h
0x13AD2  sub_13C58（施用）
```

## 2. 三選一：字母表 `ds:A5E8h`

```
ds:A5E8h = 53 49 41 00 …   ＝ "SIA\0"
```

| 索引 | 鍵 | 是什麼 | 分支 |
|---:|---|---|---|
| 0 | `S` | **Skill**（技能） | `loc_13B29` |
| 1 | `I` | **Item**（物品） | `loc_13B24` → `loc_13C0B` |
| 2 | `A` | **Attribute**（屬性） | `loc_13BDE` |

`sub_13AE4` 的開頭：

```
0x13AE4  ds:A5E1h ← al            ; 隊員編號
0x13AE7  sub_13C52
0x13AEA  al ← ds:472Ch
         ≠ 0 → al ← 0，跳過選單直接走技能那條   ; ← **旗標**
loc_13AF6:
0x13AF6  al ← 4；sub_16CB2        ; 印訊息 4（問要用什麼）
0x13AFB  ds:4680h ← 0A5E8h        ; ↑ 那張字母表
0x13B01  ds:7DF3h ← 0x404         ; 熱區遮罩（docs/re/43）
0x13B0D  sub_173B0                ; 等按鍵 → 字母表索引
0x13B10  js → stc; retn           ; ESC ＝ 取消
loc_13B12:
0x13B12  ds:A5E3h ← al            ; 存起來，之後 sub_13C58 要看
         0 → loc_13B29／1 → loc_13C0B／其餘 → loc_13BDE
```

⚠ **`ds:472Ch` 非 0 時不問，直接當成選了技能**。那個旗標是什麼還沒追——
它在 `sub_13AE4` 裡被讀了三次（`0x13AEA`、`0x13B2F`、`0x13B6E`），
形狀像「這次呼叫是從別的地方進來的」（`sub_13AE4` 的另一個呼叫端是 `0x1240E`）。

## 3. 技能那條：清單來自角色記錄

```
loc_13B29:
  al ← ds:A5E1h；sub_19614          ; 選中那個角色
  ds:472Ch ≠ 0 → 直接取 +0x80 那一格，是 0 就放棄
  否則:
    ds:7DF3h ← 0x434
    sub_198F0                        ; 清單選擇（回索引）
    ＝ 0FFh → 整個重問
    ≥ 0FDh → 回 loc_13B29
loc_13B5E:
  al <<= 1；al += 0x7Eh              ; 索引 → 記錄位移
  al ← [角色 + 那個位移]              ; **技能編號**
  ds:A5E4h ← al
```

`al × 2 + 0x7E` 是技能陣列的定址：索引 1 → `+0x80`、索引 2 → `+0x82`……
與 `docs/re/15` 的「`+0x80` 起 30 × 2」一致（清單的索引從 1 起算）。

推論等級：**已確認**（三選一的字母表是資料裡直接讀出來的；
技能陣列的定址與既有筆記互相印證）。

## 4. 施用：`sub_13C58` 的骨架

介面是三個暫存器：`al` ＝ 選項（0 Skill／1 Item／2 Attribute）、
`bl` ＝ 選中的編號、`dl` ＝ 另一個參數（語意未解）。

```
0x13C58  ds:A5E0h ← al        ; 選項
         ds:A5D9h ← dl
         ds:A5DBh ← bl        ; 技能／物品／屬性的編號
         ds:A5DCh ← ds:4661h  ; 誰在用
         sub_19614            ; 選中那個角色

         al ＝ 2 → loc_13CB6（屬性）
         al ＝ 1 → loc_13CD0（物品）
         其餘   → 技能那條
```

三條分支的形狀一樣，只差**目標判定**那一支：

| 分支 | 前置 | 目標判定 |
|---|---|---|
| 技能（0） | `sub_13E32` → `sub_13E13` | **`sub_140DD`** |
| 物品（1） | `sub_13E32` → `sub_13E18` | **`sub_14090`** |
| 屬性（2） | `sub_13E32` → `sub_13E1D` | **`sub_14126`** |

三條都先過 `sub_13E3D`（失敗 → `loc_13D03`：`al ← 1; stc; retn`），
然後匯合到 `loc_13CE8`：

```
loc_13CE8:
  ds:A5DAh ← bl
  sub_13E85
  sub_14175
  sub_142E2；and al, 8
    ≠ 0 → sub_142B1(bl ＝ ds:A5DAh)
```

### 4.1 三個特例

```
技能 0x19（25 ＝ Medic）或 0x20（32 ＝ Doctor） → loc_13D82
技能 9（Perception）且 sub_140DD 回的值 ≠ 0     → sub_17A6B → sub_13E58
```

技能編號對照 `docs/re/17` 的字串 1–35。**治療與察覺自成一條路**，
不走共同的效果套用——這與它們在遊戲裡的作用一致（一個改 CON、一個給訊息）。

推論等級：**強證據**（分支結構與呼叫序列逐條讀完；三支目標判定的內容還沒讀，
「哪一支對應哪一種」是從分支位置推的）。

### 4.2 收尾：兩條路都會改寫腳下那一格

```
loc_13D18（成功）：
  al ← bl                       ; 位移：預設 4；記錄 +0x00 的 bit3 設起來時
  call sub_13E7F                ; 由 sub_142B1 依「通過的是哪一條條件」算
  call sub_17CFF                ; ← 改寫地圖格
  CF ＝ 0 → sub_1652D(方向)     ; 只捲動不移動
  CF ＝ 1 → sub_1651A(方向)     ; 走一步

loc_13D3F（沒有吻合的條件）：
  記錄 +0x08 非 0 → 印 +0x03、套懲罰（全隊 sub_14296 或單人 sub_14193）
  call sub_13E7F
  al ← 6；call sub_17CFF        ; ← 用位移 6 改寫
```

⚠ **「沒有吻合的條件」不是什麼都不做。** 多數記錄的 `+0x06`／`+0x07` 是
`0xFF 0xFF`（不改），所以看起來像沒事——但那是資料的選擇，不是機制。

remake 這一側以前只印訊息、不改寫，於是**鑰匙插進去會說 `It works!`，
但門後面什麼都不會變**。科奇斯基地的四根圓柱就卡在這裡
（`docs/re/100` §5）。已改：`Party.UseOn`。

## 5. 還沒解的

- 匯合段的 `sub_13E85`／`sub_14175`／`sub_142E2`／`sub_142B1`。
- `loc_13D82`（Medic／Doctor）與 `loc_13D3F`／`loc_13D16` 幾條收尾。
- `ds:472Ch` 旗標的語意。

（三支目標判定的內容在 §6；`sub_13AE4` 的另一個呼叫端 `0x1240E` 是
**戰鬥的 `Use` 指令**，見 [`108`](108-combat-use-and-hire.md) §1；
`sub_19727` 是戰鬥面板的邊界設定，見 [`105`](105-enc-empty-round-and-menu-region.md) §2。）

## 6. 三支目標判定：USE 查的就是條件閘表

三支的形狀一模一樣，只差比對的型別：

```
sub_140DD(al ＝ 技能編號)   ; 型別 0
sub_14090(bl ＝ 物品槽號)   ; 型別 1（先把槽號換成物品 ID 再比）
sub_14126(al ＝ 記錄位移)   ; 型別 2

  bl ← 0x0A                        ; ← **地圖記錄 +0x0A**
loc:
  al ← [ds:46AEh + bl]
  ＝ 0FFh → stc; retn              ; 表尾 ＝ 沒有吻合的條件
  sub_1417F                        ; ＝ (byte >> 1) >> 4 ＝ **高 3 位 ＝ 型別**
  型別不對 → 下一筆（bl += 2）
  bl++；[ds:46AEh + bl] ＝ 那個編號 → 命中
命中:
  sub_1418A                        ; ＝ byte & 0x1F ＝ **低 5 位 ＝ 難度**
  技能 → sub_180F0(技能, 難度)      ; 技能檢定（docs/re/32 §3）
  物品 → sub_19A58(槽號)            ; 消耗一次（docs/re/32 §6）
  屬性 → sub_1820C(難度, 位移)      ; 屬性檢定（docs/re/32 §4）
```

**那張表就是條件閘串列**（`docs/re/32` §6、`docs/spec/07` §4）——
走上去自動評估與 `USE` 主動指定用的是同一份資料，差別在：

| | 掃法 |
|---|---|
| 走上去（`Eval`） | 逐條試到**成功**為止 |
| `USE`（`UseGate`） | 只試**型別與編號都吻合**的那一條 |

⚠ **物品那條只認型別 1**（`sub_14090` 的 `cmp al, 1`），
而自動評估把 1／5／6／7 都當成找物品。照原版保留這個差別。

## 7. remake 這一側

**已接**：規則層 `Party.UseGate`（`internal/game/gates.go`）＋
play 層的三層選單（`internal/play/use.go`）。三條分支一起接，
缺一條就會變成「按了沒反應」——那是最難查的一種半成品（`docs/re/79` §5）。

清單的呈現是重製決策（原版用清單框架 ＋ 上下鍵，這裡用數字鍵），
判定那一層照原版。**超過九項還沒分頁**。

⚠ 名字要查對表：技能與物品名在 `ds:B270h`（第 **2** 張），
`exeString` 走的是第 1 張——拿錯會安靜地取到別的句子
（技能 1 在那張是 `Radio?YesNo`）。物品的索引是 **ID ＋ 36**，
寫成 35 會整批位移一格而畫面看起來完全正常。

## 8. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/use_full.json 0x13AE4 --callers
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/use_apply.json 0x13C58
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/use_targets.json 0x140DD 0x14090 0x14126

tools/go.sh test ./internal/game/ -run TestUseGateMatchesOnlyTheRightEntry -v

python3 -c "
d=open('workplace/analysis/unpacked/wl.merged.exe','rb').read()
off=0x1CE20+0xA5E8-0x10000+0xb0
print(d[off:off+8])"
```

## 9. 這一輪學到的（寫成規則）

- **選單的字母表就是規格書。** 三個 byte `53 49 41` 一次講完了 `USE` 有幾條路、
  分別是什麼——比追三條分支的程式碼快得多。
  **遇到「等按鍵 → 查表 → 分支」的形狀，先把那張表倒出來。**
- **顯示順序不是索引，第二次踩同一個坑。** `docs/re/41` 寫著「字母表在 `ds:A5E8h`」
  卻沒把它倒出來，直接照字串 4 的顯示文字（Item／Skill／Attribute）編號——
  而字母表是 `SIA`。**看到「字母表在某處」就當場把它讀出來**，
  不要用旁邊的顯示文字代替（`docs/re/38` §2 已經記過一次同型的教訓）。
- **兩個機制用同一份資料時，差別在掃法不在格式。** `USE` 與走上去自動評估
  查的是同一張條件閘表，一個「找吻合的那條」、一個「逐條試到成功」。
  **先問「是不是同一份資料」，再問「哪裡不一樣」。**
- **`export_function.py` 給的是完整 listing，`export_forced.py` 只給一個 chunk。**
  這一輪先用 forced 拿到零散片段，換成 function 之後一次拿到 366 bytes。
  **函式已經被 IDA 認出來時用前者，沒被認出來（像 `loc_1B8AD`）才用後者。**
