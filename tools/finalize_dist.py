#!/usr/bin/env python3
"""Recheck six actual packages and write the version's local delivery manifest."""
import hashlib
import json
import os
from pathlib import Path
import re
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
    for mode in ("patch", "full-local"):
        for platform in ("linux-x64", "windows-x64", "macos-universal"):
            suffix = ".AppImage" if platform == "linux-x64" else ".zip"
            package = out / mode / platform / f"wasteland-cht-{version}-{platform}-{mode}{suffix}"
            result = verify(package, version, mode, platform)
            receipt = smoke / f"{mode}-{platform}-content.json"
            receipt.write_text(json.dumps(result, ensure_ascii=False, indent=2) + "\n")
            packages.append({"path": str(package.relative_to(out)), "bytes": result["bytes"],
                             "sha256": result["sha256"], "mode": mode, "platform": platform,
                             "rights": result["rights"], "receipt": str(receipt.relative_to(out))})
            print(f"VERIFIED {mode} {platform}", flush=True)
    lines = [f"# Wasteland 荒野遊俠 {version}", "", "本機測試交付，未上傳或發布。", "",
             "| 平台 | 引擎與翻譯 | 本機完整版 |", "|---|---|---|"]
    for platform in ("linux-x64", "windows-x64", "macos-universal"):
        rows = [next(p for p in packages if p["platform"] == platform and p["mode"] == m)
                for m in ("patch", "full-local")]
        lines.append(f"| {platform} | [patch]({rows[0]['path']}) | [full-local]({rows[1]['path']}) |")
    lines += ["", "full-local 含原版、字型、HD2 衍生圖片與本機配樂，只供本機使用，不可散布。",
              "patch 不含上述素材，玩家須自備合法原版。舊 HD 已移除，完整版支援 HD2 A／B。",
              "", "六包均完成內容、架構、版本與雜湊驗證。實際啟動檢查與畫面見 smoke/。",
              "Linux 容器／Wine 與原生 Windows／macOS 實機驗收分開記錄，不把交叉建置當成實機通過。",
              "", "來源與重現入口：專案 docs/packaging.md、tools/dist.sh、tools/finalize_dist.py。", ""]
    (out / "README.md").write_text("\n".join(lines))
    inputs = [REPO / p for p in ("tools/package.sh", "tools/package_container.sh", "tools/pack_wrap.sh",
              "tools/verify_package.py", "tools/finalize_dist.py", "cmd/wasteland/main.go",
              "cmd/wl-setup/main.go", "artwork/manifests/hd2-production.json", "LICENSE")]
    inputs += [REPO / "artpacks" / m / "manifest.json" for m in ("hd2-a", "hd2-b")]
    files = [{"path": str(p.relative_to(out)), "bytes": p.stat().st_size, "sha256": sha(p)}
             for p in sorted(out.rglob("*")) if p.is_file() and p.name != "SHA256SUMS.json"]
    manifest = {"schema": 1, "version": version, "publication": "local-only; no tag or upload",
                "packages": packages, "files": files,
                "inputs": [{"path": str(p.relative_to(REPO)), "sha256": sha(p)} for p in inputs],
                "tool_images": {key: os.environ.get(key, "not supplied") for key in
                                ("WL_GO_IMAGE_ID", "WL_OSXCROSS_IMAGE_ID", "WL_PKG_IMAGE_ID")}}
    (out / "SHA256SUMS.json").write_text(json.dumps(manifest, ensure_ascii=False, indent=2) + "\n")
    print(f"PASS {version}: 6 packages, {len(files)} delivery files")


if __name__ == "__main__":
    finalize(sys.argv[1])
