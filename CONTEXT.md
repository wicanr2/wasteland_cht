# CONTEXT — 專案脈絡與文件索引

> 這份是全專案的單一入口。對話被壓縮、或換一個新 session 接手時先讀這份，
> 再依索引跳到需要的文件。工作紀律與硬規則在 [`CLAUDE.md`](./CLAUDE.md)。
> **逆向結果的速查表在 [`docs/re/00-master-index.md`](docs/re/00-master-index.md)。**
>
> 最後更新：2026-08-15

---

## 0. 目標（使用者定案 2026-08-15）

**完成 wasteland remake。** 逆向工程告一段落之後進入實作循環，
**每一輪都照這個順序**：

```
讀 docs/re/ 的逆向文件 → 寫 docs/spec/NN（標 READY）→ 實作 ＋ 驗收 → 登記接線
```

這與 `CLAUDE.md` §0 的四道閘門是同一件事，只是把它變成每一輪的固定節奏。
**規格沒標 READY 就不寫那一塊引擎程式碼。**

最後那一步（G4）是後來補上的：前三步各自都會綠，而結論仍然可以躺在筆記裡沒人用。
命中公式的 Agility 與對手行動值、敵人的目標選擇都解過了，remake 卻一直傳著寫死的常數
（`docs/re/88`、`89`）。現在每一份筆記都要在
[`docs/re/00-wiring-status.md`](docs/re/00-wiring-status.md) 登記，
`TestWiringStatus`／`TestPlaceholders` 雙向守著。

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
| 專案規範 | `CLAUDE.md`（四道閘門、IDA 工具鏈、文件政策） |
| remake 工項 | [`WORKLIST.md`](WORKLIST.md)（含已完成的） |
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
| **文字變形機制** | **已解**：`0x0A`／`0x0C`／`0x0E` 是同一個骨架的三個實例（分段 ＋ 選擇子），分別對應**單複數／性別／him-her-it 三選一**；`0x0F` 印出數量、與 `0x0A` 共用選擇子。三個碼可以互相巢狀，外層沒選中時內層連分隔碼都看不到。中文化要照著保留分段（`docs/re/28`）|
| **字型與文字編碼** | **已解**：兩套字型、兩套索引。主文字字型**內嵌在 `wl.exe`**（`seg003:0xCA60`，128 字 × 8 bytes、單色、索引 ＝ ASCII − 0x20）；`colorf.fnt` 是彩色選單字型（兩組同形不同色的字模，`ds:722Fh` 選色，不是字元碼重映射）。18 個文字控制碼解出 14 個（`docs/re/14`） |
| 文字輸出 | `sub_1786E` 印字串（`ds:4680h`）→ 每字元處理器可切換（`ds:B265h`）。**遊戲文字是 5-bit 打包的**（`docs/re/17`）；只有少數介面字串（`Yes`、`CREATE DELETE PLAY`、`Money = $`）以明文 ASCII 存在執行檔裡 |
| `wla.bin` overlay | 26 個 slot 的 API 表、EGA mode 0Dh、列位址表、畫字元（字型 172 字 × 32 bytes、8×8、4 平面）、清除矩形（`docs/re/04`） |
| **亂數產生器** | **已解**：`sub_18E6B` 是 `ds:465Ch`–`4660h` 五個位元組的進位鏈，映像初值全零、全檔沒有種子設定，熵來自鍵盤輪詢次數。擲骰層四支（d6／dN／累加 Nd6／2d6 同點續擲）全部讀完並以模型驗證（`docs/re/13`、`tools/rng.py`） |
| **文字編碼** | **已解**：遊戲文字是 **5-bit 打包 ＋ 60 字元對照表**。執行檔九張表 426 條、42 個地圖區塊各一張表共 4,401 條，**合計 4,827 條非空全部解出**（`docs/re/17`、`18`）|
| **MSQ 區塊佈局** | **已解**：地圖區長度由選擇表決定（`0x600` ＝ 邊長 32／`0x1800` ＝ 邊長 64，只有 4 塊是大地圖），之後是 0x5C bytes 的記錄區標頭，41/42 個區塊第一個 section 落在 `P+0x5C`。取記錄走兩層索引（`sub_17CB1`），記錄指標落在 `ds:46AEh`（`docs/re/16`） |
| **地圖三層與圖磚** | **已解**：地圖是正方形（邊長在記錄區標頭 `+0x2C`，32 或 64），分三層——第 1 層 4 bits／格是 section 型別、第 2 層 1 byte／格是記錄編號、**第 3 層（Huffman 尾段）是畫面上的圖形編號**（0–9 是 `IC0_9.WLF` 的疊圖，≥10 是圖磚編號 −10；`0x420 ＋ 10 × 128 ＝ 0x920` 剛好接上圖磚組）。**圖磚在 `ALLHTDS`**：9 組、每組 66–163 張，一張 128 bytes ＝ **16 × 16 packed 4bpp**，由標頭 `+0x30` 選組。畫一格走 `螢幕 ← (背景 AND 遮罩) OR 疊圖`（overlay slot 4），42 張地圖的縮圖都畫得出來（`docs/re/24`）|
| **遊戲時鐘** | **已解**：24 小時制（`ds:4658h` 分的小數／`4659h` 分／`465Ah` 時），走一步推進的量寫在該地圖的記錄區標頭 `+0x34/+0x35`——荒野 4 分鐘、一般室內 15 秒。時鐘畫在外框上緣（字元欄 28、列 0）；**晝夜門檻 6 時與 18 時**，夜間換圖形、記錄也有白天／夜間兩套欄位。每 16 刻跑一次隨時間的角色處理（`docs/re/27`）|
| **地圖事件處理** | **八支全解**：1 ＝ 遠看才顯示的描述（站上去反而不印）、2 ＝ 條件串列（門／鎖／檢定）、4 ＝ 印一句、5 ＝ **寶箱**（第一次踩到才把類別擲成具體物品與數量，寫回記錄）、6 ＝ **設施與腳本**（第二層分派：記錄 `+0x00` bit7 設 → `ds:A4E0h` 5 個設施畫面，商店已確認；bit7 沒設 → `ds:A4EAh` **44 個腳本指令，全部讀完**：條件與晝夜分支、調整遭遇率與種類數、生成地圖物件、批次改記錄、時間戳與倒數，`docs/re/34`）、8 ＝ 多選一選單、9／12 ＝ 印訊息、10 ＝ 傳送（`docs/re/29`）|
| **移動與事件觸發** | **已解（骨架）**：走一步 ＝ 可否進入（四道閘）→ 捲動（四個方向各一支 ＋ overlay slot 11–14）→ 腳步音效 → 遭遇擲骰 → 重畫與觸發。踩上去由 `ds:AA87h` 這張 **16 筆跳表**依地圖第 1 層的 nibble 分派；7 種是空的，8 種有專屬處理（10 ＝ 傳送、12 ＝ 印訊息）。**時間系統沒找到**（`docs/re/26`）|
| **畫面版面** | **已解**：320 × 200 mode 0Dh。地圖／圖片視窗 **288 × 128 @ (8, 8)**（19 × 9 格、四邊半格裁切、隊伍固定在第 (9,4) 格），外框在欄 0–37／字元列 0–17，訊息視窗欄 1–38／字元列 18–23（6 行）。`ds:46B9h` 切換地圖與隊伍名單，兩者共用同一塊視窗（`docs/re/25`）|
| **圖片格式** | **已解**：全部是 **packed 4bpp ＋ 列間 XOR delta**，而且 **XOR 的回看距離就是一列的 byte 數**——`ALLPICS` 是 48 → 96 × 84（共 82 張），`TITLE.PIC` 是 144 → 288 × 128。`ALLPICS` 的解碼在 overlay slot 2（`sub_10144`）、`TITLE.PIC` 在 `start` 內嵌，已用 `tools/decode_pic.py` 重現（`docs/re/23`）|
| **商店與物品** | **已解**：商店由地圖記錄設定，**價格 ＝ 基礎價 − (基礎價 >> n)**。物品資料表在 `ds:7A31h`，**95 筆 × 8 bytes**——與字串表裡的 95 個物品名兩個獨立來源吻合（`docs/re/22`）|
| **七個屬性** | **已確認**：Strength／IQ／Luck／Speed／Agility／Dexterity／Charisma 在角色記錄 `+0x0E`–`+0x14`（選單字串 ＋ `sub al,'1'; add al,0Eh` 兩行釘死）。屬性→修正值有死區 9–13、兩側各半格。**角色建立也已解**：屬性 ＝ 5d6 取最高三顆、MAXCON ＝ 同擲法 ＋18、技能點 ＝ IQ（`docs/re/21`）|
| **戰鬥判定** | **兩條路徑都已定位**：隊伍打敵方（`0x1AF52`）與敵方打隊伍（`0x1B04C`）共用同一支累加器，判定前綴機器碼相同、只差 `jb`／`jnb`——累加的是隊伍成員的本事，方向隨攻守翻轉。傷害 ＝ 基底 ＋ Nd6，兩邊護甲都是 N 顆 d6 的吸收。**敵方 HP 在 `ds:46C8h + 編號×2`**，減到 ≤0 夾成 0；角色 CON 可為負並分五級傷勢（`docs/re/20`）|
| **效果與傷害** | **已解**：地圖事件對角色的所有效果由記錄 `+0x08`／`+0x09` 兩個 byte 描述（哪個欄位、加或減、固定值或 Nd6），共用 `sub_141FA` 一個出口。護甲吸收 ＝ **AC 顆 d6 的和**（`docs/re/19`）|
| **存檔** | **已解**：在 `game1`／`game2` 檔尾的一個 MSQ 資源（`0x000253C5`／`0x00028BC7`，**seek 是 32-bit 的 `cx:dx`**）。`0x800` 加密段 ＝ 8 × 256：第 0 筆是全域狀態（四組隊伍槽表、時鐘、32-bit 序號、地點名稱），第 1–7 筆是角色。**checksum ＝ 0 − Σ 明文位元組**，改寫必須重算。出廠的四個 Ranger（Hell Razor／Angela Deth／Thrasher／Snake Vargas）把角色記錄欄位表整個驗過一遍（`docs/re/30`）|
| **經驗值與升級** | **已解**：經驗值在角色記錄 `+0x21`–`+0x23`（24-bit 累計，升級**不扣**），**升到等級 L 要 (L² − L) × 512**（1,024／3,072／6,144…）。升級把等級寫 `+0x24`、技能點 `+0x20` ＋1、播音效 4、查 `ds:D522h`（等級 → 階級編號，**50 階對到等級 1–131**，`Private` 到 `General Argent`）決定要不要印升階訊息，經驗值夠可連升。技能學習：IQ 需求 ＝ 技能資料 `+0x00 >> 3`、基礎費用 ＝ `& 7`、升到等級 L 的費用 ＝ 基礎 × 2^(L−1)（`docs/re/31`）|
| **檢定與技能成長** | **已解**：技能資料表在 `ds:BA20h`（36 筆 × 2 bytes，第二個 byte 是檢定屬性的記錄位移）；35 條的 IQ 與費用**與官方手冊列的 27 條逐值吻合**。檢定 ＝ 2d6 續擲（< 5 直接失敗）＋ 屬性 ＋ 技能等級×3 **≥ 難度 × 5 ＋ 15**，成功之後技能還有機率自己 ＋1 級（上限 ＝ 角色等級）。地圖記錄 `+0x0A` 起每 2 bytes 一個條件（型別／難度／參數），涵蓋技能、屬性、等級、性別、隊伍人數、金錢、持有物品（`docs/re/32`）|
| **狀態與療傷** | **已解**：角色記錄 `+0x28` 是**八個狀態位元**（Radiation poisoning／Wasteland Herpes／Bug byte／Sewer rot／Desert dust／Rabies／D6／D7），高四位會隨時間惡化。每 16 刻跑一次體力處理：健康的人每 64 刻 ＋1 CON、生病的人每 64 刻 −1，CON 掉破 −50 直接歸零。設施身分也定了——`0x1C260` 是醫生、`0x1BBA0` 是技能訓練師（`docs/re/35`）|
| **角色記錄** | **定址已確認**：記錄 ＝ `0x7131 ＋ 角色編號 × 256`，每筆 256 bytes，經隊伍槽表兩層間接。名字、金錢（24-bit）、七個屬性、MAXCON／CON、**技能與物品兩個 30 槽陣列（已分辨）**、傷勢門檻（−11／−20／−30／−40）都已定位（`docs/re/15`） |
| 逐指令基準 | 整個 CODE 區倒成 JSON（20,177 條指令、827 個全域、4,932 筆直接定址存取），後續形狀比對改在離線做（`tools/ida/export_listing.py`、`export_memops.py`） |
| 儲存層 | 雙模式（硬碟 DOS 檔案／磁片 `int 25h` 絕對磁區）與分流旗標；資源表 8 筆全解，六個檔名的引用點就在表的 `+6` 欄位（`docs/re/05`） |
| 英文手冊 | 全文轉 markdown，7 章 646 行（`docs/manual/`） |
| 段落書 | 162 段全部轉錄，編號連續無缺（`docs/paragraphs/`）。**三層防拷結構已辨識**：3 個陷阱段落（1／22／145）、64 段變體組（同場景不同密語）、33 段火星誘餌假劇情 |

