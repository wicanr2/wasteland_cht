# 89：敵人打誰是隨機重抽，以及「倒下」與「死亡」是兩個判準

日期：2026-08-15 ｜ 接 `docs/re/20` §1.1、`docs/re/88`（命中累加值）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`CONTEXT.md` 把「戰鬥中敵人打誰的規則」列為未解，而答案早就在
`docs/re/20` §1.1 的反組譯裡——只是沒有人回頭把它接進 remake。
接上去之後戰鬥立刻卡住，逼出第二件事：`Dead` 與 `Down` 一直是同一個函式。

---

## 1. 敵人隨機挑，挑到倒下的就重抽

```
0x1B04C  sub_19D0E            ; 還有沒有人能打；沒有就跳出整條路
loc_1B054:
0x1B054  al ← ds:4653h        ; 隊伍人數
0x1B057  sub_18E41            ; roll(1..人數)
0x1B05A  ds:CF84h ← al        ; 目標編號（**1-based**）
0x1B05D  sub_172BB            ; 這個人能不能被打
0x1B060  72f2  jb short 1B054 ; CF 設 → **重抽**
```

沒有「挑血最少的」「挑前排的」「輪流」——**一律等機率隨機**，
挑到不能打的人就整個重來。重抽沒有次數上限，
靠進這條路之前的 `sub_19D0E`（`0x1B04C`）保證至少有一個人可打。

推論等級：**已確認**（六條指令，`jb` 的目標就是迴圈開頭）。

## 2. `sub_172BB` ＝ CON ≤ 0，`sub_172AE` ＝ CON ＝ 0

兩支長得很像，判的不是同一件事：

```
0x172AE  bl ← 1Eh
         al ← [角色 +0x1E]     ; CON 高位
         dec bl
         or  al, [角色 +0x1D]  ; ZF ⟺ 兩個 byte 都是 0
         retn

0x172BB  sub_19614             ; 先選中那筆角色記錄
         bl ← 1Eh
         al ← [角色 +0x1E]
         test al, al
         js  → stc             ; **高位為負 → 不能行動**
         dec bl
         or  al, [角色 +0x1D]
         jz  → stc             ; 兩個 byte 都 0 → 不能行動
         clc                   ; 其餘 → 可以
```

| 函式 | 判準 | 用在哪 |
|---|---|---|
| `sub_172AE` | CON **＝ 0** | 體力處理與角色管理（10 個呼叫端，`docs/re/35` §2） |
| `sub_172BB` | CON **≤ 0** | 下不下令（`docs/re/38` §1）、能不能被挑中（§1）、還有沒有人能打（§3） |

CON 是有號的：負值代表重傷等級（負到 −40 以下有五級傷勢，`docs/re/19`）。
**重傷倒下的人還沒死，但在戰鬥裡什麼都不能做。**

## 3. 全滅判定數的是「能行動的人數」

```
0x19D0E  ds:CD43h ← 0            ; 計數
         ds:CD44h ← 隊伍人數     ; 迴圈變數
loc_19D16:
         sub_172BB               ; CF 清 → 這個人能打
         jb  → 不計
         ds:CD43h++
         ds:CD44h−−；≠ 0 → 回 loc_19D16
         al ← ds:CD43h; test al, al   ; ZF ⟺ 一個都沒有
```

所以「全滅」的定義是 **CON ≤ 0 的人佔滿全隊**，不是「每個人 CON 都剛好是 0」。

## 4. remake 這一側

`Character.Dead()` 一直是 `CON == 0`（照 `sub_172AE` 寫的），而戰鬥的
四個判斷全部用它——行動順序、下不下令、敵人挑目標、`PartyLeft`。
CON 負值的人因此被當成好手好腳。

新增 `Character.Down()`（`CON <= 0`），戰鬥四處全部改用它；
`Tick16` 那條路留 `Dead()`（它的原版呼叫端就是 `sub_172AE`）。

### 4.1 症狀：戰鬥永遠打不完，而且沒有任何斷言會紅

接上隨機目標之後 `TestBattleRunsToCompletion` 紅了——「200 步之內戰鬥沒結束」。
逐步印出來：

```
步 19：敵人 4 活人 4 Done=false Cmd=[2 0 0 0]
步 20：敵人 3 活人 4 Done=true  Cmd=[0 0 0 0]
步 21：敵人 3 活人 4 Done=true  Cmd=[0 0 0 0]     ← 之後一直這樣
```

四個人全部 CON ≤ 0：指令階段（照 `sub_172BB` 寫的）一個人都問不到，
`Cmd` 全是 0；而 `Over()` 走 `PartyLeft()`（照 `Dead()` 寫的）認為隊伍
還有四個人。**兩邊各用一個判準，中間就出現一個誰都推不動的狀態。**

先前不會發作，是因為 remake 的敵人永遠打「第一個 `Dead()` 為 false 的人」，
而那個人 CON 一掉進負值就永遠符合這個條件——**傷害全部集中在他身上**，
其餘三人毫髮無傷，照樣輸出。目標一分散，四個人就一起滑進負值。

⚠ **這個 bug 在改動之前就在程式裡**，隨機目標只是讓它顯形。
`Dead()` 的註解甚至寫著 `sub_172AE` ——**函式引對了，用錯地方**。

### 4.2 門檻

`TestNegativeConIsDownNotDead`：CON −5 的人 `Dead()` 為 false、`Down()` 為 true，
三人隊伍（−5／20／0）的 `PartyLeft()` ＝ 1，行動順序裡不出現倒下的人。

端到端那道是既有的 `TestBattleRunsToCompletion`——它本來就會紅，
只是先前沒有東西讓它紅。

## 5. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/enemy_target.json 0x1B04C 0x172BB
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/enemy_target2.json 0x1B054 0x172D2
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/down_vs_dead.json 0x172AE 0x19D0E

tools/go.sh test ./internal/game/ -run TestNegativeConIsDownNotDead -v
tools/go.sh test ./internal/play/ -run TestBattleRunsToCompletion -v
tools/go.sh test ./internal/play/ -run TestCombatManyBattles -v
```

## 6. 這一輪學到的（寫成規則）

- **「未解」有時候只是「沒接上」。** 敵人目標選擇在 `CONTEXT.md` 掛了很久，
  而那六條指令從 `docs/re/20` 寫成的那天起就在 §1.1 裡。
  **宣告某件事未解之前，先 grep 自己的筆記。**
- **把正確行為接上去，會讓沉睡的 bug 醒過來。** `Dead`／`Down` 混用寫在
  程式裡很久，因為「敵人永遠打第一個人」剛好遮住它。
  **修好一個近似之後要重跑全部測試，紅的那個往往不是你剛改的地方。**
- **兩個判斷用不同判準，就會生出誰都推不動的狀態。** 一邊說「沒人能下令」、
  一邊說「隊伍還有四個人」——各自都自洽。
  **同一個概念只能有一個函式**，語意不同就取兩個名字，而不是共用一個。
