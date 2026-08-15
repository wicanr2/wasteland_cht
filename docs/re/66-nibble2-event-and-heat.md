# 66：nibble 2 的事件處理，與還沒解的沙漠高溫

日期：2026-08-15 ｜ 接 `docs/re/65`（第三道閘）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
實機截圖見 `docs/re/47` 的環境。

---

## 1. 閘與事件是同一支

`docs/re/26` §5 的事件表把 nibble 2 指到 `0x13EC9`，而 `docs/re/65` 讀的
第三道閘 `sub_13E9B` 也呼叫 `sub_13EC9`。**兩張表指向同一支函式**：

| 時機 | 誰呼叫 | 結果 |
|---|---|---|
| 走進去之前 | `sub_13E9B`（閘） | CF 決定擋不擋 |
| 走進去之後 | 事件分派表 `ds:AA87h[2]` | 印訊息、跑條件、收尾 |

所以**走得過去的 nibble 2 格，踩上去還是有事發生**——
`sub_13EC9` 開頭的 `mov bl, 1; call sub_16D1A` 印的是記錄 **`+0x01`** 的訊息。

remake 原本對 `EventGate` 只回一句固定的 `BLOCKED BY SOMETHING.`，
現在改成印記錄 `+0x01`（走得過去時），與原版一致。

## 2. 沙漠高溫：線索齊了但入口沒解

實機在世界地圖走一段之後：

```
It's getting warmer.
It is very warm.
It is very warm.
Hell Razor gets hurt for 1 point of damage.   （四個人各一次）
```

已經對上的部分：

| 元素 | 位置 |
|---|---|
| 三階段訊息 | 資源 0 的字串 **20 / 19 / 22**（`getting warmer` → `very warm` → `VERY hot!`） |
| 訊息的持有者 | 資源 0 的 **section 2 記錄 7–12**，`+0x00 ＝ 0xE1`、`+0x01 ＝ 19／20／22` |
| `gets hurt for` | 執行檔字串表 1 第 99 條，全檔只有 `0x158B2` 印它（`sub_157D6`，傷害結算） |
| 扣血的路 | `ds:4718h` 非 0 才印那句——那個旗標由 `sub_141FA` 的 CON 分支設（`docs/re/55` §2） |

**缺的是入口**：資源 0 裡**沒有任何格子指到 section 2 的記錄 7–12**
（全圖掃過，指到高溫字串的只有一格 nibble 10）。
所以那六筆記錄是被別的東西選出來的——三筆一組、訊息由弱到強，
形狀像**隨時間遞增的階段**，而不是踩到某一格。

候選（都還沒讀）：

- 腳本 opcode 35（`0x1AB0E`，啟動一個計時）與 opcode 14（`0x1A7E8`，倒數顯示）
- `sub_13C58` 與 `sub_14296`：`sub_14193`（改角色欄位）的另外兩個呼叫端
- 記錄 `+0x08`／`+0x09`（`1d 82` 之類）與 `+0x0A` 起的條件串列

## 3. 對 remake 的意思

**長途旅行目前不會被高溫消耗**。這不影響「走得到哪裡」，
但會讓沙漠比原版好活——水壺（物品 `0x2F`）與相關的腳本 opcode 7
（`0x1A699`：沒帶的人 CON 設成 −5）目前也沒有觸發點。

## 4. 可重跑的完整指令

```bash
# 三階段訊息與它們的持有者
tools/go.sh test ./internal/play/ -run TestProbeWhoUsesWarmStrings -v   # 一次性探針，見 git log

# 誰印 " gets hurt for "
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_range_refs.py \
  workplace/analysis/dumps/hurt.json 0x63 0x64
```

## 5. 這一輪學到的（寫成規則）

- **同一支函式出現在兩張表裡，就是兩個機制。** `sub_13EC9` 既是移動閘
  也是踩上去的事件處理；只讀其中一個呼叫端會漏掉另一半的行為。
- **「訊息在哪」比「機制在哪」好找。** 高溫的三階段字串一掃就到，
  而且立刻指出它們住在 section 2 的哪幾筆記錄——
  **先找字串，再找誰持有它，最後才追誰觸發**。
- **沒有格子指到的記錄不代表沒用到。** 高溫那六筆就是這樣；
  這與 `TRANSTBL`（真的沒人讀，`docs/re/56`）不同，
  差別在於**實機看得到它的效果**——有效果就一定有入口，只是還沒找到。
