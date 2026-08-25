# 新版美術目前狀態

> 更新：2026-08-21。這是唯一現況表；概念稿與歷史 rejected manifest 不可單獨
> 用來宣稱完成。

| 交付項 | 必要總量 | 目前正式候選 | 狀態／證據 |
|---|---:|---:|---|
| faithful-hd tileset 0 | 66 | 66 | 尺寸／數量 PASS；真正 `play.Scene` review PASS |
| faithful-hd tileset 1 | 141 | 141 | 尺寸／數量 PASS；兩張真正 `play.Scene` review PASS |
| faithful-hd tileset 2 | 163 | 163 | 尺寸／數量 PASS；真正 `play.Scene` review PASS |
| faithful-hd tileset 3 | 107 | 107 | 尺寸／數量 PASS；真正 `play.Scene` review PASS |
| faithful-hd tileset 4 | 127 | 127 | 尺寸／數量 PASS；真正 `play.Scene` review PASS |
| faithful-hd tileset 5 | 118 | 118 | 尺寸／數量 PASS；真正 `play.Scene` review PASS |
| faithful-hd tileset 6 | 90 | 90 | 尺寸／數量 PASS；真正 `play.Scene` review PASS |
| faithful-hd tileset 7 | 104 | 104 | 尺寸／數量 PASS；真正 `play.Scene` review PASS |
| faithful-hd tileset 8 | 135 | 135 | 尺寸／數量 PASS；真正 `play.Scene` review PASS |
| faithful-hd 疊圖／隊伍 sprite | 10 個原版語意＋4方向隊伍動畫 | 10 靜態＋12 行走格 | alpha／安全邊界／每方向動作／Scene tick PASS |
| faithful-hd 場景 | 82 | 82 | 000–081 完整 bundle；真正 `play.Scene` renderer 已接 |
| faithful-hd 標題／結局 | 2 | 2 | 864×384；真正 `play.Scene` renderer 已接 |
| reimagined 3/4 tileset／環境 | 9 組、1,051 張 | 1,051 | 64×48、九組完整；動態可視範圍 18–25×8–15 |
| reimagined 個別角色 sprite | 最多 7 人、8 方向、全動作／武器層 | 728＋32 | idle／walk／attack／hurt／down／interact；四類武器獨立 alpha 層 |
| reimagined 分層場景 | 82＋標題＋結局 | 84 | 場景 768×432；全畫面 1280×720；完整 manifest |
| 遊戲內原子切換 | 3 模式 | 3 | F2 設定按 V 循環；錯誤維持舊模式；只有 reimagined 動態 16:9 |

## tileset 0 驗收

- 候選位置：`artpacks/faithful-hd/assets/tileset-0/000.png`–`065.png`。
- 每張 48×48 PNG；`tools/art/check_tiles.py --tileset 0` 為 PASS。
- 真實地圖：資源 0、隊伍座標 (55,62)、輸出 864×384；本機證據
  `workplace/art-preview/faithful-map-0-resliced.png`，SHA-256
  `f96700fedefcee205756636e118ded2da492d877230f530448fee47509cef364`。
- 真正 `play.Scene` 的新版 terrain＋alpha icon＋中文 HUD 證據：
  `workplace/art-preview/scene-faithful-map-0.png`，SHA-256
  `1ba0384b8a027003ea7d7904a0285f7310f9a5af247ba89b6489a3d37a6ffce8`。
- 目前只證明資源 0 的 terrain／靜態 sprite 垂直鏈，不證明其他地圖或完整模式可玩。

## tileset 1 驗收

- 候選位置：`artpacks/faithful-hd/assets/tileset-1/000.png`–`140.png`。
- 每張 48×48 PNG；`tools/art/check_tiles.py --tileset 1` 為 PASS。
- 真正 `play.Scene` 已同時載入新版 terrain、alpha icon、隊伍動畫與中文 HUD：
  - 資源 8：`workplace/art-preview/scene-faithful-map-8.png`，SHA-256
    `1df479e85f666e555b9202d58e393aaa032d44f48327e03778751453de524210`。
  - 資源 9：`workplace/art-preview/scene-faithful-map-9.png`，SHA-256
    `a8a6e6d15a47ebff33c9a94d0cf8f276299c7ca4ef7feb3c952169a332fa9935`。
- 九批來源、處理步驟與逐張雜湊記錄於
  `artwork/manifests/tileset-1-batch-*.json`；此驗收不外推為場景或完整模式已完成。

## tileset 2 驗收

- 候選位置：`artpacks/faithful-hd/assets/tileset-2/000.png`–`162.png`。
- 每張 48×48 PNG；`tools/art/check_tiles.py --tileset 2` 為 PASS。
- 真正 `play.Scene` 證據：資源 2 的工業內裝、NPC、隊伍動畫與中文 HUD，
  `workplace/art-preview/scene-faithful-map-2.png`，SHA-256
  `0280836042827e01e9d61912bc91b9738bf1acbf902624cc1601270cce4f7519`。
- 十一批來源、處理步驟與逐張雜湊記錄於
  `artwork/manifests/tileset-2-batch-*.json`；此驗收不外推為場景或完整模式已完成。

## tileset 3 驗收

