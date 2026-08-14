# 08：MSQ 區塊的加密解開了 —— 42/42 通過原版自己的驗證

日期：2026-08-13 ｜ 狀態：**已確認**（42 個區塊全部通過原版的 checksum 條件，並解出可讀遊戲文字）

## 1. 演算法（`sub_11A59`，`0x11A59`，74 bytes）

```asm
0x11A59  mov  ax, ds:9176h      ; 從檔案讀進來的 checksum
0x11A5C  xor  al, ah            ; key ＝ 低位元組 XOR 高位元組
0x11A5E  mov  cx, ds:46B0h
0x11A62  inc  cx / inc cx       ; 第一段長度 ＝ 緩衝區大小 ＋ 2
0x11A64  mov  si, 0
0x11A67  xor  [si], al          ; ── 解密迴圈
0x11A69  add  al, 1Fh           ;    每個 byte 之後 key ＋= 0x1F
0x11A6B  inc  si
0x11A6C  loop loc_11A67
0x11A6E  dec  si / dec si
0x11A70  mov  cx, [si]          ; 剛解出來的最後 2 bytes ＝ 總長度
0x11A73  sub  cx, ds:46B0h      ; 扣掉已解的部分
0x11A7B  xor  [si], al          ; ── 用同一條 key 序列接著解剩下的
0x11A7D  add  al, 1Fh
0x11A80  loop loc_11A7B
; ── 驗證 ──
0x11A8A  mov  al, [si]
0x11A8C  sub  bx, ax            ; bx ＝ 全部 byte 的負和
0x11A8F  loop loc_11A8A
0x11A91  cmp  bx, ds:9176h      ; 要等於 checksum
0x11A95  jnz  short loc_11A9D   ; 不等 → CF=1（失敗）
```

化成程式：

```python
key = (checksum & 0xFF) ^ (checksum >> 8)
for i, c in enumerate(cipher):
    plain[i] = c ^ key
    key = (key + 0x1F) & 0xFF
assert (-sum(plain)) & 0xFFFF == checksum
```

分兩段只是因為「總長度」寫在第一段末尾的 2 bytes（第一段長度由緩衝區大小 `ds:46B0h` 決定），
key 序列在兩段之間是**連續的**，不重設。

## 2. 區塊佈局

| 位移 | 大小 | 內容 |
|---|---:|---|
| `+0` | 4 | magic `msq0`（`game1`）／`msq1`（`game2`） |
| `+4` | 2 | checksum（同時是金鑰種子） |
| `+6` | 其餘 | 加密資料 |

載入器 `sub_1841F` 的動作與這個佈局完全對得上：
`cx = 2, dx = 0x9176` 讀 checksum，接著 `cx = 表值 − 6` 讀本體
（`−6` 正好是 magic 4 ＋ checksum 2）。

## 3. 區塊長度表：`ds:BD86h`

**兩個資料檔共用同一張表**，依資源編號（0–49）排列，順序與目錄一致：

```
10958 5739 9600 10097 8188 4530 10140 | 4320 | 7270 6933 9748 | 5278 …
└──────────── game1 #0–#6 ───────────┘ └game2#0┘ └ game1#7–9 ┘ └game2#1
```

對照目錄的前八項 `80 81 82 83 84 85 86 40`，逐項吻合：
高 2 bits 決定去哪個檔案取，長度依序取用。

第 15 項（資源編號 14）的長度是 **0**，而目錄那一項正是 `0x0E`（高 2 bits ＝ `0x00`）。
這印證了 `docs/re/07` §3 的解讀：**`0x00` 表示這個資源編號沒有對應的 MSQ 區塊**。

另一張表 `ds:BD22h`（8583、4997、9223…）是載入器實際要求的讀取量。
三個長度的關係在 `docs/re/18` §3 解開：加密長度 < 讀取量 < 區塊總長度，
中間夾的分別是字串表與 Huffman 尾段。

## 4. 驗證結果

`tools/decrypt_msq.py` 對全部 42 個區塊執行：

| 檢查 | 結果 |
|---|---|
| magic 是 `msq0`／`msq1` | **42／42** |
| 解密後 `-(sum) == checksum` | **42／42** |

checksum 是 16-bit，單一區塊碰巧通過的機率是 1/65536；42 個全部通過，
再加上長度表與目錄的獨立吻合，**這套理解可以當已確認用**。

解密後也直接看得到遊戲內容，不再是高熵亂數：

| 區塊 | 英文單字數 | 抽樣 |
|---|---:|---|
| `game2` #16 | 53 | `Titanium`、`Vanadium`、`Clawer`、`Vulture`、`Chopter` |
| `game2` #35 | 86 | `Brother`、字母序列 `OPQRSTUVWXWYZ`、`abcdefghijklmno` |
| `game2` #36 | 76 | `ETech`、`section`、`REDHAWK` |
| `game1` #0 | 90 | 大量重複的 `f`（地圖填充，不是文字） |

`Titanium`／`Vanadium`／`Vulture` 這些是遊戲裡的裝備與怪物名，
`REDHAWK` 是劇情裡的地名——內容確實是遊戲資料，解密方向正確。

⚠ 但**大部分區塊解密後仍然沒有可讀文字**。合理的解讀是：
解密只是第一層，敘述文字另有編碼（原版手冊提到遊戲會即時組出句子），
或者多數區塊本來就是地圖與數值表。**目前沒有證據分辨，不要當成「文字都在這幾塊」。**

## 5. 未解與下一步

| 項目 | 狀態 |
|---|---|
| `ds:BD22h` 讀取量與區塊長度為何不同 | **已解**：差額是 Huffman 尾段（`docs/re/12`）。而讀取量與**加密長度**的差額是字串表（`docs/re/18` §3）|
| 解密後的區塊內部結構（哪裡是地圖、哪裡是文字、哪裡是數值表） | 未解——下一步主線 |
| 敘述文字的第二層編碼 | 未證實其存在，但多數區塊無可讀文字這件事需要解釋 |
| 緩衝區大小 `ds:46B0h`（`0x1800`／`0x600`）由什麼決定 | 只知與 `ds:4655h` 及表 `ds:BF1Ch` 有關 |

## 6. 重跑方式

```sh
docker run --rm --network none --memory 1g --cpus 1 --pids-limit 256 \
  --user "$(id -u):$(id -g)" -v "$PWD:/workspace" -w /workspace \
  ida-pro-9.4-idapython:py312-v1 /opt/venv/bin/python3 tools/decrypt_msq.py \
    workplace/analysis/unpacked/wl.unpacked.exe \
    workplace/orig/wastland/game1 workplace/orig/wastland/game2 \
    docs/re/generated/resources/msq-decrypted.json
```
