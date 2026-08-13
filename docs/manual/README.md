# Wasteland 官方英文手冊（1988，IBM PC 版）

原版隨盒手冊的純文字檔，逐節整理成 markdown。英文內容全部照原樣保留，
沒有翻譯、沒有摘要、沒有修訂錯字；中文只出現在章節框架與註記。

翻譯要等文字編碼與遊戲內文字表解出來之後才做，這一輪只做結構化。

## 分章索引

| 檔案 | 對應原文章節 | 行數 |
|---|---|---|
| [`01-introduction-and-characters.md`](01-introduction-and-characters.md) | Table of Contents、INTRODUCTION、OBJECTIVE、THE PARTY、CREATING CHARACTERS | 99 |
| [`02-character-profile.md`](02-character-profile.md) | CHARACTER PROFILE（Attributes and Personal Statistics／ITEMS.／Skills）、Onscreen Statistics、Getting Promoted、Recruiting Allies | 135 |
| [`03-commands-and-combat.md`](03-commands-and-combat.md) | Commands、Combat（Hand-to-Hand／Missile Weapon／Selective Fire／Explosives） | 59 |
| [`04-weapons.md`](04-weapons.md) | Weapons List（Long／Medium／Short Range Weapons） | 24 |
| [`05-injuries-time-places.md`](05-injuries-time-places.md) | INJURIES AND DEATH、TIME AND DISTANCE、SPECIAL PLACES（Quartz／Needles／Vegas／Deserts） | 49 |
| [`06-command-summary.md`](06-command-summary.md) | COMMAND SUMMARY 全節（Game Play 至 Macro Functions） | 133 |
| [`07-credits.md`](07-credits.md) | CREDITS | 20 |

原文目錄共列 31 個項目（17 個一級章節、14 個子節），本專案切成 7 個檔案，章節本身沒有增刪。

## 轉成表格的部分

原文用縮排與 tab 排版的清單，這裡改成 markdown 表格，欄位拆法如下。文字內容沒有改動。

| 表格 | 位置 | 拆欄方式 |
|---|---|---|
| 目錄 | 01 | 原文 `章節名<TAB>頁碼`，子節縮排改用全形空白 |
| 屬性與個人統計 | 02 | 原文 `Strength (ST):  說明`，冒號前後拆兩欄 |
| 物品畫面選項 | 02 | 同上 |
| 技能表 | 02 | 原文是 `IQ 3` 分組標題 ＋ `Brawling (1):  說明`，拆成「最低 IQ／技能／首級點數／說明」四欄 |
| 畫面統計欄位 | 02 | 同屬性表 |
| 主選單指令、遭遇選項 | 03 | 同屬性表；原文的 `NOTE:`／`CAUTION:` 與補充段落屬於前一個指令，併入同一格並以空行分隔 |
| 武器表 | 04 | 原文以 `Long/Medium/Short Range Weapons` 分組，改成「射程分類」欄 |
| Ranger Center 選項、非戰鬥指令、戰鬥指令、三個角色畫面的選項 | 06 | 原文 `指令<TAB>說明`，直接對應兩欄 |
| CREDITS | 07 | 原文 `職務<TAB>人名`，同一職務多列人名併入一格 |

## 轉錄處理

原檔是 CP1252 編碼、CRLF 換行、每行硬斷在 110 字元左右的純文字。轉成 markdown 時做了四件事，
都只動排版不動字：

1. **編碼**：CP1252 → UTF-8。非 ASCII 位元組只有三個：`0x92`（167 次）、`0x93`（34 次）、
   `0x94`（34 次），即 `’ “ ”`。
2. **重排段落**：硬斷行接回成一段。原文一句話中間的兩個空白（打字機式句距）在 markdown 裡
   會被瀏覽器合併，這裡統一存成單一空白。
3. **斷字接合**：原文有 13 處在行尾用連字號斷字，接回原詞：
   `Switzerland`、`oblivious`、`enables`、`dismissed`、`automatically`、`confirm`、`disbanded`、
   `bullets`、`distribution`、`knowledge`、`throwing`、`somebody`、`deterioration`。
   除此之外的連字號都是原詞的一部分，沒有動。
4. **標題**：原文標題大小寫本身就不一致（目錄寫 `ONSCREEN STATISTICS`，本文寫
   `Onscreen Statistics`；目錄寫 `Items`，本文寫 `ITEMS.`），一律照本文原樣當標題文字。

轉錄完成後用字詞多重集比對過產出與原檔，除了移進表格欄位而被拆掉的冒號、以及中文框架自身的
字詞之外，沒有落差。

## 原文的錯漏與不一致

原文有相當多的錯字、漏字與前後不一致，**全部照原樣保留**。以下是選錄，不是完整清單，
目的是讓後續翻譯的人不必每次都懷疑自己看錯。

