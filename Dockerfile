# Build
FROM golang:1.26.4 AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY *.go ./

RUN go build -o /ouchat-log-ingester

FROM debian:13-slim

WORKDIR /

# Installation propre de FFmpeg certifié pour l'architecture de ton Pi
RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /

COPY --from=build /ouchat-log-ingester /ouchat-log-ingester

EXPOSE 3010 8080

ENTRYPOINT ["/ouchat-log-ingester"]