# 118：物品表有四份，商店按記錄 `+0x06` 換一份——那一欄是「這家店賣什麼」

日期：2026-08-20 ｜ 接 [`45`](45-item-data-and-weapon-damage.md) §2（物品表在存檔區）、
[`42`](42-facility-loops.md) §2／§4（買、賣與庫存 `+0x02`）、
[`117`](117-save-globals-and-facility-screen.md)（同一輪的設施對拍）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`；
實機截圖 `workplace/dosbox/shots/44-buy.png`、`47-buy2.png`。

高池鎮商店的買清單，原版與 remake 差一件：**remake 多賣一個 `Engine`（$500）**。
價錢、名字、順序其餘全部一致，所以不是解碼錯——是**清單的來源就不是同一張表**。

---

## 1. 檔案裡有四份物品表

`0x185E6` 是物品表的載入器：

```
0x185E6  al ← ds:46C4h            ; 想要的那一組
0x185E9  cmp al, ds:46C5h         ; 已經載進來的是哪一組
0x185ED  相同 → 直接 retn          ; **不重讀**
0x185EF  ds:46C5h ← al
0x185F4  ds:9168h ← 0x80          ; 檔案選 game1
0x185F9  第 4 組 → ds:9168h ← 0x40 ; 第 4 組在 game2
0x1860B  dx:cx ← 0x000253C5       ; game1 的存檔資源位移
0x18618  第 4 組 → 0x00028BC7      ; game2 的
0x1861E  dx:cx ＋= 0x1206          ; 存檔資源 → 物品表資源
0x1862E  第 0–2 組再 ＋= [ds:BE20h + 組 × 2]   ; 0／0x2FE／0x5FC
0x18645  讀 0x2F8 bytes 到 ds:7A31h
```

檔案裡實際存在的 `msq` 區塊正好對上：

| 組 | 檔案 | 檔案位移 |
|---:|---|---|
| 0 | `game1` | `0x265CB` |
| 1 | `game1` | `0x268C9` |
| 2 | `game1` | `0x26BC7` |
| 4 | `game2` | `0x29DCD`（到檔尾剛好 6 ＋ 760 bytes）|

第 3 組沒有對應的區塊，出貨資料也沒有人用它。

**四份的差別只有庫存那一欄（`+0x02`）**：與第 0 組逐 byte 比，第 1 組差 3 筆、
第 2 組差 22 筆，全部落在 `+0x02`。價錢、傷害、類別、彈匣一個 byte 都沒變。

推論等級：**已確認**（載入器逐行讀出；四個區塊的位移與檔案內容對上；
逐 byte 差異用 `tools/dump_items.py` 的解密函式掃過）。

## 2. 哪一組由設施記錄的 `+0x06` 決定

商店進場（`0x1BE50` 與 `0x1BEA2` 兩條路）：

```
bl ← 6；al ← [ds:46AEh + 6]      ; 設施記錄 +0x06
ds:46C4h ← al
call sub_185E6                   ; 需要的話換一份
```

出貨資料裡七家商店用到的值：

| 組 | 商店 |
|---:|---|
| 0 | 交易車廂（資源 8）、農業站雜貨店（資源 9）|
| 1 | 高池鎮商店（資源 10）|
| 2 | 石英城商場（資源 1）|
| 4 | 維加斯的市場與黑市（資源 21，兩家共用）|

⚠ **醫生的 `+0x06` 不是庫存組**：實測值是 50、70、80、150、200、250——
那是每點治療費（`docs/re/42` §5）。只有商店那條路把它寫進 `ds:46C4h`。

⚠ **離開商店不會換回去**。`ds:46C5h` 就這樣留著，直到下一家店要別的一組。
所以它是活著的全域狀態，賣東西的 `+1` 也是寫回**當時那一組**
（`0x1BF5F` 的寫回路徑同樣讀 `ds:46C5h`）。

開場時 `0x1633C` 先把 `ds:46C5h` 設成 `0xFF`（＝ 還沒載入任何一組），
再 `call sub_185E6` 載 `ds:46C4h` 的初值 0 ——**所以開場是第 0 組**。

推論等級：**已確認**（兩處寫入點 ＋ 出貨資料的實際值 ＋ 實機清單逐行對上）。

## 3. 對得上實機的那一步

高池鎮商店（第 1 組）的第一頁，原版與 remake 現在逐行相同：

```
1)  10 Book        2)  10 Canteen   3)  15 Crowbar
4) 150 Gas mask    5) 300 Geiger counter
6)  10 Hand mirror 7)   1 Jug       8)  50 Map        9)  10 Match
```

`Engine` 在第 0 組的庫存是 `0xFF`、在第 1 組是 `0`，所以第 1 組不賣它；
`Shovel` 與 `Snake squeezin` 反過來（第 0 組 0、第 1 組 `0xFF`），
實機第二頁確實有這兩件。**這三件正好就是第 0 組與第 1 組的全部差異**。

## 4. remake 這一側

| 項目 | 狀態 |
|---|---|
| 開場載第 0 組（`game1`）| **已接**（`play.New`）|
| 進商店照記錄 `+0x06` 換組 | **已接**（`Scene.loadItemStock`，`internal/play/itemstock.go`）|
| 賣東西的 `+1` 寫回**當時那一組** | **已接**（`SetItemTable(itemStockFile, itemStockSlot, …)`）|
| 醫生不套用 | **已接**（只有 `FacilityShop` 走那一支）|

⚠ 之前 remake 三處都寫死「存檔那個檔案的第 0 組」，所以**每一家店賣的東西
都一樣**，而且賣掉的東西會記到別的組去。

## 5. 可重跑

```bash
# 原版：走進高池鎮商店、翻兩頁
tools/go.sh run ./cmd/wl-save -dir workplace/dosbox/game -map 10 -at 30,25
tools/dosbox.sh "wait:6;key:Return;wait:3;key:p;wait:4;key:i;wait:2;key:y;wait:4;\
key:1;wait:2;key:b;wait:3;shot:44-buy;key:k;wait:3;shot:47-buy2"

