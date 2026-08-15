# 85：地圖上的敵人圖示

日期：2026-08-15 ｜ 接 `docs/re/78`（遭遇生成）、`docs/re/84`（呈現層門檻）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 見 `docs/re/01`。

`docs/re/84` §4 列的最後一項有畫面影響的缺口。

---

## 1. `0x147BA`：三步查表

```
0x147BA  dl ← 0
loop:    bl ← ds:A5B1h[dl]              ; 三個位移：03 05 07
         al ← 敵方記錄[bl]              ; 那一組的型別
         非 0 → 用這一組
         dl++；到 3 為止

0x147D3  al 的 bit7 設 → 跳過
         sub_12A4C(al)                  ; 定址敵人資料表第 al 筆
         bl ← 6；al ← 那一筆的 +0x06
         al ← ds:AA17h[al]              ; ← **圖示編號**
         ds:A608h ＝ 0 → 不畫
         ...畫...
```

`ds:A5B1h` ＝ `03 05 07`，與 `ReadSpawnGroups` 的 `enemyTypeAt` 是同一組
（`docs/re/37`）。`ds:AA17h` ＝ `00 06 03 04 02 01 …`——
**那張表 `docs/re/48` §3 早就解過**，`internal/game/mapicons.go` 的
`kindIcon`／`KindIcon` 就是它，而 `EnemyData.Kind` 就是 `+0x06`。

所以整條鏈的每一段都已經解過也實作過了，缺的只是**沒有人把它們接起來**：
`ViewIcons` 只處理 nibble 4／5／9，遇到 nibble 15 直接跳過。

## 2. remake 這一側

`World.enemyIcon(x, y)`：

```
section 15 記錄的 +0x03 ＝ 敵人種類編號
  → 敵人資料表第 n 筆的 +0x06（＝ EnemyData.Kind）
  → KindIcon（ds:AA17h）
```

任何一步取不到就不疊圖——原版 `0x147D3` 的 `js` 也是遇到 bit7 就跳過。

驗收：

```
tools/go.sh test ./internal/play/ -run TestEnemyCellGetsIcon -v
    敵人格 (12,3) 疊上圖示 6
```

圖示 6 ＝ `IconAnimal`（種類 1）。生成器放的格子現在畫得出來了。

## 3. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/enemyicon.json 0x147BA 0x147E0 0x14800

tools/go.sh test ./internal/play/ -run TestEnemyCellGetsIcon -v
```

## 4. 這一輪學到的（寫成規則）

- **「還沒解」與「解了沒接」要分開列。** `docs/re/84` §4 把這一項寫成
  「還沒解」，實際上三段查表全部解過（`docs/re/37`、`docs/re/48`），
  只是 `ViewIcons` 沒接。**列缺口時要先 grep 一次相關的常數名**——
  `kindIcon` 就在 `internal/game/mapicons.go` 裡等著。
- **同一張表會在兩份筆記裡出現。** `ds:AA17h` 在 `docs/re/48` 是
  「疊圖編號表」、在這裡是「敵人圖示查表的第三段」。
  **總表（`00-master-index.md`）的位址欄是唯一能把它們對起來的地方。**
