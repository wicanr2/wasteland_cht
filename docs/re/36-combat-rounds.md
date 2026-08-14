# 36：戰鬥的回合與行動順序

日期：2026-08-15 ｜ 對應盤點 **D2**（戰鬥流程）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

---

## 1. 敵方是「三組 × 每組最多 10 個」

`0x1ADC6`–`0x1AE50` 是兩層迴圈：

```
外層  ds:CF80h 從 0 數到 2          ← **三組敵人**
內層  bl 從 0 每次 +2 數到 0x14     ← 每組 10 個（HP 陣列一個 2 bytes）
```

`ds:46C8h` 起是每組的敵人資料（`docs/re/20` §5.1）：

| 位移 | 內容 |
|---|---|
| `+0x00`–`+0x13` | 10 個敵人的 16-bit HP |
| `+0x14`–`+0x1D` | **每個敵人一個 byte 的「這回合還沒行動」旗標** |

內層先測 HP 是不是 0（死了就跳過），再測 `+0x14 + n` 那個 byte；
非 0 才排進行動順序，**而且排完立刻清成 0**（`0x1AE06`）——
所以那個 byte 就是「這一輪還沒動過」。

## 2. 行動順序表：`ds:7931h`

```
0x1AE0A  sub_19BF8          ; 累加器 ← 0
0x1AE0D  sub_19C84          ; 2d6 逢同點續擲
0x1AE10  sub_19C2C          ; 累加
0x1AE13  sub_1B15F          ; → sub_12A40（攻擊資料定址）→ sub_12ABA 取欄位
0x1AE16  shl al,1 ×3        ; **× 8**
0x1AE1E  sub_19C2C          ; 累加
0x1AE21  sub_1B0E2          ; 寫進 ds:7931h ＋ 游標（16-bit）
0x1AE26  sub_1B0DA(2)       ; 游標 += 2
```

所以：

```
行動值 = 2d6（逢同點續擲）+ 攻擊資料的某個欄位 × 8
```

游標是 `ds:CF83h`；跳過整組時一次前進 `0x14`（`0x1AE38`），
所以**表的版面是「三組 × 10 格 × 2 bytes」的固定格子**，不是壓縮過的清單。
`sub_1B0F9` 用另一個游標 `ds:CF82h` 把某一格清成 0（那一個單位出局）。

隊伍成員也走同一支 `sub_1B15F`（`0x1B046`、`0x1B123` 兩個呼叫端），
所以**雙方排在同一張表裡**。

推論等級：**強證據**（迴圈邊界、旗標的清除、寫入位址都讀到；
`sub_12ABA` 取的是攻擊資料的哪個欄位還沒對上）。

## 3. 還沒解的

- `sub_12ABA` 取的欄位（行動值的 ×8 那一項）是攻擊資料的哪個位移。
- 行動值算出來之後怎麼排序、誰先動。
- 逃跑與隊形。
- 隊伍那一側（`0x1B046`／`0x1B123`）的完整流程。

## 4. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py workplace/analysis/dumps/round.json \
  0x1ADC6 0x1ADD0 0x1AE00

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/round2.json \
  0x1B0DA 0x1B0E2 0x1B15F 0x1B0F9 0x12ABA --callers
```

## 5. 這一輪學到的（寫成規則）

- **「排進表就把旗標清掉」是回合制的指紋。** 看到「讀一個 byte、非 0 才做事、
  做完寫 0」，而且外面包著兩層固定邊界的迴圈，那就是回合初始化——
  比追資料流快得多。
- **固定格子的表比壓縮清單好認**：游標一次跳 `0x14` 而不是逐格前進，
  說明表的版面是預留好的，索引可以直接對回單位編號。
