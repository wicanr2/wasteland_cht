# Wasteland（1988, Interplay／EA）荒野遊俠 — 逆向、重製與繁體中文化

把 1988 年的 DOS 版《Wasteland》完整逆向，在 Go / Ebiten 上重寫引擎，再做繁體中文化。
定位是**文化資產保存**：留下一份華人玩家能玩完、能讀懂、也能考證的版本，
連同 1990 年代台灣代理商留下的中文說明書一起保存。

方法論沿用《冬之魔》(`~/cht/daemon_winter/`) 那一套並收緊一處：
**這個專案的逆向一律以 IDA Pro 為唯一主線，而且 RE 沒確認完就不准動工。**

---

## 0. 最高紀律：RE 未確認，不寫任何一行引擎程式

這是本專案的第一條硬規則，凌駕進度與交付順序。任何機制（移動、戰鬥、技能判定、
遭遇、存檔、地圖載入、文字顯示、隨機數）在 IDA 裡讀出來並寫成筆記之前，
**不得憑手冊、攻略、社群 wiki、其他重製專案或直覺實作**。

### 三道閘門，逐道通過

| 閘門 | 判準 | 產出 |
|---|---|---|
| **G1 — RE 確認** | 該機制的程式碼在 IDA 裡定位到、讀完、寫成筆記，含**輸入檔 SHA-256 + IDA 線性位址 + 函式名** | `docs/re/NN-*.md` |
| **G2 — 規格收攏** | 筆記收攏成可實作的規格，每條敘述標推論等級，未解處明寫「未解」不留空白 | `docs/spec/*.md` 標 **READY** |
| **G3 — 實作** | 只有 READY 的規格可以動手；實作完成要對原版行為驗收，不是只跑單元測試 | Go 程式碼 + 驗收證據 |

- **`docs/spec/` 沒有 READY 標記 = 不准開 `internal/game/`。** 這一條不因為「先寫個雛形」放寬。
- 卡住不是動工的理由。撞牆時換方法（靜態溯源 → xref → IDC 掃描 → DOSBox 實跑對拍），
  在筆記記下「卡在哪、試過什麼、下一個入口是哪個位址」，不要寫「暫緩」或「先照手冊做」。
- 例外只有一種：**RE 真的定位不到，而手冊／實機有明確證據**。這時要在規格裡明寫
  「依手冊暫代，RE 未定位」，並列入待補清單；不得在正文寫成已確認。

### 推論等級（每條斷言都要帶）

`已確認`（讀到程式碼並解釋得通）／`強證據`（多處互相印證但未直讀）／
`假說`（合理但只有單一來源）／`未解`。降級比升級容易犯錯：
**新結論與既有已確認常數衝突時，先假設新結論是錯的。**

---

## 1. IDA Pro 9.4 環境與鐵則

### 1.1 工具鏈（IDAPython 走得通，做法對齊 `~/cht/civ1/`）

```
IDA 安裝與工具包：/home/anr2/ida_94_official/dist
本專案使用 image：ida-pro-9.4-idapython:py312-v1
```

這個 image 的 IDAPython **實測可用**（IDA 9.4 + Python 3.12.3，`/opt/venv/bin/python3`），
所以本專案的匯出腳本一律寫 Python，不寫 IDC。

- 「那個 image 的 IDAPython 不能用、只能寫 IDC」是**別的專案**在別的 image 上的結論，
  不是本專案鐵則；以本專案的實測為準，換 image 或換版本時重測一次再下結論。
- image 缺件時，修正可重現的建置來源或建立**明確版本**的新 revision
  （做法見 `~/cht/civ1/docker/ida-pro-9.4-civ1.Dockerfile`：裝 `libpython3.12` 再用
  `idapyswitch --force-path` 把 IDAPython 的 ABI 釘在 image 內的 so），
  **不得臨時掛主機 Python、主機套件或未鎖版 library 讓它「剛好能跑」**，
  也不要因為一次啟動失敗就另建一個功能重複的 IDA image。

