# CONTEXT — 專案脈絡與文件索引

> 這份是全專案的單一入口。對話被壓縮、或換一個新 session 接手時先讀這份，
> 再依索引跳到需要的文件。工作紀律與硬規則在 [`CLAUDE.md`](./CLAUDE.md)。
> **逆向結果的速查表在 [`docs/re/00-master-index.md`](docs/re/00-master-index.md)。**
>
> 最後更新：2026-08-17

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
| 英文手冊 | 全文轉 markdown ＋ **中英對照**，7 章（`docs/manual/`）。技能與武器譯名對過遊戲內字串表 |
| 攻略 | **自建**（`docs/walkthrough/`，八章 ＋ `generated/` 四份機器產出的表）。資料由 `cmd/wl-atlas` 從 `game1`／`game2` 倒出，不是翻譯來的 |
| 段落書 | 162 段全部轉錄，編號連續無缺（`docs/paragraphs/`）。**三層防拷結構已辨識**：3 個陷阱段落（1／22／145）、64 段變體組（同場景不同密語）、33 段火星誘餌假劇情 |
| **結局的觸發點** | **已解**：資料裡沒有跳表第 4 格是設計不是漏掃——索引 4 由主迴圈的 `sub_1CB30`（`0x16C28`）在科奇斯基地自毀倒數 240 刻到期時**自己合成**（`al ← 84h`）。倒數由腳本 opcode 35 啟動，全 42 張地圖只有資源 20 記錄 4 用它（`docs/re/100`）。`TestCochiseEndgame` 從四把鑰匙走到結局 |
| **功能鍵與背景音樂** | **完成**（規格 27，**原版都沒有**）：F1 說明、F2 設定、F5／F9 快速存讀檔、F10 離開（先問 → 先存 → 才退）、ESC 一律只取消。快速存檔走獨立的 `WLQS` 檔，不碰玩家的原版資料。配樂**十首**，由 `tools/make_music.py` 譜、Roland MT-32／CM-32L 算成 ogg，跟著場景與晝夜切換（`internal/play/music.go`）；**ROM 與 ogg 都不入版控** |
| **畫面解析度與字型** | 畫布 **960×600**（原版 3×）；中文倚天 **24×24**（`STD.24M` 用 `tools/etunpack.py` 解開）、英數倚天半形 **`ASCFONT.24`**（16×24）。**英數不拿 8×8 放大**——放大後筆劃比中文粗，混排會露餡（規格 10 §2）|
| **滑鼠** | 規格 29。游標用原版 `CURS`（`internal/assets/curs.go`）並**跟著狀態換圖形**（`docs/re/112`：0 預設／1 可點／2–5 上下左右／6 中央方框，第 7 個原版選不到）。**控制走 remake 自己的路**：點到哪一格就送那一格的字元、點地圖往那個方向走、右鍵取消。點擊一律翻成與鍵盤等價的輸入再走同一條 `Update` |
| **片頭** | `docs/re/113`。標題畫面按非 `S` 的鍵就放原版的六頁開場字幕（第 0 張字串表，每頁 255 個計時器刻），播完循環；按 `S` 照樣開始遊戲 |
| **戰鬥指令 `H` 雇用** | `docs/re/110`、`114`、規格 17 §4。遭遇記錄 `+0x09` 的 bit1 ＝ **不敵對**、高 4 位 ＝ section 17 的 NPC 記錄編號 → 整筆 256 bytes 抄進隊伍 → 魅力對決 → 7 人上限。**出貨資料裡有 14 個雇得到的人**（FELICIA、ACE、JACKIE、CHRISTINA…），在 section 3；友善的那一組不會攻擊你，而**你一開槍就翻臉** |
| **攻略斷言對照** | `docs/walkthrough/swm-005/mechanics-claims.md`：軟體世界攻略的 31 條機制斷言逐條拿逆向驗過（**相符 25、有出入 5、未定 1**）。未定的那一條（M-10）缺的只有「攻略說的『廢坑』是哪一張圖」|
| **敵人的名字** | `docs/re/114` §6。原版印的是**這張地圖自己的明文名字表**（`Juveniles`、`Woman`、`City Slicker`），不是種類名；遭遇記錄 `+0x09` 的 bit0 決定走哪一條。328 條全部中文化（`translations/zh-Hant/monsters.tsv`）|
| **肖像框** | `docs/re/115`。畫面左上那一塊是**一張圖 ＋ 一行 12 格置中的說明**，兩者都是跨畫面模式活著的全域（設施畫面用同一組）。戰鬥每回合挑一次：有敵人就是那一組的 `+0x07` 與名字，一組都挑不到就是遊俠（圖 8、`Ranger`）|
| **設施畫面** | `docs/re/117` §2、`docs/re/118`。進店就切成**名單模式**：左邊肖像框（店主的圖 ＋ 招牌）、右上選單區（欄 15–38）、底下隊伍名單、指令列照留。流程是「招呼語 → 誰要進去？ → 買／賣」，`P` 是**集中金錢**（把其他隊員的錢全搬給櫃檯前這個人），清單一頁九行、滿了寫 `MORE!`。**每家店有自己的庫存表**（記錄 `+0x06` 選四份之一，四份只有庫存那一欄不同）|
| **敵人在地圖上移動** | `docs/re/116`。每回合算一次計畫（換位置／逃跑／朝你衝／不動），執行時**把遭遇那一格搬過去**（舊格清成 nibble 0）並印一句話。有射程武器的看「這一格有多不划算」決定要不要動，近戰的看距離與士氣。接上之後命中基礎值 50／60 兩條都會走到 |
| **專有名詞與地名的體例** | **中文（英文）**，英文放全形括號裡（使用者定案 2026-08-18）——玩家要對得回原版畫面、官方手冊與當年的攻略。普通名詞不加括號。`build_lang.py` 對 `place:`／`monster:` 把格數上限放寬成一行 38 格 |
| **全隊倒下** | 規格 28。三分支（什麼都不做／自動換隊／死亡畫面）；死亡畫面的圖與兩段文字從映像直讀 |
| **交付物打包** | `tools/dist.sh` → `dist-all/`：`release/`（可散布，附 `setup.sh` 讓玩家從自己的原版產合成映像）、`local/`（本機完整包，含原版資料與音樂）、`promo/`（推廣片）。腳本會檢查 release 沒混進不可散布的檔 |

