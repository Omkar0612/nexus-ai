FROM golang:1.22-alpine AS builder
WORKDIR /app
RUN apk add --no-cache gcc musl-dev sqlite-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build -ldflags="-s -w" -o nexus ./cmd/nexus

FROM alpine:3.19
RUN apk add --no-cache ca-certificates sqlite-libs python3 py3-pip
WORKDIR /app
COPY --from=builder /app/nexus .
COPY --from=builder /app/config ./config
COPY --from=builder /app/workers ./workers
COPY --from=builder /app/skills ./skills
ENV NEXUS_CONFIG=/config/nexus.toml
EXPOSE 7700
HEALTHCHECK --interval=30s --timeout=5s CMD wget -qO- http://localhost:7700/api/health || exit 1
ENTRYPOINT ["./nexus", "start"]
