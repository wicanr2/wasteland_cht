# 攻略中的機制斷言與逆向對照表

這篇攻略是 1990 年前後一位玩家的實際遊玩紀錄，裡面夾雜著大量對遊戲規則的敘述——
有些是操作說明（「按 E 鍵戰鬥」「站在他下面」），有些是條件判斷
（「最好有 2 級的拆除炸彈技能」），有些是作者自己的推論（「這支血杖大概是假的」）。

**這些斷言全部都是待驗證的線索，不是規格的依據。** `CLAUDE.md` §5 的 oracle 優先序
把社群攻略排在最後一位，所以對照的方向是固定的：

> 拿 `docs/re/` 的逆向結果去**驗證**這些斷言，不是拿這些斷言去修正逆向結果。

驗證的結果有三種，記在「狀態」欄：

| 狀態 | 意思 |
|---|---|
| 待查 | 還沒去對 |
| 相符 | 逆向結果支持這條敘述 |
| 有出入 | 逆向結果和這條敘述不一致，差異寫在備註 |

## 對照表

| 編號 | 出處 | 斷言 | 該查哪裡 | 狀態 |
|---|---|---|---|---|
| M-01 | [p.56](p56.md)、[p.57](p57.md) | 拆除炸彈技能**至少 2 級**才能拆 Felicia 身上的炸彈、才能安全拿走一碰就炸的 TNT | [`re/32`](../../re/32-skill-checks-and-xp.md) 技能檢定、[`re/69`](../../re/69-gate-flags.md) 閘旗標、[`re/80`](../../re/80-trainer-skill-list.md) 技能表 | 待查 |
| M-02 | [p.56](p56.md)、[p.57](p57.md) | 隊伍有名額上限；要讓新人入隊得先叫舊隊員「走路」／「回去休養」 | [`re/108`](../../re/108-combat-use-and-hire.md)、[`re/110`](../../re/110-hire-resolution.md)、[`re/103`](../../re/103-roster-line-columns.md) | 待查 |
| M-03 | [p.57](p57.md) | 友善 NPC（Christina）**在戰鬥中被誤傷會立刻轉為敵對**；戰鬥中沒碰到她就能直接招募 | [`re/110`](../../re/110-hire-resolution.md)、[`re/20`](../../re/20-combat-resolution.md) | 待查 |
| M-04 | [p.57](p57.md) | 可以把 NPC 隊員身上的裝備（UZI）取走給玩家角色用 | [`re/108`](../../re/108-combat-use-and-hire.md)、[`re/22`](../../re/22-shop-and-items.md) | 待查 |
| M-05 | [p.57](p57.md) | 和 NPC 交易要**站在特定相對位置**（「須站在他下面」）；給酒鬼酒要**重複兩次**才觸發情報 | [`re/26`](../../re/26-movement-and-triggers.md)、[`re/34`](../../re/34-map-script-opcodes.md) | 待查 |
| M-06 | [p.57](p57.md) | 「血杖驗證桌」：把血杖**使用到桌上**，假的會消失、真的會留下 | [`re/34`](../../re/34-map-script-opcodes.md)、[`re/29`](../../re/29-map-event-handlers.md) | 待查 |
| M-07 | [p.58](p58.md) | 在沙堆上**重複使用攀爬技能可以練等** | [`re/31`](../../re/31-experience-and-skills.md)、[`re/32`](../../re/32-skill-checks-and-xp.md) | 待查 |
| M-08 | [p.58](p58.md) | 血神殿棋盤檢查站：走正確路徑到機械人左方，**回答步數 30**；踏出路徑第一次砲塔昇起、之後持續射擊；步數答錯要重走 | [`re/46`](../../re/46-typed-answers-and-text-input.md)、[`re/34`](../../re/34-map-script-opcodes.md)、[`generated/passwords.md`](../generated/passwords.md) | 待查 |
| M-09 | [p.58](p58.md) | 蓋氏計數器在接近輻射區時會發出提示 | [`re/55`](../../re/55-radiation-and-armour-bypass.md) | 待查 |
| M-10 | [p.58](p58.md) | 廢坑中可取得兩支 **M1989 突擊步槍** | [`re/45`](../../re/45-item-data-and-weapon-damage.md)、[`re/50`](../../re/50-unnamed-items.md) | 待查 |
| M-11 | [p.58](p58.md) | 電網開關是有 **ON／OFF** 兩個選項的互動物件 | [`re/34`](../../re/34-map-script-opcodes.md) | 待查 |
| M-12 | [p.59](p59.md) | 過血池的方法是「**找一個隊員不斷的往上使用游泳技能**」 | [`re/62`](../../re/62-fourth-gate-terrain-blocking.md)、[`re/65`](../../re/65-third-gate-conditions.md) | 待查 |
| M-13 | [p.59](p59.md) | 血池中食人魚造成持續傷害，**防護衣物夠強就免疫** | [`re/55`](../../re/55-radiation-and-armour-bypass.md)、[`re/67`](../../re/67-gate-penalty-and-canteen.md) | 待查 |
| M-14 | [p.59](p59.md) | **敵人的像沒出現就無法攻擊**；解法是用 `RUN` 指令往前移動直到它出現 | [`re/38`](../../re/38-combat-commands-and-flee.md)、[`re/107`](../../re/107-command-resolution.md)、[`re/40`](../../re/40-combat-screen.md) | 待查 |
| M-15 | [p.59](p59.md) | 地板上的大圓盤是陷阱，會跳出提示問要不要跳過去（答 `Yes`） | [`re/34`](../../re/34-map-script-opcodes.md)、[`re/64`](../../re/64-enter-location-prompt.md) | 待查 |
| M-16 | [p.59](p59.md) | Vegas 的殺人機器**主動攻擊**且裝甲很高（「鐵皮蛋的皮又厚」） | [`re/37`](../../re/37-enemy-records-and-hp.md)、[`re/101`](../../re/101-enemy-move-plan-table.md) | 待查 |
| M-17 | [p.59](p59.md) | 付錢給乞丐可換情報（兩大頭頭的爭霸、Fat Freddy 辦公室密碼） | [`re/42`](../../re/42-facility-loops.md)、[`re/34`](../../re/34-map-script-opcodes.md) | 待查 |
| M-18 | [p.60](p60.md) | **蠍式（`Scorpition`）免疫子彈類武器**，只有 LAW 火箭之類打得動 | [`re/55`](../../re/55-radiation-and-armour-bypass.md) 護甲穿透、[`re/45`](../../re/45-item-data-and-weapon-damage.md)、[`re/19`](../../re/19-effects-and-damage.md) | 待查 |
| M-19 | [p.60](p60.md) | 醫院療傷**要花錢**，而且要經過一段遊戲內時間 | [`re/35`](../../re/35-status-and-healing.md)、[`re/73`](../../re/73-shop-and-doctor-entry.md)、[`re/27`](../../re/27-game-clock.md) | 待查 |
| M-20 | [p.60](p60.md)、[p.61](p61.md) | 音波鑰匙（`Sonic key`）**不只蠍式會掉**，打其他殺人機器也能取得；它能開炸不開的門 | [`re/50`](../../re/50-unnamed-items.md)、[`re/69`](../../re/69-gate-flags.md)、[`generated/gates.md`](../generated/gates.md) | 待查 |
| M-21 | [p.60](p60.md) | 對 Fat Freddy 的要求回答 `NO` 會**直接觸發戰鬥** | [`re/46`](../../re/46-typed-answers-and-text-input.md)、[`re/34`](../../re/34-map-script-opcodes.md) | 待查 |
| M-22 | [p.60](p60.md) | 招募囚犯 Covenant 要先**開門**（選 `Open`）再**解掉他身上的鎖** | [`re/110`](../../re/110-hire-resolution.md)、[`re/69`](../../re/69-gate-flags.md) | 待查 |
| M-23 | [p.60](p60.md) | 蕈狀雲神殿門口的第三題要**打字輸入 `Bloodstaff`** | [`re/46`](../../re/46-typed-answers-and-text-input.md)、[`generated/passwords.md`](../generated/passwords.md) | 待查 |
| M-24 | [p.60](p60.md) | 神殿圖書館可學**能量武器（`Energy`）技能**，對應遊戲中確實存在能量武器 | [`re/80`](../../re/80-trainer-skill-list.md)、[`re/45`](../../re/45-item-data-and-weapon-damage.md)、[`re/52`](../../re/52-trainer-facility.md) | 待查 |
| M-25 | [p.60](p60.md) | **有些門鎖開鎖技能開不了**，只能用 TNT 炸 | [`re/69`](../../re/69-gate-flags.md)、[`re/65`](../../re/65-third-gate-conditions.md) | 待查 |
| M-26 | [p.60](p60.md) | 放射能套裝（`Rad Suit`）的**防護力數值是 5**，可防放射能 | [`re/55`](../../re/55-radiation-and-armour-bypass.md)、[`re/45`](../../re/45-item-data-and-weapon-damage.md) | 待查 |
| M-27 | [p.60](p60.md) | 向 Needles 的牧師說出 `DIPSTICK` 才能取回血杖 | [`re/46`](../../re/46-typed-answers-and-text-input.md)、[`generated/passwords.md`](../generated/passwords.md) | 待查 |
| M-28 | [p.61](p61.md) | 繩子要**站在岸邊、向下使用**才能搭橋過水道 | [`re/62`](../../re/62-fourth-gate-terrain-blocking.md)、[`re/65`](../../re/65-third-gate-conditions.md) | 待查 |
| M-29 | [p.61](p61.md) | 標有 `Use Shovel` 的碎牆要用**鏟子**挖開；挖開後地圖格會改變 | [`re/65`](../../re/65-third-gate-conditions.md)、[`re/68`](../../re/68-cell-rewrite.md) | 待查 |
| M-30 | [p.61](p61.md) | 下水道中 `Cyborg` 和 `Hexborg` 兩種敵人特別危險 | [`re/37`](../../re/37-enemy-records-and-hp.md)、[`re/78`](../../re/78-encounter-spawn.md) | 待查 |
| M-31 | [p.61](p61.md) | Max 的組裝要在四張組合桌上依序「選 (2) → 原地使用零件」，最後各選 (3) | [`re/34`](../../re/34-map-script-opcodes.md)、[`re/107`](../../re/107-command-resolution.md)、[`spec/25`](../../spec/25-facility-menus.md) | 待查 |

