
# 11：Huffman 解碼器實作，四個資料檔全部解開

日期：2026-08-13 ｜ 狀態：**已確認**（解出的長度與檔案宣告的完全相同，且整個檔案 100% 用完）

## 1. 演算法

### 位元讀取（`sub_11C54`）

```asm
0x11C54  mov  ax, ds:9507h     ; 一次讀兩個 byte：AL ＝ 遮罩、AH ＝ 目前 byte
0x11C57  or   al, al
0x11C59  jz   short loc_11C67  ; 遮罩用完 → 取下一個 byte
0x11C5B  and  ah, al           ; 測試該位元
0x11C5D  shr  al, 1            ; 遮罩右移
0x11C5F  mov  ds:9507h, al
```

遮罩從 `0x80` 起、測試後右移，所以是 **MSB first**。

### 節點配置（`sub_11CA4`）

```asm
0x11CA4  mov  ax, ds:9301h     ; 已配置數
0x11CA7  cmp  ax, 300h         ; 上限 768
0x11CB4  shl  ax, 1 ×2
0x11CB8  add  ax, bx           ; ×5
0x11CBA  add  ax, 950Eh        ; 節點區基址
```

節點 5 bytes：`+0` 左指標、`+2` 右指標、`+4` 值。根節點固定在 `ds:9509h`，
配置從 `ds:950Eh` 起（剛好是根的下一個）。

### 建樹（`sub_11C28`，前序）

```
bit ＝ 1 → 葉節點，接著讀 8 bits 當值
bit ＝ 0 → 內部節點，配置左右子節點，遞迴左子樹 → 讀掉一個分隔 bit → 遞迴右子樹
```

那個分隔 bit 是 `0x11C44` 的 `call sub_11C54`，結果沒有被使用。
**實作時必須照做**——少讀一個 bit 整條位元流就會錯位。

### 解碼（`sub_11B83`）

```asm
0x11BBD  mov  di, 9509h        ; 從根開始
0x11BC3  cmp  word ptr [di], 0
0x11BC6  jz   short loc_11BDD  ; 左指標為 0 → 葉節點
0x11BCC  test ah, al
0x11BCE  jnz  short loc_11BD6
0x11BD0  mov  di, [di]         ; bit ＝ 0 → 左
0x11BD6  mov  di, [di+2]       ; bit ＝ 1 → 右
0x11BE0  mov  al, [di+4]       ; 葉節點的值
0x11BE7  stosb                 ; 輸出
```

葉節點的判定是**左指標為 0**，不是額外的旗標。

## 2. 驗證

`tools/huffman.py` 對四個檔案執行：

| 檔案 | 第一塊解出 | 檔案宣告 | 子區塊數 | 位元流用掉 |
|---|---:|---:|---:|---|
| `end.cpa` | 18,432 | 18,432 | 1 | — |
| `allhtds1` | 8,448 | 8,448 | 4 | 34,307／34,307 |
| `allhtds2` | 16,256 | 16,256 | 5 | 39,230／39,230 |
| `allpics1` | 4,032 | 4,032 | 66 | 105,866／105,866 |
| `allpics2` | 4,032 | 4,032 | 98 | 133,433／133,433 |

**每個子區塊解出的長度都精確等於它自己宣告的大小，而且整個檔案一個 byte 不剩地用完。**
解碼器若有任何偏差，位元流會錯位，長度不可能連續 173 次都對。

## 3. 子區塊怎麼串接

檔案是多個容器直接相接，**下一個的起點就是上一個位元流消耗完的位置**——
沒有目錄、沒有對齊填充。

先前試過「用 4 bytes 前綴當步進」，鏈接不起來（`docs/re/10` 之前的嘗試）：
前綴是**解壓後**的大小，而串接要的是**壓縮後**的長度，兩者無關。
實測對照：`allhtds1` 第一塊消耗到 5,122，而檔案裡下一個 `msq` 出現在 5,126，
正好差 4 bytes 的前綴。

## 4. 解開之後是什麼

| 檔案 | 內容形狀 |
|---|---|
| `allpics1`／`allpics2` | **交替兩種區塊**：4,032 bytes 的一種（可列印 0.04–0.14，大量重複 nibble）與 600–2,500 bytes 的一種（可列印 0.23–0.40，開頭像一串 16-bit 小數值）。前者是圖，後者疑似該圖的參數表 |
| `allhtds1`／`allhtds2` | 4–5 個大區塊（8K–21K），可列印 0.10–0.26，同樣像圖形資料 |
| `end.cpa` | 18,432 bytes，載入到 `seg003:0x920`，與 `TITLE.PIC` 同一個位址與大小 |

`4032 = 96 × 84 ÷ 2`，符合 4bpp 圖形，但**尺寸未經驗證**，先不當結論。

## 5. 劇情敘述文字仍然沒找到

到目前為止已解開的全部內容：MSQ 區塊（地圖 ＋ 名稱字串）、`ALLPICS`、`ALLHTDS`、`END.CPA`，
**沒有一處是敘述文字**。

下一個入口很明確：`sub_1841F`（地圖載入器）在 XOR 解密之後**還呼叫了一次解壓**——

```asm
0x184C5  call sub_11A59        ; XOR 解密
0x184CC  mov  al, 0
0x184CC  call sub_11AE8        ; 建 Huffman 樹（AL=0 跳過 magic 驗證）
0x184CF  mov  cx, 1000h        ; 4,096 bytes
0x184D2  mov  dx, 3448h
0x184D5  call sub_11B83        ; 解出到 ds:3448h
```

也就是說 **MSQ 區塊裡還有一段 Huffman 壓縮資料**，解出 4,096 bytes。
`docs/re/09` 說的「前 2048 bytes 是地圖」對的是 XOR 解密後的形式，那層仍然成立；
壓縮段在哪個位移、與地圖如何並存，還沒讀出來——`sub_11AE8` 的來源指標是
`ds:472Fh + ds:92E8h`（磁碟緩衝），在硬碟模式下這條路徑要先讀懂 `sub_118D2`。

## 6. 未解與下一步

| 項目 | 狀態 |
|---|---|
| MSQ 區塊內 Huffman 段的位置 | 未解——**這是找文字的下一個入口** |
| `sub_118D2`（緩衝補充） | 未讀，硬碟／磁片兩種模式的差異卡在這 |
| `allpics` 小區塊的欄位語意 | 未解 |
| 圖形的實際尺寸與像素格式 | 未驗證 |
| `ALLPICS`／`ALLHTDS` 的位移表（`ds:BB18h`／`ds:BDFCh`） | 未倒出 |

## 7. 重跑方式

```sh
docker run --rm --network none --memory 1g --cpus 1 --pids-limit 256 \
  --user "$(id -u):$(id -g)" -v "$PWD:/workspace" -w /workspace \
  ida-pro-9.4-idapython:py312-v1 /opt/venv/bin/python3 \
  tools/huffman.py workplace/orig/wastland/allhtds1 --all
```
