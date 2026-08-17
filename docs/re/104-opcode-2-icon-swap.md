# 104：opcode 2 是把兩張圖形對調

日期：2026-08-17 ｜ 接 `docs/re/34`（指令表）、`docs/re/04`（overlay slot 表）、
`docs/re/24`（圖磚與疊圖）、`docs/re/102`（走不到的 opcode）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/34` 把 opcode 2 標成 `?`（「用 `+0x03`／`+0x04` 呼叫 overlay 的
`sub_10036`（畫面）」），`docs/re/102` §5 因此把它列成直譯器最後一個洞。

**答案早就寫在 `docs/re/04` 的 slot 表裡**：`sub_10036` 是 overlay 的第 18 格。

---

## 0. 這一輪最該記住的事

> **兩份筆記各記了一半，而沒有人把它們對起來。**
> `docs/re/34` 只寫「呼叫 `sub_10036`」，`docs/re/04` 只寫「slot 18 交換兩張圖磚」，
> 兩份都是對的，中間差的是「`sub_10036` ＝ slot 18」這一步——
> 而那一步只要拿 `0x10036` 去 §1 的 thunk 表上查一次就有。
>
> **看到「呼叫 overlay 的 `sub_100xx`」時，先算 `(位址 − 0x10000) / 3` 得到 slot 編號，
> 再查 `docs/re/04`。** 那張表 26 格全部有語意。

## 1. `sub_10036` ＝ overlay slot 18

overlay 的入口是 `0x10000` 起、每格 3 bytes 的 `jmp` thunk 表（`docs/re/04` §1）：

```
0x10036  e9210c    jmp sub_10C5A          ; (0x36 ÷ 3) ＝ 18
```

呼叫端（`0x1A515`）：

```
0x1A515  bl ← 3；al ← [記錄 +0x03]
0x1A51D  bl ← 4；cl ← [記錄 +0x04]
0x1A521  call sub_10036
0x1A524  clc；retn                        ; ← 腳本繼續跑
```

## 2. `sub_10C5A(al, cl)` 做兩件事，兩件都是 `xchg`

```
0x10C5A  push ds
0x10C5B  ah ← al；ch ← cl                 ; 參數 a、b
0x10C5F  dl ← ah；dh ← ch                 ; 留一份原值
0x10C63  al ← 0；cl ← 0                   ; ax ＝ a × 256、cx ＝ b × 256
0x10C67  shr ax, 1；shr cx, 1             ; ×128
0x10C6B  ax += 420h；cx += 420h
0x10C72  bx ← ax
0x10C74  ax ← ds:9165h；ds ← ax           ; ← 切到 seg003
0x10C79  si ← cx；cx ← 80h                ; 128 bytes
loc_10C7E:
         al ← [bx]；xchg al, [si]；[bx] ← al ; **對調**
         si++；bx++；loop
0x10C88  al ← dl；cmp al, 0Ah
0x10C8C  jb  loc_10C95                    ; a < 10 → 用 a
0x10C8E  cmp dh, 0Ah
0x10C91  jnb loc_10CB0                    ; b ≥ 10 → 兩個都 ≥ 10，遮罩不動
0x10C93  al ← dh                          ; 否則用 b
loc_10C95:
0x10C97  cl ← 5；shl ax, cl               ; ×32
0x10C9B  ax += 0DA60h
0x10CA0  si ← 0DF60h；cx ← 20h            ; 32 bytes
loc_10CA6:
         al ← [bx]；xchg al, [si]；[bx] ← al ; **對調**
