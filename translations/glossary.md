# 譯名表

**這份是譯名的唯一真相。** 遇到衝突以這裡為準，改譯名要改這裡，
不是改個別的翻譯檔（`CLAUDE.md` §3.3）。

1990 年軟體世界中文說明書的既有譯名**優先採用**——那是當年台灣玩家看到的字，
是這個專案要保存的東西。與原文不一致時記在「取捨」欄，不要默默改成「正確」的。

## 地名

| 原文 | 譯名 | 取捨 |
|---|---|---|
| Wasteland | 荒野遊俠 | 軟體世界的譯名，遊戲標題沿用 |
| Quartz | 石英城 | |
| Highpool | 高池鎮 | |
| Needles | 針岩城 | |
| Las Vegas | 拉斯維加斯 | |
| Agricultural Center | 農業中心 | |
| Ranger Center | 遊俠中心 | 存檔裡的 `Ranger Ctr.` 縮寫要另外處理 |
| Savage Village | 野人村 | |
| Desert Nomads | 沙漠遊牧民族 | |
| Base Cochise | 科奇斯基地 | |
| Guardian Citadel | 守護者堡壘 | |
| Darwin | 達爾文 |
| Project Darwin | 達爾文計畫 |
| Highpool Creek | 高池溪 |
| Camp Highpool | 高池營 |
| Sleeper One | 沉睡者一號 |
| Temple of the Servants of the Mushroom Cloud | 蘑菇雲僕從神殿 |
| Center of Holy Knowledge | 聖知識中心 |
| Nuclear First Aid | 核子急救站 |
| Boulder Highway | 巨石公路 |
| Maryland Parkway | 馬里蘭大道 |
| Charleston Blvd. | 查爾斯頓大道 |
| Tropicana Ave. | 特羅皮卡納大道 | |

## 屬性

| 原文 | 譯名 |
|---|---|
| Strength | 力量 |
| IQ | 智力 |
| Luck | 幸運 |
| Speed | 速度 |
| Agility | 敏捷 |
| Dexterity | 靈巧 |
| Charisma | 魅力 |
| CON | 體力 |
| MAXCON | 體力上限 |
| AC | 護甲 |

## 技能（部分，35 條見 `docs/re/32` §2）

| 原文 | 譯名 |
|---|---|
| Brawling | 鬥毆 |
| Climb | 攀爬 |
| Swim | 游泳 |
| Perception | 觀察 |
| Medic | 醫護 |
| Doctor | 醫術 |
| Picklock | 開鎖 |
| Gamble | 賭博 |
| Safecrack | 開保險箱 |
| Demolitions | 爆破 |
| Bomb disarm | 拆彈 |
| Alarm disarm | 解除警報 |
| Toaster repair | 修烤麵包機 |

## 狀態與疾病（八個位元，見 `docs/re/35` §1）

| 原文 | 譯名 | 取捨 |
|---|---|---|
| Radiation poisoning | 輻射中毒 | |
| Wasteland Herpes | 荒野皰疹 | **「疱」不在 Big5，一律用「皰」** |
| Bug byte | 蟲咬 | |
| Sewer rot | 下水道腐病 | |
| Desert dust | 沙漠塵肺 | |
| Rabies | 狂犬病 | |
| D6／D7 | （原版佔位，不譯） | 原文就是這兩個字 |

## 傷勢等級

| 原文 | 譯名 |
|---|---|
| UNC（unconscious） | 昏迷 |
| SER（serious） | 重傷 |
| CRT（critical） | 危急 |
| MRT（mortal） | 瀕死 |
| COM（coma） | 深度昏迷 |

## 階級（50 階，見 `docs/re/31` §2.1）

| 原文 | 譯名 |
|---|---|
| Private | 二等兵 |
| Private 1st class | 一等兵 |
| Specialist | 專業兵 |
| Corporal | 下士 |
| Sergeant | 中士 |
| Lieutenant | 中尉 |
| Captain | 上尉 |
| Major | 少校 |
| Colonel | 上校 |
| Commander | 指揮官 |
| General | 將軍 |

## 用字慣例

