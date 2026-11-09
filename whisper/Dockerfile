# ═══ تحويل الصوت لنص — داخل سيرفرنا ═══
# قرار (ع) 10-05: الفويس ما يطلع من السيرفر. هالحاوية تشغّل whisper.cpp
# (مفتوح المصدر) على الشبكة الداخلية بس — ماكو منفذ مفتوح برّا.
#
# حجم النموذج: build-arg WHISPER_MODEL
#   small (افتراضي) ≈ ٥٠٠MB، يحتاج ~١GB رام — أدق بالعربي.
#   base            ≈ ١٥٠MB، ~٥٠٠MB رام — أسرع وأضعف.
FROM debian:bookworm-slim AS build
ARG WHISPER_VERSION=v1.7.6
ARG WHISPER_MODEL=small
RUN apt-get update && apt-get install -y --no-install-recommends \
      build-essential cmake git ca-certificates curl && rm -rf /var/lib/apt/lists/*
WORKDIR /src
RUN git clone --depth 1 --branch ${WHISPER_VERSION} https://github.com/ggml-org/whisper.cpp.git .
RUN cmake -B build -DCMAKE_BUILD_TYPE=Release -DWHISPER_BUILD_TESTS=OFF -DGGML_NATIVE=OFF -DBUILD_SHARED_LIBS=OFF \
 && cmake --build build -j"$(nproc)" --target whisper-server
RUN ./models/download-ggml-model.sh ${WHISPER_MODEL} \
 && mv models/ggml-${WHISPER_MODEL}.bin /model.bin

FROM debian:bookworm-slim
# ffmpeg: الفويس يجي webm/ogg/m4a من المتصفحات، والخادم يحوّله لـwav (--convert).
RUN apt-get update && apt-get install -y --no-install-recommends ffmpeg ca-certificates \
 && rm -rf /var/lib/apt/lists/* && useradd -u 1000 -m whisper
COPY --from=build /src/build/bin/whisper-server /usr/local/bin/whisper-server
COPY --from=build /model.bin /models/model.bin
USER whisper
WORKDIR /tmp
EXPOSE 8080
ENTRYPOINT ["whisper-server", "-m", "/models/model.bin", "--host", "0.0.0.0", "--port", "8080", "--convert", "-l", "ar", "-t", "2"]