### 進行中／未開始

| 項目 | 狀態 |
|---|---|
| 解包映像實跑驗證 | **已做**。解包版在 headless DOSBox 跑起來，畫面與原版 `compare -metric AE` ＝ **0**（`docs/re/47` §3） |
| 資料格式 | 已解：`wla.bin`（overlay 程式碼）、`title.pic`（XOR 串流）、`colorf.fnt`（172 字 × 32 bytes，格式與用途都已驗證）、主文字字型（內嵌）、`GAME1`／`GAME2` 的定址方式。`allpics1/2` 的 82 張圖（`docs/re/23`）、`allhtds1/2` 的 9 組圖磚（`docs/re/24`）。`ic0_9.wlf`／`masks.wlf` 的格式與用途（`docs/re/24` §2.3）。未解：`allpics*` 交錯的參數區、`transtbl`、`curs`、`end.cpa` |
| 劇情敘述文字 | **已解**：執行檔九張打包表 426 條 ＋ 地圖區塊 4,401 條，合計 4,827 條非空（`docs/re/17`、`18`）。**與段落書的分工也解了**：`Read paragraph N.` 就是敘述字串的一部分（83 處、82 個編號），沒有段落機制；陷阱段落 1／22／145 零引用（`docs/re/33`）|
| MSQ 尾段 | 已解：無 magic 的 Huffman 流，42/42 解出 4,096 或 1,024 bytes ＝ 地圖第 3 層（每格 1 byte，`docs/re/24`） |
| **Huffman 解壓** | **已實作並驗證**（`tools/huffman.py`）：`allhtds1/2`、`allpics1/2`、`end.cpa` 共 173 個子區塊全部解出，長度精確吻合、檔案 100% 用完（`docs/re/11`） |
| 載入器分工 | 已解：`DL`＝0 ALLPICS、1 GAME／存檔、2 ALLHTDS、6 END.CPA，各有位移表 |
| 說明書整理 | **四份全部完成**：英文手冊（中英對照）、段落書、軟體世界中文說明書逐頁轉錄、攻略（自建） |
| **規格（G2）** | **二十六份全部 READY**，沒有未寫的規格了（`docs/spec/00-index.md`）|
| **`internal/assets`** | **已實作並通過驗收**：SHA-256 驗證、資源定址、MSQ 解密、Huffman、5-bit 文字、兩套字型、圖片／圖磚／地圖三層。9 個測試全綠，含 `Raw` 的 byte-for-byte round-trip（`tools/go.sh test ./...`）|
| **`internal/textlayout`** | **已實作**：18 個控制碼含**變形機制與巢狀**（單複數／性別／三選一／數量）、組行與分頁。4,889 條語料全部排得過，未解碼只剩 `0x08` 的 7 次 |
| **`internal/render`** | **已實作**：合成 320 × 200 索引畫面。地圖視窗逐像素驗過剛好 288 × 128 @ (8,8)、捲動一格 ＝ 位移 16 像素 |
| **`internal/game/rng`** | **已實作**：進位鏈與四支擲骰 ＋ 5d6 取三。驗收數列（前七項 ＝ 二項式係數）、分佈、300 萬次不重複全過 |
| **翻譯目錄** | **管線已通**：`tools/extract_strings.py` 抽出 **4,827 條原文**（426 ＋ 4,401，與 `docs/re/17`／`18` 吻合）→ 人翻的 UTF-8 TSV → `tools/build_lang.py` 編成 **UTF-8 `.cat`**（v2）。**編譯時擋五種錯**：格數超過原文、控制碼數量不符、Big5 缺字（倚天畫不出來的字，這關沒放鬆）、`\x10` 後的按鍵被翻掉、不可翻的槽被翻了。Go 只讀 `.cat`，**整條管線 UTF-8，只在畫字那一刻轉 Big5**（`render.DrawRune`，唯一一處；`docs/utf8-experiment.md`）。端到端驗過：踩到有翻譯的格子畫面上出現中文。**已全部翻完：5,243／5,243 可翻條目**（原版語料 4,827 槽扣掉 21 個不可翻的槽——未用槽的解碼雜訊與純控制碼，清單在 `translations/untranslatable.tsv`，build 會擋「不該有譯文的 key 被翻了」）。**另外三類不在字串表裡**：`ui:` 介面文字、`place:` 店家招牌、`monster:` 敵人名字（328 條，`tools/extract_monster_names.py` 抽出來，`docs/re/114` §6）。共用文本層 `_shared.tsv` 的 185 條涵蓋 1,425 個 key |
| **中文排版** | **已決策並實作**：內部畫布 **960 × 600**（原解析乾淨 3×），原版素材 nearest 放大、倚天直繪不縮字（`rulebook/81`）。一個中文字剛好佔原版一個字元格，所以 `docs/re/25` 的座標完全不用重算、訊息視窗仍是 6 行 × 38 格。字型優先 24 點（`STDFONT.24` ＋ `SPCFONT.24` ＋ 半形 `ASCFONT.24`），只有 15 點時照原尺寸畫在格子中央；Big5 分區索引已過 oracle（「一」是一條橫線、「中」「猴」可辨識、全形標點不落 fallback）。**字型檔玩家自備，沒有時遊戲照跑英文**（`docs/spec/10`）|
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
| **地圖指令列** | **已解入口**：底部固定七項 `Use Enc Order Disband View Save Radio`，處理程式在 `ds:AB1Ch`（`sub_16C7C` 設 `ds:468Dh`／`ds:468Fh`，`ds:4689h ＝ 6` ＝ 最大索引）。**升級的入口是 RADIO**（`0x15260` → `loc_1B8AD`）——問一句、答 Y 就逐人檢查經驗值並升級，可連升、播音效 4。**七項全部接上**，`Save` 與 `Radio` 各帶一次 Y／N 確認（`docs/re/91`、`docs/re/92`、`docs/re/93`、`docs/re/94`）|
| **音效** | **已接**：`internal/audio` 的位元組碼直譯器 → `Synth`（方波，照 8253 除數算，不換算成浮點）→ Ebiten。走一步播 1、腳本 op 播 7、升級播 4（`docs/re/44` §6）|
| **行動順序** | **已解**：兩邊公式不同——敵人 ＝ 2d6 ＋ 敵人資料 `+0x02` **× 8**，隊伍 ＝ 2d6 ＋ Speed（角色記錄 `+0x11`）＋ Brawling × 3（**沒有 ×8**）。2d6 逢同點續擲。而且**只有下攻擊令的人才排進行動表**（`0x1AE78` 的 `cmp al, 2`）——迴避、換武器、使用物品的人這一回合連格子都不占（`docs/re/90`）|
| **敵人打誰** | **已解**：`roll(1..隊伍人數)` 隨機挑，挑到 CON ≤ 0 的人就**整個重抽**（`0x1B054`）——沒有仇恨、前排或輪流。重抽沒有上限，靠 `sub_19D0E`（還有沒有人能打）在進來之前擋住。⚠ `sub_172BB`（CON ≤ 0 ＝ 倒下）與 `sub_172AE`（CON ＝ 0 ＝ 死）是兩個判準，戰鬥全部走前者（`docs/re/89`）|
| **戰鬥指令與逃跑** | **已解**：一回合開頭逐人下令，指令碼寫進 `ds:46D8h`、參數寫進 `ds:46DAh`。**指令碼是熱鍵字母表 `' ' H A W R E L U` 的索引，不是選單的顯示順序**（選單印 Run/Use/Hire/Evade/Attack/Weapon/Load）。`ds:A43Bh` 是 10 筆處理程式跳表；**迴避的處理程式是空的**，效果全在命中門檻的基礎值（迴避 60 > 攻擊 50 > 其他 40）。**逃跑不擲骰也沒有失敗分支**：問 Party／Single，選一個方向就走。CON ≤ 0 的人不下令（`docs/re/38`、`docs/spec/14`）|
| **四支指令處理程式** | **已解**：Hire／Weapon／Load／Use 加上已解的迴避、攻擊、逃跑，**七支全部是同一個形狀**——「檢查前提 → 不成立就印一句話重問；成立就開選單、回傳選擇」。**它們都不執行動作**（換武器沒真的換、裝填沒真的填），動作在結算階段依指令碼與參數進行。物品陣列的數與找都認記錄 `+0xBD` 起 30 槽、stride 2。Use 是唯一會寫狀態的：把一個 byte 記在**以角色編號為索引**的一格（`ds:A9FDh`）（`docs/re/41`、`docs/spec/17`）|
| **設施的互動迴圈** | **已解**：商店主迴圈（選人 → Buy／Sell → 清單 → 換人／離開）；賣一件會**把那個槽清成 0、加錢、店家庫存 +1**，清單會標記裝備中的但不擋著不讓賣。**物品表 `+0x02` 是庫存量不是旗標**（0 ＝ 缺貨不列出、`0xFF` ＝ 無限、其餘賣一件 +1）——只看讀取端會以為是旗標，找到 `add al, 1` 才知道。醫生的逐點治療**每一輪重算 MAXCON − CON**，一次按鍵只治一點只扣一次錢，錢不夠就停在中途（`docs/re/42`、`docs/spec/18`）|
| **段落手札** | **已解＋已定案**：原版沒有段落機制，`Read paragraph N.` 就是一條普通敘述（`docs/re/33`）。引用表由 `tools/extract_paragraph_refs.py` 從**英文原文**抽成 `docs/re/generated/paragraph-refs.tsv`（83 條引用、82 個編號）——**不在執行期解析翻譯過的文字**，否則譯者的用字會決定遊戲讀不讀得到段落。手札**預設全開**（忠於紙本），分成正文（會被引用的 82 段）與附錄（其餘 80 段，含陷阱段 1／22／145），附錄標明是 1988 年防拷設計；**一段都不刪**（`docs/spec/19`）。**正文自成一個目錄檔**：譯稿 markdown（`translations/zh-Hant/paragraphs/`）→ `tools/build_paragraphs.py` → `translations/paragraphs-zh-Hant.cat`（key `para:<編號>`，與翻譯目錄同格式所以 Go 不必多一份讀檔器）→ `internal/play.Journal`。段落書是紙本、不在原版語料裡，所以**不算在 4,827 條**；譯稿 162／162 已完成|
| **輸入層與選項熱鍵** | **已解**：滑鼠不產生事件，只把游標位置換成一個按鍵碼再走鍵盤路徑（`sub_18EFE` 先問滑鼠、問不到才 `int 16h`）。21 筆熱區表 `ds:0CAEBh`，**第一個「遮罩有開且座標命中」的區域就決定結果**；遮罩是每個畫面在等待輸入前設的 32 位元字（`ds:7DF3h`／`7DF5h`）。**控制碼 `\x10` 的用途是滑鼠**：把緊接的字元登記到「每列一格」的熱鍵表 `ds:8DDCh`，點哪一列就送出哪一格。所以 **`\x10` 後面那個字母翻掉會同時弄壞滑鼠（送出 Big5 首位元組）與鍵盤（比對仍是原字母，畫面上卻沒有提示）**——寫法固定為 `\x10Y 是`，`tools/build_lang.py` 會擋（`docs/re/43`、`docs/spec/20`）|
| **遭遇怎麼冒出來** | **已解**：`sub_14664` 掃**地圖視窗那 9 × 19 格**，第 1 層 nibble ＝ 3 或 15 的就是遭遇格。取該格的地圖記錄，算與隊伍的距離：大於記錄 `+0x00`（察覺上限）整格略過；不大於 `+0x01`（主動距離）或小於該組的接戰值就進佇列。佇列在 `ds:A96Fh`，**4 隊伍組 × 4 槽 × 4 bytes（x／y／地圖／距離）**，槽內由近到遠排，**每次掃描整個清空重建**——它是視野快照不是待辦清單。掃完會把不在佇列裡的敵方記錄整筆清掉（`docs/re/39`、`docs/spec/15`）|
| **戰鬥畫面** | **已解，版面四塊都對到實機**：`ds:46B9h` 一翻，畫面上半整個換掉——**肖像**在欄 1–13／列 1–12、**指令選單與戰鬥訊息**在欄 15–38／列 1–13（`sub_19727` 設的四個邊界，不是訊息視窗，`docs/re/105` §2）、**名單**從列 14 排到列 23（**一行 39 欄**、行首序號 ＋ `>`、六個欄位在 `0x11`/`0x15`/`0x18`/`0x1C`/`0x20`，`docs/re/103`）、指令列照留。八行選單一行一個，`docs/re/40` §5「一行放不下要重排」的問題不存在。指令階段每問一個人就重畫一次名單。一回合的訊息序列（遭遇開始 → 下令 → 命中/未命中 → 傷勢 → 經驗值）逐條對到字串編號。**中文化的硬規則只剩一條：熱鍵不跟著翻譯走**（`docs/re/40`、`docs/spec/16`）。⚠ 未解：有敵人時肖像畫的是誰 |
| **存檔** | `docs/re/117` §1。**目前地圖存在全域狀態那 14 bytes 的位移 7（`ds:4655h`）**，不是只在隊伍槽表——原版讀檔只抄那 14 bytes 就載地圖。人數（`ds:4653h`）與已用掉的最大角色記錄編號（`ds:4656h`）同一塊，雇用來的隊員少寫這兩格的話原版讀檔時他不在隊伍裡 |
| **戰鬥** | **已實作**：命中判定（兩邊方向相反、夾在 100）、敵方傷害（基底 ＋ Nd6）、隊伍傷害（五項相加、沒有骰）、護甲吸收 ＝ N 顆 d6、角色 CON 可為負與五級傷勢、敵人 HP 夾在 0、擊殺經驗值。**命中累加值的四個項全部接上**：基礎 ＋ Brawling（技能 1，寫死）× 3 ＋ Agility − 對手行動值（近戰類別先做 8-bit ×4，否則另加 5）；敵方那條的基礎值取被打者這回合的指令（迴避 60／攻擊 50／其餘 40），**迴避因此才真的有效**。隊伍那條的基礎值也解了：查 `ds:711Dh` 這張**每回合重排的敵方移動計畫表**（`sub_14BF0` 先清成 `0xFF`），`0xFF` ＝ 這隻不動 → **50**，有計畫 → 60。敵人在地圖上的移動已經實作（`docs/re/116`：九個步向、換位置／逃跑／逼近三條分支、清舊格寫新格），所以 50 與 60 兩條都會走到——**計畫寫成 `NoMovePlan` 而不是 0**，因為 0 是合法的步向（`docs/re/88`、`docs/re/101`、`docs/re/116`）|
| **音效** | **已解**：跑在 `int 08h` 上的位元組碼直譯器（126.36 Hz，每 6 tick 鏈回 BIOS）。**四個聲部是優先序仲裁不是混音**——PC 喇叭一次只發一個音，編號小的贏。四個 opcode（設欄位／計數器迴圈／換寫入基底／音符），每個聲部有滑音、線性分段封套、顫音、移調。**九首全在 `seg005` 的 864 bytes 裡**，不在外部檔案。⚠ 半音表的 **E 走音 −2.91 音分**（原始資料的錯字，`0x714F` 應為 `0x711E`），重製版照抄不修。**呼叫端全檔掃過**：只有五個直接呼叫點（音效 1／2／4／5／7），0／3／6／8 沒有人播；**音效 6 是無限循環**，只有「全部靜音」關得掉（`docs/re/44`、`docs/spec/08`、`internal/audio`）|

