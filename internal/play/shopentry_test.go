package play

import (
	"strings"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/input"
)

// enterVia 走一步踩上傳送格、答 Y，回傳當時的場景。
func enterVia(t *testing.T, id, x, y int, dir input.Direction) *Scene {
	t.Helper()
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadMap(id, uint8(x), uint8(y)); err != nil {
		t.Fatalf("載入地圖 %d 失敗：%v", id, err)
	}
	if _, err := s.Update(input.Input{Dir: dir}); err != nil {
		t.Fatalf("走進傳送格失敗：%v", err)
	}
	if !s.Asking() {
		t.Fatalf("沒有問 Enter new location?，訊息是 %q", s.Message())
	}
	if _, err := s.Update(input.Input{Char: 'Y'}); err != nil {
		t.Fatalf("答 Y 失敗：%v", err)
	}
	return s
}

// TestTeleportPatchOpensFacility 驗商店與醫生的入口（docs/re/73）。
//
// 傳送記錄的 `+0x04`／`+0x05` 在傳送收尾（`sub_169B1(4)`）把落點改寫成
// (nibble 6, 設施記錄)，那一格的事件才把設施畫面叫出來。
// 22 筆設施全部走這條路——**沒有任何一筆是靠初始地圖上的 nibble 6 格子**。
func TestTeleportPatchOpensFacility(t *testing.T) {
	for _, c := range []struct {
		name    string
		id      int
		x, y    int
		dir     input.Direction
		wantFac string
	}{
		{"Store", 10, 30, 25, input.DirUp, "Store"},
		{"Market", 21, 2, 22, input.DirUp, "Market"},
	} {
		t.Run(c.name, func(t *testing.T) {
			s := enterVia(t, c.id, c.x, c.y, c.dir)
			if !s.InFacility() {
				t.Fatalf("答 Y 之後沒有進設施，訊息是 %q", s.Message())
			}
			got := s.Facility().Facility.Name
			if !strings.Contains(got, c.wantFac) {
				t.Errorf("進的是 %q，預期含 %q", got, c.wantFac)
			}
		})
	}
}

// TestRangerCenterOpensFacility 是同一條路的起點：出廠存檔那一格。
func TestRangerCenterOpensFacility(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatalf("開場失敗：%v", err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirUp}); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Update(input.Input{Dir: input.DirDown}); err != nil {
		t.Fatal(err)
	}
	if !s.Asking() {
		t.Fatalf("沒有問 Enter new location?")
	}
	if _, err := s.Update(input.Input{Char: 'Y'}); err != nil {
		t.Fatal(err)
	}
	if !s.InFacility() {
		t.Fatalf("答 Y 之後沒有進設施，訊息是 %q", s.Message())
	}
	if got := s.Facility().Facility.Name; !strings.Contains(got, "Ranger") {
		t.Errorf("進的是 %q，預期含 Ranger", got)
	}
}
