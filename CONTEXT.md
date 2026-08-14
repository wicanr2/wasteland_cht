# CONTEXT — 專案脈絡與文件索引

> 這份是全專案的單一入口。對話被壓縮、或換一個新 session 接手時先讀這份，
> 再依索引跳到需要的文件。工作紀律與硬規則在 [`CLAUDE.md`](./CLAUDE.md)。
> **逆向結果的速查表在 [`docs/re/00-master-index.md`](docs/re/00-master-index.md)。**
>
> 最後更新：2026-08-14

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
| 基準資料庫 | 打包版（不可用作證據）、解包版 `wl.unpacked.exe`（614 函式）、**合成映像 `wl.merged.exe`（解包 ＋ overlay，641 函式，所有筆記的引用基準）** |
| 開機序列 | `start` 全解：`info` 兩 bytes 的安裝資訊、七個開機載入的素材與各自的目標位址與長度（`docs/re/03`） |
| 檔名表 | 13 筆全部定位（`0x25FAD`–`0x26028`），引用點逐一對上 |
| 檔案 I/O | open／read／close 三個包裝函式與「請插入磁片」重試迴圈已解 |
| `TITLE.PIC` 解碼 | XOR 自參考串流（`out[n] = in[n] XOR out[n-0x90]`），回看距離 `0x90` 就是列寬 → **288 × 128 packed 4bpp**（`docs/re/23`） |
| RE 工具 | `export_file_io.py`（中斷掃描＋字串引用）、`export_function.py`（函式完整倒出）、`apply_overlay.py` |
| 資源目錄 | `GAME1`／`GAME2` 的定址三張表全解（目錄 `0x28CE9`、位移表 A `0x28A9A`／B `0x28AEA`）（`docs/re/06`） |
| MSQ 區塊 | 兩個資料檔切開：`game1` 20 塊全部 `msq0`、`game2` 22 塊全部 `msq1`；數量與目錄索引三重吻合（`docs/re/07`） |
| **MSQ 加密** | **已解**：`key = lo(checksum) XOR hi(checksum)`，逐 byte XOR 後 `key += 0x1F`。42/42 通過原版自己的 checksum 驗證，解密後出現 `Titanium`／`Vanadium`／`REDHAWK` 等遊戲文字（`docs/re/08`） |
| 中文說明書 | 全 60 頁節轉錄完成，含 IBM 版補充說明書與約 200 條當年譯名（`docs/manual-cht/`） |
| **字型與文字編碼** | **已解**：兩套字型、兩套索引。主文字字型**內嵌在 `wl.exe`**（`seg003:0xCA60`，128 字 × 8 bytes、單色、索引 ＝ ASCII − 0x20）；`colorf.fnt` 是彩色選單字型（兩組同形不同色的字模，`ds:722Fh` 選色，不是字元碼重映射）。18 個文字控制碼解出 14 個（`docs/re/14`） |
| 文字輸出 | `sub_1786E` 印字串（`ds:4680h`）→ 每字元處理器可切換（`ds:B265h`）。**遊戲文字是 5-bit 打包的**（`docs/re/17`）；只有少數介面字串（`Yes`、`CREATE DELETE PLAY`、`Money = $`）以明文 ASCII 存在執行檔裡 |
| `wla.bin` overlay | 26 個 slot 的 API 表、EGA mode 0Dh、列位址表、畫字元（字型 172 字 × 32 bytes、8×8、4 平面）、清除矩形（`docs/re/04`） |
| **亂數產生器** | **已解**：`sub_18E6B` 是 `ds:465Ch`–`4660h` 五個位元組的進位鏈，映像初值全零、全檔沒有種子設定，熵來自鍵盤輪詢次數。擲骰層四支（d6／dN／累加 Nd6／2d6 同點續擲）全部讀完並以模型驗證（`docs/re/13`、`tools/rng.py`） |
| **文字編碼** | **已解**：遊戲文字是 **5-bit 打包 ＋ 60 字元對照表**。執行檔九張表 426 條、42 個地圖區塊各一張表共 4,401 條，**合計 4,827 條非空全部解出**（`docs/re/17`、`18`）|
| **MSQ 區塊佈局** | **已解**：地圖區長度由選擇表決定（`0x600` ＝ 邊長 32／`0x1800` ＝ 邊長 64，只有 4 塊是大地圖），之後是 0x5C bytes 的記錄區標頭，41/42 個區塊第一個 section 落在 `P+0x5C`。取記錄走兩層索引（`sub_17CB1`），記錄指標落在 `ds:46AEh`（`docs/re/16`） |
| **地圖三層與圖磚** | **已解**：地圖是正方形（邊長在記錄區標頭 `+0x2C`，32 或 64），分三層——第 1 層 4 bits／格是 section 型別、第 2 層 1 byte／格是記錄編號、**第 3 層（Huffman 尾段）是畫面上的圖形編號**（0–9 是 `IC0_9.WLF` 的疊圖，≥10 是圖磚編號 −10；`0x420 ＋ 10 × 128 ＝ 0x920` 剛好接上圖磚組）。**圖磚在 `ALLHTDS`**：9 組、每組 66–163 張，一張 128 bytes ＝ **16 × 16 packed 4bpp**，由標頭 `+0x30` 選組。畫一格走 `螢幕 ← (背景 AND 遮罩) OR 疊圖`（overlay slot 4），42 張地圖的縮圖都畫得出來（`docs/re/24`）|
| **遊戲時鐘** | **已解**：24 小時制（`ds:4658h` 分的小數／`4659h` 分／`465Ah` 時），走一步推進的量寫在該地圖的記錄區標頭 `+0x34/+0x35`——荒野 4 分鐘、一般室內 15 秒。時鐘畫在外框上緣（字元欄 28、列 0）；**晝夜門檻 6 時與 18 時**，夜間換圖形、記錄也有白天／夜間兩套欄位。每 16 刻跑一次隨時間的角色處理（`docs/re/27`）|
| **移動與事件觸發** | **已解（骨架）**：走一步 ＝ 可否進入（四道閘）→ 捲動（四個方向各一支 ＋ overlay slot 11–14）→ 腳步音效 → 遭遇擲骰 → 重畫與觸發。踩上去由 `ds:AA87h` 這張 **16 筆跳表**依地圖第 1 層的 nibble 分派；7 種是空的，8 種有專屬處理（10 ＝ 傳送、12 ＝ 印訊息）。**時間系統沒找到**（`docs/re/26`）|
| **畫面版面** | **已解**：320 × 200 mode 0Dh。地圖／圖片視窗 **288 × 128 @ (8, 8)**（19 × 9 格、四邊半格裁切、隊伍固定在第 (9,4) 格），外框在欄 0–37／字元列 0–17，訊息視窗欄 1–38／字元列 18–23（6 行）。`ds:46B9h` 切換地圖與隊伍名單，兩者共用同一塊視窗（`docs/re/25`）|
| **圖片格式** | **已解**：全部是 **packed 4bpp ＋ 列間 XOR delta**，而且 **XOR 的回看距離就是一列的 byte 數**——`ALLPICS` 是 48 → 96 × 84（共 82 張），`TITLE.PIC` 是 144 → 288 × 128。`ALLPICS` 的解碼在 overlay slot 2（`sub_10144`）、`TITLE.PIC` 在 `start` 內嵌，已用 `tools/decode_pic.py` 重現（`docs/re/23`）|
| **商店與物品** | **已解**：商店由地圖記錄設定，**價格 ＝ 基礎價 − (基礎價 >> n)**。物品資料表在 `ds:7A31h`，**95 筆 × 8 bytes**——與字串表裡的 95 個物品名兩個獨立來源吻合（`docs/re/22`）|
| **七個屬性** | **已確認**：Strength／IQ／Luck／Speed／Agility／Dexterity／Charisma 在角色記錄 `+0x0E`–`+0x14`（選單字串 ＋ `sub al,'1'; add al,0Eh` 兩行釘死）。屬性→修正值有死區 9–13、兩側各半格。**角色建立也已解**：屬性 ＝ 5d6 取最高三顆、MAXCON ＝ 同擲法 ＋18、技能點 ＝ IQ（`docs/re/21`）|
| **戰鬥判定** | **兩條路徑都已定位**：隊伍打敵方（`0x1AF52`）與敵方打隊伍（`0x1B04C`）共用同一支累加器，判定前綴機器碼相同、只差 `jb`／`jnb`——累加的是隊伍成員的本事，方向隨攻守翻轉。傷害 ＝ 基底 ＋ Nd6，兩邊護甲都是 N 顆 d6 的吸收。**敵方 HP 在 `ds:46C8h + 編號×2`**，減到 ≤0 夾成 0；角色 CON 可為負並分五級傷勢（`docs/re/20`）|
| **效果與傷害** | **已解**：地圖事件對角色的所有效果由記錄 `+0x08`／`+0x09` 兩個 byte 描述（哪個欄位、加或減、固定值或 Nd6），共用 `sub_141FA` 一個出口。護甲吸收 ＝ **AC 顆 d6 的和**（`docs/re/19`）|
| **角色記錄** | **定址已確認**：記錄 ＝ `0x7131 ＋ 角色編號 × 256`，每筆 256 bytes，經隊伍槽表兩層間接。名字、金錢（24-bit）、七個屬性、MAXCON／CON、**技能與物品兩個 30 槽陣列（已分辨）**、傷勢門檻（−11／−20／−30／−40）都已定位（`docs/re/15`） |
| 逐指令基準 | 整個 CODE 區倒成 JSON（20,177 條指令、827 個全域、4,932 筆直接定址存取），後續形狀比對改在離線做（`tools/ida/export_listing.py`、`export_memops.py`） |
| 儲存層 | 雙模式（硬碟 DOS 檔案／磁片 `int 25h` 絕對磁區）與分流旗標；資源表 8 筆全解，六個檔名的引用點就在表的 `+6` 欄位（`docs/re/05`） |
| 英文手冊 | 全文轉 markdown，7 章 646 行（`docs/manual/`） |
| 段落書 | 162 段全部轉錄，編號連續無缺（`docs/paragraphs/`）。**三層防拷結構已辨識**：3 個陷阱段落（1／22／145）、64 段變體組（同場景不同密語）、33 段火星誘餌假劇情 |