## 3. 文件索引

| 文件 | 內容 |
|---|---|
| [`CLAUDE.md`](./CLAUDE.md) | 專案規範：四道閘門（含 G4 接線）、IDA 鐵則、文件與中文化政策、環境硬規則 |
| [`docs/re/00-master-index.md`](docs/re/00-master-index.md) | **RE 總表**：位址換算、資料格式、結構佈局、位址表、關鍵函式、工具。**查已知事實先看這份** |
| [`docs/re/00-remake-knowledge-gaps.md`](docs/re/00-remake-knowledge-gaps.md) | **RE 完成度檢查表**：remake 需要的每一項知識、狀態與入口 |
| [`docs/re/00-function-index.md`](docs/re/00-function-index.md) | 函式索引（641 個，已分析 464）。讀任何 `sub_XXXXX` 前先查 |
| [`docs/re/00-wiring-status.md`](docs/re/00-wiring-status.md) | **接線狀態**：108 份筆記的結論（已接 104、未接 0、不適用 4），remake 有沒有真的用上。`TestWiringStatus` 雙向守著（`CLAUDE.md` §0 的 G4）|
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
| [`docs/re/88-hit-accumulator.md`](docs/re/88-hit-accumulator.md) | 命中累加值 `sub_1B108` 的四個項（Brawling×3、Agility、對手行動值、基礎值）全部落地 |
| [`docs/re/89-enemy-target-and-down.md`](docs/re/89-enemy-target-and-down.md) | 敵人隨機挑目標、挑到倒下的整個重抽；「倒下」（CON ≤ 0）與「死」（CON ＝ 0）是兩個判準 |
| [`docs/re/90-party-initiative.md`](docs/re/90-party-initiative.md) | 隊伍的行動值公式（沒有 ×8），以及「只有下攻擊令的人才排進行動表」 |
| [`docs/re/91-map-command-bar.md`](docs/re/91-map-command-bar.md) | 地圖指令列的七個處理程式；**升級的入口是 RADIO** |
| [`docs/re/92-use-command.md`](docs/re/92-use-command.md) | `USE` 的第一層 Skill／Item／Attribute 三選一，與施用的骨架 |
| [`docs/re/93-order-disband-view.md`](docs/re/93-order-disband-view.md) | `ORDER`／`DISBAND`／`VIEW`：兩支要多隊伍、一支不用 |
| [`docs/re/94-enc-command.md`](docs/re/94-enc-command.md) | `ENC` 不是新指令，是自動遭遇驅動器的手動入口 |
| [`docs/re/95-main-menu.md`](docs/re/95-main-menu.md) | 主選單只有一個選項，而且**沒有「讀檔」**——存檔就是進度本身 |
| [`docs/re/96-ending.md`](docs/re/96-ending.md) | 結局掛在設施跳表 `ds:A4E0h` 的第 4 格；`END.CPA` 第二段是動畫腳本 |
| [`docs/re/97-playtest-sampling.md`](docs/re/97-playtest-sampling.md) | 抽樣試玩第一輪：七段流程各走一遍，修掉六個「編得過、測得過、玩不動」的缺口，剩下的列在 §4 |
| [`docs/re/98-a0-wiring.md`](docs/re/98-a0-wiring.md) | 補完 A0：中文的三層接線、店家庫存跟著存檔走、「全隊倒下」確認是原版行為 |
| [`docs/re/99-party-wipe.md`](docs/re/99-party-wipe.md) | 全隊陣亡的三分支；救得回來 ＝ CON ∈ [−10, −1] 且狀態位元全 0 |
| [`docs/re/100-ending-trigger.md`](docs/re/100-ending-trigger.md) | **結局的觸發點**：不在資料裡，在主迴圈的 `sub_1CB30`；自毀倒數 240 刻、opcode 35、科奇斯基地反應爐層的完整啟動序列 |
| [`docs/re/101-enemy-move-plan-table.md`](docs/re/101-enemy-move-plan-table.md) | `ds:711Dh` ＝ 敵方這一回合的**移動計畫**（每回合由 `sub_14BF0` 清成 `0xFF`）；隊伍那條的命中基礎值 50／60 從這裡來 |
| [`docs/re/102-unreachable-opcodes.md`](docs/re/102-unreachable-opcodes.md) | 走不到的 12 個腳本 opcode 逐支讀完：分隊位置、寄放隊員、倒數、批次改寫、鄰格比對、時間戳 |
| [`docs/re/103-roster-line-columns.md`](docs/re/103-roster-line-columns.md) | 名片行的 `AMM` 三道閘與 `WEAPON` 取單數形；實機截圖對出行寬 39 欄與行首序號 |
| [`docs/re/104-opcode-2-icon-swap.md`](docs/re/104-opcode-2-icon-swap.md) | 腳本 opcode 2 ＝ overlay slot 18：把兩張圖形（含遮罩）對調 |
| [`docs/re/105-enc-empty-round-and-menu-region.md`](docs/re/105-enc-empty-round-and-menu-region.md) | `ENC` 在空地上也能跑一回合（字串 `0x14`）；戰鬥的指令選單畫在欄 15–38、列 1–13 |
| [`docs/re/106-text-scroll.md`](docs/re/106-text-scroll.md) | 文字滿了會**捲動**不是切掉；捲動速度 9 段（`ds:465Bh`，`<`／`>` 調）；順帶查掉 `ds:465Bh` 不是時鐘的時 |
| [`docs/re/107-command-resolution.md`](docs/re/107-command-resolution.md) | 指令的**結算階段**跳表 `ds:A568h`：換武器、裝填、迴避各自真的做了什麼 |
| [`docs/re/108-combat-use-and-hire.md`](docs/re/108-combat-use-and-hire.md) | 戰鬥 `Use`／`Hire` 的參數格式；九向位移表與「對角那四格選不到」|
| [`docs/spec/00-index.md`](docs/spec/00-index.md) | **規格索引與閘門狀態**：哪些可以動工、其餘擋在什麼上 |
| [`docs/spec/01-assets-and-formats.md`](docs/spec/01-assets-and-formats.md) | READY：資源定址、解密、Huffman、5-bit 文字、字型、圖片、圖磚、地圖三層 ＋ Go 介面草案 |
| [`docs/spec/02-rng-and-dice.md`](docs/spec/02-rng-and-dice.md) | READY：進位鏈亂數與四支擲骰，含驗收數列 |
| [`docs/spec/03-screen-and-text.md`](docs/spec/03-screen-and-text.md) | READY：畫布、五個視窗、座標單位、控制碼、中文版面的兩條路 |
| [`docs/spec/27-remake-ui-additions.md`](docs/spec/27-remake-ui-additions.md) | READY：**原版沒有的東西**——F1／F2／F5／F9／F10、ESC 只取消、`WLQS` 快速存檔、MT-32 背景音樂。不引用 IDA 位址 |
| `internal/` | 已實作：`assets`（規格 01）、`textlayout`／`render`（規格 03）、`game/rng`（規格 02）。`tools/go.sh` 是 Go 的唯一入口，編譯與測試走 docker |
| [`docs/manual-cht/`](docs/manual-cht/) | 軟體世界 1990 中文說明書全 60 頁節轉錄 ＋ 當年譯名表 |
| [`docs/manual/`](docs/manual/) | 官方英文手冊全文 markdown，**中英對照** |
| [`docs/paragraphs/`](docs/paragraphs/) | 段落書 162 段全文與索引，含防拷結構標註 |
| [`docs/walkthrough/`](docs/walkthrough/) | **自建攻略**：八章正文 ＋ `generated/` 四份機器產出的表（地圖與傳送、條件閘、問答密語、設施）。來源是 `cmd/wl-atlas` ＋ `tools/summarize_walkthrough.py` |
| [`docs/promo-video.md`](docs/promo-video.md) | 推廣片的三段管線、視覺 token 的出處、六種版面、踩過的坑。**成品不入版控** |
| [`docs/mt32-rhythm-probe.md`](docs/mt32-rhythm-probe.md) | MT-32 節奏鍵位的量測：哪些鍵沒有指派（含 GM 的腳踏鈸 42）、怎麼從能量與過零率挑鼓 |
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

