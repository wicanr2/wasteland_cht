#!/usr/bin/env bash
# 把 site/ 發佈到 GitHub Pages（走 gh-pages 分支，**不用 GitHub Actions**）。
#
#	tools/publish_site.sh
#
# ⚠ **為什麼是這個做法**：Pages 的「Deploy from a branch」模式，資料夾只能選
# `/` 或 `/docs`（GitHub 的限制，API 給別的值直接回 422），而這個 repo 的
# `docs/` 放的是逆向筆記與規格，不是網站。用 Actions 可以只發佈 `site/`，
# 但那要 Actions 額度。所以改成：**把 `site/` 的內容推成 `gh-pages` 分支的根目錄**。
#
# 流程：重跑 build.py → 把 site/*.html 複製進 gh-pages 的 worktree → commit → push。
# master 不受影響；gh-pages 只有網頁，沒有歷史包袱。
#
# 邊界寫在腳本本體：worktree 用完就移除、Python 走 docker、不動 master 的檔案。

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WORK="$(mktemp -d)"
BRANCH=gh-pages

cleanup() {
    git -C "$ROOT" worktree remove --force "$WORK" 2>/dev/null || true
    rm -rf "$WORK"
}
trap cleanup EXIT

# 1. 重新產生網頁（HTML 是產物，來源是 docs/walkthrough/swm-005/*.md）
docker run --rm --log-opt max-size=10m --log-opt max-file=3 --network none \
    -v "$ROOT:/w" -w /w -u "$(id -u):$(id -g)" python:3.12-slim \
    python3 site/build.py

# 2. 自包含檢查：不引 CDN／web font，也沒有 <img>。
#    這一關擋的是「哪天有人手改 HTML」，不是裝飾。
if grep -rlE '<(img|script src="http|link[^>]*href="http)' "$ROOT"/site/*.html; then
    echo "site/ 有外部資源或圖片，違反自包含的規矩" >&2
    exit 1
fi

# 3. 準備 gh-pages 的 worktree（沒有這個分支就開一個沒有歷史的）
if git -C "$ROOT" show-ref --verify --quiet "refs/heads/$BRANCH"; then
    git -C "$ROOT" worktree add "$WORK" "$BRANCH" >/dev/null
else
    git -C "$ROOT" worktree add --detach "$WORK" >/dev/null
    git -C "$WORK" checkout --orphan "$BRANCH" >/dev/null
    git -C "$WORK" rm -rq --cached . 2>/dev/null || true
    find "$WORK" -mindepth 1 -maxdepth 1 ! -name .git -exec rm -rf {} +
fi

# 4. 只放網頁。**先清空再複製**，不然刪掉的頁面會留在線上。
find "$WORK" -mindepth 1 -maxdepth 1 ! -name .git -exec rm -rf {} +
cp "$ROOT"/site/*.html "$WORK"/
# .nojekyll：不讓 GitHub 拿 Jekyll 再處理一次，我們給的就是成品。
touch "$WORK/.nojekyll"

git -C "$WORK" add -A
if git -C "$WORK" diff --cached --quiet; then
    echo "網頁沒有變動，不用發佈。"
    exit 0
fi
git -C "$WORK" commit -q -m "發佈網站（來源 $(git -C "$ROOT" rev-parse --short HEAD)）"
git -C "$WORK" push -q origin "$BRANCH"

echo "---"
echo "已發佈：https://wicanr2.github.io/wasteland_cht/"
