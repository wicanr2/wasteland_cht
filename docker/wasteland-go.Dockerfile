# Go 建置環境：golang ＋ Ebiten 在 Linux 需要的 X11／GL／音訊標頭。
#
# 為什麼要自己一個 image：Ebiten 的視窗層走 cgo（`internal/glfw`），
# 官方 golang image 沒有 X11 標頭，`go build` 會停在
# `fatal error: X11/Xlib.h: No such file or directory`。
#
# 建置（這一步會連網裝套件，之後 tools/go.sh 仍然 --network none）：
#
#   docker build --network host \
#     -t wasteland-go:1.24-x11 -f docker/wasteland-go.Dockerfile docker/
#
# 版本刻意寫死：換版本就換 tag，不要覆蓋既有 tag（`CLAUDE.md` §1.1 的做法）。
FROM golang:1.24-bookworm

RUN apt-get update && apt-get install -y --no-install-recommends \
        libx11-dev \
        libxrandr-dev \
        libxinerama-dev \
        libxcursor-dev \
        libxi-dev \
        libxxf86vm-dev \
        libgl1-mesa-dev \
        libasound2-dev \
        pkg-config \
    && rm -rf /var/lib/apt/lists/*
