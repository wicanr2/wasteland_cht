# 05：雙模式儲存層 —— 磁片絕對磁區 vs 硬碟 DOS 檔案

日期：2026-08-13 ｜ 輸入：`wl.unpacked.exe`（`b5eb39f0…31a0`）與 `wl.merged.exe`（`cd5b07ea…8118`）

這一篇回答 `docs/re/03` §9 留下的問題：`GAME1`／`GAME2`／`ALLPICS*`／`ALLHTDS*`／`END.CPA`
這六個檔在開機時沒有被載入，也**沒有任何指令引用它們的位址**——那它們是怎麼被開啟的。

答案：檔名不在指令裡，在一張資源表的欄位裡。

## 1. 兩種模式，一個開關

`start` 讀 `info` 的第一個 byte，`'A'`／`'B'` 以外的值（散布版是 `'C'`）
會把 `ds:A414h`（`byte_27234`）設成 1（`docs/re/03` §3）。整個儲存層就靠這個旗標分流：

| 旗標 | 模式 | 開檔 | 讀取 |
|---|---|---|---|
| `ds:A414h ≠ 0` | 硬碟 | `AH=3Dh` DOS open | `AH=3Fh` DOS read |
| `ds:A414h = 0` | 磁片 | 不開檔，只設定磁區座標 | `int 25h` 絕對磁區讀 ＋ 自建緩衝 |

## 2. `sub_11445`：開啟資源 N

入口參數是 `DL` ＝ 資源編號。

```asm
0x11445  pushf
0x11446  mov  byte ptr ds:92E7h, 0     ; 緩衝狀態歸零
0x1144B  mov  word ptr ds:92EFh, 0
0x11451  mov  word ptr ds:92E8h, 0
0x11459  mov  bx, dx
0x1145B  cmp  bx, 6
0x1145E  jnb  short loc_1146A
0x11460  cmp  byte ptr ds:9168h, 40h   ; 磁碟切換旗標
0x11465  jnz  short loc_1146A
0x11467  add  bx, 3                    ; 0/1/2 → 3/4/5
0x1146A  shl  bx, 1 ×3                 ; ×8 ＝ 每筆 8 bytes
0x11470  add  bx, 92AAh                ; ＋ 資源表基址
0x11474  cmp  byte ptr ds:0A414h, 0
0x11479  jnz  short loc_114B0          ; → 硬碟路徑
; ── 磁片路徑：把表裡的磁區座標抄進工作變數 ──
0x1147B  mov  al, [bx]      → ds:92E2h  ; 磁碟機
0x1148B  mov  al, [bx+1]    → ds:92E4h  ; cylinder
0x11491  mov  al, [bx+2]    → ds:92E5h  ; head
0x11497  mov  al, [bx+3]    → ds:92E6h  ; sector
0x1149D  mov  al, [bx+4]    → ds:92E3h  ; 用途未解
; ── 硬碟路徑 ──
0x114B4  add  bx, 6
0x114B7  mov  dx, [bx]                 ; 檔名指標（word）
0x114B9  mov  al, 2
0x114BB  jmp  sub_11384                ; DOS open
```

磁片模式下 `ds:92E2h` 還會被覆寫：若 `ds:A415h` ＝ 0，改用 `ds:92F4h`（開機取得的預設磁碟機）。

`bx < 6` 且 `ds:9168h == 0x40` 時索引 +3，正好把 Disk 1 的三個資源換成 Disk 2 的三個（見 §3）。

## 3. 資源表（`ds:92AAh` ＝ 線性 `0x260CA`）

每筆 8 bytes，實際使用 8 筆：

| idx | +0 drive | +1 cyl | +2 head | +3 sector | +4 | +6 檔名指標 | 檔名 |
|---:|---:|---:|---:|---:|---:|---|---|
| 0 | 0 | 0 | 0 | 0 | 0 | `918Dh` | `ALLPICS1` |
| 1 | 0 | 0 | 0 | 0 | 0 | `919Fh` | `GAME1` |
| 2 | 0 | 0 | 0 | 0 | 0 | `91ABh` | `ALLHTDS1` |
| 3 | 1 | 0 | 0 | 0 | 1 | `9196h` | `ALLPICS2` |
| 4 | 1 | 0 | 0 | 0 | 1 | `91A5h` | `GAME2` |
| 5 | 1 | 0 | 0 | 0 | 1 | `91B4h` | `ALLHTDS2` |
| 6 | 0 | 0 | 0 | 0 | 0 | `91BDh` | `END.CPA` |
| 7 | 1 | 1 | 0 | 1 | 3 | `0000h` | **無檔名** |

