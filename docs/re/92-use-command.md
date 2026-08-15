# 92：`USE` 指令的第一層 —— Skill／Item／Attribute 三選一

日期：2026-08-15 ｜ 接 `docs/re/91`（指令列的七個入口）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`USE` 是指令列七項裡最深的一支，也是**擋住「走完主線」驗收**的那一項——
很多劇情觸發要用技能或物品。這一份解開它的第一層與三條分支的入口。

---

## 1. 挑人：`0x13A80`

```
0x13A80  sub_19727；sub_1728C
0x13A86  al ← ds:4653h（隊伍人數）
0x13A89  ＝ 1 → loc_13AAF          ; **一個人就不問**，直接用他
0x13A8D  al ← 2；sub_16CB2         ; 印訊息 2（`Which player?`）
0x13A92  al ← 0；sub_1721B         ; 挑隊員（1–9）
0x13A97  ＝ 0 → jmp sub_163C4      ; 取消 → 回地圖
loc_13AAF:
0x13AAF  ds:A5E5h ← al             ; 記住是誰
0x13AB2  sub_172BB                 ; CON ≤ 0（docs/re/89 §2）
0x13AB5  → 不行就印訊息 3 回地圖
loc_13AC2:
0x13AC2  sub_13AE4（挑要用什麼）
0x13AC8  失敗 → 回頭重問
0x13ACB  ds:4661h ← ds:A5E5h
0x13AD2  sub_13C58（施用）
```

## 2. 三選一：字母表 `ds:A5E8h`

```
ds:A5E8h = 53 49 41 00 …   ＝ "SIA\0"
```

| 索引 | 鍵 | 是什麼 | 分支 |
|---:|---|---|---|
| 0 | `S` | **Skill**（技能） | `loc_13B29` |
| 1 | `I` | **Item**（物品） | `loc_13B24` → `loc_13C0B` |
| 2 | `A` | **Attribute**（屬性） | `loc_13BDE` |

`sub_13AE4` 的開頭：

```
0x13AE4  ds:A5E1h ← al            ; 隊員編號
0x13AE7  sub_13C52
0x13AEA  al ← ds:472Ch
         ≠ 0 → al ← 0，跳過選單直接走技能那條   ; ← **旗標**
loc_13AF6:
0x13AF6  al ← 4；sub_16CB2        ; 印訊息 4（問要用什麼）
0x13AFB  ds:4680h ← 0A5E8h        ; ↑ 那張字母表
0x13B01  ds:7DF3h ← 0x404         ; 熱區遮罩（docs/re/43）
0x13B0D  sub_173B0                ; 等按鍵 → 字母表索引
0x13B10  js → stc; retn           ; ESC ＝ 取消
loc_13B12:
0x13B12  ds:A5E3h ← al            ; 存起來，之後 sub_13C58 要看
         0 → loc_13B29／1 → loc_13C0B／其餘 → loc_13BDE
```

⚠ **`ds:472Ch` 非 0 時不問，直接當成選了技能**。那個旗標是什麼還沒追——
它在 `sub_13AE4` 裡被讀了三次（`0x13AEA`、`0x13B2F`、`0x13B6E`），
形狀像「這次呼叫是從別的地方進來的」（`sub_13AE4` 的另一個呼叫端是 `0x1240E`）。

## 3. 技能那條：清單來自角色記錄

```
loc_13B29:
  al ← ds:A5E1h；sub_19614          ; 選中那個角色
  ds:472Ch ≠ 0 → 直接取 +0x80 那一格，是 0 就放棄
  否則:
    ds:7DF3h ← 0x434
    sub_198F0                        ; 清單選擇（回索引）
    ＝ 0FFh → 整個重問
    ≥ 0FDh → 回 loc_13B29
loc_13B5E:
  al <<= 1；al += 0x7Eh              ; 索引 → 記錄位移
  al ← [角色 + 那個位移]              ; **技能編號**
  ds:A5E4h ← al
```

`al × 2 + 0x7E` 是技能陣列的定址：索引 1 → `+0x80`、索引 2 → `+0x82`……
與 `docs/re/15` 的「`+0x80` 起 30 × 2」一致（清單的索引從 1 起算）。

推論等級：**已確認**（三選一的字母表是資料裡直接讀出來的；
技能陣列的定址與既有筆記互相印證）。

## 4. 還沒解的

- `sub_13C58`（443 bytes）：**施用**那一半。開頭依 `ds:A5E0h`（＝ 選項）
  分三條，其中還有 `sub_13E32`／`sub_13E13` 兩支沒讀。
- `loc_13C0B`（物品）與 `loc_13BDE`（屬性）兩條分支的內容。
- `ds:472Ch` 旗標的語意，以及 `sub_13AE4` 的另一個呼叫端 `0x1240E`。
- `sub_13C52`、`sub_19727`、`sub_16F20` 三支輔助函式。

## 5. remake 這一側

**還沒接**。`USE` 目前按下去只印一句 `Use: not wired yet`（`docs/re/91` §3）。
接上去要三條分支都有，缺一條就會變成「按了沒反應」——
而那是最難查的一種半成品（`docs/re/79` §5 的教訓：空殼畫得出來、測試全綠）。

## 6. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/use_full.json 0x13AE4 --callers

python3 -c "
d=open('workplace/analysis/unpacked/wl.merged.exe','rb').read()
off=0x1CE20+0xA5E8-0x10000+0xb0
print(d[off:off+8])"
```

## 7. 這一輪學到的（寫成規則）

- **選單的字母表就是規格書。** 三個 byte `53 49 41` 一次講完了 `USE` 有幾條路、
  分別是什麼——比追三條分支的程式碼快得多。
  **遇到「等按鍵 → 查表 → 分支」的形狀，先把那張表倒出來。**
- **`export_function.py` 給的是完整 listing，`export_forced.py` 只給一個 chunk。**
  這一輪先用 forced 拿到零散片段，換成 function 之後一次拿到 366 bytes。
  **函式已經被 IDA 認出來時用前者，沒被認出來（像 `loc_1B8AD`）才用後者。**