### 進行中／未開始

| 項目 | 狀態 |
|---|---|
| 解包映像實跑驗證 | **未做**。要在 DOSBox 跑起來與原版對照，才能把「解包等同原版」升為已確認 |
| 資料格式 | 已解：`wla.bin`（overlay 程式碼）、`title.pic`（XOR 串流）、`colorf.fnt`（172 字 × 32 bytes，格式與用途都已驗證）、主文字字型（內嵌）、`GAME1`／`GAME2` 的定址方式。`allpics1/2` 的 82 張圖（`docs/re/23`）、`allhtds1/2` 的 9 組圖磚（`docs/re/24`）。`ic0_9.wlf`／`masks.wlf` 的格式與用途（`docs/re/24` §2.3）。未解：`allpics*` 交錯的參數區、`transtbl`、`curs`、`end.cpa` |
| 劇情敘述文字 | **已解**：執行檔九張打包表 426 條 ＋ 地圖區塊 4,401 條，合計 4,827 條非空（`docs/re/17`、`18`）。與紙本段落書 162 段的分工待段落呼叫表解出 |
| MSQ 尾段 | 已解：無 magic 的 Huffman 流，42/42 解出 4,096 或 1,024 bytes ＝ 地圖第 3 層（每格 1 byte，`docs/re/24`） |
| **Huffman 解壓** | **已實作並驗證**（`tools/huffman.py`）：`allhtds1/2`、`allpics1/2`、`end.cpa` 共 173 個子區塊全部解出，長度精確吻合、檔案 100% 用完（`docs/re/11`） |
| 載入器分工 | 已解：`DL`＝0 ALLPICS、1 GAME／存檔、2 ALLHTDS、6 END.CPA，各有位移表 |
| 說明書整理 | 英文手冊、段落書、軟體世界中文說明書都完成；社群攻略未開始 |
| **規格（G2）** | **三份 READY**：`docs/spec/01`（資產與資料格式）、`02`（亂數與擲骰）、`03`（畫面與文字）。其餘五份要等對應逆向補完 |
| **`internal/assets`** | **已實作並通過驗收**：SHA-256 驗證、資源定址、MSQ 解密、Huffman、5-bit 文字、兩套字型、圖片／圖磚／地圖三層。9 個測試全綠，含 `Raw` 的 byte-for-byte round-trip（`tools/go.sh test ./...`）|
| **`internal/textlayout`** | **已實作**：18 個控制碼（未解的回報成事件、絕不當文字印）、組行與分頁。4,889 條語料全部排得過 |
| **`internal/render`** | **已實作**：合成 320 × 200 索引畫面。地圖視窗逐像素驗過剛好 288 × 128 @ (8,8)、捲動一格 ＝ 位移 16 像素 |
| **`internal/game/rng`** | **已實作**：進位鏈與四支擲骰 ＋ 5d6 取三。驗收數列（前七項 ＝ 二項式係數）、分佈、300 萬次不重複全過 |
| Go 引擎（其餘） | `internal/ui`（Ebiten 那層）未做；規則層仍不得開始 |

