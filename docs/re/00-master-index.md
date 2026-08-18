# 00：RE 總表

> **這份是逆向結果的速查表。** 要找「某個位址是什麼」「某個格式怎麼解」
> 「某個機制在哪份筆記」時查這裡，不要翻十八份文件。
>
> 四份 `00-*` 各有分工：
> **本表**＝已知的事實速查；
> [`00-remake-knowledge-gaps.md`](00-remake-knowledge-gaps.md)＝還缺什麼；
> [`00-function-index.md`](00-function-index.md)＝641 個函式誰解過；
> [`00-wiring-status.md`](00-wiring-status.md)＝**已經解出來的，接上了沒有**。
>
> ⚠ **這份表寫的是「原版怎麼做」，不保證 remake 照著做了。**
> 要回答「這件事接上了沒」一律查接線表——它由 `TestWiringStatus` 守著，
> 雙向都會紅；本表沒有那道檢查（`CLAUDE.md` §0 的 G4）。
>
> 最後更新：2026-08-15

---

## 1. 位址換算（先讀這段，不然所有位址都會對不上）

分析基準是 `wl.merged.exe`（解包映像 ＋ `wla.bin` overlay，本專案合成），
SHA-256 `cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

| 段 | 類別 | 線性範圍 | 用途 |
|---|---|---|---|
| `seg000` | CODE | `0x10000`–`0x1CB67` | 主程式。前 `0x10B6` 是 `wla.bin` overlay 覆蓋區 |
| `seg001` | CODE | `0x1CB67`–`0x1CE20` | 常駐 ISR（計時器、PC 喇叭），697 bytes |
| `seg002` | DATA | `0x1CE20`–`0x2AE20` | **DS 基底**。低位 `0x0000`–`0x45xx` 是 MSQ 區塊緩衝區 |
| `seg003` | — | `0x2AE20`–`0x39000` | 素材段（字型、圖、轉換表）。段值存在 `ds:9165h` |
| `seg004` | STACK | `0x39000`–`0x39200` | |
| `seg005` | — | `0x39200`–`0x39560` | |

```
線性位址 = 0x1CE20 + <ds 位移>            ← 一般變數
線性位址 = 0x2AE20 + <seg003 位移>        ← 字型與素材（繪圖碼會先 mov ds, ds:9165h）
檔案位移 = 線性位址 − 0x10000 + 176       ← 176 = MZ header 的 e_cparhdr × 16
```

**⚠ 兩個常踩的坑**

- 繪圖與字型程式碼會先 `mov ds, ds:9165h` 換到 `seg003`，之後的 `ds:XXXX`
  要用 `0x2AE20` 當基底。用 `seg002` 算會解出雜訊，而且不會報錯（`docs/re/14` §2）。
- 反組譯文字顯示 `segment:offset`，符號名用線性位址。**不要 grep `.asm` 找位址**。

## 2. 資料檔一覽

| 檔案 | 格式 | 狀態 | 文件 |
|---|---|---|---|
| `wl.exe` | MZ 16-bit，Microsoft EXEPACK | 已解包 | [`02`](02-exepack-unpack.md) |
| `wla.bin` | 26 個 slot 的程式碼 overlay，載到 `CS:0000` | 已解 | [`04`](04-overlay-wla-bin.md) |
| `game1`／`game2` | 42 個 MSQ 區塊（XOR ＋ Huffman ＋ 打包文字） | 已解 | [`06`](06-resource-directory.md)–[`09`](09-msq-map-structure.md)、[`16`](16-msq-block-layout.md)、[`18`](18-block-text.md) |
| `colorf.fnt` | 172 字 × 32 bytes，8×8、EGA 4 平面 | 已解 | [`14`](14-fonts-and-text-encoding.md) §3 |
| `title.pic` | XOR 自參考串流 `out[n]=in[n]^out[n−0x90]` | 已解 | [`03`](03-boot-and-asset-loading.md) §6 |
| `allpics1/2` | 82 張 96 × 84 packed 4bpp 圖，與參數區交錯 | 圖已解，參數區未解 | [`23`](23-picture-format.md) |
| `allhtds1/2` | 9 組圖磚，每組 66–163 張 16 × 16 packed 4bpp | 已解 | [`24`](24-map-layers-and-tiles.md) §3 |
| `end.cpa` | Huffman 容器 | 容器已解 | [`11`](11-huffman-decoder.md) |
| `masks.wlf` | 10 張 16 × 16 **1-bit 遮罩**（32 bytes 一張），與 `ic0_9.wlf` 一一對應 | 已解 | [`47`](47-dosbox-oracle.md) §6.4 |
| `curs`／`ic0_9.wlf`／`transtbl` | 未加密，載入位址已知；`ic0_9` 的第 7 張 ＝ 隊伍圖示 | 部分已解 | [`03`](03-boot-and-asset-loading.md) §5、[`47`](47-dosbox-oracle.md) |
| `info` | 2 bytes 安裝資訊 | 已解 | [`03`](03-boot-and-asset-loading.md) §3 |
| `paragraphs.txt`／`manual.txt` | 明文 | 已轉錄 | `docs/paragraphs/`、`docs/manual/` |

**主文字字型不是獨立檔案**，內嵌在 `wl.exe` 的 `seg003:0xCA60`。

## 3. 演算法速查

| 機制 | 一句話 | 實作 | 文件 |
|---|---|---|---|
| EXEPACK 解包 | 反向 RLE ＋ relocation 重建 | `tools/unpack_exepack.py` | [`02`](02-exepack-unpack.md) |
| MSQ 解密 | `key = lo(cs) ^ hi(cs)`；逐 byte XOR 後 `key += 0x1F` | `tools/decrypt_msq.py` | [`08`](08-msq-encryption.md) |
| **checksum** | **＝ 0 − Σ 明文位元組**（改寫內容後必須重算） | `tools/dump_save.py` | [`08`](08-msq-encryption.md) §0、[`30`](30-save-layout.md) §2.1 |
| **加密長度** | **＝ 區塊標頭第一個 word，不是整個區塊** | `tools/decode_block_text.py` | [`18`](18-block-text.md) §2 |
| Huffman | 前序編碼的樹 ＋ 位元流，無 magic 的尾段也是同一套 | `tools/huffman.py` | [`10`](10-huffman-compression.md)、[`11`](11-huffman-decoder.md) |
| 文字打包 | 5-bit 符號 ＋ 60 bytes 字元對照表；`0x1E` 轉大寫、`0x1F` escape | `tools/decode_text.py` | [`17`](17-packed-text.md) |
| 亂數 | 五個 byte 的進位鏈，無乘除，初值全零，熵來自鍵盤輪詢次數 | `tools/rng.py` | [`13`](13-rng.md) |
| 擲骰 | dN ＝ 遮罩 ＋ 拒絕取樣，回傳 1..N 等機率 | `tools/rng.py` | [`13`](13-rng.md) §3 |
| 字型繪製 | 主文字 8 bytes 單色；彩色字型 32 bytes、4 平面連續存放 | `tools/dump_font.py` | [`14`](14-fonts-and-text-encoding.md) |
| **圖片** | packed 4bpp ＋ 列間 XOR delta；**回看距離 ＝ 列的 byte 數** | `tools/decode_pic.py` | [`23`](23-picture-format.md) |
| **圖磚** | 同上，一張 128 bytes ＝ 16 × 16；載入時轉成 EGA 4 平面 | `tools/decode_pic.py tile` | [`24`](24-map-layers-and-tiles.md) §3 |
| 走一步 | 可否進入 → 捲動 ＋ 補兩排 → 遭遇擲骰 → 依 nibble 分派事件 | — | [`26`](26-movement-and-triggers.md) |
| **遊戲時鐘** | 24 小時制；每步 ＋ 標頭 `+0x34/35`（分 × 256）。晝夜門檻 6／18 時 | — | [`27`](27-game-clock.md) |
| **升級門檻** | **升到等級 L 要累計經驗值 (L² − L) × 512**；升級不扣經驗值，每級 ＋1 技能點 | — | [`31`](31-experience-and-skills.md) §1 |
| **技能費用** | IQ 需求 ＝ 技能資料 `+0x00 >> 3`；升到等級 L 的費用 ＝ (`+0x00 & 7`) × 2^(L−1)，飽和 `0xFF` | — | [`31`](31-experience-and-skills.md) §3 |
| **技能／屬性檢定** | 2d6 續擲（< 5 直接失敗）＋ 屬性 ＋ 技能等級×3 **≥ 難度 × 5 ＋ 15** | — | [`32`](32-skill-checks-and-xp.md) §3 |
| **擊殺經驗值** | 攻擊資料 `+0x00/+0x01`（16-bit） × ((`+0x04` 低 4 位) ＋ 1) | — | [`32`](32-skill-checks-and-xp.md) §8 |

## 4. 資料結構

### 4.1 MSQ 區塊（`game1`／`game2` 的一個資源）

```
+0x0000                地圖第 1 層：一個 byte 兩格（4 bits）
+D²÷2                  地圖第 2 層：一格 1 byte
+P（＝ D²×1.5）         記錄區標頭（0x5C bytes），第一個 word ＝ 加密長度 L
+P+0x5C                各 section
+L                     字串表（不加密）
+（讀取量 − 6）         Huffman 尾段 ＝ 地圖第 3 層，D² bytes（載到 ds:3448h）
+（總長度 − 6）         結束
```

| P | 邊長 D | 區塊數 | 判定 |
|---|---:|---:|---|
| `0x600` | 32 | 38 | `ds:BF1Ch[資源編號] != 0x40` |
| `0x1800` | 64 | 4 | `ds:BF1Ch[資源編號] == 0x40` |

**記錄區標頭（P 起算）**

| 位移 | 內容 |
|---|---|
| `+0x00` | 加密長度 L ／ 同時是字串表基址 |
| `+0x02` | 明文 NUL 分隔的敵人名表位址 |
| `+0x04` | 一張 8 bytes 一筆的記錄表位址 |
| `+0x06`…`+0x28` | section 位移表（型別 → 位移要查執行檔的 `ds:B9E0h`） |
| `+0x2C` | **地圖邊長 D**（32 或 64） |
| `+0x33` | **地圖範圍外要畫的圖形編號** |
| `+0x34`–`+0x35` | **走一步推進的時間**（16-bit ＝ 分鐘 × 256） |
| `+0x36` | 走一步推進的刻（`ds:4722h`／32-bit 累計用；多數地圖 ＝ `+0x34` ÷ 64） |
| `+0x2F` | 遭遇出現機率的分母 |
| `+0x30` | **圖磚組編號**（0–8） |
| `+0x31` | 遭遇種類數 |
| `+0x32` | 遭遇槽位上限 |

**section**（型別 3、5 已驗證）：`[16-bit 指標陣列][記錄本體]`，
陣列長度不另存——**第一個非空指標指到哪，陣列就到哪為止**。

### 4.2 地圖記錄（`ds:46AEh` 指到的那筆）

**⚠ 欄位語意隨 section 型別而不同**，不要當成單一結構。

| 位移 | section 15（遭遇槽） | section 3 |
|---|---|---|
| `+0x00`／`+0x01` | 座標 | — |
| `+0x03` | 種類編號（`1..標頭+0x31`，實測最大 5） | 訊息編號（98% 落在該區塊字串數內，`0xFF` 是哨兵） |
| `+0x04` | 擲骰結果 | — |
| `+0x05`／`+0x07` | 與 `+0x03` 一起做「此槽是否為空」判斷 | — |
| `+0x08`–`+0x0A` | bit7 ＝ 續接，低 7 位是值（變長串列） | 同左 |

### 4.2b 存檔（`game1`／`game2` 檔尾的一個 MSQ 資源）

```
game1 0x000253C5、game2 0x00028BC7   ← seek 是 cx:dx，32-bit
+0x00  magic  +0x04 checksum  +0x06  0x800 加密段 ＝ 8 × 256
         第 0 筆：全域狀態    第 1–7 筆：角色記錄
