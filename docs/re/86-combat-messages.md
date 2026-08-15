# 86：戰鬥訊息的主詞與受詞

日期：2026-08-15 ｜ 接 `docs/re/81`（戰鬥門檻）、`docs/re/40`（戰鬥畫面）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/81` §3 列的三處暫代之一。

---

## 1. 敵人名稱在執行檔字串表 1 的 `0x52 + Kind`

`EnemyKind.MessageID()` ＝ `0x52 + Kind`（原版 `0x129FF` 的 `add al, 52h`）。
查出來的內容帶單複數控制碼（`0x0A`，`docs/re/17` §4.1）：

| 種類 | 字串 | 內容 |
|---:|---:|---|
| 1 | 83 | `Animal\n\ns\n` |
| 2 | 84 | `Mutant\n\ns\n` |
| 3 | 85 | `Humanoid\n\ns\n` |
| 4 | 86 | `Cyborg\n\ns\n` |
| 5 | 87 | `Robot\n\ns\n` |

格式是「單數 `\n\n` 複數後綴 `\n`」——`Animal` / `Animals`。
`Scene.enemyNames()` 取第一個 `0x0A` 之前那一段當單數名稱。

## 2. 修正前後

| | 修正前 | 修正後 |
|---|---|---|
| 敵人打中 | `hits Hell Razor for 13` | `Animal hits Hell Razor for 13` |
| 敵人落空 | `misses.` | `Animal misses.` |
| 隊伍打中 | `Angela Deth hits for 8` | `Angela Deth hits Animal for 8` |
| 敵人死亡 | `died!` | `Animal died!` |

四句裡有三句缺主詞或受詞。玩家看到的是一串不知道誰打誰的句子，
而**沒有任何測試會紅**——訊息是字串，長度對、格式對、內容錯。

查不到名稱時回 `It`，**不留空白**：留空會變成 `" misses."`，
在訊息視窗裡看起來像排版壞掉而不是資料缺失。

## 3. 門檻

`TestCombatMessagesHaveSubjects`：打一場，檢查每一行訊息**不以動詞或空白開頭**。

```go
for _, bad := range []string{" ", "misses", "hits", "died"} {
    if strings.HasPrefix(l, bad) { t.Errorf("訊息缺主詞：%q", l) }
}
```

## 4. 可重跑的完整指令

```bash
tools/go.sh test ./internal/play/ -run TestCombatMessagesHaveSubjects -v
```

## 5. 這一輪學到的（寫成規則）

- **訊息的缺陷不會讓程式出錯。** 前六把尺量的是「跑不跑得動」「畫不畫得出來」，
  這一項兩邊都正常——缺的是**句子的意思**。
  **文字輸出要有形狀門檻**（開頭是不是名字、有沒有受詞），不然只能靠讀。
- **一個缺口通常有對稱的另一半。** `docs/re/81` 只記了「敵人 miss 訊息沒名字」，
  補的時候才發現隊伍打敵人也缺受詞、死亡訊息也缺主詞。
  **修一處時把同一族的訊息一起看過。**
