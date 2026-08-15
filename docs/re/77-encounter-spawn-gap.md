# 77：遭遇覆蓋率盤點 —— 敵人格是生出來的，而 remake 沒有生成器

日期：2026-08-15 ｜ 接 `docs/re/76`（腳本覆蓋率）、`docs/re/39`（遭遇掃描）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 見 `docs/re/01`。

`docs/re/76` 量完腳本之後，同一把尺量戰鬥，量出一個比預期大的缺口。

---

## 1. 資料面：查得到，但幾乎都是空的

`TestEncounterTableCoverage` 掃 42 張地圖：

| 項目 | 數字 |
|---|---:|
| 有敵人資料表的地圖 | **42／42** |
| section 15 的記錄 | 159 筆 |
| 有敵人的組 | 17 組 |
| 型別在資料表裡查不到 | **0 組** |
| 整筆空的記錄 | **150 筆** |

`Battle.Spawn` 在 `table.Lookup(sg.Type)` 查不到時會靜靜跳過那一組——
症狀是「打起來少了一批敵人」。**這一項是乾淨的**：0 組查不到。

但 159 筆記錄裡 150 筆整筆是空的。

## 2. 為什麼是空的：section 15 是生成器的槽

`sub_16890` 每走一步跑一次（`docs/re/26` §6 記成「擲遭遇」，只涵蓋前四行）：

```
0x16890  bl ← 0x2F；al ← 標頭[+0x2F]        ; 遭遇分母
         ＝ 0 → retn
         sub_18E41(al)；≠ 1 → retn          ; 1／分母 的機率

loc_168A5（找空槽）:
         bl ← 0x0F；sub_17CB1(ds:46B4h)     ; section 15 的第 n 筆
         記錄[3] | 記錄[5] | 記錄[7] ＝ 0 → 用這一槽
         ds:46B4h++；到 標頭[+0x32] 為止    ; 槽數上限
         沒有空槽 → retn

loc_168CF（生成）:
         sub_18E41(標頭[+0x31])             ; 擲敵人種類（1..種類數）
         ds:AA3Ah ← al
         sub_12A4C；sub_12A8D → dl          ; 定址那一族的三張 13 項表
         記錄[0] ← [ds:AA60h + dl]
         記錄[1] ← [ds:AA6Dh + dl]
         ds:AA34h ← [ds:AA7Ah + dl]
         ... 在視窗內擲位置、檢查可通行 ...
loc_1695B:
         記錄[4] ← 擲出來的數量
         記錄[3] ← ds:AA3Ah                 ; 種類
         ds:46B3h ← 0x0F                    ; **新的第 1 層 ＝ nibble 15**
         jmp sub_17D50                      ; 寫進地圖
```

**敵人格是每走一步當場生出來的**，寫進 section 15 的空槽並把地圖上某一格
改成 nibble 15。出貨資料裡那 150 筆空槽就是留給它的位置；
9 筆非空的是固定遭遇。

`TestNoEnemyCellsInShippedMaps` 驗證了另一半：**42 張地圖的 nibble 15 格數 ＝ 0**。

## 3. remake 的缺口

`World.rollEncounter` 只做了 `sub_16890` 的前四行（擲 1／分母），
回一個 bool；`Scene.StartEncounter` 接著用 `ScanEncounters` 去**找**視窗裡的
敵人格——但地圖上一格都沒有，因為沒有人生成。

實測：

```
tools/go.sh run ./cmd/wl-play -script "path=32:62,left,right ×40" -trace | grep -c ATTACKED
0
```

**105 步、0 次遭遇。** 這不是機率問題，是結構問題：掃描器永遠掃不到東西。

⚠ 這個缺口沒有任何錯誤訊息，而且 `rollEncounter` 的單元測試是綠的——
它測的是擲骰，而擲骰是對的。**缺的是擲中之後那一段。**

## 4. 要補什麼

| 元件 | 狀態 |
|---|---|
| 標頭 `+0x2F`／`+0x31`／`+0x32`（分母／種類數／槽數） | 已解，`internal/game/script.go` 已在用 |
| 遭遇表 `ds:AA60h`／`ds:AA6Dh`／`ds:AA7Ah`（三張 13 項） | 位址已記（`docs/re/00-master-index.md`），**內容還沒解** |
| `sub_12A4C`／`sub_12A8D` 的索引算法 | 未解 |
| 找空槽、擲位置、寫回地圖 | 組語已讀（§2），還沒實作 |
| `ScanEncounters`（掃視窗、距離門檻） | **已實作**（`docs/re/39`） |

所以缺的是「生成」這一半，「掃描」那一半早就好了。

## 5. 可重跑的完整指令

```bash
tools/go.sh test ./internal/play/ -run TestEncounterTableCoverage -v
tools/go.sh test ./internal/play/ -run TestNoEnemyCellsInShippedMaps -v

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/enc3.json 0x16890
```

## 6. 這一輪學到的（寫成規則）

- **同一把尺量下一個子系統。** 腳本量完覆蓋率補了 37 格，
  用同樣的方法量戰鬥，量出「隨機遭遇完全沒發生」——
  這種缺口靠玩是會發現的，但要玩很久；量一次是幾秒。
- **「掃描器找不到東西」與「沒有東西可找」在程式裡長得一樣。**
  `ScanEncounters` 每一步都正確地回報「附近沒有敵人」，
  而那句話是真的——因為沒有人生成敵人。
  **回報「沒有」的元件要能區分這兩種沒有。**
- **單元測試綠 ≠ 那條路通。** `rollEncounter` 測的是擲骰，擲骰沒錯；
  錯的是擲中之後接不上。**測試要跟著資料流走到底**，
  不然每一段都對、整條路不通。
