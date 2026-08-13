# DS000：`wl.exe` 基準清冊（IDA 匯出）

> 由 `tools/ida/export_inventory.py` 匯出、
> `tools/summarize_inventory.py` 整理，內容全部是工具輸出，未加語意推論。

| 項目 | 值 |
|---|---|
| 輸入檔 | `target.exe` |
| SHA-256 | `098aef9b4fe4fea3b8d0d134f82fe11a6dac608839ebd175e168cf0271b93b4f` |
| 大小 | 62,549 bytes |
| processor | `metapc` |
| segments | 3 |
| entry points | 1 |
| 自動辨識 functions | 340 |
| IDA 認定 strings | 40 |
| 沒有直接 caller 的 functions | 22 |

## Segments

| name | class | start | end | size | bitness |
|---|---|---|---|---|---|
| `seg000` | CODE | 0x10000 | 0x1F0C0 | 61,632 | 16-bit |
| `seg001` | CODE | 0x1F0C0 | 0x39700 | 108,096 | 16-bit |
| `seg002` | STACK | 0x39700 | 0x39780 | 128 | 16-bit |

## Entry points

| ordinal | linear | segment:offset | name |
|---|---|---|---|
| 127186 | 0x1F0D2 | seg001+0x12 | `start` |

## 最大的 20 個函式

| linear | segment:offset | IDA 名稱 | size | callers |
|---|---|---|---|---|
| 0x14D03 | seg000+0x4D03 | `sub_14D03` | 995 | 1 |
| 0x16D48 | seg000+0x6D48 | `sub_16D48` | 444 | 1 |
| 0x14755 | seg000+0x4755 | `sub_14755` | 312 | 1 |
| 0x106CC | seg000+0x6CC | `sub_106CC` | 292 | 2 |
| 0x15FD5 | seg000+0x5FD5 | `sub_15FD5` | 256 | 1 |
| 0x164CD | seg000+0x64CD | `sub_164CD` | 255 | 2 |
| 0x17421 | seg000+0x7421 | `sub_17421` | 252 | 2 |
| 0x18E37 | seg000+0x8E37 | `sub_18E37` | 224 | 1 |
| 0x1086E | seg000+0x86E | `sub_1086E` | 214 | 8 |
| 0x17E57 | seg000+0x7E57 | `sub_17E57` | 208 | 2 |
| 0x10581 | seg000+0x581 | `sub_10581` | 199 | 1 |
| 0x1515C | seg000+0x515C | `sub_1515C` | 196 | 1 |
| 0x132CA | seg000+0x32CA | `sub_132CA` | 184 | 1 |
| 0x15969 | seg000+0x5969 | `sub_15969` | 178 | 1 |
| 0x13179 | seg000+0x3179 | `sub_13179` | 156 | 1 |
| 0x15EDF | seg000+0x5EDF | `sub_15EDF` | 148 | 2 |
| 0x17AA4 | seg000+0x7AA4 | `sub_17AA4` | 148 | 2 |
| 0x16F7B | seg000+0x6F7B | `sub_16F7B` | 144 | 4 |
| 0x15D00 | seg000+0x5D00 | `sub_15D00` | 143 | 1 |
| 0x175EF | seg000+0x75EF | `sub_175EF` | 142 | 1 |

## 被呼叫最多的 20 個函式

