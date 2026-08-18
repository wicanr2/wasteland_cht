# 101：`ds:711Dh` 是敵方這一回合的移動計畫表

日期：2026-08-16 ｜ 接 `docs/re/20`（戰鬥結算）、`docs/re/88`（命中累加器）、
`docs/re/87`（敵人在地圖上移動）、`docs/re/37`（敵方記錄）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

隊伍攻擊敵方的命中基礎值查一張表（`0x1AF64`）：回 `0xFF` 用 50、否則 60。
`docs/re/20` §6 與 `docs/re/88` §5 把那張表記成「未解」，
`docs/re/87` §1 只確認了它在映像裡靜態全 0、是執行期填的。

那張表是**每一筆遭遇這一回合打算怎麼移動**。填它的是 `sub_14BF0`。

---

## 1. 全檔只有四處碰它

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_range_refs.py \
  workplace/analysis/dumps/table711d.json 0x711D 0x7200
```

掃過 20,177 條指令，範圍 `[0x711D, 0x7200)` 共 31 筆命中，
其中運算元值**正好是 `0x711D`** 的只有四處（其餘落在 `0x7131`、`0x71A9`、
`0x71B7`、`0x71CD` 這幾張相鄰但無關的表）：

| 位址 | 函式 | 指令 | 角色 |
|---|---|---|---|
| `0x14BF4` | `sub_14BF0` | `mov [bx+711Dh], al` | 清成 `0xFF`（初始化迴圈）|
| `0x14DFC` | `sub_14BF0` | `mov [bx+711Dh], al` | **唯一的實際寫入** |
| `0x15050` | `sub_15036` | `mov al, [bx+711Dh]` | 讀出來執行移動（`docs/re/87`）|
| `0x1AF64` | （不在函式內）| `mov al, [bx+711Dh]` | **命中基礎值 50／60** |

⚠ 這個掃描比對的是**運算元數值**，所以「先把位址載進暫存器再間接存取」
（`mov di, 711Dh` ／ 更大範圍的整段清零）不會出現在上面。
`mov di, 711Dh` 這一形式仍會被掃到（值就是 `0x711D`），全檔沒有；
整段清零則沒有排除，但 `sub_14BF0` 開頭自己就清了 16 格，
所以不需要別人代勞。

推論等級：**已確認**。

## 2. 表長 16 格，一格對一筆敵方記錄

`sub_14BF0` 的開頭：

```
0x14BF0  mov  bl, 0
0x14BF2  mov  al, 0FFh
loc_14BF4:
0x14BF4  mov  [bx+711Dh], al
0x14BF8  inc  bl
0x14BFA  cmp  bl, 10h            ; 16 格
0x14BFD  jnz  short loc_14BF4
```

**16 格，每一回合進來先全部清成 `0xFF`。**

寫入端的索引怎麼算（`0x14DEE`）：

```
0x14DEA  al ← ds:0A63Ch          ; 這一輪決定的步向（0–8），沒決定就不會走到這裡
0x14DED  push ax
0x14DEE  al ← ds:0A63Ah          ; 隊伍組
0x14DF1  shl  al, 1
0x14DF3  shl  al, 1              ; ×4
0x14DF5  add  al, ds:0A639h      ; ＋ 這一組面對的第幾筆遭遇
0x14DF9  bl ← al
0x14DFB  pop  ax
0x14DFC  mov  [bx+711Dh], al
```

`ds:A63Ah`（外層 0 起、上界 `ds:4657h`）與 `ds:A639h`（內層 0–3）這一對，
正是 `sub_137F4(al, dl)` 的兩個參數——
`ds:46C8h ← 0x6B31 + 0x178 × al + 0x5E × dl`（`docs/re/37` §1）。
`0x5E` ＝ 94 ＝ 一筆敵方記錄，`0x178` ＝ 4 × 94。
所以 16 筆敵方記錄本身就是 4 × 4 的排法，而 `ds:711Dh` 是**與它同索引的平行陣列**：

> **`ds:711Dh[記錄編號]`，記錄編號 ＝ 隊伍組 × 4 ＋ 遭遇序號，共 16 筆。**

`sub_14F5D` 是這一對索引的定址捷徑（`sub_137F4(ds:A63Ah, ds:A639h)`），
`sub_14BF0` 全程用它。

推論等級：**已確認**（初始化迴圈、索引算式與 `sub_137F4` 的參數三處互相對上）。

## 3. 讀取端的索引繞了一圈，繞回同一個編號

命中判定那一側（`0x1AF27`–`0x1AF64`）看起來在算別的東西：

```
0x1AF27  bl ← ds:0CF84h          ; 攻擊者（隊伍成員槽）
0x1AF2B  call sub_13449          ; → al ＝ 目標編號
0x1AF2E  ds:0CF81h ← al
...
0x1AF4C  al ← ds:0CF81h
0x1AF4F  call sub_13A56          ; ← 拆成三個數
0x1AF52  ds:0CF80h ← bl
0x1AF56  ds:0CF86h ← dl
0x1AF5A  shl  al, 1
0x1AF5C  shl  al, 1              ; ×4
0x1AF5E  add  al, ds:0CF86h
0x1AF62  bl ← al
0x1AF64  al ← [bx+711Dh]
0x1AF68  cmp  al, 0FFh
0x1AF6A  jz   loc_1AF71
0x1AF6C  al ← 3Ch                ; 60
0x1AF6E  jmp  loc_1AF73
loc_1AF71:
0x1AF71  al ← 32h                ; 50
```

`sub_13A56(n)` 是一支除以 3 的拆解：

```
0x13A56  dl ← 0FFh
loc_13A58:
0x13A58  inc  dl
0x13A5A  sub  al, 3
0x13A5C  cmc
0x13A5D  jb   short loc_13A58    ; al ≥ 3 就繼續
0x13A5F  add  al, 3
0x13A61  bl ← al                 ; ＝ n mod 3   → ds:CF80h（記錄裡的第幾個 30-byte 子組）
0x13A63  al ← dl                 ; ＝ n div 3
0x13A65  push ax
0x13A66  and  al, 3
0x13A68  dl ← al                 ; ＝ (n div 3) & 3
0x13A6A  pop  ax
0x13A6B  shr  al, 1
0x13A6D  shr  al, 1              ; ＝ (n div 3) >> 2
```

於是 `0x1AF5A` 的 `al × 4 + dl` ＝ `((q >> 2) << 2) | (q & 3)` ＝ `q`（`q < 16`）。
**繞一圈之後就是記錄編號本身**，與 §2 的寫入索引同一個東西。
`n mod 3` 那一半是記錄裡的第幾個 30-byte 子組（`docs/re/37` §2.2）——
原版打的是「一群」不是「一隻」。

推論等級：**已確認**。

## 4. 值的版面：高 2 位是訊息，低 6 位是步向

三個寫入點只差在加了什麼：

| 來源 | 加值 | `docs/re/87` 讀出來的訊息（`ds:A643h`，字串 72／74／75）|
|---|---|---|
| `0x14DEA` 直接落下 | ＋0 | `" move\ns\n\n to a better position."` |
| `0x14E08`（`sub_14E45` 之後）| `add al, 40h` | `" run\ns\n\n away."` |
| `0x14E37`（`sub_14E4A` 之後）| `add al, 80h` | `" run\ns\n\n at you."` |

`sub_15036` 拆回去的方式（`docs/re/87` §1）完全吻合：
`ds:A635h ← al & 0x3F`（步向）、`ds:A634h ← al >> 6`（訊息索引 0–2）。

低 6 位是 0–8 的步向索引：`sub_14BF0` 的搜尋迴圈跑 `ds:A63Dh` 0–8
（`0x14DC0` 的 `cmp dl, 9`），成功的那一個寫進 `ds:A63Ch`（`0x14DB7`）。
`[si-55F2h]` ＝ `ds:AA0Eh` 是 `00 01 …08` 的恆等表；
`sub_164ED` 再用它查 `ds:AAB1h` 的 9 個近指標，第 4 格（正中央）指向
`0x18016`，其餘八格指向 `0x16506`–`0x16517` 的八支短常式——
**3 × 3 鄰域，中央 ＝ 不動**。

`sub_14E45`（`al ← 1`）與 `sub_14E4A`（`al ← 0`）共用 `loc_14E4C`：
先算隊伍與這一筆遭遇的座標差的正負號，`al ≠ 0` 那一支再把兩個正負號取反。
所以 `sub_14E4A` ＝ 朝隊伍走、`sub_14E45` ＝ 往反方向走，
與「run at you」「run away」對得上。

> | 值 | 意思 |
> |---|---|
> | `0xFF` | **這一回合這一筆遭遇不移動** |
> | `0x00`–`0x08` | 換一個位置（`" moves to a better position."`）|
> | `0x40`–`0x48` | 逃跑 |
> | `0x80`–`0x88` | 朝隊伍衝 |

推論等級：**已確認**（三個寫入點、拆解點與三句訊息互相對上）。

## 5. 所以命中基礎值的意思是

> **這一回合會移動的敵人比較好打（60），留在原地的比較難打（50）。**

判定方向不要記反：隊伍攻擊那條是 `roll(1..100) < 累加值` 才命中
（`docs/re/20` §1.2），所以**累加值越大越容易打中**。

`sub_14BF0` 由遭遇驅動器 `sub_11CD0` 在 `0x11DB5` 呼叫，
位置在隊伍下指令（`sub_11F76`，`0x11DD8`）**之前**、
執行移動（`sub_15036`，`0x11E3C`）之後的下一輪之前——
也就是**每一個戰鬥回合重算一次**，不是進戰鬥時算一次。

推論等級：**已確認**（呼叫點與迴圈結構見 `docs/re/94` §2）。

## 6. 走哪一條分支：三條的判定條件

`sub_14BF0` 對 `ds:A5AEh` 的三個位移（`+0x04`／`+0x06`／`+0x08`）逐一取出
這一筆遭遇的三組敵人，每一組跑一次下面這段：

```
0x14C55  bl ← 9；al ← [ds:46C6h + 9]；and al, 4
         ≠ 0 → loc_14D31 → loc_14E3B    ; 這一筆遭遇整個不移動
