# 54：設施畫面的版面

日期：2026-08-15 ｜ 對應 `docs/spec/23` §2 留的「圖擺在哪」

輸入：`allpics1` 的 SHA-256 見 `docs/re/01`；
實機截圖來自 `tools/dosbox.sh`（環境見 `docs/re/47`）。

---

`docs/re/29` §5.4 讀出「進設施會載一張 `ALLPICS` 圖」，但**擺在哪、多大、
地點名印在哪**程式碼裡沒有直接寫（那是 overlay 的繪圖 slot 在做的）。
這一份用實機截圖量出來。

## 1. 量法

原版走到 Ranger Center，截一張圖，然後拿**解碼後的 `ALLPICS1` 圖片**
在畫面上搜位置——與 `TITLE.PIC` 那次同一招（`docs/re/47` §9）：

```
最佳：第 3 張圖，位置 (8, 8)，逐像素 7,938/8,064 ＝ 98.44%
```

**位置與圖片編號同時得到答案**：`(8, 8)` 就是地圖視窗的原點，
第 3 張正是 `ds:A4E0h` 表裡編號 3 那一支（存檔／管理設施，`docs/re/29` §5.4）——
**與程式碼讀出來的編號互相印證**。

## 2. 版面

| 元素 | 位置 | 大小 |
|---|---|---|
| 設施圖 | (8, 8) ＝ 地圖視窗原點 | 96 × 84 |
| 地點名 | **字元列 12**、字元欄 1 起 | 一行 |
| 分隔線 | 螢幕列 105 | 橫貫 |

⚠ **不是「整個地圖視窗換成圖」。** 圖只佔視窗左邊 96 × 84，
**右邊的地圖照常露出來**（截圖裡看得到地形與沙漠）。
規格 23 §2 原本那句「把地圖換成那張圖」講得太寬。

## 3. 沒對上的 126 個像素是動畫

不吻合的像素全部落在**螢幕列 16–26、x 65–85**——建築物右上角那個
碟形天線的位置，不是地點名那一帶。

那是 `ALLPICS` 交錯的參數區在畫的**局部動畫**（`sub_10A7A` 拆表、
`sub_10B11` 逐格 XOR 上去，`docs/re/23` §5）。把動畫照表 A 的順序
重建、疊到第 3 格，這 126 個像素**全部消失**：

```
python3 tools/verify_pic_anim.py workplace/orig/wastland/allpics1 3 \
  --screen workplace/dosbox/shots/db05.ppm --at 8,8
→ 疊到第 3 步（格 3）：差 0 像素
```

推論等級：位置、尺寸、差異來源全部 **已確認**（實機逐像素 100% 吻合）。

## 4. 可重跑的完整指令

```bash
# 走到 Ranger Center 並截圖
tools/dosbox.sh "wait:6;key:Return;wait:3;key:p;wait:4;key:d;wait:2;type:2;\
wait:1;key:Return;wait:2;key:Left;wait:2;key:Left;wait:2;shot:db05"

docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  -v "$PWD/workplace/dosbox/shots:/shots" --entrypoint sh wasteland-dosbox:latest \
  -c 'convert /shots/db05.png -crop 320x200+0+0 +repage /shots/db05.ppm'

# 搜位置（把 ALLPICS1 的每一張在畫面上滑一遍）
python3 - <<'PY'
import sys; sys.path.insert(0,'tools')
from compare_screen import read_ppm, to_index
# …見 §1，完整腳本在 commit 訊息引用的那一輪
PY
```

## 5. 這一輪學到的（寫成規則）

- **「搜位置」同時回答了「是哪一張」。** 拿全部候選圖在畫面上滑一遍，
  最佳解的**編號**與**座標**一起掉出來——比先假設編號再驗位置省一半功夫，
  而且兩邊互相印證。
- **對不上的那幾個像素往往是另一個未解的機制。** 126 個像素不是誤差，
  它們集中在一個小區域，而那個區域正好是 A9 那段沒解的參數區在管的東西。
  **殘差的分布會指向下一個題目。**
- **「換成一張圖」這種說法要用畫面驗。** 程式碼只說「載入並畫出來」，
  「畫多大、蓋掉什麼」得看螢幕——我原本寫成整個視窗換掉，實際只佔左邊三分之一。
