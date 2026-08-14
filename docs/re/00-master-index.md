# 00：RE 總表

> **這份是逆向結果的速查表。** 要找「某個位址是什麼」「某個格式怎麼解」
> 「某個機制在哪份筆記」時查這裡，不要翻十八份文件。
>
> 三份 `00-*` 各有分工：
> **本表**＝已知的事實速查；
> [`00-remake-knowledge-gaps.md`](00-remake-knowledge-gaps.md)＝還缺什麼；
> [`00-function-index.md`](00-function-index.md)＝641 個函式誰解過。
>
> 最後更新：2026-08-14

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
| `allpics1/2` | Huffman 容器，66／98 個子區塊 | 容器已解，內容未解 | [`11`](11-huffman-decoder.md) |
| `allhtds1/2` | Huffman 容器，4／5 個大塊 | 容器已解，內容未解 | [`11`](11-huffman-decoder.md) |
| `end.cpa` | Huffman 容器 | 容器已解 | [`11`](11-huffman-decoder.md) |
| `curs`／`masks.wlf`／`ic0_9.wlf`／`transtbl` | 未加密，載入位址已知 | 內容未解 | [`03`](03-boot-and-asset-loading.md) §5 |
| `info` | 2 bytes 安裝資訊 | 已解 | [`03`](03-boot-and-asset-loading.md) §3 |
| `paragraphs.txt`／`manual.txt` | 明文 | 已轉錄 | `docs/paragraphs/`、`docs/manual/` |

**主文字字型不是獨立檔案**，內嵌在 `wl.exe` 的 `seg003:0xCA60`。

## 3. 演算法速查

| 機制 | 一句話 | 實作 | 文件 |
|---|---|---|---|
| EXEPACK 解包 | 反向 RLE ＋ relocation 重建 | `tools/unpack_exepack.py` | [`02`](02-exepack-unpack.md) |
| MSQ 解密 | `key = lo(cs) ^ hi(cs)`；逐 byte XOR 後 `key += 0x1F` | `tools/decrypt_msq.py` | [`08`](08-msq-encryption.md) |
| **加密長度** | **＝ 區塊標頭第一個 word，不是整個區塊** | `tools/decode_block_text.py` | [`18`](18-block-text.md) §2 |
| Huffman | 前序編碼的樹 ＋ 位元流，無 magic 的尾段也是同一套 | `tools/huffman.py` | [`10`](10-huffman-compression.md)、[`11`](11-huffman-decoder.md) |
| 文字打包 | 5-bit 符號 ＋ 60 bytes 字元對照表；`0x1E` 轉大寫、`0x1F` escape | `tools/decode_text.py` | [`17`](17-packed-text.md) |
| 亂數 | 五個 byte 的進位鏈，無乘除，初值全零，熵來自鍵盤輪詢次數 | `tools/rng.py` | [`13`](13-rng.md) |
| 擲骰 | dN ＝ 遮罩 ＋ 拒絕取樣，回傳 1..N 等機率 | `tools/rng.py` | [`13`](13-rng.md) §3 |
| 字型繪製 | 主文字 8 bytes 單色；彩色字型 32 bytes、4 平面連續存放 | `tools/dump_font.py` | [`14`](14-fonts-and-text-encoding.md) |

## 4. 資料結構

### 4.1 MSQ 區塊（`game1`／`game2` 的一個資源）

```
+0x0000                地圖：一個 byte 兩格（4 bits），長度 P
+P                     記錄區標頭（0x5C bytes），第一個 word ＝ 加密長度 L
+P+0x5C                各 section
+L                     字串表（不加密）
+（讀取量 − 6）         Huffman 尾段：tile 圖形，解出 4,096 或 1,024 bytes
+（總長度 − 6）         結束
```

| P | 地圖尺寸 | 區塊數 | 判定 |
|---|---|---:|---|
| `0x600` | 64 × 48 | 38 | `ds:BF1Ch[資源編號] != 0x40` |
| `0x1800` | 128 × 96 | 4 | `ds:BF1Ch[資源編號] == 0x40` |

**記錄區標頭（P 起算）**

| 位移 | 內容 |
|---|---|
| `+0x00` | 加密長度 L ／ 同時是字串表基址 |
| `+0x02` | 明文 NUL 分隔的敵人名表位址 |
| `+0x04` | 一張 8 bytes 一筆的記錄表位址 |
| `+0x06`…`+0x28` | section 位移表（型別 → 位移要查執行檔的 `ds:B9E0h`） |
| `+0x2F` | 遭遇出現機率的分母 |
| `+0x31` | 遭遇種類數 |
| `+0x32` | 遭遇槽位上限 |
| `+0x33` | 未解 |

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

