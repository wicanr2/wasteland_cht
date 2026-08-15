# 00：函式索引

> 由 `tools/gen_func_index.py` 產生。**讀任何 `sub_XXXXX` 之前先查這張表**——
> 筆記超過二三十份之後，靠記憶一定會重讀已經解過的函式。

輸入：`wl.merged.exe（解包映像＋wla.bin overlay，本專案合成）`，SHA-256 `cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`

- 自動辨識函式：**641**
- 已在筆記中出現：**430**
- 尚未碰過：**211**

## 已分析（依呼叫端數量排序）

| 位址 | segment:offset | 大小 | callers | 出現於 |
|---|---|---:|---:|---|
| `0x16CB2` | seg000+0x6CB2 | 11 | 88 | re/19-effects-and-damage.md、re/21-attributes.md、re/38-combat-commands-and-flee.md、re/40-combat-screen.md、re/43-input-and-hotkeys.md、re/64-enter-location-prompt.md |
| `0x10039` | seg000+0x39 | 3 | 71 | re/04-overlay-wla-bin.md、re/06-resource-directory.md |
| `0x19614` | seg000+0x9614 | 38 | 43 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/15-character-record.md、re/19-effects-and-damage.md、re/28-text-variants.md、re/38-combat-commands-and-flee.md、re/55-radiation-and-armour-bypass.md、re/67-gate-penalty-and-canteen.md |
| `0x17208` | seg000+0x7208 | 19 | 29 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/15-character-record.md |
| `0x1786E` | seg000+0x786E | 46 | 28 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/03-boot-and-asset-loading.md、re/06-resource-directory.md、re/14-fonts-and-text-encoding.md、re/46-typed-answers-and-text-input.md |
| `0x18E90` | seg000+0x8E90 | 29 | 28 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md、re/21-attributes.md、re/26-movement-and-triggers.md、re/46-typed-answers-and-text-input.md、re/72-facility-entry-and-command-bar.md |
| `0x137F4` | seg000+0x37F4 | 52 | 27 | re/00-master-index.md、re/20-combat-resolution.md、re/32-skill-checks-and-xp.md、re/37-enemy-records-and-hp.md |
| `0x16149` | seg000+0x6149 | 60 | 27 | re/00-master-index.md、re/25-screen-layout.md、re/37-enemy-records-and-hp.md、re/51-encounter-driver.md |
| `0x19C2C` | seg000+0x9C2C | 40 | 26 | re/00-master-index.md、re/20-combat-resolution.md、re/21-attributes.md、re/32-skill-checks-and-xp.md、re/36-combat-rounds.md |
| `0x19EFC` | seg000+0x9EFC | 22 | 26 | re/00-master-index.md、re/14-fonts-and-text-encoding.md、re/28-text-variants.md、re/51-encounter-driver.md、re/55-radiation-and-armour-bypass.md、re/58-line-flush-and-scrollback.md |
| `0x13A56` | seg000+0x3A56 | 28 | 25 | re/20-combat-resolution.md、re/32-skill-checks-and-xp.md、re/37-enemy-records-and-hp.md |
| `0x1728C` | seg000+0x728C | 34 | 25 | re/25-screen-layout.md、re/29-map-event-handlers.md、re/38-combat-commands-and-flee.md、re/40-combat-screen.md、re/41-command-handlers.md、re/42-facility-loops.md、re/51-encounter-driver.md、re/52-trainer-facility.md、re/72-facility-entry-and-command-bar.md |
| `0x18E41` | seg000+0x8E41 | 30 | 24 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md、re/20-combat-resolution.md、re/29-map-event-handlers.md、re/37-enemy-records-and-hp.md |
| `0x163C4` | seg000+0x63C4 | 50 | 22 | re/25-screen-layout.md、re/29-map-event-handlers.md、re/38-combat-commands-and-flee.md、re/40-combat-screen.md、re/51-encounter-driver.md、re/64-enter-location-prompt.md |
| `0x1785E` | seg000+0x785E | 7 | 20 | re/14-fonts-and-text-encoding.md、re/40-combat-screen.md |
| `0x17ACE` | seg000+0x7ACE | 17 | 20 | re/17-packed-text.md |
| `0x17CB1` | seg000+0x7CB1 | 33 | 19 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/16-msq-block-layout.md、re/24-map-layers-and-tiles.md、re/71-nibble12-batch-patch.md、re/74-heat-entry-and-gate-display.md |
| `0x172BB` | seg000+0x72BB | 25 | 18 | re/00-master-index.md、re/20-combat-resolution.md、re/26-movement-and-triggers.md、re/34-map-script-opcodes.md、re/38-combat-commands-and-flee.md、re/39-encounter-scan.md、re/42-facility-loops.md、re/45-item-data-and-weapon-damage.md、re/51-encounter-driver.md、re/52-trainer-facility.md、re/69-gate-flags.md |
| `0x19BF8` | seg000+0x9BF8 | 4 | 17 | re/20-combat-resolution.md、re/21-attributes.md、re/32-skill-checks-and-xp.md、re/36-combat-rounds.md、re/45-item-data-and-weapon-damage.md |
| `0x1C561` | seg000+0xC561 | 15 | 16 | re/17-packed-text.md、re/29-map-event-handlers.md、re/35-status-and-healing.md、re/42-facility-loops.md |
| `0x13787` | seg000+0x3787 | 16 | 15 | re/20-combat-resolution.md、re/37-enemy-records-and-hp.md |
| `0x19E30` | seg000+0x9E30 | 6 | 15 | re/14-fonts-and-text-encoding.md、re/40-combat-screen.md、re/41-command-handlers.md |
| `0x10EBE` | seg000+0xEBE | 60 | 14 | re/00-remake-knowledge-gaps.md、re/04-overlay-wla-bin.md |
| `0x12A40` | seg000+0x2A40 | 12 | 14 | re/00-master-index.md、re/20-combat-resolution.md、re/32-skill-checks-and-xp.md、re/36-combat-rounds.md、re/37-enemy-records-and-hp.md |
| `0x19C69` | seg000+0x9C69 | 9 | 14 | re/32-skill-checks-and-xp.md、re/36-combat-rounds.md |
| `0x19DC3` | seg000+0x9DC3 | 103 | 14 | re/00-master-index.md、re/14-fonts-and-text-encoding.md、re/43-input-and-hotkeys.md、re/46-typed-answers-and-text-input.md、re/53-list-framework.md |
| `0x16D34` | seg000+0x6D34 | 128 | 13 | re/00-master-index.md、re/22-shop-and-items.md、re/41-command-handlers.md、re/42-facility-loops.md、re/52-trainer-facility.md、re/53-list-framework.md |
| `0x176A2` | seg000+0x76A2 | 6 | 13 | re/28-text-variants.md |
| `0x17AE0` | seg000+0x7AE0 | 16 | 13 | re/00-master-index.md、re/21-attributes.md、re/22-shop-and-items.md、re/29-map-event-handlers.md、re/32-skill-checks-and-xp.md、re/37-enemy-records-and-hp.md、re/41-command-handlers.md、re/42-facility-loops.md、re/45-item-data-and-weapon-damage.md |
| `0x17C20` | seg000+0x7C20 | 73 | 13 | re/00-master-index.md、re/24-map-layers-and-tiles.md、re/26-movement-and-triggers.md、re/39-encounter-scan.md、re/48-map-icons.md |
| `0x19D2F` | seg000+0x9D2F | 30 | 13 | re/00-master-index.md、re/37-enemy-records-and-hp.md、re/39-encounter-scan.md、re/45-item-data-and-weapon-damage.md |
| `0x142E2` | seg000+0x42E2 | 11 | 12 | re/55-radiation-and-armour-bypass.md、re/65-third-gate-conditions.md、re/69-gate-flags.md |
| `0x1789C` | seg000+0x789C | 4 | 12 | re/06-resource-directory.md、re/14-fonts-and-text-encoding.md、re/40-combat-screen.md、re/41-command-handlers.md |
| `0x178A3` | seg000+0x78A3 | 22 | 12 | re/00-master-index.md、re/17-packed-text.md |
| `0x1B735` | seg000+0xB735 | 130 | 12 | re/29-map-event-handlers.md |
| `0x11445` | seg000+0x1445 | 121 | 11 | re/00-master-index.md、re/05-storage-layer.md、re/06-resource-directory.md、re/09-msq-map-structure.md、re/10-huffman-compression.md |
| `0x137CE` | seg000+0x37CE | 38 | 11 | re/00-master-index.md、re/37-enemy-records-and-hp.md |
| `0x14AE1` | seg000+0x4AE1 | 9 | 11 | re/39-encounter-scan.md |
| `0x16DB4` | seg000+0x6DB4 | 143 | 11 | re/00-master-index.md、re/42-facility-loops.md、re/52-trainer-facility.md、re/53-list-framework.md |
| `0x16F20` | seg000+0x6F20 | 21 | 11 | re/41-command-handlers.md |
| `0x17033` | seg000+0x7033 | 88 | 11 | re/25-screen-layout.md、re/42-facility-loops.md |
| `0x171B9` | seg000+0x71B9 | 15 | 11 | re/14-fonts-and-text-encoding.md、re/15-character-record.md、re/40-combat-screen.md、re/51-encounter-driver.md |
| `0x196C9` | seg000+0x96C9 | 18 | 11 | re/15-character-record.md、re/17-packed-text.md、re/20-combat-resolution.md、re/38-combat-commands-and-flee.md、re/41-command-handlers.md |
| `0x19727` | seg000+0x9727 | 73 | 11 | re/14-fonts-and-text-encoding.md、re/28-text-variants.md、re/40-combat-screen.md、re/46-typed-answers-and-text-input.md、re/52-trainer-facility.md、re/72-facility-entry-and-command-bar.md |
| `0x19C84` | seg000+0x9C84 | 40 | 11 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md、re/21-attributes.md、re/32-skill-checks-and-xp.md、re/36-combat-rounds.md |
| `0x1A26C` | seg000+0xA26C | 9 | 11 | re/14-fonts-and-text-encoding.md、re/40-combat-screen.md |
| `0x1001B` | seg000+0x1B | 3 | 10 | re/04-overlay-wla-bin.md、re/14-fonts-and-text-encoding.md |
| `0x118C3` | seg000+0x18C3 | 15 | 10 | re/05-storage-layer.md、re/06-resource-directory.md |
| `0x12A8D` | seg000+0x2A8D | 16 | 10 | re/20-combat-resolution.md、re/32-skill-checks-and-xp.md、re/37-enemy-records-and-hp.md |
| `0x172AE` | seg000+0x72AE | 13 | 10 | re/15-character-record.md、re/42-facility-loops.md |
| `0x17857` | seg000+0x7857 | 7 | 10 | re/14-fonts-and-text-encoding.md |
| `0x17C69` | seg000+0x7C69 | 9 | 10 | re/20-combat-resolution.md、re/45-item-data-and-weapon-damage.md |
| `0x19720` | seg000+0x9720 | 7 | 10 | re/64-enter-location-prompt.md、re/67-gate-penalty-and-canteen.md、re/69-gate-flags.md、re/74-heat-entry-and-gate-display.md |
| `0x198CD` | seg000+0x98CD | 35 | 10 | re/00-master-index.md、re/15-character-record.md、re/20-combat-resolution.md、re/32-skill-checks-and-xp.md、re/65-third-gate-conditions.md |
| `0x19D0E` | seg000+0x9D0E | 33 | 10 | re/20-combat-resolution.md、re/38-combat-commands-and-flee.md、re/39-encounter-scan.md、re/51-encounter-driver.md |
| `0x19D4D` | seg000+0x9D4D | 57 | 10 | re/00-master-index.md、re/37-enemy-records-and-hp.md、re/39-encounter-scan.md |
| `0x11384` | seg000+0x1384 | 37 | 9 | re/03-boot-and-asset-loading.md、re/05-storage-layer.md、re/56-transtbl.md |
| `0x113A9` | seg000+0x13A9 | 9 | 9 | re/03-boot-and-asset-loading.md |
| `0x113B2` | seg000+0x13B2 | 39 | 9 | re/03-boot-and-asset-loading.md、re/05-storage-layer.md、re/56-transtbl.md |
| `0x11534` | seg000+0x1534 | 177 | 9 | re/06-resource-directory.md、re/30-save-layout.md、re/45-item-data-and-weapon-damage.md |
| `0x118D2` | seg000+0x18D2 | 214 | 9 | re/03-boot-and-asset-loading.md、re/05-storage-layer.md、re/11-huffman-decoder.md |
| `0x173D2` | seg000+0x73D2 | 122 | 9 | re/72-facility-entry-and-command-bar.md |
| `0x17852` | seg000+0x7852 | 5 | 9 | re/42-facility-loops.md |
| `0x178A0` | seg000+0x78A0 | 3 | 9 | re/16-msq-block-layout.md、re/17-packed-text.md、re/18-block-text.md、re/22-shop-and-items.md、re/46-typed-answers-and-text-input.md、re/52-trainer-facility.md、re/67-gate-penalty-and-canteen.md、re/69-gate-flags.md、re/74-heat-entry-and-gate-display.md |
| `0x17CD2` | seg000+0x7CD2 | 25 | 9 | re/00-master-index.md、re/16-msq-block-layout.md、re/29-map-event-handlers.md、re/34-map-script-opcodes.md、re/46-typed-answers-and-text-input.md、re/69-gate-flags.md、re/70-nibble1-and-facility-entry.md |
| `0x18D27` | seg000+0x8D27 | 5 | 9 | re/41-command-handlers.md |
| `0x19E53` | seg000+0x9E53 | 97 | 9 | re/00-master-index.md、re/14-fonts-and-text-encoding.md、re/43-input-and-hotkeys.md、re/51-encounter-driver.md、re/69-gate-flags.md |
| `0x1BB5D` | seg000+0xBB5D | 15 | 9 | re/00-master-index.md、re/17-packed-text.md、re/31-experience-and-skills.md |
| `0x125B7` | seg000+0x25B7 | 6 | 8 | re/38-combat-commands-and-flee.md、re/51-encounter-driver.md |
| `0x1393E` | seg000+0x393E | 9 | 8 | re/00-master-index.md、re/16-msq-block-layout.md、re/32-skill-checks-and-xp.md |
| `0x1651A` | seg000+0x651A | 19 | 8 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/26-movement-and-triggers.md、re/32-skill-checks-and-xp.md |
| `0x16EE8` | seg000+0x6EE8 | 56 | 8 | re/42-facility-loops.md |
| `0x1721B` | seg000+0x721B | 52 | 8 | re/42-facility-loops.md、re/43-input-and-hotkeys.md、re/46-typed-answers-and-text-input.md、re/52-trainer-facility.md |
| `0x18350` | seg000+0x8350 | 97 | 8 | re/00-master-index.md、re/24-map-layers-and-tiles.md、re/25-screen-layout.md、re/27-game-clock.md、re/51-encounter-driver.md、re/60-teleport-and-map-change.md |
| `0x18D2C` | seg000+0x8D2C | 5 | 8 | re/41-command-handlers.md |
| `0x1968C` | seg000+0x968C | 32 | 8 | re/00-master-index.md、re/15-character-record.md、re/32-skill-checks-and-xp.md、re/41-command-handlers.md、re/65-third-gate-conditions.md |
| `0x19BF0` | seg000+0x9BF0 | 8 | 8 | re/00-master-index.md、re/19-effects-and-damage.md、re/55-radiation-and-armour-bypass.md、re/67-gate-penalty-and-canteen.md |
| `0x1CA8D` | seg000+0xCA8D | 11 | 8 | re/00-master-index.md、re/31-experience-and-skills.md、re/52-trainer-facility.md |
| `0x115E5` | seg000+0x15E5 | 199 | 7 | re/00-master-index.md、re/03-boot-and-asset-loading.md、re/05-storage-layer.md、re/09-msq-map-structure.md、re/16-msq-block-layout.md、re/30-save-layout.md |
| `0x119DB` | seg000+0x19DB | 53 | 7 | re/06-resource-directory.md、re/45-item-data-and-weapon-damage.md |
| `0x11B83` | seg000+0x1B83 | 165 | 7 | re/00-master-index.md、re/11-huffman-decoder.md |
| `0x17868` | seg000+0x7868 | 6 | 7 | re/40-combat-screen.md |
| `0x17CFF` | seg000+0x7CFF | 53 | 7 | re/00-master-index.md、re/29-map-event-handlers.md、re/46-typed-answers-and-text-input.md、re/55-radiation-and-armour-bypass.md、re/59-playtest-against-original.md、re/62-fourth-gate-terrain-blocking.md、re/68-cell-rewrite.md、re/69-gate-flags.md、re/70-nibble1-and-facility-entry.md、re/71-nibble12-batch-patch.md、re/73-shop-and-doctor-entry.md |
| `0x17DC7` | seg000+0x7DC7 | 25 | 7 | re/26-movement-and-triggers.md、re/29-map-event-handlers.md、re/70-nibble1-and-facility-entry.md |
| `0x18DB4` | seg000+0x8DB4 | 10 | 7 | re/67-gate-penalty-and-canteen.md、re/69-gate-flags.md、re/74-heat-entry-and-gate-display.md |
| `0x190A8` | seg000+0x90A8 | 16 | 7 | re/00-master-index.md、re/37-enemy-records-and-hp.md |
| `0x196B2` | seg000+0x96B2 | 18 | 7 | re/37-enemy-records-and-hp.md、re/41-command-handlers.md |
| `0x199F1` | seg000+0x99F1 | 11 | 7 | re/00-master-index.md、re/22-shop-and-items.md、re/29-map-event-handlers.md、re/39-encounter-scan.md、re/45-item-data-and-weapon-damage.md |
| `0x19AC8` | seg000+0x9AC8 | 5 | 7 | re/00-master-index.md |
| `0x19B81` | seg000+0x9B81 | 49 | 7 | re/22-shop-and-items.md、re/42-facility-loops.md |
| `0x19BC0` | seg000+0x9BC0 | 44 | 7 | re/00-master-index.md、re/15-character-record.md、re/19-effects-and-damage.md、re/20-combat-resolution.md、re/31-experience-and-skills.md、re/32-skill-checks-and-xp.md |
| `0x1A3E1` | seg000+0xA3E1 | 17 | 7 | re/17-packed-text.md |
| `0x1CA98` | seg000+0xCA98 | 17 | 7 | re/00-master-index.md、re/31-experience-and-skills.md |
| `0x11AE8` | seg000+0x1AE8 | 116 | 6 | re/10-huffman-compression.md、re/11-huffman-decoder.md、re/12-msq-tail-and-text-model.md |
| `0x12619` | seg000+0x2619 | 29 | 6 | re/43-input-and-hotkeys.md、re/51-encounter-driver.md、re/64-enter-location-prompt.md |
| `0x12AC5` | seg000+0x2AC5 | 13 | 6 | re/39-encounter-scan.md |
| `0x13878` | seg000+0x3878 | 56 | 6 | re/00-master-index.md、re/39-encounter-scan.md、re/45-item-data-and-weapon-damage.md |
| `0x13924` | seg000+0x3924 | 21 | 6 | re/38-combat-commands-and-flee.md |
| `0x13939` | seg000+0x3939 | 5 | 6 | re/40-combat-screen.md |
| `0x169B1` | seg000+0x69B1 | 11 | 6 | re/00-master-index.md、re/26-movement-and-triggers.md、re/29-map-event-handlers.md、re/46-typed-answers-and-text-input.md、re/55-radiation-and-armour-bypass.md、re/59-playtest-against-original.md、re/60-teleport-and-map-change.md、re/68-cell-rewrite.md、re/70-nibble1-and-facility-entry.md、re/71-nibble12-batch-patch.md、re/73-shop-and-doctor-entry.md、re/75-desert-heat-entry.md |
| `0x17029` | seg000+0x7029 | 5 | 6 | re/40-combat-screen.md、re/52-trainer-facility.md |
| `0x171E3` | seg000+0x71E3 | 37 | 6 | re/39-encounter-scan.md |
| `0x173B0` | seg000+0x73B0 | 34 | 6 | re/00-master-index.md、re/38-combat-commands-and-flee.md、re/43-input-and-hotkeys.md、re/51-encounter-driver.md |
| `0x176D0` | seg000+0x76D0 | 47 | 6 | re/19-effects-and-damage.md、re/42-facility-loops.md |
| `0x18E5F` | seg000+0x8E5F | 12 | 6 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md、re/21-attributes.md |
| `0x18EFE` | seg000+0x8EFE | 262 | 6 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md、re/14-fonts-and-text-encoding.md、re/26-movement-and-triggers.md、re/43-input-and-hotkeys.md、re/46-typed-answers-and-text-input.md、re/47-dosbox-oracle.md、re/59-playtest-against-original.md |
| `0x19D86` | seg000+0x9D86 | 44 | 6 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md、re/19-effects-and-damage.md、re/20-combat-resolution.md、re/45-item-data-and-weapon-damage.md、re/55-radiation-and-armour-bypass.md、re/67-gate-penalty-and-canteen.md |
| `0x19E2A` | seg000+0x9E2A | 6 | 6 | re/14-fonts-and-text-encoding.md、re/35-status-and-healing.md、re/41-command-handlers.md、re/42-facility-loops.md |
| `0x19EB4` | seg000+0x9EB4 | 4 | 6 | re/26-movement-and-triggers.md、re/38-combat-commands-and-flee.md、re/51-encounter-driver.md |
| `0x1B0DA` | seg000+0xB0DA | 8 | 6 | re/36-combat-rounds.md |
| `0x1C213` | seg000+0xC213 | 15 | 6 | re/17-packed-text.md、re/35-status-and-healing.md、re/42-facility-loops.md |
| `0x1C22F` | seg000+0xC22F | 36 | 6 | re/42-facility-loops.md |
| `0x1003F` | seg000+0x3F | 3 | 5 | re/04-overlay-wla-bin.md |
| `0x1142B` | seg000+0x142B | 13 | 5 | re/34-map-script-opcodes.md、re/44-audio.md、re/60-teleport-and-map-change.md、re/76-script-opcode-coverage.md |
| `0x129E9` | seg000+0x29E9 | 87 | 5 | re/16-msq-block-layout.md、re/47-dosbox-oracle.md |
| `0x142D7` | seg000+0x42D7 | 11 | 5 | re/65-third-gate-conditions.md、re/67-gate-penalty-and-canteen.md、re/69-gate-flags.md |
| `0x14B93` | seg000+0x4B93 | 9 | 5 | re/39-encounter-scan.md |
| `0x15C8C` | seg000+0x5C8C | 10 | 5 | re/45-item-data-and-weapon-damage.md |
| `0x15DAF` | seg000+0x5DAF | 21 | 5 | re/60-teleport-and-map-change.md |
| `0x16619` | seg000+0x6619 | 45 | 5 | re/00-master-index.md、re/25-screen-layout.md、re/26-movement-and-triggers.md |
| `0x169EB` | seg000+0x69EB | 13 | 5 | re/00-master-index.md、re/16-msq-block-layout.md、re/26-movement-and-triggers.md、re/62-fourth-gate-terrain-blocking.md、re/64-enter-location-prompt.md |
| `0x16D1A` | seg000+0x6D1A | 14 | 5 | re/26-movement-and-triggers.md、re/29-map-event-handlers.md、re/55-radiation-and-armour-bypass.md、re/62-fourth-gate-terrain-blocking.md、re/66-nibble2-event-and-heat.md、re/69-gate-flags.md |
| `0x16F4C` | seg000+0x6F4C | 27 | 5 | re/00-master-index.md、re/22-shop-and-items.md、re/32-skill-checks-and-xp.md |
| `0x16F70` | seg000+0x6F70 | 185 | 5 | re/04-overlay-wla-bin.md、re/25-screen-layout.md |
| `0x176A8` | seg000+0x76A8 | 14 | 5 | re/27-game-clock.md、re/28-text-variants.md |
| `0x17920` | seg000+0x7920 | 3 | 5 | re/26-movement-and-triggers.md、re/29-map-event-handlers.md、re/55-radiation-and-armour-bypass.md、re/70-nibble1-and-facility-entry.md、re/71-nibble12-batch-patch.md |
| `0x17A6B` | seg000+0x7A6B | 62 | 5 | re/25-screen-layout.md |
| `0x17B15` | seg000+0x7B15 | 41 | 5 | re/15-character-record.md、re/19-effects-and-damage.md、re/22-shop-and-items.md、re/42-facility-loops.md |
| `0x17D50` | seg000+0x7D50 | 42 | 5 | re/46-typed-answers-and-text-input.md、re/68-cell-rewrite.md、re/69-gate-flags.md |
| `0x17FC8` | seg000+0x7FC8 | 38 | 5 | re/00-master-index.md、re/12-msq-tail-and-text-model.md、re/24-map-layers-and-tiles.md |
| `0x17FEE` | seg000+0x7FEE | 40 | 5 | re/00-master-index.md、re/24-map-layers-and-tiles.md、re/25-screen-layout.md |
| `0x18024` | seg000+0x8024 | 171 | 5 | re/00-master-index.md、re/16-msq-block-layout.md、re/25-screen-layout.md、re/27-game-clock.md、re/48-map-icons.md |
| `0x183B1` | seg000+0x83B1 | 110 | 5 | re/05-storage-layer.md、re/06-resource-directory.md、re/07-msq-blocks.md、re/10-huffman-compression.md、re/27-game-clock.md |
| `0x18860` | seg000+0x8860 | 200 | 5 | re/43-input-and-hotkeys.md |
| `0x189B1` | seg000+0x89B1 | 158 | 5 | re/04-overlay-wla-bin.md |
| `0x18D85` | seg000+0x8D85 | 9 | 5 | re/00-master-index.md、re/37-enemy-records-and-hp.md |
| `0x190A6` | seg000+0x90A6 | 2 | 5 | re/00-master-index.md、re/29-map-event-handlers.md、re/52-trainer-facility.md、re/72-facility-entry-and-command-bar.md |
| `0x19CAC` | seg000+0x9CAC | 32 | 5 | re/32-skill-checks-and-xp.md |
| `0x1B7B7` | seg000+0xB7B7 | 3 | 5 | re/29-map-event-handlers.md |
| `0x1BE31` | seg000+0xBE31 | 15 | 5 | re/17-packed-text.md、re/29-map-event-handlers.md、re/35-status-and-healing.md、re/52-trainer-facility.md |
| `0x1CBD3` | seg001+0x6C | 86 | 5 | re/00-master-index.md、re/26-movement-and-triggers.md、re/34-map-script-opcodes.md、re/44-audio.md、re/60-teleport-and-map-change.md、re/76-script-opcode-coverage.md |
| `0x1000C` | seg000+0xC | 3 | 4 | re/04-overlay-wla-bin.md、re/48-map-icons.md、re/49-save-roundtrip-on-hardware.md |
| `0x116AC` | seg000+0x16AC | 132 | 4 | re/03-boot-and-asset-loading.md、re/05-storage-layer.md |
| `0x1272E` | seg000+0x272E | 6 | 4 | re/51-encounter-driver.md |
| `0x12A4C` | seg000+0x2A4C | 42 | 4 | re/16-msq-block-layout.md、re/28-text-variants.md、re/32-skill-checks-and-xp.md、re/37-enemy-records-and-hp.md、re/48-map-icons.md |
| `0x14085` | seg000+0x4085 | 11 | 4 | re/26-movement-and-triggers.md、re/62-fourth-gate-terrain-blocking.md、re/65-third-gate-conditions.md |
| `0x1417F` | seg000+0x417F | 11 | 4 | re/26-movement-and-triggers.md、re/32-skill-checks-and-xp.md |
| `0x1418A` | seg000+0x418A | 9 | 4 | re/32-skill-checks-and-xp.md |
| `0x163F6` | seg000+0x63F6 | 26 | 4 | re/26-movement-and-triggers.md、re/70-nibble1-and-facility-entry.md |
| `0x16646` | seg000+0x6646 | 47 | 4 | re/00-master-index.md、re/25-screen-layout.md |
| `0x169CF` | seg000+0x69CF | 28 | 4 | re/00-master-index.md、re/25-screen-layout.md |
| `0x16C6F` | seg000+0x6C6F | 13 | 4 | re/25-screen-layout.md |
| `0x17451` | seg000+0x7451 | 102 | 4 | re/00-master-index.md、re/14-fonts-and-text-encoding.md |
| `0x17574` | seg000+0x7574 | 271 | 4 | re/00-master-index.md、re/14-fonts-and-text-encoding.md、re/29-map-event-handlers.md、re/72-facility-entry-and-command-bar.md |
| `0x17AF5` | seg000+0x7AF5 | 32 | 4 | re/41-command-handlers.md、re/42-facility-loops.md |
| `0x17DF1` | seg000+0x7DF1 | 81 | 4 | re/00-master-index.md、re/27-game-clock.md |
| `0x17E42` | seg000+0x7E42 | 390 | 4 | re/04-overlay-wla-bin.md |
| `0x18134` | seg000+0x8134 | 18 | 4 | re/00-master-index.md、re/32-skill-checks-and-xp.md |
| `0x18801` | seg000+0x8801 | 89 | 4 | re/10-huffman-compression.md、re/27-game-clock.md、re/29-map-event-handlers.md、re/30-save-layout.md、re/72-facility-entry-and-command-bar.md |
| `0x18DCE` | seg000+0x8DCE | 115 | 4 | re/27-game-clock.md、re/28-text-variants.md |
| `0x190B8` | seg000+0x90B8 | 64 | 4 | re/38-combat-commands-and-flee.md |
| `0x194E8` | seg000+0x94E8 | 21 | 4 | re/32-skill-checks-and-xp.md |
| `0x1968A` | seg000+0x968A | 2 | 4 | re/22-shop-and-items.md、re/42-facility-loops.md |
| `0x19895` | seg000+0x9895 | 39 | 4 | re/42-facility-loops.md |
| `0x19A1D` | seg000+0x9A1D | 36 | 4 | re/00-master-index.md、re/15-character-record.md、re/19-effects-and-damage.md |
| `0x19CCC` | seg000+0x9CCC | 29 | 4 | re/20-combat-resolution.md、re/32-skill-checks-and-xp.md |
| `0x19CE9` | seg000+0x9CE9 | 21 | 4 | re/19-effects-and-damage.md、re/35-status-and-healing.md |
| `0x1BE16` | seg000+0xBE16 | 9 | 4 | re/52-trainer-facility.md |
| `0x1BE1F` | seg000+0xBE1F | 11 | 4 | re/52-trainer-facility.md |
| `0x1C68E` | seg000+0xC68E | 43 | 4 | re/00-master-index.md、re/31-experience-and-skills.md |
| `0x1000F` | seg000+0xF | 3 | 3 | re/04-overlay-wla-bin.md、re/14-fonts-and-text-encoding.md、re/58-line-flush-and-scrollback.md |
| `0x10D4D` | seg000+0xD4D | 181 | 3 | re/04-overlay-wla-bin.md |
| `0x10E02` | seg000+0xE02 | 67 | 3 | re/04-overlay-wla-bin.md |
| `0x11730` | seg000+0x1730 | 292 | 3 | re/03-boot-and-asset-loading.md、re/05-storage-layer.md、re/30-save-layout.md |
| `0x11A10` | seg000+0x1A10 | 73 | 3 | re/06-resource-directory.md、re/07-msq-blocks.md、re/08-msq-encryption.md、re/22-shop-and-items.md、re/30-save-layout.md、re/45-item-data-and-weapon-damage.md |
| `0x11AA3` | seg000+0x1AA3 | 69 | 3 | re/09-msq-map-structure.md、re/30-save-layout.md |
| `0x11C28` | seg000+0x1C28 | 44 | 3 | re/10-huffman-compression.md、re/11-huffman-decoder.md |
| `0x11C54` | seg000+0x1C54 | 60 | 3 | re/10-huffman-compression.md、re/11-huffman-decoder.md |
| `0x123D1` | seg000+0x23D1 | 17 | 3 | re/38-combat-commands-and-flee.md |
| `0x125BD` | seg000+0x25BD | 85 | 3 | re/38-combat-commands-and-flee.md |
| `0x12738` | seg000+0x2738 | 11 | 3 | re/24-map-layers-and-tiles.md、re/41-command-handlers.md |
| `0x12ABA` | seg000+0x2ABA | 11 | 3 | re/32-skill-checks-and-xp.md、re/36-combat-rounds.md、re/37-enemy-records-and-hp.md |
| `0x1354F` | seg000+0x354F | 18 | 3 | re/51-encounter-driver.md |
| `0x14175` | seg000+0x4175 | 5 | 3 | re/00-master-index.md、re/69-gate-flags.md |
| `0x1417A` | seg000+0x417A | 5 | 3 | re/00-master-index.md、re/65-third-gate-conditions.md、re/67-gate-penalty-and-canteen.md、re/69-gate-flags.md |
| `0x14193` | seg000+0x4193 | 103 | 3 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md、re/19-effects-and-damage.md、re/55-radiation-and-armour-bypass.md、re/65-third-gate-conditions.md、re/66-nibble2-event-and-heat.md、re/67-gate-penalty-and-canteen.md、re/69-gate-flags.md、re/74-heat-entry-and-gate-display.md |
| `0x142ED` | seg000+0x42ED | 39 | 3 | re/00-master-index.md、re/65-third-gate-conditions.md、re/67-gate-penalty-and-canteen.md、re/69-gate-flags.md、re/74-heat-entry-and-gate-display.md |
| `0x14659` | seg000+0x4659 | 6 | 3 | re/37-enemy-records-and-hp.md |
| `0x14A88` | seg000+0x4A88 | 34 | 3 | re/00-master-index.md、re/39-encounter-scan.md |
| `0x14AEA` | seg000+0x4AEA | 76 | 3 | re/00-master-index.md、re/39-encounter-scan.md |
| `0x15755` | seg000+0x5755 | 88 | 3 | re/00-master-index.md、re/20-combat-resolution.md、re/21-attributes.md、re/45-item-data-and-weapon-damage.md |
| `0x157D6` | seg000+0x57D6 | 342 | 3 | re/00-master-index.md、re/13-rng.md、re/15-character-record.md、re/19-effects-and-damage.md、re/20-combat-resolution.md、re/55-radiation-and-armour-bypass.md、re/66-nibble2-event-and-heat.md |
| `0x159C7` | seg000+0x59C7 | 81 | 3 | re/00-master-index.md、re/20-combat-resolution.md、re/32-skill-checks-and-xp.md |
| `0x15A30` | seg000+0x5A30 | 84 | 3 | re/20-combat-resolution.md |
| `0x15A84` | seg000+0x5A84 | 25 | 3 | re/19-effects-and-damage.md、re/20-combat-resolution.md、re/28-text-variants.md |
| `0x15BFE` | seg000+0x5BFE | 5 | 3 | re/19-effects-and-damage.md |
| `0x15C96` | seg000+0x5C96 | 14 | 3 | re/41-command-handlers.md |
| `0x163C1` | seg000+0x63C1 | 3 | 3 | re/26-movement-and-triggers.md、re/51-encounter-driver.md |
| `0x1649E` | seg000+0x649E | 57 | 3 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/26-movement-and-triggers.md |
| `0x166D3` | seg000+0x66D3 | 67 | 3 | re/06-resource-directory.md |
| `0x16716` | seg000+0x6716 | 70 | 3 | re/00-master-index.md、re/47-dosbox-oracle.md、re/48-map-icons.md、re/49-save-roundtrip-on-hardware.md |
| `0x1676A` | seg000+0x676A | 99 | 3 | re/00-master-index.md、re/26-movement-and-triggers.md、re/27-game-clock.md |
| `0x16F35` | seg000+0x6F35 | 23 | 3 | re/41-command-handlers.md |
| `0x17533` | seg000+0x7533 | 49 | 3 | re/14-fonts-and-text-encoding.md |
| `0x17564` | seg000+0x7564 | 16 | 3 | re/72-facility-entry-and-command-bar.md |
| `0x178B9` | seg000+0x78B9 | 82 | 3 | re/17-packed-text.md |
| `0x1790B` | seg000+0x790B | 21 | 3 | re/16-msq-block-layout.md、re/17-packed-text.md、re/18-block-text.md、re/74-heat-entry-and-gate-display.md |
| `0x17923` | seg000+0x7923 | 148 | 3 | re/25-screen-layout.md |
| `0x17B3E` | seg000+0x7B3E | 30 | 3 | re/15-character-record.md、re/19-effects-and-damage.md、re/22-shop-and-items.md、re/32-skill-checks-and-xp.md、re/65-third-gate-conditions.md |
| `0x17B8F` | seg000+0x7B8F | 56 | 3 | re/00-master-index.md、re/17-packed-text.md、re/46-typed-answers-and-text-input.md |
| `0x17C72` | seg000+0x7C72 | 63 | 3 | re/00-master-index.md、re/24-map-layers-and-tiles.md、re/39-encounter-scan.md、re/46-typed-answers-and-text-input.md、re/68-cell-rewrite.md |
| `0x18016` | seg000+0x8016 | 14 | 3 | re/26-movement-and-triggers.md、re/27-game-clock.md |
| `0x1818E` | seg000+0x818E | 126 | 3 | re/00-master-index.md、re/32-skill-checks-and-xp.md |
| `0x182FA` | seg000+0x82FA | 20 | 3 | re/20-combat-resolution.md、re/21-attributes.md、re/45-item-data-and-weapon-damage.md |
| `0x18DBE` | seg000+0x8DBE | 16 | 3 | re/74-heat-entry-and-gate-display.md |
| `0x18E6B` | seg000+0x8E6B | 37 | 3 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/13-rng.md |
| `0x19004` | seg000+0x9004 | 52 | 3 | re/14-fonts-and-text-encoding.md |
| `0x19394` | seg000+0x9394 | 192 | 3 | re/21-attributes.md、re/41-command-handlers.md、re/43-input-and-hotkeys.md |
| `0x197BB` | seg000+0x97BB | 87 | 3 | re/00-master-index.md、re/25-screen-layout.md |
| `0x198F0` | seg000+0x98F0 | 139 | 3 | re/43-input-and-hotkeys.md |
| `0x1B15F` | seg000+0xB15F | 10 | 3 | re/20-combat-resolution.md、re/36-combat-rounds.md、re/37-enemy-records-and-hp.md |
| `0x1BDFF` | seg000+0xBDFF | 23 | 3 | re/52-trainer-facility.md |
| `0x1BE2A` | seg000+0xBE2A | 7 | 3 | re/52-trainer-facility.md |
| `0x1BE40` | seg000+0xBE40 | 13 | 3 | re/52-trainer-facility.md |
| `0x1C548` | seg000+0xC548 | 18 | 3 | re/42-facility-loops.md |
| `0x1C9DE` | seg000+0xC9DE | 54 | 3 | re/21-attributes.md、re/45-item-data-and-weapon-damage.md |
| `0x1CA14` | seg000+0xCA14 | 121 | 3 | re/46-typed-answers-and-text-input.md |
| `0x10000` | seg000+0x0 | 3 | 2 | re/00-master-index.md、re/02-exepack-unpack.md、re/03-boot-and-asset-loading.md、re/04-overlay-wla-bin.md、re/06-resource-directory.md、re/61-map-id-table.md |
| `0x10042` | seg000+0x42 | 3 | 2 | re/04-overlay-wla-bin.md |
| `0x10045` | seg000+0x45 | 3 | 2 | re/04-overlay-wla-bin.md |
| `0x11CA4` | seg000+0x1CA4 | 35 | 2 | re/10-huffman-compression.md、re/11-huffman-decoder.md |
| `0x12262` | seg000+0x2262 | 34 | 2 | re/41-command-handlers.md |
| `0x12440` | seg000+0x2440 | 222 | 2 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/27-game-clock.md、re/35-status-and-healing.md |
| `0x1251E` | seg000+0x251E | 25 | 2 | re/35-status-and-healing.md |
| `0x12537` | seg000+0x2537 | 19 | 2 | re/27-game-clock.md、re/35-status-and-healing.md |
| `0x1254A` | seg000+0x254A | 7 | 2 | re/27-game-clock.md、re/35-status-and-healing.md |
| `0x1295D` | seg000+0x295D | 89 | 2 | re/20-combat-resolution.md |
| `0x12A76` | seg000+0x2A76 | 23 | 2 | re/00-master-index.md、re/20-combat-resolution.md、re/22-shop-and-items.md、re/32-skill-checks-and-xp.md |
| `0x12A9D` | seg000+0x2A9D | 14 | 2 | re/37-enemy-records-and-hp.md |
| `0x12AAB` | seg000+0x2AAB | 15 | 2 | re/32-skill-checks-and-xp.md、re/37-enemy-records-and-hp.md |
| `0x12C80` | seg000+0x2C80 | 54 | 2 | re/00-master-index.md、re/29-map-event-handlers.md、re/34-map-script-opcodes.md、re/60-teleport-and-map-change.md、re/70-nibble1-and-facility-entry.md、re/71-nibble12-batch-patch.md |
| `0x1379E` | seg000+0x379E | 48 | 2 | re/00-master-index.md、re/24-map-layers-and-tiles.md、re/39-encounter-scan.md |
| `0x138B0` | seg000+0x38B0 | 17 | 2 | re/37-enemy-records-and-hp.md、re/39-encounter-scan.md |
| `0x13AE4` | seg000+0x3AE4 | 366 | 2 | re/21-attributes.md、re/41-command-handlers.md |
| `0x13C58` | seg000+0x3C58 | 443 | 2 | re/66-nibble2-event-and-heat.md、re/68-cell-rewrite.md、re/70-nibble1-and-facility-entry.md |
| `0x13EC0` | seg000+0x3EC0 | 9 | 2 | re/65-third-gate-conditions.md |
| `0x1407E` | seg000+0x407E | 7 | 2 | re/65-third-gate-conditions.md |
| `0x141FA` | seg000+0x41FA | 156 | 2 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/19-effects-and-damage.md、re/35-status-and-healing.md、re/55-radiation-and-armour-bypass.md、re/66-nibble2-event-and-heat.md、re/67-gate-penalty-and-canteen.md |
| `0x14296` | seg000+0x4296 | 27 | 2 | re/00-master-index.md、re/66-nibble2-event-and-heat.md、re/69-gate-flags.md |
| `0x142B1` | seg000+0x42B1 | 38 | 2 | re/00-master-index.md、re/69-gate-flags.md |
| `0x14314` | seg000+0x4314 | 22 | 2 | re/69-gate-flags.md、re/74-heat-entry-and-gate-display.md |
| `0x1465F` | seg000+0x465F | 5 | 2 | re/51-encounter-driver.md |
| `0x14664` | seg000+0x4664 | 915 | 2 | re/00-master-index.md、re/24-map-layers-and-tiles.md、re/37-enemy-records-and-hp.md、re/38-combat-commands-and-flee.md、re/39-encounter-scan.md、re/48-map-icons.md、re/51-encounter-driver.md |
| `0x149F7` | seg000+0x49F7 | 110 | 2 | re/00-master-index.md、re/39-encounter-scan.md |
| `0x14A65` | seg000+0x4A65 | 35 | 2 | re/00-master-index.md、re/39-encounter-scan.md |
| `0x14AAA` | seg000+0x4AAA | 29 | 2 | re/39-encounter-scan.md |
| `0x14AC7` | seg000+0x4AC7 | 26 | 2 | re/39-encounter-scan.md |
| `0x14B36` | seg000+0x4B36 | 93 | 2 | re/37-enemy-records-and-hp.md |
| `0x14B9C` | seg000+0x4B9C | 17 | 2 | re/37-enemy-records-and-hp.md |
| `0x14F5D` | seg000+0x4F5D | 10 | 2 | re/37-enemy-records-and-hp.md、re/39-encounter-scan.md |
| `0x14FDE` | seg000+0x4FDE | 88 | 2 | re/37-enemy-records-and-hp.md、re/39-encounter-scan.md |
| `0x15705` | seg000+0x5705 | 51 | 2 | re/00-master-index.md、re/20-combat-resolution.md、re/21-attributes.md |
| `0x1597E` | seg000+0x597E | 73 | 2 | re/20-combat-resolution.md |
| `0x15A9D` | seg000+0x5A9D | 242 | 2 | re/00-remake-knowledge-gaps.md、re/20-combat-resolution.md |
| `0x15B9C` | seg000+0x5B9C | 15 | 2 | re/20-combat-resolution.md |
| `0x15C19` | seg000+0x5C19 | 14 | 2 | re/20-combat-resolution.md |
| `0x15CA4` | seg000+0x5CA4 | 14 | 2 | re/41-command-handlers.md |
| `0x15CC0` | seg000+0x5CC0 | 17 | 2 | re/20-combat-resolution.md |
| `0x15CD1` | seg000+0x5CD1 | 14 | 2 | re/20-combat-resolution.md |
| `0x162E1` | seg000+0x62E1 | 43 | 2 | re/00-master-index.md、re/44-audio.md |
| `0x16428` | seg000+0x6428 | 118 | 2 | re/00-master-index.md、re/48-map-icons.md |
| `0x164E8` | seg000+0x64E8 | 5 | 2 | re/26-movement-and-triggers.md |
| `0x1652D` | seg000+0x652D | 56 | 2 | re/26-movement-and-triggers.md |
| `0x167CE` | seg000+0x67CE | 105 | 2 | re/00-master-index.md、re/26-movement-and-triggers.md、re/48-map-icons.md |
| `0x16890` | seg000+0x6890 | 280 | 2 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/26-movement-and-triggers.md、re/39-encounter-scan.md、re/74-heat-entry-and-gate-display.md |
| `0x16AC7` | seg000+0x6AC7 | 14 | 2 | re/60-teleport-and-map-change.md |
| `0x16B17` | seg000+0x6B17 | 13 | 2 | re/43-input-and-hotkeys.md、re/60-teleport-and-map-change.md、re/64-enter-location-prompt.md |
| `0x16CBD` | seg000+0x6CBD | 10 | 2 | re/17-packed-text.md、re/40-combat-screen.md |
| `0x16D30` | seg000+0x6D30 | 4 | 2 | re/53-list-framework.md |
| `0x16E43` | seg000+0x6E43 | 103 | 2 | re/53-list-framework.md |
| `0x16EAA` | seg000+0x6EAA | 19 | 2 | re/53-list-framework.md |
| `0x17357` | seg000+0x7357 | 51 | 2 | re/40-combat-screen.md |
| `0x17690` | seg000+0x7690 | 6 | 2 | re/42-facility-loops.md |
| `0x17B5C` | seg000+0x7B5C | 24 | 2 | re/22-shop-and-items.md、re/45-item-data-and-weapon-damage.md |
| `0x17BC7` | seg000+0x7BC7 | 64 | 2 | re/00-master-index.md、re/17-packed-text.md |
| `0x17D34` | seg000+0x7D34 | 14 | 2 | re/00-master-index.md、re/46-typed-answers-and-text-input.md、re/69-gate-flags.md |
| `0x17D47` | seg000+0x7D47 | 9 | 2 | re/46-typed-answers-and-text-input.md、re/69-gate-flags.md |
| `0x17DE0` | seg000+0x7DE0 | 17 | 2 | re/60-teleport-and-map-change.md |
| `0x180CF` | seg000+0x80CF | 15 | 2 | re/48-map-icons.md |
| `0x180F0` | seg000+0x80F0 | 68 | 2 | re/00-master-index.md、re/32-skill-checks-and-xp.md、re/65-third-gate-conditions.md |
| `0x18146` | seg000+0x8146 | 72 | 2 | re/00-master-index.md、re/32-skill-checks-and-xp.md |
| `0x1820C` | seg000+0x820C | 124 | 2 | re/00-master-index.md、re/32-skill-checks-and-xp.md、re/65-third-gate-conditions.md |
| `0x1830E` | seg000+0x830E | 12 | 2 | re/20-combat-resolution.md |
| `0x184E8` | seg000+0x84E8 | 252 | 2 | re/00-master-index.md、re/05-storage-layer.md、re/07-msq-blocks.md、re/10-huffman-compression.md、re/23-picture-format.md、re/29-map-event-handlers.md、re/37-enemy-records-and-hp.md |
| `0x185E6` | seg000+0x85E6 | 110 | 2 | re/10-huffman-compression.md |
| `0x19362` | seg000+0x9362 | 50 | 2 | re/15-character-record.md、re/17-packed-text.md |
| `0x19454` | seg000+0x9454 | 74 | 2 | re/00-master-index.md、re/15-character-record.md |
| `0x1949E` | seg000+0x949E | 74 | 2 | re/00-master-index.md、re/45-item-data-and-weapon-damage.md |
| `0x1963A` | seg000+0x963A | 80 | 2 | re/15-character-record.md、re/41-command-handlers.md |
| `0x196C4` | seg000+0x96C4 | 5 | 2 | re/55-radiation-and-armour-bypass.md、re/59-playtest-against-original.md |
| `0x197AE` | seg000+0x97AE | 13 | 2 | re/40-combat-screen.md |
| `0x19812` | seg000+0x9812 | 2 | 2 | re/40-combat-screen.md |
| `0x1997B` | seg000+0x997B | 94 | 2 | re/00-master-index.md、re/15-character-record.md |
| `0x19A0F` | seg000+0x9A0F | 13 | 2 | re/00-master-index.md、re/45-item-data-and-weapon-damage.md |
| `0x19A41` | seg000+0x9A41 | 23 | 2 | re/15-character-record.md |
| `0x19A58` | seg000+0x9A58 | 68 | 2 | re/00-master-index.md、re/32-skill-checks-and-xp.md、re/44-audio.md、re/65-third-gate-conditions.md、re/67-gate-penalty-and-canteen.md |
| `0x19B67` | seg000+0x9B67 | 26 | 2 | re/31-experience-and-skills.md、re/51-encounter-driver.md |
| `0x19CFE` | seg000+0x9CFE | 16 | 2 | re/42-facility-loops.md |
| `0x1A00A` | seg000+0xA00A | 53 | 2 | re/28-text-variants.md |
| `0x1A045` | seg000+0xA045 | 51 | 2 | re/28-text-variants.md |
| `0x1A07E` | seg000+0xA07E | 71 | 2 | re/28-text-variants.md |
| `0x1A0C5` | seg000+0xA0C5 | 298 | 2 | re/00-master-index.md、re/14-fonts-and-text-encoding.md、re/43-input-and-hotkeys.md、re/46-typed-answers-and-text-input.md、re/58-line-flush-and-scrollback.md |
| `0x1B0E2` | seg000+0xB0E2 | 15 | 2 | re/36-combat-rounds.md |
| `0x1B0F1` | seg000+0xB0F1 | 8 | 2 | re/20-combat-resolution.md |
| `0x1B0F9` | seg000+0xB0F9 | 15 | 2 | re/36-combat-rounds.md |
| `0x1B108` | seg000+0xB108 | 87 | 2 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/20-combat-resolution.md、re/21-attributes.md |
| `0x1B8A0` | seg000+0xB8A0 | 13 | 2 | re/31-experience-and-skills.md |
| `0x1BA72` | seg000+0xBA72 | 192 | 2 | re/00-master-index.md、re/28-text-variants.md、re/31-experience-and-skills.md |
| `0x1C1C2` | seg000+0xC1C2 | 10 | 2 | re/22-shop-and-items.md、re/42-facility-loops.md |
| `0x1C1CC` | seg000+0xC1CC | 64 | 2 | re/00-master-index.md、re/22-shop-and-items.md |
| `0x1C510` | seg000+0xC510 | 56 | 2 | re/35-status-and-healing.md、re/42-facility-loops.md |
| `0x1CAA9` | seg000+0xCAA9 | 23 | 2 | re/46-typed-answers-and-text-input.md |
| `0x1CAC0` | seg000+0xCAC0 | 17 | 2 | re/46-typed-answers-and-text-input.md |
| `0x1CAD1` | seg000+0xCAD1 | 83 | 2 | re/00-master-index.md、re/21-attributes.md |
| `0x10003` | seg000+0x3 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10006` | seg000+0x6 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10009` | seg000+0x9 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10015` | seg000+0x15 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x1001E` | seg000+0x1E | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10021` | seg000+0x21 | 3 | 1 | re/04-overlay-wla-bin.md、re/26-movement-and-triggers.md |
| `0x10024` | seg000+0x24 | 3 | 1 | re/26-movement-and-triggers.md |
| `0x10027` | seg000+0x27 | 3 | 1 | re/26-movement-and-triggers.md |
| `0x1002A` | seg000+0x2A | 3 | 1 | re/04-overlay-wla-bin.md、re/26-movement-and-triggers.md |
| `0x1002D` | seg000+0x2D | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10030` | seg000+0x30 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10033` | seg000+0x33 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10036` | seg000+0x36 | 3 | 1 | re/04-overlay-wla-bin.md、re/34-map-script-opcodes.md |
| `0x1003C` | seg000+0x3C | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x10048` | seg000+0x48 | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x1004B` | seg000+0x4B | 3 | 1 | re/04-overlay-wla-bin.md |
| `0x1004E` | seg000+0x4E | 58 | 1 | re/04-overlay-wla-bin.md、re/56-transtbl.md |
| `0x10088` | seg000+0x88 | 188 | 1 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/04-overlay-wla-bin.md、re/24-map-layers-and-tiles.md、re/56-transtbl.md |
| `0x10144` | seg000+0x144 | 18 | 1 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/04-overlay-wla-bin.md、re/23-picture-format.md |
| `0x10156` | seg000+0x156 | 325 | 1 | re/00-master-index.md、re/04-overlay-wla-bin.md、re/56-transtbl.md |
| `0x1029B` | seg000+0x29B | 881 | 1 | re/00-master-index.md、re/04-overlay-wla-bin.md、re/24-map-layers-and-tiles.md、re/48-map-icons.md、re/56-transtbl.md |
| `0x1060C` | seg000+0x60C | 189 | 1 | re/00-master-index.md、re/04-overlay-wla-bin.md、re/14-fonts-and-text-encoding.md、re/56-transtbl.md |
| `0x10762` | seg000+0x762 | 118 | 1 | re/00-master-index.md、re/04-overlay-wla-bin.md、re/25-screen-layout.md |
| `0x107D8` | seg000+0x7D8 | 161 | 1 | re/04-overlay-wla-bin.md |
| `0x108A3` | seg000+0x8A3 | 92 | 1 | re/04-overlay-wla-bin.md |
| `0x109BC` | seg000+0x9BC | 94 | 1 | re/04-overlay-wla-bin.md |
| `0x10A1A` | seg000+0xA1A | 96 | 1 | re/04-overlay-wla-bin.md |
| `0x10A7A` | seg000+0xA7A | 151 | 1 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/04-overlay-wla-bin.md、re/23-picture-format.md、re/54-facility-screen-layout.md |
| `0x10B11` | seg000+0xB11 | 329 | 1 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/04-overlay-wla-bin.md、re/23-picture-format.md、re/54-facility-screen-layout.md |
| `0x10C5A` | seg000+0xC5A | 92 | 1 | re/04-overlay-wla-bin.md、re/56-transtbl.md |
| `0x10CB6` | seg000+0xCB6 | 151 | 1 | re/00-master-index.md、re/04-overlay-wla-bin.md、re/14-fonts-and-text-encoding.md、re/56-transtbl.md |
| `0x10E45` | seg000+0xE45 | 85 | 1 | re/04-overlay-wla-bin.md |
| `0x10F12` | seg000+0xF12 | 82 | 1 | re/04-overlay-wla-bin.md、re/24-map-layers-and-tiles.md、re/25-screen-layout.md、re/56-transtbl.md |
| `0x10F64` | seg000+0xF64 | 67 | 1 | re/04-overlay-wla-bin.md、re/25-screen-layout.md、re/56-transtbl.md |
| `0x10FA7` | seg000+0xFA7 | 44 | 1 | re/04-overlay-wla-bin.md、re/56-transtbl.md |
| `0x11438` | seg000+0x1438 | 13 | 1 | re/44-audio.md |
| `0x11854` | seg000+0x1854 | 111 | 1 | re/03-boot-and-asset-loading.md |
| `0x119A8` | seg000+0x19A8 | 50 | 1 | re/03-boot-and-asset-loading.md |
| `0x11A59` | seg000+0x1A59 | 74 | 1 | re/00-master-index.md、re/08-msq-encryption.md、re/11-huffman-decoder.md、re/16-msq-block-layout.md、re/18-block-text.md |
| `0x11C90` | seg000+0x1C90 | 20 | 1 | re/10-huffman-compression.md |
| `0x11CD0` | seg000+0x1CD0 | 678 | 1 | re/00-master-index.md、re/51-encounter-driver.md |
| `0x11F76` | seg000+0x1F76 | 279 | 1 | re/00-master-index.md、re/38-combat-commands-and-flee.md、re/40-combat-screen.md、re/51-encounter-driver.md |
| `0x12551` | seg000+0x2551 | 96 | 1 | re/51-encounter-driver.md |
| `0x12636` | seg000+0x2636 | 161 | 1 | re/37-enemy-records-and-hp.md、re/38-combat-commands-and-flee.md、re/40-combat-screen.md |
| `0x12760` | seg000+0x2760 | 24 | 1 | re/51-encounter-driver.md |
| `0x13580` | seg000+0x3580 | 70 | 1 | re/51-encounter-driver.md |
| `0x13651` | seg000+0x3651 | 70 | 1 | re/37-enemy-records-and-hp.md |
| `0x136C7` | seg000+0x36C7 | 95 | 1 | re/37-enemy-records-and-hp.md |
| `0x1372B` | seg000+0x372B | 55 | 1 | re/37-enemy-records-and-hp.md |
| `0x13762` | seg000+0x3762 | 37 | 1 | re/00-master-index.md、re/37-enemy-records-and-hp.md、re/68-cell-rewrite.md、re/70-nibble1-and-facility-entry.md、re/71-nibble12-batch-patch.md |
| `0x139CE` | seg000+0x39CE | 105 | 1 | re/43-input-and-hotkeys.md |
| `0x13E9B` | seg000+0x3E9B | 37 | 1 | re/00-master-index.md、re/26-movement-and-triggers.md、re/62-fourth-gate-terrain-blocking.md、re/65-third-gate-conditions.md、re/66-nibble2-event-and-heat.md、re/67-gate-penalty-and-canteen.md |
| `0x13EC9` | seg000+0x3EC9 | 437 | 1 | re/00-master-index.md、re/26-movement-and-triggers.md、re/32-skill-checks-and-xp.md、re/65-third-gate-conditions.md、re/66-nibble2-event-and-heat.md、re/67-gate-penalty-and-canteen.md、re/68-cell-rewrite.md、re/69-gate-flags.md |
| `0x140DD` | seg000+0x40DD | 73 | 1 | re/32-skill-checks-and-xp.md |
| `0x14126` | seg000+0x4126 | 79 | 1 | re/32-skill-checks-and-xp.md |
| `0x14480` | seg000+0x4480 | 473 | 1 | re/00-master-index.md、re/37-enemy-records-and-hp.md、re/39-encounter-scan.md、re/51-encounter-driver.md |
| `0x14BAD` | seg000+0x4BAD | 63 | 1 | re/37-enemy-records-and-hp.md |
| `0x14BF0` | seg000+0x4BF0 | 597 | 1 | re/37-enemy-records-and-hp.md、re/39-encounter-scan.md |
| `0x15036` | seg000+0x5036 | 261 | 1 | re/37-enemy-records-and-hp.md、re/39-encounter-scan.md、re/51-encounter-driver.md |
| `0x15672` | seg000+0x5672 | 13 | 1 | re/20-combat-resolution.md |
| `0x15738` | seg000+0x5738 | 29 | 1 | re/45-item-data-and-weapon-damage.md |
| `0x15A18` | seg000+0x5A18 | 24 | 1 | re/32-skill-checks-and-xp.md |
| `0x15CE0` | seg000+0x5CE0 | 78 | 1 | re/00-master-index.md、re/26-movement-and-triggers.md、re/62-fourth-gate-terrain-blocking.md、re/68-cell-rewrite.md |
| `0x160A8` | seg000+0x60A8 | 57 | 1 | re/51-encounter-driver.md |
| `0x161C0` | seg000+0x61C0 | 263 | 1 | re/25-screen-layout.md |
| `0x1630C` | seg000+0x630C | 35 | 1 | re/72-facility-entry-and-command-bar.md |
| `0x16356` | seg000+0x6356 | 32 | 1 | re/26-movement-and-triggers.md |
| `0x16390` | seg000+0x6390 | 7 | 1 | re/17-packed-text.md |
| `0x16410` | seg000+0x6410 | 23 | 1 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/26-movement-and-triggers.md |
| `0x1656D` | seg000+0x656D | 23 | 1 | re/26-movement-and-triggers.md、re/32-skill-checks-and-xp.md、re/44-audio.md |
| `0x16675` | seg000+0x6675 | 94 | 1 | re/00-master-index.md、re/25-screen-layout.md |
| `0x16AD5` | seg000+0x6AD5 | 66 | 1 | re/00-master-index.md、re/26-movement-and-triggers.md、re/64-enter-location-prompt.md、re/73-shop-and-doctor-entry.md |
| `0x16C7C` | seg000+0x6C7C | 30 | 1 | re/72-facility-entry-and-command-bar.md |
| `0x1708B` | seg000+0x708B | 256 | 1 | re/00-master-index.md、re/15-character-record.md、re/17-packed-text.md |
| `0x1738A` | seg000+0x738A | 11 | 1 | re/40-combat-screen.md |
| `0x17395` | seg000+0x7395 | 19 | 1 | re/40-combat-screen.md |
| `0x17B80` | seg000+0x7B80 | 15 | 1 | re/17-packed-text.md |
| `0x17D7A` | seg000+0x7D7A | 65 | 1 | re/00-master-index.md、re/46-typed-answers-and-text-input.md、re/68-cell-rewrite.md |
| `0x1841F` | seg000+0x841F | 201 | 1 | re/00-master-index.md、re/05-storage-layer.md、re/06-resource-directory.md、re/07-msq-blocks.md、re/08-msq-encryption.md、re/09-msq-map-structure.md、re/10-huffman-compression.md、re/11-huffman-decoder.md、re/12-msq-tail-and-text-model.md、re/16-msq-block-layout.md、re/18-block-text.md、re/24-map-layers-and-tiles.md、re/61-map-id-table.md |
| `0x186B6` | seg000+0x86B6 | 142 | 1 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/10-huffman-compression.md、re/23-picture-format.md、re/24-map-layers-and-tiles.md |
| `0x18744` | seg000+0x8744 | 189 | 1 | re/00-master-index.md、re/05-storage-layer.md、re/07-msq-blocks.md、re/09-msq-map-structure.md、re/10-huffman-compression.md、re/27-game-clock.md、re/30-save-layout.md |
| `0x18928` | seg000+0x8928 | 136 | 1 | re/43-input-and-hotkeys.md |
| `0x18B4C` | seg000+0x8B4C | 20 | 1 | re/14-fonts-and-text-encoding.md |
| `0x18D8E` | seg000+0x8D8E | 37 | 1 | re/00-master-index.md、re/29-map-event-handlers.md、re/46-typed-answers-and-text-input.md |
| `0x18EAD` | seg000+0x8EAD | 76 | 1 | re/14-fonts-and-text-encoding.md |
| `0x19202` | seg000+0x9202 | 321 | 1 | re/45-item-data-and-weapon-damage.md |
| `0x19C72` | seg000+0x9C72 | 9 | 1 | re/20-combat-resolution.md |
| `0x19F12` | seg000+0x9F12 | 224 | 1 | re/00-master-index.md、re/14-fonts-and-text-encoding.md、re/28-text-variants.md、re/58-line-flush-and-scrollback.md |
| `0x1A308` | seg000+0xA308 | 35 | 1 | re/72-facility-entry-and-command-bar.md |
| `0x1B7FE` | seg000+0xB7FE | 119 | 1 | re/05-storage-layer.md、re/07-msq-blocks.md、re/10-huffman-compression.md、re/23-picture-format.md |
| `0x1B875` | seg000+0xB875 | 34 | 1 | re/25-screen-layout.md |
| `0x1BF5F` | seg000+0xBF5F | 188 | 1 | re/22-shop-and-items.md、re/42-facility-loops.md |
| `0x1C073` | seg000+0xC073 | 205 | 1 | re/22-shop-and-items.md、re/29-map-event-handlers.md、re/42-facility-loops.md |
| `0x1C140` | seg000+0xC140 | 72 | 1 | re/22-shop-and-items.md、re/42-facility-loops.md、re/45-item-data-and-weapon-damage.md |
| `0x1C5B0` | seg000+0xC5B0 | 88 | 1 | re/32-skill-checks-and-xp.md、re/35-status-and-healing.md |
| `0x1C6C9` | seg000+0xC6C9 | 789 | 1 | re/00-master-index.md、re/21-attributes.md、re/45-item-data-and-weapon-damage.md、re/47-dosbox-oracle.md |
| `0x1CB30` | seg000+0xCB30 | 55 | 1 | re/27-game-clock.md、re/60-teleport-and-map-change.md、re/70-nibble1-and-facility-entry.md、re/72-facility-entry-and-command-bar.md、re/74-heat-entry-and-gate-display.md |
| `0x1CB67` | seg001+0x0 | 10 | 1 | re/00-master-index.md、re/02-exepack-unpack.md、re/44-audio.md |
| `0x1CB75` | seg001+0xE | 53 | 1 | re/00-master-index.md、re/00-remake-knowledge-gaps.md、re/44-audio.md |
| `0x1CBAA` | seg001+0x43 | 41 | 1 | re/00-remake-knowledge-gaps.md |
| `0x1CC76` | seg001+0x10F | 83 | 1 | re/00-master-index.md、re/44-audio.md |
| `0x1CCC9` | seg001+0x162 | 137 | 1 | re/00-master-index.md、re/44-audio.md |
| `0x1CD52` | seg001+0x1EB | 194 | 1 | re/00-master-index.md、re/44-audio.md |
| `0x110B6` | seg000+0x10B6 | 648 | 0 | re/00-master-index.md、re/02-exepack-unpack.md、re/03-boot-and-asset-loading.md、re/04-overlay-wla-bin.md |
| `0x16D14` | seg000+0x6D14 | 6 | 0 | re/46-typed-answers-and-text-input.md |
| `0x17748` | seg000+0x7748 | 241 | 0 | re/46-typed-answers-and-text-input.md |

