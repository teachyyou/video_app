# ---------- build: собираем Go бинарь ----------
FROM golang:1.25.1 AS build
LABEL name="video_backend"

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY ./src ./src
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/app ./src/cmd/app

# ---------- ffmpeg: берём ffmpeg/ffprobe с NVENC ----------
FROM jrottenberg/ffmpeg:6.1-nvidia AS ffmpeg-nv

# ---------- run: CUDA runtime с библиотеками NVIDIA ----------
FROM nvidia/cuda:12.3.1-runtime-ubuntu22.04 AS run

ENV DEBIAN_FRONTEND=noninteractive

# базовые пакеты
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata \
    libgomp1 \
    libexpat1 \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Копируем ffmpeg/ffprobe и нужные либы из стадии с NVENC
COPY --from=ffmpeg-nv /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg
COPY --from=ffmpeg-nv /usr/local/bin/ffprobe /usr/local/bin/ffprobe
COPY --from=ffmpeg-nv /usr/local/lib/ /usr/local/lib/
ENV LD_LIBRARY_PATH="/usr/local/lib:${LD_LIBRARY_PATH}"

# Некорневой пользователь
RUN useradd -r -s /usr/sbin/nologin app
USER app

# Приложение и ресурсы
COPY --from=build /app/app /app/app
COPY ./resources /app/resources

EXPOSE 8080
ENTRYPOINT ["/app/app"]
