# 11：翻譯目錄

狀態：**READY** ｜ 日期：2026-08-15
對應 `translations/`、`tools/extract_strings.py`、`tools/build_lang.py`、`internal/lang`

---

## 1. 依據

| 規格內容 | 來源 | 推論等級 |
|---|---|---|
| 執行檔九張表 444 槽／426 條非空 | [`docs/re/17`](../re/17-packed-text.md) | 已確認 |
| 42 個地圖區塊 4,445 槽／4,401 條非空 | [`docs/re/18`](../re/18-block-text.md) | 已確認 |
| 文字控制碼與變形機制 | [`docs/re/14`](../re/14-fonts-and-text-encoding.md)、[`28`](../re/28-text-variants.md) | 已確認 |
| 中文排版格與字寬 | [`docs/spec/10`](10-cjk-layout.md) | — |
| 段落編號是純文字 | [`docs/re/33`](../re/33-paragraph-references.md) | 已確認 |

## 2. 分層：原文、翻譯、編譯產物

```
translations/source/*.tsv     原文（抽出來的，機器產生，不手改）
translations/zh-Hant/*.tsv    中文（人翻的，唯一要手動維護的東西）
translations/glossary.md      譯名表（唯一真相）
translations/zh-Hant.cat      編譯產物（Big5，Go 讀這個）
```

- **原文檔是機器產生的**：`tools/extract_strings.py` 從原版資料抽，任何時候可以重跑。
  重跑之後原文有變 ＝ 抽取工具改了，要回頭看，不是手動改原文。
- ⚠ **`translations/source/` 不入版控**：那是原版的英文文本，屬於原版資料
  （`CLAUDE.md` §7 不散布）。每個人從自己那份原版抽。
  公開的只有 `zh-Hant/`（我們翻的）與 `glossary.md`。
- **中文檔是人翻的**，UTF-8、TSV、一行一條，diff 得動。
- **`.cat` 是編譯產物**：`tools/build_lang.py` 把 UTF-8 轉成 Big5 並檢查排版，
  Go 只讀它。**Big5 編碼的知識放在 Python，Go 不需要依賴任何編碼函式庫。**

## 3. Key 的形狀

Key 要**穩定**（原版資料不變就不會變）而且**看得出來源**：

```
exe:<表編號>:<槽>          執行檔的九張表，表編號 0–8
blk:<檔名>:<資源編號>:<槽>  地圖區塊，例 blk:game1:12:37
```

槽編號就是原版的字串索引，**不是流水號**——這樣新增／刪除都不會讓別的 key 位移。

## 4. TSV 的欄位

```
key <TAB> text
```

- `text` 裡的控制碼寫成 `\xNN`（`\r`、`\x06`、`\x10` 這些在原版是有意義的，
  **不能刪掉也不能自己加**）。
- 空字串的槽**不進檔案**（4,889 槽裡有 62 個是空的）。
- 註解行以 `#` 開頭。

## 5. 譯者規則（build 會擋）

| 規則 | 為什麼 | build 的行為 |
|---|---|---|
| **格數不得超過原文** | 訊息視窗只有 6 行 × 38 格（`docs/spec/10` §3） | 超過就是錯誤 |
| **控制碼要一字不差地保留** | `\x0A`／`\x0C`／`\x0E`／`\x0F` 是文字變形的分段（`docs/re/28`），少一個就會選錯段 | 數量對不上就是錯誤 |
| Big5 沒有的字要換掉 | 倚天字型是 Big5 排列 | 列出來並拒絕編譯 |
| 譯名以 `glossary.md` 為準 | 一致性 | 目前只警告 |

**中文比英文密度高**，所以「格數不得超過原文」在實務上很鬆——
`You are being attacked!` 是 23 格、「你們遭到攻擊！」只有 8 格。
會擋到的通常是**漏刪原文的空白或誤加註解**。

## 6. `.cat` 的格式

```
magic "WLCAT\0"  (6 bytes)
版本   uint16
條數   uint32
逐條：keyLen uint16、key（UTF-8）、textLen uint16、text（Big5 bytes）
```

刻意做成最笨的格式：**Go 那邊只要會讀就好**，所有判斷都在編譯時做完了。

## 7. 執行期的行為

```
lang.Load(path)      載不到 → 整個中文關掉，遊戲跑英文（不得崩）
lang.Lookup(key)     沒有這條 → 回 false，呼叫者用原文
```

**沒翻的條目就顯示原文**，不留空白也不顯示 key——半成品的中文化要能玩。

## 8. 驗收條件

1. 抽取：`exe` 426 條 ＋ `blk` 4,401 條 ＝ **4,827 條**，與 `docs/re/17`／`18` 的數字一致。
2. Key 唯一，而且重跑抽取兩次結果完全相同（byte-for-byte）。
3. 控制碼 round-trip：TSV 的 `\xNN` 解回來要與原文的 bytes 完全相同。
4. 編譯：格數超過原文、控制碼數量不符、Big5 缺字，三種都要擋下來並指出是哪一條。
5. Go 載入：查得到的回中文，查不到的回 false；**沒有 `.cat` 檔時遊戲照跑**。
6. 端到端：`-mode play` 踩到一個有翻譯的訊息格，畫面上出現中文。