### 4.3 角色記錄（256 bytes）

定址：`0x7131 + 角色編號 × 256`，經隊伍槽表兩層間接
（`sub_17208` → `sub_19614`）。詳見 [`15`](15-character-record.md)。

| 位移 | 大小 | 語意 | 等級 |
|---|---|---|---|
| `+0x00` | 字串 | 名字（NUL 結尾） | 已確認 |
| `+0x15`–`+0x17` | 24-bit | 可花用的計數器（比大小 ＋ 借位扣，語意像金錢） | 強證據 |
| `+0x21`–`+0x23` | 24-bit | 只加不減的計數器，語意未解 | 強證據（結構） |
| `+0x26`–`+0x27` | 16-bit | 受傷前的 CON 備份 | 強證據 |
| `+0x18` | byte | 性別（0 ＝ Male、1 ＝ Female） | 已確認 |
| `+0x19` | byte | 國籍（0–4 ＝ U.S./Russian/Mexican/Indian/Chinese） | 已確認 |
| `+0x1A` | byte | AC（護甲等級） | 強證據 |
| `+0x1B`–`+0x1C` | 16-bit | MAXCON（最大體力） | 強證據 |
| `+0x1D`–`+0x1E` | 16-bit 有號 | CON（目前體力，可為負） | 強證據 |
| `+0x1F` | byte | 裝備索引（`×2 + 0xBC` 指進 `+0xBD` 那個陣列） | 強證據 |
| `+0x28` | byte | 非 0 → 名片行反白 | 假說 |
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

### 4.5 一次攻擊（敵方打隊伍）

```
挑目標   roll(1..隊伍人數)，不合格重抽
門檻     基礎(40/50/60) ＋ 被攻擊者技能×3 ＋ 欄位 ± 距離修正，夾在 100
命中     roll(1..100) ≥ 門檻          ← 大於等於才命中，注意 cmc
傷害     (武器資料 +0x05 高 4 位) ＋ (武器資料 +0x03) 顆 d6
護甲     吸收 ＝ 被攻擊者 AC 顆 d6 的和
結算     CON −= (傷害 − 吸收)
```

詳見 [`20`](20-combat-resolution.md)、[`19`](19-effects-and-damage.md)。

### 4.6 字串表

```
表基址 +0x00 … +0x3B    60 bytes 字元對照表（符號 → ASCII，0 ＝ 結束）
表基址 +0x3C …          16-bit 位移表，每 4 個字串一項，位移相對於 +0x3C
```

取第 N 個字串：跳到第 `N >> 2` 項位移，再解掉 `N & 3` 個。
**每張表有自己的字元對照表**，不能共用。

| 來源 | 張數 | 字串數 |
|---|---:|---:|
| 執行檔（常數基址，見 §5.1） | 9 | 442 |
| 42 個 MSQ 區塊（各一張） | 42 | 4,493 |
| **合計** | **51** | **4,935** |

字串內的機制：`\n` 分隔字根／單數字尾／複數字尾；`0x0B` 插入角色名字；
`0x0C` 夾 his/her 做性別選字；`0x0D` 段內換行。

## 5. 位址表

### 5.1 執行檔內的表（`ds:` 位移，線性 ＝ ＋`0x1CE20`）

| `ds:` | 線性 | 內容 |
|---|---|---|
| `0xA703` | `0x27523` | 字串表：開場字幕與製作名單 |
| `0xAA60`／`0xAA6D`／`0xAA7A` | `0x27880`… | 遭遇生成用的三張 13 項表 |
| `0xAB3E` | `0x2795E` | 字串表：無線電、隊伍、戰鬥 |
| `0xB233` | `0x28053` | 傷勢等級 → 訊息碼（`85 9A 9B 9C 9D 84`） |
| `0xB239` | `0x28059` | 隊伍槽表位移（`00 0E 1C 2A`，間隔 14） |
| `0xB270` | `0x28090` | 字串表：技能、物品、介面（170 條） |
| `0xB9E0` | `0x28800` | **section 型別 → 記錄區標頭位移**（24 項） |
| `0xBD22` | `0x28B42` | 各資源的讀取量 |
| `0xBD86` | `0x28BA6` | 各資源的區塊總長度 |
| `0xBEC9` | `0x28CE9` | 資源目錄（50 項，`0xFF` 結束；高 2 bits ＝ 哪個檔案） |
| `0xBF1C` | `0x28D3C` | 地圖尺寸選擇表（`0x40` → `0x1800`） |
| `0xCCCE` | `0x29AEE` | 傷勢門檻（`F5 EC E2 D8` ＝ −11／−20／−30／−40） |
| `0xCD50` | `0x29B70` | 文字控制碼跳表（直繪版，18 項） |
| `0xCD74` | `0x29B94` | 文字控制碼跳表（組行版，18 項） |
| `0xCE4B` | `0x29C6B` | 字串表：角色建立 |
| `0xD18E` | `0x29FAE` | 字串表：**結局敘述** |
| `0xD622` | `0x2A442` | 字串表：階級（64 條） |
| `0xDACC` | `0x2A8EC` | 字串表：技能學習 |
| `0xDBF8` | `0x2AA18` | 字串表：商店 |
| `0xDCED` | `0x2AB0D` | 字串表：疾病與狀態 |

