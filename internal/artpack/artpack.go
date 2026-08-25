// Package artpack loads complete, prevalidated modern-art bundles.
// It deliberately does not know Ebiten, the original ROM, or game rules.
package artpack

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"image"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
)

type Mode string

const (
	ModeFaithfulHD Mode = "faithful-hd"
	ModeReimagined Mode = "reimagined"
)

type Canvas struct {
	Width       int  `json:"width"`
	Height      int  `json:"height"`
	Responsive  bool `json:"responsive"`
	MaxViewCols int  `json:"max_view_cols,omitempty"`
	MaxViewRows int  `json:"max_view_rows,omitempty"`
}

type Asset struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`
	Path   string `json:"path"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
	SHA256 string `json:"sha256"`
}

type Manifest struct {
	Schema int     `json:"schema"`
	ID     string  `json:"id"`
	Mode   Mode    `json:"mode"`
	Canvas Canvas  `json:"canvas"`
	Assets []Asset `json:"assets"`
}

type Bundle struct {
	Dir      string
	Manifest Manifest
	byID     map[string]Asset
}

func (b *Bundle) Asset(id string) (Asset, bool) {
	a, ok := b.byID[id]
	return a, ok
}

// Load performs the prepare half of an atomic theme switch. Callers commit by
// replacing their current *Bundle only after this function succeeds.
func Load(dir string) (*Bundle, error) {
	raw, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return nil, fmt.Errorf("讀取美術包 manifest：%w", err)
	}
	var m Manifest
	d := json.NewDecoder(strings.NewReader(string(raw)))
	d.DisallowUnknownFields()
	if err := d.Decode(&m); err != nil {
		return nil, fmt.Errorf("解析美術包 manifest：%w", err)
	}
	if err := validateManifest(m); err != nil {
		return nil, err
	}
	b := &Bundle{Dir: dir, Manifest: m, byID: make(map[string]Asset, len(m.Assets))}
	for _, a := range m.Assets {
		if _, exists := b.byID[a.ID]; exists {
			return nil, fmt.Errorf("美術包資產 ID 重複：%s", a.ID)
		}
		if err := validateAsset(dir, a); err != nil {
			return nil, fmt.Errorf("資產 %s：%w", a.ID, err)
		}
		b.byID[a.ID] = a
	}
	return b, nil
}

func validateManifest(m Manifest) error {
	if m.Schema != 1 {
		return fmt.Errorf("不支援的美術包 schema：%d", m.Schema)
	}
	if m.ID == "" {
		return fmt.Errorf("美術包缺少 id")
	}
	if m.Mode != ModeFaithfulHD && m.Mode != ModeReimagined {
		return fmt.Errorf("不支援的美術模式：%q", m.Mode)
	}
	if m.Canvas.Width <= 0 || m.Canvas.Height <= 0 {
		return fmt.Errorf("畫布尺寸必須大於零")
	}
	if m.Mode == ModeFaithfulHD && m.Canvas.Responsive {
		return fmt.Errorf("faithful-hd 必須維持固定 4:3 畫布")
	}
	if m.Mode == ModeReimagined && !m.Canvas.Responsive {
		return fmt.Errorf("reimagined 必須使用響應式畫布")
	}
	return nil
}

func validateAsset(dir string, a Asset) error {
	if a.ID == "" || a.Kind == "" || a.Path == "" {
		return fmt.Errorf("id、kind 與 path 不得為空")
	}
	if filepath.IsAbs(a.Path) || filepath.Clean(a.Path) != a.Path || strings.HasPrefix(a.Path, ".."+string(filepath.Separator)) || a.Path == ".." {
		return fmt.Errorf("path 必須是包內的乾淨相對路徑：%q", a.Path)
	}
	if strings.ToLower(filepath.Ext(a.Path)) != ".png" {
		return fmt.Errorf("第一階段只接受 PNG：%q", a.Path)
	}
	if a.Width <= 0 || a.Height <= 0 {
		return fmt.Errorf("尺寸必須大於零")
	}
	want, err := hex.DecodeString(a.SHA256)
	if err != nil || len(want) != sha256.Size {
		return fmt.Errorf("SHA-256 格式錯誤")
	}
	p := filepath.Join(dir, a.Path)
	raw, err := os.ReadFile(p)
	if err != nil {
		return fmt.Errorf("讀取 %s：%w", a.Path, err)
	}
	got := sha256.Sum256(raw)
	if !strings.EqualFold(hex.EncodeToString(got[:]), a.SHA256) {
		return fmt.Errorf("SHA-256 不符")
	}
	f, err := os.Open(p)
	if err != nil {
		return err
	}
	defer f.Close()
	cfg, format, err := image.DecodeConfig(f)
	if err != nil {
		return fmt.Errorf("解碼 PNG：%w", err)
	}
	if format != "png" || cfg.Width != a.Width || cfg.Height != a.Height {
		return fmt.Errorf("實際為 %s %d×%d，manifest 為 PNG %d×%d", format, cfg.Width, cfg.Height, a.Width, a.Height)
	}
	return nil
}
