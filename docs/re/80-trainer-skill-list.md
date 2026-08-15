# 80：訓練師的技能清單

日期：2026-08-15 ｜ 接 `docs/re/52`（訓練師流程）、`docs/re/79`（設施覆蓋率）

輸入：`wl.merged.exe`，SHA-256
`cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118`。

`docs/re/79` 量到「八家訓練師進得去但一個技能都學不到」。缺的是清單來源。

---

## 1. 清單沒有篩選，篩選在選完之後

`0x1BBA0`（設施 2）的主迴圈：

```
0x1BC30  ds:469Dh ＝ 1 → loc_1BC86
         sub_1BE40；sub_1BE2A(3)          ; 印提示
         sub_16DB4                        ; **清單框架**（docs/re/53）
         bl ← 0x20；al ← 角色記錄 +0x20   ; 技能點數
         sub_1BD7F(al)                    ; 把它印出來
         ds:7DF3h ← 0x434
         sub_16D34 → al                   ; 清單選擇
         al ＝ 0FFh → 回頭

0x1BC60（選完之後才篩）:
         dl ← 1；sub_1CA8D → 基礎費用
         費用 > 角色記錄 +0x20 → 字串 7 ＋ 音效 7，回頭
         否則學起來
```

`sub_1BD7F` 只是**印技能點數**（`sub_176A2` → `sub_176A8` 把值丟進
`ds:466Bh` 再走數字轉換），不是建清單。

**清單框架列的是整張技能資料表**，而 IQ 需求、費用、角色技能欄還有沒有空位
（`sub_1BDFF`）全都在選完之後才檢查。所以**每家訓練師列的東西一樣**——
記錄裡沒有清單，也不需要有。

推論等級：**強證據**。清單框架與選後檢查都是直讀的；
「清單本身不篩選」是從「篩選發生在選之後」推出來的。

## 2. `sub_1BDFF` 不是店家清單

```
0x1BDFF  bl ← 0x80
         al ＝ 角色記錄[bl] → 找到 → clc
         bl += 2；到 0xBC 為止
         沒找到 → stc
```

角色記錄 `+0x80`–`+0xBC` 是技能陣列（每筆 2 bytes，`docs/re/15`）。
`sub_1BDFF(0)` ＝ 「陣列裡還有沒有 ID 0 的空槽」，回 CF ＝ 1 表示**學不下了**
（`docs/re/52` §2 的字串 2「沒有可學的」）。與這家店教什麼無關。

## 3. remake 這一側

`Scene.trainableSkills()` 讀 `Rom.SkillTableRaw()`（36 筆 × 2，`ds:BA20h`），
過濾掉 `BaseCost ＝ 0` 的槽——出貨資料裡 36 筆的費用全部 > 0，所以一筆都沒濾掉。
`EnterFacility` 對 `FacilityTrainer` 填 `fs.Skills`。

```
tools/go.sh test ./internal/play/ -run TestTrainerListsSkills -v
    Library 列出 36 個技能，第一個 ID=0 IQ=1 費用=1
tools/go.sh test ./internal/play/ -run TestTrainerEntryEndToEnd -v
    走進『  Library   』，列出 36 個技能
```

第二個測試是端到端：從地圖走上傳送格、答 Y、進設施、確認 Kind 與清單。

`docs/re/79` 的反向門檻（「訓練師八家全空」）如預期紅了，已改成正向門檻
「一家都不能空」。

## 4. 可重跑的完整指令

```bash
WL_IDA_TARGET="$PWD/workplace/analysis/unpacked/wl.merged.exe" \
  tools/ida.sh run tools/ida/export_forced.py \
  workplace/analysis/dumps/trainer3.json 0x1BBA0 0x1BC30 0x1BC60

tools/go.sh test ./internal/play/ -run TestTrainerListsSkills -v
tools/go.sh test ./internal/play/ -run TestTrainerEntryEndToEnd -v
tools/go.sh test ./internal/play/ -run TestFacilityCoverage -v
```

## 5. 這一輪學到的（寫成規則）

- **「清單從哪來」的答案可能是「不從哪來」。** 找了三支函式想挖出每家店的
  技能清單，最後發現原版根本沒有那個概念——列全部、選完再篩。
  **資料裡找不到某個欄位時，先確認那個機制是不是真的需要它。**
- **名字像的函式不一定做那件事。** `sub_1BD7F` 出現在建清單的位置、
  參數是技能點數，看起來就是「依點數建清單」；實際上它只是印那個數字。
  **跟到最後一層再下結論**（`sub_1BD7F` → `sub_176A2` → `sub_176A8` 才看到 `ds:466Bh`）。
- **反向門檻兌現了。** `docs/re/79` 寫的「八家全空，第一家接上就紅」
  在這一輪如期紅了，逼著回來更新文件與門檻——比一句 TODO 註解有效。
