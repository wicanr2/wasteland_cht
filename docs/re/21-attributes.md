# 21：七個屬性、修正值階梯與檢定骰

日期：2026-08-14 ｜ 對應盤點 **D1**（屬性與技能表、角色建立）、補完 **D2**／**D3**

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

---

## 1. 七個屬性

字串表 `ds:AB3Eh` 的第 11 條就是選單本身：

```
'Which:\r\r\x11\x101) Strength\r\x102) IQ\r\x103) Luck\r\x104) Speed\r
 \x105) Agility\r\x106) Dexterity\r\x107) Charisma\r'
```

印它的地方（`sub_13AE4`）緊接著把按鍵換成記錄位移，一行就釘死了配置：

```
0x13BDE  b00b      mov  al, 0Bh
0x13BE0  e8cf30    call sub_16CB2        ; 印上面那個選單
0x13BEF  e89e32    call sub_18E90        ; 等按鍵
0x13BF4  3c31      cmp  al, 31h          ; ┐ 只收 '1'–'7'
0x13BF9  3c38      cmp  al, 38h          ; ┘
0x13BFE  2c31      sub  al, 31h          ; → 0–6
0x13C00  040e      add  al, 0Eh          ; → 0x0E–0x14
```

| 屬性 | 記錄位移 |
|---|---|
| Strength | `+0x0E` |
| IQ | `+0x0F` |
| Luck | `+0x10` |
| Speed | `+0x11` |
| Agility | `+0x12` |
| Dexterity | `+0x13` |
| Charisma | `+0x14` |

同一段的另一條分支是 `shl al, 1 ／ add al, 0BBh`——選技能陣列
（`docs/re/15` §2.1），並用 `dl` 標記選到的是屬性（2）還是技能（1）。

推論等級：**已確認**（選單文字、按鍵範圍檢查、位移算術三者一致）。

### 1.1 獨立佐證：四支扇入

戰鬥程式碼裡有四個入口，各自只設一個 `bl` 就跳進同一段：

```
0x15705  sub_15705:  mov bl, 10h   ; Luck
0x1570A  loc_1570A:  mov bl, 13h   ; Dexterity
0x1570F  loc_1570F:  mov bl, 12h   ; Agility
0x15714  loc_15714:  mov bl, 0Eh   ; Strength
0x15716  loc_15716:  共用尾段
```

四個值全部落在 `0x0E`–`0x14` 區間內，而且是四個不同的屬性——
這是與 §1 完全獨立的一條證據。

## 2. 屬性 → 修正值（`loc_15716`）

共用尾段把屬性值換成一個有號修正值：

```
0x15716  b200      mov  dl, 0
0x15718  8b3eb546  mov  di, ds:46B5h
0x1571C  8a01      mov  al, [bx+di]      ; 屬性值 v
0x1571E  3c0d      cmp  al, 0Dh
0x15720  f5        cmc
0x15721  7210      jb   short loc_15733  ; v >= 13
0x15723  3c09      cmp  al, 9
0x15725  f5        cmc
0x15726  7303      jnb  short loc_1572B  ; v <  9
0x15728  b000      mov  al, 0            ; 9 <= v <= 12 → 0
0x1572A  c3        retn

loc_1572B:                                ; v < 9（負修正）
0x1572B  2c09      sub  al, 9            ; v − 9（負數）
0x1572D  feca      dec  dl               ; 高位 ← 0xFF
0x1572F  f9        stc                   ; ┐ 帶號右移一位
0x15730  d0d8      rcr  al, 1            ; ┘
0x15732  c3        retn

loc_15733:                                ; v >= 13（正修正）
0x15733  2c0c      sub  al, 0Ch          ; v − 12
0x15735  d0e8      shr  al, 1
0x15737  c3        retn
```

寫成式子：

```
v ≤ 8       → 修正 ＝ floor((v − 9) / 2)     （負）
9 ≤ v ≤ 12  → 修正 ＝ 0
v ≥ 13      → 修正 ＝ (v − 12) >> 1          （正）
```

實際值：

| v | 4 | 5 | 6 | 7 | 8 | 9–12 | 13 | 14 | 15 | 16 | 18 | 20 |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| 修正 | −3 | −2 | −2 | −1 | −1 | 0 | 0 | +1 | +1 | +2 | +3 | +4 |