0x10CB0  pop ds；bx ← 0；dx ← 0；retn
```

兩個位址都在 `seg003`（`ds:9165h` 存的段值，`docs/re/14` §2）：

| 位址 | 是什麼 |
|---|---|
| `seg003:0x420 ＋ n × 128` | **疊圖與圖磚連續的那一張表**：0–9 是 `IC0_9.WLF`，≥10 是圖磚（`docs/re/24` §2.3）|
| `seg003:0xDA60 ＋ n × 32` | `MASKS.WLF` 的第 n 張遮罩（`docs/re/24` §2.3）|
| `seg003:0xDF60` | 遮罩表的**第 40 格**（`0xDA60 ＋ 40 × 32`），映像裡**全 0** |

所以：

> **opcode 2 ＝ 把編號 a 與 b 的圖形對調**，
> 並把兩者之中編號 < 10 的那一張（a 優先）的**遮罩與第 40 格對調**。

遮罩全 0 的意思是「一格背景都不留」——合成式是
`螢幕 ← (背景 AND 遮罩) OR 疊圖`（`docs/re/00-master-index` §）。
所以換上去之後那個編號會畫成**不透明**。

推論等級：**已確認**（thunk 表、兩段 `xchg` 迴圈、三個位址與 `0xDF60` 的內容
都直讀；`0xDF60` 的 32 bytes 從映像倒出來確認是全 0）。

## 3. 出貨資料怎麼用它

六筆記錄（全部 0 格），`+0x03`／`+0x04` 是：

| 資源 | 記錄 | a | b |
|---:|---:|---:|---:|
| 0 | 4 | `07` | `2C` |
| 0 | 5 | `07` | `2C` |
| 26 | 0 | `07` | `5F` |
| 26 | 1 | `5F` | `07` |
| 26 | 2 | `07` | `5D` |
| 26 | 3 | `5D` | `07` |

**編號 7 是隊伍的疊圖**（`docs/re/48`：0 塗黑、1–4／6 敵人、5 寶箱、
**7 隊伍**、8 輻射區、9 其他分隊）。`0x2C`／`0x5D`／`0x5F` 都 ≥ 10，
是圖磚 34／83／85。

資源 26 那四筆兩兩成對（`07↔5F`、`5F↔07`、`07↔5D`、`5D↔07`），
而 `xchg` 是對稱的——**參數對調再跑一次就換回來**。

所以這個機制是：**隊伍在地圖上變成一塊地形**（疊圖換掉、遮罩換成全 0
所以畫成不透明），再跑一次換回來。

推論等級：**已確認**（機制與參數）／**強證據**（「變裝」這個用途：
編號 7 的語意、參數成對、遮罩換成不透明三件事互相印證，
但**沒有任何格子指到這六筆記錄**，所以遊戲裡看不到）。

## 4. remake 接上去的樣子

規則層只回報編號（`ScriptResult.Swap`），動圖是呈現層的事——
與音效、倒數同一條規矩。`render.Graphics.Swap(a, b)` 做兩件事：

1. 在「疊圖 ＋ 圖磚」的連續索引空間上對調兩張圖。
2. 編號 < 10 的那一張的遮罩與暫存格對調；暫存格初值是**全 false**（＝ 不透明）。

⚠ **編號超出範圍要報錯不吞。** 那代表資料或索引算錯了，
畫成空白會把問題藏起來（`docs/spec/01` §3）。

## 5. 順帶的結果：44 種 opcode 全部實作完

`TestScriptOpcodeCoverage` 的「未實作」現在只剩四個**索引越界值**
（1282／2271／26478／29813，共 5 筆記錄）——那是
「這一筆記錄根本不是腳本」的訊號（`docs/re/76` §3），本來就該擋掉。

門檻降到底了：`missRecords ≤ 5`、`len(unhandled) ≤ 4`，
再加一條「落進來的不准是合法編號」。

## 6. 可重跑的完整指令

```bash
tools/go.sh test ./internal/game/ -run TestOpOverlayReportsIconSwap -v
tools/go.sh test ./internal/render/ -run TestGraphicsSwap -v
tools/go.sh test ./internal/play/ -run TestScriptOpcodeCoverage -v
```

`0x10000`–`0x10070`、`0x10C5A`、`0x1A515` 的逐指令來源是
`workplace/analysis/dumps/listing.json`，不必再進 IDA。

`seg003:0xDF60` 那 32 bytes（線性 `0x38D80`、檔案位移 `0x28E30`）：

```bash
python3 - <<'EOF'
data = open('workplace/analysis/unpacked/wl.merged.exe','rb').read()
print(data[0x28E30:0x28E50].hex(' '))
EOF
```

## 7. 這一輪學到的（寫成規則）

- **「語意未解」要先查自己的其他筆記。** 這一格掛了兩份筆記的 `?`，
  而答案在第三份裡躺著。**下「未解」之前先 grep 那個位址**——
  `0x10036` 在 `docs/re/04` 的表上是一整列。
- **overlay 的 `sub_100xx` 一律先換算 slot 編號**：`(位址 − 0x10000) ÷ 3`。
  那張 thunk 表每格固定 3 bytes，換算是機械的。
- **看到成對出現的參數（`a,b` 與 `b,a`）先想「這是開關」。** `xchg` 型的操作
  沒有「還原」指令，還原就是再做一次；資料裡成對出現就是在用這個性質。
- **暫存格的初值要去映像裡看，不要假設。** `0xDF60` 全 0 這件事決定了
  變裝之後畫成不透明還是透明——猜錯的話畫面完全不同，而程式碼本身看不出來。