所有 IDA 操作走 `tools/ida.sh`，邊界寫在腳本本體（`--rm`、`--network none`、限 CPU／記憶體／pids、
工作區唯讀掛載、只有輸出目錄可寫、退出前 `chown` 回使用者）：

```bash
tools/ida.sh build                        # 建 canonical .i64 + .asm
tools/ida.sh run tools/ida/export_x.py docs/re/generated/ida94/x.json
WL_IDA_TARGET=/abs/path/other.exe tools/ida.sh build   # 換分析目標
```

### 1.2 證據流：export（IDA 內）→ summarize（純 Python）

兩段式，中間隔一層 JSON：

| 階段 | 腳本 | 產出 | 規則 |
|---|---|---|---|
| 匯出 | `tools/ida/export_*.py` | `docs/re/generated/ida94/*.json` | 只讀資料庫，開頭驗輸入 SHA-256，不在已知清單就拒絕產出 |
| 整理 | `tools/summarize_*.py` | `docs/re/generated/ida94/*.md` | 純 stdlib，不需要 IDA，**只排序整理不加語意推論** |
| 分析 | 人寫 `docs/re/NN-*.md` | 結論與推論等級 | 引用上面兩層的檔案與位址 |

工具輸出與人的判斷分兩層，是為了讓「這句話是機器讀出來的，還是我推的」永遠分得開。
每份 `docs/re/NN-*.md` 都要附**可重跑的完整指令**。

每筆證據至少帶：輸入檔 SHA-256、IDA 線性位址、`segment:offset`、原始 bytes、
未過濾的反組譯、原始 IDA 名稱（`sub_xxx`／`word_xxx`）、推論等級。

### 1.3 非破壞性與原始身分保存

- **canonical `.i64` 不可被寫回。** `tools/ida.sh run` 會先把資料庫複製到容器暫存層再跑腳本。
- **不得把 `sub_xxx`／`word_xxx` 改名成推測語意。** 名稱與位址是可稽核身分；
  語意用「附加欄位」的索引表表達，顯示時排在組語旁邊，不取代原始 operand。
- 要加註只能在明確命名的副本上，且每條註記保留原始名稱、位址與推論等級。

### 1.4 踩過的坑（照抄，不要重踩）

- **headless 的 `print` 不進 stdout**：結果一律寫檔，並且**驗證輸出檔存在且非空**——
  `exit 0` 不代表腳本真的產出了東西（`tools/ida.sh run` 已內建這道檢查）。
- **⚠ `get_operand_value()` 會把 16-bit 立即數符號擴展**：`0x91C5` 回傳
  `0xFFFFFFFFFFFF91C5`。拿去算位址會一筆都對不上，而**症狀是安靜的零命中**，
  和「真的沒人引用」長得一模一樣。16-bit 專案一律 `& 0xFFFF`（`docs/re/03` §7）。
- **`export_range_refs.py` 的範圍是半開區間 `[lo, hi)`**：查單一位址要寫
  `0xA5C5 0xA5C6`，寫成 `0xA5C5 0xA5C5` 會回零命中。腳本已加 guard 拒絕跑，
  但同一類邊界錯誤在別的工具還會再出現——**零命中之前先拿一個已知會命中的
  位址做正對照**（`docs/re/74` §3）。
- **任何過濾閾值都會製造假零**：ASCII 掃描的最短長度設 4，就會漏掉 `CURS`、`info`
  這種四字元檔名。下「沒有」的結論前先確認過濾器本身沒有洞。
- **偶發：資料庫載入異常會讓 `retrieve_input_file_sha256()` 回 `None`**，
  batch log 裡的 root filename 會是亂碼。**重跑就好，不要因此把雜湊驗證拿掉**——
  先用另一支已知可用的腳本跑同一份 `.i64` 做正對照，確認是偶發還是資料庫壞了。
- **IDA 9.4 有 API 改名**：`ida_idaapi.get_kernel_version` 已不存在，
  `getseg`／`get_segm_name`／`get_segm_class` 會噴 DeprecationWarning。
  遇到 `AttributeError` 先確認 9.4 的新名稱，不要假設舊教學還成立。
