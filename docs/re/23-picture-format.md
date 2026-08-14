# 23：圖片格式 —— packed 4bpp ＋ 列間 XOR delta

日期：2026-08-14 ｜ 對應盤點 **A9**（`ALLPICS`）、補完 **A14** 與 `docs/re/03` §6

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`allpics1/2`、`title.pic` 的 SHA-256 見 `docs/re/01`。

---

## 1. 結論

原版的圖片全部是**同一套格式**，只有尺寸不同：

```
packed 4bpp        一個 byte 兩個像素，高 4 位在左
列間 XOR delta     out[n + stride] ＝ in[n + stride] XOR out[n]
                   **XOR 的回看距離就是一列的 byte 數**
```

| 來源 | 大小 | stride | 尺寸 | 解碼碼在哪 |
|---|---:|---:|---|---|
| `ALLPICS1/2` 的圖片子區塊 | 4,032 | 48 | **96 × 84** | overlay slot 2（`sub_10144`） |
| `TITLE.PIC` | 18,432 | 144 | **288 × 128** | `start` 內嵌（`docs/re/03` §6） |
| `END.CPA` | 18,432 | 144（推測） | 288 × 128（推測） | `sub_1B7FE` |

`stride × 高 ＝ 檔案大小` 三個來源都整除，而且**每張圖都用滿 16 種顏色**——
與 EGA mode 0Dh 的 16 色調色盤一致。

## 2. 解碼：`sub_10144`（overlay slot 2）

整支只有 18 bytes：

```
0x10144  8b362d47  mov  si, ds:472Dh     ; 圖片緩衝區
0x10148  bb2e00    mov  bx, 2Eh          ; 46
0x1014B  b9c807    mov  cx, 7C8h         ; 1,992 次
0x1014E  ad        lodsw                 ; ax ← [si]，si += 2
0x1014F  3100      xor  [bx+si], ax      ; [si + 46] ^= ax
0x10151  e2fb      loop loc_1014E
```

`lodsw` 之後 `si` 已經加 2，所以寫入點是**讀取點 ＋ 48**。
1,992 × 2 ＝ 3,984，加上跳過的前 48 bytes 正好是 **4,032**——緩衝區剛好用完。

`si ≥ 48` 之後讀到的是**已經解過的內容**，所以這是滾動的自參考解碼，
順序不能顛倒。與 `TITLE.PIC` 的 `0x111C5` 那段是同一個寫法，只有距離不同
（`0x90` ＝ 144）。

## 3. 為什麼會走錯一次

第一次把它當成 EGA 4 平面（`COLORF.FNT` 是 4 平面，`docs/re/14` §3），
畫出來是雜訊。用兩個量測把方向掰回來：

| 量測 | 未解碼 | 解碼後 |
|---|---:|---:|
| `0x00` 的個數 | 3,077 ／ 4,032 | 661 |
| 位元組自相關的最高峰 | stride **48**（0.696） | stride **48**（0.768） |

- 自相關在 48 有明顯尖峰 → **一列就是 48 bytes**，與 XOR 距離一致。
- 解碼後自相關**上升**（0.696 → 0.768）→ 方向對了（delta 解開後列與列更像）。
- `4,032 ＝ 96 × 84 ÷ 2` → 4bpp 而不是 4 平面。

**「同一個檔案裡有 4 平面字型」不代表圖片也是 4 平面。**
`COLORF.FNT` 是 4 平面（字元要用 EGA 的 map mask 一次寫一個平面），
圖片是 packed 4bpp（整塊搬進畫面）——兩種在同一個遊戲裡並存。

## 4. `ALLPICS` 的容器結構

`split_all` 的結果是嚴格交替：

```
allpics1: 66 個子區塊 ＝ 33 × (4,032 圖片 ＋ 變動長度參數區)
allpics2: 98 個子區塊 ＝ 49 × (4,032 圖片 ＋ 變動長度參數區)
```

共 **82 張圖片**。參數區長度 430–2,490 bytes 不等，**內容未解**——
`sub_10A7A`（overlay slot 16）會把它拆成兩張以 `0xFF`／`0xFFFF` 分隔的指標表，
形狀像動畫的影格清單。

載入在 `sub_184E8`：先解壓 `0xFC0` bytes 到 `ds:472Dh`（圖片），
再解壓參數區到 `ds:2700h`（或圖片編號 `0x34` 時到 `ds:0`），
然後呼叫 slot 2 解 delta、slot 16 處理參數。
`ds:BF12h` 是目前的圖片編號、`ds:BEFBh` 是上一張——**相同就直接返回**，
所以重複要求同一張圖不會重解一次。

**圖片編號是敵人資料的 `+0x07`**（`docs/re/37` §3.2）：遭遇時載入那種敵人的
肖像圖，而同一個編號查 `ds:A920h` 還會決定文字裡用 him／her／it。

## 5. 還沒解的

- **參數區的格式**（`sub_10A7A` 拆出來的兩張指標表）。
- `ALLHTDS1/2`：4／5 個 8,448–20,864 bytes 的大塊，**大小各不相同**，
  所以不是固定尺寸的圖；`sub_186B6` 解壓到 `seg003:0x2F60`。用途未解。
- `END.CPA` 的 stride 是不是 144（大小與 `TITLE.PIC` 相同，但解碼路徑
  `sub_1B7FE` 還沒逐條讀）。
- 調色盤（盤點 A14）：16 色全部用到，但**沒找到設定調色盤的程式碼**，
  推測是 mode 0Dh 的預設 EGA 調色盤。

## 6. 可重跑的完整指令

```bash
python3 tools/decode_pic.py allpics workplace/orig/wastland/allpics1 0
python3 tools/decode_pic.py title   workplace/orig/wastland/title.pic

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/pictures.json \
  0x10144 0x10A7A 0x184E8 0x186B6 0x1B7FE --callers
```

## 7. 這一輪學到的（寫成規則）

- **XOR 自參考解碼的回看距離，就是那筆資料的列寬。** 這是免費的幾何資訊：
  `TITLE.PIC` 的 `0x90` 給了 288 px、`ALLPICS` 的 48 給了 96 px，
  兩個都不用猜。看到 delta 解碼先把距離記下來當 stride。
- **像素格式要量，不要類推。** 同一個遊戲裡 4 平面與 packed 4bpp 並存，
  用途不同（字型 vs 整塊圖）。判斷方法是算術：
  `4,032 ＝ 96 × 84 ÷ 2` 對得上 4bpp，對不上 4 平面。
- **「解碼後更有結構」是可以量的。** 自相關上升、尖峰位置不變，
  就是方向對了；靠肉眼看文字圖容易被雜訊誤導。
