# 02：`wl.exe` 是 EXEPACK 打包的，解包後才是真正的分析對象

日期：2026-08-13 ｜ 狀態：格式與解包過程已確認；「解包映像等同原版執行行為」目前是**強證據**（見 §5）

## 1. 打包器身分

`wl.exe` 映像尾端 405 bytes 是解壓 stub。stub 開頭 18 bytes 是 EXEPACK header，
偏移 0x10 處的 `RB` 簽章（檔案偏移 62,160）確認打包器是 **Microsoft EXEPACK**。

```
檔案 0x00F2C0：b6 10 00 00 00 00 95 01 00 02 00 29 56 29 01 00
檔案 0x00F2D0：52 42 8c c0 05 10 00 0e ...        ← 'R' 'B'
```

| EXEPACK 欄位 | 值 | 意義 |
|---|---|---|
| `real_ip` | `0x10B6` | 解包後真正的進入點 offset |
| `real_cs` | `0x0000` | 進入點 segment |
| `exepack_size` | 405 | stub 長度，**與映像剩餘長度完全相同** |
| `real_sp` / `real_ss` | `0x0200` / `0x2900` | 解包後的堆疊 |
| `dest_len` | 10,582 段落 | 解包後映像 169,312 bytes |
| `skip_len` | 1 段落 | |
| 簽章 | `RB` | |

## 2. 解包器

`tools/unpack_exepack.py`，純 Python（stdlib），依 EXEPACK 格式本身實作，不依賴外部解包工具。
選擇自己寫的理由：每一步都要能對原始 bytes 交代，而且失敗時要明確報錯，
不能產出一份「看起來像樣、其實錯位」的映像。

演算法：從壓縮資料尾端往前解，每個命令是 `[資料][長度:word][命令:byte]`。

| 命令 | 動作 |
|---|---|
| `0xB0` / `0xB1` | fill：讀 1 byte 值，往目的緩衝區反向填 `length` 次 |
| `0xB2` / `0xB3` | copy：從來源反向搬 `length` bytes |
| bit 0 = 1 | 這是最後一個命令 |

解完之後，來源剩下的前段是未經 RLE 的原始資料，直接搬進目的緩衝區前端。

實測統計：249 個命令（126 fill、123 copy）、尾端 15 bytes 的 `0xFF` 填充、
78 bytes 的字面前綴，61,632 bytes 壓縮資料還原成 169,312 bytes（壓縮比 0.364）。

## 3. ⚠ relocation table 的起點：`corrupt#` 的 `#` 不是訊息的一部分

stub 裡的錯誤訊息在 IDA 顯示成 `Packed file is corrupt#`，看起來像是訊息帶了個 `#` 收尾、
後面再跟一個 NUL。**這是誤導。**

實際上訊息只有 `Packed file is corrupt`（22 bytes），緊接著就是 relocation table，
而 `#` = `0x23` = 35 正是**第一組 relocation 的 count 低位元組**，後面的 `0x00` 是它的高位元組。

把 `#\0` 當成訊息的一部分往後跳兩個 byte，整張表就會錯開，讀出來的 count 變成 `0x10B9`
這種不可能的大數——第一次實作正是這樣失敗的。

正確解析（16 組，segment 高位 `0x0000`…`0xF000`，每組 `[count:word][offsets…]`）：

| 群組 | segment base | count |
|---|---|---:|
| 0 | `0x0000` | 35 |
| 1 | `0x1000` | 1 |
| 2–15 | `0x2000`…`0xF000` | 0 |

合計 36 筆 relocation，表本身 104 bytes ＝ 16 個 count（32）＋ 36 個 offset（72）。
**這個數字剛好用完 stub 尾端到檔尾的空間，是解析起點正確的算術證據**；
解包器也把「表之後還有非零資料」列為失敗條件，錯位會直接報錯而不是默默產出。

## 4. 輸出