0x14C7A  sub_12A40(組)                  ; 選定這一組的敵人資料
0x14C7D  sub_12A8D  → 資料 +0x05 & 0x0F ; **武器類別**（`docs/re/37` §3.2）
0x14C80  sub_19D2F  → ds:CD00h 有沒有這個類別
0x14C83  jnb → loc_14CCA                ; 有 ＝ **有射程** → 「換位置」
0x14C85  loc_12AA2  → 資料 +0x06        ; **敵人種類**（`docs/re/37` §3.2）
0x14C88  ＝ 1（Animal）→ loc_14CA5
0x14C8C  否則跑士氣檢查，jb → loc_14CA5，否則 → 0x14E03（逃跑）
loc_14CA5:
0x14CB1  al ← [ds:46C8h + 3]            ; 記錄標頭 +3 ＝ 與隊伍的距離
0x14CB7  cmp al, 10h；cmc；jnb → 不動     ; **距離 < 16 → 不動**
0x14CBC  否則 → 0x14E0C                 ; → 朝隊伍衝
```

`cmc` ＋ `jnb` 是這份執行檔到處在用的「小於」慣用法：
`cmp` 借位設 CF，`cmc` 反過來，`jnb`（CF ＝ 0）於是等於「小於」。
同一個形狀在 `0x14C3E` 的組迴圈上界判斷再出現一次，兩處互為佐證。

### 6.1 `ds:CD00h` ＝ 有射程的武器類別

`sub_19D2F(al)` 從 `ds:CD00h` 逐 byte 比對，遇到負值就停：

```
ds:CD00h  0d 0a 0b 0c 02 03 04 05 06 07 08 09 ff
```

12 個值 ＝ 類別 2–13，**少的正好是 0 與 1**（徒手與近戰，`docs/re/45` §3.1）。
所以第一個分岔是：**手上有射程武器的敵人走「換位置」，只有近戰的往下走。**

### 6.2 士氣：整組剩下的血量低於一隻的基礎血量的一半就跑

`0x14C8C` 那四行拆開來是：

```
sub_14FDE   ; ds:46BEh ← 這一組十隻的 16-bit 血量總和（+0x00…+0x13，`docs/re/37` §2）
sub_12AAB   ; dl:al ← 敵人資料 +0x00/+0x01 ＝ 這一型的基礎血量
sub_19BFC   ; ds:46C0h ← 上面那個 16-bit
shr ds:46C1h,1 / rcr ds:46C0h,1   ; ÷ 2
sub_19C69   ; CF ← (ds:46BEh ≥ ds:46C0h)
jb → loc_14CA5                     ; 還撐得住 → 逼近或不動
否則 → 0x14E03                     ; **逃跑**
```

`sub_14FDE` 的迴圈從 `bl = 0x13` 每次減 2 走到 0，正好是十筆 16-bit 血量；
`sub_19C04` 是 16-bit 累加，`sub_19BEC` 在迴圈前把累加器清零。

**種類 1（Animal）不做這個檢查**——動物不會逃。

### 6.3 「換位置」那一條也分種類

```
loc_14CCA  al ← 資料 +0x06（種類）
  ＝ 4（Cyborg）→ sub_14F7A(1)
  ＝ 5（Robot） → 常數 0Fh
  否則          → sub_14F7A(2)