**確定的錯誤不留在這裡。** 斷言被推翻、而且正確答案已經寫進 `docs/re/*` 與程式碼之後，
這一列就刪掉——正文寫的就是現況，教訓寫在各份筆記的「這一輪學到的（寫成規則）」，
表裡再留一份只會變成第二個要維護的真相。

**留在這裡的只有一種**：斷言確定錯了，但**正確答案還沒定案**——
那張表的長度就是「已經知道錯、還不知道對」的數量。

| 斷言 | 為什麼錯 | 目前知道的 |
|---|---|---|
| 閒置不推進時間是因為「主迴圈卡在 `sub_18E90` 的忙等迴圈裡」 | 只讀了 `sub_18E90` 有忙等，沒有回頭確認地圖主迴圈是不是真的走那一支 | 地圖主迴圈讀鍵用的是**非阻塞**的 `sub_18EFE`（`0x16BC1`）。**現象仍然成立**（實機 45 秒時鐘沒動），但機制未解，線索在 `docs/re/26` §1.2 |
| MSQ 資源的前 2 bytes 是長度 | 把 `mov cx, [bx]` 當成「取剛讀進來的 header」，其實 `ds:46B0h` 是別處設好的緩衝區位址 | 那 4 bytes 是 magic `msq0`／`msq1`；長度來源仍未解（`docs/re/07` §7） |

