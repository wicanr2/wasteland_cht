# 10：第二層是 Huffman —— 容器格式與各資料檔的載入器

日期：2026-08-13 ｜ 輸入：`wl.merged.exe`（`cd5b07ea…8118`）

`docs/re/09` 留下的問題是「劇情文字在哪」。追下去發現另一件事：
`ALLPICS`／`ALLHTDS`／`END.CPA` 這幾個檔在 XOR 解密之外還有**第二層 Huffman 壓縮**。

## 1. 各資源的載入器

`sub_11445`（開資源 N）的呼叫端依 `DL` 分工：

| `DL` | 資源 | 載入器 | 位移表 |
|---:|---|---|---|
| 0 | `ALLPICS1`／`ALLPICS2` | `sub_184E8`（252 bytes） | `ds:BB18h` |
| 1 | `GAME1`／`GAME2` | `sub_183B1`（存檔寫）／`sub_1841F`（地圖讀） | `ds:BC7Ah`／`ds:BCCAh` |
| 1 | 存檔 | `sub_185E6`、`sub_18744`、`sub_18801` | 固定位移 `0x53C5`／`0x8BC7` |
| 2 | `ALLHTDS1`／`ALLHTDS2` | `sub_186B6`（142 bytes） | `ds:BDFCh` |
| 6 | `END.CPA` | `sub_1B7FE`（119 bytes） | 無（整檔） |

`END.CPA` 那支把整條鏈交代得最清楚：

```asm
0x1B7FF  mov  dl, 6
0x1B801  call sub_11445        ; 開 END.CPA
0x1B804  mov  cx, 5223h        ; 21,027 ＝ end.cpa 的完整大小
0x1B807  mov  ds:92EAh, cx
0x1B80D  call sub_11AE8        ; 解壓
0x1B812  mov  cx, 4800h        ; 18,432
0x1B815  mov  dx, 920h         ; → seg003:0x920（與 TITLE.PIC 同一個位址）
```

`0x5223` 與檔案大小、`0x4800` 與檔案前 4 bytes 的前綴都完全相同。

## 2. 容器格式

```
+0   4 bytes   解壓後大小（32-bit）
+4   3 bytes   'm' 's' 'q'
+7   1 byte    磁碟編號（0 或 1）
+8   …         Huffman 位元流
```

⚠ 注意這裡的 magic 是 `msq` ＋ **二進位 0／1**，
與 `GAME1`／`GAME2` 內 MSQ 區塊的 `msq0`／`msq1`（ASCII 數字）**不同**。
兩者只差最後一個 byte，很容易看錯。

`sub_11AE8` 讀完 8 bytes header 就驗 magic：

```asm
0x11B1A  mov  al, 6Dh          ; 'm'
0x11B1C  mov  ah, 73h          ; 's'
0x11B1E  mov  bl, 71h          ; 'q'
0x11B20  mov  bh, ds:92E3h     ; 期望的磁碟編號（來自資源表 +4 欄位）
0x11B24  cmp  ax, ds:92FDh
0x11B2A  cmp  bx, ds:92FFh
0x11B2E  jnz  short loc_11B56  ; 不符 → CF=1
```

順帶解出 `docs/re/05` 留的疑問：**資源表的 `+4` 欄位（`ds:92E3h`）就是磁碟編號**，
在這裡被拿來與檔案內的 magic 對照，確認玩家插的是正確那片。

## 3. 位元讀取器

```asm
0x11B36  lodsb
0x11B37  mov  ds:9508h, al      ; 目前的 byte
0x11B3A  mov  ds:9505h, si      ; 來源指標
0x11B3E  mov  byte ptr ds:9507h, 80h   ; 位元遮罩，從最高位開始
0x11B43  mov  ax, 9509h         ; 樹的建構區
0x11B46  call sub_11C28
```

| 位址 | 用途 |
|---|---|
| `ds:9505h` | 來源指標 |
| `ds:9507h` | 位元遮罩（`0x80` 起，MSB first） |
| `ds:9508h` | 目前 byte |
| `ds:9509h` 起 | Huffman 樹節點 |

## 4. Huffman 樹的建構（`sub_11C28`）

44 bytes 的遞迴函式，把樹直接編在位元流開頭：

```asm
0x11C29  call sub_11C54        ; 讀 1 bit
0x11C2C  jnz  short loc_11C4C  ; bit ＝ 1 → 葉節點
; ── bit ＝ 0：內部節點 ──
0x11C2E  call sub_11CA4        ; 配置節點 → DI（左）
0x11C33  call sub_11CA4        ; 配置節點 → AX（右）
0x11C39  mov  [si], di         ; +0 左子
0x11C3B  mov  [si+2], ax       ; +2 右子
0x11C41  call sub_11C28        ; 遞迴：左子樹
0x11C48  call sub_11C28        ; 遞迴：右子樹
; ── bit ＝ 1：葉節點 ──
0x11C4C  call sub_11C90        ; 讀 8 bits
0x11C50  mov  [si+4], al       ; +4 值
```

**節點是 5 bytes**：左指標 2、右指標 2、值 1。
樹以前序方式編碼：`0` 表示內部節點（接著就是左右子樹），`1` 表示葉節點（接著 8 bits 的值）。

## 5. 這解釋了什麼

`docs/re/07` §6 觀察到 42 個 MSQ 區塊全部高熵、`docs/re/09` 又發現只有名稱字串是明文——
現在有了合理解釋的第一半：**這套資料的通用管線是 XOR 加密外層 ＋ Huffman 內層**。

但**還不能直接套用到 MSQ 區塊**：`sub_1841F`（地圖載入）確實也呼叫 `sub_11AE8`（`0x184CC`，`AL=0`），
而 `AL=0` 會跳過 magic 驗證那一段。那條路徑的資料是不是也走 Huffman、
以及地圖那 2048 bytes 是解壓前還是解壓後的形式，**都還沒確認**。
`docs/re/09` 的地圖結論是對「XOR 解密後」的資料做的，那層驗證（自相關 42/42）本身仍然成立。

## 6. 未解與下一步

| 項目 | 狀態 |
|---|---|
| Huffman 解碼主迴圈（建完樹之後） | 未讀，`sub_11AE8` 之後的呼叫端還沒追 |
| `sub_11C54`（讀 1 bit）／`sub_11C90`（讀 8 bits）／`sub_11CA4`（配置節點） | 只知職責，未逐條讀 |
| `ALLPICS`／`ALLHTDS` 的內部位移表（`ds:BB18h`／`ds:BDFCh`） | 未倒出 |
| 地圖區塊是否也經過 Huffman | 未確認 |
| 劇情敘述文字 | 仍未找到。`ALLHTDS` 解壓後是最可能的位置 |

下一輪最划算的一項：把 Huffman 解碼器實作出來，解開 `ALLHTDS1`，看裡面是什麼。
