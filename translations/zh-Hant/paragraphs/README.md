# 段落書的繁體中文譯稿

英文原文在 [`docs/paragraphs/`](../../../docs/paragraphs/)，這裡的四個檔一一鏡射它，
段落標題一律 `## 段落 N`。編譯：

```bash
python3 tools/build_paragraphs.py     # → translations/paragraphs-zh-Hant.cat
```

編譯會擋三種錯：編號重複或超出 1–162、Big5 編不出來的字、只有標題沒有正文。
沒翻的段落**不會**變成空白——`internal/play.Journal.Text(n)` 回 nil，
由呼叫端決定顯示英文原文還是「尚未翻譯」。

## 進度

| 檔案 | 範圍 | 狀態 |
|---|---|---|
| `01-paragraphs-001-040.md` | 1–40 | 已翻 |
| `02-paragraphs-041-080.md` | 41–80 | 已翻 |
| `03-paragraphs-081-120.md` | 81–120 | 已翻 |
| `04-paragraphs-121-162.md` | 121–162 | 已翻 |

**162 段全部翻完。** `internal/play` 的測試會逐一檢查 1–162 都查得到正文，
缺一段就紅（`docs/spec/19` §6 驗收條件 3）。

## 譯法

- **玩家要輸入的密語、號碼、人名保留原文**：`MORTAL`／`AZRAEL`／`TYRANNOSAURUS`／
  `BUZZARD`／`PHOENIX`／`Hotspur`／`Falstaff`／`Cretian`／`Proteus`，
  以及拆彈碼與身分手環上的數字。理由與遊戲內文字相同（`translations/glossary.md`
  的「翻到會壞掉的東西」）。
- **鐵路遊牧民族的宗族名保留原文**（`Atchison`／`Topeka`／`Sante Fe`／
  `Chattanooga`／`Amtrak`／`Conrail`／`Hider`），與 `game1:8` 一致。
- **職稱沿用 1990 年軟體世界說明書**：Engineer ＝ 工程師、Hobo ＝ 賢人。
- **惡搞照惡搞翻**：段落 20 的壁畫、段落 25 的棒球梗、段落 16 的火星線
  都是原版的玩笑，直譯會把梗弄丟。
- 原文轉錄保留了幾處印刷瑕疵（引號打成連字號、`Finster` 在第 14 段寫成 `Finger`）。
  中文這一份是給玩家讀的正文，**引號照正常收尾**；人名的拼法差異照原文翻，不統一。

## Big5 踩過的字

| 想寫 | 改寫成 |
|---|---|
| `卐` | 「納粹十字」（段落 20） |
| `獁`（猛獁） | 「長毛象」（段落 21） |
| `冩`（寫的異體字，段落 65 的錯字文想用） | 「卸」——同樣是輸入法選錯字的效果，而且在 Big5 裡 |

## 段落 65：錯字文

全書唯一一段刻意寫錯的文字。原文是半文盲守衛的日記
（`secrits`／`rite`／`fegit`／`Mushrum`／`Deth Masheens`／`Siborgs`），
而他在最後一段還誇自己寫得好——那是整段的笑點。
中文用**輸入法選錯字**的別字重現（秘蜜／往記／頭法／音為／摩菇雲／
機氣人／舞器／改造仁／卸），不要改成通順的中文。
