package artpack

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeBundle(t *testing.T, mutate func(*Manifest)) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "tile.png")
	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	im := image.NewRGBA(image.Rect(0, 0, 48, 48))
	im.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	if err := png.Encode(f, im); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	sum := sha256.Sum256(raw)
	m := Manifest{
		Schema: 1, ID: "test-hd", Mode: ModeFaithfulHD,
		Canvas: Canvas{Width: 960, Height: 600},
		Assets: []Asset{{ID: "tileset.0.0", Kind: "tile", Path: "tile.png", Width: 48, Height: 48, SHA256: hex.EncodeToString(sum[:])}},
	}
	if mutate != nil {
		mutate(&m)
	}
	b, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	return dir
}

func TestLoadValidBundle(t *testing.T) {
	b, err := Load(writeBundle(t, nil))
	if err != nil {
		t.Fatal(err)
	}
	if b.Manifest.Mode != ModeFaithfulHD {
		t.Fatalf("mode = %q", b.Manifest.Mode)
	}
	if _, ok := b.Asset("tileset.0.0"); !ok {
		t.Fatal("資產索引沒有 tileset.0.0")
	}
}

func TestLoadRejectsBadHash(t *testing.T) {
	dir := writeBundle(t, func(m *Manifest) { m.Assets[0].SHA256 = strings.Repeat("0", 64) })
	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "SHA-256 不符") {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadRejectsWrongDimensions(t *testing.T) {
	dir := writeBundle(t, func(m *Manifest) { m.Assets[0].Width = 47 })
	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "實際為") {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadRejectsTraversal(t *testing.T) {
	dir := writeBundle(t, func(m *Manifest) { m.Assets[0].Path = "../tile.png" })
	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "乾淨相對路徑") {
		t.Fatalf("err = %v", err)
	}
}

func TestLoadRejectsDuplicateID(t *testing.T) {
	dir := writeBundle(t, func(m *Manifest) { m.Assets = append(m.Assets, m.Assets[0]) })
	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "ID 重複") {
		t.Fatalf("err = %v", err)
	}
}

func TestModeCanvasContract(t *testing.T) {
	dir := writeBundle(t, func(m *Manifest) { m.Canvas.Responsive = true })
	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "固定 4:3") {
		t.Fatalf("err = %v", err)
	}
}

func TestManagerFailedSelectKeepsCurrent(t *testing.T) {
	m := NewManager(t.TempDir())
	if err := m.Select("faithful-hd"); err == nil {
		t.Fatal("缺少 bundle 應切換失敗")
	}
	if got := m.Current(); got.Mode != "original" || got.Bundle != nil {
		t.Fatalf("失敗後 current = %#v", got)
	}
}

func TestManagerCommitsOnlyValidatedBundle(t *testing.T) {
	root := t.TempDir()
	dir := writeBundle(t, nil)
	target := filepath.Join(root, "faithful-hd")
	if err := os.Rename(dir, target); err != nil {
		t.Fatal(err)
	}
	m := NewManager(root)
	if err := m.Select("faithful-hd"); err != nil {
		t.Fatal(err)
	}
	if got := m.Current(); got.Mode != "faithful-hd" || got.Bundle == nil {
		t.Fatalf("current = %#v", got)
	}
	if err := m.Select("unknown"); err == nil {
		t.Fatal("未知模式應失敗")
	}
	if got := m.Current(); got.Mode != "faithful-hd" || got.Bundle == nil {
		t.Fatalf("未知模式不應改 current：%#v", got)
	}
}

func TestCheckedInPrototypes(t *testing.T) {
	for _, tc := range []struct {
		dir  string
		mode Mode
	}{
		{"../../artpacks/prototype-faithful-hd", ModeFaithfulHD},
		{"../../artpacks/prototype-reimagined", ModeReimagined},
	} {
		b, err := Load(tc.dir)
		if err != nil {
			t.Fatalf("%s：%v", tc.dir, err)
		}
		if b.Manifest.Mode != tc.mode {
			t.Fatalf("%s mode = %q", tc.dir, b.Manifest.Mode)
		}
	}
}
