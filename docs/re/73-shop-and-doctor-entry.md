# 73：商店與醫生的入口 —— 傳送記錄的 `+0x04`／`+0x05`

日期：2026-08-15 ｜ 接 `docs/re/72`（實機解出進地點的路徑）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 見 `docs/re/01`。

擋住「玩得通」最久的一項解開了。答案在 `docs/re/60` §2 已經抄下來的一行裡。

---

## 1. 機制

```
0x16A24  ds:AAD3h ← 0；sub_169B1(4)
```

`sub_169B1(al)` ＝ `sub_17CFF(al, ds:46A6h, ds:46A7h)` ＝ **用記錄的位移 al
改寫隊伍腳下那一格**。傳送用的位移是 **4**，所以傳送記錄的
`+0x04`／`+0x05` 是「**進去之後這一格變成什麼**」。

完整路徑：

```
踩上 nibble 10（記錄 +0x00 的 bit6 設）
  ↓ sub_16AD5 → Enter new location? → Yes
sub_16A10：先 sub_169B1(4) 把腳下改寫成 (記錄 +0x04, 記錄 +0x05)
  ↓ 再依 +0x03 決定去哪（＝ 自己這張地圖時等於原地不動）
腳下那一格現在是 nibble 6 ＋ 設施記錄
  ↓ 事件分派
設施畫面（ds:A4E0h 跳表）
```

## 2. 資料：22 筆，全中

掃 42 張地圖的 599 筆傳送記錄，`+0x04`／`+0x05` 指到 nibble 6 的有 **22 筆，
而那 22 筆全部指到設施記錄**——沒有一筆例外：

| 資源 | 記錄 | 設施 |
|---:|---:|---|
| 1 | 4 | `Q. Emporium` |
| 7 | 0 | `Clone prep.` |
| 8 | 0 | `Trading car` |
| 9 | 0 | `AG. store` |
| 10 | 7 ／ 8 | `Infirmary` ／ `Store` |
| 12 | 8 ／ 9 | `LV Infirmary` ／ `Vegas Lib.` |
| 13 | 5 ／ 6 | `Books` ／ `Sickbay` |
| 21 | 19–22 | `Market`／`Blackmarket`／`Patch em' up`／`Library` |
| 26 | 25 ／ 26 | `Old Doc Bobs` ／ `New Thoughts` |
| 32 | 6 | `Thrift store` |
| 36 | 7 | `Tech section` |
| 38 | 8 ／ 9 | `NUCLEAR AID` ／ `HOLY KNOWING` |

「32 筆設施只有 2 筆有格子指到」的統計沒有錯，它只是問錯了問題——
**設施格不在初始地圖上，是傳送收尾當場改寫出來的。**

## 3. 實機與 remake

```
tools/dosbox.sh "wait:6;key:Return;wait:4;key:p;wait:6;key:i;wait:3;key:k;wait:4;key:y;wait:5;shot:entered"
```

實機答 Y 之後直接是設施畫面：`Ranger Ctr.` 的圖、隊伍名冊、
底部 `CREATE DELETE PLAY`（跳表索引 3，`docs/re/72` §3）。

remake 接上之後：

```
tools/go.sh run ./cmd/wl-play -script "up,down,y" -trace
   3 y   地圖 0 (55,62) 01:08 設施 …          ← Ranger Ctr.
tools/go.sh run ./cmd/wl-play -script "map=10:30:25,up,y" -trace
   3 y   地圖 10 (30,24) 01:00 設施 …         ← Store
tools/go.sh run ./cmd/wl-play -script "map=21:2:22,up,y" -trace
   3 y   地圖 21 (2,21) 01:00 設施 …          ← Market
```

`TestTeleportPatchOpensFacility`、`TestRangerCenterOpensFacility`。

## 4. remake 這一側

| 位置 | 做了什麼 |
|---|---|
| `internal/game/world.go` | `PatchHere(record, at)`：用記錄的位移改寫腳下那一格 |
| `internal/play/play.go` | `doTeleport` 開頭先 `PatchHere(rec, 4)`；傳送收尾呼叫 `enterFacilityHere()` |

`enterFacilityHere` 檢查腳下是不是 nibble 6 且記錄 `+0x00` 的 bit7 設，
是就進設施。設施記錄的 `+0x01`／`+0x02` 是 `fd fd`（沿用 ＝ 不改，
`docs/re/69` §7），所以不會繞回傳送格。

## 5. 可重跑的完整指令

```bash
tools/go.sh test ./internal/play/ -run TestTeleportPatchOpensFacility -v
tools/go.sh test ./internal/play/ -run TestRangerCenterOpensFacility -v
tools/go.sh run ./cmd/wl-play -script "map=10:30:25,up,y" -trace
tools/dosbox.sh "wait:6;key:Return;wait:4;key:p;wait:6;key:i;wait:3;key:k;wait:4;key:y;wait:5;shot:entered"
```

## 6. 這一輪學到的（寫成規則）

- **答案在自己三個月前抄下來的那一行裡。** `docs/re/60` §2 寫著
  `0x16A24 ds:AAD3h ← 0；sub_169B1(4)`——那時只當成流程的一步抄下來，
  沒問「位移 4 是什麼」。後來花了四輪掃遍所有其他改寫來源。
  **要說「找不到」之前先重讀自己已經抄過的組語，逐行問每一個常數是什麼。**
- **統計問錯問題不會有錯誤訊息。** 「幾筆設施有格子指到」算得完全正確，
  數字也對，但那個問題本身預設了「格子是靜態的」。
  **在會自我改寫的系統裡，任何以初始狀態為基礎的可達性統計都要標明前提。**
- **一個 22/22 的命中率比任何推論都強。** 掃出來的 22 筆全部是設施、
  沒有一筆雜訊——這種分布不可能是巧合，可以直接停止找別的解釋。
