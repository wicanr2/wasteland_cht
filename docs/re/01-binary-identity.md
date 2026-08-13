# 01：分析目標的身分與 `wl.exe` 的第一份資料庫

日期：2026-08-13 ｜ 狀態：已確認（工具輸出）＋ 一項結論（見 §4）

## 1. 輸入來源

原版取自 `Wasteland_1988.zip`（torrentzip，`TORRENTZIPPED-ADA91CE7`），
解到 `workplace/orig/wastland/`。20 個檔案，`wastland/` 目錄名是原樣保留的拼寫。

| 檔案 | bytes | SHA-256 |
|---|---:|---|
| `wl.exe` | 62,549 | `098aef9b4fe4fea3b8d0d134f82fe11a6dac608839ebd175e168cf0271b93b4f` |
| `wla.bin` | 4,209 | `85ab88dd9ab6a67390d6fc6dcba1ba4bfc6801a5d0b9cecf074d263c4d9e0c62` |
| `game1` | 159,429 | `44521eef76fab492d38c0c0bcb171be3f06a4063ce1e0f49175f0b7370a7841d` |
| `game2` | 172,235 | `6de4df6dd1af0a0210c2d6e282ecd1327f2acc7440c955f65045fd6e8a135c47` |
| `allpics1` | 105,866 | `a1bf92f9b037023d99ee4e4bd22547e1fce176608c73494d707676b7d7ce6044` |
| `allpics2` | 133,433 | `b39cb802cfda1deacb5aa575c076bdcb5999bc978794f030d63c28a3642f4a36` |
| `allhtds1` | 34,307 | `c01d0d23d9d073be3a15a8d92fbaed8e9c117ad2fe42e669f8d39b2d6dbc8c42` |
| `allhtds2` | 39,230 | `c0b2d0299d28010db37d0c7714d2879c32ead0f31ba35fdb01c183defe0ec11f` |
| `end.cpa` | 21,027 | `c48197e4b4d9fb81c39ca69e24fd1228cab90647ca6a2e20d809ecccb01ae731` |
| `title.pic` | 18,432 | `5169de0bd7efad4bf12ba8135744d9daeb1939bb7f01bd5ca5126e626d223800` |
| `colorf.fnt` | 5,504 | `db34077fe0bcc331734ba660605f8ffe4da33bf6c97794f4c712286eb0430d51` |
| `ic0_9.wlf` | 1,280 | `d8bbeae054a25852817841b905ae093fb104587e89d90749fb8fd9ec6ca38ddc` |
| `masks.wlf` | 320 | `6f355ad5841f050e2af7f353081f53e2f876b26d6f6970edf25ca8d3c2b17f66` |
| `curs` | 2,048 | `6ffbe904fc19ff8d355146fa5eb51c611e242cfa9f35538643414175d1cbaf7e` |
| `transtbl` | 800 | `1d1a265562bf2f0ecb0bb4add32015e66eae7510b93499e707dfc9c8c36d91ba` |
| `paragraphs.txt` | 72,771 | `ba50b061a0ed4326518bf92fe071e77ea8010d44ddaf889eff434be6b3e8eb92` |
| `manual.txt` | 53,322 | `4b222c7dc22229bffce83455989a1d164182f40396ffcf54a9445c09a2bc342e` |
| `readme.txt` | 164 | `9ff6cc25057bb3affe65e32ab4012909fca55621328b511a49d540f46d3ce54b` |
| `info` | 2 | `7e46dde720f00e74467c313a1142b572a18a5f03561bc08d6633de9a09d9eaa6` |

`readme.txt` 說明這份是 21st Century Oldies 的散布版，並提醒玩家需要
`paragraphs.txt` 才能遊玩（原版的段落書防拷）。這是散布者加的檔，不是 1988 年原始內容。

## 2. `wl.exe` 的 MZ header

| 欄位 | 值 | 說明 |
|---|---|---|
| `e_magic` | `MZ` | |
| `e_cblp` / `e_cp` | 85 / 123 | 映像 62,549 bytes，**與檔案大小相同 → 沒有 overlay** |
| `e_crlc` | 0 | **沒有 relocation entry** |
| `e_cparhdr` | 32 | header 512 bytes |
| `e_minalloc` / `e_maxalloc` | 6,738 / 65,535 | |
| `e_ss:e_sp` | `2970:0080` | |
| `e_cs:e_ip` | `0F0C:0012` | 進入點在映像**最後 405 bytes** 那一段 |
| `e_lfarlc` | 0x1E | |

「0 個 relocation」加上「進入點在映像尾端一小段」是打包執行檔的典型輪廓。

## 3. 第一份 IDA 資料庫（直接分析原檔）

工具：`ida-pro-9.4-idapython:py312-v1`（IDA 9.4、Python 3.12.3），
指令 `tools/ida.sh build`，資料庫 `workplace/analysis/ida94/wl.i64`（gitignore）。

| 項目 | 值 |
|---|---:|
| segments | 3（seg000 CODE、seg001 CODE、seg002 STACK） |
| entry points | 1（`start` ＠ `0x1F0D2` ＝ seg001+0x12） |
| 自動辨識 functions | 340 |
| IDA 認定 strings | 40 |
| 沒有直接 caller 的 functions | 22 |

匯出：`docs/re/generated/ida94/inventory.json` ／ `inventory.md`
（`tools/ida/export_inventory.py` → `tools/summarize_inventory.py`）。

`start` 的內容是十幾行的搬移常式：

```asm
start           proc far
                mov     ax, es
                add     ax, 10h
                push    cs
                pop     ds
                mov     word_1F0C4, ax
                add     ax, word_1F0CC
                mov     es, ax
                mov     cx, word_1F0C6
                mov     di, cx
                dec     di
                mov     si, di
                std
                rep movsb          ; 反向搬移整個映像
                mov     dx, word_1F0CE
                push    ax
                mov     ax, 38h
                push    ax
                retf                ; 跳進搬完之後的解壓器
```

40 筆字串裡有 `Packed file is corrupt`（`0x1F1D7`），與上面這段合起來指向 EXEPACK。
**40 筆字串的 xref 全部是 0**——沒有任何程式碼引用它們，因為引用它們的程式碼還在壓縮狀態。

## 4. 結論：這份資料庫不可作為分析依據

`wl.exe` 是 **Microsoft EXEPACK 打包**的執行檔（驗證見 `02-exepack-unpack.md`）。
IDA 對原檔自動辨識出的 340 個函式，多數是把壓縮資料當成指令解讀的結果；
40 筆字串是碰巧未被壓縮打散的片段，不代表程式的字串表。

因此：

- **不得引用 `wl.i64` 的函式位址、名稱或 xref 當作任何機制的證據。**
  這份資料庫的用途只有一個：證明打包器身分。
- 後續所有逆向以解包後的映像為準，見 `02-exepack-unpack.md`。

## 5. 重跑方式

```sh
unzip -q -o Wasteland_1988.zip -d workplace/orig/
(cd workplace/orig/wastland && sha256sum *)

tools/ida.sh build
tools/ida.sh run tools/ida/export_inventory.py docs/re/generated/ida94/inventory.json

docker run --rm --network none --memory 1g --cpus 1 --pids-limit 256 \
  --user "$(id -u):$(id -g)" -v "$PWD:/workspace" -w /workspace \
  ida-pro-9.4-idapython:py312-v1 \
  /opt/venv/bin/python3 tools/summarize_inventory.py \
    docs/re/generated/ida94/inventory.json \
    docs/re/generated/ida94/inventory.md
```