- **崩掉的腳本會弄壞 `.i64`**，之後所有指令回「Failed to initialize IDA as library」——
  這訊息像授權失效，其實只是那份資料庫壞了。判斷方法：拿另一個 `.i64` 跑同一支已知可用的腳本。
- **不要 grep `.asm` 找位址。** 16-bit 的反組譯文字顯示 `segment:offset`，符號名用的是線性位址，
  整份 `.asm` 連一個 5 位十六進位常數都不會有。零命中與「真的沒人碰」長得一模一樣。
  `.asm` 只能拿來導航，不能取代 xref。
- **問「這個變數是誰在寫」要查 xref 圖並看 xref type**，不要自己解析指令文字
  （`push` 的第 0 個運算元是來源不是目的，會把三十個 `push` 判成寫）。
  範圍型結構逐 byte 查 `get_first_dref_to`／`get_next_dref_to`。
- **xref 圖看不到把位址當純數字算的程式碼**。看到「大量 reader 幾乎沒有 writer」，
  要追取址點與間接寫入，不能因此宣稱欄位唯讀。
- **不要用 `grep -v` 過濾組語**。被濾掉的 `mov`／`add`／`shl` 常常正是索引計算與 stride。
  要縮短輸出就用位址範圍或函式邊界裁切。
- **動手之前先查 `docs/re/00-master-index.md`（RE 總表）。** 位址換算、資料格式、
  結構佈局、關鍵函式、工具全部在那一份，不要翻十八份文件，也不要憑記憶。
  三份 `00-*` 的分工：**總表**＝已知的事實；`00-remake-knowledge-gaps.md`＝還缺什麼；
  `00-function-index.md`＝641 個函式誰解過。
- **讀任何 `sub_XXXXX` 之前先查函式索引**（`docs/re/00-function-index.md`）。
  筆記超過二三十份之後，靠記憶一定會重讀已經解過的函式。
- **「唯一」「只有一處」沒有全檔掃描佐證就不要寫。**
- Ghidra／objdump／capstone 只用於獨立交叉驗證或重驗原始 bytes，
  **不得把 IDA 的匯出結果換個工具重讀，冒充第二份獨立證據**。
- `.i64`／`.asm`／解包後的 binary 全部 gitignore。

DOSBox 只在「RE 讀出來的行為需要實機確認」時用（docker 化，可送鍵、可截圖），
它是 IDA 的驗證工具，不是替代品。`cycles` 要寫死，不要用 `auto`——可重現性的敵人。

---

## 2. 原版素材盤點

`Wasteland_1988.zip` → `wastland/`（20 檔，玩家自備，不入版控）：

| 檔案 | 目前認知 | 等級 |
|---|---|---|
| `wl.exe` (62 KB) | 主執行檔，16-bit DOS MZ，**Microsoft EXEPACK 打包**。分析一律用解包後的 `wl.unpacked.exe`（`docs/re/02`），打包版的位址與函式不可引用 | 已確認 |
| `wla.bin` (4.2 KB) | 未解，可能是 overlay 或常駐常式 | 未解 |
| `game1` / `game2` (159 / 172 KB) | 主資料檔（地圖、劇情、實體）。社群說法是 MSQ 區塊 + 壓縮，**未經本專案驗證** | 假說 |
| `allpics1` / `allpics2` | 圖像集 | 假說 |
| `allhtds1` / `allhtds2` | 未解（HTDS 命名意義不明） | 未解 |
| `title.pic`、`end.cpa` | 標題畫面、結局畫面 | 假說 |
| `colorf.fnt`、`ic0_9.wlf`、`masks.wlf`、`curs` | 字型、數字圖示、遮罩、游標 | 假說 |
| `transtbl` (800 B) | 未解，名稱像轉換表 | 未解 |
| `paragraphs.txt` (72 KB) | 防拷段落書，同時是劇情文本主體 | 已確認（可讀） |
| `manual.txt` (53 KB) | 官方英文手冊 | 已確認（可讀） |
| `info`、`readme.txt` | 發行說明 | 已確認 |

