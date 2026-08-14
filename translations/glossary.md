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
| Darwin | 達爾文 | |

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

選項要寫成 `R 逃跑`——原版的 `\x10` 捕捉下一個字元當熱鍵標示，
把字母留在最前面兩邊都對得上（`docs/re/40` §4）。

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
| Mutie Cuties | 變種美女 | 舞團 |

## 地點（石英城）

| 原文 | 譯名 |
|---|---|
| Stagecoach Inn | 驛馬車旅館 |
| Housekeeping | 清潔部 |
| Courthouse | 法院 |
| Devastation Row | 荒廢街 |
| Quail Trail | 鵪鶉小徑 |
| Moon Rd. | 月亮路 |
| Loop Dr. | 環路 |
| Quartz City Jail | 石英城監獄 |
| officer's club | 軍官俱樂部 |

## 物品與俚語

| 原文 | 譯名 | 取捨 |
|---|---|---|
| Snake Squeezins | 蛇酒 | 全大寫時是玩家要輸入的關鍵字，**不翻** |
| Plain sludge | 普通泥漿 | 酒名，直譯保留原本的粗俗感 |
| Fancy sludge | 高級泥漿 | |
| Firewater | 烈火水 | |
| servomotor | 伺服馬達 | |

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

## 標點：人名的間隔號

⚠ 用 **`·`（U+00B7，Big5 `A150`）**，不要用 `‧`（U+2027）——
後者不在 Big5 裡，編譯會擋下來。同類的坑：`・`（U+30FB）也不行；
`．`（U+FF0E→`A144`）與 `•`（`A145`）在 Big5 裡，但那是句點與項目符號，不是名字間隔號。