+0x806 0xA00 未加密段 ＝ **十組按鍵巨集**（10 × 256，出廠全零）
```

全域狀態：`+0x00/0x0E/0x1C/0x2A` 四組隊伍槽表（各 14 bytes）、
`+0x78` 起 14 bytes 是 `ds:464Eh`–`465Bh`（視窗原點、累計、隊伍人數、時鐘）、
`+0xC8` 地點名稱、`+0xF5` 32-bit 存檔序號。詳見 [`30`](30-save-layout.md)。

### 4.3 角色記錄（256 bytes）

定址：`0x7131 + 角色編號 × 256`，經隊伍槽表兩層間接
（`sub_17208` → `sub_19614`）。詳見 [`15`](15-character-record.md)。

| 位移 | 大小 | 語意 | 等級 |
|---|---|---|---|
| `+0x00` | 字串 | 名字（NUL 結尾） | 已確認 |
| `+0x15`–`+0x17` | 24-bit | **金錢**（商店比價與扣款，`docs/re/22` §5） | 已確認 |
| `+0x21`–`+0x23` | 24-bit | **經驗值**（累計，升級不扣；門檻見 §3） | 已確認 |
| `+0x26`–`+0x27` | 16-bit | 受傷前的 CON 備份 | 強證據 |
| `+0x18` | byte | 性別（0 ＝ Male、1 ＝ Female） | 已確認 |
| `+0x19` | byte | 國籍（0–4 ＝ U.S./Russian/Mexican/Indian/Chinese） | 已確認 |
| `+0x0E`–`+0x14` | 7 bytes | **七個屬性**：Strength／IQ／Luck／Speed／Agility／Dexterity／Charisma | 已確認 |
| `+0x20` | byte | 可用技能點（建立時 ＝ IQ；每升一級 ＋1，飽和 255） | 已確認 |
| `+0x24` | byte | **等級**（建立時 1，升級時寫入新等級） | 已確認 |
| `+0x32`–`+0x4A` | 字串 | 階級名（建立時 ＝ `PRIVATE`）。**寫入端 `0x1BB6C` 抄到 NUL 為止、沒有長度檢查**，欄位大小是被 `+0x4B` 擋出來的；最長的階級 `Lieutenant Commander` 寫到 `+0x46`（[`109`](109-character-record-tail.md) §4）| 已確認 |
| `+0x4B`／`+0x4C` | 2 bytes | 參與過摧毀 Base Cochise ／ 總部已經表揚過（各只用 bit0，[`96`](96-ending.md) §5）| 已確認 |
| `+0x4D`–`+0x7F` | 51 bytes | **沒有任何存取點**、出廠全零（[`109`](109-character-record-tail.md)）| 已確認 |
| `+0x1A` | byte | **AC**（護甲等級）＝ 裝備護甲的物品表 `+0x06`（`sub_1949E`，[`45`](45-item-data-and-weapon-damage.md) §3.4） | 已確認 |
| `+0x1B`–`+0x1C` | 16-bit | MAXCON（最大體力） | 強證據 |
| `+0x1D`–`+0x1E` | 16-bit 有號 | CON（目前體力，可為負） | 強證據 |
| `+0x1F` | byte | **裝備武器的物品槽**（`sub_19AC8`：`× 2 + 0xBB` 指進物品陣列） | 已確認 |
| `+0x25` | byte | **裝備護甲的物品槽**（同上；`sub_1949E` 用這兩個判斷「再選一次 ＝ 卸下」） | 已確認 |
| `+0x28` | byte | **八個狀態位元**（疾病；高四位隨時間惡化，`docs/re/35`） | 已確認 |
| `+0x80`–`+0xBB` | 30 × 2 | **技能**（偶數位移 ＝ 編號，上限 36；奇數位移 ＝ 等級） | 已確認 |
| `+0xBD`–`+0xF8` | 30 × 2 | **物品**（奇數位移 ＝ 編號；偶數位移 ＝ 附屬 byte） | 已確認 |

分辨依據：`sub_1997B` 印 `LVL SKILL` 表頭並走 `+0x80`、帶 `cmp al, 24h`（技能名剛好 35 個）；
`sub_19454` 印 `ITEM` 表頭並走 `+0xBD`。`+0x1F` 的裝備索引指進物品陣列。

傷勢等級：CON 越負等級越高，門檻 `−11 / −20 / −30 / −40`（`ds:CCCEh`），
對應字串 `UNC / SER / CRT / MRT / COM`，CON ＝ 0 走骷髏字模。

### 4.4 效果記錄（地圖記錄的 `+0x08`／`+0x09`）

地圖事件對角色做的所有事都由這兩個 byte 描述（`sub_14193` → `sub_141FA`）：

| byte | bit 7 | bit 0–6 |
|---|---|---|
| `+0x08` | 0 ＝ 值是 N 顆 d6；1 ＝ 固定值 | 要改的**角色記錄欄位位移** |
| `+0x09` | 0 ＝ 加；1 ＝ 減 | 固定值，或骰子數 N |

特判：欄位 `0x1D` 走傷害流程（含護甲）、`0x15`／`0x21` 走 24-bit 進位鏈，
其餘一律單 byte 加減（減到負的夾在 0）。詳見 [`19`](19-effects-and-damage.md)。

**傷害與護甲**：護甲吸收 ＝ **AC（`+0x1A`）顆 d6 的和**；
實際傷害 ＝ 傷害 − 吸收，再從 16-bit 有號的 CON 扣掉。

### 4.5 屬性 → 修正值

```
v ≤ 8       修正 ＝ floor((v − 9) / 2)   （負）
9 ≤ v ≤ 12  修正 ＝ 0
v ≥ 13      修正 ＝ (v − 12) >> 1        （正；13 也是 0，所以死區實際是 9–13）
```

入口：`sub_15705`＝Luck、`loc_1570A`＝Dexterity、`loc_1570F`＝Agility、
`loc_15714`＝Strength，四支扇入 `loc_15716`。詳見 [`21`](21-attributes.md)。

### 4.6 角色建立（`sub_1C6C9`）

```
整筆清零                rec[0..255] ← 0
性別                    roll(1..2)
七個屬性 +0x0E–+0x14    每格 ＝ 5d6 取最高三顆（3–18，期望 13.43）
MAXCON ＝ CON           5d6 取最高三顆 ＋ 18（21–36）
rec[+0x20] ← IQ         初始技能點
rec[+0x24] ← 1
rec[+0x32…] ← "PRIVATE" 初始階級字串
```

### 4.7 物品資料表（`ds:7A31h`，執行期載入 `0x2F8` bytes）

**95 筆 × 8 bytes**，索引從 1 開始（`sub_17AE0` 的基址是 `0x7A39` ＝ 表首 ＋ 8）。

⚠ **表在存檔區，不在執行檔裡**，每個存檔槽各一份 —— 它是**可變的遊戲狀態**
（[`45`](45-item-data-and-weapon-damage.md) §2）：

```
偏移 = 0x000253C5 ＋ 0x1206 ＋ ds:BE20h[槽]   （game1，槽 = 0 / 0x2FE / 0x5FC）
       0x00028BC7 ＋ 0x1206                   （game2）
```

| 位移 | 內容 |
|---|---|
| `+0x00`–`+0x01` | 基礎價（16-bit） |
| `+0x02` | 這家店的庫存量（0 ＝ 缺貨、`0xFF` ＝ 無限；賣一件 ＋1） |
| `+0x03` | `>> 3` ＝ **物品類別**（1 近戰／2–13 有射程的武器／14 彈藥／15 護甲／16 一般／17 雜物／18 劇情） |
| `+0x04` | **容量**（裝滿時的次數；槍是彈匣容量）——發裝備／換彈時寫進物品槽的附屬 byte |
| `+0x05` | **使用技能的編號**（技能表的 36 條） |
| `+0x06` | **骰數**：武器 ＝ 傷害 Nd6、護甲 ＝ AC |
| `+0x07` | **要用的彈藥物品編號**（0 ＝ 不用） |

**價格 ＝ 基礎價 − (基礎價 >> n)**，`n` 來自商店那筆地圖記錄的 `+0x03`
（0 ＝ 原價、1 ＝ 半價、2 ＝ 75%…）。詳見 [`22`](22-shop-and-items.md)、
[`45`](45-item-data-and-weapon-damage.md)。

### 4.7b 敵人資料表（記錄區標頭 `+0x04` 指到，8 bytes 一筆）

**和物品表不同的另一張表**——兩張共用同一支定址器（`sub_16F4C`），
差別在誰設基址（[`32`](32-skill-checks-and-xp.md) §1）。索引 ＝ 地圖記錄裡的敵人型別，
第 0 筆恆為全零（＝ 沒有敵人）。

| 位移 | 內容 |
|---|---|
| `+0x00`–`+0x01` | **基礎血量，同時也是經驗值基值**（16-bit，[`37`](37-enemy-records-and-hp.md) §3） |
| `+0x02` | 行動值欄位（× 8 進行動值；也被命中門檻減去） |
| `+0x03` | 傷害骰數 N |
| `+0x04` 低 4 位 | 經驗值倍數 −1 |
| `+0x04` 高 4 位 | **沒有讀者**（資料裡有值 0–10，程式沒讀） |
| `+0x05` 低 4 位 | **武器類別**，與物品表同一套編碼（`ds:CD00h` 那張清單） |
| `+0x05` 高 4 位 | 傷害基底 |
| `+0x06` | **敵人種類**：1 Animal／2 Mutant／3 Humanoid／4 Cyborg／5 Robot（397 筆全在 1–5） |
| `+0x07` | **肖像圖編號**（`ALLPICS`）：遭遇時 `sub_190A8` → `sub_184E8` 載那張圖；同一個編號查 `ds:A920h`（79 筆，0／1／2）得到文字碼 `0x0E` 的 him／her／it |

### 4.7d 敵方記錄（`ds:46C8h` 指到，94 bytes 一筆，16 筆在 `0x6B31`）

| 位移 | 內容 |
|---|---|
| `+0x00`／`+0x01` | 遭遇所在的地圖座標 x／y |
| `+0x02` | 資源（地圖）編號 |
| `+0x03` | 與隊伍的距離（10 × 歐氏，查 `ds:CD0Dh`） |
| `+0x04` 起 | **3 組 × 30 bytes**：每組 `+0x00`–`+0x13` 是 10 個 16-bit 血量、`+0x14`–`+0x1D` 是「還沒行動」旗標 |

血量 ＝ `⌊基礎血量 / 4⌋ ＋ 1d(基礎低位) ＋ 256 × 1d(基礎高位)`，
每隻各擲各的（[`37`](37-enemy-records-and-hp.md) §3）。

### 4.7c 技能資料表（`ds:BA20h`，36 筆 × 2 bytes）

| 位移 | 內容 |
|---|---|
| `+0x00` | `>> 3` ＝ 最低 IQ；`& 7` ＝ 基礎技能點費用 |
| `+0x01` | 檢定用的屬性——**值就是角色記錄的位移**（`0x0E`–`0x14`） |

35 條的數值表在 [`32`](32-skill-checks-and-xp.md) §2。

### 4.8 戰鬥判定

兩條攻擊路徑共用 `sub_1B108` 累加到 `ds:46C0h`，**累加的永遠是隊伍成員的本事**，
判定方向隨攻守翻轉（判定前綴機器碼相同，只差 `jb`／`jnb`）：

| 路徑 | 位址 | 基礎值 | 命中條件 |
|---|---|---|---|
| 隊伍打敵方 | `0x1AF52` | 目標這回合的移動計畫（`ds:711Dh`）：不動 50／會動 60 | `roll(1..100) < 累加值` |
| 敵方打隊伍 | `0x1B04C` | 被打者這回合的指令：迴避 60／攻擊 50／其餘 40 | `roll(1..100) ≥ 累加值` |

累加值 ＝ 基礎 ＋ **Brawling（技能 1，寫死）** × 3 ＋ **Agility**
− 對手行動值（近戰類別先做 **8-bit** ×4，否則累加值另加 5），夾在 100
（[`88`](88-hit-accumulator.md)）。

```
傷害     (武器資料 +0x05 高 4 位) ＋ (武器資料 +0x03) 顆 d6
護甲     吸收 ＝ N 顆 d6 的和（角色 N ＝ AC；敵方 N 來自 loc_12A92）
結算     角色：CON（記錄 +0x1D，16-bit 有號）−= 傷害，**可為負**
         敵方：HP（ds:46C8h + 編號×2，16-bit）−= 傷害，**≤0 夾成 0**
