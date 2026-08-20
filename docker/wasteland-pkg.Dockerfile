# 封裝用的 image：在 Go build image 上補 AppImage 需要的 squashfs 工具。
#
# 為什麼要另一個 image：`wasteland-go` 只裝編譯要的東西，而 AppImage 的內層
# 是一份 squashfs。zip 不另裝——Python 的 `zipfile` 才控制得到 UTF-8 旗標
# 與 Unix 權限位元（`tools/pack_zip.py`）。
#
# 建置（這一步會連網裝套件，之後封裝仍然 --network none）：
#
#   docker build --network host \
#     -t wasteland-pkg:1.24-x11 -f docker/wasteland-pkg.Dockerfile docker/
#
# 版本刻意寫死：換版本就換 tag，不要覆蓋既有 tag（`CLAUDE.md` §1.1 的做法）。
FROM wasteland-go:1.24-x11
RUN apt-get update \
 && apt-get install -y --no-install-recommends squashfs-tools \
 && rm -rf /var/lib/apt/lists/*
