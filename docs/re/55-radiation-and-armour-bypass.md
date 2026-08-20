# 55：輻射結算與「無視護甲」旗標（`ds:46EFh`）

日期：2026-08-15 ｜ 接 `docs/re/48` §5（`ds:46EFh` 的讀取端）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
地圖資料 `game1`／`game2` 的雜湊見 `docs/re/01`。

`docs/re/48` 解出 nibble 9 ＝ 輻射區，但 `ds:46EFh` 只追到寫入端。
這一份補上讀取端，順帶把整條輻射結算讀完。

---

## 1. `ds:46EFh` 是「這一次結算跳過護甲吸收」

讀取端全檔只有一處：`sub_157D6` 的 `0x15840`（傷害結算的主函式）。

```
0x15840  mov  al, ds:46EFh
0x15843  test al, al
0x15845  jz   loc_1584F        ; ＝ 0 → 照常算護甲
0x15847  mov  al, 0
0x15849  mov  dl, al           ; ≠ 0 → 顆數 0
0x1584B  test al, al
0x1584D  jz   loc_15861        ; 一定成立 → 跳過下面三條
0x1584F  mov  bl, 1Ah
0x15851  mov  di, ds:46B5h
0x15855  mov  al, [bx+di]      ; 角色記錄 +0x1A ＝ AC
0x1585A  mov  dl, al
0x1585C  mov  al, 0
0x1585E  call sub_19D86        ; 0 ＋ AC 顆 d6（docs/re/13 §3.3）
0x15861  mov  ds:0A69Ah, al    ; 吸收值（16-bit）
0x15864  mov  ds:0A69Bh, dl
```

非 0 時 `dl` ＝ 0，`sub_19D86` 那條被跳過，`ds:A69Ah/A69Bh` 直接吃到
上一步留下的 0 ——**吸收 ＝ 0，傷害全額扣在 CON 上**。

## 2. 它是參數，不是狀態

三個寫入點都在**呼叫結算的前後**，沒有一個是持久的：

| 位址 | 寫什麼 |
|---|---|
| `0x1423C` | `sub_142E2() & 1`，緊接著就 `call sub_157D6` |
| `0x14252` | 結算回來立刻清 0 |
| `0x14417` | 輻射迴圈開始前設 1（見 §3） |

`sub_142E2`（`0x142E2`，12 個呼叫端）讀的是**目前地圖記錄的第一個 byte**：

```
mov  bl, 0
mov  di, ds:46AEh    ; 目前地圖記錄的指標
mov  al, [bx+di]
test al, al
retn
```

`sub_16D1A`（印訊息那一支）開頭一模一樣，非 0 才 `jmp sub_17920` ——
**兩支讀的是同一個 byte，那個 byte 是訊息編號**。
資料佐證：資源 26 與 38 的 nibble 9 記錄 `+0x00` 全部是 0，
而 `docs/re/48` §5 記的正是「字串編號 0（不印訊息）」。

所以對輻射格而言，`ds:46EFh` ＝ **訊息編號的 bit0**。

⚠ **這個 byte 的語意隨 section 型別而異**（同 `docs/re/16` §3.1 的教訓：
型別就是 nibble 本身）。`sub_142E2` 的另外 11 個呼叫端在腳本指令那一區，
它們的記錄 `+0x00` 不是訊息編號。**「訊息編號 ＆ 1」只是輻射這條路上的樣子**，
不是這個 byte 的定義。

