# 65：第三道閘的另一半 —— nibble 2 是條件式的

日期：2026-08-15 ｜ 接 `docs/re/64`（第三道閘的前半：進新地點的確認）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

---

## 1. `sub_13E9B`

```
0x13E9B  ds:A5D8h ← dl；ds:A5D7h ← bl    ; 目標座標
0x13EA3  call sub_13EC0
0x13EA6  call sub_14085                  ; nibble ＝ 2 → CF ＝ 1
0x13EA9  jnb loc_13EBC                   ; 不是 nibble 2 → **放行**
0x13EAB  call sub_1407E
0x13EAE  jnb loc_13EBC                   ; → 放行
0x13EB0  call sub_13EC0
0x13EB3  call sub_13EC9                  ; ← nibble 2 的條件串列（docs/re/32 §6）
0x13EB6  jb  loc_13EA3                   ; CF ＝ 1 → **回頭再判一次**
0x13EB8  al ≠ 0 → stc（擋住）
0x13EBC  clc; retn                       ; 放行
```

`sub_14085` 只有三條：取目標格的 nibble，等於 2 就 `stc`。

**所以 nibble 2 不是「一律擋」，是「先判條件」。** 條件串列的解析與判定
（型別、難度、技能、物品）在 `docs/re/32` §6，remake 也已經寫好
（`internal/game/gates.go` 的 `ParseGates`／`Eval`）。

## 1.1 三條路，前兩條是直讀的

```
記錄 +0x00 的 bit7 設 → 放行          ; sub_1407E 的 `js`
bit6 沒設             → 放行          ; sub_13EC9 的 `shl` ＋ `js`
兩者都不成立          → 跑條件串列    ; 記錄 +0x0A 起，0xFF 結束
```

資料面（42 張地圖全掃）：

| 路 | 格數 |
|---|---:|
| bit7 設 → 直接放行 | **1,522** |
| bit6 沒設 → 放行 | **1,031** |
| 要判定 | **146** |

**把整個 nibble 2 當成牆會擋掉 2,553 格（94%）本該能走的格子。**

擋住時印的是記錄 **`+0x01`**（`sub_13EC9` 的 `mov bl, 1`），不是 `+0x00`。

## 2. remake 原本把它無條件擋住

`internal/game/world.go` 的 `blocking` 把 nibble 2 列成無條件擋，
所以：

- **門永遠打不開**——通過檢定也走不過去。
- `EventGate` 是**死碼**：走不過去就不會觸發 `trigger()`，那個分支永遠不會跑。

**已修**：前兩條路直接照 bit7／bit6 放行（那是直讀的，沒有判定風險），
只有那 146 格維持擋住——與修正前的行為相同，不會更糟。

⚠ **剩下的 146 格不要「先放行再說」。** 沒接上判定就放行等於門永遠開著，
會跳過整段遊戲流程；那比擋住更難發現。

## 3. 那 146 格要接上去，還缺什麼

| 需要 | 現況 |
|---|---|
| `ParseGates` / `Eval` | 已實作，但**只試目前這一個角色**——原版是逐個隊員試（`ds:A5D3h` 是迴圈變數，上限 `ds:4653h`） |
| `SkillTable`（技能資料表） | `Scene` 目前只載了物品表 |
| 判定失敗要印的訊息 | `sub_13EC9` 還沒逐條讀完 |
| 通過之後會不會改寫地圖格 | 未讀 |

## 4. 還沒讀的

- `sub_13EC9` 的後半（`0x13FB1` 起）：判定跑完之後的收尾——
  `sub_14193`（腳本改屬性）、`sub_142D7`／`sub_142ED`／`sub_1417A`，
  以及 `sub_142E2 & 0x10`／`& 0x20` 兩個旗標分支。
- 五種條件型別的判定函式：`sub_1968C`（技能檢定）、`sub_17B3E`（型別 4）、
  `sub_1820C`（型別 2）、`sub_198CD` ＋ `sub_180F0`（型別 0）。
- `sub_1407E`、`sub_13EC0`。

## 5. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/gate3.json 0x13E9B 0x14085 --callers
```

## 6. 這一輪學到的（寫成規則）

- **「已解的規則」與「接上去的規則」是兩件事。** 條件串列的解析與判定
  早就寫好也測過，但移動那一側從來沒呼叫它——
  **兩邊各自都綠，中間是斷的。**
- **無條件擋住是個安全但會累積的近似。** 它讓遊戲跑得動、測試全綠，
  代價是整類內容（門後面的東西）永遠到不了，而且不會有任何錯誤訊息。
  **這種近似要寫進文件，不然下一個人會以為它是原版行為。**
