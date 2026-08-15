# 95：主選單只有一個選項，而且沒有「讀檔」

日期：2026-08-15 ｜ 接 `docs/re/03`（開機序列）、`docs/re/91`（指令列的雙表結構）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`WORKLIST` 把 5.3 寫成「主選單（新遊戲／讀檔）」。讀完之後這個描述要改：
**原版的主選單整個只有 `Start` 一個字，第二支處理程式是一支 `retn`。**

---

## 1. 開機序列的尾巴

`start`（`0x110B6`，`docs/re/03`）載完素材、裝好滑鼠之後：

```
0x11332  call sub_1CB75          ; 掛 int 08h ＋ 設 8253 → **音效驅動**
0x11337  xor bx, bx / xor dx, dx
0x1133B  call sub_161C0          ; ← 進主流程
```

`sub_1CB75` 把 `int 08h` 換成自己的處理程式，再對 8253 通道 2 寫
`0x24E3`（`out 43h, 0B6h` 之後兩次 `out 40h`）——與 `docs/re/44` 的音效格式對上，
**音效驅動是在開機序列的最後一步才裝的**。

## 2. `sub_161C0`：選單 ＋ 片頭，兩者共存

```
0x161EB  sub_1001B                    ; 換畫面
0x16207  call sub_1630C               ; ← **主選單**（尾呼叫 sub_173D2 等按鍵）
0x1620A  call sub_173D2               ; 再等一次
0x1620D  call sub_166D3               ; 逐 byte 送 ds:AA4Dh 的序列（不是文字）
…
loc_1623F:
  sub_162C7；sub_16385(1)；sub_162CF(3)
  sub_162C7；sub_16385(2)；sub_162CF(3)
  …（al ＝ 1,2,3,5,6,7,9,0Ah…11h）
0x162C4  jmp loc_1623F                ; **無限循環**
```

後半段是 attract mode：選單等不到按鍵就一直播。
`sub_16385` ＝ `sub_1785E` ＋ `sub_16390` ＋ `sub_17923`，
`sub_162CF(dl)` 是 `dl` 次的延遲迴圈。

## 3. 主選單本體：`sub_1630C`

與地圖指令列（`sub_16C7C`，`docs/re/91` §1）**是同一套雙表結構**：

```
ds:468Dh ← 0A6FFh      ; [A6FFh] ＝ A6F4h → 標籤字串
ds:468Fh ← 0A701h      ; [A701h] ＝ A6FBh → 處理程式陣列
ds:468Ah ← 0；ds:4688h ← 0
ds:468Ch ← 18h
ds:4689h ← 1           ; **最大索引 1** → 兩項
jmp sub_173D2          ; 等按鍵
```

倒出來的內容：

```
ds:A6F4h = "Start\0"
ds:A6FBh = 30 63 2F 63   → 0x16330、0x1632F
```

| 索引 | 標籤 | 處理程式 | 做什麼 |
|---:|---|---|---|
| 0 | `Start` | `0x16330` | `sub_173D2` → `sub_16356` → `ds:46C5h ← 0FFh` → `sub_185E6` → `sub_18744` → `ds:46E0h ← 0FFh` → `jmp loc_16B31`（**地圖迴圈**，`docs/re/91` §1 的呼叫端就在那裡）|
| 1 | —— | `0x1632F` | **`retn`**，一條指令。回開機序列繼續播片頭 |

`ds:46E0h ← 0FFh` 這一句同時解釋了 `ENC` 的一個分支：
`docs/re/94` §2 的「是主地圖就直接開打」比的正是 `ds:46E0h`。

推論等級：**已確認**（兩張表是從映像直接倒出來的；兩支處理程式逐指令讀完）。

## 4. 沒有「新遊戲／讀檔」這回事

標籤字串只有 `Start`，處理程式只有兩支，其中一支是空的。存檔就是
`GAME1`／`GAME2` 本身（`docs/re/16`），所以「開始」＝ **接著磁碟上的狀態走**。

角色的建立與刪除在 Ranger Center 的 `CREATE`／`DELETE`／`PLAY`
（設施 3，`docs/re/72` §3）——**不在主選單**。
1988 年的存檔設計沒有「檔位」概念，重製版照做就不必發明一套存檔管理。

## 5. remake 這一側

**已接**：`Scene.BeginTitle`／`updateTitle`／`drawTitle`
（`internal/play/mainmenu.go`），`cmd/wasteland -mode play` 從標題畫面開場。

- 按 `S` 進地圖；**其餘的鍵什麼都不做**，包括 ESC——照索引 1 那支 `retn`。
- 截圖與測試要跳過標題時用 `-skip-title`。
- **attract mode 沒做**：`sub_16385` 那一長串播的是什麼還沒讀
  （`sub_1785E`／`sub_16390`／`sub_17923` 三支都未解），不猜。
  標題畫面停著等按鍵，其餘照原版。

## 6. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/mainmenu.json 0x1CB75 0x161C0
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/menu3.json 0x1632F 0x16330

python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xA6FF 2
python3 tools/dump_word_table.py workplace/analysis/unpacked/wl.merged.exe 0xA701 2 --code

tools/go.sh test ./internal/play/ -run TestTitleStartsOnlyOnStartKey -v
```

⚠ 手算 `ds:` 位移對應的檔案位移時，**線性位址與檔案位移是兩件事**。
`dump_word_table.py` 的輸出同時印兩個（`ds:A6FFh` ＝ 線性 `0x2751F` ＝
檔案位移 `0x175CF`）；拿線性位址去索引檔案 bytes 會讀到一整片零，
而**一整片零和「這裡是 BSS」長得一模一樣**——會得出「這張表是空的」這種錯結論。

## 7. 這一輪學到的（寫成規則）

- **工項的描述本身要當成待驗證的斷言。** 「主選單（新遊戲／讀檔）」這個名字
  是照現代遊戲的習慣寫的，原版根本沒有那兩項。
  **照著工項名稱做，就會做出原版沒有的東西**——而且不會有任何測試抓到。
- **同一套 UI 框架會在好幾個地方出現。** 主選單與地圖指令列共用
  `ds:468Dh`／`ds:468Fh`／`ds:4689h` 這組全域。
  解過一次之後第二次只要倒兩張表，五分鐘的事。
  **看到熟悉的全域被設值，先假設是同一套框架。**
- **一條指令的處理程式也是答案。** `0x1632F` 只有一個 `retn`，
  它回答的是「選第二項會怎樣」——不是漏讀，是原版就這樣。