- idx 0–2 是 Scenario Disk 1 的三個資源，idx 3–5 是 Disk 2 的同名對應。
  §2 的 `+3` 就是在這兩組之間切換。
- **idx 7 沒有檔名，只有磁區座標**（cylinder 1、head 0、sector 1）。
  這是唯一一筆「只能用磁區路徑存取」的資源，而 `int 26h`（絕對磁區寫）只有一個呼叫端——
  **這一筆很可能是存檔區**。硬碟模式下它的檔名指標是 0，那條路徑會怎麼走還沒追，
  在追出來之前不要斷言硬碟版怎麼存檔。
- `+4` 欄位（0／1／3）進 `ds:92E3h`，用途未解；值與磁片編號相關但沒有證據支持它就是磁片號。

## 4. `sub_115E5`：讀取 N bytes（兩條路）

```asm
0x115EE  cmp  byte_27234, 0
0x115F3  jz   short loc_11600
0x115F5  mov  bx, word_25F94        ; 目前 handle
0x115FD  jmp  sub_113B2             ; ── 硬碟：直接 DOS read
```

磁片路徑則自己做緩衝與磁區推進：

```asm
0x1161D  cmp  byte ptr ds:92E7h, 0  ; 緩衝裡還有資料嗎
0x11625  call sub_116AC             ; 沒有 → int 25h 讀一批進來
0x11629  mov  si, ds:472Fh          ; 緩衝區
0x1162D  add  si, ds:92E8h          ; ＋ 已消耗位移
0x1164B  rep movsw                  ; 複製給呼叫端
0x11673  mov  dl, ds:92E7h
0x11677  add  dl, ds:92E6h          ; sector += 這次讀的磁區數
0x1167F  cmp  dl, 9                 ; 一磁軌 9 個磁區
0x11687  mov  byte ptr ds:92E6h, 0
0x1168C  inc  byte ptr ds:92E5h     ; head++
0x11690  cmp  byte ptr ds:92E5h, 2  ; 兩面
0x11697  mov  byte ptr ds:92E5h, 0
0x1169C  inc  byte ptr ds:92E4h     ; cylinder++
```

「9 磁區／磁軌、2 面」與 `sub_116AC` 的位址計算一致（`docs/re/03` §8），
確認原版媒體是 **360 KB 5.25 吋磁片**。

超出範圍時走的是 `int 3`（`0x1166C`、`0x11686`）——原版留下的除錯陷阱，不是正常路徑。

## 5. 呼叫端

`sub_11445`（開資源）有 11 個呼叫端，`sub_115E5`（讀）有 7 個，
分布在 `sub_183B1`／`sub_1841F`／`sub_184E8`／`sub_18744`／`sub_1B7FE` 等——
這些就是各類資產的載入器，是下一輪的分析對象。

## 6. ⚠ 方法論：指令掃描找不到「表驅動」的引用

這七個檔名在 xref 圖與立即數掃描裡都是**零命中**，即使把 overlay 疊上去重掃也一樣。
零命中在這裡完全不代表「沒人用」——檔名位址是**資料表的欄位值**，
從來沒有以立即數形式出現在任何一條指令裡。

找到它的路徑不是搜尋，而是從 sink 往回：
`int 21h AH=3Dh` → 包裝函式 `sub_11384` → 列出它的呼叫端 →
除了 `start` 的七次固定載入，只剩一個 `sub_11445` → 讀它就看到 `[bx+6]`。

**下「這個東西沒被使用」的結論前，先確認自己搜的是不是它實際被引用的形式。**

## 7. 未解與下一步

| 項目 | 狀態 |
|---|---|
| idx 7 在硬碟模式下如何存取（檔名指標為 0） | 未追 |
| 資源表 `+4` 欄位（`ds:92E3h`）的語意 | 未解 |
| `ds:9168h == 0x40` 這個磁碟切換旗標由誰設定 | 未追 |
| `GAME1`／`GAME2` 的內部結構 | 未解，需要先看載入器怎麼解析 |
| `sub_183B1`／`sub_1841F`／`sub_184E8`／`sub_18744` 這幾個載入器 | 未解 |

## 8. 重跑方式

```sh
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
    workplace/analysis/dumps/storage-layer.json \
    0x11445 0x115E5 0x118C3 0x11730 0x118D2 --callers
```

資源表可直接從解包映像讀：線性 `0x260CA` ＝ 檔案位移 `0x1617A`，8 bytes × 8 筆。
