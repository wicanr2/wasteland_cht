# 114：遭遇記錄 `+0x09` 是「這一組跟你什麼關係」

日期：2026-08-18 ｜ 接 `docs/re/110`（`Hire` 的結算）、`docs/re/39`（遭遇掃描）、
`docs/re/78`（遭遇生成）、`docs/re/101` §6（移動計畫）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
`game1`／`game2` 見 `docs/re/01`。

`docs/re/110` §5.2 記著一個開口：雇用的旗標（遭遇記錄 `+0x09` 的 bit1）
在出貨資料裡一筆都找不到，所以「按 `H` 到底能不能真的雇到人」是未知的。

**找得到，而且有 14 筆**——它們在 **section 3**，不是 section 15。

---

## 1. 兩個 section 都是遭遇，一個靜態一個執行期

| section | 筆數 | 誰寫的 |
|---:|---:|---|
| **3** | 528 | **出貨資料**。42 張地圖共 365 格 nibble 3 指過來 |
| **15** | 120 | **執行期**。`sub_16890` 生成隨機遭遇時挑一個空槽並把那一格寫成 nibble 15（`docs/re/78` §1）|

出貨資料裡 **nibble 15 一格都沒有**——那個 nibble 只在遊戲跑起來之後才出現。

兩邊共用同一支讀取路徑：`sub_13787` 拿遭遇的 (x, y) 去查那一格的
**nibble 當 section 型別**、第 2 層當記錄編號（`sub_17CD2` → `sub_17CB1`），
所以同一份程式碼同時服務兩種來源。remake 這一側早就對了
（`game.IsEncounterCell` 收 3 與 15，`docs/re/39` §2）。

⚠ **上一輪只掃了 section 15，掃出零命中。** 零命中與「資料裡沒有」
長得一模一樣——而這裡兩個 section 的記錄**長度也一樣**（多數 12 bytes），
所以連長度都不會露出破綻。

## 2. 四個欄位

| 位元 | 意思 | 讀取端 |
|---|---|---|
| bit0 | **名字從地圖的明文名字表取**（清掉就印種類名，字串表 1 的 `0x52 + Kind`，`docs/re/86`）| `0x129E9` → `loc_12A04` 讀記錄區標頭 `+0x02`／`+0x03` |
| **bit1** | **這一組不敵對** | `sub_12AC5`（第二個 `shr` 把它送進 CF）|
| bit2 | 這一筆遭遇**不參與移動**（`docs/re/101` §6）| `0x14C55` 的 `and al, 4` |
| 高 4 位 | **section 17 的 NPC 記錄編號**（0 ＝ 沒有）| `sub_15C08`（`>> 4`）|

`sub_12AC5` 的六個消費端全部同一個方向——bit1 設起來的那一組：

| 位址 | 做什麼 |
|---|---|
| `0x14C50` | **不排移動計畫**（不會朝你走過來）|
| `0x1ADBC` | **不攻擊隊伍**（敵方回合直接跳過整段結算）|
| `0x149A2` | 不算進 `ds:A610h`「附近有敵人」|
| `0x132D5` | **可以雇用**（`Hire` 的第一道閘）|
| `0x127B9`、`0x1364B` | 描述與訊息走另一條 |

所以 bit1 讀成「**友善**」最貼切。⚠ **友善 ≠ 可雇用**：
`Hire` 還要高 4 位非 0（`0x132DD` 的 `cmp al, 0`），
而出貨資料裡 section 3 共 **126 筆友善**的遭遇，其中只有 14 筆帶得動 NPC 編號。

推論等級：**已確認**（六個消費端逐支讀完，方向一致）。

## 3. 開槍就翻臉

`sub_15C19` 是唯一會改 bit1 的地方：

```
0x15C19  bl ← al；sub_13449          ; 攻擊者 → 目標編號
         sub_13A56／sub_137F4／sub_13787 ; 選定那一筆遭遇
0x15C27  bl ← 9
0x15C2D  al ← [ds:46C6h + 9]
0x15C2F  or  al, 1                   ; 名字改用明文名字表
0x15C31  and al, 0FDh                ; **清掉友善**
0x15C33  [ds:46C6h + 9] ← al
```