上表**除了標「已確認」的以外都要由 IDA 讀出載入路徑來證實**，不得憑副檔名或社群文章寫進規格。
做任何一類資產的 pipeline 之前，先把同類檔**列舉全**（`strings wl.exe | grep -i <關鍵字>`、
看載入清單），不要假設單一檔含全部。

`112-荒野遊俠.rar` → 軟體世界中文說明書掃描 33 張 JPG（1990 年台灣代理版）+ 補完計劃說明。

---

## 3. 文件與中文化政策

### 3.1 說明書要整理成完整 markdown

四份都要，一份都不省：

| 來源 | 產出 | 要求 |
|---|---|---|
| 軟體世界中文說明書（33 張掃描） | `docs/manual-cht/` | **逐頁完整轉錄**，保留 1990 年代台灣代理商的用語、譯名、行文與編排脈絡 |
| 官方英文 `manual.txt` | `docs/manual/` | 轉 markdown + 繁中翻譯，當規則的輔助 oracle |
| `paragraphs.txt` 段落書 | `docs/paragraphs/` | 整理成 markdown 並翻譯，這是劇情文本主體 |
| 社群攻略 | `docs/walkthrough/` | 繁中化，標明來源與版本 |

**中文說明書那份是保存重點。** 它是當年台灣遊戲文化的一手史料：把「荒野遊俠」這個譯名、
當年的術語選字、代理商寫給玩家的導言與體例都留下來，不要用現代用語「潤飾」掉。
與原文不一致之處照原樣轉錄，另加註記說明差異，不要直接改成正確答案。
圖說、表格、手繪示意都要交代，掃描件裡讀不清的字標 `〔字跡不清〕` 而不是猜一個填上。

### 3.2 段落書（手札）直接做進遊戲

原版把劇情段落印在紙本上、遊戲裡只給編號，是 1980 年代的防拷手段。

- **防拷機制不移植。** 重製版不要求玩家查紙本、不做編號驗證、不重現任何複製保護。
- **段落內容以遊戲內手札呈現**：讀到編號時直接把該段中文文字顯示出來，
  並保留一份可隨時翻閱的手札介面（做法參考冬之魔的遊戲內手札）。
- 段落文字走翻譯目錄，不寫死在 Go 原始碼裡。

### 3.3 譯名與文字工程

- 統一譯名表 `translations/glossary.md` 是唯一真相，中文說明書的既有譯名優先採用
  （這是保存的一部分），與原文衝突時記在表上說明取捨。
- 所有玩家看得到的字串（介面、選單、說明、劇情）都放語言資料檔，Go 只留穩定的 key。
- 中文一律點陣字（倚天 16×15），字型檔玩家自備、不隨專案散布。
- 中英混排的排版格、標點禁則、畫布尺寸等決策等 RE 讀出原版版面之後再定，不預先假設。

---

## 4. Remake 技術決策

Go / Ebiten，分層照冬之魔那套（已驗證過在無頭環境可測）：

```
internal/assets  純解碼，回傳 image.RGBA，不認識 Ebiten
internal/game    規則層，不認識畫面
internal/ui      Ebiten 呈現層（textlayout／layout 刻意不依賴 Ebiten）
cmd/wasteland    進入點
```

- 建置與測試一律走 docker（`tools/go.sh`），不污染系統環境。
- 存檔策略是**改寫不是重建**：從原始 bytes 出發只蓋已解欄位，未解區域一個 byte 都不動，
  驗收要做到讀出再寫回 byte-for-byte 相同。
- 測試存檔一律寫到 `/tmp` 或明確的測試輸出目錄，不覆蓋原版資料。
- 驗收要對原版行為，不是對自己的測試。測試全綠不等於玩得通；
  用 debug 捷徑跑完的流程要另外驗一次「正常玩家路徑」。

---

## 5. 工作紀律

- **完整性 > 投報。** 不准用「成本高、效益低」當跳過任何素材、格式或版本的理由。
  卡關就換方法並記錄，不要寫「暫緩／低投報」當結論。
