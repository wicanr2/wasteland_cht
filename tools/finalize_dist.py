#!/usr/bin/env python3
"""Recheck six actual packages and write the version's local delivery manifest."""
import hashlib
import json
import os
from pathlib import Path
import re
import subprocess
import sys

from verify_package import REPO, verify


def sha(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()


def finalize(version):
    if not re.fullmatch(r"v\.\d+\.\d+\.\d+-\d{8}", version):
        raise ValueError("必須指定完整版本")
    out = REPO / "dist-all" / version
    smoke = out / "smoke"
    smoke.mkdir(exist_ok=True)
    packages = []
    source_commits = set()
    for mode in ("patch", "full-local"):
        for platform in ("linux-x64", "windows-x64", "macos-universal"):
            suffix = ".AppImage" if platform == "linux-x64" else ".zip"
            package = out / mode / platform / f"wasteland-cht-{version}-{platform}-{mode}{suffix}"
            result = verify(package, version, mode, platform)
            source_commits.add(result["source_commit"])
            receipt = smoke / f"{mode}-{platform}-content.json"
            receipt.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
            packages.append({"path": str(package.relative_to(out)), "bytes": result["bytes"],
                             "sha256": result["sha256"], "mode": mode, "platform": platform,
                             "rights": result["rights"], "receipt": str(receipt.relative_to(out))})
            print(f"VERIFIED {mode} {platform}", flush=True)
    assert len(source_commits) == 1, f"同版封包來源提交不一致：{source_commits}"
    runtime_checks = {}
    for label, folder, mode, platform, needs_visual in (
        ("Linux 引擎包", "patch-linux-gui", "patch", "linux-x64", False),
        ("Linux 完整包 A／B", "full-local-linux-gui", "full-local", "linux-x64", True),
        ("Windows Wine 補充試玩 A／B", "full-local-windows-wine", "full-local", "windows-x64", True),
    ):
        receipt = smoke / folder / "result.json"
        check = {"status": "not-run"}
        if receipt.is_file():
            data = json.loads(receipt.read_text())
            package = next(p for p in packages if p["mode"] == mode and p["platform"] == platform)
            assert data["status"] == "passed" and data["version"] == version
            assert data["exit_code"] == 0 and data["quick_save"]["version"] == 2
            assert data.get("load_restores_exact_crop"), receipt
            assert data["package_sha256"] == package["sha256"], receipt
            assert data.get("modern_ui_verified") is True, receipt
            assert data.get("languages_verified") == ["zh-Hant", "zh-Hans"], receipt
            driver = receipt.parent / ("check_package_gui.sh" if platform == "windows-x64" else "driver.sh")
            assert sha(driver) == data["driver_sha256"], driver
            if needs_visual:
                assert data.get("visual_modes_verified") is True, receipt
            for screenshot in data["screenshots"]:
                assert sha(receipt.parent / screenshot["path"]) == screenshot["sha256"]
            check = {"status": "passed", "receipt": str(receipt.relative_to(out)), "sha256": sha(receipt)}
        runtime_checks[label] = check
    runtime_checks["Windows 原生實機"] = {"status": "not-run"}
    runtime_checks["macOS 原生實機"] = {"status": "not-run"}
    lines = [f"# Wasteland 荒野遊俠 {version}", "", "本機測試交付，未上傳或發布。", "",
             "| 平台 | 引擎與翻譯 | 本機完整版 |", "|---|---|---|"]
    for platform in ("linux-x64", "windows-x64", "macos-universal"):
        rows = [next(p for p in packages if p["platform"] == platform and p["mode"] == m)
                for m in ("patch", "full-local")]
        lines.append(f"| {platform} | [patch]({rows[0]['path']}) | [full-local]({rows[1]['path']}) |")
    lines += ["", "full-local 含原版、倚天字型、HD2 衍生圖片與本機配樂，只供本機使用，不可散布。",
              "patch 內建已授權圓體與 Noto，不含原版、倚天、HD2 圖片或配樂，玩家須自備合法原版。",
              "預設使用 18 像素圓體、細邊框與工具列；美術與字型可獨立切換。舊 HD 已移除。",
              "內含繁體與簡體中文；由工具列設定 F2 的語言按鈕或 L 切換。未完成的指令須先關閉。",
              "", "六包均完成內容、架構、版本與雜湊驗證。實際啟動檢查與畫面見 smoke/。",
              "Linux 容器／Wine 與原生 Windows／macOS 實機驗收分開記錄，不把交叉建置當成實機通過。",
              "", "| 實際啟動檢查 | 狀態 |", "|---|---|"]
    for label, check in runtime_checks.items():
        lines.append(f"| {label} | {'通過' if check['status'] == 'passed' else '未執行'} |")
    lines += ["", "工具列可選原版／A／B、圓體／Noto、16／18／20 字級、配樂及音量，並可直接存讀檔。",
              "F2／V／B 的既有熱鍵仍可使用；HD2 首次切換請等待畫面更新，避免連按。",
              "macOS 包包含 x86_64／arm64，部署目標 11.0；尚未做 Mac 實機驗收。",
              "", "來源與重現入口：專案 docs/packaging.md、tools/dist.sh、tools/finalize_dist.py。", ""]
    if (smoke / "hd2-ab-final.png").is_file():
        lines += ["", "[HD2 A／B 實際遊戲畫面](smoke/hd2-ab-final.png)", ""]
    if (smoke / "rounded-ui-review.png").is_file():
        lines += ["", "[新介面實際操作截圖](smoke/rounded-ui-review.png)", ""]
    (out / "README.md").write_text("\n".join(lines))
    inputs = [REPO / p for p in ("tools/package.sh", "tools/package_container.sh", "tools/pack_wrap.sh",
              "tools/verify_package.py", "tools/finalize_dist.py", "cmd/wasteland/main.go",
              "cmd/wl-setup/main.go", "internal/artpack/artpack.go", "internal/artpack/artpack_test.go",
              "internal/play/panels.go", "translations/zh-Hant.cat", "artwork/manifests/hd2-production.json", "LICENSE")]
    inputs += [REPO / p for p in ("internal/play/presentation.go", "internal/presentation/canvas.go",
               "internal/render/hires.go", "internal/ui/ui.go", "artwork/manifests/ui-rounded-approved.json",
               "tools/check_package_gui.sh", "tools/check_package_wine_server.sh")]
    inputs += [REPO / p for p in ("internal/play/locales.go", "tools/build_lang.py", "tools/build_paragraphs.py",
               "tools/make_zh_hans.py", "tools/localization-inputs.json", "translations/zh-Hans.cat",
               "translations/paragraphs-zh-Hant.cat", "translations/paragraphs-zh-Hans.cat")]
    inputs += sorted((REPO / "internal/presentation/fonts").iterdir())
    inputs += [REPO / "artpacks" / m / "manifest.json" for m in ("hd2-a", "hd2-b")]
    files = [{"path": str(p.relative_to(out)), "bytes": p.stat().st_size, "sha256": sha(p)}
             for p in sorted(out.rglob("*")) if p.is_file() and p.name != "SHA256SUMS.json"]
    manifest = {"schema": 1, "version": version, "source_commit": next(iter(source_commits)),
                "publication": "local-only; no tag or upload",
                "validation_commit": subprocess.check_output(["git", "-C", str(REPO), "rev-parse", "HEAD"], text=True).strip(),
                "validation_tracked_changes": subprocess.run(["git", "-C", str(REPO), "diff", "--quiet"]).returncode != 0,
                "packages": packages, "files": files,
                "runtime_checks": runtime_checks,
                "inputs": [{"path": str(p.relative_to(REPO)), "sha256": sha(p)} for p in inputs],
                "tool_images": {key: os.environ.get(key, "not supplied") for key in
                                ("WL_GO_IMAGE_ID", "WL_OSXCROSS_IMAGE_ID", "WL_PKG_IMAGE_ID", "WL_WINE_IMAGE_ID")}}
    (out / "SHA256SUMS.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n")
    print(f"PASS {version}: 6 packages, {len(files)} delivery files")


if __name__ == "__main__":
    finalize(sys.argv[1])
