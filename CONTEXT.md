# CONTEXT — 專案脈絡與文件索引

> 這份是全專案的單一入口。對話被壓縮、或換一個新 session 接手時先讀這份，
> 再依索引跳到需要的文件。工作紀律與硬規則在 [`CLAUDE.md`](./CLAUDE.md)。
>
> 最後更新：2026-08-13

---

## 1. 這個專案在做什麼

把 1988 年 Interplay／EA 的 DOS 版《Wasteland》（台灣代理譯名「荒野遊俠」）
完整逆向，在 Go / Ebiten 上重寫引擎，再做繁體中文化。定位是文化資產保存，
連同 1990 年軟體世界的中文說明書一起留下來。

三條硬性原則（細節見 `CLAUDE.md`）：

1. **RE 沒確認完不動工。** 機制要先在 IDA 裡讀出來、寫成筆記、收攏成 READY 規格，才能寫引擎。
2. **完整性 > 投報。** 不用成本效益當跳過任何素材或格式的理由。
3. **一手資料贏二手推論。** 檔案內容 > 檔名、社群文章、其他重製專案。

## 2. 現況一覽

### 已完成

| 項目 | 狀態 |
|---|---|
| 專案規範 | `CLAUDE.md`（三道閘門、IDA 工具鏈、文件政策） |
| IDA 工具鏈 | `tools/ida.sh`（build／run／raw），image `ida-pro-9.4-idapython:py312-v1`，IDAPython 實測可用 |
| 原版身分 | 20 個檔案的 SHA-256 全部記錄（`docs/re/01`） |
| 打包器識別 | `wl.exe` 是 Microsoft EXEPACK（`docs/re/02`） |
| 解包器 | `tools/unpack_exepack.py`，169,312 bytes 映像還原成功，36 筆 relocation 重建 |
| 基準資料庫 | 打包版（不可用作證據）與解包版（614 函式）各一份，清冊已匯出 |
| 開機序列 | `start` 全解：`info` 兩 bytes 的安裝資訊、七個開機載入的素材與各自的目標位址與長度（`docs/re/03`） |
| 檔名表 | 13 筆全部定位（`0x25FAD`–`0x26028`），引用點逐一對上 |
| 檔案 I/O | open／read／close 三個包裝函式與「請插入磁片」重試迴圈已解 |
| `TITLE.PIC` 解碼 | XOR 自參考串流（`out[n] = in[n] XOR out[n-0x90]`），長度算術完全吻合 |
| RE 工具 | `export_file_io.py`（中斷掃描＋字串引用）、`export_function.py`（函式完整倒出）、`apply_overlay.py` |
| 資源目錄 | `GAME1`／`GAME2` 的定址三張表全解（目錄 `0x28CE9`、位移表 A `0x28A9A`／B `0x28AEA`），資源自帶 2 bytes 長度（`docs/re/06`） |
| 文字輸出 | `sub_1786E` 印字串（`ds:4680h`）→ overlay slot 19 畫字元；`wl.exe` 內介面文字是明文 ASCII |
| `wla.bin` overlay | 26 個 slot 的 API 表、EGA mode 0Dh、列位址表、畫字元（字型 172 字 × 32 bytes、8×8、4 平面）、清除矩形（`docs/re/04`） |
| 儲存層 | 雙模式（硬碟 DOS 檔案／磁片 `int 25h` 絕對磁區）與分流旗標；資源表 8 筆全解，六個檔名的引用點就在表的 `+6` 欄位（`docs/re/05`） |
| 英文手冊 | 全文轉 markdown，7 章 646 行（`docs/manual/`） |
| 段落書 | 162 段全部轉錄，編號連續無缺（`docs/paragraphs/`）。**三層防拷結構已辨識**：3 個陷阱段落（1／22／145）、64 段變體組（同場景不同密語）、33 段火星誘餌假劇情 |

### 進行中／未開始

| 項目 | 狀態 |
|---|---|
| 解包映像實跑驗證 | **未做**。要在 DOSBox 跑起來與原版對照，才能把「解包等同原版」升為已確認 |
| 資料格式 | 全部未解：`game1`／`game2`／`allpics*`／`allhtds*`／`wla.bin`／`transtbl`／`curs`／`colorf.fnt` |
| 文字編碼 | 未解。`wl.exe` 有兩串疑似編碼表（`0x1DD98`、`0x1C01B`），是下一個入口 |
| 說明書整理 | 未開始（四份：軟體世界中文版 33 張掃描、英文 `manual.txt`、`paragraphs.txt`、社群攻略） |
| Go 引擎 | **未開始，且依規定不得開始**，要等規格 READY |