## 3. 文件索引

| 文件 | 內容 |
|---|---|
| [`CLAUDE.md`](./CLAUDE.md) | 專案規範：三道閘門、IDA 鐵則、文件與中文化政策、環境硬規則 |
| [`docs/re/00-master-index.md`](docs/re/00-master-index.md) | **RE 總表**：位址換算、資料格式、結構佈局、位址表、關鍵函式、工具。**查已知事實先看這份** |
| [`docs/re/00-remake-knowledge-gaps.md`](docs/re/00-remake-knowledge-gaps.md) | **RE 完成度檢查表**：remake 需要的每一項知識、狀態與入口 |
| [`docs/re/00-function-index.md`](docs/re/00-function-index.md) | 函式索引（641 個，已分析 256）。讀任何 `sub_XXXXX` 前先查 |
| [`docs/re/01-binary-identity.md`](docs/re/01-binary-identity.md) | 20 檔 SHA-256、`wl.exe` 的 MZ header、第一份資料庫與「不可用作證據」的結論 |
| [`docs/re/02-exepack-unpack.md`](docs/re/02-exepack-unpack.md) | EXEPACK 格式、解包器、relocation 起點的坑、解包後基準資料庫 |
| [`docs/re/03-boot-and-asset-loading.md`](docs/re/03-boot-and-asset-loading.md) | 開機序列、`info` 安裝資訊、檔名表、七個開機素材的載入位址、`TITLE.PIC` XOR 解碼 |
| [`docs/re/04-overlay-wla-bin.md`](docs/re/04-overlay-wla-bin.md) | `wla.bin` overlay 機制、26 個 slot 的 API 表、繪圖層三支 |
| [`docs/re/05-storage-layer.md`](docs/re/05-storage-layer.md) | 雙模式儲存、資源表結構、六個資料檔的開啟路徑 |
| [`docs/re/06-resource-directory.md`](docs/re/06-resource-directory.md) | `GAME1`／`GAME2` 的資源目錄與位移表、資源 magic、文字輸出層 |
| [`docs/re/07-msq-blocks.md`](docs/re/07-msq-blocks.md) | 42 個 MSQ 區塊的完整清單、數量的三重驗證、內容是加密／壓縮的 |
| [`docs/re/08-msq-encryption.md`](docs/re/08-msq-encryption.md) | MSQ 加密演算法與 42/42 驗證、區塊佈局、長度表 |
| [`docs/re/09-msq-map-structure.md`](docs/re/09-msq-map-structure.md) | 地圖層（三層、邊長、自相關驗證）、名稱字串是明文、存檔兩份輪替 |
| [`docs/re/10-huffman-compression.md`](docs/re/10-huffman-compression.md) | 第二層 Huffman：容器格式、位元讀取器、樹的前序編碼、各資源的載入器對照 |
| [`docs/re/11-huffman-decoder.md`](docs/re/11-huffman-decoder.md) | 解碼器實作與驗證、子區塊串接規則、四個資料檔全部解開 |
| [`docs/re/12-msq-tail-and-text-model.md`](docs/re/12-msq-tail-and-text-model.md) | 兩張長度表的差別＝尾段、尾段是無 magic 的 Huffman、文字在各來源的分佈 |
| [`docs/re/13-rng.md`](docs/re/13-rng.md) | 亂數產生器與擲骰層：進位鏈演算法、狀態初值、拒絕取樣、四支包裝函式、參考模型與驗證 |
| [`docs/re/14-fonts-and-text-encoding.md`](docs/re/14-fonts-and-text-encoding.md) | 兩套字型與兩套索引、`ds:722Fh` 是選色不是重映射、18 個文字控制碼、選單系統 |
| [`docs/re/15-character-record.md`](docs/re/15-character-record.md) | 角色記錄的兩層定址、欄位表、傷勢等級與門檻、名片行的欄位座標 |
| [`docs/re/16-msq-block-layout.md`](docs/re/16-msq-block-layout.md) | MSQ 區塊的整體佈局、地圖區長度與尺寸、記錄區標頭、section 位移表與兩層索引 |
| [`docs/re/17-packed-text.md`](docs/re/17-packed-text.md) | 5-bit 打包文字的格式與解碼器、執行檔九張字串表、單複數與性別機制、名片欄位對照 |
| [`docs/re/18-block-text.md`](docs/re/18-block-text.md) | 地圖區塊的字串表：加密長度就是字串表基址、完整區塊佈局、4,401 條敘述文字 |
| [`docs/re/19-effects-and-damage.md`](docs/re/19-effects-and-damage.md) | 資料驅動的效果系統（記錄 `+0x08`／`+0x09`）、傷害與護甲、單複數選擇器 |
| [`docs/re/20-combat-resolution.md`](docs/re/20-combat-resolution.md) | 命中判定（d100 對門檻）、武器傷害公式、一次攻擊的完整流程 |
| [`docs/re/21-attributes.md`](docs/re/21-attributes.md) | 七個屬性的記錄位移、屬性→修正值階梯、檢定骰、角色建立 |
| [`docs/re/22-shop-and-items.md`](docs/re/22-shop-and-items.md) | 商店、價格公式、物品資料表（95 筆 × 8 bytes） |
| [`docs/re/23-picture-format.md`](docs/re/23-picture-format.md) | 圖片格式：packed 4bpp ＋ 列間 XOR delta、`ALLPICS` 容器與 82 張圖 |
| [`docs/re/24-map-layers-and-tiles.md`](docs/re/24-map-layers-and-tiles.md) | 地圖的三層結構、邊長與定址、`ALLHTDS` 九組 16 × 16 圖磚 |
| [`docs/re/25-screen-layout.md`](docs/re/25-screen-layout.md) | 畫面版面：兩套座標單位、地圖／圖片視窗、外框、訊息視窗、隊伍名單 |
| [`docs/re/26-movement-and-triggers.md`](docs/re/26-movement-and-triggers.md) | 走一步的流程、四方向捲動與補畫、nibble → 事件處理的 16 筆跳表 |
| [`docs/re/27-game-clock.md`](docs/re/27-game-clock.md) | 遊戲時鐘：24 小時制、每步推進量、晝夜門檻、隨時間的角色處理 |
| [`docs/spec/00-index.md`](docs/spec/00-index.md) | **規格索引與閘門狀態**：哪些可以動工、其餘擋在什麼上 |
| [`docs/spec/01-assets-and-formats.md`](docs/spec/01-assets-and-formats.md) | READY：資源定址、解密、Huffman、5-bit 文字、字型、圖片、圖磚、地圖三層 ＋ Go 介面草案 |
| [`docs/spec/02-rng-and-dice.md`](docs/spec/02-rng-and-dice.md) | READY：進位鏈亂數與四支擲骰，含驗收數列 |
| [`docs/spec/03-screen-and-text.md`](docs/spec/03-screen-and-text.md) | READY：畫布、五個視窗、座標單位、控制碼、中文版面的兩條路 |
| `internal/` | 已實作：`assets`（規格 01）、`textlayout`／`render`（規格 03）、`game/rng`（規格 02）。`tools/go.sh` 是 Go 的唯一入口，編譯與測試走 docker |
| [`docs/manual-cht/`](docs/manual-cht/) | 軟體世界 1990 中文說明書全 60 頁節轉錄 ＋ 當年譯名表 |
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
| `GAME1` 等六個檔名也沒有引用（修正掃描器後仍是 0） | 只查了「位址以立即數出現」這一種形式 | 檔名位址是資源表 `+6` 欄位的資料值，開啟路徑是 `sub_11445`（`docs/re/05` §6） |
| `bcdefghijklmdenopq` 是文字解碼的頻率表 | 憑字串長相猜的，沒有追引用點 | 它是 `sub_166D3` 逐字印到畫面上的內容，不是解碼表（`docs/re/06` §3） |
| MSQ 區塊的地圖是固定 2048 bytes，記錄表從 `+0x800` 起 | `2048` 是從「64×64 × 4 bits」推出來的，不是讀出來的；自相關驗證通過的是「這區是地圖」，不是「這區到哪為止」 | 地圖區長度由執行檔的選擇表決定：`0x600` 或 `0x1800`，記錄區接在後面（`docs/re/16` §1） |
| 地圖是 64 × 48（`0x600`）或 128 × 96（`0x1800`）的單一平面 | 自相關的峰值出現在 64 nibble，於是把整個地圖區當成一張圖。實際上峰值來自**占三分之二的 byte 層**（邊長 32 的 byte 列在 nibble 視角下正好是 64） | 地圖是**正方形**，邊長寫在記錄區標頭 `+0x2C`（32 或 64），而且分**三層**（`docs/re/24` §1–§2） |
| MSQ 的 Huffman 尾段是該地圖的 tile 圖形 | 只看內容形狀（大量重複的小值）與 `ALLPICS` 相似，沒追載入後誰讀它 | 尾段是**地圖的第 3 層**，每格 1 byte、長度 D²；圖磚圖形在 `ALLHTDS`（`docs/re/24`） |
| `ds:722Fh` 的 `+0x1C` 是字元碼重映射 | 只看到算術沒把字模畫出來 | 是**選色**：`colorf.fnt` 有兩組形狀相同、顏色不同的字模，相距 `0x1C`（`docs/re/14` §3） |
| 用 xref 圖就能問出「誰在寫這個全域」 | 這份資料庫 4,932 筆直接定址存取裡只有 22 筆進了 xref 圖，其餘 `mov ax, ds:46B7h` 這類根本沒建 xref。零命中與「真的沒人寫」長得一樣 | 自己解碼運算元並用 IDA 的 `CF_CHGn`／`CF_USEn` 判讀寫（`tools/ida/export_memops.py`、`docs/re/13` §1） |
| RNG 應該是移位／加法型 | 排除 `mul` 之後只憑印象推的 | 是**純加法**的進位鏈，一個移位都沒有；靠「讀寫同一全域」的形狀找到，不是靠指令（`docs/re/13` §2） |
| 解密要對整個 MSQ 區塊做 | 原版只解到「標頭第一個 word」那個長度為止；多解的那一段被 XOR 破壞成高熵資料，看起來像壓縮或未知格式 | 加密長度 ＝ 標頭第一個 word，而**同一個 word 就是字串表的基址**——字串表接在加密區之後且不加密（`docs/re/18` §2）|
| 原版遊戲檔案裡沒有長篇敘述文字 | 依據是「掃不到 24 字元以上的 ASCII」，而遊戲文字是 5-bit 打包的，本來就掃不到 | 執行檔九張表 442 條（含完整的結局敘述）＋ 42 個地圖區塊 4,493 條，合計 **4,935 條**（`docs/re/17`、`18`） |
| ` etraoishlnd` 是文字解碼的頻率表 | 憑字串長相猜的；第一次推翻時只否定了「是頻率表」，沒找出它到底是什麼，於是同一個字串被猜了兩次 | 它是結局敘述字串表的**字元對照表**，`0x1B7BF` 就是載入它的地方（`docs/re/17` §3） |
| 遊戲文字共 4,935 條（執行檔 442 ＋ 區塊 4,493） | 兩個數字都是工具的產物而不是資料本身：`decode_block_text.py` 把「整個區塊」餵給解碼器，但原版只把 `ds:BD22h` 那麼長讀進緩衝區，多出來的壓縮尾段被 5-bit 解成看起來像文字的雜訊；另一支則用「看起來像不像文字」的啟發式裁掉尾巴，連真字串一起丟 | **4,889 個字串槽、4,827 條非空**（執行檔 444／426、區塊 4,445／4,401）。Go 與 Python 兩套獨立實作逐塊相同（`internal/assets`、`docs/re/18`） |
| MSQ 資源的前 2 bytes 是長度 | 把 `mov cx, [bx]` 當成「取剛讀進來的 header」，其實 `ds:46B0h` 是別處設好的緩衝區位址 | 那 4 bytes 是 magic `msq0`／`msq1`；長度來源仍未解（`docs/re/07` §7） |

