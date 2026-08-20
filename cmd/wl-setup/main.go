// wl-setup 從玩家自己的原版資料產生遊戲要用的合成映像 `wl.merged.exe`。
//
//	wl-setup -rom <原版資料目錄> -out build/wl.merged.exe
//
// 兩步：EXEPACK 解包（`docs/re/02`）＋ 疊 `wla.bin`（`docs/re/03` §5）。
// 兩步都在 `internal/exepack`，與 `tools/unpack_exepack.py` ＋
// `tools/apply_overlay.py` 產生**一模一樣的位元組**。
//
// ⚠ **這一步不能替玩家做。** 合成映像是原版執行檔的衍生物，不隨專案散布；
// 玩家用自己那份合法原版跑一次，產物留在自己機器上。
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wicanr2/wasteland_cht/internal/exepack"
)

// knownMerged 是 1988 年那份 DOS 版跑出來的合成映像雜湊（`CLAUDE.md` §1.2）。
// 對不上不是錯誤——別的發行版本、別的磁片映像本來就會不一樣，
// 所以只提醒不中止。
const knownMerged = "cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118"

func main() {
	rom := flag.String("rom", ".", "原版資料目錄（要含 wl.exe 與 wla.bin）")
	out := flag.String("out", "build/wl.merged.exe", "輸出的合成映像")
	quiet := flag.Bool("quiet", false, "只在出錯時說話")
	flag.Parse()

	if err := run(*rom, *out, *quiet); err != nil {
		fmt.Fprintln(os.Stderr, "wl-setup:", err)
		os.Exit(1)
	}
}

func run(rom, out string, quiet bool) error {
	// 找檔案時大小寫都試：原版磁片上是大寫，解壓工具常常會轉成小寫。
	exePath, err := find(rom, "wl.exe")
	if err != nil {
		return err
	}
	overlayPath, err := find(rom, "wla.bin")
	if err != nil {
		return err
	}
	packed, err := os.ReadFile(exePath)
	if err != nil {
		return fmt.Errorf("讀 %s：%w", exePath, err)
	}
	overlay, err := os.ReadFile(overlayPath)
	if err != nil {
		return fmt.Errorf("讀 %s：%w", overlayPath, err)
	}

	unpacked, st, err := exepack.Unpack(packed)
	if err != nil {
		return fmt.Errorf("解包 %s：%w", exePath, err)
	}
	merged, err := exepack.ApplyOverlay(unpacked, overlay)
	if err != nil {
		return fmt.Errorf("疊 overlay：%w", err)
	}

	if dir := filepath.Dir(out); dir != "" && dir != "." {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("建立輸出目錄：%w", err)
		}
	}
	if err := os.WriteFile(out, merged, 0o644); err != nil {
		return fmt.Errorf("寫 %s：%w", out, err)
	}
	if quiet {
		return nil
	}

	sum := sha256.Sum256(merged)
	digest := hex.EncodeToString(sum[:])
	fmt.Printf("解包：%d 個命令（fill %d／copy %d），%d → %d bytes\n",
		st.Commands, st.Fills, st.Copies, st.PackedBytes, st.UnpackedBytes)
	fmt.Printf("合成映像：%s（%d bytes）\nSHA-256：%s\n", out, len(merged), digest)
	if digest != knownMerged {
		fmt.Printf("提示：與本專案分析用的那份（%s…）不同。\n"+
			"      不同的發行版本本來就會不一樣，遊戲照樣跑得動；\n"+
			"      有對不上的畫面或文字時，這一行是第一個要提的線索。\n",
			knownMerged[:16])
	}
	return nil
}

// find 在目錄裡找檔案，大小寫不敏感。
func find(dir, name string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", fmt.Errorf("讀不到原版資料目錄 %s：%w", dir, err)
	}
	for _, e := range entries {
		if strings.EqualFold(e.Name(), name) {
			return filepath.Join(dir, e.Name()), nil
		}
	}
	return "", fmt.Errorf("原版資料目錄 %s 裡找不到 %s", dir, name)
}
