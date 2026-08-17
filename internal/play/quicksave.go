package play

// 快速存檔與快速讀檔（F5／F9，重製版自己加的）。
//
// 原版沒有這種東西：存檔就是 `GAME1`／`GAME2` 本身，指令列的 `Save` 就地寫回
// （`docs/re/95` §4）。而**那份檔案是玩家自備的原版資料**，`assets.Open` 驗
// SHA-256，寫過就開不起來。所以快速存檔走一份**獨立的檔**，
// 一個 byte 都不碰玩家的原版目錄。
//
// 檔案格式是自訂容器，裡面放的是**解開後的明文**：
//
//	"WLQS"        magic
//	u16           版本
//	u8 ＋ bytes   來源檔名（"game1"／"game2"）
//	u32 ＋ bytes  存檔明文（0x800）
//	u32 ＋ bytes  尾段（0xA00 的按鍵巨集）
//	u32 ＋ bytes  物品表明文（店家庫存）
//
// 存明文而不是原版那種加密資源，是因為這個檔只有我們自己讀——
// 加密與 checksum 是「寫回 GAME1／GAME2」時才需要的（`docs/re/08`）。

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/wasteland_cht/internal/game"
	"github.com/wicanr2/wasteland_cht/internal/game/rng"
	"github.com/wicanr2/wasteland_cht/internal/input"
)

const (
	quickMagic   = "WLQS"
	quickVersion = 1
)

// DefaultQuickSavePath 是沒指定時快速存檔放哪。
//
// 走 XDG：`$XDG_DATA_HOME/wasteland-cht/quicksave.wlq`，
// 沒設 `XDG_DATA_HOME` 就用 `~/.local/share/`。**不放在遊戲資料目錄裡**——
// 那裡是玩家自備的原版檔案。
func DefaultQuickSavePath() string {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "" // 家目錄都問不到就不給預設值，由呼叫端決定
		}
		base = filepath.Join(home, ".local", "share")
	}
	return filepath.Join(base, "wasteland-cht", "quicksave.wlq")
}

// SetQuickSavePath 指定快速存檔的檔案路徑；空字串 ＝ 關掉這個功能。
func (s *Scene) SetQuickSavePath(p string) { s.quickPath = p }

// QuickSavePath 回報目前的快速存檔路徑（空 ＝ 沒開）。
func (s *Scene) QuickSavePath() string { return s.quickPath }

// QuickSave 把目前的狀態寫進快速存檔。
func (s *Scene) QuickSave() error {
	if s.quickPath == "" {
		return fmt.Errorf("沒有指定快速存檔的位置")
	}
	if s.save == nil {
		return fmt.Errorf("還沒載入存檔")
	}
	if err := s.StoreTo(s.save); err != nil {
		return err
	}
	var buf []byte
	buf = append(buf, quickMagic...)
	buf = binary.LittleEndian.AppendUint16(buf, quickVersion)
	buf = append(buf, byte(len(s.save.File)))
	buf = append(buf, s.save.File...)
	for _, part := range [][]byte{s.save.Plain, s.save.Tail, s.itemsRaw} {
		buf = binary.LittleEndian.AppendUint32(buf, uint32(len(part)))
		buf = append(buf, part...)
	}
	if err := os.MkdirAll(filepath.Dir(s.quickPath), 0o755); err != nil {
		return err
	}
	// 先寫暫存檔再改名：中途當掉不會留下一個讀不回來的半截存檔。
	tmp := s.quickPath + ".tmp"
	if err := os.WriteFile(tmp, buf, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.quickPath)
}

// QuickLoad 讀回快速存檔並重建世界。
func (s *Scene) QuickLoad() error {
	if s.quickPath == "" {
		return fmt.Errorf("沒有指定快速存檔的位置")
	}
	if s.save == nil {
		return fmt.Errorf("還沒載入存檔")
	}
	raw, err := os.ReadFile(s.quickPath)
	if err != nil {
		return err
	}
	parts, file, err := parseQuick(raw)
	if err != nil {
		return err
	}
	// ⚠ **長度不合就拒絕**：明文的長度是存檔佈局的一部分（`docs/re/30`），
	// 硬套進去會把後面的角色記錄整個錯位，而症狀是「角色資料變成亂碼」。
	if len(parts[0]) != len(s.save.Plain) || len(parts[1]) != len(s.save.Tail) {
		return fmt.Errorf("快速存檔的長度不合（明文 %d／尾段 %d）",
			len(parts[0]), len(parts[1]))
	}
	if file != s.save.File {
		return fmt.Errorf("快速存檔來自 %s，目前載入的是 %s", file, s.save.File)
	}
	copy(s.save.Plain, parts[0])
	copy(s.save.Tail, parts[1])
	if len(parts[2]) > 0 {
		s.itemsRaw = parts[2]
		s.items = game.ParseItemTable(parts[2])
	}
	return s.reloadFromSave()
}

