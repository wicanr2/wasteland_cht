# DS000：`wl.exe` 基準清冊（IDA 匯出）

> 由 `tools/ida/export_inventory.py` 匯出、
> `tools/summarize_inventory.py` 整理，內容全部是工具輸出，未加語意推論。

| 項目 | 值 |
|---|---|
| 輸入檔 | `target.exe` |
| SHA-256 | `b5eb39f094e0274165eab5e1584e78ff5b54c7228d8db273573d2bd951ea31a0` |
| 大小 | 169,488 bytes |
| processor | `metapc` |
| segments | 6 |
| entry points | 1 |
| 自動辨識 functions | 614 |
| IDA 認定 strings | 43 |
| 沒有直接 caller 的 functions | 2 |

## Segments

| name | class | start | end | size | bitness |
|---|---|---|---|---|---|
| `seg000` | CODE | 0x10000 | 0x1CB67 | 52,071 | 16-bit |
| `seg001` | CODE | 0x1CB67 | 0x1CE20 | 697 | 16-bit |
| `seg002` | UNK | 0x1CE20 | 0x2AE20 | 57,344 | 16-bit |
| `seg003` | UNK | 0x2AE20 | 0x39000 | 57,824 | 16-bit |
| `seg004` | STACK | 0x39000 | 0x39200 | 512 | 16-bit |
| `seg005` | UNK | 0x39200 | 0x39560 | 864 | 16-bit |

## Entry points

| ordinal | linear | segment:offset | name |
|---|---|---|---|
| 69814 | 0x110B6 | seg000+0x10B6 | `start` |

## 最大的 20 個函式

| linear | segment:offset | IDA 名稱 | size | callers |
|---|---|---|---|---|
| 0x14664 | seg000+0x4664 | `sub_14664` | 915 | 2 |
| 0x12D84 | seg000+0x2D84 | `sub_12D84` | 733 | 1 |
| 0x1B170 | seg000+0xB170 | `sub_1B170` | 677 | 1 |
| 0x14BF0 | seg000+0x4BF0 | `sub_14BF0` | 597 | 1 |
| 0x1C6C9 | seg000+0xC6C9 | `sub_1C6C9` | 491 | 1 |
| 0x14480 | seg000+0x4480 | `sub_14480` | 473 | 1 |
| 0x110B6 | seg000+0x10B6 | `start` | 456 | 26 |
| 0x13C58 | seg000+0x3C58 | `sub_13C58` | 443 | 2 |
| 0x13EC9 | seg000+0x3EC9 | `sub_13EC9` | 437 | 1 |
| 0x15500 | seg000+0x5500 | `sub_15500` | 370 | 3 |
| 0x13AE4 | seg000+0x3AE4 | `sub_13AE4` | 366 | 2 |
| 0x130C8 | seg000+0x30C8 | `sub_130C8` | 347 | 1 |
| 0x157D6 | seg000+0x57D6 | `sub_157D6` | 342 | 3 |
| 0x19202 | seg000+0x9202 | `sub_19202` | 321 | 1 |
| 0x1A0C5 | seg000+0xA0C5 | `sub_1A0C5` | 298 | 2 |
| 0x11730 | seg000+0x1730 | `sub_11730` | 292 | 3 |
| 0x16890 | seg000+0x6890 | `sub_16890` | 280 | 2 |
| 0x11F76 | seg000+0x1F76 | `sub_11F76` | 279 | 1 |
| 0x17574 | seg000+0x7574 | `sub_17574` | 271 | 4 |
| 0x18EFE | seg000+0x8EFE | `sub_18EFE` | 262 | 6 |

## 被呼叫最多的 20 個函式