… ds:A63Eh ← 常數（Robot 0Fh／Cyborg 0Bh／其餘 7）
… ds:46C0h ← sub_156CB(距離, 組)
… ds:46BEh ← 敵人的行動值（sub_12ABA，`docs/re/36` §2）＋ 上面那個常數
… sub_19C69；jnb → loc_14D34（真的換位置），否則 loc_14E3B（不動）
```

`sub_14F7A` 回 CF 時直接跳 `loc_14CA2` ＝ 逃跑。

推論等級：**已確認**（`sub_19D2F`、`loc_12AA2`、`sub_14FDE`、`sub_19BFC`／
`sub_19C04`／`sub_19C69` 逐支讀完，資料端的兩個欄位在 `docs/re/37` §3.2 已有 42 個區塊
397 筆的值域佐證）。

### 6.4 `ds:46C6h + 9` 的 bit2 是資料裡的旗標

同一個 byte 的 bit1 是「這一組可以雇用」、高 4 位是 NPC 記錄編號
（`docs/re/110` §2）。bit2 ＝ **這一筆遭遇不參與移動**。
它在出貨資料裡，不是執行期算出來的。

## 7. 對 remake 的意思：基礎值是 50 不是 60

remake 沒有實作敵人在地圖上移動那一段（`docs/re/87` §2），
所以**每一格都會是 `0xFF`**，照原版的規則基礎值就是 **50**。

`internal/play/round.go` 原本寫死 60，理由是「表未解，先用『否則』那一支」。
表解出來之後那個理由不成立了：沒有移動計畫的敵人走的是 `0xFF` 那一支。

接法是把計畫本身放進 `game.Battle`（`MovePlan`，`NoMovePlan` ＝ `-1`），
命中基礎值由 `game.HitBase(plan)` 算——
**不是把常數從 60 改成 50**。差別在於實作了地圖移動之後，
60 那一支會自己亮起來，不必有人記得回來改。

⚠ 這是一個玩得出來的差異：命中率整體降 10 個百分點，戰鬥會變長。
它是照著解出來的機制走的結果，不是調數值。

§6 的三條分支**目前接不上去**：它們決定的是敵人在地圖上往哪一格走，
而重製版沒有實作戰鬥中的地圖移動（`docs/re/87` §2）。三條分支的判定條件
與門檻在 §6 已經寫全，實作地圖移動的那一輪照著接即可；
在那之前計畫表每一格都是 `0xFF`，基礎值 50 這一支不受影響。

## 8. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_range_refs.py \
  workplace/analysis/dumps/table711d.json 0x711D 0x7200

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/f14bf0.json 0x14BF0 --callers

python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xAA0E 9 --bytes
python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xAAB1 9

# §6 的兩張表
python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xCD00 13 --bytes
python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xA5AE 6 --bytes
```