func parseQuick(raw []byte) (parts [3][]byte, file string, err error) {
	at := 0
	take := func(n int) ([]byte, error) {
		if at+n > len(raw) {
			return nil, fmt.Errorf("快速存檔在第 %d byte 就沒了", len(raw))
		}
		b := raw[at : at+n]
		at += n
		return b, nil
	}
	head, err := take(6)
	if err != nil {
		return parts, "", err
	}
	if string(head[:4]) != quickMagic {
		return parts, "", fmt.Errorf("這不是快速存檔（magic ＝ %q）", head[:4])
	}
	if v := binary.LittleEndian.Uint16(head[4:]); v != quickVersion {
		return parts, "", fmt.Errorf("快速存檔版本 %d，這一版讀 %d", v, quickVersion)
	}
	n, err := take(1)
	if err != nil {
		return parts, "", err
	}
	name, err := take(int(n[0]))
	if err != nil {
		return parts, "", err
	}
	for i := range parts {
		l, err := take(4)
		if err != nil {
			return parts, "", err
		}
		b, err := take(int(binary.LittleEndian.Uint32(l)))
		if err != nil {
			return parts, "", err
		}
		parts[i] = append([]byte(nil), b...)
	}
	return parts, string(name), nil
}

// reloadFromSave 從（剛換過的）存檔明文重建隊伍、時鐘與所在地圖。
//
// 與 `New` 走同一組函式，所以讀檔之後的狀態與「重開遊戲讀那一份存檔」一致。
// **所有子模式要清掉**：讀檔時可能正停在戰鬥或設施畫面上，
// 留著會拿舊世界的指標去畫新世界。
func (s *Scene) reloadFromSave() error {
	party, mapID, err := loadParty(s.save)
	if err != nil {
		return err
	}
	block, err := s.rom.BlockByID(mapID)
	if err != nil {
		return fmt.Errorf("載入地圖 %d：%w", mapID, err)
	}
	tiles, err := s.rom.Tileset(block.Tileset)
	if err != nil {
		return err
	}
	skills := s.world.Skills
	s.gfx.Tiles = tiles
	s.world = game.NewWorld(block, party, rng.New())
	s.world.Clock = loadClock(s.save)
	s.world.Skills = skills
	s.blockFile, s.blockID = block.Resource.File, block.Resource.ID
	g := s.save.SlotGroups()[0]
	s.back = game.Return{X: g.BackX, Y: g.BackY, MapID: g.BackMap}
	s.resetModes()
	s.dirty = true
	return nil
}

// resetModes 把所有子模式關掉，回到地圖模式。
func (s *Scene) resetModes() {
	s.combat, s.facility = nil, nil
	s.roster = rosterState{}
	s.order = orderState{}
	s.use = useState{}
	s.question = questionState{}
	s.ending = endingState{}
	s.journalOpen, s.disband, s.encAsk = false, false, 0
	s.help, s.settingsOpen, s.quitAsk = false, false, false
	// ⚠ `input.Direction` 的零值是 `DirUp`，**一定要明寫 DirNone**——
	// 忘了寫的話讀完檔畫面會停在「進新地點？」等 Y／N。
	s.asking = input.DirNone
	s.message, s.cjk = "", ""
}

// doQuickSave／doQuickLoad 是 F5／F9 的路由：做完把結果留在訊息列。
//
// **失敗要說出來。** 快速存檔沒寫成功卻不吭聲，玩家會以為存好了。
func (s *Scene) doQuickSave() (bool, error) {
	if err := s.QuickSave(); err != nil {
		s.message = "QUICKSAVE FAILED: " + err.Error()
		s.cjk = s.cjkFmtText("quick.savefailed", err.Error())
	} else {
		s.sayEN("Quick saved.", "quick.saved")
	}
	s.dirty = true
	return true, nil
}

func (s *Scene) doQuickLoad() (bool, error) {
	if err := s.QuickLoad(); err != nil {
		s.message = "QUICKLOAD FAILED: " + err.Error()
		s.cjk = s.cjkFmtText("quick.loadfailed", err.Error())
		s.dirty = true
		return true, nil
	}
	s.sayEN("Quick loaded.", "quick.loaded")
	s.dirty = true
	return true, nil
}
