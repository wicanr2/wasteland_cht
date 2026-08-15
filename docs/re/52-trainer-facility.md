# 52：技能訓練師的流程

日期：2026-08-15 ｜ 接 `docs/re/42` §7（「技能訓練師的流程」那一項）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/31` §3 解了**學一個技能要花多少點**，`docs/re/42` 解了商店與醫生的
互動迴圈，訓練師那一支（`0x1BBA0`）一直掛著沒讀。這一份補上。

---

## 1. 進場（`0x1BBA0`–`0x1BBD8`）

與其他四個設施同一個形狀（`docs/re/29` §5.4），只有起點與圖號不同：

```
0x1BBA8  bl ← 4；把記錄 +0x04 起的 13 bytes 抄到 ds:7201h   ; 地點名
0x1BBC1  sub_190A6(2)                                       ; ALLPICS 第 2 張
0x1BBC6  sub_1728C                                          ; 畫面模式 ← 1
0x1BBC9  bl ← 3；sub_178A0(記錄 +0x03)                      ; 招呼字串
0x1BBD4  sub_1BE31(5)                                       ; 「要教誰？」
```

`sub_1BE31(al)` 是這個設施專用的印字串：它把字串表基址換成 **`ds:DACCh`**
（`docs/re/17` §3 的技能學習字串表）再印第 `al` 條。

## 2. 主迴圈

```
sub_1721B          ; 選人；回 0 → 離開設施
選到的人：
    ds:DAA5h ← 編號
    sub_172BB 不能行動（CON ≤ 0） → 字串 6，回頭重選
    sub_17029 ／ sub_19727                    ; 重畫那一列
    ds:4699h ← 1；ds:4698h ← 1
    bl ← 0x20；al ← 角色記錄 +0x20            ; **技能點數**
    al ＝ 0 → 字串 1，回頭重選
    sub_1BDFF(0) 回 CF ＝ 1 → 字串 2，回頭重選 ; 沒有可學的
    選一個技能：
        sub_1CA8D → 基礎費用（技能資料 +0x00 & 7，docs/re/31 §3）
        費用 > 技能點 → 字串 7 ＋ 音效 7，回頭
        否則 sub_1BE1F ／ sub_1BE16 ／ 學起來
```

⚠ **「技能點數為 0」與「沒有可學的」是兩條不同的路**（字串 1 與 2），
兩條都**回到選人**而不是離開設施——與商店的「背包滿了」同一種形狀
（`docs/re/42` §2）。

## 3. 費用與學習規則不在這裡

扣點與升級的算法早就解過，這一支只是把它們串起來：

| 規則 | 在哪 |
|---|---|
| IQ 需求 ＝ 技能資料 `+0x00 >> 3` | `docs/re/31` §3 |
| 基礎費用 ＝ 技能資料 `+0x00 & 7` | `docs/re/31` §3 |
| 升到等級 L 的費用 ＝ 基礎 × 2^(L−1) | `docs/re/31` §3 |
| 技能陣列在記錄 `+0x80`、每筆 2 bytes | `docs/re/15` |

## 4. 還沒讀的

- `sub_1BDFF`：可學技能的清單怎麼組（回 CF ＝ 1 表示沒有）。
  它與商店的 `sub_16DB4`／`sub_16D34` 是同一套清單框架，那套框架仍未解
  （`docs/re/42` §7）。
- `sub_1BE16`／`sub_1BE1F`／`sub_1BE2A`／`sub_1BE40` 四支短常式的分工。

## 5. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/trainer2.json 0x1BBA0

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/trainer.json 0x1BE31
```

## 6. 這一輪學到的（寫成規則）

- **五個設施的開頭是同一個模板**，差別只有三個常數（名稱起點、圖號、
  招呼字串的位移）。讀第五支的時候先假設它一樣，再去找不一樣的地方，
  比從頭讀快得多——而且**不一樣的地方就是它的身分**。
- **「這條路走不通」在原版幾乎都是回上一層，不是離開。** 商店的背包滿了、
  訓練師的沒點數、沒東西可學，三處都是印一句話回到選人。
  重製版把它們寫成錯誤分支會讓玩家莫名其妙被踢出設施。
