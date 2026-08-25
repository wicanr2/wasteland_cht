package play

// 背景音樂的曲目對照（重製版自己加的，`docs/spec/27` §3）。
//
// **原版沒有 BGM**，所以這張表不是逆向結論，是我們的編排決定。分區的**依據**
// 倒是資料講的：地圖邊長寫在記錄區標頭（`docs/re/24`），科奇斯基地的地圖編號
// 範圍是既有常數（`BaseCochiseMaps`），晝夜門檻取原版的 6 時與 18 時
// （`docs/re/27` §5，`game.Clock.Night`）。
//
// ⚠ 曲名就是檔名：`-music` 指到的目錄裡要有 `<曲名>.ogg`。
// **少一首不會報錯**，播放器會維持現況（`internal/ui/music.go`），
// 所以加曲目時 `MusicTracks` 與 `tools/make_music.py` 的 `TUNES` 要一起改，
// `TestMusicTracksMatchTheComposer` 守著兩邊一致。

// MusicTracks 是這一版用到的全部曲名。
var MusicTracks = []string{
	"theme",    // 標題與主選單
	"desert",   // 大地圖，白天
	"night",    // 大地圖，夜間
	"town",     // 一般 32 × 32 的城鎮與室內
	"facility", // 商店、醫生、訓練師、遊俠中心
	"sewer",    // 下水道與地底隧道
	"vegas",    // 拉斯維加斯
	"base",     // 科奇斯基地
	"combat",   // 戰鬥
	"ending",   // 結局
}

// 分區用得到的地圖編號。
const (
	// vegasMap 是拉斯維加斯（資源 12，64 × 64，`docs/walkthrough/generated/maps.md`）。
	vegasMap = 12
	// overworldDim 是大地圖的邊長。42 張地圖裡只有四張是 64，其餘都是 32
	// （`docs/re/24`），所以邊長就足以分辨「在荒漠上」與「在建築裡」。
	overworldDim = 64
)

// sewerMaps 是走地底的那幾張。判準取自地圖自己的入口敘述
// （`docs/walkthrough/generated/maps.md`）：資源 1 是「你在黏滑的地底爬行」，
// 資源 29 是「你沿著一條漆黑的隧道往下走」。
//
// 科奇斯基地底下那幾層（0x11–0x14）也在地底，但它們走 `base`——
// 那裡的重點不是潮濕是緊繃。
var sewerMaps = map[int]bool{1: true, 29: true}

// MusicTrack 是這一幀該播的曲子（`ui.Musical`）。
//
// 順序有意義：**畫面模式先於地點**。在拉斯維加斯的商店裡打起來時放的是
// 戰鬥曲，不是賭場曲——玩家當下在意的是那場架。
func (s *Scene) MusicTrack() string {
	switch {
	case s.title:
		return "theme"
	case s.ending.active:
		return "ending"
	case s.combat != nil:
		return "combat"
	case s.facility != nil, s.roster.active:
		return "facility"
	}
	return s.mapTrack()
}

// mapTrack 是「人在地圖上」時的曲子：先看地點，再看晝夜。
func (s *Scene) mapTrack() string {
	switch {
	case s.blockID >= BaseCochiseFirst && s.blockID < BaseCochiseEnd:
		return "base"
	case s.blockID == vegasMap:
		return "vegas"
	case sewerMaps[s.blockID]:
		return "sewer"
	}
	if s.world != nil && s.world.Block != nil && s.world.Block.Dim >= overworldDim {
		// 大地圖才分晝夜：室內看不到天色，換曲只會讓人覺得莫名其妙。
		if s.world.Clock.Night() {
			return "night"
		}
		return "desert"
	}
	return "town"
}

// MusicSetting 是玩家在 F2 設定裡選的開關與音量（`ui.Musical`）。
func (s *Scene) MusicSetting() (bool, int, string) {
	return s.settings.MusicOn, s.settings.MusicVol, s.settings.MusicVariant
}