### 進行中／未開始

| 項目 | 狀態 |
|---|---|
| 解包映像實跑驗證 | **已做**。解包版在 headless DOSBox 跑起來，畫面與原版 `compare -metric AE` ＝ **0**（`docs/re/47` §3） |
| 資料格式 | 已解：`wla.bin`（overlay 程式碼）、`title.pic`（XOR 串流）、`colorf.fnt`（172 字 × 32 bytes，格式與用途都已驗證）、主文字字型（內嵌）、`GAME1`／`GAME2` 的定址方式。`allpics1/2` 的 82 張圖（`docs/re/23`）、`allhtds1/2` 的 9 組圖磚（`docs/re/24`）。`ic0_9.wlf`／`masks.wlf` 的格式與用途（`docs/re/24` §2.3）。未解：`allpics*` 交錯的參數區、`transtbl`、`curs`、`end.cpa` |
| 劇情敘述文字 | **已解**：執行檔九張打包表 426 條 ＋ 地圖區塊 4,401 條，合計 4,827 條非空（`docs/re/17`、`18`）。**與段落書的分工也解了**：`Read paragraph N.` 就是敘述字串的一部分（83 處、82 個編號），沒有段落機制；陷阱段落 1／22／145 零引用（`docs/re/33`）|
| MSQ 尾段 | 已解：無 magic 的 Huffman 流，42/42 解出 4,096 或 1,024 bytes ＝ 地圖第 3 層（每格 1 byte，`docs/re/24`） |
| **Huffman 解壓** | **已實作並驗證**（`tools/huffman.py`）：`allhtds1/2`、`allpics1/2`、`end.cpa` 共 173 個子區塊全部解出，長度精確吻合、檔案 100% 用完（`docs/re/11`） |
| 載入器分工 | 已解：`DL`＝0 ALLPICS、1 GAME／存檔、2 ALLHTDS、6 END.CPA，各有位移表 |
| 說明書整理 | 英文手冊、段落書、軟體世界中文說明書都完成；社群攻略未開始 |
| **規格（G2）** | **二十六份全部 READY**，沒有未寫的規格了（`docs/spec/00-index.md`）|
| **`internal/assets`** | **已實作並通過驗收**：SHA-256 驗證、資源定址、MSQ 解密、Huffman、5-bit 文字、兩套字型、圖片／圖磚／地圖三層。9 個測試全綠，含 `Raw` 的 byte-for-byte round-trip（`tools/go.sh test ./...`）|
| **`internal/textlayout`** | **已實作**：18 個控制碼含**變形機制與巢狀**（單複數／性別／三選一／數量）、組行與分頁。4,889 條語料全部排得過，未解碼只剩 `0x08` 的 7 次 |
| **`internal/render`** | **已實作**：合成 320 × 200 索引畫面。地圖視窗逐像素驗過剛好 288 × 128 @ (8,8)、捲動一格 ＝ 位移 16 像素 |
| **`internal/game/rng`** | **已實作**：進位鏈與四支擲骰 ＋ 5d6 取三。驗收數列（前七項 ＝ 二項式係數）、分佈、300 萬次不重複全過 |
| **翻譯目錄** | **管線已通**：`tools/extract_strings.py` 抽出 **4,827 條原文**（426 ＋ 4,401，與 `docs/re/17`／`18` 吻合）→ 人翻的 UTF-8 TSV → `tools/build_lang.py` 編成 Big5 `.cat`。**編譯時擋五種錯**：格數超過原文、控制碼數量不符、Big5 缺字、`\x10` 後的按鍵被翻掉、不可翻的槽被翻了。Go 只讀 `.cat`，不依賴任何編碼函式庫。端到端驗過：踩到有翻譯的格子畫面上出現中文。**已全部翻完：4,806／4,806 可翻條目**（原文 4,827 槽扣掉 21 個不可翻的槽——未用槽的解碼雜訊與純控制碼，清單在 `translations/untranslatable.tsv`，build 會擋「不該有譯文的 key 被翻了」）。共用文本層 `_shared.tsv` 的 185 條涵蓋 1,425 個 key |
| **中文排版** | **已決策並實作**：內部畫布拉到 **640 × 400**（原解析乾淨 2×），原版素材 nearest 放大、倚天 16 × 15 直繪不縮字（`rulebook/81`）。一個中文字剛好佔原版一個字元格，所以 `docs/re/25` 的座標完全不用重算、訊息視窗仍是 6 行 × 38 格。字型走 `STDFONT.15` ＋ `SPCFONT.15`，Big5 分區索引已過 oracle（「一」是一條橫線、「中」「猴」可辨識、全形標點不落 fallback）。**字型檔玩家自備，沒有時遊戲照跑英文**（`docs/spec/10`）|
| **`internal/play`** | **可以走的遊戲場景**：從出廠存檔開場（四個 Ranger、時鐘 01:00、座標 (55, 62)），方向鍵走規則層。**存檔寫回也通過**：什麼都沒改就寫回是零差異，走 40 步之後只有座標／視窗原點／時鐘那 10 個 byte 變動。`cmd/wasteland -mode play` 開視窗、`cmd/wl-shot -mode play` 無頭輸出 PNG。與 `internal/viewer`（純檢視器、零規則）分開兩個套件 |
| **`internal/input`／`ui`／`viewer`** | **已實作**：與函式庫無關的按鍵模型（ESC 取消、F10 離開）、Ebiten 上色與送圖、資產檢視器場景。`cmd/wl-shot` 無頭輸出 PNG（對拍用）、`cmd/wasteland` 開視窗 |
| Go 建置環境 | `docker/wasteland-go.Dockerfile`（golang 1.24 ＋ X11／GL 標頭，Ebiten 走 cgo 需要）。相依從**唯讀掛載的本機模組快取**當 file proxy 取得，`tools/go.sh` 仍 `--network none` |
| **`internal/game`** | **已實作**：走一步（四道閘、被擋住就什麼都不推進）、24 小時制時鐘、每 16 刻的體力恢復與惡化、事件分派骨架。9 個測試全綠，含「400 步裡至少走成 100 步」的假綠防呆 |
| **存檔與角色記錄** | **已實作**：`internal/assets/save.go`（容器、XOR、checksum 重算、四組隊伍槽表）＋ `internal/game/character.go`（欄位讀寫）。**round-trip byte-for-byte 通過**，改一個欄位只動那三個 byte。升級門檻、技能費用、學技能的兩道檢查也一起做完 |
| **世界事件與檢定** | **已實作**：section 記錄定址（型別 ＝ nibble）、條件串列、技能與屬性檢定、檢定成功的自動練等、物品消耗、地圖腳本直譯器（遭遇控制／分支／狀態／變性等已實作，其餘明確回報「還沒做」不假裝成 nop）。掃過 42 張地圖：2,699 個 nibble 2 的格子解出 7,181 條條件、1,445 個 nibble 6 的格子 opcode 全在 0–43 |
| **設施** | **已實作**：商店價格（基礎價 −(基礎價 >> n)）、醫生的檢查／療傷／治病三種收費（價格欄位在地圖記錄 `+0x04`／`+0x05`／`+0x06`）、技能訓練師走既有的學技能公式。踩到設施格會顯示地點名稱（`docs/spec/09`）|
| **戰鬥回合** | **已解**：戰場是三組 × 每組 10 個，行動值 ＝ 2d6 續擲 ＋ 敵人資料 `+0x02` × 8，寫進 `ds:7931h` 的**固定格子**行動表，雙方排在同一張表。**誰先動也解了**：原版沒有排序演算法，是「線性掃一遍挑最大、做完清零、再掃」，而且比較是 `A ≥ B` 所以**平手取後面的格子**——隊伍排在敵人後面，平手時隊伍先動。剩逃跑與隊形未解（`docs/re/36`、`docs/spec/12`）|
| **敵人資料表** | **已解（8 個 byte 全部）**：`+0x00/+0x01` 基礎血量兼經驗值基值、`+0x02` 行動值、`+0x03` 傷害骰數、`+0x04` 低 4 位經驗值倍數 −1、`+0x05` 低 4 位**武器類別**（與物品表同一套編碼）高 4 位傷害基底、`+0x06` **敵人種類**（1 Animal／2 Mutant／3 Humanoid／4 Cyborg／5 Robot，397 筆全在 1–5，AI 對 1／4／5 有特判）、`+0x07` **肖像圖編號**（`ALLPICS`；同一個編號也決定文字碼 `0x0E` 的 him／her／it）。⚠ `+0x04` **高 4 位沒有任何讀者**，資料裡有值但程式沒讀（`docs/re/37` §3.2）|
| **起始裝備** | **已解**：`sub_1C9DE` 發三張清單（`ds:DECFh`／`DED9h`／`DEE3h`），`roll(1..2)` 挑 `.45` 或 9mm 手槍 ＋ 八個彈匣，再加繩／水壺／撬棍／刀／鏡子／火柴。**物品陣列的附屬 byte 初值 ＝ 物品表 `+0x04`（容量）**。十個編號翻出來全部合理，是物品表編號的獨立驗證（`docs/re/21` §5.1）|
| **`wla.bin` overlay** | **26 個 slot 全部有語意**（`docs/re/04` §3）。關鍵發現：`sub_10EBE`（14 個呼叫端、幾乎每個畫圖 slot 開頭那一句）**不是繪圖的一部分，是「畫之前先把滑鼠游標收起來」**；slot 20／21／22 是同一組的存背景／還原／防閃爍，游標尺寸 **24 × 16**。slot 6／7／8 是同一支「填一行」，填的值是自我修改的立即數 |
| **存檔改寫** | **實機驗收過**：round-trip 4,614 bytes 完全相同；改隊伍座標與時鐘寫回去，**原版讀得進去**並停在指定的座標與時刻。同一幀與我們畫的逐像素比對，夜間與白天各 **0 個不同的像素**（`docs/re/49`）|
| **遭遇驅動器** | **已解**：`sub_11CD0`（678 bytes）是地圖與戰鬥中間那一層——掃描 → 抄一份每個角色的經驗值 → 畫面模式 ← 1 → 四組各問一次指令、一起結算 → 打完**前後相減**報「X gains N experience.」→ 回地圖。原版沒有「這一場拿了多少」這個欄位（`docs/re/51`）|
| **設施在資料裡的分布** | 42 張地圖合計 **7 筆**設施記錄（醫生 2、商店 1、訓練師 3、存檔 1），但**現在有格子指到的只有 2 筆**——其餘要靠密語或腳本改寫地圖格才走得到。只掃「格子指到的記錄」會得到 2 筆，看起來像「這遊戲沒有商店也沒有醫生」（`docs/re/29` §5.4.1）|
| **商店的兩個價格** | **已解**：買與賣是**同一個公式**（`基礎價 − 基礎價 >> n`），只差指數——買用商店記錄 `+0x03`、賣用 `+0x04`。`sub_1C1CC`（買）與 `sub_1C1C2`（賣）各 10 bytes，共用同一段尾巴（`docs/re/22` §3.1）|
| **圖片動畫（A9）** | **已解**：參數區是兩張表——A 是每個通道的**播放腳本**（第一個 byte 是初始延遲，其後 `格、延遲` 交錯、`0xFF` 循環），B 是每一格的**像素**。消費端是 overlay slot 17，依 BIOS 計時器 `0040:006C` 推進、逐格 **XOR 進 EGA 平面**。元素 word：相位 ＝ `w & 3`（起點 ＋ `2 × 相位`）、`列, 欄 ＝ divmod((w >> 2) & 0x3FF, 12)`、酬載 `(w >> 12) + 1` bytes、一 byte 兩像素。列位址表 `ds:8E09h` 是 slot 0 那張 `ds:8DF9h` 往後 8 筆，**y ＋ 8 已烘在表裡**。實機逐像素 0 差異；「一輪播完 XOR 回到底圖」在 82 張圖全部成立（`docs/re/23` §5、`tools/verify_pic_anim.py`）|
| **地圖疊圖** | **十張全部解出**：0 塗黑、1–4／6 是五種敵人（`ds:AA17h` 的種類→編號表）、5 寶箱、7 隊伍、8 輻射區、9 其他分隊。**nibble 9 ＝ 輻射區**（六張地圖的訊息全是輻射，疊圖是輻射三葉標誌），而且**只在夜間畫**——實機 05:56 有、06:00 沒有，同樣那幾格（`docs/re/48`）|
| **地圖畫面對拍** | **0 個不同的像素**（`docs/re/47` §6.4）。收斂過程 404 → 288 → 0 抓出兩個缺口：隊伍圖示 ＝ **疊圖編號 7**（`MASKS.WLF` ＝ 10 張 1-bit 遮罩），以及**原版地圖視窗最上面一列留黑**（內容從 `y = 9` 起；圖片視窗沒有這一列）|
| **閒置不推進時間** | **實機確認**：站著不按鍵 45 秒，原版時鐘一分鐘都沒走——**「站著讓時間流逝」不會發生**（`docs/re/47` §6.1）。**機制未解**：主迴圈讀鍵是非阻塞的，所以不是「卡在等按鍵」；線索與下一個入口在 `docs/re/26` §1.2 |
| **DOSBox 參考環境** | **已建**（`tools/dosbox.sh` ＋ `docker/dosbox/`，全程 headless）。第一批對拍：解包版與原版畫面**逐像素相同**；`TITLE.PIC` 解碼 **36,864/36,864 吻合**、位置 (8, 8) 由掃描得出、預設 EGA 調色盤同時獲證；隊伍名單畫面一次印證五條斷言，其中 `AMM 18` 直接驗掉「物品表 `+0x04` ＝ 容量」（`docs/re/47`）|
| **問答的實際分布** | **已掃完**：42 個區塊共 **單鍵 117 題、打字 68 題**（另有 28 筆標成打字但沒有一條答案打得出來 → 不是問答）。答案是 `MUERTE`／`THANATOS`／`DIPSTICK`／`ROSEBUD` 這類大寫英文密語，**126 條已接進翻譯建置守則**（`translations/must-not-translate.tsv` ＋ `build_lang.py`，實測會擋）|
| **按鍵巨集** | **已解**：F1–F10 播放、Alt+F1–Alt+F10 錄製，10 組 × 256 bytes（`ds:C062h`，狀態變數緊接在 `0xCA62` 之後）。錄的是**轉大寫之前**的原始鍵碼。**存檔那 `0xA00` 未加密尾段就是它**（`0xA00` ＝ 10 × 256），而且不進 checksum（`docs/re/43` §6、`docs/re/30` §6）|
| **改寫地圖格** | **已解**：`sub_17CFF`（7 個呼叫端共用）拿記錄裡的兩個 byte ＝（新的第 1 層 nibble、新的第 2 層記錄編號）直接改寫指定的格子；第一個 byte 的 bit7 設起來 ＝ 不改。**答對密語不是設旗標，是把那一格從「擋路」改成「通道」**（`docs/re/46` §4.1）|
| **打字回答與密語** | **已解**：nibble 8 是**問答不是選單**——記錄 `+0x00` 的 bit7 決定「按一個鍵」還是「打一行字」（上限 16 bytes）。答案清單在 `+0x03` 起，逐一解出來與輸入比對（`sub_18D8E`，逐 byte 全等），命中第 N 個走第 N 條分支。**按鍵在讀進來的當下就轉大寫**，所以比對端不折疊大小寫。角色名字走同一支輸入（上限 13 bytes）。中文化硬約束見 `docs/re/46` §6 |
| **物品資料表** | **已解（8 個欄位全部）**：`+0x03 >> 3` ＝ 類別（18 類；2–13 是「有射程的武器」，清單在 `ds:CD00h`）、`+0x04` 彈匣容量、`+0x05` 使用技能編號、`+0x06` 骰數（武器＝傷害 Nd6、護甲＝AC）、`+0x07` 彈藥物品編號。⚠ **表不在執行檔裡，在存檔區**，每個存檔槽各一份 —— 它是可變的遊戲狀態（`docs/re/45`）|
| **隊伍傷害的第一項** | **已解**：`sub_15755` ＝ `0 ＋（武器 +0x06）顆 d6`，反戰車武器（類別 8／9）顆數變 `2N − x`（主攻擊路徑 `x ＝ 0`）。**所以「五項相加沒有骰」是錯的**——第一項自己就在擲骰（`docs/re/45` §4）|
| **自動步** | **已解**：捲動跳表有 **5 筆**，第 5 筆（方向碼 4）是 `clc; retn`——隊伍不動，但遭遇擲骰與重畫照跑。兩個觸發點：節拍滿 `0x400`（無音效）與 `ds:46E1h` 非 0（有腳步音效）。與 `ds:916Bh` 的防刷是同一套設計的兩面（`docs/re/26` §1.1）|
| **檢定的經驗值開關** | **已解**：`ds:916Bh` 全檔只有一個寫入點（`sub_1651A`：`← 方向碼 − 4`）與兩個讀者（兩支檢定）。方向碼 0–3 是玩家走的、**4 是自動步**（時間流逝或 `ds:46E1h` 觸發），只有 4 會留下 0 → 檢定成功不給經驗值。**這是防刷**：站著讓時間過所觸發的檢定沒有經驗值（`docs/re/32` §7.1）|
| **遭遇與敵人血量** | **已解**：敵方記錄 94 bytes ＝ 4 bytes 標頭（遭遇的 x／y、地圖編號、與隊伍的距離）＋ 3 組 × 30 bytes（10 個 16-bit 血量 ＋ 10 個行動旗標），16 筆在 `0x6B31`。三組的型別與數量寫在地圖記錄 `+0x03`–`+0x08`；血量 ＝ `⌊基礎/4⌋ ＋ 1d(基礎低位) ＋ 256 × 1d(基礎高位)`，**每隻各擲各的**，基礎值取自敵人資料表 `+0x00/+0x01`（同一組 byte 也是擊殺經驗值的基值）。距離是 `ds:CD0Dh` 那張 5 × 10 的表，**逐格照抄不能用公式**（`docs/re/37`、`docs/spec/13`）|
| **地圖指令列** | **已解入口**：底部固定七項 `Use Enc Order Disband View Save Radio`，處理程式在 `ds:AB1Ch`（`sub_16C7C` 設 `ds:468Dh`／`ds:468Fh`，`ds:4689h ＝ 6` ＝ 最大索引）。**升級的入口是 RADIO**（`0x15260` → `loc_1B8AD`）——問一句、答 Y 就逐人檢查經驗值並升級，可連升、播音效 4。remake 已接 `Save` 與 `Radio`，其餘五項只定位到入口（`docs/re/91`）|
| **音效** | **已接**：`internal/audio` 的位元組碼直譯器 → `Synth`（方波，照 8253 除數算，不換算成浮點）→ Ebiten。走一步播 1、腳本 op 播 7、升級播 4（`docs/re/44` §6）|
| **行動順序** | **已解**：兩邊公式不同——敵人 ＝ 2d6 ＋ 敵人資料 `+0x02` **× 8**，隊伍 ＝ 2d6 ＋ Speed（角色記錄 `+0x11`）＋ Brawling × 3（**沒有 ×8**）。2d6 逢同點續擲。而且**只有下攻擊令的人才排進行動表**（`0x1AE78` 的 `cmp al, 2`）——迴避、換武器、使用物品的人這一回合連格子都不占（`docs/re/90`）|
| **敵人打誰** | **已解**：`roll(1..隊伍人數)` 隨機挑，挑到 CON ≤ 0 的人就**整個重抽**（`0x1B054`）——沒有仇恨、前排或輪流。重抽沒有上限，靠 `sub_19D0E`（還有沒有人能打）在進來之前擋住。⚠ `sub_172BB`（CON ≤ 0 ＝ 倒下）與 `sub_172AE`（CON ＝ 0 ＝ 死）是兩個判準，戰鬥全部走前者（`docs/re/89`）|
| **戰鬥指令與逃跑** | **已解**：一回合開頭逐人下令，指令碼寫進 `ds:46D8h`、參數寫進 `ds:46DAh`。**指令碼是熱鍵字母表 `' ' H A W R E L U` 的索引，不是選單的顯示順序**（選單印 Run/Use/Hire/Evade/Attack/Weapon/Load）。`ds:A43Bh` 是 10 筆處理程式跳表；**迴避的處理程式是空的**，效果全在命中門檻的基礎值（迴避 60 > 攻擊 50 > 其他 40）。**逃跑不擲骰也沒有失敗分支**：問 Party／Single，選一個方向就走。CON ≤ 0 的人不下令（`docs/re/38`、`docs/spec/14`）|
| **四支指令處理程式** | **已解**：Hire／Weapon／Load／Use 加上已解的迴避、攻擊、逃跑，**七支全部是同一個形狀**——「檢查前提 → 不成立就印一句話重問；成立就開選單、回傳選擇」。**它們都不執行動作**（換武器沒真的換、裝填沒真的填），動作在結算階段依指令碼與參數進行。物品陣列的數與找都認記錄 `+0xBD` 起 30 槽、stride 2。Use 是唯一會寫狀態的：把一個 byte 記在**以角色編號為索引**的一格（`ds:A9FDh`）（`docs/re/41`、`docs/spec/17`）|
| **設施的互動迴圈** | **已解**：商店主迴圈（選人 → Buy／Sell → 清單 → 換人／離開）；賣一件會**把那個槽清成 0、加錢、店家庫存 +1**，清單會標記裝備中的但不擋著不讓賣。**物品表 `+0x02` 是庫存量不是旗標**（0 ＝ 缺貨不列出、`0xFF` ＝ 無限、其餘賣一件 +1）——只看讀取端會以為是旗標，找到 `add al, 1` 才知道。醫生的逐點治療**每一輪重算 MAXCON − CON**，一次按鍵只治一點只扣一次錢，錢不夠就停在中途（`docs/re/42`、`docs/spec/18`）|
| **段落手札** | **已解＋已定案**：原版沒有段落機制，`Read paragraph N.` 就是一條普通敘述（`docs/re/33`）。引用表由 `tools/extract_paragraph_refs.py` 從**英文原文**抽成 `docs/re/generated/paragraph-refs.tsv`（83 條引用、82 個編號）——**不在執行期解析翻譯過的文字**，否則譯者的用字會決定遊戲讀不讀得到段落。手札**預設全開**（忠於紙本），分成正文（會被引用的 82 段）與附錄（其餘 80 段，含陷阱段 1／22／145），附錄標明是 1988 年防拷設計；**一段都不刪**（`docs/spec/19`）。**正文自成一個目錄檔**：譯稿 markdown（`translations/zh-Hant/paragraphs/`）→ `tools/build_paragraphs.py` → `translations/paragraphs-zh-Hant.cat`（key `para:<編號>`，與翻譯目錄同格式所以 Go 不必多一份讀檔器）→ `internal/play.Journal`。段落書是紙本、不在原版語料裡，所以**不算在 4,827 條**；譯稿 162／162 已完成|
| **輸入層與選項熱鍵** | **已解**：滑鼠不產生事件，只把游標位置換成一個按鍵碼再走鍵盤路徑（`sub_18EFE` 先問滑鼠、問不到才 `int 16h`）。21 筆熱區表 `ds:0CAEBh`，**第一個「遮罩有開且座標命中」的區域就決定結果**；遮罩是每個畫面在等待輸入前設的 32 位元字（`ds:7DF3h`／`7DF5h`）。**控制碼 `\x10` 的用途是滑鼠**：把緊接的字元登記到「每列一格」的熱鍵表 `ds:8DDCh`，點哪一列就送出哪一格。所以 **`\x10` 後面那個字母翻掉會同時弄壞滑鼠（送出 Big5 首位元組）與鍵盤（比對仍是原字母，畫面上卻沒有提示）**——寫法固定為 `\x10Y 是`，`tools/build_lang.py` 會擋（`docs/re/43`、`docs/spec/20`）|
| **遭遇怎麼冒出來** | **已解**：`sub_14664` 掃**地圖視窗那 9 × 19 格**，第 1 層 nibble ＝ 3 或 15 的就是遭遇格。取該格的地圖記錄，算與隊伍的距離：大於記錄 `+0x00`（察覺上限）整格略過；不大於 `+0x01`（主動距離）或小於該組的接戰值就進佇列。佇列在 `ds:A96Fh`，**4 隊伍組 × 4 槽 × 4 bytes（x／y／地圖／距離）**，槽內由近到遠排，**每次掃描整個清空重建**——它是視野快照不是待辦清單。掃完會把不在佇列裡的敵方記錄整筆清掉（`docs/re/39`、`docs/spec/15`）|
| **戰鬥畫面** | **已解**：**戰鬥沒有專屬畫面**——`ds:46B9h` 一翻，地圖視窗就換成隊伍名單（表頭在字元列 14、每人一行、六個欄位在 `0x11`/`0x15`/`0x18`/`0x1C`/`0x20`），訊息與選單都印在原本的訊息視窗。指令階段每問一個人就重畫一次名單。一回合的訊息序列（遭遇開始 → 下令 → 命中/未命中 → 傷勢 → 經驗值）逐條對到字串編號。**中文化兩條硬規則：熱鍵不跟著翻譯走、七個選項一行放不下要重排**（`docs/re/40`、`docs/spec/16`）|
| **戰鬥** | **已實作**：命中判定（兩邊方向相反、夾在 100）、敵方傷害（基底 ＋ Nd6）、隊伍傷害（五項相加、沒有骰）、護甲吸收 ＝ N 顆 d6、角色 CON 可為負與五級傷勢、敵人 HP 夾在 0、擊殺經驗值。**命中累加值的四個項全部接上**：基礎 ＋ Brawling（技能 1，寫死）× 3 ＋ Agility − 對手行動值（近戰類別先做 8-bit ×4，否則另加 5）；敵方那條的基礎值取被打者這回合的指令（迴避 60／攻擊 50／其餘 40），**迴避因此才真的有效**。隊伍那條的基礎值查表 `0x711D` 未解，用 60（`docs/re/88`）|
| **音效** | **已解**：跑在 `int 08h` 上的位元組碼直譯器（126.36 Hz，每 6 tick 鏈回 BIOS）。**四個聲部是優先序仲裁不是混音**——PC 喇叭一次只發一個音，編號小的贏。四個 opcode（設欄位／計數器迴圈／換寫入基底／音符），每個聲部有滑音、線性分段封套、顫音、移調。**九首全在 `seg005` 的 864 bytes 裡**，不在外部檔案。⚠ 半音表的 **E 走音 −2.91 音分**（原始資料的錯字，`0x714F` 應為 `0x711E`），重製版照抄不修。**呼叫端全檔掃過**：只有五個直接呼叫點（音效 1／2／4／5／7），0／3／6／8 沒有人播；**音效 6 是無限循環**，只有「全部靜音」關得掉（`docs/re/44`、`docs/spec/08`、`internal/audio`）|

