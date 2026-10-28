package play

// 重製版的地圖狀態容器，契約見 docs/spec/27 的 2026-10-28 補充。
// Raw 與地圖三層分開保存：Block.SetCell 不會改寫 Raw。
import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"

	"github.com/wicanr2/wasteland_cht/internal/assets"
)

const worldStateFile = "world-state.json"

type savedMap struct {
	ID      int    `json:"id"`
	Raw     []byte `json:"raw"`
	Terrain []byte `json:"terrain"`
	Record  []byte `json:"record"`
	Graphic []byte `json:"graphic"`
}

type worldState struct {
	Version  int        `json:"version"`
	SaveHash string     `json:"save_sha256"`
	Group    int        `json:"active_group"`
	Stock    byte       `json:"item_stock"`
	Maps     []savedMap `json:"maps"`
}

func (s *Scene) sessionBlock(id int) (*assets.Block, error) {
	if b := s.maps[id]; b != nil {
		return b, nil
	}
	return s.rom.BlockByID(id)
}

func (s *Scene) saveHash() string {
	h := sha256.New()
	h.Write(s.save.Plain)
	h.Write(s.save.Tail)
	return hex.EncodeToString(h.Sum(nil))
}

func (s *Scene) encodeWorldState() ([]byte, error) {
	maps := make(map[int]*assets.Block, len(s.maps)+1)
	for id, b := range s.maps {
		maps[id] = b
	}
	maps[s.blockID] = s.world.Block
	ids := make([]int, 0, len(maps))
	for id := range maps {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	state := worldState{Version: 1, SaveHash: s.saveHash(), Group: s.groupID, Stock: s.itemStock}
	for _, id := range ids {
		b := maps[id]
		state.Maps = append(state.Maps, savedMap{id, b.Raw, b.Terrain, b.Record, b.Graphic})
	}
	return json.Marshal(state)
}

func (s *Scene) decodeWorldState(raw []byte) (map[int]*assets.Block, error) {
	maps := make(map[int]*assets.Block)
	if len(raw) == 0 {
		return maps, nil // WLQS v1 沒有地圖歷史。
	}
	if len(raw) > 16<<20 {
		return nil, fmt.Errorf("地圖存檔超過大小上限")
	}
	var state worldState
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(&state); err != nil {
		return nil, fmt.Errorf("解析地圖存檔：%w", err)
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return nil, fmt.Errorf("地圖存檔尾端有多餘資料")
	}
	if state.Version != 1 || state.SaveHash != s.saveHash() || len(state.Maps) > 64 ||
		state.Group < 0 || state.Group >= PartyGroupCount || state.Stock > 3 {
		return nil, fmt.Errorf("地圖存檔版本或角色存檔不匹配")
	}
	for _, m := range state.Maps {
		if _, exists := maps[m.ID]; exists {
			return nil, fmt.Errorf("地圖 %d 重複", m.ID)
		}
		b, err := s.rom.BlockByID(m.ID)
		if err != nil {
			return nil, err
		}
		if len(m.Raw) != len(b.Raw) || len(m.Terrain) != len(b.Terrain) ||
			len(m.Record) != len(b.Record) || len(m.Graphic) != len(b.Graphic) {
			return nil, fmt.Errorf("地圖 %d 的資料長度不合", m.ID)
		}
		if int(m.Raw[b.HeaderAt+0x2C]) != b.Dim || int(m.Raw[b.HeaderAt+0x30]) != b.Tileset {
			return nil, fmt.Errorf("地圖 %d 的結構不合", m.ID)
		}
		for _, v := range m.Terrain {
			if v > 15 {
				return nil, fmt.Errorf("地圖 %d 含非法地形", m.ID)
			}
		}
		copy(b.Raw, m.Raw) // Header 是 Raw 的切片，副作用一起保留。
		copy(b.Terrain, m.Terrain)
		copy(b.Record, m.Record)
		copy(b.Graphic, m.Graphic)
		maps[m.ID] = b
	}
	return maps, nil
}

// 只在 decodeWorldState 成功後使用，v1 快速存檔回到預設第 0 組。
func (s *Scene) restoreWorldMetadata(raw []byte) {
	var state worldState
	_ = json.Unmarshal(raw, &state)
	s.groupID, s.itemStock = state.Group, state.Stock
}

func (s *Scene) storeWorldState(dir string) error {
	raw, err := s.encodeWorldState()
	if err != nil {
		return err
	}
	p := filepath.Join(dir, worldStateFile)
	if err := os.WriteFile(p+".tmp", raw, 0o644); err != nil {
		return err
	}
	return os.Rename(p+".tmp", p)
}

func (s *Scene) loadWorldState(dir string) error {
	raw, err := os.ReadFile(filepath.Join(dir, worldStateFile))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if len(raw) == 0 {
		return fmt.Errorf("地圖存檔是空檔案")
	}
	maps, err := s.decodeWorldState(raw)
	if err != nil {
		return err
	}
	next := *s
	next.maps = maps
	next.restoreWorldMetadata(raw)
	if err := next.reloadFromSave(); err != nil {
		return err
	}
	*s = next
	return nil
}
