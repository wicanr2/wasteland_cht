# 47：DOSBox 參考環境，與第一批實機對拍

`CONTEXT.md` §7.2 掛著「在 DOSBox 跑解包版與原版對照」與「規格 04 §8.5 的
逐像素對拍」。環境建好了，第一批對拍也跑完了。

⚠ **DOSBox 是 IDA 的驗證工具，不是替代品**（`CLAUDE.md` §1）。
它回答的是「讀出來的東西對不對」，不回答「這段程式在做什麼」。

---

## 1. 結論

| 事實 | 等級 |
|---|---|
| 解包版與原版跑出來的畫面**逐像素相同**（`compare -metric AE` ＝ 0） | 已確認 |
| `TITLE.PIC` 的解碼**逐像素正確**：36,864/36,864 | 已確認 |
| 圖片視窗在 **(8, 8)**——由對拍掃出來，不是假設 | 已確認 |
| mode 0Dh 用的就是**預設 EGA 調色盤** | 已確認 |
| 物品表 `+0x04`（容量）＝ 畫面上的 `AMM` 欄 | 已確認 |
| 地圖視窗逐像素**完全相同**（0 個不同的像素，§6.4） | 已確認 |
| 隊伍圖示 ＝ 疊圖編號 **7**；`MASKS.WLF` ＝ 10 張 1-bit 遮罩（§6.4） | 已確認 |
| 原版地圖視窗**最上面一列留黑**，內容從 `y = 9` 起（§6.4） | 已確認 |
| **站著不動時鐘不會走**——主迴圈等按鍵時不回頂端（§6.1） | 已確認 |
| 走一步 ＝ 荒野 4 分鐘（§6.2） | 已確認 |

## 2. 環境

```bash
tools/dosbox.sh                                   # 等 6 秒截一張
tools/dosbox.sh "wait:5;shot:01;key:Return;wait:3;shot:02"
tools/dosbox.sh "<timeline>" "fixed 3000" WLU     # 換執行檔
```

全程 docker（`docker/dosbox/`），headless：Xvfb ＋ xdotool ＋ ImageMagick。
邊界寫在腳本本體——`--rm`、限 CPU／記憶體／pids、log 上限、
原版目錄**不掛進容器**（第一次跑會複製一份可寫副本到 `workplace/dosbox/game/`）。

`msdostest`（567 款實測設定的那個 repo）**沒有 Wasteland**，設定是自己調的，
但照它的原則：

```
core=normal
cputype=386
cycles=fixed 3000      ← 寫死，不用 auto
machine=ega            ← 原版是 mode 0Dh（docs/re/04 §4）
```

⚠ **`cycles=auto` 是可重現性的敵人**：同一串按鍵每次跑到的遊戲內時間點不同，
截圖就對不起來。

兩個 headless 專屬的坑（照抄，不要重踩）：

- 沒有 window manager，`xdotool windowactivate` 會失敗（不支援 `_NET_ACTIVE_WINDOW`）。
  要用 `windowfocus`（`XSetInputFocus`）。
- **不能用 `xdotool key --window <id>`**——那走 `XSendEvent`，SDL 預設不理合成事件，
  按鍵送了等於沒送。要用全域 `xdotool key`（XTest）。

## 3. 解包正確性

`tools/unpack_exepack.py` 的產物（169,488 bytes）複製成 `WLU.EXE` 丟進遊戲目錄跑：

```bash
tools/dosbox.sh "wait:6;shot:02-unpacked" "fixed 3000" WLU
docker run --rm -v "$PWD/workplace/dosbox/shots:/shots" \
  --entrypoint sh wasteland-dosbox:latest \
  -c 'compare -metric AE /shots/01-boot.png /shots/02-unpacked.png null:'
# → 0
```

**0 個不同的像素。** 解包版跑得起來，而且畫面與原版完全一樣。
`docs/re/02` §5 那條「要在 DOSBox 跑起來與原版對照，才能把『解包等同原版』
升為已確認」——升了。

## 4. 逐像素對拍：`TITLE.PIC`

`tools/compare_screen.py` 把截圖的每個像素對到最近的 EGA 顏色，
再與 `tools/decode_pic.py` 的 4-bit 索引比。**位置是掃出來的**：

```
截圖 320 × 200；解碼器 288 × 128
最佳位置 (8, 8)：36864/36864 吻合 = 100.00%
次佳位置 (7, 8) 差 4689 個像素（粗掃）
```