## 7. Worklist

> **remake 的工項清單搬到 [`WORKLIST.md`](WORKLIST.md)**（含已完成的，逐項標狀態與證據）。
> 這一節只留 RE 側的完成度與還沒解的逆向。

**RE 完成度**：資料格式、文字與資產層打通；規則層（戰鬥判定、回合結構、指令與逃跑、
遭遇生成與敵人血量、屬性、效果、商店、升級、檢定與技能成長）與世界互動層
（移動、時鐘、事件分派、44 個腳本指令、段落引用、遭遇掃描）都已解。
641 個函式裡人寫的筆記涵蓋 464 個（177 個沒碰過，多半是無呼叫端的死碼與 C 執行期）。
完整缺口見 [`docs/re/00-remake-knowledge-gaps.md`](docs/re/00-remake-knowledge-gaps.md)。

**Remake 進度**：**二十六份規格全部 READY 並實作完成**（資產、亂數、畫面與文字、移動與時鐘、
角色與存檔、戰鬥、世界事件、設施、中文排版、翻譯管線、回合結構、遭遇生成、戰鬥指令、
遭遇掃描、戰鬥畫面、指令處理程式、設施互動迴圈、段落手札、輸入層、音效、遭遇迴圈、回合結算、設施場景、模式路由、設施選單、圖片動畫），
`cmd/wasteland -mode play` 從**標題畫面**開場（按 `S` 進遊戲，`docs/re/95`），
可以走地圖、用指令列七項（`USE`／`ENC`／`ORDER`／`DISBAND`／`VIEW`／`SAVE`／`RADIO` 全部接上）、
遇敵進戰鬥（名單畫面下指令、逐回合打完）、踩進設施買賣、治療、學技能、在 Ranger Center 建角色（名字可打中文）、存檔，設施圖上的局部動畫會動。
**結局玩得到**：科奇斯基地反應爐層的四根圓柱（四把鑰匙）→ 按鈕 → 紅黃綠藍四站 →
自毀倒數 240 刻 → 結局。觸發點不在資料裡而在主迴圈（`docs/re/100`）。
**中文化完成**：目錄 **5,243／5,243** 條 ＋ 段落書 162 段，**而且畫面上真的是中文**——
戰鬥全程、設施選單與清單、角色管理、指令列訊息都走目錄（`docs/re/98` §2）。
那 5,243 條 ＝ 原版語料 4,827 條扣掉 21 個不可翻的槽（4,806）
＋ 重製版自己的 `ui:` 與店家招牌 `place:` 共 109 條
＋ **明文敵人名字 `monster:` 328 條**（`docs/re/114` §6）。**`tools/build_lang.py` 的輸出是權威數字**——
`README.md`、`tools/promo/make_promo.sh`、`tools/dist.sh` 裡的字面值都要跟著它改。

