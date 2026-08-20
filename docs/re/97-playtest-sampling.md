# 97：抽樣試玩（第二輪）—— 七段流程各走一遍

日期：2026-08-16 ｜ 接 `docs/re/59`（第一輪，正常玩家路徑）、`WORKLIST.md` 7.5

驗收方式是使用者定的（`WORKLIST.md` 開頭）：**不從頭玩到結局**，
改成挑代表性的地點與流程各走一段。這一份記走了哪七段、每一段拿什麼當
「原版會怎樣」的依據、看到什麼。

一句話結論：**七段裡有六段一走就撞到缺口**，而且沒有一個會讓測試變紅——
它們全都是「編得過、測得過、玩不動」那一類。

---

## 1. 方法與可重現指令

無頭跑 `cmd/wl-play`，每一步印出狀態。這一輪為了看得到東西補了幾樣：

| 補了什麼 | 為什麼 |
|---|---|
| `Scene.Mode()` | 訊息列看不出「進了設施但選單沒開」，模式看得出來 |
| `Scene.CJK()` ＋ `lang.FromBig5` | trace 把中文印在〔〕裡，中文化的洞當場看得到 |
| `wl-play -lang／-font／-refs／-paragraphs` | 原本只跑英文，驗不到中文那條路 |
| `wl-play -poll N` | 見 §5.1：不推進亂數時「沒遇到敵人」不成立 |
| `wl-play -save-dir／-modified` | 存檔重開要寫得出去、再讀得回來 |

```bash
tools/go.sh run ./cmd/wl-play -script "up,down,Y,C,W,A,N,G,enter,P" -trace
tools/go.sh run ./cmd/wl-play -script "map=10:30:25,up,Y,B,esc,S,esc" -trace
tools/go.sh run ./cmd/wl-play -poll 11 -script "@12:2,right,left,…（128 次）,A,…" -trace
tools/go.sh run ./cmd/wl-play -script "map=2:25:6,up" -trace
```

---

## 2. 七段流程

| # | 流程 | 結果 | 依據 |
|---|---|---|---|
| 1 | Ranger Center 建角色 | 進得去、**選單叫不出來** → 已修 | `docs/re/72` §3 |
| 2 | 出城遇敵打一場 | 第 128 步遭遇、四人下令、打完給經驗值 ✔ | `docs/re/51`、`78` |
| 3 | 進商店買賣 | 開得了店、**按買回報「背包滿了」、全館 $0** → 已修 | `docs/re/42`、`22` |
| 4 | `USE` 開一道條件閘 | 三層選單全中文、判定會失敗也會成功 ✔ | `docs/re/92` |
| 5 | 讀一段段落 | 資源 2 (25,5) 石英城酒館 → 第 23 段中文正文 ✔ | `docs/re/33`、`docs/spec/19` |
| 6 | 存檔重開 | **按了 `Save` 檔案沒動** → 已修 | `docs/re/30` §6 |
| 7 | 走進 Base Cochise 觸發結局 | **資料裡沒有觸發點**（§4.2） | `docs/re/96` |

第 2 段的細節：地圖 0 的 (12, 2) 一帶來回走，第 128 步跳出
`YOU ARE BEING ATTACKED!`，四個人各按 `A`，兩回合打完，
`The party gains 56 experience.` 回地圖。中間的命中、未命中、經驗值
逐條印得出來——**規則層是通的，通的部分全是英文**（§4.1）。

---

## 3. 這一輪修掉的

### 3.1 物品陣列讀到 0 就停 → 買不了東西、賣完重讀掉東西

`readSlots` 一遇到編號 0 就 `break`，於是出廠 Ranger 的 `Items` 長度是 15
而不是 30。兩個後果：

- `FirstEmptyItemSlot` 掃到切片長度為止，找不到空槽 → 商店回報
  「Your inventory is full.」，**任何一家店都不能買東西**。
- 賣掉中間一件（`Sell` 把那一槽清成 `Slot{}`）之後寫回記錄再讀出來，
  洞後面的物品整批消失。

原版是**固定 30 槽、0 ＝ 空，而且中間就是會有洞**：
`docs/re/15` 的 `+0xBD`–`+0xF8` ＝ 30 × 2；`docs/re/42` §3
賣掉是「把那兩個 byte 清成 0」、不搬動後面的；同一份 §3.1 的繪製回呼
明寫 `al ＝ 0` → 這一列跳過。而角色記錄 `+0x1F`／`+0x25` 存的是**槽號**，
往前搬會讓裝備指到別件東西。

