# 121：胖佛萊迪那一題答 No ＝ 毒氣，不是當場開打

日期：2026-08-20 ｜ 接 `docs/re/46` §4（nibble 8 問答）、`docs/re/70`（nibble 1）、
`docs/re/71`（nibble 12 批次改寫）、`docs/re/34`（腳本指令表）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 的雜湊見 `docs/re/01`。

攻略說「對胖佛萊迪回答 `NO`，他隨即翻臉開火」（M-21）。結果會開打沒錯，
但中間隔著三層改寫與一個**防毒面具檢定**——而那個檢定的分支方向，
remake 抄反了。

---

## 1. 那一題

資源 40（game2，拉斯維加斯）section 8 記錄 2，單鍵模式：

```
Fat Freddy looks at you expectantly. Will you take his offer?
  Yes / No
```

答案兩條（`Y`／`N`），後續動作三組（`docs/re/46` §4 的
「答案數 ＋ 1」，最後一組是答錯）：

| 答 | 改寫成 | 那是什麼 |
|---|---|---|
| `Y` | (1, 33) | 「"Good, good!" ...Here's $1000 on account.」→ 再改寫成 (5, 11) ＝ 寶箱 |
| `N` | (1, 1) | §3 那條鏈 |
| 其他鍵 | (8, 2) | **原地**——問題再問一次 |

## 2. 答 No 的第一層：nibble 1 記錄 1

```
訊息：「"You may change your mind," he sneers, and pushes a button on his desk.
       You hear a faint hissing sound...」
收尾改寫 → (6, 11)
```

nibble 1 是氛圍敘述串列 ＋ 收尾改寫（`docs/re/70` §1），這一筆只有一條訊息，
所以位移 1 那兩個 byte 就是新的 `(nibble, 記錄)`。

## 3. 第二層：section 6 記錄 11 ＝ opcode 7

那一格變成腳本格。記錄本體只有 7 bytes：

```
01 00 00 | 01 00 | 0c 0d
 ↑  sel ＝ 1 → opcode 表[1] ＝ 7
           +0x03/+0x04 ＝ (1, 0)   ← 有人帶防毒面具
                     +0x05/+0x06 ＝ (12, 13)  ← 一個都沒有
```

opcode 7（`0x1A699`）：全隊掃**物品 0x2F ＝ Gas mask**，
沒有的人 `CON ← −5`（舊值備份到 `+0x26`）。

## 4. 分支方向：有人帶著走 `+0x03`

```
0x1A699  ds:CF2Ch ← 1                     ; **旗標初值 1**
loop:    sub_19614(隊員)；sub_1968C(2Fh)
0x1A6A9  jb  0x1A6B4                      ; CF ＝ 1 ＝ **找不到** → 去扣血
0x1A6AB  al ← 0；ds:CF2Ch ← 0             ; 找到 → 清旗標，跳過扣血
0x1A6B4  記錄 +0x1D（CON）← 0FFFBh，舊值 → +0x26
0x1A6D4  bl ← 3
0x1A6D6  ds:CF2Ch ＝ 0 → 保持 3           ; **有人帶著 → +0x03**
0x1A6DD  否則 bl ← 5                      ; 一個都沒有 → +0x05
```

旗標是「**還沒有人帶著**」，不是「有人帶著」——初值 1、找到才清 0，
所以 `test/jz` 那一步保持的 3 才是有面具的那條。

資料端印證：`+0x03` 指到的 (1, 0) 是
「Those with gas masks manage to get them on in time.」——
方向如果反過來，這句話會印在一個面具都沒有的隊伍頭上。

⚠ remake 這一側原本抄反了（`internal/game/script.go` 的 `OpNeedItem`），
單元測試也照著反的方向寫，所以兩邊互相印證了一個錯的答案。
**光看程式碼與測試不會發現，是資料把它抓出來的。**

## 5. 兩條路各自發生什麼

### 有人戴上面具 → (1, 0) → (12, 18)

nibble 12 的批次改寫（`docs/re/71`）只有一筆，旗標 `0x80` ＝ 最後一筆：

```
訊息 26：「Combat flares. Freddy's henchmen go for their guns,
          and may the Cloud help anyone who gets in their way!」
(6, 23) ← (nibble 3, 記錄 11)     ; **放一組遭遇**
腳下     ← (0, 0)                 ; 問答那一格清成空地
```

**這就是攻略作者遇到的「翻臉開火」。**

### 一個都沒有 → (12, 13)

