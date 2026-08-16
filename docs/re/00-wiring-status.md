# 00：接線狀態 —— RE 解出來的東西，remake 有沒有真的用上

> **這份回答的問題只有一個：某份筆記的結論，程式碼裡有沒有人在用。**
>
> 四份 `00-*` 的分工：
> [`00-master-index.md`](00-master-index.md)＝已知的事實速查；
> [`00-remake-knowledge-gaps.md`](00-remake-knowledge-gaps.md)＝還缺什麼；
> [`00-function-index.md`](00-function-index.md)＝641 個函式誰解過；
> **本表**＝已經解出來的，接上了沒有。
>
> 這份表由 `TestWiringStatus`（`internal/wiring`）守著，**雙向都會紅**：
> 新寫一份筆記卻沒登記會紅；登記成「已接」但程式碼裡沒有人引用會紅；
> 標成「不適用」卻被引用了也會紅。**靠記憶回頭接，已經漏過三次**
> （Agility、對手行動值、敵人目標選擇，見 `docs/re/88`、`89`）。

## 這份表怎麼維護

- 狀態只有三種：**已接**（程式碼或測試裡有 `docs/re/NN` 的引用）、
  **未接**（結論成立但還沒實作，理由欄寫清楚缺什麼）、
  **不適用**（考證、方法或分析基礎，沒有可實作的結論）。
- 引用寫在**用到那個結論的地方**，不是集中在檔頭——
  這樣讀程式碼的人當場看得到依據，機器也數得到。
- ⚠ **「有引用」只證明有人讀過那份筆記，不證明實作是對的。**
  正確性靠各自的門檻測試（覆蓋率、分布、round-trip），這份表只擋「整份忘了接」。

## 還在暫代的位置

RE 沒解出來、程式碼裡先用一個值頂著的地方。**每一處在程式碼裡都寫著「暫代」**，
而這張表與那些註解由 `TestPlaceholders` 雙向對帳：新增一處卻沒登記會紅，
解掉了卻沒清註解也會紅。

**目前一處都沒有。** 最後三處在 2026-08-15 清掉：隊伍行動值解出來了
（`docs/re/90`）、EGA 調色盤本來就已確認（`docs/re/23` §7，只是註解沒跟上），
斷字與字型顏色改列為重製決策（見下）。

## 重製版自訂的地方

原版怎麼做**不查**，由重製版自己決定——這不是債，是定案。
與「暫代」的差別在於：暫代等著 RE 補上，這些不等。

| Go 位置 | 決定了什麼 | 為什麼不追原版 |
|---|---|---|
| `internal/textlayout/textlayout.go` | 自動換行硬斷，不退到前一個空白 | 排版是重製版自己的事；中文為主的譯文本來就逐字斷行。⚠ 要改成英文不切字的話，`Events` 的 Line／Col 要一起搬 |
| `internal/render/render.go` | 單色字型畫成 EGA 15（亮白） | 16 色裡與訊息視窗底色對比最高；原版取哪個顏色沒查也不打算查 |

（使用者定案 2026-08-15。）

## 統計

筆記 **100** 份：已接 **95**、未接 **0**、不適用 **5**。