修法：`readSlots` 讀滿 30 格、`FirstEmptyItemSlot` 上界改成 `ItemSlots`、
`UseItem` 清成 0 不搬動、`LearnSkill`／`Buy`／`GiveStartingKit` 改用
第一個空槽。門檻：`TestBuyListOpensWithRoomLeft`、`TestSoldSlotSurvivesReload`。

### 3.2 折價指數 0 ＝ 全館免費

`ShopPrice(base, 0)` 算 `base − (base >> 0)` ＝ 0。原版 `sub_1C1CC` 是
**`dl ＝ 0` 直接 return**、右移迴圈根本不跑，`docs/re/22` §3 的比例表
也寫著 n ＝ 0 → 實付 100%。

⚠ 這一條原本有測試守著，而**測試把誤讀寫成了斷言**（`{100, 0, 0}`，
註解還寫「原版的 0 就是全免」）。修的時候要連測試一起改，
否則改對了會變紅。

### 3.3 Ranger Center 的 `CREATE`／`DELETE`／`PLAY` 走不到

`enterRosterIfNeeded` 寫好了，**零個呼叫端**；`beginRoster` 只有測試直接叫。
所以走進 Ranger Center 之後 C／D／P 三個鍵通通沒反應——
而 `WORKLIST` 5.4 標「完成」，單元測試也全綠。

修法：掛在 `EnterFacility` 這個**進場的唯一入口**（兩條路——踩到 nibble 6、
傳送收尾改寫腳下——都經過它）。

### 3.4 `Save` 只改記憶體，沒有寫檔

`cmdSave` 呼叫 `StoreTo` 之後就報「Game saved.」。檔案一個 byte 都沒動，
下一次開遊戲什麼都沒保留。

修法：`assets.(*Rom).WriteSave`（只換 `Offset` 起那一段，其餘照抄）＋
`Scene.SetSaveDir` ＋ `cmd/wasteland -save-dir`。**沒給可寫目錄就照實說**，
不報「存好了」。門檻：`TestSaveCommandWritesToDisk`（寫檔 → 重開 →
位置與時鐘接得上）、`TestSaveWithoutDirSaysSo`。

寫入端解在 `docs/re/30` §5.1（`sub_18801`）：**序號每次 ＋1**，
寫回哪一份由磁碟旗標 `ds:9168h` 決定——也就是目前這張地圖住在哪一個檔案。
remake 照抄序號那一半（`Save.BumpSerial`，`WriteSave` 每次先做），
寫回的檔案則**固定是讀進來的那一份**（重製決策，理由在該節）。

### 3.5 設施進場是空的，要先按一個鍵選單才出現

`refresh` 只在 `Key()` 裡叫。原版主迴圈一進去就印
「Do you want to: Buy / Sell」（`docs/re/42` §1）。
修法：`EnterFacility` 對商店／醫生／訓練師先 refresh 一次。

### 3.6 清單印編號不印名字

商店與訓練師的清單列的是 `item 43`、`skill 5`。名稱表早就解過
（`docs/re/17` §4，`Scene.itemName`／`skillName` 也已經在 `USE` 那條路上用著），
只是 `FacilityScene` 拿不到。修法：進場時把兩支查名函式掛上去。

### 3.7 `USE` 收工之後中文選單留在畫面上

`applyUse` 換掉 `message` 卻沒清 `cjk`，結果技能清單一直疊在結果那一行底下。

---

## 4. 沒修、留給下一輪的

### 4.1 畫面上還有一大片英文（最大的一塊）

中文化的**文本**是完整的（4,806 條 ＋ 162 段），洞在**接線**：

| 哪裡 | 現況 |
|---|---|
| 戰鬥全程 | `YOU ARE BEING ATTACKED!`、`X gains N experience.`、`X misses.`、`The party gains N experience.` 都是 Go 字面值 |
| 指令結果 | `Hell Razor uses Brawling. It fails.`、`Game saved.`、`HQ: nothing to report.` |
| 角色管理 | `CREATE DELETE PLAY (C/D/P)`、`Name:`、`X joins the Rangers.` |
| 設施選單 | `Do you want to: Buy / Sell`、`PRICE ITEM`、`Your inventory is full.` |
| 原版字串表 | `exeString`（表 1）三個消費端**不查翻譯目錄**——`Enter new location?YesNo` 就是這樣一直是英文 |

