# 48：`IC0_9.WLF` 十張疊圖各自是什麼

日期：2026-08-15 ｜ 對應盤點 **A12**（`IC0_9.WLF`／`MASKS.WLF`）、**A5**（地圖）、
補上 `docs/re/24` §2.3 留的那一題

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`ic0_9.wlf` SHA-256 `d8bbeae054a25852817841b905ae093fb104587e89d90749fb8fd9ec6ca38ddc`；
`masks.wlf`、`game1`／`game2` 的 SHA-256 見 `docs/re/01`。

---

`docs/re/24` §2.3 把這張表的**位置與筆數**算清楚了（`seg003:0x0420`，10 × 128 bytes，
4 平面 16 × 16，遮罩在 `seg003:0xDA60`），但編號 0–9 各是什麼沒有答案——
**程式碼裡只有編號，沒有名字**。這一份把十個編號全部綁定到語意。

方法是兩邊夾：**誰傳這個編號**（全檔掃出四個呼叫點）＋ **它長什麼樣**
（把檔案畫出來看）。單靠任何一邊都不夠：只看圖會變成猜圖，只看程式碼會停在
「這裡傳了一個 4」。

## 1. 結論

| 編號 | 語意 | 圖形（`tools/dump_icons.py` 畫出來的） | 等級 |
|---:|---|---|---|
| 0 | **把一格塗黑**（遮罩全 0、圖形全 0） | 全黑 | 已確認 |
| 1 | 敵人種類 5 ＝ `Robot` | 淺藍白色的機械人形 | 已確認 |
| 2 | 敵人種類 4 ＝ `Cyborg` | 紅橘裝甲人形 | 已確認 |
| 3 | 敵人種類 2 ＝ `Mutant` | 綠色佝僂人形 | 已確認 |
| 4 | 敵人種類 3 ＝ `Humanoid`，**也是查不到種類時的預設** | 戴白頭盔、青衣桃紅褲的人 | 已確認 |
| 5 | **nibble 5 的格子**：寶箱／掉落物 | 灰色布袋 | 已確認 |
| 6 | 敵人種類 1 ＝ `Animal` | 桃紅色四足生物 | 已確認 |
| 7 | **隊伍自己** | 綠衣橘髮的人 | 已確認 |
| 8 | **nibble 9 的格子**：輻射區，**只在夜間畫** | 黃色輻射三葉標誌 | 已確認 |
| 9 | **其他分隊**（`DISBAND` 拆出去的那幾組） | 桃紅色人字形 ＋ 青色方塊 | 已確認 |

`MASKS.WLF` 與這十張一一對應：320 bytes ＝ 10 × 32，**一個位元一個像素**
（16 列 × 2 bytes），不是 4 平面。合成式是 `螢幕 ← (背景 AND 遮罩) OR 疊圖`，
所以**遮罩為 0 的像素背景被清掉、螢幕上就等於疊圖本身**——第 4 節的實機比對
靠的就是這一點。

⚠ **0 號不是「不畫」，是「塗黑」。** 它的遮罩全 0、圖形全 0，合成之後那一格
變成純黑。地圖第 3 層有 448 格的值是 0，畫出來就是黑色地形。

## 2. 四個呼叫點（全檔掃描，不是抽樣）

疊圖只有一個入口：overlay slot 4（`0x1000C` → `sub_1029B`）。掃它的
`CodeRefsTo` 得到**四個**呼叫點，全部列在這裡：

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_call_context.py \
  workplace/analysis/dumps/icon-callers.json 0x1000C 0x1029B --before 20
