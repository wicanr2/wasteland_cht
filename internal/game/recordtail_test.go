package game

// 角色記錄尾巴那塊空地與階級欄的實際大小（`docs/re/109`）。

import (
	"bytes"
	"testing"
)

// 階級欄的邊界是**下一個已知欄位**擋出來的，不是宣告出來的：
// 原版的寫入迴圈抄到 NUL 為止、沒有長度檢查（`docs/re/109` §4）。
//
// ⚠ 這一條用**實際最長的階級**驗：`Lieutenant Commander` 20 個字元。
// 舊的 14 bytes 欄位會把它截成 `Lieutenant Com`——而截斷之後照樣存得回去、
// 照樣讀得出來，**畫面上只是少了半個字，沒有任何錯誤**。
func TestRankFieldHoldsTheLongestRank(t *testing.T) {
	const longest = "Lieutenant Commander"
	raw := make([]byte, 256)
	c := &Character{Rank: longest}
	c.StoreTo(raw)
	if got := LoadCharacter(raw).Rank; got != longest {
		t.Errorf("階級 = %q，預期 %q", got, longest)
	}
	// NUL 要落在 +0x46（0x32 + 20），下一個已知欄位 +0x4B 之前。
	if raw[recRank+len(longest)] != 0 {
		t.Error("字串沒有 NUL 結尾")
	}
	if raw[recMission] != 0 {
		t.Error("寫到 +0x4B 的任務旗標了")
	}
}

// 原版**不清尾巴**（抄到 NUL 就 retn），所以我們也不能清——
// 存檔裡 NUL 之後的殘骸要原樣留著，否則就不是 round-trip。
func TestRankStoreLeavesTheTailAlone(t *testing.T) {
	raw := make([]byte, 256)
	for i := recRank; i < recRankEnd; i++ {
		raw[i] = 0xAA
	}
	(&Character{Rank: "PRIVATE"}).StoreTo(raw)
	if got := string(raw[recRank : recRank+7]); got != "PRIVATE" {
		t.Fatalf("字串沒寫進去：%q", got)
	}
	if raw[recRank+7] != 0 {
		t.Fatal("沒有補 NUL")
	}
	for i := recRank + 8; i < recRankEnd; i++ {
		if raw[i] != 0xAA {
			t.Fatalf("位移 %#x 被清成 %#x 了，原版不會動這裡", i, raw[i])
		}
	}
}

// 階級長到蓋掉旗標時要截斷。原版沒有這道防護（它的字串表最長 20 個字元），
// 我們有——因為譯文可能更長。
// ⚠ 直接測 `putRank`，不走 `StoreTo`：`StoreTo` 在階級之後才寫兩個旗標，
// 會把溢位蓋回去，於是測不出溢位（**這種假綠只有在別的欄位剛好蓋在
// 出事的位置上時才會發生**，比一般的假綠更難注意到）。
func TestRankCannotClobberMissionFlag(t *testing.T) {
	raw := make([]byte, 256)
	raw[recMission] = 0xA5
	raw[recPraised] = 0xA5
	putRank(raw, string(bytes.Repeat([]byte("X"), 64)))
	if raw[recMission] != 0xA5 || raw[recPraised] != 0xA5 {
		t.Fatalf("兩個旗標被階級字串蓋掉了：%#x %#x", raw[recMission], raw[recPraised])
	}
	// 截斷之後最後一格仍然要是 NUL，否則讀回來會把旗標當成字串的一部分。
	if raw[recRankEnd-1] != 0 {
		t.Fatalf("截斷之後沒有 NUL：%#x", raw[recRankEnd-1])
	}
}

// `+0x4D`–`+0x7F` 出廠全零（`docs/re/109` §3）。
//
// ⚠ 這一條**不是**在驗「程式碼沒有存取點」——那是靜態盤點的結論（§2）。
// 它驗的是另一半：出廠資料裡那塊也是空的。兩件事都成立才敢說那塊是空地。
func TestRecordTailIsZeroInFactorySaves(t *testing.T) {
	rom := openRom(t)
	for _, file := range []string{"game1", "game2"} {
		save, err := rom.LoadSave(file)
		if err != nil {
			t.Fatalf("%s 讀存檔：%v", file, err)
		}
		for n := 1; n <= 7; n++ {
			raw, err := save.Record(n)
			if err != nil {
				t.Fatalf("%s 記錄 %d：%v", file, n, err)
			}
			for i := 0x4D; i <= 0x7F; i++ {
				if raw[i] != 0 {
					t.Errorf("%s 記錄 %d 的 +%#04x ＝ %#04x，預期 0", file, n, i, raw[i])
				}
			}
		}
	}
}

// 出廠記錄讀出來再寫回去要 byte-for-byte 相同（`CLAUDE.md` §4）。
// 階級欄改大之後這一條特別重要：多寫一個 byte 就會紅。
func TestFactoryRecordRoundTrips(t *testing.T) {
	rom := openRom(t)
	save, err := rom.LoadSave("game1")
	if err != nil {
		t.Fatalf("讀存檔：%v", err)
	}
	for n := 1; n <= 7; n++ {
		raw, err := save.Record(n)
		if err != nil {
			t.Fatalf("記錄 %d：%v", n, err)
		}
		before := append([]byte(nil), raw...)
		LoadCharacter(raw).StoreTo(raw)
		if !bytes.Equal(before, raw) {
			for i := range before {
				if before[i] != raw[i] {
					t.Fatalf("記錄 %d 的 +%#04x：%#04x → %#04x", n, i, before[i], raw[i])
				}
			}
		}
	}
}
