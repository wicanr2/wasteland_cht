# 內建介面字型

新版介面內建兩份原始字型，壓縮儲存不變更其字型名稱或解壓後位元組。
原始字型的 SIL Open Font License 1.1 與本專案 RRSAL-1.0 分別保存。

| 字型 | 上游來源與固定提交 | 原文授權 |
|---|---|---|
| 源泉圓體 TW R，預設 | [ButTaiwan/gensen-font](https://github.com/ButTaiwan/gensen-font/tree/d347d3fffcb45e08857052433a0b432ed4f7ace8) | [GenSen-OFL.txt](../internal/presentation/fonts/GenSen-OFL.txt) |
| Noto Sans CJK TC Regular，可切換 | [notofonts/noto-cjk Sans2.004](https://github.com/notofonts/noto-cjk/tree/523d033d6cb47f4a80c58a35753646f5c3608a78) | [Noto-OFL.txt](../internal/presentation/fonts/Noto-OFL.txt) |

下載網址、上游提交、原始及嵌入檔案 SHA-256 見
[來源清單](../internal/presentation/fonts/sources.json)。
來源清單及兩份授權原文會附在之後建立的封包 `docs/`。
本頁不表示已更新先前的 `v.1.0.0-20261029` 本機包。

預設介面不需要倚天字型。`-ui classic` 保留舊版點陣字驗證路徑，倚天仍由玩家自備。
介面幾何、字級與定版依據見 [規格 33](spec/33-rounded-ui.md)。