- **oracle 優先序**：原版執行檔行為（IDA 讀出） > DOSBox 實跑觀察 > 官方英文手冊 >
  軟體世界中文說明書 > 社群攻略。低優先來源只能在高優先來源缺席時暫代，且要標明。
- **一手資料贏二手推論**：檔案內容 > 檔名、目錄結構、命名慣例、其他人的文章。
- **文件要寫現況，不寫「當初怎麼錯的」。** 斷言被推翻就把正文改寫成正確答案，
  推翻紀錄集中到 `CONTEXT.md` 的「已被推翻的斷言」一處，正文最多留一個指標。
  教訓寫成規則，不要寫成事件敘述。
- **每一輪收尾**：更新 markdown → 清掉被推翻的斷言 → commit → 更新 `CONTEXT.md` 現況。
- 未解的位元組要能原樣 round-trip。不要為了讓 worklist 看起來完整而發明遊戲規則。

`CONTEXT.md` 一旦建立就是全專案單一入口（現況、文件索引、術語表、oracle 優先序、
已被推翻的斷言、worklist）。對話被壓縮或換 session 接手時先讀它。

---

## 6. 環境硬規則

- 編譯一律走 docker；Python 一律 docker + uv.venv，不動系統環境。
- 手打 `docker run` 一律帶 `--rm --log-opt max-size=10m --log-opt max-file=3`。
- **禁止任何 `docker image prune`／`system prune`／`volume prune`／`builder prune`／
  `rmi`／`container prune`。** 這台機器同時放著多個客戶專案的 image，誤刪過一次。
  只清理自己這次建立的容器。
- 派 subagent 時，上述邊界要寫進 prompt，不能只靠 agent 自律。

---

## 7. 不做的事

- 不散布原版執行檔、資料檔、美術、音樂、掃描件與倚天字型。公開產出只有引擎程式碼與翻譯文本，
  玩家自備合法原版。原版資料一律 gitignore。
- 不做、不移植、不重現複製保護機制。
- 不在還沒讀懂原版的情況下「先做個能跑的版本」。

---

## 8. 目錄結構（規劃）

```
docs/re/          IDA 逆向筆記
docs/re/00-master-index.md          RE 總表（速查：位址、格式、結構、函式、工具）
docs/re/00-remake-knowledge-gaps.md 缺口檢查表
docs/re/00-function-index.md        函式索引
docs/formats/     資料格式規格
docs/spec/        可實作規格（只有 READY 的能動工）
docs/manual/      官方英文手冊 markdown + 繁中
docs/manual-cht/  軟體世界中文說明書逐頁轉錄
docs/paragraphs/  段落書整理與翻譯
docs/walkthrough/ 繁中攻略
translations/     譯名表與語言資料
tools/            ida.sh、go.sh、IDC 腳本、索引產生器
workplace/        原版資料解壓後的工作區（gitignore）
internal/、cmd/   Go 引擎
```

---

## 9. 現在的狀態與下一步

現況與 worklist 的單一真相來源是 [`CONTEXT.md`](./CONTEXT.md)，動手前先讀它。
**逆向結果的速查表是 [`docs/re/00-master-index.md`](docs/re/00-master-index.md)**——
remake 階段要查「某個位址是什麼、某個格式怎麼解」時查那一份。

一句話現況：三道閘門都走過了——資料格式、文字層、規則層與世界互動層已解，
**26 份規格全部 READY 並實作完成**，`cmd/wasteland -mode play` 可以從出廠存檔
走地圖、遇敵打完一場、進設施買賣治療學技能、存檔。
還沒解的集中在沒進遊戲主線的東西（`CURS` 與 `TRANSTBL` 的消費端），
逐項與「為什麼不擋實作」列在 `CONTEXT.md` §7.2。

## 素材位置

- GitHub repo：https://github.com/wicanr2/wasteland_cht （目前 private）
- 原版遊戲：`./Wasteland_1988.zip`
- 軟體世界中文說明書：`./112-荒野遊俠.rar`
- 方法論參考專案：`~/cht/daemon_winter/`（`CLAUDE.md`、`CONTEXT.md`、`AGENTS.md`）