### 5.2 `seg003` 的素材位址（線性 ＝ ＋`0x2AE20`）

| `seg003:` | 線性 | 內容 |
|---|---|---|
| `0x0100` | `0x2AF20` | `TRANSTBL`（800 bytes，用途未解） |
| `0x0420` | `0x2B240` | `IC0_9.WLF`（1,280 bytes） |
| `0x0920` | `0x2B740` | `TITLE.PIC`（18,432 bytes） |
| `0xB4E0` | `0x36300` | `COLORF.FNT`（5,504 bytes ＝ 172 × 32） |
| `0xCA60` | `0x37880` | **主文字字型**（內嵌，128 × 8 bytes） |
| `0xDA60` | `0x38880` | `MASKS.WLF` |

`CURS` 載到 `seg002:0x7E0B`。

### 5.3 執行期變數（`ds:` 位移）

| `ds:` | 內容 |
|---|---|
| `0x4654` | 目前隊伍組別 |
| `0x4655` | 目前資源（地圖）編號 |
| `0x465C`–`0x4660` | **RNG 狀態**（5 bytes，映像初值全零） |
| `0x4661` | 目前 section 的起點 |
| `0x4667` | 最後一個按鍵碼（小寫已轉大寫） |
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
| `0x46C8` | 另一組記錄的位址（`0x6B31 + 0x178a + 0x5Eb`） |
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

完整索引在 [`00-function-index.md`](00-function-index.md)（641 個，已分析 177）。
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
| `0x19D86` | 3 | `base + Nd6` | [`13`](13-rng.md) §3.3 |
| `0x19C84` | 11 | 2d6 逢同點續擲 | [`13`](13-rng.md) §3.4 |
| `0x14193` | 3 | **讀效果記錄**（`+0x08`／`+0x09`）並算出值 | [`19`](19-effects-and-damage.md) §2 |
| `0x141FA` | 2 | **套用效果**到角色欄位（地圖事件的唯一出口） | [`19`](19-effects-and-damage.md) §3 |
| `0x157D6` | 3 | **傷害套用**（護甲擲骰、扣 CON、傷勢） | [`19`](19-effects-and-damage.md) §4 |
| `0x1B108` | 2 | **累加命中門檻**（飽和加法，夾在 100） | [`20`](20-combat-resolution.md) §3 |
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
| `0x18744` | — | 存檔載入（兩份輪替，比 32-bit 序號） | [`09`](09-msq-map-structure.md) §4 |
| `0x1CB75` | — | 掛 `int 08h`，重設 8253 channel 0 | [`04`](04-overlay-wla-bin.md) |
| `0x1CC76` | — | PC 喇叭發聲（`out 42h` ＋ `out 61h`） | 盤點 F2 |
| `0x1CD52` | — | 計時器驅動的音效位元組碼直譯器 | 盤點 F2 |

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
| `tools/rng.py` | 亂數與擲骰的參考模型（附自我測試） | 否 |
| `tools/unpack_exepack.py`／`apply_overlay.py` | 解包／合成分析映像 | 否 |
| `tools/gen_func_index.py` | 產生 `00-function-index.md` | 否 |

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
| [`18`](18-block-text.md) | 地圖區塊的 4,493 條敘述文字 |
| [`19`](19-effects-and-damage.md) | 資料驅動的效果系統、傷害與護甲、單複數選擇器 |
| [`20`](20-combat-resolution.md) | 命中判定（d100 對門檻）、武器傷害公式、一次攻擊的完整流程 |

## 9. 引用這份表時的紀律

- **本表只寫已確認與強證據**。假說與未解一律去
  [`00-remake-knowledge-gaps.md`](00-remake-knowledge-gaps.md) 查，
  不要因為「總表沒寫」就當成不存在。
- 表裡的位址全部以 `wl.merged.exe` 為準。**打包版 `wl.exe` 的位址不可引用**。
- 斷言被推翻時，**改本表的正文**，把推翻紀錄寫進 `CONTEXT.md` 的
  「已被推翻的斷言」，不要在這裡留歷史。