```

| 呼叫點 | 所在函式 | `al`（疊圖編號）從哪來 |
|---|---|---|
| `0x14816` | `sub_14664`（遭遇掃描） | `ds:AA17h[敵人種類]`——見 §3 |
| `0x16488` | `sub_16428` | 常數 **9** |
| `0x16756` | `sub_16716` | 常數 **7** |
| `0x180C9` | `sub_18024`（依 nibble 畫一格） | 常數 **5**／常數 **8**／記錄 `+0x01` |

`ah`（背景）四處都一樣：地圖第 3 層那一格的 byte。

### 2.1 `sub_18024`：依第 1 層的 nibble 分三條路

```
0x1803C  call sub_17C20        ; al ← 第 1 層 nibble
0x1803F  cmp  al, 5   / jz 18082 → mov al, 5     ; 寶箱／掉落物
0x18043  cmp  al, 4   / jz 1806E                 ; 記錄 +0x01
0x18047  cmp  al, 9   / jnz 18088                ; 其餘 → 當一般圖磚
0x1805D  mov  al, ds:465Ah     ; 時（docs/re/27）
0x18060  cmp  al, 6   / jb  18068
0x18064  cmp  al, 12h / jb  18088                ; 6 ≤ 時 < 18 → 不畫
0x18068  mov  al, 8                              ; 其餘時段 → 輻射標誌
```

nibble 4 那條路：

```
0x1806E  call sub_180CF        ; ds:46AEh ← 這一格的記錄
0x18071  mov  bl, 1
0x18077  mov  al, [bx+di]      ; 記錄 +0x01
0x18079  and  al, 7Fh
0x1807B  cmp  al, 0Ah
0x1807D  jb   180A5            ; < 10 → 疊圖
0x1807F  jmp  18098            ; ≥ 10 → 當 ALLHTDS 圖磚畫（不套遮罩）
```

**`< 10` 這個門檻就是 `IC0_9.WLF` 的張數。** 編號空間與地圖第 3 層同一套：
0–9 是這十張，≥ 10 是圖磚（編號 ＝ 值 − 10，`docs/re/24` §2.3）。

`sub_167CE` 每走一步重畫 nibble 為 **4／5／9** 的格子（`docs/re/26` §5）——
那三種正好就是這裡的三條路。**「會動的格子」與「會畫疊圖的格子」是同一組。**

## 3. `ds:AA17h`：敵人種類 → 疊圖編號

`sub_14664`（遭遇掃描，`docs/re/39`）取到一筆遭遇之後：

```
0x147D3  mov  bl, 3            ; 預設 3（Humanoid）
0x147D5  cmp  al, 0
0x147D7  js   loc_147E6        ; bit7 設起來 → 用預設，不去讀記錄
0x147D9  call sub_12A4C
0x147DC  mov  bl, 6
0x147DE  mov  di, ds:4665h
0x147E2  mov  al, [bx+di]      ; al ← 敵人記錄 +0x06 ＝ 種類（docs/re/37 §3.2）
0x147E4  mov  bl, al
0x147E6  mov  al, [bx-55E9h]   ; al ← ds:AA17h[種類]
```

線性 `0x27837`（檔案位移 `0x178E7`），原始 bytes `00 06 03 04 02 01`：

| 種類 | 名稱（字串 `0x52 + 種類`） | 疊圖 |
|---:|---|---:|
| 0 | —（資料裡不存在） | 0 |
| 1 | `Animal` | 6 |
| 2 | `Mutant` | 3 |
| 3 | `Humanoid` | 4 |
| 4 | `Cyborg` | 2 |
| 5 | `Robot` | 1 |

42 個區塊 397 筆敵人資料的 `+0x06` **全部落在 1–5**（`docs/re/37` §3.2），
所以這張表六項就夠用，第 0 項是佔位。表後面還有四個 `01`——那是 6–9 的
填充值，資料到不了。

**圖與名稱互相印證**：6 是四足生物、3 是佝僂的綠色人形、4 是普通人、
2 是紅色裝甲、1 是白色機械——五個都對得上，而且**這五個對應是表決定的，
不是我照圖排的**。

## 4. 實機驗證（DOSBox，`docs/re/47` 的環境）

判準：遮罩為 0 的像素在畫面上必須**逐像素等於疊圖本身**（§1 的合成式）。
底下是什麼地形完全不影響，所以「疊圖畫在哪」可以脫離地圖單獨驗。

```bash
python3 tools/compare_screen.py icons <截圖.ppm> \
  workplace/orig/wastland/ic0_9.wlf workplace/orig/wastland/masks.wlf
