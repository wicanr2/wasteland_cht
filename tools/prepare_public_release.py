#!/usr/bin/env python3
"""Prepare patch-only Release notes/checksums in Docker; never upload anything."""
import hashlib
import json
from pathlib import Path
import re
import subprocess
import sys

ROOT = Path(__file__).resolve().parents[1]


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main():
    version = sys.argv[1]
    assert re.fullmatch(r"v\.\d+\.\d+\.\d+-\d{8}", version)
    out = ROOT / "dist-all" / version
    local = json.loads((out / "SHA256SUMS.json").read_text())
    assert local["version"] == version
    checks = local["runtime_checks"]
    for label in ("Linux 引擎包", "Linux 完整包 A／B", "Windows Wine 補充試玩 A／B"):
        assert checks[label]["status"] == "passed", label
    packages = [p for p in local["packages"] if p["mode"] == "patch"]
    assert len(packages) == 3
    assert {p["platform"] for p in packages} == {"linux-x64", "windows-x64", "macos-universal"}
    for p in packages:
        path = out / p["path"]
        assert p["rights"] == "engine-and-translations"
        assert path.stat().st_size == p["bytes"] and sha(path) == p["sha256"]
    target = subprocess.check_output(["git", "-C", str(ROOT), "rev-parse", "HEAD"], text=True).strip()
    # The tag may add verified documentation; shipped code and inputs must match.
    subprocess.run(["git", "-C", str(ROOT), "diff", "--exit-code", local["source_commit"], target,
                    "--", "cmd", "internal", "translations", "artpacks", "artwork", "go.mod", "go.sum"], check=True)
    public = {"schema":1, "version":version, "tag_commit":target,
              "build_source_commit":local["source_commit"], "rights":"patch-only; original data and HD2 art excluded",
              "verification":{k:v["status"] for k,v in checks.items()},
              "assets":[{"name":Path(p["path"]).name,"bytes":p["bytes"],"sha256":p["sha256"],
                         "platform":p["platform"]} for p in packages]}
    notes = f"""# Wasteland 荒野遊俠 {version}

繁簡中文重製版，提供 Windows x64、Linux x64 與 macOS x86_64／arm64。

## 新版成果

- 繁體／簡體中文，各 5,289 條一般文字與 162 段手札；工具列「設定 F2」中的語言按鈕或 L 可切換。
- 內建源泉圓體與 Noto 黑體，可選 16／18／20 字級；常駐工具列集中美術、配樂、音量與存讀檔。
- A／B 預設開啟 HD 原版銀灰框、直排 WASTELAND 與藍色計量表，F2 的 C 框線可關閉。原版美術介面維持。
- README 保留故事、台灣版本及當年攻略史，補齊新舊截圖、框線開關對照、素材總覽及全部 82 組 HD 事件圖對照。
- 重新錄製 75 秒推廣影片，加入 A／B 的 HD 框線與 F2／C 開關操作；README 提供成片預覽，含素材的影片僅留本機。

## 回報問題

- 已修正全隊死亡後無法重新開始，以及 F9 讀檔殘留死亡等畫面狀態。
- 已修正擊敗怪物後的遭遇寫回；清怪、門鎖與寶箱等地圖變更可跨地圖及存讀檔保留。
- 已修正目前分隊的遭遇／NPC 互動接線，JACKIE 雇用已抽測。
- 空地按 E 確認後執行空回合是原版行為，保留且不生成怪物。
- 門鎖與其他 NPC 的原始回報仍缺具體位置；16 筆門閘及 JACKIE 抽測不能代表所有回報已修復。

## 下載與執行

本次附件為可散布的引擎與翻譯包。玩家需自備合法原版資料，包含 wl.exe、wla.bin、game1、game2。
公開包不含原版資料、倚天字型、HD2 衍生圖包或配樂；README 的 A／B 展示使用本機完整素材。

- Windows：下載 windows-x64-patch.zip，解壓後執行「開始遊戲.bat」，依提示指定原版目錄。
- Linux：下載 linux-x64-patch.AppImage，賦予執行權限後啟動，指定原版目錄。
- macOS：下載 macos-universal-patch.zip，解壓後開啟「荒野遊俠.app」並選取原版目錄；未簽章或公證。
- SHA256SUMS.txt 可核對三個附件；release-manifest.json 記錄精確版本、建置／tag 提交與雜湊。

## 驗證與相容性

完整 Go 測試與靜態檢查通過。實際 Linux 公開／完整包與 Windows Wine 完整包已從正常標題抽測，
包括 A／B 框線、繁簡、字型／字級、設施、物品／手札、存讀檔及退出。
原生 Windows／macOS 及人工全程通關尚未驗收。
WLQS v2 可讀取舊 v1，舊程式無法讀取新版存檔。

[完整成果與實機截圖](https://github.com/wicanr2/wasteland_cht/blob/{version}/README.md)

程式與譯文採 RRSAL-1.0；原版素材權利屬各權利人。
"""
    folder = out / "patch"
    assert folder.is_dir() and folder.stat().st_uid == 1000
    text = "".join(p["sha256"] + "  " + Path(p["path"]).name + "\n" for p in packages)
    for name, content in (("SHA256SUMS.txt",text),("release-manifest.json",json.dumps(public,ensure_ascii=False,indent=2)+"\n"),("RELEASE-NOTES.md",notes)):
        path=folder/name
        assert not path.exists() or path.stat().st_uid==1000
        path.write_text(content,encoding="utf-8")
    print(f"READY: {version}; 3 patch packages; notes, public checksums and manifest; tag target {target}")


if __name__ == "__main__":
    main()
