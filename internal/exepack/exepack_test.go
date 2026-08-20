package exepack

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

// 合成映像的雜湊（`CLAUDE.md` §1.2、`docs/re/02`）。Go 版與 Python 版必須
// 產生**一模一樣的位元組**——這一支就是那個對照。
const shaMerged = "cd5b07eaa55f1e1578caa1b05f0bd5331355cd119f387e61b1a8906738e78118"

func sha(b []byte) string {
	s := sha256.Sum256(b)
	return hex.EncodeToString(s[:])
}

// 原版資料是玩家自備的，不在版控裡；沒有就跳過。
func origDir(t *testing.T) string {
	t.Helper()
	dir := os.Getenv("WL_DATA")
	if dir == "" {
		dir = filepath.Join("..", "..", "workplace", "orig", "wastland")
	}
	if _, err := os.Stat(filepath.Join(dir, "wl.exe")); err != nil {
		t.Skipf("沒有原版資料（%s）", dir)
	}
	return dir
}

// 解包 ＋ 疊 overlay 的結果要與 Python 那條路完全相同。
//
// ⚠ **比的是最終雜湊不是「跑得起來」**：一份長度正確、內容錯位的映像
// 照樣載入得了，錯的只有後面每一個位址（`docs/re/02`）。
func TestUnpackMatchesTheKnownMergedImage(t *testing.T) {
	dir := origDir(t)
	packed, err := os.ReadFile(filepath.Join(dir, "wl.exe"))
	if err != nil {
		t.Fatalf("讀 wl.exe：%v", err)
	}
	unpacked, st, err := Unpack(packed)
	if err != nil {
		t.Fatalf("解包失敗：%v", err)
	}
	if st.Commands == 0 || st.UnpackedBytes == 0 {
		t.Fatalf("統計不合理：%+v", st)
	}
	t.Logf("解包：%d 個命令（fill %d／copy %d）、%d → %d bytes",
		st.Commands, st.Fills, st.Copies, st.PackedBytes, st.UnpackedBytes)

	overlay, err := os.ReadFile(filepath.Join(dir, "wla.bin"))
	if err != nil {
		t.Fatalf("讀 wla.bin：%v", err)
	}
	merged, err := ApplyOverlay(unpacked, overlay)
	if err != nil {
		t.Fatalf("疊 overlay 失敗：%v", err)
	}
	if got := sha(merged); got != shaMerged {
		t.Errorf("合成映像雜湊 ＝ %s\n預期 %s", got, shaMerged)
	}
}

// 壞掉的輸入要明確報錯，不要產出一份看起來像樣的錯映像。
func TestUnpackRejectsGarbage(t *testing.T) {
	for _, c := range []struct {
		name string
		data []byte
	}{
		{"空的", nil},
		{"不是 MZ", []byte("PK\x03\x04this is a zip file, not an exe")},
		{"MZ 但沒有 stub", append([]byte("MZ"), make([]byte, 0x40)...)},
	} {
		if _, _, err := Unpack(c.data); err == nil {
			t.Errorf("%s：應該報錯", c.name)
		}
	}
}

// overlay 比映像大就要拒絕（而不是靜靜截斷）。
func TestApplyOverlayChecksSize(t *testing.T) {
	base := make([]byte, 64)
	copy(base, "MZ")
	base[8] = 2 // header 32 bytes
	if _, err := ApplyOverlay(base, make([]byte, 1024)); err == nil {
		t.Error("overlay 超出映像範圍應該報錯")
	}
	out, err := ApplyOverlay(base, []byte{1, 2, 3})
	if err != nil {
		t.Fatalf("正常情況不該失敗：%v", err)
	}
	if out[32] != 1 || out[33] != 2 || out[34] != 3 {
		t.Errorf("overlay 沒有疊在 header 之後：%v", out[30:36])
	}
	if &out[0] == &base[0] {
		t.Error("不該就地改寫呼叫端的緩衝區")
	}
}