- **不要用現代用語「潤飾」掉當年的譯法**——這是文化保存，不是重譯。
- 全形標點：`，。！？「」`。倚天的 `SPCFONT.15` 有這些字模。
- 一行不得超過原文的格數（`docs/spec/11` §5，build 會擋）。

## 傷勢狀態（名單 CON 欄印的五個字，`docs/re/15` §4）

| 原文 | 全稱 | 譯名 | 取捨 |
|---|---|---|---|
| UNC | unconscious | 昏迷 | 原文 3 格、中文 2 格，欄寬放得下 |
| SER | seriously wounded | 重傷 | |
| CRT | critically wounded | 危急 | |
| MRT | mortally wounded | 瀕死 | |
| COM | comatose | 昏死 | 與「昏迷」要分得開，所以不用「深度昏迷」 |

## 戰鬥指令（熱鍵**不跟著翻譯走**，`docs/re/38` §2）

| 鍵 | 原文 | 譯名 |
|---|---|---|
| R | Run | 逃跑 |
| U | Use | 使用 |
| H | Hire | 雇用 |
| E | Evade | 迴避 |
| A | Attack | 攻擊 |
| W | Weapon | 換武器 |
| L | Load/unjam | 裝填 |

選項要寫成 `R 逃跑`。`\x10` 後面那個字元同時是畫面上的第一個字**與點擊該列
送出的按鍵**，把字母留在最前面，鍵盤與滑鼠才都對得上（`docs/re/43` §5）。

## 敵人類別（`exe:1:83`–`87`）

| 原文 | 譯名 |
|---|---|
| Animal | 野獸 |
| Mutant | 變種人 |
| Humanoid | 人形生物 |
| Cyborg | 改造人 |
| Robot | 機器人 |

## 人物與幫派（石英城一帶）

| 原文 | 譯名 | 出處 |
|---|---|---|
| Ugly | 阿醜 | `game1:2`、`game1:3` |
| Ugly's gang | 阿醜幫 | |
| Scott | 史考特 | 酒吧老闆 |
| Scott's Bar | 史考特酒吧 | |
| Ellen | 艾倫 | 女侍，蘿莉的母親 |
| Laurie | 蘿莉 | 坐輪椅，艾倫的女兒 |
| Mayor Pedros | 佩德羅斯鎮長 | |
| Cookie | 餅乾 | 廚子 |
| Cookie's Chuckwagon | 餅乾的伙食車 | 餐廳招牌 |
| Mulefoot | 騾腳 | |
| Squint | 瞇眼 | 槍手 |
| Head Crusher | 碎顱者 | |
| the Obliterator | 殲滅者 | |
| Ace | 王牌 | |
| Ugly John | 醜約翰 | `game1:4`，阿醜的全名 |
| Felicia Pedros | 費莉西亞·佩德羅斯 | 鎮長夫人 |
| Danny Citrine | 丹尼·西特林 | 法院裡被刑求的人 |
| Stinger | 刺針 | 幫派小隊長，關鍵字時不翻 |
| Pigface | 豬臉 | 同上 |
| Huey／Dewey | 休伊／杜威 | 三胞胎其中兩個 |
| Wolf | 狼牙 | 狗 |
| Chugbum | 查格本 | |
| Irwin John Finster | 厄文·約翰·芬斯特 | `game2:23`，全大寫時是玩家要輸入的答案，**不翻** |
| Red Ryder | 紅騎士 | 1930 年代美國西部漫畫主角；同名空氣槍的型號保留原文 |
| Rex | 雷克斯 | 巴比的狗 |
| Bobby | 巴比 | |
| Jackie | 潔琪 | |
| Mr. Jumbo | 巨無霸先生 | |
| Joe McCarthy | 喬·麥卡錫 | 1936 年洋基隊總教練 |
| Lou Gehrig | 魯·蓋瑞克 | |
| Fat Freddy | 胖佛萊迪 | `game2:40`，全大寫時是關鍵字，**不翻** |
| Faran Brygo | 法蘭·布萊哥 | 同上 |
| Covenant | 「聖約」 | 人名就是這個字，加引號免得被讀成普通名詞 |
| Max | 麥克斯 | |
| Charmaine | 夏曼 | 蘑菇雲教派的大祭司 |
| Dr. Michael Scott | 麥可·史考特醫生 | `game2:38` |
| Slicerdicer | 切片機 | 機器人 |
| Death Machine | 死亡機器 |
| Spam Shade | 史班·夏德 | `game1:34` 的偵探，名字本身是黑色電影偵探的戲仿 |
| Bill Dugan | 比爾·杜根 | 血之神殿代表；也是原版製作人員之一 |
| Harry | 哈利 | 加油站 |
| Todd | 陶德 | 全大寫時是玩家要輸入的名字，**不翻** |
| robotclerk／robot-cop | 機器警員 | | |
| Mutie Cuties | 變種美女 | 舞團 |
| Gus | 葛斯 | `game2:41` 黑桃賭場的酒保 |
| Big Al／Little Al | 大艾爾／小艾爾 | 同上 |
| Kutie | 酷蒂 | 同上，拿 Mac 17 的那個 |
| Tin Man | 錫人 | 酒窖牆上的塗鴉 |
| Engineer | 工程師 | `game1:8`，1990 說明書的譯法；全大寫時是關鍵字，**不翻** |
| Hobo | 賢人 | 同上。他在遊戲裡是先知不是流浪漢，當年的譯法比直譯準 |
| Brakeman | 煞車員 | 同上；說明書當年沒譯 |
| Atchison／Topeka／Sante Fe | **保留原文** | 鐵路遊牧民族的三個宗族。1990 說明書也保留；那是 AT&SF 鐵路公司的名字拆成三家，同時是玩家要打的關鍵字 |