| linear | segment:offset | IDA 名稱 | callers | size |
|---|---|---|---|---|
| 0x1EF75 | seg000+0xEF75 | `sub_1EF75` | 26 | 6 |
| 0x1EF72 | seg000+0xEF72 | `sub_1EF72` | 16 | 3 |
| 0x152DE | seg000+0x52DE | `sub_152DE` | 15 | 11 |
| 0x1EF6C | seg000+0xEF6C | `sub_1EF6C` | 14 | 6 |
| 0x15C13 | seg000+0x5C13 | `sub_15C13` | 10 | 3 |
| 0x167A1 | seg000+0x67A1 | `sub_167A1` | 10 | 7 |
| 0x1680A | seg000+0x680A | `sub_1680A` | 10 | 68 |
| 0x1086E | seg000+0x86E | `sub_1086E` | 8 | 214 |
| 0x1542E | seg000+0x542E | `sub_1542E` | 8 | 9 |
| 0x15228 | seg000+0x5228 | `sub_15228` | 7 | 18 |
| 0x16169 | seg000+0x6169 | `sub_16169` | 7 | 48 |
| 0x1701A | seg000+0x701A | `sub_1701A` | 7 | 9 |
| 0x15220 | seg000+0x5220 | `sub_15220` | 6 | 8 |
| 0x16D1E | seg000+0x6D1E | `sub_16D1E` | 6 | 15 |
| 0x102FD | seg000+0x2FD | `sub_102FD` | 5 | 24 |
| 0x10ABC | seg000+0xABC | `sub_10ABC` | 5 | 60 |
| 0x1615F | seg000+0x615F | `sub_1615F` | 5 | 6 |
| 0x16395 | seg000+0x6395 | `sub_16395` | 5 | 5 |
| 0x17DBA | seg000+0x7DBA | `sub_17DBA` | 5 | 15 |
| 0x17F70 | seg000+0x7F70 | `nullsub_3` | 5 | 1 |

## IDA 認定的字串（全部 40 筆）

| linear | len | xrefs | 內容 |
|---|---|---|---|
| 0x1BDD8 | 20 | 0 | `Scenario Disk 1 in ` |
| 0x1BDEC | 20 | 0 | `Scenario Disk 2 in ` |
| 0x1BE00 | 17 | 0 | `Program Disk in ` |
| 0x1BE11 | 14 | 0 | `Data Disk in ` |
| 0x1C01B | 24 | 0 | `lcpmuhywfbv:8.91g!q6-'=` |
| 0x1C280 | 38 | 0 | `Use Enc Order Disband View Save Radio` |
| 0x1C2DA | 19 | 0 | `bcdefghijklmdenopq` |
| 0x1D9BA | 20 | 0 | `$*2:CLU_()-29@HQYb?` |
| 0x1DA87 | 19 | 0 | `CREATE DELETE PLAY` |
| 0x1DC01 | 36 | 0 | `Setup printer to the top of form.` |
| 0x1DC25 | 6 | 0 | `	80N` |
| 0x1DC2B | 7 | 0 | `Name: ` |
| 0x1DC32 | 7 | 0 | `Rank: ` |
| 0x1DC39 | 6 | 0 | `ST =` |
| 0x1DC3F | 7 | 0 | `  IQ =` |
| 0x1DC46 | 7 | 0 | `  LK =` |
| 0x1DC4D | 7 | 0 | `  SP =` |
| 0x1DC54 | 7 | 0 | `  AGL=` |
| 0x1DC5B | 7 | 0 | `  DEX=` |
| 0x1DC62 | 7 | 0 | `  CHR=` |
| 0x1DC69 | 7 | 0 | `  SKP=` |
| 0x1DC70 | 7 | 0 | `  AC =` |
| 0x1DC77 | 11 | 0 | `MaxCon = ` |
| 0x1DC82 | 7 | 0 | `Con = ` |
| 0x1DC89 | 10 | 0 | `Money = $` |
| 0x1DC93 | 6 | 0 | `Sex: ` |
| 0x1DC99 | 7 | 0 | `Male  ` |
| 0x1DCA0 | 7 | 0 | `Female` |
| 0x1DCA7 | 8 | 0 | `  Nat: ` |
| 0x1DCB4 | 8 | 0 | `Russian` |
| 0x1DCBC | 8 | 0 | `Mexican` |
| 0x1DCC4 | 7 | 0 | `Indian` |
| 0x1DCCB | 8 | 0 | `Chinese` |
| 0x1DCD3 | 19 | 0 | `Weapon Equipped: ` |
| 0x1DCE6 | 17 | 0 | `Armor Equipped: ` |
| 0x1DCF7 | 10 | 0 | `Items:` |
| 0x1DD01 | 16 | 0 | `Level-Skill:` |
| 0x1DD2A | 45 | 0 | `Printing in progress. Press ESC to abort.` |
| 0x1DD98 | 22 | 0 | ` etraoishlndgcyupmb,` |
| 0x1F1D7 | 24 | 0 | `Packed file is corrupt#` |
