# 00：函式索引

> 由 `tools/gen_func_index.py` 產生。**讀任何 `sub_XXXXX` 之前先查這張表**——
> 筆記超過二三十份之後，靠記憶一定會重讀已經解過的函式。

輸入：`wl.merged.exe（解包映像＋wla.bin overlay，本專案合成）`，SHA-256 `cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`

- 自動辨識函式：**641**
- 已在筆記中出現：**177**
- 尚未碰過：**464**

## 已分析（依呼叫端數量排序）

| 位址 | segment:offset | 大小 | callers | 出現於 |
|---|---|---:|---:|---|
| `0x16CB2` | seg000+0x6CB2 | 11 | 88 | re/19-effects-and-damage.md |
| `0x10039` | seg000+0x39 | 3 | 71 | re/04-overlay-wla-bin.md、re/06-resource-directory.md |
| `0x19614` | seg000+0x9614 | 38 | 43 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/15-character-record.md、re/19-effects-and-damage.md |
| `0x17208` | seg000+0x7208 | 19 | 29 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/15-character-record.md |
| `0x1786E` | seg000+0x786E | 46 | 28 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/03-boot-and-asset-loading.md、re/06-resource-directory.md、re/14-fonts-and-text-encoding.md |
| `0x18E90` | seg000+0x8E90 | 29 | 28 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md |
| `0x137F4` | seg000+0x37F4 | 52 | 27 | re/00-remake-knowledge-gaps.md、re/20-combat-resolution.md |
| `0x19C2C` | seg000+0x9C2C | 40 | 26 | re/00-master-index.md、re/20-combat-resolution.md |
| `0x19EFC` | seg000+0x9EFC | 22 | 26 | re/14-fonts-and-text-encoding.md |
| `0x13A56` | seg000+0x3A56 | 28 | 25 | re/20-combat-resolution.md |
| `0x18E41` | seg000+0x8E41 | 30 | 24 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md、re/20-combat-resolution.md |
| `0x1785E` | seg000+0x785E | 7 | 20 | re/14-fonts-and-text-encoding.md |
| `0x17ACE` | seg000+0x7ACE | 17 | 20 | re/17-packed-text.md |
| `0x17CB1` | seg000+0x7CB1 | 33 | 19 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/16-msq-block-layout.md |
| `0x172BB` | seg000+0x72BB | 25 | 18 | re/20-combat-resolution.md |
| `0x19BF8` | seg000+0x9BF8 | 4 | 17 | re/20-combat-resolution.md |
| `0x1C561` | seg000+0xC561 | 15 | 16 | re/17-packed-text.md |
| `0x13787` | seg000+0x3787 | 16 | 15 | re/20-combat-resolution.md |
| `0x19E30` | seg000+0x9E30 | 6 | 15 | re/14-fonts-and-text-encoding.md |
| `0x12A40` | seg000+0x2A40 | 12 | 14 | re/20-combat-resolution.md |
| `0x19DC3` | seg000+0x9DC3 | 103 | 14 | re/00-master-index.md、re/14-fonts-and-text-encoding.md |
| `0x176A2` | seg000+0x76A2 | 6 | 13 | re/14-fonts-and-text-encoding.md |
| `0x1789C` | seg000+0x789C | 4 | 12 | re/06-resource-directory.md、re/14-fonts-and-text-encoding.md |
| `0x178A3` | seg000+0x78A3 | 22 | 12 | re/00-master-index.md、re/17-packed-text.md |
| `0x11445` | seg000+0x1445 | 121 | 11 | re/00-master-index.md、re/05-storage-layer.md、re/06-resource-directory.md、re/09-msq-map-structure.md、re/10-huffman-compression.md |
| `0x171B9` | seg000+0x71B9 | 15 | 11 | re/14-fonts-and-text-encoding.md、re/15-character-record.md |
| `0x196C9` | seg000+0x96C9 | 18 | 11 | re/15-character-record.md、re/17-packed-text.md |
| `0x19727` | seg000+0x9727 | 73 | 11 | re/14-fonts-and-text-encoding.md |
| `0x19C84` | seg000+0x9C84 | 40 | 11 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md |
| `0x1A26C` | seg000+0xA26C | 9 | 11 | re/14-fonts-and-text-encoding.md |
| `0x1001B` | seg000+0x1B | 3 | 10 | re/04-overlay-wla-bin.md、re/14-fonts-and-text-encoding.md |
| `0x118C3` | seg000+0x18C3 | 15 | 10 | re/05-storage-layer.md、re/06-resource-directory.md |
| `0x12A8D` | seg000+0x2A8D | 16 | 10 | re/20-combat-resolution.md |
| `0x172AE` | seg000+0x72AE | 13 | 10 | re/15-character-record.md |
| `0x17857` | seg000+0x7857 | 7 | 10 | re/14-fonts-and-text-encoding.md |
| `0x17C69` | seg000+0x7C69 | 9 | 10 | re/20-combat-resolution.md |
| `0x198CD` | seg000+0x98CD | 35 | 10 | re/15-character-record.md、re/20-combat-resolution.md |
| `0x19D0E` | seg000+0x9D0E | 33 | 10 | re/20-combat-resolution.md |
| `0x11384` | seg000+0x1384 | 37 | 9 | re/03-boot-and-asset-loading.md、re/05-storage-layer.md |
| `0x113A9` | seg000+0x13A9 | 9 | 9 | re/03-boot-and-asset-loading.md |
| `0x113B2` | seg000+0x13B2 | 39 | 9 | re/03-boot-and-asset-loading.md、re/05-storage-layer.md |
| `0x11534` | seg000+0x1534 | 177 | 9 | re/06-resource-directory.md |
| `0x118D2` | seg000+0x18D2 | 214 | 9 | re/03-boot-and-asset-loading.md、re/05-storage-layer.md、re/11-huffman-decoder.md |
| `0x178A0` | seg000+0x78A0 | 3 | 9 | re/00-remake-knowledge-gaps.md、re/16-msq-block-layout.md、re/17-packed-text.md、re/18-block-text.md |
| `0x19E53` | seg000+0x9E53 | 97 | 9 | re/00-master-index.md、re/14-fonts-and-text-encoding.md |
| `0x1BB5D` | seg000+0xBB5D | 15 | 9 | re/17-packed-text.md |
| `0x1393E` | seg000+0x393E | 9 | 8 | re/00-master-index.md、re/16-msq-block-layout.md |
| `0x1968C` | seg000+0x968C | 32 | 8 | re/15-character-record.md |
| `0x19BF0` | seg000+0x9BF0 | 8 | 8 | re/19-effects-and-damage.md |
| `0x115E5` | seg000+0x15E5 | 199 | 7 | re/00-master-index.md、re/03-boot-and-asset-loading.md、re/05-storage-layer.md、re/09-msq-map-structure.md、re/16-msq-block-layout.md |
| `0x119DB` | seg000+0x19DB | 53 | 7 | re/06-resource-directory.md |
| `0x11B83` | seg000+0x1B83 | 165 | 7 | re/00-master-index.md、re/11-huffman-decoder.md |
| `0x19BC0` | seg000+0x9BC0 | 44 | 7 | re/15-character-record.md、re/19-effects-and-damage.md |
| `0x1A3E1` | seg000+0xA3E1 | 17 | 7 | re/17-packed-text.md |
| `0x11AE8` | seg000+0x1AE8 | 116 | 6 | re/10-huffman-compression.md、re/11-huffman-decoder.md、re/12-msq-tail-and-text-model.md |
| `0x176D0` | seg000+0x76D0 | 47 | 6 | re/19-effects-and-damage.md |
| `0x18E5F` | seg000+0x8E5F | 12 | 6 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md |
| `0x18EFE` | seg000+0x8EFE | 262 | 6 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md、re/14-fonts-and-text-encoding.md |
| `0x19D86` | seg000+0x9D86 | 44 | 6 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md、re/19-effects-and-damage.md、re/20-combat-resolution.md |
| `0x19E2A` | seg000+0x9E2A | 6 | 6 | re/14-fonts-and-text-encoding.md |
| `0x1C213` | seg000+0xC213 | 15 | 6 | re/17-packed-text.md |
| `0x1003F` | seg000+0x3F | 3 | 5 | re/04-overlay-wla-bin.md |
| `0x129E9` | seg000+0x29E9 | 87 | 5 | re/16-msq-block-layout.md |
| `0x16F70` | seg000+0x6F70 | 185 | 5 | re/04-overlay-wla-bin.md |
| `0x17B15` | seg000+0x7B15 | 41 | 5 | re/15-character-record.md、re/19-effects-and-damage.md |
| `0x18024` | seg000+0x8024 | 171 | 5 | re/16-msq-block-layout.md |
| `0x183B1` | seg000+0x83B1 | 110 | 5 | re/05-storage-layer.md、re/06-resource-directory.md、re/07-msq-blocks.md、re/10-huffman-compression.md |
| `0x189B1` | seg000+0x89B1 | 158 | 5 | re/04-overlay-wla-bin.md |
| `0x1BE31` | seg000+0xBE31 | 15 | 5 | re/17-packed-text.md |
| `0x1000C` | seg000+0xC | 3 | 4 | re/04-overlay-wla-bin.md |
| `0x116AC` | seg000+0x16AC | 132 | 4 | re/03-boot-and-asset-loading.md、re/05-storage-layer.md |
| `0x12A4C` | seg000+0x2A4C | 42 | 4 | re/16-msq-block-layout.md |
| `0x17451` | seg000+0x7451 | 102 | 4 | re/00-master-index.md、re/14-fonts-and-text-encoding.md |
| `0x17574` | seg000+0x7574 | 271 | 4 | re/00-master-index.md、re/14-fonts-and-text-encoding.md |
| `0x17E42` | seg000+0x7E42 | 390 | 4 | re/04-overlay-wla-bin.md |
| `0x18801` | seg000+0x8801 | 89 | 4 | re/00-remake-knowledge-gaps.md、re/10-huffman-compression.md |
| `0x19A1D` | seg000+0x9A1D | 36 | 4 | re/00-master-index.md、re/15-character-record.md、re/19-effects-and-damage.md |
| `0x19CCC` | seg000+0x9CCC | 29 | 4 | re/20-combat-resolution.md |
| `0x19CE9` | seg000+0x9CE9 | 21 | 4 | re/19-effects-and-damage.md |
| `0x1000F` | seg000+0xF | 3 | 3 | re/04-overlay-wla-bin.md、re/14-fonts-and-text-encoding.md |
| `0x10D4D` | seg000+0xD4D | 181 | 3 | re/04-overlay-wla-bin.md |
| `0x10E02` | seg000+0xE02 | 67 | 3 | re/04-overlay-wla-bin.md |
| `0x11730` | seg000+0x1730 | 292 | 3 | re/03-boot-and-asset-loading.md、re/05-storage-layer.md |
| `0x11A10` | seg000+0x1A10 | 73 | 3 | re/06-resource-directory.md、re/07-msq-blocks.md |
| `0x11AA3` | seg000+0x1AA3 | 69 | 3 | re/09-msq-map-structure.md |
| `0x11C28` | seg000+0x1C28 | 44 | 3 | re/10-huffman-compression.md、re/11-huffman-decoder.md |
| `0x11C54` | seg000+0x1C54 | 60 | 3 | re/10-huffman-compression.md、re/11-huffman-decoder.md |
| `0x14193` | seg000+0x4193 | 103 | 3 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md、re/19-effects-and-damage.md |
| `0x142ED` | seg000+0x42ED | 39 | 3 | re/00-remake-knowledge-gaps.md |
| `0x157D6` | seg000+0x57D6 | 342 | 3 | re/00-master-index.md、re/13-rng.md、re/15-character-record.md、re/19-effects-and-damage.md、re/20-combat-resolution.md |
| `0x15A84` | seg000+0x5A84 | 25 | 3 | re/19-effects-and-damage.md |
| `0x15BFE` | seg000+0x5BFE | 5 | 3 | re/19-effects-and-damage.md |
| `0x166D3` | seg000+0x66D3 | 67 | 3 | re/06-resource-directory.md |
| `0x17533` | seg000+0x7533 | 49 | 3 | re/14-fonts-and-text-encoding.md |
| `0x178B9` | seg000+0x78B9 | 82 | 3 | re/17-packed-text.md |
| `0x1790B` | seg000+0x790B | 21 | 3 | re/16-msq-block-layout.md、re/17-packed-text.md、re/18-block-text.md |
| `0x17B3E` | seg000+0x7B3E | 30 | 3 | re/15-character-record.md、re/19-effects-and-damage.md |
| `0x17B8F` | seg000+0x7B8F | 56 | 3 | re/00-master-index.md、re/17-packed-text.md |
| `0x18E6B` | seg000+0x8E6B | 37 | 3 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md |
| `0x19004` | seg000+0x9004 | 52 | 3 | re/14-fonts-and-text-encoding.md |
| `0x1B15F` | seg000+0xB15F | 10 | 3 | re/20-combat-resolution.md |
| `0x10000` | seg000+0x0 | 3 | 2 | re/00-master-index.md、re/02-exepack-unpack.md、re/03-boot-and-asset-loading.md、re/04-overlay-wla-bin.md、re/06-resource-directory.md |
| `0x10042` | seg000+0x42 | 3 | 2 | re/04-overlay-wla-bin.md |
| `0x10045` | seg000+0x45 | 3 | 2 | re/04-overlay-wla-bin.md |
| `0x11CA4` | seg000+0x1CA4 | 35 | 2 | re/10-huffman-compression.md、re/11-huffman-decoder.md |
| `0x12A76` | seg000+0x2A76 | 23 | 2 | re/00-master-index.md、re/20-combat-resolution.md |
| `0x141FA` | seg000+0x41FA | 156 | 2 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/19-effects-and-damage.md |
| `0x16890` | seg000+0x6890 | 280 | 2 | re/00-master-index.md、re/00-remake-knowledge-gaps.md |
| `0x16CBD` | seg000+0x6CBD | 10 | 2 | re/17-packed-text.md |
| `0x17BC7` | seg000+0x7BC7 | 64 | 2 | re/00-master-index.md、re/17-packed-text.md |
| `0x184E8` | seg000+0x84E8 | 252 | 2 | re/05-storage-layer.md、re/07-msq-blocks.md、re/10-huffman-compression.md |
| `0x185E6` | seg000+0x85E6 | 110 | 2 | re/00-remake-knowledge-gaps.md、re/10-huffman-compression.md |
| `0x19362` | seg000+0x9362 | 50 | 2 | re/15-character-record.md、re/17-packed-text.md |
| `0x19454` | seg000+0x9454 | 74 | 2 | re/00-master-index.md、re/15-character-record.md |
| `0x1963A` | seg000+0x963A | 80 | 2 | re/15-character-record.md |
| `0x1997B` | seg000+0x997B | 94 | 2 | re/00-master-index.md、re/15-character-record.md |
| `0x19A41` | seg000+0x9A41 | 23 | 2 | re/15-character-record.md |
| `0x1A00A` | seg000+0xA00A | 53 | 2 | re/14-fonts-and-text-encoding.md |
| `0x1A045` | seg000+0xA045 | 51 | 2 | re/14-fonts-and-text-encoding.md |
| `0x1A07E` | seg000+0xA07E | 71 | 2 | re/14-fonts-and-text-encoding.md |
| `0x1A0C5` | seg000+0xA0C5 | 298 | 2 | re/14-fonts-and-text-encoding.md |
| `0x1B0F1` | seg000+0xB0F1 | 8 | 2 | re/20-combat-resolution.md |
| `0x1B108` | seg000+0xB108 | 87 | 2 | re/00-master-index.md、re/20-combat-resolution.md |
| `0x10003` | seg000+0x3 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10006` | seg000+0x6 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10009` | seg000+0x9 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10015` | seg000+0x15 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x1001E` | seg000+0x1E | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10021` | seg000+0x21 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x1002A` | seg000+0x2A | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x1002D` | seg000+0x2D | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10030` | seg000+0x30 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10033` | seg000+0x33 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10036` | seg000+0x36 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x1003C` | seg000+0x3C | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10048` | seg000+0x48 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x1004B` | seg000+0x4B | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x1004E` | seg000+0x4E | 58 | 1 | re/04-overlay-wla-bin.md |
| `0x10088` | seg000+0x88 | 188 | 1 | re/04-overlay-wla-bin.md |
| `0x10144` | seg000+0x144 | 18 | 1 | re/04-overlay-wla-bin.md |
| `0x10156` | seg000+0x156 | 325 | 1 | re/04-overlay-wla-bin.md |
| `0x1029B` | seg000+0x29B | 881 | 1 | re/04-overlay-wla-bin.md |
| `0x1060C` | seg000+0x60C | 189 | 1 | re/00-master-index.md、re/04-overlay-wla-bin.md、re/14-fonts-and-text-encoding.md |
| `0x10762` | seg000+0x762 | 118 | 1 | re/04-overlay-wla-bin.md |
| `0x107D8` | seg000+0x7D8 | 161 | 1 | re/04-overlay-wla-bin.md |
| `0x108A3` | seg000+0x8A3 | 92 | 1 | re/04-overlay-wla-bin.md |
| `0x109BC` | seg000+0x9BC | 94 | 1 | re/04-overlay-wla-bin.md |
| `0x10A1A` | seg000+0xA1A | 96 | 1 | re/04-overlay-wla-bin.md |
| `0x10A7A` | seg000+0xA7A | 151 | 1 | re/04-overlay-wla-bin.md |
| `0x10B11` | seg000+0xB11 | 329 | 1 | re/04-overlay-wla-bin.md |
| `0x10C5A` | seg000+0xC5A | 92 | 1 | re/04-overlay-wla-bin.md |
| `0x10CB6` | seg000+0xCB6 | 151 | 1 | re/00-master-index.md、re/04-overlay-wla-bin.md、re/14-fonts-and-text-encoding.md |
| `0x10E45` | seg000+0xE45 | 85 | 1 | re/04-overlay-wla-bin.md |
| `0x10F12` | seg000+0xF12 | 82 | 1 | re/04-overlay-wla-bin.md |
| `0x10F64` | seg000+0xF64 | 67 | 1 | re/04-overlay-wla-bin.md |
| `0x10FA7` | seg000+0xFA7 | 44 | 1 | re/04-overlay-wla-bin.md |
| `0x11854` | seg000+0x1854 | 111 | 1 | re/03-boot-and-asset-loading.md |
| `0x119A8` | seg000+0x19A8 | 50 | 1 | re/03-boot-and-asset-loading.md |
| `0x11A59` | seg000+0x1A59 | 74 | 1 | re/00-master-index.md、re/08-msq-encryption.md、re/11-huffman-decoder.md、re/16-msq-block-layout.md、re/18-block-text.md |
| `0x11C90` | seg000+0x1C90 | 20 | 1 | re/10-huffman-compression.md |
| `0x16390` | seg000+0x6390 | 7 | 1 | re/17-packed-text.md |
| `0x1708B` | seg000+0x708B | 256 | 1 | re/00-master-index.md、re/15-character-record.md、re/17-packed-text.md |
| `0x17B80` | seg000+0x7B80 | 15 | 1 | re/17-packed-text.md |
| `0x1841F` | seg000+0x841F | 201 | 1 | re/00-master-index.md、re/05-storage-layer.md、re/06-resource-directory.md、re/07-msq-blocks.md、re/08-msq-encryption.md、re/09-msq-map-structure.md、re/10-huffman-compression.md、re/11-huffman-decoder.md、re/12-msq-tail-and-text-model.md、re/16-msq-block-layout.md、re/18-block-text.md |
| `0x186B6` | seg000+0x86B6 | 142 | 1 | re/10-huffman-compression.md |
| `0x18744` | seg000+0x8744 | 189 | 1 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/05-storage-layer.md、re/07-msq-blocks.md、re/09-msq-map-structure.md、re/10-huffman-compression.md |
| `0x18B4C` | seg000+0x8B4C | 20 | 1 | re/14-fonts-and-text-encoding.md |
| `0x18EAD` | seg000+0x8EAD | 76 | 1 | re/14-fonts-and-text-encoding.md |
| `0x19C72` | seg000+0x9C72 | 9 | 1 | re/20-combat-resolution.md |
| `0x19F12` | seg000+0x9F12 | 224 | 1 | re/14-fonts-and-text-encoding.md |
| `0x1B7FE` | seg000+0xB7FE | 119 | 1 | re/05-storage-layer.md、re/07-msq-blocks.md、re/10-huffman-compression.md |
| `0x1CB67` | seg001+0x0 | 10 | 1 | re/00-master-index.md、re/02-exepack-unpack.md |
| `0x1CB75` | seg001+0xE | 53 | 1 | re/00-master-index.md、re/00-remake-knowledge-gaps.md |
| `0x1CBAA` | seg001+0x43 | 41 | 1 | re/00-remake-knowledge-gaps.md |
| `0x1CC76` | seg001+0x10F | 83 | 1 | re/00-master-index.md、re/00-remake-knowledge-gaps.md |
| `0x1CD52` | seg001+0x1EB | 194 | 1 | re/00-master-index.md、re/00-remake-knowledge-gaps.md |
| `0x110B6` | seg000+0x10B6 | 648 | 0 | re/00-master-index.md、re/02-exepack-unpack.md、re/03-boot-and-asset-loading.md、re/04-overlay-wla-bin.md |