## 7. Worklist

**RE 完成度**：資料格式、文字與資產層打通（含地圖三層、圖磚與圖片），
戰鬥／屬性／效果／商店四塊規則已解；世界互動層仍是主要缺口。
641 個函式裡人寫的筆記涵蓋 256 個。完整缺口見
[`docs/re/00-remake-knowledge-gaps.md`](docs/re/00-remake-knowledge-gaps.md)。

**下一步（依「對 remake 的阻擋程度」排序）**

1. **存檔內部欄位（C2）** —— 角色記錄已解（`docs/re/15`），存檔是它的外層容器；
   時鐘與 32-bit 累計那 14 bytes 已知會存進去（`docs/re/27` §7），是好的切入點。
2. **nibble 5／8／9 的事件處理函式** —— 分派已解，但那三支 IDA 沒建成程式碼
   （`0x15280`／`0x15160`／`0x14410`），要指定位址重新分析（`docs/re/26` §8）。
   `sub_12537`／`sub_1254A`（隨時間恢復／惡化的公式）也在這一批。
3. **段落編號顯示（E3）** —— 段落書與遊戲內文字的分工要重新釐清。
4. **中文排版決策** —— 訊息視窗只有 6 行 × 38 個 8×8 字元，換成 16 × 15 中文
   放不下，要連同文字控制碼一起重排（`docs/re/25` §3）。
5. 逐一解 overlay 其餘的未解 slot（26 個裡已解 12 個）。
6. 追資源表 idx 7（無檔名，疑似存檔區）在硬碟模式下怎麼存取。
7. 音效（F2）：確定是 PC 喇叭——`sub_1CC76` 寫 8253 channel 2（`out 42h`）
   再開 61h 閘。待解的是 `sub_1CD52` 那套位元組碼的指令集與曲目資料在哪。
8. 在 DOSBox 跑解包版與原版對照，驗證解包正確性（`docs/re/02` §5 的待辦）。
9. 段落書的防拷結構要接進 remake 設計：變體組與火星誘餌劇情不能照抄成線性手札，
   要等段落呼叫表解出來才知道遊戲實際會叫哪一段。

**可以開始**：`internal/assets`、`internal/ui`／`internal/textlayout`、`internal/game/rng`
——這三塊的規格已標 READY（`docs/spec/00-index.md`）。
**仍不得開始**：規則層（角色、戰鬥、世界互動、存檔）的任何 Go 程式碼。
