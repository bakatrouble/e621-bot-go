FROM golang:1.26-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go \
    GOCACHE=/go/cache GOPATH=/go/path go mod download
RUN apk add --no-cache file-dev gcc musl-dev
COPY . .
RUN --mount=type=cache,target=/go \
    CGO_ENABLED=1 GOOS=linux GOCACHE=/go/cache GOPATH=/go/path go build -v -o /e621-bot-go

FROM alpine:latest
LABEL org.opencontainers.image.source=https://github.com/bakatrouble/e621-bot-go
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add --no-cache file-dev ffmpeg
WORKDIR /
COPY --from=builder /e621-bot-go /e621-bot-go
RUN mkdir /cache /logs && chown appuser:appgroup /cache /logs
USER appuser:appgroup
ENTRYPOINT ["/e621-bot-go"]
