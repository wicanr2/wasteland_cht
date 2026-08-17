# UTF-8 管線實驗（分支 `utf8-experiment`）

日期：2026-08-17 ｜ 基準：`master` 的 `e1d2906`

問題：**為什麼整條管線走 Big5？UTF-8 不行嗎？**

答案：**行**。這個分支把它做完了，測試全綠。下面是量到的數字與代價。

---

## 1. 改了什麼

```
目錄（.cat）        Big5 bytes  →  UTF-8
lang.Lookup         []byte      →  string
textlayout          RenderBytes →  ＋ Render（string）
訊息管線 s.cjk       []byte      →  string
逐格走訪             逐 byte 配對 →  range string（逐 rune）
畫字元               DrawCJK(hi, lo) → DrawRune(r)  ← **唯一轉 Big5 的地方**
```

`tools/build_lang.py` 仍然**在編譯期做 Big5 覆蓋檢查**（倚天畫不出來的字擋掉），
只是存進 `.cat` 的是 UTF-8。那一關沒有放鬆。

## 2. 量到的數字

| 項目 | master（Big5） | 這個分支（UTF-8） |
|---|---|---|
| 改動 | — | 43 檔、+536／−516 行（**淨 ＋20 行**）|
| `zh-Hant.cat` | 221,573 bytes | 280,650 bytes（**＋27%**）|
| `paragraphs-*.cat` | 42,564 bytes | 62,481 bytes（＋47%）|
| 程式碼裡的 Big5 觸點 | `internal/play` 53 處、`textlayout` 5、`render` 4、`cmd` 4 | **執行期只剩 1 處**（`render/hires.go` 的 `DrawRune`），其餘全是註解 |
| 畫一個漢字的轉碼 | 0 ns | 3,710 ns／24 字（**加快取後 510 ns，0 配置**）|
| 中文名字（輸入／顯示） | 6 字（13 bytes Big5） | **10 字**（30 bytes，`input.MaxName`）|
| 中文名字（**存得住**）| 6 字 | **10 字**（原版存檔 4 字 ＋ 側車檔，§4）|
| 測試 | 全綠 | 全綠 |
| 畫面 | — | 三張截圖與 master **逐位元組相同**（第四張差在 `，選擇：` 那個逗號修正，不是這次改的）|

### 轉碼成本

畫面上每個漢字每一幀都要轉一次 Big5。第一版直接呼叫 `x/text`：

```
BenchmarkRuneToBig5-14   3710 ns/op   960 B/op   72 allocs/op   （24 個字）
```

一畫面兩百個字 × 60 fps ＝ **每秒三萬六千次配置**，對遊戲迴圈是真的 GC 壓力。
加一層 `map[rune][2]byte` 快取（Big5 是固定對照表，永遠不會失效）之後：

```
BenchmarkRuneToBig5-14    510 ns/op     0 B/op    0 allocs/op
```

**7 倍、零配置。** 這是必要的，不是可選的。

## 3. 消掉的東西

改成 UTF-8 之後**下面這些程式碼與規則整批不需要了**：

| 消失的 | 它以前為什麼存在 |
|---|---|
| `singularBytes` | `singular` 的 Big5 版，逐 byte 切 `\n`，靠「`0x0A` 不可能是 Big5 尾位元組」這個前提才安全 |
| `trimSpaceBytes` | `strings.TrimSpace` 的 []byte 版 |
| `cjkCells` 的「≥ 0x80 就吃兩個」 | Big5 的格數換算。**對 UTF-8 是錯的**（漢字三個 byte）|
| `clipCells` 的兩兩配對與「落單高位元組」防呆 | 同上 |
| `eachCell` 的高／低位元組配對 | `range string` 直接給 rune |
| `lang.FromBig5` 的兩個呼叫端 | `wl-atlas` 要輸出 JSON、`wl-play` 要印 trace，都得先解回 UTF-8 |
| `TestCJKMessagesAreValidBig5` | 擋「UTF-8 漏進 Big5 串」——**那個 bug 類別不存在了** |
| 「要接進 Big5 串的字面值只能是 ASCII」這條規則 | 今天才因為 `、` 踩過一次 |