## 地點

| 原文 | 譯名 |
|---|---|
| Stagecoach Inn | 驛馬車旅館 |
| Housekeeping | 清潔部 |
| Courthouse | 法院 |
| Devastation Row | 荒廢街 |
| Quail Trail | 鵪鶉小徑 |
| Moon Rd. | 月亮路 |
| Loop Dr. | 環路 |
| Hobo Dogs | 遊民熱狗 | 針岩城的速食攤 |
| Crowley's Occult Shop | 克勞利神祕小舖 | |
| Temple of Blood | 血之神殿 | |
| Quartz City Jail | 石英城監獄 |
| officer's club | 軍官俱樂部 |
| Spade's Casino | 黑桃賭場 | `game2:41` |
| Base Cochise 的機器人工廠 | 反應爐核心室／勞安室／機器人維修室／保安電子 | `game2:20` 的四個指示牌 |
| Rail Nomads camp | 鐵路遊牧民族營地 | `game1:8` |
| Caboose／Trading Car／Casino Car | 守車／交易車廂／賭場車廂 | 同上 |

## 物品與俚語

| 原文 | 譯名 | 取捨 |
|---|---|---|
| Snake Squeezins | 蛇酒 | 全大寫時是玩家要輸入的關鍵字，**不翻** |
| Plain sludge | 普通泥漿 | 酒名，直譯保留原本的粗俗感 |
| Fancy sludge | 高級泥漿 | |
| Firewater | 烈火水 | |
| servomotor | 伺服馬達 |
| Bloodstaff | 血杖 |
| sonic key | 音波鑰匙 |
| the Great Glow | 大輝光 | |

## 不翻的詞

密語與玩家要輸入／選取的關鍵字一律保留原文：
`INFO`／`ROOM`／`CHAT`／`MONEY`／`GANG`／`SECRET`／`MAYOR`／`COURTHOUSE`／
`BACK WAY`／`SNAKE SQUEEZINS`／`RIDDLES`／`RIDDLER`／`DANCER`／`DRINK`／
`BYE`／`GOODBYE`／`THANKS`／`THANK YOU`／`TOAST`／`R`／`LETTER R`／
`THE LETTER R`／`UGLY`／`Y`／`N`／`A`–`F`，以及
`URAQT2`／`URABUTLN`／`MUERTE`／`THANATOS`／`UQTU`／`GEN QUART THANA TOES`。