`0x1AF22`–`0x1AFA0`、`0x13A56`、`0x14E45`／`0x14E4A`、`0x164ED`、`0x14F5D`
不必再進 IDA——全檔逐指令的 JSON 已經在
`workplace/analysis/dumps/listing.json`（`tools/ida/export_listing.py` 產）。

## 9. 這一輪學到的（寫成規則）

- **「這張表未解」常常只是「還沒問誰寫它」。** 這一格卡了三份筆記
  （`20`、`87`、`88`），而解開它用的是一次範圍掃描 ＋ 一次函式匯出。
  **看到「表是執行期填的」就接著問寫入點，不要停在那一句。**
- **索引算式繞遠路不代表語意不同。** 讀取端 `al × 4 + dl` 看起來與寫入端
  無關，拆開 `sub_13A56` 之後才看出它是 `(q >> 2) << 2 | (q & 3)` ＝ `q`。
  **兩端的索引要各自化簡到最簡形再比，不要比表面形狀。**
- **常數寫死的註解要連「為什麼選這一支」一起寫。** 原本的註解寫著
  「表還沒解，所以用『否則』那一支的 60」——理由留著，所以表一解出來
  就知道該回頭改哪裡。**沒寫理由的暫代值，解出來的那天沒有人找得到它。**
- **`cmc` ＋ `jnb` 是「小於」。** 這份執行檔的比較幾乎都長這樣，
  照字面把 `jnb` 讀成「大於等於」會把每一個邊界判斷讀反。
  **拿同一份執行檔裡語意已知的另一處對照一次**，比查指令手冊快。
