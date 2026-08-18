# 113：片頭播的是六頁開場字幕，而且有一句永遠播不到

日期：2026-08-18 ｜ 接 `docs/re/95`（主選單與開機序列）、`docs/re/17`（打包字串表）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/95` §2 把 `sub_161C0` 的後半記成「attract mode：選單等不到按鍵就一直播」，
播什麼沒有解。它播的是**第 0 張字串表**（開場字幕與製作名單）的六頁。

---

## 1. 進到片頭要按兩次鍵，不是等逾時

```
0x16207  call sub_1630C          ; 主選單，尾呼叫 sub_173D2 等按鍵
0x1620A  call sub_173D2          ; **再等一次**
0x1620D  call sub_166D3          ; 逐 byte 送 ds:AA4Dh 的序列
0x1623F  loc_1623F:              ; ← 片頭迴圈的頭
…
0x162C4  jmp loc_1623F           ; 無限循環
```

`sub_1630C` 的最後一行是 `jmp sub_173D2`，所以主選單的按鍵等待就是
`0x16207` 這一次。按 `S` 走索引 0 的處理程式進地圖迴圈；按別的鍵走索引 1
（`0x1632F` 只有一個 `retn`）回到 `0x1620A`，在那裡**再等一次按鍵**，
然後才進片頭。片頭跑起來之後沒有出口——`0x162C4` 無條件跳回頭。

⚠ **不是閒置逾時觸發的。** 「attract mode」是形容它的樣子，
不是它的觸發條件。

推論等級：**已確認**。

## 2. 一頁的三個動作

```
0x16385  sub_16385(al):
  push ax
  call sub_1785E        ; ds:B265h ← 9E53h
  pop  ax
  call sub_16390        ; ds:4692h ← 0A703h   ← **字串表基址**
  jmp  sub_17923        ; 印第 al 條
```

`ds:4692h` 是打包字串表的基址（`docs/re/17` §3）；`0xA703` 那一張是
**開場字幕與製作名單**，20 個槽。所以 `sub_16385(n)` ＝「印第 0 張表的第 n 條」。

## 3. 播出順序

`sub_162C7` ＝ 清畫面（`ds:46FEh ← 0`，`jmp sub_17A6B`）。
把 `0x1623F`–`0x162C4` 照著讀下來：

| 頁 | 清畫面 | 字串槽 | 之後 |
|---:|:---:|---|---|
| 1 | ✓ | 1 | `sub_162CF(3)` |
| 2 | ✓ | 2 | `sub_162CF(3)` |
| 3 | ✓ | 3 | `sub_162CF(3)` |
| 4 | ✓ | 5、6、7 | 5 與 6 之後各 `sub_18DB4(0Ah)`，7 之後 `sub_162CF(3)` |
| 5 | ✓ | 9、10、11、12 | `sub_162CF(3)` |
| 6 | ✓ | 13、14、15、16、17 | `sub_162CF(0Ah)` |

內容（`ExeStrings()` 第 0 張，`tools/extract_strings.py` 解出來的原文）：

| 槽 | 原文 |
|---:|---|
| 1 | `Electronic Arts and / Interplay Productions / proudly present` |
| 2 | `WASTELAND! / Copyright 1986-88 by Interplay Productions.` |
| 3 | `Written by Alan Pavlish / IBM version by Michael Quarles` |
| 5 | `Place : EARTH` |
| 6 | `Year  : 1998` |
| 7 | `Status: DEFCON 1` |
| 9–12 | `Diplomatic solutions to the world's problems fail and war erupts as some madmen press ahead with their insane dreams.` |
| 13–17 | `Current condition: / High concentrations of radiation produce random storms and mutations. / Somehow life continues in the Wasteland!` |

推論等級：**已確認**（15 個 `sub_16385` 呼叫端逐一列出，值與順序照反組譯）。

### 3.1 槽 8 印不出來

第 0 張表的槽 8 是 ` Computer defense initiative activated.`——
一句完整的句子，而**全檔 15 個 `sub_16385` 呼叫端沒有一個傳 8**。
槽 4、18、19 也沒有：4 與 19 是空的，18 是解碼雜訊（在
`translations/untranslatable.tsv` 裡）。

所以槽 8 是**寫了但沒接上的一句台詞**。這不是掃描有洞：
`sub_16385` 是這張表唯一的入口（`sub_16390` 是全檔唯一寫
`ds:4692h ← 0A703h` 的地方，`docs/re/17` §3 的九張表就是這樣列出來的）。