| 項目 | 值 |
|---|---|
| 檔案 | `workplace/analysis/unpacked/wl.unpacked.exe`（gitignore） |
| 大小 | 169,488 bytes（header 176 ＋ 映像 169,312） |
| SHA-256 | `b5eb39f094e0274165eab5e1584e78ff5b54c7228d8db273573d2bd951ea31a0` |
| 進入點 | `0000:10B6` |
| 堆疊 | `2900:0200` |
| relocation | 36 筆（重建進新 MZ header） |
| 報告 | `workplace/analysis/unpacked/unpack-report.json` |

## 5. 解包正確性的證據與界線

**支持解包正確的四項證據：**

1. 長度自洽：`dest_len` 宣告 169,312 bytes，解包實際產出 169,312 bytes，
   RLE 命令流以「最後一個命令」旗標正常結束，沒有剩餘資料。
2. relocation table 剛好用完 stub 空間（§3），且表後全為 0。
3. IDA 對解包映像的自動分析：函式從 340 → **614**，
   **沒有 caller 的函式從 22 → 2**。壓縮資料被誤判成程式碼時不會有這種收斂。
4. 進入點對得上：EXEPACK header 宣告 `0000:10B6`，
   IDA 在解包映像的 `start` 落在 `0x110B6`（載入基址 `0x10000` ＋ `0x10B6`）。
   解包前完全看不到的字串也出現了，例如資料檔名表
   `ALLPICS1`／`ALLPICS2`／`ALLHTDS1`／`ALLHTDS2`／`END.CPA`／`WLA.BIN`／
   `COLORF.FNT`／`IC0_9.WLF`／`MASKS.WLF`／`TRANSTBL`／`TITLE.PIC`。

**還沒做的驗證**：解包映像**尚未實際執行過**。要把「解包映像的行為等同原版」從
強證據升為已確認，需要在 DOSBox 裡跑起來，並與原版 `wl.exe` 做同輸入的行為對照。
在那之前，任何依賴「執行時期行為」的結論都要標明這個前提。

## 6. 解包後的基準資料庫

`workplace/analysis/ida94/wl.unpacked.i64`（gitignore）。

| 項目 | 值 |
|---|---:|
| segments | 6 |
| entry points | 1（`start` ＠ `0x110B6`） |
| 自動辨識 functions | 614 |
| IDA 認定 strings | 43 |
| 沒有直接 caller 的 functions | 2 |

| segment | class | 範圍 | 大小 |
|---|---|---|---:|
| `seg000` | CODE | `0x10000`–`0x1CB67` | 52,071 |
| `seg001` | CODE | `0x1CB67`–`0x1CE20` | 697 |
| `seg002` | UNK | `0x1CE20`–`0x2AE20` | 57,344 |
| `seg003` | UNK | `0x2AE20`–`0x39000` | 57,824 |
| `seg004` | STACK | `0x39000`–`0x39200` | 512 |
| `seg005` | UNK | `0x39200`–`0x39560` | 864 |

匯出：`docs/re/generated/ida94/inventory-unpacked.json` ／ `inventory-unpacked.md`。

**43 筆字串仍然偏少**，而遊戲顯然有大量文字。合理的方向是文字在 `game1`／`game2`
或經過編碼；`wl.exe` 裡兩串看起來像編碼表的字串（`0x1DD98` 的 ` etraoishlndgcyupmb,`
與 `0x1C01B` 的 `lcpmuhywfbv:8.91g!q6-'=`）是下一步的入口。**目前這只是待驗證的方向，不是結論。**

## 7. 重跑方式

```sh
docker run --rm --network none --memory 1g --cpus 1 --pids-limit 256 \
  --user "$(id -u):$(id -g)" -v "$PWD:/workspace" -w /workspace \
  ida-pro-9.4-idapython:py312-v1 \
  /opt/venv/bin/python3 tools/unpack_exepack.py \
    workplace/orig/wastland/wl.exe \
    workplace/analysis/unpacked/wl.unpacked.exe \
    workplace/analysis/unpacked/unpack-report.json

WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.unpacked.exe" tools/ida.sh build
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.unpacked.exe" \
  tools/ida.sh run tools/ida/export_inventory.py \
    docs/re/generated/ida94/inventory-unpacked.json
```
