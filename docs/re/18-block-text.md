# 18：地圖區塊的文字 —— 4,401 條全部解出

日期：2026-08-14 ｜ 對應盤點 **A6** 的文字部分、**E3**、中文化的主體

輸入：`wl.merged.exe` SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 的 SHA-256 見 `docs/re/01`。

---

## 1. 結論

42 個 MSQ 區塊**每一個都有自己的 5-bit 打包字串表**，格式與執行檔那九張相同
（`docs/re/17`），但**每個區塊有自己的字元對照表**。

全部解出來是 **4,445 個字串槽、4,401 條非空**，最長 696 個字元。內容是遊戲的敘述文字：
地點進出訊息、物件描述、NPC 對話、劇情段落、製作名單。

加上執行檔的 444 個槽（426 條非空），原版的文字總量是 **4,889 個槽、4,827 條非空**。

## 2. 為什麼之前看起來是亂碼

關鍵在**加密長度**。`sub_11A59` 不是把整個區塊解密，只解**區塊標頭第一個 word**
那麼多 bytes：

```
0x11A5E  8b0eb046    mov  cx, ds:46B0h    ; ＝ 標頭位置 P
0x11A62  41 41       inc  cx / inc cx     ; 先解 P+2 bytes，才讀得到長度
0x11A64  be0000      mov  si, 0
0x11A67  3004        xor  [si], al        ; 第一段
…
0x11A6E  4e 4e       dec  si / dec  si    ; si ← P
0x11A70  8b0c        mov  cx, [si]        ; ← 加密區總長度 L
0x11A73  2b0eb046    sub  cx, ds:46B0h
0x11A79  49 49       dec  cx / dec cx     ; cx ＝ L − P − 2
0x11A7B  3004        xor  [si], al        ; 第二段
```

而**同一個 word 又被 `sub_1790B` 當成字串表的基址**（寫進 `ds:4692h`）。
兩件事合起來就是：

> **字串表從加密區結束的地方開始，而且不加密。**

先前把整個區塊都拿去 XOR，等於把字串表整段攪爛，看起來就是高熵資料
（`docs/re/17` §6 把這一項列為未解，就是卡在這裡）。

### 2.1 兩條獨立推導互相佐證

| 推導方式 | 結果 |
|---|---|
| 讀區塊標頭第一個 word | L |
| 從頭累加負和，找 checksum 相符的位置 | 同一個 L |

**41／42 個區塊兩者完全相同**（第 42 個是 `game2` 資源 41，字串表位移表有一項
指到區塊外，解到一半停）。兩條推導的機制完全無關，同時吻合可以當已確認用。

## 3. 完整的區塊佈局

```
+0x0000                     地圖（P bytes：0x600 或 0x1800）
+P                          記錄區標頭（0x5C bytes），第一個 word ＝ L
+P+0x5C                     各 section（記錄）
+L                          字串表（不加密）
   +0x00 … +0x3B            60 bytes 字元對照表（每個區塊自己一份）
   +0x3C …                  16-bit 位移表，每 4 個字串一項
+（read_len − 6）            Huffman 尾段（tile 圖形，`docs/re/12`）
+（block_len − 6）           區塊結束
```

三個長度的關係因此完全對上：

| 長度 | 來源 | 涵蓋到 |
|---|---|---|
| `L` | 區塊標頭第一個 word | 加密區（地圖 ＋ 記錄） |
| `ds:BD22h` − 6 | 載入器讀取量 | 加密區 ＋ 字串表 |
| `ds:BD86h` − 6 | 區塊總長度 | 全部，含 Huffman 尾段 |

`docs/re/08` §3 把「`ds:BD22h` 讀取量與區塊長度為何不同」列為未解——
差額就是 Huffman 尾段（`docs/re/12` 已解），而 `L` 與讀取量的差額就是字串表。

## 4. 解出來的內容

```bash
python3 tools/decode_block_text.py \
  workplace/analysis/unpacked/wl.merged.exe \
  workplace/orig/wastland/game1 workplace/orig/wastland/game2 \
  docs/re/generated/ida94/block-strings.json
```

| 統計 | 值 |
|---|---:|
| 有文字的區塊 | 42／42 |
| 字串槽 | 4,445 |
| 非空字串 | 4,401 |
| 最長一條 | 696 字元 |
| 每個區塊的字串數 | 48 – 192 |

內容的形狀（節錄，只列足以說明類型的短句）：

| 類型 | 例 |
|---|---|
| 地點進出 | `Entering Quartz.`、`Entering Highpool.`、`Leaving Quartz.` |
| 物件與環境 | `The door is locked.`、`An old abandoned shack.` |
| 技能檢定結果 | `The door didn't open.`、`You swim next to the edge.` |
| 敘述段落 | `As you exit the from the tunnel, you startle a sleeping hobo and he runs off.` |
| 讀取類 | `You see some graffiti carved into the table. It says: ` |
| 製作名單 | 資源 33 有一條 696 字元的完整 credits（`IBM VERSION: Michael Quarles`…） |

`\r`（`0x0D`）是段內換行，`0x0B`／`0x0C` 等控制碼與執行檔那批相同
（`docs/re/14` §4.1）。

## 5. 對中文化的意義

- **這 4,401 條是中文化的主體**，比段落書（162 段）大一個數量級。
- 每個區塊一張字元對照表，**抽取與回寫都要逐區塊處理**，不能共用一張表。
- 重製版不需要沿用 5-bit 打包（沒有記憶體壓力），但**抽原文一定要用這支解碼器**。
- 字串表不加密這件事，讓「只改文字不動其他」在原版上理論可行；
  不過本專案是重寫引擎，不走改檔路線。

## 6. 還沒解的

- **字串索引怎麼對到地圖上的位置**：記錄本體（`ds:46AEh`）的 `+0x03` 有 34 處
  存取並餵給 `sub_178A0`（印訊息），形狀像字串編號，但欄位語意還沒逐一確認
  （`docs/re/16` §4）。
- 段落書 162 段與這 4,401 條的關係：遊戲內既然有敘述，
  紙本段落書的角色要重新界定（`docs/re/12` §4）。
- `game2` 資源 41 的位移表有一項指到區塊外，最後幾組沒解出來。

## 7. 可重跑的完整指令

```bash
python3 tools/decode_block_text.py \
  workplace/analysis/unpacked/wl.merged.exe \
  workplace/orig/wastland/game1 workplace/orig/wastland/game2 \
  docs/re/generated/ida94/block-strings.json

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/block-text.json \
  0x11A59 0x1790B 0x1841F --callers
```

## 8. 這一輪學到的（寫成規則）

- **同一個欄位可以有兩個用途，而那正是解法。** 區塊標頭第一個 word 既是
  「加密區長度」也是「字串表基址」——因為設計上它們本來就是同一個邊界。
  分別追這兩條線索時各自卡住，合起來才通。
- **「解密整個檔案」是一個沒有根據的預設。** 原版只解到某個長度，
  而多解的那一段會被 XOR 破壞成高熵資料——**症狀是「這裡是壓縮／未知格式」，
  和真的壓縮長得一模一樣**。看到高熵先問「這一段真的在加密範圍內嗎」。
