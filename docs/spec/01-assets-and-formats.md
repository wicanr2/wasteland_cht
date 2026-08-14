# 01：資產與資料格式規格 — **READY**

範圍：把玩家自備的原版檔案解成程式可用的資料——資源定址、解密、解壓、
文字、字型、圖片、圖磚、地圖。對應 `internal/assets`（純解碼，不認識 Ebiten）。

日期：2026-08-14

---

## 1. 依據

| 條目 | 推論等級 | 來源 |
|---|---|---|
| 執行檔解包（EXEPACK） | 已確認 | [`re/02`](../re/02-exepack-unpack.md) |
| 資源目錄與位移表 | 已確認 | [`re/06`](../re/06-resource-directory.md)、[`re/07`](../re/07-msq-blocks.md) |
| MSQ 加密（42/42） | 已確認 | [`re/08`](../re/08-msq-encryption.md)、[`re/18`](../re/18-block-text.md) §2 |
| Huffman 解壓（173 塊） | 已確認 | [`re/10`](../re/10-huffman-compression.md)、[`re/11`](../re/11-huffman-decoder.md) |
| 5-bit 打包文字（4,827 條非空） | 已確認 | [`re/17`](../re/17-packed-text.md)、[`re/18`](../re/18-block-text.md) |
| 兩套字型 | 已確認 | [`re/14`](../re/14-fonts-and-text-encoding.md) |
| 圖片格式（82 張 ＋ 標題） | 已確認 | [`re/23`](../re/23-picture-format.md) |
| 圖磚與地圖三層（42/42、9/9） | 已確認 | [`re/24`](../re/24-map-layers-and-tiles.md) |
| 記錄區標頭欄位 | 已確認（列出的那些） | [`re/16`](../re/16-msq-block-layout.md)、[`re/27`](../re/27-game-clock.md) §8 |

位址換算與速查表：[`re/00-master-index.md`](../re/00-master-index.md) §1、§5。

## 2. 規格本體

### 2.1 輸入檔案與身分驗證

玩家自備 20 個原版檔。載入前**必須驗 SHA-256**（清單在 [`re/01`](../re/01-binary-identity.md)），
不在清單裡就拒絕載入並明確告知——這是避免「用錯版本卻默默解出垃圾」的唯一防線。

`wl.exe` 是 EXEPACK 打包版，**所有位址都以解包後的映像為準**。
執行期需要從執行檔取出的東西有三類：主文字字型、九張字串表、各種常數表。
建議在建置期用 `tools/` 產出一份中介檔（JSON／二進位皆可），
執行期不再解析原版執行檔——但**中介檔不得隨專案散布**（`CLAUDE.md` §7）。

### 2.2 資源定址（`GAME1`／`GAME2`）

```
資源目錄     ds:BEC9h  逐 byte 到 0xFF 結束；高 2 bits：0x80 → GAME1、0x40 → GAME2
區塊總長度   ds:BD86h  每個資源一個 word
讀取量       ds:BD22h  每個資源一個 word（＝ 交給 XOR 解密的長度）
地圖大小選擇 ds:BF1Ch  每個資源一個 byte：0x40 → 地圖區 0x1800，其餘 → 0x600
```

同一個檔案內的區塊**首尾相接**：第 n 個區塊的位移 ＝ 前面所有區塊總長度的和。
共 42 個 MSQ 區塊（`game1` 20、`game2` 22）。

### 2.3 MSQ 區塊

```
+0x00  magic  'msq0'（game1）／'msq1'（game2）
+0x04  checksum（16-bit）
+0x06  加密段開始
```

解密：

```go
key := byte(checksum&0xFF) ^ byte(checksum>>8)
for i := range body {
    body[i] ^= key
    key += 0x1F
}
```

**只解 `word(記錄區標頭第一個 word)` 那麼長**，不是整個區塊
（[`re/18`](../re/18-block-text.md) §2）。多解的部分會被破壞成高熵資料，
症狀跟「這段是壓縮的」一模一樣。

解密後的佈局：

```
0            地圖第 1 層：D²÷2 bytes（4 bits／格）
D²÷2         地圖第 2 層：D² bytes（1 byte／格）
P ＝ D²×1.5  記錄區標頭 0x5C bytes（P 就是 0x600 或 0x1800）
P+0x5C       各 section
L            字串表（不加密；L ＝ 標頭第一個 word）
讀取量−6     Huffman 尾段 ＝ 地圖第 3 層（D² bytes）
```

