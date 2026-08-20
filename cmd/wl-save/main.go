// 指令 wl-save 把存檔讀出來、（可選）改幾個欄位、再寫回原版資料檔。
//
// 存在的理由是**實機驗收**：`CLAUDE.md` §4 說存檔策略是「改寫不是重建」，
// 而那句話只有在原版真的讀得進去我們寫出來的檔時才成立。單元測試證不了這件事。
//
//	# 只驗 round-trip：讀出來再編碼回去，必須與檔案裡的 bytes 完全相同
//	go run ./cmd/wl-save -dir workplace/dosbox/game -check
//
//	# 改隊伍座標與時鐘，寫回去（會就地改 -dir 底下的 game1／game2）
//	go run ./cmd/wl-save -dir workplace/dosbox/game -at 57,39 -hour 2
//
// ⚠ **只對可寫的副本用**。`workplace/orig/wastland/` 是唯讀的原版目錄，
// 不要指到那裡（這支會拒絕沒有寫入權限的目錄，但別靠它擋）。
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

// 視窗原點與隊伍座標的差（docs/re/25 §2.1：隊伍固定在第 (9, 4) 格）。
const (
	viewOffsetX = 9
	viewOffsetY = 4
)

func main() {
	dir := flag.String("dir", "", "**可寫的**原版資料目錄副本")
	check := flag.Bool("check", false, "只驗 round-trip，不寫檔")
	at := flag.String("at", "", "把隊伍移到 x,y")
	mapID := flag.Int("map", -1, "把隊伍搬到這張地圖（資源編號 0–41）")
	hour := flag.Int("hour", -1, "把時鐘的「時」設成這個值（0–23）")
	flag.Parse()

	if err := run(*dir, *check, *at, *hour, *mapID); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤：", err)
		os.Exit(1)
	}
}

func run(dir string, check bool, at string, hour, mapID int) error {
	if dir == "" {
		return fmt.Errorf("要指定 -dir")
	}
	// 寫過一次之後檔案雜湊就不對了，所以這支一律走不驗雜湊的入口——
	// 否則同一份副本只能改一次。
	rom, err := assets.OpenModified(dir)
	if err != nil {
		return err
	}

	for _, name := range []string{"game1", "game2"} {
		sv, err := rom.LoadSave(name)
		if err != nil {
			return fmt.Errorf("%s：%w", name, err)
		}
		raw, err := rom.File(name)
		if err != nil {
			return err
		}
		span := raw[sv.Offset : sv.Offset+len(sv.Bytes())]

		// round-trip：沒改任何東西時，編碼回去必須一個 byte 都不差。
		if got := sv.Bytes(); !bytes.Equal(got, span) {
			n := 0
			for i := range got {
				if got[i] != span[i] {
					n++
				}
			}
			return fmt.Errorf("%s 的存檔 round-trip 不一致：%d／%d 個 byte 不同",
				name, n, len(got))
		}
		fmt.Printf("%s：round-trip 相同（%d bytes，序號 %d，地點 %q）\n",
			name, len(span), sv.Serial(), sv.Place())

		if check {
			continue
		}
		if changed, err := apply(sv, at, hour, mapID); err != nil {
			return fmt.Errorf("%s：%w", name, err)
		} else if !changed {
			continue
		}
		copy(span, sv.Bytes())
		out := filepath.Join(dir, name)
		if err := os.WriteFile(out, raw, 0o644); err != nil {
			return fmt.Errorf("寫回 %s：%w", out, err)
		}
		fmt.Printf("%s：已寫回（checksum 重算）\n", name)
	}
	return nil
}

// apply 改寫已解欄位。**未解區域一個 byte 都不碰**——
// 只動隊伍槽表的座標與全域狀態的時鐘，其餘原樣留著。
func apply(sv *assets.Save, at string, hour, mapID int) (bool, error) {
	changed := false
	if at != "" {
		var x, y int
		if _, err := fmt.Sscanf(strings.TrimSpace(at), "%d,%d", &x, &y); err != nil {
			return false, fmt.Errorf("-at 要寫成 x,y：%w", err)
		}
		if x < 0 || x > 255 || y < 0 || y > 255 {
			return false, fmt.Errorf("座標 (%d, %d) 超出 0–255", x, y)
		}
		g := sv.SlotGroups()[0]
		slot := sv.Plain[g.RawIndex : g.RawIndex+14]
		slot[8], slot[9] = byte(x), byte(y)
		// 視窗原點是**另外存的**（ds:464Eh／464Fh 的副本），原版不從座標推。
		// 只改座標不改它，載進去畫面會偏——這一格是 docs/re/25 §2.1 的
		// 「隊伍固定在第 (9, 4) 格」。
		gl := sv.Globals()
		gl[0], gl[1] = byte(x-viewOffsetX), byte(y-viewOffsetY)
		changed = true
	}
	if mapID >= 0 {
		// 隊伍槽表 `+0x0A` ＝ 那一組在哪張地圖（`docs/re/60` §3）。
		//
		// ⚠ **座標要跟著改**：換了地圖卻留著舊座標，隊伍會落在新地圖的
		// 同一個 (x, y) 上——可能是牆裡面。搭配 `-at` 一起用。
		if mapID > 41 {
			return false, fmt.Errorf("-map 要在 0–41，得到 %d", mapID)
		}
		g := sv.SlotGroups()[0]
		sv.Plain[g.RawIndex+10] = byte(mapID)
		// ⚠ **兩個地方都要寫**：原版讀檔只看全域狀態那 14 bytes 裡的
		// `ds:4655h`（相對位移 7），槽表的 `+0x0A` 是「每一組各自在哪」
		// （`docs/re/117`）。只寫槽表的話原版會用**舊地圖 ＋ 新座標**開場，
		// 而畫面上看起來只是「傳送到奇怪的地方」。
		sv.Globals()[7] = byte(mapID)
		changed = true
	}
	if hour >= 0 {
		if hour > 23 {
			return false, fmt.Errorf("-hour 要在 0–23，得到 %d", hour)
		}
		sv.Globals()[12] = byte(hour)
		changed = true
	}
	return changed, nil
}