```

詳見 [`20`](20-combat-resolution.md)、[`19`](19-effects-and-damage.md)。

### 4.9 字串表

```
表基址 +0x00 … +0x3B    60 bytes 字元對照表（符號 → ASCII，0 ＝ 結束）
表基址 +0x3C …          16-bit 位移表，每 4 個字串一項，位移相對於 +0x3C
```

取第 N 個字串：跳到第 `N >> 2` 項位移，再解掉 `N & 3` 個。
**每張表有自己的字元對照表**，不能共用。

| 來源 | 張數 | 非空字串數 |
|---|---:|---:|
| 執行檔（常數基址，見 §5.1） | 9 | 442 |
| 42 個 MSQ 區塊（各一張） | 42 | 4,401 |
| **合計** | **51** | **4,827** |

字串內的機制（**同一個骨架的四個實例**，`docs/re/28`）：
`0x0A` 單複數二選一（選擇子 ＝ 數量）、`0x0C` 性別二選一（角色記錄 `+0x18`）、
`0x0E` him／her／it 三選一（目標類別）、`0x0F` 印出數量；
另有 `0x0B` 插入角色名字、`0x0D` 段內換行。三個變形碼可互相巢狀。

### 4.9b 畫面版面（320 × 200，mode 0Dh）

```
欄 0…37、字元列 0…17     外框（彩色字型的框線字元）
(8, 8) 起 288 × 128       地圖／圖片／隊伍名單共用（19 × 9 格，四邊各半格）
欄 1…38、字元列 18…23     訊息視窗（6 行 × 38 個 8×8 字元）

戰鬥（`ds:46B9h` ＝ 1）——訊息視窗那 6 列是空的，畫面上半整個換掉：
欄 1…13、字元列 1…12      肖像框：圖 ＋ 說明（列 12，12 格置中）——
                          最近的那一組敵人，挑不到就是遊俠（圖 8、`Ranger`）
欄 15…38、字元列 1…13     指令選單與戰鬥訊息（`sub_19727` 設的四個邊界）
字元列 14…23              名單（表頭一列 ＋ 每人一行，**一行 39 欄**、行首序號 ＋ `>`）
字元列 24                 指令列（戰鬥中照留）
```

單位：`ds:4674h`／`4675h` ＝ byte 欄，`ds:4676h`／`4677h` ＝ 像素列，
`ds:4672h`／`4673h` ＝ 游標的欄／**字元列**，列位址表在 `ds:8DF9h`。
`ds:46B9h` ＝ 0 地圖、1 隊伍名單。詳見 [`25`](25-screen-layout.md)、
戰鬥那幾塊見 [`105`](105-enc-empty-round-and-menu-region.md) §2、[`103`](103-roster-line-columns.md)
與 [`115`](115-portrait-box.md)（肖像框）。

### 4.10 地圖（三層，正方形，邊長 D 在記錄區標頭 `+0x2C`）

```
偏移 0        第 1 層  4 bits／格   D²÷2 bytes   這一格屬於哪一種 section
偏移 D²÷2     第 2 層  1 byte／格   D²   bytes   該 section 裡的第幾筆記錄
Huffman 尾段  第 3 層  1 byte／格   D²   bytes   畫面上的圖形編號，載到 ds:3448h
偏移 D²×1.5   記錄區標頭 0x5C bytes（見 §4.1）
```

D 只有 32（38 個地圖）與 64（4 個地圖）兩種。第 1 層取值 `sub_17C20`
（偶數行取高 4 位）、第 2 層 `sub_17C72`、第 3 層 `sub_17FC8`。
圖磚組編號在記錄區標頭 `+0x30`（0–8，0–3 在 `ALLHTDS1`、4–8 在 `ALLHTDS2`）。

**第 3 層的圖形編號**指進 `seg003:0x420` 起那張連續的 128 bytes／筆的表：

```
0–9    IC0_9.WLF 的十個疊圖
≥10    圖磚，編號 ＝ 值 − 10          （0x420 ＋ 10 × 128 ＝ 0x920 ＝ 圖磚組起點）
```

畫一格：`螢幕 ← (背景[0x420+值×128] AND 遮罩[0xDA60+疊圖×32]) OR 疊圖[0x420+疊圖×128]`
（overlay slot 4 ＝ `sub_1029B`）。螢幕座標 `ds:4685h` ＝ byte 欄（地圖行 ×2）、
`ds:4686h` ＝ 像素列（地圖列 ×16）。

### 4.11 圖片

```
解碼（word 為單位，n 由 0 起每次 +2）：
    out[n + stride] ^= out[n]        ← 讀到的是「已經解過的」內容，順序不能顛倒