⚠ 謎語 `game1:2:79`（「America 和 Australia 的正中央」＝ 字母 R）與
`game1:2:83`（八個字母稱讚女侍 ＝ `URABUTLN`，念作 "you are a beautiful one"）
**靠英文拼字才成立**。題目照翻但把 `America`／`Australia` 留原文，
答案不翻——這樣玩家看得懂題目，也答得出來。

## Big5 沒有的字

倚天字型是 Big5 排列，編不出來的字 build 會擋。踩過的：

| 想寫 | 為什麼不行 | 改寫成 |
|---|---|---|
| `疱` | 不在 Big5 | `皰`（荒野皰疹） |
| `卐` | 不在 Big5 | 「納粹十字」（段落 20 的壁畫） |
| `獁` | 不在 Big5 | 「長毛象」（段落 21，原文 mastodon） |

## 標點：人名的間隔號

⚠ 用 **`·`（U+00B7，Big5 `A150`）**，不要用 `‧`（U+2027）——
後者不在 Big5 裡，編譯會擋下來。同類的坑：`・`（U+30FB）也不行；
`．`（U+FF0E→`A144`）與 `•`（`A145`）在 Big5 裡，但那是句點與項目符號，不是名字間隔號。

## 翻到會壞掉的東西

有些內容**翻了玩家就答不出來**，一律保留原文：

| 類型 | 例子 | 為什麼 |
|---|---|---|
| 玩家要輸入的答案 | `NEW YORK YANKEES`／`ICEBERG`／`ZERO`／`512` | 輸入比對是英文 |
| 數列題目 | `2, 4, 8, 16, ?` | 答案是數字，題目照原樣 |
| 靠英文拼字的謎語 | `America`／`Australia` 的正中央 ＝ `R` | 翻了謎面就消失 |
| 選項的字母前綴 | `A)`／`O)`／`R 逃跑` | 那是玩家按的鍵 |
| **`\x10` 後面那一個字元** | `\x10Yes` → `\x10Y 是`，不是 `\x10是` | 見下 |

### `\x10` 後面的字母一個都不能動

那個字元同時是**畫面上的第一個字**與**點擊該列送出的按鍵**——原版把它登記到
「每列一格」的熱鍵表，滑鼠點哪一列就送出哪一格（`docs/re/43`）。翻成中文之後
滑鼠會送出 Big5 的首位元組（點了沒反應），鍵盤比對的仍然是原本那個字母
（玩家不知道要按什麼）。**兩邊都壞，而且壞法不一樣。**

寫法是字母留在最前面，中文接在一個半形空白之後：

```
原文    \x11\x10Yes\x0D\x10No\x0D
中文    \x11\x10Y 是\x0D\x10N 否\x0D
```

`tools/build_lang.py` 會逐條比對 `\x10` 後的字元序列，不同就拒絕編譯。

⚠ **題目裡的專有名詞跟著答案走。** `game2:23` 問「1936 年世界大賽誰贏的」，
答案要打 `NEW YORK YANKEES`，所以題目裡不能把隊名翻成「紐約洋基」；
但敘述裡的 Joe McCarthy、Lou Gehrig 不是答案，照翻。

## 惡搞與戲仿（`game1:33`）

針岩城的電影海報牆全是當年片名的戲仿。**照戲仿的方式翻**，讓中文讀者也接得到梗：

| 原文 | 譯名 | 梗 |
|---|---|---|
| Gone With the Nuke Storm | 《核爆隨風而逝》 | 《亂世佳人》(Gone With the Wind) |
| Crater Raider | 《彈坑奇兵》 | 《法櫃奇兵》(Raiders of the Lost Ark) |
| The Radheads | 《輻射頭》 | |
| Humongous and the Mutants | 《巨無霸與變種人》 | |
| Mario, Luigi and the Fireballs | 《馬利歐、路易吉與火球》 | |

水晶球裡那個巫師罵「你以為這是 Bard's Tale 嗎」——那是 Interplay 自家的另一款遊戲，
台灣譯名《冰城傳奇》，用既有譯名。

⚠ **製作人員名單（`game1:33:92`）裡的人名一律保留原文**，只翻職稱標籤。
那是史料，不是敘述。

