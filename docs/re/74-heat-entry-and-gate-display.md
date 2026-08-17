# 74：高溫記錄的入口（再排除三條）與 `sub_142ED` 的顯示層

日期：2026-08-15 ｜ 接 `docs/re/66`（高溫的線索）、`docs/re/69`（條件閘旗標）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 見 `docs/re/01`。

---

## 1. `sub_142ED`：用指定的捲動速度把那一句慢慢印出來

```
sub_142ED:
  sub_19720                    ; 讀 ds:4707h，非 0 就走另一條
  dl ← ds:A5C5h                ; 上一次留下來的捲動速度等級
  sub_14314(dl)                ; ds:A5E7h ← ds:46F7h；ds:A5E6h ← ds:465Bh（存）
                               ; ds:46F7h ← 1（開捲動動畫）；ds:465Bh ← dl（設速度）
  sub_178A0(記錄 +0x03)        ; → sub_1790B，印那一條訊息
  sub_18DB4(dl ＝ 4)           ; → sub_18DBE(0FFh) × 4，**延遲**
  ds:465Bh ← ds:A5E6h
  ds:46F7h ← ds:A5E7h          ; **還原**
```

`ds:465Bh` 是**文字捲動的速度等級（0–8）**、`ds:46F7h` 是**捲動動畫的開關**
（兩個都在 [`106`](106-text-scroll.md)）。所以這一段是
「**開動畫、設速度 → 印訊息 → 停一下 → 把速度與開關還原**」。

`ds:A5C5h` 是這個速度等級的**保存格**：`0x141D6` 與 `0x142F0` 讀它當初始速度，
`sub_14193`（套用懲罰）在 `0x141F1`–`0x141F4` 把訊息印完當下的 `ds:465Bh`
寫回去。訊息捲動中玩家按 `<`／`>` 調過的速度因此會**留到下一則訊息**。

推論等級：**已確認**（四張速度表逐項對上，見 [`106`](106-text-scroll.md) §3）。

⚠ **時鐘不在這條路上。** 遊戲時鐘的「時」是 `ds:465Ah`（`docs/re/27`），
全檔五個引用點（`0x1678C`／`0x16798`／`0x17E0A`／`0x1805D`／`0x1A528`）
沒有一個在這裡。相鄰一個 byte 的兩個變數是兩件事。

remake 這一側只印訊息，**不做捲動動畫也不做延遲**——
呈現層的節奏由 remake 自己決定（[`106`](106-text-scroll.md) §7），標為刻意不移植。

## 2. 記錄 7–12 的入口：再排除三條

`docs/re/66` 定位了三階段訊息與扣血路徑，`docs/re/67` §2 解出參數，
但入口仍未解。這一輪用 `docs/re/73` 學到的方法（**改寫出來的格子**）重查：

| 路 | 結果 |
|---|---|
| 資源 0 的所有改寫來源指向 (nibble 2, 記錄 7–12) | **0 處**（掃 nibble 1 串列尾、nibble 2 的 `+0x04`／`+0x06` 與逐條件表、nibble 6 的位移 1、nibble 10 的 `+0x04`、nibble 12 的批次、nibble 4／9 的 `+0x01`／`+0x02`） |
| 別的地圖有沒有同樣的三階段記錄 | 42 張地圖裡訊息含 `warm`／`hot` 的 section 2 記錄**只有資源 0 那六筆**，全部 0 格 |
| 腳本 opcode 取 section 2 | `sub_17CB1` 的 19 個呼叫端裡，opcode 那 13 個傳的是 section **3** 與 **5**（`0x1A5D0`／`0x1A628`／`0x1A70B` 等），沒有一個傳 2 |
| 每走一步跑的 `sub_16890` | 那是**遭遇生成器**：`bl ← 0x0F` 取 section 15，在地圖上放敵人格。與高溫無關 |

三筆一組、由弱到強、兩組參數不同（減 2/4/6 d6 與 1/2/3 d6）的形狀仍然指向
「隨某個計數遞增的階段」，但**選出那一筆記錄的程式碼還沒找到**。

還沒走的：`ds:722Ah`／`722Bh` 那組計時旗標（`sub_1CB30` 用它做設施 4，
但寫入端還沒掃）、以及實機——世界地圖上要走很多步才會觸發，
`-emit-keys` 產一條夠長的沙漠路徑再送 DOSBox 是可行的下一步。

## 3. 工具坑：`export_range_refs` 的範圍是半開區間

查單一位址寫成 `0xA5C5 0xA5C5` 會回**零命中**，而 `sub_142ED` 自己就讀那個位址——
零命中與「真的沒人碰」長得一模一樣。正確寫法是 `0xA5C5 0xA5C6`。

已在 `tools/ida/export_range_refs.py` 加 guard：下界 ≥ 上界直接拒絕產出，
讓它變成明顯的錯誤而不是安靜的假零。

## 4. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/heatshow.json 0x178A0 0x18DB4 0x19720

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_range_refs.py \
  workplace/analysis/dumps/a5c5b.json 0xA5C5 0xA5C6

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/secrec.json 0x17CB1 --callers
```

## 5. 這一輪學到的（寫成規則）

- **零命中之前先做正對照。** `ds:A5C5h` 掃出 0 處，但 `sub_142ED` 的第二行
  就在讀它——**已知會命中的東西沒命中，就是工具壞了**，不是資料的結論。
  這一步只花一次重跑，而信了那個 0 會把整條懲罰路徑判成死碼。
- **同一招不會對每個缺口都有效。** 「設施格是改寫出來的」解開了商店，
  同樣的掃描套在高溫記錄上是 0 處。**方法有效之後仍要各自驗證，
  不要把上一次的解釋直接搬過來。**
- **半開區間的邊界要寫進工具本身。** 這個坑第二次踩就該讓工具拒絕跑，
  而不是靠筆記提醒——筆記會被忘記，`SystemExit` 不會。
