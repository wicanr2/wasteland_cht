# 靜態網站

把 [`docs/walkthrough/swm-005/`](../docs/walkthrough/swm-005/) 的整理成果做成可瀏覽的網頁。
內容來源只有那個目錄，這裡不放任何獨立的正文——要改內容改 markdown，不要改 `.html`。

## 怎麼產生

```bash
python3 site/build.py
```

只用 Python 標準函式庫，不需要套件、不需要 docker、不需要網路。
每次跑會重寫 `site/*.html`，產出九頁（index ＋ 六頁轉錄 ＋ 用語表 ＋ 機制對照表）。

改完 markdown 記得重跑一次，否則網站上還是舊的。

## 怎麼在本機預覽

```bash
python3 -m http.server 8000 --directory site
```

然後開 <http://localhost:8000/>。

直接用瀏覽器開 `site/index.html` 也可以，頁面之間的相對連結一樣會通。

## 為什麼是產生器 ＋ 靜態 HTML，不是 Jekyll

兩個做法都試算過，選了維護成本較低的這個：

| | 產生器 ＋ 靜態 HTML | Jekyll |
|---|---|---|
| 本機預覽要裝什麼 | 什麼都不用（Python 3 內建） | Ruby、bundler、jekyll，或改用 docker |
| GitHub 端要設定什麼 | 指到分支目錄就結束 | 走 Pages 的 Jekyll 建置，多一層會失敗的東西 |
| 外部資源 | 零，CSS 內嵌在每頁 | 預設主題從 remote gem 來，要另外處理才能自包含 |
| 產出可預期性 | 版本庫裡看得到最終 HTML，diff 看得出改了什麼 | 產物在 CI 裡，本機和線上可能不一致 |
| 代價 | 改 markdown 後要記得重跑 | 不用重跑 |

唯一的代價是那個「要記得重跑」。相對於為了預覽一頁文字而在專案裡多一條 Ruby
工具鏈，這個代價可以接受——尤其這個網站的內容是**已經定稿的史料轉錄**，
不是會天天改的文件。

排版上的取捨：正文寬度限制在 40em、行高 1.9、用襯線字，都是為了中文長文的可讀性。
字型只寫 family 名稱讓系統自己挑，沒有下載任何字型檔——專案不散布倚天字型，
網頁這邊也沒有必要引進 web font。深色與淺色模式都定義了色票，跟隨系統設定。

## 要發佈到 GitHub Pages 需要做什麼

repo 目前是 private，Pages 沒有開。真要發佈的話，在 GitHub 上的設定是：

**已經開好了**：https://wicanr2.github.io/wasteland_cht/

發佈方式是**手動推 `gh-pages` 分支**，一行指令：

```bash
tools/publish_site.sh
```

它會重跑 `build.py`、檢查沒有外部資源、把 `site/*.html` 推成 `gh-pages` 的根目錄。
Pages 的設定是「Deploy from a branch」→ `gh-pages` → `/`。

⚠ **兩個限制夾出這個做法**：
- 分支模式的資料夾只能選 `/` 或 `/docs`（GitHub 的限制，API 給別的值直接回 422），
  而這個 repo 的 `docs/` 放的是逆向筆記與規格，不是網站。
- 用 GitHub Actions 可以只發佈 `site/`，但**這個帳號沒有 Actions 額度**。

所以把 `site/` 的內容推成另一個分支的根目錄。代價是**改完要自己跑一次腳本**，
不像 Actions 會自動觸發。

### 發佈前要先確認的事

這個 repo 的政策是**不散布原版素材**（`CLAUDE.md` §7）。開 Pages 等於把 `/site`
底下的東西公開，所以發佈前要確認：

- `site/` 底下**只有 `.html`、`build.py` 和這份 README**，沒有掃描件、沒有遊戲資料
- 轉錄文字本身沒有夾帶原版的圖檔或執行檔片段

目前的產出符合這兩點：九個 HTML 全部是文字，沒有 `<img>`，也沒有任何外部請求。

還有一點要先想清楚：**這份轉錄的內容來自《軟體世界》雜誌**。
自用保存和公開發佈是兩件事，公開之前該確認一次授權狀態。這個決定不在本目錄的範圍內，
留給專案維護者判斷。

## 檔案

| 檔案 | 說明 |
|---|---|
| `build.py` | 產生器，唯一需要手動維護的程式 |
| `index.html` | 由 `swm-005/README.md` 產生 |
| `p56.html` – `p61.html` | 六頁轉錄 |
| `glossary.html` | 用語與譯名 |
| `mechanics-claims.html` | 機制斷言與逆向對照 |

`.html` 全部是產生物。手動改了會在下次跑 `build.py` 時被蓋掉。
