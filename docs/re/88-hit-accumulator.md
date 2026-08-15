# 88：命中累加值 `sub_1B108` 的四個項全部落地

日期：2026-08-15 ｜ 接 `docs/re/20` §3（形狀）、`docs/re/21`（屬性）、`docs/re/37` §3（敵人資料）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/20` §3 讀出了 `sub_1B108` 的呼叫序列，但三個項的**語意**當時掛著：
技能編號是誰、`loc_1570F` 是哪個欄位、`sub_1B15F` 減的是什麼。
三個答案都已經散在別份筆記裡，這一份把它們接起來並寫進 remake。

---

## 1. 完整公式

```
acc  = base                          ; 40／50／60（§5）
acc += Brawling 等級 × 3             ; 技能 ID 1，寫死
acc += Agility                       ; 角色記錄 +0x12
E    = 對手行動值                    ; 敵人資料 +0x02
if 對手武器類別 == 1:  E = (E << 2) & 0xFF   ; **8-bit 位移**
else:                  acc += 5
acc -= E
acc  = min(acc, 100)
```

加減全走 `sub_19C2C`（16-bit 飽和加，借位夾 0、進位夾 `0xFFFF`）。

## 2. 技能編號寫死是 1（Brawling）

```
0x1B0F1  b001      mov  al, 1
0x1B0F3  e8d7e7    call sub_198CD      ; 技能陣列裡找 ID == al 的那一筆
0x1B0F6  e9d3eb    jmp  sub_19CCC      ; ×3
```

`al` 不是參數——它在函式第一條指令就被設成常數 1。技能 1 是 **Brawling**
（`docs/re/17` 字串 1–35 的第一個）。

**拿步槍的人命中也是加 Brawling 等級。** 武器技能只在別的地方用
（技能檢定 `docs/re/32`、傷害骰 `docs/re/45` §4），不進命中累加值。

`sub_198CD` 掃角色記錄 `+0x80`–`+0xBC`（30 × 2），找不到回 0：

```
0x198CD  b380      mov  bl, 80h
0x198CF  8b3eb546  mov  di, ds:46B5h    ; ← 那筆角色記錄
0x198D3  3a01      cmp  al, [bx+di]
0x198D5  740e      jz   short loc_198E5
0x198D7  fec3      inc  bl              ; ×2 ＝ stride 2
0x198DB  80fbbc    cmp  bl, 0BCh
0x198DE  75ef      jnz  short loc_198CF
0x198E0  b000      mov  al, 0           ; 沒學過 ＝ 0
```

推論等級：**已確認**。

## 3. `loc_1570F` ＝ Agility

`bl = 0x12`，而角色記錄的七個屬性從 `+0x0E` 起（`docs/re/21`），
`0x0E + 4` ＝ `0x12`，第 4 個屬性是 **Agility**。

`docs/re/20` 舊版把這一項列在未解清單裡，`docs/re/21` §153 已經解掉——
**兩份筆記各有一半答案，誰都沒去接另一半**。

## 4. 減的是對手的行動值，不是距離

```
0x1B117  a081cf    mov  al, ds:0CF81h
0x1B11A  e83989    call sub_13A56       ; 全域組編號 → (a, b, 組)
0x1B11D  e8d486    call sub_137F4
0x1B120  e86486    call sub_13787       ; 定址到那筆敵方記錄
0x1B123  e83900    call sub_1B15F       ; ＝ 敵人資料 +0x02（行動值欄位）
0x1B126  50        push ax
0x1B127  e86379    call sub_12A8D       ; ＝ 同一筆的 +0x05 & 0x0F（武器類別）
0x1B12A  3c01      cmp  al, 1
0x1B12C  740a      jz   short loc_1B138
0x1B12E  b005      mov  al, 5
0x1B132  e8f7ea    call sub_19C2C       ; 不是近戰 → 累加值 += 5

loc_1B138:
0x1B138  58        pop  ax
0x1B139  d0e0      shl  al, 1
0x1B13B  d0e0      shl  al, 1           ; ← **8-bit**，高位丟掉
0x1B13D  50        push ax