`cjkPageLabel` 從「先 `lang.ToBig5` 再回傳，編不出來就不印」變成一行字串串接。

## 4. 名字：側車檔（使用者定案 2026-08-17）

原本的代價是「存進存檔的中文名字從 6 字掉到 4 字」——角色記錄的名字欄是
13 bytes（`+0x00`–`+0x0C`），Big5 一字 2 bytes ＝ 6 字，UTF-8 一字 3 bytes ＝ 4 字。

⚠ **那一格不能加長。** `+0x0E` 起就是七個屬性，往後接著金錢、性別、國籍、
AC、MaxCON、CON——名字擴到 30 bytes 會一路蓋到 `+0x1D`，存檔整個壞掉。

定案的做法是**側車檔**（`internal/play/names.go`）：

```
game1／game2   +0x00  「沙漠遊」                ← 截斷版，原版遊戲讀得懂
WLNAMES（新）   槽 1   「沙漠遊俠一二三四五六」    ← 重製版讀這個
```

- 上限 **30 bytes ＝ 10 個中文字**（`input.MaxName`）。
  30 是**畫面決定的**：名單的名字欄有 15 格，10 個字放得下還留 5 格。
- 原版存檔那 13 bytes 照舊寫截斷版，`game.NameForSave` **在 rune 邊界截**——
  從中間切下去存檔裡會留半個字，而那個亂碼會寫進玩家的存檔。
- **短名字不進側車檔**（存檔那一格就夠了）。
- 側車檔不在、壞掉、版本不合、內容不是合法 UTF-8 → 退回存檔裡那 13 bytes，
  遊戲照跑。**寫側車檔失敗不算存檔失敗**（進度都在，掉的只有那截名字）。

先例是快速存檔的 `WLQS`（規格 27）：**不碰玩家的原版資料**是同一條原則，
`CLAUDE.md` §4 的「改寫不是重建」也沒有被動到——存檔本體仍然是
byte-for-byte round-trip。

其餘代價：`.cat` 大 27%（59 KB）、多一個檔、畫圖多一層 map 查詢（已量，可忽略）。

## 5. 結論與建議

技術上**成立**：淨增 20 行、測試全綠、畫面逐位元組相同、消掉一整類 bug。

名字那個代價已經用側車檔解掉了（§4），而且**比原來的 Big5 版更好**：
6 個字 → 10 個字。

剩下的只有 `.cat` 大 27% 與多一個檔，兩個都不影響玩家。
**建議合併。**

## 6. 可重跑

```bash
git checkout utf8-experiment
tools/go.sh test ./...
tools/go.sh test ./internal/render/ -bench BenchmarkRuneToBig5 -run XXX -benchmem
tools/go.sh run ./cmd/wl-shot -mode play -out /tmp/a.png     # 與 docs/images/01-map.png 比對
```

## 7. 這一輪學到的（寫成規則）

- **「為什麼用 X」的答案會過期。** 規格 11 §2 寫著「Big5 編碼的知識放在 Python，
  Go 不需要依賴任何編碼函式庫」——那是當初的主要理由，
  後來「名字可打中文」讓 `internal/lang` 引進了 `x/text`，那條理由就失效了，
  而規格沒跟著改。**引進一個相依之前，先查它會不會推翻某個既有決策的前提。**
- **編碼決策要分成三層問**：儲存（`.cat`）、傳遞（管線）、消費（字型索引）。
  這裡真正被外部限制綁死的只有第三層（倚天是 Big5 排列）與存檔的名字欄；
  中間那層是自由的，而**代價全部發生在中間那層**。
- **每幀每字的轉換要量配置次數，不是只量時間。** 3,710 ns 看起來沒問題，
  72 次配置才是遊戲迴圈的真問題。