**抽樣試玩第一輪已做**（`docs/re/97`）：七段流程各走一遍，
建角色、遇敵打完、商店買賣、`USE` 開閘、讀段落、存檔重開都通了。
那一輪修掉六個「編得過、測得過、玩不動」的缺口——物品欄讀到 0 就停、
折價指數 0 變全免、角色管理叫不出來、`Save` 沒寫檔、設施進場沒選單、
清單印編號不印名字。接著把它列的 A0 五項補完（`docs/re/98`）：
中文接線、店家庫存持久化、病名，加上確認「全隊倒下遊戲照走」是原版行為。
結局的觸發點也解了（`docs/re/100`）：跳表第 4 格在 42 張地圖裡零筆是**設計**——
索引 4 由 `sub_1CB30` 在自毀倒數到期時合成。連帶補上三段接線：
`USE` 的收尾改寫、nibble 8 問答的呈現層、腳本 opcode 35。
**`TestCochiseEndgame` 從四把鑰匙走到結局，這個遊戲玩得完了。**

### 7.1 下一輪要做的

> **交接用的 TODO 在 [`WORKLIST.md`](WORKLIST.md) 的「TODO：下一個 session
> 從這裡接」一節**，順序在該節末尾的「下一步的順序」表。
> T1–T39 都已完成（T13–T29 在 2026-08-17、T30–T39 在 2026-08-18），
> **現在沒有一項擋著把遊戲玩完**。換 session 先讀那一份。

