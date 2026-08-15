# 62：第四道閘 —— nibble 11 是山與牆

日期：2026-08-15 ｜ 起因是實機對拍：**同一串按鍵，原版被山擋住，remake 走過去了**

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
實機環境見 `docs/re/47`。

`docs/re/26` §3 把四道閘列出來，但第四道 `sub_15CE0` 一直沒讀，
所以擋路清單掛著「**下界不是全集**」。這一份補上，而缺的那一項是最大的一項。

---

## 1. `sub_15CE0` 全文

```
0x15CE0  ds:A6B0h ← dl；ds:A6B1h ← bl      ; 目標座標存起來
0x15CE8  call sub_169EB                    ; al ← 目標格的 nibble，ds:46AEh ← 該筆記錄
0x15CEB  al ＝ 0Bh → loc_15D06             ; **nibble 11**
0x15CEF  al ＝ 4  → loc_15CF5              ; nibble 4：再看記錄
0x15CF3  clc; retn                         ; 其餘 → 可以走

loc_15CF5（nibble 4）：
  al ← 記錄 +0x01
  bit7 沒設 → clc retn（可以走）
  bit7 設   → dl ← 2

loc_15D06（nibble 11）：dl ← 1

loc_15D08:
  ds:A6B2h ← dl
  bl ← 0；call sub_16D1A                   ; **印記錄 +0x00 的訊息**
  al ← 記錄[dl]                            ; dl ＝ 1（nibble 11）或 2（nibble 4）
  bit7 設 → 跳過
  否則 sub_17CFF(al, 目標座標)             ; **改寫這一格**
  stc; retn                                ; 擋住
```

| nibble | 擋不擋 | 條件 |
|---:|---|---|
| **11** | **一律擋** | 無 |
| 4 | 條件式 | 記錄 `+0x01` 的 bit7 設起來才擋 |

## 2. nibble 11 是地圖上第二多的東西

全檔重數（`docs/re/61` §3）：**20,495 格、42 張地圖全部都有**，
僅次於 nibble 0 的 24,221。

`docs/re/26` §5 的**事件**分派表把 nibble 11 記成 `clc; retn`（踩上去什麼都不做）——
那是對的，但**永遠不會被呼叫**，因為第四道閘先擋住了。
兩張表講的是兩件事：一張是「踩上去做什麼」，一張是「能不能踩上去」。

擋住時印的是那一格記錄 `+0x00` 指的訊息。實機與重製版現在逐字相同：

```
This mountain is in your way. Go around.
```

## 3. 實機對拍

同一串按鍵（開場 → `P` → 19 個 Up → 26 個 Left）：

| | 結果 |
|---|---|
| 原版 | 03:20，`This mountain is in your way. Go around.`（訊息視窗兩行） |
| 修正前的 remake | 03:56，**穿過山走進 Quartz**（地圖 1 的 (29, 3)） |
| 修正後的 remake | 03:16，同一句訊息，停在 (40, 43) |

⚠ **穿山的症狀不是「明顯的錯」，是「路線與原版不同」。** 兩邊都跑得動、
都沒有崩潰，只有把同一串按鍵送進兩邊才看得出來。
單元測試永遠測不到這一項——它測的是「擋住的格子有沒有擋住」，
而漏掉的那個 nibble 根本不在清單裡。

## 4. 還沒做的

- `sub_17CFF` 的**改寫地圖格**：撞上去之後若記錄 `+0x01`／`+0x02` 的 bit7 沒設，
  原版會把那一格改寫（大概是「撞開的門」這類）。remake 還沒做。
- `sub_13E9B`（第三道閘，含 `sub_14085`）也還沒逐條讀完。

## 5. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/gate4.json 0x15CE0 --callers

# 兩邊走同一串
tools/go.sh run ./cmd/wl-play -script "up,x19,left,x26" -trace
tools/dosbox.sh "wait:6;key:Return;wait:4;key:p;wait:5;\
$(python3 -c "print(';'.join(['key:Up','wait:1']*19 + ['key:Left','wait:1']*26))");wait:3;shot:quartz1"
```

## 6. 這一輪學到的（寫成規則）

- **「下界不是全集」這種註記要當成待辦，不是免責聲明。** `docs/re/26` §3
  誠實寫了擋路清單不完整，但那句話擺了很久沒人回頭補——
  而缺的正是地圖上第二多的地形。
- **同一個 nibble 在兩張表裡可以有兩種答案。** nibble 11 在事件表裡是
  「什麼都不做」、在移動閘裡是「一律擋住」。看到「這個值沒有處理」
  要先問**是哪張表沒有處理**。
- **把同一串按鍵送進兩邊，是最便宜的整體驗收。** 這一輪的起點只是
  「走到 Quartz 拍張照」，結果先撞出擋路的缺口——
  **對拍不必先知道要找什麼。**
