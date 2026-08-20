# 128：`ds:471Fh` 是「這一行的序號要反白」——三個進入點共用一個名片行

日期：2026-08-20 ｜ 接 [`111`](111-roster-inverse-video.md)（反白旗標 `ds:4678h`）、
[`125`](125-roster-box.md)（名單框與行號）、[`117`](117-save-globals-and-facility-screen.md) §2（設施畫面）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

實機截圖上，站在櫃檯前的那個人**行首序號是反白的**
（`workplace/dosbox/shots/43-menu.png` 的 `1>`），而問「誰要進去？」的前一張
（`42-shop.png`）一格反白都沒有。這一份把那個反白的來源讀完：
它不是名單繪製的一部分，是一個獨立的旗標 `ds:471Fh`，
由三個進入點各自設定，再落到同一支名片行繪製。

---

## 1. 一個本體，三個進入點

全檔只有四處碰 `ds:471Fh`（`export_range_refs.py`，半開區間 `0x471F 0x4720`）：

```
0x17029  b3 ff        mov  bl, 0FFh          ; ← sub_17029：反白**開**
0x1702B  e9 09 00     jmp  loc_17037
0x1702E  b3 00        mov  bl, 0             ; ← sub_1702E：反白**關**
0x17030  e9 04 00     jmp  loc_17037
0x17033  8a 1e 1f 47  mov  bl, ds:471Fh      ; ← sub_17033：**沿用上一次**
0x17037  88 1e 1f 47  mov  ds:471Fh, bl      ; 共同本體從這裡開始
0x1703B  ds:B242h ← al                       ; 要畫的是第幾個人
0x1703E  call sub_19614                      ; 依編號設角色記錄位址
0x17041  ds:46B9h ＝ 0（地圖模式）→ retn      ; 名單沒顯示就不用畫
0x1704A  ds:4654h ≠ ds:B1E0h → jmp sub_1728C ; 這一組換過了 → 整個畫面重來
0x17057  call sub_19083                      ; 游標存檔
0x1705A  ds:4675h ← 0x26；ds:4674h ← 1        ; 行寬 38、起始欄 1
0x17064  ds:4673h ← 編號 ＋ 0x0F              ; 列 ＝ 序號 ＋ 15（docs/re/125）
0x1706C  ds:4672h ← 1                         ; 欄 1
0x17071  al ← 3；call sub_19DC3               ; 清掉那一列
0x17076  編號 ≤ ds:4653h（人數）→ call sub_1708B ; 畫那一行
0x17085  call sub_19060；jmp sub_1785E        ; 游標還原、收尾
```

⚠ **`0x17033`／`0x17037` 那一對看起來是原地寫回，其實不是。**
`mov bl, ds:471Fh` 緊接 `mov ds:471Fh, bl` 單獨讀是一條沒有作用的指令，
而它之所以存在，是因為**另外兩個進入點跳進第二條**：三支公用同一個本體，
差別只有 `bl` 從哪裡來。這是 16-bit 常見的寫法，
`0x1702B` 與 `0x17030` 兩個 `jmp loc_17037` 就是證據。

推論等級：**已確認**（三個進入點與本體逐條讀完）。

## 2. 反白的範圍只有序號與 `>`

`sub_1708B`（畫一行的本體，`docs/re/103`）開頭：

```
0x17094  ds:4678h ← ds:471Fh       ; 反白旗標 ← 這一行的設定
0x1709A  印序號
0x170A0  印 `>`（0x3E）
0x170A5  call sub_19E30            ; **反白關**
```

`ds:4678h` 是文字繪製的反白旗標（`docs/re/111`：非 0 就走
`lodsb ／ not al ／ stosb`）。旗標在**印完序號的下一條指令就被關掉**，
所以畫面上反白的只有序號那幾格，名字與其他欄位照常。
`43-menu.png` 上的 `1>` 是一塊白底黑字、`Hell Razor` 是正常的白字，兩者吻合。

推論等級：**已確認**（程式碼順序 ＋ 實機截圖）。

## 3. 誰會開

六個呼叫端，每一個都先把「是誰」寫進自己那個變數，再叫 `sub_17029`：

| 位址 | 編號來自 | 意思 | 佐證 |
|---|---|---|---|
| `0x1200A` | `ds:A432h` | **戰鬥指令階段正在下令的那個人**。同一段在 `0x12004` 先 `sub_17357` 重畫整張名單（順手把旗標清 0），再用反白版重畫他那一行 | [`38`](38-combat-commands-and-flee.md) §1 |
| `0x13C55` | `ds:A5E1h` | `USE` 指令選中的隊員 | [`92`](92-use-command.md) |
| `0x19133` | `ds:CCEDh` | 選人之後的金錢畫面（§4） | 本份 |
| `0x1BBFE` | `ds:DAA5h` | 訓練師：來學技能的那個人 | [`52`](52-trainer-facility.md) §2 |
| `0x1BF18` | `ds:DBF5h` | 商店：走到櫃檯前的那個人 | [`117`](117-save-globals-and-facility-screen.md) §2.1 |
| `0x1C2B5` | `ds:DCE1h` | 醫生：就診的那個人 | [`119`](119-doctor-and-trainer-entry.md) §1 |

