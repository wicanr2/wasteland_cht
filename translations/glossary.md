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
