# 翻譯目錄

規格見 [`docs/spec/11`](../docs/spec/11-translation-catalogue.md)。

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
