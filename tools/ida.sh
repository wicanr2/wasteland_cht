#!/usr/bin/env bash
# IDA Pro 9.4 批次分析 wrapper（本專案唯一的逆向入口）
#
#   tools/ida.sh build                     建立 canonical .i64 + .asm
#   tools/ida.sh run <script.py> <out>     跑 IDAPython 匯出腳本（唯讀分析）
#   tools/ida.sh raw <idat 參數...>        直接下 idat 指令（除錯用）
#
# 邊界（寫在腳本本體，不是寫在對話裡）：
#   - 一律 --rm --network none，限制 CPU／記憶體／pids
#   - 工作區唯讀掛載，只有輸出目錄可寫
#   - canonical .i64 不直接餵給腳本；先複製到容器暫存層，避免任何寫回
#   - 退出前把輸出檔 chown 回目前使用者，不留 root-owned 檔案
#   - 不做任何 docker prune／rmi；容器一律用完即刪

set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
IMAGE="${WL_IDA_IMAGE:-ida-pro-9.4-idapython:py312-v1}"
ANALYSIS="$ROOT/workplace/analysis/ida94"
TARGET="${WL_IDA_TARGET:-$ROOT/workplace/orig/wastland/wl.exe}"
DB_NAME="$(basename "${TARGET%.*}").i64"

docker_run() {
  local outdir="$1"; shift
  mkdir -p "$outdir"
  docker run --rm \
    --network none --memory 2g --cpus 2 --pids-limit 256 \
    --log-opt max-size=10m --log-opt max-file=3 \
    -e OUTPUT_UID="$(id -u)" -e OUTPUT_GID="$(id -g)" \
    -v "$ROOT:/workspace:ro" \
    -v "$outdir:/output" \
    -w /output "$IMAGE" \
    sh -lc "$*"
}

cmd="${1:-}"; shift || true

case "$cmd" in
  build)
    [ -f "$TARGET" ] || { echo "找不到分析目標：$TARGET" >&2; exit 1; }
    rel="${TARGET#"$ROOT"/}"
    docker_run "$ANALYSIS" "
      cp /workspace/$rel /tmp/target.exe &&
      idat -B -L/output/ida-batch.log -o/output/$DB_NAME /tmp/target.exe
      rc=\$?
      [ -f /tmp/target.asm ] && cp /tmp/target.asm /output/${DB_NAME%.i64}.asm
      chown \"\$OUTPUT_UID:\$OUTPUT_GID\" /output/* 2>/dev/null || true
      exit \$rc"
    echo "→ $ANALYSIS/$DB_NAME"
    ;;

  run)
    script="${1:?用法：tools/ida.sh run <tools/ida/xxx.py> <輸出路徑>}"
    out="${2:?用法：tools/ida.sh run <tools/ida/xxx.py> <輸出路徑>}"
    [ -f "$ANALYSIS/$DB_NAME" ] || { echo "還沒建資料庫，先跑 tools/ida.sh build" >&2; exit 1; }
    outdir="$(cd "$(dirname "$out")" 2>/dev/null && pwd || { mkdir -p "$(dirname "$out")" && cd "$(dirname "$out")" && pwd; })"
    outfile="$(basename "$out")"
    docker_run "$outdir" "
      cp /workspace/workplace/analysis/ida94/$DB_NAME /tmp/$DB_NAME &&
      idat -A -L/output/$outfile.ida.log -S'/workspace/${script#"$ROOT"/} /output/$outfile' /tmp/$DB_NAME
      rc=\$?
      chown \"\$OUTPUT_UID:\$OUTPUT_GID\" /output/* 2>/dev/null || true
      [ -s /output/$outfile ] || { echo '輸出檔不存在或為空——腳本沒有真的跑完' >&2; exit 1; }
      exit \$rc"
    echo "→ $outdir/$outfile"
    ;;

  raw)
    docker_run "$ANALYSIS" "$*"
    ;;

  *)
    sed -n '2,8p' "${BASH_SOURCE[0]}"
    exit 1
    ;;
esac