## 3. 文件索引

| 文件 | 內容 |
|---|---|
| [`CLAUDE.md`](./CLAUDE.md) | 專案規範：四道閘門（含 G4 接線）、IDA 鐵則、文件與中文化政策、環境硬規則 |
| [`docs/re/00-master-index.md`](docs/re/00-master-index.md) | **RE 總表**：位址換算、資料格式、結構佈局、位址表、關鍵函式、工具。**查已知事實先看這份** |
| [`docs/re/00-remake-knowledge-gaps.md`](docs/re/00-remake-knowledge-gaps.md) | **RE 完成度檢查表**：remake 需要的每一項知識、狀態與入口 |
| [`docs/re/00-function-index.md`](docs/re/00-function-index.md) | 函式索引（641 個，已分析 435）。讀任何 `sub_XXXXX` 前先查 |
| [`docs/re/00-wiring-status.md`](docs/re/00-wiring-status.md) | **接線狀態**：95 份筆記的結論，remake 有沒有真的用上。`TestWiringStatus` 雙向守著（`CLAUDE.md` §0 的 G4）|
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
| [`docs/re/28-text-variants.md`](docs/re/28-text-variants.md) | 文字變形：單複數／性別／三選一／數量，四個碼的骨架與選擇子來源 |
| [`docs/re/29-map-event-handlers.md`](docs/re/29-map-event-handlers.md) | 地圖事件處理函式（寶箱／選單／訊息），以及強制分析 IDA 漏掉位址的做法 |
| [`docs/re/30-save-layout.md`](docs/re/30-save-layout.md) | 存檔佈局、checksum 算法、隊伍槽表、四個預設 Ranger |
| [`docs/re/31-experience-and-skills.md`](docs/re/31-experience-and-skills.md) | 升級門檻 (L² − L) × 512、階級表、技能學習與費用公式 |
| [`docs/re/32-skill-checks-and-xp.md`](docs/re/32-skill-checks-and-xp.md) | 資料表定址器、技能資料表、檢定與練等、條件串列、經驗值三來源 |
| [`docs/re/33-paragraph-references.md`](docs/re/33-paragraph-references.md) | 段落編號怎麼出現、陷阱段落零引用、密語是遊戲內謎題 |
| [`docs/re/34-map-script-opcodes.md`](docs/re/34-map-script-opcodes.md) | 地圖腳本直譯器的 44 個指令 |
| [`docs/re/35-status-and-healing.md`](docs/re/35-status-and-healing.md) | 八個狀態位元與疾病名、體力隨時間恢復與惡化、醫生與訓練師 |
| [`docs/re/36-combat-rounds.md`](docs/re/36-combat-rounds.md) | 回合結構、行動順序表、「掃描挑最大」而不是排序 |
| [`docs/re/45-item-data-and-weapon-damage.md`](docs/re/45-item-data-and-weapon-damage.md) | 物品資料表八個欄位、表在存檔區、武器傷害骰（`sub_15755`）、`ds:CD00h` 的武器類別清單 |
| [`docs/re/37-enemy-records-and-hp.md`](docs/re/37-enemy-records-and-hp.md) | 敵方記錄的兩層版面、敵人血量的擲法、敵人資料表、距離表 |
| [`docs/re/38-combat-commands-and-flee.md`](docs/re/38-combat-commands-and-flee.md) | 戰鬥的指令階段、七個指令碼與跳表、迴避與逃跑 |
| [`docs/re/39-encounter-scan.md`](docs/re/39-encounter-scan.md) | 遭遇怎麼冒出來：視窗掃描、遭遇佇列、兩個距離門檻 |
| [`docs/re/40-combat-screen.md`](docs/re/40-combat-screen.md) | 戰鬥畫面：名單模式、一回合的訊息序列、中文化的兩條硬規則 |
| [`docs/re/41-command-handlers.md`](docs/re/41-command-handlers.md) | 四支指令處理程式：共同形狀「檢查 → 選參數 → 回傳」，不執行動作 |
| [`docs/re/42-facility-loops.md`](docs/re/42-facility-loops.md) | 商店買賣的挑物品流程、逐點付錢治療、物品表 `+0x02` 是庫存量 |
| [`docs/re/43-input-and-hotkeys.md`](docs/re/43-input-and-hotkeys.md) | 鍵盤的三種比對、21 筆滑鼠熱區表、`\x10` 登記的每列熱鍵表 |
| [`docs/re/44-audio.md`](docs/re/44-audio.md) | 音效：計時器 ISR、四個聲部、位元組碼指令集、九首資料與呼叫端掃描、走音的 E |
| [`docs/re/46-typed-answers-and-text-input.md`](docs/re/46-typed-answers-and-text-input.md) | 打字回答與密語比對、文字輸入常式、中文化硬約束 |
| [`docs/re/47-dosbox-oracle.md`](docs/re/47-dosbox-oracle.md) | DOSBox 參考環境；解包正確性、逐像素對拍、地圖視窗那一列 |
| [`docs/re/48-map-icons.md`](docs/re/48-map-icons.md) | `IC0_9.WLF` 十張疊圖的語意、nibble 9 ＝ 輻射區、實機對拍 |
| [`docs/re/49-save-roundtrip-on-hardware.md`](docs/re/49-save-roundtrip-on-hardware.md) | 存檔改寫的實機驗收：寫得回去、原版讀得進去、兩張畫面各 0 個不同的像素 |
| [`docs/re/50-unnamed-items.md`](docs/re/50-unnamed-items.md) | 物品 70／71／72 的空名字：被清空不是遺失、字母序夾出 H–M、哪些欄位不算證據 |
| [`docs/re/51-encounter-driver.md`](docs/re/51-encounter-driver.md) | 遭遇驅動器 `sub_11CD0`：地圖與戰鬥之間那一層、四組一起結算、經驗值前後相減 |
| [`docs/re/52-trainer-facility.md`](docs/re/52-trainer-facility.md) | 技能訓練師的流程：五個設施同一個模板、三條「走不通」都回選人 |
| [`docs/re/53-list-framework.md`](docs/re/53-list-framework.md) | 清單框架：列與索引的對應表、三個回傳值、I／K 翻頁、每頁列數不是常數 |
| [`docs/re/54-facility-screen-layout.md`](docs/re/54-facility-screen-layout.md) | 設施畫面版面：圖在 (8, 8) 96 × 84、地點名字元列 12、殘差 ＝ A9 的動畫 |
| [`docs/re/55-radiation-and-armour-bypass.md`](docs/re/55-radiation-and-armour-bypass.md) | 輻射結算：每人 `+0x01` 顆 d6 扣 CON ＋ 中毒；`ds:46EFh` ＝ 這次跳過護甲吸收 |
| [`docs/re/56-transtbl.md`](docs/re/56-transtbl.md) | `TRANSTBL` ＝ 50 組 × 16 對照表，載入之後沒有人讀；滑鼠初始化 |
| [`docs/re/57-curs.md`](docs/re/57-curs.md) | `CURS` ＝ 8 個 32 × 16 的滑鼠游標，左半遮罩右半圖形；EGA 平面連續 |
| [`docs/re/58-line-flush-and-scrollback.md`](docs/re/58-line-flush-and-scrollback.md) | 控制碼 `0x08` ＝ 沖出這一行不捲動；scrollback ＝ `seg003:0x8CE0` 的 40 × 256 環形 |
| [`docs/re/59-playtest-against-original.md`](docs/re/59-playtest-against-original.md) | 正常玩家路徑對原版驗收：輻射帶團滅是原版行為、Rad suit 免疫、熵沒接上、出不了起始地圖 |
| [`docs/re/60-teleport-and-map-change.md`](docs/re/60-teleport-and-map-change.md) | 傳送會換地圖；槽表 `+0x0B`–`+0x0D` 是**回程**不是目的地 |
| [`docs/re/61-map-id-table.md`](docs/re/61-map-id-table.md) | 地圖編號表 `ds:BF1Ch`：bit7 設 → 建築內部（資源 5／11）；nibble 11 是第二多的地形 |
| [`docs/re/62-fourth-gate-terrain-blocking.md`](docs/re/62-fourth-gate-terrain-blocking.md) | 第四道閘：**nibble 11 是山與牆**（20,495 格），擋住時印記錄的訊息 |
| [`docs/re/63-resource-id-vs-index.md`](docs/re/63-resource-id-vs-index.md) | 資源目錄的 **ID ≠ 切片索引**（42 個裡 28 個不同）；遊戲的地圖編號一律是 ID |
| [`docs/re/64-enter-location-prompt.md`](docs/re/64-enter-location-prompt.md) | 進新地點要先問 `Enter new location?`；判準是記錄 `+0x00` 的 **bit6** |
| [`docs/re/65-third-gate-conditions.md`](docs/re/65-third-gate-conditions.md) | nibble 2 是**條件式**：bit7 或 ¬bit6 直接放行（2,553 格），只有 146 格要判定 |
| [`docs/re/66-nibble2-event-and-heat.md`](docs/re/66-nibble2-event-and-heat.md) | nibble 2 的閘與事件是同一支；沙漠高溫的訊息與扣血路徑已定位，入口未解 |
| [`docs/re/67-gate-penalty-and-canteen.md`](docs/re/67-gate-penalty-and-canteen.md) | 條件閘的獎懲在記錄 `+0x08`／`+0x09`；高溫的條件是物品 44 ＝ `Canteen` |
| [`docs/re/68-cell-rewrite.md`](docs/re/68-cell-rewrite.md) | 改寫地圖格 `sub_17CFF`：七個呼叫端只差一個位移；**這個遊戲的狀態就是改格子**；游不過河會被沖到下游（四邊對拍） |
| [`docs/re/87-enemy-map-movement.md`](docs/re/87-enemy-map-movement.md) | `sub_15036` ＝ 敵人在地圖上移動（`move to a better position.` 那三句），**不是**目標表 |
| [`docs/re/86-combat-messages.md`](docs/re/86-combat-messages.md) | 戰鬥訊息的主詞與受詞：敵人名稱 ＝ 字串 `0x52 + Kind`（單複數格式取單數） |
| [`docs/re/85-enemy-map-icon.md`](docs/re/85-enemy-map-icon.md) | 地圖上的敵人圖示：section 15 種類 → 敵人資料 `+0x06` → `ds:AA17h`（三段都早已解過，缺的是接線） |
| [`docs/re/84-render-coverage.md`](docs/re/84-render-coverage.md) | 呈現層三道門檻：42 張地圖、23 家設施、戰鬥畫面都畫得出來且值域正確 |
| [`docs/re/83-translation-coverage.md`](docs/re/83-translation-coverage.md) | 中文化覆蓋率：4,806／4,827，缺的 21 條全在 untranslatable 清單裡，孤兒 key ＝ 0 |
| [`docs/re/82-save-round-trip.md`](docs/re/82-save-round-trip.md) | 存檔的三道門檻：byte-for-byte round-trip、改動限縮、存讀一致 |
| [`docs/re/81-combat-loop-coverage.md`](docs/re/81-combat-loop-coverage.md) | 戰鬥迴圈端到端門檻；抓到「背包槽號當物品 ID 查表」讓傷害高十倍 |
| [`docs/re/80-trainer-skill-list.md`](docs/re/80-trainer-skill-list.md) | 訓練師列**整張技能表**，篩選（IQ／費用／技能欄空位）在選完之後 |
| [`docs/re/79-facility-coverage.md`](docs/re/79-facility-coverage.md) | 設施覆蓋率：跳表索引 **≥ 5 就是 opcode**（9 筆記錄）；訓練師八家清單全空 |
| [`docs/re/78-encounter-spawn.md`](docs/re/78-encounter-spawn.md) | **遭遇生成器 `sub_16890`**：找空槽 → 擲種類 → 沿九向之一走 N 步找空地 → 放 nibble 15 |
| [`docs/re/77-encounter-spawn-gap.md`](docs/re/77-encounter-spawn-gap.md) | **隨機遭遇完全沒發生**：敵人格是 `sub_16890` 每步生成的，remake 只做了擲骰那一段 |
| [`docs/re/76-script-opcode-coverage.md`](docs/re/76-script-opcode-coverage.md) | 腳本 opcode 覆蓋率：**有格子指到的一個都不缺**（0 格），剩 17 種 0 格的 |
| [`docs/re/75-desert-heat-entry.md`](docs/re/75-desert-heat-entry.md) | **沙漠高溫的入口**：腳本 opcode 3 依晝夜把格子換成 nibble 2 記錄 7–9／10–12，跑完 `fd fd` 改回原樣 |
| [`docs/re/74-heat-entry-and-gate-display.md`](docs/re/74-heat-entry-and-gate-display.md) | `sub_142ED` ＝ 暫換時鐘的時 ＋ 印 `+0x03` ＋ 延遲 ＋ 還原；高溫入口再排除三條 |
| [`docs/re/73-shop-and-doctor-entry.md`](docs/re/73-shop-and-doctor-entry.md) | **商店與醫生的入口**：傳送記錄 `+0x04`／`+0x05` 在收尾把落點改寫成 nibble 6 ＋ 設施（22/22） |
| [`docs/re/72-facility-entry-and-command-bar.md`](docs/re/72-facility-entry-and-command-bar.md) | 進地點的完整路徑（nibble 6 → nibble 12 改寫 → nibble 10 → 確認）；地圖指令列 `USE ENC ORDER…` |
| [`docs/re/71-nibble12-batch-patch.md`](docs/re/71-nibble12-batch-patch.md) | nibble 12 是**遠端批次改寫器**（2,450 筆）；商店入口的資料側四條路也掃完 |
| [`docs/re/70-nibble1-and-facility-entry.md`](docs/re/70-nibble1-and-facility-entry.md) | nibble 1 ＝ 氛圍敘述串列 ＋ 收尾改寫；`0xFE`／`0xFD` 的 153 處用途；商店入口再排除四條 |
| [`docs/re/69-gate-flags.md`](docs/re/69-gate-flags.md) | 條件閘的四個旗標（記錄 `+0x00` 低位）；條件串列的 `0xFF` 之後接一張**逐條件改寫表**；`0xFE`／`0xFD` 的沿用暫存 |
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
| 未解的文字控制碼有四個、在語料裡出現 2,127 次，是呈現層最大的缺口 | `0x0A` 在 `docs/re/14` 的控制碼表裡掛「未解」，但 `docs/re/17` §4.1 早就從字串形狀認出它是單複數分隔——**同一個 byte，兩份文件兩種狀態**，統計時照著沒對齊的那一份數 | `0x0A`／`0x0C`／`0x0E` 是同一套文字變形機制（單複數／性別／三選一）、`0x0F` 印數量、`0x07` 開文字框，全部已解。**只剩 `0x08` 未解，語料裡 7 次**（`docs/re/28`） |
| 存檔在 `GAME1` 的 `0x53C5`、`GAME2` 的 `0x8BC7` | 只讀了 `mov dx, 53C5h` 就當成位移，漏掉下一行的 `mov cx, 2`——16-bit 程式的檔案位移是 `cx:dx` 兩個暫存器 | `0x000253C5`／`0x00028BC7`，是檔尾的一個 MSQ 資源（`docs/re/30` §1） |
| 控制碼 `\x10` 只是「顯示用的熱鍵標示」，翻成 `\x10是` 沒問題 | 只讀到寫入端（`ds:4703h` 旗標 → `ds:8DDCh[列]`），沒追讀取端就下結論。**唯一的讀者是滑鼠**，而測試從來只用鍵盤，所以壞掉不會有人發現 | 那個字元是「點這一列等同按哪個鍵」，翻掉會同時弄壞滑鼠與鍵盤提示；寫法固定為 `\x10Y 是`（`docs/re/43` §5）|
| `0xFE`／`0xFD`（沿用上一格改寫前的值）在出貨資料裡一筆都沒用到 | 掃描是**逐格**走的，只看得到「有格子指到的記錄」；而改寫可以把任何一格變成任何一筆記錄，沒有格子指到的記錄一樣會被執行 | 逐 section 記錄掃是 **153 處**，最大的用戶是設施記錄的 `+0x01`／`+0x02` ＝ `fd fd`＝跑完把這一格改回原樣（`docs/re/70` §3） |
| nibble 12 只是「印訊息那一族」 | `docs/re/26` §5 只讀了 `0x12BD0` 的第一行；那一行正確，但後面 120 個 byte 是另一回事 | nibble 12 是**遠端批次改寫器**：`+0x01` 起每 5 bytes 一筆（旗標／目標 x／目標 y／新第 1 層／新第 2 層），可以改寫地圖上任意多格，出貨資料共 2,450 筆（`docs/re/71`） |
| nibble 1 是「遠看才顯示的描述，站上去反而不印」 | `docs/re/26` §5 只讀了 `sub_16CD0` 的前幾行就下註解，那一段其實是「連續踩同一種格子不重印」 | nibble 1 是**氛圍敘述**：記錄 `+0x00` 起的訊息串列（bit7 結束，可以多條），串列長度就是收尾改寫的位移（`docs/re/70` §1） |
| 資源 0 的 section 2 記錄 2 是「沙漠」，踩壞會被傳送 | 只看到「條件失敗 → 改成傳送格」的形狀，沒讀 `+0x03` 的訊息 | 那是**河流**：`You are caught by the river's current.`，條件是技能 7（游泳）難度 6；傳送記錄 `0x22` 的 bit7 設、位移 (0,+1) ＝ **往下游沖一格**，下游那一格是「平靜的水」（`docs/re/68` §3） |
| 高溫記錄 7–12「沒有格子指到，所以入口未解」 | 同一個錯誤前提：初始地圖上沒有，不代表遊戲中沒有。它們是**每走一步當場貼上、跑完撕掉**的 | 腳本 opcode 3（`0x1A526`）依晝夜把 nibble 6 記錄 1–3 那一格換成 nibble 2 記錄 7–9（白天）或 10–12（夜間），條件閘收尾的 `fd fd` 再改回原樣（`docs/re/75`） |
| `Script`（腳本直譯器）的 `Record[0]` 就是 opcode | 沒追 `sub_12C80` 的腳本路徑——那裡先 `sub_17CB1(bl ＝ 0x10)` 把索引換成 word | `Record[0]` 是 section `0x10` 的**索引**，取出來的 word 才是 opcode。資源 0 記錄 1 的 `+0x00 ＝ 1` 對到的是 opcode **3** 不是 1（`docs/re/75` §1） |
| 設施記錄「只有 2 筆有格子指到，所以大部分進不去」 | 統計本身算得正確，但它預設「格子是靜態的」——在會自我改寫的系統裡，以初始狀態為基礎的可達性統計沒有意義 | 設施格是**傳送收尾當場改寫出來的**：`sub_169B1(4)` 用傳送記錄的 `+0x04`／`+0x05` 改寫落點，599 筆傳送記錄裡有 22 筆這樣做，而那 22 筆全部指到設施（`docs/re/73`） |
| 條件閘的收尾只有「全部人都過 ＝ 位移 4」與「有人沒過 ＝ 位移 6」兩種改寫 | 只讀了 `0x1406D` 與 `0x14045` 兩個 `mov al, <常數>`，沒追 `0x1405D` 那條——它的 al 是 `sub_142B1` 算出來的 | 記錄 `+0x00` 的 `& 8` 那一族（61 筆）改寫位移**依通過的是哪一條條件而不同**：條件串列的 `0xFF` 之後接一張逐條件的改寫表（`docs/re/69` §4） |
| `sub_15036` 是原版的戰鬥目標表 | 寫在 `internal/play/round.go` 的註解裡，沒有文件支持也沒有人查過 | 那一支是**敵人在地圖上移動**：清舊格、寫新格、更新遭遇佇列，訊息是 `move to a better position.`／`run away.`／`run at you.`（`docs/re/87`） |
| 物品資料 `+0x03` ＝ 傷害骰數、`+0x05` 高 4 位 ＝ 傷害基底 | 只讀了 `sub_12A76` 內部對 `ds:4665h` 的存取，沒回頭看 `sub_12A40` 把定址器的基址換成哪一張表——原版只有一支定址器，基址（`ds:4694h`）與 stride（`ds:469Bh`）都是全域變數 | 那是**攻擊資料表**（記錄區標頭 `+0x04`，8 bytes 一筆）的欄位；物品表的 `+0x03` 是**物品類別**（`docs/re/32` §1、§8.1） |
| 物品資料 `+0x03` 同一個 byte 有「類別」與「傷害骰數」兩種語意，兩邊都要保留 | 同上——那是兩張不同的表，不是一個欄位兩種用途 | 物品表 `+0x03` 只有「類別」一種語意（`docs/re/29` §4） |
| 存檔全域狀態的地點名稱在 `+0xC8` | `tools/dump_save.py` 印的時候做了 `.strip('.')`，把前面八個 NUL 一起吃掉，看起來剛好是 `Ranger Ctr.`——**過濾器把偏移蓋住了**。Go 那邊沒有那個 strip，一寫就露餡 | `+0xD0`（＝ 記憶體 `ds:7201h`，全域記錄載到 `0x7131`）。`docs/re/30` §3.2 補上「全域位移就是記憶體位移」這條換算 |
| 只有 section 型別 3、5、15、16、17 會被當成指標陣列取用 | 依據是 `sub_17CB1` 的 18 個靜態呼叫端——它們傳的都是常數。真正的通路是 `sub_169EB`，它把**那一格的 nibble** 當型別動態傳進去，靜態呼叫端清單看不到 | **section 型別 ＝ 第 1 層的 nibble 本身**（`docs/re/16` §3.1）。實作照這條改之後，42 張地圖的 2,699 個 nibble 2 格子解出 7,181 條條件、1,445 個 nibble 6 格子的 opcode 全部落在 0–43 |
| `ds:46C8h + 小位移` 有兩種讀法，可能是兩筆不同的記錄 | 只追了讀取點。`ds:46C8h` 是**指標**，全檔 55 處存取裡只有 2 處在寫它——寫入點才是答案 | 一筆 94-byte 記錄 ＝ 4 bytes 標頭 ＋ 3 組 × 30 bytes。`sub_137F4` 停在標頭（座標／地圖／距離）、`sub_137CE` 停在某一組（10 個 16-bit 血量 ＋ 10 個行動旗標），兩種讀法都對（`docs/re/37` §1）|
| `sub_1B108` 裡那段是「距離修正」 | 憑函式名旁邊有 `sub_13A56`（拆組編號）就把整串當成距離計算，沒讀到 `sub_1B15F` 的最後一步 | 那一串是定址，實際取的是**敵人資料表 `+0x02`**（`docs/re/20` §3、`docs/re/37` §3.1）。真正的距離是 `sub_19D4D` 查 `ds:CD0Dh` |
| 距離表 ＝ `trunc(10 × √(dx²+dy²))` ＋ 兩格例外 | 先用四捨五入比，看到 5 格不符就改成截斷，只驗了那 5 格有沒有變好，**沒有把 50 格重跑一次**——截斷其實有 12 格不符 | 那 50 格是**資料**，任何取整都重現不了；remake 逐格照抄（`docs/re/37` §4）。抓到它的是單元測試裡寫死的原始 bytes |
| A9 的疊法「已解」 | 只讀了 `sub_10B11` 的程式碼就下結論，**沒有回頭重建畫面驗證**。讀得懂每一條指令不等於重建得出輸出 | 缺的是兩件結構性的事：表 A 第一個 byte 是**延遲**不是格編號，以及**相位 ＝ 左邊缺幾對像素**（起點 ＋ `2 × 相位`）。補上之後實機差 0，且「一輪播完回到底圖」在 82 張圖全成立（`docs/re/23` §5.3）|
| 閒置不推進時間是因為「主迴圈卡在 `sub_18E90` 的忙等迴圈裡」 | 只讀了 `sub_18E90` 有忙等，沒有回頭確認地圖主迴圈是不是真的走那一支 | 地圖主迴圈讀鍵用的是**非阻塞**的 `sub_18EFE`（`0x16BC1`）。**現象仍然成立**（實機 45 秒時鐘沒動），但機制未解，線索在 `docs/re/26` §1.2 |
| 隊伍打敵方的傷害是「五項相加，沒有骰」 | 五個來源都讀到了，但只讀到「呼叫誰」，沒把 `sub_15755` 讀完——它的最後一條是 `jmp sub_19D86`，而那支正是 `docs/re/13` 早就驗過的「基底 ＋ N 顆 d6」 | 第一項就是**武器的傷害骰** `0 ＋（物品表 +0x06）顆 d6`，另外四項才不擲骰（`docs/re/45` §4）|
| `sub_15755` 是「距離／射程項」 | 名字旁邊讀到 `ds:46CCh`／`46CDh`，就把它當成距離相關；那兩個其實是接戰值與武器類別 | 它是武器傷害骰；反戰車武器（類別 8／9）顆數變 `2N − x`（`docs/re/45` §4）|
| `ds:DECFh`／`DED9h`／`DEE3h` 是名字輸入樣板 | 建角色時看到「擲性別 → 選一張表」就照直覺命名，沒讀 `sub_1C9DE` | 那是三張**起始裝備清單**：手槍 ＋ 八個彈匣（依 `roll(1..2)` 二選一）＋ 繩／水壺／撬棍／刀／鏡子／火柴（`docs/re/21` §5.1）|
| 物品資料 `+0x04` ＝ 彈匣容量 | 只看武器那幾列（`.45` 是 7、Uzi 是 40），沒看非武器：Match 是 40、Rope 是 1 | 是**容量**——這件東西裝滿時能用幾次；發裝備／換彈時寫進物品槽的附屬 byte（`docs/re/45` §3.2）|
| nibble 8 是「多選一的選單」，選項範圍由 `sub_1721B` 把關 | 只讀到 `0x1518E` 的單鍵分支就停了，`jns` 的另一邊整個打字流程在文件裡等於不存在；`sub_1721B`（挑隊員 1–9）根本不在這條路上 | nibble 8 是**問答**：`+0x00` 的 bit7 決定單鍵或打字，答案清單在 `+0x03` 起、逐一比對（`docs/re/46` §4）|
| 捲動跳表 `ds:AAA7h` 有 4 筆（四個方向） | 只數了「四個方向處理程式」就回頭認表，沒用下一張表的位址去定邊界 | **5 筆**：`0xAAA7 + 5 × 2 ＝ 0xAAB1` 正好是下一張表。第 5 筆（索引 4）是 `clc; retn`，給「原地不動的一步」用（`docs/re/26` §1.1）|
| `ds:46EFh` ＝「踩在輻射上」的旗標，而且只有一支寫它 | 只掃到一個寫入點就下結論。實際有三處，漏掉的兩處（設定與清除）正是決定語意的那兩處 | 它是**參數不是狀態**：設 → 呼叫結算 → 清。效果是**這一次結算跳過護甲吸收**，值來自地圖記錄 `+0x00` 的 bit0（`docs/re/55`）|
| nibble 4／9 的訊息編號 ＝ 這一格的第 2 層值 | 第 2 層是「第幾筆記錄」，訊息編號在**記錄 `+0x00`**。兩者值域差很遠，但都是小整數，錯了只會印到別條訊息、不會噴錯 | `sub_16D1A` 讀 `[ds:46AEh + 0]`，0 就不印。資源 0 的輻射格編號是 23 ＝ `The ground seems to glow here.`（`docs/re/29` §2、§5.1）|
| 輻射結算的 `cmp al, 29h` 是拿角色記錄 `+0x01` 比 `0x29`，用意未解 | 只讀了那三行，沒追 `bl` 從哪來——`sub_196C4` 先把 `bl` 設成護甲槽指到的位移 | 比的是**護甲的物品編號**：41 ＝ `Rad suit`，穿著的人整個跳過，不扣血也不中毒（`docs/re/59` §2.2）|
| 進新地點的詢問由記錄 `+0x00` 的 **bit7** 決定 | 漏看了 `jns` 前面那一條 `shl al, 1`——移位之後的符號位是原本的 bit6 | **bit6**。Quartz 入口是 `0x41`（bit6 設、bit7 沒設）所以會問（`docs/re/64`）|
| 地圖編號可以直接拿去索引 `Resources()` 的切片 | 兩套編號的前 14 筆剛好相同，所以早期沒踩到 | **42 個區塊裡 28 個索引 ≠ ID**（從索引 14 起偏移）。遊戲的地圖編號一律是**目錄 ID**，拿索引去載入不會報錯、只會走進別的城鎮（`docs/re/63`）|
| 擋住移動的 nibble 只有 2、3、10、15 | `docs/re/26` §3 誠實寫了「下界不是全集」，但那句話擺著沒人回頭補；第四道閘 `sub_15CE0` 一直沒讀 | **nibble 11 一律擋**（山與牆，20,495 格、42 張地圖全部都有），nibble 4 條件式。實機對拍才發現：同一串按鍵原版被山擋住而 remake 穿了過去（`docs/re/62`）|
| 隊伍槽表 `+0x0B`／`+0x0C` 是 nibble 10 的**傳送目的地** | 只追了讀取點。`0x16A10` 的第一個動作就是把**目前**座標與地圖寫進 `+0x0B`–`+0x0D` | 那是**回程**。目的地在地圖記錄 `+0x01`／`+0x02`（座標）與 `+0x03`（**目標地圖編號**，`0xFF` ＝ 回程）；`+0x0D` 之前沒有語意，它是回程的地圖編號（`docs/re/60`）|
| 敵人資料 `+0x07` ＝ 代名詞索引 | 只追了一個消費者（`sub_12A4C` → `ds:A920h` → 文字碼 `0x0E`）就命名，漏掉 `0x1268F` 那一個 | **肖像圖編號**（`ALLPICS`）：`sub_190A8` → `sub_184E8` 就是圖片載入器，而 `ds:A920h[肖像編號]` 決定 him／her／it——一個編號兩個用途，講的是同一件事（`docs/re/37` §3.2）|
| 資源表 idx 7（無檔名、只有磁區座標）很可能是存檔區 | 依據是「唯一只能走磁區路徑的一筆」＋「`int 26h` 只有一個呼叫端」，兩個都是旁證，沒有去看**誰在用索引 7** | **沒有人用**。`sub_11445` 的 10 個呼叫端傳的索引只有 `{0,1,2,6}`，`+3` 切換也只把 0–2 變成 3–5。存檔在 `GAME1`／`GAME2` 檔尾的 MSQ 資源（`docs/re/30`）；idx 7 是磁片版的遺留（`docs/re/05` §3）|
| 設施 3 是**存檔處**（remake 的 `FacilitySave`） | 五種設施裡只有它沒有明顯的店面行為，就照「Ranger Center 會存檔」的印象命名，沒去讀 `0x1A2C0` 的選單字串 | `ds:CE42h` → `ds:CE12h` ＝ **`CREATE DELETE PLAY`**：那是**角色管理**畫面（建立／刪除／開始遊戲）。存檔走的是指令列的 `Save`（`docs/re/72` §3、`docs/re/91`）|
| 隊伍成員的行動值也走 `sub_1B15F`，跟敵人同一條公式 | `docs/re/36` §2 只讀了敵人那個迴圈，看到「雙方排在同一張表」就假設公式共用；`sub_1B15F` 另外兩個呼叫端其實在命中判定那條路上 | 隊伍走 `0x1AE7C`：2d6 ＋ Speed ＋ Brawling × 3，**不乘 8** 也不碰 `sub_1B15F`（`docs/re/90` §1）|
| EGA 調色盤：「全檔 58 處 EGA 埠存取沒有一次碰 `0x3C0`」 | 掃描結論對（原版不設調色盤），但理由寫得太滿——實際上有一支碰 `0x3C0` 的常式 | 那支設的是**邊框色**（index `0x11`），而且入口被 patch 成 `retn`（`nullsub_1`），三個呼叫端全部沒作用；那 16 格仍然沒人寫（`docs/re/23` §7.1）|
| CON ≤ 0 的人只要 CON 不是剛好 0 就還能行動 | `Character.Dead()` 照 `sub_172AE`（CON ＝ 0）寫，而戰鬥的四個判斷全用了它。**CON 是有號的**，重傷會掉進負值 | 戰鬥一律用 `sub_172BB`（CON ≤ 0）：不能下令、不會被敵人挑中、不算在全滅判定裡。混用的症狀是**戰鬥永遠打不完**——全隊倒下卻誰都下不了令，而 `Over()` 認為隊伍還有人（`docs/re/89`）|
| `Use` 的三個選項照字串 4 的顯示順序編號（Item 0／Skill 1／Attribute 2） | `docs/re/41` 讀到「字母表在 `ds:A5E8h`」卻沒把它倒出來，直接照顯示文字「Use: Item / Skill / Attribute」編號 | 字母表是 `53 49 41` ＝ **`SIA`**，所以 **Skill 0／Item 1／Attribute 2**。`sub_173B0` 回的是字母表的索引，與顯示順序無關——與 `docs/re/38` §2「選單顯示的順序不是指令碼」同一個坑（`docs/re/92` §2）|
| 命中累加值裡的技能是「拿什麼武器就用什麼技能」，另外還有一個距離懲罰 | 兩個都是照戰鬥系統的常識填的：`HitChance` 收 `skillID` 與 `distancePenalty` 兩個參數，呼叫端傳 `w.Skill, 0, 0`——**編得過、測得過、玩得動**，沒有人回頭問那兩個零是什麼 | 技能編號**寫死是 1（Brawling）**（`sub_1B0F1` 第一條指令 `mov al, 1`）；被減的是**對手的行動值**（敵人資料 `+0x02`），與距離無關（`docs/re/88`）|
| MSQ 資源的前 2 bytes 是長度 | 把 `mov cx, [bx]` 當成「取剛讀進來的 header」，其實 `ds:46B0h` 是別處設好的緩衝區位址 | 那 4 bytes 是 magic `msq0`／`msq1`；長度來源仍未解（`docs/re/07` §7） |
| `END.CPA` 要先跳過 4 bytes 檔頭，再拿 `+0x04` 的 `msq` 當容器解密 | 把長度欄看成檔頭是憑長相猜的；接著猜了六種 checksum／body 起點的組合——**那已經不是推論是試參數** | 整份從第 0 個 byte 起就是 Huffman，`Decompress` 自己會讀那 4 bytes 的長度欄（`0x4800`）。解錯時值域一樣是 0–15，**是顏色分布抓出來的**（最多的一種佔 6% vs 正確的 22%）（`docs/re/23` §9）|
| `Enc` 是七項指令裡最長的一支，還沒解 | 指令表指的是**中途入口** `0x11CE7`，`docs/re/51` 記的是函式起點 `sub_11CD0`——同一支函式在筆記裡長出兩個身分 | `Enc` 就是遭遇驅動器的手動入口，`docs/re/51` 整份講的就是它。新的只有「不在這張地圖的隊伍要先問一句 Y／N」（`docs/re/94`）|
| 主選單有「新遊戲／讀檔」兩條路 | 工項名稱是照現代遊戲的習慣寫的，沒去讀 `sub_1630C` 的兩張表 | 標籤字串整個只有 `Start`，兩支處理程式裡第二支是一條 `retn`。存檔就是 `GAME1`／`GAME2` 本身，角色增刪在 Ranger Center（`docs/re/95`）|

