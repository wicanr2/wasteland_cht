package play

// 一幀畫面的配置次數（GC 壓力）。**兩個分支要能跑同一支**——
// 所以這裡只碰兩邊型別相同的 API（`LoadCatalogue`／`LoadFont`／`Update`／`HiFrame`），
// 不碰 `s.cjk`（master 是 []byte、分支是 string）。

import (
	"os"
	"testing"

	"github.com/wicanr2/wasteland_cht/internal/assets"
	"github.com/wicanr2/wasteland_cht/internal/input"
	"github.com/wicanr2/wasteland_cht/internal/render"
)

func benchScene(b *testing.B) *Scene {
	b.Helper()
	dir := os.Getenv("WL_DATA")
	if dir == "" {
		dir = "../../workplace/orig/wastland"
	}
	rom, err := assets.Open(dir)
	if err != nil {
		b.Skipf("開啟原版資料失敗：%v", err)
	}
	if err := rom.LoadImage("../../workplace/analysis/unpacked/wl.merged.exe"); err != nil {
		b.Skipf("載入分析映像失敗：%v", err)
	}
	s, err := New(rom)
	if err != nil {
		b.Fatalf("開場失敗：%v", err)
	}
	if err := s.LoadCatalogue("../../translations/zh-Hant.cat"); err != nil {
		b.Skipf("載入翻譯目錄失敗：%v", err)
	}
	fdir := os.Getenv("WL_ETEN")
	if fdir == "" {
		fdir = "../../workplace/eten24"
	}
	if err := s.LoadFont(fdir); err != nil {
		b.Skipf("載入倚天字型失敗：%v", err)
	}
	return s
}

// BenchmarkHiFrame 量「整幀合成」——Ebiten 的 Draw **每一幀都無條件呼叫一次**
// （`internal/ui/ui.go`），所以這支量到的就是 60 fps 下每秒要付的成本。
func BenchmarkHiFrame(b *testing.B) {
	s := benchScene(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		h := s.HiFrame()
		if h == nil {
			b.Fatal("沒有畫面")
		}
	}
}

// BenchmarkHiFrameDraw 是**視窗那一層真正跑的路**：合成一幀 ＋ 上色寫進
// 重複用的緩衝區（`internal/ui` 的 Draw）。與 `BenchmarkHiFrameToImage`
// 的差別就是那 2.3 MB 要不要每幀重配。
func BenchmarkHiFrameDraw(b *testing.B) {
	s := benchScene(b)
	pix := make([]byte, render.RGBABytes)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if !s.HiFrame().WriteRGBA(pix) {
			b.Fatal("上色失敗")
		}
	}
}

// BenchmarkHiFrameToImage 加上 `ToImage()`——一次性用途（截圖工具）那一條。
func BenchmarkHiFrameToImage(b *testing.B) {
	s := benchScene(b)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		im := s.HiFrame().ToImage()
		if im == nil {
			b.Fatal("沒有畫面")
		}
	}
}

// BenchmarkHiFrameCombat 是**中文最多的畫面**：戰鬥的名單 ＋ 提示 ＋ 指令列。
func BenchmarkHiFrameCombat(b *testing.B) {
	s := benchScene(b)
	if _, err := s.Update(input.Input{Char: 'E'}); err != nil {
		b.Fatalf("E：%v", err)
	}
	if _, err := s.Update(input.Input{Char: 'Y'}); err != nil {
		b.Fatalf("Y：%v", err)
	}
	if !s.InCombat() {
		b.Skip("這一格開不了戰鬥")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if h := s.HiFrame(); h == nil {
			b.Fatal("沒有畫面")
		}
	}
}
