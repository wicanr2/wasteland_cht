# 129：哪一個畫面畫哪幾個標籤——`ds:7DF3h` 是一個 32-bit 的熱區遮罩

日期：2026-08-20 ｜ 接 [`126`](126-box-labels.md)（標籤的兩張表）、
[`112`](112-mouse-cursor-and-hotzones.md)（熱區表）、[`127`](127-roster-mode-boxes.md)（名單模式的框）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/126` 把標籤的**字模串**與**版面表**讀完了，末尾掛著一句
「哪一個畫面畫哪幾筆仍然未解——目前是拿實機截圖一張一張認的」。
答案是一個 32-bit 的遮罩：**每個畫面在進迴圈之前寫一次「我這個畫面有哪些熱區」**，
畫框那一層再去比對「現在畫著的是哪些」，差異的才畫或擦。

---

## 1. 遮罩：`ds:7DF3h`（低 16 位）＋ `ds:7DF5h`（高 16 位）

`sub_18A4F` 從 `bx ＝ 4`、`dx ＝ 0x10` 開始，每輪 `dx` 左移一位、`bx` 加一，
到 `bx ＝ 0x15`（21）為止——**bit n 對應熱區 n**，而熱區 4–20 與版面表
一筆對一筆（`docs/re/126` §3）。所以：

```
標籤 k 的開關 ＝ 遮罩的 bit (k + 4)
熱區 0–3（名單／地圖視窗／圖片／指令列）＝ bit 0–3
```

`bx` 走到 16 時換讀 `ds:7DF5h`（`0x18AAA`），所以兩個 word 合起來是一個
32-bit 遮罩，只用到低 21 位。

## 2. 只畫差異：`ds:8DD2h`／`ds:8DD4h` 是「現在畫著的」

```
0x18A4F  ax ← ds:7DF3h；ds:CA69h ← ax      ; 想要的
0x18A55  ax ^= ds:8DD2h                    ; **與已畫的取 XOR ＝ 差異**
0x18A5F  這一位有差 →
           想要的有這一位 → sub_18B6B（畫）
           否則              → sub_18B3C（擦）
0x18A9D  dx <<= 1；bx++；到 0x15 為止
```

`ds:CA6Bh` 是另一道閘（非 0 才畫），`dx` 的 bit8 那條分支走的是熱區本身的
啟用／停用（`sub_18AC4`／`sub_18B1B`）。**畫與擦共用同一份版面表**，
所以標籤不會擦錯位置。

## 3. 兩種版面：前四筆是直的

`sub_18B6B(bl ＝ 熱區編號)` 先算出版面表那一筆：

```
0x18B89  bx ← (bl − 4) × 6 ＋ 0xCBBD       ; **6 bytes 一筆、熱區 4 ＝ 第 0 筆**
0x18B9C  ds:8DD0h ← [bx]（欄）；ds:8DD1h ← [bx+1]（列）
0x18BA7  cl ← [bx+2]（長度）；si ← [bx+4]（字模串）
0x18BB1  cmp dl, 8 →
           ≥ 8：逐字模 sub_10039，**欄由繪製常式自己前進**（橫的）
           < 8：逐字模 sub_10039 之後 `inc 列` ＋ `dec 欄`（直的）
