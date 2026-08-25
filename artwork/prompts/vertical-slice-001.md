# Vertical slice 001：沙漠聚落雙模式概念板

- 日期：2026-08-21（Asia/Taipei）
- 工具：OpenAI image generation（本 Codex 工作階段）
- 用途：可丟棄 prototype；比較 `faithful-hd` 與 `reimagined`，不作正式資產。
- 禁止：既有 IP 角色、品牌、Fallout 專屬符號、特定在世藝術家風格。

## 提示詞

製作一張橫向 16:9 的《荒野遊俠》remake 美術方向比較板；同一個核戰後美國西南
沙漠聚落，左右兩半且沒有文字。左半為嚴格俯視正交網格、可切成 tile 的現代精緻
像素美術；右半為同一地點的固定 3/4 俯視遊戲構圖，有長陰影、沙塵、分層高度與
清楚角色輪廓。兩側使用赭石、沙黃、鏽紅、灰藍，保持 1980 年代冷戰末日氣質。

## 輸出與後處理

原始生成圖由工具保存在工作階段生成目錄；使用 Docker 唯讀掛載後分割：

- 左半縮至 960×600：`prototype-faithful-hd.png`
- 右半裁成 16:9 並縮至 1280×720：`prototype-reimagined.png`

分割圖只用於垂直切片及 UI／manifest 測試。正式 tileset 必須另以固定索引、無縫
邊界及鄰接規則重新生成，不能直接從概念板裁小格冒充。