## 尚未碰過的函式（依大小排序，前 60 個）

大的通常是主邏輯，是後續分析的優先對象。

| 位址 | segment:offset | 大小 | callers |
|---|---|---:|---:|
| `0x14664` | seg000+0x4664 | 915 | 2 |
| `0x1C6C9` | seg000+0xC6C9 | 789 | 1 |
| `0x12D84` | seg000+0x2D84 | 733 | 1 |
| `0x11CD0` | seg000+0x1CD0 | 678 | 1 |
| `0x1B170` | seg000+0xB170 | 677 | 1 |
| `0x14BF0` | seg000+0x4BF0 | 597 | 1 |
| `0x14480` | seg000+0x4480 | 473 | 1 |
| `0x13C58` | seg000+0x3C58 | 443 | 2 |
| `0x13EC9` | seg000+0x3EC9 | 437 | 1 |
| `0x15500` | seg000+0x5500 | 370 | 3 |
| `0x13AE4` | seg000+0x3AE4 | 366 | 2 |
| `0x130C8` | seg000+0x30C8 | 347 | 1 |
| `0x19202` | seg000+0x9202 | 321 | 1 |
| `0x11F76` | seg000+0x1F76 | 279 | 1 |
| `0x161C0` | seg000+0x61C0 | 263 | 1 |
| `0x15036` | seg000+0x5036 | 261 | 1 |
| `0x15A9D` | seg000+0x5A9D | 242 | 2 |
| `0x17748` | seg000+0x7748 | 241 | 0 |
| `0x12440` | seg000+0x2440 | 222 | 2 |
| `0x19130` | seg000+0x9130 | 210 | 3 |
| `0x1C073` | seg000+0xC073 | 205 | 1 |
| `0x18860` | seg000+0x8860 | 200 | 5 |
| `0x19394` | seg000+0x9394 | 192 | 3 |
| `0x1BA72` | seg000+0xBA72 | 192 | 2 |
| `0x1BF5F` | seg000+0xBF5F | 188 | 1 |
| `0x121A8` | seg000+0x21A8 | 186 | 2 |
| `0x12636` | seg000+0x2636 | 161 | 1 |
| `0x10FD3` | seg000+0xFD3 | 158 | 4 |
| `0x17923` | seg000+0x7923 | 148 | 3 |
| `0x18B6B` | seg000+0x8B6B | 148 | 2 |
| `0x16DB4` | seg000+0x6DB4 | 143 | 11 |
| `0x135C6` | seg000+0x35C6 | 139 | 1 |
| `0x198F0` | seg000+0x98F0 | 139 | 3 |
| `0x134B2` | seg000+0x34B2 | 138 | 1 |
| `0x1CCC9` | seg001+0x162 | 137 | 1 |
| `0x18928` | seg000+0x8928 | 136 | 1 |
| `0x172D4` | seg000+0x72D4 | 131 | 4 |
| `0x19ACD` | seg000+0x9ACD | 130 | 2 |
| `0x1B735` | seg000+0xB735 | 130 | 12 |
| `0x14E4A` | seg000+0x4E4A | 129 | 1 |
| `0x16D34` | seg000+0x6D34 | 128 | 13 |
| `0x15D30` | seg000+0x5D30 | 127 | 1 |
| `0x179B7` | seg000+0x79B7 | 126 | 5 |
| `0x1818E` | seg000+0x818E | 126 | 3 |
| `0x174B7` | seg000+0x74B7 | 124 | 1 |
| `0x1820C` | seg000+0x820C | 124 | 2 |
| `0x173D2` | seg000+0x73D2 | 122 | 9 |
| `0x1CA14` | seg000+0xCA14 | 121 | 3 |
| `0x1B43C` | seg000+0xB43C | 120 | 4 |
| `0x114BE` | seg000+0x14BE | 118 | 2 |
| `0x16428` | seg000+0x6428 | 118 | 2 |
| `0x18A4F` | seg000+0x8A4F | 117 | 2 |
| `0x18DCE` | seg000+0x8DCE | 115 | 4 |
| `0x19506` | seg000+0x9506 | 115 | 2 |
| `0x19814` | seg000+0x9814 | 114 | 2 |
| `0x149F7` | seg000+0x49F7 | 110 | 2 |
| `0x1A1EF` | seg000+0xA1EF | 108 | 4 |
| `0x139CE` | seg000+0x39CE | 105 | 1 |
| `0x167CE` | seg000+0x67CE | 105 | 2 |
| `0x160E1` | seg000+0x60E1 | 104 | 2 |