loc_1B13E:
0x1B13E  58        pop  ax
0x1B13F  34ff      xor  al, 0FFh
0x1B141  0401      add  al, 1           ; al ＝ −E
0x1B143  b2ff      mov  dl, 0FFh        ; 符號位 → 16-bit 減
0x1B145  e8e4ea    call sub_19C2C
```

`sub_1B15F ＝ sub_12A40(ds:CF80h) ＋ sub_12ABA`，取的是**敵人資料表 `+0x02`**
（`docs/re/37` §3.1 的行動值欄位，同一個值 × 8 就是回合行動值）。
所以這一項是「對手動得多快」，跟距離沒有關係——
`internal/play/round.go` 舊註解寫的「射程與距離懲罰」是誤名。

⚠ **`shl al` 是 8-bit**：行動值 64 以上乘 4 會繞回小數字（64 × 4 ＝ 0）。
出貨資料裡的行動值都在 0–20，所以實際遊戲碰不到，
但 remake 照抄位移寬度，不要改成 `int` 乘法。

夾值那一段：

```
0x1B148  sub_19BEC    ; ds:46BEh／46BFh 歸零
0x1B14B  ds:46BEh = 100
0x1B150  sub_19C72    ; ax ＝ 累加值；cmp ax, 46BEh；cmc
0x1B153  jnb  → retn  ; 累加值 < 100 就結束
0x1B155  sub_19BF8    ; 累加值歸零
0x1B158  累加值 += 100
```

推論等級：**已確認**（`sub_19BEC`／`sub_19BF8`／`sub_19C72` 都逐指令讀過）。

## 5. 基礎值：敵方那條已解，隊伍那條卡在一張表

| 路徑 | 基礎值怎麼來 | 狀態 |
|---|---|---|
| 敵方打隊伍（`0x1B06D`） | `ds:46D8h[目標]` ＝ **被打者這回合的指令**：迴避 5 → 60、攻擊 2 → 50、其餘 40 | **已確認**（`docs/re/38` §1） |
| 隊伍打敵方（`0x1AF40`） | `ds:711Dh[ds:CF86h + 攻擊者×4]`：回 `0xFF` 用 50、否則 60 | **那張表未解** |

迴避的指令處理程式是空的——**「迴避」的全部效果就是這個 60**（`docs/re/38` §2）。

remake 的 `Command.DefenceBase()` 早就照這張表寫好了，但 `enemyActs`
一直傳寫死的 60，所以迴避在遊戲裡沒有任何作用。這一輪接上：

```
被命中率：迴避 0.405、攻擊 0.504、其餘 0.604
```

隊伍那條仍用 60，並在 `baseHitDefault` 的註解寫明缺的是哪張表。

## 6. 兩條攻擊路徑傳的是同一組敵人

`sub_1B108` 的呼叫端只有兩個：`0x1AF75`（隊伍打敵方）與 `0x1B085`（敵方打隊伍）。
兩邊都用 `ds:CF80h`／`ds:CF81h`——**那一組敵人**。所以：

| 路徑 | `c`（累加誰的本事） | `foe`（減誰的行動值） |
|---|---|---|
| 隊伍打敵方 | 出手的隊伍成員 | 目標那組敵人 |
| 敵方打隊伍 | 被打的隊伍成員 | 出手那組敵人 |

累加的永遠是隊伍那一邊，方向靠比較符號翻轉（`docs/re/20` §2）。

## 7. remake 這一側

| 位置 | 做了什麼 |
|---|---|
| `internal/game/combat.go` | `HitChance(c, base, foe EnemyData)`，`SkillBrawling ＝ 1` |
| `internal/play/round.go` | 兩條路徑各傳那一組敵人的 `EnemyData`；敵方那條改用 `Phase.Defence(目標)` |

舊簽名的 `skillID`／`fieldValue`／`distancePenalty` 三個參數全部消失——
前兩個是常數、第三個是誤名。呼叫端原本傳的是 `w.Skill, 0, 0`，
**其中兩個零就是這一輪補上的東西**。

### 7.1 分布門檻

`TestHitChanceSpread` 拿 42 張地圖全部 125 筆敵人資料，對出廠隊伍第一個人各算一次：

```
125 筆敵人資料、17 種累加值（近戰 73 筆、夾到 100 的 0 筆）
  累加值 50：1 筆   55：20 筆   65：8 筆   68：10 筆
        72：19 筆   76：43 筆   85：1 筆   （其餘 10 種各 1–7 筆）
```

三道門檻：至少要有兩種累加值（否則公式等於沒作用）、不能全部夾在 100
或全部是 0、出貨資料裡要有近戰敵人（否則 ×4 那條路一次都沒走過）。

### 7.2 表長是標頭 `+0x31`

第一版測試掃出 **17256 筆**敵人資料，其中 8039 筆算出累加值 0。
原因是 `Block.EnemyData()` 回的是「位移到區塊尾」，沒有結尾——
後面的資料全被當成敵人。真正的筆數是標頭 `+0x31`（生成器擲 1..種類數，
`docs/re/78` §1），限縮後是 125 筆。

## 8. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/hit_acc.json 0x1B108 0x1B0F1 0x1B15F 0x198CD
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/hit_acc2.json 0x1B138 0x19C72 0x12A8D
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/hit_acc3.json 0x19BEC 0x19BF8 0x19C2C

tools/go.sh test ./internal/game/ -run TestHitChanceMeleeShiftIsEightBit -v
tools/go.sh test ./internal/game/ -run TestEvadeLowersIncomingHitRate -v
tools/go.sh test ./internal/play/ -run TestHitChanceSpread -v
```

## 9. 這一輪學到的（寫成規則）

- **未解清單要跨文件對帳。** 三個項裡有兩個早就解了（Agility 在 `docs/re/21`、
  敵人 `+0x02` 在 `docs/re/37`），只是沒有人回到 `docs/re/20` 把它們填進去。
  **每次解出一個欄位，要回頭查誰在等它**——不然筆記會維持在「各自都對、合起來仍是未解」。
- **參數列裡的常數零是待辦事項的藏身處。** `HitChance(m, base, w.Skill, 0, 0)`
  編得過、測得過、玩得動，兩個零卻是整個公式缺的那一半。
  **簽名裡出現「呼叫端全部傳 0」的參數，就是還沒接上的訊號。**
- **同一個坑一輪踩兩次：解過的東西沒接上去。** 命中的四個項有三個、
  基礎值那張對照表有一整張，全都在別份筆記裡躺著，而 remake 傳的是寫死的常數。
  **「已確認」不等於「已接上」——規格 READY 之後還要有人回來查呼叫端。**
- **回傳「到結尾」的存取器會讓量測靜靜地失真。** `EnemyData()` 沒有長度，
  掃出來的 17256 筆有一半是雜訊，而分布圖看起來完全正常。
  **量之前先確認被量的東西有沒有邊界。**
