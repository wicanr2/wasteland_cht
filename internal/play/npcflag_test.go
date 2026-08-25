package play

import "testing"

// +0x29 ＝ NPC 旗標的一手佐證（docs/re/133 §1）：
// 出廠四個 Ranger 全是 0、可雇 NPC（section 17 有名字的）非 0。
func TestNPCFlagInShippedData(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	for n := 1; n <= 4; n++ {
		raw, err := s.save.Record(n)
		if err != nil {
			t.Fatal(err)
		}
		if raw[0x29] != 0 {
			t.Errorf("出廠記錄 %d（%q）的 +0x29 應該是 0，得到 %#x",
				n, string(raw[0:8]), raw[0x29])
		}
	}
	// 三個確定可雇的 NPC（docs/re/110）。
	for _, tc := range []struct {
		block, rec int
		name       string
	}{{4, 1, "FELICIA"}, {4, 2, "ACE"}, {10, 1, "JACKIE"}} {
		blk, err := s.rom.BlockByID(tc.block)
		if err != nil {
			t.Fatal(err)
		}
		raw, err := blk.SectionRecord(17, tc.rec)
		if err != nil {
			t.Fatal(err)
		}
		if got := string(raw[0:len(tc.name)]); got != tc.name {
			t.Fatalf("圖 %d 記錄 %d 應該是 %s，得到 %q", tc.block, tc.rec, tc.name, got)
		}
		if raw[0x29] == 0 {
			t.Errorf("%s 的 +0x29 應該非 0（NPC），得到 0", tc.name)
		}
	}
}