一次驗掉三件事：

1. **XOR delta 解碼逐像素正確**（`docs/re/23` §2）。
2. **圖片視窗在 (8, 8)**——與 `docs/re/25` 從程式碼讀出來的一致，
   而這次是從畫面上**掃**出來的，兩個獨立來源。
3. **調色盤就是預設的 EGA 16 色**。`docs/re/23` §5 掛著「沒找到設定調色盤的
   程式碼，推測是 mode 0Dh 的預設」——索引比對 100% 吻合，推測成立。
   若調色盤被改過，比對不可能全中。

⚠ 工具會印**次佳位置的差距**。差距小於一列寬就代表沒對準——
那時候「吻合率 99%」是假的。這裡差 4,689 個像素（一列才 288），對準無疑。

## 5. 隊伍名單畫面：五條斷言一次印證

`tools/dosbox.sh "wait:5;shot:10;key:Return;wait:3;shot:11;key:Return;wait:3;shot:12"`

```
1)Hell Razor      AC 0  AMM  0  MAX 28  CON 28  WEAPON Crowbar
2)Angela Deth     AC 0  AMM 18  MAX 27  CON 27  WEAPON VP912 9
3)Thrasher        AC 0  AMM  0  MAX 34  CON 34  WEAPON Knife
4)Snake Vargas    AC 0  AMM 18  MAX 31  CON 31  WEAPON VP912 9
Ranger Ctr.                              CREATE DELETE PLAY
```

| 畫面上的東西 | 對應的斷言 |
|---|---|
| 四個名字與 MAX／CON | `docs/re/30` §4 的出廠角色記錄 |
| **`AMM 18`** | **物品表 `+0x04`（容量）** ＝ `VP91Z 9mm pistol` 的 18（`docs/re/45` §3.2）；Crowbar 與 Knife 的容量是 0，畫面上就是 0 |
| `Ranger Ctr.` | 全域狀態 `+0xD0` 的地點名稱（`docs/re/30` §3.2） |
| `You cannot create any more characters.` | `sub_1C6C9` 讀出的那條字串（`docs/re/21` §5） |
| `CREATE`／`DELETE`／`PLAY` 首字母上色 | `\x10` 的每列熱鍵機制（`docs/re/43` §2） |

**`AMM` 那一欄是這一批裡最有價值的一條**：它把「物品表 `+0x04` 是容量」
與「物品陣列的附屬 byte 初值 ＝ 容量」兩條靜態推論一次接到畫面上。
`+0x04` 一開始被讀成「彈匣容量」，後來因為 Match ＝ 40、Rope ＝ 1 才改成
「裝滿時能用幾次」——畫面證實了改對了。

## 6. 地圖畫面

進地圖的按鍵序列：`wait:5 → Return（標題）→ wait:3 → p（名單上的 PLAY）`。
⚠ 在名單上按 Return 會觸發 **CREATE**，跳出「You cannot create any more
characters.」的 modal——那時候按 `p` 會被 modal 吃掉。**直接按 `p`。**

### 6.1 站著不動：時鐘不會走

進地圖後 45 秒不按任何鍵，四張截圖的時鐘全部停在 `01:16`。

所以 `docs/re/26` §1.2 那個問題有答案了：**`ds:46E1h` 的原地踏步與節拍自動步
都只在主迴圈跑一輪時發生，而主迴圈在等按鍵時卡在 `sub_18E90` 的忙等迴圈裡**
（`sub_18EFE` 是非阻塞的 `int 16h ah=1`，`sub_18E90` 在外面自己轉），
根本回不到迴圈頂端。**「站著讓時間流逝」實際上不會發生。**

### 6.2 走一步：時鐘正好 +4 分鐘

`01:24 → 01:28 → 01:32`，每按一次方向鍵前進 4 分鐘——
`docs/re/27` 的「荒野每步 4 分鐘」與規格 04 §8.5 條件 3 實機驗證。

### 6.3 遭遇的訊息印證了兩條斷言

走幾步之後觸發遭遇：

```
From the depths of the wasteland appears a hostile adversary.

1 Humanoid
  appears at 14 feet.
```

- `Humanoid` ＝ 字串表 1 的第 `0x55` 條 ＝ `0x52 + 3`，也就是敵人資料
  `+0x06` ＝ 3（`docs/re/37` §3.2）。**種類欄位實機確認。**
