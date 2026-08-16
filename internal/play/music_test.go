package play

import (
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"
)

// TestMusicTracksMatchTheComposer：Go 這邊列的曲名要與譜曲腳本產的檔名一致。
//
// 兩邊分開寫就會漂，而漂掉**不會報錯**：`internal/ui/music.go` 找不到曲子時
// 維持現況（比突然靜音好），所以少一首的症狀是「走到某個地方音樂就不換了」。
func TestMusicTracksMatchTheComposer(t *testing.T) {
	src, err := os.ReadFile("../../tools/make_music.py")
	if err != nil {
		t.Skipf("讀不到譜曲腳本：%v", err)
	}
	block := regexp.MustCompile(`(?s)TUNES = \{(.*?)\n\}`).FindSubmatch(src)
	if block == nil {
		t.Fatal("在 tools/make_music.py 裡找不到 TUNES")
	}
	var composed []string
	for _, m := range regexp.MustCompile(`"([a-z]+)":`).FindAllSubmatch(block[1], -1) {
		composed = append(composed, string(m[1]))
	}
	want := append([]string(nil), MusicTracks...)
	sort.Strings(want)
	sort.Strings(composed)
	if strings.Join(want, ",") != strings.Join(composed, ",") {
		t.Errorf("曲目對不上：\n  Go   %v\n  譜曲 %v", want, composed)
	}
}

// TestMusicTrackCoversEveryMap：42 張地圖、日夜兩種，回的曲名都要在清單裡。
//
// 這一條擋的是「某張地圖對到一首不存在的曲子」——那時候玩家走進去會發現
// 音樂就停在上一首，沒有任何錯誤訊息。
func TestMusicTrackCoversEveryMap(t *testing.T) {
	rom := openRom(t)
	s, err := New(rom)
	if err != nil {
		t.Fatal(err)
	}
	known := map[string]bool{}
	for _, n := range MusicTracks {
		known[n] = false
	}

	resources, err := rom.Resources()
	if err != nil {
		t.Fatal(err)
	}
	maps := 0
	for _, res := range resources {
		if err := s.LoadMap(res.ID, 1, 1); err != nil {
			continue // 不是地圖的資源（存檔、圖片…）
		}
		maps++
		for _, hour := range []uint8{12, 23} {
			s.World().Clock.Hour = hour
			got := s.MusicTrack()
			if _, ok := known[got]; !ok {
				t.Errorf("資源 %d（%d 時）對到不存在的曲子 %q", res.ID, hour, got)
				continue
			}
			known[got] = true
		}
	}
	if maps < 40 {
		t.Fatalf("只載到 %d 張地圖，這一條沒驗到東西", maps)
	}

	// 地圖模式用得到五首；另外五首靠場景模式，下面那條測。
	for _, n := range []string{"desert", "night", "town", "sewer", "vegas", "base"} {
		if !known[n] {
			t.Errorf("走過 %d 張地圖，沒有一張對到 %q", maps, n)
		}
	}
	t.Logf("%d 張地圖 × 日夜兩種都對得到曲子", maps)
}

// TestMusicTrackFollowsMode：曲子跟著場景走，而且**畫面模式先於地點**。
//
// **這張對照表是重製版的決定，不是逆向結論**——原版沒有背景音樂
// （九首 PC 喇叭音效，`docs/re/44`）。所以這一條驗的是我們自己訂的規則。
func TestMusicTrackFollowsMode(t *testing.T) {
	s := newScene(t)
	// 出廠存檔在大地圖上、時鐘 01:00 ＝ 夜間（門檻 6 時與 18 時）。
	if got := s.MusicTrack(); got != "night" {
		t.Errorf("出廠存檔（01:00 的大地圖）放 %q，預期 night", got)
	}
	s.World().Clock.Hour = 12
	if got := s.MusicTrack(); got != "desert" {
		t.Errorf("同一張地圖白天放 %q，預期 desert", got)
	}

	// 拉斯維加斯有自己的曲子，而且不分晝夜。
	if err := s.LoadMap(vegasMap, 1, 1); err != nil {
		t.Fatalf("載不到拉斯維加斯：%v", err)
	}
	for _, hour := range []uint8{12, 23} {
		s.World().Clock.Hour = hour
		if got := s.MusicTrack(); got != "vegas" {
			t.Errorf("拉斯維加斯（%d 時）放 %q，預期 vegas", hour, got)
		}
	}

	// 在賭城裡打起來時放戰鬥曲：當下在意的是那場架，不是地點。
	if err := s.LoadMap(4, 18, 2); err != nil {
		t.Fatal(err)
	}
	if _, err := s.StartEncounter(); err != nil {
		t.Fatal(err)
	}
	if got := s.MusicTrack(); got != "combat" {
		t.Errorf("戰鬥中放 %q，預期 combat", got)
	}

	s2 := newScene(t)
	s2.BeginTitle()
	if got := s2.MusicTrack(); got != "theme" {
		t.Errorf("標題畫面放 %q，預期 theme", got)
	}
	// ⚠ 結局要用另一個場景測。標題排在最前面（它是最外層的畫面），
	// 同一個場景上先開標題再開結局，回的會是 theme——而**實際流程走不到
	// 那個狀態**：結局是從地圖上觸發的，那時候標題早就關了。
	s3 := newScene(t)
	s3.BeginEnding()
	if got := s3.MusicTrack(); got != "ending" {
		t.Errorf("結局放 %q，預期 ending", got)
	}
}

// TestBaseCochiseHasItsOwnMusic：科奇斯基地那五張走同一首。
//
// 範圍用的是既有常數（結局清算也靠它），所以這一條順便釘住「兩處用同一個範圍」。
func TestBaseCochiseHasItsOwnMusic(t *testing.T) {
	s := newScene(t)
	for id := BaseCochiseFirst; id < BaseCochiseEnd; id++ {
		if err := s.LoadMap(id, 1, 1); err != nil {
			t.Errorf("載不到地圖 %d：%v", id, err)
			continue
		}
		if got := s.MusicTrack(); got != "base" {
			t.Errorf("地圖 0x%X 放 %q，預期 base", id, got)
		}
	}
}
