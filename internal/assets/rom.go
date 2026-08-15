// Package assets 把玩家自備的原版《Wasteland》資料解成程式可用的形式。
//
// 這一層是**純解碼**：不認識 Ebiten、不認識遊戲規則，輸出 []byte、string 與
// image.RGBA（docs/spec/01-assets-and-formats.md）。
//
// 兩個貫穿整包的紀律：
//
//   - 載入前驗 SHA-256。用錯版本的症狀是「解出垃圾但不報錯」，
//     雜湊是唯一擋得住的地方。
//   - 未解的位元組原樣保留（Block.Raw）。存檔策略是改寫不是重建，
//     沒讀懂的區域一個 byte 都不能動。
package assets

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// KnownFiles 是原版 20 個檔案的 SHA-256（docs/re/01）。
// 這裡只放雜湊與長度，不含任何原版內容。
var KnownFiles = map[string]struct {
	Size int64
	Hash string
}{
	"wl.exe":         {62549, "098aef9b4fe4fea3b8d0d134f82fe11a6dac608839ebd175e168cf0271b93b4f"},
	"wla.bin":        {4209, "85ab88dd9ab6a67390d6fc6dcba1ba4bfc6801a5d0b9cecf074d263c4d9e0c62"},
	"game1":          {159429, "44521eef76fab492d38c0c0bcb171be3f06a4063ce1e0f49175f0b7370a7841d"},
	"game2":          {172235, "6de4df6dd1af0a0210c2d6e282ecd1327f2acc7440c955f65045fd6e8a135c47"},
	"allpics1":       {105866, "a1bf92f9b037023d99ee4e4bd22547e1fce176608c73494d707676b7d7ce6044"},
	"allpics2":       {133433, "b39cb802cfda1deacb5aa575c076bdcb5999bc978794f030d63c28a3642f4a36"},
	"allhtds1":       {34307, "c01d0d23d9d073be3a15a8d92fbaed8e9c117ad2fe42e669f8d39b2d6dbc8c42"},
	"allhtds2":       {39230, "c0b2d0299d28010db37d0c7714d2879c32ead0f31ba35fdb01c183defe0ec11f"},
	"end.cpa":        {21027, "c48197e4b4d9fb81c39ca69e24fd1228cab90647ca6a2e20d809ecccb01ae731"},
	"title.pic":      {18432, "5169de0bd7efad4bf12ba8135744d9daeb1939bb7f01bd5ca5126e626d223800"},
	"colorf.fnt":     {5504, "db34077fe0bcc331734ba660605f8ffe4da33bf6c97794f4c712286eb0430d51"},
	"ic0_9.wlf":      {1280, "d8bbeae054a25852817841b905ae093fb104587e89d90749fb8fd9ec6ca38ddc"},
	"masks.wlf":      {320, "6f355ad5841f050e2af7f353081f53e2f876b26d6f6970edf25ca8d3c2b17f66"},
	"curs":           {2048, "6ffbe904fc19ff8d355146fa5eb51c611e242cfa9f35538643414175d1cbaf7e"},
	"transtbl":       {800, "1d1a265562bf2f0ecb0bb4add32015e66eae7510b93499e707dfc9c8c36d91ba"},
	"paragraphs.txt": {72771, "ba50b061a0ed4326518bf92fe071e77ea8010d44ddaf889eff434be6b3e8eb92"},
	"manual.txt":     {53322, "4b222c7dc22229bffce83455989a1d164182f40396ffcf54a9445c09a2bc342e"},
	"readme.txt":     {164, "9ff6cc25057bb3affe65e32ab4012909fca55621328b511a49d540f46d3ce54b"},
	"info":           {2, "7e46dde720f00e74467c313a1142b572a18a5f03561bc08d6633de9a09d9eaa6"},
}

// Rom 是一份驗過雜湊的原版資料。
type Rom struct {
	dir   string
	files map[string][]byte

	// image 是解包並合成 overlay 之後的分析映像（tools/unpack_exepack.py ＋
	// apply_overlay.py 的產物）。字型、字串表與各種常數表都在裡面，
	// 但它是本專案的產物、不隨專案散布，所以要另外載入。
	image []byte
}

// Open 讀入一個目錄裡的原版檔案並逐一驗 SHA-256。
//
// 檔名一律以小寫比對——原版是 DOS 大寫檔名，解壓工具給的大小寫不一定一致。
func Open(dir string) (*Rom, error) { return open(dir, true) }

// OpenModified 開一個**已經被改過**的資料目錄副本，跳過 SHA-256 驗證。
//
// 只有一個正當用途：`cmd/wl-save` 這種「寫回存檔再讓原版讀」的驗收流程——
// 寫過一次之後檔案雜湊當然就不對了，再驗下去這條路只能走一次。
// ⚠ 一般載入一律用 Open。**跳過驗證等於放棄「拿到的是正版資料」這個保證**，
// 所以這支不接受原版目錄以外的任何用法，錯了要看得出來是誰跳過的。
func OpenModified(dir string) (*Rom, error) { return open(dir, false) }