### 2.4 記錄區標頭（自 P 起算）

| 位移 | 內容 |
|---|---|
| `+0x00` | 加密長度 L，同時是字串表基址 |
| `+0x02` | 明文 NUL 分隔的敵人名表位移 |
| `+0x04` | 8 bytes 一筆的記錄表位移 |
| `+0x06`…`+0x28` | section 位移表（型別 → 位移，要查執行檔 `ds:B9E0h` 的 24 項表） |
| `+0x2C` | **地圖邊長 D**（實測只有 32 與 64） |
| `+0x2F` | 遭遇機率分母 |
| `+0x30` | **圖磚組編號**（0–8） |
| `+0x31` | 遭遇種類數 |
| `+0x32` | 遭遇槽位上限 |
| `+0x33` | **地圖範圍外要畫的圖形編號** |
| `+0x34`–`+0x35` | 走一步推進的時間（分鐘 × 256） |
| `+0x36` | 走一步推進的刻 |

section 的形狀（型別 3、5 已驗證）：`[16-bit 指標陣列][記錄本體]`，
**陣列長度不另存——第一個非空指標指到哪，陣列就到哪為止**。

### 2.5 Huffman

容器與位元流見 [`re/10`](../re/10-huffman-compression.md)；
參考實作 `tools/huffman.py`（173 個子區塊全部解出，長度精確吻合）。
兩個必須照抄的細節：

- **無 magic 的尾段也是同一套**，載入器走的是跳過 magic 驗證的那條路徑
- 一個檔案裡多個子區塊**首尾相接**，要一路解到檔案用完

### 2.6 5-bit 打包文字

```
表基址 +0x00 … +0x3B   60 bytes 字元對照表（符號 → ASCII，0 ＝ 結束）
表基址 +0x3C …         16-bit 位移表，每 4 個字串一項，位移相對於 +0x3C
```

取第 N 個字串：跳到第 `N>>2` 項位移，再解掉 `N&3` 個。符號 LSB-first 取 5 bits；
`0x1E` ＝ 下一個字元轉大寫、`0x1F` ＝ escape（讀下一個符號 ＋ 0x1E）。
**每張表有自己的字元對照表，不能共用。**

| 來源 | 張數 | 字串槽 | 非空 |
|---|---:|---:|---:|
| 執行檔（基址見 `re/00-master-index` §5.1） | 9 | 444 | 426 |
| 42 個 MSQ 區塊（各一張） | 42 | 4,445 | 4,401 |
| 合計 | 51 | **4,889** | **4,827** |

字串內的機制：`\n` 分隔字根／單數字尾／複數字尾；`0x0B` 插入角色名字；
`0x0C` 夾 his/her 做性別選字；`0x0D` 段內換行。
**中文化的原文抽取一律走這條路**（`tools/decode_text.py`、`tools/decode_block_text.py`）。

### 2.7 字型

| 字型 | 位置 | 一個字 | 格式 | 索引 |
|---|---|---:|---|---|
| 主文字 | 執行檔 `seg003:0xCA60` | 8 bytes | 8×8 單色 | ASCII − 0x20（128 字，96–99 是四個箭頭） |
| 彩色選單 | `COLORF.FNT` | 32 bytes | 8×8、EGA 4 平面、平面連續存放 | 172 字；`0x18` 暖色／`0x34` 冷色兩組同形字模 |

### 2.8 圖片與圖磚（同一套格式）

```
packed 4bpp：一個 byte 兩個像素，高 4 位在左
列間 XOR delta：out[n+stride] ^= out[n]（以 word 為單位，n 由 0 每次 +2）
**回看距離 stride 就是一列的 byte 數**
```

| 來源 | 一筆大小 | stride | 尺寸 | 數量 |
|---|---:|---:|---|---:|
| `ALLPICS1/2` 的圖片子區塊 | 4,032 | 48 | 96 × 84 | 82（33 ＋ 49） |
| `TITLE.PIC` | 18,432 | 144 | 288 × 128 | 1 |
| `ALLHTDS1/2` 的圖磚 | 128 | 8 | **16 × 16** | 9 組共 1,051 張 |