| linear | segment:offset | IDA 名稱 | callers | size |
|---|---|---|---|---|
| 0x16CB2 | seg000+0x6CB2 | `sub_16CB2` | 88 | 11 |
| 0x10039 | seg000+0x39 | `j_start_8` | 71 | 3 |
| 0x19614 | seg000+0x9614 | `sub_19614` | 43 | 38 |
| 0x17208 | seg000+0x7208 | `sub_17208` | 29 | 19 |
| 0x1786E | seg000+0x786E | `sub_1786E` | 28 | 46 |
| 0x18E90 | seg000+0x8E90 | `sub_18E90` | 28 | 29 |
| 0x137F4 | seg000+0x37F4 | `sub_137F4` | 27 | 52 |
| 0x16149 | seg000+0x6149 | `sub_16149` | 27 | 60 |
| 0x110B6 | seg000+0x10B6 | `start` | 26 | 456 |
| 0x19C2C | seg000+0x9C2C | `sub_19C2C` | 26 | 40 |
| 0x19EFC | seg000+0x9EFC | `sub_19EFC` | 26 | 22 |
| 0x13A56 | seg000+0x3A56 | `sub_13A56` | 25 | 28 |
| 0x1728C | seg000+0x728C | `sub_1728C` | 25 | 34 |
| 0x19C04 | seg000+0x9C04 | `sub_19C04` | 25 | 40 |
| 0x18E41 | seg000+0x8E41 | `sub_18E41` | 24 | 30 |
| 0x163C4 | seg000+0x63C4 | `sub_163C4` | 22 | 6 |
| 0x1785E | seg000+0x785E | `sub_1785E` | 20 | 7 |
| 0x17ACE | seg000+0x7ACE | `sub_17ACE` | 20 | 17 |
| 0x17CB1 | seg000+0x7CB1 | `sub_17CB1` | 19 | 33 |
| 0x172BB | seg000+0x72BB | `sub_172BB` | 18 | 25 |

## IDA 認定的字串（全部 43 筆）

| linear | len | xrefs | 內容 |
|---|---|---|---|
| 0x26029 | 20 | 0 | `Scenario Disk 1 in ` |
| 0x2603D | 20 | 0 | `Scenario Disk 2 in ` |
| 0x26051 | 17 | 0 | `Program Disk in ` |
| 0x26062 | 14 | 0 | `Data Disk in ` |
| 0x277EC | 38 | 0 | `Use Enc Order Disband View Save Radio` |
| 0x2786D | 19 | 0 | `bcdefghijklmdenopq` |
| 0x29B4D | 19 | 0 | `$*2:CLU_()-29@HQYb` |
| 0x29C10 | 21 | 0 | `Quit game?YesNo` |
| 0x29C32 | 19 | 0 | `CREATE DELETE PLAY` |
| 0x29D76 | 41 | 0 | `Enter new name for cloned player .>` |
| 0x29DF3 | 36 | 0 | `Setup printer to the top of form.` |
| 0x29E17 | 6 | 0 | `	80N` |
| 0x29E1D | 7 | 0 | `Name: ` |
| 0x29E24 | 7 | 0 | `Rank: ` |
| 0x29E2B | 6 | 0 | `ST =` |
| 0x29E31 | 7 | 0 | `  IQ =` |
| 0x29E38 | 7 | 0 | `  LK =` |
| 0x29E3F | 7 | 0 | `  SP =` |
| 0x29E46 | 7 | 0 | `  AGL=` |
| 0x29E4D | 7 | 0 | `  DEX=` |
| 0x29E54 | 7 | 0 | `  CHR=` |
| 0x29E5B | 7 | 0 | `  SKP=` |
| 0x29E62 | 7 | 0 | `  AC =` |
| 0x29E69 | 11 | 0 | `MaxCon = ` |
| 0x29E74 | 7 | 0 | `Con = ` |
| 0x29E7B | 10 | 0 | `Money = $` |
| 0x29E85 | 6 | 0 | `Sex: ` |
| 0x29E8B | 7 | 0 | `Male  ` |
| 0x29E92 | 7 | 0 | `Female` |
| 0x29E99 | 8 | 0 | `  Nat: ` |
| 0x29EA6 | 8 | 0 | `Russian` |
| 0x29EAE | 8 | 0 | `Mexican` |
| 0x29EB6 | 7 | 0 | `Indian` |
| 0x29EBD | 8 | 0 | `Chinese` |
| 0x29EC5 | 19 | 0 | `Weapon Equipped: ` |
| 0x29ED8 | 17 | 0 | `Armor Equipped: ` |
| 0x29EE9 | 10 | 0 | `Items:` |
| 0x29EF3 | 16 | 0 | `Level-Skill:` |
| 0x29F1C | 45 | 0 | `Printing in progress. Press ESC to abort.` |
| 0x29FAE | 22 | 0 | ` etraoishlndgcyupmb,` |
| 0x2A381 | 204 | 0 | `   !!!"""###$$$$%%%%&&&&''''(((())))****++++,,,,----....////0000111122222222223222222222422222222252222222226222222222722222222232222222224222222222522222222262222222227222222222222222222234567e anrotics` |
| 0x2AC8D | 39 | 0 | `Your life has ended in The Wasteland.` |
| 0x2AD48 | 46 | 0 | ` is already in the game. Enter a new name.` |