| 位置 | 原文 | 說明 |
|---|---|---|
| THE PARTY | `It you don’t use all of the four slots` | 應為 `If` |
| CREATING CHARACTERS 步驟 3 | `press <RETUREN>` | 應為 `<RETURN>` |
| Intelligence (IQ) | `the mot important attribute` | 應為 `most` |
| Luck (LK) | `avoid more damage then unlucky ones` | 應為 `than` |
| Agility (AGL) | `jump on tables  The higher` | 缺句點 |
| Charisma (CHR) | `a trivial trait. it might well` | 句點應為逗號 |
| Sex | `what bathroom he or she has access to` | 句尾缺句點 |
| Reload（物品畫面） | `you’re asked it you want to Reload` | 應為 `if` |
| Skills | `The skills you posses` | 應為 `possess` |
| Skills | `the higher you IQ` | 應為 `your` |
| Skills / LVL | `the skill level goes up  Skills also improve` | 缺句點 |
| Picklock | `where other don’t want you to go.,` | 句尾多一個逗號 |
| Gamble | `you’ll; also be able to spot` | 多一個分號 |
| Forgery | `Someday you my just need` | 應為 `may` |
| Bureaucracy | `so you can get when you want` | 疑似漏字（`get where`） |
| Safecrack | `An experience practitioner` | 應為 `experienced` |
| 技能重排說明 | `Enter the number of he skill` | 應為 `the` |
| Assault Rifle 技能 vs 武器表 | `M1089A1` vs `M1989A1` | 同一把槍兩種型號寫法 |
| 控制鍵寫法 | `<CONTROL>R` 與 `<CONTROL>-R` 併存 | 全書不統一 |
| Ammunition (AMM) | `left in you equipped weapon` | 應為 `your` |
| Maximum Constitution (MAX) | `a life-threatening illness. like radiation poisoning` | 句點應為逗號 |
| Constitution (CON) | `unless he get medical assistance` | 應為 `gets` |
| Getting Promoted | `when you think you’re accumulated enough experience points` | 應為 `you’ve` |
| Getting Promoted | `add to any attribute your choose Put both points` | 應為 `you choose.` |
| Recruiting Allies | `on how to hire an NPC` | 句尾缺句點 |
| Commands | `listed across the button of the screen` | 應為 `bottom` |
| Enc | `on several characters..` | 兩個句點 |
| CAUTION（Disband） | `you’ll never , ever see them again` | 逗號前多一個空白 |
| Combat | `parties can disband and more to different maps` | 應為 `move` |
| Attack | `your’re asked if you want to fire Single` | 應為 `you’re` |
| Run | `the he’s disbanded from the party  (This is impossible` | 多一個 `the`，且缺句點 |
| Missile Weapon Combat | `must be re-equipped after firing Automatic weapons` | 兩句之間缺句點 |
| Explosives | `may kill the bad guns in a hurry` | 應為 `guys` |
| M1989A1 NATO Assault Rifle | `a solider could kill an enemy` | 應為 `soldier` |
| UZI 27 SMG | `built specifically for fighting terrorists has has proven` | 應為 `and has` |
| VP91Z 9mm Pistol | `featuring an 18-shot capacity of this weapon reduces the need to reload` | 句構不通，疑似漏字 |
| 1911A1 .45 Pistol | `to stop Moro rebels in the Phillipines` | 應為 `Philippines` |
| INJURIES AND DEATH | `Like unconscious characters, whey can do nothing` | 應為 `they` |
| Little Old Quartz | `About the only trouble Quarts has these days` | 同一節內 `Quartz`／`Quarts` 併存 |
| Game Play | `how it saves and keeps track of the fame` | 應為 `game` |
| Game Play | `you can pick u where you left off` | 應為 `pick up` |
| Game Play | `However, if your computer will search for the last place it saved.` | 條件子句沒有主句，疑似整段漏字 |
| Game Play | `assumes, of course, that three’s someplace to return to` | 應為 `there’s` |
| Selecting Options | `unEquip if the exception` | 應為 `is` |
| Movement Commands | `L to more right`、`hold down the mouse button to more in that direction` | 兩處都應為 `move` |
| Save（非戰鬥指令） | `ask “Quite Game (Y/N)?”` | 應為 `Quit`；同一格後面又寫成 `Quit Game(Y/N?)`，括號位置也不同 |
| `<CONTROL>A` | `not hired NPC’s` | 複數誤用所有格 |
| Reload（第二畫面） | `Reload weapon,  (Only appears` | 逗號應為句點 |
| Reload（第二畫面） | `Equipped weapon.)` | 句中大寫 |
| Macro Functions | `(Example:  If you want time to pass ... function key twice.` | 括號沒有收尾 |
| INTRODUCTION | `who repeatedly attached in attempts to reclaim` | 應為 `attacked` |
| INTRODUCTION | `after first believing ... the nuclear malestrom` | 應為 `maelstrom` |
| INTRODUCTION | `Because of each communities’ suspicions towards one another` | 應為 `communities’`／`community’s` 混用 |

## 缺漏的一節

原文目錄列有 `PARAGRAPHS 20`（第 20 至 45 頁），但 `manual.txt` 本體沒有這一節，
從 `Deadly Deserts` 直接跳到 `COMMAND SUMMARY`。段落書是另一個檔案
`paragraphs.txt`，整理結果見 [`../paragraphs/README.md`](../paragraphs/README.md)。

---

來源：`workplace/orig/wastland/manual.txt`（53,322 bytes，1,030 行，CP1252／CRLF，
SHA-256 `4b222c7dc22229bffce83455989a1d164182f40396ffcf54a9445c09a2bc342e`）
