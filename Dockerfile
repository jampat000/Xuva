# Stage 1: build frontend
# The publish:go-static script builds the Svelte SPA and writes the output to
# server/internal/webapp/static-next, which the Go binary then embeds.
FROM node:20-alpine AS frontend-builder

WORKDIR /app/apps/web/svelte

COPY apps/web/svelte/package*.json ./
RUN npm ci

COPY apps/web/svelte/ ./

# Create the embed target directory before running the publish script.
RUN mkdir -p /app/server/internal/webapp/static-next

# Build SPA and publish into Go's embed directory in one step.
RUN npm run publish:go-static

# Stage 2: build Go binary with embedded frontend
FROM golang:1.26-alpine AS go-builder

ARG XUVA_VERSION=dev
ARG XUVA_COMMIT=unknown
ARG XUVA_BUILD_DATE=

RUN apk add --no-cache git

WORKDIR /app/server

COPY server/go.mod server/go.sum ./
RUN go mod download

COPY server/ ./

# Bring in the published frontend so go:embed can find it.
COPY --from=frontend-builder /app/server/internal/webapp/static-next ./internal/webapp/static-next

RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w -X github.com/jampat000/Xuva/server/internal/buildinfo.Version=${XUVA_VERSION} -X github.com/jampat000/Xuva/server/internal/buildinfo.Commit=${XUVA_COMMIT} -X github.com/jampat000/Xuva/server/internal/buildinfo.Date=${XUVA_BUILD_DATE}" \
    -o /xuva ./cmd/Xuva

# Stage 3: minimal runtime image
FROM alpine:3.20

# ffmpeg + ffprobe for transcoding/probing; ca-certificates for HTTPS metadata.
RUN apk add --no-cache ffmpeg ca-certificates tzdata

# Non-root user for least-privilege operation.
RUN addgroup -g 1000 xuva && adduser -u 1000 -G xuva -s /bin/sh -D xuva

COPY --from=go-builder /xuva /usr/local/bin/xuva

RUN mkdir -p /data && chown xuva:xuva /data

VOLUME /data

USER xuva

# Environment variables
# XUVA_DATA_DIR      - path for database, cache, and generated files
# XUVA_MOVIES_PATH   - path where movie files are stored
# XUVA_TV_PATH       - path where TV episode files are stored
# XUVA_HTTP_ADDR     - listen address:port
# XUVA_SERVER_NAME   - display name shown in the web UI (default: Xuva)
# XUVA_FFMPEG_PATH   - override ffmpeg binary path
# XUVA_FFPROBE_PATH  - override ffprobe binary path
ENV XUVA_DATA_DIR=/data \
    XUVA_HTTP_ADDR=0.0.0.0:8097

EXPOSE 8097

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
  CMD wget -q -O /dev/null "http://127.0.0.1:8097/api/health" || exit 1

ENTRYPOINT ["/usr/local/bin/xuva"]