畫素：packed 4bpp，一個 byte 兩個像素，高 4 位在左
```

| 來源 | 大小 | stride | 尺寸 | 張數 |
|---|---:|---:|---|---:|
| `ALLPICS1` | 4,032 | 48 | 96 × 84 | 33 |
| `ALLPICS2` | 4,032 | 48 | 96 × 84 | 49 |
| `TITLE.PIC` | 18,432 | 144 | 288 × 128 | 1 |
| `END.CPA` | 18,432 | 144（推測） | 288 × 128（推測） | 1 |

`ALLPICS` 的子區塊嚴格交錯：一張圖 ＋ 一段變動長度的參數區（430–2,490 bytes，未解）。

`ALLHTDS` 的圖磚是同一套格式的最小單位：一張 128 bytes、stride 8 → **16 × 16**，
每張各自帶 8 bytes 的種子列（delta 不跨圖磚）。

## 5. 位址表

### 5.1 執行檔內的表（`ds:` 位移，線性 ＝ ＋`0x1CE20`）

| `ds:` | 線性 | 內容 |
|---|---|---|
| `0xA703` | `0x27523` | 字串表：開場字幕與製作名單 |
| `0xAA60`／`0xAA6D`／`0xAA7A` | `0x27880`… | 遭遇生成用的三張 13 項表 |
| `0xAA87` | `0x278A7` | **地圖 nibble → 事件處理函式**（16 項，`docs/re/26` §5、`29`） |
| `0xAA17` | `0x27837` | **敵人種類 → 疊圖編號**（6 項 `00 06 03 04 02 01`，`docs/re/48` §3） |
| `0xA4E0` | `0x27300` | nibble 6 的第二層跳表 A（5 項，設施畫面：商店已確認，`docs/re/29` §5.4） |
| `0xA4EA` | `0x2730A` | nibble 6 的**腳本指令表**（44 個 opcode，末尾 word ＝ 0，`docs/re/34`） |
| `0xAAA7` | `0x278C7` | 方向 → 捲動函式（4 項：上／下／左／右） |
| `0xAAB1` | `0x278D1` | 方向 → 座標增減（4 項） |
| `0xCAEB` | `0x2990B` | **滑鼠熱區表**（21 筆 × 5 word：x1／y1／x2／y2／handler，`docs/re/43`） |
| `0xC05D` | `0x28E7D` | **地圖四個楔形送出去的鍵**（4 bytes ＝ `I K J L`，`docs/re/112` §4） |
| `0x7E0B` | `0x24C2B` | `CURS` 載入位址；游標索引在 `ds:8DCDh`，繪製在 `0x10D4D`（`docs/re/112`） |
| `0xCD00` | `0x29B20` | **有射程的武器類別清單**（`0d 0a 0b 0c 02…09 ff` ＝ 類別 2–13，`docs/re/45` §3.1） |
| `0xCD0D` | `0x29B2D` | **距離表** 5 列 × 10 行（`sub_19D4D` 查它，`docs/re/37` §4） |
| `0xA5AE`／`0xA5B1` | `0x273CE` | 三組敵人的**數量位移**（`04 06 08`）與**型別位移**（`03 05 07`，`docs/re/37` §3） |
| `0xCBBD` | `0x299DD` | 固定按鈕的鍵（17 筆 × 6 bytes，鍵在 `+0x03`） |
| `0xC05D` | `0x28E7D` | 地圖視窗四象限的方向鍵（`49 4B 4A 4C` ＝ `I K J L`） |
| `0xA44D` | `0x2726D` | 戰鬥指令熱鍵表（`" HAWRELU"`，`docs/re/38` §2） |
| `0xAB3E` | `0x2795E` | 字串表：無線電、隊伍、戰鬥 |
| `0xB233` | `0x28053` | 傷勢等級 → 訊息碼（`85 9A 9B 9C 9D 84`） |
| `0xB239` | `0x28059` | 隊伍槽表位移（`00 0E 1C 2A`，間隔 14） |
| `0xB270` | `0x28090` | 字串表：技能、物品、介面（170 條） |
| `0xB9E0` | `0x28800` | **section 型別 → 記錄區標頭位移**（24 項） |
| `0xBA20` | `0x28840` | **技能資料表**（36 筆 × 2 bytes：`+0` 高 5 位 IQ／低 3 位費用、`+1` 檢定屬性的記錄位移） |
| `0xBD22` | `0x28B42` | 各資源的讀取量 |
| `0xBD86` | `0x28BA6` | 各資源的區塊總長度 |
| `0xBEC9` | `0x28CE9` | 資源目錄（50 項，`0xFF` 結束；高 2 bits ＝ 哪個檔案） |
| `0xBF1C` | `0x28D3C` | 地圖尺寸選擇表（`0x40` → `0x1800`） |
| `0xCCCE` | `0x29AEE` | 傷勢門檻（`F5 EC E2 D8` ＝ −11／−20／−30／−40） |
| `0xCF4E` | `0x29D6E` | 狀態位元遮罩（`01 02 04 08 10 20 40 80`，`docs/re/35`） |
| `0xCD50` | `0x29B70` | 文字控制碼跳表（直繪版，18 項） |
| `0xCD74` | `0x29B94` | 文字控制碼跳表（組行版，18 項） |
| `0xCE4B` | `0x29C6B` | 字串表：角色建立 |
| `0xD18E` | `0x29FAE` | 字串表：**結局敘述** |
| `0xD522` | `0x2A342` | **等級 → 階級編號**（byte 表，50 階對到等級 1–131；前 11 級一級一階，之後 2／3／4 級一階） |
| `0xD622` | `0x2A442` | 字串表：階級（64 條） |
| `0xDACC` | `0x2A8EC` | 字串表：技能學習 |
| `0xDBF8` | `0x2AA18` | 字串表：商店 |
| `0xDCED` | `0x2AB0D` | 字串表：疾病與狀態 |

### 5.2 `seg003` 的素材位址（線性 ＝ ＋`0x2AE20`）

| `seg003:` | 線性 | 內容 |
|---|---|---|
| `0x0100` | `0x2AF20` | `TRANSTBL`（800 bytes ＝ **50 組 × 16 的索引對照表**；載入之後沒有人讀，[`56`](56-transtbl.md)） |
| `0x0420` | `0x2B240` | `IC0_9.WLF`：10 個 16 × 16 疊圖（128 bytes／筆，4 平面） |
| `0x0920` | `0x2B740` | `TITLE.PIC`（18,432 bytes）；進遊戲後同一塊改放**圖磚組**（4 平面，128 bytes／張） |
| `0xB4E0` | `0x36300` | `COLORF.FNT`（5,504 bytes ＝ 172 × 32） |
| `0xCA60` | `0x37880` | **主文字字型**（內嵌，128 × 8 bytes） |
| `0xDA60` | `0x38880` | `MASKS.WLF`：10 張 16 × 16 遮罩（32 bytes／筆） |

`CURS` 載到 `seg002:0x7E0B`。

### 5.3 執行期變數（`ds:` 位移）

| `ds:` | 內容 |
|---|---|
| `0x4654` | 目前隊伍組別 |
| `0x4655` | 目前資源（地圖）編號 |
| `0x4665`／`0x4666` | 共用定址器算出來的表項指標（`sub_16F4C`，`docs/re/32` §1） |
| `0x46A6`／`0x46A7` | **隊伍座標 x／y** |
| `0x465C`–`0x4660` | **RNG 狀態**（5 bytes，映像初值全零） |
| `0x4661` | 目前 section 的起點 |
| `0x4667` | 最後一個按鍵碼（小寫已轉大寫） |
| `0x4703` | `\x10` 的「下一個字元要登記成熱鍵」旗標（`docs/re/43`） |
| `0x8DDC` | **每列熱鍵表**（25 格，`\x10` 寫、滑鼠讀；`\x11` 清空） |
| `0x7DF3`／`0x7DF5` | **滑鼠熱區遮罩**（低 16 位／高 16 位，第 i 位對應第 i 塊熱區） |
| `0x7DF7` | 滑鼠按鍵狀態位元 |
| `0x7DF9`／`0x7DFB` | 滑鼠座標 x／y（`int 33h` AX=3） |
| `0x8DDB` | 有沒有滑鼠驅動程式（0 → 完全不呼叫 `int 33h`） |
| `0x4672`／`0x4673` | 文字欄位座標（欄／列） |
| `0x4675` | 行寬上限 |
| `0x4678` | 反白旗標（非 0 → `not al`） |
| `0x4680`／`0x4681` | 目前字串指標 |
| `0x4688` | 選單選中的第幾個詞 |
| `0x468B` | 該詞在字串裡的位移 |
| `0x4692`／`0x4693` | **字串表基址**（換表就是改這裡） |
| `0x46A1`／`0x46A2` | 字元對照表基址 |
| `0x46AE` | **目前地圖記錄的指標**（全檔 150 處以上以它當基址） |
| `0x46B0` | **記錄區標頭位置 P**（`0x600` 或 `0x1800`） |
| `0x46B5`／`0x46B6` | **目前角色記錄的位址** |
| `0x46B7` | 隊伍槽表位址 |
| `0x46BB`／`0x46BC`／`0x46BD` | 文字解碼：剩餘位元數／大寫旗標／位元緩衝 |
| `0x46C0`／`0x46C1` | 命中門檻累加器（`sub_19C2C` 加減） |
| `0x46BE`／`0x46BF` | **傷害／加減值**（16-bit，`sub_19BF0` 寫入） |
| `0x46EF` | **這一次結算跳過護甲吸收**（設 → 呼叫 → 清；值 ＝ 地圖記錄 `+0x00` 的 bit0） |
| `0x46F0` | 批次結算中：`sub_19EFC` 整支跳過 |
| `0xA96F` | **遭遇佇列**（4 組 × 4 槽 × 4 bytes，每次掃描重建） |
| `0x7111` ＋ 組 | 這一格對該組的距離（`0xFF` ＝ 不適用） |
| `0x711D` ＋ 記錄 | **敵方這一回合的移動計畫**（16 格，`0xFF` ＝ 不動；高 2 位訊息、低 6 位步向）。填它的是 `sub_14BF0`，用它的是 `sub_15036`（執行）與 `0x1AF64`（命中基礎值），[`101`](101-enemy-move-plan-table.md) |
| `0x7212` ＋ 組 | 該組的接戰值（0／0x0F／0xFE） |
| `0x46D8` ＋ 成員 | **這回合的戰鬥指令碼**（`' ' H A W R E L U` 的索引） |
| `0x46DA` ＋ 成員 | 該指令的參數（逃跑方向、攻擊目標…） |
| `0x46C6`／`0x46C7` | 遭遇所在地圖記錄的指標（`sub_13762` 抄進 `ds:46AEh`） |
| `0x46C8` | **敵方記錄的指標**（`sub_137F4` ＝ 標頭、`sub_137CE` ＝ 某一組的 30-byte 區塊，`docs/re/37`） |
| `0x470B` | 角色記錄 `+0x18` 的快取 |
| `0x722F` | 彩色字型選色（0 → 冷色 bank） |
| `0x7131` | 角色記錄區起點（存檔載入目標） |
| `0x8DD0`／`0x8DD1` | 彩色字型的游標欄／列 |
| `0x9165` | `seg003` 的段值 |
| `0x9168` | 磁碟旗標（`0x80` ＝ GAME1、`0x40` ＝ GAME2） |
| `0x9176` | 目前區塊的 checksum |
| `0xB265` | 每字元處理器的函式指標 |
| `0xC059` | dN 的上限暫存 |
| `0xCD3F`–`0xCD4A` | 擲骰用暫存 |

## 6. 關鍵函式

完整索引在 [`00-function-index.md`](00-function-index.md)（641 個，已分析 461）。
下表只列「解 remake 時最常回頭查」的。

| 位址 | 呼叫端 | 作用 | 文件 |
|---|---:|---|---|
| `0x110B6` | — | `start`：開機序列與七個素材的載入 | [`03`](03-boot-and-asset-loading.md) |
| `0x11445` | 11 | 開檔（走資源表 `+6` 的檔名） | [`05`](05-storage-layer.md) |
| `0x115E5` | 7 | 讀進緩衝區 | [`03`](03-boot-and-asset-loading.md) |
| `0x11A59` | 1 | **MSQ 解密 ＋ checksum 驗證** | [`08`](08-msq-encryption.md)、[`18`](18-block-text.md) |
| `0x11B83` | 7 | Huffman 解碼 | [`11`](11-huffman-decoder.md) |
| `0x1841F` | 1 | **地圖區塊載入器**（決定 P、解密、解壓） | [`16`](16-msq-block-layout.md) |
| `0x17CB1` | 18 | **取第 n 筆地圖記錄**（兩層索引 → `ds:46AEh`） | [`16`](16-msq-block-layout.md) §3 |
| `0x1393E` | 7 | 讀記錄區標頭的任一 byte | [`16`](16-msq-block-layout.md) §2 |
| `0x17208` | 28 | 算隊伍槽表位址 | [`15`](15-character-record.md) |
| `0x19614` | 43 | **算角色記錄位址**（全檔呼叫端最多） | [`15`](15-character-record.md) |
| `0x1708B` | 1 | 名片行（NAME/AC/AMM/MAX/CON/WEAPON） | [`15`](15-character-record.md) §4 |
| `0x19A1D` | 4 | CON → 傷勢等級 | [`15`](15-character-record.md) §3 |
| `0x18E6B` | 3 | **亂數產生器** | [`13`](13-rng.md) |
| `0x18E41` | 24 | dN（1..N） | [`13`](13-rng.md) §3.2 |
| `0x18E5F` | 5 | d6 | [`13`](13-rng.md) §3.1 |
| `0x157D6` | 3 | **傷害結算**：CON −= (傷害 − 吸收)；`ds:46EFh` 非 0 就跳過護甲那 N 顆 d6 | [`55`](55-radiation-and-armour-bypass.md) §1 |
| `0x141FA` | 2 | 對角色記錄的某個欄位加減（`bl` 有號負 ＝ 減）；欄位 `0x1D` 走傷害結算 | [`55`](55-radiation-and-armour-bypass.md) §3 |
| `0x14410` | — | **輻射格結算**：逐一對隊員擲 `+0x01` 顆 d6 扣 CON ＋ 加 Radiation poisoning | [`55`](55-radiation-and-armour-bypass.md) §3 |
| `0x13EC9` | 1 | **條件閘的主體**：逐個角色跑條件串列，收尾由記錄 `+0x00` 低位的四個旗標決定 | [`69`](69-gate-flags.md) §2 |
| `0x16890` | — | **遭遇生成器**：擲 1／標頭 `+0x2F` → 找 section 15 空槽 → 擲種類 → 沿方向走 N 步找空地 → 放 nibble 15 | [`87`](87-enemy-map-movement.md) | `sub_15036` ＝ 敵人在地圖上移動（`ds:A643h` 的三句訊息定身分）；`ds:711Dh` 是執行期填的 |
| [`86`](86-combat-messages.md) | 敵人名稱 ＝ 執行檔字串表 1 的 `0x52 + Kind`（`Animal\n\ns\n` 取單數） |
| [`85`](85-enemy-map-icon.md) | 敵人圖示 ＝ `ds:A5B1h`（03 05 07）→ 敵人資料 `+0x06` → `ds:AA17h`；三段都在別的筆記解過 |
| [`84`](84-render-coverage.md) | 呈現層門檻：42 張地圖 ＋ 23 家設施 ＋ 戰鬥畫面；像素值域必須在 EGA 16 色內 |
| [`83`](83-translation-coverage.md) | 中文化覆蓋率門檻：缺的必須在 untranslatable 清單裡、孤兒 key ＝ 0 |
| [`82`](82-save-round-trip.md) | 存檔三道門檻（byte-for-byte／改動限縮／存讀一致）；`+0x0A` 所在地圖還沒存 |
| [`81`](81-combat-loop-coverage.md) | 戰鬥端到端門檻；`EquipIndex` 是**背包槽號**不是物品 ID（拿去查表傷害會高十倍） |
| [`80`](80-trainer-skill-list.md) | 訓練師的清單 ＝ 整張技能資料表；`sub_1BD7F` 只是印技能點數、`sub_1BDFF` 是查技能欄空位 |
| [`79`](79-facility-coverage.md) | 設施 `+0x00 & 0x7F` **≥ 5 就是 opcode −5**（`ds:A4E0h` 沒有上限檢查）；設施覆蓋率門檻 |
| [`78`](78-encounter-spawn.md) |
| `0x1A526` | — | **opcode 3**：`6 ≤ ds:465Ah < 18` → 用記錄 `+0x03`／`+0x04`，否則 `+0x05`／`+0x06`，搬進 `+0x01`／`+0x02` | [`75`](75-desert-heat-entry.md) §2 |
| `0x12BD0` | 1 | **nibble 12**：印記錄 `+0x00`，跑 `+0x01` 起的批次改寫表，最後改寫腳下 | [`71`](71-nibble12-batch-patch.md) §1 |
| `0x15160` | 1 | **nibble 8** 問答；答對的改寫位移 ＝ `3 + 答案數 + 2 × 序號`（`0x1522F`） | [`71`](71-nibble12-batch-patch.md) §5 |
| `0x16CD0` | 1 | **nibble 1**：印記錄開頭的訊息串列（bit7 結束），條數當位移改寫這一格 | [`70`](70-nibble1-and-facility-entry.md) §1 |
| `0x169B1` | 6 | 改寫**隊伍腳下**那一格（`sub_17CFF(al, ds:46A6h, ds:46A7h)`） | [`70`](70-nibble1-and-facility-entry.md) §1 |
| `0x12C80` | 2 | 設施／腳本分派：bit7 設 → `ds:A4E0h` 跳表，沒設 → `ds:A4EAh`（**同一張表差 5 個 word**） | [`70`](70-nibble1-and-facility-entry.md) §4.1 |
| `0x17D34` | 2 | `0xFE`／`0xFD`：把 `ds:46FCh`／`46FDh`（上一格改寫前的值）當成新值 | [`69`](69-gate-flags.md) §7 |
| `0x17CD2` | 9 | 讀某一格目前的 (nibble, 記錄) | [`69`](69-gate-flags.md) §7 |
| `0x142B1` | 2 | `& 8` 那族的改寫位移 ＝（條件串列 `0xFF` 的位置 + 1）＋ 2 × 通過的條件序號 | [`69`](69-gate-flags.md) §4 |
| `0x14296` | 2 | 對**全隊每個人**各套一次 `sub_14193` 的懲罰 | [`69`](69-gate-flags.md) §5 |
| `0x142ED` | 3 | 暫時把時鐘的時換成 `ds:A5C5h`，顯示記錄 `+0x03`，再還原 | [`69`](69-gate-flags.md) §5 |
| `0x14175` | 2 | 印記錄 `+0x02`（條件通過） | [`69`](69-gate-flags.md) §3 |
| `0x1417A` | 3 | 印記錄 `+0x03`（條件沒過且沒人受罰） | [`69`](69-gate-flags.md) §3 |
| `0x19F12` | 1 | **沖出目前這一行**：斷字 → 寫 scrollback（`seg003:0x8CE0`，40 × 256 環形）→ 送畫面 → 續行 [`58`](58-line-flush-and-scrollback.md) §2 |

| `0x19EFC` | 多 | 換行（`0x0D`）：`ds:46F0h` 非 0 就整支跳過，否則 `sub_19F12` ＋ `sub_1A0C5`（捲動與延遲） | [`58`](58-line-flush-and-scrollback.md) §3 |
| `0x19D86` | 3 | `base + Nd6` | [`13`](13-rng.md) §3.3 |
| `0x19C84` | 11 | 2d6 逢同點續擲（通用**檢定骰**） | [`13`](13-rng.md) §3.4、[`21`](21-attributes.md) §3 |
| `0x14193` | 3 | **讀效果記錄**（`+0x08`／`+0x09`）並算出值 | [`19`](19-effects-and-damage.md) §2 |
| `0x141FA` | 2 | **套用效果**到角色欄位（地圖事件的唯一出口） | [`19`](19-effects-and-damage.md) §3 |
| `0x157D6` | 3 | **傷害套用**（護甲擲骰、扣 CON、傷勢） | [`19`](19-effects-and-damage.md) §4 |
| `0x1B108` | 2 | **累加命中門檻**（飽和加法，夾在 100） | [`88`](88-hit-accumulator.md) §1 |
| `0x1B0F1` | 2 | 取 Brawling 等級 × 3（`mov al, 1` 寫死） | [`88`](88-hit-accumulator.md) §2 |
| `0x198CD` | 10 | 角色技能陣列裡找某個技能 ID 的等級（找不到回 0） | [`88`](88-hit-accumulator.md) §2 |
| `0x1B15F` | 3 | 取那一組敵人資料的 `+0x02`（行動值欄位） | [`88`](88-hit-accumulator.md) §4 |
| `0x172AE` | 10 | CON 兩個 byte 都是 0 ＝ **死** | [`89`](89-enemy-target-and-down.md) §2 |
| `0x172BB` | — | CON **≤ 0** ＝ **倒下**（不能行動、不會被挑中） | [`89`](89-enemy-target-and-down.md) §2 |
| `0x19D0E` | 10 | 數還有幾個人能行動（ZF ⟺ 全滅） | [`89`](89-enemy-target-and-down.md) §3 |
| `0x10E9A` | 3 | **`nullsub_1`**：邊框色常式的入口被 patch 成 `retn`，三個呼叫端全部沒作用 | [`23`](23-picture-format.md) §7.1 |
| `0x19BEC`／`0x19BF8` | — | 把比較暫存 `ds:46BEh`／累加器 `ds:46C0h` 歸零 | [`88`](88-hit-accumulator.md) §4 |
| `0x1C6C9` | 1 | **角色建立** | [`21`](21-attributes.md) §5 |
| `0x1CAD1` | 2 | 屬性擲法：5d6 取最高三顆 | [`21`](21-attributes.md) §5.1 |
| `0x17AE0` | 12 | 物品資料表定址（基址 `0x7A39`、stride 8） | [`22`](22-shop-and-items.md) §2 |
| `0x1C1CC` | 2 | 價格公式 | [`22`](22-shop-and-items.md) §3 |
| `0x14664` | — | **遭遇掃描**（掃視窗、填遭遇佇列、同步敵方記錄） | [`39`](39-encounter-scan.md) |
| `0x14A88` | 3 | 寫一筆遭遇佇列（x／y／地圖／距離） | [`39`](39-encounter-scan.md) §1 |
| `0x14AEA` | 3 | 依 (地圖, x, y) 找遭遇佇列 | [`39`](39-encounter-scan.md) §1 |
| `0x149F7` | 2 | 對四個隊伍組各評一次這一格 → `[0x7111 + 組]` | [`39`](39-encounter-scan.md) §4 |
| `0x14A65` | 2 | 挑**距離最近**的那一組（從 `0xFF` 起跳） | [`39`](39-encounter-scan.md) §4 |
| `0x11F76` | — | **戰鬥指令階段**（逐人下令，寫 `ds:46D8h`／`ds:46DAh`） | [`38`](38-combat-commands-and-flee.md) §1 |
| `0x173B0` | 多 | 按鍵 → 字母表索引（`ds:4680h` 線性掃） | [`38`](38-combat-commands-and-flee.md) §2 |
| `0x123A6` | — | **逃跑**（Party／Single → 選方向，不擲骰） | [`38`](38-combat-commands-and-flee.md) §3 |
| `0x172BB` | 多 | 能不能下令（CON ≤ 0 不行） | [`38`](38-combat-commands-and-flee.md) §4 |
| `0x137F4` | 27 | **敵方記錄定址**（`0x6B31 + 0x178a + 0x5Eb`，停在標頭） | [`37`](37-enemy-records-and-hp.md) §1 |
| `0x137CE` | 11 | 同上再 `+4 + 0x1E×g`，停在某一組的 30-byte 區塊 | [`37`](37-enemy-records-and-hp.md) §1 |
| `0x14480` | — | **遭遇佇列 → 敵方記錄**（敵人血量就在這裡擲） | [`37`](37-enemy-records-and-hp.md) §3 |
| `0x18E41` | 多 | **1dN**（2^k−1 遮罩 ＋ 拒絕取樣；`n ≤ 1` 原樣回傳） | [`13`](13-rng.md) §3.2 |
| `0x19D4D` | 多 | **距離**＝查 `ds:CD0Dh`[10·\|Δy\| + \|Δx\|]（5 × 10 的資料表） | [`37`](37-enemy-records-and-hp.md) §4 |
| `0x18D85` | 多 | abs | [`37`](37-enemy-records-and-hp.md) §4 |
| `0x12A76` | 2 | **武器傷害**＝基底 ＋ Nd6 | [`20`](20-combat-resolution.md) §4 |
| `0x19C2C` | 24 | 16-bit 飽和加減（累加器 `ds:46C0h`） | [`20`](20-combat-resolution.md) §3 |
| `0x18E90` | 28 | 等待按鍵（ESC 以 CF 回報） | [`03`](03-boot-and-asset-loading.md) §8 |
| `0x18EFE` | 6 | 鍵盤輪詢（順便推進 RNG） | [`13`](13-rng.md) §2.1 |
| `0x1786E` | 28 | 印字串 | [`06`](06-resource-directory.md) §3 |
| `0x178A3` | — | 印第 N 個字串（打包表） | [`17`](17-packed-text.md) §2.3 |
| `0x17B8F` | — | 取下一個字元（5-bit 解碼 ＋ 大寫旗標） | [`17`](17-packed-text.md) §2.2 |
| `0x17BC7` | — | 5-bit 位元讀取器 | [`17`](17-packed-text.md) §2.1 |
| `0x19E53` | 9 | 每字元處理：組行緩衝 | [`14`](14-fonts-and-text-encoding.md) §4 |
| `0x19DC3` | 11 | 每字元處理：直接繪製 | [`14`](14-fonts-and-text-encoding.md) §4 |
| `0x1060C` | — | 畫一個字元（主文字字型，overlay slot 5） | [`14`](14-fonts-and-text-encoding.md) §2 |
| `0x10CB6` | — | 畫一個字元（彩色字型，overlay slot 19） | [`14`](14-fonts-and-text-encoding.md) §3 |
| `0x17451` | 4 | 畫選單詞（彩色字型，置中） | [`14`](14-fonts-and-text-encoding.md) §5 |
| `0x17574` | 4 | 選單：按鍵比對字首字母 | [`14`](14-fonts-and-text-encoding.md) §5 |
| `0x16890` | 2 | **遭遇生成**（讀標頭 `+0x2F/+0x31/+0x32`、擲骰、填 section 15 槽） | 本表 §4.2 |
| `0x14BF0` | — | **排這一回合的敵方移動計畫**：16 格 `ds:711Dh` 清成 `0xFF`，再逐筆遭遇決定換位置／逃跑／逼近 | [`101`](101-enemy-move-plan-table.md) |
| `0x13A56` | 1 | 把敵人編號拆成 `(n div 3) >> 2`、`(n div 3) & 3`、`n mod 3` —— 前兩個重組回記錄編號，第三個是記錄裡的 30-byte 子組 | [`101`](101-enemy-move-plan-table.md) §3 |
| `0x18350` | — | **換地圖**：讀邊長 `+0x2C`、圖磚組 `+0x30`，必要時重載圖磚 | [`24`](24-map-layers-and-tiles.md) |
| `0x17C20` | — | 取地圖第 1 層的 nibble（section 型別） | [`24`](24-map-layers-and-tiles.md) §2.1 |
| `0x17C72` | — | 算地圖第 2 層的列基址（`ds:46ACh`） | [`24`](24-map-layers-and-tiles.md) §2.1 |
| `0x17FC8` | — | 算地圖第 3 層的列基址（`ds:46CAh`） | [`24`](24-map-layers-and-tiles.md) §2.1 |
| `0x1379E` | — | 用第 1／2 層的值取出該格的記錄 | [`24`](24-map-layers-and-tiles.md) §2.2 |
| `0x1029B` | 4 | **畫一格地圖**（背景 AND 遮罩 OR 疊圖，overlay slot 4） | [`24`](24-map-layers-and-tiles.md) §2.3 |
| `0x10156` | 1 | 畫一格地圖（不透明，overlay slot 3） | [`24`](24-map-layers-and-tiles.md) §2.3 |
| `0x18024` | 5 | 畫一格：依第 1 層 nibble 選疊圖（5 寶箱／8 輻射／記錄 `+0x01`），再叫 slot 3／4 | [`48`](48-map-icons.md) §2.1 |
| `0x16675` | 1 | **重畫整個地圖視窗**（19 欄 × 9 列） | [`25`](25-screen-layout.md) §2.1 |
| `0x16619`／`0x16646` | 各 1 | 畫一欄（9 格）／畫一列（19 格） | [`25`](25-screen-layout.md) §2.1 |
| `0x169CF` | 1 | 座標在不在地圖視窗內（19 × 9） | [`25`](25-screen-layout.md) §2.1 |
| `0x16149` | 1 | 算視窗原點 ＝ 隊伍座標 − (9, 4) | [`25`](25-screen-layout.md) §2.1 |
| `0x16716` | — | **畫隊伍圖示**：疊圖編號 7 ＋ overlay slot 4 | [`48`](48-map-icons.md) §2 |
| `0x16428` | — | **畫其他分隊**：掃四組隊伍槽，同地圖的畫疊圖 9 | [`48`](48-map-icons.md) §2 |
| `0x197BB` | 2 | 畫外框（欄 0–37、字元列 0–17） | [`25`](25-screen-layout.md) §2.3 |
| `0x10762` | 10 | 清除矩形（同時定義了兩套座標單位） | [`25`](25-screen-layout.md) §1 |
| `0x1651A` | 4 | **走一步**（方向 0 上／1 下／2 左／3 右） | [`26`](26-movement-and-triggers.md) §1 |
| `0x1676A` | 1 | **推進遊戲時鐘**（每步 ＋ 標頭 `+0x34`／`+0x36`） | [`27`](27-game-clock.md) §2 |
| `0x17DF1` | 1 | 畫時鐘（外框上緣，字元欄 28、列 0） | [`27`](27-game-clock.md) §4 |
| `0x12440` | 2 | **每 16 刻的體力處理**：健康每 64 刻 ＋1、生病每 64 刻 −1、CON 破 −50 死亡 | [`35`](35-status-and-healing.md) §2 |
| `0x1649E` | 1 | 能不能走（邊界 ＋ nibble ＋ 三道檢查） | [`26`](26-movement-and-triggers.md) §3 |
| `0x16410` | 1 | **事件分派**（`ds:AA87h`[nibble] 間接呼叫） | [`26`](26-movement-and-triggers.md) §5 |
| `0x169EB` | 多 | 取這一格的 nibble 並把 `ds:46AEh` 指到該筆記錄 | [`26`](26-movement-and-triggers.md) §5 |
| `0x15280` | 跳表 | **寶箱**：類別 → 擲出具體物品與數量，寫回記錄 | [`29`](29-map-event-handlers.md) §4 |
| `0x15453` | 1 | 數出某個類別有幾件物品（掃物品表 1–94） | [`29`](29-map-event-handlers.md) §4 |
| `0x15160` | 跳表 | **問答**：單鍵或打字，答案清單在記錄 `+0x03` 起 | [`46`](46-typed-answers-and-text-input.md) §4 |
| `0x17750` | 2 | **文字輸入本體**（緩衝區 `ds:4680h`、上限 `ds:4684h`） | [`46`](46-typed-answers-and-text-input.md) §2 |
| `0x18D8E` | 2 | **字串比對**：兩個 NUL 結尾字串逐 byte 全等 | [`46`](46-typed-answers-and-text-input.md) §3 |
| `0x17CFF` | 7 | **改寫地圖格**：記錄裡兩個 byte ＝（新第 1 層 nibble、新第 2 層記錄編號） | [`46`](46-typed-answers-and-text-input.md) §4.1 |
| `0x17D7A` | 1 | 寫第 1 層的 nibble（讀出、清舊、or 新值、寫回） | [`46`](46-typed-answers-and-text-input.md) §4.1 |
| `0x18EFE` | 多 | 鍵盤讀取；含 **F1–F10 巨集錄放**（10 × 256 bytes @ `ds:C062h`）；出口把 `a`–`z` 轉成大寫 | [`43`](43-input-and-hotkeys.md) §6、[`46`](46-typed-answers-and-text-input.md) §2.1 |
| `0x167CE` | 2 | 重畫視窗內 nibble 為 4／5／9 的格子（會動的） | [`26`](26-movement-and-triggers.md) §5 |
| `0x17FEE` | — | 地圖座標 → 螢幕座標（`ds:4685h`／`ds:4686h`） | [`24`](24-map-layers-and-tiles.md) §2.3 |
| `0x10088` | — | 圖磚 packed 4bpp → EGA 4 平面（overlay） | [`24`](24-map-layers-and-tiles.md) §3.1 |
| `0x184E8` | — | **載入一張 `ALLPICS` 圖**（解壓 `0xFC0` ＋ 參數區，再叫 slot 2／16） | [`23`](23-picture-format.md) §4 |
| `0x10144` | — | **圖片 delta 解碼**（overlay slot 2，18 bytes） | [`23`](23-picture-format.md) §2 |
| `0x10A7A` | — | 拆圖片參數區成兩張表：A ＝ `(延遲, 格)` 播放腳本、B ＝ 每格像素（overlay slot 16） | [`23`](23-picture-format.md) §5.1 |
| `0x10B11` | — | 局部動畫：依 BIOS `0040:006C` 推進，逐格 **XOR 進 EGA 平面**（overlay slot 17）；列位址表 `ds:8E09h` ＝ slot 0 那張 `ds:8DF9h` 往後 8 筆，y ＋ 8 已烘在表裡 | [`23`](23-picture-format.md) §5.2 |
| `0x186B6` | — | **載入一組圖磚**到 `seg003:0x2F60`（含 delta 解碼） | [`24`](24-map-layers-and-tiles.md) §3 |
| `0x18744` | — | 存檔載入（兩份輪替，比 32-bit 序號） | [`09`](09-msq-map-structure.md) §4 |
| `0x190A6` | 5 | 設施畫面：載入 ALLPICS 圖 ＋ 印 13 bytes 的地點名稱 | [`29`](29-map-event-handlers.md) §5.4 |
| `0x1BA72` | — | **顯示角色卡 ＋ 升級加 1 點技能點**（`0x1BB18`） | [`31`](31-experience-and-skills.md) §2 |
| `0x1BB5D` | — | 印階級名（把 `ds:4692h` 指到 `0xD622`） | [`31`](31-experience-and-skills.md) §2 |
| `0x132AC` | — | **`Hire` 的結算**：section 17 取 NPC 記錄 → 抄 256 bytes → 魅力對決 → 槽表補位（7 人上限）| [`110`](110-hire-resolution.md) |
| `0x19E2A`／`0x19E30` | 6／2 | **反白開／關**（`ds:4678h`）——卡彈的武器名、生病的隊員 | [`111`](111-roster-inverse-video.md) |
| `0x1BB6C` | — | **把階級名抄進記錄 `+0x32`**：逐字元到 NUL 為止，位移在 `ds:D430h`，**沒有長度檢查** | [`109`](109-character-record-tail.md) §4 |
| `0x1C68E` | — | **技能費用**：基礎 × 2^(等級−1)，飽和 `0xFF` | [`31`](31-experience-and-skills.md) §3.1 |
| `0x1CA8D`／`0x1CA98` | 各 2 | 技能資料 `+0x00` 拆欄位：`& 7` ＝ 費用／`>> 3` ＝ IQ 需求 | [`31`](31-experience-and-skills.md) §3 |
| `0x19BC0` | 6 | **經驗值 += `ds:466Bh` 24-bit**（溢位飽和 `0xFFFFFF`） | [`32`](32-skill-checks-and-xp.md) §7 |
| `0x16F4C` | 5 | **資料表定址器**：`ds:4665h ← ds:4694h ＋ 索引 × ds:469Bh`（基址與 stride 是可換的） | [`32`](32-skill-checks-and-xp.md) §1 |
| `0x18134` | 2 | 把定址器切到技能資料表（`ds:BA20h`，stride 2） | [`32`](32-skill-checks-and-xp.md) §2 |
| `0x12A40` | 13 | 把定址器切到攻擊資料表（記錄區標頭 `+0x04`，stride 8） | [`32`](32-skill-checks-and-xp.md) §8.1 |
| `0x180F0` | 2 | **技能檢定**（含成功後練等） | [`32`](32-skill-checks-and-xp.md) §3 |
| `0x18146` | 2 | 檢定的骰與加成（2d6 續擲 ＋ 屬性 ＋ 技能×3） | [`32`](32-skill-checks-and-xp.md) §3 |
| `0x1818E` | 3 | **技能自動升級**（上限 ＝ 角色等級） | [`32`](32-skill-checks-and-xp.md) §5 |
| `0x1820C` | 2 | **屬性／等級／性別檢定** | [`32`](32-skill-checks-and-xp.md) §4 |
| `0x13EC9` | 1 | **走地圖記錄的條件串列**（`+0x0A` 起，每 2 bytes 一條） | [`32`](32-skill-checks-and-xp.md) §6 |
| `0x159C7` | 3 | **算擊殺經驗值** | [`32`](32-skill-checks-and-xp.md) §8 |
| `0x198CD` | 10 | 取某個技能的等級（掃 `+0x80`，沒學過回 0） | [`32`](32-skill-checks-and-xp.md) §3 |
| `0x1968C` | 8 | 找某件物品（掃 `+0xBD`） | [`32`](32-skill-checks-and-xp.md) §6 |
| `0x19A58` | 2 | 消耗物品一次（附屬 byte 低 6 位是次數，歸零就移除） | [`32`](32-skill-checks-and-xp.md) §6 |
| `0x1CB75` | — | 掛 `int 08h`，重設 8253 channel 0 | [`04`](04-overlay-wla-bin.md) |
| `0x1CC76` | — | 更新四個聲部並仲裁出一個送進喇叭 | [`44`](44-audio.md) §3.2 |
| `0x1CD52` | — | **音效位元組碼直譯器**（四個 opcode） | [`44`](44-audio.md) §4 |
| `0x1CBD3` | — | 播一個音效（`al` ＝ 編號 0–8） | [`44`](44-audio.md) §6 |
| `0x1CCC9` | — | 一個聲部的一個 tick（滑音、封套、顫音） | [`44`](44-audio.md) §3.1 |
| `0x162E1` | 2 | 等 `dl` 個計時器 tick（忙等 `ds:D150h` 變化） | [`44`](44-audio.md) §8.1 |
| `0x15755` | 3 | **武器傷害骰**：`0 ＋（物品 +0x06）顆 d6`；類別 8／9 加倍 | [`45`](45-item-data-and-weapon-damage.md) §4 |
| `0x13878` | 6 | 算接戰值（`ds:46CCh`）並把裝備武器設成當前定址的物品 | [`45`](45-item-data-and-weapon-damage.md) §4.1 |
| `0x199F1` | 7 | 物品 `+0x03 >> 3` ＝ **類別**（跳的是 `loc_17C6B`，右移三次不是四次） | [`45`](45-item-data-and-weapon-damage.md) §3.1 |
| `0x19A0F` | 2 | 物品 `+0x06`（骰數）；入口 `0x19A14` 取 `+0x04`（彈匣） | [`45`](45-item-data-and-weapon-damage.md) §3.2 |
| `0x19D2F` | 13 | 「這個武器類別在 `ds:CD00h` 清單裡嗎」＝ 有沒有射程 | [`45`](45-item-data-and-weapon-damage.md) §3.1 |
| `0x10D4D` | 1 | **畫滑鼠游標**：`CURS ＋ ds:8DCDh × 256`，遮罩 AND ＋ 圖形 OR | [`112`](112-mouse-cursor-and-hotzones.md) §2 |
| `0x18C62` | 1 | 地圖視窗的滑鼠熱區：中央方框 → 游標 6 ＋ ESC，四個楔形 → 游標 2–5 ＋ `I K J L` | [`112`](112-mouse-cursor-and-hotzones.md) §4 |
| `0x16385` | 15 | 片頭印一條字幕（設字串表 `ds:A703h` 再印第 `al` 條） | [`113`](113-attract-mode.md) §2 |
| `0x198F0` | 2 | **技能清單選擇**（`USE` 的 `S`）：建表 → 清單框架 → 可還原的備份 | [`92`](92-use-command.md) §3.1 |
| `0x14FDE` | 1 | 把一組敵人十筆 16-bit 血量加總進 `ds:46BEh`（士氣判定用） | [`101`](101-enemy-move-plan-table.md) §6.2 |
| `0x12AC5` | 6 | 遭遇記錄 `+0x09 >> 2`，**CF ＝ bit1（不敵對）** | [`114`](114-friendly-encounters.md) §2 |
| `0x15C08` | 1 | 遭遇記錄 `+0x09 >> 4` ＝ **section 17 的 NPC 記錄編號** | [`114`](114-friendly-encounters.md) §2 |
| `0x15C19` | 2 | **翻臉**：清掉 `+0x09` 的 bit1、設 bit0（隊伍一開槍就跑，命中判定之前）| [`114`](114-friendly-encounters.md) §3 |
| `0x13787` | 多 | 拿遭遇的 (x, y) 查那一格的 nibble 當 section、第 2 層當編號 → `ds:46C6h` | [`114`](114-friendly-encounters.md) §1 |
| `0x1949E` | 2 | 裝備／卸下一件物品；護甲會把 `+0x06` 寫進記錄 `+0x1A` | [`45`](45-item-data-and-weapon-damage.md) §3.4 |

## 7. 工具

| 工具 | 做什麼 | 需要 IDA |
|---|---|---|
| `tools/ida.sh` | IDA 的唯一入口（`build`／`run`／`raw`） | — |
| `tools/ida/export_listing.py` | 整個 CODE 區倒成逐指令 JSON（20,177 條） | 是 |
| `tools/ida/export_memops.py` | 全檔直接定址存取（827 個全域、4,932 筆） | 是 |
| `tools/ida/export_function.py` | 指定函式完整倒出（含呼叫端） | 是 |
| `tools/ida/export_mnemonic.py` | 掃指定助憶碼 | 是 |
| `tools/ida/export_inventory.py`／`export_file_io.py`／`export_range_refs.py` | 清冊／中斷與字串／範圍引用 | 是 |
| `tools/summarize_record_fields.py` | 從逐指令 JSON 掃記錄欄位（基址暫存器型存取） | 否 |
| `tools/summarize_msq_layout.py` | MSQ 區塊佈局與 section 統計 | 否 |
| `tools/decode_text.py` | 執行檔九張打包字串表 | 否 |
| `tools/decode_block_text.py` | 42 個區塊各自的字串表 | 否 |
| `tools/decrypt_msq.py`／`split_resources.py`／`huffman.py` | 解密／切資源／解壓 | 否 |
| `tools/dump_font.py` | 兩套字型畫成文字圖 | 否 |
| `tools/decode_pic.py` | 圖片與圖磚的 delta 解碼 ＋ 4bpp 畫成文字圖 | 否 |
| `tools/summarize_map_layers.py` | 地圖三層的長度、邊長、圖磚組驗證 | 否 |
| `tools/render_map.py` | 用第 3 層把一張地圖畫成一格一字元的縮圖 | 否 |
| `tools/dump_word_table.py` | 倒出執行檔裡的 16-bit 表（跳表、位移表） | 否 |
| `tools/dump_save.py` | 解開存檔區並列出隊伍、時鐘、角色 | 否 |
| `tools/dump_items.py` | 解開物品資料表並配上物品名／技能名 | 否 |
| `tools/summarize_questions.py` | 42 個區塊的 nibble 8 問答 ＋ 產生「不可翻譯」守則清單 | 否 |
| `tools/dosbox.sh` | headless DOSBox 跑原版：送鍵、截圖（`docker/dosbox/`） | 否 |
| `tools/compare_screen.py` | 截圖與解碼器輸出**逐像素**對拍（位置用掃的，會報次佳差距） | 否 |
| `tools/ida/export_forced.py` | 強制把 IDA 漏掉的位址分析成程式碼再倒出 | 是 |
| `tools/rng.py` | 亂數與擲骰的參考模型（附自我測試） | 否 |
| `tools/unpack_exepack.py`／`apply_overlay.py` | 解包／合成分析映像 | 否 |
| `tools/summarize_sfx.py` | `seg005` 的音高／音長／音效表 ＋ 九首位元組碼反組譯 | 否 |
| `tools/scan_callers.py` | 全檔掃某函式的直接呼叫點（far ＋ 同段 near），替 xref 做正對照 | 否 |
| `tools/gen_func_index.py` | 產生 `00-function-index.md` | 否 |
| `cmd/wl-atlas` | 42 張地圖裡玩家碰得到的東西倒成 JSON（設施招牌、條件閘、問答、傳送、藏東西的格）。攻略的資料來源 | 否 |
| `tools/summarize_walkthrough.py` | 把上面那份 JSON 整理成 `docs/walkthrough/generated/` 四份表 | 否 |

產物一律落在 `docs/re/generated/ida94/`（工具輸出，不含人的推論）。

## 8. 文件索引

| 文件 | 內容 |
|---|---|
| [`01`](01-binary-identity.md) | 20 個檔案的 SHA-256、MZ header、打包版不可引用 |
| [`02`](02-exepack-unpack.md) | EXEPACK 格式與解包器 |
| [`03`](03-boot-and-asset-loading.md) | 開機序列、檔名表、七個素材的載入位址、`TITLE.PIC` |
| [`04`](04-overlay-wla-bin.md) | overlay 機制、26 個 slot、EGA mode 0Dh |
| [`05`](05-storage-layer.md) | 雙模式儲存、資源表 |
| [`06`](06-resource-directory.md) | `GAME1`／`GAME2` 的目錄與位移表 |
| [`07`](07-msq-blocks.md) | 42 個 MSQ 區塊的切分與三重驗證 |
| [`08`](08-msq-encryption.md) | XOR 加密與 checksum |
| [`09`](09-msq-map-structure.md) | 地圖層、明文名稱、存檔兩份輪替 |
| [`10`](10-huffman-compression.md)／[`11`](11-huffman-decoder.md) | Huffman 格式與解碼器 |
| [`12`](12-msq-tail-and-text-model.md) | 尾段是 Huffman、文字在各來源的分佈 |
| [`13`](13-rng.md) | 亂數與擲骰層 |
| [`14`](14-fonts-and-text-encoding.md) | 兩套字型、18 個文字控制碼、選單 |
| [`15`](15-character-record.md) | 角色記錄的定址與欄位 |
| [`16`](16-msq-block-layout.md) | MSQ 區塊佈局與兩層記錄索引 |
| [`17`](17-packed-text.md) | 5-bit 打包文字與執行檔九張表 |
| [`18`](18-block-text.md) | 地圖區塊的 4,401 條敘述文字 |
| [`19`](19-effects-and-damage.md) | 資料驅動的效果系統、傷害與護甲、單複數選擇器 |
| [`20`](20-combat-resolution.md) | 命中判定（d100 對門檻）、武器傷害公式、一次攻擊的完整流程 |
| [`88`](88-hit-accumulator.md) | 命中累加值的四個項（Brawling×3、Agility、對手行動值、基礎值） |
| [`89`](89-enemy-target-and-down.md) | 敵人隨機挑目標與重抽；CON ≤ 0（倒下）與 CON ＝ 0（死）的分野 |
| [`90`](90-party-initiative.md) | 隊伍的行動值公式，以及「只有下攻擊令的人才排進行動表」 |
| [`91`](91-map-command-bar.md) | 地圖指令列的七個處理程式；**升級的入口是 RADIO** |
| [`92`](92-use-command.md) | `USE` 的第一層：Skill／Item／Attribute 三選一 |
| [`93`](93-order-disband-view.md) | `ORDER`／`DISBAND`／`VIEW`；後兩支的前置是多隊伍 |
| [`94`](94-enc-command.md) | `ENC` ＝ 遭遇驅動器的手動入口（`0x11CE7` 是 `sub_11CD0` 的中途入口）|
| [`95`](95-main-menu.md) | 主選單 `sub_1630C`：只有 `Start` 一項，沒有新遊戲／讀檔 |
| [`96`](96-ending.md) | 結局 `0x1B4F0` ＝ 設施跳表 `ds:A4E0h` 第 4 格；`END.CPA` 第二段是動畫腳本 |
| [`97`](97-playtest-sampling.md) | 抽樣試玩：七段流程各走一遍，六個「編得過、測得過、玩不動」的缺口 |
| [`98`](98-a0-wiring.md) | 中文的三層接線、店家庫存跟著存檔走、「全隊倒下遊戲照走」是原版行為 |
| [`99`](99-party-wipe.md) | 全隊陣亡三分支（什麼都不做／自動換隊／死亡畫面）；救得回來 ＝ CON ∈ [−10, −1] 且狀態位元全 0 |
| [`100`](100-ending-trigger.md) | **結局的觸發點**：資料裡沒有第 4 格，是 `sub_1CB30`（主迴圈 `0x16C28`）在自毀倒數 240 刻到期時合成 `al ← 84h`；倒數由腳本 opcode 35（資源 20 記錄 4）啟動 |
| [`101`](101-enemy-move-plan-table.md) | `ds:711Dh` ＝ 敵方這一回合的**移動計畫**（16 格，每回合由 `sub_14BF0` 清成 `0xFF`）；命中基礎值 `0xFF` → **50**、否則 60 |
| [`102`](102-unreachable-opcodes.md) | 走不到的 12 個腳本 opcode 逐支讀完：分隊位置、寄放隊員、倒數、批次改寫、鄰格比對、時間戳 |
| [`103`](103-roster-line-columns.md) | 名片行是 **39 欄**、行首是**序號 ＋ `>`**；`AMM` 三道閘、`WEAPON` 取單數形（`Kni\nfe\nves\n` → `Knife`）|
| [`104`](104-opcode-2-icon-swap.md) | 腳本 opcode 2 ＝ overlay slot 18：把兩張圖形對調（含遮罩），換掉地圖上某個圖示的長相 |
| [`105`](105-enc-empty-round-and-menu-region.md) | `ENC` 在**沒有敵人**時問字串 `0x14`，答 Y 照樣進指令階段；戰鬥的指令選單畫在**欄 15–38、列 1–13**，不是訊息視窗 |
| [`106`](106-text-scroll.md) | 文字區塊滿了**捲一個字元列**（overlay slot 10），不是切掉；`ds:465Bh` ＝ 捲動速度 0–8（`<`／`>` 調），四張 9 筆表乘起來恆為 8 |
| [`107`](107-command-resolution.md) | 指令有**第二張跳表** `ds:A568h`（結算階段）：換裝備是切換、裝填吃掉整個彈匣、迴避只印一句；`ds:A43Bh` 只管下令 |
| [`108`](108-combat-use-and-hire.md) | 戰鬥 `Use` 的參數 ＝ `(選項 << 4) | 方向`，編號另存 `ds:A9FDh[角色編號]`；九向位移表 `ds:AAB1h`（第 5–8 格是對角，選不到）|
| [`109`](109-character-record-tail.md) | 角色記錄 `+0x4D`–`+0x7F` 那 51 bytes **沒有任何存取點**；階級欄是 `+0x32`–`+0x4A`（25 bytes），寫入端沒有長度檢查 |
| [`110`](110-hire-resolution.md) | `Hire` 的結算：遭遇記錄 `+0x09` 的 bit1（不敵對，見 `114`）＋ 高 4 位 ＝ section 17 的 NPC 編號 → **整筆 256 bytes 抄進隊伍** → 魅力對決 → 7 人上限 |
| [`111`](111-roster-inverse-video.md) | `sub_19E2A` ＝ 反白開（`ds:4678h`）：卡彈的武器名與有狀態的隊員畫成反白 |
| [`112`](112-mouse-cursor-and-hotzones.md) | 游標索引 `ds:8DCDh`、繪製常式 `0x10D4D`；八個圖形 ＝ 預設／可點／上下左右／中央方框，**第 7 個選不到**。地圖四個楔形送 `ds:C05Dh` 的 `I K J L` |
| [`113`](113-attract-mode.md) | 片頭 ＝ 第 0 張字串表的六頁，每頁 255 個計時器刻；進去要按過兩次非 `S` 的鍵，**槽 8 寫了沒有呼叫端** |
| [`114`](114-friendly-encounters.md) | 遭遇記錄 `+0x09`：bit0 名字來源／**bit1 不敵對**／bit2 不移動／高 4 位 ＝ NPC 編號。靜態遭遇在 **section 3**（出貨 14 筆可雇用），隨機的在 section 15。**開槍就翻臉**（`sub_15C19`）|
| [`115`](115-portrait-box.md) | 肖像框 ＝ `ds:46FAh`（圖）＋ `ds:7201h`（13 bytes 說明，＝ 存檔 `+0xD0`），**跨畫面模式活著**。戰鬥每回合 `sub_12636` 挑**最近的那一組**（`[ds:46C8h+3]` 最小），挑不到就畫遊俠（圖 8、字串 96）|
| [`21`](21-attributes.md) | 七個屬性的記錄位移、屬性→修正值階梯、檢定骰、角色建立 |
| [`22`](22-shop-and-items.md) | 商店、價格公式、物品資料表（95 筆 × 8 bytes） |
| [`23`](23-picture-format.md) | 圖片格式：packed 4bpp ＋ 列間 XOR delta、82 張 `ALLPICS` |
| [`24`](24-map-layers-and-tiles.md) | 地圖三層結構、`ALLHTDS` 九組圖磚、16 × 16 圖磚格式 |
| [`25`](25-screen-layout.md) | 畫面版面：座標單位、地圖／圖片視窗 288 × 128、訊息視窗、隊伍名單 |
| [`26`](26-movement-and-triggers.md) | 走一步的流程、四方向捲動、nibble → 事件處理的 16 筆跳表 |
| [`27`](27-game-clock.md) | 遊戲時鐘：24 小時制、每步推進量、晝夜門檻、隨時間的角色處理 |
| [`28`](28-text-variants.md) | 文字變形：單複數／性別／三選一／數量的骨架與選擇子 |
| [`29`](29-map-event-handlers.md) | 地圖事件處理：寶箱、選單、訊息；強制分析 IDA 漏掉的位址 |
| [`30`](30-save-layout.md) | 存檔佈局：位置、加密與 checksum、全域狀態、四組隊伍槽表、四個預設 Ranger |
| [`31`](31-experience-and-skills.md) | 經驗值與升級門檻 (L² − L) × 512、階級表、技能學習與費用公式 |
| [`32`](32-skill-checks-and-xp.md) | 資料表定址器、技能資料表（35 條數值）、技能／屬性檢定、練等、條件串列、經驗值三來源 |
| [`33`](33-paragraph-references.md) | 段落編號 ＝ 敘述文字的一部分（83 處）、陷阱段落零引用、密語是遊戲內謎題 |
| [`34`](34-map-script-opcodes.md) | 地圖腳本直譯器的 44 個指令 |
| [`35`](35-status-and-healing.md) | 八個狀態位元與疾病名、每 16 刻的體力恢復與惡化、醫生與訓練師 |
| [`36`](36-combat-rounds.md) | 戰鬥的回合：三組 × 10、行動旗標、行動順序表 |
| [`37`](37-enemy-records-and-hp.md) | 敵方記錄的兩層版面、敵人血量的擲法、敵人資料表、距離表 |
| [`38`](38-combat-commands-and-flee.md) | 戰鬥的指令階段、七個指令碼與跳表、迴避與逃跑 |
| [`39`](39-encounter-scan.md) | 遭遇怎麼冒出來：視窗掃描、遭遇佇列、兩個距離門檻 |
| [`40`](40-combat-screen.md) | 戰鬥畫面：名單模式、一回合的訊息序列、中文化的兩條硬規則 |
| [`41`](41-command-handlers.md) | 四支指令處理程式：共同形狀「檢查 → 選參數 → 回傳」，不執行動作 |
| [`42`](42-facility-loops.md) | 商店買賣的挑物品流程、逐點付錢治療、物品表 `+0x02` 是庫存量 |
| [`43`](43-input-and-hotkeys.md) | 鍵盤三種比對、21 筆滑鼠熱區表、`\x10` 登記的每列熱鍵表 |
| [`44`](44-audio.md) | 音效：計時器 ISR、四個聲部、位元組碼指令集、九首資料與呼叫端 |
| [`45`](45-item-data-and-weapon-damage.md) | 物品資料表八個欄位、表在存檔區、武器傷害骰（`sub_15755`） |
| [`46`](46-typed-answers-and-text-input.md) | 打字回答與密語比對、文字輸入常式、按鍵轉大寫、中文化硬約束 |
| [`47`](47-dosbox-oracle.md) | DOSBox 參考環境；解包正確性、`TITLE.PIC` 逐像素、調色盤、隊伍名單五條斷言 |
| [`48`](48-map-icons.md) | `IC0_9.WLF` 十張疊圖的語意（隊伍、五種敵人、寶箱、輻射區、其他分隊）、nibble 9 ＝ 輻射區 |
| [`49`](49-save-roundtrip-on-hardware.md) | 存檔改寫的實機驗收：round-trip、原版讀得進去、隊伍那一格的背景取地形 |
| [`50`](50-unnamed-items.md) | 物品 70／71／72 的空名字：被清空不是遺失、字母序把開頭夾在 H–M、哪些欄位不算證據 |
| [`51`](51-encounter-driver.md) | 遭遇驅動器 `sub_11CD0`：地圖與戰鬥之間那一層、四組一起結算、經驗值前後相減 |
| [`52`](52-trainer-facility.md) | 技能訓練師的流程：五個設施同一個模板、三條「走不通」都回選人 |
| [`53`](53-list-framework.md) | 清單框架 `sub_16DB4`／`sub_16D34`：列與索引的對應表、三個回傳值、I／K 翻頁 |
| [`78`](78-encounter-spawn.md) | 遭遇生成器全解：方向跳表 `ds:AAB1h` 九向、三張 13 項表、`+0x05 & 0x0F` 當索引 |
| [`77`](77-encounter-spawn-gap.md) | 敵人格是 `sub_16890` 每步生成的（section 15 是槽）；remake 缺生成器 → 隨機遭遇 0 次 |
| [`76`](76-script-opcode-coverage.md) | 腳本 opcode 覆蓋率盤點：37 格缺口 → **0 格**；`Handled ＝ false` 的統計與門檻測試 |
| [`75`](75-desert-heat-entry.md) | 沙漠高溫 ＝ 腳本 opcode 3 的晝夜分支（白天記錄 7–9、夜間 10–12）；`CF ＝ 0` 表示同一步繼續跑 |
| [`74`](74-heat-entry-and-gate-display.md) | `sub_142ED` ＝ 暫換時鐘的時 ＋ 印 `+0x03` ＋ 延遲 ＋ 還原；`export_range_refs` 是半開區間（單一位址要寫 `X X+1`） |
| [`73`](73-shop-and-doctor-entry.md) | **商店與醫生的入口**：`sub_169B1(4)` 用傳送記錄 `+0x04`／`+0x05` 改寫落點成 nibble 6 ＋ 設施（22 筆全中） |
| [`72`](72-facility-entry-and-command-bar.md) | 進地點 ＝ nibble 12 先把設施格改寫成 nibble 10；跳表索引 3 ＝ `CREATE DELETE PLAY`；指令列 `ds:A9CCh` |
| [`71`](71-nibble12-batch-patch.md) | nibble 12 ＝ 遠端批次改寫（`+0x01` 起每 5 bytes：旗標／x／y／新第 1 層／新第 2 層）；nibble 8 答對的改寫位移 ＝ `3 + 答案數 + 2n` |
| [`70`](70-nibble1-and-facility-entry.md) | nibble 1 ＝ 氛圍敘述串列（bit7 結束）＋ 收尾改寫；設施跳表 `ds:A4E0h` 與腳本跳表差 5 個 word；商店入口的程式碼側已封閉 |
| [`69`](69-gate-flags.md) | 條件閘的四個旗標（記錄 `+0x00` 低位）：`& 4` 有人過就算過、`& 8` 有人過就收尾且改寫位移依條件而定、`& 0x10` 全隊各罰一次、`& 0x20` 逐個角色跑；`0xFE`／`0xFD` ＝ 沿用上一格改寫前的值 |
| [`68`](68-cell-rewrite.md) | `sub_17CFF` 改寫地圖格：條件閘用 `+0x04`／`+0x06`、地形閘用 `+0x01`／`+0x02`；bit7 ＝ 不改 |
| [`67`](67-gate-penalty-and-canteen.md) | 條件閘的獎懲參數（`+0x08` ＝ 欄位／固定或擲骰、`+0x09` ＝ 量／加減）；`sub_14193` 全解 |
| [`66`](66-nibble2-event-and-heat.md) | nibble 2 的閘與事件同一支 `sub_13EC9`（踩上去印 `+0x01`）；沙漠高溫的線索 |
| [`65`](65-third-gate-conditions.md) | 第三道閘 `sub_13E9B`：nibble 2 先判條件串列再決定擋不擋（remake 目前無條件擋 ＝ 近似） |
| [`64`](64-enter-location-prompt.md) | 第三道閘 `sub_16AD5`：記錄 `+0x00` 的 bit6 → 問 `Enter new location?`（字串表 1 第 103 條），選 No 那一步整個不算 |
| [`63`](63-resource-id-vs-index.md) | 資源目錄 ID 與 `Resources()` 索引不同（28/42）；遊戲的地圖編號是 ID，拿索引會安靜載錯地圖 |
| [`62`](62-fourth-gate-terrain-blocking.md) | 第四道閘 `sub_15CE0`：nibble 11 一律擋、nibble 4 條件式；擋住時印記錄 `+0x00` 的訊息 |
| [`61`](61-map-id-table.md) | `ds:BF1Ch` 一表兩用：低半部決定標頭 `0x600`／`0x1800`，高半部把建築編號換成資源 5／11 |
| [`60`](60-teleport-and-map-change.md) | nibble 10 ＝ 傳送並換地圖；記錄 `+0x03` ＝ 目標地圖（`0xFF` ＝ 回程）；槽表 `+0x0B`–`+0x0D` ＝ 回程；**`+0x00` 低 6 位 ＝ 地點名字串編號**（`sub_16B17`，§3.1）|
| [`59`](59-playtest-against-original.md) | 玩家路徑對原版驗收：`cmd/wl-play`、輻射帶團滅是原版行為、Rad suit（物品 41）免疫 |
| [`58`](58-line-flush-and-scrollback.md) | 控制碼 `0x08` ＝ 沖出一行不捲動（`0x0D` 多包一層 `sub_19EFC`）；scrollback 40 × 256 環形 |
| [`57`](57-curs.md) | `CURS` ＝ 8 個 32 × 16 的滑鼠游標（左半遮罩、右半 2 色圖形）；平面連續不是逐列交錯 |
| [`56`](56-transtbl.md) | `TRANSTBL` ＝ 50 組 × 16 對照表；三層掃描都找不到消費端；順帶接上滑鼠初始化 |
| [`55`](55-radiation-and-armour-bypass.md) | 輻射結算迴圈（`0x14410`）、`ds:46EFh` ＝ 跳過護甲吸收、`sub_141FA` 的加減欄位分派 |
| [`54`](54-facility-screen-layout.md) | 設施畫面的版面：圖在 (8, 8) 96 × 84、地點名字元列 12、殘差 ＝ A9 的動畫（疊到第 3 格後差 0） |

## 9. 引用這份表時的紀律

- **本表只寫已確認與強證據**。假說與未解一律去
  [`00-remake-knowledge-gaps.md`](00-remake-knowledge-gaps.md) 查，
  不要因為「總表沒寫」就當成不存在。
- 表裡的位址全部以 `wl.merged.exe` 為準。**打包版 `wl.exe` 的位址不可引用**。
- 斷言被推翻時，**改本表的正文**，把推翻紀錄寫進 `CONTEXT.md` 的
  「已被推翻的斷言」，不要在這裡留歷史。
