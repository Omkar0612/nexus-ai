# syntax=docker/dockerfile:1
# ---- Build stage ----
FROM golang:1.26-alpine AS builder
WORKDIR /app

# gcc + musl needed for CGO (sqlite3)
RUN apk add --no-cache gcc musl-dev sqlite-dev

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w" \
    -trimpath \
    -o nexus ./cmd/nexus

# ---- Runtime stage ----
# alpine:3.19 — minimal attack surface; no python3/pip in runtime image
FROM alpine:3.19

# Run as non-root user — prevents container privilege escalation
RUN addgroup -S nexus && adduser -S -G nexus -u 10001 nexus

RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app
COPY --from=builder /app/nexus .

# Use explicit non-root user
USER nexus

ENV NEXUS_CONFIG=/config/nexus.toml
EXPOSE 7700

HEALTHCHECK --interval=30s --timeout=5s --start-period=5s CMD \
    wget -qO- http://localhost:7700/api/health || exit 1

ENTRYPOINT ["./nexus", "start"]