# remake：同一份資料
tools/go.sh run ./cmd/wl-play -rom workplace/dosbox/game -modified -script "up,Y,B" -trace

# 四份表的逐 byte 差異
docker run --rm --log-opt max-size=10m --log-opt max-file=3 --network none \
  -v "$PWD:/w" -w /w -u "$(id -u):$(id -g)" python:3.12-slim python3 -c "
import sys; sys.path.insert(0,'tools')
from dump_items import decrypt, SAVE_BASE, ITEM_DELTA, ITEM_BYTES, SLOT_STRIDE
blob=open('workplace/orig/wastland/game1','rb').read()
base=decrypt(blob, SAVE_BASE['game1']+ITEM_DELTA, ITEM_BYTES)[0]
for slot in (1,2):
    p=decrypt(blob, SAVE_BASE['game1']+ITEM_DELTA+slot*SLOT_STRIDE, ITEM_BYTES)[0]
    d={}
    for i,(a,b) in enumerate(zip(base,p)):
        if a!=b: d[i%8]=d.get(i%8,0)+1
    print('slot',slot,'差異欄位→筆數',d)"
```

## 6. 這一輪學到的（寫成規則）

- **「同一張表」要問「有幾份」。** 這張表在檔案裡有四份，載入器用一個變數
  記著現在是哪一份——只讀「第 0 份」的程式碼會**永遠自洽**，
  因為四份的欄位長得一模一樣，只有內容不同。
- **差異只有一欄時，症狀會偽裝成資料問題。** 多出來的 `Engine` 看起來像
  「解碼把某個旗標讀錯了」，實際上是整張表拿錯。
  **先確認兩邊讀的是同一份資料，再懷疑解碼。**
- **實機對拍要用同一份存檔。** 一開始拿 remake 的乾淨資料對 DOSBox 的可寫副本，
  兩邊的庫存本來就不同——那會把真正的差異蓋掉。
