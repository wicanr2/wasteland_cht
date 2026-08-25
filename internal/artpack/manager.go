package artpack

import (
	"fmt"
	"path/filepath"
)

// Selection is the active presentation. Original has no Bundle because it is
// decoded from the player-supplied ROM by internal/assets.
type Selection struct {
	Mode   string
	Bundle *Bundle
}

// Manager owns the commit point for presentation switching. A failed load
// never mutates current.
type Manager struct {
	root    string
	current Selection
}

func NewManager(root string) *Manager {
	return &Manager{root: root, current: Selection{Mode: "original"}}
}

func (m *Manager) Current() Selection { return m.current }

func (m *Manager) Select(mode string) error {
	if mode == "original" {
		m.current = Selection{Mode: mode}
		return nil
	}
	if mode != string(ModeFaithfulHD) && mode != string(ModeReimagined) {
		return fmt.Errorf("未知美術模式：%q", mode)
	}
	if m.root == "" {
		return fmt.Errorf("未設定新版美術包目錄")
	}
	// Prepare: validate the whole candidate without touching current.
	b, err := Load(filepath.Join(m.root, mode))
	if err != nil {
		return fmt.Errorf("切換到 %s 失敗：%w", mode, err)
	}
	if string(b.Manifest.Mode) != mode {
		return fmt.Errorf("切換到 %s 失敗：manifest mode 是 %s", mode, b.Manifest.Mode)
	}
	// Commit: one assignment is the only mutation.
	m.current = Selection{Mode: mode, Bundle: b}
	return nil
}
