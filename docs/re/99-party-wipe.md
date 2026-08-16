# 99：全隊陣亡怎麼處理 —— Grim Reaper 那一格

日期：2026-08-16 ｜ 接 `docs/re/98` §4（那一節的結論由本份取代）、
`docs/re/89`（倒下 vs 死）、`docs/re/93`（`View`）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

---

## 1. 結論

原版對「這一組全倒了」有**兩條**處置，走哪一條看倒得多重：

| 情形 | 原版做什麼 |
|---|---|
| 這一組還有人站著 | 什麼都不做，照常玩 |
| 全倒但還救得回來 | **自動切到下一支隊伍**（`sub_160A8` ＝ `View`） |
| 全倒而且每個人都到底 | **死亡畫面**：地點名換成 `Grim Reaper`、換一張圖、印 `Your life has ended in The Wasteland.`、等按鍵 |

死亡畫面**不是**遊戲內的重新開始——原版沒有讀檔選單（`docs/re/95`），
重來要離開遊戲、跑 `SETUP.EXE` 選第 4 項「Restart game with old characters」
（IBM 版補充說明書，`docs/manual-cht/ch07`）。

## 2. 死亡畫面（`0x1C570`）

```
0x1c570  b200          mov  dl, 0
0x1c572  8bf2          mov  si, dx
0x1c574  8a8460de      mov  al, [si-21A0h]     ; ← ds:DE60h
0x1c578  8bf2          mov  si, dx
0x1c57a  88840172      mov  [si+7201h], al     ; → ds:7201h（地點名）
0x1c57e  fec2          inc  dl
0x1c580  80fa0d        cmp  dl, 0Dh            ; 13 bytes
0x1c583  75ed          jnz  short loc_1C572
0x1c585  b03b          mov  al, 3Bh
0x1c587  e81ccb        call sub_190A6          ; 載入 ALLPICS 第 0x3B（59）張
0x1c58a  e8ffac        call sub_1728C
0x1c58d  e897d1        call sub_19727          ; 訊息區設定（docs/re/40 §2）
0x1c590  b86dde        mov  ax, 0DE6Dh
0x1c593  a38046        mov  ds:4680h, ax
0x1c596  e8d5b2        call sub_1786E          ; 印字串（docs/re/14）
0x1c599  c706f37d0000  mov  word ptr ds:7DF3h, 0
0x1c59f  c706f57d0100  mov  word ptr ds:7DF5h, 1   ; 滑鼠熱區遮罩（docs/re/43）
0x1c5a5  e8e8c8        call sub_18E90              ; 等按鍵
```

兩段文字都是**明文 ASCII**，直接讀映像就看得到：

| 位址 | 內容 |
|---|---|
| `ds:DE60h` | `Grim Reaper ` ＋ `\0`（13 bytes） |
| `ds:DE6Dh` | `Your life has ended in The Wasteland.\r` |

`ds:7201h` 是**地點名**（存檔 `+0xD0`，13 bytes 明文，`docs/re/30`）。
寫進去這一格的還有角色管理（`0x1A2CF`）、無線電（`0x1B8C2`）、
訓練師（`0x1BBB4`）、商店（`0x1BE72`）、醫生（`0x1C274`）——
每個畫面把地點名換成自己的招牌。**死亡畫面就是同一個形狀的第六張**
（圖 ＋ 地點名 ＋ 一句話），只是它不掛在設施跳表上，是手寫的一段。

推論等級：**已確認**（逐指令讀完，兩段文字從映像直讀）。

## 3. 誰跳到那裡（`0x16C2B`）

```
0x16c2b  a05646      mov  al, ds:4656h      ; 這一組的人數
0x16c2e  84c0        test al, al
0x16c30  742f        jz   loc_16C61         ; 0 人 → 死亡畫面
0x16c32  b001        mov  al, 1             ; i ← 1
0x16c34  a2e0aa      mov  ds:0AAE0h, al
0x16c37  e8ea29      call loc_19624         ; 選第 i 個角色（ds:46B5h ← 0x7131 + i×256）
0x16c3a  e88106      call loc_172BE         ; CF ＝ 1 ⇔ CON ≤ 0（倒下）
0x16c3d  7325        jnb  loc_16C64         ; 這個人還站著 → 離開迴圈
0x16c3f  e8db2d      call sub_19A1D         ; dl ← 傷勢等級
0x16c42  80fa00      cmp  dl, 0
0x16c45  750c        jnz  loc_16C53         ; 有傷勢等級 → 看下一個人
0x16c47  b328        mov  bl, 28h           ; 記錄 +0x28 ＝ 八個狀態位元
0x16c49  8b3eb546    mov  di, ds:46B5h
0x16c4d  8a01        mov  al, [bx+di]
0x16c4f  84c0        test al, al
0x16c51  7411        jz   loc_16C64         ; 沒有狀態 → 離開迴圈
0x16c53  …           i ＋ 1，還有人就回 0x16C34
0x16c61  e90c59      jmp  loc_1C570         ; 全部走完 → 死亡畫面
0x16c64  e8a730      call sub_19D0E         ; 還有沒有人站著
0x16c67  7503        jnz  loc_16C6C
0x16c69  e83cf4      call sub_160A8         ; ← 沒有 → View：切到下一支隊伍
```

