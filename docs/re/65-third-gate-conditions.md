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

## 2. remake 目前把它無條件擋住

`internal/game/world.go` 的 `blocking` 把 nibble 2 列成無條件擋，
所以：

- **門永遠打不開**——通過檢定也走不過去。
- `EventGate` 是**死碼**：走不過去就不會觸發 `trigger()`，那個分支永遠不會跑。

⚠ **不要因此就把 nibble 2 從 blocking 拿掉。** 沒接上條件判定就放行，
等於門永遠開著——比永遠關著更糟（會跳過整段遊戲流程）。
要改就要一起接上 `Eval`。

## 3. 接上去要準備的

| 需要 | 現況 |
|---|---|
| `ParseGates` / `Eval` | 已實作 |
| `SkillTable`（技能資料表） | `Scene` 目前只載了物品表 |
| 判定失敗要印的訊息 | `sub_13EC9` 還沒逐條讀完 |
| 通過之後會不會改寫地圖格 | 未讀 |

## 4. 還沒讀的

- `sub_13EC9`：條件串列的執行（`docs/re/32` §6 解了資料格式，流程沒讀完）。
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
