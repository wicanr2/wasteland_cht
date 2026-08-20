# 87：`sub_15036` 是敵人在地圖上移動，不是目標選擇

日期：2026-08-15 ｜ 接 `docs/re/51`（遭遇驅動器）、`docs/re/86`（戰鬥訊息）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`internal/play/round.go` 的註解把 `sub_15036` 說成「原版的目標表」。
讀完之後那句話是錯的。

---

## 1. 它做什麼

```
0x15036  ds:A631h ← 0（索引）；ds:A630h ← 0（組）
loop:
  sub_137F4(ds:A630h)              ; 定址敵方記錄第 n 組
  al ← ds:711Dh[ds:A631h]          ; 這一隻的動作
  ＝ 0FFh → 下一個
  ds:A635h ← al & 0x3F             ; 低 6 位
  ds:A634h ← al >> 6               ; 高 2 位 ＝ **訊息索引**
  x ← ds:46C8h[0]；y ← ds:46C8h[1] ; 遭遇佇列的座標
  sub_1513B；sub_16840             ; 算新座標、檢查有沒有別的隊伍
  sub_17C20 ≠ 0 → 放棄             ; 目的地要是空地
  ds:46B3h ← 0；ds:46B4h ← 0；sub_17D50   ; **把舊格子清成 nibble 0**
  更新遭遇佇列的 x／y
  把新格子設回原本的 (nibble, 記錄)        ; sub_17D50
  sub_19EFC；sub_13787
  al ← ds:A643h[ds:A634h]；sub_16CB2       ; 印訊息
```

`ds:A643h` 的三個訊息（執行檔字串表 1 的 72／74／75）把它的身分講死了：

```
" move\ns\n\n to a better position."
" run\ns\n\n away."
" run\ns\n\n at you."
```

**這是敵人在地圖上換位置**：清掉舊格、寫新格、更新遭遇佇列、印一句話。
`\n`（`0x0A`）是單複數分隔（`docs/re/17` §4.1）。

`ds:711Dh` 靜態全 0——**那張表是執行期填的**，不是出貨資料。

## 2. 對 remake 的意思

這一支與「打誰」無關。**打誰解在 [`89`](89-enemy-target-and-down.md) §1**
（`roll(1..隊伍人數)` 隨機挑，挑到 CON ≤ 0 的就整個重抽），
remake 那一側是 `internal/play/round.go` 的 `pickEnemyTarget`。

移動本身也接上了：計畫在 `game.PlanEnemyMove`、執行在
`internal/play/enemymove.go`，逐項見 [`116`](116-enemy-move-execution.md)。

## 3. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/target.json 0x15036
```

## 4. 這一輪學到的（寫成規則）

- **程式碼註解裡的斷言要跟文件一樣認真。** 「原版走 `sub_15036` 的目標表」
  寫在 `round.go` 的註解裡，沒有任何文件支持它，也沒有人查過。
  **註解裡的 `sub_xxxxx` 要嘛附文件出處，要嘛標成猜測。**
- **訊息字串是最快的身分證。** 讀了 40 條指令還在猜這一支做什麼，
  取出 `ds:A643h` 的三句話之後一秒確定。
  **看到「印訊息」那一步就先把字串撈出來。**