## 尚未碰過的函式（依大小排序，前 60 個）

大的通常是主邏輯，是後續分析的優先對象。

| 位址 | segment:offset | 大小 | callers |
|---|---|---:|---:|
| `0x12D84` | seg000+0x2D84 | 733 | 1 |
| `0x1B170` | seg000+0xB170 | 677 | 1 |
| `0x15500` | seg000+0x5500 | 370 | 3 |
| `0x130C8` | seg000+0x30C8 | 347 | 1 |
| `0x19130` | seg000+0x9130 | 210 | 3 |
| `0x121A8` | seg000+0x21A8 | 186 | 2 |
| `0x10FD3` | seg000+0xFD3 | 158 | 4 |
| `0x18B6B` | seg000+0x8B6B | 148 | 2 |
| `0x135C6` | seg000+0x35C6 | 139 | 1 |
| `0x134B2` | seg000+0x34B2 | 138 | 1 |
| `0x172D4` | seg000+0x72D4 | 131 | 4 |
| `0x19ACD` | seg000+0x9ACD | 130 | 2 |
| `0x14E4A` | seg000+0x4E4A | 129 | 1 |
| `0x15D30` | seg000+0x5D30 | 127 | 1 |
| `0x179B7` | seg000+0x79B7 | 126 | 5 |
| `0x174B7` | seg000+0x74B7 | 124 | 1 |
| `0x1B43C` | seg000+0xB43C | 120 | 4 |
| `0x114BE` | seg000+0x14BE | 118 | 2 |
| `0x18A4F` | seg000+0x8A4F | 117 | 2 |
| `0x19506` | seg000+0x9506 | 115 | 2 |
| `0x19814` | seg000+0x9814 | 114 | 2 |
| `0x1A1EF` | seg000+0xA1EF | 108 | 4 |
| `0x160E1` | seg000+0x60E1 | 104 | 2 |
| `0x12284` | seg000+0x2284 | 101 | 2 |
| `0x122E9` | seg000+0x22E9 | 101 | 2 |
| `0x16018` | seg000+0x6018 | 96 | 3 |
| `0x1095D` | seg000+0x95D | 95 | 1 |
| `0x108FF` | seg000+0x8FF | 94 | 1 |
| `0x138C1` | seg000+0x38C1 | 93 | 9 |
| `0x14ECB` | seg000+0x4ECB | 91 | 3 |
| `0x18AC4` | seg000+0x8AC4 | 87 | 3 |
| `0x18288` | seg000+0x8288 | 84 | 2 |
| `0x13828` | seg000+0x3828 | 80 | 1 |
| `0x14090` | seg000+0x4090 | 77 | 1 |
| `0x1567F` | seg000+0x567F | 76 | 1 |
| `0x126D7` | seg000+0x26D7 | 72 | 2 |
| `0x1BD37` | seg000+0xBD37 | 72 | 1 |
| `0x196DB` | seg000+0x96DB | 69 | 2 |
| `0x14F9D` | seg000+0x4F9D | 65 | 1 |
| `0x16840` | seg000+0x6840 | 64 | 3 |
| `0x12B8A` | seg000+0x2B8A | 63 | 1 |
| `0x1395B` | seg000+0x395B | 62 | 1 |
| `0x19770` | seg000+0x9770 | 62 | 4 |
| `0x12CFE` | seg000+0x2CFE | 58 | 4 |
| `0x17709` | seg000+0x7709 | 58 | 2 |
| `0x1A430` | seg000+0xA430 | 58 | 1 |
| `0x13AAF` | seg000+0x3AAF | 53 | 4 |
| `0x13697` | seg000+0x3697 | 48 | 2 |
| `0x16078` | seg000+0x6078 | 48 | 2 |
| `0x1718B` | seg000+0x718B | 46 | 1 |
| `0x13086` | seg000+0x3086 | 43 | 2 |
| `0x10879` | seg000+0x879 | 42 | 1 |
| `0x157AD` | seg000+0x57AD | 41 | 4 |
| `0x13999` | seg000+0x3999 | 40 | 1 |
| `0x1724F` | seg000+0x724F | 40 | 5 |
| `0x18320` | seg000+0x8320 | 40 | 3 |
| `0x19C04` | seg000+0x9C04 | 40 | 25 |
| `0x1B4B4` | seg000+0xB4B4 | 40 | 7 |
| `0x1BB6C` | seg000+0xBB6C | 40 | 1 |
| `0x11B5C` | seg000+0x1B5C | 39 | 3 |
