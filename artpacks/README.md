# 新版美術包

正式目錄固定為：

```text
artpacks/faithful-hd/manifest.json
artpacks/reimagined/manifest.json
```

`original` 直接使用玩家自備 ROM，不在此建立 manifest。遊戲切換新版模式前會由
`internal/artpack` 全量驗證候選目錄；任何缺檔、錯誤尺寸或雜湊不符都維持原模式。

`prototype-*` 目錄只供垂直切片，不是可發佈的完整主題，也不會被正式模式選取。
