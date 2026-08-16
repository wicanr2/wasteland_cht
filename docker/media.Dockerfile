# 音訊與影片的工具 image：MIDI 渲染（MT-32 與 SoundFont 兩條路）＋ ffmpeg ＋ ImageMagick。
#
# 為什麼合成一個 image：每次 `apt install ffmpeg imagemagick fonts-noto-cjk` 要拉 ~200MB，
# 一輪合成大半時間花在安裝上（KB game-promo-video-ffmpeg 雷 #2）。建一次、之後零安裝。
#
#   docker build -f docker/media.Dockerfile -t wasteland-media:latest .
#
# munt 不在 Debian 套件庫，從原始碼編 mt32emu_smf2wav（SMF → WAV）。
# **ROM 不進 image**：MT-32／CM-32L ROM 是 Roland 的韌體，執行時唯讀掛載進來。
FROM debian:bookworm-slim

RUN apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends \
        ca-certificates git cmake g++ make pkg-config \
        ffmpeg imagemagick fonts-noto-cjk vorbis-tools \
        fluidsynth fluid-soundfont-gm \
    && rm -rf /var/lib/apt/lists/*

# mt32emu_smf2wav 要 GLIB2（CMakeLists.txt 的 find_package(GLIB2)）與 libsndfile。
RUN apt-get update -qq && DEBIAN_FRONTEND=noninteractive apt-get install -y -qq --no-install-recommends \
        libglib2.0-dev libsndfile1-dev \
    && rm -rf /var/lib/apt/lists/*

# munt：mt32emu 函式庫 ＋ smf2wav 轉檔器
RUN git clone --depth 1 https://github.com/munt/munt /src/munt \
    && cmake -S /src/munt/mt32emu -B /build/mt32emu -DCMAKE_BUILD_TYPE=Release \
    && cmake --build /build/mt32emu -j2 && cmake --install /build/mt32emu \
    && ldconfig \
    && cmake -S /src/munt/mt32emu_smf2wav -B /build/smf2wav -DCMAKE_BUILD_TYPE=Release \
    && cmake --build /build/smf2wav -j2 && cmake --install /build/smf2wav \
    && ldconfig \
    && rm -rf /src /build

# ImageMagick 預設 policy 會擋 `@` 讀檔（KB 雷 #5）。只放寬本地檔讀取。
RUN sed -i 's/rights="none" pattern="@\*"/rights="read" pattern="@*"/' /etc/ImageMagick-6/policy.xml || true

WORKDIR /work
