# 19：資料驅動的效果系統與傷害計算

日期：2026-08-14 ｜ 對應盤點 **D3**（武器傷害）、**E5**（技能檢定與陷阱）、部分 **E2**

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

---

## 1. 結論

地圖記錄對角色做的所有事——扣血、給錢、加屬性——**都由記錄裡的兩個 byte 描述**，
不是每種效果各寫一段程式碼。這兩個 byte 說的是：

> 「把角色記錄的**第 X 個 byte**，**加上／減去**一個**固定值或 N 顆 d6**。」

只有兩個欄位編號被特判：`0x1D`（體力）走完整的護甲流程，
`0x15`／`0x21`（兩個 24-bit 計數器）走進位鏈。其餘一律當單一 byte 加減。

## 2. 效果記錄：`+0x08` 與 `+0x09`

`sub_14193`（103 bytes、3 個呼叫端）從目前的地圖記錄（`ds:46AEh`）取兩個 byte：

```
0x1419C  b308      mov  bl, 8
0x1419E  8b3eae46  mov  di, ds:46AEh
0x141A2  8a01      mov  al, [bx+di]      ; 記錄 +0x08
0x141A4  84c0      test al, al
0x141A6  74ea      jz   short locret     ; 0 ＝ 沒有效果
0x141A8  50        push ax
0x141A9  247f      and  al, 7Fh
0x141AB  a2c6a5    mov  ds:0A5C6h, al    ; ← 要改的欄位位移
0x141AE  58        pop  ax
0x141AF  84c0      test al, al
0x141B1  790f      jns  short loc_141C2  ; bit7 ＝ 0 → 擲骰
0x141B3  fec3      inc  bl               ; ┐ bit7 ＝ 1：固定值
0x141B9  8a01      mov  al, [bx+di]      ; │ 記錄 +0x09
0x141BB  247f      and  al, 7Fh          ; │
0x141BD  b200      mov  dl, 0            ; ┘
0x141BF  eb12      jmp  loc_141D3
loc_141C2:                                ; bit7 ＝ 0：Nd6
0x141C4  fec3      inc  bl
0x141C8  8a01      mov  al, [bx+di]      ; 記錄 +0x09
0x141CA  247f      and  al, 7Fh          ; N
0x141CC  8ad0      mov  dl, al
0x141CE  b000      mov  al, 0
0x141D0  e8b3bb    call sub_19D86        ; ← 0 ＋ N 顆 d6
loc_141D3:
0x141D3  e81a7a    call sub_19BF0        ; ds:46BEh/46BFh ← 值（16-bit）
…
0x141DD  b309      mov  bl, 9
0x141E3  8a01      mov  al, [bx+di]      ; 記錄 +0x09 **完整 byte**
0x141E5  8ad8      mov  bl, al           ; → 傳給 sub_141FA 當加減旗標
0x141EE  e809 00   call sub_141FA
```

| byte | bit 7 | bit 0–6 |
|---|---|---|
| `+0x08` | 0 ＝ 值是 **N 顆 d6**；1 ＝ 值是**固定值** | **要改的角色記錄欄位位移** |
| `+0x09` | 0 ＝ **加**；1 ＝ **減** | 固定值，或骰子數 N |

`+0x09` 一個 byte 同時提供「值」與「加減」——低 7 位當值、bit 7 當號誌。

推論等級：**已確認**（程式碼讀完；兩條分支都落回同一個 `+0x09`）。

## 3. 套用：`sub_141FA`

```
                       bl（＝記錄 +0x09 完整 byte）
                              │
              ┌───────────────┴───────────────┐
          bit7 ＝ 1（減）                bit7 ＝ 0（加）
              │                               │
   欄位 ＝ 0x1D ─→ sub_157D6（傷害，走護甲）   欄位 ＝ 0x21 ─→ sub_19BC0（24-bit 加）
   欄位 ＝ 0x15 ─→ sub_17B15（24-bit 減）      其餘 ─→ rec[欄位] += 值（溢位就不加）
   其餘 ─→ rec[欄位] −= 值（下限 0）
```

原始碼：

```
0x1420C  84db      test bl, bl
0x1420E  795a      jns  short loc_1426A      ; 加
0x14210  8a1ec2a5  mov  bl, ds:0A5C2h        ; 欄位位移
0x14214  80fb1d    cmp  bl, 1Dh
0x14217  7416      jz   short loc_1422F      ; 體力 → 傷害流程
0x14219  80fb15    cmp  bl, 15h
0x1421C  7438      jz   short loc_14256      ; 24-bit 計數器 → 借位鏈
0x1421E  8b3eb546  mov  di, ds:46B5h
0x14222  8a01      mov  al, [bx+di]
0x14224  2a06be46  sub  al, ds:46BEh
0x14228  7302      jnb  short loc_1422C
0x1422A  b000      mov  al, 0                ; 減到負的就夾在 0
0x1422C  8801      mov  [bx+di], al
```

**這就是 E5（門／鎖／陷阱／技能檢定）與 E2（地圖事件）對角色產生後果的唯一出口。**
remake 只要實作這一支，所有地圖事件的效果就都對了——不需要逐一還原每種陷阱。

## 4. 傷害與護甲：`sub_157D6`

進入時 `al`／`dl` ＝ 16-bit 傷害、`bl` ＝ 目標角色槽。

