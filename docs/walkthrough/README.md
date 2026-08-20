# 荒野遊俠攻略

這份攻略不是翻譯來的，是從**遊戲資料本身**長出來的。

每一條「這扇門要開鎖或撬棍」「這句密語是 MUERTE」「這裡有一家商店叫石英城商場」，
都來自 `game1`／`game2` 兩個資料檔裡的地圖記錄，由 [`cmd/wl-atlas`](../../cmd/wl-atlas)
倒出來、[`tools/summarize_walkthrough.py`](../../tools/summarize_walkthrough.py) 整理成表。
所以這裡不會出現「我記得好像要先跟某人講話」這種二手記憶，也不會有抄來的錯誤。
反過來說，資料看不出來的東西這裡也不會寫——**沒把握的事寧可留白**。

## 怎麼讀

正文按地區分章，每一章講三件事：怎麼進去、裡面有什麼、走得出哪裡。
窮舉的清單放在 [`generated/`](generated/)，四份：

| 檔案 | 內容 |
|---|---|
| [`generated/maps.md`](generated/maps.md) | 42 張地圖與它們之間的所有傳送 |
| [`generated/gates.md`](generated/gates.md) | 每一格擋路的東西，以及過得去的條件 |
| [`generated/passwords.md`](generated/passwords.md) | 所有問答與收得下的答案 |
| [`generated/facilities.md`](generated/facilities.md) | 商店、醫生、訓練師與地圖腳本 |

座標一律寫成 `(x, y)`，原點在地圖左上角。地圖用**資源編號**指稱
（`game1`／`game2` 裡的 MSQ 區塊編號），因為那是資料裡唯一穩定的識別碼——
同一個地名在不同層會對到不同的資源。

## 劇透

謎題的答案、密碼、機關的解法**全部直接寫出來**。這份東西的用途是查，不是導覽。
不想被劇透就只看每一章的「怎麼進去」那一段。

## 這些數字怎麼來的

```bash
tools/go.sh run ./cmd/wl-atlas -dir workplace/orig/wastland \
    -lang translations/zh-Hant.cat -out workplace/atlas.json
python3 tools/summarize_walkthrough.py workplace/atlas.json docs/walkthrough/generated
```

`wl-atlas` 只讀不寫，中間那份 `atlas.json` 含原版文本，不進版控。
機制的依據在 `docs/re/`：地圖格的分類看 [`docs/spec/07`](../spec/07-world-events.md)，
條件閘看 [`docs/re/65`](../re/65-third-gate-conditions.md)，
問答看 [`docs/re/46`](../re/46-typed-answers-and-text-input.md)，
傳送與地點名看 [`docs/re/60`](../re/60-teleport-and-map-change.md)。

## 章節

| 章 | 涵蓋 |
|---|---|
| [01 世界與沙漠](01-world.md) | 42 張地圖的全貌、沙漠上十個入口的座標、沙漠本身的危險 |
| [02 石英城](02-quartz.md) | 資源 1–6：石英城、史考特酒吧、驛馬車旅館、法院、幫派藏身處 |
| [03 高池鎮、農業中心、鐵路遊牧民族](03-highpool-agcenter-nomads.md) | 資源 8、9、10、29 |
| [04 針岩城](04-needles.md) | 資源 26、27、28、31、32、33、34 |
| [05 拉斯維加斯](05-las-vegas.md) | 資源 11、12、24、25、38、39、40、41 |
| [06 達爾文與守護者堡壘](06-darwin-and-citadel.md) | 資源 21、22、23、35、36、42 |
| [07 沉睡者基地與科奇斯基地](07-sleeper-and-cochise.md) | 資源 7、13、15、16–20 |
| [08 礦坑與野人村](08-mine-and-savage-village.md) | 資源 43、49 |

譯名以 [`translations/glossary.md`](../../translations/glossary.md) 為準。

## 當年的中文攻略（保存，不是取代）

兩份雜誌攻略逐頁轉錄，另附「機制斷言逐條拿逆向驗過」的對照表。
**它們是史料**：用字、譯名、體例照原樣留著，與上面那份自建攻略互不取代。

| 目錄 | 來源 | 版本 | 特色 |
|---|---|---|---|
| [`jingxun-027/`](jingxun-027/) | 精訊電腦 第 27 期，1988 年 12 月 | **Apple II** | 目前找得到最早的中文攻略，與美國原版同年。第一人稱手記體，六張手繪地圖；〈江湖經驗〉記的是磁片版的玩法 |
| [`swm-005/`](swm-005/) | 軟體世界 第 5 期，1990 年 | DOS | 第三人稱通關指南，31 條機制斷言全部對過（相符 26、有出入 5、未定 0）|

⚠ **兩份的譯名不一樣**：精訊譯 Ranger 為「遊騎兵」，軟體世界譯「遊俠」，
而 1990 年代理版的封面寫「荒野遊俠」。三種都保留。
