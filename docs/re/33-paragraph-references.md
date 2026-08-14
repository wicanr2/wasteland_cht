# 33：段落編號在遊戲裡怎麼出現

日期：2026-08-15 ｜ 對應盤點 **E3**（段落編號顯示）

輸入：`game1`／`game2`（SHA-256 見 `docs/re/01`）、`wl.merged.exe`。

---

## 1. 結論：沒有「段落機制」，只有一句話

原版**沒有**段落編號的資料結構、對照表或專屬處理函式。
`Read paragraph 23.` 就是一條普通的敘述字串，和「你看到一扇門」用同一張表、
同一支印字串的程式碼（`docs/re/18`）。

- 42 個地圖區塊的 4,401 條敘述裡，**83 處**寫著 `Read paragraph N.`，
  涵蓋 **82 個不同編號**（只有第 124 段被引用兩次）。
- 執行檔那九張字串表**一次都沒有**。
- 有些是整條就這一句，有些混在句子中間：
  `"There is a diary under the console! It begins just after the Holocaust. Read paragraph 6."`

所以**中文化是純文字層的事**：把那一句換成該段落的中文正文，或換成「翻開手札第 N 則」
之類的導引。不需要在引擎裡做任何段落相關的機制（`CLAUDE.md` §3.2 本來就規定
防拷不移植）。

## 2. 唯一的例外：編號由執行期拼出來

`game2` 資源 15 的字串 61 是 `\rRead paragraph `——**沒有編號**。
同一張表裡有 `'1'`–`'8'`、`'Y'` 這類單字元字串（槽 50–59、64），
而該區塊有一個選單：

```
Select info file:
  1) Col. John Smith.
  2) Maj. Harrison Edsel.
  3) Maj. Peregrine Cite.
  …
```

也就是**選了哪一項，就接上哪一個數字**。中文化時這一處不能只翻譯整句，
要保留「前綴 ＋ 數字」兩段的拼接（或改寫成不需要拼接的句型）。

## 3. 用遊戲資料回頭驗證段落書的分層

`docs/paragraphs/README.md` 把 162 段分成三層（陷阱／變體組／火星誘餌），
當時的依據只有段落原文互相對照，是**二手推論**。現在有一手資料可以對：

| 分層 | 段數 | 被遊戲引用的情況 |
|---|---:|---|
| **陷阱**（1、22、145） | 3 | **一段都沒有** ✓ |
| **變體組** | 31 組 68 段 | 25 組**剛好一段**被引用；3 組完全沒有；3 組有兩段以上 |
| **火星誘餌** | 26 | 只有第 42 段被引用 |
| 其餘 | 94 | 51 段被引用 |

**陷阱段落零引用**，正好對上「這段只存在於紙本、用來抓沒買正版的人」的定位——
遊戲永遠不會叫玩家去讀它。變體組 25/31 剛好一段被引用，也支持
「同一個場景印了好幾份、只有一份是這一版在用的」這個說法。

三組例外（`[14,42]`、`[34,146]`、`[53,88,132]` 兩段以上被引用；
`[19,63]`、`[36,70]`、`[74,78]` 一段都沒有）要回頭逐段看原文——
可能是變體判定過寬，也可能同一個場景真的會依狀態印不同編號。

推論等級：**已確認**（引用清單直接來自遊戲資料）／變體組的解釋仍是**強證據**。

## 4. 密語是遊戲內機制，不是防拷

段落書的另一個用途是給密語。這些密語**在遊戲資料裡就有明文**：

```
"…their general HQ, probably a hideout, uses a password THANATOS."
"We have learned that the password into the courthouse is MUERTE."
"The password is now KAPUT."
```

守衛會問（`"Gimme the password."`、`"Halt" shouts a voice from above.
"What is the password?"`），玩家要打字。**這是遊戲謎題，不是複製保護**——
`CLAUDE.md` 說不移植的是防拷，密語要留。

⚠ **中文化決策未定**：玩家要輸入的是英文單字。輸入層改成中文、或保留英文輸入
但把提示中文化，等輸入比對的程式碼讀出來再定（還沒定位）。

## 5. 可重跑的完整指令

```bash
python3 tools/decode_block_text.py workplace/analysis/unpacked/wl.merged.exe \
  workplace/orig/wastland/game1 workplace/orig/wastland/game2 /tmp/blocks.json

python3 - <<'PY'
import json, re
d = json.load(open('/tmp/blocks.json'))
hits = {}
for b in d['blocks']:
    for i, s in enumerate(b['strings']):
        for m in re.finditer(r'[Rr]ead paragraph\s+(\d+)', s or ''):
            hits.setdefault(int(m.group(1)), []).append((b['file'], b['resource_id'], i))
print(len(hits), '個編號、', sum(map(len, hits.values())), '處引用')
print('沒被引用：', [n for n in range(1, 163) if n not in hits])
PY
```

## 6. 這一輪學到的（寫成規則）

- **「這個機制在哪」有時候的答案是「沒有機制」。** 段落編號找了很久的處理函式，
  結果它從頭到尾只是敘述文字的一部分。**先在資料裡搜字面**，再去找程式碼——
  搜到 83 處明文只花了一次解碼，而追程式碼會一直撲空。
- **二手推論可以用一手資料回頭驗。** 段落書的分層當初是靠原文互相對照推的；
  遊戲資料裡的引用清單是獨立來源，兩邊對上才把「陷阱段落」這件事釘死。