```

`dl` 是熱區編號，`< 8` 就是版面表的**第 0–3 筆 ＝ 四個箭頭**。
它們都在欄 39，而 39 已經是最後一欄——橫著排會排到畫面外，
所以那個 `dec ds:8DD0h` 正好抵銷繪製常式的欄前進，字模變成**往下疊**。
於是版面表的「長度 2」對這四筆是**兩列高**，不是兩格寬。

擦除那一支（`0x18BDA` 起，同一支的另一半）也分兩種：
橫的填字模 `0x12`、直的填 `0x13` ——那是外框的橫線與直線（`docs/re/124` §2），
所以擦掉標籤就是把框線補回去。

推論等級：**已確認**（三段程式碼逐條讀完；實機 `20-map.png` 的箭頭
確實佔列 18–19 與 22–23，逐像素比對過）。

## 4. 每個畫面的遮罩

`mov word ptr ds:7DF3h, imm` 全檔 54 處，逐處解出來的表在
[`generated/ida94/box-labels.md`](generated/ida94/box-labels.md)（`tools/summarize_box_labels.py`
第二段，純 stdlib 掃指令樣式，不需要 IDA）。挑出已定名的畫面：

| 畫面 | 位址 | 遮罩 | 標籤 |
|---|---|---|---|
| 地圖（名單收起）| `0x16BB5` | `0x40CA` | 訊息視窗的上下箭頭（2、3）＋ `ROSTER ON` |
| 地圖（名單展開）| `0x16BAB` | `0x200B` | `ROSTER OFF` |
| 戰鬥指令階段（`sub_11F76`）| `0x12023` | `0x00080404` | `ESC` ＋ `MAP` |
| 設施進場選人（`sub_1721B`）| `0x17225` | `0x0404` | **只有 `ESC`** |
| 商店主選單 | `0x1BF2C` | `0x8404` | `ESC` ＋ `POOL MONEY` |
| 商店買賣清單（`sub_1C073`）| `0x1C0AC` | `0x8434` | 清單的上下箭頭 ＋ `ESC` ＋ `POOL MONEY` |
| 醫生 | `0x1C345`／`0x1C3BD`／`0x1C44F` | `0x8404` | `ESC` ＋ `POOL MONEY` |
| 訓練師（`0x1C6C9`）| `0x1C7D0` | `0x0404` | `ESC` |
| 訓練師的技能清單 | `0x1C8C2` | `0x0C34` | 清單的上下箭頭 ＋ `ESC` ＋ `NEXT` |
| 選人之後的金錢畫面（`sub_19130`）| `0x19139` | `0x00060C04` | `ESC` ＋ `NEXT` ＋ `POOL` ＋ `DIVIDE` |
| 捲動速度（`sub_19004`）| `0x1900B` | `0x0300` | `SLOW` ＋ `FAST` |

推論等級：**遮罩的值與解碼已確認**（掃出來的立即數 ＋ §1 的位元對應）；
**「哪個位址屬於哪個畫面」分兩級**——`0x16BB5`／`0x12023`／`0x17225`／`0x1C0AC`
的函式在筆記裡已定名（`docs/re/26`、`38`、`119`、`117`），是**已確認**；
`0x1BF2C`（讀 `ds:DBF5h` ＝ 商店那條路的隊員編號）與 `0x1C6C9`
（緊接技能費用 `sub_1C68E`、遮罩裡沒有 `POOL MONEY`，與 `docs/re/119` 的
訓練師觀察一致）是**強證據**。

三處實機逐像素對過：`20-map.png` 的欄 39 在列 18–19 與 22–23 有箭頭；
`47-buy2.png` 的欄 39 在列 3–4 與 10–11 有箭頭，而同一張圖的**列 23 沒有
`ROSTER OFF`**；`42-shop.png`（選人那一步）連 `POOL MONEY` 都沒有。

⚠ **`ROSTER OFF` 只屬於地圖畫面。** 戰鬥與設施各自設自己的遮罩，
兩者都沒有那一位——把「名單露出來就畫 `ROSTER OFF`」當成通則會多畫一個按鈕。

⚠ **這張表只涵蓋立即數那一種寫法。** 地圖主迴圈是 `mov ds:7DF3h, ax`
（值由 `ds:46B9h` 選），掃不到；所以「某個標籤沒有畫面用到」不能只憑這張表下結論。
`SLOW`／`FAST` 與 `ENTER` 就是靠這張表才找到出處的。

## 5. remake 這一側

`Scene.boxLabels()` 現在照這張表回傳（繪製與滑鼠共用同一支）：

| 畫面 | 畫哪些 |
|---|---|
| 地圖 | `LabelMsgUp`／`LabelMsgDown`／`LabelRosterOn` |
| 戰鬥 | `LabelEsc`／`LabelMap` |
| 設施 | `LabelEsc` ＋（有 `P` 才）`LabelPoolMoney` ＋（清單超過一頁才）兩個箭頭，訓練師再加 `LabelNext` |

箭頭用 `BoxLabel.Vertical`（字模往下疊），按鍵存的是**原版的擴充碼**
（`0xC8`／`0xD0`／`0xC9`／`0xD1`），在 `internal/play/mouse.go` 轉成
重製版的動作：清單那一組送翻頁鍵 `I`／`K`，訊息視窗那一組送
`ScrollUp`／`ScrollDown`。

順帶修掉一個實際擋路的 bug：**訓練師的技能清單翻不了頁**。
它掛在主迴圈上（`Step` 一直是 `StepMain`），而翻頁的判準原本是
「不在主迴圈」——36 個技能只摸得到前 9 個，畫面上卻只是「清單就這麼長」。
原版的清單按鍵處理（`sub_16D34`）三種設施共用，訓練師清單也有上下箭頭，
所以判準改成「現在有沒有一份翻得了頁的清單」。

## 6. 可重跑的完整指令

```bash
# 版面表 ＋ 每個畫面的遮罩（純 stdlib）
python3 tools/summarize_box_labels.py workplace/analysis/unpacked/wl.merged.exe

# 三段程式碼
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/f18A4F.json 0x18A4F

# 實機：箭頭在欄 39 的哪幾列（兩張圖相減）
docker run --rm --log-opt max-size=10m --log-opt max-file=3 \
  -v "$PWD/workplace/dosbox/shots:/shots" --entrypoint sh wasteland-dosbox:latest \
  -c 'convert /shots/20-map.png -crop 24x64+296+136 +repage -filter point -resize 600% /shots/zoom.png'
```

## 7. 這一輪學到的（寫成規則）

- **「哪個畫面畫哪個東西」這種問題，先找有沒有一個開關集合。** 一張一張截圖去認
  是 O(畫面數)，而畫面自己會宣告——找到那個宣告（這裡是一個 32-bit 遮罩），
  剩下的就是解碼。**先問「誰決定的」，不要先問「這一張有什麼」。**
- **版面表的欄位對不同記錄可以有不同語意。** 「長度 2」對橫的標籤是兩格寬、
  對直的箭頭是兩列高——差別在畫的那一支用 `dl < 8` 分兩條路。
  **看到同一張表被兩段程式碼讀，要分別讀完再下結論。**
- **一個看得見的缺件，常常牽出一個看不見的 bug。** 這一輪是為了補箭頭才去讀遮罩，
  結果讀到「訓練師的清單也有箭頭」——而 remake 那份清單根本翻不了頁。
  **UI 元件缺席時，順手問一句「那個元件本來是給誰用的」。**