## 7. Worklist

> **remake 的工項清單搬到 [`WORKLIST.md`](WORKLIST.md)**（含已完成的，逐項標狀態與證據）。
> 這一節只留 RE 側的完成度與還沒解的逆向。

**RE 完成度**：資料格式、文字與資產層打通；規則層（戰鬥判定、回合結構、指令與逃跑、
遭遇生成與敵人血量、屬性、效果、商店、升級、檢定與技能成長）與世界互動層
（移動、時鐘、事件分派、44 個腳本指令、段落引用、遭遇掃描）都已解。
641 個函式裡人寫的筆記涵蓋 435 個（206 個沒碰過，多半是無呼叫端的死碼與 C 執行期）。
完整缺口見 [`docs/re/00-remake-knowledge-gaps.md`](docs/re/00-remake-knowledge-gaps.md)。

**Remake 進度**：**二十六份規格全部 READY 並實作完成**（資產、亂數、畫面與文字、移動與時鐘、
角色與存檔、戰鬥、世界事件、設施、中文排版、翻譯管線、回合結構、遭遇生成、戰鬥指令、
遭遇掃描、戰鬥畫面、指令處理程式、設施互動迴圈、段落手札、輸入層、音效、遭遇迴圈、回合結算、設施場景、模式路由、設施選單、圖片動畫），
`cmd/wasteland -mode play` 從**標題畫面**開場（按 `S` 進遊戲，`docs/re/95`），
可以走地圖、用指令列七項（`USE`／`ENC`／`ORDER`／`DISBAND`／`VIEW`／`SAVE`／`RADIO` 全部接上）、
遇敵進戰鬥（名單畫面下指令、逐回合打完）、踩進設施買賣、治療、學技能、在 Ranger Center 建角色（名字可打中文）、存檔，設施圖上的局部動畫會動。
結局圖也解得出來（`wl-shot -mode end`），還沒接進流程。
**中文化的文本工作已全部完成**：4,806 條可翻字串 ＋ 段落書 162 段。

