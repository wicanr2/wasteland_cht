# 90：隊伍的行動值，以及誰會被排進行動表

日期：2026-08-15 ｜ 接 `docs/re/36`（回合與行動順序）、`docs/re/89`

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/36` §2 解出敵人那一半，隊伍那一半在 remake 裡一直頂著 0。
兩邊排進同一張表，公式卻不是同一條。

---

## 1. 兩條公式

敵人（`0x1AE0A`，`docs/re/36` §2 已解）：

```
sub_19BF8              ; 累加器 ← 0
sub_19C84              ; 2d6，逢同點續擲
sub_19C2C              ; 累加
sub_1B15F              ; 敵人資料 +0x02
shl al,1 ×3            ; **× 8**
sub_19C2C              ; 累加
sub_1B0E2              ; 寫進 ds:7931h
```

隊伍（`0x1AE7C`）：

```
sub_19BF8              ; 累加器 ← 0
sub_19C84              ; 2d6，逢同點續擲
sub_19C2C              ; 累加
bl ← 11h
di ← ds:46B5h
al ← [bx+di]           ; **角色記錄 +0x11**
dl ← 0
sub_19C2C              ; 累加
sub_1B0F1              ; **Brawling（技能 1）等級 × 3**
sub_19C2C              ; 累加
sub_1B0E2              ; 寫進同一張表
```

```
敵人行動值 ＝ 2d6 ＋ 敵人資料 +0x02 × 8
隊伍行動值 ＝ 2d6 ＋ Speed ＋ Brawling × 3      ← 沒有 ×8
```

`+0x11` 是七個屬性裡的第 4 個。屬性從 `+0x0E` 起（`docs/re/21`），
所以 `+0x11` ＝ **Speed**——與 `loc_1570F` 的 `+0x12` ＝ Agility 相隔一格，兩邊互相印證。

`sub_1B0F1` 與命中累加值用的是同一支（`docs/re/88` §2）：`mov al, 1` 寫死，
所以**行動順序也吃 Brawling，不管拿什麼武器**。

推論等級：**已確認**（兩段都逐指令讀完，累加器的歸零與寫入點一致）。

⚠ 出貨資料的量級差很大：敵人那條乘 8（`+0x02` 實測 0–20 → 0–160），
隊伍那條不乘（Speed 加 Brawling×3，出廠角色在 15–30）。
**敵人普遍先動**，這是資料與公式共同造成的，不是哪一邊寫錯。

## 2. 只有下攻擊令的人才排進行動表

```
loc_1AE5F  al ← 2；sub_1B0DA          ; 游標先 += 2
           al ← 1                     ; 成員編號從 1 起
loc_1AE66  ds:CF84h ← al
           sub_172BB                  ; CON ≤ 0 → jb loc_1AE9B（跳過）
           bl ← ds:CF84h
           al ← [ds:46D8h + bl]       ; 這個人這回合的指令碼
0x1AE78    cmp al, 2                  ; **≠ Attack → jnz loc_1AE9B**
           …算行動值、寫進表…
loc_1AE9B  al ← 2；sub_1B0DA          ; 不論排不排，游標都 += 2
           成員++；< ds:4653h → 回 loc_1AE66
```

**迴避、換武器、使用物品、雇用、裝填的人這一回合根本不會被叫到**——
不是「輪到他但什麼都不做」，是連格子都不占。指令碼的編號見 `docs/re/38` §2
（`' ' H A W R E L U` 的索引，Attack ＝ 2）。

跳過的人游標照樣前進，所以**表的格子與成員編號是固定對應的**，
跟敵人那邊「三組 × 10 格 × 2 bytes」是同一種固定版面。

回合結束後（`0x1AEEB`–`0x1AF0A`）把下過攻擊令的人的指令碼與參數清 0——
指令是「這一回合」的，不是持續狀態。

## 3. remake 這一側

| 位置 | 改了什麼 |
|---|---|
| `internal/game/rounds.go` | `BeginRound(acting func(member int) bool)`：兩條公式各自寫在 game 層，不再由呼叫者傳「一個欄位」 |
| `internal/play/round.go` | `speedOf` 整支刪掉，改成 `acting`（查 `Phase.Cmd[i] == CmdAttack`） |

舊簽名 `BeginRound(speedOf func(Combatant) int)` 把行動值寫成
「2d6 ＋ 回呼 × 8」，隊伍那邊回 0——**兩條公式被硬塞成一條**，
而 game 層其實兩邊的欄位都拿得到，回呼從一開始就不需要。

### 3.1 門檻

```
tools/go.sh test ./internal/game/ -run TestInitiativeFormulasDiffer -v
    隊伍 18–43（底 15）；敵人 27–62（底 24）
```

檢查三件事：隊伍的下界 ＝ `Speed ＋ Brawling×3 ＋ 2`、敵人的下界 ＝ `+0x02 × 8 ＋ 2`、
**隊伍的下界不得大到看起來像被乘了 8**。第三條是反向門檻：
把隊伍那條誤乘 8 不會讓任何既有斷言紅，戰鬥照樣打得完，只有先後順序整個偏掉。

`TestRoundOrderIsStable` 加了反向的一半：沒人下攻擊令時行動表裡只剩敵人。

## 4. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/order.json 0x1AE0A
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/order3.json 0x1AE5F 0x1AD8A

tools/go.sh test ./internal/game/ -run TestInitiativeFormulasDiffer -v
tools/go.sh test ./internal/play/ -run TestRoundOrderIsStable -v
```

## 5. 這一輪學到的（寫成規則）

- **「同一張表」不等於「同一條公式」。** 敵人與隊伍排進同一個 `ds:7931h`，
  舊實作因此把兩邊寫成一條式子加一個回呼。
  **共用資料結構不保證共用算法**——兩邊的寫入點要各讀一次。
- **參數化不是保留未解的好辦法。** 「那個欄位還沒對上，所以收進來當參數」
  看起來很保守，實際上它把「隊伍那邊回 0」凍結成了行為，
  而且讓呼叫端看起來像已經接好了（`docs/re/88` §9 的同一條規則）。
- **dump 起點落在指令中間會生出假指令。** 這一輪 `0x1AE60`／`0x1AE70`／`0x1AE80`
  各解出一條不存在的指令（`add ch, al`、`test cl, bh`、`add ch, dh`），
  因為前一條是 3–4 bytes。**16-bit 反組譯要從跳躍目標開始 dump**，
  不要從整齊的位址開始。
