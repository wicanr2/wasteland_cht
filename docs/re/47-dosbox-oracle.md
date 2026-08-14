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

## 6. 還沒做的

- **`ds:46E1h` 觸發原地踏步的頻率**（`docs/re/26` §1.2）：要進到地圖上、
  不按任何鍵、看時鐘與遭遇會不會自己動。要先走完 `PLAY` 之後的流程。
- 地圖畫面的逐像素對拍（規格 04 §8.5 的第二半）：要先能穩定進到地圖。
- 存檔 round-trip 的實機驗收：改寫存檔之後原版還讀不讀得進去。

## 7. 可重跑的完整指令

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

## 8. 這一輪學到的（寫成規則）

- **對拍要把「位置」也當成待驗的東西。** 把偏移寫死的話，「解碼錯」與
  「對錯位置」會長得一模一樣。掃出最佳位置、並印出**次佳的差距**——
  差距不夠大就代表沒對準，那時候的吻合率是假的。
- **一張截圖可以同時驗很多條斷言，但要事先列出你在驗什麼。** 隊伍名單那張
  一次印證了五條；如果只是「看起來對」就過去了，那五條還是沒被驗過。
- **實機能驗的不只是畫面。** `AMM` 是一個數字欄位，它驗的是兩層靜態推論
  （欄位語意 ＋ 初始化路徑）。找對拍目標時優先找**畫面上直接顯示原始欄位值**的地方。