推論等級：**已確認**。

## 4. 兩種等法

```
0x162CF  sub_162CF(dl):
  ds:0A6E3h ← dl        ; 傳進來的 3 或 10 存這裡
  dl ← 0FFh             ; **實際的等待長度是 255**
  call sub_162E1

0x162E1  sub_162E1(dl):
  ds:0A6E2h ← dl
  ds:7DF3h ← 8；ds:7DF5h ← 0        ; 熱區範圍（滑鼠）
loc_162F1:
  call sub_18EFE；call sub_17574     ; 收鍵盤／滑鼠
  ax ← ds:0D150h
  cmp ax, ds:0A6E6h；jz $-           ; **等計時器變一次**
  ds:0A6E6h ← ax
  dec dl；jnz loc_162E1
```

`sub_162CF(dl)` 的參數不是等待長度——它被存進 `ds:A6E3h`，
真正傳給 `sub_162E1` 的是寫死的 `0xFF`。所以**每一頁都是 255 個計時器刻**。
`ds:D150h` 是遊戲的計時器計數（地圖迴圈的 `0x16B65`／`0x16B85` 也讀它），
照結局那一段用的同一個換算（BIOS 一刻 ＝ 1/18.2 秒，`docs/re/96` §4）
是每頁約 14 秒。

`sub_18DB4(dl)` 是另一回事：`dl` 次的空轉迴圈（255 × 255 的 `sbb`），
沒有計時器也沒有鍵盤——第 4 頁那三行「Place／Year／Status」之間的短停頓。
它的長度**跟著 CPU 速度跑**，在現代機器上等於沒有。

推論等級：**已確認**（三支常式讀完）；
「一刻 ＝ 1/18.2 秒」是**強證據**——`ds:D150h` 的寫入端在中斷處理裡，
靜態掃不到，這裡沿用 `docs/re/96` §4 對同一個計數的換算。

## 5. 對 remake 的意思

這六頁是遊戲的開場故事，而且**整張表都在翻譯目錄裡**（`exe:0:*`）。
重製版原本只有標題圖與 `Start` 一個選項，這一段從來沒有出現過。

接法在 `internal/play/attract.go`：標題畫面按下非 `S` 鍵就開始播，
六頁照原版的順序與分頁，播完回到第一頁循環；播放中按 `S` 照樣開始遊戲。

與原版的差異，逐條寫在那個檔案的檔頭：

| | 原版 | 重製版 | 為什麼 |
|---|---|---|---|
| 進入 | 按兩次非 `S` 鍵 | 按一次 | 第二次等待在畫面上沒有任何提示，看起來像沒反應 |
| 版面 | 清整個畫面印文字 | 同樣清掉標題圖，文字從第 0 列起 | 一樣 |
| 每頁長度 | 255 刻（約 14 秒）| 同 | 照原版 |
| 第 4 頁的行間停頓 | `sub_18DB4`，跟 CPU 速度跑 | 固定 18 刻（約 1 秒）| 原版那個值在現代機器上是 0，等於三行同時出現 |

## 6. 可重跑的完整指令

```bash
docker run --rm --network none --log-opt max-size=10m --log-opt max-file=3 \
  -v "$PWD":/w -w /w -u "$(id -u):$(id -g)" python:3.12-slim \
  python3 tools/scan_addr_refs.py workplace/analysis/dumps/listing.json 0xA703 0xA705

tools/go.sh test ./internal/play/ -run TestAttract -v
```

## 7. 這一輪學到的（寫成規則）

- **「等不到按鍵就播」這種描述要回頭對呼叫圖。** 片頭的觸發是兩次按鍵，
  不是逾時；把行為的**樣子**寫成觸發**條件**，下一個人會照著實作一個
  原版沒有的計時器。
- **參數不一定是它看起來的那個參數。** `sub_162CF(3)` 的 `3` 沒有進等待迴圈，
  等待長度是常式裡寫死的 `0xFF`。**追到值真正被用到的那一行**再說它是什麼。
- **表裡有槽不代表遊戲印得出來。** 槽 8 是完整的一句台詞而沒有任何呼叫端。
  翻譯覆蓋率算得到它，玩家永遠看不到——**「有翻譯」與「玩得到」要分開量**。
