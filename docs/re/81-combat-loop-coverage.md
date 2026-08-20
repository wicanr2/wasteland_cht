# 81：戰鬥迴圈的端到端門檻，與一個槽號當 ID 的 bug

日期：2026-08-15 ｜ 接 `docs/re/78`（遭遇生成）、`docs/re/79`（設施覆蓋率）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

腳本、遭遇、設施都有覆蓋率門檻之後，第四把尺量戰鬥迴圈本身。

---

## 1. 端到端門檻

`TestCombatRunsToCompletion`：從地圖走到遭遇 → 全隊下攻擊令 →
把回合跑到結束 → `FinishEncounter` 回地圖。整條路要跑得完。

```
第 73 步打起來，敵人 3 隻
第 2 回合結束，打贏了
收尾：全滅=false 經驗 map[Hell Razor:14 Snake Vargas:14 Thrasher:14]
```

`TestCombatManyBattles` 連打 12 場，統計勝負、回合數分布與「200 回合還沒完」。
一場打得完只證明路是通的；**偶發的卡死只有多打幾場才看得到**。

## 2. 12 場全部第 1 回合結束 → 查出傷害爆表

第一次跑出來的分布是 `map[1:12]`——12 場全部一回合打完，太整齊。
逐場印出來：

```
敵人 HP 8／2／5
misses. ×3
Snake Vargas hits for 15
Hell Razor hits for 112     ← 這個
Thrasher hits for 18
```

**112 點**。出廠存檔 Hell Razor 裝的是物品 16，`Dice ＝ 3`——
3 顆 d6 加上技能與屬性修正到不了 112。

原因在 `weaponOf`：

```go
w, ok := s.Items.Get(c.EquipIndex)   // ← 錯
```

`Character.EquipIndex` 是 `Equip(slot)` 存進去的**背包槽號**（0–29），
而 `ItemTable.Get(id)` 要的是**物品 ID**。拿槽號當 ID 查表會取到
完全不相干的物品——剛好取到一把骰數很大的重武器。

修正：

```go
if int(c.EquipIndex) >= len(c.Items) {
    return game.ItemData{}
}
w, ok := s.Items.Get(c.Items[c.EquipIndex].ID)
```

修完之後回合數分布變成 `map[2:12]`。

⚠ **這個 bug 不會讓任何測試變紅**：戰鬥照打、有人受傷、有人死、經驗照結算。
它只是讓玩家的傷害大約高十倍。抓到它的不是斷言，是**「12 場全部一回合」
這個分布本身看起來不對**。

## 3. 還沒接的

- ~~敵人的 miss 訊息沒有名字~~ → 已補（`docs/re/86`）。
- ~~命中累加值的射程與距離懲罰~~ → 那不是距離，是**對手的行動值**；
  四個項全部接上（`docs/re/88`）。
- 敵人的目標選擇解在 `docs/re/89` §1：`roll(1..隊伍人數)` 隨機挑，
  挑到 CON ≤ 0 的就整個重抽。（`sub_15036` 不是目標表，
  那是敵人在地圖上移動，見 `docs/re/87`。）

## 4. 可重跑的完整指令

```bash
tools/go.sh test ./internal/play/ -run TestCombatRunsToCompletion -v
tools/go.sh test ./internal/play/ -run TestCombatManyBattles -v
```

## 5. 這一輪學到的（寫成規則）

- **分布比斷言先看到問題。** 「12 場全部第 1 回合」沒有違反任何斷言，
  但它就是不對。**統計量測要把分布印出來，不要只印通過與否。**
- **索引與識別碼混用是最安靜的一類 bug。** 槽號與物品 ID 都是小整數、
  查表都會成功、回傳的都是合法物品——**型別一樣、語意不同的兩個值，
  在 Go 裡沒有任何東西會擋**。這種地方要嘛換成具名型別，要嘛在註解寫死。
- **量測的價值不只在找缺口。** 前三把尺量出的是「沒做的東西」，
  這一把量出的是「做了但做錯的東西」——後者更難靠讀程式碼發現。