推論等級：`ds:46EFh` 的**效果**（跳過護甲吸收）**已確認**；
它的**值來自地圖記錄 `+0x00` 的 bit0` **已確認**（三支函式直讀同一個 byte，
中間沒有改 `ds:46AEh` 的呼叫）。
**為什麼設計成訊息編號的奇偶** —— **未解**。

## 3. 輻射結算的迴圈（`0x14410`）

`docs/re/29` §2 記的是「印訊息、設旗標」，實際上整條結算都在這裡：

```
0x14410  bl ← 0
0x14412  call sub_16D1A               ; 印記錄 +0x00 指的訊息（0 就不印）
0x14417  ds:46EFh ← 1                 ; 這一輪先設 1
0x1441A  ds:46F0h ← 1                 ; 批次中：sub_19EFC 整支跳過
0x1441D  ds:0A5F0h ← al               ; ← 迴圈頂，al ＝ 隊員編號
0x14420  call sub_19614               ; 依編號設 ds:46B5h（角色記錄位址）
0x14423  call sub_196C4              ; bl ← 護甲槽（+0x25）指到的物品陣列位移
0x14426  bl ＝ 0（沒穿護甲）→ 直接受傷
0x1442C  al ← 角色記錄[bl]            ; 護甲的**物品編號**
0x14432  cmp al, 29h ＝ 41 → **跳過這個人**  ; 物品 41 ＝ Rad suit
0x14436  bl ← 1；dl ← 地圖記錄 +0x01  ; **骰子顆數**
0x14442  call sub_19D86               ; 0 ＋ dl 顆 d6
0x14445  call sub_19BF0               ; 結果 → ds:46BEh/46BFh（傷害）
0x14448  al ← 隊員編號；dl ← 1Dh；bl ← 0FFh
0x1444F  call sub_141FA               ; 扣 CON
0x14452  角色記錄 +0x28 |= ds:A5F1h[0] ; **狀態位元 0**
0x14462  隊員編號 ＋1；< ds:4653h 就回迴圈頂
0x14470  ds:46EFh ← 0；ds:46F0h ← 0
0x14478  al ← 2；jmp sub_169B1
```

- `ds:A5F1h` 起是 `01 02 04 08 10 20 40 80`（位元遮罩表，與 `ds:CF4Eh` 同一份內容），
  索引 0 → `0x01` → **Radiation poisoning**（`docs/re/35` §1 的 bit 0）。
- **結算從頭到尾不看蓋氏計數器**（物品 48）。那件東西是提示裝置，走的是另一條路：
  畫面右緣的計量表與逼近時的滴答聲，見 [`120`](120-geiger-counter.md)。
- **穿著 `Rad suit`（物品 41）的人整個跳過**——不扣血也不中毒。
  `sub_196C4` 取護甲槽 `+0x25` 指到的物品編號，`0x14432` 拿它比 `0x29`；
  物品表第 41 筆就叫 `Rad suit`（`docs/re/59` §2.2）。
- `sub_141FA(al ＝ 隊員編號, dl ＝ 欄位位移, bl)`：`bl` 有號負 ＝ 減、正 ＝ 加；
  欄位 `0x1D`（CON）走傷害結算、`0x15`／`0x21`（金錢／經驗）走 24-bit 專用路徑，
  其餘直接對記錄那個 byte 加減並夾在 0–255。
- `ds:46F0h` 只有 `sub_19EFC` 讀：非 0 就整支 `retn`。
  **批次結算期間把逐次的那件事關掉**，結束再打開。

## 4. 資料：211 格輻射區長什麼樣

`tools/summarize_radiation.py` 的輸出：

| 資源 | 格數 | `+0x00`（訊息編號） | bit0 ＝ 1 的格 | `+0x01`（骰數） |
|---:|---:|---|---:|---|
| 0 | 36 | 23 | 36 | 3、5 |
| 19 | 60 | 28／29／30 | 20 | 4／6／10 |
| 20 | 44 | 69／70／71 | 28 | 2／3／4 |
| 26 | 54 | 0（不印） | 0 | 2、4 |
| 31 | 1 | 12 | 0 | 6 |
| 38 | 16 | 0（不印） | 0 | 4 |

合計 211 格，**84 格的傷害無視護甲**（訊息編號是奇數的那些），
骰數落在 2–10 顆 d6。

## 5. 還沒讀的

- `sub_169B1(2)` → `sub_17CFF` 是**改寫地圖格**的通路，但 211 個輻射格的
  記錄 `+0x02` **全部是 `0xFF`**（bit7 設 ＝ 不改），所以踩過不會變地形。
- `sub_141FA` 的另一條入口 `sub_14193`（3 個呼叫端，都在腳本指令區）——
  那是「腳本改屬性」的通路，值得單獨讀。

## 6. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_range_refs.py \
  workplace/analysis/dumps/rad46ef.json 0x46EF 0x46F0

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/rad_reader.json 0x157D6 0x141FA --callers

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/rad_setters.json 0x142E2 0x14410

python3 tools/summarize_radiation.py \
  workplace/analysis/unpacked/wl.merged.exe \
  workplace/orig/wastland/game1 workplace/orig/wastland/game2
```

## 7. 這一輪學到的（寫成規則）

- **「只有這支寫它」要掃過才能講。** `docs/re/48` 寫成只有一處，實際是三處，
  而漏掉的兩處正好是決定語意的那兩處（設定與清除）。
  掃一次的成本是一條指令，猜錯的成本是整個機制解反。
- **同一個 byte 被兩支函式讀，就拿另一支的用途去驗。** `sub_142E2` 單看
  只是「讀一個 byte」，但 `sub_16D1A` 拿同一個 byte 當訊息編號並在 0 時不印，
  而資料裡那兩張地圖的值剛好全是 0 —— **三邊對上才敢說它是什麼**。
- **設了又清的全域變數是參數，不是狀態。** 看到 `寫 → call → 清` 這個形狀
  就不要把它當成「玩家踩在某個東西上」的旗標；它的生命週期只有一次呼叫。