- 候選位置：`artpacks/faithful-hd/assets/tileset-3/000.png`–`106.png`。
- 每張 48×48 PNG；`tools/art/check_tiles.py --tileset 3` 為 PASS。
- 真正 `play.Scene` 證據：資源 1 的道路、屋頂、水岸、植被、隊伍動畫與中文 HUD，
  `workplace/art-preview/scene-faithful-map-1.png`，SHA-256
  `bdde1488da8f6b49507383c5c69eae970996254f373bbe0272cea8c51ece9c47`。
- 七批來源、處理步驟與逐張雜湊記錄於
  `artwork/manifests/tileset-3-batch-*.json`；此驗收不外推為場景或完整模式已完成。

## tileset 4 驗收

- 候選位置：`artpacks/faithful-hd/assets/tileset-4/000.png`–`126.png`。
- 每張 48×48 PNG；`tools/art/check_tiles.py --tileset 4` 為 PASS。
- 真正 `play.Scene` 證據：資源 11 的地下設施、門、牆面、隊伍動畫與中文 HUD，
  `workplace/art-preview/scene-faithful-map-11.png`，SHA-256
  `522bc544b8c7e5564a8d4334566caec13ddd9a6ff9b0ff462269e147048b7d41`。
- 八批來源、處理步驟與逐張雜湊記錄於
  `artwork/manifests/tileset-4-batch-*.json`；此驗收不外推為場景或完整模式已完成。

## tileset 5 驗收

- 候選位置：`artpacks/faithful-hd/assets/tileset-5/000.png`–`117.png`。
- 每張 48×48 PNG；`tools/art/check_tiles.py --tileset 5` 為 PASS。
- 真正 `play.Scene` 證據：資源 38 的力場、科技牆、家具、角色與中文 HUD，
  `workplace/art-preview/scene-faithful-map-38.png`，SHA-256
  `d83580147320b3cbc1ed95c9c28aa16b2672a3a7bb0e19061ae124d641ec1dd6`。
- 八批來源、處理步驟與逐張雜湊記錄於
  `artwork/manifests/tileset-5-batch-*.json`；此驗收不外推為場景或完整模式已完成。

## tileset 6 驗收

- 候選位置：`artpacks/faithful-hd/assets/tileset-6/000.png`–`089.png`。
- 每張 48×48 PNG；`tools/art/check_tiles.py --tileset 6` 為 PASS。
- 真正 `play.Scene` 證據：資源 12 的沙漠道路、建築剖面、車輛與中文 HUD，
  `workplace/art-preview/scene-faithful-map-12.png`，SHA-256
  `8b5957f4aa75ee175e2cbf8292ef756cb180308fd50e8024f22c8eda9600f533`。
- 六批來源、處理步驟與逐張雜湊記錄於
  `artwork/manifests/tileset-6-batch-*.json`；此驗收不外推為場景或完整模式已完成。

## tileset 7 驗收

- 候選位置：`artpacks/faithful-hd/assets/tileset-7/000.png`–`103.png`。
- 每張 48×48 PNG；`tools/art/check_tiles.py --tileset 7` 為 PASS。
- 真正 `play.Scene` 證據：資源 7 的科技牆、力場地板、控制物件與中文 HUD，
  `workplace/art-preview/scene-faithful-map-7.png`，SHA-256
  `1d5451088849ceeb4df9c6c02e4c7f12ab3f0ee9e02523cf7246d32827475571`。
- 七批來源、處理步驟與逐張雜湊記錄於
  `artwork/manifests/tileset-7-batch-*.json`；此驗收不外推為場景或完整模式已完成。

## tileset 8 驗收

- 候選位置：`artpacks/faithful-hd/assets/tileset-8/000.png`–`134.png`。
- 每張 48×48 PNG；`tools/art/check_tiles.py --tileset 8` 為 PASS。
- 真正 `play.Scene` 證據：資源 16 的機械、彩色地表、牆體、障礙與中文 HUD，
  `workplace/art-preview/scene-faithful-map-16.png`，SHA-256
  `e66b3e77905cf8a1477bd947438853436c2f52a8eb187f96077567f723e453c5`。
- 九批來源、處理步驟與逐張雜湊記錄於
  `artwork/manifests/tileset-8-batch-*.json`。
- 九組 tileset 合計 1,051 張 48×48 圖塊均有正式候選；這只證明地圖圖塊覆蓋，
  不代表 82 張場景、reimagined 模式或玩家設定切換已完成。

## 現行垂直鏈與剩餘驗收

1. `artpacks/faithful-hd/manifest.json` 列出 1,157 個資產；固定 960×720（4:3）
   畫布使用 contain 呈現既有 960×600 內容，不裁切。
2. `artpacks/reimagined/manifest.json` 列出 1,905 個資產；由
   `tools/art/build_reimagined.py` 從已接受 master 可重現產生，逐檔記錄尺寸與
   SHA-256；1280×720 為預設，960×540 至 1920×1080 依視窗重繪。
3. 正常 `play.Scene` 證據為 `workplace/art-preview/reimagined-live-1280x720.png`、
   `reimagined-live-1600x900.png` 與 `reimagined-title-1280x720.png`。剩餘 gate 是
   場景／戰鬥／結局抽樣、三模式連續切換、全套測試與封包規則檢查。
