FROM golang:1.25.1 AS build
LABEL name="video_backend"

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY ./src ./src
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/app ./src/cmd/app

FROM debian:bookworm-slim AS run

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates tzdata \
    libgomp1 \
    libexpat1 \
    ffmpeg \
 && rm -rf /var/lib/apt/lists/*

WORKDIR /app

RUN useradd -r -s /usr/sbin/nologin app

COPY --from=build /app/app /app/app
COPY ./resources /app/resources

RUN chown -R app:app /app

USER app

EXPOSE 8080
ENTRYPOINT ["/app/app"]
