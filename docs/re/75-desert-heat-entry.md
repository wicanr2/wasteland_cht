# 75：沙漠高溫的入口 —— 腳本 opcode 3 的晝夜分支

日期：2026-08-15 ｜ 接 `docs/re/66`（線索）、`docs/re/67`（參數）、`docs/re/74`（排除）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 見 `docs/re/01`；實機環境見 `docs/re/47`。

跨四份筆記的缺口解開了。

---

## 1. 逐步對拍找到的那一格

實機與 remake 送同一串鍵（`path=32:62` 的 25 步），每 6 步截一張圖：

| 步 | 實機 | remake（修正前） |
|---:|---|---|
| 6 | 沒有訊息 | 沒有訊息 |
| 10–12 | `It is very warm.` × 3 | `SOMETHING HAPPENS.` × 3 |

`SOMETHING HAPPENS.` 是 remake 對「nibble 6 但不是設施」的 fallback，
所以差異在**腳本**。第 10 步踩到的是地圖 0 的 **(47,62)**：

```
nibble 6 記錄 1：01 | 00 00 | 02 07 | 02 0a | 01 00 00 02 08
                 ↑    ↑       ↑       ↑
             section  下一步  分支 A   分支 B
             0x10 的   ＝改寫  nibble 2 nibble 2
             索引 1    對      記錄 7   記錄 10
```

`+0x00` ＝ 1 是 **section `0x10` 的索引**，那個 word 才是 opcode
（`docs/re/71` §5.3）——查出來是 **3**。

## 2. opcode 3（`0x1A526`）：晝夜分支

```
0x1A526  bl ← 5
         al ← ds:465Ah                 ; 時鐘的「時」
         al < 6      → loc_1A535       ; 夜間，bl 保持 5
         al ≥ 12h(18) → loc_1A535      ; 夜間
         bl ← 3                        ; 6 ≤ 時 < 18 ＝ 白天
loc_1A535:
         al ← 記錄[bl]；push
         bl++；al ← 記錄[bl]
         記錄[2] ← 那個 byte
         記錄[1] ← push 的那個
         clc; retn                     ; ← CF ＝ 0：**繼續跑**
```

把選中的一對搬進 `+0x01`／`+0x02`，也就是**下一步**。
而 `sub_12C70` 收尾的 `sub_169B1(1)` 用位移 1 改寫這一格——
`docs/re/71` §5.1 說過，這台直譯器的程式計數器就是地圖格本身，
所以「換下一步」與「改寫這一格」是同一個動作。

| 時段 | 用哪一對 | 換成 | 扣多少 |
|---|---|---|---|
| 06:00–17:59 | `+0x03`／`+0x04` | nibble 2 記錄 **7／8／9** | 2／4／6 顆 d6 |
| 18:00–05:59 | `+0x05`／`+0x06` | nibble 2 記錄 **10／11／12** | 1／2／3 顆 d6 |

**兩組三階段就是白天與夜晚。** `docs/re/67` §2 只看到參數由弱到強，
沒看出分組的理由——理由在這裡。

`DawnHour ＝ 6`／`DuskHour ＝ 18` 是 `docs/re/27` §5 早就解出的晝夜門檻，
opcode 3 用的是同一組常數。

## 3. `CF ＝ 0` 表示「這一步還沒完」

opcode 3 回 `clc`，所以原版**當場繼續跑**改寫後那一格的事件——
於是同一步就印出高溫訊息並扣血。remake 原本一步只跑一個事件，
症狀是「踩上去只說 SOMETHING HAPPENS.，走開再走回來才變熱」。

`internal/game/world.go` 的事件處理改成迴圈（上限 `maxEventChain ＝ 8`，
正常資料碰不到），腳本回報 `Continue` 就再跑一次 `trigger`。

## 4. `fd fd fd fd`：跑完改回原樣

記錄 7–12 的 `+0x04`–`+0x07` 全是 `fd`（沿用改寫前的值，`docs/re/69` §7），
所以條件閘收尾會把那一格**改回 nibble 6 記錄 1**。
下一步再踩又重跑一次——**高溫是每走一步都算的，不是一次性事件**。

這也解釋了為什麼記錄 7–12 在初始地圖上一格都沒有：
它們是每一步當場貼上去、跑完撕掉的。

## 5. 對拍結果

```
tools/go.sh run ./cmd/wl-play -script "path=32:62" -trace
  10 (47,62) 01:40  It is very warm.
  11 (46,62) 01:44  It is very warm.
  12 (45,62) 01:48  It is very warm.
  13 (44,62) 01:52  It's getting warmer.
```

實機同一串鍵在同一步印同一句（`docs/re/75` 的 `seg2.png`：
時鐘 02:36、訊息視窗三行 `It is very warm.`）。
開場 01:00 是夜間，所以走的是記錄 10–12 那一組——與實機一致。

`TestDesertHeatViaScript`。

## 6. 順帶修正

- **`Script` 是死碼**：直譯器 44 個 opcode 早就實作好也測過，
  但生產程式碼從來沒呼叫過它，`nibble 6` 的事件只印一句 fallback。現已接上。
- **`Script.Step()` 把 `Record[0]` 當 opcode**：那是 section `0x10` 的索引。
  新增 `NewScript` 查表，查不到時 `Handled ＝ false`（不當成 nop）。
- **nibble 2 記錄 6 是「水」不是沙漠**：`+0x03` ＝
  `As calm as the water is here, you have trouble keeping your head above water.`，
  條件是技能 7，沒過每步扣 1。

## 7. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/op3.json 0x1A526

tools/go.sh test ./internal/play/ -run TestDesertHeatViaScript -v
tools/go.sh run ./cmd/wl-play -script "path=32:62" -trace

# 實機逐段截圖（每 6 個按鍵一張）
TL=$(tools/go.sh run ./cmd/wl-play -script "path=32:62" -emit-keys | grep '^timeline: ' | sed 's/^timeline: //')
# 把 TL 每 6 個 key 插一次 shot 之後送進 tools/dosbox.sh
```

## 8. 這一輪學到的（寫成規則）

- **兩邊送同一串鍵、逐段截圖，差異會自己跳出來。** 找了四輪的入口，
  在「第 10 步實機說 warm、remake 說 SOMETHING HAPPENS」那一刻就結束了。
  **有 oracle 的時候，先做逐步對拍再做靜態掃描。**
- **`SOMETHING HAPPENS.` 這種 fallback 是啞掉的錯誤。** 它讓遊戲跑得動、
  測試全綠，代價是一整類內容安靜地不發生。
  **fallback 要能被統計**——如果 remake 記錄了「這一步走了 fallback」，
  第一次跑地圖就會看到它出現幾百次。
- **`Write` 會覆蓋既有檔案。** 這一輪用 `Write` 建 `internal/game/script.go`，
  砍掉了已經寫好的 44 個 opcode 實作，靠 `git checkout` 救回。
  **要新增內容到可能已存在的檔案，先 `Read` 或 `ls`。**
