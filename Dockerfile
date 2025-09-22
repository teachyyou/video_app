FROM golang:latest AS build
LABEL authors="kudri"
LABEL name="video_backend"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY ./src src/

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o app ./src/cmd/app

FROM alpine:3.20 AS run

WORKDIR /app

RUN adduser -D -H app && apk add --no-cache ca-certificates tzdata

COPY --from=build /app/app /app/app
COPY ./resources resources/

EXPOSE 8080

USER app

ENTRYPOINT ["/app/app"]