```

| 驗的是 | 怎麼做 | 結果 |
|---|---|---|
| 7 ＝ 隊伍 | 進地圖截圖 | (144, 64) **1 處**吻合，正是隊伍那一格 |
| 9 ＝ 其他分隊 | `DISBAND` → 選 2 號 → 往北拆出去 | (144, 80) 出現 **1 處**；主隊再往北一步變 (144, 96) |
| 8 ＝ 輻射區 | 往北走進輻射帶（02:44） | 畫面上 34 處，訊息 `The ground seems to glow here.` |
| 8 只在夜間 | 原地來回踏步等到天亮 | **05:56 有、06:00 沒有**，同樣那幾格 |
| 6 ＝ Animal | 荒漠遊走遇到遭遇接近 | (128, 64) 出現 1 處，就在隊伍旁邊 |

重製版這一側跑同一支工具（`cmd/wl-shot -mode play -at 57,39 -hour 2`／`-hour 12`）：
夜間 **32 處**吻合編號 8、白天 **0 處**，格線對齊與原版相同。
判準是同一個，所以兩邊的數字可以直接對照。

「05:56 有、06:00 沒有」直接對上 `cmp al, 6`：**畫面在跨過 06:00 的那一步
翻面，位置一格沒動**。這種前後對照比單張截圖強——它排掉了「剛好那張圖沒畫」。

⚠ **遭遇是 RNG 驅動的，不可重跑。** 同一串按鍵、同樣 `cycles=fixed 3000`，
第一次跑出遭遇、第二次沒有——亂數種子來自計時器節拍（`docs/re/13`）。
畫面內容（標題、地圖、時鐘）可重現，**遭遇不可**；靠「再跑一次」來重現遭遇
會白等。要驗遭遇相關的東西就多跑幾輪並接受隨機性。

## 5. nibble 9 ＝ 輻射區

疊圖 8 是輻射三葉標誌，而它只畫在 nibble 9 的格子上。反過來查資料：
六張有 nibble 9 的地圖，那些格子指到的訊息**全部與輻射有關**：

| 資源 | 格數 | 訊息 |
|---:|---:|---|
| 0 | 36 | `The ground seems to glow here.` |
| 19 | 60 | `A wave of heat washes over you and the ground shimmers.`／`Your eyes tear and a ringing fills your ears.`／`Your whole body feels ready to combust, and your skin itches mightily.` |
| 20 | 44 | `The floor is covered with Plutonium dust.`／`On the floor is an open container of U235.`／`A pile of Wastelandium glows brightly here.` |
| 26 | 54 | 字串編號 0（不印訊息） |
| 31 | 1 | `A broken radioactive waste container is lying here buried in the trash.` |
| 38 | 16 | 字串編號 0（不印訊息） |

nibble 9 的處理函式是 `0x14410`：印記錄 `+0x00` 指的訊息，然後**對每個隊員
擲 `+0x01` 顆 d6 扣 CON，並加上 Radiation poisoning 的狀態位元**。
整條結算與 `ds:46EFh`（這一次結算跳不跳過護甲吸收）在
[`55`](55-radiation-and-armour-bypass.md)。

推論等級：nibble 9 ＝ 輻射區 **已確認**（六張地圖的資料、圖形、實機訊息三邊一致）。

## 6. 資料面的統計

`tools/summarize_icons.py` 的輸出在
[`generated/ida94/icons.md`](generated/ida94/icons.md)：

- 第 3 層值落在 0–9 的格子：**457 格**（0 號 448、1 號 1、2 號 3、7 號 5），
  與 `docs/re/24` §2.3 獨立算出來的數字相同。
- 會動的三種 nibble：**nibble 4 有 218 格、nibble 5 有 43 格、nibble 9 有 211 格**。
- nibble 4 的 218 格裡，記錄 `+0x01` **< 10 的只有 15 格**（0 號 2、1 號 1、7 號 12），
  其餘 197 格是圖磚，6 格的記錄查不到。

「nibble 4 的格子大多不是疊圖」這件事很容易看反：`+0x01` 直接當疊圖編號讀
會讀到 32、41、96 這種值，看起來像壞掉的資料，其實只是圖磚編號。
**門檻是程式碼寫的（`cmp al, 0Ah`），不是我設的。**

## 7. 可重跑的完整指令

```bash
# 疊圖與遮罩畫成一張圖（上：疊圖／中：遮罩／下：合成示意）
python3 tools/dump_icons.py workplace/orig/wastland/ic0_9.wlf \
  workplace/orig/wastland/masks.wlf workplace/shots/icons.png

# 四個呼叫點與各自的前 20 條指令
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_call_context.py \
  workplace/analysis/dumps/icon-callers.json 0x1000C 0x1029B --before 20

# sub_18024 / sub_16428 / sub_14664 完整反組譯
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/iconfuncs.json 0x18024 0x16428 0x14664

# 資料面統計
python3 tools/summarize_icons.py workplace/analysis/unpacked/wl.merged.exe \
  workplace/orig/wastland/game1 workplace/orig/wastland/game2 \
  docs/re/generated/ida94/icons.md

# 實機：拆隊 → 其他分隊的標記
tools/dosbox.sh "wait:6;key:Return;wait:3;key:p;wait:4;key:d;wait:2;type:2;\
wait:1;key:Return;wait:2;key:Up;wait:3;shot:grp1"

# 實機：走進輻射帶，等到天亮（截圖橫跨 06:00）
#   20 步向北 → 40 步原地來回 → 24 步原地來回並逐步截圖
python3 tools/compare_screen.py icons workplace/dosbox/shots/dawn13.ppm \
  workplace/orig/wastland/ic0_9.wlf workplace/orig/wastland/masks.wlf
```

## 8. 這一輪學到的（寫成規則）

- **「這個編號是什麼」要兩邊夾。** 只讀程式碼會停在「這裡傳了一個 4」，
  只看圖會變成猜圖。程式碼給對應關係、圖給名字，兩邊獨立才互相印證得了。
- **編號空間常常是共用的。** 疊圖 0–9 與圖磚 ≥ 10 是**同一個編號空間**，
  分界寫在程式碼的 `cmp al, 0Ah` 裡。不套那個門檻就會把圖磚編號讀成壞資料。
- **實機驗證要找「同一位置的前後對照」，不要只截一張。** 「05:56 有、06:00 沒有」
  排掉了「剛好那張沒畫」；單張「有輻射標誌」證不了「只在夜間」。
- **RNG 驅動的畫面不可重跑。** 固定 `cycles` 保證的是**指令節拍**可重現，
  不保證**亂數**可重現——種子來自計時器。把「再跑一次就會一樣」當前提會白等。
- **反過來查資料可以把語意釘死。** 疊圖 8 是輻射標誌 → 去看所有 nibble 9 的
  格子指到什麼訊息 → 六張地圖全是輻射。這比「讀處理函式的四條指令」強得多，
  因為它涵蓋了全部資料而不是一條路徑。