## 3. 文件索引

| 文件 | 內容 |
|---|---|
| [`CLAUDE.md`](./CLAUDE.md) | 專案規範：三道閘門、IDA 鐵則、文件與中文化政策、環境硬規則 |
| [`docs/re/01-binary-identity.md`](docs/re/01-binary-identity.md) | 20 檔 SHA-256、`wl.exe` 的 MZ header、第一份資料庫與「不可用作證據」的結論 |
| [`docs/re/02-exepack-unpack.md`](docs/re/02-exepack-unpack.md) | EXEPACK 格式、解包器、relocation 起點的坑、解包後基準資料庫 |
| [`docs/re/03-boot-and-asset-loading.md`](docs/re/03-boot-and-asset-loading.md) | 開機序列、`info` 安裝資訊、檔名表、七個開機素材的載入位址、`TITLE.PIC` XOR 解碼 |
| [`docs/re/04-overlay-wla-bin.md`](docs/re/04-overlay-wla-bin.md) | `wla.bin` overlay 機制、26 個 slot 的 API 表、繪圖層三支 |
| [`docs/re/05-storage-layer.md`](docs/re/05-storage-layer.md) | 雙模式儲存、資源表結構、六個資料檔的開啟路徑 |
| [`docs/re/06-resource-directory.md`](docs/re/06-resource-directory.md) | `GAME1`／`GAME2` 的資源目錄與位移表、資源 header、文字輸出層 |
| [`docs/manual/`](docs/manual/) | 官方英文手冊全文 markdown |
| [`docs/paragraphs/`](docs/paragraphs/) | 段落書 162 段全文與索引，含防拷結構標註 |
| `docs/re/generated/ida94/` | 工具匯出的清冊（JSON ＋ markdown），不含人的推論 |

## 4. oracle 優先序

原版執行檔行為（IDA 讀出） > DOSBox 實跑觀察 > 官方英文手冊 >
軟體世界中文說明書 > 社群攻略。低優先來源只能在高優先缺席時暫代，且要標明。

## 5. 術語表

| 詞 | 定義 |
|---|---|
| 打包版 | 原始 `wl.exe`，EXEPACK 壓縮狀態。位址不可引用 |
| 解包版 | `wl.unpacked.exe`，`tools/unpack_exepack.py` 的輸出，所有逆向的基準 |
| 線性位址 | IDA database linear address（如 `0x110B6`）。與 `segment:offset` 併記 |
| 段落書 | `paragraphs.txt`，原版的紙本防拷文本，重製版改做遊戲內手札 |
| 推論等級 | 已確認／強證據／假說／未解 |

## 6. 已被推翻的斷言

| 斷言 | 為什麼錯 | 正確答案 |
|---|---|---|
| `wl.exe` 裡 `Packed file is corrupt#` 的 `#` 是錯誤訊息的一部分 | IDA 把後續資料一起顯示成字串了 | `#`＝`0x23`＝35，是第一組 relocation 的 count（`docs/re/02` §3） |
| 檔名表沒有任何程式碼引用（xref 與立即數掃描都是 0） | 立即數掃描器沒遮罩符號擴展，`0x91C5` 被讀成 `0xFFFF...91C5`，算出的位址不存在 | 檔名全部由 `mov dx, <位移>` 直接引用，修正後 `seg002` 有 50 個字串對上引用點（`docs/re/03` §7） |

## 7. Worklist

**下一步（按順序）**

1. 建資源編號 → 資產類型的對照：資源目錄與兩張位移表已解（`docs/re/06`），
   下一步是找出每個編號對應什麼（地圖／文字／圖），這是通往劇情文字的主線。
2. 用位移表把 `GAME1`／`GAME2` 實際切開，逐個資源看內容與 header。
3. 解 `ds:722Fh` 的字元碼重映射（`docs/re/04` §5），中文化前必須先解。
4. 逐一解 overlay 其餘 21 個 slot，特別是 `0x1029B`（881 bytes）。
5. 追資源表 idx 7（無檔名，疑似存檔區）在硬碟模式下怎麼存取。
6. 在 DOSBox 跑解包版與原版對照，驗證解包正確性（`docs/re/02` §5 的待辦）。
7. 產函式索引 `docs/re/00-function-index.md`（讀任何 `sub_xxx` 前要先查）。
8. 段落書的防拷結構要接進 remake 設計：變體組與火星誘餌劇情不能照抄成線性手札，
   要等段落呼叫表解出來才知道遊戲實際會叫哪一段。

**不得開始**：`internal/` 下的任何 Go 引擎程式碼，直到對應規格標 READY。
