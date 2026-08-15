# 91：地圖指令列的七個處理程式 —— 升級的入口是 RADIO

日期：2026-08-15 ｜ 接 `docs/re/72` §4（指令列的字串）、`docs/re/31`（升級公式）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/31` 解完升級的門檻與動作，卻沒解「**什麼時候會檢查升級**」，
remake 因此把 `LevelUp()` 寫好了放著，**一個呼叫端都沒有**。
答案在地圖指令列：升級是玩家自己按 `R` 用無線電呼叫總部要來的。

---

## 1. 兩張表，不是一張

`sub_16C7C`（`0x16B36` 唯一呼叫端）：

```
ds:468Dh ← 0AB18h      ; 選項字串位址表
ds:468Fh ← 0AB1Ah      ; 處理程式位址表
ds:468Ah ← 0           ; 目前組號
ds:4688h ← 0
ds:4689h ← 6           ; **最大索引**，不是組數 → 0–6 共七項
jmp sub_173D2
```

兩張表都是**指標的指標**（一組一個 word）：

```
ds:AB18h ＝ A9CCh  → 字串 "Use Enc Order Disband View Save Radio\0IKJL..."（docs/re/72 §4）
ds:AB1Ah ＝ AB1Ch  → 處理程式陣列從 ds:AB1Ch 起
```

`ds:AB1Ch` 的七個 word 依序對到那七個字：

| 索引 | 指令 | 處理程式 | 第一印象 |
|---:|---|---|---|
| 0 | `Use` | `0x13A80` | `sub_19727` → 隊伍只有一人就直接用他，否則印訊息 2 問 `Which player?`（`sub_1721B` 挑隊員） |
| 1 | `Enc` | `0x11CE7` | 最長的一支（46 條起跳），開頭清 `ds:A436h`／`ds:CA64h` 並把人數搬進 `ds:71B7h` |
| 2 | `Order` | `0x12AE0` | `ds:4653h`（人數）減 1 為 0 就 `retn`——**一個人不用排隊形** |
| 3 | `Disband` | `0x15E77` | 同樣先擋「只有一個人」 |
| 4 | `View` | `0x160A8` | 拿 `ds:4654h` 寫進 `ds:A6C2h`／`ds:A6C3h`，再與 `ds:4657h` 比對後 ±1 |
| 5 | `Save` | `0x1A290` | 印訊息 `0x49` → `sub_19B4F` 等 Y／N → `sub_1A2B1` → 寫檔（`ds:4680h ← 0CDF0h`） |
| 6 | **`Radio`** | **`0x15260`** | **升級**，見 §2 |

`ds:4689h ＝ 6` 與七個字、七個 word 三邊互相印證。

推論等級：**已確認**（表的設定逐指令讀完；七個位址是從 `ds:AB1Ch` 直接倒出來的）。

## 2. RADIO ＝ 問一句，答 Y 就呼叫總部

```
0x15260  al ← 1
0x15262  sub_16CB2        ; 印執行檔字串 1
0x15265  sub_19B4F        ; 等 Y／N（熱區遮罩 0x404；按 'N' → CF 設）
0x15268  jb  loc_15270    ; 答 N → 走人
0x1526A  sub_1A2B1        ; sub_18801 ＋ sub_183B1（換畫面）
0x1526D  jmp loc_1B8AD    ; → 升級流程
```

`Save`（`0x1A290`）走的是同一組「印訊息 → `sub_19B4F` → `sub_1A2B1`」，
只是訊息編號不同（`0x49` vs `1`）——**兩個指令的確認流程是同一套**。

### 2.1 升級流程有兩個迴圈

`loc_1B8AD` 進來之後先搬 13 bytes 的表（`ds:D515h` → `ds:7201h`）並載圖 8，
然後對每個隊員跑**兩輪**：

```
第一輪（loc_1B8D4）：
    loc_19624                  ; 選中這個人
    loc_172BE                  ; CON ≤ 0 → 跳過（docs/re/89 §2）
    角色 +0x4B 的 bit0 ＝ 0 → 跳過
    角色 +0x4C 的 bit0 ≠ 0 → 跳過
    角色 +0x4C 的 bit0 ← 1
    ds:D436h ＝ 0 → 印字串 0x3D（sub_1BB5D）並 ds:D436h++
    載圖 11；**sub_1B8A0 ＝ 播音效 4**；sub_19720
    ds:4687h ← 0x0A            ; 訊息的數量選擇子（docs/re/28）

第二輪（0x1B920）：
    loc_19624；loc_172BE       ; 同樣跳過倒下的人
    0x1B93D                    ; 升級門檻檢查（docs/re/31 §1）
    夠 → 0x1BA08 升級（docs/re/31 §2），可以連升
```

`+0x4B` bit0 ＝ **參與過摧毀 Base Cochise**（結局那一段逐人設起來），
`+0x4C` bit0 ＝ **總部已經表揚過**，所以那段賀詞一個人只聽得到一次。
賀詞是**階級表**（`ds:D622h`）的第 `0x3D` 條，不是第 1 張表——
`sub_1BB5D` 印之前會換 `ds:4692h`（`docs/re/96` §5）。

推論等級：**已確認**（指令流、跳轉目標與兩個旗標的語意都讀到）。

## 3. remake 這一側

指令列在 remake 裡**整條不存在**——地圖模式只認方向鍵。
這一輪先把框架與兩個確認得起來的指令接上：

| 指令 | 這一輪 | 備註 |
|---|---|---|
| `Save` | **已接** | 存檔邏輯本來就有（`Scene.StoreTo`），缺的只是入口 |
| `Radio` | **已接（兩輪）** | 第一輪表揚（`+0x4B` → 賀詞 → `+0x4C`），第二輪逐人檢查經驗值並升級，可連升；音效 4 照播 |
| `Use`／`Enc`／`Order`／`Disband`／`View` | 未接 | 入口已定位，見 §1 的表 |

Radio 的第一輪一度沒做，因為 `+0x4B`／`+0x4C` 的語意未解；
`docs/re/96` §5 把兩個旗標解開之後補上了。

## 4. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/cmdbar.json 0x16C7C
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/cmds.json 0x13A80 0x11CE7 0x12AE0 0x15E77 0x160A8 0x1A290
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/radio.json 0x1B8AD 0x1B8DF 0x1B8F1 0x1B90E 0x1B91F

python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xAB1C 7 --code

tools/go.sh test ./internal/play/ -run TestRadioLevelsUp -v
```

## 5. 這一輪學到的（寫成規則）

- **「公式解完」不等於「知道什麼時候用」。** `docs/re/31` 把升級門檻、
  階級階梯、每級給幾點技能點全解完了，卻沒有人問「誰呼叫它」——
  remake 因此有一支零呼叫端的 `LevelUp()`。
  **解完一個機制要補一句「入口在哪」，沒補就是只解了一半。**
- **零呼叫端的公開函式是接線漏洞的化石。** `LevelUp()` 編得過、測得過，
  只是永遠不會執行。`docs/re/00-wiring-status.md` 擋得住「整份筆記沒接」，
  擋不住「函式沒人叫」——後者要靠讀程式碼發現。
- **自己掃 bytes 找呼叫端要三種都掃。** 這一輪找 `0x1B8AD` 的來源時，
  只掃 `E8`（call）得到零命中，補上 `E9`／`EB`（jmp）才找到那條尾呼叫。
  **零命中之前先確認掃描涵蓋了所有可能的編碼。**
