# 72：進地點的完整路徑，與地圖指令列

日期：2026-08-15 ｜ 接 `docs/re/71`（批次改寫）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
實機環境見 `docs/re/47`。

三輪靜態排除之後改用實機，第一步就把「Ranger Center 怎麼進去」整條解開了。

---

## 1. 出廠存檔就站在設施格上，而站著不會觸發

出廠存檔的位置是地圖 0 的 **(55, 62)**，而那一格是 **nibble 6 記錄 0**
（設施 `Ranger Ctr.`，32 筆設施裡有格子指到的兩筆之一）。

實機開場畫面是世界地圖，不是設施畫面——**事件只在踩上去的那一步觸發**，
被存檔放在那裡不算踩。

## 2. 走開再走回來：nibble 12 先把那一格換掉

```
tools/dosbox.sh "wait:6;key:Return;wait:4;key:p;wait:6;key:i;wait:3;shot:step1;key:k;wait:4;shot:step2"
```

| 步驟 | 實機畫面 |
|---|---|
| `i`（上）→ (55,61) | 沒有訊息 |
| `k`（下）→ (55,62) | `Enter new location?` ／ `Yes` ／ `No`，訊息 `Entering Ranger Center.` |

(55,61) 是 **nibble 12 記錄 2**，它的批次表第一筆是：

```
+0x01: c0 37 3e 0a 27
       flags ＝ 0xC0（bit7 最後一筆、bit6 跳過重畫、bit0 ＝ 0 相對原點）
       x ＝ 0x37 ＝ 55、y ＝ 0x3E ＝ 62
       新值 ＝ nibble 10、記錄 39
```

**踩上 (55,61) 就把 (55,62) 從「設施」改寫成「傳送格」。**
所以完整路徑是：

```
nibble 6（設施記錄，初始狀態）
  ↓ 走過旁邊的 nibble 12 → 批次改寫
nibble 10 記錄 39（傳送 ＋ 記錄 +0x00 的 bit6 → 問一句）
  ↓ 踩上去 → Enter new location? → Yes
傳送進 Ranger Center 的內部地圖
```

remake 走同一串鍵得到同一個結果（`TestRangerCenterEntry`）——
這同時驗證了 `docs/re/71` 的批次改寫實作。

## 3. 設施 3 ＝ Ranger Center 的角色管理選單

`Ranger Ctr.` 記錄的 `+0x00` ＝ `0x83`，跳表索引 **3** ＝ `0x1A2C0`：

```
0x1A2C0  sub_18801
         記錄 +0x03 起 13 bytes → ds:7201h      ; 設施名稱
         sub_1A308                              ; 選單：ds:468Dh ← 0xCE42、
                                                ;       ds:4689h ← 2（兩組）
         al ← 3；sub_190A6；sub_1728C；sub_19727 ; 畫面
         sub_18E90 → sub_17574                  ; 讀鍵 → 清單選單處理
         jmp loc_1A2F4                          ; 迴圈
```

`ds:CE42h` 的內容是 `0xCE12`，而 `ds:CE12h` ＝ **`CREATE DELETE PLAY`**。

所以設施 3 是**建立／刪除角色與開始遊戲**的畫面，不是通用的店面框架。
`ds:A4E0h` 跳表的五個設施因此是：

| 索引 | 位址 | 是什麼 |
|---:|---|---|
| 0 | `0x1C260` | 醫生（`Infirmary`、`Patch em' up`、`Old Doc Bobs`、`NUCLEAR AID`） |
| 1 | `0x1BE50` | 商店（`AG. store`、`Store`、`Market`、`Blackmarket`） |
| 2 | `0x1BBA0` | 圖書館／訓練（`Library`、`New Thoughts`、`HOLY KNOWING`） |
| 3 | `0x1A2C0` | **Ranger Center 的角色管理**（`CREATE DELETE PLAY`） |
| 4 | `0x1B4F0` | `sub_1CB30` 計時到期寫死呼叫的那一個 |

## 4. 地圖指令列

`sub_16C7C` 設的選單表 `ds:AB18h`，第 0 組指到 `ds:A9CCh`：

```
"Use Enc Order Disband View Save Radio\0IKJL..."
```

實機底部的指令列逐字相同：`USE ENC ORDER DISBAND VIEW SAVE RADIO`。
`IKJL` 是方向鍵（I 上、K 下、J 左、L 右）。

按 `U` 之後畫面問 `Which player?` 並列出隊伍名冊——
**`USE` 是「用某個角色的技能或物品」，不是進店的指令**。

選單框架的三個欄位（`docs/re/53` 的清單框架共用）：

| 欄位 | 意義 |
|---|---|
| `ds:468Dh` | 選項字串位址表（`sub_17564`：`ds:4680h ← [表 + ds:468Ah × 2]`） |
| `ds:468Fh` | 第二張表（`sub_17574` 的 `0x17655` 用） |
| `ds:4689h` | 組數（指令列 6、設施 3 是 2、`sub_1630C` 的 `Start` 是 1） |

## 5. 商店與醫生：問題的形狀變了

`Ranger Ctr.` 的格子**初始就是 nibble 6**，而 nibble 12 把它改成傳送格。
其餘 30 筆設施沒有初始格子，掃過的五條改寫來源也沒有指向它們
（`docs/re/71` §5）。

新的著手點是**反過來問**：既然改寫可以把設施格換掉，
那也可以把別的格子換成設施格——而 nibble 12 的批次表是唯一能改寫
「非腳下」格子的來源。先前掃它得到「目標是 nibble 6 的 32 筆、設施 0 筆」，
但那次掃描對 `flags & 1`（相對隊伍座標）那一族只做了寬鬆的合理性檢查。
**下一步是把相對座標那一族單獨列出來逐筆看**，它們的目標編號不受地圖邊界限制，
先前的過濾可能把真的命中濾掉了。

## 6. 可重跑的完整指令

```bash
tools/go.sh test ./internal/play/ -run TestRangerCenterEntry -v

tools/dosbox.sh "wait:6;key:Return;wait:4;key:p;wait:6;shot:before;key:u;wait:3;shot:use1"
tools/dosbox.sh "wait:6;key:Return;wait:4;key:p;wait:6;key:i;wait:3;shot:step1;key:k;wait:4;shot:step2"

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/menu2.json 0x1630C 0x16C7C 0x173D2 0x17564
```

## 7. 這一輪學到的（寫成規則）

- **靜態掃三輪不如實機走兩步。** 「Ranger Center 怎麼進去」的答案是
  「旁邊那一格先把它改成傳送格」——這在資料上完全看得出來，
  但**沒有人會想到去查隔壁那一格**。`docs/re/60` §8 從第一輪就寫了要用實機，
  我繞了三輪才照做。**排除法連續失敗兩次之後就該換 oracle。**
- **「初始狀態」不是「唯一狀態」。** 統計「幾筆設施有格子指到」時，
  兩邊都在動：格子會被改成設施，設施格也會被改成別的。
  單看初始地圖的統計對這種系統沒有意義。
- **推翻自己上一輪的假說要當場改寫，不要留著。** `docs/re/71` §5.5 曾把
  設施 3 猜成「通用選單，商店與醫生從那裡分派」——`ds:CE12h` 是
  `CREATE DELETE PLAY` 就否定了它。**假說寫進文件時要標推論等級，
  被推翻時連同推論一起改掉。**
