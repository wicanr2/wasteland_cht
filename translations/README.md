# 翻譯目錄

規格見 [`docs/spec/11`](../docs/spec/11-translation-catalogue.md)。
繁簡切換見 [規格 34](../docs/spec/34-zh-hans-and-language-switch.md)。
簡體來源為 `zh-Hans/*.tsv`，其中 `paragraphs.tsv` 保存 162 段正文。
兩語 `.cat` 均使用 UTF-8；Big5 只限制倚天驗證路徑，不代表簡體編碼。

簡體初稿工具為 [make_zh_hans.py](../tools/make_zh_hans.py)，固定工具來源與 SHA-256
見 [localization-inputs.json](../tools/localization-inputs.json)。在 Docker 內使用該 wheel，
產出的 TSV 可直接維護；遊戲執行時不依賴轉換工具。編譯命令：
`python3 tools/build_lang.py --locale zh-Hans` 與
`python3 tools/build_paragraphs.py --locale zh-Hans`，均須在既有 Docker 工具映像內執行。

```
source/       原文（機器產生，**不入版控**——原版資料不散布）
zh-Hant/      中文（人翻，唯一要手動維護的東西）
glossary.md   譯名表（唯一真相）
zh-Hant.cat   編譯產物（Big5，Go 讀這個）
```

## 怎麼跑

```bash
# 1. 從自己那份原版抽出原文（4,827 條）
python3 tools/extract_strings.py \
  workplace/analysis/unpacked/wl.merged.exe \
  workplace/orig/wastland/game1 workplace/orig/wastland/game2

# 2. 翻譯：編輯 zh-Hant/*.tsv，一行一條 `key<TAB>譯文`

# 3. 編譯（會擋格數超標、控制碼數量不符、Big5 缺字）
python3 tools/build_lang.py
```

## 譯者要守的三條

1. **格數不得超過原文**——訊息視窗只有 6 行 × 38 格。
2. **控制碼一字不差地保留**（`\x0A`／`\x0C`／`\x0E`／`\x0F` 是文字變形的分段，
   少一個就會選錯段）。
3. **譯名以 `glossary.md` 為準**；1990 年軟體世界說明書的既有譯名優先。

三條都在 `tools/build_lang.py` 裡有對應的檢查，前兩條會直接擋下編譯。
