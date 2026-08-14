# Wasteland（1988, Interplay／EA）逆向工程參考環境
# DOSBox 0.74-3（Debian bookworm 套件版）＋ headless 截圖與自動化輸入。
#
# 用途：當「原版實際行為」的 oracle。IDA 讀出來的東西要實機確認時用它，
# **不是 IDA 的替代品**（CLAUDE.md §1）。
#
# 全程 headless：Xvfb 起虛擬 framebuffer、xdotool 送鍵、ImageMagick 截圖。
# image 本身不含任何遊戲資產——資料由 tools/dosbox.sh bind mount 進來。
FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y --no-install-recommends \
        dosbox \
        xvfb \
        x11-apps \
        imagemagick \
        xdotool \
        procps \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /work

COPY entrypoint.sh /usr/local/bin/dosbox-entrypoint.sh
RUN chmod +x /usr/local/bin/dosbox-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/dosbox-entrypoint.sh"]
