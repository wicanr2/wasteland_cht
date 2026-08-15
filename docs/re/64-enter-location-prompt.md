# 64：進新地點的確認（第三道閘的一半）

日期：2026-08-15 ｜ 起因是實機對拍：**原版問了一句，remake 直接走進去**

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
實機環境見 `docs/re/47`。

`docs/re/26` §3 把第三道閘記成「`sub_16AD5` → CF（nibble 10 且記錄
`+0x00` 的 bit7 ＝ 1 時走詢問流程）」。**bit 記錯了，而且流程沒讀。**

---

## 1. `sub_16AD5` 全文

```
0x16AD5  call sub_169EB          ; al ← 目標格的 nibble
0x16AD8  al ≠ 0Ah → 不問
0x16ADE  al ← 記錄 +0x00
0x16AE4  shl al, 1
0x16AE6  test al, al
0x16AE8  jns → 不問              ; ← **原本的 bit6** 設起來才問
0x16AEA  sub_16B17；sub_19720
0x16AF0  al ← 67h；sub_16CB2     ; **印字串 0x67**
0x16AF5  call sub_12619          ; 問 Yes／No
0x16AF8  CF ＝ 1（選 No）→ loc_16B06
0x16AFA  ds:AAD3h ← 1            ; 選 Yes
0x16AFF  ds:AAD4h ← 0；clc; retn ; 可以走

loc_16B06（選 No）：
  sub_163C4
  ds:AAD3h ← 0；ds:AAD4h ← 1
  stc; retn                      ; 擋住
```

**判準是 bit6 不是 bit7。** `shl al, 1` 之後 `jns` 看的是移位後的符號位，
也就是原本的 bit6。Quartz 入口的記錄 `+0x00` 是 `0x41`（`0100 0001`）——
bit6 設、bit7 沒設，所以會問。

字串 `0x67` 是執行檔字串表 1 的第 103 條：

```
\rEnter new location?\r\r\x11\x10Yes\r\x10No\r
```

`\x10` 是熱鍵標記（`docs/re/14` §4.1），所以熱鍵是 **Y** 與 **N**。
與實機畫面逐字相同。

## 2. 兩個旗標接回 `sub_16A10`

`ds:AAD3h`／`ds:AAD4h` 就是傳送那一支開頭檢查的東西（`docs/re/60` §2）：

| 旗標 | 誰設 | 傳送那邊怎麼用 |
|---|---|---|
| `ds:AAD4h` | 選 No 時 ← 1 | `0x16A10` 一開頭：非 0 就整支不做 |
| `ds:AAD3h` | 選 Yes 時 ← 1 | `0x16A1A`：＝ 0 時先跑 `sub_16B17` |

**答 No 之後那一步整個不算**：不移動、不推進時鐘、不掃遭遇。

## 3. 實機對拍

同一串 51 步（`cmd/wl-play` 的 `-emit-keys` 直接產出 timeline）：

| | 畫面 |
|---|---|
| 原版 | `Enter new location?` ／ `Yes` ／ `No`，訊息 `Entering Quartz.` |
| 修正前的 remake | **沒問，直接進 Quartz** |
| 修正後 | 停在原地問，答 Y 才進去、答 N 留在原地 |

## 4. 對 remake 的意思

規則層多一個狀態：`StepResult.Ask` 非 0 時**這一步停在原地等回答**。
呈現層記住原本按的方向，收到 Y 就 `World.Confirm()` 再走一次。

⚠ **答 No 不是「被擋住」**：被擋住會印那一格的訊息（`docs/re/62`），
答 No 什麼都不印，畫面回到走路狀態。

## 5. 還沒讀的

- `sub_12619`（問 Yes／No 的那一支）與 `sub_16B17`。
- `sub_163C4`（選 No 時跑的）。

## 6. 可重跑的完整指令

```bash
# 產生 timeline 並送進 DOSBox
TL=$(tools/go.sh run ./cmd/wl-play -script "path=29:43" -emit-keys \
  | grep '^timeline: ' | sed 's/^timeline: //')
tools/dosbox.sh "wait:6;key:Return;wait:4;key:p;wait:5;$TL;wait:3;shot:quartz2"

tools/go.sh test ./internal/play/ -run TestEnterLocationAsksAndNoStays -v
```

## 7. 這一輪學到的（寫成規則）

- **`shl` 之後的 `jns` 看的是移位前的前一個 bit。** 筆記寫成 bit7 是漏了
  那一條 `shl`。**位元判斷要連著移位一起讀**，單看 `jns` 會差一位。
- **「走過去」與「走過去之前問一句」在畫面上差很多，在程式裡只差一道閘。**
  這種缺口只有把同一串按鍵送進兩邊才看得出來——
  remake 那邊完全正常，只是少了一句話。
- **驗證工具要能產出實機的輸入。** `-emit-keys` 讓「remake 走的那條路」
  直接變成 DOSBox 的 timeline，兩邊跑的是同一串鍵；
  沒有它，每次對拍都要手工湊按鍵，而湊錯了會以為是行為不同。