| # | 筆記 | 狀態 | 接在哪／為什麼 |
|---:|---|---|---|
| 01 | [分析目標的身分與 `wl.exe` 的第一份資料庫](01-binary-identity.md) | 已接 | `internal/assets/rom.go`、`tools/ida/export_forced.py`（共 3 處） |
| 02 | [`wl.exe` 是 EXEPACK 打包的，解包後才是真正的分析對象](02-exepack-unpack.md) | 已接 | `internal/assets/rom.go` |
| 03 | [開機序列、檔名表與資產載入](03-boot-and-asset-loading.md) | 已接 | `tools/apply_overlay.py`、`tools/decode_pic.py`（共 3 處） |
| 04 | [`wla.bin` overlay 與它提供的繪圖層](04-overlay-wla-bin.md) | 已接 | `internal/assets/pic.go` |
| 05 | [雙模式儲存層 —— 磁片絕對磁區 vs 硬碟 DOS 檔案](05-storage-layer.md) | 不適用 | 讀檔層的考證。`sub_11AE8`／`sub_11B83`（8 bytes 段標頭 ＋ Huffman 解壓）已解，但 remake 用自己的 `Decompress`，不重現原版的緩衝管理與磁區讀取 |
| 06 | [`GAME1`／`GAME2` 的資源目錄，與文字輸出層](06-resource-directory.md) | 已接 | `internal/assets/msq.go`、`tools/split_resources.py` |
| 07 | [`GAME1`／`GAME2` 切開了 —— 42 個 MSQ 區塊](07-msq-blocks.md) | 已接 | `internal/assets/msq.go` |
| 08 | [MSQ 區塊的加密解開了 —— 42/42 通過原版自己的驗證](08-msq-encryption.md) | 已接 | `internal/assets/msq.go`、`tools/decrypt_msq.py`（共 3 處） |
| 09 | [MSQ 區塊的內部 —— 地圖層](09-msq-map-structure.md) | 已接 | `internal/assets/msq.go` |
| 10 | [第二層是 Huffman —— 容器格式與各資料檔的載入器](10-huffman-compression.md) | 已接 | `internal/assets/huffman.go`、`tools/huffman.py` |
| 11 | [](11-huffman-decoder.md) | 已接 | `internal/assets/huffman.go` |
| 12 | [MSQ 區塊的尾段與文字的分佈](12-msq-tail-and-text-model.md) | 已接 | `internal/assets/msq.go` |
| 13 | [亂數產生器與擲骰層](13-rng.md) | 已接 | `internal/game/rng/rng.go`、`tools/rng.py` |
| 14 | [字型、字元編碼與文字控制碼](14-fonts-and-text-encoding.md) | 已接 | `internal/assets/font.go`、`internal/input/input.go`（共 6 處） |
| 15 | [角色記錄的定址與欄位](15-character-record.md) | 已接 | `internal/game/character.go`、`internal/game/combat.go`（共 6 處） |
| 16 | [MSQ 區塊的整體佈局與記錄區](16-msq-block-layout.md) | 已接 | `internal/assets/msq.go`、`tools/summarize_msq_layout.py` |
| 17 | [5-bit 打包文字與執行檔的字串表](17-packed-text.md) | 已接 | `internal/assets/text.go`、`internal/game/combat.go`（共 7 處） |
| 18 | [地圖區塊的文字 —— 4,401 條全部解出](18-block-text.md) | 已接 | `internal/assets/msq.go`、`internal/assets/text.go`（共 3 處） |
| 19 | [資料驅動的效果系統與傷害計算](19-effects-and-damage.md) | 已接 | `internal/game/combat.go`、`internal/game/party.go`（共 3 處） |
| 20 | [命中判定與武器傷害](20-combat-resolution.md) | 已接 | `internal/game/combat.go`、`internal/game/commands.go`（共 6 處） |
| 21 | [七個屬性、修正值階梯與檢定骰](21-attributes.md) | 已接 | `internal/game/character.go`、`internal/game/combat.go`（共 8 處） |
| 22 | [商店、價格公式與物品資料表](22-shop-and-items.md) | 已接 | `internal/game/facility.go`、`internal/game/facilityloop.go`（共 3 處） |
| 23 | [圖片格式 —— packed 4bpp ＋ 列間 XOR delta](23-picture-format.md) | 已接 | `internal/assets/pic.go`、`internal/play/encounter.go`（共 10 處） |
| 24 | [地圖的三層結構與 `ALLHTDS` 圖磚](24-map-layers-and-tiles.md) | 已接 | `internal/assets/msq.go`、`internal/assets/pic.go`（共 7 處） |
| 25 | [畫面版面 —— 座標系統與五個視窗](25-screen-layout.md) | 已接 | `cmd/wl-save/main.go`、`internal/game/world.go`（共 7 處） |
| 26 | [移動、捲動與地圖事件觸發](26-movement-and-triggers.md) | 已接 | `internal/game/mapicons.go`、`internal/game/world.go`（共 6 處） |
| 27 | [遊戲時鐘 —— 24 小時制，走一步花多久由地圖決定](27-game-clock.md) | 已接 | `internal/game/clock.go`、`internal/game/rng/rng.go`（共 4 處） |
| 28 | [文字的變形機制 —— 單複數、性別、數量](28-text-variants.md) | 已接 | `internal/textlayout/textlayout.go`、`tools/build_lang.py`（共 6 處） |
| 29 | [地圖事件處理函式 —— 八個 nibble 全解](29-map-event-handlers.md) | 已接 | `internal/game/facility.go`、`internal/game/world.go`（共 6 處） |
| 30 | [存檔的內部結構](30-save-layout.md) | 已接 | `internal/assets/save.go`、`tools/dump_save.py`（共 4 處） |
| 31 | [經驗值、升級與技能學習](31-experience-and-skills.md) | 已接 | `internal/game/levels.go`、`internal/play/command.go`（共 5 處） |
| 32 | [檢定、技能成長與經驗值來源](32-skill-checks-and-xp.md) | 已接 | `internal/assets/rom.go`、`internal/game/checks.go`（共 10 處） |
| 33 | [段落編號在遊戲裡怎麼出現](33-paragraph-references.md) | 已接 | `internal/game/journal.go`、`tools/extract_paragraph_refs.py`（共 6 處） |
| 34 | [地圖腳本的 44 個指令](34-map-script-opcodes.md) | 已接 | `internal/assets/msq.go`、`internal/game/script.go`（共 5 處） |
| 35 | [狀態、疾病與隨時間的恢復](35-status-and-healing.md) | 已接 | `internal/game/facility.go`、`internal/game/party.go`（共 4 處） |
| 36 | [戰鬥的回合與行動順序](36-combat-rounds.md) | 已接 | `internal/game/rounds.go`、`internal/game/rounds_test.go` |
| 37 | [敵方記錄、血量的來源，與距離表](37-enemy-records-and-hp.md) | 已接 | `internal/assets/msq.go`、`internal/game/combat.go`（共 9 處） |
| 38 | [戰鬥的指令階段與逃跑](38-combat-commands-and-flee.md) | 已接 | `internal/game/commands.go`、`internal/game/handlers.go`（共 4 處） |
| 39 | [遭遇怎麼冒出來——視窗掃描與遭遇佇列](39-encounter-scan.md) | 已接 | `internal/game/encounterscan.go`、`internal/game/world.go` |
| 40 | [戰鬥畫面——名單模式與訊息序列](40-combat-screen.md) | 已接 | `internal/play/combat.go`、`internal/play/command.go`（共 6 處） |
| 41 | [四支指令處理程式（Hire／Weapon／Use／Load）](41-command-handlers.md) | 已接 | `internal/game/handlers.go` |
| 42 | [設施的互動迴圈（商店與醫生）](42-facility-loops.md) | 已接 | `internal/game/facility.go`、`internal/game/facilityloop.go`（共 5 處） |
| 43 | [按鍵從哪來——鍵盤、滑鼠熱區，與 `\x10` 的真正用途](43-input-and-hotkeys.md) | 已接 | `internal/assets/save.go`、`internal/input/mouse.go`（共 9 處） |
| 44 | [音效——PC 喇叭、四個聲部與位元組碼直譯器](44-audio.md) | 已接 | `cmd/wasteland/main.go`、`internal/assets/rom.go`（共 14 處） |
| 45 | [物品資料表全解，與隊伍傷害的第一項](45-item-data-and-weapon-damage.md) | 已接 | `internal/assets/save.go`、`internal/game/combat.go`（共 10 處） |
| 46 | [打字回答、文字輸入與字串比對](46-typed-answers-and-text-input.md) | 已接 | `internal/assets/msq.go`、`internal/game/answers.go`（共 12 處） |
| 47 | [DOSBox 參考環境，與第一批實機對拍](47-dosbox-oracle.md) | 已接 | `internal/game/world.go`、`internal/play/play.go`（共 4 處） |
| 48 | [`IC0_9.WLF` 十張疊圖各自是什麼](48-map-icons.md) | 已接 | `internal/assets/rom.go`、`internal/game/mapicons.go`（共 7 處） |
| 49 | [存檔改寫的實機驗收](49-save-roundtrip-on-hardware.md) | 已接 | `internal/assets/assets_test.go` |
| 50 | [物品 70／71／72 為什麼沒有名字](50-unnamed-items.md) | 不適用 | 物品 70／71／72 的身分考證——這份 DOS 版問不出更多，資料原樣 round-trip |
| 51 | [遭遇驅動器——地圖與戰鬥之間那一層](51-encounter-driver.md) | 已接 | `internal/play/encounter.go`、`internal/play/play.go`（共 3 處） |
| 52 | [技能訓練師的流程](52-trainer-facility.md) | 已接 | `internal/play/facility.go`、`internal/play/shop.go`（共 3 處） |
| 53 | [清單框架（`sub_16DB4` ／ `sub_16D34`）](53-list-framework.md) | 已接 | `internal/play/shop.go`、`internal/play/shop_test.go` |
| 54 | [設施畫面的版面](54-facility-screen-layout.md) | 已接 | `internal/play/play.go`、`internal/render/render.go`（共 5 處） |
| 55 | [輻射結算與「無視護甲」旗標（`ds:46EFh`）](55-radiation-and-armour-bypass.md) | 已接 | `internal/game/radiation.go`、`tools/summarize_radiation.py`（共 3 處） |
| 56 | [`TRANSTBL` 是 50 組 16 色對照表，而且沒有人讀它](56-transtbl.md) | 已接 | `tools/summarize_transtbl.py` |
| 57 | [`CURS` 是 8 個滑鼠游標（遮罩 ＋ 圖形並排）](57-curs.md) | 已接 | `tools/dump_curs.py` |
| 58 | [控制碼 `0x08` ＝ 沖出一行不捲動；順帶解出 scrollback](58-line-flush-and-scrollback.md) | 已接 | `internal/game/world.go`、`internal/textlayout/textlayout.go`（共 3 處） |
| 59 | [正常玩家路徑對原版驗收（第一輪）](59-playtest-against-original.md) | 不適用 | 實機對拍的驗收紀錄與方法，不產生程式碼 |
| 60 | [傳送與換地圖（nibble 10）](60-teleport-and-map-change.md) | 已接 | `internal/assets/save.go`、`internal/game/teleport.go`（共 5 處） |
| 61 | [地圖編號表 —— 建築內部怎麼進去](61-map-id-table.md) | 已接 | `internal/assets/mapid.go`、`internal/play/play.go`（共 3 處） |
| 62 | [第四道閘 —— nibble 11 是山與牆](62-fourth-gate-terrain-blocking.md) | 已接 | `internal/game/world.go`、`internal/play/play.go`（共 5 處） |
| 63 | [資源編號與切片索引是兩件事](63-resource-id-vs-index.md) | 已接 | `internal/play/play.go`、`internal/assets/assets_test.go` |
| 64 | [進新地點的確認（第三道閘的一半）](64-enter-location-prompt.md) | 已接 | `internal/game/world.go`、`internal/play/play.go`（共 3 處） |
| 65 | [第三道閘的另一半 —— nibble 2 是條件式的](65-third-gate-conditions.md) | 已接 | `internal/game/world.go`、`internal/game/world_events_test.go` |
| 66 | [nibble 2 的事件處理，與還沒解的沙漠高溫](66-nibble2-event-and-heat.md) | 已接 | `internal/game/world.go`、`internal/play/play.go` |
| 67 | [條件閘的獎懲參數，與水壺](67-gate-penalty-and-canteen.md) | 已接 | `internal/game/gates.go`、`internal/game/world.go`（共 4 處） |
| 68 | [改寫地圖格（`sub_17CFF`）](68-cell-rewrite.md) | 已接 | `internal/game/gates.go`、`internal/game/world_events_test.go` |
| 69 | [條件閘的四個旗標（記錄 `+0x00` 的低位）](69-gate-flags.md) | 已接 | `internal/game/answers.go`、`internal/game/gates.go`（共 5 處） |
| 70 | [nibble 1 的氛圍敘述，與商店入口的排除紀錄](70-nibble1-and-facility-entry.md) | 已接 | `internal/game/world.go`、`internal/play/nibble1_test.go` |
| 71 | [nibble 12 是遠端批次改寫器](71-nibble12-batch-patch.md) | 已接 | `internal/game/world.go`、`internal/play/nibble12_test.go` |
| 72 | [進地點的完整路徑，與地圖指令列](72-facility-entry-and-command-bar.md) | 已接 | `internal/play/command.go`、`internal/render/render.go`（共 3 處） |
| 73 | [商店與醫生的入口 —— 傳送記錄的 `+0x04`／`+0x05`](73-shop-and-doctor-entry.md) | 已接 | `internal/game/world.go`、`internal/play/play.go`（共 4 處） |
| 74 | [高溫記錄的入口（再排除三條）與 `sub_142ED` 的顯示層](74-heat-entry-and-gate-display.md) | 已接 | `internal/play/heat_test.go` |
| 75 | [沙漠高溫的入口 —— 腳本 opcode 3 的晝夜分支](75-desert-heat-entry.md) | 已接 | `internal/game/script.go`、`internal/game/world.go`（共 3 處） |
| 76 | [腳本 opcode 的覆蓋率盤點](76-script-opcode-coverage.md) | 已接 | `internal/play/combatloop_test.go` |
| 77 | [遭遇覆蓋率盤點 —— 敵人格是生出來的，而 remake 沒有生成器](77-encounter-spawn-gap.md) | 已接 | `internal/play/combatcoverage_test.go`、`internal/play/combatloop_test.go`（共 3 處） |
| 78 | [遭遇生成器 `sub_16890`](78-encounter-spawn.md) | 已接 | `internal/assets/rom.go`、`internal/game/encounter.go`（共 9 處） |
| 79 | [設施覆蓋率盤點 —— 跳表索引 ≥ 5 就是 opcode](79-facility-coverage.md) | 已接 | `internal/game/script.go`、`internal/game/world.go`（共 6 處） |
| 80 | [訓練師的技能清單](80-trainer-skill-list.md) | 已接 | `internal/play/facilitycoverage_test.go`、`internal/play/trainer_test.go` |
| 81 | [戰鬥迴圈的端到端門檻，與一個槽號當 ID 的 bug](81-combat-loop-coverage.md) | 已接 | `internal/play/combatloop_test.go` |
| 82 | [存檔的三道門檻](82-save-round-trip.md) | 已接 | `internal/play/save_test.go` |
| 83 | [中文化的覆蓋率門檻](83-translation-coverage.md) | 已接 | `internal/play/lang_coverage_test.go` |
| 84 | [呈現層的三道門檻](84-render-coverage.md) | 已接 | `internal/play/render_coverage_test.go` |
| 85 | [地圖上的敵人圖示](85-enemy-map-icon.md) | 已接 | `internal/play/combat.go`、`internal/play/round.go`（共 3 處） |
| 86 | [戰鬥訊息的主詞與受詞](86-combat-messages.md) | 已接 | `internal/play/combatloop_test.go` |
| 87 | [`sub_15036` 是敵人在地圖上移動，不是目標選擇](87-enemy-map-movement.md) | 已接 | `internal/play/round.go` |
| 88 | [命中累加值 `sub_1B108` 的四個項全部落地](88-hit-accumulator.md) | 已接 | `internal/game/combat.go`、`internal/play/round.go`（共 3 處） |
| 89 | [敵人打誰是隨機重抽，以及「倒下」與「死亡」是兩個判準](89-enemy-target-and-down.md) | 已接 | `internal/game/party.go`、`internal/play/command.go`（共 5 處） |
| 90 | [隊伍的行動值，以及誰會被排進行動表](90-party-initiative.md) | 已接 | `internal/game/rounds.go`、`internal/play/round.go`（共 4 處） |
| 91 | [地圖指令列的七個處理程式 —— 升級的入口是 RADIO](91-map-command-bar.md) | 已接 | `internal/play/command.go`、`internal/play/play.go`（共 4 處） |
| 92 | [`USE` 指令 —— Skill／Item／Attribute 三選一，與施用的骨架](92-use-command.md) | 已接 | `internal/game/gates.go`、`internal/play/play.go`（共 5 處） |
| 93 | [`ORDER`／`DISBAND`／`VIEW` —— 兩支要多隊伍，一支不用](93-order-disband-view.md) | 已接 | `internal/play/command.go`、`internal/play/command_test.go` |
| 94 | [`ENC` —— 它不是新指令，是自動遭遇的手動入口](94-enc-command.md) | 已接 | `internal/play/enc.go`、`internal/play/command.go` |
| 95 | [主選單只有一個選項，而且沒有「讀檔」](95-main-menu.md) | 已接 | `internal/play/mainmenu.go`、`cmd/wasteland/main.go` |
| 96 | [結局 —— 它掛在設施跳表的第 4 格](96-ending.md) | 已接 | `internal/assets/endanim.go`、`internal/game/character.go`、`internal/play/ending.go`（共 7 處） |
| 97 | [抽樣試玩（第二輪）—— 七段流程各走一遍](97-playtest-sampling.md) | 不適用 | 驗收紀錄與方法，不產生程式碼（修掉的缺口各自引用它們的依據筆記） |
| 98 | [補完 A0 —— 中文接線、庫存持久化，與結局觸發點的定位](98-a0-wiring.md) | 不適用 | 同上：接線紀錄與一份未解的定位，結論各自引用 `docs/re/40`／`42`／`45`／`96` |
| 99 | [全隊陣亡怎麼處理 —— Grim Reaper 那一格](99-party-wipe.md) | 已接 | `game.Party.Wipe` 三分支 ＋ `internal/play/wipe.go`：全倒自動 `View`、救不回來走死亡畫面（圖 0x3B、地點名與訊息從映像直讀）。規格 28 |
| 100 | [結局的觸發點 —— 不在資料裡，在主迴圈裡](100-ending-trigger.md) | 已接 | `internal/game/selfdestruct.go`、`internal/game/script.go`、`internal/game/gates.go`、`internal/play/question.go`、`internal/play/play.go` |