```
sub_19614(bl)                       ; → ds:46B5h ＝ 目標角色記錄
if CON（+0x1D/+0x1E）＝ 0：stc; retn  ; 已經死了，不再處理

; ── 護甲吸收 ──
dl = rec[+0x1A]                     ; AC
al = 0
sub_19D86                           ; ← 擲 AC 顆 d6 相加
實際傷害 = 傷害 − 護甲吸收（16-bit 借位）
if 實際傷害 <= 0：印「沒造成傷害」並返回

; ── 扣血 ──
sub_15A84(實際傷害)                  ; 印「N 點傷害」，含單複數
sub_19CE9(實際傷害)                  ; CON −= 實際傷害（16-bit）
rec[+0x26]/rec[+0x27] ← 受傷前的值
if CON 低於 −50：sub_15BFE
sub_19A1D → 傷勢等級 → 查 ds:A69Eh → 印狀態訊息
```

**護甲的機制是擲骰減傷：吸收量 ＝ AC 顆 d6 的和。**
`sub_19D86` 的呼叫是 `al = 0, dl = rec[+0x1A]`，也就是 `0 + AC × d6`。

扣血本體（`sub_19CE9`，21 bytes）是乾淨的 16-bit 減法：

```
0x19CE9  8ad4      mov  ah, dl
0x19CEB  b31d      mov  bl, 1Dh
0x19CED  8b3eb546  mov  di, ds:46B5h
0x19CF1  8b09      mov  cx, [bx+di]      ; CON
0x19CF3  2bc8      sub  cx, ax
0x19CF5  890b      mov  [bx+di], cx      ; CON −= 傷害
```

**CON 是有號 16-bit，會掉到負的**，傷勢等級就是靠負值分級
（門檻 −11／−20／−30／−40，`docs/re/15` §3）。

推論等級：**已確認**（三支函式讀完）。

## 5. 單複數在這裡也出現

`sub_15A84`（25 bytes）決定訊息要用單數還是複數：

```
0x15A84  b302      mov  bl, 2            ; 預設複數
0x15A86  80fa00    cmp  dl, 0
0x15A89  7506      jnz  short loc_15A91
0x15A8B  3c01      cmp  al, 1
0x15A8D  7502      jnz  short loc_15A91
0x15A8F  fecb      dec  bl               ; 值恰為 1 → 單數
0x15A91  881e8746  mov  ds:4687h, bl
0x15A95  e83821    call sub_176D0        ; 印數字
0x15A98  b059      mov  al, 59h
0x15A9A  e915 12   jmp  sub_16CB2        ; 印訊息 0x59
```

`ds:4687h` 就是 `docs/re/17` §4.1 那個 `\n` 分隔機制的選擇器：
1 ＝ 用單數字尾、2 ＝ 用複數字尾。**中文化時這個值沒有意義，但格式要能 round-trip。**

## 6. 兩個 24-bit 計數器

| 欄位 | 語意 | 路徑 | 函式 |
|---|---|---|---|
| `+0x15`–`+0x17` | **金錢** | 減（先比大小再借位扣） | `sub_17B3E` ＋ `sub_17B15` |
| `+0x21`–`+0x23` | **經驗值** | 加（進位鏈） | `sub_19BC0` |

兩個都是 24-bit、都只有一個方向的路徑，但**是不同欄位**。
語意由別的地方定案：金錢在商店比價與扣款（`docs/re/22` §5）、
經驗值在升級門檻 `(L² − L) × 512` 的比較（`docs/re/31` §1）。

⚠ 效果記錄可以直接指定欄位 `0x21`（§3 的分派），也就是**地圖事件本身就能給經驗值**——
不是只有戰鬥才給。

## 7. 還沒解的

- **命中判定**：`sub_157D6` 收到的是已經算好的傷害，計算它的地方還沒讀
  （`sub_141FA` 的另一個呼叫端 `0x1444F`、以及 `0x1B0C6` 那條路徑）。
- `sub_15BFE`（CON 低於 −50 時呼叫）做什麼。
- 地圖記錄的 `+0x08/+0x09` 是哪個 section 型別才有——目前只知道
  `sub_14193` 會讀，取記錄的路徑還沒往回追到 section 型別。

## 8. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py workplace/analysis/dumps/effects.json \
  0x14193 0x141FA 0x157D6 0x19CE9 0x15A84 0x19D86 0x19BC0 0x17B15 --callers

python3 tools/summarize_record_fields.py workplace/analysis/dumps/listing.json \
  > docs/re/generated/ida94/record-fields.json
```

## 9. 這一輪學到的（寫成規則）

- **看到「每種效果一段程式碼」的預期落空時，先找共用出口。** 陷阱、寶箱、
  技能檢定看起來是三件事，實際上共用 `sub_141FA` 一支，差別全在資料。
  remake 照著資料驅動實作，比逐一還原省一個數量級。
- **線性追蹤暫存器會把互斥分支疊加。** `sub_14193` 兩條分支各 `inc bl` 一次，
  線性掃描報出 `+0x0A`，而實際上兩邊都是 `+0x09`。
  `tools/summarize_record_fields.py` 已改成**在往前跳的目標清掉追蹤狀態**
  （往回跳不清——迴圈進入前的那個值正是陣列起點）。
  修正後 `ds:46AEh` 的位移從 12 個降為 11 個，少掉的那個本來就不存在。