**中間有一段死區**：9–13 都是 0（13 落在正的分支，但 `(13−12)>>1 ＝ 0`）。
負的那一半用 `stc ／ rcr` 做帶號右移——這是 8086 沒有 `sar` 以外選擇時的常見寫法，
效果是 floor 除法而不是截斷。

推論等級：**已確認**（逐條讀完，數值以位元運算逐一驗算過）。

## 3. 檢定骰：`sub_19C84`

`sub_19C84`（2d6 逢同點續擲，`docs/re/13` §3.4）**不是角色建立用的**，
是通用的**檢定骰**。11 個呼叫端幾乎都是同一個形狀：

```
call sub_19BF8      ; 累加器 ds:46C0h ← 0
call sub_19C84      ; 2d6（逢同點續擲）
call sub_19C2C      ; 累加器 += 骰值
mov  bl, 11h        ; ← 例：Speed
mov  di, ds:46B5h
…                   ; 再把屬性修正加進去
```

或是直接當成功判定：

```
call sub_19C84
cmp  al, 5
cmc
jnb  失敗
```

`sub_19C84` 的期望值是 8.4、最小值 3（`docs/re/13` §3.4 已驗），
所以門檻 4／5 是很寬鬆的檢定。

**這修正了先前把它標成「屬性生成入口」的推測**——
角色建立用的是另一支（`sub_1CAD1`，見 §5.1）。

## 4. 這些修正用在哪

`docs/re/20` §6 列了「隊伍攻擊路徑的傷害由五個來源累加，各來源未解」，
現在其中三個有答案了：

| 來源 | 內容 |
|---|---|
| `sub_15705` | **Luck** 的修正值 |
| `loc_1570A` | **Dexterity** 的修正值 |
| `loc_15714` | **Strength** 的修正值 |
| `sub_15755` | 武器傷害骰（不是屬性修正，[`45`](45-item-data-and-weapon-damage.md) §4） |
| `sub_182FA` | 裝備武器的技能等級 × 3（`docs/re/20` §4.1） |

命中判定的累加器（`sub_1B108`，`docs/re/20` §3）用的是 `loc_1570F`＝**Agility**。

所以目前可寫出來的是：

```
命中累加值 ＝ 基礎 ＋ 技能等級×3 ＋ Agility 修正 ± 距離修正，夾在 100
隊伍傷害   ＝ … ＋ Strength 修正 ＋ Dexterity 修正 ＋ Luck 修正 ＋ 兩個未解來源
```

## 5. 角色建立：`sub_1C6C9`

印 `'Create a character.'`／`'Keep this char?'` 的那一支（789 bytes、1 個呼叫端）。

```
找空槽                      掃隊伍，rec[+0x29] ＝ 0 的算空；滿 4 個就印
                            'You cannot create any more characters.'
整筆清零                    rec[0..255] ← 0
性別                        roll(1..2) → 決定用哪張輸入樣板（ds:DECFh／ds:DED9h）
七個屬性                    bl 從 0x0E 跑到 0x14，每格 ＝ sub_1CAD1()
MAXCON ＝ CON               sub_1CAD1() ＋ 18，兩者寫成同一個值
rec[+0x24] ← 1              等級 ＝ 1（升級規則見 docs/re/31）
rec[+0x20] ← rec[+0x0F]     技能點 ＝ IQ
rec[+0x32…] ← "PRIVATE"     從 ds:DEC7h 複製到 0 為止（初始階級字串）
```

`rec[+0x20] ← IQ` 與字串 `'Skill points = '`、`'   IQ PTS LVL   SKILL'`
（`docs/re/17`）互相對上，所以 **`+0x20` 是可用技能點**。

### 5.1 屬性擲法：`sub_1CAD1` ＝ 5d6 取最高三顆