### 7.1 下一輪要做的（照 §0 的節奏：讀 RE → 寫 spec → 實作）

1. ~~翻譯本體~~ —— **已完成（4,806／4,806 可翻條目）**。
   原文 4,827 個槽扣掉 21 個**不可翻**的槽（未用槽的解碼雜訊與純控制碼，
   `docs/re/17` §1）＝ 4,806 條可翻條目，全部翻完。
   清單在 `translations/untranslatable.tsv`，`tools/build_lang.py` 會擋
   「不該有譯文的 key 被翻了」，`tools/untranslated.py` 也扣掉它們——
   **這個數字有結論，不留永遠減不掉的餘數。**
   共用文本層 `_shared.tsv` 的 185 條翻譯涵蓋 1,425 個 key。

2. ~~段落書的 162 段中文翻譯~~ —— **已完成（162／162）**。
   譯稿在 `translations/zh-Hant/paragraphs/*.md`，`tools/build_paragraphs.py`
   編成 `translations/paragraphs-zh-Hant.cat`，`internal/play.Journal` 讀它；
   測試逐一檢查 1–162 都查得到正文，缺一段就紅。
   **這批不算在 4,827 條裡**——段落書是紙本，不在原版的字串語料裡。

### 7.2 還沒解的逆向

**都不擋實作**：對應的機制都有 READY 規格，缺的是「為什麼」不是「怎麼做」。
完整檢查表在 [`docs/re/00-remake-knowledge-gaps.md`](docs/re/00-remake-knowledge-gaps.md)。