func open(dir string, verify bool) (*Rom, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("讀取原版資料目錄：%w", err)
	}

	rom := &Rom{dir: dir, files: make(map[string][]byte, len(KnownFiles))}
	var problems []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := strings.ToLower(entry.Name())
		want, known := KnownFiles[name]
		if !known {
			continue // 目錄裡的其他檔案不管，但也不會被用到
		}
		data, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			problems = append(problems, fmt.Sprintf("%s：讀不到（%v）", name, err))
			continue
		}
		if verify {
			sum := sha256.Sum256(data)
			if got := hex.EncodeToString(sum[:]); got != want.Hash {
				problems = append(problems,
					fmt.Sprintf("%s：SHA-256 是 %s，應該是 %s", name, got, want.Hash))
				continue
			}
		}
		rom.files[name] = data
	}

	for name := range KnownFiles {
		if _, ok := rom.files[name]; !ok {
			if !containsProblem(problems, name) {
				problems = append(problems, fmt.Sprintf("%s：缺少", name))
			}
		}
	}

	if len(problems) > 0 {
		sort.Strings(problems)
		return nil, fmt.Errorf("原版資料驗證失敗（%d 項）：\n  %s",
			len(problems), strings.Join(problems, "\n  "))
	}
	return rom, nil
}

func containsProblem(problems []string, name string) bool {
	for _, p := range problems {
		if strings.HasPrefix(p, name+"：") {
			return true
		}
	}
	return false
}

// File 回傳某個原版檔案的內容（唯讀，呼叫端不得修改）。
func (r *Rom) File(name string) ([]byte, error) {
	data, ok := r.files[strings.ToLower(name)]
	if !ok {
		return nil, fmt.Errorf("沒有這個原版檔案：%s", name)
	}
	return data, nil
}

// LoadImage 載入解包合成映像（wl.merged.exe）。
//
// 字型、九張字串表與資源定址用的常數表都在執行檔裡，而原版的 wl.exe 是
// EXEPACK 打包的，位址對不上——所以要用工具解包後的映像（docs/re/02）。
// 這份映像是本專案的產物，不隨專案散布。
func (r *Rom) LoadImage(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("載入分析映像：%w", err)
	}
	if len(data) < 0x40 || data[0] != 'M' || data[1] != 'Z' {
		return fmt.Errorf("%s 不是 MZ 執行檔", path)
	}
	r.image = data
	return nil
}

// 位址換算（docs/re/00-master-index.md §1）。
const (
	dsBase    = 0x1CE20 // seg002：一般變數的線性基底
	seg003    = 0x2AE20 // 素材段的線性基底
	loadBase  = 0x10000 // 映像載入基底
	imageHdr  = 8       // MZ header 的 e_cparhdr 欄位位移
	paragraph = 16

	tblSkillData = 0xBA20 // 技能資料表（docs/re/32 §2）
	skillCount   = 36
)

// fileOffset 把線性位址換成分析映像裡的檔案位移。
func (r *Rom) fileOffset(linear int) (int, error) {
	if r.image == nil {
		return 0, fmt.Errorf("還沒載入分析映像（先呼叫 LoadImage）")
	}
	headerBytes := int(le16(r.image, imageHdr)) * paragraph
	off := linear - loadBase + headerBytes
	if off < 0 || off >= len(r.image) {
		return 0, fmt.Errorf("線性位址 %#x 換算出的檔案位移 %#x 超出映像", linear, off)
	}
	return off, nil
}

// SkillTableRaw 回傳技能資料表的原始 bytes（`ds:BA20h`，36 筆 × 2，docs/re/32 §2）。
func (r *Rom) SkillTableRaw() ([]byte, error) {
	off, err := r.dsOffset(tblSkillData)
	if err != nil {
		return nil, err
	}
	if off+skillCount*2 > len(r.image) {
		return nil, fmt.Errorf("技能資料表超出映像")
	}
	return r.image[off : off+skillCount*2], nil
}

// dsOffset 把 ds: 位移換成分析映像裡的檔案位移。
func (r *Rom) dsOffset(off int) (int, error) { return r.fileOffset(dsBase + off) }

// 音效資料段（docs/re/44 §5）。九首音效的表與位元組碼都在這裡，
// 不在任何外部檔案裡。
const (
	seg005     = 0x39200
	seg005Size = 0x360
)

// AudioData 取出音效段的原始 bytes（864 bytes）。
// 純解碼層只負責把它切出來，怎麼解讀是 internal/audio 的事。
func (r *Rom) AudioData() ([]byte, error) {
	off, err := r.fileOffset(seg005)
	if err != nil {
		return nil, fmt.Errorf("音效資料段：%w", err)
	}
	if off+seg005Size > len(r.image) {
		return nil, fmt.Errorf("音效資料段超出映像（%#x + %#x）", off, seg005Size)
	}
	out := make([]byte, seg005Size)
	copy(out, r.image[off:off+seg005Size])
	return out, nil
}

func le16(b []byte, at int) uint16 { return uint16(b[at]) | uint16(b[at+1])<<8 }
func le32(b []byte, at int) uint32 {
	return uint32(b[at]) | uint32(b[at+1])<<8 | uint32(b[at+2])<<16 | uint32(b[at+3])<<24
}

// kindIconTable 是 ds:AA17h：敵人種類 → 疊圖編號（docs/re/48 §3）。
const (
	kindIconTable = 0xAA17
	kindIconLen   = 6 // 種類 0–5；資料裡實際用到的是 1–5
)

// KindIconTable 取出敵人種類 → 疊圖編號的對照表。
//
// 讓規則層拿它與自己寫死的常數對過，**寫死的值才有人守著**——
// 抄一份數字進 Go 而沒有東西比對它，改壞了不會有人知道。
func (r *Rom) KindIconTable() ([]byte, error) {
	off, err := r.dsOffset(kindIconTable)
	if err != nil {
		return nil, fmt.Errorf("種類→疊圖表：%w", err)
	}
	if off+kindIconLen > len(r.image) {
		return nil, fmt.Errorf("種類→疊圖表超出映像（%#x）", off)
	}
	out := make([]byte, kindIconLen)
	copy(out, r.image[off:off+kindIconLen])
	return out, nil
}
