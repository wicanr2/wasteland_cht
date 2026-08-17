# 105：`ENC` 在空地上也能跑一回合，以及選單用的是另一塊區域

日期：2026-08-17 ｜ 接 `docs/re/40`（戰鬥畫面）、`docs/re/94`（`ENC`）、
`docs/re/47`（DOSBox oracle）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
截圖 `workplace/dosbox/shots/21-enc.png`、`24-encyes.png`。

兩件事都是**實機對拍抓出來的**，而兩件事的程式碼都早就在 `.asm` 裡。

---

## 1. `ENC` 在沒有敵人時會問一句，答 Y 照樣進指令階段

按 `E` 之後畫面（`21-enc.png`）：

```
This party is not being attacked. Do you want them to execute a battle
round?

Yes
No
```

`docs/re/94` §2 只記了另一條路（字串 `0x36`「別組不在這張地圖上」）。
這一條是字串 **`0x14`**，在 `sub_11F76` 裡（**每一組的指令階段開頭**）：

```
0x11F76  al ← ds:4654h；sub_13924        ; 選這一組
0x11F7C  bl ← ds:4653h（人數）
loc_11F82: [ds:46D8h + bl] ← 0          ; 清掉每個人這一回合的指令
0x11F8C  sub_19D0E；jnz loc_11FB3        ; 這一組還有人站著嗎
0x11F91  clc；retn                       ; 全倒 → 跳過這一組
loc_11FB3:
0x11FB3  dl ← 3                          ; 掃三個敵方子組
loc_11FB5:
0x11FB9  sub_137F4(ds:4654h)             ; 定址敵方記錄
0x11FC5  al ← [ds:46C8h + 0]
0x11FC7  test al, al；jnz loc_11FEF      ; **非 0 ＝ 有敵人** → 直接進指令階段
0x11FCB  dl−−；jns loc_11FB5             ; 下一個子組
0x11FD9  al ← 14h；sub_16CB2             ; "This party is not being attacked. "
0x11FDE  al ← 4Ch；sub_16CB2             ; "Do you want them to execute a battle round?"
0x11FE3  sub_12619                       ; 等 Y／N
0x11FEB  jnb loc_11FEF                   ; Y → 進指令階段
0x11FED  clc；retn                       ; N → 跳過這一組
```

答 `Y` 之後（`24-encyes.png`）畫面切成戰鬥模式，開始逐人問指令。
**所以「在空地上按 `E`」是原版讓玩家換武器、裝填、用道具的正規入口**——
那些都是戰鬥指令，地圖上沒有對應的按鍵。

推論等級：**已確認**（逐指令讀完 ＋ 兩張實機截圖）。

⚠ **兩條路要分清楚**：`0x14` 問的是「**這一組**沒有敵人在打」，
`0x36` 問的是「**別組**不在這張地圖上」。remake 原本只接了後者，
前者回一句 `Nothing to fight here.`——而那**不是原版的字串**，
是重製版自己寫的。

## 2. 指令選單用的是比訊息視窗高的另一塊區域

`docs/re/40` §5 把「選單的 9 行怎麼塞進 6 行的訊息視窗」列為未解，
並且已經指出線索：`sub_19727` 設的是 `ds:4674h ← 0x0E` 而不是訊息視窗的 `0x12`。

截圖把它定死了。`24-encyes.png` 的訊息區：

```
Hell Razor, choose:
Run
Use
Hire
Evade
Attack
Weapon
Load/Unjam
```

八行**全部看得到**，區域在畫面右上、肖像框的右邊。與 `sub_19727` 的設定對上：

| 設定 | 值 | 意思 |
|---|---|---|
| `ds:4674h ← 0x0E`，然後 `inc` | 15 | 左邊界 ＝ 字元欄 15 |
| `ds:4675h ← 0x27`，然後 `dec` | 38 | 右邊界 ＝ 字元欄 38 |
| `ds:4677h ← 0x70` | 112 | 下邊界 ＝ 像素列 112 ＝ 字元列 14 |
| `sub_19812(0x27, 0x0D)` | 39 × 13 | 第一次進來清掉的範圍 |

> **戰鬥的指令選單畫在字元欄 15–38、列 1–13**（13 列 × 24 欄），
> 不是訊息視窗那 6 列。八行綽綽有餘，`docs/re/40` §5 的問題不存在。

左邊那一塊（欄 1–13、列 1–12）放的是**肖像**，截圖上是一張遊俠的圖、
下面一行 `Ranger`。

推論等級：**已確認**（版面數字與 `sub_19727` 的四個設定逐項對上）。

⚠ **未解**：指令階段的肖像畫的是誰。截圖裡沒有敵人，出現的是遊俠的圖，
所以**至少在沒有敵人時畫的是隊伍這一邊**。有敵人時是不是換成敵人肖像
（remake 現在的做法，`docs/re/37` §3.2）還沒對拍——那要一張真的遭遇截圖。

## 3. remake 現在的狀態

| 項目 | 狀態 |
|---|---|
| `ENC` 空地問一句 ＋ 進指令階段 | **已接**（`internal/play/enc.go` 的 `beginEmptyRound`）|
| 指令選單的版面 | **還是重製決策**：七項排成一行擠進訊息視窗 |

第二項現在**不再是「未 RE」**——原版怎麼做已經知道了，只是還沒實作。
要照原版做得動三件事：肖像框、右邊那一塊文字區、以及名單那一塊的共存。
列在 `WORKLIST` 的「下一步的順序」裡。

## 4. 可重跑的完整指令

```bash
tools/dosbox.sh "wait:6;key:Return;wait:3;key:p;wait:4;shot:20-map;key:e;wait:3;shot:21-enc"
tools/dosbox.sh "wait:6;key:Return;wait:3;key:p;wait:4;key:e;wait:3;key:y;wait:3;shot:24-encyes"
tools/go.sh test ./internal/play/ -run TestEncOffersARoundWithNoEnemies -v
```

## 5. 這一輪學到的（寫成規則）

- **一句「找不到就印自己寫的話」會蓋掉一整條原版路徑。** `Nothing to fight here.`
  看起來只是個提示，實際上站在 `0x14` 那條分支的位置上。
  **重製版自己寫的字串要標出來**（`ui:` 前綴已經做到），
  而且**每一條都該問一次「原版在這個位置說什麼」**。
- **同一句話的兩個變體要當成兩條路。** `0x14` 與 `0x36` 後面接的都是 `0x4C`，
  很容易讀成同一件事；它們在程式碼裡差了 60 個位址、條件也不同。
- **開著的問題要回頭看自己留的線索。** `docs/re/40` §5 已經寫下
  「`ds:4674h ← 0x0E` 看起來是另一塊區域」，缺的只是一張截圖去確認。
  **列「未解」時要一起列「下一個入口」，而那個入口要真的去走。**