| # | 還沒解的 | 為什麼不擋 |
|---|---|---|
| A12b | `CURS` 的**消費端**：哪個圖形對應哪個狀態，以及資料的 16 寬為何與 slot 21 的 24 寬對不上 | 版面已解（8 個 32 × 16，遮罩 ＋ 圖形並排，`docs/re/57`）；遊戲主線不用滑鼠 |
| — | `0x13FC8`–`0x13FD9` 第一個受罰者才跑的欄位前置處理（`docs/re/69` §6）；`sub_1790B` 怎麼取那個被塞進時鐘的數字（`docs/re/74` §1） | `sub_142ED` 的形狀已解（暫換時鐘的時 → 印 `+0x03` → 延遲 → 還原），remake 刻意只印訊息 |
| — | 17 種腳本 opcode 還沒實作（`docs/re/76`）：0、1、2、4、7、9、14、32、34、35、36、37、39 與四個非法值 | **有格子指到的一個都不缺**；剩下的都要靠改寫才到得了，覆蓋率有測試守著（`missCells != 0` 就紅） |
| — | 隊伍攻擊那條的命中基礎值查表 `ds:711Dh`（索引 `ds:CF86h + 攻擊者×4`，回 `0xFF` → 50 否則 60，`docs/re/88` §5） | 兩個候選值只差 10 點，remake 用 60；敵方那條已完全解 |
| A13 | `TRANSTBL` 的**用途**（形狀已解：50 組 × 16 的索引對照表） | 三層掃描都找不到消費端，與資源 idx 7 同一種遺留（`docs/re/56`） |
| — | 物品 `+0x03` 低 3 位、敵人 `+0x04` 高 4 位**都沒有讀取端** | 資料裡有值但程式沒讀；原樣 round-trip，不給語意 |
| — | 物品 70／71／72 原本是什麼（`docs/re/50`） | 名字是被清空的、資料完整、字母序把開頭夾在 H–M。**這份 DOS 版問不出更多**；要答案得看別的平台版本 |