下面兩條是已經結案的紀錄，留著是因為它們是驗收數字的來源。


1. ~~翻譯本體~~ —— **已完成（5,243／5,243 可翻條目，`build_lang.py` 的輸出是權威數字）**。
   原版語料 4,827 個槽扣掉 21 個**不可翻**的槽（未用槽的解碼雜訊與純控制碼，
   `docs/re/17` §1）＝ 4,806 條，加上重製版自己的 `ui:` 與 `places:` 共 109 條
   ＋ 328 條明文敵人名字 ＝ 目錄裡的 5,243 條，全部翻完。
   清單在 `translations/untranslatable.tsv`，`tools/build_lang.py` 會擋
   「不該有譯文的 key 被翻了」，`tools/untranslated.py` 也扣掉它們——
   **這個數字有結論，不留永遠減不掉的餘數。**
   共用文本層 `_shared.tsv` 的 154 條翻譯涵蓋 1,425 個 key。
   ⚠ **改了條數要一起改三處字面值**：`tools/promo/make_promo.sh` 兩處、
   `tools/dist.sh` 的 release README 一處。它們不會讓任何測試變紅。

2. ~~段落書的 162 段中文翻譯~~ —— **已完成（162／162）**。
   譯稿在 `translations/zh-Hant/paragraphs/*.md`，`tools/build_paragraphs.py`
   編成 `translations/paragraphs-zh-Hant.cat`，`internal/play.Journal` 讀它；
   測試逐一檢查 1–162 都查得到正文，缺一段就紅。
   **這批不算在 4,827 條裡**——段落書是紙本，不在原版的字串語料裡。