```
0x1CAD2  b204      mov  dl, 4
0x1CAD4  e888c3    call sub_18E5F        ; ┐ d6 × 5，存進 5 格暫存
0x1CAD9  8884f1df  mov  [si-200Fh], al   ; │
0x1CADD  feca      dec  dl               ; │
0x1CADF  79f3      jns  short loc_1CAD4  ; ┘ dl ＝ 4,3,2,1,0
0x1CAE1  …         氣泡排序（升冪）
0x1CB10  b000      mov  al, 0
0x1CB12  b303      mov  bl, 3            ; ┐ 從最大的那端加三個
0x1CB14  b204      mov  dl, 4            ; │
0x1CB18  0284f1df  add  al, [si-200Fh]   ; │
0x1CB20  75f4      jnz  short loc_1CB16  ; ┘
```

**擲 5 顆 d6，排序後把最大的三顆相加。**

| | 值 |
|---|---|
| 值域 | 3 – 18 |
| 期望值 | **13.43** |
| 眾數 | 14（14.85%） |
| ≤ 8 的機率 | 4.1% |
| ≥ 16 的機率 | 23.4% |

擲法配 §2 的修正值階梯看很合理：期望值 13.43 落在死區（9–13）的上緣，
所以**大多數屬性給 +1 左右的修正**，極端值才拉開差距。

MAXCON 起始 ＝ 同一支再擲一次 ＋ 18，也就是 **21–36**。

推論等級：**已確認**（程式碼讀完；分佈以窮舉 6⁵ 種組合算出）。

### 5.2 一個要看原始 bytes 才不會讀錯的地方

清零那一段，反組譯文字看起來像「把 `bl` 寫進 `rec[bl]`」：

```
0x1C71B  b3 00        mov  bl, 0
0x1C71D  8a c3        mov  al, bl
0x1C71F  8b 3e b5 46  mov  di, ds:46B5h
0x1C723  88 01        mov  [bx+di], al
0x1C725  fe c3        inc  bl
0x1C727  75 f6        jnz  short …
```

`75 f6` 的目標是 `0x1C729 − 10 ＝ 0x1C71F`，**不是 `0x1C71D`**——
`mov al, bl` 在迴圈外，只執行一次，所以 `al` 全程是 0，這是乾淨的清零。
只看反組譯很容易把它讀成「填入 0,1,2,…,255」。

## 6. 還沒解的

- 名字輸入的樣板（`ds:DECFh`／`ds:DED9h`／`ds:DEE3h` ＋ `sub_1C9DE`）。
- 國籍在建立時怎麼選（`+0x19`，`docs/re/17` §4.3 已知欄位語意）。
- `+0x0D`（屬性區前一個 byte）是什麼。
- IQ（`+0x0F`）、Speed（`+0x11`）、Charisma（`+0x14`）用在哪裡還沒逐一追。
  `docs/re/15` 記過 `+0x0F` 有 15 處存取、`+0x11`／`+0x13`／`+0x14` 各有幾處。
- 屬性上限（`sub_19394` 那條路徑會不會擋）。

## 7. 可重跑的完整指令

```bash
python3 tools/decode_text.py workplace/analysis/unpacked/wl.merged.exe \
  docs/re/generated/ida94/exe-strings.json     # 表 0xAB3E 第 11 條就是屬性選單

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/attributes.json \
  0x13AE4 0x15705 0x19C84 0x1B108 0x1C6C9 0x1CAD1 --callers
```

`0x1570A`／`0x1570F`／`0x15714`／`0x15716` 不是函式起點，
要用 `export_listing.py` 的逐指令 JSON 讀。

## 8. 這一輪學到的（寫成規則）

- **選單文字就是欄位表。** `1) Strength … 7) Charisma` 加上緊接著的
  `sub al,'1' ／ add al, 0Eh`，兩行就把七個欄位位移全部定死——
  比從存取點逐一反推快一個數量級。**解出字串表之後，先回頭看選單類的字串**。
- **`jcc` 的目標要用原始 bytes 算，不要看反組譯的標籤位置。** §5.2 那個清零迴圈，
  目標差兩個 byte 就是「全部清零」與「填入 0..255」的差別。
- **「入口已定」不等於「猜對了」。** `sub_19C84` 先前被標成屬性生成的入口，
  實際上是通用檢定骰。列入口的時候要寫清楚憑什麼（這次憑的是「分佈特別」，
  那只夠支持「它是某種檢定」，不足以支持「它是角色建立」）。