兩個呼叫端：

- **`0x1AF7B`**：隊伍攻擊的結算裡，就在命中基礎值查表（`0x1AF64`，`docs/re/101` §3）
  之後、**命中判定之前**——所以**打不中也算數，開槍就翻臉**。
- `0x12DAE`：另一條路（遭遇初始化那一族），還沒細讀。

這正是攻略那句話的機制（`docs/walkthrough/swm-005` 的 M-03）：

> 友善 NPC（Christina）在戰鬥中被誤傷會立刻轉為敵對；戰鬥中沒碰到她就能直接招募

翻臉是**寫回地圖記錄**的，所以離開戰鬥也不會復原。

推論等級：**已確認**。

## 4. 出貨資料裡的 14 個人

section 3 裡 bit1 設起來、而且高 4 位指得到 section 17 記錄的：

| 圖 | 記錄 | `+0x09` | NPC | 出貨資料有格子指過來 |
|---|---:|---|---|---|
| `game1` 4 | 3/46 | `0x13` | FELICIA | — |
| `game1` 4 | 3/47 | `0x23` | ACE | (1,27) |
| `game1` 6 | 3/34 | `0x23` | DAN CITRINE | — |
| `game1` 6 | 3/35 | `0x13` | MAYOR PEDROS | — |
| `game1` 10 | 3/11 | `0x17` | JACKIE | (28,1) |
| `game1` 27 | 3/4 | `0x13` | RALF | — |
| `game1` 32 | 3/16 | `0x13` | CHRISTINA | (19,24) |
| `game1` 34 | 3/31 | `0x13` | MORT | (29,1) |
| `game2` 19 | 3/2 | `0x13` | VAX | — |
| `game2` 21 | 3/0 | `0x17` | MAD DOG FARGO | — |
| `game2` 21 | 3/1 | `0x27` | METAL MANIAC | — |
| `game2` 36 | 3/19 | `0x13` | REDHAWK | (11,28) |
| `game2` 38 | 3/15 | `0x13` | DR. MIKE SCOT | (23,26) |
| `game2` 40 | 3/15 | `0x17` | COVENANT | — |

**六筆一開始就有格子指過來**，其餘八筆沒有——那些人要等劇情把某一格改寫成
指向他們的 nibble 3（`docs/re/71` 的批次改寫，或 `docs/re/34` 的腳本）。
攻略裡「先開門、再解掉他身上的鎖，Covenant 才會出現」與這個形狀一致
（`docs/walkthrough/swm-005` 的 M-22）。

`0x17`／`0x27` 那幾筆同時設了 bit2 ＝ **站著不動**，
`0x13` 那幾筆會跟著移動計畫走。

`+0x00` ＝ 0 是空槽記號（`docs/re/37` §2.1）——掃描要先測它，
否則 section 的最後一筆會讀到字串表的位元組並且看起來像一筆合法記錄。

推論等級：**已確認**（名字直接從 section 17 的角色記錄讀出，
與 `docs/re/110` §5.1 是同一批人）。

## 5. remake 這一側

`internal/game/hire.go` 加兩支：

| 函式 | 對應 |
|---|---|
| `Friendly(rec)` | bit1 |
| `TurnHostile(rec)` | `sub_15C19`（清 bit1、設 bit0）|

接在戰鬥回合裡：

- `enemyActs`：`Friendly` 為真就整段跳過——**友善的那一組一輪都不會出手**。
- `partyActs`：找到目標之後、命中判定之前呼叫 `TurnHostile`。

`EncRecord` 是地圖區塊 `Raw` 的切片，所以翻臉直接寫進區塊，
這一場戰鬥結束之後仍然成立。

⚠ **重製版目前不把地圖區塊寫回資料檔**（只寫存檔那一段，`docs/re/49`），
所以翻臉在**重開遊戲之後會復原**。原版的地圖改寫是持久的
（`docs/re/71` 的批次改寫、寶箱第一次踩到就寫回記錄）——
這是一整類還沒接的東西，不是雇用專屬的，記在這裡當指標。

## 6. 順帶解出來的：bit0 那條路是「敵人的專屬名字」