`sub_19D0E` 逐指令讀完了：它掃這一組每個人，`sub_172BB` 回 CF ＝ 0（站著）
就把 `ds:CD43h` ＋1，最後 `test al, al; retn`——**ZF ⇔ 一個人都沒站著**。
所以 `0x16C67` 的 `jnz` 跳過 `View`，落下來的那一條就是「全倒 → 換隊」。

推論等級：**已確認**（跳轉與被呼叫的三支都逐指令讀完）。

### 3.1 逐人判準還沒對完

迴圈要走完（＝ 死亡畫面）的條件是每個人都「倒下 ＋ 傷勢等級 ≠ 0
或狀態位元 ≠ 0」。`sub_19A1D` 的形狀讀出來了——`dl` 從 5 起，
`CON ＝ 0` 直接回 5，否則拿 CON 低位與 `ds:CCCDh` 起那張門檻表比、
逐次 `dec dl`，掉出迴圈就是 0——但**哪一個 `dl` 對應哪一條傷勢帶**
還沒與 `docs/re/15` §3 的 −11／−20／−30／−40 逐值對過。

推論等級：**強證據**（分支結構確定，語意對照未完成）。
下一個入口：`ds:CCCDh` 起四個 byte 的值。

## 4. 手冊怎麼說

官方英文手冊（`docs/manual/06-command-summary.md`，描述的是 Apple II 流程）：

> What can you do if a character dies? DO NOT ENTER A NEW LOCATION OR SAVE THE GAME!
> Turn off your computer and reboot, and your character will live again, but without
> anything they acquired since you last saved. If all the characters die in the midst
> of general carnage and mayhem, your computer will state the obvious:
> **“Your life in Wasteland is over.”**

DOS 版的實際字串是 `Your life has ended in The Wasteland.`——**措辭不同、機制相同**。
手冊在這裡是**次級 oracle**（`CLAUDE.md` §5 的優先序）：它指出「有這麼一件事」，
確切內容以執行檔為準。

手冊同時說明了為什麼遊戲內沒有重新開始的入口：救回角色的方法是
**不要存檔、關機重開**，靠的是「存檔只在你主動存或換地點時發生」。

## 5. remake 現在缺什麼

| 項目 | 現況 |
|---|---|
| 全倒自動換隊 | **沒有**。remake 全倒之後就停在地圖上，什麼都不會發生 |
| 死亡畫面 | **沒有**。圖 59 載得到、地點名欄位也在，只是沒人叫 |
| `SETUP.EXE` 的重來 | 不做——那是 DOS 安裝程式，不是遊戲的一部分 |

兩項都要先收攏成規格（G2）再動工，列在 `WORKLIST.md`。

## 6. 可重跑的完整指令

```bash
# 死亡畫面那一段與跳進去的迴圈
python3 - <<'PY'
import json, bisect
D=json.load(open('workplace/analysis/dumps/listing.json'))
INS=D['instructions']; EAS=[int(i['ea'],16) for i in INS]
def show(lo,hi):
    i=bisect.bisect_left(EAS,lo)
    while i<len(EAS) and EAS[i]<hi:
        x=INS[i]; print(f"{EAS[i]:#07x}  {x['bytes']:<12} {x['disasm']}"); i+=1
show(0x16C2B,0x16C6C); show(0x1C570,0x1C5A8)
PY
```

兩段文字用 `Rom.DsBytes(0xDE50, 0x90)` 直接讀映像。

## 7. 這一輪學到的（寫成規則）

- **「掃過了、沒有」要先問掃的是不是全部的語料。** 上一輪掃九張打包字串表
  找不到任何「遊戲結束」，還做了正對照（同一次掃描找得到 ` died!`）——
  正對照過了，結論還是錯的，因為**正對照與目標在同一個不完整的語料裡**。
  `CONTEXT.md` §2 早就寫著「只有少數介面字串以明文 ASCII 存在執行檔裡」。
  **正對照要放在懷疑的那個邊界外面**：要證明「執行檔裡沒有這句話」，
  對照組就得是一句已知的明文 ASCII 字串，不是另一條打包字串。
- **手冊是找「有沒有這件事」的好起點，不是找「怎麼做」的來源。**
  一句 “your computer will state the obvious” 就把整條路指出來了，
  而確切字串與流程仍然只能從執行檔讀。
- **同一個形狀出現第六次，就該去找它。** 圖 ＋ 地點名 ＋ 一句話 ＋ 等按鍵，
  五個設施都是這個形狀；死亡畫面是第六個，只是沒掛在跳表上。
  **找「還有誰寫 `ds:7201h`」比找字串快。**