### 7.2 還沒解的逆向

**都不擋實作**：對應的機制都有 READY 規格，缺的是「為什麼」不是「怎麼做」。
完整檢查表在 [`docs/re/00-remake-knowledge-gaps.md`](docs/re/00-remake-knowledge-gaps.md)。

> ⚠ 下表是**還沒解的**。**解完了還沒接上**的是另一回事，在
> [`docs/re/00-wiring-status.md`](docs/re/00-wiring-status.md) 標「未接」：
> 條件閘扣 CON 的護甲吸收（`120`／`122`，全檔 105 筆閘扣多了）、
> 玩家攻擊的目標篩選（`123`）、輻射計量表（`120`）。
> **那三條會改變玩家感受到的規則**，順位在 `WORKLIST.md` 的「下一步的順序」。

| # | 還沒解的 | 為什麼不擋 |
|---|---|---|
| A12b | `CURS` 的資料 16 寬為何與 slot 21 的 24 寬對不上 | 版面已解（`docs/re/57`），**消費端也解完並接上了**（`docs/re/112`：`ds:8DCDh` 是索引、`0x10D4D` 是繪製常式，八個圖形對應七種狀態，第 7 個原版選不到）|
| — | `0x13FC8`–`0x13FD9` 第一個受罰者才跑的欄位前置處理（`docs/re/69` §6）| `sub_142ED` 整支已解（開捲動動畫 → 設速度 → 印 `+0x03` → 延遲 → 還原，`docs/re/74` §1、`docs/re/106`），remake 刻意只印訊息不做動畫 |
| — | section `0x10` 那 **4 個索引越界值**（5 筆記錄）是什麼意思。**44 個 opcode 本身全部解完也全部實作完**（`docs/re/102`、`104`）| 原版對它們也是拒絕，remake 照做。覆蓋率有測試守著（`missCells != 0` 就紅）。要讓那四個值消失只有一種可能——section `0x10` 的解讀改了（`docs/re/71` §5.2），不是調數字 |
| A13 | `TRANSTBL` 的**用途**（形狀已解：50 組 × 16 的索引對照表） | 三層掃描都找不到消費端，與資源 idx 7 同一種遺留（`docs/re/56`） |
| — | 物品 `+0x03` 低 3 位、敵人 `+0x04` **高** 4 位都沒有讀取端 | 資料裡有值但程式沒讀；原樣 round-trip，不給語意。⚠ 敵人 `+0x04` 的**低** 4 位有兩個用途（護甲骰數與經驗值倍數，`docs/re/37` §3.3），不要混起來 |
| — | 物品 70／71／72 原本是什麼（`docs/re/50`） | 名字是被清空的、資料完整、字母序把開頭夾在 H–M。**這份 DOS 版問不出更多**；要答案得看別的平台版本 |
| — | 遭遇記錄 `+0x09` 還有 **bit3 與 bit4–7 以外的位元**沒有讀取端 | bit0／bit1／bit2 與高 4 位都解完了（`docs/re/114` §2），剩下的在資料裡沒有出現過 |
| — | `Save` 寫回兩份存檔的哪一份、32-bit 序號怎麼推進（`docs/re/97` §3.4）| remake 寫回讀進來的那一份，round-trip 與實機讀取都驗過（`docs/re/49`）|