前四類是重製版自己的介面文字，該走 `ui:` 那組 key
（`translations/{source,zh-Hant}/ui.tsv` ＋ `lang_coverage_test.go` 的
`uiCatalogueKeys`）。第五類是原版字串，譯文已經在目錄裡，只差查表。

### 4.2 結局在資料裡沒有觸發點

`TestFacilityCoverage` 掃 42 張地圖的 nibble 6 記錄：醫生 7、商店 7、
訓練 8、角色管理 1，**結局（kind 4）0 筆**。`docs/re/96` 解的是
程式那一側（跳表第 4 格 → `0x1B4F0`），資料那一側**誰把格子指到 kind 4
還沒定位**。remake 目前只有 `EnterFacility` 會叫 `BeginEnding`，
所以玩家走不到結局；`TestEndingPlaysThrough` 是直接呼叫的。

下一個入口：17 個沒有格子指到的腳本 opcode（`docs/re/76`）裡有沒有
「開結局」那一個；或 nibble 12 的批次改寫（`docs/re/71`）會不會生出這種記錄。

### 4.3 店家庫存不進存檔

賣一件會讓店家庫存 `+1`（`docs/re/42` §3），remake 記在
`shopState.Stock`，離開設施就丟。原版的物品表**在存檔區、每個存檔槽一份**
（`docs/re/45` §2），所以庫存是跟著存檔走的遊戲狀態。

### 4.4 全隊倒下之後遊戲照走

走進地圖 0 北邊的輻射帶，三步之內四個人 CON 全部歸零（這一段本身
與 `docs/re/59` 一致：輻射帶團滅是原版行為），**然後還可以繼續按方向鍵撞山**。
原版在全隊倒下時怎麼處理還沒讀。

### 4.5 兩個小的

- 醫生的治病清單印 `status bit 3`：八個狀態位元的名字在 `docs/re/35` §1
  有，但還沒接到清單上。
- 進設施之後訊息列停在 `TELEPORT.`，地點名只出現在設施畫面那幾行。

---

## 5. 這一輪學到的（寫成規則）

### 5.1 無頭驗收預設不推進亂數，所以「沒發生」不是結論

`Scene.PollRNG` 是熵的唯一來源（`docs/re/13`：產生器初值全零、
靠鍵盤輪詢推進），無頭工具刻意不叫它以保可重現。代價是序列退化成
固定的前幾項——第一次量到「走 49 步一次遭遇都沒有」時，那個數字
**同時相容於「遭遇壞了」與「抽樣方式壞了」**。加了 `-poll 11` 之後
第 128 步就打起來了。

規則：**無頭跑遭遇、寶箱、戰鬥這類靠擲骰的東西，先給 `-poll`**；
不給就不准寫「沒有觸發」。

### 5.2 「有測試守著」不等於那條路走得到

這一輪的六個缺口全部發生在**有測試的子系統**上：角色管理有
`roster_test.go`（直接叫 `beginRoster`）、商店有 `shop_test.go`
（直接叫 `Key`）、存檔有 round-trip 門檻（只驗編碼、不驗寫檔）。
每一個測試都從**中間**進場，於是「玩家怎麼走到這裡」沒有人驗。

規則：**接線測試要從玩家的第一個按鍵開始**，不從場景中間的函式開始。
`TestRangerCenterEntry` 就是對的形狀（走一步、走回去、答 Y），
它只是停在「進得去」而沒有再按一個 `C`。

### 5.3 測試會把誤讀鎖起來

`{100, 0, 0}` 那一列不只是沒抓到 bug，它**把 bug 寫成了規格**，
還附了一句自信的註解。而 `docs/re/22` §3 的比例表就在旁邊，
兩者互相矛盾了不知道多久。

規則：**斷言要引用出處那一行，不要只寫結論**。引用寫上去之後，
「這條與 RE 對不上」是讀得出來的；只寫 `{100, 0, 0}` 讀不出來。

### 5.4 資料模型抄形狀，不要抄「看起來夠用的形狀」

出廠存檔的物品欄剛好連續 15 格，所以「讀到 0 就停」在所有既有測試裡
都成立，直到有人賣掉中間一件。原版的形狀是**固定 30 槽 ＋ 空洞**，
證據在 `docs/re/15` 與 `docs/re/42` §3.1（`al ＝ 0` → 跳過）——
兩份都早就寫著。

規則：**陣列型欄位一律照原版的固定長度讀滿**，空值當資料不當結尾。
