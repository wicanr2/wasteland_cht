# 82：存檔的三道門檻

日期：2026-08-15 ｜ 接 `docs/re/30`（存檔格式）、`docs/re/81`（戰鬥門檻）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 見 `docs/re/01`。

第五把尺量存檔。`CLAUDE.md` §4 寫死了驗收標準：
**「從原始 bytes 出發只蓋已解欄位，未解區域一個 byte 都不動，
讀出再寫回 byte-for-byte 相同」**——這一份把那句話變成三個測試。

---

## 1. 讀出再寫回 byte-for-byte 相同

`TestSaveRoundTripsByteForByte`：對 `game1` 與 `game2` 各讀一次存檔，
呼叫 `Save.Bytes()` 重新編碼，與檔案裡原本那 4,614 bytes 逐 byte 比對。

```
game1：4614 bytes 完全相同
game2：4614 bytes 完全相同
```

這道門檻同時涵蓋了 XOR 解密、checksum 重算與尾段保留——
只要有人在編碼路徑上「重建」了什麼，數字就對不上。

## 2. 玩一輪之後，改動要限縮在已解欄位

`TestSaveKeepsUnknownBytesAfterPlaying`：走到打起來、打完一場、收尾，
再 `StoreTo` 一份新讀的存檔，數有幾個 byte 變了。

```
玩了一輪（含一場戰鬥）之後，存檔明文有 12／2048 bytes 改變
```

12 bytes ＝ 座標 2 ＋ 時鐘 6 ＋ 角色記錄裡變動的那幾個。
門檻設在 200——**全部 2,048 bytes 都變就是在重建而不是改寫**。

既有的 `TestStoreToAfterWalkTouchesOnlyKnownBytes` 只走了幾步；
這一個把事件、改寫地圖格、戰鬥結算都跑過一輪。

## 3. 該存的真的存進去了

前兩道守的是「沒動到不該動的」，第三道守相反的方向。
`TestSaveLoadRoundTripKeepsState` 玩 20 步（含戰鬥）之後 `StoreTo`，
再從存檔的明文讀回座標、時鐘與第一個角色：

```
存讀循環一致：(12,2) 2:20 Hell Razor CON=28
```

## 4. 目前存進去的欄位

| 位置 | 內容 |
|---|---|
| 隊伍槽表 `+0x08`／`+0x09` | 隊伍座標 |
| 全域區 `+0`／`+1` | 視窗原點 |
| 全域區 `+2`–`+4` | 24-bit 總時間 |
| 全域區 `+10`–`+12` | 分的小數、分、時 |
| 每個角色的記錄 | `Character.StoreTo`：名字、七個屬性、金錢、性別、國籍、AC、CON／MaxCON、裝備槽、技能點 |

⚠ **隊伍槽表 `+0x0A`（所在地圖）目前沒有寫**——註解說要寫，程式碼只寫了
`+0x08`／`+0x09`。換地圖之後存檔會記成還在原地圖。列進待補。

## 5. 可重跑的完整指令

```bash
tools/go.sh test ./internal/play/ -run TestSaveRoundTripsByteForByte -v
tools/go.sh test ./internal/play/ -run TestSaveKeepsUnknownBytesAfterPlaying -v
tools/go.sh test ./internal/play/ -run TestSaveLoadRoundTripKeepsState -v
```

## 6. 這一輪學到的（寫成規則）

- **同一個子系統要兩個方向的門檻。** 「沒動到不該動的」與「該存的真的存了」
  是兩件事，只測一邊會漏掉另一邊——一份什麼都不寫的 `StoreTo` 可以完美通過
  第一道門檻。
- **把規格裡的句子直接變成測試。** `CLAUDE.md` §4 的「byte-for-byte 相同」
  寫了很久，但在這一輪之前沒有任何東西在檢查它。
  **規格裡的驗收標準要有對應的斷言，不然它只是一句話。**
- **量測會順手照出沒寫完的程式碼。** `+0x0A`（所在地圖）的註解寫著要存，
  實際上沒存——是列欄位表的時候發現的，不是測試紅的。
