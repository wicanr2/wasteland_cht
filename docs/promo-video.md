# 推廣片：怎麼重跑、為什麼長這樣

76 秒、1280 × 720、十四段。素材全部由專案自己的工具產生，沒有開過剪輯軟體，
整條路可重跑。

```bash
docker build -f docker/media.Dockerfile -t wasteland-media .   # 一次就好
tools/render_music.sh        # 配樂（需要自備 MT-32 ROM）
tools/promo/shots.sh         # 十張截圖（wl-shot，無頭）
tools/promo.sh               # 合成 → workplace/promo/out/
```

⚠ **成品不進版控**：畫面是原版的、聲音是 MT-32 算的，與原版資料、倚天字型同一個政策。
腳本進版控、產物不進（`workplace/` 整個 gitignore）。

---

## 1. 三段管線

| 段 | 腳本 | 產出 |
|---|---|---|
| 擷取 | `tools/promo/shots.sh` → `cmd/wl-shot` | `workplace/promo/shots/*.png`（640 × 400） |
| 配樂 | `tools/render_music.sh` → `tools/make_music.py` ＋ munt | `workplace/music/theme.ogg` |
| 合成 | `tools/promo.sh` → `tools/promo/make_promo.sh` ＋ `theme.sh` | `workplace/promo/out/wasteland-cht-promo.mp4` |

每一段的輸入輸出都是檔案，中間沒有手工步驟。截圖的座標、地圖編號、按鍵序列
全部寫死在 `shots.sh` 裡——**同一份原版資料跑兩次會得到同一批圖**。

## 2. 配樂的來歷（要講清楚）

**原版沒有背景音樂。** DOS 版只有九首 PC 喇叭音效，其中只有音效 4 是旋律，1.8 秒
（`docs/re/44`）。所以片子裡的曲子不是抽自原版，是重製版自己寫的
（`tools/make_music.py`），用 Roland MT-32／CM-32L 算成音檔——
1988 年 DOS 遊戲的高階音源，年代與機種對得上。

這件事要在任何對外場合講明，不能讓人以為聽到的是 1988 年的原聲。

## 3. 這一部的視覺 token：「鏽與沙」

色票不是憑喜好挑的，是從遊戲自己的畫面量出來的：

```bash
convert shot.png -resize 100x100 -colors 8 -depth 8 -format %c histogram:info:- | sort -rn
```

四張截圖跑下來，第一名一律是近黑 `#020202`（EGA 底色，佔比四成以上），
接著是鏽色 `#9A5331`（標題畫面的地面、沙漠的岩層、結局的土坡）與
沙黃 `#F0E046`（荒漠地圖的主色）。所以背景走近黑到暗鏽的漸層，
標題用沙黃、框線用鏽色。

母題選「軍規檔案」：掃描線 ＋ 四角的框線記號。主角是一支準軍事單位
（Desert Rangers），而 1988 年的 EGA 畫面本來就帶 CRT 味。

字體用黑體（Noto Sans CJK），不用襯線——襯線是奇幻題材的語彙，這一部是核戰後的沙漠。

## 4. 六種版面輪流用

單一版面重複十四段會很單調，所以 `make_promo.sh` 準備了六個版面函式：

| 版面 | 用在哪 |
|---|---|
| `card` 標題卡 | 開場、結尾 |
| `slide` 框內截圖 ＋ 下方字幕 | 七段遊戲畫面 |
| `compare2` 上下對照 | 原版英文 ↔ 繁體中文 |
| `quote` 引用卡 | 1990 年說明書與 CD 盒文案 |
| `stat5` 數據卡 | 專案帳目 |

判斷單不單調的方法：合成完抽四五幀並排看，如果四格長得都一樣就回去換版面。
`make_promo.sh` 收尾會自動抽 `frame-{3,15,28,45,62}.png`。

## 5. 分鏡（保存敘事）

世界觀 → 一手史料 → 玩得到的東西 → 這些字怎麼來的 → 結局 → 帳。

| # | 版面 | 內容 |
|---|---|---|
| 1 | card | 荒野遊俠 / WASTELAND · Interplay 1988 |
| 2 | slide | 標題畫面：1998 年，衛星從天上消失 |
| 3 | quote | 說明書〈一、故事介紹〉：Desert Rangers 誕生了 |
| 4 | slide | 荒漠地圖：走一步，時鐘跳四分鐘 |
| 5 | slide | 戰鬥：公式從執行檔讀出來 |
| 6 | compare2 | 技能清單的英文 ↔ 中文 |
| 7 | slide | 問答：逐位元組比對 |
| 8 | quote | CD 盒背面文案：拿起你的烏茲衝鋒槍 |
| 9 | slide | 手札：162 段劇情，不移植防拷 |
| 10 | slide | 設施：招牌、選單、清單都是中文 |
| 11 | slide | F1／F2／F5／F9 與背景音樂 |
| 12 | slide | 結局：自毀倒數 240 步 |
| 13 | stat5 | 42 張地圖 / 100 份逆向筆記 / 4,873 條文本 / 162 段 / 27 份規格 |
| 14 | card | 專案位址與版權立場 |

## 6. 踩過的坑（照抄，不要重踩）

- **不要用 `zoompan`。** `-loop 1 -t S` 配上 `fps` 濾鏡會讓 zoompan 的 `d` 變成
  「每個輸入幀輸出 d 幀」——6 秒 25fps 算成兩萬多幀，CPU 燒滿好幾分鐘，
  而輸出一直是空的。靜態圖加淡入淡出就夠了。
- **配樂比影片短時不要用 `-shortest`**：它會以音軌為準把結尾整張卡截掉。
  先 `aloop=loop=-1` 再 `atrim=0:$DUR`，兩軌自然等長。驗收看 `ffprobe`
  的視訊與音訊長度是不是都等於分鏡總長。
- **`-out` 要給相對於 repo 根目錄的路徑**：`tools/go.sh` 在容器裡把 repo 掛到
  `/src` 並 `-w /src`，主機的絕對路徑在裡面不存在，症狀是
  `no such file or directory` 指著一個明明存在的目錄。
- **`-lang ""` 沒有用**：空字串經過 shell 會被吃掉，後面的 `-out` 跟著錯位，
  圖寫到 `shot.png` 去。要英文對照組就給一個不存在的檔名（`-lang no-such.cat`）,
  `wl-shot` 對載不到翻譯是容忍的。
- **截圖放大只能用整數倍配 `-filter point`**。截圖本身已經是原版 320 × 200 的
  2 倍，再乘 1.5 剛好讓原版一個像素變成 3 × 3；非整數倍會讓有些像素兩格寬、
  有些三格寬，pixel art 看起來就髒了。
- **裁訊息視窗要停在 y < 384**：指令列在字元列 24，切太高會把它的上半截帶進來，
  看起來像壞掉的字。
- CPU：`docker run --cpus=2` ＋ ffmpeg 一律 `-preset veryfast -threads 2`。
  工具 image 先建好，不要每次 `apt install`。