- 走的是 `sub_129E9`「沒有專屬名字就印種類名」那一條路。
- `14 feet` 來自距離表 `ds:CD0Dh`（`docs/re/37` §4，逐格照抄不能用公式）。
- 「1 Humanoid」是單數，沒有加 `s`——`0x0A` 的單複數變形（`docs/re/28`）。

### 6.4 地圖視窗逐像素對拍：**0 個不同的像素**

`cmd/wl-shot -mode play` 從存檔的起始位置畫一幀，與原版同一位置的截圖比
地圖視窗那 288 × 128。過程中修掉兩個缺口：

| 輪次 | 不同的像素 | 抓到什麼 |
|---:|---:|---|
| 1 | 404 | **隊伍圖示沒畫**——差異全部集中在第 (9, 4) 格 |
| 2 | 288 | 全部落在**第 0 列**，橫跨整個寬度 |
| 3 | **0** | — |

**第一個缺口**：`sub_16716` 把 `al ← 7`、座標寫進 `ds:4685h`／`4686h`，
再叫 overlay slot 4（背景 AND 遮罩 OR 疊圖）。所以**隊伍圖示是疊圖編號 7**——
`docs/re/24` 那條「`IC0_9.WLF` 那十個圖形各自是什麼」的第一個答案。
順帶把 `MASKS.WLF` 解了：320 bytes ＝ 10 張 × 32，**一個位元一個像素**
（16 列 × 2 bytes），不是 4 平面。

**第二個缺口**：原版地圖視窗**最上面那一列留黑**，內容從 `y = 9` 開始。
⚠ 這**不是**視窗位置錯了——`TITLE.PIC` 在 (8, 8) 滿 128 列 100% 吻合，
而且我們的第 1–127 列與原版逐像素相同。是**地圖繪製專屬**的一列差別。
重製版用 `MapClip()`（`ViewClip()` 少最上面一列），`DrawPicture` 不套。

## 7. 還沒做的

- 存檔 round-trip 的實機驗收：改寫存檔之後原版還讀不讀得進去。
- 戰鬥、商店、設施畫面的對拍（要先走到那些畫面）。

## 8. 可重跑的完整指令

```bash
# 原版
tools/dosbox.sh "wait:6;shot:01-boot"

# 解包版（先把 tools/unpack_exepack.py 的產物複製成 WLU.EXE）
cp workplace/analysis/unpacked/wl.unpacked.exe workplace/dosbox/game/WLU.EXE
tools/dosbox.sh "wait:6;shot:02-unpacked" "fixed 3000" WLU

# PNG → PPM（比對工具吃 PPM，純 stdlib 讀得動）
docker run --rm -v "$PWD/workplace/dosbox/shots:/shots" \
  --entrypoint sh wasteland-dosbox:latest \
  -c 'convert /shots/01-boot.png -crop 320x200+0+0 +repage /shots/01-boot.ppm'

# 逐像素對拍
python3 tools/compare_screen.py title \
  workplace/dosbox/shots/01-boot.ppm workplace/orig/wastland/title.pic
```

## 9. 這一輪學到的（寫成規則）

- **對拍要把「位置」也當成待驗的東西。** 把偏移寫死的話，「解碼錯」與
  「對錯位置」會長得一模一樣。掃出最佳位置、並印出**次佳的差距**——
  差距不夠大就代表沒對準，那時候的吻合率是假的。
- **一張截圖可以同時驗很多條斷言，但要事先列出你在驗什麼。** 隊伍名單那張
  一次印證了五條；如果只是「看起來對」就過去了，那五條還是沒被驗過。
- **實機能驗的不只是畫面。** `AMM` 是一個數字欄位，它驗的是兩層靜態推論
  （欄位語意 ＋ 初始化路徑）。找對拍目標時優先找**畫面上直接顯示原始欄位值**的地方。
- **對拍不是一次過關的事，是「差異收斂」的過程。** 404 → 288 → 0：
  每一輪的差異**分布**都指向下一個缺口（集中在一格 → 隊伍圖示；
  集中在一列 → 裁切差一列）。所以工具要能報「差在哪裡」，不能只報「差幾個」。
- **靜態讀不出來的一列，實機一眼就看到。** 「地圖視窗最上面一列留黑」
  沒有任何常數寫著它，只有把兩張圖疊起來才會冒出來——
  這正是 `CLAUDE.md` 說「驗收要對原版行為，不是對自己的測試」的意思。
