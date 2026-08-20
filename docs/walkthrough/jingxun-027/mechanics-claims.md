# 精訊 27 期〈遊騎兵手記〉的機制斷言，逐條拿逆向驗過

〈遊騎兵手記〉是敘事體，大部分篇幅在講四個人怎麼走這一趟；
可以拿逆向結果驗的**機制敘述**抽出來是 14 條。

⚠ **這一篇寫的是 Apple II 版**，本專案逆向的是 DOS 版。
所以這裡多一種狀態：**版本差異**——不是攻略記錯，是兩個平台本來就不一樣。

| 狀態 | 意思 |
|---|---|
| 相符 | 逆向結果支持這條敘述 |
| 版本差異 | 這條在 Apple II 版成立，DOS 版的做法不同或不存在 |
| 有出入 | 逆向結果與敘述不一致，而且不像是版本差異 |
| 未定 | 逆向還沒有對應的機制可以驗——**不是「錯」，是「還沒有辦法判斷」** |

2026-08-20 對過一輪：**相符 9、版本差異 2、有出入 1、未定 2**。

| # | 出處 | 斷言 | 對照 | 狀態 |
|---|---|---|---|---|
| J-01 | [p.49](p49.md) | 水壺是行走沙漠所必需，可以紓解炎熱 | [`re/67`](../../re/67-gate-penalty-and-canteen.md)、[`generated/gates.md`](../generated/gates.md) | **相符** |
| J-02 | [p.49](p49.md) | 遇上斷崖時有可能用到繩索 | [`generated/gates.md`](../generated/gates.md)（物品 54）| **相符** |
| J-03 | [p.49](p49.md)、[p.51](p51.md) | 高池鎮的淨水機缺一具引擎；引擎每具 500 塊 | [`generated/ida94/items.md`](../../re/generated/ida94/items.md)：`Engine` 是物品 46，基礎價 **500** | **相符** |
| J-04 | [p.49](p49.md) | 先學「理解」技能才找得到那個洞穴 | [`re/32`](../../re/32-skill-checks-and-xp.md) 技能檢定開路 | **相符** |
| J-05 | [p.50](p50.md) | 流浪漢要 `Snake Squeezin`，喝十罐才倒 | 攜帶檢查成立（[`generated/gates.md`](../generated/gates.md) 有四處要蛇汁），**「十罐」那個次數沒有對應的機制** | **未定** |
| J-06 | [p.51](p51.md) | 礦坑深處的怪物叫 `Shadowclaw` | 42 張地圖的字串表裡**一個 `claw` 都沒有**（只有世界地圖的 `claws` 與資源 15 的 `razor-clawed attackers`）。敵人名字存在各地圖自己的字串表（[`re/18`](../../re/18-block-text.md)），所以掃得乾淨 | **有出入** |
| J-07 | [p.51](p51.md)、[p.67](p67.md) | 礦坑東南方那口箱子有一條繩索與四具防毒面具 | [`generated/ida94/chests.md`](../../re/generated/ida94/chests.md)：資源 43 的第 2 筆 ＝ `54 Rope、47 Gas mask ×4` | **相符** |
| J-08 | [p.51](p51.md) | 賭博要專門的技能 | 技能表裡有 `Gamble`（[`re/80`](../../re/80-trainer-skill-list.md) 的 36 筆）| **相符** |
| J-09 | [p.52](p52.md) | 石英城走下水道不會遇敵，走街道會 | **地面與下水道是同一張圖**（資源 1，`1 → 1` 的自我傳送「你在黏滑的地底爬行」），所以**遭遇分母一樣是 30**（[`generated/maps.md`](../generated/maps.md)）。差別要看遭遇來源的分布，還沒掃 | **未定** |
| J-10 | [p.52](p52.md) | 城裡唯一的醫院叫 `Quack's Clinic` | 資源 1 的字串：`Entering Dr. Quack's emergency clinic.` | **相符** |
| J-11 | [p.53](p53.md) | 老人的三個謎題答案是 `TOAST`、`R`、`URABUTLN` | [`generated/passwords.md`](../generated/passwords.md) 三題逐條對上，而且是**打字輸入** | **相符** |
| J-12 | [p.53](p53.md) | 對女侍艾倫說 `URABUTLN` 會拿到鑰匙 | 同上：艾倫那一格接受 `DRINK`、`CHAT`、`URABUTLN` | **相符** |
| J-13 | [p.53](p53.md) | 把 `Visa Card` 給 Head Crusher 換 Atchison 帳篷密語 | `Visa card` 是物品 67；問答關鍵字表裡也有 `VISA CARD`（[`generated/ida94/questions.md`](../../re/generated/ida94/questions.md) 資源 8）| **相符** |
| J-14 | [p.67](p67.md) | 抽換資料片可以復原狀態，靠它反覆刷 Gas Mask 賣錢升級 | 磁片抽換是 Apple II 版的操作；DOS 版的存檔是 `game1`／`game2` 檔尾的 MSQ 資源（[`re/30`](../../re/30-save-layout.md)），沒有磁片可換。**道理成立**（寶箱第一次踩到才擲並寫回，[`re/29`](../../re/29-map-event-handlers.md) §4），做法不通用 | **版本差異** |
| J-15 | [p.67](p67.md) | `Ctrl`－`R` 可以重排技能與物品 | DOS 版的按鍵表裡沒有這個組合（[`re/43`](../../re/43-input-and-hotkeys.md)）| **版本差異** |

---

## 這一輪對出來的兩件事

**「一條繩索、四具防毒面具」是同一筆寶箱記錄。** 這是 Apple II 版的攻略與
DOS 版資料互相印證的一處：兩個平台的礦坑那口箱子放的是同一批東西，
連件數都一樣。件數是這一輪才讀出來的欄位——`tools/summarize_chests.py`
現在會把記錄裡每個物品後面那個 byte 印出來（`×4` 是出貨寫死的，
`×1dN` 是踩到才擲）。

**`Sante Fe` 不是作者拼錯，是遊戲自己拼錯。** 問答關鍵字表裡就是 `SANTE FE`
（`generated/ida94/questions.md` 資源 8 第 3 筆），照正確的 `Santa Fe` 打反而問不到。

## 下一輪的入口

- **J-09**（下水道不遇敵）：掃資源 1 的遭遇來源分布——遭遇是從**視窗內的敵人組**
  掃出來的（[`re/39`](../../re/39-encounter-scan.md)），同一張圖上不同區域的
  來源密度可以差很多。這一條回答得了就從「未定」收掉。
- **J-05**（十罐蛇汁）：找流浪漢那一格的記錄，看攜帶檢查後面接的是不是一個計數器。
- **J-06**（Shadowclaw）：要答案得看 Apple II 版的資料檔，DOS 版問不出更多。