```
訊息 0x65：「Everything is spinning. The gas knocks you out
            in just a couple of seconds.」
11 筆改寫，最後一筆旗標 0xC0（最後一筆 ＋ 跳過重畫）：
  (1,12) (1,11) (2,11) (2,12) → nibble 2（條件串列）
  (4, 9)                      → nibble 2 記錄 7
  (6,23)                      → **nibble 3 記錄 11（同一組遭遇）**
  (1,10) (2,10) (3,10) (3,11) (3,12) → nibble 1 記錄 35
      （「Looks like your buddies have run into a bit of a jam.
        They are bound by ropes which you swiftly untie...」）
腳下 ← (10, 4) ＝ **傳送**：同一張地圖（`+0x03` ＝ 0x28 ＝ 40）的 (1, 12)
```

沒帶面具的人 CON ＝ −5 躺著，隊伍被搬到牢房那一區，同伴綁著繩子。
**遭遇一樣放在 (6, 23)**，只是要先脫身才走得回去。

推論等級：§1–§5 全部**已確認**（程式碼與出貨資料互相印證）。

## 6. 對攻略 M-21 的結論

> 對 Fat Freddy 的要求回答 `NO` 會直接觸發戰鬥

**相符**，但「直接」要拆開：答 `No` 觸發的是**毒氣**，戰鬥是毒氣結算之後
放下去的那一組遭遇。隊上有防毒面具就當場開打，一個都沒有就先被毒昏、
搬進牢房，那一組遭遇仍然在原地等著。

## 7. 順帶修好的兩個工具洞

`tools/scan_item_refs.py` 的兩個假零，兩個都會讓「這個物品沒有人用」變得不可信：

| 洞 | 症狀 | 修法 |
|---|---|---|
| 指標表的**第 0 項可以是 0** | `array()` 拿第 0 項當「第一筆記錄」→ 讀到 0 → 整個 section 跳過。資源 40 的寶箱表就是這樣，43 格報成 42 格 | 取**第一個非 0 的項**，再檢查它容不容得下自己的索引 |
| 腳本的閘門是「地圖上有沒有 nibble 6 的格子」 | **出貨地圖沒有 nibble 6 的格子不代表沒有腳本**——問答的答案分支會把某一格改寫成 (6, N)，這一份追的那條鏈正是如此。42 筆腳本記錄報成 42 筆，實際是 141 筆 | 閘門改成「section 6 的指標有沒有和別的型別撞在一起」（撞了才是這張圖沒有這個 section）|

修完之後正對照仍然全綠（opcode 全部落在 0–43、寶箱類別全部落在 0–18），
物品 48 的「0 次」也仍然成立（`docs/re/120` §1）——**只是現在的證據強了三倍**。

⚠ `docs/re/76` 把「0 格的 opcode 只能靠改寫到得了」的改寫路徑列成
批次（`71`）與傳送落點（`73`），**漏了問答的答案分支**（`46` §4.1）。
opcode 7 就是靠它到得了的。

## 8. 可重跑的完整指令

```bash
# 問答記錄與答案
python3 tools/summarize_questions.py workplace/analysis/unpacked/wl.merged.exe \
  workplace/orig/wastland/game1 workplace/orig/wastland/game2 \
  docs/re/generated/ida94/block-strings.json out.md

# 腳本記錄與 opcode（修過 array() 之後）
python3 tools/scan_item_refs.py workplace/analysis/unpacked/wl.merged.exe \
  workplace/orig/wastland/game1 workplace/orig/wastland/game2 out2.md

# opcode 7 的本體
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_function.py \
  workplace/analysis/dumps/target.json 0x1A699
```

## 9. 這一輪學到的（寫成規則）

- **旗標的名字要從初值讀起。** `ds:CF2Ch` 初值 1、找到才清 0，所以它是
  「還沒有人帶著」。把它當成「有人帶著」，整個分支就反了，
  而**程式碼與測試會互相印證那個錯**——只有資料能抓出來。
  **布林分支解完，回頭找一筆真實資料驗方向。**
- **「這張圖沒有 X 格」不等於「這張圖不會有 X 格」。** 地圖是會被改寫的，
  出貨資料只是初始狀態。用「出貨有沒有這種格子」當掃描閘門，
  會把**只在劇情之後才存在的東西**整批漏掉，而且是安靜地漏。
- **指標表裡的 0 是合法的洞，不是結束。** 拿第 0 項推表長會在
  「第 0 筆不存在」的表上安靜歸零。要取第一個非 0 的項。