六處的形狀完全相同，所以這個旗標的語意是統一的：
**「現在輪到／選中的是這個人」**，與「這一格有問題」那兩種反白
（`docs/re/111`：狀態位元反白 `MAX` 欄、卡彈反白武器名）互不干涉。

## 4. 誰會關

| 位址 | 動作 |
|---|---|
| `0x17362` | `sub_17357`（重畫整張名單）開頭 `call sub_19E30` 之後 `ds:471Fh ← al`。`sub_19E30` 是 `mov al, 0` 開頭的，所以這裡寫進去的是 **0** ——**整張重畫一定不帶反白** |
| `0x1701C` | `sub_16F70`（畫名單框，`docs/re/125`）用 `sub_1702E` 逐行畫 1–7 列 |
| `0x191FF` | `sub_19130` 離開時用 `sub_1702E` 把那個人重畫一次 |

`sub_19130(al ＝ 隊員編號)` 是**選完人之後的那個小迴圈**：

```
0x19130  ds:CCEDh ← al；call sub_17029      ; 反白他的序號
0x19139  ds:7DF3h ← 0x0C04；ds:7DF5h ← 6    ; 熱區遮罩（docs/re/43）
0x19145  call sub_18E90                      ; 等按鍵
         0x0D／空白 → 離開
         'P' → sub_19B81（集中金錢，docs/re/117 §3），回頭
         'D' → 0x19165：取這個人的錢、清成 0、逐一發還給每個成員
0x191FC  al ← ds:CCEDh；jmp sub_1702E        ; 離開前把反白關掉
```

兩個呼叫端（`0x1276F`、`0x1A421`）都先拿按鍵碼 `ds:4667h & 0x0F` 當編號，
所以入口是「在名單上按 `1`–`9` 選一個人」。

推論等級：`sub_19130` 的按鍵分支**已確認**（整支讀完），
`'D'` 那一段的語意標**強證據**（指令序列是「取錢 → 清零 → 逐人發還」，
但沒有實機對過）。

## 5. remake 這一側

`RosterRow.IndexInverse` ＋ `Scene.selectedMember()`：戰鬥取
`CombatScene.Turn`、設施取 `FacilityScene.Who()`。
`Who()` 在「誰要進去？」那一步回 `false`——那時候還沒有人被選中，
與 `42-shop.png` 一致。反白範圍由 `InverseAt` 算，
與另外兩個旗標共用同一組欄座標。

## 6. 可重跑的完整指令

```bash
# 誰碰 ds:471Fh（半開區間，單一位址要寫 lo 與 lo+1）
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_range_refs.py \
  workplace/analysis/dumps/ref471F.json 0x471F 0x4720

# 三個進入點與本體
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/f17033.json 0x17033

# 誰叫「反白開」／「反白關」（離線掃全檔指令，不必再開 IDA）
python3 - <<'PY'
import json
ins = json.load(open('workplace/analysis/dumps/listing.json'))['instructions']
for off, name in ((0x7029, '反白開'), (0x702E, '反白關')):
    for i in ins:
        if any(o['kind'] in ('near', 'far') and int(o['addr'], 16) == off
               for o in i['ops']):
            print(name, i['ea'], i['disasm'])
PY
```

實機對照：`workplace/dosbox/shots/42-shop.png`（沒有人反白）
與 `43-menu.png`（`1>` 反白）。

## 7. 這一輪學到的（寫成規則）

- **看到「讀一個變數又原封不動寫回去」，先找有沒有別的進入點跳進第二條指令。**
  16-bit 的程式碼常用「三個入口、一個本體」省空間，而中間那個入口
  單獨讀起來像是無用指令。判它多餘之前，先掃一次誰跳到那個位址。
- **旗標的語意要由「誰寫它」決定，不是由「它長得像什麼」決定。**
  `ds:471Fh` 與 `ds:4678h` 都是反白，但一個是「選中的人」、一個是
  「這一格有問題」，六個寫入端的形狀一致才定得下來。
- **一個 UI 細節缺席時，實機截圖是最快的入口。** 這條路是從
  「remake 的名單沒有反白的序號」開始的，比從函式清單找快得多。
