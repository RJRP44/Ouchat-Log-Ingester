# Build
FROM golang:1.26.4 AS build

WORKDIR /app

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY *.go ./

RUN go build -o /ouchat-log-ingester

# Deploy
FROM gcr.io/distroless/base-debian13

WORKDIR /

COPY --from=build /ouchat-log-ingester /ouchat-log-ingester

EXPOSE 3010 8080

ENTRYPOINT ["/ouchat-log-ingester"]