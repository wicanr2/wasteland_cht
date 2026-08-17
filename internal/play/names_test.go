package play

import (
	"os"
	"path/filepath"
	"testing"
	"unicode/utf8"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

// 側車檔自己 round-trip。
func TestLongNamesRoundTrip(t *testing.T) {
	dir := t.TempDir()
	var want longNames
	want[1] = "沙漠遊俠一二三四五" // 27 bytes，10 個字
	want[3] = "Hell Razor the Third"
	if err := storeLongNames(dir, want); err != nil {
		t.Fatalf("寫側車檔：%v", err)
	}
	got := loadLongNames(dir)
	if got != want {
		t.Errorf("讀回來不一樣：\n got %q\nwant %q", got, want)
	}
}

// 檔案不在就回空的——**不要因此讓遊戲開不起來**。
func TestLongNamesMissingFileIsFine(t *testing.T) {
	if got := loadLongNames(t.TempDir()); got != (longNames{}) {
		t.Errorf("沒有檔案卻讀出東西：%q", got)
	}
	if got := loadLongNames(""); got != (longNames{}) {
		t.Errorf("沒有目錄卻讀出東西：%q", got)
	}
}

// 檔案壞掉也不能崩，而且**壞掉的那一槽要丟掉不能塞半個字**。
func TestLongNamesRejectsBrokenFile(t *testing.T) {
	for name, raw := range map[string][]byte{
		"不是側車檔":  []byte("this is not it"),
		"截斷":     []byte("WLNM\x01\x00\xff\xff"),
		"版本不合":   []byte("WLNM\x09\x00"),
		"半個 UTF-8": append([]byte("WLNM\x01\x00\x02\x00"), 0xE6, 0xB2),
	} {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, namesFile), raw, 0o644); err != nil {
			t.Fatal(err)
		}
		got := loadLongNames(dir)
		for i, n := range got {
			if n != "" && !utf8.ValidString(n) {
				t.Errorf("%s：第 %d 槽讀出不合法的 UTF-8：% x", name, i, n)
			}
		}
	}
}

// **端到端**：打一個 10 個中文字的名字 → 存檔 → 重新開場 → 名字完整回來。
//
// ⚠ 這一條是整個側車檔的理由。原版存檔那一格只有 13 bytes，
// 少了側車檔的話重開之後名字會變成前四個字——**而畫面上看起來很正常**，
// 玩家只會覺得「我剛剛不是打了十個字嗎」。
func TestLongNameSurvivesSaveAndReload(t *testing.T) {
	const long = "沙漠遊俠一二三四五六" // 30 bytes，10 個字
	if len(long) != 30 {
		t.Fatalf("測試資料應該是 30 bytes，實際 %d", len(long))
	}
	if game.NameFitsSave(long) {
		t.Fatal("30 bytes 不該塞得進 13 bytes 的欄位——這條測試就沒意義了")
	}

	dir := t.TempDir()
	s := newScene(t)
	s.SetSaveDir(dir)
	m := s.World().Party.Members[0]
	m.Name = long

	if err := s.StoreTo(s.save); err != nil {
		t.Fatalf("寫回：%v", err)
	}
	if err := storeLongNames(dir, s.longNames); err != nil {
		t.Fatalf("寫側車檔：%v", err)
	}

	// 存檔裡那一格應該是**截斷過的**，而且截在 rune 邊界。
	raw, err := s.save.Record(1)
	if err != nil {
		t.Fatal(err)
	}
	short := game.LoadCharacter(raw).Name
	if len(short) > game.NameFieldBytes {
		t.Errorf("存檔那一格 %d bytes，超過 %d", len(short), game.NameFieldBytes)
	}
	if !utf8.ValidString(short) {
		t.Errorf("存檔裡截在字中間了：% x", short)
	}

	// 重新開場：側車檔要把完整名字帶回來。
	s2 := newScene(t)
	s2.SetSaveDir(dir)
	if got := s2.longNames[1]; got != long {
		t.Fatalf("側車檔沒讀回來：%q", got)
	}
	if got := s2.World().Party.Members[0].Name; got != long {
		t.Errorf("重開之後名字是 %q，預期 %q", got, long)
	}
}

// 短名字**不進側車檔**——存檔那一格就夠了，不要多存一份。
func TestShortNameDoesNotUseSidecar(t *testing.T) {
	dir := t.TempDir()
	s := newScene(t)
	s.SetSaveDir(dir)
	s.World().Party.Members[0].Name = "Hell Razor" // 10 bytes
	if err := s.StoreTo(s.save); err != nil {
		t.Fatal(err)
	}
	if got := s.longNames[1]; got != "" {
		t.Errorf("短名字不該進側車檔，得到 %q", got)
	}
}

// 輸入層收得下 10 個中文字（30 bytes），第 11 個要被擋掉。
func TestNameEntryStopsAtThirtyBytes(t *testing.T) {
	if input.MaxName != 30 {
		t.Fatalf("上限應該是 30 bytes，實際 %d", input.MaxName)
	}
	var e input.TextEntry
	e.Max = input.MaxName
	for _, r := range "沙漠遊俠一二三四五六七" { // 11 個字 ＝ 33 bytes
		e.KeyRune(r, nil)
	}
	got := string(e.Text())
	if got != "沙漠遊俠一二三四五六" {
		t.Errorf("應該停在第 10 個字，得到 %q（%d bytes）", got, len(got))
	}
	// ⚠ **不能停在半個字上**——那會讓畫面出現一格空白。
	if !utf8.ValidString(got) {
		t.Errorf("停在字中間了：% x", got)
	}
}