`0x129E9` 讀 `+0x09` 的 bit0；**清的**印種類名（字串表 1 的 `0x52 + Kind`，
`docs/re/86`），**設的**走 `loc_12A04`：

```
0x12A04  ds:4680h ← 記錄區標頭 +0x02/+0x03      ; ← 明文名字表的起點（`docs/re/09` §3）
0x12A18  bl ← ds:A5B1h[組]（03／05／07）
0x12A1C  dl ← 遭遇記錄[bl]                       ; ＝ 這一組的敵人型別
0x12A24  dl ＝ 0 → 直接印第 0 條
loc_12A2B: 掃到 NUL；dl−−；不為 0 就再掃一條      ; **跳過 dl 條**
0x12A3D  jmp loc_17877                           ; 印
```

所以 **名字編號 ＝ 那一組的敵人型別編號**，而名字表的第 0 條是空的
（前導 NUL），等於 1 起算——與敵人資料表同一套編號。實測：

| 圖 | 型別 | 名字 | 那一格是誰 |
|---:|---:|---|---|
| `game1` 10 | 7 | `Juv\nenile\nies\n` → `Juveniles` | JACKIE |
| `game1` 32 | 6 | `Woman` | CHRISTINA |
| `game1` 34 | 4 | `City Slicker` | MORT |

出貨資料裡 `+0x09` 幾乎都設著 bit0，所以**原版絕大多數的遭遇印的是專屬名字**，
不是種類名。remake 目前一律印種類名（`Names.Name(Kind)`），所以畫面上
到處都是「人形生物」。

**接上去了**：`Block.MonsterNames()` 讀那張表，
`CombatScene.properName` 在 bit0 設著時取 `BlockNames[Type]`，
`enemyLabel`／`zhEnemy`／雇用清單都走它。⚠ 索引用的是**遭遇記錄裡的型別**
（`Enemy.Type`），不是 `Data.Kind`——後者是 1–5 的大分類，一張地圖上會重複。

中文化走與地點招牌（`place:`）同一套顯示時查表：
`tools/extract_monster_names.py` 抽出 **328 條**去重的名字進
`translations/source/monsters.tsv`，譯文在 `translations/zh-Hant/monsters.tsv`
（`docs/spec/11` §3、§5）。名字表的筆數比一開始估的 124 條多——
那個數字是只掃前 2 KB 的結果。

## 7. 可重跑的完整指令

```bash
docker run --rm --network none --log-opt max-size=10m --log-opt max-file=3 \
  -v "$PWD":/w -w /w -u "$(id -u):$(id -g)" python:3.12-slim \
  python3 tools/scan_addr_refs.py workplace/analysis/dumps/listing.json 0x46C6 0x46C8

tools/go.sh test ./internal/play/ -run TestHireableEncountersInShippedData -v
tools/go.sh test ./internal/play/ -run TestJackieEncounterIsHireable -v
tools/go.sh run ./cmd/wl-play -script "map=10:28:2,fight,H,1" -trace
```

## 8. 這一輪學到的（寫成規則）

- **「掃過了，沒有」要連掃描範圍一起寫進結論。** 上一輪的結論是
  「出貨資料裡找不到設起 bit1 的靜態遭遇記錄」，而它真正成立的範圍是
  **section 15**。範圍寫在句子裡，下一個人第一眼就看得出該去哪裡找。
- **同一種東西有兩個來源時，先問「另一個在哪」。** 這個遊戲的遭遇分
  出貨的與執行期生成的，兩者記錄格式一樣、讀取路徑一樣、
  連長度都一樣——唯一的差別是 section 型別。
- **不要用「長度剛好 N」當記錄的篩選條件。** 遭遇記錄多數 12 bytes，
  但資料裡也有 11 bytes 的，而**每個 section 的最後一筆量不到長度**
  （沒有下一個指標）。用等號會安靜地漏掉六個 NPC。
  要篩就用原版自己的哨兵（`+0x00` ＝ 0 是空槽）。
- **旗標的名字要照消費端取，不要照第一個發現它的用途取。** 這個位元
  先在 `Hire` 被讀到，於是被記成「可雇用」；六個消費端讀完之後
  它其實是「不敵對」，而可雇用只是其中一個後果。