- `ALLPICS` 的子區塊嚴格交錯：一張圖 ＋ 一段變動長度參數區（**參數區未解**，見 §3）
- 圖磚**每張各自有 8 bytes 種子列**，delta 不跨圖磚
- 原版載入圖磚時會轉成 EGA 4 平面；**remake 不需要轉**，直接用 packed 4bpp 解到 RGBA

### 2.9 地圖三層與圖形編號

```
第 1 層（位移 0，4 bits／格）    這一格屬於哪一種 section（0 ＝ 沒東西）
第 2 層（位移 D²÷2，1 byte／格） 該 section 裡的第幾筆記錄
第 3 層（Huffman 尾段，1 byte）  畫面上的圖形編號
```

第 1 層取值：偶數行取高 4 位、奇數行取低 4 位，列基址 ＝ 列號 × D÷2。

圖形編號的解讀（[`re/24`](../re/24-map-layers-and-tiles.md) §2.3）：

```
0–9   IC0_9.WLF 的十個 16×16 圖形
≥10   該地圖圖磚組的第 (值 − 10) 張
```

地圖範圍外的格子畫標頭 `+0x33` 指定的圖形編號。

### 2.10 建議的 Go 介面

```go
package assets

type Rom struct{ ... }                    // 驗過 SHA-256 的一份原版資料

func Open(dir string) (*Rom, error)       // 驗雜湊、建索引，失敗要說明是哪個檔
func (r *Rom) Block(id int) (*Block, error)   // 解密 ＋ 解壓一個 MSQ 區塊

type Block struct {
    Dim      int      // 邊長 D
    Tileset  int      // 標頭 +0x30
    Terrain  []byte   // 第 1 層，已展開成一格一個 byte（值 0–15）
    Record   []byte   // 第 2 層
    Graphic  []byte   // 第 3 層
    Header   [0x5C]byte
    Strings  []string // 該區塊的字串表
    Raw      []byte   // 解密後的原始 bytes，未解區域原樣保留
}

func (r *Rom) Tileset(n int) ([]*image.RGBA, error) // 16×16 × 66–163 張
func (r *Rom) Picture(n int) (*image.RGBA, error)   // 96×84
func (r *Rom) Title() (*image.RGBA, error)          // 288×128
func (r *Rom) ExeStrings() [][]string               // 九張表
func (r *Rom) FontMain() *Font                      // 8×8 單色
func (r *Rom) FontColor() *Font                     // 8×8 四平面
```

`Block.Raw` 一定要留：存檔策略是**改寫不是重建**（`CLAUDE.md` §4），
未解區域一個 byte 都不能動。

## 3. 未解與邊界

| 項目 | 狀態 | 實作時怎麼辦 |
|---|---|---|
| `ALLPICS` 交錯的參數區 | 未解（形狀像動畫影格表） | 原樣保留，先不做圖片動畫 |
| EGA 調色盤 | **未解** | 暫用 EGA mode 0Dh 標準 16 色，**在程式碼與文件標明是暫代**，等 RE 補上 |
| `TRANSTBL`（800 bytes） | 未解 | 不載入 |
| `CURS`（2,048 bytes） | 未解 | 不載入；游標自行設計 |
| `ALLHTDS` 圖磚組 163 張以外的位置 | 已知緩衝區後段會被當圖磚用 | 圖形編號超出該組張數時，回傳明確錯誤而不是靜靜畫錯 |
| 記錄本體的欄位語意 | 隨 section 型別而不同，多數未解 | `internal/assets` 只回傳 bytes，不解讀語意 |

## 4. 驗收條件

1. **雜湊驗證**：20 個檔案的 SHA-256 全部對上；改一個 byte 就要拒絕載入。
2. **全量解碼**：42 個 MSQ 區塊全部通過原版自己的 checksum；
   173 個 Huffman 子區塊長度精確吻合；9 個圖磚組長度都整除 128。
3. **文字**：解出 4,889 個字串槽、4,827 條非空，與 `tools/decode_text.py`／`decode_block_text.py` 一致。
4. **round-trip**：`Block.Raw` 原樣寫回，與原始檔 **byte-for-byte 相同**。
5. **對原版畫面**：任取一張 `ALLPICS` 圖與一組圖磚，在 DOSBox 跑原版截同一張圖，
   與解碼結果**逐像素比對**（調色盤未定案前比索引值，不比 RGB）。
6. 地圖三層的長度檢查：前兩層填滿地圖區、尾段 ＝ D²，42/42。

第 5 項是**對原版行為**的驗收，不能用「測試全綠」代替。
