# 26：圖片的局部動畫

狀態：**READY** ｜ 日期：2026-08-15 ｜ 對應 `internal/assets`、`internal/render`

`ALLPICS` 的每張圖後面跟著一段參數區，那是**局部動畫**——
設施畫面上的碟形天線、遭遇肖像上會動的部件都靠它。
規格 23 只講了「進設施要畫哪張圖」，這一份補「那張圖怎麼動」。

---

## 1. 依據

| 規格內容 | 來源 | 推論等級 |
|---|---|---|
| 兩張表的拆法（`sub_10A7A`） | [`re/23`](../re/23-picture-format.md) §5.1 | 已確認 |
| 表 A ＝ `(延遲, 格)` 播放腳本 | [`re/23`](../re/23-picture-format.md) §5.1 | 已確認 |
| 元素 word 的解法與 XOR 疊法 | [`re/23`](../re/23-picture-format.md) §5.2 | 已確認 |
| 實機逐像素 0 差異 ＋ 82 張圖的循環恆等式 | [`re/23`](../re/23-picture-format.md) §5.3 | 已確認 |
| 圖擺在 (8, 8)、96 × 84 | [`re/54`](../re/54-facility-screen-layout.md) §2 | 已確認 |

## 2. 資料

參數區接在圖片子區塊後面，同樣是 Huffman 壓縮，解開之後：

```
word  n           表 A 的 bytes 數
n bytes           表 A：以 0xFF 分隔的記錄，一筆 ＝ 一個動畫通道
word  m           表 B 的 bytes 數
m bytes           表 B：以 0xFFFF 分隔的格，一格 ＝ 若干元素
```

**表 A 一筆**（一個通道的腳本）：

```
延遲₀, 格₀, 延遲₁, 格₁, …        ← 第一個 byte 是初始倒數，不是格編號
```

放完最後一筆就從頭再來，**無限循環**。

**表 B 一個元素**：

```
word w
    相位 ← w & 3
    列, 欄 ← divmod((w >> 2) & 0x3FF, 12)      ← 一列 12 個 byte ＝ 96 像素
    長度 ← (w >> 12) + 1                        ← 酬載 bytes 數
長度 bytes 酬載                                  ← 一個 byte 兩個像素，高 nibble 在左
```

畫的位置：`x ← 欄 × 8 ＋ 2 × 相位`、`y ← 列`（**圖片內座標**），
從那裡開始逐像素 **XOR** 上去。

⚠ **相位不是「跳過幾個 byte」，是「左邊缺幾對像素」。** 原版靠部分組
（`dl = 3 − 相位`）讓第一組只吃 `4 − 相位` 個 byte，那些像素落在螢幕 byte 的
低位——也就是右半邊。少乘這個 2 會讓整段左移，肉眼看不太出來，
逐像素比對才會露餡。

⚠ **y 的 ＋8 不要重複加。** 原版的列位址表已經把「圖擺在螢幕 y ＝ 8」
烘進去了（`ds:8E09h` ＝ `ds:8DF9h` 往後 8 筆）。重製版拿的是**圖片內座標**，
螢幕位移由畫圖的人加一次就好。

## 3. 播放

```
每一拍（BIOS 計時器 0040:006C 跳一次）：
    每個通道：
        倒數 > 0 → 倒數 −1，這一拍不動
        否則     → 取腳本的下一個格，XOR 疊上去；再讀下一個 byte 當新的倒數
                   讀到 0xFF 就回腳本開頭
```

⚠ **XOR 疊上去之後不還原。** 畫面是「底圖 ⊕ 已播過的所有格」，
原版沒有存背景也沒有重畫底圖。這能成立是因為**一輪播完所有格 XOR 互相抵消**——
這條恆等式在 `allpics1`／`allpics2` 的 82 張圖上全部成立，可以拿來當自檢。

## 4. Go 介面

```go
// internal/assets

// PicAnim 是一張圖的局部動畫。
type PicAnim struct {
    Channels []AnimChannel // 表 A：每個通道一份腳本
    Frames   [][]AnimElem  // 表 B：每一格的元素
}

// AnimChannel 是一個通道的播放腳本。Delay[i] 是播 Frame[i] 之前要等幾拍。
type AnimChannel struct {
    Delay []byte
    Frame []byte
}

// AnimElem 是一格裡的一段橫向像素，XOR 到 (X, Y) 起的位置。
// Pixels 已經拆成一個像素一個 byte（值 0–15），X 已含相位偏移。
type AnimElem struct {
    X, Y   int
    Pixels []byte
}

// DecodePicAnim 解一張圖後面的參數區；沒有參數區時回傳零值與 nil。
func DecodePicAnim(raw []byte) (PicAnim, error)

// internal/render

// PicPlayer 依拍數推進一張圖的動畫，並把結果 XOR 進畫布。
type PicPlayer struct{ /* … */ }

func NewPicPlayer(a assets.PicAnim) *PicPlayer

// Tick 推進一拍，回傳這一拍要疊的元素（可能是空的）。
func (p *PicPlayer) Tick() []assets.AnimElem
```

## 5. 未解與邊界

- **一拍多長沒有量。** 原版比對的是 BIOS 計時器低位元組跳沒跳
  （18.2 Hz 的計數器），所以一拍 ＝ 一個 tick ≈ 55 ms，但原版**只在
  值變了才推進**，實際頻率受主迴圈影響。重製版先用 55 ms 一拍，
  之後與實機錄影對拍再調——這一項標 **假說**，其餘都是已確認。
- 多通道的圖（表 A 有多筆）沒有實機對拍過；結構與單通道相同。

## 6. 驗收條件

1. **逐像素**：`allpics1` 第 3 張圖，底圖疊上格 0–3 之後，與
   `workplace/dosbox/shots/db05.ppm` 在 (8, 8) 起的 96 × 84 區域**完全相同**。
2. **循環恆等式**：`allpics1`／`allpics2` 全部 82 張圖，把腳本跑完一輪之後
   畫面回到底圖，一個像素都不差。
3. **相位**：至少一個相位非 0 的元素，其起點是 `欄 × 8 ＋ 2 × 相位`。
4. **沒有參數區的圖**不當成錯誤，回零值。
5. **不改底圖緩衝區**：疊完之後解碼出來的那份 96 × 84 仍與解碼當下相同。