## 特別值得先查的三條

這幾條如果和逆向結果對得上，可以直接補進 remake 的驗收清單；對不上則是攻略記錯，
要在對照表註明。

**M-18（蠍式免疫子彈）** — 這是整篇攻略裡對戰鬥系統最強的一條斷言，
而且 [`re/55`](../../re/55-radiation-and-armour-bypass.md) 的主題正好就是護甲與穿透。
如果原版真的有「某類武器無視／無法穿透某個護甲值」的規則，這條會是它的玩家側證據。

**M-03（誤傷 NPC 會翻臉）** — 招募流程的分支條件。
[`re/110`](../../re/110-hire-resolution.md) 已經解過招募，可以直接對這條。

**M-14（敵人未顯示就無法攻擊）** — 這條描述的是戰鬥介面的可見性規則，
會直接影響 remake 的戰鬥畫面該不該畫出還沒「出現」的敵人。

## 順帶記下的原文誤植

這些不是機制，是轉錄時發現的拼寫／排版問題，集中列在這裡方便查：

| 位置 | 原文 | 應為 | 處理 |
|---|---|---|---|
| [p.56](p56.md) 節標題 | `Oneedles` | `Needles` | 照原樣保留 |
| [p.59](p59.md) | `Atohison` | `Atchison` | 照原樣保留 |
| [p.60](p60.md) | `Scorpition` | `Scorpitron` | 照原樣保留 |
| [p.61](p61.md) | `Sever motor` | `Servo motor` | 照原樣保留 |
| [p.61](p61.md) | `Power Conventer` | `Power Converter` | 照原樣保留 |
| [p.58](p58.md) 等處 | 幅射 | 輻射 | 照原樣保留 |
| [p.61](p61.md) | 徙水道 | 下水道 | 照原樣保留 |
